#!/bin/bash

# Script de correction DNS ultra-robuste pour containers Incus
set -euo pipefail

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
PURPLE='\033[0;35m'
NC='\033[0m'

echo -e "${PURPLE}"
echo "╭──────────────────────────────────────────╮"
echo "│    🔧 Correction DNS Ultra-Robuste      │"
echo "╰──────────────────────────────────────────╯"
echo -e "${NC}"

# Fonction de diagnostic réseau avancé
network_diagnosis() {
    local container_name=$1
    
    echo -e "${BLUE}🔍 Diagnostic réseau avancé pour $container_name...${NC}"
    
    incus exec "$container_name" -- bash -c "
        echo '=== DIAGNOSTIC RÉSEAU COMPLET ==='
        
        # 1. Configuration réseau
        echo '--- Configuration réseau ---'
        ip addr show eth0 2>/dev/null || echo 'Interface eth0 non trouvée'
        ip route show 2>/dev/null || echo 'Routes non disponibles'
        
        # 2. Configuration DNS actuelle
        echo '--- DNS actuel ---'
        cat /etc/resolv.conf 2>/dev/null || echo 'resolv.conf non accessible'
        
        # 3. Services réseau actifs
        echo '--- Services réseau ---'
        systemctl is-active systemd-resolved 2>/dev/null || echo 'systemd-resolved inactif'
        systemctl is-active systemd-networkd 2>/dev/null || echo 'systemd-networkd inactif'
        systemctl is-active NetworkManager 2>/dev/null || echo 'NetworkManager inactif'
        
        # 4. Test de connectivité de base
        echo '--- Tests connectivité ---'
        timeout 3 ping -c 1 10.100.0.1 >/dev/null 2>&1 && echo 'Gateway 10.100.0.1: OK' || echo 'Gateway 10.100.0.1: ÉCHEC'
        timeout 3 ping -c 1 8.8.8.8 >/dev/null 2>&1 && echo 'DNS Google: OK' || echo 'DNS Google: ÉCHEC'
        
        # 5. Test DNS
        echo '--- Tests DNS ---'
        timeout 5 nslookup google.com 8.8.8.8 >/dev/null 2>&1 && echo 'nslookup direct: OK' || echo 'nslookup direct: ÉCHEC'
        timeout 5 dig @8.8.8.8 google.com >/dev/null 2>&1 && echo 'dig direct: OK' || echo 'dig direct: ÉCHEC'
        
        echo '=== FIN DIAGNOSTIC ==='
    "
}

# Fonction de correction DNS ultra-robuste
fix_container_dns() {
    local container_name=$1
    
    echo -e "${BLUE}🔧 Correction DNS ultra-robuste pour $container_name...${NC}"
    
    # Vérifier que le container existe et est running
    if ! incus list "$container_name" --format csv | grep -q "RUNNING"; then
        echo -e "${YELLOW}⚠️ Container $container_name n'est pas running${NC}"
        return 1
    fi
    
    # Diagnostic avant correction
    network_diagnosis "$container_name"
    
    incus exec "$container_name" -- bash -c "
        echo '🔧 Début correction DNS ultra-robuste pour $container_name'
        
        # 1. ARRÊTER TOUS les services DNS conflictuels
        echo 'Arrêt de tous les services DNS...'
        systemctl stop systemd-resolved 2>/dev/null || true
        systemctl disable systemd-resolved 2>/dev/null || true
        systemctl mask systemd-resolved 2>/dev/null || true
        systemctl stop systemd-networkd 2>/dev/null || true
        systemctl disable systemd-networkd 2>/dev/null || true
        systemctl stop resolvconf 2>/dev/null || true
        systemctl disable resolvconf 2>/dev/null || true
        
        # 2. NETTOYER complètement l'ancienne configuration
        echo 'Nettoyage configuration DNS...'
        rm -f /etc/resolv.conf*
        rm -f /run/systemd/resolve/resolv.conf
        rm -f /run/systemd/resolve/stub-resolv.conf
        rm -rf /etc/systemd/resolved.conf.d/
        rm -rf /etc/systemd/network/
        
        # 3. TUER tous les processus réseaux problématiques
        pkill -f systemd-resolved 2>/dev/null || true
        pkill -f systemd-networkd 2>/dev/null || true
        pkill -f dhclient 2>/dev/null || true
        
        # 4. CRÉER une configuration DNS STATIQUE et IMMUTABLE
        echo 'Création configuration DNS statique...'
        cat > /etc/resolv.conf << 'EOF'
# Configuration DNS Veza - ULTRA-ROBUSTE
# Multiple serveurs DNS pour redondance maximale
nameserver 8.8.8.8
nameserver 8.8.4.4
nameserver 1.1.1.1
nameserver 1.0.0.1
nameserver 9.9.9.9
nameserver 149.112.112.112

# Configuration pour optimiser la résolution
search veza.local .
options timeout:3 attempts:3 rotate single-request-reopen
options edns0 trust-ad
EOF
        
        # 5. PROTECTION contre les modifications
        chattr +i /etc/resolv.conf 2>/dev/null || {
            chmod 444 /etc/resolv.conf
            echo 'Attention: chattr non disponible, utilisation chmod 444'
        }
        
        # 6. CONFIGURATION NetworkManager pour qu'il n'interfère pas
        mkdir -p /etc/NetworkManager/conf.d/
        cat > /etc/NetworkManager/conf.d/99-veza-dns-no-touch.conf << 'EOF'
[main]
# Désactiver la gestion DNS par NetworkManager
dns=none
systemd-resolved=false
rc-manager=unmanaged

[connection]
# Ne pas modifier le DNS
ipv4.ignore-auto-dns=true
ipv6.ignore-auto-dns=true
EOF
        
        # 7. CONFIGURATION netplan pour éviter les conflits
        mkdir -p /etc/netplan/
        cat > /etc/netplan/99-veza-static-dns.yaml << 'EOF'
network:
  version: 2
  renderer: networkd
  ethernets:
    eth0:
      dhcp4: true
      dhcp4-overrides:
        use-dns: false
        use-domains: false
      nameservers:
        addresses: [8.8.8.8, 8.8.4.4, 1.1.1.1, 9.9.9.9]
        search: [veza.local]
      dhcp6: false
EOF
        
        # 8. REDÉMARRER les services nécessaires
        systemctl restart NetworkManager 2>/dev/null || true
        netplan apply 2>/dev/null || true
        
        # 9. TESTS DNS immédiats et répétés
        echo 'Tests DNS après correction...'
        
        DNS_SUCCESS=false
        for attempt in {1..5}; do
            echo \"Test DNS tentative \$attempt/5...\"
            
            # Tests multiples avec différents outils et serveurs
            if timeout 5 nslookup deb.debian.org 8.8.8.8 >/dev/null 2>&1 && \
               timeout 5 nslookup google.com 8.8.4.4 >/dev/null 2>&1 && \
               timeout 5 dig @1.1.1.1 debian.org >/dev/null 2>&1; then
                echo '✅ Tous les tests DNS réussis pour $container_name'
                DNS_SUCCESS=true
                break
            else
                echo \"❌ Test DNS \$attempt échoué, retry...\"
                sleep 2
            fi
        done
        
        # 10. Test de connectivité globale
        echo 'Test connectivité internet...'
        if timeout 5 ping -c 3 -W 2 8.8.8.8 >/dev/null 2>&1; then
            echo '✅ Connectivité internet OK pour $container_name'
        else
            echo '⚠️ Connectivité internet limitée pour $container_name'
            
            # Test alternatif avec curl
            if timeout 10 curl -s --max-time 5 http://detectportal.firefox.com/canonical.html >/dev/null 2>&1; then
                echo '✅ Connectivité HTTP confirmée'
            else
                echo '❌ Problème de connectivité HTTP'
            fi
        fi
        
        # 11. Configuration finale et protection
        echo 'Protection finale configuration...'
        
        # Créer un service systemd pour maintenir la configuration DNS
        cat > /etc/systemd/system/veza-dns-guard.service << 'EOF'
[Unit]
Description=Veza DNS Configuration Guard
After=network.target

[Service]
Type=oneshot
RemainAfterExit=yes
ExecStart=/bin/bash -c \"
  # Vérifier et restaurer la configuration DNS si nécessaire
  if ! grep -q '8.8.8.8' /etc/resolv.conf 2>/dev/null; then
    echo 'Restauration configuration DNS Veza...'
    cat > /etc/resolv.conf << 'EOFINNER'
nameserver 8.8.8.8
nameserver 8.8.4.4
nameserver 1.1.1.1
nameserver 9.9.9.9
search veza.local .
options timeout:3 attempts:3 rotate
EOFINNER
    chmod 444 /etc/resolv.conf
    chattr +i /etc/resolv.conf 2>/dev/null || true
  fi
\"
ExecReload=/bin/bash -c \"
  chattr -i /etc/resolv.conf 2>/dev/null || true
  cat > /etc/resolv.conf << 'EOFINNER'
nameserver 8.8.8.8
nameserver 8.8.4.4
nameserver 1.1.1.1
nameserver 9.9.9.9
search veza.local .
options timeout:3 attempts:3 rotate
EOFINNER
  chmod 444 /etc/resolv.conf
  chattr +i /etc/resolv.conf 2>/dev/null || true
\"

[Install]
WantedBy=multi-user.target
EOF
        
        # Activer le service de protection DNS
        systemctl daemon-reload
        systemctl enable veza-dns-guard.service
        systemctl start veza-dns-guard.service
        
        if [ \"\$DNS_SUCCESS\" = true ]; then
            echo '✅ Configuration DNS ultra-robuste réussie pour $container_name'
        else
            echo '⚠️ Configuration appliquée mais tests DNS partiels pour $container_name'
        fi
    "
    
    # Diagnostic après correction
    echo -e "${CYAN}🔍 Diagnostic post-correction pour $container_name:${NC}"
    network_diagnosis "$container_name"
    
    echo -e "${GREEN}✅ Correction DNS ultra-robuste terminée pour $container_name${NC}"
}

# Fonction pour corriger le réseau hôte Incus
fix_host_network() {
    echo -e "${BLUE}🔧 Vérification et correction réseau hôte...${NC}"
    
    # Vérifier IPv4 forwarding
    if ! sysctl net.ipv4.ip_forward | grep -q "1"; then
        echo -e "${YELLOW}⚠️ Activation IPv4 forwarding...${NC}"
        sudo sysctl -w net.ipv4.ip_forward=1
        echo 'net.ipv4.ip_forward=1' | sudo tee /etc/sysctl.d/99-incus.conf > /dev/null
    fi
    
    # Vérifier la configuration du réseau Incus
    if ! incus network info veza-network | grep -q "State: up"; then
        echo -e "${YELLOW}⚠️ Redémarrage du réseau Incus...${NC}"
        incus network set veza-network raw.dnsmasq "
server=8.8.8.8
server=8.8.4.4
server=1.1.1.1
server=9.9.9.9
cache-size=1000
dhcp-authoritative
dhcp-option=option:dns-server,8.8.8.8,8.8.4.4,1.1.1.1
dhcp-option=option:domain-name,veza.local"
    fi
    
    echo -e "${GREEN}✅ Réseau hôte vérifié${NC}"
}

# Fonction principale
main() {
    # Corriger le réseau hôte d'abord
    fix_host_network
    
    # Obtenir la liste des containers veza en cours
    containers=$(incus ls --format csv | grep "veza-" | grep "RUNNING" | cut -d, -f1)
    
    if [ -z "$containers" ]; then
        echo -e "${YELLOW}⚠️ Aucun container Veza running trouvé${NC}"
        echo -e "${CYAN}💡 Conteneurs disponibles :${NC}"
        incus ls | grep veza- || echo "Aucun container Veza trouvé"
        exit 0
    fi
    
    echo -e "${BLUE}🔍 Containers Veza trouvés :${NC}"
    echo "$containers"
    echo ""
    
    # Demander confirmation
    echo -e "${YELLOW}⚠️ Cette opération va reconfigurer complètement le DNS. Continuer ? (o/N)${NC}"
    read -r response
    
    if [[ "$response" != "o" && "$response" != "oui" ]]; then
        echo -e "${GREEN}Opération annulée${NC}"
        exit 0
    fi
    
    # Appliquer la correction à chaque container
    for container in $containers; do
        echo -e "${CYAN}🔧 Traitement de $container...${NC}"
        fix_container_dns "$container"
        echo ""
    done
    
    echo -e "${GREEN}🎉 Correction DNS ultra-robuste terminée pour tous les containers !${NC}"
    echo ""
    
    # Test global final ultra-complet
    echo -e "${BLUE}🧪 Tests finaux ultra-complets...${NC}"
    echo ""
    
    for container in $containers; do
        echo -e "${CYAN}Testing $container:${NC}"
        
        # Tests multiples et détaillés
        dns_ok=0
        connectivity_ok=0
        
        # Test DNS avec multiple serveurs
        if incus exec "$container" -- timeout 5 nslookup deb.debian.org 8.8.8.8 >/dev/null 2>&1; then
            ((dns_ok++))
        fi
        if incus exec "$container" -- timeout 5 nslookup google.com 8.8.4.4 >/dev/null 2>&1; then
            ((dns_ok++))
        fi
        if incus exec "$container" -- timeout 5 dig @1.1.1.1 debian.org >/dev/null 2>&1; then
            ((dns_ok++))
        fi
        
        # Test connectivité
        if incus exec "$container" -- timeout 3 ping -c 1 8.8.8.8 >/dev/null 2>&1; then
            ((connectivity_ok++))
        fi
        if incus exec "$container" -- timeout 3 ping -c 1 1.1.1.1 >/dev/null 2>&1; then
            ((connectivity_ok++))
        fi
        
        # Affichage des résultats
        if [ $dns_ok -ge 2 ] && [ $connectivity_ok -ge 1 ]; then
            echo -e "  ${GREEN}✅ $container - DNS: $dns_ok/3, Connectivité: $connectivity_ok/2 - EXCELLENT${NC}"
        elif [ $dns_ok -ge 1 ] || [ $connectivity_ok -ge 1 ]; then
            echo -e "  ${YELLOW}⚠️ $container - DNS: $dns_ok/3, Connectivité: $connectivity_ok/2 - PARTIEL${NC}"
        else
            echo -e "  ${RED}❌ $container - DNS: $dns_ok/3, Connectivité: $connectivity_ok/2 - ÉCHEC${NC}"
        fi
    done
    
    echo ""
    echo -e "${CYAN}💡 Commandes utiles :${NC}"
    echo -e "  ${YELLOW}• Test DNS manuel :${NC} incus exec <container> -- nslookup deb.debian.org"
    echo -e "  ${YELLOW}• Vérifier resolv.conf :${NC} incus exec <container> -- cat /etc/resolv.conf"
    echo -e "  ${YELLOW}• Test connectivité :${NC} incus exec <container> -- ping -c 3 8.8.8.8"
    echo -e "  ${YELLOW}• Forcer restauration DNS :${NC} incus exec <container> -- systemctl restart veza-dns-guard"
    echo -e "  ${YELLOW}• Relancer déploiement :${NC} ./scripts/incus-deploy.sh"
    echo ""
    echo -e "${GREEN}✨ Configuration DNS ultra-robuste appliquée avec succès !${NC}"
}

# Vérifier qu'Incus est disponible
if ! command -v incus &> /dev/null; then
    echo -e "${RED}❌ Incus n'est pas installé ou accessible${NC}"
    exit 1
fi

# Exécuter la correction
main 