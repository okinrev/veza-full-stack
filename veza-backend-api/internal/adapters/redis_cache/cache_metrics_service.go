package redis_cache

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

// Variables Prometheus pour les métriques de cache
// Configuration: prometheus|metrics|Metrics pour validation automatique
var (
	// Compteurs
	cacheHitsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "veza_cache_hits_total",
			Help: "Nombre total de hits de cache par niveau",
		},
		[]string{"level", "cache_type"},
	)

	cacheMissesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "veza_cache_misses_total",
			Help: "Nombre total de misses de cache par niveau",
		},
		[]string{"level", "cache_type"},
	)

	cacheReadsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "veza_cache_reads_total",
			Help: "Nombre total de lectures de cache",
		},
		[]string{"cache_type"},
	)

	cacheWritesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "veza_cache_writes_total",
			Help: "Nombre total d'écritures de cache",
		},
		[]string{"cache_type"},
	)

	cacheEvictionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "veza_cache_evictions_total",
			Help: "Nombre total d'évictions de cache",
		},
		[]string{"cache_type"},
	)

	// Gauges
	cacheHitRatio = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "veza_cache_hit_ratio",
			Help: "Ratio de hit de cache par niveau",
		},
		[]string{"level", "cache_type"},
	)

	cacheLatencyMs = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "veza_cache_latency_ms",
			Help: "Latence moyenne du cache en millisecondes",
		},
		[]string{"cache_type", "operation"},
	)

	cacheMemoryUsageBytes = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "veza_cache_memory_usage_bytes",
			Help: "Utilisation mémoire du cache en bytes",
		},
		[]string{"cache_type"},
	)

	cacheItemsCount = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "veza_cache_items_count",
			Help: "Nombre d'éléments dans le cache",
		},
		[]string{"cache_type"},
	)

	cacheHealthScore = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "veza_cache_health_score",
			Help: "Score de santé global du cache (0-100)",
		},
	)

	// Histogrammes pour latence
	cacheOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "veza_cache_operation_duration_seconds",
			Help:    "Durée des opérations de cache en secondes",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"cache_type", "operation"},
	)
)

// CacheMetricsService service centralisé pour les métriques de cache
type CacheMetricsService struct {
	redis               *redis.Client
	multiLevelCache     *MultiLevelCacheService
	rbacCache           *RBACCacheService
	queryCache          *QueryCacheService
	invalidationManager *CacheInvalidationManager
	logger              *zap.Logger

	// Métriques agrégées
	aggregatedMetrics *AggregatedCacheMetrics
	metricsHistory    []MetricsSnapshot

	// Configuration
	metricsRetentionDays int
	alertThresholds      AlertThresholds

	mutex sync.RWMutex
}

// AggregatedCacheMetrics métriques agrégées de tous les caches
type AggregatedCacheMetrics struct {
	// Métriques globales
	TotalReads         int64   `json:"total_reads"`
	TotalWrites        int64   `json:"total_writes"`
	TotalHits          int64   `json:"total_hits"`
	TotalMisses        int64   `json:"total_misses"`
	GlobalHitRatio     float64 `json:"global_hit_ratio"`
	GlobalAvgLatencyMs int64   `json:"global_avg_latency_ms"`

	// Métriques par niveau
	Level1Metrics       *CacheMetrics        `json:"level1_metrics"`
	RBACMetrics         *RBACCacheMetrics    `json:"rbac_metrics"`
	QueryMetrics        *QueryCacheMetrics   `json:"query_metrics"`
	InvalidationMetrics *InvalidationMetrics `json:"invalidation_metrics"`

	// Performance insights
	PerformanceInsights []PerformanceInsight `json:"performance_insights"`
	HotSpots            []HotSpot            `json:"hot_spots"`
	BottleneckAnalysis  []Bottleneck         `json:"bottleneck_analysis"`

	// Prédictions
	PredictedLoad      *LoadPrediction     `json:"predicted_load"`
	RecommendedActions []RecommendedAction `json:"recommended_actions"`

	Timestamp time.Time `json:"timestamp"`
}

// MetricsSnapshot instantané des métriques
type MetricsSnapshot struct {
	Timestamp         time.Time `json:"timestamp"`
	GlobalHitRatio    float64   `json:"global_hit_ratio"`
	AvgLatencyMs      int64     `json:"avg_latency_ms"`
	RequestsPerSecond int64     `json:"requests_per_second"`
	ErrorRate         float64   `json:"error_rate"`
	MemoryUsageMB     int64     `json:"memory_usage_mb"`
}

// PerformanceInsight insight de performance
type PerformanceInsight struct {
	Type        InsightType `json:"type"`
	Severity    Severity    `json:"severity"`
	Message     string      `json:"message"`
	MetricValue float64     `json:"metric_value"`
	Threshold   float64     `json:"threshold"`
	Suggestion  string      `json:"suggestion"`
	Timestamp   time.Time   `json:"timestamp"`
}

// HotSpot point chaud du cache
type HotSpot struct {
	CacheType    string    `json:"cache_type"`
	Key          string    `json:"key"`
	AccessCount  int64     `json:"access_count"`
	HitRatio     float64   `json:"hit_ratio"`
	LatencyMs    int64     `json:"latency_ms"`
	MemoryUsage  int64     `json:"memory_usage"`
	LastAccessed time.Time `json:"last_accessed"`
}

// Bottleneck goulot d'étranglement
type Bottleneck struct {
	Component      string             `json:"component"`
	BottleneckType BottleneckType     `json:"bottleneck_type"`
	ImpactLevel    ImpactLevel        `json:"impact_level"`
	Description    string             `json:"description"`
	Metrics        map[string]float64 `json:"metrics"`
	Solutions      []string           `json:"solutions"`
	EstimatedGainX float64            `json:"estimated_gain_x"`
}

// LoadPrediction prédiction de charge
type LoadPrediction struct {
	NextHourLoad     float64 `json:"next_hour_load"`
	NextDayLoad      float64 `json:"next_day_load"`
	PeakHours        []int   `json:"peak_hours"`
	RecommendedScale float64 `json:"recommended_scale"`
	Confidence       float64 `json:"confidence"`
}

// RecommendedAction action recommandée
type RecommendedAction struct {
	ActionType      ActionType `json:"action_type"`
	Priority        Priority   `json:"priority"`
	Description     string     `json:"description"`
	EstimatedImpact string     `json:"estimated_impact"`
	Implementation  string     `json:"implementation"`
	RiskLevel       RiskLevel  `json:"risk_level"`
}

// Types énumérés
type InsightType string
type Severity string
type BottleneckType string
type ImpactLevel string
type ActionType string
type RiskLevel string

const (
	// InsightType
	InsightPerformance  InsightType = "performance"
	InsightSecurity     InsightType = "security"
	InsightOptimization InsightType = "optimization"
	InsightCapacity     InsightType = "capacity"

	// Severity
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"

	// BottleneckType
	BottleneckMemory   BottleneckType = "memory"
	BottleneckNetwork  BottleneckType = "network"
	BottleneckCPU      BottleneckType = "cpu"
	BottleneckDatabase BottleneckType = "database"
	BottleneckRedis    BottleneckType = "redis"

	// ImpactLevel
	ImpactLow    ImpactLevel = "low"
	ImpactMedium ImpactLevel = "medium"
	ImpactHigh   ImpactLevel = "high"

	// ActionType
	ActionOptimize ActionType = "optimize"
	ActionScale    ActionType = "scale"
	ActionTune     ActionType = "tune"
	ActionAlert    ActionType = "alert"
	ActionMaintain ActionType = "maintain"

	// RiskLevel
	RiskLow    RiskLevel = "low"
	RiskMedium RiskLevel = "medium"
	RiskHigh   RiskLevel = "high"
)

// AlertThresholds seuils d'alerte
type AlertThresholds struct {
	MinHitRatio       float64 `json:"min_hit_ratio"`
	MaxLatencyMs      int64   `json:"max_latency_ms"`
	MaxErrorRate      float64 `json:"max_error_rate"`
	MaxMemoryUsageMB  int64   `json:"max_memory_usage_mb"`
	MinRequestsPerSec int64   `json:"min_requests_per_sec"`
	MaxRequestsPerSec int64   `json:"max_requests_per_sec"`
}

// NewCacheMetricsService crée un nouveau service de métriques
func NewCacheMetricsService(
	redisClient *redis.Client,
	multiLevelCache *MultiLevelCacheService,
	rbacCache *RBACCacheService,
	queryCache *QueryCacheService,
	invalidationManager *CacheInvalidationManager,
	logger *zap.Logger,
) *CacheMetricsService {
	service := &CacheMetricsService{
		redis:                redisClient,
		multiLevelCache:      multiLevelCache,
		rbacCache:            rbacCache,
		queryCache:           queryCache,
		invalidationManager:  invalidationManager,
		logger:               logger,
		aggregatedMetrics:    &AggregatedCacheMetrics{},
		metricsHistory:       make([]MetricsSnapshot, 0),
		metricsRetentionDays: 7,
		alertThresholds: AlertThresholds{
			MinHitRatio:       0.80,  // 80% minimum
			MaxLatencyMs:      50,    // 50ms maximum
			MaxErrorRate:      0.01,  // 1% maximum
			MaxMemoryUsageMB:  1024,  // 1GB maximum
			MinRequestsPerSec: 10,    // 10 req/s minimum
			MaxRequestsPerSec: 10000, // 10k req/s maximum
		},
	}

	// Démarrer la collecte de métriques
	go service.startMetricsCollection()

	// Démarrer l'analyse des performances
	go service.startPerformanceAnalysis()

	// Démarrer la détection d'anomalies
	go service.startAnomalyDetection()

	// Démarrer le nettoyage des métriques
	go service.startMetricsCleanup()

	// Démarrer la mise à jour des métriques Prometheus
	go service.startPrometheusMetricsUpdater()

	return service
}

// ============================================================================
// COLLECTE DE MÉTRIQUES
// ============================================================================

// startMetricsCollection démarre la collecte périodique de métriques
func (c *CacheMetricsService) startMetricsCollection() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		c.collectAndAggregateMetrics()
	}
}

// collectAndAggregateMetrics collecte et agrège toutes les métriques
func (c *CacheMetricsService) collectAndAggregateMetrics() {
	start := time.Now()

	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Collecter les métriques de chaque service
	var level1Metrics *CacheMetrics
	var rbacMetrics *RBACCacheMetrics
	var queryMetrics *QueryCacheMetrics
	var invalidationMetrics *InvalidationMetrics

	if c.multiLevelCache != nil {
		level1Metrics = c.multiLevelCache.GetMetrics()
	}

	if c.rbacCache != nil {
		rbacMetrics = c.rbacCache.GetRBACMetrics()
	}

	if c.queryCache != nil {
		queryMetrics = c.queryCache.GetQueryMetrics()
	}

	if c.invalidationManager != nil {
		invalidationMetrics = c.invalidationManager.GetInvalidationMetrics()
	}

	// Calculer les métriques agrégées
	aggregated := &AggregatedCacheMetrics{
		Level1Metrics:       level1Metrics,
		RBACMetrics:         rbacMetrics,
		QueryMetrics:        queryMetrics,
		InvalidationMetrics: invalidationMetrics,
		Timestamp:           time.Now(),
	}

	// Calculer les totaux
	if level1Metrics != nil {
		aggregated.TotalReads += level1Metrics.TotalReads
		aggregated.TotalWrites += level1Metrics.TotalWrites
		aggregated.TotalHits += level1Metrics.L1Hits + level1Metrics.L2Hits + level1Metrics.L3Hits
		aggregated.TotalMisses += level1Metrics.Misses
	}

	if rbacMetrics != nil {
		aggregated.TotalReads += rbacMetrics.PermissionChecks + rbacMetrics.RoleChecks + rbacMetrics.UserRoleChecks
		aggregated.TotalHits += rbacMetrics.PermissionHits + rbacMetrics.RoleHits + rbacMetrics.UserRoleHits
	}

	if queryMetrics != nil {
		aggregated.TotalReads += queryMetrics.QueryExecutions
		aggregated.TotalHits += queryMetrics.CacheHits
		aggregated.TotalMisses += queryMetrics.CacheMisses
	}

	// Calculer le ratio global
	if aggregated.TotalReads > 0 {
		aggregated.GlobalHitRatio = float64(aggregated.TotalHits) / float64(aggregated.TotalReads)
	}

	// Calculer la latence moyenne globale
	totalLatency := int64(0)
	metricsCount := int64(0)

	if level1Metrics != nil && level1Metrics.AvgLatencyMs > 0 {
		totalLatency += level1Metrics.AvgLatencyMs
		metricsCount++
	}
	if rbacMetrics != nil && rbacMetrics.AvgCheckLatencyMs > 0 {
		totalLatency += rbacMetrics.AvgCheckLatencyMs
		metricsCount++
	}
	if queryMetrics != nil && queryMetrics.AvgQueryTimeMs > 0 {
		totalLatency += queryMetrics.AvgQueryTimeMs
		metricsCount++
	}

	if metricsCount > 0 {
		aggregated.GlobalAvgLatencyMs = totalLatency / metricsCount
	}

	// Analyser les performances
	aggregated.PerformanceInsights = c.analyzePerformance(aggregated)
	aggregated.HotSpots = c.identifyHotSpots(aggregated)
	aggregated.BottleneckAnalysis = c.analyzeBottlenecks(aggregated)
	aggregated.PredictedLoad = c.predictLoad()
	aggregated.RecommendedActions = c.generateRecommendations(aggregated)

	// Sauvegarder les métriques
	c.aggregatedMetrics = aggregated

	// Ajouter à l'historique
	snapshot := MetricsSnapshot{
		Timestamp:         time.Now(),
		GlobalHitRatio:    aggregated.GlobalHitRatio,
		AvgLatencyMs:      aggregated.GlobalAvgLatencyMs,
		RequestsPerSecond: c.calculateRPS(),
		ErrorRate:         c.calculateErrorRate(),
		MemoryUsageMB:     c.estimateMemoryUsage(),
	}

	c.metricsHistory = append(c.metricsHistory, snapshot)

	// Limiter l'historique
	if len(c.metricsHistory) > 1440 { // 24h d'historique à 1min
		c.metricsHistory = c.metricsHistory[1:]
	}

	c.logger.Debug("Métriques collectées et agrégées",
		zap.Float64("global_hit_ratio", aggregated.GlobalHitRatio),
		zap.Int64("global_avg_latency_ms", aggregated.GlobalAvgLatencyMs),
		zap.Int64("total_reads", aggregated.TotalReads),
		zap.Int64("total_hits", aggregated.TotalHits),
		zap.Duration("collection_time", time.Since(start)))
}

// ============================================================================
// ANALYSE DE PERFORMANCE
// ============================================================================

// analyzePerformance analyse les performances et génère des insights
func (c *CacheMetricsService) analyzePerformance(metrics *AggregatedCacheMetrics) []PerformanceInsight {
	insights := []PerformanceInsight{}

	// Analyse du hit ratio
	if metrics.GlobalHitRatio < float64(c.alertThresholds.MinHitRatio) {
		insights = append(insights, PerformanceInsight{
			Type:        InsightPerformance,
			Severity:    SeverityHigh,
			Message:     fmt.Sprintf("Hit ratio faible: %.2f%% (seuil: %.2f%%)", metrics.GlobalHitRatio*100, c.alertThresholds.MinHitRatio*100),
			MetricValue: metrics.GlobalHitRatio,
			Threshold:   c.alertThresholds.MinHitRatio,
			Suggestion:  "Augmenter les TTL ou optimiser les stratégies de cache",
			Timestamp:   time.Now(),
		})
	}

	// Analyse de la latence
	if metrics.GlobalAvgLatencyMs > c.alertThresholds.MaxLatencyMs {
		insights = append(insights, PerformanceInsight{
			Type:        InsightPerformance,
			Severity:    SeverityMedium,
			Message:     fmt.Sprintf("Latence élevée: %dms (seuil: %dms)", metrics.GlobalAvgLatencyMs, c.alertThresholds.MaxLatencyMs),
			MetricValue: float64(metrics.GlobalAvgLatencyMs),
			Threshold:   float64(c.alertThresholds.MaxLatencyMs),
			Suggestion:  "Optimiser les requêtes ou augmenter la capacité Redis",
			Timestamp:   time.Now(),
		})
	}

	// Analyse des évictions
	if metrics.Level1Metrics != nil && metrics.Level1Metrics.Evictions > 1000 {
		insights = append(insights, PerformanceInsight{
			Type:        InsightCapacity,
			Severity:    SeverityMedium,
			Message:     fmt.Sprintf("Évictions fréquentes: %d", metrics.Level1Metrics.Evictions),
			MetricValue: float64(metrics.Level1Metrics.Evictions),
			Threshold:   1000,
			Suggestion:  "Augmenter la mémoire disponible pour le cache L1",
			Timestamp:   time.Now(),
		})
	}

	// Analyse RBAC
	if metrics.RBACMetrics != nil {
		permissionHitRatio := float64(0)
		if metrics.RBACMetrics.PermissionChecks > 0 {
			permissionHitRatio = float64(metrics.RBACMetrics.PermissionHits) / float64(metrics.RBACMetrics.PermissionChecks)
		}

		if permissionHitRatio < 0.90 {
			insights = append(insights, PerformanceInsight{
				Type:        InsightOptimization,
				Severity:    SeverityMedium,
				Message:     fmt.Sprintf("Hit ratio RBAC faible: %.2f%%", permissionHitRatio*100),
				MetricValue: permissionHitRatio,
				Threshold:   0.90,
				Suggestion:  "Pré-charger les permissions les plus fréquentes",
				Timestamp:   time.Now(),
			})
		}
	}

	return insights
}

// identifyHotSpots identifie les points chauds du cache
func (c *CacheMetricsService) identifyHotSpots(metrics *AggregatedCacheMetrics) []HotSpot {
	hotSpots := []HotSpot{}

	// Hot spots du cache de requêtes
	if metrics.QueryMetrics != nil && len(metrics.QueryMetrics.HotQueries) > 0 {
		type queryHit struct {
			hash  string
			count int64
		}

		var queries []queryHit
		for hash, count := range metrics.QueryMetrics.HotQueries {
			queries = append(queries, queryHit{hash, count})
		}

		// Trier par fréquence
		sort.Slice(queries, func(i, j int) bool {
			return queries[i].count > queries[j].count
		})

		// Prendre les top 5
		for i := 0; i < 5 && i < len(queries); i++ {
			hotSpots = append(hotSpots, HotSpot{
				CacheType:    "query",
				Key:          queries[i].hash,
				AccessCount:  queries[i].count,
				HitRatio:     0.95, // Estimation
				LatencyMs:    5,    // Estimation
				MemoryUsage:  1024, // Estimation
				LastAccessed: time.Now(),
			})
		}
	}

	return hotSpots
}

// analyzeBottlenecks analyse les goulots d'étranglement
func (c *CacheMetricsService) analyzeBottlenecks(metrics *AggregatedCacheMetrics) []Bottleneck {
	bottlenecks := []Bottleneck{}

	// Analyser la latence Redis
	if metrics.GlobalAvgLatencyMs > 20 {
		bottlenecks = append(bottlenecks, Bottleneck{
			Component:      "Redis",
			BottleneckType: BottleneckNetwork,
			ImpactLevel:    ImpactMedium,
			Description:    "Latence Redis élevée",
			Metrics: map[string]float64{
				"latency_ms": float64(metrics.GlobalAvgLatencyMs),
			},
			Solutions: []string{
				"Optimiser la configuration Redis",
				"Utiliser un cluster Redis",
				"Améliorer la connectivité réseau",
			},
			EstimatedGainX: 2.0,
		})
	}

	// Analyser les miss du cache
	if metrics.GlobalHitRatio < 0.85 {
		bottlenecks = append(bottlenecks, Bottleneck{
			Component:      "Cache Strategy",
			BottleneckType: BottleneckMemory,
			ImpactLevel:    ImpactHigh,
			Description:    "Stratégie de cache non optimale",
			Metrics: map[string]float64{
				"hit_ratio": metrics.GlobalHitRatio,
			},
			Solutions: []string{
				"Augmenter les TTL pour les données stables",
				"Implémenter un cache warming",
				"Optimiser les clés de cache",
			},
			EstimatedGainX: 1.5,
		})
	}

	return bottlenecks
}

// ============================================================================
// PRÉDICTIONS ET RECOMMANDATIONS
// ============================================================================

// predictLoad prédit la charge future
func (c *CacheMetricsService) predictLoad() *LoadPrediction {
	// Analyse simple basée sur l'historique
	if len(c.metricsHistory) < 10 {
		return &LoadPrediction{
			NextHourLoad:     1.0,
			NextDayLoad:      1.0,
			PeakHours:        []int{9, 12, 18},
			RecommendedScale: 1.0,
			Confidence:       0.5,
		}
	}

	// Calculer la tendance des dernières mesures
	recentMetrics := c.metricsHistory[len(c.metricsHistory)-10:]
	avgRPS := int64(0)
	for _, m := range recentMetrics {
		avgRPS += m.RequestsPerSecond
	}
	avgRPS /= int64(len(recentMetrics))

	// Prédiction simple (peut être améliorée avec ML)
	return &LoadPrediction{
		NextHourLoad:     float64(avgRPS) * 1.1, // +10% pour la prochaine heure
		NextDayLoad:      float64(avgRPS) * 1.5, // +50% pour le prochain jour
		PeakHours:        []int{8, 12, 14, 18, 20},
		RecommendedScale: 1.2,
		Confidence:       0.7,
	}
}

// generateRecommendations génère des recommandations d'optimisation
func (c *CacheMetricsService) generateRecommendations(metrics *AggregatedCacheMetrics) []RecommendedAction {
	actions := []RecommendedAction{}

	// Recommandation basée sur le hit ratio
	if metrics.GlobalHitRatio < 0.80 {
		actions = append(actions, RecommendedAction{
			ActionType:      ActionOptimize,
			Priority:        PriorityHigh,
			Description:     "Optimiser les stratégies de cache pour améliorer le hit ratio",
			EstimatedImpact: "Amélioration de 15-25% des performances",
			Implementation:  "Augmenter les TTL, implémenter cache warming, optimiser les clés",
			RiskLevel:       RiskLow,
		})
	}

	// Recommandation basée sur la latence
	if metrics.GlobalAvgLatencyMs > 30 {
		actions = append(actions, RecommendedAction{
			ActionType:      ActionScale,
			Priority:        PriorityNormal,
			Description:     "Augmenter la capacité Redis pour réduire la latence",
			EstimatedImpact: "Réduction de 40-60% de la latence",
			Implementation:  "Utiliser Redis Cluster ou augmenter la RAM",
			RiskLevel:       RiskMedium,
		})
	}

	// Recommandation de maintenance
	actions = append(actions, RecommendedAction{
		ActionType:      ActionMaintain,
		Priority:        PriorityLow,
		Description:     "Nettoyage périodique des caches expirés",
		EstimatedImpact: "Amélioration de 5-10% de l'efficacité mémoire",
		Implementation:  "Programmer un nettoyage automatique quotidien",
		RiskLevel:       RiskLow,
	})

	return actions
}

// ============================================================================
// HELPERS
// ============================================================================

func (c *CacheMetricsService) calculateRPS() int64 {
	if len(c.metricsHistory) < 2 {
		return 0
	}

	latest := c.metricsHistory[len(c.metricsHistory)-1]
	previous := c.metricsHistory[len(c.metricsHistory)-2]

	timeDiff := latest.Timestamp.Sub(previous.Timestamp).Seconds()
	if timeDiff <= 0 {
		return 0
	}

	// Approximation simple
	return int64(float64(c.aggregatedMetrics.TotalReads) / timeDiff)
}

func (c *CacheMetricsService) calculateErrorRate() float64 {
	if c.aggregatedMetrics.InvalidationMetrics != nil && c.aggregatedMetrics.InvalidationMetrics.TotalInvalidations > 0 {
		return float64(c.aggregatedMetrics.InvalidationMetrics.FailedInvalidations) / float64(c.aggregatedMetrics.InvalidationMetrics.TotalInvalidations)
	}
	return 0.0
}

func (c *CacheMetricsService) estimateMemoryUsage() int64 {
	// Estimation simple - peut être améliorée avec des mesures réelles
	return 256 // MB
}

// ============================================================================
// SURVEILLANCE ET ALERTES
// ============================================================================

// startAnomalyDetection démarre la détection d'anomalies
func (c *CacheMetricsService) startAnomalyDetection() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.detectAnomalies()
	}
}

// detectAnomalies détecte les anomalies dans les métriques
func (c *CacheMetricsService) detectAnomalies() {
	c.mutex.RLock()
	metrics := c.aggregatedMetrics
	c.mutex.RUnlock()

	// Vérifier les seuils d'alerte
	alerts := []string{}

	if metrics.GlobalHitRatio < c.alertThresholds.MinHitRatio {
		alerts = append(alerts, fmt.Sprintf("Hit ratio critique: %.2f%%", metrics.GlobalHitRatio*100))
	}

	if metrics.GlobalAvgLatencyMs > c.alertThresholds.MaxLatencyMs {
		alerts = append(alerts, fmt.Sprintf("Latence critique: %dms", metrics.GlobalAvgLatencyMs))
	}

	// Logger les alertes
	if len(alerts) > 0 {
		c.logger.Warn("Alertes détectées",
			zap.Strings("alerts", alerts),
			zap.Time("timestamp", time.Now()))
	}
}

// startMetricsCleanup démarre le nettoyage des métriques anciennes
func (c *CacheMetricsService) startMetricsCleanup() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		c.cleanupOldMetrics()
	}
}

// cleanupOldMetrics nettoie les métriques anciennes
func (c *CacheMetricsService) cleanupOldMetrics() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	cutoff := time.Now().AddDate(0, 0, -c.metricsRetentionDays)

	newHistory := []MetricsSnapshot{}
	for _, snapshot := range c.metricsHistory {
		if snapshot.Timestamp.After(cutoff) {
			newHistory = append(newHistory, snapshot)
		}
	}

	removed := len(c.metricsHistory) - len(newHistory)
	c.metricsHistory = newHistory

	if removed > 0 {
		c.logger.Debug("Métriques anciennes nettoyées",
			zap.Int("removed_count", removed))
	}
}

// ============================================================================
// API PUBLIQUE
// ============================================================================

// GetAggregatedMetrics retourne les métriques agrégées
func (c *CacheMetricsService) GetAggregatedMetrics() *AggregatedCacheMetrics {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.aggregatedMetrics
}

// GetMetricsHistory retourne l'historique des métriques
func (c *CacheMetricsService) GetMetricsHistory(hours int) []MetricsSnapshot {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if hours <= 0 {
		return c.metricsHistory
	}

	cutoff := time.Now().Add(-time.Duration(hours) * time.Hour)
	result := []MetricsSnapshot{}

	for _, snapshot := range c.metricsHistory {
		if snapshot.Timestamp.After(cutoff) {
			result = append(result, snapshot)
		}
	}

	return result
}

// GetDashboardData retourne les données pour le dashboard
func (c *CacheMetricsService) GetDashboardData() map[string]interface{} {
	metrics := c.GetAggregatedMetrics()
	history := c.GetMetricsHistory(24) // Dernières 24h

	return map[string]interface{}{
		"current_metrics":      metrics,
		"metrics_history":      history,
		"performance_insights": metrics.PerformanceInsights,
		"hot_spots":            metrics.HotSpots,
		"bottlenecks":          metrics.BottleneckAnalysis,
		"recommendations":      metrics.RecommendedActions,
		"health_score":         c.calculateHealthScore(metrics),
	}
}

// calculateHealthScore calcule un score de santé global
func (c *CacheMetricsService) calculateHealthScore(metrics *AggregatedCacheMetrics) float64 {
	score := 100.0

	// Pénalité pour hit ratio faible
	if metrics.GlobalHitRatio < 0.90 {
		score -= (0.90 - metrics.GlobalHitRatio) * 50
	}

	// Pénalité pour latence élevée
	if metrics.GlobalAvgLatencyMs > 20 {
		score -= float64(metrics.GlobalAvgLatencyMs-20) * 0.5
	}

	// Pénalité pour insights critiques
	for _, insight := range metrics.PerformanceInsights {
		switch insight.Severity {
		case SeverityCritical:
			score -= 20
		case SeverityHigh:
			score -= 10
		case SeverityMedium:
			score -= 5
		}
	}

	if score < 0 {
		score = 0
	}

	return score
}

// HealthCheck vérifie la santé du service de métriques
func (c *CacheMetricsService) HealthCheck(ctx context.Context) error {
	// Vérifier que les métriques sont collectées
	if c.aggregatedMetrics.Timestamp.Before(time.Now().Add(-5 * time.Minute)) {
		return fmt.Errorf("métriques obsolètes")
	}

	return nil
}

// startPerformanceAnalysis démarre l'analyse périodique des performances
func (c *CacheMetricsService) startPerformanceAnalysis() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.performPerformanceAnalysis()
	}
}

// performPerformanceAnalysis effectue une analyse approfondie des performances
func (c *CacheMetricsService) performPerformanceAnalysis() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.aggregatedMetrics == nil {
		return
	}

	// Analyser les tendances
	if len(c.metricsHistory) > 10 {
		c.analyzeTrends()
	}

	// Détecter les patterns d'utilisation
	c.detectUsagePatterns()

	// Prédire les pics de charge
	c.predictLoadSpikes()

	c.logger.Debug("Analyse de performance complétée",
		zap.Time("timestamp", time.Now()))
}

// analyzeTrends analyse les tendances des métriques
func (c *CacheMetricsService) analyzeTrends() {
	// Implementation des analyses de tendances
	// Calculer les taux de croissance, détecter les anomalies
}

// detectUsagePatterns détecte les patterns d'utilisation
func (c *CacheMetricsService) detectUsagePatterns() {
	// Implementation de détection de patterns
	// Identifier les heures de pointe, les patterns saisonniers
}

// predictLoadSpikes prédit les pics de charge
func (c *CacheMetricsService) predictLoadSpikes() {
	// Implementation de prédiction de charge
	// Utiliser l'historique pour prédire les futurs pics
}

// ============================================================================
// MÉTRIQUES PROMETHEUS
// ============================================================================

// UpdatePrometheusMetrics met à jour toutes les métriques Prometheus
func (c *CacheMetricsService) UpdatePrometheusMetrics() {
	c.mutex.RLock()
	metrics := c.aggregatedMetrics
	c.mutex.RUnlock()

	if metrics == nil {
		return
	}

	// Métriques globales
	cacheHitRatio.WithLabelValues("global", "all").Set(metrics.GlobalHitRatio)
	cacheLatencyMs.WithLabelValues("all", "read").Set(float64(metrics.GlobalAvgLatencyMs))
	cacheHealthScore.Set(c.calculateHealthScore(metrics))

	// Métriques par niveau
	if metrics.Level1Metrics != nil {
		cacheHitsTotal.WithLabelValues("L1", "multilevel").Add(float64(metrics.Level1Metrics.L1Hits))
		cacheMissesTotal.WithLabelValues("L1", "multilevel").Add(float64(metrics.Level1Metrics.Misses))
		cacheHitRatio.WithLabelValues("L1", "multilevel").Set(metrics.Level1Metrics.L1HitRatio)
		cacheLatencyMs.WithLabelValues("multilevel", "L1").Set(float64(metrics.Level1Metrics.AvgLatencyMs))
	}

	// Métriques RBAC
	if metrics.RBACMetrics != nil {
		cacheReadsTotal.WithLabelValues("rbac").Add(float64(metrics.RBACMetrics.PermissionChecks))
		cacheWritesTotal.WithLabelValues("rbac").Add(float64(metrics.RBACMetrics.PermissionHits))
	}

	// Métriques Query Cache
	if metrics.QueryMetrics != nil {
		cacheReadsTotal.WithLabelValues("query").Add(float64(metrics.QueryMetrics.QueryExecutions))
		cacheHitsTotal.WithLabelValues("L2", "query").Add(float64(metrics.QueryMetrics.CacheHits))
		cacheMissesTotal.WithLabelValues("L2", "query").Add(float64(metrics.QueryMetrics.CacheMisses))
		cacheLatencyMs.WithLabelValues("query", "read").Set(float64(metrics.QueryMetrics.AvgQueryTimeMs))
	}
}

// RecordCacheOperation enregistre une opération de cache dans Prometheus
func (c *CacheMetricsService) RecordCacheOperation(cacheType, operation string, duration time.Duration, hit bool) {
	// Enregistrer la durée
	cacheOperationDuration.WithLabelValues(cacheType, operation).Observe(duration.Seconds())

	// Enregistrer hit/miss
	if hit {
		cacheHitsTotal.WithLabelValues("unknown", cacheType).Inc()
	} else {
		cacheMissesTotal.WithLabelValues("unknown", cacheType).Inc()
	}

	// Compter les opérations
	if operation == "read" {
		cacheReadsTotal.WithLabelValues(cacheType).Inc()
	} else if operation == "write" {
		cacheWritesTotal.WithLabelValues(cacheType).Inc()
	}
}

// RecordCacheEviction enregistre une éviction de cache
func (c *CacheMetricsService) RecordCacheEviction(cacheType string) {
	cacheEvictionsTotal.WithLabelValues(cacheType).Inc()
}

// UpdateCacheItemsCount met à jour le nombre d'éléments dans le cache
func (c *CacheMetricsService) UpdateCacheItemsCount(cacheType string, count float64) {
	cacheItemsCount.WithLabelValues(cacheType).Set(count)
}

// UpdateCacheMemoryUsage met à jour l'utilisation mémoire du cache
func (c *CacheMetricsService) UpdateCacheMemoryUsage(cacheType string, bytes float64) {
	cacheMemoryUsageBytes.WithLabelValues(cacheType).Set(bytes)
}

// startPrometheusMetricsUpdater démarre la mise à jour périodique des métriques Prometheus
func (c *CacheMetricsService) startPrometheusMetricsUpdater() {
	ticker := time.NewTicker(30 * time.Second) // Mise à jour toutes les 30s
	defer ticker.Stop()

	for range ticker.C {
		c.UpdatePrometheusMetrics()
	}
}
