/// Service gRPC pour la gestion des streams

use std::sync::Arc;
use tokio::sync::mpsc;
use tracing::{info};

use crate::error::AppError;
use crate::core::{StreamManager, StreamEvent};

/// Implémentation du service Stream gRPC
#[derive(Debug)]
pub struct StreamServiceImpl {
    stream_manager: Arc<StreamManager>,
    event_sender: mpsc::UnboundedSender<StreamEvent>,
}

impl StreamServiceImpl {
    pub fn new(
        stream_manager: Arc<StreamManager>,
        event_sender: mpsc::UnboundedSender<StreamEvent>,
    ) -> Self {
        Self {
            stream_manager,
            event_sender,
        }
    }

    pub async fn test_service(&self) -> Result<String, AppError> {
        info!("🧪 Test du service Stream gRPC");
        Ok("Stream service OK".to_string())
    }
}
