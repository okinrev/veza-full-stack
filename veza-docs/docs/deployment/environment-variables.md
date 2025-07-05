---
id: environment-variables
title: Variables d'Environnement
sidebar_label: Variables d'Environnement
description: Guide complet des variables d'environnement pour la plateforme Veza
---

# Variables d'Environnement - Veza Platform

Ce guide détaille toutes les variables d'environnement utilisées dans la plateforme Veza.

## Vue d'ensemble

### Structure des variables
```bash
# Environnements
VEZA_ENV=production|staging|development

# Services
VEZA_SERVICE_NAME=backend-api|chat-server|stream-server

# Base de données
DATABASE_URL=postgres://user:pass@host:port/db
DATABASE_MAX_CONNECTIONS=25
DATABASE_TIMEOUT=30s

# Cache
REDIS_URL=redis://host:port
REDIS_PASSWORD=secret
REDIS_DB=0

# Sécurité
JWT_SECRET=your-secret-key
API_KEY=your-api-key
CORS_ORIGINS=https://veza.com,https://app.veza.com

# Monitoring
PROMETHEUS_ENABLED=true
JAEGER_ENDPOINT=http://jaeger:14268/api/traces
LOG_LEVEL=info|debug|warn|error
```

## Variables par service

### Backend API (veza-backend-api)

#### Configuration de base
```bash
# Environnement
ENV=production
SERVICE_NAME=backend-api
SERVICE_VERSION=1.0.0

# Serveur
PORT=8080
HOST=0.0.0.0
READ_TIMEOUT=30s
WRITE_TIMEOUT=30s
IDLE_TIMEOUT=60s
```

#### Base de données
```bash
# PostgreSQL
DATABASE_URL=postgres://veza_user:password@postgres:5432/veza_prod?sslmode=require
DATABASE_MAX_OPEN_CONNS=25
DATABASE_MAX_IDLE_CONNS=5
DATABASE_CONN_MAX_LIFETIME=5m
DATABASE_TIMEOUT=30s

# Migrations
MIGRATION_PATH=./migrations
MIGRATION_AUTO_RUN=true
```

#### Cache et sessions
```bash
# Redis
REDIS_URL=redis://redis:6379
REDIS_PASSWORD=redis-password
REDIS_DB=0
REDIS_POOL_SIZE=10
REDIS_DIAL_TIMEOUT=5s
REDIS_READ_TIMEOUT=3s
REDIS_WRITE_TIMEOUT=3s

# Sessions
SESSION_SECRET=session-secret-key
SESSION_TIMEOUT=24h
SESSION_CLEANUP_INTERVAL=1h
```

#### Authentification et sécurité
```bash
# JWT
JWT_SECRET=your-super-secret-jwt-key
JWT_EXPIRATION=24h
JWT_REFRESH_EXPIRATION=168h
JWT_ISSUER=veza-platform

# OAuth2
OAUTH_GOOGLE_CLIENT_ID=google-client-id
OAUTH_GOOGLE_CLIENT_SECRET=google-client-secret
OAUTH_GITHUB_CLIENT_ID=github-client-id
OAUTH_GITHUB_CLIENT_SECRET=github-client-secret

# API Keys
API_KEY=your-api-key
API_KEY_HEADER=X-API-Key
API_RATE_LIMIT=1000
API_RATE_LIMIT_WINDOW=1m
```

#### Services externes
```bash
# Email
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM=noreply@veza.com

# SMS (Twilio)
TWILIO_ACCOUNT_SID=your-account-sid
TWILIO_AUTH_TOKEN=your-auth-token
TWILIO_PHONE_NUMBER=+1234567890

# Storage (AWS S3)
AWS_ACCESS_KEY_ID=your-access-key
AWS_SECRET_ACCESS_KEY=your-secret-key
AWS_REGION=eu-west-1
AWS_S3_BUCKET=veza-media
AWS_S3_ENDPOINT=https://s3.eu-west-1.amazonaws.com

# CDN
CDN_URL=https://cdn.veza.com
CDN_API_KEY=your-cdn-api-key
```

#### Monitoring et observabilité
```bash
# Prometheus
PROMETHEUS_ENABLED=true
PROMETHEUS_PORT=9090
PROMETHEUS_PATH=/metrics

# Jaeger (Tracing)
JAEGER_ENABLED=true
JAEGER_ENDPOINT=http://jaeger:14268/api/traces
JAEGER_SERVICE_NAME=veza-backend-api

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
LOG_OUTPUT=stdout
LOG_FILE_PATH=/var/log/veza/backend-api.log

# Health checks
HEALTH_CHECK_ENABLED=true
HEALTH_CHECK_PORT=8081
HEALTH_CHECK_PATH=/health
```

### Chat Server (veza-chat-server)

#### Configuration de base
```bash
# Environnement
ENV=production
SERVICE_NAME=chat-server
SERVICE_VERSION=1.0.0

# Serveur WebSocket
WS_PORT=8081
WS_HOST=0.0.0.0
WS_READ_BUFFER_SIZE=1024
WS_WRITE_BUFFER_SIZE=1024
WS_MAX_CONNECTIONS=10000
```

#### Base de données et cache
```bash
# PostgreSQL
DATABASE_URL=postgres://veza_user:password@postgres:5432/veza_prod?sslmode=require
DATABASE_MAX_CONNECTIONS=20
DATABASE_TIMEOUT=30s

# Redis
REDIS_URL=redis://redis:6379
REDIS_PASSWORD=redis-password
REDIS_DB=1
REDIS_POOL_SIZE=20
REDIS_PUBSUB_CHANNELS=chat_messages,user_status,room_events
```

#### Authentification
```bash
# JWT
JWT_SECRET=your-super-secret-jwt-key
JWT_EXPIRATION=24h
JWT_ISSUER=veza-platform

# API Keys
API_KEY=your-api-key
API_KEY_HEADER=X-API-Key
```

#### Modération et sécurité
```bash
# Modération
MODERATION_ENABLED=true
MODERATION_API_KEY=your-moderation-api-key
MODERATION_ENDPOINT=https://api.moderation.com/v1/check

# Rate limiting
RATE_LIMIT_ENABLED=true
RATE_LIMIT_MESSAGES_PER_MINUTE=60
RATE_LIMIT_CONNECTIONS_PER_IP=10
RATE_LIMIT_BURST_SIZE=100
```

#### Monitoring
```bash
# Prometheus
PROMETHEUS_ENABLED=true
PROMETHEUS_PORT=9091
PROMETHEUS_PATH=/metrics

# Jaeger
JAEGER_ENABLED=true
JAEGER_ENDPOINT=http://jaeger:14268/api/traces
JAEGER_SERVICE_NAME=veza-chat-server

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
LOG_OUTPUT=stdout
LOG_FILE_PATH=/var/log/veza/chat-server.log
```

### Stream Server (veza-stream-server)

#### Configuration de base
```bash
# Environnement
ENV=production
SERVICE_NAME=stream-server
SERVICE_VERSION=1.0.0

# Serveur
HTTP_PORT=8082
HTTP_HOST=0.0.0.0
WS_PORT=8083
WS_HOST=0.0.0.0
```

#### Audio et streaming
```bash
# Audio processing
AUDIO_SAMPLE_RATE=44100
AUDIO_CHANNELS=2
AUDIO_BITRATE=128k
AUDIO_FORMAT=mp3|aac|flac

# Streaming
STREAM_CHUNK_SIZE=4096
STREAM_BUFFER_SIZE=8192
STREAM_TIMEOUT=30s
STREAM_MAX_CONNECTIONS=1000

# Transcoding
TRANSCODER_ENABLED=true
TRANSCODER_THREADS=4
TRANSCODER_QUALITY=high
TRANSCODER_FORMATS=mp3,aac,flac
```

#### Storage et CDN
```bash
# AWS S3
AWS_ACCESS_KEY_ID=your-access-key
AWS_SECRET_ACCESS_KEY=your-secret-key
AWS_REGION=eu-west-1
AWS_S3_BUCKET=veza-audio
AWS_S3_ENDPOINT=https://s3.eu-west-1.amazonaws.com

# CDN
CDN_URL=https://cdn.veza.com
CDN_API_KEY=your-cdn-api-key
CDN_CACHE_TTL=3600
```

#### Monitoring
```bash
# Prometheus
PROMETHEUS_ENABLED=true
PROMETHEUS_PORT=9092
PROMETHEUS_PATH=/metrics

# Jaeger
JAEGER_ENABLED=true
JAEGER_ENDPOINT=http://jaeger:14268/api/traces
JAEGER_SERVICE_NAME=veza-stream-server

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
LOG_OUTPUT=stdout
LOG_FILE_PATH=/var/log/veza/stream-server.log
```

## Configuration par environnement

### Développement
```bash
# .env.development
ENV=development
LOG_LEVEL=debug
DATABASE_URL=postgres://user:pass@localhost:5432/veza_dev
REDIS_URL=redis://localhost:6379
JWT_SECRET=dev-secret-key
API_KEY=dev-api-key
PROMETHEUS_ENABLED=false
JAEGER_ENABLED=false
```

### Staging
```bash
# .env.staging
ENV=staging
LOG_LEVEL=info
DATABASE_URL=postgres://user:pass@postgres-staging:5432/veza_staging
REDIS_URL=redis://redis-staging:6379
JWT_SECRET=staging-secret-key
API_KEY=staging-api-key
PROMETHEUS_ENABLED=true
JAEGER_ENABLED=true
```

### Production
```bash
# .env.production
ENV=production
LOG_LEVEL=warn
DATABASE_URL=postgres://user:pass@postgres-prod:5432/veza_prod?sslmode=require
REDIS_URL=redis://redis-prod:6379
JWT_SECRET=production-super-secret-key
API_KEY=production-api-key
PROMETHEUS_ENABLED=true
JAEGER_ENABLED=true
```

## Gestion des secrets

### Kubernetes Secrets
```yaml
# k8s/secrets.yaml
apiVersion: v1
kind: Secret
metadata:
  name: veza-secrets
  namespace: veza-prod
type: Opaque
data:
  JWT_SECRET: <base64-encoded-secret>
  DATABASE_PASSWORD: <base64-encoded-password>
  REDIS_PASSWORD: <base64-encoded-password>
  API_KEY: <base64-encoded-api-key>
  AWS_ACCESS_KEY_ID: <base64-encoded-access-key>
  AWS_SECRET_ACCESS_KEY: <base64-encoded-secret-key>
```

### Docker Compose
```yaml
# docker-compose.yml
version: '3.8'
services:
  backend-api:
    environment:
      - JWT_SECRET=${JWT_SECRET}
      - DATABASE_URL=${DATABASE_URL}
      - REDIS_URL=${REDIS_URL}
    env_file:
      - .env.production
```

### Helm Values
```yaml
# helm/veza/values.yaml
global:
  environment: production

backend-api:
  env:
    JWT_SECRET: "{{ .Values.secrets.jwtSecret }}"
    DATABASE_URL: "{{ .Values.database.url }}"
    REDIS_URL: "{{ .Values.redis.url }}"

secrets:
  jwtSecret: "your-secret-key"
  databasePassword: "your-db-password"
  redisPassword: "your-redis-password"
```

## Validation des variables

### Script de validation
```bash
#!/bin/bash
# scripts/validate-env.sh

required_vars=(
    "ENV"
    "DATABASE_URL"
    "REDIS_URL"
    "JWT_SECRET"
    "API_KEY"
)

missing_vars=()

for var in "${required_vars[@]}"; do
    if [ -z "${!var}" ]; then
        missing_vars+=("$var")
    fi
done

if [ ${#missing_vars[@]} -ne 0 ]; then
    echo "❌ Variables d'environnement manquantes:"
    printf '%s\n' "${missing_vars[@]}"
    exit 1
fi

echo "✅ Toutes les variables d'environnement requises sont définies"
```

### Validation Go
```go
// internal/config/validator.go
package config

import (
    "fmt"
    "os"
)

type Validator struct {
    required []string
}

func NewValidator() *Validator {
    return &Validator{
        required: []string{
            "ENV",
            "DATABASE_URL",
            "REDIS_URL",
            "JWT_SECRET",
            "API_KEY",
        },
    }
}

func (v *Validator) Validate() error {
    for _, env := range v.required {
        if os.Getenv(env) == "" {
            return fmt.Errorf("variable d'environnement requise manquante: %s", env)
        }
    }
    return nil
}
```

## Bonnes pratiques

### 1. Sécurité
```bash
# ✅ Bon - Utilisation de secrets
JWT_SECRET=${JWT_SECRET}
DATABASE_PASSWORD=${DATABASE_PASSWORD}

# ❌ Mauvais - Secrets en dur
JWT_SECRET=my-secret-key
DATABASE_PASSWORD=password123
```

### 2. Validation
```bash
# ✅ Bon - Validation des variables
if [ -z "$DATABASE_URL" ]; then
    echo "❌ DATABASE_URL est requis"
    exit 1
fi

# ❌ Mauvais - Pas de validation
# Utilisation directe sans vérification
```

### 3. Documentation
```bash
# ✅ Bon - Documentation des variables
# DATABASE_URL: URL de connexion PostgreSQL
# JWT_SECRET: Clé secrète pour signer les JWT
# API_KEY: Clé API pour l'authentification

# ❌ Mauvais - Pas de documentation
# Variables sans explication
```

### 4. Environnements
```bash
# ✅ Bon - Séparation des environnements
.env.development
.env.staging
.env.production

# ❌ Mauvais - Un seul fichier
.env
```

## Outils utiles

### Validation automatique
```bash
# Script de validation
make validate-env

# Validation dans le CI/CD
npm run validate-env
```

### Documentation automatique
```bash
# Génération de la documentation
make docs-env

# Export des variables
make export-env
```

## Conclusion

La gestion des variables d'environnement est cruciale pour la sécurité et la configuration de la plateforme Veza. Suivez ces bonnes pratiques pour maintenir un environnement sécurisé et bien configuré.

### Ressources supplémentaires
- [Guide de déploiement](./deployment-guide.md)
- [Guide de sécurité](../security/README.md)
- [Guide de monitoring](../monitoring/README.md) 