#!/bin/bash

# Script de nettoyage complet Incus pour Veza
# Supprime tous les containers, profils, réseaux et stockage

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
echo "│      🧹 Veza - Nettoyage Ultra-Complet  │"
echo "╰──────────────────────────────────────────╯"
echo -e "${NC}"

# Fonction pour supprimer tous les containers Veza
clean_containers() {
    echo -e "${BLUE}🗑️  Suppression des containers...${NC}"
    
    # Liste des containers Veza
    local containers=("veza-frontend" "veza-backend" "veza-chat" "veza-stream" "veza-postgres" "veza-haproxy" "veza-redis" "veza-storage")
    
    for container in "${containers[@]}"; do
        if incus ls --format csv | grep -q "^$container,"; then
            echo -e "${YELLOW}Arrêt et suppression de $container...${NC}"
            incus stop "$container" --force || true
            incus delete "$container" || true
        fi
    done
    
    # Supprimer tous les autres containers qui pourraient exister (sauf default)
    while IFS=',' read -r name state; do
        if [[ -n "$name" && "$name" != "NAME" && "$name" != "default" ]]; then
            echo -e "${YELLOW}Suppression du container restant: $name...${NC}"
            incus stop "$name" --force || true
            incus delete "$name" || true
        fi
    done < <(incus ls --format csv)
}

# Fonction pour supprimer les profils
clean_profiles() {
    echo -e "${BLUE}👤 Suppression des profils...${NC}"
    
    local profiles=("veza-storage" "veza-database" "veza-app" "veza-base")
    
    for profile in "${profiles[@]}"; do
        if incus profile show "$profile" >/dev/null 2>&1; then
            echo -e "${YELLOW}Suppression du profil $profile...${NC}"
            incus profile delete "$profile" || true
        fi
    done
}

# Fonction pour supprimer les volumes de stockage
clean_storage_volumes() {
    echo -e "${BLUE}💾 Suppression des volumes de stockage...${NC}"
    
    if incus storage show veza-zfs-pool >/dev/null 2>&1; then
        local volumes=("uploads" "audio" "backups" "zfs-cache")
        
        for volume in "${volumes[@]}"; do
            if incus storage volume show veza-zfs-pool "$volume" >/dev/null 2>&1; then
                echo -e "${YELLOW}Suppression du volume $volume...${NC}"
                incus storage volume delete veza-zfs-pool "$volume" || true
            fi
        done
    fi
}

# Fonction pour supprimer le stockage
clean_storage() {
    echo -e "${BLUE}🗄️  Suppression du stockage...${NC}"
    
    if incus storage show veza-zfs-pool >/dev/null 2>&1; then
        echo -e "${YELLOW}Suppression du pool veza-zfs-pool...${NC}"
        incus storage delete veza-zfs-pool || true
    fi
}

# Fonction pour supprimer le réseau et nettoyer la configuration
clean_network() {
    echo -e "${BLUE}🌐 Nettoyage complet du réseau...${NC}"
    
    # Arrêter le réseau veza-network
    if incus network show veza-network >/dev/null 2>&1; then
        echo -e "${YELLOW}Arrêt du réseau veza-network...${NC}"
        
        # Lister les containers attachés et les détacher
        local attached_containers
        attached_containers=$(incus network show veza-network | grep -A 10 "used_by:" | grep -E "^\s*-" | sed 's/^\s*-\s*//' | grep "^/1.0/instances" | cut -d'/' -f4 || true)
        
        if [ -n "$attached_containers" ]; then
            echo -e "${YELLOW}Détachement des containers du réseau...${NC}"
            for container in $attached_containers; do
                echo -e "${CYAN}Détachement de $container...${NC}"
                incus stop "$container" --force 2>/dev/null || true
            done
        fi
        
        # Supprimer le réseau
        echo -e "${YELLOW}Suppression du réseau veza-network...${NC}"
        incus network delete veza-network || true
    fi
    
    # Nettoyer les règles iptables liées à Incus (optionnel)
    echo -e "${BLUE}🔧 Nettoyage des règles réseau hôte...${NC}"
    
    # Supprimer les règles iptables personnalisées Veza (si elles existent)
    sudo iptables -t nat -D POSTROUTING -s 10.100.0.0/24 -j MASQUERADE 2>/dev/null || true
    sudo iptables -D FORWARD -i veza-network -j ACCEPT 2>/dev/null || true
    sudo iptables -D FORWARD -o veza-network -j ACCEPT 2>/dev/null || true
    
    # Nettoyer la configuration DNS hôte si elle a été modifiée
    if [ -f "/etc/systemd/resolved.conf.d/incus.conf" ]; then
        echo -e "${YELLOW}Suppression configuration DNS hôte...${NC}"
        sudo rm -f /etc/systemd/resolved.conf.d/incus.conf
        sudo systemctl restart systemd-resolved 2>/dev/null || true
    fi
    
    # Nettoyer les configurations sysctl
    if [ -f "/etc/sysctl.d/99-incus-forwarding.conf" ]; then
        echo -e "${YELLOW}Nettoyage configuration sysctl...${NC}"
        sudo rm -f /etc/sysctl.d/99-incus-forwarding.conf
        sudo rm -f /etc/sysctl.d/99-incus.conf
    fi
}

# Fonction pour vérifier et nettoyer les processus réseau orphelins
clean_network_processes() {
    echo -e "${BLUE}🔍 Nettoyage des processus réseau orphelins...${NC}"
    
    # Arrêter les processus dnsmasq liés à veza-network (si ils existent)
    local dnsmasq_pids
    dnsmasq_pids=$(pgrep -f "dnsmasq.*veza-network" 2>/dev/null || true)
    
    if [ -n "$dnsmasq_pids" ]; then
        echo -e "${YELLOW}Arrêt des processus dnsmasq orphelins...${NC}"
        echo "$dnsmasq_pids" | xargs sudo kill -TERM 2>/dev/null || true
        sleep 2
        # Force kill si nécessaire
        dnsmasq_pids=$(pgrep -f "dnsmasq.*veza-network" 2>/dev/null || true)
        if [ -n "$dnsmasq_pids" ]; then
            echo "$dnsmasq_pids" | xargs sudo kill -KILL 2>/dev/null || true
        fi
    fi
}

# Fonction pour nettoyer les répertoires locaux
clean_local_directories() {
    echo -e "${BLUE}📁 Nettoyage des répertoires locaux...${NC}"
    
    # Demander confirmation avant de supprimer les données
    echo -e "${YELLOW}⚠️ Voulez-vous aussi supprimer les données locales ? (logs, data, uploads, etc.) (o/N)${NC}"
    read -r response
    
    if [[ "$response" == "o" || "$response" == "oui" ]]; then
        echo -e "${YELLOW}Suppression des répertoires de données...${NC}"
        
        # Nettoyer les répertoires de données
        rm -rf data/postgres/* 2>/dev/null || true
        rm -rf data/redis/* 2>/dev/null || true
        rm -rf logs/* 2>/dev/null || true
        rm -rf uploads/* 2>/dev/null || true
        rm -rf audio/* 2>/dev/null || true
        rm -rf backups/* 2>/dev/null || true
        
        echo -e "${GREEN}✅ Répertoires de données nettoyés${NC}"
    else
        echo -e "${CYAN}ℹ️ Répertoires de données préservés${NC}"
    fi
}

# Fonction pour réinitialiser complètement Incus (optionnel)
reset_incus_completely() {
    echo -e "${BLUE}🔄 Option de réinitialisation complète d'Incus...${NC}"
    echo -e "${RED}⚠️ ATTENTION: Ceci va supprimer TOUTE la configuration Incus !${NC}"
    echo -e "${YELLOW}Voulez-vous réinitialiser complètement Incus ? (oui/NON)${NC}"
    read -r response
    
    if [[ "$response" == "oui" ]]; then
        echo -e "${RED}🚨 Réinitialisation complète d'Incus...${NC}"
        
        # Arrêter tous les containers
        incus list --format csv | cut -d, -f1 | grep -v "^NAME$" | xargs -I {} incus stop {} --force 2>/dev/null || true
        
        # Supprimer tous les containers
        incus list --format csv | cut -d, -f1 | grep -v "^NAME$" | xargs -I {} incus delete {} 2>/dev/null || true
        
        # Supprimer tous les profils personnalisés
        incus profile list --format csv | cut -d, -f1 | grep -v "^NAME$" | grep -v "^default$" | xargs -I {} incus profile delete {} 2>/dev/null || true
        
        # Supprimer tous les réseaux personnalisés
        incus network list --format csv | cut -d, -f1 | grep -v "^NAME$" | grep -v "^incusbr0$" | xargs -I {} incus network delete {} 2>/dev/null || true
        
        # Supprimer tous les pools de stockage personnalisés
        incus storage list --format csv | cut -d, -f1 | grep -v "^NAME$" | grep -v "^incus_storage$" | xargs -I {} incus storage delete {} 2>/dev/null || true
        
        echo -e "${GREEN}✅ Incus complètement réinitialisé${NC}"
        echo -e "${CYAN}💡 Vous devrez reconfigurer Incus avec: sudo incus admin init${NC}"
    else
        echo -e "${GREEN}Réinitialisation complète annulée${NC}"
    fi
}

# Fonction de vérification finale
final_verification() {
    echo -e "${BLUE}🔍 Vérification finale de la suppression...${NC}"
    
    echo -e "${CYAN}📊 État final :${NC}"
    echo ""
    echo -e "${BLUE}Containers restants :${NC}"
    if incus ls --format csv | grep -q "veza-"; then
        echo -e "${RED}⚠️ Containers Veza trouvés :${NC}"
        incus ls | grep veza- || true
    else
        echo -e "${GREEN}✅ Aucun container Veza restant${NC}"
    fi
    
    echo ""
    echo -e "${BLUE}Profils restants :${NC}"
    if incus profile list --format csv | grep -q "veza-"; then
        echo -e "${RED}⚠️ Profils Veza trouvés :${NC}"
        incus profile list | grep veza- || true
    else
        echo -e "${GREEN}✅ Aucun profil Veza restant${NC}"
    fi
    
    echo ""
    echo -e "${BLUE}Réseaux restants :${NC}"
    if incus network list --format csv | grep -q "veza-"; then
        echo -e "${RED}⚠️ Réseaux Veza trouvés :${NC}"
        incus network list | grep veza- || true
    else
        echo -e "${GREEN}✅ Aucun réseau Veza restant${NC}"
    fi
    
    echo ""
    echo -e "${BLUE}Stockage restant :${NC}"
    if incus storage list --format csv | grep -q "veza-"; then
        echo -e "${RED}⚠️ Stockage Veza trouvé :${NC}"
        incus storage list | grep veza- || true
    else
        echo -e "${GREEN}✅ Aucun stockage Veza restant${NC}"
    fi
}

# Fonction principale de nettoyage
main() {
    echo -e "${RED}⚠️  ATTENTION: Cette opération va supprimer TOUS les containers et données Veza !${NC}"
    echo -e "${YELLOW}Types de nettoyage disponibles :${NC}"
    echo -e "  1. ${CYAN}Nettoyage standard${NC} - Containers, profils, réseau, stockage"
    echo -e "  2. ${YELLOW}Nettoyage complet${NC} - Standard + données locales"
    echo -e "  3. ${RED}Nettoyage ultra-complet${NC} - Complet + réinitialisation Incus"
    echo -e "  4. ${GREEN}Annuler${NC}"
    echo ""
    echo -e "${YELLOW}Votre choix (1-4) :${NC}"
    read -r choice
    
    case $choice in
        1)
            echo -e "${BLUE}🚀 Début du nettoyage standard...${NC}"
            ;;
        2)
            echo -e "${BLUE}🚀 Début du nettoyage complet...${NC}"
            ;;
        3)
            echo -e "${BLUE}🚀 Début du nettoyage ultra-complet...${NC}"
            ;;
        4|*)
            echo -e "${GREEN}Opération annulée${NC}"
            exit 0
            ;;
    esac
    
    # Nettoyage dans l'ordre correct pour éviter les dépendances
    clean_containers
    clean_network_processes
    clean_profiles
    clean_storage_volumes
    clean_storage
    clean_network
    
    # Nettoyage des données locales si choix 2 ou 3
    if [[ "$choice" == "2" || "$choice" == "3" ]]; then
        clean_local_directories
    fi
    
    # Réinitialisation complète si choix 3
    if [[ "$choice" == "3" ]]; then
        reset_incus_completely
    fi
    
    # Vérification finale
    final_verification
    
    echo ""
    echo -e "${GREEN}🎉 Nettoyage terminé avec succès !${NC}"
    echo ""
    echo -e "${CYAN}💡 Prochaines étapes suggérées :${NC}"
    echo -e "  • ${YELLOW}Reconfigurer l'infrastructure :${NC} ./scripts/incus-setup.sh"
    echo -e "  • ${YELLOW}Redéployer les services :${NC} ./scripts/incus-deploy.sh"
    echo -e "  • ${YELLOW}Vérifier l'état :${NC} ./scripts/incus-status.sh"
}

# Vérifier qu'Incus est disponible
if ! command -v incus &> /dev/null; then
    echo -e "${RED}❌ Incus n'est pas installé ou accessible${NC}"
    exit 1
fi

# Exécuter le nettoyage
main 