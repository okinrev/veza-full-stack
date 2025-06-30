package monitoring

import (
	"runtime"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// PrometheusMetrics contient toutes les métriques Prometheus
type PrometheusMetrics struct {
	// Métriques HTTP
	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
	HTTPRequestsActive  prometheus.Gauge
	HTTPResponseSize    *prometheus.HistogramVec

	// Métriques Auth
	AuthOperationsTotal *prometheus.CounterVec
	AuthTokensGenerated *prometheus.CounterVec
	AuthFailuresTotal   *prometheus.CounterVec
	AuthSessionsActive  prometheus.Gauge

	// Métriques Database
	DBConnectionsActive prometheus.Gauge
	DBConnectionsTotal  *prometheus.CounterVec
	DBQueryDuration     *prometheus.HistogramVec
	DBQueryTotal        *prometheus.CounterVec

	// Métriques Cache Redis
	CacheOperationsTotal   *prometheus.CounterVec
	CacheHitRatio          prometheus.Gauge
	CacheConnectionsActive prometheus.Gauge
	CacheLatency           *prometheus.HistogramVec

	// Métriques Business
	UsersActive   prometheus.Gauge
	UsersTotal    prometheus.Gauge
	MessagesTotal *prometheus.CounterVec
	RoomsActive   prometheus.Gauge

	// Métriques gRPC
	GRPCRequestsTotal     *prometheus.CounterVec
	GRPCDuration          *prometheus.HistogramVec
	GRPCConnectionsActive prometheus.Gauge

	// Métriques System
	SystemMemoryUsage prometheus.Gauge
	SystemCPUUsage    prometheus.Gauge
	GoroutinesActive  prometheus.Gauge

	// Registry
	registry *prometheus.Registry
	logger   *zap.Logger
}

// NewPrometheusMetrics crée une nouvelle instance des métriques
func NewPrometheusMetrics(logger *zap.Logger) *PrometheusMetrics {
	registry := prometheus.NewRegistry()

	metrics := &PrometheusMetrics{
		registry: registry,
		logger:   logger,

		// Métriques HTTP
		HTTPRequestsTotal: promauto.With(registry).NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "veza_backend",
				Subsystem: "http",
				Name:      "requests_total",
				Help:      "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status_code"},
		),

		HTTPRequestDuration: promauto.With(registry).NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "veza_backend",
				Subsystem: "http",
				Name:      "request_duration_seconds",
				Help:      "HTTP request duration in seconds",
				Buckets:   prometheus.ExponentialBuckets(0.001, 2, 15),
			},
			[]string{"method", "endpoint", "status_code"},
		),

		HTTPRequestsActive: promauto.With(registry).NewGauge(
			prometheus.GaugeOpts{
				Namespace: "veza_backend",
				Subsystem: "http",
				Name:      "requests_active",
				Help:      "Current number of active HTTP requests",
			},
		),

		HTTPResponseSize: promauto.With(registry).NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "veza_backend",
				Subsystem: "http",
				Name:      "response_size_bytes",
				Help:      "HTTP response size in bytes",
				Buckets:   prometheus.ExponentialBuckets(100, 10, 6),
			},
			[]string{"method", "endpoint"},
		),

		// Métriques Auth
		AuthOperationsTotal: promauto.With(registry).NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "veza_backend",
				Subsystem: "auth",
				Name:      "operations_total",
				Help:      "Total number of auth operations",
			},
			[]string{"operation", "status"},
		),

		AuthTokensGenerated: promauto.With(registry).NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "veza_backend",
				Subsystem: "auth",
				Name:      "tokens_generated_total",
				Help:      "Total number of JWT tokens generated",
			},
			[]string{"type"},
		),

		AuthFailuresTotal: promauto.With(registry).NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "veza_backend",
				Subsystem: "auth",
				Name:      "failures_total",
				Help:      "Total number of auth failures",
			},
			[]string{"reason"},
		),

		AuthSessionsActive: promauto.With(registry).NewGauge(
			prometheus.GaugeOpts{
				Namespace: "veza_backend",
				Subsystem: "auth",
				Name:      "sessions_active",
				Help:      "Current number of active user sessions",
			},
		),

		// Métriques Database
		DBConnectionsActive: promauto.With(registry).NewGauge(
			prometheus.GaugeOpts{
				Namespace: "veza_backend",
				Subsystem: "database",
				Name:      "connections_active",
				Help:      "Current number of active database connections",
			},
		),

		DBConnectionsTotal: promauto.With(registry).NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "veza_backend",
				Subsystem: "database",
				Name:      "connections_total",
				Help:      "Total number of database connections created",
			},
			[]string{"status"},
		),

		DBQueryDuration: promauto.With(registry).NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "veza_backend",
				Subsystem: "database",
				Name:      "query_duration_seconds",
				Help:      "Database query duration in seconds",
				Buckets:   prometheus.ExponentialBuckets(0.001, 2, 12),
			},
			[]string{"operation", "table"},
		),

		DBQueryTotal: promauto.With(registry).NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "veza_backend",
				Subsystem: "database",
				Name:      "queries_total",
				Help:      "Total number of database queries",
			},
			[]string{"operation", "table", "status"},
		),

		// Métriques Cache
		CacheOperationsTotal: promauto.With(registry).NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "veza_backend",
				Subsystem: "cache",
				Name:      "operations_total",
				Help:      "Total number of cache operations",
			},
			[]string{"operation", "status"},
		),

		CacheHitRatio: promauto.With(registry).NewGauge(
			prometheus.GaugeOpts{
				Namespace: "veza_backend",
				Subsystem: "cache",
				Name:      "hit_ratio",
				Help:      "Cache hit ratio (0-1)",
			},
		),

		CacheConnectionsActive: promauto.With(registry).NewGauge(
			prometheus.GaugeOpts{
				Namespace: "veza_backend",
				Subsystem: "cache",
				Name:      "connections_active",
				Help:      "Current number of active cache connections",
			},
		),

		CacheLatency: promauto.With(registry).NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "veza_backend",
				Subsystem: "cache",
				Name:      "latency_seconds",
				Help:      "Cache operation latency in seconds",
				Buckets:   prometheus.ExponentialBuckets(0.0001, 2, 10),
			},
			[]string{"operation"},
		),

		// Métriques Business
		UsersActive: promauto.With(registry).NewGauge(
			prometheus.GaugeOpts{
				Namespace: "veza_backend",
				Subsystem: "business",
				Name:      "users_active",
				Help:      "Current number of active users (last 5 minutes)",
			},
		),

		UsersTotal: promauto.With(registry).NewGauge(
			prometheus.GaugeOpts{
				Namespace: "veza_backend",
				Subsystem: "business",
				Name:      "users_total",
				Help:      "Total number of registered users",
			},
		),

		MessagesTotal: promauto.With(registry).NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "veza_backend",
				Subsystem: "business",
				Name:      "messages_total",
				Help:      "Total number of messages sent",
			},
			[]string{"type"},
		),

		RoomsActive: promauto.With(registry).NewGauge(
			prometheus.GaugeOpts{
				Namespace: "veza_backend",
				Subsystem: "business",
				Name:      "rooms_active",
				Help:      "Current number of active chat rooms",
			},
		),

		// Métriques gRPC
		GRPCRequestsTotal: promauto.With(registry).NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "veza_backend",
				Subsystem: "grpc",
				Name:      "requests_total",
				Help:      "Total number of gRPC requests",
			},
			[]string{"service", "method", "code"},
		),

		GRPCDuration: promauto.With(registry).NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "veza_backend",
				Subsystem: "grpc",
				Name:      "request_duration_seconds",
				Help:      "gRPC request duration in seconds",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"service", "method"},
		),

		GRPCConnectionsActive: promauto.With(registry).NewGauge(
			prometheus.GaugeOpts{
				Namespace: "veza_backend",
				Subsystem: "grpc",
				Name:      "connections_active",
				Help:      "Current number of active gRPC connections",
			},
		),

		// Métriques System
		SystemMemoryUsage: promauto.With(registry).NewGauge(
			prometheus.GaugeOpts{
				Namespace: "veza_backend",
				Subsystem: "system",
				Name:      "memory_usage_bytes",
				Help:      "Current memory usage in bytes",
			},
		),

		SystemCPUUsage: promauto.With(registry).NewGauge(
			prometheus.GaugeOpts{
				Namespace: "veza_backend",
				Subsystem: "system",
				Name:      "cpu_usage_percent",
				Help:      "Current CPU usage percentage",
			},
		),

		GoroutinesActive: promauto.With(registry).NewGauge(
			prometheus.GaugeOpts{
				Namespace: "veza_backend",
				Subsystem: "system",
				Name:      "goroutines_active",
				Help:      "Current number of active goroutines",
			},
		),
	}

	// Démarrer la collecte des métriques système
	go metrics.collectSystemMetrics()

	if logger != nil {
		logger.Info("✅ Métriques Prometheus initialisées")
	}

	return metrics
}

// PrometheusMiddleware middleware Gin pour les métriques HTTP
func (m *PrometheusMetrics) PrometheusMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Incrémenter les requêtes actives
		m.HTTPRequestsActive.Inc()
		defer m.HTTPRequestsActive.Dec()

		start := time.Now()

		// Traiter la requête
		c.Next()

		// Calculer la durée
		duration := time.Since(start).Seconds()

		// Extraire les labels
		method := c.Request.Method
		endpoint := c.FullPath()
		if endpoint == "" {
			endpoint = "unknown"
		}
		statusCode := strconv.Itoa(c.Writer.Status())

		// Incrémenter les métriques
		m.HTTPRequestsTotal.WithLabelValues(method, endpoint, statusCode).Inc()
		m.HTTPRequestDuration.WithLabelValues(method, endpoint, statusCode).Observe(duration)

		// Taille de la réponse
		responseSize := float64(c.Writer.Size())
		if responseSize > 0 {
			m.HTTPResponseSize.WithLabelValues(method, endpoint).Observe(responseSize)
		}
	})
}

// RecordAuthOperation enregistre une opération d'authentification
func (m *PrometheusMetrics) RecordAuthOperation(operation, status string) {
	m.AuthOperationsTotal.WithLabelValues(operation, status).Inc()
}

// RecordTokenGeneration enregistre la génération d'un token
func (m *PrometheusMetrics) RecordTokenGeneration(tokenType string) {
	m.AuthTokensGenerated.WithLabelValues(tokenType).Inc()
}

// RecordAuthFailure enregistre un échec d'authentification
func (m *PrometheusMetrics) RecordAuthFailure(reason string) {
	m.AuthFailuresTotal.WithLabelValues(reason).Inc()
}

// RecordDBOperation enregistre une opération de base de données
func (m *PrometheusMetrics) RecordDBOperation(operation, table, status string, duration time.Duration) {
	m.DBQueryTotal.WithLabelValues(operation, table, status).Inc()
	m.DBQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
}

// RecordCacheOperation enregistre une opération de cache
func (m *PrometheusMetrics) RecordCacheOperation(operation, status string, duration time.Duration) {
	m.CacheOperationsTotal.WithLabelValues(operation, status).Inc()
	m.CacheLatency.WithLabelValues(operation).Observe(duration.Seconds())
}

// RecordGRPCRequest enregistre une requête gRPC
func (m *PrometheusMetrics) RecordGRPCRequest(service, method, code string, duration time.Duration) {
	m.GRPCRequestsTotal.WithLabelValues(service, method, code).Inc()
	m.GRPCDuration.WithLabelValues(service, method).Observe(duration.Seconds())
}

// UpdateBusinessMetrics met à jour les métriques business
func (m *PrometheusMetrics) UpdateBusinessMetrics(activeUsers, totalUsers, activeRooms int) {
	m.UsersActive.Set(float64(activeUsers))
	m.UsersTotal.Set(float64(totalUsers))
	m.RoomsActive.Set(float64(activeRooms))
}

// GetHandler retourne le handler Prometheus pour l'endpoint /metrics
func (m *PrometheusMetrics) GetHandler() gin.HandlerFunc {
	handler := promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})
	return gin.WrapH(handler)
}

// collectSystemMetrics collecte les métriques système périodiquement
func (m *PrometheusMetrics) collectSystemMetrics() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)

		m.SystemMemoryUsage.Set(float64(memStats.Alloc))
		m.GoroutinesActive.Set(float64(runtime.NumGoroutine()))
	}
}
