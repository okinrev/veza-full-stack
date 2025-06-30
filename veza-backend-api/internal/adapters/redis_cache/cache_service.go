package redis_cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	"github.com/okinrev/veza-web-app/internal/domain/services"
)

// cacheService implémentation Redis du CacheService
type cacheService struct {
	client     *redis.Client
	logger     *zap.Logger
	defaultTTL time.Duration
}

// NewCacheService crée une nouvelle instance du service de cache Redis
func NewCacheService(client *redis.Client, defaultTTL time.Duration) services.CacheService {
	return &cacheService{
		client:     client,
		defaultTTL: defaultTTL,
	}
}

// Get récupère une valeur depuis le cache
func (c *cacheService) Get(ctx context.Context, key string) (interface{}, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Clé non trouvée
		}
		return nil, err
	}

	// Tenter de désérialiser en JSON
	var result interface{}
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		// Si échec JSON, retourner comme string
		return val, nil
	}

	return result, nil
}

// Set stocke une valeur dans le cache
func (c *cacheService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if ttl == 0 {
		ttl = c.defaultTTL
	}

	// Sérialiser en JSON
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, data, ttl).Err()
}

// Delete supprime une clé du cache
func (c *cacheService) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

// Exists vérifie si une clé existe
func (c *cacheService) Exists(ctx context.Context, key string) (bool, error) {
	count, err := c.client.Exists(ctx, key).Result()
	return count > 0, err
}

// ListPush ajoute des valeurs à une liste
func (c *cacheService) ListPush(ctx context.Context, key string, values ...interface{}) error {
	return c.client.LPush(ctx, key, values...).Err()
}

// ListPop récupère et supprime un élément de la liste
func (c *cacheService) ListPop(ctx context.Context, key string) (interface{}, error) {
	return c.client.LPop(ctx, key).Result()
}

// ListRange récupère une tranche de la liste
func (c *cacheService) ListRange(ctx context.Context, key string, start, stop int64) ([]interface{}, error) {
	result, err := c.client.LRange(ctx, key, start, stop).Result()
	if err != nil {
		return nil, err
	}

	// Convertir []string en []interface{}
	interfaces := make([]interface{}, len(result))
	for i, v := range result {
		interfaces[i] = v
	}

	return interfaces, nil
}

// ListLength retourne la longueur de la liste
func (c *cacheService) ListLength(ctx context.Context, key string) (int64, error) {
	return c.client.LLen(ctx, key).Result()
}

// SetAdd ajoute des membres à un set
func (c *cacheService) SetAdd(ctx context.Context, key string, members ...interface{}) error {
	return c.client.SAdd(ctx, key, members...).Err()
}

// SetRemove supprime des membres d'un set
func (c *cacheService) SetRemove(ctx context.Context, key string, members ...interface{}) error {
	return c.client.SRem(ctx, key, members...).Err()
}

// SetMembers retourne tous les membres d'un set
func (c *cacheService) SetMembers(ctx context.Context, key string) ([]interface{}, error) {
	result, err := c.client.SMembers(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	interfaces := make([]interface{}, len(result))
	for i, v := range result {
		interfaces[i] = v
	}

	return interfaces, nil
}

// SetIsMember vérifie si un membre est dans le set
func (c *cacheService) SetIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return c.client.SIsMember(ctx, key, member).Result()
}

// HashSet définit un champ dans un hash
func (c *cacheService) HashSet(ctx context.Context, key, field string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.client.HSet(ctx, key, field, data).Err()
}

// HashGet récupère un champ d'un hash
func (c *cacheService) HashGet(ctx context.Context, key, field string) (interface{}, error) {
	val, err := c.client.HGet(ctx, key, field).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var result interface{}
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		return val, nil
	}

	return result, nil
}

// HashGetAll récupère tous les champs d'un hash
func (c *cacheService) HashGetAll(ctx context.Context, key string) (map[string]interface{}, error) {
	result, err := c.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	hashMap := make(map[string]interface{})
	for field, value := range result {
		var parsedValue interface{}
		if err := json.Unmarshal([]byte(value), &parsedValue); err != nil {
			parsedValue = value
		}
		hashMap[field] = parsedValue
	}

	return hashMap, nil
}

// HashDelete supprime des champs d'un hash
func (c *cacheService) HashDelete(ctx context.Context, key string, fields ...string) error {
	return c.client.HDel(ctx, key, fields...).Err()
}

// Increment incrémente une valeur
func (c *cacheService) Increment(ctx context.Context, key string) (int64, error) {
	return c.client.Incr(ctx, key).Result()
}

// IncrementBy incrémente une valeur par un montant
func (c *cacheService) IncrementBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.client.IncrBy(ctx, key, value).Result()
}

// Expire définit un TTL sur une clé
func (c *cacheService) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return c.client.Expire(ctx, key, ttl).Err()
}

// TTL retourne le TTL d'une clé
func (c *cacheService) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.client.TTL(ctx, key).Result()
}

// Pipeline crée un pipeline Redis
func (c *cacheService) Pipeline() services.Pipeline {
	return &pipeline{
		pipe: c.client.Pipeline(),
	}
}

// Keys retourne les clés correspondant au pattern
func (c *cacheService) Keys(ctx context.Context, pattern string) ([]string, error) {
	return c.client.Keys(ctx, pattern).Result()
}

// DeletePattern supprime toutes les clés correspondant au pattern
func (c *cacheService) DeletePattern(ctx context.Context, pattern string) error {
	keys, err := c.client.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return c.client.Del(ctx, keys...).Err()
	}

	return nil
}

// Ping teste la connexion Redis
func (c *cacheService) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// Info retourne les informations Redis
func (c *cacheService) Info(ctx context.Context) (map[string]string, error) {
	info, err := c.client.Info(ctx).Result()
	if err != nil {
		return nil, err
	}

	// Parser les informations Redis (format key:value)
	infoMap := make(map[string]string)
	infoMap["raw"] = info
	infoMap["status"] = "connected"

	return infoMap, nil
}

// FlushDB vide la base de données Redis
func (c *cacheService) FlushDB(ctx context.Context) error {
	return c.client.FlushDB(ctx).Err()
}

// pipeline implémentation du Pipeline
type pipeline struct {
	pipe     redis.Pipeliner
	commands []*pipelineCommand
}

type pipelineCommand struct {
	result *services.PipelineResult
	cmd    redis.Cmder
}

// Get ajoute une commande Get au pipeline
func (p *pipeline) Get(key string) *services.PipelineResult {
	result := &services.PipelineResult{}
	cmd := p.pipe.Get(context.Background(), key)

	p.commands = append(p.commands, &pipelineCommand{
		result: result,
		cmd:    cmd,
	})

	return result
}

// Set ajoute une commande Set au pipeline
func (p *pipeline) Set(key string, value interface{}, ttl time.Duration) *services.PipelineResult {
	result := &services.PipelineResult{}

	data, err := json.Marshal(value)
	if err != nil {
		result.Error = err
		return result
	}

	cmd := p.pipe.Set(context.Background(), key, data, ttl)

	p.commands = append(p.commands, &pipelineCommand{
		result: result,
		cmd:    cmd,
	})

	return result
}

// Delete ajoute une commande Delete au pipeline
func (p *pipeline) Delete(key string) *services.PipelineResult {
	result := &services.PipelineResult{}
	cmd := p.pipe.Del(context.Background(), key)

	p.commands = append(p.commands, &pipelineCommand{
		result: result,
		cmd:    cmd,
	})

	return result
}

// Increment ajoute une commande Increment au pipeline
func (p *pipeline) Increment(key string) *services.PipelineResult {
	result := &services.PipelineResult{}
	cmd := p.pipe.Incr(context.Background(), key)

	p.commands = append(p.commands, &pipelineCommand{
		result: result,
		cmd:    cmd,
	})

	return result
}

// Execute exécute toutes les commandes du pipeline
func (p *pipeline) Execute(ctx context.Context) ([]*services.PipelineResult, error) {
	_, err := p.pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, err
	}

	// Remplir les résultats
	results := make([]*services.PipelineResult, len(p.commands))
	for i, cmd := range p.commands {
		results[i] = cmd.result

		if cmd.cmd.Err() != nil && cmd.cmd.Err() != redis.Nil {
			results[i].Error = cmd.cmd.Err()
		} else {
			// Récupérer la valeur selon le type de commande
			switch c := cmd.cmd.(type) {
			case *redis.StringCmd:
				results[i].Value = c.Val()
			case *redis.IntCmd:
				results[i].Value = c.Val()
			case *redis.StatusCmd:
				results[i].Value = c.Val()
			}
		}
	}

	return results, nil
}
