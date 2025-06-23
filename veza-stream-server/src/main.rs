// file: stream_server/src/main.rs

use stream_server::{
    config::Config,
    middleware::{
        logging::request_logging_middleware,
        rate_limit::rate_limit_middleware,
        security::security_headers_middleware,
    },
    AppState,
};
use axum::{
    http::{header, HeaderValue, Method},
    response::Json,
    routing::get,
    Router,
};
use std::{collections::HashMap, net::SocketAddr, sync::Arc, time::Duration};
use tokio::signal;
use tower::ServiceBuilder;
use tower_http::{
    compression::CompressionLayer,
    cors::{AllowOrigin, Any, CorsLayer},
    timeout::TimeoutLayer,
};
use tracing::{error, info, warn};
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};


#[tokio::main]
async fn main() -> std::result::Result<(), Box<dyn std::error::Error>> {
    // Configuration du système de logging
    init_logging()?;
    
    info!("🚀 Démarrage du Stream Server");
    
    // Chargement de la configuration
    let config = Arc::new(Config::from_env()
        .map_err(|e| format!("Erreur de configuration: {}", e))?);
    config.validate()
        .map_err(|e| format!("Erreur de validation: {}", e))?;
    
    info!("✅ Configuration chargée et validée");
    
    // Création de l'état de l'application
    let app_state = create_app_state(config.clone()).await?;
    
    info!("✅ État de l'application initialisé");
    
    // Démarrage des tâches de background
    start_background_tasks(&app_state).await;
    
    // Création du routeur avec tous les middlewares
    let app = create_router(app_state);
    
    // Configuration de l'adresse d'écoute
    let addr = SocketAddr::from(([0, 0, 0, 0], config.port));
    
    info!("🌐 Serveur démarré sur http://{}", addr);
    info!("📁 Répertoire audio: {}", config.audio_dir);
    info!("🔐 Origines autorisées: {:?}", config.allowed_origins);
    
    // Démarrage du serveur avec graceful shutdown
    let listener = tokio::net::TcpListener::bind(&addr).await
        .map_err(|e| format!("Impossible de démarrer le serveur: {}", e))?;
    
    axum::serve(listener, app)
        .with_graceful_shutdown(shutdown_signal())
        .await
        .map_err(|e| format!("Erreur du serveur: {}", e))?;
    
    info!("👋 Serveur arrêté proprement");
    Ok(())
}

fn init_logging() -> std::result::Result<(), Box<dyn std::error::Error>> {
    let log_level = std::env::var("RUST_LOG")
        .unwrap_or_else(|_| "stream_server=info,tower_http=debug".to_string());
    
    tracing_subscriber::registry()
        .with(
            tracing_subscriber::EnvFilter::try_from_default_env()
                .unwrap_or_else(|_| log_level.into()),
        )
        .with(tracing_subscriber::fmt::layer().with_target(false))
        .init();
    
    Ok(())
}

async fn create_app_state(config: Arc<Config>) -> std::result::Result<AppState, Box<dyn std::error::Error>> {
    use stream_server::{
        analytics::AnalyticsEngine,
        audio::{compression::CompressionEngine, processing::AudioProcessor},
        auth::AuthManager,
        cache::FileCache,
        health::HealthMonitor,
        notifications::NotificationService,
        streaming::{adaptive::AdaptiveStreamingManager, websocket::WebSocketManager},
        utils::metrics::Metrics,
    };
    
    // Création du cache de fichiers
    let cache = Arc::new(FileCache::new(
        Duration::from_secs(config.cache.ttl_seconds),
        (config.cache.max_size_mb * 1024 * 1024) as usize,
    ));
    
    // Création du système de métriques
    let metrics = Arc::new(Metrics::new(config.clone()));
    
    // Création du moteur d'analytics
    let analytics = Arc::new(
        AnalyticsEngine::new(&config.database.url, config.clone())
            .await
            .map_err(|e| format!("Erreur analytics: {}", e))?,
    );
    
    // Création du processeur audio
    let audio_processor = Arc::new(AudioProcessor::new(config.clone()));
    
    // Création du gestionnaire de streaming adaptatif
    let adaptive_streaming = Arc::new(AdaptiveStreamingManager::new(config.clone()));
    
    // Création du moniteur de santé
    let health_monitor = Arc::new(HealthMonitor::new(config.clone()));
    
    // Création du gestionnaire d'authentification
    let auth_manager = Arc::new(
        AuthManager::new(config.clone())
            .map_err(|e| format!("Erreur auth: {}", e))?,
    );
    
    // Création du moteur de compression
    let compression_engine = Arc::new(CompressionEngine::new(config.clone()));
    
    // Création du service de notifications
    let notification_service = Arc::new(NotificationService::new(config.clone()));
    
    // Création du gestionnaire WebSocket
    let websocket_manager = Arc::new(WebSocketManager::new());
    
    Ok(AppState {
        config,
        cache,
        metrics,
        analytics,
        audio_processor,
        adaptive_streaming,
        health_monitor,
        auth_manager,
        compression_engine,
        notification_service,
        websocket_manager,
    })
}

async fn start_background_tasks(state: &AppState) {
    info!("🔄 Démarrage des tâches de background");
    
    // Démarrage des workers de compression
    state.compression_engine.start_workers().await;
    
    // Démarrage des workers de notifications
    state.notification_service.start_delivery_workers().await;
    
    // Démarrage des tâches d'analytics
    state.analytics.start_background_tasks().await;
    
    // Démarrage du monitoring de santé
    state.health_monitor.start_monitoring().await;
    
    // Démarrage de la collecte de métriques
    state.metrics.start_collection().await;
    
    info!("✅ Tâches de background démarrées");
}

fn create_router(state: AppState) -> Router {
    // Configuration CORS
    let cors = if state.config.allowed_origins.contains(&"*".to_string()) {
        warn!("⚠️  CORS configuré pour toutes les origines - non recommandé en production");
        CorsLayer::new()
            .allow_origin(Any)
            .allow_methods([Method::GET, Method::POST, Method::OPTIONS])
            .allow_headers(Any)
            .expose_headers([
                header::CONTENT_RANGE,
                header::CONTENT_LENGTH,
                header::ACCEPT_RANGES,
            ])
    } else {
        let origins: std::result::Result<Vec<_>, _> = state
            .config
            .allowed_origins
            .iter()
            .map(|origin| origin.parse::<HeaderValue>())
            .collect();
        
        match origins {
            Ok(origins) => {
                let mut cors_layer = CorsLayer::new()
                    .allow_methods([Method::GET, Method::POST, Method::OPTIONS])
                    .allow_headers([header::AUTHORIZATION, header::CONTENT_TYPE, header::RANGE])
                    .expose_headers([
                        header::CONTENT_RANGE,
                        header::CONTENT_LENGTH,
                        header::ACCEPT_RANGES,
                    ]);
                
                for origin in origins {
                    cors_layer = cors_layer.allow_origin(AllowOrigin::exact(origin));
                }
                
                cors_layer
            },
            Err(e) => {
                error!("❌ Erreur de configuration CORS: {}", e);
                CorsLayer::new().allow_origin(Any)
            }
        }
    };
    
    // Stack de middlewares
    let middleware_stack = ServiceBuilder::new()
        .layer(TimeoutLayer::new(Duration::from_secs(30)))
        .layer(CompressionLayer::new())
        .layer(cors)
        .layer(axum::middleware::from_fn_with_state(
            state.clone(),
            security_headers_middleware,
        ))
        .layer(axum::middleware::from_fn_with_state(
            state.clone(),
            rate_limit_middleware,
        ))
        .layer(axum::middleware::from_fn_with_state(
            state.clone(),
            request_logging_middleware,
        ));
    
    // Routes principales
    Router::new()
        .route("/", get(|| async { "🎵 Stream Server - Serveur de streaming audio" }))
        .route("/health", get(health_check))
        .route("/health/detailed", get(detailed_health_check))
        .route("/metrics", get(metrics_endpoint))
        .route("/stream/:filename", get(stream_audio))
        .layer(middleware_stack)
        .with_state(state)
}

async fn shutdown_signal() {
    let ctrl_c = async {
        signal::ctrl_c()
            .await
            .expect("Impossible d'installer le handler Ctrl+C");
    };
    
    #[cfg(unix)]
    let terminate = async {
        signal::unix::signal(signal::unix::SignalKind::terminate())
            .expect("Impossible d'installer le handler SIGTERM")
            .recv()
            .await;
    };
    
    #[cfg(not(unix))]
    let terminate = std::future::pending::<()>();
    
    tokio::select! {
        _ = ctrl_c => {
            info!("📱 Signal Ctrl+C reçu, arrêt du serveur...");
        },
        _ = terminate => {
            info!("📱 Signal SIGTERM reçu, arrêt du serveur...");
        }
    }
}

async fn health_check() -> Json<serde_json::Value> {
    Json(serde_json::json!({
        "status": "healthy",
        "timestamp": std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap()
            .as_secs(),
        "service": "stream_server",
        "version": "0.2.0"
    }))
}

async fn detailed_health_check(
    axum::extract::State(state): axum::extract::State<AppState>,
) -> Json<serde_json::Value> {
    let health_status = state.health_monitor.get_health_status().await;
    Json(serde_json::to_value(health_status).unwrap_or_default())
}

async fn metrics_endpoint(
    axum::extract::State(state): axum::extract::State<AppState>,
) -> Json<serde_json::Value> {
    Json(state.metrics.get_metrics().await)
}

async fn stream_audio(
    axum::extract::Path(filename): axum::extract::Path<String>,
    axum::extract::Query(params): axum::extract::Query<HashMap<String, String>>,
    axum::extract::State(state): axum::extract::State<AppState>,
    headers: axum::http::HeaderMap,
) -> std::result::Result<axum::response::Response, (axum::http::StatusCode, String)> {
    use stream_server::{
        error::AppError,
        utils::{validate_filename, build_safe_path, serve_partial_file, validate_signature},
    };
    
    // Validation des paramètres
    let expires = params.get("expires").ok_or((
        axum::http::StatusCode::BAD_REQUEST,
        "Missing expires parameter".to_string(),
    ))?;
    
    let sig = params.get("sig").ok_or((
        axum::http::StatusCode::BAD_REQUEST,
        "Missing signature parameter".to_string(),
    ))?;
    
    // Validation du nom de fichier
    let validated_filename = validate_filename(&filename)
        .map_err(|_| (axum::http::StatusCode::BAD_REQUEST, "Invalid filename".to_string()))?;
    
    // Validation de la signature
    if !validate_signature(&state.config, &validated_filename, expires, sig) {
        return Err((
            axum::http::StatusCode::FORBIDDEN,
            "Invalid signature".to_string(),
        ));
    }
    
    // Construction du chemin sécurisé
    let file_path = build_safe_path(&state.config, &format!("{}.mp3", validated_filename))
        .map_err(|_| (axum::http::StatusCode::NOT_FOUND, "File not found".to_string()))?;
    
    // Streaming du fichier
    serve_partial_file(&state.config, file_path, headers)
        .await
        .map_err(|e| match e {
            AppError::FileNotFound => (axum::http::StatusCode::NOT_FOUND, "File not found".to_string()),
            AppError::InvalidRange => (axum::http::StatusCode::RANGE_NOT_SATISFIABLE, "Invalid range".to_string()),
            _ => (axum::http::StatusCode::INTERNAL_SERVER_ERROR, "Internal error".to_string()),
        })
}
