//! Module serveur gRPC pour le Stream Server

use std::sync::Arc;
use tonic::{transport::Server, Request, Response, Status};
use tracing::{info, debug};
use crate::Config;

// Importation des bindings protobuf générés
pub mod stream {
    include!("generated/veza.stream.rs");
}

pub mod auth {
    include!("generated/veza.common.auth.rs");
}

use stream::{
    stream_service_server::{StreamService, StreamServiceServer},
    *,
};

/// Implémentation du service gRPC Stream
#[derive(Clone)]
pub struct StreamServiceImpl {
    pub config: Arc<Config>,
}

impl StreamServiceImpl {
    pub fn new(config: Arc<Config>) -> Self {
        Self { config }
    }
}

#[tonic::async_trait]
impl StreamService for StreamServiceImpl {
    /// Créer un stream
    async fn create_stream(
        &self,
        request: Request<CreateStreamRequest>,
    ) -> Result<Response<CreateStreamResponse>, Status> {
        let req = request.into_inner();
        debug!("Creating stream: {}", req.title);

        let stream_id = uuid::Uuid::new_v4().to_string();
        let stream_key = uuid::Uuid::new_v4().to_string();

        let stream = Stream {
            id: stream_id.clone(),
            title: req.title.clone(),
            description: req.description.clone(),
            category: req.category,
            visibility: req.visibility,
            streamer_id: req.streamer_id,
            streamer_username: format!("user_{}", req.streamer_id),
            status: 0, // Created
            current_quality: req.default_quality,
            listener_count: 0,
            created_at: chrono::Utc::now().timestamp(),
            started_at: 0,
            duration: 0,
            is_recording: false,
            metadata: None,
        };

        Ok(Response::new(CreateStreamResponse {
            stream: Some(stream),
            stream_key: stream_key.clone(),
            rtmp_url: format!("rtmp://localhost:1935/live/{}", stream_key),
            error: String::new(),
        }))
    }

    /// Démarrer un stream
    async fn start_stream(&self, request: Request<StartStreamRequest>) -> Result<Response<StartStreamResponse>, Status> {
        let req = request.into_inner();
        Ok(Response::new(StartStreamResponse {
            success: true,
            stream_url: format!("http://localhost:8081/stream/{}", req.stream_id),
            hls_urls: vec![],
            error: String::new(),
        }))
    }

    /// Rejoindre un stream  
    async fn join_stream(&self, request: Request<JoinStreamRequest>) -> Result<Response<JoinStreamResponse>, Status> {
        let req = request.into_inner();
        Ok(Response::new(JoinStreamResponse {
            success: true,
            stream_url: format!("http://localhost:8081/stream/{}/listen", req.stream_id),
            actual_quality: req.preferred_quality,
            buffer_duration: 3000,
            error: String::new(),
        }))
    }

    /// Changer la qualité audio
    async fn change_quality(&self, _request: Request<ChangeQualityRequest>) -> Result<Response<ChangeQualityResponse>, Status> {
        Ok(Response::new(ChangeQualityResponse { success: true, new_stream_url: String::new(), error: String::new() }))
    }

    /// Obtenir les métriques audio
    async fn get_audio_metrics(&self, request: Request<GetAudioMetricsRequest>) -> Result<Response<AudioMetrics>, Status> {
        let req = request.into_inner();
        let metrics = AudioMetrics {
            stream_id: req.stream_id,
            current_bitrate: 128,
            buffer_health: 95,
            latency: 150.0,
            dropped_frames: 0,
            quality_stats: None,
            measured_at: chrono::Utc::now().timestamp(),
        };
        Ok(Response::new(metrics))
    }

    // Implémentations simplifiées des autres méthodes
    async fn stop_stream(&self, _request: Request<StopStreamRequest>) -> Result<Response<StopStreamResponse>, Status> {
        Ok(Response::new(StopStreamResponse { success: true, summary: None, error: String::new() }))
    }
    
    async fn get_stream_info(&self, request: Request<GetStreamInfoRequest>) -> Result<Response<Stream>, Status> {
        let req = request.into_inner();
        let stream = Stream {
            id: req.stream_id,
            title: "Demo Stream".to_string(),
            description: "Test stream".to_string(),
            category: 0,
            visibility: 0,
            streamer_id: 1,
            streamer_username: "streamer_1".to_string(),
            status: 1,
            current_quality: 2,
            listener_count: 0,
            created_at: chrono::Utc::now().timestamp(),
            started_at: chrono::Utc::now().timestamp(),
            duration: 0,
            is_recording: false,
            metadata: None,
        };
        Ok(Response::new(stream))
    }
    
    async fn list_active_streams(&self, _request: Request<ListActiveStreamsRequest>) -> Result<Response<ListActiveStreamsResponse>, Status> {
        Ok(Response::new(ListActiveStreamsResponse { streams: vec![], total: 0, error: String::new() }))
    }
    
    async fn leave_stream(&self, _request: Request<LeaveStreamRequest>) -> Result<Response<LeaveStreamResponse>, Status> {
        Ok(Response::new(LeaveStreamResponse { success: true, listen_duration: 0, error: String::new() }))
    }
    
    async fn get_listeners(&self, _request: Request<GetListenersRequest>) -> Result<Response<GetListenersResponse>, Status> {
        Ok(Response::new(GetListenersResponse { listeners: vec![], total_count: 0, error: String::new() }))
    }
    
    async fn set_volume(&self, _request: Request<SetVolumeRequest>) -> Result<Response<SetVolumeResponse>, Status> {
        Ok(Response::new(SetVolumeResponse { success: true, error: String::new() }))
    }
    
    async fn start_recording(&self, _request: Request<StartRecordingRequest>) -> Result<Response<StartRecordingResponse>, Status> {
        Ok(Response::new(StartRecordingResponse { success: true, recording_id: String::new(), error: String::new() }))
    }
    
    async fn stop_recording(&self, _request: Request<StopRecordingRequest>) -> Result<Response<StopRecordingResponse>, Status> {
        Ok(Response::new(StopRecordingResponse { success: true, recording: None, error: String::new() }))
    }
    
    async fn get_recordings(&self, _request: Request<GetRecordingsRequest>) -> Result<Response<GetRecordingsResponse>, Status> {
        Ok(Response::new(GetRecordingsResponse { recordings: vec![], total: 0, error: String::new() }))
    }
    
    async fn get_stream_analytics(&self, request: Request<GetStreamAnalyticsRequest>) -> Result<Response<StreamAnalytics>, Status> {
        let req = request.into_inner();
        let analytics = StreamAnalytics {
            stream_id: req.stream_id,
            start_time: chrono::Utc::now().timestamp(),
            end_time: chrono::Utc::now().timestamp(),
            unique_listeners: 0,
            max_concurrent: 0,
            total_listen_time: 0,
            average_session_duration: 0.0,
            geographic_distribution: std::collections::HashMap::new(),
            hourly_activity: vec![],
        };
        Ok(Response::new(analytics))
    }
    
    async fn get_user_listening_history(&self, _request: Request<GetUserListeningHistoryRequest>) -> Result<Response<UserListeningHistory>, Status> {
        Ok(Response::new(UserListeningHistory {
            user_id: 0,
            sessions: vec![],
            total_listen_time: 0,
            streams_listened: 0,
        }))
    }

    type SubscribeToStreamEventsStream = tokio_stream::wrappers::ReceiverStream<Result<StreamEvent, Status>>;

    async fn subscribe_to_stream_events(&self, _request: Request<SubscribeToStreamEventsRequest>) -> Result<Response<Self::SubscribeToStreamEventsStream>, Status> {
        let (tx, rx) = tokio::sync::mpsc::channel(10);
        let _tx = tx.clone();
        Ok(Response::new(tokio_stream::wrappers::ReceiverStream::new(rx)))
    }
}

/// Démarrer le serveur gRPC du stream
pub async fn start_grpc_server(config: Arc<Config>) -> Result<(), Box<dyn std::error::Error>> {
    let addr = "0.0.0.0:50052".parse()?;
    let stream_service = StreamServiceImpl::new(config);

    info!("🚀 Stream gRPC Server starting on {}", addr);

    Server::builder()
        .add_service(StreamServiceServer::new(stream_service))
        .serve(addr)
        .await?;

    Ok(())
} 