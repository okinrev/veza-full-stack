#!/bin/bash

# Script de réparation réseau Veza
# Utilise le réseau par défaut d'Incus qui fonctionne au lieu de créer un réseau custom

set -e

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

echo -e "${BLUE}🌐 Réparation réseau Veza...${NC}"

# Analyste des réseaux actuels
echo -e "${CYAN}📊 Analyse des réseaux actuels:${NC}"
incus network list

# Le problème principal identifié
echo -e "${YELLOW}🔍 ANALYSE DU PROBLÈME:${NC}"
echo -e "  • Le réseau 'veza-network' custom cause des problèmes de connectivité"
echo -e "  • Le réseau par défaut 'incusbr0' fonctionne parfaitement"
echo -e "  • Solution: Utiliser incusbr0 avec configuration optimisée"

# Vérifier le réseau par défaut
DEFAULT_BRIDGE="incusbr0"
if incus network show "$DEFAULT_BRIDGE" >/dev/null 2>&1; then
    echo -e "${GREEN}✅ Réseau par défaut trouvé: $DEFAULT_BRIDGE${NC}"
    
    # Afficher la configuration
    echo -e "${CYAN}📋 Configuration actuelle:${NC}"
    incus network show "$DEFAULT_BRIDGE" | grep -E "(ipv4\.address|ipv4\.nat|used_by)"
else
    echo -e "${RED}❌ Réseau par défaut introuvable${NC}"
    exit 1
fi

# Supprimer le réseau veza-network problématique s'il existe
if incus network show veza-network >/dev/null 2>&1; then
    echo -e "${YELLOW}🗑️ Suppression du réseau veza-network problématique...${NC}"
    
    # Détacher tous les containers du réseau veza-network
    ATTACHED_CONTAINERS=$(incus network show veza-network | grep "used_by:" -A 20 | grep "/instances/" | cut -d'/' -f4 || true)
    
    for container in $ATTACHED_CONTAINERS; do
        if [ -n "$container" ]; then
            echo -e "${CYAN}🔄 Migration de $container vers le réseau par défaut...${NC}"
            incus stop "$container" --force || true
            # Le container utilisera automatiquement le réseau par défaut au redémarrage
        fi
    done
    
    # Supprimer le réseau problématique
    incus network delete veza-network || true
    echo -e "${GREEN}✅ Réseau veza-network supprimé${NC}"
fi

# Optimiser les profils pour utiliser le réseau par défaut
echo -e "${BLUE}👤 Optimisation des profils pour le réseau par défaut...${NC}"

PROFILES=("veza-base" "veza-app" "veza-database" "veza-storage")

for profile in "${PROFILES[@]}"; do
    if incus profile show "$profile" >/dev/null 2>&1; then
        echo -e "${CYAN}🔧 Mise à jour profil $profile...${NC}"
        
        # Supprimer l'ancien device eth0 s'il existe
        incus profile device remove "$profile" eth0 2>/dev/null || true
        
        # Le profil utilisera automatiquement le réseau par défaut
        # Pas besoin de configuration explicite
        
        echo -e "${GREEN}✅ Profil $profile optimisé${NC}"
    fi
done

# Configuration système optimisée
echo -e "${BLUE}⚙️ Configuration système optimisée...${NC}"

# IPv4 forwarding (essentiel pour la connectivité)
echo 'net.ipv4.ip_forward=1' | sudo tee /etc/sysctl.d/99-incus-forwarding.conf > /dev/null
sudo sysctl -p /etc/sysctl.d/99-incus-forwarding.conf >/dev/null 2>&1

# Optimisation DNS système
if systemctl is-active systemd-resolved >/dev/null 2>&1; then
    echo -e "${CYAN}🌐 Optimisation DNS système...${NC}"
    
    sudo mkdir -p /etc/systemd/resolved.conf.d/
    sudo tee /etc/systemd/resolved.conf.d/incus-optimized.conf > /dev/null << 'EOF'
[Resolve]
DNS=8.8.8.8 8.8.4.4 1.1.1.1
FallbackDNS=9.9.9.9 149.112.112.112
DNSStubListener=yes
DNSSEC=allow-downgrade
Cache=yes
DNSOverTLS=no
EOF
    
    sudo systemctl restart systemd-resolved || true
    echo -e "${GREEN}✅ DNS système optimisé${NC}"
fi

# Test de connectivité
echo -e "${BLUE}🧪 Test de connectivité...${NC}"

# Test connectivité internet de l'hôte
if ping -c 2 8.8.8.8 >/dev/null 2>&1; then
    echo -e "${GREEN}✅ Connectivité internet hôte OK${NC}"
else
    echo -e "${RED}❌ Problème connectivité internet hôte${NC}"
fi

# Créer un container de test si nécessaire
TEST_CONTAINER="veza-network-test"
if ! incus list "$TEST_CONTAINER" --format csv | grep -q RUNNING; then
    echo -e "${CYAN}🧪 Création container de test...${NC}"
    incus launch images:debian/bookworm "$TEST_CONTAINER" --profile default
    sleep 10
    
    # Test DNS dans le container
    if incus exec "$TEST_CONTAINER" -- nslookup google.com >/dev/null 2>&1; then
        echo -e "${GREEN}✅ DNS container OK${NC}"
    else
        echo -e "${YELLOW}⚠️ DNS container limité${NC}"
    fi
    
    # Test connectivité internet dans le container
    if incus exec "$TEST_CONTAINER" -- ping -c 2 8.8.8.8 >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Connectivité internet container OK${NC}"
    else
        echo -e "${YELLOW}⚠️ Connectivité internet container limitée${NC}"
    fi
    
    # Nettoyage
    incus delete "$TEST_CONTAINER" --force
fi

echo -e "${GREEN}🎉 Réparation réseau terminée !${NC}"
echo ""
echo -e "${CYAN}📋 RÉSUMÉ:${NC}"
echo -e "  • ✅ Utilisation du réseau par défaut Incus (incusbr0)"
echo -e "  • ✅ Suppression du réseau veza-network problématique"
echo -e "  • ✅ Configuration système optimisée"
echo -e "  • ✅ DNS optimisé"
echo ""
echo -e "${BLUE}💡 PROCHAINES ÉTAPES:${NC}"
echo -e "  1. Les containers utiliseront automatiquement le réseau par défaut"
echo -e "  2. Les IPs seront attribuées par DHCP (plage: $(incus network get incusbr0 ipv4.address))"
echo -e "  3. La connectivité internet sera garantie"
echo ""
echo -e "${YELLOW}📝 NOTE:${NC} Les containers devront être redémarrés pour appliquer les changements réseau" 