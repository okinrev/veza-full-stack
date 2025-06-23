# Streaming Module Documentation  

Le module streaming fournit des fonctionnalités avancées de streaming adaptatif et de communication temps réel via WebSocket pour le serveur audio.

## Table des Matières

- [Vue d'ensemble](#vue-densemble)
- [Architecture](#architecture)
- [Streaming Adaptatif](#streaming-adaptatif)
- [WebSocket Manager](#websocket-manager)
- [Types et Structures](#types-et-structures)
- [Protocoles de Communication](#protocoles-de-communication)
- [API Reference](#api-reference)
- [Exemples d'utilisation](#exemples-dutilisation)
- [Intégration](#intégration)

## Vue d'ensemble

Le module streaming se compose de deux parties principales :
- **Streaming Adaptatif** : Ajuste automatiquement la qualité selon les conditions réseau
- **WebSocket Manager** : Gère les connexions temps réel pour les événements et commandes

### Fonctionnalités clés

**Streaming Adaptatif :**
- Détection automatique de la bande passante
- Adaptation de la qualité en temps réel
- Support HLS (HTTP Live Streaming)
- Profils de qualité multiples
- Métriques de performance

**WebSocket :**
- Connexions bidirectionnelles temps réel
- Système d'événements sophistiqué
- Commandes avec réponses asynchrones
- Souscriptions sélectives aux événements
- Diffusion (broadcast) efficace

## Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│    Client       │    │   Adaptive       │    │   WebSocket     │
│                 │    │   Streaming      │    │   Manager       │
│ - HLS Player    │◄──►│                  │◄──►│                 │
│ - WebSocket     │    │ - Quality        │    │ - Events        │
│ - Bandwidth     │    │   Selection      │    │ - Commands      │
│   Detection     │    │ - Performance    │    │ - Broadcast     │
│                 │    │   Monitoring     │    │ - Sessions      │
└─────────────────┘    └──────────────────┘    └─────────────────┘
         │                        │                        │
         └────────────────────────┼────────────────────────┘
                                  ▼
                    ┌─────────────────────────┐
                    │    Analytics Engine     │
                    │                         │
                    │ - Session Tracking      │
                    │ - Performance Metrics   │
                    │ - Quality Statistics    │
                    │ - User Behavior         │
                    └─────────────────────────┘
```

## Streaming Adaptatif

### AdaptiveStreamingManager

```rust
pub struct AdaptiveStreamingManager {
    config: Arc<Config>,                                    // Configuration système
    sessions: Arc<RwLock<HashMap<String, StreamingSession>>>, // Sessions actives
    profiles: Vec<AdaptiveProfile>,                         // Profils de qualité
}
```

### AdaptiveProfile

```rust
pub struct AdaptiveProfile {
    pub quality_id: String,             // Identifiant unique (high, medium, low)
    pub bitrate_kbps: u32,             // Débit en kbps
    pub sample_rate: u32,              // Taux d'échantillonnage
    pub channels: u8,                  // Nombre de canaux
    pub codec: String,                 // Codec utilisé
    pub file_extension: String,        // Extension de fichier
    pub bandwidth_estimate_kbps: u32,  // Bande passante estimée requise
}
```

**Profils prédéfinis :**

```rust
impl AdaptiveProfile {
    pub fn high_quality() -> Self {
        Self {
            quality_id: "high".to_string(),
            bitrate_kbps: 320,
            sample_rate: 44100,
            channels: 2,
            codec: "aac".to_string(),
            file_extension: "aac".to_string(),
            bandwidth_estimate_kbps: 400, // Marge pour les variations réseau
        }
    }
    
    pub fn medium_quality() -> Self {
        Self {
            quality_id: "medium".to_string(),
            bitrate_kbps: 192,
            sample_rate: 44100,
            channels: 2,
            codec: "aac".to_string(),
            file_extension: "aac".to_string(),
            bandwidth_estimate_kbps: 250,
        }
    }
    
    pub fn low_quality() -> Self {
        Self {
            quality_id: "low".to_string(),
            bitrate_kbps: 128,
            sample_rate: 22050,
            channels: 2,
            codec: "mp3".to_string(),
            file_extension: "mp3".to_string(),
            bandwidth_estimate_kbps: 160,
        }
    }
    
    pub fn mobile_quality() -> Self {
        Self {
            quality_id: "mobile".to_string(),
            bitrate_kbps: 96,
            sample_rate: 22050,
            channels: 2,
            codec: "aac".to_string(),
            file_extension: "aac".to_string(),
            bandwidth_estimate_kbps: 120,
        }
    }
}
```

### ClientCapabilities

```rust
pub struct ClientCapabilities {
    pub estimated_bandwidth_kbps: u32,      // Bande passante estimée
    pub buffer_duration_ms: u32,            // Durée du buffer
    pub connection_type: ConnectionType,    // Type de connexion
    pub device_type: DeviceType,           // Type d'appareil
    pub preferred_quality: Option<String>,  // Qualité préférée
    pub adaptive_enabled: bool,            // Adaptation automatique activée
}
```

### ConnectionType

```rust
pub enum ConnectionType {
    Wifi,           // WiFi
    Cellular4G,     // 4G/LTE
    Cellular3G,     // 3G
    Cellular2G,     // 2G/Edge
    Ethernet,       // Ethernet filaire
    Unknown,        // Type inconnu
}
```

### DeviceType

```rust
pub enum DeviceType {
    Desktop,        // Ordinateur de bureau
    Mobile,         // Téléphone mobile
    Tablet,         // Tablette
    SmartSpeaker,   // Enceinte connectée
    CarAudio,       // Système audio voiture
    Unknown,        // Appareil inconnu
}
```

### StreamingSession

```rust
pub struct StreamingSession {
    pub session_id: String,                     // ID unique de session
    pub track_id: String,                       // ID de la piste
    pub client_id: String,                      // ID du client
    pub current_quality: String,                // Qualité actuelle
    pub client_capabilities: ClientCapabilities, // Capacités du client
    pub performance_metrics: PerformanceMetrics, // Métriques de performance
    pub created_at: SystemTime,                 // Timestamp de création
    pub last_updated: SystemTime,               // Dernière mise à jour
}
```

### PerformanceMetrics

```rust
pub struct PerformanceMetrics {
    pub buffer_health_percentage: f32,      // Santé du buffer (0-100%)
    pub download_speed_kbps: u32,          // Vitesse de téléchargement
    pub packet_loss_percentage: f32,       // Taux de perte de paquets
    pub latency_ms: u32,                   // Latence réseau
    pub rebuffer_count: u32,               // Nombre de rebuffering
    pub rebuffer_duration_ms: u32,         // Durée totale de rebuffering
    pub quality_switches: u32,             // Nombre de changements de qualité
}
```

### HLS (HTTP Live Streaming)

#### MasterPlaylist

```rust
pub struct MasterPlaylist {
    pub version: u8,                    // Version HLS
    pub streams: Vec<StreamInfo>,       // Informations des streams
}

pub struct StreamInfo {
    pub bandwidth: u32,                 // Bande passante requise
    pub codecs: String,                 // Codecs utilisés
    pub resolution: Option<String>,     // Résolution (pour vidéo)
    pub url: String,                    // URL du playlist spécifique
}
```

#### HLSManifest

```rust
pub struct HLSManifest {
    pub version: u8,                    // Version HLS
    pub target_duration: u32,           // Durée cible des segments
    pub media_sequence: u32,            // Numéro de séquence
    pub segments: Vec<HLSSegment>,      // Liste des segments
    pub end_list: bool,                 // Fin de playlist (VOD)
}

pub struct HLSSegment {
    pub duration: f32,                  // Durée du segment
    pub url: String,                    // URL du segment
    pub byte_range: Option<(u64, u64)>, // Range d'octets (optionnel)
}
```

## WebSocket Manager

### WebSocketManager

```rust
pub struct WebSocketManager {
    connections: Arc<RwLock<HashMap<Uuid, WebSocketConnection>>>, // Connexions actives
    global_sender: broadcast::Sender<WebSocketEvent>,            // Canal de diffusion
    stats: Arc<RwLock<WebSocketStats>>,                         // Statistiques
}
```

### WebSocketConnection

```rust
pub struct WebSocketConnection {
    pub id: Uuid,                           // ID unique de connexion
    pub user_id: Option<String>,            // ID utilisateur (optionnel)
    pub ip_address: String,                 // Adresse IP
    pub connected_at: SystemTime,           // Timestamp de connexion
    pub last_activity: SystemTime,          // Dernière activité
    pub subscribed_events: Vec<String>,     // Événements souscrits
    pub sender: broadcast::Sender<WebSocketEvent>, // Canal personnel
}
```

### WebSocketStats

```rust
struct WebSocketStats {
    total_connections: u64,         // Total des connexions
    current_connections: u32,       // Connexions actuelles
    messages_sent: u64,            // Messages envoyés
    messages_received: u64,        // Messages reçus
    events_broadcasted: u64,       // Événements diffusés
}
```

## Types et Structures

### WebSocketEvent (Événements sortants)

```rust
pub enum WebSocketEvent {
    /// Événements de lecture
    PlaybackStarted {
        track_id: String,
        user_id: String,
        timestamp: u64,
    },
    PlaybackPaused {
        track_id: String,
        user_id: String,
        position_ms: u64,
    },
    PlaybackResumed {
        track_id: String,
        user_id: String,
        position_ms: u64,
    },
    PlaybackStopped {
        track_id: String,
        user_id: String,
        total_played_ms: u64,
    },
    PlaybackProgress {
        track_id: String,
        user_id: String,
        position_ms: u64,
        buffer_percentage: f32,
    },

    /// Événements de playlist
    PlaylistUpdated {
        playlist_id: String,
        action: PlaylistAction,
        track_id: Option<String>,
        position: Option<usize>,
    },
    PlaylistShared {
        playlist_id: String,
        from_user: String,
        to_users: Vec<String>,
    },

    /// Événements sociaux
    TrackLiked {
        track_id: String,
        user_id: String,
        total_likes: u64,
    },
    TrackShared {
        track_id: String,
        from_user: String,
        to_users: Vec<String>,
        message: Option<String>,
    },
    UserFollowed {
        follower_id: String,
        followed_id: String,
    },

    /// Événements système
    ServerMessage {
        message: String,
        level: MessageLevel,
    },
    RateLimitWarning {
        remaining_requests: u32,
        reset_time_seconds: u64,
    },

    /// Réponses aux commandes
    CommandResponse {
        command_id: String,
        success: bool,
        data: Option<serde_json::Value>,
        error: Option<String>,
    },

    /// Statistiques en temps réel
    LiveStats {
        concurrent_listeners: u32,
        top_tracks: Vec<LiveTrackStats>,
        server_load: f32,
    },
}
```

### WebSocketCommand (Commandes entrantes)

```rust
pub enum WebSocketCommand {
    /// Commandes de lecture
    Play {
        command_id: String,
        track_id: String,
        start_position_ms: Option<u64>,
    },
    Pause {
        command_id: String,
        track_id: String,
    },
    Stop {
        command_id: String,
        track_id: String,
    },
    Seek {
        command_id: String,
        track_id: String,
        position_ms: u64,
    },

    /// Commandes de playlist
    AddToPlaylist {
        command_id: String,
        playlist_id: String,
        track_id: String,
        position: Option<usize>,
    },
    RemoveFromPlaylist {
        command_id: String,
        playlist_id: String,
        track_id: String,
    },

    /// Commandes d'abonnement
    Subscribe {
        command_id: String,
        events: Vec<String>,
        filters: Option<HashMap<String, String>>,
    },
    Unsubscribe {
        command_id: String,
        events: Vec<String>,
    },

    /// Commandes de statut
    GetStatus {
        command_id: String,
    },
    Ping {
        command_id: String,
    },
}
```

### Enums supportés

```rust
pub enum PlaylistAction {
    Added,      // Piste ajoutée
    Removed,    // Piste supprimée
    Reordered,  // Ordre modifié
    Cleared,    // Playlist vidée
}

pub enum MessageLevel {
    Info,       // Information
    Warning,    // Avertissement
    Error,      // Erreur
}
```

## Protocoles de Communication

### Connexion WebSocket

**URL de connexion :**
```
ws://localhost:8082/api/ws?user_id=123&token=jwt_token_here
```

**Paramètres de query :**
- `user_id` (optionnel) : Identifiant utilisateur
- `token` (optionnel) : Token JWT pour authentification

### Format des messages

**Commande (Client → Serveur) :**
```json
{
  "command": "Play",
  "command_id": "cmd_123",
  "track_id": "song_456",
  "start_position_ms": 30000
}
```

**Événement (Serveur → Client) :**
```json
{
  "event": "PlaybackStarted",
  "track_id": "song_456",
  "user_id": "user_123",
  "timestamp": 1640995200
}
```

**Réponse de commande :**
```json
{
  "event": "CommandResponse",
  "command_id": "cmd_123",
  "success": true,
  "data": {
    "track_duration_ms": 210000,
    "quality": "high"
  },
  "error": null
}
```

### Souscription aux événements

```json
{
  "command": "Subscribe",
  "command_id": "sub_001",
  "events": ["PlaybackStarted", "PlaybackProgress", "TrackLiked"],
  "filters": {
    "user_id": "user_123",
    "track_genre": "electronic"
  }
}
```

## API Reference

### AdaptiveStreamingManager Methods

#### `new(config: Arc<Config>) -> Self`
Crée une nouvelle instance du gestionnaire de streaming adaptatif.

#### `start_quality_monitor(&self)`
Démarre le monitoring automatique de la qualité.

#### `create_session(session_id: String, track_id: String) -> AdaptiveProfile`
Crée une nouvelle session de streaming et retourne le profil initial.

#### `update_session_quality(session_id: &str, quality: String)`
Met à jour la qualité d'une session existante.

#### `generate_master_playlist(track_id: &str) -> Result<String, Error>`
Génère un master playlist HLS pour une piste.

#### `generate_quality_playlist(track_id: &str, quality: &str) -> Result<String, Error>`
Génère un playlist HLS pour une qualité spécifique.

#### `get_streaming_stats() -> serde_json::Value`
Retourne les statistiques de streaming.

### WebSocketManager Methods

#### `new() -> Self`
Crée une nouvelle instance du gestionnaire WebSocket.

#### `handle_websocket(ws: WebSocketUpgrade, user_id: Option<String>, ip_address: String) -> Response`
Gère une nouvelle connexion WebSocket.

#### `broadcast_event(event: WebSocketEvent)`
Diffuse un événement à toutes les connexions.

#### `send_to_connection(connection_id: Uuid, event: WebSocketEvent)`
Envoie un événement à une connexion spécifique.

#### `send_to_user(user_id: &str, event: WebSocketEvent)`
Envoie un événement à toutes les connexions d'un utilisateur.

#### `cleanup_inactive_connections(max_idle_duration: Duration)`
Nettoie les connexions inactives.

#### `get_stats() -> serde_json::Value`
Retourne les statistiques WebSocket.

## Exemples d'utilisation

### Streaming adaptatif complet

```rust
use stream_server::streaming::adaptive::{AdaptiveStreamingManager, ClientCapabilities, ConnectionType, DeviceType};

async fn example_adaptive_streaming() -> Result<(), Box<dyn std::error::Error>> {
    let config = Arc::new(Config::from_env()?);
    let streaming_manager = AdaptiveStreamingManager::new(config);
    
    // Démarrer le monitoring de qualité
    streaming_manager.start_quality_monitor().await;
    
    // Simuler les capacités d'un client mobile
    let client_capabilities = ClientCapabilities {
        estimated_bandwidth_kbps: 150, // Connexion limitée
        buffer_duration_ms: 5000,      // Buffer de 5 secondes
        connection_type: ConnectionType::Cellular3G,
        device_type: DeviceType::Mobile,
        preferred_quality: None,        // Laisser l'adaptation automatique
        adaptive_enabled: true,
    };
    
    // Créer une session de streaming
    let session_id = "session_mobile_001".to_string();
    let track_id = "epic_song_123".to_string();
    
    let initial_profile = streaming_manager.create_session(session_id.clone(), track_id.clone()).await;
    
    println!("🎵 Session créée: {}", session_id);
    println!("🎚️  Qualité initiale: {} ({}kbps)", initial_profile.quality_id, initial_profile.bitrate_kbps);
    
    // Générer le master playlist HLS
    let master_playlist = streaming_manager.generate_master_playlist(&track_id).await?;
    println!("\n📋 Master Playlist HLS:");
    println!("{}", master_playlist);
    
    // Simuler une amélioration de la connexion
    tokio::time::sleep(Duration::from_secs(10)).await;
    
    // Mettre à jour vers une meilleure qualité
    streaming_manager.update_session_quality(&session_id, "medium".to_string()).await;
    println!("\n📶 Qualité mise à jour vers: medium");
    
    // Afficher les statistiques
    let stats = streaming_manager.get_streaming_stats().await;
    println!("\n📊 Statistiques de streaming:");
    println!("{}", serde_json::to_string_pretty(&stats)?);
    
    Ok(())
}
```

### Gestion WebSocket avancée

```rust
use stream_server::streaming::websocket::{WebSocketManager, WebSocketEvent, WebSocketCommand};

async fn example_websocket_management() -> Result<(), Box<dyn std::error::Error>> {
    let ws_manager = Arc::new(WebSocketManager::new());
    
    // Simuler une connexion WebSocket (normalement géré par axum)
    println!("🔌 Simulation d'une connexion WebSocket");
    
    // Créer des événements de test
    let events = vec![
        WebSocketEvent::PlaybackStarted {
            track_id: "song_001".to_string(),
            user_id: "user_123".to_string(),
            timestamp: SystemTime::now().duration_since(UNIX_EPOCH)?.as_secs(),
        },
        WebSocketEvent::PlaybackProgress {
            track_id: "song_001".to_string(),
            user_id: "user_123".to_string(),
            position_ms: 45000, // 45 secondes
            buffer_percentage: 87.5,
        },
        WebSocketEvent::TrackLiked {
            track_id: "song_001".to_string(),
            user_id: "user_456".to_string(),
            total_likes: 127,
        },
        WebSocketEvent::ServerMessage {
            message: "Nouvelle fonctionnalité disponible!".to_string(),
            level: MessageLevel::Info,
        },
    ];
    
    // Diffuser les événements
    for event in events {
        ws_manager.broadcast_event(event).await;
        tokio::time::sleep(Duration::from_millis(500)).await;
    }
    
    // Statistiques temps réel
    let live_stats = WebSocketEvent::LiveStats {
        concurrent_listeners: 1247,
        top_tracks: vec![
            LiveTrackStats {
                track_id: "trending_001".to_string(),
                title: "Summer Vibes".to_string(),
                artist: "DJ Sunshine".to_string(),
                current_listeners: 89,
            },
            LiveTrackStats {
                track_id: "trending_002".to_string(),
                title: "Night Drive".to_string(),
                artist: "Neon Pulse".to_string(),
                current_listeners: 67,
            },
        ],
        server_load: 0.32, // 32% de charge
    };
    
    ws_manager.broadcast_event(live_stats).await;
    
    // Afficher les statistiques WebSocket
    let stats = ws_manager.get_stats().await;
    println!("\n📊 Statistiques WebSocket:");
    println!("{}", serde_json::to_string_pretty(&stats)?);
    
    Ok(())
}
```

### Client WebSocket interactif

```rust
use tokio_tungstenite::{connect_async, tungstenite::Message};
use serde_json::{json, Value};

async fn example_websocket_client() -> Result<(), Box<dyn std::error::Error>> {
    // Se connecter au serveur WebSocket
    let url = "ws://localhost:8082/api/ws?user_id=client_001";
    println!("🔌 Connexion à: {}", url);
    
    let (ws_stream, _) = connect_async(url).await?;
    let (mut write, mut read) = ws_stream.split();
    
    // Tâche pour écouter les messages du serveur
    let read_task = tokio::spawn(async move {
        while let Some(msg) = read.next().await {
            match msg {
                Ok(Message::Text(text)) => {
                    if let Ok(event) = serde_json::from_str::<Value>(&text) {
                        println!("📨 Événement reçu: {}", serde_json::to_string_pretty(&event).unwrap());
                    }
                }
                Ok(Message::Close(_)) => {
                    println!("🔌 Connexion fermée par le serveur");
                    break;
                }
                Err(e) => {
                    println!("❌ Erreur WebSocket: {}", e);
                    break;
                }
                _ => {}
            }
        }
    });
    
    // Envoyer des commandes de test
    let commands = vec![
        // S'abonner aux événements de lecture
        json!({
            "command": "Subscribe",
            "command_id": "sub_001",
            "events": ["PlaybackStarted", "PlaybackProgress", "PlaybackStopped"],
            "filters": {
                "user_id": "client_001"
            }
        }),
        
        // Lancer la lecture d'une piste
        json!({
            "command": "Play",
            "command_id": "play_001",
            "track_id": "favorite_song",
            "start_position_ms": 0
        }),
        
        // Simuler une progression
        json!({
            "command": "Seek",
            "command_id": "seek_001",
            "track_id": "favorite_song",
            "position_ms": 60000
        }),
        
        // Demander le statut
        json!({
            "command": "GetStatus",
            "command_id": "status_001"
        }),
        
        // Ping pour tester la connexion
        json!({
            "command": "Ping",
            "command_id": "ping_001"
        }),
    ];
    
    for (i, command) in commands.iter().enumerate() {
        println!("📤 Envoi de la commande {}:", i + 1);
        println!("{}", serde_json::to_string_pretty(command)?);
        
        let message = Message::Text(serde_json::to_string(command)?);
        write.send(message).await?;
        
        // Attendre entre les commandes
        tokio::time::sleep(Duration::from_secs(2)).await;
    }
    
    // Attendre un peu pour voir les réponses
    tokio::time::sleep(Duration::from_secs(10)).await;
    
    // Fermer la connexion
    write.close().await?;
    read_task.await?;
    
    Ok(())
}
```

### Intégration avec analytics

```rust
use stream_server::{streaming::websocket::*, analytics::AnalyticsEngine};

async fn example_streaming_with_analytics() -> Result<(), Box<dyn std::error::Error>> {
    let config = Arc::new(Config::from_env()?);
    let analytics = Arc::new(AnalyticsEngine::new("sqlite:analytics.db", config.clone()).await?);
    let ws_manager = Arc::new(WebSocketManager::new());
    
    // Commencer une session d'écoute
    let session_id = analytics.start_play_session(
        Some("user_123".to_string()),
        "epic_track_001".to_string(),
        "192.168.1.100".to_string(),
        Some("WebSocket Client/1.0".to_string()),
        180_000, // 3 minutes
        "high".to_string(),
        Platform::Web,
        Some("https://musicapp.com".to_string()),
    ).await;
    
    // Notifier via WebSocket
    let start_event = WebSocketEvent::PlaybackStarted {
        track_id: "epic_track_001".to_string(),
        user_id: "user_123".to_string(),
        timestamp: SystemTime::now().duration_since(UNIX_EPOCH)?.as_secs(),
    };
    ws_manager.broadcast_event(start_event).await;
    
    // Simuler la progression avec analytics et WebSocket
    for position in (0..=180_000).step_by(15_000) {
        // Mettre à jour les analytics
        analytics.update_play_progress(
            session_id,
            position,
            Some(90.0 + (position as f32 / 180_000.0) * 10.0), // Buffer 90-100%
        ).await;
        
        // Notifier la progression via WebSocket
        let progress_event = WebSocketEvent::PlaybackProgress {
            track_id: "epic_track_001".to_string(),
            user_id: "user_123".to_string(),
            position_ms: position,
            buffer_percentage: 90.0 + (position as f32 / 180_000.0) * 10.0,
        };
        ws_manager.broadcast_event(progress_event).await;
        
        println!("📊 Position: {}ms, Buffer: {:.1}%", position, 90.0 + (position as f32 / 180_000.0) * 10.0);
        
        tokio::time::sleep(Duration::from_millis(200)).await; // Simulation
    }
    
    // Terminer la session
    analytics.end_play_session(session_id, None).await;
    
    let stop_event = WebSocketEvent::PlaybackStopped {
        track_id: "epic_track_001".to_string(),
        user_id: "user_123".to_string(),
        total_played_ms: 180_000,
    };
    ws_manager.broadcast_event(stop_event).await;
    
    println!("✅ Session terminée avec analytics et notifications WebSocket");
    
    Ok(())
}
```

## Intégration

### Avec le serveur principal

```rust
// Dans main.rs
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let config = Arc::new(Config::from_env()?);
    
    // Créer les gestionnaires de streaming
    let adaptive_streaming = Arc::new(AdaptiveStreamingManager::new(config.clone()));
    let websocket_manager = Arc::new(WebSocketManager::new());
    
    // Démarrer le monitoring adaptatif
    adaptive_streaming.start_quality_monitor().await;
    
    // Router avec endpoints de streaming
    let app = Router::new()
        // HLS Master Playlist
        .route("/hls/:track_id/master.m3u8", get(hls_master_playlist))
        .route("/hls/:track_id/:quality/playlist.m3u8", get(hls_quality_playlist))
        
        // WebSocket
        .route("/api/ws", get(websocket_handler))
        
        // Streaming adaptatif
        .route("/stream/:filename", get(adaptive_stream_handler))
        
        .with_state(AppState {
            adaptive_streaming,
            websocket_manager,
            // ... autres composants
        });
    
    // Démarrer le serveur
    let listener = tokio::net::TcpListener::bind("0.0.0.0:8082").await?;
    axum::serve(listener, app).await?;
    
    Ok(())
}

// Handler pour streaming adaptatif
async fn adaptive_stream_handler(
    Path(filename): Path<String>,
    Query(params): Query<HashMap<String, String>>,
    State(state): State<AppState>,
    headers: HeaderMap,
) -> Result<Response, (StatusCode, String)> {
    // Détecter les capacités du client
    let client_capabilities = detect_client_capabilities(&headers, &params);
    
    // Créer une session de streaming
    let session_id = Uuid::new_v4().to_string();
    let profile = state.adaptive_streaming.create_session(session_id.clone(), filename.clone()).await;
    
    // Servir le fichier avec la qualité adaptée
    serve_adaptive_file(&state.config, &filename, &profile, headers).await
}
```

### Avec le frontend React

```typescript
// Hook pour streaming adaptatif
export function useAdaptiveStreaming() {
  const [currentQuality, setCurrentQuality] = useState<string>('medium');
  const [bufferHealth, setBufferHealth] = useState<number>(100);
  const [isAdaptive, setIsAdaptive] = useState<boolean>(true);
  
  const createStreamingSession = async (trackId: string): Promise<string> => {
    const response = await fetch('/api/streaming/session', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        track_id: trackId,
        client_capabilities: {
          estimated_bandwidth_kbps: navigator.connection?.downlink * 1000 || 1000,
          buffer_duration_ms: 5000,
          connection_type: detectConnectionType(),
          device_type: detectDeviceType(),
          adaptive_enabled: isAdaptive,
        }
      }),
    });
    
    const { session_id, initial_profile } = await response.json();
    setCurrentQuality(initial_profile.quality_id);
    return session_id;
  };
  
  const getHLSMasterPlaylist = (trackId: string): string => {
    return `/hls/${trackId}/master.m3u8`;
  };
  
  return {
    currentQuality,
    bufferHealth,
    isAdaptive,
    setIsAdaptive,
    createStreamingSession,
    getHLSMasterPlaylist,
  };
}

// Hook pour WebSocket
export function useWebSocket() {
  const [socket, setSocket] = useState<WebSocket | null>(null);
  const [connectionStatus, setConnectionStatus] = useState<'connecting' | 'connected' | 'disconnected'>('disconnected');
  const [events, setEvents] = useState<WebSocketEvent[]>([]);
  
  const connect = useCallback((userId?: string, token?: string) => {
    const params = new URLSearchParams();
    if (userId) params.append('user_id', userId);
    if (token) params.append('token', token);
    
    const url = `ws://localhost:8082/api/ws?${params.toString()}`;
    const ws = new WebSocket(url);
    
    ws.onopen = () => {
      setConnectionStatus('connected');
      console.log('🔌 WebSocket connecté');
    };
    
    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        setEvents(prev => [...prev.slice(-99), data]); // Garder les 100 derniers événements
        console.log('📨 Événement reçu:', data);
      } catch (error) {
        console.error('❌ Erreur parsing WebSocket:', error);
      }
    };
    
    ws.onclose = () => {
      setConnectionStatus('disconnected');
      console.log('🔌 WebSocket déconnecté');
    };
    
    ws.onerror = (error) => {
      console.error('❌ Erreur WebSocket:', error);
    };
    
    setSocket(ws);
  }, []);
  
  const sendCommand = useCallback((command: WebSocketCommand) => {
    if (socket && socket.readyState === WebSocket.OPEN) {
      socket.send(JSON.stringify(command));
      console.log('📤 Commande envoyée:', command);
    } else {
      console.warn('⚠️ WebSocket non connecté');
    }
  }, [socket]);
  
  const disconnect = useCallback(() => {
    if (socket) {
      socket.close();
      setSocket(null);
    }
  }, [socket]);
  
  return {
    connectionStatus,
    events,
    connect,
    sendCommand,
    disconnect,
  };
}

// Composant de lecteur audio avec streaming adaptatif
export function AdaptiveAudioPlayer({ trackId }: { trackId: string }) {
  const { createStreamingSession, getHLSMasterPlaylist, currentQuality } = useAdaptiveStreaming();
  const { connect, sendCommand, events, connectionStatus } = useWebSocket();
  const audioRef = useRef<HTMLAudioElement>(null);
  const [isPlaying, setIsPlaying] = useState(false);
  const [sessionId, setSessionId] = useState<string>('');
  
  useEffect(() => {
    // Créer la session de streaming
    createStreamingSession(trackId).then(setSessionId);
    
    // Connecter WebSocket
    connect('user_123', 'jwt_token_here');
  }, [trackId]);
  
  const handlePlay = () => {
    if (audioRef.current) {
      audioRef.current.play();
      setIsPlaying(true);
      
      // Envoyer commande WebSocket
      sendCommand({
        command: 'Play',
        command_id: `play_${Date.now()}`,
        track_id: trackId,
        start_position_ms: (audioRef.current.currentTime * 1000),
      });
    }
  };
  
  const handlePause = () => {
    if (audioRef.current) {
      audioRef.current.pause();
      setIsPlaying(false);
      
      sendCommand({
        command: 'Pause',
        command_id: `pause_${Date.now()}`,
        track_id: trackId,
      });
    }
  };
  
  return (
    <div className="adaptive-audio-player">
      <audio
        ref={audioRef}
        src={getHLSMasterPlaylist(trackId)}
        onTimeUpdate={() => {
          if (audioRef.current) {
            const progress = audioRef.current.currentTime * 1000;
            // Envoyer progression via WebSocket
            sendCommand({
              command: 'Seek',
              command_id: `progress_${Date.now()}`,
              track_id: trackId,
              position_ms: progress,
            });
          }
        }}
      />
      
      <div className="controls">
        <button onClick={isPlaying ? handlePause : handlePlay}>
          {isPlaying ? '⏸️' : '▶️'}
        </button>
        
        <div className="quality-info">
          Qualité: {currentQuality}
        </div>
        
        <div className="connection-status">
          WebSocket: {connectionStatus}
        </div>
      </div>
      
      <div className="events-log">
        <h4>Événements temps réel:</h4>
        {events.slice(-5).map((event, i) => (
          <div key={i} className="event">
            {JSON.stringify(event, null, 2)}
          </div>
        ))}
      </div>
    </div>
  );
}
```

Cette documentation complète du module streaming vous permet d'implémenter un système de streaming adaptatif sophistiqué avec communication temps réel pour votre plateforme audio. 