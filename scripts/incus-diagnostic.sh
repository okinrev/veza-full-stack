#!/bin/bash

# Script de diagnostic et réparation Incus pour Veza
# Résout les problèmes de réseau et de configuration

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
echo "│    🔧 Veza - Diagnostic et Réparation   │"
echo "╰──────────────────────────────────────────╯"
echo -e "${NC}"

# Fonction pour afficher l'état actuel
show_status() {
    echo -e "${BLUE}📊 État actuel du système Incus${NC}"
    echo "----------------------------------------"
    
    echo -e "${CYAN}Réseaux :${NC}"
    incus network list
    echo ""
    
    echo -e "${CYAN}Profils :${NC}"
    incus profile list
    echo ""
    
    echo -e "${CYAN}Containers :${NC}"
    incus ls
    echo ""
}

# Fonction pour nettoyer et reconfigurer le réseau
fix_network() {
    echo -e "${YELLOW}🔧 Réparation de la configuration réseau...${NC}"
    
    # Arrêter le container problématique
    if incus ls --format csv | grep -q "veza-postgres"; then
        echo -e "${BLUE}Arrêt du container veza-postgres...${NC}"
        incus stop veza-postgres --force || true
    fi
    
    # Vérifier et reconfigurer le réseau veza-network
    if incus network show veza-network >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Réseau veza-network existe${NC}"
        
        # Vérifier la configuration DHCP
        echo -e "${BLUE}Configuration du serveur DHCP...${NC}"
        incus network set veza-network ipv4.dhcp true
        incus network set veza-network ipv4.dhcp.ranges "10.100.0.10-10.100.0.250"
        
        # Redémarrer le réseau
        echo -e "${BLUE}Redémarrage du réseau...${NC}"
        incus network set veza-network ipv4.nat true
        
    else
        echo -e "${RED}❌ Réseau veza-network manquant, création...${NC}"
        incus network create veza-network \
            ipv4.address=10.100.0.1/24 \
            ipv4.nat=true \
            ipv4.dhcp=true \
            ipv4.dhcp.ranges="10.100.0.10-10.100.0.250" \
            ipv6.address=none \
            dns.domain=veza.local
    fi
    
    # Reconfigurer le container
    echo -e "${BLUE}Reconfiguration du container veza-postgres...${NC}"
    
    # Supprimer l'ancienne configuration réseau
    incus config device remove veza-postgres eth0 || true
    
    # Ajouter la nouvelle configuration réseau
    incus config device add veza-postgres eth0 nic \
        network=veza-network \
        ipv4.address=10.100.0.15
    
    # Démarrer le container
    echo -e "${BLUE}Démarrage du container...${NC}"
    incus start veza-postgres
    
    # Attendre que le container soit complètement démarré
    echo -e "${BLUE}Attente du démarrage complet...${NC}"
    sleep 5
    
    # Forcer l'obtention de l'IP
    echo -e "${BLUE}Configuration IP dans le container...${NC}"
    incus exec veza-postgres -- ip addr flush dev eth0 || true
    incus exec veza-postgres -- dhclient -r eth0 || true
    incus exec veza-postgres -- dhclient eth0 || true
    
    # Alternative : configuration manuelle
    if ! incus exec veza-postgres -- ip addr show eth0 | grep -q "10.100.0.15"; then
        echo -e "${YELLOW}Configuration IP manuelle...${NC}"
        incus exec veza-postgres -- ip addr add 10.100.0.15/24 dev eth0 || true
        incus exec veza-postgres -- ip route add default via 10.100.0.1 || true
    fi
}

# Fonction pour tester la connectivité
test_connectivity() {
    echo -e "${BLUE}🔍 Tests de connectivité${NC}"
    echo "----------------------------------------"
    
    # Test ping vers le container
    echo -e "${CYAN}Test ping vers veza-postgres (10.100.0.15)...${NC}"
    if ping -c 2 -W 3 10.100.0.15 >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Ping réussi${NC}"
    else
        echo -e "${RED}❌ Ping échoué${NC}"
    fi
    
    # Test depuis le container vers l'hôte
    echo -e "${CYAN}Test depuis le container vers l'hôte...${NC}"
    if incus exec veza-postgres -- ping -c 2 -W 3 10.100.0.1 >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Connectivité container -> hôte OK${NC}"
    else
        echo -e "${RED}❌ Connectivité container -> hôte échouée${NC}"
    fi
    
    # Test de résolution DNS
    echo -e "${CYAN}Test de résolution DNS...${NC}"
    if incus exec veza-postgres -- nslookup google.com >/dev/null 2>&1; then
        echo -e "${GREEN}✅ DNS OK${NC}"
    else
        echo -e "${RED}❌ DNS échoué${NC}"
    fi
}

# Fonction principale
main() {
    echo -e "${BLUE}🚀 Début du diagnostic...${NC}"
    
    # Afficher l'état actuel
    show_status
    
    # Demander confirmation pour la réparation
    echo -e "${YELLOW}🔧 Voulez-vous procéder à la réparation automatique ? (o/N)${NC}"
    read -r response
    
    if [[ "$response" =~ ^[oO][uU]?[iI]?$ ]]; then
        # Effectuer la réparation
        fix_network
        
        # Afficher le nouvel état
        echo -e "${GREEN}📊 Nouvel état après réparation${NC}"
        show_status
        
        # Tester la connectivité
        test_connectivity
        
        echo -e "${GREEN}🎉 Diagnostic et réparation terminés !${NC}"
        echo ""
        echo -e "${BLUE}📋 Résumé :${NC}"
        echo -e "  • Container : ${YELLOW}veza-postgres${NC}"
        echo -e "  • IP : ${YELLOW}10.100.0.15${NC}"
        echo -e "  • Réseau : ${YELLOW}veza-network (10.100.0.0/24)${NC}"
        echo ""
        echo -e "${CYAN}💡 Commandes utiles :${NC}"
        echo -e "  • État : ${YELLOW}incus ls${NC}"
        echo -e "  • Logs : ${YELLOW}incus info veza-postgres${NC}"
        echo -e "  • Shell : ${YELLOW}incus exec veza-postgres -- bash${NC}"
        
    else
        echo -e "${YELLOW}Diagnostic seulement - aucune modification effectuée${NC}"
    fi
}

# Vérifier qu'Incus est disponible
if ! command -v incus &> /dev/null; then
    echo -e "${RED}❌ Incus n'est pas installé ou accessible${NC}"
    exit 1
fi

# Exécuter le diagnostic
main 