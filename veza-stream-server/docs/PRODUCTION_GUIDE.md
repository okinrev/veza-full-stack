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

### **Limites et Quotas Recommandés**
```yaml
production_limits:
  # CPU & Memory
  cpu_request: "2000m"      # 2 CPU cores minimum
  cpu_limit: "8000m"        # 8 CPU cores maximum
  memory_request: "4Gi"     # 4GB RAM minimum
  memory_limit: "16Gi"      # 16GB RAM maximum
  
  # Network
  max_connections: 100000
  bandwidth_limit: "1Gbps"
  
  # Storage
  ephemeral_storage: "10Gi"
  logs_retention: "30d"
  
  # Application
  max_message_rate: 100     # messages/second/user
  max_file_upload: "200MB"
  concurrent_streams: 10000
```

---

## 🚀 DÉPLOIEMENT PRODUCTION

### **1. Pré-requis Infrastructure**
- **Kubernetes** 1.25+ ou Docker Swarm
- **PostgreSQL** 14+ avec High Availability
- **Redis** 7.0+ cluster mode
- **Load Balancer** avec SSL termination
- **Monitoring** : Prometheus + Grafana stack

### **2. Health Checks**
```yaml
healthcheck:
  readiness:
    path: /health/ready
    port: 8080
    timeout: 5s
    period: 10s
    
  liveness:
    path: /health/live
    port: 8080
    timeout: 5s
    period: 30s
    failure_threshold: 3
```

### **3. Graceful Shutdown**
```rust
// Configuration de graceful shutdown (30s)
tokio::select! {
    _ = signal::ctrl_c() => {
        info!("🛑 Graceful shutdown initiated");
        
        // 1. Stop accepting new connections
        server.stop_accepting().await;
        
        // 2. Wait for existing connections to finish (max 30s)
        timeout(Duration::from_secs(30), 
               server.wait_for_connections()).await;
        
        // 3. Force close remaining connections
        server.force_close().await;
        
        info!("✅ Graceful shutdown completed");
    }
}
```

---

## 📊 MONITORING & ALERTING

### **Métriques Clés à Surveiller**

#### **Performance Metrics**
```prometheus
# Latency (target: P99 < 50ms)
http_request_duration_seconds{quantile="0.99"} < 0.05

# Throughput (target: > 10k req/s)
rate(http_requests_total[1m]) > 10000

# Error Rate (target: < 0.1%)
rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m]) < 0.001
```

#### **Resource Metrics**
```prometheus
# CPU Usage (alert: > 80%)
cpu_usage_percent > 80

# Memory Usage (alert: > 85%)
memory_usage_percent > 85

# Connection Count (alert: > 90k)
websocket_connections_active > 90000
```

#### **Business Metrics**
```prometheus
# Active Users (alert: < 1k unusual drop)
increase(active_users_total[5m]) < -1000

# Message Success Rate (alert: < 99.9%)
message_delivery_success_rate < 0.999

# Stream Quality (alert: > 5% degraded)
stream_quality_degraded_percent > 5
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
          
      - alert: HighErrorRate  
        expr: rate(http_errors_total[5m]) / rate(http_requests_total[5m]) > 0.01
        for: 1m
        labels:
          severity: critical
          
      - alert: ServiceDown
        expr: up == 0
        for: 30s
        labels:
          severity: critical
```

---

## 🔒 SÉCURITÉ PRODUCTION

### **1. Network Security**
- **TLS 1.3** obligatoire pour toutes les connexions
- **Certificate pinning** pour communications inter-services
- **Network policies** Kubernetes restrictives
- **DDoS protection** avec rate limiting intelligent

### **2. Data Protection**
```rust
// Encryption at rest
let encryption_key = load_key_from_vault().await?;
let encrypted_data = AES_256_GCM.encrypt(&data, &encryption_key)?;

// Encryption in transit
let tls_config = TlsConfig::builder()
    .cert_file("/certs/server.crt")
    .key_file("/certs/server.key")
    .min_tls_version(TlsVersion::TLSv1_3)
    .build()?;
```

### **3. Authentication & Authorization**
- **JWT** avec rotation automatique (24h)
- **RBAC** granulaire par resource
- **API keys** avec scopes limités
- **2FA** obligatoire pour comptes privilégiés

---

## 🔄 MAINTENANCE & OPÉRATIONS

### **1. Mise à Jour Rolling**
```bash
# 1. Update image version
kubectl set image deployment/stream-server \
    stream-server=veza/stream-server:v2.1.0

# 2. Monitor rollout
kubectl rollout status deployment/stream-server

# 3. Validate health
kubectl get pods -l app=stream-server
```

### **2. Backup & Recovery**
- **Database** : Point-in-time recovery (PITR) avec PostgreSQL
- **Configuration** : Git-ops avec validation automatique
- **Logs** : Rétention 30 jours avec archivage S3
- **Metrics** : Rétention 1 an avec downsampling

### **3. Scaling Operations**
```bash
# Horizontal scaling
kubectl scale deployment/stream-server --replicas=10

# Vertical scaling (HPA)
kubectl autoscale deployment/stream-server \
    --cpu-percent=70 --min=5 --max=50
```

---

## 🚨 RUNBOOKS INCIDENTS

### **Incident 1 : High Latency**
```bash
# 1. Diagnostic rapide
kubectl top pods -l app=stream-server
kubectl logs -l app=stream-server --tail=100

# 2. Scaling immédiat si CPU > 80%
kubectl scale deployment/stream-server --replicas=20

# 3. Investigation
kubectl exec -it stream-server-xxx -- /bin/bash
htop  # Vérifier CPU/Memory
ss -tulpn  # Vérifier connexions réseau
```

### **Incident 2 : Service Down**
```bash
# 1. Restart rapid
kubectl rollout restart deployment/stream-server

# 2. Check dependencies
kubectl get pods -l app=postgres
kubectl get pods -l app=redis

# 3. Traffic rerouting
kubectl patch service/stream-server -p '{"spec":{"selector":{"app":"stream-server-backup"}}}'
```

### **Incident 3 : Memory Leak**
```bash
# 1. Memory profiling
kubectl exec stream-server-xxx -- /bin/bash
curl http://localhost:9090/debug/pprof/heap > heap.prof

# 2. Graceful restart par batch
for pod in $(kubectl get pods -l app=stream-server -o name); do
    kubectl delete $pod
    sleep 30  # Wait for replacement
done
```

---

## 📈 PERFORMANCE TUNING

### **1. OS Level Optimizations**
```bash
# Network tuning
echo 'net.core.somaxconn=65535' >> /etc/sysctl.conf
echo 'net.core.netdev_max_backlog=5000' >> /etc/sysctl.conf
echo 'net.ipv4.tcp_max_syn_backlog=65535' >> /etc/sysctl.conf

# File descriptor limits
echo '* soft nofile 1048576' >> /etc/security/limits.conf
echo '* hard nofile 1048576' >> /etc/security/limits.conf
```

### **2. Application Tuning**
```rust
// Tokio runtime optimization
let rt = tokio::runtime::Builder::new_multi_thread()
    .worker_threads(num_cpus::get() * 2)
    .max_blocking_threads(512)
    .thread_stack_size(2 * 1024 * 1024)  // 2MB stack
    .enable_all()
    .build()?;
```

### **3. Database Optimization**
```sql
-- Connection pooling
ALTER SYSTEM SET max_connections = 500;
ALTER SYSTEM SET shared_buffers = '4GB';
ALTER SYSTEM SET effective_cache_size = '12GB';
ALTER SYSTEM SET work_mem = '64MB';

-- Index optimization for chat messages
CREATE INDEX CONCURRENTLY idx_messages_room_created 
    ON messages(room_id, created_at DESC) 
    WHERE deleted_at IS NULL;
```

---

## 🔍 TROUBLESHOOTING

### **Problèmes Fréquents**

#### **1. WebSocket Connections Dropping**
```bash
# Check load balancer timeout
kubectl describe ingress stream-server

# Verify heartbeat configuration
grep -r "ping_interval" src/

# Monitor connection metrics
curl http://localhost:9090/metrics | grep websocket
```

#### **2. Audio Stream Latency**
```rust
// Verify buffer configuration
let buffer_config = BufferConfig {
    target_latency: Duration::from_millis(50),
    max_buffer_size: 1024 * 8,  // 8KB
    adaptive: true,
};
```

#### **3. Memory Usage Growth**
```bash
# Check for connection leaks
ss -s | grep tcp
lsof -p $(pgrep stream-server) | wc -l

# Monitor memory pools
curl http://localhost:9090/debug/pprof/allocs
```

---

## 📋 CHECKLIST PRODUCTION

### **Pre-Deployment**
- [ ] Load testing completed (100k+ connections)
- [ ] Security audit passed
- [ ] Monitoring configured
- [ ] Backup strategy validated
- [ ] Disaster recovery tested

### **Post-Deployment**
- [ ] Health checks passing
- [ ] Metrics collecting properly
- [ ] Logs flowing to aggregation
- [ ] Alerts configured and tested
- [ ] Performance within targets

### **Weekly Maintenance**
- [ ] Check resource utilization trends
- [ ] Review error logs
- [ ] Update security patches
- [ ] Validate backup integrity
- [ ] Performance regression testing

---

**🎯 Cette documentation garantit un déploiement production robuste et maintenable des modules Rust Veza.** 