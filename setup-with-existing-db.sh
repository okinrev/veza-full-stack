#!/bin/bash

# Script de configuration pour Veza avec base de données existante
# Ce script configure l'application en préservant vos données existantes

set -e

# Couleurs pour les messages
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${PURPLE}"
echo "╭─────────────────────────────────────────────────╮"
echo "│   🎵 Veza - Configuration avec BDD Existante   │"
echo "╰─────────────────────────────────────────────────╯"
echo -e "${NC}"

# Vérifier si le dump de base de données existe
DB_DUMP_FILE=""
if [ -f "veza_db_dump_21_06_2025.sql" ]; then
    DB_DUMP_FILE="veza_db_dump_21_06_2025.sql"
elif [ -f "veza_db_dump.sql" ]; then
    DB_DUMP_FILE="veza_db_dump.sql"
else
    echo -e "${YELLOW}📁 Chercher le fichier dump de base de données...${NC}"
    echo "Fichiers SQL trouvés :"
    find . -name "*.sql" -type f | head -10
    echo ""
    read -p "Entrez le chemin vers votre dump de base de données : " DB_DUMP_FILE
    
    if [ ! -f "$DB_DUMP_FILE" ]; then
        echo -e "${RED}❌ Fichier non trouvé : $DB_DUMP_FILE${NC}"
        exit 1
    fi
fi

echo -e "${GREEN}✅ Dump de base de données trouvé : $DB_DUMP_FILE${NC}"

# Configuration initiale
echo -e "${BLUE}🔧 Configuration initiale...${NC}"
mkdir -p logs uploads audio ssl backups haproxy-errors

if [ ! -f ".env" ]; then
    cp env.example .env
    echo -e "${CYAN}📄 Fichier .env créé à partir de l'exemple${NC}"
else
    echo -e "${GREEN}✅ Fichier .env déjà présent${NC}"
fi

# Demander les paramètres de connexion à la base de données existante
echo -e "${BLUE}🗄️ Configuration de la base de données existante${NC}"
echo "Entrez les paramètres de connexion à votre base de données :"

read -p "Host de la base de données [localhost] : " DB_HOST
DB_HOST=${DB_HOST:-localhost}

read -p "Port de la base de données [5432] : " DB_PORT
DB_PORT=${DB_PORT:-5432}

read -p "Nom de la base de données [veza_db] : " DB_NAME
DB_NAME=${DB_NAME:-veza_db}

read -p "Utilisateur de la base de données [veza_user] : " DB_USER
DB_USER=${DB_USER:-veza_user}

read -s -p "Mot de passe de la base de données : " DB_PASSWORD
echo ""

# Mettre à jour le fichier .env avec les paramètres de base de données
echo -e "${YELLOW}📝 Mise à jour du fichier .env...${NC}"
sed -i "s/DB_HOST=.*/DB_HOST=$DB_HOST/" .env
sed -i "s/DB_PORT=.*/DB_PORT=$DB_PORT/" .env
sed -i "s/DB_NAME=.*/DB_NAME=$DB_NAME/" .env
sed -i "s/DB_USERNAME=.*/DB_USERNAME=$DB_USER/" .env
sed -i "s/DB_PASSWORD=.*/DB_PASSWORD=$DB_PASSWORD/" .env
sed -i "s|DATABASE_URL=.*|DATABASE_URL=postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable|" .env

echo -e "${GREEN}✅ Configuration de base de données mise à jour${NC}"

# Option de démarrage
echo ""
echo -e "${BLUE}🚀 Options de démarrage :${NC}"
echo "1. Utiliser ma base de données existante (externe)"
echo "2. Importer ma base de données dans Docker et la migrer"
echo "3. Démarrer sans base de données (configuration manuelle)"

read -p "Choisissez une option [1-3] : " START_OPTION

case $START_OPTION in
    1)
        echo -e "${BLUE}🔄 Configuration pour base de données externe...${NC}"
        
        # Tester la connexion
        echo -e "${YELLOW}🔍 Test de connexion à la base de données...${NC}"
        if command -v psql &> /dev/null; then
            if PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "\q" 2>/dev/null; then
                echo -e "${GREEN}✅ Connexion à la base de données réussie${NC}"
                
                # Demander si on veut exécuter la migration
                read -p "Voulez-vous exécuter la migration pour ajouter les nouvelles fonctionnalités ? [y/N] : " RUN_MIGRATION
                if [[ $RUN_MIGRATION =~ ^[Yy]$ ]]; then
                    echo -e "${BLUE}🔄 Exécution de la migration...${NC}"
                    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f migrate-existing-db.sql
                    echo -e "${GREEN}✅ Migration terminée${NC}"
                fi
            else
                echo -e "${RED}❌ Impossible de se connecter à la base de données${NC}"
                echo "Vérifiez vos paramètres de connexion et réessayez"
                exit 1
            fi
        else
            echo -e "${YELLOW}⚠️ psql non installé - impossible de tester la connexion${NC}"
            echo "Assurez-vous que votre base de données est accessible"
        fi
        
        # Modifier docker-compose pour utiliser la base externe
        echo -e "${YELLOW}📝 Configuration Docker pour base externe...${NC}"
        cp docker-compose.yml docker-compose.yml.backup
        
        # Commenter le service postgres dans docker-compose
        sed -i '/postgres:/,/networks:/ { /postgres:/!{ /networks:/!d; }; }' docker-compose.yml
        
        echo -e "${BLUE}🚀 Démarrage des services (sans PostgreSQL)...${NC}"
        docker-compose up -d --scale postgres=0
        ;;
    
    2)
        echo -e "${BLUE}📦 Configuration avec Docker et import de base de données...${NC}"
        
        # Copier le dump dans le dossier d'initialisation
        cp "$DB_DUMP_FILE" ./init-existing-db.sql
        
        # Modifier docker-compose pour inclure le dump
        if ! grep -q "init-existing-db.sql" docker-compose.yml; then
            sed -i '/init-db.sql/a\      - ./init-existing-db.sql:/docker-entrypoint-initdb.d/01-existing-db.sql' docker-compose.yml
        fi
        
        echo -e "${BLUE}🚀 Démarrage des services avec import...${NC}"
        docker-compose up -d
        
        # Attendre que PostgreSQL soit prêt
        echo -e "${YELLOW}⏳ Attente du démarrage de PostgreSQL...${NC}"
        sleep 15
        
        # Exécuter la migration
        echo -e "${BLUE}🔄 Exécution de la migration...${NC}"
        docker-compose exec postgres psql -U veza_user -d veza_db -f /docker-entrypoint-initdb.d/migrate-existing-db.sql
        ;;
    
    3)
        echo -e "${BLUE}🚀 Démarrage basique des services...${NC}"
        docker-compose up -d
        echo -e "${YELLOW}⚠️ Configuration manuelle requise pour la base de données${NC}"
        ;;
    
    *)
        echo -e "${RED}❌ Option invalide${NC}"
        exit 1
        ;;
esac

# Installation des dépendances si Make est disponible
if command -v make &> /dev/null; then
    echo -e "${BLUE}📦 Installation des dépendances...${NC}"
    make install-deps
fi

# Affichage des informations finales
echo ""
echo -e "${GREEN}🎉 Configuration terminée avec succès !${NC}"
echo ""
echo -e "${BLUE}📋 Informations de connexion :${NC}"
echo -e "  🌐 Application : ${GREEN}http://localhost${NC}"
echo -e "  📊 HAProxy Stats : ${GREEN}http://localhost:8404/stats${NC} (admin:veza-admin-2024!)"
echo -e "  🎨 Frontend (dev) : ${GREEN}http://localhost:5173${NC}"
echo -e "  ⚙️ Backend API : ${GREEN}http://localhost:8080${NC}"
echo -e "  💬 Chat WebSocket : ${GREEN}ws://localhost:8081/ws${NC}"
echo -e "  🎵 Stream WebSocket : ${GREEN}ws://localhost:8082/ws${NC}"
echo ""
echo -e "${BLUE}📝 Commandes utiles :${NC}"
echo -e "  📊 Voir les logs : ${YELLOW}make logs${NC}"
echo -e "  🏥 Vérifier la santé : ${YELLOW}make health${NC}"
echo -e "  ⏹️ Arrêter les services : ${YELLOW}make docker-down${NC}"
echo ""
echo -e "${CYAN}💡 Consultez le README.md pour plus d'informations${NC}" 