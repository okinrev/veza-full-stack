#!/bin/bash
# Script de synchronisation et build rapide pour développement (Version Incus)

set -euo pipefail

# Configuration
WORKSPACE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# Couleurs
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

# Fonction pour obtenir l'IP d'un container
get_container_ip() {
    incus list "$1" --format csv | cut -d, -f3 | grep -E '^10\.' | head -1
}

# Fonction pour vérifier si un container existe et est démarré
check_container() {
    local container=$1
    if ! incus list "$container" --format csv | grep -q "RUNNING"; then
        echo -e "${RED}❌ Container $container n'est pas en cours d'exécution${NC}"
        return 1
    fi
    return 0
}

# Fonction de sync et build pour un service
sync_and_build() {
    local service=$1
    local action=${2:-all} # sync, build, restart, ou all
    
    echo -e "${BLUE}[$service] Traitement...${NC}"
    
    case "$service" in
        backend)
            local container="veza-backend"
            check_container "$container" || return 1
            
            if [[ "$action" == "sync" || "$action" == "all" ]]; then
                echo "Synchronisation du backend..."
                
                # Créer le répertoire app s'il n'existe pas
                incus exec "$container" -- mkdir -p /app
                
                # Copier les fichiers Go
                echo "  → Copie des fichiers source..."
                incus file push -r "$WORKSPACE_DIR/veza-backend-api/cmd" "$container/app/" 2>/dev/null || true
                incus file push -r "$WORKSPACE_DIR/veza-backend-api/internal" "$container/app/" 2>/dev/null || true
                incus file push -r "$WORKSPACE_DIR/veza-backend-api/pkg" "$container/app/" 2>/dev/null || true
                incus file push "$WORKSPACE_DIR/veza-backend-api/go.mod" "$container/app/" 2>/dev/null || true
                incus file push "$WORKSPACE_DIR/veza-backend-api/go.sum" "$container/app/" 2>/dev/null || true
                
                # Copier la configuration
                if [[ -f "$WORKSPACE_DIR/configs/backend.env" ]]; then
                    echo "  → Copie de la configuration..."
                    incus file push "$WORKSPACE_DIR/configs/backend.env" "$container/app/.env" 2>/dev/null || true
                fi
            fi
            
            if [[ "$action" == "build" || "$action" == "all" ]]; then
                echo "Compilation du backend..."
                incus exec "$container" -- bash -c "cd /app && PATH=\$PATH:/usr/local/go/bin go mod tidy"
                incus exec "$container" -- bash -c "cd /app && PATH=\$PATH:/usr/local/go/bin go build -o veza-backend cmd/server/main.go"
            fi
            
            if [[ "$action" == "restart" || "$action" == "all" ]]; then
                echo "Redémarrage du backend..."
                incus exec "$container" -- systemctl restart veza-backend 2>/dev/null || echo "  ⚠️  Service veza-backend non configuré"
            fi
            ;;
            
        chat)
            local container="veza-chat"
            check_container "$container" || return 1
            
            if [[ "$action" == "sync" || "$action" == "all" ]]; then
                echo "Synchronisation du chat server..."
                
                # Créer le répertoire app s'il n'existe pas
                incus exec "$container" -- mkdir -p /app
                
                # Copier les fichiers Rust
                echo "  → Copie des fichiers source..."
                incus file push -r "$WORKSPACE_DIR/veza-chat-server/src" "$container/app/" 2>/dev/null || true
                incus file push "$WORKSPACE_DIR/veza-chat-server/Cargo.toml" "$container/app/" 2>/dev/null || true
                incus file push "$WORKSPACE_DIR/veza-chat-server/Cargo.lock" "$container/app/" 2>/dev/null || true
                
                # Copier la configuration
                if [[ -f "$WORKSPACE_DIR/configs/chat.env" ]]; then
                    echo "  → Copie de la configuration..."
                    incus file push "$WORKSPACE_DIR/configs/chat.env" "$container/app/.env" 2>/dev/null || true
                fi
            fi
            
            if [[ "$action" == "build" || "$action" == "all" ]]; then
                echo "Compilation du chat server..."
                incus exec "$container" -- bash -c "source /root/.cargo/env && cd /app && cargo build --release"
            fi
            
            if [[ "$action" == "restart" || "$action" == "all" ]]; then
                echo "Redémarrage du chat server..."
                incus exec "$container" -- systemctl restart veza-chat 2>/dev/null || echo "  ⚠️  Service veza-chat non configuré"
            fi
            ;;
            
        stream)
            local container="veza-stream"
            check_container "$container" || return 1
            
            if [[ "$action" == "sync" || "$action" == "all" ]]; then
                echo "Synchronisation du stream server..."
                
                # Créer le répertoire app s'il n'existe pas
                incus exec "$container" -- mkdir -p /app
                
                # Copier les fichiers Rust
                echo "  → Copie des fichiers source..."
                incus file push -r "$WORKSPACE_DIR/veza-stream-server/src" "$container/app/" 2>/dev/null || true
                incus file push "$WORKSPACE_DIR/veza-stream-server/Cargo.toml" "$container/app/" 2>/dev/null || true
                incus file push "$WORKSPACE_DIR/veza-stream-server/Cargo.lock" "$container/app/" 2>/dev/null || true
                
                # Copier la configuration
                if [[ -f "$WORKSPACE_DIR/configs/stream.env" ]]; then
                    echo "  → Copie de la configuration..."
                    incus file push "$WORKSPACE_DIR/configs/stream.env" "$container/app/.env" 2>/dev/null || true
                fi
            fi
            
            if [[ "$action" == "build" || "$action" == "all" ]]; then
                echo "Compilation du stream server..."
                incus exec "$container" -- bash -c "source /root/.cargo/env && cd /app && cargo build --release"
            fi
            
            if [[ "$action" == "restart" || "$action" == "all" ]]; then
                echo "Redémarrage du stream server..."
                incus exec "$container" -- systemctl restart veza-stream 2>/dev/null || echo "  ⚠️  Service veza-stream non configuré"
            fi
            ;;
            
        frontend)
            local container="veza-frontend"
            check_container "$container" || return 1
            
            if [[ "$action" == "sync" || "$action" == "all" ]]; then
                echo "Synchronisation du frontend..."
                
                # Créer le répertoire app s'il n'existe pas
                incus exec "$container" -- mkdir -p /app
                
                # Copier les fichiers React
                echo "  → Copie des fichiers source..."
                incus file push -r "$WORKSPACE_DIR/veza-frontend/src" "$container/app/" 2>/dev/null || true
                incus file push -r "$WORKSPACE_DIR/veza-frontend/public" "$container/app/" 2>/dev/null || true
                incus file push "$WORKSPACE_DIR/veza-frontend/package.json" "$container/app/" 2>/dev/null || true
                incus file push "$WORKSPACE_DIR/veza-frontend/package-lock.json" "$container/app/" 2>/dev/null || true
                incus file push "$WORKSPACE_DIR/veza-frontend/vite.config.ts" "$container/app/" 2>/dev/null || true
                incus file push "$WORKSPACE_DIR/veza-frontend/tsconfig.json" "$container/app/" 2>/dev/null || true
                incus file push "$WORKSPACE_DIR/veza-frontend/tailwind.config.js" "$container/app/" 2>/dev/null || true
                incus file push "$WORKSPACE_DIR/veza-frontend/index.html" "$container/app/" 2>/dev/null || true
                
                # Copier la configuration
                if [[ -f "$WORKSPACE_DIR/configs/frontend.env" ]]; then
                    echo "  → Copie de la configuration..."
                    incus file push "$WORKSPACE_DIR/configs/frontend.env" "$container/app/.env" 2>/dev/null || true
                fi
            fi
            
            if [[ "$action" == "build" || "$action" == "all" ]]; then
                echo "Installation des dépendances et build du frontend..."
                incus exec "$container" -- bash -c "cd /app && npm install"
                incus exec "$container" -- bash -c "cd /app && npm run build"
            fi
            
            if [[ "$action" == "restart" || "$action" == "all" ]]; then
                echo "Redémarrage du frontend..."
                incus exec "$container" -- systemctl restart veza-frontend 2>/dev/null || echo "  ⚠️  Service veza-frontend non configuré"
            fi
            ;;
            
        all)
            # Traiter tous les services
            for s in backend chat stream frontend; do
                sync_and_build "$s" "$action"
                echo ""
            done
            ;;
            
        *)
            echo -e "${YELLOW}Service inconnu: $service${NC}"
            echo "Services disponibles: backend, chat, stream, frontend, all"
            exit 1
            ;;
    esac
    
    echo -e "${GREEN}[$service] Terminé!${NC}"
}

# Vérifier la santé d'un service
check_health() {
    local service=$1
    local ip
    
    case "$service" in
        backend)
            ip=$(get_container_ip veza-backend)
            if curl -s "http://$ip:8080/health" &>/dev/null; then
                echo -e "${GREEN}✅ Backend API OK${NC}"
            else
                echo -e "${YELLOW}⚠️  Backend API KO${NC}"
            fi
            ;;
        chat)
            ip=$(get_container_ip veza-chat)
            if curl -s "http://$ip:3001/health" &>/dev/null; then
                echo -e "${GREEN}✅ Chat Server OK${NC}"
            else
                echo -e "${YELLOW}⚠️  Chat Server KO${NC}"
            fi
            ;;
        stream)
            ip=$(get_container_ip veza-stream)
            if curl -s "http://$ip:3002/health" &>/dev/null; then
                echo -e "${GREEN}✅ Stream Server OK${NC}"
            else
                echo -e "${YELLOW}⚠️  Stream Server KO${NC}"
            fi
            ;;
        frontend)
            ip=$(get_container_ip veza-frontend)
            if curl -s "http://$ip:3000" &>/dev/null; then
                echo -e "${GREEN}✅ Frontend OK${NC}"
            else
                echo -e "${YELLOW}⚠️  Frontend KO${NC}"
            fi
            ;;
    esac
}

# Menu d'aide
show_help() {
    echo "Usage: $0 <service> [action]"
    echo ""
    echo "Services:"
    echo "  backend   - API Go"
    echo "  chat      - Chat Server Rust"
    echo "  stream    - Stream Server Rust"  
    echo "  frontend  - Frontend React"
    echo "  all       - Tous les services"
    echo ""
    echo "Actions:"
    echo "  all       - Sync + Build + Restart (défaut)"
    echo "  sync      - Synchroniser uniquement"
    echo "  build     - Compiler uniquement"
    echo "  restart   - Redémarrer uniquement"
    echo ""
    echo "Exemples:"
    echo "  $0 backend          # Sync, build et restart backend"
    echo "  $0 frontend sync    # Sync frontend uniquement"
    echo "  $0 all build        # Build tous les services"
}

# Point d'entrée principal
main() {
    if [[ $# -eq 0 ]]; then
        show_help
        exit 1
    fi
    
    local service=$1
    local action=${2:-all}
    
    if [[ "$service" == "help" || "$service" == "-h" || "$service" == "--help" ]]; then
        show_help
        exit 0
    fi
    
    sync_and_build "$service" "$action"
}

main "$@" 