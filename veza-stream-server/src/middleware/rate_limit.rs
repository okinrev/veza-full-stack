use axum::{
    extract::{Request, State},
    http::{HeaderMap, StatusCode},
    middleware::Next,
    response::Response,
};
use std::time::Instant;
use tracing::{warn, debug};
use crate::AppState;

pub async fn rate_limit_middleware(
    State(state): State<AppState>,
    request: Request,
    next: Next,
) -> Result<Response, StatusCode> {
    let headers = request.headers();
    let client_ip = extract_client_ip(headers);
    
    // Vérifier les limites de taux
    if !check_rate_limit(&state, &client_ip).await {
        warn!(
            client_ip = %client_ip,
            "Rate limit dépassé"
        );
        
        state.metrics.increment_rate_limited();
        
        return Err(StatusCode::TOO_MANY_REQUESTS);
    }
    
    // Enregistrer la requête
    record_request(&state, &client_ip).await;
    
    let response = next.run(request).await;
    Ok(response)
}

fn extract_client_ip(headers: &HeaderMap) -> String {
    // Vérifier les headers de proxy dans l'ordre de priorité
    if let Some(forwarded_for) = headers.get("x-forwarded-for") {
        if let Ok(forwarded_str) = forwarded_for.to_str() {
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
    
    "unknown".to_string()
}

async fn check_rate_limit(state: &AppState, client_ip: &str) -> bool {
    // Implémentation basique du rate limiting
    // Dans une vraie application, on utiliserait un store externe comme Redis
    
    let max_requests_per_minute = state.config.security.rate_limit_requests_per_minute;
    let _now = Instant::now();
    
    // Pour cette implémentation basique, on permet toutes les requêtes
    // En production, il faudrait implémenter un vrai système de rate limiting
    debug!(
        client_ip = %client_ip,
        limit = max_requests_per_minute,
        "Vérification du rate limit"
    );
    
    true
}

async fn record_request(_state: &AppState, client_ip: &str) {
    // Enregistrer la requête pour les statistiques
    debug!(
        client_ip = %client_ip,
        "Requête enregistrée"
    );
} 