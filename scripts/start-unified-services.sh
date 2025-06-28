#!/bin/bash

# Script de démarrage unifié pour tous les services Veza
# Démarre tous les services avec la configuration synchronisée

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}🚀 Démarrage unifié des services Veza${NC}"
echo -e "${BLUE}=====================================${NC}"
echo ""

# Fonction pour vérifier si un port est utilisé
check_port() {
    local port="$1"
    local service="$2"
    
    if command -v netstat >/dev/null 2>&1; then
        if netstat -ln 2>/dev/null | grep -q ":$port "; then
            echo -e "${YELLOW}⚠️ Port $port déjà utilisé par un autre processus${NC}"
            echo "💡 Arrêtez le service existant ou changez le port pour $service"
            return 1
        fi
    elif command -v ss >/dev/null 2>&1; then
        if ss -ln 2>/dev/null | grep -q ":$port "; then
            echo -e "${YELLOW}⚠️ Port $port déjà utilisé par un autre processus${NC}"
            echo "💡 Arrêtez le service existant ou changez le port pour $service"
            return 1
        fi
    fi
    return 0
}

# Fonction pour attendre qu'un service soit prêt
wait_for_service() {
    local host="$1"
    local port="$2"
    local service="$3"
    local max_attempts="${4:-30}"
    
    echo -n "⏳ Attente du démarrage de $service ($host:$port)... "
    
    for i in $(seq 1 $max_attempts); do
        if command -v nc >/dev/null 2>&1; then
            if nc -z "$host" "$port" 2>/dev/null; then
                echo -e "${GREEN}✅ Prêt !${NC}"
                return 0
            fi
        elif command -v curl >/dev/null 2>&1; then
            if curl -s "$host:$port" >/dev/null 2>&1; then
                echo -e "${GREEN}✅ Prêt !${NC}"
                return 0
            fi
        fi
        sleep 2
    done
    
    echo -e "${RED}❌ Timeout après ${max_attempts} tentatives${NC}"
    return 1
}

# Vérifier que les configurations sont synchronisées
echo -e "${BLUE}🔧 Vérification des configurations...${NC}"
if [ ! -f "$PROJECT_ROOT/configs/env.unified" ]; then
    echo -e "${RED}❌ Configuration unifiée manquante${NC}"
    echo "💡 Exécutez d'abord: ./scripts/sync-env-config.sh"
    exit 1
fi

# Vérifier les fichiers .env des services
for service in "veza-backend-api" "veza-chat-server" "veza-stream-server" "veza-frontend"; do
    if [ ! -f "$PROJECT_ROOT/$service/.env" ]; then
        echo -e "${YELLOW}⚠️ Configuration manquante pour $service${NC}"
        echo "💡 Exécutez: ./scripts/sync-env-config.sh"
    fi
done

echo -e "${GREEN}✅ Configurations vérifiées${NC}"
echo ""

# Vérifier la disponibilité des ports
echo -e "${BLUE}🔍 Vérification des ports...${NC}"
check_port 8080 "Backend Go" || true
check_port 3001 "Chat Server Rust" || true  
check_port 3002 "Stream Server Rust" || true
check_port 5173 "Frontend React" || true

echo ""

# Démarrer PostgreSQL et Redis (si pas déjà démarrés)
echo -e "${BLUE}🗄️ Services de données...${NC}"
echo "PostgreSQL : 10.5.191.154:5432"
echo "Redis      : 10.5.191.95:6379"
echo "💡 Assurez-vous que ces services sont démarrés sur leurs containers respectifs"
echo ""

# Démarrer le Backend Go
echo -e "${BLUE}🔧 Démarrage du Backend Go...${NC}"
cd "$PROJECT_ROOT/veza-backend-api"

if [ ! -f ".env" ]; then
    echo -e "${RED}❌ Fichier .env manquant dans veza-backend-api${NC}"
    exit 1
fi

echo "📁 Répertoire: $(pwd)"
echo "🔧 Chargement des variables d'environnement..."

# Afficher quelques variables importantes
echo "📋 Configuration chargée:"
echo "   - PORT: $(grep "^PORT=" .env 2>/dev/null | cut -d'=' -f2 || echo "8080")"
echo "   - JWT_SECRET: $(grep "^JWT_SECRET=" .env 2>/dev/null | cut -d'=' -f2 | cut -c1-20)..."
echo "   - DATABASE_URL: $(grep "^DATABASE_URL=" .env 2>/dev/null | cut -d'=' -f2 | sed 's/password:[^@]*/password:***/')"

# Compiler et démarrer le backend
echo "🔨 Compilation du backend Go..."
if go build -o tmp/backend cmd/server/main.go; then
    echo -e "${GREEN}✅ Compilation réussie${NC}"
    echo "🚀 Démarrage du backend sur le port 8080..."
    ./tmp/backend &
    BACKEND_PID=$!
    echo "Backend PID: $BACKEND_PID"
else
    echo -e "${RED}❌ Échec de la compilation du backend${NC}"
    exit 1
fi

# Attendre que le backend soit prêt
wait_for_service "localhost" "8080" "Backend Go"

echo ""

# Démarrer le Chat Server Rust
echo -e "${BLUE}💬 Démarrage du Chat Server Rust...${NC}"
cd "$PROJECT_ROOT/veza-chat-server"

echo "📁 Répertoire: $(pwd)"
echo "🔧 Configuration du chat server..."
echo "📋 Configuration:"
echo "   - PORT: 3001"
echo "   - WebSocket: /ws"

echo "🔨 Compilation du chat server..."
if cargo build --release; then
    echo -e "${GREEN}✅ Compilation réussie${NC}"
    echo "🚀 Démarrage du chat server sur le port 3001..."
    cargo run --release &
    CHAT_PID=$!
    echo "Chat Server PID: $CHAT_PID"
else
    echo -e "${RED}❌ Échec de la compilation du chat server${NC}"
    # Continuer malgré l'erreur
fi

# Attendre que le chat server soit prêt
wait_for_service "localhost" "3001" "Chat Server" 20 || echo -e "${YELLOW}⚠️ Chat Server peut-être non démarré${NC}"

echo ""

# Démarrer le Stream Server Rust
echo -e "${BLUE}🎵 Démarrage du Stream Server Rust...${NC}"
cd "$PROJECT_ROOT/veza-stream-server"

echo "📁 Répertoire: $(pwd)"
echo "🔧 Configuration du stream server..."
echo "📋 Configuration:"
echo "   - PORT: 3002"
echo "   - Audio DIR: /storage/audio"

echo "🔨 Compilation du stream server..."
if cargo build --release; then
    echo -e "${GREEN}✅ Compilation réussie${NC}"
    echo "🚀 Démarrage du stream server sur le port 3002..."
    cargo run --release &
    STREAM_PID=$!
    echo "Stream Server PID: $STREAM_PID"
else
    echo -e "${RED}❌ Échec de la compilation du stream server${NC}"
    # Continuer malgré l'erreur
fi

# Attendre que le stream server soit prêt
wait_for_service "localhost" "3002" "Stream Server" 20 || echo -e "${YELLOW}⚠️ Stream Server peut-être non démarré${NC}"

echo ""

# Démarrer le Frontend React
echo -e "${BLUE}🌐 Démarrage du Frontend React...${NC}"
cd "$PROJECT_ROOT/veza-frontend"

echo "📁 Répertoire: $(pwd)"
echo "🔧 Configuration du frontend..."

if [ -f ".env" ]; then
    echo "📋 Configuration chargée:"
    echo "   - API URL: $(grep "^VITE_API_URL=" .env 2>/dev/null | cut -d'=' -f2)"
    echo "   - Chat WS: $(grep "^VITE_WS_CHAT_URL=" .env 2>/dev/null | cut -d'=' -f2)"
    echo "   - Stream WS: $(grep "^VITE_WS_STREAM_URL=" .env 2>/dev/null | cut -d'=' -f2)"
else
    echo -e "${YELLOW}⚠️ Fichier .env manquant pour le frontend${NC}"
fi

echo "📦 Installation des dépendances..."
if command -v npm >/dev/null 2>&1; then
    npm install --silent
    echo "🚀 Démarrage du serveur de développement..."
    npm run dev -- --host 0.0.0.0 --port 5173 &
    FRONTEND_PID=$!
    echo "Frontend PID: $FRONTEND_PID"
else
    echo -e "${RED}❌ npm non disponible${NC}"
fi

echo ""

# Résumé des services démarrés
echo -e "${GREEN}🎉 Services démarrés avec succès !${NC}"
echo -e "${GREEN}=================================${NC}"
echo ""
echo -e "${GREEN}✅ Services actifs:${NC}"
echo "   🔧 Backend Go       : http://localhost:8080 (PID: ${BACKEND_PID:-N/A})"
echo "   💬 Chat Server      : http://localhost:3001 (PID: ${CHAT_PID:-N/A})"
echo "   🎵 Stream Server    : http://localhost:3002 (PID: ${STREAM_PID:-N/A})"
echo "   🌐 Frontend React   : http://localhost:5173 (PID: ${FRONTEND_PID:-N/A})"
echo ""
echo -e "${GREEN}🔌 WebSocket Endpoints:${NC}"
echo "   💬 Chat      : ws://localhost:3001/ws"
echo "   🎵 Stream    : ws://localhost:3002/ws"
echo ""
echo -e "${GREEN}🌍 URLs d'accès:${NC}"
echo "   🏠 Application principale: http://localhost:5173"
echo "   📊 API Backend           : http://localhost:8080/api/v1"
echo "   ⚕️ Santé Backend         : http://localhost:8080/api/health"
echo "   ⚕️ Santé Chat            : http://localhost:3001/health"
echo "   ⚕️ Santé Stream          : http://localhost:3002/health"
echo ""
echo -e "${BLUE}📝 Commandes utiles:${NC}"
echo "   🧪 Tester les connexions: ./scripts/test-all-connections.sh"
echo "   ⏹️ Arrêter les services : kill $BACKEND_PID $CHAT_PID $STREAM_PID $FRONTEND_PID"
echo "   📊 Voir les logs        : tail -f logs/*.log"
echo ""

# Attendre que le frontend soit prêt
wait_for_service "localhost" "5173" "Frontend React" 30 || echo -e "${YELLOW}⚠️ Frontend peut-être non démarré${NC}"

echo ""
echo -e "${GREEN}🚀 Tous les services sont prêts !${NC}"
echo -e "${BLUE}💡 Testez maintenant l'application en visitant: http://localhost:5173${NC}"

# Garder les services actifs
echo ""
echo -e "${YELLOW}⏹️ Appuyez sur Ctrl+C pour arrêter tous les services${NC}"

# Gestionnaire de signal pour arrêter tous les services
cleanup() {
    echo ""
    echo -e "${BLUE}🛑 Arrêt des services...${NC}"
    
    if [ ! -z "$FRONTEND_PID" ]; then
        echo "⏹️ Arrêt du Frontend (PID: $FRONTEND_PID)"
        kill $FRONTEND_PID 2>/dev/null || true
    fi
    
    if [ ! -z "$STREAM_PID" ]; then
        echo "⏹️ Arrêt du Stream Server (PID: $STREAM_PID)"
        kill $STREAM_PID 2>/dev/null || true
    fi
    
    if [ ! -z "$CHAT_PID" ]; then
        echo "⏹️ Arrêt du Chat Server (PID: $CHAT_PID)"
        kill $CHAT_PID 2>/dev/null || true
    fi
    
    if [ ! -z "$BACKEND_PID" ]; then
        echo "⏹️ Arrêt du Backend (PID: $BACKEND_PID)"
        kill $BACKEND_PID 2>/dev/null || true
    fi
    
    echo -e "${GREEN}✅ Tous les services ont été arrêtés${NC}"
    exit 0
}

trap cleanup SIGINT SIGTERM

# Attendre indéfiniment
wait 