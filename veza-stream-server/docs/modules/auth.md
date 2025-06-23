# Authentication Module Documentation

Le module d'authentification fournit un système complet de gestion des utilisateurs, des rôles, des permissions et des tokens JWT pour le serveur de streaming.

## Table des Matières

- [Vue d'ensemble](#vue-densemble)
- [Architecture de sécurité](#architecture-de-sécurité)
- [Types et Structures](#types-et-structures)
- [Système de rôles](#système-de-rôles)
- [Système de permissions](#système-de-permissions)
- [Gestion des tokens](#gestion-des-tokens)
- [AuthManager](#authmanager)
- [Middleware](#middleware)
- [API Reference](#api-reference)
- [Exemples d'utilisation](#exemples-dutilisation)
- [Intégration](#intégration)

## Vue d'ensemble

Le système d'authentification comprend :
- **Authentification JWT** avec access/refresh tokens
- **Système de rôles** hiérarchique et flexible
- **Permissions granulaires** pour un contrôle d'accès fin
- **Révocation de tokens** avec sessions
- **Tiers d'abonnement** intégrés
- **Middleware de sécurité** pour la protection des routes
- **Gestion des sessions** avec nettoyage automatique

## Architecture de sécurité

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Client        │    │   AuthManager    │    │   Database      │
│                 │    │                  │    │                 │
│ - Login Request │───►│ - Verify Creds   │───►│ - User Storage  │
│ - Store Tokens  │◄───│ - Generate JWT   │◄───│ - Session Mgmt  │
│ - API Requests  │    │ - Validate Tokens│    │ - Permissions   │
└─────────────────┘    └──────────────────┘    └─────────────────┘
         │                        │                        
         └────────────────────────┼────────────────────────
                                  ▼
                    ┌─────────────────────────┐
                    │     Middleware          │
                    │                         │
                    │ - Token Validation      │
                    │ - Permission Checks     │
                    │ - Rate Limiting         │
                    │ - Security Headers      │
                    └─────────────────────────┘
```

## Types et Structures

### Claims (JWT Payload)

```rust
pub struct Claims {
    pub sub: String,                    // Subject (user ID)
    pub username: String,               // Nom d'utilisateur
    pub email: Option<String>,          // Email (optionnel)
    pub roles: Vec<Role>,              // Rôles de l'utilisateur
    pub permissions: Vec<Permission>,   // Permissions explicites
    pub exp: u64,                      // Expiration timestamp
    pub iat: u64,                      // Issued at timestamp
    pub iss: String,                   // Issuer (nom du service)
    pub aud: String,                   // Audience (destinataire)
    pub session_id: String,            // ID de session pour révocation
}
```

**Champs détaillés :**
- `sub` : Identifiant unique de l'utilisateur
- `username` : Nom d'utilisateur pour affichage
- `email` : Adresse email (optionnel pour les comptes invités)
- `roles` : Liste des rôles attribués
- `permissions` : Permissions explicites supplémentaires
- `exp` : Timestamp d'expiration Unix
- `iat` : Timestamp de création Unix
- `iss` : "stream-server" (identifiant du service)
- `aud` : "stream-api" (audience ciblée)
- `session_id` : UUID pour identifier et révoquer la session

### Role (Rôles système)

```rust
pub enum Role {
    Admin,       // Administrateur système complet
    Moderator,   // Modérateur de contenu
    User,        // Utilisateur standard
    Premium,     // Utilisateur premium
    Artist,      // Artiste/Créateur de contenu
    Guest,       // Invité (accès limité)
}
```

**Hiérarchie des rôles :**
```
Admin
├── Moderator
│   ├── Artist
│   └── Premium
│       └── User
│           └── Guest
```

### Permission (Permissions granulaires)

```rust
pub enum Permission {
    // Permissions de streaming
    StreamAudio,           // Écouter de l'audio
    StreamHighQuality,     // Streaming haute qualité
    StreamUnlimited,       // Streaming illimité
    
    // Permissions de contenu
    UploadAudio,          // Uploader des fichiers audio
    DeleteAudio,          // Supprimer des fichiers
    ModifyMetadata,       // Modifier les métadonnées
    
    // Permissions administratives
    ViewAnalytics,        // Consulter les analytics
    ManageUsers,          // Gérer les utilisateurs
    SystemAdmin,          // Administration système
    
    // Permissions sociales
    CreatePlaylists,      // Créer des playlists
    ShareContent,         // Partager du contenu
    Comment,              // Commenter
    Like,                 // Liker du contenu
    
    // Permissions avancées
    AccessAPI,            // Accès à l'API
    ManageSubscriptions,  // Gérer les abonnements
    ViewReports,          // Consulter les rapports
}
```

### LoginRequest

```rust
pub struct LoginRequest {
    pub username: String,               // Nom d'utilisateur ou email
    pub password: String,               // Mot de passe
    pub remember_me: Option<bool>,      // Session prolongée
}
```

### LoginResponse

```rust
pub struct LoginResponse {
    pub access_token: String,           // Token d'accès JWT
    pub refresh_token: String,          // Token de rafraîchissement
    pub token_type: String,            // "Bearer"
    pub expires_in: u64,               // Durée de validité (secondes)
    pub user_info: UserInfo,           // Informations utilisateur
}
```

### UserInfo

```rust
pub struct UserInfo {
    pub id: String,                            // ID utilisateur unique
    pub username: String,                      // Nom d'utilisateur
    pub email: Option<String>,                 // Email
    pub roles: Vec<Role>,                     // Rôles attribués
    pub permissions: Vec<Permission>,          // Permissions explicites
    pub subscription_tier: SubscriptionTier,  // Niveau d'abonnement
    pub created_at: u64,                      // Date de création
    pub last_login: u64,                      // Dernière connexion
}
```

### SubscriptionTier

```rust
pub enum SubscriptionTier {
    Free,         // Gratuit (limité)
    Premium,      // Premium (standard)
    Artist,       // Artiste/Créateur
    Enterprise,   // Entreprise
}
```

**Avantages par tier :**
- **Free** : Streaming limité, qualité standard, publicités
- **Premium** : Streaming illimité, haute qualité, sans pub
- **Artist** : Toutes les fonctionnalités Premium + upload, analytics
- **Enterprise** : API complète, analytics avancées, support prioritaire

## Système de rôles

### Attribution automatique des permissions

```rust
impl Role {
    pub fn default_permissions(&self) -> Vec<Permission> {
        match self {
            Role::Admin => vec![
                // Toutes les permissions
                Permission::StreamAudio,
                Permission::StreamHighQuality,
                Permission::StreamUnlimited,
                Permission::UploadAudio,
                Permission::DeleteAudio,
                Permission::ModifyMetadata,
                Permission::ViewAnalytics,
                Permission::ManageUsers,
                Permission::SystemAdmin,
                Permission::CreatePlaylists,
                Permission::ShareContent,
                Permission::Comment,
                Permission::Like,
                Permission::AccessAPI,
                Permission::ManageSubscriptions,
                Permission::ViewReports,
            ],
            
            Role::Moderator => vec![
                Permission::StreamAudio,
                Permission::StreamHighQuality,
                Permission::StreamUnlimited,
                Permission::DeleteAudio,
                Permission::ModifyMetadata,
                Permission::ViewAnalytics,
                Permission::CreatePlaylists,
                Permission::ShareContent,
                Permission::Comment,
                Permission::Like,
                Permission::ViewReports,
            ],
            
            Role::Artist => vec![
                Permission::StreamAudio,
                Permission::StreamHighQuality,
                Permission::StreamUnlimited,
                Permission::UploadAudio,
                Permission::ModifyMetadata,
                Permission::ViewAnalytics,
                Permission::CreatePlaylists,
                Permission::ShareContent,
                Permission::Comment,
                Permission::Like,
                Permission::AccessAPI,
            ],
            
            Role::Premium => vec![
                Permission::StreamAudio,
                Permission::StreamHighQuality,
                Permission::StreamUnlimited,
                Permission::CreatePlaylists,
                Permission::ShareContent,
                Permission::Comment,
                Permission::Like,
            ],
            
            Role::User => vec![
                Permission::StreamAudio,
                Permission::CreatePlaylists,
                Permission::ShareContent,
                Permission::Comment,
                Permission::Like,
            ],
            
            Role::Guest => vec![
                Permission::StreamAudio,
            ],
        }
    }
}
```

### Vérification hiérarchique

```rust
impl Role {
    pub fn has_role_level(&self, required_role: &Role) -> bool {
        match (self, required_role) {
            (Role::Admin, _) => true,
            (Role::Moderator, Role::Artist | Role::Premium | Role::User | Role::Guest) => true,
            (Role::Artist, Role::Premium | Role::User | Role::Guest) => true,
            (Role::Premium, Role::User | Role::Guest) => true,
            (Role::User, Role::Guest) => true,
            _ => self == required_role,
        }
    }
}
```

## Système de permissions

### Vérification de permissions

```rust
impl AuthManager {
    pub fn has_permission(&self, claims: &Claims, required_permission: Permission) -> bool {
        // Vérifier permissions explicites
        if claims.permissions.contains(&required_permission) {
            return true;
        }
        
        // Vérifier permissions des rôles
        for role in &claims.roles {
            if role.default_permissions().contains(&required_permission) {
                return true;
            }
        }
        
        false
    }
    
    pub fn has_any_role(&self, claims: &Claims, required_roles: &[Role]) -> bool {
        claims.roles.iter().any(|user_role| {
            required_roles.iter().any(|required_role| {
                user_role.has_role_level(required_role)
            })
        })
    }
}
```

## Gestion des tokens

### Structure des tokens

**Access Token (JWT) :**
- **Durée** : 15 minutes (configurable)
- **Usage** : Authentification des requêtes API
- **Contenu** : Claims complets avec permissions

**Refresh Token (JWT) :**
- **Durée** : 7 jours (configurable)
- **Usage** : Renouvellement des access tokens
- **Contenu** : Informations minimales (user_id, session_id)

### Cycle de vie des tokens

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│  Initial Login  │───►│  Token Usage     │───►│  Token Refresh  │
│                 │    │                  │    │                 │
│ - Validate creds│    │ - API requests   │    │ - Validate      │
│ - Generate JWT  │    │ - Auto-refresh   │    │   refresh token │
│ - Return tokens │    │ - Permission     │    │ - Generate new  │
│                 │    │   checks         │    │   access token  │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

### Révocation de tokens

```rust
impl AuthManager {
    pub async fn revoke_token(&self, session_id: &str) {
        let mut revoked_tokens = self.revoked_tokens.write().await;
        revoked_tokens.insert(
            session_id.to_string(), 
            SystemTime::now()
                .duration_since(UNIX_EPOCH)
                .unwrap()
                .as_secs()
        );
    }
    
    pub async fn is_token_revoked(&self, session_id: &str) -> bool {
        let revoked_tokens = self.revoked_tokens.read().await;
        revoked_tokens.contains_key(session_id)
    }
}
```

## AuthManager

### Structure principale

```rust
pub struct AuthManager {
    config: Arc<Config>,                                              // Configuration
    encoding_key: EncodingKey,                                       // Clé de signature JWT
    decoding_key: DecodingKey,                                       // Clé de vérification JWT
    validation: Validation,                                          // Paramètres de validation JWT
    revoked_tokens: Arc<tokio::sync::RwLock<HashMap<String, u64>>>, // Tokens révoqués
}
```

### Méthodes principales

#### Authentification

```rust
impl AuthManager {
    pub async fn authenticate_user(&self, username: &str, password: &str) -> Result<UserInfo, AuthError> {
        // 1. Rechercher l'utilisateur dans la base de données
        // 2. Vérifier le mot de passe avec bcrypt
        // 3. Retourner les informations utilisateur
        
        // Simulation d'authentification
        if username == "admin" && password == "admin123" {
            Ok(UserInfo {
                id: "admin_001".to_string(),
                username: "admin".to_string(),
                email: Some("admin@streamserver.com".to_string()),
                roles: vec![Role::Admin],
                permissions: vec![], // Les permissions viennent des rôles
                subscription_tier: SubscriptionTier::Enterprise,
                created_at: 1640995200, // 2022-01-01
                last_login: SystemTime::now()
                    .duration_since(UNIX_EPOCH)
                    .unwrap()
                    .as_secs(),
            })
        } else {
            Err(AuthError::InvalidCredentials)
        }
    }
}
```

#### Génération de tokens

```rust
impl AuthManager {
    pub async fn generate_tokens(&self, user_info: &UserInfo, remember_me: bool) -> Result<(String, String), AuthError> {
        let session_id = Uuid::new_v4().to_string();
        let now = Utc::now();
        
        // Access token (courte durée)
        let access_claims = Claims {
            sub: user_info.id.clone(),
            username: user_info.username.clone(),
            email: user_info.email.clone(),
            roles: user_info.roles.clone(),
            permissions: self.collect_all_permissions(&user_info.roles),
            exp: (now + Duration::minutes(15)).timestamp() as u64,
            iat: now.timestamp() as u64,
            iss: "stream-server".to_string(),
            aud: "stream-api".to_string(),
            session_id: session_id.clone(),
        };
        
        // Refresh token (longue durée)
        let refresh_duration = if remember_me { Duration::days(30) } else { Duration::days(7) };
        let refresh_claims = Claims {
            sub: user_info.id.clone(),
            username: user_info.username.clone(),
            email: None, // Pas d'infos sensibles dans le refresh token
            roles: vec![],
            permissions: vec![],
            exp: (now + refresh_duration).timestamp() as u64,
            iat: now.timestamp() as u64,
            iss: "stream-server".to_string(),
            aud: "stream-refresh".to_string(),
            session_id,
        };
        
        let access_token = encode(&Header::default(), &access_claims, &self.encoding_key)
            .map_err(|e| AuthError::TokenGenerationError(e.to_string()))?;
            
        let refresh_token = encode(&Header::default(), &refresh_claims, &self.encoding_key)
            .map_err(|e| AuthError::TokenGenerationError(e.to_string()))?;
        
        Ok((access_token, refresh_token))
    }
}
```

#### Validation de tokens

```rust
impl AuthManager {
    pub async fn validate_token(&self, token: &str) -> TokenValidationResult {
        match decode::<Claims>(token, &self.decoding_key, &self.validation) {
            Ok(token_data) => {
                let claims = token_data.claims;
                
                // Vérifier si le token est révoqué
                if self.is_token_revoked(&claims.session_id).await {
                    return TokenValidationResult {
                        valid: false,
                        claims: None,
                        error: Some("Token revoked".to_string()),
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
            },
        }
    }
}
```

## Middleware

### Middleware d'authentification

```rust
pub async fn auth_middleware(
    State(auth_manager): State<Arc<AuthManager>>,
    mut request: Request,
    next: Next,
) -> Result<Response, StatusCode> {
    // Extraire le token de l'en-tête Authorization
    let token = extract_token_from_headers(request.headers())
        .ok_or(StatusCode::UNAUTHORIZED)?;
    
    // Valider le token
    let validation_result = auth_manager.validate_token(&token).await;
    
    if !validation_result.valid {
        return Err(StatusCode::UNAUTHORIZED);
    }
    
    // Ajouter les claims à la requête pour les handlers suivants
    if let Some(claims) = validation_result.claims {
        request.extensions_mut().insert(claims);
    }
    
    Ok(next.run(request).await)
}
```

### Middleware de permissions

```rust
pub fn require_permission(required_permission: Permission) -> impl Fn(Request, Next) -> Pin<Box<dyn Future<Output = Result<Response, StatusCode>> + Send>> + Clone {
    move |request: Request, next: Next| {
        let required_permission = required_permission.clone();
        Box::pin(async move {
            // Récupérer les claims depuis les extensions de la requête
            let claims = request.extensions()
                .get::<Claims>()
                .ok_or(StatusCode::UNAUTHORIZED)?;
            
            // Vérifier la permission
            let auth_manager = request.extensions()
                .get::<Arc<AuthManager>>()
                .ok_or(StatusCode::INTERNAL_SERVER_ERROR)?;
                
            if !auth_manager.has_permission(claims, required_permission) {
                return Err(StatusCode::FORBIDDEN);
            }
            
            Ok(next.run(request).await)
        })
    }
}
```

### Middleware de rôles

```rust
pub fn require_role(required_role: Role) -> impl Fn(Request, Next) -> Pin<Box<dyn Future<Output = Result<Response, StatusCode>> + Send>> + Clone {
    move |request: Request, next: Next| {
        let required_role = required_role.clone();
        Box::pin(async move {
            let claims = request.extensions()
                .get::<Claims>()
                .ok_or(StatusCode::UNAUTHORIZED)?;
            
            let auth_manager = request.extensions()
                .get::<Arc<AuthManager>>()
                .ok_or(StatusCode::INTERNAL_SERVER_ERROR)?;
                
            if !auth_manager.has_any_role(claims, &[required_role]) {
                return Err(StatusCode::FORBIDDEN);
            }
            
            Ok(next.run(request).await)
        })
    }
}
```

## API Reference

### AuthManager Methods

#### `new(config: Arc<Config>) -> Result<Self, AuthError>`
Crée une nouvelle instance d'AuthManager avec configuration JWT.

#### `authenticate_user(username: &str, password: &str) -> Result<UserInfo, AuthError>`
Authentifie un utilisateur avec ses identifiants.

#### `generate_tokens(user_info: &UserInfo, remember_me: bool) -> Result<(String, String), AuthError>`
Génère une paire access/refresh token pour un utilisateur.

#### `validate_token(token: &str) -> TokenValidationResult`
Valide un token JWT et retourne les claims si valide.

#### `refresh_token(refresh_token: &str) -> Result<(String, String), AuthError>`
Génère un nouveau access token à partir d'un refresh token valide.

#### `revoke_token(session_id: &str)`
Révoque une session en ajoutant son ID à la liste noire.

#### `has_permission(claims: &Claims, required_permission: Permission) -> bool`
Vérifie si un utilisateur a une permission spécifique.

#### `has_role(claims: &Claims, required_role: Role) -> bool`
Vérifie si un utilisateur a un rôle spécifique.

#### `login(request: LoginRequest) -> Result<LoginResponse, AuthError>`
Processus de connexion complet (authentification + génération de tokens).

### Handlers API

#### `login_handler(State, Json<LoginRequest>) -> Result<Json<LoginResponse>, (StatusCode, String)>`
Endpoint de connexion POST /api/auth/login.

#### `refresh_handler(State, Json<RefreshTokenRequest>) -> Result<Json<LoginResponse>, (StatusCode, String)>`
Endpoint de rafraîchissement POST /api/auth/refresh.

#### `logout_handler(State, Request) -> Result<Json<serde_json::Value>, (StatusCode, String)>`
Endpoint de déconnexion POST /api/auth/logout.

#### `user_info_handler(Request) -> Result<Json<Claims>, StatusCode>`
Endpoint d'informations utilisateur GET /api/auth/me.

## Exemples d'utilisation

### Connexion d'un utilisateur

```rust
use stream_server::auth::{AuthManager, LoginRequest};

async fn example_login() -> Result<(), Box<dyn std::error::Error>> {
    let config = Arc::new(Config::from_env()?);
    let auth_manager = AuthManager::new(config)?;
    
    // Tentative de connexion
    let login_request = LoginRequest {
        username: "alice_musician".to_string(),
        password: "secure_password123".to_string(),
        remember_me: Some(true),
    };
    
    match auth_manager.login(login_request).await {
        Ok(response) => {
            println!("✅ Connexion réussie!");
            println!("🆔 User: {}", response.user_info.username);
            println!("🎭 Rôles: {:?}", response.user_info.roles);
            println!("💎 Abonnement: {:?}", response.user_info.subscription_tier);
            println!("🔑 Access Token: {}...", &response.access_token[..20]);
            println!("♻️  Refresh Token: {}...", &response.refresh_token[..20]);
            println!("⏱️  Expire dans: {}s", response.expires_in);
        }
        Err(e) => {
            println!("❌ Échec de connexion: {}", e);
        }
    }
    
    Ok(())
}
```

### Validation et rafraîchissement de tokens

```rust
async fn example_token_lifecycle() -> Result<(), Box<dyn std::error::Error>> {
    let config = Arc::new(Config::from_env()?);
    let auth_manager = AuthManager::new(config)?;
    
    // Supposons que nous avons des tokens depuis une connexion précédente
    let access_token = "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9...";
    let refresh_token = "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9...";
    
    // Valider l'access token
    let validation_result = auth_manager.validate_token(access_token).await;
    
    if validation_result.valid {
        if let Some(claims) = validation_result.claims {
            println!("✅ Token valide pour: {}", claims.username);
            println!("🛡️  Permissions: {:?}", claims.permissions);
        }
    } else {
        println!("❌ Token invalide: {:?}", validation_result.error);
        
        // Essayer de rafraîchir le token
        match auth_manager.refresh_token(refresh_token).await {
            Ok((new_access_token, new_refresh_token)) => {
                println!("🔄 Tokens rafraîchis avec succès");
                println!("🔑 Nouveau access token: {}...", &new_access_token[..20]);
            }
            Err(e) => {
                println!("❌ Échec du rafraîchissement: {}", e);
                // L'utilisateur doit se reconnecter
            }
        }
    }
    
    Ok(())
}
```

### Système de permissions avancé

```rust
use stream_server::auth::{Permission, Role};

async fn example_permission_system() -> Result<(), Box<dyn std::error::Error>> {
    let config = Arc::new(Config::from_env()?);
    let auth_manager = AuthManager::new(config)?;
    
    // Simuler des claims utilisateur
    let artist_claims = Claims {
        sub: "artist_123".to_string(),
        username: "DJ_Producer".to_string(),
        email: Some("dj@example.com".to_string()),
        roles: vec![Role::Artist],
        permissions: vec![Permission::AccessAPI], // Permission supplémentaire
        exp: (Utc::now() + Duration::hours(1)).timestamp() as u64,
        iat: Utc::now().timestamp() as u64,
        iss: "stream-server".to_string(),
        aud: "stream-api".to_string(),
        session_id: "session_789".to_string(),
    };
    
    // Tests de permissions
    let permissions_to_test = vec![
        Permission::StreamAudio,
        Permission::UploadAudio,
        Permission::ViewAnalytics,
        Permission::ManageUsers,
        Permission::SystemAdmin,
    ];
    
    println!("🎭 Utilisateur: {} (rôles: {:?})", artist_claims.username, artist_claims.roles);
    println!("🔍 Test des permissions:");
    
    for permission in permissions_to_test {
        let has_permission = auth_manager.has_permission(&artist_claims, permission.clone());
        let status = if has_permission { "✅" } else { "❌" };
        println!("  {} {:?}", status, permission);
    }
    
    // Test de rôles hiérarchiques
    let roles_to_test = vec![
        Role::Guest,
        Role::User,
        Role::Premium,
        Role::Artist,
        Role::Moderator,
        Role::Admin,
    ];
    
    println!("\n🏆 Test des rôles:");
    for role in roles_to_test {
        let has_role = auth_manager.has_any_role(&artist_claims, &[role.clone()]);
        let status = if has_role { "✅" } else { "❌" };
        println!("  {} {:?}", status, role);
    }
    
    Ok(())
}
```

### Utilisation des middleware

```rust
use axum::{Router, routing::get, middleware::from_fn_with_state};
use stream_server::auth::{auth_middleware, require_permission, require_role};

fn create_protected_routes(auth_manager: Arc<AuthManager>) -> Router {
    Router::new()
        // Route publique - pas d'authentification requise
        .route("/public", get(public_handler))
        
        // Routes authentifiées - token JWT requis
        .route("/profile", get(profile_handler))
        .route("/playlists", get(playlists_handler))
        .layer(from_fn_with_state(auth_manager.clone(), auth_middleware))
        
        // Routes avec permissions spécifiques
        .route("/upload", post(upload_handler))
        .layer(from_fn(require_permission(Permission::UploadAudio)))
        
        // Routes avec rôles spécifiques
        .route("/admin", get(admin_handler))
        .layer(from_fn(require_role(Role::Admin)))
        
        // Analytics - permission spéciale
        .route("/analytics", get(analytics_handler))
        .layer(from_fn(require_permission(Permission::ViewAnalytics)))
        
        .with_state(auth_manager)
}

// Handlers d'exemple
async fn public_handler() -> &'static str {
    "Public content - no auth required"
}

async fn profile_handler(claims: Extension<Claims>) -> Json<Claims> {
    Json(claims.0)
}

async fn upload_handler(claims: Extension<Claims>) -> String {
    format!("Upload authorized for user: {}", claims.username)
}

async fn admin_handler(claims: Extension<Claims>) -> String {
    format!("Admin access granted to: {}", claims.username)
}

async fn analytics_handler(claims: Extension<Claims>) -> String {
    format!("Analytics access for: {} (roles: {:?})", claims.username, claims.roles)
}
```

### Intégration avec WebSocket

```rust
use stream_server::auth::AuthManager;
use axum::extract::ws::{WebSocketUpgrade, WebSocket};

async fn websocket_auth_handler(
    ws: WebSocketUpgrade,
    Query(params): Query<HashMap<String, String>>,
    State(auth_manager): State<Arc<AuthManager>>,
) -> Response {
    // Vérifier le token dans les paramètres de query
    let token = params.get("token")
        .ok_or_else(|| "Token required".to_string());
    
    match token {
        Ok(token) => {
            let validation_result = auth_manager.validate_token(token).await;
            
            if validation_result.valid {
                ws.on_upgrade(move |socket| handle_authenticated_socket(socket, validation_result.claims))
            } else {
                // Réponse d'erreur HTTP
                Response::builder()
                    .status(StatusCode::UNAUTHORIZED)
                    .body("Invalid token".into())
                    .unwrap()
            }
        }
        Err(e) => Response::builder()
            .status(StatusCode::BAD_REQUEST)
            .body(e.into())
            .unwrap()
    }
}

async fn handle_authenticated_socket(socket: WebSocket, claims: Option<Claims>) {
    if let Some(claims) = claims {
        println!("🔌 WebSocket connecté pour: {}", claims.username);
        // Logique WebSocket avec utilisateur authentifié
    }
}
```

## Intégration

### Avec le serveur principal

```rust
// Dans main.rs
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let config = Arc::new(Config::from_env()?);
    
    // Créer le gestionnaire d'authentification
    let auth_manager = Arc::new(AuthManager::new(config.clone())?);
    
    // Router avec authentification
    let app = Router::new()
        // Routes publiques
        .route("/", get(|| async { "Stream Server API" }))
        .route("/health", get(health_check))
        
        // Routes d'authentification
        .route("/api/auth/login", post(login_handler))
        .route("/api/auth/refresh", post(refresh_handler))
        .route("/api/auth/logout", post(logout_handler))
        
        // Routes protégées
        .route("/api/auth/me", get(user_info_handler))
        .route("/api/profile", get(profile_handler))
        .layer(from_fn_with_state(auth_manager.clone(), auth_middleware))
        
        // Streaming avec authentification
        .route("/stream/:filename", get(stream_with_auth))
        .layer(from_fn_with_state(auth_manager.clone(), auth_middleware))
        .layer(from_fn(require_permission(Permission::StreamAudio)))
        
        .with_state(auth_manager);
    
    // Démarrer le serveur
    let listener = tokio::net::TcpListener::bind("0.0.0.0:8082").await?;
    axum::serve(listener, app).await?;
    
    Ok(())
}
```

### Avec l'API Go

```go
// Structure JWT Claims correspondante
type JWTClaims struct {
    Sub         string   `json:"sub"`
    Username    string   `json:"username"`
    Email       *string  `json:"email,omitempty"`
    Roles       []string `json:"roles"`
    Permissions []string `json:"permissions"`
    Exp         int64    `json:"exp"`
    Iat         int64    `json:"iat"`
    Iss         string   `json:"iss"`
    Aud         string   `json:"aud"`
    SessionID   string   `json:"session_id"`
    jwt.RegisteredClaims
}

// Client d'authentification
type AuthClient struct {
    baseURL     string
    httpClient  *http.Client
    jwtSecret   []byte
}

func (c *AuthClient) ValidateToken(tokenString string) (*JWTClaims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
        return c.jwtSecret, nil
    })
    
    if err != nil {
        return nil, err
    }
    
    if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
        return claims, nil
    }
    
    return nil, errors.New("invalid token")
}

func (c *AuthClient) HasPermission(claims *JWTClaims, permission string) bool {
    for _, p := range claims.Permissions {
        if p == permission {
            return true
        }
    }
    return false
}

// Middleware Go pour validation JWT
func (c *AuthClient) JWTMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
            return
        }
        
        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        claims, err := c.ValidateToken(tokenString)
        if err != nil {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }
        
        // Ajouter les claims au contexte
        ctx := context.WithValue(r.Context(), "claims", claims)
        next.ServeHTTP(w, r.WithContext(ctx))
    }
}
```

### Avec le frontend React

```typescript
// Types TypeScript correspondants
interface Claims {
  sub: string;
  username: string;
  email?: string;
  roles: Role[];
  permissions: Permission[];
  exp: number;
  iat: number;
  iss: string;
  aud: string;
  session_id: string;
}

interface LoginRequest {
  username: string;
  password: string;
  remember_me?: boolean;
}

interface LoginResponse {
  access_token: string;
  refresh_token: string;
  token_type: string;
  expires_in: number;
  user_info: UserInfo;
}

// Service d'authentification
class AuthService {
  private accessToken: string | null = null;
  private refreshToken: string | null = null;
  
  async login(credentials: LoginRequest): Promise<LoginResponse> {
    const response = await fetch('/api/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(credentials),
    });
    
    if (!response.ok) {
      throw new Error('Login failed');
    }
    
    const loginResponse: LoginResponse = await response.json();
    
    // Stocker les tokens
    this.setTokens(loginResponse.access_token, loginResponse.refresh_token);
    
    return loginResponse;
  }
  
  async refreshAccessToken(): Promise<string> {
    if (!this.refreshToken) {
      throw new Error('No refresh token available');
    }
    
    const response = await fetch('/api/auth/refresh', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: this.refreshToken }),
    });
    
    if (!response.ok) {
      throw new Error('Token refresh failed');
    }
    
    const { access_token, refresh_token } = await response.json();
    this.setTokens(access_token, refresh_token);
    
    return access_token;
  }
  
  async apiRequest<T>(url: string, options: RequestInit = {}): Promise<T> {
    let token = this.accessToken;
    
    // Essayer la requête avec le token actuel
    let response = await this.makeRequest(url, token, options);
    
    // Si le token est expiré, essayer de le rafraîchir
    if (response.status === 401) {
      try {
        token = await this.refreshAccessToken();
        response = await this.makeRequest(url, token, options);
      } catch (error) {
        // Échec du rafraîchissement - rediriger vers login
        this.logout();
        throw new Error('Authentication required');
      }
    }
    
    return response.json();
  }
  
  private async makeRequest(url: string, token: string | null, options: RequestInit) {
    return fetch(url, {
      ...options,
      headers: {
        ...options.headers,
        ...(token && { Authorization: `Bearer ${token}` }),
      },
    });
  }
  
  private setTokens(accessToken: string, refreshToken: string) {
    this.accessToken = accessToken;
    this.refreshToken = refreshToken;
    
    // Stocker dans localStorage
    localStorage.setItem('access_token', accessToken);
    localStorage.setItem('refresh_token', refreshToken);
  }
  
  logout() {
    this.accessToken = null;
    this.refreshToken = null;
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
  }
}

// Hook React pour l'authentification
export function useAuth() {
  const [user, setUser] = useState<UserInfo | null>(null);
  const [loading, setLoading] = useState(true);
  const authService = new AuthService();
  
  useEffect(() => {
    const initAuth = async () => {
      const token = localStorage.getItem('access_token');
      if (token) {
        try {
          const userInfo = await authService.apiRequest<Claims>('/api/auth/me');
          setUser(userInfo);
        } catch (error) {
          // Token invalide
          authService.logout();
        }
      }
      setLoading(false);
    };
    
    initAuth();
  }, []);
  
  const login = async (credentials: LoginRequest) => {
    const response = await authService.login(credentials);
    setUser(response.user_info);
    return response;
  };
  
  const logout = () => {
    authService.logout();
    setUser(null);
  };
  
  const hasPermission = (permission: Permission): boolean => {
    return user?.permissions.includes(permission) || false;
  };
  
  const hasRole = (role: Role): boolean => {
    return user?.roles.includes(role) || false;
  };
  
  return {
    user,
    loading,
    login,
    logout,
    hasPermission,
    hasRole,
    apiRequest: authService.apiRequest.bind(authService),
  };
}
```

Cette documentation complète du module d'authentification vous permet de mettre en place un système de sécurité robuste et flexible pour votre plateforme de streaming audio, avec une intégration parfaite entre Rust, Go et React. 