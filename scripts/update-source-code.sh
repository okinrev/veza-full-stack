#!/bin/bash

# Script de mise à jour du code source dans les containers
# Supporte 3 méthodes : rsync, git, archive

set -e

WORKSPACE_DIR="$(cd "$(dirname "$0")/.." && pwd)"

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

echo -e "${BLUE}📝 Mise à jour du code source Veza...${NC}"

# Configuration des projets
declare -A PROJECTS=(
    ["backend"]="veza-backend-api:veza-backend:/app/backend"
    ["chat"]="veza-chat-server:veza-chat:/app/chat"
    ["stream"]="veza-stream-server:veza-stream:/app/stream"
    ["frontend"]="veza-frontend:veza-frontend:/app/frontend"
)

# Méthode 1: Rsync (recommandée pour le développement)
update_with_rsync() {
    local project_name=$1
    local project_info="${PROJECTS[$project_name]}"
    local source_dir=$(echo "$project_info" | cut -d':' -f1)
    local container=$(echo "$project_info" | cut -d':' -f2)
    local target_path=$(echo "$project_info" | cut -d':' -f3)
    
    echo -e "${CYAN}🔄 Rsync $project_name...${NC}"
    
    if [ ! -d "$WORKSPACE_DIR/$source_dir" ]; then
        echo -e "${RED}❌ Répertoire source manquant: $source_dir${NC}"
        return 1
    fi
    
    if ! incus list "$container" --format csv | grep -q RUNNING; then
        echo -e "${YELLOW}⚠️ Container $container non démarré${NC}"
        return 1
    fi
    
    # Synchronisation avec rsync via incus file push
    incus exec "$container" -- mkdir -p "$target_path"
    incus file push -r "$WORKSPACE_DIR/$source_dir/." "$container$target_path/"
    
    echo -e "${GREEN}✅ $project_name synchronisé${NC}"
}

# Méthode 2: Git clone/pull
update_with_git() {
    local project_name=$1
    local git_repo=$2
    local project_info="${PROJECTS[$project_name]}"
    local container=$(echo "$project_info" | cut -d':' -f2)
    local target_path=$(echo "$project_info" | cut -d':' -f3)
    
    echo -e "${CYAN}🔄 Git pull $project_name...${NC}"
    
    if ! incus list "$container" --format csv | grep -q RUNNING; then
        echo -e "${YELLOW}⚠️ Container $container non démarré${NC}"
        return 1
    fi
    
    incus exec "$container" -- bash -c "
        if [ -d '$target_path/.git' ]; then
            cd '$target_path' && git pull
        else
            rm -rf '$target_path'
            git clone '$git_repo' '$target_path'
        fi
    "
    
    echo -e "${GREEN}✅ $project_name mis à jour via Git${NC}"
}

# Méthode 3: Archive tar.gz
update_with_archive() {
    local project_name=$1
    local project_info="${PROJECTS[$project_name]}"
    local source_dir=$(echo "$project_info" | cut -d':' -f1)
    local container=$(echo "$project_info" | cut -d':' -f2)
    local target_path=$(echo "$project_info" | cut -d':' -f3)
    
    echo -e "${CYAN}🔄 Archive $project_name...${NC}"
    
    if [ ! -d "$WORKSPACE_DIR/$source_dir" ]; then
        echo -e "${RED}❌ Répertoire source manquant: $source_dir${NC}"
        return 1
    fi
    
    # Créer l'archive
    local archive_name="${project_name}-$(date +%Y%m%d-%H%M%S).tar.gz"
    local archive_path="/tmp/$archive_name"
    
    tar -czf "$archive_path" -C "$WORKSPACE_DIR" "$source_dir"
    
    # Envoyer et extraire dans le container
    incus file push "$archive_path" "$container/tmp/"
    incus exec "$container" -- bash -c "
        mkdir -p '$target_path'
        cd /tmp
        tar -xzf '$archive_name' --strip-components=1 -C '$target_path'
        rm -f '$archive_name'
    "
    
    # Nettoyer
    rm -f "$archive_path"
    
    echo -e "${GREEN}✅ $project_name déployé via archive${NC}"
}

# Compilation après mise à jour
compile_project() {
    local project_name=$1
    local project_info="${PROJECTS[$project_name]}"
    local container=$(echo "$project_info" | cut -d':' -f2)
    local target_path=$(echo "$project_info" | cut -d':' -f3)
    
    echo -e "${CYAN}🔨 Compilation $project_name...${NC}"
    
    case "$project_name" in
        "backend")
            incus exec "$container" -- bash -c "
                cd '$target_path'
                export PATH=/usr/local/go/bin:\$PATH
                go mod tidy
                go build -o veza-backend ./cmd/server/main.go
            "
            ;;
        "chat"|"stream")
            incus exec "$container" -- bash -c "
                cd '$target_path'
                export PATH=/root/.cargo/bin:\$PATH
                cargo build --release
            "
            ;;
        "frontend")
            incus exec "$container" -- bash -c "
                cd '$target_path'
                npm install
                # npm run build  # Optionnel pour production
            "
            ;;
    esac
    
    echo -e "${GREEN}✅ $project_name compilé${NC}"
}

# Redémarrage des services
restart_services() {
    echo -e "${BLUE}🔄 Redémarrage des services...${NC}"
    
    local services=("veza-backend" "veza-chat" "veza-stream" "veza-frontend")
    
    for service in "${services[@]}"; do
        local container_name=$(echo "$service" | sed 's/veza-/veza-/')
        if incus list "$container_name" --format csv | grep -q RUNNING; then
            echo -e "${CYAN}🔄 Redémarrage $service...${NC}"
            incus exec "$container_name" -- systemctl restart "$service" 2>/dev/null || true
        fi
    done
    
    echo -e "${GREEN}✅ Services redémarrés${NC}"
}

# Fonction d'aide
show_help() {
    echo -e "${BLUE}Usage:${NC}"
    echo -e "  $0 <method> [project]"
    echo ""
    echo -e "${BLUE}Méthodes:${NC}"
    echo -e "  ${GREEN}rsync${NC}     - Synchronisation directe (recommandé pour dev)"
    echo -e "  ${GREEN}git${NC}       - Via Git clone/pull (nécessite repo distant)"
    echo -e "  ${GREEN}archive${NC}   - Via archive tar.gz"
    echo -e "  ${GREEN}all${NC}       - Tous les projets avec rsync"
    echo ""
    echo -e "${BLUE}Projets:${NC}"
    for project in "${!PROJECTS[@]}"; do
        echo -e "  • ${CYAN}$project${NC}"
    done
    echo ""
    echo -e "${BLUE}Exemples:${NC}"
    echo -e "  $0 rsync backend    # Sync backend via rsync"
    echo -e "  $0 git backend      # Update backend via git"
    echo -e "  $0 all              # Sync tous les projets"
}

# Fonction principale
main() {
    local method=${1:-help}
    local project=${2:-}
    
    case "$method" in
        "rsync")
            if [ -z "$project" ]; then
                echo -e "${RED}❌ Projet requis pour rsync${NC}"
                show_help
                exit 1
            fi
            
            if [ -z "${PROJECTS[$project]}" ]; then
                echo -e "${RED}❌ Projet inconnu: $project${NC}"
                exit 1
            fi
            
            update_with_rsync "$project"
            compile_project "$project"
            restart_services
            ;;
        "git")
            if [ -z "$project" ]; then
                echo -e "${RED}❌ Projet requis pour git${NC}"
                show_help
                exit 1
            fi
            
            # Pour Git, il faudrait configurer les URLs des dépôts
            echo -e "${YELLOW}⚠️ Méthode Git pas encore configurée${NC}"
            echo -e "${BLUE}💡 Configurez les URLs Git dans le script${NC}"
            ;;
        "archive")
            if [ -z "$project" ]; then
                echo -e "${RED}❌ Projet requis pour archive${NC}"
                show_help
                exit 1
            fi
            
            if [ -z "${PROJECTS[$project]}" ]; then
                echo -e "${RED}❌ Projet inconnu: $project${NC}"
                exit 1
            fi
            
            update_with_archive "$project"
            compile_project "$project"
            restart_services
            ;;
        "all")
            echo -e "${BLUE}🔄 Synchronisation complète avec rsync...${NC}"
            for project in "${!PROJECTS[@]}"; do
                update_with_rsync "$project"
                compile_project "$project"
            done
            restart_services
            ;;
        "help"|"-h"|"--help")
            show_help
            ;;
        *)
            echo -e "${RED}❌ Méthode inconnue: $method${NC}"
            show_help
            exit 1
            ;;
    esac
}

main "$@" 