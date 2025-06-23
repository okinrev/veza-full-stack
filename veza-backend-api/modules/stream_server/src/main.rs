// file: stream_server/src/main.rs

mod routes;
mod utils;

use axum::{serve, Router,  http::Method};
use std::net::SocketAddr;
use tokio::net::TcpListener;
use tracing_subscriber;
use routes::stream_route;
use tower_http::cors::{CorsLayer, Any};
use axum::http::HeaderName;


#[tokio::main]
async fn main() {
    tracing_subscriber::fmt::init();

    let exposed_headers = vec![
        HeaderName::from_static("content-range"),
        HeaderName::from_static("content-length"),
        HeaderName::from_static("accept-ranges"),
    ];

    let cors = CorsLayer::new()
        .allow_origin(Any)
        .allow_methods([Method::GET])
        .allow_headers(Any)
        .expose_headers(exposed_headers);

    let app = Router::new()
        .nest("/stream", stream_route())
        .layer(cors); // ← ici

    let addr = SocketAddr::from(([127, 0, 0, 1], 8082));
    let listener = TcpListener::bind(addr).await.unwrap();

    tracing::info!("🎧 Serveur de streaming démarré sur {}", addr);

    serve(listener, app).await.unwrap();
}
