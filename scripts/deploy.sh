#!/bin/bash

# ╭──────────────────────────────────────────────────────────────╮
# │               🚀 Veza - Déploiement Unifié                  │
# │          Script de déploiement complet et optimisé          │
# ╰──────────────────────────────────────────────────────────────╯

set -e

# Couleurs pour les messages
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

# Fonctions utilitaires
log() { echo -e "${BLUE}[$(date '+%H:%M:%S')]${NC} $1"; }
success() { echo -e "${GREEN}✅ $1${NC}"; }
warning() { echo -e "${YELLOW}⚠️  $1${NC}"; }
error() { echo -e "${RED}❌ $1${NC}"; exit 1; }
header() { 
    echo -e "${PURPLE}${BOLD}"
    echo "╭──────────────────────────────────────────────────────────────╮"
    echo "│ $1"
    echo "╰──────────────────────────────────────────────────────────────╯"
    echo -e "${NC}"
}

# Variables globales
WORKSPACE_DIR=$(pwd)
IMAGE="images:debian/bookworm"
DEPLOYMENT_MODE="production"
SKIP_SETUP=false
FORCE_REBUILD=false
COMMAND=""

# Configuration des IPs (utilise les IPs existantes des containers)
get_container_ip() {
    local container_name=$1
    incus list "$container_name" -c 4 --format csv | cut -d' ' -f1 | head -n1
}

# Fonction d'aide
show_help() {
    echo -e "${CYAN}${BOLD}Veza - Script de Déploiement Unifié${NC}"
    echo ""
    echo -e "${BLUE}Usage:${NC} $0 [OPTIONS] [COMMAND]"
    echo ""
    echo -e "${YELLOW}Commandes disponibles:${NC}"
    echo -e "  ${GREEN}setup${NC}       - Configuration initiale d'Incus uniquement"
    echo -e "  ${GREEN}deploy${NC}      - Déploiement complet (setup + all containers)"
    echo -e "  ${GREEN}apps${NC}        - Déployer seulement les applications"
    echo -e "  ${GREEN}infrastructure${NC} - Déployer seulement l'infrastructure (DB, Redis, Storage)"
    echo -e "  ${GREEN}rebuild${NC}     - Reconstruction complète avec nettoyage"
    echo -e "  ${GREEN}status${NC}      - Vérifier le statut de tous les containers"
    echo -e "  ${GREEN}test${NC}        - Tester le déploiement"
    echo -e "  ${GREEN}clean${NC}       - Nettoyer complètement l'environnement"
    echo ""
    echo -e "${YELLOW}Options:${NC}"
    echo -e "  ${GREEN}--dev${NC}           - Mode développement"
    echo -e "  ${GREEN}--production${NC}    - Mode production (défaut)"
    echo -e "  ${GREEN}--skip-setup${NC}    - Ignorer la configuration initiale"
    echo -e "  ${GREEN}--force${NC}         - Forcer la reconstruction"
    echo -e "  ${GREEN}--help${NC}          - Afficher cette aide"
    echo ""
    echo -e "${YELLOW}Exemples:${NC}"
    echo -e "  $0 deploy              # Déploiement complet en mode production"
    echo -e "  $0 deploy --dev        # Déploiement complet en mode développement"
    echo -e "  $0 apps --force        # Redéployer seulement les applications"
    echo -e "  $0 rebuild             # Reconstruction complète"
    echo ""
}

# Parser les arguments
parse_arguments() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --dev)
                DEPLOYMENT_MODE="development"
                shift
                ;;
            --production)
                DEPLOYMENT_MODE="production"
                shift
                ;;
            --skip-setup)
                SKIP_SETUP=true
                shift
                ;;
            --force)
                FORCE_REBUILD=true
                shift
                ;;
            --help|-h)
                show_help
                exit 0
                ;;
            setup|deploy|apps|infrastructure|rebuild|status|test|clean)
                COMMAND=$1
                shift
                ;;
            *)
                error "Option inconnue: $1. Utilisez --help pour voir les options disponibles."
                ;;
        esac
    done

    if [ -z "${COMMAND:-}" ]; then
        COMMAND="deploy"
    fi
}

# Vérifier les prérequis
check_prerequisites() {
    log "Vérification des prérequis..."
    
    if ! command -v incus &> /dev/null; then
        error "Incus n'est pas installé. Installez-le avec: sudo snap install incus --channel=latest/stable"
    fi
    
    success "Prérequis vérifiés"
}

# Configuration initiale
setup_incus() {
    if [ "$SKIP_SETUP" = true ]; then
        log "Configuration initiale ignorée (--skip-setup)"
        return 0
    fi
    
    header "🔧 Configuration Initiale d'Incus"
    
    # Utiliser le script de setup existant
    if [ -f "$WORKSPACE_DIR/scripts/incus-setup.sh" ]; then
        chmod +x "$WORKSPACE_DIR/scripts/incus-setup.sh"
        "$WORKSPACE_DIR/scripts/incus-setup.sh"
    else
        error "Script incus-setup.sh non trouvé"
    fi
    
    success "Configuration Incus terminée"
}

# Déployer l'infrastructure
deploy_infrastructure() {
    header "🏗️ Finalisation de l'Infrastructure"
    
    # Vérifier PostgreSQL
    if incus exec veza-postgres -- systemctl is-active postgresql >/dev/null 2>&1; then
        success "PostgreSQL déjà actif"
    else
        log "Configuration PostgreSQL..."
        configure_postgresql
    fi
    
    # Vérifier Redis
    if incus exec veza-redis -- systemctl is-active redis-server >/dev/null 2>&1; then
        success "Redis déjà actif"
    else
        log "Configuration Redis..."
        configure_redis
    fi
    
    # Configurer le stockage NFS
    log "Configuration du stockage NFS..."
    configure_storage
    
    success "Infrastructure configurée"
}

# Configurer PostgreSQL
configure_postgresql() {
    log "Configuration avancée de PostgreSQL..."
    
    incus exec veza-postgres -- bash -c "
        # Configuration PostgreSQL pour l'accès réseau
        echo \"listen_addresses = '*'\" >> /etc/postgresql/15/main/postgresql.conf || true
        echo \"host all all 10.5.0.0/16 md5\" >> /etc/postgresql/15/main/pg_hba.conf || true
        
        # Redémarrer PostgreSQL
        systemctl restart postgresql
        
        # Créer utilisateur et base si nécessaire
        sudo -u postgres psql -c \"CREATE USER veza_user WITH PASSWORD 'veza_password';\" 2>/dev/null || true
        sudo -u postgres psql -c \"CREATE DATABASE veza_db OWNER veza_user;\" 2>/dev/null || true
        sudo -u postgres psql -c \"GRANT ALL PRIVILEGES ON DATABASE veza_db TO veza_user;\" 2>/dev/null || true
    "
    
    success "PostgreSQL configuré"
}

# Configurer Redis
configure_redis() {
    log "Configuration avancée de Redis..."
    
    incus exec veza-redis -- bash -c "
        # Configuration Redis pour l'accès réseau
        sed -i 's/bind 127.0.0.1/bind 0.0.0.0/' /etc/redis/redis.conf
        sed -i 's/protected-mode yes/protected-mode no/' /etc/redis/redis.conf
        
        systemctl restart redis-server
    "
    
    success "Redis configuré"
}

# Configurer le stockage NFS
configure_storage() {
    log "Installation et configuration du stockage NFS..."
    
    incus exec veza-storage -- bash -c "
        export DEBIAN_FRONTEND=noninteractive
        
        # Installer NFS si pas déjà fait
        if ! command -v exportfs &> /dev/null; then
            apt-get update -qq
            apt-get install -y nfs-kernel-server
        fi
        
        # Créer les répertoires de stockage
        mkdir -p /storage/{uploads,audio,backups,cache}
        chown -R nobody:nogroup /storage
        chmod -R 755 /storage
        
        # Configuration des exports NFS
        cat > /etc/exports << 'EOF'
/storage/uploads 10.5.0.0/16(rw,sync,no_subtree_check,no_root_squash)
/storage/audio 10.5.0.0/16(rw,sync,no_subtree_check,no_root_squash)
/storage/backups 10.5.0.0/16(rw,sync,no_subtree_check,no_root_squash)
/storage/cache 10.5.0.0/16(rw,sync,no_subtree_check,no_root_squash)
EOF
        
        # Démarrer les services NFS
        systemctl enable nfs-kernel-server
        systemctl restart nfs-kernel-server
        exportfs -ra
    "
    
    success "Stockage NFS configuré"
}

# Déployer les applications
deploy_applications() {
    header "🚀 Déploiement des Applications"
    
    log "Déploiement Backend Go..."
    deploy_backend
    
    log "Déploiement Chat Server Rust..."
    deploy_chat
    
    log "Déploiement Stream Server Rust..."
    deploy_stream
    
    log "Déploiement Frontend React..."
    deploy_frontend
    
    success "Applications déployées"
}

# Déployer le backend Go
deploy_backend() {
    local backend_ip=$(get_container_ip veza-backend)
    local postgres_ip=$(get_container_ip veza-postgres)
    local redis_ip=$(get_container_ip veza-redis)
    local frontend_ip=$(get_container_ip veza-frontend)
    
    log "Installation des dépendances Go..."
    incus exec veza-backend -- bash -c "
        export DEBIAN_FRONTEND=noninteractive
        
        # Installer Go si pas déjà fait
        if ! command -v go &> /dev/null; then
            apt-get update -qq
            apt-get install -y curl wget git build-essential
            
            # Installation Go
            wget -q https://go.dev/dl/go1.21.6.linux-amd64.tar.gz
            tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz
            rm go1.21.6.linux-amd64.tar.gz
            echo 'export PATH=\$PATH:/usr/local/go/bin' >> /etc/profile
        fi
    "
    
    # Copier le code source si pas déjà fait
    if ! incus exec veza-backend -- test -d /app/veza-backend-api; then
        log "Copie du code source backend..."
        incus file push -r "$WORKSPACE_DIR/veza-backend-api/" veza-backend/app/
    fi
    
    # Créer la configuration .env
    log "Configuration de l'environnement backend..."
    cat > /tmp/backend.env << EOF
DATABASE_URL=postgres://veza_user:veza_password@${postgres_ip}:5432/veza_db?sslmode=disable
REDIS_URL=redis://${redis_ip}:6379
JWT_SECRET=veza_jwt_secret_key_2025
SERVER_PORT=8080
SERVER_HOST=0.0.0.0
CORS_ORIGINS=http://${frontend_ip}:5173,http://$(get_container_ip veza-haproxy)
LOG_LEVEL=info
ENVIRONMENT=$DEPLOYMENT_MODE
EOF
    
    incus file push /tmp/backend.env veza-backend/app/.env
    rm /tmp/backend.env
    
    # Build et installation du service
    log "Construction et installation du backend..."
    incus exec veza-backend -- bash -c "
        cd /app/veza-backend-api
        export PATH=\$PATH:/usr/local/go/bin
        
        # Télécharger et nettoyer les dépendances
        go mod download
        go mod tidy
        
        # Build avec la méthode validée
        go build -buildvcs=false -o ../veza-backend ./cmd/server/
        
        chmod +x ../veza-backend
    "
    
    # Créer le service systemd
    log "Création du service systemd..."
    incus exec veza-backend -- bash -c "
cat > /etc/systemd/system/veza-backend.service << 'EOF'
[Unit]
Description=Veza Backend API
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/app
Environment=PATH=/usr/local/go/bin:\$PATH
ExecStart=/app/veza-backend
Restart=always
RestartSec=5
EnvironmentFile=-/app/.env

[Install]
WantedBy=multi-user.target
EOF
"
    
    # Démarrer le service
    incus exec veza-backend -- systemctl daemon-reload
    incus exec veza-backend -- systemctl enable veza-backend
    incus exec veza-backend -- systemctl restart veza-backend
    
    # Vérifier le démarrage
    sleep 5
    if incus exec veza-backend -- systemctl is-active veza-backend >/dev/null 2>&1; then
        success "Backend Go déployé et actif ($backend_ip:8080)"
    else
        warning "Backend démarré mais vérification nécessaire"
    fi
}

# Déployer le chat server Rust
deploy_chat() {
    local chat_ip=$(get_container_ip veza-chat)
    
    log "Installation des dépendances Rust..."
    incus exec veza-chat -- bash -c "
        export DEBIAN_FRONTEND=noninteractive
        
        # Installer Rust si pas déjà fait
        if ! command -v cargo &> /dev/null; then
            apt-get update -qq
            apt-get install -y curl build-essential
            
            # Installation Rust
            curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
            source ~/.cargo/env
        fi
    "
    
    # Copier le code source si pas déjà fait
    if ! incus exec veza-chat -- test -d /app/veza-chat-server; then
        log "Copie du code source chat server..."
        incus file push -r "$WORKSPACE_DIR/veza-chat-server/" veza-chat/app/
    fi
    
    # Build l'application
    log "Construction du chat server..."
    incus exec veza-chat -- bash -c "
        cd /app/veza-chat-server
        source ~/.cargo/env
        
        # Build en mode release
        cargo build --release || cargo build
        
        # Copier le binaire
        if [ -f target/release/veza-chat-server ]; then
            cp target/release/veza-chat-server ../veza-chat
        elif [ -f target/debug/veza-chat-server ]; then
            cp target/debug/veza-chat-server ../veza-chat
        fi
        
        chmod +x ../veza-chat
    "
    
    # Créer le service systemd
    incus exec veza-chat -- bash -c "
cat > /etc/systemd/system/veza-chat.service << 'EOF'
[Unit]
Description=Veza Chat Server
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/app
ExecStart=/app/veza-chat
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF
"
    
    # Démarrer le service
    incus exec veza-chat -- systemctl daemon-reload
    incus exec veza-chat -- systemctl enable veza-chat
    incus exec veza-chat -- systemctl restart veza-chat
    
    success "Chat Server Rust déployé ($chat_ip:8081)"
}

# Déployer le stream server Rust
deploy_stream() {
    local stream_ip=$(get_container_ip veza-stream)
    
    log "Installation des dépendances Rust pour stream..."
    incus exec veza-stream -- bash -c "
        export DEBIAN_FRONTEND=noninteractive
        
        # Installer Rust si pas déjà fait
        if ! command -v cargo &> /dev/null; then
            apt-get update -qq
            apt-get install -y curl build-essential
            
            # Installation Rust
            curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
            source ~/.cargo/env
        fi
    "
    
    # Copier le code source si pas déjà fait
    if ! incus exec veza-stream -- test -d /app/veza-stream-server; then
        log "Copie du code source stream server..."
        incus file push -r "$WORKSPACE_DIR/veza-stream-server/" veza-stream/app/
    fi
    
    # Build l'application
    log "Construction du stream server..."
    incus exec veza-stream -- bash -c "
        cd /app/veza-stream-server
        source ~/.cargo/env
        
        # Build en mode release
        cargo build --release || cargo build
        
        # Copier le binaire
        if [ -f target/release/veza-stream-server ]; then
            cp target/release/veza-stream-server ../veza-stream
        elif [ -f target/debug/veza-stream-server ]; then
            cp target/debug/veza-stream-server ../veza-stream
        fi
        
        chmod +x ../veza-stream
    "
    
    # Créer le service systemd
    incus exec veza-stream -- bash -c "
cat > /etc/systemd/system/veza-stream.service << 'EOF'
[Unit]
Description=Veza Stream Server
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/app
ExecStart=/app/veza-stream
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF
"
    
    # Démarrer le service
    incus exec veza-stream -- systemctl daemon-reload
    incus exec veza-stream -- systemctl enable veza-stream
    incus exec veza-stream -- systemctl restart veza-stream
    
    success "Stream Server Rust déployé ($stream_ip:8082)"
}

# Déployer le frontend React
deploy_frontend() {
    local frontend_ip=$(get_container_ip veza-frontend)
    local backend_ip=$(get_container_ip veza-backend)
    local chat_ip=$(get_container_ip veza-chat)
    local stream_ip=$(get_container_ip veza-stream)
    
    log "Installation de Node.js..."
    incus exec veza-frontend -- bash -c "
        export DEBIAN_FRONTEND=noninteractive
        
        # Installer Node.js si pas déjà fait
        if ! command -v node &> /dev/null; then
            apt-get update -qq
            apt-get install -y curl
            
            # Installation Node.js LTS
            curl -fsSL https://deb.nodesource.com/setup_lts.x | bash -
            apt-get install -y nodejs
        fi
    "
    
    # Copier le code source si pas déjà fait
    if ! incus exec veza-frontend -- test -d /app/veza-frontend; then
        log "Copie du code source frontend..."
        incus file push -r "$WORKSPACE_DIR/veza-frontend/" veza-frontend/app/
    fi
    
    # Créer la configuration .env
    log "Configuration de l'environnement frontend..."
    cat > /tmp/frontend.env << EOF
VITE_API_URL=http://${backend_ip}:8080
VITE_WS_CHAT_URL=ws://${chat_ip}:8081
VITE_WS_STREAM_URL=ws://${stream_ip}:8082
VITE_APP_NAME=Veza
VITE_APP_VERSION=1.0.0
VITE_ENVIRONMENT=$DEPLOYMENT_MODE
EOF
    
    incus file push /tmp/frontend.env veza-frontend/app/.env
    rm /tmp/frontend.env
    
    # Installation des dépendances et build
    log "Installation des dépendances et build du frontend..."
    incus exec veza-frontend -- bash -c "
        cd /app/veza-frontend
        
        # Installer les dépendances
        npm install
        
        # Build pour la production ou démarrer en dev
        if [ '$DEPLOYMENT_MODE' = 'production' ]; then
            npm run build
            
            # Installer un serveur web simple
            npm install -g serve
            
            # Créer le service pour servir les fichiers
            cat > /etc/systemd/system/veza-frontend.service << 'EOFSERVICE'
[Unit]
Description=Veza Frontend Server
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/app/veza-frontend
ExecStart=/usr/local/bin/serve -s dist -l 5173
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOFSERVICE
        else
            # Mode développement - serveur de dev
            cat > /etc/systemd/system/veza-frontend.service << 'EOFSERVICE'
[Unit]
Description=Veza Frontend Dev Server
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/app/veza-frontend
ExecStart=/usr/bin/npm run dev -- --host 0.0.0.0
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOFSERVICE
        fi
    "
    
    # Démarrer le service
    incus exec veza-frontend -- systemctl daemon-reload
    incus exec veza-frontend -- systemctl enable veza-frontend
    incus exec veza-frontend -- systemctl restart veza-frontend
    
    success "Frontend React déployé ($frontend_ip:5173)"
}

# Vérifier le statut du déploiement
check_deployment_status() {
    header "📊 Statut du Déploiement"
    
    local containers=("veza-postgres" "veza-redis" "veza-storage" "veza-backend" "veza-chat" "veza-stream" "veza-frontend" "veza-haproxy")
    
    for container in "${containers[@]}"; do
        if incus info "$container" >/dev/null 2>&1; then
            local ip=$(get_container_ip "$container")
            local status=$(incus list "$container" -c s --format csv)
            
            if [ "$status" = "RUNNING" ]; then
                success "$container: RUNNING ($ip)"
            else
                warning "$container: $status ($ip)"
            fi
        else
            error "$container: NON TROUVÉ"
        fi
    done
}

# Nettoyage complet
clean_deployment() {
    header "🧹 Nettoyage Complet"
    
    warning "Ceci va supprimer tous les containers et profils Veza"
    read -p "Continuer ? (y/N) " confirm && [ "$confirm" = "y" ] || exit 1
    
    # Arrêter et supprimer les containers
    local containers=("veza-haproxy" "veza-frontend" "veza-stream" "veza-chat" "veza-backend" "veza-storage" "veza-redis" "veza-postgres")
    
    for container in "${containers[@]}"; do
        log "Suppression de $container..."
        incus stop "$container" 2>/dev/null || true
        incus delete "$container" 2>/dev/null || true
    done
    
    # Supprimer les profils
    incus profile delete veza-storage veza-database veza-app veza-base 2>/dev/null || true
    
    # Supprimer le réseau
    incus network delete veza-network 2>/dev/null || true
    
    success "Nettoyage terminé"
}

# Fonction principale
main() {
    header "🚀 Veza - Déploiement Unifié ($DEPLOYMENT_MODE)"
    
    parse_arguments "$@"
    check_prerequisites
    
    case $COMMAND in
        setup)
            setup_incus
            ;;
        infrastructure)
            deploy_infrastructure
            ;;
        apps)
            deploy_applications
            ;;
        deploy)
            setup_incus
            deploy_infrastructure
            deploy_applications
            check_deployment_status
            ;;
        rebuild)
            FORCE_REBUILD=true
            clean_deployment
            setup_incus
            deploy_infrastructure
            deploy_applications
            check_deployment_status
            ;;
        status)
            check_deployment_status
            ;;
        test)
            # Lancer les tests
            if [ -f "$WORKSPACE_DIR/scripts/test.sh" ]; then
                chmod +x "$WORKSPACE_DIR/scripts/test.sh"
                "$WORKSPACE_DIR/scripts/test.sh"
            else
                warning "Script de test non trouvé"
            fi
            ;;
        clean)
            clean_deployment
            ;;
        *)
            error "Commande inconnue: $COMMAND"
            ;;
    esac
    
    success "Opération $COMMAND terminée avec succès!"
}

# Exécution
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi 