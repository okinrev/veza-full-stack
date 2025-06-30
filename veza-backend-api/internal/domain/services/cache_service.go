package services

import (
	"context"
	"time"
)

// CacheService interface pour les opérations de cache
type CacheService interface {
	// Opérations basiques
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)

	// Opérations sur les listes
	ListPush(ctx context.Context, key string, values ...interface{}) error
	ListPop(ctx context.Context, key string) (interface{}, error)
	ListRange(ctx context.Context, key string, start, stop int64) ([]interface{}, error)
	ListLength(ctx context.Context, key string) (int64, error)

	// Opérations sur les sets
	SetAdd(ctx context.Context, key string, members ...interface{}) error
	SetRemove(ctx context.Context, key string, members ...interface{}) error
	SetMembers(ctx context.Context, key string) ([]interface{}, error)
	SetIsMember(ctx context.Context, key string, member interface{}) (bool, error)

	// Opérations sur les hash maps
	HashSet(ctx context.Context, key, field string, value interface{}) error
	HashGet(ctx context.Context, key, field string) (interface{}, error)
	HashGetAll(ctx context.Context, key string) (map[string]interface{}, error)
	HashDelete(ctx context.Context, key string, fields ...string) error

	// Opérations avancées
	Increment(ctx context.Context, key string) (int64, error)
	IncrementBy(ctx context.Context, key string, value int64) (int64, error)
	Expire(ctx context.Context, key string, ttl time.Duration) error
	TTL(ctx context.Context, key string) (time.Duration, error)

	// Opérations de batch
	Pipeline() Pipeline

	// Gestion des patterns
	Keys(ctx context.Context, pattern string) ([]string, error)
	DeletePattern(ctx context.Context, pattern string) error

	// Monitoring
	Ping(ctx context.Context) error
	Info(ctx context.Context) (map[string]string, error)
	FlushDB(ctx context.Context) error
}

// Pipeline interface pour les opérations en batch
type Pipeline interface {
	Get(key string) *PipelineResult
	Set(key string, value interface{}, ttl time.Duration) *PipelineResult
	Delete(key string) *PipelineResult
	Increment(key string) *PipelineResult
	Execute(ctx context.Context) ([]*PipelineResult, error)
}

// PipelineResult résultat d'une opération pipeline
type PipelineResult struct {
	Error error
	Value interface{}
}
