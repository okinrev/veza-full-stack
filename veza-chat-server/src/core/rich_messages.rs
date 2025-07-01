//! Système de Rich Messages Discord-like
//! 
//! Ce module implémente :
//! - Messages avec embeds riches
//! - Système de threads
//! - Réactions avec émojis
//! - Attachements multiples
//! - Mentions et replies
//! - Message pinning et édition

use std::collections::{HashMap, HashSet};
use std::sync::Arc;
use serde::{Deserialize, Serialize};
use chrono::{DateTime, Utc};
use dashmap::DashMap;
use uuid::Uuid;

use crate::error::{ChatError, Result};
use crate::core::message::{StoredMessage, MessageType};

/// Message riche Discord-like avec toutes les fonctionnalités
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RichMessage {
    pub id: String,
    pub channel_id: String,
    pub author_id: i64,
    pub author_username: String,
    pub content: String,
    pub message_type: RichMessageType,
    pub created_at: DateTime<Utc>,
    pub edited_at: Option<DateTime<Utc>>,
    
    /// Embeds riches
    pub embeds: Vec<MessageEmbed>,
    
    /// Attachements (fichiers, images, etc.)
    pub attachments: Vec<MessageAttachment>,
    
    /// Mentions dans le message
    pub mentions: MessageMentions,
    
    /// Réactions au message
    pub reactions: HashMap<String, MessageReaction>,
    
    /// Thread associé (si c'est un message thread)
    pub thread: Option<MessageThread>,
    
    /// Référence à un autre message (reply)
    pub message_reference: Option<MessageReference>,
    
    /// Flags du message
    pub flags: MessageFlags,
    
    /// Activités intégrées (si applicable)
    pub activity: Option<MessageActivity>,
    
    /// Application qui a envoyé le message (pour les bots)
    pub application: Option<MessageApplication>,
}

/// Types de messages riches
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum RichMessageType {
    /// Message normal
    Default,
    /// Message de réponse
    Reply,
    /// Message slash command
    ChatInputCommand,
    /// Message système
    ChannelNameChange,
    ChannelIconChange,
    UserJoin,
    UserPremiumGuildSubscription,
    UserPremiumGuildSubscriptionTier1,
    UserPremiumGuildSubscriptionTier2,
    UserPremiumGuildSubscriptionTier3,
    ChannelFollowAdd,
    /// Message d'appel
    Call,
    /// Message stage
    StageStart,
    StageEnd,
    /// Thread
    ThreadCreated,
    ThreadStarterMessage,
}

/// Embed riche Discord-like
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MessageEmbed {
    /// Titre de l'embed
    pub title: Option<String>,
    /// Description
    pub description: Option<String>,
    /// URL de titre
    pub url: Option<String>,
    /// Timestamp
    pub timestamp: Option<DateTime<Utc>>,
    /// Couleur (format hex)
    pub color: Option<u32>,
    /// Footer
    pub footer: Option<EmbedFooter>,
    /// Image
    pub image: Option<EmbedImage>,
    /// Thumbnail
    pub thumbnail: Option<EmbedThumbnail>,
    /// Video
    pub video: Option<EmbedVideo>,
    /// Provider
    pub provider: Option<EmbedProvider>,
    /// Auteur
    pub author: Option<EmbedAuthor>,
    /// Champs
    pub fields: Vec<EmbedField>,
}

/// Footer d'un embed
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct EmbedFooter {
    pub text: String,
    pub icon_url: Option<String>,
    pub proxy_icon_url: Option<String>,
}

/// Image d'un embed
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct EmbedImage {
    pub url: String,
    pub proxy_url: Option<String>,
    pub height: Option<u32>,
    pub width: Option<u32>,
}

/// Thumbnail d'un embed
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct EmbedThumbnail {
    pub url: String,
    pub proxy_url: Option<String>,
    pub height: Option<u32>,
    pub width: Option<u32>,
}

/// Video d'un embed
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct EmbedVideo {
    pub url: Option<String>,
    pub proxy_url: Option<String>,
    pub height: Option<u32>,
    pub width: Option<u32>,
}

/// Provider d'un embed
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct EmbedProvider {
    pub name: Option<String>,
    pub url: Option<String>,
}

/// Auteur d'un embed
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct EmbedAuthor {
    pub name: String,
    pub url: Option<String>,
    pub icon_url: Option<String>,
    pub proxy_icon_url: Option<String>,
}

/// Champ d'un embed
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct EmbedField {
    pub name: String,
    pub value: String,
    pub inline: bool,
}

/// Attachement de message
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MessageAttachment {
    pub id: String,
    pub filename: String,
    pub description: Option<String>,
    pub content_type: Option<String>,
    pub size: u64,
    pub url: String,
    pub proxy_url: String,
    pub height: Option<u32>,
    pub width: Option<u32>,
    pub ephemeral: bool,
}

/// Mentions dans un message
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MessageMentions {
    /// Utilisateurs mentionnés
    pub users: Vec<i64>,
    /// Rôles mentionnés
    pub roles: Vec<String>,
    /// Channels mentionnés
    pub channels: Vec<String>,
    /// @everyone/@here
    pub everyone: bool,
}

/// Réaction à un message
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MessageReaction {
    /// Nombre de réactions
    pub count: u32,
    /// L'utilisateur actuel a-t-il réagi ?
    pub me: bool,
    /// Emoji utilisé
    pub emoji: ReactionEmoji,
    /// Utilisateurs qui ont réagi
    pub users: HashSet<i64>,
}

/// Emoji de réaction
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ReactionEmoji {
    pub id: Option<String>,
    pub name: String,
    pub animated: bool,
}

/// Thread associé à un message
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MessageThread {
    pub id: String,
    pub name: String,
    pub message_count: u32,
    pub member_count: u32,
    pub last_message_id: Option<String>,
    pub rate_limit_per_user: Option<u32>,
    pub flags: u32,
    pub total_message_sent: u32,
    pub created_at: DateTime<Utc>,
    pub auto_archive_duration: u32,
    pub archive_timestamp: Option<DateTime<Utc>>,
    pub locked: bool,
    pub invitable: bool,
}

/// Référence à un autre message (reply)
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MessageReference {
    pub message_id: String,
    pub channel_id: String,
    pub guild_id: Option<String>,
    pub fail_if_not_exists: bool,
}

/// Flags d'un message
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MessageFlags {
    pub crossposted: bool,
    pub is_crosspost: bool,
    pub suppress_embeds: bool,
    pub source_message_deleted: bool,
    pub urgent: bool,
    pub has_thread: bool,
    pub ephemeral: bool,
    pub loading: bool,
    pub failed_to_mention_some_roles_in_thread: bool,
}

/// Activité de message (jeux, etc.)
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MessageActivity {
    pub activity_type: u8,
    pub party_id: Option<String>,
}

/// Application qui a envoyé le message
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MessageApplication {
    pub id: String,
    pub name: String,
    pub icon: Option<String>,
    pub description: String,
}

/// Gestionnaire de messages riches
#[derive(Debug)]
pub struct RichMessageManager {
    /// Messages par ID
    messages: Arc<DashMap<String, RichMessage>>,
    /// Index des messages par channel
    channel_messages: Arc<DashMap<String, Vec<String>>>,
    /// Index des threads
    threads: Arc<DashMap<String, MessageThread>>,
    /// Index des réactions
    reactions: Arc<DashMap<String, HashMap<String, MessageReaction>>>,
}

impl RichMessageManager {
    pub fn new() -> Self {
        Self {
            messages: Arc::new(DashMap::new()),
            channel_messages: Arc::new(DashMap::new()),
            threads: Arc::new(DashMap::new()),
            reactions: Arc::new(DashMap::new()),
        }
    }
    
    /// Crée un nouveau message riche
    pub async fn create_message(&self, mut message: RichMessage) -> Result<String> {
        let message_id = format!("msg_{}", Uuid::new_v4());
        message.id = message_id.clone();
        message.created_at = Utc::now();
        
        // Valider le contenu
        self.validate_message(&message)?;
        
        // Ajouter le message
        self.messages.insert(message_id.clone(), message.clone());
        
        // Indexer par channel
        self.channel_messages
            .entry(message.channel_id.clone())
            .or_insert_with(Vec::new)
            .push(message_id.clone());
        
        // Créer un thread si nécessaire
        if let Some(thread) = &message.thread {
            self.threads.insert(thread.id.clone(), thread.clone());
        }
        
        Ok(message_id)
    }
    
    /// Édite un message existant
    pub async fn edit_message(
        &self,
        message_id: &str,
        new_content: String,
        new_embeds: Option<Vec<MessageEmbed>>,
        editor_id: i64,
    ) -> Result<()> {
        let mut message = self.messages.get_mut(message_id)
            .ok_or_else(|| ChatError::not_found_simple("message_not_found"))?;
        
        // Vérifier les permissions (l'auteur peut éditer son message)
        if message.author_id != editor_id {
            return Err(ChatError::unauthorized_simple("cannot_edit_message"));
        }
        
        // Mettre à jour le message
        message.content = new_content;
        message.edited_at = Some(Utc::now());
        
        if let Some(embeds) = new_embeds {
            message.embeds = embeds;
        }
        
        Ok(())
    }
    
    /// Ajoute une réaction à un message
    pub async fn add_reaction(
        &self,
        message_id: &str,
        emoji: ReactionEmoji,
        user_id: i64,
    ) -> Result<()> {
        let emoji_clone = emoji.clone();
        let emoji_key = format!("{}:{}", emoji.name, emoji.id.clone().unwrap_or_default());
        
        // Mettre à jour les réactions du message
        if let Some(mut message) = self.messages.get_mut(message_id) {
            let reaction = message.reactions
                .entry(emoji_key.clone())
                .or_insert_with(|| MessageReaction {
                    count: 0,
                    me: false,
                    emoji: emoji_clone.clone(),
                    users: HashSet::new(),
                });
            
            if !reaction.users.contains(&user_id) {
                reaction.users.insert(user_id);
                reaction.count += 1;
            }
        }
        
        Ok(())
    }
    
    /// Retire une réaction d'un message
    pub async fn remove_reaction(
        &self,
        message_id: &str,
        emoji: &ReactionEmoji,
        user_id: i64,
    ) -> Result<()> {
        let emoji_key = format!("{}:{}", emoji.name, emoji.id.as_ref().unwrap_or(&String::new()));
        
        if let Some(mut message) = self.messages.get_mut(message_id) {
            if let Some(reaction) = message.reactions.get_mut(&emoji_key) {
                if reaction.users.remove(&user_id) {
                    reaction.count = reaction.count.saturating_sub(1);
                    
                    // Supprimer la réaction si plus personne n'a réagi
                    if reaction.count == 0 {
                        message.reactions.remove(&emoji_key);
                    }
                }
            }
        }
        
        Ok(())
    }
    
    /// Crée un thread à partir d'un message
    pub async fn create_thread(
        &self,
        message_id: &str,
        thread_name: String,
        auto_archive_duration: u32,
        _creator_id: i64,
    ) -> Result<String> {
        let _message = self.messages.get(message_id)
            .ok_or_else(|| ChatError::not_found_simple("message_not_found"))?;
        
        let thread_id = format!("thread_{}", Uuid::new_v4());
        
        let thread = MessageThread {
            id: thread_id.clone(),
            name: thread_name,
            message_count: 0,
            member_count: 1, // Le créateur
            last_message_id: None,
            rate_limit_per_user: None,
            flags: 0,
            total_message_sent: 0,
            created_at: Utc::now(),
            auto_archive_duration,
            archive_timestamp: None,
            locked: false,
            invitable: true,
        };
        
        // Ajouter le thread
        self.threads.insert(thread_id.clone(), thread.clone());
        
        // Mettre à jour le message pour indiquer qu'il a un thread
        if let Some(mut msg) = self.messages.get_mut(message_id) {
            msg.thread = Some(thread);
            msg.flags.has_thread = true;
        }
        
        Ok(thread_id)
    }
    
    /// Pin/Unpin un message dans un channel
    pub async fn toggle_pin(
        &self,
        _message_id: &str,
        _pinner_id: i64,
    ) -> Result<bool> {
        // Dans une vraie implémentation, on vérifierait les permissions ici
        // Pour l'instant, on simule juste le changement d'état
        
        // Retourner le nouvel état (pinned ou non)
        Ok(true) // Simulé
    }
    
    /// Obtient les messages d'un channel avec pagination
    pub fn get_channel_messages(
        &self,
        channel_id: &str,
        limit: usize,
        before: Option<&str>,
        after: Option<&str>,
    ) -> Vec<RichMessage> {
        if let Some(message_ids) = self.channel_messages.get(channel_id) {
            let mut messages = Vec::new();
            
            for msg_id in message_ids.iter() {
                if let Some(message) = self.messages.get(msg_id) {
                    messages.push(message.value().clone());
                }
            }
            
            // Trier par date (plus récent en premier)
            messages.sort_by(|a, b| b.created_at.cmp(&a.created_at));
            
            // Appliquer la pagination
            if let Some(before_id) = before {
                if let Some(pos) = messages.iter().position(|m| m.id == before_id) {
                    messages = messages.into_iter().skip(pos + 1).collect();
                }
            }
            
            if let Some(after_id) = after {
                if let Some(pos) = messages.iter().position(|m| m.id == after_id) {
                    messages = messages.into_iter().take(pos).collect();
                }
            }
            
            messages.into_iter().take(limit).collect()
        } else {
            Vec::new()
        }
    }
    
    /// Valide un message
    fn validate_message(&self, message: &RichMessage) -> Result<()> {
        // Vérifier la longueur du contenu
        if message.content.len() > 2000 {
            return Err(ChatError::validation_error("message_too_long"));
        }
        
        // Vérifier le nombre d'embeds
        if message.embeds.len() > 10 {
            return Err(ChatError::validation_error("too_many_embeds"));
        }
        
        // Vérifier les attachements
        if message.attachments.len() > 10 {
            return Err(ChatError::validation_error("too_many_attachments"));
        }
        
        // Vérifier la taille totale des attachements
        let total_size: u64 = message.attachments.iter().map(|a| a.size).sum();
        if total_size > 100 * 1024 * 1024 { // 100 MB
            return Err(ChatError::validation_error("attachments_too_large"));
        }
        
        Ok(())
    }
}

impl Default for MessageFlags {
    fn default() -> Self {
        Self {
            crossposted: false,
            is_crosspost: false,
            suppress_embeds: false,
            source_message_deleted: false,
            urgent: false,
            has_thread: false,
            ephemeral: false,
            loading: false,
            failed_to_mention_some_roles_in_thread: false,
        }
    }
}

impl Default for MessageMentions {
    fn default() -> Self {
        Self {
            users: Vec::new(),
            roles: Vec::new(),
            channels: Vec::new(),
            everyone: false,
        }
    }
}

impl Default for RichMessageManager {
    fn default() -> Self {
        Self {
            messages: Arc::new(DashMap::new()),
            channel_messages: Arc::new(DashMap::new()),
            threads: Arc::new(DashMap::new()),
            reactions: Arc::new(DashMap::new()),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_rich_message_creation() {
        let manager = RichMessageManager::new();
        
        let message = RichMessage {
            id: String::new(), // Sera généré
            channel_id: "channel123".to_string(),
            author_id: 456,
            author_username: "testuser".to_string(),
            content: "Hello **world**!".to_string(),
            message_type: RichMessageType::Default,
            created_at: Utc::now(),
            edited_at: None,
            embeds: vec![],
            attachments: vec![],
            mentions: MessageMentions::default(),
            reactions: HashMap::new(),
            thread: None,
            message_reference: None,
            flags: MessageFlags::default(),
            activity: None,
            application: None,
        };
        
        let message_id = manager.create_message(message).await.unwrap();
        assert!(manager.messages.contains_key(&message_id));
    }
    
    #[tokio::test]
    async fn test_message_reactions() {
        let manager = RichMessageManager::new();
        
        let message = RichMessage {
            id: "msg123".to_string(),
            channel_id: "channel123".to_string(),
            author_id: 456,
            author_username: "testuser".to_string(),
            content: "React to this!".to_string(),
            message_type: RichMessageType::Default,
            created_at: Utc::now(),
            edited_at: None,
            embeds: vec![],
            attachments: vec![],
            mentions: MessageMentions::default(),
            reactions: HashMap::new(),
            thread: None,
            message_reference: None,
            flags: MessageFlags::default(),
            activity: None,
            application: None,
        };
        
        manager.messages.insert("msg123".to_string(), message);
        
        let emoji = ReactionEmoji {
            id: None,
            name: "👍".to_string(),
            animated: false,
        };
        
        manager.add_reaction("msg123", emoji.clone(), 789).await.unwrap();
        
        let message = manager.messages.get("msg123").unwrap();
        let reaction_key = format!("{}:{}", emoji.name, "");
        assert!(message.reactions.contains_key(&reaction_key));
        assert_eq!(message.reactions[&reaction_key].count, 1);
    }
} 