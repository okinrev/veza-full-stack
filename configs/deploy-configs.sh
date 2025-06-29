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
