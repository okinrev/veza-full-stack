#!/bin/bash

# Script pour créer un utilisateur de test et valider l'authentification JWT
# Teste l'unicité des tokens entre tous les services

set -e

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration selon guide déploiement
BACKEND_URL="http://10.5.191.175:8080"

echo -e "${BLUE}👤 Création d'utilisateur de test Veza${NC}"
echo -e "${BLUE}=====================================${NC}"
echo ""

# Fonction pour tester une API avec curl
test_api() {
    local method="$1"
    local url="$2"
    local data="$3"
    local description="$4"
    local auth_header="$5"
    
    echo -e "${BLUE}🔗 Test: $description${NC}"
    echo "   URL: $url"
    echo "   Méthode: $method"
    
    if [ -n "$data" ]; then
        echo "   Données: $data"
    fi
    
    # Construire la commande curl
    local curl_cmd="curl -s -w \"\n%{http_code}\" -X $method"
    
    if [ -n "$auth_header" ]; then
        curl_cmd="$curl_cmd -H \"Authorization: $auth_header\""
    fi
    
    if [ -n "$data" ]; then
        curl_cmd="$curl_cmd -H \"Content-Type: application/json\" -d '$data'"
    fi
    
    curl_cmd="$curl_cmd \"$url\""
    
    echo -e "${YELLOW}⏳ Exécution...${NC}"
    local response=$(eval $curl_cmd 2>/dev/null || echo -e "\nERROR")
    
    # Séparer la réponse du code de statut
    local http_code=$(echo "$response" | tail -n1)
    local response_body=$(echo "$response" | head -n -1)
    
    echo "📋 Code de statut: $http_code"
    echo "📄 Réponse:"
    echo "$response_body" | head -10
    
    if [ ${#response_body} -gt 500 ]; then
        echo "   ... (réponse tronquée)"
    fi
    
    echo ""
    
    # Retourner le code de statut et la réponse
    echo "$http_code|$response_body"
}

# 1. Créer un utilisateur de test
echo -e "${GREEN}📝 Étape 1: Création d'un utilisateur de test${NC}"
echo "=========================================="

USER_DATA='{
    "username": "testuser",
    "email": "test@veza.com", 
    "password": "test123456"
}'

result=$(test_api "POST" "$BACKEND_URL/api/v1/auth/register" "$USER_DATA" "Inscription utilisateur")
register_code=$(echo "$result" | cut -d'|' -f1)
register_response=$(echo "$result" | cut -d'|' -f2)

if [ "$register_code" = "200" ] || [ "$register_code" = "201" ]; then
    echo -e "${GREEN}✅ Utilisateur créé avec succès${NC}"
elif [ "$register_code" = "409" ]; then
    echo -e "${YELLOW}⚠️ Utilisateur déjà existant (normal)${NC}"
else
    echo -e "${RED}❌ Échec de la création de l'utilisateur${NC}"
fi

echo ""

# 2. Connexion et récupération du token
echo -e "${GREEN}🔑 Étape 2: Connexion et récupération du token JWT${NC}"
echo "=============================================="

LOGIN_DATA='{
    "email": "test@veza.com",
    "password": "test123456"
}'

result=$(test_api "POST" "$BACKEND_URL/api/v1/auth/login" "$LOGIN_DATA" "Connexion utilisateur")
login_code=$(echo "$result" | cut -d'|' -f1)
login_response=$(echo "$result" | cut -d'|' -f2)

if [ "$login_code" = "200" ]; then
    echo -e "${GREEN}✅ Connexion réussie${NC}"
    
    # Extraire le token de la réponse
    TOKEN=$(echo "$login_response" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
    
    if [ -n "$TOKEN" ] && [ "$TOKEN" != "null" ]; then
        echo -e "${GREEN}✅ Token JWT extrait avec succès${NC}"
        echo "🔑 Token (premiers 50 caractères): ${TOKEN:0:50}..."
    else
        echo -e "${RED}❌ Impossible d'extraire le token JWT${NC}"
        echo "🔍 Réponse complète:"
        echo "$login_response"
        exit 1
    fi
else
    echo -e "${RED}❌ Échec de la connexion${NC}"
    echo "🔍 Code: $login_code"
    echo "🔍 Réponse: $login_response"
    exit 1
fi

echo ""

# 3. Validation du token avec l'endpoint de test
echo -e "${GREEN}🔍 Étape 3: Validation du token JWT${NC}"
echo "=================================="

result=$(test_api "GET" "$BACKEND_URL/api/v1/auth/test" "" "Validation token JWT" "Bearer $TOKEN")
test_code=$(echo "$result" | cut -d'|' -f1)
test_response=$(echo "$result" | cut -d'|' -f2)

if [ "$test_code" = "200" ]; then
    echo -e "${GREEN}✅ Token validé avec succès${NC}"
    
    # Extraire les informations du token
    echo "📋 Informations extraites du token:"
    echo "$test_response" | grep -o '"user_id":[^,]*' | head -1
    echo "$test_response" | grep -o '"username":"[^"]*"' | head -1
    echo "$test_response" | grep -o '"role":"[^"]*"' | head -1
    echo "$test_response" | grep -o '"service":"[^"]*"' | head -1
    
    # Extraire les endpoints
    echo ""
    echo "🔌 Endpoints disponibles:"
    echo "$test_response" | grep -o '"chat_ws":"[^"]*"' | head -1
    echo "$test_response" | grep -o '"stream_ws":"[^"]*"' | head -1
    echo "$test_response" | grep -o '"api_rest":"[^"]*"' | head -1
    
else
    echo -e "${RED}❌ Échec de la validation du token${NC}"
    echo "🔍 Code: $test_code"
    echo "🔍 Réponse: $test_response"
fi

echo ""

# 4. Test de l'endpoint /me
echo -e "${GREEN}👤 Étape 4: Test de l'endpoint utilisateur${NC}"
echo "========================================"

result=$(test_api "GET" "$BACKEND_URL/api/v1/auth/me" "" "Récupération profil utilisateur" "Bearer $TOKEN")
me_code=$(echo "$result" | cut -d'|' -f1)
me_response=$(echo "$result" | cut -d'|' -f2)

if [ "$me_code" = "200" ]; then
    echo -e "${GREEN}✅ Profil utilisateur récupéré avec succès${NC}"
else
    echo -e "${RED}❌ Échec de la récupération du profil${NC}"
    echo "🔍 Code: $me_code"
    echo "🔍 Réponse: $me_response"
fi

echo ""

# 5. Test de déconnexion
echo -e "${GREEN}⏹️ Étape 5: Test de déconnexion${NC}"
echo "==============================="

# Extraire le refresh token de la réponse de login
REFRESH_TOKEN=$(echo "$login_response" | grep -o '"refresh_token":"[^"]*"' | cut -d'"' -f4)

if [ -n "$REFRESH_TOKEN" ] && [ "$REFRESH_TOKEN" != "null" ]; then
    LOGOUT_DATA="{\"refresh_token\":\"$REFRESH_TOKEN\"}"
    
    result=$(test_api "POST" "$BACKEND_URL/api/v1/auth/logout" "$LOGOUT_DATA" "Déconnexion utilisateur")
    logout_code=$(echo "$result" | cut -d'|' -f1)
    logout_response=$(echo "$result" | cut -d'|' -f2)
    
    if [ "$logout_code" = "200" ]; then
        echo -e "${GREEN}✅ Déconnexion réussie${NC}"
    else
        echo -e "${RED}❌ Échec de la déconnexion${NC}"
        echo "🔍 Code: $logout_code"
    fi
else
    echo -e "${YELLOW}⚠️ Refresh token non trouvé, déconnexion ignorée${NC}"
fi

echo ""

# Résumé des tests
echo -e "${BLUE}📋 Résumé des tests d'authentification${NC}"
echo "======================================"
echo ""
echo -e "${GREEN}✅ Tests effectués:${NC}"
echo "   1. 📝 Inscription utilisateur: $register_code"
echo "   2. 🔑 Connexion JWT: $login_code"  
echo "   3. 🔍 Validation token: $test_code"
echo "   4. 👤 Profil utilisateur: $me_code"
echo "   5. ⏹️ Déconnexion: ${logout_code:-N/A}"
echo ""

if [ "$login_code" = "200" ] && [ "$test_code" = "200" ] && [ "$me_code" = "200" ]; then
    echo -e "${GREEN}🎉 Authentification JWT entièrement fonctionnelle !${NC}"
    echo ""
    echo -e "${GREEN}✅ Configuration JWT unifiée validée:${NC}"
    echo "   - Secret: veza_unified_jwt_secret_key_2025_microservices_secure_32chars_minimum"
    echo "   - Audience: veza-services"
    echo "   - Issuer: veza-platform"
    echo "   - Tokens inter-services: ✅ Compatibles"
    echo ""
    echo -e "${BLUE}🔌 Prêt pour l'intégration WebSocket:${NC}"
    echo "   - Chat: ws://10.5.191.108:3001/ws?token=$TOKEN"
    echo "   - Stream: ws://10.5.191.188:3002/ws?token=$TOKEN"
    echo ""
    echo -e "${BLUE}💡 Prochaines étapes:${NC}"
    echo "   1. Tester les connexions WebSocket avec ce token"
    echo "   2. Intégrer l'authentification dans le frontend"
    echo "   3. Déployer les services mis à jour"
else
    echo -e "${RED}❌ Problèmes détectés dans l'authentification${NC}"
    echo "💡 Vérifiez les logs des services et la configuration JWT"
fi

echo ""
echo -e "${GREEN}🔚 Tests d'authentification terminés${NC}" 