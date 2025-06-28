#!/bin/bash

# Script de synchronisation de la configuration Veza
# Copie la configuration unifiée vers tous les services

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
UNIFIED_CONFIG="$PROJECT_ROOT/configs/env.unified"

echo "🔧 Synchronisation de la configuration Veza..."
echo "📁 Répertoire du projet: $PROJECT_ROOT"
echo "📄 Configuration source: $UNIFIED_CONFIG"

# Vérifier que le fichier unifié existe
if [ ! -f "$UNIFIED_CONFIG" ]; then
    echo "❌ Fichier de configuration unifié introuvable: $UNIFIED_CONFIG"
    exit 1
fi

# Fonction pour créer le .env d'un service
create_service_env() {
    local service_dir="$1"
    local service_name="$2"
    local env_file="$service_dir/.env"
    
    echo "📝 Création de $env_file pour $service_name..."
    
    # Créer le répertoire si nécessaire
    mkdir -p "$service_dir"
    
    # Copier la configuration de base
    cp "$UNIFIED_CONFIG" "$env_file"
    
    # Ajouter des configurations spécifiques au service
    case "$service_name" in
        "backend")
            echo "" >> "$env_file"
            echo "# Configuration spécifique Backend Go" >> "$env_file"
            echo "PORT=8080" >> "$env_file"
            echo "HOST=0.0.0.0" >> "$env_file"
            ;;
        "chat")
            echo "" >> "$env_file"
            echo "# Configuration spécifique Chat Rust" >> "$env_file"
            echo "CHAT_SERVER_BIND_ADDR=0.0.0.0:3001" >> "$env_file"
            ;;
        "stream")
            echo "" >> "$env_file"
            echo "# Configuration spécifique Stream Rust" >> "$env_file"
            echo "STREAM_SERVER_PORT=3002" >> "$env_file"
            echo "STREAM_SERVER_BIND_ADDR=0.0.0.0:3002" >> "$env_file"
            ;;
        "frontend")
            # Pour le frontend, ne garder que les variables VITE_
            echo "# Veza Frontend Configuration - UNIFIÉ" > "$env_file"
            echo "# Généré automatiquement depuis configs/env.unified" >> "$env_file"
            echo "" >> "$env_file"
            grep "^VITE_" "$UNIFIED_CONFIG" >> "$env_file"
            ;;
    esac
    
    echo "✅ $service_name configuré"
}

# Synchroniser tous les services
echo ""
echo "🔄 Synchronisation des services..."

create_service_env "$PROJECT_ROOT/veza-backend-api" "backend"
create_service_env "$PROJECT_ROOT/veza-chat-server" "chat" 
create_service_env "$PROJECT_ROOT/veza-stream-server" "stream"
create_service_env "$PROJECT_ROOT/veza-frontend" "frontend"

# Créer aussi un .env global si possible
if [ -w "$PROJECT_ROOT" ]; then
    echo "📄 Création du .env global..."
    cp "$UNIFIED_CONFIG" "$PROJECT_ROOT/.env"
    echo "✅ .env global créé"
fi

echo ""
echo "🎯 Résumé de la synchronisation:"
echo "   - Backend Go: $PROJECT_ROOT/veza-backend-api/.env"
echo "   - Chat Rust: $PROJECT_ROOT/veza-chat-server/.env"  
echo "   - Stream Rust: $PROJECT_ROOT/veza-stream-server/.env"
echo "   - Frontend React: $PROJECT_ROOT/veza-frontend/.env"
echo ""
echo "✅ Configuration synchronisée avec succès !"
echo ""
echo "🔧 Prochaines étapes:"
echo "   1. Redémarrer tous les services"
echo "   2. Vérifier les connexions WebSocket"
echo "   3. Tester l'authentification JWT" 