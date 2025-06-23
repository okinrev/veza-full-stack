pub mod error;
pub mod config;
pub mod auth;
pub mod cache;
pub mod health;
pub mod analytics;
pub mod notifications;
pub mod streaming;
pub mod audio;
pub mod utils;
pub mod middleware;

// Re-exports pour faciliter l'utilisation
pub use error::{AppError, Result};
pub use config::Config;

// Types principaux
use std::sync::Arc;
use crate::{
    analytics::AnalyticsEngine,
    audio::{AudioProcessor, CompressionEngine},
    auth::AuthManager,
    cache::FileCache,
    health::HealthMonitor,
    notifications::NotificationService,
    streaming::{AdaptiveStreamingManager, WebSocketManager},
    utils::Metrics,
};

#[derive(Clone)]
pub struct AppState {
    pub config: Arc<Config>,
    pub cache: Arc<FileCache>,
    pub metrics: Arc<Metrics>,
    pub analytics: Arc<AnalyticsEngine>,
    pub audio_processor: Arc<AudioProcessor>,
    pub adaptive_streaming: Arc<AdaptiveStreamingManager>,
    pub health_monitor: Arc<HealthMonitor>,
    pub auth_manager: Arc<AuthManager>,
    pub compression_engine: Arc<CompressionEngine>,
    pub notification_service: Arc<NotificationService>,
    pub websocket_manager: Arc<WebSocketManager>,
} 