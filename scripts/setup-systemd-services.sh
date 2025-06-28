#!/bin/bash

# Script de création des services systemd pour tous les composants Veza
# Services optimisés pour le développement avec rsync

set -e

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

echo -e "${BLUE}"
echo "╭──────────────────────────────────────────╮"
echo "│     🔧 Configuration Services systemd     │"
echo "╰──────────────────────────────────────────╯"
echo -e "${NC}"

# Configuration des services Backend Go
setup_backend_service() {
    echo -e "${CYAN}🔧 Configuration service Backend Go...${NC}"
    
    incus exec veza-backend -- bash -c '
        # Service systemd pour le backend Go
        cat > /etc/systemd/system/veza-backend.service << EOF
[Unit]
Description=Veza Backend API Go Server
After=network.target postgresql.service redis.service
Wants=postgresql.service redis.service
StartLimitIntervalSec=0

[Service]
Type=simple
Restart=always
RestartSec=1
User=root
WorkingDirectory=/opt/veza/backend
Environment=PATH=/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
Environment=GOPATH=/opt/veza/go
Environment=DATABASE_URL=postgres://veza_user:veza_password@veza-postgres:5432/veza_db?sslmode=disable
Environment=REDIS_URL=redis://veza-redis:6379
Environment=JWT_SECRET=your-super-secret-jwt-key-change-in-production
Environment=PORT=8080
Environment=UPLOADS_DIR=/storage/uploads
Environment=LOGS_DIR=/app/logs
ExecStartPre=/bin/mkdir -p /app/logs
ExecStart=/opt/veza/backend/main
StandardOutput=journal
StandardError=journal
SyslogIdentifier=veza-backend

[Install]
WantedBy=multi-user.target
EOF

        # Script de build
        cat > /opt/veza/backend/build.sh << '\''EOF'\''
#!/bin/bash
cd /opt/veza/backend
export PATH=/usr/local/go/bin:$PATH
export GOPATH=/opt/veza/go

echo "🔨 Building Veza Backend..."
go mod tidy
go build -o main ./cmd/server/main.go
echo "✅ Build completed"
EOF

        chmod +x /opt/veza/backend/build.sh
        
        # Activer le service (ne pas démarrer maintenant)
        systemctl daemon-reload
        systemctl enable veza-backend.service
    '
}

# Configuration des services Chat Rust
setup_chat_service() {
    echo -e "${CYAN}💬 Configuration service Chat Rust...${NC}"
    
    incus exec veza-chat -- bash -c '
        # Service systemd pour le serveur chat Rust
        cat > /etc/systemd/system/veza-chat.service << EOF
[Unit]
Description=Veza Chat Server Rust
After=network.target postgresql.service redis.service
Wants=postgresql.service redis.service
StartLimitIntervalSec=0

[Service]
Type=simple
Restart=always
RestartSec=1
User=root
WorkingDirectory=/opt/veza/chat
Environment=PATH=/root/.cargo/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
Environment=DATABASE_URL=postgres://veza_user:veza_password@veza-postgres:5432/veza_db?sslmode=disable
Environment=REDIS_URL=redis://veza-redis:6379
Environment=PORT=8081
Environment=RUST_LOG=info
Environment=LOGS_DIR=/app/logs
ExecStartPre=/bin/mkdir -p /app/logs
ExecStart=/opt/veza/chat/target/release/veza-chat-server
StandardOutput=journal
StandardError=journal
SyslogIdentifier=veza-chat

[Install]
WantedBy=multi-user.target
EOF

        # Script de build
        cat > /opt/veza/chat/build.sh << '\''EOF'\''
#!/bin/bash
cd /opt/veza/chat
export PATH=/root/.cargo/bin:$PATH

echo "🔨 Building Veza Chat Server..."
cargo build --release
echo "✅ Build completed"
EOF

        chmod +x /opt/veza/chat/build.sh
        
        # Activer le service
        systemctl daemon-reload
        systemctl enable veza-chat.service
    '
}

# Configuration des services Stream Rust
setup_stream_service() {
    echo -e "${CYAN}🎵 Configuration service Stream Rust...${NC}"
    
    incus exec veza-stream -- bash -c '
        # Service systemd pour le serveur stream Rust
        cat > /etc/systemd/system/veza-stream.service << EOF
[Unit]
Description=Veza Stream Server Rust
After=network.target postgresql.service redis.service nfs-common.service
Wants=postgresql.service redis.service nfs-common.service
StartLimitIntervalSec=0

[Service]
Type=simple
Restart=always
RestartSec=1
User=root
WorkingDirectory=/opt/veza/stream
Environment=PATH=/root/.cargo/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
Environment=DATABASE_URL=postgres://veza_user:veza_password@veza-postgres:5432/veza_db?sslmode=disable
Environment=REDIS_URL=redis://veza-redis:6379
Environment=PORT=8082
Environment=AUDIO_DIR=/storage/audio
Environment=RUST_LOG=info
Environment=LOGS_DIR=/app/logs
ExecStartPre=/bin/mkdir -p /app/logs
ExecStartPre=/bin/mkdir -p /storage/audio
ExecStart=/opt/veza/stream/target/release/veza-stream-server
StandardOutput=journal
StandardError=journal
SyslogIdentifier=veza-stream

[Install]
WantedBy=multi-user.target
EOF

        # Script de build
        cat > /opt/veza/stream/build.sh << '\''EOF'\''
#!/bin/bash
cd /opt/veza/stream
export PATH=/root/.cargo/bin:$PATH

echo "🔨 Building Veza Stream Server..."
cargo build --release
echo "✅ Build completed"
EOF

        chmod +x /opt/veza/stream/build.sh
        
        # Activer le service
        systemctl daemon-reload
        systemctl enable veza-stream.service
    '
}

# Configuration des services Frontend React
setup_frontend_service() {
    echo -e "${CYAN}⚛️ Configuration service Frontend React...${NC}"
    
    incus exec veza-frontend -- bash -c '
        # Service systemd pour le frontend React
        cat > /etc/systemd/system/veza-frontend.service << EOF
[Unit]
Description=Veza Frontend React Development Server
After=network.target
StartLimitIntervalSec=0

[Service]
Type=simple
Restart=always
RestartSec=1
User=root
WorkingDirectory=/opt/veza/frontend
Environment=PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
Environment=NODE_ENV=development
Environment=VITE_API_URL=http://veza-backend:8080/api/v1
Environment=VITE_WS_CHAT_URL=ws://10.5.191.108:3001/ws
Environment=VITE_WS_STREAM_URL=ws://veza-stream:8082/ws
Environment=LOGS_DIR=/app/logs
ExecStartPre=/bin/mkdir -p /app/logs
ExecStart=/usr/bin/npm run dev -- --host 0.0.0.0 --port 3000
StandardOutput=journal
StandardError=journal
SyslogIdentifier=veza-frontend

[Install]
WantedBy=multi-user.target
EOF

        # Script de build et installation des dépendances
        cat > /opt/veza/frontend/build.sh << '\''EOF'\''
#!/bin/bash
cd /opt/veza/frontend

echo "📦 Installing dependencies..."
npm install

echo "🔨 Building Veza Frontend..."
npm run build
echo "✅ Build completed"
EOF

        # Script de démarrage du dev server
        cat > /opt/veza/frontend/dev.sh << '\''EOF'\''
#!/bin/bash
cd /opt/veza/frontend

echo "📦 Installing dependencies..."
npm install

echo "🚀 Starting dev server..."
npm run dev -- --host 0.0.0.0 --port 3000
EOF

        chmod +x /opt/veza/frontend/build.sh
        chmod +x /opt/veza/frontend/dev.sh
        
        # Activer le service
        systemctl daemon-reload
        systemctl enable veza-frontend.service
    '
}

# Configuration des services HAProxy
setup_haproxy_service() {
    echo -e "${CYAN}⚖️ Configuration service HAProxy...${NC}"
    
    incus exec veza-haproxy -- bash -c '
        # Configuration HAProxy
        cat > /etc/haproxy/haproxy.cfg << EOF
global
    daemon
    maxconn 4096
    log stdout local0
    chroot /var/lib/haproxy
    stats socket /run/haproxy/admin.sock mode 660 level admin
    stats timeout 30s
    user haproxy
    group haproxy

defaults
    mode http
    timeout connect 5000ms
    timeout client 50000ms
    timeout server 50000ms
    option httplog
    log global

# Frontend principal
frontend veza_frontend
    bind *:80
    # Redirection des API vers le backend
    acl is_api path_beg /api/
    use_backend veza_backend if is_api
    
    # Redirection des WebSockets chat
    acl is_chat_ws path_beg /ws/chat
    use_backend veza_chat if is_chat_ws
    
    # Redirection des WebSockets stream
    acl is_stream_ws path_beg /ws/stream
    use_backend veza_stream if is_stream_ws
    
    # Par défaut, frontend React
    default_backend veza_frontend_app

# Backend pour l'\''API Go
backend veza_backend
    balance roundrobin
    server backend1 veza-backend:8080 check

# Backend pour le chat Rust
backend veza_chat
    balance roundrobin
    option forwardfor
    http-request set-header X-Forwarded-Proto https if { ssl_fc }
    server chat1 veza-chat:8081 check

# Backend pour le stream Rust
backend veza_stream
    balance roundrobin
    option forwardfor
    http-request set-header X-Forwarded-Proto https if { ssl_fc }
    server stream1 veza-stream:8082 check

# Backend pour le frontend React
backend veza_frontend_app
    balance roundrobin
    server frontend1 veza-frontend:3000 check

# Interface de statistiques
listen stats
    bind *:8404
    stats enable
    stats uri /stats
    stats refresh 30s
    stats admin if TRUE
EOF

        # Redémarrer HAProxy
        systemctl restart haproxy
        systemctl enable haproxy
    '
}

# Scripts de gestion
create_management_scripts() {
    echo -e "${CYAN}📜 Création des scripts de gestion...${NC}"
    
    # Script pour démarrer tous les services
    cat > scripts/start-all-services.sh << 'EOF'
#!/bin/bash

echo "🚀 Démarrage de tous les services Veza..."

# Démarrer les services dans l'ordre
echo "📊 Démarrage PostgreSQL et Redis..."
incus exec veza-postgres -- systemctl start postgresql
incus exec veza-redis -- systemctl start redis-server

echo "🗄️ Démarrage du storage..."
incus exec veza-storage -- systemctl start nfs-kernel-server

sleep 3

echo "🔧 Démarrage du backend..."
incus exec veza-backend -- systemctl start veza-backend

echo "💬 Démarrage du chat..."
incus exec veza-chat -- systemctl start veza-chat

echo "🎵 Démarrage du stream..."
incus exec veza-stream -- systemctl start veza-stream

echo "⚛️ Démarrage du frontend..."
incus exec veza-frontend -- systemctl start veza-frontend

echo "⚖️ Démarrage de HAProxy..."
incus exec veza-haproxy -- systemctl restart haproxy

echo "✅ Tous les services sont démarrés!"
echo ""
echo "🌐 Accès:"
echo "  - Frontend: http://$(incus ls veza-haproxy -c 4 --format csv | cut -d' ' -f1)"
echo "  - Stats HAProxy: http://$(incus ls veza-haproxy -c 4 --format csv | cut -d' ' -f1):8404/stats"
EOF

    # Script pour arrêter tous les services
    cat > scripts/stop-all-services.sh << 'EOF'
#!/bin/bash

echo "🛑 Arrêt de tous les services Veza..."

incus exec veza-haproxy -- systemctl stop haproxy
incus exec veza-frontend -- systemctl stop veza-frontend
incus exec veza-stream -- systemctl stop veza-stream
incus exec veza-chat -- systemctl stop veza-chat
incus exec veza-backend -- systemctl stop veza-backend
incus exec veza-storage -- systemctl stop nfs-kernel-server
incus exec veza-redis -- systemctl stop redis-server
incus exec veza-postgres -- systemctl stop postgresql

echo "✅ Tous les services sont arrêtés!"
EOF

    # Script pour voir le statut
    cat > scripts/status-all-services.sh << 'EOF'
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
if [ -n "$haproxy_ip" ]; then
    echo "  - Application: http://$haproxy_ip"
    echo "  - HAProxy Stats: http://$haproxy_ip:8404/stats"
fi
EOF

    chmod +x scripts/start-all-services.sh
    chmod +x scripts/stop-all-services.sh
    chmod +x scripts/status-all-services.sh
}

# Fonction principale
main() {
    echo -e "${CYAN}🔧 Configuration des services systemd...${NC}"
    echo ""
    
    # Vérifier que les containers existent
    containers=("veza-backend" "veza-chat" "veza-stream" "veza-frontend" "veza-haproxy")
    for container in "${containers[@]}"; do
        if ! incus list "$container" --format csv | grep -q "$container"; then
            echo -e "${RED}❌ Container $container non trouvé${NC}"
            echo -e "${YELLOW}Exécutez d'abord: ./scripts/setup-manual-containers.sh${NC}"
            exit 1
        fi
    done
    
    # Configurer tous les services
    setup_backend_service
    setup_chat_service
    setup_stream_service
    setup_frontend_service
    setup_haproxy_service
    
    # Créer les scripts de gestion
    create_management_scripts
    
    echo ""
    echo -e "${GREEN}🎉 Tous les services systemd sont configurés !${NC}"
    echo ""
    echo -e "${BLUE}💡 Commandes utiles:${NC}"
    echo -e "  - Démarrer tous: ${YELLOW}./scripts/start-all-services.sh${NC}"
    echo -e "  - Arrêter tous: ${YELLOW}./scripts/stop-all-services.sh${NC}"
    echo -e "  - Voir statut: ${YELLOW}./scripts/status-all-services.sh${NC}"
    echo ""
    echo -e "${CYAN}📝 Prochaine étape: Configuration rsync${NC}"
    echo -e "  ${YELLOW}./scripts/setup-rsync.sh${NC}"
}

main "$@" 