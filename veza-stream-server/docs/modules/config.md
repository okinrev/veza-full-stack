# Module Configuration - Documentation Détaillée

## Vue d'ensemble

Le module de configuration (`src/config/mod.rs`) centralise toute la configuration du Stream Server. Il utilise les variables d'environnement avec des valeurs par défaut sensées et inclut une validation complète.

## Structures Principales

### `Config`
Structure principale contenant toute la configuration de l'application.

```rust
pub struct Config {
    // Configuration de base
    pub secret_key: String,                    // Clé secrète pour les signatures
    pub port: u16,                            // Port d'écoute du serveur
    pub audio_dir: String,                    // Répertoire des fichiers audio
    pub allowed_origins: Vec<String>,         // Origines CORS autorisées
    pub max_file_size: u64,                   // Taille max des fichiers (bytes)
    pub max_range_size: u64,                  // Taille max des Range Requests (bytes)
    pub signature_tolerance: i64,             // Tolérance de signature (secondes)
    
    // Sous-configurations
    pub database: DatabaseConfig,
    pub cache: CacheConfig,
    pub security: SecurityConfig,
    pub performance: PerformanceConfig,
    pub monitoring: MonitoringConfig,
    pub notifications: NotificationConfig,
    pub compression: CompressionConfig,
    pub environment: Environment,
}
```

### `DatabaseConfig`
Configuration pour la base de données SQLite.

```rust
pub struct DatabaseConfig {
    pub url: String,                          // URL de connexion SQLite
    pub max_connections: u32,                 // Pool de connexions max
    pub min_connections: u32,                 // Pool de connexions min
    pub connection_timeout: Duration,         // Timeout de connexion
    pub idle_timeout: Duration,               // Timeout d'inactivité
    pub max_lifetime: Duration,               // Durée de vie max d'une connexion
    pub enable_logging: bool,                 // Logs des requêtes SQL
    pub migrate_on_start: bool,               // Migration auto au démarrage
}
```

**Variables d'environnement :**
- `DATABASE_URL` (défaut: `"sqlite:stream_server.db"`)
- `DATABASE_MAX_CONNECTIONS` (défaut: `10`)
- `DATABASE_MIN_CONNECTIONS` (défaut: `1`) 
- `DATABASE_CONNECTION_TIMEOUT` (défaut: `30` secondes)
- `DATABASE_IDLE_TIMEOUT` (défaut: `300` secondes)
- `DATABASE_MAX_LIFETIME` (défaut: `3600` secondes)
- `DATABASE_ENABLE_LOGGING` (défaut: `false`)
- `DATABASE_MIGRATE_ON_START` (défaut: `true`)

### `CacheConfig`
Configuration du système de cache.

```rust
pub struct CacheConfig {
    pub max_size_mb: u64,                     // Taille max du cache (MB)
    pub ttl_seconds: u64,                     // Time To Live (secondes)
    pub cleanup_interval: Duration,           // Intervalle de nettoyage
    pub compression_enabled: bool,            // Compression du cache
    pub redis_url: Option<String>,            // URL Redis (optionnel)
    pub redis_pool_size: Option<u32>,         // Taille du pool Redis
}
```

**Variables d'environnement :**
- `CACHE_MAX_SIZE_MB` (défaut: `256`)
- `CACHE_TTL_SECONDS` (défaut: `3600`)
- `CACHE_CLEANUP_INTERVAL` (défaut: `300` secondes)
- `CACHE_COMPRESSION_ENABLED` (défaut: `true`)
- `REDIS_URL` (optionnel)
- `REDIS_POOL_SIZE` (défaut: `10`)

### `SecurityConfig`
Configuration de sécurité.

```rust
pub struct SecurityConfig {
    pub jwt_secret: Option<String>,           // Clé secrète JWT (optionnel)
    pub jwt_expiration: Duration,             // Expiration des JWT
    pub bcrypt_cost: u32,                     // Coût bcrypt pour les mots de passe
    pub rate_limit_requests_per_minute: u32,  // Limite de requêtes/minute
    pub rate_limit_burst: u32,                // Burst autorisé
    pub cors_max_age: Duration,               // Cache CORS max-age
    pub csrf_protection: bool,                // Protection CSRF
    pub secure_headers: bool,                 // Headers de sécurité
    pub tls_cert_path: Option<String>,        // Certificat TLS (optionnel)
    pub tls_key_path: Option<String>,         // Clé privée TLS (optionnel)
}
```

**Variables d'environnement :**
- `JWT_SECRET` (optionnel, requis en production)
- `JWT_EXPIRATION` (défaut: `3600` secondes)
- `BCRYPT_COST` (défaut: `10`)
- `RATE_LIMIT_RPM` (défaut: `60`)
- `RATE_LIMIT_BURST` (défaut: `10`)
- `CORS_MAX_AGE` (défaut: `86400` secondes)
- `CSRF_PROTECTION` (défaut: `true`)
- `SECURE_HEADERS` (défaut: `true`)
- `TLS_CERT_PATH` (optionnel)
- `TLS_KEY_PATH` (optionnel)

### `PerformanceConfig`
Configuration de performance.

```rust
pub struct PerformanceConfig {
    pub worker_threads: Option<usize>,        // Nombre de threads workers
    pub max_blocking_threads: Option<usize>,  // Threads bloquants max
    pub thread_stack_size: Option<usize>,     // Taille de stack des threads
    pub tcp_nodelay: bool,                    // Option TCP_NODELAY
    pub tcp_keepalive: Option<Duration>,      // TCP keepalive
    pub buffer_size: usize,                   // Taille des buffers
    pub max_concurrent_streams: usize,        // Streams concurrents max
    pub stream_timeout: Duration,             // Timeout des streams
    pub compression_level: u8,                // Niveau de compression HTTP
}
```

**Variables d'environnement :**
- `WORKER_THREADS` (défaut: auto-détecté)
- `MAX_BLOCKING_THREADS` (défaut: `512`)
- `THREAD_STACK_SIZE` (défaut: système)
- `TCP_NODELAY` (défaut: `true`)
- `TCP_KEEPALIVE` (défaut: `60` secondes)
- `BUFFER_SIZE` (défaut: `8192`)
- `MAX_CONCURRENT_STREAMS` (défaut: `100`)
- `STREAM_TIMEOUT` (défaut: `30` secondes)
- `COMPRESSION_LEVEL` (défaut: `6`)

### `MonitoringConfig`
Configuration du monitoring.

```rust
pub struct MonitoringConfig {
    pub metrics_enabled: bool,                // Métriques activées
    pub metrics_port: u16,                    // Port des métriques
    pub health_check_interval: Duration,      // Intervalle health checks
    pub log_level: String,                    // Niveau de logs
    pub log_format: LogFormat,                // Format des logs
    pub jaeger_endpoint: Option<String>,      // Endpoint Jaeger (optionnel)
    pub prometheus_namespace: String,         // Namespace Prometheus
    pub alert_webhooks: Vec<String>,          // Webhooks d'alertes
}
```

**Variables d'environnement :**
- `METRICS_ENABLED` (défaut: `true`)
- `METRICS_PORT` (défaut: `9090`)
- `HEALTH_CHECK_INTERVAL` (défaut: `30` secondes)
- `RUST_LOG` (défaut: `"stream_server=info"`)
- `LOG_FORMAT` (défaut: `"pretty"`, options: `"pretty"`, `"json"`, `"compact"`)
- `JAEGER_ENDPOINT` (optionnel)
- `PROMETHEUS_NAMESPACE` (défaut: `"stream_server"`)
- `ALERT_WEBHOOKS` (séparés par virgules)

### `NotificationConfig`
Configuration des notifications.

```rust
pub struct NotificationConfig {
    pub enabled: bool,                        // Notifications activées
    pub max_queue_size: usize,                // Taille max de la queue
    pub delivery_workers: usize,              // Nombre de workers de livraison
    pub retry_attempts: u32,                  // Tentatives de retry
    pub retry_delay: Duration,                // Délai entre les retries
    pub batch_size: usize,                    // Taille des batches
    pub email_provider: Option<EmailProvider>, // Fournisseur email
    pub sms_provider: Option<SmsProvider>,    // Fournisseur SMS
    pub push_provider: Option<PushProvider>,  // Fournisseur push
}
```

**Variables d'environnement :**
- `NOTIFICATIONS_ENABLED` (défaut: `true`)
- `NOTIFICATION_QUEUE_SIZE` (défaut: `1000`)
- `NOTIFICATION_WORKERS` (défaut: `2`)
- `NOTIFICATION_RETRY_ATTEMPTS` (défaut: `3`)
- `NOTIFICATION_RETRY_DELAY` (défaut: `60` secondes)
- `NOTIFICATION_BATCH_SIZE` (défaut: `10`)

### `CompressionConfig`
Configuration de la compression audio.

```rust
pub struct CompressionConfig {
    pub enabled: bool,                        // Compression activée
    pub output_dir: String,                   // Répertoire de sortie
    pub temp_dir: String,                     // Répertoire temporaire
    pub max_concurrent_jobs: usize,           // Jobs concurrents max
    pub cleanup_after_days: u32,              // Nettoyage après X jours
    pub ffmpeg_path: Option<String>,          // Chemin vers FFmpeg
    pub quality_profiles: Vec<String>,        // Profils de qualité
}
```

**Variables d'environnement :**
- `COMPRESSION_ENABLED` (défaut: `false`)
- `COMPRESSION_OUTPUT_DIR` (défaut: `"compressed"`)
- `COMPRESSION_TEMP_DIR` (défaut: `"temp"`)
- `COMPRESSION_MAX_JOBS` (défaut: `2`)
- `COMPRESSION_CLEANUP_DAYS` (défaut: `7`)
- `FFMPEG_PATH` (optionnel)
- `COMPRESSION_PROFILES` (défaut: `"high,medium,low"`)

## Énumérations

### `Environment`
Environnement d'exécution.

```rust
pub enum Environment {
    Development,  // Développement
    Testing,      // Tests
    Staging,      // Staging
    Production,   // Production
}
```

**Variable d'environnement :**
- `ENVIRONMENT` (défaut: `"development"`)

### `LogFormat`
Format des logs.

```rust
pub enum LogFormat {
    Pretty,   // Format lisible par humain
    Json,     // Format JSON structuré
    Compact,  // Format compact
}
```

## Fournisseurs Externes

### `EmailProvider`
Configuration pour l'envoi d'emails.

```rust
pub struct EmailProvider {
    pub provider_type: String,           // Type de fournisseur
    pub api_key: Option<String>,         // Clé API
    pub smtp_host: Option<String>,       // Host SMTP
    pub smtp_port: Option<u16>,          // Port SMTP
    pub smtp_username: Option<String>,   // Utilisateur SMTP
    pub smtp_password: Option<String>,   // Mot de passe SMTP
    pub from_email: String,              // Email expéditeur
    pub from_name: String,               // Nom expéditeur
}
```

**Variables d'environnement :**
- `EMAIL_PROVIDER_TYPE` (ex: `"smtp"`, `"sendgrid"`, `"mailgun"`)
- `EMAIL_API_KEY`
- `SMTP_HOST`
- `SMTP_PORT`
- `SMTP_USERNAME`
- `SMTP_PASSWORD`
- `EMAIL_FROM_EMAIL`
- `EMAIL_FROM_NAME`

### `SmsProvider`
Configuration pour l'envoi de SMS.

```rust
pub struct SmsProvider {
    pub provider_type: String,           // Type de fournisseur
    pub api_key: String,                 // Clé API
    pub api_secret: Option<String>,      // Secret API
    pub from_number: String,             // Numéro expéditeur
}
```

**Variables d'environnement :**
- `SMS_PROVIDER_TYPE` (ex: `"twilio"`, `"nexmo"`)
- `SMS_API_KEY`
- `SMS_API_SECRET`
- `SMS_FROM_NUMBER`

### `PushProvider`
Configuration pour les notifications push.

```rust
pub struct PushProvider {
    pub provider_type: String,           // Type de fournisseur
    pub api_key: String,                 // Clé API
    pub project_id: Option<String>,      // ID du projet
    pub bundle_id: Option<String>,       // Bundle ID (iOS)
}
```

**Variables d'environnement :**
- `PUSH_PROVIDER_TYPE` (ex: `"fcm"`, `"apns"`)
- `PUSH_API_KEY`
- `PUSH_PROJECT_ID`
- `PUSH_BUNDLE_ID`

## Méthodes Principales

### `Config::from_env()`
Charge la configuration depuis les variables d'environnement.

```rust
pub fn from_env() -> Result<Self, ConfigError>
```

**Exemple d'utilisation :**
```rust
let config = Config::from_env()
    .map_err(|e| format!("Erreur de configuration: {}", e))?;
```

### `Config::validate()`
Valide la configuration chargée.

```rust
pub fn validate(&self) -> Result<(), ConfigError>
```

**Validations effectuées :**
- Port dans la plage valide (1-65535)
- Répertoire audio existe et accessible
- Clé secrète suffisamment forte
- JWT secret présent en production
- Tailles de fichiers cohérentes

### Méthodes d'introspection

```rust
pub fn is_production(&self) -> bool        // Vérifie si en production
pub fn is_development(&self) -> bool       // Vérifie si en développement
pub fn redis_enabled(&self) -> bool        // Vérifie si Redis est configuré
pub fn tls_enabled(&self) -> bool          // Vérifie si TLS est activé
pub fn metrics_enabled(&self) -> bool      // Vérifie si les métriques sont activées
pub fn notifications_enabled(&self) -> bool // Vérifie si les notifications sont activées
pub fn compression_enabled(&self) -> bool  // Vérifie si la compression est activée
```

## Erreurs de Configuration

### `ConfigError`
Énumération des erreurs de configuration possibles.

```rust
pub enum ConfigError {
    InvalidPort,                    // Port invalide
    InvalidAudioDir,                // Répertoire audio invalide
    WeakSecretKey,                  // Clé secrète faible
    MissingJwtSecret,               // JWT secret manquant en production
    InvalidFileSize,                // Taille de fichier invalide
    InvalidRangeSize,               // Taille de range invalide
    InvalidSignatureTolerance,      // Tolérance de signature invalide
    InvalidDatabaseConfig,          // Configuration de base de données invalide
}
```

## Exemple de Configuration Complète

### Fichier `.env` pour développement

```env
# Configuration de base
STREAM_SERVER_PORT=8082
AUDIO_DIR=./audio
SECRET_KEY=development-secret-key-32-chars
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8080
MAX_FILE_SIZE=104857600
MAX_RANGE_SIZE=10485760
SIGNATURE_TOLERANCE=60

# Base de données
DATABASE_URL=sqlite:stream_server.db
DATABASE_MAX_CONNECTIONS=10
DATABASE_ENABLE_LOGGING=true

# Cache
CACHE_MAX_SIZE_MB=256
CACHE_TTL_SECONDS=3600

# Sécurité
JWT_SECRET=jwt-secret-key-for-development
JWT_EXPIRATION=3600
RATE_LIMIT_RPM=120

# Performance
WORKER_THREADS=4
MAX_CONCURRENT_STREAMS=50

# Monitoring
METRICS_ENABLED=true
RUST_LOG=stream_server=debug,tower_http=info
LOG_FORMAT=pretty

# Notifications
NOTIFICATIONS_ENABLED=true
EMAIL_PROVIDER_TYPE=smtp
SMTP_HOST=localhost
SMTP_PORT=1025

# Compression
COMPRESSION_ENABLED=false
```

### Fichier `.env` pour production

```env
# Configuration de base
STREAM_SERVER_PORT=8082
AUDIO_DIR=/app/audio
SECRET_KEY=super-secure-production-key-32chars
ALLOWED_ORIGINS=https://yourdomain.com,https://admin.yourdomain.com
MAX_FILE_SIZE=104857600
MAX_RANGE_SIZE=10485760
SIGNATURE_TOLERANCE=30

# Base de données
DATABASE_URL=sqlite:/data/stream_server.db
DATABASE_MAX_CONNECTIONS=20
DATABASE_ENABLE_LOGGING=false

# Cache
CACHE_MAX_SIZE_MB=512
CACHE_TTL_SECONDS=7200
REDIS_URL=redis://redis:6379

# Sécurité
JWT_SECRET=production-jwt-secret-very-long-and-secure
JWT_EXPIRATION=1800
RATE_LIMIT_RPM=60
CORS_MAX_AGE=3600

# Performance
WORKER_THREADS=8
MAX_CONCURRENT_STREAMS=200
COMPRESSION_LEVEL=9

# Monitoring
METRICS_ENABLED=true
METRICS_PORT=9090
RUST_LOG=stream_server=info,tower_http=warn
LOG_FORMAT=json
PROMETHEUS_NAMESPACE=streamserver

# Notifications
NOTIFICATIONS_ENABLED=true
EMAIL_PROVIDER_TYPE=sendgrid
EMAIL_API_KEY=SG.your-sendgrid-api-key
EMAIL_FROM_EMAIL=noreply@yourdomain.com
EMAIL_FROM_NAME=Stream Server

# Compression
COMPRESSION_ENABLED=true
COMPRESSION_MAX_JOBS=4
FFMPEG_PATH=/usr/bin/ffmpeg

# Environnement
ENVIRONMENT=production
```

## Intégration avec Backend API

### Go
```go
type StreamServerConfig struct {
    BaseURL     string
    SecretKey   string
    JWTSecret   string
    Environment string
}

func (c *StreamServerConfig) GenerateSignedURL(filename string, duration time.Duration) string {
    expires := time.Now().Add(duration).Unix()
    signature := c.generateSignature(filename, expires)
    
    return fmt.Sprintf("%s/stream/%s?expires=%d&sig=%s",
        c.BaseURL, filename, expires, signature)
}
```

### Variables d'environnement critiques pour l'intégration
- `SECRET_KEY` : **OBLIGATOIRE** - Doit être partagée avec le backend API
- `JWT_SECRET` : **RECOMMANDÉ** - Pour la validation des tokens
- `ALLOWED_ORIGINS` : **CRITIQUE** - Doit inclure l'origine du frontend
- `SIGNATURE_TOLERANCE` : **IMPORTANT** - Tolérance de décalage horaire

## Bonnes Pratiques

### Sécurité
1. **Clés secrètes** : Générer des clés de 32+ caractères aléatoirement
2. **CORS** : Limiter aux domaines strictement nécessaires en production
3. **TLS** : Toujours activer en production
4. **Rate limiting** : Ajuster selon la charge attendue

### Performance
1. **Cache** : Augmenter la taille en production
2. **Workers** : Ajuster selon le nombre de CPU
3. **Connexions DB** : Monitorer et ajuster selon la charge
4. **Compression** : Activer uniquement si nécessaire

### Monitoring
1. **Logs** : Format JSON en production pour l'analyse
2. **Métriques** : Toujours activées pour surveiller la santé
3. **Health checks** : Intégrer dans le load balancer
4. **Alertes** : Configurer les webhooks pour les incidents

---

**Note** : Cette configuration est conçue pour être flexible et sécurisée par défaut. Chaque paramètre peut être ajusté selon les besoins spécifiques de votre déploiement. 