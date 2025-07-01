// internal/config/config.go
package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Redis    RedisConfig
	NATS     NATSConfig
	Cache    CacheConfig
	Queue    QueueConfig
}

type ServerConfig struct {
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	Environment     string
}

type DatabaseConfig struct {
	URL          string
	Host         string
	Port         string
	Username     string
	Password     string
	Database     string
	SSLMode      string
	MaxOpenConns int
	MaxIdleConns int
	MaxLifetime  time.Duration
}

type JWTConfig struct {
	Secret          string
	ExpirationTime  time.Duration
	RefreshTime     time.Duration
	RefreshTTL      time.Duration
	RefreshRotation bool
}

type RedisConfig struct {
	URL          string
	Host         string
	Port         string
	Password     string
	Database     int
	MaxRetries   int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	PoolSize     int
	PoolTimeout  time.Duration
	IdleTimeout  time.Duration
	MaxConnAge   time.Duration
	EnableTLS    bool
}

type NATSConfig struct {
	URL                   string
	ClusterID             string
	ClientID              string
	MaxReconnects         int
	ReconnectWait         time.Duration
	ConnectTimeout        time.Duration
	MaxPendingMsgs        int
	MaxPendingBytes       int64
	EnableJetStream       bool
	StreamRetentionPolicy string
}

type CacheConfig struct {
	EnableLevel1     bool
	EnableLevel2     bool
	EnableLevel3     bool
	MaxMemoryMB      int
	CompressionLevel int
	StatsInterval    time.Duration
	WarmupEnabled    bool
	WarmupInterval   time.Duration
}

type QueueConfig struct {
	MaxWorkers          int
	MaxQueueSize        int
	ProcessingTimeout   time.Duration
	RetryMaxAttempts    int
	RetryBackoffBase    time.Duration
	DeadLetterQueueSize int
}

func New() *Config {
	// Récupérer DATABASE_URL depuis l'environnement
	databaseURL := getEnv("DATABASE_URL", "")
	if databaseURL == "" {
		// Construire l'URL si pas définie
		host := getEnv("DATABASE_HOST", "localhost")
		port := getEnv("DATABASE_PORT", "5432")
		username := getEnv("DATABASE_USER", "postgres")
		password := getEnv("DATABASE_PASSWORD", "")
		database := getEnv("DATABASE_NAME", "veza_dev")
		sslmode := "disable"

		databaseURL = "postgres://" + username + ":" + password + "@" + host + ":" + port + "/" + database + "?sslmode=" + sslmode
	}

	return &Config{
		Server: ServerConfig{
			Port:            getEnv("PORT", "8080"),
			ReadTimeout:     getDurationEnv("READ_TIMEOUT", 10*time.Second),
			WriteTimeout:    getDurationEnv("WRITE_TIMEOUT", 10*time.Second),
			ShutdownTimeout: getDurationEnv("SHUTDOWN_TIMEOUT", 30*time.Second),
			Environment:     getEnv("ENVIRONMENT", "development"),
		},
		Database: DatabaseConfig{
			URL:          databaseURL,
			Host:         getEnv("DATABASE_HOST", "localhost"),
			Port:         getEnv("DATABASE_PORT", "5432"),
			Username:     getEnv("DATABASE_USER", "postgres"),
			Password:     getEnv("DATABASE_PASSWORD", ""),
			Database:     getEnv("DATABASE_NAME", "veza_dev"),
			SSLMode:      "disable",
			MaxOpenConns: getIntEnv("DATABASE_MAX_OPEN_CONNS", 100), // Optimisé pour haute charge
			MaxIdleConns: getIntEnv("DATABASE_MAX_IDLE_CONNS", 25),
			MaxLifetime:  getDurationEnv("DATABASE_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		JWT: JWTConfig{
			Secret:          getEnv("JWT_ACCESS_SECRET", "your-super-secret-key-change-in-production"),
			ExpirationTime:  getDurationEnv("JWT_ACCESS_TTL", 15*time.Minute),
			RefreshTime:     getDurationEnv("JWT_REFRESH_TTL", 7*24*time.Hour),
			RefreshTTL:      getDurationEnv("JWT_REFRESH_TTL", 7*24*time.Hour),
			RefreshRotation: getBoolEnv("JWT_REFRESH_ROTATION", true),
		},
		Redis: RedisConfig{
			URL:          getEnv("REDIS_URL", ""),
			Host:         getEnv("REDIS_HOST", "localhost"),
			Port:         getEnv("REDIS_PORT", "6379"),
			Password:     getEnv("REDIS_PASSWORD", ""),
			Database:     getIntEnv("REDIS_DATABASE", 0),
			MaxRetries:   getIntEnv("REDIS_MAX_RETRIES", 3),
			DialTimeout:  getDurationEnv("REDIS_DIAL_TIMEOUT", 5*time.Second),
			ReadTimeout:  getDurationEnv("REDIS_READ_TIMEOUT", 3*time.Second),
			WriteTimeout: getDurationEnv("REDIS_WRITE_TIMEOUT", 3*time.Second),
			PoolSize:     getIntEnv("REDIS_POOL_SIZE", 100), // Optimisé pour haute charge
			PoolTimeout:  getDurationEnv("REDIS_POOL_TIMEOUT", 5*time.Second),
			IdleTimeout:  getDurationEnv("REDIS_IDLE_TIMEOUT", 5*time.Minute),
			MaxConnAge:   getDurationEnv("REDIS_MAX_CONN_AGE", 10*time.Minute),
			EnableTLS:    getBoolEnv("REDIS_ENABLE_TLS", false),
		},
		NATS: NATSConfig{
			URL:                   getEnv("NATS_URL", "nats://localhost:4222"),
			ClusterID:             getEnv("NATS_CLUSTER_ID", "veza-cluster"),
			ClientID:              getEnv("NATS_CLIENT_ID", "veza-backend"),
			MaxReconnects:         getIntEnv("NATS_MAX_RECONNECTS", 10),
			ReconnectWait:         getDurationEnv("NATS_RECONNECT_WAIT", 2*time.Second),
			ConnectTimeout:        getDurationEnv("NATS_CONNECT_TIMEOUT", 5*time.Second),
			MaxPendingMsgs:        getIntEnv("NATS_MAX_PENDING_MSGS", 10000),
			MaxPendingBytes:       getInt64Env("NATS_MAX_PENDING_BYTES", 64*1024*1024), // 64MB
			EnableJetStream:       getBoolEnv("NATS_ENABLE_JETSTREAM", true),
			StreamRetentionPolicy: getEnv("NATS_STREAM_RETENTION", "limits"),
		},
		Cache: CacheConfig{
			EnableLevel1:     getBoolEnv("CACHE_ENABLE_L1", true),
			EnableLevel2:     getBoolEnv("CACHE_ENABLE_L2", true),
			EnableLevel3:     getBoolEnv("CACHE_ENABLE_L3", false),
			MaxMemoryMB:      getIntEnv("CACHE_MAX_MEMORY_MB", 512),
			CompressionLevel: getIntEnv("CACHE_COMPRESSION_LEVEL", 1),
			StatsInterval:    getDurationEnv("CACHE_STATS_INTERVAL", 30*time.Second),
			WarmupEnabled:    getBoolEnv("CACHE_WARMUP_ENABLED", true),
			WarmupInterval:   getDurationEnv("CACHE_WARMUP_INTERVAL", 10*time.Minute),
		},
		Queue: QueueConfig{
			MaxWorkers:          getIntEnv("QUEUE_MAX_WORKERS", 50), // Optimisé pour haute charge
			MaxQueueSize:        getIntEnv("QUEUE_MAX_SIZE", 10000), // 10k éléments max
			ProcessingTimeout:   getDurationEnv("QUEUE_PROCESSING_TIMEOUT", 30*time.Second),
			RetryMaxAttempts:    getIntEnv("QUEUE_RETRY_MAX_ATTEMPTS", 3),
			RetryBackoffBase:    getDurationEnv("QUEUE_RETRY_BACKOFF_BASE", 1*time.Second),
			DeadLetterQueueSize: getIntEnv("QUEUE_DLQ_SIZE", 1000),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getInt64Env(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}
