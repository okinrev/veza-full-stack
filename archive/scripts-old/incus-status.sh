#!/bin/bash

# Script de vérification du statut des containers Incus Veza

set -e

# Couleurs pour les messages
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

echo -e "${PURPLE}"
echo "╭──────────────────────────────────────────╮"
echo "│        🔍 Veza - Statut Containers      │"
echo "╰──────────────────────────────────────────╯"
echo -e "${NC}"

# Configuration des containers
declare -A CONTAINERS=(
    ["veza-postgres"]="10.100.0.15:5432"
    ["veza-redis"]="10.100.0.17:6379"
    ["veza-storage"]="10.100.0.18:2049"
    ["veza-backend"]="10.100.0.12:8080"
    ["veza-chat"]="10.100.0.13:8081"
    ["veza-stream"]="10.100.0.14:8082"
    ["veza-frontend"]="10.100.0.11:5173"
    ["veza-haproxy"]="10.100.0.16:80"
)

# Fonction pour vérifier la connectivité
check_connectivity() {
    local container=$1
    local ip_port=$2
    local ip=$(echo $ip_port | cut -d: -f1)
    local port=$(echo $ip_port | cut -d: -f2)
    
    if timeout 3 bash -c "</dev/tcp/$ip/$port" &>/dev/null; then
        echo -e "${GREEN}✅${NC}"
    else
        echo -e "${RED}❌${NC}"
    fi
}

# Fonction pour obtenir l'utilisation des ressources
get_resource_usage() {
    local container=$1
    
    # CPU et Mémoire
    local cpu_mem=$(incus info $container 2>/dev/null | grep -E "(CPU usage|Memory usage)" | awk '{print $3}' | paste -sd ";" -)
    if [ -n "$cpu_mem" ]; then
        echo "$cpu_mem"
    else
        echo "N/A;N/A"
    fi
}

# Fonction pour vérifier les services
check_service() {
    local container=$1
    local service=$2
    
    if incus exec $container -- systemctl is-active $service &>/dev/null; then
        echo -e "${GREEN}✅${NC}"
    else
        echo -e "${RED}❌${NC}"
    fi
}

echo -e "${BLUE}📊 État des containers Veza${NC}"
echo -e "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

printf "%-15s %-12s %-15s %-10s %-15s %-10s\n" "CONTAINER" "STATUT" "IP" "PORT" "RESSOURCES" "SERVICE"
echo -e "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

for container in "${!CONTAINERS[@]}"; do
    ip_port=${CONTAINERS[$container]}
    ip=$(echo $ip_port | cut -d: -f1)
    port=$(echo $ip_port | cut -d: -f2)
    
    # Statut du container
    if incus info $container &>/dev/null; then
        status=$(incus list $container -c s --format csv 2>/dev/null || echo "UNKNOWN")
        if [ "$status" = "RUNNING" ]; then
            status_icon="${GREEN}🟢 RUNNING${NC}"
        else
            status_icon="${RED}🔴 $status${NC}"
        fi
    else
        status_icon="${RED}🚫 NOT FOUND${NC}"
        status="NOT_FOUND"
    fi
    
    # Connectivité
    connectivity=$(check_connectivity $container $ip_port)
    
    # Ressources
    if [ "$status" = "RUNNING" ]; then
        resources=$(get_resource_usage $container)
        cpu=$(echo $resources | cut -d';' -f1)
        mem=$(echo $resources | cut -d';' -f2)
        resource_str="${cpu}/${mem}"
    else
        resource_str="N/A"
    fi
    
    # Service spécifique selon le container
    if [ "$status" = "RUNNING" ]; then
        case $container in
            "veza-postgres")
                service_status=$(check_service $container postgresql)
                ;;
            "veza-redis")
                service_status=$(check_service $container redis-server)
                ;;
            "veza-storage")
                service_status=$(check_service $container nfs-kernel-server)
                ;;
            "veza-haproxy")
                service_status=$(check_service $container haproxy)
                ;;
            *)
                service_status="${YELLOW}⚙️${NC}"
                ;;
        esac
    else
        service_status="${RED}❌${NC}"
    fi
    
    printf "%-25s %-20s %-15s %-10s %-15s %-10s\n" \
        "$container" \
        "$status_icon" \
        "$ip" \
        "$connectivity $port" \
        "$resource_str" \
        "$service_status"
done

echo ""

# Vérification du réseau
echo -e "${BLUE}🌐 État du réseau${NC}"
echo -e "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if incus network show veza-network &>/dev/null; then
    network_status="${GREEN}✅ veza-network actif${NC}"
    
    # Obtenir les informations du réseau
    network_info=$(incus network show veza-network 2>/dev/null)
    ipv4_addr=$(echo "$network_info" | grep "ipv4.address:" | awk '{print $2}')
    
    echo -e "Réseau : $network_status"
    echo -e "Plage IP : ${CYAN}$ipv4_addr${NC}"
else
    echo -e "${RED}❌ Réseau veza-network non trouvé${NC}"
fi

echo ""

# Health checks des services
echo -e "${BLUE}🏥 Health Checks${NC}"
echo -e "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# PostgreSQL Health Check
if incus exec veza-postgres -- pg_isready -U veza_user -d veza_db &>/dev/null; then
    echo -e "PostgreSQL : ${GREEN}✅ Connecté${NC}"
else
    echo -e "PostgreSQL : ${RED}❌ Déconnecté${NC}"
fi

# Redis Health Check
if incus exec veza-redis -- redis-cli ping &>/dev/null | grep -q PONG; then
    echo -e "Redis      : ${GREEN}✅ Connecté${NC}"
else
    echo -e "Redis      : ${RED}❌ Déconnecté${NC}"
fi

# Backend API Health Check
if curl -s http://10.100.0.12:8080/health &>/dev/null; then
    echo -e "Backend    : ${GREEN}✅ API disponible${NC}"
else
    echo -e "Backend    : ${RED}❌ API indisponible${NC}"
fi

# HAProxy Stats
if curl -s http://10.100.0.16:8404/stats &>/dev/null; then
    echo -e "HAProxy    : ${GREEN}✅ Stats disponibles${NC}"
else
    echo -e "HAProxy    : ${RED}❌ Stats indisponibles${NC}"
fi

echo ""

# Résumé final
echo -e "${BLUE}📋 Résumé${NC}"
echo -e "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

running_count=0
total_count=${#CONTAINERS[@]}

for container in "${!CONTAINERS[@]}"; do
    if incus list $container -c s --format csv 2>/dev/null | grep -q "RUNNING"; then
        ((running_count++))
    fi
done

if [ $running_count -eq $total_count ]; then
    echo -e "${GREEN}🎉 Tous les containers sont opérationnels ($running_count/$total_count)${NC}"
else
    echo -e "${YELLOW}⚠️ $running_count/$total_count containers en cours d'exécution${NC}"
fi

echo ""
echo -e "${CYAN}💡 Commandes utiles :${NC}"
echo -e "  • ${YELLOW}incus list${NC} - Liste des containers"
echo -e "  • ${YELLOW}./scripts/incus-logs.sh [container]${NC} - Voir les logs"
echo -e "  • ${YELLOW}incus exec [container] -- bash${NC} - Se connecter à un container"
echo -e "  • ${YELLOW}./scripts/incus-restart.sh [container]${NC} - Redémarrer un container" 