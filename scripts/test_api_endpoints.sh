#!/bin/bash

# Script de test complet des endpoints API Veza
set -e

BASE_URL="http://localhost:8080"
API_URL="$BASE_URL/api/v1"

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

ACCESS_TOKEN=""

log_info() { echo -e "${BLUE}ℹ️  $1${NC}"; }
log_success() { echo -e "${GREEN}✅ $1${NC}"; }
log_error() { echo -e "${RED}❌ $1${NC}"; }

test_endpoint() {
    local method="$1"
    local endpoint="$2"
    local description="$3"
    
    log_info "Test: $description"
    
    if [[ "$method" == "GET" ]]; then
        response=$(curl -s -w '%{http_code}' -o /tmp/response.json "$endpoint")
    fi
    
    http_code="${response: -3}"
    
    if [[ "$http_code" == "200" ]] || [[ "$http_code" == "201" ]]; then
        log_success "$description - HTTP $http_code"
        if [[ -f /tmp/response.json ]]; then
            cat /tmp/response.json | python3 -m json.tool 2>/dev/null | head -5 || echo "Response: $(cat /tmp/response.json)"
        fi
    else
        log_error "$description - HTTP $http_code"
    fi
    echo ""
}

main() {
    echo "🧪 TESTS RAPIDES DE L'API VEZA"
    echo "==============================="
    
    # Vérifier que le serveur fonctionne
    if ! curl -s "$BASE_URL/api/health" > /dev/null; then
        log_error "Serveur non accessible sur $BASE_URL"
        exit 1
    fi
    
    # Tests essentiels
    test_endpoint "GET" "$BASE_URL/api/health" "Health check"
    test_endpoint "GET" "$API_URL/users" "List users"
    test_endpoint "GET" "$API_URL/rooms" "List rooms"
    test_endpoint "GET" "$API_URL/search?q=test" "Global search"
    test_endpoint "GET" "$API_URL/tags" "List tags"
    
    log_success "Tests de base terminés"
}

main "$@"
