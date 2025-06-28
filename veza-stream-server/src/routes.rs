use axum::{
    extract::{Query, State},
    http::{HeaderMap, StatusCode},
    response::{IntoResponse, Response},
    routing::{get, post},
    Json, Router,
};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;

use crate::{AppState, AppError, Result};

#[derive(Deserialize)]
pub struct StreamRequest {
    pub file: String,
    pub signature: String,
    pub timestamp: i64,
}

#[derive(Serialize)]
pub struct StreamResponse {
    pub success: bool,
    pub message: String,
}

pub fn create_routes() -> Router<AppState> {
    Router::new()
        .route("/health", get(health_check))
        .route("/stream", get(stream_file_handler))
        .route("/metadata", get(file_metadata_handler))
        .route("/websocket", get(websocket_handler))
}

async fn health_check() -> impl IntoResponse {
    Json(serde_json::json!({
        "status": "ok",
        "service": "stream-server"
    }))
}

async fn stream_file_handler(
    Query(params): Query<StreamRequest>,
    headers: HeaderMap,
    State(_state): State<AppState>,
) -> Result<Response> {
    // Validation basique
    if params.file.is_empty() {
        return Err(AppError::ValidationError("File parameter is required".to_string()));
    }

    // Pour les tests, on retourne une réponse simple
    let response = StreamResponse {
        success: true,
        message: format!("Stream request for file: {}", params.file),
    };

    Ok(Json(response).into_response())
}

async fn file_metadata_handler(
    Query(params): Query<HashMap<String, String>>,
    State(_state): State<AppState>,
) -> Result<Json<serde_json::Value>> {
    let file_path = params.get("file")
        .ok_or_else(|| AppError::ValidationError("File parameter is required".to_string()))?;

    // Métadonnées simulées pour les tests
    let metadata = serde_json::json!({
        "file": file_path,
        "size": 1024000,
        "format": "mp3",
        "duration": 180
    });

    Ok(Json(metadata))
}

async fn websocket_handler() -> Result<Response> {
    // Handler WebSocket basique pour les tests
    Ok((StatusCode::SWITCHING_PROTOCOLS, "WebSocket upgrade").into_response())
} 