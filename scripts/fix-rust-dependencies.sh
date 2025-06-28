#!/bin/bash

echo "🔧 Installation des dépendances OpenSSL pour Rust..."

# Chat container - OpenSSL dependencies
echo "💬 Installation OpenSSL dans veza-chat..."
incus exec veza-chat -- bash -c "
    export DEBIAN_FRONTEND=noninteractive
    apt-get update -qq
    apt-get install -y libssl-dev pkg-config
    echo 'Dépendances OpenSSL installées pour chat ✅'
"

# Stream container - OpenSSL dependencies
echo "🎵 Installation OpenSSL dans veza-stream..."
incus exec veza-stream -- bash -c "
    export DEBIAN_FRONTEND=noninteractive
    apt-get update -qq
    apt-get install -y libssl-dev pkg-config
    echo 'Dépendances OpenSSL installées pour stream ✅'
"

echo "✅ Toutes les dépendances OpenSSL sont installées !" 