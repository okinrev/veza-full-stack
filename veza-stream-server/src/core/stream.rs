/// Core Stream Management pour production
/// Support de 10k+ streams simultanés avec gestion optimisée
use std::sync::Arc;
use std::time::{Duration, Instant};
use std::collections::HashMap;

use dashmap::DashMap;
use parking_lot::RwLock;
use tokio::sync::{broadcast, mpsc};
use uuid::Uuid;
use serde::{Serialize, Deserialize};
use tracing::{info, warn, error, debug};

use crate::core::AudioFormat;
use crate::error::AppError;

/// Gestionnaire principal des streams en production
#[derive(Debug)]
pub struct StreamManager {
    /// Streams actifs avec accès concurrent optimisé
    streams: Arc<DashMap<Uuid, LiveStream>>,
    /// Pool d'encodeurs réutilisables
    encoder_pool: Arc<crate::core::EncoderPool>,
    /// Gestionnaire de buffers adaptatifs
    buffer_manager: Arc<crate::core::BufferManager>,
    /// Analytics temps réel
    analytics: Arc<StreamAnalytics>,
    /// Événements globaux (nouveaux streams, fin, etc.)
    event_sender: broadcast::Sender<StreamEvent>,
    /// Configuration globale
    config: Arc<RwLock<StreamConfig>>,
}

/// Représentation d'un stream live en cours
#[derive(Debug)]
pub struct LiveStream {
    pub id: Uuid,
    pub title: String,
    pub description: Option<String>,
    pub creator_id: i64,
    pub source: StreamSource,
    pub outputs: Vec<StreamOutput>,
    pub encoder_pipeline: Arc<crate::core::EncoderPipeline>,
    pub buffer: Arc<crate::core::AdaptiveBuffer>,
    pub listeners: Arc<DashMap<Uuid, Listener>>,
    pub metadata: Arc<RwLock<StreamMetadata>>,
    pub analytics: StreamAnalytics,
    pub started_at: Instant,
    pub status: StreamStatus,
}

/// Sources possibles pour un stream
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
pub enum StreamSource {
    /// Upload direct de fichier audio
    File { 
        path: String,
        format: AudioFormat,
        duration: Option<Duration>,
    },
    /// Stream live depuis microphone/input
    Live { 
        input_device: String,
        format: AudioFormat,
        bitrate: u32,
    },
    /// URL externe (autre stream, radio, etc.)
    External { 
        url: String,
        format: Option<AudioFormat>,
    },
    /// Stream généré (synthèse, silence, etc.)
    Generated { 
        generator_type: String,
        parameters: HashMap<String, String>,
    },
}

/// Formats de sortie pour distribution
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct StreamOutput {
    pub format: AudioFormat,
    pub bitrate: u32,
    pub protocol: StreamProtocol,
    pub endpoint: String,
    pub listeners_count: Arc<std::sync::atomic::AtomicU64>,
}

/// Protocoles de streaming supportés
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum StreamProtocol {
    /// HTTP Live Streaming (HLS)
    HLS {
        segment_duration: Duration,
        playlist_size: usize,
    },
    /// Dynamic Adaptive Streaming (DASH)
    DASH {
        segment_duration: Duration,
        adaptation_sets: Vec<String>,
    },
    /// WebRTC pour ultra-low latency
    WebRTC {
        ice_servers: Vec<String>,
        turn_credentials: Option<String>,
    },
    /// WebSocket direct streaming
    WebSocket {
        compression: bool,
        binary_mode: bool,
    },
    /// RTMP pour compatibilité
    RTMP {
        server_url: String,
        stream_key: String,
    },
}

/// Informations sur un listener connecté
#[derive(Debug, Clone)]
pub struct Listener {
    pub id: Uuid,
    pub user_id: Option<i64>,
    pub ip_address: String,
    pub user_agent: Option<String>,
    pub connected_at: Instant,
    pub current_quality: String,
    pub bandwidth_estimate: u32,
    pub buffer_health: f32,
    pub session_data: HashMap<String, String>,
}

/// Métadonnées temps réel du stream
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct StreamMetadata {
    pub current_position: Duration,
    pub total_duration: Option<Duration>,
    pub current_track: Option<TrackInfo>,
    pub next_track: Option<TrackInfo>,
    pub volume: f32,
    pub playback_speed: f32,
    pub effects_enabled: Vec<String>,
    pub tags: Vec<String>,
    pub language: Option<String>,
    pub artwork_url: Option<String>,
}

/// Informations sur une piste audio
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TrackInfo {
    pub title: String,
    pub artist: Option<String>,
    pub album: Option<String>,
    pub duration: Option<Duration>,
    pub isrc: Option<String>,
    pub bpm: Option<f32>,
    pub key: Option<String>,
    pub genre: Option<String>,
}

/// Status du stream
#[derive(Debug, Clone, Copy, Serialize, Deserialize, PartialEq)]
pub enum StreamStatus {
    Starting,
    Live,
    Paused,
    Buffering,
    Reconnecting,
    Ending,
    Ended,
    Error,
}

/// Configuration globale du streaming
#[derive(Debug, Clone)]
pub struct StreamConfig {
    pub max_concurrent_streams: usize,
    pub max_listeners_per_stream: usize,
    pub default_audio_format: AudioFormat,
    pub adaptive_bitrate_enabled: bool,
    pub cdn_endpoints: Vec<String>,
    pub analytics_enabled: bool,
    pub recording_enabled: bool,
    pub transcoding_enabled: bool,
}

/// Analytics temps réel pour un stream
#[derive(Debug, Clone, Default)]
pub struct StreamAnalytics {
    pub current_listeners: u64,
    pub peak_listeners: u64,
    pub total_listening_time: Duration,
    pub average_session_duration: Duration,
    pub quality_switches: u64,
    pub buffering_events: u64,
    pub geographic_distribution: HashMap<String, u64>,
    pub device_distribution: HashMap<String, u64>,
    pub bandwidth_stats: BandwidthStats,
}

/// Statistiques de bande passante
#[derive(Debug, Clone, Default)]
pub struct BandwidthStats {
    pub average_bitrate: f64,
    pub peak_bitrate: u32,
    pub total_bytes_sent: u64,
    pub compression_ratio: f32,
}

/// Événements du système de streaming
#[derive(Debug, Clone)]
pub enum StreamEvent {
    StreamStarted { stream_id: Uuid, creator_id: i64 },
    StreamEnded { stream_id: Uuid, duration: Duration },
    ListenerJoined { stream_id: Uuid, listener_id: Uuid },
    ListenerLeft { stream_id: Uuid, listener_id: Uuid },
    QualityChanged { stream_id: Uuid, listener_id: Uuid, new_quality: String },
    ErrorOccurred { stream_id: Uuid, error: String },
    AnalyticsUpdate { stream_id: Uuid, analytics: StreamAnalytics },
}

impl Default for StreamConfig {
    fn default() -> Self {
        Self {
            max_concurrent_streams: 10_000,
            max_listeners_per_stream: 100_000,
            default_audio_format: AudioFormat {
                sample_rate: 44100,
                channels: 2,
                bit_depth: 16,
            },
            adaptive_bitrate_enabled: true,
            cdn_endpoints: vec![],
            analytics_enabled: true,
            recording_enabled: false,
            transcoding_enabled: true,
        }
    }
}

impl StreamManager {
    /// Crée un nouveau gestionnaire de streams
    pub fn new(config: StreamConfig) -> Result<Self, AppError> {
        let (event_sender, _) = broadcast::channel(10_000);
        
        Ok(Self {
            streams: Arc::new(DashMap::new()),
            encoder_pool: Arc::new(crate::core::EncoderPool::new()?),
            buffer_manager: Arc::new(crate::core::BufferManager::new()),
            analytics: Arc::new(StreamAnalytics::default()),
            event_sender,
            config: Arc::new(RwLock::new(config)),
        })
    }
    
    /// Démarre un nouveau stream
    pub async fn create_stream(
        &self,
        creator_id: i64,
        source: StreamSource,
        outputs: Vec<StreamOutput>,
        metadata: StreamMetadata,
    ) -> Result<Uuid, AppError> {
        let config = self.config.read();
        
        // Vérifier les limites
        if self.streams.len() >= config.max_concurrent_streams {
            return Err(AppError::LimitExceeded {
                limit: config.max_concurrent_streams,
                current: self.streams.len(),
            });
        }
        
        let stream_id = Uuid::new_v4();
        
        // Créer le pipeline d'encodage
        let encoder_pipeline = self.encoder_pool.create_pipeline(&source, &outputs).await?;
        
        // Créer le buffer adaptatif
        let buffer = self.buffer_manager.create_buffer(stream_id).await?;
        
        let stream = LiveStream {
            id: stream_id,
            title: metadata.current_track.as_ref()
                .map(|t| t.title.clone())
                .unwrap_or_else(|| format!("Stream {}", stream_id)),
            description: None,
            creator_id,
            source,
            outputs,
            encoder_pipeline,
            buffer,
            listeners: Arc::new(DashMap::new()),
            metadata: Arc::new(RwLock::new(metadata)),
            analytics: StreamAnalytics::default(),
            started_at: Instant::now(),
            status: StreamStatus::Starting,
        };
        
        self.streams.insert(stream_id, stream);
        
        // Émettre l'événement
        let _ = self.event_sender.send(StreamEvent::StreamStarted {
            stream_id,
            creator_id,
        });
        
        info!("Stream créé: {} par utilisateur {}", stream_id, creator_id);
        Ok(stream_id)
    }
    
    /// Ajoute un listener à un stream
    pub async fn add_listener(
        &self,
        stream_id: Uuid,
        listener: Listener,
    ) -> Result<(), AppError> {
        let stream = self.streams.get(&stream_id)
            .ok_or_else(|| AppError::NotFound { stream_id })?;
        
        let config = self.config.read();
        if stream.listeners.len() >= config.max_listeners_per_stream {
            return Err(AppError::ListenerLimitExceeded {
                stream_id,
                limit: config.max_listeners_per_stream,
            });
        }
        
        let listener_id = listener.id;
        stream.listeners.insert(listener_id, listener);
        
        // Émettre l'événement
        let _ = self.event_sender.send(StreamEvent::ListenerJoined {
            stream_id,
            listener_id,
        });
        
        debug!("Listener {} ajouté au stream {}", listener_id, stream_id);
        Ok(())
    }
    
    /// Retire un listener d'un stream
    pub async fn remove_listener(
        &self,
        stream_id: Uuid,
        listener_id: Uuid,
    ) -> Result<(), AppError> {
        let stream = self.streams.get(&stream_id)
            .ok_or_else(|| AppError::NotFound { stream_id })?;
        
        stream.listeners.remove(&listener_id);
        
        // Émettre l'événement
        let _ = self.event_sender.send(StreamEvent::ListenerLeft {
            stream_id,
            listener_id,
        });
        
        debug!("Listener {} retiré du stream {}", listener_id, stream_id);
        Ok(())
    }
    
    /// Termine un stream
    pub async fn end_stream(&self, stream_id: Uuid) -> Result<(), AppError> {
        let (_, stream) = self.streams.remove(&stream_id)
            .ok_or_else(|| AppError::NotFound { stream_id })?;
        
        let duration = stream.started_at.elapsed();
        
        // Émettre l'événement
        let _ = self.event_sender.send(StreamEvent::StreamEnded {
            stream_id,
            duration,
        });
        
        info!("Stream {} terminé après {:?}", stream_id, duration);
        Ok(())
    }
    
    /// Obtient les statistiques globales
    pub fn get_global_stats(&self) -> HashMap<String, u64> {
        let mut stats = HashMap::new();
        stats.insert("total_streams".to_string(), self.streams.len() as u64);
        
        let total_listeners: u64 = self.streams
            .iter()
            .map(|stream| stream.listeners.len() as u64)
            .sum();
        stats.insert("total_listeners".to_string(), total_listeners);
        
        stats
    }
    
    /// Abonnement aux événements du système
    pub fn subscribe_events(&self) -> broadcast::Receiver<StreamEvent> {
        self.event_sender.subscribe()
    }
} 