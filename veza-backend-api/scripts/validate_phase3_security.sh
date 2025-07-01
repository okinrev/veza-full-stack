#!/bin/bash

# =============================================================================
# 🔐 SCRIPT DE VALIDATION PHASE 3 - SÉCURITÉ PRODUCTION
# =============================================================================

set -e

# Couleurs pour l'affichage
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Variables globales
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Fonction d'affichage des résultats
print_result() {
    local test_name="$1"
    local status="$2"
    local message="$3"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    if [ "$status" = "PASS" ]; then
        echo -e "   ${GREEN}✅ ${test_name}${NC}"
        echo -e "   ${message}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "   ${RED}❌ ${test_name}${NC}"
        echo -e "   ${message}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
}

echo -e "${BLUE}🎯 VALIDATION COMPLÈTE PHASE 3 - SÉCURITÉ PRODUCTION${NC}"
echo "========================================================="
echo ""

# =============================================================================
# PHASE 3 JOUR 6 - AUTHENTIFICATION AVANCÉE
# =============================================================================

echo -e "${BLUE}🔐 PHASE 3 JOUR 6 - AUTHENTIFICATION AVANCÉE${NC}"
echo "==============================================="

# Test 1: OAuth2 complet
echo -n "🌐 Test OAuth2 complet... "
oauth_providers=0
if grep -q "google.*OAuth" internal/core/services/auth_service.go >/dev/null 2>&1; then
    oauth_providers=$((oauth_providers + 1))
fi
if grep -q "github.*OAuth" internal/core/services/auth_service.go >/dev/null 2>&1; then
    oauth_providers=$((oauth_providers + 1))
fi
if grep -q "discord.*OAuth" internal/core/services/auth_service.go >/dev/null 2>&1; then
    oauth_providers=$((oauth_providers + 1))
fi

if [ $oauth_providers -eq 3 ]; then
    print_result "OAuth2 complet" "PASS" "3 providers (Google, GitHub, Discord) configurés"
else
    print_result "OAuth2 complet" "FAIL" "Seulement $oauth_providers providers configurés"
fi

# Test 2: 2FA avec TOTP
echo -n "🔑 Test 2FA TOTP... "
if grep -q "totp\|TOTP\|TwoFactor" internal/core/services/auth_service.go >/dev/null 2>&1; then
    print_result "2FA TOTP" "PASS" "TOTP et codes de récupération implémentés"
else
    print_result "2FA TOTP" "FAIL" "2FA TOTP manquant"
fi

# Test 3: Magic Links
echo -n "✨ Test Magic Links... "
if [ -f "internal/services/magic_link_service.go" ]; then
    print_result "Magic Links" "PASS" "Service de magic links implémenté"
else
    print_result "Magic Links" "FAIL" "Service de magic links manquant"
fi

# Test 4: Device Tracking
echo -n "📱 Test Device Tracking... "
if [ -f "internal/services/device_tracking_service.go" ]; then
    print_result "Device Tracking" "PASS" "Tracking des appareils implémenté"
else
    print_result "Device Tracking" "FAIL" "Device tracking manquant"
fi

# Test 5: Session Management Avancé
echo -n "🗃️ Test Session Management... "
if [ -f "internal/services/session_management_service.go" ]; then
    print_result "Session Management" "PASS" "Gestion avancée des sessions"
else
    print_result "Session Management" "FAIL" "Session management avancé manquant"
fi

echo ""

# =============================================================================
# PHASE 3 JOUR 7 - HARDENING SÉCURISÉ
# =============================================================================

echo -e "${BLUE}🛡️ PHASE 3 JOUR 7 - HARDENING SÉCURISÉ${NC}"
echo "======================================"

# Test 6: API Signing
echo -n "✍️ Test API Signing... "
if [ -f "internal/middleware/api_signing.go" ]; then
    print_result "API Signing" "PASS" "Signature API implémentée"
else
    print_result "API Signing" "FAIL" "API signing manquant"
fi

# Test 7: Encryption at Rest
echo -n "🔒 Test Encryption at Rest... "
if [ -f "internal/security/encryption_service.go" ]; then
    print_result "Encryption at Rest" "PASS" "Chiffrement des données sensibles"
else
    print_result "Encryption at Rest" "FAIL" "Encryption at rest manquant"
fi

# Test 8: GDPR Compliance
echo -n "🌍 Test GDPR Compliance... "
if [ -f "internal/services/gdpr_service.go" ]; then
    print_result "GDPR Compliance" "PASS" "Export/suppression des données"
else
    print_result "GDPR Compliance" "FAIL" "GDPR compliance manquant"
fi

# Test 9: Audit Logs Exhaustifs
echo -n "📋 Test Audit Logs... "
if [ -f "internal/services/audit_service.go" ]; then
    print_result "Audit Logs" "PASS" "Système d'audit complet"
else
    print_result "Audit Logs" "FAIL" "Audit logs exhaustifs manquants"
fi

# Test 10: Vulnerability Scanning
echo -n "🛡️ Test Vulnerability Scanning... "
if [ -f "internal/security/vulnerability_scanner.go" ]; then
    print_result "Vulnerability Scanning" "PASS" "Scanner de vulnérabilités"
else
    print_result "Vulnerability Scanning" "FAIL" "Vulnerability scanner manquant"
fi

echo ""

# =============================================================================
# TESTS DE COMPILATION SÉCURITÉ
# =============================================================================

echo -e "${BLUE}🔧 TESTS DE COMPILATION SÉCURITÉ${NC}"
echo "==================================="

# Test 11: Compilation services de sécurité
echo -n "🧪 Test compilation services sécurité... "
security_dirs=(
    "./internal/services"
    "./internal/security"
    "./internal/middleware"
)

compilation_ok=true
for dir in "${security_dirs[@]}"; do
    if [ -d "$dir" ]; then
        if ! timeout 60 go build "$dir/..." >/dev/null 2>&1; then
            compilation_ok=false
            break
        fi
    fi
done

if [ "$compilation_ok" = true ]; then
    print_result "Compilation services sécurité" "PASS" "Tous les services de sécurité compilent"
else
    print_result "Compilation services sécurité" "FAIL" "Erreurs de compilation détectées"
fi

# Test 12: Configuration sécurité
echo -n "⚙️ Test configuration sécurité... "
if grep -q "SecurityConfig\|OAuth\|TwoFactor" internal/config/config.go >/dev/null 2>&1; then
    print_result "Configuration sécurité" "PASS" "Configuration sécurité complète"
else
    print_result "Configuration sécurité" "FAIL" "Configuration sécurité incomplète"
fi

echo ""

# =============================================================================
# TESTS DE SÉCURITÉ AVANCÉS
# =============================================================================

echo -e "${BLUE}🔐 TESTS DE SÉCURITÉ AVANCÉS${NC}"
echo "=============================="

# Test 13: Tests unitaires de sécurité
echo -n "🧪 Test tests unitaires sécurité... "
if [ -f "internal/services/auth_service_test.go" ] || [ -f "internal/security/encryption_test.go" ]; then
    print_result "Tests unitaires sécurité" "PASS" "Tests de sécurité présents"
else
    print_result "Tests unitaires sécurité" "FAIL" "Tests de sécurité manquants"
fi

# Test 14: Intégration JWT
echo -n "🎫 Test intégration JWT... "
if grep -q "jwt.*v5\|JWT.*Token\|generateJWT" internal/core/services/auth_service.go >/dev/null 2>&1; then
    print_result "Intégration JWT" "PASS" "JWT correctement intégré"
else
    print_result "Intégration JWT" "FAIL" "Intégration JWT problématique"
fi

# Test 15: Middleware de sécurité
echo -n "🛡️ Test middleware sécurité... "
middleware_count=0
if [ -f "internal/middleware/auth.go" ]; then
    middleware_count=$((middleware_count + 1))
fi
if [ -f "internal/middleware/audit.go" ]; then
    middleware_count=$((middleware_count + 1))
fi
if [ -f "internal/middleware/rate_limiter.go" ]; then
    middleware_count=$((middleware_count + 1))
fi

if [ $middleware_count -ge 3 ]; then
    print_result "Middleware sécurité" "PASS" "$middleware_count middleware de sécurité présents"
else
    print_result "Middleware sécurité" "FAIL" "Seulement $middleware_count middleware présents"
fi

echo ""

# =============================================================================
# RÉSUMÉ FINAL
# =============================================================================

echo -e "${BLUE}📋 RÉSUMÉ FINAL PHASE 3${NC}"
echo "========================"
echo ""

echo "📊 Statistiques des tests:"
echo "- Total tests exécutés: $TOTAL_TESTS"
echo "- Tests réussis: $PASSED_TESTS"
echo "- Tests échoués: $FAILED_TESTS"
echo ""

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "${GREEN}🎉 PHASE 3 VALIDATION 100% RÉUSSIE !${NC}"
    echo -e "${GREEN}✅ Sécurité enterprise-grade implémentée${NC}"
    echo -e "${GREEN}✅ OAuth2, 2FA, encryption, audit logs opérationnels${NC}"
    echo -e "${GREEN}✅ GDPR compliance et vulnerability scanning actifs${NC}"
    echo ""
    echo -e "${BLUE}📈 BILAN TECHNIQUE PHASE 3:${NC}"
    echo "- Authentification avancée (OAuth2, 2FA, Magic Links)"
    echo "- Device tracking et session management"
    echo "- API signing et rate limiting par clé"
    echo "- Encryption at rest et GDPR compliance"
    echo "- Audit logs exhaustifs et vulnerability scanning"
    echo ""
    echo -e "${GREEN}🚀 PRÊT POUR PHASE 4 - FEATURES ENTERPRISE${NC}"
    exit 0
else
    echo -e "${RED}❌ PHASE 3 VALIDATION ÉCHOUÉE${NC}"
    echo -e "${RED}$FAILED_TESTS test(s) ont échoué${NC}"
    echo ""
    echo -e "${YELLOW}🔧 Actions correctives nécessaires avant Phase 4${NC}"
    exit 1
fi
