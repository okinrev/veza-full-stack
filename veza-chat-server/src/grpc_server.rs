//! Module serveur gRPC pour le Chat Server

use std::sync::Arc;
use tonic::{transport::Server, Request, Response, Status};
use tracing::{info, debug};

// Importation des bindings protobuf générés
pub mod chat {
    include!("generated/veza.chat.rs");
}

pub mod auth {
    include!("generated/veza.common.auth.rs");
}

use chat::{
    chat_service_server::{ChatService, ChatServiceServer},
    *,
};

use crate::{
    config::ServerConfig,
    simple_message_store::SimpleMessageStore,
};

/// Implémentation du service gRPC Chat
#[derive(Clone)]
pub struct ChatServiceImpl {
    pub config: Arc<ServerConfig>,
    pub message_store: Arc<SimpleMessageStore>,
}

impl ChatServiceImpl {
    pub fn new(config: Arc<ServerConfig>, message_store: Arc<SimpleMessageStore>) -> Self {
        Self {
            config,
            message_store,
        }
    }
}

#[tonic::async_trait]
impl ChatService for ChatServiceImpl {
    /// Créer une salle de chat
    async fn create_room(
        &self,
        request: Request<CreateRoomRequest>,
    ) -> Result<Response<CreateRoomResponse>, Status> {
        let _req = request.into_inner();
        debug!("Creating room: {}", _req.name);

        // Validation des données
        if _req.name.trim().is_empty() {
            return Ok(Response::new(CreateRoomResponse {
                room: None,
                error: "Room name cannot be empty".to_string(),
            }));
        }

        // Génération d'un ID unique pour la salle
        let room_id = uuid::Uuid::new_v4().to_string();
        let now = chrono::Utc::now().timestamp();

        // Création de la salle
        let room = Room {
            id: room_id.clone(),
            name: _req.name.clone(),
            description: _req.description.clone(),
            r#type: _req.r#type,
            visibility: _req.visibility,
            created_by: _req.created_by,
            created_at: now,
            member_count: 1,
            online_count: 1,
            is_active: true,
        };

        info!("Room created: {} (ID: {})", _req.name, room_id);

        Ok(Response::new(CreateRoomResponse {
            room: Some(room),
            error: String::new(),
        }))
    }

    /// Rejoindre une salle
    async fn join_room(
        &self,
        request: Request<JoinRoomRequest>,
    ) -> Result<Response<JoinRoomResponse>, Status> {
        let _req = request.into_inner();
        debug!("User {} joining room {}", _req.user_id, _req.room_id);

        // Création du membre
        let member = RoomMember {
            user_id: _req.user_id,
            username: format!("user_{}", _req.user_id),
            role: 0, // Member
            joined_at: chrono::Utc::now().timestamp(),
            is_online: true,
            last_seen: chrono::Utc::now().timestamp(),
        };

        info!("User {} joined room {}", _req.user_id, _req.room_id);

        Ok(Response::new(JoinRoomResponse {
            success: true,
            member: Some(member),
            error: String::new(),
        }))
    }

    /// Envoyer un message  
    async fn send_message(
        &self,
        request: Request<SendMessageRequest>,
    ) -> Result<Response<SendMessageResponse>, Status> {
        let _req = request.into_inner();
        debug!("Sending message to room {} from user {}", _req.room_id, _req.sender_id);

        // Validation
        if _req.content.trim().is_empty() {
            return Ok(Response::new(SendMessageResponse {
                message: None,
                error: "Message content cannot be empty".to_string(),
            }));
        }

        let message_id = uuid::Uuid::new_v4().to_string();
        let now = chrono::Utc::now().timestamp();

        let message = Message {
            id: message_id.clone(),
            room_id: _req.room_id.clone(),
            sender_id: _req.sender_id,
            sender_username: format!("user_{}", _req.sender_id),
            content: _req.content.clone(),
            r#type: _req.r#type,
            created_at: now,
            updated_at: now,
            is_edited: false,
            is_deleted: false,
            reply_to: _req.reply_to.clone(),
            reactions: vec![],
        };

        info!("Message sent: {} in room {}", message_id, _req.room_id);

        Ok(Response::new(SendMessageResponse {
            message: Some(message),
            error: String::new(),
        }))
    }

    // Implémentation simplifiée des autres méthodes
    async fn leave_room(&self, request: Request<LeaveRoomRequest>) -> Result<Response<LeaveRoomResponse>, Status> {
        let _req = request.into_inner();
        Ok(Response::new(LeaveRoomResponse { success: true, error: String::new() }))
    }
    
    async fn get_room_info(&self, request: Request<GetRoomInfoRequest>) -> Result<Response<Room>, Status> {
        let _req = request.into_inner();
        let room = Room {
            id: _req.room_id,
            name: "Demo Room".to_string(),
            description: "Test room".to_string(),
            r#type: 0, // Public
            visibility: 0, // Open
            created_by: 1,
            created_at: chrono::Utc::now().timestamp(),
            member_count: 1,
            online_count: 1,
            is_active: true,
        };
        Ok(Response::new(room))
    }
    
    async fn list_rooms(&self, _request: Request<ListRoomsRequest>) -> Result<Response<ListRoomsResponse>, Status> {
        Ok(Response::new(ListRoomsResponse { rooms: vec![], total: 0, error: String::new() }))
    }
    
    async fn get_message_history(&self, _request: Request<GetMessageHistoryRequest>) -> Result<Response<GetMessageHistoryResponse>, Status> {
        Ok(Response::new(GetMessageHistoryResponse { messages: vec![], has_more: false, error: String::new() }))
    }
    
    async fn delete_message(&self, _request: Request<DeleteMessageRequest>) -> Result<Response<DeleteMessageResponse>, Status> {
        Ok(Response::new(DeleteMessageResponse { success: true, error: String::new() }))
    }
    
    async fn send_direct_message(&self, _request: Request<SendDirectMessageRequest>) -> Result<Response<SendDirectMessageResponse>, Status> {
        Ok(Response::new(SendDirectMessageResponse { message: None, error: String::new() }))
    }
    
    async fn get_direct_messages(&self, _request: Request<GetDirectMessagesRequest>) -> Result<Response<GetDirectMessagesResponse>, Status> {
        Ok(Response::new(GetDirectMessagesResponse { messages: vec![], has_more: false, error: String::new() }))
    }
    
    async fn mute_user(&self, _request: Request<MuteUserRequest>) -> Result<Response<MuteUserResponse>, Status> {
        Ok(Response::new(MuteUserResponse { success: true, error: String::new() }))
    }
    
    async fn ban_user(&self, _request: Request<BanUserRequest>) -> Result<Response<BanUserResponse>, Status> {
        Ok(Response::new(BanUserResponse { success: true, error: String::new() }))
    }
    
    async fn moderate_message(&self, _request: Request<ModerateMessageRequest>) -> Result<Response<ModerateMessageResponse>, Status> {
        Ok(Response::new(ModerateMessageResponse { success: true, error: String::new() }))
    }
    
    async fn get_room_stats(&self, request: Request<GetRoomStatsRequest>) -> Result<Response<RoomStats>, Status> {
        let _req = request.into_inner();
        let stats = RoomStats {
            room_id: _req.room_id,
            total_members: 1,
            online_members: 1,
            messages_today: 0,
            total_messages: 0,
            active_users: vec![],
        };
        Ok(Response::new(stats))
    }
    
    async fn get_user_activity(&self, request: Request<GetUserActivityRequest>) -> Result<Response<UserActivity>, Status> {
        let _req = request.into_inner();
        let activity = UserActivity {
            user_id: _req.user_id,
            rooms_joined: 0,
            messages_sent: 0,
            last_activity: chrono::Utc::now().timestamp(),
            is_online: true,
            current_status: "active".to_string(),
        };
        Ok(Response::new(activity))
    }
}

/// Démarrer le serveur gRPC du chat
pub async fn start_grpc_server(
    config: Arc<crate::config::ServerConfig>,
    message_store: Arc<SimpleMessageStore>,
) -> Result<(), Box<dyn std::error::Error>> {
    let addr = format!("0.0.0.0:{}", config.server.grpc_port).parse()?;
    let chat_service = ChatServiceImpl::new(config.clone(), message_store);

    info!("🚀 Chat gRPC Server starting on {}", addr);

    Server::builder()
        .add_service(ChatServiceServer::new(chat_service))
        .serve(addr)
        .await?;

    Ok(())
} 