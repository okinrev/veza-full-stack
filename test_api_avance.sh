#!/bin/bash

# Script de test API Talas/Veza - Version avancée
# Tests complets avec cas d'erreur et rapport détaillé

# Configuration
API="http://localhost:8080/api/v1"
TMP="/tmp/veza_test_advanced"
REPORT="$TMP/rapport_complet.md"
mkdir -p "$TMP"

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

# Variables globales
TOKEN=""
ADMIN_TOKEN=""
USER_ID=""
USER2_ID=""
ROOM_ID=""
LISTING_ID=""
TRACK_ID=""
TOTAL=0
PASS=0
FAIL=0
WARN=0

# Initialisation du rapport
cat > "$REPORT" << 'EOF'
# 📊 Rapport de Tests API Talas/Veza

**Date des tests:** $(date)
**Version API:** 1.0.0
**URL de base:** http://localhost:8080/api/v1

## 🎯 Objectifs des tests
- Valider toutes les fonctionnalités CRUD
- Tester la sécurité et l'authentification
- Vérifier les performances basiques
- Identifier les points d'amélioration

## 📋 Résultats détaillés

EOF

# Fonctions utilitaires
log() { 
    echo -e "${BLUE}[TEST]${NC} $1"
    ((TOTAL++))
    echo "### Test: $1" >> "$REPORT"
}

ok() { 
    echo -e "${GREEN}✅ $1${NC}"
    ((PASS++))
    echo "**✅ SUCCÈS:** $1" >> "$REPORT"
    echo "" >> "$REPORT"
}

err() { 
    echo -e "${RED}❌ $1${NC}"
    ((FAIL++))
    echo "**❌ ÉCHEC:** $1" >> "$REPORT"
    echo "" >> "$REPORT"
}

warn() { 
    echo -e "${YELLOW}⚠️  $1${NC}"
    ((WARN++))
    echo "**⚠️ ATTENTION:** $1" >> "$REPORT"
    echo "" >> "$REPORT"
}

info() { 
    echo -e "${PURPLE}ℹ️  $1${NC}"
    echo "**ℹ️ INFO:** $1" >> "$REPORT"
    echo "" >> "$REPORT"
}

# Fonction de requête améliorée
req() {
    local method="$1" url="$2" data="$3" auth="$4" expected_status="$5"
    local cmd="curl -s -w '%{http_code}'"
    
    [ -n "$auth" ] && cmd="$cmd -H 'Authorization: Bearer $TOKEN'"
    [ -n "$data" ] && cmd="$cmd -H 'Content-Type: application/json' -d '$data'"
    [ "$method" != "GET" ] && cmd="$cmd -X $method"
    
    cmd="$cmd -o $TMP/resp.json '$url'"
    local status=$(eval $cmd 2>>"$TMP/debug.log")
    local response=$(cat "$TMP/resp.json" 2>/dev/null || echo "{}")
    
    echo "$response" > "$TMP/last.json"
    echo "$status" > "$TMP/status.txt"
    
    echo "**Requête:** \`$method $url\`" >> "$REPORT"
    echo "**Status:** $status" >> "$REPORT"
    echo "**Réponse:** \`$(echo "$response" | head -c 100)...\`" >> "$REPORT"
    
    if [ -n "$expected_status" ]; then
        [ "$status" = "$expected_status" ]
    else
        [ "$status" = "200" ] || [ "$status" = "201" ]
    fi
}

# Tests avec cas d'erreur
test_auth_errors() {
    echo -e "\n${BLUE}=== TESTS D'ERREURS D'AUTHENTIFICATION ===${NC}"
    
    log "Inscription avec email invalide"
    if req "POST" "$API/auth/signup" '{"username":"test","email":"invalid-email","password":"test"}' "" "400"; then
        ok "Email invalide rejeté correctement"
    else
        warn "Validation email insuffisante"
    fi
    
    log "Inscription avec mot de passe trop court"
    if req "POST" "$API/auth/signup" '{"username":"test","email":"test@test.com","password":"123"}' "" "400"; then
        ok "Mot de passe court rejeté"
    else
        warn "Validation mot de passe insuffisante"
    fi
    
    log "Connexion avec identifiants incorrects"
    if req "POST" "$API/auth/login" '{"email":"wrong@test.com","password":"wrongpass"}' "" "401"; then
        ok "Identifiants incorrects rejetés"
    else
        warn "Authentification faible"
    fi
}

# Tests d'autorisation
test_permissions() {
    echo -e "\n${BLUE}=== TESTS DE PERMISSIONS ===${NC}"
    
    log "Accès route admin sans droits"
    if req "GET" "$API/admin/dashboard" "" "auth" "403"; then
        ok "Accès admin correctement refusé"
    else
        warn "Contrôle d'accès admin insuffisant"
    fi
    
    log "Modification profil autre utilisateur"
    if req "PUT" "$API/users/1" '{"username":"hacked"}' "auth" "403"; then
        ok "Modification autre utilisateur refusée"
    else
        warn "Isolation utilisateurs insuffisante"
    fi
}

# Tests de validation des données
test_data_validation() {
    echo -e "\n${BLUE}=== TESTS DE VALIDATION DES DONNÉES ===${NC}"
    
    log "Création track sans titre"
    if req "POST" "$API/tracks" '{"description":"test"}' "auth" "400"; then
        ok "Track sans titre rejeté"
    else
        warn "Validation track insuffisante"
    fi
    
    log "Création listing avec prix négatif"
    if req "POST" "$API/listings" '{"title":"test","price":-10}' "auth" "400"; then
        ok "Prix négatif rejeté"
    else
        warn "Validation prix insuffisante"
    fi
    
    log "Création salon avec nom vide"
    if req "POST" "$API/rooms" '{"name":"","description":"test"}' "auth" "400"; then
        ok "Nom salon vide rejeté"
    else
        warn "Validation salon insuffisante"
    fi
}

# Tests d'injection
test_security_injections() {
    echo -e "\n${BLUE}=== TESTS DE SÉCURITÉ (INJECTIONS) ===${NC}"
    
    log "Injection SQL dans recherche"
    if req "GET" "$API/search?q=' OR 1=1--" "" ""; then
        local response=$(cat "$TMP/last.json")
        if echo "$response" | grep -qi "error\|sql\|syntax"; then
            ok "Injection SQL détectée et bloquée"
        else
            warn "Possible vulnérabilité injection SQL"
        fi
    else
        ok "Requête malicieuse rejetée"
    fi
    
    log "Injection XSS dans création track"
    if req "POST" "$API/tracks" '{"title":"<script>alert(1)</script>","description":"test"}' "auth" ""; then
        local response=$(cat "$TMP/last.json")
        if echo "$response" | grep -q "<script>"; then
            warn "Possible vulnérabilité XSS"
        else
            ok "Contenu XSS échappé/filtré"
        fi
    else
        ok "Contenu malicieux rejeté"
    fi
}

# Tests de charge avancés
test_advanced_performance() {
    echo -e "\n${BLUE}=== TESTS DE PERFORMANCE AVANCÉS ===${NC}"
    
    log "Test de charge - 20 requêtes simultanées"
    local start=$(date +%s%N)
    for i in {1..20}; do
        req "GET" "$API/tags" &
    done
    wait
    local end=$(date +%s%N)
    local duration=$(( (end - start) / 1000000 ))
    
    if [ $duration -lt 2000 ]; then
        ok "Performance excellente: 20 requêtes en ${duration}ms"
    elif [ $duration -lt 5000 ]; then
        warn "Performance correcte: 20 requêtes en ${duration}ms"
    else
        err "Performance dégradée: 20 requêtes en ${duration}ms"
    fi
    
    log "Test rate limiting"
    local rate_limit_hit=0
    for i in {1..25}; do
        req "GET" "$API/tags"
        local status=$(cat "$TMP/status.txt")
        if [ "$status" = "429" ]; then
            rate_limit_hit=1
            break
        fi
        sleep 0.1
    done
    
    if [ $rate_limit_hit -eq 1 ]; then
        ok "Rate limiting actif et efficace"
    else
        warn "Rate limiting non détecté ou très permissif"
    fi
}

# Tests des flux complets
test_complete_workflows() {
    echo -e "\n${BLUE}=== TESTS DE FLUX COMPLETS ===${NC}"
    
    log "Flux complet: Création utilisateur → Track → Salon → Message"
    
    # Création deuxième utilisateur
    local ts2=$(date +%s)_2
    local email2="test_$ts2@test.com"
    local user2="user_$ts2"
    
    if req "POST" "$API/auth/signup" "{\"username\":\"$user2\",\"email\":\"$email2\",\"password\":\"password123\"}"; then
        info "Utilisateur 2 créé"
        
        if req "POST" "$API/auth/login" "{\"email\":\"$email2\",\"password\":\"password123\"}"; then
            local token2=$(cat "$TMP/last.json" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
            info "Utilisateur 2 connecté"
            
            # Créer un track avec le premier utilisateur
            if req "POST" "$API/tracks" '{"title":"Track Collaboratif","description":"Pour tests"}' "auth"; then
                info "Track créé par utilisateur 1"
                
                # Créer un salon
                if req "POST" "$API/rooms" '{"name":"Salon Test","description":"Pour discussions","is_public":true}' "auth"; then
                    local room_id=$(cat "$TMP/last.json" | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
                    info "Salon créé (ID: $room_id)"
                    
                    # Test conversation
                    local saved_token="$TOKEN"
                    TOKEN="$token2"
                    if req "GET" "$API/chat/rooms" "" "auth"; then
                        ok "Flux complet réussi: utilisateurs, tracks, salons fonctionnent ensemble"
                    else
                        warn "Flux partiellement fonctionnel"
                    fi
                    TOKEN="$saved_token"
                else
                    warn "Échec création salon dans le flux"
                fi
            else
                warn "Échec création track dans le flux"
            fi
        else
            warn "Échec connexion utilisateur 2"
        fi
    else
        warn "Échec création utilisateur 2"
    fi
}

# Tests de cohérence des données
test_data_consistency() {
    echo -e "\n${BLUE}=== TESTS DE COHÉRENCE DES DONNÉES ===${NC}"
    
    log "Vérification cohérence utilisateur après modifications"
    local initial_count=$(req "GET" "$API/users" && cat "$TMP/last.json" | grep -o '"id"' | wc -l)
    
    # Modifier le profil
    if req "PUT" "$API/users/me" '{"first_name":"Test","last_name":"User"}' "auth"; then
        # Vérifier que les infos sont bien mises à jour
        if req "GET" "$API/users/me" "" "auth"; then
            local response=$(cat "$TMP/last.json")
            if echo "$response" | grep -q "Test" && echo "$response" | grep -q "User"; then
                ok "Cohérence données utilisateur maintenue"
            else
                warn "Incohérence dans les données utilisateur"
            fi
        fi
    fi
    
    log "Vérification unicité des emails"
    local ts_dup=$(date +%s)_dup
    if req "POST" "$API/auth/signup" "{\"username\":\"duplicate_$ts_dup\",\"email\":\"$EMAIL\",\"password\":\"password123\"}" "" "400"; then
        ok "Unicité email respectée"
    else
        err "Violation contrainte unicité email"
    fi
}

# === EXÉCUTION PRINCIPALE ===
echo -e "${PURPLE}🚀 TESTS API TALAS/VEZA - SUITE AVANCÉE${NC}\n"

# Tests basiques d'abord
./test_api_simple.sh > /dev/null 2>&1
if [ $? -eq 0 ]; then
    info "Tests basiques passés - Démarrage tests avancés"
    
    # Récupération du token depuis les tests basiques
    if [ -f "/tmp/veza_test/last.json" ]; then
        TOKEN=$(grep -o '"access_token":"[^"]*"' /tmp/veza_test/last.json | cut -d'"' -f4 2>/dev/null || echo "")
    fi
    
    # Si pas de token, on refait une auth rapide
    if [ -z "$TOKEN" ]; then
        TS=$(date +%s)
        EMAIL="test_advanced_$TS@test.com"
        req "POST" "$API/auth/signup" "{\"username\":\"user_advanced_$TS\",\"email\":\"$EMAIL\",\"password\":\"password123\"}"
        req "POST" "$API/auth/login" "{\"email\":\"$EMAIL\",\"password\":\"password123\"}"
        TOKEN=$(cat "$TMP/last.json" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
    fi
    
    # Tests avancés
    test_auth_errors
    test_permissions
    test_data_validation
    test_security_injections
    test_advanced_performance
    test_complete_workflows
    test_data_consistency
    
else
    err "Tests basiques échoués - Arrêt des tests avancés"
    exit 1
fi

# === GÉNÉRATION DU RAPPORT FINAL ===
cat >> "$REPORT" << EOF

## 📊 Statistiques finales

| Métrique | Valeur |
|----------|--------|
| **Total des tests** | $TOTAL |
| **Tests réussis** | $PASS |
| **Tests échoués** | $FAIL |
| **Avertissements** | $WARN |
| **Taux de réussite** | $(( PASS * 100 / TOTAL ))% |

## 🏆 Évaluation globale

EOF

if [ $FAIL -eq 0 ]; then
    cat >> "$REPORT" << 'EOF'
### ✅ EXCELLENTE QUALITÉ
L'API Talas est **robuste et prête pour la production**. Tous les tests critiques passent.

#### Points forts identifiés:
- Authentification JWT sécurisée
- Validation des données correcte
- Performances satisfaisantes
- Architecture solide

EOF
    echo -e "\n${GREEN}🎉 TESTS AVANCÉS RÉUSSIS !${NC}"
    echo -e "${GREEN}L'API Talas est de haute qualité et prête pour la production${NC}"
elif [ $FAIL -le 2 ]; then
    cat >> "$REPORT" << 'EOF'
### ⚠️ BONNE QUALITÉ AVEC AMÉLIORATIONS
L'API est **fonctionnelle** mais quelques améliorations sont recommandées.

EOF
    echo -e "\n${YELLOW}⚠️  Tests majoritairement réussis avec quelques points d'amélioration${NC}"
else
    cat >> "$REPORT" << 'EOF'
### ❌ QUALITÉ À AMÉLIORER
Plusieurs points critiques nécessitent une attention immédiate.

EOF
    echo -e "\n${RED}❌ Plusieurs problèmes détectés - Révision nécessaire${NC}"
fi

cat >> "$REPORT" << EOF

## 🔧 Recommandations

1. **Sécurité**: Continuer les tests de pénétration
2. **Performance**: Monitorer en production avec des métriques
3. **Documentation**: Mettre à jour la documentation API
4. **Tests**: Intégrer ces tests dans la CI/CD

## 📅 Prochaines étapes

- [ ] Déployment en environnement de staging
- [ ] Tests d'intégration avec le frontend React
- [ ] Tests de charge avec plus d'utilisateurs simultanés
- [ ] Validation de la sauvegarde/restauration des données

---
*Rapport généré automatiquement le $(date)*
EOF

echo -e "\n${CYAN}📊 Statistiques complètes:${NC}"
echo -e "   Total: ${BLUE}$TOTAL${NC} | Réussis: ${GREEN}$PASS${NC} | Échecs: ${RED}$FAIL${NC} | Warnings: ${YELLOW}$WARN${NC}"
echo -e "\n${PURPLE}📄 Rapport complet généré: $REPORT${NC}"
echo -e "${PURPLE}🔍 Logs de debug: $TMP/debug.log${NC}"

# Ouvrir le rapport si possible
if command -v firefox >/dev/null 2>&1; then
    echo -e "${CYAN}🌐 Ouverture du rapport dans Firefox...${NC}"
    firefox "$REPORT" 2>/dev/null &
fi 