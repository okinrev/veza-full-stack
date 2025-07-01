//! Système de channels Discord-like avancé
//! 
//! Ce module implémente un système de channels complet avec :
//! - Types de channels variés (Text, Voice, Stage, Forum, etc.)
//! - Permissions granulaires par rôle et utilisateur
//! - Catégories et organisation hiérarchique
//! - Support vocal avec gestion des membres connectés
//! - Slow mode et limitations
//! - Statistiques détaillées

use std::collections::{HashMap, HashSet};
use std::sync::Arc;
use serde::{Deserialize, Serialize};
use chrono::{DateTime, Utc};
use dashmap::DashMap;
use uuid::Uuid;

use crate::permissions::{Permission, UserPermissions};
use crate::error::{ChatError, Result};

/// Types de channels disponibles (Discord-like)
#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub enum ChannelType {
    /// Channel texte standard
    Text,
    /// Channel vocal
    Voice,
    /// Channel d'annonces (un seul sens)
    Announcement,
    /// Channel stage pour events
    Stage,
    /// Channel forum avec threads
    Forum,
    /// Channel de news
    News,
    /// Channel privé (DM)
    DirectMessage,
}

/// Configuration d'un channel
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ChannelConfig {
    /// Nom du channel
    pub name: String,
    /// Description/Topic
    pub topic: Option<String>,
    /// Type de channel
    pub channel_type: ChannelType,
    /// NSFW ?
    pub nsfw: bool,
    /// Slow mode (secondes entre messages)
    pub slowmode_delay: Option<u32>,
    /// Bitrate pour les channels vocaux (kbps)
    pub bitrate: Option<u32>,
    /// Limite d'utilisateurs pour channels vocaux
    pub user_limit: Option<u32>,
    /// Position dans la liste
    pub position: u32,
    /// ID de la catégorie parent
    pub parent_id: Option<String>,
}

/// Permissions spécifiques à un channel
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ChannelPermissions {
    /// Permissions par rôle
    pub role_permissions: HashMap<String, HashSet<ChannelPermission>>,
    /// Permissions par utilisateur (overrides)
    pub user_permissions: HashMap<i64, HashSet<ChannelPermission>>,
}

/// Permissions granulaires pour les channels
#[derive(Debug, Clone, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub enum ChannelPermission {
    // Permissions générales
    ViewChannel,
    ManageChannel,
    ManagePermissions,
    CreateInvite,
    
    // Messages
    SendMessages,
    SendTTSMessages,
    ManageMessages,
    EmbedLinks,
    AttachFiles,
    ReadMessageHistory,
    MentionEveryone,
    UseExternalEmojis,
    UseExternalStickers,
    AddReactions,
    UseSlashCommands,
    UseThreads,
    CreatePublicThreads,
    CreatePrivateThreads,
    SendMessagesInThreads,
    
    // Vocal
    Connect,
    Speak,
    MuteMembers,
    DeafenMembers,
    MoveMembers,
    UseVoiceActivity,
    Priorityspeaker,
    Stream,
    UseEmbeddedActivities,
    UseSoundboard,
    
    // Avancé
    ManageWebhooks,
    ManageEvents,
    RequestToSpeak,
}

/// Structure d'un channel Discord-like
#[derive(Debug, Clone, Serialize)]
pub struct Channel {
    pub id: String,
    pub config: ChannelConfig,
    pub permissions: ChannelPermissions,
    pub created_at: DateTime<Utc>,
    pub last_message_id: Option<String>,
    pub last_activity: DateTime<Utc>,
    
    /// Membres connectés (pour channels vocaux)
    #[serde(skip)]
    pub connected_members: Arc<DashMap<i64, VoiceMember>>,
    
    /// Statistiques du channel
    pub stats: ChannelStats,
}

/// Membre connecté à un channel vocal
#[derive(Debug, Clone, Serialize)]
pub struct VoiceMember {
    pub user_id: i64,
    pub username: String,
    pub joined_at: DateTime<Utc>,
    pub is_muted: bool,
    pub is_deafened: bool,
    pub is_streaming: bool,
    pub is_camera_on: bool,
}

/// Statistiques d'un channel
#[derive(Debug, Default, Clone, Serialize)]
pub struct ChannelStats {
    pub total_messages: u64,
    pub total_members: u64,
    pub active_members_today: u64,
    pub peak_concurrent_users: u64,
    pub last_peak_at: Option<DateTime<Utc>>,
}

/// Gestionnaire de channels
#[derive(Debug)]
pub struct ChannelManager {
    /// Channels par ID
    channels: Arc<DashMap<String, Channel>>,
    /// Index des channels par serveur
    server_channels: Arc<DashMap<String, HashSet<String>>>,
    /// Catégories
    categories: Arc<DashMap<String, ChannelCategory>>,
}

/// Catégorie de channels
#[derive(Debug, Clone, Serialize)]
pub struct ChannelCategory {
    pub id: String,
    pub name: String,
    pub position: u32,
    pub server_id: String,
    pub created_at: DateTime<Utc>,
}

impl ChannelManager {
    pub fn new() -> Self {
        Self {
            channels: Arc::new(DashMap::new()),
            server_channels: Arc::new(DashMap::new()),
            categories: Arc::new(DashMap::new()),
        }
    }
    
    /// Crée un nouveau channel
    pub async fn create_channel(
        &self,
        server_id: &str,
        config: ChannelConfig,
        _creator_id: i64,
        creator_permissions: &UserPermissions,
    ) -> Result<String> {
        // Vérifier les permissions
        if !creator_permissions.has_permission(&Permission::ManageChannels) {
            return Err(ChatError::unauthorized_simple("insufficient_permissions"));
        }
        
        let channel_id = format!("ch_{}", Uuid::new_v4());
        
        let channel = Channel {
            id: channel_id.clone(),
            config,
            permissions: ChannelPermissions {
                role_permissions: HashMap::new(),
                user_permissions: HashMap::new(),
            },
            created_at: Utc::now(),
            last_message_id: None,
            last_activity: Utc::now(),
            connected_members: Arc::new(DashMap::new()),
            stats: ChannelStats::default(),
        };
        
        // Ajouter le channel
        self.channels.insert(channel_id.clone(), channel);
        
        // Indexer par serveur
        self.server_channels
            .entry(server_id.to_string())
            .or_insert_with(HashSet::new)
            .insert(channel_id.clone());
        
        Ok(channel_id)
    }
    
    /// Vérifie si un utilisateur peut voir un channel
    pub fn can_view_channel(&self, channel_id: &str, user_permissions: &UserPermissions) -> bool {
        if let Some(channel) = self.channels.get(channel_id) {
            self.check_channel_permission(
                &channel.permissions,
                user_permissions,
                &ChannelPermission::ViewChannel,
            )
        } else {
            false
        }
    }
    
    /// Vérifie si un utilisateur peut envoyer des messages dans un channel
    pub fn can_send_messages(&self, channel_id: &str, user_permissions: &UserPermissions) -> bool {
        if let Some(channel) = self.channels.get(channel_id) {
            self.check_channel_permission(
                &channel.permissions,
                user_permissions,
                &ChannelPermission::SendMessages,
            )
        } else {
            false
        }
    }
    
    /// Joint un utilisateur à un channel vocal
    pub async fn join_voice_channel(
        &self,
        channel_id: &str,
        user_id: i64,
        username: String,
        user_permissions: &UserPermissions,
    ) -> Result<()> {
        let channel = self.channels.get(channel_id)
            .ok_or_else(|| ChatError::not_found_simple("channel_not_found"))?;
        
        // Vérifier le type de channel
        if !matches!(channel.config.channel_type, ChannelType::Voice | ChannelType::Stage) {
            return Err(ChatError::validation_error("not_voice_channel"));
        }
        
        // Vérifier les permissions
        if !self.check_channel_permission(
            &channel.permissions,
            user_permissions,
            &ChannelPermission::Connect,
        ) {
            return Err(ChatError::unauthorized_simple("cannot_connect"));
        }
        
        // Vérifier la limite d'utilisateurs
        if let Some(limit) = channel.config.user_limit {
            if channel.connected_members.len() >= limit as usize {
                return Err(ChatError::validation_error("channel_full"));
            }
        }
        
        let voice_member = VoiceMember {
            user_id,
            username,
            joined_at: Utc::now(),
            is_muted: false,
            is_deafened: false,
            is_streaming: false,
            is_camera_on: false,
        };
        
        channel.connected_members.insert(user_id, voice_member);
        
        Ok(())
    }
    
    /// Quitte un channel vocal
    pub async fn leave_voice_channel(&self, channel_id: &str, user_id: i64) -> Result<()> {
        if let Some(channel) = self.channels.get(channel_id) {
            channel.connected_members.remove(&user_id);
        }
        Ok(())
    }
    
    /// Met à jour les permissions d'un channel
    pub async fn update_channel_permissions(
        &self,
        channel_id: &str,
        target_type: PermissionTargetType,
        target_id: String,
        permissions: HashSet<ChannelPermission>,
        user_permissions: &UserPermissions,
    ) -> Result<()> {
        if !user_permissions.has_permission(&Permission::ManageChannels) {
            return Err(ChatError::unauthorized_simple("insufficient_permissions"));
        }
        
        if let Some(mut channel) = self.channels.get_mut(channel_id) {
            match target_type {
                PermissionTargetType::Role => {
                    channel.permissions.role_permissions.insert(target_id, permissions);
                }
                PermissionTargetType::User => {
                    if let Ok(user_id) = target_id.parse::<i64>() {
                        channel.permissions.user_permissions.insert(user_id, permissions);
                    }
                }
            }
        }
        
        Ok(())
    }
    
    /// Active le slow mode sur un channel
    pub async fn set_slowmode(
        &self,
        channel_id: &str,
        delay_seconds: Option<u32>,
        user_permissions: &UserPermissions,
    ) -> Result<()> {
        if !user_permissions.has_permission(&Permission::ManageChannels) {
            return Err(ChatError::unauthorized_simple("insufficient_permissions"));
        }
        
        if let Some(mut channel) = self.channels.get_mut(channel_id) {
            channel.config.slowmode_delay = delay_seconds;
        }
        
        Ok(())
    }
    
    /// Obtient les channels d'un serveur organisés par catégories
    pub fn get_server_channels(&self, server_id: &str) -> Vec<ChannelWithCategory> {
        let mut result = Vec::new();
        
        if let Some(channel_ids) = self.server_channels.get(server_id) {
            for channel_id in channel_ids.iter() {
                if let Some(channel) = self.channels.get(channel_id) {
                    let category = channel.config.parent_id.as_ref()
                        .and_then(|id| self.categories.get(id))
                        .map(|cat| cat.value().clone());
                    
                    result.push(ChannelWithCategory {
                        channel: channel.value().clone(),
                        category,
                    });
                }
            }
        }
        
        // Trier par position
        result.sort_by_key(|ch| ch.channel.config.position);
        result
    }
    
    /// Vérifie une permission spécifique pour un channel
    fn check_channel_permission(
        &self,
        channel_permissions: &ChannelPermissions,
        user_permissions: &UserPermissions,
        required_permission: &ChannelPermission,
    ) -> bool {
        // Permissions utilisateur spécifiques (override)
        if let Some(user_perms) = channel_permissions.user_permissions.get(&user_permissions.user_id) {
            return user_perms.contains(required_permission);
        }
        
        // Permissions de rôle
        for role in &user_permissions.roles {
            let role_str = format!("{:?}", role);
            if let Some(role_perms) = channel_permissions.role_permissions.get(&role_str) {
                if role_perms.contains(required_permission) {
                    return true;
                }
            }
        }
        
        // Permissions par défaut (everyone peut voir les channels publics)
        matches!(required_permission, ChannelPermission::ViewChannel | ChannelPermission::SendMessages)
    }
}

/// Type de cible pour les permissions
#[derive(Debug, Clone)]
pub enum PermissionTargetType {
    Role,
    User,
}

/// Channel avec sa catégorie
#[derive(Debug, Clone, Serialize)]
pub struct ChannelWithCategory {
    pub channel: Channel,
    pub category: Option<ChannelCategory>,
}

impl Default for ChannelType {
    fn default() -> Self {
        Self::Text
    }
}

impl Default for ChannelConfig {
    fn default() -> Self {
        Self {
            name: "general".to_string(),
            topic: None,
            channel_type: ChannelType::Text,
            nsfw: false,
            slowmode_delay: None,
            bitrate: Some(64), // 64 kbps par défaut
            user_limit: None,
            position: 0,
            parent_id: None,
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::permissions::{Role, UserPermissions};

    #[tokio::test]
    async fn test_channel_creation() {
        let manager = ChannelManager::new();
        let mut permissions = UserPermissions::new_user(123);
        permissions.add_role(Role::Admin);
        
        let config = ChannelConfig {
            name: "test-channel".to_string(),
            channel_type: ChannelType::Text,
            ..Default::default()
        };
        
        let channel_id = manager.create_channel("server1", config, 123, &permissions)
            .await.unwrap();
        
        assert!(manager.channels.contains_key(&channel_id));
        assert!(manager.can_view_channel(&channel_id, &permissions));
    }
    
    #[tokio::test]
    async fn test_voice_channel_join() {
        let manager = ChannelManager::new();
        let mut permissions = UserPermissions::new_user(123);
        permissions.add_role(Role::User);
        
        let config = ChannelConfig {
            name: "voice-channel".to_string(),
            channel_type: ChannelType::Voice,
            user_limit: Some(10),
            ..Default::default()
        };
        
        let channel_id = manager.create_channel("server1", config, 123, &permissions)
            .await.unwrap();
        
        manager.join_voice_channel(&channel_id, 123, "testuser".to_string(), &permissions)
            .await.unwrap();
        
        if let Some(channel) = manager.channels.get(&channel_id) {
            assert_eq!(channel.connected_members.len(), 1);
        }
    }
} 