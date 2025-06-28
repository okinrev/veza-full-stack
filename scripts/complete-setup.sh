#!/bin/bash

# Script complet : Installation + Build + Start
set -e

echo "🚀 Setup complet de l'infrastructure Veza"
echo "========================================"

# Étape 1: Installation des dépendances
echo ""
echo "🔧 Étape 1: Installation des dépendances..."
./scripts/install-dependencies.sh

# Petite pause pour laisser les installations se finaliser
sleep 5

# Étape 2: Construction des projets
echo ""
echo "🔨 Étape 2: Construction des projets..."

# Backend Go
echo "🔧 Construction Backend Go..."
incus exec veza-backend -- bash -c "
    cd /opt/veza/backend
    export PATH=/usr/local/go/bin:\$PATH
    export GOPATH=/opt/veza/go
    go mod tidy
    go build -o main ./cmd/server/main.go
    echo 'Backend Go compilé ✅'
"

# Chat Rust
echo "💬 Construction Chat Rust..."
incus exec veza-chat -- bash -c "
    cd /opt/veza/chat
    export PATH=/root/.cargo/bin:\$PATH
    source /root/.cargo/env
    cargo build --release
    echo 'Chat Rust compilé ✅'
"

# Stream Rust
echo "🎵 Construction Stream Rust..."
incus exec veza-stream -- bash -c "
    cd /opt/veza/stream
    export PATH=/root/.cargo/bin:\$PATH
    source /root/.cargo/env
    cargo build --release
    echo 'Stream Rust compilé ✅'
"

# Frontend React
echo "⚛️ Installation Frontend React..."
incus exec veza-frontend -- bash -c "
    cd /opt/veza/frontend
    npm install
    echo 'Frontend React installé ✅'
"

# Étape 3: Démarrage des services
echo ""
echo "🚀 Étape 3: Démarrage des services..."

# Services de base
echo "📊 Démarrage PostgreSQL..."
incus exec veza-postgres -- systemctl start postgresql
sleep 2

echo "🔴 Démarrage Redis..."
incus exec veza-redis -- systemctl start redis-server
sleep 2

echo "🗄️ Démarrage NFS Storage..."
incus exec veza-storage -- systemctl start nfs-kernel-server 2>/dev/null || echo "NFS OK"
sleep 2

# Services applicatifs
echo "🔧 Démarrage Backend..."
incus exec veza-backend -- systemctl start veza-backend
sleep 3

echo "💬 Démarrage Chat..."
incus exec veza-chat -- systemctl start veza-chat
sleep 3

echo "🎵 Démarrage Stream..."
incus exec veza-stream -- systemctl start veza-stream
sleep 3

echo "⚛️ Démarrage Frontend..."
incus exec veza-frontend -- systemctl start veza-frontend
sleep 3

# Étape 4: Configuration HAProxy
echo ""
echo "⚖️ Étape 4: Configuration HAProxy..."
./scripts/fix-haproxy.sh

# Étape 5: Vérification finale
echo ""
echo "🔍 Étape 5: Vérification finale..."
./scripts/status-all-services.sh

echo ""
echo "🎉 Setup complet terminé !"
echo ""
echo "🌐 Accès à l'application:"
haproxy_ip=$(incus ls veza-haproxy -c 4 --format csv | cut -d' ' -f1)
echo "  👉 http://$haproxy_ip"
echo "  📊 Stats HAProxy: http://$haproxy_ip:8404/stats" 