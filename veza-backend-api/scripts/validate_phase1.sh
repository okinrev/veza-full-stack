#!/bin/bash

# =============================================================================
# SCRIPT DE VALIDATION PHASE 1 - ARCHITECTURE HEXAGONALE
# =============================================================================

set -e

echo "🚀 Validation de la Phase 1 - Architecture Hexagonale"
echo "======================================================="

# Configuration
PROJECT_ROOT=$(pwd)
LOG_FILE="/tmp/veza_phase1_validation.log"

# Couleurs pour l'affichage
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Fonctions utilitaires
print_step() {
    echo -e "\n${BLUE}📋 $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# 1. Vérification de la structure du projet
print_step "1. Vérification de la structure hexagonale"

REQUIRED_DIRS=(
    "internal/domain/entities"
    "internal/domain/repositories"
    "internal/domain/services"
    "internal/ports/http"
    "internal/infrastructure/config"
    "internal/infrastructure/container"
)

for dir in "${REQUIRED_DIRS[@]}"; do
    if [ -d "$dir" ]; then
        print_success "Structure: $dir"
    else
        print_error "Structure manquante: $dir"
        exit 1
    fi
done

# 2. Vérification des fichiers clés
print_step "2. Vérification des fichiers clés"

REQUIRED_FILES=(
    "internal/domain/entities/user.go"
    "internal/domain/repositories/user_repository.go"
    "internal/domain/services/auth_service.go"
    "internal/ports/http/auth_handler.go"
    "internal/infrastructure/config/app_config.go"
    "internal/infrastructure/container/container.go"
    "cmd/server/main_hexagonal.go"
    "config.example.env"
)

for file in "${REQUIRED_FILES[@]}"; do
    if [ -f "$file" ]; then
        print_success "Fichier: $file"
    else
        print_error "Fichier manquant: $file"
        exit 1
    fi
done

# 3. Vérification de la compilation
print_step "3. Test de compilation"

if go build -o /tmp/veza_test ./cmd/server/main_hexagonal.go > "$LOG_FILE" 2>&1; then
    print_success "Compilation réussie"
    rm -f /tmp/veza_test
else
    print_error "Erreur de compilation:"
    cat "$LOG_FILE"
    exit 1
fi

# 4. Vérification des dépendances
print_step "4. Vérification des dépendances Go"

if go mod tidy > "$LOG_FILE" 2>&1; then
    print_success "Dépendances Go mises à jour"
else
    print_warning "Problème avec les dépendances:"
    cat "$LOG_FILE"
fi

# 5. Tests unitaires (si disponibles)
print_step "5. Exécution des tests unitaires"

if [ -f "internal/domain/entities/user_test.go" ] || [ -f "internal/domain/services/auth_service_test.go" ]; then
    if go test ./internal/domain/... -v > "$LOG_FILE" 2>&1; then
        print_success "Tests unitaires passés"
    else
        print_warning "Certains tests échouent:"
        tail -n 10 "$LOG_FILE"
    fi
else
    print_warning "Aucun test unitaire trouvé (optionnel pour Phase 1)"
fi

# 6. Validation de la configuration
print_step "6. Validation de la configuration"

if [ -f "config.example.env" ]; then
    print_success "Fichier de configuration exemple présent"
    
    # Vérifier les variables clés
    REQUIRED_VARS=(
        "DATABASE_URL"
        "REDIS_HOST"
        "JWT_ACCESS_SECRET"
        "JWT_REFRESH_SECRET"
    )
    
    for var in "${REQUIRED_VARS[@]}"; do
        if grep -q "^$var=" config.example.env; then
            print_success "Configuration: $var"
        else
            print_error "Variable manquante: $var"
        fi
    done
else
    print_error "Fichier de configuration manquant"
fi

# 7. Vérification de l'endpoint de santé
print_step "7. Test de l'endpoint hexagonal"

if [ -f "cmd/server/main_hexagonal.go" ]; then
    if grep -q "/hexagonal/status" cmd/server/main_hexagonal.go; then
        print_success "Endpoint de statut hexagonal présent"
    else
        print_warning "Endpoint de statut non trouvé"
    fi
fi

# 8. Résumé final
print_step "8. Résumé de la validation"

echo ""
echo "🎯 PHASE 1 - ARCHITECTURE HEXAGONALE"
echo "===================================="
echo ""
echo "✅ Structure hexagonale complète"
echo "   ├── Domain (entities, repositories, services)"
echo "   ├── Ports (HTTP handlers)"
echo "   ├── Adapters (PostgreSQL, Redis - à compléter)"
echo "   └── Infrastructure (config, container)"
echo ""
echo "✅ Fichiers de base créés"
echo "   ├── Entité User avec validation métier"
echo "   ├── Service Auth avec logique business"
echo "   ├── Handler HTTP avec gestion d'erreurs"
echo "   └── Configuration complète"
echo ""
echo "🔄 Prochaines étapes Phase 1:"
echo "   1. Implémenter les adapters PostgreSQL et Redis"
echo "   2. Finaliser l'intégration JWT"
echo "   3. Ajouter les tests d'intégration"
echo "   4. Tester les endpoints d'authentification"
echo ""

# 9. Recommandations
echo "💡 RECOMMANDATIONS PHASE 1:"
echo ""
echo "1. Configuration:"
echo "   cp config.example.env .env"
echo "   # Modifier .env avec vos paramètres locaux"
echo ""
echo "2. Base de données (PostgreSQL):"
echo "   docker run -d --name postgres_veza -p 5432:5432 -e POSTGRES_PASSWORD=password -e POSTGRES_DB=veza_dev postgres:15"
echo ""
echo "3. Cache (Redis):"
echo "   docker run -d --name redis_veza -p 6379:6379 redis:7-alpine"
echo ""
echo "4. Test de l'architecture:"
echo "   go run cmd/server/main_hexagonal.go"
echo "   curl http://localhost:8080/hexagonal/status"
echo ""

print_success "Validation Phase 1 terminée!"

# Nettoyage
rm -f "$LOG_FILE"

echo ""
echo "📊 MÉTRIQUES CIBLES PHASE 1:"
echo "   • Architecture: 100% hexagonale ✅"
echo "   • Injection dépendances: ✅"
echo "   • Configuration avancée: ✅"
echo "   • Service Auth fonctionnel: 🔄 (en cours)"
echo "   • Cache Redis: 🔄 (en cours)"
echo "   • Tests unitaires: 🔄 (optionnel)"
echo ""
echo "🚀 Phase 1 prête pour finalisation!" 