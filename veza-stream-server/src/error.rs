/// Gestion centralisée des erreurs du stream server
/// 
/// Hiérarchie d'erreurs pour un debugging efficace et une gestion robuste

use std::fmt;
use axum::http::StatusCode;
use axum::response::{IntoResponse, Response};
use axum::Json;
use serde_json::json;
use serde::{Serialize, Deserialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum AppError {
    // Configuration et initialisation
    ConfigError { message: String },
    InitializationError { message: String },
    
    // Erreurs réseau et connexion
    NetworkError { message: String },
    ConnectionClosed,
    ConnectionTimeout,
    
    // Erreurs audio et streaming
    AudioError { message: String },
    StreamingError { message: String },
    EncodingError { message: String },
    DecodingError { message: String },
    
    // Erreurs de codec
    UnsupportedCodec { codec: String },
    InvalidSampleRate { rate: u32, supported: Vec<u32> },
    InvalidChannelCount { channels: u8 },
    
    // Erreurs de fichier et stockage
    FileError { message: String },
    NotFound { resource: String },
    StorageError { message: String },
    
    // Erreurs de sérialisation
    SerializationError,
    InvalidData { message: String },
    
    // Erreurs de ressources
    ResourceExhausted { resource: String },
    NotEnoughData,
    BufferOverflow,
    
    // Erreurs d'autorisation
    Unauthorized,
    Forbidden,
    
    // Erreurs de thread et concurrence
    AlreadyRunning,
    ThreadError { message: String },
    
    // Erreurs génériques
    InternalError { message: String },
    ExternalServiceError { service: String, message: String },
}

impl fmt::Display for AppError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            AppError::ConfigError { message } => write!(f, "Configuration error: {}", message),
            AppError::InitializationError { message } => write!(f, "Initialization error: {}", message),
            AppError::NetworkError { message } => write!(f, "Network error: {}", message),
            AppError::ConnectionClosed => write!(f, "Connection closed"),
            AppError::ConnectionTimeout => write!(f, "Connection timeout"),
            AppError::AudioError { message } => write!(f, "Audio error: {}", message),
            AppError::StreamingError { message } => write!(f, "Streaming error: {}", message),
            AppError::EncodingError { message } => write!(f, "Encoding error: {}", message),
            AppError::DecodingError { message } => write!(f, "Decoding error: {}", message),
            AppError::UnsupportedCodec { codec } => write!(f, "Unsupported codec: {}", codec),
            AppError::InvalidSampleRate { rate, supported } => write!(f, "Invalid sample rate: {} (supported: {:?})", rate, supported),
            AppError::InvalidChannelCount { channels } => write!(f, "Invalid channel count: {}", channels),
            AppError::FileError { message } => write!(f, "File error: {}", message),
            AppError::NotFound { resource } => write!(f, "Not found: {}", resource),
            AppError::StorageError { message } => write!(f, "Storage error: {}", message),
            AppError::SerializationError => write!(f, "Serialization error"),
            AppError::InvalidData { message } => write!(f, "Invalid data: {}", message),
            AppError::ResourceExhausted { resource } => write!(f, "Resource exhausted: {}", resource),
            AppError::NotEnoughData => write!(f, "Not enough data"),
            AppError::BufferOverflow => write!(f, "Buffer overflow"),
            AppError::Unauthorized => write!(f, "Unauthorized access"),
            AppError::Forbidden => write!(f, "Forbidden access"),
            AppError::AlreadyRunning => write!(f, "Process already running"),
            AppError::ThreadError { message } => write!(f, "Thread error: {}", message),
            AppError::InternalError { message } => write!(f, "Internal error: {}", message),
            AppError::ExternalServiceError { service, message } => write!(f, "External service error: {} - {}", service, message),
        }
    }
}

impl std::error::Error for AppError {}

impl IntoResponse for AppError {
    fn into_response(self) -> Response {
        let (status, error_message) = match &self {
            AppError::ConfigError { message } => (StatusCode::INTERNAL_SERVER_ERROR, message),
            AppError::InitializationError { message } => (StatusCode::INTERNAL_SERVER_ERROR, message),
            AppError::NetworkError { message } => (StatusCode::INTERNAL_SERVER_ERROR, message),
            AppError::ConnectionClosed => (StatusCode::INTERNAL_SERVER_ERROR, "Connection closed"),
            AppError::ConnectionTimeout => (StatusCode::INTERNAL_SERVER_ERROR, "Connection timeout"),
            AppError::AudioError { message } => (StatusCode::INTERNAL_SERVER_ERROR, message),
            AppError::StreamingError { message } => (StatusCode::INTERNAL_SERVER_ERROR, message),
            AppError::EncodingError { message } => (StatusCode::INTERNAL_SERVER_ERROR, message),
            AppError::DecodingError { message } => (StatusCode::INTERNAL_SERVER_ERROR, message),
                         AppError::UnsupportedCodec { codec } => (StatusCode::BAD_REQUEST, &format!("Unsupported codec: {}", codec)),
             AppError::InvalidSampleRate { rate, supported } => (StatusCode::BAD_REQUEST, &format!("Invalid sample rate: {} (supported: {:?})", rate, supported)),
             AppError::InvalidChannelCount { channels } => (StatusCode::BAD_REQUEST, &format!("Invalid channel count: {}", channels)),
            AppError::FileError { message } => (StatusCode::INTERNAL_SERVER_ERROR, message),
                         AppError::NotFound { resource } => (StatusCode::NOT_FOUND, &format!("Not found: {}", resource)),
             AppError::StorageError { message } => (StatusCode::INTERNAL_SERVER_ERROR, message),
             AppError::SerializationError => (StatusCode::INTERNAL_SERVER_ERROR, "Serialization error"),
             AppError::InvalidData { message } => (StatusCode::BAD_REQUEST, message),
             AppError::ResourceExhausted { resource } => (StatusCode::INTERNAL_SERVER_ERROR, &format!("Resource exhausted: {}", resource)),
             AppError::NotEnoughData => (StatusCode::BAD_REQUEST, "Not enough data"),
             AppError::BufferOverflow => (StatusCode::BAD_REQUEST, "Buffer overflow"),
             AppError::Unauthorized => (StatusCode::UNAUTHORIZED, "Unauthorized access"),
             AppError::Forbidden => (StatusCode::FORBIDDEN, "Forbidden access"),
             AppError::AlreadyRunning => (StatusCode::CONFLICT, "Process already running"),
             AppError::ThreadError { message } => (StatusCode::INTERNAL_SERVER_ERROR, message),
             AppError::InternalError { message } => (StatusCode::INTERNAL_SERVER_ERROR, message),
             AppError::ExternalServiceError { service, message } => (StatusCode::INTERNAL_SERVER_ERROR, &format!("External service error: {} - {}", service, message)),
        };

        let body = Json(json!({
            "error": error_message,
            "status": status.as_u16(),
        }));

        (status, body).into_response()
    }
}

pub type Result<T> = std::result::Result<T, AppError>; 