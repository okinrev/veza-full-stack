/// Module de traitement audio en temps réel pour le streaming
/// 
/// Implémentation des buffers circulaires, resampling adaptatif
/// et gestion de la latence ultra-faible pour streaming live

use std::collections::VecDeque;
use std::sync::Arc;
use std::time::{Duration, Instant, SystemTime};
use parking_lot::{RwLock, Mutex};
use serde::{Serialize, Deserialize};
use crate::error::AppError;
use crate::audio::effects::{EffectsChain, AudioEffect};

/// Buffer circulaire thread-safe pour audio temps réel
#[derive(Debug)]
pub struct RingBuffer<T> {
    buffer: Vec<T>,
    write_index: usize,
    read_index: usize,
    size: usize,
    filled: bool,
}

impl<T: Clone + Default> RingBuffer<T> {
    pub fn new(size: usize) -> Self {
        Self {
            buffer: vec![T::default(); size],
            write_index: 0,
            read_index: 0,
            size,
            filled: false,
        }
    }
    
    pub fn write(&mut self, data: &[T]) -> usize {
        let mut written = 0;
        
        for &item in data.iter() {
            if self.is_full() && !self.filled {
                break;
            }
            
            self.buffer[self.write_index] = item.clone();
            self.write_index = (self.write_index + 1) % self.size;
            written += 1;
            
            if self.write_index == self.read_index {
                self.filled = true;
            }
        }
        
        written
    }
    
    pub fn read(&mut self, data: &mut [T]) -> usize {
        let mut read = 0;
        
        for item in data.iter_mut() {
            if self.is_empty() {
                break;
            }
            
            *item = self.buffer[self.read_index].clone();
            self.read_index = (self.read_index + 1) % self.size;
            read += 1;
            
            if self.read_index == self.write_index {
                self.filled = false;
            }
        }
        
        read
    }
    
    pub fn is_empty(&self) -> bool {
        !self.filled && self.read_index == self.write_index
    }
    
    pub fn is_full(&self) -> bool {
        self.filled
    }
    
    pub fn available_read(&self) -> usize {
        if self.is_empty() {
            0
        } else if self.filled {
            self.size
        } else if self.write_index >= self.read_index {
            self.write_index - self.read_index
        } else {
            self.size - self.read_index + self.write_index
        }
    }
    
    pub fn available_write(&self) -> usize {
        self.size - self.available_read()
    }
}

/// Processeur audio temps réel multi-threaded
#[derive(Debug)]
pub struct RealtimeAudioProcessor {
    /// Buffer d'entrée
    input_buffer: Arc<Mutex<RingBuffer<f32>>>,
    /// Buffer de sortie
    output_buffer: Arc<Mutex<RingBuffer<f32>>>,
    /// Chaîne d'effets
    effects_chain: Arc<Mutex<EffectsChain>>,
    /// Configuration
    config: RealtimeConfig,
    /// Métriques temps réel
    metrics: Arc<RwLock<RealtimeMetrics>>,
    /// Thread de traitement
    processing_thread: Option<std::thread::JoinHandle<()>>,
    /// Signal d'arrêt
    shutdown_signal: Arc<std::sync::atomic::AtomicBool>,
}

/// Configuration du processeur temps réel
#[derive(Debug, Clone)]
pub struct RealtimeConfig {
    pub sample_rate: u32,
    pub channels: u8,
    pub buffer_size: usize,
    pub max_latency_ms: f32,
    pub enable_adaptive_buffering: bool,
    pub enable_jitter_compensation: bool,
    pub thread_priority: ThreadPriority,
}

/// Priorité des threads audio
#[derive(Debug, Clone)]
pub enum ThreadPriority {
    Normal,
    High,
    Realtime,
}

/// Métriques temps réel
#[derive(Debug, Clone, Default)]
pub struct RealtimeMetrics {
    /// Latence actuelle en microsecondes
    pub current_latency_us: u64,
    /// Latence moyenne sur 1 seconde
    pub average_latency_us: u64,
    /// Pic de latence
    pub peak_latency_us: u64,
    /// Underruns (buffer vide)
    pub buffer_underruns: u64,
    /// Overruns (buffer plein)
    pub buffer_overruns: u64,
    /// Samples traités
    pub samples_processed: u64,
    /// Utilisation CPU du thread audio
    pub cpu_usage_percent: f32,
    /// Jitter (variation de latence)
    pub jitter_us: u64,
    /// Qualité du signal (SNR)
    pub signal_quality_db: f32,
}

/// Resampler adaptatif pour compensation de drift
#[derive(Debug)]
pub struct AdaptiveResampler {
    /// Ratio de resampling actuel
    ratio: f64,
    /// Ratio cible
    target_ratio: f64,
    /// Coefficients de filtre anti-aliasing
    filter_coeffs: Vec<f32>,
    /// Historique pour interpolation
    history: VecDeque<f32>,
    /// Configuration
    config: ResamplerConfig,
}

/// Configuration du resampler
#[derive(Debug, Clone)]
pub struct ResamplerConfig {
    pub max_ratio_deviation: f64,
    pub adaptation_speed: f64,
    pub filter_quality: FilterQuality,
    pub enable_anti_aliasing: bool,
}

#[derive(Debug, Clone)]
pub enum FilterQuality {
    Low,      // Rapide, qualité basique
    Medium,   // Compromis équilibré
    High,     // Haute qualité, plus de CPU
    Audiophile, // Qualité maximale
}

/// Gestionnaire de latence adaptative
#[derive(Debug)]
pub struct LatencyManager {
    /// Buffer adaptatif
    adaptive_buffer: RingBuffer<f32>,
    /// Latence cible
    target_latency_ms: f32,
    /// Latence mesurée
    measured_latency_ms: f32,
    /// Historique des latences
    latency_history: VecDeque<f32>,
    /// Contrôleur PID pour ajustement
    pid_controller: PIDController,
}

/// Contrôleur PID pour latence
#[derive(Debug)]
pub struct PIDController {
    kp: f32, // Proportionnel
    ki: f32, // Intégral
    kd: f32, // Dérivé
    integral: f32,
    previous_error: f32,
    last_update: Instant,
}

impl RealtimeAudioProcessor {
    pub fn new(config: RealtimeConfig) -> Result<Self, AppError> {
        let buffer_size = config.buffer_size;
        
        Ok(Self {
            input_buffer: Arc::new(Mutex::new(RingBuffer::new(buffer_size * 4))),
            output_buffer: Arc::new(Mutex::new(RingBuffer::new(buffer_size * 4))),
            effects_chain: Arc::new(Mutex::new(EffectsChain::new())),
            config,
            metrics: Arc::new(RwLock::new(RealtimeMetrics::default())),
            processing_thread: None,
            shutdown_signal: Arc::new(std::sync::atomic::AtomicBool::new(false)),
        })
    }
    
    /// Démarre le traitement temps réel
    pub fn start(&mut self) -> Result<(), AppError> {
        if self.processing_thread.is_some() {
            return Err(AppError::AlreadyRunning);
        }
        
        self.shutdown_signal.store(false, std::sync::atomic::Ordering::Relaxed);
        
        let input_buffer = self.input_buffer.clone();
        let output_buffer = self.output_buffer.clone();
        let effects_chain = self.effects_chain.clone();
        let config = self.config.clone();
        let metrics = self.metrics.clone();
        let shutdown = self.shutdown_signal.clone();
        
        let handle = std::thread::Builder::new()
            .name("realtime-audio".to_string())
            .spawn(move || {
                Self::processing_loop(
                    input_buffer,
                    output_buffer,
                    effects_chain,
                    config,
                    metrics,
                    shutdown,
                );
            })
            .map_err(|e| AppError::ThreadError { 
                message: format!("Failed to start audio thread: {}", e) 
            })?;
        
        self.processing_thread = Some(handle);
        Ok(())
    }
    
    /// Arrête le traitement
    pub fn stop(&mut self) -> Result<(), AppError> {
        if let Some(handle) = self.processing_thread.take() {
            self.shutdown_signal.store(true, std::sync::atomic::Ordering::Relaxed);
            handle.join().map_err(|_| AppError::ThreadError { 
                message: "Failed to join audio thread".to_string() 
            })?;
        }
        Ok(())
    }
    
    /// Boucle de traitement principale
    fn processing_loop(
        input_buffer: Arc<Mutex<RingBuffer<f32>>>,
        output_buffer: Arc<Mutex<RingBuffer<f32>>>,
        effects_chain: Arc<Mutex<EffectsChain>>,
        config: RealtimeConfig,
        metrics: Arc<RwLock<RealtimeMetrics>>,
        shutdown: Arc<std::sync::atomic::AtomicBool>,
    ) {
        let mut processing_buffer = vec![0.0f32; config.buffer_size];
        let frame_duration = Duration::from_micros(
            (config.buffer_size as u64 * 1_000_000) / config.sample_rate as u64
        );
        
        while !shutdown.load(std::sync::atomic::Ordering::Relaxed) {
            let start_time = Instant::now();
            
            // Lecture depuis le buffer d'entrée
            let samples_read = {
                let mut input = input_buffer.lock();
                input.read(&mut processing_buffer)
            };
            
            if samples_read == 0 {
                // Buffer underrun
                let mut metrics_guard = metrics.write();
                metrics_guard.buffer_underruns += 1;
                
                // Attendre un peu avant de réessayer
                std::thread::sleep(Duration::from_micros(100));
                continue;
            }
            
            // Traitement des effets
            {
                let mut effects = effects_chain.lock();
                if let Err(e) = effects.process(
                    &mut processing_buffer[..samples_read],
                    config.sample_rate,
                    config.channels,
                ) {
                    eprintln!("Effect processing error: {:?}", e);
                }
            }
            
            // Écriture vers le buffer de sortie
            let samples_written = {
                let mut output = output_buffer.lock();
                output.write(&processing_buffer[..samples_read])
            };
            
            if samples_written < samples_read {
                // Buffer overrun
                let mut metrics_guard = metrics.write();
                metrics_guard.buffer_overruns += 1;
            }
            
            // Mise à jour des métriques
            let processing_time = start_time.elapsed();
            let mut metrics_guard = metrics.write();
            metrics_guard.current_latency_us = processing_time.as_micros() as u64;
            metrics_guard.samples_processed += samples_read as u64;
            
            // Calcul utilisation CPU (approximation)
            let cpu_usage = (processing_time.as_micros() as f32 / frame_duration.as_micros() as f32) * 100.0;
            metrics_guard.cpu_usage_percent = cpu_usage.min(100.0);
            
            // Attendre pour maintenir le timing
            if processing_time < frame_duration {
                std::thread::sleep(frame_duration - processing_time);
            }
        }
    }
    
    /// Ajoute des données audio à traiter
    pub fn write_input(&self, samples: &[f32]) -> Result<usize, AppError> {
        let mut input = self.input_buffer.lock();
        Ok(input.write(samples))
    }
    
    /// Lit les données audio traitées
    pub fn read_output(&self, buffer: &mut [f32]) -> Result<usize, AppError> {
        let mut output = self.output_buffer.lock();
        Ok(output.read(buffer))
    }
    
    /// Obtient les métriques actuelles
    pub fn get_metrics(&self) -> RealtimeMetrics {
        self.metrics.read().clone()
    }
    
    /// Ajoute un effet à la chaîne
    pub fn add_effect(&self, effect: Box<dyn AudioEffect>) {
        let mut effects = self.effects_chain.lock();
        effects.add_effect(effect);
    }
}

impl AdaptiveResampler {
    pub fn new(config: ResamplerConfig) -> Self {
        Self {
            ratio: 1.0,
            target_ratio: 1.0,
            filter_coeffs: Self::generate_filter_coeffs(&config),
            history: VecDeque::with_capacity(64),
            config,
        }
    }
    
    fn generate_filter_coeffs(config: &ResamplerConfig) -> Vec<f32> {
        match config.filter_quality {
            FilterQuality::Low => vec![1.0], // Pas de filtrage
            FilterQuality::Medium => vec![0.25, 0.5, 0.25], // Filtre simple
            FilterQuality::High => {
                // Filtre Kaiser-Bessel
                let taps = 32;
                let mut coeffs = Vec::with_capacity(taps);
                for i in 0..taps {
                    let x = (i as f32 - taps as f32 / 2.0) / (taps as f32 / 2.0);
                    let sinc = if x == 0.0 { 1.0 } else { (std::f32::consts::PI * x).sin() / (std::f32::consts::PI * x) };
                    let window = 0.54 - 0.46 * (2.0 * std::f32::consts::PI * i as f32 / (taps - 1) as f32).cos();
                    coeffs.push(sinc * window);
                }
                coeffs
            },
            FilterQuality::Audiophile => {
                // Filtre très haute qualité (plus de taps)
                let taps = 128;
                let mut coeffs = Vec::with_capacity(taps);
                for i in 0..taps {
                    let x = (i as f32 - taps as f32 / 2.0) / (taps as f32 / 2.0);
                    let sinc = if x == 0.0 { 1.0 } else { (std::f32::consts::PI * x).sin() / (std::f32::consts::PI * x) };
                    let window = 0.54 - 0.46 * (2.0 * std::f32::consts::PI * i as f32 / (taps - 1) as f32).cos();
                    coeffs.push(sinc * window);
                }
                coeffs
            }
        }
    }
    
    /// Resample un buffer audio
    pub fn process(&mut self, input: &[f32], output: &mut Vec<f32>) -> Result<(), AppError> {
        // Mise à jour progressive du ratio vers la cible
        let ratio_diff = self.target_ratio - self.ratio;
        self.ratio += ratio_diff * self.config.adaptation_speed;
        
        output.clear();
        
        // Resampling avec interpolation linéaire (simplifié)
        let mut input_pos = 0.0;
        
        while input_pos < input.len() as f64 - 1.0 {
            let index = input_pos as usize;
            let frac = input_pos - index as f64;
            
            // Interpolation linéaire
            let sample = if index + 1 < input.len() {
                input[index] * (1.0 - frac as f32) + input[index + 1] * frac as f32
            } else {
                input[index]
            };
            
            output.push(sample);
            input_pos += self.ratio;
        }
        
        Ok(())
    }
    
    /// Ajuste le ratio de resampling
    pub fn set_target_ratio(&mut self, ratio: f64) {
        let max_dev = self.config.max_ratio_deviation;
        self.target_ratio = ratio.clamp(1.0 - max_dev, 1.0 + max_dev);
    }
}

impl PIDController {
    pub fn new(kp: f32, ki: f32, kd: f32) -> Self {
        Self {
            kp, ki, kd,
            integral: 0.0,
            previous_error: 0.0,
            last_update: Instant::now(),
        }
    }
    
    pub fn update(&mut self, setpoint: f32, measured: f32) -> f32 {
        let now = Instant::now();
        let dt = now.duration_since(self.last_update).as_secs_f32();
        self.last_update = now;
        
        let error = setpoint - measured;
        
        // Terme proportionnel
        let p_term = self.kp * error;
        
        // Terme intégral
        self.integral += error * dt;
        let i_term = self.ki * self.integral;
        
        // Terme dérivé
        let derivative = (error - self.previous_error) / dt;
        let d_term = self.kd * derivative;
        
        self.previous_error = error;
        
        p_term + i_term + d_term
    }
}

impl Default for RealtimeConfig {
    fn default() -> Self {
        Self {
            sample_rate: 44100,
            channels: 2,
            buffer_size: 256, // ~5.8ms @ 44.1kHz
            max_latency_ms: 10.0,
            enable_adaptive_buffering: true,
            enable_jitter_compensation: true,
            thread_priority: ThreadPriority::High,
        }
    }
}

impl Default for ResamplerConfig {
    fn default() -> Self {
        Self {
            max_ratio_deviation: 0.1, // ±10%
            adaptation_speed: 0.01,
            filter_quality: FilterQuality::Medium,
            enable_anti_aliasing: true,
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    
    #[test]
    fn test_ring_buffer() {
        let mut buffer = RingBuffer::new(4);
        
        // Test écriture
        let written = buffer.write(&[1.0, 2.0, 3.0]);
        assert_eq!(written, 3);
        assert_eq!(buffer.available_read(), 3);
        
        // Test lecture
        let mut output = [0.0; 2];
        let read = buffer.read(&mut output);
        assert_eq!(read, 2);
        assert_eq!(output, [1.0, 2.0]);
        assert_eq!(buffer.available_read(), 1);
    }
    
    #[test]
    fn test_realtime_processor() {
        let config = RealtimeConfig::default();
        let mut processor = RealtimeAudioProcessor::new(config).unwrap();
        
        // Test démarrage/arrêt
        assert!(processor.start().is_ok());
        std::thread::sleep(Duration::from_millis(10));
        assert!(processor.stop().is_ok());
    }
    
    #[test]
    fn test_adaptive_resampler() {
        let config = ResamplerConfig::default();
        let mut resampler = AdaptiveResampler::new(config);
        
        let input = vec![1.0, 0.0, -1.0, 0.0];
        let mut output = Vec::new();
        
        assert!(resampler.process(&input, &mut output).is_ok());
        assert!(!output.is_empty());
    }
}
