use std::collections::{HashMap, VecDeque};
use std::sync::Arc;
use std::time::{Duration, SystemTime, UNIX_EPOCH};
use tokio::sync::{broadcast, RwLock};
use serde::{Deserialize, Serialize};
use tracing::{debug, error, info, warn};
use crate::config::Config;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Notification {
    pub id: String,
    pub user_id: String,
    pub notification_type: NotificationType,
    pub title: String,
    pub message: String,
    pub data: Option<serde_json::Value>,
    pub priority: NotificationPriority,
    pub channels: Vec<NotificationChannel>,
    pub created_at: u64,
    pub expires_at: Option<u64>,
    pub read: bool,
    pub delivered: bool,
    pub delivery_attempts: u32,
    pub tags: Vec<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq, Hash)]
pub enum NotificationType {
    // Notifications de streaming
    TrackStarted,
    TrackFinished,
    PlaylistUpdated,
    QualityChanged,
    
    // Notifications sociales
    NewFollower,
    NewLike,
    NewComment,
    NewShare,
    
    // Notifications système
    SystemMaintenance,
    ServiceDegraded,
    ServiceRestored,
    
    // Notifications de contenu
    NewTrackUploaded,
    TrackApproved,
    TrackRejected,
    
    // Notifications de sécurité
    LoginFromNewDevice,
    PasswordChanged,
    SuspiciousActivity,
    
    // Notifications promotionnelles
    SubscriptionExpiring,
    NewFeature,
    SpecialOffer,
    
    // Notifications personnalisées
    Custom(String),
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq, PartialOrd, Ord)]
pub enum NotificationPriority {
    Low,
    Normal,
    High,
    Critical,
    Emergency,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq, Hash)]
pub enum NotificationChannel {
    WebSocket,
    Email,
    SMS,
    Push,
    InApp,
    Webhook,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct NotificationTemplate {
    pub id: String,
    pub notification_type: NotificationType,
    pub title_template: String,
    pub message_template: String,
    pub default_channels: Vec<NotificationChannel>,
    pub default_priority: NotificationPriority,
    pub variables: Vec<String>,
    pub localization: HashMap<String, LocalizedTemplate>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct LocalizedTemplate {
    pub title: String,
    pub message: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct UserPreferences {
    pub user_id: String,
    pub enabled_channels: HashMap<NotificationChannel, bool>,
    pub type_preferences: HashMap<NotificationType, NotificationPreference>,
    pub quiet_hours: Option<QuietHours>,
    pub frequency_limits: HashMap<NotificationType, FrequencyLimit>,
    pub language: String,
    pub timezone: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct NotificationPreference {
    pub enabled: bool,
    pub channels: Vec<NotificationChannel>,
    pub priority_threshold: NotificationPriority,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct QuietHours {
    pub start_hour: u8,  // 0-23
    pub end_hour: u8,    // 0-23
    pub timezone: String,
    pub enabled_days: Vec<u8>, // 0=Sunday, 1=Monday, etc.
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FrequencyLimit {
    pub max_per_hour: u32,
    pub max_per_day: u32,
    pub cooldown_minutes: u32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct NotificationStats {
    pub total_sent: u64,
    pub total_delivered: u64,
    pub total_read: u64,
    pub delivery_rate: f32,
    pub read_rate: f32,
    pub channel_stats: HashMap<NotificationChannel, ChannelStats>,
    pub type_stats: HashMap<String, TypeStats>,
    pub recent_failures: Vec<DeliveryFailure>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ChannelStats {
    pub sent: u64,
    pub delivered: u64,
    pub failed: u64,
    pub average_delivery_time_ms: u64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TypeStats {
    pub sent: u64,
    pub read: u64,
    pub average_read_time_minutes: u64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DeliveryFailure {
    pub notification_id: String,
    pub channel: NotificationChannel,
    pub error: String,
    pub timestamp: u64,
    pub retry_count: u32,
}

pub struct NotificationService {
    config: Arc<Config>,
    templates: Arc<RwLock<HashMap<String, NotificationTemplate>>>,
    user_preferences: Arc<RwLock<HashMap<String, UserPreferences>>>,
    pending_notifications: Arc<RwLock<VecDeque<Notification>>>,
    notification_history: Arc<RwLock<HashMap<String, Notification>>>,
    stats: Arc<RwLock<NotificationStats>>,
    websocket_sender: broadcast::Sender<Notification>,
    delivery_workers: usize,
}

impl NotificationService {
    pub fn new(config: Arc<Config>) -> Self {
        let (websocket_sender, _) = broadcast::channel(1000);
        
        let delivery_workers = config.performance.worker_threads.unwrap_or(4);
        
        Self {
            config,
            templates: Arc::new(RwLock::new(HashMap::new())),
            user_preferences: Arc::new(RwLock::new(HashMap::new())),
            pending_notifications: Arc::new(RwLock::new(VecDeque::new())),
            notification_history: Arc::new(RwLock::new(HashMap::new())),
            stats: Arc::new(RwLock::new(NotificationStats {
                total_sent: 0,
                total_delivered: 0,
                total_read: 0,
                delivery_rate: 0.0,
                read_rate: 0.0,
                channel_stats: HashMap::new(),
                type_stats: HashMap::new(),
                recent_failures: Vec::new(),
            })),
            websocket_sender,
            delivery_workers,
        }
    }

    pub async fn start_delivery_workers(&self) {
        info!("📬 Démarrage de {} workers de notifications", self.delivery_workers);
        
        for worker_id in 0..self.delivery_workers {
            let service = self.clone();
            tokio::spawn(async move {
                service.delivery_worker_loop(worker_id).await;
            });
        }

        // Worker de nettoyage des anciennes notifications
        let service = self.clone();
        tokio::spawn(async move {
            service.cleanup_worker().await;
        });
    }

    async fn delivery_worker_loop(&self, worker_id: usize) {
        debug!("Worker de notifications {} démarré", worker_id);
        
        loop {
            let notification = {
                let mut pending = self.pending_notifications.write().await;
                pending.pop_front()
            };

            if let Some(notification) = notification {
                debug!("Worker {} traite la notification {}", worker_id, notification.id);
                self.deliver_notification(notification).await;
            } else {
                // Pas de notification, attendre un peu
                tokio::time::sleep(Duration::from_millis(100)).await;
            }
        }
    }

    async fn cleanup_worker(&self) {
        let mut interval = tokio::time::interval(Duration::from_secs(3600)); // 1 heure
        
        loop {
            interval.tick().await;
            self.cleanup_expired_notifications().await;
            self.cleanup_old_history().await;
            self.update_statistics().await;
        }
    }

    pub async fn send_notification(&self, mut notification: Notification) -> Result<String, NotificationError> {
        // Valider la notification
        self.validate_notification(&notification)?;
        
        // Appliquer les préférences utilisateur
        notification = self.apply_user_preferences(notification).await?;
        
        // Vérifier les limites de fréquence
        if !self.check_frequency_limits(&notification).await {
            return Err(NotificationError::FrequencyLimitExceeded);
        }

        // Générer un ID si nécessaire
        if notification.id.is_empty() {
            notification.id = uuid::Uuid::new_v4().to_string();
        }

        // Ajouter à l'historique
        {
            let mut history = self.notification_history.write().await;
            history.insert(notification.id.clone(), notification.clone());
        }

        // Ajouter à la queue de livraison
        {
            let mut pending = self.pending_notifications.write().await;
            pending.push_back(notification.clone());
        }

        // Mettre à jour les stats
        {
            let mut stats = self.stats.write().await;
            stats.total_sent += 1;
        }

        info!("📨 Notification {} ajoutée à la queue pour l'utilisateur {}", 
              notification.id, notification.user_id);

        Ok(notification.id)
    }

    pub async fn send_from_template(
        &self,
        template_id: &str,
        user_id: &str,
        variables: HashMap<String, String>,
        override_channels: Option<Vec<NotificationChannel>>,
        override_priority: Option<NotificationPriority>,
    ) -> Result<String, NotificationError> {
        let template = {
            let templates = self.templates.read().await;
            templates.get(template_id).cloned()
                .ok_or_else(|| NotificationError::TemplateNotFound(template_id.to_string()))?
        };

        // Récupérer les préférences utilisateur pour la localisation
        let user_prefs = self.get_user_preferences(user_id).await;
        let language = user_prefs.as_ref().map(|p| p.language.as_str()).unwrap_or("en");

        // Utiliser le template localisé si disponible
        let (title_template, message_template) = if let Some(localized) = template.localization.get(language) {
            (&localized.title, &localized.message)
        } else {
            (&template.title_template, &template.message_template)
        };

        // Remplacer les variables dans le template
        let title = self.replace_template_variables(title_template, &variables);
        let message = self.replace_template_variables(message_template, &variables);

        let notification = Notification {
            id: String::new(), // Sera généré automatiquement
            user_id: user_id.to_string(),
            notification_type: template.notification_type.clone(),
            title,
            message,
            data: Some(serde_json::to_value(variables).unwrap_or_default()),
            priority: override_priority.unwrap_or(template.default_priority.clone()),
            channels: override_channels.unwrap_or(template.default_channels.clone()),
            created_at: SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_secs(),
            expires_at: None,
            read: false,
            delivered: false,
            delivery_attempts: 0,
            tags: Vec::new(),
        };

        self.send_notification(notification).await
    }

    async fn deliver_notification(&self, mut notification: Notification) {
        let start_time = SystemTime::now();
        
        for channel in &notification.channels.clone() {
            let delivery_result = match channel {
                NotificationChannel::WebSocket => self.deliver_websocket(&notification).await,
                NotificationChannel::Email => self.deliver_email(&notification).await,
                NotificationChannel::SMS => self.deliver_sms(&notification).await,
                NotificationChannel::Push => self.deliver_push(&notification).await,
                NotificationChannel::InApp => self.deliver_in_app(&notification).await,
                NotificationChannel::Webhook => self.deliver_webhook(&notification).await,
            };

            match delivery_result {
                Ok(_) => {
                    debug!("✅ Notification {} livrée via {:?}", notification.id, channel);
                    self.update_channel_stats(channel, true, start_time).await;
                }
                Err(e) => {
                    error!("❌ Échec de livraison de la notification {} via {:?}: {:?}", 
                           notification.id, channel, e);
                    self.update_channel_stats(channel, false, start_time).await;
                    self.record_delivery_failure(&notification, channel, &e).await;
                }
            }
        }

        notification.delivered = true;
        notification.delivery_attempts += 1;

        // Mettre à jour dans l'historique
        {
            let mut history = self.notification_history.write().await;
            history.insert(notification.id.clone(), notification);
        }

        // Mettre à jour les stats globales
        {
            let mut stats = self.stats.write().await;
            stats.total_delivered += 1;
            stats.delivery_rate = stats.total_delivered as f32 / stats.total_sent as f32;
        }
    }

    async fn deliver_websocket(&self, notification: &Notification) -> Result<(), NotificationError> {
        // Envoyer via le canal WebSocket
        self.websocket_sender.send(notification.clone())
            .map_err(|e| NotificationError::DeliveryFailed(format!("WebSocket: {}", e)))?;
        Ok(())
    }

    async fn deliver_email(&self, notification: &Notification) -> Result<(), NotificationError> {
        // Simuler l'envoi d'email (à implémenter avec votre provider SMTP)
        debug!("📧 Envoi d'email pour la notification {}", notification.id);
        tokio::time::sleep(Duration::from_millis(100)).await; // Simuler la latence
        Ok(())
    }

    async fn deliver_sms(&self, notification: &Notification) -> Result<(), NotificationError> {
        // Simuler l'envoi de SMS (à implémenter avec votre provider SMS)
        debug!("📱 Envoi de SMS pour la notification {}", notification.id);
        tokio::time::sleep(Duration::from_millis(200)).await;
        Ok(())
    }

    async fn deliver_push(&self, notification: &Notification) -> Result<(), NotificationError> {
        // Simuler l'envoi de push notification (à implémenter avec FCM/APNs)
        debug!("🔔 Envoi de push notification pour {}", notification.id);
        tokio::time::sleep(Duration::from_millis(150)).await;
        Ok(())
    }

    async fn deliver_in_app(&self, notification: &Notification) -> Result<(), NotificationError> {
        // Les notifications in-app sont stockées et récupérées via API
        debug!("📋 Notification in-app stockée pour {}", notification.id);
        Ok(())
    }

    async fn deliver_webhook(&self, notification: &Notification) -> Result<(), NotificationError> {
        // Simuler l'envoi vers un webhook (à implémenter avec reqwest)
        debug!("🔗 Envoi webhook pour la notification {}", notification.id);
        tokio::time::sleep(Duration::from_millis(300)).await;
        Ok(())
    }

    pub async fn mark_as_read(&self, notification_id: &str, user_id: &str) -> Result<(), NotificationError> {
        let mut history = self.notification_history.write().await;
        
        if let Some(notification) = history.get_mut(notification_id) {
            if notification.user_id != user_id {
                return Err(NotificationError::Unauthorized);
            }
            
            if !notification.read {
                notification.read = true;
                
                // Mettre à jour les stats
                let mut stats = self.stats.write().await;
                stats.total_read += 1;
                stats.read_rate = stats.total_read as f32 / stats.total_delivered as f32;
            }
            
            Ok(())
        } else {
            Err(NotificationError::NotificationNotFound(notification_id.to_string()))
        }
    }

    pub async fn get_user_notifications(
        &self,
        user_id: &str,
        limit: Option<usize>,
        offset: Option<usize>,
        unread_only: bool,
    ) -> Vec<Notification> {
        let history = self.notification_history.read().await;
        let limit = limit.unwrap_or(50);
        let offset = offset.unwrap_or(0);
        
        let mut user_notifications: Vec<_> = history.values()
            .filter(|n| n.user_id == user_id)
            .filter(|n| !unread_only || !n.read)
            .cloned()
            .collect();
        
        // Trier par date de création (plus récent en premier)
        user_notifications.sort_by(|a, b| b.created_at.cmp(&a.created_at));
        
        user_notifications.into_iter()
            .skip(offset)
            .take(limit)
            .collect()
    }

    pub async fn get_websocket_receiver(&self) -> broadcast::Receiver<Notification> {
        self.websocket_sender.subscribe()
    }

    pub async fn register_template(&self, template: NotificationTemplate) {
        let mut templates = self.templates.write().await;
        templates.insert(template.id.clone(), template);
    }

    pub async fn update_user_preferences(&self, preferences: UserPreferences) {
        let mut user_prefs = self.user_preferences.write().await;
        user_prefs.insert(preferences.user_id.clone(), preferences);
    }

    pub async fn get_user_preferences(&self, user_id: &str) -> Option<UserPreferences> {
        let user_prefs = self.user_preferences.read().await;
        user_prefs.get(user_id).cloned()
    }

    pub async fn get_statistics(&self) -> NotificationStats {
        self.stats.read().await.clone()
    }

    // Méthodes utilitaires privées

    fn validate_notification(&self, notification: &Notification) -> Result<(), NotificationError> {
        if notification.user_id.is_empty() {
            return Err(NotificationError::InvalidNotification("user_id is required".to_string()));
        }
        
        if notification.title.is_empty() {
            return Err(NotificationError::InvalidNotification("title is required".to_string()));
        }
        
        if notification.channels.is_empty() {
            return Err(NotificationError::InvalidNotification("at least one channel is required".to_string()));
        }

        Ok(())
    }

    async fn apply_user_preferences(&self, mut notification: Notification) -> Result<Notification, NotificationError> {
        if let Some(prefs) = self.get_user_preferences(&notification.user_id).await {
            // Filtrer les canaux selon les préférences
            notification.channels.retain(|channel| {
                *prefs.enabled_channels.get(channel).unwrap_or(&true)
            });

            // Vérifier les préférences par type
            if let Some(type_pref) = prefs.type_preferences.get(&notification.notification_type) {
                if !type_pref.enabled {
                    return Err(NotificationError::NotificationDisabled);
                }
                
                if notification.priority < type_pref.priority_threshold {
                    return Err(NotificationError::PriorityTooLow);
                }

                // Utiliser les canaux préférés si spécifiés
                if !type_pref.channels.is_empty() {
                    notification.channels = type_pref.channels.clone();
                }
            }

            // Vérifier les heures de silence
            if let Some(quiet_hours) = &prefs.quiet_hours {
                if self.is_in_quiet_hours(quiet_hours).await && notification.priority < NotificationPriority::Critical {
                    return Err(NotificationError::QuietHours);
                }
            }
        }

        if notification.channels.is_empty() {
            return Err(NotificationError::NoEnabledChannels);
        }

        Ok(notification)
    }

    async fn check_frequency_limits(&self, notification: &Notification) -> bool {
        // Simuler la vérification des limites de fréquence
        // Dans une implémentation réelle, on vérifierait la base de données
        true
    }

    async fn is_in_quiet_hours(&self, _quiet_hours: &QuietHours) -> bool {
        // Simuler la vérification des heures de silence
        // Dans une implémentation réelle, on vérifierait l'heure actuelle selon le fuseau horaire
        false
    }

    fn replace_template_variables(&self, template: &str, variables: &HashMap<String, String>) -> String {
        let mut result = template.to_string();
        for (key, value) in variables {
            result = result.replace(&format!("{{{}}}", key), value);
        }
        result
    }

    async fn update_channel_stats(&self, channel: &NotificationChannel, success: bool, start_time: SystemTime) {
        let mut stats = self.stats.write().await;
        let channel_stats = stats.channel_stats.entry(channel.clone()).or_insert(ChannelStats {
            sent: 0,
            delivered: 0,
            failed: 0,
            average_delivery_time_ms: 0,
        });

        channel_stats.sent += 1;
        if success {
            channel_stats.delivered += 1;
        } else {
            channel_stats.failed += 1;
        }

        let delivery_time = start_time.elapsed().unwrap_or_default().as_millis() as u64;
        channel_stats.average_delivery_time_ms = 
            (channel_stats.average_delivery_time_ms + delivery_time) / 2;
    }

    async fn record_delivery_failure(&self, notification: &Notification, channel: &NotificationChannel, error: &NotificationError) {
        let mut stats = self.stats.write().await;
        
        let failure = DeliveryFailure {
            notification_id: notification.id.clone(),
            channel: channel.clone(),
            error: error.to_string(),
            timestamp: SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_secs(),
            retry_count: notification.delivery_attempts,
        };

        stats.recent_failures.push(failure);
        
        // Garder seulement les 100 derniers échecs
        if stats.recent_failures.len() > 100 {
            stats.recent_failures.remove(0);
        }
    }

    async fn cleanup_expired_notifications(&self) {
        let now = SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_secs();
        let mut history = self.notification_history.write().await;
        
        history.retain(|_, notification| {
            notification.expires_at.map_or(true, |expires| expires > now)
        });
    }

    async fn cleanup_old_history(&self) {
        let cutoff = SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_secs() - (30 * 24 * 3600); // 30 jours
        let mut history = self.notification_history.write().await;
        
        history.retain(|_, notification| notification.created_at > cutoff);
    }

    async fn update_statistics(&self) {
        // Mettre à jour les statistiques périodiquement
        let mut stats = self.stats.write().await;
        
        if stats.total_sent > 0 {
            stats.delivery_rate = stats.total_delivered as f32 / stats.total_sent as f32;
        }
        
        if stats.total_delivered > 0 {
            stats.read_rate = stats.total_read as f32 / stats.total_delivered as f32;
        }
    }
}

impl Clone for NotificationService {
    fn clone(&self) -> Self {
        Self {
            config: self.config.clone(),
            templates: self.templates.clone(),
            user_preferences: self.user_preferences.clone(),
            pending_notifications: self.pending_notifications.clone(),
            notification_history: self.notification_history.clone(),
            stats: self.stats.clone(),
            websocket_sender: self.websocket_sender.clone(),
            delivery_workers: self.delivery_workers,
        }
    }
}

#[derive(Debug, thiserror::Error)]
pub enum NotificationError {
    #[error("Notification invalide: {0}")]
    InvalidNotification(String),
    
    #[error("Template non trouvé: {0}")]
    TemplateNotFound(String),
    
    #[error("Notification non trouvée: {0}")]
    NotificationNotFound(String),
    
    #[error("Limite de fréquence dépassée")]
    FrequencyLimitExceeded,
    
    #[error("Type de notification désactivé")]
    NotificationDisabled,
    
    #[error("Priorité trop faible")]
    PriorityTooLow,
    
    #[error("Heures de silence actives")]
    QuietHours,
    
    #[error("Aucun canal activé")]
    NoEnabledChannels,
    
    #[error("Non autorisé")]
    Unauthorized,
    
    #[error("Échec de livraison: {0}")]
    DeliveryFailed(String),
} 