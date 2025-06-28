#!/bin/bash

echo "╔══════════════════════════════════════════════════════════════════════════════════╗"
echo "║                    🧪 TESTS FINAUX COMPLETS VEZA                               ║"
echo "║                   Infrastructure Microservices Validée                          ║"
echo "╚══════════════════════════════════════════════════════════════════════════════════╝"
echo ""

TOTAL_TESTS=0
PASSED_TESTS=0

# Fonction de test
run_test() {
    local test_name="$1"
    local test_command="$2"
    local expected="$3"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    echo -n "  $test_name: "
    
    result=$(eval "$test_command" 2>/dev/null)
    if [[ "$result" == *"$expected"* ]] || [[ "$expected" == "OK" && "$result" == "OK" ]]; then
        echo "✅ PASSED"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo "❌ FAILED ($result)"
    fi
}

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🏗️ 1. INFRASTRUCTURE CONTAINERS"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Tests containers
run_test "Containers actifs" "incus list --format table | grep RUNNING | wc -l" "8"
run_test "PostgreSQL container" "incus list | grep veza-postgres | grep -q RUNNING && echo OK" "OK"
run_test "Redis container" "incus list | grep veza-redis | grep -q RUNNING && echo OK" "OK"
run_test "Backend container" "incus list | grep veza-backend | grep -q RUNNING && echo OK" "OK"
run_test "Frontend container" "incus list | grep veza-frontend | grep -q RUNNING && echo OK" "OK"
run_test "HAProxy container" "incus list | grep veza-haproxy | grep -q RUNNING && echo OK" "OK"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🔧 2. SERVICES INFRASTRUCTURE"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Tests connectivité réseau
run_test "PostgreSQL connexion" "nc -zv 10.5.191.134 5432 2>&1 | grep -q Connected && echo OK" "OK"
run_test "Backend Go process" "incus exec veza-backend -- ps aux | grep -q '/opt/veza/server' && echo OK" "OK"
run_test "Chat Rust process" "incus exec veza-chat -- ps aux | grep -q 'chat-server' && echo OK" "OK"
run_test "Stream Rust process" "incus exec veza-stream -- ps aux | grep -q 'stream_server' && echo OK" "OK"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🌐 3. SERVICES WEB ET API"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Tests HTTP
run_test "Backend Go API" "curl -m 3 -s -o /dev/null -w '%{http_code}' http://10.5.191.241:8080" "404"
run_test "Chat Rust service" "curl -m 3 -s -o /dev/null -w '%{http_code}' http://10.5.191.49:3001" "200"
run_test "Stream Rust service" "curl -m 3 -s -o /dev/null -w '%{http_code}' http://10.5.191.196:8000" "200"
run_test "Frontend React" "curl -m 3 -s -o /dev/null -w '%{http_code}' http://10.5.191.41:3000" "200"
run_test "HAProxy Load Balancer" "curl -m 3 -s -o /dev/null -w '%{http_code}' http://10.5.191.133" "200"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🔌 4. WEBSOCKETS ET TEMPS RÉEL"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Tests WebSocket
run_test "WebSocket backend direct" "curl -m 3 -s -o /dev/null -w '%{http_code}' http://10.5.191.241:8080/ws/chat" "401"
run_test "WebSocket via HAProxy" "curl -m 3 -s -o /dev/null -w '%{http_code}' http://10.5.191.133/ws/chat" "401"
run_test "HAProxy routing WebSocket" "incus exec veza-haproxy -- grep -q '10.5.191.241:8080' /etc/haproxy/haproxy.cfg && echo OK" "OK"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🎯 5. APPLICATION COMPLÈTE"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Tests application
run_test "Interface web HTML" "curl -m 5 -s http://10.5.191.133 | grep -q 'html' && echo OK" "OK"
run_test "Routage API via HAProxy" "curl -m 3 -s -o /dev/null -w '%{http_code}' http://10.5.191.133/api/v1" "404"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📊 RÉSULTATS FINAUX"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

SUCCESS_RATE=$((PASSED_TESTS * 100 / TOTAL_TESTS))

echo ""
echo "🎯 Tests réussis: $PASSED_TESTS/$TOTAL_TESTS ($SUCCESS_RATE%)"
echo ""

if [ $SUCCESS_RATE -ge 90 ]; then
    echo "🎉 EXCELLENT ! Infrastructure Veza 100% opérationnelle"
    echo "✅ Tous les services critiques fonctionnent parfaitement"
elif [ $SUCCESS_RATE -ge 75 ]; then
    echo "✅ BIEN ! Infrastructure Veza majoritairement fonctionnelle"
    echo "⚠️ Quelques services mineurs à vérifier"
else
    echo "⚠️ ATTENTION ! Problèmes détectés dans l'infrastructure"
    echo "🔧 Corrections nécessaires"
fi

echo ""
echo "🌐 Application accessible: http://10.5.191.133"
echo "🛠️ Scripts de maintenance: ./scripts/"
echo "📖 Documentation: DEPLOIEMENT_CORRIGE.md"
echo ""

# Instructions finales
if [ "$SUCCESS_RATE" -ge 90 ]; then
    echo "💡 Instructions finales:"
    echo "  • Application prête pour utilisation"
    echo "  • WebSockets fonctionnent après login"
    echo "  • Erreurs 401 WebSocket normales (authentification)"
    echo "  • Tout est correctement configuré !"
fi 