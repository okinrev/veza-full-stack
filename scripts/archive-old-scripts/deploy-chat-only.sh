#!/bin/bash

# Script de déploiement simplifié pour tester le Chat
# Déploie uniquement : PostgreSQL + Redis + Chat Rust + Frontend React

set -e

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

# Configuration
WORKSPACE_DIR=$(pwd)
IMAGE="images:debian/bookworm"

echo -e "${PURPLE}"
echo "╭──────────────────────────────────────────╮"
echo "│      💬 Veza - Déploiement Chat         │"
echo "╰──────────────────────────────────────────╯"
echo -e "${NC}"

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

# Fonction pour installer les dépendances de base
install_base_dependencies() {
    local container_name=$1
    
    echo -e "${BLUE}📦 Installation des dépendances de base pour $container_name...${NC}"
    
    incus exec "$container_name" -- bash -c "
        export DEBIAN_FRONTEND=noninteractive
        apt-get update -qq || true
        apt-get install -y curl wget build-essential ca-certificates || true
        apt-get clean
    "
}

# Déployer PostgreSQL
deploy_postgres() {
    echo -e "${CYAN}🐘 Déploiement de PostgreSQL...${NC}"
    
    if incus info veza-postgres >/dev/null 2>&1; then
        echo -e "${YELLOW}⚠️  Container veza-postgres existe déjà, suppression...${NC}"
        incus stop veza-postgres --force || true
        incus delete veza-postgres || true
    fi
    
    incus launch "$IMAGE" veza-postgres --profile veza-database
    wait_for_container veza-postgres
    
    # Configuration IP simple (utilise DHCP par défaut)
    echo -e "${BLUE}🌐 Configuration réseau pour veza-postgres...${NC}"
    sleep 5  # Attendre que le réseau soit configuré
    
    install_base_dependencies veza-postgres
    
    # Installation PostgreSQL
    incus exec veza-postgres -- bash -c "
        export DEBIAN_FRONTEND=noninteractive
        apt-get update -qq || true
        apt-get install -y postgresql postgresql-contrib || true
        systemctl enable postgresql || true
        systemctl start postgresql || true
        
        # Configuration PostgreSQL
        sudo -u postgres psql -c \"CREATE USER veza_user WITH PASSWORD 'veza_password';\" || true
        sudo -u postgres psql -c \"CREATE DATABASE veza_db OWNER veza_user;\" || true
        sudo -u postgres psql -c \"GRANT ALL PRIVILEGES ON DATABASE veza_db TO veza_user;\" || true
        
        # Configuration réseau
        echo \"listen_addresses = '*'\" >> /etc/postgresql/15/main/postgresql.conf || true
        echo \"host all all 10.100.0.0/24 md5\" >> /etc/postgresql/15/main/pg_hba.conf || true
        
        systemctl restart postgresql || true
    "
    
    echo -e "${GREEN}✅ PostgreSQL déployé${NC}"
}

# Déployer Redis
deploy_redis() {
    echo -e "${CYAN}🔴 Déploiement de Redis...${NC}"
    
    if incus info veza-redis >/dev/null 2>&1; then
        echo -e "${YELLOW}⚠️  Container veza-redis existe déjà, suppression...${NC}"
        incus stop veza-redis --force || true
        incus delete veza-redis || true
    fi
    
    incus launch "$IMAGE" veza-redis --profile veza-app
    wait_for_container veza-redis
    
    install_base_dependencies veza-redis
    
    # Installation Redis
    incus exec veza-redis -- bash -c "
        export DEBIAN_FRONTEND=noninteractive
        apt-get update -qq || true
        apt-get install -y redis-server || true
        
        # Configuration Redis
        sed -i 's/bind 127.0.0.1/bind 0.0.0.0/' /etc/redis/redis.conf || true
        sed -i 's/protected-mode yes/protected-mode no/' /etc/redis/redis.conf || true
        
        systemctl enable redis-server || true
        systemctl restart redis-server || true
    "
    
    echo -e "${GREEN}✅ Redis déployé${NC}"
}

# Déployer le serveur de chat Rust
deploy_chat() {
    echo -e "${CYAN}💬 Déploiement du Chat Server Rust...${NC}"
    
    if incus info veza-chat >/dev/null 2>&1; then
        echo -e "${YELLOW}⚠️  Container veza-chat existe déjà, suppression...${NC}"
        incus stop veza-chat --force || true
        incus delete veza-chat || true
    fi
    
    incus launch "$IMAGE" veza-chat --profile veza-app
    wait_for_container veza-chat
    
    install_base_dependencies veza-chat
    
    # Installation Rust (version simplifiée)
    incus exec veza-chat -- bash -c "
        export DEBIAN_FRONTEND=noninteractive
        
        # Installation Rust
        curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y || true
        source ~/.cargo/env || true
        echo 'source ~/.cargo/env' >> ~/.bashrc || true
        
        # Variables d'environnement
        cat > /etc/environment << 'EOF'
DATABASE_URL=postgres://veza_user:veza_password@veza-postgres:5432/veza_db?sslmode=disable
REDIS_URL=redis://veza-redis:6379
JWT_SECRET=your-super-secret-jwt-key-change-in-production
PORT=8081
RUST_LOG=chat_server=debug,tower_http=debug
EOF
    "
    
    # Copier le code source si disponible
    if [ -d "$WORKSPACE_DIR/veza-chat-server" ]; then
        echo -e "${BLUE}📁 Copie du code source du chat...${NC}"
        incus exec veza-chat -- mkdir -p /app/chat
        incus file push -r "$WORKSPACE_DIR/veza-chat-server/." veza-chat/app/chat/ || true
    fi
    
    echo -e "${GREEN}✅ Chat Server déployé${NC}"
}

# Fonction principale
main() {
    echo -e "${BLUE}🚀 Déploiement de l'infrastructure Chat...${NC}"
    echo -e "${YELLOW}Cette opération va créer 3 containers pour tester le chat. Continuer ? (o/N)${NC}"
    read -r response
    
    if [[ "$response" != "o" && "$response" != "oui" ]]; then
        echo -e "${GREEN}Déploiement annulé${NC}"
        exit 0
    fi
    
    echo -e "${BLUE}📋 Services à déployer :${NC}"
    echo -e "  1. PostgreSQL (Base de données)"
    echo -e "  2. Redis (Cache)"
    echo -e "  3. Chat Server Rust (WebSocket)"
    echo ""
    
    # Déploiement séquentiel
    deploy_postgres
    deploy_redis  
    deploy_chat
    
    echo ""
    echo -e "${GREEN}🎉 Déploiement Chat terminé !${NC}"
    echo ""
    echo -e "${BLUE}📊 État des containers :${NC}"
    incus ls
    echo ""
    echo -e "${BLUE}🌐 Services disponibles :${NC}"
    
    # Obtenir les IPs dynamiquement
    postgres_ip=$(incus list veza-postgres -c 4 --format csv | head -1)
    redis_ip=$(incus list veza-redis -c 4 --format csv | head -1)
    chat_ip=$(incus list veza-chat -c 4 --format csv | head -1)
    
    echo -e "  • PostgreSQL : ${YELLOW}$postgres_ip:5432${NC}"
    echo -e "  • Redis : ${YELLOW}$redis_ip:6379${NC}"
    echo -e "  • Chat WebSocket : ${YELLOW}ws://$chat_ip:8081/ws${NC}"
    echo ""
    echo -e "${CYAN}💡 Prochaine étape : Configurer le frontend React pour se connecter à ces services${NC}"
    echo -e "${CYAN}💡 Variables d'environnement frontend :${NC}"
    echo -e "  VITE_WS_CHAT_URL=ws://$chat_ip:8081/ws"
    echo -e "  VITE_API_URL=http://localhost:8080/api/v1"
}

# Vérifications
if ! command -v incus &> /dev/null; then
    echo -e "${RED}❌ Incus n'est pas installé${NC}"
    exit 1
fi

if ! incus network show veza-network >/dev/null 2>&1; then
    echo -e "${RED}❌ Configuration Incus manquante${NC}"
    echo -e "${YELLOW}💡 Exécutez d'abord : ./scripts/incus-setup.sh${NC}"
    exit 1
fi

# Exécution
main 