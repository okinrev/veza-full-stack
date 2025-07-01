use std::collections::HashMap;
use chrono::{DateTime, Utc};
use sqlx::{PgPool, Postgres, Transaction, Row};
use serde::{Serialize, Deserialize};
use crate::error::{ChatError, Result};
use crate::models::MessageReaction;
use uuid::Uuid;

/// Types de messages différenciés
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum MessageType {
    RoomMessage,
    DirectMessage,
    SystemMessage,
}

impl ToString for MessageType {
    fn to_string(&self) -> String {
        match self {
            MessageType::RoomMessage => "room".to_string(),
            MessageType::DirectMessage => "direct".to_string(),
            MessageType::SystemMessage => "system".to_string(),
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

impl ToString for MessageStatus {
    fn to_string(&self) -> String {
        match self {
            MessageStatus::Sent => "active".to_string(),
            MessageStatus::Delivered => "delivered".to_string(),
            MessageStatus::Read => "read".to_string(),
            MessageStatus::Edited => "edited".to_string(),
            MessageStatus::Deleted => "deleted".to_string(),
        }
    }
}

/// Message unifié avec séparation logique
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Message {
    pub id: i64,
    pub message_type: MessageType,
    pub content: String,
    pub author_id: i32,
    pub author_username: String,
    
    // Pour les messages de salon
    pub room_id: Option<String>,
    
    // Pour les messages directs
    pub recipient_id: Option<i32>,
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
    pub reactions: HashMap<String, Vec<i32>>, // emoji -> liste d'user_ids
    
    // Attachments
    pub attachments: Vec<String>,
    
    // Mentions
    pub mentions: Vec<i32>, // user_ids mentionnés
    
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
            message.id as i32,
            message.message_type.to_string(),
            message.content,
            message.author_id,
            message.author_username,
            message.room_id.map(|r| r.parse::<i32>().unwrap_or(0)),
            message.recipient_id,
            message.recipient_username,
            message.parent_message_id,
            message.thread_count,
            message.status.to_string(),
            message.is_pinned,
            message.is_edited,
            message.original_content,
            message.created_at.naive_utc(),
            message.updated_at
            )
            .execute(&self.db)
            .await
        .map_err(|e| ChatError::database_error("save_message", e))?;

        // Sauvegarder les mentions si elles existent
        if !message.mentions.is_empty() {
            for mention in &message.mentions {
            sqlx::query!(
                    "INSERT INTO message_mentions (message_id, user_id) VALUES ($1, $2)",
                    message.id as i32,
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
    pub async fn get_message(&self, message_id: i32) -> Result<Option<Message>> {
        let row = sqlx::query!(
            r#"SELECT m.*, ARRAY_AGG(mm.user_id) as mention_ids
            FROM messages m
            LEFT JOIN message_mentions mm ON m.id = mm.message_id
               WHERE m.id = $1
               GROUP BY m.id"#,
            message_id
        )
        .fetch_optional(&self.db)
        .await
        .map_err(|e| ChatError::database_error("get_message", e))?;

        match row {
            Some(row) => {
                let message = self.row_to_message(row).await?;
                Ok(Some(message))
            }
            None => Ok(None),
        }
    }

    /// Récupère les messages d'un salon avec pagination
    pub async fn get_room_messages(&self, room_id: Option<i32>, limit: i32, before_id: Option<i32>) -> Result<Vec<Message>> {
        let mut query = sqlx::QueryBuilder::new(
            r#"SELECT m.*, ARRAY_REMOVE(ARRAY_AGG(mm.user_id), NULL) as mention_ids
               FROM messages m
               LEFT JOIN message_mentions mm ON m.id = mm.message_id
               WHERE m.room_id = "#
        );
        
        query.push_bind(room_id);
        
        if let Some(before) = before_id {
            query.push(" AND m.id < ");
            query.push_bind(before);
        }
        
        query.push(" GROUP BY m.id ORDER BY m.created_at DESC LIMIT ");
        query.push_bind(limit);

        let rows = query
            .build()
            .fetch_all(&self.db)
            .await
            .map_err(|e| ChatError::database_error("get_room_messages", e))?;

        let mut result = Vec::new();
        for row in rows {
            result.push(self.row_to_message(row).await?);
        }

        Ok(result)
    }

    /// Récupère les messages directs entre deux utilisateurs
    pub async fn get_dm_messages(&self, user1_id: i32, user2_id: i32, limit: i32, before_id: Option<i32>) -> Result<Vec<Message>> {
        let mut query = sqlx::QueryBuilder::new(
            r#"SELECT m.*, ARRAY_AGG(mm.user_id) as mention_ids
               FROM messages m
               LEFT JOIN message_mentions mm ON m.id = mm.message_id
               WHERE ((m.author_id = "#
        );
        
        query.push_bind(user1_id);
        query.push(" AND m.recipient_id = ");
        query.push_bind(user2_id);
        query.push(") OR (m.author_id = ");
        query.push_bind(user2_id);
        query.push(" AND m.recipient_id = ");
        query.push_bind(user1_id);
        query.push("))");
        
        if let Some(before) = before_id {
            query.push(" AND m.id < ");
            query.push_bind(before);
        }
        
        query.push(" GROUP BY m.id ORDER BY m.created_at DESC LIMIT ");
        query.push_bind(limit);

        let messages = query
            .build()
            .fetch_all(&self.db)
            .await
            .map_err(|e| ChatError::database_error("get_dm_messages", e))?;

        let mut result = Vec::new();
        for row in messages {
            result.push(self.row_to_message(row).await?);
        }

        Ok(result)
    }

    /// Épingle un message
    pub async fn pin_message(&self, message_id: i32, room_id: Option<i32>) -> Result<()> {
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
    pub async fn unpin_message(&self, message_id: i32) -> Result<()> {
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
    pub async fn get_pinned_messages(&self, room_id: Option<i32>) -> Result<Vec<Message>> {
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
            result.push(self.row_to_message(row).await?);
        }

        Ok(result)
    }

    /// Marque les messages comme lus pour un utilisateur
    pub async fn mark_messages_as_read(&self, user_id: i32, conversation_id: Option<String>) -> Result<()> {
        // Vérifier que l'utilisateur peut marquer ces messages comme lus
        if let Some(conv_id) = &conversation_id {
            // Vérifier que l'utilisateur fait partie de cette conversation
            let is_participant = self.is_conversation_participant(user_id, conv_id).await?;
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
    pub async fn count_unread_messages(&self, user_id: i32) -> Result<HashMap<String, i64>> {
        let counts = sqlx::query!(
            r#"SELECT 
                   CASE 
                       WHEN m.room_id IS NOT NULL THEN CONCAT('room_', m.room_id)
                       ELSE CONCAT('dm_', LEAST(m.author_id, m.recipient_id), '_', GREATEST(m.author_id, m.recipient_id))
                   END as conversation_key,
                   COUNT(*) as unread_count
               FROM messages m
               LEFT JOIN message_read_status mrs ON m.id = mrs.message_id AND mrs.user_id = $1
               WHERE (m.room_id IS NOT NULL OR m.recipient_id = $1 OR m.author_id = $1)
                 AND mrs.message_id IS NULL
                 AND m.author_id != $1
               GROUP BY conversation_key"#,
            user_id
        )
        .fetch_all(&self.db)
        .await
        .map_err(|e| ChatError::database_error("count_unread_messages", e))?;

        let mut result = HashMap::new();
        for row in counts {
            if let (Some(key), Some(count)) = (row.conversation_key, row.unread_count) {
                result.insert(key, count);
            }
        }

        Ok(result)
    }

    /// Vérifie si un utilisateur participe à une conversation
    async fn is_conversation_participant(&self, user_id: i32, _conversation_id: &str) -> Result<bool> {
        // Pour l'instant, on suppose que l'utilisateur peut participer
        // TODO: Implémenter la vérification réelle
        Ok(true)
    }

    /// Ajoute une réaction à un message
    pub async fn add_reaction(&self, message_id: i32, user_id: i32, emoji: &str) -> Result<()> {
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
    pub async fn remove_reaction(&self, message_id: i32, user_id: i32, emoji: &str) -> Result<()> {
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
    pub async fn get_message_reactions(&self, message_id: i32) -> Result<Vec<MessageReaction>> {
        let reactions = sqlx::query_as!(
            MessageReaction,
            "SELECT * FROM message_reactions WHERE message_id = $1 ORDER BY created_at",
            message_id
        )
        .fetch_all(&self.db)
        .await
        .map_err(|e| ChatError::database_error("get_message_reactions", e))?;

        Ok(reactions)
    }

    /// Édite un message
    pub async fn edit_message(&self, message_id: i32, new_content: &str, editor_user_id: i32) -> Result<()> {
        // Récupérer le message existant
        let message = sqlx::query!(
            "SELECT author_id, content FROM messages WHERE id = $1",
            message_id
        )
        .fetch_optional(&self.db)
        .await
        .map_err(|e| ChatError::database_error("edit_message_fetch", e))?;

        let message = message.ok_or_else(|| ChatError::configuration_error("Message non trouvé"))?;

        // Vérifier que seul l'auteur peut éditer le message
        if message.author_id != Some(editor_user_id) {
            return Err(ChatError::unauthorized("Seul l'auteur peut éditer ce message"));
        }

        // Sauvegarder le contenu original si c'est la première édition
        let original_content = if message.content != new_content {
            Some(message.content)
        } else {
            None
        };

        // Mettre à jour le message
        sqlx::query!(
            r#"UPDATE messages 
               SET content = $1, is_edited = true, updated_at = NOW(),
                   original_content = COALESCE(original_content, $2)
               WHERE id = $3"#,
            new_content,
            original_content,
            message_id
        )
        .execute(&self.db)
        .await
        .map_err(|e| ChatError::database_error("edit_message_update", e))?;

        Ok(())
    }

    /// Supprime un message (soft delete)
    pub async fn delete_message(&self, message_id: i32, deleter_user_id: i32, is_moderator: bool) -> Result<()> {
        // Récupérer le message existant
        let message = sqlx::query!(
            "SELECT author_id FROM messages WHERE id = $1",
            message_id
        )
        .fetch_optional(&self.db)
            .await
        .map_err(|e| ChatError::database_error("delete_message_fetch", e))?;

        let message = message.ok_or_else(|| ChatError::configuration_error("Message non trouvé"))?;

        // Vérifier les permissions
        if message.author_id != Some(deleter_user_id) && !is_moderator {
            return Err(ChatError::unauthorized("Seul l'auteur ou un modérateur peut supprimer ce message"));
        }

        // Soft delete du message
        sqlx::query!(
            "UPDATE messages SET status = 'deleted', updated_at = NOW() WHERE id = $1",
            message_id
        )
        .execute(&self.db)
        .await
        .map_err(|e| ChatError::database_error("delete_message_update", e))?;

        Ok(())
    }

    /// Recherche des messages avec un terme
    pub async fn search_messages(&self, query: &str, user_id: i32, limit: i32) -> Result<Vec<Message>> {
        let messages = sqlx::query!(
            r#"SELECT m.*, ARRAY_AGG(mm.user_id) as mention_ids
            FROM messages m
               LEFT JOIN message_mentions mm ON m.id = mm.message_id
               WHERE (m.content ILIKE $1 OR m.author_username ILIKE $1)
                 AND (m.room_id IS NOT NULL OR m.author_id = $2 OR m.recipient_id = $2)
                 AND m.status = 'active'
               GROUP BY m.id
            ORDER BY m.created_at DESC
               LIMIT $3"#,
            format!("%{}%", query),
            user_id,
            limit
        )
                .fetch_all(&self.db)
                .await
        .map_err(|e| ChatError::database_error("search_messages", e))?;

        let mut result = Vec::new();
        for row in messages {
            result.push(self.row_to_message(row).await?);
        }

        Ok(result)
    }

    /// Statistiques des messages pour un utilisateur
    pub async fn get_user_message_stats(&self, user_id: i32) -> Result<MessageStats> {
        let stats = sqlx::query!(
            r#"SELECT 
                   COUNT(*) as total_sent,
                   COUNT(CASE WHEN room_id IS NOT NULL THEN 1 END) as room_messages,
                   COUNT(CASE WHEN room_id IS NULL THEN 1 END) as dm_messages,
                   COUNT(CASE WHEN is_edited THEN 1 END) as edited_messages,
                   COUNT(CASE WHEN is_pinned THEN 1 END) as pinned_messages
               FROM messages 
               WHERE author_id = $1 AND status = 'active'"#,
            user_id
        )
        .fetch_one(&self.db)
        .await
        .map_err(|e| ChatError::database_error("get_user_message_stats", e))?;

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
        let stats = sqlx::query!(
            r#"SELECT 
                   COUNT(*) as total_messages,
                   COUNT(DISTINCT author_id) as unique_authors,
                   COUNT(CASE WHEN room_id IS NOT NULL THEN 1 END) as room_messages,
                   COUNT(CASE WHEN room_id IS NULL THEN 1 END) as dm_messages,
                   COUNT(CASE WHEN is_edited THEN 1 END) as edited_messages,
                   COUNT(CASE WHEN status = 'deleted' THEN 1 END) as deleted_messages
               FROM messages"#
        )
        .fetch_one(&self.db)
        .await
        .map_err(|e| ChatError::database_error("get_global_message_stats", e))?;

        Ok(GlobalMessageStats {
            total_messages: stats.total_messages.unwrap_or(0) as u64,
            unique_authors: stats.total_users as u64,
            room_messages: stats.total_room_messages as u64,
            dm_messages: stats.total_dm_messages as u64,
            edited_messages: 0, // Champ non disponible dans la base
            deleted_messages: 0, // Champ non disponible dans la base
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

    /// Convertit une ligne SQL en objet Message
    async fn row_to_message(&self, row: sqlx::postgres::PgRow) -> Result<Message> {
        use sqlx::Row;
        
        let message_id: i64 = row.get("id")?;
        let mention_ids: Option<Vec<i32>> = row.get("mention_ids");
        let mentions: Vec<i32> = mention_ids.unwrap_or_default();
        
        let message_type_str: String = row.get("message_type")?;
        let message_type = match message_type_str.as_str() {
            "room" => MessageType::RoomMessage,
            "direct" => MessageType::DirectMessage,
            "system" => MessageType::SystemMessage,
            _ => MessageType::RoomMessage,
        };

        let content: String = row.get("content")?;
        let author_id: i32 = row.get("author_id")?;
        let author_username: String = row.get("author_username")?;
        let room_id: Option<String> = row.get("room_id")?;
        let recipient_id: Option<i32> = row.get("recipient_id")?;
        let recipient_username: Option<String> = row.get("recipient_username")?;
        let created_at: DateTime<Utc> = row.get("created_at")?;
        let updated_at: Option<DateTime<Utc>> = row.get("updated_at").ok();
        
        let status_str: String = row.get("status")?;
        let status = match status_str.as_str() {
            "active" => MessageStatus::Sent,
                "deleted" => MessageStatus::Deleted,
            "archived" => MessageStatus::Deleted,
                _ => MessageStatus::Sent,
        };

        let is_pinned: bool = row.get("is_pinned").unwrap_or(false);
        let is_edited: bool = row.get("is_edited").unwrap_or(false);
        let original_content: Option<String> = row.get("original_content").unwrap_or(None);
        let parent_message_id: Option<i64> = row.get("parent_message_id").unwrap_or(None);
        let thread_count: i32 = row.get("thread_count").unwrap_or(0);

        Ok(Message {
            id: message_id,
            message_type,
            content,
            author_id,
            author_username,
            room_id,
            recipient_id,
            recipient_username,
            created_at,
            updated_at,
            status,
            is_pinned,
            is_edited,
            original_content,
            parent_message_id,
            thread_count,
            reactions: HashMap::new(),
            attachments: Vec::new(),
            mentions,
            is_flagged: row.get("is_flagged").unwrap_or(false),
            moderation_notes: row.get("moderation_notes").unwrap_or(None),
        })
    }

    /// Sauvegarde en lot de messages (pour l'import/sync)
    pub async fn batch_save_messages(&self, messages: &[Message]) -> Result<()> {
        let mut tx = self.db.begin().await.map_err(|e| ChatError::database_error("batch_save_begin", e))?;

        for message in messages {
            sqlx::query!(
                r#"INSERT INTO messages 
                   (id, message_type, content, author_id, author_username, room_id, recipient_id, recipient_username, 
                    parent_message_id, thread_count, status, is_pinned, is_edited, original_content, created_at, updated_at) 
                   VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
                   ON CONFLICT (id) DO NOTHING"#,
                message.id as i32,
                message.message_type.to_string(),
                message.content,
                message.author_id,
                message.author_username,
                message.room_id.map(|r| r.parse::<i32>().unwrap_or(0)),
                message.recipient_id,
                message.recipient_username,
                message.parent_message_id,
                message.thread_count,
                message.status.to_string(),
                message.is_pinned,
                message.is_edited,
                message.original_content,
                message.created_at.naive_utc(),
                message.updated_at
            )
            .execute(&mut *tx)
            .await
            .map_err(|e| ChatError::database_error("batch_save_insert", e))?;
        }

        tx.commit().await.map_err(|e| ChatError::database_error("batch_save_commit", e))?;
        Ok(())
    }

    /// Archive les anciens messages 
    pub async fn archive_old_messages(&self, days_to_archive: i32) -> Result<u64> {
        let result = sqlx::query(
            &format!("UPDATE messages SET status = 'archived' WHERE created_at < NOW() - INTERVAL '{} days' AND status = 'active'", days_to_archive)
        )
        .execute(&self.db)
        .await
        .map_err(|e| ChatError::database_error("archive_old_messages", e))?;

        Ok(result.rows_affected())
    }

    /// Restaure les messages archivés
    pub async fn restore_archived_messages(&self, message_ids: &[i32]) -> Result<u64> {
        let result = sqlx::query!(
            "UPDATE messages SET status = 'active' WHERE id = ANY($1) AND status = 'archived'",
            message_ids
        )
        .execute(&self.db)
        .await
        .map_err(|e| ChatError::database_error("restore_archived_messages", e))?;

        Ok(result.rows_affected())
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
        let total_messages = sqlx::query_scalar!(
            "SELECT COUNT(*) FROM messages WHERE status != 'deleted'"
        )
        .fetch_one(&self.db)
        .await
        .map_err(|e| ChatError::database_error("database_operation", e))?
        .unwrap_or(0);

        let room_messages = sqlx::query_scalar!(
            "SELECT COUNT(*) FROM messages WHERE message_type = 'room_message' AND status != 'deleted'"
        )
        .fetch_one(&self.db)
        .await
        .map_err(|e| ChatError::database_error("database_operation", e))?
        .unwrap_or(0);

        let direct_messages = sqlx::query_scalar!(
            "SELECT COUNT(*) FROM messages WHERE message_type = 'direct_message' AND status != 'deleted'"
        )
        .fetch_one(&self.db)
        .await
        .map_err(|e| ChatError::database_error("database_operation", e))?
        .unwrap_or(0);

        let messages_today = sqlx::query_scalar!(
            "SELECT COUNT(*) FROM messages WHERE created_at >= CURRENT_DATE AND status != 'deleted'"
        )
        .fetch_one(&self.db)
        .await
        .map_err(|e| ChatError::database_error("database_operation", e))?
        .unwrap_or(0);

        let messages_this_week = sqlx::query_scalar!(
            "SELECT COUNT(*) FROM messages WHERE created_at >= CURRENT_DATE - INTERVAL '7 days' AND status != 'deleted'"
        )
        .fetch_one(&self.db)
        .await
        .map_err(|e| ChatError::database_error("database_operation", e))?
        .unwrap_or(0);

        // Top salons par nombre de messages
        let top_rooms_rows = sqlx::query!(
            r#"
            SELECT room_id, COUNT(*) as message_count
            FROM messages 
            WHERE message_type = 'room_message' 
              AND status != 'deleted'
              AND room_id IS NOT NULL
            GROUP BY room_id
            ORDER BY message_count DESC
            LIMIT 10
            "#
        )
        .fetch_all(&self.db)
        .await
        .map_err(|e| ChatError::database_error("database_operation", e))?;

        let top_rooms = top_rooms_rows.into_iter()
            .map(|row| (row.get("room_id").unwrap_or_default().unwrap_or_else(|| "unknown".to_string()), row.message_count.unwrap_or(0)))
            .collect();

        // Utilisateurs les plus actifs
        let active_users_rows = sqlx::query!(
            r#"
            SELECT author_id, author_username, COUNT(*) as message_count
            FROM messages 
            WHERE status != 'deleted'
              AND created_at >= CURRENT_DATE - INTERVAL '30 days'
            GROUP BY author_id, author_username
            ORDER BY message_count DESC
            LIMIT 10
            "#
        )
        .fetch_all(&self.db)
        .await
        .map_err(|e| ChatError::database_error("database_operation", e))?;

        let active_users = active_users_rows.into_iter()
            .map(|row| (row.get("author_id").unwrap_or_default(), row.get("author_username").unwrap_or_default(), row.message_count.unwrap_or(0)))
            .collect();

        // Récupérer les messages édités
        let edited_messages = sqlx::query_scalar!(
            "SELECT COUNT(*) FROM messages WHERE is_edited = true AND status != 'deleted'"
        )
        .fetch_one(&self.db)
        .await
        .map_err(|e| ChatError::database_error("database_operation", e))?
        .unwrap_or(0);

        // Récupérer les messages épinglés
        let pinned_messages = sqlx::query_scalar!(
            "SELECT COUNT(*) FROM messages WHERE is_pinned = true AND status != 'deleted'"
        )
        .fetch_one(&self.db)
        .await
        .map_err(|e| ChatError::database_error("database_operation", e))?
        .unwrap_or(0);

        Ok(MessageStats {
            total_sent: total_messages as u64,
            room_messages: room_messages as u64,
            dm_messages: direct_messages as u64,
            edited_messages: edited_messages as u64,
            pinned_messages: pinned_messages as u64,
        })
    }
} 