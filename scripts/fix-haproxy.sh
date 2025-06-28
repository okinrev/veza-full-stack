#!/bin/bash

echo "⚖️ Correction de la configuration HAProxy..."

# Récupérer les IPs depuis l'hôte
BACKEND_IP=$(incus ls veza-backend -c 4 --format csv | cut -d' ' -f1)
FRONTEND_IP=$(incus ls veza-frontend -c 4 --format csv | cut -d' ' -f1)

echo "📊 IPs détectées:"
echo "  Backend: $BACKEND_IP"
echo "  Frontend: $FRONTEND_IP"

# Configuration HAProxy avec les vraies IPs
incus exec veza-haproxy -- bash -c "
cat > /etc/haproxy/haproxy.cfg << 'EOF'
global
    daemon
    maxconn 4096
    log stdout local0

defaults
    mode http
    timeout connect 5000ms
    timeout client 50000ms
    timeout server 50000ms
    option httplog
    log global

# Interface de stats HAProxy
listen stats
    bind *:8404
    stats enable
    stats uri /stats
    stats refresh 30s
    stats admin if TRUE

# Frontend principal
frontend veza_main
    bind *:80
    
    # Redirection API vers backend
    acl is_api path_beg /api/
    use_backend veza_backend if is_api
    
    # Par défaut vers frontend React
    default_backend veza_frontend

# Backend pour l'API Go
backend veza_backend
    balance roundrobin
    server backend1 $BACKEND_IP:8080 check

# Backend pour le frontend React
backend veza_frontend
    balance roundrobin
    server frontend1 $FRONTEND_IP:3000 check
EOF

# Test de la configuration
echo 'Test de la configuration HAProxy...'
haproxy -c -f /etc/haproxy/haproxy.cfg

if [ \$? -eq 0 ]; then
    echo 'Configuration HAProxy valide'
    systemctl restart haproxy
    systemctl enable haproxy
    echo 'HAProxy redémarré avec succès'
else
    echo 'Erreur dans la configuration HAProxy'
    exit 1
fi
" 