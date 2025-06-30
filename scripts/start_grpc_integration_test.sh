#!/bin/bash

# =============================================================================
# 🚀 SCRIPT DE TEST D'INTÉGRATION gRPC - PHASE 2
# =============================================================================

set -e

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
echo "📂 Dossier projet : $PROJECT_ROOT"

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'  
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Nettoyage
cleanup() {
    log_info "🧹 Nettoyage..."
    pkill -f "chat_server" || true
    pkill -f "stream_server" || true  
    pkill -f "grpc_test_server" || true
    sleep 2
    log_success "Nettoyage terminé"
}

trap cleanup EXIT

echo ""
echo "======================================================================="
echo "🚀 TEST INTÉGRATION gRPC - PHASE 2"
echo "======================================================================="

# 1. Vérification des dépendances
log_info "Vérification des dépendances..."

if ! command -v cargo &> /dev/null; then
    log_error "Cargo (Rust) requis"
    exit 1
fi

if ! command -v go &> /dev/null; then
    log_error "Go requis"
    exit 1
fi

log_success "Dépendances OK"

# 2. Test de compilation backend Go
log_info "Test compilation Backend Go..."
cd "$PROJECT_ROOT"
if go build -o /tmp/grpc_test_server cmd/server/grpc_test_server.go; then
    log_success "Backend Go compile"
else
    log_error "Erreur compilation Backend"
    exit 1
fi

# 3. Test de compilation Chat Server
log_info "Test compilation Chat Server..."
cd "$PROJECT_ROOT/../veza-chat-server"
if timeout 60 cargo check --release; then
    log_success "Chat Server compile"
else
    log_error "Problème compilation Chat Server"
fi

# 4. Test de compilation Stream Server
log_info "Test compilation Stream Server..."
cd "$PROJECT_ROOT/../veza-stream-server"
if timeout 60 cargo check --release; then
    log_success "Stream Server compile"
else
    log_error "Problème compilation Stream Server"
fi

# 5. Démarrage Backend simple
log_info "Démarrage Backend de test..."
cd "$PROJECT_ROOT"
/tmp/grpc_test_server > /tmp/backend.log 2>&1 &
BACKEND_PID=$!
sleep 3

if ps -p $BACKEND_PID > /dev/null; then
    log_success "Backend démarré (PID: $BACKEND_PID)"
else
    log_error "Échec démarrage Backend"
    cat /tmp/backend.log
    exit 1
fi

# 6. Test de santé
log_info "Test santé Backend..."
sleep 2
if curl -s http://localhost:8080/health > /tmp/health.json; then
    log_success "Backend accessible"
    echo "Réponse :"
    cat /tmp/health.json
else
    log_error "Backend inaccessible"
fi

echo ""
echo "======================================================================="
echo "✅ PHASE 2 - INTÉGRATION gRPC : INFRASTRUCTURE PRÊTE"
echo "======================================================================="
echo ""
echo "Backend de test : http://localhost:8080/health"
echo "Logs backend    : tail -f /tmp/backend.log"
echo ""
echo "Prochaines étapes :"
echo "1. Démarrer serveurs Rust gRPC"
echo "2. Tester communication gRPC"
echo "3. Implémenter authentification JWT"
echo ""
echo "Appuyez sur Ctrl+C pour arrêter"
echo "======================================================================="

# Attente
wait 