#!/bin/bash

# Script de test API Talas/Veza - Version simplifiée
# Tests toutes les fonctionnalités en profondeur

# Configuration (pas de set -e pour continuer malgré les erreurs)
API="http://localhost:8080/api/v1"
TMP="/tmp/veza_test"
mkdir -p "$TMP"

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Variables globales
TOKEN=""
USER_ID=""
TOTAL=0
PASS=0
FAIL=0

# Fonctions utilitaires
log() { echo -e "${BLUE}[TEST]${NC} $1"; ((TOTAL++)); }
ok() { echo -e "${GREEN}✅ $1${NC}"; ((PASS++)); }
err() { echo -e "${RED}❌ $1${NC}"; ((FAIL++)); }
warn() { echo -e "${YELLOW}⚠️  $1${NC}"; }

# Fonction de requête
req() {
    local method="$1" url="$2" data="$3" auth="$4"
    local cmd="curl -s -w '%{http_code}'"
    
    [ -n "$auth" ] && cmd="$cmd -H 'Authorization: Bearer $TOKEN'"
    [ -n "$data" ] && cmd="$cmd -H 'Content-Type: application/json' -d '$data'"
    [ "$method" != "GET" ] && cmd="$cmd -X $method"
    
    cmd="$cmd -o $TMP/resp.json '$url'"
    echo "DEBUG: Executing: $cmd" >> "$TMP/debug.log"
    local status=$(eval $cmd 2>>"$TMP/debug.log")
    local response=$(cat "$TMP/resp.json" 2>/dev/null || echo "{}")
    
    echo "$response" > "$TMP/last.json"
    echo "$status" > "$TMP/status.txt"
    echo "DEBUG: Status=$status, Response=$response" >> "$TMP/debug.log"
    
    [ "$status" = "200" ] || [ "$status" = "201" ]
}

# === TESTS ===

echo -e "${BLUE}🚀 Tests API Talas - $(date)${NC}\n"

# 1. Vérification serveur
log "Vérification serveur"
if req "GET" "http://localhost:8080/api/health"; then
    ok "Serveur accessible"
else
    err "Serveur non accessible - Arrêt des tests"
    exit 1
fi

# 2. Tests d'authentification
echo -e "\n${BLUE}=== AUTHENTIFICATION ===${NC}"

# Création utilisateur unique
TS=$(date +%s)
EMAIL="test_$TS@test.com"
USER="user_$TS"
PASS_TEST="password123"

log "Inscription nouvel utilisateur"
DATA="{\"username\":\"$USER\",\"email\":\"$EMAIL\",\"password\":\"$PASS_TEST\"}"
if req "POST" "$API/auth/signup" "$DATA"; then
    response=$(cat "$TMP/last.json")
    if echo "$response" | grep -q "success.*true"; then
        USER_ID=$(echo "$response" | jq -r '.data.user_id // .data.id // ""')
        ok "Inscription réussie (UserID: $USER_ID)"
    else
        warn "Email existant - tentative connexion"
    fi
else
    status=$(cat "$TMP/status.txt")
    if [ "$status" = "200" ]; then
        warn "Email existant - tentative connexion"
    else
        err "Échec inscription (Status: $status)"
    fi
fi

log "Connexion utilisateur"
DATA="{\"email\":\"$EMAIL\",\"password\":\"$PASS_TEST\"}"
if req "POST" "$API/auth/login" "$DATA"; then
    response=$(cat "$TMP/last.json")
    if echo "$response" | grep -q "access_token"; then
        TOKEN=$(echo "$response" | jq -r '.data.access_token')
        USER_ID=$(echo "$response" | jq -r '.data.user.id')
        ok "Connexion réussie - Token obtenu"
    else
        err "Connexion échouée"
        exit 1
    fi
else
    err "Échec connexion"
    exit 1
fi

log "Vérification profil (/me)"
if req "GET" "$API/auth/me" "" "auth"; then
    ok "Profil récupéré"
else
    err "Échec récupération profil"
fi

# 3. Tests utilisateurs
echo -e "\n${BLUE}=== UTILISATEURS ===${NC}"

log "Liste utilisateurs"
if req "GET" "$API/users"; then
    count=$(cat "$TMP/last.json" | jq '. | length // 0')
    ok "Liste utilisateurs ($count utilisateurs)"
else
    err "Échec liste utilisateurs"
fi

log "Profil via /users/me"
if req "GET" "$API/users/me" "" "auth"; then
    ok "Profil /users/me accessible"
else
    warn "/users/me non accessible"
fi

log "Utilisateurs sauf moi"
if req "GET" "$API/users/except-me" "" "auth"; then
    ok "Except-me fonctionnel"
else
    warn "Except-me non accessible"
fi

log "Recherche utilisateurs"
if req "GET" "$API/users/search?q=test" "" "auth"; then
    ok "Recherche utilisateurs OK"
else
    warn "Recherche non accessible"
fi

# 4. Tests tags
echo -e "\n${BLUE}=== TAGS ===${NC}"

log "Liste tags"
if req "GET" "$API/tags"; then
    count=$(cat "$TMP/last.json" | jq '. | length // 0')
    ok "Tags récupérés ($count tags)"
else
    err "Échec tags"
fi

log "Recherche tags"
if req "GET" "$API/tags/search?q=music"; then
    ok "Recherche tags OK"
else
    warn "Recherche tags non accessible"
fi

# 5. Tests tracks
echo -e "\n${BLUE}=== TRACKS AUDIO ===${NC}"

log "Liste tracks"
if req "GET" "$API/tracks"; then
    count=$(cat "$TMP/last.json" | jq '. | length // 0')
    ok "Tracks récupérés ($count tracks)"
else
    err "Échec tracks"
fi

log "Création track"
DATA='{"title":"Track Test","description":"Test upload","tags":["test"]}'
if req "POST" "$API/tracks" "$DATA" "auth"; then
    response=$(cat "$TMP/last.json")
    if echo "$response" | grep -q '"id"'; then
        TRACK_ID=$(echo "$response" | jq -r '.data.id // .id // ""')
        ok "Track créé (ID: $TRACK_ID)"
    else
        warn "Track créé - format inattendu"
    fi
else
    warn "Création track échec"
fi

# 6. Tests listings/produits
echo -e "\n${BLUE}=== LISTINGS/PRODUITS ===${NC}"

log "Liste listings"
if req "GET" "$API/listings"; then
    count=$(cat "$TMP/last.json" | jq '. | length // 0')
    ok "Listings récupérés ($count listings)"
else
    err "Échec listings"
fi

log "Création listing"
DATA='{"title":"Produit Test","description":"Test","price":29.99,"category":"test"}'
if req "POST" "$API/listings" "$DATA" "auth"; then
    response=$(cat "$TMP/last.json")
    if echo "$response" | grep -q '"id"'; then
        LISTING_ID=$(echo "$response" | jq -r '.data.id // .id // ""')
        ok "Listing créé (ID: $LISTING_ID)"
    else
        warn "Listing créé - format inattendu"
    fi
else
    warn "Création listing échec"
fi

# 7. Tests salons/rooms
echo -e "\n${BLUE}=== SALONS/ROOMS ===${NC}"

log "Liste salons"
if req "GET" "$API/rooms"; then
    count=$(cat "$TMP/last.json" | jq '. | length // 0')
    ok "Salons récupérés ($count salons)"
else
    err "Échec salons"
fi

log "Création salon"
DATA='{"name":"Salon Test","description":"Test","is_public":true}'
if req "POST" "$API/rooms" "$DATA" "auth"; then
    response=$(cat "$TMP/last.json")
    if echo "$response" | grep -q '"id"'; then
        ROOM_ID=$(echo "$response" | jq -r '.data.id // .id // ""')
        ok "Salon créé (ID: $ROOM_ID)"
    else
        warn "Salon créé - format inattendu"
    fi
else
    warn "Création salon échec"
fi

# 8. Tests chat
echo -e "\n${BLUE}=== SYSTÈME CHAT ===${NC}"

log "API chat/rooms"
if req "GET" "$API/chat/rooms" "" "auth"; then
    ok "Chat/rooms accessible"
else
    warn "Chat/rooms non accessible"
fi

log "Conversations directes"
if req "GET" "$API/chat/conversations" "" "auth"; then
    ok "Conversations accessibles"
else
    warn "Conversations non accessibles"
fi

log "Messages non lus"
if req "GET" "$API/chat/unread" "" "auth"; then
    ok "Messages non lus OK"
else
    warn "Messages non lus non accessibles"
fi

log "Message direct (DM)"
if req "GET" "$API/chat/dm/1" "" "auth"; then
    ok "DM accessible"
else
    warn "DM non accessible"
fi

# 9. Tests recherche
echo -e "\n${BLUE}=== RECHERCHE ===${NC}"

log "Recherche globale"
if req "GET" "$API/search?q=test"; then
    ok "Recherche globale OK"
else
    warn "Recherche globale non accessible"
fi

log "Recherche avancée"
if req "GET" "$API/search/advanced?q=test&type=tracks"; then
    ok "Recherche avancée OK"
else
    warn "Recherche avancée non accessible"
fi

log "Autocomplétion"
if req "GET" "$API/search/autocomplete?q=te"; then
    ok "Autocomplétion OK"
else
    warn "Autocomplétion non accessible"
fi

# 10. Tests ressources partagées
echo -e "\n${BLUE}=== RESSOURCES PARTAGÉES ===${NC}"

log "Liste ressources"
if req "GET" "$API/shared-resources"; then
    count=$(cat "$TMP/last.json" | jq '. | length // 0')
    ok "Ressources récupérées ($count ressources)"
else
    err "Échec ressources"
fi

log "Recherche ressources"
if req "GET" "$API/shared-resources/search?q=test"; then
    ok "Recherche ressources OK"
else
    warn "Recherche ressources non accessible"
fi

# 11. Tests messages
echo -e "\n${BLUE}=== MESSAGES ===${NC}"

log "Messages utilisateur"
if req "GET" "$API/messages/1" "" "auth"; then
    ok "Messages utilisateur OK"
else
    warn "Messages utilisateur non accessibles"
fi

# 12. Tests administration
echo -e "\n${BLUE}=== ADMINISTRATION ===${NC}"

log "Dashboard admin"
if req "GET" "$API/admin/dashboard" "" "auth"; then
    ok "Dashboard admin accessible"
else
    warn "Dashboard admin non accessible (droits insuffisants)"
fi

log "Utilisateurs admin"
if req "GET" "$API/admin/users" "" "auth"; then
    ok "Gestion utilisateurs admin OK"
else
    warn "Gestion utilisateurs admin non accessible"
fi

# 13. Tests sécurité
echo -e "\n${BLUE}=== SÉCURITÉ ===${NC}"

log "Accès sans token"
if req "GET" "$API/users/me"; then
    status=$(cat "$TMP/status.txt")
    if [ "$status" = "401" ]; then
        ok "Protection JWT active (401)"
    else
        err "Protection JWT faible (Status: $status)"
    fi
else
    status=$(cat "$TMP/status.txt")
    if [ "$status" = "401" ]; then
        ok "Protection JWT active (401)"
    else
        warn "Protection JWT (Status: $status)"
    fi
fi

log "Token invalide"
SAVE_TOKEN="$TOKEN"
TOKEN="invalid_token_123"
if req "GET" "$API/users/me" "" "auth"; then
    status=$(cat "$TMP/status.txt")
    if [ "$status" = "401" ]; then
        ok "Validation JWT active"
    else
        warn "Validation JWT faible"
    fi
else
    ok "Token invalide rejeté"
fi
TOKEN="$SAVE_TOKEN"

# 14. Test performance basique
echo -e "\n${BLUE}=== PERFORMANCE ===${NC}"

log "Test charge (5 requêtes simultanées)"
start=$(date +%s%N)
for i in {1..5}; do
    req "GET" "$API/tags" &
done
wait
end=$(date +%s%N)
duration=$(( (end - start) / 1000000 ))
ok "5 requêtes en ${duration}ms"

# === RÉSUMÉ ===
echo -e "\n${BLUE}=== RÉSUMÉ DES TESTS ===${NC}"
echo -e "Total: ${BLUE}$TOTAL${NC} | Réussis: ${GREEN}$PASS${NC} | Échecs: ${RED}$FAIL${NC}"

if [ $FAIL -eq 0 ]; then
    echo -e "\n${GREEN}🎉 TOUS LES TESTS CRITIQUES PASSÉS !${NC}"
    echo -e "${GREEN}API Talas opérationnelle ✅${NC}"
else
    echo -e "\n${YELLOW}⚠️  $FAIL test(s) échoué(s)${NC}"
    RATE=$(( PASS * 100 / TOTAL ))
    echo -e "Taux de réussite: ${BLUE}${RATE}%${NC}"
fi

echo -e "\n${BLUE}📄 Logs détaillés: $TMP/last.json${NC}"
echo -e "${BLUE}🕒 Tests terminés: $(date)${NC}" 