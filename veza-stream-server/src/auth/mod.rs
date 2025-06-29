use std::collections::HashMap;
use std::sync::Arc;
use std::time::{Duration, SystemTime, UNIX_EPOCH};
use axum::{
    extract::{Request, State},
    http::{HeaderMap, HeaderValue, StatusCode},
    middleware::Next,
    response::Response,
    Json,
};
use serde::{Deserialize, Serialize};
use jsonwebtoken::{decode, encode, Algorithm, DecodingKey, EncodingKey, Header, Validation};
use tracing::{debug, error, warn};
use crate::config::Config;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Claims {
    pub sub: String,        // Subject (user ID)
    pub username: String,   // Username
    pub email: Option<String>,
    pub roles: Vec<Role>,
    pub permissions: Vec<Permission>,
    pub exp: u64,          // Expiration time
    pub iat: u64,          // Issued at
    pub iss: String,       // Issuer
    pub aud: String,       // Audience
    pub session_id: String, // Session ID pour la révocation
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum Role {
    Admin,
    Moderator,
    User,
    Premium,
    Artist,
    Guest,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum Permission {
    // Permissions de streaming
    StreamAudio,
    StreamHighQuality,
    StreamUnlimited,
    
    // Permissions de contenu
    UploadAudio,
    DeleteAudio,
    ModifyMetadata,
    
    // Permissions administratives
    ViewAnalytics,
    ManageUsers,
    SystemAdmin,
    
    // Permissions sociales
    CreatePlaylists,
    ShareContent,
    Comment,
    Like,
    
    // Permissions avancées
    AccessAPI,
    ManageSubscriptions,
    ViewReports,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct LoginRequest {
    pub username: String,
    pub password: String,
    pub remember_me: Option<bool>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct LoginResponse {
    pub access_token: String,
    pub refresh_token: String,
    pub token_type: String,
    pub expires_in: u64,
    pub user_info: UserInfo,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct UserInfo {
    pub id: String,
    pub username: String,
    pub email: Option<String>,
    pub roles: Vec<Role>,
    pub permissions: Vec<Permission>,
    pub subscription_tier: SubscriptionTier,
    pub created_at: u64,
    pub last_login: u64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum SubscriptionTier {
    Free,
    Premium,
    Artist,
    Enterprise,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RefreshTokenRequest {
    pub refresh_token: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TokenValidationResult {
    pub valid: bool,
    pub claims: Option<Claims>,
    pub error: Option<String>,
}

pub struct AuthManager {
    config: Arc<Config>,
    encoding_key: EncodingKey,
    decoding_key: DecodingKey,
    validation: Validation,
    revoked_tokens: Arc<tokio::sync::RwLock<HashMap<String, u64>>>, // session_id -> revocation_time
}

impl AuthManager {
    pub fn new(config: Arc<Config>) -> Result<Self, AuthError> {
        let jwt_secret = config.security.jwt_secret
            .as_ref()
            .ok_or(AuthError::ConfigurationError("JWT_SECRET not configured".to_string()))?;

        let encoding_key = EncodingKey::from_secret(jwt_secret.as_bytes());
        let decoding_key = DecodingKey::from_secret(jwt_secret.as_bytes());
        
        let mut validation = Validation::new(Algorithm::HS256);
        validation.set_audience(&["veza-services"]);
        validation.set_issuer(&["veza-platform"]);

        Ok(Self {
            config,
            encoding_key,
            decoding_key,
            validation,
            revoked_tokens: Arc::new(tokio::sync::RwLock::new(HashMap::new())),
        })
    }

    pub async fn authenticate_user(&self, username: &str, password: &str) -> Result<UserInfo, AuthError> {
        // Simuler une authentification (à remplacer par votre logique réelle)
        if username == "admin" && password == "admin123" {
            Ok(UserInfo {
                id: "admin_001".to_string(),
                username: username.to_string(),
                email: Some("admin@example.com".to_string()),
                roles: vec![Role::Admin],
                permissions: vec![
                    Permission::StreamAudio,
                    Permission::StreamHighQuality,
                    Permission::StreamUnlimited,
                    Permission::ViewAnalytics,
                    Permission::ManageUsers,
                    Permission::SystemAdmin,
                    Permission::AccessAPI,
                ],
                subscription_tier: SubscriptionTier::Enterprise,
                created_at: SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_secs(),
                last_login: SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_secs(),
            })
        } else if username == "user" && password == "user123" {
            Ok(UserInfo {
                id: "user_001".to_string(),
                username: username.to_string(),
                email: Some("user@example.com".to_string()),
                roles: vec![Role::User],
                permissions: vec![
                    Permission::StreamAudio,
                    Permission::CreatePlaylists,
                    Permission::ShareContent,
                    Permission::Comment,
                    Permission::Like,
                ],
                subscription_tier: SubscriptionTier::Free,
                created_at: SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_secs() - 86400,
                last_login: SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_secs(),
            })
        } else {
            Err(AuthError::InvalidCredentials)
        }
    }

    pub async fn generate_tokens(&self, user_info: &UserInfo, remember_me: bool) -> Result<(String, String), AuthError> {
        let now = SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_secs();
        let session_id = uuid::Uuid::new_v4().to_string();
        
        // Durée d'expiration selon remember_me
        let expires_in = if remember_me {
            self.config.security.jwt_expiration.as_secs() * 7 // 7 fois plus long
        } else {
            self.config.security.jwt_expiration.as_secs()
        };

        let claims = Claims {
            sub: user_info.id.clone(),
            username: user_info.username.clone(),
            email: user_info.email.clone(),
            roles: user_info.roles.clone(),
            permissions: user_info.permissions.clone(),
            exp: now + expires_in,
            iat: now,
            iss: "stream_server".to_string(),
            aud: "stream_server".to_string(),
            session_id: session_id.clone(),
        };

        let access_token = encode(&Header::default(), &claims, &self.encoding_key)
            .map_err(|e| AuthError::TokenGenerationError(e.to_string()))?;

        // Refresh token avec une durée plus longue
        let refresh_claims = Claims {
            exp: now + (expires_in * 2), // 2x plus long que l'access token
            ..claims.clone()
        };

        let refresh_token = encode(&Header::default(), &refresh_claims, &self.encoding_key)
            .map_err(|e| AuthError::TokenGenerationError(e.to_string()))?;

        Ok((access_token, refresh_token))
    }

    pub async fn validate_token(&self, token: &str) -> TokenValidationResult {
        match decode::<Claims>(token, &self.decoding_key, &self.validation) {
            Ok(token_data) => {
                let claims = token_data.claims;
                
                // Vérifier si le token est révoqué
                let revoked_tokens = self.revoked_tokens.read().await;
                if revoked_tokens.contains_key(&claims.session_id) {
                    return TokenValidationResult {
                        valid: false,
                        claims: None,
                        error: Some("Token has been revoked".to_string()),
                    };
                }

                // Vérifier l'expiration
                let now = SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_secs();
                if claims.exp < now {
                    return TokenValidationResult {
                        valid: false,
                        claims: None,
                        error: Some("Token has expired".to_string()),
                    };
                }

                TokenValidationResult {
                    valid: true,
                    claims: Some(claims),
                    error: None,
                }
            }
            Err(e) => TokenValidationResult {
                valid: false,
                claims: None,
                error: Some(e.to_string()),
            }
        }
    }

    pub async fn refresh_token(&self, refresh_token: &str) -> Result<(String, String), AuthError> {
        let validation_result = self.validate_token(refresh_token).await;
        
        if !validation_result.valid {
            return Err(AuthError::InvalidToken(
                validation_result.error.unwrap_or("Invalid refresh token".to_string())
            ));
        }

        let claims = validation_result.claims.unwrap();
        
        // Créer un nouveau UserInfo à partir des claims
        let user_info = UserInfo {
            id: claims.sub,
            username: claims.username,
            email: claims.email,
            roles: claims.roles,
            permissions: claims.permissions,
            subscription_tier: SubscriptionTier::Free, // À déterminer selon la logique métier
            created_at: claims.iat,
            last_login: SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_secs(),
        };

        // Révoquer l'ancien token
        self.revoke_token(&claims.session_id).await;

        // Générer de nouveaux tokens
        self.generate_tokens(&user_info, false).await
    }

    pub async fn revoke_token(&self, session_id: &str) {
        let mut revoked_tokens = self.revoked_tokens.write().await;
        let now = SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_secs();
        revoked_tokens.insert(session_id.to_string(), now);
        
        // Nettoyer les anciens tokens révoqués (plus de 24h)
        let cutoff = now - (24 * 3600);
        revoked_tokens.retain(|_, &mut revocation_time| revocation_time > cutoff);
    }

    pub fn has_permission(&self, claims: &Claims, required_permission: Permission) -> bool {
        claims.permissions.contains(&required_permission)
    }

    pub fn has_role(&self, claims: &Claims, required_role: Role) -> bool {
        claims.roles.contains(&required_role)
    }

    pub fn has_any_role(&self, claims: &Claims, required_roles: &[Role]) -> bool {
        claims.roles.iter().any(|role| required_roles.contains(role))
    }

    pub async fn login(&self, request: LoginRequest) -> Result<LoginResponse, AuthError> {
        let user_info = self.authenticate_user(&request.username, &request.password).await?;
        let (access_token, refresh_token) = self.generate_tokens(&user_info, request.remember_me.unwrap_or(false)).await?;
        
        Ok(LoginResponse {
            access_token,
            refresh_token,
            token_type: "Bearer".to_string(),
            expires_in: self.config.security.jwt_expiration.as_secs(),
            user_info,
        })
    }
}

impl Clone for AuthManager {
    fn clone(&self) -> Self {
        Self {
            config: self.config.clone(),
            encoding_key: self.encoding_key.clone(),
            decoding_key: self.decoding_key.clone(),
            validation: self.validation.clone(),
            revoked_tokens: self.revoked_tokens.clone(),
        }
    }
}

// Middleware d'authentification
pub async fn auth_middleware(
    State(auth_manager): State<Arc<AuthManager>>,
    mut request: Request,
    next: Next,
) -> Result<Response, StatusCode> {
    let headers = request.headers();
    
    let token = extract_token_from_headers(headers)
        .ok_or(StatusCode::UNAUTHORIZED)?;

    let validation_result = auth_manager.validate_token(&token).await;
    
    if !validation_result.valid {
        warn!("Token validation failed: {:?}", validation_result.error);
        return Err(StatusCode::UNAUTHORIZED);
    }

    let claims = validation_result.claims.unwrap();
    
    // Ajouter les claims à la requête pour les handlers suivants
    request.extensions_mut().insert(claims);
    
    Ok(next.run(request).await)
}

// Middleware de vérification des permissions
pub fn require_permission(required_permission: Permission) -> impl Fn(Request, Next) -> std::pin::Pin<Box<dyn std::future::Future<Output = Result<Response, StatusCode>> + Send>> + Clone {
    move |request: Request, next: Next| {
        let required_permission = required_permission.clone();
        Box::pin(async move {
            let claims = request.extensions().get::<Claims>()
                .ok_or(StatusCode::UNAUTHORIZED)?;

            if !claims.permissions.contains(&required_permission) {
                warn!("User {} lacks required permission: {:?}", claims.username, required_permission);
                return Err(StatusCode::FORBIDDEN);
            }

            Ok(next.run(request).await)
        })
    }
}

// Middleware de vérification des rôles
pub fn require_role(required_role: Role) -> impl Fn(Request, Next) -> std::pin::Pin<Box<dyn std::future::Future<Output = Result<Response, StatusCode>> + Send>> + Clone {
    move |request: Request, next: Next| {
        let required_role = required_role.clone();
        Box::pin(async move {
            let claims = request.extensions().get::<Claims>()
                .ok_or(StatusCode::UNAUTHORIZED)?;

            if !claims.roles.contains(&required_role) {
                warn!("User {} lacks required role: {:?}", claims.username, required_role);
                return Err(StatusCode::FORBIDDEN);
            }

            Ok(next.run(request).await)
        })
    }
}

fn extract_token_from_headers(headers: &HeaderMap) -> Option<String> {
    let auth_header = headers.get("Authorization")?;
    let auth_str = auth_header.to_str().ok()?;
    
    if auth_str.starts_with("Bearer ") {
        Some(auth_str[7..].to_string())
    } else {
        None
    }
}

// Handlers pour les routes d'authentification
pub async fn login_handler(
    State(auth_manager): State<Arc<AuthManager>>,
    Json(request): Json<LoginRequest>,
) -> Result<Json<LoginResponse>, (StatusCode, String)> {
    match auth_manager.login(request).await {
        Ok(response) => Ok(Json(response)),
        Err(e) => {
            error!("Login failed: {:?}", e);
            match e {
                AuthError::InvalidCredentials => Err((StatusCode::UNAUTHORIZED, "Invalid credentials".to_string())),
                AuthError::TokenGenerationError(msg) => Err((StatusCode::INTERNAL_SERVER_ERROR, msg)),
                _ => Err((StatusCode::INTERNAL_SERVER_ERROR, "Authentication failed".to_string())),
            }
        }
    }
}

pub async fn refresh_handler(
    State(auth_manager): State<Arc<AuthManager>>,
    Json(request): Json<RefreshTokenRequest>,
) -> Result<Json<LoginResponse>, (StatusCode, String)> {
    match auth_manager.refresh_token(&request.refresh_token).await {
        Ok((access_token, refresh_token)) => {
            // Simuler la récupération des infos utilisateur
            let user_info = UserInfo {
                id: "refreshed_user".to_string(),
                username: "user".to_string(),
                email: None,
                roles: vec![Role::User],
                permissions: vec![Permission::StreamAudio],
                subscription_tier: SubscriptionTier::Free,
                created_at: 0,
                last_login: SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_secs(),
            };

            let response = LoginResponse {
                access_token,
                refresh_token,
                token_type: "Bearer".to_string(),
                expires_in: auth_manager.config.security.jwt_expiration.as_secs(),
                user_info,
            };

            Ok(Json(response))
        }
        Err(e) => {
            error!("Token refresh failed: {:?}", e);
            Err((StatusCode::UNAUTHORIZED, "Token refresh failed".to_string()))
        }
    }
}

pub async fn logout_handler(
    State(auth_manager): State<Arc<AuthManager>>,
    request: Request,
) -> Result<Json<serde_json::Value>, (StatusCode, String)> {
    if let Some(claims) = request.extensions().get::<Claims>() {
        auth_manager.revoke_token(&claims.session_id).await;
        debug!("User {} logged out", claims.username);
    }

    Ok(Json(serde_json::json!({
        "message": "Successfully logged out"
    })))
}

pub async fn user_info_handler(
    request: Request,
) -> Result<Json<Claims>, StatusCode> {
    let claims = request.extensions().get::<Claims>()
        .ok_or(StatusCode::UNAUTHORIZED)?;

    Ok(Json(claims.clone()))
}

#[derive(Debug, thiserror::Error)]
pub enum AuthError {
    #[error("Invalid credentials")]
    InvalidCredentials,
    
    #[error("Invalid token: {0}")]
    InvalidToken(String),
    
    #[error("Token generation error: {0}")]
    TokenGenerationError(String),
    
    #[error("Configuration error: {0}")]
    ConfigurationError(String),
    
    #[error("Database error: {0}")]
    DatabaseError(String),
} 