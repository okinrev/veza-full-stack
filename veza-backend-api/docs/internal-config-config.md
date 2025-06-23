# Documentation - internal/config/config.go

## Vue d'ensemble

Le package `config` gère la configuration de l'application en centralisant toutes les variables d'environnement et les paramètres de configuration. Il utilise le pattern Builder pour construire une configuration complète avec des valeurs par défaut.

## Structures de Configuration

### `Config`

Structure principale qui contient toute la configuration de l'application.

```go
type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    JWT      JWTConfig
}
```

### `ServerConfig`

Configuration du serveur HTTP.

```go
type ServerConfig struct {
    Port            string        // Port d'écoute (défaut: "8080")
    ReadTimeout     time.Duration // Timeout de lecture (défaut: 10s)
    WriteTimeout    time.Duration // Timeout d'écriture (défaut: 10s)
    ShutdownTimeout time.Duration // Timeout d'arrêt gracieux (défaut: 30s)
    Environment     string        // Environnement (défaut: "development")
}
```

**Variables d'environnement** :
- `PORT` - Port du serveur
- `READ_TIMEOUT` - Timeout de lecture (format: "10s", "1m")
- `WRITE_TIMEOUT` - Timeout d'écriture
- `SHUTDOWN_TIMEOUT` - Timeout d'arrêt
- `ENVIRONMENT` - Environnement (development/production)

### `DatabaseConfig`

Configuration de la base de données PostgreSQL.

```go
type DatabaseConfig struct {
    URL          string        // URL complète de connexion
    Host         string        // Hôte de la base (défaut: "localhost")
    Port         string        // Port de la base (défaut: "5432")
    Username     string        // Nom d'utilisateur (défaut: "postgres")
    Password     string        // Mot de passe
    Database     string        // Nom de la base (défaut: "veza_db")
    SSLMode      string        // Mode SSL (défaut: "disable")
    MaxOpenConns int           // Nombre max de connexions ouvertes (défaut: 25)
    MaxIdleConns int           // Nombre max de connexions inactives (défaut: 25)
    MaxLifetime  time.Duration // Durée de vie max des connexions (défaut: 5m)
}
```

**Variables d'environnement** :
- `DATABASE_URL` - URL complète (prioritaire sur les autres)
- `DB_HOST` - Hôte de la base
- `DB_PORT` - Port de la base
- `DB_USERNAME` - Nom d'utilisateur
- `DB_PASSWORD` - Mot de passe
- `DB_NAME` - Nom de la base
- `DB_SSLMODE` - Mode SSL
- `DB_MAX_OPEN_CONNS` - Connexions ouvertes max
- `DB_MAX_IDLE_CONNS` - Connexions inactives max
- `DB_MAX_LIFETIME` - Durée de vie des connexions

### `JWTConfig`

Configuration JWT pour l'authentification.

```go
type JWTConfig struct {
    Secret         string        // Clé secrète de signature
    ExpirationTime time.Duration // Durée de validité (défaut: 24h)
    RefreshTime    time.Duration // Durée de validité refresh (défaut: 7j)
}
```

**Variables d'environnement** :
- `JWT_SECRET` - Clé secrète (défaut: clé de développement)
- `JWT_EXPIRATION` - Durée de validité du token
- `JWT_REFRESH_TIME` - Durée de validité du refresh token

## Fonctions Principales

### `New()`

**Description** : Crée une nouvelle instance de configuration en lisant les variables d'environnement.

**Signature** :
```go
func New() *Config
```

**Comportement** :
1. Lit toutes les variables d'environnement
2. Applique les valeurs par défaut si variables manquantes
3. Construit l'URL de base de données si `DATABASE_URL` non fournie
4. Retourne une configuration complète

**Exemple d'utilisation** :
```go
cfg := config.New()
fmt.Printf("Port: %s", cfg.Server.Port)
fmt.Printf("DB URL: %s", cfg.Database.URL)
```

### `getEnv()`

**Description** : Fonction utilitaire pour récupérer une variable d'environnement avec valeur par défaut.

**Signature** :
```go
func getEnv(key, defaultValue string) string
```

**Paramètres** :
- `key` - Nom de la variable d'environnement
- `defaultValue` - Valeur par défaut si variable non définie

**Retour** : Valeur de la variable ou valeur par défaut

### `getIntEnv()`

**Description** : Fonction utilitaire pour récupérer une variable d'environnement entière.

**Signature** :
```go
func getIntEnv(key string, defaultValue int) int
```

**Paramètres** :
- `key` - Nom de la variable d'environnement
- `defaultValue` - Valeur par défaut si variable non définie ou invalide

**Retour** : Valeur entière de la variable ou valeur par défaut

### `getDurationEnv()`

**Description** : Fonction utilitaire pour récupérer une variable d'environnement de durée.

**Signature** :
```go
func getDurationEnv(key string, defaultValue time.Duration) time.Duration
```

**Paramètres** :
- `key` - Nom de la variable d'environnement
- `defaultValue` - Valeur par défaut si variable non définie ou invalide

**Retour** : Durée parsée ou valeur par défaut

**Formats acceptés** : "10s", "5m", "1h", "24h", etc.

## Construction de l'URL de Base de Données

Si `DATABASE_URL` n'est pas fournie, l'URL est construite automatiquement :

```go
databaseURL = "postgres://" + username + ":" + password + "@" + host + ":" + port + "/" + database + "?sslmode=" + sslmode
```

**Exemple** :
```
postgres://postgres:password@localhost:5432/veza_db?sslmode=disable
```

## Exemples de Configuration

### Fichier .env de Développement

```env
# Serveur
PORT=8080
ENVIRONMENT=development
READ_TIMEOUT=10s
WRITE_TIMEOUT=10s
SHUTDOWN_TIMEOUT=30s

# Base de données
DB_HOST=localhost
DB_PORT=5432
DB_USERNAME=postgres
DB_PASSWORD=mypassword
DB_NAME=veza_db
DB_SSLMODE=disable
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=25
DB_MAX_LIFETIME=5m

# JWT
JWT_SECRET=your-super-secret-development-key
JWT_EXPIRATION=24h
JWT_REFRESH_TIME=168h
```

### Fichier .env de Production

```env
# Serveur
PORT=80
ENVIRONMENT=production
READ_TIMEOUT=30s
WRITE_TIMEOUT=30s
SHUTDOWN_TIMEOUT=60s

# Base de données
DATABASE_URL=postgres://user:password@prod-db.example.com:5432/veza_prod?sslmode=require
DB_MAX_OPEN_CONNS=50
DB_MAX_IDLE_CONNS=10
DB_MAX_LIFETIME=1h

# JWT
JWT_SECRET=your-super-secure-production-key-change-this
JWT_EXPIRATION=1h
JWT_REFRESH_TIME=24h
```

## Utilisation dans l'Application

### Initialisation

```go
// Dans main.go
cfg := config.New()

// Utilisation dans les services
db, err := database.NewConnection(cfg.Database.URL)
jwtService := auth.NewJWTService(cfg.JWT.Secret)
```

### Accès aux Valeurs

```go
// Configuration serveur
port := cfg.Server.Port
environment := cfg.Server.Environment

// Configuration base de données
dbURL := cfg.Database.URL
maxConns := cfg.Database.MaxOpenConns

// Configuration JWT
jwtSecret := cfg.JWT.Secret
expiration := cfg.JWT.ExpirationTime
```

## Sécurité

### Variables Sensibles

**Important** : Les variables suivantes ne doivent jamais être exposées :
- `JWT_SECRET` - Clé de signature JWT
- `DB_PASSWORD` - Mot de passe base de données
- `DATABASE_URL` - Peut contenir des informations sensibles

### Bonnes Pratiques

1. **Secrets en Production** : Utiliser un gestionnaire de secrets (AWS Secrets Manager, Kubernetes Secrets)
2. **Variables d'Environnement** : Ne jamais commiter les fichiers `.env` de production
3. **Rotation des Clés** : Changer régulièrement `JWT_SECRET` en production
4. **SSL** : Utiliser `sslmode=require` en production

## Validation

Le système de configuration n'effectue pas de validation stricte, mais applique des valeurs par défaut saines. Pour une validation plus stricte, vous pouvez ajouter :

```go
func (c *Config) Validate() error {
    if c.JWT.Secret == "your-super-secret-key-change-in-production" {
        return errors.New("JWT secret must be changed in production")
    }
    
    if c.Server.Environment == "production" && c.Database.SSLMode == "disable" {
        return errors.New("SSL must be enabled in production")
    }
    
    return nil
}
```

## Intégration avec les Modules

### Frontend React

Le frontend utilise les mêmes endpoints configurés :
```javascript
const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1';
```

### Modules Rust

Les modules Rust peuvent lire les mêmes variables d'environnement :
```rust
let database_url = std::env::var("DATABASE_URL").expect("DATABASE_URL must be set");
let jwt_secret = std::env::var("JWT_SECRET").expect("JWT_SECRET must be set");
```

## Extensions Possibles

1. **Validation Schema** : Validation stricte des valeurs
2. **Configuration Hot-Reload** : Rechargement à chaud
3. **Configurations par Environnement** : Fichiers séparés par environnement
4. **Chiffrement** : Chiffrement des valeurs sensibles 