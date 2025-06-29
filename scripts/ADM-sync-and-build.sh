#!/bin/bash
# Script de synchronisation et build rapide pour développement

set -euo pipefail

# Configuration
WORKSPACE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# Couleurs
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Fonction pour obtenir l'IP d'un container
get_container_ip() {
    incus list "$1" --format csv | cut -d, -f3 | grep -E '^10\.' | head -1
}

# Fonction de sync et build pour un service
sync_and_build() {
    local service=$1
    local action=${2:-all} # sync, build, restart, ou all
    
    echo -e "${BLUE}[$service] Traitement...${NC}"
    
    case "$service" in
        backend)
            local ip=$(get_container_ip veza-backend)
            
            if [[ "$action" == "sync" || "$action" == "all" ]]; then
                echo "Synchronisation..."
                rsync -avz --delete \
                    --exclude '.git' \
                    --exclude '*.log' \
                    --exclude '.env.local' \
                    "$WORKSPACE_DIR/veza-backend-api/" \
                    "root@$ip:/app/"
            fi
            
            if [[ "$action" == "build" || "$action" == "all" ]]; then
                echo "Compilation..."
                ssh "root@$ip" "cd /app && PATH=\$PATH:/usr/local/go/bin go build -o veza-backend cmd/server/main.go"
            fi
            
            if [[ "$action" == "restart" || "$action" == "all" ]]; then
                echo "Redémarrage..."
                ssh "root@$ip" "systemctl restart veza-backend"
            fi
            ;;
            
        chat)
            local ip=$(get_container_ip veza-chat)
            
            if [[ "$action" == "sync" || "$action" == "all" ]]; then
                echo "Synchronisation..."
                rsync -avz --delete \
                    --exclude '.git' \
                    --exclude 'target' \
                    --exclude '*.log' \
                    "$WORKSPACE_DIR/veza-chat-server/" \
                    "root@$ip:/app/"
            fi
            
            if [[ "$action" == "build" || "$action" == "all" ]]; then
                echo "Compilation..."
                ssh "root@$ip" "source /root/.cargo/env && cd /app && cargo build --release"
            fi
            
            if [[ "$action" == "restart" || "$action" == "all" ]]; then
                echo "Redémarrage..."
                ssh "root@$ip" "systemctl restart veza-chat"
            fi
            ;;
            
        stream)
            local ip=$(get_container_ip veza-stream)
            
            if [[ "$action" == "sync" || "$action" == "all" ]]; then
                echo "Synchronisation..."
                rsync -avz --delete \
                    --exclude '.git' \
                    --exclude 'target' \
                    --exclude '*.log' \
                    "$WORKSPACE_DIR/veza-stream-server/" \
                    "root@$ip:/app/"
            fi
            
            if [[ "$action" == "build" || "$action" == "all" ]]; then
                echo "Compilation..."
                ssh "root@$ip" "source /root/.cargo/env && cd /app && cargo build --release"
            fi
            
            if [[ "$action" == "restart" || "$action" == "all" ]]; then
                echo "Redémarrage..."
                ssh "root@$ip" "systemctl restart veza-stream"
            fi
            ;;
            
        frontend)
            local ip=$(get_container_ip veza-frontend)
            
            if [[ "$action" == "sync" || "$action" == "all" ]]; then
                echo "Synchronisation..."
                rsync -avz --delete \
                    --exclude '.git' \
                    --exclude 'node_modules' \
                    --exclude 'dist' \
                    --exclude '.env.local' \
                    "$WORKSPACE_DIR/veza-frontend/" \
                    "root@$ip:/app/"
            fi
            
            if [[ "$action" == "build" || "$action" == "all" ]]; then
                echo "Installation dépendances et build..."
                ssh "root@$ip" "cd /app && npm install && npm run build"
            fi
            
            if [[ "$action" == "restart" || "$action" == "all" ]]; then
                echo "Redémarrage..."
                ssh "root@$ip" "systemctl restart veza-frontend"
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
    local port
    
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

# Main
if [ $# -eq 0 ]; then
    show_help
    exit 1
fi

service=$1
action=${2:-all}

# Exécuter l'action
sync_and_build "$service" "$action"

# Vérifier la santé après redémarrage
if [[ "$action" == "restart" || "$action" == "all" ]]; then
    echo ""
    echo "Vérification santé..."
    sleep 3
    
    if [ "$service" == "all" ]; then
        for s in backend chat stream frontend; do
            check_health "$s"
        done
    else
        check_health "$service"
    fi
fi
