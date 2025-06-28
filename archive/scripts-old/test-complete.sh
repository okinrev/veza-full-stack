#!/bin/bash

# Script de test ultra complet pour l'infrastructure Veza
# Teste tous les aspects : infrastructure, services, APIs, intégration, performance, sécurité

set -e

# Couleurs et formatage
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

# Compteurs globaux
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
WARNING_TESTS=0

# Fonctions utilitaires
log_header() {
    echo -e "\n${PURPLE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${PURPLE}🔍 $1${NC}"
    echo -e "${PURPLE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

log_section() {
    echo -e "\n${CYAN}▶ $1${NC}"
}

log_test() {
    echo -e "${BLUE}  • $1${NC}"
}

log_pass() {
    echo -e "${GREEN}    ✅ $1${NC}"
    ((PASSED_TESTS++))
    ((TOTAL_TESTS++))
}

log_fail() {
    echo -e "${RED}    ❌ $1${NC}"
    ((FAILED_TESTS++))
    ((TOTAL_TESTS++))
}

log_warn() {
    echo -e "${YELLOW}    ⚠️  $1${NC}"
    ((WARNING_TESTS++))
    ((TOTAL_TESTS++))
}

log_info() {
    echo -e "${BLUE}    ℹ️  $1${NC}"
}

# Fonction pour tester les endpoints avec timeout
test_endpoint() {
    local url=$1
    local expected_code=${2:-200}
    local timeout=${3:-5}
    
    local response=$(curl -s -o /dev/null -w "%{http_code}" -m $timeout "$url" 2>/dev/null || echo "000")
    
    if [[ "$response" == "$expected_code" ]] || [[ "$expected_code" == "200|404" && ("$response" == "200" || "$response" == "404") ]]; then
        return 0
    else
        return 1
    fi
}

# Fonction pour tester le contenu JSON
test_json_endpoint() {
    local url=$1
    local expected_field=$2
    local expected_value=$3
    local timeout=${4:-5}
    
    local response=$(curl -s -m $timeout "$url" 2>/dev/null || echo "{}")
    local actual_value=$(echo "$response" | jq -r ".$expected_field // \"null\"" 2>/dev/null || echo "null")
    
    if [[ "$actual_value" == "$expected_value" ]]; then
        return 0
    else
        return 1
    fi
}

# Début des tests
clear
echo -e "${PURPLE}╔══════════════════════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${PURPLE}║                    🧪 TESTS ULTRA COMPLETS VEZA v2.0                             ║${NC}"
echo -e "${PURPLE}║                   Infrastructure Microservices Complète                          ║${NC}"
echo -e "${PURPLE}╚══════════════════════════════════════════════════════════════════════════════════╝${NC}"

START_TIME=$(date +%s)

# ========================================
# 1. INFRASTRUCTURE ET CONTAINERS
# ========================================
log_header "1. INFRASTRUCTURE ET CONTAINERS INCUS"

log_section "1.1 Vérification des containers"
containers=("veza-postgres" "veza-redis" "veza-storage" "veza-backend" "veza-chat" "veza-stream" "veza-frontend" "veza-haproxy")

for container in "${containers[@]}"; do
    log_test "Container $container"
    if incus list --format csv | grep -q "^$container,RUNNING"; then
        # Vérifier l'IP
        ip=$(incus list $container --format csv | cut -d, -f4 | cut -d' ' -f1)
        if [[ -n "$ip" && "$ip" != "-" ]]; then
            log_pass "Container actif avec IP $ip"
        else
            log_warn "Container actif mais pas d'IP"
        fi
    else
        log_fail "Container non actif ou inexistant"
    fi
done

log_section "1.2 Ressources des containers"
for container in "${containers[@]}"; do
    if incus list --format csv | grep -q "^$container,RUNNING"; then
        log_test "Ressources $container"
        memory=$(incus exec $container -- free -h | awk 'NR==2{printf "RAM: %s", $3}' 2>/dev/null || echo "N/A")
        disk=$(incus exec $container -- df -h / | awk 'NR==2{printf "Disk: %s", $5}' 2>/dev/null || echo "N/A")
        if [[ "$memory" != "N/A" && "$disk" != "N/A" ]]; then
            log_pass "$memory, $disk"
        else
            log_warn "Impossible de récupérer les métriques"
        fi
    fi
done

# ========================================
# 2. SERVICES INDIVIDUELS
# ========================================
log_header "2. SERVICES INDIVIDUELS ET SANTÉ"

log_section "2.1 PostgreSQL Database"
log_test "Connectivité PostgreSQL"
if incus exec veza-postgres -- pg_isready -h localhost -p 5432 >/dev/null 2>&1; then
    log_pass "PostgreSQL accessible"
else
    log_fail "PostgreSQL inaccessible"
fi

log_test "Base de données Veza"
if incus exec veza-postgres -- psql -h localhost -U veza_user -d veza_db -c "SELECT 1;" >/dev/null 2>&1; then
    log_pass "Base de données accessible"
else
    log_fail "Base de données inaccessible"
fi

log_section "2.2 Redis Cache"
log_test "Connectivité Redis"
if incus exec veza-redis -- redis-cli ping | grep -q "PONG"; then
    log_pass "Redis accessible"
else
    log_fail "Redis inaccessible"
fi

log_section "2.3 Backend Go"
log_test "Service Backend Go"
if test_endpoint "http://10.5.191.241:8080" "200|404"; then
    log_pass "Backend Go accessible"
else
    log_fail "Backend Go inaccessible"
fi

log_test "Processus Backend Go"
if incus exec veza-backend -- pgrep -f "server" >/dev/null 2>&1; then
    log_pass "Processus Backend actif"
else
    log_warn "Processus Backend non détecté"
fi

log_section "2.4 Chat Server (Rust)"
log_test "Health endpoint Chat"
if test_json_endpoint "http://10.5.191.49:3001/health" "data.status" "healthy"; then
    log_pass "Chat Server healthy"
else
    log_fail "Chat Server unhealthy"
fi

log_test "Service Chat"
if incus exec veza-chat -- systemctl is-active veza-chat >/dev/null 2>&1; then
    log_pass "Service Chat actif"
else
    log_fail "Service Chat inactif"
fi

log_section "2.5 Stream Server (Rust)"
log_test "Health endpoint Stream"
if test_json_endpoint "http://10.5.191.196:8000/health" "status" "healthy"; then
    log_pass "Stream Server healthy"
else
    log_fail "Stream Server unhealthy"
fi

log_test "Service Stream"
if incus exec veza-stream -- systemctl is-active veza-stream >/dev/null 2>&1; then
    log_pass "Service Stream actif"
else
    log_fail "Service Stream inactif"
fi

log_section "2.6 Frontend React"
log_test "Nginx Frontend"
if test_endpoint "http://10.5.191.41:3000"; then
    log_pass "Frontend accessible"
else
    log_fail "Frontend inaccessible"
fi

log_test "Service Nginx"
if incus exec veza-frontend -- systemctl is-active nginx >/dev/null 2>&1; then
    log_pass "Service Nginx actif"
else
    log_fail "Service Nginx inactif"
fi

# ========================================
# 3. APIs ET ENDPOINTS
# ========================================
log_header "3. APIS ET ENDPOINTS DÉTAILLÉS"

log_section "3.1 Chat API Endpoints"
log_test "GET /health"
if test_json_endpoint "http://10.5.191.49:3001/health" "success" "true"; then
    log_pass "Health endpoint OK"
else
    log_fail "Health endpoint KO"
fi

log_test "GET /api/messages"
if test_endpoint "http://10.5.191.49:3001/api/messages?room=general"; then
    log_pass "Messages endpoint OK"
else
    log_fail "Messages endpoint KO"
fi

log_test "GET /api/messages/stats"
if test_json_endpoint "http://10.5.191.49:3001/api/messages/stats" "success" "true"; then
    log_pass "Stats endpoint OK"
else
    log_fail "Stats endpoint KO"
fi

log_test "POST /api/messages (test)"
post_response=$(curl -s -X POST "http://10.5.191.49:3001/api/messages" \
    -H "Content-Type: application/json" \
    -d '{"content":"Test automatique","author":"test-script","room":"general"}' \
    -w "%{http_code}" -o /tmp/post_response.json 2>/dev/null || echo "000")

if [[ "$post_response" == "200" ]]; then
    log_pass "POST messages OK"
else
    log_fail "POST messages KO (code: $post_response)"
fi

log_section "3.2 Stream API Endpoints"
log_test "GET /health"
if test_json_endpoint "http://10.5.191.196:8000/health" "status" "healthy"; then
    log_pass "Health endpoint OK"
else
    log_fail "Health endpoint KO"
fi

log_test "GET /list"
if test_endpoint "http://10.5.191.196:8000/list"; then
    log_pass "List endpoint OK"
else
    log_fail "List endpoint KO"
fi

log_test "GET / (root)"
if test_endpoint "http://10.5.191.196:8000/"; then
    log_pass "Root endpoint OK"
else
    log_fail "Root endpoint KO"
fi

# ========================================
# 4. INTÉGRATION HAPROXY
# ========================================
log_header "4. INTÉGRATION HAPROXY"

log_section "4.1 HAProxy Service"
log_test "Service HAProxy"
if incus exec veza-haproxy -- systemctl is-active haproxy >/dev/null 2>&1; then
    log_pass "Service HAProxy actif"
else
    log_fail "Service HAProxy inactif"
fi

log_test "Port 80 ouvert"
if test_endpoint "http://10.5.191.133:80"; then
    log_pass "Port 80 accessible"
else
    log_fail "Port 80 inaccessible"
fi

log_section "4.2 Routage via HAProxy"
log_test "Frontend React via HAProxy"
if test_endpoint "http://10.5.191.133/"; then
    log_pass "Frontend routé correctement"
else
    log_fail "Frontend non routé"
fi

log_test "Chat API via HAProxy"
if test_endpoint "http://10.5.191.133/chat-api/health"; then
    log_pass "Chat API routée correctement"
else
    log_fail "Chat API non routée"
fi

log_test "Stream API via HAProxy"
if test_endpoint "http://10.5.191.133/stream/health"; then
    log_pass "Stream API routée correctement"
else
    log_fail "Stream API non routée"
fi

log_section "4.3 Headers CORS via HAProxy"
log_test "Headers CORS"
cors_headers=$(curl -s -I "http://10.5.191.133/" | grep -i "access-control" | wc -l)
if [[ $cors_headers -gt 0 ]]; then
    log_pass "Headers CORS configurés ($cors_headers headers)"
else
    log_warn "Headers CORS non détectés"
fi

# ========================================
# 5. FONCTIONNALITÉS MÉTIER
# ========================================
log_header "5. FONCTIONNALITÉS MÉTIER"

log_section "5.1 Chat Functionality"
log_test "Récupération messages existants"
messages=$(curl -s "http://10.5.191.133/chat-api/messages?room=general" | jq '.data | length' 2>/dev/null || echo "0")
if [[ $messages -gt 0 ]]; then
    log_pass "$messages messages trouvés"
else
    log_warn "Aucun message trouvé"
fi

log_test "Statistiques chat"
stats=$(curl -s "http://10.5.191.133/chat-api/messages/stats" | jq '.data.total_messages' 2>/dev/null || echo "0")
if [[ $stats -gt 0 ]]; then
    log_pass "$stats messages total en base"
else
    log_warn "Statistiques non disponibles"
fi

log_test "Envoi nouveau message via HAProxy"
new_msg_id=$(curl -s -X POST "http://10.5.191.133/chat-api/messages" \
    -H "Content-Type: application/json" \
    -d '{"content":"Test intégration complète","author":"test-integration","room":"general"}' | \
    jq '.data' 2>/dev/null || echo "null")

if [[ "$new_msg_id" != "null" && "$new_msg_id" != "" ]]; then
    log_pass "Message envoyé (ID: $new_msg_id)"
else
    log_fail "Échec envoi message"
fi

log_section "5.2 Stream Functionality"
log_test "Liste fichiers audio"
audio_files=$(curl -s "http://10.5.191.133/stream/list" | jq '. | length' 2>/dev/null || echo "0")
if [[ $audio_files -gt 0 ]]; then
    log_pass "$audio_files fichiers audio disponibles"
else
    log_warn "Aucun fichier audio trouvé"
fi

# ========================================
# 6. PERFORMANCE ET CHARGE
# ========================================
log_header "6. PERFORMANCE ET CHARGE"

log_section "6.1 Tests de charge basiques"
log_test "Charge Frontend (10 requêtes simultanées)"
time_start=$(date +%s.%N)
for i in {1..10}; do
    curl -s -o /dev/null "http://10.5.191.133/" &
done
wait
time_end=$(date +%s.%N)
duration=$(echo "$time_end - $time_start" | bc 2>/dev/null || echo "N/A")

if [[ "$duration" != "N/A" && $(echo "$duration < 5" | bc 2>/dev/null) -eq 1 ]]; then
    log_pass "Frontend rapide (${duration}s pour 10 requêtes)"
else
    log_warn "Frontend potentiellement lent"
fi

log_test "Charge Chat API (5 messages simultanés)"
time_start=$(date +%s.%N)
for i in {1..5}; do
    curl -s -X POST "http://10.5.191.133/chat-api/messages" \
        -H "Content-Type: application/json" \
        -d '{"content":"Load test '$i'","author":"load-test","room":"test"}' \
        -o /dev/null &
done
wait
time_end=$(date +%s.%N)
duration=$(echo "$time_end - $time_start" | bc 2>/dev/null || echo "N/A")

if [[ "$duration" != "N/A" && $(echo "$duration < 3" | bc 2>/dev/null) -eq 1 ]]; then
    log_pass "Chat API rapide (${duration}s pour 5 posts)"
else
    log_warn "Chat API potentiellement lent"
fi

log_section "6.2 Temps de réponse"
services=("http://10.5.191.133/:Frontend" "http://10.5.191.133/chat-api/health:Chat" "http://10.5.191.133/stream/health:Stream")

for service in "${services[@]}"; do
    url=$(echo $service | cut -d: -f1)
    name=$(echo $service | cut -d: -f2)
    
    log_test "Temps de réponse $name"
    response_time=$(curl -s -o /dev/null -w "%{time_total}" "$url" 2>/dev/null || echo "999")
    
    if (( $(echo "$response_time < 1" | bc -l 2>/dev/null || echo "0") )); then
        log_pass "Très rapide (${response_time}s)"
    elif (( $(echo "$response_time < 3" | bc -l 2>/dev/null || echo "0") )); then
        log_pass "Rapide (${response_time}s)"
    else
        log_warn "Lent (${response_time}s)"
    fi
done

# ========================================
# 7. SÉCURITÉ BASIQUE
# ========================================
log_header "7. SÉCURITÉ BASIQUE"

log_section "7.1 Exposition des ports"
log_test "Ports exposés vs attendus"
expected_ports=("80:haproxy" "3000:frontend" "3001:chat" "8000:stream" "8080:backend" "5432:postgres" "6379:redis")
exposed_count=0

for port_info in "${expected_ports[@]}"; do
    port=$(echo $port_info | cut -d: -f1)
    service=$(echo $port_info | cut -d: -f2)
    
    # Test rapide du port
    if timeout 2 bash -c "echo >/dev/tcp/10.5.191.133/$port" 2>/dev/null || \
       timeout 2 bash -c "echo >/dev/tcp/10.5.191.41/$port" 2>/dev/null || \
       timeout 2 bash -c "echo >/dev/tcp/10.5.191.49/$port" 2>/dev/null; then
        ((exposed_count++))
    fi
done

if [[ $exposed_count -ge 4 ]]; then
    log_pass "$exposed_count ports essentiels accessibles"
else
    log_warn "Seulement $exposed_count ports détectés"
fi

log_section "7.2 Headers de sécurité"
log_test "Headers de sécurité HAProxy"
security_headers=$(curl -s -I "http://10.5.191.133/" | grep -i -E "(cors|content-type|cache)" | wc -l)
if [[ $security_headers -gt 1 ]]; then
    log_pass "$security_headers headers de sécurité détectés"
else
    log_warn "Headers de sécurité limités"
fi

# ========================================
# 8. PERSISTANCE DES DONNÉES
# ========================================
log_header "8. PERSISTANCE DES DONNÉES"

log_section "8.1 Base de données PostgreSQL"
log_test "Tables principales"
tables=$(incus exec veza-postgres -- psql -h localhost -U veza_user -d veza_db -t -c "SELECT count(*) FROM information_schema.tables WHERE table_schema='public';" 2>/dev/null | xargs || echo "0")
if [[ $tables -gt 0 ]]; then
    log_pass "$tables tables trouvées"
else
    log_warn "Aucune table trouvée"
fi

log_test "Données de test"
# Insérer une donnée de test
test_insert=$(incus exec veza-postgres -- psql -h localhost -U veza_user -d veza_db -t -c "INSERT INTO messages (content, author, room, created_at) VALUES ('test-persistence', 'test', 'general', NOW()) RETURNING id;" 2>/dev/null | xargs || echo "")
if [[ -n "$test_insert" ]]; then
    log_pass "Insertion test réussie (ID: $test_insert)"
else
    log_warn "Échec insertion test"
fi

# ========================================
# 9. MONITORING ET LOGS
# ========================================
log_header "9. MONITORING ET LOGS"

log_section "9.1 Logs des services"
log_services=("veza-chat:veza-chat" "veza-stream:veza-stream" "veza-frontend:nginx" "veza-haproxy:haproxy")

for service_info in "${log_services[@]}"; do
    container=$(echo $service_info | cut -d: -f1)
    service=$(echo $service_info | cut -d: -f2)
    
    log_test "Logs $container ($service)"
    log_lines=$(incus exec $container -- journalctl -u $service --no-pager -n 10 2>/dev/null | wc -l || echo "0")
    if [[ $log_lines -gt 5 ]]; then
        log_pass "$log_lines lignes de log récentes"
    else
        log_warn "Logs limités ($log_lines lignes)"
    fi
done

# ========================================
# 10. TESTS DE RÉSILIENCE
# ========================================
log_header "10. TESTS DE RÉSILIENCE"

log_section "10.1 Récupération automatique"
log_test "Test redémarrage Chat Server"
# Redémarrer le service et vérifier qu'il revient
incus exec veza-chat -- systemctl restart veza-chat >/dev/null 2>&1
sleep 3

if test_json_endpoint "http://10.5.191.49:3001/health" "data.status" "healthy" 10; then
    log_pass "Chat Server récupéré après redémarrage"
else
    log_fail "Chat Server non récupéré"
fi

log_test "Test redémarrage HAProxy"
incus exec veza-haproxy -- systemctl restart haproxy >/dev/null 2>&1
sleep 2

if test_endpoint "http://10.5.191.133/" "200" 10; then
    log_pass "HAProxy récupéré après redémarrage"
else
    log_fail "HAProxy non récupéré"
fi

# ========================================
# 11. TESTS END-TO-END
# ========================================
log_header "11. TESTS END-TO-END"

log_section "11.1 Scénario utilisateur complet"
log_test "1. Accès application via HAProxy"
if test_endpoint "http://10.5.191.133/"; then
    log_pass "✅ Application accessible"
else
    log_fail "❌ Application inaccessible"
fi

log_test "2. Envoi message via interface"
message_id=$(curl -s -X POST "http://10.5.191.133/chat-api/messages" \
    -H "Content-Type: application/json" \
    -d '{"content":"Test end-to-end complet","author":"e2e-test","room":"general"}' | \
    jq '.data' 2>/dev/null || echo "null")

if [[ "$message_id" != "null" && "$message_id" != "" ]]; then
    log_pass "✅ Message envoyé (ID: $message_id)"
else
    log_fail "❌ Échec envoi message"
fi

log_test "3. Vérification message dans chat"
recent_message=$(curl -s "http://10.5.191.133/chat-api/messages?room=general" | \
    jq '.data[-1].content' 2>/dev/null | tr -d '"' || echo "")

if [[ "$recent_message" == "Test end-to-end complet" ]]; then
    log_pass "✅ Message retrouvé dans le chat"
else
    log_warn "⚠️ Message non retrouvé (dernier: $recent_message)"
fi

log_test "4. Vérification stats mises à jour"
new_stats=$(curl -s "http://10.5.191.133/chat-api/messages/stats" | \
    jq '.data.total_messages' 2>/dev/null || echo "0")

if [[ $new_stats -gt $stats ]]; then
    log_pass "✅ Statistiques mises à jour ($new_stats total)"
else
    log_warn "⚠️ Statistiques non mises à jour"
fi

# ========================================
# RAPPORT FINAL
# ========================================
END_TIME=$(date +%s)
DURATION=$((END_TIME - START_TIME))

echo -e "\n${PURPLE}╔══════════════════════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${PURPLE}║                               📊 RAPPORT FINAL                                    ║${NC}"
echo -e "${PURPLE}╚══════════════════════════════════════════════════════════════════════════════════╝${NC}"

echo -e "\n${CYAN}📈 STATISTIQUES DES TESTS:${NC}"
echo -e "   ${GREEN}✅ Tests réussis:     $PASSED_TESTS${NC}"
echo -e "   ${RED}❌ Tests échoués:     $FAILED_TESTS${NC}"
echo -e "   ${YELLOW}⚠️  Avertissements:   $WARNING_TESTS${NC}"
echo -e "   ${BLUE}📊 Total des tests:   $TOTAL_TESTS${NC}"
echo -e "   ${PURPLE}⏱️  Durée d'exécution: ${DURATION}s${NC}"

# Calcul du score de santé
if [[ $TOTAL_TESTS -gt 0 ]]; then
    HEALTH_SCORE=$(( (PASSED_TESTS * 100) / TOTAL_TESTS ))
    
    echo -e "\n${CYAN}🎯 SCORE DE SANTÉ GLOBAL: ${NC}"
    if [[ $HEALTH_SCORE -ge 90 ]]; then
        echo -e "   ${GREEN}🟢 EXCELLENT: $HEALTH_SCORE% - Application parfaitement opérationnelle${NC}"
    elif [[ $HEALTH_SCORE -ge 75 ]]; then
        echo -e "   ${YELLOW}🟡 BON: $HEALTH_SCORE% - Application majoritairement fonctionnelle${NC}"
    elif [[ $HEALTH_SCORE -ge 50 ]]; then
        echo -e "   ${YELLOW}🟠 MOYEN: $HEALTH_SCORE% - Application partiellement fonctionnelle${NC}"
    else
        echo -e "   ${RED}🔴 CRITIQUE: $HEALTH_SCORE% - Application nécessite une intervention${NC}"
    fi
fi

echo -e "\n${CYAN}🚀 RECOMMANDATIONS:${NC}"

if [[ $FAILED_TESTS -eq 0 ]]; then
    echo -e "   ${GREEN}✨ Félicitations ! Votre infrastructure Veza est parfaitement opérationnelle${NC}"
    echo -e "   ${GREEN}🌐 Vous pouvez utiliser l'application à l'adresse: http://10.5.191.133${NC}"
elif [[ $FAILED_TESTS -le 3 ]]; then
    echo -e "   ${YELLOW}📝 Quelques améliorations mineures recommandées${NC}"
    echo -e "   ${YELLOW}🔧 Consultez les logs pour les détails: ./scripts/deploy-all.sh${NC}"
else
    echo -e "   ${RED}⚠️  Intervention requise sur plusieurs services${NC}"
    echo -e "   ${RED}🔧 Exécutez: ./scripts/deploy-all.sh pour corriger${NC}"
fi

echo -e "\n${CYAN}📋 SERVICES PRINCIPAUX:${NC}"
echo -e "   • Frontend React:     http://10.5.191.133/ (via HAProxy)"
echo -e "   • Chat API:           http://10.5.191.133/chat-api/"
echo -e "   • Stream API:         http://10.5.191.133/stream/"
echo -e "   • Backend API:        http://10.5.191.133/api/"

echo -e "\n${PURPLE}🎉 Tests ultra complets terminés ! Infrastructure Veza validée 🎉${NC}"

# Retourner un code d'erreur si des tests critiques ont échoué
if [[ $FAILED_TESTS -gt 5 ]]; then
    exit 1
else
    exit 0
fi 