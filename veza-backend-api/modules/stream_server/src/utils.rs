// file: qstream_server/src/utils.rs

use axum::http::{HeaderMap, StatusCode};
use axum::response::{Response};
use tokio::fs::File;
use tokio::io::{AsyncReadExt, SeekFrom, AsyncSeekExt};
use std::path::PathBuf;
use mime_guess;
use hmac::{Hmac, Mac};
use sha2::Sha256;
use chrono::Utc;

const SECRET_KEY: &str = "&!zg0N4f1L4@0Z1NkUc9heFiwZMI4KmRzVd0viTU7#JO$YrJ&g!!h54G";

pub async fn serve_partial_file(path: PathBuf, headers: HeaderMap) -> Result<Response, (StatusCode, String)> {
    let metadata = tokio::fs::metadata(&path).await.map_err(|_| {
        (StatusCode::INTERNAL_SERVER_ERROR, "Erreur de lecture fichier".into())
    })?;
    let total_size = metadata.len();

    let range = headers
        .get("range")
        .and_then(|v| v.to_str().ok())
        .and_then(parse_range);

    let mime = mime_guess::from_path(&path).first_or_octet_stream();

    let mut file = File::open(&path).await.map_err(|_| {
        (StatusCode::INTERNAL_SERVER_ERROR, "Erreur ouverture fichier".into())
    })?;

    let (start, end) = range.unwrap_or((0, total_size - 1));
    let chunk_size = end - start + 1;

    file.seek(SeekFrom::Start(start)).await.unwrap();
    let mut buffer = vec![0; chunk_size as usize];
    file.read_exact(&mut buffer).await.unwrap();

    let body = axum::body::Body::from(buffer);

    let mut response = Response::new(body);
    *response.status_mut() = StatusCode::PARTIAL_CONTENT;

    let headers = response.headers_mut();
    headers.insert("Content-Type", mime.to_string().parse().unwrap());
    headers.insert("Accept-Ranges", "bytes".parse().unwrap());
    headers.insert("Content-Length", chunk_size.to_string().parse().unwrap());
    headers.insert("Content-Range", format!("bytes {}-{}/{}", start, end, total_size).parse().unwrap());

    Ok(response)
}

fn parse_range(header: &str) -> Option<(u64, u64)> {
    // Ex: "bytes=0-1023"
    if !header.starts_with("bytes=") {
        return None;
    }
    let parts: Vec<_> = header["bytes=".len()..].split('-').collect();
    if parts.len() != 2 {
        return None;
    }
    let start = parts[0].parse::<u64>().ok()?;
    let end = parts[1].parse::<u64>().ok()?;
    Some((start, end))
}

pub fn validate_signature(filename: &str, expires: &str, sig: &str) -> bool {
    let expires_int = match expires.parse::<i64>() {
        Ok(val) => val,
        Err(_) => return false,
    };

    if chrono::Utc::now().timestamp() > expires_int {
        return false;
    }

    let to_sign = format!("{}|{}", filename, expires);
    let mut mac = Hmac::<Sha256>::new_from_slice(SECRET_KEY.as_bytes()).unwrap();
    mac.update(to_sign.as_bytes());
    let result = mac.finalize();
    let expected_sig = hex::encode(result.into_bytes());

    expected_sig == sig
}
