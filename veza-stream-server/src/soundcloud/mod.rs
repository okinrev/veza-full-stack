/// Features SoundCloud-like pour streaming production
/// 
/// Modules implémentés :
/// - Upload & Management multi-format
/// - Playback Experience avancée
/// - Social Features complètes 
/// - Discovery & Algorithmes ML
/// - Creator Tools & Analytics

pub mod upload;
pub mod management;
pub mod playback;
pub mod social;
pub mod discovery;
pub mod creator;
pub mod waveform;

// Re-exports pour faciliter l'usage
pub use upload::*;
pub use management::*;
pub use playback::*;
// pub use social::*;
pub use discovery::*;
pub use creator::*;
pub use waveform::*; 