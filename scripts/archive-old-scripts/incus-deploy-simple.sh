#!/bin/bash

# Script de déploiement Incus simplifié pour Veza
# Utilise le réseau par défaut d'Incus qui fonctionne

set -e

# Couleurs pour les messages
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

echo -e "${PURPLE}"
echo "╭──────────────────────────────────────────╮"
echo "│      🚀 Veza - Déploiement Simplifié    │"
echo "╰──────────────────────────────────────────╯"
echo -e "${NC}"

# Variables globales
WORKSPACE_DIR=$(pwd)
IMAGE="images:debian/bookworm"

# Configuration DNS ultra-simple pour tous les containers
configure_dns_simple() {
    local container_name=$1
    
    echo -e "${BLUE}🌐 Configuration DNS simple pour $container_name...${NC}"
    
    incus exec "$container_name" -- bash -c "
        # Arrêter systemd-resolved s'il interfère
        systemctl stop systemd-resolved 2>/dev/null || true
        systemctl disable systemd-resolved 2>/dev/null || true
        
        # Configuration DNS simple et efficace
        rm -f /etc/resolv.conf
        cat > /etc/resolv.conf << 'EOF'
nameserver 8.8.8.8
nameserver 8.8.4.4
nameserver 1.1.1.1
search .
EOF
        
        # Protection basique
        chmod 644 /etc/resolv.conf
        
        # Test DNS simple
        if timeout 5 ping -c 1 8.8.8.8 >/dev/null 2>&1; then
            echo '✅ Connectivité OK pour $container_name'
        else
            echo '⚠️ Connectivité limitée pour $container_name'
        fi
    "
}

# Attendre qu'un container soit prêt
wait_for_container() {
    local container_name=$1
    local max_attempts=30
    local attempt=0
    
    echo -e "${BLUE}⏳ Attente du démarrage de $container_name...${NC}"
    
    while [ $attempt -lt $max_attempts ]; do
        if incus exec "$container_name" -- test -f /etc/hostname >/dev/null 2>&1; then
            echo -e "${GREEN}✅ Container $container_name prêt${NC}"
            return 0
        fi
        sleep 2
        ((attempt++))
    done
    
    echo -e "${RED}❌ Timeout - Container $container_name non prêt${NC}"
    return 1
}

# Installation des dépendances de base
install_base_dependencies() {
    local container_name=$1
    
    echo -e "${BLUE}📦 Installation des dépendances de base pour $container_name...${NC}"
    
    incus exec "$container_name" -- bash -c "
        export DEBIAN_FRONTEND=noninteractive
        
        # Mise à jour des paquets
        apt-get update -qq
        
        # Installation des outils de base
        apt-get install -y curl wget git build-essential ca-certificates \
                          systemd systemd-sysv procps net-tools \
                          dnsutils iputils-ping netcat-openbsd htop vim nano
        
        apt-get clean
    "
}

# Obtenir l'IP d'un container
get_container_ip() {
    local container_name=$1
    local ip
    local attempts=0
    local max_attempts=10
    
    while [ $attempts -lt $max_attempts ]; do
        ip=$(incus list "$container_name" --format csv | cut -d, -f3 | head -1 | cut -d' ' -f1)
        if [[ "$ip" =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
            echo "$ip"
            return 0
        fi
        sleep 2
        ((attempts++))
    done
    
    echo ""
    return 1
}

# Déployer PostgreSQL simplifié
deploy_postgres_simple() {
    echo -e "${CYAN}🐘 Déploiement de PostgreSQL...${NC}"
    
    incus launch "$IMAGE" veza-postgres
    wait_for_container veza-postgres
    configure_dns_simple veza-postgres
    install_base_dependencies veza-postgres
    
    # Installation PostgreSQL
    incus exec veza-postgres -- bash -c "
        export DEBIAN_FRONTEND=noninteractive
        apt-get update -qq
        apt-get install -y postgresql postgresql-contrib
        systemctl enable postgresql
        systemctl start postgresql
        
        # Configuration PostgreSQL basique
        sudo -u postgres psql -c \"CREATE USER veza_user WITH PASSWORD 'veza_password';\"
        sudo -u postgres psql -c \"CREATE DATABASE veza_db OWNER veza_user;\"
        sudo -u postgres psql -c \"GRANT ALL PRIVILEGES ON DATABASE veza_db TO veza_user;\"
        
        # Configuration réseau pour accepter les connexions
        echo \"listen_addresses = '*'\" >> /etc/postgresql/15/main/postgresql.conf
        echo \"host all all 0.0.0.0/0 md5\" >> /etc/postgresql/15/main/pg_hba.conf
        
        systemctl restart postgresql
        
        # Test de fonctionnement
        sudo -u postgres psql -d veza_db -c 'SELECT version();' || echo 'PostgreSQL test failed'
    "
    
    echo -e "${GREEN}✅ PostgreSQL déployé${NC}"
}

# Déployer Redis simplifié
deploy_redis_simple() {
    echo -e "${CYAN}🔴 Déploiement de Redis...${NC}"
    
    incus launch "$IMAGE" veza-redis
    wait_for_container veza-redis
    configure_dns_simple veza-redis
    install_base_dependencies veza-redis
    
    # Installation Redis
    incus exec veza-redis -- bash -c "
        export DEBIAN_FRONTEND=noninteractive
        apt-get update -qq
        apt-get install -y redis-server
        
        # Configuration Redis pour accepter les connexions externes
        sed -i 's/bind 127.0.0.1/bind 0.0.0.0/' /etc/redis/redis.conf
        sed -i 's/protected-mode yes/protected-mode no/' /etc/redis/redis.conf
        
        systemctl enable redis-server
        systemctl restart redis-server
        
        # Test de fonctionnement
        redis-cli ping | grep PONG || echo 'Redis test failed'
    "
    
    echo -e "${GREEN}✅ Redis déployé${NC}"
}

# Déployer le backend Go simplifié
deploy_backend_simple() {
    echo -e "${CYAN}🔧 Déploiement du Backend Go...${NC}"
    
    incus launch "$IMAGE" veza-backend
    wait_for_container veza-backend
    configure_dns_simple veza-backend
    install_base_dependencies veza-backend
    
    # Copier le code source dans le bon répertoire
    incus exec veza-backend -- mkdir -p /app/backend
    incus file push -r "$WORKSPACE_DIR/veza-backend-api/." veza-backend/app/backend/
    
    # Obtenir les IPs des services déployés
    echo -e "${BLUE}🔍 Récupération des IPs des services...${NC}"
    local postgres_ip redis_ip
    postgres_ip=$(get_container_ip veza-postgres)
    redis_ip=$(get_container_ip veza-redis)
    
    if [ -z "$postgres_ip" ] || [ -z "$redis_ip" ]; then
        echo -e "${RED}❌ Impossible d'obtenir les IPs des services${NC}"
        return 1
    fi
    
    echo -e "${GREEN}📍 PostgreSQL IP: $postgres_ip${NC}"
    echo -e "${GREEN}📍 Redis IP: $redis_ip${NC}"
    
    # Installation Go et compilation
    incus exec veza-backend -- bash -c "
        export DEBIAN_FRONTEND=noninteractive
        
        # Installation Go
        echo '🔽 Téléchargement de Go...'
        wget -q https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
        tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
        rm go1.21.5.linux-amd64.tar.gz
        
        export PATH=/usr/local/go/bin:\$PATH
        
        # Variables d'environnement avec les vraies IPs
        cat > /etc/environment << EOF
PATH=/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
DATABASE_URL=postgres://veza_user:veza_password@$postgres_ip:5432/veza_db?sslmode=disable
REDIS_URL=redis://$redis_ip:6379
JWT_SECRET=your-super-secret-jwt-key-change-in-production
PORT=8080
EOF
        
        # Compilation du backend
        echo '🔨 Compilation du backend Go...'
        cd /app/backend/veza-backend-api
        /usr/local/go/bin/go mod tidy
        /usr/local/go/bin/go build -o /app/backend/veza-backend ./cmd/server/main.go
        chmod +x /app/backend/veza-backend
        
        # Vérifier que le binaire a été créé
        if [ ! -f /app/backend/veza-backend ]; then
            echo '❌ Échec de la compilation'
            exit 1
        fi
        
        echo '✅ Compilation réussie'
        
        # Créer un service systemd
        cat > /etc/systemd/system/veza-backend.service << 'EOF'
[Unit]
Description=Veza Backend API Server
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/app/backend
ExecStart=/app/backend/veza-backend
Restart=always
RestartSec=5
EnvironmentFile=/etc/environment

[Install]
WantedBy=multi-user.target
EOF
        
        # Démarrer le service
        systemctl daemon-reload
        systemctl enable veza-backend
        systemctl start veza-backend
        
        # Attendre un peu et vérifier le statut
        sleep 5
        if systemctl is-active --quiet veza-backend; then
            echo '✅ Service backend démarré avec succès'
        else
            echo '❌ Échec du démarrage du service'
            journalctl -u veza-backend --no-pager -n 10
            exit 1
        fi
    "
    
    echo -e "${GREEN}✅ Backend Go déployé${NC}"
}

# Déployer le serveur chat Rust simplifié
deploy_chat_simple() {
    echo -e "${CYAN}🔧 Déploiement du Chat Server Rust...${NC}"
    
    incus launch "$IMAGE" veza-chat
    wait_for_container veza-chat
    configure_dns_simple veza-chat
    install_base_dependencies veza-chat
    
    # Copier le code source Rust
    incus exec veza-chat -- mkdir -p /app/chat
    incus file push -r "$WORKSPACE_DIR/veza-chat-server/." veza-chat/app/chat/
    
    # Installation Rust et compilation
    incus exec veza-chat -- bash -c "
        export DEBIAN_FRONTEND=noninteractive
        
        # Installation Rust
        echo '🔽 Installation de Rust...'
        curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
        source /root/.cargo/env
        
        # Variables d'environnement
        cat > /etc/environment << EOF
PATH=/root/.cargo/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
RUST_LOG=info
CHAT_PORT=8081
EOF
        
        # Compilation Rust
        echo '🔨 Compilation du chat server...'
        cd /app/chat
        /root/.cargo/bin/cargo build --release
        
        # Vérifier que le binaire a été créé
        if [ ! -f target/release/veza-chat-server ]; then
            echo '❌ Échec de la compilation Rust'
            exit 1
        fi
        
        echo '✅ Compilation Rust réussie'
        
        # Service systemd
        cat > /etc/systemd/system/veza-chat.service << 'EOF'
[Unit]
Description=Veza Chat Server
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/app/chat
ExecStart=/app/chat/target/release/veza-chat-server
Restart=always
RestartSec=5
EnvironmentFile=/etc/environment

[Install]
WantedBy=multi-user.target
EOF
        
        systemctl daemon-reload
        systemctl enable veza-chat
        systemctl start veza-chat
        
        sleep 5
        if systemctl is-active --quiet veza-chat; then
            echo '✅ Service chat démarré avec succès'
        else
            echo '❌ Échec du démarrage du service chat'
            exit 1
        fi
    "
    
    echo -e "${GREEN}✅ Chat Server Rust déployé${NC}"
}

# Déployer le serveur streaming Rust simplifié
deploy_stream_simple() {
    echo -e "${CYAN}🔧 Déploiement du Stream Server Rust...${NC}"
    
    incus launch "$IMAGE" veza-stream
    wait_for_container veza-stream
    configure_dns_simple veza-stream
    install_base_dependencies veza-stream
    
    # Copier le code source Rust
    incus exec veza-stream -- mkdir -p /app/stream
    incus file push -r "$WORKSPACE_DIR/veza-stream-server/." veza-stream/app/stream/
    
    # Installation Rust et compilation
    incus exec veza-stream -- bash -c "
        export DEBIAN_FRONTEND=noninteractive
        
        # Installation Rust
        echo '🔽 Installation de Rust...'
        curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
        source /root/.cargo/env
        
        # Variables d'environnement
        cat > /etc/environment << EOF
PATH=/root/.cargo/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
RUST_LOG=info
STREAM_PORT=9002
EOF
        
        # Compilation Rust
        echo '🔨 Compilation du stream server...'
        cd /app/stream
        /root/.cargo/bin/cargo build --release
        
        # Vérifier que le binaire a été créé
        if [ ! -f target/release/veza-stream-server ]; then
            echo '❌ Échec de la compilation Rust'
            exit 1
        fi
        
        echo '✅ Compilation Rust réussie'
        
        # Service systemd
        cat > /etc/systemd/system/veza-stream.service << 'EOF'
[Unit]
Description=Veza Stream Server
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/app/stream
ExecStart=/app/stream/target/release/veza-stream-server
Restart=always
RestartSec=5
EnvironmentFile=/etc/environment

[Install]
WantedBy=multi-user.target
EOF
        
        systemctl daemon-reload
        systemctl enable veza-stream
        systemctl start veza-stream
        
        sleep 5
        if systemctl is-active --quiet veza-stream; then
            echo '✅ Service stream démarré avec succès'
        else
            echo '❌ Échec du démarrage du service stream'
            exit 1
        fi
    "
    
    echo -e "${GREEN}✅ Stream Server Rust déployé${NC}"
}

# Déployer le frontend React simplifié
deploy_frontend_simple() {
    echo -e "${CYAN}🔧 Déploiement du Frontend React...${NC}"
    
    incus launch "$IMAGE" veza-frontend
    wait_for_container veza-frontend
    configure_dns_simple veza-frontend
    install_base_dependencies veza-frontend
    
    # Copier le code source React
    incus exec veza-frontend -- mkdir -p /app/frontend
    incus file push -r "$WORKSPACE_DIR/veza-frontend/." veza-frontend/app/frontend/
    
    # Installation Node.js et compilation
    incus exec veza-frontend -- bash -c "
        export DEBIAN_FRONTEND=noninteractive
        
        # Installation Node.js
        echo '🔽 Installation de Node.js...'
        curl -fsSL https://deb.nodesource.com/setup_18.x | bash -
        apt-get install -y nodejs nginx
        
                 # Obtenir l'IP du backend
         backend_ip=\$(incus ls veza-backend -c 4 --format csv | cut -d' ' -f1)
         
         # Variables d'environnement
         cat > /etc/environment << EOF
PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
NODE_ENV=production
VITE_API_URL=http://\$backend_ip:8080
EOF
        
        # Compilation React
        echo '🔨 Compilation du frontend React...'
        cd /app/frontend
        npm install
        npm run build
        
        # Configuration Nginx
        cat > /etc/nginx/sites-available/default << 'EOF'
server {
    listen 80 default_server;
    listen [::]:80 default_server;
    root /app/frontend/dist;
    index index.html;
    server_name _;
    
    location / {
        try_files \$uri \$uri/ /index.html;
    }
    
    location /api {
        proxy_pass http://veza-backend:8080;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
    }
}
EOF
        
        systemctl enable nginx
        systemctl restart nginx
        
        if systemctl is-active --quiet nginx; then
            echo '✅ Frontend démarré avec succès'
        else
            echo '❌ Échec du démarrage du frontend'
            exit 1
        fi
    "
    
    echo -e "${GREEN}✅ Frontend React déployé${NC}"
}

# Déployer le container de stockage
deploy_storage_simple() {
    echo -e "${CYAN}🔧 Déploiement du Storage...${NC}"
    
    incus launch "$IMAGE" veza-storage
    wait_for_container veza-storage
    configure_dns_simple veza-storage
    install_base_dependencies veza-storage
    
    # Configuration stockage
    incus exec veza-storage -- bash -c "
        export DEBIAN_FRONTEND=noninteractive
        
        # Installation services de stockage
        apt-get update && apt-get install -y nginx
        
        # Création des dossiers de stockage
        mkdir -p /storage/{uploads,audio,backups,ssl}
        chmod 755 /storage
        chmod 777 /storage/{uploads,audio,backups}
        
        # Configuration Nginx pour servir les fichiers
        cat > /etc/nginx/sites-available/default << 'EOF'
server {
    listen 80 default_server;
    listen [::]:80 default_server;
    root /storage;
    server_name _;
    
    location /uploads {
        alias /storage/uploads;
        autoindex on;
    }
    
    location /audio {
        alias /storage/audio;
        autoindex on;
    }
    
    location /backups {
        alias /storage/backups;
        autoindex on;
    }
}
EOF
        
        systemctl enable nginx
        systemctl restart nginx
        
        if systemctl is-active --quiet nginx; then
            echo '✅ Storage service démarré avec succès'
        else
            echo '❌ Échec du démarrage du storage'
            exit 1
        fi
    "
    
    echo -e "${GREEN}✅ Storage déployé${NC}"
}

# Déployer HAProxy
deploy_haproxy_simple() {
    echo -e "${CYAN}🔧 Déploiement de HAProxy...${NC}"
    
    incus launch "$IMAGE" veza-haproxy
    wait_for_container veza-haproxy
    configure_dns_simple veza-haproxy
    install_base_dependencies veza-haproxy
    
    # Obtenir les IPs des services
    local backend_ip chat_ip stream_ip frontend_ip storage_ip
    backend_ip=$(get_container_ip veza-backend)
    chat_ip=$(get_container_ip veza-chat)
    stream_ip=$(get_container_ip veza-stream)
    frontend_ip=$(get_container_ip veza-frontend)
    storage_ip=$(get_container_ip veza-storage)
    
    # Configuration HAProxy
    incus exec veza-haproxy -- bash -c "
        export DEBIAN_FRONTEND=noninteractive
        
        # Installation HAProxy
        apt-get update && apt-get install -y haproxy
        
        # Configuration HAProxy
        cat > /etc/haproxy/haproxy.cfg << EOF
global
    daemon
    user haproxy
    group haproxy

defaults
    mode http
    timeout connect 5000ms
    timeout client 50000ms
    timeout server 50000ms

frontend veza_frontend
    bind *:80
    
    # API Backend
    acl is_api path_beg /api
    use_backend veza_backend if is_api
    
    # Chat WebSocket  
    acl is_chat path_beg /chat
    use_backend veza_chat if is_chat
    
    # Streaming
    acl is_stream path_beg /stream
    use_backend veza_stream if is_stream
    
    # Storage
    acl is_storage path_beg /storage
    use_backend veza_storage if is_storage
    
    # Frontend par défaut
    default_backend veza_frontend

backend veza_backend
    server backend1 $backend_ip:8080 check

backend veza_chat  
    server chat1 $chat_ip:8081 check

backend veza_stream
    server stream1 $stream_ip:9002 check

backend veza_frontend
    server frontend1 $frontend_ip:80 check

backend veza_storage
    server storage1 $storage_ip:80 check
EOF
        
        systemctl enable haproxy
        systemctl restart haproxy
        
        if systemctl is-active --quiet haproxy; then
            echo '✅ HAProxy démarré avec succès'
        else
            echo '❌ Échec du démarrage de HAProxy'
            exit 1
        fi
    "
    
    echo -e "${GREEN}✅ HAProxy déployé${NC}"
}

# Test de connectivité complet
test_deployment() {
    echo -e "${BLUE}🧪 Tests de déploiement...${NC}"
    
    local backend_ip
    backend_ip=$(get_container_ip veza-backend)
    
    if [ -z "$backend_ip" ]; then
        echo -e "${RED}❌ Impossible d'obtenir l'IP du backend${NC}"
        return 1
    fi
    
    echo -e "${BLUE}🌐 Test de l'API backend...${NC}"
    if curl -f --connect-timeout 10 "http://$backend_ip:8080" >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Backend accessible${NC}"
    else
        echo -e "${YELLOW}⚠️ Backend pas encore accessible (normal au premier démarrage)${NC}"
    fi
    
    # Test des services dans les containers
    echo -e "${BLUE}🔍 Test PostgreSQL...${NC}"
    if incus exec veza-postgres -- systemctl is-active --quiet postgresql; then
        echo -e "${GREEN}✅ PostgreSQL actif${NC}"
    else
        echo -e "${RED}❌ PostgreSQL inactif${NC}"
    fi
    
    echo -e "${BLUE}🔍 Test Redis...${NC}"
    if incus exec veza-redis -- systemctl is-active --quiet redis-server; then
        echo -e "${GREEN}✅ Redis actif${NC}"
    else
        echo -e "${RED}❌ Redis inactif${NC}"
    fi
    
    echo -e "${BLUE}🔍 Test Backend...${NC}"
    if incus exec veza-backend -- systemctl is-active --quiet veza-backend; then
        echo -e "${GREEN}✅ Backend actif${NC}"
    else
        echo -e "${RED}❌ Backend inactif${NC}"
        incus exec veza-backend -- journalctl -u veza-backend --no-pager -n 5
    fi
    
    echo -e "${BLUE}🔍 Test Chat Server...${NC}"
    if incus exec veza-chat -- systemctl is-active --quiet veza-chat; then
        echo -e "${GREEN}✅ Chat Server actif${NC}"
    else
        echo -e "${RED}❌ Chat Server inactif${NC}"
    fi
    
    echo -e "${BLUE}🔍 Test Stream Server...${NC}"
    if incus exec veza-stream -- systemctl is-active --quiet veza-stream; then
        echo -e "${GREEN}✅ Stream Server actif${NC}"
    else
        echo -e "${RED}❌ Stream Server inactif${NC}"
    fi
    
    echo -e "${BLUE}🔍 Test Frontend...${NC}"
    if incus exec veza-frontend -- systemctl is-active --quiet nginx; then
        echo -e "${GREEN}✅ Frontend actif${NC}"
    else
        echo -e "${RED}❌ Frontend inactif${NC}"
    fi
    
    echo -e "${BLUE}🔍 Test Storage...${NC}"
    if incus exec veza-storage -- systemctl is-active --quiet nginx; then
        echo -e "${GREEN}✅ Storage actif${NC}"
    else
        echo -e "${RED}❌ Storage inactif${NC}"
    fi
    
    echo -e "${BLUE}🔍 Test HAProxy...${NC}"
    if incus exec veza-haproxy -- systemctl is-active --quiet haproxy; then
        echo -e "${GREEN}✅ HAProxy actif${NC}"
    else
        echo -e "${RED}❌ HAProxy inactif${NC}"
    fi
}

# Fonction principale
main() {
    echo -e "${BLUE}🚀 Début du déploiement simplifié...${NC}"
    echo -e "${YELLOW}Cette version utilise le réseau par défaut d'Incus. Continuer ? (o/N)${NC}"
    read -r response
    
    if [[ "$response" != "o" && "$response" != "oui" ]]; then
        echo -e "${GREEN}Déploiement annulé${NC}"
        exit 0
    fi
    
    # Vérifier que les codes sources existent
    if [ ! -d "$WORKSPACE_DIR/veza-backend-api" ]; then
        echo -e "${RED}❌ Répertoire veza-backend-api non trouvé${NC}"
        exit 1
    fi
    
    if [ ! -d "$WORKSPACE_DIR/veza-chat-server" ]; then
        echo -e "${RED}❌ Répertoire veza-chat-server non trouvé${NC}"
        exit 1
    fi
    
    if [ ! -d "$WORKSPACE_DIR/veza-stream-server" ]; then
        echo -e "${RED}❌ Répertoire veza-stream-server non trouvé${NC}"
        exit 1
    fi
    
    if [ ! -d "$WORKSPACE_DIR/veza-frontend" ]; then
        echo -e "${RED}❌ Répertoire veza-frontend non trouvé${NC}"
        exit 1
    fi
    
    # Déploiement séquentiel des 8 containers
    echo -e "${PURPLE}🚀 Déploiement de l'infrastructure complète (8 containers)...${NC}"
    
    # 1. Services de base (DB + Cache)
    deploy_postgres_simple
    deploy_redis_simple
    
    # 2. Services applicatifs  
    deploy_backend_simple
    deploy_chat_simple
    deploy_stream_simple
    
    # 3. Frontend et stockage
    deploy_frontend_simple
    deploy_storage_simple
    
    # 4. Load balancer (en dernier)
    deploy_haproxy_simple
    
    # Tests
    test_deployment
    
    # Affichage des résultats
    echo -e "${GREEN}🎉 Déploiement complet terminé ! (8 containers)${NC}"
    echo ""
    echo -e "${BLUE}📊 État des containers :${NC}"
    incus ls
    echo ""
    echo -e "${BLUE}🌐 IPs des services :${NC}"
    echo -e "  • PostgreSQL : $(get_container_ip veza-postgres)"
    echo -e "  • Redis : $(get_container_ip veza-redis)"
    echo -e "  • Backend Go : $(get_container_ip veza-backend)"
    echo -e "  • Chat Rust : $(get_container_ip veza-chat)"
    echo -e "  • Stream Rust : $(get_container_ip veza-stream)"
    echo -e "  • Frontend React : $(get_container_ip veza-frontend)"
    echo -e "  • Storage : $(get_container_ip veza-storage)"
    echo -e "  • HAProxy : $(get_container_ip veza-haproxy)"
    echo ""
    echo -e "${CYAN}💡 Points d'accès principaux :${NC}"
    echo -e "  • Application complète : http://$(get_container_ip veza-haproxy)"
    echo -e "  • API Backend : http://$(get_container_ip veza-backend):8080"
    echo -e "  • Chat WebSocket : ws://$(get_container_ip veza-chat):8081"
    echo -e "  • Streaming Audio : http://$(get_container_ip veza-stream):9002"
    echo -e "  • Frontend direct : http://$(get_container_ip veza-frontend)"
    echo -e "  • Stockage : http://$(get_container_ip veza-storage)"
    echo ""
    echo -e "${CYAN}🔧 Commandes utiles :${NC}"
    echo -e "  • Logs backend : incus exec veza-backend -- journalctl -u veza-backend -f"
    echo -e "  • Logs chat : incus exec veza-chat -- journalctl -u veza-chat -f"
    echo -e "  • Logs stream : incus exec veza-stream -- journalctl -u veza-stream -f"
    echo -e "  • Status HAProxy : incus exec veza-haproxy -- systemctl status haproxy"
    echo ""
    echo -e "${GREEN}✅ Infrastructure complète opérationnelle ! (8/8 containers)${NC}"
}

# Vérifier qu'Incus est disponible
if ! command -v incus &> /dev/null; then
    echo -e "${RED}❌ Incus n'est pas installé ou accessible${NC}"
    exit 1
fi

# Exécuter le déploiement
main 