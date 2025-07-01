package database

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

// QueryOptimizationService optimise les requêtes pour éviter les problèmes N+1
type QueryOptimizationService struct {
	db     *sqlx.DB
	logger *zap.Logger
	
	// Cache des requêtes préparées
	preparedStatements map[string]*sqlx.Stmt
	
	// Métriques
	metrics *QueryMetrics
	
	// Configuration
	config *QueryConfig
}

// QueryConfig configuration pour l'optimisation des requêtes
type QueryConfig struct {
	EnablePreparedStatements bool          `json:"enable_prepared_statements"`
	QueryTimeout            time.Duration `json:"query_timeout"`
	SlowQueryThreshold      time.Duration `json:"slow_query_threshold"`
	MaxBatchSize            int           `json:"max_batch_size"`
	EnableQueryPlan         bool          `json:"enable_query_plan"`
	CacheSize               int           `json:"cache_size"`
}

// DefaultQueryConfig configuration par défaut
func DefaultQueryConfig() *QueryConfig {
	return &QueryConfig{
		EnablePreparedStatements: true,
		QueryTimeout:            30 * time.Second,
		SlowQueryThreshold:      100 * time.Millisecond,
		MaxBatchSize:           1000,
		EnableQueryPlan:        true,
		CacheSize:              100,
	}
}

// QueryMetrics métriques pour les requêtes
type QueryMetrics struct {
	queryDuration    *prometheus.HistogramVec
	slowQueries      *prometheus.CounterVec
	batchOperations  *prometheus.CounterVec
	preparedStmtHits *prometheus.CounterVec
	n1Problems       *prometheus.CounterVec
}

// NewQueryMetrics crée les métriques Prometheus
func NewQueryMetrics() *QueryMetrics {
	return &QueryMetrics{
		queryDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name: "db_query_queryDuration_seconds",
			Help: "Durée des requêtes optimisées",
			Buckets: []float64{0.001, 0.01, 0.1, 0.5, 1.0, 2.0, 5.0},
		}, []string{"query_type", "optimization"}),
		slowQueries: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "db_slow_queries_total",
			Help: "Nombre de requêtes lentes détectées",
		}, []string{"query_type", "threshold"}),
		batchOperations: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "db_batch_operations_total",
			Help: "Nombre d'opérations batch",
		}, []string{"operation_type", "batch_size"}),
		preparedStmtHits: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "db_prepared_stmt_hits_total",
			Help: "Cache hits pour requêtes préparées",
		}, []string{"statement_key"}),
		n1Problems: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "db_n1_problems_prevented_total",
			Help: "Problèmes N+1 évités par optimisation",
		}, []string{"entity_type"}),
	}
}

// NewQueryOptimizationService crée un nouveau service d'optimisation
func NewQueryOptimizationService(db *sqlx.DB, logger *zap.Logger) *QueryOptimizationService {
	return &QueryOptimizationService{
		db:                 db,
		logger:            logger,
		preparedStatements: make(map[string]*sqlx.Stmt),
		metrics:           NewQueryMetrics(),
		config:            DefaultQueryConfig(),
	}
}

// OptimizedUserQuery requêtes optimisées pour les utilisateurs
type OptimizedUserQuery struct {
	service *QueryOptimizationService
}

// GetUsersWithRoles récupère les utilisateurs avec leurs rôles (évite N+1)
func (q *OptimizedUserQuery) GetUsersWithRoles(ctx context.Context, userIDs []int64) ([]UserWithRole, error) {
	if len(userIDs) == 0 {
		return []UserWithRole{}, nil
	}
	
	start := time.Now()
	defer func() {
		_ = time.Since(start)
		q.service.metrics.queryDuration.WithLabelValues("users_with_roles", "batch").Observe(time.Since(start).Seconds())
		
		if time.Since(start) > q.service.config.SlowQueryThreshold {
			q.service.metrics.slowQueries.WithLabelValues("users_with_roles", "100ms").Inc()
		}
	}()
	
	// Éviter problème N+1 avec une seule requête JOIN
	query := `
		SELECT 
			u.id, u.username, u.email, u.first_name, u.last_name,
			u.role, u.status, u.is_active, u.created_at,
			r.name as role_name, r.permissions, r.description
		FROM users u
		LEFT JOIN roles r ON u.role = r.code
		WHERE u.id = ANY($1)
		ORDER BY u.created_at DESC
	`
	
	// Convertir slice en format PostgreSQL array
	pqArray := fmt.Sprintf("{%s}", strings.Join(int64SliceToStringSlice(userIDs), ","))
	
	var results []UserWithRole
	if err := q.service.db.SelectContext(ctx, &results, query, pqArray); err != nil {
		return nil, fmt.Errorf("erreur récupération users avec roles: %w", err)
	}
	
	// Métrique pour éviter N+1
	q.service.metrics.n1Problems.WithLabelValues("user_roles").Inc()
	
	return results, nil
}

// GetUsersWithStats récupère les utilisateurs avec leurs statistiques
func (q *OptimizedUserQuery) GetUsersWithStats(ctx context.Context, limit, offset int) ([]UserWithStats, error) {
	start := time.Now()
	defer func() {
		_ = time.Since(start)
		q.service.metrics.queryDuration.WithLabelValues("users_with_stats", "join").Observe(time.Since(start).Seconds())
	}()
	
	// Requête optimisée avec JOIN et sous-requêtes pour éviter N+1
	query := `
		SELECT 
			u.id, u.username, u.email, u.first_name, u.last_name,
			u.role, u.status, u.is_active, u.created_at,
			COALESCE(stats.message_count, 0) as message_count,
			COALESCE(stats.last_activity, u.created_at) as last_activity,
			COALESCE(stats.room_count, 0) as room_count
		FROM users u
		LEFT JOIN (
			SELECT 
				user_id,
				COUNT(*) as message_count,
				MAX(created_at) as last_activity,
				COUNT(DISTINCT room_id) as room_count
			FROM chat_messages 
			WHERE created_at > CURRENT_DATE - INTERVAL '30 days'
			GROUP BY user_id
		) stats ON u.id = stats.user_id
		WHERE u.is_active = true
		ORDER BY u.created_at DESC
		LIMIT $1 OFFSET $2
	`
	
	var results []UserWithStats
	if err := q.service.db.SelectContext(ctx, &results, query, limit, offset); err != nil {
		return nil, fmt.Errorf("erreur récupération users avec stats: %w", err)
	}
	
	q.service.metrics.n1Problems.WithLabelValues("user_stats").Inc()
	
	return results, nil
}

// OptimizedChatQuery requêtes optimisées pour le chat
type OptimizedChatQuery struct {
	service *QueryOptimizationService
}

// GetRoomsWithMessages récupère les rooms avec leurs derniers messages (évite N+1)
func (q *OptimizedChatQuery) GetRoomsWithMessages(ctx context.Context, userID int64, limit int) ([]RoomWithLastMessage, error) {
	start := time.Now()
	defer func() {
		_ = time.Since(start)
		q.service.metrics.queryDuration.WithLabelValues("rooms_with_messages", "window_function").Observe(time.Since(start).Seconds())
	}()
	
	// Utilisation de window functions pour éviter N+1
	query := `
		SELECT DISTINCT ON (r.id)
			r.id, r.name, r.description, r.room_type, r.is_active,
			r.created_at, r.updated_at,
			m.id as last_message_id, m.content as last_message_content,
			m.created_at as last_message_at, m.user_id as last_message_user_id,
			u.username as last_message_username
		FROM chat_rooms r
		LEFT JOIN chat_messages m ON r.id = m.room_id
		LEFT JOIN users u ON m.user_id = u.id
		WHERE r.is_active = true
		  AND (r.room_type = 'public' OR EXISTS(
			  SELECT 1 FROM room_members rm 
			  WHERE rm.room_id = r.id AND rm.user_id = $1
		  ))
		ORDER BY r.id, m.created_at DESC
		LIMIT $2
	`
	
	var results []RoomWithLastMessage
	if err := q.service.db.SelectContext(ctx, &results, query, userID, limit); err != nil {
		return nil, fmt.Errorf("erreur récupération rooms avec messages: %w", err)
	}
	
	q.service.metrics.n1Problems.WithLabelValues("room_messages").Inc()
	
	return results, nil
}

// GetMessagesWithUsers récupère les messages avec les infos utilisateurs en batch
func (q *OptimizedChatQuery) GetMessagesWithUsers(ctx context.Context, roomID int64, limit, offset int) ([]MessageWithUser, error) {
	start := time.Now()
	defer func() {
		_ = time.Since(start)
		q.service.metrics.queryDuration.WithLabelValues("messages_with_users", "join").Observe(time.Since(start).Seconds())
	}()
	
	// JOIN optimisé avec INCLUDE index
	query := `
		SELECT 
			m.id, m.content, m.message_type, m.created_at, m.updated_at,
			m.user_id, u.username, u.avatar, u.role,
			m.reply_to_id, rm.content as reply_content, ru.username as reply_username
		FROM chat_messages m
		INNER JOIN users u ON m.user_id = u.id
		LEFT JOIN chat_messages rm ON m.reply_to_id = rm.id
		LEFT JOIN users ru ON rm.user_id = ru.id
		WHERE m.room_id = $1
		ORDER BY m.created_at DESC
		LIMIT $2 OFFSET $3
	`
	
	var results []MessageWithUser
	if err := q.service.db.SelectContext(ctx, &results, query, roomID, limit, offset); err != nil {
		return nil, fmt.Errorf("erreur récupération messages avec users: %w", err)
	}
	
	q.service.metrics.n1Problems.WithLabelValues("message_users").Inc()
	
	return results, nil
}

// BatchOperations opérations en lot pour optimiser les performances
type BatchOperations struct {
	service *QueryOptimizationService
}

// BatchInsertUsers insertion en lot d'utilisateurs
func (b *BatchOperations) BatchInsertUsers(ctx context.Context, users []UserInsert) error {
	if len(users) == 0 {
		return nil
	}
	
	batchSize := b.service.config.MaxBatchSize
	if len(users) > batchSize {
		return fmt.Errorf("batch trop large: %d > %d", len(users), batchSize)
	}
	
	start := time.Now()
	defer func() {
		_ = time.Since(start)
		b.service.metrics.queryDuration.WithLabelValues("batch_insert_users", "bulk").Observe(time.Since(start).Seconds())
		b.service.metrics.batchOperations.WithLabelValues("insert", fmt.Sprintf("%d", len(users))).Inc()
	}()
	
	// Construire la requête VALUES en lot
	var valueStrings []string
	var valueArgs []interface{}
	
	for i, user := range users {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d)", 
			i*7+1, i*7+2, i*7+3, i*7+4, i*7+5, i*7+6, i*7+7))
		valueArgs = append(valueArgs, user.Username, user.Email, user.PasswordHash, 
			user.FirstName, user.LastName, user.Role, user.Status)
	}
	
	query := fmt.Sprintf(`
		INSERT INTO users (username, email, password_hash, first_name, last_name, role, status)
		VALUES %s
		ON CONFLICT (email) DO NOTHING
	`, strings.Join(valueStrings, ","))
	
	_, err := b.service.db.ExecContext(ctx, query, valueArgs...)
	if err != nil {
		return fmt.Errorf("erreur batch insert users: %w", err)
	}
	
	return nil
}

// BatchUpdateUsers mise à jour en lot d'utilisateurs
func (b *BatchOperations) BatchUpdateUsers(ctx context.Context, updates []UserUpdate) error {
	if len(updates) == 0 {
		return nil
	}
	
	start := time.Now()
	defer func() {
		_ = time.Since(start)
		b.service.metrics.batchOperations.WithLabelValues("update", fmt.Sprintf("%d", len(updates))).Inc()
	}()
	
	// Utiliser UPDATE avec CASE WHEN pour mise à jour en lot
	var ids []int64
	var usernames []string
	var statuses []string
	
	for _, update := range updates {
		ids = append(ids, update.ID)
		usernames = append(usernames, update.Username)
		statuses = append(statuses, update.Status)
	}
	
	// Requête optimisée avec unnest
	query := `
		UPDATE users SET
			username = u.username,
			status = u.status,
			updated_at = CURRENT_TIMESTAMP
		FROM (
			SELECT unnest($1::bigint[]) as id,
				   unnest($2::text[]) as username,
				   unnest($3::text[]) as status
		) u
		WHERE users.id = u.id
	`
	
	_, err := b.service.db.ExecContext(ctx, query, 
		fmt.Sprintf("{%s}", strings.Join(int64SliceToStringSlice(ids), ",")),
		fmt.Sprintf("{%s}", strings.Join(usernames, ",")),
		fmt.Sprintf("{%s}", strings.Join(statuses, ",")))
	
	if err != nil {
		return fmt.Errorf("erreur batch update users: %w", err)
	}
	
	return nil
}

// PreparedStatements gestion des requêtes préparées
func (s *QueryOptimizationService) GetPreparedStatement(key, query string) (*sqlx.Stmt, error) {
	// Vérifier le cache
	if stmt, exists := s.preparedStatements[key]; exists {
		s.metrics.preparedStmtHits.WithLabelValues(key).Inc()
		return stmt, nil
	}
	
	// Préparer la nouvelle requête
	stmt, err := s.db.Preparex(query)
	if err != nil {
		return nil, fmt.Errorf("erreur préparation requête %s: %w", key, err)
	}
	
	// Ajouter au cache
	s.preparedStatements[key] = stmt
	
	s.logger.Debug("Nouvelle requête préparée", 
		zap.String("key", key),
		zap.String("query", query),
	)
	
	return stmt, nil
}

// Types de données pour les requêtes optimisées
type UserWithRole struct {
	ID          int64     `db:"id"`
	Username    string    `db:"username"`
	Email       string    `db:"email"`
	FirstName   string    `db:"first_name"`
	LastName    string    `db:"last_name"`
	Role        string    `db:"role"`
	Status      string    `db:"status"`
	IsActive    bool      `db:"is_active"`
	CreatedAt   time.Time `db:"created_at"`
	RoleName    string    `db:"role_name"`
	Permissions string    `db:"permissions"`
	Description string    `db:"description"`
}

type UserWithStats struct {
	ID           int64     `db:"id"`
	Username     string    `db:"username"`
	Email        string    `db:"email"`
	FirstName    string    `db:"first_name"`
	LastName     string    `db:"last_name"`
	Role         string    `db:"role"`
	Status       string    `db:"status"`
	IsActive     bool      `db:"is_active"`
	CreatedAt    time.Time `db:"created_at"`
	MessageCount int64     `db:"message_count"`
	LastActivity time.Time `db:"last_activity"`
	RoomCount    int64     `db:"room_count"`
}

type RoomWithLastMessage struct {
	ID                   int64     `db:"id"`
	Name                 string    `db:"name"`
	Description          string    `db:"description"`
	RoomType            string    `db:"room_type"`
	IsActive            bool      `db:"is_active"`
	CreatedAt           time.Time `db:"created_at"`
	UpdatedAt           time.Time `db:"updated_at"`
	LastMessageID       *int64    `db:"last_message_id"`
	LastMessageContent  *string   `db:"last_message_content"`
	LastMessageAt       *time.Time `db:"last_message_at"`
	LastMessageUserID   *int64    `db:"last_message_user_id"`
	LastMessageUsername *string   `db:"last_message_username"`
}

type MessageWithUser struct {
	ID           int64     `db:"id"`
	Content      string    `db:"content"`
	MessageType  string    `db:"message_type"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
	UserID       int64     `db:"user_id"`
	Username     string    `db:"username"`
	Avatar       *string   `db:"avatar"`
	Role         string    `db:"role"`
	ReplyToID    *int64    `db:"reply_to_id"`
	ReplyContent *string   `db:"reply_content"`
	ReplyUsername *string  `db:"reply_username"`
}

type UserInsert struct {
	Username     string
	Email        string
	PasswordHash string
	FirstName    string
	LastName     string
	Role         string
	Status       string
}

type UserUpdate struct {
	ID       int64
	Username string
	Status   string
}

// Utilitaires
func int64SliceToStringSlice(slice []int64) []string {
	result := make([]string, len(slice))
	for i, v := range slice {
		result[i] = fmt.Sprintf("%d", v)
	}
	return result
}
