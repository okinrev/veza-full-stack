#!/bin/bash

# 🚀 Script de démarrage du serveur production Veza

echo "🚀 DÉMARRAGE SERVEUR PRODUCTION VEZA"
echo "===================================="
echo ""

# Couleurs pour les logs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# Configuration de l'environnement
export ENVIRONMENT=production
export PORT=8080
export DATABASE_HOST=localhost
export DATABASE_PORT=5432
export DATABASE_USER=postgres
export DATABASE_PASSWORD=
export DATABASE_NAME=veza_production
export JWT_ACCESS_SECRET=production-super-secret-key-change-me

echo -e "${BLUE}📋 Configuration:${NC}"
echo -e "   Environment: ${GREEN}$ENVIRONMENT${NC}"
echo -e "   Port: ${GREEN}$PORT${NC}"
echo -e "   Database: ${GREEN}$DATABASE_NAME${NC}"
echo ""

# Vérification des prérequis
echo -e "${YELLOW}🔍 Vérification des prérequis...${NC}"

# Go version
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go n'est pas installé${NC}"
    exit 1
fi

GO_VERSION=$(go version | cut -d' ' -f3)
echo -e "   Go version: ${GREEN}$GO_VERSION${NC}"

# PostgreSQL
if ! command -v psql &> /dev/null; then
    echo -e "${YELLOW}⚠️  PostgreSQL client pas trouvé, mais peut être sur un serveur distant${NC}"
else
    echo -e "   PostgreSQL: ${GREEN}✓ Disponible${NC}"
fi

echo ""

# Compilation
echo -e "${YELLOW}🔨 Compilation du serveur production...${NC}"
cd /home/senke/Documents/veza-full-stack/veza-backend-api

# Build du serveur production
go build -o ./tmp/production-server ./cmd/production-server/main.go

if [ $? -ne 0 ]; then
    echo -e "${RED}❌ Erreur de compilation${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Compilation réussie${NC}"
echo ""

# Préparation base de données (optionnel)
echo -e "${YELLOW}🗄️  Préparation base de données...${NC}"
echo -e "${BLUE}   Note: Le serveur va tenter de se connecter automatiquement${NC}"
echo ""

# Démarrage du serveur
echo -e "${PURPLE}🚀 DÉMARRAGE DU SERVEUR PRODUCTION...${NC}"
echo ""

./tmp/production-server

# Note: Le serveur s'arrêtera avec Ctrl+C
echo ""
echo -e "${GREEN}✅ Arrêt propre du serveur production${NC}" 