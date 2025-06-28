#!/bin/bash

# Script de déploiement des containers de base avec dépendances
# Crée les 8 containers avec toutes les dépendances installées, prêts pour l'export

set -e

# Configuration
WORKSPACE_DIR="$(cd "$(dirname "$0")/.." && pwd)"
IMAGE="images:debian/bookworm"

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

echo -e "${PURPLE}"
echo "╭──────────────────────────────────────────╮"
echo "│    🏗️ Déploiement Containers de Base     │"
echo "╰──────────────────────────────────────────╯"
echo -e "${NC}"

# Configuration des containers
declare -A CONTAINERS=(
    ["postgres"]="veza-postgres"
    ["redis"]="veza-redis"
    ["storage"]="veza-storage"
    ["backend"]="veza-backend"
    ["chat"]="veza-chat"
    ["stream"]="veza-stream"
    ["frontend"]="veza-frontend"
    ["haproxy"]="veza-haproxy"
)

# Vérifications préalables
check_requirements() {
    echo -e "${BLUE}🔍 Vérification des prérequis...${NC}"
    
    # Vérifier Incus
    if ! command -v incus &> /dev/null; then
        echo -e "${RED}❌ Incus n'est pas installé${NC}"
        exit 1
    fi
    
    if ! incus info >/dev/null 2>&1; then
        echo -e "${RED}❌ Incus n'est pas initialisé ou accessible${NC}"
        echo -e "${YELLOW}💡 Exécutez: sudo incus admin init${NC}"
        exit 1
    fi
    
    # Vérifier le réseau par défaut
    if ! incus network show incusbr0 >/dev/null 2>&1; then
        echo -e "${RED}❌ Réseau par défaut incusbr0 introuvable${NC}"
        echo -e "${YELLOW}💡 Exécutez: sudo incus admin init${NC}"
        exit 1
    fi
    
    # Vérifier l'image
    if ! incus image list | grep -q "debian/bookworm"; then
        echo -e "${BLUE}📥 Téléchargement de l'image Debian Bookworm...${NC}"
        incus image copy images:debian/bookworm local: --alias debian/bookworm || {
            echo -e "${RED}❌ Impossible de télécharger l'image${NC}"
            exit 1
        }
    fi
    
    echo -e "${GREEN}✅ Prérequis vérifiés${NC}"
}

# Fonction de création sécurisée d'un container
create_container_safe() {
    local container_name=$1
    local timeout=60
    
    echo -e "${BLUE}🚀 Création du container $container_name...${NC}"
    
    # Vérifier si le container existe déjà
    if incus list "$container_name" --format csv | grep -q "$container_name"; then
        local status
        status=$(incus list "$container_name" --format csv | cut -d, -f2)
        echo -e "${YELLOW}⚠️ Container $container_name existe déjà (état: $status)${NC}"
        
        if [ "$status" = "RUNNING" ]; then
            echo -e "${CYAN}ℹ️ Container déjà running, arrêt pour réinitialisation...${NC}"
            incus stop "$container_name" --force || true
            sleep 2
        fi
        
        echo -e "${CYAN}🗑️ Suppression du container existant...${NC}"
        incus delete "$container_name" --force || true
        sleep 2
    fi
    
    # Créer le container avec le profil default uniquement
    echo -e "${CYAN}📦 Lancement de $container_name avec profil default...${NC}"
    incus launch "$IMAGE" "$container_name" --profile default
    
    # Attendre que le container soit complètement démarré
    echo -e "${BLUE}⏳ Attente du démarrage complet de $container_name...${NC}"
    local attempt=0
    while [ $attempt -lt $timeout ]; do
        if incus exec "$container_name" -- test -f /etc/hostname >/dev/null 2>&1; then
            echo -e "${GREEN}✅ Container $container_name prêt${NC}"
            return 0
        fi
        sleep 2
        ((attempt += 2))
        if [ $((attempt % 10)) -eq 0 ]; then
            echo -e "${CYAN}⏳ Attente en cours... ${attempt}s/${timeout}s${NC}"
        fi
    done
    
    echo -e "${RED}❌ Timeout: $container_name non prêt après ${timeout}s${NC}"
    return 1
}

# Fonction d'installation des dépendances de base (améliorée)
install_base_deps() {
    local container=$1
    echo -e "${BLUE}📦 Installation dépendances de base pour $container...${NC}"
    
    incus exec "$container" -- bash -c '
        export DEBIAN_FRONTEND=noninteractive
        
        echo "📥 Mise à jour des paquets..."
        for i in {1..3}; do
            if apt-get update -qq; then
                break
            else
                echo "Retry apt update (tentative $i/3)"
                sleep 5
            fi
        done
        
        echo "📦 Installation des outils de base..."
        apt-get install -y \
            curl wget git \
            build-essential ca-certificates gnupg lsb-release \
            systemd systemd-sysv init-system-helpers \
            procps net-tools dnsutils iputils-ping \
            netcat-openbsd htop vim nano \
            software-properties-common apt-transport-https \
            unzip zip rsync
            
        # Configuration DNS simple et efficace
        echo "🌐 Configuration DNS..."
        cat > /etc/resolv.conf << EOF
nameserver 8.8.8.8
nameserver 8.8.4.4
nameserver 1.1.1.1
EOF
        chmod 644 /etc/resolv.conf
        
        # Test DNS
        timeout 5 nslookup google.com 8.8.8.8 >/dev/null 2>&1 && echo "✅ DNS fonctionnel" || echo "⚠️ DNS limité"
        
        # Nettoyage
        apt-get clean
        rm -rf /var/lib/apt/lists/*
        
        echo "✅ Dépendances de base installées pour '"$container"'"
    '
}

# Déploiement PostgreSQL
deploy_postgres() {
    local container="${CONTAINERS[postgres]}"
    echo -e "${CYAN}🐘 Déploiement PostgreSQL...${NC}"
    
    create_container_safe "$container" || return 1
    install_base_deps "$container"
    
    incus exec "$container" -- bash -c '
        export DEBIAN_FRONTEND=noninteractive
        
        # Installation PostgreSQL
        apt-get update -qq
        apt-get install -y postgresql postgresql-contrib postgresql-client
        
        # Configuration de base
        systemctl enable postgresql
        
        # Créer un utilisateur et base de données par défaut
        sudo -u postgres createdb veza_db || true
        sudo -u postgres psql -c "CREATE USER veza_user WITH PASSWORD '\''veza_password'\'';" || true
        sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE veza_db TO veza_user;" || true
        
        # Configuration pour accepter les connexions réseau
        echo "listen_addresses = '\''*'\''" >> /etc/postgresql/15/main/postgresql.conf
        echo "host all all 0.0.0.0/0 md5" >> /etc/postgresql/15/main/pg_hba.conf
        
        apt-get clean
    '
    
    echo -e "${GREEN}✅ PostgreSQL configuré${NC}"
}

# Déploiement Redis
deploy_redis() {
    local container="${CONTAINERS[redis]}"
    echo -e "${CYAN}🔴 Déploiement Redis...${NC}"
    
    create_container_safe "$container" || return 1
    install_base_deps "$container"
    
    incus exec "$container" -- bash -c '
        export DEBIAN_FRONTEND=noninteractive
        
        # Installation Redis
        apt-get update -qq
        apt-get install -y redis-server
        
        # Configuration pour accepter les connexions réseau
        sed -i "s/bind 127.0.0.1/bind 0.0.0.0/" /etc/redis/redis.conf
        sed -i "s/protected-mode yes/protected-mode no/" /etc/redis/redis.conf
        
        systemctl enable redis-server
        apt-get clean
    '
    
    echo -e "${GREEN}✅ Redis configuré${NC}"
}

# Déploiement Storage (NFS)
deploy_storage() {
    local container="${CONTAINERS[storage]}"
    echo -e "${CYAN}🗄️ Déploiement Storage...${NC}"
    
    create_container_safe "$container" || return 1
    install_base_deps "$container"
    
    incus exec "$container" -- bash -c '
        export DEBIAN_FRONTEND=noninteractive
        
        # Installation NFS
        apt-get update -qq
        apt-get install -y nfs-kernel-server nfs-common
        
        # Création des répertoires de stockage
        mkdir -p /storage/{uploads,audio,backups,cache}
        chown -R nobody:nogroup /storage
        chmod -R 755 /storage
        
        # Configuration NFS exports
        cat > /etc/exports << EOF
/storage/uploads *(rw,sync,no_subtree_check,no_root_squash)
/storage/audio *(rw,sync,no_subtree_check,no_root_squash)
/storage/backups *(rw,sync,no_subtree_check,no_root_squash)
/storage/cache *(rw,sync,no_subtree_check,no_root_squash)
EOF
        
        systemctl enable nfs-kernel-server
        apt-get clean
    '
    
    echo -e "${GREEN}✅ Storage NFS configuré${NC}"
}

# Déploiement Backend (Go)
deploy_backend() {
    local container="${CONTAINERS[backend]}"
    echo -e "${CYAN}🔧 Déploiement Backend...${NC}"
    
    create_container_safe "$container" || return 1
    install_base_deps "$container"
    
    incus exec "$container" -- bash -c '
        export DEBIAN_FRONTEND=noninteractive
        
        # Installation Go
        wget -q https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
        tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
        rm go1.21.5.linux-amd64.tar.gz
        
        # Configuration Go
        echo "export PATH=/usr/local/go/bin:\$PATH" >> /etc/profile
        echo "export GOPATH=/app/go" >> /etc/profile
        
        # Installation client NFS
        apt-get update -qq
        apt-get install -y nfs-common
        
        # Création des répertoires de travail
        mkdir -p /app/{backend,uploads,logs}
        
        apt-get clean
    '
    
    echo -e "${GREEN}✅ Backend (Go) configuré${NC}"
}

# Déploiement Chat (Rust)
deploy_chat() {
    local container="${CONTAINERS[chat]}"
    echo -e "${CYAN}💬 Déploiement Chat...${NC}"
    
    create_container_safe "$container" || return 1
    install_base_deps "$container"
    
    incus exec "$container" -- bash -c '
        export DEBIAN_FRONTEND=noninteractive
        
        # Installation des dépendances de compilation Rust
        apt-get update -qq
        apt-get install -y pkg-config libssl-dev libpq-dev
        
        # Installation Rust
        curl --proto "=https" --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y --default-toolchain stable
        
        # Configuration Rust
        echo "export PATH=/root/.cargo/bin:\$PATH" >> /etc/profile
        source /root/.cargo/env
        
        # Création des répertoires de travail
        mkdir -p /app/{chat,logs}
        
        apt-get clean
    '
    
    echo -e "${GREEN}✅ Chat (Rust) configuré${NC}"
}

# Déploiement Stream (Rust)
deploy_stream() {
    local container="${CONTAINERS[stream]}"
    echo -e "${CYAN}🎵 Déploiement Stream...${NC}"
    
    create_container_safe "$container" || return 1
    install_base_deps "$container"
    
    incus exec "$container" -- bash -c '
        export DEBIAN_FRONTEND=noninteractive
        
        # Installation des dépendances de compilation Rust + audio
        apt-get update -qq
        apt-get install -y pkg-config libssl-dev libpq-dev ffmpeg nfs-common
        
        # Installation Rust
        curl --proto "=https" --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y --default-toolchain stable
        
        # Configuration Rust
        echo "export PATH=/root/.cargo/bin:\$PATH" >> /etc/profile
        source /root/.cargo/env
        
        # Création des répertoires de travail
        mkdir -p /app/{stream,logs}
        mkdir -p /storage/audio
        
        apt-get clean
    '
    
    echo -e "${GREEN}✅ Stream (Rust) configuré${NC}"
}

# Déploiement Frontend (Node.js)
deploy_frontend() {
    local container="${CONTAINERS[frontend]}"
    echo -e "${CYAN}⚛️ Déploiement Frontend...${NC}"
    
    create_container_safe "$container" || return 1
    install_base_deps "$container"
    
    incus exec "$container" -- bash -c '
        export DEBIAN_FRONTEND=noninteractive
        
        # Installation Node.js 20
        curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
        apt-get install -y nodejs
        
        # Installation des outils globaux
        npm install -g npm@latest
        
        # Création des répertoires de travail
        mkdir -p /app/{frontend,logs}
        
        apt-get clean
    '
    
    echo -e "${GREEN}✅ Frontend (Node.js) configuré${NC}"
}

# Déploiement HAProxy
deploy_haproxy() {
    local container="${CONTAINERS[haproxy]}"
    echo -e "${CYAN}⚖️ Déploiement HAProxy...${NC}"
    
    create_container_safe "$container" || return 1
    install_base_deps "$container"
    
    incus exec "$container" -- bash -c '
        export DEBIAN_FRONTEND=noninteractive
        
        # Installation HAProxy
        apt-get update -qq
        apt-get install -y haproxy
        
        # Configuration de base HAProxy
        systemctl enable haproxy
        
        # Création du répertoire de configuration
        mkdir -p /app/logs
        
        apt-get clean
    '
    
    echo -e "${GREEN}✅ HAProxy configuré${NC}"
}

# Configuration simplifiée (utilise uniquement le profil default)
configure_system() {
    echo -e "${BLUE}⚙️ Configuration système...${NC}"
    echo -e "${GREEN}✅ Profil 'default' standard utilisé pour tous les containers${NC}"
    echo -e "${GREEN}✅ Réseau 'incusbr0' par défaut utilisé${NC}"
    echo -e "${GREEN}✅ Configuration simplifiée et robuste${NC}"
}

# Fonction principale
main() {
    echo -e "${BLUE}🚀 Démarrage du déploiement des containers de base...${NC}"
    
    # Vérifications et configuration
    check_requirements
    configure_system
    
    # Déploiement séquentiel des containers
    echo -e "${BLUE}📋 Déploiement des 8 containers de base...${NC}"
    
    deploy_postgres
    deploy_redis
    deploy_storage
    deploy_backend
    deploy_chat
    deploy_stream
    deploy_frontend
    deploy_haproxy
    
    echo -e "${GREEN}🎉 Déploiement des containers de base terminé !${NC}"
    echo ""
    echo -e "${CYAN}📊 État des containers:${NC}"
    incus ls
    echo ""
    echo -e "${BLUE}💡 Prochaines étapes:${NC}"
    echo -e "  1. Tester la connectivité: ${YELLOW}./scripts/veza-manager.sh status${NC}"
    echo -e "  2. Exporter les containers: ${YELLOW}./scripts/veza-manager.sh export${NC}"
    echo -e "  3. Déployer le code source: ${YELLOW}./scripts/veza-manager.sh update${NC}"
}

main "$@" 