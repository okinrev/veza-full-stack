//! Message Types et Protocol
//! 
//! Types de messages optimisés pour Discord-like features
//! avec support threads, réactions, mentions, etc.

use uuid::Uuid;
use serde::{Serialize, Deserialize};
use chrono::{DateTime, Utc};

/// Message stocké avec métadonnées complètes
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct StoredMessage {
    pub id: Uuid,
    pub content: String,
    pub author_id: i64,
    pub timestamp: DateTime<Utc>,
    pub message_type: MessageType,
    pub room_id: String,
    
    // Features Discord-like
    pub thread_id: Option<Uuid>,
    pub reply_to: Option<Uuid>,
    pub mentions: Vec<i64>,
    pub reactions: Vec<MessageReaction>,
    pub attachments: Vec<MessageAttachment>,
    pub embeds: Vec<MessageEmbed>,
    
    // Modération
    pub edited_at: Option<DateTime<Utc>>,
    pub deleted_at: Option<DateTime<Utc>>,
    pub moderation_flags: ModerationFlags,
}

/// Type de message
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum MessageType {
    Text,
    File,
    Image,
    Voice,
    Video,
    System,
    ThreadStart,
    ThreadReply,
}

/// Réaction à un message
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MessageReaction {
    pub emoji: String,
    pub users: Vec<i64>,
    pub count: u32,
}

/// Pièce jointe
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MessageAttachment {
    pub id: Uuid,
    pub filename: String,
    pub content_type: String,
    pub size: u64,
    pub url: String,
    pub proxy_url: Option<String>,
    pub width: Option<u32>,
    pub height: Option<u32>,
}

/// Embed riche (Discord-like)
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MessageEmbed {
    pub title: Option<String>,
    pub description: Option<String>,
    pub url: Option<String>,
    pub color: Option<u32>,
    pub timestamp: Option<DateTime<Utc>>,
    pub footer: Option<EmbedFooter>,
    pub image: Option<EmbedImage>,
    pub thumbnail: Option<EmbedThumbnail>,
    pub author: Option<EmbedAuthor>,
    pub fields: Vec<EmbedField>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct EmbedFooter {
    pub text: String,
    pub icon_url: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct EmbedImage {
    pub url: String,
    pub width: Option<u32>,
    pub height: Option<u32>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct EmbedThumbnail {
    pub url: String,
    pub width: Option<u32>,
    pub height: Option<u32>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct EmbedAuthor {
    pub name: String,
    pub url: Option<String>,
    pub icon_url: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct EmbedField {
    pub name: String,
    pub value: String,
    pub inline: bool,
}

/// Flags de modération
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct ModerationFlags {
    pub is_flagged: bool,
    pub is_spam: bool,
    pub toxicity_score: Option<f32>,
    pub auto_moderated: bool,
    pub manual_review: bool,
}

/// Message entrant du WebSocket
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(tag = "type", content = "data")]
pub enum IncomingMessage {
    // Messages basiques
    SendMessage {
        room_id: String,
        content: String,
        reply_to: Option<Uuid>,
        thread_id: Option<Uuid>,
    },
    
    EditMessage {
        message_id: Uuid,
        content: String,
    },
    
    DeleteMessage {
        message_id: Uuid,
    },
    
    // Réactions
    AddReaction {
        message_id: Uuid,
        emoji: String,
    },
    
    RemoveReaction {
        message_id: Uuid,
        emoji: String,
    },
    
    // Salles
    JoinRoom {
        room_id: String,
    },
    
    LeaveRoom {
        room_id: String,
    },
    
    // Présence
    UpdatePresence {
        status: super::user::PresenceStatus,
        activity: Option<String>,
    },
    
    StartTyping {
        room_id: String,
    },
    
    StopTyping {
        room_id: String,
    },
    
    // Threads
    CreateThread {
        message_id: Uuid,
        name: String,
    },
    
    // Modération
    ReportMessage {
        message_id: Uuid,
        reason: String,
    },
}

/// Message sortant vers le WebSocket
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(tag = "type", content = "data")]
pub enum OutgoingMessage {
    // Messages
    MessageReceived {
        message: StoredMessage,
    },
    
    MessageEdited {
        message_id: Uuid,
        content: String,
        edited_at: DateTime<Utc>,
    },
    
    MessageDeleted {
        message_id: Uuid,
        deleted_at: DateTime<Utc>,
    },
    
    // Réactions
    ReactionAdded {
        message_id: Uuid,
        emoji: String,
        user_id: i64,
    },
    
    ReactionRemoved {
        message_id: Uuid,
        emoji: String,
        user_id: i64,
    },
    
    // Présence
    UserPresenceUpdate {
        user_id: i64,
        status: super::user::PresenceStatus,
        activity: Option<String>,
    },
    
    TypingStart {
        room_id: String,
        user_id: i64,
    },
    
    TypingStop {
        room_id: String,
        user_id: i64,
    },
    
    // Salles
    RoomJoined {
        room_id: String,
        user_id: i64,
    },
    
    RoomLeft {
        room_id: String,
        user_id: i64,
    },
    
    // Système
    Error {
        message: String,
        code: Option<String>,
    },
    
    ActionConfirmed {
        action: String,
        success: bool,
    },
    
    // Threads
    ThreadCreated {
        thread_id: Uuid,
        parent_message_id: Uuid,
        name: String,
        creator_id: i64,
    },
}

impl StoredMessage {
    pub fn new_text_message(
        author_id: i64,
        room_id: String,
        content: String,
    ) -> Self {
        Self {
            id: Uuid::new_v4(),
            content,
            author_id,
            timestamp: Utc::now(),
            message_type: MessageType::Text,
            room_id,
            thread_id: None,
            reply_to: None,
            mentions: Vec::new(),
            reactions: Vec::new(),
            attachments: Vec::new(),
            embeds: Vec::new(),
            edited_at: None,
            deleted_at: None,
            moderation_flags: ModerationFlags::default(),
        }
    }

    pub fn add_reaction(&mut self, emoji: String, user_id: i64) {
        if let Some(reaction) = self.reactions.iter_mut()
            .find(|r| r.emoji == emoji) {
            if !reaction.users.contains(&user_id) {
                reaction.users.push(user_id);
                reaction.count += 1;
            }
        } else {
            self.reactions.push(MessageReaction {
                emoji,
                users: vec![user_id],
                count: 1,
            });
        }
    }

    pub fn remove_reaction(&mut self, emoji: &str, user_id: i64) {
        if let Some(reaction) = self.reactions.iter_mut()
            .find(|r| r.emoji == emoji) {
            if let Some(pos) = reaction.users.iter().position(|&id| id == user_id) {
                reaction.users.remove(pos);
                reaction.count -= 1;
                
                // Supprimer la réaction si plus personne
                if reaction.count == 0 {
                    self.reactions.retain(|r| r.emoji != emoji);
                }
            }
        }
    }

    pub fn is_deleted(&self) -> bool {
        self.deleted_at.is_some()
    }

    pub fn is_edited(&self) -> bool {
        self.edited_at.is_some()
    }
}
