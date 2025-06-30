#!/bin/bash

# Configuration et création de la base de données Veza pour le développement

set -e

echo "🗄️  Configuration de la base de données Veza..."

# Variables de configuration
DB_NAME="veza_dev"
DB_USER="veza_user"
DB_PASSWORD="veza_password"
DB_HOST="localhost"
DB_PORT="5432"

# Couleurs pour les logs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Vérifier si PostgreSQL est installé
check_postgresql() {
    log_info "Vérification de PostgreSQL..."
    
    if ! command -v psql &> /dev/null; then
        log_error "PostgreSQL n'est pas installé"
        log_info "Installation sur Ubuntu/Debian: sudo apt-get install postgresql postgresql-contrib"
        log_info "Installation sur CentOS/RHEL: sudo yum install postgresql postgresql-server"
        log_info "Installation sur macOS: brew install postgresql"
        exit 1
    fi
    
    log_success "PostgreSQL est installé"
}

# Vérifier si le service PostgreSQL fonctionne
check_postgresql_service() {
    log_info "Vérification du service PostgreSQL..."
    
    if ! systemctl is-active --quiet postgresql && ! pgrep -x postgres > /dev/null; then
        log_warning "Service PostgreSQL non démarré"
        log_info "Tentative de démarrage..."
        
        if command -v systemctl &> /dev/null; then
            sudo systemctl start postgresql
        else
            log_info "Veuillez démarrer PostgreSQL manuellement"
            exit 1
        fi
    fi
    
    log_success "Service PostgreSQL actif"
}

# Créer l'utilisateur et la base de données
setup_database() {
    log_info "Configuration de la base de données..."
    
    # Se connecter comme utilisateur postgres et créer l'utilisateur
    sudo -u postgres psql -c "CREATE USER $DB_USER WITH PASSWORD '$DB_PASSWORD';" 2>/dev/null || {
        log_warning "L'utilisateur $DB_USER existe déjà"
    }
    
    # Donner les privilèges
    sudo -u postgres psql -c "ALTER USER $DB_USER CREATEDB;"
    
    # Créer la base de données
    sudo -u postgres psql -c "CREATE DATABASE $DB_NAME OWNER $DB_USER;" 2>/dev/null || {
        log_warning "La base de données $DB_NAME existe déjà"
    }
    
    # Accorder tous les privilèges
    sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;"
    
    log_success "Base de données configurée"
}

# Tester la connexion
test_connection() {
    log_info "Test de connexion à la base de données..."
    
    export PGPASSWORD="$DB_PASSWORD"
    
    if psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "SELECT version();" > /dev/null 2>&1; then
        log_success "Connexion à la base de données réussie"
    else
        log_error "Impossible de se connecter à la base de données"
        exit 1
    fi
}

# Créer les tables de base
create_basic_tables() {
    log_info "Création des tables de base..."
    
    export PGPASSWORD="$DB_PASSWORD"
    
    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME << 'EOF'
-- Table des utilisateurs
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(30) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    bio TEXT,
    avatar VARCHAR(500),
    role VARCHAR(20) DEFAULT 'user' NOT NULL,
    status VARCHAR(20) DEFAULT 'active' NOT NULL,
    is_active BOOLEAN DEFAULT true NOT NULL,
    is_verified BOOLEAN DEFAULT false NOT NULL,
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

-- Table des tokens de rafraîchissement
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(500) UNIQUE NOT NULL,
    expires_at BIGINT NOT NULL,
    created_at BIGINT DEFAULT extract(epoch from now()) NOT NULL
);

-- Table des salles de chat
CREATE TABLE IF NOT EXISTS rooms (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_private BOOLEAN DEFAULT false NOT NULL,
    max_members INTEGER DEFAULT 50,
    created_by BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

-- Table des messages
CREATE TABLE IF NOT EXISTS messages (
    id BIGSERIAL PRIMARY KEY,
    room_id BIGINT NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id),
    content TEXT NOT NULL,
    message_type VARCHAR(20) DEFAULT 'text' NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

-- Index pour performances
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token ON refresh_tokens(token);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_messages_room_id ON messages(room_id);
CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);

EOF
    
    log_success "Tables créées avec succès"
}

# Menu principal
main() {
    echo "🚀 Configuration de la base de données Veza"
    echo "=========================================="
    
    check_postgresql
    check_postgresql_service
    setup_database
    test_connection
    create_basic_tables
    
    echo ""
    log_success "Configuration terminée !"
    echo ""
    log_info "Détails de connexion :"
    echo "  Host: $DB_HOST"
    echo "  Port: $DB_PORT"
    echo "  Database: $DB_NAME"
    echo "  User: $DB_USER"
    echo "  Password: $DB_PASSWORD"
    echo ""
    log_info "URL de connexion :"
    echo "  postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable"
    echo ""
}

# Exécuter le script principal
main "$@" 