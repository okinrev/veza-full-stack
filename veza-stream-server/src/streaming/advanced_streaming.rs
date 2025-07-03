// Advanced Streaming Engine for Phase 5

use std::collections::HashMap;
use std::sync::Arc;
use std::time::{Duration, SystemTime};
use tokio::sync::{RwLock, broadcast};
use serde::{Deserialize, Serialize};
use tracing::{info, debug, warn, span, Level};
use uuid::Uuid;
use serde_json;

use super::webrtc::{WebRTCManager, WebRTCConfig};
use super::sync_manager::{SyncManager, SyncConfig};
use super::live_recording::{LiveRecordingManager, RecordingConfig, RecordingQuality};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AdvancedStreamingConfig {
    pub webrtc: WebRTCConfig,
    pub sync: SyncConfig,
    pub recording: RecordingConfig,
    pub max_concurrent_streams: usize,
    pub adaptive_quality: bool,
    pub bandwidth_monitoring: bool,
    pub analytics_enabled: bool,
    pub failover_support: bool,
}

impl Default for AdvancedStreamingConfig {
    fn default() -> Self {
        Self {
            webrtc: WebRTCConfig::default(),
            sync: SyncConfig::default(),
            recording: RecordingConfig::default(),
            max_concurrent_streams: 100,
            adaptive_quality: true,
            bandwidth_monitoring: true,
            analytics_enabled: true,
            failover_support: true,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct StreamSession {
    pub session_id: String,
    pub stream_id: String,
    pub user_id: String,
    pub stream_type: StreamType,
    pub state: StreamState,
    pub start_time: SystemTime,
    pub end_time: Option<SystemTime>,
    pub current_quality: String,
    pub listeners: Vec<ListenerInfo>,
    pub recording_id: Option<String>,
    pub webrtc_peer_id: Option<String>,
    pub sync_client_id: Option<String>,
    pub analytics: StreamAnalytics,
    pub metadata: StreamMetadata,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum StreamType {
    Audio,
    Video,
    AudioVideo,
    Screen,
    Chat,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum StreamState {
    Initializing,
    Starting,
    Live,
    Paused,
    Buffering,
    Ending,
    Completed,
    Failed,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ListenerInfo {
    pub listener_id: String,
    pub user_id: String,
    pub joined_at: SystemTime,
    pub connection_type: String,
    pub quality_preference: String,
    pub bandwidth_kbps: u32,
    pub is_synchronized: bool,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct StreamAnalytics {
    pub total_listeners: u32,
    pub peak_listeners: u32,
    pub average_listener_duration_ms: u64,
    pub total_data_transferred_mb: f32,
    pub average_bitrate_kbps: u32,
    pub buffer_events: u32,
    pub quality_switches: u32,
    pub connection_drops: u32,
    pub geographic_distribution: HashMap<String, u32>,
}

impl Default for StreamAnalytics {
    fn default() -> Self {
        Self {
            total_listeners: 0,
            peak_listeners: 0,
            average_listener_duration_ms: 0,
            total_data_transferred_mb: 0.0,
            average_bitrate_kbps: 0,
            buffer_events: 0,
            quality_switches: 0,
            connection_drops: 0,
            geographic_distribution: HashMap::new(),
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct StreamMetadata {
    pub title: String,
    pub description: Option<String>,
    pub tags: Vec<String>,
    pub category: String,
    pub language: String,
    pub thumbnail_url: Option<String>,
    pub duration_ms: Option<u64>,
    pub is_public: bool,
    pub scheduled_start: Option<SystemTime>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(tag = "type", content = "data")]
pub enum StreamingMessage {
    StreamStarted {
        session_id: String,
        stream_id: String,
        quality: String,
    },
    StreamEnded {
        session_id: String,
        duration_ms: u64,
    },
    ListenerJoined {
        session_id: String,
        listener: ListenerInfo,
    },
    ListenerLeft {
        session_id: String,
        listener_id: String,
    },
    QualityChanged {
        session_id: String,
        old_quality: String,
        new_quality: String,
    },
    BufferEvent {
        session_id: String,
        listener_id: String,
        event_type: String,
    },
    AnalyticsUpdate {
        session_id: String,
        analytics: StreamAnalytics,
    },
    Error {
        session_id: String,
        error: String,
    },
}

/// Moteur de streaming avancé Phase 5
#[derive(Clone)]
pub struct AdvancedStreamingEngine {
    config: AdvancedStreamingConfig,
    sessions: Arc<RwLock<HashMap<String, StreamSession>>>,
    webrtc_manager: Arc<WebRTCManager>,
    sync_manager: Arc<SyncManager>,
    recording_manager: Arc<LiveRecordingManager>,
    streaming_tx: broadcast::Sender<StreamingMessage>,
    analytics_collector: Arc<RwLock<HashMap<String, StreamAnalytics>>>,
}

impl AdvancedStreamingEngine {
    pub fn new(config: AdvancedStreamingConfig) -> Self {
        let webrtc_manager = Arc::new(WebRTCManager::new(config.webrtc.clone()));
        let sync_manager = Arc::new(SyncManager::new(config.sync.clone()));
        let recording_manager = Arc::new(LiveRecordingManager::new(config.recording.clone()));
        let (streaming_tx, _) = broadcast::channel(1000);

        Self {
            config,
            sessions: Arc::new(RwLock::new(HashMap::new())),
            webrtc_manager,
            sync_manager,
            recording_manager,
            streaming_tx,
            analytics_collector: Arc::new(RwLock::new(HashMap::new())),
        }
    }

    /// Démarrer le moteur de streaming avancé
    pub async fn start(&self) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
        info!("Starting Advanced Streaming Engine Phase 5");

        // Démarrer tous les gestionnaires
        self.webrtc_manager.start().await?;
        self.sync_manager.start().await?;
        self.recording_manager.start().await?;

        // Démarrer les services internes
        self.start_session_monitor().await;
        self.start_quality_adapter().await;
        
        if self.config.analytics_enabled {
            self.start_analytics_collector().await;
        }

        if self.config.bandwidth_monitoring {
            self.start_bandwidth_monitor().await;
        }

        info!("Advanced Streaming Engine Phase 5 started successfully");
        Ok(())
    }

    /// Créer une nouvelle session de streaming
    pub async fn create_stream_session(
        &self,
        user_id: String,
        stream_type: StreamType,
        metadata: StreamMetadata,
        enable_recording: bool,
    ) -> Result<String, Box<dyn std::error::Error + Send + Sync>> {
        let span = span!(Level::INFO, "create_stream_session", user_id = %user_id);
        let _enter = span.enter();

        let mut sessions = self.sessions.write().await;
        
        if sessions.len() >= self.config.max_concurrent_streams {
            return Err("Maximum number of concurrent streams reached".into());
        }

        let session_id = Uuid::new_v4().to_string();
        let stream_id = format!("stream_{}", Uuid::new_v4().simple());

        // Créer la session de streaming
        let mut session = StreamSession {
            session_id: session_id.clone(),
            stream_id: stream_id.clone(),
            user_id: user_id.clone(),
            stream_type,
            state: StreamState::Initializing,
            start_time: SystemTime::now(),
            end_time: None,
            current_quality: "medium".to_string(),
            listeners: Vec::new(),
            recording_id: None,
            webrtc_peer_id: None,
            sync_client_id: None,
            analytics: StreamAnalytics::default(),
            metadata,
        };

        // Initialiser WebRTC peer
        match self.webrtc_manager.create_peer_session(
            format!("peer_{}", session_id),
            session_id.clone(),
        ).await {
            Ok(peer) => {
                session.webrtc_peer_id = Some(peer.peer_id);
                info!("WebRTC peer created for session: {}", session_id);
            }
            Err(e) => {
                warn!("Failed to create WebRTC peer: {}", e);
            }
        }

        // Ajouter client de synchronisation
        match self.sync_manager.add_client(
            format!("sync_{}", session_id),
            session_id.clone(),
        ).await {
            Ok(client) => {
                session.sync_client_id = Some(client.client_id);
                info!("Sync client added for session: {}", session_id);
            }
            Err(e) => {
                warn!("Failed to add sync client: {}", e);
            }
        }

        // Démarrer l'enregistrement si demandé
        if enable_recording {
            let recording_quality = RecordingQuality::high();
            let recording_metadata = crate::streaming::live_recording::RecordingMetadata {
                title: Some(session.metadata.title.clone()),
                artist: Some(user_id.clone()),
                album: None,
                genre: Some(session.metadata.category.clone()),
                duration_ms: 0,
                bitrate: recording_quality.bitrate,
                sample_rate: recording_quality.sample_rate,
                channels: recording_quality.channels,
                file_size_bytes: 0,
                creation_time: SystemTime::now(),
                tags: session.metadata.tags.iter()
                    .enumerate()
                    .map(|(i, tag)| (format!("tag_{}", i), tag.clone()))
                    .collect(),
            };

            match self.recording_manager.start_recording(
                session_id.clone(),
                stream_id.clone(),
                recording_quality,
                recording_metadata,
            ).await {
                Ok(recording_id) => {
                    session.recording_id = Some(recording_id.clone());
                    info!("Recording started for session: {} with ID: {}", session_id, recording_id);
                }
                Err(e) => {
                    warn!("Failed to start recording: {}", e);
                }
            }
        }

        session.state = StreamState::Starting;
        sessions.insert(session_id.clone(), session);

        info!("Created stream session: {} for user: {}", session_id, user_id);

        // Envoyer message de démarrage
        let start_msg = StreamingMessage::StreamStarted {
            session_id: session_id.clone(),
            stream_id,
            quality: "medium".to_string(),
        };

        if let Err(e) = self.streaming_tx.send(start_msg) {
            warn!("Failed to send stream started message: {}", e);
        }

        Ok(session_id)
    }

    /// Ajouter un listener à une session
    pub async fn add_listener(
        &self,
        session_id: &str,
        user_id: String,
        connection_type: String,
        quality_preference: String,
        bandwidth_kbps: u32,
    ) -> Result<String, Box<dyn std::error::Error + Send + Sync>> {
        let mut sessions = self.sessions.write().await;
        
        if let Some(session) = sessions.get_mut(session_id) {
            let listener_id = Uuid::new_v4().to_string();
            
            let listener = ListenerInfo {
                listener_id: listener_id.clone(),
                user_id: user_id.clone(),
                joined_at: SystemTime::now(),
                connection_type,
                quality_preference,
                bandwidth_kbps,
                is_synchronized: false,
            };

            session.listeners.push(listener.clone());
            session.analytics.total_listeners += 1;
            
            if session.listeners.len() as u32 > session.analytics.peak_listeners {
                session.analytics.peak_listeners = session.listeners.len() as u32;
            }

            info!("Added listener {} to session {}", listener_id, session_id);

            // Envoyer message de listener ajouté
            let listener_msg = StreamingMessage::ListenerJoined {
                session_id: session_id.to_string(),
                listener,
            };

            if let Err(e) = self.streaming_tx.send(listener_msg) {
                warn!("Failed to send listener joined message: {}", e);
            }

            // Synchroniser le nouveau listener
            if let Some(sync_client_id) = &session.sync_client_id {
                // Ici on ajouterait la logique de synchronisation du listener
                debug!("Synchronizing new listener with sync client: {}", sync_client_id);
            }

            Ok(listener_id)
        } else {
            Err("Session not found".into())
        }
    }

    /// Supprimer un listener d'une session
    pub async fn remove_listener(
        &self,
        session_id: &str,
        listener_id: &str,
    ) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
        let mut sessions = self.sessions.write().await;
        
        if let Some(session) = sessions.get_mut(session_id) {
            if let Some(pos) = session.listeners.iter().position(|l| l.listener_id == listener_id) {
                let listener = session.listeners.remove(pos);
                
                // Calculer la durée d'écoute
                if let Ok(duration) = listener.joined_at.elapsed() {
                    let duration_ms = duration.as_millis() as u64;
                    
                    // Mettre à jour les analytics
                    let total_duration = session.analytics.average_listener_duration_ms * (session.analytics.total_listeners - 1) as u64 + duration_ms;
                    session.analytics.average_listener_duration_ms = total_duration / session.analytics.total_listeners as u64;
                }

                info!("Removed listener {} from session {}", listener_id, session_id);

                // Envoyer message de listener parti
                let left_msg = StreamingMessage::ListenerLeft {
                    session_id: session_id.to_string(),
                    listener_id: listener_id.to_string(),
                };

                if let Err(e) = self.streaming_tx.send(left_msg) {
                    warn!("Failed to send listener left message: {}", e);
                }

                Ok(())
            } else {
                Err("Listener not found in session".into())
            }
        } else {
            Err("Session not found".into())
        }
    }

    /// Terminer une session de streaming
    pub async fn end_stream_session(&self, session_id: &str) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
        let mut sessions = self.sessions.write().await;
        
        if let Some(session) = sessions.get_mut(session_id) {
            session.state = StreamState::Ending;
            session.end_time = Some(SystemTime::now());

            let duration_ms = if let Ok(duration) = session.start_time.elapsed() {
                duration.as_millis() as u64
            } else {
                0
            };

            // Arrêter l'enregistrement si actif
            if let Some(recording_id) = &session.recording_id {
                if let Err(e) = self.recording_manager.stop_recording(recording_id).await {
                    warn!("Failed to stop recording {}: {}", recording_id, e);
                }
            }

            // Nettoyer les ressources WebRTC
            if let Some(peer_id) = &session.webrtc_peer_id {
                self.webrtc_manager.remove_peer(peer_id).await;
            }

            // Nettoyer les ressources de synchronisation
            if let Some(sync_client_id) = &session.sync_client_id {
                self.sync_manager.remove_client(sync_client_id).await;
            }

            session.state = StreamState::Completed;

            info!("Ended stream session: {} (duration: {}ms)", session_id, duration_ms);

            // Envoyer message de fin
            let end_msg = StreamingMessage::StreamEnded {
                session_id: session_id.to_string(),
                duration_ms,
            };

            if let Err(e) = self.streaming_tx.send(end_msg) {
                warn!("Failed to send stream ended message: {}", e);
            }

            Ok(())
        } else {
            Err("Session not found".into())
        }
    }

    /// Obtenir les statistiques globales en temps réel
    pub async fn get_global_stats(&self) -> serde_json::Value {
        let sessions = self.sessions.read().await;
        let webrtc_stats = self.webrtc_manager.get_real_time_stats().await;
        let sync_stats = self.sync_manager.get_sync_stats().await;
        let recording_stats = self.recording_manager.get_recording_stats().await;

        let total_sessions = sessions.len();
        let active_sessions = sessions.values()
            .filter(|s| matches!(s.state, StreamState::Live))
            .count();

        let total_listeners: u32 = sessions.values()
            .map(|s| s.listeners.len() as u32)
            .sum();

        let total_data_transferred_mb: f32 = sessions.values()
            .map(|s| s.analytics.total_data_transferred_mb)
            .sum();

        let avg_bitrate: f32 = if active_sessions > 0 {
            sessions.values()
                .filter(|s| matches!(s.state, StreamState::Live))
                .map(|s| s.analytics.average_bitrate_kbps as f32)
                .sum::<f32>() / active_sessions as f32
        } else {
            0.0
        };

        serde_json::json!({
            "phase5_streaming_stats": {
                "total_sessions": total_sessions,
                "active_sessions": active_sessions,
                "total_listeners": total_listeners,
                "total_data_transferred_mb": total_data_transferred_mb,
                "average_bitrate_kbps": avg_bitrate,
                "max_concurrent_streams": self.config.max_concurrent_streams,
                "adaptive_quality_enabled": self.config.adaptive_quality,
                "webrtc": webrtc_stats,
                "synchronization": sync_stats,
                "recording": recording_stats
            }
        })
    }

    /// Démarrer le moniteur de sessions
    async fn start_session_monitor(&self) {
        let sessions = self.sessions.clone();
        let _streaming_tx = self.streaming_tx.clone();

        tokio::spawn(async move {
            let mut interval = tokio::time::interval(Duration::from_secs(10));
            
            loop {
                interval.tick().await;
                
                let sessions_guard = sessions.read().await;
                for (session_id, session) in sessions_guard.iter() {
                    match session.state {
                        StreamState::Live => {
                            debug!("Session {} live with {} listeners", 
                                   session_id, session.listeners.len());
                        }
                        StreamState::Failed => {
                            warn!("Session {} in failed state", session_id);
                        }
                        _ => {}
                    }
                }
            }
        });
    }

    /// Démarrer l'adaptateur de qualité
    async fn start_quality_adapter(&self) {
        if !self.config.adaptive_quality {
            return;
        }

        let sessions = self.sessions.clone();
        let _streaming_tx = self.streaming_tx.clone();

        tokio::spawn(async move {
            let mut interval = tokio::time::interval(Duration::from_secs(5));
            
            loop {
                interval.tick().await;
                
                let mut sessions_guard = sessions.write().await;
                for (session_id, session) in sessions_guard.iter_mut() {
                    if matches!(session.state, StreamState::Live) {
                        // Logique d'adaptation de qualité basée sur les listeners
                        let avg_bandwidth: f32 = if !session.listeners.is_empty() {
                            session.listeners.iter()
                                .map(|l| l.bandwidth_kbps as f32)
                                .sum::<f32>() / session.listeners.len() as f32
                        } else {
                            1000.0
                        };

                        let new_quality = if avg_bandwidth > 500.0 {
                            "high"
                        } else if avg_bandwidth > 200.0 {
                            "medium"
                        } else {
                            "low"
                        };

                        if new_quality != session.current_quality {
                            let old_quality = session.current_quality.clone();
                            session.current_quality = new_quality.to_string();
                            session.analytics.quality_switches += 1;

                            let quality_msg = StreamingMessage::QualityChanged {
                                session_id: session_id.clone(),
                                old_quality,
                                new_quality: new_quality.to_string(),
                            };

                            if let Err(e) = _streaming_tx.send(quality_msg) {
                                warn!("Failed to send quality change message: {}", e);
                            }
                        }
                    }
                }
            }
        });
    }

    /// Démarrer le collecteur d'analytics
    async fn start_analytics_collector(&self) {
        let sessions = self.sessions.clone();
        let analytics_collector = self.analytics_collector.clone();

        tokio::spawn(async move {
            let mut interval = tokio::time::interval(Duration::from_secs(30));
            
            loop {
                interval.tick().await;
                
                let sessions_guard = sessions.read().await;
                let mut analytics_guard = analytics_collector.write().await;
                
                for (session_id, session) in sessions_guard.iter() {
                    analytics_guard.insert(session_id.clone(), session.analytics.clone());
                }
            }
        });
    }

    /// Démarrer le moniteur de bande passante
    async fn start_bandwidth_monitor(&self) {
        let sessions = self.sessions.clone();

        tokio::spawn(async move {
            let mut interval = tokio::time::interval(Duration::from_secs(15));
            
            loop {
                interval.tick().await;
                
                let sessions_guard = sessions.read().await;
                for (session_id, session) in sessions_guard.iter() {
                    if matches!(session.state, StreamState::Live) {
                        let total_bandwidth: u32 = session.listeners.iter()
                            .map(|l| l.bandwidth_kbps)
                            .sum();
                        
                        debug!("Session {} bandwidth usage: {} kbps", 
                               session_id, total_bandwidth);
                    }
                }
            }
        });
    }

    /// Obtenir un receiver pour les messages de streaming
    pub fn get_streaming_receiver(&self) -> broadcast::Receiver<StreamingMessage> {
        self.streaming_tx.subscribe()
    }

    /// Obtenir les statistiques d'une session spécifique
    pub async fn get_session_stats(&self, session_id: &str) -> Option<StreamSession> {
        let sessions = self.sessions.read().await;
        sessions.get(session_id).cloned()
    }
}
