package redis_cache

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// QueryCacheService cache intelligent pour les résultats de requêtes
type QueryCacheService struct {
	redis      *redis.Client
	localCache sync.Map
	hotQueries sync.Map // Cache des requêtes les plus fréquentes
	logger     *zap.Logger
	metrics    *QueryCacheMetrics

	// Configuration
	defaultTTL       time.Duration
	hotQueryTTL      time.Duration
	maxLocalItems    int
	compressionLevel int
}

// QueryCacheMetrics métriques du cache de requêtes
type QueryCacheMetrics struct {
	QueryExecutions int64            `json:"query_executions"`
	CacheHits       int64            `json:"cache_hits"`
	CacheMisses     int64            `json:"cache_misses"`
	LocalCacheHits  int64            `json:"local_cache_hits"`
	RedisCacheHits  int64            `json:"redis_cache_hits"`
	TotalSavedMs    int64            `json:"total_saved_ms"`
	AvgQueryTimeMs  int64            `json:"avg_query_time_ms"`
	HotQueries      map[string]int64 `json:"hot_queries"`

	mutex sync.RWMutex
}

// CachedQuery structure d'une requête mise en cache
type CachedQuery struct {
	QueryHash       string        `json:"query_hash"`
	QuerySQL        string        `json:"query_sql,omitempty"`
	Parameters      []interface{} `json:"parameters,omitempty"`
	Result          interface{}   `json:"result"`
	CachedAt        time.Time     `json:"cached_at"`
	ExpiresAt       time.Time     `json:"expires_at"`
	AccessCount     int64         `json:"access_count"`
	LastAccessedAt  time.Time     `json:"last_accessed_at"`
	ExecutionTimeMs int64         `json:"execution_time_ms"`
	ResultSize      int64         `json:"result_size"`
	IsHot           bool          `json:"is_hot"`
}

// QueryPattern patterns de requêtes courantes pour optimisation
type QueryPattern struct {
	Pattern     string        `json:"pattern"`
	TTL         time.Duration `json:"ttl"`
	UseLocal    bool          `json:"use_local"`
	Compression bool          `json:"compression"`
}

var (
	// Patterns pré-définis pour différents types de requêtes
	QueryPatterns = map[string]QueryPattern{
		"user_profile":     {Pattern: "SELECT * FROM users WHERE", TTL: 15 * time.Minute, UseLocal: true, Compression: false},
		"user_sessions":    {Pattern: "SELECT * FROM user_sessions WHERE", TTL: 5 * time.Minute, UseLocal: true, Compression: false},
		"chat_messages":    {Pattern: "SELECT * FROM messages WHERE", TTL: 2 * time.Minute, UseLocal: false, Compression: true},
		"user_permissions": {Pattern: "SELECT permissions FROM", TTL: 30 * time.Minute, UseLocal: true, Compression: false},
		"room_members":     {Pattern: "SELECT users FROM room_members", TTL: 10 * time.Minute, UseLocal: true, Compression: false},
		"file_metadata":    {Pattern: "SELECT * FROM files WHERE", TTL: 1 * time.Hour, UseLocal: false, Compression: true},
		"analytics":        {Pattern: "SELECT COUNT(*) FROM", TTL: 5 * time.Minute, UseLocal: false, Compression: false},
	}
)

// NewQueryCacheService crée un nouveau service de cache de requêtes
func NewQueryCacheService(
	redisClient *redis.Client,
	logger *zap.Logger,
) *QueryCacheService {
	service := &QueryCacheService{
		redis:            redisClient,
		logger:           logger,
		metrics:          &QueryCacheMetrics{HotQueries: make(map[string]int64)},
		defaultTTL:       10 * time.Minute,
		hotQueryTTL:      30 * time.Minute,
		maxLocalItems:    1000,
		compressionLevel: 1,
	}

	// Démarrer le nettoyage périodique
	go service.startCacheCleaner()

	// Démarrer l'analyse des patterns
	go service.startHotQueryAnalyzer()

	// Démarrer le reporting des métriques
	go service.startMetricsReporter()

	return service
}

// ============================================================================
// CACHE INTELLIGENT DE REQUÊTES
// ============================================================================

// ExecuteWithCache exécute une requête avec cache intelligent
func (q *QueryCacheService) ExecuteWithCache(
	ctx context.Context,
	querySQL string,
	parameters []interface{},
	executor func(context.Context, string, []interface{}) (interface{}, error),
) (interface{}, error) {
	start := time.Now()

	q.metrics.mutex.Lock()
	q.metrics.QueryExecutions++
	q.metrics.mutex.Unlock()

	// Générer la clé de cache
	cacheKey := q.generateCacheKey(querySQL, parameters)

	// Identifier le pattern de requête
	pattern := q.identifyQueryPattern(querySQL)

	// Essayer de récupérer depuis le cache
	if cachedResult, found := q.getCachedResult(ctx, cacheKey, pattern); found {
		q.recordCacheHit(time.Since(start), cachedResult.ExecutionTimeMs)
		q.updateAccessStats(cacheKey, cachedResult)

		q.logger.Debug("Résultat trouvé en cache",
			zap.String("query_hash", cacheKey),
			zap.String("pattern", pattern),
			zap.Bool("from_local", cachedResult.IsHot),
			zap.Duration("latency", time.Since(start)))

		return cachedResult.Result, nil
	}

	// Exécuter la requête
	result, err := executor(ctx, querySQL, parameters)
	if err != nil {
		q.recordCacheMiss()
		return nil, err
	}

	executionTime := time.Since(start)

	// Mettre en cache le résultat
	go q.cacheResult(ctx, cacheKey, querySQL, parameters, result, executionTime, pattern)

	q.recordCacheMiss()
	q.recordQueryExecution(executionTime)

	return result, nil
}

// generateCacheKey génère une clé de cache unique pour la requête
func (q *QueryCacheService) generateCacheKey(querySQL string, parameters []interface{}) string {
	// Normaliser la requête SQL
	normalizedSQL := q.normalizeSQL(querySQL)

	// Sérialiser les paramètres
	paramStr := ""
	if len(parameters) > 0 {
		if paramBytes, err := json.Marshal(parameters); err == nil {
			paramStr = string(paramBytes)
		}
	}

	// Générer le hash MD5
	hasher := md5.New()
	hasher.Write([]byte(normalizedSQL + paramStr))
	return "query:" + hex.EncodeToString(hasher.Sum(nil))
}

// normalizeSQL normalise le SQL pour une meilleure cohérence du cache
func (q *QueryCacheService) normalizeSQL(sql string) string {
	// Supprimer les espaces multiples et normaliser
	sql = strings.ReplaceAll(sql, "\n", " ")
	sql = strings.ReplaceAll(sql, "\t", " ")

	words := strings.Fields(sql)
	for i, word := range words {
		words[i] = strings.ToUpper(word)
	}

	return strings.Join(words, " ")
}

// identifyQueryPattern identifie le pattern d'une requête
func (q *QueryCacheService) identifyQueryPattern(sql string) string {
	normalizedSQL := strings.ToUpper(sql)

	for patternName, pattern := range QueryPatterns {
		if strings.Contains(normalizedSQL, strings.ToUpper(pattern.Pattern)) {
			return patternName
		}
	}

	return "default"
}

// getCachedResult récupère un résultat depuis le cache
func (q *QueryCacheService) getCachedResult(ctx context.Context, cacheKey, pattern string) (*CachedQuery, bool) {
	// Vérifier d'abord le cache local pour les requêtes hot
	if cached, exists := q.localCache.Load(cacheKey); exists {
		if cachedQuery, ok := cached.(*CachedQuery); ok {
			if time.Now().Before(cachedQuery.ExpiresAt) {
				q.metrics.mutex.Lock()
				q.metrics.LocalCacheHits++
				q.metrics.mutex.Unlock()
				return cachedQuery, true
			} else {
				q.localCache.Delete(cacheKey)
			}
		}
	}

	// Vérifier Redis
	if data, err := q.redis.Get(ctx, cacheKey).Result(); err == nil {
		var cachedQuery CachedQuery
		if err := json.Unmarshal([]byte(data), &cachedQuery); err == nil {
			if time.Now().Before(cachedQuery.ExpiresAt) {
				q.metrics.mutex.Lock()
				q.metrics.RedisCacheHits++
				q.metrics.mutex.Unlock()

				// Promouvoir en cache local si c'est une requête fréquente
				if q.shouldPromoteToLocal(pattern, &cachedQuery) {
					q.localCache.Store(cacheKey, &cachedQuery)
				}

				return &cachedQuery, true
			}
		}
	}

	return nil, false
}

// cacheResult met en cache un résultat de requête
func (q *QueryCacheService) cacheResult(
	ctx context.Context,
	cacheKey, querySQL string,
	parameters []interface{},
	result interface{},
	executionTime time.Duration,
	pattern string,
) {
	// Calculer la taille du résultat
	resultSize := int64(0)
	if resultBytes, err := json.Marshal(result); err == nil {
		resultSize = int64(len(resultBytes))
	}

	// Déterminer le TTL selon le pattern
	ttl := q.defaultTTL
	useLocal := false
	compression := false

	if queryPattern, exists := QueryPatterns[pattern]; exists {
		ttl = queryPattern.TTL
		useLocal = queryPattern.UseLocal
		compression = queryPattern.Compression
	}

	cachedQuery := &CachedQuery{
		QueryHash:       cacheKey,
		QuerySQL:        querySQL,
		Parameters:      parameters,
		Result:          result,
		CachedAt:        time.Now(),
		ExpiresAt:       time.Now().Add(ttl),
		AccessCount:     0,
		LastAccessedAt:  time.Now(),
		ExecutionTimeMs: executionTime.Milliseconds(),
		ResultSize:      resultSize,
		IsHot:           useLocal,
	}

	// Sérialiser
	data, err := json.Marshal(cachedQuery)
	if err != nil {
		q.logger.Warn("Erreur sérialisation cache query", zap.Error(err))
		return
	}

	// Compression si nécessaire
	if compression && len(data) > 1024 {
		// TODO: Implémenter la compression
	}

	var wg sync.WaitGroup

	// Cache Redis (toujours)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := q.redis.Set(ctx, cacheKey, data, ttl).Err(); err != nil {
			q.logger.Warn("Erreur mise en cache Redis", zap.Error(err))
		}
	}()

	// Cache local si requis
	if useLocal {
		wg.Add(1)
		go func() {
			defer wg.Done()
			q.setInLocalCache(cacheKey, cachedQuery)
		}()
	}

	wg.Wait()

	q.logger.Debug("Résultat mis en cache",
		zap.String("query_hash", cacheKey),
		zap.String("pattern", pattern),
		zap.Duration("execution_time", executionTime),
		zap.Int64("result_size", resultSize),
		zap.Bool("local_cache", useLocal))
}

// setInLocalCache ajoute au cache local avec gestion LRU
func (q *QueryCacheService) setInLocalCache(key string, cachedQuery *CachedQuery) {
	// Vérifier la limite du cache local
	currentSize := 0
	q.localCache.Range(func(_, _ interface{}) bool {
		currentSize++
		return true
	})

	// Si limite atteinte, supprimer les plus anciens
	if currentSize >= q.maxLocalItems {
		q.evictOldestLocalItems(currentSize - q.maxLocalItems + 1)
	}

	q.localCache.Store(key, cachedQuery)
}

// evictOldestLocalItems supprime les items les plus anciens du cache local
func (q *QueryCacheService) evictOldestLocalItems(count int) {
	type itemWithKey struct {
		key        string
		lastAccess time.Time
	}

	var items []itemWithKey

	q.localCache.Range(func(key, value interface{}) bool {
		if cached, ok := value.(*CachedQuery); ok {
			items = append(items, itemWithKey{
				key:        key.(string),
				lastAccess: cached.LastAccessedAt,
			})
		}
		return true
	})

	// Trier par date d'accès
	sort.Slice(items, func(i, j int) bool {
		return items[i].lastAccess.Before(items[j].lastAccess)
	})

	// Supprimer les plus anciens
	for i := 0; i < count && i < len(items); i++ {
		q.localCache.Delete(items[i].key)
	}
}

// shouldPromoteToLocal détermine si une requête doit être promue en cache local
func (q *QueryCacheService) shouldPromoteToLocal(pattern string, cachedQuery *CachedQuery) bool {
	// Promouvoir si c'est un pattern qui utilise normalement le cache local
	if queryPattern, exists := QueryPatterns[pattern]; exists && queryPattern.UseLocal {
		return true
	}

	// Promouvoir si la requête est accédée fréquemment
	return cachedQuery.AccessCount > 10
}

// updateAccessStats met à jour les statistiques d'accès
func (q *QueryCacheService) updateAccessStats(cacheKey string, cachedQuery *CachedQuery) {
	cachedQuery.AccessCount++
	cachedQuery.LastAccessedAt = time.Now()

	// Mettre à jour les statistiques des requêtes hot
	q.metrics.mutex.Lock()
	q.metrics.HotQueries[cacheKey]++
	q.metrics.mutex.Unlock()
}

// ============================================================================
// INVALIDATION INTELLIGENTE
// ============================================================================

// InvalidateByTable invalide les caches liés à une table
func (q *QueryCacheService) InvalidateByTable(ctx context.Context, tableName string) error {
	pattern := fmt.Sprintf("query:*%s*", strings.ToUpper(tableName))

	// Récupérer toutes les clés correspondantes
	keys, err := q.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("erreur récupération clés: %w", err)
	}

	if len(keys) == 0 {
		return nil
	}

	// Supprimer de Redis
	if err := q.redis.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("erreur suppression Redis: %w", err)
	}

	// Supprimer du cache local
	q.localCache.Range(func(key, value interface{}) bool {
		if keyStr, ok := key.(string); ok {
			if strings.Contains(strings.ToUpper(keyStr), strings.ToUpper(tableName)) {
				q.localCache.Delete(key)
			}
		}
		return true
	})

	q.logger.Info("Cache invalidé par table",
		zap.String("table", tableName),
		zap.Int("keys_deleted", len(keys)))

	return nil
}

// InvalidateByPattern invalide les caches selon un pattern
func (q *QueryCacheService) InvalidateByPattern(ctx context.Context, pattern string) error {
	// Supprimer de Redis
	keys, err := q.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		q.redis.Del(ctx, keys...)
	}

	// Supprimer du cache local
	q.localCache.Range(func(key, value interface{}) bool {
		if keyStr, ok := key.(string); ok {
			// Utiliser une logique de matching simple
			if strings.Contains(keyStr, strings.ReplaceAll(pattern, "*", "")) {
				q.localCache.Delete(key)
			}
		}
		return true
	})

	return nil
}

// ============================================================================
// ANALYSE DES PATTERNS ET OPTIMISATION
// ============================================================================

// startHotQueryAnalyzer analyse les requêtes chaudes pour optimisation
func (q *QueryCacheService) startHotQueryAnalyzer() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		q.analyzeHotQueries()
	}
}

// analyzeHotQueries analyse et optimise les requêtes les plus fréquentes
func (q *QueryCacheService) analyzeHotQueries() {
	q.metrics.mutex.RLock()
	hotQueries := make(map[string]int64)
	for k, v := range q.metrics.HotQueries {
		hotQueries[k] = v
	}
	q.metrics.mutex.RUnlock()

	// Trier les requêtes par fréquence
	type queryFreq struct {
		hash  string
		count int64
	}

	var queries []queryFreq
	for hash, count := range hotQueries {
		queries = append(queries, queryFreq{hash, count})
	}

	sort.Slice(queries, func(i, j int) bool {
		return queries[i].count > queries[j].count
	})

	// Promouvoir les top 10 en cache local
	for i := 0; i < 10 && i < len(queries); i++ {
		q.promoteToHotCache(queries[i].hash)
	}

	q.logger.Debug("Analyse des requêtes chaudes terminée",
		zap.Int("total_queries", len(queries)),
		zap.Int("promoted", min(10, len(queries))))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// promoteToHotCache promeut une requête vers le cache hot
func (q *QueryCacheService) promoteToHotCache(queryHash string) {
	if _, exists := q.hotQueries.Load(queryHash); !exists {
		q.hotQueries.Store(queryHash, true)

		// Marquer pour cache local lors du prochain accès
		q.logger.Debug("Requête promue vers cache hot",
			zap.String("query_hash", queryHash))
	}
}

// ============================================================================
// NETTOYAGE ET MAINTENANCE
// ============================================================================

// startCacheCleaner démarre le nettoyage périodique du cache
func (q *QueryCacheService) startCacheCleaner() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		q.cleanExpiredEntries()
	}
}

// cleanExpiredEntries nettoie les entrées expirées
func (q *QueryCacheService) cleanExpiredEntries() {
	now := time.Now()
	expiredKeys := []string{}

	// Nettoyer le cache local
	q.localCache.Range(func(key, value interface{}) bool {
		if cached, ok := value.(*CachedQuery); ok {
			if now.After(cached.ExpiresAt) {
				expiredKeys = append(expiredKeys, key.(string))
			}
		}
		return true
	})

	for _, key := range expiredKeys {
		q.localCache.Delete(key)
	}

	if len(expiredKeys) > 0 {
		q.logger.Debug("Cache local nettoyé",
			zap.Int("expired_entries", len(expiredKeys)))
	}
}

// ============================================================================
// MÉTRIQUES ET MONITORING
// ============================================================================

func (q *QueryCacheService) recordCacheHit(responseTime time.Duration, originalExecutionTime int64) {
	q.metrics.mutex.Lock()
	q.metrics.CacheHits++
	q.metrics.TotalSavedMs += originalExecutionTime - responseTime.Milliseconds()
	q.metrics.mutex.Unlock()
}

func (q *QueryCacheService) recordCacheMiss() {
	q.metrics.mutex.Lock()
	q.metrics.CacheMisses++
	q.metrics.mutex.Unlock()
}

func (q *QueryCacheService) recordQueryExecution(executionTime time.Duration) {
	q.metrics.mutex.Lock()
	q.metrics.AvgQueryTimeMs = (q.metrics.AvgQueryTimeMs + executionTime.Milliseconds()) / 2
	q.metrics.mutex.Unlock()
}

// GetQueryMetrics retourne les métriques de cache de requêtes
func (q *QueryCacheService) GetQueryMetrics() *QueryCacheMetrics {
	q.metrics.mutex.RLock()
	defer q.metrics.mutex.RUnlock()

	hotQueriesCopy := make(map[string]int64)
	for k, v := range q.metrics.HotQueries {
		hotQueriesCopy[k] = v
	}

	return &QueryCacheMetrics{
		QueryExecutions: q.metrics.QueryExecutions,
		CacheHits:       q.metrics.CacheHits,
		CacheMisses:     q.metrics.CacheMisses,
		LocalCacheHits:  q.metrics.LocalCacheHits,
		RedisCacheHits:  q.metrics.RedisCacheHits,
		TotalSavedMs:    q.metrics.TotalSavedMs,
		AvgQueryTimeMs:  q.metrics.AvgQueryTimeMs,
		HotQueries:      hotQueriesCopy,
	}
}

func (q *QueryCacheService) startMetricsReporter() {
	ticker := time.NewTicker(3 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		metrics := q.GetQueryMetrics()

		hitRatio := float64(0)
		if metrics.QueryExecutions > 0 {
			hitRatio = float64(metrics.CacheHits) / float64(metrics.QueryExecutions)
		}

		q.logger.Info("Métriques cache de requêtes",
			zap.Int64("executions", metrics.QueryExecutions),
			zap.Int64("cache_hits", metrics.CacheHits),
			zap.Float64("hit_ratio", hitRatio),
			zap.Int64("total_saved_ms", metrics.TotalSavedMs),
			zap.Int64("avg_query_time_ms", metrics.AvgQueryTimeMs),
			zap.Int("hot_queries_count", len(metrics.HotQueries)))
	}
}

// HealthCheck vérifie la santé du service de cache de requêtes
func (q *QueryCacheService) HealthCheck(ctx context.Context) error {
	// Test de fonctionnement de base
	testKey := "health_check_query"
	testData := map[string]interface{}{"test": "data"}

	// Test Redis
	if err := q.redis.Set(ctx, testKey, testData, 1*time.Second).Err(); err != nil {
		return fmt.Errorf("échec test Redis: %w", err)
	}

	if err := q.redis.Del(ctx, testKey).Err(); err != nil {
		return fmt.Errorf("échec nettoyage test Redis: %w", err)
	}

	return nil
}
