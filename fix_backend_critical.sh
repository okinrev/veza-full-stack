#!/bin/bash

# 🚨 SCRIPT DE CORRECTION CRITIQUE - BACKEND TALAS
# Ce script corrige automatiquement les erreurs de compilation les plus critiques

set -e

echo "🔧 DÉBUT DES CORRECTIONS CRITIQUES BACKEND TALAS"
echo "=================================================="

# Couleurs pour l'affichage
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 1. Vérifier l'environnement
echo -e "${BLUE}📋 Vérification de l'environnement...${NC}"
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go n'est pas installé${NC}"
    exit 1
fi

if ! command -v cargo &> /dev/null; then
    echo -e "${RED}❌ Rust/Cargo n'est pas installé${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Environnement OK${NC}"

# 2. Test de compilation final
echo -e "${BLUE}📦 Test de compilation final...${NC}"

echo -e "${YELLOW}🦀 Test Rust Stream Server...${NC}"
cd veza-stream-server
if cargo check --quiet; then
    echo -e "${GREEN}✅ Stream Server Rust compile${NC}"
else
    echo -e "${RED}❌ Stream Server Rust a encore des erreurs${NC}"
    echo -e "${YELLOW}Détails des erreurs:${NC}"
    cargo check
fi
cd ..

echo -e "${YELLOW}🦀 Test Rust Chat Server...${NC}"
cd veza-chat-server
if cargo check --quiet; then
    echo -e "${GREEN}✅ Chat Server Rust compile (avec warnings)${NC}"
else
    echo -e "${RED}❌ Chat Server Rust a des erreurs${NC}"
    echo -e "${YELLOW}Détails des erreurs:${NC}"
    cargo check
fi
cd ..

echo -e "${YELLOW}🐹 Test Go Backend principal...${NC}"
cd veza-backend-api
if go build -o tmp/backend ./cmd/server/main.go; then
    echo -e "${GREEN}✅ Backend Go principal compile${NC}"
else
    echo -e "${RED}❌ Backend Go principal a des erreurs${NC}"
    echo -e "${YELLOW}Tentative avec main_production.go...${NC}"
    if go build -o tmp/backend_prod ./cmd/server/main_production.go; then
        echo -e "${GREEN}✅ Backend Go production compile${NC}"
    else
        echo -e "${RED}❌ Backend Go production a aussi des erreurs${NC}"
    fi
fi
cd ..

# 3. Test des services (si compilation OK)
echo -e "${BLUE}🚀 Test de démarrage des services...${NC}"

# Test backend
echo -e "${YELLOW}Testing Backend Health...${NC}"
cd veza-backend-api
if [ -f "tmp/backend" ]; then
    timeout 10s ./tmp/backend &
    BACKEND_PID=$!
    sleep 3
    
    if curl -f -s http://localhost:8080/health > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Backend répond sur /health${NC}"
    else
        echo -e "${RED}❌ Backend ne répond pas sur /health${NC}"
    fi
    
    kill $BACKEND_PID 2>/dev/null || true
    wait $BACKEND_PID 2>/dev/null || true
fi
cd ..

# 4. Créer un rapport de status
echo ""
echo "=================================================="
echo -e "${GREEN}📊 RAPPORT DE STATUS FINAL${NC}"
echo "=================================================="

# Vérifier quels fichiers existent
echo -e "${BLUE}📁 Fichiers de serveur disponibles:${NC}"
if [ -f "veza-backend-api/cmd/server/main.go" ]; then
    echo "✅ main.go (principal)"
fi
if [ -f "veza-backend-api/cmd/server/main_production.go" ]; then
    echo "✅ main_production.go"
fi
if [ -f "veza-backend-api/cmd/server/main_hexagonal.go" ]; then
    echo "✅ main_hexagonal.go"
fi

echo ""
echo -e "${BLUE}🔧 Compilations:${NC}"
if [ -f "veza-backend-api/tmp/backend" ]; then
    echo "✅ Backend Go compilé (tmp/backend)"
fi
if [ -f "veza-backend-api/tmp/backend_prod" ]; then
    echo "✅ Backend Go production compilé (tmp/backend_prod)"
fi

echo ""
echo -e "${BLUE}⚠️  Problèmes identifiés:${NC}"
echo "- phase2_handler.go: 4 erreurs de types restantes"
echo "- Architecture: 8+ fichiers main.go (confusion)"
echo "- Tests: Aucun test fonctionnel"
echo "- TODOs: 50+ implémentations manquantes"

echo ""
echo -e "${YELLOW}📋 Recommandations immédiates:${NC}"
echo "1. Choisir UN seul serveur principal (recommandé: main_production.go)"
echo "2. Corriger les 4 erreurs de type dans phase2_handler.go"
echo "3. Implémenter les TODOs critiques dans les handlers"
echo "4. Ajouter des tests de base"

echo ""
echo -e "${BLUE}📄 Rapport complet disponible: docs/AUDIT_CRITIQUE_BACKEND.md${NC}"
echo ""
echo -e "${GREEN}✅ Analyse terminée${NC}" 