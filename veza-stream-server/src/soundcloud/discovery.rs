/// Module de découverte et recommandations SoundCloud-like
/// 
/// Features :
/// - Algorithmes de recommandation ML
/// - Trending tracks par genre/région
/// - Charts Top 50, New & Hot
/// - Station radio continue
/// - Découverte personnalisée
/// - Analytics d'engagement

use std::sync::Arc;
use std::collections::{HashMap, BTreeMap, VecDeque};
use std::time::{Duration, SystemTime, Instant};

use serde::{Serialize, Deserialize};
use uuid::Uuid;
use tokio::sync::RwLock;
use parking_lot::Mutex;
use tracing::{info, debug, warn};

use crate::error::AppError;
use crate::soundcloud::social::{SocialManager, UserSocialStats};

/// Gestionnaire principal de la découverte
#[derive(Debug)]
pub struct DiscoveryEngine {
    /// Algorithmes de recommandation
    recommendation_engine: Arc<RecommendationEngine>,
    /// Gestionnaire de trending
    trending_manager: Arc<TrendingManager>,
    /// Charts globaux et par genre
    charts_manager: Arc<ChartsManager>,
    /// Stations radio personnalisées
    radio_manager: Arc<RadioManager>,
    /// Analytics d'engagement
    engagement_tracker: Arc<EngagementTracker>,
    /// Configuration
    config: DiscoveryConfig,
}

/// Moteur de recommandations ML
#[derive(Debug)]
pub struct RecommendationEngine {
    /// Données d'écoute utilisateur
    user_listening_history: Arc<RwLock<HashMap<i64, UserListeningProfile>>>,
    /// Similarité entre tracks
    track_similarity_matrix: Arc<RwLock<TrackSimilarityMatrix>>,
    /// Clusters d'utilisateurs similaires
    user_clusters: Arc<RwLock<UserClusters>>,
    /// Modèles ML entraînés
    ml_models: Arc<MLModels>,
    /// Configuration
    config: RecommendationConfig,
}

/// Profil d'écoute d'un utilisateur
#[derive(Debug, Clone)]
pub struct UserListeningProfile {
    pub user_id: i64,
    pub listening_history: VecDeque<ListeningEvent>,
    pub genre_preferences: HashMap<String, GenrePreference>,
    pub artist_preferences: HashMap<String, ArtistPreference>,
    pub tempo_preferences: TempoPreferences,
    pub discovery_preferences: DiscoveryPreferences,
    pub last_updated: SystemTime,
}

/// Événement d'écoute
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ListeningEvent {
    pub track_id: Uuid,
    pub listened_at: SystemTime,
    pub duration_listened: Duration,
    pub completion_percentage: f32,
    pub source: ListeningSource,
    pub skipped: bool,
    pub liked: bool,
    pub reposted: bool,
    pub shared: bool,
}

/// Source d'écoute
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum ListeningSource {
    Search,
    Recommendation,
    Trending,
    Chart,
    Radio,
    Playlist,
    Artist,
    Album,
    Feed,
    External,
}

/// Préférence de genre musical
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct GenrePreference {
    pub genre: String,
    pub weight: f32,
    pub listen_count: u32,
    pub average_completion: f32,
    pub last_listened: SystemTime,
    pub trending: bool,
}

/// Préférence d'artiste
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ArtistPreference {
    pub artist: String,
    pub weight: f32,
    pub tracks_listened: u32,
    pub average_completion: f32,
    pub last_listened: SystemTime,
    pub following: bool,
}

/// Préférences de tempo
#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct TempoPreferences {
    pub preferred_bpm_range: (f32, f32),
    pub energy_level: f32, // 0.0 - 1.0
    pub danceability: f32, // 0.0 - 1.0
    pub valence: f32,      // 0.0 - 1.0 (positivity)
}

/// Préférences de découverte
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DiscoveryPreferences {
    pub familiarity_ratio: f32,    // 0.0 = toujours nouveau, 1.0 = toujours familier
    pub genre_diversity: f32,      // 0.0 - 1.0
    pub popularity_bias: f32,      // 0.0 = underground, 1.0 = mainstream
    pub recency_preference: f32,   // 0.0 = classiques, 1.0 = nouveau
    pub language_preferences: Vec<String>,
    pub explicit_content: bool,
}

/// Matrice de similarité entre tracks
#[derive(Debug, Clone)]
pub struct TrackSimilarityMatrix {
    /// track_id -> Vec<(similar_track_id, similarity_score)>
    similarities: HashMap<Uuid, Vec<TrackSimilarity>>,
    last_computed: SystemTime,
}

impl Default for TrackSimilarityMatrix {
    fn default() -> Self {
        Self {
            similarities: HashMap::new(),
            last_computed: SystemTime::now(),
        }
    }
}

/// Similarité entre deux tracks
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TrackSimilarity {
    pub track_id: Uuid,
    pub similarity_score: f32, // 0.0 - 1.0
    pub similarity_factors: SimilarityFactors,
}

/// Facteurs de similarité
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SimilarityFactors {
    pub genre_similarity: f32,
    pub tempo_similarity: f32,
    pub key_similarity: f32,
    pub energy_similarity: f32,
    pub audio_features_similarity: f32,
    pub collaborative_filtering: f32, // Basé sur les écoutes utilisateur
}

/// Clusters d'utilisateurs similaires
#[derive(Debug, Clone)]
pub struct UserClusters {
    /// cluster_id -> Vec<user_id>
    clusters: HashMap<u32, Vec<i64>>,
    /// user_id -> cluster_id
    user_to_cluster: HashMap<i64, u32>,
    cluster_profiles: HashMap<u32, ClusterProfile>,
    last_computed: SystemTime,
}

impl Default for UserClusters {
    fn default() -> Self {
        Self {
            clusters: HashMap::new(),
            user_to_cluster: HashMap::new(),
            cluster_profiles: HashMap::new(),
            last_computed: SystemTime::now(),
        }
    }
}

/// Profil d'un cluster d'utilisateurs
#[derive(Debug, Clone)]
pub struct ClusterProfile {
    pub cluster_id: u32,
    pub size: usize,
    pub dominant_genres: Vec<String>,
    pub average_age_range: Option<(u8, u8)>,
    pub geographic_regions: Vec<String>,
    pub listening_patterns: ListeningPatterns,
    pub discovery_behavior: DiscoveryBehavior,
}

/// Patterns d'écoute d'un cluster
#[derive(Debug, Clone, Default)]
pub struct ListeningPatterns {
    pub peak_listening_hours: Vec<u8>, // 0-23
    pub average_session_duration: Duration,
    pub skip_rate: f32,
    pub completion_rate: f32,
    pub playlist_usage: f32,
    pub social_engagement: f32,
}

/// Comportement de découverte d'un cluster
#[derive(Debug, Clone, Default)]
pub struct DiscoveryBehavior {
    pub openness_to_new: f32,
    pub genre_exploration: f32,
    pub trending_influence: f32,
    pub social_influence: f32,
    pub recommendation_acceptance: f32,
}

/// Modèles ML pour recommandations
#[derive(Debug)]
pub struct MLModels {
    /// Modèle de collaborative filtering
    collaborative_model: Arc<Mutex<CollaborativeFilteringModel>>,
    /// Modèle de content-based filtering
    content_model: Arc<Mutex<ContentBasedModel>>,
    /// Modèle hybride
    hybrid_model: Arc<Mutex<HybridModel>>,
    /// Modèle de tendances
    trending_model: Arc<Mutex<TrendingModel>>,
}

/// Modèle de collaborative filtering (simulation)
#[derive(Debug, Default)]
pub struct CollaborativeFilteringModel {
    /// Matrice utilisateur-item
    user_item_matrix: HashMap<(i64, Uuid), f32>,
    /// Facteurs latents utilisateur
    user_factors: HashMap<i64, Vec<f32>>,
    /// Facteurs latents tracks
    item_factors: HashMap<Uuid, Vec<f32>>,
    model_accuracy: f32,
    last_trained: Option<SystemTime>,
}

/// Modèle content-based (simulation)
#[derive(Debug, Default)]
pub struct ContentBasedModel {
    /// Features audio par track
    track_features: HashMap<Uuid, AudioFeatures>,
    /// Poids des features
    feature_weights: Vec<f32>,
    model_accuracy: f32,
    last_trained: Option<SystemTime>,
}

/// Features audio d'une track
#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct AudioFeatures {
    pub tempo: f32,
    pub key: i8,
    pub energy: f32,
    pub danceability: f32,
    pub valence: f32,
    pub acousticness: f32,
    pub instrumentalness: f32,
    pub speechiness: f32,
    pub loudness: f32,
    pub duration_ms: u32,
    pub time_signature: u8,
}

/// Modèle hybride combinant collaborative et content-based
#[derive(Debug, Default)]
pub struct HybridModel {
    collaborative_weight: f32,
    content_weight: f32,
    popularity_weight: f32,
    recency_weight: f32,
    social_weight: f32,
    model_accuracy: f32,
    last_trained: Option<SystemTime>,
}

/// Modèle de tendances
#[derive(Debug, Default)]
pub struct TrendingModel {
    /// Scores de tendance par track
    trending_scores: HashMap<Uuid, TrendingScore>,
    /// Decay factor pour l'ancienneté
    time_decay_factor: f32,
    /// Weights pour différents signaux
    play_weight: f32,
    like_weight: f32,
    share_weight: f32,
    comment_weight: f32,
    velocity_weight: f32,
}

/// Score de tendance d'une track
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TrendingScore {
    pub track_id: Uuid,
    pub score: f32,
    pub plays_velocity: f32,    // Plays par heure
    pub likes_velocity: f32,    // Likes par heure  
    pub shares_velocity: f32,   // Shares par heure
    pub comments_velocity: f32, // Comments par heure
    pub geographic_spread: f32, // Dispersion géographique
    pub last_updated: SystemTime,
}

/// Gestionnaire de trending
#[derive(Debug)]
pub struct TrendingManager {
    /// Trending global
    global_trending: Arc<RwLock<Vec<TrendingTrack>>>,
    /// Trending par genre
    genre_trending: Arc<RwLock<HashMap<String, Vec<TrendingTrack>>>>,
    /// Trending par région
    regional_trending: Arc<RwLock<HashMap<String, Vec<TrendingTrack>>>>,
    /// Configuration
    config: TrendingConfig,
}

/// Track dans le trending
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TrendingTrack {
    pub track_id: Uuid,
    pub position: u32,
    pub previous_position: Option<u32>,
    pub trend_direction: TrendDirection,
    pub trending_score: f32,
    pub velocity_score: f32,
    pub time_in_trending: Duration,
    pub peak_position: u32,
}

/// Direction de la tendance
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum TrendDirection {
    Up(u32),    // Positions gagnées
    Down(u32),  // Positions perdues
    Stable,
    New,        // Nouvelle entry
}

/// Configuration du trending
#[derive(Debug, Clone)]
pub struct TrendingConfig {
    pub update_interval: Duration,
    pub trending_window: Duration, // Fenêtre de temps pour calcul
    pub max_trending_items: usize,
    pub min_plays_threshold: u32,
    pub geographic_regions: Vec<String>,
    pub decay_factor: f32,
}

/// Gestionnaire de charts
#[derive(Debug)]
pub struct ChartsManager {
    /// Chart global Top 50
    global_chart: Arc<RwLock<Chart>>,
    /// Charts par genre
    genre_charts: Arc<RwLock<HashMap<String, Chart>>>,
    /// Chart "New & Hot"
    new_hot_chart: Arc<RwLock<Chart>>,
    /// Chart découvertes de la semaine
    weekly_discovery_chart: Arc<RwLock<Chart>>,
    /// Configuration
    config: ChartsConfig,
}

/// Chart musical
#[derive(Debug, Clone)]
pub struct Chart {
    pub chart_type: ChartType,
    pub period: ChartPeriod,
    pub entries: Vec<ChartEntry>,
    pub last_updated: SystemTime,
    pub total_entries: usize,
}

/// Type de chart
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum ChartType {
    Global,
    Genre(String),
    NewHot,
    WeeklyDiscovery,
    Regional(String),
}

/// Période du chart
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum ChartPeriod {
    Daily,
    Weekly,
    Monthly,
    AllTime,
}

/// Entrée dans un chart
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ChartEntry {
    pub position: u32,
    pub previous_position: Option<u32>,
    pub track_id: Uuid,
    pub chart_score: f32,
    pub plays_count: u64,
    pub weeks_on_chart: u32,
    pub peak_position: u32,
    pub trend: TrendDirection,
}

/// Configuration des charts
#[derive(Debug, Clone)]
pub struct ChartsConfig {
    pub update_interval: Duration,
    pub chart_size: usize,
    pub min_chart_threshold: u32,
    pub supported_genres: Vec<String>,
    pub new_track_window: Duration, // Pour "New & Hot"
}

/// Gestionnaire de stations radio
#[derive(Debug)]
pub struct RadioManager {
    /// Stations actives
    active_stations: Arc<RwLock<HashMap<Uuid, RadioStation>>>,
    /// Stations par utilisateur
    user_stations: Arc<RwLock<HashMap<i64, Vec<Uuid>>>>,
    /// Configuration
    config: RadioConfig,
}

/// Station radio personnalisée
#[derive(Debug, Clone)]
pub struct RadioStation {
    pub id: Uuid,
    pub user_id: i64,
    pub station_type: RadioStationType,
    pub name: String,
    pub description: Option<String>,
    pub seed_tracks: Vec<Uuid>,
    pub generated_queue: VecDeque<Uuid>,
    pub played_tracks: Vec<Uuid>,
    pub current_position: usize,
    pub last_updated: SystemTime,
    pub total_listening_time: Duration,
}

/// Type de station radio
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum RadioStationType {
    /// Basée sur des tracks seed
    TrackSeed(Vec<Uuid>),
    /// Basée sur des artistes
    ArtistSeed(Vec<String>),
    /// Basée sur un genre
    GenreSeed(String),
    /// Découverte personnalisée
    PersonalizedDiscovery,
    /// Trending mix
    TrendingMix,
    /// Deep cuts (tracks moins connues)
    DeepCuts,
    /// Focus genre avec évolution
    GenreEvolution(String),
}

/// Configuration des stations radio
#[derive(Debug, Clone)]
pub struct RadioConfig {
    pub max_stations_per_user: usize,
    pub queue_size: usize,
    pub similarity_threshold: f32,
    pub diversity_factor: f32,
    pub discovery_ratio: f32, // Ratio de tracks inconnues
}

/// Tracker d'engagement
#[derive(Debug)]
pub struct EngagementTracker {
    /// Métriques d'engagement par recommandation
    recommendation_metrics: Arc<RwLock<HashMap<Uuid, RecommendationMetrics>>>,
    /// Feedback utilisateur
    user_feedback: Arc<RwLock<HashMap<i64, Vec<UserFeedback>>>>,
    /// A/B tests actifs
    ab_tests: Arc<RwLock<HashMap<String, ABTest>>>,
}

/// Métriques d'une recommandation
#[derive(Debug, Clone, Default)]
pub struct RecommendationMetrics {
    pub recommendation_id: Uuid,
    pub user_id: i64,
    pub track_id: Uuid,
    pub algorithm_used: String,
    pub confidence_score: f32,
    pub clicked: bool,
    pub played: bool,
    pub completion_rate: f32,
    pub liked: bool,
    pub shared: bool,
    pub feedback_score: Option<f32>, // 1-5 stars
    pub timestamp: SystemTime,
}

/// Feedback utilisateur
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct UserFeedback {
    pub feedback_id: Uuid,
    pub user_id: i64,
    pub track_id: Uuid,
    pub feedback_type: FeedbackType,
    pub score: Option<f32>,
    pub comment: Option<String>,
    pub timestamp: SystemTime,
}

/// Type de feedback
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum FeedbackType {
    Like,
    Dislike,
    NotInterested,
    PlayedOften,
    Shared,
    AddedToPlaylist,
    Rating(f32), // 1-5 stars
    TextFeedback(String),
}

/// Test A/B pour algorithmes
#[derive(Debug, Clone)]
pub struct ABTest {
    pub test_id: String,
    pub test_name: String,
    pub algorithm_a: String,
    pub algorithm_b: String,
    pub user_assignments: HashMap<i64, String>, // user_id -> algorithm
    pub metrics_a: ABTestMetrics,
    pub metrics_b: ABTestMetrics,
    pub start_date: SystemTime,
    pub end_date: SystemTime,
    pub active: bool,
}

/// Métriques d'un test A/B
#[derive(Debug, Clone, Default)]
pub struct ABTestMetrics {
    pub total_users: u32,
    pub total_recommendations: u32,
    pub click_through_rate: f32,
    pub play_rate: f32,
    pub completion_rate: f32,
    pub like_rate: f32,
    pub share_rate: f32,
    pub average_rating: f32,
}

/// Configuration du discovery
#[derive(Debug, Clone)]
pub struct DiscoveryConfig {
    pub recommendation_config: RecommendationConfig,
    pub trending_config: TrendingConfig,
    pub charts_config: ChartsConfig,
    pub radio_config: RadioConfig,
    pub max_recommendations_per_request: usize,
    pub enable_ab_testing: bool,
    pub ml_model_retrain_interval: Duration,
}

/// Configuration des recommandations
#[derive(Debug, Clone)]
pub struct RecommendationConfig {
    pub max_similar_tracks: usize,
    pub similarity_threshold: f32,
    pub diversity_factor: f32,
    pub popularity_boost: f32,
    pub recency_boost: f32,
    pub social_boost: f32,
    pub cold_start_fallback: bool,
    pub explicit_content_filter: bool,
}

impl Default for DiscoveryConfig {
    fn default() -> Self {
        Self {
            recommendation_config: RecommendationConfig::default(),
            trending_config: TrendingConfig::default(),
            charts_config: ChartsConfig::default(),
            radio_config: RadioConfig::default(),
            max_recommendations_per_request: 50,
            enable_ab_testing: true,
            ml_model_retrain_interval: Duration::from_secs(3600), // 1 heure
        }
    }
}

impl Default for RecommendationConfig {
    fn default() -> Self {
        Self {
            max_similar_tracks: 100,
            similarity_threshold: 0.5,
            diversity_factor: 0.3,
            popularity_boost: 0.1,
            recency_boost: 0.15,
            social_boost: 0.2,
            cold_start_fallback: true,
            explicit_content_filter: false,
        }
    }
}

impl Default for TrendingConfig {
    fn default() -> Self {
        Self {
            update_interval: Duration::from_secs(300), // 5 minutes
            trending_window: Duration::from_secs(86400), // 24 heures
            max_trending_items: 50,
            min_plays_threshold: 100,
            geographic_regions: vec![
                "US".to_string(),
                "UK".to_string(), 
                "DE".to_string(),
                "FR".to_string(),
                "JP".to_string(),
            ],
            decay_factor: 0.95,
        }
    }
}

impl Default for ChartsConfig {
    fn default() -> Self {
        Self {
            update_interval: Duration::from_secs(3600), // 1 heure
            chart_size: 50,
            min_chart_threshold: 1000,
            supported_genres: vec![
                "Electronic".to_string(),
                "Hip Hop".to_string(),
                "Rock".to_string(),
                "Pop".to_string(),
                "Jazz".to_string(),
                "Classical".to_string(),
                "R&B".to_string(),
                "Country".to_string(),
                "Reggae".to_string(),
                "Blues".to_string(),
            ],
            new_track_window: Duration::from_secs(604800), // 1 semaine
        }
    }
}

impl Default for RadioConfig {
    fn default() -> Self {
        Self {
            max_stations_per_user: 10,
            queue_size: 100,
            similarity_threshold: 0.6,
            diversity_factor: 0.25,
            discovery_ratio: 0.3,
        }
    }
}

impl DiscoveryEngine {
    /// Crée un nouveau moteur de découverte
    pub async fn new(
        config: DiscoveryConfig,
        social_manager: Arc<SocialManager>,
    ) -> Result<Self, AppError> {
        let recommendation_engine = Arc::new(RecommendationEngine::new(config.recommendation_config.clone()).await?);
        let trending_manager = Arc::new(TrendingManager::new(config.trending_config.clone()));
        let charts_manager = Arc::new(ChartsManager::new(config.charts_config.clone()));
        let radio_manager = Arc::new(RadioManager::new(config.radio_config.clone()));
        let engagement_tracker = Arc::new(EngagementTracker::new());
        
        Ok(Self {
            recommendation_engine,
            trending_manager,
            charts_manager,
            radio_manager,
            engagement_tracker,
            config,
        })
    }
    
    /// Obtient des recommandations personnalisées pour un utilisateur
    pub async fn get_personalized_recommendations(
        &self,
        user_id: i64,
        count: usize,
        seed_tracks: Option<Vec<Uuid>>,
    ) -> Result<Vec<RecommendationResult>, AppError> {
        let count = count.min(self.config.max_recommendations_per_request);
        
        // Obtenir le profil utilisateur
        let user_profile = self.recommendation_engine
            .get_user_profile(user_id)
            .await
            .unwrap_or_else(|| UserListeningProfile::new(user_id));
        
        // Générer les recommandations selon différents algorithmes
        let mut recommendations = Vec::new();
        
        // 60% collaborative filtering
        let collaborative_count = (count as f32 * 0.6) as usize;
        let mut collaborative = self.recommendation_engine
            .get_collaborative_recommendations(user_id, collaborative_count)
            .await?;
        recommendations.append(&mut collaborative);
        
        // 30% content-based
        let content_count = (count as f32 * 0.3) as usize;
        let mut content_based = self.recommendation_engine
            .get_content_based_recommendations(user_id, content_count, seed_tracks.clone())
            .await?;
        recommendations.append(&mut content_based);
        
        // 10% trending/social
        let trending_count = count - recommendations.len();
        let mut trending = self.get_trending_recommendations(user_id, trending_count).await?;
        recommendations.append(&mut trending);
        
        // Diversifier et re-scorer
        let final_recommendations = self.diversify_and_score_recommendations(
            recommendations,
            &user_profile,
            count,
        ).await?;
        
        // Tracker les recommandations pour analytics
        self.track_recommendations(&final_recommendations, user_id).await?;
        
        Ok(final_recommendations)
    }
    
    /// Obtient les tracks trending
    pub async fn get_trending_tracks(
        &self,
        genre: Option<String>,
        region: Option<String>,
        limit: Option<usize>,
    ) -> Result<Vec<TrendingTrack>, AppError> {
        self.trending_manager.get_trending(genre, region, limit).await
    }
    
    /// Obtient un chart spécifique
    pub async fn get_chart(
        &self,
        chart_type: ChartType,
        period: ChartPeriod,
        limit: Option<usize>,
    ) -> Result<Chart, AppError> {
        self.charts_manager.get_chart(chart_type, period, limit).await
    }
    
    /// Crée une station radio personnalisée
    pub async fn create_radio_station(
        &self,
        user_id: i64,
        station_type: RadioStationType,
        name: String,
    ) -> Result<Uuid, AppError> {
        self.radio_manager.create_station(user_id, station_type, name).await
    }
    
    /// Obtient les recommendations trending
    async fn get_trending_recommendations(
        &self,
        user_id: i64,
        count: usize,
    ) -> Result<Vec<RecommendationResult>, AppError> {
        // Simulation - obtenir les tracks trending qui correspondent au profil utilisateur
        let trending = self.trending_manager.get_trending(None, None, Some(count * 2)).await?;
        
        let mut recommendations = Vec::new();
        for (i, track) in trending.iter().enumerate().take(count) {
            recommendations.push(RecommendationResult {
                track_id: track.track_id,
                confidence_score: 0.8 - (i as f32 * 0.01), // Score décroissant
                reason: RecommendationReason::Trending,
                algorithm_used: "trending".to_string(),
                metadata: None,
            });
        }
        
        Ok(recommendations)
    }
    
    /// Diversifie et score les recommandations
    async fn diversify_and_score_recommendations(
        &self,
        mut recommendations: Vec<RecommendationResult>,
        user_profile: &UserListeningProfile,
        target_count: usize,
    ) -> Result<Vec<RecommendationResult>, AppError> {
        // Supprimer les doublons
        recommendations.sort_by_key(|r| r.track_id);
        recommendations.dedup_by_key(|r| r.track_id);
        
        // Re-scorer selon le profil utilisateur
        for rec in &mut recommendations {
            rec.confidence_score = self.calculate_personalized_score(rec, user_profile).await;
        }
        
        // Trier par score
        recommendations.sort_by(|a, b| b.confidence_score.partial_cmp(&a.confidence_score).unwrap());
        
        // Diversifier par genre
        let diversified = self.diversify_by_genre(recommendations, target_count).await;
        
        Ok(diversified)
    }
    
    /// Calcule un score personnalisé
    async fn calculate_personalized_score(
        &self,
        recommendation: &RecommendationResult,
        user_profile: &UserListeningProfile,
    ) -> f32 {
        // Simulation de scoring personnalisé
        let mut score = recommendation.confidence_score;
        
        // Ajuster selon les préférences de découverte
        match recommendation.reason {
            RecommendationReason::Trending => {
                score *= user_profile.discovery_preferences.popularity_bias;
            }
            RecommendationReason::SimilarToLiked => {
                score *= (1.0 - user_profile.discovery_preferences.familiarity_ratio);
            }
            _ => {}
        }
        
        score.clamp(0.0, 1.0)
    }
    
    /// Diversifie par genre
    async fn diversify_by_genre(
        &self,
        recommendations: Vec<RecommendationResult>,
        target_count: usize,
    ) -> Vec<RecommendationResult> {
        // Simulation de diversification par genre
        recommendations.into_iter().take(target_count).collect()
    }
    
    /// Tracker les recommandations pour analytics
    async fn track_recommendations(
        &self,
        recommendations: &[RecommendationResult],
        user_id: i64,
    ) -> Result<(), AppError> {
        for rec in recommendations {
            let metrics = RecommendationMetrics {
                recommendation_id: Uuid::new_v4(),
                user_id,
                track_id: rec.track_id,
                algorithm_used: rec.algorithm_used.clone(),
                confidence_score: rec.confidence_score,
                clicked: false,
                played: false,
                completion_rate: 0.0,
                liked: false,
                shared: false,
                feedback_score: None,
                timestamp: SystemTime::now(),
            };
            
            self.engagement_tracker.track_recommendation(metrics).await;
        }
        Ok(())
    }
}

/// Résultat d'une recommandation
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RecommendationResult {
    pub track_id: Uuid,
    pub confidence_score: f32,
    pub reason: RecommendationReason,
    pub algorithm_used: String,
    pub metadata: Option<HashMap<String, serde_json::Value>>,
}

/// Raison de la recommandation
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum RecommendationReason {
    SimilarToLiked,
    PopularInGenre,
    FriendsAlsoLike,
    Trending,
    BasedOnHistory,
    NewRelease,
    DeepCut,
    PersonalizedRadio,
}

impl UserListeningProfile {
    fn new(user_id: i64) -> Self {
        Self {
            user_id,
            listening_history: VecDeque::new(),
            genre_preferences: HashMap::new(),
            artist_preferences: HashMap::new(),
            tempo_preferences: TempoPreferences::default(),
            discovery_preferences: DiscoveryPreferences {
                familiarity_ratio: 0.7,
                genre_diversity: 0.5,
                popularity_bias: 0.6,
                recency_preference: 0.4,
                language_preferences: vec!["en".to_string()],
                explicit_content: false,
            },
            last_updated: SystemTime::now(),
        }
    }
}

impl RecommendationEngine {
    async fn new(config: RecommendationConfig) -> Result<Self, AppError> {
        Ok(Self {
            user_listening_history: Arc::new(RwLock::new(HashMap::new())),
            track_similarity_matrix: Arc::new(RwLock::new(TrackSimilarityMatrix::default())),
            user_clusters: Arc::new(RwLock::new(UserClusters::default())),
            ml_models: Arc::new(MLModels::new()),
            config,
        })
    }
    
    async fn get_user_profile(&self, user_id: i64) -> Option<UserListeningProfile> {
        let profiles = self.user_listening_history.read().await;
        profiles.get(&user_id).cloned()
    }
    
    async fn get_collaborative_recommendations(
        &self,
        user_id: i64,
        count: usize,
    ) -> Result<Vec<RecommendationResult>, AppError> {
        // Simulation collaborative filtering
        let mut recommendations = Vec::new();
        
        for i in 0..count {
            recommendations.push(RecommendationResult {
                track_id: Uuid::new_v4(),
                confidence_score: 0.9 - (i as f32 * 0.05),
                reason: RecommendationReason::SimilarToLiked,
                algorithm_used: "collaborative_filtering".to_string(),
                metadata: None,
            });
        }
        
        Ok(recommendations)
    }
    
    async fn get_content_based_recommendations(
        &self,
        user_id: i64,
        count: usize,
        seed_tracks: Option<Vec<Uuid>>,
    ) -> Result<Vec<RecommendationResult>, AppError> {
        // Simulation content-based filtering
        let mut recommendations = Vec::new();
        
        for i in 0..count {
            recommendations.push(RecommendationResult {
                track_id: Uuid::new_v4(),
                confidence_score: 0.85 - (i as f32 * 0.03),
                reason: RecommendationReason::BasedOnHistory,
                algorithm_used: "content_based".to_string(),
                metadata: None,
            });
        }
        
        Ok(recommendations)
    }
}

impl MLModels {
    fn new() -> Self {
        Self {
            collaborative_model: Arc::new(Mutex::new(CollaborativeFilteringModel::default())),
            content_model: Arc::new(Mutex::new(ContentBasedModel::default())),
            hybrid_model: Arc::new(Mutex::new(HybridModel::default())),
            trending_model: Arc::new(Mutex::new(TrendingModel::default())),
        }
    }
}

impl TrendingManager {
    fn new(config: TrendingConfig) -> Self {
        Self {
            global_trending: Arc::new(RwLock::new(Vec::new())),
            genre_trending: Arc::new(RwLock::new(HashMap::new())),
            regional_trending: Arc::new(RwLock::new(HashMap::new())),
            config,
        }
    }
    
    async fn get_trending(
        &self,
        genre: Option<String>,
        region: Option<String>,
        limit: Option<usize>,
    ) -> Result<Vec<TrendingTrack>, AppError> {
        let limit = limit.unwrap_or(self.config.max_trending_items);
        
        let trending = match (genre, region) {
            (Some(g), _) => {
                let genre_trending = self.genre_trending.read().await;
                genre_trending.get(&g).cloned().unwrap_or_default()
            }
            (None, Some(r)) => {
                let regional_trending = self.regional_trending.read().await;
                regional_trending.get(&r).cloned().unwrap_or_default()
            }
            (None, None) => {
                let global_trending = self.global_trending.read().await;
                global_trending.clone()
            }
        };
        
        Ok(trending.into_iter().take(limit).collect())
    }
}

impl ChartsManager {
    fn new(config: ChartsConfig) -> Self {
        Self {
            global_chart: Arc::new(RwLock::new(Chart {
                chart_type: ChartType::Global,
                period: ChartPeriod::Weekly,
                entries: Vec::new(),
                last_updated: SystemTime::now(),
                total_entries: 0,
            })),
            genre_charts: Arc::new(RwLock::new(HashMap::new())),
            new_hot_chart: Arc::new(RwLock::new(Chart {
                chart_type: ChartType::NewHot,
                period: ChartPeriod::Weekly,
                entries: Vec::new(),
                last_updated: SystemTime::now(),
                total_entries: 0,
            })),
            weekly_discovery_chart: Arc::new(RwLock::new(Chart {
                chart_type: ChartType::WeeklyDiscovery,
                period: ChartPeriod::Weekly,
                entries: Vec::new(),
                last_updated: SystemTime::now(),
                total_entries: 0,
            })),
            config,
        }
    }
    
    async fn get_chart(
        &self,
        chart_type: ChartType,
        period: ChartPeriod,
        limit: Option<usize>,
    ) -> Result<Chart, AppError> {
        let limit = limit.unwrap_or(self.config.chart_size);
        
        let mut chart = match chart_type {
            ChartType::Global => self.global_chart.read().await.clone(),
            ChartType::NewHot => self.new_hot_chart.read().await.clone(),
            ChartType::WeeklyDiscovery => self.weekly_discovery_chart.read().await.clone(),
            ChartType::Genre(genre) => {
                let genre_charts = self.genre_charts.read().await;
                genre_charts.get(&genre).cloned().unwrap_or_else(|| Chart {
                    chart_type: ChartType::Genre(genre),
                    period,
                    entries: Vec::new(),
                    last_updated: SystemTime::now(),
                    total_entries: 0,
                })
            }
            ChartType::Regional(region) => Chart {
                chart_type: ChartType::Regional(region),
                period,
                entries: Vec::new(),
                last_updated: SystemTime::now(),
                total_entries: 0,
            }
        };
        
        chart.entries.truncate(limit);
        Ok(chart)
    }
}

impl RadioManager {
    fn new(config: RadioConfig) -> Self {
        Self {
            active_stations: Arc::new(RwLock::new(HashMap::new())),
            user_stations: Arc::new(RwLock::new(HashMap::new())),
            config,
        }
    }
    
    async fn create_station(
        &self,
        user_id: i64,
        station_type: RadioStationType,
        name: String,
    ) -> Result<Uuid, AppError> {
        // Vérifier la limite par utilisateur
        {
            let user_stations = self.user_stations.read().await;
            if let Some(stations) = user_stations.get(&user_id) {
                if stations.len() >= self.config.max_stations_per_user {
                    return Err(AppError::ValidationError(format!(
                        "Max stations limit reached: {}",
                        self.config.max_stations_per_user
                    )));
                }
            }
        }
        
        let station_id = Uuid::new_v4();
        let station = RadioStation {
            id: station_id,
            user_id,
            station_type,
            name,
            description: None,
            seed_tracks: Vec::new(),
            generated_queue: VecDeque::new(),
            played_tracks: Vec::new(),
            current_position: 0,
            last_updated: SystemTime::now(),
            total_listening_time: Duration::ZERO,
        };
        
        // Ajouter la station
        {
            let mut active_stations = self.active_stations.write().await;
            active_stations.insert(station_id, station);
        }
        
        // Ajouter à l'utilisateur
        {
            let mut user_stations = self.user_stations.write().await;
            user_stations.entry(user_id).or_insert_with(Vec::new).push(station_id);
        }
        
        info!("Station radio créée: {} pour utilisateur {}", station_id, user_id);
        Ok(station_id)
    }
}

impl EngagementTracker {
    fn new() -> Self {
        Self {
            recommendation_metrics: Arc::new(RwLock::new(HashMap::new())),
            user_feedback: Arc::new(RwLock::new(HashMap::new())),
            ab_tests: Arc::new(RwLock::new(HashMap::new())),
        }
    }
    
    async fn track_recommendation(&self, metrics: RecommendationMetrics) {
        let mut recommendation_metrics = self.recommendation_metrics.write().await;
        recommendation_metrics.insert(metrics.recommendation_id, metrics);
    }
} 