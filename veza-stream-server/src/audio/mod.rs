/// Module Audio Processing pour Stream Server
/// 
/// Implémente le traitement audio temps réel avec effets optimisés SIMD
/// et gestion de latence ultra-faible pour streaming professionnel

pub mod effects;
pub mod realtime;
pub mod compression;
pub mod processing;

pub use effects::*;
pub use realtime::*;
pub use compression::*;
pub use processing::*;

/// Re-exports pour faciliter l'usage
pub use effects::{
    AudioEffect, 
    EffectsChain, 
    SIMDCompressor, 
    EffectFactory,
    EffectParameter,
    EffectsPerformanceMetrics
};

pub use realtime::{
    RealtimeAudioProcessor,
    RealtimeConfig,
    RealtimeMetrics,
    AdaptiveResampler,
    RingBuffer,
    ThreadPriority
};
