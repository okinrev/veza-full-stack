package redis_cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	"github.com/okinrev/veza-web-app/internal/domain/repositories"
)

// cacheRepository implémentation Redis du CacheRepository
type cacheRepository struct {
	client     *redis.Client
	logger     *zap.Logger
	defaultTTL time.Duration
}

// NewCacheRepository crée une nouvelle instance du repository de cache Redis
func NewCacheRepository(client *redis.Client, defaultTTL time.Duration, logger *zap.Logger) repositories.CacheRepository {
	return &cacheRepository{
		client:     client,
		logger:     logger,
		defaultTTL: defaultTTL,
	}
}

// Get récupère une valeur depuis le cache et la désérialise dans dest
func (c *cacheRepository) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("key not found: %s", key)
		}
		return fmt.Errorf("failed to get key %s: %w", key, err)
	}

	// Désérialiser en JSON dans dest
	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return fmt.Errorf("failed to unmarshal value for key %s: %w", key, err)
	}

	return nil
}

// Set stocke une valeur dans le cache avec sérialisation JSON
func (c *cacheRepository) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if ttl == 0 {
		ttl = c.defaultTTL
	}

	// Sérialiser en JSON
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value for key %s: %w", key, err)
	}

	if err := c.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set key %s: %w", key, err)
	}

	return nil
}

// Delete supprime une clé du cache
func (c *cacheRepository) Delete(ctx context.Context, key string) error {
	result := c.client.Del(ctx, key)
	if err := result.Err(); err != nil {
		return fmt.Errorf("failed to delete key %s: %w", key, err)
	}
	return nil
}

// Exists vérifie si une clé existe dans le cache
func (c *cacheRepository) Exists(ctx context.Context, key string) (bool, error) {
	count, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check existence of key %s: %w", key, err)
	}
	return count > 0, nil
}

// TTL retourne le temps de vie restant d'une clé
func (c *cacheRepository) TTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := c.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get TTL for key %s: %w", key, err)
	}
	return ttl, nil
}

// Expire définit un nouveau TTL sur une clé existante
func (c *cacheRepository) Expire(ctx context.Context, key string, ttl time.Duration) error {
	if err := c.client.Expire(ctx, key, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set TTL for key %s: %w", key, err)
	}
	return nil
}

// Increment incrémente une valeur numérique
func (c *cacheRepository) Increment(ctx context.Context, key string, value int64) (int64, error) {
	result, err := c.client.IncrBy(ctx, key, value).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment key %s: %w", key, err)
	}
	return result, nil
}

// Keys retourne toutes les clés correspondant au pattern
func (c *cacheRepository) Keys(ctx context.Context, pattern string) ([]string, error) {
	keys, err := c.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get keys for pattern %s: %w", pattern, err)
	}
	return keys, nil
}

// DeletePattern supprime toutes les clés correspondant au pattern
func (c *cacheRepository) DeletePattern(ctx context.Context, pattern string) error {
	// Récupérer toutes les clés correspondant au pattern
	keys, err := c.Keys(ctx, pattern)
	if err != nil {
		return fmt.Errorf("failed to get keys for pattern %s: %w", pattern, err)
	}

	if len(keys) == 0 {
		return nil // Aucune clé à supprimer
	}

	// Supprimer toutes les clés en une seule commande
	if err := c.client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("failed to delete keys for pattern %s: %w", pattern, err)
	}

	if c.logger != nil {
		c.logger.Info("Deleted keys by pattern",
			zap.String("pattern", pattern),
			zap.Int("count", len(keys)),
		)
	}

	return nil
}

// Pipeline exécute une série d'opérations en pipeline
func (c *cacheRepository) Pipeline(ctx context.Context, operations []repositories.CacheOperation) error {
	pipe := c.client.Pipeline()

	// Ajouter toutes les opérations au pipeline
	for _, op := range operations {
		switch op.Type {
		case "set":
			data, err := json.Marshal(op.Value)
			if err != nil {
				return fmt.Errorf("failed to marshal value for pipeline operation: %w", err)
			}
			pipe.Set(ctx, op.Key, data, op.TTL)

		case "delete":
			pipe.Del(ctx, op.Key)

		case "increment":
			if value, ok := op.Value.(int64); ok {
				pipe.IncrBy(ctx, op.Key, value)
			} else {
				return fmt.Errorf("invalid value type for increment operation: %T", op.Value)
			}

		case "expire":
			pipe.Expire(ctx, op.Key, op.TTL)

		default:
			return fmt.Errorf("unsupported pipeline operation type: %s", op.Type)
		}
	}

	// Exécuter le pipeline
	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return fmt.Errorf("failed to execute pipeline: %w", err)
	}

	if c.logger != nil {
		c.logger.Debug("Executed cache pipeline",
			zap.Int("operations", len(operations)),
		)
	}

	return nil
}
