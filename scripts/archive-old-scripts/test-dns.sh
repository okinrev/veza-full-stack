#!/bin/bash

# Script de test DNS pour diagnostiquer les problèmes Incus
set -euo pipefail

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

echo -e "${BLUE}🔍 Test de résolution DNS depuis l'hôte et Incus...${NC}"

# Test 1: DNS de l'hôte
echo -e "${BLUE}📡 Test DNS hôte...${NC}"
echo -e "  • Test nslookup deb.debian.org:"
nslookup deb.debian.org 8.8.8.8 || echo -e "${RED}❌ Échec${NC}"

echo -e "  • Test curl vers deb.debian.org:"
timeout 10 curl -I http://deb.debian.org/ || echo -e "${RED}❌ Échec${NC}"

echo -e "  • Test ping 8.8.8.8:"
ping -c 2 8.8.8.8 || echo -e "${RED}❌ Échec${NC}"

# Test 2: Incus réseau
echo -e "${BLUE}🌐 Test réseau Incus...${NC}"
if incus network show veza-network >/dev/null 2>&1; then
    echo -e "${GREEN}✅ Réseau veza-network existe${NC}"
    incus network show veza-network | grep -E "(ipv4\.|dns\.)"
else
    echo -e "${RED}❌ Réseau veza-network manquant${NC}"
fi

# Test 3: Container temporaire
echo -e "${BLUE}🧪 Test container temporaire...${NC}"

# Créer un container temporaire
if ! incus list test-dns --format csv | grep -q "test-dns"; then
    echo -e "  • Création container test..."
    incus launch images:debian/bookworm test-dns --profile veza-base 2>/dev/null || {
        echo -e "${RED}❌ Échec création container${NC}"
        exit 1
    }
    
    # Attendre le démarrage
    sleep 5
fi

echo -e "  • Test DNS dans container:"
incus exec test-dns -- bash -c "
    echo 'Configuration DNS container:'
    cat /etc/resolv.conf
    echo ''
    
    echo 'Test nslookup depuis container:'
    timeout 5 nslookup deb.debian.org 8.8.8.8 || echo 'DNS échoué'
    
    echo 'Test ping depuis container:'
    timeout 5 ping -c 2 8.8.8.8 || echo 'Ping échoué'
    
    echo 'Test curl depuis container:'
    timeout 10 curl -I http://deb.debian.org/ || echo 'HTTP échoué'
"

# Nettoyer
echo -e "${BLUE}🧹 Nettoyage container test...${NC}"
incus delete test-dns --force

echo -e "${GREEN}✅ Tests DNS terminés${NC}" 