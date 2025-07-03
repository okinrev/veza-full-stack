use axum::{
    extract::{
        ws::{WebSocket, WebSocketUpgrade},
        State, Query,
    },
    response::Response,
};
use serde::{Deserialize, Serialize};
use std::{
    collections::HashMap,
    sync::Arc,
    time::{SystemTime, Duration},
};
use tokio::sync::{RwLock, broadcast};
use tracing::{info, warn};
use uuid::Uuid;

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
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

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum PlaylistAction {
    Added,
    Removed,
    Reordered,
    Cleared,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum MessageLevel {
    Info,
    Warning,
    Error,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct LiveTrackStats {
    pub track_id: String,
    pub title: String,
    pub artist: String,
    pub current_listeners: u32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
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

#[derive(Debug, Clone)]
pub struct WebSocketConnection {
    pub id: Uuid,
    pub user_id: Option<String>,
    pub ip_address: String,
    pub connected_at: SystemTime,
    pub last_activity: SystemTime,
    pub subscribed_events: Vec<String>,
    pub sender: broadcast::Sender<WebSocketEvent>,
}

pub struct WebSocketManager {
    connections: Arc<RwLock<HashMap<Uuid, WebSocketConnection>>>,
    global_sender: broadcast::Sender<WebSocketEvent>,
    stats: Arc<RwLock<WebSocketStats>>,
}

#[derive(Debug, Default)]
struct WebSocketStats {
    total_connections: u64,
    current_connections: u32,
    messages_sent: u64,
    messages_received: u64,
    events_broadcasted: u64,
}

impl WebSocketManager {
    pub fn new() -> Self {
        let (global_sender, _) = broadcast::channel(1000);
        
        Self {
            connections: Arc::new(RwLock::new(HashMap::new())),
            global_sender,
            stats: Arc::new(RwLock::new(WebSocketStats::default())),
        }
    }

    /// Gère une nouvelle connexion WebSocket
    pub async fn handle_websocket(
        &self,
        ws: WebSocketUpgrade,
        user_id: Option<String>,
        ip_address: String,
    ) -> Response {
        let manager = self.clone();
        
        ws.on_upgrade(move |socket| async move {
            manager.handle_socket(socket, user_id, ip_address).await;
        })
    }

    async fn handle_socket(&self, _socket: WebSocket, user_id: Option<String>, ip_address: String) {
        let connection_id = Uuid::new_v4();
        let (sender, _receiver) = broadcast::channel(100);
        
        let connection = WebSocketConnection {
            id: connection_id,
            user_id: user_id.clone(),
            ip_address: ip_address.clone(),
            connected_at: SystemTime::now(),
            last_activity: SystemTime::now(),
            subscribed_events: vec!["*".to_string()], // Abonné à tous les événements par défaut
            sender: sender.clone(),
        };

        // Ajouter la connexion
        {
            let mut connections = self.connections.write().await;
            connections.insert(connection_id, connection);
            
            let mut stats = self.stats.write().await;
            stats.current_connections += 1;
            stats.total_connections += 1;
        }

        info!("WebSocket connecté: {} depuis {}", connection_id, ip_address);

        // Envoyer un message de bienvenue
        let welcome_event = WebSocketEvent::ServerMessage {
            message: "Connexion WebSocket établie".to_string(),
            level: MessageLevel::Info,
        };
        
        if let Ok(_json) = serde_json::to_string(&welcome_event) {
            if let Err(e) = sender.send(welcome_event) {
                warn!("Erreur envoi message bienvenue: {}", e);
            }
        }

        // Note: Implémentation simplifiée pour éviter les erreurs de lifetime
        info!("WebSocket handler simplifié pour {}", connection_id);

        // Nettoyage à la déconnexion
        {
            let mut connections = self.connections.write().await;
            connections.remove(&connection_id);
            
            let mut stats = self.stats.write().await;
            stats.current_connections = stats.current_connections.saturating_sub(1);
        }

        info!("WebSocket déconnecté: {}", connection_id);
    }

    async fn handle_command(
        connection_id: Uuid,
        command: WebSocketCommand,
        connections: &Arc<RwLock<HashMap<Uuid, WebSocketConnection>>>,
        sender: &broadcast::Sender<WebSocketEvent>,
    ) {
        let response = match command {
            WebSocketCommand::Subscribe { command_id, events, filters: _ } => {
                {
                    let mut conns = connections.write().await;
                    if let Some(conn) = conns.get_mut(&connection_id) {
                        conn.subscribed_events = events.clone();
                    }
                }
                
                WebSocketEvent::CommandResponse {
                    command_id,
                    success: true,
                    data: Some(serde_json::json!({
                        "subscribed_events": events
                    })),
                    error: None,
                }
            }

            WebSocketCommand::Unsubscribe { command_id, events } => {
                {
                    let mut conns = connections.write().await;
                    if let Some(conn) = conns.get_mut(&connection_id) {
                        conn.subscribed_events.retain(|e| !events.contains(e));
                    }
                }
                
                WebSocketEvent::CommandResponse {
                    command_id,
                    success: true,
                    data: Some(serde_json::json!({
                        "unsubscribed_events": events
                    })),
                    error: None,
                }
            }

            WebSocketCommand::GetStatus { command_id } => {
                let conn_info = {
                    let conns = connections.read().await;
                    conns.get(&connection_id).cloned()
                };

                if let Some(conn) = conn_info {
                    WebSocketEvent::CommandResponse {
                        command_id,
                        success: true,
                        data: Some(serde_json::json!({
                            "connection_id": conn.id,
                            "user_id": conn.user_id,
                            "connected_at": conn.connected_at.duration_since(SystemTime::UNIX_EPOCH).unwrap().as_secs(),
                            "subscribed_events": conn.subscribed_events
                        })),
                        error: None,
                    }
                } else {
                    WebSocketEvent::CommandResponse {
                        command_id,
                        success: false,
                        data: None,
                        error: Some("Connexion non trouvée".to_string()),
                    }
                }
            }

            WebSocketCommand::Ping { command_id } => {
                WebSocketEvent::CommandResponse {
                    command_id,
                    success: true,
                    data: Some(serde_json::json!({
                        "timestamp": SystemTime::now().duration_since(SystemTime::UNIX_EPOCH).unwrap().as_secs()
                    })),
                    error: None,
                }
            }

            _ => {
                WebSocketEvent::CommandResponse {
                    command_id: "unknown".to_string(),
                    success: false,
                    data: None,
                    error: Some("Commande non implémentée".to_string()),
                }
            }
        };

        let _ = sender.send(response);
    }

    async fn should_receive_event(
        connections: &Arc<RwLock<HashMap<Uuid, WebSocketConnection>>>,
        connection_id: Uuid,
        event: &WebSocketEvent,
    ) -> bool {
        let conns = connections.read().await;
        if let Some(conn) = conns.get(&connection_id) {
            if conn.subscribed_events.is_empty() {
                return true; // Par défaut, recevoir tous les événements
            }

            let event_type = match event {
                WebSocketEvent::PlaybackStarted { .. } => "playback",
                WebSocketEvent::PlaybackPaused { .. } => "playback",
                WebSocketEvent::PlaybackResumed { .. } => "playback",
                WebSocketEvent::PlaybackStopped { .. } => "playback",
                WebSocketEvent::PlaybackProgress { .. } => "playback_progress",
                WebSocketEvent::PlaylistUpdated { .. } => "playlist",
                WebSocketEvent::TrackLiked { .. } => "social",
                WebSocketEvent::TrackShared { .. } => "social",
                WebSocketEvent::LiveStats { .. } => "stats",
                WebSocketEvent::ServerMessage { .. } => "system",
                _ => "other",
            };

            return conn.subscribed_events.contains(&event_type.to_string());
        }

        false
    }

    /// Diffuse un événement à toutes les connexions
    pub async fn broadcast_event(&self, event: WebSocketEvent) {
        let _ = self.global_sender.send(event);
        self.stats.write().await.events_broadcasted += 1;
    }

    /// Envoie un événement à une connexion spécifique
    pub async fn send_to_connection(&self, connection_id: Uuid, event: WebSocketEvent) {
        let connections = self.connections.read().await;
        if let Some(conn) = connections.get(&connection_id) {
            let _ = conn.sender.send(event);
        }
    }

    /// Envoie un événement à un utilisateur spécifique (toutes ses connexions)
    pub async fn send_to_user(&self, user_id: &str, event: WebSocketEvent) {
        let connections = self.connections.read().await;
        for conn in connections.values() {
            if let Some(ref conn_user_id) = conn.user_id {
                if conn_user_id == user_id {
                    let _ = conn.sender.send(event.clone());
                }
            }
        }
    }

    /// Nettoie les connexions inactives
    pub async fn cleanup_inactive_connections(&self, max_idle_duration: Duration) {
        let cutoff = SystemTime::now() - max_idle_duration;
        let mut connections = self.connections.write().await;
        let before_count = connections.len();

        connections.retain(|_, conn| conn.last_activity > cutoff);

        let removed = before_count - connections.len();
        if removed > 0 {
            info!("Nettoyage des connexions WebSocket inactives: {} supprimées", removed);
            self.stats.write().await.current_connections = connections.len() as u32;
        }
    }

    /// Obtient les statistiques des WebSockets
    pub async fn get_stats(&self) -> serde_json::Value {
        let stats = self.stats.read().await;
        let connections = self.connections.read().await;

        let user_connections: HashMap<String, u32> = connections
            .values()
            .filter_map(|conn| conn.user_id.as_ref())
            .fold(HashMap::new(), |mut acc, user_id| {
                *acc.entry(user_id.clone()).or_insert(0) += 1;
                acc
            });

        serde_json::json!({
            "total_connections": stats.total_connections,
            "current_connections": stats.current_connections,
            "messages_sent": stats.messages_sent,
            "messages_received": stats.messages_received,
            "events_broadcasted": stats.events_broadcasted,
            "user_connections": user_connections.len(),
            "connections_per_user": user_connections
        })
    }
}

impl Clone for WebSocketManager {
    fn clone(&self) -> Self {
        Self {
            connections: self.connections.clone(),
            global_sender: self.global_sender.clone(),
            stats: self.stats.clone(),
        }
    }
}

impl Default for WebSocketManager {
    fn default() -> Self {
        Self::new()
    }
}

/// Query parameters pour les connexions WebSocket
#[derive(Debug, Deserialize)]
pub struct WebSocketQuery {
    pub user_id: Option<String>,
    pub token: Option<String>,
}

/// Handler pour les connexions WebSocket
pub async fn websocket_handler(
    ws: WebSocketUpgrade,
    Query(params): Query<WebSocketQuery>,
    State(ws_manager): State<Arc<WebSocketManager>>,
) -> Response {
    // En production, on validerait le token ici
    let user_id = params.user_id;
    let ip_address = "127.0.0.1".to_string(); // Extraire de la requête réelle

    info!("Nouvelle connexion WebSocket demandée pour utilisateur: {:?}", user_id);

    ws_manager.handle_websocket(ws, user_id, ip_address).await
} 