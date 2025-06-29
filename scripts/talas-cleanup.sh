#!/bin/bash

# ================================================================================
# TALAS CLEANUP SCRIPT - Nettoyage Complet et Réorganisation
# ================================================================================
# Ce script nettoie le projet et prépare l'architecture unifiée Talas

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

# Fonction d'affichage
log_info() { echo -e "${BLUE}ℹ️  $1${NC}"; }
log_success() { echo -e "${GREEN}✅ $1${NC}"; }
log_warning() { echo -e "${YELLOW}⚠️  $1${NC}"; }
log_error() { echo -e "${RED}❌ $1${NC}"; }
log_header() { echo -e "${PURPLE}🚀 $1${NC}"; }

show_header() {
    echo -e "${PURPLE}"
    echo "╭─────────────────────────────────────────────────╮"
    echo "│           🧹 TALAS CLEANUP SCRIPT              │"
    echo "│      Nettoyage et Réorganisation Complète      │"
    echo "╰─────────────────────────────────────────────────╯"
    echo -e "${NC}"
}

# Fonction de confirmation
confirm_action() {
    echo -e "${YELLOW}⚠️  Cette opération va modifier la structure du projet.${NC}"
    echo -e "${YELLOW}   Voulez-vous continuer ? (oui/non)${NC}"
    read -r response
    case "$response" in
        [oO][uU][iI]|[yY][eE][sS]|[yY])
            return 0
            ;;
        *)
            echo -e "${RED}❌ Opération annulée.${NC}"
            exit 1
            ;;
    esac
}

# 1. Archiver veza-basic-frontend
archive_basic_frontend() {
    log_header "1. Archivage de veza-basic-frontend"
    
    if [ -d "$PROJECT_ROOT/veza-basic-frontend" ]; then
        log_info "Création du dossier archive/ si nécessaire"
        mkdir -p "$PROJECT_ROOT/archive"
        
        log_info "Déplacement de veza-basic-frontend vers archive/"
        mv "$PROJECT_ROOT/veza-basic-frontend" "$PROJECT_ROOT/archive/"
        
        # Créer un README dans archive
        cat > "$PROJECT_ROOT/archive/README.md" << 'EOF'
# Archive Talas

Ce dossier contient les anciens composants archivés du projet Talas.

## Contenu

### veza-basic-frontend/
Ancien frontend HTML/Alpine.js - **À CONSERVER UNIQUEMENT COMME RÉFÉRENCE**

**⚠️ IMPORTANT :** Ce frontend est obsolète et ne doit plus être utilisé.
Le frontend actif est maintenant `veza-frontend/` (React + TypeScript).

### Historique
- Archivé le : $(date)
- Raison : Migration vers architecture unifiée Talas
- Status : Archive de référence uniquement

EOF
        
        log_success "veza-basic-frontend archivé avec succès"
    else
        log_warning "veza-basic-frontend n'existe pas, rien à archiver"
    fi
}

# 2. Nettoyer le dossier scripts/
clean_scripts_directory() {
    log_header "2. Nettoyage du dossier scripts/"
    
    cd "$PROJECT_ROOT/scripts"
    
    # Scripts fonctionnels à conserver
    KEEP_SCRIPTS=(
        "talas-admin.sh"
        "talas-cleanup.sh"
        "veza-manager.sh"
        "complete-setup.sh"
        "build-and-start.sh"
        "start-unified-services.sh"
        "test-all-connections.sh"
        "sync-env-config.sh"
        "setup-database.sh"
        "init-database.sql"
        "create-test-user.sh"
        "install-dependencies.sh"
        "network-fix.sh"
        "deploy-base-containers.sh"
        "update-source-code.sh"
    )
    
    # Créer dossier archive si nécessaire
    mkdir -p "archive-old-scripts"
    
    # Archiver les scripts non essentiels
    log_info "Archivage des scripts obsolètes..."
    
    for script in *.sh *.md *.conf; do
        if [ -f "$script" ]; then
            # Vérifier si le script doit être conservé
            should_keep=false
            for keep_script in "${KEEP_SCRIPTS[@]}"; do
                if [[ "$script" == "$keep_script" ]]; then
                    should_keep=true
                    break
                fi
            done
            
            if [ "$should_keep" = false ]; then
                if [[ ! "$script" =~ ^archive- ]]; then
                    log_info "Archivage: $script"
                    mv "$script" "archive-old-scripts/"
                fi
            fi
        fi
    done
    
    # Créer la nouvelle structure organisée
    mkdir -p deploy test maintenance
    
    log_success "Dossier scripts/ nettoyé et organisé"
}

# 3. Supprimer fichiers temporaires et inutiles
clean_temp_files() {
    log_header "3. Suppression des fichiers temporaires"
    
    cd "$PROJECT_ROOT"
    
    # Fichiers temporaires à supprimer
    TEMP_PATTERNS=(
        "**/*.tmp"
        "**/*.log"
        "**/node_modules/.cache"
        "**/target/debug"
        "**/target/release"
        "**/.DS_Store"
        "**/Thumbs.db"
        "**/*.swp"
        "**/*.swo"
        "**/*~"
        "logs/*.log"
        "uploads/temp_*"
        "backups/temp_*"
        "**/core.*"
        "**/.pytest_cache"
        "**/__pycache__"
    )
    
    log_info "Nettoyage des fichiers temporaires..."
    
    for pattern in "${TEMP_PATTERNS[@]}"; do
        find . -path "./.*" -prune -o -name "$pattern" -print0 2>/dev/null | xargs -0 rm -rf 2>/dev/null || true
    done
    
    # Nettoyer les dossiers de build spécifiques
    if [ -d "veza-chat-server/target" ]; then
        log_info "Nettoyage du cache Rust chat-server..."
        rm -rf veza-chat-server/target/debug 2>/dev/null || true
    fi
    
    if [ -d "veza-stream-server/target" ]; then
        log_info "Nettoyage du cache Rust stream-server..."
        rm -rf veza-stream-server/target/debug 2>/dev/null || true
    fi
    
    if [ -d "veza-frontend/node_modules/.cache" ]; then
        log_info "Nettoyage du cache npm frontend..."
        rm -rf veza-frontend/node_modules/.cache 2>/dev/null || true
    fi
    
    log_success "Fichiers temporaires supprimés"
}

# 4. Créer la structure documentaire unifiée
create_docs_structure() {
    log_header "4. Création de la structure documentaire"
    
    cd "$PROJECT_ROOT"
    
    # Créer la structure docs/ unifiée
    mkdir -p docs/{integration,api,deployment,architecture}
    
    # Déplacer et organiser la documentation existante
    if [ -f "GUIDE_DEPLOYMENT_FINAL.md" ]; then
        mv "GUIDE_DEPLOYMENT_FINAL.md" "docs/deployment/"
    fi
    
    if [ -f "archi.md" ]; then
        mv "archi.md" "docs/architecture/"
    fi
    
    # Créer les fichiers de documentation de base
    cat > "docs/README.md" << 'EOF'
# Documentation Talas

## Structure

- `integration/` - Guides d'intégration entre services
- `api/` - Documentation des APIs
- `deployment/` - Guides de déploiement
- `architecture/` - Documentation technique et architecturale

## Fichiers Principaux

- `INTEGRATION.md` - Guide d'intégration complète (À CRÉER)
- `API.md` - Documentation API unifiée (À CRÉER)
- `DEPLOYMENT.md` - Guide de déploiement unifié (À CRÉER)
EOF
    
    log_success "Structure documentaire créée"
}

# 5. Nettoyer les configurations redondantes
clean_configs() {
    log_header "5. Nettoyage des configurations"
    
    cd "$PROJECT_ROOT/configs"
    
    # Créer backup des configs importantes
    log_info "Sauvegarde des configurations importantes..."
    cp env.unified "env.unified.backup.$TIMESTAMP"
    
    # Supprimer les configs obsolètes/redondantes
    OBSOLETE_CONFIGS=(
        "env.old"
        "env.backup"
        "*.bak"
        "haproxy.cfg.old"
    )
    
    for config in "${OBSOLETE_CONFIGS[@]}"; do
        rm -f $config 2>/dev/null || true
    done
    
    log_success "Configurations nettoyées"
}

# 6. Créer le résumé du nettoyage
create_cleanup_summary() {
    log_header "6. Création du rapport de nettoyage"
    
    cat > "$PROJECT_ROOT/CLEANUP_REPORT_$TIMESTAMP.md" << EOF
# Rapport de Nettoyage Talas - $TIMESTAMP

## Actions Effectuées

### ✅ Archivage
- \`veza-basic-frontend/\` → \`archive/veza-basic-frontend/\`
- Scripts obsolètes → \`scripts/archive-old-scripts/\`

### ✅ Nettoyage
- Fichiers temporaires supprimés
- Caches de build nettoyés
- Configurations redondantes supprimées

### ✅ Réorganisation
- Structure \`docs/\` unifiée créée
- Scripts organisés par fonction
- Configuration centralisée

## Structure Finale

\`\`\`
veza-full-stack/
├── veza-backend-api/      # API Go principale
├── veza-frontend/         # Frontend React actif
├── veza-chat-server/      # Chat WebSocket Rust
├── veza-stream-server/    # Stream Audio Rust
├── archive/
│   └── veza-basic-frontend/  # Ancien frontend archivé
├── scripts/
│   ├── talas-admin.sh     # Script principal (À CRÉER)
│   ├── deploy/            # Scripts de déploiement
│   ├── test/              # Scripts de test
│   └── maintenance/       # Scripts de maintenance
├── configs/
│   ├── env.unified        # Variables d'environnement principales
│   └── [autres configs]
└── docs/
    ├── INTEGRATION.md     # Guide d'intégration (À CRÉER)
    ├── API.md            # Documentation API (À CRÉER)
    ├── DEPLOYMENT.md     # Guide de déploiement (À CRÉER)
    └── [dossiers thématiques]
\`\`\`

## Prochaines Étapes

1. ✅ Créer \`scripts/talas-admin.sh\`
2. ✅ Unifier l'authentification JWT
3. ✅ Implémenter tests d'intégration
4. ✅ Créer documentation technique

## Backups Créés

- \`configs/env.unified.backup.$TIMESTAMP\`

---
Nettoyage effectué automatiquement par \`scripts/talas-cleanup.sh\`
EOF
    
    log_success "Rapport de nettoyage créé: CLEANUP_REPORT_$TIMESTAMP.md"
}

# Fonction principale
main() {
    show_header
    
    log_info "Début du nettoyage du projet Talas..."
    log_info "Répertoire du projet: $PROJECT_ROOT"
    
    confirm_action
    
    # Exécution des étapes
    archive_basic_frontend
    clean_scripts_directory
    clean_temp_files
    create_docs_structure
    clean_configs
    create_cleanup_summary
    
    echo ""
    log_header "🎉 Nettoyage Talas terminé avec succès !"
    echo ""
    log_success "Le projet est maintenant prêt pour l'intégration unifiée."
    log_info "Consultez le rapport: CLEANUP_REPORT_$TIMESTAMP.md"
    echo ""
    log_info "Prochaines étapes:"
    echo "  1. Créer scripts/talas-admin.sh"
    echo "  2. Unifier l'authentification JWT"
    echo "  3. Implémenter les tests d'intégration"
    echo ""
}

# Vérification des arguments
if [ "$1" = "--help" ] || [ "$1" = "-h" ]; then
    show_header
    echo "Usage: $0"
    echo ""
    echo "Ce script effectue un nettoyage complet du projet Talas :"
    echo "- Archive veza-basic-frontend"
    echo "- Nettoie le dossier scripts/"
    echo "- Supprime les fichiers temporaires"
    echo "- Réorganise la documentation"
    echo "- Crée un rapport détaillé"
    exit 0
fi

# Exécution
main "$@" 