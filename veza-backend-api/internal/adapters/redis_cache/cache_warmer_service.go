package redis_cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	"github.com/okinrev/veza-web-app/internal/constants"
)

// CacheWarmerService service pour préchauffer les caches critiques
type CacheWarmerService struct {
	redis           *redis.Client
	multiLevelCache *MultiLevelCacheService
	rbacCache       *RBACCacheService
	queryCache      *QueryCacheService
	logger          *zap.Logger

	// Configuration
	config *CacheWarmerConfig

	// État du service
	isWarming  bool
	lastWarmup time.Time
	mutex      sync.RWMutex
}

// CacheWarmerConfig configuration du cache warmer
type CacheWarmerConfig struct {
	Enabled            bool          `json:"enabled"`
	WarmupInterval     time.Duration `json:"warmup_interval"`     // Intervalle entre warmups
	WarmupTimeout      time.Duration `json:"warmup_timeout"`      // Timeout pour warmup complet
	ConcurrentWarmups  int           `json:"concurrent_warmups"`  // Nombre de warmups simultanés
	PreloadUserData    bool          `json:"preload_user_data"`   // Précharger données utilisateurs
	PreloadPermissions bool          `json:"preload_permissions"` // Précharger permissions RBAC
	PreloadQueries     bool          `json:"preload_queries"`     // Précharger requêtes fréquentes
	PreloadOnStartup   bool          `json:"preload_on_startup"`  // Préchauffer au démarrage
	WarmupBatchSize    int           `json:"warmup_batch_size"`   // Taille des batches de warmup
}

// DefaultCacheWarmerConfig configuration par défaut
func DefaultCacheWarmerConfig() *CacheWarmerConfig {
	return &CacheWarmerConfig{
		Enabled:            true,
		WarmupInterval:     10 * time.Minute, // Warmup toutes les 10 minutes
		WarmupTimeout:      5 * time.Minute,  // Timeout de 5 minutes
		ConcurrentWarmups:  5,                // 5 warmups simultanés
		PreloadUserData:    true,
		PreloadPermissions: true,
		PreloadQueries:     true,
		PreloadOnStartup:   true,
		WarmupBatchSize:    100, // Batches de 100 items
	}
}

// CacheWarmupStats statistiques de warmup
type CacheWarmupStats struct {
	LastWarmupTime  time.Time     `json:"last_warmup_time"`
	WarmupDuration  time.Duration `json:"warmup_duration"`
	ItemsWarmedUp   int64         `json:"items_warmed_up"`
	CacheHitsBefore int64         `json:"cache_hits_before"`
	CacheHitsAfter  int64         `json:"cache_hits_after"`
	WarmupErrors    int64         `json:"warmup_errors"`
	WarmupSuccess   bool          `json:"warmup_success"`
}

// NewCacheWarmerService crée un nouveau service de cache warming
func NewCacheWarmerService(
	redisClient *redis.Client,
	multiLevelCache *MultiLevelCacheService,
	rbacCache *RBACCacheService,
	queryCache *QueryCacheService,
	logger *zap.Logger,
) *CacheWarmerService {
	config := DefaultCacheWarmerConfig()

	service := &CacheWarmerService{
		redis:           redisClient,
		multiLevelCache: multiLevelCache,
		rbacCache:       rbacCache,
		queryCache:      queryCache,
		logger:          logger,
		config:          config,
		isWarming:       false,
	}

	// Démarrer le préchauffage automatique si activé
	if config.Enabled {
		go service.startPeriodicWarmup()

		// Préchauffage initial au démarrage
		if config.PreloadOnStartup {
			go service.InitialWarmup()
		}
	}

	return service
}

// InitialWarmup préchauffage initial au démarrage de l'application
func (c *CacheWarmerService) InitialWarmup() error {
	c.logger.Info("Démarrage du préchauffage initial des caches")

	ctx, cancel := context.WithTimeout(context.Background(), c.config.WarmupTimeout)
	defer cancel()

	return c.WarmupCaches(ctx)
}

// WarmupCaches préchauffe tous les caches configurés
func (c *CacheWarmerService) WarmupCaches(ctx context.Context) error {
	c.mutex.Lock()
	if c.isWarming {
		c.mutex.Unlock()
		return fmt.Errorf("warmup déjà en cours")
	}
	c.isWarming = true
	c.mutex.Unlock()

	defer func() {
		c.mutex.Lock()
		c.isWarming = false
		c.lastWarmup = time.Now()
		c.mutex.Unlock()
	}()

	start := time.Now()
	var wg sync.WaitGroup
	errors := make(chan error, 10)
	itemsWarmed := int64(0)

	// Préchauffage parallèle des différents types de cache
	tasks := []func(context.Context) (int64, error){}

	if c.config.PreloadUserData {
		tasks = append(tasks, c.warmupUserData)
	}
	if c.config.PreloadPermissions {
		tasks = append(tasks, c.warmupPermissions)
	}
	if c.config.PreloadQueries {
		tasks = append(tasks, c.warmupFrequentQueries)
	}

	// Exécuter les tâches de warmup
	for _, task := range tasks {
		wg.Add(1)
		go func(taskFunc func(context.Context) (int64, error)) {
			defer wg.Done()

			if items, err := taskFunc(ctx); err != nil {
				errors <- err
			} else {
				itemsWarmed += items
			}
		}(task)
	}

	// Attendre la fin de tous les warmups
	go func() {
		wg.Wait()
		close(errors)
	}()

	// Collecter les erreurs
	var warmupErrors []error
	for err := range errors {
		warmupErrors = append(warmupErrors, err)
	}

	duration := time.Since(start)

	c.logger.Info("Préchauffage des caches terminé",
		zap.Duration("duration", duration),
		zap.Int64("items_warmed", itemsWarmed),
		zap.Int("errors", len(warmupErrors)),
	)

	if len(warmupErrors) > 0 {
		return fmt.Errorf("erreurs pendant le warmup: %v", warmupErrors)
	}

	return nil
}

// warmupUserData préchauffe les données utilisateurs courantes
func (c *CacheWarmerService) warmupUserData(ctx context.Context) (int64, error) {
	c.logger.Debug("Préchauffage données utilisateurs")

	// Simuler la récupération des utilisateurs actifs récents
	activeUserIDs := []int64{1, 2, 3, 4, 5} // À remplacer par vraie logique

	var wg sync.WaitGroup
	itemsWarmed := int64(0)
	errors := make(chan error, len(activeUserIDs))

	for _, userID := range activeUserIDs {
		wg.Add(1)
		go func(id int64) {
			defer wg.Done()

			// Préchauffer les sessions utilisateur
			sessionKey := fmt.Sprintf("session:user:%d", id)
			if err := c.warmupCacheKey(ctx, sessionKey, map[string]interface{}{
				"user_id": id,
				"active":  true,
				"warmed":  time.Now(),
			}); err != nil {
				errors <- err
				return
			}

			itemsWarmed++
		}(userID)
	}

	wg.Wait()
	close(errors)

	// Vérifier s'il y a eu des erreurs
	for err := range errors {
		return itemsWarmed, err
	}

	return itemsWarmed, nil
}

// warmupPermissions préchauffe les permissions RBAC fréquentes
func (c *CacheWarmerService) warmupPermissions(ctx context.Context) (int64, error) {
	c.logger.Debug("Préchauffage permissions RBAC")

	// Préchauffer les permissions pour les rôles principaux
	roles := []constants.Role{
		constants.RoleUser,
		constants.RoleAdmin,
		constants.RoleModerator,
		constants.RoleSuperAdmin,
	}

	itemsWarmed := int64(0)

	for _, role := range roles {
		// Préchauffer les permissions du rôle
		permKey := fmt.Sprintf("rbac:role:%s:permissions", string(role))
		permissions := constants.RolePermissions[role]

		if err := c.warmupCacheKey(ctx, permKey, permissions); err != nil {
			return itemsWarmed, err
		}

		itemsWarmed++
	}

	return itemsWarmed, nil
}

// warmupFrequentQueries préchauffe les requêtes fréquentes
func (c *CacheWarmerService) warmupFrequentQueries(ctx context.Context) (int64, error) {
	c.logger.Debug("Préchauffage requêtes fréquentes")

	// Requêtes fréquentes à préchauffer
	frequentQueries := map[string]interface{}{
		"query:users:active_count":    1000,
		"query:rooms:public_list":     []string{"general", "tech", "random"},
		"query:analytics:daily_stats": map[string]int{"users": 500, "messages": 1500},
	}

	itemsWarmed := int64(0)

	for key, value := range frequentQueries {
		if err := c.warmupCacheKey(ctx, key, value); err != nil {
			return itemsWarmed, err
		}
		itemsWarmed++
	}

	return itemsWarmed, nil
}

// warmupCacheKey préchauffe une clé de cache spécifique
func (c *CacheWarmerService) warmupCacheKey(ctx context.Context, key string, value interface{}) error {
	// Utiliser le cache multi-niveaux pour le warmup
	if c.multiLevelCache != nil {
		return c.multiLevelCache.Set(ctx, key, value, 1*time.Hour)
	}

	// Fallback sur Redis direct
	return c.redis.Set(ctx, key, value, 1*time.Hour).Err()
}

// startPeriodicWarmup démarre le préchauffage périodique
func (c *CacheWarmerService) startPeriodicWarmup() {
	ticker := time.NewTicker(c.config.WarmupInterval)
	defer ticker.Stop()

	for range ticker.C {
		if !c.config.Enabled {
			continue
		}

		c.logger.Debug("Début préchauffage périodique")

		ctx, cancel := context.WithTimeout(context.Background(), c.config.WarmupTimeout)
		if err := c.WarmupCaches(ctx); err != nil {
			c.logger.Error("Erreur préchauffage périodique", zap.Error(err))
		}
		cancel()
	}
}

// GetWarmupStats retourne les statistiques de warmup
func (c *CacheWarmerService) GetWarmupStats() *CacheWarmupStats {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return &CacheWarmupStats{
		LastWarmupTime: c.lastWarmup,
		WarmupSuccess:  !c.lastWarmup.IsZero(),
	}
}

// IsWarmingUp indique si un warmup est en cours
func (c *CacheWarmerService) IsWarmingUp() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.isWarming
}

// ForceWarmup force un préchauffage immédiat
func (c *CacheWarmerService) ForceWarmup() error {
	c.logger.Info("Préchauffage forcé des caches")

	ctx, cancel := context.WithTimeout(context.Background(), c.config.WarmupTimeout)
	defer cancel()

	return c.WarmupCaches(ctx)
}

// UpdateConfig met à jour la configuration du cache warmer
func (c *CacheWarmerService) UpdateConfig(config *CacheWarmerConfig) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.config = config
	c.logger.Info("Configuration cache warmer mise à jour")
}

// HealthCheck vérifie la santé du service de cache warming
func (c *CacheWarmerService) HealthCheck(ctx context.Context) error {
	if !c.config.Enabled {
		return nil
	}

	c.mutex.RLock()
	lastWarmup := c.lastWarmup
	c.mutex.RUnlock()

	// Vérifier que le dernier warmup n'est pas trop ancien
	maxAge := c.config.WarmupInterval * 2
	if time.Since(lastWarmup) > maxAge {
		return fmt.Errorf("dernier warmup trop ancien: %v", time.Since(lastWarmup))
	}

	return nil
}
