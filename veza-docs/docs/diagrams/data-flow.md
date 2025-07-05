# Flux de Données - Veza Platform

## Vue d'ensemble des Flux de Données

Ce document décrit les flux de données principaux dans l'architecture Veza, incluant les interactions entre les services, les bases de données, et les clients.

## Diagramme de Flux Principal

```mermaid
flowchart TD
    subgraph "Clients"
        Web[🌐 Web Client]
        Mobile[📱 Mobile Client]
        Desktop[💻 Desktop Client]
    end
    
    subgraph "Load Balancer"
        LB[⚖️ Load Balancer<br/>Nginx/Traefik]
    end
    
    subgraph "Backend Services"
        API[🔧 Backend API<br/>Go/Gin]
        Chat[💬 Chat Server<br/>Rust/Axum]
        Stream[🎵 Stream Server<br/>Rust/Tokio]
    end
    
    subgraph "Data Layer"
        PostgreSQL[(🗄️ PostgreSQL<br/>Primary Database)]
        Redis[(⚡ Redis<br/>Cache & Sessions)]
        S3[(☁️ S3 Storage<br/>Media Files)]
    end
    
    subgraph "Message Queue"
        NATS[🔄 NATS<br/>Event Bus]
        RedisQ[📨 Redis Queue<br/>Background Jobs]
    end
    
    subgraph "External Services"
        Auth[🔐 Auth Service<br/>JWT/OAuth2]
        Email[📧 Email Service<br/>SMTP/SendGrid]
        SMS[📱 SMS Service<br/>Twilio]
        CDN[🌍 CDN<br/>CloudFront]
    end
    
    %% Client to Load Balancer
    Web --> LB
    Mobile --> LB
    Desktop --> LB
    
    %% Load Balancer to Services
    LB --> API
    LB --> Chat
    LB --> Stream
    
    %% API Data Flow
    API --> PostgreSQL
    API --> Redis
    API --> NATS
    API --> Auth
    
    %% Chat Server Data Flow
    Chat --> PostgreSQL
    Chat --> Redis
    Chat --> NATS
    
    %% Stream Server Data Flow
    Stream --> S3
    Stream --> Redis
    Stream --> NATS
    Stream --> CDN
    
    %% Message Queue Flow
    NATS --> Email
    NATS --> SMS
    RedisQ --> API
    
    %% External Service Flow
    Auth --> API
    CDN --> Stream
```

## Flux d'Authentification

```mermaid
sequenceDiagram
    participant Client
    participant API
    participant Auth
    participant DB
    participant Redis
    participant NATS
    
    Client->>API: POST /auth/login
    API->>DB: Validate User Credentials
    DB->>API: User Data
    API->>Auth: Generate JWT Token
    Auth->>API: JWT Token
    API->>Redis: Store Session
    API->>NATS: Publish Login Event
    API->>Client: JWT Token + User Data
    
    Client->>API: GET /api/protected (with JWT)
    API->>Auth: Validate JWT Token
    Auth->>API: Token Valid
    API->>Redis: Check Session
    API->>Client: Protected Data
```

## Flux de Chat en Temps Réel

```mermaid
sequenceDiagram
    participant Client1
    participant Client2
    participant ChatServer
    participant DB
    participant Redis
    participant NATS
    
    Client1->>ChatServer: WebSocket Connect
    ChatServer->>Redis: Store Connection
    ChatServer->>NATS: Publish User Online
    
    Client2->>ChatServer: WebSocket Connect
    ChatServer->>Redis: Store Connection
    ChatServer->>NATS: Publish User Online
    
    Client1->>ChatServer: Send Message
    ChatServer->>DB: Store Message
    ChatServer->>Redis: Cache Message
    ChatServer->>NATS: Publish Message Event
    ChatServer->>Client2: Broadcast Message
    ChatServer->>Client1: Message Confirmation
    
    Client2->>ChatServer: Mark as Read
    ChatServer->>DB: Update Read Status
    ChatServer->>NATS: Publish Read Event
    ChatServer->>Client1: Read Receipt
```

## Flux de Streaming Audio

```mermaid
sequenceDiagram
    participant Broadcaster
    participant StreamServer
    participant S3
    participant CDN
    participant NATS
    participant Viewer
    
    Broadcaster->>StreamServer: Start Stream
    StreamServer->>S3: Create Stream Record
    StreamServer->>NATS: Publish Stream Start
    StreamServer->>Broadcaster: Stream URL
    
    Broadcaster->>StreamServer: Audio Data (WebSocket)
    StreamServer->>StreamServer: Process Audio
    StreamServer->>S3: Store Processed Audio
    StreamServer->>CDN: Upload to CDN
    StreamServer->>NATS: Publish Audio Event
    
    Viewer->>StreamServer: Join Stream
    StreamServer->>CDN: Get Stream URL
    StreamServer->>Viewer: Stream URL
    CDN->>Viewer: Audio Stream
```

## Flux de Notifications

```mermaid
sequenceDiagram
    participant API
    participant NATS
    participant EmailService
    participant SMSService
    participant PushService
    participant User
    
    API->>NATS: Publish Notification Event
    NATS->>EmailService: Email Notification
    NATS->>SMSService: SMS Notification
    NATS->>PushService: Push Notification
    
    EmailService->>User: Email
    SMSService->>User: SMS
    PushService->>User: Push Notification
```

## Flux de Cache

```mermaid
flowchart LR
    subgraph "Cache Levels"
        L1[L1 Cache<br/>Memory]
        L2[L2 Cache<br/>Redis]
        L3[L3 Cache<br/>Database]
    end
    
    subgraph "Cache Operations"
        Read[📖 Read Request]
        Write[✏️ Write Request]
        Invalidate[🗑️ Invalidation]
    end
    
    Read --> L1
    L1 -->|Cache Miss| L2
    L2 -->|Cache Miss| L3
    L3 -->|Data Found| L2
    L2 -->|Store| L1
    L1 -->|Return| Read
    
    Write --> L1
    Write --> L2
    Write --> L3
    Write --> Invalidate
    
    Invalidate --> L1
    Invalidate --> L2
```

## Flux de Base de Données

```mermaid
flowchart TD
    subgraph "Application Layer"
        API[Backend API]
        Chat[Chat Server]
        Stream[Stream Server]
    end
    
    subgraph "Database Layer"
        Master[(Master DB<br/>PostgreSQL)]
        Replica[(Read Replica<br/>PostgreSQL)]
        Analytics[(Analytics DB<br/>PostgreSQL)]
    end
    
    subgraph "Connection Pool"
        Pool[Connection Pool<br/>100 Connections]
    end
    
    API --> Pool
    Chat --> Pool
    Stream --> Pool
    
    Pool --> Master
    Pool --> Replica
    Pool --> Analytics
    
    Master -->|Replication| Replica
    Master -->|ETL| Analytics
```

## Flux d'Événements

```mermaid
flowchart LR
    subgraph "Event Producers"
        API[Backend API]
        Chat[Chat Server]
        Stream[Stream Server]
    end
    
    subgraph "Event Bus"
        NATS[NATS JetStream]
    end
    
    subgraph "Event Consumers"
        Email[Email Service]
        SMS[SMS Service]
        Analytics[Analytics Service]
        Notification[Notification Service]
    end
    
    API --> NATS
    Chat --> NATS
    Stream --> NATS
    
    NATS --> Email
    NATS --> SMS
    NATS --> Analytics
    NATS --> Notification
```

## Flux de Monitoring

```mermaid
flowchart TD
    subgraph "Services"
        API[Backend API]
        Chat[Chat Server]
        Stream[Stream Server]
    end
    
    subgraph "Monitoring Stack"
        Prometheus[Prometheus<br/>Metrics Collection]
        Grafana[Grafana<br/>Visualization]
        Jaeger[Jaeger<br/>Distributed Tracing]
        ELK[ELK Stack<br/>Logs]
    end
    
    subgraph "Alerts"
        AlertManager[Alert Manager]
        Slack[Slack Notifications]
        Email[Email Alerts]
    end
    
    API --> Prometheus
    Chat --> Prometheus
    Stream --> Prometheus
    
    API --> Jaeger
    Chat --> Jaeger
    Stream --> Jaeger
    
    API --> ELK
    Chat --> ELK
    Stream --> ELK
    
    Prometheus --> Grafana
    Prometheus --> AlertManager
    AlertManager --> Slack
    AlertManager --> Email
```

## Flux de Sécurité

```mermaid
flowchart TD
    subgraph "Security Layer"
        JWT[JWT Authentication]
        RBAC[RBAC Authorization]
        RateLimit[Rate Limiting]
        CORS[CORS Policy]
        TLS[TLS Encryption]
    end
    
    subgraph "Request Flow"
        Request[Incoming Request]
        Auth[Authentication]
        Authz[Authorization]
        Process[Process Request]
        Response[Response]
    end
    
    Request --> TLS
    TLS --> CORS
    CORS --> RateLimit
    RateLimit --> JWT
    JWT --> Auth
    Auth --> RBAC
    RBAC --> Authz
    Authz --> Process
    Process --> Response
    Response --> TLS
```

## Métriques de Performance

### Latence Cible

| Service | Latence Cible | Latence Actuelle |
|---------|---------------|------------------|
| API REST | < 100ms | 50ms |
| WebSocket | < 50ms | 20ms |
| Database | < 10ms | 5ms |
| Cache | < 5ms | 2ms |

### Throughput Cible

| Service | Throughput Cible | Throughput Actuel |
|---------|------------------|-------------------|
| API Requests | 10,000 req/s | 8,000 req/s |
| WebSocket Connections | 50,000 concurrent | 30,000 concurrent |
| Database Queries | 5,000 q/s | 4,000 q/s |
| Cache Operations | 50,000 ops/s | 45,000 ops/s |

### Disponibilité

| Service | SLA Cible | Disponibilité Actuelle |
|---------|-----------|----------------------|
| API | 99.9% | 99.95% |
| Chat | 99.99% | 99.98% |
| Streaming | 99.5% | 99.7% |
| Database | 99.99% | 99.99% |

## Points de Contrôle

### Health Checks

- **API Health** : `/api/health`
- **Chat Health** : `/health`
- **Stream Health** : `/health`
- **Database Health** : Ping + Query
- **Redis Health** : Ping + Memory Usage
- **NATS Health** : Connection + JetStream

### Monitoring Points

- **Request Rate** : Requêtes par seconde
- **Error Rate** : Taux d'erreur
- **Response Time** : Temps de réponse
- **Connection Count** : Nombre de connexions
- **Memory Usage** : Utilisation mémoire
- **CPU Usage** : Utilisation CPU

---

**Dernière mise à jour** : $(date)
**Version du diagramme** : 1.0.0 