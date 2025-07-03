/// Module de gestion des buffers adaptatifs pour streaming production
/// 
/// Features :
/// - Adaptive buffering selon bande passante
/// - Prédiction intelligente des besoins
/// - Gestion des interruptions réseau
/// - Optimisation mémoire avec pools

use std::sync::Arc;
use std::time::{Duration, Instant};
use std::collections::{VecDeque, HashMap};

use parking_lot::RwLock;
use dashmap::DashMap;
use tokio::sync::Mutex;
use uuid::Uuid;
use serde::{Serialize, Deserialize};
use tracing::{info, debug};

use crate::error::AppError;

/// Gestionnaire principal des buffers adaptatifs
#[derive(Debug)]
pub struct BufferManager {
    /// Buffers actifs indexés par stream_id
    buffers: Arc<DashMap<Uuid, Arc<AdaptiveBuffer>>>,
    /// Pool de chunks réutilisables pour optimiser la mémoire
    chunk_pool: Arc<Mutex<Vec<AudioChunk>>>,
    /// Configuration globale
    config: Arc<RwLock<BufferConfig>>,
    /// Métriques de performance
    metrics: Arc<BufferMetrics>,
    /// Analyseur de bande passante pour ajustements
    _bandwidth_analyzer: Arc<BandwidthAnalyzer>,
}

/// Buffer adaptatif pour un stream spécifique
#[derive(Debug)]
pub struct AdaptiveBuffer {
    pub stream_id: Uuid,
    /// Buffer principal avec chunks ordonnés
    buffer: Arc<RwLock<VecDeque<AudioChunk>>>,
    /// Taille cible du buffer (adaptatif)
    target_size: Arc<std::sync::atomic::AtomicUsize>,
    /// Taille maximale autorisée
    max_size: usize,
    /// État actuel du buffer
    status: Arc<RwLock<BufferStatus>>,
    /// Statistiques temps réel
    stats: Arc<RwLock<BufferStats>>,
    /// Configuration spécifique
    config: BufferStreamConfig,
    /// Prédicteur de besoins
    predictor: Arc<BufferPredictor>,
}

/// Chunk audio avec métadonnées
#[derive(Debug, Clone)]
pub struct AudioChunk {
    pub id: Uuid,
    pub stream_id: Uuid,
    pub sequence_number: u64,
    pub data: Arc<Vec<u8>>,
    pub format: AudioFormat,
    pub timestamp: Instant,
    pub duration: Duration,
    pub size_bytes: usize,
    pub quality_level: String,
    pub compression_ratio: f32,
}

/// Format audio pour les chunks
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct AudioFormat {
    pub codec: String,
    pub bitrate: u32,
    pub sample_rate: u32,
    pub channels: u8,
    pub bit_depth: u8,
}

/// État du buffer
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum BufferStatus {
    /// Buffer vide, en attente de données
    Empty,
    /// Remplissage initial
    Filling { progress: f32 },
    /// État optimal pour streaming
    Optimal,
    /// Buffer plein, données peuvent être perdues
    Full,
    /// Sous-remplissage, risque d'interruption
    Underrun { severity: UnderrunSeverity },
    /// Sur-remplissage, latence élevée
    Overrun { latency_ms: u32 },
    /// Erreur critique
    Error { message: String },
}

/// Sévérité des sous-remplissages
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum UnderrunSeverity {
    Low,     // 80-90% du target
    Medium,  // 60-80% du target
    High,    // 40-60% du target
    Critical, // <40% du target
}

/// Statistiques du buffer
#[derive(Debug, Clone, Default)]
pub struct BufferStats {
    pub current_size: usize,
    pub peak_size: usize,
    pub average_size: f32,
    pub fill_rate: f32,      // chunks/sec
    pub drain_rate: f32,     // chunks/sec
    pub underruns_count: u32,
    pub overruns_count: u32,
    pub total_chunks_processed: u64,
    pub average_chunk_size: usize,
    pub memory_usage_mb: f32,
    pub latency_ms: u32,
}

/// Configuration globale des buffers
#[derive(Debug, Clone)]
pub struct BufferConfig {
    pub default_target_size: usize,
    pub max_buffer_size: usize,
    pub chunk_pool_size: usize,
    pub enable_adaptive_sizing: bool,
    pub enable_prediction: bool,
    pub memory_limit_mb: usize,
    pub cleanup_interval: Duration,
    pub metrics_update_interval: Duration,
}

/// Configuration spécifique à un stream
#[derive(Debug, Clone)]
pub struct BufferStreamConfig {
    pub initial_target_size: usize,
    pub min_target_size: usize,
    pub max_target_size: usize,
    pub adaptation_speed: f32,
    pub quality_priorities: Vec<String>,
    pub enable_quality_switching: bool,
    pub preload_duration: Duration,
}

/// Métriques globales des buffers
#[derive(Debug, Default)]
pub struct BufferMetrics {
    pub total_buffers_active: std::sync::atomic::AtomicUsize,
    pub total_memory_usage_mb: std::sync::atomic::AtomicU64,
    pub total_chunks_processed: std::sync::atomic::AtomicU64,
    pub average_latency_ms: std::sync::atomic::AtomicU32,
    pub underruns_per_minute: std::sync::atomic::AtomicU32,
    pub adaptation_events: std::sync::atomic::AtomicU64,
}

/// Analyseur de bande passante et prédicteur
#[derive(Debug)]
pub struct BandwidthAnalyzer {
    /// Historique des mesures de bande passante
    measurements: Arc<RwLock<VecDeque<BandwidthMeasurement>>>,
    /// Prédictions calculées
    predictions: Arc<RwLock<HashMap<Uuid, BandwidthPrediction>>>,
    /// Configuration de l'analyseur
    config: BandwidthAnalyzerConfig,
}

/// Mesure de bande passante
#[derive(Debug, Clone)]
pub struct BandwidthMeasurement {
    pub timestamp: Instant,
    pub stream_id: Uuid,
    pub available_bandwidth: u32, // bits/sec
    pub used_bandwidth: u32,
    pub latency_ms: u32,
    pub packet_loss: f32,
    pub jitter_ms: f32,
}

/// Prédiction de bande passante
#[derive(Debug, Clone)]
pub struct BandwidthPrediction {
    pub predicted_bandwidth: u32,
    pub confidence: f32,
    pub recommended_target_size: usize,
    pub quality_recommendation: String,
    pub next_update: Instant,
}

/// Configuration de l'analyseur de bande passante
#[derive(Debug, Clone)]
pub struct BandwidthAnalyzerConfig {
    pub measurement_window: Duration,
    pub prediction_horizon: Duration,
    pub update_interval: Duration,
    pub smoothing_factor: f32,
}

/// Prédicteur de besoins du buffer
#[derive(Debug)]
pub struct BufferPredictor {
    /// Historique de consommation par stream
    consumption_history: Arc<RwLock<HashMap<Uuid, VecDeque<BufferStateSnapshot>>>>,
    /// Prédictions actives
    predictions: Arc<RwLock<HashMap<Uuid, BufferPrediction>>>,
    /// Modèle de prédiction ML
    _model: Arc<RwLock<PredictionModel>>,
    /// Configuration du prédicteur
    config: PredictorConfig,
}

/// Snapshot de l'état du buffer pour prédiction
#[derive(Debug, Clone)]
pub struct BufferStateSnapshot {
    pub timestamp: Instant,
    pub buffer_size: usize,
    pub fill_rate: f32,
    pub drain_rate: f32,
    pub bandwidth: u32,
    pub quality_level: String,
    pub listener_count: u32,
}

/// Modèle de prédiction (simple moving average + trend)
#[derive(Debug, Clone)]
pub struct PredictionModel {
    pub window_size: usize,
    pub trend_weight: f32,
    pub seasonal_component: f32,
}

/// Configuration du prédicteur
#[derive(Debug, Clone)]
pub struct PredictorConfig {
    pub history_size: usize,
    pub prediction_accuracy_threshold: f32,
    pub adaptation_threshold: f32,
    pub min_samples_for_prediction: usize,
}

impl Default for BufferConfig {
    fn default() -> Self {
        Self {
            default_target_size: 50,  // chunks
            max_buffer_size: 200,     // chunks
            chunk_pool_size: 1000,
            enable_adaptive_sizing: true,
            enable_prediction: true,
            memory_limit_mb: 1024,    // 1GB
            cleanup_interval: Duration::from_secs(60),
            metrics_update_interval: Duration::from_secs(5),
        }
    }
}

impl Default for AudioFormat {
    fn default() -> Self {
        Self {
            codec: "opus".to_string(),
            bitrate: 128_000,
            sample_rate: 44100,
            channels: 2,
            bit_depth: 16,
        }
    }
}

impl BufferManager {
    /// Crée un nouveau gestionnaire de buffers
    pub fn new() -> Self {
        Self {
            buffers: Arc::new(DashMap::new()),
            chunk_pool: Arc::new(Mutex::new(Vec::new())),
            config: Arc::new(RwLock::new(BufferConfig::default())),
            metrics: Arc::new(BufferMetrics::default()),
            _bandwidth_analyzer: Arc::new(BandwidthAnalyzer::new()),
        }
    }
    
    /// Crée un buffer adaptatif pour un stream
    pub async fn create_buffer(&self, stream_id: Uuid) -> Result<Arc<AdaptiveBuffer>, AppError> {
        let config = self.config.read();
        
        let stream_config = BufferStreamConfig {
            initial_target_size: config.default_target_size,
            min_target_size: config.default_target_size / 4,
            max_target_size: config.max_buffer_size,
            adaptation_speed: 0.1,
            quality_priorities: vec!["medium".to_string(), "high".to_string(), "low".to_string()],
            enable_quality_switching: true,
            preload_duration: Duration::from_millis(2000),
        };
        
        let buffer = Arc::new(AdaptiveBuffer {
            stream_id,
            buffer: Arc::new(RwLock::new(VecDeque::new())),
            target_size: Arc::new(std::sync::atomic::AtomicUsize::new(config.default_target_size)),
            max_size: config.max_buffer_size,
            status: Arc::new(RwLock::new(BufferStatus::Empty)),
            stats: Arc::new(RwLock::new(BufferStats::default())),
            config: stream_config,
            predictor: Arc::new(BufferPredictor::new()),
        });
        
        self.buffers.insert(stream_id, buffer.clone());
        self.metrics.total_buffers_active.fetch_add(1, std::sync::atomic::Ordering::Relaxed);
        
        info!("Buffer adaptatif créé pour stream: {}", stream_id);
        Ok(buffer)
    }
    
    /// Supprime un buffer
    pub async fn remove_buffer(&self, stream_id: Uuid) -> Result<(), AppError> {
        if let Some((_, buffer)) = self.buffers.remove(&stream_id) {
            // Libérer les chunks dans le pool
            let chunks = buffer.buffer.read().clone();
            self.return_chunks_to_pool(chunks.into_iter().collect()).await;
            
            self.metrics.total_buffers_active.fetch_sub(1, std::sync::atomic::Ordering::Relaxed);
            info!("Buffer supprimé pour stream: {}", stream_id);
        }
        Ok(())
    }
    
    /// Ajoute un chunk à un buffer
    pub async fn add_chunk(&self, chunk: AudioChunk) -> Result<(), AppError> {
        let buffer = self.buffers.get(&chunk.stream_id)
            .ok_or_else(|| AppError::BufferNotFound { stream_id: chunk.stream_id.to_string() })?;
        
        buffer.add_chunk(chunk).await
    }
    
    /// Récupère le prochain chunk d'un buffer
    pub async fn get_next_chunk(&self, stream_id: Uuid) -> Result<Option<AudioChunk>, AppError> {
        let buffer = self.buffers.get(&stream_id)
            .ok_or_else(|| AppError::BufferNotFound { stream_id: stream_id.to_string() })?;
        
        buffer.get_next_chunk().await
    }
    
    /// Retourne des chunks au pool pour réutilisation
    async fn return_chunks_to_pool(&self, chunks: Vec<AudioChunk>) {
        let mut pool = self.chunk_pool.lock().await;
        let config = self.config.read();
        
        // Limiter la taille du pool
        let max_to_add = config.chunk_pool_size.saturating_sub(pool.len());
        let chunks_to_add = chunks.into_iter().take(max_to_add);
        
        for mut chunk in chunks_to_add {
            // Nettoyer le chunk pour réutilisation
            chunk.id = Uuid::new_v4();
            chunk.data = Arc::new(Vec::new());
            pool.push(chunk);
        }
    }
    
    /// Obtient les métriques globales
    pub fn get_metrics(&self) -> BufferMetrics {
        // Clone des métriques atomiques
        BufferMetrics {
            total_buffers_active: std::sync::atomic::AtomicUsize::new(
                self.metrics.total_buffers_active.load(std::sync::atomic::Ordering::Relaxed)
            ),
            total_memory_usage_mb: std::sync::atomic::AtomicU64::new(
                self.metrics.total_memory_usage_mb.load(std::sync::atomic::Ordering::Relaxed)
            ),
            total_chunks_processed: std::sync::atomic::AtomicU64::new(
                self.metrics.total_chunks_processed.load(std::sync::atomic::Ordering::Relaxed)
            ),
            average_latency_ms: std::sync::atomic::AtomicU32::new(
                self.metrics.average_latency_ms.load(std::sync::atomic::Ordering::Relaxed)
            ),
            underruns_per_minute: std::sync::atomic::AtomicU32::new(
                self.metrics.underruns_per_minute.load(std::sync::atomic::Ordering::Relaxed)
            ),
            adaptation_events: std::sync::atomic::AtomicU64::new(
                self.metrics.adaptation_events.load(std::sync::atomic::Ordering::Relaxed)
            ),
        }
    }
}

impl AdaptiveBuffer {
    /// Ajoute un chunk au buffer
    pub async fn add_chunk(&self, chunk: AudioChunk) -> Result<(), AppError> {
        let mut buffer = self.buffer.write();
        let target_size = self.target_size.load(std::sync::atomic::Ordering::Relaxed);
        
        // Vérifier si on dépasse la taille max
        if buffer.len() >= self.max_size {
            self.update_status(BufferStatus::Full).await;
            return Err(AppError::BufferFull { stream_id: self.stream_id.to_string() });
        }
        
        // Ajouter le chunk
        buffer.push_back(chunk);
        
        // Mettre à jour les statistiques
        self.update_stats(&buffer).await;
        
        // Adapter la taille si nécessaire
        if self.config.enable_quality_switching {
            self.adapt_buffer_size().await?;
        }
        
        // Mettre à jour le statut
        let fill_ratio = buffer.len() as f32 / target_size as f32;
        self.update_status_from_fill_ratio(fill_ratio).await;
        
        Ok(())
    }
    
    /// Récupère le prochain chunk
    pub async fn get_next_chunk(&self) -> Result<Option<AudioChunk>, AppError> {
        let mut buffer = self.buffer.write();
        let chunk = buffer.pop_front();
        
        if let Some(ref _chunk) = chunk {
            // Mettre à jour les statistiques
            self.update_stats(&buffer).await;
            
            // Vérifier les underruns
            let target_size = self.target_size.load(std::sync::atomic::Ordering::Relaxed);
            let fill_ratio = buffer.len() as f32 / target_size as f32;
            
            if fill_ratio < 0.4 {
                self.update_status(BufferStatus::Underrun { 
                    severity: UnderrunSeverity::Critical 
                }).await;
            } else if fill_ratio < 0.6 {
                self.update_status(BufferStatus::Underrun { 
                    severity: UnderrunSeverity::High 
                }).await;
            }
        }
        
        Ok(chunk)
    }
    
    /// Adapte la taille du buffer selon les conditions
    async fn adapt_buffer_size(&self) -> Result<(), AppError> {
        // Obtenir les prédictions
        let prediction = self.predictor.predict_needs().await?;
        
        let current_target = self.target_size.load(std::sync::atomic::Ordering::Relaxed);
        let new_target = self.calculate_optimal_size(prediction).await;
        
        // Appliquer l'adaptation progressive
        if new_target != current_target {
            let adapted_target = self.apply_adaptation_speed(current_target, new_target);
            self.target_size.store(adapted_target, std::sync::atomic::Ordering::Relaxed);
            
            debug!("Buffer {} adapté: {} -> {} chunks", 
                   self.stream_id, current_target, adapted_target);
        }
        
        Ok(())
    }
    
    /// Calcule la taille optimale selon les prédictions
    async fn calculate_optimal_size(&self, _prediction: BufferPrediction) -> usize {
        // Algorithme simplifié - en production, utiliser ML plus sophistiqué
        let stats = self.stats.read();
        let current_size = stats.current_size;
        let drain_rate = stats.drain_rate;
        let fill_rate = stats.fill_rate;
        
        if drain_rate > fill_rate * 1.2 {
            // Augmenter le buffer si on consomme plus qu'on reçoit
            (current_size as f32 * 1.5) as usize
        } else if fill_rate > drain_rate * 1.5 {
            // Diminuer le buffer si on reçoit beaucoup plus qu'on consomme
            (current_size as f32 * 0.8) as usize
        } else {
            current_size
        }
    }
    
    /// Applique la vitesse d'adaptation configurée
    fn apply_adaptation_speed(&self, current: usize, target: usize) -> usize {
        let speed = self.config.adaptation_speed;
        let delta = target as f32 - current as f32;
        let adapted_delta = delta * speed;
        
        let new_target = current as f32 + adapted_delta;
        let clamped = new_target
            .max(self.config.min_target_size as f32)
            .min(self.config.max_target_size as f32);
        
        clamped as usize
    }
    
    /// Met à jour les statistiques
    async fn update_stats(&self, buffer: &VecDeque<AudioChunk>) {
        let mut stats = self.stats.write();
        stats.current_size = buffer.len();
        
        if buffer.len() > stats.peak_size {
            stats.peak_size = buffer.len();
        }
        
        // Calculer la moyenne mobile
        stats.average_size = stats.average_size * 0.9 + buffer.len() as f32 * 0.1;
        
        // Calculer l'usage mémoire
        let memory_mb = buffer.iter()
            .map(|chunk| chunk.size_bytes)
            .sum::<usize>() as f32 / 1024.0 / 1024.0;
        stats.memory_usage_mb = memory_mb;
        
        stats.total_chunks_processed += 1;
    }
    
    /// Met à jour le statut selon le ratio de remplissage
    async fn update_status_from_fill_ratio(&self, fill_ratio: f32) {
        let status = match fill_ratio {
            r if r < 0.1 => BufferStatus::Empty,
            r if r < 0.5 => BufferStatus::Filling { progress: r },
            r if r >= 0.5 && r <= 0.9 => BufferStatus::Optimal,
            r if r > 0.9 => BufferStatus::Full,
            _ => BufferStatus::Optimal,
        };
        
        self.update_status(status).await;
    }
    
    /// Met à jour le statut du buffer
    async fn update_status(&self, status: BufferStatus) {
        *self.status.write() = status;
    }
    
    /// Obtient le statut actuel
    pub async fn get_status(&self) -> BufferStatus {
        self.status.read().clone()
    }
    
    /// Obtient les statistiques actuelles
    pub async fn get_stats(&self) -> BufferStats {
        self.stats.read().clone()
    }
}

impl BandwidthAnalyzer {
    /// Crée un nouveau analyseur de bande passante
    pub fn new() -> Self {
        Self {
            measurements: Arc::new(RwLock::new(VecDeque::new())),
            predictions: Arc::new(RwLock::new(HashMap::new())),
            config: BandwidthAnalyzerConfig {
                measurement_window: Duration::from_secs(60),
                prediction_horizon: Duration::from_secs(30),
                update_interval: Duration::from_secs(5),
                smoothing_factor: 0.8,
            },
        }
    }
    
    /// Ajoute une mesure de bande passante
    pub async fn add_measurement(&self, measurement: BandwidthMeasurement) {
        let mut measurements = self.measurements.write();
        measurements.push_back(measurement.clone());
        
        // Limiter la taille de l'historique
        let cutoff_time = Instant::now() - self.config.measurement_window;
        while let Some(front) = measurements.front() {
            if front.timestamp < cutoff_time {
                measurements.pop_front();
            } else {
                break;
            }
        }
        
        // Mettre à jour les prédictions
        self.update_prediction(measurement.stream_id).await;
    }
    
    /// Met à jour la prédiction pour un stream
    async fn update_prediction(&self, stream_id: Uuid) {
        let measurements = self.measurements.read();
        let stream_measurements: Vec<_> = measurements
            .iter()
            .filter(|m| m.stream_id == stream_id)
            .collect();
        
        if stream_measurements.len() < 3 {
            return; // Pas assez de données
        }
        
        // Calcul simple de prédiction (moyenne pondérée)
        let total_weight: f32 = stream_measurements.len() as f32;
        let predicted_bandwidth = stream_measurements
            .iter()
            .enumerate()
            .map(|(i, m)| {
                let weight = (i + 1) as f32 / total_weight;
                m.available_bandwidth as f32 * weight
            })
            .sum::<f32>() as u32;
        
        let confidence = (stream_measurements.len() as f32 / 10.0).min(1.0);
        
        let prediction = BandwidthPrediction {
            predicted_bandwidth,
            confidence,
            recommended_target_size: self.calculate_recommended_buffer_size(predicted_bandwidth),
            quality_recommendation: self.recommend_quality(predicted_bandwidth),
            next_update: Instant::now() + self.config.update_interval,
        };
        
        self.predictions.write().insert(stream_id, prediction);
    }
    
    /// Calcule la taille de buffer recommandée selon la bande passante
    fn calculate_recommended_buffer_size(&self, bandwidth: u32) -> usize {
        match bandwidth {
            0..=64_000 => 100,       // Faible bande passante = gros buffer
            64_001..=256_000 => 75,  // Moyen
            256_001..=1_000_000 => 50, // Bon
            _ => 25,                 // Excellent = petit buffer
        }
    }
    
    /// Recommande une qualité selon la bande passante
    fn recommend_quality(&self, bandwidth: u32) -> String {
        match bandwidth {
            0..=64_000 => "low".to_string(),
            64_001..=128_000 => "medium".to_string(),
            128_001..=256_000 => "high".to_string(),
            _ => "lossless".to_string(),
        }
    }
    
    /// Obtient la prédiction pour un stream
    pub async fn get_prediction(&self, stream_id: Uuid) -> Option<BandwidthPrediction> {
        self.predictions.read().get(&stream_id).cloned()
    }
}

impl BufferPredictor {
    /// Crée un nouveau prédicteur
    pub fn new() -> Self {
        Self {
            consumption_history: Arc::new(RwLock::new(HashMap::new())),
            predictions: Arc::new(RwLock::new(HashMap::new())),
            _model: Arc::new(RwLock::new(PredictionModel {
                window_size: 10,
                trend_weight: 0.3,
                seasonal_component: 0.1,
            })),
            config: PredictorConfig {
                history_size: 100,
                prediction_accuracy_threshold: 0.8,
                adaptation_threshold: 0.1,
                min_samples_for_prediction: 5,
            },
        }
    }
    
    /// Prédit les besoins futurs du buffer
    pub async fn predict_needs(&self) -> Result<BufferPrediction, AppError> {
        let history = self.consumption_history.read();
        
        if history.len() < self.config.min_samples_for_prediction {
            return Err(AppError::InsufficientData);
        }
        
        // Collecte de toutes les snapshots de tous les streams
        let mut all_snapshots = Vec::new();
        for snapshots in history.values() {
            all_snapshots.extend(snapshots.iter().cloned());
        }
        
        // Trier par timestamp et prendre les 5 plus récents
        all_snapshots.sort_by_key(|s| s.timestamp);
        let recent: Vec<_> = all_snapshots.into_iter().rev().take(5).collect();
        
        if recent.is_empty() {
            return Err(AppError::InsufficientData);
        }
        
        let avg_size = recent.iter().map(|s| s.buffer_size as f32).sum::<f32>() / recent.len() as f32;
        let avg_fill_rate = recent.iter().map(|s| s.fill_rate).sum::<f32>() / recent.len() as f32;
        
        Ok(BufferPrediction {
            predicted_size: avg_size as usize,
            predicted_fill_rate: avg_fill_rate,
            confidence: 0.7, // Simple heuristique
            horizon: Duration::from_secs(30),
        })
    }
}

/// Prédiction des besoins du buffer
#[derive(Debug, Clone)]
pub struct BufferPrediction {
    pub predicted_size: usize,
    pub predicted_fill_rate: f32,
    pub confidence: f32,
    pub horizon: Duration,
} 