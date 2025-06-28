//! Serveur de chat Veza - Version avec WebSocket
//! 
//! Version complète du serveur de chat avec support WebSocket et HTTP REST.

use axum::{
    extract::{Query, State, WebSocketUpgrade, ws::{Message, WebSocket}},
    http::StatusCode,
    response::{Response, Json},
    routing::{get, post},
    Router,
};
use std::collections::HashMap;
use std::sync::Arc;
use tokio::net::TcpListener;
use serde::{Deserialize, Serialize};
use tracing::{info, warn, error};
use futures_util::{SinkExt, StreamExt};
use tokio::sync::RwLock;

use chat_server::{
    simple_message_store::{SimpleMessageStore, SimpleMessage},
    websocket::{WebSocketManager, IncomingMessage, OutgoingMessage},
    error::ChatError,
    models::User,
};

/// État global de l'application
#[derive(Clone)]
struct AppState {
    store: Arc<SimpleMessageStore>,
    ws_manager: Arc<WebSocketManager>,
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
    data: T,
    message: Option<String>,
}

impl<T> ApiResponse<T> {
    fn success(data: T) -> Self {
        Self {
            success: true,
            data,
            message: None,
        }
    }
    
    fn _error(message: String) -> Self 
    where 
        T: Default 
    {
        Self {
            success: false,
            data: T::default(),
            message: Some(message),
        }
    }
}

#[tokio::main]
async fn main() -> Result<(), ChatError> {
    // Configuration du logging
    tracing_subscriber::fmt()
        .with_max_level(tracing::Level::INFO)
        .init();

    info!("🚀 Démarrage du serveur de chat Veza...");

    // Initialisation du store de messages
    let store = Arc::new(SimpleMessageStore::new());
    
    // Initialisation du gestionnaire WebSocket
    let ws_manager = Arc::new(WebSocketManager::new());
    
    let state = AppState {
        store,
        ws_manager,
    };

    // Configuration des routes avec WebSocket
    let app = Router::new()
        .route("/health", get(health_check))
        .route("/api/messages", get(get_messages))
        .route("/api/messages", post(send_message))
        .route("/api/messages/stats", get(get_stats))
        .route("/ws", get(websocket_handler))  // ✨ NOUVEAU: Endpoint WebSocket
        .with_state(state);

    // Démarrage du serveur
    let listener = TcpListener::bind("0.0.0.0:3001").await
        .map_err(|e| ChatError::configuration_error(&format!("Bind error: {}", e)))?;

    info!("✅ Serveur démarré sur http://0.0.0.0:3001");
    info!("📊 Endpoints disponibles:");
    info!("   - GET  /health          - Vérification de santé");
    info!("   - GET  /api/messages    - Récupération des messages");
    info!("   - POST /api/messages    - Envoi de message");
    info!("   - GET  /api/messages/stats - Statistiques");
    info!("   - GET  /ws              - WebSocket Chat (🆕)");
    
    axum::serve(listener, app).await
        .map_err(|e| ChatError::configuration_error(&format!("Server error: {}", e)))?;
    
    Ok(())
}

/// 🆕 Handler pour les connexions WebSocket
async fn websocket_handler(
    ws: WebSocketUpgrade,
    State(state): State<AppState>,
) -> Response {
    info!("🔌 Nouvelle connexion WebSocket demandée");
    
    ws.on_upgrade(move |socket| handle_websocket(socket, state))
}

/// 🆕 Gestion d'une connexion WebSocket individuelle
async fn handle_websocket(socket: WebSocket, state: AppState) {
    info!("✅ Connexion WebSocket établie");
    
    let (mut sender, mut receiver) = socket.split();
    
    // Envoyer un message de bienvenue
    let welcome_msg = OutgoingMessage::ActionConfirmed {
        action: "connected".to_string(),
        success: true,
    };
    
    if let Ok(json) = serde_json::to_string(&welcome_msg) {
        if sender.send(Message::Text(json)).await.is_err() {
            error!("❌ Impossible d'envoyer le message de bienvenue");
            return;
        }
    }
    
    // Boucle de gestion des messages
    while let Some(msg) = receiver.next().await {
        match msg {
            Ok(Message::Text(text)) => {
                info!("📨 Message WebSocket reçu: {}", text);
                
                // Parser le message JSON
                match serde_json::from_str::<IncomingMessage>(&text) {
                    Ok(incoming) => {
                        if let Err(e) = handle_incoming_message(incoming, &state, &mut sender).await {
                            warn!("⚠️ Erreur traitement message: {}", e);
                        }
                    }
                    Err(e) => {
                        warn!("⚠️ Message JSON invalide: {}", e);
                        let error_msg = OutgoingMessage::Error {
                            message: format!("Message JSON invalide: {}", e),
                        };
                        if let Ok(json) = serde_json::to_string(&error_msg) {
                            let _ = sender.send(Message::Text(json)).await;
                        }
                    }
                }
            }
            Ok(Message::Close(_)) => {
                info!("👋 Connexion WebSocket fermée");
                break;
            }
            Ok(Message::Ping(data)) => {
                info!("🏓 Ping reçu");
                if sender.send(Message::Pong(data)).await.is_err() {
                    break;
                }
            }
            Ok(_) => {
                // Ignore les autres types de messages
            }
            Err(e) => {
                error!("❌ Erreur WebSocket: {}", e);
                break;
            }
        }
    }
    
    info!("🔌 Connexion WebSocket terminée");
}

/// 🆕 Traite un message entrant via WebSocket
async fn handle_incoming_message(
    message: IncomingMessage,
    state: &AppState,
    sender: &mut futures_util::stream::SplitSink<WebSocket, Message>,
) -> Result<(), ChatError> {
    match message {
        IncomingMessage::SendMessage { conversation_id, content, parent_message_id: _ } => {
            info!("💬 Envoi de message via WebSocket: '{}'", content);
            
            // Convertir l'UUID en room_id (simplifié)
            let room_id = conversation_id.to_string();
            
            // Enregistrer le message via le store
            let message_id = state.store.send_simple_message(
                &content,
                "websocket_user", // TODO: Utiliser le vrai utilisateur authentifié
                Some(&room_id),
                false,
            ).await?;
            
            // Confirmer l'envoi
            let confirmation = OutgoingMessage::ActionConfirmed {
                action: "message_sent".to_string(),
                success: true,
            };
            
            if let Ok(json) = serde_json::to_string(&confirmation) {
                if let Err(e) = sender.send(Message::Text(json)).await {
                    warn!("⚠️ Erreur envoi confirmation: {}", e);
                    return Err(ChatError::configuration_error("Erreur envoi confirmation"));
                }
            }
            
            info!("✅ Message WebSocket envoyé - ID: {}", message_id);
        }
        
        IncomingMessage::JoinConversation { conversation_id } => {
            info!("👥 Rejoindre conversation: {}", conversation_id);
            
            let confirmation = OutgoingMessage::ActionConfirmed {
                action: "joined_conversation".to_string(),
                success: true,
            };
            
            if let Ok(json) = serde_json::to_string(&confirmation) {
                if let Err(e) = sender.send(Message::Text(json)).await {
                    warn!("⚠️ Erreur envoi confirmation join: {}", e);
                    return Err(ChatError::configuration_error("Erreur envoi confirmation join"));
                }
            }
        }
        
        IncomingMessage::LeaveConversation { conversation_id } => {
            info!("👋 Quitter conversation: {}", conversation_id);
            
            let confirmation = OutgoingMessage::ActionConfirmed {
                action: "left_conversation".to_string(),
                success: true,
            };
            
            if let Ok(json) = serde_json::to_string(&confirmation) {
                if let Err(e) = sender.send(Message::Text(json)).await {
                    warn!("⚠️ Erreur envoi confirmation leave: {}", e);
                    return Err(ChatError::configuration_error("Erreur envoi confirmation leave"));
                }
            }
        }
        
        IncomingMessage::MarkAsRead { conversation_id, message_id } => {
            info!("✓ Marquer comme lu: conversation={}, message={}", conversation_id, message_id);
            
            let confirmation = OutgoingMessage::ActionConfirmed {
                action: "marked_as_read".to_string(),
                success: true,
            };
            
            if let Ok(json) = serde_json::to_string(&confirmation) {
                if let Err(e) = sender.send(Message::Text(json)).await {
                    warn!("⚠️ Erreur envoi confirmation read: {}", e);
                    return Err(ChatError::configuration_error("Erreur envoi confirmation read"));
                }
            }
        }
        
        IncomingMessage::Ping => {
            info!("🏓 Ping WebSocket");
            
            let pong = OutgoingMessage::Pong;
            if let Ok(json) = serde_json::to_string(&pong) {
                if let Err(e) = sender.send(Message::Text(json)).await {
                    warn!("⚠️ Erreur envoi pong: {}", e);
                    return Err(ChatError::configuration_error("Erreur envoi pong"));
                }
            }
        }
    }
    
    Ok(())
}

/// Endpoint de vérification de santé
async fn health_check() -> Json<ApiResponse<HashMap<String, String>>> {
    let mut info = HashMap::new();
    info.insert("status".to_string(), "healthy".to_string());
    info.insert("service".to_string(), "veza-chat-server".to_string());
    info.insert("version".to_string(), "0.3.0".to_string());
    info.insert("websocket".to_string(), "enabled".to_string());
    
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
    stats.insert("websocket_enabled".to_string(), 1);
    
    Json(ApiResponse::success(stats))
}
