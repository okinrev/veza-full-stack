//file: backend/modules/chat_server/src/hub/common.rs

use std::collections::HashMap;
use std::sync::Arc;
use std::time::{Duration, Instant};
use tokio::sync::RwLock;
use sqlx::PgPool;

use crate::client::Client;
// use crate::rate_limiter::RateLimiter;
use crate::config::ServerConfig;
// use crate::cache::CacheManager;
// use crate::monitoring::ChatMetrics;
// use crate::moderation::ModerationSystem;
// use crate::presence::PresenceManager;
// use crate::authentication::UserSession;

// Types temporaires pour la compilation
#[derive(Debug, Clone)]
pub struct UserSession {
    pub user_id: i32,
    pub username: String,
}

// Commenté car le ReactionManager n'est pas encore disponible
// use crate::hub::reactions::ReactionManager;

pub struct ChatHub {
    pub clients: Arc<RwLock<HashMap<i32, Client>>>,
    pub rooms: Arc<RwLock<HashMap<String, Vec<i32>>>>,
    pub db: PgPool,
    // pub rate_limiter: RateLimiter,
    pub config: ServerConfig,
    pub stats: Arc<RwLock<HubStats>>,
    
    // Nouveaux systèmes intégrés (initialisés séparément)
    // pub cache: CacheManager,
    // pub metrics: ChatMetrics,
    // pub presence: PresenceManager,
    // pub connections: Arc<RwLock<HashMap<i32, UserSession>>>,
    // pub moderation: ModerationSystem,
    // pub reactions: ReactionManager, // Commenté temporairement
}

#[derive(Debug, Default, Clone)]
pub struct HubStats {
    pub total_connections: u64,
    pub active_connections: u64,
    pub total_messages: u64,
    pub total_rooms_created: u64,
    pub uptime_start: Option<Instant>,
}

impl HubStats {
    pub fn new() -> Self {
        Self {
            uptime_start: Some(Instant::now()),
            ..Default::default()
        }
    }

    pub fn uptime(&self) -> Duration {
        self.uptime_start.map_or(Duration::ZERO, |start| start.elapsed())
    }
}

impl ChatHub {
    pub fn new(db: PgPool) -> Self {
        let config = ServerConfig::default();
        Self {
            db,
            clients: Arc::new(RwLock::new(HashMap::new())),
            rooms: Arc::new(RwLock::new(HashMap::new())),
            // rate_limiter: RateLimiter::new(config.limits.max_messages_per_minute),
            config,
            stats: Arc::new(RwLock::new(HubStats::new())),
            
            // Initialisation des nouveaux systèmes
            // cache: CacheManager::new(),
            // metrics: ChatMetrics::new(),
            // presence,
            // connections: Arc::new(RwLock::new(HashMap::new())),
            // moderation,
            // reactions: ReactionManager::new(), // Commenté temporairement
        }
    }

    pub async fn register(&self, user_id: i32, client: Client) {
        tracing::debug!(user_id = %user_id, username = %client.username, "🔧 Début register");
        
        let mut clients = self.clients.write().await;
        let clients_before = clients.len();
        
        clients.insert(user_id, client);

        // Mise à jour des statistiques
        let mut stats = self.stats.write().await;
        stats.total_connections += 1;
        stats.active_connections = clients.len() as u64;
        
        tracing::info!(
            user_id = %user_id, 
            clients_before = %clients_before, 
            clients_after = %clients.len(), 
            total_connections = %stats.total_connections,
            "👤 Enregistrement du client"
        );
    }

    pub async fn unregister(&self, user_id: i32) {
        tracing::debug!(user_id = %user_id, "🔧 Début unregister");
        
        let mut clients = self.clients.write().await;
        let clients_before = clients.len();
        
        if let Some(removed_client) = clients.remove(&user_id) {
            // Mise à jour des statistiques
            let mut stats = self.stats.write().await;
            stats.active_connections = clients.len() as u64;
            
            tracing::info!(
                user_id = %user_id, 
                username = %removed_client.username, 
                clients_before = %clients_before, 
                clients_after = %clients.len(),
                active_connections = %stats.active_connections,
                connection_duration = ?removed_client.connection_duration(),
                "🚪 Déconnexion du client"
            );
        } else {
            tracing::warn!(user_id = %user_id, clients_count = %clients.len(), "⚠️ Tentative de déconnexion d'un client non enregistré");
        }
        
        // Nettoyer les salons
        let mut rooms = self.rooms.write().await;
        let mut rooms_cleaned = 0;
        let mut total_removals = 0;
        
        for (room_name, user_list) in rooms.iter_mut() {
            let before_len = user_list.len();
            user_list.retain(|&id| id != user_id);
            let after_len = user_list.len();
            
            if before_len != after_len {
                total_removals += before_len - after_len;
                rooms_cleaned += 1;
                tracing::debug!(user_id = %user_id, room = %room_name, members_before = %before_len, members_after = %after_len, "🧹 Utilisateur retiré du salon");
            }
        }
        
        if rooms_cleaned > 0 {
            tracing::info!(user_id = %user_id, rooms_cleaned = %rooms_cleaned, total_removals = %total_removals, "🧹 Nettoyage des salons terminé");
        } else {
            tracing::debug!(user_id = %user_id, "🧹 Aucun salon à nettoyer");
        }
    }

    /// Vérifie le rate limiting pour un utilisateur
    pub async fn check_rate_limit(&self, _user_id: i32) -> bool {
        // self.rate_limiter.check_and_update(user_id).await
        false
    }

    /// Incrémente le compteur de messages
    pub async fn increment_message_count(&self) {
        let mut stats = self.stats.write().await;
        stats.total_messages += 1;
    }

    /// Retourne les statistiques du hub
    pub async fn get_stats(&self) -> HubStats {
        self.stats.read().await.clone()
    }

    /// Nettoie les connexions mortes (heartbeat timeout)
    pub async fn cleanup_dead_connections(&self) {
        let timeout = Duration::from_secs(self.config.server.heartbeat_interval.as_secs() as u64 * 3); // 3x heartbeat interval
        let mut dead_clients = Vec::new();
        
        {
            let clients = self.clients.read().await;
            for (user_id, client) in clients.iter() {
                if !client.is_alive(timeout) {
                    dead_clients.push(*user_id);
                }
            }
        }

        for user_id in dead_clients {
            tracing::warn!(user_id = %user_id, timeout_seconds = %timeout.as_secs(), "💀 Connexion morte détectée, nettoyage");
            self.unregister(user_id).await;
        }
    }

    /// Envoie un ping à tous les clients connectés
    pub async fn ping_all_clients(&self) {
        let clients = self.clients.read().await;
        let mut successful_pings = 0;
        let mut failed_pings = 0;

        for client in clients.values() {
            if client.send_ping() {
                successful_pings += 1;
            } else {
                failed_pings += 1;
            }
        }

        if failed_pings > 0 {
            tracing::warn!(successful_pings = %successful_pings, failed_pings = %failed_pings, "🏓 Ping terminé avec des échecs");
        } else {
            tracing::debug!(successful_pings = %successful_pings, "🏓 Ping de tous les clients réussi");
        }
    }

    /// Ajoute une connexion utilisateur
    pub async fn add_connection(&self, _user_id: i32, _session: UserSession) {
        // let mut connections = self.connections.write().await;
        // connections.insert(user_id, session);
    }

    /// Supprime une connexion utilisateur
    pub async fn remove_connection(&self, _user_id: i32) {
        // let mut connections = self.connections.write().await;
        // connections.remove(&user_id);
    }

    /// Vérifie si un utilisateur est connecté
    pub async fn is_user_connected(&self, _user_id: i32) -> bool {
        // let connections = self.connections.read().await;
        // connections.contains_key(&user_id)
        false
    }

    /// Ajoute un utilisateur à un salon
    pub async fn add_user_to_room(&self, room: &str, user_id: i32) {
        let mut rooms = self.rooms.write().await;
        rooms.entry(room.to_string()).or_insert_with(Vec::new).push(user_id);
    }

    /// Supprime un utilisateur d'un salon
    pub async fn remove_user_from_room(&self, room: &str, user_id: i32) {
        let mut rooms = self.rooms.write().await;
        if let Some(users) = rooms.get_mut(room) {
            users.retain(|&id| id != user_id);
            if users.is_empty() {
                rooms.remove(room);
            }
        }
    }

    /// Récupère les utilisateurs d'un salon
    pub async fn get_room_users(&self, room: &str) -> Vec<i32> {
        let rooms = self.rooms.read().await;
        rooms.get(room).cloned().unwrap_or_default()
    }

    /// Diffuse un message à tous les utilisateurs d'un salon
    pub async fn broadcast_to_room(&self, room: &str, _message: &str, exclude_user: Option<i32>) {
        let users = self.get_room_users(room).await;
        // let connections = self.connections.read().await;

        for user_id in users {
            if let Some(excluded) = exclude_user {
                if user_id == excluded {
                    continue;
                }
            }

            // if let Some(session) = connections.get(&user_id) {
            //     // Ici on devrait envoyer le message via WebSocket
            //     // Pour l'instant on fait juste un log
            //     tracing::info!(
            //         user_id = %user_id,
            //         room = %room,
            //         message = %message,
            //         "📡 Message diffusé"
            //     );
            // }
        }
    }
}
