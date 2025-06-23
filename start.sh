#!/bin/bash

# Script de démarrage pour Veza - Application Web Unifiée avec Incus
# Ce script configure et démarre tous les services avec containers Incus

set -e

# Couleurs pour les messages
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'  
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

# Banner avec art ASCII
echo -e "${PURPLE}"
cat << "EOF"
╔══════════════════════════════════════════════════════════════════════════════╗
║                                                                              ║
║    ██╗   ██╗███████╗███████╗ █████╗     ██╗███╗   ██╗ ██████╗██╗   ██╗███████╗║
║    ██║   ██║██╔════╝╚══███╔╝██╔══██╗    ██║████╗  ██║██╔════╝██║   ██║██╔════╝║
║    ██║   ██║█████╗    ███╔╝ ███████║    ██║██╔██╗ ██║██║     ██║   ██║███████╗║
║    ╚██╗ ██╔╝██╔══╝   ███╔╝  ██╔══██║    ██║██║╚██╗██║██║     ██║   ██║╚════██║║
║     ╚████╔╝ ███████╗███████╗██║  ██║    ██║██║ ╚████║╚██████╗╚██████╔╝███████║║
║      ╚═══╝  ╚══════╝╚══════╝╚═╝  ╚═╝    ╚═╝╚═╝  ╚═══╝ ╚═════╝ ╚═════╝ ╚══════╝║
║                                                                              ║  
║                    🚀 Application Web Unifiée - Incus                       ║
║                          8 Containers Microservices                         ║
║                                                                              ║
╚══════════════════════════════════════════════════════════════════════════════╝
EOF
echo -e "${NC}"

# Fonction d'aide
show_help() {
    echo -e "${BLUE}Utilisation : $0 [OPTION]${NC}"
    echo ""
    echo -e "${CYAN}Options disponibles :${NC}"
    echo -e "  ${GREEN}setup${NC}      - Configuration initiale du projet et Incus"
    echo -e "  ${GREEN}dev${NC}        - Démarrage complet avec Incus (8 containers)"  
    echo -e "  ${GREEN}dev-local${NC}  - Démarrage en mode développement local (sans containers)"
    echo -e "  ${GREEN}build${NC}      - Construction de tous les services"
    echo -e "  ${GREEN}test${NC}       - Exécution de tous les tests"
    echo -e "  ${GREEN}clean${NC}      - Nettoyage des fichiers temporaires"
    echo -e "  ${GREEN}status${NC}     - Vérification du statut des containers"
    echo -e "  ${GREEN}stop${NC}       - Arrêt de tous les containers Incus"
    echo -e "  ${GREEN}logs${NC}       - Voir les logs de tous les containers"
    echo -e "  ${GREEN}help${NC}       - Afficher cette aide"
    echo ""
    echo -e "${YELLOW}Exemples :${NC}"
    echo -e "  $0 setup       # Configuration initiale"
    echo -e "  $0 dev         # Démarrage Incus (recommandé)"
    echo -e "  $0 dev-local   # Démarrage local sans containers"
    echo -e "  $0 status      # Vérifier le statut"
    echo ""
    echo -e "${CYAN}🐧 Architecture Incus - 8 Containers :${NC}"
    echo -e "  • ${YELLOW}veza-postgres${NC}  - Base de données PostgreSQL (10.100.0.15)"
    echo -e "  • ${YELLOW}veza-redis${NC}     - Cache Redis (10.100.0.17)"
    echo -e "  • ${YELLOW}veza-storage${NC}   - Stockage ZFS + NFS (10.100.0.18)"
    echo -e "  • ${YELLOW}veza-backend${NC}   - API Backend Go (10.100.0.12)"
    echo -e "  • ${YELLOW}veza-chat${NC}      - Serveur Chat Rust (10.100.0.13)"
    echo -e "  • ${YELLOW}veza-stream${NC}    - Serveur Stream Rust (10.100.0.14)"
    echo -e "  • ${YELLOW}veza-frontend${NC}  - Interface React (10.100.0.11)"
    echo -e "  • ${YELLOW}veza-haproxy${NC}   - Load Balancer HAProxy (10.100.0.16)"
    echo ""
    echo -e "${CYAN}💾 Gestion ZFS Storage :${NC}"
    echo -e "  • ${YELLOW}make zfs-status${NC}    - Statut du pool ZFS"
    echo -e "  • ${YELLOW}make zfs-snapshot${NC}  - Créer des snapshots"
    echo -e "  • ${YELLOW}make zfs-monitor${NC}   - Monitoring en temps réel"
    echo -e "  • ${YELLOW}make zfs-compress${NC}  - Stats compression"
    echo ""
}

# Fonction de vérification des prérequis
check_requirements() {
    echo -e "${BLUE}🔍 Vérification des prérequis...${NC}"
    
    local missing_requirements=()
    
    # Vérifier Incus (obligatoire)
    if ! command -v incus &> /dev/null; then
        missing_requirements+=("Incus")
    fi
    
    # Vérifier Make
    if ! command -v make &> /dev/null; then
        missing_requirements+=("Make")
    fi
    
    # Vérifier Node.js (pour développement local)
    if ! command -v node &> /dev/null; then
        echo -e "${YELLOW}⚠️ Node.js non trouvé (nécessaire pour développement local)${NC}"
    fi
    
    # Vérifier Go (pour développement local)
    if ! command -v go &> /dev/null; then
        echo -e "${YELLOW}⚠️ Go non trouvé (nécessaire pour développement local)${NC}"
    fi
    
    # Vérifier Rust (pour développement local)
    if ! command -v cargo &> /dev/null; then
        echo -e "${YELLOW}⚠️ Rust/Cargo non trouvé (nécessaire pour développement local)${NC}"
    fi
    
    if [ ${#missing_requirements[@]} -ne 0 ]; then
        echo -e "${RED}❌ Prérequis manquants :${NC}"
        for req in "${missing_requirements[@]}"; do
            echo -e "  • $req"
        done
        echo ""
        echo -e "${BLUE}📋 Instructions d'installation :${NC}"
        echo -e "${CYAN}Incus :${NC} sudo snap install incus --channel=latest/stable"
        echo -e "${CYAN}       ${NC} sudo incus admin init"
        echo -e "${CYAN}Make :${NC} sudo apt install make (Ubuntu/Debian)"
        echo -e "${CYAN}Node.js :${NC} https://nodejs.org/"
        echo -e "${CYAN}Go :${NC} https://golang.org/dl/"
        echo -e "${CYAN}Rust :${NC} https://rustup.rs/"
        echo ""
        exit 1
    fi
    
    echo -e "${GREEN}✅ Prérequis principaux installés${NC}"
}

# Fonction de configuration
setup_project() {
    echo -e "${BLUE}🔧 Configuration du projet Veza avec Incus...${NC}"
    
    # Vérifier les prérequis
    check_requirements
    
    # Utiliser le Makefile pour la configuration
    make setup
    make incus-setup
    
    echo -e "${GREEN}✅ Configuration terminée !${NC}"
    echo ""
    echo -e "${BLUE}📋 Prochaines étapes :${NC}"
    echo -e "  1. Modifier le fichier ${YELLOW}.env${NC} selon vos besoins"
    echo -e "  2. Lancer ${YELLOW}$0 dev${NC} pour démarrer avec Incus"
    echo -e "  3. Accéder à ${YELLOW}http://10.100.0.16${NC} pour l'application"
}

# Fonction de démarrage développement Incus
start_dev() {
    echo -e "${BLUE}🚀 Démarrage en mode développement avec Incus...${NC}"
    echo -e "${YELLOW}Les services seront accessibles sur :${NC}"
    echo -e "  • Application HAProxy : ${GREEN}http://10.100.0.16${NC}"
    echo -e "  • HAProxy Stats       : ${GREEN}http://10.100.0.16:8404/stats${NC}"
    echo -e "  • Frontend React      : ${GREEN}http://10.100.0.11:5173${NC}"
    echo -e "  • Backend API         : ${GREEN}http://10.100.0.12:8080${NC}"
    echo -e "  • Chat WebSocket      : ${GREEN}ws://10.100.0.13:8081/ws${NC}"
    echo -e "  • Stream WebSocket    : ${GREEN}ws://10.100.0.14:8082/ws${NC}"
    echo ""
    
    # Utiliser le Makefile pour démarrer avec Incus
    make incus-dev
}

# Fonction de démarrage développement local
start_dev_local() {
    echo -e "${BLUE}🚀 Démarrage en mode développement local...${NC}"
    echo -e "${YELLOW}Les services seront accessibles sur :${NC}"
    echo -e "  • Frontend React  : ${GREEN}http://localhost:5173${NC}"
    echo -e "  • Backend API     : ${GREEN}http://localhost:8080${NC}"
    echo -e "  • Chat WebSocket  : ${GREEN}ws://localhost:8081/ws${NC}"
    echo -e "  • Stream WebSocket: ${GREEN}ws://localhost:8082/ws${NC}"
    echo ""
    echo -e "${CYAN}💡 Appuyez sur Ctrl+C pour arrêter tous les services${NC}"
    echo ""
    
    # Utiliser le Makefile pour démarrer en mode développement local
    make dev-local
}

# Fonction de build
build_all() {
    echo -e "${BLUE}🔨 Construction de tous les services...${NC}"
    make build
}

# Fonction de test
test_all() {
    echo -e "${BLUE}🧪 Exécution de tous les tests...${NC}"
    make test
}

# Fonction de nettoyage
clean_all() {
    echo -e "${BLUE}🧹 Nettoyage des fichiers temporaires...${NC}"
    make clean
}

# Fonction de vérification de statut
status_check() {
    echo -e "${BLUE}📊 Vérification du statut des containers...${NC}"
    make status
}

# Fonction d'arrêt
stop_all() {
    echo -e "${BLUE}🛑 Arrêt de tous les containers Incus...${NC}"
    make incus-stop
}

# Fonction de logs
logs_all() {
    echo -e "${BLUE}📝 Logs de tous les containers...${NC}"
    make logs
}

# Script principal
main() {
    # Vérifier les arguments
    case "${1:-help}" in
        "setup")
            setup_project
            ;;
        "dev")
            check_requirements
            start_dev
            ;;
        "dev-local")
            check_requirements
            start_dev_local
            ;;
        "build")
            check_requirements
            build_all
            ;;
        "test")
            check_requirements
            test_all
            ;;
        "clean")
            clean_all
            ;;
        "status")
            check_requirements
            status_check
            ;;
        "stop")
            check_requirements
            stop_all
            ;;
        "logs")
            check_requirements
            logs_all
            ;;
        "help"|*)
            show_help
            ;;
    esac
}

# Exécuter le script principal
main "$@"