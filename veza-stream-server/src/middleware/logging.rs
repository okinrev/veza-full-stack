use axum::{
    extract::{Request, State},
    http::{HeaderMap, HeaderName, Method, StatusCode, Uri},
    middleware::Next,
    response::Response,
};
use std::time::Instant;
use tracing::{info, warn, error};
use uuid::Uuid;
use crate::AppState;

pub async fn request_logging_middleware(
    State(state): State<AppState>,
    mut request: Request,
    next: Next,
) -> Response {
    let start = Instant::now();
    let _method = request.method().clone();
    let uri = request.uri().clone();
    let headers = request.headers().clone();
    let request_id = Uuid::new_v4();
    
    // Extraire l'IP du client
    let client_ip = extract_client_ip(&headers);
    
    // Ajouter l'ID de requête aux headers
    request.headers_mut().insert(
        HeaderName::from_static("x-request-id"),
        request_id.to_string().parse().unwrap()
    );
    
    // Enregistrer le début de la requête
    info!(
        request_id = %request_id,
        method = %_method,
        uri = %uri,
        client_ip = %client_ip,
        user_agent = ?headers.get("user-agent"),
        "Début de requête"
    );
    
    // Traiter la requête
    let response = next.run(request).await;
    let status = response.status();
    let duration = start.elapsed();
    
    // Incrémenter les métriques
    state.metrics.increment_requests();
    if status.is_success() {
        state.metrics.increment_successful_requests();
    } else {
        state.metrics.increment_failed_requests();
    }
    
    // Détecter les requêtes suspectes
    if is_suspicious_request(&_method, &uri, &headers, status) {
        warn!(
            request_id = %request_id,
            method = %_method,
            uri = %uri,
            client_ip = %client_ip,
            status = %status,
            duration_ms = duration.as_millis(),
            "Requête suspecte détectée"
        );
    }
    
    // Enregistrer la réponse
    let log_level = if status.is_server_error() {
        tracing::Level::ERROR
    } else if status.is_client_error() {
        tracing::Level::WARN
    } else {
        tracing::Level::INFO
    };
    
    match log_level {
        tracing::Level::ERROR => error!(
            request_id = %request_id,
            method = %_method,
            uri = %uri,
            client_ip = %client_ip,
            status = %status,
            duration_ms = duration.as_millis(),
            "Requête terminée avec erreur serveur"
        ),
        tracing::Level::WARN => warn!(
            request_id = %request_id,
            method = %_method,
            uri = %uri,
            client_ip = %client_ip,
            status = %status,
            duration_ms = duration.as_millis(),
            "Requête terminée avec erreur client"
        ),
        _ => info!(
            request_id = %request_id,
            method = %_method,
            uri = %uri,
            client_ip = %client_ip,
            status = %status,
            duration_ms = duration.as_millis(),
            "Requête terminée avec succès"
        ),
    }
    
    response
}

fn extract_client_ip(headers: &HeaderMap) -> String {
    // Vérifier les headers de proxy dans l'ordre de priorité
    if let Some(forwarded_for) = headers.get("x-forwarded-for") {
        if let Ok(forwarded_str) = forwarded_for.to_str() {
            // Prendre la première IP de la liste
            if let Some(first_ip) = forwarded_str.split(',').next() {
                return first_ip.trim().to_string();
            }
        }
    }
    
    if let Some(real_ip) = headers.get("x-real-ip") {
        if let Ok(ip_str) = real_ip.to_str() {
            return ip_str.to_string();
        }
    }
    
    if let Some(forwarded) = headers.get("forwarded") {
        if let Ok(forwarded_str) = forwarded.to_str() {
            // Parser le header Forwarded (RFC 7239)
            for directive in forwarded_str.split(';') {
                if let Some(for_part) = directive.strip_prefix("for=") {
                    return for_part.trim_matches('"').to_string();
                }
            }
        }
    }
    
    // Fallback
    "unknown".to_string()
}

fn is_suspicious_request(
    _method: &Method,
    uri: &Uri,
    headers: &HeaderMap,
    status: StatusCode,
) -> bool {
    let path = uri.path();
    let query = uri.query().unwrap_or("");
    
    // Détecter les tentatives d'injection SQL
    if contains_injection_patterns(path) || contains_injection_patterns(query) {
        return true;
    }
    
    // Détecter les tentatives de traversée de répertoire
    if contains_dangerous_patterns(path) {
        return true;
    }
    
    // Détecter les tentatives de scan de ports ou de vulnérabilités
    if path.contains("/.well-known/") || 
       path.contains("/admin") || 
       path.contains("/wp-admin") ||
       path.contains("/phpmyadmin") {
        return true;
    }
    
    // Détecter les User-Agents suspects
    if let Some(user_agent) = headers.get("user-agent") {
        if let Ok(ua_str) = user_agent.to_str() {
            let ua_lower = ua_str.to_lowercase();
            if ua_lower.contains("bot") && 
               !ua_lower.contains("googlebot") && 
               !ua_lower.contains("bingbot") {
                return true;
            }
        }
    }
    
    // Détecter les erreurs 404 répétées sur des chemins sensibles
    if status == StatusCode::NOT_FOUND && (
        path.contains("admin") || 
        path.contains("config") || 
        path.contains("backup")
    ) {
        return true;
    }
    
    false
}

fn contains_dangerous_patterns(input: &str) -> bool {
    let dangerous_patterns = [
        "../", "..\\", "..%2f", "..%5c",
        "%2e%2e%2f", "%2e%2e%5c",
        "etc/passwd", "windows/system32",
        "/proc/", "/sys/",
    ];
    
    let input_lower = input.to_lowercase();
    dangerous_patterns.iter().any(|&pattern| input_lower.contains(pattern))
}

fn contains_injection_patterns(input: &str) -> bool {
    let injection_patterns = [
        "union select", "drop table", "insert into",
        "delete from", "update set", "create table",
        "<script", "javascript:", "onload=",
        "onerror=", "eval(", "alert(",
        "document.cookie", "window.location",
    ];
    
    let input_lower = input.to_lowercase();
    injection_patterns.iter().any(|&pattern| input_lower.contains(pattern))
}

#[cfg(test)]
mod tests {
    use super::*;
    use axum::http::{Method, Uri};

    #[test]
    fn test_extract_client_ip() {
        let mut headers = HeaderMap::new();
        
        // Test avec X-Forwarded-For
        headers.insert("x-forwarded-for", "192.168.1.1, 10.0.0.1".parse().unwrap());
        assert_eq!(extract_client_ip(&headers), "192.168.1.1");

        // Test avec X-Real-IP
        headers.clear();
        headers.insert("x-real-ip", "203.0.113.1".parse().unwrap());
        assert_eq!(extract_client_ip(&headers), "203.0.113.1");

        // Test sans headers
        headers.clear();
        assert_eq!(extract_client_ip(&headers), "unknown");
    }

    #[test]
    fn test_suspicious_request_detection() {
        let headers = HeaderMap::new();
        
        // Test avec méthode suspecte
        let method = Method::POST;
        let uri: Uri = "/stream/test".parse().unwrap();
        assert!(is_suspicious_request(&method, &uri, &headers, StatusCode::OK));

        // Test avec chemin suspect
        let method = Method::GET;
        let uri: Uri = "/admin/login".parse().unwrap();
        assert!(is_suspicious_request(&method, &uri, &headers, StatusCode::OK));

        // Test avec requête normale
        let method = Method::GET;
        let uri: Uri = "/stream/music.mp3".parse().unwrap();
        assert!(!is_suspicious_request(&method, &uri, &headers, StatusCode::OK));
    }
} 