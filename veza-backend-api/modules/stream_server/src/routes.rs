use axum::{
    extract::{Path},
    http::{HeaderMap, StatusCode, Request},
    routing::get,
    Router,
    response::Response,
    body::Body,
};
use std::{path::PathBuf, collections::HashMap};
use crate::utils::validate_signature;
use url::form_urlencoded;

pub fn stream_route() -> Router {
    Router::new()
        .route("/:id", get(stream_audio))
}

pub async fn stream_audio(
    Path(id): Path<String>,
    headers: HeaderMap,
    request: Request<Body>,
) -> Result<Response, (StatusCode, String)> {
    let file_path = format!("audio/{}.mp3", id);
    let path = PathBuf::from(&file_path);

    let query = request.uri().query().unwrap_or("");
    let params: HashMap<_, _> = form_urlencoded::parse(query.as_bytes()).into_owned().collect();

    let expires = match params.get("expires") {
        Some(e) => e,
        None => return Err((StatusCode::FORBIDDEN, "Paramètre expires manquant".into())),
    };

    let sig = match params.get("sig") {
        Some(s) => s,
        None => return Err((StatusCode::FORBIDDEN, "Signature manquante".into())),
    };

    if !validate_signature(&id, expires, sig) {
        return Err((StatusCode::FORBIDDEN, "Signature invalide".into()));
    }

    if !path.exists() {
        return Err((StatusCode::NOT_FOUND, "Fichier non trouvé".into()));
    }

    crate::utils::serve_partial_file(path, headers).await
}
