package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config structure principale de configuration
type Config struct {
	// Serveur
	Server ServerConfig `json:"server"`

	// Base de données
	Database DatabaseConfig `json:"database"`

	// Redis
	Redis RedisConfig `json:"redis"`

	// JWT
	JWT JWTConfig `json:"jwt"`

	// OAuth2
	OAuth OAuth2Config `json:"oauth"`

	// Email
	Email EmailConfig `json:"email"`

	// Rate Limiting
	RateLimit RateLimitConfig `json:"rate_limit"`

	// Security
	Security SecurityConfig `json:"security"`

	// Observabilité
	Observability ObservabilityConfig `json:"observability"`

	// gRPC
	GRPC GRPCConfig `json:"grpc"`
}

// ServerConfig configuration du serveur HTTP
type ServerConfig struct {
	Port         int           `json:"port"`
	Host         string        `json:"host"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
	Environment  string        `json:"environment"`
}

// DatabaseConfig configuration PostgreSQL
type DatabaseConfig struct {
	Host            string        `json:"host"`
	Port            int           `json:"port"`
	User            string        `json:"user"`
	Password        string        `json:"password"`
	DBName          string        `json:"db_name"`
	SSLMode         string        `json:"ssl_mode"`
	MaxOpenConns    int           `json:"max_open_conns"`
	MaxIdleConns    int           `json:"max_idle_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `json:"conn_max_idle_time"`
}

// RedisConfig configuration Redis
type RedisConfig struct {
	Host          string        `json:"host"`
	Port          int           `json:"port"`
	Password      string        `json:"password"`
	DB            int           `json:"db"`
	PoolSize      int           `json:"pool_size"`
	MinIdleConns  int           `json:"min_idle_conns"`
	MaxConnAge    time.Duration `json:"max_conn_age"`
	PoolTimeout   time.Duration `json:"pool_timeout"`
	IdleTimeout   time.Duration `json:"idle_timeout"`
	IdleCheckFreq time.Duration `json:"idle_check_freq"`
}

// JWTConfig configuration JWT
type JWTConfig struct {
	Secret        string        `json:"-"` // Ne pas exposer dans les logs
	Expiry        time.Duration `json:"expiry"`
	RefreshExpiry time.Duration `json:"refresh_expiry"`
	Issuer        string        `json:"issuer"`

	// Rotation des secrets
	RotationEnabled    bool          `json:"rotation_enabled"`
	RotationInterval   time.Duration `json:"rotation_interval"`
	SecretHistorySize  int           `json:"secret_history_size"`
	CurrentSecretIndex int           `json:"current_secret_index"`
	SecretHistory      []string      `json:"-"` // Historique des secrets
}

// OAuth2Config configuration OAuth2
type OAuth2Config struct {
	Google  OAuthProviderConfig `json:"google"`
	GitHub  OAuthProviderConfig `json:"github"`
	Discord OAuthProviderConfig `json:"discord"`
}

// OAuthProviderConfig configuration pour un provider OAuth2
type OAuthProviderConfig struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"-"` // Ne pas exposer dans les logs
	RedirectURL  string `json:"redirect_url"`
	Enabled      bool   `json:"enabled"`
}

// EmailConfig configuration email (SMTP)
type EmailConfig struct {
	SMTPHost     string `json:"smtp_host"`
	SMTPPort     int    `json:"smtp_port"`
	SMTPUser     string `json:"smtp_user"`
	SMTPPassword string `json:"-"` // Ne pas exposer dans les logs
	FromEmail    string `json:"from_email"`
	FromName     string `json:"from_name"`
	UseTLS       bool   `json:"use_tls"`
}

// RateLimitConfig configuration rate limiting
type RateLimitConfig struct {
	Enabled         bool          `json:"enabled"`
	RequestsPerMin  int           `json:"requests_per_min"`
	BurstSize       int           `json:"burst_size"`
	CleanupInterval time.Duration `json:"cleanup_interval"`
	BanDuration     time.Duration `json:"ban_duration"`

	// Rate limiting par endpoint
	Endpoints map[string]EndpointRateLimit `json:"endpoints"`
}

// EndpointRateLimit configuration rate limiting spécifique par endpoint
type EndpointRateLimit struct {
	RequestsPerMin int `json:"requests_per_min"`
	BurstSize      int `json:"burst_size"`
}

// SecurityConfig configuration sécurité
type SecurityConfig struct {
	// CORS
	CORS CORSConfig `json:"cors"`

	// Headers de sécurité
	Headers SecurityHeaders `json:"headers"`

	// 2FA
	TwoFactor TwoFactorConfig `json:"two_factor"`

	// Session
	Session SessionConfig `json:"session"`

	// Audit
	Audit AuditConfig `json:"audit"`
}

// CORSConfig configuration CORS
type CORSConfig struct {
	AllowedOrigins   []string `json:"allowed_origins"`
	AllowedMethods   []string `json:"allowed_methods"`
	AllowedHeaders   []string `json:"allowed_headers"`
	ExposedHeaders   []string `json:"exposed_headers"`
	AllowCredentials bool     `json:"allow_credentials"`
	MaxAge           int      `json:"max_age"`
}

// SecurityHeaders headers de sécurité
type SecurityHeaders struct {
	ContentSecurityPolicy   string `json:"content_security_policy"`
	StrictTransportSecurity string `json:"strict_transport_security"`
	XFrameOptions           string `json:"x_frame_options"`
	XContentTypeOptions     string `json:"x_content_type_options"`
	ReferrerPolicy          string `json:"referrer_policy"`
}

// TwoFactorConfig configuration 2FA
type TwoFactorConfig struct {
	Issuer         string        `json:"issuer"`
	WindowSize     int           `json:"window_size"`
	RecoveryLength int           `json:"recovery_length"`
	ValidityPeriod time.Duration `json:"validity_period"`
}

// SessionConfig configuration sessions
type SessionConfig struct {
	CookieName     string        `json:"cookie_name"`
	CookieDomain   string        `json:"cookie_domain"`
	CookiePath     string        `json:"cookie_path"`
	CookieSecure   bool          `json:"cookie_secure"`
	CookieHTTPOnly bool          `json:"cookie_http_only"`
	CookieSameSite string        `json:"cookie_same_site"`
	MaxSessions    int           `json:"max_sessions"`
	IdleTimeout    time.Duration `json:"idle_timeout"`
}

// AuditConfig configuration audit logging
type AuditConfig struct {
	Enabled         bool     `json:"enabled"`
	RetentionDays   int      `json:"retention_days"`
	SensitiveFields []string `json:"sensitive_fields"`
	LogLevel        string   `json:"log_level"`
}

// ObservabilityConfig configuration observabilité
type ObservabilityConfig struct {
	Metrics MetricsConfig `json:"metrics"`
	Tracing TracingConfig `json:"tracing"`
	Logging LoggingConfig `json:"logging"`
}

// MetricsConfig configuration métriques Prometheus
type MetricsConfig struct {
	Enabled   bool   `json:"enabled"`
	Path      string `json:"path"`
	Port      int    `json:"port"`
	Namespace string `json:"namespace"`
	Subsystem string `json:"subsystem"`
}

// TracingConfig configuration tracing distribué
type TracingConfig struct {
	Enabled     bool    `json:"enabled"`
	ServiceName string  `json:"service_name"`
	Endpoint    string  `json:"endpoint"`
	SampleRate  float64 `json:"sample_rate"`
}

// LoggingConfig configuration logging
type LoggingConfig struct {
	Level      string `json:"level"`
	Format     string `json:"format"` // json, console
	Output     string `json:"output"` // stdout, stderr, file
	FilePath   string `json:"file_path"`
	MaxSize    int    `json:"max_size"` // MB
	MaxBackups int    `json:"max_backups"`
	MaxAge     int    `json:"max_age"` // days
	Compress   bool   `json:"compress"`
}

// GRPCConfig configuration gRPC
type GRPCConfig struct {
	Enabled bool   `json:"enabled"`
	Host    string `json:"host"`
	Port    int    `json:"port"`

	// TLS
	TLS GRPCTLSConfig `json:"tls"`

	// Services
	Services GRPCServicesConfig `json:"services"`
}

// GRPCTLSConfig configuration TLS pour gRPC
type GRPCTLSConfig struct {
	Enabled  bool   `json:"enabled"`
	CertFile string `json:"cert_file"`
	KeyFile  string `json:"key_file"`
	CAFile   string `json:"ca_file"`
}

// GRPCServicesConfig configuration services gRPC
type GRPCServicesConfig struct {
	Chat   GRPCServiceConfig `json:"chat"`
	Stream GRPCServiceConfig `json:"stream"`
}

// GRPCServiceConfig configuration d'un service gRPC
type GRPCServiceConfig struct {
	Enabled bool   `json:"enabled"`
	Host    string `json:"host"`
	Port    int    `json:"port"`
}

// LoadConfig charge la configuration depuis les variables d'environnement
func LoadConfig() (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			Port:         getEnvAsInt("SERVER_PORT", 8080),
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			ReadTimeout:  getEnvAsDuration("SERVER_READ_TIMEOUT", "15s"),
			WriteTimeout: getEnvAsDuration("SERVER_WRITE_TIMEOUT", "15s"),
			IdleTimeout:  getEnvAsDuration("SERVER_IDLE_TIMEOUT", "60s"),
			Environment:  getEnv("ENVIRONMENT", "development"),
		},

		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnvAsInt("DB_PORT", 5432),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", "password"),
			DBName:          getEnv("DB_NAME", "veza_dev"),
			SSLMode:         getEnv("DB_SSL_MODE", "disable"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 100),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", "1h"),
			ConnMaxIdleTime: getEnvAsDuration("DB_CONN_MAX_IDLE_TIME", "30m"),
		},

		Redis: RedisConfig{
			Host:          getEnv("REDIS_HOST", "localhost"),
			Port:          getEnvAsInt("REDIS_PORT", 6379),
			Password:      getEnv("REDIS_PASSWORD", ""),
			DB:            getEnvAsInt("REDIS_DB", 0),
			PoolSize:      getEnvAsInt("REDIS_POOL_SIZE", 20),
			MinIdleConns:  getEnvAsInt("REDIS_MIN_IDLE_CONNS", 5),
			MaxConnAge:    getEnvAsDuration("REDIS_MAX_CONN_AGE", "30m"),
			PoolTimeout:   getEnvAsDuration("REDIS_POOL_TIMEOUT", "4s"),
			IdleTimeout:   getEnvAsDuration("REDIS_IDLE_TIMEOUT", "5m"),
			IdleCheckFreq: getEnvAsDuration("REDIS_IDLE_CHECK_FREQ", "1m"),
		},

		JWT: JWTConfig{
			Secret:        getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-in-production"),
			Expiry:        getEnvAsDuration("JWT_EXPIRY", "15m"),
			RefreshExpiry: getEnvAsDuration("JWT_REFRESH_EXPIRY", "7d"),
			Issuer:        getEnv("JWT_ISSUER", "veza-api"),

			// Configuration rotation des secrets
			RotationEnabled:    getEnvAsBool("JWT_ROTATION_ENABLED", true),
			RotationInterval:   getEnvAsDuration("JWT_ROTATION_INTERVAL", "24h"),
			SecretHistorySize:  getEnvAsInt("JWT_SECRET_HISTORY_SIZE", 5),
			CurrentSecretIndex: 0,
			SecretHistory:      []string{},
		},

		OAuth: OAuth2Config{
			Google: OAuthProviderConfig{
				ClientID:     getEnv("OAUTH_GOOGLE_CLIENT_ID", ""),
				ClientSecret: getEnv("OAUTH_GOOGLE_CLIENT_SECRET", ""),
				RedirectURL:  getEnv("OAUTH_GOOGLE_REDIRECT_URL", ""),
				Enabled:      getEnvAsBool("OAUTH_GOOGLE_ENABLED", false),
			},
			GitHub: OAuthProviderConfig{
				ClientID:     getEnv("OAUTH_GITHUB_CLIENT_ID", ""),
				ClientSecret: getEnv("OAUTH_GITHUB_CLIENT_SECRET", ""),
				RedirectURL:  getEnv("OAUTH_GITHUB_REDIRECT_URL", ""),
				Enabled:      getEnvAsBool("OAUTH_GITHUB_ENABLED", false),
			},
			Discord: OAuthProviderConfig{
				ClientID:     getEnv("OAUTH_DISCORD_CLIENT_ID", ""),
				ClientSecret: getEnv("OAUTH_DISCORD_CLIENT_SECRET", ""),
				RedirectURL:  getEnv("OAUTH_DISCORD_REDIRECT_URL", ""),
				Enabled:      getEnvAsBool("OAUTH_DISCORD_ENABLED", false),
			},
		},

		Email: EmailConfig{
			SMTPHost:     getEnv("SMTP_HOST", "localhost"),
			SMTPPort:     getEnvAsInt("SMTP_PORT", 587),
			SMTPUser:     getEnv("SMTP_USER", ""),
			SMTPPassword: getEnv("SMTP_PASSWORD", ""),
			FromEmail:    getEnv("EMAIL_FROM", "noreply@veza.dev"),
			FromName:     getEnv("EMAIL_FROM_NAME", "Veza"),
			UseTLS:       getEnvAsBool("SMTP_USE_TLS", true),
		},

		RateLimit: RateLimitConfig{
			Enabled:         getEnvAsBool("RATE_LIMIT_ENABLED", true),
			RequestsPerMin:  getEnvAsInt("RATE_LIMIT_REQUESTS_PER_MIN", 60),
			BurstSize:       getEnvAsInt("RATE_LIMIT_BURST_SIZE", 20),
			CleanupInterval: getEnvAsDuration("RATE_LIMIT_CLEANUP_INTERVAL", "5m"),
			BanDuration:     getEnvAsDuration("RATE_LIMIT_BAN_DURATION", "1h"),
			Endpoints:       make(map[string]EndpointRateLimit),
		},

		Security: SecurityConfig{
			CORS: CORSConfig{
				AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:5173"},
				AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
				AllowedHeaders:   []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
				ExposedHeaders:   []string{"X-Total-Count", "X-Page-Count"},
				AllowCredentials: true,
				MaxAge:           300,
			},
			Headers: SecurityHeaders{
				ContentSecurityPolicy:   "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'",
				StrictTransportSecurity: "max-age=31536000; includeSubDomains",
				XFrameOptions:           "DENY",
				XContentTypeOptions:     "nosniff",
				ReferrerPolicy:          "strict-origin-when-cross-origin",
			},
			TwoFactor: TwoFactorConfig{
				Issuer:         "Veza",
				WindowSize:     1,
				RecoveryLength: 8,
				ValidityPeriod: getEnvAsDuration("2FA_VALIDITY_PERIOD", "5m"),
			},
			Session: SessionConfig{
				CookieName:     "veza_session",
				CookieDomain:   getEnv("SESSION_COOKIE_DOMAIN", ""),
				CookiePath:     "/",
				CookieSecure:   getEnvAsBool("SESSION_COOKIE_SECURE", false),
				CookieHTTPOnly: true,
				CookieSameSite: "lax",
				MaxSessions:    getEnvAsInt("SESSION_MAX_SESSIONS", 5),
				IdleTimeout:    getEnvAsDuration("SESSION_IDLE_TIMEOUT", "30m"),
			},
			Audit: AuditConfig{
				Enabled:         getEnvAsBool("AUDIT_ENABLED", true),
				RetentionDays:   getEnvAsInt("AUDIT_RETENTION_DAYS", 90),
				SensitiveFields: []string{"password", "token", "secret"},
				LogLevel:        getEnv("AUDIT_LOG_LEVEL", "info"),
			},
		},

		Observability: ObservabilityConfig{
			Metrics: MetricsConfig{
				Enabled:   getEnvAsBool("METRICS_ENABLED", true),
				Path:      getEnv("METRICS_PATH", "/metrics"),
				Port:      getEnvAsInt("METRICS_PORT", 9090),
				Namespace: getEnv("METRICS_NAMESPACE", "veza"),
				Subsystem: getEnv("METRICS_SUBSYSTEM", "api"),
			},
			Tracing: TracingConfig{
				Enabled:     getEnvAsBool("TRACING_ENABLED", false),
				ServiceName: getEnv("TRACING_SERVICE_NAME", "veza-api"),
				Endpoint:    getEnv("TRACING_ENDPOINT", ""),
				SampleRate:  getEnvAsFloat("TRACING_SAMPLE_RATE", 0.1),
			},
			Logging: LoggingConfig{
				Level:      getEnv("LOG_LEVEL", "info"),
				Format:     getEnv("LOG_FORMAT", "json"),
				Output:     getEnv("LOG_OUTPUT", "stdout"),
				FilePath:   getEnv("LOG_FILE_PATH", "logs/app.log"),
				MaxSize:    getEnvAsInt("LOG_MAX_SIZE", 100),
				MaxBackups: getEnvAsInt("LOG_MAX_BACKUPS", 3),
				MaxAge:     getEnvAsInt("LOG_MAX_AGE", 28),
				Compress:   getEnvAsBool("LOG_COMPRESS", true),
			},
		},

		GRPC: GRPCConfig{
			Enabled: getEnvAsBool("GRPC_ENABLED", true),
			Host:    getEnv("GRPC_HOST", "0.0.0.0"),
			Port:    getEnvAsInt("GRPC_PORT", 9000),
			TLS: GRPCTLSConfig{
				Enabled:  getEnvAsBool("GRPC_TLS_ENABLED", false),
				CertFile: getEnv("GRPC_TLS_CERT_FILE", ""),
				KeyFile:  getEnv("GRPC_TLS_KEY_FILE", ""),
				CAFile:   getEnv("GRPC_TLS_CA_FILE", ""),
			},
			Services: GRPCServicesConfig{
				Chat: GRPCServiceConfig{
					Enabled: getEnvAsBool("GRPC_CHAT_ENABLED", true),
					Host:    getEnv("GRPC_CHAT_HOST", "localhost"),
					Port:    getEnvAsInt("GRPC_CHAT_PORT", 9001),
				},
				Stream: GRPCServiceConfig{
					Enabled: getEnvAsBool("GRPC_STREAM_ENABLED", true),
					Host:    getEnv("GRPC_STREAM_HOST", "localhost"),
					Port:    getEnvAsInt("GRPC_STREAM_PORT", 9002),
				},
			},
		},
	}

	// Configuration des endpoints rate limit spécifiques
	config.RateLimit.Endpoints = map[string]EndpointRateLimit{
		"/auth/login":    {RequestsPerMin: 5, BurstSize: 2},
		"/auth/register": {RequestsPerMin: 3, BurstSize: 1},
		"/auth/reset":    {RequestsPerMin: 2, BurstSize: 1},
	}

	return config, nil
}

// Validate valide la configuration
func (c *Config) Validate() error {
	// Validation JWT
	if c.JWT.Secret == "" || c.JWT.Secret == "your-super-secret-jwt-key-change-in-production" {
		if c.Server.Environment == "production" {
			return fmt.Errorf("JWT secret must be set in production")
		}
	}

	// Validation rotation JWT
	if c.JWT.RotationEnabled {
		if c.JWT.RotationInterval < time.Hour {
			return fmt.Errorf("JWT rotation interval must be at least 1 hour")
		}
		if c.JWT.SecretHistorySize < 2 {
			return fmt.Errorf("JWT secret history size must be at least 2")
		}
		if c.JWT.SecretHistorySize > 10 {
			return fmt.Errorf("JWT secret history size must not exceed 10")
		}
	}

	if c.Database.Password == "" {
		return fmt.Errorf("database password must be set")
	}

	if c.Server.Environment == "production" {
		if !c.Security.Session.CookieSecure {
			return fmt.Errorf("cookies must be secure in production")
		}

		if c.Security.Headers.StrictTransportSecurity == "" {
			return fmt.Errorf("HSTS header must be set in production")
		}
	}

	return nil
}

// IsDevelopment vérifie si on est en mode développement
func (c *Config) IsDevelopment() bool {
	return c.Server.Environment == "development"
}

// IsProduction vérifie si on est en mode production
func (c *Config) IsProduction() bool {
	return c.Server.Environment == "production"
}

// GetDatabaseURL retourne l'URL de connexion PostgreSQL
func (c *Config) GetDatabaseURL() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.DBName,
		c.Database.SSLMode,
	)
}

// GetRedisAddr retourne l'adresse Redis
func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port)
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// getEnv récupère une variable d'environnement avec une valeur par défaut
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt récupère une variable d'environnement comme entier
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsBool récupère une variable d'environnement comme booléen
func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// getEnvAsFloat récupère une variable d'environnement comme float64
func getEnvAsFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

// getEnvAsDuration récupère une variable d'environnement comme duration
func getEnvAsDuration(key, defaultValue string) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	if duration, err := time.ParseDuration(defaultValue); err == nil {
		return duration
	}
	return 0
}
