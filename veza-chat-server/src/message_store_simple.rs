use std::collections::HashMap;
use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use sqlx::{PgPool, FromRow};
use crate::error::{ChatError, Result};

// Modèle de message simplifié correspondant exactement au schéma unifié
#[derive(Debug, Clone, Serialize, Deserialize, FromRow)]
pub struct SimpleMessage {
    pub id: i32,
    pub author_id: Option<i32>,
    pub author_username: Option<String>,
    pub recipient_id: Option<i32>,
    pub recipient_username: Option<String>,
    pub room: Option<String>,
    pub room_id: Option<i32>,
    pub content: String,
    pub created_at: Option<chrono::NaiveDateTime>,
    pub message_type: Option<String>,
    pub is_pinned: Option<bool>,
    pub is_edited: Option<bool>,
    pub original_content: Option<String>,
    pub status: Option<String>,
    pub parent_message_id: Option<i64>,
    pub updated_at: Option<DateTime<Utc>>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MessageStats {
    pub total_messages: u64,
    pub room_messages: u64,
    pub direct_messages: u64,
    pub today_messages: u64,
}

/// Store de messages simple et fonctionnel
pub struct SimpleMessageStore {
    db: PgPool,
}

impl SimpleMessageStore {
    pub fn new(db: PgPool) -> Self {
        Self { db }
    }

    /// Envoie un message dans un salon
    pub async fn send_room_message(
        &self,
        room_name: &str,
        user_id: i32,
        username: &str,
        content: &str,
    ) -> Result<i32> {
        let result = sqlx::query!(
            r#"INSERT INTO messages (author_id, author_username, room, content, message_type) 
               VALUES ($1, $2, $3, $4, 'text') RETURNING id"#,
            user_id,
            username,
            room_name,
            content
        )
        .fetch_one(&self.db)
        .await
        .map_err(|e| ChatError::database(&format!("Erreur lors de l'envoi du message: {}", e)))?;

        Ok(result.id)
    }

    /// Envoie un message direct
    pub async fn send_direct_message(
        &self,
        from_user_id: i32,
        from_username: &str,
        to_user_id: i32,
        to_username: &str,
        content: &str,
    ) -> Result<i32> {
        let result = sqlx::query!(
            r#"INSERT INTO messages (author_id, author_username, recipient_id, recipient_username, content, message_type) 
               VALUES ($1, $2, $3, $4, $5, 'text') RETURNING id"#,
            from_user_id,
            from_username,
            to_user_id,
            to_username,
            content
        )
        .fetch_one(&self.db)
        .await
        .map_err(|e| ChatError::database(&format!("Erreur lors de l'envoi du message direct: {}", e)))?;

        Ok(result.id)
    }

    /// Récupère les messages d'un salon
    pub async fn get_room_messages(&self, room_name: &str, limit: i32) -> Result<Vec<SimpleMessage>> {
        let messages = sqlx::query_as!(
            SimpleMessage,
            r#"SELECT id, author_id, author_username, recipient_id, recipient_username, 
                      room, room_id, content, created_at, message_type, is_pinned, 
                      is_edited, original_content, status, parent_message_id, updated_at
               FROM messages 
               WHERE room = $1 AND status != 'deleted'
               ORDER BY created_at DESC 
               LIMIT $2"#,
            room_name,
            limit as i64
        )
        .fetch_all(&self.db)
        .await
        .map_err(|e| ChatError::database(&format!("Erreur lors de la récupération des messages: {}", e)))?;

        Ok(messages)
    }

    /// Récupère les messages directs entre deux utilisateurs
    pub async fn get_direct_messages(&self, user1_id: i32, user2_id: i32, limit: i32) -> Result<Vec<SimpleMessage>> {
        let messages = sqlx::query_as!(
            SimpleMessage,
            r#"SELECT id, author_id, author_username, recipient_id, recipient_username, 
                      room, room_id, content, created_at, message_type, is_pinned, 
                      is_edited, original_content, status, parent_message_id, updated_at
               FROM messages 
               WHERE ((author_id = $1 AND recipient_id = $2) OR (author_id = $2 AND recipient_id = $1))
                 AND room IS NULL 
                 AND status != 'deleted'
               ORDER BY created_at DESC 
               LIMIT $3"#,
            user1_id,
            user2_id,
            limit as i64
        )
        .fetch_all(&self.db)
        .await
        .map_err(|e| ChatError::database(&format!("Erreur lors de la récupération des messages directs: {}", e)))?;

        Ok(messages)
    }

    /// Épingle/désépingle un message
    pub async fn toggle_pin_message(&self, message_id: i32, pinned: bool) -> Result<()> {
        sqlx::query!(
            "UPDATE messages SET is_pinned = $1, updated_at = NOW() WHERE id = $2",
            pinned,
            message_id
        )
        .execute(&self.db)
        .await
        .map_err(|e| ChatError::database(&format!("Erreur lors de l'épinglage: {}", e)))?;

        Ok(())
    }

    /// Supprime un message (soft delete)
    pub async fn delete_message(&self, message_id: i32, user_id: i32) -> Result<()> {
        // Vérifier que l'utilisateur peut supprimer ce message
        let message = sqlx::query!(
            "SELECT author_id FROM messages WHERE id = $1",
            message_id
        )
        .fetch_optional(&self.db)
        .await
        .map_err(|e| ChatError::database(&format!("Erreur lors de la vérification: {}", e)))?;

        if let Some(msg) = message {
            if msg.author_id != Some(user_id) {
                return Err(ChatError::unauthorized("Vous ne pouvez pas supprimer ce message"));
            }
        } else {
            return Err(ChatError::not_found("Message non trouvé"));
        }

        sqlx::query!(
            "UPDATE messages SET status = 'deleted', updated_at = NOW() WHERE id = $1",
            message_id
        )
        .execute(&self.db)
        .await
        .map_err(|e| ChatError::database(&format!("Erreur lors de la suppression: {}", e)))?;

        Ok(())
    }

    /// Modifie un message
    pub async fn edit_message(&self, message_id: i32, new_content: &str, user_id: i32) -> Result<()> {
        // Récupérer le message avec son contenu original
        let message = sqlx::query!(
            "SELECT author_id, content FROM messages WHERE id = $1",
            message_id
        )
        .fetch_optional(&self.db)
        .await
        .map_err(|e| ChatError::database(&format!("Erreur lors de la vérification: {}", e)))?;

        if let Some(msg) = message {
            if msg.author_id != Some(user_id) {
                return Err(ChatError::unauthorized("Vous ne pouvez pas modifier ce message"));
            }

            // Sauvegarder le contenu original si c'est la première modification
            sqlx::query!(
                r#"UPDATE messages 
                   SET content = $1, 
                       is_edited = true, 
                       original_content = COALESCE(original_content, $2),
                       updated_at = NOW() 
                   WHERE id = $3"#,
                new_content,
                msg.content,
                message_id
            )
            .execute(&self.db)
            .await
            .map_err(|e| ChatError::database(&format!("Erreur lors de la modification: {}", e)))?;

            Ok(())
        } else {
            Err(ChatError::not_found("Message non trouvé"))
        }
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
        .map_err(|e| ChatError::database(&format!("Erreur lors de la vérification: {}", e)))?;

        if existing.is_some() {
            return Err(ChatError::bad_request("Réaction déjà ajoutée"));
        }

        sqlx::query!(
            "INSERT INTO message_reactions (message_id, user_id, emoji) VALUES ($1, $2, $3)",
            message_id,
            user_id,
            emoji
        )
        .execute(&self.db)
        .await
        .map_err(|e| ChatError::database(&format!("Erreur lors de l'ajout de la réaction: {}", e)))?;

        Ok(())
    }

    /// Retire une réaction d'un message
    pub async fn remove_reaction(&self, message_id: i32, user_id: i32, emoji: &str) -> Result<()> {
        let result = sqlx::query!(
            "DELETE FROM message_reactions WHERE message_id = $1 AND user_id = $2 AND emoji = $3",
            message_id,
            user_id,
            emoji
        )
        .execute(&self.db)
        .await
        .map_err(|e| ChatError::database(&format!("Erreur lors de la suppression: {}", e)))?;

        if result.rows_affected() == 0 {
            return Err(ChatError::not_found("Réaction non trouvée"));
        }

        Ok(())
    }

    /// Marque les messages d'un salon comme lus
    pub async fn mark_room_as_read(&self, room_name: &str, user_id: i32) -> Result<()> {
        sqlx::query!(
            r#"INSERT INTO message_read_status (user_id, message_id)
               SELECT $1, m.id
               FROM messages m
               WHERE m.room = $2 AND m.status != 'deleted'
               ON CONFLICT (user_id, message_id) DO NOTHING"#,
            user_id,
            room_name
        )
        .execute(&self.db)
        .await
        .map_err(|e| ChatError::database(&format!("Erreur lors du marquage comme lu: {}", e)))?;

        Ok(())
    }

    /// Récupère les statistiques des messages
    pub async fn get_message_stats(&self) -> Result<MessageStats> {
        let total_messages = sqlx::query_scalar!(
            "SELECT COUNT(*) FROM messages WHERE status != 'deleted'"
        )
        .fetch_one(&self.db)
        .await
        .map_err(|e| ChatError::database(&format!("Erreur lors du comptage: {}", e)))?
        .unwrap_or(0) as u64;

        let room_messages = sqlx::query_scalar!(
            "SELECT COUNT(*) FROM messages WHERE room IS NOT NULL AND status != 'deleted'"
        )
        .fetch_one(&self.db)
        .await
        .map_err(|e| ChatError::database(&format!("Erreur lors du comptage: {}", e)))?
        .unwrap_or(0) as u64;

        let direct_messages = sqlx::query_scalar!(
            "SELECT COUNT(*) FROM messages WHERE room IS NULL AND recipient_id IS NOT NULL AND status != 'deleted'"
        )
        .fetch_one(&self.db)
        .await
        .map_err(|e| ChatError::database(&format!("Erreur lors du comptage: {}", e)))?
        .unwrap_or(0) as u64;

        let today_messages = sqlx::query_scalar!(
            "SELECT COUNT(*) FROM messages WHERE created_at >= CURRENT_DATE AND status != 'deleted'"
        )
        .fetch_one(&self.db)
        .await
        .map_err(|e| ChatError::database(&format!("Erreur lors du comptage: {}", e)))?
        .unwrap_or(0) as u64;

        Ok(MessageStats {
            total_messages,
            room_messages,
            direct_messages,
            today_messages,
        })
    }
} 