---
id: stream-server-architecture
title: Architecture Stream Server
sidebar_label: Stream Server Architecture
---

# 🎵 Architecture Stream Server - Veza

## 📋 Vue d'ensemble

Le **veza-stream-server** est un service de streaming audio haute performance développé en Rust avec Tokio et Axum. Il gère le streaming audio en temps réel, la compression, le transcodage, et l'analytics pour la plateforme Veza.

## 🎯 Responsabilités Principales

- **Streaming Audio** : Diffusion audio en temps réel
- **Transcodage** : Conversion entre formats audio (MP3, FLAC, AAC, OGG)
- **Compression** : Optimisation de la bande passante
- **Cache Management** : Mise en cache des fichiers audio
- **Analytics** : Métriques de streaming et analytics
- **Adaptive Streaming** : Adaptation de la qualité selon la bande passante
- **Live Recording** : Enregistrement de sessions en direct
- **SoundCloud Integration** : Intégration avec SoundCloud

## 🏗️ Architecture Globale

```mermaid
graph TB
    subgraph "Clients"
        WebClient[Web Client]
        MobileClient[Mobile Client]
        DesktopClient[Desktop Client]
        StreamClient[Stream Client]
    end
    
    subgraph "Stream Server (Rust)"
        HTTPHandler[HTTP Handler]
        WebSocketHandler[WebSocket Handler]
        StreamManager[Stream Manager]
        AudioProcessor[Audio Processor]
        CompressionEngine[Compression Engine]
        TranscodingService[Transcoding Service]
        CacheManager[Cache Manager]
        AnalyticsEngine[Analytics Engine]
    end
    
    subgraph "Core Services"
        AdaptiveStreaming[Adaptive Streaming]
        LiveRecording[Live Recording]
        SoundCloudIntegration[SoundCloud Integration]
        NotificationService[Notification Service]
        HealthMonitor[Health Monitor]
    end
    
    subgraph "Infrastructure"
        FileSystem[File System]
        Redis[(Redis Cache)]
        PostgreSQL[(PostgreSQL)]
        NATS[NATS Event Bus]
        Prometheus[Prometheus Metrics]
    end
    
    WebClient --> HTTPHandler
    MobileClient --> HTTPHandler
    DesktopClient --> HTTPHandler
    StreamClient --> WebSocketHandler
    
    HTTPHandler --> StreamManager
    WebSocketHandler --> StreamManager
    
    StreamManager --> AudioProcessor
    StreamManager --> CompressionEngine
    StreamManager --> TranscodingService
    StreamManager --> CacheManager
    StreamManager --> AnalyticsEngine
    
    AudioProcessor --> AdaptiveStreaming
    AudioProcessor --> LiveRecording
    AudioProcessor --> SoundCloudIntegration
    
    StreamManager --> NotificationService
    StreamManager --> HealthMonitor
    
    CacheManager --> Redis
    AnalyticsEngine --> PostgreSQL
    StreamManager --> NATS
    HealthMonitor --> Prometheus
    AudioProcessor --> FileSystem
```

## 📁 Structure des Dossiers

```
veza-stream-server/
├── src/
│   ├── main.rs                    # Point d'entrée principal
│   ├── lib.rs                     # Configuration de la bibliothèque
│   ├── config.rs                  # Configuration de l'application
│   ├── error.rs                   # Gestion des erreurs
│   ├── routes.rs                  # Définition des routes
│   ├── audio/                     # Traitement audio
│   │   ├── mod.rs                 # Module audio
│   │   ├── compression.rs         # Compression audio
│   │   ├── effects.rs             # Effets audio
│   │   ├── processing.rs          # Traitement audio
│   │   └── codecs/                # Codecs audio
│   │       ├── mod.rs             # Module codecs
│   │       ├── aac.rs             # Codec AAC
│   │       ├── flac.rs            # Codec FLAC
│   │       ├── mp3.rs             # Codec MP3
│   │       └── ogg.rs             # Codec OGG
│   ├── streaming/                 # Services de streaming
│   │   ├── mod.rs                 # Module streaming
│   │   ├── adaptive.rs            # Streaming adaptatif
│   │   ├── live_recording.rs      # Enregistrement en direct
│   │   └── websocket.rs           # WebSocket streaming
│   ├── core/                      # Logique métier centrale
│   │   ├── mod.rs                 # Module core
│   │   ├── buffer.rs              # Gestion des buffers
│   │   └── encoder.rs             # Encodeurs audio
│   ├── auth/                      # Authentification
│   │   ├── mod.rs                 # Module auth
│   ├── cache/                     # Gestion du cache
│   │   ├── mod.rs                 # Module cache
│   ├── analytics/                 # Analytics
│   │   ├── mod.rs                 # Module analytics
│   ├── health/                    # Monitoring de santé
│   │   ├── mod.rs                 # Module health
│   ├── monitoring/                # Monitoring et alertes
│   │   ├── mod.rs                 # Module monitoring
│   │   ├── metrics.rs             # Métriques
│   │   └── health_checks.rs       # Vérifications de santé
│   ├── middleware/                # Middleware HTTP
│   │   ├── mod.rs                 # Module middleware
│   │   ├── logging.rs             # Logging middleware
│   │   ├── rate_limit.rs          # Rate limiting
│   │   └── security.rs            # Headers de sécurité
│   ├── soundcloud/                # Intégration SoundCloud
│   │   ├── mod.rs                 # Module SoundCloud
│   │   ├── oauth.rs               # OAuth SoundCloud
│   │   └── upload.rs              # Upload vers SoundCloud
│   └── utils/                     # Utilitaires
│       ├── mod.rs                 # Module utils
│       └── metrics.rs             # Utilitaires métriques
├── proto/                         # Définitions protobuf
├── dashboards/                    # Dashboards Grafana
├── scripts/                       # Scripts utilitaires
├── Cargo.toml                     # Dépendances Rust
└── build.rs                       # Script de build
```

## 🔄 Flux de Données

### 1. Streaming Audio

```mermaid
sequenceDiagram
    participant Client
    participant StreamManager
    participant AudioProcessor
    participant CompressionEngine
    participant CacheManager
    participant FileSystem
    
    Client->>StreamManager: Request Audio Stream
    StreamManager->>CacheManager: Check Cache
    CacheManager-->>StreamManager: Cache Miss
    StreamManager->>FileSystem: Load Audio File
    FileSystem-->>StreamManager: Raw Audio Data
    StreamManager->>AudioProcessor: Process Audio
    AudioProcessor->>CompressionEngine: Compress Audio
    CompressionEngine-->>AudioProcessor: Compressed Audio
    AudioProcessor-->>StreamManager: Processed Audio
    StreamManager->>CacheManager: Store in Cache
    StreamManager-->>Client: Stream Audio Data
```

### 2. Live Recording

```mermaid
sequenceDiagram
    participant Client
    participant StreamManager
    participant LiveRecording
    participant FileSystem
    participant Analytics
    
    Client->>StreamManager: Start Live Stream
    StreamManager->>LiveRecording: Initialize Recording
    LiveRecording->>FileSystem: Create Recording File
    StreamManager-->>Client: Stream Started
    
    loop Audio Stream
        Client->>StreamManager: Audio Data
        StreamManager->>LiveRecording: Record Audio
        LiveRecording->>FileSystem: Write Audio Data
        StreamManager->>Analytics: Update Metrics
    end
    
    Client->>StreamManager: Stop Stream
    StreamManager->>LiveRecording: Finalize Recording
```

## 🔌 API Endpoints

### Streaming Endpoints

#### GET /api/v1/streams/stream_id
**Description** : Récupération d'un stream audio

**Response** :
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "title": "Live Music Session",
    "description": "Live music streaming",
    "is_live": true,
    "viewer_count": 250,
    "stream_url": "https://stream.veza.com/live/uuid.m3u8",
    "started_at": "2024-01-01T12:00:00Z"
  }
}
```

#### POST /api/v1/streams
**Description** : Création d'un nouveau stream

**Request Body** :
```json
{
  "title": "My Live Stream",
  "description": "Stream description",
  "is_private": false,
  "quality": "high"
}
```

### Audio Processing Endpoints

#### POST /api/v1/audio/compress
**Description** : Compression d'un fichier audio

**Request Body** :
```json
{
  "file_id": "uuid",
  "format": "mp3",
  "bitrate": 128,
  "quality": "high"
}
```

#### POST /api/v1/audio/transcode
**Description** : Transcodage d'un fichier audio

**Request Body** :
```json
{
  "file_id": "uuid",
  "input_format": "flac",
  "output_format": "aac",
  "bitrate": 256
}
```

## 🔧 Configuration

### Variables d'Environnement

```bash
# Serveur
STREAM_SERVER_PORT=8080
STREAM_SERVER_HOST=0.0.0.0

# Base de données
DATABASE_URL=postgresql://user:pass@localhost/veza_stream
REDIS_URL=redis://localhost:6379

# NATS
NATS_URL=nats://localhost:4222

# Audio
AUDIO_CACHE_DIR=/var/cache/veza/audio
AUDIO_UPLOAD_DIR=/var/uploads/veza/audio
MAX_AUDIO_FILE_SIZE=100MB

# Streaming
STREAM_BUFFER_SIZE=8192
STREAM_CHUNK_SIZE=4096
MAX_CONCURRENT_STREAMS=100

# Monitoring
PROMETHEUS_PORT=9090
GRAFANA_PORT=3000
```

### Configuration Rust

```rust
#[derive(Debug, Deserialize)]
pub struct Config {
    pub server: ServerConfig,
    pub database: DatabaseConfig,
    pub redis: RedisConfig,
    pub nats: NatsConfig,
    pub audio: AudioConfig,
    pub streaming: StreamingConfig,
    pub monitoring: MonitoringConfig,
}

#[derive(Debug, Deserialize)]
pub struct ServerConfig {
    pub port: u16,
    pub host: String,
    pub workers: usize,
}

#[derive(Debug, Deserialize)]
pub struct AudioConfig {
    pub cache_dir: String,
    pub upload_dir: String,
    pub max_file_size: usize,
    pub supported_formats: Vec<String>,
}
```

## 🔒 Sécurité

### Authentification
- **JWT Tokens** : Authentification des utilisateurs
- **API Keys** : Authentification des services
- **OAuth2** : Intégration SoundCloud

### Autorisation
- **RBAC** : Contrôle d'accès basé sur les rôles
- **Resource Permissions** : Permissions granulaires
- **Rate Limiting** : Limitation de taux par utilisateur

### Sécurité des Données
- **HTTPS/WSS** : Chiffrement en transit
- **File Validation** : Validation des fichiers audio
- **Input Sanitization** : Nettoyage des entrées
- **CORS** : Configuration Cross-Origin

## 📊 Monitoring & Analytics

### Métriques Prometheus

```rust
// Métriques de streaming
pub const STREAMS_ACTIVE: &str = "veza_streams_active_total";
pub const STREAMS_CREATED: &str = "veza_streams_created_total";
pub const STREAMS_ENDED: &str = "veza_streams_ended_total";

// Métriques audio
pub const AUDIO_FILES_PROCESSED: &str = "veza_audio_files_processed_total";
pub const AUDIO_PROCESSING_DURATION: &str = "veza_audio_processing_duration_seconds";
pub const AUDIO_CACHE_HITS: &str = "veza_audio_cache_hits_total";
pub const AUDIO_CACHE_MISSES: &str = "veza_audio_cache_misses_total";

// Métriques système
pub const MEMORY_USAGE: &str = "veza_memory_usage_bytes";
pub const CPU_USAGE: &str = "veza_cpu_usage_percent";
pub const DISK_USAGE: &str = "veza_disk_usage_bytes";
```

### Dashboards Grafana

- **Streaming Overview** : Vue d'ensemble des streams
- **Audio Processing** : Métriques de traitement audio
- **System Health** : Santé du système
- **Performance** : Métriques de performance

### Alertes

```yaml
# Alertes Prometheus
groups:
  - name: veza-stream-server
    rules:
      - alert: HighStreamLatency
        expr: veza_stream_latency_seconds > 5
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High stream latency detected"
          
      - alert: AudioProcessingErrors
        expr: rate(veza_audio_processing_errors_total[5m]) > 0.1
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "High audio processing error rate"
```

## 🧪 Tests

### Tests Unitaires
```rust
#[cfg(test)]
mod tests {
    use super::*;
    
    #[tokio::test]
    async fn test_audio_compression() {
        let processor = AudioProcessor::new();
        let result = processor.compress("test.flac", "mp3", 128).await;
        assert!(result.is_ok());
    }
    
    #[tokio::test]
    async fn test_stream_creation() {
        let manager = StreamManager::new();
        let stream = manager.create_stream("Test Stream").await;
        assert!(stream.is_ok());
    }
}
```

### Tests d'Intégration
```rust
#[tokio::test]
async fn test_full_streaming_flow() {
    // Test complet du flux de streaming
    let client = TestClient::new();
    
    // Créer un stream
    let stream = client.create_stream("Test").await.unwrap();
    
    // Démarrer le streaming
    let stream_id = stream.id.clone();
    client.start_stream(&stream_id).await.unwrap();
    
    // Vérifier que le stream est actif
    let status = client.get_stream_status(&stream_id).await.unwrap();
    assert_eq!(status, "active");
    
    // Arrêter le stream
    client.stop_stream(&stream_id).await.unwrap();
}
```

## 🚀 Déploiement

### Docker

```dockerfile
FROM rust:1.70 as builder
WORKDIR /app
COPY . .
RUN cargo build --release

FROM debian:bullseye-slim
RUN apt-get update && apt-get install -y \
    ffmpeg \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/target/release/veza-stream-server /usr/local/bin/
COPY --from=builder /app/config /app/config

EXPOSE 8080
CMD ["veza-stream-server"]
```

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: veza-stream-server
spec:
  replicas: 3
  selector:
    matchLabels:
      app: veza-stream-server
  template:
    metadata:
      labels:
        app: veza-stream-server
    spec:
      containers:
      - name: stream-server
        image: veza/stream-server:latest
        ports:
        - containerPort: 8080
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: veza-secrets
              key: database-url
        - name: REDIS_URL
          valueFrom:
            secretKeyRef:
              name: veza-secrets
              key: redis-url
        resources:
          requests:
            memory: "512Mi"
            cpu: "500m"
          limits:
            memory: "2Gi"
            cpu: "2000m"
```

## 📈 Performance

### Benchmarks

```rust
#[bench]
fn bench_audio_compression(b: &mut Bencher) {
    let processor = AudioProcessor::new();
    let audio_data = load_test_audio();
    
    b.iter(|| {
        processor.compress_sync(&audio_data, "mp3", 128)
    });
}

#[bench]
fn bench_stream_creation(b: &mut Bencher) {
    let manager = StreamManager::new();
    
    b.iter(|| {
        manager.create_stream_sync("Test Stream")
    });
}
```

### Optimisations

- **Async/Await** : Traitement asynchrone
- **Connection Pooling** : Pool de connexions DB
- **Memory Pooling** : Pool de buffers audio
- **Compression** : Compression des données
- **Caching** : Cache multi-niveaux
- **Load Balancing** : Répartition de charge

## 🔗 Intégrations

### SoundCloud
- **OAuth2** : Authentification
- **Upload API** : Upload de fichiers
- **Discovery API** : Recherche de contenu
- **Management API** : Gestion des comptes

### NATS
- **Event Publishing** : Publication d'événements
- **Event Subscription** : Abonnement aux événements
- **Request/Response** : Communication synchrone

### Prometheus
- **Metrics Collection** : Collecte de métriques
- **Health Checks** : Vérifications de santé
- **Alerting** : Système d'alertes

## 📚 Documentation

### API Documentation
- **OpenAPI/Swagger** : Documentation interactive
- **Postman Collection** : Collection de tests
- **Examples** : Exemples d'utilisation

### Guides
- **Getting Started** : Guide de démarrage
- **Deployment** : Guide de déploiement
- **Troubleshooting** : Guide de dépannage

### Références
- **Configuration** : Référence de configuration
- **API Reference** : Référence API complète
- **Architecture** : Documentation d'architecture 