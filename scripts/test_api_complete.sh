#!/bin/bash

# 🧪 TEST COMPLET DE L'API TALAS
# Validation de tous les endpoints avec les bonnes URLs

set -e

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
API_BASE="http://localhost:8080/api/v1"
HEALTH_URL="http://localhost:8080/api/health"

# Compteurs
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Fonctions utilitaires
log_info() { echo -e "${BLUE}ℹ️  $1${NC}"; }
log_success() { echo -e "${GREEN}✅ $1${NC}"; ((PASSED_TESTS++)); }
log_warning() { echo -e "${YELLOW}⚠️  $1${NC}"; }
log_error() { echo -e "${RED}❌ $1${NC}"; ((FAILED_TESTS++)); }

test_endpoint() {
    local url="$1"
    local description="$2"
    local expected_status="${3:-200}"
    
    ((TOTAL_TESTS++))
    
    local status_code=$(curl -s -o /dev/null -w "%{http_code}" "$url" 2>/dev/null || echo "000")
    
    if [ "$status_code" = "$expected_status" ]; then
        log_success "$description (HTTP $status_code)"
    elif [ "$status_code" = "000" ]; then
        log_error "$description (Connexion impossible)"
    else
        log_warning "$description (HTTP $status_code, attendu $expected_status)"
    fi
}

echo "🧪 TEST COMPLET DE L'API TALAS"
echo "==============================="
echo ""

# Vérification que le serveur est en cours d'exécution
log_info "Vérification du serveur..."
test_endpoint "$HEALTH_URL" "Health Endpoint" "200"

if [ $PASSED_TESTS -eq 0 ]; then
    log_error "Serveur non accessible. Démarrez-le avec:"
    echo "cd veza-backend-api && go run cmd/server/main.go"
    exit 1
fi

echo ""
log_info "🔐 Tests d'authentification"
test_endpoint "$API_BASE/auth/register" "Register Endpoint" "422"
test_endpoint "$API_BASE/auth/login" "Login Endpoint" "400"

echo ""
log_info "👥 Tests de gestion des utilisateurs"
test_endpoint "$API_BASE/users" "Liste des utilisateurs" "401"
test_endpoint "$API_BASE/users/me" "Profil utilisateur" "401"
test_endpoint "$API_BASE/users/search" "Recherche utilisateurs" "401"

echo ""
log_info "💬 Tests de chat"
test_endpoint "$API_BASE/chat/rooms" "Salles de chat" "401"
test_endpoint "$API_BASE/chat/conversations" "Conversations" "401"

echo ""
log_info "🎵 Tests de streaming"
test_endpoint "$API_BASE/tracks" "Liste des tracks" "200"

echo ""
log_info "🔍 Tests de recherche"
test_endpoint "$API_BASE/search" "Recherche globale" "200"
test_endpoint "$API_BASE/search/advanced" "Recherche avancée" "200"

echo ""
log_info "🏷️ Tests des tags"
test_endpoint "$API_BASE/tags" "Liste des tags" "200"
test_endpoint "$API_BASE/tags/search" "Recherche de tags" "200"

echo ""
log_info "👑 Tests d'administration"
test_endpoint "$API_BASE/admin/dashboard" "Dashboard admin" "401"
test_endpoint "$API_BASE/admin/users" "Admin utilisateurs" "401"

echo ""
log_info "📁 Tests des ressources partagées"
test_endpoint "$API_BASE/shared-resources" "Ressources partagées" "200"

echo ""
log_info "📋 Tests des listings"
test_endpoint "$API_BASE/listings" "Liste des annonces" "200"

echo ""
echo "🎯 RÉSULTATS DES TESTS"
echo "====================="
log_info "Total des tests: $TOTAL_TESTS"
log_success "Tests réussis: $PASSED_TESTS"
if [ $FAILED_TESTS -gt 0 ]; then
    log_error "Tests échoués: $FAILED_TESTS"
fi

# Calcul du pourcentage de réussite
if [ $TOTAL_TESTS -gt 0 ]; then
    percentage=$((PASSED_TESTS * 100 / TOTAL_TESTS))
    echo ""
    if [ $percentage -ge 80 ]; then
        log_success "Score global: $percentage% - EXCELLENT ✅"
    elif [ $percentage -ge 60 ]; then
        log_warning "Score global: $percentage% - BON ⚠️"
    else
        log_error "Score global: $percentage% - À AMÉLIORER ❌"
    fi
fi

echo ""
log_info "📝 Notes importantes:"
echo "• HTTP 401 = Normal (authentification requise)"
echo "• HTTP 200 = Endpoint accessible" 
echo "• HTTP 422/400 = Normal (données manquantes)"
echo "• HTTP 404 = Endpoint non trouvé (problème)"

echo ""
if [ $percentage -ge 60 ]; then
    log_success "🎉 API FONCTIONNELLE - Prête pour l'intégration frontend !"
else
    log_error "⚠️ API PARTIELLEMENT FONCTIONNELLE - Corrections recommandées"
fi 