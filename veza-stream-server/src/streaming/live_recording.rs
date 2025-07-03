// Live Recording module for Phase 5

use std::collections::HashMap;
use std::sync::Arc;
use std::time::{Duration, SystemTime};
use std::path::{Path, PathBuf};
use tokio::sync::{RwLock, broadcast};
use serde::{Deserialize, Serialize};
use tracing::{info, debug, warn, span, Level};
use uuid::Uuid;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RecordingConfig {
    pub output_directory: PathBuf,
    pub max_concurrent_recordings: usize,
    pub segment_duration_ms: u32,
    pub output_formats: Vec<AudioFormat>,
    pub real_time_transcoding: bool,
    pub metadata_injection: bool,
    pub compression_enabled: bool,
    pub quality_profiles: Vec<RecordingQuality>,
}

impl Default for RecordingConfig {
    fn default() -> Self {
        Self {
            output_directory: PathBuf::from("./recordings"),
            max_concurrent_recordings: 50,
            segment_duration_ms: 30000, // 30 secondes par segment
            output_formats: vec![
                AudioFormat::Mp3 { bitrate: 320, sample_rate: 44100 },
                AudioFormat::Flac { sample_rate: 44100, bit_depth: 24 },
                AudioFormat::Wav { sample_rate: 44100, bit_depth: 16 },
            ],
            real_time_transcoding: true,
            metadata_injection: true,
            compression_enabled: true,
            quality_profiles: vec![
                RecordingQuality::high(),
                RecordingQuality::medium(),
                RecordingQuality::low(),
            ],
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum AudioFormat {
    Mp3 { bitrate: u32, sample_rate: u32 },
    Flac { sample_rate: u32, bit_depth: u8 },
    Wav { sample_rate: u32, bit_depth: u8 },
    Opus { bitrate: u32, sample_rate: u32 },
    Aac { bitrate: u32, sample_rate: u32 },
}

impl AudioFormat {
    pub fn get_extension(&self) -> &'static str {
        match self {
            AudioFormat::Mp3 { .. } => "mp3",
            AudioFormat::Flac { .. } => "flac",
            AudioFormat::Wav { .. } => "wav",
            AudioFormat::Opus { .. } => "opus",
            AudioFormat::Aac { .. } => "aac",
        }
    }

    pub fn get_mime_type(&self) -> &'static str {
        match self {
            AudioFormat::Mp3 { .. } => "audio/mpeg",
            AudioFormat::Flac { .. } => "audio/flac",
            AudioFormat::Wav { .. } => "audio/wav",
            AudioFormat::Opus { .. } => "audio/opus",
            AudioFormat::Aac { .. } => "audio/aac",
        }
    }

    pub fn get_bitrate(&self) -> u32 {
        match self {
            AudioFormat::Mp3 { bitrate, .. } => *bitrate,
            AudioFormat::Flac { sample_rate, bit_depth } => sample_rate * (*bit_depth as u32) * 2,
            AudioFormat::Wav { sample_rate, bit_depth } => sample_rate * (*bit_depth as u32) * 2,
            AudioFormat::Opus { bitrate, .. } => *bitrate,
            AudioFormat::Aac { bitrate, .. } => *bitrate,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RecordingQuality {
    pub name: String,
    pub bitrate: u32,
    pub sample_rate: u32,
    pub channels: u8,
    pub format: AudioFormat,
    pub target_file_size_mb: Option<u32>,
}

impl RecordingQuality {
    pub fn high() -> Self {
        Self {
            name: "high".to_string(),
            bitrate: 320,
            sample_rate: 44100,
            channels: 2,
            format: AudioFormat::Flac { sample_rate: 44100, bit_depth: 24 },
            target_file_size_mb: None,
        }
    }

    pub fn medium() -> Self {
        Self {
            name: "medium".to_string(),
            bitrate: 192,
            sample_rate: 44100,
            channels: 2,
            format: AudioFormat::Mp3 { bitrate: 192, sample_rate: 44100 },
            target_file_size_mb: Some(50),
        }
    }

    pub fn low() -> Self {
        Self {
            name: "low".to_string(),
            bitrate: 128,
            sample_rate: 22050,
            channels: 1,
            format: AudioFormat::Mp3 { bitrate: 128, sample_rate: 22050 },
            target_file_size_mb: Some(25),
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct LiveRecording {
    pub recording_id: String,
    pub session_id: String,
    pub stream_id: String,
    pub state: RecordingState,
    pub start_time: SystemTime,
    pub end_time: Option<SystemTime>,
    pub duration_ms: u64,
    pub file_paths: HashMap<String, PathBuf>,
    pub metadata: RecordingMetadata,
    pub segments: Vec<RecordingSegment>,
    pub transcoding_jobs: Vec<TranscodingJob>,
    pub stats: RecordingStats,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum RecordingState {
    Preparing,
    Recording,
    Transcoding,
    Completed,
    Failed,
    Stopped,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RecordingMetadata {
    pub title: Option<String>,
    pub artist: Option<String>,
    pub album: Option<String>,
    pub genre: Option<String>,
    pub duration_ms: u64,
    pub bitrate: u32,
    pub sample_rate: u32,
    pub channels: u8,
    pub file_size_bytes: u64,
    pub creation_time: SystemTime,
    pub tags: HashMap<String, String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RecordingSegment {
    pub segment_id: String,
    pub start_time_ms: u64,
    pub duration_ms: u32,
    pub file_path: PathBuf,
    pub file_size_bytes: u64,
    pub checksum: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TranscodingJob {
    pub job_id: String,
    pub input_format: AudioFormat,
    pub output_format: AudioFormat,
    pub progress_percent: f32,
    pub state: TranscodingState,
    pub started_at: SystemTime,
    pub estimated_completion: Option<SystemTime>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum TranscodingState {
    Queued,
    Processing,
    Completed,
    Failed,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RecordingStats {
    pub bytes_recorded: u64,
    pub segments_created: u32,
    pub transcoding_jobs_completed: u32,
    pub transcoding_jobs_failed: u32,
    pub average_bitrate: u32,
    pub peak_bitrate: u32,
    pub recording_efficiency: f32, // 0.0 - 1.0
    pub disk_usage_mb: f32,
}

impl Default for RecordingStats {
    fn default() -> Self {
        Self {
            bytes_recorded: 0,
            segments_created: 0,
            transcoding_jobs_completed: 0,
            transcoding_jobs_failed: 0,
            average_bitrate: 0,
            peak_bitrate: 0,
            recording_efficiency: 1.0,
            disk_usage_mb: 0.0,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(tag = "type", content = "data")]
pub enum RecordingMessage {
    StartRecording {
        recording_id: String,
        session_id: String,
        quality: RecordingQuality,
        metadata: RecordingMetadata,
    },
    StopRecording {
        recording_id: String,
    },
    SegmentCompleted {
        recording_id: String,
        segment: RecordingSegment,
    },
    TranscodingUpdate {
        job_id: String,
        progress: f32,
        state: TranscodingState,
    },
    RecordingError {
        recording_id: String,
        error: String,
    },
    StatsUpdate {
        recording_id: String,
        stats: RecordingStats,
    },
}

/// Gestionnaire d'enregistrement temps réel
#[derive(Clone)]
pub struct LiveRecordingManager {
    config: RecordingConfig,
    recordings: Arc<RwLock<HashMap<String, LiveRecording>>>,
    recording_tx: broadcast::Sender<RecordingMessage>,
    transcoding_queue: Arc<RwLock<Vec<TranscodingJob>>>,
    stats_collector: Arc<RwLock<HashMap<String, RecordingStats>>>,
}

impl LiveRecordingManager {
    pub fn new(config: RecordingConfig) -> Self {
        let (recording_tx, _) = broadcast::channel(1000);

        Self {
            config,
            recordings: Arc::new(RwLock::new(HashMap::new())),
            recording_tx,
            transcoding_queue: Arc::new(RwLock::new(Vec::new())),
            stats_collector: Arc::new(RwLock::new(HashMap::new())),
        }
    }

    /// Démarrer le gestionnaire d'enregistrement
    pub async fn start(&self) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
        info!("Starting Live Recording Manager with max {} concurrent recordings", 
              self.config.max_concurrent_recordings);

        // Créer le répertoire de sortie
        tokio::fs::create_dir_all(&self.config.output_directory).await?;

        // Démarrer le processeur de transcodage
        if self.config.real_time_transcoding {
            self.start_transcoding_processor().await;
        }

        // Démarrer le moniteur de segments
        self.start_segment_monitor().await;

        // Démarrer le collecteur de statistiques
        self.start_stats_collector().await;

        Ok(())
    }

    /// Commencer un nouvel enregistrement
    pub async fn start_recording(
        &self,
        session_id: String,
        stream_id: String,
        quality: RecordingQuality,
        metadata: RecordingMetadata,
    ) -> Result<String, Box<dyn std::error::Error + Send + Sync>> {
        let span = span!(Level::INFO, "start_recording", session_id = %session_id);
        let _enter = span.enter();

        let mut recordings = self.recordings.write().await;
        
        if recordings.len() >= self.config.max_concurrent_recordings {
            return Err("Maximum number of concurrent recordings reached".into());
        }

        let recording_id = Uuid::new_v4().to_string();
        
        let recording = LiveRecording {
            recording_id: recording_id.clone(),
            session_id: session_id.clone(),
            stream_id: stream_id.clone(),
            state: RecordingState::Preparing,
            start_time: SystemTime::now(),
            end_time: None,
            duration_ms: 0,
            file_paths: HashMap::new(),
            metadata: metadata.clone(),
            segments: Vec::new(),
            transcoding_jobs: Vec::new(),
            stats: RecordingStats::default(),
        };

        recordings.insert(recording_id.clone(), recording);

        info!("Started live recording: {} for session: {} stream: {}", 
              recording_id, session_id, stream_id);

        // Envoyer message de début d'enregistrement
        let start_msg = RecordingMessage::StartRecording {
            recording_id: recording_id.clone(),
            session_id,
            quality,
            metadata,
        };

        if let Err(e) = self.recording_tx.send(start_msg) {
            warn!("Failed to send recording start message: {}", e);
        }

        // Initialiser l'enregistrement
        self.initialize_recording(&recording_id).await?;

        Ok(recording_id)
    }

    /// Initialiser un enregistrement
    async fn initialize_recording(&self, recording_id: &str) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
        let mut recordings = self.recordings.write().await;
        
        if let Some(recording) = recordings.get_mut(recording_id) {
            // Créer les chemins de fichiers pour chaque format
            for format in &self.config.output_formats {
                let filename = format!(
                    "{}_{}_{}_{}.{}",
                    recording.session_id,
                    recording.stream_id,
                    recording_id,
                    chrono::Utc::now().format("%Y%m%d_%H%M%S"),
                    format.get_extension()
                );
                
                let file_path = self.config.output_directory.join(filename);
                recording.file_paths.insert(format.get_extension().to_string(), file_path);
            }

            recording.state = RecordingState::Recording;
            
            // Planifier les tâches de transcodage si activé
            if self.config.real_time_transcoding {
                self.schedule_transcoding_jobs(recording_id).await?;
            }
        }

        Ok(())
    }

    /// Planifier les tâches de transcodage
    async fn schedule_transcoding_jobs(&self, recording_id: &str) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
        let recordings = self.recordings.read().await;
        
        if let Some(_recording) = recordings.get(recording_id) {
            let mut queue = self.transcoding_queue.write().await;
            
            // Créer des tâches pour chaque format de sortie
            for format in &self.config.output_formats {
                let job = TranscodingJob {
                    job_id: Uuid::new_v4().to_string(),
                    input_format: AudioFormat::Wav { sample_rate: 44100, bit_depth: 16 }, // Format source
                    output_format: format.clone(),
                    progress_percent: 0.0,
                    state: TranscodingState::Queued,
                    started_at: SystemTime::now(),
                    estimated_completion: None,
                };

                queue.push(job);
            }
        }

        Ok(())
    }

    /// Arrêter un enregistrement
    pub async fn stop_recording(&self, recording_id: &str) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
        let mut recordings = self.recordings.write().await;
        
        if let Some(recording) = recordings.get_mut(recording_id) {
            recording.state = RecordingState::Transcoding;
            recording.end_time = Some(SystemTime::now());
            
            if let Ok(duration) = recording.start_time.elapsed() {
                recording.duration_ms = duration.as_millis() as u64;
            }

            info!("Stopped live recording: {} (duration: {}ms)", 
                  recording_id, recording.duration_ms);

            // Envoyer message d'arrêt
            let stop_msg = RecordingMessage::StopRecording {
                recording_id: recording_id.to_string(),
            };

            if let Err(e) = self.recording_tx.send(stop_msg) {
                warn!("Failed to send recording stop message: {}", e);
            }

            // Finaliser l'enregistrement
            self.finalize_recording(recording_id).await?;
        }

        Ok(())
    }

    /// Finaliser un enregistrement
    async fn finalize_recording(&self, recording_id: &str) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
        let mut recordings = self.recordings.write().await;
        
        if let Some(recording) = recordings.get_mut(recording_id) {
            // Calculer les statistiques finales
            recording.stats.recording_efficiency = self.calculate_recording_efficiency(recording);
            
            // Injecter les métadonnées si activé
            if self.config.metadata_injection {
                self.inject_metadata(recording).await?;
            }

            recording.state = RecordingState::Completed;
            
            info!("Finalized recording: {} with {} segments", 
                  recording_id, recording.segments.len());
        }

        Ok(())
    }

    /// Calculer l'efficacité d'enregistrement
    fn calculate_recording_efficiency(&self, recording: &LiveRecording) -> f32 {
        if recording.duration_ms == 0 {
            return 0.0;
        }

        let expected_bytes = (recording.duration_ms * recording.metadata.bitrate as u64) / 8000;
        let actual_bytes = recording.stats.bytes_recorded;

        if expected_bytes == 0 {
            1.0
        } else {
            (actual_bytes as f32) / (expected_bytes as f32)
        }
    }

    /// Injecter les métadonnées dans les fichiers
    async fn inject_metadata(&self, recording: &LiveRecording) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
        for (format_name, file_path) in &recording.file_paths {
            match format_name.as_str() {
                "mp3" => {
                    // Injection métadonnées MP3 (ID3)
                    self.inject_mp3_metadata(file_path, &recording.metadata).await?;
                }
                "flac" => {
                    // Injection métadonnées FLAC
                    self.inject_flac_metadata(file_path, &recording.metadata).await?;
                }
                _ => {
                    debug!("Metadata injection not supported for format: {}", format_name);
                }
            }
        }

        Ok(())
    }

    /// Injecter métadonnées MP3
    async fn inject_mp3_metadata(&self, _file_path: &Path, _metadata: &RecordingMetadata) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
        // Implémentation simplifiée - en production utiliser une librairie comme id3
        debug!("MP3 metadata injection would be implemented here");
        Ok(())
    }

    /// Injecter métadonnées FLAC
    async fn inject_flac_metadata(&self, _file_path: &Path, _metadata: &RecordingMetadata) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
        // Implémentation simplifiée - en production utiliser une librairie comme metaflac
        debug!("FLAC metadata injection would be implemented here");
        Ok(())
    }

    /// Obtenir les statistiques en temps réel
    pub async fn get_recording_stats(&self) -> serde_json::Value {
        let recordings = self.recordings.read().await;
        let transcoding_queue = self.transcoding_queue.read().await;
        
        let total_recordings = recordings.len();
        let active_recordings = recordings.values()
            .filter(|r| matches!(r.state, RecordingState::Recording))
            .count();
        
        let completed_recordings = recordings.values()
            .filter(|r| matches!(r.state, RecordingState::Completed))
            .count();

        let total_segments: u32 = recordings.values()
            .map(|r| r.segments.len() as u32)
            .sum();

        let total_disk_usage_mb: f32 = recordings.values()
            .map(|r| r.stats.disk_usage_mb)
            .sum();

        let pending_transcoding_jobs = transcoding_queue.len();

        serde_json::json!({
            "recording_stats": {
                "total_recordings": total_recordings,
                "active_recordings": active_recordings,
                "completed_recordings": completed_recordings,
                "total_segments": total_segments,
                "total_disk_usage_mb": total_disk_usage_mb,
                "pending_transcoding_jobs": pending_transcoding_jobs,
                "max_concurrent_recordings": self.config.max_concurrent_recordings,
                "output_formats": self.config.output_formats.len(),
                "real_time_transcoding_enabled": self.config.real_time_transcoding
            }
        })
    }

    /// Démarrer le processeur de transcodage
    async fn start_transcoding_processor(&self) {
        let queue = self.transcoding_queue.clone();
        let recording_tx = self.recording_tx.clone();

        tokio::spawn(async move {
            let mut interval = tokio::time::interval(Duration::from_secs(1));
            
            loop {
                interval.tick().await;
                
                let mut queue_guard = queue.write().await;
                
                // Traiter les tâches en attente
                if let Some(mut job) = queue_guard.pop() {
                    job.state = TranscodingState::Processing;
                    let recording_tx_clone = recording_tx.clone();
                    
                    // Simulation du transcodage
                    tokio::spawn(async move {
                        // Ici on implémenterait le vrai transcodage
                        tokio::time::sleep(Duration::from_secs(2)).await;
                        
                        let update_msg = RecordingMessage::TranscodingUpdate {
                            job_id: job.job_id.clone(),
                            progress: 100.0,
                            state: TranscodingState::Completed,
                        };

                        if let Err(e) = recording_tx_clone.send(update_msg) {
                            warn!("Failed to send transcoding update: {}", e);
                        }
                    });
                }
            }
        });
    }

    /// Démarrer le moniteur de segments
    async fn start_segment_monitor(&self) {
        let recordings = self.recordings.clone();
        let segment_duration = Duration::from_millis(self.config.segment_duration_ms as u64);

        tokio::spawn(async move {
            let mut interval = tokio::time::interval(segment_duration);
            
            loop {
                interval.tick().await;
                
                let mut recordings_guard = recordings.write().await;
                for (recording_id, recording) in recordings_guard.iter_mut() {
                    if matches!(recording.state, RecordingState::Recording) {
                        // Créer un nouveau segment
                        let segment = RecordingSegment {
                            segment_id: Uuid::new_v4().to_string(),
                            start_time_ms: recording.duration_ms,
                            duration_ms: segment_duration.as_millis() as u32,
                            file_path: PathBuf::from(format!("segment_{}_{}.tmp", recording_id, recording.segments.len())),
                            file_size_bytes: 0, // Sera calculé
                            checksum: "".to_string(), // Sera calculé
                        };

                        recording.segments.push(segment);
                        recording.stats.segments_created += 1;
                    }
                }
            }
        });
    }

    /// Démarrer le collecteur de statistiques
    async fn start_stats_collector(&self) {
        let recordings = self.recordings.clone();

        tokio::spawn(async move {
            let mut interval = tokio::time::interval(Duration::from_secs(5));
            
            loop {
                interval.tick().await;
                
                let recordings_guard = recordings.read().await;
                for (recording_id, recording) in recordings_guard.iter() {
                    if matches!(recording.state, RecordingState::Recording) {
                        debug!("Recording {} active with {} segments", 
                               recording_id, recording.segments.len());
                    }
                }
            }
        });
    }

    /// Obtenir un receiver pour les messages d'enregistrement
    pub fn get_recording_receiver(&self) -> broadcast::Receiver<RecordingMessage> {
        self.recording_tx.subscribe()
    }

    /// Supprimer un enregistrement
    pub async fn remove_recording(&self, recording_id: &str) -> bool {
        let mut recordings = self.recordings.write().await;
        if recordings.remove(recording_id).is_some() {
            info!("Removed recording: {}", recording_id);
            true
        } else {
            false
        }
    }
}
