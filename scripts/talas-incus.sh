#!/bin/bash

# ================================================================================
# TALAS INCUS - Administration Complète Infrastructure Containers
# ================================================================================
# Gestion complète de l'infrastructure Talas avec 8 containers Incus
# Fonctionnalités : déploiement, services, compilation, logs, debug, nettoyage

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
VERSION="1.0.0"

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

# Configuration des containers et services
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

declare -A SERVICES=(
    ["postgres"]="postgresql"
    ["redis"]="redis-server"
    ["storage"]="nfs-kernel-server"
    ["backend"]="veza-backend"
    ["chat"]="veza-chat"
    ["stream"]="veza-stream"
    ["frontend"]="veza-frontend"
    ["haproxy"]="haproxy"
)

declare -A SERVICE_PORTS=(
    ["postgres"]="5432"
    ["redis"]="6379"
    ["storage"]="2049"
    ["backend"]="8080"
    ["chat"]="3001"
    ["stream"]="3002"
    ["frontend"]="5173"
    ["haproxy"]="80"
)

# Fonctions utilitaires
log_info() { echo -e "${BLUE}ℹ️  $1${NC}"; }
log_success() { echo -e "${GREEN}✅ $1${NC}"; }
log_warning() { echo -e "${YELLOW}⚠️  $1${NC}"; }
log_error() { echo -e "${RED}❌ $1${NC}"; }
log_header() { echo -e "${PURPLE}${BOLD}🚀 $1${NC}"; }

show_header() {
    echo -e "${PURPLE}${BOLD}"
    echo "╭─────────────────────────────────────────────────────╮"
    echo "│         🏗️  TALAS INCUS ADMIN v${VERSION}              │"
    echo "│      Administration Infrastructure Containers       │"
    echo "│        8 Containers • Services • Debug • Logs      │"
    echo "╰─────────────────────────────────────────────────────╯"
    echo -e "${NC}"
}

show_help() {
    show_header
    echo -e "${BOLD}COMMANDES INFRASTRUCTURE:${NC}"
    echo ""
    echo -e "${GREEN}  deploy${NC}         Déployer les 8 containers avec dépendances"
    echo -e "${GREEN}  setup${NC}          Configurer les services systemd"
    echo -e "${GREEN}  compile${NC}        Compiler le code dans les containers"
    echo -e "${GREEN}  update${NC}         Rsync + compilation + restart automatique"
    echo -e "${GREEN}  quick [export]${NC} Import + update (workflow rapide)"
    echo ""
    echo -e "${BOLD}GESTION DES SERVICES:${NC}"
    echo ""
    echo -e "${GREEN}  start [service]${NC}    Démarrer service(s)"
    echo -e "${GREEN}  stop [service]${NC}     Arrêter service(s)"
    echo -e "${GREEN}  restart [service]${NC}  Redémarrer service(s)"
    echo -e "${GREEN}  status${NC}             État de tous les services"
    echo ""
    echo -e "${BOLD}MONITORING & DEBUG:${NC}"
    echo ""
    echo -e "${GREEN}  logs [service]${NC}     Voir les logs d'un service"
    echo -e "${GREEN}  health${NC}             Vérification de santé complète"
    echo -e "${GREEN}  debug${NC}              Mode debug avancé"
    echo -e "${GREEN}  network-fix${NC}        Réparer les problèmes réseau"
    echo ""
    echo -e "${BOLD}MAINTENANCE:${NC}"
    echo ""
    echo -e "${GREEN}  clean${NC}              Supprimer toute l'infrastructure"
    echo -e "${GREEN}  export${NC}             Exporter les containers configurés"
    echo -e "${GREEN}  import${NC}             Importer les containers"
    echo ""
    echo -e "${BOLD}CONTAINERS DISPONIBLES:${NC}"
    echo -e "  ${YELLOW}postgres${NC}  - Base de données (port 5432)"
    echo -e "  ${YELLOW}redis${NC}     - Cache mémoire (port 6379)"
    echo -e "  ${YELLOW}storage${NC}   - Stockage NFS (port 2049)"
    echo -e "  ${YELLOW}backend${NC}   - API Go (port 8080)"
    echo -e "  ${YELLOW}chat${NC}      - WebSocket Rust (port 3001)"
    echo -e "  ${YELLOW}stream${NC}    - Audio Rust (port 3002)"
    echo -e "  ${YELLOW}frontend${NC}  - React Dev (port 5173)"
    echo -e "  ${YELLOW}haproxy${NC}   - Load Balancer (port 80)"
    echo ""
}

# Vérifications préalables
check_requirements() {
    log_info "Vérification des prérequis..."
    
    if ! command -v incus &> /dev/null; then
        log_error "Incus n'est pas installé"
        exit 1
    fi
    
    if ! incus info >/dev/null 2>&1; then
        log_error "Incus n'est pas initialisé"
        exit 1
    fi
    
    log_success "Prérequis satisfaits"
}

# Déploiement complet des 8 containers
deploy_infrastructure() {
    log_header "Déploiement Infrastructure Complète (8 containers)"
    
    check_requirements
    
    log_info "Déploiement via deploy-base-containers.sh..."
    if [ -f "$SCRIPT_DIR/deploy-base-containers.sh" ]; then
        bash "$SCRIPT_DIR/deploy-base-containers.sh"
        log_success "Infrastructure déployée"
    else
        log_error "Script deploy-base-containers.sh non trouvé"
        return 1
    fi
}

# Configuration des services systemd
setup_services() {
    log_header "Configuration des Services Systemd"
    
    # Récupération des IPs des containers
    local POSTGRES_IP=$(incus ls veza-postgres --format csv --columns 4 | cut -d' ' -f1)
    local REDIS_IP=$(incus ls veza-redis --format csv --columns 4 | cut -d' ' -f1)
    
    log_info "IPs des services: PostgreSQL=$POSTGRES_IP, Redis=$REDIS_IP"
    
    # Backend Go
    log_info "Configuration service Backend Go..."
    incus exec veza-backend -- bash -c "
        mkdir -p /opt/veza/backend
        cat > /etc/systemd/system/veza-backend.service << 'EOF'
[Unit]
Description=Veza Backend API Server
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/veza/backend
ExecStart=/opt/veza/backend/bin/server
Restart=always
RestartSec=5
Environment=DATABASE_URL=postgres://veza_user:veza_password@$POSTGRES_IP:5432/veza_db
Environment=REDIS_URL=redis://$REDIS_IP:6379
Environment=JWT_SECRET=talas_jwt_secret_key
Environment=PORT=8080

[Install]
WantedBy=multi-user.target
EOF
        systemctl daemon-reload
        systemctl enable veza-backend
    "
    
    # Chat Server Rust
    log_info "Configuration service Chat Server..."
    incus exec veza-chat -- bash -c "
        mkdir -p /opt/veza/chat
        cat > /etc/systemd/system/veza-chat.service << 'EOF'
[Unit]
Description=Veza Chat WebSocket Server
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/app/chat/veza-chat-server
ExecStart=/app/chat/veza-chat-server/target/release/chat-server
Restart=always
RestartSec=5
Environment=PORT=3001
Environment=DATABASE_URL=postgres://veza_user:veza_password@$POSTGRES_IP:5432/veza_db
Environment=REDIS_URL=redis://$REDIS_IP:6379

[Install]
WantedBy=multi-user.target
EOF
        systemctl daemon-reload
        systemctl enable veza-chat
    "
    
    # Stream Server Rust
    log_info "Configuration service Stream Server..."
    incus exec veza-stream -- bash -c "
        mkdir -p /opt/veza/stream
        cat > /etc/systemd/system/veza-stream.service << 'EOF'
[Unit]
Description=Veza Stream Audio Server
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/app/stream/veza-stream-server
ExecStart=/app/stream/veza-stream-server/target/release/veza-stream-server
Restart=always
RestartSec=5
Environment=PORT=3002

[Install]
WantedBy=multi-user.target
EOF
        systemctl daemon-reload
        systemctl enable veza-stream
    "
    
    # Frontend React
    log_info "Configuration service Frontend..."
    incus exec veza-frontend -- bash -c "
        mkdir -p /opt/veza/frontend
        cat > /etc/systemd/system/veza-frontend.service << 'EOF'
[Unit]
Description=Veza Frontend Development Server
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/app/frontend/veza-frontend
ExecStart=/usr/bin/npm run dev -- --host 0.0.0.0 --port 5173
Restart=always
RestartSec=5
Environment=NODE_ENV=development

[Install]
WantedBy=multi-user.target
EOF
        systemctl daemon-reload
        systemctl enable veza-frontend
    "
    
    log_success "Services systemd configurés"
}

# Compilation dans les containers
compile_all() {
    log_header "Compilation dans les Containers"
    
    # Backend Go
    log_info "Compilation Backend Go..."
    incus exec veza-backend -- bash -c "
        cd /app/backend/veza-backend-api
        if [ -f go.mod ]; then
            mkdir -p /opt/veza/backend/bin
            export PATH=/usr/local/go/bin:\$PATH
            go build -o /opt/veza/backend/bin/server ./cmd/server
            echo '✅ Backend Go compilé'
        else
            echo '⚠️ Code source Backend non trouvé dans /app/backend/veza-backend-api'
            ls -la /app/backend/ || echo 'Répertoire vide'
        fi
    "
    
    # Chat Server Rust
    log_info "Compilation Chat Server Rust..."
    incus exec veza-chat -- bash -c "
        cd /app/chat/veza-chat-server
        if [ -f Cargo.toml ]; then
            export PATH=/root/.cargo/bin:\$PATH
            cargo build --release
            mkdir -p /opt/veza/chat/target/release
            cp target/release/* /opt/veza/chat/target/release/ 2>/dev/null || true
            echo '✅ Chat Server compilé'
        else
            echo '⚠️ Code source Chat non trouvé dans /app/chat/veza-chat-server'
            ls -la /app/chat/ || echo 'Répertoire vide'
        fi
    "
    
    # Stream Server Rust  
    log_info "Compilation Stream Server Rust..."
    incus exec veza-stream -- bash -c "
        cd /app/stream/veza-stream-server
        if [ -f Cargo.toml ]; then
            export PATH=/root/.cargo/bin:\$PATH
            cargo build --release
            mkdir -p /opt/veza/stream/target/release
            cp target/release/* /opt/veza/stream/target/release/ 2>/dev/null || true
            echo '✅ Stream Server compilé'
        else
            echo '⚠️ Code source Stream non trouvé'
            ls -la /app/stream || echo 'Répertoire stream vide'
        fi
    "
    
    # Frontend React
    log_info "Installation dépendances Frontend..."
    incus exec veza-frontend -- bash -c "
        cd /app/frontend/veza-frontend
        if [ -f package.json ]; then
            npm install
            echo '✅ Frontend préparé'
        else
            echo '⚠️ Code source Frontend non trouvé dans /app/frontend/veza-frontend'
            ls -la /app/frontend/ || echo 'Répertoire vide'
        fi
    "
    
    log_success "Compilation terminée dans tous les containers"
}

# Mise à jour complète (rsync + compile + restart)
update_and_restart() {
    log_header "Mise à Jour Complète (Rsync + Compile + Restart)"
    
    # Utiliser le script de mise à jour existant
    log_info "1. Synchronisation du code source..."
    if [ -f "$SCRIPT_DIR/update-source-code.sh" ]; then
        bash "$SCRIPT_DIR/update-source-code.sh" all
    else
        log_warning "Script update-source-code.sh non trouvé"
    fi
    
    # Compilation
    log_info "2. Compilation..."
    compile_all
    
    # Redémarrage des services
    log_info "3. Redémarrage des services..."
    restart_service "backend"
    restart_service "chat"
    restart_service "stream"
    restart_service "frontend"
    
    log_success "Mise à jour complète terminée"
}

# Gestion des services
start_service() {
    local service_name=$1
    if [ -z "$service_name" ]; then
        # Démarrer tous les services
        for service in "${!SERVICES[@]}"; do
            start_service "$service"
        done
        return
    fi
    
    local container="${CONTAINERS[$service_name]}"
    local service="${SERVICES[$service_name]}"
    
    log_info "Démarrage $service_name..."
    if incus exec "$container" -- systemctl start "$service"; then
        log_success "$service_name démarré"
    else
        log_error "Échec démarrage $service_name"
    fi
}

stop_service() {
    local service_name=$1
    if [ -z "$service_name" ]; then
        # Arrêter tous les services
        for service in "${!SERVICES[@]}"; do
            stop_service "$service"
        done
        return
    fi
    
    local container="${CONTAINERS[$service_name]}"
    local service="${SERVICES[$service_name]}"
    
    log_info "Arrêt $service_name..."
    if incus exec "$container" -- systemctl stop "$service"; then
        log_success "$service_name arrêté"
    else
        log_error "Échec arrêt $service_name"
    fi
}

# Alias pour arrêter tous les services
stop_all_services() {
    stop_service
}

restart_service() {
    local service_name=$1
    if [ -z "$service_name" ]; then
        # Redémarrer tous les services
        for service in "${!SERVICES[@]}"; do
            restart_service "$service"
        done
        return
    fi
    
    local container="${CONTAINERS[$service_name]}"
    local service="${SERVICES[$service_name]}"
    
    log_info "Redémarrage $service_name..."
    if incus exec "$container" -- systemctl restart "$service"; then
        log_success "$service_name redémarré"
    else
        log_error "Échec redémarrage $service_name"
    fi
}

# État des services
check_status() {
    log_header "État de l'Infrastructure Talas"
    
    echo -e "${BOLD}Container         Service           Port    Status${NC}"
    echo -e "${BOLD}────────────────────────────────────────────────────${NC}"
    
    for service_name in "${!CONTAINERS[@]}"; do
        local container="${CONTAINERS[$service_name]}"
        local service="${SERVICES[$service_name]}"
        local port="${SERVICE_PORTS[$service_name]}"
        
        # Vérifier si le container existe et est running
        if ! incus list "$container" --format csv | grep -q "RUNNING"; then
            printf "%-17s %-17s %-7s %s\n" "$container" "$service" "$port" "❌ Container arrêté"
            continue
        fi
        
        # Vérifier le service
        if incus exec "$container" -- systemctl is-active "$service" >/dev/null 2>&1; then
            printf "%-17s %-17s %-7s %s\n" "$container" "$service" "$port" "✅ Actif"
        else
            printf "%-17s %-17s %-7s %s\n" "$container" "$service" "$port" "❌ Inactif"
        fi
    done
    
    echo ""
    log_info "État des containers:"
    incus ls --format table --columns n,s,4
}

# Logs des services
show_logs() {
    local service_name=$1
    if [ -z "$service_name" ]; then
        log_error "Service non spécifié"
        log_info "Usage: $0 logs <service>"
        log_info "Services: ${!SERVICES[*]}"
        return 1
    fi
    
    local container="${CONTAINERS[$service_name]}"
    local service="${SERVICES[$service_name]}"
    
    log_info "Logs de $service_name (Ctrl+C pour quitter):"
    incus exec "$container" -- journalctl -u "$service" -f --no-pager
}

# Vérification de santé
health_check() {
    log_header "Vérification de Santé Complète"
    
    log_info "Test de connectivité inter-services..."
    
    # Test PostgreSQL
    if incus exec veza-backend -- nc -z localhost 5432 2>/dev/null; then
        log_success "PostgreSQL accessible"
    else
        log_warning "PostgreSQL non accessible"
    fi
    
    # Test Redis
    if incus exec veza-backend -- nc -z localhost 6379 2>/dev/null; then
        log_success "Redis accessible"
    else
        log_warning "Redis non accessible"
    fi
    
    # Test Backend API
    if incus exec veza-frontend -- curl -s -f http://localhost:8080/health >/dev/null 2>&1; then
        log_success "Backend API répond"
    else
        log_warning "Backend API ne répond pas"
    fi
    
    # Test ports des services
    for service_name in "${!SERVICE_PORTS[@]}"; do
        local container="${CONTAINERS[$service_name]}"
        local port="${SERVICE_PORTS[$service_name]}"
        
        if incus exec "$container" -- netstat -ln | grep -q ":$port "; then
            log_success "Port $port ($service_name) ouvert"
        else
            log_warning "Port $port ($service_name) fermé"
        fi
    done
    
    log_info "Utilisation des ressources:"
    incus info --resources
}

# Mode debug avancé
debug_mode() {
    log_header "Mode Debug Avancé"
    
    log_info "Diagnostic réseau..."
    incus network list
    
    log_info "État détaillé des containers..."
    for container in "${CONTAINERS[@]}"; do
        echo -e "\n${CYAN}=== $container ===${NC}"
        incus info "$container"
        echo -e "\n${YELLOW}Processus:${NC}"
        incus exec "$container" -- ps aux | head -10
        echo -e "\n${YELLOW}Mémoire:${NC}"
        incus exec "$container" -- free -h
        echo -e "\n${YELLOW}Réseau:${NC}"
        incus exec "$container" -- ip addr show
    done
}

# Réparation réseau
network_fix() {
    log_header "Réparation Réseau"
    
    if [ -f "$SCRIPT_DIR/network-fix.sh" ]; then
        bash "$SCRIPT_DIR/network-fix.sh"
    else
        log_error "Script network-fix.sh non trouvé"
    fi
}

# Nettoyage complet
clean_infrastructure() {
    log_header "Suppression Complète de l'Infrastructure"
    
    log_warning "Cette action va supprimer TOUS les containers Talas"
    echo -e "${YELLOW}Continuer ? (oui/non)${NC}"
    read -r response
    
    if [[ "$response" != "oui" ]]; then
        log_info "Opération annulée"
        return 0
    fi
    
    log_info "Arrêt et suppression des containers..."
    for container in "${CONTAINERS[@]}"; do
        if incus list "$container" --format csv | grep -q "$container"; then
            log_info "Suppression $container..."
            incus stop "$container" --force 2>/dev/null || true
            incus delete "$container" 2>/dev/null || true
        fi
    done
    
    log_success "Infrastructure complètement supprimée"
}

# Export des containers
export_containers() {
    log_header "Export des Containers Configurés"
    
    # Création du dossier d'export avec timestamp
    local export_dir="$PROJECT_ROOT/exports/$(date +%Y%m%d_%H%M%S)"
    mkdir -p "$export_dir"
    
    log_info "Arrêt des services avant export..."
    stop_all_services
    
    local exported_count=0
    for service_name in "${!CONTAINERS[@]}"; do
        local container="${CONTAINERS[$service_name]}"
        if incus list "$container" --format csv | grep -q "RUNNING\|STOPPED"; then
            log_info "Export $container ($service_name)..."
            if incus export "$container" "$export_dir/$container.tar.xz" --compression=xz; then
                ((exported_count++))
                log_success "✅ $container exporté"
            else
                log_error "❌ Échec export $container"
            fi
        else
            log_warning "Container $container non trouvé, ignoré"
        fi
    done
    
    # Création d'un fichier de métadonnées
    cat > "$export_dir/export_info.txt" << EOF
Export Talas Infrastructure
Date: $(date)
Containers exportés: $exported_count
Version: $VERSION
Status: $(if [ $exported_count -gt 0 ]; then echo "SUCCESS"; else echo "FAILED"; fi)
EOF
    
    # Création d'un lien symbolique vers le dernier export
    ln -sfn "$export_dir" "$PROJECT_ROOT/exports/latest"
    
    log_success "🎉 $exported_count containers exportés vers exports/$(basename "$export_dir")"
    log_info "💡 Accès rapide via: exports/latest/"
}

# Import des containers
import_containers() {
    log_header "Import des Containers"
    
    local import_dir=""
    
    # Si un argument est fourni, l'utiliser comme répertoire
    if [ -n "$2" ]; then
        import_dir="$PROJECT_ROOT/exports/$2"
    else
        # Sinon utiliser le dernier export
        import_dir="$PROJECT_ROOT/exports/latest"
    fi
    
    if [ ! -d "$import_dir" ]; then
        log_error "Dossier d'import non trouvé: $import_dir"
        log_info "💡 Exports disponibles:"
        if [ -d "$PROJECT_ROOT/exports" ]; then
            ls -la "$PROJECT_ROOT/exports/" | grep "^d" | awk '{print $9}' | grep -v "^\\.$\\|^\\.\\.\\|^latest"
        fi
        return 1
    fi
    
    log_info "Nettoyage des containers existants..."
    clean_infrastructure_silent
    
    local imported_count=0
    for export_file in "$import_dir"/*.tar.xz; do
        if [ -f "$export_file" ]; then
            local container_name=$(basename "$export_file" .tar.xz)
            log_info "Import $container_name..."
            if incus import "$export_file"; then
                incus start "$container_name"
                ((imported_count++))
                log_success "✅ $container_name importé et démarré"
            else
                log_error "❌ Échec import $container_name"
            fi
        fi
    done
    
    log_success "🎉 $imported_count containers importés"
    
    # Attendre que les containers soient prêts
    log_info "Attente stabilisation des services..."
    sleep 10
    
    check_status
}

# Nettoyage silencieux pour import
clean_infrastructure_silent() {
    for container in "${CONTAINERS[@]}"; do
        if incus list "$container" --format csv | grep -q "$container"; then
            incus stop "$container" --force 2>/dev/null || true
            incus delete "$container" 2>/dev/null || true
        fi
    done
}

# Workflow rapide : import + update
quick_deploy() {
    log_header "Déploiement Rapide (Import + Update)"
    
    import_containers "$@"
    
    if [ $? -eq 0 ]; then
        log_info "🔄 Mise à jour du code..."
        update_and_restart
    else
        log_error "Import échoué, déploiement rapide annulé"
        return 1
    fi
}

# Fonction principale
main() {
    case "$1" in
        "deploy")
            deploy_infrastructure
            ;;
        "setup")
            setup_services
            ;;
        "compile")
            compile_all
            ;;
        "update")
            update_and_restart
            ;;
        "quick")
            quick_deploy "$@"
            ;;
        "start")
            start_service "$2"
            ;;
        "stop")
            stop_service "$2"
            ;;
        "restart")
            restart_service "$2"
            ;;
        "status")
            check_status
            ;;
        "logs")
            show_logs "$2"
            ;;
        "health")
            health_check
            ;;
        "debug")
            debug_mode
            ;;
        "network-fix")
            network_fix
            ;;
        "clean")
            clean_infrastructure
            ;;
        "export")
            export_containers
            ;;
        "import")
            import_containers
            ;;
        "help"|"-h"|"--help"|"")
            show_help
            ;;
        *)
            log_error "Commande inconnue: $1"
            show_help
            exit 1
            ;;
    esac
}

# Exécution
main "$@" 