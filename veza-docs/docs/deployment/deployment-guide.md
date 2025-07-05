# 🚀 Guide de Déploiement - Veza

## 📋 Vue d'ensemble

Ce guide détaille le déploiement complet de la plateforme Veza en production, incluant tous les services (backend API, chat server, stream server) avec leurs dépendances.

## 🏗️ Architecture de Déploiement

```mermaid
graph TB
    subgraph "Load Balancer"
        HAProxy[HAProxy]
        SSL[SSL Termination]
    end
    
    subgraph "Application Servers"
        BackendAPI1[Backend API 1]
        BackendAPI2[Backend API 2]
        ChatServer1[Chat Server 1]
        ChatServer2[Chat Server 2]
        StreamServer1[Stream Server 1]
        StreamServer2[Stream Server 2]
    end
    
    subgraph "Database Layer"
        PostgreSQL1[(PostgreSQL Primary)]
        PostgreSQL2[(PostgreSQL Replica)]
        Redis1[(Redis Primary)]
        Redis2[(Redis Replica)]
    end
    
    subgraph "Message Queue"
        NATS1[NATS Server 1]
        NATS2[NATS Server 2]
        NATS3[NATS Server 3]
    end
    
    subgraph "Monitoring"
        Prometheus[Prometheus]
        Grafana[Grafana]
        AlertManager[Alert Manager]
    end
    
    subgraph "Storage"
        S3[S3 Storage]
        CDN[CDN]
    end
    
    HAProxy --> BackendAPI1
    HAProxy --> BackendAPI2
    HAProxy --> ChatServer1
    HAProxy --> ChatServer2
    HAProxy --> StreamServer1
    HAProxy --> StreamServer2
    
    BackendAPI1 --> PostgreSQL1
    BackendAPI2 --> PostgreSQL1
    ChatServer1 --> PostgreSQL2
    ChatServer2 --> PostgreSQL2
    
    BackendAPI1 --> Redis1
    BackendAPI2 --> Redis1
    ChatServer1 --> Redis2
    ChatServer2 --> Redis2
    
    BackendAPI1 --> NATS1
    ChatServer1 --> NATS2
    StreamServer1 --> NATS3
    
    BackendAPI1 --> Prometheus
    ChatServer1 --> Prometheus
    StreamServer1 --> Prometheus
    
    StreamServer1 --> S3
    StreamServer2 --> S3
    S3 --> CDN
```

## 🎯 Prérequis

### Infrastructure

- **Serveurs** : 6 serveurs minimum (2 par service)
- **CPU** : 4 cores minimum par serveur
- **RAM** : 8GB minimum par serveur
- **Storage** : 100GB SSD minimum par serveur
- **Network** : 1Gbps minimum

### Logiciels

- **Docker** : 20.10+
- **Docker Compose** : 2.0+
- **Kubernetes** : 1.24+ (optionnel)
- **PostgreSQL** : 15+
- **Redis** : 7+
- **NATS** : 2.9+

## 🔧 Configuration par Environnement

### Variables d'Environnement Globales

```bash
# Environnement
ENVIRONMENT=production
NODE_ENV=production

# Domaine
DOMAIN=veza.com
API_DOMAIN=api.veza.com
CHAT_DOMAIN=chat.veza.com
STREAM_DOMAIN=stream.veza.com

# SSL
SSL_CERT_PATH=/etc/ssl/certs/veza.com.crt
SSL_KEY_PATH=/etc/ssl/private/veza.com.key

# Monitoring
PROMETHEUS_URL=http://prometheus:9090
GRAFANA_URL=http://grafana:3000
```

### Backend API Configuration

```bash
# Backend API
BACKEND_PORT=8080
BACKEND_HOST=0.0.0.0

# Database
DATABASE_URL=postgresql://veza_user:secure_password@postgres-primary:5432/veza_prod
DATABASE_MAX_OPEN_CONNS=100
DATABASE_MAX_IDLE_CONNS=25
DATABASE_CONN_MAX_LIFETIME=300

# Redis
REDIS_URL=redis://redis-primary:6379
REDIS_POOL_SIZE=100
REDIS_PASSWORD=secure_redis_password

# JWT
JWT_ACCESS_SECRET=your-super-secure-production-jwt-secret-key-here
JWT_ACCESS_TTL=900
JWT_REFRESH_TTL=604800

# NATS
NATS_URL=nats://nats-1:4222,nats://nats-2:4222,nats://nats-3:4222
NATS_CLUSTER_ID=veza-cluster
NATS_CLIENT_ID=veza-backend-1

# Services
CHAT_SERVER_URL=grpc://chat-server-1:3001
STREAM_SERVER_URL=grpc://stream-server-1:3002
```

### Chat Server Configuration

```bash
# Chat Server
CHAT_PORT=3001
CHAT_HOST=0.0.0.0

# Database
CHAT_DATABASE_URL=postgresql://veza_user:secure_password@postgres-replica:5432/veza_prod
CHAT_DATABASE_MAX_CONNECTIONS=50

# Redis
CHAT_REDIS_URL=redis://redis-replica:6379
CHAT_REDIS_POOL_SIZE=50
CHAT_REDIS_PASSWORD=secure_redis_password

# JWT
CHAT_JWT_SECRET=your-super-secure-production-jwt-secret-key-here

# NATS
CHAT_NATS_URL=nats://nats-1:4222,nats://nats-2:4222,nats://nats-3:4222
CHAT_NATS_CLUSTER_ID=veza-cluster
CHAT_NATS_CLIENT_ID=veza-chat-1

# Performance
CHAT_MAX_CONNECTIONS=10000
CHAT_MESSAGE_QUEUE_SIZE=10000
CHAT_WORKER_THREADS=50
```

### Stream Server Configuration

```bash
# Stream Server
STREAM_PORT=3002
STREAM_HOST=0.0.0.0

# Audio
AUDIO_DIR=/var/audio
AUDIO_CACHE_SIZE_MB=2048
AUDIO_CACHE_TTL_SECONDS=7200

# Database
STREAM_DATABASE_URL=postgresql://veza_user:secure_password@postgres-replica:5432/veza_prod
STREAM_DATABASE_MAX_CONNECTIONS=25

# Redis
STREAM_REDIS_URL=redis://redis-replica:6379
STREAM_REDIS_POOL_SIZE=25
STREAM_REDIS_PASSWORD=secure_redis_password

# Storage
S3_BUCKET=veza-audio-storage
S3_REGION=us-east-1
S3_ACCESS_KEY=your_s3_access_key
S3_SECRET_KEY=your_s3_secret_key

# CDN
CDN_URL=https://cdn.veza.com
CDN_API_KEY=your_cdn_api_key

# Performance
STREAM_MAX_CONNECTIONS=5000
STREAM_BUFFER_SIZE_MB=512
STREAM_WORKER_THREADS=25
```

## 🐳 Déploiement avec Docker

### Docker Compose Production

```yaml
# docker-compose.production.yml
version: '3.8'

services:
  # Load Balancer
  haproxy:
    image: haproxy:2.6
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./config/haproxy.cfg:/usr/local/etc/haproxy/haproxy.cfg:ro
      - ./ssl:/etc/ssl:ro
    depends_on:
      - backend-api-1
      - backend-api-2
      - chat-server-1
      - chat-server-2
      - stream-server-1
      - stream-server-2
    networks:
      - veza-network

  # Backend API Services
  backend-api-1:
    build:
      context: ./veza-backend-api
      dockerfile: Dockerfile.production
    environment:
      - ENVIRONMENT=production
      - DATABASE_URL=postgresql://veza_user:secure_password@postgres-primary:5432/veza_prod
      - REDIS_URL=redis://redis-primary:6379
      - JWT_ACCESS_SECRET=${JWT_ACCESS_SECRET}
      - NATS_URL=nats://nats-1:4222,nats://nats-2:4222,nats://nats-3:4222
    depends_on:
      - postgres-primary
      - redis-primary
      - nats-1
    networks:
      - veza-network
    deploy:
      replicas: 1
      resources:
        limits:
          cpus: '2'
          memory: 2G
        reservations:
          cpus: '1'
          memory: 1G

  backend-api-2:
    build:
      context: ./veza-backend-api
      dockerfile: Dockerfile.production
    environment:
      - ENVIRONMENT=production
      - DATABASE_URL=postgresql://veza_user:secure_password@postgres-primary:5432/veza_prod
      - REDIS_URL=redis://redis-primary:6379
      - JWT_ACCESS_SECRET=${JWT_ACCESS_SECRET}
      - NATS_URL=nats://nats-1:4222,nats://nats-2:4222,nats://nats-3:4222
    depends_on:
      - postgres-primary
      - redis-primary
      - nats-1
    networks:
      - veza-network
    deploy:
      replicas: 1
      resources:
        limits:
          cpus: '2'
          memory: 2G
        reservations:
          cpus: '1'
          memory: 1G

  # Chat Server Services
  chat-server-1:
    build:
      context: ./veza-chat-server
      dockerfile: Dockerfile.production
    environment:
      - ENVIRONMENT=production
      - CHAT_DATABASE_URL=postgresql://veza_user:secure_password@postgres-replica:5432/veza_prod
      - CHAT_REDIS_URL=redis://redis-replica:6379
      - CHAT_JWT_SECRET=${JWT_ACCESS_SECRET}
      - CHAT_NATS_URL=nats://nats-1:4222,nats://nats-2:4222,nats://nats-3:4222
    depends_on:
      - postgres-replica
      - redis-replica
      - nats-2
    networks:
      - veza-network
    deploy:
      replicas: 1
      resources:
        limits:
          cpus: '4'
          memory: 4G
        reservations:
          cpus: '2'
          memory: 2G

  chat-server-2:
    build:
      context: ./veza-chat-server
      dockerfile: Dockerfile.production
    environment:
      - ENVIRONMENT=production
      - CHAT_DATABASE_URL=postgresql://veza_user:secure_password@postgres-replica:5432/veza_prod
      - CHAT_REDIS_URL=redis://redis-replica:6379
      - CHAT_JWT_SECRET=${JWT_ACCESS_SECRET}
      - CHAT_NATS_URL=nats://nats-1:4222,nats://nats-2:4222,nats://nats-3:4222
    depends_on:
      - postgres-replica
      - redis-replica
      - nats-2
    networks:
      - veza-network
    deploy:
      replicas: 1
      resources:
        limits:
          cpus: '4'
          memory: 4G
        reservations:
          cpus: '2'
          memory: 2G

  # Stream Server Services
  stream-server-1:
    build:
      context: ./veza-stream-server
      dockerfile: Dockerfile.production
    environment:
      - ENVIRONMENT=production
      - STREAM_DATABASE_URL=postgresql://veza_user:secure_password@postgres-replica:5432/veza_prod
      - STREAM_REDIS_URL=redis://redis-replica:6379
      - S3_BUCKET=${S3_BUCKET}
      - S3_REGION=${S3_REGION}
      - S3_ACCESS_KEY=${S3_ACCESS_KEY}
      - S3_SECRET_KEY=${S3_SECRET_KEY}
    volumes:
      - audio-storage:/var/audio
    depends_on:
      - postgres-replica
      - redis-replica
      - nats-3
    networks:
      - veza-network
    deploy:
      replicas: 1
      resources:
        limits:
          cpus: '4'
          memory: 4G
        reservations:
          cpus: '2'
          memory: 2G

  stream-server-2:
    build:
      context: ./veza-stream-server
      dockerfile: Dockerfile.production
    environment:
      - ENVIRONMENT=production
      - STREAM_DATABASE_URL=postgresql://veza_user:secure_password@postgres-replica:5432/veza_prod
      - STREAM_REDIS_URL=redis://redis-replica:6379
      - S3_BUCKET=${S3_BUCKET}
      - S3_REGION=${S3_REGION}
      - S3_ACCESS_KEY=${S3_ACCESS_KEY}
      - S3_SECRET_KEY=${S3_SECRET_KEY}
    volumes:
      - audio-storage:/var/audio
    depends_on:
      - postgres-replica
      - redis-replica
      - nats-3
    networks:
      - veza-network
    deploy:
      replicas: 1
      resources:
        limits:
          cpus: '4'
          memory: 4G
        reservations:
          cpus: '2'
          memory: 2G

  # Database Layer
  postgres-primary:
    image: postgres:15
    environment:
      - POSTGRES_DB=veza_prod
      - POSTGRES_USER=veza_user
      - POSTGRES_PASSWORD=secure_password
    volumes:
      - postgres-primary-data:/var/lib/postgresql/data
      - ./config/postgres/postgresql.conf:/etc/postgresql/postgresql.conf:ro
    command: postgres -c config_file=/etc/postgresql/postgresql.conf
    networks:
      - veza-network
    deploy:
      replicas: 1
      resources:
        limits:
          cpus: '2'
          memory: 4G
        reservations:
          cpus: '1'
          memory: 2G

  postgres-replica:
    image: postgres:15
    environment:
      - POSTGRES_DB=veza_prod
      - POSTGRES_USER=veza_user
      - POSTGRES_PASSWORD=secure_password
    volumes:
      - postgres-replica-data:/var/lib/postgresql/data
      - ./config/postgres/postgresql-replica.conf:/etc/postgresql/postgresql.conf:ro
    command: postgres -c config_file=/etc/postgresql/postgresql.conf
    depends_on:
      - postgres-primary
    networks:
      - veza-network
    deploy:
      replicas: 1
      resources:
        limits:
          cpus: '2'
          memory: 4G
        reservations:
          cpus: '1'
          memory: 2G

  # Cache Layer
  redis-primary:
    image: redis:7-alpine
    command: redis-server --requirepass secure_redis_password --maxmemory 2gb --maxmemory-policy allkeys-lru
    volumes:
      - redis-primary-data:/data
    networks:
      - veza-network
    deploy:
      replicas: 1
      resources:
        limits:
          cpus: '1'
          memory: 3G
        reservations:
          cpus: '0.5'
          memory: 1G

  redis-replica:
    image: redis:7-alpine
    command: redis-server --replicaof redis-primary 6379 --requirepass secure_redis_password
    depends_on:
      - redis-primary
    networks:
      - veza-network
    deploy:
      replicas: 1
      resources:
        limits:
          cpus: '1'
          memory: 3G
        reservations:
          cpus: '0.5'
          memory: 1G

  # Message Queue
  nats-1:
    image: nats:2.9-alpine
    command: nats-server --cluster_name veza-cluster --server_name nats-1 --cluster nats://0.0.0.0:6222 --routes nats://nats-2:6222,nats://nats-3:6222
    networks:
      - veza-network
    deploy:
      replicas: 1

  nats-2:
    image: nats:2.9-alpine
    command: nats-server --cluster_name veza-cluster --server_name nats-2 --cluster nats://0.0.0.0:6222 --routes nats://nats-1:6222,nats://nats-3:6222
    networks:
      - veza-network
    deploy:
      replicas: 1

  nats-3:
    image: nats:2.9-alpine
    command: nats-server --cluster_name veza-cluster --server_name nats-3 --cluster nats://0.0.0.0:6222 --routes nats://nats-1:6222,nats://nats-2:6222
    networks:
      - veza-network
    deploy:
      replicas: 1

  # Monitoring
  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./config/prometheus/prometheus.yml:/etc/prometheus/prometheus.yml:ro
      - prometheus-data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
      - '--web.console.templates=/etc/prometheus/consoles'
      - '--storage.tsdb.retention.time=200h'
      - '--web.enable-lifecycle'
    networks:
      - veza-network

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=secure_grafana_password
    volumes:
      - grafana-data:/var/lib/grafana
      - ./config/grafana/dashboards:/etc/grafana/provisioning/dashboards:ro
      - ./config/grafana/datasources:/etc/grafana/provisioning/datasources:ro
    depends_on:
      - prometheus
    networks:
      - veza-network

  alertmanager:
    image: prom/alertmanager:latest
    ports:
      - "9093:9093"
    volumes:
      - ./config/alertmanager/alertmanager.yml:/etc/alertmanager/alertmanager.yml:ro
      - alertmanager-data:/alertmanager
    command:
      - '--config.file=/etc/alertmanager/alertmanager.yml'
      - '--storage.path=/alertmanager'
    networks:
      - veza-network

networks:
  veza-network:
    driver: bridge

volumes:
  postgres-primary-data:
  postgres-replica-data:
  redis-primary-data:
  redis-replica-data:
  prometheus-data:
  grafana-data:
  alertmanager-data:
  audio-storage:
```

### Configuration HAProxy

```haproxy
# config/haproxy.cfg
global
    daemon
    maxconn 50000
    log stdout format raw local0 info

defaults
    mode http
    timeout connect 5000ms
    timeout client 50000ms
    timeout server 50000ms
    log global

frontend http_front
    bind *:80
    bind *:443 ssl crt /etc/ssl/veza.com.pem
    redirect scheme https if !{ ssl_fc }
    
    # ACLs
    acl is_api hdr(host) -i api.veza.com
    acl is_chat hdr(host) -i chat.veza.com
    acl is_stream hdr(host) -i stream.veza.com
    
    # Routing
    use_backend backend_api if is_api
    use_backend backend_chat if is_chat
    use_backend backend_stream if is_stream

backend backend_api
    balance roundrobin
    option httpchk GET /health
    server api1 backend-api-1:8080 check
    server api2 backend-api-2:8080 check

backend backend_chat
    balance roundrobin
    option httpchk GET /health
    server chat1 chat-server-1:3001 check
    server chat2 chat-server-2:3001 check

backend backend_stream
    balance roundrobin
    option httpchk GET /health
    server stream1 stream-server-1:3002 check
    server stream2 stream-server-2:3002 check
```

## ☸️ Déploiement Kubernetes

### Namespace

```yaml
# k8s/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: veza
  labels:
    name: veza
```

### ConfigMaps et Secrets

```yaml
# k8s/configmaps.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: veza-config
  namespace: veza
data:
  ENVIRONMENT: "production"
  DOMAIN: "veza.com"
  API_DOMAIN: "api.veza.com"
  CHAT_DOMAIN: "chat.veza.com"
  STREAM_DOMAIN: "stream.veza.com"
---
apiVersion: v1
kind: Secret
metadata:
  name: veza-secrets
  namespace: veza
type: Opaque
data:
  JWT_ACCESS_SECRET: <base64-encoded-secret>
  DATABASE_PASSWORD: <base64-encoded-password>
  REDIS_PASSWORD: <base64-encoded-password>
  S3_ACCESS_KEY: <base64-encoded-key>
  S3_SECRET_KEY: <base64-encoded-secret>
```

### Deployments

```yaml
# k8s/deployments.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend-api
  namespace: veza
spec:
  replicas: 2
  selector:
    matchLabels:
      app: backend-api
  template:
    metadata:
      labels:
        app: backend-api
    spec:
      containers:
      - name: backend-api
        image: veza/backend-api:latest
        ports:
        - containerPort: 8080
        envFrom:
        - configMapRef:
            name: veza-config
        - secretRef:
            name: veza-secrets
        resources:
          limits:
            cpu: "2"
            memory: "2Gi"
          requests:
            cpu: "1"
            memory: "1Gi"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: chat-server
  namespace: veza
spec:
  replicas: 2
  selector:
    matchLabels:
      app: chat-server
  template:
    metadata:
      labels:
        app: chat-server
    spec:
      containers:
      - name: chat-server
        image: veza/chat-server:latest
        ports:
        - containerPort: 3001
        envFrom:
        - configMapRef:
            name: veza-config
        - secretRef:
            name: veza-secrets
        resources:
          limits:
            cpu: "4"
            memory: "4Gi"
          requests:
            cpu: "2"
            memory: "2Gi"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: stream-server
  namespace: veza
spec:
  replicas: 2
  selector:
    matchLabels:
      app: stream-server
  template:
    metadata:
      labels:
        app: stream-server
    spec:
      containers:
      - name: stream-server
        image: veza/stream-server:latest
        ports:
        - containerPort: 3002
        envFrom:
        - configMapRef:
            name: veza-config
        - secretRef:
            name: veza-secrets
        resources:
          limits:
            cpu: "4"
            memory: "4Gi"
          requests:
            cpu: "2"
            memory: "2Gi"
        volumeMounts:
        - name: audio-storage
          mountPath: /var/audio
      volumes:
      - name: audio-storage
        persistentVolumeClaim:
          claimName: audio-storage-pvc
```

### Services

```yaml
# k8s/services.yaml
apiVersion: v1
kind: Service
metadata:
  name: backend-api-service
  namespace: veza
spec:
  selector:
    app: backend-api
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
  type: ClusterIP
---
apiVersion: v1
kind: Service
metadata:
  name: chat-server-service
  namespace: veza
spec:
  selector:
    app: chat-server
  ports:
  - protocol: TCP
    port: 80
    targetPort: 3001
  type: ClusterIP
---
apiVersion: v1
kind: Service
metadata:
  name: stream-server-service
  namespace: veza
spec:
  selector:
    app: stream-server
  ports:
  - protocol: TCP
    port: 80
    targetPort: 3002
  type: ClusterIP
```

### Ingress

```yaml
# k8s/ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: veza-ingress
  namespace: veza
  annotations:
    kubernetes.io/ingress.class: "nginx"
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/force-ssl-redirect: "true"
spec:
  tls:
  - hosts:
    - api.veza.com
    - chat.veza.com
    - stream.veza.com
    secretName: veza-tls
  rules:
  - host: api.veza.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: backend-api-service
            port:
              number: 80
  - host: chat.veza.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: chat-server-service
            port:
              number: 80
  - host: stream.veza.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: stream-server-service
            port:
              number: 80
```

## 🔄 CI/CD Pipeline

### GitHub Actions

```yaml
# .github/workflows/deploy.yml
name: Deploy to Production

on:
  push:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: '1.21'
    
    - name: Set up Rust
      uses: actions-rs/toolchain@v1
      with:
        toolchain: stable
    
    - name: Test Backend API
      run: |
        cd veza-backend-api
        go test ./...
    
    - name: Test Chat Server
      run: |
        cd veza-chat-server
        cargo test
    
    - name: Test Stream Server
      run: |
        cd veza-stream-server
        cargo test

  build:
    needs: test
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v2
    
    - name: Login to Docker Hub
      uses: docker/login-action@v2
      with:
        username: ${{ secrets.DOCKER_USERNAME }}
        password: ${{ secrets.DOCKER_PASSWORD }}
    
    - name: Build and push Backend API
      uses: docker/build-push-action@v4
      with:
        context: ./veza-backend-api
        file: ./veza-backend-api/Dockerfile.production
        push: true
        tags: veza/backend-api:latest,veza/backend-api:${{ github.sha }}
    
    - name: Build and push Chat Server
      uses: docker/build-push-action@v4
      with:
        context: ./veza-chat-server
        file: ./veza-chat-server/Dockerfile.production
        push: true
        tags: veza/chat-server:latest,veza/chat-server:${{ github.sha }}
    
    - name: Build and push Stream Server
      uses: docker/build-push-action@v4
      with:
        context: ./veza-stream-server
        file: ./veza-stream-server/Dockerfile.production
        push: true
        tags: veza/stream-server:latest,veza/stream-server:${{ github.sha }}

  deploy:
    needs: build
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Deploy to Production
      run: |
        # Déploiement avec kubectl ou docker-compose
        kubectl apply -f k8s/
        # ou
        docker-compose -f docker-compose.production.yml up -d
```

## 📊 Monitoring et Alertes

### Configuration Prometheus

```yaml
# config/prometheus/prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - "alert_rules.yml"

alerting:
  alertmanagers:
    - static_configs:
        - targets:
          - alertmanager:9093

scrape_configs:
  - job_name: 'backend-api'
    static_configs:
      - targets: ['backend-api-1:8080', 'backend-api-2:8080']
    metrics_path: '/metrics'
    scrape_interval: 10s

  - job_name: 'chat-server'
    static_configs:
      - targets: ['chat-server-1:3001', 'chat-server-2:3001']
    metrics_path: '/metrics'
    scrape_interval: 10s

  - job_name: 'stream-server'
    static_configs:
      - targets: ['stream-server-1:3002', 'stream-server-2:3002']
    metrics_path: '/metrics'
    scrape_interval: 10s

  - job_name: 'postgres'
    static_configs:
      - targets: ['postgres-primary:5432', 'postgres-replica:5432']

  - job_name: 'redis'
    static_configs:
      - targets: ['redis-primary:6379', 'redis-replica:6379']
```

### Règles d'Alerte

```yaml
# config/prometheus/alert_rules.yml
groups:
  - name: veza_alerts
    rules:
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.1
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "High error rate detected"
          description: "Error rate is {{ $value }} errors per second"

      - alert: HighLatency
        expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 1
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "High latency detected"
          description: "95th percentile latency is {{ $value }} seconds"

      - alert: HighMemoryUsage
        expr: (container_memory_usage_bytes / container_spec_memory_limit_bytes) > 0.8
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High memory usage"
          description: "Memory usage is {{ $value | humanizePercentage }}"

      - alert: HighCPUUsage
        expr: (rate(container_cpu_usage_seconds_total[5m]) * 100) > 80
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High CPU usage"
          description: "CPU usage is {{ $value }}%"

      - alert: DatabaseConnectionHigh
        expr: pg_stat_database_numbackends > 80
        for: 2m
        labels:
          severity: critical
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

### Configuration AlertManager

```yaml
# config/alertmanager/alertmanager.yml
global:
  smtp_smarthost: 'smtp.gmail.com:587'
  smtp_from: 'alerts@veza.com'
  smtp_auth_username: 'alerts@veza.com'
  smtp_auth_password: 'your_smtp_password'

route:
  group_by: ['alertname']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'team-veza'

receivers:
  - name: 'team-veza'
    email_configs:
      - to: 'team@veza.com'
        send_resolved: true
    slack_configs:
      - api_url: 'https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK'
        channel: '#alerts'
        send_resolved: true
    webhook_configs:
      - url: 'https://api.pagerduty.com/v2/enqueue'
        send_resolved: true
```

## 🔒 Sécurité

### SSL/TLS Configuration

```nginx
# config/nginx/ssl.conf
ssl_protocols TLSv1.2 TLSv1.3;
ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512:ECDHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES256-GCM-SHA384;
ssl_prefer_server_ciphers off;
ssl_session_cache shared:SSL:10m;
ssl_session_timeout 10m;
ssl_stapling on;
ssl_stapling_verify on;
add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
add_header X-Frame-Options DENY always;
add_header X-Content-Type-Options nosniff always;
add_header X-XSS-Protection "1; mode=block" always;
```

### Firewall Configuration

```bash
# UFW Configuration
ufw default deny incoming
ufw default allow outgoing

# Allow SSH
ufw allow ssh

# Allow HTTP/HTTPS
ufw allow 80/tcp
ufw allow 443/tcp

# Allow specific ports for services
ufw allow 8080/tcp  # Backend API
ufw allow 3001/tcp  # Chat Server
ufw allow 3002/tcp  # Stream Server

# Allow monitoring ports
ufw allow 9090/tcp  # Prometheus
ufw allow 3000/tcp  # Grafana

# Enable firewall
ufw enable
```

## 📈 Performance et Optimisation

### Configuration PostgreSQL

```sql
-- config/postgres/postgresql.conf
# Memory Configuration
shared_buffers = 2GB
effective_cache_size = 6GB
work_mem = 16MB
maintenance_work_mem = 256MB

# Connection Configuration
max_connections = 200
max_worker_processes = 8
max_parallel_workers_per_gather = 4
max_parallel_workers = 8

# WAL Configuration
wal_buffers = 16MB
checkpoint_completion_target = 0.9
wal_writer_delay = 200ms

# Query Optimization
random_page_cost = 1.1
effective_io_concurrency = 200
```

### Configuration Redis

```redis
# config/redis/redis.conf
# Memory Configuration
maxmemory 2gb
maxmemory-policy allkeys-lru

# Persistence
save 900 1
save 300 10
save 60 10000

# Network
tcp-keepalive 300
timeout 0

# Performance
tcp-backlog 511
databases 16
```

## 🚨 Troubleshooting

### Commandes de Diagnostic

```bash
# Vérification de santé des services
curl -f http://localhost:8080/health
curl -f http://localhost:3001/health
curl -f http://localhost:3002/health

# Vérification des logs
docker-compose logs backend-api-1
docker-compose logs chat-server-1
docker-compose logs stream-server-1

# Vérification des métriques
curl http://localhost:9090/api/v1/query?query=up

# Vérification de la base de données
docker exec -it postgres-primary psql -U veza_user -d veza_prod -c "SELECT version();"

# Vérification de Redis
docker exec -it redis-primary redis-cli ping

# Vérification de NATS
docker exec -it nats-1 nats-server --version
```

### Scripts de Maintenance

```bash
#!/bin/bash
# scripts/maintenance.sh

echo "Starting maintenance..."

# Backup database
echo "Creating database backup..."
docker exec postgres-primary pg_dump -U veza_user veza_prod > backup_$(date +%Y%m%d_%H%M%S).sql

# Cleanup old logs
echo "Cleaning old logs..."
find /var/log -name "*.log" -mtime +7 -delete

# Restart services
echo "Restarting services..."
docker-compose restart backend-api-1 backend-api-2
docker-compose restart chat-server-1 chat-server-2
docker-compose restart stream-server-1 stream-server-2

echo "Maintenance completed!"
```

---

**Dernière mise à jour** : $(date)
**Version** : 1.0.0
**Maintenu par** : Équipe Veza 