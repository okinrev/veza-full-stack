//! Store de messages simple et fonctionnel

use std::collections::HashMap;
use chrono::{DateTime, Utc};
use serde::{Serialize, Deserialize};
use tokio::sync::RwLock;
use crate::error::{ChatError, Result};

/// Message simple pour test
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SimpleMessage {
    pub id: i32,
    pub content: String,
    pub author: String,
    pub timestamp: DateTime<Utc>,
    pub room: Option<String>,
    pub is_direct: bool,
}

/// Store en mémoire pour les tests
pub struct SimpleMessageStore {
    messages: RwLock<Vec<SimpleMessage>>,
    next_id: RwLock<i32>,
}

impl SimpleMessageStore {
    pub fn new() -> Self {
        Self {
            messages: RwLock::new(Vec::new()),
            next_id: RwLock::new(1),
        }
    }

    /// Envoi d'un message simple
    pub async fn send_simple_message(
        &self,
        content: &str,
        author: &str,
        room: Option<&str>,
        is_direct: bool,
    ) -> Result<i32> {
        let mut next_id = self.next_id.write().await;
        let id = *next_id;
        *next_id += 1;

        let message = SimpleMessage {
            id,
            content: content.to_string(),
            author: author.to_string(),
            timestamp: Utc::now(),
            room: room.map(|s| s.to_string()),
            is_direct,
        };

        let mut messages = self.messages.write().await;
        messages.push(message);

        Ok(id)
    }

    /// Récupération des messages d'un salon
    pub async fn get_room_messages(&self, room_name: &str, limit: i32) -> Result<Vec<SimpleMessage>> {
        let messages = self.messages.read().await;
        let filtered: Vec<SimpleMessage> = messages
            .iter()
            .filter(|msg| {
                if let Some(ref msg_room) = msg.room {
                    msg_room == room_name && !msg.is_direct
                } else {
                    false
                }
            })
            .take(limit as usize)
            .cloned()
            .collect();

        Ok(filtered)
    }

    /// Récupération des messages directs
    pub async fn get_direct_messages(&self, user1: &str, user2: &str, limit: i32) -> Result<Vec<SimpleMessage>> {
        let messages = self.messages.read().await;
        let filtered: Vec<SimpleMessage> = messages
            .iter()
            .filter(|msg| {
                msg.is_direct && 
                ((msg.author == user1) || (msg.author == user2))
            })
            .take(limit as usize)
            .cloned()
            .collect();

        Ok(filtered)
    }

    /// Autres méthodes simplifiées
    pub async fn pin_message(&self, _message_id: i32) -> Result<()> { Ok(()) }
    pub async fn message_exists(&self, message_id: i32) -> Result<bool> { 
        let messages = self.messages.read().await;
        Ok(messages.iter().any(|msg| msg.id == message_id))
    }
    pub async fn delete_message(&self, message_id: i32) -> Result<()> {
        let mut messages = self.messages.write().await;
        messages.retain(|msg| msg.id != message_id);
        Ok(())
    }
    pub async fn edit_message(&self, message_id: i32, new_content: &str) -> Result<()> {
        let mut messages = self.messages.write().await;
        if let Some(msg) = messages.iter_mut().find(|msg| msg.id == message_id) {
            msg.content = new_content.to_string();
            Ok(())
        } else {
            Err(ChatError::not_found("message", &message_id.to_string()))
        }
    }
    pub async fn add_reaction(&self, _message_id: i32, _user_id: i32, _emoji: &str) -> Result<()> { Ok(()) }
    pub async fn remove_reaction(&self, _message_id: i32, _user_id: i32, _emoji: &str) -> Result<()> { Ok(()) }
    pub async fn mark_as_read(&self, _user_id: i32, _conversation_id: &str) -> Result<()> { Ok(()) }
    pub async fn count_unread(&self, _user_id: i32) -> Result<i64> { Ok(0) }
    pub async fn count_unread_dms(&self, _user_id: i32) -> Result<i64> { Ok(0) }
    pub async fn count_unread_mentions(&self, _user_id: i32) -> Result<i64> { Ok(0) }
    pub async fn count_reactions(&self, _message_id: i32) -> Result<i64> { Ok(0) }
} 