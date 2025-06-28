#!/bin/bash

# Veza Manager - Script principal unifié de gestion de l'infrastructure
# Toutes les fonctionnalités intégrées : setup, deploy, status, logs, services, etc.

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORKSPACE_DIR="$(dirname "$SCRIPT_DIR")"
VERSION="2.0.0"

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

# Logo et titre
show_header() {
    echo -e "${PURPLE}"
    echo "╭──────────────────────────────────────────╮"
    echo "│      🚀 VEZA MANAGER v${VERSION}           │"
    echo "│    Infrastructure Management Complète   │"
    echo "╰──────────────────────────────────────────╯"
    echo -e "${NC}"
}

# Configuration des containers et services
declare -A CONTAINERS=(
    ["postgres"]="veza-postgres"
    ["redis"]="veza-redis"
    ["storage"]="veza-storage"
    ["backend"]="veza-backend"
    ["chat"]="veza-chat"
    ["stream"]="veza-stream"
    ["frontend"]="veza-frontend"
    ["haproxy"]="veza-haproxy"
)

declare -A SERVICES=(
    ["postgres"]="postgresql"
    ["redis"]="redis-server"
    ["storage"]="nfs-kernel-server"
    ["backend"]="veza-backend"
    ["chat"]="veza-chat"
    ["stream"]="veza-stream"
    ["frontend"]="veza-frontend"
    ["haproxy"]="haproxy"
)

# Fonction d'aide complète
show_help() {
    show_header
    echo -e "${BLUE}COMMANDES PRINCIPALES:${NC}"
    echo -e "  ${GREEN}setup${NC}        - Configuration initiale complète"
    echo -e "  ${GREEN}deploy${NC}       - Déploiement complet de l'infrastructure"
    echo -e "  ${GREEN}status${NC}       - État complet de l'infrastructure"
    echo -e "  ${GREEN}health${NC}       - Vérification de santé complète"
    echo ""
    echo -e "${BLUE}GESTION DES SERVICES:${NC}"
    echo -e "  ${GREEN}start${NC}    [service]  - Démarrer tous les services ou un service spécifique"
    echo -e "  ${GREEN}stop${NC}     [service]  - Arrêter tous les services ou un service spécifique"  
    echo -e "  ${GREEN}restart${NC}  [service]  - Redémarrer tous les services ou un service spécifique"
    echo -e "  ${GREEN}logs${NC}     <service>  - Afficher les logs d'un service"
    echo ""
    echo -e "${BLUE}BUILD ET SYNCHRONISATION:${NC}"
    echo -e "  ${GREEN}build${NC}    [service]  - Compiler tous les projets ou un service spécifique"
    echo -e "  ${GREEN}sync${NC}        - Synchroniser le code source (rsync)"
    echo -e "  ${GREEN}watch${NC}       - Surveillance automatique et sync en temps réel"
    echo -e "  ${GREEN}build-start${NC} - Build complet + démarrage de tous les services"
    echo ""
    echo -e "${BLUE}CONTAINERS ET DÉPLOIEMENT:${NC}"
    echo -e "  ${GREEN}export${NC}       - Exporter les containers de base"
    echo -e "  ${GREEN}import${NC}       - Importer les containers de base"
    echo -e "  ${GREEN}clean${NC}        - Nettoyage complet"
    echo ""
    echo -e "${BLUE}RÉSEAU ET MAINTENANCE:${NC}"
    echo -e "  ${GREEN}network-fix${NC}  - Réparer les problèmes réseau"
    echo -e "  ${GREEN}update${NC}       - Mettre à jour le code source"
    echo -e "  ${GREEN}fix-deps${NC}     - Installer/réparer les dépendances"
    echo ""
    echo -e "${BLUE}BASE DE DONNÉES:${NC}"
    echo -e "  ${GREEN}init-db${NC}      - Initialiser la base de données PostgreSQL"
    echo ""
    echo -e "${BLUE}EXEMPLES:${NC}"
    echo -e "  $0 setup                    # Configuration complète"
    echo -e "  $0 deploy                   # Déploiement complet"
    echo -e "  $0 status                   # État global"
    echo -e "  $0 build-start              # Build complet + démarrage"
    echo -e "  $0 sync                     # Synchroniser le code"
    echo -e "  $0 watch                    # Surveillance auto + sync"
    echo -e "  $0 build backend            # Compiler uniquement le backend"
    echo -e "  $0 start backend            # Démarrer le backend"
    echo -e "  $0 logs chat -f            # Logs du chat en temps réel"
    echo -e "  $0 health                  # Vérification de santé"
    echo -e "  $0 fix-deps                # Réparer les dépendances"
    echo ""
    echo -e "${CYAN}Services disponibles: ${!CONTAINERS[*]}${NC}"
}

# Vérifications préalables
check_requirements() {
    if ! command -v incus &> /dev/null; then
        echo -e "${RED}❌ Incus n'est pas installé${NC}"
        exit 1
    fi
    
    if ! incus info >/dev/null 2>&1; then
        echo -e "${RED}❌ Incus n'est pas initialisé${NC}"
        echo -e "${YELLOW}💡 Exécutez: sudo incus admin init${NC}"
        exit 1
    fi
}

# Configuration réseau optimisée
setup_network() {
    echo -e "${BLUE}🌐 Configuration réseau optimisée...${NC}"
    
    # Utiliser le réseau par défaut d'Incus (incusbr0)
    local default_bridge="incusbr0"
    
    if ! incus network show "$default_bridge" >/dev/null 2>&1; then
        echo -e "${RED}❌ Réseau par défaut Incus introuvable${NC}"
        echo -e "${YELLOW}💡 Exécutez: sudo incus admin init${NC}"
        exit 1
    fi
    
    # Afficher les informations réseau
    local network_info
    network_info=$(incus network get "$default_bridge" ipv4.address 2>/dev/null || echo "DHCP automatique")
    echo -e "${GREEN}✅ Réseau par défaut Incus: $default_bridge ($network_info)${NC}"
    
    # Supprimer tout réseau veza-network obsolète s'il existe
    if incus network show veza-network >/dev/null 2>&1; then
        echo -e "${YELLOW}🗑️ Suppression réseau obsolète veza-network...${NC}"
        
        # Détacher containers du réseau obsolète
        local attached_containers
        attached_containers=$(incus network show veza-network | grep "used_by:" -A 20 | grep "/instances/" | cut -d'/' -f4 2>/dev/null || true)
        
        for container in $attached_containers; do
            if [ -n "$container" ]; then
                echo -e "${CYAN}🔄 Détachement de $container...${NC}"
                incus stop "$container" --force || true
            fi
        done
        
        incus network delete veza-network || true
        echo -e "${GREEN}✅ Réseau obsolète supprimé${NC}"
    fi
    
    # Configuration système pour optimiser la connectivité
    echo -e "${BLUE}🔧 Optimisation système...${NC}"
    
    # IPv4 forwarding (essentiel pour les containers)
    echo 'net.ipv4.ip_forward=1' | sudo tee /etc/sysctl.d/99-incus-forwarding.conf > /dev/null
    sudo sysctl -p /etc/sysctl.d/99-incus-forwarding.conf >/dev/null 2>&1 || true
    
    echo -e "${GREEN}✅ Configuration réseau terminée - Tous les containers utiliseront incusbr0${NC}"
    return 0
}

# Configuration des profils
setup_profiles() {
    echo -e "${BLUE}👤 Configuration des profils...${NC}"
    
    # Vérifier que le profil default existe (toujours présent dans Incus)
    if ! incus profile show default >/dev/null 2>&1; then
        echo -e "${RED}❌ Profil default introuvable - problème Incus${NC}"
        exit 1
    fi
    
    # Profil de base Veza (optionnel, pour configurations spécifiques futures)
    if ! incus profile show veza-base >/dev/null 2>&1; then
        echo -e "${BLUE}📋 Création du profil veza-base...${NC}"
        incus profile create veza-base
        incus profile copy default veza-base --force
        
        # Configuration optimisée de base
        incus profile set veza-base limits.cpu 2
        incus profile set veza-base limits.memory 2GB
        incus profile set veza-base security.nesting true
        incus profile set veza-base security.privileged false
        
        echo -e "${GREEN}✅ Profil veza-base créé${NC}"
    else
        echo -e "${CYAN}ℹ️ Profil veza-base existe déjà${NC}"
    fi
    
    echo -e "${GREEN}✅ Tous les containers utiliseront le profil 'default' pour maximum de compatibilité${NC}"
}

# Déploiement complet avec les scripts existants
deploy_infrastructure() {
    echo -e "${BLUE}🚀 Déploiement de l'infrastructure complète...${NC}"
    
    # 1. Créer les containers
    if [ -f "$SCRIPT_DIR/setup-manual-containers.sh" ]; then
        echo -e "${CYAN}📦 Création des containers...${NC}"
        bash "$SCRIPT_DIR/setup-manual-containers.sh"
    else
        echo -e "${RED}❌ Script setup-manual-containers.sh introuvable${NC}"
        exit 1
    fi
    
    # 2. Configurer les services systemd
    if [ -f "$SCRIPT_DIR/setup-systemd-services.sh" ]; then
        echo -e "${CYAN}⚙️ Configuration services systemd...${NC}"
        bash "$SCRIPT_DIR/setup-systemd-services.sh"
    else
        echo -e "${YELLOW}⚠️ Script setup-systemd-services.sh introuvable${NC}"
    fi
    
    # 3. Configurer rsync et SSH
    if [ -f "$SCRIPT_DIR/setup-rsync.sh" ]; then
        echo -e "${CYAN}🔄 Configuration rsync et SSH...${NC}"
        bash "$SCRIPT_DIR/setup-rsync.sh"
    else
        echo -e "${YELLOW}⚠️ Script setup-rsync.sh introuvable${NC}"
    fi
    
    # 4. Setup complet final
    if [ -f "$SCRIPT_DIR/complete-setup.sh" ]; then
        echo -e "${CYAN}🏗️ Setup complet final...${NC}"
        bash "$SCRIPT_DIR/complete-setup.sh"
    else
        echo -e "${YELLOW}⚠️ Script complete-setup.sh introuvable${NC}"
    fi
    
    echo -e "${GREEN}🎉 Déploiement terminé${NC}"
}

# État complet de l'infrastructure
show_status() {
    show_header
    echo -e "${BLUE}📊 État de l'infrastructure Veza${NC}"
    echo ""
    
    # État des containers
    echo -e "${CYAN}🏠 Containers:${NC}"
    incus ls --format=table --columns=n,s,4,6
    echo ""
    
    # État des services
    echo -e "${CYAN}⚙️ Services:${NC}"
    printf "%-15s %-15s %-15s\n" "Service" "Container" "État"
    printf "%-15s %-15s %-15s\n" "-------" "---------" "----"
    
    for service in "${!CONTAINERS[@]}"; do
        local container="${CONTAINERS[$service]}"
        local service_name="${SERVICES[$service]}"
        
        if incus list "$container" --format csv | grep -q RUNNING; then
            if incus exec "$container" -- systemctl is-active "$service_name" >/dev/null 2>&1; then
                status="${GREEN}✅ ACTIF${NC}"
            else
                status="${YELLOW}⏸️ ARRÊTÉ${NC}"
            fi
        else
            status="${RED}❌ CONTAINER ARRÊTÉ${NC}"
        fi
        
        printf "%-15s %-15s %s\n" "$service" "$container" "$status"
    done
    echo ""
    
    # IPs des containers
    echo -e "${CYAN}🌐 Adresses IP:${NC}"
    for service in "${!CONTAINERS[@]}"; do
        local container="${CONTAINERS[$service]}"
        if incus list "$container" --format csv | grep -q RUNNING; then
            local ip=$(incus list "$container" -c 4 --format csv | head -1)
            printf "%-15s %-15s %s\n" "$service" "$container" "$ip"
        fi
    done
}

# Vérification de santé complète
health_check() {
    echo -e "${BLUE}🏥 Vérification de santé de l'infrastructure${NC}"
    echo ""
    
    # Test containers actifs
    echo -e "${CYAN}📦 Containers:${NC}"
    local containers_ok=0
    local containers_total=0
    
    # Liste des containers dans l'ordre
    local container_list=("veza-postgres" "veza-redis" "veza-storage" "veza-backend" "veza-chat" "veza-stream" "veza-frontend" "veza-haproxy")
    
    for container in "${container_list[@]}"; do
        ((containers_total++))
        
        if incus list "$container" --format csv 2>/dev/null | grep -q RUNNING; then
            echo -e "  ${GREEN}✅ $container running${NC}"
            ((containers_ok++))
        else
            echo -e "  ${RED}❌ $container stopped${NC}"
        fi
    done
    
    echo -e "  📊 Total: $containers_ok/$containers_total containers actifs"
    echo ""
    
    # Test services
    echo -e "${CYAN}⚙️ Services:${NC}"
    local services_ok=0
    local services_total=0
    
    # Services correspondants
    local service_names=("postgresql" "redis-server" "nfs-kernel-server" "veza-backend" "veza-chat" "veza-stream" "veza-frontend" "haproxy")
    local service_labels=("postgres" "redis" "storage" "backend" "chat" "stream" "frontend" "haproxy")
    
    for i in "${!container_list[@]}"; do
        local container="${container_list[$i]}"
        local service_name="${service_names[$i]}"
        local service_label="${service_labels[$i]}"
        ((services_total++))
        
        if incus list "$container" --format csv 2>/dev/null | grep -q RUNNING; then
            if incus exec "$container" -- systemctl is-active "$service_name" >/dev/null 2>&1; then
                echo -e "  ${GREEN}✅ $service_label actif${NC}"
                ((services_ok++))
            else
                echo -e "  ${RED}❌ $service_label inactif${NC}"
            fi
        else
            echo -e "  ${YELLOW}⚠️ $service_label (container arrêté)${NC}"
        fi
    done
    
    echo -e "  📊 Total: $services_ok/$services_total services actifs"
    echo ""
    
    # Test connectivité réseau
    echo -e "${CYAN}🌐 Connectivité:${NC}"
    
    # Test PostgreSQL depuis backend
    if incus list veza-backend --format csv 2>/dev/null | grep -q RUNNING && incus list veza-postgres --format csv 2>/dev/null | grep -q RUNNING; then
        local postgres_ip=$(incus list veza-postgres -c 4 --format csv 2>/dev/null | head -1 | cut -d' ' -f1)
        if incus exec veza-backend -- timeout 5 nc -z "$postgres_ip" 5432 >/dev/null 2>&1; then
            echo -e "  ${GREEN}✅ Backend → PostgreSQL${NC}"
        else
            echo -e "  ${RED}❌ Backend → PostgreSQL${NC}"
        fi
    fi
    
    # Test Redis depuis backend
    if incus list veza-backend --format csv 2>/dev/null | grep -q RUNNING && incus list veza-redis --format csv 2>/dev/null | grep -q RUNNING; then
        local redis_ip=$(incus list veza-redis -c 4 --format csv 2>/dev/null | head -1 | cut -d' ' -f1)
        if incus exec veza-backend -- timeout 5 nc -z "$redis_ip" 6379 >/dev/null 2>&1; then
            echo -e "  ${GREEN}✅ Backend → Redis${NC}"
        else
            echo -e "  ${RED}❌ Backend → Redis${NC}"
        fi
    fi
    
    # Test internet depuis containers
    local internet_ok=0
    local internet_total=0
    for container in "${container_list[@]}"; do
        if incus list "$container" --format csv 2>/dev/null | grep -q RUNNING; then
            ((internet_total++))
            if incus exec "$container" -- timeout 3 ping -c 1 8.8.8.8 >/dev/null 2>&1; then
                ((internet_ok++))
            fi
        fi
    done
    
    if [ $internet_ok -eq $internet_total ] && [ $internet_total -gt 0 ]; then
        echo -e "  ${GREEN}✅ Connectivité internet ($internet_ok/$internet_total)${NC}"
    elif [ $internet_ok -gt 0 ]; then
        echo -e "  ${YELLOW}⚠️ Connectivité internet partielle ($internet_ok/$internet_total)${NC}"
    else
        echo -e "  ${RED}❌ Pas de connectivité internet${NC}"
    fi
}

# Gestion des services
manage_service() {
    local action=$1
    local service_target=$2
    
    if [ -z "$service_target" ]; then
        # Action sur tous les services
        echo -e "${BLUE}🔧 ${action^} de tous les services...${NC}"
        
        for service in "${!CONTAINERS[@]}"; do
            local container="${CONTAINERS[$service]}"
            local service_name="${SERVICES[$service]}"
            
            if incus list "$container" --format csv | grep -q RUNNING; then
                echo -e "${CYAN}$action $service...${NC}"
                case $action in
                    "start")
                        incus exec "$container" -- systemctl start "$service_name" || echo -e "${RED}❌ Échec${NC}"
                        ;;
                    "stop")
                        incus exec "$container" -- systemctl stop "$service_name" || echo -e "${RED}❌ Échec${NC}"
                        ;;
                    "restart")
                        incus exec "$container" -- systemctl restart "$service_name" || echo -e "${RED}❌ Échec${NC}"
                        ;;
                esac
            else
                echo -e "${YELLOW}⚠️ Container $container non running${NC}"
            fi
        done
    else
        # Action sur un service spécifique
        if [[ -n "${CONTAINERS[$service_target]}" ]]; then
            local container="${CONTAINERS[$service_target]}"
            local service_name="${SERVICES[$service_target]}"
            
            echo -e "${CYAN}$action $service_target ($service_name dans $container)...${NC}"
            
            if incus list "$container" --format csv | grep -q RUNNING; then
                case $action in
                    "start")
                        incus exec "$container" -- systemctl start "$service_name"
                        echo -e "${GREEN}✅ $service_target démarré${NC}"
                        ;;
                    "stop")
                        incus exec "$container" -- systemctl stop "$service_name"
                        echo -e "${GREEN}✅ $service_target arrêté${NC}"
                        ;;
                    "restart")
                        incus exec "$container" -- systemctl restart "$service_name"
                        echo -e "${GREEN}✅ $service_target redémarré${NC}"
                        ;;
                esac
            else
                echo -e "${RED}❌ Container $container non running${NC}"
                exit 1
            fi
        else
            echo -e "${RED}❌ Service '$service_target' inconnu${NC}"
            echo -e "${CYAN}Services disponibles: ${!CONTAINERS[*]}${NC}"
            exit 1
        fi
    fi
}

# Affichage des logs
show_logs() {
    local service_target=$1
    local follow_mode=${2:-false}
    
    if [[ -z "$service_target" ]]; then
        echo -e "${RED}❌ Service requis pour les logs${NC}"
        echo -e "${CYAN}Services disponibles: ${!CONTAINERS[*]}${NC}"
        exit 1
    fi
    
    if [[ -n "${CONTAINERS[$service_target]}" ]]; then
        local container="${CONTAINERS[$service_target]}"
        local service_name="${SERVICES[$service_target]}"
        
        echo -e "${BLUE}📋 Logs de $service_target ($service_name)${NC}"
        echo -e "${YELLOW}Container: $container${NC}"
        echo ""
        
        if incus list "$container" --format csv | grep -q RUNNING; then
            if [ "$follow_mode" = "follow" ]; then
                echo -e "${CYAN}📡 Suivi en temps réel (Ctrl+C pour arrêter)${NC}"
                incus exec "$container" -- journalctl -u "$service_name" -f --no-pager
            else
                echo -e "${CYAN}📖 Dernières 50 lignes${NC}"
                incus exec "$container" -- journalctl -u "$service_name" -n 50 --no-pager
            fi
        else
            echo -e "${RED}❌ Container $container non running${NC}"
            exit 1
        fi
    else
        echo -e "${RED}❌ Service '$service_target' inconnu${NC}"
        echo -e "${CYAN}Services disponibles: ${!CONTAINERS[*]}${NC}"
        exit 1
    fi
}

# Nettoyage complet
clean_infrastructure() {
    echo -e "${RED}⚠️ ATTENTION: Suppression complète de l'infrastructure${NC}"
    echo -e "${YELLOW}Voulez-vous continuer ? (oui/NON)${NC}"
    read -r response
    
    if [[ "$response" != "oui" ]]; then
        echo -e "${GREEN}Nettoyage annulé${NC}"
        return 0
    fi
    
    echo -e "${BLUE}🧹 Nettoyage en cours...${NC}"
    
    # Arrêter et supprimer tous les containers
    for service in "${!CONTAINERS[@]}"; do
        local container="${CONTAINERS[$service]}"
        if incus list "$container" --format csv | grep -q "RUNNING\|STOPPED"; then
            echo -e "${YELLOW}🗑️ Suppression de $container...${NC}"
            incus stop "$container" --force || true
            incus delete "$container" || true
        fi
    done
    
    # Supprimer les profils
    for profile in veza-database veza-app veza-base; do
        if incus profile show "$profile" >/dev/null 2>&1; then
            echo -e "${YELLOW}👤 Suppression profil $profile...${NC}"
            incus profile delete "$profile" || true
        fi
    done
    
    echo -e "${GREEN}🎉 Nettoyage terminé${NC}"
}

# Export des containers de base
export_base_containers() {
    echo -e "${BLUE}📦 Export des containers de base...${NC}"
    
    local export_dir="$WORKSPACE_DIR/containers-export"
    mkdir -p "$export_dir"
    
    for service in "${!CONTAINERS[@]}"; do
        local container="${CONTAINERS[$service]}"
        if incus list "$container" --format csv | grep -q "RUNNING\|STOPPED"; then
            echo -e "${CYAN}📤 Export de $container...${NC}"
            incus export "$container" "$export_dir/${container}-base.tar.gz"
            echo -e "${GREEN}✅ $container exporté${NC}"
        fi
    done
    
    echo -e "${GREEN}🎉 Export terminé dans: $export_dir${NC}"
}

# Import des containers de base
import_base_containers() {
    echo -e "${BLUE}📦 Import des containers de base...${NC}"
    
    local export_dir="$WORKSPACE_DIR/containers-export"
    
    if [ ! -d "$export_dir" ]; then
        echo -e "${RED}❌ Répertoire d'export introuvable: $export_dir${NC}"
        exit 1
    fi
    
    for service in "${!CONTAINERS[@]}"; do
        local container="${CONTAINERS[$service]}"
        local export_file="$export_dir/${container}-base.tar.gz"
        
        if [ -f "$export_file" ]; then
            if ! incus list "$container" --format csv | grep -q "RUNNING\|STOPPED"; then
                echo -e "${CYAN}📥 Import de $container...${NC}"
                incus import "$export_file" "$container"
                echo -e "${GREEN}✅ $container importé${NC}"
            else
                echo -e "${YELLOW}⚠️ $container existe déjà${NC}"
            fi
        fi
    done
}

# Synchronisation du code source
sync_code() {
    echo -e "${BLUE}🔄 Synchronisation du code source...${NC}"
    
    if [ -f "$SCRIPT_DIR/quick-sync.sh" ]; then
        bash "$SCRIPT_DIR/quick-sync.sh"
    else
        echo -e "${RED}❌ Script quick-sync.sh introuvable${NC}"
        exit 1
    fi
}

# Surveillance automatique
watch_code() {
    echo -e "${BLUE}👁️ Démarrage de la surveillance automatique...${NC}"
    echo -e "${YELLOW}Ctrl+C pour arrêter${NC}"
    
    if [ -f "$SCRIPT_DIR/watch-and-sync.sh" ]; then
        bash "$SCRIPT_DIR/watch-and-sync.sh"
    else
        echo -e "${RED}❌ Script watch-and-sync.sh introuvable${NC}"
        exit 1
    fi
}

# Compilation des projets
build_projects() {
    local target=${1:-all}
    
    echo -e "${BLUE}🔨 Compilation des projets...${NC}"
    
    if [ "$target" = "all" ]; then
        echo -e "${CYAN}🔧 Compilation de tous les projets...${NC}"
        
        # Backend Go
        echo -e "${YELLOW}📦 Backend Go...${NC}"
        incus exec veza-backend -- bash -c "
            cd /opt/veza/backend
            export PATH=/usr/local/go/bin:\$PATH
            export GOPATH=/opt/veza/go
            go mod tidy
            go build -o main ./cmd/server/main.go
            echo 'Backend Go compilé ✅'
        "
        
        # Chat Rust
        echo -e "${YELLOW}💬 Chat Rust...${NC}"
        incus exec veza-chat -- bash -c "
            cd /opt/veza/chat
            export PATH=/root/.cargo/bin:\$PATH
            source /root/.cargo/env
            cargo build --release
            echo 'Chat Rust compilé ✅'
        "
        
        # Stream Rust
        echo -e "${YELLOW}🎵 Stream Rust...${NC}"
        incus exec veza-stream -- bash -c "
            cd /opt/veza/stream
            export PATH=/root/.cargo/bin:\$PATH
            source /root/.cargo/env
            cargo build --release
            echo 'Stream Rust compilé ✅'
        "
        
        # Frontend React
        echo -e "${YELLOW}🎨 Frontend React...${NC}"
        incus exec veza-frontend -- bash -c "
            cd /opt/veza/frontend
            export PATH=/usr/local/node/bin:\$PATH
            npm install
            npm run build
            echo 'Frontend React compilé ✅'
        "
        
    elif [ "$target" = "backend" ]; then
        echo -e "${CYAN}🔧 Compilation Backend Go...${NC}"
        incus exec veza-backend -- bash -c "
            cd /opt/veza/backend
            export PATH=/usr/local/go/bin:\$PATH
            export GOPATH=/opt/veza/go
            go mod tidy
            go build -o main ./cmd/server/main.go
        "
        
    elif [ "$target" = "chat" ]; then
        echo -e "${CYAN}💬 Compilation Chat Rust...${NC}"
        incus exec veza-chat -- bash -c "
            cd /opt/veza/chat
            export PATH=/root/.cargo/bin:\$PATH
            source /root/.cargo/env
            cargo build --release
        "
        
    elif [ "$target" = "stream" ]; then
        echo -e "${CYAN}🎵 Compilation Stream Rust...${NC}"
        incus exec veza-stream -- bash -c "
            cd /opt/veza/stream
            export PATH=/root/.cargo/bin:\$PATH
            source /root/.cargo/env
            cargo build --release
        "
        
    elif [ "$target" = "frontend" ]; then
        echo -e "${CYAN}🎨 Compilation Frontend React...${NC}"
        incus exec veza-frontend -- bash -c "
            cd /opt/veza/frontend
            export PATH=/usr/local/node/bin:\$PATH
            npm install
            npm run build
        "
        
    else
        echo -e "${RED}❌ Cible de build inconnue: $target${NC}"
        echo -e "${CYAN}Cibles disponibles: all, backend, chat, stream, frontend${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}✅ Compilation terminée${NC}"
}

# Build complet + démarrage
build_and_start() {
    echo -e "${BLUE}🚀 Build complet + démarrage des services...${NC}"
    
    if [ -f "$SCRIPT_DIR/build-and-start.sh" ]; then
        bash "$SCRIPT_DIR/build-and-start.sh"
    else
        echo -e "${YELLOW}📝 Utilisation du build intégré...${NC}"
        build_projects all
        sleep 3
        manage_service start
    fi
}

# Réparation des dépendances
fix_dependencies() {
    echo -e "${BLUE}🔧 Réparation des dépendances...${NC}"
    
    # Installation des dépendances de base
    if [ -f "$SCRIPT_DIR/install-dependencies.sh" ]; then
        echo -e "${CYAN}📦 Installation dépendances principales...${NC}"
        bash "$SCRIPT_DIR/install-dependencies.sh"
    fi
    
    # Réparation Rust OpenSSL
    if [ -f "$SCRIPT_DIR/fix-rust-dependencies.sh" ]; then
        echo -e "${CYAN}🦀 Réparation dépendances Rust...${NC}"
        bash "$SCRIPT_DIR/fix-rust-dependencies.sh"
    fi
    
    # Réparation HAProxy
    if [ -f "$SCRIPT_DIR/fix-haproxy.sh" ]; then
        echo -e "${CYAN}⚖️ Réparation HAProxy...${NC}"
        bash "$SCRIPT_DIR/fix-haproxy.sh"
    fi
    
    echo -e "${GREEN}✅ Réparation des dépendances terminée${NC}"
}

# Initialisation de la base de données
init_database() {
    echo -e "${BLUE}🗄️ Initialisation de la base de données PostgreSQL...${NC}"
    
    if [ -f "$SCRIPT_DIR/setup-database.sh" ]; then
        bash "$SCRIPT_DIR/setup-database.sh"
    else
        echo -e "${RED}❌ Script setup-database.sh introuvable${NC}"
        exit 1
    fi
}

# Fonction principale
main() {
    local command=${1:-help}
    local target=${2:-}
    local option=${3:-}
    
    check_requirements
    
    case "$command" in
        "help"|"-h"|"--help")
            show_help
            ;;
        "setup")
            show_header
            echo -e "${BLUE}🚀 Configuration initiale complète...${NC}"
            setup_network
            setup_profiles
            echo -e "${GREEN}🎉 Setup terminé${NC}"
            ;;
        "deploy")
            show_header
            deploy_infrastructure
            ;;
        "status")
            show_status
            ;;
        "health")
            show_header
            health_check
            ;;
        "start"|"stop"|"restart")
            manage_service "$command" "$target"
            ;;
        "logs")
            if [[ "$option" = "follow" || "$option" = "-f" ]]; then
                show_logs "$target" "follow"
            else
                show_logs "$target"
            fi
            ;;
        "build")
            show_header
            build_projects "$target"
            ;;
        "sync")
            show_header
            sync_code
            ;;
        "watch")
            show_header
            watch_code
            ;;
        "build-start")
            show_header
            build_and_start
            ;;
        "fix-deps")
            show_header
            fix_dependencies
            ;;
        "export")
            show_header
            export_base_containers
            ;;
        "import")
            show_header
            import_base_containers
            ;;
        "clean")
            show_header
            clean_infrastructure
            ;;
        "network-fix")
            show_header
            if [ -f "$SCRIPT_DIR/network-fix.sh" ]; then
                bash "$SCRIPT_DIR/network-fix.sh"
            else
                echo -e "${RED}❌ Script network-fix.sh introuvable${NC}"
                exit 1
            fi
            ;;
        "update")
            show_header
            if [ -f "$SCRIPT_DIR/update-source-code.sh" ]; then
                bash "$SCRIPT_DIR/update-source-code.sh" "${target:-all}"
            else
                echo -e "${RED}❌ Script update-source-code.sh introuvable${NC}"
                exit 1
            fi
            ;;
        "init-db")
            show_header
            init_database
            ;;
        *)
            echo -e "${RED}❌ Commande inconnue: $command${NC}"
            show_help
            exit 1
            ;;
    esac
}

# Exécution
main "$@" 