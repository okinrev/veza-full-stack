#!/bin/bash

# Script de débogage des connexions rsync/SSH

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

echo -e "${BLUE}🔍 Débogage des connexions Veza${NC}"
echo ""

SSH_KEY="$HOME/.ssh/veza_rsa"
SSH_OPTS="-i $SSH_KEY -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o LogLevel=ERROR"

# Tester les connexions SSH
containers=("veza-backend" "veza-chat" "veza-stream" "veza-frontend")

for container in "${containers[@]}"; do
    echo -e "${CYAN}📡 Test connexion $container...${NC}"
    
    # Obtenir l'IP
    container_ip=$(incus ls "$container" -c 4 --format csv | cut -d' ' -f1)
    
    if [ -z "$container_ip" ]; then
        echo -e "${RED}❌ Pas d'IP pour $container${NC}"
        continue
    fi
    
    echo -e "${BLUE}  IP: $container_ip${NC}"
    
    # Test ping
    if ping -c 1 -W 2 "$container_ip" >/dev/null 2>&1; then
        echo -e "${GREEN}  ✅ Ping OK${NC}"
    else
        echo -e "${RED}  ❌ Ping échoué${NC}"
        continue
    fi
    
    # Test SSH
    if ssh $SSH_OPTS "root@$container_ip" "echo 'SSH OK'" >/dev/null 2>&1; then
        echo -e "${GREEN}  ✅ SSH OK${NC}"
    else
        echo -e "${RED}  ❌ SSH échoué${NC}"
        continue
    fi
    
    # Test service SSH
    ssh_status=$(ssh $SSH_OPTS "root@$container_ip" "systemctl is-active ssh" 2>/dev/null || echo "failed")
    echo -e "${BLUE}  Service SSH: $ssh_status${NC}"
    
    echo ""
done

echo -e "${CYAN}🔑 Vérification clé SSH...${NC}"
if [ -f "$SSH_KEY" ]; then
    echo -e "${GREEN}✅ Clé privée trouvée: $SSH_KEY${NC}"
    echo -e "${BLUE}   Permissions: $(ls -l $SSH_KEY | cut -d' ' -f1)${NC}"
else
    echo -e "${RED}❌ Clé privée manquante: $SSH_KEY${NC}"
fi

if [ -f "$SSH_KEY.pub" ]; then
    echo -e "${GREEN}✅ Clé publique trouvée: $SSH_KEY.pub${NC}"
else
    echo -e "${RED}❌ Clé publique manquante: $SSH_KEY.pub${NC}"
fi
