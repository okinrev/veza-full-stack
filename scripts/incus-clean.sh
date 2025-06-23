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
NC='\033[0m'

echo -e "${PURPLE}"
echo "╭──────────────────────────────────────────╮"
echo "│      🧹 Veza - Nettoyage Complet        │"
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
    
    # Supprimer tous les autres containers qui pourraient exister
    while IFS=',' read -r name state; do
        if [[ -n "$name" && "$name" != "NAME" ]]; then
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

# Fonction pour supprimer le réseau
clean_network() {
    echo -e "${BLUE}🌐 Suppression du réseau...${NC}"
    
    if incus network show veza-network >/dev/null 2>&1; then
        echo -e "${YELLOW}Suppression du réseau veza-network...${NC}"
        incus network delete veza-network || true
    fi
}

# Fonction principale de nettoyage
main() {
    echo -e "${RED}⚠️  ATTENTION: Cette opération va supprimer TOUS les containers et données Veza !${NC}"
    echo -e "${YELLOW}Voulez-vous continuer ? (oui/NON)${NC}"
    read -r response
    
    if [[ "$response" != "oui" ]]; then
        echo -e "${GREEN}Opération annulée${NC}"
        exit 0
    fi
    
    echo -e "${BLUE}🚀 Début du nettoyage complet...${NC}"
    
    # Nettoyage dans l'ordre correct
    clean_containers
    clean_profiles
    clean_storage_volumes
    clean_storage
    clean_network
    
    # Vérification finale
    echo -e "${GREEN}✅ Nettoyage terminé !${NC}"
    echo ""
    echo -e "${BLUE}📊 État final :${NC}"
    echo -e "${CYAN}Containers :${NC}"
    incus ls
    echo ""
    echo -e "${CYAN}Profils :${NC}"
    incus profile list
    echo ""
    echo -e "${CYAN}Réseaux :${NC}"
    incus network list
    echo ""
    echo -e "${CYAN}Stockage :${NC}"
    incus storage list
}

# Vérifier qu'Incus est disponible
if ! command -v incus &> /dev/null; then
    echo -e "${RED}❌ Incus n'est pas installé ou accessible${NC}"
    exit 1
fi

# Exécuter le nettoyage
main 