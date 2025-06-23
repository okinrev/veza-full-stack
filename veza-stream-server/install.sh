#!/bin/bash

# Script d'installation pour Stream Server
# Usage: ./install.sh

set -e

GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}🚀 Installation de Stream Server${NC}"
echo ""

# Vérification de Rust
if ! command -v cargo &> /dev/null; then
    echo -e "${RED}❌ Rust n'est pas installé${NC}"
    echo -e "${YELLOW}Installation de Rust...${NC}"
    curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
    source ~/.cargo/env
    echo -e "${GREEN}✅ Rust installé${NC}"
else
    echo -e "${GREEN}✅ Rust détecté: $(rustc --version)${NC}"
fi

# Installation des dépendances système
echo -e "${YELLOW}📦 Installation des dépendances système...${NC}"
if command -v apt-get >/dev/null 2>&1; then
    sudo apt-get update && sudo apt-get install -y pkg-config libssl-dev jq
elif command -v dnf >/dev/null 2>&1; then
    sudo dnf install -y pkgconfig openssl-devel jq
elif command -v pacman >/dev/null 2>&1; then
    sudo pacman -S pkg-config openssl jq
else
    echo -e "${YELLOW}⚠️  Gestionnaire de paquets non reconnu - installation manuelle requise${NC}"
fi

# Configuration des répertoires
echo -e "${YELLOW}📁 Création des répertoires...${NC}"
mkdir -p audio logs

# Configuration du fichier .env
if [ ! -f .env ]; then
    echo -e "${YELLOW}⚙️  Configuration du fichier .env...${NC}"
    cp env.example .env
    
    # Génération d'une clé secrète
    if command -v openssl >/dev/null 2>&1; then
        SECRET_KEY=$(openssl rand -hex 32)
        sed -i "s/SECRET_KEY=.*/SECRET_KEY=$SECRET_KEY/" .env
        echo -e "${GREEN}✅ Clé secrète générée automatiquement${NC}"
    else
        echo -e "${YELLOW}⚠️  OpenSSL non trouvé - générez manuellement une clé secrète${NC}"
    fi
    
    echo -e "${GREEN}✅ Fichier .env créé${NC}"
else
    echo -e "${GREEN}✅ Fichier .env existant${NC}"
fi

# Compilation du projet principal
echo -e "${YELLOW}🔨 Compilation du projet principal...${NC}"
cargo build --release

# Compilation des outils
echo -e "${YELLOW}🔧 Compilation des outils...${NC}"
cd tools
cargo build --release
cd ..

echo ""
echo -e "${GREEN}🎉 Installation terminée avec succès !${NC}"
echo ""
echo -e "${YELLOW}Prochaines étapes :${NC}"
echo -e "1. Ajoutez vos fichiers audio dans le répertoire 'audio/'"
echo -e "2. Éditez le fichier '.env' selon vos besoins"
echo -e "3. Lancez le serveur avec: make run"
echo -e "4. Testez avec: curl http://localhost:8082/health"
echo ""
echo -e "${GREEN}🔗 Commandes utiles :${NC}"
echo -e "  make help       - Aide sur les commandes"
echo -e "  make run        - Lancer le serveur"
echo -e "  make dev        - Mode développement"
echo -e "  make test       - Lancer les tests"
echo -e "  make docker-run - Lancer avec Docker"
echo ""
echo -e "${GREEN}📚 Documentation complète disponible dans archives/README.md${NC}" 