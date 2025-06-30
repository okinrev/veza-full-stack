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
	Secret         string
	ExpirationTime time.Duration
	RefreshTime    time.Duration
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
			MaxOpenConns: getIntEnv("DATABASE_MAX_OPEN_CONNS", 25),
			MaxIdleConns: getIntEnv("DATABASE_MAX_IDLE_CONNS", 25),
			MaxLifetime:  getDurationEnv("DATABASE_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		JWT: JWTConfig{
			Secret:         getEnv("JWT_ACCESS_SECRET", "your-super-secret-key-change-in-production"),
			ExpirationTime: getDurationEnv("JWT_ACCESS_TTL", 15*time.Minute),
			RefreshTime:    getDurationEnv("JWT_REFRESH_TTL", 7*24*time.Hour),
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
