#!/bin/bash

# Script complet : Build + Start tous les services Veza
set -e

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

echo -e "${BLUE}"
echo "╭──────────────────────────────────────────╮"
echo "│     🔨 Build & Start Tous les Services    │"
echo "╰──────────────────────────────────────────╯"
echo -e "${NC}"

# Étape 1: Construction Backend Go
echo -e "${CYAN}🔧 Construction du Backend Go...${NC}"
incus exec veza-backend -- bash -c "
    cd /opt/veza/backend
    export PATH=/usr/local/go/bin:\$PATH
    export GOPATH=/opt/veza/go
    
    echo '📦 Installation des dépendances Go...'
    go mod tidy
    
    echo '🔨 Compilation du backend...'
    go build -o main ./cmd/server/main.go
    
    echo '✅ Backend Go compilé avec succès'
"

# Étape 2: Construction Chat Rust
echo -e "${CYAN}💬 Construction du Chat Server Rust...${NC}"
incus exec veza-chat -- bash -c "
    cd /opt/veza/chat
    export PATH=/root/.cargo/bin:\$PATH
    
    echo '📦 Installation des dépendances Rust...'
    cargo check
    
    echo '🔨 Compilation du chat server...'
    cargo build --release
    
    echo '✅ Chat Server Rust compilé avec succès'
"

# Étape 3: Construction Stream Rust
echo -e "${CYAN}🎵 Construction du Stream Server Rust...${NC}"
incus exec veza-stream -- bash -c "
    cd /opt/veza/stream
    export PATH=/root/.cargo/bin:\$PATH
    
    echo '📦 Installation des dépendances Rust...'
    cargo check
    
    echo '🔨 Compilation du stream server...'
    cargo build --release
    
    echo '✅ Stream Server Rust compilé avec succès'
"

# Étape 4: Installation Frontend React
echo -e "${CYAN}⚛️ Installation du Frontend React...${NC}"
incus exec veza-frontend -- bash -c "
    cd /opt/veza/frontend
    
    echo '📦 Installation des dépendances Node.js...'
    npm install
    
    echo '✅ Frontend React prêt'
"

# Étape 5: Démarrage des services de base
echo -e "${BLUE}🚀 Démarrage des services de base...${NC}"

echo -e "${CYAN}📊 Démarrage PostgreSQL...${NC}"
incus exec veza-postgres -- systemctl start postgresql
sleep 2

echo -e "${CYAN}🔴 Démarrage Redis...${NC}"
incus exec veza-redis -- systemctl start redis-server
sleep 2

echo -e "${CYAN}🗄️ Démarrage NFS Storage...${NC}"
incus exec veza-storage -- systemctl start nfs-kernel-server 2>/dev/null || echo "NFS déjà démarré"
sleep 2

# Étape 6: Démarrage des services applicatifs
echo -e "${BLUE}🔧 Démarrage des services applicatifs...${NC}"

echo -e "${CYAN}🔧 Démarrage Backend Go...${NC}"
incus exec veza-backend -- systemctl start veza-backend
sleep 3

echo -e "${CYAN}💬 Démarrage Chat Server...${NC}"
incus exec veza-chat -- systemctl start veza-chat
sleep 3

echo -e "${CYAN}🎵 Démarrage Stream Server...${NC}"
incus exec veza-stream -- systemctl start veza-stream
sleep 3

echo -e "${CYAN}⚛️ Démarrage Frontend React...${NC}"
incus exec veza-frontend -- systemctl start veza-frontend
sleep 3

# Étape 7: Installation et configuration HAProxy
echo -e "${CYAN}⚖️ Installation et configuration HAProxy...${NC}"
incus exec veza-haproxy -- bash -c "
    # Installation HAProxy si pas déjà installé
    if ! command -v haproxy &> /dev/null; then
        echo 'Installation HAProxy...'
        export DEBIAN_FRONTEND=noninteractive
        apt-get update -qq
        apt-get install -y haproxy
        systemctl enable haproxy
    fi
    
    # Configuration HAProxy avec les vraies IPs
    echo 'Configuration HAProxy...'
    cat > /etc/haproxy/haproxy.cfg << 'EOF'
global
    daemon
    maxconn 4096
    log stdout local0

defaults
    mode http
    timeout connect 5000ms
    timeout client 50000ms
    timeout server 50000ms
    option httplog
    log global

# Interface de stats HAProxy
listen stats
    bind *:8404
    stats enable
    stats uri /stats
    stats refresh 30s
    stats admin if TRUE

# Frontend principal
frontend veza_main
    bind *:80
    
    # Redirection API vers backend
    acl is_api path_beg /api/
    use_backend veza_backend if is_api
    
    # Par défaut vers frontend React
    default_backend veza_frontend

# Backend pour l'API Go
backend veza_backend
    balance roundrobin
    server backend1 \$(incus ls veza-backend -c 4 --format csv | cut -d' ' -f1):8080 check

# Backend pour le frontend React
backend veza_frontend
    balance roundrobin
    server frontend1 \$(incus ls veza-frontend -c 4 --format csv | cut -d' ' -f1):3000 check
EOF
    
    # Remplacer les variables par les vraies IPs
    BACKEND_IP=\$(incus ls veza-backend -c 4 --format csv | cut -d' ' -f1)
    FRONTEND_IP=\$(incus ls veza-frontend -c 4 --format csv | cut -d' ' -f1)
    
    sed -i \"s/\$(incus ls veza-backend -c 4 --format csv | cut -d' ' -f1)/\$BACKEND_IP/g\" /etc/haproxy/haproxy.cfg
    sed -i \"s/\$(incus ls veza-frontend -c 4 --format csv | cut -d' ' -f1)/\$FRONTEND_IP/g\" /etc/haproxy/haproxy.cfg
    
    # Démarrer HAProxy
    systemctl restart haproxy
    systemctl enable haproxy
"

# Étape 8: Vérification des services
echo -e "${BLUE}🔍 Vérification des services...${NC}"

services=(
    "veza-postgres:postgresql"
    "veza-redis:redis-server"
    "veza-backend:veza-backend"
    "veza-chat:veza-chat"
    "veza-stream:veza-stream"
    "veza-frontend:veza-frontend"
    "veza-haproxy:haproxy"
)

all_good=true

for service in "${services[@]}"; do
    container="${service%:*}"
    service_name="${service#*:}"
    
    echo -n -e "${BLUE}📦 $container ($service_name): ${NC}"
    
    status=$(incus exec "$container" -- systemctl is-active "$service_name" 2>/dev/null || echo "inactive")
    
    case $status in
        "active")
            echo -e "${GREEN}✅ Active${NC}"
            ;;
        *)
            echo -e "${RED}❌ $status${NC}"
            all_good=false
            ;;
    esac
done

echo ""

if [ "$all_good" = true ]; then
    echo -e "${GREEN}🎉 Tous les services sont démarrés avec succès !${NC}"
else
    echo -e "${YELLOW}⚠️ Certains services ont des problèmes${NC}"
fi

# Étape 9: Affichage des URLs d'accès
echo ""
echo -e "${BLUE}🌐 URLs d'accès:${NC}"

HAPROXY_IP=$(incus ls veza-haproxy -c 4 --format csv | cut -d' ' -f1)
FRONTEND_IP=$(incus ls veza-frontend -c 4 --format csv | cut -d' ' -f1)
BACKEND_IP=$(incus ls veza-backend -c 4 --format csv | cut -d' ' -f1)

echo -e "  ${GREEN}• Application complète: ${YELLOW}http://$HAPROXY_IP${NC} (via HAProxy)"
echo -e "  ${GREEN}• HAProxy Stats: ${YELLOW}http://$HAPROXY_IP:8404/stats${NC}"
echo -e "  ${GREEN}• Frontend direct: ${YELLOW}http://$FRONTEND_IP:3000${NC}"
echo -e "  ${GREEN}• Backend API: ${YELLOW}http://$BACKEND_IP:8080${NC}"

echo ""
echo -e "${CYAN}💡 Commandes utiles:${NC}"
echo -e "  ${YELLOW}./scripts/status-all-services.sh${NC}     # Voir le statut"
echo -e "  ${YELLOW}./scripts/quick-sync.sh component${NC}    # Resync un composant"
echo -e "  ${YELLOW}./scripts/watch-and-sync.sh${NC}         # Surveillance automatique"

echo ""
echo -e "${GREEN}🚀 Infrastructure Veza prête pour le développement !${NC}" 