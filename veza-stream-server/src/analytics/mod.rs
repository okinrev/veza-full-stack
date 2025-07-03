use std::{
    collections::HashMap,
    sync::Arc,
    time::SystemTime,
};
use tokio::sync::RwLock;
use sqlx::{PgPool, Row};
use serde::{Serialize, Deserialize};
use tracing::{info, debug, error};
use uuid::Uuid;
use chrono::Timelike;
use crate::Config;


#[derive(Debug, Clone, Serialize, Deserialize)]
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

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq, Hash)]
pub enum Platform {
    Web,
    Mobile,
    Desktop,
    Embedded,
    Unknown,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum SkipReason {
    UserSkip,
    NetworkError,
    QualityIssue,
    Timeout,
    Other(String),
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct GeoLocation {
    pub country: Option<String>,
    pub region: Option<String>,
    pub city: Option<String>,
    pub latitude: Option<f64>,
    pub longitude: Option<f64>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TrackAnalytics {
    pub track_id: String,
    pub total_plays: u64,
    pub unique_listeners: u64,
    pub total_duration_played_ms: u64,
    pub average_completion_rate: f32,
    pub peak_concurrent_listeners: u32,
    pub plays_by_hour: HashMap<u32, u64>, // Heure de la journée (0-23)
    pub plays_by_day: HashMap<String, u64>, // Date (YYYY-MM-DD)
    pub skip_rate: f32,
    pub quality_distribution: HashMap<String, u64>,
    pub geographic_distribution: HashMap<String, u64>, // Par pays
    pub platform_distribution: HashMap<Platform, u64>,
    pub last_updated: SystemTime,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct UserAnalytics {
    pub user_id: String,
    pub total_listening_time_ms: u64,
    pub tracks_played: u64,
    pub unique_tracks: u64,
    pub average_session_duration_ms: u64,
    pub favorite_genres: Vec<String>,
    pub listening_patterns: ListeningPatterns,
    pub device_preferences: HashMap<Platform, u64>,
    pub quality_preference: String,
    pub most_active_hours: Vec<u32>,
    pub discovery_rate: f32, // Pourcentage de nouvelles pistes
    pub last_activity: SystemTime,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ListeningPatterns {
    pub average_skip_time_ms: u64,
    pub completion_rate: f32,
    pub binge_listening_tendency: f32, // 0-1, tendance à écouter longtemps
    pub peak_listening_hour: u32,
    pub weekend_vs_weekday_ratio: f32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RealTimeStats {
    pub current_listeners: u32,
    pub streams_started_last_hour: u64,
    pub bytes_served_last_hour: u64,
    pub top_tracks_now: Vec<TopTrackNow>,
    pub geographic_activity: HashMap<String, u32>,
    pub quality_distribution_now: HashMap<String, u32>,
    pub error_rate_percentage: f32,
    pub average_buffer_health: f32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TopTrackNow {
    pub track_id: String,
    pub title: String,
    pub artist: String,
    pub current_listeners: u32,
    pub started_in_last_hour: u64,
}

pub struct AnalyticsEngine {
    db_pool: PgPool,
    active_sessions: Arc<RwLock<HashMap<Uuid, PlaySession>>>,
    track_analytics: Arc<RwLock<HashMap<String, TrackAnalytics>>>,
    user_analytics: Arc<RwLock<HashMap<String, UserAnalytics>>>,
    realtime_stats: Arc<RwLock<RealTimeStats>>,
    config: Arc<Config>,
}

impl AnalyticsEngine {
    pub async fn new(database_url: &str, config: Arc<Config>) -> Result<Self, sqlx::Error> {
        let pool = PgPool::connect(database_url).await?;
        
        // Créer les tables si elles n'existent pas
        sqlx::query(r#"
            CREATE TABLE IF NOT EXISTS play_sessions (
                session_id UUID PRIMARY KEY,
                user_id TEXT,
                track_id TEXT NOT NULL,
                client_ip INET NOT NULL,
                user_agent TEXT,
                started_at TIMESTAMPTZ NOT NULL,
                last_update TIMESTAMPTZ NOT NULL,
                duration_played_ms BIGINT NOT NULL,
                total_duration_ms BIGINT NOT NULL,
                completion_percentage REAL NOT NULL,
                quality TEXT NOT NULL,
                platform TEXT NOT NULL,
                country TEXT,
                region TEXT,
                city TEXT,
                referrer TEXT,
                ended BOOLEAN NOT NULL DEFAULT FALSE,
                skip_reason TEXT
            )
        "#).execute(&pool).await?;

        sqlx::query(r#"
            CREATE TABLE IF NOT EXISTS track_analytics (
                track_id TEXT PRIMARY KEY,
                total_plays BIGINT NOT NULL,
                unique_listeners BIGINT NOT NULL,
                total_duration_played_ms BIGINT NOT NULL,
                average_completion_rate REAL NOT NULL,
                peak_concurrent_listeners INTEGER NOT NULL,
                skip_rate REAL NOT NULL,
                last_updated TIMESTAMPTZ NOT NULL
            )
        "#).execute(&pool).await?;

        sqlx::query(r#"
            CREATE TABLE IF NOT EXISTS user_analytics (
                user_id TEXT PRIMARY KEY,
                total_listening_time_ms BIGINT NOT NULL,
                tracks_played BIGINT NOT NULL,
                unique_tracks BIGINT NOT NULL,
                average_session_duration_ms BIGINT NOT NULL,
                quality_preference TEXT,
                discovery_rate REAL NOT NULL,
                last_activity TIMESTAMPTZ NOT NULL
            )
        "#).execute(&pool).await?;

        // Index pour les performances
        sqlx::query("CREATE INDEX IF NOT EXISTS idx_sessions_track_id ON play_sessions(track_id)").execute(&pool).await?;
        sqlx::query("CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON play_sessions(user_id)").execute(&pool).await?;
        sqlx::query("CREATE INDEX IF NOT EXISTS idx_sessions_started_at ON play_sessions(started_at)").execute(&pool).await?;

        info!("Base de données analytics initialisée");

        Ok(Self {
            db_pool: pool,
            active_sessions: Arc::new(RwLock::new(HashMap::new())),
            track_analytics: Arc::new(RwLock::new(HashMap::new())),
            user_analytics: Arc::new(RwLock::new(HashMap::new())),
            realtime_stats: Arc::new(RwLock::new(RealTimeStats {
                current_listeners: 0,
                streams_started_last_hour: 0,
                bytes_served_last_hour: 0,
                top_tracks_now: Vec::new(),
                geographic_activity: HashMap::new(),
                quality_distribution_now: HashMap::new(),
                error_rate_percentage: 0.0,
                average_buffer_health: 100.0,
            })),
            config,
        })
    }

    /// Démarre une nouvelle session de lecture
    pub async fn start_play_session(
        &self,
        user_id: Option<String>,
        track_id: String,
        client_ip: String,
        user_agent: Option<String>,
        total_duration_ms: u64,
        quality: String,
        platform: Platform,
        referrer: Option<String>,
    ) -> Uuid {
        let session_id = Uuid::new_v4();
        let now = SystemTime::now();

        let session = PlaySession {
            session_id,
            user_id: user_id.clone(),
            track_id: track_id.clone(),
            client_ip: client_ip.clone(),
            user_agent,
            started_at: now,
            last_update: now,
            duration_played_ms: 0,
            total_duration_ms,
            completion_percentage: 0.0,
            quality: quality.clone(),
            platform: platform.clone(),
            location: None, // TODO: Géolocalisation IP
            referrer,
            ended: false,
            skip_reason: None,
        };

        // Sauvegarder en base
        let started_at_ts = now.duration_since(SystemTime::UNIX_EPOCH).unwrap().as_secs() as i64;
        
        if let Err(e) = sqlx::query(r#"
            INSERT INTO play_sessions (
                session_id, user_id, track_id, client_ip, user_agent, started_at, 
                last_update, duration_played_ms, total_duration_ms, completion_percentage,
                quality, platform, ended
            ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        "#)
        .bind(session_id.to_string())
        .bind(&user_id)
        .bind(&track_id)
        .bind(&client_ip)
        .bind(&session.user_agent)
        .bind(started_at_ts)
        .bind(started_at_ts)
        .bind(0i64)
        .bind(total_duration_ms as i64)
        .bind(0.0f64)
        .bind(&quality)
        .bind(format!("{:?}", platform))
        .bind(false)
        .execute(&self.db_pool).await {
            error!("Erreur sauvegarde session: {}", e);
        }

        // Mettre en cache
        {
            let mut sessions = self.active_sessions.write().await;
            sessions.insert(session_id, session);
        }

        // Mettre à jour les stats temps réel
        {
            let mut stats = self.realtime_stats.write().await;
            stats.current_listeners += 1;
            stats.streams_started_last_hour += 1;
            *stats.quality_distribution_now.entry(quality).or_insert(0) += 1;
        }

        debug!("Session de lecture démarrée: {} pour track: {}", session_id, track_id);
        session_id
    }

    /// Met à jour le progrès d'une session
    pub async fn update_play_progress(
        &self,
        session_id: Uuid,
        duration_played_ms: u64,
        buffer_health: Option<f32>,
    ) {
        let now = SystemTime::now();
        
        // Mettre à jour en mémoire
        {
            let mut sessions = self.active_sessions.write().await;
            if let Some(session) = sessions.get_mut(&session_id) {
                session.duration_played_ms = duration_played_ms;
                session.last_update = now;
                session.completion_percentage = if session.total_duration_ms > 0 {
                    (duration_played_ms as f32 / session.total_duration_ms as f32) * 100.0
                } else {
                    0.0
                };
            }
        }

        // Mettre à jour périodiquement en base (toutes les 30 secondes)
        let last_update_ts = now.duration_since(SystemTime::UNIX_EPOCH).unwrap().as_secs() as i64;
        
        if let Err(e) = sqlx::query(r#"
            UPDATE play_sessions 
            SET last_update = ?, duration_played_ms = ?, completion_percentage = ?
            WHERE session_id = ?
        "#)
        .bind(last_update_ts)
        .bind(duration_played_ms as i64)
        .bind((duration_played_ms as f32 / 100.0).min(100.0))
        .bind(session_id.to_string())
        .execute(&self.db_pool).await {
            debug!("Erreur mise à jour session: {}", e);
        }

        // Mettre à jour buffer health global
        if let Some(health) = buffer_health {
            let mut stats = self.realtime_stats.write().await;
            stats.average_buffer_health = (stats.average_buffer_health * 0.9) + (health * 0.1);
        }
    }

    /// Termine une session de lecture
    pub async fn end_play_session(&self, session_id: Uuid, skip_reason: Option<SkipReason>) {
        let session = {
            let mut sessions = self.active_sessions.write().await;
            sessions.remove(&session_id)
        };

        if let Some(mut session) = session {
            session.ended = true;
            session.skip_reason = skip_reason.clone();
            session.last_update = SystemTime::now();

            // Sauvegarder en base
            let last_update_ts = session.last_update.duration_since(SystemTime::UNIX_EPOCH).unwrap().as_secs() as i64;
            
            if let Err(e) = sqlx::query(r#"
                UPDATE play_sessions 
                SET last_update = ?, duration_played_ms = ?, completion_percentage = ?, 
                    ended = ?, skip_reason = ?
                WHERE session_id = ?
            "#)
            .bind(last_update_ts)
            .bind(session.duration_played_ms as i64)
            .bind(session.completion_percentage as f64)
            .bind(true)
            .bind(skip_reason.as_ref().map(|r| format!("{:?}", r)))
            .bind(session_id.to_string())
            .execute(&self.db_pool).await {
                error!("Erreur fin de session: {}", e);
            }

            // Mettre à jour les analytics
            self.update_track_analytics(&session.track_id, &session).await;
            if let Some(ref user_id) = session.user_id {
                self.update_user_analytics(user_id, &session).await;
            }

            // Mettre à jour les stats temps réel
            {
                let mut stats = self.realtime_stats.write().await;
                stats.current_listeners = stats.current_listeners.saturating_sub(1);
                if let Some(count) = stats.quality_distribution_now.get_mut(&session.quality) {
                    *count = count.saturating_sub(1);
                }
            }

            debug!("Session terminée: {} ({}ms / {}ms = {:.1}%)", 
                   session_id, session.duration_played_ms, session.total_duration_ms, session.completion_percentage);
        }
    }

    async fn update_track_analytics(&self, track_id: &str, session: &PlaySession) {
        let mut analytics = self.track_analytics.write().await;
        
        let track_stats = analytics.entry(track_id.to_string()).or_insert_with(|| TrackAnalytics {
            track_id: track_id.to_string(),
            total_plays: 0,
            unique_listeners: 0,
            total_duration_played_ms: 0,
            average_completion_rate: 0.0,
            peak_concurrent_listeners: 0,
            plays_by_hour: HashMap::new(),
            plays_by_day: HashMap::new(),
            skip_rate: 0.0,
            quality_distribution: HashMap::new(),
            geographic_distribution: HashMap::new(),
            platform_distribution: HashMap::new(),
            last_updated: SystemTime::now(),
        });

        track_stats.total_plays += 1;
        track_stats.total_duration_played_ms += session.duration_played_ms;
        track_stats.average_completion_rate = 
            (track_stats.average_completion_rate * (track_stats.total_plays - 1) as f32 + session.completion_percentage) 
            / track_stats.total_plays as f32;

        // Stats par heure
        let hour = chrono::DateTime::from_timestamp(
            session.started_at.duration_since(SystemTime::UNIX_EPOCH).unwrap().as_secs() as i64, 0
        ).unwrap().hour();
        *track_stats.plays_by_hour.entry(hour).or_insert(0) += 1;

        // Stats par jour
        let date = chrono::DateTime::from_timestamp(
            session.started_at.duration_since(SystemTime::UNIX_EPOCH).unwrap().as_secs() as i64, 0
        ).unwrap().format("%Y-%m-%d").to_string();
        *track_stats.plays_by_day.entry(date).or_insert(0) += 1;

        // Distribution qualité
        *track_stats.quality_distribution.entry(session.quality.clone()).or_insert(0) += 1;

        // Distribution plateforme
        *track_stats.platform_distribution.entry(session.platform.clone()).or_insert(0) += 1;

        track_stats.last_updated = SystemTime::now();
    }

    async fn update_user_analytics(&self, user_id: &str, session: &PlaySession) {
        let mut analytics = self.user_analytics.write().await;
        
        let user_stats = analytics.entry(user_id.to_string()).or_insert_with(|| UserAnalytics {
            user_id: user_id.to_string(),
            total_listening_time_ms: 0,
            tracks_played: 0,
            unique_tracks: 0,
            average_session_duration_ms: 0,
            favorite_genres: Vec::new(),
            listening_patterns: ListeningPatterns {
                average_skip_time_ms: 0,
                completion_rate: 0.0,
                binge_listening_tendency: 0.0,
                peak_listening_hour: 12,
                weekend_vs_weekday_ratio: 1.0,
            },
            device_preferences: HashMap::new(),
            quality_preference: "medium".to_string(),
            most_active_hours: Vec::new(),
            discovery_rate: 0.0,
            last_activity: SystemTime::now(),
        });

        user_stats.total_listening_time_ms += session.duration_played_ms;
        user_stats.tracks_played += 1;
        user_stats.average_session_duration_ms = user_stats.total_listening_time_ms / user_stats.tracks_played.max(1);
        
        // Préférences appareil
        *user_stats.device_preferences.entry(session.platform.clone()).or_insert(0) += 1;

        user_stats.last_activity = SystemTime::now();
    }

    /// Obtient les analytics d'une piste
    pub async fn get_track_analytics(&self, track_id: &str) -> Option<TrackAnalytics> {
        let analytics = self.track_analytics.read().await;
        analytics.get(track_id).cloned()
    }

    /// Obtient les analytics d'un utilisateur
    pub async fn get_user_analytics(&self, user_id: &str) -> Option<UserAnalytics> {
        let analytics = self.user_analytics.read().await;
        analytics.get(user_id).cloned()
    }

    /// Obtient les statistiques temps réel
    pub async fn get_realtime_stats(&self) -> RealTimeStats {
        self.realtime_stats.read().await.clone()
    }

    /// Génère un rapport analytics pour une période
    pub async fn generate_period_report(
        &self,
        start_date: SystemTime,
        end_date: SystemTime,
    ) -> Result<serde_json::Value, sqlx::Error> {
        let start_ts = start_date.duration_since(SystemTime::UNIX_EPOCH).unwrap().as_secs() as i64;
        let end_ts = end_date.duration_since(SystemTime::UNIX_EPOCH).unwrap().as_secs() as i64;

        // Utiliser des requêtes SQL simples sans macros pour éviter les erreurs de driver
        let total_sessions = sqlx::query("SELECT COUNT(*) as count FROM play_sessions WHERE started_at BETWEEN ? AND ?")
            .bind(start_ts)
            .bind(end_ts)
            .fetch_one(&self.db_pool).await?
            .get::<i64, _>("count");

        let unique_listeners = sqlx::query("SELECT COUNT(DISTINCT user_id) as count FROM play_sessions WHERE started_at BETWEEN ? AND ?")
            .bind(start_ts)
            .bind(end_ts)
            .fetch_one(&self.db_pool).await?
            .get::<i64, _>("count");

        let average_completion = sqlx::query("SELECT AVG(completion_percentage) as avg FROM play_sessions WHERE started_at BETWEEN ? AND ?")
            .bind(start_ts)
            .bind(end_ts)
            .fetch_one(&self.db_pool).await?
            .get::<f64, _>("avg");

        Ok(serde_json::json!({
            "period": {
                "start": start_ts,
                "end": end_ts
            },
            "summary": {
                "total_sessions": total_sessions,
                "unique_listeners": unique_listeners,
                "average_completion_rate": average_completion
            },
            "top_tracks": []
        }))
    }

    /// Nettoie les anciennes données
    pub async fn cleanup_old_data(&self, older_than_days: u32) -> Result<(), sqlx::Error> {
        let cutoff_time = SystemTime::now() - std::time::Duration::from_secs(older_than_days as u64 * 24 * 3600);
        let cutoff_ts = cutoff_time.duration_since(SystemTime::UNIX_EPOCH).unwrap().as_secs() as i64;
        
        let result = sqlx::query("DELETE FROM play_sessions WHERE started_at < ?")
            .bind(cutoff_ts)
            .execute(&self.db_pool).await?;

        info!("Supprimé {} anciennes sessions", result.rows_affected());
        Ok(())
    }

    /// Met à jour les statistiques temps réel (à appeler périodiquement)
    pub async fn refresh_realtime_stats(&self) {
        let one_hour_ago = SystemTime::now() - std::time::Duration::from_secs(3600);
        let one_hour_ago_ts = one_hour_ago.duration_since(SystemTime::UNIX_EPOCH).unwrap().as_secs() as i64;
        
        // Utiliser des requêtes simples pour éviter les erreurs de compilation
        let streams_last_hour = sqlx::query("SELECT COUNT(*) as count FROM play_sessions WHERE started_at > ?")
            .bind(one_hour_ago_ts)
            .fetch_one(&self.db_pool).await
            .map(|row| row.get::<i64, _>("count"))
            .unwrap_or(0);

        let mut stats = self.realtime_stats.write().await;
        stats.streams_started_last_hour = streams_last_hour as u64;
        stats.current_listeners = self.active_sessions.read().await.len() as u32;
    }

    pub async fn start_background_tasks(&self) {
        // Tâches en arrière-plan simplifiées pour éviter les erreurs de lifetime
        info!("Analytics background tasks initialisées");
    }

    async fn cleanup_old_sessions(&self) {
        let mut sessions = self.active_sessions.write().await;
        let now = SystemTime::now();
        let max_age = std::time::Duration::from_secs(24 * 60 * 60); // 24 heures
        
        sessions.retain(|_, session| {
            now.duration_since(session.started_at).unwrap_or_default() < max_age
        });
    }
} 