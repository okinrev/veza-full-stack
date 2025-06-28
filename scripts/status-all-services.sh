#!/bin/bash

echo "📊 Statut des services Veza..."
echo ""

containers=("veza-postgres" "veza-redis" "veza-storage" "veza-backend" "veza-chat" "veza-stream" "veza-frontend" "veza-haproxy")
services=("postgresql" "redis-server" "nfs-kernel-server" "veza-backend" "veza-chat" "veza-stream" "veza-frontend" "haproxy")

for i in "${!containers[@]}"; do
    container="${containers[$i]}"
    service="${services[$i]}"
    
    echo -n "📦 $container ($service): "
    status=$(incus exec "$container" -- systemctl is-active "$service" 2>/dev/null || echo "inactive")
    
    case $status in
        "active")
            echo -e "\033[0;32m✅ Active\033[0m"
            ;;
        "inactive"|"failed")
            echo -e "\033[0;31m❌ Inactive\033[0m"
            ;;
        *)
            echo -e "\033[0;33m⚠️ $status\033[0m"
            ;;
    esac
done

echo ""
echo "🌐 URLs d'accès:"
haproxy_ip=$(incus ls veza-haproxy -c 4 --format csv | cut -d' ' -f1)
frontend_ip=$(incus ls veza-frontend -c 4 --format csv | cut -d' ' -f1)
backend_ip=$(incus ls veza-backend -c 4 --format csv | cut -d' ' -f1)

if [ -n "$haproxy_ip" ]; then
    echo "  - Application: http://$haproxy_ip"
    echo "  - HAProxy Stats: http://$haproxy_ip:8404/stats"
fi
if [ -n "$frontend_ip" ]; then
    echo "  - Frontend direct: http://$frontend_ip:3000"
fi
if [ -n "$backend_ip" ]; then
    echo "  - Backend API: http://$backend_ip:8080"
fi 