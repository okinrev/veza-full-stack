# Analytics Module Documentation

Le module analytics fournit un système complet de suivi, analyse et reporting pour le serveur de streaming audio.

## Table des Matières

- [Vue d'ensemble](#vue-densemble)
- [Architecture](#architecture)
- [Types et Structures](#types-et-structures)
- [AnalyticsEngine](#analyticsengine)
- [Sessions de lecture](#sessions-de-lecture)
- [Analytics utilisateur](#analytics-utilisateur)
- [Analytics par piste](#analytics-par-piste)
- [Statistiques temps réel](#statistiques-temps-réel)
- [API Reference](#api-reference)
- [Exemples d'utilisation](#exemples-dutilisation)
- [Intégration](#intégration)

## Vue d'ensemble

Le système d'analytics suit et analyse :
- **Sessions de lecture** : Comportement d'écoute en temps réel
- **Analytics utilisateur** : Habitudes et préférences individuelles
- **Analytics par piste** : Performance et popularité des contenus
- **Statistiques globales** : Métriques de performance du service
- **Données géographiques** : Répartition géographique des utilisateurs
- **Données de performance** : Qualité de streaming et santé du buffer

## Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│  Play Session   │    │  User Analytics  │    │ Track Analytics │
│   (Real-time)   │    │  (Aggregated)    │    │  (Aggregated)   │
└─────────────────┘    └──────────────────┘    └─────────────────┘
         │                        │                        │
         └───────────────────────────────────────────────────────┘
                                  │
                    ┌─────────────────────────┐
                    │   AnalyticsEngine       │
                    │                         │
                    │ - Session Management    │
                    │ - Data Aggregation      │
                    │ - Report Generation     │
                    │ - Background Tasks      │
                    └─────────────────────────┘
                                  │
                    ┌─────────────────────────┐
                    │     SQLite Database     │
                    │                         │
                    │ - Session History       │
                    │ - User Profiles         │
                    │ - Track Statistics      │
                    │ - Performance Metrics   │
                    └─────────────────────────┘
```

## Types et Structures

### PlaySession

```rust
pub struct PlaySession {
    pub session_id: Uuid,
    pub user_id: Option<String>,
    pub track_id: String,
    pub client_ip: String,
    pub user_agent: Option<String>,
    pub started_at: SystemTime,
    pub last_update: SystemTime,
    pub duration_played_ms: u64,
    pub total_duration_ms: u64,
    pub completion_percentage: f32,
    pub quality: String,
    pub platform: Platform,
    pub location: Option<GeoLocation>,
    pub referrer: Option<String>,
    pub ended: bool,
    pub skip_reason: Option<SkipReason>,
}
```

**Champs détaillés :**
- `session_id` : Identifiant unique de session
- `user_id` : ID utilisateur (optionnel pour les invités)
- `track_id` : Identifiant de la piste audio
- `client_ip` : Adresse IP du client
- `user_agent` : Agent utilisateur du navigateur
- `started_at` : Timestamp de début de session
- `last_update` : Dernière mise à jour de la session
- `duration_played_ms` : Durée réellement écoutée
- `total_duration_ms` : Durée totale de la piste
- `completion_percentage` : Pourcentage d'achèvement (0-100)
- `quality` : Qualité audio (high, medium, low, mobile)
- `platform` : Plateforme d'écoute
- `location` : Localisation géographique
- `referrer` : URL de référence
- `ended` : Session terminée ou non
- `skip_reason` : Raison d'arrêt prématuré

### Platform

```rust
pub enum Platform {
    Web,        // Navigateur web
    Mobile,     // Application mobile
    Desktop,    // Application desktop
    Embedded,   // Dispositif embarqué
    Unknown,    // Plateforme non identifiée
}
```

### SkipReason

```rust
pub enum SkipReason {
    UserSkip,       // Skip volontaire
    NetworkError,   // Erreur réseau
    QualityIssue,   // Problème de qualité
    Timeout,        // Timeout de connexion
    Other(String),  // Autre raison
}
```

### GeoLocation

```rust
pub struct GeoLocation {
    pub country: Option<String>,    // Code pays (FR, US, etc.)
    pub region: Option<String>,     // Région/État
    pub city: Option<String>,       // Ville
    pub latitude: Option<f64>,      // Latitude
    pub longitude: Option<f64>,     // Longitude
}
```

### TrackAnalytics

```rust
pub struct TrackAnalytics {
    pub track_id: String,
    pub total_plays: u64,                                    // Nombre total de lectures
    pub unique_listeners: u64,                               // Auditeurs uniques
    pub total_duration_played_ms: u64,                      // Durée totale écoutée
    pub average_completion_rate: f32,                        // Taux d'achèvement moyen
    pub peak_concurrent_listeners: u32,                      // Pic d'auditeurs simultanés
    pub plays_by_hour: HashMap<u32, u64>,                  // Répartition par heure (0-23)
    pub plays_by_day: HashMap<String, u64>,                // Répartition par jour (YYYY-MM-DD)
    pub skip_rate: f32,                                     // Taux de skip
    pub quality_distribution: HashMap<String, u64>,         // Répartition par qualité
    pub geographic_distribution: HashMap<String, u64>,      // Répartition géographique
    pub platform_distribution: HashMap<Platform, u64>,     // Répartition par plateforme
    pub last_updated: SystemTime,                           // Dernière mise à jour
}
```

### UserAnalytics

```rust
pub struct UserAnalytics {
    pub user_id: String,
    pub total_listening_time_ms: u64,                       // Temps d'écoute total
    pub tracks_played: u64,                                 // Nombre de pistes jouées
    pub unique_tracks: u64,                                 // Pistes uniques écoutées
    pub average_session_duration_ms: u64,                  // Durée moyenne des sessions
    pub favorite_genres: Vec<String>,                       // Genres préférés
    pub listening_patterns: ListeningPatterns,              // Motifs d'écoute
    pub device_preferences: HashMap<Platform, u64>,         // Préférences de dispositif
    pub quality_preference: String,                         // Qualité préférée
    pub most_active_hours: Vec<u32>,                       // Heures les plus actives
    pub discovery_rate: f32,                               // Taux de découverte (0-1)
    pub last_activity: SystemTime,                         // Dernière activité
}
```

### ListeningPatterns

```rust
pub struct ListeningPatterns {
    pub average_skip_time_ms: u64,              // Temps moyen avant skip
    pub completion_rate: f32,                   // Taux d'achèvement moyen
    pub binge_listening_tendency: f32,          // Tendance à l'écoute prolongée (0-1)
    pub peak_listening_hour: u32,               // Heure d'écoute principale
    pub weekend_vs_weekday_ratio: f32,          // Ratio weekend/semaine
}
```

### RealTimeStats

```rust
pub struct RealTimeStats {
    pub current_listeners: u32,                              // Auditeurs actuels
    pub streams_started_last_hour: u64,                     // Streams démarrés dernière heure
    pub bytes_served_last_hour: u64,                        // Bytes servis dernière heure
    pub top_tracks_now: Vec<TopTrackNow>,                   // Top pistes en ce moment
    pub geographic_activity: HashMap<String, u32>,           // Activité par pays
    pub quality_distribution_now: HashMap<String, u32>,     // Répartition qualité actuelle
    pub error_rate_percentage: f32,                         // Taux d'erreur
    pub average_buffer_health: f32,                         // Santé moyenne du buffer
}
```

### TopTrackNow

```rust
pub struct TopTrackNow {
    pub track_id: String,           // ID de la piste
    pub title: String,              // Titre
    pub artist: String,             // Artiste
    pub current_listeners: u32,     // Auditeurs actuels
    pub started_in_last_hour: u64,  // Démarrages dernière heure
}
```

## AnalyticsEngine

### Structure principale

```rust
pub struct AnalyticsEngine {
    db_pool: SqlitePool,                                        // Pool de connexions DB
    active_sessions: Arc<RwLock<HashMap<Uuid, PlaySession>>>,  // Sessions actives
    track_analytics: Arc<RwLock<HashMap<String, TrackAnalytics>>>, // Analytics par piste
    user_analytics: Arc<RwLock<HashMap<String, UserAnalytics>>>,   // Analytics utilisateur
    realtime_stats: Arc<RwLock<RealTimeStats>>,                // Stats temps réel
    config: Arc<Config>,                                        // Configuration
}
```

### Cycle de vie des sessions

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  Session Start  │───►│  Progress       │───►│  Session End    │
│                 │    │  Updates        │    │                 │
│ - Create UUID   │    │                 │    │ - Calculate     │
│ - Store client  │    │ - Update time   │    │   completion    │
│ - Begin tracking│    │ - Track buffer  │    │ - Update stats  │
│                 │    │ - Monitor skip  │    │ - Cleanup       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Sessions de lecture

### Démarrage de session

```rust
let session_id = analytics_engine.start_play_session(
    Some("user123".to_string()),     // User ID (optionnel)
    "track456".to_string(),          // Track ID
    "192.168.1.100".to_string(),     // Client IP
    Some("Mozilla/5.0...".to_string()), // User Agent
    180_000,                         // Durée totale (ms)
    "high".to_string(),              // Qualité
    Platform::Web,                   // Plateforme
    Some("https://app.com".to_string()), // Referrer
).await;
```

### Mise à jour de progression

```rust
analytics_engine.update_play_progress(
    session_id,
    45_000,        // Position actuelle (ms)
    Some(85.5),    // Santé du buffer (%)
).await;
```

### Fin de session

```rust
analytics_engine.end_play_session(
    session_id,
    Some(SkipReason::UserSkip), // Raison d'arrêt (optionnel)
).await;
```

## Analytics utilisateur

### Calcul automatique des métriques

```rust
impl UserAnalytics {
    // Temps d'écoute moyen par session
    pub fn average_session_duration(&self) -> Duration {
        Duration::from_millis(self.average_session_duration_ms)
    }
    
    // Taux de découverte (nouvelles pistes / total)
    pub fn discovery_rate(&self) -> f32 {
        if self.tracks_played == 0 { 0.0 }
        else { self.unique_tracks as f32 / self.tracks_played as f32 }
    }
    
    // Heure d'écoute principale
    pub fn peak_listening_hour(&self) -> u32 {
        self.listening_patterns.peak_listening_hour
    }
}
```

### Patterns d'écoute

L'engine analyse automatiquement :
- **Tendance au binge** : Basée sur la durée des sessions
- **Heures d'activité** : Distribution des écoutes par heure
- **Ratio weekend/semaine** : Différence de comportement
- **Taux de skip** : Fréquence d'arrêt prématuré

## Analytics par piste

### Métriques automatiques

```rust
impl TrackAnalytics {
    // Taux d'achèvement moyen
    pub fn completion_rate(&self) -> f32 {
        self.average_completion_rate
    }
    
    // Popularité relative (auditeurs uniques / total plays)
    pub fn virality_score(&self) -> f32 {
        if self.total_plays == 0 { 0.0 }
        else { self.unique_listeners as f32 / self.total_plays as f32 }
    }
    
    // Heure de pic d'écoute
    pub fn peak_listening_hour(&self) -> Option<u32> {
        self.plays_by_hour.iter()
            .max_by_key(|(_, &count)| count)
            .map(|(&hour, _)| hour)
    }
}
```

## Statistiques temps réel

### Mise à jour automatique

```rust
impl AnalyticsEngine {
    pub async fn refresh_realtime_stats(&self) {
        let mut stats = self.realtime_stats.write().await;
        
        // Calculer auditeurs actuels
        let active_sessions = self.active_sessions.read().await;
        stats.current_listeners = active_sessions.len() as u32;
        
        // Top pistes actuelles
        let mut track_counts: HashMap<String, u32> = HashMap::new();
        for session in active_sessions.values() {
            *track_counts.entry(session.track_id.clone()).or_insert(0) += 1;
        }
        
        // Convertir en TopTrackNow
        stats.top_tracks_now = track_counts.into_iter()
            .map(|(track_id, count)| TopTrackNow {
                track_id: track_id.clone(),
                title: self.get_track_title(&track_id).unwrap_or("Unknown".to_string()),
                artist: self.get_track_artist(&track_id).unwrap_or("Unknown".to_string()),
                current_listeners: count,
                started_in_last_hour: self.count_recent_starts(&track_id).await,
            })
            .collect();
    }
}
```

## API Reference

### AnalyticsEngine Methods

#### `new(database_url: &str, config: Arc<Config>) -> Result<Self, sqlx::Error>`
Crée une nouvelle instance de l'AnalyticsEngine.

#### `start_play_session(...) -> Uuid`
Démarre une nouvelle session de lecture et retourne l'ID de session.

**Paramètres :**
- `user_id` : ID utilisateur (optionnel)
- `track_id` : ID de la piste
- `client_ip` : Adresse IP du client
- `user_agent` : User agent du navigateur
- `total_duration_ms` : Durée totale de la piste
- `quality` : Qualité de streaming
- `platform` : Plateforme d'écoute
- `referrer` : URL de référence

#### `update_play_progress(session_id, duration_played_ms, buffer_health)`
Met à jour la progression d'une session.

#### `end_play_session(session_id, skip_reason)`
Termine une session de lecture et met à jour les analytics.

#### `get_track_analytics(track_id) -> Option<TrackAnalytics>`
Récupère les analytics d'une piste spécifique.

#### `get_user_analytics(user_id) -> Option<UserAnalytics>`
Récupère les analytics d'un utilisateur spécifique.

#### `get_realtime_stats() -> RealTimeStats`
Récupère les statistiques temps réel.

#### `generate_period_report(start_date, end_date) -> Result<serde_json::Value, sqlx::Error>`
Génère un rapport pour une période donnée.

#### `cleanup_old_data(older_than_days) -> Result<(), sqlx::Error>`
Nettoie les données anciennes.

#### `start_background_tasks()`
Démarre les tâches de fond (nettoyage, agrégation).

## Exemples d'utilisation

### Session de lecture complète

```rust
use stream_server::analytics::{AnalyticsEngine, Platform};

async fn example_full_session() -> Result<(), Box<dyn std::error::Error>> {
    let config = Arc::new(Config::from_env()?);
    let analytics = AnalyticsEngine::new("sqlite:analytics.db", config).await?;
    
    // Démarrer une session
    let session_id = analytics.start_play_session(
        Some("user_123".to_string()),
        "track_456".to_string(),
        "192.168.1.100".to_string(),
        Some("Mozilla/5.0 (Windows NT 10.0; Win64; x64)".to_string()),
        210_000, // 3 minutes 30 secondes
        "high".to_string(),
        Platform::Web,
        Some("https://musicapp.com/playlist/123".to_string()),
    ).await;
    
    println!("🎵 Session démarrée: {}", session_id);
    
    // Simuler la progression
    for position in (0..=210_000).step_by(30_000) {
        analytics.update_play_progress(
            session_id,
            position,
            Some(90.0 + (position as f32 / 210_000.0) * 10.0), // Buffer 90-100%
        ).await;
        
        println!("⏱️  Position: {}ms", position);
        tokio::time::sleep(Duration::from_millis(100)).await; // Simulation
    }
    
    // Terminer la session
    analytics.end_play_session(session_id, None).await;
    println!("✅ Session terminée");
    
    Ok(())
}
```

### Analytics utilisateur détaillées

```rust
async fn example_user_analytics() -> Result<(), Box<dyn std::error::Error>> {
    let config = Arc::new(Config::from_env()?);
    let analytics = AnalyticsEngine::new("sqlite:analytics.db", config).await?;
    
    // Récupérer les analytics d'un utilisateur
    if let Some(user_analytics) = analytics.get_user_analytics("user_123").await {
        println!("👤 Analytics pour l'utilisateur: {}", user_analytics.user_id);
        println!("🎧 Temps d'écoute total: {:.1}h", user_analytics.total_listening_time_ms as f64 / 3600000.0);
        println!("🎵 Pistes jouées: {}", user_analytics.tracks_played);
        println!("🔍 Pistes uniques: {}", user_analytics.unique_tracks);
        println!("📊 Taux de découverte: {:.1}%", user_analytics.discovery_rate * 100.0);
        println!("⏱️  Durée moyenne des sessions: {:.1}min", user_analytics.average_session_duration_ms as f64 / 60000.0);
        
        // Habitudes d'écoute
        println!("\n📈 Patterns d'écoute:");
        println!("  - Taux d'achèvement: {:.1}%", user_analytics.listening_patterns.completion_rate * 100.0);
        println!("  - Tendance au binge: {:.1}%", user_analytics.listening_patterns.binge_listening_tendency * 100.0);
        println!("  - Heure de pic: {}h", user_analytics.listening_patterns.peak_listening_hour);
        
        // Genres préférés
        println!("\n🎶 Genres préférés:");
        for (i, genre) in user_analytics.favorite_genres.iter().take(5).enumerate() {
            println!("  {}. {}", i + 1, genre);
        }
        
        // Plateformes utilisées
        println!("\n📱 Plateformes:");
        for (platform, count) in &user_analytics.device_preferences {
            println!("  - {:?}: {} sessions", platform, count);
        }
    }
    
    Ok(())
}
```

### Analytics par piste

```rust
async fn example_track_analytics() -> Result<(), Box<dyn std::error::Error>> {
    let config = Arc::new(Config::from_env()?);
    let analytics = AnalyticsEngine::new("sqlite:analytics.db", config).await?;
    
    // Récupérer les analytics d'une piste
    if let Some(track_analytics) = analytics.get_track_analytics("track_456").await {
        println!("🎵 Analytics pour la piste: {}", track_analytics.track_id);
        println!("▶️  Total plays: {}", track_analytics.total_plays);
        println!("👥 Auditeurs uniques: {}", track_analytics.unique_listeners);
        println!("📊 Taux d'achèvement: {:.1}%", track_analytics.average_completion_rate * 100.0);
        println!("⏭️  Taux de skip: {:.1}%", track_analytics.skip_rate * 100.0);
        println!("🔥 Pic d'auditeurs simultanés: {}", track_analytics.peak_concurrent_listeners);
        
        // Durée totale écoutée
        let total_hours = track_analytics.total_duration_played_ms as f64 / 3600000.0;
        println!("⏱️  Durée totale écoutée: {:.1}h", total_hours);
        
        // Répartition par qualité
        println!("\n🎚️  Répartition par qualité:");
        for (quality, count) in &track_analytics.quality_distribution {
            let percentage = (*count as f64 / track_analytics.total_plays as f64) * 100.0;
            println!("  - {}: {} ({:.1}%)", quality, count, percentage);
        }
        
        // Top 5 pays
        println!("\n🌍 Top pays:");
        let mut countries: Vec<_> = track_analytics.geographic_distribution.iter().collect();
        countries.sort_by_key(|(_, &count)| std::cmp::Reverse(count));
        for (i, (country, count)) in countries.iter().take(5).enumerate() {
            println!("  {}. {}: {} plays", i + 1, country, count);
        }
        
        // Heures de pic
        println!("\n⏰ Heures les plus populaires:");
        let mut hours: Vec<_> = track_analytics.plays_by_hour.iter().collect();
        hours.sort_by_key(|(_, &count)| std::cmp::Reverse(count));
        for (hour, count) in hours.iter().take(3) {
            println!("  - {}h: {} plays", hour, count);
        }
    }
    
    Ok(())
}
```

### Statistiques temps réel

```rust
async fn example_realtime_stats() -> Result<(), Box<dyn std::error::Error>> {
    let config = Arc::new(Config::from_env()?);
    let analytics = AnalyticsEngine::new("sqlite:analytics.db", config).await?;
    
    // Démarrer les tâches de fond
    analytics.start_background_tasks().await;
    
    // Afficher les stats temps réel
    loop {
        let stats = analytics.get_realtime_stats().await;
        
        println!("🔴 LIVE - {} auditeurs", stats.current_listeners);
        println!("📈 Streams dernière heure: {}", stats.streams_started_last_hour);
        println!("💾 Bytes servis: {:.1} MB", stats.bytes_served_last_hour as f64 / 1_000_000.0);
        println!("❌ Taux d'erreur: {:.2}%", stats.error_rate_percentage);
        println!("📊 Santé buffer moyenne: {:.1}%", stats.average_buffer_health);
        
        // Top pistes en ce moment
        println!("\n🎵 Top pistes actuellement:");
        for (i, track) in stats.top_tracks_now.iter().take(5).enumerate() {
            println!("  {}. {} - {} ({} auditeurs)", 
                i + 1, track.title, track.artist, track.current_listeners);
        }
        
        // Activité géographique
        println!("\n🌍 Activité par pays:");
        let mut geo_activity: Vec<_> = stats.geographic_activity.iter().collect();
        geo_activity.sort_by_key(|(_, &count)| std::cmp::Reverse(count));
        for (country, count) in geo_activity.iter().take(5) {
            println!("  - {}: {} auditeurs", country, count);
        }
        
        println!("\n" + "─".repeat(50));
        tokio::time::sleep(Duration::from_secs(10)).await;
    }
}
```

### Rapport de période

```rust
async fn example_period_report() -> Result<(), Box<dyn std::error::Error>> {
    let config = Arc::new(Config::from_env()?);
    let analytics = AnalyticsEngine::new("sqlite:analytics.db", config).await?;
    
    // Rapport des 7 derniers jours
    let end_date = SystemTime::now();
    let start_date = end_date - Duration::from_secs(7 * 24 * 3600);
    
    let report = analytics.generate_period_report(start_date, end_date).await?;
    
    println!("📋 Rapport des 7 derniers jours:");
    println!("{}", serde_json::to_string_pretty(&report)?);
    
    Ok(())
}
```

## Intégration

### Avec le serveur principal

```rust
// Dans main.rs
let analytics_engine = Arc::new(
    AnalyticsEngine::new(&config.database.url, config.clone()).await?
);

// Démarrer les tâches de fond
analytics_engine.start_background_tasks().await;

// Inclure dans l'état de l'application
let app_state = AppState {
    analytics: analytics_engine,
    // ... autres composants
};
```

### Avec l'API de streaming

```rust
// Middleware pour tracker les sessions
async fn track_streaming_session(
    State(state): State<AppState>,
    req: Request,
    next: Next,
) -> Response {
    let session_id = state.analytics.start_play_session(
        extract_user_id(&req),
        extract_track_id(&req),
        extract_client_ip(&req),
        extract_user_agent(&req),
        get_track_duration(&req).await,
        extract_quality(&req),
        detect_platform(&req),
        extract_referrer(&req),
    ).await;
    
    // Stocker l'ID de session pour les mises à jour
    req.extensions_mut().insert(session_id);
    
    let response = next.run(req).await;
    
    // Optionnel: terminer la session si erreur
    if response.status().is_server_error() {
        state.analytics.end_play_session(
            session_id, 
            Some(SkipReason::NetworkError)
        ).await;
    }
    
    response
}
```

### Avec l'API REST

```rust
// Endpoint pour analytics utilisateur
async fn get_user_analytics(
    Path(user_id): Path<String>,
    State(state): State<AppState>,
) -> Result<Json<UserAnalytics>, StatusCode> {
    match state.analytics.get_user_analytics(&user_id).await {
        Some(analytics) => Ok(Json(analytics)),
        None => Err(StatusCode::NOT_FOUND),
    }
}

// Endpoint pour stats temps réel
async fn get_realtime_stats(
    State(state): State<AppState>,
) -> Json<RealTimeStats> {
    Json(state.analytics.get_realtime_stats().await)
}

// Endpoint pour rapport de période
async fn get_period_report(
    Query(params): Query<HashMap<String, String>>,
    State(state): State<AppState>,
) -> Result<Json<serde_json::Value>, StatusCode> {
    let days = params.get("days")
        .and_then(|d| d.parse::<u64>().ok())
        .unwrap_or(7);
    
    let end_date = SystemTime::now();
    let start_date = end_date - Duration::from_secs(days * 24 * 3600);
    
    match state.analytics.generate_period_report(start_date, end_date).await {
        Ok(report) => Ok(Json(report)),
        Err(_) => Err(StatusCode::INTERNAL_SERVER_ERROR),
    }
}
```

### Avec le frontend React

```typescript
// Types correspondants
interface PlaySession {
  sessionId: string;
  userId?: string;
  trackId: string;
  clientIp: string;
  userAgent?: string;
  startedAt: number;
  lastUpdate: number;
  durationPlayedMs: number;
  totalDurationMs: number;
  completionPercentage: number;
  quality: string;
  platform: Platform;
  location?: GeoLocation;
  referrer?: string;
  ended: boolean;
  skipReason?: SkipReason;
}

// Service d'analytics
export class AnalyticsService {
  private ws: WebSocket | null = null;
  
  async startSession(trackId: string, quality: string): Promise<string> {
    const response = await fetch('/api/analytics/session/start', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ trackId, quality }),
    });
    const { sessionId } = await response.json();
    return sessionId;
  }
  
  async updateProgress(sessionId: string, position: number, bufferHealth?: number) {
    await fetch(`/api/analytics/session/${sessionId}/progress`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ position, bufferHealth }),
    });
  }
  
  async endSession(sessionId: string, skipReason?: SkipReason) {
    await fetch(`/api/analytics/session/${sessionId}/end`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ skipReason }),
    });
  }
  
  async getUserAnalytics(userId: string): Promise<UserAnalytics> {
    const response = await fetch(`/api/analytics/users/${userId}`);
    return response.json();
  }
  
  async getRealTimeStats(): Promise<RealTimeStats> {
    const response = await fetch('/api/analytics/realtime');
    return response.json();
  }
  
  // WebSocket pour stats temps réel
  subscribeToRealTimeStats(callback: (stats: RealTimeStats) => void) {
    this.ws = new WebSocket('/api/analytics/realtime/ws');
    this.ws.onmessage = (event) => {
      const stats = JSON.parse(event.data);
      callback(stats);
    };
  }
}
```

Cette documentation complète du module analytics vous permet d'implémenter un système de suivi et d'analyse complet pour votre plateforme de streaming audio. 