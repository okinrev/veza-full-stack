#!/bin/bash

# ╭──────────────────────────────────────────────────────────────╮
# │              🗄️ Veza Database Setup Script                  │
# │          Script d'initialisation PostgreSQL                 │
# ╰──────────────────────────────────────────────────────────────╯

set -e

# Couleurs pour l'affichage
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
CONTAINER_NAME="veza-postgres"
DB_NAME="veza_db"
DB_USER="veza_user"
DB_PASSWORD="veza_password"
SCRIPT_DIR="$(dirname "$0")"
SQL_FILE="$SCRIPT_DIR/init-database.sql"

# Fonctions utilitaires
print_header() {
    echo -e "${BLUE}╭─────────────────────────────────────────────────────────────╮${NC}"
    echo -e "${BLUE}│ $1${NC}"
    echo -e "${BLUE}╰─────────────────────────────────────────────────────────────╯${NC}"
}

print_step() {
    echo -e "${YELLOW}⚡ $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

# Vérification des prérequis
check_prerequisites() {
    print_step "Vérification des prérequis..."
    
    # Vérifier si le container existe
    if ! incus list | grep -q "$CONTAINER_NAME"; then
        print_error "Container $CONTAINER_NAME non trouvé !"
        print_info "Lancez d'abord: ./veza deploy"
        exit 1
    fi
    
    # Vérifier si le container est démarré
    if ! incus list | grep "$CONTAINER_NAME" | grep -q "RUNNING"; then
        print_error "Container $CONTAINER_NAME n'est pas démarré !"
        print_info "Lancez: ./veza start"
        exit 1
    fi
    
    # Vérifier si le fichier SQL existe
    if [ ! -f "$SQL_FILE" ]; then
        print_error "Fichier $SQL_FILE non trouvé !"
        exit 1
    fi
    
    print_success "Prérequis validés"
}

# Test de connexion PostgreSQL
test_postgres_connection() {
    print_step "Test de connexion PostgreSQL..."
    
    if incus exec "$CONTAINER_NAME" -- sudo -u postgres psql -c "SELECT version();" > /dev/null 2>&1; then
        print_success "PostgreSQL est accessible"
        return 0
    else
        print_error "Impossible de se connecter à PostgreSQL"
        return 1
    fi
}

# Attendre que PostgreSQL soit prêt
wait_for_postgres() {
    print_step "Attente du démarrage de PostgreSQL..."
    
    local attempts=0
    local max_attempts=30
    
    while [ $attempts -lt $max_attempts ]; do
        if test_postgres_connection; then
            return 0
        fi
        
        echo -n "."
        sleep 2
        attempts=$((attempts + 1))
    done
    
    print_error "PostgreSQL n'est pas prêt après $max_attempts tentatives"
    return 1
}

# Créer la base de données et l'utilisateur
setup_database() {
    print_step "Configuration de la base de données..."
    
    # Créer la base de données si elle n'existe pas
    incus exec "$CONTAINER_NAME" -- sudo -u postgres psql -c "
        CREATE DATABASE $DB_NAME;
    " 2>/dev/null || true
    
    # Créer l'utilisateur si il n'existe pas
    incus exec "$CONTAINER_NAME" -- sudo -u postgres psql -c "
        CREATE USER $DB_USER WITH PASSWORD '$DB_PASSWORD';
        GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;
        ALTER USER $DB_USER CREATEDB;
    " 2>/dev/null || true
    
    print_success "Base de données configurée"
}

# Exécuter le script d'initialisation
run_init_script() {
    print_step "Exécution du script d'initialisation..."
    
    # Copier le script SQL dans le container
    incus file push "$SQL_FILE" "$CONTAINER_NAME/tmp/init-database.sql"
    
    # Exécuter le script
    if incus exec "$CONTAINER_NAME" -- sudo -u postgres psql -d "$DB_NAME" -f /tmp/init-database.sql; then
        print_success "Script d'initialisation exécuté avec succès"
    else
        print_error "Erreur lors de l'exécution du script d'initialisation"
        return 1
    fi
    
    # Nettoyer le fichier temporaire
    incus exec "$CONTAINER_NAME" -- rm -f /tmp/init-database.sql
}

# Vérifier l'installation
verify_installation() {
    print_step "Vérification de l'installation..."
    
    # Tester quelques requêtes
    local tables_count=$(incus exec "$CONTAINER_NAME" -- sudo -u postgres psql -d "$DB_NAME" -t -c "
        SELECT COUNT(*) FROM information_schema.tables 
        WHERE table_schema = 'public' AND table_type = 'BASE TABLE';
    " | tr -d ' ')
    
    local users_count=$(incus exec "$CONTAINER_NAME" -- sudo -u postgres psql -d "$DB_NAME" -t -c "
        SELECT COUNT(*) FROM users;
    " | tr -d ' ')
    
    print_success "Tables créées: $tables_count"
    print_success "Utilisateurs de test: $users_count"
    
    if [ "$tables_count" -gt "10" ] && [ "$users_count" -gt "0" ]; then
        print_success "Installation vérifiée avec succès !"
        return 0
    else
        print_error "Problème détecté dans l'installation"
        return 1
    fi
}

# Afficher les informations de connexion
show_connection_info() {
    print_header "Informations de connexion"
    echo -e "${GREEN}🔑 Comptes utilisateurs créés:${NC}"
    echo -e "   ${YELLOW}Admin:${NC} admin@veza.local / password"
    echo -e "   ${YELLOW}Demo:${NC}  demo@veza.local / password"
    echo -e "   ${YELLOW}Test:${NC}  loulou@free.fr / password"
    echo ""
    echo -e "${GREEN}🗄️ Base de données:${NC}"
    echo -e "   ${YELLOW}Host:${NC}     veza-postgres.lxd (10.5.191.154)"
    echo -e "   ${YELLOW}Database:${NC} $DB_NAME"
    echo -e "   ${YELLOW}User:${NC}     $DB_USER"
    echo -e "   ${YELLOW}Password:${NC} $DB_PASSWORD"
    echo ""
    echo -e "${GREEN}📊 Tables créées:${NC}"
    echo -e "   • users, refresh_tokens"
    echo -e "   • categories, products, user_products"
    echo -e "   • tracks, shared_resources"
    echo -e "   • listings, offers"
    echo -e "   • rooms, room_members, messages"
    echo -e "   • tags, files"
}

# Fonction principale
main() {
    print_header "🗄️ Initialisation Base de Données Veza"
    
    check_prerequisites
    wait_for_postgres
    setup_database
    run_init_script
    verify_installation
    show_connection_info
    
    echo ""
    print_success "Base de données Veza initialisée avec succès ! 🚀"
    print_info "Vous pouvez maintenant redémarrer le backend: ./veza restart backend"
}

# Gestion des options
case "${1:-}" in
    --help|-h)
        echo "Usage: $0 [options]"
        echo ""
        echo "Options:"
        echo "  --help, -h    Afficher cette aide"
        echo "  --force, -f   Forcer la réinitialisation (supprime toutes les données)"
        echo ""
        echo "Ce script initialise la base de données PostgreSQL pour Veza."
        exit 0
        ;;
    --force|-f)
        print_header "🗄️ Réinitialisation Forcée de la Base de Données"
        print_error "ATTENTION: Ceci va supprimer TOUTES les données existantes !"
        read -p "Êtes-vous sûr ? (tapez 'CONFIRMER' pour continuer): " confirm
        if [ "$confirm" = "CONFIRMER" ]; then
            main
        else
            print_info "Opération annulée"
            exit 0
        fi
        ;;
    "")
        main
        ;;
    *)
        print_error "Option inconnue: $1"
        print_info "Utilisez --help pour voir les options disponibles"
        exit 1
        ;;
esac 