#!/bin/bash

# Script de configuration rsync pour synchronisation automatique
# Synchronise le code local vers les containers et redémarre les services

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
echo "│       🔄 Configuration Rsync et Sync      │"
echo "╰──────────────────────────────────────────╯"
echo -e "${NC}"

WORKSPACE_DIR="$(pwd)"

# Configuration des clés SSH pour chaque container
setup_ssh_keys() {
    echo -e "${CYAN}🔑 Configuration des clés SSH...${NC}"
    
    # Créer clé SSH si elle n'existe pas
    if [ ! -f ~/.ssh/veza_rsa ]; then
        echo -e "${BLUE}📋 Génération de la clé SSH...${NC}"
        ssh-keygen -t rsa -b 4096 -f ~/.ssh/veza_rsa -N "" -C "veza-development"
    fi
    
    # Containers qui ont besoin de rsync
    containers=("veza-backend" "veza-chat" "veza-stream" "veza-frontend")
    
    for container in "${containers[@]}"; do
        echo -e "${BLUE}🔐 Configuration SSH pour $container...${NC}"
        
        # Installer et configurer SSH dans le container
        incus exec "$container" -- bash -c '
            export DEBIAN_FRONTEND=noninteractive
            apt-get update -qq
            apt-get install -y openssh-server rsync
            
            # Configuration SSH
            mkdir -p /root/.ssh
            chmod 700 /root/.ssh
            
            # Configurer sshd
            sed -i "s/#PermitRootLogin prohibit-password/PermitRootLogin yes/" /etc/ssh/sshd_config
            sed -i "s/#PubkeyAuthentication yes/PubkeyAuthentication yes/" /etc/ssh/sshd_config
            sed -i "s/#PasswordAuthentication yes/PasswordAuthentication no/" /etc/ssh/sshd_config
            
            systemctl enable ssh
            systemctl restart ssh
        '
        
        # Copier la clé publique
        cat ~/.ssh/veza_rsa.pub | incus exec "$container" -- bash -c 'cat > /root/.ssh/authorized_keys'
        incus exec "$container" -- chmod 600 /root/.ssh/authorized_keys
        
        echo -e "${GREEN}✅ SSH configuré pour $container${NC}"
    done
}

# Configuration des mappings rsync
setup_rsync_mappings() {
    echo -e "${CYAN}🗂️ Configuration des mappings rsync...${NC}"
    
    # Créer le fichier de configuration rsync
    cat > scripts/rsync-config.conf << 'EOF'
# Configuration des mappings rsync pour Veza
# Format: SOURCE_LOCAL:CONTAINER:DESTINATION_REMOTE:SERVICE_NAME

# Backend Go
veza-backend-api/:veza-backend:/opt/veza/backend/:veza-backend

# Chat Server Rust
veza-chat-server/:veza-chat:/opt/veza/chat/:veza-chat

# Stream Server Rust
veza-stream-server/:veza-stream:/opt/veza/stream/:veza-stream

# Frontend React
veza-frontend/:veza-frontend:/opt/veza/frontend/:veza-frontend
EOF
    
    echo -e "${GREEN}✅ Configuration rsync créée${NC}"
}

# Script de synchronisation rapide
create_quick_sync_script() {
    echo -e "${CYAN}⚡ Création du script de synchronisation rapide...${NC}"
    
    cat > scripts/quick-sync.sh << 'EOF'
#!/bin/bash

# Script de synchronisation rapide
# Usage: ./scripts/quick-sync.sh [composant] [--build] [--restart]

set -e

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# Paramètres
COMPONENT="$1"
BUILD_FLAG="$2"
RESTART_FLAG="$3"

# Configuration SSH
SSH_KEY="$HOME/.ssh/veza_rsa"
SSH_OPTS="-i $SSH_KEY -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o LogLevel=ERROR"

# Fonction de synchronisation pour un composant
sync_component() {
    local component=$1
    local container=$2
    local local_path=$3
    local remote_path=$4
    local service=$5
    local build_needed=$6
    
    echo -e "${CYAN}🔄 Synchronisation $component...${NC}"
    
    # Obtenir l'IP du container
    local container_ip=$(incus ls "$container" -c 4 --format csv | cut -d' ' -f1)
    if [ -z "$container_ip" ]; then
        echo -e "${RED}❌ Impossible d'obtenir l'IP de $container${NC}"
        return 1
    fi
    
    # Vérifier que le répertoire local existe
    if [ ! -d "$local_path" ]; then
        echo -e "${RED}❌ Répertoire local $local_path non trouvé${NC}"
        return 1
    fi
    
    # Synchroniser avec rsync
    echo -e "${BLUE}📋 rsync: $local_path -> $container:$remote_path${NC}"
    rsync -avz --delete \
        --exclude='.git' \
        --exclude='node_modules' \
        --exclude='target' \
        --exclude='dist' \
        --exclude='build' \
        --exclude='.next' \
        --exclude='*.log' \
        -e "ssh $SSH_OPTS" \
        "$local_path" "root@$container_ip:$remote_path"
    
    # Build si nécessaire
    if [ "$build_needed" = "true" ] || [ "$BUILD_FLAG" = "--build" ]; then
        echo -e "${BLUE}🔨 Build de $component...${NC}"
        
        case $component in
            "backend")
                ssh $SSH_OPTS "root@$container_ip" "cd $remote_path && ./build.sh"
                ;;
            "chat"|"stream")
                ssh $SSH_OPTS "root@$container_ip" "cd $remote_path && ./build.sh"
                ;;
            "frontend")
                ssh $SSH_OPTS "root@$container_ip" "cd $remote_path && npm install"
                ;;
        esac
    fi
    
    # Redémarrer le service si demandé
    if [ "$BUILD_FLAG" = "--restart" ] || [ "$RESTART_FLAG" = "--restart" ]; then
        echo -e "${BLUE}🔄 Redémarrage du service $service...${NC}"
        ssh $SSH_OPTS "root@$container_ip" "systemctl restart $service"
        
        # Vérifier le statut
        sleep 2
        status=$(ssh $SSH_OPTS "root@$container_ip" "systemctl is-active $service" 2>/dev/null || echo "failed")
        
        case $status in
            "active")
                echo -e "${GREEN}✅ Service $service redémarré avec succès${NC}"
                ;;
            *)
                echo -e "${RED}❌ Échec redémarrage du service $service${NC}"
                echo -e "${YELLOW}Vérifiez les logs: incus exec $container -- journalctl -u $service -n 20${NC}"
                ;;
        esac
    fi
    
    echo -e "${GREEN}✅ $component synchronisé${NC}"
}

# Fonction de synchronisation complète
sync_all() {
    echo -e "${CYAN}🚀 Synchronisation complète de tous les composants...${NC}"
    echo ""
    
    # Lire la configuration
    while IFS=':' read -r local_path container remote_path service; do
        # Ignorer les commentaires et lignes vides
        [[ "$local_path" =~ ^#.*$ ]] && continue
        [[ -z "$local_path" ]] && continue
        
        # Déterminer le composant
        component=$(basename "$local_path" | sed 's/\/$//')
        component=$(echo "$component" | sed 's/veza-//' | sed 's/-server//' | sed 's/-api//')
        
        sync_component "$component" "$container" "$local_path" "$remote_path" "$service" "false"
        echo ""
    done < scripts/rsync-config.conf
}

# Afficher l'aide
show_help() {
    echo -e "${BLUE}🔄 Script de synchronisation rapide Veza${NC}"
    echo ""
    echo -e "${CYAN}Usage:${NC}"
    echo "  $0                    # Synchroniser tous les composants"
    echo "  $0 backend            # Synchroniser seulement le backend"
    echo "  $0 chat               # Synchroniser seulement le chat"
    echo "  $0 stream             # Synchroniser seulement le stream"
    echo "  $0 frontend           # Synchroniser seulement le frontend"
    echo "  $0 backend --build    # Synchroniser et builder"
    echo "  $0 backend --restart  # Synchroniser et redémarrer le service"
    echo ""
    echo -e "${CYAN}Exemples:${NC}"
    echo "  $0 backend --build --restart    # Sync + build + restart"
    echo "  $0 frontend --restart           # Sync frontend + restart"
}

# Fonction principale
main() {
    case "$COMPONENT" in
        "backend")
            sync_component "backend" "veza-backend" "veza-backend-api/" "/opt/veza/backend/" "veza-backend" "false"
            ;;
        "chat")
            sync_component "chat" "veza-chat" "veza-chat-server/" "/opt/veza/chat/" "veza-chat" "false"
            ;;
        "stream")
            sync_component "stream" "veza-stream" "veza-stream-server/" "/opt/veza/stream/" "veza-stream" "false"
            ;;
        "frontend")
            sync_component "frontend" "veza-frontend" "veza-frontend/" "/opt/veza/frontend/" "veza-frontend" "false"
            ;;
        "help"|"-h"|"--help")
            show_help
            ;;
        "")
            sync_all
            ;;
        *)
            echo -e "${RED}❌ Composant inconnu: $COMPONENT${NC}"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

main "$@"
EOF

    chmod +x scripts/quick-sync.sh
    echo -e "${GREEN}✅ Script de synchronisation rapide créé${NC}"
}

# Script de surveillance des changements avec inotify
create_watch_script() {
    echo -e "${CYAN}👁️ Création du script de surveillance automatique...${NC}"
    
    cat > scripts/watch-and-sync.sh << 'EOF'
#!/bin/bash

# Script de surveillance automatique avec inotify
# Synchronise automatiquement quand des fichiers changent

set -e

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# Vérifier inotify-tools
if ! command -v inotifywait &> /dev/null; then
    echo -e "${YELLOW}📦 Installation d'inotify-tools...${NC}"
    sudo dnf install -y inotify-tools || sudo apt-get install -y inotify-tools
fi

echo -e "${BLUE}"
echo "╭──────────────────────────────────────────╮"
echo "│     👁️ Surveillance Automatique des      │"
echo "│           Changements de Code             │"
echo "╰──────────────────────────────────────────╯"
echo -e "${NC}"

# Configuration
DEBOUNCE_TIME=2  # Attendre 2 secondes avant de synchroniser

# Fonction de surveillance pour un répertoire
watch_directory() {
    local dir=$1
    local component=$2
    
    echo -e "${CYAN}👁️ Surveillance de $dir ($component)...${NC}"
    
    inotifywait -m -r -e modify,create,delete,move \
        --exclude='(\.git|node_modules|target|dist|build|\.next|.*\.log)' \
        "$dir" | while read path action file; do
        
        echo -e "${YELLOW}📝 Changement détecté: $path$file ($action)${NC}"
        
        # Debounce - attendre un peu puis synchroniser
        sleep $DEBOUNCE_TIME
        
        echo -e "${BLUE}🔄 Synchronisation de $component...${NC}"
        ./scripts/quick-sync.sh "$component" --restart
        
        echo -e "${GREEN}✅ $component synchronisé et redémarré${NC}"
        echo "────────────────────────────────────────"
    done
}

# Surveillance en parallèle
main() {
    local component="$1"
    
    case "$component" in
        "backend")
            watch_directory "veza-backend-api" "backend"
            ;;
        "chat")
            watch_directory "veza-chat-server" "chat"
            ;;
        "stream")
            watch_directory "veza-stream-server" "stream"
            ;;
        "frontend")
            watch_directory "veza-frontend" "frontend"
            ;;
        "all"|"")
            echo -e "${CYAN}🚀 Surveillance de tous les composants...${NC}"
            echo -e "${YELLOW}Appuyez sur Ctrl+C pour arrêter${NC}"
            echo ""
            
            # Lancer en arrière-plan
            watch_directory "veza-backend-api" "backend" &
            watch_directory "veza-chat-server" "chat" &
            watch_directory "veza-stream-server" "stream" &
            watch_directory "veza-frontend" "frontend" &
            
            # Attendre
            wait
            ;;
        "help"|"-h"|"--help")
            echo -e "${BLUE}👁️ Script de surveillance automatique${NC}"
            echo ""
            echo -e "${CYAN}Usage:${NC}"
            echo "  $0              # Surveiller tous les composants"
            echo "  $0 backend      # Surveiller seulement le backend"
            echo "  $0 chat         # Surveiller seulement le chat"
            echo "  $0 stream       # Surveiller seulement le stream"
            echo "  $0 frontend     # Surveiller seulement le frontend"
            ;;
        *)
            echo -e "${RED}❌ Composant inconnu: $component${NC}"
            exit 1
            ;;
    esac
}

main "$@"
EOF

    chmod +x scripts/watch-and-sync.sh
    echo -e "${GREEN}✅ Script de surveillance automatique créé${NC}"
}

# Script de débogage des connexions
create_debug_script() {
    echo -e "${CYAN}🔍 Création du script de débogage...${NC}"
    
    cat > scripts/debug-connections.sh << 'EOF'
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
EOF

    chmod +x scripts/debug-connections.sh
    echo -e "${GREEN}✅ Script de débogage créé${NC}"
}

# Fonction principale
main() {
    echo -e "${CYAN}🔄 Configuration rsync et synchronisation...${NC}"
    echo ""
    
    # Vérifier que les containers existent
    containers=("veza-backend" "veza-chat" "veza-stream" "veza-frontend")
    for container in "${containers[@]}"; do
        if ! incus list "$container" --format csv | grep -q "$container"; then
            echo -e "${RED}❌ Container $container non trouvé${NC}"
            echo -e "${YELLOW}Exécutez d'abord: ./scripts/setup-manual-containers.sh${NC}"
            exit 1
        fi
    done
    
    # Configuration complète
    setup_ssh_keys
    setup_rsync_mappings
    create_quick_sync_script
    create_watch_script
    create_debug_script
    
    echo ""
    echo -e "${GREEN}🎉 Configuration rsync terminée !${NC}"
    echo ""
    echo -e "${BLUE}💡 Commandes utiles:${NC}"
    echo -e "  - Sync tous: ${YELLOW}./scripts/quick-sync.sh${NC}"
    echo -e "  - Sync backend: ${YELLOW}./scripts/quick-sync.sh backend --restart${NC}"
    echo -e "  - Surveillance auto: ${YELLOW}./scripts/watch-and-sync.sh${NC}"
    echo -e "  - Debug connexions: ${YELLOW}./scripts/debug-connections.sh${NC}"
    echo ""
    echo -e "${CYAN}🚀 Prêt pour le développement !${NC}"
    echo -e "  1. Démarrez les services: ${YELLOW}./scripts/start-all-services.sh${NC}"
    echo -e "  2. Synchronisez le code: ${YELLOW}./scripts/quick-sync.sh${NC}"
    echo -e "  3. Activez la surveillance: ${YELLOW}./scripts/watch-and-sync.sh${NC}"
}

main "$@" 