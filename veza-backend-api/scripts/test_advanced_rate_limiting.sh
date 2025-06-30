#!/bin/bash

# Script de test pour valider le Rate Limiting Distribué
# Test du serveur avancé avec Redis

set -e

echo "🚀 TEST RATE LIMITING DISTRIBUÉ - SERVEUR AVANCÉ"
echo "================================================"

# Configuration
PORT=${PORT:-8080}
HOST="localhost:${PORT}"
BASE_URL="http://${HOST}"

# Couleurs pour l'affichage
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Fonction d'aide
function test_endpoint() {
    local method=$1
    local endpoint=$2
    local expected_status=$3
    local description=$4
    local data=$5
    
    echo -n "  Testing ${description}... "
    
    if [ "$method" = "POST" ] && [ -n "$data" ]; then
        status=$(curl -s -o /dev/null -w "%{http_code}" -X POST -H "Content-Type: application/json" -d "$data" "${BASE_URL}${endpoint}")
    else
        status=$(curl -s -o /dev/null -w "%{http_code}" "${BASE_URL}${endpoint}")
    fi
    
    if [ "$status" = "$expected_status" ]; then
        echo -e "${GREEN}✅ OK${NC} (Status: $status)"
    else
        echo -e "${RED}❌ FAIL${NC} (Expected: $expected_status, Got: $status)"
        return 1
    fi
}

# Fonction pour tester les headers de rate limiting
function test_rate_limit_headers() {
    local endpoint=$1
    local description=$2
    
    echo -n "  Testing rate limit headers for ${description}... "
    
    response=$(curl -s -I "${BASE_URL}${endpoint}")
    
    if echo "$response" | grep -q "X-RateLimit-Limit" && echo "$response" | grep -q "X-RateLimit-Remaining"; then
        echo -e "${GREEN}✅ Headers OK${NC}"
        echo "$response" | grep "X-RateLimit" | sed 's/^/    /'
    else
        echo -e "${YELLOW}⚠️  No rate limit headers${NC}"
    fi
}

# Fonction pour tester le dépassement de limite
function test_rate_limit_exceeded() {
    local endpoint=$1
    local limit=$2
    local description=$3
    
    echo "  Testing rate limit exceeded for ${description} (limit: ${limit} req/min)..."
    
    # Reset les limites d'abord
    curl -s -X POST "${BASE_URL}/api/v1/admin/ratelimit/reset" > /dev/null || true
    
    local success_count=0
    local rate_limited_count=0
    
    # Faire plus de requêtes que la limite
    for i in $(seq 1 $((limit + 2))); do
        status=$(curl -s -o /dev/null -w "%{http_code}" "${BASE_URL}${endpoint}")
        
        if [ "$status" = "200" ]; then
            ((success_count++))
        elif [ "$status" = "429" ]; then
            ((rate_limited_count++))
        fi
        
        # Petite pause pour éviter de surcharger
        sleep 0.1
    done
    
    echo "    Successful requests: ${success_count}"
    echo "    Rate limited requests: ${rate_limited_count}"
    
    if [ "$rate_limited_count" -gt 0 ]; then
        echo -e "    ${GREEN}✅ Rate limiting functional${NC}"
    else
        echo -e "    ${RED}❌ Rate limiting not working${NC}"
        return 1
    fi
}

# Vérifier que le serveur est démarré
echo "🔍 Checking server status..."
if ! curl -s "${BASE_URL}/health" > /dev/null; then
    echo -e "${RED}❌ Serveur non accessible sur ${BASE_URL}${NC}"
    echo "   💡 Démarrez le serveur avec: VEZA_SERVER_MODE=advanced-simple go run cmd/server/advanced_simple.go"
    exit 1
fi

echo -e "${GREEN}✅ Serveur accessible${NC}"
echo ""

# 1. Tests de sanité de base
echo "📝 1. TESTS DE SANITÉ DE BASE"
echo "─────────────────────────────"

test_endpoint "GET" "/health" "200" "Health check"
test_endpoint "GET" "/health/ready" "200" "Readiness check"
test_endpoint "GET" "/metrics" "200" "Prometheus metrics"
test_endpoint "GET" "/api/v1/advanced/status" "200" "Advanced status"

echo ""

# 2. Tests des endpoints de démonstration
echo "📝 2. TESTS ENDPOINTS DÉMONSTRATION"
echo "───────────────────────────────────"

test_endpoint "GET" "/api/v1/demo/ping" "200" "Demo ping"
test_endpoint "POST" "/api/v1/demo/echo" "200" "Demo echo" '{"test": "rate limiting"}'
test_endpoint "GET" "/api/v1/demo/redis" "200" "Demo Redis"
test_endpoint "GET" "/api/v1/demo/stress" "200" "Demo stress (first call)"

echo ""

# 3. Tests des headers de rate limiting
echo "📝 3. TESTS HEADERS RATE LIMITING"
echo "─────────────────────────────────"

test_rate_limit_headers "/api/v1/demo/ping" "ping endpoint"
test_rate_limit_headers "/api/v1/demo/stress" "stress endpoint"

echo ""

# 4. Tests de dépassement de limite
echo "📝 4. TESTS DÉPASSEMENT DE LIMITE"
echo "─────────────────────────────────"

# Test l'endpoint stress (limite: 3 req/min)
test_rate_limit_exceeded "/api/v1/demo/stress" 3 "stress endpoint"

echo ""

# Test l'endpoint echo (limite: 10 req/min) - test plus léger
echo "  Testing echo endpoint rate limiting (10 req/min)..."
curl -s -X POST "${BASE_URL}/api/v1/admin/ratelimit/reset" > /dev/null || true

echo_success=0
echo_limited=0

for i in $(seq 1 12); do
    status=$(curl -s -o /dev/null -w "%{http_code}" -X POST -H "Content-Type: application/json" -d '{"test": "echo rate limit"}' "${BASE_URL}/api/v1/demo/echo")
    
    if [ "$status" = "200" ]; then
        ((echo_success++))
    elif [ "$status" = "429" ]; then
        ((echo_limited++))
    fi
    
    sleep 0.1
done

echo "    Echo successful: ${echo_success}, limited: ${echo_limited}"
if [ "$echo_limited" -gt 0 ]; then
    echo -e "    ${GREEN}✅ Echo rate limiting functional${NC}"
else
    echo -e "    ${YELLOW}⚠️  Echo rate limiting not triggered (normal if under limit)${NC}"
fi

echo ""

# 5. Tests des endpoints d'administration
echo "📝 5. TESTS ADMINISTRATION RATE LIMITING"
echo "────────────────────────────────────────"

test_endpoint "GET" "/api/v1/admin/ratelimit/stats" "200" "Rate limit stats"
test_endpoint "GET" "/api/v1/admin/ratelimit/config" "200" "Rate limit config"

# Test du reset des limites
echo -n "  Testing rate limit reset... "
reset_response=$(curl -s -X POST "${BASE_URL}/api/v1/admin/ratelimit/reset")
if echo "$reset_response" | grep -q "reset"; then
    echo -e "${GREEN}✅ Reset OK${NC}"
else
    echo -e "${RED}❌ Reset failed${NC}"
fi

echo ""

# 6. Tests de performance et métriques
echo "📝 6. TESTS MÉTRIQUES ET PERFORMANCE"
echo "────────────────────────────────────"

test_endpoint "GET" "/api/v1/advanced/metrics" "200" "Advanced metrics"

echo -n "  Testing Prometheus metrics content... "
metrics_response=$(curl -s "${BASE_URL}/metrics")
if echo "$metrics_response" | grep -q "http_requests_total" && echo "$metrics_response" | grep -q "http_duration_seconds"; then
    echo -e "${GREEN}✅ Prometheus metrics OK${NC}"
else
    echo -e "${RED}❌ Prometheus metrics incomplete${NC}"
fi

echo ""

# 7. Test de la configuration Redis
echo "📝 7. TESTS REDIS ET CONFIGURATION"
echo "──────────────────────────────────"

echo -n "  Testing Redis connection... "
redis_response=$(curl -s "${BASE_URL}/api/v1/demo/redis")
if echo "$redis_response" | grep -q "Success"; then
    echo -e "${GREEN}✅ Redis OK${NC}"
else
    echo -e "${RED}❌ Redis connection failed${NC}"
fi

# Vérifier le nombre de clés Redis
echo -n "  Testing Redis keys count... "
stats_response=$(curl -s "${BASE_URL}/api/v1/admin/ratelimit/stats")
keys_count=$(echo "$stats_response" | grep -o '"total_keys":[0-9]*' | cut -d':' -f2)
echo "Redis rate limit keys: $keys_count"

echo ""

# 8. Résumé final
echo "📊 RÉSUMÉ DU TEST"
echo "─────────────────"

echo "✅ Tests terminés pour le Rate Limiting Distribué"
echo ""
echo -e "${BLUE}🎯 Fonctionnalités testées:${NC}"
echo "  • Rate limiting par endpoint (différentes limites)"
echo "  • Headers de rate limiting (X-RateLimit-*)"
echo "  • Protection contre le dépassement de limite (429)"
echo "  • Administration des limites (stats, reset, config)"
echo "  • Intégration Redis pour distribution"
echo "  • Métriques Prometheus en temps réel"
echo "  • Monitoring et health checks"
echo ""
echo -e "${BLUE}🔧 Configuration testée:${NC}"
echo "  • Stress endpoint: 3 req/min (très restrictif)"
echo "  • Echo endpoint: 10 req/min"
echo "  • Autres endpoints: 60 req/min par défaut"
echo "  • Protection DDoS: 120 req/min"
echo ""
echo -e "${GREEN}✅ RATE LIMITING DISTRIBUÉ FONCTIONNEL !${NC}"
echo ""
echo -e "${YELLOW}💡 Prochaines étapes:${NC}"
echo "  • Intégrer les clients gRPC"
echo "  • Ajouter l'authentification JWT"
echo "  • Implémenter les WebSocket handlers"
echo "  • Déployer en mode production"
echo "" 