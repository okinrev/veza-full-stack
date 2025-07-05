---
id: monitoring-metrics
sidebar_label: Metrics
---

# Metrics - Veza Platform

Ce dossier regroupe la documentation sur les métriques, Prometheus, Grafana, etc.

## Index
- À compléter : ajouter la documentation sur les métriques exposées, dashboards, etc.

## Navigation
- [Retour au schéma principal](../../diagrams/architecture-overview.md)

## 📋 Vue d'ensemble

Ce guide détaille les métriques de monitoring et d'observabilité de la plateforme Veza.

## 🎯 Métriques Système

### Métriques Infrastructure
```yaml
# Métriques CPU
cpu_usage_percent: 45.2
cpu_load_average: 1.2
cpu_temperature: 65.0

# Métriques Mémoire
memory_usage_percent: 78.5
memory_available_gb: 2.1
memory_swap_usage_percent: 5.2

# Métriques Disque
disk_usage_percent: 67.8
disk_read_bytes_per_sec: 1024000
disk_write_bytes_per_sec: 2048000

# Métriques Réseau
network_rx_bytes_per_sec: 5120000
network_tx_bytes_per_sec: 2560000
network_connections: 1250
```

### Métriques Application
```yaml
# Métriques API
api_requests_per_second: 150
api_response_time_avg_ms: 245
api_error_rate_percent: 2.1
api_requests_by_endpoint:
  /api/users: 45
  /api/messages: 78
  /api/streams: 27

# Métriques Base de Données
db_connections_active: 45
db_connections_idle: 12
db_query_time_avg_ms: 15.3
db_queries_per_second: 89

# Métriques Cache
cache_hit_rate_percent: 87.5
cache_memory_usage_mb: 512
cache_evictions_per_second: 2.1
```

## 📈 Métriques Métier

### Métriques Utilisateurs
```yaml
# Utilisateurs
total_users: 15420
active_users_today: 3245
new_users_this_week: 156
users_online_now: 234

# Engagement
average_session_duration_minutes: 45
messages_sent_today: 12540
streams_started_today: 89
rooms_created_today: 23
```

### Métriques Performance
```yaml
# Performance Chat
message_delivery_time_avg_ms: 125
message_delivery_success_rate: 99.8
websocket_connections: 567
websocket_messages_per_second: 234

# Performance Streaming
stream_startup_time_avg_seconds: 3.2
stream_bitrate_avg_kbps: 128
stream_quality_score: 4.7
stream_viewers_avg: 45
```

## 🚨 Alertes et Seuils

### Seuils Critiques
```yaml
# Infrastructure
cpu_usage_critical: 90
memory_usage_critical: 95
disk_usage_critical: 90
network_error_rate_critical: 5

# Application
api_response_time_critical: 2000
api_error_rate_critical: 5
db_connection_failure_critical: 10
cache_hit_rate_critical: 70

# Métier
user_experience_critical: 3.0
message_delivery_failure_critical: 2
stream_startup_failure_critical: 10
```

### Configuration d'Alertes
```yaml
alerts:
  - name: "High CPU Usage"
    condition: "cpu_usage > 90"
    duration: "5m"
    severity: "critical"
    
  - name: "API Response Time High"
    condition: "api_response_time_avg > 2000"
    duration: "2m"
    severity: "warning"
    
  - name: "Database Connection Issues"
    condition: "db_connection_failures > 10"
    duration: "1m"
    severity: "critical"
```

## 📊 Dashboards

### Dashboard Infrastructure
```yaml
panels:
  - title: "CPU Usage"
    type: "graph"
    metrics: ["cpu_usage_percent"]
    thresholds: [70, 90]
    
  - title: "Memory Usage"
    type: "graph"
    metrics: ["memory_usage_percent"]
    thresholds: [80, 95]
    
  - title: "Network Traffic"
    type: "graph"
    metrics: ["network_rx_bytes_per_sec", "network_tx_bytes_per_sec"]
```

### Dashboard Application
```yaml
panels:
  - title: "API Requests"
    type: "graph"
    metrics: ["api_requests_per_second"]
    
  - title: "Response Time"
    type: "graph"
    metrics: ["api_response_time_avg_ms"]
    
  - title: "Error Rate"
    type: "graph"
    metrics: ["api_error_rate_percent"]
```

## 📚 Ressources

- [Guide de Monitoring](../README.md)
- [Alertes](../alerts/README.md)
- [Logs](../logs/README.md)

---

**Dernière mise à jour** : $(date)
**Version** : 1.0.0 