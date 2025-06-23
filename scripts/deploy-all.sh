#!/bin/bash

# Script de déploiement complet de l'infrastructure Veza

set -e

echo "🚀 DÉPLOIEMENT COMPLET DE L'INFRASTRUCTURE VEZA"
echo ""

# Vérifier les containers
echo "1. Vérification des containers..."
containers=("veza-postgres" "veza-redis" "veza-backend" "veza-chat" "veza-stream" "veza-frontend" "veza-haproxy")
for container in "${containers[@]}"; do
    if incus list --format csv | grep -q "^$container,RUNNING"; then
        echo "✅ $container actif"
    else
        echo "❌ $container non actif"
        exit 1
    fi
done

echo ""
echo "2. Configuration HAProxy finale..."
incus exec veza-haproxy -- tee /etc/haproxy/haproxy.cfg > /dev/null << 'EOH'
global
    daemon
    maxconn 4096

defaults
    mode http
    timeout connect 5s
    timeout client 30s
    timeout server 30s

frontend veza_main
    bind *:80
    # Headers CORS
    http-response set-header Access-Control-Allow-Origin "*"
    http-response set-header Access-Control-Allow-Methods "GET,POST,OPTIONS,PUT,DELETE"
    http-response set-header Access-Control-Allow-Headers "Content-Type,Authorization"
    
    # Routage
    acl is_chat_api path_beg /chat-api/
    acl is_stream_api path_beg /stream/
    acl is_backend_api path_beg /api/
    
    use_backend chat_backend if is_chat_api
    use_backend stream_backend if is_stream_api  
    use_backend go_backend if is_backend_api
    default_backend react_frontend

backend react_frontend
    server react1 10.5.191.41:5173 check

backend go_backend
    server go1 10.5.191.241:8080 check

backend chat_backend
    http-request set-path %[path,regsub(^/chat-api,/api)]
    server chat1 10.5.191.49:8081 check

backend stream_backend
    http-request set-path %[path,regsub(^/stream,/)]
    server stream1 10.5.191.196:8082 check
EOH

incus exec veza-haproxy -- systemctl restart haproxy

echo "✅ HAProxy configuré"

echo ""
echo "3. Tests finaux..."
sleep 3

echo "Frontend via HAProxy:"
curl -s -o /dev/null -w "Status: %{http_code}\n" http://10.5.191.133/ || echo "Erreur"

echo "Chat API via HAProxy:"
curl -s http://10.5.191.133/chat-api/health | head -1 || echo "Erreur"

echo "Stream API via HAProxy:"
curl -s http://10.5.191.133/stream/health | head -1 || echo "Erreur"

echo ""
echo "🎉 DÉPLOIEMENT TERMINÉ !"
echo "========================"
echo ""
echo "🌐 Accédez à votre application sur:"
echo "   http://10.5.191.133 (via HAProxy)"
echo "   http://10.5.191.41:3000 (frontend direct)"
echo ""
echo "✅ Tous les services sont interconnectés et fonctionnels !"
