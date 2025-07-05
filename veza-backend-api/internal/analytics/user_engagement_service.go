package analytics

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"go.uber.org/zap"
)

// UserEngagementService service de tracking d'engagement utilisateur
type UserEngagementService struct {
	db     *sql.DB
	logger *zap.Logger
	cache  EngagementCache
}

// UserEngagementMetrics métriques d'engagement utilisateur
type UserEngagementMetrics struct {
	DAU             int64                   `json:"dau"`                   // Daily Active Users
	WAU             int64                   `json:"wau"`                   // Weekly Active Users
	MAU             int64                   `json:"mau"`                   // Monthly Active Users
	NewUsers        int64                   `json:"new_users"`             // Nouveaux utilisateurs
	ReturningUsers  int64                   `json:"returning_users"`       // Utilisateurs de retour
	SessionDuration float64                 `json:"avg_session_duration"`  // Durée moyenne de session (minutes)
	PagesPerSession float64                 `json:"avg_pages_per_session"` // Pages par session
	BounceRate      float64                 `json:"bounce_rate"`           // Taux de rebond (%)
	RetentionRate   float64                 `json:"retention_rate"`        // Taux de rétention (%)
	ChurnRate       float64                 `json:"churn_rate"`            // Taux de désabonnement (%)
	EngagementScore float64                 `json:"engagement_score"`      // Score d'engagement (0-100)
	TopActions      []ActionMetric          `json:"top_actions"`           // Actions les plus fréquentes
	DeviceBreakdown map[string]int64        `json:"device_breakdown"`      // Répartition par device
	GeographicData  map[string]int64        `json:"geographic_data"`       // Données géographiques
	CohortAnalysis  map[string]CohortMetric `json:"cohort_analysis"`       // Analyse de cohorte
	FunnelMetrics   map[string]FunnelStep   `json:"funnel_metrics"`        // Métriques d'entonnoir
}

// ActionMetric métrique d'action utilisateur
type ActionMetric struct {
	Action      string  `json:"action"`
	Count       int64   `json:"count"`
	UniqueUsers int64   `json:"unique_users"`
	AvgDuration float64 `json:"avg_duration_seconds"`
}

// CohortMetric métrique de cohorte
type CohortMetric struct {
	CohortDate     time.Time `json:"cohort_date"`
	InitialSize    int64     `json:"initial_size"`
	Day1Retention  float64   `json:"day1_retention"`
	Day7Retention  float64   `json:"day7_retention"`
	Day30Retention float64   `json:"day30_retention"`
	Revenue        float64   `json:"revenue"`
}

// FunnelStep étape d'entonnoir de conversion
type FunnelStep struct {
	Step        string  `json:"step"`
	Users       int64   `json:"users"`
	Conversions int64   `json:"conversions"`
	Rate        float64 `json:"conversion_rate"`
}

// UserSession session utilisateur
type UserSession struct {
	ID             string                 `json:"id"`
	UserID         string                 `json:"user_id"`
	StartTime      time.Time              `json:"start_time"`
	EndTime        *time.Time             `json:"end_time,omitempty"`
	Duration       *int64                 `json:"duration_seconds,omitempty"`
	PageViews      int64                  `json:"page_views"`
	Actions        int64                  `json:"actions"`
	DeviceType     string                 `json:"device_type"`
	UserAgent      string                 `json:"user_agent"`
	IPAddress      string                 `json:"ip_address"`
	Country        string                 `json:"country"`
	City           string                 `json:"city"`
	ReferrerSource string                 `json:"referrer_source"`
	EntryPage      string                 `json:"entry_page"`
	ExitPage       string                 `json:"exit_page"`
	IsNewUser      bool                   `json:"is_new_user"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// UserAction action utilisateur trackée
type UserAction struct {
	ID         string                 `json:"id"`
	UserID     string                 `json:"user_id"`
	SessionID  string                 `json:"session_id"`
	Action     string                 `json:"action"`
	Category   string                 `json:"category"`
	Label      string                 `json:"label,omitempty"`
	Value      *float64               `json:"value,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
	Duration   *int64                 `json:"duration_seconds,omitempty"`
	Properties map[string]interface{} `json:"properties"`
}

// EngagementCache interface de cache pour les métriques
type EngagementCache interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, ttl time.Duration)
	Delete(key string)
}

// NewUserEngagementService crée un nouveau service d'engagement utilisateur
func NewUserEngagementService(db *sql.DB, logger *zap.Logger, cache EngagementCache) *UserEngagementService {
	return &UserEngagementService{
		db:     db,
		logger: logger,
		cache:  cache,
	}
}

// TrackSession démarre le tracking d'une session utilisateur
func (s *UserEngagementService) TrackSession(ctx context.Context, session *UserSession) error {
	query := `
		INSERT INTO user_sessions (
			id, user_id, start_time, device_type, user_agent, ip_address,
			country, city, referrer_source, entry_page, is_new_user, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	metadataJSON, _ := json.Marshal(session.Metadata)

	_, err := s.db.ExecContext(ctx, query,
		session.ID, session.UserID, session.StartTime, session.DeviceType,
		session.UserAgent, session.IPAddress, session.Country, session.City,
		session.ReferrerSource, session.EntryPage, session.IsNewUser, metadataJSON)

	if err != nil {
		s.logger.Error("Failed to track session", zap.Error(err))
		return fmt.Errorf("failed to track session: %w", err)
	}

	s.logger.Debug("Session tracked",
		zap.String("session_id", session.ID),
		zap.String("user_id", session.UserID))

	return nil
}

// EndSession termine une session utilisateur
func (s *UserEngagementService) EndSession(ctx context.Context, sessionID string, exitPage string) error {
	query := `
		UPDATE user_sessions 
		SET end_time = NOW(), 
		    duration = EXTRACT(EPOCH FROM (NOW() - start_time)),
		    exit_page = $2
		WHERE id = $1`

	_, err := s.db.ExecContext(ctx, query, sessionID, exitPage)
	if err != nil {
		s.logger.Error("Failed to end session", zap.Error(err))
		return fmt.Errorf("failed to end session: %w", err)
	}

	return nil
}

// TrackAction enregistre une action utilisateur
func (s *UserEngagementService) TrackAction(ctx context.Context, action *UserAction) error {
	query := `
		INSERT INTO user_actions (
			id, user_id, session_id, action, category, label, value,
			timestamp, duration, properties
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	propertiesJSON, _ := json.Marshal(action.Properties)

	_, err := s.db.ExecContext(ctx, query,
		action.ID, action.UserID, action.SessionID, action.Action,
		action.Category, action.Label, action.Value, action.Timestamp,
		action.Duration, propertiesJSON)

	if err != nil {
		s.logger.Error("Failed to track action", zap.Error(err))
		return fmt.Errorf("failed to track action: %w", err)
	}

	// Mettre à jour le compteur d'actions de la session
	updateQuery := `UPDATE user_sessions SET actions = actions + 1 WHERE id = $1`
	if _, err := s.db.ExecContext(ctx, updateQuery, action.SessionID); err != nil {
		s.logger.Error("Failed to update session stats", zap.Error(err))
	}

	return nil
}

// TrackPageView enregistre une vue de page
func (s *UserEngagementService) TrackPageView(ctx context.Context, sessionID, page string) error {
	updateQuery := `UPDATE user_sessions SET page_views = page_views + 1 WHERE id = $1`
	_, err := s.db.ExecContext(ctx, updateQuery, sessionID)

	if err != nil {
		return fmt.Errorf("failed to track page view: %w", err)
	}

	// Enregistrer aussi comme action
	action := &UserAction{
		ID:        generateActionID(),
		SessionID: sessionID,
		Action:    "page_view",
		Category:  "navigation",
		Label:     page,
		Timestamp: time.Now(),
	}

	return s.TrackAction(ctx, action)
}

// GetDailyActiveUsers retourne le nombre d'utilisateurs actifs quotidiens
func (s *UserEngagementService) GetDailyActiveUsers(ctx context.Context, date time.Time) (int64, error) {
	cacheKey := fmt.Sprintf("dau:%s", date.Format("2006-01-02"))

	if cached, exists := s.cache.Get(cacheKey); exists {
		return cached.(int64), nil
	}

	query := `
		SELECT COUNT(DISTINCT user_id) 
		FROM user_sessions 
		WHERE DATE(start_time) = DATE($1)`

	var dau int64
	err := s.db.QueryRowContext(ctx, query, date).Scan(&dau)
	if err != nil {
		return 0, fmt.Errorf("failed to get DAU: %w", err)
	}

	// Cache pour 1 heure
	s.cache.Set(cacheKey, dau, time.Hour)

	return dau, nil
}

// GetWeeklyActiveUsers retourne le nombre d'utilisateurs actifs hebdomadaires
func (s *UserEngagementService) GetWeeklyActiveUsers(ctx context.Context, date time.Time) (int64, error) {
	weekStart := date.AddDate(0, 0, -6) // 7 jours en arrière

	query := `
		SELECT COUNT(DISTINCT user_id) 
		FROM user_sessions 
		WHERE start_time >= $1 AND start_time <= $2`

	var wau int64
	err := s.db.QueryRowContext(ctx, query, weekStart, date).Scan(&wau)
	if err != nil {
		return 0, fmt.Errorf("failed to get WAU: %w", err)
	}

	return wau, nil
}

// GetMonthlyActiveUsers retourne le nombre d'utilisateurs actifs mensuels
func (s *UserEngagementService) GetMonthlyActiveUsers(ctx context.Context, date time.Time) (int64, error) {
	monthStart := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
	monthEnd := monthStart.AddDate(0, 1, -1)

	query := `
		SELECT COUNT(DISTINCT user_id) 
		FROM user_sessions 
		WHERE start_time >= $1 AND start_time <= $2`

	var mau int64
	err := s.db.QueryRowContext(ctx, query, monthStart, monthEnd).Scan(&mau)
	if err != nil {
		return 0, fmt.Errorf("failed to get MAU: %w", err)
	}

	return mau, nil
}

// GetEngagementMetrics retourne les métriques d'engagement complètes
func (s *UserEngagementService) GetEngagementMetrics(ctx context.Context, dateRange DateRange) (*UserEngagementMetrics, error) {
	metrics := &UserEngagementMetrics{}

	// DAU/WAU/MAU
	var err error
	metrics.DAU, err = s.GetDailyActiveUsers(ctx, dateRange.End)
	if err != nil {
		return nil, err
	}

	metrics.WAU, err = s.GetWeeklyActiveUsers(ctx, dateRange.End)
	if err != nil {
		return nil, err
	}

	metrics.MAU, err = s.GetMonthlyActiveUsers(ctx, dateRange.End)
	if err != nil {
		return nil, err
	}

	// Nouveaux vs Utilisateurs de retour
	metrics.NewUsers, metrics.ReturningUsers, err = s.getNewVsReturningUsers(ctx, dateRange)
	if err != nil {
		return nil, err
	}

	// Durée moyenne de session
	metrics.SessionDuration, err = s.getAverageSessionDuration(ctx, dateRange)
	if err != nil {
		return nil, err
	}

	// Pages par session
	metrics.PagesPerSession, err = s.getAveragePagesPerSession(ctx, dateRange)
	if err != nil {
		return nil, err
	}

	// Taux de rebond
	metrics.BounceRate, err = s.getBounceRate(ctx, dateRange)
	if err != nil {
		return nil, err
	}

	// Score d'engagement
	metrics.EngagementScore = s.calculateEngagementScore(metrics)

	// Actions les plus fréquentes
	metrics.TopActions, err = s.getTopActions(ctx, dateRange, 10)
	if err != nil {
		return nil, err
	}

	// Répartition par device
	metrics.DeviceBreakdown, err = s.getDeviceBreakdown(ctx, dateRange)
	if err != nil {
		return nil, err
	}

	// Données géographiques
	metrics.GeographicData, err = s.getGeographicData(ctx, dateRange)
	if err != nil {
		return nil, err
	}

	return metrics, nil
}

// ============================================================================
// MÉTHODES PRIVÉES POUR CALCULS SPÉCIFIQUES
// ============================================================================

func (s *UserEngagementService) getNewVsReturningUsers(ctx context.Context, dateRange DateRange) (int64, int64, error) {
	query := `
		SELECT 
			COUNT(CASE WHEN is_new_user = true THEN 1 END) as new_users,
			COUNT(CASE WHEN is_new_user = false THEN 1 END) as returning_users
		FROM user_sessions 
		WHERE start_time >= $1 AND start_time <= $2`

	var newUsers, returningUsers int64
	err := s.db.QueryRowContext(ctx, query, dateRange.Start, dateRange.End).Scan(&newUsers, &returningUsers)

	return newUsers, returningUsers, err
}

func (s *UserEngagementService) getAverageSessionDuration(ctx context.Context, dateRange DateRange) (float64, error) {
	query := `
		SELECT AVG(duration) / 60.0 as avg_duration_minutes
		FROM user_sessions 
		WHERE start_time >= $1 AND start_time <= $2 AND duration IS NOT NULL`

	var avgDuration sql.NullFloat64
	err := s.db.QueryRowContext(ctx, query, dateRange.Start, dateRange.End).Scan(&avgDuration)

	if err != nil || !avgDuration.Valid {
		return 0, err
	}

	return avgDuration.Float64, nil
}

func (s *UserEngagementService) getAveragePagesPerSession(ctx context.Context, dateRange DateRange) (float64, error) {
	query := `
		SELECT AVG(page_views) 
		FROM user_sessions 
		WHERE start_time >= $1 AND start_time <= $2`

	var avgPages sql.NullFloat64
	err := s.db.QueryRowContext(ctx, query, dateRange.Start, dateRange.End).Scan(&avgPages)

	if err != nil || !avgPages.Valid {
		return 0, err
	}

	return avgPages.Float64, nil
}

func (s *UserEngagementService) getBounceRate(ctx context.Context, dateRange DateRange) (float64, error) {
	query := `
		SELECT 
			COUNT(CASE WHEN page_views = 1 THEN 1 END) * 100.0 / COUNT(*) as bounce_rate
		FROM user_sessions 
		WHERE start_time >= $1 AND start_time <= $2`

	var bounceRate sql.NullFloat64
	err := s.db.QueryRowContext(ctx, query, dateRange.Start, dateRange.End).Scan(&bounceRate)

	if err != nil || !bounceRate.Valid {
		return 0, err
	}

	return bounceRate.Float64, nil
}

func (s *UserEngagementService) getTopActions(ctx context.Context, dateRange DateRange, limit int) ([]ActionMetric, error) {
	query := `
		SELECT 
			action,
			COUNT(*) as count,
			COUNT(DISTINCT user_id) as unique_users,
			AVG(COALESCE(duration, 0)) as avg_duration
		FROM user_actions 
		WHERE timestamp >= $1 AND timestamp <= $2
		GROUP BY action
		ORDER BY count DESC
		LIMIT $3`

	rows, err := s.db.QueryContext(ctx, query, dateRange.Start, dateRange.End, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var actions []ActionMetric
	for rows.Next() {
		var action ActionMetric
		err := rows.Scan(&action.Action, &action.Count, &action.UniqueUsers, &action.AvgDuration)
		if err != nil {
			continue
		}
		actions = append(actions, action)
	}

	return actions, nil
}

func (s *UserEngagementService) getDeviceBreakdown(ctx context.Context, dateRange DateRange) (map[string]int64, error) {
	query := `
		SELECT device_type, COUNT(DISTINCT user_id)
		FROM user_sessions 
		WHERE start_time >= $1 AND start_time <= $2
		GROUP BY device_type`

	rows, err := s.db.QueryContext(ctx, query, dateRange.Start, dateRange.End)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	breakdown := make(map[string]int64)
	for rows.Next() {
		var deviceType string
		var count int64
		if err := rows.Scan(&deviceType, &count); err != nil {
			continue
		}
		breakdown[deviceType] = count
	}

	return breakdown, nil
}

func (s *UserEngagementService) getGeographicData(ctx context.Context, dateRange DateRange) (map[string]int64, error) {
	query := `
		SELECT country, COUNT(DISTINCT user_id)
		FROM user_sessions 
		WHERE start_time >= $1 AND start_time <= $2 AND country IS NOT NULL
		GROUP BY country
		ORDER BY COUNT(DISTINCT user_id) DESC
		LIMIT 20`

	rows, err := s.db.QueryContext(ctx, query, dateRange.Start, dateRange.End)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	geoData := make(map[string]int64)
	for rows.Next() {
		var country string
		var count int64
		if err := rows.Scan(&country, &count); err != nil {
			continue
		}
		geoData[country] = count
	}

	return geoData, nil
}

// calculateEngagementScore calcule un score d'engagement composite (0-100)
func (s *UserEngagementService) calculateEngagementScore(metrics *UserEngagementMetrics) float64 {
	// Algorithme de scoring basé sur plusieurs facteurs
	score := 0.0

	// Facteur durée de session (max 25 points)
	if metrics.SessionDuration > 0 {
		score += min(metrics.SessionDuration/10.0*25, 25) // 10 minutes = 25 points
	}

	// Facteur pages par session (max 20 points)
	if metrics.PagesPerSession > 0 {
		score += min(metrics.PagesPerSession/5.0*20, 20) // 5 pages = 20 points
	}

	// Facteur taux de rebond inversé (max 25 points)
	if metrics.BounceRate >= 0 {
		score += (100 - metrics.BounceRate) / 100 * 25
	}

	// Facteur rétention (max 30 points)
	if metrics.RetentionRate > 0 {
		score += metrics.RetentionRate / 100 * 30
	}

	return min(score, 100)
}

// DateRange plage de dates pour les requêtes
type DateRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// Helper functions
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func generateActionID() string {
	return fmt.Sprintf("action_%d_%d", time.Now().UnixNano(), rand.Intn(10000))
}
