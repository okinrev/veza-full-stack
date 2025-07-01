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

// IndexType type d'index PostgreSQL
type IndexType string

const (
	BTReeIndex   IndexType = "btree"    // Index B-Tree standard
	HashIndex    IndexType = "hash"     // Index hash pour égalités
	GinIndex     IndexType = "gin"      // Index GIN pour full-text search
	GistIndex    IndexType = "gist"     // Index GIST pour données géographiques
	BrinIndex    IndexType = "brin"     // Index BRIN pour très grandes tables
	PartialIndex IndexType = "partial"  // Index partiel avec condition WHERE
)

// IndexDefinition définition d'un index optimisé
type IndexDefinition struct {
	Name        string    `json:"name"`         // Nom de l'index
	Table       string    `json:"table"`        // Table cible
	Columns     []string  `json:"columns"`      // Colonnes indexées
	Type        IndexType `json:"type"`         // Type d'index
	Unique      bool      `json:"unique"`       // Index unique
	Partial     string    `json:"partial"`      // Condition WHERE pour index partiel
	Include     []string  `json:"include"`      // Colonnes INCLUDE (PostgreSQL 11+)
	Concurrent  bool      `json:"concurrent"`   // Création concurrente
	Priority    int       `json:"priority"`     // Priorité (1=critique, 5=faible)
	Description string    `json:"description"`  // Description de l'optimisation
}

// CriticalIndexes index critiques pour performance 100k+ users
var CriticalIndexes = []IndexDefinition{
	// Users - Authentification haute fréquence
	{
		Name:        "idx_users_email_active",
		Table:       "users",
		Columns:     []string{"email"},
		Type:        BTReeIndex,
		Unique:      true,
		Partial:     "is_active = true",  // Seulement users actifs
		Concurrent:  true,
		Priority:    1,
		Description: "Login par email - critique pour auth",
	},
	{
		Name:        "idx_users_username_active",
		Table:       "users",
		Columns:     []string{"username"},
		Type:        BTReeIndex,
		Unique:      true,
		Partial:     "is_active = true",
		Concurrent:  true,
		Priority:    1,
		Description: "Login par username - critique pour auth",
	},
	{
		Name:        "idx_users_role_status",
		Table:       "users",
		Columns:     []string{"role", "status"},
		Type:        BTReeIndex,
		Include:     []string{"id", "username", "email"},
		Concurrent:  true,
		Priority:    2,
		Description: "Filtrage par rôle et statut avec données INCLUDE",
	},
	
	// Sessions - Gestion des connexions
	{
		Name:        "idx_user_sessions_token_active",
		Table:       "user_sessions",
		Columns:     []string{"refresh_token"},
		Type:        BTReeIndex,
		Unique:      true,
		Partial:     "expires_at > CURRENT_TIMESTAMP",  // Seulement sessions non expirées
		Concurrent:  true,
		Priority:    1,
		Description: "Validation tokens - critique pour sessions",
	},
	{
		Name:        "idx_user_sessions_user_expires",
		Table:       "user_sessions",
		Columns:     []string{"user_id", "expires_at"},
		Type:        BTReeIndex,
		Concurrent:  true,
		Priority:    1,
		Description: "Sessions par utilisateur avec expiration",
	},
	
	// Chat - Messages haute fréquence
	{
		Name:        "idx_chat_messages_room_created",
		Table:       "chat_messages",
		Columns:     []string{"room_id", "created_at"},
		Type:        BTReeIndex,
		Include:     []string{"id", "user_id", "content", "message_type"},
		Concurrent:  true,
		Priority:    1,
		Description: "Messages par room ordonnés par date - critique pour chat",
	},
	{
		Name:        "idx_chat_messages_user_created",
		Table:       "chat_messages",
		Columns:     []string{"user_id", "created_at"},
		Type:        BTReeIndex,
		Concurrent:  true,
		Priority:    2,
		Description: "Messages par utilisateur pour historique",
	},
	{
		Name:        "idx_chat_messages_content_gin",
		Table:       "chat_messages",
		Columns:     []string{"content"},
		Type:        GinIndex,
		Concurrent:  true,
		Priority:    3,
		Description: "Recherche full-text dans messages",
	},
	
	// Chat Rooms - Gestion des salles
	{
		Name:        "idx_chat_rooms_type_active",
		Table:       "chat_rooms",
		Columns:     []string{"room_type", "is_active"},
		Type:        BTReeIndex,
		Include:     []string{"id", "name", "created_at"},
		Concurrent:  true,
		Priority:    2,
		Description: "Filtrage rooms par type et statut actif",
	},
	
	// Rate Limiting - Anti-spam
	{
		Name:        "idx_rate_limits_key_window",
		Table:       "rate_limits",
		Columns:     []string{"limit_key", "window_start"},
		Type:        BTReeIndex,
		Unique:      true,
		Concurrent:  true,
		Priority:    1,
		Description: "Rate limiting par clé et fenêtre temporelle",
	},
	
	// Audit Logs - Monitoring
	{
		Name:        "idx_audit_logs_created_level",
		Table:       "audit_logs",
		Columns:     []string{"created_at", "level"},
		Type:        BTReeIndex,
		Concurrent:  true,
		Priority:    3,
		Description: "Logs par date et niveau pour monitoring",
	},
	{
		Name:        "idx_audit_logs_user_action",
		Table:       "audit_logs",
		Columns:     []string{"user_id", "action"},
		Type:        BTReeIndex,
		Partial:     "user_id IS NOT NULL",  // Exclure logs système
		Concurrent:  true,
		Priority:    3,
		Description: "Actions par utilisateur pour audit",
	},
	
	// Files/Uploads - Gestion fichiers
	{
		Name:        "idx_files_user_created",
		Table:       "files",
		Columns:     []string{"uploaded_by", "created_at"},
		Type:        BTReeIndex,
		Include:     []string{"id", "filename", "file_size", "mime_type"},
		Concurrent:  true,
		Priority:    2,
		Description: "Fichiers par utilisateur avec métadonnées",
	},
	{
		Name:        "idx_files_status_type",
		Table:       "files",
		Columns:     []string{"status", "mime_type"},
		Type:        BTReeIndex,
		Concurrent:  true,
		Priority:    3,
		Description: "Filtrage fichiers par statut et type MIME",
	},
}

// IndexOptimizationService gère l'optimisation des index PostgreSQL
type IndexOptimizationService struct {
	db     *sqlx.DB
	logger *zap.Logger
	
	// Métriques
	metrics *IndexMetrics
	
	// Configuration
	config *IndexConfig
}

// IndexConfig configuration pour l'optimisation des index
type IndexConfig struct {
	EnableConcurrentCreation bool          `json:"enable_concurrent_creation"`
	MaxConcurrentOperations  int           `json:"max_concurrent_operations"`
	AnalyzeAfterCreation     bool          `json:"analyze_after_creation"`
	ReindexThreshold         time.Duration `json:"reindex_threshold"`         // Seuil pour reindexer
	MaintenanceWindow        time.Duration `json:"maintenance_window"`        // Fenêtre de maintenance
	MonitoringInterval       time.Duration `json:"monitoring_interval"`       // Intervalle monitoring
}

// DefaultIndexConfig configuration par défaut
func DefaultIndexConfig() *IndexConfig {
	return &IndexConfig{
		EnableConcurrentCreation: true,
		MaxConcurrentOperations:  3,
		AnalyzeAfterCreation:     true,
		ReindexThreshold:         7 * 24 * time.Hour,  // Reindex hebdomadaire
		MaintenanceWindow:        2 * time.Hour,       // 2h de maintenance
		MonitoringInterval:       1 * time.Hour,       // Monitoring horaire
	}
}

// IndexMetrics métriques pour les index
type IndexMetrics struct {
	indexCreationDuration *prometheus.HistogramVec
	indexSize             *prometheus.GaugeVec
	indexUsage            *prometheus.CounterVec
	reindexOperations     *prometheus.CounterVec
	indexHealth           *prometheus.GaugeVec
}

// NewIndexMetrics crée les métriques Prometheus
func NewIndexMetrics() *IndexMetrics {
	return &IndexMetrics{
		indexCreationDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name: "db_index_creation_duration_seconds",
			Help: "Durée de création des index",
		}, []string{"index_name", "table_name", "index_type"}),
		indexSize: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "db_index_size_bytes",
			Help: "Taille des index en bytes",
		}, []string{"index_name", "table_name"}),
		indexUsage: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "db_index_usage_total",
			Help: "Nombre d'utilisations des index",
		}, []string{"index_name", "scan_type"}),
		reindexOperations: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "db_reindex_operations_total",
			Help: "Nombre d'opérations de reindex",
		}, []string{"index_name", "operation_type"}),
		indexHealth: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "db_index_health_score",
			Help: "Score de santé des index (0-1)",
		}, []string{"index_name", "table_name"}),
	}
}

// NewIndexOptimizationService crée un nouveau service d'optimisation d'index
func NewIndexOptimizationService(db *sqlx.DB, logger *zap.Logger) *IndexOptimizationService {
	return &IndexOptimizationService{
		db:      db,
		logger:  logger,
		metrics: NewIndexMetrics(),
		config:  DefaultIndexConfig(),
	}
}

// CreateCriticalIndexes crée tous les index critiques pour la performance
func (s *IndexOptimizationService) CreateCriticalIndexes(ctx context.Context) error {
	s.logger.Info("Début création des index critiques", 
		zap.Int("total_indexes", len(CriticalIndexes)),
	)
	
	// Grouper par priorité
	priorityGroups := s.groupIndexesByPriority()
	
	// Créer par ordre de priorité
	for priority := 1; priority <= 5; priority++ {
		indexes := priorityGroups[priority]
		if len(indexes) == 0 {
			continue
		}
		
		s.logger.Info("Création index priorité", 
			zap.Int("priority", priority),
			zap.Int("count", len(indexes)),
		)
		
		if err := s.createIndexGroup(ctx, indexes); err != nil {
			return fmt.Errorf("erreur création index priorité %d: %w", priority, err)
		}
	}
	
	s.logger.Info("Tous les index critiques créés avec succès")
	return nil
}

// groupIndexesByPriority groupe les index par priorité
func (s *IndexOptimizationService) groupIndexesByPriority() map[int][]IndexDefinition {
	groups := make(map[int][]IndexDefinition)
	
	for _, idx := range CriticalIndexes {
		priority := idx.Priority
		if priority == 0 {
			priority = 3 // Priorité par défaut
		}
		groups[priority] = append(groups[priority], idx)
	}
	
	return groups
}

// createIndexGroup crée un groupe d'index en parallèle
func (s *IndexOptimizationService) createIndexGroup(ctx context.Context, indexes []IndexDefinition) error {
	// Canal pour limiter la concurrence
	semaphore := make(chan struct{}, s.config.MaxConcurrentOperations)
	errChan := make(chan error, len(indexes))
	
	// Lancer les créations en parallèle
	for _, idx := range indexes {
		go func(index IndexDefinition) {
			semaphore <- struct{}{} // Acquérir
			defer func() { <-semaphore }() // Libérer
			
			if err := s.createSingleIndex(ctx, index); err != nil {
				errChan <- fmt.Errorf("erreur index %s: %w", index.Name, err)
				return
			}
			errChan <- nil
		}(idx)
	}
	
	// Attendre toutes les créations
	var errors []error
	for i := 0; i < len(indexes); i++ {
		if err := <-errChan; err != nil {
			errors = append(errors, err)
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("erreurs création index: %v", errors)
	}
	
	return nil
}

// createSingleIndex crée un index unique
func (s *IndexOptimizationService) createSingleIndex(ctx context.Context, idx IndexDefinition) error {
	start := time.Now()
	
	// Vérifier si l'index existe déjà
	exists, err := s.indexExists(ctx, idx.Name)
	if err != nil {
		return fmt.Errorf("erreur vérification existence: %w", err)
	}
	if exists {
		s.logger.Info("Index déjà existant", zap.String("index", idx.Name))
		return nil
	}
	
	// Générer la requête SQL
	query := s.buildCreateIndexQuery(idx)
	
	s.logger.Info("Création index", 
		zap.String("index", idx.Name),
		zap.String("table", idx.Table),
		zap.String("type", string(idx.Type)),
		zap.Strings("columns", idx.Columns),
	)
	
	// Exécuter la création
	if _, err := s.db.ExecContext(ctx, query); err != nil {
		s.logger.Error("Erreur création index", 
			zap.String("index", idx.Name),
			zap.Error(err),
			zap.String("query", query),
		)
		return fmt.Errorf("création index échouée: %w", err)
	}
	
	duration := time.Since(start)
	
	// Analyser la table après création si configuré
	if s.config.AnalyzeAfterCreation {
		if err := s.analyzeTable(ctx, idx.Table); err != nil {
			s.logger.Warn("Erreur ANALYZE après création index", 
				zap.String("table", idx.Table),
				zap.Error(err),
			)
		}
	}
	
	// Métriques
	s.metrics.indexCreationDuration.WithLabelValues(
		idx.Name, idx.Table, string(idx.Type),
	).Observe(duration.Seconds())
	
	s.logger.Info("Index créé avec succès", 
		zap.String("index", idx.Name),
		zap.Duration("duration", duration),
	)
	
	return nil
}

// buildCreateIndexQuery construit la requête SQL de création d'index
func (s *IndexOptimizationService) buildCreateIndexQuery(idx IndexDefinition) string {
	var query strings.Builder
	
	// CREATE [UNIQUE] INDEX [CONCURRENTLY] nom
	query.WriteString("CREATE ")
	if idx.Unique {
		query.WriteString("UNIQUE ")
	}
	query.WriteString("INDEX ")
	if idx.Concurrent && s.config.EnableConcurrentCreation {
		query.WriteString("CONCURRENTLY ")
	}
	query.WriteString(idx.Name)
	
	// ON table
	query.WriteString(" ON ")
	query.WriteString(idx.Table)
	
	// USING type
	if idx.Type != BTReeIndex {
		query.WriteString(" USING ")
		query.WriteString(string(idx.Type))
	}
	
	// (colonnes)
	query.WriteString(" (")
	query.WriteString(strings.Join(idx.Columns, ", "))
	query.WriteString(")")
	
	// INCLUDE (colonnes) pour PostgreSQL 11+
	if len(idx.Include) > 0 {
		query.WriteString(" INCLUDE (")
		query.WriteString(strings.Join(idx.Include, ", "))
		query.WriteString(")")
	}
	
	// WHERE condition pour index partiel
	if idx.Partial != "" {
		query.WriteString(" WHERE ")
		query.WriteString(idx.Partial)
	}
	
	return query.String()
}

// indexExists vérifie si un index existe
func (s *IndexOptimizationService) indexExists(ctx context.Context, indexName string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM pg_indexes 
			WHERE indexname = $1
		)
	`
	var exists bool
	err := s.db.GetContext(ctx, &exists, query, indexName)
	return exists, err
}

// analyzeTable lance ANALYZE sur une table
func (s *IndexOptimizationService) analyzeTable(ctx context.Context, tableName string) error {
	query := fmt.Sprintf("ANALYZE %s", tableName)
	_, err := s.db.ExecContext(ctx, query)
	return err
}

// GetIndexStats retourne les statistiques des index
func (s *IndexOptimizationService) GetIndexStats(ctx context.Context) (map[string]interface{}, error) {
	query := `
		SELECT 
			schemaname,
			tablename,
			indexname,
			pg_size_pretty(pg_relation_size(indexrelid)) as size,
			idx_scan,
			idx_tup_read,
			idx_tup_fetch
		FROM pg_stat_user_indexes 
		ORDER BY pg_relation_size(indexrelid) DESC
	`
	
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var stats []map[string]interface{}
	for rows.Next() {
		var schema, table, index, size string
		var scan, tupRead, tupFetch int64
		
		if err := rows.Scan(&schema, &table, &index, &size, &scan, &tupRead, &tupFetch); err != nil {
			continue
		}
		
		stats = append(stats, map[string]interface{}{
			"schema":     schema,
			"table":      table,
			"index":      index,
			"size":       size,
			"scans":      scan,
			"tup_read":   tupRead,
			"tup_fetch":  tupFetch,
		})
	}
	
	return map[string]interface{}{
		"indexes": stats,
	}, nil
}
