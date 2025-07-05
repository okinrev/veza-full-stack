---
id: advanced-configuration
title: Configuration Avancée - Veza Platform
sidebar_label: Configuration Avancée
---

# Configuration Avancée - Veza Platform

> **Guide complet pour la configuration avancée de la plateforme Veza**

## Vue d'ensemble

Ce guide couvre les configurations avancées pour optimiser les performances, la sécurité et la scalabilité de la plateforme Veza.

## Configuration de l'Environnement

### Variables d'Environnement Avancées

```bash
# Configuration de base
VEZA_ENV=production
VEZA_LOG_LEVEL=info
VEZA_TIMEZONE=UTC

# Base de données
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_DB=veza_production
POSTGRES_USER=veza_user
POSTGRES_PASSWORD=secure_password
POSTGRES_SSL_MODE=require
POSTGRES_MAX_CONNECTIONS=100
POSTGRES_IDLE_CONNECTIONS=10

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=redis_password
REDIS_DB=0
REDIS_POOL_SIZE=20

# NATS
NATS_URL=nats://localhost:4222
NATS_CLUSTER_ID=veza_cluster
NATS_CLIENT_ID=veza_client

# JWT
JWT_SECRET=your_jwt_secret_key
JWT_EXPIRATION=24h
JWT_REFRESH_EXPIRATION=168h

# Streaming
STREAM_BUCKET=veza-streams
STREAM_REGION=us-east-1
STREAM_CDN_URL=https://cdn.veza.com

# Analytics
ANALYTICS_ENABLED=true
ANALYTICS_BATCH_SIZE=100
ANALYTICS_FLUSH_INTERVAL=30s

# Monitoring
PROMETHEUS_ENABLED=true
PROMETHEUS_PORT=9090
JAEGER_ENABLED=true
JAEGER_ENDPOINT=http://localhost:14268

# Security
CORS_ORIGINS=https://veza.com,https://app.veza.com
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=1m

# Moderation
MODERATION_ENABLED=true
MODERATION_AI_ENDPOINT=https://api.openai.com/v1
MODERATION_AI_KEY=your_openai_key
MODERATION_THRESHOLD=0.7
```

### Configuration par Environnement

#### Développement

```yaml
# config/development.yaml
environment: development
debug: true
log_level: debug
database:
  host: localhost
  port: 5432
  name: veza_dev
  ssl_mode: disable
redis:
  host: localhost
  port: 6379
  password: ""
streaming:
  bucket: veza-dev-streams
  cdn_enabled: false
monitoring:
  prometheus_enabled: false
  jaeger_enabled: false
security:
  cors_origins: ["http://localhost:3000"]
  rate_limit_requests: 1000
```

#### Production

```yaml
# config/production.yaml
environment: production
debug: false
log_level: info
database:
  host: veza-db.production.com
  port: 5432
  name: veza_production
  ssl_mode: require
  max_connections: 100
  idle_connections: 10
redis:
  host: veza-redis.production.com
  port: 6379
  password: "${REDIS_PASSWORD}"
  pool_size: 50
streaming:
  bucket: veza-production-streams
  cdn_enabled: true
  cdn_url: "https://cdn.veza.com"
monitoring:
  prometheus_enabled: true
  jaeger_enabled: true
security:
  cors_origins: ["https://veza.com", "https://app.veza.com"]
  rate_limit_requests: 100
```

## Configuration de la Base de Données

### Optimisation PostgreSQL

```sql
-- Configuration des connexions
ALTER SYSTEM SET max_connections = 200;
ALTER SYSTEM SET shared_buffers = '256MB';
ALTER SYSTEM SET effective_cache_size = '1GB';
ALTER SYSTEM SET maintenance_work_mem = '64MB';
ALTER SYSTEM SET checkpoint_completion_target = 0.9;
ALTER SYSTEM SET wal_buffers = '16MB';
ALTER SYSTEM SET default_statistics_target = 100;

-- Configuration pour les requêtes
ALTER SYSTEM SET random_page_cost = 1.1;
ALTER SYSTEM SET effective_io_concurrency = 200;
ALTER SYSTEM SET work_mem = '4MB';

-- Configuration du logging
ALTER SYSTEM SET log_statement = 'all';
ALTER SYSTEM SET log_min_duration_statement = 1000;
ALTER SYSTEM SET log_checkpoints = on;
ALTER SYSTEM SET log_connections = on;
ALTER SYSTEM SET log_disconnections = on;

-- Recharger la configuration
SELECT pg_reload_conf();
```

### Index Optimisés

```sql
-- Index pour les utilisateurs
CREATE INDEX CONCURRENTLY idx_users_email ON users(email);
CREATE INDEX CONCURRENTLY idx_users_username ON users(username);
CREATE INDEX CONCURRENTLY idx_users_created_at ON users(created_at);

-- Index pour les messages
CREATE INDEX CONCURRENTLY idx_messages_room_id ON messages(room_id);
CREATE INDEX CONCURRENTLY idx_messages_created_at ON messages(created_at);
CREATE INDEX CONCURRENTLY idx_messages_user_id ON messages(user_id);

-- Index pour les streams
CREATE INDEX CONCURRENTLY idx_streams_user_id ON streams(user_id);
CREATE INDEX CONCURRENTLY idx_streams_status ON streams(status);
CREATE INDEX CONCURRENTLY idx_streams_created_at ON streams(created_at);

-- Index composites
CREATE INDEX CONCURRENTLY idx_messages_room_created ON messages(room_id, created_at);
CREATE INDEX CONCURRENTLY idx_streams_user_status ON streams(user_id, status);
```

## Configuration Redis

### Optimisation Redis

```redis
# Configuration Redis pour la production
# redis.conf

# Mémoire
maxmemory 2gb
maxmemory-policy allkeys-lru

# Persistance
save 900 1
save 300 10
save 60 10000
appendonly yes
appendfsync everysec

# Réseau
tcp-keepalive 300
timeout 0

# Logging
loglevel notice
logfile /var/log/redis/redis.log

# Sécurité
requirepass your_redis_password
```

### Scripts de Monitoring Redis

```bash
#!/bin/bash
# scripts/monitor_redis.sh

REDIS_HOST="localhost"
REDIS_PORT="6379"
REDIS_PASSWORD="your_password"

# Métriques de base
echo "=== Redis Metrics ==="
redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD info memory
redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD info stats
redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD info clients

# Clés par type
echo "=== Key Types ==="
redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD --scan --pattern "*" | wc -l
redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD --scan --pattern "user:*" | wc -l
redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD --scan --pattern "session:*" | wc -l
redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD --scan --pattern "chat:*" | wc -l
```

## Configuration NATS

### Configuration NATS Server

```yaml
# nats-server.conf
port: 4222
http_port: 8222
cluster {
  port: 6222
  listen: 0.0.0.0:6222
  routes: [
    "nats://nats-1:6222"
    "nats://nats-2:6222"
    "nats://nats-3:6222"
  ]
}

jetstream {
  store_dir: "/data/jetstream"
  max_mem: 1GB
  max_file: 10GB
}

authorization {
  users: [
    {
      nkey: UDXU4RCSJNZOIQHZNWXHXORDPRTGNJAHAHFRONTANEFL
      permissions: {
        publish: {
          allow: ["veza.*"]
        }
        subscribe: {
          allow: ["veza.*"]
        }
      }
    }
  ]
}
```

### Scripts de Monitoring NATS

```bash
#!/bin/bash
# scripts/monitor_nats.sh

NATS_URL="nats://localhost:4222"

echo "=== NATS Server Info ==="
curl -s http://localhost:8222/varz | jq '.'

echo "=== NATS Connections ==="
curl -s http://localhost:8222/connz | jq '.'

echo "=== NATS Subscriptions ==="
curl -s http://localhost:8222/subsz | jq '.'

echo "=== JetStream Info ==="
curl -s http://localhost:8222/jsz | jq '.'
```

## Configuration de Sécurité

### Configuration CORS

```go
// internal/middleware/cors.go
package middleware

import (
    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
    "time"
)

func CORSMiddleware() gin.HandlerFunc {
    return cors.New(cors.Config{
        AllowOrigins:     []string{"https://veza.com", "https://app.veza.com"},
        AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    })
}
```

### Rate Limiting

```go
// internal/middleware/rate_limiter.go
package middleware

import (
    "github.com/gin-gonic/gin"
    "golang.org/x/time/rate"
    "sync"
    "time"
)

type IPRateLimiter struct {
    ips    map[string]*rate.Limiter
    mu     *sync.RWMutex
    rate   rate.Limit
    burst  int
}

func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
    return &IPRateLimiter{
        ips:   make(map[string]*rate.Limiter),
        mu:    &sync.RWMutex{},
        rate:  r,
        burst: b,
    }
}

func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
    i.mu.Lock()
    defer i.mu.Unlock()

    limiter, exists := i.ips[ip]
    if !exists {
        limiter = rate.NewLimiter(i.rate, i.burst)
        i.ips[ip] = limiter
    }

    return limiter
}

func RateLimiter() gin.HandlerFunc {
    limiter := NewIPRateLimiter(rate.Limit(100), 200)
    
    return func(c *gin.Context) {
        ip := c.ClientIP()
        if !limiter.GetLimiter(ip).Allow() {
            c.JSON(429, gin.H{"error": "Too many requests"})
            c.Abort()
            return
        }
        c.Next()
    }
}
```

## Configuration de Monitoring

### Prometheus Configuration

```yaml
# prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - "veza_rules.yml"

scrape_configs:
  - job_name: 'veza-backend-api'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
    scrape_interval: 5s

  - job_name: 'veza-chat-server'
    static_configs:
      - targets: ['localhost:8081']
    metrics_path: '/metrics'
    scrape_interval: 5s

  - job_name: 'veza-stream-server'
    static_configs:
      - targets: ['localhost:8082']
    metrics_path: '/metrics'
    scrape_interval: 5s

  - job_name: 'postgres'
    static_configs:
      - targets: ['localhost:9187']

  - job_name: 'redis'
    static_configs:
      - targets: ['localhost:9121']
```

### Règles d'Alerte

```yaml
# veza_rules.yml
groups:
  - name: veza_alerts
    rules:
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.1
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "High error rate detected"
          description: "Error rate is {{ $value }} errors per second"

      - alert: HighResponseTime
        expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 1
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "High response time detected"
          description: "95th percentile response time is {{ $value }} seconds"

      - alert: DatabaseConnectionsHigh
        expr: postgres_exporter_postgresql_connections > 80
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "High database connections"
          description: "Database has {{ $value }} active connections"

      - alert: RedisMemoryHigh
        expr: redis_memory_used_bytes / redis_memory_max_bytes > 0.8
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "High Redis memory usage"
          description: "Redis memory usage is {{ $value | humanizePercentage }}"
```

## Configuration de Logging

### Configuration des Logs

```go
// internal/logger/logger.go
package logger

import (
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
    "os"
)

func NewLogger(level string) *zap.Logger {
    var logLevel zapcore.Level
    switch level {
    case "debug":
        logLevel = zapcore.DebugLevel
    case "info":
        logLevel = zapcore.InfoLevel
    case "warn":
        logLevel = zapcore.WarnLevel
    case "error":
        logLevel = zapcore.ErrorLevel
    default:
        logLevel = zapcore.InfoLevel
    }

    config := zap.NewProductionConfig()
    config.Level = zap.NewAtomicLevelAt(logLevel)
    config.OutputPaths = []string{"stdout", "/var/log/veza/application.log"}
    config.ErrorOutputPaths = []string{"stderr", "/var/log/veza/error.log"}
    config.EncoderConfig.TimeKey = "timestamp"
    config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
    config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

    logger, err := config.Build()
    if err != nil {
        panic(err)
    }

    return logger
}
```

### Rotation des Logs

```bash
#!/bin/bash
# scripts/log_rotation.sh

LOG_DIR="/var/log/veza"
MAX_SIZE="100M"
MAX_FILES=10

# Rotation des logs d'application
logrotate -f << EOF
$LOG_DIR/application.log {
    daily
    missingok
    rotate 10
    compress
    delaycompress
    notifempty
    create 644 veza veza
    postrotate
        systemctl reload veza-backend-api
    endscript
}

$LOG_DIR/error.log {
    daily
    missingok
    rotate 10
    compress
    delaycompress
    notifempty
    create 644 veza veza
    postrotate
        systemctl reload veza-backend-api
    endscript
}
EOF
```

## Configuration de Performance

### Optimisation Go

```go
// cmd/server/main.go
package main

import (
    "runtime"
    "github.com/gin-gonic/gin"
)

func main() {
    // Optimisation du runtime Go
    runtime.GOMAXPROCS(runtime.NumCPU())
    
    // Configuration Gin pour la production
    gin.SetMode(gin.ReleaseMode)
    
    // Configuration des middlewares de performance
    router := gin.New()
    router.Use(gin.Recovery())
    router.Use(gin.Logger())
    
    // ... reste de la configuration
}
```

### Optimisation Rust

```rust
// Cargo.toml
[profile.release]
opt-level = 3
lto = true
codegen-units = 1
panic = "abort"
strip = true

[profile.release.build-override]
opt-level = 3
codegen-units = 1
```

## Scripts de Maintenance

### Script de Backup

```bash
#!/bin/bash
# scripts/backup.sh

BACKUP_DIR="/backups/veza"
DATE=$(date +%Y%m%d_%H%M%S)
DB_NAME="veza_production"

# Backup PostgreSQL
pg_dump -h localhost -U veza_user -d $DB_NAME | gzip > $BACKUP_DIR/db_$DATE.sql.gz

# Backup Redis
redis-cli --rdb $BACKUP_DIR/redis_$DATE.rdb

# Backup fichiers de configuration
tar -czf $BACKUP_DIR/config_$DATE.tar.gz /etc/veza/

# Nettoyage des anciens backups (garde 30 jours)
find $BACKUP_DIR -name "*.gz" -mtime +30 -delete
find $BACKUP_DIR -name "*.rdb" -mtime +30 -delete
find $BACKUP_DIR -name "*.tar.gz" -mtime +30 -delete

echo "Backup completed: $DATE"
```

### Script de Health Check

```bash
#!/bin/bash
# scripts/health_check.sh

# Vérification des services
services=("veza-backend-api" "veza-chat-server" "veza-stream-server")

for service in "${services[@]}"; do
    if systemctl is-active --quiet $service; then
        echo "$service is running"
    else
        echo "$service is not running"
        systemctl restart $service
    fi
done

# Vérification de la base de données
if pg_isready -h localhost -p 5432; then
    echo "PostgreSQL is ready"
else
    echo "PostgreSQL is not ready"
    exit 1
fi

# Vérification de Redis
if redis-cli ping > /dev/null 2>&1; then
    echo "Redis is ready"
else
    echo "Redis is not ready"
    exit 1
fi

# Vérification de NATS
if curl -s http://localhost:8222/healthz > /dev/null; then
    echo "NATS is ready"
else
    echo "NATS is not ready"
    exit 1
fi

echo "All services are healthy"
```

---

## 🔗 Liens croisés

- [Guide de Déploiement](../deployment/README.md)
- [Monitoring](../monitoring/README.md)
- [Sécurité](../security/README.md)
- [Tests](../testing/README.md)

---

## Pour aller plus loin

- [Architecture](../architecture/README.md)
- [API Reference](../api/README.md)
- [Database Schema](../database/README.md)
- [Troubleshooting](../troubleshooting/README.md) 