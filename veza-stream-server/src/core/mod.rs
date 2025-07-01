/// Core streaming modules pour production
/// 
/// Cette architecture supporte :
/// - 10k+ streams simultanés  
/// - 100k+ listeners par stream
/// - Adaptive bitrate seamless
/// - Multi-codec (Opus, AAC, MP3, FLAC)
/// - Synchronisation précise multi-client

pub mod stream;
pub mod encoder;
pub mod buffer;
pub mod sync;

// Re-exports pour faciliter l'usage
pub use stream::*;
pub use encoder::*;
pub use buffer::*;
pub use sync::*; 