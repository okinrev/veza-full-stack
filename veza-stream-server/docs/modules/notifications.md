# Notifications Module Documentation

Le module notifications fournit un système complet de notifications multi-canal avec gestion des templates, préférences utilisateur et statistiques de livraison.

## Table des Matières

- [Vue d'ensemble](#vue-densemble)
- [Architecture](#architecture)
- [Types de notifications](#types-de-notifications)
- [Canaux de diffusion](#canaux-de-diffusion)
- [Système de templates](#système-de-templates)
- [Préférences utilisateur](#préférences-utilisateur)
- [Types et Structures](#types-et-structures)
- [NotificationService](#notificationservice)
- [API Reference](#api-reference)
- [Exemples d'utilisation](#exemples-dutilisation)
- [Intégration](#intégration)

## Vue d'ensemble

Le système de notifications permet :
- **Notifications multi-canal** : WebSocket, Email, SMS, Push, In-App, Webhook
- **Système de templates** : Templates personnalisables avec variables
- **Préférences utilisateur** : Contrôle granulaire par type et canal
- **Heures de silence** : Respect des plages horaires
- **Limites de fréquence** : Prévention du spam
- **Livraison asynchrone** : Workers dédiés avec retry
- **Statistiques complètes** : Taux de livraison et de lecture

## Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Trigger       │    │  Notification    │    │   Delivery      │
│                 │    │   Service        │    │   Workers       │
│ - User Action   │───►│                  │───►│                 │
│ - System Event  │    │ - Template       │    │ - WebSocket     │
│ - Scheduled     │    │   Processing     │    │ - Email         │
│ - External API  │    │ - User Prefs     │    │ - SMS           │
│                 │    │ - Queue Mgmt     │    │ - Push          │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                │
                    ┌──────────────────────┐
                    │   Storage Layer      │
                    │                      │
                    │ - Templates DB       │
                    │ - User Preferences   │
                    │ - Notification Log   │
                    │ - Statistics         │
                    └──────────────────────┘
```

## Types de notifications

### NotificationType

```rust
pub enum NotificationType {
    // Notifications de streaming
    TrackStarted,           // Lecture commencée
    TrackFinished,          // Lecture terminée
    PlaylistUpdated,        // Playlist modifiée
    QualityChanged,         // Qualité ajustée
    
    // Notifications sociales
    NewFollower,            // Nouveau follower
    NewLike,                // Nouveau like
    NewComment,             // Nouveau commentaire
    NewShare,               // Nouveau partage
    
    // Notifications système
    SystemMaintenance,      // Maintenance programmée
    ServiceDegraded,        // Service dégradé
    ServiceRestored,        // Service restauré
    
    // Notifications de contenu
    NewTrackUploaded,       // Nouveau contenu disponible
    TrackApproved,          // Contenu approuvé
    TrackRejected,          // Contenu rejeté
    
    // Notifications de sécurité
    LoginFromNewDevice,     // Connexion depuis nouvel appareil
    PasswordChanged,        // Mot de passe modifié
    SuspiciousActivity,     // Activité suspecte
    
    // Notifications promotionnelles
    SubscriptionExpiring,   // Abonnement expirant
    NewFeature,             // Nouvelle fonctionnalité
    SpecialOffer,           // Offre spéciale
    
    // Notifications personnalisées
    Custom(String),         // Type personnalisé
}
```

### NotificationPriority

```rust
pub enum NotificationPriority {
    Low,        // Faible priorité (différable)
    Normal,     // Priorité normale
    High,       // Haute priorité (immédiate)
    Critical,   // Critique (bypass des préférences)
    Emergency,  // Urgence (tous canaux forcés)
}
```

## Canaux de diffusion

### NotificationChannel

```rust
pub enum NotificationChannel {
    WebSocket,  // Temps réel via WebSocket
    Email,      // Email SMTP/Service
    SMS,        // SMS via service tiers
    Push,       // Push notifications mobiles
    InApp,      // Notifications in-app
    Webhook,    // Webhook HTTP
}
```

### Capacités par canal

| Canal | Temps réel | Rich Media | Guaranteed | Offline |
|-------|------------|------------|------------|---------|
| WebSocket | ✅ | ✅ | ❌ | ❌ |
| Email | ❌ | ✅ | ✅ | ✅ |
| SMS | ❌ | ❌ | ✅ | ✅ |
| Push | ⚡ | ⚠️ | ⚠️ | ✅ |
| InApp | ✅ | ✅ | ❌ | ❌ |
| Webhook | ✅ | ✅ | ⚠️ | ❌ |

## Système de templates

### NotificationTemplate

```rust
pub struct NotificationTemplate {
    pub id: String,                                           // ID unique du template
    pub notification_type: NotificationType,                 // Type de notification
    pub title_template: String,                              // Template du titre
    pub message_template: String,                            // Template du message
    pub default_channels: Vec<NotificationChannel>,          // Canaux par défaut
    pub default_priority: NotificationPriority,             // Priorité par défaut
    pub variables: Vec<String>,                              // Variables requises
    pub localization: HashMap<String, LocalizedTemplate>,   // Traductions
}
```

### LocalizedTemplate

```rust
pub struct LocalizedTemplate {
    pub title: String,      // Titre localisé
    pub message: String,    // Message localisé
}
```

### Variables de template

Les templates supportent des variables avec la syntaxe `{{variable}}` :

```rust
// Template exemple
let template = NotificationTemplate {
    id: "new_follower".to_string(),
    notification_type: NotificationType::NewFollower,
    title_template: "Nouveau follower!".to_string(),
    message_template: "{{follower_name}} a commencé à vous suivre. Vous avez maintenant {{total_followers}} followers!".to_string(),
    variables: vec!["follower_name".to_string(), "total_followers".to_string()],
    // ...
};
```

## Préférences utilisateur

### UserPreferences

```rust
pub struct UserPreferences {
    pub user_id: String,                                              // ID utilisateur
    pub enabled_channels: HashMap<NotificationChannel, bool>,        // Canaux activés
    pub type_preferences: HashMap<NotificationType, NotificationPreference>, // Préfs par type
    pub quiet_hours: Option<QuietHours>,                            // Heures de silence
    pub frequency_limits: HashMap<NotificationType, FrequencyLimit>, // Limites de fréquence
    pub language: String,                                            // Langue préférée
    pub timezone: String,                                            // Fuseau horaire
}
```

### NotificationPreference

```rust
pub struct NotificationPreference {
    pub enabled: bool,                               // Activé pour ce type
    pub channels: Vec<NotificationChannel>,          // Canaux autorisés
    pub priority_threshold: NotificationPriority,   // Seuil de priorité minimum
}
```

### QuietHours

```rust
pub struct QuietHours {
    pub start_hour: u8,         // Heure de début (0-23)
    pub end_hour: u8,           // Heure de fin (0-23)
    pub timezone: String,       // Fuseau horaire
    pub enabled_days: Vec<u8>,  // Jours actifs (0=Dimanche)
}
```

### FrequencyLimit

```rust
pub struct FrequencyLimit {
    pub max_per_hour: u32,      // Maximum par heure
    pub max_per_day: u32,       // Maximum par jour
    pub cooldown_minutes: u32,  // Délai minimum entre envois
}
```

## Types et Structures

### Notification (Structure principale)

```rust
pub struct Notification {
    pub id: String,                         // ID unique
    pub user_id: String,                    // ID du destinataire
    pub notification_type: NotificationType, // Type de notification
    pub title: String,                      // Titre
    pub message: String,                    // Message
    pub data: Option<serde_json::Value>,    // Données supplémentaires
    pub priority: NotificationPriority,    // Priorité
    pub channels: Vec<NotificationChannel>, // Canaux de diffusion
    pub created_at: u64,                   // Timestamp de création
    pub expires_at: Option<u64>,           // Expiration (optionnel)
    pub read: bool,                        // Lu par l'utilisateur
    pub delivered: bool,                   // Livré avec succès
    pub delivery_attempts: u32,            // Tentatives de livraison
    pub tags: Vec<String>,                 // Tags pour filtrage
}
```

### NotificationStats

```rust
pub struct NotificationStats {
    pub total_sent: u64,                                        // Total envoyé
    pub total_delivered: u64,                                   // Total livré
    pub total_read: u64,                                        // Total lu
    pub delivery_rate: f32,                                     // Taux de livraison
    pub read_rate: f32,                                         // Taux de lecture
    pub channel_stats: HashMap<NotificationChannel, ChannelStats>, // Stats par canal
    pub type_stats: HashMap<String, TypeStats>,                // Stats par type
    pub recent_failures: Vec<DeliveryFailure>,                 // Échecs récents
}
```

### ChannelStats

```rust
pub struct ChannelStats {
    pub sent: u64,                          // Notifications envoyées
    pub delivered: u64,                     // Notifications livrées
    pub failed: u64,                        // Échecs de livraison
    pub average_delivery_time_ms: u64,      // Temps moyen de livraison
}
```

### TypeStats

```rust
pub struct TypeStats {
    pub sent: u64,                          // Notifications envoyées
    pub read: u64,                          // Notifications lues
    pub average_read_time_minutes: u64,     // Temps moyen avant lecture
}
```

### DeliveryFailure

```rust
pub struct DeliveryFailure {
    pub notification_id: String,    // ID de la notification
    pub channel: NotificationChannel, // Canal qui a échoué
    pub error: String,               // Message d'erreur
    pub timestamp: u64,              // Timestamp de l'échec
    pub retry_count: u32,            // Nombre de tentatives
}
```

## NotificationService

### Structure principale

```rust
pub struct NotificationService {
    config: Arc<Config>,                                          // Configuration
    templates: Arc<RwLock<HashMap<String, NotificationTemplate>>>, // Templates
    user_preferences: Arc<RwLock<HashMap<String, UserPreferences>>>, // Préférences
    pending_notifications: Arc<RwLock<VecDeque<Notification>>>,  // Queue des notifications
    notification_history: Arc<RwLock<HashMap<String, Notification>>>, // Historique
    stats: Arc<RwLock<NotificationStats>>,                      // Statistiques
    websocket_sender: broadcast::Sender<Notification>,          // Canal WebSocket
    delivery_workers: usize,                                    // Nombre de workers
}
```

### Cycle de vie d'une notification

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Creation      │───►│   Processing     │───►│   Delivery      │
│                 │    │                  │    │                 │
│ - Template      │    │ - User Prefs     │    │ - Channel       │
│ - Variables     │    │ - Frequency      │    │   Selection     │
│ - Priority      │    │ - Quiet Hours    │    │ - Retry Logic   │
│ - Target User   │    │ - Localization   │    │ - Statistics    │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

## API Reference

### NotificationService Methods

#### `new(config: Arc<Config>) -> Self`
Crée une nouvelle instance du service de notifications.

#### `start_delivery_workers(&self)`
Démarre les workers de livraison en arrière-plan.

#### `send_notification(notification: Notification) -> Result<String, NotificationError>`
Envoie une notification directement.

#### `send_from_template(template_id, user_id, variables, override_channels, override_priority) -> Result<String, NotificationError>`
Envoie une notification basée sur un template.

**Paramètres :**
- `template_id` : ID du template à utiliser
- `user_id` : ID du destinataire
- `variables` : Variables pour le template
- `override_channels` : Canaux spécifiques (optionnel)
- `override_priority` : Priorité spécifique (optionnel)

#### `mark_as_read(notification_id: &str, user_id: &str) -> Result<(), NotificationError>`
Marque une notification comme lue.

#### `get_user_notifications(user_id, limit, offset, unread_only) -> Vec<Notification>`
Récupère les notifications d'un utilisateur.

#### `register_template(template: NotificationTemplate)`
Enregistre un nouveau template.

#### `update_user_preferences(preferences: UserPreferences)`
Met à jour les préférences d'un utilisateur.

#### `get_user_preferences(user_id: &str) -> Option<UserPreferences>`
Récupère les préférences d'un utilisateur.

#### `get_statistics() -> NotificationStats`
Récupère les statistiques de notifications.

### Templates prédéfinis

Le service inclut des templates par défaut :

```rust
// Template pour nouveau follower
let new_follower_template = NotificationTemplate {
    id: "new_follower".to_string(),
    notification_type: NotificationType::NewFollower,
    title_template: "Nouveau follower!".to_string(),
    message_template: "{{follower_name}} a commencé à vous suivre!".to_string(),
    default_channels: vec![NotificationChannel::InApp, NotificationChannel::Push],
    default_priority: NotificationPriority::Normal,
    variables: vec!["follower_name".to_string()],
    localization: HashMap::from([
        ("en".to_string(), LocalizedTemplate {
            title: "New follower!".to_string(),
            message: "{{follower_name}} started following you!".to_string(),
        }),
    ]),
};

// Template pour maintenance système
let maintenance_template = NotificationTemplate {
    id: "system_maintenance".to_string(),
    notification_type: NotificationType::SystemMaintenance,
    title_template: "Maintenance programmée".to_string(),
    message_template: "Une maintenance est prévue le {{date}} de {{start_time}} à {{end_time}}. Le service pourra être temporairement indisponible.".to_string(),
    default_channels: vec![
        NotificationChannel::Email,
        NotificationChannel::InApp,
        NotificationChannel::Push,
    ],
    default_priority: NotificationPriority::High,
    variables: vec!["date".to_string(), "start_time".to_string(), "end_time".to_string()],
    localization: HashMap::new(),
};
```

## Exemples d'utilisation

### Envoi de notification simple

```rust
use stream_server::notifications::{NotificationService, Notification, NotificationType, NotificationPriority, NotificationChannel};

async fn example_simple_notification() -> Result<(), Box<dyn std::error::Error>> {
    let config = Arc::new(Config::from_env()?);
    let notification_service = NotificationService::new(config);
    
    // Démarrer les workers de livraison
    notification_service.start_delivery_workers().await;
    
    // Créer une notification simple
    let notification = Notification {
        id: Uuid::new_v4().to_string(),
        user_id: "user_123".to_string(),
        notification_type: NotificationType::TrackStarted,
        title: "Lecture en cours".to_string(),
        message: "Vous écoutez maintenant 'Bohemian Rhapsody' de Queen".to_string(),
        data: Some(serde_json::json!({
            "track_id": "queen_bohemian_rhapsody",
            "artist": "Queen",
            "album": "A Night at the Opera",
            "duration_ms": 355000
        })),
        priority: NotificationPriority::Normal,
        channels: vec![NotificationChannel::InApp, NotificationChannel::WebSocket],
        created_at: SystemTime::now().duration_since(UNIX_EPOCH)?.as_secs(),
        expires_at: None,
        read: false,
        delivered: false,
        delivery_attempts: 0,
        tags: vec!["music".to_string(), "playback".to_string()],
    };
    
    // Envoyer la notification
    let notification_id = notification_service.send_notification(notification).await?;
    println!("📨 Notification envoyée: {}", notification_id);
    
    Ok(())
}
```

### Utilisation des templates

```rust
async fn example_template_notification() -> Result<(), Box<dyn std::error::Error>> {
    let config = Arc::new(Config::from_env()?);
    let notification_service = NotificationService::new(config);
    
    // Démarrer les workers
    notification_service.start_delivery_workers().await;
    
    // Enregistrer un template personnalisé
    let like_template = NotificationTemplate {
        id: "track_liked".to_string(),
        notification_type: NotificationType::NewLike,
        title_template: "❤️ Votre piste a été likée!".to_string(),
        message_template: "{{liker_name}} a aimé votre piste '{{track_title}}'. Vous avez maintenant {{total_likes}} likes!".to_string(),
        default_channels: vec![NotificationChannel::InApp, NotificationChannel::Push],
        default_priority: NotificationPriority::Normal,
        variables: vec![
            "liker_name".to_string(),
            "track_title".to_string(),
            "total_likes".to_string(),
        ],
        localization: HashMap::from([
            ("en".to_string(), LocalizedTemplate {
                title: "❤️ Your track was liked!".to_string(),
                message: "{{liker_name}} liked your track '{{track_title}}'. You now have {{total_likes}} likes!".to_string(),
            }),
        ]),
    };
    
    notification_service.register_template(like_template).await;
    
    // Utiliser le template
    let variables = HashMap::from([
        ("liker_name".to_string(), "Alice Musicienne".to_string()),
        ("track_title".to_string(), "Summer Vibes".to_string()),
        ("total_likes".to_string(), "42".to_string()),
    ]);
    
    let notification_id = notification_service.send_from_template(
        "track_liked",
        "artist_789",
        variables,
        None, // Utiliser les canaux par défaut
        None, // Utiliser la priorité par défaut
    ).await?;
    
    println!("📨 Notification template envoyée: {}", notification_id);
    
    Ok(())
}
```

### Gestion des préférences utilisateur

```rust
async fn example_user_preferences() -> Result<(), Box<dyn std::error::Error>> {
    let config = Arc::new(Config::from_env()?);
    let notification_service = NotificationService::new(config);
    
    // Configurer les préférences d'un utilisateur
    let user_preferences = UserPreferences {
        user_id: "user_456".to_string(),
        
        // Canaux activés
        enabled_channels: HashMap::from([
            (NotificationChannel::InApp, true),
            (NotificationChannel::Push, true),
            (NotificationChannel::Email, false), // Désactivé
            (NotificationChannel::SMS, false),   // Désactivé
            (NotificationChannel::WebSocket, true),
        ]),
        
        // Préférences par type de notification
        type_preferences: HashMap::from([
            (NotificationType::NewFollower, NotificationPreference {
                enabled: true,
                channels: vec![NotificationChannel::InApp, NotificationChannel::Push],
                priority_threshold: NotificationPriority::Normal,
            }),
            (NotificationType::SystemMaintenance, NotificationPreference {
                enabled: true,
                channels: vec![NotificationChannel::InApp, NotificationChannel::Email],
                priority_threshold: NotificationPriority::High,
            }),
            (NotificationType::SpecialOffer, NotificationPreference {
                enabled: false, // Pas de notifications promotionnelles
                channels: vec![],
                priority_threshold: NotificationPriority::Critical,
            }),
        ]),
        
        // Heures de silence (22h à 8h)
        quiet_hours: Some(QuietHours {
            start_hour: 22,
            end_hour: 8,
            timezone: "Europe/Paris".to_string(),
            enabled_days: vec![1, 2, 3, 4, 5], // Lundi à vendredi
        }),
        
        // Limites de fréquence
        frequency_limits: HashMap::from([
            (NotificationType::NewLike, FrequencyLimit {
                max_per_hour: 5,
                max_per_day: 20,
                cooldown_minutes: 10,
            }),
            (NotificationType::NewComment, FrequencyLimit {
                max_per_hour: 3,
                max_per_day: 15,
                cooldown_minutes: 15,
            }),
        ]),
        
        language: "fr".to_string(),
        timezone: "Europe/Paris".to_string(),
    };
    
    // Enregistrer les préférences
    notification_service.update_user_preferences(user_preferences).await;
    
    // Test avec les préférences appliquées
    let notification = Notification {
        id: Uuid::new_v4().to_string(),
        user_id: "user_456".to_string(),
        notification_type: NotificationType::NewFollower,
        title: "Nouveau follower!".to_string(),
        message: "DJ_Producer a commencé à vous suivre!".to_string(),
        data: None,
        priority: NotificationPriority::Normal,
        channels: vec![NotificationChannel::Email], // Sera filtré selon les préférences
        created_at: SystemTime::now().duration_since(UNIX_EPOCH)?.as_secs(),
        expires_at: None,
        read: false,
        delivered: false,
        delivery_attempts: 0,
        tags: vec!["social".to_string()],
    };
    
    let notification_id = notification_service.send_notification(notification).await?;
    println!("📨 Notification avec préférences: {}", notification_id);
    
    Ok(())
}
```

### Notifications système avec priorité critique

```rust
async fn example_system_notifications() -> Result<(), Box<dyn std::error::Error>> {
    let config = Arc::new(Config::from_env()?);
    let notification_service = NotificationService::new(config);
    
    // Démarrer les workers
    notification_service.start_delivery_workers().await;
    
    // Notification de maintenance (haute priorité)
    let maintenance_vars = HashMap::from([
        ("date".to_string(), "15 janvier 2024".to_string()),
        ("start_time".to_string(), "02:00".to_string()),
        ("end_time".to_string(), "04:00".to_string()),
        ("services".to_string(), "Streaming audio".to_string()),
    ]);
    
    let maintenance_id = notification_service.send_from_template(
        "system_maintenance",
        "all_users", // Notification broadcast
        maintenance_vars,
        Some(vec![
            NotificationChannel::Email,
            NotificationChannel::InApp,
            NotificationChannel::Push,
        ]),
        Some(NotificationPriority::High),
    ).await?;
    
    println!("🔧 Notification de maintenance: {}", maintenance_id);
    
    // Notification d'urgence (priorité critique - bypass des préférences)
    let emergency_notification = Notification {
        id: Uuid::new_v4().to_string(),
        user_id: "all_users".to_string(),
        notification_type: NotificationType::ServiceDegraded,
        title: "🚨 Service dégradé".to_string(),
        message: "Nous rencontrons actuellement des difficultés techniques. Nos équipes travaillent à la résolution du problème.".to_string(),
        data: Some(serde_json::json!({
            "incident_id": "INC-2024-001",
            "severity": "high",
            "estimated_resolution": "30 minutes",
            "status_page": "https://status.streamserver.com"
        })),
        priority: NotificationPriority::Critical,
        channels: vec![
            NotificationChannel::InApp,
            NotificationChannel::Push,
            NotificationChannel::WebSocket,
            NotificationChannel::Email,
        ],
        created_at: SystemTime::now().duration_since(UNIX_EPOCH)?.as_secs(),
        expires_at: Some(
            (SystemTime::now() + Duration::from_secs(3600))
                .duration_since(UNIX_EPOCH)?
                .as_secs()
        ), // Expire dans 1 heure
        read: false,
        delivered: false,
        delivery_attempts: 0,
        tags: vec!["system".to_string(), "incident".to_string(), "critical".to_string()],
    };
    
    let emergency_id = notification_service.send_notification(emergency_notification).await?;
    println!("🚨 Notification d'urgence: {}", emergency_id);
    
    Ok(())
}
```

### Statistiques et monitoring

```rust
async fn example_notification_statistics() -> Result<(), Box<dyn std::error::Error>> {
    let config = Arc::new(Config::from_env()?);
    let notification_service = NotificationService::new(config);
    
    // Attendre un peu pour accumuler des statistiques
    tokio::time::sleep(Duration::from_secs(60)).await;
    
    // Récupérer les statistiques
    let stats = notification_service.get_statistics().await;
    
    println!("📊 Statistiques de notifications:");
    println!("  Total envoyé: {}", stats.total_sent);
    println!("  Total livré: {}", stats.total_delivered);
    println!("  Total lu: {}", stats.total_read);
    println!("  Taux de livraison: {:.2}%", stats.delivery_rate * 100.0);
    println!("  Taux de lecture: {:.2}%", stats.read_rate * 100.0);
    
    // Statistiques par canal
    println!("\n📱 Statistiques par canal:");
    for (channel, channel_stats) in &stats.channel_stats {
        println!("  {:?}:", channel);
        println!("    Envoyé: {}", channel_stats.sent);
        println!("    Livré: {}", channel_stats.delivered);
        println!("    Échec: {}", channel_stats.failed);
        println!("    Temps moyen: {}ms", channel_stats.average_delivery_time_ms);
        
        let success_rate = if channel_stats.sent > 0 {
            (channel_stats.delivered as f32 / channel_stats.sent as f32) * 100.0
        } else { 0.0 };
        println!("    Taux de succès: {:.2}%", success_rate);
    }
    
    // Statistiques par type
    println!("\n📋 Statistiques par type:");
    for (notification_type, type_stats) in &stats.type_stats {
        println!("  {}:", notification_type);
        println!("    Envoyé: {}", type_stats.sent);
        println!("    Lu: {}", type_stats.read);
        
        let read_rate = if type_stats.sent > 0 {
            (type_stats.read as f32 / type_stats.sent as f32) * 100.0
        } else { 0.0 };
        println!("    Taux de lecture: {:.2}%", read_rate);
        println!("    Temps moyen avant lecture: {}min", type_stats.average_read_time_minutes);
    }
    
    // Échecs récents
    if !stats.recent_failures.is_empty() {
        println!("\n❌ Échecs récents:");
        for failure in stats.recent_failures.iter().take(10) {
            println!("  - {} via {:?}: {} (tentative {})", 
                failure.notification_id, 
                failure.channel,
                failure.error,
                failure.retry_count
            );
        }
    }
    
    Ok(())
}
```

### Intégration avec WebSocket

```rust
use stream_server::notifications::NotificationService;
use stream_server::streaming::websocket::WebSocketManager;

async fn example_websocket_notifications() -> Result<(), Box<dyn std::error::Error>> {
    let config = Arc::new(Config::from_env()?);
    let notification_service = Arc::new(NotificationService::new(config));
    let ws_manager = Arc::new(WebSocketManager::new());
    
    // Démarrer les services
    notification_service.start_delivery_workers().await;
    
    // Récupérer le récepteur WebSocket pour les notifications
    let mut notification_receiver = notification_service.get_websocket_receiver().await;
    
    // Tâche pour relayer les notifications vers WebSocket
    let ws_manager_clone = ws_manager.clone();
    tokio::spawn(async move {
        while let Ok(notification) = notification_receiver.recv().await {
            // Convertir la notification en événement WebSocket
            let ws_event = WebSocketEvent::ServerMessage {
                message: format!("{}: {}", notification.title, notification.message),
                level: match notification.priority {
                    NotificationPriority::Critical | NotificationPriority::Emergency => MessageLevel::Error,
                    NotificationPriority::High => MessageLevel::Warning,
                    _ => MessageLevel::Info,
                },
            };
            
            // Envoyer à l'utilisateur spécifique ou broadcast
            if notification.user_id == "all_users" {
                ws_manager_clone.broadcast_event(ws_event).await;
            } else {
                ws_manager_clone.send_to_user(&notification.user_id, ws_event).await;
            }
        }
    });
    
    // Test: envoyer une notification qui sera relayée via WebSocket
    let test_notification = Notification {
        id: Uuid::new_v4().to_string(),
        user_id: "websocket_user".to_string(),
        notification_type: NotificationType::NewLike,
        title: "Nouveau like!".to_string(),
        message: "Votre piste a reçu un nouveau like!".to_string(),
        data: None,
        priority: NotificationPriority::Normal,
        channels: vec![NotificationChannel::WebSocket],
        created_at: SystemTime::now().duration_since(UNIX_EPOCH)?.as_secs(),
        expires_at: None,
        read: false,
        delivered: false,
        delivery_attempts: 0,
        tags: vec!["social".to_string()],
    };
    
    notification_service.send_notification(test_notification).await?;
    
    println!("🔄 Notification envoyée via WebSocket");
    
    Ok(())
}
```

## Intégration

### Avec le serveur principal

```rust
// Dans main.rs
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let config = Arc::new(Config::from_env()?);
    
    // Créer le service de notifications
    let notification_service = Arc::new(NotificationService::new(config.clone()));
    
    // Démarrer les workers de livraison
    notification_service.start_delivery_workers().await;
    
    // Router avec endpoints de notifications
    let app = Router::new()
        // Endpoints de notifications
        .route("/api/notifications", get(get_user_notifications))
        .route("/api/notifications/:id/read", post(mark_notification_read))
        .route("/api/notifications/preferences", get(get_notification_preferences))
        .route("/api/notifications/preferences", put(update_notification_preferences))
        .route("/api/notifications/send", post(send_notification_handler))
        
        // Stats admin
        .route("/admin/notifications/stats", get(get_notification_stats))
        
        .with_state(AppState {
            notification_service,
            // ... autres composants
        });
    
    // Démarrer le serveur
    let listener = tokio::net::TcpListener::bind("0.0.0.0:8082").await?;
    axum::serve(listener, app).await?;
    
    Ok(())
}

// Handlers d'API
async fn get_user_notifications(
    Query(params): Query<HashMap<String, String>>,
    State(state): State<AppState>,
    claims: Extension<Claims>,
) -> Result<Json<Vec<Notification>>, StatusCode> {
    let limit = params.get("limit")
        .and_then(|l| l.parse().ok())
        .unwrap_or(50);
    let offset = params.get("offset")
        .and_then(|o| o.parse().ok())
        .unwrap_or(0);
    let unread_only = params.get("unread_only")
        .map(|u| u == "true")
        .unwrap_or(false);
    
    let notifications = state.notification_service
        .get_user_notifications(&claims.sub, Some(limit), Some(offset), unread_only)
        .await;
    
    Ok(Json(notifications))
}

async fn mark_notification_read(
    Path(notification_id): Path<String>,
    State(state): State<AppState>,
    claims: Extension<Claims>,
) -> Result<Json<serde_json::Value>, StatusCode> {
    match state.notification_service
        .mark_as_read(&notification_id, &claims.sub)
        .await
    {
        Ok(_) => Ok(Json(json!({"success": true}))),
        Err(_) => Err(StatusCode::NOT_FOUND),
    }
}
```

### Avec l'API Go

```go
// Structures correspondantes
type Notification struct {
    ID               string                 `json:"id"`
    UserID          string                 `json:"user_id"`
    NotificationType string                 `json:"notification_type"`
    Title           string                 `json:"title"`
    Message         string                 `json:"message"`
    Data            map[string]interface{} `json:"data,omitempty"`
    Priority        string                 `json:"priority"`
    Channels        []string               `json:"channels"`
    CreatedAt       int64                  `json:"created_at"`
    ExpiresAt       *int64                 `json:"expires_at,omitempty"`
    Read            bool                   `json:"read"`
    Delivered       bool                   `json:"delivered"`
    DeliveryAttempts int                   `json:"delivery_attempts"`
    Tags            []string               `json:"tags"`
}

// Client de notifications
type NotificationClient struct {
    baseURL    string
    httpClient *http.Client
}

func (c *NotificationClient) SendNotification(notification Notification) error {
    jsonData, err := json.Marshal(notification)
    if err != nil {
        return err
    }
    
    resp, err := c.httpClient.Post(
        fmt.Sprintf("%s/api/notifications/send", c.baseURL),
        "application/json",
        bytes.NewBuffer(jsonData),
    )
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("failed to send notification: %d", resp.StatusCode)
    }
    
    return nil
}

func (c *NotificationClient) GetUserNotifications(userID string, limit, offset int) ([]Notification, error) {
    url := fmt.Sprintf("%s/api/notifications?limit=%d&offset=%d", c.baseURL, limit, offset)
    
    resp, err := c.httpClient.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var notifications []Notification
    if err := json.NewDecoder(resp.Body).Decode(&notifications); err != nil {
        return nil, err
    }
    
    return notifications, nil
}
```

### Avec le frontend React

```typescript
// Types TypeScript
interface Notification {
  id: string;
  user_id: string;
  notification_type: string;
  title: string;
  message: string;
  data?: any;
  priority: 'Low' | 'Normal' | 'High' | 'Critical' | 'Emergency';
  channels: string[];
  created_at: number;
  expires_at?: number;
  read: boolean;
  delivered: boolean;
  delivery_attempts: number;
  tags: string[];
}

interface UserPreferences {
  user_id: string;
  enabled_channels: Record<string, boolean>;
  type_preferences: Record<string, NotificationPreference>;
  quiet_hours?: QuietHours;
  frequency_limits: Record<string, FrequencyLimit>;
  language: string;
  timezone: string;
}

// Service de notifications
export class NotificationService {
  private baseURL: string;
  
  constructor(baseURL: string) {
    this.baseURL = baseURL;
  }
  
  async getUserNotifications(limit = 50, offset = 0, unreadOnly = false): Promise<Notification[]> {
    const params = new URLSearchParams({
      limit: limit.toString(),
      offset: offset.toString(),
      unread_only: unreadOnly.toString(),
    });
    
    const response = await fetch(`${this.baseURL}/api/notifications?${params}`);
    return response.json();
  }
  
  async markAsRead(notificationId: string): Promise<void> {
    await fetch(`${this.baseURL}/api/notifications/${notificationId}/read`, {
      method: 'POST',
    });
  }
  
  async getUserPreferences(): Promise<UserPreferences> {
    const response = await fetch(`${this.baseURL}/api/notifications/preferences`);
    return response.json();
  }
  
  async updateUserPreferences(preferences: UserPreferences): Promise<void> {
    await fetch(`${this.baseURL}/api/notifications/preferences`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(preferences),
    });
  }
}

// Hook React pour les notifications
export function useNotifications() {
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [unreadCount, setUnreadCount] = useState(0);
  const [preferences, setPreferences] = useState<UserPreferences | null>(null);
  const service = new NotificationService('/api');
  
  useEffect(() => {
    loadNotifications();
    loadPreferences();
  }, []);
  
  const loadNotifications = async () => {
    const data = await service.getUserNotifications();
    setNotifications(data);
    setUnreadCount(data.filter(n => !n.read).length);
  };
  
  const loadPreferences = async () => {
    const prefs = await service.getUserPreferences();
    setPreferences(prefs);
  };
  
  const markAsRead = async (notificationId: string) => {
    await service.markAsRead(notificationId);
    setNotifications(prev => 
      prev.map(n => n.id === notificationId ? { ...n, read: true } : n)
    );
    setUnreadCount(prev => Math.max(0, prev - 1));
  };
  
  const updatePreferences = async (newPreferences: UserPreferences) => {
    await service.updateUserPreferences(newPreferences);
    setPreferences(newPreferences);
  };
  
  return {
    notifications,
    unreadCount,
    preferences,
    markAsRead,
    updatePreferences,
    refresh: loadNotifications,
  };
}

// Composant de notifications
export function NotificationCenter() {
  const { notifications, unreadCount, markAsRead } = useNotifications();
  const [isOpen, setIsOpen] = useState(false);
  
  return (
    <div className="notification-center">
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="notification-trigger"
      >
        🔔
        {unreadCount > 0 && (
          <span className="notification-badge">{unreadCount}</span>
        )}
      </button>
      
      {isOpen && (
        <div className="notification-dropdown">
          <div className="notification-header">
            <h3>Notifications ({unreadCount} non lues)</h3>
          </div>
          
          <div className="notification-list">
            {notifications.map(notification => (
              <div
                key={notification.id}
                className={`notification-item ${!notification.read ? 'unread' : ''}`}
                onClick={() => markAsRead(notification.id)}
              >
                <div className="notification-content">
                  <h4>{notification.title}</h4>
                  <p>{notification.message}</p>
                  <small>
                    {new Date(notification.created_at * 1000).toLocaleString()}
                  </small>
                </div>
                
                <div className="notification-meta">
                  <span className={`priority ${notification.priority.toLowerCase()}`}>
                    {notification.priority}
                  </span>
                  {!notification.read && <span className="unread-dot" />}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
```

Cette documentation complète du module notifications vous permet d'implémenter un système de notifications sophistiqué et flexible pour votre plateforme de streaming audio. 