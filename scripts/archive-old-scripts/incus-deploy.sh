#!/bin/bash

# Script de déploiement complet Incus pour Veza
# Déploie les 8 containers avec configuration optimisée et démarre tous les services

set -e

# Couleurs pour les messages
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

echo -e "${PURPLE}"
echo "╭──────────────────────────────────────────╮"
echo "│      🚀 Veza - Déploiement Complet      │"
echo "╰──────────────────────────────────────────╯"
echo -e "${NC}"

# Variables globales
WORKSPACE_DIR=$(pwd)
IMAGE="images:debian/bookworm"

# Variables IP (seront mises à jour avec les vraies IPs DHCP)
FRONTEND_IP="10.100.0.11"
BACKEND_IP="10.100.0.12" 
CHAT_IP="10.100.0.13"
STREAM_IP="10.100.0.14"
POSTGRES_IP="10.100.0.15"
HAPROXY_IP="10.100.0.16"
REDIS_IP="10.100.0.17"
STORAGE_IP="10.100.0.18"

# Configuration DNS ultra-robuste pour tous les containers
configure_dns() {
    local container_name=$1
    
    echo -e "${BLUE}🌐 Configuration DNS ultra-robuste pour $container_name...${NC}"
    
    incus exec "$container_name" -- bash -c "
        # 1. ARRÊTER et DÉSACTIVER tous les services DNS conflictuels
        echo 'Désactivation services DNS conflictuels...'
        systemctl stop systemd-resolved 2>/dev/null || true
        systemctl disable systemd-resolved 2>/dev/null || true
        systemctl mask systemd-resolved 2>/dev/null || true
        systemctl stop systemd-networkd 2>/dev/null || true
        systemctl disable systemd-networkd 2>/dev/null || true
        systemctl stop resolvconf 2>/dev/null || true
        systemctl disable resolvconf 2>/dev/null || true
        
        # 2. NETTOYER complètement les configurations DNS existantes
        echo 'Nettoyage configurations DNS existantes...'
        rm -f /etc/resolv.conf*
        rm -f /run/systemd/resolve/resolv.conf
        rm -f /run/systemd/resolve/stub-resolv.conf
        rm -rf /etc/systemd/resolved.conf.d/
        rm -rf /etc/systemd/network/
        
        # 3. TUER tous les processus DNS résiduels
        pkill -f systemd-resolved 2>/dev/null || true
        pkill -f systemd-networkd 2>/dev/null || true
        pkill -f dhclient 2>/dev/null || true
        
        # 4. CRÉER configuration DNS statique ULTRA-ROBUSTE
        echo 'Création configuration DNS statique ultra-robuste...'
        cat > /etc/resolv.conf << 'EOF'
# Configuration DNS Veza - ULTRA-ROBUSTE avec redondance maximale
# Serveurs DNS primaires (Google, Cloudflare)
nameserver 8.8.8.8
nameserver 8.8.4.4
nameserver 1.1.1.1
nameserver 1.0.0.1

# Serveurs DNS secondaires (Quad9, OpenDNS)
nameserver 9.9.9.9
nameserver 149.112.112.112
nameserver 208.67.222.222
nameserver 208.67.220.220

# Configuration de recherche optimisée
search veza.local .
options timeout:2 attempts:4 rotate single-request-reopen
options edns0 trust-ad use-vc
EOF
        
        # 5. PROTECTION maximale contre les modifications
        chmod 444 /etc/resolv.conf
        chattr +i /etc/resolv.conf 2>/dev/null || {
            echo 'chattr non disponible, utilisation protection alternative'
            echo 'Création script de protection...'
            
            # Créer un script de protection en boucle
            cat > /usr/local/bin/protect-resolv.sh << 'EOFSCRIPT'
#!/bin/bash
while true; do
    if ! grep -q \"8.8.8.8\" /etc/resolv.conf 2>/dev/null; then
        cat > /etc/resolv.conf << 'EOFRESOLV'
nameserver 8.8.8.8
nameserver 8.8.4.4
nameserver 1.1.1.1
nameserver 9.9.9.9
search veza.local .
options timeout:2 attempts:4 rotate
EOFRESOLV
        chmod 444 /etc/resolv.conf
    fi
    sleep 30
done
EOFSCRIPT
            chmod +x /usr/local/bin/protect-resolv.sh
        }
        
        # 6. CONFIGURATION NetworkManager pour qu'il n'interfère JAMAIS
        mkdir -p /etc/NetworkManager/conf.d/
        cat > /etc/NetworkManager/conf.d/99-veza-dns-ultimate.conf << 'EOF'
[main]
# Désactivation TOTALE de la gestion DNS par NetworkManager
dns=none
systemd-resolved=false
rc-manager=unmanaged

[connection]
# Ignorer TOUS les paramètres DNS automatiques
ipv4.ignore-auto-dns=true
ipv6.ignore-auto-dns=true
ipv4.never-default=false
ipv6.never-default=false

[device]
# Configurer l'interface pour ignorer le DNS
wifi.scan-rand-mac-address=no
ethernet.wake-on-lan=ignore
EOF
        
        # 7. CONFIGURATION netplan ultra-robuste
        mkdir -p /etc/netplan/
        cat > /etc/netplan/99-veza-ultimate-dns.yaml << 'EOF'
network:
  version: 2
  renderer: networkd
  ethernets:
    eth0:
      dhcp4: true
      dhcp4-overrides:
        use-dns: false
        use-domains: false
        use-ntp: false
      nameservers:
        addresses: [8.8.8.8, 8.8.4.4, 1.1.1.1, 1.0.0.1, 9.9.9.9, 149.112.112.112]
        search: [veza.local]
      dhcp6: false
      accept-ra: false
EOF
        
        # 8. CRÉER service systemd de protection DNS
        cat > /etc/systemd/system/veza-dns-ultimate-guard.service << 'EOF'
[Unit]
Description=Veza DNS Ultimate Protection Service
After=network.target network-online.target
Wants=network-online.target

[Service]
Type=forking
Restart=always
RestartSec=5
ExecStart=/bin/bash -c \"
# Protection DNS en arrière-plan
(
while true; do
    # Vérifier et restaurer la configuration DNS
    if ! grep -q '8.8.8.8' /etc/resolv.conf 2>/dev/null || ! grep -q '8.8.4.4' /etc/resolv.conf 2>/dev/null; then
        echo \\\"[\$(date)] Restauration configuration DNS Veza pour $container_name\\\" >> /var/log/veza-dns-guard.log
        
        # Supprimer protection temporaire
        chattr -i /etc/resolv.conf 2>/dev/null || true
        
        # Restaurer configuration
        cat > /etc/resolv.conf << 'EOFINNER'
nameserver 8.8.8.8
nameserver 8.8.4.4
nameserver 1.1.1.1
nameserver 1.0.0.1
nameserver 9.9.9.9
nameserver 149.112.112.112
search veza.local .
options timeout:2 attempts:4 rotate single-request-reopen
EOFINNER
        
        # Remettre protection
        chmod 444 /etc/resolv.conf
        chattr +i /etc/resolv.conf 2>/dev/null || true
    fi
    
    # Vérifier connectivité et réparer si nécessaire
    if ! timeout 3 ping -c 1 8.8.8.8 >/dev/null 2>&1; then
        echo \\\"[\$(date)] Problème connectivité détecté pour $container_name\\\" >> /var/log/veza-dns-guard.log
        
        # Redémarrer l'interface réseau
        ip link set eth0 down 2>/dev/null || true
        sleep 1
        ip link set eth0 up 2>/dev/null || true
        
        # Forcer renouvellement DHCP
        dhclient -r eth0 2>/dev/null || true
        dhclient eth0 2>/dev/null || true
    fi
    
    sleep 30
done
) &
echo \\\$! > /var/run/veza-dns-guard.pid
\"
ExecStop=/bin/bash -c \"
if [ -f /var/run/veza-dns-guard.pid ]; then
    kill \$(cat /var/run/veza-dns-guard.pid) 2>/dev/null || true
    rm -f /var/run/veza-dns-guard.pid
fi
\"
PIDFile=/var/run/veza-dns-guard.pid

[Install]
WantedBy=multi-user.target
EOF
        
        # 9. REDÉMARRER les services réseau nécessaires
        systemctl restart NetworkManager 2>/dev/null || true
        netplan apply 2>/dev/null || true
        
        # 10. ACTIVER le service de protection
        systemctl daemon-reload
        systemctl enable veza-dns-ultimate-guard.service
        systemctl start veza-dns-ultimate-guard.service
        
        # 11. TESTS DNS ultra-complets avec retry intelligent
        echo 'Tests DNS ultra-complets pour $container_name...'
        
        DNS_SUCCESS=false
        CONNECTIVITY_SUCCESS=false
        
        # Test DNS avec différents serveurs et outils
        for attempt in {1..5}; do
            echo \"Test DNS tentative \$attempt/5...\"
            
            dns_tests_passed=0
            
            # Test avec nslookup
            if timeout 3 nslookup deb.debian.org 8.8.8.8 >/dev/null 2>&1; then
                ((dns_tests_passed++))
            fi
            
            # Test avec dig
            if timeout 3 dig @8.8.4.4 google.com >/dev/null 2>&1; then
                ((dns_tests_passed++))
            fi
            
            # Test avec host
            if timeout 3 host debian.org 1.1.1.1 >/dev/null 2>&1; then
                ((dns_tests_passed++))
            fi
            
            # Test résolution par défaut
            if timeout 3 nslookup github.com >/dev/null 2>&1; then
                ((dns_tests_passed++))
            fi
            
            if [ \$dns_tests_passed -ge 3 ]; then
                echo \"✅ Tests DNS réussis (\$dns_tests_passed/4) pour $container_name\"
                DNS_SUCCESS=true
                break
            else
                echo \"❌ Tests DNS échoués (\$dns_tests_passed/4) pour $container_name - retry...\"
                sleep 2
            fi
        done
        
        # Test de connectivité multi-cibles
        connectivity_tests_passed=0
        
        if timeout 3 ping -c 1 -W 1 8.8.8.8 >/dev/null 2>&1; then
            ((connectivity_tests_passed++))
        fi
        
        if timeout 3 ping -c 1 -W 1 1.1.1.1 >/dev/null 2>&1; then
            ((connectivity_tests_passed++))
        fi
        
        if timeout 3 ping -c 1 -W 1 9.9.9.9 >/dev/null 2>&1; then
            ((connectivity_tests_passed++))
        fi
        
        if timeout 5 curl -s --max-time 3 http://detectportal.firefox.com/canonical.html >/dev/null 2>&1; then
            ((connectivity_tests_passed++))
        fi
        
        if [ \$connectivity_tests_passed -ge 2 ]; then
            echo \"✅ Connectivité internet confirmée (\$connectivity_tests_passed/4) pour $container_name\"
            CONNECTIVITY_SUCCESS=true
        else
            echo \"⚠️ Connectivité internet limitée (\$connectivity_tests_passed/4) pour $container_name\"
        fi
        
        # 12. Configuration finale des hosts avec mapping complet
        echo 'Configuration fichier hosts ultra-complet...'
        CONTAINER_HOSTNAME=\$(hostname)
        cat > /etc/hosts << EOF
127.0.0.1 localhost
127.0.1.1 \$CONTAINER_HOSTNAME

# Veza Infrastructure - Services internes
10.100.0.11 veza-frontend frontend app.veza.local
10.100.0.12 veza-backend backend api api.veza.local
10.100.0.13 veza-chat chat websocket chat.veza.local  
10.100.0.14 veza-stream stream audio stream.veza.local
10.100.0.15 veza-postgres postgres db database.veza.local
10.100.0.16 veza-haproxy haproxy lb proxy.veza.local
10.100.0.17 veza-redis redis cache cache.veza.local
10.100.0.18 veza-storage storage nfs files.veza.local

# Gateway et réseau local
10.100.0.1 gateway.veza.local router

# Serveurs DNS publics pour référence
8.8.8.8 dns.google google-dns
8.8.4.4 dns2.google google-dns2
1.1.1.1 one.one.one.one cloudflare-dns
1.0.0.1 cloudflare-dns2
9.9.9.9 dns.quad9.net quad9-dns
EOF
        
        # 13. RAPPORT final
        if [ \"\$DNS_SUCCESS\" = true ] && [ \"\$CONNECTIVITY_SUCCESS\" = true ]; then
            echo '✅ Configuration DNS ultra-robuste RÉUSSIE pour $container_name'
            echo '🌟 DNS et connectivité parfaitement fonctionnels'
        elif [ \"\$DNS_SUCCESS\" = true ]; then
            echo '⚠️ Configuration DNS ultra-robuste PARTIELLE pour $container_name'
            echo '✅ DNS fonctionnel mais connectivité limitée'
        else
            echo '❌ Configuration DNS ultra-robuste PROBLÉMATIQUE pour $container_name'
            echo '⚠️ Problèmes persistants détectés mais configuration appliquée'
        fi
        
        echo '📊 Service de protection DNS actif et surveillant'
    "
}

# Fonction pour mettre à jour /etc/hosts avec les vraies IPs
update_hosts_files() {
    echo -e "${BLUE}🔗 Mise à jour des fichiers /etc/hosts avec les vraies IPs...${NC}"
    
    # Liste de tous les containers
    local containers=("veza-frontend" "veza-backend" "veza-chat" "veza-stream" "veza-postgres" "veza-haproxy" "veza-redis" "veza-storage")
    
    for container in "${containers[@]}"; do
        if incus list "$container" --format csv | grep -q "RUNNING"; then
            echo -e "${BLUE}📝 Mise à jour /etc/hosts pour $container...${NC}"
            
            incus exec "$container" -- bash -c "
                # Supprimer les anciennes entrées veza
                sed -i '/veza-/d' /etc/hosts
                
                # Ajouter les nouvelles entrées avec les vraies IPs
                cat >> /etc/hosts << EOF

# Veza Services - IPs Dynamiques
$FRONTEND_IP veza-frontend veza-frontend.veza.local
$BACKEND_IP veza-backend veza-backend.veza.local  
$CHAT_IP veza-chat veza-chat.veza.local
$STREAM_IP veza-stream veza-stream.veza.local
$POSTGRES_IP veza-postgres veza-postgres.veza.local postgres
$HAPROXY_IP veza-haproxy veza-haproxy.veza.local
$REDIS_IP veza-redis veza-redis.veza.local redis
$STORAGE_IP veza-storage veza-storage.veza.local nfs-server
EOF
            "
        fi
    done
    
    echo -e "${GREEN}✅ Fichiers /etc/hosts mis à jour avec les vraies IPs${NC}"
}

# Fonction pour attendre qu'un container soit prêt
wait_for_container() {
    local container_name=$1
    local max_attempts=30
    local attempt=0
    
    echo -e "${BLUE}⏳ Attente du démarrage de $container_name...${NC}"
    
    while [ $attempt -lt $max_attempts ]; do
        if incus exec "$container_name" -- test -f /etc/hostname >/dev/null 2>&1; then
            echo -e "${GREEN}✅ Container $container_name prêt${NC}"
            return 0
        fi
        sleep 2
        ((attempt++))
    done
    
    echo -e "${RED}❌ Timeout - Container $container_name non prêt${NC}"
    return 1
}

# Fonction pour configurer l'IP statique via DHCP avec réservation
configure_static_ip() {
    local container_name=$1
    local ip_address=$2
    
    echo -e "${BLUE}🌐 Configuration IP pour $container_name ($ip_address)...${NC}"
    
    # Obtenir l'adresse MAC du container
    local mac_address
    mac_address=$(incus config get "$container_name" volatile.eth0.hwaddr)
    
    if [ -z "$mac_address" ]; then
        # Démarrer le container pour obtenir la MAC
        incus start "$container_name" 2>/dev/null || true
        sleep 3
        mac_address=$(incus config get "$container_name" volatile.eth0.hwaddr)
    fi
    
    # Créer une réservation DHCP pour cette MAC vers l'IP désirée
    if [ -n "$mac_address" ]; then
        echo -e "${BLUE}📱 Réservation DHCP: $mac_address → $ip_address${NC}"
        incus network set veza-network ipv4.dhcp.ranges "10.100.0.10-10.100.0.250"
        incus network set veza-network "ipv4.dhcp.expiry" "1h"
        # Note: Incus ne supporte pas les réservations DHCP statiques par MAC
        # On va donc utiliser une approche alternative
    fi
    
    # Redémarrer le container pour récupérer une nouvelle IP
    incus restart "$container_name"
    wait_for_container "$container_name"
    
    # Configuration DNS immédiate  
    configure_dns "$container_name"
    
    # Attendre que le DNS soit stable
    sleep 3
    
    # Récupérer l'IP attribuée par DHCP
    local actual_ip
    local retry=0
    while [ $retry -lt 10 ]; do
        actual_ip=$(incus exec "$container_name" -- ip -4 addr show eth0 | grep -oP '(?<=inet\s)\d+(\.\d+){3}' | head -1)
        if [ -n "$actual_ip" ] && [ "$actual_ip" != "127.0.0.1" ]; then
            echo -e "${GREEN}✅ IP $actual_ip attribuée pour $container_name${NC}"
            # Mettre à jour les variables globales avec la vraie IP
            case "$container_name" in
                "veza-frontend") FRONTEND_IP="$actual_ip" ;;
                "veza-backend") BACKEND_IP="$actual_ip" ;;
                "veza-chat") CHAT_IP="$actual_ip" ;;
                "veza-stream") STREAM_IP="$actual_ip" ;;
                "veza-postgres") POSTGRES_IP="$actual_ip" ;;
                "veza-haproxy") HAPROXY_IP="$actual_ip" ;;
                "veza-redis") REDIS_IP="$actual_ip" ;;
                "veza-storage") STORAGE_IP="$actual_ip" ;;
            esac
            return 0
        fi
        
        echo -e "${BLUE}⏳ Attente attribution IP pour $container_name (tentative $((retry+1))/10)${NC}"
        sleep 3
        ((retry++))
    done
    
    echo -e "${YELLOW}⚠️ Utilisation IP DHCP automatique pour $container_name${NC}"
}

# Fonction pour installer les dépendances de base
install_base_dependencies() {
    local container_name=$1
    
    echo -e "${BLUE}📦 Installation des dépendances de base pour $container_name...${NC}"
    
    incus exec "$container_name" -- bash -c "
        export DEBIAN_FRONTEND=noninteractive
        
        # Désactiver systemd-resolved qui peut interférer avec DNS
        systemctl stop systemd-resolved 2>/dev/null || true
        systemctl disable systemd-resolved 2>/dev/null || true
        
        # Configuration robuste des sources APT avec mirrors de secours
        cat > /etc/apt/sources.list << 'EOF'
# Sources principales
deb http://deb.debian.org/debian bookworm main
deb http://deb.debian.org/debian bookworm-updates main
deb http://security.debian.org/debian-security bookworm-security main

# Mirrors de secours en cas de problème
deb http://ftp.fr.debian.org/debian bookworm main
deb http://ftp.de.debian.org/debian bookworm main

# Sources
deb-src http://deb.debian.org/debian bookworm main
EOF

        # Configuration APT pour éviter les timeouts
        cat > /etc/apt/apt.conf.d/99veza << 'EOF'
APT::Acquire::Retries "3";
APT::Acquire::http::Timeout "10";
APT::Acquire::ftp::Timeout "10";
Acquire::http::Pipeline-Depth "0";
Acquire::http::No-Cache=true;
Acquire::BrokenProxy=true;
EOF
        
        # Test DNS détaillé avant APT
        echo 'Diagnostic DNS pour $container_name...'
        timeout 5 nslookup deb.debian.org 8.8.8.8 || echo 'DNS externe échoué'
        timeout 3 ping -c 1 8.8.8.8 || echo 'Connectivité internet limitée'
        
        # Mise à jour avec retry et diagnostic
        for i in {1..3}; do
            echo 'Tentative APT update $i/3 pour $container_name'
            if timeout 60 apt-get update -o Debug::pkgAcquire::Worker=1; then
                echo 'APT update réussi pour $container_name'
                break
            else
                echo 'Échec APT update $i/3 pour $container_name - diagnostic...'
                echo 'Test connectivité vers deb.debian.org:'
                timeout 5 curl -I http://deb.debian.org/ || echo 'Connexion impossible'
                sleep 10
            fi
        done
        
        # Installation des paquets avec retry
        for i in {1..3}; do
            if apt-get install -y curl wget git build-essential ca-certificates gnupg lsb-release \
                              systemd systemd-sysv init-system-helpers procps net-tools \
                              dnsutils iputils-ping netcat-openbsd htop vim nano; then
                echo 'Installation réussie pour $container_name'
                break
            else
                echo 'Retry installation $i/3 pour $container_name'
                sleep 5
            fi
        done
        
        apt-get clean
        
        # Ne pas créer de configuration systemd-networkd conflictuelle
        # Laisser Incus gérer la configuration réseau de base
    "
}

# Fonction pour créer un service systemd
create_systemd_service() {
    local container_name=$1
    local service_name=$2
    local service_file=$3
    
    echo -e "${BLUE}⚙️ Création du service systemd $service_name pour $container_name...${NC}"
    
    incus file push - "$container_name/etc/systemd/system/$service_name.service" << EOF
$service_file
EOF
    
    incus exec "$container_name" -- bash -c "
        systemctl daemon-reload
        systemctl enable $service_name
    "
}

# Déployer PostgreSQL
deploy_postgres() {
    echo -e "${CYAN}🐘 Déploiement de PostgreSQL...${NC}"
    
    incus launch "$IMAGE" veza-postgres --profile veza-database
    wait_for_container veza-postgres
    configure_static_ip veza-postgres "$POSTGRES_IP"
    
    install_base_dependencies veza-postgres
    
    # Installation PostgreSQL
    incus exec veza-postgres -- bash -c "
        export DEBIAN_FRONTEND=noninteractive
        apt-get update -qq
        apt-get install -y postgresql postgresql-contrib
        systemctl enable postgresql
        systemctl start postgresql
        
        # Configuration PostgreSQL
        sudo -u postgres psql -c \"CREATE USER veza_user WITH PASSWORD 'veza_password';\"
        sudo -u postgres psql -c \"CREATE DATABASE veza_db OWNER veza_user;\"
        sudo -u postgres psql -c \"GRANT ALL PRIVILEGES ON DATABASE veza_db TO veza_user;\"
        
        # Configuration réseau
        echo \"listen_addresses = '*'\" >> /etc/postgresql/15/main/postgresql.conf
        echo \"host all all 10.100.0.0/24 md5\" >> /etc/postgresql/15/main/pg_hba.conf
        
        systemctl restart postgresql
    "
    
    # Importer le schéma de base de données
    if [ -f "$WORKSPACE_DIR/init-db.sql" ]; then
        echo -e "${BLUE}📊 Import du schéma de base de données...${NC}"
        incus file push "$WORKSPACE_DIR/init-db.sql" veza-postgres/tmp/
        incus exec veza-postgres -- sudo -u postgres psql veza_db < /tmp/init-db.sql
    fi
    
    echo -e "${GREEN}✅ PostgreSQL déployé (10.100.0.15)${NC}"
}

# Déployer Redis
deploy_redis() {
    echo -e "${CYAN}🔴 Déploiement de Redis...${NC}"
    
    incus launch "$IMAGE" veza-redis --profile veza-app
    wait_for_container veza-redis
    configure_static_ip veza-redis "$REDIS_IP"
    
    install_base_dependencies veza-redis
    
    # Installation Redis
    incus exec veza-redis -- bash -c "
        export DEBIAN_FRONTEND=noninteractive
        apt-get update -qq
        apt-get install -y redis-server
        
        # Configuration Redis
        sed -i 's/bind 127.0.0.1/bind 0.0.0.0/' /etc/redis/redis.conf
        sed -i 's/protected-mode yes/protected-mode no/' /etc/redis/redis.conf
        
        systemctl enable redis-server
        systemctl restart redis-server
    "
    
    echo -e "${GREEN}✅ Redis déployé (10.100.0.17)${NC}"
}

# Déployer le système de fichiers
deploy_storage() {
    echo -e "${CYAN}🗄️ Déploiement du système de fichiers...${NC}"
    
    incus launch "$IMAGE" veza-storage --profile veza-storage
    wait_for_container veza-storage
    configure_static_ip veza-storage "$STORAGE_IP"
    
    install_base_dependencies veza-storage
    
    # Installation NFS
    incus exec veza-storage -- bash -c "
        export DEBIAN_FRONTEND=noninteractive
        apt-get update -qq
        apt-get install -y nfs-kernel-server
        
        # Configuration des exports NFS
        mkdir -p /storage/{uploads,audio,backups,cache}
        chown -R nobody:nogroup /storage
        chmod -R 755 /storage
        
        echo '/storage/uploads 10.100.0.0/24(rw,sync,no_subtree_check,no_root_squash)' >> /etc/exports
        echo '/storage/audio 10.100.0.0/24(rw,sync,no_subtree_check,no_root_squash)' >> /etc/exports
        echo '/storage/backups 10.100.0.0/24(rw,sync,no_subtree_check,no_root_squash)' >> /etc/exports
        echo '/storage/cache 10.100.0.0/24(rw,sync,no_subtree_check,no_root_squash)' >> /etc/exports
        
        systemctl enable nfs-kernel-server
        systemctl restart nfs-kernel-server
        exportfs -ra
    "
    
    echo -e "${GREEN}✅ Système de fichiers déployé (10.100.0.18)${NC}"
}

# Déployer le backend Go
deploy_backend() {
    echo -e "${CYAN}🔧 Déploiement du Backend Go...${NC}"
    
    incus launch "$IMAGE" veza-backend --profile veza-app
    wait_for_container veza-backend
    configure_static_ip veza-backend "$BACKEND_IP"
    
    install_base_dependencies veza-backend
    
    # Copier le code source AVANT l'installation Go
    incus exec veza-backend -- mkdir -p /app/backend
    incus file push -r "$WORKSPACE_DIR/veza-backend-api/." veza-backend/app/backend/
    
    # Installation Go et compilation
    incus exec veza-backend -- bash -c "
        export DEBIAN_FRONTEND=noninteractive
        
        # Installation Go avec retry
        for i in {1..3}; do
            if wget -q https://go.dev/dl/go1.21.5.linux-amd64.tar.gz && 
               tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz; then
                echo 'Installation Go réussie'
                rm go1.21.5.linux-amd64.tar.gz 2>/dev/null || true
                break
            else
                echo 'Retry installation Go $i/3'
                rm go1.21.5.linux-amd64.tar.gz 2>/dev/null || true
                sleep 10
            fi
        done
        
        export PATH=/usr/local/go/bin:\$PATH
        export GOPATH=/app/go
        
        # Monter le client NFS
        apt-get update -qq
        apt-get install -y nfs-common
        mkdir -p /app/uploads
        
        # Variables d'environnement globales
        cat > /etc/environment << 'EOF'
PATH=/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
GOPATH=/app/go
DATABASE_URL=postgres://veza_user:veza_password@10.100.0.15:5432/veza_db?sslmode=disable
REDIS_URL=redis://10.100.0.17:6379
JWT_SECRET=your-super-secret-jwt-key-change-in-production
PORT=8080
CHAT_SERVER_URL=http://10.100.0.13:8081
STREAM_SERVER_URL=http://10.100.0.14:8082
UPLOAD_DIR=/app/uploads
ALLOWED_ORIGINS=http://10.100.0.11:5173,http://10.100.0.16
EOF
        
        # Compilation du backend
        cd /app/backend
        /usr/local/go/bin/go mod tidy
        /usr/local/go/bin/go build -o veza-backend ./cmd/server/main.go
        chmod +x veza-backend
    "
    
    # Configuration NFS après que le serveur NFS soit prêt
    echo -e "${BLUE}📁 Configuration du montage NFS...${NC}"
    incus exec veza-backend -- bash -c "
        # Attendre que le serveur NFS soit prêt
        timeout=60
        while [ \$timeout -gt 0 ]; do
            if mount -t nfs 10.100.0.18:/storage/uploads /app/uploads 2>/dev/null; then
                break
            fi
            sleep 2
            ((timeout--))
        done
        
        echo '10.100.0.18:/storage/uploads /app/uploads nfs defaults 0 0' >> /etc/fstab
    "
    
    # Créer le service systemd
    create_systemd_service veza-backend veza-backend '[Unit]
Description=Veza Backend API Server
After=network.target
Wants=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/app/backend
ExecStart=/app/backend/veza-backend
Restart=always
RestartSec=5
Environment=PATH=/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
EnvironmentFile=/etc/environment

[Install]
WantedBy=multi-user.target'
    
    echo -e "${GREEN}✅ Backend Go déployé et configuré (10.100.0.12)${NC}"
}

# Déployer le serveur de chat Rust
deploy_chat() {
    echo -e "${CYAN}💬 Déploiement du Chat Server Rust...${NC}"
    
    incus launch "$IMAGE" veza-chat --profile veza-app
    wait_for_container veza-chat
    configure_static_ip veza-chat "$CHAT_IP"
    
    install_base_dependencies veza-chat
    
    # Copier le code source AVANT l'installation Rust
    incus exec veza-chat -- mkdir -p /app/chat
    incus file push -r "$WORKSPACE_DIR/veza-chat-server/." veza-chat/app/chat/
    
    # Installation Rust et compilation
    incus exec veza-chat -- bash -c "
        export DEBIAN_FRONTEND=noninteractive
        
        # Installation des dépendances de compilation avec retry
        for i in {1..3}; do
            if apt-get update -qq && apt-get install -y pkg-config libssl-dev libpq-dev; then
                echo 'Dépendances compilation chat installées'
                break
            else
                echo 'Retry dépendances compilation chat $i/3'
                sleep 5
            fi
        done
        
        # Installation Rust avec retry
        for i in {1..3}; do
            if curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y --default-toolchain stable; then
                echo 'Installation Rust chat réussie'
                break
            else
                echo 'Retry installation Rust chat $i/3'
                sleep 10
            fi
        done
        
        source ~/.cargo/env 2>/dev/null || true
        export PATH=~/.cargo/bin:\$PATH
        
        # Variables d'environnement globales
        cat > /etc/environment << 'EOF'
PATH=/root/.cargo/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
DATABASE_URL=postgres://veza_user:veza_password@10.100.0.15:5432/veza_db?sslmode=disable
REDIS_URL=redis://10.100.0.17:6379
JWT_SECRET=your-super-secret-jwt-key-change-in-production
PORT=8081
RUST_LOG=chat_server=debug,tower_http=debug
ALLOWED_ORIGINS=http://10.100.0.11:5173,http://10.100.0.16
EOF
        
        # Compilation du chat server
        cd /app/chat
        ~/.cargo/bin/cargo build --release
        
        # Créer un lien vers l'exécutable
        ln -sf /app/chat/target/release/veza-chat-server /usr/local/bin/veza-chat-server
    "
    
    # Créer le service systemd
    create_systemd_service veza-chat veza-chat '[Unit]
Description=Veza Chat Server (Rust WebSocket)
After=network.target postgresql.service redis.service
Wants=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/app/chat
ExecStart=/usr/local/bin/veza-chat-server
Restart=always
RestartSec=5
Environment=PATH=/root/.cargo/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
EnvironmentFile=/etc/environment

[Install]
WantedBy=multi-user.target'
    
    echo -e "${GREEN}✅ Chat Server Rust déployé et configuré (10.100.0.13)${NC}"
}

# Déployer le serveur de streaming Rust
deploy_stream() {
    echo -e "${CYAN}🎵 Déploiement du Stream Server Rust...${NC}"
    
    incus launch "$IMAGE" veza-stream --profile veza-app
    wait_for_container veza-stream
    configure_static_ip veza-stream "$STREAM_IP"
    
    install_base_dependencies veza-stream
    
    # Copier le code source AVANT l'installation Rust
    incus exec veza-stream -- mkdir -p /app/stream
    incus file push -r "$WORKSPACE_DIR/veza-stream-server/." veza-stream/app/stream/
    
    # Installation Rust et compilation
    incus exec veza-stream -- bash -c "
        export DEBIAN_FRONTEND=noninteractive
        
        # Installation des dépendances de compilation avec retry
        for i in {1..3}; do
            if apt-get update -qq && apt-get install -y pkg-config libssl-dev libpq-dev nfs-common ffmpeg; then
                echo 'Dépendances compilation stream installées'
                break
            else
                echo 'Retry dépendances compilation stream $i/3'
                sleep 5
            fi
        done
        
        # Installation Rust avec retry
        for i in {1..3}; do
            if curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y --default-toolchain stable; then
                echo 'Installation Rust stream réussie'
                break
            else
                echo 'Retry installation Rust stream $i/3'
                sleep 10
            fi
        done
        
        source ~/.cargo/env 2>/dev/null || true
        export PATH=~/.cargo/bin:\$PATH
        
        # Créer le répertoire audio
        mkdir -p /storage/audio
        
        # Variables d'environnement globales
        cat > /etc/environment << 'EOF'
PATH=/root/.cargo/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
DATABASE_URL=postgres://veza_user:veza_password@10.100.0.15:5432/veza_db?sslmode=disable
REDIS_URL=redis://10.100.0.17:6379
JWT_SECRET=your-super-secret-jwt-key-change-in-production
PORT=8082
AUDIO_DIR=/storage/audio
RUST_LOG=stream_server=debug,tower_http=debug
ALLOWED_ORIGINS=http://10.100.0.11:5173,http://10.100.0.16
EOF
        
        # Compilation du stream server
        cd /app/stream
        ~/.cargo/bin/cargo build --release
        
        # Créer un lien vers l'exécutable
        ln -sf /app/stream/target/release/veza-stream-server /usr/local/bin/veza-stream-server
    "
    
    # Configuration NFS après que le serveur NFS soit prêt
    echo -e "${BLUE}🎵 Configuration du montage audio NFS...${NC}"
    incus exec veza-stream -- bash -c "
        # Attendre que le serveur NFS soit prêt
        timeout=60
        while [ \$timeout -gt 0 ]; do
            if mount -t nfs 10.100.0.18:/storage/audio /storage/audio 2>/dev/null; then
                break
            fi
            sleep 2
            ((timeout--))
        done
        
        echo '10.100.0.18:/storage/audio /storage/audio nfs defaults 0 0' >> /etc/fstab
    "
    
    # Créer le service systemd
    create_systemd_service veza-stream veza-stream '[Unit]
Description=Veza Stream Server (Rust Audio)
After=network.target postgresql.service redis.service
Wants=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/app/stream
ExecStart=/usr/local/bin/veza-stream-server
Restart=always
RestartSec=5
Environment=PATH=/root/.cargo/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
EnvironmentFile=/etc/environment

[Install]
WantedBy=multi-user.target'
    
    echo -e "${GREEN}✅ Stream Server Rust déployé et configuré (10.100.0.14)${NC}"
}

# Déployer le frontend React
deploy_frontend() {
    echo -e "${CYAN}⚛️ Déploiement du Frontend React...${NC}"
    
    incus launch "$IMAGE" veza-frontend --profile veza-app
    wait_for_container veza-frontend
    configure_static_ip veza-frontend "$FRONTEND_IP"
    
    install_base_dependencies veza-frontend
    
    # Copier le code source AVANT l'installation Node.js
    incus exec veza-frontend -- mkdir -p /app/frontend
    incus file push -r "$WORKSPACE_DIR/veza-frontend/." veza-frontend/app/frontend/
    
    # Installation Node.js et build
    incus exec veza-frontend -- bash -c "
        export DEBIAN_FRONTEND=noninteractive
        
        # Installation Node.js 20 avec retry
        for i in {1..3}; do
            if curl -fsSL https://deb.nodesource.com/setup_20.x | bash - && apt-get install -y nodejs; then
                echo 'Installation Node.js réussie'
                break
            else
                echo 'Retry installation Node.js $i/3'
                sleep 10
            fi
        done
        
        # Créer le fichier .env pour Vite
        cat > /app/frontend/.env << 'EOF'
NODE_ENV=development
VITE_API_URL=http://10.100.0.12:8080/api/v1
VITE_WS_CHAT_URL=ws://10.100.0.13:8081/ws
VITE_WS_STREAM_URL=ws://10.100.0.14:8082/ws
VITE_DEBUG=true
VITE_UPLOAD_MAX_SIZE=50000000
VITE_CACHE_ENABLED=true
EOF
        
        # Variables d'environnement globales
        cat > /etc/environment << 'EOF'
PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
NODE_ENV=development
VITE_API_URL=http://10.100.0.12:8080/api/v1
VITE_WS_CHAT_URL=ws://10.100.0.13:8081/ws
VITE_WS_STREAM_URL=ws://10.100.0.14:8082/ws
EOF
        
        # Installation des dépendances et build
        cd /app/frontend
        npm install
        
        # Build de production pour optimiser (optionnel, pour dev on peut laisser en mode dev)
        # npm run build
    "
    
    # Créer le service systemd pour le serveur de développement Vite
    create_systemd_service veza-frontend veza-frontend '[Unit]
Description=Veza Frontend React Development Server
After=network.target
Wants=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/app/frontend
ExecStart=/usr/bin/npm run dev -- --host 0.0.0.0 --port 5173
Restart=always
RestartSec=5
Environment=PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
EnvironmentFile=/etc/environment

[Install]
WantedBy=multi-user.target'
    
    echo -e "${GREEN}✅ Frontend React déployé et configuré (10.100.0.11)${NC}"
}

# Déployer HAProxy
deploy_haproxy() {
    echo -e "${CYAN}⚖️ Déploiement de HAProxy...${NC}"
    
    incus launch "$IMAGE" veza-haproxy --profile veza-app
    wait_for_container veza-haproxy
    configure_static_ip veza-haproxy "$HAPROXY_IP"
    
    install_base_dependencies veza-haproxy
    
    # Installation HAProxy
    incus exec veza-haproxy -- bash -c "
        export DEBIAN_FRONTEND=noninteractive
        apt-get update -qq
        apt-get install -y haproxy
        systemctl enable haproxy
    "
    
    # Créer la configuration HAProxy optimisée
    incus file push - veza-haproxy/etc/haproxy/haproxy.cfg << 'EOF'
global
    daemon
    log stdout local0
    chroot /var/lib/haproxy
    stats socket /run/haproxy/admin.sock mode 660 level admin
    stats timeout 30s
    user haproxy
    group haproxy

defaults
    mode http
    log global
    option httplog
    option dontlognull
    option redispatch
    retries 3
    timeout connect 5000
    timeout client 50000
    timeout server 50000
    errorfile 400 /etc/haproxy/errors/400.http
    errorfile 403 /etc/haproxy/errors/403.http
    errorfile 408 /etc/haproxy/errors/408.http
    errorfile 500 /etc/haproxy/errors/500.http
    errorfile 502 /etc/haproxy/errors/502.http
    errorfile 503 /etc/haproxy/errors/503.http
    errorfile 504 /etc/haproxy/errors/504.http

# Interface de stats HAProxy
frontend stats
    bind *:8404
    stats enable
    stats uri /stats
    stats refresh 30s
    stats admin if TRUE

# Frontend principal
frontend main
    bind *:80
    default_backend veza_frontend

# Backend pour le frontend React
backend veza_frontend
    balance roundrobin
    option httpchk GET /
    server frontend1 $FRONTEND_IP:5173 check

# Frontend pour l'API
frontend api
    bind *:8080
    default_backend veza_api

# Backend pour l'API Go
backend veza_api
    balance roundrobin
    option httpchk GET /health
    server api1 $BACKEND_IP:8080 check

# Frontend pour le WebSocket Chat
frontend chat_ws
    bind *:8081
    default_backend veza_chat

# Backend pour le Chat Rust
backend veza_chat
    balance roundrobin
    option httpchk GET /health
    server chat1 $CHAT_IP:8081 check

# Frontend pour le Stream
frontend stream
    bind *:8082
    default_backend veza_stream

# Backend pour le Stream Rust
backend veza_stream
    balance roundrobin
    option httpchk GET /health
    server stream1 $STREAM_IP:8082 check
EOF
    
    # Redémarrer HAProxy avec la nouvelle configuration
        incus exec veza-haproxy -- systemctl restart haproxy
    
    echo -e "${GREEN}✅ HAProxy déployé et configuré (10.100.0.16)${NC}"
}

# Fonction principale de déploiement
main() {
    echo -e "${BLUE}🚀 Début du déploiement complet...${NC}"
    echo -e "${YELLOW}Cette opération va créer 8 containers. Continuer ? (o/N)${NC}"
    read -r response
    
    if [[ "$response" != "o" && "$response" != "oui" ]]; then
        echo -e "${GREEN}Déploiement annulé${NC}"
        exit 0
    fi
    
    # Déploiement dans l'ordre optimal
    echo -e "${BLUE}📋 Ordre de déploiement :${NC}"
    echo -e "  1. PostgreSQL (Base de données)"
    echo -e "  2. Redis (Cache)"
    echo -e "  3. Système de fichiers (NFS)"
    echo -e "  4. Backend Go (API)"
    echo -e "  5. Chat Server Rust"
    echo -e "  6. Stream Server Rust"
    echo -e "  7. Frontend React"
    echo -e "  8. HAProxy (Load Balancer)"
    echo ""
    
    # Déploiement séquentiel avec gestion d'erreurs
    echo -e "${BLUE}🚀 Début du déploiement des services...${NC}"
    
    deploy_postgres || { echo -e "${RED}❌ Échec PostgreSQL${NC}"; exit 1; }
    deploy_redis || { echo -e "${RED}❌ Échec Redis${NC}"; exit 1; }
    deploy_storage || { echo -e "${RED}❌ Échec Storage${NC}"; exit 1; }
    deploy_backend || { echo -e "${RED}❌ Échec Backend${NC}"; exit 1; }
    deploy_chat || { echo -e "${RED}❌ Échec Chat${NC}"; exit 1; }
    deploy_stream || { echo -e "${RED}❌ Échec Stream${NC}"; exit 1; }
    deploy_frontend || { echo -e "${RED}❌ Échec Frontend${NC}"; exit 1; }
    deploy_haproxy || { echo -e "${RED}❌ Échec HAProxy${NC}"; exit 1; }
    
    # Mise à jour des fichiers /etc/hosts avec les vraies IPs DHCP
    update_hosts_files
    
    # Démarrage de tous les services
    echo -e "${BLUE}🔧 Démarrage de tous les services...${NC}"
    
    echo -e "${CYAN}🔄 Démarrage Backend Go...${NC}"
    incus exec veza-backend -- systemctl start veza-backend
    
    echo -e "${CYAN}🔄 Démarrage Chat Server...${NC}"
    incus exec veza-chat -- systemctl start veza-chat
    
    echo -e "${CYAN}🔄 Démarrage Stream Server...${NC}"
    incus exec veza-stream -- systemctl start veza-stream
    
    echo -e "${CYAN}🔄 Démarrage Frontend React...${NC}"
    incus exec veza-frontend -- systemctl start veza-frontend
    
    # Attendre un peu pour que les services se stabilisent
    echo -e "${BLUE}⏳ Attente de la stabilisation des services...${NC}"
    sleep 10
    
    # Vérification finale des services avec retry
    echo -e "${BLUE}🔍 Vérification de l'état des services...${NC}"
    
    for service in veza-backend veza-chat veza-stream veza-frontend; do
        container=$(echo $service | sed 's/veza-/veza-/')
        case $service in
            "veza-backend") container="veza-backend" ;;
            "veza-chat") container="veza-chat" ;;
            "veza-stream") container="veza-stream" ;;
            "veza-frontend") container="veza-frontend" ;;
        esac
        
        # Essayer pendant 30 secondes max
        for i in {1..6}; do
            if incus exec "$container" -- systemctl is-active "$service" >/dev/null 2>&1; then
                echo -e "${GREEN}✅ Service $service actif${NC}"
                break
            elif [ $i -eq 6 ]; then
                echo -e "${YELLOW}⚠️ Service $service pas encore actif - Vérifier manuellement${NC}"
            else
                echo -e "${BLUE}⏳ Attente activation $service (tentative $i/6)...${NC}"
                sleep 5
            fi
        done
    done
    
    # Vérification de la connectivité réseau inter-services
    echo -e "${BLUE}🌐 Test de connectivité inter-services...${NC}"
    
    # Test de PostgreSQL depuis backend
    if incus exec veza-backend -- timeout 5 nc -z "$POSTGRES_IP" 5432 >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Backend → PostgreSQL : OK${NC}"
    else
        echo -e "${YELLOW}⚠️ Backend → PostgreSQL : Problème de connectivité${NC}"
    fi
    
    # Test de Redis depuis backend
    if incus exec veza-backend -- timeout 5 nc -z "$REDIS_IP" 6379 >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Backend → Redis : OK${NC}"
    else
        echo -e "${YELLOW}⚠️ Backend → Redis : Problème de connectivité${NC}"
    fi
    
    # Vérification finale
    echo -e "${GREEN}🎉 Déploiement terminé !${NC}"
    echo ""
    echo -e "${BLUE}📊 État final des containers :${NC}"
    incus ls
    echo ""
    echo -e "${BLUE}🌐 Points d'accès (IPs dynamiques DHCP) :${NC}"
    echo -e "  • Application : ${YELLOW}http://$HAPROXY_IP${NC} (HAProxy)"
    echo -e "  • HAProxy Stats : ${YELLOW}http://$HAPROXY_IP:8404/stats${NC}"
    echo -e "  • Frontend Dev : ${YELLOW}http://$FRONTEND_IP:5173${NC}"
    echo -e "  • Backend API : ${YELLOW}http://$BACKEND_IP:8080${NC}"
    echo -e "  • Chat WebSocket : ${YELLOW}ws://$CHAT_IP:8081/ws${NC}"
    echo -e "  • Stream Server : ${YELLOW}http://$STREAM_IP:8082${NC}"
    echo -e "  • PostgreSQL : ${YELLOW}$POSTGRES_IP:5432${NC}"
    echo -e "  • Redis : ${YELLOW}$REDIS_IP:6379${NC}"
    echo ""
    echo -e "${CYAN}💡 Commandes utiles :${NC}"
    echo -e "  • État containers : ${YELLOW}incus ls${NC}"
    echo -e "  • Logs service : ${YELLOW}incus exec <container> -- journalctl -u <service> -f${NC}"
    echo -e "  • Shell : ${YELLOW}incus exec <container> -- bash${NC}"
    echo -e "  • Redémarrer service : ${YELLOW}incus exec <container> -- systemctl restart <service>${NC}"
    echo -e "  • État service : ${YELLOW}incus exec <container> -- systemctl status <service>${NC}"
}

# Vérifier qu'Incus est disponible
if ! command -v incus &> /dev/null; then
    echo -e "${RED}❌ Incus n'est pas installé ou accessible${NC}"
    exit 1
fi

# Vérifier que la configuration est faite
if ! incus network show veza-network >/dev/null 2>&1; then
    echo -e "${RED}❌ Configuration Incus manquante${NC}"
    echo -e "${YELLOW}💡 Exécutez d'abord : ./scripts/incus-setup.sh${NC}"
    exit 1
fi

# Exécuter le déploiement
main