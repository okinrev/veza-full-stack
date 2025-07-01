//! Room Management pour Chat Production
//! 
//! Gestion des salles de chat avec permissions Discord-like
//! et optimisations pour haute performance.

use std::sync::Arc;
use dashmap::DashMap;
use tokio::sync::RwLock;
use uuid::Uuid;
use serde::{Serialize, Deserialize};
use chrono::{DateTime, Utc};

use super::message::*;
use super::user::*;

/// Salle de chat optimisée
pub struct Room {
    /// Identifiant de la salle
    pub id: String,
    
    /// Nom de la salle
    pub name: String,
    
    /// Membres connectés
    pub members: Arc<DashMap<Uuid, RoomMember>>,
    
    /// Configuration de la salle
    pub settings: RoomSettings,
    
    /// Buffer circulaire pour messages récents
    pub message_buffer: Arc<RwLock<MessageBuffer>>,
    
    /// Tracker de présence
    pub presence_tracker: Arc<PresenceTracker>,
}

/// Membre d'une salle
#[derive(Debug, Clone)]
pub struct RoomMember {
    pub connection_id: Uuid,
    pub user_id: i64,
    pub joined_at: DateTime<Utc>,
    pub permissions: RoomPermissions,
    pub status: PresenceStatus,
}

/// Permissions dans une salle (Discord-like)
#[derive(Debug, Clone, PartialEq)]
pub struct RoomPermissions {
    // Permissions générales
    pub view_channel: bool,
    pub send_messages: bool,
    pub embed_links: bool,
    pub attach_files: bool,
    pub read_message_history: bool,
    pub mention_everyone: bool,
    pub use_external_emojis: bool,
    pub add_reactions: bool,
    
    // Permissions modération
    pub manage_messages: bool,
    pub manage_channel: bool,
    pub kick_members: bool,
    pub ban_members: bool,
    
    // Permissions voix
    pub connect_voice: bool,
    pub speak: bool,
    pub mute_members: bool,
    pub move_members: bool,
}

impl Default for RoomPermissions {
    fn default() -> Self {
        Self {
            view_channel: true,
            send_messages: true,
            embed_links: true,
            attach_files: true,
            read_message_history: true,
            mention_everyone: false,
            use_external_emojis: true,
            add_reactions: true,
            manage_messages: false,
            manage_channel: false,
            kick_members: false,
            ban_members: false,
            connect_voice: true,
            speak: true,
            mute_members: false,
            move_members: false,
        }
    }
}

/// Configuration d'une salle
#[derive(Debug, Clone)]
pub struct RoomSettings {
    pub is_public: bool,
    pub max_members: Option<usize>,
    pub rate_limit: Option<u32>,
    pub enable_file_upload: bool,
    pub enable_voice: bool,
    pub channel_type: ChannelType,
    pub topic: Option<String>,
    pub slow_mode: Option<u32>, // secondes entre messages
}

/// Type de channel Discord-like
#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub enum ChannelType {
    Text,
    Voice,
    Announcement,
    Stage,
    Forum,
    Category,
}

impl Default for RoomSettings {
    fn default() -> Self {
        Self {
            is_public: true,
            max_members: None,
            rate_limit: Some(10),
            enable_file_upload: true,
            enable_voice: false,
            channel_type: ChannelType::Text,
            topic: None,
            slow_mode: None,
        }
    }
}

/// Buffer circulaire pour messages récents
pub struct MessageBuffer {
    messages: Vec<StoredMessage>,
    capacity: usize,
    index: usize,
}

impl MessageBuffer {
    pub fn new(capacity: usize) -> Self {
        Self {
            messages: Vec::with_capacity(capacity),
            capacity,
            index: 0,
        }
    }

    pub fn add_message(&mut self, message: StoredMessage) {
        if self.messages.len() < self.capacity {
            self.messages.push(message);
        } else {
            self.messages[self.index] = message;
            self.index = (self.index + 1) % self.capacity;
        }
    }

    pub fn get_recent_messages(&self, limit: usize) -> Vec<&StoredMessage> {
        let len = self.messages.len().min(limit);
        if self.messages.len() < self.capacity {
            self.messages.iter().rev().take(len).collect()
        } else {
            let mut result = Vec::with_capacity(len);
            for i in 0..len {
                let idx = (self.index + self.capacity - 1 - i) % self.capacity;
                result.push(&self.messages[idx]);
            }
            result
        }
    }
}

impl Room {
    pub fn new(id: String, name: String, settings: RoomSettings) -> Self {
        Self {
            id,
            name,
            members: Arc::new(DashMap::new()),
            settings,
            message_buffer: Arc::new(RwLock::new(MessageBuffer::new(1000))),
            presence_tracker: Arc::new(PresenceTracker::new()),
        }
    }

    /// Ajoute un membre à la salle
    pub async fn add_member(
        &self,
        connection_id: Uuid,
        user_id: i64,
        permissions: RoomPermissions,
    ) -> Result<(), &'static str> {
        if let Some(max) = self.settings.max_members {
            if self.members.len() >= max {
                return Err("Room is full");
            }
        }

        let member = RoomMember {
            connection_id,
            user_id,
            joined_at: Utc::now(),
            permissions,
            status: PresenceStatus::Online,
        };

        self.members.insert(connection_id, member);
        self.presence_tracker.update_status(user_id, PresenceStatus::Online);

        Ok(())
    }

    /// Retire un membre de la salle
    pub async fn remove_member(&self, connection_id: Uuid) {
        if let Some((_, member)) = self.members.remove(&connection_id) {
            // Vérifier si c'était la dernière connexion de cet utilisateur
            let user_still_connected = self.members.iter()
                .any(|entry| entry.value().user_id == member.user_id);
            
            if !user_still_connected {
                self.presence_tracker.update_status(member.user_id, PresenceStatus::Invisible);
            }
        }
    }

    /// Vérifie les permissions d'un membre
    pub fn check_permission(
        &self,
        connection_id: Uuid,
        permission: fn(&RoomPermissions) -> bool,
    ) -> bool {
        self.members.get(&connection_id)
            .map(|member| permission(&member.permissions))
            .unwrap_or(false)
    }
}
