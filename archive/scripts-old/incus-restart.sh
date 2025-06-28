#!/bin/bash

# Script de redémarrage des containers Incus Veza

# Couleurs pour les messages
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

CONTAINER_NAME=$1
ACTION=${2:-restart}

if [ -z "$CONTAINER_NAME" ]; then
    echo -e "${PURPLE}"
    echo "╭──────────────────────────────────────────╮"
    echo "│       🔄 Veza - Gestion Containers      │"
    echo "╰──────────────────────────────────────────╯"
    echo -e "${NC}"
    
    echo -e "${BLUE}Usage : $0 <container|all> [start|stop|restart]${NC}"
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
    echo -e "  • ${YELLOW}all${NC}            - Tous les containers"
    echo ""
    echo -e "${BLUE}Actions disponibles :${NC}"
    echo -e "  • ${GREEN}start${NC}   - Démarrer le(s) container(s)"
    echo -e "  • ${RED}stop${NC}    - Arrêter le(s) container(s)"
    echo -e "  • ${YELLOW}restart${NC} - Redémarrer le(s) container(s) (défaut)"
    echo ""
    echo -e "${BLUE}Exemples :${NC}"
    echo -e "  $0 veza-backend restart  # Redémarrer le backend"
    echo -e "  $0 all stop             # Arrêter tous les containers"
    echo -e "  $0 veza-postgres start  # Démarrer PostgreSQL"
    exit 1
fi

# Configuration des containers dans l'ordre de démarrage
CONTAINERS_ORDER=(
    "veza-postgres"
    "veza-redis"
    "veza-storage"
    "veza-backend"
    "veza-chat"
    "veza-stream"
    "veza-frontend"
    "veza-haproxy"
)

# Ordre inverse pour l'arrêt
CONTAINERS_REVERSE_ORDER=(
    "veza-haproxy"
    "veza-frontend"
    "veza-stream"
    "veza-chat"
    "veza-backend"
    "veza-storage"
    "veza-redis"
    "veza-postgres"
)

# Fonction pour démarrer un container
start_container() {
    local container=$1
    
    echo -e "${BLUE}🚀 Démarrage de $container...${NC}"
    
    if ! incus info $container &>/dev/null; then
        echo -e "${RED}❌ Container $container non trouvé${NC}"
        return 1
    fi
    
    local status=$(incus list $container -c s --format csv 2>/dev/null)
    
    if [ "$status" = "RUNNING" ]; then
        echo -e "${YELLOW}⚠️ Container $container déjà en cours d'exécution${NC}"
        return 0
    fi
    
    if incus start $container; then
        echo -e "${GREEN}✅ Container $container démarré${NC}"
        
        # Attendre que le container soit complètement prêt
        echo -e "${YELLOW}⏳ Attente de l'initialisation...${NC}"
        sleep 3
        
        # Vérifications spécifiques selon le container
        case $container in
            "veza-postgres")
                for i in {1..30}; do
                    if incus exec $container -- pg_isready -U veza_user -d veza_db &>/dev/null; then
                        echo -e "${GREEN}✅ PostgreSQL prêt${NC}"
                        break
                    fi
                    sleep 2
                done
                ;;
            "veza-redis")
                for i in {1..15}; do
                    if incus exec $container -- redis-cli ping &>/dev/null | grep -q PONG; then
                        echo -e "${GREEN}✅ Redis prêt${NC}"
                        break
                    fi
                    sleep 1
                done
                ;;
            "veza-haproxy")
                for i in {1..20}; do
                    if curl -s http://10.100.0.16:8404/stats &>/dev/null; then
                        echo -e "${GREEN}✅ HAProxy prêt${NC}"
                        break
                    fi
                    sleep 1
                done
                ;;
        esac
        
    else
        echo -e "${RED}❌ Erreur lors du démarrage de $container${NC}"
        return 1
    fi
}

# Fonction pour arrêter un container
stop_container() {
    local container=$1
    
    echo -e "${BLUE}🛑 Arrêt de $container...${NC}"
    
    if ! incus info $container &>/dev/null; then
        echo -e "${RED}❌ Container $container non trouvé${NC}"
        return 1
    fi
    
    local status=$(incus list $container -c s --format csv 2>/dev/null)
    
    if [ "$status" = "STOPPED" ]; then
        echo -e "${YELLOW}⚠️ Container $container déjà arrêté${NC}"
        return 0
    fi
    
    if incus stop $container; then
        echo -e "${GREEN}✅ Container $container arrêté${NC}"
    else
        echo -e "${RED}❌ Erreur lors de l'arrêt de $container${NC}"
        return 1
    fi
}

# Fonction pour redémarrer un container
restart_container() {
    local container=$1
    
    echo -e "${BLUE}🔄 Redémarrage de $container...${NC}"
    
    stop_container $container
    sleep 2
    start_container $container
}

# Fonction pour gérer tous les containers
manage_all_containers() {
    local action=$1
    
    case $action in
        "start")
            echo -e "${GREEN}🚀 Démarrage de tous les containers...${NC}"
            for container in "${CONTAINERS_ORDER[@]}"; do
                start_container $container
                echo ""
            done
            ;;
        "stop")
            echo -e "${RED}🛑 Arrêt de tous les containers...${NC}"
            for container in "${CONTAINERS_REVERSE_ORDER[@]}"; do
                stop_container $container
                echo ""
            done
            ;;
        "restart")
            echo -e "${YELLOW}🔄 Redémarrage de tous les containers...${NC}"
            # Arrêter tous les containers dans l'ordre inverse
            for container in "${CONTAINERS_REVERSE_ORDER[@]}"; do
                stop_container $container
            done
            echo ""
            echo -e "${BLUE}⏳ Pause avant redémarrage...${NC}"
            sleep 5
            echo ""
            # Démarrer tous les containers dans l'ordre normal
            for container in "${CONTAINERS_ORDER[@]}"; do
                start_container $container
                echo ""
            done
            ;;
    esac
}

# Main
echo -e "${PURPLE}"
echo "╭──────────────────────────────────────────╮"
echo "│       🔄 Veza - Gestion Containers      │"
echo "╰──────────────────────────────────────────╯"
echo -e "${NC}"

if [ "$CONTAINER_NAME" = "all" ]; then
    manage_all_containers $ACTION
else
    case $ACTION in
        "start")
            start_container $CONTAINER_NAME
            ;;
        "stop")
            stop_container $CONTAINER_NAME
            ;;
        "restart")
            restart_container $CONTAINER_NAME
            ;;
        *)
            echo -e "${RED}❌ Action inconnue : $ACTION${NC}"
            echo -e "${BLUE}Actions disponibles : start, stop, restart${NC}"
            exit 1
            ;;
    esac
fi

echo ""
echo -e "${CYAN}💡 Commandes utiles :${NC}"
echo -e "  • ${YELLOW}./scripts/incus-status.sh${NC} - Vérifier le statut"
echo -e "  • ${YELLOW}./scripts/incus-logs.sh $CONTAINER_NAME${NC} - Voir les logs"
echo -e "  • ${YELLOW}incus exec $CONTAINER_NAME -- bash${NC} - Se connecter au container" 