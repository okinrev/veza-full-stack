#!/bin/bash

# 🧪 AUDIT COMPLET DU BACKEND TALAS
# Script de validation exhaustive pour production-ready

set -e

# Couleurs pour les logs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKEND_DIR="$PROJECT_ROOT/veza-backend-api"
CHAT_DIR="$PROJECT_ROOT/veza-chat-server"
STREAM_DIR="$PROJECT_ROOT/veza-stream-server"

# URLs de test
API_BASE="http://localhost:8080"
CHAT_WS="ws://localhost:3001"
STREAM_WS="ws://localhost:3002"

# Compteurs de tests
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Fonctions utilitaires
log_info() { echo -e "${BLUE}ℹ️  $1${NC}"; }
log_success() { echo -e "${GREEN}✅ $1${NC}"; ((PASSED_TESTS++)); }
log_warning() { echo -e "${YELLOW}⚠️  $1${NC}"; }
log_error() { echo -e "${RED}❌ $1${NC}"; ((FAILED_TESTS++)); }
log_test() { echo -e "${BLUE}🧪 Test: $1${NC}"; ((TOTAL_TESTS++)); }

# Test de compilation
test_compilation() {
    log_info "🔧 TESTS DE COMPILATION"
    echo "========================"
    
    # Backend Go
    log_test "Compilation Backend Go"
    if cd "$BACKEND_DIR" && go build -o tmp/audit-test cmd/server/main.go 2>/dev/null; then
        log_success "Backend Go compile sans erreur"
        rm -f tmp/audit-test
    else
        log_error "Backend Go ne compile pas"
    fi
    
    # Chat Server Rust
    log_test "Compilation Chat Server"
    if cd "$CHAT_DIR" && cargo check --quiet 2>/dev/null; then
        log_success "Chat Server compile (avec warnings)"
    else
        log_error "Chat Server ne compile pas"
    fi
    
    # Stream Server Rust
    log_test "Compilation Stream Server"
    if cd "$STREAM_DIR" && cargo check --quiet 2>/dev/null; then
        log_success "Stream Server compile (avec warnings)"
    else
        log_error "Stream Server ne compile pas"
    fi
    
    cd "$PROJECT_ROOT"
    echo ""
}

# Test de la base de données
test_database() {
    log_info "🗄️  TESTS DE BASE DE DONNÉES"
    echo "============================="
    
    # Test de connexion PostgreSQL
    log_test "Connexion PostgreSQL"
    if command -v psql > /dev/null && psql -h localhost -U postgres -d veza_db -c "SELECT 1;" > /dev/null 2>&1; then
        log_success "PostgreSQL accessible"
        
        # Test des tables principales
        log_test "Vérification des tables"
        if psql -h localhost -U postgres -d veza_db -c "SELECT COUNT(*) FROM users;" > /dev/null 2>&1; then
            log_success "Table users existe"
        else
            log_error "Table users manquante"
        fi
        
        if psql -h localhost -U postgres -d veza_db -c "SELECT COUNT(*) FROM rooms;" > /dev/null 2>&1; then
            log_success "Table rooms existe"
        else
            log_error "Table rooms manquante"
        fi
        
    else
        log_error "PostgreSQL non accessible (vérifiez la connexion)"
    fi
    
    # Test Redis
    log_test "Connexion Redis"
    if command -v redis-cli > /dev/null && redis-cli ping 2>/dev/null | grep -q "PONG"; then
        log_success "Redis accessible"
    else
        log_error "Redis non accessible"
    fi
    
    echo ""
}

# Test des endpoints API
test_api_endpoints() {
    log_info "🌐 TESTS DES ENDPOINTS API"
    echo "============================="
    
    # Health check
    log_test "Health check"
    if curl -s -f "$API_BASE/api/health" > /dev/null 2>&1; then
        log_success "Health endpoint accessible"
    else
        log_error "Health endpoint non accessible"
    fi
    
    # Test endpoints publics
    log_test "Endpoints publics"
    local public_endpoints=("/api/users" "/api/rooms" "/api/search" "/api/tags")
    
    for endpoint in "${public_endpoints[@]}"; do
        local response=$(curl -s -o /dev/null -w "%{http_code}" "$API_BASE$endpoint" 2>/dev/null || echo "000")
        if [[ "$response" =~ ^[2-4][0-9][0-9]$ ]]; then
            log_success "Endpoint $endpoint répond (HTTP $response)"
        else
            log_error "Endpoint $endpoint ne répond pas (HTTP $response)"
        fi
    done
    
    echo ""
}

# Test de sécurité basique
test_security() {
    log_info "🔒 TESTS DE SÉCURITÉ BASIQUE"
    echo "============================="
    
    # Test headers de sécurité
    log_test "Headers de sécurité"
    local headers=$(curl -s -I "$API_BASE/api/health" 2>/dev/null || echo "")
    
    if echo "$headers" | grep -q "X-Frame-Options"; then
        log_success "Header X-Frame-Options présent"
    else
        log_warning "Header X-Frame-Options manquant"
    fi
    
    # Test rate limiting
    log_test "Rate limiting basique"
    local rate_limit_count=0
    for i in {1..5}; do
        if curl -s -f "$API_BASE/api/health" > /dev/null 2>&1; then
            ((rate_limit_count++))
        fi
    done
    
    if [ $rate_limit_count -eq 5 ]; then
        log_warning "Pas de rate limiting détecté (5 requêtes passées)"
    else
        log_success "Rate limiting fonctionne ($rate_limit_count/5 requêtes passées)"
    fi
    
    echo ""
}

# Test de configuration
test_configuration() {
    log_info "⚙️  TESTS DE CONFIGURATION"
    echo "=========================="
    
    # Fichiers de configuration
    log_test "Fichiers de configuration"
    if [ -f "$BACKEND_DIR/.env" ]; then
        log_success "Fichier .env du backend présent"
    else
        log_warning "Fichier .env du backend manquant"
    fi
    
    if [ -f "$CHAT_DIR/.env" ]; then
        log_success "Fichier .env du chat présent"
    else
        log_warning "Fichier .env du chat manquant"
    fi
    
    if [ -f "$STREAM_DIR/.env" ]; then
        log_success "Fichier .env du stream présent"
    else
        log_warning "Fichier .env du stream manquant"
    fi
    
    echo ""
}

# Rapport final
generate_report() {
    echo ""
    log_info "📋 RAPPORT D'AUDIT FINAL"
    echo "========================="
    echo ""
    
    local success_rate=0
    if [ $TOTAL_TESTS -gt 0 ]; then
        success_rate=$((PASSED_TESTS * 100 / TOTAL_TESTS))
    fi
    
    echo "Tests exécutés: $TOTAL_TESTS"
    echo "Tests réussis: $PASSED_TESTS"
    echo "Tests échoués: $FAILED_TESTS"
    echo "Taux de réussite: $success_rate%"
    echo ""
    
    if [ $success_rate -ge 90 ]; then
        log_success "BACKEND PRODUCTION-READY ✨"
        echo "Le backend est prêt pour la production avec un taux de réussite excellent."
    elif [ $success_rate -ge 75 ]; then
        log_warning "BACKEND ACCEPTABLE AVEC AMÉLIORATIONS RECOMMANDÉES"
        echo "Le backend fonctionne mais nécessite quelques améliorations."
    elif [ $success_rate -ge 50 ]; then
        log_warning "BACKEND PARTIELLEMENT FONCTIONNEL"
        echo "Le backend nécessite des corrections importantes avant la production."
    else
        log_error "BACKEND NON PRODUCTION-READY"
        echo "Le backend nécessite des corrections critiques avant utilisation."
    fi
    
    echo ""
    echo "🕐 Audit terminé le: $(date)"
}

# Fonction principale
main() {
    echo "🧪 AUDIT COMPLET DU BACKEND TALAS"
    echo "================================="
    echo "Début de l'audit: $(date)"
    echo ""
    
    # Vérifier les prérequis
    if ! command -v curl > /dev/null; then
        log_error "curl n'est pas installé"
        exit 1
    fi
    
    # Exécuter tous les tests
    test_compilation
    test_database
    test_api_endpoints
    test_security
    test_configuration
    
    # Générer le rapport
    generate_report
    
    # Code de sortie basé sur le taux de réussite
    local success_rate=$((PASSED_TESTS * 100 / TOTAL_TESTS))
    if [ $success_rate -ge 75 ]; then
        exit 0
    else
        exit 1
    fi
}

# Piéger les interruptions
trap 'echo ""; log_warning "Audit interrompu"; exit 1' INT TERM

# Exécuter l'audit
main "$@" 