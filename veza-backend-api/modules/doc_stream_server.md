# Stream Server - Documentation

## Vue d'ensemble

Le Stream Server est un serveur HTTP développé en Rust utilisant Axum pour la diffusion sécurisée de fichiers audio avec support du streaming partiel (HTTP Range Requests). Il implémente un système de signatures temporisées pour sécuriser l'accès aux ressources.

## Architecture

### Structure du projet

```
src/
├── main.rs      # Point d'entrée et configuration du serveur
├── routes.rs    # Définition des routes et handlers
└── utils.rs     # Utilitaires (streaming, validation signatures)
```

## Fonctionnalités principales

### 1. Streaming audio sécurisé
- Support des requêtes HTTP Range pour le streaming partiel
- Validation de signatures temporisées (signed URLs)
- Diffusion de fichiers MP3 avec headers appropriés

### 2. Sécurité par signatures
- URLs signées avec expiration temporelle
- Protection HMAC-SHA256 contre la falsification
- Validation automatique des permissions d'accès

### 3. Support CORS
- Configuration CORS pour les applications web
- Headers exposés pour le contrôle de plage
- Support des requêtes cross-origin

## Configuration

### Structure des fichiers

```
audio/
├── track1.mp3
├── track2.mp3
└── ...
```

### Variables d'environnement

```bash
# Port d'écoute (optionnel, défaut: 8082)
STREAM_SERVER_PORT=8082

# Clé secrète pour les signatures (définie dans le code)
SECRET_KEY="&!zg0N4f1L4@0Z1NkUc9heFiwZMI4KmRzVd0viTU7#JO$YrJ&g!!h54G"
```

### Dépendances principales

```toml
[dependencies]
axum = "0.7"
tokio = { version = "1.36", features = ["full"] }
tower = "0.4"
mime_guess = "2.0"
tokio-util = { version = "0.7", features = ["codec"] }
headers = "0.4"
axum-extra = { version = "0.9", features = ["typed-header"] }
tower-http = { version = "0.5", features = ["cors"] }
hmac = "0.12"
sha2 = "0.10"
hex = "0.4"
url = "2.5"
chrono = { version = "0.4", features = ["clock"] }
```

## API REST

### Endpoint de streaming

#### GET `/stream/{id}`

Diffuse un fichier audio avec support du streaming partiel.

**Paramètres d'URL :**
- `id` : Identifiant du fichier (sans extension)

**Paramètres de requête obligatoires :**
- `expires` : Timestamp Unix d'expiration
- `sig` : Signature HMAC-SHA256 de sécurité

**Headers supportés :**
- `Range` : Pour les requêtes de plage (ex: `bytes=0-1023`)

**Exemple d'URL signée :**
```
http://localhost:8082/stream/track1?expires=1609459200&sig=abc123def456...
```

**Réponses :**

##### 206 Partial Content (succès avec Range)
```http
HTTP/1.1 206 Partial Content
Content-Type: audio/mpeg
Accept-Ranges: bytes
Content-Length: 1024
Content-Range: bytes 0-1023/3145728

[données binaires audio]
```

##### 200 OK (succès sans Range)
```http
HTTP/1.1 200 OK
Content-Type: audio/mpeg
Content-Length: 3145728

[données binaires audio complètes]
```

##### 403 Forbidden (signature invalide)
```http
HTTP/1.1 403 Forbidden

Signature invalide
```

##### 404 Not Found (fichier inexistant)
```http
HTTP/1.1 404 Not Found

Fichier non trouvé
```

## Système de signatures

### Génération de signatures

```rust
use hmac::{Hmac, Mac};
use sha2::Sha256;

fn generate_signature(filename: &str, expires: i64, secret: &str) -> String {
    let to_sign = format!("{}|{}", filename, expires);
    let mut mac = Hmac::<Sha256>::new_from_slice(secret.as_bytes()).unwrap();
    mac.update(to_sign.as_bytes());
    let result = mac.finalize();
    hex::encode(result.into_bytes())
}
```

### Exemple de génération d'URL signée

```rust
fn create_signed_url(filename: &str, expires_in_seconds: i64) -> String {
    let expires = chrono::Utc::now().timestamp() + expires_in_seconds;
    let signature = generate_signature(filename, expires, SECRET_KEY);
    
    format!(
        "http://localhost:8082/stream/{}?expires={}&sig={}",
        filename, expires, signature
    )
}

// Utilisation
let url = create_signed_url("track1", 3600); // Expire dans 1 heure
```

### Validation côté serveur

1. **Vérification de l'expiration** : `current_time <= expires`
2. **Recalcul de la signature** : HMAC-SHA256 du payload
3. **Comparaison sécurisée** : signature fournie vs signature attendue

## Configuration CORS

### Headers exposés

```rust
let exposed_headers = vec![
    HeaderName::from_static("content-range"),
    HeaderName::from_static("content-length"),
    HeaderName::from_static("accept-ranges"),
];
```

### Politique CORS

```rust
let cors = CorsLayer::new()
    .allow_origin(Any)           // Toutes les origines (à restreindre en production)
    .allow_methods([Method::GET]) // Seulement GET
    .allow_headers(Any)          // Tous les headers
    .expose_headers(exposed_headers);
```

## Gestion du streaming partiel

### Format des requêtes Range

```http
Range: bytes=0-1023        # Octets 0 à 1023
Range: bytes=1024-2047     # Octets 1024 à 2047
Range: bytes=1024-         # Depuis l'octet 1024 jusqu'à la fin
```

### Parsing des headers Range

```rust
fn parse_range(header: &str) -> Option<(u64, u64)> {
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
```

### Lecture de fichier partielle

```rust
pub async fn serve_partial_file(
    path: PathBuf, 
    headers: HeaderMap
) -> Result<Response, (StatusCode, String)> {
    let metadata = tokio::fs::metadata(&path).await?;
    let total_size = metadata.len();
    
    let range = headers
        .get("range")
        .and_then(|v| v.to_str().ok())
        .and_then(parse_range);
    
    let (start, end) = range.unwrap_or((0, total_size - 1));
    let chunk_size = end - start + 1;
    
    let mut file = File::open(&path).await?;
    file.seek(SeekFrom::Start(start)).await?;
    
    let mut buffer = vec![0; chunk_size as usize];
    file.read_exact(&mut buffer).await?;
    
    // Construction de la réponse avec headers appropriés
    // ...
}
```

## Sécurité

### Mesures de sécurité implémentées

1. **URLs signées temporisées** pour contrôler l'accès
2. **Validation HMAC** pour prévenir la falsification
3. **Expiration automatique** des liens
4. **Validation des chemins** pour éviter les directory traversal
5. **Headers de sécurité** appropriés

### Recommandations de sécurité

#### Pour la production

```rust
// Restriction CORS plus stricte
let cors = CorsLayer::new()
    .allow_origin("https://yourdomain.com".parse::<HeaderValue>().unwrap())
    .allow_methods([Method::GET])
    .allow_headers([header::RANGE, header::AUTHORIZATION])
    .expose_headers(exposed_headers);
```

#### Variables d'environnement sécurisées

```bash
# Utiliser une clé secrète forte et unique
SECRET_KEY=$(openssl rand -hex 32)

# Restriction des origines
ALLOWED_ORIGINS=https://app.example.com,https://admin.example.com
```

#### Validation des chemins

```rust
fn validate_file_path(id: &str) -> Result<PathBuf, Error> {
    // Validation du nom de fichier
    if id.contains("..") || id.contains('/') || id.contains('\\') {
        return Err(Error::InvalidPath);
    }
    
    // Limitation aux caractères alphanumériques
    if !id.chars().all(|c| c.is_alphanumeric() || c == '_' || c == '-') {
        return Err(Error::InvalidCharacters);
    }
    
    Ok(PathBuf::from(format!("audio/{}.mp3", id)))
}
```

## Performance

### Optimisations implémentées

1. **Streaming par chunks** pour éviter le chargement complet en mémoire
2. **Support des requêtes Range** pour économiser la bande passante
3. **Headers de cache** appropriés
4. **Async I/O** avec Tokio pour la concurrence

### Métriques recommandées

- Débit de streaming (MB/s)
- Nombre de connexions simultanées
- Latence des requêtes
- Taux d'erreur 403/404
- Utilisation CPU/mémoire

### Optimisations avancées

```rust
// Cache des métadonnées de fichiers
use std::collections::HashMap;
use tokio::sync::RwLock;

struct FileCache {
    metadata: RwLock<HashMap<String, (u64, SystemTime)>>, // taille, modified
}

// Headers de cache
response.headers_mut().insert(
    "Cache-Control", 
    "public, max-age=3600".parse().unwrap()
);
response.headers_mut().insert(
    "ETag", 
    format!("\"{}\"", file_hash).parse().unwrap()
);
```

## Monitoring et logging

### Configuration du logging

```rust
tracing_subscriber::fmt()
    .with_env_filter("stream_server=info")
    .json()
    .init();
```

### Événements à tracer

```rust
// Requêtes de streaming
tracing::info!(
    file_id = %id,
    range = ?range,
    client_ip = %client_ip,
    "Streaming request"
);

// Erreurs de sécurité
tracing::warn!(
    file_id = %id,
    signature = %sig,
    expires = %expires,
    "Invalid signature attempt"
);

// Métriques de performance
tracing::debug!(
    file_id = %id,
    bytes_served = chunk_size,
    duration_ms = elapsed.as_millis(),
    "Stream completed"
);
```

## Déploiement

### Compilation

```bash
cd backend/modules/stream_server
cargo build --release
```

### Structure de déploiement

```
/opt/stream_server/
├── stream_server          # Binaire
├── audio/                 # Dossier des fichiers audio
│   ├── track1.mp3
│   └── track2.mp3
└── config.env            # Variables d'environnement
```

### Service systemd

```ini
[Unit]
Description=Stream Server
After=network.target

[Service]
Type=simple
User=streamserver
WorkingDirectory=/opt/stream_server
ExecStart=/opt/stream_server/stream_server
EnvironmentFile=/opt/stream_server/config.env
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

### Docker

```dockerfile
FROM rust:1.70 as builder
WORKDIR /app
COPY . .
RUN cargo build --release

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y ca-certificates
WORKDIR /app
COPY --from=builder /app/target/release/stream_server ./
COPY audio/ ./audio/
EXPOSE 8082
CMD ["./stream_server"]
```

### Reverse proxy (Nginx)

```nginx
upstream stream_server {
    server 127.0.0.1:8082;
}

server {
    listen 443 ssl;
    server_name streams.example.com;
    
    location /stream/ {
        proxy_pass http://stream_server;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        
        # Support des requêtes Range
        proxy_http_version 1.1;
        proxy_set_header Range $http_range;
        
        # Cache pour les ressources statiques
        expires 1h;
        add_header Cache-Control "public, immutable";
    }
}
```

## Tests et développement

### Tests unitaires

```bash
cargo test
```

### Tests d'intégration

```rust
#[tokio::test]
async fn test_valid_signature() {
    let app = create_test_app().await;
    
    let expires = chrono::Utc::now().timestamp() + 3600;
    let signature = generate_signature("test", expires, SECRET_KEY);
    
    let response = app
        .oneshot(
            Request::builder()
                .uri(&format!("/stream/test?expires={}&sig={}", expires, signature))
                .body(Body::empty())
                .unwrap(),
        )
        .await
        .unwrap();
    
    assert_eq!(response.status(), StatusCode::OK);
}
```

### Génération de liens de test

```rust
fn main() {
    let url = create_signed_url("test_track", 3600);
    println!("Test URL: {}", url);
}
```

### Debugging

1. **Logs détaillés** : `RUST_LOG=stream_server=debug`
2. **Test des signatures** avec des outils externes
3. **Monitoring des requêtes** avec curl
4. **Validation des headers** HTTP

```bash
# Test d'une requête Range
curl -H "Range: bytes=0-1023" \
     "http://localhost:8082/stream/test?expires=1609459200&sig=..."

# Inspection des headers
curl -I "http://localhost:8082/stream/test?expires=1609459200&sig=..."
```