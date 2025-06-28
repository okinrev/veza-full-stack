#!/bin/bash

# Script de déploiement complet Incus pour Veza
# Déploie les 8 containers avec configuration optimisée

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
echo "│      🚀 Veza - Déploiement Complet      │"
echo "╰──────────────────────────────────────────╯"
echo -e "${NC}"

# Variables globales
WORKSPACE_DIR=$(pwd)
IMAGE="images:debian/bookworm"

# Fonction pour attendre qu'un container soit prêt
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

# Fonction pour configurer l'IP statique
configure_static_ip() {
    local container_name=$1
    local ip_address=$2
    
    echo -e "${BLUE}🌐 Configuration IP statique pour $container_name ($ip_address)...${NC}"
    
    # Configurer l'IP statique en surchargeant le device du profil
    incus config device override "$container_name" eth0 \
        ipv4.address="$ip_address"
    
    # Redémarrer le container pour appliquer la configuration
    incus restart "$container_name"
    wait_for_container "$container_name"
    
    # Vérifier la configuration IP
    local retry=0
    while [ $retry -lt 10 ]; do
        if incus exec "$container_name" -- ip addr show eth0 | grep -q "$ip_address"; then
            echo -e "${GREEN}✅ IP $ip_address configurée pour $container_name${NC}"
            return 0
        fi
        
        # Forcer la configuration IP si nécessaire
        incus exec "$container_name" -- ip addr flush dev eth0 || true
        incus exec "$container_name" -- ip addr add "$ip_address/24" dev eth0 || true
        incus exec "$container_name" -- ip route add default via 10.100.0.1 || true
        
        sleep 3
        ((retry++))
    done
    
    echo -e "${YELLOW}⚠️ Configuration IP manuelle pour $container_name${NC}"
}

# Fonction pour installer les dépendances de base
install_base_dependencies() {
    local container_name=$1
    
    echo -e "${BLUE}📦 Installation des dépendances de base pour $container_name...${NC}"
    
    incus exec "$container_name" -- bash -c "
        export DEBIAN_FRONTEND=noninteractive
        apt-get update -qq
        apt-get install -y curl wget git build-essential ca-certificates gnupg lsb-release
        apt-get clean
    "
}

# Déployer PostgreSQL
deploy_postgres() {
    echo -e "${CYAN}🐘 Déploiement de PostgreSQL...${NC}"
    
    incus launch "$IMAGE" veza-postgres --profile veza-database
    wait_for_container veza-postgres
    configure_static_ip veza-postgres 10.100.0.15
    
    install_base_dependencies veza-postgres
    
    # Installation PostgreSQL
    incus exec veza-postgres -- bash -c "
        export DEBIAN_FRONTEND=noninteractive
        apt-get update -qq
        apt-get install -y postgresql postgresql-contrib
        systemctl enable postgresql
        systemctl start postgresql
        
        # Configuration PostgreSQL
        sudo -u postgres psql -c \"CREATE USER veza_user WITH PASSWORD 'veza_password';\"
        sudo -u postgres psql -c \"CREATE DATABASE veza_db OWNER veza_user;\"
        sudo -u postgres psql -c \"GRANT ALL PRIVILEGES ON DATABASE veza_db TO veza_user;\"
        
        # Configuration réseau
        echo \"listen_addresses = '*'\" >> /etc/postgresql/15/main/postgresql.conf
        echo \"host all all 10.100.0.0/24 md5\" >> /etc/postgresql/15/main/pg_hba.conf
        
        systemctl restart postgresql
    "
    
    # Importer le schéma de base de données
    if [ -f "$WORKSPACE_DIR/init-db.sql" ]; then
        echo -e "${BLUE}📊 Import du schéma de base de données...${NC}"
        incus file push "$WORKSPACE_DIR/init-db.sql" veza-postgres/tmp/
        incus exec veza-postgres -- sudo -u postgres psql veza_db < /tmp/init-db.sql
    fi
    
    echo -e "${GREEN}✅ PostgreSQL déployé (10.100.0.15)${NC}"
}

# Déployer Redis
deploy_redis() {
    echo -e "${CYAN}🔴 Déploiement de Redis...${NC}"
    
    incus launch "$IMAGE" veza-redis --profile veza-app
    wait_for_container veza-redis
    configure_static_ip veza-redis 10.100.0.17
    
    install_base_dependencies veza-redis
    
    # Installation Redis
    incus exec veza-redis -- bash -c "
        export DEBIAN_FRONTEND=noninteractive
        apt-get update -qq
        apt-get install -y redis-server
        
        # Configuration Redis
        sed -i 's/bind 127.0.0.1/bind 0.0.0.0/' /etc/redis/redis.conf
        sed -i 's/protected-mode yes/protected-mode no/' /etc/redis/redis.conf
        
        systemctl enable redis-server
        systemctl restart redis-server
    "
    
    echo -e "${GREEN}✅ Redis déployé (10.100.0.17)${NC}"
}

# Déployer le système de fichiers
deploy_storage() {
    echo -e "${CYAN}🗄️ Déploiement du système de fichiers...${NC}"
    
    incus launch "$IMAGE" veza-storage --profile veza-storage
    wait_for_container veza-storage
    configure_static_ip veza-storage 10.100.0.18
    
    install_base_dependencies veza-storage
    
    # Installation NFS
    incus exec veza-storage -- bash -c "
        export DEBIAN_FRONTEND=noninteractive
        apt-get update -qq
        apt-get install -y nfs-kernel-server
        
        # Configuration des exports NFS
        mkdir -p /storage/{uploads,audio,backups,cache}
        chown -R nobody:nogroup /storage
        chmod -R 755 /storage
        
        echo '/storage/uploads 10.100.0.0/24(rw,sync,no_subtree_check,no_root_squash)' >> /etc/exports
        echo '/storage/audio 10.100.0.0/24(rw,sync,no_subtree_check,no_root_squash)' >> /etc/exports
        echo '/storage/backups 10.100.0.0/24(rw,sync,no_subtree_check,no_root_squash)' >> /etc/exports
        echo '/storage/cache 10.100.0.0/24(rw,sync,no_subtree_check,no_root_squash)' >> /etc/exports
        
        systemctl enable nfs-kernel-server
        systemctl restart nfs-kernel-server
        exportfs -ra
    "
    
    echo -e "${GREEN}✅ Système de fichiers déployé (10.100.0.18)${NC}"
}

# Déployer le backend Go
deploy_backend() {
    echo -e "${CYAN}🔧 Déploiement du Backend Go...${NC}"
    
    incus launch "$IMAGE" veza-backend --profile veza-app
    wait_for_container veza-backend
    configure_static_ip veza-backend 10.100.0.12
    
    install_base_dependencies veza-backend
    
    # Installation Go
    incus exec veza-backend -- bash -c "
        export DEBIAN_FRONTEND=noninteractive
        
        # Installation Go
        wget -q https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
        tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
        echo 'export PATH=\$PATH:/usr/local/go/bin' >> /etc/profile
        echo 'export GOPATH=/app/go' >> /etc/profile
        
        # Monter le client NFS
        apt-get install -y nfs-common
        mkdir -p /app/uploads
        mount -t nfs 10.100.0.18:/storage/uploads /app/uploads
        echo '10.100.0.18:/storage/uploads /app/uploads nfs defaults 0 0' >> /etc/fstab
        
        # Variables d'environnement
        cat > /etc/environment << EOF
DATABASE_URL=postgres://veza_user:veza_password@10.100.0.15:5432/veza_db?sslmode=disable
REDIS_URL=redis://10.100.0.17:6379
JWT_SECRET=your-super-secret-jwt-key-change-in-production
PORT=8080
CHAT_SERVER_URL=http://10.100.0.13:8081
STREAM_SERVER_URL=http://10.100.0.14:8082
EOF
    "
    
    # Copier le code source
    incus exec veza-backend -- mkdir -p /app/backend
    incus file push -r "$WORKSPACE_DIR/veza-backend-api/." veza-backend/app/backend/
    
    echo -e "${GREEN}✅ Backend Go déployé (10.100.0.12)${NC}"
}

# Déployer le serveur de chat Rust
deploy_chat() {
    echo -e "${CYAN}💬 Déploiement du Chat Server Rust...${NC}"
    
    incus launch "$IMAGE" veza-chat --profile veza-app
    wait_for_container veza-chat
    configure_static_ip veza-chat 10.100.0.13
    
    install_base_dependencies veza-chat
    
    # Installation Rust
    incus exec veza-chat -- bash -c "
        export DEBIAN_FRONTEND=noninteractive
        
        # Installation Rust
        curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
        source ~/.cargo/env
        echo 'source ~/.cargo/env' >> ~/.bashrc
        
        # Variables d'environnement
        cat > /etc/environment << EOF
DATABASE_URL=postgres://veza_user:veza_password@10.100.0.15:5432/veza_db?sslmode=disable
REDIS_URL=redis://10.100.0.17:6379
JWT_SECRET=your-super-secret-jwt-key-change-in-production
PORT=8081
RUST_LOG=chat_server=debug,tower_http=debug
EOF
    "
    
    # Copier le code source
    incus exec veza-chat -- mkdir -p /app/chat
    incus file push -r "$WORKSPACE_DIR/veza-chat-server/." veza-chat/app/chat/
    
    echo -e "${GREEN}✅ Chat Server déployé (10.100.0.13)${NC}"
}

# Déployer le serveur de streaming Rust
deploy_stream() {
    echo -e "${CYAN}🎵 Déploiement du Stream Server Rust...${NC}"
    
    incus launch "$IMAGE" veza-stream --profile veza-app
    wait_for_container veza-stream
    configure_static_ip veza-stream 10.100.0.14
    
    install_base_dependencies veza-stream
    
    # Installation Rust
    incus exec veza-stream -- bash -c "
        export DEBIAN_FRONTEND=noninteractive
        
        # Installation Rust
        curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
        source ~/.cargo/env
        echo 'source ~/.cargo/env' >> ~/.bashrc
        
        # Monter le stockage audio
        apt-get install -y nfs-common
        mkdir -p /storage/audio
        mount -t nfs 10.100.0.18:/storage/audio /storage/audio
        echo '10.100.0.18:/storage/audio /storage/audio nfs defaults 0 0' >> /etc/fstab
        
        # Variables d'environnement
        cat > /etc/environment << EOF
DATABASE_URL=postgres://veza_user:veza_password@10.100.0.15:5432/veza_db?sslmode=disable
REDIS_URL=redis://10.100.0.17:6379
JWT_SECRET=your-super-secret-jwt-key-change-in-production
PORT=8082
AUDIO_DIR=/storage/audio
RUST_LOG=stream_server=debug,tower_http=debug
ALLOWED_ORIGINS=http://10.100.0.11:5173,http://10.100.0.16
EOF
    "
    
    # Copier le code source
    incus exec veza-stream -- mkdir -p /app/stream
    incus file push -r "$WORKSPACE_DIR/veza-stream-server/." veza-stream/app/stream/
    
    echo -e "${GREEN}✅ Stream Server déployé (10.100.0.14)${NC}"
}

# Déployer le frontend React
deploy_frontend() {
    echo -e "${CYAN}⚛️ Déploiement du Frontend React...${NC}"
    
    incus launch "$IMAGE" veza-frontend --profile veza-app
    wait_for_container veza-frontend
    configure_static_ip veza-frontend 10.100.0.11
    
    install_base_dependencies veza-frontend
    
    # Installation Node.js
    incus exec veza-frontend -- bash -c "
        export DEBIAN_FRONTEND=noninteractive
        
        # Installation Node.js 20
        curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
        apt-get install -y nodejs
        
        # Variables d'environnement
        cat > /etc/environment << EOF
NODE_ENV=development
VITE_API_URL=http://10.100.0.12:8080/api/v1
VITE_WS_CHAT_URL=ws://10.100.0.13:8081/ws
VITE_WS_STREAM_URL=ws://10.100.0.14:8082/ws
EOF
    "
    
    # Copier le code source
    incus exec veza-frontend -- mkdir -p /app/frontend
    incus file push -r "$WORKSPACE_DIR/veza-frontend/." veza-frontend/app/frontend/
    
    echo -e "${GREEN}✅ Frontend React déployé (10.100.0.11)${NC}"
}

# Déployer HAProxy
deploy_haproxy() {
    echo -e "${CYAN}⚖️ Déploiement de HAProxy...${NC}"
    
    incus launch "$IMAGE" veza-haproxy --profile veza-app
    wait_for_container veza-haproxy
    configure_static_ip veza-haproxy 10.100.0.16
    
    install_base_dependencies veza-haproxy
    
    # Installation HAProxy
    incus exec veza-haproxy -- bash -c "
        export DEBIAN_FRONTEND=noninteractive
        apt-get update -qq
        apt-get install -y haproxy
        systemctl enable haproxy
    "
    
    # Copier la configuration HAProxy
    if [ -f "$WORKSPACE_DIR/haproxy.cfg" ]; then
        incus file push "$WORKSPACE_DIR/haproxy.cfg" veza-haproxy/etc/haproxy/haproxy.cfg
        incus exec veza-haproxy -- systemctl restart haproxy
    fi
    
    echo -e "${GREEN}✅ HAProxy déployé (10.100.0.16)${NC}"
}

# Fonction principale de déploiement
main() {
    echo -e "${BLUE}🚀 Début du déploiement complet...${NC}"
    echo -e "${YELLOW}Cette opération va créer 8 containers. Continuer ? (o/N)${NC}"
    read -r response
    
    if [[ "$response" != "o" && "$response" != "oui" ]]; then
        echo -e "${GREEN}Déploiement annulé${NC}"
        exit 0
    fi
    
    # Déploiement dans l'ordre optimal
    echo -e "${BLUE}📋 Ordre de déploiement :${NC}"
    echo -e "  1. PostgreSQL (Base de données)"
    echo -e "  2. Redis (Cache)"
    echo -e "  3. Système de fichiers (NFS)"
    echo -e "  4. Backend Go (API)"
    echo -e "  5. Chat Server Rust"
    echo -e "  6. Stream Server Rust"
    echo -e "  7. Frontend React"
    echo -e "  8. HAProxy (Load Balancer)"
    echo ""
    
    # Déploiement séquentiel
    deploy_postgres
    deploy_redis
    deploy_storage
    deploy_backend
    deploy_chat
    deploy_stream
    deploy_frontend
    deploy_haproxy
    
    # Vérification finale
    echo -e "${GREEN}🎉 Déploiement terminé !${NC}"
    echo ""
    echo -e "${BLUE}📊 État final des containers :${NC}"
    incus ls
    echo ""
    echo -e "${BLUE}🌐 Points d'accès :${NC}"
    echo -e "  • Application : ${YELLOW}http://10.100.0.16${NC} (HAProxy)"
    echo -e "  • HAProxy Stats : ${YELLOW}http://10.100.0.16:8404/stats${NC}"
    echo -e "  • Frontend Dev : ${YELLOW}http://10.100.0.11:5173${NC}"
    echo -e "  • Backend API : ${YELLOW}http://10.100.0.12:8080${NC}"
    echo -e "  • PostgreSQL : ${YELLOW}10.100.0.15:5432${NC}"
    echo -e "  • Redis : ${YELLOW}10.100.0.17:6379${NC}"
    echo ""
    echo -e "${CYAN}💡 Commandes utiles :${NC}"
    echo -e "  • État : ${YELLOW}incus ls${NC}"
    echo -e "  • Logs : ${YELLOW}incus info <container>${NC}"
    echo -e "  • Shell : ${YELLOW}incus exec <container> -- bash${NC}"
    echo -e "  • Arrêt : ${YELLOW}incus stop <container>${NC}"
    echo -e "  • Démarrage : ${YELLOW}incus start <container>${NC}"
}

# Vérifier qu'Incus est disponible
if ! command -v incus &> /dev/null; then
    echo -e "${RED}❌ Incus n'est pas installé ou accessible${NC}"
    exit 1
fi

# Vérifier que la configuration est faite
if ! incus network show veza-network >/dev/null 2>&1; then
    echo -e "${RED}❌ Configuration Incus manquante${NC}"
    echo -e "${YELLOW}💡 Exécutez d'abord : ./scripts/incus-setup.sh${NC}"
    exit 1
fi

# Exécuter le déploiement
main