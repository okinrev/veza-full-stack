#!/bin/bash

# Installation des dépendances dans les containers
set -e

echo "🔧 Installation des dépendances manquantes..."

# Backend Go - Installation de Go
echo "📦 Installation de Go dans veza-backend..."
incus exec veza-backend -- bash -c "
    if ! command -v go &> /dev/null; then
        echo 'Installation de Go...'
        cd /tmp
        wget -q https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
        tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
        echo 'export PATH=/usr/local/go/bin:\$PATH' >> /root/.bashrc
        echo 'export GOPATH=/opt/veza/go' >> /root/.bashrc
        mkdir -p /opt/veza/go
        echo 'Go installé avec succès'
    else
        echo 'Go déjà installé'
    fi
"

# Chat Rust - Installation de Rust
echo "🦀 Installation de Rust dans veza-chat..."
incus exec veza-chat -- bash -c "
    if ! command -v cargo &> /dev/null; then
        echo 'Installation de Rust...'
        curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
        echo 'export PATH=/root/.cargo/bin:\$PATH' >> /root/.bashrc
        source /root/.cargo/env
        echo 'Rust installé avec succès'
    else
        echo 'Rust déjà installé'
    fi
"

# Stream Rust - Installation de Rust
echo "🎵 Installation de Rust dans veza-stream..."
incus exec veza-stream -- bash -c "
    if ! command -v cargo &> /dev/null; then
        echo 'Installation de Rust...'
        curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
        echo 'export PATH=/root/.cargo/bin:\$PATH' >> /root/.bashrc
        source /root/.cargo/env
        echo 'Rust installé avec succès'
    else
        echo 'Rust déjà installé'
    fi
"

# Frontend React - Installation de Node.js
echo "⚛️ Installation de Node.js dans veza-frontend..."
incus exec veza-frontend -- bash -c "
    if ! command -v npm &> /dev/null; then
        echo 'Installation de Node.js...'
        curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
        apt-get install -y nodejs
        echo 'Node.js installé avec succès'
    else
        echo 'Node.js déjà installé'
    fi
"

echo "✅ Toutes les dépendances sont installées !" 