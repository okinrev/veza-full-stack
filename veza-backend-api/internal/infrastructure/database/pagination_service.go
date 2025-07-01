package database

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

// PaginationType type de pagination
type PaginationType string

const (
	OffsetPagination PaginationType = "offset" // LIMIT/OFFSET traditionnel
	CursorPagination PaginationType = "cursor" // Pagination par curseur (plus performant)
	KeysetPagination PaginationType = "keyset" // Pagination par keyset (très performant)
)

// PaginationConfig configuration pour la pagination
type PaginationConfig struct {
	DefaultPageSize    int            `json:"default_page_size"`   // Taille page par défaut
	MaxPageSize        int            `json:"max_page_size"`       // Taille page maximum
	DefaultType        PaginationType `json:"default_type"`        // Type par défaut
	EnableOptimization bool           `json:"enable_optimization"` // Optimisations automatiques
	CacheResults       bool           `json:"cache_results"`       // Cache des résultats
	CacheTTL           time.Duration  `json:"cache_ttl"`           // TTL du cache
}

// DefaultPaginationConfig configuration par défaut optimisée
func DefaultPaginationConfig() *PaginationConfig {
	return &PaginationConfig{
		DefaultPageSize:    20,
		MaxPageSize:        100,
		DefaultType:        CursorPagination,
		EnableOptimization: true,
		CacheResults:       true,
		CacheTTL:           5 * time.Minute,
	}
}

// PaginationRequest paramètres de pagination
type PaginationRequest struct {
	// Pagination par offset
	Page     int `json:"page,omitempty"`      // Numéro de page (1-based)
	PageSize int `json:"page_size,omitempty"` // Taille de page

	// Pagination par curseur
	Cursor    string `json:"cursor,omitempty"`    // Curseur pour page suivante
	Direction string `json:"direction,omitempty"` // "next" ou "prev"

	// Tri
	SortBy    string `json:"sort_by,omitempty"`    // Colonne de tri
	SortOrder string `json:"sort_order,omitempty"` // "asc" ou "desc"

	// Filtres
	Filters map[string]interface{} `json:"filters,omitempty"`

	// Type de pagination forcé
	Type PaginationType `json:"type,omitempty"`
}

// PaginationResponse réponse paginée
type PaginationResponse struct {
	Data        interface{}     `json:"data"`        // Données paginées
	Pagination  PaginationMeta  `json:"pagination"`  // Métadonnées pagination
	Performance PerformanceMeta `json:"performance"` // Métriques performance
}

// PaginationMeta métadonnées de pagination
type PaginationMeta struct {
	CurrentPage int    `json:"current_page"`          // Page actuelle
	PageSize    int    `json:"page_size"`             // Taille de page
	TotalCount  int64  `json:"total_count,omitempty"` // Total items (si disponible)
	TotalPages  int    `json:"total_pages,omitempty"` // Total pages (si disponible)
	HasNext     bool   `json:"has_next"`              // Y a-t-il une page suivante
	HasPrev     bool   `json:"has_prev"`              // Y a-t-il une page précédente
	NextCursor  string `json:"next_cursor,omitempty"` // Curseur page suivante
	PrevCursor  string `json:"prev_cursor,omitempty"` // Curseur page précédente
	Type        string `json:"type"`                  // Type de pagination utilisé
}

// PerformanceMeta métriques de performance
type PerformanceMeta struct {
	QueryTime    time.Duration `json:"query_time"`    // Temps requête
	TotalTime    time.Duration `json:"total_time"`    // Temps total
	CacheHit     bool          `json:"cache_hit"`     // Cache hit
	IndexUsed    bool          `json:"index_used"`    // Index utilisé
	RowsExamined int64         `json:"rows_examined"` // Lignes examinées
}

// PaginationService service de pagination intelligente
type PaginationService struct {
	db      *sqlx.DB
	logger  *zap.Logger
	config  *PaginationConfig
	metrics *PaginationMetrics
}

// PaginationMetrics métriques pour la pagination
type PaginationMetrics struct {
	paginationRequests *prometheus.CounterVec
	queryDuration      *prometheus.HistogramVec
	cacheHits          *prometheus.CounterVec
	pageSize           *prometheus.HistogramVec
	optimizations      *prometheus.CounterVec
}

// NewPaginationMetrics crée les métriques Prometheus
func NewPaginationMetrics() *PaginationMetrics {
	return &PaginationMetrics{
		paginationRequests: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "db_pagination_requests_total",
			Help: "Nombre de requêtes de pagination",
		}, []string{"type", "table", "optimization"}),
		queryDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "db_pagination_query_duration_seconds",
			Help:    "Durée des requêtes de pagination",
			Buckets: []float64{0.001, 0.01, 0.1, 0.5, 1.0, 2.0, 5.0},
		}, []string{"type", "table"}),
		cacheHits: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "db_pagination_cache_hits_total",
			Help: "Cache hits pour pagination",
		}, []string{"table", "cache_type"}),
		pageSize: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "db_pagination_page_size",
			Help:    "Taille des pages demandées",
			Buckets: []float64{10, 20, 50, 100, 200, 500, 1000},
		}, []string{"table"}),
		optimizations: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "db_pagination_optimizations_total",
			Help: "Optimisations appliquées automatiquement",
		}, []string{"optimization_type", "table"}),
	}
}

// NewPaginationService crée un nouveau service de pagination
func NewPaginationService(db *sqlx.DB, logger *zap.Logger) *PaginationService {
	return &PaginationService{
		db:      db,
		logger:  logger,
		config:  DefaultPaginationConfig(),
		metrics: NewPaginationMetrics(),
	}
}

// PaginateUsers pagination optimisée pour les utilisateurs
func (s *PaginationService) PaginateUsers(ctx context.Context, req PaginationRequest) (*PaginationResponse, error) {
	startTime := time.Now()

	// Normaliser la requête
	s.normalizeRequest(&req)

	// Choisir le type optimal de pagination
	paginationType := s.choosePaginationType(req, "users")

	var response *PaginationResponse
	var err error

	switch paginationType {
	case CursorPagination:
		response, err = s.paginateUsersCursor(ctx, req)
	case KeysetPagination:
		response, err = s.paginateUsersKeyset(ctx, req)
	default:
		response, err = s.paginateUsersOffset(ctx, req)
	}

	if err != nil {
		return nil, err
	}

	// Ajouter métriques performance
	totalTime := time.Since(startTime)
	response.Performance.TotalTime = totalTime

	// Métriques Prometheus
	s.metrics.paginationRequests.WithLabelValues(string(paginationType), "users", "auto").Inc()
	s.metrics.queryDuration.WithLabelValues(string(paginationType), "users").Observe(response.Performance.QueryTime.Seconds())
	s.metrics.pageSize.WithLabelValues("users").Observe(float64(req.PageSize))

	return response, nil
}

// paginateUsersCursor pagination par curseur (très performant)
func (s *PaginationService) paginateUsersCursor(ctx context.Context, req PaginationRequest) (*PaginationResponse, error) {
	queryStart := time.Now()

	// Construire la requête avec curseur
	var whereClause string
	var args []interface{}
	argIndex := 1

	// Filtres de base
	conditions := []string{"is_active = true"}

	// Condition de curseur
	if req.Cursor != "" {
		cursorID, err := s.decodeCursor(req.Cursor)
		if err != nil {
			return nil, fmt.Errorf("curseur invalide: %w", err)
		}

		if req.Direction == "prev" {
			conditions = append(conditions, fmt.Sprintf("id > $%d", argIndex))
		} else {
			conditions = append(conditions, fmt.Sprintf("id < $%d", argIndex))
		}
		args = append(args, cursorID)
		argIndex++
	}

	// Filtres additionnels
	for key, value := range req.Filters {
		switch key {
		case "role":
			conditions = append(conditions, fmt.Sprintf("role = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "status":
			conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
			args = append(args, value)
			argIndex++
		}
	}

	whereClause = strings.Join(conditions, " AND ")

	// Ordre de tri optimisé (utilise l'index sur id)
	orderClause := "ORDER BY id DESC"
	if req.Direction == "prev" {
		orderClause = "ORDER BY id ASC"
	}

	// Requête principale avec LIMIT+1 pour détecter page suivante
	finalQuery := fmt.Sprintf(`
		SELECT id, username, email, first_name, last_name, role, status, 
		       is_active, is_verified, created_at, updated_at
		FROM users 
		WHERE %s 
		%s 
		LIMIT $%d
	`, whereClause, orderClause, argIndex)

	args = append(args, req.PageSize+1) // +1 pour détecter hasNext

	rows, err := s.db.QueryContext(ctx, finalQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("erreur requête cursor pagination: %w", err)
	}
	defer rows.Close()

	var users []map[string]interface{}
	var lastID int64

	for rows.Next() {
		if len(users) >= req.PageSize {
			break // On a atteint la limite
		}

		var user map[string]interface{}
		// Scanner les colonnes (simplifié pour l'exemple)
		var id int64
		var username, email string
		// ... autres colonnes

		if err := rows.Scan(&id, &username, &email /* autres colonnes */); err != nil {
			continue
		}

		user = map[string]interface{}{
			"id":       id,
			"username": username,
			"email":    email,
			// ... autres champs
		}

		users = append(users, user)
		lastID = id
	}

	queryTime := time.Since(queryStart)

	// Déterminer s'il y a page suivante/précédente
	hasNext := len(users) == req.PageSize && rows.Next()
	hasPrev := req.Cursor != ""

	// Générer curseurs
	var nextCursor, prevCursor string
	if hasNext && len(users) > 0 {
		nextCursor = s.encodeCursor(lastID)
	}
	if hasPrev && len(users) > 0 {
		prevCursor = s.encodeCursor(lastID)
	}

	response := &PaginationResponse{
		Data: users,
		Pagination: PaginationMeta{
			PageSize:   req.PageSize,
			HasNext:    hasNext,
			HasPrev:    hasPrev,
			NextCursor: nextCursor,
			PrevCursor: prevCursor,
			Type:       string(CursorPagination),
		},
		Performance: PerformanceMeta{
			QueryTime: queryTime,
			IndexUsed: true, // Utilise l'index sur id
			CacheHit:  false,
		},
	}

	return response, nil
}

// paginateUsersKeyset pagination par keyset (pour tri complexe)
func (s *PaginationService) paginateUsersKeyset(ctx context.Context, req PaginationRequest) (*PaginationResponse, error) {
	// Implémentation pagination keyset pour tri par created_at + id
	queryStart := time.Now()

	// Cette méthode est optimale pour tri par colonnes indexées
	// Exemple: ORDER BY created_at DESC, id DESC

	var whereClause string
	var args []interface{}

	// Condition keyset si curseur fourni
	if req.Cursor != "" {
		cursorData, err := s.decodeKeysetCursor(req.Cursor)
		if err != nil {
			return nil, fmt.Errorf("curseur keyset invalide: %w", err)
		}

		// WHERE (created_at, id) < (cursor_date, cursor_id)
		whereClause = "WHERE (created_at, id) < ($1, $2)"
		args = append(args, cursorData.CreatedAt, cursorData.ID)
	} else {
		whereClause = "WHERE is_active = true"
	}

	// Construction de la requête keyset
	finalQuery := fmt.Sprintf(`
		SELECT id, username, email, first_name, last_name, role, status,
		       is_active, is_verified, created_at, updated_at
		FROM users 
		%s
		ORDER BY created_at DESC, id DESC
		LIMIT $%d
	`, whereClause, len(args)+1)

	args = append(args, req.PageSize+1)

	// Exécuter la requête...
	_ = finalQuery // Utiliser la requête construite
	queryTime := time.Since(queryStart)

	// Construire la réponse similaire à cursor pagination
	response := &PaginationResponse{
		Performance: PerformanceMeta{
			QueryTime: queryTime,
			IndexUsed: true,
		},
	}

	return response, nil
}

// paginateUsersOffset pagination offset classique (fallback)
func (s *PaginationService) paginateUsersOffset(ctx context.Context, req PaginationRequest) (*PaginationResponse, error) {
	queryStart := time.Now()

	offset := (req.Page - 1) * req.PageSize

	// Si offset très élevé, recommander cursor pagination
	if offset > 10000 {
		s.metrics.optimizations.WithLabelValues("cursor_recommendation", "users").Inc()
		s.logger.Warn("Offset élevé détecté, cursor pagination recommandée",
			zap.Int("offset", offset),
		)
	}

	// Requête avec OFFSET/LIMIT
	query := `
		SELECT id, username, email, first_name, last_name, role, status,
		       is_active, is_verified, created_at, updated_at
		FROM users 
		WHERE is_active = true
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := s.db.QueryContext(ctx, query, req.PageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("erreur pagination offset: %w", err)
	}
	defer rows.Close()

	var users []map[string]interface{}
	// Scanner les résultats...

	queryTime := time.Since(queryStart)

	// Compter le total (coûteux mais nécessaire pour OFFSET)
	var totalCount int64
	countQuery := "SELECT COUNT(*) FROM users WHERE is_active = true"
	if err := s.db.GetContext(ctx, &totalCount, countQuery); err != nil {
		s.logger.Warn("Erreur calcul total count", zap.Error(err))
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(req.PageSize)))

	response := &PaginationResponse{
		Data: users,
		Pagination: PaginationMeta{
			CurrentPage: req.Page,
			PageSize:    req.PageSize,
			TotalCount:  totalCount,
			TotalPages:  totalPages,
			HasNext:     req.Page < totalPages,
			HasPrev:     req.Page > 1,
			Type:        string(OffsetPagination),
		},
		Performance: PerformanceMeta{
			QueryTime:    queryTime,
			IndexUsed:    true,
			RowsExamined: totalCount, // OFFSET examine toutes les lignes précédentes
		},
	}

	return response, nil
}

// Utilitaires pour gestion curseurs et configuration

// normalizeRequest normalise les paramètres de pagination
func (s *PaginationService) normalizeRequest(req *PaginationRequest) {
	if req.PageSize <= 0 {
		req.PageSize = s.config.DefaultPageSize
	}
	if req.PageSize > s.config.MaxPageSize {
		req.PageSize = s.config.MaxPageSize
		s.metrics.optimizations.WithLabelValues("page_size_capped", "").Inc()
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.SortOrder == "" {
		req.SortOrder = "desc"
	}
}

// choosePaginationType choisit le type optimal de pagination
func (s *PaginationService) choosePaginationType(req PaginationRequest, table string) PaginationType {
	if req.Type != "" {
		return req.Type
	}

	// Si curseur fourni, utiliser cursor pagination
	if req.Cursor != "" {
		return CursorPagination
	}

	// Si tri complexe, utiliser keyset
	if req.SortBy != "" && req.SortBy != "id" {
		s.metrics.optimizations.WithLabelValues("keyset_optimization", table).Inc()
		return KeysetPagination
	}

	// Si page élevée, recommander cursor
	if req.Page > 100 {
		s.metrics.optimizations.WithLabelValues("cursor_auto", table).Inc()
		return CursorPagination
	}

	return s.config.DefaultType
}

// Encodage/décodage curseurs (simplifiés)
func (s *PaginationService) encodeCursor(id int64) string {
	return fmt.Sprintf("%d", id) // Simplifié - utiliser base64 en production
}

func (s *PaginationService) decodeCursor(cursor string) (int64, error) {
	var id int64
	_, err := fmt.Sscanf(cursor, "%d", &id)
	return id, err
}

type KeysetCursor struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

func (s *PaginationService) decodeKeysetCursor(cursor string) (*KeysetCursor, error) {
	// Implémentation simplifiée - décoder JSON en production
	return &KeysetCursor{}, nil
}
