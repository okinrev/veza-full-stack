use std::fmt;
use axum::http::StatusCode;
use axum::response::{IntoResponse, Response};
use axum::Json;
use serde_json::json;

#[derive(Debug)]
pub enum AppError {
    // Erreurs de configuration
    Config(String),
    
    // Erreurs d'authentification
    Unauthorized,
    InvalidToken,
    TokenExpired,
    
    // Erreurs de streaming
    FileNotFound,
    InvalidRange,
    StreamingError(String),
    
    // Erreurs de base de données
    Database(String),
    
    // Erreurs de cache
    CacheError(String),
    
    // Erreurs de validation
    ValidationError(String),
    
    // Erreurs de rate limiting
    RateLimited,
    
    // Erreurs d'I/O
    IoError(String),
    
    // Erreurs de parsing
    ParseError(String),
    
    // Erreurs internes
    Internal(String),
}

impl fmt::Display for AppError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            AppError::Config(msg) => write!(f, "Configuration error: {}", msg),
            AppError::Unauthorized => write!(f, "Unauthorized access"),
            AppError::InvalidToken => write!(f, "Invalid token"),
            AppError::TokenExpired => write!(f, "Token expired"),
            AppError::FileNotFound => write!(f, "File not found"),
            AppError::InvalidRange => write!(f, "Invalid range request"),
            AppError::StreamingError(msg) => write!(f, "Streaming error: {}", msg),
            AppError::Database(msg) => write!(f, "Database error: {}", msg),
            AppError::CacheError(msg) => write!(f, "Cache error: {}", msg),
            AppError::ValidationError(msg) => write!(f, "Validation error: {}", msg),
            AppError::RateLimited => write!(f, "Rate limit exceeded"),
            AppError::IoError(msg) => write!(f, "I/O error: {}", msg),
            AppError::ParseError(msg) => write!(f, "Parsing error: {}", msg),
            AppError::Internal(msg) => write!(f, "Internal error: {}", msg),
        }
    }
}

impl std::error::Error for AppError {}

impl IntoResponse for AppError {
    fn into_response(self) -> Response {
        let (status, error_message) = match &self {
            AppError::Config(_) => (StatusCode::INTERNAL_SERVER_ERROR, "Configuration error"),
            AppError::Unauthorized => (StatusCode::UNAUTHORIZED, "Unauthorized"),
            AppError::InvalidToken => (StatusCode::UNAUTHORIZED, "Invalid token"),
            AppError::TokenExpired => (StatusCode::UNAUTHORIZED, "Token expired"),
            AppError::FileNotFound => (StatusCode::NOT_FOUND, "File not found"),
            AppError::InvalidRange => (StatusCode::RANGE_NOT_SATISFIABLE, "Invalid range"),
            AppError::StreamingError(_) => (StatusCode::INTERNAL_SERVER_ERROR, "Streaming error"),
            AppError::Database(_) => (StatusCode::INTERNAL_SERVER_ERROR, "Database error"),
            AppError::CacheError(_) => (StatusCode::INTERNAL_SERVER_ERROR, "Cache error"),
            AppError::ValidationError(_) => (StatusCode::BAD_REQUEST, "Validation error"),
            AppError::RateLimited => (StatusCode::TOO_MANY_REQUESTS, "Rate limit exceeded"),
            AppError::IoError(_) => (StatusCode::INTERNAL_SERVER_ERROR, "I/O error"),
            AppError::ParseError(_) => (StatusCode::BAD_REQUEST, "Parsing error"),
            AppError::Internal(_) => (StatusCode::INTERNAL_SERVER_ERROR, "Internal error"),
        };

        let body = Json(json!({
            "error": error_message,
            "status": status.as_u16(),
        }));

        (status, body).into_response()
    }
}

pub type Result<T> = std::result::Result<T, AppError>; 