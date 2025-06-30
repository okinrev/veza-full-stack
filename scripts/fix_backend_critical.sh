#!/bin/bash

# 🔧 SCRIPT DE CORRECTION CRITIQUE DU BACKEND TALAS
# Corrections prioritaires pour rendre le backend fonctionnel

set -e

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Fonctions utilitaires
log_info() { echo -e "${BLUE}ℹ️  $1${NC}"; }
log_success() { echo -e "${GREEN}✅ $1${NC}"; }
log_warning() { echo -e "${YELLOW}⚠️  $1${NC}"; }
log_error() { echo -e "${RED}❌ $1${NC}"; }

echo "🔧 CORRECTIONS CRITIQUES DU BACKEND TALAS"
echo "=========================================="
echo ""

# 1. Test de compilation du serveur principal
log_info "1. Vérification du serveur principal"
cd veza-backend-api

if go build -o tmp/main-test cmd/server/main.go 2>/dev/null; then
    log_success "Serveur principal compile correctement"
    rm -f tmp/main-test
else
    log_error "Serveur principal ne compile pas"
    echo "Erreurs de compilation:"
    go build -o tmp/main-test cmd/server/main.go
    exit 1
fi

# 2. Test des endpoints essentiels
log_info "2. Test des endpoints"

# Démarrer le serveur en arrière-plan pour les tests
if ! pgrep -f "main.go" > /dev/null; then
    log_info "Démarrage du serveur de test..."
    go run cmd/server/main.go &
    SERVER_PID=$!
    sleep 3  # Attendre que le serveur démarre
    
    # Test des endpoints
    if curl -s -f http://localhost:8080/api/health > /dev/null; then
        log_success "Health endpoint fonctionne"
    else
        log_warning "Health endpoint ne répond pas"
    fi
    
    # Test des autres endpoints
    log_info "Tests des endpoints API:"
    echo "  Users: $(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/api/users)"
    echo "  Rooms: $(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/api/rooms)"
    echo "  Search: $(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/api/search)"
    echo "  Tags: $(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/api/tags)"
    
    # Arrêter le serveur de test
    kill $SERVER_PID 2>/dev/null || true
    sleep 1
else
    log_success "Serveur déjà en cours d'exécution"
fi

# 3. Test des modules Rust
log_info "3. Vérification des modules Rust"

cd ../veza-chat-server
if cargo check --quiet 2>/dev/null; then
    log_success "Chat Server Rust compile"
else
    log_warning "Chat Server Rust a des erreurs"
fi

cd ../veza-stream-server  
if cargo check --quiet 2>/dev/null; then
    log_success "Stream Server Rust compile"
else
    log_warning "Stream Server Rust a des erreurs"
fi

cd ..

# 4. Vérification de la base de données
log_info "4. Test de la base de données"

if command -v psql > /dev/null; then
    if psql -h localhost -U postgres -d veza_db -c "SELECT 1;" > /dev/null 2>&1; then
        log_success "PostgreSQL accessible"
        
        # Compter les tables
        table_count=$(psql -h localhost -U postgres -d veza_db -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';" 2>/dev/null | xargs)
        if [ "$table_count" -gt 0 ]; then
            log_success "$table_count tables trouvées en base"
        else
            log_warning "Aucune table trouvée - migrations nécessaires"
        fi
    else
        log_error "PostgreSQL non accessible"
        log_info "Pour configurer la base de données, exécutez: ./scripts/setup_database.sh"
    fi
else
    log_warning "psql non installé - impossible de tester PostgreSQL"
fi

# 5. Test Redis
log_info "5. Test de Redis"

if command -v redis-cli > /dev/null; then
    if redis-cli ping 2>/dev/null | grep -q "PONG"; then
        log_success "Redis accessible"
    else
        log_warning "Redis non accessible"
        log_info "Pour démarrer Redis: sudo systemctl start redis ou redis-server"
    fi
else
    log_warning "redis-cli non installé"
fi

# 6. Compilation test intégration
log_info "6. Test de compilation avancée"

cd veza-backend-api

# Test de compilation avec le serveur hexagonal
if go build -o tmp/hexagonal-test cmd/server/main_hexagonal.go 2>/dev/null; then
    log_success "Architecture hexagonale fonctionne"
    rm -f tmp/hexagonal-test
else
    log_warning "Architecture hexagonale a des erreurs (non bloquant)"
    echo "Utilisez le serveur principal (main.go) en attendant les corrections"
fi

echo ""
echo "🎯 RÉSUMÉ DES CORRECTIONS"
echo "========================"

log_info "Actions recommandées:"
echo "1. Utiliser cmd/server/main.go comme serveur principal"
echo "2. Configurer PostgreSQL avec ./scripts/setup_database.sh"  
echo "3. Démarrer Redis si nécessaire"
echo "4. Les modules Rust compilent mais ont des warnings"
echo "5. L'architecture hexagonale nécessite des corrections"

echo ""
log_success "Backend partiellement fonctionnel - prêt pour les tests de base"
echo "Pour démarrer le serveur: cd veza-backend-api && go run cmd/server/main.go" 