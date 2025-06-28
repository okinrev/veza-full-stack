#!/bin/bash

# Script de surveillance automatique avec inotify
# Synchronise automatiquement quand des fichiers changent

set -e

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# Vérifier inotify-tools
if ! command -v inotifywait &> /dev/null; then
    echo -e "${YELLOW}📦 Installation d'inotify-tools...${NC}"
    sudo dnf install -y inotify-tools || sudo apt-get install -y inotify-tools
fi

echo -e "${BLUE}"
echo "╭──────────────────────────────────────────╮"
echo "│     👁️ Surveillance Automatique des      │"
echo "│           Changements de Code             │"
echo "╰──────────────────────────────────────────╯"
echo -e "${NC}"

# Configuration
DEBOUNCE_TIME=2  # Attendre 2 secondes avant de synchroniser

# Fonction de surveillance pour un répertoire
watch_directory() {
    local dir=$1
    local component=$2
    
    echo -e "${CYAN}👁️ Surveillance de $dir ($component)...${NC}"
    
    inotifywait -m -r -e modify,create,delete,move \
        --exclude='(\.git|node_modules|target|dist|build|\.next|.*\.log)' \
        "$dir" | while read path action file; do
        
        echo -e "${YELLOW}📝 Changement détecté: $path$file ($action)${NC}"
        
        # Debounce - attendre un peu puis synchroniser
        sleep $DEBOUNCE_TIME
        
        echo -e "${BLUE}🔄 Synchronisation de $component...${NC}"
        ./scripts/quick-sync.sh "$component" --restart
        
        echo -e "${GREEN}✅ $component synchronisé et redémarré${NC}"
        echo "────────────────────────────────────────"
    done
}

# Surveillance en parallèle
main() {
    local component="$1"
    
    case "$component" in
        "backend")
            watch_directory "veza-backend-api" "backend"
            ;;
        "chat")
            watch_directory "veza-chat-server" "chat"
            ;;
        "stream")
            watch_directory "veza-stream-server" "stream"
            ;;
        "frontend")
            watch_directory "veza-frontend" "frontend"
            ;;
        "all"|"")
            echo -e "${CYAN}🚀 Surveillance de tous les composants...${NC}"
            echo -e "${YELLOW}Appuyez sur Ctrl+C pour arrêter${NC}"
            echo ""
            
            # Lancer en arrière-plan
            watch_directory "veza-backend-api" "backend" &
            watch_directory "veza-chat-server" "chat" &
            watch_directory "veza-stream-server" "stream" &
            watch_directory "veza-frontend" "frontend" &
            
            # Attendre
            wait
            ;;
        "help"|"-h"|"--help")
            echo -e "${BLUE}👁️ Script de surveillance automatique${NC}"
            echo ""
            echo -e "${CYAN}Usage:${NC}"
            echo "  $0              # Surveiller tous les composants"
            echo "  $0 backend      # Surveiller seulement le backend"
            echo "  $0 chat         # Surveiller seulement le chat"
            echo "  $0 stream       # Surveiller seulement le stream"
            echo "  $0 frontend     # Surveiller seulement le frontend"
            ;;
        *)
            echo -e "${RED}❌ Composant inconnu: $component${NC}"
            exit 1
            ;;
    esac
}

main "$@"
