# 📚 DOCUMENTATION ULTRA-COMPLÈTE DU PROJET VEZA

> **Plateforme de streaming audio et chat en temps réel avec architecture microservices**

## 🎯 Vue d'ensemble du Projet

**Veza** est une plateforme moderne de streaming audio et de chat en temps réel construite avec une architecture microservices robuste. Le projet comprend trois services principaux interconnectés :

- **veza-backend-api** : API REST principale en Go (Gin) avec authentification JWT
- **veza-chat-server** : Service de chat en temps réel en Rust avec WebSocket
- **veza-stream-server** : Service de streaming audio en Rust avec compression adaptative

### 🎵 Fonctionnalités Principales

- **Streaming Audio Adaptatif** : Support de multiples formats (MP3, AAC, OGG, FLAC) avec ajustement automatique de la qualité
- **Chat en Temps Réel** : Communication instantanée via WebSocket avec modération automatique
- **Authentification Sécurisée** : JWT, OAuth2, Magic Links
- **Analytics Avancés** : Métriques en temps réel, dashboards Grafana
- **Monitoring Complet** : Health checks, alertes, logs structurés

## 🏗️ Architecture Globale

```mermaid
graph TB
    subgraph "Frontend (React/TypeScript)"
        UI[Interface Utilisateur]
        WS[WebSocket Client]
        AUDIO[Audio Player]
    end
    
    subgraph "API Gateway & Load Balancer"
        LB[Load Balancer - Nginx]
        REST[REST API - Go/Gin]
        GRPC[gRPC Gateway]
    end
    
    subgraph "Services Backend"
        CHAT[Chat Server - Rust/Axum]
        STREAM[Stream Server - Rust/Axum]
        AUTH[Authentication Service]
        ANALYTICS[Analytics Service]
        NOTIFICATIONS[Notification Service]
    end
    
    subgraph "Infrastructure"
        DB[(PostgreSQL 15+)]
        REDIS[(Redis 7+)]
        NATS[NATS Event Bus]
        MONITORING[Prometheus/Grafana]
        STORAGE[Object Storage]
    end
    
    subgraph "External Services"
        OAUTH[OAuth Providers]
        CDN[CDN]
        EMAIL[Email Service]
    end
    
    UI --> LB
    WS --> CHAT
    AUDIO --> STREAM
    LB --> REST
    LB --> GRPC
    REST --> AUTH
    REST --> ANALYTICS
    GRPC --> CHAT
    GRPC --> STREAM
    CHAT --> DB
    STREAM --> DB
    CHAT --> REDIS
    STREAM --> REDIS
    CHAT --> NATS
    STREAM --> NATS
    REST --> MONITORING
    CHAT --> MONITORING
    STREAM --> MONITORING
    AUTH --> OAUTH
    NOTIFICATIONS --> EMAIL
    STREAM --> STORAGE
    STREAM --> CDN
```

## 📁 Structure de Documentation

### 🏛️ Architecture
- [Architecture Globale](./architecture/global-architecture.md) - Vue d'ensemble complète
- [Architecture Backend](./architecture/backend-architecture.md) - API REST en Go
- [Architecture Chat Server](./architecture/chat-server-architecture.md) - Service de chat Rust
- [Architecture Stream Server](./architecture/stream-server-architecture.md) - Service de streaming Rust
- [Configuration Backend](./architecture/backend-config.md) - Variables d'environnement et configs

### 🔌 API & Développement
- [Vue d'ensemble API](./api/api-overview.md) - Introduction aux APIs
- [Référence des Endpoints](./api/endpoints-reference.md) - Documentation complète des endpoints
- [API Backend](./api/backend-api.md) - Documentation spécifique du backend
- [Authentification](./api/authentication/) - JWT, OAuth2, Magic Links
- [WebSocket API](./api/websocket/) - API temps réel
- [gRPC API](./api/grpc/) - Communication inter-services

### 🗄️ Base de Données
- [Schéma Complet](./database/schema.md) - Structure des tables et relations
- [Migrations](./database/migrations/) - Scripts de migration
- [Requêtes Optimisées](./database/queries/) - Requêtes SQL optimisées
- [Relations](./database/relationships/) - Mapping des relations

### 🚀 Déploiement & Infrastructure
- [Guide de Déploiement](./deployment/deployment-guide.md) - Instructions de déploiement
- [Variables d'Environnement](./deployment/environment-variables.md) - Configuration
- [Pipeline CI/CD](./deployment/ci-cd/) - Intégration continue
- [Monitoring](./deployment/monitoring/) - Surveillance et alertes

### 🧪 Tests & Qualité
- [Stratégie de Tests](./testing/test-strategy.md) - Approche de test
- [Tests Unitaires](./testing/unit-tests/) - Tests unitaires par service
- [Tests d'Intégration](./testing/integration-tests/) - Tests d'intégration
- [Tests de Performance](./testing/performance-tests/) - Tests de charge

### 📊 Monitoring & Analytics
- [Métriques](./monitoring/metrics/) - Métriques Prometheus
- [Alertes](./monitoring/alerts/) - Configuration des alertes
- [Dashboards Grafana](./monitoring/grafana/) - Dashboards de monitoring
- [Logs](./monitoring/logs/) - Gestion des logs

### 🔒 Sécurité
- [Guide de Sécurité](./security/security-guide.md) - Bonnes pratiques
- [Authentification](./security/authentication/) - Mécanismes d'auth
- [Autorisation](./security/authorization/) - Gestion des permissions
- [Audit](./security/audit/) - Traçabilité et audit

## 🎯 Métriques du Projet

### 📈 Statistiques Techniques
- **Services** : 3 microservices (Go + Rust)
- **Langages** : Go 1.21+, Rust 1.70+
- **Base de Données** : PostgreSQL 15+, Redis 7+
- **Communication** : gRPC, WebSocket, REST API
- **Monitoring** : Prometheus, Grafana, Jaeger
- **Event Bus** : NATS
- **Containerisation** : Docker, Docker Compose
- **CI/CD** : GitHub Actions

### 📊 Couverture de Documentation
- [x] Architecture : 85% (en cours d'amélioration)
- [x] API : 70% (en cours d'amélioration)
- [x] Base de Données : 80% (en cours d'amélioration)
- [ ] Services : 60% (à compléter)
- [ ] Tests : 40% (à compléter)
- [ ] Déploiement : 65% (à compléter)
- [ ] Sécurité : 75% (à compléter)

## 🚀 Démarrage Rapide

### Prérequis Système
```bash
# Vérification des prérequis
go version  # >= 1.21
rustc --version  # >= 1.70
docker --version  # >= 20.10
docker-compose --version  # >= 2.0
psql --version  # >= 15.0
redis-server --version  # >= 7.0
```

### Installation et Configuration

#### 1. Cloner le Repository
```bash
git clone https://github.com/okinrev/veza-full-stack.git
cd veza-full-stack
```

#### 2. Configuration des Variables d'Environnement
```bash
# Backend API
cp veza-backend-api/config.example.env veza-backend-api/.env
# Éditer veza-backend-api/.env avec vos valeurs

# Chat Server
cp veza-chat-server/env.example veza-chat-server/.env
# Éditer veza-chat-server/.env avec vos valeurs

# Stream Server
cp veza-stream-server/env.example veza-stream-server/.env
# Éditer veza-stream-server/.env avec vos valeurs
```

#### 3. Démarrage avec Docker Compose
```bash
# Démarrer tous les services
docker-compose up -d

# Vérifier l'état des services
docker-compose ps

# Voir les logs
docker-compose logs -f
```

#### 4. Développement Local
```bash
# Backend API
cd veza-backend-api
go mod download
go run cmd/server/main.go

# Chat Server
cd veza-chat-server
cargo build
cargo run

# Stream Server
cd veza-stream-server
cargo build
cargo run
```

### Accès aux Services
- **Frontend** : http://localhost:3000
- **Backend API** : http://localhost:8080
- **Chat Server** : http://localhost:3001
- **Stream Server** : http://localhost:3002
- **Grafana** : http://localhost:3003
- **Prometheus** : http://localhost:9090

## 🔧 Configuration Avancée

### Variables d'Environnement Critiques

#### Backend API (.env)
```bash
# Base de données
DATABASE_URL=postgresql://user:password@localhost:5432/veza_db
REDIS_URL=redis://localhost:6379

# JWT
JWT_SECRET=your-super-secret-jwt-key
JWT_EXPIRATION=24h

# OAuth
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret

# NATS
NATS_URL=nats://localhost:4222
```

#### Chat Server (.env)
```bash
# Base de données
DATABASE_URL=postgresql://user:password@localhost:5432/veza_db

# WebSocket
WS_PORT=3001
WS_HOST=0.0.0.0

# Modération
MODERATION_ENABLED=true
PROFANITY_FILTER=true
```

#### Stream Server (.env)
```bash
# Base de données
DATABASE_URL=postgresql://user:password@localhost:5432/veza_db

# Streaming
STREAM_PORT=3002
STREAM_HOST=0.0.0.0

# Storage
STORAGE_PATH=/app/storage
MAX_FILE_SIZE=100MB
```

## 📝 Guide de Contribution

### Standards de Code
- **Go** : `gofmt`, `golint`, `go vet`
- **Rust** : `rustfmt`, `clippy`
- **Tests** : Couverture minimale de 80%
- **Documentation** : JSDoc/TSDoc pour toutes les fonctions

### Workflow de Développement
1. **Fork** du repository
2. **Branch** feature : `git checkout -b feature/nom-feature`
3. **Développement** avec tests
4. **Commit** : `git commit -m "feat: description"`
5. **Push** : `git push origin feature/nom-feature`
6. **Pull Request** avec description détaillée

### Tests Obligatoires
```bash
# Backend API
cd veza-backend-api
go test ./... -v -cover

# Chat Server
cd veza-chat-server
cargo test

# Stream Server
cd veza-stream-server
cargo test
```

## 🔗 Liens Utiles

### Documentation
- [API Reference](./api/endpoints-reference.md)
- [Architecture](./architecture/backend-architecture.md)
- [Déploiement](./deployment/deployment-guide.md)
- [Troubleshooting](./guides/troubleshooting.md)

### Développement
- [Guide de Développement](./guides/development-guide.md)
- [Standards de Code](./guides/coding-standards.md)
- [Tests](./testing/test-strategy.md)

### Production
- [Monitoring](./monitoring/)
- [Sécurité](./security/)
- [Performance](./guides/performance.md)

## 🆘 Support et Communauté

### Ressources d'Aide
- **GitHub Issues** : [Signaler un bug](https://github.com/okinrev/veza-full-stack/issues)
- **Discussions** : [Forum communautaire](https://github.com/okinrev/veza-full-stack/discussions)
- **Documentation** : Cette documentation complète
- **Wiki** : [Wiki du projet](https://github.com/okinrev/veza-full-stack/wiki)

### Contact
- **Email** : support@veza.com
- **Discord** : [Serveur Discord](https://discord.gg/veza)
- **Twitter** : [@VezaPlatform](https://twitter.com/VezaPlatform)

---

<div className="alert alert--info">
  <strong>💡 Conseil</strong> : Commencez par la section <a href="./architecture/backend-architecture.md">Architecture</a> pour comprendre l'ensemble du système, puis explorez les services qui vous intéressent.
</div>

<div className="alert alert--warning">
  <strong>⚠️ Important</strong> : Cette documentation est en constante amélioration. Si vous trouvez des erreurs ou des sections manquantes, n'hésitez pas à contribuer !
</div>

---

**Dernière mise à jour** : $(date)
**Version** : 1.0.0
**Maintenu par** : Équipe Veza
**Licence** : MIT 