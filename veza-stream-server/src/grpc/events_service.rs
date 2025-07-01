/// Service gRPC pour les événements

use std::sync::Arc;
use tokio::sync::mpsc;
use tracing::{info, debug, error};
use serde::{Serialize, Deserialize};

use crate::error::AppError;
use crate::core::StreamEvent;

/// Service d'événements gRPC
#[derive(Debug)]
pub struct EventsServiceImpl {
    event_sender: mpsc::UnboundedSender<StreamEvent>,
}

impl EventsServiceImpl {
    pub fn new(event_sender: mpsc::UnboundedSender<StreamEvent>) -> Self {
        Self { event_sender }
    }

    pub async fn publish_event(&self, event: StreamEvent) -> Result<(), AppError> {
        info!("📡 Publication événement: {:?}", event);
        
        self.event_sender.send(event)
            .map_err(|e| AppError::NetworkError(format!("Failed to send event: {}", e)))?;
        
        Ok(())
    }
}
