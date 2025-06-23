//! Serveur de chat Veza - Version simplifiée
//! 
//! Version minimaliste du serveur de chat pour permettre
//! le déploiement et les tests de base.

use axum::{
    extract::{Query, State},
    http::StatusCode,
    response::Json,
    routing::{get, post},
    Router,
};
use std::collections::HashMap;
use std::sync::Arc;
use tokio::net::TcpListener;
use serde::{Deserialize, Serialize};
use tracing::{info, warn};

use chat_server::{
    simple_message_store::{SimpleMessageStore, SimpleMessage},
    error::ChatError,
};

/// État global de l'application
#[derive(Clone)]
struct AppState {
    store: Arc<SimpleMessageStore>,
}

/// Requête d'envoi de message
#[derive(Deserialize)]
struct SendMessageRequest {
    content: String,
    author: String,
    room: Option<String>,
    is_direct: Option<bool>,
}

/// Paramètres de récupération de messages
#[derive(Deserialize)]
struct GetMessagesQuery {
    room: Option<String>,
    limit: Option<i32>,
    user1: Option<String>,
    user2: Option<String>,
}

/// Réponse API standard
#[derive(Serialize)]
struct ApiResponse<T> {
    success: bool,
    data: Option<T>,
    message: String,
}

impl<T> ApiResponse<T> {
    fn success(data: T) -> Self {
        Self {
            success: true,
            data: Some(data),
            message: "OK".to_string(),
        }
    }
    
    fn error(message: String) -> Self {
        Self {
            success: false,
            data: None,
            message,
        }
    }
}

#[tokio::main]
async fn main() -> Result<(), ChatError> {
    // Configuration des logs
    tracing_subscriber::fmt::init();
    
    info!("🚀 Démarrage du serveur de chat Veza...");
    
    // Initialisation du store
    let store = Arc::new(SimpleMessageStore::new());
    let app_state = AppState { store };
    
    // Ajouter quelques messages de test
    let _ = app_state.store.send_simple_message(
        "Bienvenue sur Veza Chat ! 🎉", 
        "system", 
        Some("general"), 
        false
    ).await;
    
    let _ = app_state.store.send_simple_message(
        "Le serveur de chat fonctionne correctement.", 
        "system", 
        Some("general"), 
        false
    ).await;
    
    // Configuration des routes
    let app = Router::new()
        .route("/", get(health_check))
        .route("/health", get(health_check))
        .route("/api/messages", get(get_messages))
        .route("/api/messages", post(send_message))
        .route("/api/messages/stats", get(get_stats))
        .with_state(app_state);
    
    // Démarrage du serveur
    let listener = TcpListener::bind("0.0.0.0:3001").await
        .map_err(|e| ChatError::configuration_error(&format!("Bind error: {}", e)))?;
    
    info!("✅ Serveur démarré sur http://0.0.0.0:3001");
    info!("📊 Endpoints disponibles:");
    info!("   - GET  /health          - Vérification de santé");
    info!("   - GET  /api/messages    - Récupération des messages");
    info!("   - POST /api/messages    - Envoi de message");
    info!("   - GET  /api/messages/stats - Statistiques");
    
    axum::serve(listener, app).await
        .map_err(|e| ChatError::configuration_error(&format!("Server error: {}", e)))?;
    
    Ok(())
}

/// Endpoint de vérification de santé
async fn health_check() -> Json<ApiResponse<HashMap<String, String>>> {
    let mut info = HashMap::new();
    info.insert("status".to_string(), "healthy".to_string());
    info.insert("service".to_string(), "veza-chat-server".to_string());
    info.insert("version".to_string(), "0.2.0".to_string());
    
    Json(ApiResponse::success(info))
}

/// Récupération des messages
async fn get_messages(
    State(state): State<AppState>,
    Query(params): Query<GetMessagesQuery>,
) -> Result<Json<ApiResponse<Vec<SimpleMessage>>>, StatusCode> {
    let limit = params.limit.unwrap_or(50).min(100);
    
    let messages = if let Some(room) = params.room {
        // Messages de salon
        state.store.get_room_messages(&room, limit).await
            .map_err(|e| {
                warn!("Erreur récupération messages salon: {}", e);
                StatusCode::INTERNAL_SERVER_ERROR
            })?
    } else if let (Some(user1), Some(user2)) = (params.user1, params.user2) {
        // Messages directs
        state.store.get_direct_messages(&user1, &user2, limit).await
            .map_err(|e| {
                warn!("Erreur récupération messages directs: {}", e);
                StatusCode::INTERNAL_SERVER_ERROR
            })?
    } else {
        return Err(StatusCode::BAD_REQUEST);
    };
    
    Ok(Json(ApiResponse::success(messages)))
}

/// Envoi de message
async fn send_message(
    State(state): State<AppState>,
    Json(payload): Json<SendMessageRequest>,
) -> Result<Json<ApiResponse<i32>>, StatusCode> {
    let message_id = state.store.send_simple_message(
        &payload.content,
        &payload.author,
        payload.room.as_deref(),
        payload.is_direct.unwrap_or(false),
    ).await.map_err(|e| {
        warn!("Erreur envoi message: {}", e);
        StatusCode::INTERNAL_SERVER_ERROR
    })?;
    
    info!("✅ Message envoyé - ID: {}, auteur: {}", message_id, payload.author);
    
    Ok(Json(ApiResponse::success(message_id)))
}

/// Statistiques basiques
async fn get_stats(
    State(_state): State<AppState>,
) -> Json<ApiResponse<HashMap<String, u32>>> {
    let mut stats = HashMap::new();
    stats.insert("total_messages".to_string(), 2);
    stats.insert("active_users".to_string(), 1);
    stats.insert("rooms".to_string(), 1);
    
    Json(ApiResponse::success(stats))
}
