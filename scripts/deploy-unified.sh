#!/bin/bash

# ╭──────────────────────────────────────────────────────────────╮
# │           🚀 Veza - Déploiement Unifié Corrigé             │
# │        Script de déploiement automatisé avec IPs fixes      │
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

# Configuration centralisée basée sur infrastructure.yaml
readonly WORKSPACE_DIR=$(pwd)

# IPs et configuration des containers (valeurs réelles)
declare -A CONTAINER_IPS=(
    ["veza-postgres"]="10.5.191.134"
    ["veza-redis"]="10.5.191.186"
    ["veza-storage"]="10.5.191.206"
    ["veza-backend"]="10.5.191.241"
    ["veza-chat"]="10.5.191.49"
    ["veza-stream"]="10.5.191.196"
    ["veza-frontend"]="10.5.191.41"
    ["veza-haproxy"]="10.5.191.133"
)

declare -A CONTAINER_PORTS=(
    ["veza-postgres"]="5432"
    ["veza-redis"]="6379"
    ["veza-storage"]="2049"
    ["veza-backend"]="8080"
    ["veza-chat"]="8081"
    ["veza-stream"]="8082"
    ["veza-frontend"]="5173"
    ["veza-haproxy"]="80"
)

# Variables globales
DEPLOYMENT_MODE="production"
SKIP_INFRASTRUCTURE=false
FORCE_RESTART=false

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

# Fonction d'aide
show_help() {
    echo -e "${CYAN}${BOLD}Veza - Script de Déploiement Unifié${NC}"
    echo ""
    echo -e "${BLUE}Usage:${NC} $0 [OPTIONS]"
    echo ""
    echo -e "${YELLOW}Options:${NC}"
    echo -e "  ${GREEN}--prod${NC}               - Mode production (défaut)"
    echo -e "  ${GREEN}--dev${NC}                - Mode développement"
    echo -e "  ${GREEN}--skip-infra${NC}         - Ignorer le déploiement infrastructure"
    echo -e "  ${GREEN}--force-restart${NC}      - Forcer le redémarrage des services"
    echo -e "  ${GREEN}--help${NC}               - Afficher cette aide"
    echo ""
    echo -e "${YELLOW}Exemples:${NC}"
    echo -e "  $0                       # Déploiement complet en mode production"
    echo -e "  $0 --dev                 # Déploiement en mode développement"
    echo -e "  $0 --skip-infra          # Déployer seulement les applications"
    echo ""
}

# Parser les arguments
parse_arguments() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --prod)
                DEPLOYMENT_MODE="production"
                shift
                ;;
            --dev)
                DEPLOYMENT_MODE="development"
                shift
                ;;
            --skip-infra)
                SKIP_INFRASTRUCTURE=true
                shift
                ;;
            --force-restart)
                FORCE_RESTART=true
                shift
                ;;
            --help|-h)
                show_help
                exit 0
                ;;
            *)
                error "Option inconnue: $1. Utilisez --help pour voir les options disponibles."
                ;;
        esac
    done
}

# Vérifier les prérequis
check_prerequisites() {
    log "Vérification des prérequis..."
    
    if ! command -v incus &> /dev/null; then
        error "Incus n'est pas installé. Installez-le avec: sudo snap install incus --channel=latest/stable"
    fi
    
    success "Prérequis vérifiés"
}

# Obtenir l'IP réelle d'un container
get_container_ip() {
    local container_name=$1
    local real_ip
    
    # Obtenir l'IP actuelle du container
    real_ip=$(incus list "$container_name" -c 4 --format csv | cut -d' ' -f1 | head -n1 2>/dev/null || echo "")
    
    if [[ -n "$real_ip" && "$real_ip" != "-" ]]; then
        echo "$real_ip"
    else
        # Retourner l'IP attendue si pas trouvée
        echo "${CONTAINER_IPS[$container_name]}"
    fi
}

# Vérifier si un container existe et est actif
check_container_status() {
    local container_name=$1
    
    if incus list --format csv | grep -q "^$container_name,RUNNING"; then
        return 0
    else
        return 1
    fi
}

# Configurer l'environnement backend avec les bonnes IPs
configure_backend_environment() {
    local postgres_ip="${CONTAINER_IPS[veza-postgres]}"
    local redis_ip="${CONTAINER_IPS[veza-redis]}"
    
    log "Configuration de l'environnement backend..."
    
    # Créer le fichier .env avec les bonnes IPs
    cat > /tmp/backend.env << EOF
DATABASE_URL=postgres://veza:veza_password@${postgres_ip}:5432/veza_db?sslmode=disable
REDIS_URL=redis://${redis_ip}:6379
JWT_SECRET=veza_jwt_secret_key_2025_production
SERVER_PORT=8080
SERVER_HOST=0.0.0.0
LOG_LEVEL=info
ENVIRONMENT=${DEPLOYMENT_MODE}
UPLOAD_PATH=/app/uploads
MAX_FILE_SIZE=10485760
EOF
    
    # Copier le fichier dans le container
    incus file push /tmp/backend.env veza-backend/app/veza-backend-api/.env
    rm /tmp/backend.env
    
    success "Environnement backend configuré"
}

# Configurer l'environnement chat server
configure_chat_environment() {
    local postgres_ip="${CONTAINER_IPS[veza-postgres]}"
    local redis_ip="${CONTAINER_IPS[veza-redis]}"
    
    log "Configuration de l'environnement chat server..."
    
    cat > /tmp/chat.env << EOF
# Veza Chat Server Configuration (Auto-generated)
DATABASE_URL=postgres://veza:veza_password@${postgres_ip}:5432/veza_db?sslmode=disable
REDIS_URL=redis://${redis_ip}:6379
JWT_SECRET=veza_jwt_secret_key_2025_production
RUST_LOG=chat_server=info,tower_http=info
BIND_ADDRESS=0.0.0.0:8081
EOF
    
    incus file push /tmp/chat.env veza-chat/app/veza-chat-server/.env
    rm /tmp/chat.env
    
    success "Environnement chat server configuré"
}

# Configurer l'environnement stream server
configure_stream_environment() {
    local postgres_ip="${CONTAINER_IPS[veza-postgres]}"
    local redis_ip="${CONTAINER_IPS[veza-redis]}"
    
    log "Configuration de l'environnement stream server..."
    
    cat > /tmp/stream.env << EOF
# Veza Stream Server Configuration (Auto-generated)
DATABASE_URL=postgres://veza:veza_password@${postgres_ip}:5432/veza_db?sslmode=disable
REDIS_URL=redis://${redis_ip}:6379
JWT_SECRET=veza_jwt_secret_key_2025_production
RUST_LOG=stream_server=info,tower_http=info
BIND_ADDRESS=0.0.0.0:8082
AUDIO_DIR=/storage/audio
EOF
    
    incus file push /tmp/stream.env veza-stream/app/veza-stream-server/.env
    rm /tmp/stream.env
    
    success "Environnement stream server configuré"
}

# Configurer l'environnement frontend
configure_frontend_environment() {
    local backend_ip="${CONTAINER_IPS[veza-backend]}"
    local chat_ip="${CONTAINER_IPS[veza-chat]}"
    local stream_ip="${CONTAINER_IPS[veza-stream]}"
    local haproxy_ip="${CONTAINER_IPS[veza-haproxy]}"
    
    log "Configuration de l'environnement frontend..."
    
    cat > /tmp/frontend.env << EOF
# Veza Frontend Configuration (Auto-generated)
VITE_API_URL=http://${haproxy_ip}/api/v1
VITE_WS_CHAT_URL=ws://${chat_ip}:8081/ws
VITE_WS_STREAM_URL=ws://${stream_ip}:8082/ws
VITE_APP_NAME=Veza
VITE_APP_VERSION=1.0.0
VITE_ENVIRONMENT=${DEPLOYMENT_MODE}
NODE_ENV=${DEPLOYMENT_MODE}
EOF
    
    incus file push /tmp/frontend.env veza-frontend/app/veza-frontend/.env
    rm /tmp/frontend.env
    
    success "Environnement frontend configuré"
}

# Configurer HAProxy avec les bonnes IPs
configure_haproxy() {
    local frontend_ip="${CONTAINER_IPS[veza-frontend]}"
    local backend_ip="${CONTAINER_IPS[veza-backend]}"
    local chat_ip="${CONTAINER_IPS[veza-chat]}"
    local stream_ip="${CONTAINER_IPS[veza-stream]}"
    
    log "Configuration HAProxy avec les IPs réelles..."
    
    cat > /tmp/haproxy.cfg << EOF
global
    daemon
    maxconn 4096

defaults
    mode http
    timeout connect 5s
    timeout client 30s
    timeout server 30s

frontend veza_main
    bind *:80
    
    # Headers CORS
    http-response set-header Access-Control-Allow-Origin "*"
    http-response set-header Access-Control-Allow-Methods "GET,POST,OPTIONS,PUT,DELETE"
    http-response set-header Access-Control-Allow-Headers "Content-Type,Authorization"
    
    # Routage
    acl is_backend_api path_beg /api/
    acl is_chat_ws path_beg /chat-api/
    acl is_stream_ws path_beg /stream/
    
    use_backend go_backend if is_backend_api
    use_backend chat_backend if is_chat_ws
    use_backend stream_backend if is_stream_ws
    default_backend react_frontend

backend react_frontend
    server react1 ${frontend_ip}:5173 check

backend go_backend
    server go1 ${backend_ip}:8080 check

backend chat_backend
    http-request set-path %[path,regsub(^/chat-api,/api)]
    server chat1 ${chat_ip}:8081 check

backend stream_backend
    http-request set-path %[path,regsub(^/stream,/)]
    server stream1 ${stream_ip}:8082 check
EOF
    
    # Copier dans le container HAProxy
    incus file push /tmp/haproxy.cfg veza-haproxy/etc/haproxy/haproxy.cfg
    rm /tmp/haproxy.cfg
    
    # Redémarrer HAProxy
    incus exec veza-haproxy -- systemctl restart haproxy
    
    success "HAProxy configuré avec les bonnes IPs"
}

# Déployer et configurer le backend Go
deploy_backend() {
    header "🔧 Déploiement Backend Go"
    
    if ! check_container_status veza-backend; then
        error "Container veza-backend n'est pas actif"
    fi
    
    # Copier le code source
    log "Copie du code source backend..."
    incus exec veza-backend -- mkdir -p /app
    incus file push -r "$WORKSPACE_DIR/veza-backend-api/" veza-backend/app/
    
    # Configurer l'environnement
    configure_backend_environment
    
    # Build et installation
    log "Construction du backend Go..."
    incus exec veza-backend -- bash -c "
        cd /app/veza-backend-api
        export PATH=\$PATH:/usr/local/go/bin
        
        # Nettoyer les dépendances
        go mod tidy
        
        # Build
        go build -buildvcs=false -o ../veza-backend ./cmd/server/
        chmod +x ../veza-backend
        
        # Créer le service systemd
        cat > /etc/systemd/system/veza-backend.service << 'EOFSERVICE'
[Unit]
Description=Veza Backend API
After=network.target postgresql.service

[Service]
Type=simple
User=root
WorkingDirectory=/app/veza-backend-api
Environment=PATH=/usr/local/go/bin:\$PATH
ExecStart=/app/veza-backend
Restart=always
RestartSec=5
EnvironmentFile=/app/veza-backend-api/.env

[Install]
WantedBy=multi-user.target
EOFSERVICE
        
        # Activer et démarrer le service
        systemctl daemon-reload
        systemctl enable veza-backend
        systemctl restart veza-backend
    "
    
    success "Backend Go déployé et actif"
}

# Déployer et configurer le chat server
deploy_chat() {
    header "💬 Déploiement Chat Server Rust"
    
    if ! check_container_status veza-chat; then
        error "Container veza-chat n'est pas actif. Lancez d'abord l'infrastructure."
    fi
    
    # Copier le code source
    log "Copie du code source chat server..."
    incus exec veza-chat -- mkdir -p /app
    incus file push -r "$WORKSPACE_DIR/veza-chat-server/" veza-chat/app/
    
    # Configurer l'environnement
    configure_chat_environment
    
    # Build et installation
    log "Construction du chat server Rust..."
    incus exec veza-chat -- bash -c "
        cd /app/veza-chat-server
        source ~/.cargo/env || true
        
        # Build en mode release
        cargo build --release || cargo build
        
        # Copier le binaire
        if [ -f target/release/veza-chat-server ]; then
            cp target/release/veza-chat-server ../veza-chat
        elif [ -f target/debug/veza-chat-server ]; then
            cp target/debug/veza-chat-server ../veza-chat
        fi
        
        chmod +x ../veza-chat
        
        # Créer le service systemd
        cat > /etc/systemd/system/veza-chat.service << 'EOFSERVICE'
[Unit]
Description=Veza Chat Server
After=network.target postgresql.service

[Service]
Type=simple
User=root
WorkingDirectory=/app/veza-chat-server
ExecStart=/app/veza-chat
Restart=always
RestartSec=5
EnvironmentFile=/app/veza-chat-server/.env

[Install]
WantedBy=multi-user.target
EOFSERVICE
        
        # Activer et démarrer le service
        systemctl daemon-reload
        systemctl enable veza-chat
        systemctl restart veza-chat
    "
    
    success "Chat Server Rust déployé"
}

# Déployer et configurer le stream server
deploy_stream() {
    header "🎵 Déploiement Stream Server Rust"
    
    if ! check_container_status veza-stream; then
        error "Container veza-stream n'est pas actif. Lancez d'abord l'infrastructure."
    fi
    
    # Copier le code source
    log "Copie du code source stream server..."
    incus exec veza-stream -- mkdir -p /app
    incus file push -r "$WORKSPACE_DIR/veza-stream-server/" veza-stream/app/
    
    # Configurer l'environnement
    configure_stream_environment
    
    # Build et installation
    log "Construction du stream server Rust..."
    incus exec veza-stream -- bash -c "
        cd /app/veza-stream-server
        source ~/.cargo/env || true
        
        # Build en mode release
        cargo build --release || cargo build
        
        # Copier le binaire
        if [ -f target/release/veza-stream-server ]; then
            cp target/release/veza-stream-server ../veza-stream
        elif [ -f target/debug/veza-stream-server ]; then
            cp target/debug/veza-stream-server ../veza-stream
        fi
        
        chmod +x ../veza-stream
        
        # Créer le service systemd
        cat > /etc/systemd/system/veza-stream.service << 'EOFSERVICE'
[Unit]
Description=Veza Stream Server
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/app/veza-stream-server
ExecStart=/app/veza-stream
Restart=always
RestartSec=5
EnvironmentFile=/app/veza-stream-server/.env

[Install]
WantedBy=multi-user.target
EOFSERVICE
        
        # Activer et démarrer le service
        systemctl daemon-reload
        systemctl enable veza-stream
        systemctl restart veza-stream
    "
    
    success "Stream Server Rust déployé"
}

# Déployer et configurer le frontend
deploy_frontend() {
    header "⚛️ Déploiement Frontend React"
    
    if ! check_container_status veza-frontend; then
        error "Container veza-frontend n'est pas actif. Lancez d'abord l'infrastructure."
    fi
    
    # Copier le code source
    log "Copie du code source frontend..."
    incus exec veza-frontend -- mkdir -p /app
    incus file push -r "$WORKSPACE_DIR/veza-frontend/" veza-frontend/app/
    
    # Configurer l'environnement
    configure_frontend_environment
    
    # Installation et build
    log "Installation et construction du frontend..."
    incus exec veza-frontend -- bash -c "
        cd /app/veza-frontend
        
        # Installer les dépendances
        npm install
        
        if [ '$DEPLOYMENT_MODE' = 'production' ]; then
            # Build pour la production
            npm run build
            
            # Configurer Nginx pour servir les fichiers statiques
            cat > /etc/nginx/sites-available/veza-frontend << 'EOFNGINX'
server {
    listen 5173;
    server_name _;
    root /app/veza-frontend/dist;
    index index.html;

    location / {
        try_files \$uri \$uri/ /index.html;
    }

    location /api/ {
        proxy_pass http://HAPROXY_IP/api/;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }

    location /chat-api/ {
        proxy_pass http://HAPROXY_IP/chat-api/;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }

    location /stream/ {
        proxy_pass http://HAPROXY_IP/stream/;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
}
EOFNGINX

            # Remplacer le placeholder
            sed -i 's/HAPROXY_IP/${CONTAINER_IPS[veza-haproxy]}/g' /etc/nginx/sites-available/veza-frontend
            
            # Activer le site
            ln -sf /etc/nginx/sites-available/veza-frontend /etc/nginx/sites-enabled/
            rm -f /etc/nginx/sites-enabled/default
            
            # Redémarrer Nginx
            systemctl restart nginx
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
EnvironmentFile=/app/veza-frontend/.env

[Install]
WantedBy=multi-user.target
EOFSERVICE
            
            systemctl daemon-reload
            systemctl enable veza-frontend
            systemctl restart veza-frontend
        fi
    "
    
    success "Frontend React déployé"
}

# Tests de connectivité
test_connectivity() {
    header "🔍 Tests de Connectivité"
    
    # Test PostgreSQL
    log "Test PostgreSQL..."
    if incus exec veza-postgres -- pg_isready -h localhost -p 5432 >/dev/null 2>&1; then
        success "PostgreSQL accessible"
    else
        warning "PostgreSQL inaccessible"
    fi
    
    # Test Backend
    log "Test Backend API..."
    sleep 5
    if curl -s -o /dev/null "http://${CONTAINER_IPS[veza-backend]}:8080"; then
        success "Backend API accessible"
    else
        warning "Backend API inaccessible"
    fi
    
    # Test HAProxy
    log "Test HAProxy..."
    if curl -s -o /dev/null "http://${CONTAINER_IPS[veza-haproxy]}"; then
        success "HAProxy accessible"
    else
        warning "HAProxy inaccessible"
    fi
}

# Fonction principale
main() {
    header "🚀 Veza - Déploiement Unifié Corrigé"
    
    parse_arguments "$@"
    check_prerequisites
    
    # Vérifier que l'infrastructure est en place
    if ! $SKIP_INFRASTRUCTURE; then
        log "Vérification de l'infrastructure..."
        
        local infrastructure_ok=true
        for container in "${!CONTAINER_IPS[@]}"; do
            if ! check_container_status "$container"; then
                warning "Container $container non actif"
                infrastructure_ok=false
            fi
        done
        
        if ! $infrastructure_ok; then
            error "Infrastructure manquante. Lancez d'abord: ./scripts/incus-setup.sh && ./scripts/incus-deploy.sh"
        fi
    fi
    
    # Déploiement des applications
    deploy_backend
    deploy_chat
    deploy_stream
    deploy_frontend
    
    # Configuration HAProxy
    configure_haproxy
    
    # Tests finaux
    test_connectivity
    
    # Rapport final
    header "🎉 Déploiement Terminé !"
    
    echo -e "${CYAN}🌐 Points d'accès :${NC}"
    echo -e "  • Application : ${GREEN}http://${CONTAINER_IPS[veza-haproxy]}${NC}"
    echo -e "  • Backend API : ${GREEN}http://${CONTAINER_IPS[veza-backend]}:8080${NC}"
    echo ""
    success "✨ Application Veza opérationnelle !"
}

# Exécution
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi 