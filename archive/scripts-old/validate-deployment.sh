#!/bin/bash

# Script de validation rapide du déploiement Veza
set -e

echo "🔍 Validation du déploiement Veza"

# Couleurs
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

success=0
total=0

# Fonction de test
test_item() {
    local name="$1"
    local command="$2"
    ((total++))
    
    echo -n "Testing $name... "
    if eval "$command" >/dev/null 2>&1; then
        echo -e "${GREEN}✅${NC}"
        ((success++))
    else
        echo -e "${RED}❌${NC}"
    fi
}

echo ""
echo "📦 Vérification des containers:"
test_item "veza-postgres" "incus list veza-postgres -c s --format csv | grep -q RUNNING"
test_item "veza-redis" "incus list veza-redis -c s --format csv | grep -q RUNNING"
test_item "veza-backend" "incus list veza-backend -c s --format csv | grep -q RUNNING"
test_item "veza-chat" "incus list veza-chat -c s --format csv | grep -q RUNNING"
test_item "veza-stream" "incus list veza-stream -c s --format csv | grep -q RUNNING"
test_item "veza-frontend" "incus list veza-frontend -c s --format csv | grep -q RUNNING"
test_item "veza-haproxy" "incus list veza-haproxy -c s --format csv | grep -q RUNNING"

echo ""
echo "🔗 Vérification des services:"
test_item "PostgreSQL" "incus exec veza-postgres -- pg_isready -h localhost -p 5432"
test_item "Redis" "incus exec veza-redis -- redis-cli ping | grep -q PONG"

echo ""
echo "🌐 Vérification des endpoints:"
test_item "Frontend" "curl -s -o /dev/null http://10.5.191.41:5173"
test_item "Backend" "curl -s -o /dev/null http://10.5.191.241:8080"
test_item "Chat Server" "curl -s -o /dev/null http://10.5.191.49:8081"
test_item "Stream Server" "curl -s -o /dev/null http://10.5.191.196:8082"
test_item "HAProxy" "curl -s -o /dev/null http://10.5.191.133"

echo ""
echo "📊 Résultats:"
echo -e "Tests réussis: ${GREEN}$success${NC}/$total"

if [[ $success -eq $total ]]; then
    echo -e "${GREEN}🎉 Tous les tests sont passés ! Déploiement validé.${NC}"
    echo ""
    echo -e "${BLUE}🌐 Accès à l'application:${NC}"
    echo -e "  • Principal: ${GREEN}http://10.5.191.133${NC}"
    echo -e "  • Frontend: ${GREEN}http://10.5.191.41:5173${NC}"
    echo -e "  • Backend API: ${GREEN}http://10.5.191.241:8080${NC}"
elif [[ $success -gt $((total/2)) ]]; then
    echo -e "${YELLOW}⚠️ Déploiement partiellement fonctionnel ($success/$total)${NC}"
    echo "Certains services peuvent prendre plus de temps à démarrer."
else
    echo -e "${RED}❌ Problèmes majeurs détectés ($success/$total)${NC}"
    echo "Vérifiez les logs avec: ./scripts/incus-logs.sh [container]"
fi

echo ""
echo -e "${BLUE}💡 Commandes utiles:${NC}"
echo -e "  • Logs: ${YELLOW}./scripts/incus-logs.sh [container]${NC}"
echo -e "  • Statut: ${YELLOW}./scripts/incus-status.sh${NC}"
echo -e "  • Tests complets: ${YELLOW}./scripts/test-complete.sh${NC}" 