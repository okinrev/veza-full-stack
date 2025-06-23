#!/bin/bash

set -e

echo "🚀 Déploiement du serveur de chat Veza (version simplifiée)"
echo "============================================================"

# Configuration
CONTAINER_NAME="veza-chat"
BINARY_NAME="chat-server"
PORT=3001

# Fonctions
build_server() {
    echo "📦 Compilation du serveur de chat..."
    cargo build --release --bin chat-server
    echo "✅ Compilation réussie"
}

deploy_to_container() {
    echo "🚢 Déploiement dans le container $CONTAINER_NAME..."
    
    # Copier le binaire
    incus file push target/release/chat-server $CONTAINER_NAME/opt/veza/
    
    # Rendre exécutable
    incus exec $CONTAINER_NAME -- chmod +x /opt/veza/chat-server
    
    # Créer le service systemd
    incus exec $CONTAINER_NAME -- tee /etc/systemd/system/veza-chat.service > /dev/null << 'EOF'
[Unit]
Description=Veza Chat Server
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/veza
ExecStart=/opt/veza/chat-server
Restart=always
RestartSec=10
Environment=RUST_LOG=info

[Install]
WantedBy=multi-user.target
EOF

    # Activer et démarrer le service
    incus exec $CONTAINER_NAME -- systemctl daemon-reload
    incus exec $CONTAINER_NAME -- systemctl enable veza-chat
    incus exec $CONTAINER_NAME -- systemctl restart veza-chat
    
    echo "✅ Service déployé et démarré"
}

test_deployment() {
    echo "🧪 Test du déploiement..."
    
    # Récupérer l'IP du container
    IP=$(incus list $CONTAINER_NAME -c 4 --format csv | cut -d' ' -f1)
    
    if [ -z "$IP" ]; then
        echo "❌ Impossible de récupérer l'IP du container"
        return 1
    fi
    
    echo "📡 Test de santé sur http://$IP:$PORT/health"
    
    # Attendre que le service démarre
    sleep 5
    
    # Test de l'endpoint de santé
    if curl -s "http://$IP:$PORT/health" | grep -q "healthy"; then
        echo "✅ Serveur de chat opérationnel sur $IP:$PORT"
        echo "📊 Endpoints disponibles :"
        echo "   - GET  http://$IP:$PORT/health"
        echo "   - GET  http://$IP:$PORT/api/messages?room=general"
        echo "   - POST http://$IP:$PORT/api/messages"
        echo "   - GET  http://$IP:$PORT/api/messages/stats"
        return 0
    else
        echo "❌ Le serveur ne répond pas correctement"
        echo "📝 Logs du service :"
        incus exec $CONTAINER_NAME -- journalctl -u veza-chat --no-pager -n 20
        return 1
    fi
}

# Vérifications préliminaires
if ! command -v incus &> /dev/null; then
    echo "❌ Incus non installé"
    exit 1
fi

if ! incus list | grep -q $CONTAINER_NAME; then
    echo "❌ Container $CONTAINER_NAME non trouvé"
    echo "📋 Containers disponibles :"
    incus list
    exit 1
fi

# Déploiement
echo "🎯 Déploiement vers le container : $CONTAINER_NAME"

build_server
deploy_to_container
test_deployment

echo ""
echo "🎉 Déploiement terminé avec succès !"
echo "📊 Pour tester l'API :"
echo "   curl http://$(incus list $CONTAINER_NAME -c 4 --format csv | cut -d' ' -f1):$PORT/health" 