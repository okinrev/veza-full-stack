# 📖 PRODUCTION GUIDE - VEZA RUST MODULES

> **Guide complet pour le déploiement et l'exploitation des modules Rust en production**  
> **Version** : 2.0 Production-Ready  
> **Dernière mise à jour** : 1er juillet 2025

---

## 🎯 APERÇU SYSTÈME

### **Architecture Production**
```
┌─────────────────────────────────────────┐
│                VEZA PLATFORM            │
├─────────────────┬───────────────────────┤
│   CHAT SERVER   │   STREAM SERVER       │
│   (Rust)        │   (Rust)              │
├─────────────────┼───────────────────────┤
│ • 100k+ WS      │ • 10k+ Streams        │
│ • <10ms latency │ • 100k+ Listeners     │
│ • E2E Encryption│ • Adaptive Bitrate    │
│ • AI Moderation │ • Real-time Effects   │
└─────────────────┴───────────────────────┘
            │
    ┌───────┼───────┐
    │   BACKEND GO  │
    │ (API Gateway) │
    └───────────────┘
```

### **Spécifications Techniques**
- **Performance** : 100k+ connexions WebSocket simultanées
- **Latency** : <10ms P99 pour messages, <50ms pour streaming
- **Throughput** : 10k+ requêtes/seconde par instance
- **Availability** : 99.99% uptime target
- **Scalability** : Horizontale avec auto-scaling

---

## 🔧 CONFIGURATION PRODUCTION

### **Variables d'Environnement**
```bash
# === CORE CONFIG ===
RUST_LOG=info
ENVIRONMENT=production
SERVICE_NAME=veza-stream-server
VERSION=2.0.0

# === NETWORK ===
HOST=0.0.0.0
PORT=8080
WS_PORT=8081
GRPC_PORT=50051

# === DATABASE ===
DATABASE_URL=postgresql://veza:secure_pass@postgres:5432/veza_prod
DATABASE_POOL_SIZE=100
DATABASE_TIMEOUT_MS=5000

# === REDIS ===
REDIS_URL=redis://redis:6379
REDIS_POOL_SIZE=50
REDIS_TTL_DEFAULT=3600

# === MONITORING ===
PROMETHEUS_PORT=9090
JAEGER_ENDPOINT=http://jaeger:14268/api/traces
METRICS_ENABLED=true
TRACING_ENABLED=true

# === PERFORMANCE ===
MAX_CONNECTIONS=100000
WORKER_THREADS=16
BLOCKING_THREADS=32
MEMORY_LIMIT_MB=8192

# === SECURITY ===
JWT_SECRET=your_production_jwt_secret_here
ENCRYPTION_KEY=your_32_byte_encryption_key_here
RATE_LIMIT_PER_MINUTE=1000
ENABLE_CORS=true
ALLOWED_ORIGINS=https://veza.live,https://app.veza.live

# === AUDIO STREAMING ===
MAX_STREAMS=10000
MAX_LISTENERS_PER_STREAM=10000
ADAPTIVE_BITRATE=true
DEFAULT_BITRATE=128
CODECS_ENABLED=opus,aac,mp3

# === CHAT ===
MAX_MESSAGE_SIZE=8192
MESSAGE_HISTORY_LIMIT=1000
MODERATION_ENABLED=true
E2E_ENCRYPTION=optional
```

---

## 🚀 DÉPLOIEMENT KUBERNETES

### **Deployment Manifest**
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: veza-stream-server
  labels:
    app: veza-stream-server
spec:
  replicas: 5
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 2
      maxUnavailable: 1
  selector:
    matchLabels:
      app: veza-stream-server
  template:
    metadata:
      labels:
        app: veza-stream-server
    spec:
      containers:
      - name: stream-server
        image: veza/stream-server:2.0.0
        ports:
        - containerPort: 8080
          name: http
        - containerPort: 8081
          name: websocket
        - containerPort: 50051
          name: grpc
        - containerPort: 9090
          name: metrics
        env:
        - name: RUST_LOG
          value: "info"
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: postgres-secret
              key: url
        resources:
          requests:
            cpu: "2000m"
            memory: "4Gi"
          limits:
            cpu: "8000m"
            memory: "16Gi"
        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 5
        livenessProbe:
          httpGet:
            path: /health/live
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
```

---

## 📊 MONITORING & ALERTING

### **Métriques Clés**
```prometheus
# Latency (target: P99 < 50ms)
http_request_duration_seconds{quantile="0.99"} < 0.05

# Throughput (target: > 10k req/s)
rate(http_requests_total[1m]) > 10000

# Error Rate (target: < 0.1%)
rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m]) < 0.001
```

### **Alerting Rules**
```yaml
groups:
  - name: veza-rust-modules
    rules:
      - alert: HighLatency
        expr: histogram_quantile(0.99, http_request_duration_seconds) > 0.1
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "High latency detected"
          
      - alert: ServiceDown
        expr: up == 0
        for: 30s
        labels:
          severity: critical
```

---

## 🔒 SÉCURITÉ PRODUCTION

### **Configuration TLS**
```rust
let tls_config = TlsConfig::builder()
    .cert_file("/certs/server.crt")
    .key_file("/certs/server.key")
    .min_tls_version(TlsVersion::TLSv1_3)
    .build()?;
```

### **Rate Limiting**
```rust
let rate_limiter = RateLimiter::new(
    1000,  // 1000 requests per minute
    Duration::from_secs(60)
);
```

---

## 🚨 RUNBOOKS INCIDENTS

### **High Latency**
1. Check CPU/Memory usage
2. Scale horizontally if needed
3. Investigate database performance

### **Service Down**
1. Restart deployment
2. Check dependencies (PostgreSQL, Redis)
3. Verify network connectivity

---

**🎯 Cette documentation garantit un déploiement production robuste des modules Rust.** 