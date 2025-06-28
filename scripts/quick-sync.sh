#!/bin/bash

# Script de synchronisation rapide
# Usage: ./scripts/quick-sync.sh [composant] [--build] [--restart]

set -e

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# Paramètres
COMPONENT="$1"
BUILD_FLAG="$2"
RESTART_FLAG="$3"

# Configuration SSH
SSH_KEY="$HOME/.ssh/veza_rsa"
SSH_OPTS="-i $SSH_KEY -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o LogLevel=ERROR"

# Fonction de synchronisation pour un composant
sync_component() {
    local component=$1
    local container=$2
    local local_path=$3
    local remote_path=$4
    local service=$5
    local build_needed=$6
    
    echo -e "${CYAN}🔄 Synchronisation $component...${NC}"
    
    # Obtenir l'IP du container
    local container_ip=$(incus ls "$container" -c 4 --format csv | cut -d' ' -f1)
    if [ -z "$container_ip" ]; then
        echo -e "${RED}❌ Impossible d'obtenir l'IP de $container${NC}"
        return 1
    fi
    
    # Vérifier que le répertoire local existe
    if [ ! -d "$local_path" ]; then
        echo -e "${RED}❌ Répertoire local $local_path non trouvé${NC}"
        return 1
    fi
    
    # Synchroniser avec rsync
    echo -e "${BLUE}📋 rsync: $local_path -> $container:$remote_path${NC}"
    rsync -avz --delete \
        --exclude='.git' \
        --exclude='node_modules' \
        --exclude='target' \
        --exclude='dist' \
        --exclude='build' \
        --exclude='.next' \
        --exclude='*.log' \
        -e "ssh $SSH_OPTS" \
        "$local_path" "root@$container_ip:$remote_path"
    
    # Build si nécessaire
    if [ "$build_needed" = "true" ] || [ "$BUILD_FLAG" = "--build" ]; then
        echo -e "${BLUE}🔨 Build de $component...${NC}"
        
        case $component in
            "backend")
                ssh $SSH_OPTS "root@$container_ip" "cd $remote_path && ./build.sh"
                ;;
            "chat"|"stream")
                ssh $SSH_OPTS "root@$container_ip" "cd $remote_path && ./build.sh"
                ;;
            "frontend")
                ssh $SSH_OPTS "root@$container_ip" "cd $remote_path && npm install"
                ;;
        esac
    fi
    
    # Redémarrer le service si demandé
    if [ "$BUILD_FLAG" = "--restart" ] || [ "$RESTART_FLAG" = "--restart" ]; then
        echo -e "${BLUE}🔄 Redémarrage du service $service...${NC}"
        ssh $SSH_OPTS "root@$container_ip" "systemctl restart $service"
        
        # Vérifier le statut
        sleep 2
        status=$(ssh $SSH_OPTS "root@$container_ip" "systemctl is-active $service" 2>/dev/null || echo "failed")
        
        case $status in
            "active")
                echo -e "${GREEN}✅ Service $service redémarré avec succès${NC}"
                ;;
            *)
                echo -e "${RED}❌ Échec redémarrage du service $service${NC}"
                echo -e "${YELLOW}Vérifiez les logs: incus exec $container -- journalctl -u $service -n 20${NC}"
                ;;
        esac
    fi
    
    echo -e "${GREEN}✅ $component synchronisé${NC}"
}

# Fonction de synchronisation complète
sync_all() {
    echo -e "${CYAN}🚀 Synchronisation complète de tous les composants...${NC}"
    echo ""
    
    # Lire la configuration
    while IFS=':' read -r local_path container remote_path service; do
        # Ignorer les commentaires et lignes vides
        [[ "$local_path" =~ ^#.*$ ]] && continue
        [[ -z "$local_path" ]] && continue
        
        # Déterminer le composant
        component=$(basename "$local_path" | sed 's/\/$//')
        component=$(echo "$component" | sed 's/veza-//' | sed 's/-server//' | sed 's/-api//')
        
        sync_component "$component" "$container" "$local_path" "$remote_path" "$service" "false"
        echo ""
    done < scripts/rsync-config.conf
}

# Afficher l'aide
show_help() {
    echo -e "${BLUE}🔄 Script de synchronisation rapide Veza${NC}"
    echo ""
    echo -e "${CYAN}Usage:${NC}"
    echo "  $0                    # Synchroniser tous les composants"
    echo "  $0 backend            # Synchroniser seulement le backend"
    echo "  $0 chat               # Synchroniser seulement le chat"
    echo "  $0 stream             # Synchroniser seulement le stream"
    echo "  $0 frontend           # Synchroniser seulement le frontend"
    echo "  $0 backend --build    # Synchroniser et builder"
    echo "  $0 backend --restart  # Synchroniser et redémarrer le service"
    echo ""
    echo -e "${CYAN}Exemples:${NC}"
    echo "  $0 backend --build --restart    # Sync + build + restart"
    echo "  $0 frontend --restart           # Sync frontend + restart"
}

# Fonction principale
main() {
    case "$COMPONENT" in
        "backend")
            sync_component "backend" "veza-backend" "veza-backend-api/" "/opt/veza/backend/" "veza-backend" "false"
            ;;
        "chat")
            sync_component "chat" "veza-chat" "veza-chat-server/" "/opt/veza/chat/" "veza-chat" "false"
            ;;
        "stream")
            sync_component "stream" "veza-stream" "veza-stream-server/" "/opt/veza/stream/" "veza-stream" "false"
            ;;
        "frontend")
            sync_component "frontend" "veza-frontend" "veza-frontend/" "/opt/veza/frontend/" "veza-frontend" "false"
            ;;
        "help"|"-h"|"--help")
            show_help
            ;;
        "")
            sync_all
            ;;
        *)
            echo -e "${RED}❌ Composant inconnu: $COMPONENT${NC}"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

main "$@"
