# 💬 Architecture Chat Server - Veza

## 📋 Vue d'ensemble

Le **veza-chat-server** est un service de chat en temps réel développé en Rust avec Tokio et Axum. Il gère les messages de chat directs et en salons via WebSocket et API REST, avec support pour la modération, la présence utilisateur et les analytics.

## 🎯 Responsabilités Principales

- **Chat en Temps Réel** : Gestion des messages via WebSocket
- **API REST** : Endpoints pour l'historique et la gestion des messages
- **Gestion des Salons** : Création, gestion et modération des salons
- **Présence Utilisateur** : Statut en ligne/hors ligne
- **Modération** : Filtrage de contenu et gestion des abus
- **Analytics** : Métriques de chat et statistiques
- **Persistance** : Stockage des messages et historique

## 🏗️ Architecture Globale

```mermaid
graph TB
    subgraph "Clients"
        WebClient[Web Client]
        MobileClient[Mobile Client]
        DesktopClient[Desktop Client]
    end
    
    subgraph "Chat Server (Rust)"
        WebSocketHandler[WebSocket Handler]
        RESTHandler[REST API Handler]
        MessageProcessor[Message Processor]
        ModerationEngine[Moderation Engine]
        PresenceManager[Presence Manager]
        AnalyticsCollector[Analytics Collector]
    end
    
    subgraph "Core Services"
        MessageStore[Message Store]
        RoomManager[Room Manager]
        UserManager[User Manager]
        CacheManager[Cache Manager]
    end
    
    subgraph "Infrastructure"
        PostgreSQL[(PostgreSQL)]
        Redis[(Redis Cache)]
        NATS[NATS Event Bus]
        Prometheus[Prometheus Metrics]
    end
    
    WebClient --> WebSocketHandler
    MobileClient --> WebSocketHandler
    DesktopClient --> WebSocketHandler
    
    WebClient --> RESTHandler
    MobileClient --> RESTHandler
    DesktopClient --> RESTHandler
    
    WebSocketHandler --> MessageProcessor
    RESTHandler --> MessageProcessor
    
    MessageProcessor --> ModerationEngine
    MessageProcessor --> PresenceManager
    MessageProcessor --> AnalyticsCollector
    
    MessageProcessor --> MessageStore
    MessageProcessor --> RoomManager
    MessageProcessor --> UserManager
    MessageProcessor --> CacheManager
    
    MessageStore --> PostgreSQL
    CacheManager --> Redis
    MessageProcessor --> NATS
    AnalyticsCollector --> Prometheus
```

## 📁 Structure des Dossiers

```
veza-chat-server/
├── src/
│   ├── main.rs                    # Point d'entrée principal
│   ├── lib.rs                     # Configuration de la bibliothèque
│   ├── config.rs                  # Configuration de l'application
│   ├── error.rs                   # Gestion des erreurs
│   ├── models.rs                  # Modèles de données
│   ├── websocket.rs               # Gestion WebSocket
│   ├── message_store.rs           # Stockage des messages
│   ├── simple_message_store.rs    # Store de messages simple
│   ├── optimized_persistence.rs   # Persistance optimisée
│   ├── moderation.rs              # Moteur de modération
│   ├── advanced_moderation.rs     # Modération avancée
│   ├── presence.rs                # Gestion de la présence
│   ├── permissions.rs             # Gestion des permissions
│   ├── authentication.rs          # Authentification
│   ├── security.rs                # Sécurité et validation
│   ├── rate_limiter.rs            # Limitation de taux
│   ├── cache.rs                   # Gestion du cache
│   ├── connection_pool.rs         # Pool de connexions
│   ├── monitoring.rs              # Monitoring et métriques
│   ├── services.rs                # Services métier
│   ├── utils.rs                   # Utilitaires
│   ├── validation.rs              # Validation des données
│   ├── client.rs                  # Client gRPC
│   ├── grpc_server.rs             # Serveur gRPC
│   ├── messages.rs                # Types de messages
│   ├── auth.rs                    # Authentification JWT
│   └── core/                      # Logique métier centrale
│       ├── connection.rs          # Gestion des connexions
│       ├── channels.rs            # Canaux de communication
│       ├── rate_limiter.rs        # Limitation de taux avancée
│       └── buffer.rs              # Gestion des buffers
│   ├── hub/                       # Hub de communication
│   │   ├── channels.rs            # Gestion des canaux
│   │   ├── room_websocket.rs      # WebSocket des salons
│   │   ├── audit.rs               # Audit des actions
│   │   ├── presence.rs            # Présence utilisateur
│   │   ├── moderation.rs          # Modération des salons
│   │   └── analytics.rs           # Analytics des salons
│   ├── generated/                 # Code généré (gRPC, protobuf)
│   └── bin/                       # Binaires utilitaires
├── proto/                         # Définitions protobuf
├── migrations/                    # Migrations de base de données
├── scripts/                       # Scripts utilitaires
├── Cargo.toml                     # Dépendances Rust
└── build.rs                       # Script de build
```

## 🔄 Flux de Données

### 1. Connexion WebSocket

```mermaid
sequenceDiagram
    participant Client
    participant WebSocketHandler
    participant AuthService
    participant ConnectionPool
    participant PresenceManager
    
    Client->>WebSocketHandler: WebSocket Upgrade Request
    WebSocketHandler->>AuthService: Validate JWT Token
    AuthService-->>WebSocketHandler: User Context
    WebSocketHandler->>ConnectionPool: Create Connection
    ConnectionPool->>PresenceManager: Update User Status
    PresenceManager-->>ConnectionPool: Status Updated
    ConnectionPool-->>WebSocketHandler: Connection Established
    WebSocketHandler-->>Client: WebSocket Connected
```

### 2. Envoi de Message

```mermaid
sequenceDiagram
    participant Client
    participant MessageProcessor
    participant ModerationEngine
    participant MessageStore
    participant RoomManager
    participant OtherClients
    
    Client->>MessageProcessor: Send Message
    MessageProcessor->>ModerationEngine: Validate Content
    ModerationEngine-->>MessageProcessor: Content Validated
    MessageProcessor->>MessageStore: Store Message
    MessageProcessor->>RoomManager: Broadcast to Room
    RoomManager->>OtherClients: Deliver Message
    MessageProcessor-->>Client: Message Confirmed
```

### 3. Gestion de la Présence

```mermaid
sequenceDiagram
    participant Client
    participant PresenceManager
    participant CacheManager
    participant OtherClients
    
    Client->>PresenceManager: Update Status
    PresenceManager->>CacheManager: Store Status
    PresenceManager->>OtherClients: Broadcast Status
    Client->>PresenceManager: Heartbeat
    PresenceManager->>CacheManager: Update Last Seen
    PresenceManager->>OtherClients: Update Online Status
```

## 🔧 Technologies Utilisées

### Framework et Librairies Principales
- **Tokio** : Runtime asynchrone
- **Axum** : Framework web HTTP/WebSocket
- **SQLx** : ORM pour PostgreSQL
- **Redis** : Cache et sessions
- **Serde** : Sérialisation/désérialisation
- **Tracing** : Logging et observabilité

### Communication
- **WebSocket** : Communication temps réel
- **gRPC** : Communication inter-services
- **REST API** : API HTTP traditionnelle
- **NATS** : Message broker

### Base de Données
- **PostgreSQL** : Base de données principale
- **Redis** : Cache et sessions

## 🚀 Points d'Entrée

### Serveur Principal (`src/main.rs`)

```rust
/**
 * @file main.rs
 * @description Point d'entrée principal du serveur de chat
 * @module veza-chat-server
 * @requires axum, tokio, tracing
 * @since 1.0.0
 * @author Équipe Veza
 */

#[tokio::main]
async fn main() -> Result<(), ChatError> {
    // 1. Configuration du logging
    // 2. Initialisation du store de messages
    // 3. Initialisation du gestionnaire WebSocket
    // 4. Configuration des routes
    // 5. Démarrage du serveur
}
```

### Configuration (`src/config.rs`)

```rust
/**
 * @struct Config
 * @description Configuration globale du serveur de chat
 * @fields server, database, redis, jwt, nats
 */
#[derive(Debug, Clone)]
pub struct Config {
    pub server: ServerConfig,
    pub database: DatabaseConfig,
    pub redis: RedisConfig,
    pub jwt: JWTConfig,
    pub nats: NATSConfig,
}
```

## 🔌 API Endpoints

### WebSocket Endpoints

#### GET /ws
**Description** : Connexion WebSocket pour le chat temps réel

**Headers** :
```http
Authorization: Bearer <jwt_token>
```

**Messages WebSocket** :

**Connexion** :
```json
{
  "type": "connect",
  "data": {
    "user_id": 123,
    "username": "john_doe",
    "room_id": 1
  }
}
```

**Envoi de message** :
```json
{
  "type": "message",
  "data": {
    "room_id": 1,
    "content": "Hello everyone!",
    "message_type": "text"
  }
}
```

**Typing indicator** :
```json
{
  "type": "typing",
  "data": {
    "room_id": 1,
    "is_typing": true
  }
}
```

**Réception de message** :
```json
{
  "type": "message",
  "data": {
    "id": 789,
    "room_id": 1,
    "user_id": 456,
    "username": "jane_doe",
    "content": "Hello everyone!",
    "message_type": "text",
    "timestamp": "2024-01-15T11:00:00Z"
  }
}
```

### REST API Endpoints

#### GET /health
**Description** : Vérification de santé du service

**Response** (200) :
```json
{
  "success": true,
  "data": {
    "status": "healthy",
    "service": "veza-chat-server",
    "version": "1.0.0",
    "timestamp": "2024-01-15T11:00:00Z",
    "connections": 150,
    "rooms": 25,
    "messages_per_second": 10.5
  },
  "message": "Service is healthy"
}
```

#### GET /api/messages
**Description** : Récupération des messages

**Query Parameters** :
- `room` : ID du salon
- `limit` : Nombre de messages (défaut: 50)
- `user1` : Premier utilisateur (messages directs)
- `user2` : Deuxième utilisateur (messages directs)

**Response** (200) :
```json
{
  "success": true,
  "data": [
    {
      "id": 789,
      "content": "Hello everyone!",
      "author": "john_doe",
      "room": "general",
      "timestamp": "2024-01-15T11:00:00Z"
    }
  ],
  "message": "Messages retrieved successfully"
}
```

#### POST /api/messages
**Description** : Envoi d'un message

**Request Body** :
```json
{
  "content": "Hello everyone!",
  "author": "john_doe",
  "room": "general",
  "is_direct": false
}
```

**Response** (201) :
```json
{
  "success": true,
  "data": 790,
  "message": "Message sent successfully"
}
```

#### GET /api/messages/stats
**Description** : Statistiques des messages

**Response** (200) :
```json
{
  "success": true,
  "data": {
    "total_messages": 50000,
    "messages_today": 1500,
    "active_rooms": 25,
    "online_users": 150,
    "messages_per_second": 10.5
  },
  "message": "Statistics retrieved successfully"
}
```

## 🛡️ Sécurité et Modération

### Système de Modération

```rust
/**
 * @struct ModerationEngine
 * @description Moteur de modération avancé
 * @fields filters, rules, actions
 */
pub struct ModerationEngine {
    filters: Vec<Box<dyn ContentFilter>>,
    rules: Vec<ModerationRule>,
    actions: Vec<ModerationAction>,
}
```

**Types de Filtres** :
- **Spam Detection** : Détection de spam
- **Profanity Filter** : Filtrage de langage inapproprié
- **Rate Limiting** : Limitation de taux par utilisateur
- **Content Analysis** : Analyse de contenu IA
- **User Reputation** : Système de réputation

**Actions de Modération** :
- **Warning** : Avertissement
- **Message Deletion** : Suppression de message
- **User Mute** : Mise en sourdine temporaire
- **User Ban** : Bannissement permanent
- **Room Lock** : Verrouillage de salon

### Authentification et Autorisation

```rust
/**
 * @struct AuthService
 * @description Service d'authentification JWT
 * @methods validate_token, get_user_context, check_permissions
 */
pub struct AuthService {
    jwt_secret: String,
    user_repository: Arc<UserRepository>,
}
```

**Validation JWT** :
```rust
pub async fn validate_token(&self, token: &str) -> Result<UserContext, AuthError> {
    // 1. Décodage du token JWT
    // 2. Vérification de la signature
    // 3. Vérification de l'expiration
    // 4. Récupération du contexte utilisateur
}
```

## 📊 Monitoring et Analytics

### Métriques Collectées

```rust
/**
 * @struct AnalyticsCollector
 * @description Collecteur de métriques et analytics
 * @fields metrics, events, performance
 */
pub struct AnalyticsCollector {
    metrics: Arc<MetricsRegistry>,
    events: Arc<EventCollector>,
    performance: Arc<PerformanceMonitor>,
}
```

**Métriques Clés** :
- **Messages par seconde** : Throughput des messages
- **Connexions actives** : Nombre d'utilisateurs connectés
- **Latence WebSocket** : Temps de réponse WebSocket
- **Taux d'erreur** : Pourcentage d'erreurs
- **Utilisation mémoire** : Consommation RAM
- **Utilisation CPU** : Charge processeur

### Dashboards Prometheus

```yaml
# Métriques personnalisées
veza_chat_messages_total{room="general"} 50000
veza_chat_connections_active 150
veza_chat_websocket_latency_seconds 0.05
veza_chat_error_rate 0.01
veza_chat_memory_usage_bytes 1073741824
```

## 🔄 Patterns Utilisés

### 1. Actor Pattern pour les Connexions

```rust
/**
 * @struct ConnectionActor
 * @description Actor pour gérer une connexion WebSocket
 * @methods handle_message, broadcast, close
 */
pub struct ConnectionActor {
    user_id: UserId,
    room_id: RoomId,
    sender: mpsc::UnboundedSender<OutgoingMessage>,
    receiver: mpsc::UnboundedReceiver<IncomingMessage>,
}
```

### 2. Repository Pattern pour les Données

```rust
/**
 * @trait MessageRepository
 * @description Interface pour l'accès aux messages
 * @methods save, find_by_room, find_by_user
 */
#[async_trait]
pub trait MessageRepository {
    async fn save(&self, message: Message) -> Result<MessageId, RepositoryError>;
    async fn find_by_room(&self, room_id: RoomId, limit: usize) -> Result<Vec<Message>, RepositoryError>;
    async fn find_by_user(&self, user_id: UserId, limit: usize) -> Result<Vec<Message>, RepositoryError>;
}
```

### 3. Observer Pattern pour les Événements

```rust
/**
 * @struct EventBus
 * @description Bus d'événements pour la communication inter-composants
 * @methods publish, subscribe, unsubscribe
 */
pub struct EventBus {
    subscribers: Arc<RwLock<HashMap<EventType, Vec<Box<dyn EventHandler>>>>>,
}
```

## 🧪 Tests

### Structure des Tests

```rust
#[cfg(test)]
mod tests {
    use super::*;
    
    #[tokio::test]
    async fn test_message_sending() {
        // Test d'envoi de message
    }
    
    #[tokio::test]
    async fn test_websocket_connection() {
        // Test de connexion WebSocket
    }
    
    #[tokio::test]
    async fn test_moderation_filter() {
        // Test de filtrage de contenu
    }
}
```

### Tests d'Intégration

```rust
#[tokio::test]
async fn test_full_chat_flow() {
    // 1. Connexion WebSocket
    // 2. Envoi de message
    // 3. Réception de message
    // 4. Vérification de persistance
}
```

## 🚀 Déploiement

### Configuration Docker

```dockerfile
FROM rust:1.70 as builder
WORKDIR /app
COPY . .
RUN cargo build --release

FROM debian:bullseye-slim
RUN apt-get update && apt-get install -y ca-certificates
COPY --from=builder /app/target/release/chat-server /usr/local/bin/
EXPOSE 3001
CMD ["chat-server"]
```

### Variables d'Environnement

```bash
# Configuration serveur
CHAT_PORT=3001
CHAT_HOST=0.0.0.0

# Base de données
DATABASE_URL=postgresql://user:pass@localhost/veza_chat
DATABASE_MAX_CONNECTIONS=100

# Redis
REDIS_URL=redis://localhost:6379
REDIS_POOL_SIZE=50

# JWT
JWT_SECRET=your-super-secret-key
JWT_EXPIRATION=3600

# NATS
NATS_URL=nats://localhost:4222
NATS_CLUSTER_ID=veza-cluster
```

## 📈 Performance

### Optimisations

- **Connection Pooling** : Pool de connexions DB optimisé
- **Message Batching** : Regroupement des messages
- **Memory Pooling** : Pool de mémoire pour les messages
- **Async I/O** : I/O asynchrone avec Tokio
- **Compression** : Compression des messages WebSocket

### Benchmarks

- **Connexions simultanées** : 10,000+
- **Messages par seconde** : 50,000+
- **Latence WebSocket** : < 10ms
- **Utilisation mémoire** : < 512MB
- **CPU usage** : < 30% en charge normale

---

**Dernière mise à jour** : $(date)
**Version** : 1.0.0
**Maintenu par** : Équipe Veza 