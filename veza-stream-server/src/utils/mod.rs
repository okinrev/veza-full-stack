// file: qstream_server/src/utils.rs

pub mod metrics;
pub mod signature;

use crate::config::Config;
use crate::error::{AppError, Result};
use axum::{
    body::Body,
    http::{HeaderMap, HeaderValue, StatusCode},
    response::Response,
};
use std::path::{Path, PathBuf};
use std::sync::Arc;
use tokio::fs::File;
use tokio::io::AsyncReadExt;
use tokio_util::io::ReaderStream;

pub use metrics::*;

pub fn validate_filename(id: &str) -> Result<String> {
    if id.is_empty() || id.len() > 255 {
        return Err(AppError::ValidationError("Invalid filename length".to_string()));
    }

    if id.contains("..") || id.contains('/') || id.contains('\\') {
        return Err(AppError::ValidationError("Path traversal attempt detected".to_string()));
    }

    let filename = id.trim();
    let allowed_chars = filename.chars().all(|c| {
        c.is_alphanumeric() || c == '.' || c == '-' || c == '_' || c == ' '
    });

    if !allowed_chars {
        return Err(AppError::ValidationError("Invalid characters in filename".to_string()));
    }

    Ok(filename.to_string())
}

pub fn build_safe_path(config: &Config, filename: &str) -> Result<PathBuf> {
    let audio_dir = Path::new(&config.audio_dir);
    let file_path = audio_dir.join(filename);

    let canonical_audio_dir = audio_dir.canonicalize()
        .map_err(|_| AppError::Config("Invalid audio directory".to_string()))?;
    
    let canonical_file_path = file_path.canonicalize()
        .map_err(|_| AppError::FileNotFound)?;

    if !canonical_file_path.starts_with(canonical_audio_dir) {
        return Err(AppError::ValidationError("Path traversal attempt detected".to_string()));
    }

    Ok(canonical_file_path)
}

pub fn parse_range(header: &str, file_size: u64) -> Option<(u64, u64)> {
    let range_str = header.strip_prefix("bytes=")?;
    
    if let Some((start_str, end_str)) = range_str.split_once('-') {
        let start = if start_str.is_empty() {
            0
        } else {
            start_str.parse::<u64>().ok()?
        };
        
        let end = if end_str.is_empty() {
            file_size.saturating_sub(1)
        } else {
            end_str.parse::<u64>().ok()?.min(file_size.saturating_sub(1))
        };
        
        if start <= end && start < file_size {
            Some((start, end))
        } else {
            None
        }
    } else {
        None
    }
}

pub async fn serve_partial_file(
    config: &Config,
    path: PathBuf,
    headers: HeaderMap,
) -> Result<Response<Body>> {
    let file = File::open(&path).await
        .map_err(|_| AppError::FileNotFound)?;
    
    let metadata = file.metadata().await
        .map_err(|_| AppError::Internal("Failed to read file metadata".to_string()))?;
    
    let file_size = metadata.len();
    
    if let Some(range_header) = headers.get("range") {
        if let Ok(range_str) = range_header.to_str() {
            if let Some((start, end)) = parse_range(range_str, file_size) {
                let content_length = end - start + 1;
                
                if content_length > config.max_range_size {
                    return Err(AppError::InvalidRange);
                }
                
                let mut file = file;
                let mut buffer = vec![0; content_length as usize];
                file.read_exact(&mut buffer).await
                    .map_err(|_| AppError::Internal("Failed to read file range".to_string()))?;
                
                let mut response = Response::builder()
                    .status(StatusCode::PARTIAL_CONTENT)
                    .header("Content-Length", content_length.to_string())
                    .header("Content-Range", format!("bytes {}-{}/{}", start, end, file_size))
                    .header("Accept-Ranges", "bytes");
                
                add_security_headers(&mut response);
                
                return Ok(response.body(Body::from(buffer))
                    .map_err(|_| AppError::Internal("Failed to create response".to_string()))?);
            }
        }
    }
    
    let stream = ReaderStream::new(file);
    let body = Body::from_stream(stream);
    
    let mut response = Response::builder()
        .status(StatusCode::OK)
        .header("Content-Length", file_size.to_string())
        .header("Accept-Ranges", "bytes");
    
    add_security_headers(&mut response);
    
    Ok(response.body(body)
        .map_err(|_| AppError::Internal("Failed to create response".to_string()))?)
}

pub fn validate_signature(config: &Config, filename: &str, expires: &str, sig: &str) -> bool {
    let expires_timestamp = match expires.parse::<i64>() {
        Ok(timestamp) => timestamp,
        Err(_) => return false,
    };
    
    let now = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .unwrap()
        .as_secs() as i64;
    
    if expires_timestamp < now - config.signature_tolerance {
        return false;
    }
    
    let expected_sig = generate_signature(filename, expires_timestamp, &config.secret_key);
    
    use subtle::ConstantTimeEq;
    expected_sig.as_bytes().ct_eq(sig.as_bytes()).into()
}

fn generate_signature(filename: &str, expires: i64, secret: &str) -> String {
    use hmac::{Hmac, Mac};
    use sha2::Sha256;
    
    type HmacSha256 = Hmac<Sha256>;
    
    let message = format!("{}:{}", filename, expires);
    let mut mac = HmacSha256::new_from_slice(secret.as_bytes()).unwrap();
    mac.update(message.as_bytes());
    let result = mac.finalize();
    
    hex::encode(result.into_bytes())
}

fn add_security_headers(response: &mut axum::http::response::Builder) {
    if let Some(headers) = response.headers_mut() {
        headers.insert("X-Content-Type-Options", HeaderValue::from_static("nosniff"));
        headers.insert("X-Frame-Options", HeaderValue::from_static("DENY"));
        headers.insert("X-XSS-Protection", HeaderValue::from_static("1; mode=block"));
        headers.insert("Content-Security-Policy", HeaderValue::from_static("default-src 'none'; media-src 'self'"));
        headers.insert("Referrer-Policy", HeaderValue::from_static("strict-origin-when-cross-origin"));
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_validate_filename() {
        assert!(validate_filename("test.mp3").is_ok());
        assert!(validate_filename("my-song_01.wav").is_ok());
        assert!(validate_filename("../../../etc/passwd").is_err());
        assert!(validate_filename("file/with/slash").is_err());
        assert!(validate_filename("file\\with\\backslash").is_err());
        assert!(validate_filename("").is_err());
    }

    #[test]
    fn test_parse_range() {
        // Test range normal
        assert_eq!(parse_range("bytes=0-1023", 2048), Some((0, 1023)));
        assert_eq!(parse_range("bytes=1024-2047", 2048), Some((1024, 2047)));
        
        // Test range ouvert
        assert_eq!(parse_range("bytes=1024-", 2048), Some((1024, 2047)));
        
        // Test suffix range
        assert_eq!(parse_range("bytes=-1024", 2048), Some((1024, 2047)));
        
        // Test range invalide
        assert_eq!(parse_range("bytes=2048-", 2048), None);
        assert_eq!(parse_range("invalid", 2048), None);
    }

    #[test]
    fn test_signature_generation() {
        let secret = "test_secret_key";
        let filename = "test.mp3";
        let expires = 1609459200i64;
        
        let sig1 = generate_signature(filename, expires, secret);
        let sig2 = generate_signature(filename, expires, secret);
        
        // Les signatures doivent être identiques
        assert_eq!(sig1, sig2);
        
        // Une signature différente avec d'autres paramètres
        let sig3 = generate_signature("other.mp3", expires, secret);
        assert_ne!(sig1, sig3);
    }
}
