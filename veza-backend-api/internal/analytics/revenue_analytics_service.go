package analytics

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// RevenueAnalyticsService service d'analytics des revenus
type RevenueAnalyticsService struct {
	db     *sql.DB
	logger *zap.Logger
	cache  EngagementCache
}

// RevenueMetrics métriques de revenus complètes
type RevenueMetrics struct {
	TotalRevenue           float64                  `json:"total_revenue"`
	RecurringRevenue       float64                  `json:"recurring_revenue"` // MRR
	OneTimeRevenue         float64                  `json:"one_time_revenue"`
	RevenueGrowth          float64                  `json:"revenue_growth_percent"`
	ARPU                   float64                  `json:"arpu"` // Average Revenue Per User
	LTV                    float64                  `json:"ltv"`  // Customer Lifetime Value
	ChurnRate              float64                  `json:"churn_rate"`
	RevenueBySource        map[string]float64       `json:"revenue_by_source"`
	RevenueByPlan          map[string]RevenueByPlan `json:"revenue_by_plan"`
	RevenueByGeography     map[string]float64       `json:"revenue_by_geography"`
	DailyRevenue           []DailyRevenueData       `json:"daily_revenue"`
	SubscriptionMetrics    SubscriptionMetrics      `json:"subscription_metrics"`
	RefundMetrics          RefundMetrics            `json:"refund_metrics"`
	PaymentMethodBreakdown map[string]float64       `json:"payment_method_breakdown"`
	ConversionFunnel       ConversionFunnelMetrics  `json:"conversion_funnel"`
	CohortRevenue          []CohortRevenueAnalysis  `json:"cohort_revenue"`
	RevenueForecasting     RevenueForecast          `json:"revenue_forecasting"`
}

// RevenueByPlan revenus par plan d'abonnement
type RevenueByPlan struct {
	PlanName    string  `json:"plan_name"`
	Revenue     float64 `json:"revenue"`
	Subscribers int64   `json:"subscribers"`
	ChurnRate   float64 `json:"churn_rate"`
	AverageStay float64 `json:"average_stay_months"`
	UpgradeRate float64 `json:"upgrade_rate"`
}

// DailyRevenueData données de revenus quotidiens
type DailyRevenueData struct {
	Date             time.Time `json:"date"`
	Revenue          float64   `json:"revenue"`
	NewSubscriptions int64     `json:"new_subscriptions"`
	Cancellations    int64     `json:"cancellations"`
	Upgrades         int64     `json:"upgrades"`
	Downgrades       int64     `json:"downgrades"`
	Refunds          float64   `json:"refunds"`
	NetRevenue       float64   `json:"net_revenue"`
}

// SubscriptionMetrics métriques d'abonnement
type SubscriptionMetrics struct {
	TotalSubscribers     int64   `json:"total_subscribers"`
	ActiveSubscribers    int64   `json:"active_subscribers"`
	NewSubscribers       int64   `json:"new_subscribers"`
	CancelledSubscribers int64   `json:"cancelled_subscribers"`
	TrialUsers           int64   `json:"trial_users"`
	TrialConversionRate  float64 `json:"trial_conversion_rate"`
	MRR                  float64 `json:"mrr"` // Monthly Recurring Revenue
	ARR                  float64 `json:"arr"` // Annual Recurring Revenue
	ChurnRate            float64 `json:"churn_rate"`
	NetRevenueRetention  float64 `json:"net_revenue_retention"`
}

// RefundMetrics métriques de remboursement
type RefundMetrics struct {
	TotalRefunds    float64            `json:"total_refunds"`
	RefundCount     int64              `json:"refund_count"`
	RefundRate      float64            `json:"refund_rate_percent"`
	AvgRefundAmount float64            `json:"avg_refund_amount"`
	RefundsByReason map[string]float64 `json:"refunds_by_reason"`
	RefundsByPlan   map[string]float64 `json:"refunds_by_plan"`
}

// ConversionFunnelMetrics métriques d'entonnoir de conversion
type ConversionFunnelMetrics struct {
	Visitors          int64   `json:"visitors"`
	SignUps           int64   `json:"sign_ups"`
	TrialStarts       int64   `json:"trial_starts"`
	PaidConversions   int64   `json:"paid_conversions"`
	VisitorToSignUp   float64 `json:"visitor_to_signup_rate"`
	SignUpToTrial     float64 `json:"signup_to_trial_rate"`
	TrialToPaid       float64 `json:"trial_to_paid_rate"`
	OverallConversion float64 `json:"overall_conversion_rate"`
}

// CohortRevenueAnalysis analyse de revenus par cohorte
type CohortRevenueAnalysis struct {
	CohortMonth       time.Time          `json:"cohort_month"`
	InitialSize       int64              `json:"initial_size"`
	MonthlyRevenue    map[string]float64 `json:"monthly_revenue"` // Revenus par mois depuis création
	RetentionRate     map[string]float64 `json:"retention_rate"`  // Taux de rétention par mois
	CumulativeRevenue float64            `json:"cumulative_revenue"`
}

// RevenueForecast prévisions de revenus
type RevenueForecast struct {
	NextMonthPrediction   float64           `json:"next_month_prediction"`
	NextQuarterPrediction float64           `json:"next_quarter_prediction"`
	AnnualPrediction      float64           `json:"annual_prediction"`
	Confidence            float64           `json:"confidence_percent"`
	ForecastDetails       []MonthlyForecast `json:"forecast_details"`
	GrowthAssumptions     GrowthAssumptions `json:"growth_assumptions"`
}

// MonthlyForecast prévision mensuelle
type MonthlyForecast struct {
	Month            time.Time `json:"month"`
	PredictedRevenue float64   `json:"predicted_revenue"`
	LowerBound       float64   `json:"lower_bound"`
	UpperBound       float64   `json:"upper_bound"`
}

// GrowthAssumptions hypothèses de croissance
type GrowthAssumptions struct {
	MonthlyGrowthRate   float64 `json:"monthly_growth_rate"`
	ChurnRateAssumption float64 `json:"churn_rate_assumption"`
	NewCustomerRate     float64 `json:"new_customer_rate"`
	PriceChangeImpact   float64 `json:"price_change_impact"`
}

// Transaction transaction de revenus
type Transaction struct {
	ID            string                 `json:"id"`
	UserID        string                 `json:"user_id"`
	Amount        float64                `json:"amount"`
	Currency      string                 `json:"currency"`
	Type          string                 `json:"type"`           // subscription, one_time, refund
	Status        string                 `json:"status"`         // completed, pending, failed
	PaymentMethod string                 `json:"payment_method"` // card, paypal, etc.
	PlanID        string                 `json:"plan_id,omitempty"`
	Country       string                 `json:"country"`
	Timestamp     time.Time              `json:"timestamp"`
	RefundReason  string                 `json:"refund_reason,omitempty"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// NewRevenueAnalyticsService crée un nouveau service d'analytics des revenus
func NewRevenueAnalyticsService(db *sql.DB, logger *zap.Logger, cache EngagementCache) *RevenueAnalyticsService {
	return &RevenueAnalyticsService{
		db:     db,
		logger: logger,
		cache:  cache,
	}
}

// TrackTransaction enregistre une transaction
func (s *RevenueAnalyticsService) TrackTransaction(ctx context.Context, transaction *Transaction) error {
	query := `
		INSERT INTO revenue_transactions (
			id, user_id, amount, currency, type, status, payment_method,
			plan_id, country, timestamp, refund_reason, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	metadataJSON, _ := json.Marshal(transaction.Metadata)

	_, err := s.db.ExecContext(ctx, query,
		transaction.ID, transaction.UserID, transaction.Amount, transaction.Currency,
		transaction.Type, transaction.Status, transaction.PaymentMethod,
		transaction.PlanID, transaction.Country, transaction.Timestamp,
		transaction.RefundReason, metadataJSON)

	if err != nil {
		s.logger.Error("Failed to track transaction", zap.Error(err))
		return fmt.Errorf("failed to track transaction: %w", err)
	}

	s.logger.Debug("Transaction tracked",
		zap.String("transaction_id", transaction.ID),
		zap.Float64("amount", transaction.Amount))

	return nil
}

// GetRevenueMetrics retourne les métriques complètes de revenus
func (s *RevenueAnalyticsService) GetRevenueMetrics(ctx context.Context, dateRange DateRange) (*RevenueMetrics, error) {
	metrics := &RevenueMetrics{}

	// Revenus totaux
	var err error
	metrics.TotalRevenue, err = s.getTotalRevenue(ctx, dateRange)
	if err != nil {
		return nil, err
	}

	// Revenus récurrents vs one-time
	metrics.RecurringRevenue, metrics.OneTimeRevenue, err = s.getRevenueBreakdown(ctx, dateRange)
	if err != nil {
		return nil, err
	}

	// Croissance des revenus
	metrics.RevenueGrowth, err = s.getRevenueGrowth(ctx, dateRange)
	if err != nil {
		return nil, err
	}

	// ARPU et LTV
	metrics.ARPU, err = s.getARPU(ctx, dateRange)
	if err != nil {
		return nil, err
	}

	metrics.LTV, err = s.getLTV(ctx, dateRange)
	if err != nil {
		return nil, err
	}

	// Revenus par source
	metrics.RevenueBySource, err = s.getRevenueBySource(ctx, dateRange)
	if err != nil {
		return nil, err
	}

	// Revenus par plan
	metrics.RevenueByPlan, err = s.getRevenueByPlan(ctx, dateRange)
	if err != nil {
		return nil, err
	}

	// Revenus par géographie
	metrics.RevenueByGeography, err = s.getRevenueByGeography(ctx, dateRange)
	if err != nil {
		return nil, err
	}

	// Revenus quotidiens
	metrics.DailyRevenue, err = s.getDailyRevenue(ctx, dateRange)
	if err != nil {
		return nil, err
	}

	// Métriques d'abonnement
	metrics.SubscriptionMetrics, err = s.getSubscriptionMetrics(ctx, dateRange)
	if err != nil {
		return nil, err
	}

	// Métriques de remboursement
	metrics.RefundMetrics, err = s.getRefundMetrics(ctx, dateRange)
	if err != nil {
		return nil, err
	}

	// Entonnoir de conversion
	metrics.ConversionFunnel, err = s.getConversionFunnel(ctx, dateRange)
	if err != nil {
		return nil, err
	}

	// Prévisions
	metrics.RevenueForecasting, err = s.getRevenueForecast(ctx, dateRange)
	if err != nil {
		return nil, err
	}

	return metrics, nil
}

// ============================================================================
// MÉTHODES PRIVÉES
// ============================================================================

func (s *RevenueAnalyticsService) getTotalRevenue(ctx context.Context, dateRange DateRange) (float64, error) {
	query := `
		SELECT COALESCE(SUM(amount), 0) 
		FROM revenue_transactions 
		WHERE timestamp >= $1 AND timestamp <= $2 
		AND status = 'completed' 
		AND type != 'refund'`

	var total float64
	err := s.db.QueryRowContext(ctx, query, dateRange.Start, dateRange.End).Scan(&total)
	return total, err
}

func (s *RevenueAnalyticsService) getRevenueBreakdown(ctx context.Context, dateRange DateRange) (float64, float64, error) {
	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN type = 'subscription' THEN amount END), 0) as recurring,
			COALESCE(SUM(CASE WHEN type = 'one_time' THEN amount END), 0) as one_time
		FROM revenue_transactions 
		WHERE timestamp >= $1 AND timestamp <= $2 
		AND status = 'completed'`

	var recurring, oneTime float64
	err := s.db.QueryRowContext(ctx, query, dateRange.Start, dateRange.End).Scan(&recurring, &oneTime)
	return recurring, oneTime, err
}

func (s *RevenueAnalyticsService) getRevenueGrowth(ctx context.Context, dateRange DateRange) (float64, error) {
	// Calculer la croissance par rapport à la période précédente
	periodLength := dateRange.End.Sub(dateRange.Start)
	previousStart := dateRange.Start.Add(-periodLength)
	previousEnd := dateRange.Start

	currentRevenue, err := s.getTotalRevenue(ctx, dateRange)
	if err != nil {
		return 0, err
	}

	previousRevenue, err := s.getTotalRevenue(ctx, DateRange{Start: previousStart, End: previousEnd})
	if err != nil {
		return 0, err
	}

	if previousRevenue == 0 {
		return 0, nil
	}

	growth := ((currentRevenue - previousRevenue) / previousRevenue) * 100
	return growth, nil
}

func (s *RevenueAnalyticsService) getARPU(ctx context.Context, dateRange DateRange) (float64, error) {
	query := `
		SELECT 
			COALESCE(SUM(amount), 0) / NULLIF(COUNT(DISTINCT user_id), 0) as arpu
		FROM revenue_transactions 
		WHERE timestamp >= $1 AND timestamp <= $2 
		AND status = 'completed' 
		AND type != 'refund'`

	var arpu sql.NullFloat64
	err := s.db.QueryRowContext(ctx, query, dateRange.Start, dateRange.End).Scan(&arpu)
	if !arpu.Valid {
		return 0, err
	}
	return arpu.Float64, err
}

func (s *RevenueAnalyticsService) getLTV(ctx context.Context, dateRange DateRange) (float64, error) {
	// LTV = (ARPU × Gross Margin %) / Churn Rate
	arpu, err := s.getARPU(ctx, dateRange)
	if err != nil {
		return 0, err
	}

	// TODO: Obtenir le taux de churn réel depuis la base de données
	churnRate := 0.05  // 5% par mois (exemple)
	grossMargin := 0.8 // 80% de marge (exemple)

	if churnRate == 0 {
		return 0, nil
	}

	ltv := (arpu * grossMargin) / churnRate
	return ltv, nil
}

func (s *RevenueAnalyticsService) getRevenueBySource(ctx context.Context, dateRange DateRange) (map[string]float64, error) {
	query := `
		SELECT 
			COALESCE(payment_method, 'unknown') as source,
			SUM(amount) as revenue
		FROM revenue_transactions 
		WHERE timestamp >= $1 AND timestamp <= $2 
		AND status = 'completed' 
		AND type != 'refund'
		GROUP BY payment_method`

	rows, err := s.db.QueryContext(ctx, query, dateRange.Start, dateRange.End)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	revenueBySource := make(map[string]float64)
	for rows.Next() {
		var source string
		var revenue float64
		if err := rows.Scan(&source, &revenue); err != nil {
			continue
		}
		revenueBySource[source] = revenue
	}

	return revenueBySource, nil
}

func (s *RevenueAnalyticsService) getRevenueByPlan(ctx context.Context, dateRange DateRange) (map[string]RevenueByPlan, error) {
	query := `
		SELECT 
			COALESCE(rt.plan_id, 'unknown') as plan_id,
			COALESCE(sp.name, 'Unknown Plan') as plan_name,
			SUM(rt.amount) as revenue,
			COUNT(DISTINCT rt.user_id) as subscribers
		FROM revenue_transactions rt
		LEFT JOIN subscription_plans sp ON sp.id = rt.plan_id
		WHERE rt.timestamp >= $1 AND rt.timestamp <= $2 
		AND rt.status = 'completed' 
		AND rt.type = 'subscription'
		GROUP BY rt.plan_id, sp.name`

	rows, err := s.db.QueryContext(ctx, query, dateRange.Start, dateRange.End)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	revenueByPlan := make(map[string]RevenueByPlan)
	for rows.Next() {
		var planID, planName string
		var revenue float64
		var subscribers int64

		if err := rows.Scan(&planID, &planName, &revenue, &subscribers); err != nil {
			continue
		}

		revenueByPlan[planID] = RevenueByPlan{
			PlanName:    planName,
			Revenue:     revenue,
			Subscribers: subscribers,
			ChurnRate:   0.0, // TODO: Calculer le churn par plan
			AverageStay: 0.0, // TODO: Calculer la durée moyenne
			UpgradeRate: 0.0, // TODO: Calculer le taux d'upgrade
		}
	}

	return revenueByPlan, nil
}

func (s *RevenueAnalyticsService) getRevenueByGeography(ctx context.Context, dateRange DateRange) (map[string]float64, error) {
	query := `
		SELECT country, SUM(amount) as revenue
		FROM revenue_transactions 
		WHERE timestamp >= $1 AND timestamp <= $2 
		AND status = 'completed' 
		AND type != 'refund'
		AND country IS NOT NULL
		GROUP BY country
		ORDER BY revenue DESC
		LIMIT 20`

	rows, err := s.db.QueryContext(ctx, query, dateRange.Start, dateRange.End)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	revenueByGeo := make(map[string]float64)
	for rows.Next() {
		var country string
		var revenue float64
		if err := rows.Scan(&country, &revenue); err != nil {
			continue
		}
		revenueByGeo[country] = revenue
	}

	return revenueByGeo, nil
}

func (s *RevenueAnalyticsService) getDailyRevenue(ctx context.Context, dateRange DateRange) ([]DailyRevenueData, error) {
	query := `
		SELECT 
			DATE(timestamp) as date,
			SUM(CASE WHEN type != 'refund' AND status = 'completed' THEN amount ELSE 0 END) as revenue,
			COUNT(CASE WHEN type = 'subscription' AND status = 'completed' THEN 1 END) as new_subscriptions,
			SUM(CASE WHEN type = 'refund' THEN amount ELSE 0 END) as refunds
		FROM revenue_transactions 
		WHERE timestamp >= $1 AND timestamp <= $2
		GROUP BY DATE(timestamp)
		ORDER BY date`

	rows, err := s.db.QueryContext(ctx, query, dateRange.Start, dateRange.End)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dailyRevenue []DailyRevenueData
	for rows.Next() {
		var data DailyRevenueData
		var refunds float64

		err := rows.Scan(
			&data.Date,
			&data.Revenue,
			&data.NewSubscriptions,
			&refunds,
		)
		if err != nil {
			continue
		}

		data.Refunds = refunds
		data.NetRevenue = data.Revenue - refunds

		dailyRevenue = append(dailyRevenue, data)
	}

	return dailyRevenue, nil
}

func (s *RevenueAnalyticsService) getSubscriptionMetrics(ctx context.Context, dateRange DateRange) (SubscriptionMetrics, error) {
	// TODO: Implémenter les métriques d'abonnement détaillées
	return SubscriptionMetrics{
		TotalSubscribers:     1250,
		ActiveSubscribers:    1180,
		NewSubscribers:       125,
		CancelledSubscribers: 45,
		TrialUsers:           350,
		TrialConversionRate:  15.5,
		MRR:                  45000.0,
		ARR:                  540000.0,
		ChurnRate:            3.8,
		NetRevenueRetention:  108.5,
	}, nil
}

func (s *RevenueAnalyticsService) getRefundMetrics(ctx context.Context, dateRange DateRange) (RefundMetrics, error) {
	query := `
		SELECT 
			COUNT(*) as refund_count,
			SUM(amount) as total_refunds,
			AVG(amount) as avg_refund
		FROM revenue_transactions 
		WHERE timestamp >= $1 AND timestamp <= $2 
		AND type = 'refund'`

	var refundCount int64
	var totalRefunds, avgRefund sql.NullFloat64

	err := s.db.QueryRowContext(ctx, query, dateRange.Start, dateRange.End).Scan(&refundCount, &totalRefunds, &avgRefund)
	if err != nil {
		return RefundMetrics{}, err
	}

	metrics := RefundMetrics{
		RefundCount: refundCount,
	}

	if totalRefunds.Valid {
		metrics.TotalRefunds = totalRefunds.Float64
	}
	if avgRefund.Valid {
		metrics.AvgRefundAmount = avgRefund.Float64
	}

	// Calculer le taux de remboursement
	totalRevenue, _ := s.getTotalRevenue(ctx, dateRange)
	if totalRevenue > 0 {
		metrics.RefundRate = (metrics.TotalRefunds / totalRevenue) * 100
	}

	return metrics, nil
}

func (s *RevenueAnalyticsService) getConversionFunnel(ctx context.Context, dateRange DateRange) (ConversionFunnelMetrics, error) {
	// TODO: Intégrer avec les métriques d'engagement utilisateur
	return ConversionFunnelMetrics{
		Visitors:          15420,
		SignUps:           2850,
		TrialStarts:       1250,
		PaidConversions:   380,
		VisitorToSignUp:   18.5,
		SignUpToTrial:     43.9,
		TrialToPaid:       30.4,
		OverallConversion: 2.46,
	}, nil
}

func (s *RevenueAnalyticsService) getRevenueForecast(ctx context.Context, dateRange DateRange) (RevenueForecast, error) {
	// Prévision simple basée sur la tendance des 3 derniers mois
	currentRevenue, _ := s.getTotalRevenue(ctx, dateRange)

	// TODO: Implémenter un modèle de prévision plus sophistiqué
	forecast := RevenueForecast{
		NextMonthPrediction:   currentRevenue * 1.08, // +8% de croissance
		NextQuarterPrediction: currentRevenue * 3.25, // +8.33% par mois
		AnnualPrediction:      currentRevenue * 13.0, // +8.33% par mois
		Confidence:            75.0,                  // 75% de confiance
		GrowthAssumptions: GrowthAssumptions{
			MonthlyGrowthRate:   8.0,
			ChurnRateAssumption: 5.0,
			NewCustomerRate:     12.0,
			PriceChangeImpact:   0.0,
		},
	}

	return forecast, nil
}
