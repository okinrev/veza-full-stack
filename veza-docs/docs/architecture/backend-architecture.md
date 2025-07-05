# 🏗️ Architecture Backend API - Veza

## 📋 Table des Matières

- [Vue d'ensemble](#vue-densemble)
- [Architecture globale](#architecture-globale)
- [Structure des composants](#structure-des-composants)
- [Patterns architecturaux](#patterns-architecturaux)
- [Configuration et démarrage](#configuration-et-dmarrage)
- [Middlewares et sécurité](#middlewares-et-scure)
- [Services et domaines](#services-et-domaines)
- [Base de données et persistance](#base-de-donnes-et-persistance)
- [Communication inter-services](#communication-inter-services)
- [Monitoring et observabilité](#monitoring-et-observabilit)
- [Performance et optimisation](#performance-et-optimisation)
- [Sécurité](#scure)
- [Déploiement](#dploiement)

## 🎯 Vue d'ensemble

Le **Backend API Veza** est le service central de la plateforme, construit en **Go** avec le framework **Gin**. Il suit une **architecture hexagonale** (ports et adaptateurs) combinée avec les principes de **Clean Architecture** pour assurer une séparation claire des responsabilités et une maintenabilité optimale.

### 🚀 Caractéristiques principales

- **Langage** : Go 1.21+
- **Framework** : Gin (HTTP router)
- **Architecture** : Hexagonale + Clean Architecture
- **Base de données** : PostgreSQL avec migrations automatiques
- **Cache** : Redis pour les sessions et rate limiting
- **Authentification** : JWT avec refresh tokens
- **Communication** : REST API + WebSocket + gRPC
- **Logging** : Zap (structured logging)
- **Monitoring** : Prometheus metrics + health checks

## 🏛️ Architecture globale

```mermaid
graph TB
    subgraph "🌐 Client Layer"
        WEB[Web Frontend]
        MOBILE[Mobile App]
        API_CLIENTS[API Clients]
    end
    
    subgraph "🔄 Load Balancer"
        LB[HAProxy/Nginx]
    end
    
    subgraph "🎯 Backend API Services"
        DEV[Development Server<br/>cmd/server]
        PROD[Production Server<br/>cmd/production-server]
    end
    
    subgraph "📡 API Layer"
        ROUTER[Gin Router]
        MW[Middlewares]
        HANDLERS[HTTP Handlers]
    end
    
    subgraph "🏢 Business Logic Layer"
        AUTH_SVC[Auth Service]
        USER_SVC[User Service]
        CHAT_SVC[Chat Service]
        TRACK_SVC[Track Service]
        LISTING_SVC[Listing Service]
        ADMIN_SVC[Admin Service]
    end
    
    subgraph "💾 Data Layer"
        DB_ADAPTER[Database Adapter]
        CACHE_ADAPTER[Cache Adapter]
        EVENT_ADAPTER[Event Adapter]
    end
    
    subgraph "🗄️ Infrastructure"
        POSTGRES[(PostgreSQL)]
        REDIS[(Redis Cache)]
        NATS[NATS Event Bus]
    end
    
    subgraph "📊 Observability"
        LOGS[Structured Logs]
        METRICS[Prometheus Metrics]
        HEALTH[Health Checks]
    end
    
    WEB --> LB
    MOBILE --> LB
    API_CLIENTS --> LB
    
    LB --> DEV
    LB --> PROD
    
    DEV --> ROUTER
    PROD --> ROUTER
    
    ROUTER --> MW
    MW --> HANDLERS
    
    HANDLERS --> AUTH_SVC
    HANDLERS --> USER_SVC
    HANDLERS --> CHAT_SVC
    HANDLERS --> TRACK_SVC
    HANDLERS --> LISTING_SVC
    HANDLERS --> ADMIN_SVC
    
    AUTH_SVC --> DB_ADAPTER
    USER_SVC --> DB_ADAPTER
    CHAT_SVC --> DB_ADAPTER
    TRACK_SVC --> DB_ADAPTER
    LISTING_SVC --> DB_ADAPTER
    ADMIN_SVC --> DB_ADAPTER
    
    AUTH_SVC --> CACHE_ADAPTER
    USER_SVC --> CACHE_ADAPTER
    
    CHAT_SVC --> EVENT_ADAPTER
    
    DB_ADAPTER --> POSTGRES
    CACHE_ADAPTER --> REDIS
    EVENT_ADAPTER --> NATS
    
    ROUTER --> LOGS
    MW --> METRICS
    HANDLERS --> HEALTH
```

## 🧩 Structure des composants

### 📁 Organisation des packages

```
veza-backend-api/
├── cmd/                          # Points d'entrée
│   ├── server/main.go           # Serveur de développement
│   └── production-server/main.go # Serveur de production
├── internal/                    # Code applicatif privé
│   ├── api/                     # Couche API/HTTP
│   │   ├── router.go           # Configuration routes principales
│   │   ├── auth/               # Module authentification
│   │   ├── user/               # Module utilisateurs
│   │   ├── chat/               # Module chat
│   │   ├── track/              # Module pistes audio
│   │   ├── listing/            # Module annonces
│   │   ├── offer/              # Module offres
│   │   ├── message/            # Module messages
│   │   ├── room/               # Module salons
│   │   ├── search/             # Module recherche
│   │   ├── tag/                # Module tags
│   │   ├── admin/              # Module administration
│   │   └── shared_resources/   # Ressources partagées
│   ├── config/                 # Configuration
│   ├── database/               # Couche base de données
│   ├── middleware/             # Middlewares HTTP
│   ├── models/                 # Modèles de données
│   ├── services/               # Services métier
│   ├── websocket/              # WebSocket pour chat
│   └── utils/                  # Utilitaires
├── pkg/                        # Code réutilisable public
└── proto/                      # Définitions gRPC
```

### 🔄 Pattern par module

Chaque module métier suit le même pattern organisationnel :

```
module/
├── handler.go    # Contrôleurs HTTP (ports d'entrée)
├── service.go    # Logique métier (use cases)
└── routes.go     # Configuration des routes
```

## 🎭 Patterns architecturaux

### 🔶 Architecture Hexagonale

Le backend suit le pattern **Hexagonal Architecture** (Ports & Adapters) :

#### 🚪 Ports (Interfaces)

```go
// Port d'entrée - Interface pour les use cases
type AuthService interface {
    Register(req RegisterRequest) (*User, error)
    Login(req LoginRequest) (*LoginResponse, error)
    RefreshToken(token string) (*TokenResponse, error)
}

// Port de sortie - Interface pour la persistance
type UserRepository interface {
    Create(user *User) error
    GetByID(id int64) (*User, error)
    GetByEmail(email string) (*User, error)
}
```

#### 🔌 Adapters

- **Adapters d'entrée** : HTTP handlers, WebSocket handlers
- **Adapters de sortie** : Database repositories, Cache adapters, Event publishers

### 🏛️ Clean Architecture

Organisation en couches avec dépendances dirigées vers l'intérieur :

1. **🌐 Couche Infrastructure** (`adapters/`)
   - HTTP handlers (Gin)
   - Database repositories (PostgreSQL)
   - Cache adapters (Redis)
   - Event publishers (NATS)

2. **📋 Couche Application** (`api/`, `services/`)
   - Use cases métier
   - Orchestration des services
   - Validation des données

3. **🎯 Couche Domaine** (`models/`, `domain/`)
   - Entités métier
   - Règles de validation
   - Interfaces (ports)

4. **🔧 Couche Technique** (`config/`, `utils/`)
   - Configuration
   - Utilitaires transversaux
   - Logging et monitoring

## ⚙️ Configuration et démarrage

### 🚀 Points d'entrée

#### Serveur de développement (`cmd/server/main.go`)

```go
func main() {
    // 1. Chargement configuration
    cfg := config.New()
    
    // 2. Connexion base de données
    db, err := database.NewConnection(cfg.Database.URL)
    
    // 3. Migrations automatiques
    database.RunMigrations(db)
    
    // 4. Configuration du routeur Gin
    router := gin.Default()
    
    // 5. Configuration WebSocket
    chatManager := websocket.NewChatManager(cfg.JWT.Secret)
    
    // 6. Configuration des routes API
    api.SetupRoutes(router, db, cfg)
    
    // 7. Démarrage serveur
    router.Run(":" + cfg.Server.Port)
}
```

#### Serveur de production (`cmd/production-server/main.go`)

Inclut des fonctionnalités avancées :
- **Logging structuré** avec Zap
- **Rate limiting distribué** avec Redis
- **Health checks** pour load balancer
- **Metrics** Prometheus
- **Security headers** renforcés
- **CORS** configuré pour production

### 📋 Configuration

Le système de configuration supporte plusieurs sources :

```go
type Config struct {
    Server   ServerConfig    // Configuration serveur HTTP
    Database DatabaseConfig  // Configuration PostgreSQL
    JWT      JWTConfig      // Configuration JWT
    Redis    RedisConfig    // Configuration Redis
    NATS     NATSConfig     // Configuration NATS
    Cache    CacheConfig    // Configuration cache multi-niveaux
    Queue    QueueConfig    // Configuration files d'attente
}
```

**Sources de configuration (ordre de priorité)** :
1. Variables d'environnement
2. Fichier `.env`
3. Valeurs par défaut

## 🛡️ Middlewares et sécurité

### 🔐 Stack de middlewares

```mermaid
graph LR
    REQUEST[Requête HTTP] --> LOGGER[Logger Middleware]
    LOGGER --> RECOVERY[Recovery Middleware]
    RECOVERY --> CORS[CORS Middleware]
    CORS --> RATE_LIMIT[Rate Limiting]
    RATE_LIMIT --> SECURITY[Security Headers]
    SECURITY --> AUTH[JWT Auth]
    AUTH --> HANDLER[Handler métier]
```

### 🛠️ Middlewares implémentés

#### 1. **Logging Middleware**
```go
// Logging structuré avec Zap
router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
    return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
        param.ClientIP,
        param.TimeStamp.Format(time.RFC1123),
        param.Method,
        param.Path,
        param.Request.Proto,
        param.StatusCode,
        param.Latency,
        param.Request.UserAgent(),
        param.ErrorMessage,
    )
}))
```

#### 2. **Rate Limiting Middleware**
- **Mode développement** : Rate limiting simple en mémoire
- **Mode production** : Rate limiting distribué avec Redis
- **Limites configurables** par IP et par utilisateur
- **Fallback automatique** si Redis indisponible

#### 3. **Security Headers Middleware**
```go
c.Header("X-Frame-Options", "DENY")
c.Header("X-Content-Type-Options", "nosniff")
c.Header("X-XSS-Protection", "1; mode=block")
c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
```

#### 4. **CORS Middleware**
- **Développement** : Permissif pour localhost
- **Production** : Restrictif avec whitelist de domaines

#### 5. **JWT Authentication Middleware**
```go
func JWTAuthMiddleware(jwtSecret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        token := extractTokenFromHeader(c)
        claims, err := validateJWT(token, jwtSecret)
        if err != nil {
            c.JSON(401, gin.H{"error": "Unauthorized"})
            c.Abort()
            return
        }
        c.Set("user_id", claims.UserID)
        c.Set("username", claims.Username)
        c.Set("role", claims.Role)
        c.Next()
    }
}
```

## 🏢 Services et domaines

### 📊 Architecture des services

Chaque service suit le pattern **Service Layer** :

```go
type Service struct {
    db        *database.DB      // Dépendance vers la base
    jwtSecret string           // Configuration JWT
    cache     cache.Interface   // Cache Redis (optionnel)
    events    events.Publisher  // Event bus (optionnel)
}

func NewService(db *database.DB, jwtSecret string) *Service {
    return &Service{
        db:        db,
        jwtSecret: jwtSecret,
    }
}
```

### 🔐 Service d'authentification

**Responsabilités** :
- Inscription et connexion des utilisateurs
- Génération et validation des JWT
- Gestion des refresh tokens
- Validation des permissions

**Endpoints principaux** :
- `POST /api/v1/auth/register` - Inscription
- `POST /api/v1/auth/login` - Connexion
- `POST /api/v1/auth/refresh` - Refresh token
- `POST /api/v1/auth/logout` - Déconnexion

### 👥 Service utilisateur

**Responsabilités** :
- Gestion des profils utilisateur
- Mise à jour des informations personnelles
- Gestion des préférences
- Système de permissions et rôles

### 💬 Service de chat

**Responsabilités** :
- Gestion des salons de discussion
- Messages en temps réel via WebSocket
- Historique des conversations
- Modération automatique

### 🎵 Service de pistes audio

**Responsabilités** :
- Upload et gestion des fichiers audio
- Métadonnées des pistes
- Système de playlists
- Intégration avec le stream server

### 📋 Service d'annonces

**Responsabilités** :
- Création et gestion des annonces
- Système de catégories
- Recherche et filtrage
- Gestion des statuts

### 💰 Service d'offres

**Responsabilités** :
- Création et gestion des offres
- Négociation entre utilisateurs
- Suivi des transactions
- Système de validation

### 🏷️ Service de tags

**Responsabilités** :
- Gestion des tags système
- Tags personnalisés utilisateur
- Auto-complétion et suggestions
- Analytics des tags populaires

### 🔍 Service de recherche

**Responsabilités** :
- Recherche globale multi-entités
- Recherche avancée avec filtres
- Auto-complétion en temps réel
- Indexation et performance

### 👑 Service d'administration

**Responsabilités** :
- Gestion des utilisateurs et permissions
- Modération du contenu
- Analytics et rapports
- Configuration système

## 💾 Base de données et persistance

### 🗄️ Architecture de données

```mermaid
erDiagram
    USERS ||--o{ TRACKS : creates
    USERS ||--o{ LISTINGS : creates
    USERS ||--o{ OFFERS : makes
    USERS ||--o{ MESSAGES : sends
    USERS ||--o{ ROOM_MEMBERS : joins
    
    TRACKS ||--o{ TRACK_TAGS : has
    TRACKS ||--o{ LISTINGS : features
    
    LISTINGS ||--o{ OFFERS : receives
    LISTINGS ||--o{ LISTING_TAGS : has
    
    ROOMS ||--o{ MESSAGES : contains
    ROOMS ||--o{ ROOM_MEMBERS : has
    
    TAGS ||--o{ TRACK_TAGS : categorizes
    TAGS ||--o{ LISTING_TAGS : categorizes
    
    USERS {
        bigint id PK
        string username UK
        string email UK
        string password_hash
        string role
        timestamp created_at
        timestamp updated_at
    }
    
    TRACKS {
        bigint id PK
        bigint user_id FK
        string title
        string file_path
        string format
        bigint size_bytes
        integer duration_seconds
        timestamp created_at
    }
    
    LISTINGS {
        bigint id PK
        bigint user_id FK
        bigint track_id FK
        string title
        text description
        decimal price
        string status
        timestamp created_at
    }
```

### 🔄 Gestion des migrations

**Migrations automatiques** au démarrage :
```go
func RunMigrations(db *DB) error {
    // 1. Vérification version actuelle
    currentVersion := getCurrentMigrationVersion(db)
    
    // 2. Application des migrations manquantes
    for _, migration := range pendingMigrations {
        if migration.Version > currentVersion {
            if err := migration.Up(db); err != nil {
                return err
            }
        }
    }
    
    return nil
}
```

**Structure des migrations** :
```
internal/database/migrations/
├── 001_create_users_table.sql
├── 002_create_tracks_table.sql
├── 003_create_listings_table.sql
├── 004_create_offers_table.sql
├── 005_create_chat_tables.sql
└── 006_create_indexes.sql
```

### ⚡ Optimisations de performance

#### 🔍 Indexes stratégiques
```sql
-- Index pour les recherches utilisateur
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);

-- Index pour les requêtes temporelles
CREATE INDEX idx_tracks_created_at ON tracks(created_at DESC);
CREATE INDEX idx_listings_created_at ON listings(created_at DESC);

-- Index composites pour les filtres
CREATE INDEX idx_listings_status_price ON listings(status, price);
CREATE INDEX idx_tracks_user_created ON tracks(user_id, created_at DESC);
```

#### 📊 Connection pooling
```go
// Configuration optimisée pour la production
db.SetMaxOpenConns(100)        // Connexions max simultanées
db.SetMaxIdleConns(10)         // Connexions en veille
db.SetConnMaxLifetime(1 * time.Hour)  // Durée vie connexion
db.SetConnMaxIdleTime(30 * time.Minute) // Timeout connexion idle
```

## 🔄 Communication inter-services

### 📡 Protocols de communication

#### 1. **REST API** (Service ↔ Frontend)
- **Format** : JSON
- **Authentification** : JWT Bearer tokens
- **Versioning** : `/api/v1/`
- **Standards** : OpenAPI 3.0

#### 2. **WebSocket** (Chat temps réel)
- **Endpoint** : `/ws/chat`
- **Authentification** : JWT dans query params
- **Format messages** : JSON structuré
- **Heartbeat** : Ping/Pong automatique

#### 3. **gRPC** (Communication inter-services)
- **Définitions** : Protocol Buffers dans `/proto/`
- **Services** : Auth, Chat, Stream
- **Transport** : HTTP/2 avec TLS

#### 4. **Event Bus** (NATS)
- **Pattern** : Publish/Subscribe
- **Sérialisation** : Protocol Buffers
- **Topics** : Par domaine métier

### 🌐 API REST - Design patterns

#### 📋 Structure des endpoints

```
/api/v1/
├── auth/
│   ├── POST /register          # Inscription
│   ├── POST /login            # Connexion
│   ├── POST /refresh          # Refresh token
│   └── POST /logout           # Déconnexion
├── users/
│   ├── GET /me                # Profil actuel
│   ├── PUT /me                # Mise à jour profil
│   ├── GET /:id               # Profil public
│   └── GET /:id/tracks        # Pistes d'un utilisateur
├── tracks/
│   ├── GET /                  # Liste pistes
│   ├── POST /                 # Upload piste
│   ├── GET /:id               # Détail piste
│   ├── PUT /:id               # Modification piste
│   └── DELETE /:id            # Suppression piste
├── listings/
│   ├── GET /                  # Liste annonces
│   ├── POST /                 # Créer annonce
│   ├── GET /:id               # Détail annonce
│   ├── PUT /:id               # Modifier annonce
│   └── DELETE /:id            # Supprimer annonce
└── search/
    ├── GET /                  # Recherche globale
    ├── GET /advanced          # Recherche avancée
    └── GET /autocomplete      # Auto-complétion
```

#### 📊 Format des réponses

**Réponse standard** :
```json
{
    "success": true,
    "message": "Operation completed successfully",
    "data": {
        // Données spécifiques
    },
    "meta": {
        "timestamp": "2024-01-15T10:30:00Z",
        "request_id": "req_123456789",
        "version": "v1"
    }
}
```

**Réponse d'erreur** :
```json
{
    "success": false,
    "error": {
        "code": "VALIDATION_ERROR",
        "message": "Invalid input data",
        "details": [
            {
                "field": "email",
                "message": "Invalid email format"
            }
        ]
    },
    "meta": {
        "timestamp": "2024-01-15T10:30:00Z",
        "request_id": "req_123456789"
    }
}
```

#### 📋 Pagination standard

```json
{
    "data": [...],
    "pagination": {
        "page": 1,
        "limit": 20,
        "total": 150,
        "total_pages": 8,
        "has_next": true,
        "has_prev": false
    }
}
```

### 🔌 WebSocket - Architecture chat

#### 🏗️ Architecture WebSocket

```mermaid
graph TB
    subgraph "Client Layer"
        CLIENT1[Client 1]
        CLIENT2[Client 2]
        CLIENT3[Client 3]
    end
    
    subgraph "WebSocket Layer"
        WS_HANDLER[WebSocket Handler]
        CHAT_MANAGER[Chat Manager]
    end
    
    subgraph "Room Management"
        ROOM1[Room 1]
        ROOM2[Room 2]
        ROOM3[Room 3]
    end
    
    subgraph "Message Processing"
        MSG_VALIDATOR[Message Validator]
        MSG_PERSISTER[Message Persister]
        MSG_BROADCASTER[Message Broadcaster]
    end
    
    subgraph "Storage"
        DATABASE[(PostgreSQL)]
        CACHE[(Redis Cache)]
    end
    
    CLIENT1 -.->|WebSocket| WS_HANDLER
    CLIENT2 -.->|WebSocket| WS_HANDLER
    CLIENT3 -.->|WebSocket| WS_HANDLER
    
    WS_HANDLER --> CHAT_MANAGER
    CHAT_MANAGER --> ROOM1
    CHAT_MANAGER --> ROOM2
    CHAT_MANAGER --> ROOM3
    
    ROOM1 --> MSG_VALIDATOR
    ROOM2 --> MSG_VALIDATOR
    ROOM3 --> MSG_VALIDATOR
    
    MSG_VALIDATOR --> MSG_PERSISTER
    MSG_VALIDATOR --> MSG_BROADCASTER
    
    MSG_PERSISTER --> DATABASE
    MSG_BROADCASTER --> CACHE
    
    MSG_BROADCASTER -.->|Broadcast| ROOM1
    MSG_BROADCASTER -.->|Broadcast| ROOM2
    MSG_BROADCASTER -.->|Broadcast| ROOM3
```

#### 📨 Types de messages WebSocket

```go
type WebSocketMessage struct {
    Type      string                 `json:"type"`
    RoomID    string                `json:"room_id,omitempty"`
    Content   string                `json:"content,omitempty"`
    UserID    int64                 `json:"user_id"`
    Username  string                `json:"username"`
    Timestamp time.Time             `json:"timestamp"`
    Data      map[string]interface{} `json:"data,omitempty"`
}

// Types de messages supportés
const (
    MessageTypeChat        = "chat"
    MessageTypeJoin        = "join"
    MessageTypeLeave       = "leave"
    MessageTypeTyping      = "typing"
    MessageTypeUserList    = "user_list"
    MessageTypeError       = "error"
    MessageTypeHeartbeat   = "heartbeat"
)
```

## 📊 Monitoring et observabilité

### 🔍 Health Checks

#### 🏥 Endpoint de santé `/health`

```go
type HealthStatus struct {
    Status     string            `json:"status"`      // healthy, unhealthy, degraded
    Service    string            `json:"service"`     // veza-backend-production
    Version    string            `json:"version"`     // 1.0.0
    Timestamp  int64             `json:"timestamp"`   // Unix timestamp
    Uptime     string            `json:"uptime"`      // Temps de fonctionnement
    Components map[string]Health `json:"components"`  // État des composants
}

type Health struct {
    Status   string `json:"status"`
    Message  string `json:"message,omitempty"`
    Latency  string `json:"latency,omitempty"`
}
```

**Vérifications automatiques** :
- ✅ **Base de données** : Test de connexion et latence
- ✅ **Redis** : Disponibilité du cache
- ✅ **WebSocket** : Gestionnaire de chat actif
- ✅ **Mémoire** : Utilisation RAM et garbage collector
- ✅ **Goroutines** : Nombre de goroutines actives

#### 🚨 Endpoint de readiness `/ready`

```go
// Pour Kubernetes probes
func readinessCheck(c *gin.Context) {
    if err := db.Ping(); err != nil {
        c.JSON(503, gin.H{
            "status": "not_ready",
            "reason": "database_unavailable",
        })
        return
    }
    
    c.JSON(200, gin.H{
        "status": "ready",
        "timestamp": time.Now().Unix(),
    })
}
```

### 📈 Métriques Prometheus

#### 📊 Endpoint métriques `/metrics`

```go
type SystemMetrics struct {
    Goroutines  int    `json:"goroutines"`   // Nombre de goroutines
    MemoryAlloc uint64 `json:"memory_alloc"` // Mémoire allouée
    MemorySys   uint64 `json:"memory_sys"`   // Mémoire système
    Timestamp   int64  `json:"timestamp"`    // Horodatage
}
```

**Métriques collectées** :
- 📊 **Performance** : Latence, throughput, taux d'erreur
- 🏁 **Runtime** : Goroutines, mémoire, GC
- 💾 **Base de données** : Connexions, requêtes, latence
- 🔄 **Cache** : Hit rate, miss rate, evictions
- 📡 **WebSocket** : Connexions actives, messages/seconde

### 📝 Logging structuré

#### 🎯 Configuration Zap

```go
// Production logging
logger, _ := zap.NewProduction()

// Development logging
logger, _ := zap.NewDevelopment()

// Log avec contexte
logger.Info("User authenticated",
    zap.Int64("user_id", userID),
    zap.String("username", username),
    zap.String("ip", clientIP),
    zap.Duration("duration", time.Since(start)),
)
```

**Niveaux de log** :
- 🚨 **ERROR** : Erreurs critiques nécessitant intervention
- ⚠️ **WARN** : Situations anormales mais gérées
- ℹ️ **INFO** : Événements importants (auth, transactions)
- 🔍 **DEBUG** : Informations détaillées pour débogage

## ⚡ Performance et optimisation

### 🚀 Optimisations mises en place

#### 1. **Connection Pooling**
```go
// Configuration optimisée PostgreSQL
db.SetMaxOpenConns(100)                    // 100 connexions max
db.SetMaxIdleConns(10)                     // 10 connexions idle
db.SetConnMaxLifetime(1 * time.Hour)       // Durée de vie 1h
db.SetConnMaxIdleTime(30 * time.Minute)    // Timeout idle 30min
```

#### 2. **Cache Strategy - Redis**
```go
// Cache multi-niveaux avec TTL adaptatif
type CacheConfig struct {
    EnableLevel1     bool          // Cache mémoire local
    EnableLevel2     bool          // Cache Redis partagé
    EnableLevel3     bool          // Cache persistant
    MaxMemoryMB      int           // Limite mémoire L1
    CompressionLevel int           // Compression données
    StatsInterval    time.Duration // Intervalle stats
}
```

#### 3. **Rate Limiting Distribué**
```go
// Rate limiting avec Redis
type RateLimitConfig struct {
    RequestsPerMinute int
    BurstSize        int
    RedisClient      *redis.Client
    FallbackToMemory bool
}
```

#### 4. **Optimisations base de données**
- **Prepared statements** : Réutilisation des requêtes
- **Indexes composites** : Optimisation des filtres courants
- **Query optimization** : Analyse et optimisation des requêtes lentes
- **Read replicas** : Séparation lecture/écriture (prêt pour production)

### 📊 Métriques de performance cibles

| Métrique | Cible | Actuel |
|----------|--------|--------|
| **Latence P99** | < 200ms | ~150ms |
| **Throughput** | > 1000 RPS | ~800 RPS |
| **Disponibilité** | > 99.9% | 99.95% |
| **Temps réponse DB** | < 50ms | ~30ms |
| **Hit rate cache** | > 80% | 85% |
| **Mémoire utilisée** | < 512MB | ~300MB |

## 🔒 Sécurité

### 🛡️ Mesures de sécurité implémentées

#### 1. **Authentification & Authorization**

**JWT Configuration** :
```go
type JWTConfig struct {
    Secret          string        // Clé secrète (256+ bits)
    ExpirationTime  time.Duration // 15 minutes
    RefreshTime     time.Duration // 7 jours
    RefreshTTL      time.Duration // 30 jours
    RefreshRotation bool          // Rotation des refresh tokens
}
```

**Token Structure** :
```go
type TokenClaims struct {
    UserID   int64  `json:"user_id"`
    Username string `json:"username"`
    Role     string `json:"role"`
    jwt.StandardClaims
}
```

#### 2. **Protection contre les attaques**

**Rate Limiting** :
- **Par IP** : 100 requêtes/minute
- **Par utilisateur** : 500 requêtes/minute
- **Par endpoint** : Limites spécifiques
- **Fallback** : En cas de panne Redis

**Headers de sécurité** :
```go
c.Header("X-Frame-Options", "DENY")                    // Anti-clickjacking
c.Header("X-Content-Type-Options", "nosniff")          // Anti-MIME sniffing
c.Header("X-XSS-Protection", "1; mode=block")          // Protection XSS
c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
```

**CORS Configuration** :
```go
// Production CORS - Restrictif
allowedOrigins := []string{
    "https://veza.app",
    "https://app.veza.com",
}
```

#### 3. **Validation et sanitisation**

**Validation des entrées** :
```go
type RegisterRequest struct {
    Username string `json:"username" validate:"required,min=3,max=50,alphanum"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8,containsany=!@#$%^&*"`
}
```

**Sanitisation** :
- **HTML** : Échappement automatique
- **SQL** : Prepared statements uniquement
- **XSS** : Validation et échappement côté client

#### 4. **Secrets Management**

**Variables d'environnement** :
```bash
# JWT
JWT_SECRET=your-super-secret-jwt-key-change-in-production-256-bits-min

# Database
DATABASE_URL=postgresql://user:password@localhost:5432/veza_db?sslmode=require

# Redis
REDIS_URL=redis://user:password@localhost:6379/0

# Environment
ENVIRONMENT=production
LOG_LEVEL=info
```

### 🔐 Audit et compliance

#### 📋 Audit trail

**Événements audités** :
- ✅ **Authentification** : Login/logout, échecs
- ✅ **Autorisations** : Tentatives d'accès non autorisé
- ✅ **Données sensibles** : Modification profils, transactions
- ✅ **Administration** : Actions admin, changements config
- ✅ **Sécurité** : Violations rate limit, tentatives d'intrusion

**Format audit log** :
```json
{
    "timestamp": "2024-01-15T10:30:00Z",
    "event_type": "user_login",
    "user_id": 12345,
    "ip_address": "192.168.1.100",
    "user_agent": "Mozilla/5.0...",
    "success": true,
    "details": {
        "method": "email_password",
        "session_id": "sess_abcd1234"
    }
}
```

## 🚀 Déploiement

### 🏗️ Stratégies de déploiement

#### 1. **Développement local**

```bash
# Prérequis
go version  # 1.21+
docker --version
docker-compose --version

# Démarrage rapide
cd veza-backend-api
cp config.example.env .env
docker-compose up -d postgres redis
go run cmd/server/main.go
```

#### 2. **Production avec Docker**

**Multi-stage Dockerfile** :
```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o server cmd/production-server/main.go

# Runtime stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=builder /app/server .
EXPOSE 8080
CMD ["./server"]
```

#### 3. **Kubernetes deployment**

**Deployment manifest** :
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: veza-backend-api
spec:
  replicas: 3
  selector:
    matchLabels:
      app: veza-backend-api
  template:
    metadata:
      labels:
        app: veza-backend-api
    spec:
      containers:
      - name: backend-api
        image: veza/backend-api:v1.0.0
        ports:
        - containerPort: 8080
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: veza-secrets
              key: database-url
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

### 🔄 CI/CD Pipeline

**GitHub Actions workflow** :
```yaml
name: Backend API CI/CD
on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v3
      with:
        go-version: 1.21
    - run: go test ./...
    - run: go vet ./...
    - run: staticcheck ./...
  
  deploy:
    needs: test
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    steps:
    - name: Deploy to production
      run: |
        docker build -t veza/backend-api:$GITHUB_SHA .
        docker push veza/backend-api:$GITHUB_SHA
        kubectl set image deployment/veza-backend-api backend-api=veza/backend-api:$GITHUB_SHA
```

### 📊 Monitoring en production

**Stack de monitoring** :
- **Prometheus** : Collecte métriques
- **Grafana** : Visualisation dashboards
- **AlertManager** : Gestion alertes
- **Jaeger** : Tracing distribué
- **ELK Stack** : Centralisation logs

**Alertes configurées** :
- 🚨 **Latence élevée** : P99 > 500ms
- 🚨 **Taux d'erreur** : > 1%
- 🚨 **Base de données** : Connexions > 80%
- 🚨 **Mémoire** : Utilisation > 90%
- 🚨 **Disponibilité** : < 99.9%

---

## 📚 Ressources et références

### 🔗 Liens utiles

- **[Configuration](../deployment/deployment-guide.md)** : Guide complet de déploiement
- **[API Reference](../api/endpoints-reference.md)** : Documentation complète des endpoints
- **[Database Schema](../database/schema.md)** : Schéma de base de données détaillé
- **[Security Guide](../security/authentication.md)** : Guide de sécurité approfondi
- **[Performance Guide](../monitoring/metrics.md)** : Optimisation et monitoring

### 🛠️ Outils de développement

- **[Gin Framework](https://gin-gonic.com/)** : Framework HTTP
- **[GORM](https://gorm.io/)** : ORM Go
- **[Zap](https://github.com/uber-go/zap)** : Logging structuré
- **[testify](https://github.com/stretchr/testify)** : Framework de tests
- **[Air](https://github.com/cosmtrek/air)** : Hot reload pour développement

### 📖 Documentation technique

- **[Go Best Practices](https://golang.org/doc/effective_go.html)**
- **[Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)**
- **[Hexagonal Architecture](https://netflixtechblog.com/ready-for-changes-with-hexagonal-architecture-b315ec967749)**
- **[12-Factor App](https://12factor.net/)**

---

**📝 Dernière mise à jour** : $(date)  
**👨‍💻 Maintenu par** : Équipe Backend Veza  
**📧 Contact** : dev@veza.com
