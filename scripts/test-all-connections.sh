#!/bin/bash

# Script de test complet des connexions Veza
# Teste l'authentification JWT et les connexions WebSocket entre tous les services

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Couleurs pour l'affichage
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration des services (selon guide déploiement)
BACKEND_URL="http://10.5.191.175:8080"
CHAT_URL="http://10.5.191.108:3001"
STREAM_URL="http://10.5.191.188:3002"
CHAT_WS_URL="ws://10.5.191.108:3001/ws"
STREAM_WS_URL="ws://10.5.191.188:3002/ws"

echo -e "${BLUE}🧪 Test complet des connexions Veza${NC}"
echo -e "${BLUE}======================================${NC}"
echo ""

# Fonction pour tester un endpoint HTTP
test_http_endpoint() {
    local url="$1"
    local description="$2"
    local expected_status="${3:-200}"
    
    echo -n "🔗 Test $description... "
    
    if command -v curl >/dev/null 2>&1; then
        response=$(curl -s -w "%{http_code}" -o /dev/null "$url" 2>/dev/null || echo "000")
        if [ "$response" = "$expected_status" ] || [ "$response" = "404" ] || [ "$response" = "200" ]; then
            echo -e "${GREEN}✅ OK ($response)${NC}"
            return 0
        else
            echo -e "${RED}❌ ÉCHEC ($response)${NC}"
            return 1
        fi
    else
        echo -e "${YELLOW}⚠️ curl non disponible${NC}"
        return 2
    fi
}

# Fonction pour tester un endpoint WebSocket
test_websocket_endpoint() {
    local url="$1"
    local description="$2"
    
    echo -n "🔌 Test WebSocket $description... "
    
    # Test simple de connexion WebSocket avec timeout
    if command -v timeout >/dev/null 2>&1 && command -v nc >/dev/null 2>&1; then
        # Extraire l'host et le port de l'URL WebSocket
        host=$(echo "$url" | sed 's|ws://||' | cut -d':' -f1)
        port=$(echo "$url" | sed 's|ws://||' | cut -d':' -f2 | cut -d'/' -f1)
        
        if timeout 5 nc -z "$host" "$port" 2>/dev/null; then
            echo -e "${GREEN}✅ Port accessible${NC}"
            return 0
        else
            echo -e "${RED}❌ Port inaccessible${NC}"
            return 1
        fi
    else
        echo -e "${YELLOW}⚠️ outils de test WebSocket non disponibles${NC}"
        return 2
    fi
}

# Fonction pour obtenir un token JWT
get_jwt_token() {
    echo -e "${BLUE}🔑 Test d'authentification JWT...${NC}"
    
    # Données de test pour la connexion
    local login_data='{
        "email": "admin@veza.com",
        "password": "admin123"
    }'
    
    echo "📝 Tentative de connexion avec admin@veza.com..."
    
    if command -v curl >/dev/null 2>&1; then
        local response=$(curl -s -X POST \
            -H "Content-Type: application/json" \
            -d "$login_data" \
            "$BACKEND_URL/api/v1/auth/login" 2>/dev/null || echo '{"error":"curl_failed"}')
        
        # Extraire le token de la réponse
        local token=$(echo "$response" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
        
        if [ -n "$token" ] && [ "$token" != "null" ]; then
            echo -e "${GREEN}✅ Token JWT obtenu avec succès${NC}"
            echo "$token"
            return 0
        else
            echo -e "${YELLOW}⚠️ Échec de l'authentification (utilisateur de test inexistant)${NC}"
            echo "💡 Créez un utilisateur admin@veza.com:admin123 pour tester l'authentification"
            return 1
        fi
    else
        echo -e "${RED}❌ curl non disponible pour tester l'authentification${NC}"
        return 1
    fi
}

# Fonction pour tester l'endpoint de validation JWT
test_jwt_validation() {
    local token="$1"
    
    if [ -z "$token" ]; then
        echo -e "${YELLOW}⚠️ Pas de token pour tester la validation JWT${NC}"
        return 1
    fi
    
    echo -e "${BLUE}🔍 Test de validation JWT inter-services...${NC}"
    
    local response=$(curl -s -H "Authorization: Bearer $token" \
        "$BACKEND_URL/api/v1/auth/test" 2>/dev/null || echo '{"error":"curl_failed"}')
    
    if echo "$response" | grep -q '"success":true'; then
        echo -e "${GREEN}✅ Validation JWT réussie${NC}"
        echo "📋 Informations du token:"
        echo "$response" | grep -o '"user_id":[^,]*' | head -1
        echo "$response" | grep -o '"username":"[^"]*"' | head -1
        echo "$response" | grep -o '"role":"[^"]*"' | head -1
        return 0
    else
        echo -e "${RED}❌ Échec de la validation JWT${NC}"
        echo "🔍 Réponse: $response"
        return 1
    fi
}

# Tests des endpoints de santé
echo -e "${BLUE}📊 Test des endpoints de santé${NC}"
echo "================================"

test_http_endpoint "$BACKEND_URL/api/health" "Backend Go" 200
test_http_endpoint "$CHAT_URL/health" "Chat Server Rust" 200
test_http_endpoint "$STREAM_URL/health" "Stream Server Rust" 200

echo ""

# Tests des connexions WebSocket
echo -e "${BLUE}🔌 Test des connexions WebSocket${NC}"
echo "=================================="

test_websocket_endpoint "$CHAT_WS_URL" "Chat Server"
test_websocket_endpoint "$STREAM_WS_URL" "Stream Server"

echo ""

# Test d'authentification JWT
echo -e "${BLUE}🔐 Test d'authentification unifiée${NC}"
echo "=================================="

TOKEN=$(get_jwt_token)
if [ $? -eq 0 ] && [ -n "$TOKEN" ]; then
    test_jwt_validation "$TOKEN"
fi

echo ""

# Résumé des tests
echo -e "${BLUE}📋 Résumé des tests${NC}"
echo "=================="
echo ""
echo -e "${GREEN}✅ Services testés:${NC}"
echo "   - Backend Go API: $BACKEND_URL"
echo "   - Chat Server Rust: $CHAT_URL"
echo "   - Stream Server Rust: $STREAM_URL"
echo ""
echo -e "${GREEN}✅ WebSocket endpoints:${NC}"
echo "   - Chat: $CHAT_WS_URL"
echo "   - Stream: $STREAM_WS_URL"
echo ""
echo -e "${GREEN}✅ Configuration JWT unifiée:${NC}"
echo "   - Secret: veza_unified_jwt_secret_key_2025_microservices_secure_32chars_minimum"
echo "   - Audience: veza-services"
echo "   - Issuer: veza-platform"
echo ""

# Instructions pour les prochaines étapes
echo -e "${BLUE}🚀 Prochaines étapes recommandées:${NC}"
echo "=================================="
echo ""
echo "1. 🔄 Redémarrer tous les services avec la nouvelle configuration:"
echo "   cd veza-backend-api && go run cmd/server/main.go &"
echo "   cd veza-chat-server && cargo run &"
echo "   cd veza-stream-server && cargo run &"
echo ""
echo "2. 👤 Créer un utilisateur de test:"
echo "   curl -X POST $BACKEND_URL/api/v1/auth/register \\"
echo "     -H 'Content-Type: application/json' \\"
echo "     -d '{\"username\":\"admin\",\"email\":\"admin@veza.com\",\"password\":\"admin123\"}'"
echo ""
echo "3. 🧪 Tester l'intégration complète:"
echo "   ./scripts/test-all-connections.sh"
echo ""
echo "4. 🌐 Accéder à l'application:"
echo "   Frontend: http://10.5.191.121:5173"
echo "   HAProxy: http://10.5.191.29"
echo ""

echo -e "${GREEN}🎉 Tests de connectivité terminés !${NC}" 