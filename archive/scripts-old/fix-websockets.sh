#!/bin/bash

echo "🔧 Correction et diagnostic WebSocket Veza"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# 1. Vérification de la configuration HAProxy
echo ""
echo "📋 1. Vérification configuration HAProxy WebSocket:"
if incus exec veza-haproxy -- grep -q "server websocket1 10.5.191.241:8080" /etc/haproxy/haproxy.cfg; then
    echo "  ✅ WebSocket correctement routé vers backend Go (10.5.191.241:8080)"
else
    echo "  🔧 Correction routage WebSocket..."
    
    # Sauvegarde et correction
    cp configs/haproxy.cfg temp_haproxy_ws.cfg
    sed -i 's/server websocket1 10.5.191.49:3001/server websocket1 10.5.191.241:8080/' temp_haproxy_ws.cfg
    sed -i 's/server websocket1 10.5.191.241:9001/server websocket1 10.5.191.241:8080/' temp_haproxy_ws.cfg
    
    incus file push temp_haproxy_ws.cfg veza-haproxy/etc/haproxy/haproxy.cfg
    incus exec veza-haproxy -- systemctl reload haproxy
    rm temp_haproxy_ws.cfg
    echo "  ✅ HAProxy WebSocket corrigé et rechargé"
fi

# 2. Test des services
echo ""
echo "📡 2. Test des services WebSocket:"

# Test backend Go WebSocket
WS_STATUS=$(curl -m 3 -s -o /dev/null -w '%{http_code}' "http://10.5.191.241:8080/ws/chat" 2>/dev/null || echo "000")
if [ "$WS_STATUS" = "401" ]; then
    echo "  ✅ Backend Go WebSocket: OK (401 = authentification requise)"
elif [ "$WS_STATUS" = "404" ]; then
    echo "  ❌ Backend Go WebSocket: Route non trouvée (404)"
else
    echo "  ⚠️ Backend Go WebSocket: Status $WS_STATUS"
fi

# Test via HAProxy
HAP_STATUS=$(curl -m 3 -s -o /dev/null -w '%{http_code}' "http://10.5.191.133/ws/chat" 2>/dev/null || echo "000")
if [ "$HAP_STATUS" = "401" ]; then
    echo "  ✅ HAProxy WebSocket: OK (401 = authentification requise)"
elif [ "$HAP_STATUS" = "404" ]; then
    echo "  ❌ HAProxy WebSocket: Route non trouvée (404)"
else
    echo "  ⚠️ HAProxy WebSocket: Status $HAP_STATUS"
fi

# 3. Vérification des processus
echo ""
echo "🔍 3. Vérification des processus:"

BACKEND_PROC=$(incus exec veza-backend -- ps aux | grep -c "/opt/veza/server" || echo "0")
if [ "$BACKEND_PROC" -gt "0" ]; then
    echo "  ✅ Backend Go actif ($BACKEND_PROC processus)"
else
    echo "  ❌ Backend Go non actif"
    echo "  🔧 Redémarrage du backend..."
    incus exec veza-backend -- systemctl restart veza-api 2>/dev/null || echo "  ⚠️ Pas de service systemd"
fi

# 4. Test de l'application
echo ""
echo "🌐 4. Test application web:"
APP_STATUS=$(curl -m 3 -s -o /dev/null -w '%{http_code}' "http://10.5.191.133" 2>/dev/null || echo "000")
if [ "$APP_STATUS" = "200" ]; then
    echo "  ✅ Application web accessible (200)"
else
    echo "  ❌ Application web: Status $APP_STATUS"
fi

# 5. Conseils
echo ""
echo "💡 5. Instructions pour tester:"
echo "  • Rafraîchissez votre navigateur (F5 ou Ctrl+R)"
echo "  • Les erreurs WebSocket 401 sont normales - l'auth est gérée côté frontend"
echo "  • URL application: http://10.5.191.133"
echo "  • Les WebSockets se connectent après login réussi"

echo ""
echo "🎉 Correction WebSocket terminée !" 