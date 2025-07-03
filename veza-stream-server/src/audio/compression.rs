use std::collections::HashMap;
use std::path::{Path, PathBuf};
use std::sync::Arc;
use std::time::{Duration, SystemTime};
use tokio::sync::RwLock;
use serde::{Deserialize, Serialize};
use crate::Config;
use tracing::{debug, info, error};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CompressionProfile {
    pub name: String,
    pub codec: AudioCodec,
    pub bitrate_kbps: u32,
    pub sample_rate: u32,
    pub channels: u8,
    pub quality_factor: f32, // 0.0 - 1.0
    pub compression_level: u8, // 0-9 for most codecs
    pub target_size_reduction: f32, // Pourcentage de réduction visé
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum AudioCodec {
    MP3,
    AAC,
    OGG,
    OPUS,
    FLAC,
    WAV,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CompressionJob {
    pub id: String,
    pub input_path: PathBuf,
    pub output_path: PathBuf,
    pub profile: CompressionProfile,
    pub status: JobStatus,
    pub progress: f32,
    pub created_at: SystemTime,
    pub started_at: Option<SystemTime>,
    pub completed_at: Option<SystemTime>,
    pub error_message: Option<String>,
    pub original_size_bytes: u64,
    pub compressed_size_bytes: Option<u64>,
    pub compression_ratio: Option<f32>,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum JobStatus {
    Pending,
    InProgress,
    Completed,
    Failed,
    Cancelled,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CompressionStats {
    pub total_jobs: u64,
    pub completed_jobs: u64,
    pub failed_jobs: u64,
    pub average_compression_ratio: f32,
    pub total_space_saved_mb: u64,
    pub processing_queue_size: usize,
    pub average_processing_time_ms: u64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CompressionRequest {
    pub input_file: String,
    pub target_quality: String,
    pub preserve_metadata: bool,
    pub async_processing: bool,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CompressionResponse {
    pub job_id: String,
    pub status: JobStatus,
    pub estimated_completion_time: Option<u64>,
    pub download_url: Option<String>,
}

pub struct CompressionEngine {
    config: Arc<Config>,
    profiles: HashMap<String, CompressionProfile>,
    active_jobs: Arc<RwLock<HashMap<String, CompressionJob>>>,
    job_queue: Arc<RwLock<Vec<String>>>,
    stats: Arc<RwLock<CompressionStats>>,
    worker_count: usize,
}

impl CompressionEngine {
    pub fn new(config: Arc<Config>) -> Self {
        let mut profiles = HashMap::new();
        
        // Profils prédéfinis
        profiles.insert("ultra_high".to_string(), CompressionProfile {
            name: "Ultra High Quality".to_string(),
            codec: AudioCodec::FLAC,
            bitrate_kbps: 1411, // CD quality
            sample_rate: 44100,
            channels: 2,
            quality_factor: 1.0,
            compression_level: 8,
            target_size_reduction: 0.3, // 30% de réduction
        });

        profiles.insert("high".to_string(), CompressionProfile {
            name: "High Quality".to_string(),
            codec: AudioCodec::AAC,
            bitrate_kbps: 320,
            sample_rate: 44100,
            channels: 2,
            quality_factor: 0.9,
            compression_level: 6,
            target_size_reduction: 0.5,
        });

        profiles.insert("medium".to_string(), CompressionProfile {
            name: "Medium Quality".to_string(),
            codec: AudioCodec::MP3,
            bitrate_kbps: 192,
            sample_rate: 44100,
            channels: 2,
            quality_factor: 0.7,
            compression_level: 5,
            target_size_reduction: 0.7,
        });

        profiles.insert("low".to_string(), CompressionProfile {
            name: "Low Quality".to_string(),
            codec: AudioCodec::MP3,
            bitrate_kbps: 128,
            sample_rate: 22050,
            channels: 2,
            quality_factor: 0.5,
            compression_level: 4,
            target_size_reduction: 0.8,
        });

        profiles.insert("mobile".to_string(), CompressionProfile {
            name: "Mobile Optimized".to_string(),
            codec: AudioCodec::OPUS,
            bitrate_kbps: 96,
            sample_rate: 48000,
            channels: 2,
            quality_factor: 0.6,
            compression_level: 3,
            target_size_reduction: 0.85,
        });

        profiles.insert("podcast".to_string(), CompressionProfile {
            name: "Podcast/Voice".to_string(),
            codec: AudioCodec::OPUS,
            bitrate_kbps: 64,
            sample_rate: 22050,
            channels: 1, // Mono pour la voix
            quality_factor: 0.8,
            compression_level: 5,
            target_size_reduction: 0.9,
        });

        let worker_count = config.performance.worker_threads.unwrap_or_else(|| {
            std::thread::available_parallelism().map(|p| p.get()).unwrap_or(4)
        });

        Self {
            config,
            profiles,
            active_jobs: Arc::new(RwLock::new(HashMap::new())),
            job_queue: Arc::new(RwLock::new(Vec::new())),
            stats: Arc::new(RwLock::new(CompressionStats {
                total_jobs: 0,
                completed_jobs: 0,
                failed_jobs: 0,
                average_compression_ratio: 0.0,
                total_space_saved_mb: 0,
                processing_queue_size: 0,
                average_processing_time_ms: 0,
            })),
            worker_count,
        }
    }

    pub async fn start_workers(&self) {
        info!("🔧 Démarrage de {} workers de compression", self.worker_count);
        
        for worker_id in 0..self.worker_count {
            let engine = self.clone();
            tokio::spawn(async move {
                engine.worker_loop(worker_id).await;
            });
        }
    }

    async fn worker_loop(&self, worker_id: usize) {
        debug!("Worker de compression {} démarré", worker_id);
        
        loop {
            // Récupérer le prochain job de la queue
            let job_id = {
                let mut queue = self.job_queue.write().await;
                queue.pop()
            };

            if let Some(job_id) = job_id {
                debug!("Worker {} traite le job {}", worker_id, job_id);
                self.process_job(&job_id).await;
            } else {
                // Pas de job, attendre un peu
                tokio::time::sleep(Duration::from_millis(100)).await;
            }
        }
    }

    pub async fn compress_audio(&self, request: CompressionRequest) -> Result<CompressionResponse, CompressionError> {
        let profile = self.profiles.get(&request.target_quality)
            .ok_or_else(|| CompressionError::InvalidProfile(request.target_quality.clone()))?;

        let input_path = PathBuf::from(&self.config.audio_dir).join(&request.input_file);
        if !input_path.exists() {
            return Err(CompressionError::FileNotFound(request.input_file));
        }

        let job_id = uuid::Uuid::new_v4().to_string();
        let output_filename = self.generate_output_filename(&request.input_file, profile);
        let output_path = PathBuf::from("compressed").join(&output_filename);

        // Créer le répertoire de sortie si nécessaire
        if let Some(parent) = output_path.parent() {
            tokio::fs::create_dir_all(parent).await
                .map_err(|e| CompressionError::IoError(e.to_string()))?;
        }

        let original_size = tokio::fs::metadata(&input_path).await
            .map_err(|e| CompressionError::IoError(e.to_string()))?
            .len();

        let job = CompressionJob {
            id: job_id.clone(),
            input_path,
            output_path,
            profile: profile.clone(),
            status: JobStatus::Pending,
            progress: 0.0,
            created_at: SystemTime::now(),
            started_at: None,
            completed_at: None,
            error_message: None,
            original_size_bytes: original_size,
            compressed_size_bytes: None,
            compression_ratio: None,
        };

        // Ajouter le job à la liste active
        {
            let mut active_jobs = self.active_jobs.write().await;
            active_jobs.insert(job_id.clone(), job);
        }

        if request.async_processing {
            // Traitement asynchrone - ajouter à la queue
            {
                let mut queue = self.job_queue.write().await;
                queue.push(job_id.clone());
            }

            {
                let mut stats = self.stats.write().await;
                stats.total_jobs += 1;
                stats.processing_queue_size = self.job_queue.read().await.len();
            }

            Ok(CompressionResponse {
                job_id,
                status: JobStatus::Pending,
                estimated_completion_time: Some(self.estimate_completion_time().await),
                download_url: None,
            })
        } else {
            // Traitement synchrone
            self.process_job(&job_id).await;
            
            let job = {
                let active_jobs = self.active_jobs.read().await;
                active_jobs.get(&job_id).cloned()
            };

            if let Some(job) = job {
                let download_url = if job.status == JobStatus::Completed {
                    Some(format!("/api/compressed/{}", output_filename))
                } else {
                    None
                };

                Ok(CompressionResponse {
                    job_id,
                    status: job.status,
                    estimated_completion_time: None,
                    download_url,
                })
            } else {
                Err(CompressionError::JobNotFound(job_id))
            }
        }
    }

    async fn process_job(&self, job_id: &str) {
        let mut job = {
            let mut active_jobs = self.active_jobs.write().await;
            if let Some(job) = active_jobs.get_mut(job_id) {
                job.status = JobStatus::InProgress;
                job.started_at = Some(SystemTime::now());
                job.clone()
            } else {
                error!("Job {} non trouvé", job_id);
                return;
            }
        };

        debug!("Début de compression pour le job {}", job_id);

        let result = self.execute_compression(&mut job).await;

        // Mettre à jour le job avec le résultat
        {
            let mut active_jobs = self.active_jobs.write().await;
            if let Some(active_job) = active_jobs.get_mut(job_id) {
                *active_job = job.clone();
            }
        }

        // Mettre à jour les statistiques
        {
            let mut stats = self.stats.write().await;
            match result {
                Ok(_) => {
                    stats.completed_jobs += 1;
                    if let Some(ratio) = job.compression_ratio {
                        stats.average_compression_ratio = 
                            (stats.average_compression_ratio * (stats.completed_jobs - 1) as f32 + ratio) / stats.completed_jobs as f32;
                    }
                    if let Some(compressed_size) = job.compressed_size_bytes {
                        let space_saved = (job.original_size_bytes - compressed_size) / (1024 * 1024);
                        stats.total_space_saved_mb += space_saved;
                    }
                }
                Err(_) => {
                    stats.failed_jobs += 1;
                }
            }
            stats.processing_queue_size = self.job_queue.read().await.len();
        }

        if result.is_ok() {
            info!("✅ Compression terminée avec succès pour le job {}", job_id);
        } else {
            error!("❌ Échec de compression pour le job {}: {:?}", job_id, result);
        }
    }

    async fn execute_compression(&self, job: &mut CompressionJob) -> Result<(), CompressionError> {
        let _start_time = SystemTime::now();

        // Simuler la progression de la compression
        for progress in (0..=100).step_by(10) {
            job.progress = progress as f32;
            
            // Mettre à jour le job dans la liste active
            {
                let mut active_jobs = self.active_jobs.write().await;
                if let Some(active_job) = active_jobs.get_mut(&job.id) {
                    active_job.progress = job.progress;
                }
            }

            // Simuler le temps de traitement
            tokio::time::sleep(Duration::from_millis(100)).await;
        }

        // Exécuter la compression réelle (simulation)
        let compression_result = self.perform_actual_compression(job).await;

        job.completed_at = Some(SystemTime::now());
        
        match compression_result {
            Ok(compressed_size) => {
                job.status = JobStatus::Completed;
                job.compressed_size_bytes = Some(compressed_size);
                job.compression_ratio = Some(compressed_size as f32 / job.original_size_bytes as f32);
                job.progress = 100.0;
                Ok(())
            }
            Err(e) => {
                job.status = JobStatus::Failed;
                job.error_message = Some(e.to_string());
                Err(e)
            }
        }
    }

    async fn perform_actual_compression(&self, job: &CompressionJob) -> Result<u64, CompressionError> {
        // Dans une implémentation réelle, ici on utiliserait FFmpeg ou une autre bibliothèque
        // Pour la démo, on simule la compression
        
        debug!("Compression de {:?} vers {:?} avec le profil {:?}", 
               job.input_path, job.output_path, job.profile.name);

        // Simuler la création du fichier compressé
        let simulated_compressed_size = (job.original_size_bytes as f32 * 
            (1.0 - job.profile.target_size_reduction)) as u64;

        // Créer un fichier de simulation
        tokio::fs::write(&job.output_path, b"compressed_audio_data_simulation").await
            .map_err(|e| CompressionError::IoError(e.to_string()))?;

        Ok(simulated_compressed_size)
    }

    fn generate_output_filename(&self, input_filename: &str, profile: &CompressionProfile) -> String {
        let input_path = Path::new(input_filename);
        let stem = input_path.file_stem().unwrap_or_default().to_string_lossy();
        let extension = match profile.codec {
            AudioCodec::MP3 => "mp3",
            AudioCodec::AAC => "aac",
            AudioCodec::OGG => "ogg",
            AudioCodec::OPUS => "opus",
            AudioCodec::FLAC => "flac",
            AudioCodec::WAV => "wav",
        };

        format!("{}_{}.{}", stem, profile.name.replace(" ", "_").to_lowercase(), extension)
    }

    async fn estimate_completion_time(&self) -> u64 {
        let queue_size = self.job_queue.read().await.len();
        let stats = self.stats.read().await;
        
        if stats.average_processing_time_ms > 0 {
            (queue_size as u64 * stats.average_processing_time_ms) / (self.worker_count as u64 * 1000)
        } else {
            300 // 5 minutes par défaut
        }
    }

    pub async fn get_job_status(&self, job_id: &str) -> Option<CompressionJob> {
        let active_jobs = self.active_jobs.read().await;
        active_jobs.get(job_id).cloned()
    }

    pub async fn cancel_job(&self, job_id: &str) -> Result<(), CompressionError> {
        {
            let mut active_jobs = self.active_jobs.write().await;
            if let Some(job) = active_jobs.get_mut(job_id) {
                if job.status == JobStatus::Pending || job.status == JobStatus::InProgress {
                    job.status = JobStatus::Cancelled;
                } else {
                    return Err(CompressionError::JobNotCancellable(job_id.to_string()));
                }
            } else {
                return Err(CompressionError::JobNotFound(job_id.to_string()));
            }
        }

        // Retirer de la queue si présent
        {
            let mut queue = self.job_queue.write().await;
            queue.retain(|id| id != job_id);
        }

        Ok(())
    }

    pub async fn get_compression_stats(&self) -> CompressionStats {
        self.stats.read().await.clone()
    }

    pub async fn list_profiles(&self) -> Vec<CompressionProfile> {
        self.profiles.values().cloned().collect()
    }

    pub async fn add_custom_profile(&mut self, name: String, profile: CompressionProfile) {
        self.profiles.insert(name, profile);
    }

    pub async fn cleanup_completed_jobs(&self, max_age: Duration) {
        let cutoff = SystemTime::now() - max_age;
        let mut active_jobs = self.active_jobs.write().await;
        
        active_jobs.retain(|_, job| {
            match job.status {
                JobStatus::Completed | JobStatus::Failed | JobStatus::Cancelled => {
                    job.completed_at.map_or(true, |completed| completed > cutoff)
                }
                _ => true,
            }
        });
    }

    pub async fn get_queue_info(&self) -> serde_json::Value {
        let queue = self.job_queue.read().await;
        let active_jobs = self.active_jobs.read().await;
        
        let pending_jobs: Vec<_> = queue.iter()
            .filter_map(|job_id| active_jobs.get(job_id))
            .collect();

        let in_progress_jobs: Vec<_> = active_jobs.values()
            .filter(|job| job.status == JobStatus::InProgress)
            .collect();

        serde_json::json!({
            "queue_size": queue.len(),
            "pending_jobs": pending_jobs,
            "in_progress_jobs": in_progress_jobs,
            "worker_count": self.worker_count,
        })
    }
}

impl Clone for CompressionEngine {
    fn clone(&self) -> Self {
        Self {
            config: self.config.clone(),
            profiles: self.profiles.clone(),
            active_jobs: self.active_jobs.clone(),
            job_queue: self.job_queue.clone(),
            stats: self.stats.clone(),
            worker_count: self.worker_count,
        }
    }
}

#[derive(Debug, thiserror::Error)]
pub enum CompressionError {
    #[error("Profile de compression invalide: {0}")]
    InvalidProfile(String),
    
    #[error("Fichier non trouvé: {0}")]
    FileNotFound(String),
    
    #[error("Erreur I/O: {0}")]
    IoError(String),
    
    #[error("Job non trouvé: {0}")]
    JobNotFound(String),
    
    #[error("Job non annulable: {0}")]
    JobNotCancellable(String),
    
    #[error("Erreur de compression: {0}")]
    CompressionFailed(String),
    
    #[error("Format audio non supporté: {0}")]
    UnsupportedFormat(String),
} 