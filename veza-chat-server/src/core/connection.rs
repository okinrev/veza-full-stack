//! Connection Manager Production-Ready
//! 
//! Gestionnaire de connexions optimisé pour 100k+ WebSocket simultanées
//! avec zero-copy broadcasting et métriques en temps réel.

use std::sync::Arc;
use std::time::Duration;
use std::collections::HashSet;
use dashmap::DashMap;
use tokio::sync::{RwLock, broadcast};
use uuid::Uuid;
use bytes::Bytes;
use serde::{Serialize, Deserialize};
use tracing::{info, warn, error, debug};
use chrono::{DateTime, Utc};

use crate::error::ChatError;

/// Gestionnaire principal des connexions WebSocket
/// Optimisé pour 100k+ connexions simultanées
#[derive(Debug, Clone)]
pub struct ConnectionManager {
    /// Connexions actives indexées par ID
    connections: Arc<DashMap<Uuid, UserConnection>>,
    
    /// Salles de chat avec leurs membres
    rooms: Arc<DashMap<String, Room>>,
    
    /// Broadcaster pour diffusion efficace
    broadcaster: Arc<BroadcastOptimizer>,
    
    /// Configuration
    config: Arc<ConnectionConfig>,
}

/// Configuration du gestionnaire de connexions
#[derive(Debug, Clone)]
pub struct ConnectionConfig {
    /// Nombre maximum de connexions simultanées
    pub max_connections: usize,
    
    /// Timeout d'inactivité avant déconnexion
    pub idle_timeout: Duration,
    
    /// Taille du buffer de diffusion
    pub broadcast_buffer_size: usize,
    
    /// Limite de messages par seconde par connexion
    pub rate_limit_per_second: u32,
}

impl Default for ConnectionConfig {
    fn default() -> Self {
        Self {
            max_connections: 100_000,  // 100k connexions
            idle_timeout: Duration::from_secs(300),  // 5 minutes
            broadcast_buffer_size: 1024,
            rate_limit_per_second: 10,
        }
    }
}

/// Connexion utilisateur individuelle
pub struct UserConnection {
    /// Identifiant unique de la connexion
    pub id: Uuid,
    
    /// ID de l'utilisateur connecté
    pub user_id: i64,
    
    /// Sender pour envoyer des messages à ce client
    pub sender: broadcast::Sender<Bytes>,
    
    /// Rate limiter individuel
    pub rate_limiter: Arc<RateLimiter>,
    
    /// Dernière activité
    pub last_activity: DateTime<Utc>,
    
    /// Salles auxquelles l'utilisateur est abonné
    pub subscriptions: HashSet<String>,
    
    /// Métadonnées de connexion
    pub metadata: ConnectionMetadata,
}

/// Métadonnées de connexion
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ConnectionMetadata {
    pub ip_address: String,
    pub user_agent: String,
    pub connected_at: DateTime<Utc>,
    pub platform: String,
}

pub use super::room::*;

/// Rate limiter par connexion
pub struct RateLimiter {
    tokens: Arc<parking_lot::Mutex<f64>>,
    last_refill: Arc<parking_lot::Mutex<DateTime<Utc>>>,
    rate: f64,
    burst: f64,
}

/// Optimiseur de diffusion zero-copy
pub struct BroadcastOptimizer {
    /// Cache de messages pré-sérialisés
    message_cache: Arc<DashMap<String, Bytes>>,
    
    /// Groupes de connexions pour routage efficace
    connection_groups: Arc<DashMap<String, Vec<broadcast::Sender<Bytes>>>>,
}

impl ConnectionManager {
    /// Crée un nouveau gestionnaire de connexions
    pub fn new(config: ConnectionConfig) -> Self {
        Self {
            connections: Arc::new(DashMap::new()),
            rooms: Arc::new(DashMap::new()),
            broadcaster: Arc::new(BroadcastOptimizer::new()),
            config: Arc::new(config),
        }
    }

    /// Ajoute une nouvelle connexion
    pub async fn add_connection(
        &self,
        user_id: i64,
        metadata: ConnectionMetadata,
    ) -> Result<(Uuid, broadcast::Receiver<Bytes>), ChatError> {
        // Vérifier la limite de connexions
        if self.connections.len() >= self.config.max_connections {
            return Err(ChatError::configuration_error("Maximum connections reached"));
        }

        let connection_id = Uuid::new_v4();
        let (sender, receiver) = broadcast::channel(self.config.broadcast_buffer_size);
        
        let connection = UserConnection {
            id: connection_id,
            user_id,
            sender,
            rate_limiter: Arc::new(RateLimiter::new(
                self.config.rate_limit_per_second as f64,
                10.0, // burst
            )),
            last_activity: Utc::now(),
            subscriptions: HashSet::new(),
            metadata,
        };

        self.connections.insert(connection_id, connection);
        
        info!(
            connection_id = %connection_id,
            user_id = user_id,
            total_connections = self.connections.len(),
            "🔌 Nouvelle connexion établie"
        );

        Ok((connection_id, receiver))
    }

    /// Diffuse un message à une salle avec parallélisation rayon
    pub async fn broadcast_to_room(
        &self,
        room_id: &str,
        message: Bytes,
    ) -> Result<usize, ChatError> {
        let start = Utc::now();
        let mut sent_count = 0;

        if let Some(room) = self.rooms.get(room_id) {
            // Utiliser rayon pour diffusion parallèle optimisée
            use rayon::prelude::*;
            
            let member_ids: Vec<Uuid> = room.members.iter()
                .map(|entry| *entry.key())
                .collect();

            sent_count = member_ids.par_iter()
                .map(|&connection_id| {
                    if let Some(connection) = self.connections.get(&connection_id) {
                        match connection.sender.send(message.clone()) {
                            Ok(_) => 1,
                            Err(_) => 0
                        }
                    } else {
                        0
                    }
                })
                .sum();
        }

        let duration = Utc::now().signed_duration_since(start).num_milliseconds() as u128;
        debug!(
            room_id = room_id,
            recipients = sent_count,
            duration_ms = duration.as_millis(),
            "📡 Message diffusé"
        );

        Ok(sent_count)
    }

    /// Statistiques en temps réel
    pub fn get_stats(&self) -> ConnectionStats {
        ConnectionStats {
            active_connections: self.connections.len(),
            active_rooms: self.rooms.len(),
            total_members: self.rooms.iter()
                .map(|room| room.members.len())
                .sum(),
        }
    }
}

/// Statistiques de connexion
#[derive(Debug, Serialize)]
pub struct ConnectionStats {
    pub active_connections: usize,
    pub active_rooms: usize,
    pub total_members: usize,
}

impl BroadcastOptimizer {
    pub fn new() -> Self {
        Self {
            message_cache: Arc::new(DashMap::new()),
            connection_groups: Arc::new(DashMap::new()),
        }
    }
}

impl RateLimiter {
    pub fn new(rate: f64, burst: f64) -> Self {
        Self {
            tokens: Arc::new(parking_lot::Mutex::new(burst)),
            last_refill: Arc::new(parking_lot::Mutex::new(Utc::now())),
            rate,
            burst,
        }
    }

    pub fn check_rate_limit(&self) -> bool {
        let now = Utc::now();
        let mut tokens = self.tokens.lock();
        let mut last_refill = self.last_refill.lock();

        // Token bucket algorithm
        let elapsed = now.signed_duration_since(*last_refill).num_seconds() as f64;
        *tokens = (*tokens + elapsed * self.rate).min(self.burst);
        *last_refill = now;

        if *tokens >= 1.0 {
            *tokens -= 1.0;
            true
        } else {
            false
        }
    }
}
