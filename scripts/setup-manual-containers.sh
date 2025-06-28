#!/bin/bash

# Script de création manuelle des 8 containers Veza
# Optimisé pour le développement avec rsync et services systemd

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
echo "│    🏗️ Création Manuelle des Containers   │"
echo "╰──────────────────────────────────────────╯"
echo -e "${NC}"

# Configuration
WORKSPACE_DIR="$(pwd)"
IMAGE="images:debian/bookworm"

# Liste des containers avec leurs rôles
declare -A CONTAINERS=(
    ["veza-postgres"]="Base de données PostgreSQL"
    ["veza-redis"]="Cache Redis"
    ["veza-storage"]="Stockage NFS"
    ["veza-backend"]="API Backend Go"
    ["veza-chat"]="Serveur Chat Rust"
    ["veza-stream"]="Serveur Stream Rust"
    ["veza-frontend"]="Interface React"
    ["veza-haproxy"]="Load Balancer"
)

# Fonction de création sécurisée d'un container
create_container() {
    local name=$1
    local description=$2
    
    echo -e "${CYAN}🚀 Création de $name - $description${NC}"
    
    # Supprimer si existe déjà
    if incus list "$name" --format csv | grep -q "$name"; then
        echo -e "${YELLOW}⚠️ $name existe déjà, suppression...${NC}"
        incus stop "$name" --force 2>/dev/null || true
        incus delete "$name" --force 2>/dev/null || true
        sleep 2
    fi
    
    # Créer le container avec profil default
    incus launch "$IMAGE" "$name" --profile default
    
    # Attendre qu'il soit prêt
    echo -e "${BLUE}⏳ Attente de $name...${NC}"
    local timeout=60
    local elapsed=0
    while [ $elapsed -lt $timeout ]; do
        if incus exec "$name" -- test -f /etc/hostname >/dev/null 2>&1; then
            echo -e "${GREEN}✅ $name prêt${NC}"
            return 0
        fi
        sleep 2
        ((elapsed += 2))
    done
    
    echo -e "${RED}❌ Timeout pour $name${NC}"
    return 1
}

# Configuration de base commune
setup_base_container() {
    local name=$1
    
    echo -e "${BLUE}📦 Configuration de base pour $name...${NC}"
    
    incus exec "$name" -- bash -c '
        export DEBIAN_FRONTEND=noninteractive
        
        # Mise à jour et outils de base
        apt-get update -qq
        apt-get install -y \
            curl wget git \
            build-essential ca-certificates \
            systemd systemd-sysv \
            procps net-tools dnsutils iputils-ping \
            rsync vim nano htop \
            software-properties-common
        
        # Configuration DNS simple
        cat > /etc/resolv.conf << EOF
nameserver 8.8.8.8
nameserver 8.8.4.4
nameserver 1.1.1.1
EOF
        
        # Créer répertoires de travail
        mkdir -p /app/{logs,scripts}
        mkdir -p /opt/veza
        
        # Nettoyage
        apt-get clean
        rm -rf /var/lib/apt/lists/*
    '
}

# Configuration spécialisée PostgreSQL
setup_postgres() {
    echo -e "${CYAN}🐘 Configuration PostgreSQL...${NC}"
    
    incus exec veza-postgres -- bash -c '
        export DEBIAN_FRONTEND=noninteractive
        
        # Installation PostgreSQL
        apt-get update -qq
        apt-get install -y postgresql postgresql-contrib postgresql-client
        
        # Configuration PostgreSQL
        systemctl enable postgresql
        systemctl start postgresql
        
        # Créer base et utilisateur
        sudo -u postgres createdb veza_db || true
        sudo -u postgres psql -c "CREATE USER veza_user WITH PASSWORD '\''veza_password'\'';" || true
        sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE veza_db TO veza_user;" || true
        
        # Configuration réseau
        echo "listen_addresses = '\''*'\''" >> /etc/postgresql/15/main/postgresql.conf
        echo "host all all 0.0.0.0/0 md5" >> /etc/postgresql/15/main/pg_hba.conf
        
        systemctl restart postgresql
    '
}

# Configuration spécialisée Redis
setup_redis() {
    echo -e "${CYAN}🔴 Configuration Redis...${NC}"
    
    incus exec veza-redis -- bash -c '
        export DEBIAN_FRONTEND=noninteractive
        
        # Installation Redis
        apt-get update -qq
        apt-get install -y redis-server
        
        # Configuration réseau
        sed -i "s/bind 127.0.0.1/bind 0.0.0.0/" /etc/redis/redis.conf
        sed -i "s/protected-mode yes/protected-mode no/" /etc/redis/redis.conf
        
        systemctl enable redis-server
        systemctl restart redis-server
    '
}

# Configuration spécialisée Storage
setup_storage() {
    echo -e "${CYAN}🗄️ Configuration Storage...${NC}"
    
    incus exec veza-storage -- bash -c '
        export DEBIAN_FRONTEND=noninteractive
        
        # Installation NFS
        apt-get update -qq
        apt-get install -y nfs-kernel-server nfs-common
        
        # Créer répertoires de stockage
        mkdir -p /storage/{uploads,audio,backups,cache}
        chown -R nobody:nogroup /storage
        chmod -R 755 /storage
        
        # Configuration NFS
        cat > /etc/exports << EOF
/storage/uploads *(rw,sync,no_subtree_check,no_root_squash)
/storage/audio *(rw,sync,no_subtree_check,no_root_squash)
/storage/backups *(rw,sync,no_subtree_check,no_root_squash)
/storage/cache *(rw,sync,no_subtree_check,no_root_squash)
EOF
        
        systemctl enable nfs-kernel-server
        systemctl restart nfs-kernel-server
        exportfs -ra
    '
}

# Configuration spécialisée Backend Go
setup_backend() {
    echo -e "${CYAN}🔧 Configuration Backend Go...${NC}"
    
    incus exec veza-backend -- bash -c '
        export DEBIAN_FRONTEND=noninteractive
        
        # Installation Go
        wget -q https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
        tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
        rm go1.21.5.linux-amd64.tar.gz
        
        # Configuration Go
        echo "export PATH=/usr/local/go/bin:\$PATH" >> /etc/profile
        echo "export GOPATH=/opt/veza/go" >> /etc/profile
        
        # Répertoires de travail
        mkdir -p /opt/veza/{backend,go}
        
        # Variables d'\''environnement pour le service
        cat > /etc/environment << EOF
PATH=/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
GOPATH=/opt/veza/go
DATABASE_URL=postgres://veza_user:veza_password@veza-postgres:5432/veza_db?sslmode=disable
REDIS_URL=redis://veza-redis:6379
JWT_SECRET=your-super-secret-jwt-key-change-in-production
PORT=8080
EOF
    '
}

# Configuration spécialisée Chat Rust
setup_chat() {
    echo -e "${CYAN}💬 Configuration Chat Rust...${NC}"
    
    incus exec veza-chat -- bash -c '
        export DEBIAN_FRONTEND=noninteractive
        
        # Dépendances Rust
        apt-get update -qq
        apt-get install -y pkg-config libssl-dev libpq-dev
        
        # Installation Rust
        curl --proto "=https" --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y --default-toolchain stable
        
        # Configuration Rust
        echo "export PATH=/root/.cargo/bin:\$PATH" >> /etc/profile
        source ~/.cargo/env
        
        # Répertoires de travail
        mkdir -p /opt/veza/chat
        
        # Variables d'\''environnement
        cat > /etc/environment << EOF
PATH=/root/.cargo/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
DATABASE_URL=postgres://veza_user:veza_password@veza-postgres:5432/veza_db?sslmode=disable
REDIS_URL=redis://veza-redis:6379
PORT=8081
RUST_LOG=info
EOF
    '
}

# Configuration spécialisée Stream Rust
setup_stream() {
    echo -e "${CYAN}🎵 Configuration Stream Rust...${NC}"
    
    incus exec veza-stream -- bash -c '
        export DEBIAN_FRONTEND=noninteractive
        
        # Dépendances Rust + Audio
        apt-get update -qq
        apt-get install -y pkg-config libssl-dev libpq-dev ffmpeg nfs-common
        
        # Installation Rust
        curl --proto "=https" --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y --default-toolchain stable
        
        # Configuration Rust
        echo "export PATH=/root/.cargo/bin:\$PATH" >> /etc/profile
        source ~/.cargo/env
        
        # Répertoires de travail
        mkdir -p /opt/veza/stream
        mkdir -p /storage/audio
        
        # Variables d'\''environnement
        cat > /etc/environment << EOF
PATH=/root/.cargo/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
DATABASE_URL=postgres://veza_user:veza_password@veza-postgres:5432/veza_db?sslmode=disable
REDIS_URL=redis://veza-redis:6379
PORT=8082
AUDIO_DIR=/storage/audio
RUST_LOG=info
EOF
    '
}

# Configuration spécialisée Frontend React
setup_frontend() {
    echo -e "${CYAN}⚛️ Configuration Frontend React...${NC}"
    
    incus exec veza-frontend -- bash -c '
        export DEBIAN_FRONTEND=noninteractive
        
        # Installation Node.js 20
        curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
        apt-get install -y nodejs
        
        # Outils globaux
        npm install -g npm@latest
        
        # Répertoires de travail
        mkdir -p /opt/veza/frontend
        
        # Variables d'\''environnement
        cat > /etc/environment << EOF
PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
NODE_ENV=development
VITE_API_URL=http://veza-backend:8080/api/v1
VITE_WS_CHAT_URL=ws://10.5.191.108:3001/ws
VITE_WS_STREAM_URL=ws://veza-stream:8082/ws
EOF
    '
}

# Configuration spécialisée HAProxy
setup_haproxy() {
    echo -e "${CYAN}⚖️ Configuration HAProxy...${NC}"
    
    incus exec veza-haproxy -- bash -c '
        export DEBIAN_FRONTEND=noninteractive
        
        # Installation HAProxy
        apt-get update -qq
        apt-get install -y haproxy
        
        # Configuration de base
        systemctl enable haproxy
        
        # Répertoire de configuration
        mkdir -p /opt/veza/haproxy
    '
}

# Fonction principale
main() {
    echo -e "${CYAN}🚀 Création des 8 containers Veza...${NC}"
    echo ""
    
    # Vérifier Incus
    if ! command -v incus &> /dev/null; then
        echo -e "${RED}❌ Incus non installé${NC}"
        exit 1
    fi
    
    # Créer tous les containers
    for container in "${!CONTAINERS[@]}"; do
        create_container "$container" "${CONTAINERS[$container]}"
        setup_base_container "$container"
    done
    
    echo ""
    echo -e "${BLUE}⚙️ Configuration spécialisée de chaque container...${NC}"
    
    # Configurations spécialisées
    setup_postgres
    setup_redis
    setup_storage
    setup_backend
    setup_chat
    setup_stream
    setup_frontend
    setup_haproxy
    
    echo ""
    echo -e "${GREEN}🎉 Tous les containers sont créés et configurés !${NC}"
    echo ""
    echo -e "${CYAN}📊 État des containers:${NC}"
    incus ls
    echo ""
    echo -e "${BLUE}💡 Prochaines étapes:${NC}"
    echo -e "  1. Créer les services systemd: ${YELLOW}./scripts/setup-systemd-services.sh${NC}"
    echo -e "  2. Configurer rsync: ${YELLOW}./scripts/setup-rsync.sh${NC}"
    echo -e "  3. Script de sync rapide: ${YELLOW}./scripts/quick-sync.sh${NC}"
}

main "$@" 