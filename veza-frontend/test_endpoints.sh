#!/bin/bash

# Configuration
BASE_URL="http://localhost:8080"
API_URL="$BASE_URL/api/v1"

# Couleurs pour les messages
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Fonction pour afficher les résultats
print_result() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✓ $2${NC}"
    else
        echo -e "${RED}✗ $2${NC}"
    fi
}

# Fonction pour tester un endpoint
test_endpoint() {
    local method=$1
    local endpoint=$2
    local data=$3
    local description=$4
    local token=$5

    echo "Testing $description..."
    
    if [ -n "$token" ]; then
        response=$(curl -s -X $method \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $token" \
            -d "$data" \
            "$API_URL$endpoint")
    else
        response=$(curl -s -X $method \
            -H "Content-Type: application/json" \
            -d "$data" \
            "$API_URL$endpoint")
    fi

    status=$?
    print_result $status "$description"
    echo "Response: $response"
    echo "----------------------------------------"
}

# 1. Test des endpoints d'authentification
echo "Testing Authentication Endpoints..."

# 1.1 Inscription
test_endpoint "POST" "/auth/register" \
    '{"username":"koubou","email":"koubou@example.com","password":"testpass123"}' \
    "Register new user"

# 1.2 Connexion
login_response=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    -d '{"email":"koubou@example.com","password":"testpass123"}' \
    "$API_URL/auth/login")

echo "Login response: $login_response"

# Extraire le token d'accès
access_token=$(echo $login_response | jq -r '.data.access_token')

if [ -z "$access_token" ] || [ "$access_token" = "null" ]; then
    echo -e "${RED}Failed to get access token. Response was: $login_response${NC}"
    exit 1
fi

echo "Got access token: $access_token"
echo "----------------------------------------"

# 2. Test des endpoints utilisateur
echo "Testing User Endpoints..."

test_endpoint "GET" "/users/me" "" "Get current user profile" "$access_token"
test_endpoint "GET" "/users" "" "Get all users" "$access_token"
test_endpoint "GET" "/users/except-me" "" "Get users except me" "$access_token"
test_endpoint "GET" "/users/search?q=test" "" "Search users" "$access_token"

# 3. Test des endpoints de chat
echo "Testing Chat Endpoints..."

test_endpoint "GET" "/chat/rooms" "" "Get public rooms" "$access_token"
test_endpoint "POST" "/chat/rooms" \
    '{"name":"Test Room","description":"A test room","is_private":false}' \
    "Create new room" "$access_token"

# 4. Test des endpoints de messages
echo "Testing Message Endpoints..."

test_endpoint "GET" "/chat/dm/1" "" "Get direct messages with user 1" "$access_token"
test_endpoint "GET" "/chat/rooms/1/messages" "" "Get room messages" "$access_token"

# 5. Test des endpoints de recherche
echo "Testing Search Endpoints..."

test_endpoint "GET" "/search?q=test" "" "Global search" "$access_token"
test_endpoint "GET" "/search/advanced?q=test" "" "Advanced search" "$access_token"
test_endpoint "GET" "/search/autocomplete?q=test" "" "Autocomplete search" "$access_token"

# 6. Test des endpoints de tags
echo "Testing Tag Endpoints..."

test_endpoint "GET" "/tags" "" "Get all tags" "$access_token"
test_endpoint "GET" "/tags/search?q=test" "" "Search tags" "$access_token"

# 7. Test des endpoints de ressources partagées
echo "Testing Shared Resources Endpoints..."

test_endpoint "GET" "/shared-resources" "" "Get all shared resources" "$access_token"
test_endpoint "GET" "/shared-resources/search?q=test" "" "Search shared resources" "$access_token"

# 8. Test des endpoints de tracks
echo "Testing Track Endpoints..."

test_endpoint "GET" "/tracks" "" "Get all tracks" "$access_token"
test_endpoint "GET" "/tracks/1" "" "Get specific track" "$access_token"

# 9. Test des endpoints de listings
echo "Testing Listing Endpoints..."

test_endpoint "GET" "/listings" "" "Get all listings" "$access_token"
test_endpoint "GET" "/listings/1" "" "Get specific listing" "$access_token"

# 10. Test des endpoints d'offres
echo "Testing Offer Endpoints..."

test_endpoint "GET" "/offers" "" "Get all offers" "$access_token"

# 11. Test des endpoints admin
echo "Testing Admin Endpoints..."

test_endpoint "GET" "/admin/dashboard" "" "Get admin dashboard" "$access_token"
test_endpoint "GET" "/admin/users" "" "Get admin users" "$access_token"
test_endpoint "GET" "/admin/analytics" "" "Get admin analytics" "$access_token"
test_endpoint "GET" "/admin/categories" "" "Get admin categories" "$access_token"

echo "All tests completed!" 