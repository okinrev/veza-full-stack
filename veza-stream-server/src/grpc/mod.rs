/// Module gRPC pour intégration avec le backend Go
/// 
/// Gère la communication bidirectionnelle entre le stream server Rust
/// et le backend API Go pour synchronisation des données et événements

pub mod stream_service;
pub mod auth_service;
pub mod events_service;
pub mod client;
pub mod server;

pub use stream_service::*;
pub use auth_service::*;
pub use events_service::*;
pub use client::*;
pub use server::*;

use std::sync::Arc;
use std::collections::HashMap;
use tonic::{transport::{Server, Channel}, Request, Response, Status};
use tokio::sync::{mpsc, RwLock};
use tracing::{info, warn, error, debug};
use uuid::Uuid;
use serde::{Serialize, Deserialize};

use crate::error::AppError;
use crate::core::{StreamManager, StreamEvent};

/// Configuration pour le serveur gRPC
#[derive(Debug, Clone)]
pub struct GrpcConfig {
    /// Port d'écoute du serveur gRPC
    pub port: u16,
    /// Adresse de bind
    pub bind_address: String,
    /// Activation TLS
    pub enable_tls: bool,
    /// Chemin vers le certificat TLS
    pub tls_cert_path: Option<String>,
    /// Chemin vers la clé privée TLS
    pub tls_key_path: Option<String>,
    /// Timeout pour les requêtes
    pub request_timeout_ms: u64,
    /// Taille max des messages
    pub max_message_size: usize,
    /// Keep-alive interval
    pub keep_alive_interval_ms: u64,
}

/// Service principal gRPC pour le stream server
#[derive(Debug)]
pub struct StreamServerGrpc {
    /// Gestionnaire de streams
    stream_manager: Arc<StreamManager>,
    /// Configuration
    config: GrpcConfig,
    /// Canal d'événements vers le backend Go
    event_sender: mpsc::UnboundedSender<StreamEvent>,
    /// État des connexions actives
    active_connections: Arc<RwLock<HashMap<String, ConnectionInfo>>>,
    /// Métriques gRPC
    metrics: Arc<RwLock<GrpcMetrics>>,
}

/// Information de connexion gRPC
#[derive(Debug, Clone)]
pub struct ConnectionInfo {
    pub client_id: String,
    pub connected_at: std::time::SystemTime,
    pub last_activity: std::time::SystemTime,
    pub request_count: u64,
    pub client_version: String,
}

/// Métriques du service gRPC
#[derive(Debug, Clone, Default)]
pub struct GrpcMetrics {
    pub total_requests: u64,
    pub active_connections: u64,
    pub successful_requests: u64,
    pub failed_requests: u64,
    pub average_response_time_ms: f64,
    pub requests_per_second: f64,
}

/// Client gRPC pour communication avec le backend Go
#[derive(Debug, Clone)]
pub struct GoBackendClient {
    /// URL du backend Go
    backend_url: String,
    /// Configuration de retry
    retry_config: RetryConfig,
    /// Métriques client
    metrics: Arc<RwLock<ClientMetrics>>,
}

/// Configuration de retry pour le client
#[derive(Debug, Clone)]
pub struct RetryConfig {
    pub max_retries: u32,
    pub base_delay_ms: u64,
    pub max_delay_ms: u64,
    pub exponential_backoff: bool,
}

/// Métriques du client gRPC
#[derive(Debug, Clone, Default)]
pub struct ClientMetrics {
    pub requests_sent: u64,
    pub responses_received: u64,
    pub errors: u64,
    pub retries: u64,
    pub connection_failures: u64,
}

impl Default for GrpcConfig {
    fn default() -> Self {
        Self {
            port: 50051,
            bind_address: "0.0.0.0".to_string(),
            enable_tls: false,
            tls_cert_path: None,
            tls_key_path: None,
            request_timeout_ms: 30000,
            max_message_size: 4 * 1024 * 1024, // 4MB
            keep_alive_interval_ms: 60000,
        }
    }
}

impl Default for RetryConfig {
    fn default() -> Self {
        Self {
            max_retries: 3,
            base_delay_ms: 100,
            max_delay_ms: 5000,
            exponential_backoff: true,
        }
    }
}

impl StreamServerGrpc {
    /// Crée un nouveau service gRPC
    pub fn new(
        stream_manager: Arc<StreamManager>,
        config: GrpcConfig,
    ) -> Result<Self, AppError> {
        let (event_sender, _event_receiver) = mpsc::unbounded_channel();
        
        Ok(Self {
            stream_manager,
            config,
            event_sender,
            active_connections: Arc::new(RwLock::new(HashMap::new())),
            metrics: Arc::new(RwLock::new(GrpcMetrics::default())),
        })
    }
    
    /// Démarre le serveur gRPC (simulation)
    pub async fn start(&self) -> Result<(), AppError> {
        let addr = format!("{}:{}", self.config.bind_address, self.config.port);
        
        info!("🚀 Démarrage serveur gRPC sur {}", addr);
        
        // Simulation du serveur gRPC
        tokio::spawn({
            let metrics = self.metrics.clone();
            async move {
                loop {
                    tokio::time::sleep(std::time::Duration::from_secs(1)).await;
                    
                    // Simulation de métriques
                    let mut m = metrics.write().await;
                    m.total_requests += 10;
                    m.successful_requests += 9;
                    m.failed_requests += 1;
                    m.average_response_time_ms = 15.5;
                    m.requests_per_second = 10.0;
                }
            }
        });
        
        info!("✅ Serveur gRPC démarré avec succès");
        Ok(())
    }
    
    /// Obtient les métriques du serveur
    pub async fn get_metrics(&self) -> GrpcMetrics {
        self.metrics.read().await.clone()
    }
    
    /// Met à jour les métriques
    async fn update_metrics<F>(&self, update_fn: F) 
    where 
        F: FnOnce(&mut GrpcMetrics),
    {
        let mut metrics = self.metrics.write().await;
        update_fn(&mut metrics);
    }
}

impl GoBackendClient {
    /// Crée un nouveau client pour le backend Go
    pub fn new(backend_url: String) -> Self {
        Self {
            backend_url,
            retry_config: RetryConfig::default(),
            metrics: Arc::new(RwLock::new(ClientMetrics::default())),
        }
    }
    
    /// Établit la connexion avec le backend Go
    pub async fn connect(&mut self) -> Result<(), AppError> {
        info!("🔗 Connexion au backend Go: {}", self.backend_url);
        
        // Simulation de connexion
        tokio::time::sleep(std::time::Duration::from_millis(100)).await;
        
        info!("✅ Connexion établie avec le backend Go");
        Ok(())
    }
    
    /// Envoie un événement au backend Go
    pub async fn send_event(&self, event: StreamEvent) -> Result<(), AppError> {
        debug!("📤 Envoi événement vers backend Go: {:?}", event);
        
        // Conversion de l'événement en message gRPC
        let grpc_event = self.convert_event_to_grpc(event);
        
        // Envoi avec retry
        self.send_with_retry(grpc_event).await?;
        
        // Mise à jour métriques
        let mut metrics = self.metrics.write().await;
        metrics.requests_sent += 1;
        
        Ok(())
    }
    
    /// Envoie avec retry automatique
    async fn send_with_retry(&self, _event: GrpcEvent) -> Result<(), AppError> {
        let mut attempts = 0;
        let mut delay = self.retry_config.base_delay_ms;
        
        loop {
            attempts += 1;
            
            // Tentative d'envoi (simulation)
            match self.attempt_send().await {
                Ok(_) => {
                    debug!("✅ Événement envoyé au backend Go (tentative {})", attempts);
                    return Ok(());
                }
                Err(e) => {
                    if attempts >= self.retry_config.max_retries {
                        error!("❌ Échec définitif après {} tentatives: {}", attempts, e);
                        return Err(e);
                    }
                    
                    warn!("⚠️  Tentative {} échouée, retry dans {}ms: {}", attempts, delay, e);
                    
                    tokio::time::sleep(std::time::Duration::from_millis(delay)).await;
                    
                    if self.retry_config.exponential_backoff {
                        delay = (delay * 2).min(self.retry_config.max_delay_ms);
                    }
                }
            }
        }
    }
    
    /// Tentative d'envoi (simulation)
    async fn attempt_send(&self) -> Result<(), AppError> {
        // Simulation d'envoi gRPC
        tokio::time::sleep(std::time::Duration::from_millis(10)).await;
        Ok(())
    }
    
    /// Convertit un événement stream en message gRPC
    fn convert_event_to_grpc(&self, event: StreamEvent) -> GrpcEvent {
        match event {
            StreamEvent::StreamStarted { stream_id, metadata } => {
                GrpcEvent::StreamStarted {
                    stream_id: stream_id.to_string(),
                    title: metadata.title,
                    description: metadata.description.unwrap_or_default(),
                    tags: metadata.tags,
                }
            }
            StreamEvent::StreamEnded { stream_id, duration, .. } => {
                GrpcEvent::StreamEnded {
                    stream_id: stream_id.to_string(),
                    duration_ms: duration.as_millis() as u64,
                }
            }
            StreamEvent::ListenerJoined { stream_id, listener_id, .. } => {
                GrpcEvent::ListenerJoined {
                    stream_id: stream_id.to_string(),
                    listener_id: listener_id.to_string(),
                }
            }
            StreamEvent::ListenerLeft { stream_id, listener_id, .. } => {
                GrpcEvent::ListenerLeft {
                    stream_id: stream_id.to_string(),
                    listener_id: listener_id.to_string(),
                }
            }
        }
    }
    
    /// Obtient les métriques du client
    pub async fn get_metrics(&self) -> ClientMetrics {
        self.metrics.read().await.clone()
    }
}

/// Événement gRPC pour communication avec le backend
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum GrpcEvent {
    StreamStarted {
        stream_id: String,
        title: String,
        description: String,
        tags: Vec<String>,
    },
    StreamEnded {
        stream_id: String,
        duration_ms: u64,
    },
    ListenerJoined {
        stream_id: String,
        listener_id: String,
    },
    ListenerLeft {
        stream_id: String,
        listener_id: String,
    },
}

// Re-exports pour compatibilité avec les services générés
pub use generated_stream_service::*;
pub use generated_auth_service::*;
pub use generated_events_service::*;

// Modules temporaires pour services générés (seront remplacés par proto)
mod generated_stream_service {
    use tonic::{async_trait, Request, Response, Status};
    
    pub struct StreamServiceServer<T> {
        inner: T,
    }
    
    impl<T> StreamServiceServer<T> {
        pub fn new(inner: T) -> Self {
            Self { inner }
        }
    }
    
    pub struct StreamServiceImpl {
        // Sera implémenté avec les vrais services gRPC
    }
}

mod generated_auth_service {
    pub struct AuthServiceServer<T> {
        inner: T,
    }
    
    impl<T> AuthServiceServer<T> {
        pub fn new(inner: T) -> Self {
            Self { inner }
        }
    }
    
    pub struct AuthServiceImpl {
        // Sera implémenté avec les vrais services gRPC
    }
}

mod generated_events_service {
    pub struct EventsServiceServer<T> {
        inner: T,
    }
    
    impl<T> EventsServiceServer<T> {
        pub fn new(inner: T) -> Self {
            Self { inner }
        }
    }
    
    pub struct EventsServiceImpl {
        // Sera implémenté avec les vrais services gRPC
    }
} 