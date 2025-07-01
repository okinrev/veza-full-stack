package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

// ConnectionPoolConfig configuration pour le pool de connexions
type ConnectionPoolConfig struct {
	// Pool principal
	MaxOpenConns    int           `json:"max_open_conns"`     // Connexions max ouvertes
	MaxIdleConns    int           `json:"max_idle_conns"`     // Connexions idle max
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"`  // Durée de vie max d'une connexion
	ConnMaxIdleTime time.Duration `json:"conn_max_idle_time"` // Temps idle max

	// Timeouts
	QueryTimeout   time.Duration `json:"query_timeout"`   // Timeout pour les requêtes
	ConnectTimeout time.Duration `json:"connect_timeout"` // Timeout pour la connexion
	PingTimeout    time.Duration `json:"ping_timeout"`    // Timeout pour le ping

	// Read replicas
	ReadReplicaEnabled bool     `json:"read_replica_enabled"`
	ReadReplicaURLs    []string `json:"read_replica_urls"`
	ReadReplicaWeight  int      `json:"read_replica_weight"` // % de requêtes vers read replicas

	// Health checks
	HealthCheckInterval time.Duration `json:"health_check_interval"`
	MaxRetries          int           `json:"max_retries"`
	RetryBackoff        time.Duration `json:"retry_backoff"`
}

// DefaultConnectionPoolConfig configuration par défaut optimisée pour 100k+ users
func DefaultConnectionPoolConfig() *ConnectionPoolConfig {
	return &ConnectionPoolConfig{
		// Pool principal - optimisé pour haute charge
		MaxOpenConns:    100,              // 100 connexions max (vs 25 avant)
		MaxIdleConns:    50,               // 50 connexions idle
		ConnMaxLifetime: 30 * time.Minute, // Renouveler connexions toutes les 30min
		ConnMaxIdleTime: 5 * time.Minute,  // Fermer connexions idle après 5min

		// Timeouts agressifs
		QueryTimeout:   30 * time.Second, // Timeout requêtes
		ConnectTimeout: 10 * time.Second, // Timeout connexion
		PingTimeout:    5 * time.Second,  // Timeout ping

		// Read replicas pour analytics
		ReadReplicaEnabled: true,
		ReadReplicaWeight:  70, // 70% des SELECT vers read replicas

		// Health monitoring
		HealthCheckInterval: 30 * time.Second,
		MaxRetries:          3,
		RetryBackoff:        time.Second,
	}
}

// ConnectionPoolService gère les pools de connexions optimisés
type ConnectionPoolService struct {
	config *ConnectionPoolConfig
	logger *zap.Logger

	// Pools de connexions
	primaryDB       *sqlx.DB
	readReplicas    []*sqlx.DB
	replicasHealthy []bool // Status de santé des read replicas

	// Context pour arrêt propre
	ctx    context.Context
	cancel context.CancelFunc
}

// NewConnectionPoolService crée un nouveau service de pool de connexions
func NewConnectionPoolService(config *ConnectionPoolConfig, logger *zap.Logger) *ConnectionPoolService {
	if config == nil {
		config = DefaultConnectionPoolConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &ConnectionPoolService{
		config: config,
		logger: logger,
		ctx:    ctx,
		cancel: cancel,
	}
}

// GetPrimaryDB retourne la connexion à la DB principale
func (s *ConnectionPoolService) GetPrimaryDB() *sqlx.DB {
	return s.primaryDB
}

// GetReadDB retourne une connexion optimale pour lecture
func (s *ConnectionPoolService) GetReadDB() *sqlx.DB {
	// Si read replicas désactivés ou non disponibles, utiliser primary
	if !s.config.ReadReplicaEnabled || len(s.readReplicas) == 0 {
		return s.primaryDB
	}

	// Choisir un read replica sain aléatoirement
	for i, replica := range s.readReplicas {
		if len(s.replicasHealthy) > i && s.replicasHealthy[i] {
			return replica
		}
	}

	// Fallback sur primary si aucun replica sain
	s.logger.Warn("Aucun read replica sain, fallback sur primary")
	return s.primaryDB
}

// Initialize initialise les pools de connexions
func (s *ConnectionPoolService) Initialize(primaryURL string, replicaURLs ...string) error {
	// Initialiser le pool principal
	db, err := sqlx.Connect("postgres", primaryURL)
	if err != nil {
		return fmt.Errorf("connexion à la DB principale échouée: %w", err)
	}

	// Configuration du pool
	db.SetMaxOpenConns(s.config.MaxOpenConns)
	db.SetMaxIdleConns(s.config.MaxIdleConns)
	db.SetConnMaxLifetime(s.config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(s.config.ConnMaxIdleTime)

	s.primaryDB = db
	s.logger.Info("Pool principal initialisé")

	// Initialiser les read replicas si fournis
	if len(replicaURLs) > 0 {
		s.readReplicas = make([]*sqlx.DB, 0, len(replicaURLs))
		s.replicasHealthy = make([]bool, len(replicaURLs))

		for i, replicaURL := range replicaURLs {
			replica, err := sqlx.Connect("postgres", replicaURL)
			if err != nil {
				s.logger.Warn("Connexion read replica échouée", zap.Error(err))
				s.replicasHealthy[i] = false
				continue
			}

			replica.SetMaxOpenConns(s.config.MaxOpenConns / 2)
			replica.SetMaxIdleConns(s.config.MaxIdleConns / 2)
			replica.SetConnMaxLifetime(s.config.ConnMaxLifetime)
			replica.SetConnMaxIdleTime(s.config.ConnMaxIdleTime)

			s.readReplicas = append(s.readReplicas, replica)
			s.replicasHealthy[i] = true

			s.logger.Info("Read replica initialisé", zap.Int("index", i))
		}
	}

	return nil
}

// GetStats retourne les statistiques du pool
func (s *ConnectionPoolService) GetStats() map[string]interface{} {
	stats := make(map[string]interface{})

	if s.primaryDB != nil {
		dbStats := s.primaryDB.Stats()
		stats["primary"] = map[string]interface{}{
			"open_connections": dbStats.OpenConnections,
			"in_use":           dbStats.InUse,
			"idle":             dbStats.Idle,
		}
	}

	replicaStats := make([]map[string]interface{}, len(s.readReplicas))
	for i, replica := range s.readReplicas {
		if replica != nil {
			dbStats := replica.Stats()
			replicaStats[i] = map[string]interface{}{
				"open_connections": dbStats.OpenConnections,
				"in_use":           dbStats.InUse,
				"idle":             dbStats.Idle,
				"healthy":          len(s.replicasHealthy) > i && s.replicasHealthy[i],
			}
		}
	}
	stats["replicas"] = replicaStats

	return stats
}

// Close ferme toutes les connexions proprement
func (s *ConnectionPoolService) Close() error {
	s.cancel()

	var errors []error

	// Fermer primary
	if s.primaryDB != nil {
		if err := s.primaryDB.Close(); err != nil {
			errors = append(errors, fmt.Errorf("erreur fermeture primary: %w", err))
		}
	}

	// Fermer read replicas
	for i, replica := range s.readReplicas {
		if replica != nil {
			if err := replica.Close(); err != nil {
				errors = append(errors, fmt.Errorf("erreur fermeture replica %d: %w", i, err))
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("erreurs fermeture pool: %v", errors)
	}

	s.logger.Info("ConnectionPoolService fermé proprement")
	return nil
}
