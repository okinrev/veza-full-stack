//! Module d'authentification pour le serveur de chat
//! 
//! Gère l'authentification des utilisateurs, les sessions et les rôles

use serde::{Deserialize, Serialize};
use chrono::{DateTime, Utc};
use std::collections::HashMap;
use crate::error::{ChatError, Result};

/// Rôles des utilisateurs dans le système de chat
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
pub enum Role {
    /// Utilisateur standard
    User,
    /// Modérateur - peut modérer les messages et gérer les salons
    Moderator,
    /// Administrateur - accès complet au système
    Admin,
    /// Utilisateur banni - accès restreint
    Banned,
}

impl Role {
    /// Vérifie si le rôle a les permissions d'administrateur
    pub fn is_admin(&self) -> bool {
        matches!(self, Role::Admin)
    }

    /// Vérifie si le rôle a les permissions de modérateur ou plus
    pub fn is_moderator_or_above(&self) -> bool {
        matches!(self, Role::Admin | Role::Moderator)
    }

    /// Vérifie si l'utilisateur est banni
    pub fn is_banned(&self) -> bool {
        matches!(self, Role::Banned)
    }

    /// Vérifie si l'utilisateur peut envoyer des messages
    pub fn can_send_messages(&self) -> bool {
        !self.is_banned()
    }

    /// Vérifie si l'utilisateur peut créer des salons
    pub fn can_create_rooms(&self) -> bool {
        matches!(self, Role::Admin | Role::Moderator | Role::User)
    }
}

/// Session utilisateur active
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct UserSession {
    /// ID unique de l'utilisateur
    pub user_id: i32,
    /// Nom d'utilisateur
    pub username: String,
    /// Rôle de l'utilisateur
    pub role: Role,
    /// Timestamp de connexion
    pub connected_at: DateTime<Utc>,
    /// Dernière activité
    pub last_activity: DateTime<Utc>,
    /// Adresse IP de connexion
    pub ip_address: String,
    /// User agent du client
    pub user_agent: Option<String>,
    /// Salons auxquels l'utilisateur est connecté
    pub active_rooms: Vec<String>,
    /// Statut de présence
    pub presence_status: PresenceStatus,
}

/// Statut de présence de l'utilisateur
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum PresenceStatus {
    Online,
    Away,
    Busy,
    Offline,
}

impl UserSession {
    /// Crée une nouvelle session utilisateur
    pub fn new(
        user_id: i32,
        username: String,
        role: Role,
        ip_address: String,
        user_agent: Option<String>,
    ) -> Self {
        let now = Utc::now();
        Self {
            user_id,
            username,
            role,
            connected_at: now,
            last_activity: now,
            ip_address,
            user_agent,
            active_rooms: Vec::new(),
            presence_status: PresenceStatus::Online,
        }
    }

    /// Met à jour la dernière activité
    pub fn update_activity(&mut self) {
        self.last_activity = Utc::now();
    }

    /// Ajoute l'utilisateur à un salon
    pub fn join_room(&mut self, room_id: String) {
        if !self.active_rooms.contains(&room_id) {
            self.active_rooms.push(room_id);
        }
        self.update_activity();
    }

    /// Retire l'utilisateur d'un salon
    pub fn leave_room(&mut self, room_id: &str) {
        self.active_rooms.retain(|r| r != room_id);
        self.update_activity();
    }

    /// Change le statut de présence
    pub fn set_presence(&mut self, status: PresenceStatus) {
        self.presence_status = status;
        self.update_activity();
    }

    /// Vérifie si l'utilisateur est dans un salon spécifique
    pub fn is_in_room(&self, room_id: &str) -> bool {
        self.active_rooms.contains(&room_id.to_string())
    }

    /// Vérifie si la session est expirée (inactivité > 1 heure)
    pub fn is_expired(&self) -> bool {
        let now = Utc::now();
        let duration = now.signed_duration_since(self.last_activity);
        duration.num_hours() > 1
    }
}

/// Gestionnaire d'authentification
pub struct AuthManager {
    /// Sessions actives (user_id -> session)
    sessions: HashMap<i32, UserSession>,
    /// Connexions WebSocket (connection_id -> user_id)
    connections: HashMap<String, i32>,
}

impl AuthManager {
    /// Crée un nouveau gestionnaire d'authentification
    pub fn new() -> Self {
        Self {
            sessions: HashMap::new(),
            connections: HashMap::new(),
        }
    }

    /// Authentifie un utilisateur et crée une session
    pub fn authenticate_user(
        &mut self,
        user_id: i32,
        username: String,
        role: Role,
        connection_id: String,
        ip_address: String,
        user_agent: Option<String>,
    ) -> Result<&UserSession> {
        // Créer ou mettre à jour la session
        let session = UserSession::new(user_id, username, role, ip_address, user_agent);
        
        // Stocker la session
        self.sessions.insert(user_id, session);
        self.connections.insert(connection_id, user_id);

        Ok(self.sessions.get(&user_id).unwrap())
    }

    /// Récupère une session par ID utilisateur
    pub fn get_session(&self, user_id: i32) -> Option<&UserSession> {
        self.sessions.get(&user_id)
    }

    /// Récupère une session par ID de connexion
    pub fn get_session_by_connection(&self, connection_id: &str) -> Option<&UserSession> {
        let user_id = self.connections.get(connection_id)?;
        self.sessions.get(user_id)
    }

    /// Met à jour l'activité d'une session
    pub fn update_activity(&mut self, user_id: i32) -> Result<()> {
        if let Some(session) = self.sessions.get_mut(&user_id) {
            session.update_activity();
            Ok(())
        } else {
            Err(ChatError::unauthorized("Session non trouvée"))
        }
    }

    /// Déconnecte un utilisateur
    pub fn disconnect_user(&mut self, connection_id: &str) -> Option<UserSession> {
        if let Some(user_id) = self.connections.remove(connection_id) {
            // Retirer de tous les salons
            if let Some(mut session) = self.sessions.remove(&user_id) {
                session.active_rooms.clear();
                session.presence_status = PresenceStatus::Offline;
                Some(session)
            } else {
                None
            }
        } else {
            None
        }
    }

    /// Nettoie les sessions expirées
    pub fn cleanup_expired_sessions(&mut self) -> Vec<i32> {
        let mut expired_users = Vec::new();
        
        self.sessions.retain(|&user_id, session| {
            if session.is_expired() {
                expired_users.push(user_id);
                false
            } else {
                true
            }
        });

        // Nettoyer aussi les connexions
        self.connections.retain(|_, &mut user_id| {
            !expired_users.contains(&user_id)
        });

        expired_users
    }

    /// Récupère toutes les sessions actives
    pub fn get_active_sessions(&self) -> Vec<&UserSession> {
        self.sessions.values().collect()
    }

    /// Récupère le nombre de sessions actives
    pub fn active_session_count(&self) -> usize {
        self.sessions.len()
    }
}

impl Default for AuthManager {
    fn default() -> Self {
        Self::new()
    }
} 