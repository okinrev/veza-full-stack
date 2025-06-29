#!/bin/bash
# 🚀 Script de Gestion Infrastructure veza avec Incus
# Gestion complète : création, export/import, sync, build, monitoring

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORKSPACE_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
INCUS_PROJECT="veza"
EXPORT_DIR="$WORKSPACE_DIR/incus-exports"
LOG_DIR="$WORKSPACE_DIR/logs"
NETWORK_NAME="veza-net"
NETWORK_SUBNET="10.10.10.1/24"

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

# Liste des containers et leurs configurations
declare -A CONTAINERS=(
    ["postgres"]="postgresql"
    ["redis"]="redis-server"
    ["backend"]="go-api"
    ["chat"]="rust-chat"
    ["stream"]="rust-stream"
    ["frontend"]="react-app"
    ["storage"]="nfs-server"
    ["haproxy"]="load-balancer"
)

# Ports des services
declare -A SERVICE_PORTS=(
    ["postgres"]="5432"
    ["redis"]="6379"
    ["backend"]="8080"
    ["chat"]="3001"
    ["stream"]="3002"
    ["frontend"]="3000"
    ["storage"]="2049"
    ["haproxy"]="80"
)

# Dossiers à synchroniser
declare -A SYNC_PATHS=(
    ["backend"]="veza-backend-api:/app"
    ["chat"]="veza-chat-server:/app"
    ["stream"]="veza-stream-server:/app"
    ["frontend"]="veza-frontend:/app"
)

# Création des dossiers nécessaires
mkdir -p "$EXPORT_DIR" "$LOG_DIR"

# Logging
log() { echo -e "${BLUE}[$(date '+%H:%M:%S')]${NC} $*"; }
success() { echo -e "${GREEN}[$(date '+%H:%M:%S')] ✅ $*${NC}"; }
error() { echo -e "${RED}[$(date '+%H:%M:%S')] ❌ $*${NC}"; exit 1; }
warning() { echo -e "${YELLOW}[$(date '+%H:%M:%S')] ⚠️  $*${NC}"; }

# Header
show_header() {
    echo -e "${PURPLE}"
    echo "╔═══════════════════════════════════════════════════╗"
    echo "║        🚀 veza Infrastructure Manager 🚀         ║"
    echo "╚═══════════════════════════════════════════════════╝"
    echo -e "${NC}"
}

# Menu principal
show_menu() {
    echo -e "${CYAN}Commandes disponibles:${NC}"
    echo "  setup       - Configuration initiale (première fois)"
    echo "  create      - Créer tous les containers de zéro"
    echo "  export      - Exporter les containers configurés"
    echo "  import      - Importer et démarrer depuis les exports"
    echo "  sync        - Synchroniser le code source"
    echo "  build       - Compiler le code dans les containers"
    echo "  restart     - Redémarrer les services"
    echo "  deploy      - Import + Sync + Build + Restart"
    echo "  status      - Vérifier l'état de tous les services"
    echo "  logs        - Voir les logs d'un service"
    echo "  shell       - Accéder à un container"
    echo "  stop        - Arrêter tous les containers"
    echo "  clean       - Nettoyer tout (containers + exports)"
    echo "  ips         - Afficher toutes les IPs"
}

# Vérifier les prérequis
check_requirements() {
    log "Vérification des prérequis..."
    
    if ! command -v incus &> /dev/null; then
        error "Incus n'est pas installé. Installez-le d'abord."
    fi
    
    if ! command -v rsync &> /dev/null; then
        error "rsync n'est pas installé. Installez-le d'abord."
    fi
    
    success "Prérequis OK"
}

# Créer le réseau
setup_network() {
    log "Configuration du réseau $NETWORK_NAME..."
    
    if ! incus network show "$NETWORK_NAME" &>/dev/null; then
        incus network create "$NETWORK_NAME" \
            ipv4.address="10.10.10.1/24" \
            ipv4.nat=true \
            ipv6.address=none \
            dns.domain=veza.local \
            raw.dnsmasq="server=8.8.8.8
server=8.8.4.4"
        success "Réseau créé"
    else
        log "Réseau déjà existant"
    fi
}

# Créer un container de base
create_base_container() {
    local name=$1
    local type=$2
    
    log "Création du container $name ($type)..."
    
    # Créer le container
    incus launch images:debian/12 "$name" --network "$NETWORK_NAME"
    
    # Attendre qu'il soit prêt
    sleep 5
    
    # Configuration DNS et installation des dépendances de base
    incus exec "$name" -- bash -c "
        # Configuration DNS
        echo 'nameserver 8.8.8.8' > /etc/resolv.conf
        echo 'nameserver 8.8.4.4' >> /etc/resolv.conf
        
        # Mise à jour et installation
        apt-get update
        apt-get install -y curl wget git vim systemd systemd-sysv \
                          htop net-tools iputils-ping dnsutils
    "
    
    # Installation spécifique par type
    case "$type" in
        "postgresql")
            install_postgresql "$name"
            ;;
        "redis-server")
            install_redis "$name"
            ;;
        "go-api")
            install_go "$name"
            ;;
        "rust-chat"|"rust-stream")
            install_rust "$name"
            ;;
        "react-app")
            install_node "$name"
            ;;
        "nfs-server")
            install_nfs "$name"
            ;;
        "load-balancer")
            install_haproxy "$name"
            ;;
    esac
    
    success "Container $name créé et configuré"
}

# Installations spécifiques
install_postgresql() {
    local container=$1
    incus exec "$container" -- bash -c "
        apt-get install -y postgresql postgresql-contrib
        
        # Configuration PostgreSQL
        sudo -u postgres psql <<EOF
CREATE USER veza WITH PASSWORD 'veza_password';
CREATE DATABASE veza_db OWNER veza;
GRANT ALL PRIVILEGES ON DATABASE veza_db TO veza;
EOF
        
        # Autoriser les connexions externes
        echo 'host all all 0.0.0.0/0 md5' >> /etc/postgresql/15/main/pg_hba.conf
        echo \"listen_addresses = '*'\" >> /etc/postgresql/15/main/postgresql.conf
        
        systemctl restart postgresql
    "
}

install_redis() {
    local container=$1
    incus exec "$container" -- bash -c "
        apt-get install -y redis-server
        
        # Configuration Redis
        sed -i 's/bind 127.0.0.1/bind 0.0.0.0/' /etc/redis/redis.conf
        sed -i 's/# requirepass/requirepass veza_redis_password/' /etc/redis/redis.conf
        
        systemctl restart redis-server
    "
}

install_go() {
    local container=$1
    incus exec "$container" -- bash -c "
        # Installer Go
        wget -q https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
        tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
        rm go1.21.5.linux-amd64.tar.gz
        
        # Configuration environnement
        echo 'export PATH=\$PATH:/usr/local/go/bin' >> /etc/profile
        echo 'export GOPATH=/root/go' >> /etc/profile
        
        # Dépendances build
        apt-get install -y build-essential
        
        # Créer le service systemd
        cat > /etc/systemd/system/veza-backend.service <<EOF
[Unit]
Description=veza Backend API
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/app
Environment=\"PATH=/usr/local/go/bin:/usr/bin:/bin\"
ExecStart=/app/veza-backend
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF
        
        systemctl daemon-reload
        systemctl enable veza-backend
    "
}

install_rust() {
    local container=$1
    incus exec "$container" -- bash -c "
        # Installer Rust
        curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
        source /root/.cargo/env
        
        # Dépendances build
        apt-get install -y build-essential pkg-config libssl-dev
        
        # Créer le service systemd
        local service_name='veza-chat'
        if [[ \$HOSTNAME == *'stream'* ]]; then
            service_name='veza-stream'
        fi
        
        cat > /etc/systemd/system/\${service_name}.service <<EOF
[Unit]
Description=veza \${service_name} Server
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/app
Environment=\"PATH=/root/.cargo/bin:/usr/bin:/bin\"
ExecStart=/app/target/release/\${service_name}
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF
        
        systemctl daemon-reload
        systemctl enable \${service_name}
    "
}

install_node() {
    local container=$1
    incus exec "$container" -- bash -c "
        # Installer Node.js
        curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
        apt-get install -y nodejs
        
        # Installer PM2
        npm install -g pm2
        
        # Créer le service systemd
        cat > /etc/systemd/system/veza-frontend.service <<EOF
[Unit]
Description=veza Frontend
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/app
ExecStart=/usr/bin/npm run preview -- --host 0.0.0.0 --port 3000
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF
        
        systemctl daemon-reload
        systemctl enable veza-frontend
    "
}

install_nfs() {
    local container=$1
    incus exec "$container" -- bash -c "
        apt-get install -y nfs-kernel-server
        
        # Configuration NFS
        mkdir -p /storage/uploads /storage/tracks
        chmod 777 /storage/uploads /storage/tracks
        
        echo '/storage *(rw,sync,no_subtree_check,no_root_squash)' > /etc/exports
        
        systemctl restart nfs-kernel-server
    "
}

install_haproxy() {
    local container=$1
    incus exec "$container" -- bash -c "
        apt-get install -y haproxy
        
        # Configuration sera mise à jour dynamiquement
        systemctl enable haproxy
    "
}

# Créer tous les containers
create_all_containers() {
    log "Création de tous les containers..."
    
    setup_network
    
    for name in "${!CONTAINERS[@]}"; do
        local container_name="veza-$name"
        if incus info "$container_name" &>/dev/null; then
            warning "Container $container_name existe déjà, skip..."
            continue
        fi
        create_base_container "$container_name" "${CONTAINERS[$name]}"
    done
    
    success "Tous les containers créés"
}

# Exporter les containers
export_containers() {
    log "Export des containers..."
    
    for name in "${!CONTAINERS[@]}"; do
        local container_name="veza-$name"
        local export_file="$EXPORT_DIR/${container_name}.tar.gz"
        
        if incus info "$container_name" &>/dev/null; then
            log "Export de $container_name..."
            incus stop "$container_name" --force 2>/dev/null || true
            incus export "$container_name" "$export_file" --optimized-storage
            success "Exporté: $export_file"
        else
            warning "Container $container_name n'existe pas"
        fi
    done
    
    success "Export terminé"
}

# Importer les containers
import_containers() {
    log "Import des containers..."
    
    setup_network
    
    for name in "${!CONTAINERS[@]}"; do
        local container_name="veza-$name"
        local export_file="$EXPORT_DIR/${container_name}.tar.gz"
        
        if [ -f "$export_file" ]; then
            if incus info "$container_name" &>/dev/null; then
                warning "Container $container_name existe déjà, suppression..."
                incus delete "$container_name" --force
            fi
            
            log "Import de $container_name..."
            incus import "$export_file" "$container_name"
            incus start "$container_name"
            success "Importé: $container_name"
        else
            error "Fichier d'export non trouvé: $export_file"
        fi
    done
    
    # Attendre que tous soient prêts
    sleep 10
    
    # Mettre à jour la configuration HAProxy avec les nouvelles IPs
    update_haproxy_config
    
    success "Import terminé"
}

# Synchroniser le code source
sync_code() {
    log "Synchronisation du code source..."
    
    for name in "${!SYNC_PATHS[@]}"; do
        local container_name="veza-$name"
        local sync_path="${SYNC_PATHS[$name]}"
        local source_dir="${sync_path%%:*}"
        local dest_dir="${sync_path##*:}"
        
        if [ -d "$WORKSPACE_DIR/$source_dir" ]; then
            log "Sync $source_dir → $container_name:$dest_dir"
            
            # Créer le dossier destination
            incus exec "$container_name" -- mkdir -p "$dest_dir"
            
            # Rsync avec exclusions
            rsync -avz --delete \
                --exclude 'node_modules' \
                --exclude 'target' \
                --exclude '.git' \
                --exclude '*.log' \
                --exclude '.env.local' \
                "$WORKSPACE_DIR/$source_dir/" \
                "root@$(get_container_ip $container_name):$dest_dir/"
                
            success "Sync $name terminé"
        else
            warning "Dossier source non trouvé: $source_dir"
        fi
    done
}

# Compiler le code
build_code() {
    log "Compilation du code..."
    
    # Backend Go
    log "Build Backend Go..."
    incus exec veza-backend -- bash -c "
        cd /app
        export PATH=\$PATH:/usr/local/go/bin
        go mod tidy
        go build -o veza-backend cmd/server/main.go
    " && success "Backend compilé"
    
    # Chat Server Rust
    log "Build Chat Server..."
    incus exec veza-chat -- bash -c "
        source /root/.cargo/env
        cd /app
        cargo build --release
    " && success "Chat Server compilé"
    
    # Stream Server Rust
    log "Build Stream Server..."
    incus exec veza-stream -- bash -c "
        source /root/.cargo/env
        cd /app
        cargo build --release
    " && success "Stream Server compilé"
    
    # Frontend React
    log "Build Frontend..."
    incus exec veza-frontend -- bash -c "
        cd /app
        npm install
        npm run build
    " && success "Frontend compilé"
}

# Redémarrer les services
restart_services() {
    log "Redémarrage des services..."
    
    incus exec veza-postgres -- systemctl restart postgresql
    incus exec veza-redis -- systemctl restart redis-server
    incus exec veza-backend -- systemctl restart veza-backend
    incus exec veza-chat -- systemctl restart veza-chat
    incus exec veza-stream -- systemctl restart veza-stream
    incus exec veza-frontend -- systemctl restart veza-frontend
    incus exec veza-haproxy -- systemctl restart haproxy
    
    success "Services redémarrés"
}

# Obtenir l'IP d'un container
get_container_ip() {
    local container=$1
    incus list "$container" --format csv | cut -d, -f3 | grep -E '^10\.' | head -1
}

# Mettre à jour la configuration HAProxy
update_haproxy_config() {
    log "Mise à jour configuration HAProxy..."
    
    # Récupérer les IPs
    local backend_ip=$(get_container_ip veza-backend)
    local chat_ip=$(get_container_ip veza-chat)
    local stream_ip=$(get_container_ip veza-stream)
    local frontend_ip=$(get_container_ip veza-frontend)
    
    # Créer la configuration
    cat > /tmp/haproxy.cfg <<EOF
global
    maxconn 4096
    log stdout local0

defaults
    mode http
    timeout connect 5000ms
    timeout client 50000ms
    timeout server 50000ms
    option httplog

frontend http_front
    bind *:80
    
    # Routes API
    acl is_api path_beg /api
    use_backend backend_api if is_api
    
    # Routes WebSocket Chat
    acl is_ws_chat path_beg /ws/chat
    use_backend chat_ws if is_ws_chat
    
    # Routes WebSocket Stream
    acl is_ws_stream path_beg /ws/stream
    use_backend stream_ws if is_ws_stream
    
    # Frontend par défaut
    default_backend frontend_web

backend backend_api
    server backend1 ${backend_ip}:8080 check

backend chat_ws
    server chat1 ${chat_ip}:3001 check

backend stream_ws
    server stream1 ${stream_ip}:3002 check

backend frontend_web
    server frontend1 ${frontend_ip}:3000 check
EOF
    
    # Copier la configuration
    incus file push /tmp/haproxy.cfg veza-haproxy/etc/haproxy/haproxy.cfg
    
    # Redémarrer HAProxy
    incus exec veza-haproxy -- systemctl restart haproxy
    
    success "HAProxy configuré"
}

# Vérifier l'état des services
check_status() {
    echo -e "${CYAN}═══════════════════════════════════════════════════${NC}"
    echo -e "${CYAN}           État de l'Infrastructure veza           ${NC}"
    echo -e "${CYAN}═══════════════════════════════════════════════════${NC}"
    
    # Tableau des containers
    printf "%-15s %-15s %-20s %-10s %-15s\n" "Container" "IP" "Service" "Port" "Status"
    echo "───────────────────────────────────────────────────────────────────────"
    
    for name in "${!CONTAINERS[@]}"; do
        local container_name="veza-$name"
        local port="${SERVICE_PORTS[$name]}"
        
        if incus info "$container_name" &>/dev/null; then
            local ip=$(get_container_ip "$container_name")
            local container_status=$(incus list "$container_name" -c s --format csv)
            
            # Vérifier le service
            local service_status="❌"
            case "$name" in
                "postgres")
                    incus exec "$container_name" -- systemctl is-active postgresql &>/dev/null && service_status="✅"
                    ;;
                "redis")
                    incus exec "$container_name" -- systemctl is-active redis-server &>/dev/null && service_status="✅"
                    ;;
                "backend"|"chat"|"stream"|"frontend")
                    incus exec "$container_name" -- systemctl is-active "veza-$name" &>/dev/null && service_status="✅"
                    ;;
                "haproxy")
                    incus exec "$container_name" -- systemctl is-active haproxy &>/dev/null && service_status="✅"
                    ;;
            esac
            
            printf "%-15s %-15s %-20s %-10s %-15s\n" \
                "$container_name" "$ip" "${CONTAINERS[$name]}" "$port" "$service_status"
        else
            printf "%-15s %-15s %-20s %-10s %-15s\n" \
                "$container_name" "N/A" "${CONTAINERS[$name]}" "$port" "❌ Missing"
        fi
    done
    
    echo ""
    
    # Tests de connectivité
    echo -e "${CYAN}Tests de connectivité:${NC}"
    
    # Test API Backend
    local backend_ip=$(get_container_ip veza-backend)
    if [ -n "$backend_ip" ]; then
        if curl -s "http://$backend_ip:8080/health" &>/dev/null; then
            success "API Backend accessible"
        else
            warning "API Backend inaccessible"
        fi
    fi
    
    # Test Frontend
    local frontend_ip=$(get_container_ip veza-frontend)
    if [ -n "$frontend_ip" ]; then
        if curl -s "http://$frontend_ip:3000" &>/dev/null; then
            success "Frontend accessible"
        else
            warning "Frontend inaccessible"
        fi
    fi
    
    # URLs d'accès
    local haproxy_ip=$(get_container_ip veza-haproxy)
    echo ""
    echo -e "${CYAN}URLs d'accès:${NC}"
    echo "  Application : http://$haproxy_ip"
    echo "  API directe : http://$backend_ip:8080"
    echo "  Frontend direct : http://$frontend_ip:3000"
}

# Voir les logs
show_logs() {
    local service=$1
    
    if [ -z "$service" ]; then
        echo "Usage: $0 logs <service>"
        echo "Services disponibles: postgres, redis, backend, chat, stream, frontend, haproxy"
        return 1
    fi
    
    local container_name="veza-$service"
    
    case "$service" in
        "postgres")
            incus exec "$container_name" -- journalctl -u postgresql -f
            ;;
        "redis")
            incus exec "$container_name" -- journalctl -u redis-server -f
            ;;
        "backend"|"chat"|"stream"|"frontend")
            incus exec "$container_name" -- journalctl -u "veza-$service" -f
            ;;
        "haproxy")
            incus exec "$container_name" -- journalctl -u haproxy -f
            ;;
        *)
            error "Service inconnu: $service"
            ;;
    esac
}

# Accéder à un container
shell_access() {
    local service=$1
    
    if [ -z "$service" ]; then
        echo "Usage: $0 shell <service>"
        echo "Services disponibles: postgres, redis, backend, chat, stream, frontend, haproxy, storage"
        return 1
    fi
    
    incus exec "veza-$service" -- bash
}

# Afficher toutes les IPs
show_ips() {
    echo -e "${CYAN}IPs des containers:${NC}"
    for name in "${!CONTAINERS[@]}"; do
        local container_name="veza-$name"
        local ip=$(get_container_ip "$container_name")
        printf "  %-15s : %s\n" "$container_name" "${ip:-N/A}"
    done
}

# Arrêter tous les containers
stop_all() {
    log "Arrêt de tous les containers..."
    for name in "${!CONTAINERS[@]}"; do
        incus stop "veza-$name" --force 2>/dev/null || true
    done
    success "Containers arrêtés"
}

# Nettoyer tout
clean_all() {
    warning "Cette action va supprimer tous les containers et exports!"
    read -p "Êtes-vous sûr? (oui/non): " confirm
    
    if [ "$confirm" != "oui" ]; then
        log "Annulé"
        return
    fi
    
    stop_all
    
    log "Suppression des containers..."
    for name in "${!CONTAINERS[@]}"; do
        incus delete "veza-$name" --force 2>/dev/null || true
    done
    
    log "Suppression du réseau..."
    incus network delete "$NETWORK_NAME" 2>/dev/null || true
    
    log "Suppression des exports..."
    rm -rf "$EXPORT_DIR"
    
    success "Nettoyage terminé"
}

# Déploiement complet
full_deploy() {
    log "Déploiement complet..."
    
    import_containers
    sync_code
    build_code
    restart_services
    check_status
    
    success "Déploiement terminé!"
}

# Configuration initiale
initial_setup() {
    log "Configuration initiale..."
    
    check_requirements
    create_all_containers
    
    # Sync initial du code
    sync_code
    
    # Build initial
    build_code
    
    # Configuration finale
    update_haproxy_config
    restart_services
    
    # Export pour réutilisation
    export_containers
    
    success "Configuration initiale terminée!"
    echo -e "${CYAN}Les containers sont exportés dans: $EXPORT_DIR${NC}"
    echo -e "${CYAN}Pour les prochains déploiements, utilisez: $0 deploy${NC}"
}

# Main
main() {
    show_header
    
    case "${1:-help}" in
        setup)
            initial_setup
            ;;
        create)
            create_all_containers
            ;;
        export)
            export_containers
            ;;
        import)
            import_containers
            ;;
        sync)
            sync_code
            ;;
        build)
            build_code
            ;;
        restart)
            restart_services
            ;;
        deploy)
            full_deploy
            ;;
        status)
            check_status
            ;;
        logs)
            show_logs "${2:-}"
            ;;
        shell)
            shell_access "${2:-}"
            ;;
        ips)
            show_ips
            ;;
        stop)
            stop_all
            ;;
        clean)
            clean_all
            ;;
        help|*)
            show_menu
            ;;
    esac
}

# Exécution
main "$@"
