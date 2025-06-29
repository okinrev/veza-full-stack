#!/bin/bash

# ================================================================================
# TALAS ADMIN - Script Principal d'Administration
# ================================================================================
# Script unifié pour administrer la plateforme Talas complète
# 4 modules : Backend Go, Frontend React, Chat Server Rust, Stream Server Rust

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
VERSION="1.0.0"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

# Configuration des services Talas
declare -A SERVICES=(
    ["backend"]="Backend API Go"
    ["frontend"]="Frontend React"
    ["chat"]="Chat Server Rust"
    ["stream"]="Stream Server Rust"
    ["postgres"]="PostgreSQL Database"
    ["redis"]="Redis Cache"
    ["haproxy"]="Load Balancer"
)

declare -A SERVICE_PORTS=(
    ["backend"]="8080"
    ["frontend"]="5173"
    ["chat"]="3001"
    ["stream"]="3002"
    ["postgres"]="5432"
    ["redis"]="6379"
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
    echo "│           🎯 TALAS ADMIN v${VERSION}                    │"
    echo "│      Administration Complète de la Plateforme      │"
    echo "│    Backend Go • Frontend React • Chat • Stream     │"
    echo "╰─────────────────────────────────────────────────────╯"
    echo -e "${NC}"
}

show_help() {
    show_header
    echo -e "${BOLD}COMMANDES PRINCIPALES:${NC}"
    echo ""
    echo -e "${GREEN}  setup${NC}          Configuration initiale complète"
    echo -e "${GREEN}  start${NC}          Démarrer tous les services"
    echo -e "${GREEN}  stop${NC}           Arrêter tous les services"
    echo -e "${GREEN}  status${NC}         Vérifier l'état de tous les services"
    echo -e "${GREEN}  logs${NC}           Voir les logs des services"
    echo -e "${GREEN}  test${NC}           Lancer les tests d'intégration"
    echo -e "${GREEN}  deploy${NC}         Déployer sur Incus (production)"
    echo -e "${GREEN}  clean${NC}          Nettoyer l'environnement"
    echo ""
    echo -e "${BOLD}SERVICES DISPONIBLES:${NC}"
    echo -e "  ${YELLOW}backend${NC}   - API Go principale (port 8080)"
    echo -e "  ${YELLOW}frontend${NC}  - Interface React (port 5173)"
    echo -e "  ${YELLOW}chat${NC}      - WebSocket Chat Rust (port 3001)"
    echo -e "  ${YELLOW}stream${NC}    - Audio Stream Rust (port 3002)"
    echo ""
}

# Chargement de la configuration
load_config() {
    log_info "Chargement de la configuration..."
    
    # Charger la configuration unifiée
    if [ -f "$PROJECT_ROOT/configs/env.unified" ]; then
        source "$PROJECT_ROOT/configs/env.unified"
        log_success "Configuration unifiée chargée"
    else
        log_warning "Configuration unifiée non trouvée, utilisation des valeurs par défaut"
    fi
    
    # Variables par défaut si non définies
    export JWT_SECRET="${JWT_SECRET:-talas_unified_jwt_secret_key_2025_microservices_secure_32chars_minimum}"
    export JWT_ISSUER="${JWT_ISSUER:-talas-platform}"
    export JWT_AUDIENCE="${JWT_AUDIENCE:-talas-services}"
    export ENVIRONMENT="${ENVIRONMENT:-development}"
}

# Configuration initiale complète
setup_system() {
    log_header "Configuration Initiale de Talas"
    
    load_config
    
    log_info "1. Configuration JWT unifiée..."
    setup_jwt_config
    
    log_success "Configuration initiale terminée avec succès !"
    log_info "Vous pouvez maintenant démarrer les services avec: ${0##*/} start"
}

# Configuration JWT unifiée
setup_jwt_config() {
    log_header "Configuration JWT Unifiée"
    
    # Créer le fichier de configuration JWT partagé
    cat > "$PROJECT_ROOT/configs/jwt.config" << EOF
# Configuration JWT Unifiée Talas
JWT_SECRET=$JWT_SECRET
JWT_ISSUER=$JWT_ISSUER
JWT_AUDIENCE=$JWT_AUDIENCE
JWT_ALGORITHM=HS256
JWT_EXPIRATION=24h
JWT_REFRESH_EXPIRATION=168h
EOF
    
    # Mettre à jour la configuration du Frontend React
    cat > "$PROJECT_ROOT/veza-frontend/.env.local" << EOF
VITE_API_URL=http://localhost:8080
VITE_WS_CHAT_URL=ws://localhost:3001/ws
VITE_WS_STREAM_URL=ws://localhost:3002/ws
VITE_JWT_ISSUER=$JWT_ISSUER
VITE_JWT_AUDIENCE=$JWT_AUDIENCE
EOF
    
    log_success "Configuration JWT unifiée terminée"
}

# Compilation de tous les services
build_all() {
    log_header "Compilation de tous les services"
    
    # Backend Go
    log_info "Compilation du Backend Go..."
    cd "$PROJECT_ROOT/veza-backend-api"
    go build -o bin/server ./cmd/server
    log_success "Backend Go compilé"
    
    # Chat Server Rust
    log_info "Compilation du Chat Server Rust..."
    cd "$PROJECT_ROOT/veza-chat-server"
    cargo build --release
    log_success "Chat Server Rust compilé"
    
    # Stream Server Rust
    log_info "Compilation du Stream Server Rust..."
    cd "$PROJECT_ROOT/veza-stream-server"
    cargo build --release
    log_success "Stream Server Rust compilé"
    
    # Frontend React
    log_info "Compilation du Frontend React..."
    cd "$PROJECT_ROOT/veza-frontend"
    npm run build
    log_success "Frontend React compilé"
    
    cd "$PROJECT_ROOT"
    log_success "Tous les services compilés avec succès"
}

# Démarrage de tous les services
start_all() {
    log_header "Démarrage de tous les services"
    
    load_config
    
    # Créer le dossier logs
    mkdir -p "$PROJECT_ROOT/logs"
    
    # Démarrer Backend Go
    log_info "Démarrage du Backend Go..."
    cd "$PROJECT_ROOT/veza-backend-api"
    if [ -f "bin/server" ]; then
        nohup ./bin/server > "$PROJECT_ROOT/logs/backend.log" 2>&1 &
        echo $! > "$PROJECT_ROOT/logs/backend.pid"
        log_success "Backend Go démarré"
    else
        log_error "Exécutable Backend Go non trouvé. Exécutez d'abord: build"
        return 1
    fi
    
    # Démarrer Chat Server Rust
    log_info "Démarrage du Chat Server Rust..."
    cd "$PROJECT_ROOT/veza-chat-server"
    if [ -f "target/release/veza-chat-server" ]; then
        nohup ./target/release/veza-chat-server > "$PROJECT_ROOT/logs/chat.log" 2>&1 &
        echo $! > "$PROJECT_ROOT/logs/chat.pid"
        log_success "Chat Server Rust démarré"
    else
        log_error "Exécutable Chat Server non trouvé. Exécutez d'abord: build"
        return 1
    fi
    
    # Démarrer Stream Server Rust
    log_info "Démarrage du Stream Server Rust..."
    cd "$PROJECT_ROOT/veza-stream-server"
    if [ -f "target/release/veza-stream-server" ]; then
        nohup ./target/release/veza-stream-server > "$PROJECT_ROOT/logs/stream.log" 2>&1 &
        echo $! > "$PROJECT_ROOT/logs/stream.pid"
        log_success "Stream Server Rust démarré"
    else
        log_error "Exécutable Stream Server non trouvé. Exécutez d'abord: build"
        return 1
    fi
    
    # Démarrer Frontend React
    log_info "Démarrage du Frontend React..."
    cd "$PROJECT_ROOT/veza-frontend"
    nohup npm run dev -- --host 0.0.0.0 --port 5173 > "$PROJECT_ROOT/logs/frontend.log" 2>&1 &
    echo $! > "$PROJECT_ROOT/logs/frontend.pid"
    log_success "Frontend React démarré"
    
    cd "$PROJECT_ROOT"
    
    log_success "Tous les services démarrés avec succès !"
    log_info "Vérifiez l'état avec: ${0##*/} status"
}

# Arrêt de tous les services
stop_all() {
    log_header "Arrêt de tous les services"
    
    local pids=("backend" "chat" "stream" "frontend")
    
    for service in "${pids[@]}"; do
        if [ -f "$PROJECT_ROOT/logs/$service.pid" ]; then
            local pid=$(cat "$PROJECT_ROOT/logs/$service.pid")
            if kill -0 "$pid" 2>/dev/null; then
                log_info "Arrêt du service $service..."
                kill -TERM "$pid" 2>/dev/null || true
                sleep 2
                if kill -0 "$pid" 2>/dev/null; then
                    kill -KILL "$pid" 2>/dev/null || true
                fi
                log_success "Service $service arrêté"
            fi
            rm -f "$PROJECT_ROOT/logs/$service.pid"
        fi
    done
    
    log_success "Tous les services arrêtés"
}

# Vérification de l'état des services
check_status() {
    log_header "État des Services Talas"
    
    echo -e "${BOLD}Service           Port    Status${NC}"
    echo -e "${BOLD}─────────────────────────────────${NC}"
    
    for service in "${!SERVICES[@]}"; do
        local port="${SERVICE_PORTS[$service]}"
        local status="❌ Arrêté"
        
        if [ -f "$PROJECT_ROOT/logs/$service.pid" ]; then
            local service_pid=$(cat "$PROJECT_ROOT/logs/$service.pid")
            if kill -0 "$service_pid" 2>/dev/null; then
                status="✅ Actif"
            fi
        fi
        
        printf "%-17s %-7s %s\n" "${SERVICES[$service]}" "$port" "$status"
    done
}

# Affichage des logs
show_logs() {
    local service="$1"
    
    if [ -z "$service" ]; then
        log_error "Service non spécifié"
        log_info "Usage: ${0##*/} logs <service>"
        log_info "Services disponibles: ${!SERVICES[*]}"
        return 1
    fi
    
    if [ ! -f "$PROJECT_ROOT/logs/$service.log" ]; then
        log_error "Fichier de log non trouvé: $service.log"
        return 1
    fi
    
    log_info "Logs du service $service:"
    tail -n 50 "$PROJECT_ROOT/logs/$service.log"
}

# Tests d'intégration basiques
run_integration_tests() {
    log_header "Tests d'Intégration Talas"
    
    log_info "Test des endpoints de santé..."
    
    # Test Backend
    if curl -s -f http://localhost:8080/health >/dev/null 2>&1; then
        log_success "Backend API accessible"
    else
        log_warning "Backend API non accessible"
    fi
    
    # Test Chat
    if curl -s -f http://localhost:3001/health >/dev/null 2>&1; then
        log_success "Chat Server accessible"
    else
        log_warning "Chat Server non accessible"
    fi
    
    # Test Stream
    if curl -s -f http://localhost:3002/health >/dev/null 2>&1; then
        log_success "Stream Server accessible"
    else
        log_warning "Stream Server non accessible"
    fi
    
    log_success "Tests d'intégration terminés"
}

# Fonction principale
main() {
    case "$1" in
        "setup")
            setup_system
            ;;
        "start")
            start_all
            ;;
        "stop")
            stop_all
            ;;
        "status")
            check_status
            ;;
        "logs")
            show_logs "$2"
            ;;
        "test")
            run_integration_tests
            ;;
        "build")
            build_all
            ;;
        "restart")
            stop_all
            sleep 2
            start_all
            ;;
        "clean")
            stop_all
            rm -f logs/*.pid logs/*.log
            log_success "Environnement nettoyé"
            ;;
        "deploy")
            log_info "Fonctionnalité de déploiement à implémenter"
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
