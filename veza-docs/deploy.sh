#!/bin/bash

# Script de déploiement pour la documentation Veza
# Usage: ./deploy.sh [production|staging|local]

set -e

# Couleurs pour les messages
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_NAME="veza-docs"
GITHUB_REPO="okinrev/veza-full-stack"
GITHUB_BRANCH="main"
DEPLOY_BRANCH="gh-pages"

# Fonction pour afficher les messages
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Fonction pour vérifier les prérequis
check_prerequisites() {
    log_info "Vérification des prérequis..."
    
    # Vérifier Node.js
    if ! command -v node &> /dev/null; then
        log_error "Node.js n'est pas installé"
        exit 1
    fi
    
    # Vérifier npm/yarn
    if ! command -v npm &> /dev/null; then
        log_error "npm n'est pas installé"
        exit 1
    fi
    
    # Vérifier git
    if ! command -v git &> /dev/null; then
        log_error "git n'est pas installé"
        exit 1
    fi
    
    log_success "Tous les prérequis sont satisfaits"
}

# Fonction pour installer les dépendances
install_dependencies() {
    log_info "Installation des dépendances..."
    
    if [ -f "package-lock.json" ]; then
        npm ci
    else
        npm install
    fi
    
    log_success "Dépendances installées"
}

# Fonction pour construire le site
build_site() {
    log_info "Construction du site..."
    
    # Nettoyer le build précédent
    if [ -d "build" ]; then
        rm -rf build
    fi
    
    # Construire le site
    npm run build
    
    if [ ! -d "build" ]; then
        log_error "La construction a échoué"
        exit 1
    fi
    
    log_success "Site construit avec succès"
}

# Fonction pour tester le site localement
test_local() {
    log_info "Démarrage du serveur de développement..."
    log_info "Le site sera disponible sur http://localhost:3000"
    log_info "Appuyez sur Ctrl+C pour arrêter"
    
    npm start
}

# Fonction pour déployer sur GitHub Pages
deploy_github_pages() {
    log_info "Déploiement sur GitHub Pages..."
    
    # Vérifier que nous sommes dans un repo git
    if [ ! -d ".git" ]; then
        log_error "Ce répertoire n'est pas un repository git"
        exit 1
    fi
    
    # Vérifier le statut git
    if [ -n "$(git status --porcelain)" ]; then
        log_warning "Il y a des modifications non commitées"
        read -p "Voulez-vous continuer ? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_info "Déploiement annulé"
            exit 0
        fi
    fi
    
    # Créer le dossier temporaire pour le déploiement
    TEMP_DIR=$(mktemp -d)
    log_info "Dossier temporaire créé: $TEMP_DIR"
    
    # Copier les fichiers construits
    cp -r build/* "$TEMP_DIR/"
    
    # Aller dans le dossier temporaire
    cd "$TEMP_DIR"
    
    # Initialiser git et ajouter les fichiers
    git init
    git add .
    git commit -m "Deploy Veza documentation $(date)"
    
    # Ajouter le remote et pousser
    git remote add origin "https://github.com/$GITHUB_REPO.git"
    git branch -M "$DEPLOY_BRANCH"
    git push -f origin "$DEPLOY_BRANCH"
    
    # Nettoyer
    cd - > /dev/null
    rm -rf "$TEMP_DIR"
    
    log_success "Déploiement terminé !"
    log_info "Le site sera disponible sur https://$GITHUB_REPO.github.io"
}

# Fonction pour déployer sur Vercel
deploy_vercel() {
    log_info "Déploiement sur Vercel..."
    
    # Vérifier si Vercel CLI est installé
    if ! command -v vercel &> /dev/null; then
        log_warning "Vercel CLI n'est pas installé"
        log_info "Installation de Vercel CLI..."
        npm install -g vercel
    fi
    
    # Déployer
    vercel --prod
    
    log_success "Déploiement Vercel terminé !"
}

# Fonction pour déployer sur Netlify
deploy_netlify() {
    log_info "Déploiement sur Netlify..."
    
    # Vérifier si Netlify CLI est installé
    if ! command -v netlify &> /dev/null; then
        log_warning "Netlify CLI n'est pas installé"
        log_info "Installation de Netlify CLI..."
        npm install -g netlify-cli
    fi
    
    # Déployer
    netlify deploy --prod --dir=build
    
    log_success "Déploiement Netlify terminé !"
}

# Fonction pour afficher l'aide
show_help() {
    echo "Script de déploiement pour la documentation Veza"
    echo ""
    echo "Usage: $0 [COMMANDE]"
    echo ""
    echo "Commandes:"
    echo "  local       - Démarrer le serveur de développement local"
    echo "  build       - Construire le site pour la production"
    echo "  github      - Déployer sur GitHub Pages"
    echo "  vercel      - Déployer sur Vercel"
    echo "  netlify     - Déployer sur Netlify"
    echo "  all         - Construire et déployer sur GitHub Pages"
    echo "  help        - Afficher cette aide"
    echo ""
    echo "Exemples:"
    echo "  $0 local"
    echo "  $0 build"
    echo "  $0 github"
    echo "  $0 all"
}

# Fonction principale
main() {
    local command=${1:-help}
    
    case $command in
        "local")
            check_prerequisites
            install_dependencies
            test_local
            ;;
        "build")
            check_prerequisites
            install_dependencies
            build_site
            log_success "Site construit dans le dossier 'build'"
            ;;
        "github")
            check_prerequisites
            install_dependencies
            build_site
            deploy_github_pages
            ;;
        "vercel")
            check_prerequisites
            install_dependencies
            build_site
            deploy_vercel
            ;;
        "netlify")
            check_prerequisites
            install_dependencies
            build_site
            deploy_netlify
            ;;
        "all")
            check_prerequisites
            install_dependencies
            build_site
            deploy_github_pages
            ;;
        "help"|*)
            show_help
            ;;
    esac
}

# Exécuter la fonction principale
main "$@" 