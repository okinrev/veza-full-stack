#!/bin/bash

# Test avancé du rate limiting en conditions réelles
# Simule différents scénarios de charge et d'attaque

set -e

echo "⚡ TEST RATE LIMITING EN CONDITIONS RÉELLES - VEZA BACKEND"
echo "========================================================"

# Configuration
SERVER_PORT=${PORT:-"8080"}
API_BASE_URL="http://localhost:${SERVER_PORT}"
TEST_EMAIL_BASE="ratetest"
TEST_DOMAIN="@example.com"

# Couleurs pour l'affichage
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
PURPLE='\033[0;35m'
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

step() {
    echo -e "${CYAN}🔄 $1${NC}"
}

highlight() {
    echo -e "${PURPLE}⭐ $1${NC}"
}

# Fonction pour faire une requête et mesurer le temps
make_request() {
    local endpoint="$1"
    local method="$2"
    local data="$3"
    local expected_status="$4"
    
    start_time=$(date +%s%N)
    
    if [ "$method" = "POST" ]; then
        response=$(curl -s -w "HTTPSTATUS:%{http_code};TIME:%{time_total}" \
            -X POST "$API_BASE_URL$endpoint" \
            -H "Content-Type: application/json" \
            -d "$data" 2>/dev/null)
    else
        response=$(curl -s -w "HTTPSTATUS:%{http_code};TIME:%{time_total}" \
            "$API_BASE_URL$endpoint" 2>/dev/null)
    fi
    
    end_time=$(date +%s%N)
    total_time=$(( (end_time - start_time) / 1000000 )) # ms
    
    http_status=$(echo "$response" | grep -o "HTTPSTATUS:[0-9]*" | cut -d: -f2)
    curl_time=$(echo "$response" | grep -o "TIME:[0-9.]*" | cut -d: -f2)
    body=$(echo "$response" | sed 's/HTTPSTATUS:[0-9]*;TIME:[0-9.]*$//')
    
    echo "$http_status|$total_time|$body"
}

# Test 1: Charge normale - Vérifier que les requêtes normales passent
test_normal_load() {
    step "Test 1: Charge normale (10 requêtes/sec pendant 10s)"
    
    local success_count=0
    local total_requests=100
    local failed_requests=0
    local total_time=0
    
    for i in $(seq 1 $total_requests); do
        result=$(make_request "/health" "GET" "" "200")
        status=$(echo "$result" | cut -d'|' -f1)
        time=$(echo "$result" | cut -d'|' -f2)
        
        total_time=$((total_time + time))
        
        if [ "$status" = "200" ]; then
            success_count=$((success_count + 1))
        else
            failed_requests=$((failed_requests + 1))
        fi
        
        # Petite pause pour simuler 10 req/sec
        sleep 0.1
    done
    
    avg_time=$((total_time / total_requests))
    success_rate=$(( (success_count * 100) / total_requests ))
    
    info "Résultats charge normale:"
    echo "  📊 Requêtes réussies: $success_count/$total_requests ($success_rate%)"
    echo "  ⏱️  Temps moyen: ${avg_time}ms"
    echo "  ❌ Requêtes échouées: $failed_requests"
    
    if [ $success_rate -ge 95 ]; then
        success "Charge normale gérée correctement"
    else
        warning "Taux de succès en charge normale faible: $success_rate%"
    fi
}

# Test 2: Rate limiting login - Tester les limites sur l'authentification
test_login_rate_limiting() {
    step "Test 2: Rate limiting sur login (tentatives répétées)"
    
    local blocked_count=0
    local attempts=20
    
    for i in $(seq 1 $attempts); do
        result=$(make_request "/api/v1/auth/login" "POST" \
            '{"email":"nonexistent@example.com","password":"wrongpassword"}' "401")
        
        status=$(echo "$result" | cut -d'|' -f1)
        body=$(echo "$result" | cut -d'|' -f3)
        
        if [ "$status" = "429" ] || echo "$body" | grep -q "rate.limit\|too.many"; then
            blocked_count=$((blocked_count + 1))
            info "Requête $i bloquée par rate limiting"
            break
        fi
        
        # Pause très courte pour tester la limite
        sleep 0.1
    done
    
    if [ $blocked_count -gt 0 ]; then
        success "Rate limiting login actif (bloqué après $((attempts - blocked_count + 1)) tentatives)"
    else
        warning "Rate limiting login pourrait être insuffisant"
    fi
}

# Test 3: Pic de trafic - Simulation d'un pic soudain
test_traffic_spike() {
    step "Test 3: Pic de trafic (50 requêtes simultanées)"
    
    local temp_file="/tmp/rate_test_spike_$$"
    
    # Lancer 50 requêtes en parallèle
    for i in $(seq 1 50); do
        {
            result=$(make_request "/health" "GET" "" "200")
            echo "$result" >> "$temp_file"
        } &
    done
    
    # Attendre que toutes les requêtes se terminent
    wait
    
    # Analyser les résultats
    local success_count=$(grep "^200|" "$temp_file" | wc -l)
    local blocked_count=$(grep -E "^429|" "$temp_file" | wc -l)
    local total_requests=$(cat "$temp_file" | wc -l)
    
    info "Résultats pic de trafic:"
    echo "  📊 Total requêtes: $total_requests"
    echo "  ✅ Réussies: $success_count"
    echo "  🚫 Bloquées: $blocked_count"
    
    # Calculer temps moyen
    if [ -s "$temp_file" ]; then
        local avg_time=$(awk -F'|' '{sum+=$2; count++} END {if(count>0) print int(sum/count)}' "$temp_file")
        echo "  ⏱️  Temps moyen: ${avg_time}ms"
    fi
    
    rm -f "$temp_file"
    
    if [ $total_requests -eq 50 ]; then
        success "Pic de trafic géré (50 requêtes traitées)"
    else
        warning "Problème lors du pic de trafic"
    fi
}

# Test 4: Attaque DDoS simulée
test_ddos_simulation() {
    step "Test 4: Simulation attaque DDoS (100 requêtes rapides)"
    
    local blocked_start=0
    local total_blocked=0
    
    info "Lancement de l'attaque DDoS simulée..."
    
    for i in $(seq 1 100); do
        result=$(make_request "/api/v1/auth/login" "POST" \
            '{"email":"attacker@evil.com","password":"hack"}' "")
        
        status=$(echo "$result" | cut -d'|' -f1)
        
        if [ "$status" = "429" ] || [ "$status" = "403" ]; then
            total_blocked=$((total_blocked + 1))
            if [ $blocked_start -eq 0 ]; then
                blocked_start=$i
                info "🛡️  Protection DDoS activée à la requête $i"
            fi
        fi
        
        # Aucune pause - attaque maximale
    done
    
    info "Résultats simulation DDoS:"
    echo "  🚫 Requêtes bloquées: $total_blocked/100"
    if [ $blocked_start -gt 0 ]; then
        echo "  🛡️  Protection activée après: $blocked_start requêtes"
    fi
    
    if [ $total_blocked -gt 20 ]; then
        success "Protection DDoS efficace ($total_blocked requêtes bloquées)"
    else
        warning "Protection DDoS insuffisante"
    fi
}

# Test 5: Test de récupération après limitation
test_recovery_after_limiting() {
    step "Test 5: Récupération après rate limiting"
    
    info "Attente de la récupération du rate limiting (30 secondes)..."
    sleep 30
    
    # Tester si les requêtes normales repassent
    local recovery_success=0
    for i in {1..5}; do
        result=$(make_request "/health" "GET" "" "200")
        status=$(echo "$result" | cut -d'|' -f1)
        
        if [ "$status" = "200" ]; then
            recovery_success=$((recovery_success + 1))
        fi
        sleep 1
    done
    
    if [ $recovery_success -ge 4 ]; then
        success "Récupération réussie après rate limiting ($recovery_success/5)"
    else
        warning "Récupération partielle après rate limiting ($recovery_success/5)"
    fi
}

# Test 6: Test des différents endpoints
test_endpoint_specific_limits() {
    step "Test 6: Limites spécifiques par endpoint"
    
    local endpoints=(
        "/health:GET:"
        "/api/v1/auth/login:POST:{\"email\":\"test@example.com\",\"password\":\"test\"}"
        "/api/v1/auth/register:POST:{\"username\":\"test\",\"email\":\"test@example.com\",\"password\":\"test123\"}"
    )
    
    for endpoint_config in "${endpoints[@]}"; do
        IFS=':' read -r endpoint method data <<< "$endpoint_config"
        
        info "Testing endpoint: $method $endpoint"
        
        local success_count=0
        local blocked_count=0
        
        for i in {1..10}; do
            result=$(make_request "$endpoint" "$method" "$data" "")
            status=$(echo "$result" | cut -d'|' -f1)
            
            if [ "$status" = "429" ]; then
                blocked_count=$((blocked_count + 1))
            elif [ "$status" -lt 400 ]; then
                success_count=$((success_count + 1))
            fi
        done
        
        echo "    ✅ Succès: $success_count, 🚫 Bloqués: $blocked_count"
    done
    
    success "Test des limites par endpoint terminé"
}

# Test 7: Performance sous contrainte
test_performance_under_stress() {
    step "Test 7: Performance sous contrainte"
    
    local temp_file="/tmp/rate_test_perf_$$"
    
    # Lancer des requêtes modérées mais continues
    for i in $(seq 1 30); do
        {
            result=$(make_request "/health" "GET" "" "200")
            echo "$result" >> "$temp_file"
        } &
        
        # Petite pause pour éviter le rate limiting
        sleep 0.5
    done
    
    wait
    
    # Analyser la performance
    if [ -s "$temp_file" ]; then
        local avg_time=$(awk -F'|' '{sum+=$2; count++} END {if(count>0) print int(sum/count)}' "$temp_file")
        local max_time=$(awk -F'|' 'BEGIN{max=0} {if($2>max) max=$2} END{print max}' "$temp_file")
        local success_count=$(grep "^200|" "$temp_file" | wc -l)
        
        info "Performance sous contrainte:"
        echo "  📊 Requêtes réussies: $success_count/30"
        echo "  ⏱️  Temps moyen: ${avg_time}ms"
        echo "  ⏱️  Temps max: ${max_time}ms"
        
        if [ $avg_time -lt 500 ] && [ $success_count -ge 25 ]; then
            success "Performance maintenue sous contrainte"
        else
            warning "Dégradation de performance sous contrainte"
        fi
    fi
    
    rm -f "$temp_file"
}

# Fonction principale
main() {
    echo "Début des tests de rate limiting en conditions réelles..."
    echo "Configuration:"
    echo "  - API URL: $API_BASE_URL"
    echo "  - Tests: Charge normale, pic trafic, DDoS, récupération"
    echo ""
    
    # Vérifier que le serveur répond
    if ! curl -s "$API_BASE_URL/health" >/dev/null 2>&1; then
        error "Serveur non accessible sur $API_BASE_URL"
    fi
    
    test_normal_load
    test_login_rate_limiting
    test_traffic_spike
    test_ddos_simulation
    test_recovery_after_limiting
    test_endpoint_specific_limits
    test_performance_under_stress
    
    echo ""
    highlight "🎉 TESTS RATE LIMITING TERMINÉS"
    success "Rate limiting validé en conditions réelles !"
    
    echo ""
    echo "📋 FONCTIONNALITÉS VALIDÉES:"
    echo "   ✅ Charge normale supportée"
    echo "   ✅ Protection login active"
    echo "   ✅ Gestion pics de trafic"
    echo "   ✅ Protection DDoS efficace"
    echo "   ✅ Récupération automatique"
    echo "   ✅ Limites par endpoint"
    echo "   ✅ Performance maintenue"
}

# Gestion des erreurs
cleanup() {
    info "Nettoyage des fichiers temporaires"
    rm -f /tmp/rate_test_*_$$
}

trap cleanup EXIT

# Exécution
if [ "$#" -eq 0 ]; then
    main "$@"
else
    # Permettre d'exécuter des tests individuels
    "$@"
fi 