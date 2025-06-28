#!/bin/bash

# Script de vérification de l'état de l'infrastructure Veza
set -euo pipefail

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

echo -e "${BLUE}📊 État de l'infrastructure Veza...${NC}"
echo ""

# Fonction pour obtenir l'IP d'un container
get_container_ip() {
    local container_name=$1
    incus exec "$container_name" -- ip -4 addr show eth0 | grep -oP '(?<=inet\s)\d+(\.\d+){3}' | head -1 2>/dev/null || echo "N/A"
}

# Fonction pour vérifier l'état d'un service
check_service() {
    local container_name=$1
    local service_name=$2
    
    if incus list "$container_name" --format csv | grep -q "RUNNING"; then
        if incus exec "$container_name" -- systemctl is-active "$service_name" >/dev/null 2>&1; then
            echo -e "${GREEN}✅ ACTIF${NC}"
        elif incus exec "$container_name" -- systemctl is-enabled "$service_name" >/dev/null 2>&1; then
            echo -e "${YELLOW}⚠️ INSTALLÉ (inactif)${NC}"
        else
            echo -e "${RED}❌ NON INSTALLÉ${NC}"
        fi
    else
        echo -e "${RED}❌ CONTAINER ARRÊTÉ${NC}"
    fi
}

# Fonction pour tester la connectivité
test_connectivity() {
    local from_container=$1
    local to_ip=$2
    local to_port=$3
    
    if incus list "$from_container" --format csv | grep -q "RUNNING"; then
        if incus exec "$from_container" -- timeout 3 nc -z "$to_ip" "$to_port" 2>/dev/null; then
            echo -e "${GREEN}✅ OK${NC}"
        else
            echo -e "${RED}❌ ÉCHEC${NC}"
        fi
    else
        echo -e "${YELLOW}⚠️ CONTAINER ARRÊTÉ${NC}"
    fi
}

# 1. État des containers
echo -e "${BLUE}🏠 Containers Incus :${NC}"
incus ls | head -1
incus ls | grep -E "(veza-|NAME)" || echo "Aucun container Veza trouvé"
echo ""

# 2. État des services par container
echo -e "${BLUE}⚙️ État des services :${NC}"
printf "%-15s %-15s %-15s %-15s\n" "Container" "IP" "Service" "État"
printf "%-15s %-15s %-15s %-15s\n" "----------" "---------------" "----------" "----------"

containers=(
    "veza-postgres:postgresql"
    "veza-redis:redis-server"
    "veza-storage:nfs-kernel-server"
    "veza-backend:veza-backend"
    "veza-chat:veza-chat"
    "veza-stream:veza-stream"
    "veza-frontend:veza-frontend"
    "veza-haproxy:haproxy"
)

for item in "${containers[@]}"; do
    container_name="${item%:*}"
    service_name="${item#*:}"
    
    if incus list "$container_name" --format csv >/dev/null 2>&1; then
        ip=$(get_container_ip "$container_name")
        status=$(check_service "$container_name" "$service_name")
        printf "%-15s %-15s %-15s %s\n" "$container_name" "$ip" "$service_name" "$status"
    else
        printf "%-15s %-15s %-15s %s\n" "$container_name" "N/A" "$service_name" "${RED}❌ INEXISTANT${NC}"
    fi
done
echo ""

# 3. Tests de connectivité inter-services
echo -e "${BLUE}🌐 Tests de connectivité :${NC}"
printf "%-25s %-15s %-15s\n" "Test" "Port" "État"
printf "%-25s %-15s %-15s\n" "------------------------" "----------" "----------"

# Obtenir les IPs
postgres_ip=$(get_container_ip "veza-postgres")
redis_ip=$(get_container_ip "veza-redis")
backend_ip=$(get_container_ip "veza-backend")
chat_ip=$(get_container_ip "veza-chat")
stream_ip=$(get_container_ip "veza-stream")
frontend_ip=$(get_container_ip "veza-frontend")
haproxy_ip=$(get_container_ip "veza-haproxy")

# Tests de connectivité
if [ "$postgres_ip" != "N/A" ] && [ "$backend_ip" != "N/A" ]; then
    status=$(test_connectivity "veza-backend" "$postgres_ip" "5432")
    printf "%-25s %-15s %s\n" "Backend → PostgreSQL" "5432" "$status"
fi

if [ "$redis_ip" != "N/A" ] && [ "$backend_ip" != "N/A" ]; then
    status=$(test_connectivity "veza-backend" "$redis_ip" "6379")
    printf "%-25s %-15s %s\n" "Backend → Redis" "6379" "$status"
fi

if [ "$backend_ip" != "N/A" ]; then
    status=$(test_connectivity "veza-backend" "$backend_ip" "8080")
    printf "%-25s %-15s %s\n" "Backend API" "8080" "$status"
fi

if [ "$chat_ip" != "N/A" ]; then
    status=$(test_connectivity "veza-chat" "$chat_ip" "8081")
    printf "%-25s %-15s %s\n" "Chat WebSocket" "8081" "$status"
fi

if [ "$stream_ip" != "N/A" ]; then
    status=$(test_connectivity "veza-stream" "$stream_ip" "8082")
    printf "%-25s %-15s %s\n" "Stream Server" "8082" "$status"
fi

if [ "$frontend_ip" != "N/A" ]; then
    status=$(test_connectivity "veza-frontend" "$frontend_ip" "5173")
    printf "%-25s %-15s %s\n" "Frontend React" "5173" "$status"
fi

if [ "$haproxy_ip" != "N/A" ]; then
    status=$(test_connectivity "veza-haproxy" "$haproxy_ip" "80")
    printf "%-25s %-15s %s\n" "HAProxy HTTP" "80" "$status"
    
    status=$(test_connectivity "veza-haproxy" "$haproxy_ip" "8404")
    printf "%-25s %-15s %s\n" "HAProxy Stats" "8404" "$status"
fi

echo ""

# 4. Points d'accès
echo -e "${BLUE}🌐 Points d'accès :${NC}"
if [ "$haproxy_ip" != "N/A" ]; then
    echo -e "  • Application : ${YELLOW}http://$haproxy_ip${NC}"
    echo -e "  • HAProxy Stats : ${YELLOW}http://$haproxy_ip:8404/stats${NC}"
fi
if [ "$frontend_ip" != "N/A" ]; then
    echo -e "  • Frontend Dev : ${YELLOW}http://$frontend_ip:5173${NC}"
fi
if [ "$backend_ip" != "N/A" ]; then
    echo -e "  • Backend API : ${YELLOW}http://$backend_ip:8080${NC}"
fi
if [ "$chat_ip" != "N/A" ]; then
    echo -e "  • Chat WebSocket : ${YELLOW}ws://$chat_ip:8081/ws${NC}"
fi
if [ "$stream_ip" != "N/A" ]; then
    echo -e "  • Stream Server : ${YELLOW}http://$stream_ip:8082${NC}"
fi

echo ""

# 5. Commandes utiles
echo -e "${CYAN}💡 Commandes utiles :${NC}"
echo -e "  • Logs service : ${YELLOW}incus exec <container> -- journalctl -u <service> -f${NC}"
echo -e "  • Shell container : ${YELLOW}incus exec <container> -- bash${NC}"
echo -e "  • Redémarrer service : ${YELLOW}incus exec <container> -- systemctl restart <service>${NC}"
echo -e "  • État détaillé : ${YELLOW}incus exec <container> -- systemctl status <service>${NC}"
echo -e "  • Redémarrer container : ${YELLOW}incus restart <container>${NC}" 