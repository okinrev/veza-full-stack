#!/bin/bash

# ╭──────────────────────────────────────────────────────────────╮
# │              🚀 Veza - Script de Déploiement Principal      │
# │             Script orchestrateur pour déploiement complet   │
# ╰──────────────────────────────────────────────────────────────╯

set -e

# Couleurs et formatage
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

# Variables globales
WORKSPACE_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
SCRIPTS_DIR="$WORKSPACE_DIR/scripts"
LOG_FILE="$WORKSPACE_DIR/deploy.log"

# Fonctions utilitaires
log() { 
    echo -e "${BLUE}[$(date '+%H:%M:%S')]${NC} $1" | tee -a "$LOG_FILE"
}
success() { 
    echo -e "${GREEN}✅ $1${NC}" | tee -a "$LOG_FILE"
}
warning() { 
    echo -e "${YELLOW}⚠️  $1${NC}" | tee -a "$LOG_FILE"
}
error() { 
    echo -e "${RED}❌ $1${NC}" | tee -a "$LOG_FILE"
    exit 1
}
header() { 
    echo -e "${PURPLE}${BOLD}"
    echo "╭──────────────────────────────────────────────────────────────╮"
    echo "│ $1"
    echo "╰──────────────────────────────────────────────────────────────╯"
    echo -e "${NC}"
}

# Fonction d'aide
show_help() {
    echo -e "${CYAN}${BOLD}Veza - Script de Déploiement Principal${NC}"
    echo ""
    echo -e "${BLUE}Usage:${NC} $0 [COMMAND] [OPTIONS]"
    echo ""
    echo -e "${YELLOW}Commandes:${NC}"
    echo -e "  ${GREEN}setup${NC}        - Configuration initiale d'Incus uniquement"
    echo -e "  ${GREEN}infrastructure${NC} - Déployer l'infrastructure (containers)"
    echo -e "  ${GREEN}applications${NC} - Déployer les applications"
    echo -e "  ${GREEN}deploy${NC}       - Déploiement complet (setup + infrastructure + applications)"
    echo -e "  ${GREEN}status${NC}       - Vérifier le statut"
    echo -e "  ${GREEN}test${NC}         - Lancer les tests"
    echo -e "  ${GREEN}clean${NC}        - Nettoyer l'environnement"
    echo -e "  ${GREEN}help${NC}         - Afficher cette aide"
    echo ""
    echo -e "${YELLOW}Options:${NC}"
    echo -e "  ${GREEN}--dev${NC}        - Mode développement"
    echo -e "  ${GREEN}--prod${NC}       - Mode production (défaut)"
    echo -e "  ${GREEN}--force${NC}      - Forcer la reconstruction"
    echo ""
    echo -e "${YELLOW}Exemples:${NC}"
    echo -e "  $0 deploy                    # Déploiement complet"
    echo -e "  $0 applications --dev        # Déployer seulement les apps en mode dev"
    echo -e "  $0 clean --force             # Nettoyage complet"
    echo ""
}

# Vérification des prérequis
check_prerequisites() {
    header "🔧 Vérification des Prérequis"
    
    # Vérifier Incus
    if ! command -v incus &> /dev/null; then
        error "Incus n'est pas installé. Installez-le avec: sudo snap install incus --channel=latest/stable"
    fi
    success "Incus disponible"
    
    # Vérifier les scripts nécessaires
    local required_scripts=(
        "incus-setup.sh"
        "incus-deploy.sh" 
        "deploy-unified.sh"
        "deploy-all.sh"
        "incus-status.sh"
        "test-complete.sh"
    )
    
    for script in "${required_scripts[@]}"; do
        if [[ ! -f "$SCRIPTS_DIR/$script" ]]; then
            error "Script manquant: $script"
        fi
        chmod +x "$SCRIPTS_DIR/$script"
    done
    success "Scripts disponibles"
    
    # Vérifier la structure du projet
    local required_dirs=(
        "veza-backend-api"
        "veza-frontend"
        "veza-chat-server"
        "veza-stream-server"
        "configs"
    )
    
    for dir in "${required_dirs[@]}"; do
        if [[ ! -d "$WORKSPACE_DIR/$dir" ]]; then
            error "Répertoire manquant: $dir"
        fi
    done
    success "Structure du projet validée"
}

# Configuration initiale
setup_infrastructure() {
    header "🏗️ Configuration Initiale"
    
    log "Exécution de incus-setup.sh..."
    "$SCRIPTS_DIR/incus-setup.sh"
    
    success "Configuration initiale terminée"
}

# Déploiement de l'infrastructure
deploy_infrastructure() {
    header "📦 Déploiement Infrastructure"
    
    log "Exécution de incus-deploy.sh..."
    "$SCRIPTS_DIR/incus-deploy.sh"
    
    success "Infrastructure déployée"
}

# Déploiement des applications
deploy_applications() {
    header "🚀 Déploiement Applications"
    
    log "Exécution de deploy-unified.sh..."
    "$SCRIPTS_DIR/deploy-unified.sh" "$@"
    
    log "Configuration finale avec deploy-all.sh..."
    "$SCRIPTS_DIR/deploy-all.sh"
    
    success "Applications déployées"
}

# Vérification du statut
check_status() {
    header "📊 Vérification du Statut"
    
    if [[ -f "$SCRIPTS_DIR/incus-status.sh" ]]; then
        "$SCRIPTS_DIR/incus-status.sh"
    else
        # Vérification basique
        log "Vérification des containers..."
        incus list
    fi
}

# Tests complets
run_tests() {
    header "🧪 Tests Complets"
    
    if [[ -f "$SCRIPTS_DIR/test-complete.sh" ]]; then
        "$SCRIPTS_DIR/test-complete.sh"
    else
        log "Lancement des tests basiques..."
        "$SCRIPTS_DIR/test.sh"
    fi
}

# Nettoyage
clean_environment() {
    header "🧹 Nettoyage"
    
    warning "Cette opération va supprimer tous les containers Veza"
    
    if [[ "${FORCE:-false}" == "true" ]]; then
        log "Nettoyage forcé..."
    else
        read -p "Continuer ? (y/N) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log "Nettoyage annulé"
            return 0
        fi
    fi
    
    if [[ -f "$SCRIPTS_DIR/incus-clean.sh" ]]; then
        "$SCRIPTS_DIR/incus-clean.sh"
    else
        # Nettoyage basique
        local containers=("veza-haproxy" "veza-frontend" "veza-stream" "veza-chat" "veza-backend" "veza-storage" "veza-redis" "veza-postgres")
        for container in "${containers[@]}"; do
            log "Suppression de $container..."
            incus stop "$container" --force 2>/dev/null || true
            incus delete "$container" 2>/dev/null || true
        done
    fi
    
    success "Nettoyage terminé"
}

# Gestion des erreurs
trap 'error "Erreur dans le déploiement. Consultez le log: $LOG_FILE"' ERR

# Parser les arguments
COMMAND=""
MODE="prod"
FORCE="false"

while [[ $# -gt 0 ]]; do
    case $1 in
        setup|infrastructure|applications|deploy|status|test|clean|help)
            COMMAND=$1
            shift
            ;;
        --dev)
            MODE="dev"
            shift
            ;;
        --prod)
            MODE="prod"
            shift
            ;;
        --force)
            FORCE="true"
            shift
            ;;
        *)
            error "Option inconnue: $1. Utilisez 'help' pour voir les options."
            ;;
    esac
done

# Commande par défaut
if [[ -z "$COMMAND" ]]; then
    COMMAND="help"
fi

# Initialiser le log
echo "=== Veza Deployment Log - $(date) ===" > "$LOG_FILE"

# Fonction principale
main() {
    header "🚀 Veza - Déploiement Principal ($MODE)"
    
    case $COMMAND in
        setup)
            check_prerequisites
            setup_infrastructure
            ;;
        infrastructure)
            check_prerequisites
            deploy_infrastructure
            ;;
        applications)
            check_prerequisites
            deploy_applications
            ;;
        deploy)
            check_prerequisites
            setup_infrastructure
            deploy_infrastructure
            deploy_applications
            check_status
            ;;
        status)
            check_status
            ;;
        test)
            run_tests
            ;;
        clean)
            clean_environment
            ;;
        help)
            show_help
            ;;
        *)
            error "Commande inconnue: $COMMAND"
            ;;
    esac
    
    success "Opération '$COMMAND' terminée avec succès!"
    
    if [[ "$COMMAND" == "deploy" || "$COMMAND" == "applications" ]]; then
        echo ""
        echo -e "${CYAN}🌐 Application Veza disponible :${NC}"
        echo -e "  • Principal : ${GREEN}http://10.5.191.133${NC}"
        echo -e "  • Frontend direct : ${GREEN}http://10.5.191.41:5173${NC}"
        echo -e "  • Backend API : ${GREEN}http://10.5.191.241:8080${NC}"
        echo ""
        echo -e "${YELLOW}💡 Commandes utiles :${NC}"
        echo -e "  • Tests : ${CYAN}$0 test${NC}"
        echo -e "  • Statut : ${CYAN}$0 status${NC}"
        echo -e "  • Logs : ${CYAN}tail -f $LOG_FILE${NC}"
    fi
}

# Exécution
main 