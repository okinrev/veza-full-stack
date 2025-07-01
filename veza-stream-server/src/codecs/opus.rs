/// Codec Opus pour streaming ultra low latency
/// 
/// Opus est le codec optimal pour streaming live :
/// - Latence ultra-faible (<5ms)
/// - Qualité excellente à tous les bitrates
/// - Support VBR/CBR adaptatif
/// - Optimisé pour voix et musique

use std::sync::Arc;
use std::time::Instant;
use parking_lot::Mutex;
use serde::{Serialize, Deserialize};
use tracing::{debug, warn, error};

use crate::error::AppError;
use super::{
    AudioEncoder, AudioDecoder, EncoderInfo, DecoderInfo, EncoderMetrics, 
    DecodedAudio, AudioSampleFormat, EncoderConfig, DecoderConfig
};

/// Implémentation de l'encodeur Opus
#[derive(Debug)]
pub struct OpusEncoderImpl {
    /// Configuration de l'encodeur
    config: OpusEncoderConfig,
    /// État interne de l'encodeur (opaque)
    encoder_state: Arc<Mutex<OpusEncoderState>>,
    /// Métriques de performance
    metrics: EncoderMetrics,
    /// Buffer interne pour les échantillons
    sample_buffer: Vec<f32>,
    /// Taille de frame configurée
    frame_size: usize,
}

/// Implémentation du décodeur Opus
#[derive(Debug)]
pub struct OpusDecoderImpl {
    /// Configuration du décodeur
    config: OpusDecoderConfig,
    /// État interne du décodeur (opaque)
    decoder_state: Arc<Mutex<OpusDecoderState>>,
    /// Buffer de sortie
    output_buffer: Vec<f32>,
}

/// Configuration spécifique à Opus pour l'encodeur
#[derive(Debug, Clone)]
pub struct OpusEncoderConfig {
    pub sample_rate: u32,
    pub channels: u8,
    pub bitrate: u32,
    pub complexity: u8,          // 0-10
    pub signal_type: OpusSignalType,
    pub application: OpusApplication,
    pub vbr_mode: OpusVbrMode,
    pub frame_duration: OpusFrameDuration,
    pub packet_loss_resilience: bool,
    pub dtx_enabled: bool,       // Discontinuous Transmission
    pub inband_fec: bool,        // Forward Error Correction
}

/// Configuration spécifique à Opus pour le décodeur
#[derive(Debug, Clone)]
pub struct OpusDecoderConfig {
    pub sample_rate: u32,
    pub channels: u8,
    pub frame_duration: OpusFrameDuration,
    pub gain_db: f32,
}

/// Type de signal pour optimisation Opus
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum OpusSignalType {
    /// Détection automatique
    Auto,
    /// Optimisé pour la voix
    Voice,
    /// Optimisé pour la musique
    Music,
}

/// Application Opus pour différents use cases
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum OpusApplication {
    /// VoIP - optimisé pour la voix
    Voip,
    /// Audio général - équilibré
    Audio,
    /// Low delay - latence minimale
    RestrictedLowDelay,
}

/// Mode Variable Bit Rate
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum OpusVbrMode {
    /// Constant Bit Rate
    CBR,
    /// Variable Bit Rate non contraint
    VBR,
    /// Variable Bit Rate contraint
    CVBR,
}

/// Durée de frame Opus
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum OpusFrameDuration {
    /// 2.5ms - Ultra low latency
    Ms2_5,
    /// 5ms - Very low latency
    Ms5,
    /// 10ms - Low latency (défaut)
    Ms10,
    /// 20ms - Standard
    Ms20,
    /// 40ms - Haute efficacité
    Ms40,
    /// 60ms - Très haute efficacité
    Ms60,
}

/// État interne de l'encodeur (simulation)
#[derive(Debug)]
struct OpusEncoderState {
    initialized: bool,
    last_frame_time: Option<Instant>,
    total_frames: u64,
    total_bytes: u64,
    last_bitrate: u32,
    bandwidth_adaptation: BandwidthAdaptation,
}

/// État interne du décodeur (simulation)
#[derive(Debug)]
struct OpusDecoderState {
    initialized: bool,
    last_decode_time: Option<Instant>,
    total_frames: u64,
    packet_loss_concealment: PacketLossConcealmentState,
}

/// Adaptation de bande passante intelligente
#[derive(Debug)]
struct BandwidthAdaptation {
    current_bandwidth: u32,
    target_bandwidth: u32,
    adaptation_rate: f32,
    quality_history: Vec<f32>,
}

/// État de dissimulation de perte de paquets
#[derive(Debug)]
struct PacketLossConcealmentState {
    last_good_frame: Option<Vec<f32>>,
    consecutive_losses: u32,
    concealment_energy: f32,
}

impl OpusFrameDuration {
    /// Convertit en millisecondes
    pub fn as_ms(&self) -> f32 {
        match self {
            Self::Ms2_5 => 2.5,
            Self::Ms5 => 5.0,
            Self::Ms10 => 10.0,
            Self::Ms20 => 20.0,
            Self::Ms40 => 40.0,
            Self::Ms60 => 60.0,
        }
    }
    
    /// Calcule la taille de frame en échantillons
    pub fn frame_size(&self, sample_rate: u32) -> usize {
        (self.as_ms() * sample_rate as f32 / 1000.0) as usize
    }
}

impl Default for OpusEncoderConfig {
    fn default() -> Self {
        Self {
            sample_rate: 48000,
            channels: 2,
            bitrate: 128_000,
            complexity: 10,
            signal_type: OpusSignalType::Auto,
            application: OpusApplication::Audio,
            vbr_mode: OpusVbrMode::VBR,
            frame_duration: OpusFrameDuration::Ms10,
            packet_loss_resilience: true,
            dtx_enabled: false,
            inband_fec: true,
        }
    }
}

impl Default for OpusDecoderConfig {
    fn default() -> Self {
        Self {
            sample_rate: 48000,
            channels: 2,
            frame_duration: OpusFrameDuration::Ms10,
            gain_db: 0.0,
        }
    }
}

impl OpusEncoderImpl {
    /// Crée un nouvel encodeur Opus
    pub fn new(config: EncoderConfig) -> Result<Self, AppError> {
        let opus_config = OpusEncoderConfig {
            sample_rate: config.sample_rate,
            channels: config.channels,
            bitrate: config.bitrate,
            complexity: config.complexity,
            signal_type: OpusSignalType::Auto,
            application: match config.latency_mode {
                crate::codecs::LatencyMode::UltraLow => OpusApplication::RestrictedLowDelay,
                _ => OpusApplication::Audio,
            },
            vbr_mode: if config.enable_vbr { OpusVbrMode::VBR } else { OpusVbrMode::CBR },
            frame_duration: match config.latency_mode {
                crate::codecs::LatencyMode::UltraLow => OpusFrameDuration::Ms2_5,
                crate::codecs::LatencyMode::Low => OpusFrameDuration::Ms5,
                _ => OpusFrameDuration::Ms10,
            },
            packet_loss_resilience: true,
            dtx_enabled: false,
            inband_fec: true,
        };
        
        let frame_size = opus_config.frame_duration.frame_size(opus_config.sample_rate) * opus_config.channels as usize;
        
        let encoder_state = OpusEncoderState {
            initialized: false,
            last_frame_time: None,
            total_frames: 0,
            total_bytes: 0,
            last_bitrate: opus_config.bitrate,
            bandwidth_adaptation: BandwidthAdaptation {
                current_bandwidth: opus_config.bitrate,
                target_bandwidth: opus_config.bitrate,
                adaptation_rate: 0.1,
                quality_history: Vec::new(),
            },
        };
        
        let mut encoder = Self {
            config: opus_config,
            encoder_state: Arc::new(Mutex::new(encoder_state)),
            metrics: EncoderMetrics::default(),
            sample_buffer: Vec::with_capacity(frame_size * 2),
            frame_size,
        };
        
        encoder.initialize()?;
        Ok(encoder)
    }
    
    /// Initialise l'encodeur Opus
    fn initialize(&mut self) -> Result<(), AppError> {
        // Validation des paramètres
        self.validate_config()?;
        
        // Simulation d'initialisation de l'encodeur Opus natif
        // En production, utiliser opus_encoder_create() de libopus
        {
            let mut state = self.encoder_state.lock();
            state.initialized = true;
        }
        
        debug!("Encodeur Opus initialisé: {}Hz, {} ch, {} bps", 
               self.config.sample_rate, 
               self.config.channels, 
               self.config.bitrate);
        
        Ok(())
    }
    
    /// Valide la configuration
    fn validate_config(&self) -> Result<(), AppError> {
        // Opus supporte spécifiquement 8kHz, 12kHz, 16kHz, 24kHz, 48kHz
        let valid_rates = [8000, 12000, 16000, 24000, 48000];
        if !valid_rates.contains(&self.config.sample_rate) {
            return Err(AppError::InvalidSampleRate {
                codec: "opus".to_string(),
                sample_rate: self.config.sample_rate,
                supported: valid_rates.to_vec(),
            });
        }
        
        // Opus supporte 1 ou 2 channels
        if self.config.channels == 0 || self.config.channels > 2 {
            return Err(AppError::InvalidChannelCount {
                codec: "opus".to_string(),
                channels: self.config.channels,
                supported: vec![1, 2],
            });
        }
        
        // Bitrate valide : 6kbps à 512kbps
        if self.config.bitrate < 6_000 || self.config.bitrate > 512_000 {
            return Err(AppError::InvalidBitrate {
                codec: "opus".to_string(),
                bitrate: self.config.bitrate,
                range: (6_000, 512_000),
            });
        }
        
        Ok(())
    }
    
    /// Adapte automatiquement le bitrate selon les conditions
    fn adapt_bitrate(&mut self, available_bandwidth: u32, packet_loss: f32) -> Result<(), AppError> {
        let mut state = self.encoder_state.lock();
        
        // Calcul du bitrate optimal
        let mut target_bitrate = available_bandwidth * 8 / 10; // 80% de la bande passante
        
        // Réduction si perte de paquets
        if packet_loss > 0.01 { // > 1%
            target_bitrate = (target_bitrate as f32 * (1.0 - packet_loss * 2.0)) as u32;
        }
        
        // Limites du codec
        target_bitrate = target_bitrate.clamp(6_000, 512_000);
        
        if target_bitrate != state.bandwidth_adaptation.current_bandwidth {
            state.bandwidth_adaptation.target_bandwidth = target_bitrate;
            
            // Adaptation progressive
            let diff = target_bitrate as f32 - state.bandwidth_adaptation.current_bandwidth as f32;
            let adjustment = diff * state.bandwidth_adaptation.adaptation_rate;
            state.bandwidth_adaptation.current_bandwidth = 
                (state.bandwidth_adaptation.current_bandwidth as f32 + adjustment) as u32;
            
            debug!("Adaptation bitrate Opus: {} -> {} bps", 
                   state.last_bitrate, state.bandwidth_adaptation.current_bandwidth);
        }
        
        Ok(())
    }
    
    /// Encode avec détection automatique du type de signal
    fn encode_with_signal_detection(&mut self, samples: &[f32]) -> Result<Vec<u8>, AppError> {
        // Analyse simple du signal pour optimiser l'encodage
        let signal_type = self.detect_signal_type(samples);
        
        // Simulation d'encodage Opus
        // En production, utiliser opus_encode() ou opus_encode_float()
        let encoded_size = self.estimate_encoded_size(samples.len());
        let mut encoded_data = vec![0u8; encoded_size];
        
        // Simulation de l'encodage avec compression réaliste
        self.simulate_opus_encoding(samples, &mut encoded_data)?;
        
        // Mise à jour des métriques
        self.update_encoding_metrics(samples.len(), encoded_data.len());
        
        debug!("Frame Opus encodée: {} samples -> {} bytes (type: {:?})", 
               samples.len(), encoded_data.len(), signal_type);
        
        Ok(encoded_data)
    }
    
    /// Détecte le type de signal pour optimiser l'encodage
    fn detect_signal_type(&self, samples: &[f32]) -> OpusSignalType {
        if samples.is_empty() {
            return OpusSignalType::Auto;
        }
        
        // Calcul de métriques simples
        let energy = samples.iter().map(|&s| s * s).sum::<f32>() / samples.len() as f32;
        let zero_crossings = samples.windows(2)
            .filter(|w| (w[0] >= 0.0) != (w[1] >= 0.0))
            .count();
        
        let zcr = zero_crossings as f32 / samples.len() as f32;
        
        // Heuristiques simples
        if zcr > 0.1 && energy > 0.01 {
            OpusSignalType::Voice
        } else if energy > 0.001 {
            OpusSignalType::Music
        } else {
            OpusSignalType::Auto
        }
    }
    
    /// Estime la taille encodée
    fn estimate_encoded_size(&self, sample_count: usize) -> usize {
        let duration_ms = (sample_count as f32 / self.config.sample_rate as f32) * 1000.0;
        let bits_per_ms = self.config.bitrate as f32 / 1000.0;
        let estimated_bits = duration_ms * bits_per_ms;
        (estimated_bits / 8.0) as usize
    }
    
    /// Simulation d'encodage Opus
    fn simulate_opus_encoding(&self, samples: &[f32], output: &mut [u8]) -> Result<(), AppError> {
        // Simulation simple - en production, utiliser libopus
        for (i, &sample) in samples.iter().enumerate() {
            if i / 4 < output.len() {
                // Quantification simulée
                let quantized = (sample * 127.0) as i8;
                output[i / 4] = quantized as u8;
            }
        }
        Ok(())
    }
    
    /// Met à jour les métriques d'encodage
    fn update_encoding_metrics(&mut self, input_samples: usize, output_bytes: usize) {
        self.metrics.frames_encoded += 1;
        self.metrics.bytes_output += output_bytes as u64;
        
        // Calcul du ratio de compression
        let input_bytes = input_samples * 4; // f32 = 4 bytes
        self.metrics.compression_ratio = input_bytes as f32 / output_bytes as f32;
        
        // Mise à jour de l'état
        {
            let mut state = self.encoder_state.lock();
            state.total_frames += 1;
            state.total_bytes += output_bytes as u64;
            state.last_frame_time = Some(Instant::now());
        }
    }
}

impl AudioEncoder for OpusEncoderImpl {
    fn encode(&mut self, samples: &[f32], sample_rate: u32, channels: u8) -> Result<Vec<u8>, AppError> {
        // Vérifier que les paramètres correspondent à la config
        if sample_rate != self.config.sample_rate || channels != self.config.channels {
            return Err(AppError::ParameterMismatch {
                expected: format!("{}Hz, {} ch", self.config.sample_rate, self.config.channels),
                received: format!("{}Hz, {} ch", sample_rate, channels),
            });
        }
        
        // Ajouter les échantillons au buffer
        self.sample_buffer.extend_from_slice(samples);
        
        let mut encoded_frames = Vec::new();
        
        // Traiter les frames complètes
        while self.sample_buffer.len() >= self.frame_size {
            let frame_samples: Vec<f32> = self.sample_buffer.drain(..self.frame_size).collect();
            let encoded_frame = self.encode_with_signal_detection(&frame_samples)?;
            encoded_frames.extend(encoded_frame);
        }
        
        Ok(encoded_frames)
    }
    
    fn finalize(&mut self) -> Result<Vec<u8>, AppError> {
        // Encoder le buffer restant avec padding si nécessaire
        if !self.sample_buffer.is_empty() {
            // Padding avec du silence
            while self.sample_buffer.len() < self.frame_size {
                self.sample_buffer.push(0.0);
            }
            
            let final_frame: Vec<f32> = self.sample_buffer.drain(..).collect();
            return self.encode_with_signal_detection(&final_frame);
        }
        
        Ok(Vec::new())
    }
    
    fn reset(&mut self) -> Result<(), AppError> {
        self.sample_buffer.clear();
        self.metrics = EncoderMetrics::default();
        
        {
            let mut state = self.encoder_state.lock();
            state.total_frames = 0;
            state.total_bytes = 0;
            state.last_frame_time = None;
        }
        
        debug!("Encodeur Opus remis à zéro");
        Ok(())
    }
    
    fn set_bitrate(&mut self, bitrate: u32) -> Result<(), AppError> {
        if bitrate < 6_000 || bitrate > 512_000 {
            return Err(AppError::InvalidBitrate {
                codec: "opus".to_string(),
                bitrate,
                range: (6_000, 512_000),
            });
        }
        
        self.config.bitrate = bitrate;
        
        {
            let mut state = self.encoder_state.lock();
            state.last_bitrate = bitrate;
            state.bandwidth_adaptation.current_bandwidth = bitrate;
            state.bandwidth_adaptation.target_bandwidth = bitrate;
        }
        
        debug!("Bitrate Opus mis à jour: {} bps", bitrate);
        Ok(())
    }
    
    fn info(&self) -> EncoderInfo {
        EncoderInfo {
            codec_name: "Opus".to_string(),
            version: "1.3.1".to_string(),
            bitrate: self.config.bitrate,
            sample_rate: self.config.sample_rate,
            channels: self.config.channels,
            bit_depth: 16, // Opus utilise toujours 16-bit en interne
            frame_size: self.frame_size,
            latency_ms: self.config.frame_duration.as_ms(),
            quality_mode: format!("{:?}", self.config.vbr_mode),
        }
    }
    
    fn metrics(&self) -> EncoderMetrics {
        self.metrics.clone()
    }
}

impl OpusDecoderImpl {
    /// Crée un nouveau décodeur Opus
    pub fn new(config: DecoderConfig) -> Result<Self, AppError> {
        let opus_config = OpusDecoderConfig {
            sample_rate: config.sample_rate,
            channels: config.channels,
            frame_duration: OpusFrameDuration::Ms10,
            gain_db: 0.0,
        };
        
        let decoder_state = OpusDecoderState {
            initialized: false,
            last_decode_time: None,
            total_frames: 0,
            packet_loss_concealment: PacketLossConcealmentState {
                last_good_frame: None,
                consecutive_losses: 0,
                concealment_energy: 0.0,
            },
        };
        
        let output_buffer_size = opus_config.frame_duration.frame_size(opus_config.sample_rate) * opus_config.channels as usize;
        
        let mut decoder = Self {
            config: opus_config,
            decoder_state: Arc::new(Mutex::new(decoder_state)),
            output_buffer: Vec::with_capacity(output_buffer_size * 2),
        };
        
        decoder.initialize()?;
        Ok(decoder)
    }
    
    /// Initialise le décodeur
    fn initialize(&mut self) -> Result<(), AppError> {
        {
            let mut state = self.decoder_state.lock();
            state.initialized = true;
        }
        
        debug!("Décodeur Opus initialisé: {}Hz, {} ch", 
               self.config.sample_rate, self.config.channels);
        Ok(())
    }
    
    /// Décode avec dissimulation de perte de paquets
    fn decode_with_plc(&mut self, data: Option<&[u8]>) -> Result<DecodedAudio, AppError> {
        let mut state = self.decoder_state.lock();
        
        match data {
            Some(encoded_data) => {
                // Décodage normal
                let decoded_samples = self.simulate_opus_decoding(encoded_data)?;
                
                // Sauvegarder pour PLC
                state.packet_loss_concealment.last_good_frame = Some(decoded_samples.clone());
                state.packet_loss_concealment.consecutive_losses = 0;
                state.total_frames += 1;
                state.last_decode_time = Some(Instant::now());
                
                Ok(DecodedAudio {
                    samples: decoded_samples,
                    sample_rate: self.config.sample_rate,
                    channels: self.config.channels,
                    duration_ms: self.config.frame_duration.as_ms() as u32,
                    format: AudioSampleFormat::F32,
                })
            }
            None => {
                // Perte de paquet - générer du contenu de remplacement
                state.packet_loss_concealment.consecutive_losses += 1;
                
                let concealed_samples = self.generate_packet_loss_concealment(&mut state)?;
                
                Ok(DecodedAudio {
                    samples: concealed_samples,
                    sample_rate: self.config.sample_rate,
                    channels: self.config.channels,
                    duration_ms: self.config.frame_duration.as_ms() as u32,
                    format: AudioSampleFormat::F32,
                })
            }
        }
    }
    
    /// Simulation de décodage Opus
    fn simulate_opus_decoding(&self, data: &[u8]) -> Result<Vec<f32>, AppError> {
        let frame_size = self.config.frame_duration.frame_size(self.config.sample_rate) * self.config.channels as usize;
        let mut samples = Vec::with_capacity(frame_size);
        
        // Simulation simple - en production, utiliser opus_decode()
        for (i, &byte) in data.iter().enumerate() {
            if samples.len() < frame_size {
                // Dé-quantification simulée
                let sample = (byte as i8) as f32 / 127.0;
                samples.push(sample);
                
                // Interpolation pour atteindre la taille de frame
                if i < data.len() - 1 {
                    for _ in 0..3 {
                        if samples.len() < frame_size {
                            samples.push(sample * 0.8); // Interpolation simple
                        }
                    }
                }
            }
        }
        
        // Compléter avec du silence si nécessaire
        while samples.len() < frame_size {
            samples.push(0.0);
        }
        
        Ok(samples)
    }
    
    /// Génère une dissimulation de perte de paquet
    fn generate_packet_loss_concealment(&self, state: &mut OpusDecoderState) -> Result<Vec<f32>, AppError> {
        let frame_size = self.config.frame_duration.frame_size(self.config.sample_rate) * self.config.channels as usize;
        
        if let Some(ref last_frame) = state.packet_loss_concealment.last_good_frame {
            // Atténuation progressive basée sur le nombre de pertes consécutives
            let attenuation = 0.8_f32.powf(state.packet_loss_concealment.consecutive_losses as f32);
            
            let mut concealed = Vec::with_capacity(frame_size);
            for (i, &sample) in last_frame.iter().enumerate() {
                if concealed.len() < frame_size {
                    // Atténuation avec légère variation pour éviter les artefacts
                    let variation = (i as f32 * 0.1).sin() * 0.01;
                    concealed.push(sample * attenuation + variation);
                }
            }
            
            // Compléter si nécessaire
            while concealed.len() < frame_size {
                concealed.push(0.0);
            }
            
            Ok(concealed)
        } else {
            // Pas de frame précédente, générer du silence
            Ok(vec![0.0; frame_size])
        }
    }
}

impl AudioDecoder for OpusDecoderImpl {
    fn decode(&mut self, data: &[u8]) -> Result<DecodedAudio, AppError> {
        if data.is_empty() {
            // Perte de paquet
            self.decode_with_plc(None)
        } else {
            self.decode_with_plc(Some(data))
        }
    }
    
    fn reset(&mut self) -> Result<(), AppError> {
        self.output_buffer.clear();
        
        {
            let mut state = self.decoder_state.lock();
            state.total_frames = 0;
            state.last_decode_time = None;
            state.packet_loss_concealment.last_good_frame = None;
            state.packet_loss_concealment.consecutive_losses = 0;
            state.packet_loss_concealment.concealment_energy = 0.0;
        }
        
        debug!("Décodeur Opus remis à zéro");
        Ok(())
    }
    
    fn info(&self) -> DecoderInfo {
        DecoderInfo {
            codec_name: "Opus".to_string(),
            version: "1.3.1".to_string(),
            sample_rate: self.config.sample_rate,
            channels: self.config.channels,
            bit_depth: 16,
            frame_size: self.config.frame_duration.frame_size(self.config.sample_rate) * self.config.channels as usize,
        }
    }
} 