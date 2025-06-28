#!/bin/bash

# Script de gestion des services Veza dans les containers Incus
# Permet de démarrer, arrêter, redémarrer et vérifier l'état des services

set -e

# Couleurs pour les messages
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

# Services et leurs containers
declare -A SERVICES=(
    ["backend"]="veza-backend:veza-backend"
    ["chat"]="veza-chat:veza-chat"
    ["stream"]="veza-stream:veza-stream"
    ["frontend"]="veza-frontend:veza-frontend"
    ["postgres"]="veza-postgres:postgresql"
    ["redis"]="veza-redis:redis-server"
    ["haproxy"]="veza-haproxy:haproxy"
)

# Fonction d'aide
show_help() {
    echo -e "${PURPLE}🔧 Gestionnaire des Services Veza${NC}"
    echo ""
    echo -e "${BLUE}Usage:${NC}"
    echo -e "  $0 <action> [service]"
    echo ""
    echo -e "${BLUE}Actions:${NC}"
    echo -e "  ${GREEN}status${NC}    - Afficher l'état de tous les services"
    echo -e "  ${GREEN}start${NC}     - Démarrer un service ou tous"
    echo -e "  ${GREEN}stop${NC}      - Arrêter un service ou tous"
    echo -e "  ${GREEN}restart${NC}   - Redémarrer un service ou tous"
    echo -e "  ${GREEN}logs${NC}      - Afficher les logs d'un service"
    echo -e "  ${GREEN}health${NC}    - Vérifier la santé des services"
    echo ""
    echo -e "${BLUE}Services disponibles:${NC}"
    for service in "${!SERVICES[@]}"; do
        echo -e "  • ${CYAN}$service${NC}"
    done
    echo ""
    echo -e "${BLUE}Exemples:${NC}"
    echo -e "  $0 status"
    echo -e "  $0 start backend"
    echo -e "  $0 restart chat"
    echo -e "  $0 logs frontend"
}

# Vérifier l'état d'un service
check_service_status() {
    local service_key=$1
    local container_service=(${SERVICES[$service_key]//:/ })
    local container=${container_service[0]}
    local service=${container_service[1]}
    
    if incus exec "$container" -- systemctl is-active "$service" >/dev/null 2>&1; then
        echo -e "${GREEN}✅ $service_key (actif)${NC}"
        return 0
    elif incus exec "$container" -- systemctl is-failed "$service" >/dev/null 2>&1; then
        echo -e "${RED}❌ $service_key (échec)${NC}"
        return 1
    else
        echo -e "${YELLOW}⏸️ $service_key (arrêté)${NC}"
        return 2
    fi
}

# Afficher l'état de tous les services
status_all() {
    echo -e "${BLUE}📊 État des services Veza:${NC}"
    echo ""
    
    for service_key in "${!SERVICES[@]}"; do
        check_service_status "$service_key"
    done
    
    echo ""
    echo -e "${BLUE}📊 État des containers:${NC}"
    incus ls --format=table --columns=n,s,4,6
}

# Démarrer un service
start_service() {
    local service_key=$1
    local container_service=(${SERVICES[$service_key]//:/ })
    local container=${container_service[0]}
    local service=${container_service[1]}
    
    echo -e "${CYAN}🔄 Démarrage de $service_key...${NC}"
    
    if incus exec "$container" -- systemctl start "$service"; then
        echo -e "${GREEN}✅ $service_key démarré${NC}"
    else
        echo -e "${RED}❌ Échec du démarrage de $service_key${NC}"
        return 1
    fi
}

# Arrêter un service
stop_service() {
    local service_key=$1
    local container_service=(${SERVICES[$service_key]//:/ })
    local container=${container_service[0]}
    local service=${container_service[1]}
    
    echo -e "${CYAN}⏹️ Arrêt de $service_key...${NC}"
    
    if incus exec "$container" -- systemctl stop "$service"; then
        echo -e "${GREEN}✅ $service_key arrêté${NC}"
    else
        echo -e "${RED}❌ Échec de l'arrêt de $service_key${NC}"
        return 1
    fi
}

# Redémarrer un service
restart_service() {
    local service_key=$1
    local container_service=(${SERVICES[$service_key]//:/ })
    local container=${container_service[0]}
    local service=${container_service[1]}
    
    echo -e "${CYAN}🔄 Redémarrage de $service_key...${NC}"
    
    if incus exec "$container" -- systemctl restart "$service"; then
        echo -e "${GREEN}✅ $service_key redémarré${NC}"
    else
        echo -e "${RED}❌ Échec du redémarrage de $service_key${NC}"
        return 1
    fi
}

# Afficher les logs d'un service
logs_service() {
    local service_key=$1
    local container_service=(${SERVICES[$service_key]//:/ })
    local container=${container_service[0]}
    local service=${container_service[1]}
    
    echo -e "${CYAN}📋 Logs de $service_key (Ctrl+C pour quitter):${NC}"
    incus exec "$container" -- journalctl -u "$service" -f --no-pager
}

# Vérifier la santé des services
health_check() {
    echo -e "${BLUE}🏥 Vérification de santé des services:${NC}"
    echo ""
    
    # Vérifier la connectivité réseau
    echo -e "${CYAN}🌐 Test de connectivité:${NC}"
    
    # Test PostgreSQL
    if incus exec veza-backend -- nc -z 10.100.0.15 5432; then
        echo -e "${GREEN}✅ PostgreSQL accessible${NC}"
    else
        echo -e "${RED}❌ PostgreSQL inaccessible${NC}"
    fi
    
    # Test Redis
    if incus exec veza-backend -- nc -z 10.100.0.17 6379; then
        echo -e "${GREEN}✅ Redis accessible${NC}"
    else
        echo -e "${RED}❌ Redis inaccessible${NC}"
    fi
    
    # Test Backend API
    if incus exec veza-frontend -- curl -s http://10.100.0.12:8080/health >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Backend API répond${NC}"
    else
        echo -e "${YELLOW}⚠️ Backend API ne répond pas${NC}"
    fi
    
    # Test Chat WebSocket
    if incus exec veza-frontend -- nc -z 10.100.0.13 8081; then
        echo -e "${GREEN}✅ Chat WebSocket accessible${NC}"
    else
        echo -e "${YELLOW}⚠️ Chat WebSocket inaccessible${NC}"
    fi
    
    # Test Stream Server
    if incus exec veza-frontend -- nc -z 10.100.0.14 8082; then
        echo -e "${GREEN}✅ Stream Server accessible${NC}"
    else
        echo -e "${YELLOW}⚠️ Stream Server inaccessible${NC}"
    fi
    
    echo ""
    echo -e "${CYAN}📊 Utilisation des ressources:${NC}"
    incus info --resources
}

# Fonction principale
main() {
    local action=$1
    local service=$2
    
    case $action in
        "status")
            status_all
            ;;
        "start")
            if [ -z "$service" ]; then
                for service_key in backend chat stream frontend; do
                    start_service "$service_key"
                done
            elif [ -n "${SERVICES[$service]}" ]; then
                start_service "$service"
            else
                echo -e "${RED}❌ Service '$service' inconnu${NC}"
                show_help
                exit 1
            fi
            ;;
        "stop")
            if [ -z "$service" ]; then
                for service_key in frontend stream chat backend; do
                    stop_service "$service_key"
                done
            elif [ -n "${SERVICES[$service]}" ]; then
                stop_service "$service"
            else
                echo -e "${RED}❌ Service '$service' inconnu${NC}"
                show_help
                exit 1
            fi
            ;;
        "restart")
            if [ -z "$service" ]; then
                for service_key in backend chat stream frontend; do
                    restart_service "$service_key"
                done
            elif [ -n "${SERVICES[$service]}" ]; then
                restart_service "$service"
            else
                echo -e "${RED}❌ Service '$service' inconnu${NC}"
                show_help
                exit 1
            fi
            ;;
        "logs")
            if [ -n "${SERVICES[$service]}" ]; then
                logs_service "$service"
            else
                echo -e "${RED}❌ Vous devez spécifier un service pour les logs${NC}"
                show_help
                exit 1
            fi
            ;;
        "health")
            health_check
            ;;
        "help"|"--help"|"-h"|"")
            show_help
            ;;
        *)
            echo -e "${RED}❌ Action '$action' inconnue${NC}"
            show_help
            exit 1
            ;;
    esac
}

# Vérifier qu'Incus est disponible
if ! command -v incus &> /dev/null; then
    echo -e "${RED}❌ Incus n'est pas installé ou accessible${NC}"
    exit 1
fi

# Exécuter la commande
main "$@" 