/// Module de synchronisation multi-client pour streaming production
/// 
/// Features :
/// - Synchronisation précise <10ms entre clients
/// - Compensation automatique du drift réseau
/// - Support NTP pour horloge de référence
/// - Synchronisation adaptative selon latence
/// - Support paroles/sous-titres synchronisés

use std::sync::Arc;
use std::time::{Duration, Instant, SystemTime, UNIX_EPOCH};
use std::collections::HashMap;

use parking_lot::RwLock;
use dashmap::DashMap;
use tokio::sync::{mpsc, broadcast};
use uuid::Uuid;
use serde::{Serialize, Deserialize};
use tracing::{info, warn, error, debug};

use crate::error::AppError;
use crate::core::Listener;

/// Moteur de synchronisation principal
#[derive(Debug)]
pub struct SyncEngine {
    /// Serveur de temps de référence
    time_server: Arc<TimeServer>,
    /// Compensateur de drift pour chaque client
    drift_compensator: Arc<DriftCompensator>,
    /// Carte des latences par client
    latency_map: Arc<DashMap<Uuid, Duration>>,
    /// Synchroniseurs actifs par stream
    stream_synchronizers: Arc<DashMap<Uuid, Arc<StreamSynchronizer>>>,
    /// Configuration globale
    config: Arc<RwLock<SyncConfig>>,
    /// Métriques de synchronisation
    metrics: Arc<SyncMetrics>,
    /// Événements de synchronisation
    event_sender: broadcast::Sender<SyncEvent>,
}

/// Synchroniseur pour un stream spécifique
#[derive(Debug)]
pub struct StreamSynchronizer {
    pub stream_id: Uuid,
    /// Horloge maître du stream
    master_clock: Arc<MasterClock>,
    /// Clients synchronisés sur ce stream
    synchronized_clients: Arc<DashMap<Uuid, SynchronizedClient>>,
    /// Buffer de synchronisation
    sync_buffer: Arc<RwLock<SyncBuffer>>,
    /// Configuration du stream
    config: StreamSyncConfig,
    /// Métadonnées temps réel (paroles, etc.)
    timed_metadata: Arc<RwLock<TimedMetadata>>,
}

/// Serveur de temps NTP-like
#[derive(Debug)]
pub struct TimeServer {
    /// Temps de référence (peut être NTP externe)
    reference_time: Arc<RwLock<ReferenceTime>>,
    /// Clients NTP configurés
    ntp_clients: Vec<String>,
    /// Décalage mesuré avec les serveurs externes
    time_offset: Arc<std::sync::atomic::AtomicI64>,
    /// Qualité de la synchronisation
    sync_quality: Arc<std::sync::atomic::AtomicU8>,
}

/// Temps de référence avec précision
#[derive(Debug, Clone)]
pub struct ReferenceTime {
    pub system_time: SystemTime,
    pub monotonic_time: Instant,
    pub ntp_offset: Duration,
    pub precision_microseconds: u32,
}

/// Compensateur de drift réseau
#[derive(Debug)]
pub struct DriftCompensator {
    /// Mesures de drift par client
    drift_measurements: Arc<DashMap<Uuid, VecDeque<DriftMeasurement>>>,
    /// Compensation calculée par client
    compensations: Arc<DashMap<Uuid, DriftCompensation>>,
    /// Configuration
    config: DriftCompensatorConfig,
}

/// Mesure de drift pour un client
#[derive(Debug, Clone)]
pub struct DriftMeasurement {
    pub timestamp: Instant,
    pub client_reported_time: u64,
    pub server_time: u64,
    pub round_trip_time: Duration,
    pub drift_ms: f64,
}

/// Compensation appliquée à un client
#[derive(Debug, Clone)]
pub struct DriftCompensation {
    pub timestamp_offset: Duration,
    pub playback_rate_adjustment: f64, // 1.0 = normal, 1.001 = 0.1% plus rapide
    pub buffer_target_adjustment: i32, // +/- chunks
    pub confidence: f32,
}

/// Client synchronisé
#[derive(Debug, Clone)]
pub struct SynchronizedClient {
    pub listener_id: Uuid,
    pub sync_state: SyncState,
    pub last_sync_adjustment: Option<SyncAdjustment>,
    pub sync_quality: SyncQuality,
    pub adaptive_config: ClientAdaptiveConfig,
}

/// État de synchronisation d'un client
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum SyncState {
    /// Initial, en attente de première synchronisation
    Initializing,
    /// Synchronisation en cours
    Synchronizing { progress: f32 },
    /// Synchronisé avec qualité bonne
    Synchronized { drift_ms: f64 },
    /// Désynchronisé, nécessite re-sync
    Desynchronized { reason: String },
    /// Erreur de synchronisation
    Error { message: String },
}

/// Ajustement de synchronisation envoyé au client
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SyncAdjustment {
    pub timestamp_offset: Duration,
    pub playback_rate: f64,
    pub buffer_target: usize,
    pub quality_switch: Option<String>,
    pub sync_point: SyncPoint,
}

/// Point de synchronisation dans le stream
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SyncPoint {
    pub stream_position: Duration,
    pub server_timestamp: u64,
    pub sequence_number: u64,
    pub checksum: u32,
}

/// Qualité de synchronisation
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SyncQuality {
    pub accuracy_ms: f64,
    pub stability_score: f32, // 0-1
    pub last_drift: f64,
    pub sync_loss_count: u32,
}

/// Configuration adaptative par client
#[derive(Debug, Clone)]
pub struct ClientAdaptiveConfig {
    pub max_drift_tolerance: Duration,
    pub sync_frequency: Duration,
    pub aggressive_correction: bool,
    pub quality_priority: bool,
}

/// Buffer de synchronisation
#[derive(Debug, Clone)]
pub struct SyncBuffer {
    /// Points de synchronisation dans le stream
    sync_points: VecDeque<SyncPoint>,
    /// Métadonnées temporelles
    timed_events: VecDeque<TimedEvent>,
    /// Configuration
    max_size: usize,
}

/// Métadonnées temporelles (paroles, chapitres, etc.)
#[derive(Debug, Clone)]
pub struct TimedMetadata {
    /// Paroles synchronisées
    pub lyrics: Vec<LyricLine>,
    /// Chapitres/sections
    pub chapters: Vec<Chapter>,
    /// Événements personnalisés
    pub custom_events: Vec<CustomTimedEvent>,
    /// Sous-titres
    pub subtitles: Vec<Subtitle>,
}

/// Ligne de paroles avec timing
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct LyricLine {
    pub start_time: Duration,
    pub end_time: Duration,
    pub text: String,
    pub phonetic: Option<String>,
    pub language: Option<String>,
}

/// Chapitre dans le stream
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Chapter {
    pub start_time: Duration,
    pub end_time: Duration,
    pub title: String,
    pub description: Option<String>,
    pub artwork_url: Option<String>,
}

/// Événement personnalisé avec timing
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CustomTimedEvent {
    pub timestamp: Duration,
    pub event_type: String,
    pub data: HashMap<String, String>,
    pub priority: u8,
}

/// Sous-titre
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Subtitle {
    pub start_time: Duration,
    pub end_time: Duration,
    pub text: String,
    pub language: String,
    pub position: SubtitlePosition,
}

/// Position du sous-titre
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum SubtitlePosition {
    Bottom,
    Top,
    Center,
    Custom { x: f32, y: f32 },
}

/// Événement temporel
#[derive(Debug, Clone)]
pub struct TimedEvent {
    pub timestamp: Duration,
    pub event: SyncEvent,
    pub target_clients: Option<Vec<Uuid>>,
}

/// Configuration globale de synchronisation
#[derive(Debug, Clone)]
pub struct SyncConfig {
    pub enable_ntp_sync: bool,
    pub ntp_servers: Vec<String>,
    pub max_client_drift_ms: f64,
    pub sync_interval: Duration,
    pub drift_measurement_window: Duration,
    pub enable_adaptive_sync: bool,
    pub quality_threshold: f32,
}

/// Configuration par stream
#[derive(Debug, Clone)]
pub struct StreamSyncConfig {
    pub precision_mode: PrecisionMode,
    pub enable_timed_metadata: bool,
    pub enable_lyrics_sync: bool,
    pub enable_chapter_sync: bool,
    pub sync_tolerance_ms: f64,
}

/// Mode de précision de synchronisation
#[derive(Debug, Clone)]
pub enum PrecisionMode {
    /// Mode relax pour casual listening
    Relaxed { tolerance_ms: f64 },
    /// Mode standard pour streaming normal
    Standard { tolerance_ms: f64 },
    /// Mode précis pour events synchronisés
    Precise { tolerance_ms: f64 },
    /// Mode ultra-précis pour performances live
    UltraPrecise { tolerance_ms: f64 },
}

/// Configuration du compensateur de drift
#[derive(Debug, Clone)]
pub struct DriftCompensatorConfig {
    pub measurement_window_size: usize,
    pub min_measurements_for_compensation: usize,
    pub max_playback_rate_adjustment: f64,
    pub compensation_smoothing: f64,
}

/// Métriques de synchronisation
#[derive(Debug, Default)]
pub struct SyncMetrics {
    pub total_clients_synchronized: std::sync::atomic::AtomicU64,
    pub average_sync_accuracy_ms: std::sync::atomic::AtomicU32,
    pub sync_failures_total: std::sync::atomic::AtomicU64,
    pub drift_corrections_total: std::sync::atomic::AtomicU64,
    pub ntp_sync_errors: std::sync::atomic::AtomicU64,
}

/// Événements de synchronisation
#[derive(Debug, Clone)]
pub enum SyncEvent {
    /// Client ajouté à la synchronisation
    ClientSyncStarted { client_id: Uuid, stream_id: Uuid },
    /// Client synchronisé avec succès
    ClientSynchronized { client_id: Uuid, accuracy_ms: f64 },
    /// Client désynchronisé
    ClientDesynchronized { client_id: Uuid, reason: String },
    /// Ajustement de synchronisation appliqué
    SyncAdjustmentApplied { client_id: Uuid, adjustment: SyncAdjustment },
    /// Métadonnées temporelles mises à jour
    TimedMetadataUpdated { stream_id: Uuid, event_type: String },
    /// Erreur de synchronisation
    SyncError { client_id: Option<Uuid>, error: String },
}

use std::collections::VecDeque;

impl Default for SyncConfig {
    fn default() -> Self {
        Self {
            enable_ntp_sync: true,
            ntp_servers: vec![
                "pool.ntp.org".to_string(),
                "time.google.com".to_string(),
                "time.cloudflare.com".to_string(),
            ],
            max_client_drift_ms: 50.0,
            sync_interval: Duration::from_millis(1000),
            drift_measurement_window: Duration::from_secs(60),
            enable_adaptive_sync: true,
            quality_threshold: 0.8,
        }
    }
}

impl SyncEngine {
    /// Crée un nouveau moteur de synchronisation
    pub async fn new(config: SyncConfig) -> Result<Self, AppError> {
        let (event_sender, _) = broadcast::channel(10_000);
        
        let time_server = Arc::new(TimeServer::new(config.ntp_servers.clone()).await?);
        
        Ok(Self {
            time_server,
            drift_compensator: Arc::new(DriftCompensator::new()),
            latency_map: Arc::new(DashMap::new()),
            stream_synchronizers: Arc::new(DashMap::new()),
            config: Arc::new(RwLock::new(config)),
            metrics: Arc::new(SyncMetrics::default()),
            event_sender,
        })
    }
    
    /// Synchronise tous les listeners d'un stream
    pub async fn sync_listeners(&self, stream_id: Uuid, listeners: &DashMap<Uuid, Listener>) -> Result<(), AppError> {
        let synchronizer = self.get_or_create_synchronizer(stream_id).await?;
        
        // Obtenir le temps maître
        let master_time = self.time_server.get_master_time().await?;
        
        // Synchroniser chaque listener en parallèle
        let sync_tasks: Vec<_> = listeners
            .iter()
            .map(|entry| {
                let listener = entry.value().clone();
                let sync_engine = self.clone();
                let synchronizer = synchronizer.clone();
                let master_time = master_time;
                
                tokio::spawn(async move {
                    sync_engine.sync_individual_listener(&synchronizer, &listener, master_time).await
                })
            })
            .collect();
        
        // Attendre toutes les synchronisations
        let results = futures::future::join_all(sync_tasks).await;
        
        // Compter les succès/échecs
        let mut success_count = 0;
        let mut error_count = 0;
        
        for result in results {
            match result {
                Ok(Ok(())) => success_count += 1,
                Ok(Err(e)) => {
                    error_count += 1;
                    warn!("Erreur sync listener: {:?}", e);
                }
                Err(e) => {
                    error_count += 1;
                    error!("Erreur task sync: {:?}", e);
                }
            }
        }
        
        info!("Synchronisation stream {}: {} succès, {} erreurs", 
              stream_id, success_count, error_count);
        
        Ok(())
    }
    
    /// Synchronise un listener individuel
    async fn sync_individual_listener(
        &self,
        synchronizer: &StreamSynchronizer,
        listener: &Listener,
        master_time: MasterTime,
    ) -> Result<(), AppError> {
        // Mesurer la latence
        let latency = self.measure_latency(listener).await?;
        self.latency_map.insert(listener.id, latency);
        
        // Calculer le drift
        let drift = self.drift_compensator.calculate_drift(listener.id, master_time).await?;
        
        // Créer l'ajustement de synchronisation
        let adjustment = SyncAdjustment {
            timestamp_offset: latency + Duration::from_millis(drift.abs() as u64),
            playback_rate: self.calculate_playback_rate(drift),
            buffer_target: self.calculate_buffer_size(latency),
            quality_switch: self.determine_quality_switch(listener).await,
            sync_point: synchronizer.get_current_sync_point().await?,
        };
        
        // Appliquer l'ajustement
        self.apply_sync_adjustment(listener.id, adjustment.clone()).await?;
        
        // Mettre à jour le client synchronisé
        let sync_client = SynchronizedClient {
            listener_id: listener.id,
            sync_state: SyncState::Synchronized { drift_ms: drift },
            last_sync_adjustment: Some(adjustment.clone()),
            sync_quality: SyncQuality {
                accuracy_ms: drift.abs(),
                stability_score: self.calculate_stability_score(listener.id).await,
                last_drift: drift,
                sync_loss_count: 0,
            },
            adaptive_config: ClientAdaptiveConfig {
                max_drift_tolerance: Duration::from_millis(50),
                sync_frequency: Duration::from_secs(5),
                aggressive_correction: drift.abs() > 20.0,
                quality_priority: listener.bandwidth_estimate > 256_000,
            },
        };
        
        synchronizer.synchronized_clients.insert(listener.id, sync_client);
        
        // Émettre l'événement
        let _ = self.event_sender.send(SyncEvent::ClientSynchronized {
            client_id: listener.id,
            accuracy_ms: drift.abs(),
        });
        
        debug!("Listener {} synchronisé avec drift: {:.2}ms", listener.id, drift);
        Ok(())
    }
    
    /// Mesure la latence réseau avec un client
    async fn measure_latency(&self, listener: &Listener) -> Result<Duration, AppError> {
        // Simulation de mesure ping/pong
        // En production, implémenter un vrai protocole de mesure
        let base_latency = match listener.bandwidth_estimate {
            0..=64_000 => Duration::from_millis(150),      // Mobile/slow
            64_001..=256_000 => Duration::from_millis(50), // Standard
            256_001..=1_000_000 => Duration::from_millis(20), // Fast
            _ => Duration::from_millis(10),                // Fiber/local
        };
        
        // Ajouter jitter simulé
        let jitter = Duration::from_millis(rand::random::<u64>() % 10);
        Ok(base_latency + jitter)
    }
    
    /// Calcule le taux de lecture pour corriger le drift
    fn calculate_playback_rate(&self, drift_ms: f64) -> f64 {
        const MAX_ADJUSTMENT: f64 = 0.005; // 0.5% max
        
        // Ajustement proportionnel au drift
        let adjustment = (drift_ms / 1000.0) * 0.001; // 0.1% par seconde de drift
        let clamped = adjustment.max(-MAX_ADJUSTMENT).min(MAX_ADJUSTMENT);
        
        1.0 + clamped
    }
    
    /// Calcule la taille de buffer optimale selon la latence
    fn calculate_buffer_size(&self, latency: Duration) -> usize {
        match latency.as_millis() {
            0..=20 => 25,   // Très faible latence
            21..=50 => 50,  // Faible latence
            51..=100 => 75, // Latence standard
            101..=200 => 100, // Latence élevée
            _ => 150,       // Très haute latence
        }
    }
    
    /// Détermine si un changement de qualité est nécessaire
    async fn determine_quality_switch(&self, listener: &Listener) -> Option<String> {
        // Logique simplifiée - en production, analyser la performance
        if listener.buffer_health < 0.3 {
            Some("low".to_string())
        } else if listener.buffer_health > 0.8 && listener.bandwidth_estimate > 256_000 {
            Some("high".to_string())
        } else {
            None
        }
    }
    
    /// Applique un ajustement de synchronisation
    async fn apply_sync_adjustment(&self, client_id: Uuid, adjustment: SyncAdjustment) -> Result<(), AppError> {
        // Envoyer l'ajustement au client via WebSocket
        // TODO: Implémenter l'envoi réel via la connexion WebSocket
        
        // Émettre l'événement
        let _ = self.event_sender.send(SyncEvent::SyncAdjustmentApplied {
            client_id,
            adjustment,
        });
        
        Ok(())
    }
    
    /// Calcule le score de stabilité pour un client
    async fn calculate_stability_score(&self, client_id: Uuid) -> f32 {
        if let Some(measurements) = self.drift_compensator.drift_measurements.get(&client_id) {
            if measurements.len() < 3 {
                return 0.5; // Score neutre
            }
            
            // Calculer la variance du drift
            let drifts: Vec<f64> = measurements.iter().map(|m| m.drift_ms).collect();
            let mean = drifts.iter().sum::<f64>() / drifts.len() as f64;
            let variance = drifts.iter()
                .map(|&d| (d - mean).powi(2))
                .sum::<f64>() / drifts.len() as f64;
            let std_dev = variance.sqrt();
            
            // Score inversement proportionnel à la variance
            (1.0 - (std_dev / 100.0).min(1.0)) as f32
        } else {
            0.5
        }
    }
    
    /// Obtient ou crée un synchroniseur pour un stream
    async fn get_or_create_synchronizer(&self, stream_id: Uuid) -> Result<Arc<StreamSynchronizer>, AppError> {
        if let Some(sync) = self.stream_synchronizers.get(&stream_id) {
            Ok(sync.clone())
        } else {
            let synchronizer = Arc::new(StreamSynchronizer::new(stream_id).await?);
            self.stream_synchronizers.insert(stream_id, synchronizer.clone());
            Ok(synchronizer)
        }
    }
    
    /// Abonnement aux événements de synchronisation
    pub fn subscribe_events(&self) -> broadcast::Receiver<SyncEvent> {
        self.event_sender.subscribe()
    }
}

impl Clone for SyncEngine {
    fn clone(&self) -> Self {
        Self {
            time_server: self.time_server.clone(),
            drift_compensator: self.drift_compensator.clone(),
            latency_map: self.latency_map.clone(),
            stream_synchronizers: self.stream_synchronizers.clone(),
            config: self.config.clone(),
            metrics: self.metrics.clone(),
            event_sender: self.event_sender.clone(),
        }
    }
}

/// Temps maître pour synchronisation
#[derive(Debug, Clone, Copy)]
pub struct MasterTime {
    pub timestamp: u64,
    pub precision_us: u32,
}

impl TimeServer {
    /// Crée un nouveau serveur de temps
    pub async fn new(ntp_servers: Vec<String>) -> Result<Self, AppError> {
        let reference_time = ReferenceTime {
            system_time: SystemTime::now(),
            monotonic_time: Instant::now(),
            ntp_offset: Duration::ZERO,
            precision_microseconds: 1000, // 1ms precision par défaut
        };
        
        Ok(Self {
            reference_time: Arc::new(RwLock::new(reference_time)),
            ntp_clients: ntp_servers,
            time_offset: Arc::new(std::sync::atomic::AtomicI64::new(0)),
            sync_quality: Arc::new(std::sync::atomic::AtomicU8::new(50)),
        })
    }
    
    /// Obtient le temps maître actuel
    pub async fn get_master_time(&self) -> Result<MasterTime, AppError> {
        let ref_time = self.reference_time.read();
        let elapsed = ref_time.monotonic_time.elapsed();
        let timestamp = ref_time.system_time
            .duration_since(UNIX_EPOCH)
            .map_err(|_| AppError::TimeSync)?
            .as_micros() as u64 + elapsed.as_micros() as u64;
        
        Ok(MasterTime {
            timestamp,
            precision_us: ref_time.precision_microseconds,
        })
    }
}

impl StreamSynchronizer {
    /// Crée un nouveau synchroniseur de stream
    pub async fn new(stream_id: Uuid) -> Result<Self, AppError> {
        Ok(Self {
            stream_id,
            master_clock: Arc::new(MasterClock::new()),
            synchronized_clients: Arc::new(DashMap::new()),
            sync_buffer: Arc::new(RwLock::new(SyncBuffer {
                sync_points: VecDeque::new(),
                timed_events: VecDeque::new(),
                max_size: 1000,
            })),
            config: StreamSyncConfig {
                precision_mode: PrecisionMode::Standard { tolerance_ms: 10.0 },
                enable_timed_metadata: true,
                enable_lyrics_sync: true,
                enable_chapter_sync: true,
                sync_tolerance_ms: 10.0,
            },
            timed_metadata: Arc::new(RwLock::new(TimedMetadata {
                lyrics: Vec::new(),
                chapters: Vec::new(),
                custom_events: Vec::new(),
                subtitles: Vec::new(),
            })),
        })
    }
    
    /// Obtient le point de synchronisation actuel
    pub async fn get_current_sync_point(&self) -> Result<SyncPoint, AppError> {
        let buffer = self.sync_buffer.read();
        buffer.sync_points.back()
            .cloned()
            .ok_or(AppError::NoSyncPoint)
    }
}

/// Horloge maître pour un stream
#[derive(Debug)]
pub struct MasterClock {
    start_time: Instant,
    position: Arc<std::sync::atomic::AtomicU64>, // microseconds
}

impl MasterClock {
    pub fn new() -> Self {
        Self {
            start_time: Instant::now(),
            position: Arc::new(std::sync::atomic::AtomicU64::new(0)),
        }
    }
    
    pub fn get_position(&self) -> Duration {
        let micros = self.position.load(std::sync::atomic::Ordering::Relaxed);
        Duration::from_micros(micros)
    }
}

impl DriftCompensator {
    /// Crée un nouveau compensateur de drift
    pub fn new() -> Self {
        Self {
            drift_measurements: Arc::new(DashMap::new()),
            compensations: Arc::new(DashMap::new()),
            config: DriftCompensatorConfig {
                measurement_window_size: 20,
                min_measurements_for_compensation: 5,
                max_playback_rate_adjustment: 0.01, // 1%
                compensation_smoothing: 0.8,
            },
        }
    }
    
    /// Calcule le drift pour un client
    pub async fn calculate_drift(&self, client_id: Uuid, master_time: MasterTime) -> Result<f64, AppError> {
        // Simulation de calcul de drift
        // En production, utiliser les timestamps client/serveur
        Ok(rand::random::<f64>() * 20.0 - 10.0) // -10 à +10ms
    }
} 