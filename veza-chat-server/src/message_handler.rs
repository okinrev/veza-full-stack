//! Module de gestion des messages avec filtrage de contenu et sécurité
//! 
//! Ce module fournit une couche de haut niveau pour traiter les messages entrants,
//! appliquer les filtres de sécurité et déléguer aux modules métier appropriés.

use std::sync::Arc;
use crate::error::{ChatError, Result};
use crate::hub::common::ChatHub;
use crate::permissions::{Role, Permission, check_permission};
use crate::security::{EnhancedSecurity, SecurityAction, ContentFilter};
use serde_json::json;
use tokio::sync::mpsc::UnboundedSender;
use tokio_tungstenite::tungstenite::Message;
use tracing::{info, warn};

/// Gestionnaire centralisé pour tous les types de messages
pub struct MessageHandler {
    security: EnhancedSecurity,
    content_filter: ContentFilter,
    hub: Arc<ChatHub>,
}

impl MessageHandler {
    pub fn new(hub: Arc<ChatHub>) -> Result<Self> {
        Ok(Self {
            security: EnhancedSecurity::new()?,
            content_filter: ContentFilter::new()?,
            hub,
        })
    }

    /// Gère les messages de salon avec permissions
    pub async fn handle_room_message(
        &mut self,
        user_id: i32,
        username: &str,
        room: &str,
        content: &str,
        session_token: &str,
        user_ip: &str,
        parent_id: Option<i64>,
    ) -> Result<()> {
        // Validation de sécurité
        self.security.validate_request(
            user_id,
            user_ip,
            session_token,
            &SecurityAction::SendMessage,
            Some(content)
        ).await?;

        // Filtrage du contenu
        let clean_room = self.content_filter.validate_content(room)?;
        let clean_content = self.content_filter.validate_content(content)?;

        info!(
            user_id = %user_id,
            username = %username,
            room = %clean_room,
            content_length = %clean_content.len(),
            "📝 Message de salon filtré et validé"
        );

        // Délégation à la logique métier - Conversion de types
        let room_id = self.get_room_id_by_name(&clean_room).await?;
        crate::hub::channels::send_room_message(&self.hub, room_id, user_id as i64, username, &clean_content, parent_id, None).await?;
        Ok(())
    }

    /// Gère les messages directs avec permissions
    pub async fn handle_direct_message(
        &mut self,
        from_user: i32,
        from_username: &str,
        to_user: i32,
        content: &str,
        session_token: &str,
        user_ip: &str,
        parent_id: Option<i64>,
    ) -> Result<()> {
        // Validation de sécurité
        self.security.validate_request(
            from_user,
            user_ip,
            session_token,
            &SecurityAction::SendDM,
            Some(content)
        ).await?;

        // Filtrage du contenu
        let clean_content = self.content_filter.validate_content(content)?;

        info!(
            from_user = %from_user,
            from_username = %from_username,
            to_user = %to_user,
            content_length = %clean_content.len(),
            "💬 Message direct filtré et validé"
        );

        // Délégation à la logique métier - Conversion de types
        let conversation_id = self.get_or_create_conversation(from_user as i64, to_user as i64).await?;
        crate::hub::direct_messages::send_dm_message(&self.hub, conversation_id, from_user as i64, from_username, &clean_content, parent_id, None).await?;
        Ok(())
    }

    /// Gère la jointure d'un salon avec permissions
    pub async fn handle_join_room(
        &mut self,
        user_id: i32,
        username: &str,
        room: &str,
        session_token: &str,
        user_ip: &str,
    ) -> Result<()> {
        // Validation de sécurité
        self.security.validate_request(
            user_id,
            user_ip,
            session_token,
            &SecurityAction::JoinRoom,
            None
        ).await?;

        // Validation du nom de salon
        let clean_room = self.content_filter.validate_content(room)?;

        // Vérification que le salon existe ou peut être créé
        // Pour l'instant, on suppose que le salon existe
        let room_exists = true;
        
        if !room_exists {
            return Err(ChatError::not_found("Salon", &clean_room));
            }

        info!(
            user_id = %user_id,
            username = %username,
            room = %clean_room,
            "🚪 Jointure de salon validée"
        );

        // Délégation à la logique métier - Conversion de types et ID de salon
        let room_id = self.get_room_id_by_name(&clean_room).await?;
        crate::hub::channels::join_room(&self.hub, room_id, user_id as i64).await?;

        // Envoi de confirmation
        Ok(())
    }

    /// Gère la récupération d'historique avec permissions
    pub async fn handle_room_history(
        &mut self,
        user_id: i32,
        user_role: &Role,
        room: &str,
        limit: Option<i32>,
        session_token: &str,
        user_ip: &str,
    ) -> Result<Vec<crate::hub::channels::RoomMessage>> {
        // Validation de sécurité pour la lecture
        self.security.validate_request(
            user_id,
            user_ip,
            session_token,
            &SecurityAction::SendMessage, // Approximation
            None
        ).await?;

        // Validation du nom de salon
        let clean_room = self.content_filter.validate_content(room)?;
        let limit = limit.unwrap_or(50).min(100); // Limiter à 100 messages max

        // Vérification des permissions de lecture
        if !self.can_read_room_history(user_id, user_role, &clean_room).await? {
            return Err(ChatError::unauthorized("Lecture de l'historique du salon"));
        }

        // Délégation à la logique métier - Conversion de types
        let room_id = self.get_room_id_by_name(&clean_room).await?;
        let messages = crate::hub::channels::fetch_room_history(&self.hub, room_id, user_id as i64, limit.into(), None).await?;

        // Envoi de la réponse
        info!(
            user_id = %user_id,
            room = %clean_room,
            message_count = %messages.len(),
            "📚 Historique de salon récupéré"
        );

        Ok(messages)
    }

    /// Gère la récupération d'historique DM avec permissions
    pub async fn handle_dm_history(
        &mut self,
        user_id: i32,
        with_user: i32,
        limit: Option<i32>,
        session_token: &str,
        user_ip: &str,
    ) -> Result<Vec<crate::hub::direct_messages::DmMessage>> {
        // Validation de sécurité
        self.security.validate_request(
            user_id,
            user_ip,
            session_token,
            &SecurityAction::SendDM,
            None
        ).await?;

        let limit = limit.unwrap_or(50).min(100);

        // Vérification que l'utilisateur peut lire cette conversation
        if !self.can_read_dm_conversation(user_id, with_user).await? {
            return Err(ChatError::unauthorized("Lecture de conversation privée"));
        }

        // Délégation à la logique métier - Conversion de types
        let conversation_id = self.get_or_create_conversation(user_id as i64, with_user as i64).await?;
        let messages = crate::hub::direct_messages::fetch_history(&self.hub, conversation_id, user_id as i64, limit.into(), None).await?;

        // Envoi de la réponse
        info!(
            user_id = %user_id,
            with_user = %with_user,
            message_count = %messages.len(),
            "💬 Historique DM récupéré"
        );

        Ok(messages)
    }

    /// Vérifie si un utilisateur peut lire l'historique d'un salon
    async fn can_read_room_history(&self, user_id: i32, user_role: &Role, _room: &str) -> Result<bool> {
        // Logique simple : les admins et modérateurs peuvent tout lire
        match user_role {
            Role::Admin | Role::Moderator => Ok(true),
            Role::User => {
                // Les utilisateurs normaux peuvent lire les salons dont ils sont membres
                // TODO: Vérifier l'appartenance au salon
                Ok(true) // Temporaire
            }
            _ => Ok(false),
        }
    }

    /// Vérifie si un utilisateur peut lire une conversation DM
    async fn can_read_dm_conversation(&self, user_id: i32, with_user: i32) -> Result<bool> {
        // Un utilisateur peut lire ses propres conversations
        if user_id == with_user {
            return Ok(true);
        }

        // TODO: Vérifier si les utilisateurs ont une conversation existante
        // ou si l'un des deux autorise les messages de inconnus
        Ok(false)
    }

    /// Récupère ou crée une conversation entre deux utilisateurs
    async fn get_or_create_conversation(&self, user1_id: i64, user2_id: i64) -> Result<i64> {
        let conversation = crate::hub::direct_messages::get_or_create_dm_conversation(&self.hub, user1_id, user2_id).await?;
        Ok(conversation.id)
    }

    /// Récupère l'ID d'un salon par son nom
    async fn get_room_id_by_name(&self, room_name: &str) -> Result<i64> {
        // Pour l'instant, retourne un ID fictif
        // TODO: Implémenter la recherche d'ID de salon par nom
        Ok(1)
    }
} 