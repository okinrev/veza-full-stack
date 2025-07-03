# 🔧 Backend API Go - Vue d'ensemble

## 🎯 Objectif

Le Backend API Go est le cœur de l'architecture Veza. Il gère l'authentification, les utilisateurs, les ressources et fournit une API REST complète pour le frontend et les autres services.

## 🏗️ Architecture

### Structure du Projet
```
veza-backend-api/
├── cmd/                    # Points d'entrée
│   ├── server/            # Serveur principal
│   └── production-server/ # Serveur de production
├── internal/              # Code interne
│   ├── api/              # Handlers HTTP
│   ├── core/             # Logique métier
│   ├── adapters/         # Adaptateurs infrastructure
│   ├── middleware/       # Middleware HTTP
│   ├── config/           # Configuration
│   └── monitoring/       # Observabilité
├── proto/                # Définitions protobuf
├── migrations/           # Migrations base de données
└── pkg/                  # Packages publics
```

### Pattern Architecture Hexagonale

```mermaid
graph TB
    subgraph "API Layer"
        HTTP[HTTP Handlers]
        gRPC[gRPC Services]
    end
    
    subgraph "Core Domain"
        Services[Services]
        Entities[Entities]
        Repositories[Repository Interfaces]
    end
    
    subgraph "Infrastructure"
        DB[(PostgreSQL)]
        Cache[(Redis)]
        Queue[NATS]
        Auth[JWT/OAuth]
    end
    
    subgraph "Adapters"
        PostgresAdapter[PostgreSQL Adapter]
        RedisAdapter[Redis Adapter]
        NATSAdapter[NATS Adapter]
    end
    
    HTTP --> Services
    gRPC --> Services
    Services --> Entities
    Services --> Repositories
    Repositories --> PostgresAdapter
    PostgresAdapter --> DB
    RedisAdapter --> Cache
    NATSAdapter --> Queue
```

## 🔧 Technologies Utilisées

### Framework & Runtime
- **Gin** : Framework HTTP haute performance
- **Go 1.23+** : Runtime moderne avec garbage collector optimisé
- **Context** : Gestion des timeouts et annulation

### Base de Données
- **PostgreSQL 15+** : Base de données principale
- **SQLx** : ORM léger avec support des requêtes préparées
- **Connection Pooling** : Pool optimisé (min: 10, max: 100)

### Cache & Session
- **Redis 7+** : Cache distribué et sessions
- **JWT** : Tokens d'authentification avec rotation
- **OAuth2** : Intégration Google, GitHub

### Communication
- **gRPC** : Communication inter-services
- **NATS** : Message queue pour événements
- **Protocol Buffers** : Sérialisation efficace

### Monitoring & Observabilité
- **Prometheus** : Métriques système et business
- **Zap** : Logging structuré JSON
- **Jaeger** : Distributed tracing
- **Health Checks** : Endpoints de santé

## 📊 Métriques Clés

### Performance
- **Latence** : < 50ms (p95)
- **Throughput** : > 10,000 req/sec
- **Connexions** : > 1,000 simultanées
- **Uptime** : > 99.9%

### Métriques Business
- **Utilisateurs actifs** : /min, /heure, /jour
- **Messages envoyés** : /min, /heure
- **Streams actifs** : /min
- **Erreurs** : taux par endpoint

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
- **Rate limiting** : 1000 req/min par utilisateur
- **CORS** configuré strictement
- **CSRF** protection
- **XSS** prevention
- **SQL injection** protection

## 🚀 Déploiement

### Environnements
- **Development** : Docker Compose
- **Staging** : Kubernetes
- **Production** : Kubernetes + Helm

### Configuration
```bash
# Variables d'environnement principales
DATABASE_URL=postgresql://user:pass@host:5432/veza
REDIS_URL=redis://localhost:6379
JWT_SECRET=your-secret-key
NATS_URL=nats://localhost:4222
```

### Health Checks
- `GET /health` : Santé générale
- `GET /ready` : Readiness check
- `GET /metrics` : Métriques Prometheus

## 📚 API Documentation

### Endpoints Principaux

#### Authentication
```
POST /api/v1/auth/login
POST /api/v1/auth/register
POST /api/v1/auth/refresh
POST /api/v1/auth/logout
```

#### Users
```
GET    /api/v1/users
GET    /api/v1/users/{id}
PUT    /api/v1/users/{id}
DELETE /api/v1/users/{id}
```

#### Chat
```
GET    /api/v1/rooms
POST   /api/v1/rooms
GET    /api/v1/rooms/{id}
PUT    /api/v1/rooms/{id}
DELETE /api/v1/rooms/{id}
```

#### Stream
```
GET    /api/v1/streams
POST   /api/v1/streams
GET    /api/v1/streams/{id}
PUT    /api/v1/streams/{id}
DELETE /api/v1/streams/{id}
```

## 🔄 Communication Inter-Services

### gRPC Services
- **Auth Service** : Validation des tokens
- **Chat Service** : Gestion des salles
- **Stream Service** : Gestion des streams

### Événements NATS
- `user.created`
- `user.updated`
- `room.created`
- `message.sent`
- `stream.started`
- `stream.stopped`

## 🧪 Tests

### Types de Tests
- **Unit Tests** : Logique métier
- **Integration Tests** : Base de données, Redis
- **E2E Tests** : Scénarios complets
- **Load Tests** : Performance

### Coverage
- **Minimum** : 80%
- **Critique** : 90% (auth, payments)

### Commandes
```bash
# Tests unitaires
go test ./...

# Tests avec coverage
go test -cover ./...

# Tests d'intégration
make test-integration

# Tests de charge
make test-load
```

## 📈 Monitoring & Alerting

### Dashboards Grafana
- **API Performance** : Latence, throughput, erreurs
- **Business Metrics** : Utilisateurs, messages, streams
- **Infrastructure** : CPU, mémoire, disque, réseau

### Alertes
- **High Error Rate** : > 5% erreurs 5xx
- **High Latency** : > 100ms p95
- **Database Issues** : Connexions, requêtes lentes
- **Memory Usage** : > 80% utilisation

## 🔧 Développement

### Prérequis
- Go 1.23+
- PostgreSQL 15+
- Redis 7+
- NATS 2.9+

### Setup Local
```bash
# Cloner et configurer
git clone <repo>
cd veza-backend-api

# Installer les dépendances
go mod download

# Configurer l'environnement
cp .env.example .env
# Éditer .env

# Lancer les services
docker-compose up -d

# Lancer l'API
go run cmd/server/main.go
```

### Hot Reload
```bash
# Avec air (recommandé)
air

# Avec reflex
reflex -r '\.go$$' -s go run cmd/server/main.go
```

## 📝 Logs

### Format
```json
{
  "level": "info",
  "timestamp": "2024-01-01T00:00:00Z",
  "message": "User authenticated",
  "user_id": "123",
  "ip": "192.168.1.1",
  "trace_id": "abc123"
}
```

### Niveaux
- **DEBUG** : Informations détaillées
- **INFO** : Événements normaux
- **WARN** : Situations anormales
- **ERROR** : Erreurs récupérables
- **FATAL** : Erreurs critiques

---

*Dernière mise à jour : 2024-01-01*
*Version : 0.2.0* 