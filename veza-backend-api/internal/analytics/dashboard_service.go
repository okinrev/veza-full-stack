package analytics

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// DashboardService service de dashboard temps réel
type DashboardService struct {
	db                     *sql.DB
	logger                 *zap.Logger
	cache                  EngagementCache
	userEngagementService  *UserEngagementService
	chatAnalyticsService   *ChatAnalyticsService
	streamAnalyticsService *StreamAnalyticsService
	revenueAnalyticsService *RevenueAnalyticsService
	
	// WebSocket pour updates temps réel
	subscribers map[string]chan *DashboardUpdate
	mutex       sync.RWMutex
	
	// Métriques en cache
	cachedMetrics *DashboardMetrics
	lastUpdate    time.Time
}

// DashboardMetrics métriques complètes du dashboard
type DashboardMetrics struct {
	Overview          OverviewMetrics           `json:"overview"`
	UserEngagement    *UserEngagementMetrics    `json:"user_engagement"`
	ChatMetrics       *ChatMetrics              `json:"chat_metrics"`
	StreamMetrics     *StreamMetrics            `json:"stream_metrics"`
	RevenueMetrics    *RevenueMetrics           `json:"revenue_metrics"`
	RealtimeMetrics   RealtimeMetrics           `json:"realtime_metrics"`
	SystemHealth      SystemHealthMetrics       `json:"system_health"`
	AlertsAndNotifications []DashboardAlert     `json:"alerts_notifications"`
	LastUpdated       time.Time                 `json:"last_updated"`
}

// OverviewMetrics métriques de vue d'ensemble
type OverviewMetrics struct {
	TotalUsers        int64   `json:"total_users"`
	ActiveUsers24h    int64   `json:"active_users_24h"`
	TotalRevenue      float64 `json:"total_revenue"`
	Revenue24h        float64 `json:"revenue_24h"`
	TotalStreams      int64   `json:"total_streams"`
	Streams24h        int64   `json:"streams_24h"`
	TotalMessages     int64   `json:"total_messages"`
	Messages24h       int64   `json:"messages_24h"`
	ConversionRate    float64 `json:"conversion_rate"`
	ChurnRate         float64 `json:"churn_rate"`
	NPS               float64 `json:"nps"`                // Net Promoter Score
	ServerUptime      float64 `json:"server_uptime"`
}

// RealtimeMetrics métriques temps réel
type RealtimeMetrics struct {
	CurrentUsers      int64                 `json:"current_users"`
	ActiveStreams     int64                 `json:"active_streams"`
	MessageRate       float64               `json:"message_rate_per_minute"`
	StreamRate        float64               `json:"stream_rate_per_minute"`
	ErrorRate         float64               `json:"error_rate_percent"`
	ResponseTime      float64               `json:"avg_response_time_ms"`
	BandwidthUsage    float64               `json:"bandwidth_usage_mbps"`
	CPUUsage          float64               `json:"cpu_usage_percent"`
	MemoryUsage       float64               `json:"memory_usage_percent"`
	DatabaseQueries   int64                 `json:"database_queries_per_second"`
	CacheHitRate      float64               `json:"cache_hit_rate_percent"`
	ActiveConnections int64                 `json:"active_connections"`
	QueueLength       int64                 `json:"queue_length"`
	GeoDistribution   map[string]int64      `json:"geo_distribution"`
}

// SystemHealthMetrics métriques de santé du système
type SystemHealthMetrics struct {
	OverallHealth     string                 `json:"overall_health"`      // healthy, warning, critical
	Services          map[string]ServiceHealth `json:"services"`
	Infrastructure    InfrastructureHealth   `json:"infrastructure"`
	Alerts            []SystemAlert          `json:"alerts"`
	LastIncident      *time.Time             `json:"last_incident,omitempty"`
	UptimePercentage  float64                `json:"uptime_percentage"`
}

// ServiceHealth santé d'un service
type ServiceHealth struct {
	Status        string    `json:"status"`         // healthy, warning, critical, down
	ResponseTime  float64   `json:"response_time_ms"`
	ErrorRate     float64   `json:"error_rate_percent"`
	LastCheck     time.Time `json:"last_check"`
	Dependencies  []string  `json:"dependencies"`
}

// InfrastructureHealth santé de l'infrastructure
type InfrastructureHealth struct {
	Database      ServiceHealth `json:"database"`
	Redis         ServiceHealth `json:"redis"`
	FileStorage   ServiceHealth `json:"file_storage"`
	CDN           ServiceHealth `json:"cdn"`
	LoadBalancer  ServiceHealth `json:"load_balancer"`
	Monitoring    ServiceHealth `json:"monitoring"`
}

// SystemAlert alerte système
type SystemAlert struct {
	ID          string    `json:"id"`
	Severity    string    `json:"severity"`    // info, warning, critical
	Service     string    `json:"service"`
	Message     string    `json:"message"`
	Timestamp   time.Time `json:"timestamp"`
	Resolved    bool      `json:"resolved"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
}

// DashboardAlert alerte dashboard
type DashboardAlert struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`        // metric_threshold, anomaly, system
	Severity    string    `json:"severity"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Metric      string    `json:"metric"`
	Threshold   float64   `json:"threshold"`
	CurrentValue float64  `json:"current_value"`
	Timestamp   time.Time `json:"timestamp"`
	Acknowledged bool     `json:"acknowledged"`
}

// DashboardUpdate update temps réel du dashboard
type DashboardUpdate struct {
	Type      string      `json:"type"`      // full_refresh, metric_update, alert
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// DashboardConfig configuration du dashboard
type DashboardConfig struct {
	RefreshInterval time.Duration `json:"refresh_interval"`
	AlertThresholds AlertThresholds `json:"alert_thresholds"`
	EnableRealtime  bool          `json:"enable_realtime"`
}

// AlertThresholds seuils d'alerte
type AlertThresholds struct {
	HighErrorRate        float64 `json:"high_error_rate"`         // %
	HighResponseTime     float64 `json:"high_response_time"`      // ms
	LowConversionRate    float64 `json:"low_conversion_rate"`     // %
	HighChurnRate        float64 `json:"high_churn_rate"`         // %
	LowCacheHitRate      float64 `json:"low_cache_hit_rate"`      // %
	HighCPUUsage         float64 `json:"high_cpu_usage"`          // %
	HighMemoryUsage      float64 `json:"high_memory_usage"`       // %
	LowDiskSpace         float64 `json:"low_disk_space"`          // %
}

// NewDashboardService crée un nouveau service de dashboard
func NewDashboardService(
	db *sql.DB,
	logger *zap.Logger,
	cache EngagementCache,
	userEngagementService *UserEngagementService,
	chatAnalyticsService *ChatAnalyticsService,
	streamAnalyticsService *StreamAnalyticsService,
	revenueAnalyticsService *RevenueAnalyticsService,
) *DashboardService {
	return &DashboardService{
		db:                      db,
		logger:                  logger,
		cache:                   cache,
		userEngagementService:   userEngagementService,
		chatAnalyticsService:    chatAnalyticsService,
		streamAnalyticsService:  streamAnalyticsService,
		revenueAnalyticsService: revenueAnalyticsService,
		subscribers:             make(map[string]chan *DashboardUpdate),
	}
}

// Start démarre le service de dashboard
func (s *DashboardService) Start(ctx context.Context, config *DashboardConfig) {
	if config == nil {
		config = &DashboardConfig{
			RefreshInterval: 30 * time.Second,
			EnableRealtime:  true,
			AlertThresholds: AlertThresholds{
				HighErrorRate:     5.0,
				HighResponseTime:  1000.0,
				LowConversionRate: 2.0,
				HighChurnRate:     10.0,
				LowCacheHitRate:   80.0,
				HighCPUUsage:      80.0,
				HighMemoryUsage:   85.0,
				LowDiskSpace:      90.0,
			},
		}
	}

	s.logger.Info("🚀 Starting dashboard service",
		zap.Duration("refresh_interval", config.RefreshInterval),
		zap.Bool("realtime_enabled", config.EnableRealtime))

	// Worker de mise à jour des métriques
	go s.metricsUpdateWorker(ctx, config)

	// Worker de détection d'alertes
	go s.alertWorker(ctx, config)

	// Worker de nettoyage des abonnés
	go s.cleanupWorker(ctx)
}

// GetDashboardMetrics retourne les métriques complètes du dashboard
func (s *DashboardService) GetDashboardMetrics(ctx context.Context, dateRange DateRange) (*DashboardMetrics, error) {
	// Vérifier le cache
	if s.cachedMetrics != nil && time.Since(s.lastUpdate) < 30*time.Second {
		return s.cachedMetrics, nil
	}

	metrics := &DashboardMetrics{}

	// Collecter toutes les métriques en parallèle
	var wg sync.WaitGroup
	var mu sync.Mutex
	errors := make([]error, 0)

	// Overview
	wg.Add(1)
	go func() {
		defer wg.Done()
		overview, err := s.getOverviewMetrics(ctx, dateRange)
		if err != nil {
			mu.Lock()
			errors = append(errors, fmt.Errorf("overview metrics: %w", err))
			mu.Unlock()
			return
		}
		metrics.Overview = overview
	}()

	// User Engagement
	wg.Add(1)
	go func() {
		defer wg.Done()
		userMetrics, err := s.userEngagementService.GetEngagementMetrics(ctx, dateRange)
		if err != nil {
			mu.Lock()
			errors = append(errors, fmt.Errorf("user engagement: %w", err))
			mu.Unlock()
			return
		}
		metrics.UserEngagement = userMetrics
	}()

	// Chat Metrics
	wg.Add(1)
	go func() {
		defer wg.Done()
		chatMetrics, err := s.chatAnalyticsService.GetChatMetrics(ctx, dateRange)
		if err != nil {
			mu.Lock()
			errors = append(errors, fmt.Errorf("chat metrics: %w", err))
			mu.Unlock()
			return
		}
		metrics.ChatMetrics = chatMetrics
	}()

	// Stream Metrics
	wg.Add(1)
	go func() {
		defer wg.Done()
		streamMetrics, err := s.streamAnalyticsService.GetStreamMetrics(ctx, dateRange)
		if err != nil {
			mu.Lock()
			errors = append(errors, fmt.Errorf("stream metrics: %w", err))
			mu.Unlock()
			return
		}
		metrics.StreamMetrics = streamMetrics
	}()

	// Revenue Metrics
	wg.Add(1)
	go func() {
		defer wg.Done()
		revenueMetrics, err := s.revenueAnalyticsService.GetRevenueMetrics(ctx, dateRange)
		if err != nil {
			mu.Lock()
			errors = append(errors, fmt.Errorf("revenue metrics: %w", err))
			mu.Unlock()
			return
		}
		metrics.RevenueMetrics = revenueMetrics
	}()

	// Realtime Metrics
	wg.Add(1)
	go func() {
		defer wg.Done()
		realtimeMetrics, err := s.getRealtimeMetrics(ctx)
		if err != nil {
			mu.Lock()
			errors = append(errors, fmt.Errorf("realtime metrics: %w", err))
			mu.Unlock()
			return
		}
		metrics.RealtimeMetrics = realtimeMetrics
	}()

	// System Health
	wg.Add(1)
	go func() {
		defer wg.Done()
		systemHealth, err := s.getSystemHealth(ctx)
		if err != nil {
			mu.Lock()
			errors = append(errors, fmt.Errorf("system health: %w", err))
			mu.Unlock()
			return
		}
		metrics.SystemHealth = systemHealth
	}()

	wg.Wait()

	// Vérifier les erreurs
	if len(errors) > 0 {
		s.logger.Error("Errors collecting dashboard metrics", zap.Any("errors", errors))
		// Continuer avec les métriques partielles
	}

	// Alertes
	alerts, err := s.getActiveAlerts(ctx)
	if err != nil {
		s.logger.Error("Failed to get alerts", zap.Error(err))
	} else {
		metrics.AlertsAndNotifications = alerts
	}

	metrics.LastUpdated = time.Now()

	// Mettre en cache
	s.cachedMetrics = metrics
	s.lastUpdate = time.Now()

	return metrics, nil
}

// SubscribeToUpdates s'abonne aux mises à jour temps réel
func (s *DashboardService) SubscribeToUpdates(subscriberID string) chan *DashboardUpdate {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	channel := make(chan *DashboardUpdate, 100)
	s.subscribers[subscriberID] = channel

	s.logger.Debug("New dashboard subscriber", zap.String("subscriber_id", subscriberID))

	return channel
}

// UnsubscribeFromUpdates se désabonne des mises à jour
func (s *DashboardService) UnsubscribeFromUpdates(subscriberID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if channel, exists := s.subscribers[subscriberID]; exists {
		close(channel)
		delete(s.subscribers, subscriberID)
		s.logger.Debug("Dashboard subscriber removed", zap.String("subscriber_id", subscriberID))
	}
}

// BroadcastUpdate diffuse une mise à jour à tous les abonnés
func (s *DashboardService) BroadcastUpdate(update *DashboardUpdate) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for subscriberID, channel := range s.subscribers {
		select {
		case channel <- update:
		default:
			s.logger.Warn("Dashboard update channel full", zap.String("subscriber_id", subscriberID))
		}
	}
}

// ============================================================================
// WORKERS ET MÉTHODES PRIVÉES
// ============================================================================

// metricsUpdateWorker met à jour les métriques périodiquement
func (s *DashboardService) metricsUpdateWorker(ctx context.Context, config *DashboardConfig) {
	ticker := time.NewTicker(config.RefreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.updateMetrics(ctx)
		}
	}
}

// alertWorker surveille les métriques et génère des alertes
func (s *DashboardService) alertWorker(ctx context.Context, config *DashboardConfig) {
	ticker := time.NewTicker(1 * time.Minute) // Vérification toutes les minutes
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.checkAlerts(ctx, config.AlertThresholds)
		}
	}
}

// cleanupWorker nettoie les abonnés inactifs
func (s *DashboardService) cleanupWorker(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.cleanupInactiveSubscribers()
		}
	}
}

// updateMetrics met à jour toutes les métriques
func (s *DashboardService) updateMetrics(ctx context.Context) {
	dateRange := DateRange{
		Start: time.Now().AddDate(0, 0, -1), // Dernières 24h
		End:   time.Now(),
	}

	metrics, err := s.GetDashboardMetrics(ctx, dateRange)
	if err != nil {
		s.logger.Error("Failed to update dashboard metrics", zap.Error(err))
		return
	}

	// Diffuser la mise à jour
	update := &DashboardUpdate{
		Type:      "full_refresh",
		Data:      metrics,
		Timestamp: time.Now(),
	}

	s.BroadcastUpdate(update)
}

// getOverviewMetrics récupère les métriques de vue d'ensemble
func (s *DashboardService) getOverviewMetrics(ctx context.Context, dateRange DateRange) (OverviewMetrics, error) {
	overview := OverviewMetrics{}

	// Requêtes parallèles pour les métriques de base
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Total utilisateurs
	wg.Add(1)
	go func() {
		defer wg.Done()
		var count int64
		query := `SELECT COUNT(*) FROM users`
		s.db.QueryRowContext(ctx, query).Scan(&count)
		mu.Lock()
		overview.TotalUsers = count
		mu.Unlock()
	}()

	// Utilisateurs actifs 24h
	wg.Add(1)
	go func() {
		defer wg.Done()
		var count int64
		query := `SELECT COUNT(DISTINCT user_id) FROM user_sessions WHERE start_time >= $1`
		s.db.QueryRowContext(ctx, query, time.Now().AddDate(0, 0, -1)).Scan(&count)
		mu.Lock()
		overview.ActiveUsers24h = count
		mu.Unlock()
	}()

	// Revenus 24h
	wg.Add(1)
	go func() {
		defer wg.Done()
		var revenue sql.NullFloat64
		query := `
			SELECT COALESCE(SUM(amount), 0) 
			FROM revenue_transactions 
			WHERE timestamp >= $1 AND status = 'completed' AND type != 'refund'`
		s.db.QueryRowContext(ctx, query, time.Now().AddDate(0, 0, -1)).Scan(&revenue)
		mu.Lock()
		if revenue.Valid {
			overview.Revenue24h = revenue.Float64
		}
		mu.Unlock()
	}()

	wg.Wait()

	// Métriques calculées
	overview.ConversionRate = 15.5 // TODO: Calculer le vrai taux
	overview.ChurnRate = 3.8       // TODO: Calculer le vrai taux
	overview.NPS = 72.0            // TODO: Intégrer avec un système de feedback
	overview.ServerUptime = 99.95  // TODO: Calculer depuis le monitoring

	return overview, nil
}

// getRealtimeMetrics récupère les métriques temps réel
func (s *DashboardService) getRealtimeMetrics(ctx context.Context) (RealtimeMetrics, error) {
	// Métriques simulées - dans un vrai système, ces données viendraient
	// d'un système de monitoring comme Prometheus, New Relic, etc.
	return RealtimeMetrics{
		CurrentUsers:      245,
		ActiveStreams:     67,
		MessageRate:       15.8,
		StreamRate:        4.2,
		ErrorRate:         0.12,
		ResponseTime:      45.7,
		BandwidthUsage:    125.6,
		CPUUsage:          34.8,
		MemoryUsage:       52.1,
		DatabaseQueries:   187,
		CacheHitRate:      94.3,
		ActiveConnections: 1250,
		QueueLength:       12,
		GeoDistribution: map[string]int64{
			"France": 85,
			"Belgium": 42,
			"Switzerland": 38,
			"Canada": 35,
			"Germany": 28,
			"UK": 17,
		},
	}, nil
}

// getSystemHealth récupère l'état de santé du système
func (s *DashboardService) getSystemHealth(ctx context.Context) (SystemHealthMetrics, error) {
	// Simulation des checks de santé
	return SystemHealthMetrics{
		OverallHealth:    "healthy",
		UptimePercentage: 99.95,
		Services: map[string]ServiceHealth{
			"api": {
				Status:       "healthy",
				ResponseTime: 45.2,
				ErrorRate:    0.12,
				LastCheck:    time.Now(),
			},
			"websocket": {
				Status:       "healthy",
				ResponseTime: 15.8,
				ErrorRate:    0.05,
				LastCheck:    time.Now(),
			},
			"streaming": {
				Status:       "healthy",
				ResponseTime: 78.5,
				ErrorRate:    0.23,
				LastCheck:    time.Now(),
			},
		},
		Infrastructure: InfrastructureHealth{
			Database: ServiceHealth{
				Status:       "healthy",
				ResponseTime: 12.5,
				ErrorRate:    0.01,
				LastCheck:    time.Now(),
			},
			Redis: ServiceHealth{
				Status:       "healthy",
				ResponseTime: 2.3,
				ErrorRate:    0.0,
				LastCheck:    time.Now(),
			},
		},
	}, nil
}

// getActiveAlerts récupère les alertes actives
func (s *DashboardService) getActiveAlerts(ctx context.Context) ([]DashboardAlert, error) {
	// TODO: Intégrer avec un vrai système d'alertes
	return []DashboardAlert{}, nil
}

// checkAlerts vérifie les seuils d'alerte
func (s *DashboardService) checkAlerts(ctx context.Context, thresholds AlertThresholds) {
	// TODO: Implémenter la logique de vérification des alertes
}

// cleanupInactiveSubscribers nettoie les abonnés inactifs
func (s *DashboardService) cleanupInactiveSubscribers() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// TODO: Implémenter le nettoyage basé sur l'activité
}
