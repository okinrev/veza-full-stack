package redis_cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	"github.com/okinrev/veza-web-app/internal/domain/entities"
)

// CacheLevel représente les différents niveaux de cache
type CacheLevel int

const (
	CacheLevel1 CacheLevel = iota // Cache L1: In-memory local (très rapide)
	CacheLevel2                   // Cache L2: Redis distributed (rapide)
	CacheLevel3                   // Cache L3: Persistent cache (moyen)
)

// CacheStrategy définit les stratégies de cache
type CacheStrategy struct {
	TTLLevel1   time.Duration // TTL pour cache mémoire local
	TTLLevel2   time.Duration // TTL pour Redis
	TTLLevel3   time.Duration // TTL pour cache persistant
	MaxMemItems int           // Nombre max d'items en mémoire
	Compression bool          // Compression des données
}

var (
	// Stratégies pré-définies pour différents types de données
	SessionCacheStrategy = CacheStrategy{
		TTLLevel1:   5 * time.Minute,  // Cache local court pour performance
		TTLLevel2:   30 * time.Minute, // Cache Redis moyen
		TTLLevel3:   2 * time.Hour,    // Cache persistant long
		MaxMemItems: 10000,
		Compression: true,
	}

	RBACCacheStrategy = CacheStrategy{
		TTLLevel1:   15 * time.Minute, // Permissions changent peu
		TTLLevel2:   1 * time.Hour,    // Cache Redis long
		TTLLevel3:   6 * time.Hour,    // Cache persistant très long
		MaxMemItems: 5000,
		Compression: true,
	}

	QueryCacheStrategy = CacheStrategy{
		TTLLevel1:   2 * time.Minute,  // Résultats de requêtes courts
		TTLLevel2:   10 * time.Minute, // Cache Redis moyen
		TTLLevel3:   1 * time.Hour,    // Cache persistant long
		MaxMemItems: 20000,
		Compression: false, // Pas de compression pour performance
	}
)

// MultiLevelCacheService cache multi-niveaux enterprise
type MultiLevelCacheService struct {
	// Redis client pour niveau 2
	redis *redis.Client

	// Cache en mémoire niveau 1 (LRU)
	memCache   sync.Map
	memCounter sync.Map
	memMutex   sync.RWMutex

	// Métriques et monitoring
	metrics *CacheMetrics
	logger  *zap.Logger

	// Configuration
	config *CacheConfig
}

// CacheConfig configuration du cache multi-niveaux
type CacheConfig struct {
	EnableLevel1     bool
	EnableLevel2     bool
	EnableLevel3     bool
	MaxMemoryMB      int
	CompressionLevel int
	StatsInterval    time.Duration
}

// CacheItem item en cache avec métadonnées
type CacheItem struct {
	Data       interface{} `json:"data"`
	ExpireAt   time.Time   `json:"expire_at"`
	CreatedAt  time.Time   `json:"created_at"`
	AccessedAt time.Time   `json:"accessed_at"`
	HitCount   int64       `json:"hit_count"`
	Size       int64       `json:"size"`
	Level      CacheLevel  `json:"level"`
}

// CacheMetrics métriques de performance du cache
type CacheMetrics struct {
	L1Hits       int64 `json:"l1_hits"`
	L2Hits       int64 `json:"l2_hits"`
	L3Hits       int64 `json:"l3_hits"`
	Misses       int64 `json:"misses"`
	Evictions    int64 `json:"evictions"`
	TotalReads   int64 `json:"total_reads"`
	TotalWrites  int64 `json:"total_writes"`
	AvgLatencyMs int64 `json:"avg_latency_ms"`

	// Performance ratios
	HitRatio   float64 `json:"hit_ratio"`
	L1HitRatio float64 `json:"l1_hit_ratio"`
	L2HitRatio float64 `json:"l2_hit_ratio"`

	mutex sync.RWMutex
}

// NewMultiLevelCacheService crée un service de cache multi-niveaux
func NewMultiLevelCacheService(
	redisClient *redis.Client,
	config *CacheConfig,
	logger *zap.Logger,
) *MultiLevelCacheService {
	service := &MultiLevelCacheService{
		redis:   redisClient,
		metrics: &CacheMetrics{},
		logger:  logger,
		config:  config,
	}

	// Démarrer le garbage collector pour le cache mémoire
	go service.startMemoryCleaner()

	// Démarrer le monitoring des métriques
	go service.startMetricsReporter()

	return service
}

// ============================================================================
// SESSION CACHE OPTIMISÉ (3.1)
// ============================================================================

// GetUserSession récupère une session utilisateur avec cache multi-niveaux
func (c *MultiLevelCacheService) GetUserSession(ctx context.Context, userID int64) (*entities.User, bool) {
	start := time.Now()
	defer func() {
		c.recordLatency(time.Since(start))
	}()

	sessionKey := fmt.Sprintf("user_session:%d", userID)

	// Niveau 1 : Cache mémoire local (ultra-rapide)
	if c.config.EnableLevel1 {
		if item, found := c.getFromMemory(sessionKey); found {
			c.recordHit(CacheLevel1)
			if user, ok := item.Data.(*entities.User); ok {
				c.logger.Debug("Session trouvée en cache L1",
					zap.Int64("user_id", userID),
					zap.Duration("latency", time.Since(start)))
				return user, true
			}
		}
	}

	// Niveau 2 : Redis distribué (rapide)
	if c.config.EnableLevel2 {
		if data, err := c.redis.Get(ctx, sessionKey).Result(); err == nil {
			var user entities.User
			if err := json.Unmarshal([]byte(data), &user); err == nil {
				c.recordHit(CacheLevel2)

				// Remonter en cache L1 pour prochaines requêtes
				if c.config.EnableLevel1 {
					c.setInMemory(sessionKey, &user, SessionCacheStrategy.TTLLevel1)
				}

				c.logger.Debug("Session trouvée en cache L2",
					zap.Int64("user_id", userID),
					zap.Duration("latency", time.Since(start)))
				return &user, true
			}
		}
	}

	c.recordMiss()
	return nil, false
}

// SetUserSession stocke une session utilisateur avec stratégie multi-niveaux
func (c *MultiLevelCacheService) SetUserSession(ctx context.Context, user *entities.User) error {
	sessionKey := fmt.Sprintf("user_session:%d", user.ID)

	// Sérialiser pour Redis
	userData, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("erreur sérialisation user: %w", err)
	}

	var wg sync.WaitGroup
	errors := make(chan error, 2)

	// Niveau 1 : Cache mémoire (parallèle)
	if c.config.EnableLevel1 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.setInMemory(sessionKey, user, SessionCacheStrategy.TTLLevel1)
		}()
	}

	// Niveau 2 : Redis (parallèle)
	if c.config.EnableLevel2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := c.redis.Set(ctx, sessionKey, userData, SessionCacheStrategy.TTLLevel2).Err(); err != nil {
				errors <- fmt.Errorf("erreur cache L2: %w", err)
			}
		}()
	}

	wg.Wait()
	close(errors)

	// Vérifier les erreurs
	for err := range errors {
		if err != nil {
			c.logger.Warn("Erreur mise en cache session", zap.Error(err))
		}
	}

	c.recordWrite()
	return nil
}

// InvalidateUserSession invalide une session sur tous les niveaux
func (c *MultiLevelCacheService) InvalidateUserSession(ctx context.Context, userID int64) error {
	sessionKey := fmt.Sprintf("user_session:%d", userID)

	// Invalidation parallèle sur tous les niveaux
	var wg sync.WaitGroup

	// Niveau 1 : Mémoire
	if c.config.EnableLevel1 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.memCache.Delete(sessionKey)
		}()
	}

	// Niveau 2 : Redis
	if c.config.EnableLevel2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.redis.Del(ctx, sessionKey)
		}()
	}

	wg.Wait()

	c.logger.Debug("Session invalidée", zap.Int64("user_id", userID))
	return nil
}

// ============================================================================
// CACHE MÉMOIRE NIVEAU 1 (LRU)
// ============================================================================

func (c *MultiLevelCacheService) getFromMemory(key string) (*CacheItem, bool) {
	if value, exists := c.memCache.Load(key); exists {
		if item, ok := value.(*CacheItem); ok {
			// Vérifier expiration
			if time.Now().Before(item.ExpireAt) {
				// Mettre à jour les statistiques d'accès
				item.AccessedAt = time.Now()
				item.HitCount++
				return item, true
			} else {
				// Item expiré, le supprimer
				c.memCache.Delete(key)
			}
		}
	}
	return nil, false
}

func (c *MultiLevelCacheService) setInMemory(key string, data interface{}, ttl time.Duration) {
	item := &CacheItem{
		Data:       data,
		ExpireAt:   time.Now().Add(ttl),
		CreatedAt:  time.Now(),
		AccessedAt: time.Now(),
		HitCount:   0,
		Level:      CacheLevel1,
	}

	c.memCache.Store(key, item)
}

// ============================================================================
// MÉTRIQUES ET MONITORING
// ============================================================================

func (c *MultiLevelCacheService) recordHit(level CacheLevel) {
	c.metrics.mutex.Lock()
	defer c.metrics.mutex.Unlock()

	switch level {
	case CacheLevel1:
		c.metrics.L1Hits++
	case CacheLevel2:
		c.metrics.L2Hits++
	case CacheLevel3:
		c.metrics.L3Hits++
	}
	c.metrics.TotalReads++
}

func (c *MultiLevelCacheService) recordMiss() {
	c.metrics.mutex.Lock()
	defer c.metrics.mutex.Unlock()

	c.metrics.Misses++
	c.metrics.TotalReads++
}

func (c *MultiLevelCacheService) recordWrite() {
	c.metrics.mutex.Lock()
	defer c.metrics.mutex.Unlock()

	c.metrics.TotalWrites++
}

func (c *MultiLevelCacheService) recordLatency(latency time.Duration) {
	c.metrics.mutex.Lock()
	defer c.metrics.mutex.Unlock()

	// Simple moving average pour la latence
	c.metrics.AvgLatencyMs = (c.metrics.AvgLatencyMs + latency.Milliseconds()) / 2
}

// GetMetrics retourne les métriques actuelles
func (c *MultiLevelCacheService) GetMetrics() *CacheMetrics {
	c.metrics.mutex.RLock()
	defer c.metrics.mutex.RUnlock()

	// Calculer les ratios
	totalHits := c.metrics.L1Hits + c.metrics.L2Hits + c.metrics.L3Hits
	if c.metrics.TotalReads > 0 {
		c.metrics.HitRatio = float64(totalHits) / float64(c.metrics.TotalReads)
		c.metrics.L1HitRatio = float64(c.metrics.L1Hits) / float64(c.metrics.TotalReads)
		c.metrics.L2HitRatio = float64(c.metrics.L2Hits) / float64(c.metrics.TotalReads)
	}

	// Retourner une copie
	return &CacheMetrics{
		L1Hits:       c.metrics.L1Hits,
		L2Hits:       c.metrics.L2Hits,
		L3Hits:       c.metrics.L3Hits,
		Misses:       c.metrics.Misses,
		Evictions:    c.metrics.Evictions,
		TotalReads:   c.metrics.TotalReads,
		TotalWrites:  c.metrics.TotalWrites,
		AvgLatencyMs: c.metrics.AvgLatencyMs,
		HitRatio:     c.metrics.HitRatio,
		L1HitRatio:   c.metrics.L1HitRatio,
		L2HitRatio:   c.metrics.L2HitRatio,
	}
}

// ============================================================================
// GARBAGE COLLECTION ET MAINTENANCE
// ============================================================================

func (c *MultiLevelCacheService) startMemoryCleaner() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.cleanExpiredMemoryItems()
	}
}

func (c *MultiLevelCacheService) cleanExpiredMemoryItems() {
	now := time.Now()
	itemsToDelete := []string{}

	c.memCache.Range(func(key, value interface{}) bool {
		if item, ok := value.(*CacheItem); ok {
			if now.After(item.ExpireAt) {
				itemsToDelete = append(itemsToDelete, key.(string))
			}
		}
		return true
	})

	for _, key := range itemsToDelete {
		c.memCache.Delete(key)
		c.metrics.mutex.Lock()
		c.metrics.Evictions++
		c.metrics.mutex.Unlock()
	}

	if len(itemsToDelete) > 0 {
		c.logger.Debug("Cache mémoire nettoyé",
			zap.Int("items_supprimés", len(itemsToDelete)))
	}
}

func (c *MultiLevelCacheService) startMetricsReporter() {
	ticker := time.NewTicker(c.config.StatsInterval)
	defer ticker.Stop()

	for range ticker.C {
		metrics := c.GetMetrics()
		c.logger.Info("Métriques cache multi-niveaux",
			zap.Int64("l1_hits", metrics.L1Hits),
			zap.Int64("l2_hits", metrics.L2Hits),
			zap.Int64("misses", metrics.Misses),
			zap.Float64("hit_ratio", metrics.HitRatio),
			zap.Int64("avg_latency_ms", metrics.AvgLatencyMs))
	}
}

// ============================================================================
// MÉTHODES GÉNÉRIQUES POUR CACHE WARMER
// ============================================================================

// Set méthode générique pour stocker des données dans le cache multi-niveaux
func (c *MultiLevelCacheService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	// Stocker en cache mémoire (L1)
	if c.config.EnableLevel1 {
		c.setInMemory(key, value, ttl)
	}

	// Stocker en Redis (L2)
	if c.config.EnableLevel2 {
		data, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("erreur sérialisation pour cache: %w", err)
		}

		if err := c.redis.Set(ctx, key, data, ttl).Err(); err != nil {
			return fmt.Errorf("erreur stockage Redis: %w", err)
		}
	}

	c.recordWrite()
	return nil
}

// Get méthode générique pour récupérer des données du cache multi-niveaux
func (c *MultiLevelCacheService) Get(ctx context.Context, key string, dest interface{}) (bool, error) {
	start := time.Now()
	defer func() {
		c.recordLatency(time.Since(start))
	}()

	// Niveau 1 : Cache mémoire local
	if c.config.EnableLevel1 {
		if item, found := c.getFromMemory(key); found {
			c.recordHit(CacheLevel1)
			// Copy data to destination
			switch v := dest.(type) {
			case *interface{}:
				*v = item.Data
			default:
				// Try JSON marshaling for type conversion
				data, err := json.Marshal(item.Data)
				if err != nil {
					return false, fmt.Errorf("erreur conversion cache L1: %w", err)
				}
				if err := json.Unmarshal(data, dest); err != nil {
					return false, fmt.Errorf("erreur désérialisation cache L1: %w", err)
				}
			}
			return true, nil
		}
	}

	// Niveau 2 : Redis distribué
	if c.config.EnableLevel2 {
		data, err := c.redis.Get(ctx, key).Result()
		if err == nil {
			c.recordHit(CacheLevel2)

			// Désérialiser les données
			if err := json.Unmarshal([]byte(data), dest); err != nil {
				return false, fmt.Errorf("erreur désérialisation Redis: %w", err)
			}

			// Remonter en cache L1
			if c.config.EnableLevel1 {
				c.setInMemory(key, dest, SessionCacheStrategy.TTLLevel1)
			}

			return true, nil
		}
	}

	c.recordMiss()
	return false, nil
}

// Health check du service de cache
func (c *MultiLevelCacheService) HealthCheck(ctx context.Context) error {
	// Test Redis
	if c.config.EnableLevel2 {
		if err := c.redis.Ping(ctx).Err(); err != nil {
			return fmt.Errorf("redis indisponible: %w", err)
		}
	}

	// Test cache mémoire
	testKey := "health_check_test"
	c.setInMemory(testKey, "test", 1*time.Second)
	if _, found := c.getFromMemory(testKey); !found {
		return fmt.Errorf("cache mémoire défaillant")
	}
	c.memCache.Delete(testKey)

	return nil
}
