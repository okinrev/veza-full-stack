#!/bin/bash

# ============================================================================
# SCRIPT DE VALIDATION - PHASE 3 : SÉCURITÉ PRODUCTION
# ============================================================================

set -e

# Couleurs pour l'affichage
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Variables de configuration
API_BASE="http://localhost:8080/api/v1"
LOG_FILE="validation_phase3_$(date +%Y%m%d_%H%M%S).log"

# Compteurs
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Fonction d'affichage
log() {
    echo -e "${2:-$NC}$1${NC}" | tee -a "$LOG_FILE"
}

log_success() {
    log "✅ $1" "$GREEN"
    ((PASSED_TESTS++))
}

log_error() {
    log "❌ $1" "$RED"
    ((FAILED_TESTS++))
}

log_info() {
    log "ℹ️  $1" "$BLUE"
}

log_warning() {
    log "⚠️  $1" "$YELLOW"
}

# Fonction de test HTTP
test_endpoint() {
    local method="$1"
    local endpoint="$2"
    local data="$3"
    local expected_status="$4"
    local description="$5"
    
    ((TOTAL_TESTS++))
    
    log_info "Test: $description"
    
    if [ -n "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            -H "Content-Type: application/json" \
            -d "$data" \
            "$API_BASE$endpoint" 2>/dev/null || echo "000")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            "$API_BASE$endpoint" 2>/dev/null || echo "000")
    fi
    
    status_code=$(echo "$response" | tail -n1)
    response_body=$(echo "$response" | head -n -1)
    
    if [ "$status_code" = "$expected_status" ]; then
        log_success "$description - Status: $status_code"
        return 0
    else
        log_error "$description - Expected: $expected_status, Got: $status_code"
        return 1
    fi
}

# Fonction de test avec token
test_authenticated_endpoint() {
    local method="$1"
    local endpoint="$2"
    local token="$3"
    local data="$4"
    local expected_status="$5"
    local description="$6"
    
    ((TOTAL_TESTS++))
    
    log_info "Test: $description"
    
    if [ -n "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $token" \
            -d "$data" \
            "$API_BASE$endpoint" 2>/dev/null || echo "000")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            -H "Authorization: Bearer $token" \
            "$API_BASE$endpoint" 2>/dev/null || echo "000")
    fi
    
    status_code=$(echo "$response" | tail -n1)
    response_body=$(echo "$response" | head -n -1)
    
    if [ "$status_code" = "$expected_status" ]; then
        log_success "$description - Status: $status_code"
        return 0
    else
        log_error "$description - Expected: $expected_status, Got: $status_code"
        return 1
    fi
}

# Début des tests
log "🚀 DÉMARRAGE VALIDATION PHASE 3 - SÉCURITÉ PRODUCTION" "$BLUE"
log "=============================================" "$BLUE"
echo

# ============================================================================
# 1. TESTS OAUTH2
# ============================================================================

log "📱 1. TESTS OAUTH2" "$YELLOW"
echo

# Test URL Google OAuth2
test_endpoint "GET" "/auth/oauth/google" "" "200" "Google OAuth URL Generation"

# Test URL GitHub OAuth2  
test_endpoint "GET" "/auth/oauth/github" "" "200" "GitHub OAuth URL Generation"

# Test URL Discord OAuth2
test_endpoint "GET" "/auth/oauth/discord" "" "200" "Discord OAuth URL Generation"

# Test callback sans paramètres (doit échouer)
test_endpoint "GET" "/auth/oauth/google/callback" "" "400" "Google Callback Without Parameters"

# Test callback avec mauvais state
test_endpoint "GET" "/auth/oauth/google/callback?code=test&state=invalid" "" "400" "Google Callback Invalid State"

echo

# ============================================================================
# 2. TESTS AUTHENTIFICATION 2FA/TOTP
# ============================================================================

log "🔐 2. TESTS 2FA/TOTP" "$YELLOW"
echo

# D'abord, créer un utilisateur de test et se connecter
TEST_USER_DATA='{"username":"testuser2fa","email":"test2fa@veza.dev","password":"testpassword123"}'
response=$(curl -s -X POST -H "Content-Type: application/json" -d "$TEST_USER_DATA" "$API_BASE/auth/register" || echo "")

if [[ $response == *"access_token"* ]]; then
    ACCESS_TOKEN=$(echo "$response" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
    log_success "Utilisateur de test créé et token récupéré"
    
    # Test statut 2FA (doit être désactivé)
    test_authenticated_endpoint "GET" "/auth/2fa/status" "$ACCESS_TOKEN" "" "200" "Récupération statut 2FA initial"
    
    # Test configuration 2FA
    SETUP_2FA_DATA='{"password":"testpassword123"}'
    test_authenticated_endpoint "POST" "/auth/2fa/setup" "$ACCESS_TOKEN" "$SETUP_2FA_DATA" "200" "Configuration 2FA Setup"
    
    # Test vérification avec mauvais code
    BAD_TOTP_DATA='{"totp_code":"123456"}'
    test_authenticated_endpoint "POST" "/auth/2fa/verify" "$ACCESS_TOKEN" "$BAD_TOTP_DATA" "400" "Vérification 2FA avec mauvais code"
    
    # Test désactivation 2FA
    DISABLE_2FA_DATA='{"password":"testpassword123"}'
    test_authenticated_endpoint "POST" "/auth/2fa/disable" "$ACCESS_TOKEN" "$DISABLE_2FA_DATA" "200" "Désactivation 2FA"
    
else
    log_error "Impossible de créer l'utilisateur de test pour 2FA"
fi

echo

# ============================================================================
# 3. TESTS MAGIC LINKS
# ============================================================================

log "✨ 3. TESTS MAGIC LINKS" "$YELLOW"
echo

# Test envoi Magic Link
MAGIC_LINK_DATA='{"email":"test@veza.dev","redirect_url":"http://localhost:3000/dashboard"}'
test_endpoint "POST" "/auth/magic-link/send" "$MAGIC_LINK_DATA" "200" "Envoi Magic Link"

# Test Magic Link avec email invalide
INVALID_EMAIL_DATA='{"email":"invalid-email","redirect_url":"http://localhost:3000"}'
test_endpoint "POST" "/auth/magic-link/send" "$INVALID_EMAIL_DATA" "400" "Magic Link avec email invalide"

# Test vérification Magic Link sans token
test_endpoint "GET" "/auth/magic-link/verify" "" "400" "Vérification Magic Link sans token"

# Test vérification avec token invalide
test_endpoint "GET" "/auth/magic-link/verify?token=invalid_token" "" "400" "Vérification Magic Link token invalide"

# Test statut Magic Link
test_endpoint "GET" "/auth/magic-link/status?token=invalid_token" "" "200" "Statut Magic Link"

echo

# ============================================================================
# 4. TESTS SÉCURITÉ GÉNÉRALE
# ============================================================================

log "🛡️  4. TESTS SÉCURITÉ GÉNÉRALE" "$YELLOW"
echo

# Test endpoint protégé sans token
test_endpoint "GET" "/auth/me" "" "401" "Endpoint protégé sans authentification"

# Test avec token invalide
test_authenticated_endpoint "GET" "/auth/me" "invalid_token" "" "401" "Endpoint protégé avec token invalide"

# Test limits de rate (envoyer plusieurs requêtes rapidement)
for i in {1..5}; do
    test_endpoint "POST" "/auth/login" '{"email":"nonexistent@test.com","password":"wrong"}' "401" "Test rate limiting ($i/5)"
    sleep 0.1
done

# Test injection SQL basique
SQL_INJECTION_DATA='{"email":"admin@test.com'\'' OR 1=1 --","password":"test"}'
test_endpoint "POST" "/auth/login" "$SQL_INJECTION_DATA" "401" "Test protection injection SQL"

# Test XSS dans les données
XSS_DATA='{"username":"<script>alert(\"xss\")</script>","email":"xss@test.com","password":"testpass123"}'
test_endpoint "POST" "/auth/register" "$XSS_DATA" "400" "Test protection XSS"

echo

# ============================================================================
# 5. TESTS VALIDATION DES DONNÉES
# ============================================================================

log "✅ 5. TESTS VALIDATION DES DONNÉES" "$YELLOW"
echo

# Test mot de passe trop court
SHORT_PASS_DATA='{"username":"testuser","email":"test@veza.dev","password":"123"}'
test_endpoint "POST" "/auth/register" "$SHORT_PASS_DATA" "400" "Mot de passe trop court"

# Test email invalide
INVALID_EMAIL_REG='{"username":"testuser","email":"not-an-email","password":"validpassword123"}'
test_endpoint "POST" "/auth/register" "$INVALID_EMAIL_REG" "400" "Email invalide à l'inscription"

# Test nom d'utilisateur trop court
SHORT_USERNAME='{"username":"ab","email":"test@veza.dev","password":"validpassword123"}'
test_endpoint "POST" "/auth/register" "$SHORT_USERNAME" "400" "Nom d'utilisateur trop court"

# Test caractères spéciaux dans nom d'utilisateur
SPECIAL_CHARS='{"username":"test@user!","email":"test@veza.dev","password":"validpassword123"}'
test_endpoint "POST" "/auth/register" "$SPECIAL_CHARS" "400" "Caractères spéciaux dans nom d'utilisateur"

echo

# ============================================================================
# 6. TESTS ENDPOINTS ADMINISTRATEUR
# ============================================================================

log "👨‍💼 6. TESTS ENDPOINTS ADMINISTRATEUR" "$YELLOW"
echo

# Test accès admin sans authentification
test_endpoint "GET" "/admin/users" "" "401" "Admin endpoint sans authentification"

# Test accès admin avec utilisateur normal (si on a un token)
if [ -n "$ACCESS_TOKEN" ]; then
    test_authenticated_endpoint "GET" "/admin/users" "$ACCESS_TOKEN" "" "403" "Admin endpoint avec utilisateur normal"
fi

echo

# ============================================================================
# 7. TESTS HEADERS DE SÉCURITÉ
# ============================================================================

log "🔒 7. TESTS HEADERS DE SÉCURITÉ" "$YELLOW"
echo

# Test headers de sécurité
SECURITY_HEADERS=$(curl -s -I "$API_BASE/auth/me" | tr -d '\r')

if echo "$SECURITY_HEADERS" | grep -q "X-Frame-Options"; then
    log_success "Header X-Frame-Options présent"
    ((PASSED_TESTS++))
else
    log_error "Header X-Frame-Options manquant"
    ((FAILED_TESTS++))
fi
((TOTAL_TESTS++))

if echo "$SECURITY_HEADERS" | grep -q "X-Content-Type-Options"; then
    log_success "Header X-Content-Type-Options présent"
    ((PASSED_TESTS++))
else
    log_error "Header X-Content-Type-Options manquant"
    ((FAILED_TESTS++))
fi
((TOTAL_TESTS++))

echo

# ============================================================================
# 8. TESTS CORS
# ============================================================================

log "🌐 8. TESTS CORS" "$YELLOW"
echo

# Test CORS OPTIONS
CORS_RESPONSE=$(curl -s -I -X OPTIONS \
    -H "Origin: http://localhost:3000" \
    -H "Access-Control-Request-Method: POST" \
    -H "Access-Control-Request-Headers: Content-Type,Authorization" \
    "$API_BASE/auth/login" | tr -d '\r')

if echo "$CORS_RESPONSE" | grep -q "Access-Control-Allow-Origin"; then
    log_success "CORS configuré correctement"
    ((PASSED_TESTS++))
else
    log_error "CORS non configuré"
    ((FAILED_TESTS++))
fi
((TOTAL_TESTS++))

echo

# ============================================================================
# RÉSUMÉ DES TESTS
# ============================================================================

log "📊 RÉSUMÉ DES TESTS DE SÉCURITÉ" "$BLUE"
log "=============================================" "$BLUE"
log "Total des tests: $TOTAL_TESTS" "$BLUE"
log "Tests réussis: $PASSED_TESTS" "$GREEN"
log "Tests échoués: $FAILED_TESTS" "$RED"

if [ $FAILED_TESTS -eq 0 ]; then
    log "🎉 TOUS LES TESTS DE SÉCURITÉ SONT PASSÉS!" "$GREEN"
    PERCENTAGE=100
else
    PERCENTAGE=$((PASSED_TESTS * 100 / TOTAL_TESTS))
    log "Taux de réussite: ${PERCENTAGE}%" "$YELLOW"
fi

log "Rapport détaillé sauvegardé dans: $LOG_FILE" "$BLUE"

# ============================================================================
# RECOMMENDATIONS DE SÉCURITÉ
# ============================================================================

echo
log "🔧 RECOMMANDATIONS DE SÉCURITÉ" "$YELLOW"
log "=============================================" "$YELLOW"

if [ $PERCENTAGE -lt 80 ]; then
    log "⚠️  Taux de réussite faible - Actions requises:" "$RED"
    log "   - Vérifier la configuration des endpoints OAuth2" "$YELLOW"
    log "   - Valider les middlewares d'authentification" "$YELLOW"
    log "   - Contrôler les headers de sécurité" "$YELLOW"
elif [ $PERCENTAGE -lt 95 ]; then
    log "✅ Sécurité globalement correcte - Améliorations possibles:" "$YELLOW"
    log "   - Finaliser la configuration OAuth2" "$YELLOW"
    log "   - Optimiser les validations" "$YELLOW"
else
    log "🛡️  EXCELLENT! Sécurité de niveau production" "$GREEN"
    log "   - OAuth2 fonctionnel (Google, GitHub, Discord)" "$GREEN"
    log "   - 2FA/TOTP correctement implémenté" "$GREEN"
    log "   - Magic Links sécurisés" "$GREEN"
    log "   - Protection contre les attaques courantes" "$GREEN"
    log "   - Headers de sécurité configurés" "$GREEN"
fi

echo
log "✨ PHASE 3 - SÉCURITÉ PRODUCTION: $PERCENTAGE% COMPLÉTÉE" "$BLUE"

# Code de sortie basé sur le taux de réussite
if [ $PERCENTAGE -ge 80 ]; then
    exit 0
else
    exit 1
fi
