package recommendations

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// RedisRecommendationCache implémentation du cache avec Redis
type RedisRecommendationCache struct {
	client *redis.Client
	logger *zap.Logger
}

// NewRedisRecommendationCache crée un nouveau cache Redis
func NewRedisRecommendationCache(client *redis.Client, logger *zap.Logger) *RedisRecommendationCache {
	return &RedisRecommendationCache{
		client: client,
		logger: logger,
	}
}

// GetRecommendations récupère les recommandations depuis le cache
func (c *RedisRecommendationCache) GetRecommendations(ctx context.Context, userID int64, context string) (*RecommendationResponse, error) {
	key := c.generateCacheKey(userID, context)

	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			// Pas de données en cache
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get from cache: %w", err)
	}

	var recommendations RecommendationResponse
	if err := json.Unmarshal([]byte(data), &recommendations); err != nil {
		c.logger.Warn("Failed to unmarshal cached recommendations", zap.Error(err))
		return nil, fmt.Errorf("failed to unmarshal cached data: %w", err)
	}

	c.logger.Debug("📦 Retrieved recommendations from cache",
		zap.Int64("user_id", userID),
		zap.String("context", context),
	)

	return &recommendations, nil
}

// SetRecommendations met en cache les recommandations
func (c *RedisRecommendationCache) SetRecommendations(ctx context.Context, userID int64, context string, recommendations *RecommendationResponse, ttl time.Duration) error {
	key := c.generateCacheKey(userID, context)

	data, err := json.Marshal(recommendations)
	if err != nil {
		return fmt.Errorf("failed to marshal recommendations: %w", err)
	}

	if err := c.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	c.logger.Debug("💾 Cached recommendations",
		zap.Int64("user_id", userID),
		zap.String("context", context),
		zap.Duration("ttl", ttl),
	)

	return nil
}

// InvalidateUserRecommendations invalide toutes les recommandations d'un utilisateur
func (c *RedisRecommendationCache) InvalidateUserRecommendations(ctx context.Context, userID int64) error {
	pattern := c.generateCacheKey(userID, "*")

	keys, err := c.client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get cache keys: %w", err)
	}

	if len(keys) > 0 {
		if err := c.client.Del(ctx, keys...).Err(); err != nil {
			return fmt.Errorf("failed to delete cache keys: %w", err)
		}

		c.logger.Info("🗑️ Invalidated user recommendations cache",
			zap.Int64("user_id", userID),
			zap.Int("keys_deleted", len(keys)),
		)
	}

	return nil
}

// generateCacheKey génère une clé de cache
func (c *RedisRecommendationCache) generateCacheKey(userID int64, context string) string {
	return fmt.Sprintf("recommendations:%d:%s", userID, context)
}

// GetCacheStats récupère les statistiques du cache
func (c *RedisRecommendationCache) GetCacheStats(ctx context.Context) (map[string]interface{}, error) {
	info, err := c.client.Info(ctx, "memory").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get Redis info: %w", err)
	}

	// Compter les clés de recommandations
	pattern := "recommendations:*"
	keys, err := c.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to count recommendation keys: %w", err)
	}

	stats := map[string]interface{}{
		"total_recommendation_keys": len(keys),
		"redis_info":                info,
	}

	return stats, nil
}

// ClearAllRecommendations vide tout le cache de recommandations
func (c *RedisRecommendationCache) ClearAllRecommendations(ctx context.Context) error {
	pattern := "recommendations:*"
	keys, err := c.client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get all recommendation keys: %w", err)
	}

	if len(keys) > 0 {
		if err := c.client.Del(ctx, keys...).Err(); err != nil {
			return fmt.Errorf("failed to delete all recommendation keys: %w", err)
		}

		c.logger.Info("🗑️ Cleared all recommendations cache",
			zap.Int("keys_deleted", len(keys)),
		)
	}

	return nil
}
