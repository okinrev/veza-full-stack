/// Module d'encodage multi-codec pour streaming production
/// 
/// Support des codecs :
/// - Opus (primary) - Ultra low latency, haute qualité
/// - AAC (fallback) - Compatibilité iOS/Safari  
/// - MP3 (legacy) - Compatibilité universelle
/// - FLAC (lossless) - Qualité studio pour premium

use std::sync::Arc;
use std::collections::HashMap;
use std::time::Duration;
use std::fmt;

use parking_lot::RwLock;
use rayon::prelude::*;
use tokio::sync::mpsc;
use uuid::Uuid;
use serde::{Serialize, Deserialize};
use tracing::{info, warn, error, debug};

use crate::core::AudioFormat;
use crate::error::AppError;
use crate::core::{StreamSource, StreamOutput};

/// Pool d'encodeurs réutilisables pour performance optimale
#[derive(Debug)]
pub struct EncoderPool {
    /// Encodeurs Opus disponibles
    opus_encoders: Arc<RwLock<Vec<Box<dyn OpusEncoder + Send + Sync>>>>,
    /// Encodeurs AAC disponibles  
    aac_encoders: Arc<RwLock<Vec<Box<dyn AacEncoder + Send + Sync>>>>,
    /// Encodeurs MP3 disponibles
    mp3_encoders: Arc<RwLock<Vec<Box<dyn Mp3Encoder + Send + Sync>>>>,
    /// Encodeurs FLAC disponibles
    flac_encoders: Arc<RwLock<Vec<Box<dyn FlacEncoder + Send + Sync>>>>,
    /// Configuration du pool
    config: EncoderPoolConfig,
    /// Métriques d'utilisation
    metrics: Arc<EncoderMetrics>,
}

/// Pipeline d'encodage pour un stream spécifique
#[derive(Debug)]
pub struct EncoderPipeline {
    pub id: Uuid,
    pub input_format: AudioFormat,
    pub outputs: Vec<EncoderOutput>,
    pub effects_chain: Vec<Box<dyn AudioEffect + Send + Sync>>,
    pub hardware_acceleration: bool,
    pub real_time_processing: bool,
    pub buffer_size: usize,
    pub processing_thread: Option<tokio::task::JoinHandle<()>>,
}

/// Configuration d'un encodeur de sortie
#[derive(Debug, Clone)]
pub struct EncoderOutput {
    pub id: Uuid,
    pub codec: AudioCodec,
    pub bitrate: u32,
    pub quality: QualityProfile,
    pub target_format: AudioFormat,
    pub encoding_preset: EncodingPreset,
    pub adaptive_bitrate: bool,
}

/// Codecs audio supportés  
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum AudioCodec {
    Opus {
        complexity: u8,  // 0-10, plus élevé = meilleure qualité
        signal_type: OpusSignalType,
        vbr_enabled: bool,
    },
    AAC {
        profile: AacProfile,
        object_type: AacObjectType,
        vbr_enabled: bool,
    },
    MP3 {
        mode: Mp3Mode,
        quality: u8,  // 0-9, 0 = meilleure qualité
        vbr_enabled: bool,
    },
    FLAC {
        compression_level: u8,  // 0-8
        verify: bool,
    },
}

/// Types de signal pour Opus
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum OpusSignalType {
    Auto,
    Voice,
    Music,
}

/// Profils AAC
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum AacProfile {
    LC,    // Low Complexity - standard
    HE,    // High Efficiency - pour bas débits
    HEv2,  // HE-AAC v2 - stéréo à très bas débit
}

/// Types d'objets AAC
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum AacObjectType {
    Main,
    LC,
    SSR,
    LTP,
}

/// Modes MP3
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum Mp3Mode {
    Stereo,
    JointStereo,
    DualChannel,
    Mono,
}

/// Profils de qualité prédéfinis  
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct QualityProfile {
    pub name: String,
    pub bitrate: u32,
    pub sample_rate: u32,
    pub channels: u8,
    pub description: String,
}

/// Presets d'encodage pour optimiser selon l'usage
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum EncodingPreset {
    /// Ultra low latency pour streaming live
    UltraLowLatency {
        max_latency_ms: u32,
        buffer_size: usize,
    },
    /// Streaming temps réel standard
    RealTime {
        target_latency_ms: u32,
        quality_priority: bool,
    },
    /// Haute qualité pour VOD
    HighQuality {
        multi_pass: bool,
        noise_reduction: bool,
    },
    /// Optimisé pour mobile/faible bande passante
    MobileOptimized {
        aggressive_compression: bool,
        adaptive_quality: bool,
    },
}

/// Configuration du pool d'encodeurs
#[derive(Debug, Clone)]
pub struct EncoderPoolConfig {
    pub opus_pool_size: usize,
    pub aac_pool_size: usize, 
    pub mp3_pool_size: usize,
    pub flac_pool_size: usize,
    pub enable_hardware_acceleration: bool,
    pub max_parallel_encodes: usize,
    pub enable_real_time_processing: bool,
}

/// Métriques d'utilisation des encodeurs
#[derive(Debug, Default)]
pub struct EncoderMetrics {
    pub opus_encodes_total: std::sync::atomic::AtomicU64,
    pub aac_encodes_total: std::sync::atomic::AtomicU64,
    pub mp3_encodes_total: std::sync::atomic::AtomicU64,
    pub flac_encodes_total: std::sync::atomic::AtomicU64,
    pub encode_errors_total: std::sync::atomic::AtomicU64,
    pub average_encode_time_ms: std::sync::atomic::AtomicU64,
    pub peak_cpu_usage: std::sync::atomic::AtomicU32,
    pub memory_usage_mb: std::sync::atomic::AtomicU64,
}

/// Trait pour les effets audio en temps réel
pub trait AudioEffect: fmt::Debug {
    fn process(&mut self, samples: &mut [f32], sample_rate: u32) -> Result<(), AppError>;
    fn latency(&self) -> Duration;
    fn enabled(&self) -> bool;
    fn set_enabled(&mut self, enabled: bool);
    fn parameters(&self) -> HashMap<String, f32>;
    fn set_parameter(&mut self, name: &str, value: f32) -> Result<(), AppError>;
}

/// Trait pour encodeurs Opus
pub trait OpusEncoder: fmt::Debug {
    fn encode(&mut self, samples: &[f32]) -> Result<Vec<u8>, AppError>;
    fn reset(&mut self) -> Result<(), AppError>;
    fn set_bitrate(&mut self, bitrate: u32) -> Result<(), AppError>;
    fn set_complexity(&mut self, complexity: u8) -> Result<(), AppError>;
}

/// Trait pour encodeurs AAC
pub trait AacEncoder: fmt::Debug {
    fn encode(&mut self, samples: &[f32]) -> Result<Vec<u8>, AppError>;
    fn reset(&mut self) -> Result<(), AppError>;
    fn set_bitrate(&mut self, bitrate: u32) -> Result<(), AppError>;
    fn set_profile(&mut self, profile: AacProfile) -> Result<(), AppError>;
}

/// Trait pour encodeurs MP3
pub trait Mp3Encoder: fmt::Debug {
    fn encode(&mut self, samples: &[f32]) -> Result<Vec<u8>, AppError>;
    fn reset(&mut self) -> Result<(), AppError>;
    fn set_bitrate(&mut self, bitrate: u32) -> Result<(), AppError>;
    fn set_quality(&mut self, quality: u8) -> Result<(), AppError>;
}

/// Trait pour encodeurs FLAC
pub trait FlacEncoder: fmt::Debug {
    fn encode(&mut self, samples: &[f32]) -> Result<Vec<u8>, AppError>;
    fn reset(&mut self) -> Result<(), AppError>;
    fn set_compression_level(&mut self, level: u8) -> Result<(), AppError>;
}

impl Default for EncoderPoolConfig {
    fn default() -> Self {
        Self {
            opus_pool_size: 50,
            aac_pool_size: 30,
            mp3_pool_size: 20,
            flac_pool_size: 10,
            enable_hardware_acceleration: true,
            max_parallel_encodes: num_cpus::get() * 2,
            enable_real_time_processing: true,
        }
    }
}

impl EncoderPool {
    /// Crée un nouveau pool d'encodeurs
    pub fn new() -> Result<Self, AppError> {
        Self::with_config(EncoderPoolConfig::default())
    }
    
    /// Crée un pool avec configuration personnalisée
    pub fn with_config(config: EncoderPoolConfig) -> Result<Self, AppError> {
        let opus_encoders = Arc::new(RwLock::new(Vec::new()));
        let aac_encoders = Arc::new(RwLock::new(Vec::new()));
        let mp3_encoders = Arc::new(RwLock::new(Vec::new()));
        let flac_encoders = Arc::new(RwLock::new(Vec::new()));
        
        // Pré-allouer les encodeurs dans le pool
        // TODO: Implémentation réelle des encodeurs
        
        Ok(Self {
            opus_encoders,
            aac_encoders,
            mp3_encoders,
            flac_encoders,
            config,
            metrics: Arc::new(EncoderMetrics::default()),
        })
    }
    
    /// Crée un pipeline d'encodage pour un stream
    pub async fn create_pipeline(
        &self,
        source: &StreamSource,
        outputs: &[StreamOutput],
    ) -> Result<Arc<EncoderPipeline>, AppError> {
        let pipeline_id = Uuid::new_v4();
        
        // Déterminer le format d'entrée depuis la source
        let input_format = match source {
            StreamSource::File { format, .. } => format.clone(),
            StreamSource::Live { format, .. } => format.clone(),
            StreamSource::External { format, .. } => format.clone().unwrap_or_default(),
            StreamSource::Generated { .. } => AudioFormat::default(),
        };
        
        // Créer les encodeurs de sortie
        let encoder_outputs = outputs.iter()
            .map(|output| self.create_encoder_output(output))
            .collect::<Result<Vec<_>, _>>()?;
        
        let pipeline = EncoderPipeline {
            id: pipeline_id,
            input_format,
            outputs: encoder_outputs,
            effects_chain: Vec::new(),
            hardware_acceleration: self.config.enable_hardware_acceleration,
            real_time_processing: self.config.enable_real_time_processing,
            buffer_size: 4096,
            processing_thread: None,
        };
        
        info!("Pipeline d'encodage créé: {}", pipeline_id);
        Ok(Arc::new(pipeline))
    }
    
    /// Crée un encodeur de sortie spécifique
    fn create_encoder_output(&self, output: &StreamOutput) -> Result<EncoderOutput, AppError> {
        let encoder_output = EncoderOutput {
            id: Uuid::new_v4(),
            codec: self.determine_codec(&output.format, output.bitrate)?,
            bitrate: output.bitrate,
            quality: self.get_quality_profile(output.bitrate),
            target_format: output.format.clone(),
            encoding_preset: self.determine_preset(&output.protocol),
            adaptive_bitrate: true,
        };
        
        Ok(encoder_output)
    }
    
    /// Détermine le codec optimal selon le format et bitrate
    fn determine_codec(&self, format: &AudioFormat, bitrate: u32) -> Result<AudioCodec, AppError> {
        match bitrate {
            0..=64 => Ok(AudioCodec::Opus {
                complexity: 5,
                signal_type: OpusSignalType::Music,
                vbr_enabled: true,
            }),
            65..=128 => Ok(AudioCodec::AAC {
                profile: AacProfile::HE,
                object_type: AacObjectType::LC,
                vbr_enabled: true,
            }),
            129..=320 => Ok(AudioCodec::MP3 {
                mode: Mp3Mode::JointStereo,
                quality: 2,
                vbr_enabled: true,
            }),
            _ => Ok(AudioCodec::FLAC {
                compression_level: 5,
                verify: false,
            }),
        }
    }
    
    /// Obtient un profil de qualité selon le bitrate
    fn get_quality_profile(&self, bitrate: u32) -> QualityProfile {
        match bitrate {
            0..=64 => QualityProfile {
                name: "Low".to_string(),
                bitrate,
                sample_rate: 22050,
                channels: 1,
                description: "Optimisé pour faible bande passante".to_string(),
            },
            65..=128 => QualityProfile {
                name: "Medium".to_string(),
                bitrate,
                sample_rate: 44100,
                channels: 2,
                description: "Qualité standard pour streaming".to_string(),
            },
            129..=256 => QualityProfile {
                name: "High".to_string(),
                bitrate,
                sample_rate: 44100,
                channels: 2,
                description: "Haute qualité pour audiophiles".to_string(),
            },
            _ => QualityProfile {
                name: "Lossless".to_string(),
                bitrate,
                sample_rate: 96000,
                channels: 2,
                description: "Qualité studio sans perte".to_string(),
            },
        }
    }
    
    /// Détermine le preset d'encodage selon le protocole
    fn determine_preset(&self, protocol: &crate::core::StreamProtocol) -> EncodingPreset {
        match protocol {
            crate::core::StreamProtocol::WebRTC { .. } => EncodingPreset::UltraLowLatency {
                max_latency_ms: 20,
                buffer_size: 512,
            },
            crate::core::StreamProtocol::WebSocket { .. } => EncodingPreset::RealTime {
                target_latency_ms: 100,
                quality_priority: false,
            },
            crate::core::StreamProtocol::HLS { .. } => EncodingPreset::HighQuality {
                multi_pass: false,
                noise_reduction: true,
            },
            crate::core::StreamProtocol::DASH { .. } => EncodingPreset::MobileOptimized {
                aggressive_compression: true,
                adaptive_quality: true,
            },
            crate::core::StreamProtocol::RTMP { .. } => EncodingPreset::RealTime {
                target_latency_ms: 2000,
                quality_priority: true,
            },
        }
    }
    
    /// Obtient les métriques d'utilisation
    pub fn get_metrics(&self) -> EncoderMetrics {
        // Clone des métriques atomiques
        EncoderMetrics {
            opus_encodes_total: std::sync::atomic::AtomicU64::new(
                self.metrics.opus_encodes_total.load(std::sync::atomic::Ordering::Relaxed)
            ),
            aac_encodes_total: std::sync::atomic::AtomicU64::new(
                self.metrics.aac_encodes_total.load(std::sync::atomic::Ordering::Relaxed)
            ),
            mp3_encodes_total: std::sync::atomic::AtomicU64::new(
                self.metrics.mp3_encodes_total.load(std::sync::atomic::Ordering::Relaxed)
            ),
            flac_encodes_total: std::sync::atomic::AtomicU64::new(
                self.metrics.flac_encodes_total.load(std::sync::atomic::Ordering::Relaxed)
            ),
            encode_errors_total: std::sync::atomic::AtomicU64::new(
                self.metrics.encode_errors_total.load(std::sync::atomic::Ordering::Relaxed)
            ),
            average_encode_time_ms: std::sync::atomic::AtomicU64::new(
                self.metrics.average_encode_time_ms.load(std::sync::atomic::Ordering::Relaxed)
            ),
            peak_cpu_usage: std::sync::atomic::AtomicU32::new(
                self.metrics.peak_cpu_usage.load(std::sync::atomic::Ordering::Relaxed)
            ),
            memory_usage_mb: std::sync::atomic::AtomicU64::new(
                self.metrics.memory_usage_mb.load(std::sync::atomic::Ordering::Relaxed)
            ),
        }
    }
}

impl EncoderPipeline {
    /// Démarre le traitement en temps réel
    pub async fn start_processing(&mut self) -> Result<(), AppError> {
        if self.processing_thread.is_some() {
            return Err(AppError::AlreadyProcessing);
        }
        
        // TODO: Implémenter le thread de traitement en temps réel
        info!("Pipeline de traitement démarré: {}", self.id);
        Ok(())
    }
    
    /// Arrête le traitement
    pub async fn stop_processing(&mut self) -> Result<(), AppError> {
        if let Some(handle) = self.processing_thread.take() {
            handle.abort();
            info!("Pipeline de traitement arrêté: {}", self.id);
        }
        Ok(())
    }
    
    /// Ajoute un effet audio à la chaîne
    pub fn add_effect(&mut self, effect: Box<dyn AudioEffect + Send + Sync>) {
        self.effects_chain.push(effect);
        debug!("Effet ajouté au pipeline: {}", self.id);
    }
    
    /// Retire un effet par index
    pub fn remove_effect(&mut self, index: usize) -> Option<Box<dyn AudioEffect + Send + Sync>> {
        if index < self.effects_chain.len() {
            Some(self.effects_chain.remove(index))
        } else {
            None
        }
    }
}

/// Profils de qualité prédéfinis pour différents usages
impl QualityProfile {
    /// Profil pour podcasts et voix
    pub fn voice() -> Self {
        Self {
            name: "Voice".to_string(),
            bitrate: 64,
            sample_rate: 22050,
            channels: 1,
            description: "Optimisé pour la voix et podcasts".to_string(),
        }
    }
    
    /// Profil standard pour musique
    pub fn music_standard() -> Self {
        Self {
            name: "Music Standard".to_string(),
            bitrate: 128,
            sample_rate: 44100,
            channels: 2,
            description: "Qualité standard pour musique".to_string(),
        }
    }
    
    /// Profil haute qualité
    pub fn music_high() -> Self {
        Self {
            name: "Music High".to_string(),
            bitrate: 256,
            sample_rate: 44100,
            channels: 2,
            description: "Haute qualité pour audiophiles".to_string(),
        }
    }
    
    /// Profil lossless
    pub fn lossless() -> Self {
        Self {
            name: "Lossless".to_string(),
            bitrate: 1411, // CD quality
            sample_rate: 44100,
            channels: 2,
            description: "Qualité CD sans perte".to_string(),
        }
    }
} 