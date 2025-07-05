# Module de Configuration - Backend API Go

## 📋 Vue d'Ensemble

Le module `internal/config/config.go` gère toute la configuration du service backend API Veza. Il centralise la lecture des variables d'environnement, la validation des paramètres et la création d'une structure de configuration unifiée pour tous les composants du système.

## 🏗️ Architecture

### Structure de Configuration

```mermaid
graph TD
    A[Config] --> B[ServerConfig]
    A --> C[DatabaseConfig]
    A --> D[JWTConfig]
    A --> E[RedisConfig]
    A --> F[NATSConfig]
    A --> G[CacheConfig]
    A --> H[QueueConfig]
    
    B --> B1[Port, Timeouts, Environment]
    C --> C1[URL, Pool, SSL]
    D --> D1[Secret, Expiration, Refresh]
    E --> E1[URL, Pool, TLS]
    F --> F1[URL, JetStream, Reconnect]
    G --> G1[Levels, Memory, Compression]
    H --> H1[Workers, Queue, Retry]
```

### Hiérarchie des Configurations

```go
type Config struct {
    Server   ServerConfig    // Configuration du serveur HTTP
    Database DatabaseConfig  // Configuration PostgreSQL
    JWT      JWTConfig       // Configuration JWT
    Redis    RedisConfig     // Configuration Redis
    NATS     NATSConfig      // Configuration NATS
    Cache    CacheConfig     // Configuration du cache
    Queue    QueueConfig     // Configuration des queues
}
```

## 🔧 Structures de Configuration

### `ServerConfig`

```go
type ServerConfig struct {
    Port            string        // Port d'écoute du serveur
    ReadTimeout     time.Duration // Timeout de lecture
    WriteTimeout    time.Duration // Timeout d'écriture
    ShutdownTimeout time.Duration // Timeout d'arrêt gracieux
    Environment     string        // Environnement (dev/prod)
}
```

**Variables d'environnement**:
- `PORT`: Port d'écoute (défaut: "8080")
- `READ_TIMEOUT`: Timeout de lecture (défaut: 10s)
- `WRITE_TIMEOUT`: Timeout d'écriture (défaut: 10s)
- `SHUTDOWN_TIMEOUT`: Timeout d'arrêt (défaut: 30s)
- `ENVIRONMENT`: Environnement (défaut: "development")

### `DatabaseConfig`

```go
type DatabaseConfig struct {
    URL          string        // URL de connexion complète
    Host         string        // Hôte de la base de données
    Port         string        // Port de la base de données
    Username     string        // Nom d'utilisateur
    Password     string        // Mot de passe
    Database     string        // Nom de la base de données
    SSLMode      string        // Mode SSL
    MaxOpenConns int           // Connexions max ouvertes
    MaxIdleConns int           // Connexions max inactives
    MaxLifetime  time.Duration // Durée de vie max des connexions
}
```

**Variables d'environnement**:
- `DATABASE_URL`: URL complète (priorité)
- `DATABASE_HOST`: Hôte (défaut: "localhost")
- `DATABASE_PORT`: Port (défaut: "5432")
- `DATABASE_USER`: Utilisateur (défaut: "postgres")
- `DATABASE_PASSWORD`: Mot de passe
- `DATABASE_NAME`: Base de données (défaut: "veza_dev")
- `DATABASE_MAX_OPEN_CONNS`: Connexions max (défaut: 100)
- `DATABASE_MAX_IDLE_CONNS`: Connexions inactives (défaut: 25)
- `DATABASE_CONN_MAX_LIFETIME`: Durée de vie (défaut: 5min)

**Construction automatique de l'URL**:
```go
if databaseURL == "" {
    host := getEnv("DATABASE_HOST", "localhost")
    port := getEnv("DATABASE_PORT", "5432")
    username := getEnv("DATABASE_USER", "postgres")
    password := getEnv("DATABASE_PASSWORD", "")
    database := getEnv("DATABASE_NAME", "veza_dev")
    sslmode := "disable"

    databaseURL = "postgres://" + username + ":" + password + "@" + host + ":" + port + "/" + database + "?sslmode=" + sslmode
}
```

### `JWTConfig`

```go
type JWTConfig struct {
    Secret          string        // Secret de signature
    ExpirationTime  time.Duration // Durée de vie des access tokens
    RefreshTime     time.Duration // Durée de vie des refresh tokens
    RefreshTTL      time.Duration // TTL des refresh tokens
    RefreshRotation bool          // Rotation automatique des refresh tokens
}
```

**Variables d'environnement**:
- `JWT_ACCESS_SECRET`: Secret de signature (défaut: "your-super-secret-key-change-in-production")
- `JWT_ACCESS_TTL`: Durée access token (défaut: 15min)
- `JWT_REFRESH_TTL`: Durée refresh token (défaut: 7 jours)
- `JWT_REFRESH_ROTATION`: Rotation automatique (défaut: true)

### `RedisConfig`

```go
type RedisConfig struct {
    URL          string        // URL de connexion Redis
    Host         string        // Hôte Redis
    Port         string        // Port Redis
    Password     string        // Mot de passe Redis
    Database     int           // Base de données Redis
    MaxRetries   int           // Nombre max de tentatives
    DialTimeout  time.Duration // Timeout de connexion
    ReadTimeout  time.Duration // Timeout de lecture
    WriteTimeout time.Duration // Timeout d'écriture
    PoolSize     int           // Taille du pool de connexions
    PoolTimeout  time.Duration // Timeout du pool
    IdleTimeout  time.Duration // Timeout d'inactivité
    MaxConnAge   time.Duration // Âge max des connexions
    EnableTLS    bool          // Activation TLS
}
```

**Variables d'environnement**:
- `REDIS_URL`: URL complète Redis
- `REDIS_HOST`: Hôte (défaut: "localhost")
- `REDIS_PORT`: Port (défaut: "6379")
- `REDIS_PASSWORD`: Mot de passe
- `REDIS_DATABASE`: Base de données (défaut: 0)
- `REDIS_MAX_RETRIES`: Tentatives max (défaut: 3)
- `REDIS_DIAL_TIMEOUT`: Timeout connexion (défaut: 5s)
- `REDIS_READ_TIMEOUT`: Timeout lecture (défaut: 3s)
- `REDIS_WRITE_TIMEOUT`: Timeout écriture (défaut: 3s)
- `REDIS_POOL_SIZE`: Taille pool (défaut: 100)
- `REDIS_POOL_TIMEOUT`: Timeout pool (défaut: 5s)
- `REDIS_IDLE_TIMEOUT`: Timeout inactivité (défaut: 5min)
- `REDIS_MAX_CONN_AGE`: Âge max connexions (défaut: 10min)
- `REDIS_ENABLE_TLS`: Activation TLS (défaut: false)

### `NATSConfig`

```go
type NATSConfig struct {
    URL                   string        // URL de connexion NATS
    ClusterID             string        // ID du cluster
    ClientID              string        // ID du client
    MaxReconnects         int           // Nombre max de reconnexions
    ReconnectWait         time.Duration // Délai de reconnexion
    ConnectTimeout        time.Duration // Timeout de connexion
    MaxPendingMsgs        int           // Messages en attente max
    MaxPendingBytes       int64         // Bytes en attente max
    EnableJetStream       bool          // Activation JetStream
    StreamRetentionPolicy string        // Politique de rétention
}
```

**Variables d'environnement**:
- `NATS_URL`: URL NATS (défaut: "nats://localhost:4222")
- `NATS_CLUSTER_ID`: ID cluster (défaut: "veza-cluster")
- `NATS_CLIENT_ID`: ID client (défaut: "veza-backend")
- `NATS_MAX_RECONNECTS`: Reconnexions max (défaut: 10)
- `NATS_RECONNECT_WAIT`: Délai reconnexion (défaut: 2s)
- `NATS_CONNECT_TIMEOUT`: Timeout connexion (défaut: 5s)
- `NATS_MAX_PENDING_MSGS`: Messages en attente (défaut: 10000)
- `NATS_MAX_PENDING_BYTES`: Bytes en attente (défaut: 64MB)
- `NATS_ENABLE_JETSTREAM`: Activation JetStream (défaut: true)
- `NATS_STREAM_RETENTION`: Politique rétention (défaut: "limits")

### `CacheConfig`

```go
type CacheConfig struct {
    EnableLevel1     bool          // Activation cache niveau 1 (mémoire)
    EnableLevel2     bool          // Activation cache niveau 2 (Redis)
    EnableLevel3     bool          // Activation cache niveau 3 (disque)
    MaxMemoryMB      int           // Mémoire max en MB
    CompressionLevel int           // Niveau de compression
    StatsInterval    time.Duration // Intervalle des statistiques
    WarmupEnabled    bool          // Activation du warmup
    WarmupInterval   time.Duration // Intervalle de warmup
}
```

**Variables d'environnement**:
- `CACHE_ENABLE_L1`: Cache niveau 1 (défaut: true)
- `CACHE_ENABLE_L2`: Cache niveau 2 (défaut: true)
- `CACHE_ENABLE_L3`: Cache niveau 3 (défaut: false)
- `CACHE_MAX_MEMORY_MB`: Mémoire max (défaut: 512MB)
- `CACHE_COMPRESSION_LEVEL`: Niveau compression (défaut: 1)
- `CACHE_STATS_INTERVAL`: Intervalle stats (défaut: 30s)
- `CACHE_WARMUP_ENABLED`: Activation warmup (défaut: true)
- `CACHE_WARMUP_INTERVAL`: Intervalle warmup (défaut: 10min)

### `QueueConfig`

```go
type QueueConfig struct {
    MaxWorkers          int           // Nombre max de workers
    MaxQueueSize        int           // Taille max de la queue
    ProcessingTimeout   time.Duration // Timeout de traitement
    RetryMaxAttempts    int           // Nombre max de tentatives
    RetryBackoffBase    time.Duration // Délai de base pour retry
    DeadLetterQueueSize int           // Taille de la DLQ
}
```

**Variables d'environnement**:
- `QUEUE_MAX_WORKERS`: Workers max (défaut: 50)
- `QUEUE_MAX_SIZE`: Taille queue max (défaut: 10000)
- `QUEUE_PROCESSING_TIMEOUT`: Timeout traitement (défaut: 30s)
- `QUEUE_RETRY_MAX_ATTEMPTS`: Tentatives max (défaut: 3)
- `QUEUE_RETRY_BACKOFF_BASE`: Délai retry (défaut: 1s)
- `QUEUE_DLQ_SIZE`: Taille DLQ (défaut: 1000)

## 🔧 Fonctions Utilitaires

### `New()`

```go
func New() *Config
```

**Objectif**: Création d'une nouvelle instance de configuration avec les valeurs par défaut et les variables d'environnement.

**Logique**:
1. Lecture de `DATABASE_URL` ou construction automatique
2. Configuration de tous les composants avec les valeurs par défaut
3. Override avec les variables d'environnement
4. Retour de la configuration complète

### `getEnv()`

```go
func getEnv(key, defaultValue string) string
```

**Objectif**: Récupération d'une variable d'environnement avec valeur par défaut.

**Paramètres**:
- `key`: Clé de la variable d'environnement
- `defaultValue`: Valeur par défaut si la variable n'existe pas

### `getIntEnv()`

```go
func getIntEnv(key string, defaultValue int) int
```

**Objectif**: Récupération d'une variable d'environnement entière.

**Gestion d'erreurs**: Retourne la valeur par défaut si la conversion échoue.

### `getDurationEnv()`

```go
func getDurationEnv(key string, defaultValue time.Duration) time.Duration
```

**Objectif**: Récupération d'une variable d'environnement de type durée.

**Format attendu**: Durée Go (ex: "30s", "5m", "1h")

### `getBoolEnv()`

```go
func getBoolEnv(key string, defaultValue bool) bool
```

**Objectif**: Récupération d'une variable d'environnement booléenne.

**Valeurs acceptées**: "true", "1", "yes", "on" pour `true`

### `getInt64Env()`

```go
func getInt64Env(key string, defaultValue int64) int64
```

**Objectif**: Récupération d'une variable d'environnement int64.

**Utilisation**: Pour les grandes valeurs (bytes, timestamps)

## 🛠️ Configuration par Environnement

### Développement

```bash
# Variables d'environnement de développement
export ENVIRONMENT=development
export PORT=8080
export DATABASE_URL=postgres://postgres:password@localhost:5432/veza_dev
export JWT_ACCESS_SECRET=dev-secret-key
export REDIS_URL=redis://localhost:6379
export NATS_URL=nats://localhost:4222
```

### Production

```bash
# Variables d'environnement de production
export ENVIRONMENT=production
export PORT=8080
export DATABASE_URL=postgres://user:password@db:5432/veza_prod
export JWT_ACCESS_SECRET=production-secret-key-change-this
export REDIS_URL=redis://redis:6379
export NATS_URL=nats://nats:4222
export REDIS_ENABLE_TLS=true
export NATS_ENABLE_JETSTREAM=true
```

### Staging

```bash
# Variables d'environnement de staging
export ENVIRONMENT=staging
export PORT=8080
export DATABASE_URL=postgres://user:password@staging-db:5432/veza_staging
export JWT_ACCESS_SECRET=staging-secret-key
export REDIS_URL=redis://staging-redis:6379
export NATS_URL=nats://staging-nats:4222
```

## 🔒 Bonnes Pratiques de Sécurité

### Secrets et Clés

1. **JWT Secret**
   - Minimum 32 caractères en production
   - Utilisation de générateurs cryptographiques
   - Rotation régulière des clés

2. **Mots de Passe Base de Données**
   - Stockage sécurisé (secrets managers)
   - Pas de hardcoding dans le code
   - Rotation régulière

3. **URLs de Connexion**
   - Utilisation de TLS en production
   - Validation des certificats
   - Isolation réseau

### Validation de Configuration

```go
// Exemple de validation à ajouter
func (c *Config) Validate() error {
    if c.Environment == "production" {
        if len(c.JWT.Secret) < 32 {
            return errors.New("JWT secret too short for production")
        }
        if !strings.Contains(c.Database.URL, "sslmode=require") {
            return errors.New("SSL required for production database")
        }
    }
    return nil
}
```

## 📊 Optimisations de Performance

### Pool de Connexions

1. **Base de Données**
   - `MaxOpenConns`: 100 (optimisé haute charge)
   - `MaxIdleConns`: 25
   - `MaxLifetime`: 5 minutes

2. **Redis**
   - `PoolSize`: 100 (optimisé haute charge)
   - `PoolTimeout`: 5 secondes
   - `IdleTimeout`: 5 minutes

3. **NATS**
   - `MaxPendingMsgs`: 10000
   - `MaxPendingBytes`: 64MB
   - `MaxReconnects`: 10

### Cache Multi-Niveaux

1. **Niveau 1 (Mémoire)**
   - Accès ultra-rapide
   - Limité par la RAM
   - Invalidation automatique

2. **Niveau 2 (Redis)**
   - Persistance entre redémarrages
   - Partage entre instances
   - TTL configurable

3. **Niveau 3 (Disque)**
   - Stockage permanent
   - Compression avancée
   - Archivage automatique

## 🧪 Tests

### Tests de Configuration

```go
func TestConfigNew(t *testing.T) {
    // Test avec variables d'environnement
    os.Setenv("PORT", "9090")
    os.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/test")
    
    config := New()
    
    assert.Equal(t, "9090", config.Server.Port)
    assert.Equal(t, "postgres://test:test@localhost:5432/test", config.Database.URL)
}

func TestConfigDefaults(t *testing.T) {
    // Test des valeurs par défaut
    config := New()
    
    assert.Equal(t, "8080", config.Server.Port)
    assert.Equal(t, "development", config.Server.Environment)
    assert.Equal(t, 100, config.Database.MaxOpenConns)
}
```

### Tests de Validation

```go
func TestConfigValidation(t *testing.T) {
    config := &Config{
        Server: ServerConfig{Environment: "production"},
        JWT: JWTConfig{Secret: "short"},
    }
    
    err := config.Validate()
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "JWT secret too short")
}
```

## 🔧 Maintenance

### Logs de Configuration

```go
func (c *Config) LogSummary() {
    log.Printf("Configuration chargée:")
    log.Printf("  - Environnement: %s", c.Server.Environment)
    log.Printf("  - Port: %s", c.Server.Port)
    log.Printf("  - Base de données: %s:%s", c.Database.Host, c.Database.Port)
    log.Printf("  - Redis: %s:%s", c.Redis.Host, c.Redis.Port)
    log.Printf("  - NATS: %s", c.NATS.URL)
}
```

### Monitoring de Configuration

```go
func (c *Config) GetMetrics() map[string]interface{} {
    return map[string]interface{}{
        "environment": c.Server.Environment,
        "database_pool_size": c.Database.MaxOpenConns,
        "redis_pool_size": c.Redis.PoolSize,
        "cache_levels_enabled": []bool{
            c.Cache.EnableLevel1,
            c.Cache.EnableLevel2,
            c.Cache.EnableLevel3,
        },
        "queue_workers": c.Queue.MaxWorkers,
    }
}
```

## 📚 Ressources Associées

- [Point d'Entrée Principal](./cmd-server-main.md)
- [Base de Données](../../database/schema.md)
- [API REST](../../api/backend-api.md)
- [Cache Redis](../cache/redis.md)
- [NATS Messaging](../messaging/nats.md)
- [Guide de Déploiement](../../deployment/deployment-guide.md) 