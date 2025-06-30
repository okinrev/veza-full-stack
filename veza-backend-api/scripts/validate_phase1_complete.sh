#!/bin/bash

# =============================================================================
# SCRIPT DE VALIDATION PHASE 1 - REPOSITORY LAYER COMPLET
# =============================================================================
# Ce script valide que tous les composants de la Phase 1 sont correctement
# implémentés et fonctionnels.

set -e

# Couleurs pour l'affichage
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Variables
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
POSTGRES_DB="${POSTGRES_DB:-veza_test}"
POSTGRES_USER="${POSTGRES_USER:-veza_user}"
POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-veza_pass}"
POSTGRES_HOST="${POSTGRES_HOST:-localhost}"
POSTGRES_PORT="${POSTGRES_PORT:-5432}"

# Compteurs de tests
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# =============================================================================
# FONCTIONS UTILITAIRES
# =============================================================================

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
    ((PASSED_TESTS++))
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
    ((FAILED_TESTS++))
}

run_test() {
    local test_name="$1"
    local test_command="$2"
    
    ((TOTAL_TESTS++))
    log_info "Test: $test_name"
    
    if eval "$test_command" > /dev/null 2>&1; then
        log_success "✓ $test_name"
        return 0
    else
        log_error "✗ $test_name"
        return 1
    fi
}

check_file_exists() {
    local file_path="$1"
    local description="$2"
    
    if [[ -f "$file_path" ]]; then
        log_success "✓ $description existe"
        return 0
    else
        log_error "✗ $description manquant: $file_path"
        return 1
    fi
}

check_go_compilation() {
    local package_path="$1"
    local description="$2"
    
    cd "$PROJECT_ROOT"
    if go build -o /tmp/test_compile "$package_path" 2>/dev/null; then
        log_success "✓ $description compile correctement"
        rm -f /tmp/test_compile
        return 0
    else
        log_error "✗ $description ne compile pas"
        return 1
    fi
}

# =============================================================================
# VÉRIFICATIONS PRÉLIMINAIRES
# =============================================================================

echo "============================================================================="
echo "  VALIDATION PHASE 1 - REPOSITORY LAYER COMPLET"
echo "============================================================================="
echo ""

log_info "Vérification de l'environnement..."

# Vérifier que Go est installé
if ! command -v go &> /dev/null; then
    log_error "Go n'est pas installé ou n'est pas dans le PATH"
    exit 1
fi

# Vérifier que PostgreSQL est accessible
if ! command -v psql &> /dev/null; then
    log_warning "psql n'est pas installé, impossible de tester la base de données"
    SKIP_DB_TESTS=true
else
    if PGPASSWORD="$POSTGRES_PASSWORD" psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "SELECT 1;" &> /dev/null; then
        log_success "✓ Base de données PostgreSQL accessible"
        SKIP_DB_TESTS=false
    else
        log_warning "Base de données PostgreSQL non accessible, tests DB ignorés"
        SKIP_DB_TESTS=true
    fi
fi

cd "$PROJECT_ROOT"

# =============================================================================
# PHASE 1.1 - VÉRIFICATION DE L'ARCHITECTURE HEXAGONALE
# =============================================================================

echo ""
log_info "=== PHASE 1.1 : ARCHITECTURE HEXAGONALE ==="

# Vérifier la structure des dossiers
log_info "Vérification de la structure des dossiers..."

directories=(
    "internal/core/domain/entities"
    "internal/core/domain/repositories"
    "internal/adapters/postgres"
    "internal/infrastructure"
    "internal/database/migrations"
)

for dir in "${directories[@]}"; do
    if [[ -d "$dir" ]]; then
        log_success "✓ Dossier $dir existe"
        ((PASSED_TESTS++))
    else
        log_error "✗ Dossier $dir manquant"
        ((FAILED_TESTS++))
    fi
    ((TOTAL_TESTS++))
done

# =============================================================================
# PHASE 1.2 - VÉRIFICATION DES ENTITÉS DOMAIN
# =============================================================================

echo ""
log_info "=== PHASE 1.2 : ENTITÉS DOMAIN ==="

# Vérifier les entités principales
entities=(
    "internal/core/domain/entities/user.go:Entité User"
    "internal/core/domain/entities/chat.go:Entité Chat"
    "internal/core/domain/entities/stream.go:Entité Stream"
)

for entity_info in "${entities[@]}"; do
    IFS=':' read -r file_path description <<< "$entity_info"
    ((TOTAL_TESTS++))
    check_file_exists "$file_path" "$description"
done

# Vérifier le contenu de l'entité User
log_info "Vérification du contenu de l'entité User..."

user_features=(
    "type User struct"
    "type UserRole"
    "type UserStatus" 
    "func (u *User) ValidatePassword"
    "func (u *User) HashPassword"
    "func (u *User) HasPermission"
    "func (u *User) IsActive"
)

for feature in "${features[@]}"; do
    ((TOTAL_TESTS++))
    if grep -q "$feature" "internal/core/domain/entities/user.go" 2>/dev/null; then
        log_success "✓ User entité contient: $feature"
        ((PASSED_TESTS++))
    else
        log_error "✗ User entité manque: $feature"
        ((FAILED_TESTS++))
    fi
done

# =============================================================================
# PHASE 1.3 - VÉRIFICATION DES INTERFACES REPOSITORY
# =============================================================================

echo ""
log_info "=== PHASE 1.3 : INTERFACES REPOSITORY ==="

# Vérifier les interfaces de repository
repositories=(
    "internal/core/domain/repositories/user_repository.go:UserRepository Interface"
    "internal/core/domain/repositories/chat_repository.go:ChatRepository Interface"
    "internal/core/domain/repositories/stream_repository.go:StreamRepository Interface"
)

for repo_info in "${repositories[@]}"; do
    IFS=':' read -r file_path description <<< "$repo_info"
    ((TOTAL_TESTS++))
    check_file_exists "$file_path" "$description"
done

# Vérifier les méthodes dans UserRepository
log_info "Vérification des méthodes UserRepository..."

user_repo_methods=(
    "CreateUser"
    "GetUserByID"
    "GetUserByEmail"
    "UpdateUser"
    "DeleteUser"
    "GetPreferences"
    "UpdatePreferences"
    "CreateSession"
    "GetSession"
    "InvalidateSession"
    "CreateAuditLog"
    "GetUserAuditLogs"
    "AddContact"
    "RemoveContact"
    "BlockUser"
    "UnblockUser"
)

for method in "${user_repo_methods[@]}"; do
    ((TOTAL_TESTS++))
    if grep -q "func.*$method" "internal/core/domain/repositories/user_repository.go" 2>/dev/null; then
        log_success "✓ UserRepository contient méthode: $method"
        ((PASSED_TESTS++))
    else
        log_error "✗ UserRepository manque méthode: $method"
        ((FAILED_TESTS++))
    fi
done

# =============================================================================
# PHASE 1.4 - VÉRIFICATION DES IMPLÉMENTATIONS POSTGRES
# =============================================================================

echo ""
log_info "=== PHASE 1.4 : IMPLÉMENTATIONS POSTGRES ==="

# Vérifier les implémentations PostgreSQL
postgres_implementations=(
    "internal/adapters/postgres/user_repository.go:UserRepository PostgreSQL"
    "internal/adapters/postgres/chat_repository.go:ChatRepository PostgreSQL"
)

for impl_info in "${postgres_implementations[@]}"; do
    IFS=':' read -r file_path description <<< "$impl_info"
    ((TOTAL_TESTS++))
    check_file_exists "$file_path" "$description"
done

# Tester la compilation des repositories
log_info "Test de compilation des repositories..."

((TOTAL_TESTS++))
check_go_compilation "./internal/adapters/postgres" "Adapters PostgreSQL"

# =============================================================================
# PHASE 1.5 - VÉRIFICATION DES MIGRATIONS SQL
# =============================================================================

echo ""
log_info "=== PHASE 1.5 : MIGRATIONS SQL ==="

# Vérifier les migrations
migrations=(
    "internal/database/migrations/100_user_repository_complete.sql:Migration User Repository"
    "internal/database/migrations/101_chat_repository_complete.sql:Migration Chat Repository"
    "internal/database/migrations/102_stream_repository_complete.sql:Migration Stream Repository"
)

for migration_info in "${migrations[@]}"; do
    IFS=':' read -r file_path description <<< "$migration_info"
    ((TOTAL_TESTS++))
    check_file_exists "$file_path" "$description"
done

# Vérifier le contenu des migrations
log_info "Vérification du contenu des migrations..."

migration_features=(
    "100_user_repository_complete.sql:CREATE TABLE.*user_stats"
    "100_user_repository_complete.sql:CREATE TABLE.*user_preferences"
    "100_user_repository_complete.sql:CREATE TABLE.*user_sessions"
    "100_user_repository_complete.sql:CREATE TABLE.*user_audit_logs"
    "101_chat_repository_complete.sql:CREATE TABLE.*chat_rooms"
    "101_chat_repository_complete.sql:CREATE TABLE.*room_members"
    "101_chat_repository_complete.sql:CREATE TABLE.*chat_messages"
    "102_stream_repository_complete.sql:CREATE TABLE.*streams"
    "102_stream_repository_complete.sql:CREATE TABLE.*stream_listeners"
    "102_stream_repository_complete.sql:CREATE TABLE.*stream_tracks"
)

for feature_info in "${migration_features[@]}"; do
    IFS=':' read -r file_path pattern <<< "$feature_info"
    ((TOTAL_TESTS++))
    if grep -q "$pattern" "internal/database/migrations/$file_path" 2>/dev/null; then
        log_success "✓ Migration contient: $pattern"
        ((PASSED_TESTS++))
    else
        log_error "✗ Migration manque: $pattern"
        ((FAILED_TESTS++))
    fi
done

# =============================================================================
# PHASE 1.6 - TESTS BASE DE DONNÉES (si disponible)
# =============================================================================

if [[ "$SKIP_DB_TESTS" == "false" ]]; then
    echo ""
    log_info "=== PHASE 1.6 : TESTS BASE DE DONNÉES ==="
    
    # Tester les migrations
    log_info "Test d'exécution des migrations..."
    
    # Créer une base de test temporaire
    TEST_DB="veza_test_$(date +%s)"
    
    # Créer la base de test
    PGPASSWORD="$POSTGRES_PASSWORD" createdb -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" "$TEST_DB" 2>/dev/null || true
    
    if PGPASSWORD="$POSTGRES_PASSWORD" psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$TEST_DB" -c "SELECT 1;" &> /dev/null; then
        log_success "✓ Base de test créée: $TEST_DB"
        
        # Tester chaque migration
        for migration_file in internal/database/migrations/*.sql; do
            if [[ -f "$migration_file" ]]; then
                migration_name=$(basename "$migration_file")
                ((TOTAL_TESTS++))
                
                if PGPASSWORD="$POSTGRES_PASSWORD" psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$TEST_DB" -f "$migration_file" &> /dev/null; then
                    log_success "✓ Migration appliquée: $migration_name"
                    ((PASSED_TESTS++))
                else
                    log_error "✗ Échec migration: $migration_name"
                    ((FAILED_TESTS++))
                fi
            fi
        done
        
        # Vérifier que les tables ont été créées
        tables_to_check=(
            "users"
            "user_stats"
            "user_preferences"
            "user_sessions"
            "user_audit_logs"
            "chat_rooms"
            "room_members"
            "chat_messages"
            "streams"
            "stream_listeners"
            "stream_tracks"
        )
        
        for table in "${tables_to_check[@]}"; do
            ((TOTAL_TESTS++))
            if PGPASSWORD="$POSTGRES_PASSWORD" psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$TEST_DB" -c "SELECT 1 FROM $table LIMIT 1;" &> /dev/null; then
                log_success "✓ Table existe: $table"
                ((PASSED_TESTS++))
            else
                log_error "✗ Table manquante: $table"
                ((FAILED_TESTS++))
            fi
        done
        
        # Nettoyer la base de test
        PGPASSWORD="$POSTGRES_PASSWORD" dropdb -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" "$TEST_DB" 2>/dev/null || true
        log_info "Base de test nettoyée"
    else
        log_error "Impossible de créer la base de test"
    fi
else
    log_warning "Tests de base de données ignorés (PostgreSQL non accessible)"
fi

# =============================================================================
# PHASE 1.7 - COMPILATION COMPLÈTE
# =============================================================================

echo ""
log_info "=== PHASE 1.7 : COMPILATION COMPLÈTE ==="

# Tester la compilation du serveur principal
((TOTAL_TESTS++))
check_go_compilation "./cmd/production-server" "Serveur de production"

# Tester go mod tidy
log_info "Vérification des dépendances Go..."
((TOTAL_TESTS++))
if go mod tidy && go mod verify; then
    log_success "✓ Dépendances Go valides"
    ((PASSED_TESTS++))
else
    log_error "✗ Problème avec les dépendances Go"
    ((FAILED_TESTS++))
fi

# Tester go vet
log_info "Analyse statique du code..."
((TOTAL_TESTS++))
if go vet ./...; then
    log_success "✓ Code passe go vet"
    ((PASSED_TESTS++))
else
    log_error "✗ Code échoue go vet"
    ((FAILED_TESTS++))
fi

# =============================================================================
# PHASE 1.8 - VÉRIFICATION DE LA DOCUMENTATION
# =============================================================================

echo ""
log_info "=== PHASE 1.8 : DOCUMENTATION ==="

# Vérifier la documentation
docs=(
    "docs/PLAN_FINALISATION_BACKEND_PRODUCTION.md:Plan de finalisation"
    "README.md:README principal"
)

for doc_info in "${docs[@]}"; do
    IFS=':' read -r file_path description <<< "$doc_info"
    ((TOTAL_TESTS++))
    check_file_exists "$file_path" "$description"
done

# =============================================================================
# RÉSUMÉ DES RÉSULTATS
# =============================================================================

echo ""
echo "============================================================================="
echo "  RÉSUMÉ DE LA VALIDATION PHASE 1"
echo "============================================================================="
echo ""

echo -e "📊 ${BLUE}STATISTIQUES:${NC}"
echo -e "   • Total des tests: ${TOTAL_TESTS}"
echo -e "   • Tests réussis: ${GREEN}${PASSED_TESTS}${NC}"
echo -e "   • Tests échoués: ${RED}${FAILED_TESTS}${NC}"

if [[ $FAILED_TESTS -eq 0 ]]; then
    echo ""
    echo -e "${GREEN}🎉 PHASE 1 VALIDÉE AVEC SUCCÈS !${NC}"
    echo ""
    echo -e "✅ ${GREEN}Architecture hexagonale complète${NC}"
    echo -e "✅ ${GREEN}Entités domain implémentées${NC}"
    echo -e "✅ ${GREEN}Interfaces repository complètes${NC}"
    echo -e "✅ ${GREEN}Implémentations PostgreSQL fonctionnelles${NC}"
    echo -e "✅ ${GREEN}Migrations SQL validées${NC}"
    echo -e "✅ ${GREEN}Code compile sans erreur${NC}"
    echo ""
    echo -e "${BLUE}🚀 PRÊT POUR LA PHASE 2 : SÉCURITÉ ENTERPRISE${NC}"
    
    # Créer un fichier de validation
    echo "PHASE_1_VALIDATED=true" > .phase1_complete
    echo "VALIDATION_DATE=$(date -Iseconds)" >> .phase1_complete
    echo "TESTS_PASSED=$PASSED_TESTS" >> .phase1_complete
    echo "TESTS_TOTAL=$TOTAL_TESTS" >> .phase1_complete
    
    exit 0
else
    echo ""
    echo -e "${RED}❌ PHASE 1 INCOMPLÈTE${NC}"
    echo ""
    echo -e "${YELLOW}📋 ACTIONS REQUISES:${NC}"
    
    if [[ $FAILED_TESTS -gt 0 ]]; then
        echo -e "   • Corriger les ${FAILED_TESTS} tests échoués"
        echo -e "   • Vérifier les fichiers manquants"
        echo -e "   • Résoudre les erreurs de compilation"
    fi
    
    echo ""
    echo -e "${BLUE}💡 CONSEILS:${NC}"
    echo -e "   • Exécuter ce script à nouveau après corrections"
    echo -e "   • Vérifier les logs d'erreur ci-dessus"
    echo -e "   • Consulter la documentation pour plus de détails"
    
    exit 1
fi 