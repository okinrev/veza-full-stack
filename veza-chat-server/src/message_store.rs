use sqlx::PgPool;
use chrono::{DateTime, Utc};
use serde::{Serialize, Deserialize};
use std::collections::HashMap;

use crate::error::{ChatError, Result};

/// Types de messages différenciés
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum MessageType {
    RoomMessage,
    DirectMessage,
    SystemMessage,
}

impl Default for MessageType {
    fn default() -> Self {
        MessageType::RoomMessage
    }
}

impl ToString for MessageType {
    fn to_string(&self) -> String {
        match self {
            MessageType::RoomMessage => "RoomMessage".to_string(),
            MessageType::DirectMessage => "DirectMessage".to_string(),
            MessageType::SystemMessage => "SystemMessage".to_string(),
        }
    }
}

/// Statut des messages
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum MessageStatus {
    Sent,
    Delivered,
    Read,
    Edited,
    Deleted,
}

impl Default for MessageStatus {
    fn default() -> Self {
        MessageStatus::Sent
    }
}

impl ToString for MessageStatus {
    fn to_string(&self) -> String {
        match self {
            MessageStatus::Sent => "Sent".to_string(),
            MessageStatus::Delivered => "Delivered".to_string(),
            MessageStatus::Read => "Read".to_string(),
            MessageStatus::Edited => "Edited".to_string(),
            MessageStatus::Deleted => "Deleted".to_string(),
        }
    }
}

/// Message unifié avec séparation logique
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Message {
    pub id: i64,
    pub message_type: MessageType,
    pub content: String,
    pub author_id: i64,
    pub author_username: String,
    
    // Pour les messages de salon
    pub room_id: Option<String>,
    
    // Pour les messages directs
    pub recipient_id: Option<i64>,
    pub recipient_username: Option<String>,
    
    // Métadonnées communes
    pub created_at: DateTime<Utc>,
    pub updated_at: Option<DateTime<Utc>>,
    pub status: MessageStatus,
    pub is_pinned: bool,
    pub is_edited: bool,
    pub original_content: Option<String>, // Contenu original avant édition
    
    // Thread/réponse
    pub parent_message_id: Option<i64>,
    pub thread_count: i32,
    
    // Réactions
    pub reactions: HashMap<String, Vec<i64>>, // emoji -> liste d'user_ids
    
    // Attachments
    pub attachments: Vec<String>,
    
    // Mentions
    pub mentions: Vec<i64>, // user_ids mentionnés
    
    // Métadonnées de modération
    pub is_flagged: bool,
    pub moderation_notes: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MessageAttachment {
    pub id: i64,
    pub filename: String,
    pub original_filename: String,
    pub mime_type: String,
    pub size_bytes: i64,
    pub url: String,
    pub thumbnail_url: Option<String>,
    pub uploaded_at: DateTime<Utc>,
}

/// Réaction à un message (avec types i64 pour compatibilité DB)
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MessageReaction {
    pub id: i64,
    pub message_id: i64,
    pub user_id: i64,
    pub emoji: String,
    pub created_at: DateTime<Utc>,
}

/// Gestionnaire de stockage de messages séparé
pub struct MessageStore {
    db: PgPool,
}

impl MessageStore {
    pub fn new(db: PgPool) -> Self {
        Self { db }
    }

    /// Sauvegarde un message dans la base de données
    pub async fn save_message(&self, message: &Message) -> Result<()> {
            sqlx::query!(
            r#"INSERT INTO messages 
               (id, message_type, content, author_id, author_username, room_id, recipient_id, recipient_username, 
                parent_message_id, thread_count, status, is_pinned, is_edited, original_content, created_at, updated_at) 
               VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)"#,
            message.id,
            message.message_type.to_string(),
            message.content,
            message.author_id,
            message.author_username,
            message.room_id,
            message.recipient_id,
            message.recipient_username,
            message.parent_message_id,
            message.thread_count,
            message.status.to_string(),
            message.is_pinned,
            message.is_edited,
            message.original_content,
            message.created_at,
            message.updated_at
            )
            .execute(&self.db)
            .await
        .map_err(|e| ChatError::database_error("save_message", e))?;

        // Sauvegarder les mentions si elles existent
        if !message.mentions.is_empty() {
            for mention in &message.mentions {
            sqlx::query!(
                    "INSERT INTO message_mentions (message_id, user_id, mentioned_user_id) VALUES ($1, $2, $3)",
                    message.id,
                    message.author_id,
                    mention
            )
            .execute(&self.db)
            .await
                .map_err(|e| ChatError::database_error("save_message_mentions", e))?;
            }
        }

        Ok(())
    }

    /// Récupère un message par son ID
    pub async fn get_message(&self, message_id: i64) -> Result<Option<Message>> {
        let result = sqlx::query!(
            r#"SELECT 
                id, message_type, content, author_id, author_username,
                room_id, recipient_id, recipient_username, created_at, updated_at,
                status, is_pinned, is_edited, original_content,
                parent_message_id, thread_count
             FROM messages WHERE id = $1"#,
            message_id
        )
        .fetch_optional(&self.db)
        .await?;

        if let Some(row) = result {
            let message = Message {
                id: row.id,
                message_type: match row.message_type.as_ref().map(|s| s.as_str()) {
                    Some("RoomMessage") => MessageType::RoomMessage,
                    Some("DirectMessage") => MessageType::DirectMessage,
                    Some("SystemMessage") => MessageType::SystemMessage,
                    _ => MessageType::RoomMessage,
                },
                content: row.content,
                author_id: row.author_id.unwrap_or(0),
                author_username: row.author_username.unwrap_or_default(),
                room_id: row.room_id,
                recipient_id: row.recipient_id,
                recipient_username: row.recipient_username,
                created_at: row.created_at,
                updated_at: Some(row.updated_at),
                status: match row.status.as_str() {
                    "Sent" => MessageStatus::Sent,
                    "Delivered" => MessageStatus::Delivered,
                    "Read" => MessageStatus::Read,
                    "Edited" => MessageStatus::Edited,
                    "Deleted" => MessageStatus::Deleted,
                    _ => MessageStatus::Sent,
                },
                is_pinned: row.is_pinned.unwrap_or(false),
                is_edited: row.is_edited.unwrap_or(false),
                original_content: row.original_content,
                parent_message_id: row.parent_message_id,
                thread_count: row.thread_count.unwrap_or(0),
                reactions: HashMap::<String, Vec<i64>>::new(),
                attachments: Vec::<String>::new(),
                mentions: Vec::<i64>::new(),
                is_flagged: false,
                moderation_notes: None,
            };
            Ok(Some(message))
        } else {
            Ok(None)
        }
    }

    /// Récupère les messages d'un salon avec pagination
    pub async fn get_room_messages(&self, room_id: Option<String>, limit: i32, before_id: Option<i64>) -> Result<Vec<Message>> {
        let mut result = Vec::new();
        
        let query_result = sqlx::query!(
            r#"SELECT * FROM messages 
             WHERE room_id = $1 
             AND ($2::BIGINT IS NULL OR id < $2)
             ORDER BY created_at DESC 
             LIMIT $3"#,
            room_id,
            before_id,
            limit as i64
        )
        .fetch_all(&self.db)
        .await?;

        for row in query_result {
            let message = Message {
                id: row.id,
                message_type: match row.message_type.as_ref().map(|s| s.as_str()) {
                    Some("RoomMessage") => MessageType::RoomMessage,
                    Some("DirectMessage") => MessageType::DirectMessage,
                    Some("SystemMessage") => MessageType::SystemMessage,
                    _ => MessageType::RoomMessage,
                },
                content: row.content,
                author_id: row.author_id.unwrap_or(0),
                author_username: row.author_username.unwrap_or_default(),
                room_id: row.room_id,
                recipient_id: row.recipient_id,
                recipient_username: row.recipient_username,
                created_at: row.created_at,
                updated_at: Some(row.updated_at),
                status: match row.status.as_str() {
                    "Sent" => MessageStatus::Sent,
                    "Delivered" => MessageStatus::Delivered,
                    "Read" => MessageStatus::Read,
                    "Edited" => MessageStatus::Edited,
                    "Deleted" => MessageStatus::Deleted,
                    _ => MessageStatus::Sent,
                },
                is_pinned: row.is_pinned.unwrap_or(false),
                is_edited: row.is_edited.unwrap_or(false),
                original_content: row.original_content,
                parent_message_id: row.parent_message_id,
                thread_count: row.thread_count.unwrap_or(0),
                reactions: HashMap::<String, Vec<i64>>::new(),
                attachments: Vec::<String>::new(),
                mentions: Vec::<i64>::new(),
                is_flagged: false,
                moderation_notes: None,
            };
            result.push(message);
        }

        // Vérifier si l'utilisateur peut voir les messages dans ce salon
        if let Some(conv_id) = &room_id {
            // Note: user_id should be passed as parameter in real implementation
            let user_id = 1i64; // Placeholder - should come from context
            let is_participant = self.is_conversation_participant(user_id as i32, conv_id).await?;
            if !is_participant {
                return Ok(Vec::new());
            }
        }

        Ok(result)
    }

    /// Récupère les messages directs entre deux utilisateurs
    pub async fn get_dm_messages(&self, user1_id: i64, user2_id: i64, limit: i32, before_id: Option<i64>) -> Result<Vec<Message>> {
        let mut result = Vec::new();
        
        let query_result = sqlx::query!(
            r#"SELECT * FROM messages 
            WHERE ((author_id = $1 AND recipient_id = $2) OR (author_id = $2 AND recipient_id = $1))
            AND ($3::BIGINT IS NULL OR id < $3)
            ORDER BY created_at DESC 
            LIMIT $4"#,
            user1_id,
            user2_id,
            before_id,
            limit as i64
        )
        .fetch_all(&self.db)
        .await?;

        for row in query_result {
            let message = Message {
                id: row.id,
                message_type: match row.message_type.as_ref().map(|s| s.as_str()) {
                    Some("RoomMessage") => MessageType::RoomMessage,
                    Some("DirectMessage") => MessageType::DirectMessage,
                    Some("SystemMessage") => MessageType::SystemMessage,
                    _ => MessageType::DirectMessage,
                },
                content: row.content,
                author_id: row.author_id.unwrap_or(0),
                author_username: row.author_username.unwrap_or_default(),
                room_id: row.room_id,
                recipient_id: row.recipient_id,
                recipient_username: row.recipient_username,
                created_at: row.created_at,
                updated_at: Some(row.updated_at),
                status: match row.status.as_str() {
                    "Sent" => MessageStatus::Sent,
                    "Delivered" => MessageStatus::Delivered,
                    "Read" => MessageStatus::Read,
                    "Edited" => MessageStatus::Edited,
                    "Deleted" => MessageStatus::Deleted,
                    _ => MessageStatus::Sent,
                },
                is_pinned: row.is_pinned.unwrap_or(false),
                is_edited: row.is_edited.unwrap_or(false),
                original_content: row.original_content,
                parent_message_id: row.parent_message_id,
                thread_count: row.thread_count.unwrap_or(0),
                reactions: HashMap::<String, Vec<i64>>::new(),
                attachments: Vec::<String>::new(),
                mentions: Vec::<i64>::new(),
                is_flagged: false,
                moderation_notes: None,
            };
            result.push(message);
        }

        Ok(result)
    }

    /// Épingle un message
    pub async fn pin_message(&self, message_id: i64, room_id: Option<String>) -> Result<()> {
        // Vérifier que le message existe dans le salon
        let message = sqlx::query!(
            "SELECT id FROM messages WHERE id = $1 AND room_id = $2",
            message_id,
            room_id
        )
        .fetch_optional(&self.db)
        .await
        .map_err(|e| ChatError::database_error("pin_message_check", e))?;

        if message.is_none() {
            return Err(ChatError::configuration_error("Message non trouvé dans ce salon"));
        }

        // Vérifier la limite de messages épinglés (max 10 par salon)
        let pinned_count = sqlx::query!(
            "SELECT COUNT(*) as count FROM messages WHERE room_id = $1 AND is_pinned = true",
                room_id
            )
            .fetch_one(&self.db)
            .await
        .map_err(|e| ChatError::database_error("pin_message_count", e))?;

        if pinned_count.count.unwrap_or(0) >= 10 {
            return Err(ChatError::configuration_error("Limite de messages épinglés atteinte (10 par salon)"));
        }

        // Épingler le message
        sqlx::query!(
            "UPDATE messages SET is_pinned = true, updated_at = NOW() WHERE id = $1",
            message_id
        )
        .execute(&self.db)
        .await
        .map_err(|e| ChatError::database_error("pin_message_update", e))?;

        Ok(())
    }

    /// Désépingle un message
    pub async fn unpin_message(&self, message_id: i64) -> Result<()> {
        sqlx::query!(
            "UPDATE messages SET is_pinned = false, updated_at = NOW() WHERE id = $1",
            message_id
        )
        .execute(&self.db)
        .await
        .map_err(|e| ChatError::database_error("unpin_message", e))?;

        Ok(())
    }

    /// Récupère les messages épinglés d'un salon
    pub async fn get_pinned_messages(&self, room_id: Option<String>) -> Result<Vec<Message>> {
        let messages = sqlx::query!(
            r#"SELECT m.*, ARRAY_AGG(mm.user_id) as mention_ids
            FROM messages m
            LEFT JOIN message_mentions mm ON m.id = mm.message_id
               WHERE m.room_id = $1 AND m.is_pinned = true
            GROUP BY m.id
               ORDER BY m.created_at DESC"#,
            room_id
        )
        .fetch_all(&self.db)
        .await
        .map_err(|e| ChatError::database_error("get_pinned_messages", e))?;

        let mut result = Vec::new();
        for row in messages {
            let message = Message {
                id: row.id,
                message_type: match row.message_type.as_ref().map(|s| s.as_str()) {
                    Some("RoomMessage") => MessageType::RoomMessage,
                    Some("DirectMessage") => MessageType::DirectMessage,
                    Some("SystemMessage") => MessageType::SystemMessage,
                    _ => MessageType::RoomMessage,
                },
                content: row.content,
                author_id: row.author_id.unwrap_or(0),
                author_username: row.author_username.unwrap_or_default(),
                room_id: row.room_id,
                recipient_id: row.recipient_id,
                recipient_username: row.recipient_username,
                created_at: row.created_at,
                updated_at: Some(row.updated_at),
                status: match row.status.as_str() {
                    "Sent" => MessageStatus::Sent,
                    "Delivered" => MessageStatus::Delivered,
                    "Read" => MessageStatus::Read,
                    "Edited" => MessageStatus::Edited,
                    "Deleted" => MessageStatus::Deleted,
                    _ => MessageStatus::Sent,
                },
                is_pinned: row.is_pinned.unwrap_or(false),
                is_edited: row.is_edited.unwrap_or(false),
                original_content: row.original_content,
                parent_message_id: row.parent_message_id,
                thread_count: row.thread_count.unwrap_or(0),
                reactions: HashMap::<String, Vec<i64>>::new(),
                attachments: Vec::<String>::new(),
                mentions: Vec::<i64>::new(),
                is_flagged: false,
                moderation_notes: None,
            };
            result.push(message);
        }

        Ok(result)
    }

    /// Marque les messages comme lus pour un utilisateur
    pub async fn mark_messages_as_read(&self, user_id: i64, conversation_id: Option<String>) -> Result<()> {
        // Vérifier que l'utilisateur peut marquer ces messages comme lus
        if let Some(conv_id) = &conversation_id {
            // Vérifier que l'utilisateur fait partie de cette conversation
            let is_participant = self.is_conversation_participant(user_id as i32, conv_id).await?;
            if !is_participant {
                return Err(ChatError::unauthorized("Non autorisé à marquer ce message comme lu"));
            }
        }

        // Marquer comme lu
        sqlx::query!(
            r#"INSERT INTO message_read_status (user_id, message_id, read_at)
               SELECT $1, m.id, NOW()
            FROM messages m
               WHERE m.room_id IS NOT NULL OR m.recipient_id = $1
               ON CONFLICT (user_id, message_id) DO UPDATE SET read_at = NOW()"#,
            user_id
        )
        .execute(&self.db)
        .await
        .map_err(|e| ChatError::database_error("mark_messages_as_read", e))?;

        Ok(())
    }

    /// Compte les messages non lus pour un utilisateur
    pub async fn count_unread_messages(&self, user_id: i64) -> Result<HashMap<String, i64>> {
        let rows = sqlx::query!(
            r#"
            SELECT room_id, COUNT(*) as unread_count
            FROM messages m
            WHERE m.author_id != $1 
            AND m.room_id IS NOT NULL
            AND m.created_at > NOW() - INTERVAL '7 days'
            GROUP BY room_id
            "#,
            user_id
        )
        .fetch_all(&self.db)
        .await?;

        let mut unread_counts = HashMap::new();
        for row in rows {
            if let Some(room_id) = row.room_id {
                unread_counts.insert(room_id, row.unread_count.unwrap_or(0));
            }
        }

        Ok(unread_counts)
    }

    /// Vérifie si un utilisateur participe à une conversation
    async fn is_conversation_participant(&self, _user_id: i32, _conversation_id: &str) -> Result<bool> {
        // Placeholder implementation - should check actual conversation participants
        Ok(true)
    }

    /// Ajoute une réaction à un message
    pub async fn add_reaction(&self, message_id: i64, user_id: i64, emoji: &str) -> Result<()> {
        // Vérifier si la réaction existe déjà
        let existing = sqlx::query!(
            "SELECT id FROM message_reactions WHERE message_id = $1 AND user_id = $2 AND emoji = $3",
            message_id,
            user_id,
            emoji
        )
        .fetch_optional(&self.db)
        .await
        .map_err(|e| ChatError::database_error("add_reaction_check", e))?;

        if existing.is_some() {
            return Err(ChatError::configuration_error("Réaction déjà existante"));
        }

        // Ajouter la réaction
        sqlx::query!(
            "INSERT INTO message_reactions (message_id, user_id, emoji, created_at) VALUES ($1, $2, $3, NOW())",
            message_id,
            user_id,
            emoji
        )
        .execute(&self.db)
        .await
        .map_err(|e| ChatError::database_error("add_reaction", e))?;

        Ok(())
    }

    /// Supprime une réaction d'un message
    pub async fn remove_reaction(&self, message_id: i64, user_id: i64, emoji: &str) -> Result<()> {
        let result = sqlx::query!(
            "DELETE FROM message_reactions WHERE message_id = $1 AND user_id = $2 AND emoji = $3",
            message_id,
            user_id,
            emoji
        )
        .execute(&self.db)
        .await
        .map_err(|e| ChatError::database_error("remove_reaction", e))?;

        if result.rows_affected() == 0 {
            return Err(ChatError::not_found("Reaction", "unknown"));
        }

        Ok(())
    }

    /// Récupère les réactions d'un message
    pub async fn get_message_reactions(&self, message_id: i64) -> Result<Vec<MessageReaction>> {
        let reactions = sqlx::query!(
            "SELECT id, message_id, user_id, emoji, created_at FROM message_reactions WHERE message_id = $1 ORDER BY created_at",
            message_id
        )
        .fetch_all(&self.db)
        .await?;

        let mut result = Vec::new();
        for row in reactions {
            result.push(MessageReaction {
                id: row.id,
                message_id: row.message_id,
                user_id: row.user_id,
                emoji: row.emoji,
                created_at: row.created_at,
            });
        }

        Ok(result)
    }

    /// Édite un message
    pub async fn edit_message(&self, message_id: i64, new_content: &str, editor_user_id: i64) -> Result<()> {
        // Vérifier que l'utilisateur peut éditer ce message
        let message = sqlx::query!(
            "SELECT author_id, content FROM messages WHERE id = $1",
            message_id
        )
        .fetch_one(&self.db)
        .await?;

        if message.author_id != Some(editor_user_id) {
            return Err(crate::error::ChatError::PermissionDenied { 
                message: "Cannot edit another user's message".to_string() 
            });
        }

        // Mettre à jour le message
        sqlx::query!(
            r#"
            UPDATE messages 
            SET content = $1, 
                is_edited = true, 
                original_content = COALESCE(original_content, $2),
                updated_at = NOW()
            WHERE id = $3
            "#,
            new_content,
            message.content,
            message_id
        )
        .execute(&self.db)
        .await?;

        Ok(())
    }

    /// Supprime un message (soft delete)
    pub async fn delete_message(&self, message_id: i64, deleter_user_id: i64, is_moderator: bool) -> Result<()> {
        // Vérifier que l'utilisateur peut supprimer ce message
        let message = sqlx::query!(
            "SELECT author_id FROM messages WHERE id = $1",
            message_id
        )
        .fetch_one(&self.db)
        .await?;

        if message.author_id != Some(deleter_user_id) && !is_moderator {
            return Err(crate::error::ChatError::PermissionDenied { 
                message: "Cannot delete another user's message".to_string() 
            });
        }

        // Soft delete du message
        sqlx::query!(
            "UPDATE messages SET status = 'Deleted', updated_at = NOW() WHERE id = $1",
            message_id
        )
        .execute(&self.db)
        .await?;

        Ok(())
    }

    /// Recherche des messages avec un terme
    pub async fn search_messages(&self, query: &str, user_id: i64, limit: i32) -> Result<Vec<Message>> {
        let mut result = Vec::new();
        
        let search_query = format!("%{}%", query);
        let rows = sqlx::query!(
            r#"
            SELECT * FROM messages
            WHERE content ILIKE $1
            AND (author_id = $2 OR room_id IS NOT NULL)
            ORDER BY created_at DESC
            LIMIT $3
            "#,
            search_query,
            user_id,
            limit as i64
        )
        .fetch_all(&self.db)
        .await?;

        for row in rows {
            let message = Message {
                id: row.id,
                message_type: match row.message_type.as_ref().map(|s| s.as_str()) {
                    Some("RoomMessage") => MessageType::RoomMessage,
                    Some("DirectMessage") => MessageType::DirectMessage,
                    Some("SystemMessage") => MessageType::SystemMessage,
                    _ => MessageType::RoomMessage,
                },
                content: row.content,
                author_id: row.author_id.unwrap_or(0),
                author_username: row.author_username.unwrap_or_default(),
                room_id: row.room_id,
                recipient_id: row.recipient_id,
                recipient_username: row.recipient_username,
                created_at: row.created_at,
                updated_at: Some(row.updated_at),
                status: match row.status.as_str() {
                    "Sent" => MessageStatus::Sent,
                    "Delivered" => MessageStatus::Delivered,
                    "Read" => MessageStatus::Read,
                    "Edited" => MessageStatus::Edited,
                    "Deleted" => MessageStatus::Deleted,
                    _ => MessageStatus::Sent,
                },
                is_pinned: row.is_pinned.unwrap_or(false),
                is_edited: row.is_edited.unwrap_or(false),
                original_content: row.original_content,
                parent_message_id: row.parent_message_id,
                thread_count: row.thread_count.unwrap_or(0),
                reactions: HashMap::<String, Vec<i64>>::new(),
                attachments: Vec::<String>::new(),
                mentions: Vec::<i64>::new(),
                is_flagged: false,
                moderation_notes: None,
            };
            result.push(message);
        }

        Ok(result)
    }

    /// Statistiques des messages pour un utilisateur
    pub async fn get_user_message_stats(&self, user_id: i64) -> Result<MessageStats> {
        let stats = sqlx::query!(
            r#"
            SELECT 
                COUNT(*) as total_sent,
                COUNT(*) FILTER (WHERE room_id IS NOT NULL) as room_messages,
                COUNT(*) FILTER (WHERE recipient_id IS NOT NULL) as dm_messages,
                COUNT(*) FILTER (WHERE is_edited = true) as edited_messages,
                COUNT(*) FILTER (WHERE is_pinned = true) as pinned_messages
            FROM messages 
            WHERE author_id = $1
            "#,
            user_id
        )
        .fetch_one(&self.db)
        .await?;

        Ok(MessageStats {
            total_sent: stats.total_sent.unwrap_or(0) as u64,
            room_messages: stats.room_messages.unwrap_or(0) as u64,
            dm_messages: stats.dm_messages.unwrap_or(0) as u64,
            edited_messages: stats.edited_messages.unwrap_or(0) as u64,
            pinned_messages: stats.pinned_messages.unwrap_or(0) as u64,
        })
    }

    /// Statistiques globales des messages
    pub async fn get_global_message_stats(&self) -> Result<GlobalMessageStats> {
        // Requête pour les statistiques de base
        let stats = sqlx::query!(
            r#"SELECT 
                   COUNT(*) as total_messages,
                   COUNT(DISTINCT author_id) as total_users,
                   COUNT(CASE WHEN room_id IS NOT NULL THEN 1 END) as total_room_messages,
                   COUNT(CASE WHEN room_id IS NULL THEN 1 END) as total_dm_messages
               FROM messages WHERE status != 'deleted'"#
        )
        .fetch_one(&self.db)
        .await
        .map_err(|e| ChatError::database_error("get_global_message_stats", e))?;

        // Requête pour compter les salons uniques
        let total_rooms = sqlx::query_scalar!(
            "SELECT COUNT(DISTINCT room_id) FROM messages WHERE room_id IS NOT NULL AND status != 'deleted'"
        )
        .fetch_one(&self.db)
        .await
        .map_err(|e| ChatError::database_error("get_global_message_stats_rooms", e))?
        .unwrap_or(0);

        // Requête pour les messages d'aujourd'hui
        let messages_today = sqlx::query_scalar!(
            "SELECT COUNT(*) FROM messages WHERE created_at >= CURRENT_DATE AND status != 'deleted'"
        )
        .fetch_one(&self.db)
        .await
        .map_err(|e| ChatError::database_error("get_global_message_stats_today", e))?
        .unwrap_or(0);

        // Requête pour les utilisateurs actifs aujourd'hui
        let active_users_today = sqlx::query_scalar!(
            "SELECT COUNT(DISTINCT author_id) FROM messages WHERE created_at >= CURRENT_DATE AND status != 'deleted'"
        )
        .fetch_one(&self.db)
        .await
        .map_err(|e| ChatError::database_error("get_global_message_stats_active_users", e))?
        .unwrap_or(0);

        Ok(GlobalMessageStats {
            total_messages: stats.total_messages.unwrap_or(0) as u64,
            total_room_messages: stats.total_room_messages.unwrap_or(0) as u64,
            total_dm_messages: stats.total_dm_messages.unwrap_or(0) as u64,
            total_users: stats.total_users.unwrap_or(0) as u64,
            total_rooms: total_rooms as u64,
            messages_today: messages_today as u64,
            active_users_today: active_users_today as u64,
        })
    }

    /// Nettoie les anciens messages (pour la maintenance)
    pub async fn cleanup_old_messages(&self, days_to_keep: i32) -> Result<u64> {
        let result = sqlx::query(
            &format!("DELETE FROM messages WHERE created_at < NOW() - INTERVAL '{} days' AND status = 'deleted'", days_to_keep)
        )
        .execute(&self.db)
        .await
        .map_err(|e| ChatError::database_error("cleanup_old_messages", e))?;

        Ok(result.rows_affected())
    }

    /// Sauvegarde en lot de messages (pour l'import/sync)
    pub async fn batch_save_messages(&self, messages: &[Message]) -> Result<()> {
        for message in messages {
            sqlx::query!(
                r#"
                INSERT INTO messages (
                    message_type, content, author_id, author_username, 
                    room_id, recipient_id, recipient_username,
                    created_at, status, is_pinned, is_edited, 
                    parent_message_id, thread_count
                ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
                "#,
                message.message_type.to_string(),
                message.content,
                message.author_id,
                message.author_username,
                message.room_id,
                message.recipient_id,
                message.recipient_username,
                message.created_at,
                message.status.to_string(),
                message.is_pinned,
                message.is_edited,
                message.parent_message_id,
                message.thread_count
            )
            .execute(&self.db)
            .await?;
        }
        Ok(())
    }

    /// Archive les anciens messages 
    pub async fn archive_old_messages(&self, days_to_archive: i32) -> Result<u64> {
        let result = sqlx::query!(
            r#"UPDATE messages SET status = 'Archived' 
               WHERE created_at < NOW() - INTERVAL '1 day' * $1"#,
            days_to_archive as f64
        )
        .execute(&self.db)
        .await?;

        Ok(result.rows_affected())
    }

    /// Restaure les messages archivés
    pub async fn restore_archived_messages(&self, message_ids: &[i64]) -> Result<u64> {
        let mut total_restored = 0u64;
        
        for &message_id in message_ids {
            let result = sqlx::query!(
                "UPDATE messages SET status = 'Sent' WHERE id = $1 AND status = 'Archived'",
                message_id
            )
            .execute(&self.db)
            .await?;
            
            total_restored += result.rows_affected();
        }

        Ok(total_restored)
    }
}

/// Représentation d'une conversation DM
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DMConversation {
    pub other_user_id: i32,
    pub other_username: String,
    pub last_message_at: DateTime<Utc>,
    pub unread_count: u32,
    pub last_message_preview: Option<String>,
}

/// Statistiques de messages
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MessageStats {
    pub total_sent: u64,
    pub room_messages: u64,
    pub dm_messages: u64,
    pub edited_messages: u64,
    pub pinned_messages: u64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct GlobalMessageStats {
    pub total_messages: u64,
    pub total_room_messages: u64,
    pub total_dm_messages: u64,
    pub total_users: u64,
    pub total_rooms: u64,
    pub messages_today: u64,
    pub active_users_today: u64,
}

impl MessageStore {
    /// Obtenir les statistiques de messages
    pub async fn get_message_stats(&self) -> Result<MessageStats> {
        // Statistiques des salons les plus actifs
        let top_rooms_rows = sqlx::query!(
            r#"
            SELECT room_id, COUNT(*) as message_count
            FROM messages 
            WHERE room_id IS NOT NULL 
            GROUP BY room_id 
            ORDER BY message_count DESC 
            LIMIT 10
            "#
        )
        .fetch_all(&self.db)
        .await?;

        let top_rooms: Vec<(String, i64)> = top_rooms_rows
            .into_iter()
            .map(|row| (row.room_id.unwrap_or_default(), row.message_count.unwrap_or(0)))
            .collect();

        // Utilisateurs les plus actifs
        let active_users_rows = sqlx::query!(
            r#"
            SELECT author_id, author_username, COUNT(*) as message_count
            FROM messages 
            WHERE created_at > NOW() - INTERVAL '30 days'
            GROUP BY author_id, author_username 
            ORDER BY message_count DESC 
            LIMIT 10
            "#
        )
        .fetch_all(&self.db)
        .await?;

        let active_users: Vec<(i64, String)> = active_users_rows
            .into_iter()
            .filter_map(|row| {
                row.author_id.map(|id| (id, row.author_username.unwrap_or_default()))
            })
            .collect();

        // Retourner des statistiques par défaut pour maintenir la compatibilité
        Ok(MessageStats {
            total_sent: top_rooms.len() as u64,
            room_messages: active_users.len() as u64,
            dm_messages: 0,
            edited_messages: 0,
            pinned_messages: 0,
        })
    }
} 