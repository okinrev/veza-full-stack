#!/bin/bash

# Script de correction rapide pour la connexion à la base de données
set -e

echo "🔧 Correction de la connexion à la base de données Veza"

# Couleurs
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

# Obtenir l'IP réelle de PostgreSQL
POSTGRES_IP=$(incus list veza-postgres -c 4 --format csv | cut -d' ' -f1 | head -n1)

if [[ -z "$POSTGRES_IP" || "$POSTGRES_IP" == "-" ]]; then
    echo -e "${RED}❌ Container veza-postgres non trouvé ou sans IP${NC}"
    echo "Lancez d'abord l'infrastructure avec: ./scripts/incus-setup.sh && ./scripts/incus-deploy.sh"
    exit 1
fi

echo -e "${BLUE}📍 IP PostgreSQL détectée: $POSTGRES_IP${NC}"

# Corriger le fichier .env du backend
echo -e "${BLUE}🔧 Correction du fichier .env backend...${NC}"

cat > veza-backend-api/.env << EOF
# Veza Backend Configuration - Auto-corrigé
DATABASE_URL=postgres://veza:veza_password@${POSTGRES_IP}:5432/veza_db?sslmode=disable
REDIS_URL=redis://10.5.191.186:6379
JWT_SECRET=veza_jwt_secret_key_2025_production
SERVER_PORT=8080
SERVER_HOST=0.0.0.0
LOG_LEVEL=info
ENVIRONMENT=production
UPLOAD_PATH=/app/uploads
MAX_FILE_SIZE=10485760
EOF

echo -e "${GREEN}✅ Fichier .env corrigé${NC}"

# Si le backend est en cours d'exécution dans un container, le corriger aussi
if incus list veza-backend -c s --format csv | grep -q RUNNING; then
    echo -e "${BLUE}🔧 Correction dans le container backend...${NC}"
    
    # Copier le nouveau .env dans le container
    incus file push veza-backend-api/.env veza-backend/app/veza-backend-api/.env
    
    # Redémarrer le service backend
    incus exec veza-backend -- systemctl restart veza-backend || true
    
    echo -e "${GREEN}✅ Service backend redémarré${NC}"
fi

# Test de connectivité
echo -e "${BLUE}🔍 Test de connectivité...${NC}"

# Tester PostgreSQL
if incus exec veza-postgres -- pg_isready -h localhost -p 5432 >/dev/null 2>&1; then
    echo -e "${GREEN}✅ PostgreSQL accessible${NC}"
else
    echo -e "${RED}❌ PostgreSQL inaccessible${NC}"
fi

# Attendre quelques secondes et tester le backend
sleep 5

if curl -s -o /dev/null http://10.5.191.241:8080 2>/dev/null; then
    echo -e "${GREEN}✅ Backend accessible${NC}"
else
    echo -e "${RED}❌ Backend inaccessible (peut prendre quelques secondes de plus)${NC}"
fi

echo ""
echo -e "${GREEN}🎉 Correction terminée !${NC}"
echo -e "${BLUE}💡 Vous pouvez maintenant relancer votre backend avec:${NC}"
echo "   cd veza-backend-api && ./bin/veza-api"
echo ""
echo -e "${BLUE}🌐 Ou accéder à l'application via HAProxy:${NC}"
echo "   http://10.5.191.133" 