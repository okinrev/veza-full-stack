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

// AnalyticsReplicaService service dédié aux requêtes analytics sur read replicas
type AnalyticsReplicaService struct {
	readReplicas []*sqlx.DB
	loadBalancer *ReadReplicaLoadBalancer
	logger       *zap.Logger
	metrics      *AnalyticsMetrics
	config       *AnalyticsConfig

	// Pool de connexions dédiées analytics
	analyticsPool *AnalyticsConnectionPool
}

// AnalyticsConfig configuration pour analytics
type AnalyticsConfig struct {
	MaxConcurrentQueries int           `json:"max_concurrent_queries"`
	QueryTimeout         time.Duration `json:"query_timeout"`
	EnableQueryCache     bool          `json:"enable_query_cache"`
	CacheTTL             time.Duration `json:"cache_ttl"`
	SlowQueryThreshold   time.Duration `json:"slow_query_threshold"`
	EnableCompression    bool          `json:"enable_compression"`
	BatchSize            int           `json:"batch_size"`
}

// DefaultAnalyticsConfig configuration par défaut
func DefaultAnalyticsConfig() *AnalyticsConfig {
	return &AnalyticsConfig{
		MaxConcurrentQueries: 10,
		QueryTimeout:         5 * time.Minute,
		EnableQueryCache:     true,
		CacheTTL:             15 * time.Minute,
		SlowQueryThreshold:   1 * time.Second,
		EnableCompression:    true,
		BatchSize:            1000,
	}
}

// ReadReplicaLoadBalancer équilibreur de charge pour read replicas
type ReadReplicaLoadBalancer struct {
	replicas     []*ReplicaNode
	currentIndex int
	mu           sync.RWMutex
	strategy     LoadBalancingStrategy
}

// LoadBalancingStrategy stratégie d'équilibrage
type LoadBalancingStrategy string

const (
	RoundRobin    LoadBalancingStrategy = "round_robin"
	LeastLatency  LoadBalancingStrategy = "least_latency"
	WeightedRound LoadBalancingStrategy = "weighted_round"
)

// ReplicaNode nœud de read replica
type ReplicaNode struct {
	DB       *sqlx.DB
	URL      string
	Weight   int           // Poids pour équilibrage
	Latency  time.Duration // Latence moyenne
	Healthy  bool          // État de santé
	LastPing time.Time     // Dernier ping
}

// AnalyticsConnectionPool pool spécialisé pour analytics
type AnalyticsConnectionPool struct {
	pools map[string]*sqlx.DB // Pool par type de requête
	mu    sync.RWMutex
}

// AnalyticsMetrics métriques pour analytics
type AnalyticsMetrics struct {
	queryDuration     *prometheus.HistogramVec
	replicaLatency    *prometheus.HistogramVec
	queryCache        *prometheus.CounterVec
	slowQueries       *prometheus.CounterVec
	replicaHealth     *prometheus.GaugeVec
	concurrentQueries prometheus.Gauge
	dataVolume        *prometheus.HistogramVec
}

// NewAnalyticsMetrics crée les métriques Prometheus
func NewAnalyticsMetrics() *AnalyticsMetrics {
	return &AnalyticsMetrics{
		queryDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "analytics_query_duration_seconds",
			Help:    "Durée des requêtes analytics",
			Buckets: []float64{0.1, 0.5, 1.0, 5.0, 10.0, 30.0, 60.0},
		}, []string{"query_type", "replica"}),
		replicaLatency: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name: "analytics_replica_latency_seconds",
			Help: "Latence des read replicas",
		}, []string{"replica_url"}),
		queryCache: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "analytics_query_cache_total",
			Help: "Cache hits/misses pour requêtes analytics",
		}, []string{"cache_type", "result"}),
		slowQueries: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "analytics_slow_queries_total",
			Help: "Nombre de requêtes analytics lentes",
		}, []string{"query_type", "threshold"}),
		replicaHealth: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "analytics_replica_health",
			Help: "État de santé des read replicas analytics",
		}, []string{"replica_url"}),
		concurrentQueries: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "analytics_concurrent_queries",
			Help: "Nombre de requêtes analytics concurrentes",
		}),
		dataVolume: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name: "analytics_data_volume_bytes",
			Help: "Volume de données retournées par requêtes analytics",
		}, []string{"query_type"}),
	}
}

// NewAnalyticsReplicaService crée un nouveau service analytics
func NewAnalyticsReplicaService(replicaURLs []string, logger *zap.Logger) *AnalyticsReplicaService {
	service := &AnalyticsReplicaService{
		logger:  logger,
		metrics: NewAnalyticsMetrics(),
		config:  DefaultAnalyticsConfig(),
		analyticsPool: &AnalyticsConnectionPool{
			pools: make(map[string]*sqlx.DB),
		},
	}

	// Initialiser les read replicas
	service.initializeReplicas(replicaURLs)

	// Initialiser le load balancer
	service.loadBalancer = &ReadReplicaLoadBalancer{
		strategy: LeastLatency,
	}

	return service
}

// initializeReplicas initialise les connexions aux read replicas
func (s *AnalyticsReplicaService) initializeReplicas(replicaURLs []string) {
	s.readReplicas = make([]*sqlx.DB, 0, len(replicaURLs))
	s.loadBalancer.replicas = make([]*ReplicaNode, 0, len(replicaURLs))

	for i, url := range replicaURLs {
		db, err := sqlx.Connect("postgres", url)
		if err != nil {
			s.logger.Error("Erreur connexion read replica",
				zap.Int("index", i),
				zap.String("url", url),
				zap.Error(err),
			)
			continue
		}

		// Configuration optimisée pour analytics (lecture seule)
		db.SetMaxOpenConns(20) // Moins de connexions par replica
		db.SetMaxIdleConns(10)
		db.SetConnMaxLifetime(1 * time.Hour)

		// Test de connexion
		if err := db.Ping(); err != nil {
			s.logger.Error("Ping read replica échoué",
				zap.Int("index", i),
				zap.Error(err),
			)
			continue
		}

		s.readReplicas = append(s.readReplicas, db)

		// Ajouter au load balancer
		node := &ReplicaNode{
			DB:       db,
			URL:      url,
			Weight:   1,
			Healthy:  true,
			LastPing: time.Now(),
		}
		s.loadBalancer.replicas = append(s.loadBalancer.replicas, node)

		s.logger.Info("Read replica initialisé pour analytics",
			zap.Int("index", i),
			zap.String("url", url),
		)
	}
}

// UserAnalytics analytics des utilisateurs
func (s *AnalyticsReplicaService) GetUserAnalytics(ctx context.Context, timeRange string) (*UserAnalyticsData, error) {
	start := time.Now()
	replica := s.loadBalancer.SelectReplica()

	defer func() {
		duration := time.Since(start)
		s.metrics.queryDuration.WithLabelValues("user_analytics", replica.URL).Observe(duration.Seconds())

		if duration > s.config.SlowQueryThreshold {
			s.metrics.slowQueries.WithLabelValues("user_analytics", "1s").Inc()
		}
	}()

	// Requête optimisée pour analytics utilisateurs
	query := `
		WITH date_range AS (
			SELECT 
				CASE 
					WHEN $1 = '7d' THEN CURRENT_DATE - INTERVAL '7 days'
					WHEN $1 = '30d' THEN CURRENT_DATE - INTERVAL '30 days'
					WHEN $1 = '90d' THEN CURRENT_DATE - INTERVAL '90 days'
					ELSE CURRENT_DATE - INTERVAL '30 days'
				END as start_date
		),
		user_stats AS (
			SELECT 
				DATE_TRUNC('day', u.created_at) as registration_date,
				COUNT(*) as new_registrations,
				COUNT(*) FILTER (WHERE u.is_verified = true) as verified_users,
				COUNT(*) FILTER (WHERE u.status = 'active') as active_users
			FROM users u, date_range dr
			WHERE u.created_at >= dr.start_date
			GROUP BY DATE_TRUNC('day', u.created_at)
		),
		activity_stats AS (
			SELECT 
				DATE_TRUNC('day', s.created_at) as activity_date,
				COUNT(DISTINCT s.user_id) as daily_active_users,
				AVG(s.duration_minutes) as avg_session_duration
			FROM user_sessions s, date_range dr
			WHERE s.created_at >= dr.start_date
			GROUP BY DATE_TRUNC('day', s.created_at)
		)
		SELECT 
			COALESCE(us.registration_date, ast.activity_date) as date,
			COALESCE(us.new_registrations, 0) as new_registrations,
			COALESCE(us.verified_users, 0) as verified_users,
			COALESCE(us.active_users, 0) as active_users,
			COALESCE(ast.daily_active_users, 0) as daily_active_users,
			COALESCE(ast.avg_session_duration, 0) as avg_session_duration
		FROM user_stats us
		FULL OUTER JOIN activity_stats ast 
			ON us.registration_date = ast.activity_date
		ORDER BY date DESC
	`

	var results []UserAnalyticsRow
	if err := replica.DB.SelectContext(ctx, &results, query, timeRange); err != nil {
		return nil, fmt.Errorf("erreur requête user analytics: %w", err)
	}

	// Calculer métriques agrégées
	analytics := &UserAnalyticsData{
		TimeRange: timeRange,
		Daily:     results,
	}

	for _, row := range results {
		analytics.TotalRegistrations += row.NewRegistrations
		analytics.TotalActiveUsers += row.DailyActiveUsers
	}

	if len(results) > 0 {
		analytics.AvgDailyActiveUsers = analytics.TotalActiveUsers / int64(len(results))
	}

	return analytics, nil
}

// ChatAnalytics analytics des conversations
func (s *AnalyticsReplicaService) GetChatAnalytics(ctx context.Context, timeRange string) (*ChatAnalyticsData, error) {
	start := time.Now()
	replica := s.loadBalancer.SelectReplica()

	defer func() {
		duration := time.Since(start)
		s.metrics.queryDuration.WithLabelValues("chat_analytics", replica.URL).Observe(duration.Seconds())
	}()

	// Requête optimisée pour analytics chat
	query := `
		WITH date_range AS (
			SELECT 
				CASE 
					WHEN $1 = '7d' THEN CURRENT_DATE - INTERVAL '7 days'
					WHEN $1 = '30d' THEN CURRENT_DATE - INTERVAL '30 days'
					WHEN $1 = '90d' THEN CURRENT_DATE - INTERVAL '90 days'
					ELSE CURRENT_DATE - INTERVAL '30 days'
				END as start_date
		),
		message_stats AS (
			SELECT 
				DATE_TRUNC('day', m.created_at) as message_date,
				COUNT(*) as total_messages,
				COUNT(DISTINCT m.user_id) as active_users,
				COUNT(DISTINCT m.room_id) as active_rooms,
				AVG(LENGTH(m.content)) as avg_message_length
			FROM chat_messages m, date_range dr
			WHERE m.created_at >= dr.start_date
			GROUP BY DATE_TRUNC('day', m.created_at)
		),
		room_stats AS (
			SELECT 
				DATE_TRUNC('day', r.created_at) as creation_date,
				COUNT(*) as new_rooms,
				COUNT(*) FILTER (WHERE r.room_type = 'public') as public_rooms,
				COUNT(*) FILTER (WHERE r.room_type = 'private') as private_rooms
			FROM chat_rooms r, date_range dr
			WHERE r.created_at >= dr.start_date
			GROUP BY DATE_TRUNC('day', r.created_at)
		)
		SELECT 
			COALESCE(ms.message_date, rs.creation_date) as date,
			COALESCE(ms.total_messages, 0) as total_messages,
			COALESCE(ms.active_users, 0) as active_users,
			COALESCE(ms.active_rooms, 0) as active_rooms,
			COALESCE(ms.avg_message_length, 0) as avg_message_length,
			COALESCE(rs.new_rooms, 0) as new_rooms,
			COALESCE(rs.public_rooms, 0) as public_rooms,
			COALESCE(rs.private_rooms, 0) as private_rooms
		FROM message_stats ms
		FULL OUTER JOIN room_stats rs 
			ON ms.message_date = rs.creation_date
		ORDER BY date DESC
	`

	var results []ChatAnalyticsRow
	if err := replica.DB.SelectContext(ctx, &results, query, timeRange); err != nil {
		return nil, fmt.Errorf("erreur requête chat analytics: %w", err)
	}

	analytics := &ChatAnalyticsData{
		TimeRange: timeRange,
		Daily:     results,
	}

	// Calculer agrégats
	for _, row := range results {
		analytics.TotalMessages += row.TotalMessages
		analytics.TotalRooms += row.NewRooms
	}

	return analytics, nil
}

// PerformanceAnalytics analytics de performance
func (s *AnalyticsReplicaService) GetPerformanceAnalytics(ctx context.Context) (*PerformanceAnalyticsData, error) {
	start := time.Now()
	replica := s.loadBalancer.SelectReplica()

	defer func() {
		duration := time.Since(start)
		s.metrics.queryDuration.WithLabelValues("performance_analytics", replica.URL).Observe(duration.Seconds())
	}()

	// Requête pour métriques de performance
	query := `
		SELECT 
			schemaname,
			tablename,
			pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size,
			seq_scan,
			seq_tup_read,
			idx_scan,
			idx_tup_fetch,
			n_tup_ins,
			n_tup_upd,
			n_tup_del,
			n_live_tup,
			n_dead_tup,
			last_vacuum,
			last_autovacuum,
			last_analyze,
			last_autoanalyze
		FROM pg_stat_user_tables 
		ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC
		LIMIT 20
	`

	rows, err := replica.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("erreur requête performance analytics: %w", err)
	}
	defer rows.Close()

	var tables []TablePerformanceData
	for rows.Next() {
		var table TablePerformanceData
		var lastVacuum, lastAutovacuum, lastAnalyze, lastAutoanalyze *time.Time

		if err := rows.Scan(
			&table.Schema, &table.TableName, &table.Size,
			&table.SeqScan, &table.SeqTupRead, &table.IdxScan, &table.IdxTupFetch,
			&table.TupIns, &table.TupUpd, &table.TupDel, &table.LiveTup, &table.DeadTup,
			&lastVacuum, &lastAutovacuum, &lastAnalyze, &lastAutoanalyze,
		); err != nil {
			continue
		}

		if lastVacuum != nil {
			table.LastVacuum = *lastVacuum
		}
		if lastAnalyze != nil {
			table.LastAnalyze = *lastAnalyze
		}

		tables = append(tables, table)
	}

	analytics := &PerformanceAnalyticsData{
		Tables:      tables,
		GeneratedAt: time.Now(),
	}

	return analytics, nil
}

// SelectReplica sélectionne le meilleur read replica
func (lb *ReadReplicaLoadBalancer) SelectReplica() *ReplicaNode {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	if len(lb.replicas) == 0 {
		return nil
	}

	// Filtrer les replicas sains
	var healthyReplicas []*ReplicaNode
	for _, replica := range lb.replicas {
		if replica.Healthy {
			healthyReplicas = append(healthyReplicas, replica)
		}
	}

	if len(healthyReplicas) == 0 {
		// Fallback sur le premier replica même s'il est unhealthy
		return lb.replicas[0]
	}

	switch lb.strategy {
	case LeastLatency:
		return lb.selectLeastLatency(healthyReplicas)
	case WeightedRound:
		return lb.selectWeightedRound(healthyReplicas)
	default: // RoundRobin
		return lb.selectRoundRobin(healthyReplicas)
	}
}

// selectLeastLatency sélectionne le replica avec la plus faible latence
func (lb *ReadReplicaLoadBalancer) selectLeastLatency(replicas []*ReplicaNode) *ReplicaNode {
	bestReplica := replicas[0]
	for _, replica := range replicas[1:] {
		if replica.Latency < bestReplica.Latency {
			bestReplica = replica
		}
	}
	return bestReplica
}

// selectRoundRobin sélection round-robin
func (lb *ReadReplicaLoadBalancer) selectRoundRobin(replicas []*ReplicaNode) *ReplicaNode {
	replica := replicas[lb.currentIndex%len(replicas)]
	lb.currentIndex++
	return replica
}

// selectWeightedRound sélection avec poids
func (lb *ReadReplicaLoadBalancer) selectWeightedRound(replicas []*ReplicaNode) *ReplicaNode {
	// Implémentation simplifiée - utiliser weighted random en production
	return replicas[0]
}

// Types de données pour analytics

type UserAnalyticsData struct {
	TimeRange           string             `json:"time_range"`
	TotalRegistrations  int64              `json:"total_registrations"`
	TotalActiveUsers    int64              `json:"total_active_users"`
	AvgDailyActiveUsers int64              `json:"avg_daily_active_users"`
	Daily               []UserAnalyticsRow `json:"daily"`
}

type UserAnalyticsRow struct {
	Date               time.Time `db:"date"`
	NewRegistrations   int64     `db:"new_registrations"`
	VerifiedUsers      int64     `db:"verified_users"`
	ActiveUsers        int64     `db:"active_users"`
	DailyActiveUsers   int64     `db:"daily_active_users"`
	AvgSessionDuration float64   `db:"avg_session_duration"`
}

type ChatAnalyticsData struct {
	TimeRange     string             `json:"time_range"`
	TotalMessages int64              `json:"total_messages"`
	TotalRooms    int64              `json:"total_rooms"`
	Daily         []ChatAnalyticsRow `json:"daily"`
}

type ChatAnalyticsRow struct {
	Date             time.Time `db:"date"`
	TotalMessages    int64     `db:"total_messages"`
	ActiveUsers      int64     `db:"active_users"`
	ActiveRooms      int64     `db:"active_rooms"`
	AvgMessageLength float64   `db:"avg_message_length"`
	NewRooms         int64     `db:"new_rooms"`
	PublicRooms      int64     `db:"public_rooms"`
	PrivateRooms     int64     `db:"private_rooms"`
}

type PerformanceAnalyticsData struct {
	Tables      []TablePerformanceData `json:"tables"`
	GeneratedAt time.Time              `json:"generated_at"`
}

type TablePerformanceData struct {
	Schema      string `db:"schemaname"`
	TableName   string `db:"tablename"`
	Size        string `db:"size"`
	SeqScan     int64  `db:"seq_scan"`
	SeqTupRead  int64  `db:"seq_tup_read"`
	IdxScan     int64  `db:"idx_scan"`
	IdxTupFetch int64  `db:"idx_tup_fetch"`
	TupIns      int64  `db:"n_tup_ins"`
	TupUpd      int64  `db:"n_tup_upd"`
	TupDel      int64  `db:"n_tup_del"`
	LiveTup     int64  `db:"n_live_tup"`
	DeadTup     int64  `db:"n_dead_tup"`
	LastVacuum  time.Time
	LastAnalyze time.Time
}
