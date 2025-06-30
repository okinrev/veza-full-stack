// Sync Manager module for Phase 5

use std::collections::HashMap;
use std::sync::Arc;
use std::time::{Duration, Instant, SystemTime};
use tokio::sync::{RwLock, broadcast, mpsc};
use serde::{Deserialize, Serialize};
use tracing::{info, debug, warn, error, span, Level};
use uuid::Uuid;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SyncConfig {
    pub sync_tolerance_ms: u32,
    pub max_clients: usize,
    pub clock_sync_interval: Duration,
    pub buffer_target_ms: u32,
    pub buffer_min_ms: u32,
    pub buffer_max_ms: u32,
    pub jitter_correction: bool,
    pub adaptive_buffering: bool,
}

impl Default for SyncConfig {
    fn default() -> Self {
        Self {
            sync_tolerance_ms: 100, // < 100ms selon objectifs Phase 5
            max_clients: 1000,       // Support 1000 listeners simultanés
            clock_sync_interval: Duration::from_secs(30),
            buffer_target_ms: 200,
            buffer_min_ms: 100,
            buffer_max_ms: 500,
            jitter_correction: true,
            adaptive_buffering: true,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SynchronizedClient {
    pub client_id: String,
    pub session_id: String,
    pub sync_state: SyncState,
    pub clock_offset_ms: i64,
    pub buffer_level_ms: u32,
    pub target_buffer_ms: u32,
    pub jitter_ms: u32,
    #[serde(skip, default = "default_instant")]
    pub last_sync: Instant,
    pub connection_quality: ConnectionQuality,
    pub sync_metrics: SyncMetrics,
    pub created_at: SystemTime,
}

fn default_instant() -> Instant {
    Instant::now()
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum SyncState {
    Initializing,
    Syncing,
    Synchronized,
    Buffering,
    Catching,
    Disconnected,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ConnectionQuality {
    pub rtt_ms: u32,
    pub jitter_ms: u32,
    pub packet_loss_percent: f32,
    pub bandwidth_kbps: u32,
    pub stability_score: f32, // 0.0 - 1.0
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SyncMetrics {
    pub sync_events: u64,
    pub buffer_underruns: u64,
    pub buffer_overruns: u64,
    pub clock_corrections: u64,
    pub quality_adjustments: u64,
    pub total_drift_ms: i64,
    pub avg_sync_accuracy_ms: f32,
}

impl Default for SyncMetrics {
    fn default() -> Self {
        Self {
            sync_events: 0,
            buffer_underruns: 0,
            buffer_overruns: 0,
            clock_corrections: 0,
            quality_adjustments: 0,
            total_drift_ms: 0,
            avg_sync_accuracy_ms: 0.0,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(tag = "type", content = "data")]
pub enum SyncMessage {
    ClockSync {
        client_id: String,
        server_timestamp: u64,
        client_timestamp: u64,
        round_trip_time: u32,
    },
    BufferAdjustment {
        client_id: String,
        target_buffer_ms: u32,
        current_buffer_ms: u32,
    },
    SyncCommand {
        client_id: String,
        command: SyncCommandType,
        timestamp: u64,
    },
    QualityUpdate {
        client_id: String,
        connection_quality: ConnectionQuality,
    },
    SyncStatus {
        client_id: String,
        sync_state: SyncState,
        accuracy_ms: f32,
    },
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum SyncCommandType {
    StartSync,
    PauseSync,
    ResetBuffer,
    ForceSync,
    AdjustClock,
}

/// Gestionnaire de synchronisation multi-clients
#[derive(Clone)]
pub struct SyncManager {
    config: SyncConfig,
    clients: Arc<RwLock<HashMap<String, SynchronizedClient>>>,
    sync_tx: broadcast::Sender<SyncMessage>,
    master_clock: Arc<RwLock<MasterClock>>,
    stream_position: Arc<RwLock<StreamPosition>>,
}

#[derive(Debug, Clone)]
pub struct MasterClock {
    pub start_time: Instant,
    pub current_position_ms: u64,
    pub playback_rate: f64,
    pub is_playing: bool,
}

#[derive(Debug, Clone)]
pub struct StreamPosition {
    pub position_ms: u64,
    pub timestamp: Instant,
    pub drift_correction: i64,
}

impl SyncManager {
    pub fn new(config: SyncConfig) -> Self {
        let (sync_tx, _) = broadcast::channel(1000);
        
        let master_clock = MasterClock {
            start_time: Instant::now(),
            current_position_ms: 0,
            playback_rate: 1.0,
            is_playing: false,
        };

        let stream_position = StreamPosition {
            position_ms: 0,
            timestamp: Instant::now(),
            drift_correction: 0,
        };

        Self {
            config,
            clients: Arc::new(RwLock::new(HashMap::new())),
            sync_tx,
            master_clock: Arc::new(RwLock::new(master_clock)),
            stream_position: Arc::new(RwLock::new(stream_position)),
        }
    }

    /// Démarrer le gestionnaire de synchronisation
    pub async fn start(&self) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
        info!("Starting Sync Manager with max {} clients", self.config.max_clients);

        // Démarrer l'horloge maître
        self.start_master_clock().await;

        // Démarrer la synchronisation périodique
        self.start_periodic_sync().await;

        // Démarrer le moniteur de qualité
        self.start_quality_monitor().await;

        // Démarrer l'adaptation de buffer
        if self.config.adaptive_buffering {
            self.start_adaptive_buffering().await;
        }

        Ok(())
    }

    /// Ajouter un nouveau client à synchroniser
    pub async fn add_client(
        &self,
        client_id: String,
        session_id: String,
    ) -> Result<SynchronizedClient, Box<dyn std::error::Error + Send + Sync>> {
        let span = span!(Level::INFO, "add_client", client_id = %client_id);
        let _enter = span.enter();

        let mut clients = self.clients.write().await;
        
        if clients.len() >= self.config.max_clients {
            return Err("Maximum number of synchronized clients reached".into());
        }

        let client = SynchronizedClient {
            client_id: client_id.clone(),
            session_id: session_id.clone(),
            sync_state: SyncState::Initializing,
            clock_offset_ms: 0,
            buffer_level_ms: self.config.buffer_target_ms,
            target_buffer_ms: self.config.buffer_target_ms,
            jitter_ms: 0,
            last_sync: Instant::now(),
            connection_quality: ConnectionQuality {
                rtt_ms: 50,
                jitter_ms: 10,
                packet_loss_percent: 0.0,
                bandwidth_kbps: 1000,
                stability_score: 1.0,
            },
            sync_metrics: SyncMetrics::default(),
            created_at: SystemTime::now(),
        };

        clients.insert(client_id.clone(), client.clone());

        info!("Added synchronized client: {} for session: {}", client_id, session_id);

        // Initier la synchronisation initiale
        self.initiate_client_sync(&client_id).await?;

        Ok(client)
    }

    /// Initier la synchronisation pour un client
    async fn initiate_client_sync(&self, client_id: &str) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
        let master_clock = self.master_clock.read().await;
        let current_time = Instant::now();
        let server_timestamp = current_time.duration_since(master_clock.start_time).as_millis() as u64;

        let sync_msg = SyncMessage::ClockSync {
            client_id: client_id.to_string(),
            server_timestamp,
            client_timestamp: 0, // Will be set by client
            round_trip_time: 0,  // Will be calculated
        };

        self.sync_tx.send(sync_msg)?;
        Ok(())
    }

    /// Traiter une réponse de synchronisation d'horloge
    pub async fn process_clock_sync_response(
        &self,
        client_id: &str,
        client_timestamp: u64,
        server_timestamp: u64,
    ) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
        let mut clients = self.clients.write().await;
        
        if let Some(client) = clients.get_mut(client_id) {
            let current_time = Instant::now();
            let master_clock = self.master_clock.read().await;
            let current_server_time = current_time.duration_since(master_clock.start_time).as_millis() as u64;
            
            // Calculer RTT et offset d'horloge
            let rtt = (current_server_time - server_timestamp) as u32;
            let clock_offset = (client_timestamp as i64) - (server_timestamp as i64) - (rtt as i64 / 2);
            
            client.clock_offset_ms = clock_offset;
            client.connection_quality.rtt_ms = rtt;
            client.last_sync = current_time;
            client.sync_metrics.sync_events += 1;
            client.sync_metrics.total_drift_ms += clock_offset.abs();
            client.sync_metrics.avg_sync_accuracy_ms = 
                client.sync_metrics.total_drift_ms as f32 / client.sync_metrics.sync_events as f32;

            // Mettre à jour l'état de synchronisation
            client.sync_state = if clock_offset.abs() <= self.config.sync_tolerance_ms as i64 {
                SyncState::Synchronized
            } else {
                SyncState::Syncing
            };

            info!("Clock sync for client {}: offset={}ms, rtt={}ms, state={:?}", 
                  client_id, clock_offset, rtt, client.sync_state);

            // Envoyer les ajustements si nécessaire
            if clock_offset.abs() > self.config.sync_tolerance_ms as i64 {
                self.send_sync_adjustment(client_id, clock_offset).await?;
            }
        }

        Ok(())
    }

    /// Envoyer un ajustement de synchronisation
    async fn send_sync_adjustment(
        &self,
        client_id: &str,
        clock_offset: i64,
    ) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
        let adjustment_msg = SyncMessage::SyncCommand {
            client_id: client_id.to_string(),
            command: SyncCommandType::AdjustClock,
            timestamp: clock_offset as u64,
        };

        self.sync_tx.send(adjustment_msg)?;
        Ok(())
    }

    /// Obtenir les statistiques de synchronisation en temps réel
    pub async fn get_sync_stats(&self) -> serde_json::Value {
        let clients = self.clients.read().await;
        let master_clock = self.master_clock.read().await;
        
        let total_clients = clients.len();
        let synchronized_clients = clients.values()
            .filter(|c| matches!(c.sync_state, SyncState::Synchronized))
            .count();

        let avg_sync_accuracy: f32 = if total_clients > 0 {
            clients.values()
                .map(|c| c.sync_metrics.avg_sync_accuracy_ms)
                .sum::<f32>() / total_clients as f32
        } else {
            0.0
        };

        let avg_rtt: f32 = if total_clients > 0 {
            clients.values()
                .map(|c| c.connection_quality.rtt_ms as f32)
                .sum::<f32>() / total_clients as f32
        } else {
            0.0
        };

        serde_json::json!({
            "sync_stats": {
                "total_clients": total_clients,
                "synchronized_clients": synchronized_clients,
                "sync_ratio": if total_clients > 0 { synchronized_clients as f32 / total_clients as f32 } else { 0.0 },
                "average_sync_accuracy_ms": avg_sync_accuracy,
                "average_rtt_ms": avg_rtt,
                "master_clock_position_ms": master_clock.current_position_ms,
                "master_clock_playing": master_clock.is_playing,
                "sync_tolerance_ms": self.config.sync_tolerance_ms,
                "target_accuracy_achieved": avg_sync_accuracy <= self.config.sync_tolerance_ms as f32
            }
        })
    }

    /// Démarrer l'horloge maître
    async fn start_master_clock(&self) {
        let master_clock = self.master_clock.clone();
        let stream_position = self.stream_position.clone();

        tokio::spawn(async move {
            let mut interval = tokio::time::interval(Duration::from_millis(10));
            
            loop {
                interval.tick().await;
                
                let mut clock = master_clock.write().await;
                let mut position = stream_position.write().await;
                
                if clock.is_playing {
                    let elapsed = clock.start_time.elapsed().as_millis() as u64;
                    clock.current_position_ms = (elapsed as f64 * clock.playback_rate) as u64;
                    
                    position.position_ms = clock.current_position_ms;
                    position.timestamp = Instant::now();
                }
            }
        });
    }

    /// Démarrer la synchronisation périodique
    async fn start_periodic_sync(&self) {
        let clients = self.clients.clone();
        let sync_tx = self.sync_tx.clone();
        let interval = self.config.clock_sync_interval;

        tokio::spawn(async move {
            let mut timer = tokio::time::interval(interval);
            
            loop {
                timer.tick().await;
                
                let clients_guard = clients.read().await;
                for client_id in clients_guard.keys() {
                    let sync_msg = SyncMessage::SyncCommand {
                        client_id: client_id.clone(),
                        command: SyncCommandType::StartSync,
                        timestamp: Instant::now().elapsed().as_millis() as u64,
                    };

                    if let Err(e) = sync_tx.send(sync_msg) {
                        warn!("Failed to send periodic sync: {}", e);
                    }
                }
            }
        });
    }

    /// Démarrer le moniteur de qualité
    async fn start_quality_monitor(&self) {
        let clients = self.clients.clone();

        tokio::spawn(async move {
            let mut interval = tokio::time::interval(Duration::from_secs(5));
            
            loop {
                interval.tick().await;
                
                let clients_guard = clients.read().await;
                for (client_id, client) in clients_guard.iter() {
                    if client.last_sync.elapsed() > Duration::from_secs(60) {
                        warn!("Client {} sync timeout", client_id);
                    }
                }
            }
        });
    }

    /// Démarrer l'adaptation de buffer
    async fn start_adaptive_buffering(&self) {
        let clients = self.clients.clone();

        tokio::spawn(async move {
            let mut interval = tokio::time::interval(Duration::from_secs(10));
            
            loop {
                interval.tick().await;
                
                let mut clients_guard = clients.write().await;
                for (_client_id, client) in clients_guard.iter_mut() {
                    // Ajuster le buffer selon les performances
                    if client.sync_metrics.buffer_underruns > client.sync_metrics.buffer_overruns {
                        // Plus de underruns = augmenter le buffer
                        let new_target = std::cmp::min(
                            client.target_buffer_ms + 50,
                            500, // Max buffer
                        );
                        client.target_buffer_ms = new_target;
                    } else if client.sync_metrics.buffer_overruns > 0 {
                        // Des overruns = réduire le buffer
                        let new_target = std::cmp::max(
                            client.target_buffer_ms.saturating_sub(25),
                            100, // Min buffer
                        );
                        client.target_buffer_ms = new_target;
                    }
                }
            }
        });
    }

    /// Obtenir un receiver pour les messages de synchronisation
    pub fn get_sync_receiver(&self) -> broadcast::Receiver<SyncMessage> {
        self.sync_tx.subscribe()
    }

    /// Supprimer un client
    pub async fn remove_client(&self, client_id: &str) -> bool {
        let mut clients = self.clients.write().await;
        if clients.remove(client_id).is_some() {
            info!("Removed synchronized client: {}", client_id);
            true
        } else {
            false
        }
    }
}
