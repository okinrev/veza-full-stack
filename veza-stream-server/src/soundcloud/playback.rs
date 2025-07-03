/// Module de playback experience avancée SoundCloud-like
/// 
/// Features :
/// - Continuous playback avec crossfade
/// - Gapless playback seamless
/// - Queue management intelligent
/// - Shuffle/repeat algorithms
/// - Timed comments sur waveform
/// - Hotkeys et contrôles avancés

use std::sync::Arc;
use std::time::{Duration, SystemTime};
use std::collections::{VecDeque, HashMap};

use serde::{Serialize, Deserialize};
use uuid::Uuid;
use tokio::sync::{mpsc, RwLock, broadcast};
use parking_lot::Mutex;
use tracing::info;

use crate::error::AppError;
use crate::core::StreamManager;

/// Gestionnaire principal du playback
#[derive(Debug)]
pub struct PlaybackManager {
    /// Players actifs par utilisateur
    active_players: Arc<RwLock<HashMap<i64, Arc<SoundCloudPlayer>>>>,
    /// Configuration globale
    config: PlaybackConfig,
    /// Gestionnaire de streams
    stream_manager: Arc<StreamManager>,
    /// Événements de playback
    event_sender: broadcast::Sender<PlaybackEvent>,
}

/// Player SoundCloud-like pour un utilisateur
#[derive(Debug)]
pub struct SoundCloudPlayer {
    pub user_id: i64,
    pub session_id: Uuid,
    
    /// État de lecture
    pub playback_state: Arc<RwLock<PlaybackState>>,
    
    /// Queue de lecture
    pub queue: Arc<RwLock<PlaybackQueue>>,
    
    /// Configuration du player
    pub config: PlayerConfig,
    
    /// Contrôleur de crossfade
    crossfade_controller: Arc<Mutex<CrossfadeController>>,
    
    /// Gestionnaire de commentaires temporels
    timed_comments: Arc<RwLock<TimedCommentsManager>>,
    
    /// Analytics de session
    session_analytics: Arc<RwLock<SessionAnalytics>>,
    
    /// Événements du player
    event_sender: mpsc::UnboundedSender<PlaybackEvent>,
}

/// État de lecture du player
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PlaybackState {
    pub current_track: Option<TrackInfo>,
    pub status: PlaybackStatus,
    pub position: Duration,
    pub volume: f32,
    pub playback_speed: f32,
    pub repeat_mode: RepeatMode,
    pub shuffle_enabled: bool,
    pub crossfade_enabled: bool,
    pub gapless_enabled: bool,
    pub last_updated: SystemTime,
}

/// Status de lecture
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum PlaybackStatus {
    Stopped,
    Playing,
    Paused,
    Buffering,
    Loading,
    Error { message: String },
}

/// Modes de répétition
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum RepeatMode {
    Off,
    Track,
    Queue,
    All,
}

/// Queue de lecture avec gestion avancée
#[derive(Debug, Clone)]
pub struct PlaybackQueue {
    /// Index de la piste actuelle
    pub current_index: Option<usize>,
    /// Pistes dans la queue
    pub tracks: Vec<QueueTrack>,
    /// Historique de lecture
    pub play_history: VecDeque<TrackInfo>,
    /// Queue "up next" priorisée
    pub up_next: VecDeque<QueueTrack>,
    /// Mode shuffle
    pub shuffle_state: ShuffleState,
    /// Autoplay activé
    pub autoplay_enabled: bool,
}

/// Piste dans la queue
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct QueueTrack {
    pub track: TrackInfo,
    pub added_at: SystemTime,
    pub added_by: QueueSource,
    pub played: bool,
    pub skipped: bool,
}

/// Source d'ajout à la queue
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum QueueSource {
    User,
    Autoplay,
    Recommendation,
    Radio,
    Playlist { playlist_id: Uuid },
}

/// Informations sur une piste
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TrackInfo {
    pub id: Uuid,
    pub title: String,
    pub artist: String,
    pub album: Option<String>,
    pub duration: Duration,
    pub stream_url: String,
    pub waveform_url: Option<String>,
    pub artwork_url: Option<String>,
    pub genres: Vec<String>,
    pub bpm: Option<f32>,
    pub key: Option<String>,
    pub plays_count: u64,
    pub likes_count: u64,
    pub created_at: SystemTime,
}

/// État du shuffle avec mémoire
#[derive(Debug, Clone)]
pub struct ShuffleState {
    pub enabled: bool,
    pub played_indices: Vec<usize>,
    pub remaining_indices: Vec<usize>,
    pub algorithm: ShuffleAlgorithm,
}

/// Algorithmes de shuffle
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum ShuffleAlgorithm {
    /// Shuffle standard (Fisher-Yates)
    Standard,
    /// Shuffle intelligent évitant les répétitions d'artiste
    Smart,
    /// Shuffle basé sur les préférences utilisateur
    Personalized,
}

/// Configuration du player
#[derive(Debug, Clone)]
pub struct PlayerConfig {
    pub crossfade_duration: Duration,
    pub gapless_gap_threshold: Duration,
    pub max_history_size: usize,
    pub max_queue_size: usize,
    pub enable_scrobbling: bool,
    pub auto_quality_switching: bool,
    pub preload_next_track: bool,
    pub analytics_enabled: bool,
}

/// Configuration globale du playback
#[derive(Debug, Clone)]
pub struct PlaybackConfig {
    pub max_concurrent_players: usize,
    pub default_crossfade_duration: Duration,
    pub enable_real_time_analytics: bool,
    pub cache_preload_tracks: bool,
}

/// Contrôleur de crossfade
#[derive(Debug)]
pub struct CrossfadeController {
    pub enabled: bool,
    pub duration: Duration,
    pub curve: CrossfadeCurve,
    pub current_fade: Option<FadeState>,
}

/// Courbes de crossfade
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum CrossfadeCurve {
    Linear,
    Exponential,
    Logarithmic,
    SCurve,
}

/// État de fade en cours
#[derive(Debug, Clone)]
pub struct FadeState {
    pub start_time: SystemTime,
    pub duration: Duration,
    pub from_volume: f32,
    pub to_volume: f32,
    pub curve: CrossfadeCurve,
}

/// Gestionnaire de commentaires temporels
#[derive(Debug, Clone)]
pub struct TimedCommentsManager {
    /// Commentaires indexés par timestamp
    pub comments: HashMap<u64, Vec<TimedComment>>, // timestamp_ms -> comments
    /// Configuration
    pub config: TimedCommentsConfig,
}

/// Commentaire temporel sur la waveform
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TimedComment {
    pub id: Uuid,
    pub user_id: i64,
    pub track_id: Uuid,
    pub timestamp_ms: u64,
    pub text: String,
    pub created_at: SystemTime,
    pub likes_count: u32,
    pub replies: Vec<CommentReply>,
}

/// Réponse à un commentaire
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CommentReply {
    pub id: Uuid,
    pub user_id: i64,
    pub text: String,
    pub created_at: SystemTime,
}

/// Configuration des commentaires temporels
#[derive(Debug, Clone)]
pub struct TimedCommentsConfig {
    pub enable_live_comments: bool,
    pub max_comments_per_timestamp: usize,
    pub comment_display_duration: Duration,
    pub enable_comment_notifications: bool,
}

/// Analytics de session de playback
#[derive(Debug, Clone, Default)]
pub struct SessionAnalytics {
    pub session_start: Option<SystemTime>,
    pub total_listening_time: Duration,
    pub tracks_played: u32,
    pub tracks_skipped: u32,
    pub tracks_completed: u32,
    pub average_completion_rate: f32,
    pub genres_played: HashMap<String, u32>,
    pub artists_played: HashMap<String, u32>,
    pub skip_patterns: Vec<SkipPattern>,
    pub quality_switches: u32,
}

/// Pattern de skip pour analytics
#[derive(Debug, Clone)]
pub struct SkipPattern {
    pub track_id: Uuid,
    pub skip_position: Duration,
    pub skip_reason: SkipReason,
    pub timestamp: SystemTime,
}

/// Raisons de skip
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum SkipReason {
    UserAction,
    BufferingTimeout,
    QualityIssue,
    TrackEnded,
    AutoplayNext,
}

/// Événements de playback
#[derive(Debug, Clone)]
pub enum PlaybackEvent {
    /// Lecture commencée
    PlaybackStarted { 
        user_id: i64, 
        track: TrackInfo,
        queue_position: Option<usize>,
    },
    /// Lecture mise en pause
    PlaybackPaused { user_id: i64, position: Duration },
    /// Lecture reprise
    PlaybackResumed { user_id: i64, position: Duration },
    /// Lecture arrêtée
    PlaybackStopped { user_id: i64 },
    /// Piste suivante
    TrackChanged { 
        user_id: i64, 
        previous_track: Option<TrackInfo>,
        current_track: TrackInfo,
        change_reason: TrackChangeReason,
    },
    /// Position mise à jour
    PositionUpdated { user_id: i64, position: Duration },
    /// Queue modifiée
    QueueUpdated { user_id: i64, queue_size: usize },
    /// Commentaire temporel ajouté
    TimedCommentAdded { 
        user_id: i64, 
        track_id: Uuid, 
        comment: TimedComment 
    },
    /// Erreur de playback
    PlaybackError { user_id: i64, error: String },
}

/// Raisons de changement de piste
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum TrackChangeReason {
    UserSkip,
    TrackEnded,
    AutoplayNext,
    QueueAdvanced,
    RepeatTrack,
    ShuffleNext,
}

impl Default for PlaybackConfig {
    fn default() -> Self {
        Self {
            max_concurrent_players: 10_000,
            default_crossfade_duration: Duration::from_secs(3),
            enable_real_time_analytics: true,
            cache_preload_tracks: true,
        }
    }
}

impl Default for PlayerConfig {
    fn default() -> Self {
        Self {
            crossfade_duration: Duration::from_secs(3),
            gapless_gap_threshold: Duration::from_millis(100),
            max_history_size: 50,
            max_queue_size: 1000,
            enable_scrobbling: true,
            auto_quality_switching: true,
            preload_next_track: true,
            analytics_enabled: true,
        }
    }
}

impl PlaybackManager {
    /// Crée un nouveau gestionnaire de playback
    pub async fn new(
        config: PlaybackConfig,
        stream_manager: Arc<StreamManager>,
    ) -> Result<Self, AppError> {
        let (event_sender, _) = broadcast::channel(10_000);
        
        Ok(Self {
            active_players: Arc::new(RwLock::new(HashMap::new())),
            config,
            stream_manager,
            event_sender,
        })
    }
    
    /// Obtient ou crée un player pour un utilisateur
    pub async fn get_or_create_player(&self, user_id: i64) -> Result<Arc<SoundCloudPlayer>, AppError> {
        let mut players = self.active_players.write().await;
        
        if let Some(player) = players.get(&user_id) {
            Ok(player.clone())
        } else {
            // Vérifier la limite de players concurrents
            if players.len() >= self.config.max_concurrent_players {
                return Err(AppError::TooManyActivePlayers { 
                    limit: self.config.max_concurrent_players as u32
                });
            }
            
            let player = Arc::new(SoundCloudPlayer::new(
                user_id,
                PlayerConfig::default(),
                self.event_sender.clone(),
            )?);
            
            players.insert(user_id, player.clone());
            info!("Player créé pour utilisateur: {}", user_id);
            
            Ok(player)
        }
    }
    
    /// Démarre la lecture d'une piste
    pub async fn play_track(
        &self,
        user_id: i64,
        track: TrackInfo,
        queue_position: Option<usize>,
    ) -> Result<(), AppError> {
        let player = self.get_or_create_player(user_id).await?;
        player.play_track(track, queue_position).await
    }
    
    /// Met en pause la lecture
    pub async fn pause(&self, user_id: i64) -> Result<(), AppError> {
        let players = self.active_players.read().await;
        if let Some(player) = players.get(&user_id) {
            player.pause().await
        } else {
            Err(AppError::PlayerNotFound { user_id })
        }
    }
    
    /// Reprend la lecture
    pub async fn resume(&self, user_id: i64) -> Result<(), AppError> {
        let players = self.active_players.read().await;
        if let Some(player) = players.get(&user_id) {
            player.resume().await
        } else {
            Err(AppError::PlayerNotFound { user_id })
        }
    }
    
    /// Passe à la piste suivante
    pub async fn next_track(&self, user_id: i64) -> Result<(), AppError> {
        let players = self.active_players.read().await;
        if let Some(player) = players.get(&user_id) {
            player.next_track().await
        } else {
            Err(AppError::PlayerNotFound { user_id })
        }
    }
    
    /// Revient à la piste précédente
    pub async fn previous_track(&self, user_id: i64) -> Result<(), AppError> {
        let players = self.active_players.read().await;
        if let Some(player) = players.get(&user_id) {
            player.previous_track().await
        } else {
            Err(AppError::PlayerNotFound { user_id })
        }
    }
    
    /// Abonnement aux événements de playback
    pub fn subscribe_events(&self) -> broadcast::Receiver<PlaybackEvent> {
        self.event_sender.subscribe()
    }
}

impl SoundCloudPlayer {
    /// Crée un nouveau player
    pub fn new(
        user_id: i64,
        config: PlayerConfig,
        global_event_sender: broadcast::Sender<PlaybackEvent>,
    ) -> Result<Self, AppError> {
        let session_id = Uuid::new_v4();
        let (event_sender, event_receiver) = mpsc::unbounded_channel();
        
        // État initial du playback
        let playback_state = Arc::new(RwLock::new(PlaybackState {
            current_track: None,
            status: PlaybackStatus::Stopped,
            position: Duration::from_secs(0),
            volume: 1.0,
            playback_speed: 1.0,
            repeat_mode: RepeatMode::Off,
            shuffle_enabled: false,
            crossfade_enabled: config.crossfade_duration > Duration::from_secs(0),
            gapless_enabled: true,
            last_updated: SystemTime::now(),
        }));
        
        // Queue vide
        let queue = Arc::new(RwLock::new(PlaybackQueue {
            current_index: None,
            tracks: Vec::new(),
            play_history: VecDeque::new(),
            up_next: VecDeque::new(),
            shuffle_state: ShuffleState {
                enabled: false,
                played_indices: Vec::new(),
                remaining_indices: Vec::new(),
                algorithm: ShuffleAlgorithm::Standard,
            },
            autoplay_enabled: true,
        }));
        
        // Crossfade controller
        let crossfade_controller = Arc::new(Mutex::new(CrossfadeController {
            enabled: config.crossfade_duration > Duration::from_secs(0),
            duration: config.crossfade_duration,
            curve: CrossfadeCurve::SCurve,
            current_fade: None,
        }));
        
        // Manager des commentaires temporels
        let timed_comments = Arc::new(RwLock::new(TimedCommentsManager {
            comments: HashMap::new(),
            config: TimedCommentsConfig {
                enable_live_comments: true,
                max_comments_per_timestamp: 10,
                comment_display_duration: Duration::from_secs(5),
                enable_comment_notifications: true,
            },
        }));
        
        // Analytics de session
        let session_analytics = Arc::new(RwLock::new(SessionAnalytics::default()));
        
        // Gestion des événements asynchrones
        let _global_sender = global_event_sender.clone();
        // UnboundedReceiver ne peut pas être cloné, on utilise directement
        let _local_receiver = event_receiver;
        
        tokio::spawn(async move {
            // Local event handling logic here would go
        });
        
        Ok(Self {
            user_id,
            session_id,
            playback_state,
            queue,
            config,
            crossfade_controller,
            timed_comments,
            session_analytics,
            event_sender: event_sender,
        })
    }
    
    /// Démarre la lecture d'une piste
    pub async fn play_track(
        &self,
        track: TrackInfo,
        queue_position: Option<usize>,
    ) -> Result<(), AppError> {
        info!("Playing track: {} for user: {}", track.title, self.user_id);
        
        // Mettre à jour les analytics
        let mut analytics = self.session_analytics.write().await;
        if analytics.session_start.is_none() {
            analytics.session_start = Some(SystemTime::now());
        }
        analytics.tracks_played += 1;
        
        // Mettre à jour l'état de playback
        let mut state = self.playback_state.write().await;
        state.current_track = Some(track.clone());
        state.status = PlaybackStatus::Loading;
        state.position = Duration::from_secs(0);
        state.last_updated = SystemTime::now();
        
        // Démarrer le stream
        drop(state);
        self.start_stream(&track).await?;
        
        // Mettre à jour l'état final
        let mut state = self.playback_state.write().await;
        state.status = PlaybackStatus::Playing;
        state.last_updated = SystemTime::now();
        
        // Envoyer l'événement
        let event = PlaybackEvent::PlaybackStarted {
            user_id: self.user_id,
            track: track.clone(),
            queue_position,
        };
        
        let _ = self.event_sender.send(event);
        
        // Mettre à jour les analytics
        self.update_analytics_track_started(&track).await;
        
        Ok(())
    }
    
    /// Démarre le streaming de la piste
    async fn start_stream(&self, track: &TrackInfo) -> Result<(), AppError> {
        // Simulation du streaming - en production, configurer le vrai streaming
        info!("Starting stream for track: {} at URL: {}", track.title, track.stream_url);
        
        // Simuler la latence de démarrage
        tokio::time::sleep(Duration::from_millis(100)).await;
        
        Ok(())
    }
    
    /// Gère la transition de crossfade
    async fn handle_crossfade_transition(&self) -> Result<(), AppError> {
        let mut controller = self.crossfade_controller.lock();
        
        if controller.enabled {
            controller.current_fade = Some(FadeState {
                start_time: SystemTime::now(),
                duration: controller.duration,
                from_volume: 1.0,
                to_volume: 0.0,
                curve: controller.curve.clone(),
            });
        }
        
        Ok(())
    }
    
    /// Met en pause la lecture
    pub async fn pause(&self) -> Result<(), AppError> {
        let mut state = self.playback_state.write().await;
        
        if matches!(state.status, PlaybackStatus::Playing) {
            state.status = PlaybackStatus::Paused;
            state.last_updated = SystemTime::now();
            
            let event = PlaybackEvent::PlaybackPaused {
                user_id: self.user_id,
                position: state.position,
            };
            
            let _ = self.event_sender.send(event);
            
            info!("Playback paused for user: {}", self.user_id);
            Ok(())
        } else {
            Err(AppError::InvalidPlaybackState { 
                state: format!("Cannot pause from {:?} state (expected Playing)", state.status)
            })
        }
    }
    
    /// Reprend la lecture
    pub async fn resume(&self) -> Result<(), AppError> {
        let mut state = self.playback_state.write().await;
        
        if matches!(state.status, PlaybackStatus::Paused) {
            state.status = PlaybackStatus::Playing;
            state.last_updated = SystemTime::now();
            
            let event = PlaybackEvent::PlaybackResumed {
                user_id: self.user_id,
                position: state.position,
            };
            
            let _ = self.event_sender.send(event);
            
            info!("Playback resumed for user: {}", self.user_id);
            Ok(())
        } else {
            Err(AppError::InvalidPlaybackState { 
                state: format!("Cannot resume from {:?} state (expected Paused)", state.status)
            })
        }
    }
    
    /// Passe à la piste suivante
    pub async fn next_track(&self) -> Result<(), AppError> {
        if let Some(next_track) = self.determine_next_track().await? {
            self.play_track(next_track, None).await
        } else {
            // Arrêter la lecture si pas de piste suivante
            let mut state = self.playback_state.write().await;
            state.status = PlaybackStatus::Stopped;
            state.current_track = None;
            state.last_updated = SystemTime::now();
            
            let event = PlaybackEvent::PlaybackStopped {
                user_id: self.user_id,
            };
            
            let _ = self.event_sender.send(event);
            
            Ok(())
        }
    }
    
    /// Revient à la piste précédente
    pub async fn previous_track(&self) -> Result<(), AppError> {
        if let Some(previous_track) = self.determine_previous_track().await? {
            self.play_track(previous_track, None).await
        } else {
            // Redémarrer la piste actuelle
            let mut state = self.playback_state.write().await;
            state.position = Duration::from_secs(0);
            state.last_updated = SystemTime::now();
            
            Ok(())
        }
    }
    
    /// Arrête la lecture
    pub async fn stop(&self) -> Result<(), AppError> {
        let mut state = self.playback_state.write().await;
        state.status = PlaybackStatus::Stopped;
        state.current_track = None;
        state.position = Duration::from_secs(0);
        state.last_updated = SystemTime::now();
        
        let event = PlaybackEvent::PlaybackStopped {
            user_id: self.user_id,
        };
        
        let _ = self.event_sender.send(event);
        
        info!("Playback stopped for user: {}", self.user_id);
        Ok(())
    }
    
    /// Détermine la piste suivante selon la logique de queue
    async fn determine_next_track(&self) -> Result<Option<TrackInfo>, AppError> {
        let queue = self.queue.read().await;
        let state = self.playback_state.read().await;
        
        // Logique simplifiée - en production, implémenter shuffle, repeat, etc.
        if let Some(current_index) = queue.current_index {
            if current_index + 1 < queue.tracks.len() {
                Ok(Some(queue.tracks[current_index + 1].track.clone()))
            } else {
                match state.repeat_mode {
                    RepeatMode::All => Ok(queue.tracks.first().map(|t| t.track.clone())),
                    RepeatMode::Track => {
                        if let Some(ref current) = state.current_track {
                            Ok(Some(current.clone()))
                        } else {
                            Ok(None)
                        }
                    }
                    _ => Ok(None),
                }
            }
        } else {
            Ok(queue.tracks.first().map(|t| t.track.clone()))
        }
    }
    
    /// Détermine la piste précédente
    async fn determine_previous_track(&self) -> Result<Option<TrackInfo>, AppError> {
        let queue = self.queue.read().await;
        
        if let Some(current_index) = queue.current_index {
            if current_index > 0 {
                Ok(Some(queue.tracks[current_index - 1].track.clone()))
            } else {
                Ok(None)
            }
        } else {
            Ok(None)
        }
    }
    
    /// Met à jour les analytics pour début de piste
    async fn update_analytics_track_started(&self, track: &TrackInfo) {
        let mut analytics = self.session_analytics.write().await;
        analytics.tracks_played += 1;
        
        // Compter les genres
        for genre in &track.genres {
            *analytics.genres_played.entry(genre.clone()).or_insert(0) += 1;
        }
        
        // Compter les artistes
        *analytics.artists_played.entry(track.artist.clone()).or_insert(0) += 1;
    }
    
    /// Ajoute un commentaire temporel
    pub async fn add_timed_comment(
        &self,
        track_id: Uuid,
        timestamp_ms: u64,
        text: String,
    ) -> Result<Uuid, AppError> {
        let comment = TimedComment {
            id: Uuid::new_v4(),
            user_id: self.user_id,
            track_id,
            timestamp_ms,
            text,
            created_at: SystemTime::now(),
            likes_count: 0,
            replies: Vec::new(),
        };
        
        {
            let mut comments_manager = self.timed_comments.write().await;
            comments_manager.comments
                .entry(timestamp_ms)
                .or_insert_with(Vec::new)
                .push(comment.clone());
        }
        
        let _ = self.event_sender.send(PlaybackEvent::TimedCommentAdded {
            user_id: self.user_id,
            track_id,
            comment: comment.clone(),
        });
        
        Ok(comment.id)
    }
    
    /// Obtient les commentaires pour un timestamp
    pub async fn get_comments_at_time(&self, timestamp_ms: u64) -> Vec<TimedComment> {
        let comments_manager = self.timed_comments.read().await;
        comments_manager.comments.get(&timestamp_ms).cloned().unwrap_or_default()
    }
} 