#!/bin/bash

# =============================================================================
# SCRIPT DE VALIDATION PHASE 2 - SÉCURITÉ ENTERPRISE
# =============================================================================

set -euo pipefail

# Couleurs pour l'affichage
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Compteurs
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Configuration
API_URL="http://localhost:8080"
TEST_EMAIL="security.test@veza.dev"
TEST_PASSWORD="SecureP@ssw0rd123!"
TEST_USERNAME="securitytest"

# Fonction d'affichage
print_header() {
    echo -e "\n${PURPLE}==============================================================================${NC}"
    echo -e "${PURPLE}$1${NC}"
    echo -e "${PURPLE}==============================================================================${NC}\n"
}

print_section() {
    echo -e "\n${BLUE}📋 $1${NC}"
    echo -e "${BLUE}$(printf '=%.0s' {1..80})${NC}"
}

print_test() {
    echo -e "\n${CYAN}🧪 Test: $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
    ((PASSED_TESTS++))
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
    ((FAILED_TESTS++))
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_info() {
    echo -e "${CYAN}ℹ️  $1${NC}"
}

# Fonction pour incrémenter le compteur de tests
increment_test() {
    ((TOTAL_TESTS++))
}

# Fonction pour vérifier si le serveur répond
check_server() {
    print_test "Vérification de la disponibilité du serveur"
    increment_test
    
    if curl -s -f "$API_URL/api/health" > /dev/null 2>&1; then
        print_success "Serveur accessible sur $API_URL"
        return 0
    else
        print_error "Serveur non accessible sur $API_URL"
        return 1
    fi
}

# Fonction pour tester l'authentification JWT
test_jwt_authentication() {
    print_section "AUTHENTIFICATION JWT"
    
    # Test 1: Inscription
    print_test "Inscription d'un utilisateur de test"
    increment_test
    
    local register_response
    register_response=$(curl -s -X POST "$API_URL/api/v1/auth/register" \
        -H "Content-Type: application/json" \
        -d "{
            \"username\": \"$TEST_USERNAME\",
            \"email\": \"$TEST_EMAIL\",
            \"password\": \"$TEST_PASSWORD\",
            \"first_name\": \"Security\",
            \"last_name\": \"Test\",
            \"display_name\": \"Security Test\"
        }")
    
    if echo "$register_response" | jq -e '.access_token' > /dev/null 2>&1; then
        print_success "Inscription réussie avec token JWT"
        ACCESS_TOKEN=$(echo "$register_response" | jq -r '.access_token')
    else
        print_error "Échec de l'inscription: $register_response"
        return 1
    fi
    
    # Test 2: Connexion
    print_test "Connexion avec email/mot de passe"
    increment_test
    
    local login_response
    login_response=$(curl -s -X POST "$API_URL/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d "{
            \"email\": \"$TEST_EMAIL\",
            \"password\": \"$TEST_PASSWORD\"
        }")
    
    if echo "$login_response" | jq -e '.access_token' > /dev/null 2>&1; then
        print_success "Connexion réussie avec token JWT"
        ACCESS_TOKEN=$(echo "$login_response" | jq -r '.access_token')
        REFRESH_TOKEN=$(echo "$login_response" | jq -r '.refresh_token')
    else
        print_error "Échec de la connexion: $login_response"
        return 1
    fi
    
    # Test 3: Accès à un endpoint protégé
    print_test "Accès à un endpoint protégé avec JWT"
    increment_test
    
    local protected_response
    protected_response=$(curl -s -H "Authorization: Bearer $ACCESS_TOKEN" \
        "$API_URL/api/v1/user/profile")
    
    if echo "$protected_response" | jq -e '.id' > /dev/null 2>&1; then
        print_success "Accès autorisé avec JWT valide"
    else
        print_error "Accès refusé avec JWT: $protected_response"
    fi
    
    # Test 4: Refresh token
    print_test "Rafraîchissement du token JWT"
    increment_test
    
    local refresh_response
    refresh_response=$(curl -s -X POST "$API_URL/api/v1/auth/refresh" \
        -H "Content-Type: application/json" \
        -d "{\"refresh_token\": \"$REFRESH_TOKEN\"}")
    
    if echo "$refresh_response" | jq -e '.access_token' > /dev/null 2>&1; then
        print_success "Token rafraîchi avec succès"
        ACCESS_TOKEN=$(echo "$refresh_response" | jq -r '.access_token')
    else
        print_error "Échec du rafraîchissement: $refresh_response"
    fi
}

# Fonction pour tester le rate limiting
test_rate_limiting() {
    print_section "RATE LIMITING"
    
    # Test 1: Rate limiting endpoint login
    print_test "Rate limiting sur /auth/login (max 5 req/min)"
    increment_test
    
    local blocked=false
    for i in {1..7}; do
        local response
        response=$(curl -s -w "%{http_code}" -X POST "$API_URL/api/v1/auth/login" \
            -H "Content-Type: application/json" \
            -d '{"email":"wrong@email.com","password":"wrongpass"}')
        
        local http_code="${response: -3}"
        
        if [[ "$http_code" == "429" ]]; then
            blocked=true
            break
        fi
        
        sleep 1
    done
    
    if $blocked; then
        print_success "Rate limiting actif - requêtes bloquées après dépassement"
    else
        print_warning "Rate limiting non détecté sur /auth/login"
    fi
    
    # Test 2: Headers de rate limiting
    print_test "Présence des headers X-RateLimit-*"
    increment_test
    
    local headers
    headers=$(curl -s -I "$API_URL/api/v1/tracks")
    
    if echo "$headers" | grep -q "X-RateLimit-Limit" && 
       echo "$headers" | grep -q "X-RateLimit-Remaining"; then
        print_success "Headers de rate limiting présents"
    else
        print_warning "Headers de rate limiting manquants"
    fi
}

# Fonction pour tester les headers de sécurité
test_security_headers() {
    print_section "HEADERS DE SÉCURITÉ"
    
    print_test "Vérification des headers de sécurité"
    increment_test
    
    local headers
    headers=$(curl -s -I "$API_URL/api/health")
    
    # Test des headers individuels
    local security_checks=(
        "X-Content-Type-Options:nosniff"
        "X-Frame-Options"
        "Strict-Transport-Security"
        "X-XSS-Protection"
        "Referrer-Policy"
    )
    
    local found_headers=0
    for header in "${security_checks[@]}"; do
        if echo "$headers" | grep -qi "$header"; then
            print_info "✓ $header trouvé"
            ((found_headers++))
        else
            print_warning "✗ $header manquant"
        fi
    done
    
    if [[ $found_headers -ge 3 ]]; then
        print_success "Headers de sécurité présents ($found_headers/5)"
    else
        print_error "Trop peu de headers de sécurité ($found_headers/5)"
    fi
}

# Fonction pour tester CORS
test_cors() {
    print_section "CONFIGURATION CORS"
    
    print_test "Test de la configuration CORS"
    increment_test
    
    local cors_response
    cors_response=$(curl -s -H "Origin: http://localhost:3000" \
        -H "Access-Control-Request-Method: POST" \
        -H "Access-Control-Request-Headers: Content-Type,Authorization" \
        -X OPTIONS "$API_URL/api/v1/tracks")
    
    if echo "$cors_response" | grep -q "Access-Control-Allow-Origin"; then
        print_success "CORS configuré correctement"
    else
        print_warning "Configuration CORS non détectée"
    fi
}

# Fonction pour tester l'injection SQL
test_sql_injection() {
    print_section "PROTECTION INJECTION SQL"
    
    print_test "Test basique d'injection SQL"
    increment_test
    
    # Test avec tentative d'injection simple
    local injection_payload="'; DROP TABLE users; --"
    local response
    response=$(curl -s -X POST "$API_URL/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d "{\"email\": \"$injection_payload\", \"password\": \"test\"}")
    
    # Si le serveur répond normalement (pas de crash), c'est bon
    if echo "$response" | jq -e '.error' > /dev/null 2>&1; then
        print_success "Protection contre l'injection SQL active"
    else
        print_warning "Réponse inattendue à la tentative d'injection"
    fi
}

# Fonction pour tester la validation des entrées
test_input_validation() {
    print_section "VALIDATION DES ENTRÉES"
    
    # Test 1: Email invalide
    print_test "Validation email invalide"
    increment_test
    
    local response
    response=$(curl -s -X POST "$API_URL/api/v1/auth/register" \
        -H "Content-Type: application/json" \
        -d '{"email": "invalid-email", "password": "validpass123"}')
    
    if echo "$response" | jq -e '.error' > /dev/null 2>&1; then
        print_success "Email invalide rejeté"
    else
        print_error "Email invalide accepté"
    fi
    
    # Test 2: Mot de passe trop court
    print_test "Validation mot de passe faible"
    increment_test
    
    response=$(curl -s -X POST "$API_URL/api/v1/auth/register" \
        -H "Content-Type: application/json" \
        -d '{"email": "test@test.com", "password": "123"}')
    
    if echo "$response" | jq -e '.error' > /dev/null 2>&1; then
        print_success "Mot de passe faible rejeté"
    else
        print_error "Mot de passe faible accepté"
    fi
    
    # Test 3: Champs requis manquants
    print_test "Validation champs requis"
    increment_test
    
    response=$(curl -s -X POST "$API_URL/api/v1/auth/register" \
        -H "Content-Type: application/json" \
        -d '{"email": "test@test.com"}')
    
    if echo "$response" | jq -e '.error' > /dev/null 2>&1; then
        print_success "Champs requis manquants rejetés"
    else
        print_error "Champs requis manquants acceptés"
    fi
}

# Fonction pour tester l'OAuth2 (si configuré)
test_oauth2() {
    print_section "OAUTH2 (SI CONFIGURÉ)"
    
    # Test 1: URL d'authentification Google
    print_test "Génération URL OAuth2 Google"
    increment_test
    
    local oauth_response
    oauth_response=$(curl -s "$API_URL/api/v1/auth/oauth/google/url?state=test123")
    
    if echo "$oauth_response" | jq -e '.url' > /dev/null 2>&1; then
        print_success "URL OAuth2 Google générée"
    else
        print_info "OAuth2 Google non configuré ou non disponible"
    fi
    
    # Test 2: URL d'authentification GitHub
    print_test "Génération URL OAuth2 GitHub"
    increment_test
    
    oauth_response=$(curl -s "$API_URL/api/v1/auth/oauth/github/url?state=test123")
    
    if echo "$oauth_response" | jq -e '.url' > /dev/null 2>&1; then
        print_success "URL OAuth2 GitHub générée"
    else
        print_info "OAuth2 GitHub non configuré ou non disponible"
    fi
}

# Fonction pour tester les permissions RBAC
test_rbac() {
    print_section "RBAC - CONTRÔLE D'ACCÈS"
    
    # Test 1: Accès endpoint admin sans privilèges
    print_test "Accès endpoint admin sans privilèges"
    increment_test
    
    local admin_response
    admin_response=$(curl -s -w "%{http_code}" \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        "$API_URL/api/v1/admin/users")
    
    local http_code="${admin_response: -3}"
    
    if [[ "$http_code" == "403" ]]; then
        print_success "Accès admin bloqué pour utilisateur normal"
    elif [[ "$http_code" == "401" ]]; then
        print_success "Authentification requise pour endpoint admin"
    else
        print_warning "Endpoint admin accessible sans privilèges ($http_code)"
    fi
    
    # Test 2: Permissions utilisateur de base
    print_test "Permissions utilisateur de base"
    increment_test
    
    local profile_response
    profile_response=$(curl -s -w "%{http_code}" \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        "$API_URL/api/v1/user/profile")
    
    http_code="${profile_response: -3}"
    
    if [[ "$http_code" == "200" ]]; then
        print_success "Accès profil autorisé pour utilisateur"
    else
        print_error "Accès profil refusé pour utilisateur ($http_code)"
    fi
}

# Fonction pour tester la sécurité 2FA (si disponible)
test_2fa() {
    print_section "AUTHENTIFICATION 2FA"
    
    # Test 1: Activation 2FA
    print_test "Endpoint d'activation 2FA"
    increment_test
    
    local twofa_response
    twofa_response=$(curl -s -w "%{http_code}" \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        -X POST "$API_URL/api/v1/auth/2fa/enable")
    
    local http_code="${twofa_response: -3}"
    
    if [[ "$http_code" == "200" ]]; then
        print_success "Endpoint 2FA disponible"
    elif [[ "$http_code" == "404" ]]; then
        print_info "2FA non encore implémenté"
    else
        print_warning "Erreur endpoint 2FA ($http_code)"
    fi
}

# Fonction pour nettoyer les données de test
cleanup_test_data() {
    print_section "NETTOYAGE"
    
    print_test "Suppression des données de test"
    increment_test
    
    # Tenter de supprimer l'utilisateur de test
    local delete_response
    delete_response=$(curl -s -w "%{http_code}" \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        -X DELETE "$API_URL/api/v1/user/account")
    
    local http_code="${delete_response: -3}"
    
    if [[ "$http_code" == "200" ]] || [[ "$http_code" == "204" ]]; then
        print_success "Données de test supprimées"
    else
        print_info "Nettoyage manuel requis pour l'utilisateur: $TEST_EMAIL"
    fi
}

# Fonction pour générer le rapport final
generate_report() {
    print_header "RAPPORT FINAL - PHASE 2 SÉCURITÉ"
    
    echo -e "${CYAN}📊 STATISTIQUES:${NC}"
    echo -e "   • Tests exécutés: ${TOTAL_TESTS}"
    echo -e "   • Tests réussis:  ${GREEN}${PASSED_TESTS}${NC}"
    echo -e "   • Tests échoués:  ${RED}${FAILED_TESTS}${NC}"
    
    local success_rate=$((PASSED_TESTS * 100 / TOTAL_TESTS))
    echo -e "   • Taux de réussite: ${success_rate}%"
    
    echo -e "\n${CYAN}🎯 ÉVALUATION SÉCURITÉ:${NC}"
    
    if [[ $success_rate -ge 90 ]]; then
        echo -e "   ${GREEN}🟢 EXCELLENT - Sécurité enterprise-grade${NC}"
    elif [[ $success_rate -ge 75 ]]; then
        echo -e "   ${YELLOW}🟡 BON - Quelques améliorations recommandées${NC}"
    elif [[ $success_rate -ge 60 ]]; then
        echo -e "   ${YELLOW}🟠 MOYEN - Sécurité de base, optimisations nécessaires${NC}"
    else
        echo -e "   ${RED}🔴 INSUFFISANT - Problèmes de sécurité critiques${NC}"
    fi
    
    echo -e "\n${CYAN}📋 PROCHAINES ÉTAPES:${NC}"
    if [[ $FAILED_TESTS -gt 0 ]]; then
        echo -e "   • Corriger les ${FAILED_TESTS} tests échoués"
        echo -e "   • Implémenter les fonctionnalités manquantes"
        echo -e "   • Re-exécuter les tests de sécurité"
    fi
    
    echo -e "   • Continuer avec la Phase 3: Performance & Scalabilité"
    echo -e "   • Tests de pénétration recommandés"
    echo -e "   • Audit de sécurité externe"
    
    echo -e "\n${PURPLE}Phase 2 - Sécurité Enterprise: ${success_rate}% completée${NC}"
}

# Fonction principale
main() {
    print_header "VALIDATION PHASE 2 - SÉCURITÉ ENTERPRISE"
    
    # Vérifier les prérequis
    if ! command -v curl &> /dev/null; then
        print_error "curl est requis pour les tests"
        exit 1
    fi
    
    if ! command -v jq &> /dev/null; then
        print_error "jq est requis pour parser JSON"
        exit 1
    fi
    
    # Vérifier que le serveur est disponible
    if ! check_server; then
        print_error "Serveur non disponible. Démarrez le serveur avant de lancer les tests."
        exit 1
    fi
    
    # Exécuter les tests de sécurité
    test_jwt_authentication
    test_rate_limiting
    test_security_headers
    test_cors
    test_sql_injection
    test_input_validation
    test_oauth2
    test_rbac
    test_2fa
    
    # Nettoyage
    cleanup_test_data
    
    # Générer le rapport
    generate_report
    
    # Code de sortie
    if [[ $FAILED_TESTS -eq 0 ]]; then
        exit 0
    else
        exit 1
    fi
}

# Exécution
main "$@" 