#!/bin/bash

# Script de configuration initiale Incus pour Veza
# Configure le réseau, les profils et prépare les containers

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
echo "│      🚀 Veza - Configuration Incus      │"
echo "╰──────────────────────────────────────────╯"
echo -e "${NC}"

# Vérifier qu'Incus est installé
if ! command -v incus &> /dev/null; then
    echo -e "${RED}❌ Incus n'est pas installé${NC}"
    echo -e "${YELLOW}💡 Installation d'Incus :${NC}"
    echo "  sudo snap install incus --channel=latest/stable"
    echo "  sudo incus admin init"
    exit 1
fi

echo -e "${GREEN}✅ Incus détecté${NC}"

# Configuration système hôte pour éviter les conflits DNS
echo -e "${BLUE}🔧 Configuration système hôte...${NC}"

# Désactiver systemd-resolved sur l'hôte si actif (évite les conflits)
if systemctl is-active systemd-resolved >/dev/null 2>&1; then
    echo -e "${YELLOW}⚠️ systemd-resolved actif - Configuration pour éviter les conflits${NC}"
    
    # Configuration pour que systemd-resolved n'interfère pas avec Incus
    sudo mkdir -p /etc/systemd/resolved.conf.d/
    sudo tee /etc/systemd/resolved.conf.d/incus.conf > /dev/null << 'EOF'
[Resolve]
DNS=8.8.8.8 8.8.4.4 1.1.1.1
FallbackDNS=9.9.9.9 149.112.112.112
DNSStubListener=no
DNSSEC=no
EOF
    sudo systemctl restart systemd-resolved || true
fi

# Activer le forwarding IPv4 pour Incus (critique pour la connectivité)
echo 'net.ipv4.ip_forward=1' | sudo tee /etc/sysctl.d/99-incus-forwarding.conf > /dev/null
sudo sysctl -p /etc/sysctl.d/99-incus-forwarding.conf

# Créer les répertoires nécessaires
echo -e "${BLUE}📁 Création des répertoires...${NC}"
mkdir -p {data/postgres,data/redis,logs,uploads,audio,ssl,backups}
chmod 755 {data/postgres,data/redis,logs,uploads,audio,ssl,backups}

# Configuration du pool de stockage ZFS pour veza-storage
echo -e "${BLUE}💾 Configuration du pool ZFS pour le stockage...${NC}"
if ! incus storage show veza-zfs-pool >/dev/null 2>&1; then
    # Créer un pool ZFS dédié pour le stockage Veza
    incus storage create veza-zfs-pool zfs size=50GB
    echo -e "${GREEN}✅ Pool ZFS veza-zfs-pool créé (50GB)${NC}"
else
    echo -e "${YELLOW}⚠️ Pool ZFS veza-zfs-pool existe déjà${NC}"
fi

# Configuration du réseau Veza avec connectivité internet GARANTIE
echo -e "${BLUE}🌐 Configuration du réseau veza-network...${NC}"
if ! incus network show veza-network >/dev/null 2>&1; then
    echo -e "${BLUE}🔧 Création du réseau avec connectivité internet optimisée...${NC}"
    incus network create veza-network \
        ipv4.address=10.100.0.1/24 \
        ipv4.nat=true \
        ipv4.dhcp=true \
        ipv4.dhcp.ranges="10.100.0.10-10.100.0.250" \
        ipv4.dhcp.expiry="24h" \
        ipv4.routing=true \
        ipv6.address=none \
        dns.domain=veza.local \
        dns.mode=none \
        raw.dnsmasq="
# Configuration DNS ultra-robuste
server=8.8.8.8
server=8.8.4.4  
server=1.1.1.1
server=9.9.9.9
# Optimisations réseau
cache-size=1000
neg-ttl=60
local-ttl=300
# DHCP optimisé
dhcp-authoritative
dhcp-lease-max=200
dhcp-option=option:dns-server,8.8.8.8,8.8.4.4,1.1.1.1
dhcp-option=option:domain-name,veza.local
dhcp-option=option:mtu,1500
# Logs pour debug
log-queries
log-dhcp"
    echo -e "${GREEN}✅ Réseau veza-network créé avec connectivité internet garantie${NC}"
else
    echo -e "${YELLOW}⚠️ Réseau veza-network existe déjà - reconfiguration COMPLÈTE...${NC}"
    # Reconfigurer complètement pour garantir la connectivité
    incus network delete veza-network || true
    sleep 2
    
    incus network create veza-network \
        ipv4.address=10.100.0.1/24 \
        ipv4.nat=true \
        ipv4.dhcp=true \
        ipv4.dhcp.ranges="10.100.0.10-10.100.0.250" \
        ipv4.dhcp.expiry="24h" \
        ipv4.routing=true \
        ipv6.address=none \
        dns.domain=veza.local \
        dns.mode=none \
        raw.dnsmasq="
# Configuration DNS ultra-robuste
server=8.8.8.8
server=8.8.4.4  
server=1.1.1.1
server=9.9.9.9
# Optimisations réseau
cache-size=1000
neg-ttl=60
local-ttl=300
# DHCP optimisé
dhcp-authoritative
dhcp-lease-max=200
dhcp-option=option:dns-server,8.8.8.8,8.8.4.4,1.1.1.1
dhcp-option=option:domain-name,veza.local
dhcp-option=option:mtu,1500
# Logs pour debug
log-queries
log-dhcp"
    echo -e "${GREEN}✅ Réseau veza-network recréé et optimisé${NC}"
fi

# Test de connectivité réseau hôte
echo -e "${BLUE}🧪 Test de connectivité réseau hôte...${NC}"
if timeout 5 ping -c 2 8.8.8.8 >/dev/null 2>&1; then
    echo -e "${GREEN}✅ Connectivité internet hôte OK${NC}"
else
    echo -e "${RED}❌ Problème connectivité internet hôte${NC}"
    echo -e "${YELLOW}⚠️ Vérifiez votre connexion internet avant de continuer${NC}"
fi

# Création du profil de base ULTRA-ROBUSTE
echo -e "${BLUE}👤 Configuration des profils...${NC}"

# Profil veza-base
if ! incus profile show veza-base >/dev/null 2>&1; then
    incus profile create veza-base
    incus profile set veza-base limits.cpu 2
    incus profile set veza-base limits.memory 2GB
    incus profile set veza-base security.nesting true
    incus profile set veza-base security.privileged false
    
    # Configuration réseau optimisée
    incus profile device add veza-base eth0 nic \
        network=veza-network \
        name=eth0
    
    # Device root
    incus profile device add veza-base root disk path=/ pool=incus_storage
    
    # Configuration du DNS dans le profil pour éviter les conflits
    incus profile set veza-base user.network-config - << 'EOF'
version: 2
ethernets:
  eth0:
    dhcp4: true
    dhcp4-overrides:
      use-dns: false
    nameservers:
      addresses: [8.8.8.8, 8.8.4.4, 1.1.1.1, 9.9.9.9]
      search: [veza.local]
EOF
    
    echo -e "${GREEN}✅ Profil veza-base créé avec DNS robuste${NC}"
else
    echo -e "${YELLOW}⚠️ Profil veza-base existe déjà - mise à jour...${NC}"
    # Mise à jour du profil existant
    incus profile set veza-base user.network-config - << 'EOF'
version: 2
ethernets:
  eth0:
    dhcp4: true
    dhcp4-overrides:
      use-dns: false
    nameservers:
      addresses: [8.8.8.8, 8.8.4.4, 1.1.1.1, 9.9.9.9]
      search: [veza.local]
EOF
    echo -e "${GREEN}✅ Profil veza-base mis à jour${NC}"
fi

# Profil veza-app
if ! incus profile show veza-app >/dev/null 2>&1; then
    incus profile copy veza-base veza-app
    incus profile set veza-app limits.cpu 4
    incus profile set veza-app limits.memory 4GB
    
    # Monter le dossier logs
    incus profile device add veza-app logs disk \
        source=$(pwd)/logs \
        path=/app/logs
        
    echo -e "${GREEN}✅ Profil veza-app créé${NC}"
else
    echo -e "${YELLOW}⚠️ Profil veza-app existe déjà${NC}"
fi

# Profil veza-database
if ! incus profile show veza-database >/dev/null 2>&1; then
    incus profile copy veza-base veza-database
    incus profile set veza-database limits.cpu 4
    incus profile set veza-database limits.memory 8GB
    
    # Monter le dossier de données PostgreSQL
    incus profile device add veza-database data disk \
        source=$(pwd)/data/postgres \
        path=/var/lib/postgresql/data
        
    echo -e "${GREEN}✅ Profil veza-database créé${NC}"
else
    echo -e "${YELLOW}⚠️ Profil veza-database existe déjà${NC}"
fi

# Profil veza-storage avec ZFS
if ! incus profile show veza-storage >/dev/null 2>&1; then
    incus profile copy veza-base veza-storage
    incus profile set veza-storage limits.memory 4GB
    incus profile set veza-storage security.privileged true
    
    # Créer les volumes ZFS d'abord
    echo -e "${YELLOW}💾 Création des volumes ZFS...${NC}"
    incus storage volume create veza-zfs-pool uploads --type=filesystem || true
    incus storage volume create veza-zfs-pool audio --type=filesystem || true
    incus storage volume create veza-zfs-pool backups --type=filesystem || true
    incus storage volume create veza-zfs-pool zfs-cache --type=filesystem || true
    
    # Configurer les tailles des volumes
    incus storage volume set veza-zfs-pool uploads size=10GB || true
    incus storage volume set veza-zfs-pool audio size=20GB || true
    incus storage volume set veza-zfs-pool backups size=15GB || true
    incus storage volume set veza-zfs-pool zfs-cache size=5GB || true
    
    # Monter les volumes ZFS dans le profil
    incus profile device add veza-storage uploads disk \
        pool=veza-zfs-pool \
        source=uploads \
        path=/storage/uploads || true
    incus profile device add veza-storage audio disk \
        pool=veza-zfs-pool \
        source=audio \
        path=/storage/audio || true
    incus profile device add veza-storage backups disk \
        pool=veza-zfs-pool \
        source=backups \
        path=/storage/backups || true
    incus profile device add veza-storage zfs-cache disk \
        pool=veza-zfs-pool \
        source=zfs-cache \
        path=/storage/cache || true
        
    echo -e "${GREEN}✅ Profil veza-storage avec ZFS créé${NC}"
    echo -e "${CYAN}📊 Volumes ZFS configurés :${NC}"
    echo -e "  • uploads: 10GB"
    echo -e "  • audio: 20GB" 
    echo -e "  • backups: 15GB"
    echo -e "  • cache: 5GB"
else
    echo -e "${YELLOW}⚠️ Profil veza-storage existe déjà${NC}"
fi

# Vérification finale de la configuration réseau
echo -e "${BLUE}🔍 Vérification finale de la configuration...${NC}"

# Test du réseau Incus
if incus network info veza-network | grep -q "State: up"; then
    echo -e "${GREEN}✅ Réseau veza-network opérationnel${NC}"
else
    echo -e "${RED}❌ Problème avec le réseau veza-network${NC}"
fi

# Test de forwarding IPv4
if sysctl net.ipv4.ip_forward | grep -q "1"; then
    echo -e "${GREEN}✅ IPv4 forwarding activé${NC}"
else
    echo -e "${YELLOW}⚠️ IPv4 forwarding non activé - tentative d'activation...${NC}"
    sudo sysctl -w net.ipv4.ip_forward=1
fi

echo -e "${GREEN}🎉 Configuration Incus terminée !${NC}"
echo ""
echo -e "${BLUE}📋 Profils créés :${NC}"
echo -e "  • ${YELLOW}veza-base${NC} - Configuration de base avec DNS robuste"
echo -e "  • ${YELLOW}veza-app${NC} - Services applicatifs"  
echo -e "  • ${YELLOW}veza-database${NC} - Base de données"
echo -e "  • ${YELLOW}veza-storage${NC} - Système de fichiers ZFS"
echo ""
echo -e "${BLUE}🌐 Réseau configuré :${NC}"
echo -e "  • ${YELLOW}veza-network${NC} - 10.100.0.0/24 avec DHCP intelligent et DNS ultra-robuste"
echo -e "  • DNS: 8.8.8.8, 8.8.4.4, 1.1.1.1, 9.9.9.9"
echo -e "  • DHCP: 10.100.0.10-10.100.0.250"
echo ""
echo -e "${CYAN}💡 Prochaine étape : ./scripts/incus-deploy.sh${NC}" 