#!/bin/bash

# Script de consultation des logs des containers Incus Veza

# Couleurs pour les messages
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

CONTAINER_NAME=$1
FOLLOW=${2:-false}

if [ -z "$CONTAINER_NAME" ]; then
    echo -e "${PURPLE}"
    echo "╭──────────────────────────────────────────╮"
    echo "│         📋 Veza - Logs Containers       │"
    echo "╰──────────────────────────────────────────╯"
    echo -e "${NC}"
    
    echo -e "${BLUE}Usage : $0 <container> [follow]${NC}"
    echo ""
    echo -e "${CYAN}Containers disponibles :${NC}"
    echo -e "  • ${YELLOW}veza-postgres${NC}  - Base de données PostgreSQL"
    echo -e "  • ${YELLOW}veza-redis${NC}     - Cache Redis"
    echo -e "  • ${YELLOW}veza-storage${NC}   - Système de fichiers NFS"
    echo -e "  • ${YELLOW}veza-backend${NC}   - API Backend Go"
    echo -e "  • ${YELLOW}veza-chat${NC}      - Serveur Chat Rust"
    echo -e "  • ${YELLOW}veza-stream${NC}    - Serveur Stream Rust"
    echo -e "  • ${YELLOW}veza-frontend${NC}  - Interface React"
    echo -e "  • ${YELLOW}veza-haproxy${NC}   - Load Balancer HAProxy"
    echo ""
    echo -e "${BLUE}Exemples :${NC}"
    echo -e "  $0 veza-backend          # Voir les logs du backend"
    echo -e "  $0 veza-chat follow      # Suivre les logs du chat en temps réel"
    echo -e "  $0 all                   # Voir les logs de tous les containers"
    exit 1
fi

# Fonction pour afficher les logs d'un container
show_logs() {
    local container=$1
    local follow_mode=$2
    
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}📋 Logs de $container${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    
    if ! incus info $container &>/dev/null; then
        echo -e "${RED}❌ Container $container non trouvé${NC}"
        return 1
    fi
    
    if [ "$follow_mode" = "follow" ]; then
        echo -e "${YELLOW}💡 Suivi en temps réel - Ctrl+C pour arrêter${NC}"
        echo ""
        
        case $container in
            "veza-postgres")
                incus exec $container -- tail -f /var/log/postgresql/postgresql-14-main.log
                ;;
            "veza-redis")
                incus exec $container -- tail -f /var/log/redis/redis-server.log
                ;;
            "veza-haproxy")
                incus exec $container -- tail -f /var/log/haproxy.log
                ;;
            *)
                # Pour les applications, logs système
                incus exec $container -- journalctl -f --no-pager
                ;;
        esac
    else
        echo -e "${YELLOW}💡 Derniers logs (100 lignes)${NC}"
        echo ""
        
        case $container in
            "veza-postgres")
                incus exec $container -- tail -n 100 /var/log/postgresql/postgresql-14-main.log 2>/dev/null || \
                incus exec $container -- journalctl -u postgresql -n 100 --no-pager
                ;;
            "veza-redis")
                incus exec $container -- tail -n 100 /var/log/redis/redis-server.log 2>/dev/null || \
                incus exec $container -- journalctl -u redis-server -n 100 --no-pager
                ;;
            "veza-storage")
                incus exec $container -- journalctl -u nfs-kernel-server -n 100 --no-pager
                ;;
            "veza-haproxy")
                incus exec $container -- tail -n 100 /var/log/haproxy.log 2>/dev/null || \
                incus exec $container -- journalctl -u haproxy -n 100 --no-pager
                ;;
            "veza-backend"|"veza-chat"|"veza-stream"|"veza-frontend")
                # Logs des applications personnalisées
                incus exec $container -- find /app/logs -name "*.log" -exec tail -n 50 {} + 2>/dev/null || \
                incus exec $container -- journalctl -n 100 --no-pager
                ;;
            *)
                incus exec $container -- journalctl -n 100 --no-pager
                ;;
        esac
    fi
}

# Afficher les logs de tous les containers
if [ "$CONTAINER_NAME" = "all" ]; then
    containers=("veza-postgres" "veza-redis" "veza-storage" "veza-backend" "veza-chat" "veza-stream" "veza-frontend" "veza-haproxy")
    
    echo -e "${PURPLE}"
    echo "╭──────────────────────────────────────────╮"
    echo "│      📋 Veza - Logs Tous Containers     │"
    echo "╰──────────────────────────────────────────╯"
    echo -e "${NC}"
    
    for container in "${containers[@]}"; do
        if incus info $container &>/dev/null; then
            show_logs $container false
            echo ""
        else
            echo -e "${YELLOW}⚠️ Container $container non trouvé, ignoré${NC}"
        fi
    done
else
    echo -e "${PURPLE}"
    echo "╭──────────────────────────────────────────╮"
    echo "│         📋 Veza - Logs Container        │"
    echo "╰──────────────────────────────────────────╯"
    echo -e "${NC}"
    
    show_logs $CONTAINER_NAME $FOLLOW
fi

echo ""
echo -e "${CYAN}💡 Commandes utiles :${NC}"
echo -e "  • ${YELLOW}incus exec $CONTAINER_NAME -- bash${NC} - Se connecter au container"
echo -e "  • ${YELLOW}./scripts/incus-status.sh${NC} - Vérifier le statut général"
echo -e "  • ${YELLOW}./scripts/incus-restart.sh $CONTAINER_NAME${NC} - Redémarrer le container" 