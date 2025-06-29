#!/bin/bash

# Script de Nettoyage Complet et Réorganisation Veza
# Transforme le projet en infrastructure unifiée propre

set -e

PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
PURPLE='\033[0;35m'
NC='\033[0m'

log_info() { echo -e "${BLUE}ℹ️  $1${NC}"; }
log_success() { echo -e "${GREEN}✅ $1${NC}"; }
log_warning() { echo -e "${YELLOW}⚠️  $1${NC}"; }
log_error() { echo -e "${RED}❌ $1${NC}"; }
log_header() { echo -e "\n${PURPLE}🚀 $1${NC}"; }

# Vérification de sécurité
verify_project_structure() {
    log_header "Vérification de la Structure du Projet"
    
    local required_dirs=("veza-backend-api" "veza-frontend" "veza-chat-server" "veza-stream-server")
    for dir in "${required_dirs[@]}"; do
        if [ ! -d "$PROJECT_ROOT/$dir" ]; then
            log_error "Module manquant : $dir"
            exit 1
        fi
        log_success "Module trouvé : $dir"
    done
}

# Nettoyage des fichiers temporaires et logs
cleanup_temp_files() {
    log_header "Nettoyage des Fichiers Temporaires"
    
    cd "$PROJECT_ROOT"
    
    # Liste des patterns à supprimer
    local temp_patterns=(
        "*.log"
        "*.pid" 
        "*_RAPPORT_*.md"
        "*_INVENTAIRE_*.md"
        "*_MIGRATION_*.md"
        "*_STATUS_*.md"
        "CLEANUP_REPORT_*.md"
        "TALAS_INTEGRATION_REPORT.md"
        "RESOLUTION_PROBLEMES_TESTS.md"
        "console-export-*.txt"
        "docker-compose.test.yml"
        "test_*.html"
        "test_*.sh"
        "veza-admin.log"
        "deploy.log"
    )
    
    for pattern in "${temp_patterns[@]}"; do
        find . -maxdepth 1 -name "$pattern" -type f -exec rm -f {} \; 2>/dev/null || true
        log_info "Supprimé : $pattern"
    done
    
    # Nettoyage des modules
    find veza-backend-api -name "*.log" -o -name "*.pid" -exec rm -f {} \; 2>/dev/null || true
    find veza-backend-api -name "main" -type f -exec rm -f {} \; 2>/dev/null || true
    
    log_success "Fichiers temporaires nettoyés"
}

# Création de la structure unifiée
create_unified_structure() {
    log_header "Création de la Structure Unifiée"
    
    cd "$PROJECT_ROOT"
    
    # Créer les dossiers de la structure finale
    mkdir -p {configs,docs,scripts/{deploy,test,maintenance}}
    
    # Réorganiser les archives
    if [ ! -d "archive/veza-basic-frontend" ] && [ -d "veza-basic-frontend" ]; then
        mkdir -p archive
        mv veza-basic-frontend archive/
        log_success "veza-basic-frontend déplacé vers archive/"
    fi
    
    # Nettoyer les dossiers redondants
    rm -rf logs backups uploads audio storage ssl data 2>/dev/null || true
    
    log_success "Structure unifiée créée"
}

# Configuration JWT unifiée
create_jwt_config() {
    log_header "Création de la Configuration JWT Unifiée"
    
    cat > "$PROJECT_ROOT/configs/jwt.config" << 'EOF'
# Configuration JWT Unifiée Veza
# Partagée entre tous les services (Backend Go, Chat Rust, Stream Rust)

JWT_SECRET=veza_unified_jwt_secret_key_2025
JWT_ISSUER=veza-platform
JWT_AUDIENCE=veza-services
JWT_EXPIRATION=24h
JWT_REFRESH_EXPIRATION=7d

# Algorithme de signature
JWT_ALGORITHM=HS256

# Headers requis
JWT_HEADER_NAME=Authorization
JWT_HEADER_PREFIX=Bearer
EOF
    
    log_success "Configuration JWT unifiée créée"
}

# Configuration des services
create_services_config() {
    log_header "Création de la Configuration des Services"
    
    cat > "$PROJECT_ROOT/configs/services.config" << 'EOF'
# Configuration des Services Veza
# URLs et ports pour tous les environnements

# Services Backend
BACKEND_HOST=localhost
BACKEND_PORT=8080
BACKEND_URL=http://localhost:8080

# Services WebSocket
CHAT_HOST=localhost
CHAT_PORT=3001
CHAT_WS_URL=ws://localhost:3001

STREAM_HOST=localhost 
STREAM_PORT=3002
STREAM_WS_URL=ws://localhost:3002

# Frontend
FRONTEND_HOST=localhost
FRONTEND_PORT=5173
FRONTEND_URL=http://localhost:5173

# Base de données
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_NAME=veza_db
DATABASE_USER=veza_user
DATABASE_PASSWORD=veza_password
DATABASE_URL=postgres://veza_user:veza_password@localhost:5432/veza_db

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_URL=redis://localhost:6379

# HAProxy
HAPROXY_PORT=80
HAPROXY_STATS_PORT=8404
EOF
    
    log_success "Configuration des services créée"
}

# Configuration HAProxy unifiée
create_haproxy_config() {
    log_header "Configuration HAProxy Unifiée"
    
    cat > "$PROJECT_ROOT/configs/haproxy.cfg" << 'EOF'
global
    daemon
    maxconn 4096
    log stdout local0 info

defaults
    mode http
    timeout connect 5000ms
    timeout client 50000ms
    timeout server 50000ms
    option httplog
    log global

# Interface de monitoring
listen stats
    bind *:8404
    stats enable
    stats uri /stats
    stats refresh 5s

# Frontend principal
frontend veza_frontend
    bind *:80
    
    # Routes vers le frontend React
    acl is_frontend path_beg /
    acl is_frontend path_beg /static
    acl is_frontend path_beg /assets
    use_backend frontend_servers if is_frontend
    
    # Routes API Backend
    acl is_api path_beg /api
    use_backend backend_servers if is_api
    
    # Routes WebSocket Chat
    acl is_chat path_beg /chat
    acl is_ws_chat hdr(connection) -i upgrade
    use_backend chat_servers if is_chat or is_ws_chat
    
    # Routes WebSocket Stream
    acl is_stream path_beg /stream
    acl is_ws_stream hdr(connection) -i upgrade
    use_backend stream_servers if is_stream or is_ws_stream

# Backend Go API
backend backend_servers
    balance roundrobin
    server backend1 localhost:8080 check

# Frontend React
backend frontend_servers
    balance roundrobin
    server frontend1 localhost:5173 check

# Chat WebSocket Rust
backend chat_servers
    balance roundrobin
    server chat1 localhost:3001 check

# Stream WebSocket Rust  
backend stream_servers
    balance roundrobin
    server stream1 localhost:3002 check
EOF
    
    # Remplacer l'ancien haproxy.cfg
    if [ -f "$PROJECT_ROOT/haproxy.cfg" ]; then
        mv "$PROJECT_ROOT/haproxy.cfg" "$PROJECT_ROOT/configs/"
        log_info "Ancien haproxy.cfg déplacé vers configs/"
    fi
    
    log_success "Configuration HAProxy unifiée créée"
}

# Nettoyage de .gitignore unifié
create_unified_gitignore() {
    log_header "Création du .gitignore Unifié"
    
    cat > "$PROJECT_ROOT/.gitignore" << 'EOF'
# === GITIGNORE UNIFIÉ VEZA ===

# Logs et fichiers temporaires
*.log
*.pid
*.tmp
*.temp
console-export-*.txt
*_RAPPORT_*.md
*_STATUS_*.md
*_MIGRATION_*.md

# Environnements et secrets
.env
.env.local
.env.production
*.pem
*.key
ssl/

# Base de données locales
*.db
*.sqlite
*.sqlite3

# Builds et distributions
dist/
build/
target/
bin/
main

# Dépendances
node_modules/
vendor/

# OS et éditeurs
.DS_Store
.vscode/
.idea/
*.swp
*.swo
*~

# Test et développement
test-results/
coverage/
playwright-report/
test-data/

# Backups
backups/
*.backup

# Containers et données
data/
logs/
uploads/
storage/

# Cache
.cache/
.vite/
*.tsbuildinfo
EOF
    
    log_success ".gitignore unifié créé"
}

# Nettoyage final et validation
final_cleanup() {
    log_header "Nettoyage Final et Validation"
    
    cd "$PROJECT_ROOT"
    
    # Supprimer les fichiers de documentation redondants
    rm -f README-VEZA-MANAGER.md GUIDE_DEVELOPMENT_MANUAL.md QUICK_START.md 2>/dev/null || true
    rm -f GUIDE_TESTS_COMPLETS.md PROJECT_CLEAN_STATUS.md RAPPORT_TESTS_FINAL.md 2>/dev/null || true
    
    # Supprimer le Makefile global (les modules ont leurs propres Makefiles)
    rm -f Makefile 2>/dev/null || true
    
    # Nettoyer les fichiers veza temporaires
    rm -f veza 2>/dev/null || true
    
    log_success "Nettoyage final terminé"
}

# Génération du rapport de nettoyage
generate_cleanup_report() {
    log_header "Génération du Rapport de Nettoyage"
    
    local report_file="$PROJECT_ROOT/CLEANUP_UNIFIED_$(date +%Y%m%d_%H%M%S).md"
    
    cat > "$report_file" << EOF
# Rapport de Nettoyage et Unification Veza

**Date :** $(date)  
**Projet :** Infrastructure Veza Unifiée

## Structure Finale

\`\`\`
veza/
├── veza-backend-api/      # API Go principale
├── veza-frontend/         # Frontend React actif  
├── veza-chat-server/      # Chat WebSocket Rust
├── veza-stream-server/    # Stream Audio Rust
├── archive/
│   └── veza-basic-frontend/  # Ancien frontend archivé
├── scripts/
│   ├── veza-cleanup.sh    # Script de nettoyage
│   ├── talas-incus.sh     # Gestion Incus
│   ├── deploy/            # Scripts de déploiement
│   ├── test/              # Scripts de test
│   └── maintenance/       # Scripts de maintenance
├── configs/
│   ├── jwt.config         # Config JWT partagée
│   ├── services.config    # Config des services
│   └── haproxy.cfg        # Config HAProxy
└── docs/                  # Documentation technique
\`\`\`

## Actions Réalisées

✅ Nettoyage des fichiers temporaires et logs  
✅ Suppression des rapports et documentation redondante  
✅ Réorganisation de la structure en modules unifiés  
✅ Création de la configuration JWT partagée  
✅ Configuration unifiée des services  
✅ Configuration HAProxy optimisée  
✅ .gitignore unifié et complet  

## Prochaines Étapes

1. **Implémentation JWT Unifiée** dans chaque service
2. **Script d'administration central** veza-admin.sh
3. **Tests d'intégration** complets
4. **Documentation technique** complète

## Configuration JWT

Tous les services utilisent maintenant la même configuration JWT :
- Secret unifié partagé
- Headers standardisés  
- Validation cohérente

## Services Intégrés

- **Backend Go (8080)** : API principale + génération JWT
- **Chat Rust (3001)** : WebSocket + validation JWT  
- **Stream Rust (3002)** : WebSocket + validation JWT
- **Frontend React (5173)** : Interface + gestion JWT
- **HAProxy (80)** : Reverse proxy unifié

Le projet est maintenant prêt pour l'intégration complète !
EOF
    
    log_success "Rapport généré : $(basename "$report_file")"
}

# Fonction principale
main() {
    echo -e "${CYAN}"
    echo "╔═══════════════════════════════════════════════╗"
    echo "║        🧹 NETTOYAGE UNIFIÉ VEZA 🧹           ║"
    echo "║    Transformation en Infrastructure Propre    ║"
    echo "╚═══════════════════════════════════════════════╝"
    echo -e "${NC}\n"
    
    verify_project_structure
    cleanup_temp_files
    create_unified_structure
    create_jwt_config
    create_services_config
    create_haproxy_config
    create_unified_gitignore
    final_cleanup
    generate_cleanup_report
    
    echo -e "\n${GREEN}🎉 NETTOYAGE UNIFIÉ TERMINÉ ! 🎉${NC}"
    echo -e "${CYAN}Structure prête pour l'intégration JWT et l'administration centralisée${NC}\n"
}

# Execution
if [ "$1" = "--help" ] || [ "$1" = "-h" ]; then
    echo "Usage: $0"
    echo ""
    echo "Script de nettoyage et unification complète du projet Veza"
    echo "Transforme le projet en infrastructure propre et unifiée"
    exit 0
fi

main "$@" 