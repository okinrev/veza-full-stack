use axum::{
    extract::{Path, Query, State},
    http::{header, HeaderMap, StatusCode},
    response::Json,
    routing::get,
    Router,
};
use serde::{Deserialize, Serialize};
use std::{
    net::SocketAddr,
    path::PathBuf,
    time::{SystemTime, UNIX_EPOCH},
};
use tokio::{fs, signal, io::AsyncReadExt};
use tower_http::{
    compression::CompressionLayer,
    cors::{Any, CorsLayer},
};
use tracing::{error, info, warn};

#[derive(Clone)]
struct AppState {
    audio_dir: PathBuf,
    port: u16,
}

#[derive(Serialize)]
struct HealthResponse {
    status: String,
    timestamp: u64,
    service: String,
    version: String,
}

#[derive(Serialize)]
struct StreamInfo {
    filename: String,
    size: Option<u64>,
    content_type: String,
}

#[derive(Deserialize)]
struct StreamParams {
    quality: Option<String>,
    start: Option<u64>,
    duration: Option<u64>,
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    // Configuration du logging
    tracing_subscriber::fmt::init();
    info!("🎵 Démarrage du Stream Server Simplifié");

    // Configuration
    let audio_dir = std::env::var("AUDIO_DIR")
        .unwrap_or_else(|_| "/opt/veza-stream/audio".to_string());
    let port = std::env::var("PORT")
        .unwrap_or_else(|_| "8000".to_string())
        .parse::<u16>()
        .unwrap_or(8000);

    // Créer le répertoire audio s'il n'existe pas
    tokio::fs::create_dir_all(&audio_dir).await?;
    
    let state = AppState {
        audio_dir: PathBuf::from(audio_dir.clone()),
        port,
    };

    info!("📁 Répertoire audio: {}", audio_dir);
    info!("🔌 Port: {}", port);

    // Configuration CORS permissive
    let cors = CorsLayer::new()
        .allow_origin(Any)
        .allow_methods([axum::http::Method::GET, axum::http::Method::OPTIONS])
        .allow_headers(Any);

    // Routes
    let app = Router::new()
        .route("/", get(root))
        .route("/health", get(health_check))
        .route("/stream/:filename", get(stream_audio))
        .route("/info/:filename", get(audio_info))
        .route("/list", get(list_audio_files))
        .layer(CompressionLayer::new())
        .layer(cors)
        .with_state(state);

    // Démarrage du serveur
    let addr = SocketAddr::from(([0, 0, 0, 0], port));
    info!("🌐 Serveur démarré sur http://{}", addr);

    let listener = tokio::net::TcpListener::bind(&addr).await?;
    axum::serve(listener, app)
        .with_graceful_shutdown(shutdown_signal())
        .await?;

    info!("👋 Serveur arrêté");
    Ok(())
}

async fn root() -> &'static str {
    "🎵 Veza Stream Server - Serveur de streaming audio simplifié"
}

async fn health_check() -> Json<HealthResponse> {
    Json(HealthResponse {
        status: "healthy".to_string(),
        timestamp: SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .unwrap()
            .as_secs(),
        service: "veza-stream-server".to_string(),
        version: "0.2.0".to_string(),
    })
}

async fn list_audio_files(State(state): State<AppState>) -> Result<Json<Vec<String>>, StatusCode> {
    let mut files = Vec::new();
    
    let mut dir = match fs::read_dir(&state.audio_dir).await {
        Ok(dir) => dir,
        Err(_) => return Ok(Json(files)),
    };

    while let Some(entry) = dir.next_entry().await.unwrap_or(None) {
        if let Some(name) = entry.file_name().to_str() {
            if name.ends_with(".mp3") || name.ends_with(".wav") || name.ends_with(".flac") {
                files.push(name.to_string());
            }
        }
    }

    files.sort();
    Ok(Json(files))
}

async fn audio_info(
    Path(filename): Path<String>,
    State(state): State<AppState>,
) -> Result<Json<StreamInfo>, StatusCode> {
    let file_path = state.audio_dir.join(&filename);
    
    if !file_path.exists() {
        return Err(StatusCode::NOT_FOUND);
    }

    let metadata = match fs::metadata(&file_path).await {
        Ok(meta) => meta,
        Err(_) => return Err(StatusCode::INTERNAL_SERVER_ERROR),
    };

    let content_type = match file_path.extension().and_then(|s| s.to_str()) {
        Some("mp3") => "audio/mpeg",
        Some("wav") => "audio/wav",
        Some("flac") => "audio/flac",
        _ => "application/octet-stream",
    };

    Ok(Json(StreamInfo {
        filename,
        size: Some(metadata.len()),
        content_type: content_type.to_string(),
    }))
}

async fn stream_audio(
    Path(filename): Path<String>,
    Query(_params): Query<StreamParams>,
    State(state): State<AppState>,
    headers: HeaderMap,
) -> Result<axum::response::Response, StatusCode> {
    let file_path = state.audio_dir.join(&filename);
    
    if !file_path.exists() {
        warn!("Fichier non trouvé: {:?}", file_path);
        return Err(StatusCode::NOT_FOUND);
    }

    let file = match fs::File::open(&file_path).await {
        Ok(file) => file,
        Err(e) => {
            error!("Erreur ouverture fichier: {}", e);
            return Err(StatusCode::INTERNAL_SERVER_ERROR);
        }
    };

    let metadata = match file.metadata().await {
        Ok(meta) => meta,
        Err(_) => return Err(StatusCode::INTERNAL_SERVER_ERROR),
    };

    let file_size = metadata.len();
    let content_type = match file_path.extension().and_then(|s| s.to_str()) {
        Some("mp3") => "audio/mpeg",
        Some("wav") => "audio/wav",
        Some("flac") => "audio/flac",
        _ => "application/octet-stream",
    };

    // Support pour les requêtes Range (streaming partiel)
    if let Some(range) = headers.get(header::RANGE) {
        if let Ok(range_str) = range.to_str() {
            if let Some(range_value) = parse_range_header(range_str, file_size) {
                let (start, end) = range_value;
                let content_length = end - start + 1;
                
                info!("Streaming partiel: {}-{}/{} pour {}", start, end, file_size, filename);
                
                let stream = create_partial_stream(file, start, content_length).await?;
                
                return Ok(axum::response::Response::builder()
                    .status(StatusCode::PARTIAL_CONTENT)
                    .header(header::CONTENT_TYPE, content_type)
                    .header(header::CONTENT_LENGTH, content_length.to_string())
                    .header(header::CONTENT_RANGE, format!("bytes {}-{}/{}", start, end, file_size))
                    .header(header::ACCEPT_RANGES, "bytes")
                    .body(axum::body::Body::from_stream(stream))
                    .unwrap());
            }
        }
    }

    // Streaming complet
    info!("Streaming complet: {} ({} bytes)", filename, file_size);
    let stream = create_full_stream(file).await?;
    
    Ok(axum::response::Response::builder()
        .status(StatusCode::OK)
        .header(header::CONTENT_TYPE, content_type)
        .header(header::CONTENT_LENGTH, file_size.to_string())
        .header(header::ACCEPT_RANGES, "bytes")
        .body(axum::body::Body::from_stream(stream))
        .unwrap())
}

fn parse_range_header(range: &str, file_size: u64) -> Option<(u64, u64)> {
    if !range.starts_with("bytes=") {
        return None;
    }
    
    let range = &range[6..];
    let parts: Vec<&str> = range.split('-').collect();
    
    if parts.len() != 2 {
        return None;
    }
    
    let start = if parts[0].is_empty() {
        file_size.saturating_sub(parts[1].parse::<u64>().ok()?)
    } else {
        parts[0].parse::<u64>().ok()?
    };
    
    let end = if parts[1].is_empty() {
        file_size - 1
    } else {
        parts[1].parse::<u64>().ok()?.min(file_size - 1)
    };
    
    if start <= end && end < file_size {
        Some((start, end))
    } else {
        None
    }
}

async fn create_full_stream(
    file: tokio::fs::File,
) -> Result<impl futures_util::Stream<Item = Result<bytes::Bytes, std::io::Error>>, StatusCode> {
    use tokio_util::io::ReaderStream;
    Ok(ReaderStream::new(file))
}

async fn create_partial_stream(
    mut file: tokio::fs::File,
    start: u64,
    length: u64,
) -> Result<impl futures_util::Stream<Item = Result<bytes::Bytes, std::io::Error>>, StatusCode> {
    use tokio::io::{AsyncSeekExt, SeekFrom};
    use tokio_util::io::ReaderStream;
    
    if let Err(_) = file.seek(SeekFrom::Start(start)).await {
        return Err(StatusCode::INTERNAL_SERVER_ERROR);
    }
    
    let limited_reader = file.take(length);
    Ok(ReaderStream::new(limited_reader))
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
 