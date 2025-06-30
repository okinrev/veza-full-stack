#!/bin/bash

# =============================================================================
# 🚀 SCRIPT DE TEST D'INTÉGRATION gRPC - PHASE 2
# =============================================================================
# 
# Ce script démarre les 3 services pour tester l'intégration gRPC :
# 1. Chat Server Rust (port 50051)
# 2. Stream Server Rust (port 50052) 
# 3. Backend Go avec clients gRPC (port 8080)
#
# Usage: ./scripts/start_grpc_integration_test.sh
# =============================================================================

set -e

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
echo "📂 Dossier projet : $PROJECT_ROOT"

# Couleurs pour l'affichage
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# Fonction pour afficher les messages
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Fonction de nettoyage
cleanup() {
    log_info "🧹 Nettoyage des processus..."
    pkill -f "chat_server" || true
    pkill -f "stream_server" || true  
    pkill -f "grpc_test_server" || true
    sleep 2
    log_success "Nettoyage terminé"
}

# Piège pour nettoyer lors d'un arrêt
trap cleanup EXIT

echo ""
echo "======================================================================="
echo "🚀 DÉMARRAGE TEST INTÉGRATION gRPC - PHASE 2"
echo "======================================================================="

# 1. Vérification des dépendances
log_info "Vérification des dépendances..."

if ! command -v cargo &> /dev/null; then
    log_error "Cargo (Rust) n'est pas installé"
    exit 1
fi

if ! command -v go &> /dev/null; then
    log_error "Go n'est pas installé"
    exit 1
fi

if ! command -v protoc &> /dev/null; then
    log_error "protoc n'est pas installé"
    exit 1
fi

log_success "Toutes les dépendances sont présentes"

# 2. Compilation des modules Rust
log_info "Compilation du Chat Server Rust..."
cd "$PROJECT_ROOT/veza-chat-server"
if cargo build --release > /tmp/chat_build.log 2>&1; then
    log_success "Chat Server compilé"
else
    log_error "Erreur compilation Chat Server"
    tail -20 /tmp/chat_build.log
    exit 1
fi

log_info "Compilation du Stream Server Rust..."
cd "$PROJECT_ROOT/veza-stream-server"
if cargo build --release > /tmp/stream_build.log 2>&1; then
    log_success "Stream Server compilé"
else
    log_error "Erreur compilation Stream Server"
    tail -20 /tmp/stream_build.log
    exit 1
fi

# 3. Démarrage des serveurs Rust en arrière-plan
log_info "Démarrage Chat Server gRPC (port 50051)..."
cd "$PROJECT_ROOT/veza-chat-server"
nohup ./target/release/chat_server > /tmp/chat_server.log 2>&1 &
CHAT_PID=$!
sleep 3

if ps -p $CHAT_PID > /dev/null; then
    log_success "Chat Server démarré (PID: $CHAT_PID)"
else
    log_error "Échec démarrage Chat Server"
    tail -20 /tmp/chat_server.log
    exit 1
fi

log_info "Démarrage Stream Server gRPC (port 50052)..."
cd "$PROJECT_ROOT/veza-stream-server"
nohup ./target/release/stream_server > /tmp/stream_server.log 2>&1 &
STREAM_PID=$!
sleep 3

if ps -p $STREAM_PID > /dev/null; then
    log_success "Stream Server démarré (PID: $STREAM_PID)"
else
    log_error "Échec démarrage Stream Server"
    tail -20 /tmp/stream_server.log
    exit 1
fi

# 4. Test de connectivité gRPC
log_info "Test de connectivité gRPC..."

# Test Chat Server
if timeout 5 bash -c "</dev/tcp/localhost/50051" 2>/dev/null; then
    log_success "Chat Server gRPC accessible sur port 50051"
else
    log_error "Chat Server gRPC inaccessible"
fi

# Test Stream Server
if timeout 5 bash -c "</dev/tcp/localhost/50052" 2>/dev/null; then
    log_success "Stream Server gRPC accessible sur port 50052"
else
    log_error "Stream Server gRPC inaccessible"
fi

# 5. Démarrage du Backend Go
log_info "Compilation et démarrage Backend Go..."
cd "$PROJECT_ROOT/veza-backend-api"

# Utiliser le serveur simple d'abord
if go run cmd/server/grpc_test_server.go > /tmp/backend.log 2>&1 &
then
    BACKEND_PID=$!
    sleep 3
    
    if ps -p $BACKEND_PID > /dev/null; then
        log_success "Backend Go démarré (PID: $BACKEND_PID)"
    else
        log_error "Échec démarrage Backend Go"
        tail -20 /tmp/backend.log
        exit 1
    fi
else
    log_error "Erreur compilation Backend Go"
    exit 1
fi

# 6. Tests d'intégration
echo ""
echo "======================================================================="
echo "🧪 TESTS D'INTÉGRATION gRPC"
echo "======================================================================="

sleep 5

# Test santé Backend
log_info "Test santé Backend Go..."
if curl -s http://localhost:8080/health > /tmp/health_test.json; then
    log_success "Backend accessible"
    echo "Réponse santé :"
    cat /tmp/health_test.json | jq . 2>/dev/null || cat /tmp/health_test.json
else
    log_error "Backend inaccessible"
fi

echo ""

# Affichage des logs en temps réel
echo "======================================================================="
echo "📊 STATUT DES SERVICES"
echo "======================================================================="

printf "%-20s %-10s %-20s\n" "SERVICE" "PID" "STATUS"
printf "%-20s %-10s %-20s\n" "-------" "---" "------"
printf "%-20s %-10s %-20s\n" "Chat Server" "$CHAT_PID" "$(ps -p $CHAT_PID >/dev/null && echo 'RUNNING' || echo 'STOPPED')"
printf "%-20s %-10s %-20s\n" "Stream Server" "$STREAM_PID" "$(ps -p $STREAM_PID >/dev/null && echo 'RUNNING' || echo 'STOPPED')"
printf "%-20s %-10s %-20s\n" "Backend Go" "$BACKEND_PID" "$(ps -p $BACKEND_PID >/dev/null && echo 'RUNNING' || echo 'STOPPED')"

echo ""
echo "======================================================================="
echo "🎯 INTÉGRATION gRPC PRÊTE !"
echo "======================================================================="
echo ""
echo "Services disponibles :"
echo "• Backend Go     : http://localhost:8080"
echo "• Chat gRPC      : localhost:50051"
echo "• Stream gRPC    : localhost:50052"
echo ""
echo "Tests disponibles :"
echo "• Santé          : curl http://localhost:8080/health"
echo "• Chat gRPC      : curl -X POST http://localhost:8080/test/chat"
echo "• Stream gRPC    : curl -X POST http://localhost:8080/test/stream"
echo ""
echo "Logs en temps réel :"
echo "• Chat Server    : tail -f /tmp/chat_server.log"
echo "• Stream Server  : tail -f /tmp/stream_server.log"
echo "• Backend        : tail -f /tmp/backend.log"
echo ""
echo "Appuyez sur Ctrl+C pour arrêter tous les services"
echo "======================================================================="

# Attente interactive
wait 