package redis_cache

import (
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/okinrev/veza-web-app/internal/infrastructure/config"
)

// NewClient crée une nouvelle connexion Redis
func NewClient(cfg config.RedisConfig) (*redis.Client, error) {
	if !cfg.Enabled {
		return nil, fmt.Errorf("Redis désactivé dans la configuration")
	}

	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.Database,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		// Configuration avancée pour la production
		MaxRetries:      3,
		MinRetryBackoff: 8 * time.Millisecond,
		MaxRetryBackoff: 512 * time.Millisecond,
		DialTimeout:     5 * time.Second,
		ReadTimeout:     3 * time.Second,
		WriteTimeout:    3 * time.Second,
		PoolTimeout:     4 * time.Second,
	})

	return client, nil
}
