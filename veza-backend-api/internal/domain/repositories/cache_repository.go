package repositories

import (
	"context"
	"time"
)

// CacheRepository définit l'interface pour les opérations de cache
type CacheRepository interface {
	// Opérations de base
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)

	// Opérations avancées
	TTL(ctx context.Context, key string) (time.Duration, error)
	Expire(ctx context.Context, key string, ttl time.Duration) error
	Increment(ctx context.Context, key string, value int64) (int64, error)

	// Opérations de patterns
	Keys(ctx context.Context, pattern string) ([]string, error)
	DeletePattern(ctx context.Context, pattern string) error

	// Opérations de pipeline
	Pipeline(ctx context.Context, operations []CacheOperation) error
}

// CacheOperation représente une opération de cache pour les pipelines
type CacheOperation struct {
	Type  string // "set", "get", "delete", etc.
	Key   string
	Value interface{}
	TTL   time.Duration
}
