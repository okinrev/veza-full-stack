#!/bin/bash
# Script de génération de configuration unifiée avec IPs dynamiques

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORKSPACE_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
CONFIG_DIR="$WORKSPACE_DIR/configs"

# Créer le dossier configs
mkdir -p "$CONFIG_DIR"

# Fonction pour obtenir l'IP d'un container
get_container_ip() {
    incus list "$1" --format csv | cut -d, -f3 | grep -E '^10\.' | head -1
}

# Récupérer toutes les IPs
echo "Récupération des IPs des containers..."

POSTGRES_IP=$(get_container_ip veza-postgres)
REDIS_IP=$(get_container_ip veza-redis)
BACKEND_IP=$(get_container_ip veza-backend)
CHAT_IP=$(get_container_ip veza-chat)
STREAM_IP=$(get_container_ip veza-stream)
FRONTEND_IP=$(get_container_ip veza-frontend)
STORAGE_IP=$(get_container_ip veza-storage)
HAPROXY_IP=$(get_container_ip veza-haproxy)

# Générer la configuration unifiée
cat > "$CONFIG_DIR/env.unified" <<EOF
# Configuration Unifiée veza - Générée automatiquement
# Date: $(date)

# IPs des Services
POSTGRES_HOST=$POSTGRES_IP
REDIS_HOST=$REDIS_IP
BACKEND_HOST=$BACKEND_IP
CHAT_HOST=$CHAT_IP
STREAM_HOST=$STREAM_IP
FRONTEND_HOST=$FRONTEND_IP
STORAGE_HOST=$STORAGE_IP
HAPROXY_HOST=$HAPROXY_IP

# Database
DATABASE_URL=postgresql://veza:veza_password@$POSTGRES_IP:5432/veza_db
DATABASE_HOST=$POSTGRES_IP
DATABASE_PORT=5432
DATABASE_NAME=veza_db
DATABASE_USER=veza
DATABASE_PASSWORD=veza_password

# Redis
REDIS_URL=redis://:veza_redis_password@$REDIS_IP:6379
REDIS_HOST=$REDIS_IP
REDIS_PORT=6379
REDIS_PASSWORD=veza_redis_password

# JWT Configuration (CRITICAL - Same for all services)
JWT_SECRET=veza_unified_jwt_secret_key_2025_microservices_secure_32chars_minimum
JWT_ISSUER=veza-platform
JWT_AUDIENCE=veza-services
JWT_EXPIRY_ACCESS=15m
JWT_EXPIRY_REFRESH=7d

# Service URLs
API_BASE_URL=http://$BACKEND_IP:8080/api/v1
CHAT_WS_URL=ws://$CHAT_IP:3001/ws
STREAM_WS_URL=ws://$STREAM_IP:3002/ws

# Frontend Config
VITE_API_URL=http://$BACKEND_IP:8080
VITE_CHAT_WS_URL=ws://$CHAT_IP:3001
VITE_STREAM_WS_URL=ws://$STREAM_IP:3002

# Public Access (via HAProxy)
PUBLIC_URL=http://$HAPROXY_IP
PUBLIC_API_URL=http://$HAPROXY_IP/api/v1
PUBLIC_CHAT_WS_URL=ws://$HAPROXY_IP/ws/chat
PUBLIC_STREAM_WS_URL=ws://$HAPROXY_IP/ws/stream

# Storage
NFS_STORAGE_PATH=/storage
UPLOAD_DIR=/storage/uploads
TRACKS_DIR=/storage/tracks

# Environment
NODE_ENV=development
GO_ENV=development
RUST_ENV=development
EOF

# Générer les .env pour chaque service
echo "Génération des fichiers .env spécifiques..."

# Backend Go .env
cat > "$CONFIG_DIR/backend.env" <<EOF
# Backend Go Environment
DATABASE_URL=postgresql://veza:veza_password@$POSTGRES_IP:5432/veza_db
REDIS_URL=redis://:veza_redis_password@$REDIS_IP:6379
JWT_SECRET=veza_unified_jwt_secret_key_2025_microservices_secure_32chars_minimum
JWT_ISSUER=veza-platform
JWT_AUDIENCE=veza-services
PORT=8080
HOST=0.0.0.0
STORAGE_HOST=$STORAGE_IP
STORAGE_PATH=/storage
EOF

# Chat Server .env
cat > "$CONFIG_DIR/chat.env" <<EOF
# Chat Server Rust Environment
DATABASE_URL=postgresql://veza:veza_password@$POSTGRES_IP:5432/veza_db
REDIS_URL=redis://:veza_redis_password@$REDIS_IP:6379
JWT_SECRET=veza_unified_jwt_secret_key_2025_microservices_secure_32chars_minimum
JWT_ISSUER=veza-platform
JWT_AUDIENCE=veza-services
PORT=3001
HOST=0.0.0.0
EOF

# Stream Server .env
cat > "$CONFIG_DIR/stream.env" <<EOF
# Stream Server Rust Environment
DATABASE_URL=postgresql://veza:veza_password@$POSTGRES_IP:5432/veza_db
REDIS_URL=redis://:veza_redis_password@$REDIS_IP:6379
JWT_SECRET=veza_unified_jwt_secret_key_2025_microservices_secure_32chars_minimum
JWT_ISSUER=veza-platform
JWT_AUDIENCE=veza-services
PORT=3002
HOST=0.0.0.0
STORAGE_HOST=$STORAGE_IP
STORAGE_PATH=/storage/tracks
EOF

# Frontend .env
cat > "$CONFIG_DIR/frontend.env" <<EOF
# Frontend React Environment
VITE_API_URL=http://$BACKEND_IP:8080
VITE_CHAT_WS_URL=ws://$CHAT_IP:3001
VITE_STREAM_WS_URL=ws://$STREAM_IP:3002
VITE_PUBLIC_URL=http://$HAPROXY_IP
EOF

# Script pour copier les configs vers les containers
cat > "$CONFIG_DIR/deploy-configs.sh" <<'SCRIPT'
#!/bin/bash
# Déployer les configurations vers les containers

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "Déploiement des configurations..."

# Backend
incus file push "$SCRIPT_DIR/backend.env" veza-backend/app/.env

# Chat
incus file push "$SCRIPT_DIR/chat.env" veza-chat/app/.env

# Stream
incus file push "$SCRIPT_DIR/stream.env" veza-stream/app/.env

# Frontend
incus file push "$SCRIPT_DIR/frontend.env" veza-frontend/app/.env

echo "Configurations déployées!"
SCRIPT

chmod +x "$CONFIG_DIR/deploy-configs.sh"

# Afficher le résumé
echo ""
echo "Configuration générée avec succès!"
echo "================================="
echo "PostgreSQL : $POSTGRES_IP:5432"
echo "Redis      : $REDIS_IP:6379"
echo "Backend    : $BACKEND_IP:8080"
echo "Chat       : $CHAT_IP:3001"
echo "Stream     : $STREAM_IP:3002"
echo "Frontend   : $FRONTEND_IP:3000"
echo "HAProxy    : $HAPROXY_IP:80"
echo ""
echo "Fichiers générés:"
echo "  - $CONFIG_DIR/env.unified"
echo "  - $CONFIG_DIR/backend.env"
echo "  - $CONFIG_DIR/chat.env"
echo "  - $CONFIG_DIR/stream.env"
echo "  - $CONFIG_DIR/frontend.env"
echo ""
echo "Pour déployer: $CONFIG_DIR/deploy-configs.sh"
