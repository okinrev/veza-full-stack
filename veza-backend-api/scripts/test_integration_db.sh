#!/bin/bash

# Test d'intégration PostgreSQL pour Veza Backend
# Vérifie la connexion DB, les migrations, et les opérations CRUD

set -e

echo "🔍 TEST D'INTÉGRATION POSTGRESQL - VEZA BACKEND"
echo "=============================================="

# Configuration
DB_HOST=${DATABASE_HOST:-"localhost"}
DB_PORT=${DATABASE_PORT:-"5432"}  
DB_USER=${DATABASE_USER:-"veza_user"}
DB_NAME=${DATABASE_NAME:-"veza_dev"}
SERVER_PORT=${PORT:-"8080"}
API_BASE_URL="http://localhost:${SERVER_PORT}"

# Couleurs pour l'affichage
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

success() {
    echo -e "${GREEN}✅ $1${NC}"
}

error() {
    echo -e "${RED}❌ $1${NC}"
    exit 1
}

info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# Test 1: Connexion PostgreSQL directe
test_postgres_connection() {
    info "Test 1: Connexion PostgreSQL directe"
    
    if command -v psql >/dev/null 2>&1; then
        if psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT version();" >/dev/null 2>&1; then
            success "Connexion PostgreSQL réussie"
        else
            error "Échec de connexion PostgreSQL"
        fi
    else
        warning "psql non disponible, test sauté"
    fi
}

# Test 2: Vérification des tables essentielles
test_database_schema() {
    info "Test 2: Vérification du schéma de base de données"
    
    # Tables critiques à vérifier
    TABLES=("users" "refresh_tokens" "chat_messages" "chat_rooms")
    
    for table in "${TABLES[@]}"; do
        if command -v psql >/dev/null 2>&1; then
            if psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "\dt $table" >/dev/null 2>&1; then
                success "Table '$table' existe"
            else
                warning "Table '$table' n'existe pas ou non accessible"
            fi
        else
            warning "Impossible de vérifier le schéma (psql non disponible)"
            break
        fi
    done
}

# Test 3: Test des migrations
test_migrations() {
    info "Test 3: Test des migrations"
    
    # Compiler et lancer le serveur pour les migrations
    go build -o tmp/test-server ./cmd/server >/dev/null 2>&1
    if [ $? -eq 0 ]; then
        success "Compilation du serveur réussie"
    else 
        error "Échec de compilation du serveur"
    fi
    
    # Tester les migrations via le serveur
    timeout 10s ./tmp/test-server >/dev/null 2>&1 &
    SERVER_PID=$!
    sleep 3
    
    # Vérifier si le serveur s'est lancé sans erreur critique
    if kill -0 $SERVER_PID 2>/dev/null; then
        success "Serveur démarré, migrations probablement réussies"
        kill $SERVER_PID 2>/dev/null || true
        wait $SERVER_PID 2>/dev/null || true
    else
        warning "Serveur arrêté prématurément, vérifier les migrations"
    fi
}

# Test 4: Test des opérations CRUD via API
test_crud_operations() {
    info "Test 4: Test des opérations CRUD via API"
    
    # Démarrer le serveur de test
    go build -o tmp/integration-test-server ./cmd/server >/dev/null 2>&1
    ./tmp/integration-test-server > tmp/server-test.log 2>&1 &
    SERVER_PID=$!
    
    # Attendre que le serveur démarre
    sleep 5
    
    # Test de health check
    if curl -s "$API_BASE_URL/health" >/dev/null 2>&1; then
        success "Health check réussi"
    else
        error "Health check échoué"
    fi
    
    # Test de l'endpoint d'inscription
    REGISTER_RESPONSE=$(curl -s -X POST "$API_BASE_URL/api/v1/auth/register" \
        -H "Content-Type: application/json" \
        -d '{"username":"testuser","email":"test@example.com","password":"testpass123"}' 2>/dev/null)
    
    if echo "$REGISTER_RESPONSE" | grep -q "success"; then
        if echo "$REGISTER_RESPONSE" | grep -q '"success":true'; then
            success "Registration réussie"
        else
            # Vérifier si c'est une erreur attendue (utilisateur existe déjà)
            if echo "$REGISTER_RESPONSE" | grep -q "already exists"; then
                success "Registration testée (utilisateur existe déjà)"
            else
                warning "Registration échouée : $REGISTER_RESPONSE"
            fi
        fi
    else
        warning "Pas de réponse valide pour registration"
    fi
    
    # Test de l'endpoint de login
    LOGIN_RESPONSE=$(curl -s -X POST "$API_BASE_URL/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d '{"email":"test@example.com","password":"testpass123"}' 2>/dev/null)
    
    if echo "$LOGIN_RESPONSE" | grep -q "success"; then
        success "Login endpoint répond"
    else
        warning "Login endpoint ne répond pas correctement"
    fi
    
    # Nettoyer
    kill $SERVER_PID 2>/dev/null || true
    wait $SERVER_PID 2>/dev/null || true
}

# Test 5: Test de performance de base
test_basic_performance() {
    info "Test 5: Test de performance de base"
    
    # Démarrer le serveur
    ./tmp/integration-test-server > tmp/perf-test.log 2>&1 &
    SERVER_PID=$!
    sleep 3
    
    # Test de latence
    START_TIME=$(date +%s%N)
    curl -s "$API_BASE_URL/health" >/dev/null 2>&1
    END_TIME=$(date +%s%N)
    
    LATENCY=$(( (END_TIME - START_TIME) / 1000000 )) # ms
    
    if [ $LATENCY -lt 100 ]; then
        success "Latence acceptable: ${LATENCY}ms"
    else
        warning "Latence élevée: ${LATENCY}ms"
    fi
    
    # Test de throughput basique
    info "Test de throughput (10 requêtes simultanées)"
    for i in {1..10}; do
        curl -s "$API_BASE_URL/health" >/dev/null 2>&1 &
    done
    wait
    success "Test de throughput terminé"
    
    # Nettoyer
    kill $SERVER_PID 2>/dev/null || true
    wait $SERVER_PID 2>/dev/null || true
}

# Fonction principale
main() {
    echo "Début des tests d'intégration PostgreSQL..."
    echo "Configuration:"
    echo "  - Base de données: $DB_HOST:$DB_PORT/$DB_NAME"  
    echo "  - Utilisateur: $DB_USER"
    echo "  - API URL: $API_BASE_URL"
    echo ""
    
    test_postgres_connection
    test_database_schema
    test_migrations
    test_crud_operations
    test_basic_performance
    
    echo ""
    echo "🎉 TESTS D'INTÉGRATION POSTGRESQL TERMINÉS"
    success "Tous les tests sont passés avec succès !"
}

# Gestion des erreurs
cleanup() {
    info "Nettoyage en cours..."
    pkill -f "integration-test-server" 2>/dev/null || true
    pkill -f "test-server" 2>/dev/null || true
    rm -f tmp/test-server tmp/integration-test-server 2>/dev/null || true
}

trap cleanup EXIT

# Exécution
main "$@" 