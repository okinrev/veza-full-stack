package postgres

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/okinrev/veza-web-app/internal/infrastructure/config"
)

// NewConnection crée une nouvelle connexion PostgreSQL avec pool optimisé
func NewConnection(cfg config.DatabaseConfig) (*sql.DB, error) {
	// Construire l'URL de connexion PostgreSQL
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("erreur ouverture base de données: %w", err)
	}

	// Configuration du pool de connexions
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	// Test de connexion
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("impossible de se connecter à la base de données: %w", err)
	}

	return db, nil
}

// CreateTables crée les tables nécessaires si elles n'existent pas
func CreateTables(db *sql.DB, logger *zap.Logger) error {
	queries := []string{
		createUsersTable,
		createRefreshTokensTable,
		createUsersIndexes,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			logger.Error("Erreur création table", zap.Error(err), zap.String("query", query))
			return fmt.Errorf("erreur création table: %w", err)
		}
	}

	logger.Info("✅ Tables PostgreSQL créées avec succès")
	return nil
}

// SQL pour création des tables
const createUsersTable = `
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(30) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    bio TEXT,
    avatar TEXT,
    role VARCHAR(20) NOT NULL DEFAULT 'user',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    is_active BOOLEAN NOT NULL DEFAULT true,
    is_verified BOOLEAN NOT NULL DEFAULT false,
    last_login_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
`

const createRefreshTokensTable = `
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token TEXT UNIQUE NOT NULL,
    expires_at BIGINT NOT NULL,
    created_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())
);
`

const createUsersIndexes = `
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token ON refresh_tokens(token);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
`

// Migrations simples pour la Phase 1
func RunMigrations(db *sql.DB, logger *zap.Logger) error {
	// Pour la Phase 1, on utilise CreateTables
	// En production, utiliser un système de migration plus robuste comme golang-migrate
	return CreateTables(db, logger)
}
