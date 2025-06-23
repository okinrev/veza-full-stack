// backend/internal/database/db.go
package database

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/lib/pq"
)

// DB wraps sql.DB to provide additional methods
type DB struct {
	*sql.DB
}

// NewConnection creates a new database connection
func NewConnection(databaseURL string) (*DB, error) {
	log.Printf("🔌 Tentative de connexion à la base de données avec l'URL: %s", databaseURL)

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Printf("❌ Erreur lors de l'ouverture de la connexion: %v", err)
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		log.Printf("❌ Erreur lors du ping de la base de données: %v", err)
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)

	log.Println("✅ Connexion à la base de données établie avec succès")
	return &DB{DB: db}, nil
}

// Query wraps sql.DB.Query to use our DB type
func (db *DB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return db.DB.Query(query, args...)
}

// QueryRow wraps sql.DB.QueryRow to use our DB type
func (db *DB) QueryRow(query string, args ...interface{}) *sql.Row {
	return db.DB.QueryRow(query, args...)
}

// Exec wraps sql.DB.Exec to use our DB type
func (db *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return db.DB.Exec(query, args...)
}

// RunMigrations runs all SQL migration files in the migrations directory
func RunMigrations(db *DB) error {
	// Create migrations table if it doesn't exist
	if err := createMigrationsTable(db); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get all migration files
	migrationFiles, err := getMigrationFiles("./internal/database/migrations")
	if err != nil {
		return fmt.Errorf("failed to get migration files: %w", err)
	}

	// Get applied migrations
	appliedMigrations, err := getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Run pending migrations
	for _, file := range migrationFiles {
		if _, applied := appliedMigrations[file]; !applied {
			log.Printf("Running migration: %s", file)

			if err := runMigrationFile(db, filepath.Join("./internal/database/migrations", file)); err != nil {
				return fmt.Errorf("failed to run migration %s: %w", file, err)
			}

			if err := recordMigration(db, file); err != nil {
				return fmt.Errorf("failed to record migration %s: %w", file, err)
			}

			log.Printf("✅ Migration completed: %s", file)
		}
	}

	log.Println("✅ All migrations completed")
	return nil
}

// createMigrationsTable creates the migrations tracking table
func createMigrationsTable(db *DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS migrations (
			id SERIAL PRIMARY KEY,
			filename VARCHAR(255) NOT NULL UNIQUE,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err := db.Exec(query)
	return err
}

// getMigrationFiles returns all .sql files in the migrations directory, sorted
func getMigrationFiles(dir string) ([]string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var migrationFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			// Remove .2 suffix if present (from your files)
			name := file.Name()
			if strings.HasSuffix(name, ".sql.2") {
				name = strings.TrimSuffix(name, ".2")
			}
			migrationFiles = append(migrationFiles, name)
		}
	}

	// Sort files to ensure consistent migration order
	sort.Strings(migrationFiles)
	return migrationFiles, nil
}

// getAppliedMigrations returns a map of applied migration filenames
func getAppliedMigrations(db *DB) (map[string]bool, error) {
	rows, err := db.Query("SELECT filename FROM migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	appliedMigrations := make(map[string]bool)
	for rows.Next() {
		var filename string
		if err := rows.Scan(&filename); err != nil {
			return nil, err
		}
		appliedMigrations[filename] = true
	}

	return appliedMigrations, rows.Err()
}

// runMigrationFile executes a single migration file
func runMigrationFile(db *DB, filePath string) error {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Split by semicolons and execute each statement
	statements := strings.Split(string(content), ";")
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("failed to execute statement: %s, error: %w", stmt, err)
		}
	}

	return nil
}

// recordMigration records that a migration has been applied
func recordMigration(db *DB, filename string) error {
	_, err := db.Exec("INSERT INTO migrations (filename) VALUES ($1)", filename)
	return err
}
