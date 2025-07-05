---
id: development-environment
title: Environnement de développement
sidebar_label: Environnement de développement
description: Guide complet pour configurer l'environnement de développement Veza
---

# Environnement de développement

Ce guide détaille la configuration complète de l'environnement de développement pour le projet Veza, incluant tous les outils, dépendances et configurations nécessaires.

## Prérequis

### Système d'exploitation
- **Linux** : Ubuntu 20.04+, CentOS 8+, Fedora 33+
- **macOS** : 10.15+ (Catalina)
- **Windows** : 10+ avec WSL2 recommandé

### Outils de base
```bash
# Vérifier les versions requises
go version        # >= 1.21
rustc --version   # >= 1.70
node --version    # >= 18.0
docker --version  # >= 20.10
docker-compose --version  # >= 2.0
```

### Espace disque
- **Minimum** : 10 GB d'espace libre
- **Recommandé** : 20 GB pour les builds et caches

## Installation des outils

### 1. Go (Backend API)
```bash
# Ubuntu/Debian
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# macOS
brew install go

# Vérification
go version
go env GOPATH
```

### 2. Rust (Chat & Stream Servers)
```bash
# Installation via rustup
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
source ~/.cargo/env

# Vérification
rustc --version
cargo --version
```

### 3. Node.js (Documentation)
```bash
# Installation via nvm (recommandé)
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash
source ~/.bashrc

# Installation de Node.js LTS
nvm install --lts
nvm use --lts

# Vérification
node --version
npm --version
```

### 4. Docker & Docker Compose
```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install docker.io docker-compose
sudo usermod -aG docker $USER

# macOS
brew install --cask docker

# Vérification
docker --version
docker-compose --version
```

### 5. Outils de développement
```bash
# Git
sudo apt-get install git  # Ubuntu/Debian
brew install git          # macOS

# Make
sudo apt-get install make  # Ubuntu/Debian
brew install make          # macOS

# Outils de base de données
sudo apt-get install postgresql-client redis-tools  # Ubuntu/Debian
brew install postgresql redis                        # macOS
```

## Configuration de l'environnement

### 1. Variables d'environnement
```bash
# Créer le fichier .env dans la racine du projet
cp config.example.env .env

# Éditer les variables selon votre environnement
nano .env
```

**Configuration minimale pour le développement :**
```bash
# Base de données
DATABASE_URL=postgresql://veza:veza@localhost:5432/veza_dev
REDIS_URL=redis://localhost:6379

# Authentification
JWT_SECRET=your-super-secret-jwt-key-for-development
JWT_ACCESS_TOKEN_EXPIRY=15m
JWT_REFRESH_TOKEN_EXPIRY=7d

# Serveurs
API_PORT=8080
WS_PORT=8081
STREAM_PORT=8082

# Logging
LOG_LEVEL=debug
LOG_FORMAT=console
```

### 2. Configuration Git
```bash
# Configuration Git globale
git config --global user.name "Votre Nom"
git config --global user.email "votre.email@example.com"

# Configuration des hooks Git
cp scripts/git-hooks/* .git/hooks/
chmod +x .git/hooks/*
```

### 3. Configuration IDE

#### VS Code
```json
// .vscode/settings.json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.formatTool": "goimports",
  "rust-analyzer.checkOnSave.command": "clippy",
  "typescript.preferences.importModuleSpecifier": "relative",
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.organizeImports": true
  }
}
```

#### Extensions recommandées
- **Go** : Go (officiel)
- **Rust** : rust-analyzer
- **TypeScript** : TypeScript and JavaScript Language Features
- **Docker** : Docker
- **Git** : GitLens
- **Markdown** : Markdown All in One

## Services de base de données

### 1. PostgreSQL
```bash
# Installation locale
sudo apt-get install postgresql postgresql-contrib  # Ubuntu/Debian
brew install postgresql                              # macOS

# Démarrage du service
sudo systemctl start postgresql  # Ubuntu/Debian
brew services start postgresql    # macOS

# Création de la base de données
sudo -u postgres psql
CREATE DATABASE veza_dev;
CREATE USER veza WITH PASSWORD 'veza';
GRANT ALL PRIVILEGES ON DATABASE veza_dev TO veza;
\q
```

### 2. Redis
```bash
# Installation locale
sudo apt-get install redis-server  # Ubuntu/Debian
brew install redis                  # macOS

# Démarrage du service
sudo systemctl start redis  # Ubuntu/Debian
brew services start redis    # macOS

# Test de connexion
redis-cli ping
```

### 3. Docker Compose (Alternative)
```yaml
# docker-compose.dev.yml
version: '3.8'

services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: veza_dev
      POSTGRES_USER: veza
      POSTGRES_PASSWORD: veza
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

volumes:
  postgres_data:
  redis_data:
```

```bash
# Démarrage des services
docker-compose -f docker-compose.dev.yml up -d

# Vérification
docker-compose -f docker-compose.dev.yml ps
```

## Configuration des projets

### 1. Backend API (Go)
```bash
cd veza-backend-api

# Installation des dépendances
go mod download
go mod tidy

# Installation des outils de développement
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/go-delve/delve/cmd/dlv@latest

# Configuration de golangci-lint
cp .golangci.yml.example .golangci.yml

# Vérification de la configuration
go vet ./...
golangci-lint run
```

### 2. Chat Server (Rust)
```bash
cd veza-chat-server

# Installation des dépendances
cargo build

# Installation des outils de développement
cargo install cargo-watch
cargo install cargo-audit
cargo install cargo-tarpaulin

# Configuration de clippy
cp .clippy.toml.example .clippy.toml

# Vérification de la configuration
cargo clippy
cargo fmt --check
```

### 3. Stream Server (Rust)
```bash
cd veza-stream-server

# Installation des dépendances
cargo build

# Configuration spécifique au streaming
cp config/stream.toml.example config/stream.toml

# Vérification de la configuration
cargo clippy
cargo fmt --check
```

### 4. Documentation (Node.js)
```bash
cd veza-docs

# Installation des dépendances
npm install

# Configuration de Docusaurus
cp docusaurus.config.ts.example docusaurus.config.ts

# Vérification de la configuration
npm run build
```

## Scripts de développement

### 1. Scripts de démarrage
```bash
# scripts/dev-start.sh
#!/bin/bash

echo "🚀 Démarrage de l'environnement de développement..."

# Vérification des prérequis
echo "📋 Vérification des prérequis..."
command -v go >/dev/null 2>&1 || { echo "❌ Go n'est pas installé"; exit 1; }
command -v cargo >/dev/null 2>&1 || { echo "❌ Rust n'est pas installé"; exit 1; }
command -v node >/dev/null 2>&1 || { echo "❌ Node.js n'est pas installé"; exit 1; }
command -v docker >/dev/null 2>&1 || { echo "❌ Docker n'est pas installé"; exit 1; }

# Démarrage des services de base
echo "🗄️ Démarrage des services de base..."
docker-compose -f docker-compose.dev.yml up -d

# Attendre que les services soient prêts
echo "⏳ Attente de la disponibilité des services..."
until pg_isready -h localhost -p 5432 -U veza; do
    echo "⏳ Attente de PostgreSQL..."
    sleep 1
done

until redis-cli ping >/dev/null 2>&1; do
    echo "⏳ Attente de Redis..."
    sleep 1
done

echo "✅ Services de base prêts"

# Démarrage des applications
echo "🔧 Démarrage des applications..."

# Backend API
echo "🌐 Démarrage du Backend API..."
cd veza-backend-api
go run cmd/server/main.go &
API_PID=$!

# Chat Server
echo "💬 Démarrage du Chat Server..."
cd ../veza-chat-server
cargo run &
CHAT_PID=$!

# Stream Server
echo "🎵 Démarrage du Stream Server..."
cd ../veza-stream-server
cargo run &
STREAM_PID=$!

# Documentation
echo "📚 Démarrage de la documentation..."
cd ../veza-docs
npm run start &
DOCS_PID=$!

echo "✅ Tous les services sont démarrés"
echo "📊 PIDs: API=$API_PID, Chat=$CHAT_PID, Stream=$STREAM_PID, Docs=$DOCS_PID"

# Fonction de nettoyage
cleanup() {
    echo "🛑 Arrêt des services..."
    kill $API_PID $CHAT_PID $STREAM_PID $DOCS_PID 2>/dev/null
    exit 0
}

trap cleanup SIGINT SIGTERM

# Attendre indéfiniment
wait
```

### 2. Scripts de test
```bash
# scripts/dev-test.sh
#!/bin/bash

echo "🧪 Exécution des tests de développement..."

# Tests Go
echo "🔍 Tests Backend API..."
cd veza-backend-api
go test ./... -v

# Tests Rust
echo "🔍 Tests Chat Server..."
cd ../veza-chat-server
cargo test

echo "🔍 Tests Stream Server..."
cd ../veza-stream-server
cargo test

# Tests de documentation
echo "🔍 Tests de documentation..."
cd ../veza-docs
npm test

echo "✅ Tous les tests sont passés"
```

### 3. Scripts de build
```bash
# scripts/dev-build.sh
#!/bin/bash

echo "🔨 Build de l'environnement de développement..."

# Build Backend API
echo "🔨 Build Backend API..."
cd veza-backend-api
go build -o bin/server cmd/server/main.go

# Build Chat Server
echo "🔨 Build Chat Server..."
cd ../veza-chat-server
cargo build

# Build Stream Server
echo "🔨 Build Stream Server..."
cd ../veza-stream-server
cargo build

# Build Documentation
echo "🔨 Build Documentation..."
cd ../veza-docs
npm run build

echo "✅ Tous les builds sont terminés"
```

## Outils de développement

### 1. Linting et formatting
```bash
# Go
golangci-lint run
goimports -w .

# Rust
cargo clippy
cargo fmt

# TypeScript/JavaScript
npm run lint
npm run format
```

### 2. Tests
```bash
# Tests unitaires
go test ./... -v
cargo test
npm test

# Tests d'intégration
go test ./... -tags=integration
cargo test --features integration
npm run test:integration

# Tests de performance
go test ./... -bench=.
cargo bench
npm run test:performance
```

### 3. Debugging
```bash
# Go avec Delve
dlv debug cmd/server/main.go

# Rust avec GDB
cargo build
gdb target/debug/veza-chat-server

# Node.js avec inspect
node --inspect npm run start
```

## Monitoring de développement

### 1. Logs
```bash
# Affichage des logs en temps réel
tail -f logs/api.log
tail -f logs/chat.log
tail -f logs/stream.log

# Logs Docker
docker-compose -f docker-compose.dev.yml logs -f
```

### 2. Métriques
```bash
# Métriques Prometheus
curl http://localhost:9090/metrics

# Métriques de base de données
psql -h localhost -U veza -d veza_dev -c "SELECT * FROM pg_stat_activity;"

# Métriques Redis
redis-cli info
```

### 3. Profiling
```bash
# Profiling Go
go tool pprof http://localhost:8080/debug/pprof/profile

# Profiling Rust
cargo install flamegraph
cargo flamegraph

# Profiling Node.js
node --prof npm run start
node --prof-process isolate-*.log
```

## Dépannage

### Problèmes courants

#### 1. Ports déjà utilisés
```bash
# Vérifier les ports utilisés
sudo netstat -tulpn | grep :8080
sudo netstat -tulpn | grep :5432
sudo netstat -tulpn | grep :6379

# Tuer les processus
sudo kill -9 <PID>
```

#### 2. Permissions Docker
```bash
# Ajouter l'utilisateur au groupe docker
sudo usermod -aG docker $USER
newgrp docker
```

#### 3. Problèmes de base de données
```bash
# Réinitialiser la base de données
sudo -u postgres dropdb veza_dev
sudo -u postgres createdb veza_dev

# Appliquer les migrations
cd veza-backend-api
go run cmd/migrate/main.go
```

#### 4. Problèmes de cache
```bash
# Nettoyer les caches
go clean -cache
cargo clean
npm cache clean --force
```

## Ressources supplémentaires

### Documentation
- [Guide de démarrage rapide](./quick-start.md)
- [Architecture du projet](../architecture/backend-architecture.md)
- [Standards de code](./code-review.md)

### Outils externes
- [Go Documentation](https://golang.org/doc/)
- [Rust Book](https://doc.rust-lang.org/book/)
- [Docusaurus Documentation](https://docusaurus.io/docs)

### Support
- Issues GitHub : [Projet Veza](https://github.com/veza/veza-full-stack/issues)
- Discussions : [GitHub Discussions](https://github.com/veza/veza-full-stack/discussions)
- Documentation : [Veza Docs](https://docs.veza.app)

## Conclusion

Cet environnement de développement vous permet de travailler efficacement sur le projet Veza avec tous les outils et configurations nécessaires. Pour toute question ou problème, n'hésitez pas à consulter la documentation ou à créer une issue. 