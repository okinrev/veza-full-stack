# 🏗️ Architecture Système Veza

## Vue d'ensemble

L'architecture Veza est conçue pour offrir une plateforme de chat et streaming audio haute performance, scalable et sécurisée. Elle suit les principes de l'architecture hexagonale et utilise des microservices spécialisés.

## 🎯 Objectifs d'Architecture

### Performance
- **Latence < 50ms** pour les messages de chat
- **Latence < 100ms** pour le streaming audio
- **Throughput > 10,000** connexions simultanées
- **Uptime > 99.9%**

### Scalabilité
- **Horizontal scaling** automatique
- **Load balancing** intelligent
- **Database sharding** pour les gros volumes
- **Cache distribué** multi-niveaux

### Sécurité
- **Authentification JWT** avec rotation
- **Chiffrement end-to-end** pour les messages
- **Rate limiting** par utilisateur et IP
- **Audit logging** complet

## 🏛️ Architecture Globale

```mermaid
graph TB
    subgraph "Frontend Layer"
        Web[🌐 Web App - React/Next.js]
        Mobile[📱 Mobile App - React Native]
        Desktop[💻 Desktop App - Electron]
    end
    
    subgraph "API Gateway Layer"
        Gateway[🚪 API Gateway - Kong/Traefik]
        LB[⚖️ Load Balancer - HAProxy]
        CDN[🌍 CDN - Cloudflare]
    end
    
    subgraph "Backend Services"
        GoAPI[🔧 Go Backend API]
        ChatServer[💬 Rust Chat Server]
        StreamServer[🎵 Rust Stream Server]
        AuthService[🔐 Auth Service]
    end
    
    subgraph "Data Layer"
        PostgreSQL[(🗄️ PostgreSQL)]
        Redis[(⚡ Redis)]
        NATS[📨 NATS Message Queue]
        MinIO[(📦 MinIO S3)]
    end
    
    subgraph "Monitoring & Observability"
        Prometheus[📊 Prometheus]
        Grafana[📈 Grafana]
        Jaeger[🔍 Jaeger Tracing]
        ELK[📝 ELK Stack]
    end
    
    Web --> CDN
    Mobile --> CDN
    Desktop --> CDN
    CDN --> Gateway
    Gateway --> LB
    LB --> GoAPI
    LB --> ChatServer
    LB --> StreamServer
    LB --> AuthService
    
    GoAPI --> PostgreSQL
    GoAPI --> Redis
    ChatServer --> PostgreSQL
    ChatServer --> Redis
    StreamServer --> PostgreSQL
    StreamServer --> Redis
    
    GoAPI --> NATS
    ChatServer --> NATS
    StreamServer --> NATS
    
    GoAPI --> MinIO
    ChatServer --> MinIO
    StreamServer --> MinIO
    
    GoAPI --> Prometheus
    ChatServer --> Prometheus
    StreamServer --> Prometheus
    Prometheus --> Grafana
    Jaeger --> Grafana
    ELK --> Grafana
```

## 🔧 Services Backend

### 1. Go Backend API (`veza-backend-api`)

**Responsabilités :**
- Authentification et autorisation
- Gestion des utilisateurs et profils
- CRUD des ressources (salles, messages, fichiers)
- Analytics et reporting
- Administration

**Technologies :**
- **Framework** : Gin (HTTP router)
- **ORM** : SQLx (PostgreSQL)
- **Cache** : Redis
- **Auth** : JWT + OAuth2
- **Monitoring** : Prometheus + Zap

**Architecture interne :**
```
internal/
├── api/           # Handlers HTTP
├── core/          # Domain logic
├── adapters/      # Infrastructure adapters
├── middleware/    # HTTP middleware
├── config/        # Configuration
└── monitoring/    # Observability
```

### 2. Rust Chat Server (`veza-chat-server`)

**Responsabilités :**
- WebSocket connections
- Messages en temps réel
- Salles de chat
- Modération automatique
- Présence utilisateurs

**Technologies :**
- **WebSocket** : Tokio-tungstenite
- **Database** : SQLx (PostgreSQL)
- **Cache** : Redis
- **Serialization** : Serde
- **Async Runtime** : Tokio

**Architecture interne :**
```
src/
├── core/          # Domain logic
├── hub/           # WebSocket hub
├── auth/          # Authentication
├── moderation/    # Content moderation
└── monitoring/    # Metrics & tracing
```

### 3. Rust Stream Server (`veza-stream-server`)

**Responsabilités :**
- Streaming audio en temps réel
- Codecs multiples (MP3, AAC, Opus)
- Gestion des sessions de streaming
- Métadonnées audio
- Recording et playback

**Technologies :**
- **Audio** : Symphonia, Hound
- **Codecs** : MinimP3, Rubato
- **WebSocket** : Axum + Tokio-tungstenite
- **Database** : SQLx (PostgreSQL)
- **Cache** : Redis

**Architecture interne :**
```
src/
├── streaming/     # Audio streaming
├── audio/         # Audio processing
├── codecs/        # Codec support
├── analytics/     # Stream analytics
└── monitoring/    # Performance metrics
```

## 🗄️ Couche Données

### PostgreSQL
- **Version** : 15+
- **Extensions** : pg_stat_statements, pg_trgm
- **Optimisations** :
  - Connection pooling (min: 10, max: 100)
  - Read replicas pour les analytics
  - Partitioning pour les gros volumes

### Redis
- **Version** : 7+
- **Usage** :
  - Sessions utilisateurs
  - Cache de données fréquentes
  - Rate limiting
  - Pub/Sub pour les événements

### NATS
- **Version** : 2.9+
- **Usage** :
  - Communication inter-services
  - Event sourcing
  - Notifications en temps réel

## 🔐 Sécurité

### Authentification
- **JWT** avec rotation automatique (1h)
- **Refresh tokens** (7 jours)
- **OAuth2** (Google, GitHub)
- **2FA** avec TOTP

### Autorisation
- **RBAC** (Role-Based Access Control)
- **Permissions granulaires**
- **Audit logging** complet

### Protection
- **Rate limiting** : 100 req/min par IP
- **CORS** configuré strictement
- **CSRF** protection
- **XSS** prevention
- **SQL injection** protection

## 📊 Monitoring & Observabilité

### Métriques
- **Prometheus** : Métriques système et business
- **Grafana** : Dashboards et alerting
- **Custom metrics** : Messages/min, streams actifs

### Logging
- **Structured logging** (JSON)
- **Log levels** : DEBUG, INFO, WARN, ERROR
- **Centralized logging** avec ELK

### Tracing
- **Jaeger** : Distributed tracing
- **OpenTelemetry** : Standard de tracing
- **Performance profiling** : CPU, mémoire

## 🚀 Déploiement

### Environnements
- **Development** : Docker Compose
- **Staging** : Kubernetes
- **Production** : Kubernetes + Helm

### CI/CD
- **GitHub Actions** : Build et tests
- **Docker** : Containerization
- **Helm** : Kubernetes deployment
- **ArgoCD** : GitOps

### Scaling
- **Horizontal Pod Autoscaler** (HPA)
- **Vertical Pod Autoscaler** (VPA)
- **Database scaling** : Read replicas
- **Cache scaling** : Redis Cluster

## 🔄 Communication Inter-Services

### gRPC
- **Protocol Buffers** pour la sérialisation
- **Bidirectional streaming**
- **Service discovery** automatique

### WebSocket
- **Real-time communication**
- **Binary messages** pour l'audio
- **Connection pooling**

### Message Queue
- **NATS** pour les événements
- **Reliability** : At-least-once delivery
- **Ordering** : FIFO par partition

## 📈 Performance

### Benchmarks
- **Chat** : 10,000+ messages/sec
- **Streaming** : 1,000+ concurrent streams
- **Database** : 100,000+ queries/sec
- **Cache** : 1M+ operations/sec

### Optimisations
- **Connection pooling**
- **Query optimization**
- **Index optimization**
- **Caching strategies**
- **Compression** (gzip, brotli)

## 🔧 Configuration

### Variables d'environnement
```bash
# Database
DATABASE_URL=postgresql://user:pass@host:5432/veza
REDIS_URL=redis://localhost:6379

# Security
JWT_SECRET=your-secret-key
ENCRYPTION_KEY=your-encryption-key

# Services
CHAT_SERVER_URL=ws://localhost:8081
STREAM_SERVER_URL=ws://localhost:8082

# Monitoring
PROMETHEUS_PORT=9090
JAEGER_ENDPOINT=http://localhost:14268
```

### Configuration par service
- **Go API** : `config/config.go`
- **Chat Server** : `config/config.rs`
- **Stream Server** : `config/config.rs`

## 🛠️ Développement

### Prérequis
- **Go** : 1.23+
- **Rust** : 1.70+
- **PostgreSQL** : 15+
- **Redis** : 7+
- **Docker** : 20.10+

### Setup local
```bash
# Cloner le repository
git clone <repo-url>
cd veza-full-stack

# Installer les dépendances
make install-deps

# Configurer l'environnement
cp .env.example .env
# Éditer .env

# Lancer les services
make up

# Générer la documentation
make docs-all
```

## 📚 Documentation

### Génération automatique
- **Go** : `godoc`, `swag`
- **Rust** : `cargo doc`
- **API** : OpenAPI 3.0, AsyncAPI
- **Architecture** : Mermaid, PlantUML

### Commandes utiles
```bash
# Documentation complète
make docs-all

# Documentation Go
make docs-go

# Documentation Rust
make docs-rust

# Diagrammes
make docs-diagrams

# Serveur de documentation
make docs-serve
```

---

*Dernière mise à jour : 2024-01-01*
*Version : 0.2.0* 