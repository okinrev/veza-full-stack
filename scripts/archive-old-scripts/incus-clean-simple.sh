#!/bin/bash

# Script de nettoyage simple pour les containers Veza
set -e

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}"
echo "╭────────────────────────────────────────╮"
echo "│      🧹 Nettoyage Containers Veza     │"
echo "╰────────────────────────────────────────╯"
echo -e "${NC}"

# Liste des containers Veza (8 au total)
CONTAINERS=("veza-haproxy" "veza-frontend" "veza-storage" "veza-stream" "veza-chat" "veza-backend" "veza-redis" "veza-postgres")

echo -e "${YELLOW}⚠️ Ce script va supprimer tous les containers Veza. Continuer ? (o/N)${NC}"
read -r response

if [[ "$response" != "o" && "$response" != "oui" ]]; then
    echo -e "${GREEN}Nettoyage annulé${NC}"
    exit 0
fi

echo -e "${BLUE}🔍 Recherche des containers Veza...${NC}"

# Obtenir la liste des containers existants
existing_containers=$(incus ls --format csv | cut -d, -f1)

for container in "${CONTAINERS[@]}"; do
    if echo "$existing_containers" | grep -q "^$container$"; then
        echo -e "${YELLOW}🗑️ Suppression de $container...${NC}"
        
        # Arrêter et supprimer le container en une fois
        echo -e "  ⏹️ Arrêt et suppression de $container..."
        incus stop "$container" --force 2>/dev/null || true
        incus delete "$container" 2>/dev/null || true
        
        echo -e "${GREEN}  ✅ $container supprimé${NC}"
    else
        echo -e "${BLUE}  ℹ️ $container n'existe pas${NC}"
    fi
done

echo ""
echo -e "${BLUE}📊 État final des containers :${NC}"
incus ls

echo ""
echo -e "${GREEN}🎉 Nettoyage terminé !${NC}"
echo -e "${CYAN}💡 Vous pouvez maintenant relancer le déploiement avec :${NC}"
echo -e "   ./scripts/incus-deploy-simple.sh" 