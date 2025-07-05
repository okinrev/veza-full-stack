# 🚀 Guide de Démarrage Rapide - Veza Platform

## 🎯 Objectif

Ce guide vous permet de configurer et démarrer l'ensemble de la plateforme Veza en moins de 30 minutes, du clone du repository au serveur fonctionnel.

## ✅ Prérequis

### 💻 Système d'exploitation

- **Linux** : Ubuntu 20.04+ / Debian 11+ / CentOS 8+ / Arch Linux
- **macOS** : 10.15+ (Catalina et plus récent)
- **Windows** : Windows 10/11 avec WSL2

### 🛠️ Outils requis

| Outil | Version | Installation |
|-------|---------|-------------|
| **Git** | 2.30+ | `sudo apt install git` |
| **Go** | 1.21+ | [Installation Go](https://golang.org/doc/install) |
| **Rust** | 1.70+ | `curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs \| sh` |
| **Node.js** | 18+ | `curl -fsSL https://deb.nodesource.com/setup_18.x \| sudo -E bash -` |
| **Docker** | 20.10+ | [Installation Docker](https://docs.docker.com/get-docker/) |
| **Docker Compose** | 2.0+ | Inclus avec Docker Desktop |
| **Make** | 4.0+ | `sudo apt install build-essential` |

### ☁️ Outils optionnels (recommandés)

```bash
# Air pour hot reload Go
go install github.com/cosmtrek/air@latest

# golangci-lint pour linting Go
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.54.2

# cargo-watch pour hot reload Rust
cargo install cargo-watch

# pnpm pour package management JS (plus rapide que npm)
npm install -g pnpm
```

## 📦 Installation complète

### 1. 📥 Clone du repository

```bash
# Clone du repository principal
git clone https://github.com/okinrev/veza-full-stack.git
cd veza-full-stack

# Vérification des submodules (si applicable)
git submodule update --init --recursive
```

### 2. 🔧 Configuration des environnements

#### Backend API (Go)

```bash
cd veza-backend-api

# Copie et configuration du fichier d'environnement
cp config.example.env .env

# Édition du fichier .env
nano .env
```

**Configuration `.env` minimale** :
```bash
# Base de données
DATABASE_URL=postgresql://veza:password@localhost:5432/veza_dev?sslmode=disable

# JWT
JWT_SECRET=your-super-secret-jwt-key-change-in-production-256-bits-minimum

# Redis
REDIS_URL=redis://localhost:6379/0

# Server
SERVER_PORT=8080
ENVIRONMENT=development

# Logs
LOG_LEVEL=debug
```

#### Chat Server (Rust)

```bash
cd ../veza-chat-server

# Configuration Rust
cp .env.example .env
nano .env
```

**Configuration Chat Server** :
```bash
# Base de données
DATABASE_URL=postgresql://veza:password@localhost:5432/veza_dev

# Server
SERVER_HOST=127.0.0.1
SERVER_PORT=3001

# JWT pour WebSocket
JWT_SECRET=your-super-secret-jwt-key-change-in-production-256-bits-minimum

# Redis pour pub/sub
REDIS_URL=redis://127.0.0.1:6379/1

# NATS Event Bus
NATS_URL=nats://localhost:4222
```

#### Stream Server (Rust)

```bash
cd ../veza-stream-server

# Configuration Stream
cp .env.example .env
nano .env
```

**Configuration Stream Server** :
```bash
# Base de données
DATABASE_URL=postgresql://veza:password@localhost:5432/veza_dev

# Server
SERVER_HOST=127.0.0.1
SERVER_PORT=3002

# Storage (S3 compatible)
AWS_ACCESS_KEY_ID=minioadmin
AWS_SECRET_ACCESS_KEY=minioadmin
AWS_ENDPOINT_URL=http://localhost:9000
AWS_BUCKET_NAME=veza-audio-files

# Redis Cache
REDIS_URL=redis://127.0.0.1:6379/2
```

### 3. 🗄️ Infrastructure avec Docker

```bash
# Retour au répertoire racine
cd ..

# Démarrage de l'infrastructure
docker-compose up -d postgres redis nats minio

# Vérification que tous les services sont démarrés
docker-compose ps
```

**Services démarrés** :
- **PostgreSQL** : Port 5432
- **Redis** : Port 6379
- **NATS** : Port 4222
- **MinIO** (S3) : Port 9000 (Console: 9001)

### 4. 🗄️ Initialisation de la base de données

#### Création de la base et des utilisateurs

```bash
# Connexion à PostgreSQL
docker-compose exec postgres psql -U postgres

-- Création utilisateur et base
CREATE USER veza WITH PASSWORD 'password';
CREATE DATABASE veza_dev OWNER veza;
GRANT ALL PRIVILEGES ON DATABASE veza_dev TO veza;

-- Extensions nécessaires
\c veza_dev
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

\q
```

#### Migrations automatiques

```bash
# Backend API - Migrations Go
cd veza-backend-api
go run cmd/server/main.go --migrate-only

# Chat Server - Migrations Rust
cd ../veza-chat-server
cargo run --bin migrate

# Stream Server - Migrations Rust  
cd ../veza-stream-server
cargo run --bin migrate
```

### 5. 📦 Installation des dépendances

#### Backend API (Go)

```bash
cd veza-backend-api

# Download des dépendances Go
go mod download
go mod verify

# Génération des fichiers protobuf (si applicable)
make gen-proto

# Build pour vérifier que tout compile
go build -o server cmd/server/main.go
```

#### Chat Server (Rust)

```bash
cd ../veza-chat-server

# Build des dépendances Rust
cargo build

# Génération des protobuf Rust
cargo build --features="gen-proto"
```

#### Stream Server (Rust)

```bash
cd ../veza-stream-server

# Build du stream server
cargo build

# Build des outils supplémentaires
cd tools
cargo build
```

#### Frontend (optionnel)

```bash
cd ../veza-frontend

# Installation des dépendances Node.js
pnpm install

# Build de production
pnpm build
```

## 🚀 Démarrage des services

### Option 1 : 🐳 Avec Docker Compose (Recommandé)

```bash
# Démarrage complet avec Docker
docker-compose up -d

# Vérification des logs
docker-compose logs -f backend-api
docker-compose logs -f chat-server
docker-compose logs -f stream-server
```

### Option 2 : 💻 Développement local

#### Terminal 1 : Backend API

```bash
cd veza-backend-api

# Hot reload avec Air
air

# Ou démarrage normal
go run cmd/server/main.go
```

#### Terminal 2 : Chat Server

```bash
cd veza-chat-server

# Hot reload avec cargo-watch
cargo watch -x run

# Ou démarrage normal
cargo run
```

#### Terminal 3 : Stream Server

```bash
cd veza-stream-server

# Hot reload avec cargo-watch  
cargo watch -x run

# Ou démarrage normal
cargo run
```

#### Terminal 4 : Frontend (optionnel)

```bash
cd veza-frontend

# Développement avec hot reload
pnpm dev

# Le frontend sera disponible sur http://localhost:3000
```

## ✅ Vérification de l'installation

### 🔍 Health checks

#### Backend API
```bash
curl http://localhost:8080/health
# Réponse attendue : {"status":"healthy","service":"veza-backend-dev",...}
```

#### Chat Server
```bash
curl http://localhost:3001/health
# Réponse attendue : {"status":"healthy","service":"veza-chat-server",...}
```

#### Stream Server
```bash
curl http://localhost:3002/health  
# Réponse attendue : {"status":"healthy","service":"veza-stream-server",...}
```

### 🗄️ Bases de données

```bash
# Test PostgreSQL
docker-compose exec postgres psql -U veza -d veza_dev -c "SELECT COUNT(*) FROM users;"

# Test Redis
docker-compose exec redis redis-cli ping
# Réponse attendue : PONG

# Test NATS
docker-compose exec nats nats server info
```

### 🌐 Interfaces web

- **Backend API** : http://localhost:8080/api/v1/
- **Chat Server** : ws://localhost:3001/ws/chat
- **Stream Server** : http://localhost:3002/stream/
- **Frontend** : http://localhost:3000
- **MinIO Console** : http://localhost:9001 (admin/password)

## 🧪 Tests de fonctionnement

### 1. 🔐 Test d'authentification

```bash
# Inscription d'un utilisateur test
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com", 
    "password": "password123"
  }'

# Connexion
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

### 2. 💬 Test WebSocket Chat

```javascript
// Dans la console du navigateur ou avec un client WebSocket
const ws = new WebSocket('ws://localhost:3001/ws/chat?token=YOUR_JWT_TOKEN');

ws.onopen = () => {
  console.log('Connecté au chat');
  ws.send(JSON.stringify({
    type: 'join',
    room_id: 'general'
  }));
};

ws.onmessage = (event) => {
  console.log('Message reçu:', JSON.parse(event.data));
};
```

### 3. 🎵 Test Upload Audio

```bash
# Upload d'un fichier audio test
curl -X POST http://localhost:3002/api/v1/upload \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -F "file=@/path/to/your/audio.mp3" \
  -F "title=Test Track"
```

## 🐛 Dépannage

### ❌ Problèmes courants

#### 1. Erreur de connexion PostgreSQL

```bash
# Vérifier que PostgreSQL est démarré
docker-compose ps postgres

# Logs PostgreSQL
docker-compose logs postgres

# Recréer le container si nécessaire
docker-compose down postgres
docker-compose up -d postgres
```

#### 2. Erreur Redis connection

```bash
# Test de connexion Redis
docker-compose exec redis redis-cli ping

# Restart Redis si nécessaire
docker-compose restart redis
```

#### 3. Port déjà utilisé

```bash
# Vérifier les ports occupés
sudo netstat -tulpn | grep :8080

# Arrêter le processus utilisant le port
sudo kill -9 <PID>
```

#### 4. Permissions Docker

```bash
# Ajouter l'utilisateur au groupe docker
sudo usermod -aG docker $USER

# Se reconnecter pour appliquer les changements
newgrp docker
```

### 🔧 Commandes de diagnostic

```bash
# Status complet des services
docker-compose ps

# Logs en temps réel
docker-compose logs -f

# Utilisation des ressources
docker stats

# Nettoyage complet (attention : perte de données)
docker-compose down -v
docker system prune -a
```

## 📚 Prochaines étapes

Après cette installation réussie :

1. **📖 Documentation** : Lisez [Architecture et patterns](./architecture-patterns.md)
2. **🔧 IDE Setup** : Suivez [Environnement de développement](./development-environment.md)
3. **🧪 Tests** : Explorez [Testing Guide](./testing-guide.md)
4. **🚀 Première contrib** : [Première contribution](./first-contribution.md)

## 🆘 Support

Si vous rencontrez des problèmes :

1. **📋 Issues** : [GitHub Issues](https://github.com/okinrev/veza-full-stack/issues)
2. **💬 Slack** : #dev-help
3. **📧 Email** : dev-team@veza.com

---

## 📋 Checklist de validation

- [ ] Repository cloné et configuré
- [ ] Tous les prérequis installés
- [ ] Infrastructure Docker démarrée
- [ ] Base de données initialisée avec migrations
- [ ] Backend API démarré et health check OK
- [ ] Chat Server démarré et WebSocket fonctionnel
- [ ] Stream Server démarré et upload fonctionnel
- [ ] Tests d'authentification réussis
- [ ] Tests WebSocket réussis
- [ ] Tests upload audio réussis

**🎉 Félicitations ! Votre environnement Veza est opérationnel !**

---

**📝 Dernière mise à jour** : $(date)  
**👨‍💻 Maintenu par** : Équipe Backend Veza  
**🔄 Version** : 1.0.0  
**⏱️ Temps estimé** : 20-30 minutes 