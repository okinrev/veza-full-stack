#!/bin/bash

# Script de gestion ZFS pour Veza Storage
# Permet de gérer le stockage ZFS depuis l'hôte

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
echo "│       💾 Veza - Gestion ZFS Storage     │"
echo "╰──────────────────────────────────────────╯"
echo -e "${NC}"

# Fonction d'aide
show_help() {
    echo -e "${BLUE}Usage : $0 [OPTION]${NC}"
    echo ""
    echo -e "${CYAN}Options disponibles :${NC}"
    echo -e "  ${GREEN}status${NC}           - Afficher le statut du pool ZFS"
    echo -e "  ${GREEN}snapshot${NC}         - Créer des snapshots des volumes"
    echo -e "  ${GREEN}list-snapshots${NC}   - Lister tous les snapshots"
    echo -e "  ${GREEN}restore${NC}          - Restaurer depuis un snapshot"
    echo -e "  ${GREEN}compress-stats${NC}   - Statistiques de compression"
    echo -e "  ${GREEN}cleanup${NC}          - Nettoyer les anciens snapshots"
    echo -e "  ${GREEN}expand${NC}           - Étendre la taille des volumes"
    echo -e "  ${GREEN}monitor${NC}          - Monitoring en temps réel"
    echo -e "  ${GREEN}help${NC}             - Afficher cette aide"
    echo ""
    echo -e "${YELLOW}Exemples :${NC}"
    echo -e "  $0 status              # Voir le statut ZFS"
    echo -e "  $0 snapshot            # Créer des snapshots"
    echo -e "  $0 compress-stats      # Voir la compression"
    echo ""
}

# Vérifier qu'Incus est accessible
check_incus() {
    if ! command -v incus &> /dev/null; then
        echo -e "${RED}❌ Incus n'est pas installé ou accessible${NC}"
        exit 1
    fi
    
    if ! incus info veza-storage &>/dev/null; then
        echo -e "${RED}❌ Container veza-storage non trouvé${NC}"
        echo -e "${YELLOW}💡 Démarrez d'abord l'application avec : ./start.sh dev${NC}"
        exit 1
    fi
}

# Statut du pool ZFS
show_status() {
    echo -e "${BLUE}📊 Statut du pool ZFS Veza${NC}"
    echo -e "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    # Statut du pool depuis l'hôte
    echo -e "${CYAN}🏠 Depuis l'hôte :${NC}"
    if incus storage show veza-zfs-pool &>/dev/null; then
        incus storage info veza-zfs-pool
    else
        echo -e "${RED}❌ Pool veza-zfs-pool non trouvé${NC}"
        return 1
    fi
    
    echo ""
    echo -e "${CYAN}📦 Depuis le container :${NC}"
    incus exec veza-storage -- veza-zfs-manage.sh status
}

# Créer des snapshots
create_snapshots() {
    echo -e "${BLUE}📸 Création des snapshots ZFS...${NC}"
    
    local timestamp=$(date +%Y%m%d_%H%M%S)
    echo -e "${YELLOW}🕒 Timestamp : $timestamp${NC}"
    
    # Créer les snapshots depuis le container
    incus exec veza-storage -- veza-zfs-manage.sh snapshot
    
    echo -e "${GREEN}✅ Snapshots créés avec succès${NC}"
}

# Lister les snapshots
list_snapshots() {
    echo -e "${BLUE}📋 Liste des snapshots ZFS${NC}"
    echo -e "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    incus exec veza-storage -- veza-zfs-manage.sh list-snapshots
}

# Statistiques de compression
show_compression_stats() {
    echo -e "${BLUE}🗜️ Statistiques de compression ZFS${NC}"
    echo -e "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    incus exec veza-storage -- veza-zfs-manage.sh compress-stats
}

# Restaurer depuis un snapshot
restore_snapshot() {
    echo -e "${BLUE}🔄 Restauration depuis un snapshot${NC}"
    echo -e "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    # Lister les snapshots disponibles
    echo -e "${YELLOW}📋 Snapshots disponibles :${NC}"
    list_snapshots
    
    echo ""
    echo -e "${YELLOW}💡 Pour restaurer un snapshot, connectez-vous au container :${NC}"
    echo -e "  incus exec veza-storage -- bash"
    echo -e "  zfs rollback veza-zfs-pool/containers/veza-storage/uploads@SNAPSHOT_NAME"
    echo ""
    echo -e "${RED}⚠️ ATTENTION : La restauration supprimera toutes les données postérieures au snapshot !${NC}"
}

# Nettoyer les anciens snapshots
cleanup_snapshots() {
    echo -e "${BLUE}🧹 Nettoyage des anciens snapshots${NC}"
    echo -e "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    # Supprimer les snapshots de plus de 30 jours
    echo -e "${YELLOW}🗑️ Suppression des snapshots de plus de 30 jours...${NC}"
    
    # Script de nettoyage dans le container
    incus exec veza-storage -- bash -c '
    cutoff_date=$(date -d "30 days ago" +%Y%m%d)
    
    for snapshot in $(zfs list -H -t snapshot -o name | grep veza-storage); do
        snapshot_date=$(echo $snapshot | grep -o "[0-9]\{8\}_[0-9]\{6\}" | cut -d_ -f1)
        if [[ "$snapshot_date" < "$cutoff_date" ]]; then
            echo "🗑️ Suppression du snapshot: $snapshot"
            zfs destroy $snapshot
        fi
    done
    
    echo "✅ Nettoyage terminé"
    '
}

# Étendre les volumes
expand_volumes() {
    echo -e "${BLUE}📈 Expansion des volumes ZFS${NC}"
    echo -e "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    echo -e "${CYAN}📊 Tailles actuelles :${NC}"
    incus exec veza-storage -- df -h /storage/
    
    echo ""
    echo -e "${YELLOW}💡 Pour étendre un volume :${NC}"
    echo -e "  incus config device set veza-storage uploads size=20GB"
    echo -e "  incus config device set veza-storage audio size=50GB"
    echo -e "  incus config device set veza-storage backups size=30GB"
    echo ""
    echo -e "${BLUE}Exemple d'expansion automatique :${NC}"
    read -p "Étendre le volume audio à 50GB ? (y/N) " confirm
    
    if [[ $confirm == [yY] ]]; then
        echo -e "${YELLOW}🔧 Extension du volume audio...${NC}"
        incus config device set veza-storage audio size=50GB
        echo -e "${GREEN}✅ Volume audio étendu à 50GB${NC}"
    fi
}

# Monitoring en temps réel
monitor_zfs() {
    echo -e "${BLUE}📡 Monitoring ZFS en temps réel${NC}"
    echo -e "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo -e "${YELLOW}💡 Appuyez sur Ctrl+C pour arrêter${NC}"
    echo ""
    
    while true; do
        clear
        echo -e "${PURPLE}🔄 Monitoring ZFS Veza - $(date)${NC}"
        echo -e "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        
        # Utilisation des volumes
        echo -e "${CYAN}📊 Utilisation des volumes :${NC}"
        incus exec veza-storage -- df -h /storage/ | grep -E "(Filesystem|/storage)"
        
        echo ""
        echo -e "${CYAN}💾 Statut ZFS :${NC}"
        incus exec veza-storage -- zfs list | grep veza-storage | head -5
        
        echo ""
        echo -e "${CYAN}🗜️ Compression :${NC}"
        incus exec veza-storage -- zfs get compressratio veza-zfs-pool/containers/veza-storage/uploads | tail -1
        
        sleep 5
    done
}

# Script principal
main() {
    case "${1:-help}" in
        "status")
            check_incus
            show_status
            ;;
        "snapshot")
            check_incus
            create_snapshots
            ;;
        "list-snapshots")
            check_incus
            list_snapshots
            ;;
        "compress-stats")
            check_incus
            show_compression_stats
            ;;
        "restore")
            check_incus
            restore_snapshot
            ;;
        "cleanup")
            check_incus
            cleanup_snapshots
            ;;
        "expand")
            check_incus
            expand_volumes
            ;;
        "monitor")
            check_incus
            monitor_zfs
            ;;
        "help"|*)
            show_help
            ;;
    esac
}

# Exécuter le script principal
main "$@"