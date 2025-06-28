use std::env;
use std::time::Duration;
use serde::{Deserialize, Serialize};

#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct Config {
    // Configuration de base
    pub secret_key: String,
    pub port: u16,
    pub audio_dir: String,
    pub allowed_origins: Vec<String>,
    pub max_file_size: u64,
    pub max_range_size: u64,
    pub signature_tolerance: i64,
    
    // Configuration de base de données
    pub database: DatabaseConfig,
    
    // Configuration de cache
    pub cache: CacheConfig,
    
    // Configuration de sécurité
    pub security: SecurityConfig,
    
    // Configuration de performance
    pub performance: PerformanceConfig,
    
    // Configuration de monitoring
    pub monitoring: MonitoringConfig,
    
    // Configuration de notifications
    pub notifications: NotificationConfig,
    
    // Configuration de compression
    pub compression: CompressionConfig,
    
    // Profil d'environnement
    pub environment: Environment,
}

#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct DatabaseConfig {
    pub url: String,
    pub max_connections: u32,
    pub min_connections: u32,
    pub connection_timeout: Duration,
    pub idle_timeout: Duration,
    pub max_lifetime: Duration,
    pub enable_logging: bool,
    pub migrate_on_start: bool,
}

#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct CacheConfig {
    pub max_size_mb: u64,
    pub ttl_seconds: u64,
    pub cleanup_interval: Duration,
    pub compression_enabled: bool,
    pub redis_url: Option<String>,
    pub redis_pool_size: Option<u32>,
}

#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct SecurityConfig {
    pub jwt_secret: Option<String>,
    pub jwt_expiration: Duration,
    pub bcrypt_cost: u32,
    pub rate_limit_requests_per_minute: u32,
    pub rate_limit_burst: u32,
    pub cors_max_age: Duration,
    pub csrf_protection: bool,
    pub secure_headers: bool,
    pub tls_cert_path: Option<String>,
    pub tls_key_path: Option<String>,
}

#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct PerformanceConfig {
    pub worker_threads: Option<usize>,
    pub max_blocking_threads: Option<usize>,
    pub thread_stack_size: Option<usize>,
    pub tcp_nodelay: bool,
    pub tcp_keepalive: Option<Duration>,
    pub buffer_size: usize,
    pub max_concurrent_streams: usize,
    pub stream_timeout: Duration,
    pub compression_level: u8,
}

#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct MonitoringConfig {
    pub metrics_enabled: bool,
    pub metrics_port: u16,
    pub health_check_interval: Duration,
    pub log_level: String,
    pub log_format: LogFormat,
    pub jaeger_endpoint: Option<String>,
    pub prometheus_namespace: String,
    pub alert_webhooks: Vec<String>,
}

#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct NotificationConfig {
    pub enabled: bool,
    pub max_queue_size: usize,
    pub delivery_workers: usize,
    pub retry_attempts: u32,
    pub retry_delay: Duration,
    pub batch_size: usize,
    pub email_provider: Option<EmailProvider>,
    pub sms_provider: Option<SmsProvider>,
    pub push_provider: Option<PushProvider>,
}

#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct CompressionConfig {
    pub enabled: bool,
    pub output_dir: String,
    pub temp_dir: String,
    pub max_concurrent_jobs: usize,
    pub cleanup_after_days: u32,
    pub ffmpeg_path: Option<String>,
    pub quality_profiles: Vec<String>,
}

#[derive(Clone, Debug, Serialize, Deserialize)]
pub enum Environment {
    Development,
    Testing,
    Staging,
    Production,
}

#[derive(Clone, Debug, Serialize, Deserialize)]
pub enum LogFormat {
    Pretty,
    Json,
    Compact,
}

#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct EmailProvider {
    pub provider_type: String, // smtp, sendgrid, mailgun, etc.
    pub api_key: Option<String>,
    pub smtp_host: Option<String>,
    pub smtp_port: Option<u16>,
    pub smtp_username: Option<String>,
    pub smtp_password: Option<String>,
    pub from_email: String,
    pub from_name: String,
}

#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct SmsProvider {
    pub provider_type: String, // twilio, nexmo, etc.
    pub api_key: String,
    pub api_secret: Option<String>,
    pub from_number: String,
}

#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct PushProvider {
    pub provider_type: String, // fcm, apns
    pub api_key: String,
    pub project_id: Option<String>,
    pub bundle_id: Option<String>,
}

impl Config {
    pub fn from_env() -> Result<Self, ConfigError> {
        let environment = match env::var("ENVIRONMENT")
            .unwrap_or_else(|_| "development".to_string())
            .to_lowercase()
            .as_str()
        {
            "production" | "prod" => Environment::Production,
            "staging" | "stage" => Environment::Staging,
            "testing" | "test" => Environment::Testing,
            _ => Environment::Development,
        };

        let config = Self {
            secret_key: env::var("SECRET_KEY")
                .unwrap_or_else(|_| "your-secret-key-change-in-production".to_string()),
            // CONFIGURATION PORT UNIFIÉE - Port 3002 selon guide déploiement
            port: env::var("STREAM_PORT")
                .or_else(|_| env::var("PORT"))
                .unwrap_or_else(|_| "3002".to_string())
                .parse()
                .map_err(|_| ConfigError::InvalidPort)?,
            audio_dir: env::var("AUDIO_DIR")
                .unwrap_or_else(|_| "./audio".to_string()),
            allowed_origins: env::var("ALLOWED_ORIGINS")
                .unwrap_or_else(|_| "*".to_string())
                .split(',')
                .map(|s| s.trim().to_string())
                .collect(),
            max_file_size: env::var("MAX_FILE_SIZE")
                .unwrap_or_else(|_| "104857600".to_string()) // 100MB
                .parse()
                .map_err(|_| ConfigError::InvalidFileSize)?,
            max_range_size: env::var("MAX_RANGE_SIZE")
                .unwrap_or_else(|_| "10485760".to_string()) // 10MB
                .parse()
                .map_err(|_| ConfigError::InvalidRangeSize)?,
            signature_tolerance: env::var("SIGNATURE_TOLERANCE")
                .unwrap_or_else(|_| "300".to_string()) // 5 minutes
                .parse()
                .map_err(|_| ConfigError::InvalidSignatureTolerance)?,

            database: DatabaseConfig {
                // CONFIGURATION DATABASE UNIFIÉE - PostgreSQL selon guide déploiement
                url: env::var("DATABASE_URL")
                    .unwrap_or_else(|_| "postgres://veza_user:veza_password@10.5.191.154:5432/veza_db".to_string()),
                max_connections: env::var("DB_MAX_CONNECTIONS")
                    .unwrap_or_else(|_| "10".to_string())
                    .parse()
                    .unwrap_or(10),
                min_connections: env::var("DB_MIN_CONNECTIONS")
                    .unwrap_or_else(|_| "1".to_string())
                    .parse()
                    .unwrap_or(1),
                connection_timeout: Duration::from_secs(
                    env::var("DB_CONNECTION_TIMEOUT")
                        .unwrap_or_else(|_| "30".to_string())
                        .parse()
                        .unwrap_or(30)
                ),
                idle_timeout: Duration::from_secs(
                    env::var("DB_IDLE_TIMEOUT")
                        .unwrap_or_else(|_| "600".to_string())
                        .parse()
                        .unwrap_or(600)
                ),
                max_lifetime: Duration::from_secs(
                    env::var("DB_MAX_LIFETIME")
                        .unwrap_or_else(|_| "3600".to_string())
                        .parse()
                        .unwrap_or(3600)
                ),
                enable_logging: env::var("DB_ENABLE_LOGGING")
                    .unwrap_or_else(|_| "false".to_string())
                    .parse()
                    .unwrap_or(false),
                migrate_on_start: env::var("DB_MIGRATE_ON_START")
                    .unwrap_or_else(|_| "true".to_string())
                    .parse()
                    .unwrap_or(true),
            },

            cache: CacheConfig {
                max_size_mb: env::var("CACHE_MAX_SIZE_MB")
                    .unwrap_or_else(|_| "256".to_string())
                    .parse()
                    .unwrap_or(256),
                ttl_seconds: env::var("CACHE_TTL_SECONDS")
                    .unwrap_or_else(|_| "3600".to_string())
                    .parse()
                    .unwrap_or(3600),
                cleanup_interval: Duration::from_secs(
                    env::var("CACHE_CLEANUP_INTERVAL")
                        .unwrap_or_else(|_| "300".to_string())
                        .parse()
                        .unwrap_or(300)
                ),
                compression_enabled: env::var("CACHE_COMPRESSION")
                    .unwrap_or_else(|_| "true".to_string())
                    .parse()
                    .unwrap_or(true),
                redis_url: env::var("REDIS_URL").ok(),
                redis_pool_size: env::var("REDIS_POOL_SIZE")
                    .ok()
                    .and_then(|s| s.parse().ok()),
            },

            security: SecurityConfig {
                jwt_secret: Some(env::var("JWT_SECRET")
                    .unwrap_or_else(|_| "veza_unified_jwt_secret_key_2025_microservices_secure_32chars_minimum".to_string())),
                jwt_expiration: Duration::from_secs(
                    env::var("JWT_EXPIRATION")
                        .unwrap_or_else(|_| "3600".to_string())
                        .parse()
                        .unwrap_or(3600)
                ),
                bcrypt_cost: env::var("BCRYPT_COST")
                    .unwrap_or_else(|_| "12".to_string())
                    .parse()
                    .unwrap_or(12),
                rate_limit_requests_per_minute: env::var("RATE_LIMIT_RPM")
                    .unwrap_or_else(|_| "60".to_string())
                    .parse()
                    .unwrap_or(60),
                rate_limit_burst: env::var("RATE_LIMIT_BURST")
                    .unwrap_or_else(|_| "10".to_string())
                    .parse()
                    .unwrap_or(10),
                cors_max_age: Duration::from_secs(
                    env::var("CORS_MAX_AGE")
                        .unwrap_or_else(|_| "3600".to_string())
                        .parse()
                        .unwrap_or(3600)
                ),
                csrf_protection: env::var("CSRF_PROTECTION")
                    .unwrap_or_else(|_| "true".to_string())
                    .parse()
                    .unwrap_or(true),
                secure_headers: env::var("SECURE_HEADERS")
                    .unwrap_or_else(|_| "true".to_string())
                    .parse()
                    .unwrap_or(true),
                tls_cert_path: env::var("TLS_CERT_PATH").ok(),
                tls_key_path: env::var("TLS_KEY_PATH").ok(),
            },

            performance: PerformanceConfig {
                worker_threads: env::var("WORKER_THREADS")
                    .ok()
                    .and_then(|s| s.parse().ok()),
                max_blocking_threads: env::var("MAX_BLOCKING_THREADS")
                    .ok()
                    .and_then(|s| s.parse().ok()),
                thread_stack_size: env::var("THREAD_STACK_SIZE")
                    .ok()
                    .and_then(|s| s.parse().ok()),
                tcp_nodelay: env::var("TCP_NODELAY")
                    .unwrap_or_else(|_| "true".to_string())
                    .parse()
                    .unwrap_or(true),
                tcp_keepalive: env::var("TCP_KEEPALIVE")
                    .ok()
                    .and_then(|s| s.parse().ok())
                    .map(Duration::from_secs),
                buffer_size: env::var("BUFFER_SIZE")
                    .unwrap_or_else(|_| "8192".to_string())
                    .parse()
                    .unwrap_or(8192),
                max_concurrent_streams: env::var("MAX_CONCURRENT_STREAMS")
                    .unwrap_or_else(|_| "1000".to_string())
                    .parse()
                    .unwrap_or(1000),
                stream_timeout: Duration::from_secs(
                    env::var("STREAM_TIMEOUT")
                        .unwrap_or_else(|_| "30".to_string())
                        .parse()
                        .unwrap_or(30)
                ),
                compression_level: env::var("COMPRESSION_LEVEL")
                    .unwrap_or_else(|_| "6".to_string())
                    .parse()
                    .unwrap_or(6),
            },

            monitoring: MonitoringConfig {
                metrics_enabled: env::var("METRICS_ENABLED")
                    .unwrap_or_else(|_| "true".to_string())
                    .parse()
                    .unwrap_or(true),
                metrics_port: env::var("METRICS_PORT")
                    .unwrap_or_else(|_| "9090".to_string())
                    .parse()
                    .unwrap_or(9090),
                health_check_interval: Duration::from_secs(
                    env::var("HEALTH_CHECK_INTERVAL")
                        .unwrap_or_else(|_| "30".to_string())
                        .parse()
                        .unwrap_or(30)
                ),
                log_level: env::var("LOG_LEVEL")
                    .unwrap_or_else(|_| "info".to_string()),
                log_format: match env::var("LOG_FORMAT")
                    .unwrap_or_else(|_| "pretty".to_string())
                    .to_lowercase()
                    .as_str()
                {
                    "json" => LogFormat::Json,
                    "compact" => LogFormat::Compact,
                    _ => LogFormat::Pretty,
                },
                jaeger_endpoint: env::var("JAEGER_ENDPOINT").ok(),
                prometheus_namespace: env::var("PROMETHEUS_NAMESPACE")
                    .unwrap_or_else(|_| "stream_server".to_string()),
                alert_webhooks: env::var("ALERT_WEBHOOKS")
                    .unwrap_or_default()
                    .split(',')
                    .filter(|s| !s.trim().is_empty())
                    .map(|s| s.trim().to_string())
                    .collect(),
            },

            notifications: NotificationConfig {
                enabled: env::var("NOTIFICATIONS_ENABLED")
                    .unwrap_or_else(|_| "true".to_string())
                    .parse()
                    .unwrap_or(true),
                max_queue_size: env::var("NOTIFICATIONS_MAX_QUEUE_SIZE")
                    .unwrap_or_else(|_| "10000".to_string())
                    .parse()
                    .unwrap_or(10000),
                delivery_workers: env::var("NOTIFICATIONS_DELIVERY_WORKERS")
                    .unwrap_or_else(|_| "4".to_string())
                    .parse()
                    .unwrap_or(4),
                retry_attempts: env::var("NOTIFICATIONS_RETRY_ATTEMPTS")
                    .unwrap_or_else(|_| "3".to_string())
                    .parse()
                    .unwrap_or(3),
                retry_delay: Duration::from_secs(
                    env::var("NOTIFICATIONS_RETRY_DELAY")
                        .unwrap_or_else(|_| "60".to_string())
                        .parse()
                        .unwrap_or(60)
                ),
                batch_size: env::var("NOTIFICATIONS_BATCH_SIZE")
                    .unwrap_or_else(|_| "100".to_string())
                    .parse()
                    .unwrap_or(100),
                email_provider: None, // Configuré séparément
                sms_provider: None,   // Configuré séparément
                push_provider: None,  // Configuré séparément
            },

            compression: CompressionConfig {
                enabled: env::var("COMPRESSION_ENABLED")
                    .unwrap_or_else(|_| "true".to_string())
                    .parse()
                    .unwrap_or(true),
                output_dir: env::var("COMPRESSION_OUTPUT_DIR")
                    .unwrap_or_else(|_| "./compressed".to_string()),
                temp_dir: env::var("COMPRESSION_TEMP_DIR")
                    .unwrap_or_else(|_| "./temp".to_string()),
                max_concurrent_jobs: env::var("COMPRESSION_MAX_CONCURRENT_JOBS")
                    .unwrap_or_else(|_| "4".to_string())
                    .parse()
                    .unwrap_or(4),
                cleanup_after_days: env::var("COMPRESSION_CLEANUP_AFTER_DAYS")
                    .unwrap_or_else(|_| "7".to_string())
                    .parse()
                    .unwrap_or(7),
                ffmpeg_path: env::var("FFMPEG_PATH").ok(),
                quality_profiles: env::var("COMPRESSION_QUALITY_PROFILES")
                    .unwrap_or_else(|_| "high,medium,low,mobile".to_string())
                    .split(',')
                    .map(|s| s.trim().to_string())
                    .collect(),
            },

            environment,
        };

        // Validation de la configuration
        config.validate()?;

        Ok(config)
    }

    pub fn validate(&self) -> Result<(), ConfigError> {
        // Validation du port
        if self.port == 0 || self.port > 65535 {
            return Err(ConfigError::InvalidPort);
        }

        // Validation du répertoire audio
        if self.audio_dir.is_empty() {
            return Err(ConfigError::InvalidAudioDir);
        }

        // Validation de la clé secrète en production
        if matches!(self.environment, Environment::Production) {
            if self.secret_key == "your-secret-key-change-in-production" {
                return Err(ConfigError::WeakSecretKey);
            }

            if self.security.jwt_secret.is_none() {
                return Err(ConfigError::MissingJwtSecret);
            }
        }

        // Validation des limites de fichiers
        if self.max_file_size == 0 {
            return Err(ConfigError::InvalidFileSize);
        }

        if self.max_range_size == 0 || self.max_range_size > self.max_file_size {
            return Err(ConfigError::InvalidRangeSize);
        }

        // Validation de la base de données
        if self.database.max_connections == 0 {
            return Err(ConfigError::InvalidDatabaseConfig);
        }

        Ok(())
    }

    pub fn is_production(&self) -> bool {
        matches!(self.environment, Environment::Production)
    }

    pub fn is_development(&self) -> bool {
        matches!(self.environment, Environment::Development)
    }

    pub fn redis_enabled(&self) -> bool {
        self.cache.redis_url.is_some()
    }

    pub fn tls_enabled(&self) -> bool {
        self.security.tls_cert_path.is_some() && self.security.tls_key_path.is_some()
    }

    pub fn metrics_enabled(&self) -> bool {
        self.monitoring.metrics_enabled
    }

    pub fn notifications_enabled(&self) -> bool {
        self.notifications.enabled
    }

    pub fn compression_enabled(&self) -> bool {
        self.compression.enabled
    }
}

#[derive(Debug, thiserror::Error)]
pub enum ConfigError {
    #[error("Port invalide")]
    InvalidPort,
    
    #[error("Répertoire audio invalide")]
    InvalidAudioDir,
    
    #[error("Clé secrète faible - changez-la en production")]
    WeakSecretKey,
    
    #[error("Clé JWT manquante en production")]
    MissingJwtSecret,
    
    #[error("Taille de fichier invalide")]
    InvalidFileSize,
    
    #[error("Taille de range invalide")]
    InvalidRangeSize,
    
    #[error("Tolérance de signature invalide")]
    InvalidSignatureTolerance,
    
    #[error("Configuration de base de données invalide")]
    InvalidDatabaseConfig,
} 