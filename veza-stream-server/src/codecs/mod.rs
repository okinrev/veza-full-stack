/// Modules de codecs audio pour streaming production
/// 
/// Codecs supportés :
/// - Opus : Ultra low latency, qualité optimale pour streaming live
/// - AAC : Compatibilité universelle iOS/Safari/mobile
/// - MP3 : Compatibilité legacy et universelle  
/// - FLAC : Qualité lossless pour premium/studio

pub mod opus;
pub mod aac;
pub mod mp3;
pub mod flac;

// Re-exports pour faciliter l'usage
pub use opus::*;
pub use aac::*;
pub use mp3::*;
pub use flac::*;

use std::fmt;
use serde::{Serialize, Deserialize};
use crate::error::AppError;

/// Trait unifié pour tous les encodeurs audio
pub trait AudioEncoder: fmt::Debug + Send + Sync {
    /// Encode des échantillons audio en format spécifique
    fn encode(&mut self, samples: &[f32], sample_rate: u32, channels: u8) -> Result<Vec<u8>, AppError>;
    
    /// Finalise l'encodage et retourne les derniers bytes
    fn finalize(&mut self) -> Result<Vec<u8>, AppError>;
    
    /// Remet l'encodeur à zéro pour réutilisation
    fn reset(&mut self) -> Result<(), AppError>;
    
    /// Configure le bitrate cible
    fn set_bitrate(&mut self, bitrate: u32) -> Result<(), AppError>;
    
    /// Obtient les informations de l'encodeur
    fn info(&self) -> EncoderInfo;
    
    /// Obtient les métriques de performance
    fn metrics(&self) -> EncoderMetrics;
}

/// Trait unifié pour tous les décodeurs audio
pub trait AudioDecoder: fmt::Debug + Send + Sync {
    /// Décode des bytes en échantillons audio
    fn decode(&mut self, data: &[u8]) -> Result<DecodedAudio, AppError>;
    
    /// Remet le décodeur à zéro
    fn reset(&mut self) -> Result<(), AppError>;
    
    /// Obtient les informations du décodeur
    fn info(&self) -> DecoderInfo;
}

/// Informations sur un encodeur
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct EncoderInfo {
    pub codec_name: String,
    pub version: String,
    pub bitrate: u32,
    pub sample_rate: u32,
    pub channels: u8,
    pub bit_depth: u8,
    pub frame_size: usize,
    pub latency_ms: f32,
    pub quality_mode: String,
}

/// Informations sur un décodeur
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DecoderInfo {
    pub codec_name: String,
    pub version: String,
    pub sample_rate: u32,
    pub channels: u8,
    pub bit_depth: u8,
    pub frame_size: usize,
}

/// Métriques de performance d'un encodeur
#[derive(Debug, Clone, Default)]
pub struct EncoderMetrics {
    pub frames_encoded: u64,
    pub bytes_output: u64,
    pub encoding_time_ms: u64,
    pub cpu_usage_percent: f32,
    pub memory_usage_mb: f32,
    pub compression_ratio: f32,
    pub quality_score: f32,
}

/// Audio décodé avec métadonnées
#[derive(Debug, Clone)]
pub struct DecodedAudio {
    pub samples: Vec<f32>,
    pub sample_rate: u32,
    pub channels: u8,
    pub duration_ms: u32,
    pub format: AudioSampleFormat,
}

/// Format des échantillons audio
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum AudioSampleFormat {
    F32,  // 32-bit float
    I16,  // 16-bit integer
    I24,  // 24-bit integer
    I32,  // 32-bit integer
}

/// Factory pour créer des encodeurs/décodeurs
pub struct CodecFactory;

impl CodecFactory {
    /// Crée un encodeur selon le codec demandé
    pub fn create_encoder(codec: &str, config: EncoderConfig) -> Result<Box<dyn AudioEncoder>, AppError> {
        match codec.to_lowercase().as_str() {
            "opus" => Ok(Box::new(opus::OpusEncoderImpl::new(config)?)),
            "aac" => Ok(Box::new(aac::AacEncoderImpl::new(config)?)),
            "mp3" => {
                let mp3_config = Self::convert_to_mp3_encoder_config(config);
                Ok(Box::new(mp3::Mp3EncoderImpl::new(mp3_config)))
            },
            "flac" => Ok(Box::new(flac::FlacEncoderImpl::new(config)?)),
            _ => Err(AppError::UnsupportedCodec { codec: codec.to_string() }),
        }
    }
    
    /// Crée un décodeur selon le codec demandé
    pub fn create_decoder(codec: &str, config: DecoderConfig) -> Result<Box<dyn AudioDecoder>, AppError> {
        match codec.to_lowercase().as_str() {
            "opus" => Ok(Box::new(opus::OpusDecoderImpl::new(config)?)),
            "aac" => Ok(Box::new(aac::AacDecoderImpl::new(config)?)),
            "mp3" => {
                let mp3_config = Self::convert_to_mp3_decoder_config(config);
                Ok(Box::new(mp3::Mp3DecoderImpl::new(mp3_config)))
            },
            "flac" => Ok(Box::new(flac::FlacDecoderImpl::new(config)?)),
            _ => Err(AppError::UnsupportedCodec { codec: codec.to_string() }),
        }
    }
    
    /// Liste tous les codecs supportés
    pub fn supported_codecs() -> Vec<CodecInfo> {
        vec![
            CodecInfo {
                name: "opus".to_string(),
                description: "Opus - Ultra low latency, optimal for live streaming".to_string(),
                mime_types: vec!["audio/opus".to_string()],
                file_extensions: vec!["opus".to_string()],
                bitrate_range: (6_000, 512_000),
                sample_rates: vec![8000, 12000, 16000, 24000, 48000],
                channels_supported: vec![1, 2],
                is_lossless: false,
                latency_ms: 2.5,
            },
            CodecInfo {
                name: "aac".to_string(),
                description: "AAC - Universal compatibility, iOS/Safari optimized".to_string(),
                mime_types: vec!["audio/aac".to_string(), "audio/mp4".to_string()],
                file_extensions: vec!["aac".to_string(), "m4a".to_string()],
                bitrate_range: (32_000, 320_000),
                sample_rates: vec![8000, 16000, 22050, 44100, 48000, 96000],
                channels_supported: vec![1, 2, 6, 8],
                is_lossless: false,
                latency_ms: 20.0,
            },
            CodecInfo {
                name: "mp3".to_string(),
                description: "MP3 - Legacy compatibility, universal support".to_string(),
                mime_types: vec!["audio/mpeg".to_string()],
                file_extensions: vec!["mp3".to_string()],
                bitrate_range: (32_000, 320_000),
                sample_rates: vec![8000, 11025, 12000, 16000, 22050, 24000, 32000, 44100, 48000],
                channels_supported: vec![1, 2],
                is_lossless: false,
                latency_ms: 26.0,
            },
            CodecInfo {
                name: "flac".to_string(),
                description: "FLAC - Lossless compression, studio quality".to_string(),
                mime_types: vec!["audio/flac".to_string()],
                file_extensions: vec!["flac".to_string()],
                bitrate_range: (700_000, 1_411_000),
                sample_rates: vec![8000, 16000, 22050, 44100, 48000, 88200, 96000, 176400, 192000],
                channels_supported: vec![1, 2, 6, 8],
                is_lossless: true,
                latency_ms: 40.0,
            },
        ]
    }
    
    /// Convertit EncoderConfig vers Mp3EncoderConfig
    fn convert_to_mp3_encoder_config(config: EncoderConfig) -> mp3::Mp3EncoderConfig {
        let quality_preset = match config.quality {
            CodecQuality::Low => mp3::Mp3QualityPreset::Economy,
            CodecQuality::Medium => mp3::Mp3QualityPreset::Standard,
            CodecQuality::High => mp3::Mp3QualityPreset::Extreme,
            CodecQuality::VeryHigh => mp3::Mp3QualityPreset::Insane,
            CodecQuality::Custom(_) => mp3::Mp3QualityPreset::Standard,
        };
        
        let encoding_mode = if config.enable_vbr {
            mp3::Mp3EncodingMode::VBR
        } else {
            mp3::Mp3EncodingMode::CBR
        };
        
        mp3::Mp3EncoderConfig {
            encoding_mode,
            bitrate: config.bitrate / 1000, // Convert to kbps
            vbr_quality: match config.quality {
                CodecQuality::Low => 7,
                CodecQuality::Medium => 4,
                CodecQuality::High => 2,
                CodecQuality::VeryHigh => 0,
                CodecQuality::Custom(f) => ((1.0 - f) * 9.0) as u8,
            },
            sample_rate: config.sample_rate,
            channels: config.channels,
            quality_preset,
            joint_stereo: config.channels == 2,
            error_protection: false,
            include_id3: true,
            copyright: false,
            original: true,
        }
    }
    
    /// Convertit DecoderConfig vers Mp3DecoderConfig
    fn convert_to_mp3_decoder_config(_config: DecoderConfig) -> mp3::Mp3DecoderConfig {
        mp3::Mp3DecoderConfig {
            frame_buffer_size: 1024,
            enable_seeking: true,
            error_tolerance: mp3::Mp3ErrorTolerance::Tolerant,
            gapless_playback: true,
            auto_eq: false,
        }
    }
}

/// Configuration pour un encodeur
#[derive(Debug, Clone)]
pub struct EncoderConfig {
    pub bitrate: u32,
    pub sample_rate: u32,
    pub channels: u8,
    pub quality: CodecQuality,
    pub latency_mode: LatencyMode,
    pub enable_vbr: bool,
    pub complexity: u8, // 0-10, plus élevé = meilleure qualité
}

/// Configuration pour un décodeur
#[derive(Debug, Clone)]
pub struct DecoderConfig {
    pub sample_rate: u32,
    pub channels: u8,
    pub output_format: AudioSampleFormat,
}

/// Qualité du codec
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum CodecQuality {
    Low,
    Medium,
    High,
    VeryHigh,
    Custom(f32), // 0.0-1.0
}

/// Mode de latence
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum LatencyMode {
    UltraLow,  // <5ms
    Low,       // <20ms
    Normal,    // <50ms
    High,      // >50ms (pour qualité max)
}

/// Informations sur un codec
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CodecInfo {
    pub name: String,
    pub description: String,
    pub mime_types: Vec<String>,
    pub file_extensions: Vec<String>,
    pub bitrate_range: (u32, u32), // (min, max) en bps
    pub sample_rates: Vec<u32>,
    pub channels_supported: Vec<u8>,
    pub is_lossless: bool,
    pub latency_ms: f32,
}

impl Default for EncoderConfig {
    fn default() -> Self {
        Self {
            bitrate: 128_000,
            sample_rate: 44100,
            channels: 2,
            quality: CodecQuality::Medium,
            latency_mode: LatencyMode::Normal,
            enable_vbr: true,
            complexity: 5,
        }
    }
}

impl Default for DecoderConfig {
    fn default() -> Self {
        Self {
            sample_rate: 44100,
            channels: 2,
            output_format: AudioSampleFormat::F32,
        }
    }
}

/// Utilitaires pour les codecs
pub mod utils {
    use super::*;
    
    /// Convertit les échantillons entre différents formats
    pub fn convert_samples(
        samples: &[f32],
        from_format: AudioSampleFormat,
        to_format: AudioSampleFormat,
    ) -> Vec<u8> {
        match (from_format, to_format) {
            (AudioSampleFormat::F32, AudioSampleFormat::I16) => {
                samples.iter()
                    .flat_map(|&sample| {
                        let scaled = (sample * 32767.0).clamp(-32768.0, 32767.0) as i16;
                        scaled.to_le_bytes()
                    })
                    .collect()
            }
            (AudioSampleFormat::F32, AudioSampleFormat::I24) => {
                samples.iter()
                    .flat_map(|&sample| {
                        let scaled = (sample * 8388607.0).clamp(-8388608.0, 8388607.0) as i32;
                        let bytes = scaled.to_le_bytes();
                        [bytes[0], bytes[1], bytes[2]]
                    })
                    .collect()
            }
            (AudioSampleFormat::F32, AudioSampleFormat::I32) => {
                samples.iter()
                    .flat_map(|&sample| {
                        let scaled = (sample * 2147483647.0).clamp(-2147483648.0, 2147483647.0) as i32;
                        scaled.to_le_bytes()
                    })
                    .collect()
            }
            _ => {
                // Format identique ou non supporté, retourner les bytes bruts
                samples.iter()
                    .flat_map(|&sample| sample.to_le_bytes())
                    .collect()
            }
        }
    }
    
    /// Calcule la compression ratio
    pub fn calculate_compression_ratio(input_size: usize, output_size: usize) -> f32 {
        if output_size == 0 {
            return 0.0;
        }
        input_size as f32 / output_size as f32
    }
    
    /// Valide les paramètres d'un codec
    pub fn validate_codec_params(
        codec_name: &str,
        sample_rate: u32,
        channels: u8,
        bitrate: u32,
    ) -> Result<(), AppError> {
        let codecs = CodecFactory::supported_codecs();
        let codec_info = codecs.iter()
            .find(|c| c.name == codec_name)
            .ok_or_else(|| AppError::UnsupportedCodec { 
                codec: codec_name.to_string() 
            })?;
        
        // Vérifier sample rate
        if !codec_info.sample_rates.contains(&sample_rate) {
            return Err(AppError::InvalidSampleRate { 
                rate: sample_rate,
            });
        }
        
        // Vérifier channels
        if !codec_info.channels_supported.contains(&channels) {
            return Err(AppError::InvalidChannelCount { 
                channels,
            });
        }
        
        // Vérifier bitrate
        let (min_bitrate, max_bitrate) = codec_info.bitrate_range;
        if bitrate < min_bitrate || bitrate > max_bitrate {
            return Err(AppError::InvalidBitrate { 
                bitrate,
                codec: codec_name.to_string(),
            });
        }
        
        Ok(())
    }
}

 
