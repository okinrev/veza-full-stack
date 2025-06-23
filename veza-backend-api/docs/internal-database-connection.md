# Documentation - internal/database/connection.go

## Vue d'ensemble

Le package `database` gère toutes les interactions avec la base de données PostgreSQL, incluant la connexion, les migrations automatiques, et l'encapsulation des opérations de base de données.

## Structure Principale

### `DB`

Structure qui encapsule `*sql.DB` pour fournir des méthodes additionnelles.

```go
type DB struct {
    *sql.DB
}
```

Cette structure permet d'étendre les fonctionnalités de la base de données standard Go tout en conservant la compatibilité.

## Fonctions Principales

### `NewConnection()`

**Description** : Crée une nouvelle connexion à la base de données PostgreSQL.

**Signature** :
```go
func NewConnection(databaseURL string) (*DB, error)
```

**Paramètres** :
- `databaseURL` - URL de connexion PostgreSQL (format: `postgres://user:password@host:port/database?sslmode=disable`)

**Retour** :
- `*DB` - Instance de base de données encapsulée
- `error` - Erreur de connexion si échec

**Comportement** :
1. Ouvre la connexion avec le driver PostgreSQL
2. Teste la connexion avec `Ping()`
3. Configure le pool de connexions (25 connexions max ouvertes et inactives)
4. Retourne l'instance encapsulée

**Exemple d'utilisation** :
```go
databaseURL := "postgres://postgres:password@localhost:5432/veza_db?sslmode=disable"
db, err := database.NewConnection(databaseURL)
if err != nil {
    log.Fatal("Database connection failed:", err)
}
defer db.Close()
```

### Méthodes de Requêtes

#### `Query()`

**Description** : Exécute une requête SELECT et retourne plusieurs lignes.

**Signature** :
```go
func (db *DB) Query(query string, args ...interface{}) (*sql.Rows, error)
```

**Exemple** :
```go
rows, err := db.Query("SELECT id, username FROM users WHERE active = $1", true)
if err != nil {
    return err
}
defer rows.Close()

for rows.Next() {
    var id int
    var username string
    err := rows.Scan(&id, &username)
    if err != nil {
        return err
    }
    fmt.Printf("ID: %d, Username: %s\n", id, username)
}
```

#### `QueryRow()`

**Description** : Exécute une requête qui retourne une seule ligne.

**Signature** :
```go
func (db *DB) QueryRow(query string, args ...interface{}) *sql.Row
```

**Exemple** :
```go
var userCount int
err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
if err != nil {
    return err
}
fmt.Printf("Total users: %d\n", userCount)
```

#### `Exec()`

**Description** : Exécute une requête INSERT, UPDATE ou DELETE.

**Signature** :
```go
func (db *DB) Exec(query string, args ...interface{}) (sql.Result, error)
```

**Exemple** :
```go
result, err := db.Exec("UPDATE users SET last_login_at = $1 WHERE id = $2", time.Now(), userID)
if err != nil {
    return err
}

rowsAffected, err := result.RowsAffected()
if err != nil {
    return err
}
fmt.Printf("Rows affected: %d\n", rowsAffected)
```

## Système de Migrations

### `RunMigrations()`

**Description** : Exécute automatiquement toutes les migrations SQL non appliquées.

**Signature** :
```go
func RunMigrations(db *DB) error
```

**Processus** :
1. Crée la table `migrations` si elle n'existe pas
2. Récupère tous les fichiers `.sql` du dossier `migrations/`
3. Compare avec les migrations déjà appliquées
4. Exécute les migrations manquantes dans l'ordre alphabétique
5. Enregistre chaque migration appliquée

**Structure de la table migrations** :
```sql
CREATE TABLE migrations (
    id SERIAL PRIMARY KEY,
    filename VARCHAR(255) NOT NULL UNIQUE,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)
```

### `createMigrationsTable()`

**Description** : Crée la table de suivi des migrations.

**Signature** :
```go
func createMigrationsTable(db *DB) error
```

### `getMigrationFiles()`

**Description** : Récupère tous les fichiers de migration du répertoire.

**Signature** :
```go
func getMigrationFiles(dir string) ([]string, error)
```

**Comportement** :
- Lit tous les fichiers `.sql` du répertoire `migrations/`
- Trie les fichiers par ordre alphabétique
- Gère les fichiers avec suffixe `.2` (nettoyage automatique)

### `getAppliedMigrations()`

**Description** : Récupère la liste des migrations déjà appliquées.

**Signature** :
```go
func getAppliedMigrations(db *DB) (map[string]bool, error)
```

**Retour** : Map des noms de fichiers déjà appliqués

### `runMigrationFile()`

**Description** : Exécute un fichier de migration spécifique.

**Signature** :
```go
func runMigrationFile(db *DB, filePath string) error
```

**Comportement** :
1. Lit le contenu du fichier SQL
2. Divise le contenu par les points-virgules
3. Exécute chaque statement séparément
4. Gère les erreurs de syntaxe SQL

### `recordMigration()`

**Description** : Enregistre qu'une migration a été appliquée.

**Signature** :
```go
func recordMigration(db *DB, filename string) error
```

## Structure des Migrations

### Conventions de Nommage

Les fichiers de migration doivent suivre cette convention :
```
001_create_users_table.sql
002_create_listings_table.sql
003_add_indexes.sql
```

**Format** : `{numéro}_{description}.sql`

### Exemple de Fichier de Migration

```sql
-- 001_create_users_table.sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(50),
    last_name VARCHAR(50),
    bio TEXT,
    avatar VARCHAR(255),
    role VARCHAR(20) DEFAULT 'user',
    is_active BOOLEAN DEFAULT true,
    is_verified BOOLEAN DEFAULT false,
    last_login_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Index pour améliorer les performances
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_active ON users(is_active);
```

## Configuration du Pool de Connexions

### Paramètres par Défaut

```go
// Configure connection pool
db.SetMaxOpenConns(25)  // Maximum 25 connexions ouvertes
db.SetMaxIdleConns(25)  // Maximum 25 connexions inactives
```

### Optimisation Production

Pour la production, ajustez selon votre charge :

```go
db.SetMaxOpenConns(100)        // Plus de connexions simultanées
db.SetMaxIdleConns(10)         // Moins de connexions inactives
db.SetConnMaxLifetime(time.Hour) // Renouvellement des connexions
```

## Gestion d'Erreurs

### Erreurs de Connexion

```go
db, err := database.NewConnection(cfg.Database.URL)
if err != nil {
    log.Fatal("Database connection failed:", err)
}
```

**Erreurs communes** :
- Mauvaise URL de connexion
- Serveur PostgreSQL inaccessible
- Mauvaises credentials
- Base de données inexistante

### Erreurs de Migration

```go
if err := database.RunMigrations(db); err != nil {
    log.Printf("Migration warning: %v", err)
}
```

**Erreurs communes** :
- Fichier SQL malformé
- Conflit de schéma
- Permissions insuffisantes
- Migration déjà appliquée partiellement

## Intégration avec l'Application

### Initialisation dans main.go

```go
// Configuration
cfg := config.New()

// Database
db, err := database.NewConnection(cfg.Database.URL)
if err != nil {
    log.Fatal("Database connection failed:", err)
}
defer db.Close()

// Migrations
if err := database.RunMigrations(db); err != nil {
    log.Printf("Migration warning: %v", err)
}
```

### Utilisation dans les Services

```go
type UserService struct {
    db *database.DB
}

func NewUserService(db *database.DB) *UserService {
    return &UserService{db: db}
}

func (s *UserService) GetUserByID(id int) (*models.User, error) {
    user := &models.User{}
    err := s.db.QueryRow(`
        SELECT id, username, email, first_name, last_name, created_at 
        FROM users WHERE id = $1 AND is_active = true
    `, id).Scan(&user.ID, &user.Username, &user.Email, &user.FirstName, &user.LastName, &user.CreatedAt)
    
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, ErrUserNotFound
        }
        return nil, err
    }
    
    return user, nil
}
```

## Sécurité

### Protection contre l'Injection SQL

**Toujours utiliser les placeholders** :
```go
// ✅ Correct - Utilise des placeholders
db.Query("SELECT * FROM users WHERE email = $1", email)

// ❌ Incorrect - Vulnérable à l'injection SQL
db.Query("SELECT * FROM users WHERE email = '" + email + "'")
```

### Gestion des Permissions

Créer un utilisateur de base de données avec permissions limitées :

```sql
-- Créer un utilisateur applicatif
CREATE USER veza_app WITH PASSWORD 'secure_password';

-- Accorder uniquement les permissions nécessaires
GRANT CONNECT ON DATABASE veza_db TO veza_app;
GRANT USAGE ON SCHEMA public TO veza_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO veza_app;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO veza_app;
```

## Performance

### Indexes Recommandés

```sql
-- Indexes pour améliorer les performances
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_listings_user_id ON listings(user_id);
CREATE INDEX idx_messages_room_id ON messages(room_id);
CREATE INDEX idx_messages_created_at ON messages(created_at);
```

### Monitoring des Requêtes

```go
// Ajouter du logging pour les requêtes lentes
func (db *DB) QueryWithLogging(query string, args ...interface{}) (*sql.Rows, error) {
    start := time.Now()
    rows, err := db.Query(query, args...)
    duration := time.Since(start)
    
    if duration > 100*time.Millisecond {
        log.Printf("Slow query (%v): %s", duration, query)
    }
    
    return rows, err
}
```

## Backup et Restauration

### Script de Backup

```bash
#!/bin/bash
pg_dump -h localhost -U postgres -d veza_db > backup_$(date +%Y%m%d_%H%M%S).sql
```

### Script de Restauration

```bash
#!/bin/bash
psql -h localhost -U postgres -d veza_db < backup_file.sql
```

## Intégration avec les Modules Rust

Les modules Rust peuvent utiliser la même base de données :

```rust
// Dans Cargo.toml
[dependencies]
sqlx = { version = "0.7", features = ["runtime-tokio-rustls", "postgres"] }

// Dans le code Rust
let database_url = std::env::var("DATABASE_URL").expect("DATABASE_URL must be set");
let pool = PgPool::connect(&database_url).await?;
```

## Tests

### Tests d'Intégration

```go
func TestDatabaseConnection(t *testing.T) {
    cfg := config.New()
    db, err := database.NewConnection(cfg.Database.URL)
    assert.NoError(t, err)
    assert.NotNil(t, db)
    
    // Test ping
    err = db.Ping()
    assert.NoError(t, err)
    
    db.Close()
}
```

### Base de Données de Test

```go
func setupTestDB() *database.DB {
    testURL := "postgres://postgres:password@localhost:5432/veza_test?sslmode=disable"
    db, err := database.NewConnection(testURL)
    if err != nil {
        panic(err)
    }
    
    // Nettoyer les données de test
    db.Exec("TRUNCATE TABLE users, listings, messages CASCADE")
    
    return db
}
``` 