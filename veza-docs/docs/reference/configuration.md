---
id: configuration
title: Configuration
sidebar_label: Configuration
---

# ⚙️ Configuration - Veza

## 📋 Vue d'ensemble

Ce guide détaille la configuration de la plateforme Veza pour tous les environnements.

## 🔧 Variables d'Environnement

### Configuration Backend API
```bash
# Base de données
DB_HOST=localhost
DB_PORT=5432
DB_NAME=veza
DB_USER=veza_user
DB_PASSWORD=secure_password

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=redis_password

# JWT
JWT_SECRET=your_jwt_secret_key
JWT_EXPIRATION=3600

# OAuth2
GOOGLE_CLIENT_ID=google_client_id
GOOGLE_CLIENT_SECRET=google_client_secret
GITHUB_CLIENT_ID=github_client_id
GITHUB_CLIENT_SECRET=github_client_secret

# Email
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your_email@gmail.com
SMTP_PASSWORD=your_email_password

# Fichiers
UPLOAD_PATH=/var/uploads
MAX_FILE_SIZE=10485760
ALLOWED_FILE_TYPES=image/jpeg,image/png,audio/mpeg

# Monitoring
PROMETHEUS_ENABLED=true
GRAFANA_ENABLED=true
LOG_LEVEL=info
```

### Configuration Chat Server
```bash
# WebSocket
WS_PORT=8081
WS_HOST=0.0.0.0
WS_PATH=/ws

# Rate Limiting
RATE_LIMIT_MESSAGES=100
RATE_LIMIT_WINDOW=3600

# Modération
MODERATION_ENABLED=true
PROFANITY_FILTER=true
SPAM_DETECTION=true

# Performance
MAX_CONNECTIONS=10000
MESSAGE_QUEUE_SIZE=1000
```

### Configuration Stream Server
```bash
# Streaming
STREAM_PORT=8082
STREAM_HOST=0.0.0.0
STREAM_PATH=/stream

# Audio
AUDIO_BITRATE=128000
AUDIO_CODEC=aac
AUDIO_SAMPLE_RATE=44100

# Cache
CACHE_ENABLED=true
CACHE_TTL=3600
CACHE_MAX_SIZE=1073741824

# Transcoding
TRANSCODING_ENABLED=true
TRANSCODING_THREADS=4
TRANSCODING_QUALITY=high
```

## 📁 Fichiers de Configuration

### Configuration Docker Compose
```yaml
# docker-compose.yml
version: '3.8'

services:
  backend-api:
    image: veza/backend-api:latest
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=postgres
      - REDIS_HOST=redis
    depends_on:
      - postgres
      - redis
    volumes:
      - ./uploads:/var/uploads
    networks:
      - veza-network

  chat-server:
    image: veza/chat-server:latest
    ports:
      - "8081:8081"
    environment:
      - REDIS_HOST=redis
      - BACKEND_API_URL=http://backend-api:8080
    depends_on:
      - redis
      - backend-api
    networks:
      - veza-network

  stream-server:
    image: veza/stream-server:latest
    ports:
      - "8082:8082"
    environment:
      - REDIS_HOST=redis
      - BACKEND_API_URL=http://backend-api:8080
    depends_on:
      - redis
      - backend-api
    volumes:
      - ./audio:/var/audio
    networks:
      - veza-network

  postgres:
    image: postgres:15
    environment:
      - POSTGRES_DB=veza
      - POSTGRES_USER=veza_user
      - POSTGRES_PASSWORD=secure_password
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - veza-network

  redis:
    image: redis:7-alpine
    command: redis-server --requirepass redis_password
    volumes:
      - redis_data:/data
    networks:
      - veza-network

  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml
    networks:
      - veza-network

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana_data:/var/lib/grafana
    networks:
      - veza-network

volumes:
  postgres_data:
  redis_data:
  grafana_data:

networks:
  veza-network:
    driver: bridge
```

### Configuration Kubernetes
```yaml
# k8s/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: veza-config
  namespace: veza
data:
  # Backend API
  DB_HOST: "postgres-service"
  DB_PORT: "5432"
  DB_NAME: "veza"
  REDIS_HOST: "redis-service"
  REDIS_PORT: "6379"
  
  # JWT
  JWT_SECRET: "your_jwt_secret_key"
  JWT_EXPIRATION: "3600"
  
  # OAuth2
  GOOGLE_CLIENT_ID: "google_client_id"
  GOOGLE_CLIENT_SECRET: "google_client_secret"
  
  # Email
  SMTP_HOST: "smtp.gmail.com"
  SMTP_PORT: "587"
  
  # Monitoring
  PROMETHEUS_ENABLED: "true"
  LOG_LEVEL: "info"
```

```yaml
# k8s/secrets.yaml
apiVersion: v1
kind: Secret
metadata:
  name: veza-secrets
  namespace: veza
type: Opaque
data:
  # Base de données
  DB_USER: dmV6YV91c2Vy
  DB_PASSWORD: c2VjdXJlX3Bhc3N3b3Jk
  
  # Redis
  REDIS_PASSWORD: cmVkaXNfcGFzc3dvcmQ=
  
  # Email
  SMTP_USER: eW91cl9lbWFpbEBnbWFpbC5jb20=
  SMTP_PASSWORD: eW91cl9lbWFpbF9wYXNzd29yZA==
  
  # OAuth2
  GOOGLE_CLIENT_SECRET: Z29vZ2xlX2NsaWVudF9zZWNyZXQ=
  GITHUB_CLIENT_SECRET: Z2l0aHViX2NsaWVudF9zZWNyZXQ=
```

## 🔒 Configuration de Sécurité

### Configuration SSL/TLS
```yaml
# nginx/ssl.conf
ssl_protocols TLSv1.2 TLSv1.3;
ssl_ciphers ECDHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384;
ssl_prefer_server_ciphers off;
ssl_session_cache shared:SSL:10m;
ssl_session_timeout 10m;
ssl_stapling on;
ssl_stapling_verify on;
```

### Configuration Firewall
```bash
# ufw/firewall.conf
# Autoriser SSH
ufw allow 22/tcp

# Autoriser HTTP/HTTPS
ufw allow 80/tcp
ufw allow 443/tcp

# Autoriser les ports d'application
ufw allow 8080/tcp  # Backend API
ufw allow 8081/tcp  # Chat Server
ufw allow 8082/tcp  # Stream Server

# Autoriser les ports de monitoring
ufw allow 9090/tcp  # Prometheus
ufw allow 3000/tcp  # Grafana

# Refuser tout le reste
ufw default deny incoming
```

## 📊 Configuration de Monitoring

### Configuration Prometheus
```yaml
# monitoring/prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - "alerts.yml"

scrape_configs:
  - job_name: 'veza-backend-api'
    static_configs:
      - targets: ['backend-api:8080']
    metrics_path: '/metrics'
    
  - job_name: 'veza-chat-server'
    static_configs:
      - targets: ['chat-server:8081']
    metrics_path: '/metrics'
    
  - job_name: 'veza-stream-server'
    static_configs:
      - targets: ['stream-server:8082']
    metrics_path: '/metrics'
    
  - job_name: 'postgres'
    static_configs:
      - targets: ['postgres-exporter:9187']
    metrics_path: '/metrics'
    
  - job_name: 'redis'
    static_configs:
      - targets: ['redis-exporter:9121']
    metrics_path: '/metrics'
```

### Configuration Grafana
```yaml
# monitoring/grafana/dashboards.yml
apiVersion: 1

providers:
  - name: 'Veza Dashboards'
    orgId: 1
    folder: ''
    type: file
    disableDeletion: false
    updateIntervalSeconds: 10
    allowUiUpdates: true
    options:
      path: /var/lib/grafana/dashboards
```

## 📚 Ressources

- [Guide de Déploiement](../deployment/deployment-guide.md)
- [Variables d'Environnement](../deployment/environment-variables.md)
- [Monitoring](../monitoring/README.md)

---

**Dernière mise à jour** : $(date)
**Version** : 1.0.0 