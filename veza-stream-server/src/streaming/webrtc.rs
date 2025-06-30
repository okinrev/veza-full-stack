use std::collections::HashMap;
use std::sync::Arc;
use std::time::{Duration, Instant, SystemTime};
use tokio::sync::{RwLock, broadcast, mpsc};
use serde::{Deserialize, Serialize};
use tracing::{info, debug, warn, error, span, Level};
use uuid::Uuid;

/// Configuration WebRTC pour streaming audio
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct WebRTCConfig {
    pub ice_servers: Vec<IceServer>,
    pub max_peers: usize,
    pub connection_timeout: Duration,
    pub heartbeat_interval: Duration,
    pub codec_preferences: Vec<AudioCodec>,
    pub bitrate_adaptation: bool,
    pub jitter_buffer_ms: u32,
}

impl Default for WebRTCConfig {
    fn default() -> Self {
        Self {
            ice_servers: vec![
                IceServer {
                    urls: vec!["stun:stun.l.google.com:19302".to_string()],
                    username: None,
                    credential: None,
                },
            ],
            max_peers: 1000,
            connection_timeout: Duration::from_secs(30),
            heartbeat_interval: Duration::from_secs(10),
            codec_preferences: vec![
                AudioCodec::Opus { bitrate: 320 },
                AudioCodec::Aac { bitrate: 256 },
                AudioCodec::Mp3 { bitrate: 192 },
            ],
            bitrate_adaptation: true,
            jitter_buffer_ms: 100,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct IceServer {
    pub urls: Vec<String>,
    pub username: Option<String>,
    pub credential: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum AudioCodec {
    Opus { bitrate: u32 },
    Aac { bitrate: u32 },
    Mp3 { bitrate: u32 },
    Pcm { sample_rate: u32 },
}

impl AudioCodec {
    pub fn get_bitrate(&self) -> u32 {
        match self {
            AudioCodec::Opus { bitrate } => *bitrate,
            AudioCodec::Aac { bitrate } => *bitrate,
            AudioCodec::Mp3 { bitrate } => *bitrate,
            AudioCodec::Pcm { sample_rate } => sample_rate * 16 * 2 / 1000,
        }
    }

    pub fn get_mime_type(&self) -> &'static str {
        match self {
            AudioCodec::Opus { .. } => "audio/opus",
            AudioCodec::Aac { .. } => "audio/aac",
            AudioCodec::Mp3 { .. } => "audio/mpeg",
            AudioCodec::Pcm { .. } => "audio/pcm",
        }
    }
}

/// Informations sur un peer WebRTC
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct WebRTCPeer {
    pub peer_id: String,
    pub session_id: String,
    pub connection_state: ConnectionState,
    pub ice_connection_state: IceConnectionState,
    pub selected_codec: Option<AudioCodec>,
    pub bandwidth_estimate: u32,
    pub rtt_ms: Option<u32>,
    pub jitter_ms: Option<u32>,
    pub packet_loss_percentage: f32,
    pub connected_at: SystemTime,
    #[serde(skip, default = "default_instant")]
    pub last_activity: Instant,
    pub stats: PeerStats,
}

fn default_instant() -> Instant {
    Instant::now()
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum ConnectionState {
    New,
    Connecting,
    Connected,
    Disconnected,
    Failed,
    Closed,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum IceConnectionState {
    New,
    Checking,
    Connected,
    Completed,
    Disconnected,
    Failed,
    Closed,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PeerStats {
    pub bytes_sent: u64,
    pub bytes_received: u64,
    pub packets_sent: u64,
    pub packets_received: u64,
    pub packets_lost: u64,
    pub audio_level: f32,
    pub quality_switches: u32,
}

impl Default for PeerStats {
    fn default() -> Self {
        Self {
            bytes_sent: 0,
            bytes_received: 0,
            packets_sent: 0,
            packets_received: 0,
            packets_lost: 0,
            audio_level: 0.0,
            quality_switches: 0,
        }
    }
}

/// Messages WebRTC pour signaling
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(tag = "type", content = "data")]
pub enum WebRTCMessage {
    Offer {
        peer_id: String,
        sdp: String,
        session_id: String,
    },
    Answer {
        peer_id: String,
        sdp: String,
    },
    IceCandidate {
        peer_id: String,
        candidate: String,
        sdp_mid: Option<String>,
        sdp_mline_index: Option<u16>,
    },
    BitrateChange {
        peer_id: String,
        new_bitrate: u32,
    },
    CodecChange {
        peer_id: String,
        codec: AudioCodec,
    },
    QualityUpdate {
        peer_id: String,
        bandwidth: u32,
        rtt: u32,
        packet_loss: f32,
    },
    PeerDisconnected {
        peer_id: String,
    },
    Error {
        peer_id: String,
        message: String,
    },
}

/// Gestionnaire WebRTC principal
#[derive(Clone)]
pub struct WebRTCManager {
    config: WebRTCConfig,
    peers: Arc<RwLock<HashMap<String, WebRTCPeer>>>,
    signaling_tx: broadcast::Sender<WebRTCMessage>,
    stats_tx: mpsc::Sender<PeerStats>,
}

impl WebRTCManager {
    pub fn new(config: WebRTCConfig) -> Self {
        let (signaling_tx, _signaling_rx) = broadcast::channel(1000);
        let (stats_tx, _stats_rx) = mpsc::channel(100);

        Self {
            config,
            peers: Arc::new(RwLock::new(HashMap::new())),
            signaling_tx,
            stats_tx,
        }
    }

    /// Démarre le gestionnaire WebRTC
    pub async fn start(&self) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
        info!("Starting WebRTC Manager with max {} peers", self.config.max_peers);

        // Démarrer le moniteur de connexions
        self.start_connection_monitor().await;
        
        // Démarrer l'adaptation de bitrate
        if self.config.bitrate_adaptation {
            self.start_bitrate_adaptation().await;
        }

        // Démarrer le collecteur de statistiques
        self.start_stats_collector().await;

        Ok(())
    }

    /// Créer une nouvelle session peer
    pub async fn create_peer_session(
        &self,
        peer_id: String,
        session_id: String,
    ) -> Result<WebRTCPeer, Box<dyn std::error::Error + Send + Sync>> {
        let span = span!(Level::INFO, "create_peer_session", peer_id = %peer_id);
        let _enter = span.enter();

        let mut peers = self.peers.write().await;
        
        if peers.len() >= self.config.max_peers {
            return Err("Maximum number of peers reached".into());
        }

        let peer = WebRTCPeer {
            peer_id: peer_id.clone(),
            session_id: session_id.clone(),
            connection_state: ConnectionState::New,
            ice_connection_state: IceConnectionState::New,
            selected_codec: None,
            bandwidth_estimate: 1000,
            rtt_ms: None,
            jitter_ms: None,
            packet_loss_percentage: 0.0,
            connected_at: SystemTime::now(),
            last_activity: Instant::now(),
            stats: PeerStats::default(),
        };

        peers.insert(peer_id.clone(), peer.clone());
        
        info!("Created WebRTC peer session: {} for session: {}", peer_id, session_id);
        Ok(peer)
    }

    /// Sélectionner le meilleur codec pour un peer
    pub async fn select_optimal_codec(
        &self,
        _peer_id: &str,
        bandwidth_estimate: u32,
    ) -> Option<AudioCodec> {
        let available_bitrates = [64, 128, 256, 320];
        let optimal_bitrate = available_bitrates
            .iter()
            .filter(|&&bitrate| bitrate <= bandwidth_estimate * 8 / 10)
            .max()
            .copied()
            .unwrap_or(64);

        for codec in &self.config.codec_preferences {
            match codec {
                AudioCodec::Opus { .. } => {
                    return Some(AudioCodec::Opus { bitrate: optimal_bitrate });
                }
                AudioCodec::Aac { .. } => {
                    return Some(AudioCodec::Aac { bitrate: optimal_bitrate });
                }
                AudioCodec::Mp3 { .. } => {
                    return Some(AudioCodec::Mp3 { bitrate: optimal_bitrate });
                }
                _ => continue,
            }
        }

        Some(AudioCodec::Opus { bitrate: optimal_bitrate })
    }

    /// Mettre à jour les statistiques d'un peer
    pub async fn update_peer_stats(
        &self,
        peer_id: &str,
        bandwidth: u32,
        rtt: u32,
        packet_loss: f32,
        jitter: u32,
    ) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
        let mut peers = self.peers.write().await;
        
        if let Some(peer) = peers.get_mut(peer_id) {
            peer.bandwidth_estimate = bandwidth;
            peer.rtt_ms = Some(rtt);
            peer.packet_loss_percentage = packet_loss;
            peer.jitter_ms = Some(jitter);
            peer.last_activity = Instant::now();

            // Envoyer message de mise à jour qualité
            let quality_msg = WebRTCMessage::QualityUpdate {
                peer_id: peer_id.to_string(),
                bandwidth,
                rtt,
                packet_loss,
            };

            if let Err(e) = self.signaling_tx.send(quality_msg) {
                warn!("Failed to send quality update: {}", e);
            }
        }

        Ok(())
    }

    /// Obtenir les statistiques en temps réel
    pub async fn get_real_time_stats(&self) -> serde_json::Value {
        let peers = self.peers.read().await;
        let peer_count = peers.len();
        let connected_peers = peers.values()
            .filter(|p| matches!(p.connection_state, ConnectionState::Connected))
            .count();

        let total_bandwidth: u32 = peers.values()
            .map(|p| p.bandwidth_estimate)
            .sum();

        let avg_rtt: f32 = {
            let rtts: Vec<u32> = peers.values()
                .filter_map(|p| p.rtt_ms)
                .collect();
            if rtts.is_empty() {
                0.0
            } else {
                rtts.iter().sum::<u32>() as f32 / rtts.len() as f32
            }
        };

        let avg_packet_loss: f32 = {
            let losses: Vec<f32> = peers.values()
                .map(|p| p.packet_loss_percentage)
                .collect();
            if losses.is_empty() {
                0.0
            } else {
                losses.iter().sum::<f32>() / losses.len() as f32
            }
        };

        serde_json::json!({
            "webrtc_stats": {
                "total_peers": peer_count,
                "connected_peers": connected_peers,
                "total_bandwidth_kbps": total_bandwidth,
                "average_rtt_ms": avg_rtt,
                "average_packet_loss_percent": avg_packet_loss,
                "max_peers": self.config.max_peers,
                "codec_preferences": self.config.codec_preferences,
                "bitrate_adaptation_enabled": self.config.bitrate_adaptation
            }
        })
    }

    /// Démarrer le moniteur de connexions
    async fn start_connection_monitor(&self) {
        let peers = self.peers.clone();
        let timeout = self.config.connection_timeout;
        let heartbeat_interval = self.config.heartbeat_interval;

        tokio::spawn(async move {
            let mut interval = tokio::time::interval(heartbeat_interval);
            
            loop {
                interval.tick().await;
                
                let mut peers_to_remove = Vec::new();
                {
                    let peers_guard = peers.read().await;
                    let now = Instant::now();
                    
                    for (peer_id, peer) in peers_guard.iter() {
                        if now.duration_since(peer.last_activity) > timeout {
                            peers_to_remove.push(peer_id.clone());
                        }
                    }
                }

                if !peers_to_remove.is_empty() {
                    let mut peers_guard = peers.write().await;
                    for peer_id in peers_to_remove {
                        if peers_guard.remove(&peer_id).is_some() {
                            warn!("Removed inactive WebRTC peer: {}", peer_id);
                        }
                    }
                }
            }
        });
    }

    /// Démarrer l'adaptation automatique de bitrate
    async fn start_bitrate_adaptation(&self) {
        let peers = self.peers.clone();
        let signaling_tx = self.signaling_tx.clone();

        tokio::spawn(async move {
            let mut interval = tokio::time::interval(Duration::from_secs(5));
            
            loop {
                interval.tick().await;
                
                let peers_guard = peers.read().await;
                for (peer_id, peer) in peers_guard.iter() {
                    // Adapter le bitrate selon les conditions réseau
                    let current_bitrate = peer.selected_codec
                        .as_ref()
                        .map(|c| c.get_bitrate())
                        .unwrap_or(128);

                    let optimal_bitrate = if peer.packet_loss_percentage > 5.0 {
                        // Réduire le bitrate si perte de paquets élevée
                        std::cmp::max(64, current_bitrate - 64)
                    } else if peer.bandwidth_estimate > current_bitrate * 12 / 10 {
                        // Augmenter le bitrate si bande passante suffisante
                        std::cmp::min(320, current_bitrate + 64)
                    } else {
                        current_bitrate
                    };

                    if optimal_bitrate != current_bitrate {
                        let msg = WebRTCMessage::BitrateChange {
                            peer_id: peer_id.clone(),
                            new_bitrate: optimal_bitrate,
                        };

                        if let Err(e) = signaling_tx.send(msg) {
                            warn!("Failed to send bitrate change: {}", e);
                        }
                    }
                }
            }
        });
    }

    /// Démarrer le collecteur de statistiques
    async fn start_stats_collector(&self) {
        let peers = self.peers.clone();

        tokio::spawn(async move {
            let mut interval = tokio::time::interval(Duration::from_secs(1));
            
            loop {
                interval.tick().await;
                
                let peers_guard = peers.read().await;
                let connected_count = peers_guard.values()
                    .filter(|p| matches!(p.connection_state, ConnectionState::Connected))
                    .count();

                if connected_count > 0 {
                    debug!("WebRTC active connections: {}", connected_count);
                }
            }
        });
    }

    /// Supprimer un peer
    pub async fn remove_peer(&self, peer_id: &str) -> bool {
        let mut peers = self.peers.write().await;
        if let Some(_peer) = peers.remove(peer_id) {
            info!("Removed WebRTC peer: {}", peer_id);
            
            let disconnect_msg = WebRTCMessage::PeerDisconnected {
                peer_id: peer_id.to_string(),
            };

            if let Err(e) = self.signaling_tx.send(disconnect_msg) {
                warn!("Failed to send peer disconnect message: {}", e);
            }
            
            true
        } else {
            false
        }
    }

    /// Obtenir un receiver pour les messages de signaling
    pub fn get_signaling_receiver(&self) -> broadcast::Receiver<WebRTCMessage> {
        self.signaling_tx.subscribe()
    }

    /// Envoyer un message de signaling
    pub async fn send_signaling_message(&self, message: WebRTCMessage) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
        self.signaling_tx.send(message)?;
        Ok(())
    }
}
