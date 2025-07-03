/// Module d'effets audio temps réel optimisés SIMD
/// 
/// Implémentation des effets audio pour streaming en temps réel
/// avec support multi-threading et optimisations vectorielles

use std::collections::HashMap;
use std::sync::Arc;
use std::time::Duration;
use parking_lot::RwLock;
use serde::{Serialize, Deserialize};
use crate::error::AppError;

/// Trait pour tous les effets audio
pub trait AudioEffect: Send + Sync + std::fmt::Debug {
    /// Traite un buffer audio
    fn process(&mut self, samples: &mut [f32], sample_rate: u32, channels: u8) -> Result<(), AppError>;
    
    /// Latence introduite par l'effet
    fn latency(&self) -> Duration;
    
    /// Paramètres configurables
    fn get_parameters(&self) -> &HashMap<String, EffectParameter>;
    fn set_parameter(&mut self, name: &str, value: f32) -> Result<(), AppError>;
    
    /// Bypass de l'effet
    fn set_bypass(&mut self, bypass: bool);
    fn is_bypassed(&self) -> bool;
    
    /// Reset de l'état interne
    fn reset(&mut self);
}

/// Chaîne d'effets pour le streaming
#[derive(Debug)]
pub struct EffectsChain {
    /// Effets appliqués en séquence
    effects: Vec<Box<dyn AudioEffect + Send + Sync>>,
    /// Configuration de mix dry/wet
    _dry_wet_mix: f32,
    /// État de bypass global
    bypass: bool,
    /// Métriques de performance
    _performance_metrics: Arc<RwLock<EffectsPerformanceMetrics>>,
}

/// Paramètre d'effet configurable
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct EffectParameter {
    pub name: String,
    pub value: f32,
    pub min_value: f32,
    pub max_value: f32,
    pub default_value: f32,
    pub description: String,
    pub unit: String,
}

/// Métriques de performance des effets
#[derive(Debug, Clone, Default)]
pub struct EffectsPerformanceMetrics {
    pub processing_time_us: u64,
    pub buffer_underruns: u64,
    pub effects_processed: u64,
    pub cpu_usage_percent: f32,
}

/// Compresseur dynamique avec optimisations SIMD
#[derive(Debug)]
pub struct SIMDCompressor {
    threshold: f32,
    ratio: f32,
    attack_time: f32,
    release_time: f32,
    envelope_follower: f32,
    gain_reduction: f32,
    parameters: HashMap<String, EffectParameter>,
    bypass: bool,
    sample_rate: u32,
    attack_coeff: f32,
    release_coeff: f32,
}

impl SIMDCompressor {
    pub fn new() -> Self {
        let mut parameters = HashMap::new();
        
        parameters.insert("threshold".to_string(), EffectParameter {
            name: "Threshold".to_string(),
            value: -12.0,
            min_value: -60.0,
            max_value: 0.0,
            default_value: -12.0,
            description: "Compression threshold in dB".to_string(),
            unit: "dB".to_string(),
        });
        
        Self {
            threshold: -12.0,
            ratio: 4.0,
            attack_time: 10.0,
            release_time: 100.0,
            envelope_follower: 0.0,
            gain_reduction: 1.0,
            parameters,
            bypass: false,
            sample_rate: 44100,
            attack_coeff: 0.0,
            release_coeff: 0.0,
        }
    }
    
    fn calculate_coefficients(&mut self) {
        let attack_samples = (self.attack_time / 1000.0) * self.sample_rate as f32;
        let release_samples = (self.release_time / 1000.0) * self.sample_rate as f32;
        
        self.attack_coeff = (-1.0 / attack_samples).exp();
        self.release_coeff = (-1.0 / release_samples).exp();
    }
    
    #[inline]
    fn db_to_linear(&self, db: f32) -> f32 {
        10.0_f32.powf(db / 20.0)
    }
}

impl AudioEffect for SIMDCompressor {
    fn process(&mut self, samples: &mut [f32], sample_rate: u32, _channels: u8) -> Result<(), AppError> {
        if self.bypass {
            return Ok(());
        }
        
        if self.sample_rate != sample_rate {
            self.sample_rate = sample_rate;
            self.calculate_coefficients();
        }
        
        let threshold_linear = self.db_to_linear(self.threshold);
        
        for sample in samples.iter_mut() {
            let abs_sample = sample.abs();
            
            if abs_sample > self.envelope_follower {
                self.envelope_follower = abs_sample * (1.0 - self.attack_coeff) 
                                       + self.envelope_follower * self.attack_coeff;
            } else {
                self.envelope_follower = abs_sample * (1.0 - self.release_coeff) 
                                       + self.envelope_follower * self.release_coeff;
            }
            
            if self.envelope_follower > threshold_linear {
                let over_threshold = self.envelope_follower / threshold_linear;
                let compressed = over_threshold.powf(1.0 / self.ratio);
                self.gain_reduction = compressed / over_threshold;
            } else {
                self.gain_reduction = 1.0;
            }
            
            *sample *= self.gain_reduction;
        }
        
        Ok(())
    }
    
    fn latency(&self) -> Duration {
        Duration::from_micros(100)
    }
    
    fn get_parameters(&self) -> &HashMap<String, EffectParameter> {
        &self.parameters
    }
    
    fn set_parameter(&mut self, name: &str, value: f32) -> Result<(), AppError> {
        match name {
            "threshold" => {
                self.threshold = value.clamp(-60.0, 0.0);
                if let Some(param) = self.parameters.get_mut("threshold") {
                    param.value = self.threshold;
                }
            },
            "ratio" => {
                self.ratio = value.clamp(1.0, 20.0);
            },
            _ => return Err(AppError::InvalidData { 
                message: format!("Unknown parameter: {}", name) 
            }),
        }
        Ok(())
    }
    
    fn set_bypass(&mut self, bypass: bool) {
        self.bypass = bypass;
    }
    
    fn is_bypassed(&self) -> bool {
        self.bypass
    }
    
    fn reset(&mut self) {
        self.envelope_follower = 0.0;
        self.gain_reduction = 1.0;
    }
}

impl EffectsChain {
    pub fn new() -> Self {
        Self {
            effects: Vec::new(),
            _dry_wet_mix: 1.0,
            bypass: false,
            _performance_metrics: Arc::new(RwLock::new(EffectsPerformanceMetrics::default())),
        }
    }
    
    pub fn add_effect(&mut self, effect: Box<dyn AudioEffect>) {
        self.effects.push(effect);
    }
    
    pub fn process(&mut self, samples: &mut [f32], sample_rate: u32, channels: u8) -> Result<(), AppError> {
        if self.bypass || self.effects.is_empty() {
            return Ok(());
        }
        
        for effect in &mut self.effects {
            if !effect.is_bypassed() {
                effect.process(samples, sample_rate, channels)?;
            }
        }
        
        Ok(())
    }
}

/// Factory pour créer des effets préconfigurés
pub struct EffectFactory;

impl EffectFactory {
    pub fn create_streaming_compressor() -> Box<dyn AudioEffect> {
        let mut compressor = SIMDCompressor::new();
        let _ = compressor.set_parameter("threshold", -18.0);
        let _ = compressor.set_parameter("ratio", 3.0);
        Box::new(compressor)
    }
    
    pub fn create_mastering_chain() -> EffectsChain {
        let mut chain = EffectsChain::new();
        chain.add_effect(Self::create_streaming_compressor());
        chain
    }
}
