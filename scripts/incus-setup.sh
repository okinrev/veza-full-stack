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

# Configuration du réseau Veza avec DHCP amélioré
echo -e "${BLUE}🌐 Configuration du réseau veza-network...${NC}"
if ! incus network show veza-network >/dev/null 2>&1; then
    incus network create veza-network \
        ipv4.address=10.100.0.1/24 \
        ipv4.nat=true \
        ipv4.dhcp=true \
        ipv4.dhcp.ranges="10.100.0.10-10.100.0.250" \
        ipv6.address=none \
        dns.domain=veza.local \
        dns.mode=managed
    echo -e "${GREEN}✅ Réseau veza-network créé avec DHCP${NC}"
else
    echo -e "${YELLOW}⚠️ Réseau veza-network existe déjà${NC}"
    # Mettre à jour la configuration DHCP
    incus network set veza-network ipv4.dhcp true
    incus network set veza-network ipv4.dhcp.ranges "10.100.0.10-10.100.0.250"
    incus network set veza-network dns.mode managed
fi

# Création du profil de base
echo -e "${BLUE}👤 Configuration des profils...${NC}"

# Profil veza-base
if ! incus profile show veza-base >/dev/null 2>&1; then
    incus profile create veza-base
    incus profile set veza-base limits.cpu 2
    incus profile set veza-base limits.memory 2GB
    incus profile set veza-base security.nesting true
    incus profile set veza-base security.privileged false
    
    # Device réseau avec configuration DNS
    incus profile device add veza-base eth0 nic network=veza-network
    
    # Device root
    incus profile device add veza-base root disk path=/ pool=incus_storage
    
    echo -e "${GREEN}✅ Profil veza-base créé${NC}"
else
    echo -e "${YELLOW}⚠️ Profil veza-base existe déjà${NC}"
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
    incus storage volume create veza-zfs-pool uploads --type=filesystem
    incus storage volume create veza-zfs-pool audio --type=filesystem  
    incus storage volume create veza-zfs-pool backups --type=filesystem
    incus storage volume create veza-zfs-pool zfs-cache --type=filesystem
    
    # Configurer les tailles des volumes
    incus storage volume set veza-zfs-pool uploads size=10GB
    incus storage volume set veza-zfs-pool audio size=20GB
    incus storage volume set veza-zfs-pool backups size=15GB
    incus storage volume set veza-zfs-pool zfs-cache size=5GB
    
    # Monter les volumes ZFS dans le profil
    incus profile device add veza-storage uploads disk \
        pool=veza-zfs-pool \
        source=uploads \
        path=/storage/uploads
    incus profile device add veza-storage audio disk \
        pool=veza-zfs-pool \
        source=audio \
        path=/storage/audio
    incus profile device add veza-storage backups disk \
        pool=veza-zfs-pool \
        source=backups \
        path=/storage/backups
    incus profile device add veza-storage zfs-cache disk \
        pool=veza-zfs-pool \
        source=zfs-cache \
        path=/storage/cache
        
    echo -e "${GREEN}✅ Profil veza-storage avec ZFS créé${NC}"
    echo -e "${CYAN}📊 Volumes ZFS configurés :${NC}"
    echo -e "  • uploads: 10GB"
    echo -e "  • audio: 20GB" 
    echo -e "  • backups: 15GB"
    echo -e "  • cache: 5GB"
else
    echo -e "${YELLOW}⚠️ Profil veza-storage existe déjà${NC}"
fi

echo -e "${GREEN}🎉 Configuration Incus terminée !${NC}"
echo ""
echo -e "${BLUE}📋 Profils créés :${NC}"
echo -e "  • ${YELLOW}veza-base${NC} - Configuration de base"
echo -e "  • ${YELLOW}veza-app${NC} - Services applicatifs"  
echo -e "  • ${YELLOW}veza-database${NC} - Base de données"
echo -e "  • ${YELLOW}veza-storage${NC} - Système de fichiers"
echo ""
echo -e "${BLUE}🌐 Réseau configuré :${NC}"
echo -e "  • ${YELLOW}veza-network${NC} - 10.100.0.0/24 avec DHCP"
echo ""
echo -e "${CYAN}💡 Prochaine étape : ./scripts/incus-deploy.sh${NC}" 