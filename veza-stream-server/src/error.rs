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
    Config(String), // Variante pour compatibilité
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
    InvalidSampleRate { rate: u32 },
    InvalidChannelCount { channels: u8 },
    InvalidBitrate { bitrate: u32, codec: String },
    
    // Erreurs de validation et parsing
    ValidationError(String),
    ParseError(String),
    ParameterMismatch { expected: String, got: String },
    InvalidRange,
    
    // Erreurs de synchronisation
    TimeSync,
    NoSyncPoint,
    
    // Erreurs de playback
    InvalidPlaybackState { state: String },
    RateLimitExceeded,
    LimitExceeded { resource: String, limit: u32 },
    ListenerLimitExceeded { current: u32, limit: u32 },
    UploadSessionNotFound { session_id: String },
    InvalidUploadState { current: String, expected: String },
    TooManyActivePlayers { limit: u32 },
    PlayerNotFound { user_id: i64 },
    
    // Erreurs de fichier et stockage
    FileError { message: String },
    FileNotFound,
    NotFound { resource: String },
    StorageError { message: String },
    
    // Erreurs de sérialisation
    SerializationError,
    InvalidData { message: String },
    
    // Erreurs de ressources
    ResourceExhausted { resource: String },
    NotEnoughData,
    BufferOverflow,
    
    // Erreurs de buffer
    BufferNotFound { stream_id: String },
    BufferFull { stream_id: String },
    InsufficientData,
    
    // Erreurs d'autorisation
    Unauthorized,
    Forbidden,
    
    // Erreurs de thread et concurrence
    AlreadyRunning,
    AlreadyProcessing,
    ThreadError { message: String },
    
    // Erreurs génériques
    InternalError { message: String },
    ExternalServiceError { service: String, message: String },
}

impl fmt::Display for AppError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            AppError::ConfigError { message } => write!(f, "Configuration error: {}", message),
            AppError::Config(message) => write!(f, "Configuration error: {}", message),
            AppError::InitializationError { message } => write!(f, "Initialization error: {}", message),
            AppError::NetworkError { message } => write!(f, "Network error: {}", message),
            AppError::ConnectionClosed => write!(f, "Connection closed"),
            AppError::ConnectionTimeout => write!(f, "Connection timeout"),
            AppError::AudioError { message } => write!(f, "Audio error: {}", message),
            AppError::StreamingError { message } => write!(f, "Streaming error: {}", message),
            AppError::EncodingError { message } => write!(f, "Encoding error: {}", message),
            AppError::DecodingError { message } => write!(f, "Decoding error: {}", message),
            AppError::UnsupportedCodec { codec } => write!(f, "Unsupported codec: {}", codec),
            AppError::InvalidSampleRate { rate } => write!(f, "Invalid sample rate: {}", rate),
            AppError::InvalidChannelCount { channels } => write!(f, "Invalid channel count: {}", channels),
            AppError::InvalidBitrate { bitrate, codec } => write!(f, "Invalid bitrate: {} for codec: {}", bitrate, codec),
            AppError::ValidationError(message) => write!(f, "Validation error: {}", message),
            AppError::ParseError(message) => write!(f, "Parse error: {}", message),
            AppError::ParameterMismatch { expected, got } => write!(f, "Parameter mismatch: expected {} but got {}", expected, got),
            AppError::InvalidRange => write!(f, "Invalid range request"),
            AppError::TimeSync => write!(f, "Time synchronization error"),
            AppError::NoSyncPoint => write!(f, "No synchronization point found"),
            AppError::InvalidPlaybackState { state } => write!(f, "Invalid playback state: {}", state),
            AppError::RateLimitExceeded => write!(f, "Rate limit exceeded"),
            AppError::LimitExceeded { resource, limit } => write!(f, "Limit exceeded for {}: max {}", resource, limit),
            AppError::ListenerLimitExceeded { current, limit } => write!(f, "Listener limit exceeded: {} current, {} limit", current, limit),
            AppError::UploadSessionNotFound { session_id } => write!(f, "Upload session not found: {}", session_id),
            AppError::InvalidUploadState { current, expected } => write!(f, "Invalid upload state: current {} but expected {}", current, expected),
            AppError::TooManyActivePlayers { limit } => write!(f, "Too many active players: limit {}", limit),
            AppError::PlayerNotFound { user_id } => write!(f, "Player not found: user_id {}", user_id),
            AppError::FileError { message } => write!(f, "File error: {}", message),
            AppError::FileNotFound => write!(f, "File not found"),
            AppError::NotFound { resource } => write!(f, "Not found: {}", resource),
            AppError::StorageError { message } => write!(f, "Storage error: {}", message),
            AppError::SerializationError => write!(f, "Serialization error"),
            AppError::InvalidData { message } => write!(f, "Invalid data: {}", message),
            AppError::ResourceExhausted { resource } => write!(f, "Resource exhausted: {}", resource),
            AppError::NotEnoughData => write!(f, "Not enough data"),
            AppError::BufferOverflow => write!(f, "Buffer overflow"),
            AppError::BufferNotFound { stream_id } => write!(f, "Buffer not found for stream: {}", stream_id),
            AppError::BufferFull { stream_id } => write!(f, "Buffer full for stream: {}", stream_id),
            AppError::InsufficientData => write!(f, "Insufficient data"),
            AppError::Unauthorized => write!(f, "Unauthorized access"),
            AppError::Forbidden => write!(f, "Forbidden access"),
            AppError::AlreadyRunning => write!(f, "Process already running"),
            AppError::AlreadyProcessing => write!(f, "Already processing"),
            AppError::ThreadError { message } => write!(f, "Thread error: {}", message),
            AppError::InternalError { message } => write!(f, "Internal error: {}", message),
            AppError::ExternalServiceError { service, message } => write!(f, "External service error: {} - {}", service, message),
        }
    }
}

impl std::error::Error for AppError {}

impl IntoResponse for AppError {
    fn into_response(self) -> Response {
        let (status, error_message) = match self {
            AppError::ConfigError { message } => (StatusCode::INTERNAL_SERVER_ERROR, message),
            AppError::Config(message) => (StatusCode::INTERNAL_SERVER_ERROR, message),
            AppError::InitializationError { message } => (StatusCode::INTERNAL_SERVER_ERROR, message),
            AppError::NetworkError { message } => (StatusCode::INTERNAL_SERVER_ERROR, message),
            AppError::ConnectionClosed => (StatusCode::INTERNAL_SERVER_ERROR, "Connection closed".to_string()),
            AppError::ConnectionTimeout => (StatusCode::INTERNAL_SERVER_ERROR, "Connection timeout".to_string()),
            AppError::AudioError { message } => (StatusCode::INTERNAL_SERVER_ERROR, message),
            AppError::StreamingError { message } => (StatusCode::INTERNAL_SERVER_ERROR, message),
            AppError::EncodingError { message } => (StatusCode::INTERNAL_SERVER_ERROR, message),
            AppError::DecodingError { message } => (StatusCode::INTERNAL_SERVER_ERROR, message),
            AppError::UnsupportedCodec { codec } => (StatusCode::BAD_REQUEST, format!("Unsupported codec: {}", codec)),
            AppError::InvalidSampleRate { rate } => (StatusCode::BAD_REQUEST, format!("Invalid sample rate: {}", rate)),
            AppError::InvalidChannelCount { channels } => (StatusCode::BAD_REQUEST, format!("Invalid channel count: {}", channels)),
            AppError::InvalidBitrate { bitrate, codec } => (StatusCode::BAD_REQUEST, format!("Invalid bitrate: {} for codec: {}", bitrate, codec)),
            AppError::ValidationError(message) => (StatusCode::BAD_REQUEST, message),
            AppError::ParseError(message) => (StatusCode::BAD_REQUEST, message),
            AppError::ParameterMismatch { expected, got } => (StatusCode::BAD_REQUEST, format!("Parameter mismatch: expected {} but got {}", expected, got)),
            AppError::InvalidRange => (StatusCode::RANGE_NOT_SATISFIABLE, "Invalid range request".to_string()),
            AppError::TimeSync => (StatusCode::INTERNAL_SERVER_ERROR, "Time synchronization error".to_string()),
            AppError::NoSyncPoint => (StatusCode::INTERNAL_SERVER_ERROR, "No synchronization point found".to_string()),
            AppError::InvalidPlaybackState { state } => (StatusCode::BAD_REQUEST, format!("Invalid playback state: {}", state)),
            AppError::RateLimitExceeded => (StatusCode::TOO_MANY_REQUESTS, "Rate limit exceeded".to_string()),
            AppError::LimitExceeded { resource, limit } => (StatusCode::TOO_MANY_REQUESTS, format!("Limit exceeded for {}: max {}", resource, limit)),
            AppError::ListenerLimitExceeded { current, limit } => (StatusCode::TOO_MANY_REQUESTS, format!("Listener limit exceeded: {} current, {} limit", current, limit)),
            AppError::UploadSessionNotFound { session_id } => (StatusCode::NOT_FOUND, format!("Upload session not found: {}", session_id)),
            AppError::InvalidUploadState { current, expected } => (StatusCode::BAD_REQUEST, format!("Invalid upload state: current {} but expected {}", current, expected)),
            AppError::TooManyActivePlayers { limit } => (StatusCode::BAD_REQUEST, format!("Too many active players: limit {}", limit)),
            AppError::PlayerNotFound { user_id } => (StatusCode::NOT_FOUND, format!("Player not found: user_id {}", user_id)),
            AppError::FileError { message } => (StatusCode::INTERNAL_SERVER_ERROR, message),
            AppError::FileNotFound => (StatusCode::NOT_FOUND, "File not found".to_string()),
            AppError::NotFound { resource } => (StatusCode::NOT_FOUND, format!("Not found: {}", resource)),
            AppError::StorageError { message } => (StatusCode::INTERNAL_SERVER_ERROR, message),
            AppError::SerializationError => (StatusCode::INTERNAL_SERVER_ERROR, "Serialization error".to_string()),
            AppError::InvalidData { message } => (StatusCode::BAD_REQUEST, message),
            AppError::ResourceExhausted { resource } => (StatusCode::INTERNAL_SERVER_ERROR, format!("Resource exhausted: {}", resource)),
            AppError::NotEnoughData => (StatusCode::BAD_REQUEST, "Not enough data".to_string()),
            AppError::BufferOverflow => (StatusCode::BAD_REQUEST, "Buffer overflow".to_string()),
            AppError::BufferNotFound { stream_id } => (StatusCode::NOT_FOUND, format!("Buffer not found for stream: {}", stream_id)),
            AppError::BufferFull { stream_id } => (StatusCode::TOO_MANY_REQUESTS, format!("Buffer full for stream: {}", stream_id)),
            AppError::InsufficientData => (StatusCode::BAD_REQUEST, "Insufficient data".to_string()),
            AppError::Unauthorized => (StatusCode::UNAUTHORIZED, "Unauthorized access".to_string()),
            AppError::Forbidden => (StatusCode::FORBIDDEN, "Forbidden access".to_string()),
            AppError::AlreadyRunning => (StatusCode::CONFLICT, "Process already running".to_string()),
            AppError::AlreadyProcessing => (StatusCode::CONFLICT, "Already processing".to_string()),
            AppError::ThreadError { message } => (StatusCode::INTERNAL_SERVER_ERROR, message),
            AppError::InternalError { message } => (StatusCode::INTERNAL_SERVER_ERROR, message),
            AppError::ExternalServiceError { service, message } => (StatusCode::INTERNAL_SERVER_ERROR, format!("External service error: {} - {}", service, message)),
        };

        let body = Json(json!({
            "error": error_message,
            "status": status.as_u16(),
        }));

        (status, body).into_response()
    }
}

pub type Result<T> = std::result::Result<T, AppError>; 
// Conversions depuis les erreurs standard
impl From<std::io::Error> for AppError {
    fn from(err: std::io::Error) -> Self {
        AppError::FileError { message: err.to_string() }
    }
}
