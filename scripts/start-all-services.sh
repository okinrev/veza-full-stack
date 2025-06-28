#!/bin/bash

# Script de démarrage de tous les services Veza
set -e

echo "🚀 Démarrage de tous les services Veza..."
echo ""

# Services de base
echo "📊 Démarrage PostgreSQL..."
incus exec veza-postgres -- systemctl start postgresql || echo "PostgreSQL déjà démarré"
sleep 2

echo "🔴 Démarrage Redis..."
incus exec veza-redis -- systemctl start redis-server || echo "Redis déjà démarré"  
sleep 2

echo "🗄️ Démarrage NFS..."
incus exec veza-storage -- systemctl start nfs-kernel-server || echo "NFS en cours de démarrage..."
sleep 3

# Services applicatifs
echo "🔧 Démarrage Backend Go..."
incus exec veza-backend -- systemctl restart veza-backend
sleep 3

echo "💬 Démarrage Chat Rust..."
incus exec veza-chat -- systemctl restart veza-chat
sleep 3

echo "🎵 Démarrage Stream Rust..."
incus exec veza-stream -- systemctl restart veza-stream
sleep 3

echo "⚛️ Démarrage Frontend React..."
incus exec veza-frontend -- systemctl restart veza-frontend
sleep 3

echo "⚖️ Démarrage HAProxy..."
incus exec veza-haproxy -- systemctl restart haproxy
sleep 2

echo ""
echo "⏳ Attente de stabilisation des services..."
sleep 10

echo ""
echo "🔍 Vérification finale..."
./scripts/status-all-services.sh

echo ""
echo "🎉 Infrastructure Veza prête !"
echo ""
echo "🌐 Accès principal:"
haproxy_ip=$(incus ls veza-haproxy -c 4 --format csv | cut -d' ' -f1)
echo "  👉 Application: http://$haproxy_ip"
echo "  📊 HAProxy Stats: http://$haproxy_ip:8404/stats" 