//! User Management et Présence
//! 
//! Gestion des utilisateurs connectés avec tracking de présence
//! et activités Discord-like.

use std::sync::Arc;
use dashmap::DashMap;
use serde::{Serialize, Deserialize};
use chrono::{DateTime, Utc};

/// Status de présence Discord-like
#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub enum PresenceStatus {
    Online,
    Idle,       // Inactif (>10 min)
    DoNotDisturb,
    Invisible,  // Apparaît offline
}

/// Activité utilisateur
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct UserActivity {
    pub activity_type: ActivityType,
    pub name: String,
    pub details: Option<String>,
    pub state: Option<String>,
    pub started_at: Option<DateTime<Utc>>,
}

/// Type d'activité
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum ActivityType {
    Playing,    // Joue à un jeu
    Streaming,  // Stream
    Listening,  // Écoute de la musique
    Watching,   // Regarde
    Custom,     // Status custom
    Competing,  // Compétition
}

/// Tracker de présence optimisé pour haute performance
#[derive(Debug)]
pub struct PresenceTracker {
    /// Status des utilisateurs
    statuses: Arc<DashMap<i64, PresenceStatus>>,
    
    /// Dernière activité
    last_seen: Arc<DashMap<i64, DateTime<Utc>>>,
    
    /// Activités en cours
    activities: Arc<DashMap<i64, UserActivity>>,
    
    /// Utilisateurs en train d'écrire par salle
    typing_users: Arc<DashMap<String, DashMap<i64, DateTime<Utc>>>>,
}

impl PresenceTracker {
    pub fn new() -> Self {
        Self {
            statuses: Arc::new(DashMap::new()),
            last_seen: Arc::new(DashMap::new()),
            activities: Arc::new(DashMap::new()),
            typing_users: Arc::new(DashMap::new()),
        }
    }

    /// Met à jour le status d'un utilisateur
    pub fn update_status(&self, user_id: i64, status: PresenceStatus) {
        self.statuses.insert(user_id, status);
        self.last_seen.insert(user_id, Utc::now());
    }

    /// Met à jour l'activité d'un utilisateur
    pub fn update_activity(&self, user_id: i64, activity: Option<UserActivity>) {
        match activity {
            Some(activity) => {
                self.activities.insert(user_id, activity);
            }
            None => {
                self.activities.remove(&user_id);
            }
        }
        self.last_seen.insert(user_id, Utc::now());
    }

    /// Obtient le status d'un utilisateur
    pub fn get_status(&self, user_id: i64) -> Option<PresenceStatus> {
        self.statuses.get(&user_id).map(|entry| entry.value().clone())
    }

    /// Obtient l'activité d'un utilisateur
    pub fn get_activity(&self, user_id: i64) -> Option<UserActivity> {
        self.activities.get(&user_id).map(|entry| entry.value().clone())
    }

    /// Vérifie si un utilisateur est en ligne
    pub fn is_online(&self, user_id: i64) -> bool {
        matches!(
            self.get_status(user_id), 
            Some(PresenceStatus::Online | PresenceStatus::Idle | PresenceStatus::DoNotDisturb)
        )
    }

    /// Démarre l'indicateur "en train d'écrire"
    pub fn start_typing(&self, user_id: i64, room_id: &str) {
        let room_key = room_id.to_string();
        let typing_room = self.typing_users.entry(room_key)
            .or_insert_with(|| DashMap::new());
        typing_room.insert(user_id, Utc::now());
    }

    /// Arrête l'indicateur "en train d'écrire"
    pub fn stop_typing(&self, user_id: i64, room_id: &str) {
        if let Some(typing_room) = self.typing_users.get(room_id) {
            typing_room.remove(&user_id);
        }
    }

    /// Obtient la liste des utilisateurs en train d'écrire
    pub fn get_typing_users(&self, room_id: &str) -> Vec<i64> {
        if let Some(typing_room) = self.typing_users.get(room_id) {
            let now = Utc::now();
            let timeout = std::time::Duration::from_secs(5); // 5 secondes timeout
            
            // Nettoyer les anciens indicateurs et retourner les actifs
            typing_room.retain(|_, last_typing| {
                now.signed_duration_since(*last_typing) < chrono::Duration::from_std(timeout).unwrap_or(chrono::Duration::seconds(5))
            });
            
            typing_room.iter().map(|entry| *entry.key()).collect()
        } else {
            Vec::new()
        }
    }

    /// Nettoie les utilisateurs inactifs
    pub fn cleanup_inactive_users(&self, inactive_threshold: std::time::Duration) -> usize {
        let now = Utc::now();
        let mut cleaned = 0;

        // Nettoyer les statuses des utilisateurs inactifs
        self.statuses.retain(|user_id, _| {
            if let Some(last_seen) = self.last_seen.get(user_id) {
                let is_active = now.signed_duration_since(*last_seen.value()) < chrono::Duration::from_std(inactive_threshold).unwrap_or(chrono::Duration::hours(1));
                if !is_active {
                    cleaned += 1;
                    // Nettoyer aussi l'activité
                    self.activities.remove(user_id);
                }
                is_active
            } else {
                false
            }
        });

        // Nettoyer les anciens indicateurs de frappe
        for typing_room in self.typing_users.iter() {
            typing_room.value().retain(|_, last_typing| {
                now.signed_duration_since(*last_typing) < chrono::Duration::seconds(5)
            });
        }

        // Supprimer les salles vides de typing
        self.typing_users.retain(|_, typing_room| {
            !typing_room.is_empty()
        });

        cleaned
    }

    /// Obtient les statistiques de présence
    pub fn get_presence_stats(&self) -> PresenceStats {
        let mut stats = PresenceStats::default();
        
        for entry in self.statuses.iter() {
            match entry.value() {
                PresenceStatus::Online => stats.online += 1,
                PresenceStatus::Idle => stats.idle += 1,
                PresenceStatus::DoNotDisturb => stats.dnd += 1,
                PresenceStatus::Invisible => stats.invisible += 1,
            }
        }
        
        stats.total = stats.online + stats.idle + stats.dnd + stats.invisible;
        stats
    }
}

/// Statistiques de présence
#[derive(Debug, Default, Serialize)]
pub struct PresenceStats {
    pub total: usize,
    pub online: usize,
    pub idle: usize,
    pub dnd: usize,
    pub invisible: usize,
}

impl Default for PresenceStatus {
    fn default() -> Self {
        Self::Online
    }
}

impl UserActivity {
    pub fn playing(name: String) -> Self {
        Self {
            activity_type: ActivityType::Playing,
            name,
            details: None,
            state: None,
            started_at: Some(Utc::now()),
        }
    }

    pub fn listening(name: String) -> Self {
        Self {
            activity_type: ActivityType::Listening,
            name,
            details: None,
            state: None,
            started_at: Some(Utc::now()),
        }
    }

    pub fn streaming(name: String, url: String) -> Self {
        Self {
            activity_type: ActivityType::Streaming,
            name,
            details: Some(url),
            state: None,
            started_at: Some(Utc::now()),
        }
    }

    pub fn custom(status: String) -> Self {
        Self {
            activity_type: ActivityType::Custom,
            name: status,
            details: None,
            state: None,
            started_at: Some(Utc::now()),
        }
    }
}
