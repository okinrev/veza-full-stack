package database

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

// DatabaseOptimizationManager orchestrateur principal des optimisations database
type DatabaseOptimizationManager struct {
	// Services d'optimisation
	connectionPool    *ConnectionPoolService
	indexOptimizer    *IndexOptimizationService
	queryOptimizer    *QueryOptimizationService
	paginationService *PaginationService
	analyticsReplica  *AnalyticsReplicaService
	
	// Configuration
	config *OptimizationConfig
	logger *zap.Logger
	
	// Métriques globales
	metrics *OptimizationMetrics
	
	// État du manager
	mu        sync.RWMutex
	isRunning bool
	ctx       context.Context
	cancel    context.CancelFunc
}

// OptimizationConfig configuration globale des optimisations
type OptimizationConfig struct {
	// Connexions
	PrimaryDatabaseURL   string   `json:"primary_database_url"`
	ReadReplicaURLs      []string `json:"read_replica_urls"`
	
	// Optimisations
	EnableIndexOptimization   bool `json:"enable_index_optimization"`
	EnableQueryOptimization   bool `json:"enable_query_optimization"`
	EnablePagination         bool `json:"enable_pagination"`
	EnableAnalyticsReplicas  bool `json:"enable_analytics_replicas"`
	
	// Maintenance
	AutoOptimizeIndexes      bool          `json:"auto_optimize_indexes"`
	MaintenanceWindow        time.Duration `json:"maintenance_window"`
	OptimizationInterval     time.Duration `json:"optimization_interval"`
	HealthCheckInterval      time.Duration `json:"health_check_interval"`
	
	// Performance
	MaxConcurrentOptimizations int           `json:"max_concurrent_optimizations"`
	OptimizationTimeout        time.Duration `json:"optimization_timeout"`
}

// DefaultOptimizationConfig configuration par défaut pour 100k+ users
func DefaultOptimizationConfig() *OptimizationConfig {
	return &OptimizationConfig{
		// Optimisations activées
		EnableIndexOptimization:  true,
		EnableQueryOptimization:  true,
		EnablePagination:        true,
		EnableAnalyticsReplicas: true,
		
		// Maintenance automatique
		AutoOptimizeIndexes:     true,
		MaintenanceWindow:       2 * time.Hour,    // 2h de maintenance
		OptimizationInterval:    6 * time.Hour,    // Optimisation toutes les 6h
		HealthCheckInterval:     1 * time.Minute,  // Health check chaque minute
		
		// Performance
		MaxConcurrentOptimizations: 3,
		OptimizationTimeout:        30 * time.Minute,
	}
}

// OptimizationMetrics métriques globales d'optimisation
type OptimizationMetrics struct {
	// Status global
	optimizationStatus *prometheus.GaugeVec
	
	// Performance globale
	overallLatency     *prometheus.HistogramVec
	throughput         *prometheus.GaugeVec
	
	// Optimisations
	optimizationsApplied *prometheus.CounterVec
	optimizationErrors   *prometheus.CounterVec
	
	// Health
	serviceHealth      *prometheus.GaugeVec
	
	// Business metrics
	concurrentUsers    prometheus.Gauge
	queryPerformance   *prometheus.HistogramVec
}

// NewOptimizationMetrics crée les métriques Prometheus
func NewOptimizationMetrics() *OptimizationMetrics {
	return &OptimizationMetrics{
		optimizationStatus: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "db_optimization_status",
			Help: "Status des optimisations database (1=actif, 0=inactif)",
		}, []string{"optimization_type"}),
		overallLatency: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name: "db_overall_latency_seconds",
			Help: "Latence globale database après optimisations",
			Buckets: []float64{0.001, 0.01, 0.05, 0.1, 0.5, 1.0},
		}, []string{"operation_type"}),
		throughput: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "db_throughput_ops_per_second",
			Help: "Débit database (opérations/seconde)",
		}, []string{"operation_type"}),
		optimizationsApplied: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "db_optimizations_applied_total",
			Help: "Nombre d'optimisations appliquées",
		}, []string{"optimization_type", "status"}),
		optimizationErrors: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "db_optimization_errors_total",
			Help: "Erreurs lors des optimisations",
		}, []string{"optimization_type", "error_type"}),
		serviceHealth: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "db_optimization_service_health",
			Help: "Santé des services d'optimisation",
		}, []string{"service"}),
		concurrentUsers: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "db_concurrent_users",
			Help: "Nombre d'utilisateurs concurrents supportés",
		}),
		queryPerformance: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name: "db_query_performance_score",
			Help: "Score de performance des requêtes (0-1)",
		}, []string{"query_category"}),
	}
}

// NewDatabaseOptimizationManager crée un nouveau manager d'optimisation
func NewDatabaseOptimizationManager(config *OptimizationConfig, logger *zap.Logger) *DatabaseOptimizationManager {
	if config == nil {
		config = DefaultOptimizationConfig()
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
	return &DatabaseOptimizationManager{
		config:  config,
		logger:  logger,
		metrics: NewOptimizationMetrics(),
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Initialize initialise tous les services d'optimisation
func (m *DatabaseOptimizationManager) Initialize() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.logger.Info("🚀 Initialisation DatabaseOptimizationManager", 
		zap.String("primary_url", m.config.PrimaryDatabaseURL),
		zap.Int("read_replicas", len(m.config.ReadReplicaURLs)),
	)
	
	// 1. Initialiser le service de pool de connexions
	if err := m.initializeConnectionPool(); err != nil {
		return fmt.Errorf("erreur initialisation connection pool: %w", err)
	}
	
	// 2. Initialiser l'optimiseur d'index
	if m.config.EnableIndexOptimization {
		if err := m.initializeIndexOptimizer(); err != nil {
			return fmt.Errorf("erreur initialisation index optimizer: %w", err)
		}
	}
	
	// 3. Initialiser l'optimiseur de requêtes
	if m.config.EnableQueryOptimization {
		if err := m.initializeQueryOptimizer(); err != nil {
			return fmt.Errorf("erreur initialisation query optimizer: %w", err)
		}
	}
	
	// 4. Initialiser le service de pagination
	if m.config.EnablePagination {
		if err := m.initializePaginationService(); err != nil {
			return fmt.Errorf("erreur initialisation pagination: %w", err)
		}
	}
	
	// 5. Initialiser les read replicas analytics
	if m.config.EnableAnalyticsReplicas && len(m.config.ReadReplicaURLs) > 0 {
		if err := m.initializeAnalyticsReplicas(); err != nil {
			return fmt.Errorf("erreur initialisation analytics replicas: %w", err)
		}
	}
	
	// Démarrer les services de maintenance
	go m.startMaintenanceServices()
	
	// Démarrer le monitoring
	go m.startMonitoring()
	
	m.isRunning = true
	
	m.logger.Info("✅ DatabaseOptimizationManager initialisé avec succès")
	return nil
}

// initializeConnectionPool initialise le service de pool de connexions
func (m *DatabaseOptimizationManager) initializeConnectionPool() error {
	config := DefaultConnectionPoolConfig()
	config.ReadReplicaURLs = m.config.ReadReplicaURLs
	config.ReadReplicaEnabled = len(m.config.ReadReplicaURLs) > 0
	
	m.connectionPool = NewConnectionPoolService(config, m.logger)
	
	if err := m.connectionPool.Initialize(m.config.PrimaryDatabaseURL, m.config.ReadReplicaURLs...); err != nil {
		return err
	}
	
	m.metrics.serviceHealth.WithLabelValues("connection_pool").Set(1)
	m.metrics.optimizationStatus.WithLabelValues("connection_pool").Set(1)
	
	m.logger.Info("✅ Connection pool initialisé", 
		zap.Int("max_connections", config.MaxOpenConns),
		zap.Int("read_replicas", len(config.ReadReplicaURLs)),
	)
	
	return nil
}

// initializeIndexOptimizer initialise l'optimiseur d'index
func (m *DatabaseOptimizationManager) initializeIndexOptimizer() error {
	primaryDB := m.connectionPool.GetPrimaryDB()
	if primaryDB == nil {
		return fmt.Errorf("primary database non disponible")
	}
	
	m.indexOptimizer = NewIndexOptimizationService(primaryDB, m.logger)
	
	// Créer les index critiques en arrière-plan
	if m.config.AutoOptimizeIndexes {
		go func() {
			ctx, cancel := context.WithTimeout(m.ctx, m.config.OptimizationTimeout)
			defer cancel()
			
			if err := m.indexOptimizer.CreateCriticalIndexes(ctx); err != nil {
				m.logger.Error("Erreur création index critiques", zap.Error(err))
				m.metrics.optimizationErrors.WithLabelValues("index", "creation").Inc()
			} else {
				m.metrics.optimizationsApplied.WithLabelValues("index", "success").Inc()
			}
		}()
	}
	
	m.metrics.serviceHealth.WithLabelValues("index_optimizer").Set(1)
	m.metrics.optimizationStatus.WithLabelValues("index_optimization").Set(1)
	
	m.logger.Info("✅ Index optimizer initialisé")
	return nil
}

// initializeQueryOptimizer initialise l'optimiseur de requêtes
func (m *DatabaseOptimizationManager) initializeQueryOptimizer() error {
	primaryDB := m.connectionPool.GetPrimaryDB()
	if primaryDB == nil {
		return fmt.Errorf("primary database non disponible")
	}
	
	m.queryOptimizer = NewQueryOptimizationService(primaryDB, m.logger)
	
	m.metrics.serviceHealth.WithLabelValues("query_optimizer").Set(1)
	m.metrics.optimizationStatus.WithLabelValues("query_optimization").Set(1)
	
	m.logger.Info("✅ Query optimizer initialisé")
	return nil
}

// initializePaginationService initialise le service de pagination
func (m *DatabaseOptimizationManager) initializePaginationService() error {
	primaryDB := m.connectionPool.GetPrimaryDB()
	if primaryDB == nil {
		return fmt.Errorf("primary database non disponible")
	}
	
	m.paginationService = NewPaginationService(primaryDB, m.logger)
	
	m.metrics.serviceHealth.WithLabelValues("pagination").Set(1)
	m.metrics.optimizationStatus.WithLabelValues("pagination").Set(1)
	
	m.logger.Info("✅ Pagination service initialisé")
	return nil
}

// initializeAnalyticsReplicas initialise les read replicas pour analytics
func (m *DatabaseOptimizationManager) initializeAnalyticsReplicas() error {
	m.analyticsReplica = NewAnalyticsReplicaService(m.config.ReadReplicaURLs, m.logger)
	
	m.metrics.serviceHealth.WithLabelValues("analytics_replicas").Set(1)
	m.metrics.optimizationStatus.WithLabelValues("analytics_replicas").Set(1)
	
	m.logger.Info("✅ Analytics replicas initialisés", 
		zap.Int("replicas_count", len(m.config.ReadReplicaURLs)),
	)
	return nil
}

// startMaintenanceServices démarre les services de maintenance automatique
func (m *DatabaseOptimizationManager) startMaintenanceServices() {
	ticker := time.NewTicker(m.config.OptimizationInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.performMaintenance()
		}
	}
}

// performMaintenance exécute la maintenance automatique
func (m *DatabaseOptimizationManager) performMaintenance() {
	m.logger.Info("🔧 Début maintenance automatique database")
	
	// Maintenance des index
	if m.indexOptimizer != nil {
		go func() {
			ctx, cancel := context.WithTimeout(m.ctx, 10*time.Minute)
			defer cancel()
			
			stats, err := m.indexOptimizer.GetIndexStats(ctx)
			if err != nil {
				m.logger.Error("Erreur récupération stats index", zap.Error(err))
			} else {
				m.logger.Debug("Stats index récupérées", zap.Any("stats", stats))
			}
		}()
	}
	
	// Health check des services
	m.performHealthChecks()
	
	m.logger.Info("✅ Maintenance automatique terminée")
}

// startMonitoring démarre le monitoring continu
func (m *DatabaseOptimizationManager) startMonitoring() {
	ticker := time.NewTicker(m.config.HealthCheckInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.updateMetrics()
		}
	}
}

// updateMetrics met à jour les métriques de performance
func (m *DatabaseOptimizationManager) updateMetrics() {
	// Métriques connection pool
	if m.connectionPool != nil {
		stats := m.connectionPool.GetStats()
		if primaryStats, ok := stats["primary"].(map[string]interface{}); ok {
			if openConns, ok := primaryStats["open_connections"].(int); ok {
				m.metrics.throughput.WithLabelValues("connections").Set(float64(openConns))
			}
		}
	}
	
	// Estimations utilisateurs concurrents (basé sur les connexions actives)
	m.metrics.concurrentUsers.Set(float64(m.estimateConcurrentUsers()))
}

// performHealthChecks vérifie la santé de tous les services
func (m *DatabaseOptimizationManager) performHealthChecks() {
	services := map[string]bool{
		"connection_pool":    m.connectionPool != nil,
		"index_optimizer":    m.indexOptimizer != nil,
		"query_optimizer":    m.queryOptimizer != nil,
		"pagination":         m.paginationService != nil,
		"analytics_replicas": m.analyticsReplica != nil,
	}
	
	for service, healthy := range services {
		healthValue := float64(0)
		if healthy {
			healthValue = 1
		}
		m.metrics.serviceHealth.WithLabelValues(service).Set(healthValue)
	}
}

// estimateConcurrentUsers estime le nombre d'utilisateurs concurrents
func (m *DatabaseOptimizationManager) estimateConcurrentUsers() int {
	// Estimation basée sur les connexions actives
	// En production, utiliser des métriques plus précises
	if m.connectionPool != nil {
		stats := m.connectionPool.GetStats()
		if primaryStats, ok := stats["primary"].(map[string]interface{}); ok {
			if inUse, ok := primaryStats["in_use"].(int); ok {
				// Estimation: ~5 utilisateurs par connexion active
				return inUse * 5
			}
		}
	}
	return 0
}

// GetOptimizedPrimaryDB retourne la connexion primary optimisée
func (m *DatabaseOptimizationManager) GetOptimizedPrimaryDB() *sqlx.DB {
	if m.connectionPool == nil {
		return nil
	}
	return m.connectionPool.GetPrimaryDB()
}

// GetOptimizedReadDB retourne une connexion read optimisée
func (m *DatabaseOptimizationManager) GetOptimizedReadDB() *sqlx.DB {
	if m.connectionPool == nil {
		return nil
	}
	return m.connectionPool.GetReadDB()
}

// GetPaginationService retourne le service de pagination
func (m *DatabaseOptimizationManager) GetPaginationService() *PaginationService {
	return m.paginationService
}

// GetAnalyticsService retourne le service analytics
func (m *DatabaseOptimizationManager) GetAnalyticsService() *AnalyticsReplicaService {
	return m.analyticsReplica
}

// GetPerformanceReport génère un rapport de performance complet
func (m *DatabaseOptimizationManager) GetPerformanceReport(ctx context.Context) (*PerformanceReport, error) {
	report := &PerformanceReport{
		GeneratedAt: time.Now(),
		Services:    make(map[string]ServiceStatus),
	}
	
	// Status des services
	report.Services["connection_pool"] = ServiceStatus{
		Enabled: m.connectionPool != nil,
		Healthy: m.connectionPool != nil,
	}
	report.Services["index_optimization"] = ServiceStatus{
		Enabled: m.indexOptimizer != nil,
		Healthy: m.indexOptimizer != nil,
	}
	report.Services["query_optimization"] = ServiceStatus{
		Enabled: m.queryOptimizer != nil,
		Healthy: m.queryOptimizer != nil,
	}
	report.Services["pagination"] = ServiceStatus{
		Enabled: m.paginationService != nil,
		Healthy: m.paginationService != nil,
	}
	report.Services["analytics_replicas"] = ServiceStatus{
		Enabled: m.analyticsReplica != nil,
		Healthy: m.analyticsReplica != nil,
	}
	
	// Métriques de performance
	if m.connectionPool != nil {
		report.ConnectionStats = m.connectionPool.GetStats()
	}
	
	// Estimation capacité
	report.EstimatedCapacity = EstimatedCapacity{
		ConcurrentUsers:    m.estimateConcurrentUsers(),
		MaxSupportedUsers:  100000, // Objectif 100k+ users
		CurrentUtilization: float64(m.estimateConcurrentUsers()) / 100000.0,
	}
	
	return report, nil
}

// Shutdown arrête proprement tous les services
func (m *DatabaseOptimizationManager) Shutdown() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if !m.isRunning {
		return nil
	}
	
	m.logger.Info("🔄 Arrêt DatabaseOptimizationManager")
	
	// Arrêter le contexte
	m.cancel()
	
	var errors []error
	
	// Fermer connection pool
	if m.connectionPool != nil {
		if err := m.connectionPool.Close(); err != nil {
			errors = append(errors, fmt.Errorf("erreur fermeture connection pool: %w", err))
		}
	}
	
	m.isRunning = false
	
	if len(errors) > 0 {
		return fmt.Errorf("erreurs pendant shutdown: %v", errors)
	}
	
	m.logger.Info("✅ DatabaseOptimizationManager arrêté proprement")
	return nil
}

// Types pour le rapport de performance

type PerformanceReport struct {
	GeneratedAt       time.Time                  `json:"generated_at"`
	Services          map[string]ServiceStatus   `json:"services"`
	ConnectionStats   map[string]interface{}     `json:"connection_stats"`
	EstimatedCapacity EstimatedCapacity          `json:"estimated_capacity"`
}

type ServiceStatus struct {
	Enabled   bool      `json:"enabled"`
	Healthy   bool      `json:"healthy"`
	LastCheck time.Time `json:"last_check,omitempty"`
}

type EstimatedCapacity struct {
	ConcurrentUsers    int     `json:"concurrent_users"`
	MaxSupportedUsers  int     `json:"max_supported_users"`
	CurrentUtilization float64 `json:"current_utilization"`
}
