# Documentation Complète - Veza Backend API

## Vue d'ensemble

Ce projet est une API backend développée en Go qui fonctionne avec des modules Rust (chat et stream server) et un frontend React. Cette documentation détaille chaque composant du système pour permettre une intégration parfaite entre tous les éléments.

## Architecture du Système

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Frontend      │    │   Backend API    │    │   Modules Rust  │
│   (React)       │◄──►│   (Go)           │◄──►│   (Chat/Stream) │
└─────────────────┘    └──────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
         └───────────────────────▼───────────────────────┘
                            PostgreSQL
```

## Composants Principaux

### 1. Backend API Go
- **Serveur HTTP** : Gin avec middleware d'authentification JWT
- **Base de données** : PostgreSQL avec migrations automatiques
- **WebSocket** : Support pour le chat temps réel
- **APIs REST** : Endpoints pour toutes les fonctionnalités

### 2. Modules Rust
- **Chat Server** : Serveur WebSocket pour messages en temps réel
- **Stream Server** : Serveur de streaming audio/vidéo

### 3. Frontend React
- **Interface utilisateur** : Application React moderne
- **Communication** : API REST + WebSocket

## Structure des Dossiers

```
veza-backend-api/
├── cmd/server/          # Point d'entrée principal
├── internal/            # Code interne du backend
│   ├── api/            # Handlers et routes API
│   ├── config/         # Configuration système
│   ├── database/       # Gestion base de données
│   ├── middleware/     # Middleware HTTP
│   ├── models/         # Modèles de données
│   ├── services/       # Logique métier
│   ├── utils/          # Utilitaires
│   └── websocket/      # Gestion WebSocket
├── modules/            # Modules Rust
│   ├── chat_server/    # Serveur de chat Rust
│   └── stream_server/  # Serveur de streaming Rust
├── pkg/               # Packages publics
└── docs/              # Documentation complète
```

## Technologies Utilisées

### Backend Go
- **Framework** : Gin (HTTP router)
- **Base de données** : PostgreSQL avec driver lib/pq
- **Authentication** : JWT avec golang-jwt/jwt
- **WebSocket** : gorilla/websocket
- **Configuration** : godotenv pour variables d'environnement

### Modules Rust
- **Framework async** : Tokio
- **WebSocket** : tokio-tungstenite
- **Sérialisation** : serde
- **Base de données** : sqlx

## Guide de Démarrage Rapide

### 1. Configuration
```bash
# Cloner le projet
git clone <repository>
cd veza-backend-api

# Configuration environnement
cp .env.example .env
# Éditer .env avec vos paramètres
```

### 2. Base de données
```bash
# Créer la base PostgreSQL
createdb veza_db

# Les migrations s'exécutent automatiquement au démarrage
```

### 3. Démarrage Backend Go
```bash
go mod tidy
go run cmd/server/main.go
```

### 4. Démarrage Modules Rust
```bash
# Chat server
cd modules/chat_server
cargo run

# Stream server
cd modules/stream_server
cargo run
```

## Documentation Détaillée

### Fichiers de Code Go
- [cmd/server/main.go](cmd-server-main.md) - Point d'entrée principal
- [internal/config/config.go](internal-config-config.md) - Configuration système
- [internal/database/connection.go](internal-database-connection.md) - Connexion base de données
- [internal/api/router.go](internal-api-router.md) - Routeur API principal
- [internal/models/user.go](internal-models-user.md) - Modèle utilisateur
- [internal/middleware/auth.go](internal-middleware-auth.md) - Middleware d'authentification
- [internal/websocket/chat.go](internal-websocket-chat.md) - Gestion WebSocket

### APIs REST
- [API Authentication](api-auth.md) - Endpoints d'authentification
- [API User](api-user.md) - Gestion des utilisateurs
- [API Chat](api-chat.md) - API de chat
- [API Track](api-track.md) - Gestion des pistes audio
- [API Listing](api-listing.md) - Gestion des annonces
- [API Offer](api-offer.md) - Gestion des offres

### Modules Rust
- [Chat Server](rust-chat-server.md) - Serveur de chat Rust
- [Stream Server](rust-stream-server.md) - Serveur de streaming Rust

### Intégration
- [Frontend Integration](frontend-integration.md) - Guide d'intégration frontend
- [WebSocket Protocol](websocket-protocol.md) - Protocole WebSocket
- [Authentication Flow](auth-flow.md) - Flux d'authentification

## Endpoints API Principaux

### Authentication
- `POST /api/v1/auth/register` - Inscription
- `POST /api/v1/auth/login` - Connexion
- `POST /api/v1/auth/refresh` - Rafraîchir token

### User Management
- `GET /api/v1/users/profile` - Profil utilisateur
- `PUT /api/v1/users/profile` - Mettre à jour profil
- `GET /api/v1/users/:id` - Obtenir utilisateur par ID

### Chat
- `GET /api/v1/chat/rooms` - Liste des salons
- `GET /api/v1/chat/rooms/:id/messages` - Messages d'un salon
- `POST /api/v1/chat/rooms/:id/messages` - Envoyer message

### WebSocket
- `GET /ws/chat?token=<jwt_token>` - Connexion WebSocket chat

## Configuration

### Variables d'Environnement
```env
# Serveur
PORT=8080
ENVIRONMENT=development

# Base de données
DATABASE_URL=postgres://user:password@localhost/veza_db
DB_HOST=localhost
DB_PORT=5432
DB_USERNAME=postgres
DB_PASSWORD=
DB_NAME=veza_db

# JWT
JWT_SECRET=your-super-secret-key
JWT_EXPIRATION=24h
```

### Configuration Rust
```toml
# Cargo.toml pour modules Rust
[dependencies]
tokio = { version = "1", features = ["full"] }
serde = { version = "1.0", features = ["derive"] }
sqlx = { version = "0.7", features = ["runtime-tokio-rustls", "postgres"] }
```

## Sécurité

### Authentification JWT
- Tokens sécurisés avec expiration
- Refresh tokens pour sessions longues
- Middleware de validation automatique

### CORS
- Configuration CORS pour développement et production
- Headers de sécurité appropriés

### Validation
- Validation des entrées utilisateur
- Sanitisation des données

## Déploiement

### Docker
```dockerfile
# Exemple Dockerfile pour backend Go
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod tidy && go build -o main cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
CMD ["./main"]
```

### Production
- Variables d'environnement de production
- Configuration SSL/TLS
- Monitoring et logging

## Support et Maintenance

Cette documentation couvre tous les aspects nécessaires pour :
- Intégrer le frontend React
- Utiliser les modules Rust
- Déployer en production
- Développer de nouvelles fonctionnalités
- Déboguer et maintenir le système

Pour des questions spécifiques, consultez les fichiers de documentation détaillée de chaque composant. 