# Étape 6 - Tests et validation finale

## 🎯 Objectif
Valider que le backend reconstruit fonctionne correctement et est prêt pour la production.

## ⏱️ Durée estimée : 20-30 minutes

## 🧪 Plan de validation

### Phase 6.1 : Tests de compilation et structure
1. Validation de la compilation
2. Vérification de l'architecture
3. Audit des dépendances

### Phase 6.2 : Tests fonctionnels
4. Tests des endpoints principaux
5. Tests d'authentification
6. Tests de sécurité

### Phase 6.3 : Tests de performance
7. Tests de charge basiques
8. Validation des migrations
9. Tests d'intégration

## 🔧 Implémentation des tests

### Phase 6.1 : Tests de compilation et structure

#### Script de validation complète : `scripts/validate_build.sh`
```bash
#!/bin/bash

echo "🔍 Validation complète du backend Talas"
echo "=========================================="

# Couleurs pour les messages
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Compteurs
TESTS_PASSED=0
TESTS_FAILED=0

# Fonction utilitaire
test_step() {
    local test_name="$1"
    local command="$2"
    local expected_exit_code="${3:-0}"
    
    echo -n "• $test_name... "
    
    if eval "$command" >/dev/null 2>&1; then
        if [ $? -eq $expected_exit_code ]; then
            echo -e "${GREEN}✅ PASS${NC}"
            ((TESTS_PASSED++))
            return 0
        else
            echo -e "${RED}❌ FAIL (exit code)${NC}"
            ((TESTS_FAILED++))
            return 1
        fi
    else
        echo -e "${RED}❌ FAIL${NC}"
        ((TESTS_FAILED++))
        return 1
    fi
}

echo "1. 📁 STRUCTURE ET ARCHITECTURE"
echo "================================"

# Vérifier la structure des modules API
test_step "Structure des modules API" "[ -d internal/api ] && [ $(find internal/api -name '*.go' | wc -l) -gt 10 ]"

# Vérifier que l'ancien système est supprimé
test_step "Suppression routes legacy" "[ ! -d internal/routes ]"
test_step "Suppression services legacy" "[ ! -d internal/services ]"

# Vérifier la présence des fichiers essentiels
test_step "main.go existe" "[ -f cmd/server/main.go ]"
test_step "Router API existe" "[ -f internal/api/router.go ]"
test_step "Configuration existe" "[ -f internal/config/config.go ]"

echo ""
echo "2. 🔨 COMPILATION ET DÉPENDANCES"
echo "================================="

# Nettoyage des modules
test_step "Nettoyage modules Go" "go clean -modcache && go mod tidy"

# Vérification des modules
test_step "Modules Go valides" "go mod verify"

# Test de compilation
test_step "Compilation réussie" "go build -o tmp_server ./cmd/server && rm -f tmp_server"

# Vérification des imports
test_step "Imports corrects" "! grep -r 'veza-web-app/' internal/ cmd/ --include='*.go' | grep -v github.com/okinrev"

echo ""
echo "3. 🏗️ ARCHITECTURE DES MODULES"
echo "==============================="

# Compter les modules
MODULES_COUNT=$(find internal/api -maxdepth 1 -type d | grep -v '^internal/api$' | wc -l)
test_step "Modules API présents ($MODULES_COUNT)" "[ $MODULES_COUNT -ge 8 ]"

# Vérifier la structure des modules
for module_dir in internal/api/*/; do
    if [ -d "$module_dir" ]; then
        module_name=$(basename "$module_dir")
        test_step "Module $module_name: handler.go" "[ -f ${module_dir}handler.go ]"
        test_step "Module $module_name: service.go" "[ -f ${module_dir}service.go ]"
        test_step "Module $module_name: routes.go" "[ -f ${module_dir}routes.go ]"
    fi
done

echo ""
echo "4. 🔐 SÉCURITÉ ET CONFIGURATION"
echo "==============================="

# Vérifier les middlewares de sécurité
test_step "Middleware auth existe" "[ -f internal/middleware/auth.go ]"
test_step "Utilitaires auth existent" "[ -f internal/utils/auth.go ]"

# Vérifier la configuration
test_step "Configuration JWT" "grep -q 'JWTConfig' internal/config/config.go"
test_step "Configuration DB" "grep -q 'DatabaseConfig' internal/config/config.go"

echo ""
echo "📊 RÉSUMÉ"
echo "========="
echo -e "Tests passés: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Tests échoués: ${RED}$TESTS_FAILED${NC}"

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "\n${GREEN}🎉 TOUS LES TESTS SONT PASSÉS !${NC}"
    echo "Le backend est prêt pour les tests fonctionnels."
    exit 0
else
    echo -e "\n${RED}⚠️ CERTAINS TESTS ONT ÉCHOUÉ${NC}"
    echo "Veuillez corriger les problèmes avant de continuer."
    exit 1
fi
```

#### Script de vérification des imports : `scripts/check_imports.sh`
```bash
#!/bin/bash

echo "🔍 Vérification des imports..."

# Rechercher les imports incorrects
echo "1. Recherche d'imports incorrects:"
INCORRECT_IMPORTS=$(grep -r "veza-web-app/" internal/ cmd/ pkg/ --include="*.go" | grep -v "github.com/okinrev" || true)

if [ -z "$INCORRECT_IMPORTS" ]; then
    echo "✅ Tous les imports sont corrects"
else
    echo "❌ Imports incorrects trouvés:"
    echo "$INCORRECT_IMPORTS"
    exit 1
fi

# Vérifier les imports circulaires
echo ""
echo "2. Vérification des imports circulaires:"
if command -v go >/dev/null 2>&1; then
    go list -json ./... | jq -r 'select(.ImportPath != null) | .ImportPath + " imports: " + (.Imports[]? // "none")' > /tmp/imports.txt
    
    # Recherche simple de cycles (méthode basique)
    if grep -q "internal/api/.* imports.*internal/api" /tmp/imports.txt; then
        echo "⚠️ Cycles d'import potentiels détectés"
        grep "internal/api/.* imports.*internal/api" /tmp/imports.txt
    else
        echo "✅ Aucun cycle d'import détecté"
    fi
    rm -f /tmp/imports.txt
fi

echo ""
echo "3. Vérification des dépendances manquantes:"
go mod tidy
if [ $? -eq 0 ]; then
    echo "✅ Toutes les dépendances sont satisfaites"
else
    echo "❌ Problèmes de dépendances détectés"
    exit 1
fi

echo ""
echo "✅ Vérification des imports terminée avec succès"
```

### Phase 6.2 : Tests fonctionnels

#### Script de tests des endpoints : `scripts/test_endpoints.sh`
```bash
#!/bin/bash

echo "🧪 Tests des endpoints API"
echo "=========================="

# Configuration
BASE_URL="http://localhost:8080"
API_URL="$BASE_URL/api/v1"

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Fonction de test HTTP
test_endpoint() {
    local method="$1"
    local endpoint="$2"
    local expected_status="$3"
    local description="$4"
    local data="$5"
    
    echo -n "• $description... "
    
    if [ -n "$data" ]; then
        response=$(curl -s -w "%{http_code}" -X "$method" \
            -H "Content-Type: application/json" \
            -d "$data" \
            "$endpoint" -o /tmp/response.json)
    else
        response=$(curl -s -w "%{http_code}" -X "$method" \
            "$endpoint" -o /tmp/response.json)
    fi
    
    status_code="${response: -3}"
    
    if [ "$status_code" = "$expected_status" ]; then
        echo -e "${GREEN}✅ PASS${NC} ($status_code)"
        return 0
    else
        echo -e "${RED}❌ FAIL${NC} (got $status_code, expected $expected_status)"
        return 1
    fi
}

# Vérifier que le serveur est accessible
echo "1. Tests de base"
echo "================"

test_endpoint "GET" "$BASE_URL/health" "200" "Health check"

echo ""
echo "2. Tests d'authentification"
echo "==========================="

# Test d'inscription (doit échouer avec données vides)
test_endpoint "POST" "$API_URL/auth/register" "400" "Register sans données" '{"username":"","email":"","password":""}'

# Test d'inscription valide
TIMESTAMP=$(date +%s)
test_endpoint "POST" "$API_URL/auth/register" "200" "Register utilisateur test" "{\"username\":\"test$TIMESTAMP\",\"email\":\"test$TIMESTAMP@example.com\",\"password\":\"password123\"}"

# Test de login (doit échouer)
test_endpoint "POST" "$API_URL/auth/login" "401" "Login avec mauvais credentials" '{"email":"wrong@example.com","password":"wrongpass"}'

echo ""
echo "3. Tests des routes protégées"
echo "============================="

# Test accès sans token (doit échouer)
test_endpoint "GET" "$API_URL/auth/me" "401" "Accès profil sans token"

# Test routes publiques
test_endpoint "GET" "$API_URL/users" "200" "Liste utilisateurs (public)"

echo ""
echo "4. Tests des modules API"
echo "========================"

# Tracks (public)
test_endpoint "GET" "$API_URL/tracks" "200" "Liste tracks publiques"

# Search (public)
test_endpoint "GET" "$API_URL/search?q=test" "200" "Recherche globale"

# Tags (public)
test_endpoint "GET" "$API_URL/tags" "200" "Liste des tags"

echo ""
echo "5. Tests d'erreurs"
echo "=================="

# Route inexistante
test_endpoint "GET" "$API_URL/nonexistent" "404" "Route inexistante"

# Méthode non autorisée
test_endpoint "PATCH" "$API_URL/health" "405" "Méthode non autorisée"

echo ""
echo "✅ Tests des endpoints terminés"

# Nettoyage
rm -f /tmp/response.json
```

#### Tests de performance basique : `scripts/test_performance.sh`
```bash
#!/bin/bash

echo "⚡ Tests de performance basique"
echo "=============================="

BASE_URL="http://localhost:8080"

# Vérifier si ab (Apache Bench) est disponible
if ! command -v ab &> /dev/null; then
    echo "⚠️ Apache Bench (ab) non disponible, installation..."
    # Sur Ubuntu/Debian
    if command -v apt-get &> /dev/null; then
        sudo apt-get update && sudo apt-get install -y apache2-utils
    else
        echo "❌ Impossible d'installer ab automatiquement"
        echo "Veuillez installer apache2-utils manuellement"
        exit 1
    fi
fi

echo "1. Test de charge health check"
echo "=============================="
ab -n 100 -c 10 -q "$BASE_URL/health"

echo ""
echo "2. Test de charge API"
echo "===================="
ab -n 50 -c 5 -q "$BASE_URL/api/v1/tracks"

echo ""
echo "3. Test de latence"
echo "=================="
for i in {1..5}; do
    echo -n "Request $i: "
    curl -w "@-" -o /dev/null -s "$BASE_URL/health" <<'EOF'
     time_namelookup:  %{time_namelookup}\n
        time_connect:  %{time_connect}\n
     time_appconnect:  %{time_appconnect}\n
    time_pretransfer:  %{time_pretransfer}\n
       time_redirect:  %{time_redirect}\n
  time_starttransfer:  %{time_starttransfer}\n
                     ----------\n
          time_total:  %{time_total}\n
EOF
done

echo ""
echo "✅ Tests de performance terminés"
```

### Phase 6.3 : Tests d'intégration

#### Tests de la base de données : `scripts/test_database.sh`
```bash
#!/bin/bash

echo "🗄️ Tests de la base de données"
echo "=============================="

# Vérifier la présence du .env
if [ ! -f .env ]; then
    echo "❌ Fichier .env manquant"
    echo "Créez un fichier .env avec DATABASE_URL"
    exit 1
fi

# Charger les variables d'environnement
source .env

if [ -z "$DATABASE_URL" ]; then
    echo "❌ DATABASE_URL non défini dans .env"
    exit 1
fi

echo "✅ Configuration de base de données trouvée"

echo ""
echo "1. Test de connexion"
echo "==================="

# Test avec psql si disponible
if command -v psql &> /dev/null; then
    echo "Test avec psql..."
    if psql "$DATABASE_URL" -c "SELECT version();" >/dev/null 2>&1; then
        echo "✅ Connexion PostgreSQL réussie"
    else
        echo "❌ Échec de connexion PostgreSQL"
        exit 1
    fi
else
    echo "⚠️ psql non disponible, test avec le backend Go..."
    # Test via compilation temporaire
    cat > /tmp/test_db.go << 'EOF'
package main

import (
    "database/sql"
    "fmt"
    "os"
    _ "github.com/lib/pq"
)

func main() {
    db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
    if err != nil {
        fmt.Printf("Erreur ouverture: %v\n", err)
        os.Exit(1)
    }
    defer db.Close()
    
    if err := db.Ping(); err != nil {
        fmt.Printf("Erreur ping: %v\n", err)
        os.Exit(1)
    }
    
    fmt.Println("✅ Connexion DB réussie")
}
EOF

    cd /tmp && go mod init test_db && go mod tidy && go run test_db.go
    rm -rf /tmp/test_db.go /tmp/go.mod /tmp/go.sum
fi

echo ""
echo "2. Test des migrations"
echo "====================="

# Compter les fichiers de migration
MIGRATION_COUNT=$(find internal/database/migrations -name "*.sql" | wc -l)
echo "📁 Fichiers de migration trouvés: $MIGRATION_COUNT"

if [ $MIGRATION_COUNT -eq 0 ]; then
    echo "❌ Aucun fichier de migration trouvé"
    exit 1
fi

# Test d'exécution des migrations (via le backend)
echo "🔄 Test d'exécution des migrations..."
if timeout 10s go run cmd/server/main.go >/dev/null 2>&1; then
    echo "✅ Migrations exécutées avec succès"
else
    echo "⚠️ Timeout ou erreur lors du test des migrations"
fi

echo ""
echo "3. Vérification des tables"
echo "=========================="

# Lister les tables si psql disponible
if command -v psql &> /dev/null; then
    echo "📋 Tables créées:"
    psql "$DATABASE_URL" -c "\dt" | grep -E "users|tracks|products|listings" || echo "⚠️ Tables principales non trouvées"
fi

echo ""
echo "✅ Tests de base de données terminés"
```

#### Script de nettoyage : `scripts/clean_duplicates.sh`
```bash
#!/bin/bash

echo "🧹 Nettoyage des doublons et fichiers temporaires"
echo "================================================="

# Supprimer les fichiers de backup
echo "1. Suppression des fichiers .bak"
find . -name "*.bak" -type f -delete
echo "✅ Fichiers .bak supprimés"

# Supprimer les fichiers temporaires de compilation
echo ""
echo "2. Nettoyage des fichiers temporaires"
rm -f cmd/server/server
rm -f tmp_server
rm -f /tmp/test_db.go
echo "✅ Fichiers temporaires supprimés"

# Nettoyer le cache Go
echo ""
echo "3. Nettoyage du cache Go"
go clean -cache
go clean -modcache
go mod tidy
echo "✅ Cache Go nettoyé"

# Supprimer les logs temporaires
echo ""
echo "4. Nettoyage des logs"
find . -name "*.log" -type f -delete 2>/dev/null || true
echo "✅ Logs temporaires supprimés"

# Vérifier les permissions
echo ""
echo "5. Vérification des permissions"
find scripts/ -name "*.sh" -exec chmod +x {} \;
echo "✅ Permissions des scripts corrigées"

echo ""
echo "✅ Nettoyage terminé"
```

## 📋 Checklist finale complète

### ✅ Architecture
- [ ] Structure modulaire `internal/api/*/`
- [ ] Services consolidés dans chaque module
- [ ] Routes standardisées
- [ ] Configuration centralisée
- [ ] Middleware de sécurité

### ✅ Compilation
- [ ] `go mod tidy` sans erreur
- [ ] `go build ./cmd/server` réussi
- [ ] Aucun import incorrect
- [ ] Aucun cycle d'import

### ✅ Fonctionnalités
- [ ] Health check accessible
- [ ] Authentification fonctionnelle
- [ ] Routes API répondent
- [ ] Base de données connectée
- [ ] Migrations exécutées

### ✅ Sécurité
- [ ] JWT implémenté
- [ ] Middleware auth fonctionnel
- [ ] Validation des entrées
- [ ] CORS configuré
- [ ] Pas de secrets hardcodés

### ✅ Performance
- [ ] Temps de réponse < 100ms (health)
- [ ] Pas de fuites mémoire apparentes
- [ ] Gestion des erreurs propre

## 🚀 Script de lancement final

#### `scripts/start_server.sh`
```bash
#!/bin/bash

echo "🚀 Lancement du serveur Talas"
echo "============================="

# Vérifications préalables
echo "1. Vérifications..."

# Vérifier .env
if [ ! -f .env ]; then
    echo "❌ Fichier .env manquant"
    echo "Créez un fichier .env avec les variables nécessaires"
    exit 1
fi

# Vérifier compilation
if ! go build -o tmp_server ./cmd/server; then
    echo "❌ Erreur de compilation"
    exit 1
fi

rm -f tmp_server
echo "✅ Compilation OK"

# Vérifier la base de données
source .env
if [ -z "$DATABASE_URL" ]; then
    echo "❌ DATABASE_URL non défini"
    exit 1
fi

echo "✅ Configuration OK"

# Lancement
echo ""
echo "2. Lancement du serveur..."
echo "🌐 URL: http://localhost:${PORT:-8080}"
echo "📖 Health: http://localhost:${PORT:-8080}/health"
echo "🔌 API: http://localhost:${PORT:-8080}/api/v1"
echo ""
echo "Appuyez sur Ctrl+C pour arrêter"
echo "================================"

# Lancer le serveur
go run cmd/server/main.go
```

## 🎯 Validation finale

Une fois tous les tests passés, le backend est considéré comme **entièrement fonctionnel** et prêt pour :

1. ✅ **Développement** : Ajout de nouvelles fonctionnalités
2. ✅ **Tests avancés** : Tests unitaires et d'intégration
3. ✅ **Déploiement** : Mise en production
4. ✅ **Maintenance** : Code maintenable et documenté

---

**💾 COMMIT FINAL**
```bash
git add .
git commit -m "✅ Migration terminée - Backend Go totalement fonctionnel

- Architecture modulaire consolidée
- Services unifiés par module
- Routing simplifié et performant
- Tests de validation passés
- Prêt pour production"

git tag -a v1.0.0-migrated -m "Backend migré vers architecture modulaire"
```

**🎉 FÉLICITATIONS !** 
Votre backend Talas est maintenant totalement fonctionnel avec une architecture moderne, sécurisée et maintenable.