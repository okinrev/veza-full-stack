---
id: monitoring-logs
title: Logs et Observabilité
sidebar_label: Logs
---

# 📝 Logs et Observabilité - Veza

## 📋 Vue d'ensemble

Ce guide détaille la gestion des logs et l'observabilité de la plateforme Veza.

## 📊 Structure des Logs

### Format de Log Standard
```json
{
  "timestamp": "2024-01-01T12:00:00.000Z",
  "level": "info",
  "service": "veza-backend-api",
  "version": "1.0.0",
  "trace_id": "trace_123456789",
  "span_id": "span_987654321",
  "user_id": "user_123",
  "session_id": "session_456",
  "ip_address": "192.168.1.100",
  "user_agent": "Mozilla/5.0...",
  "message": "User login successful",
  "metadata": {
    "login_method": "password",
    "success": true,
    "duration_ms": 245
  }
}
```

### Niveaux de Log
```yaml
log_levels:
  trace: 0    # Détails très fins
  debug: 1    # Informations de débogage
  info: 2     # Informations générales
  warn: 3     # Avertissements
  error: 4    # Erreurs
  fatal: 5    # Erreurs critiques
```

## 🔍 Logs par Service

### Logs Backend API
```javascript
// Exemple de log d'API
const apiLog = {
  timestamp: new Date().toISOString(),
  level: 'info',
  service: 'veza-backend-api',
  endpoint: '/api/v1/users',
  method: 'GET',
  status_code: 200,
  response_time_ms: 125,
  user_id: 'user_123',
  ip_address: req.ip,
  user_agent: req.headers['user-agent']
};
```

### Logs Chat Server
```javascript
// Exemple de log de chat
const chatLog = {
  timestamp: new Date().toISOString(),
  level: 'info',
  service: 'veza-chat-server',
  event: 'message_sent',
  room_id: 'room_456',
  user_id: 'user_123',
  message_id: 'msg_789',
  message_length: 150,
  delivery_status: 'delivered'
};
```

### Logs Stream Server
```javascript
// Exemple de log de streaming
const streamLog = {
  timestamp: new Date().toISOString(),
  level: 'info',
  service: 'veza-stream-server',
  event: 'stream_started',
  stream_id: 'stream_123',
  user_id: 'user_456',
  bitrate: 128000,
  codec: 'aac',
  quality: 'high'
};
```

## 📈 Agrégation et Analyse

### Configuration Fluentd
```yaml
# fluentd/conf/fluent.conf
<source>
  @type tail
  path /var/log/containers/*.log
  pos_file /var/log/fluentd-containers.log.pos
  tag kubernetes.*
  read_from_head true
  <parse>
    @type json
    time_format %Y-%m-%dT%H:%M:%S.%NZ
  </parse>
</source>

<filter kubernetes.**>
  @type kubernetes_metadata
  @id filter_kube_metadata
</filter>

<match kubernetes.**>
  @type elasticsearch
  host elasticsearch
  port 9200
  logstash_format true
  logstash_prefix k8s
  <buffer>
    @type file
    path /var/log/fluentd-buffers/kubernetes.system.buffer
    flush_mode interval
    retry_type exponential_backoff
    flush_interval 5s
    retry_forever false
    retry_max_interval 30
    chunk_limit_size 2M
    queue_limit_length 8
    overflow_action block
  </buffer>
</match>
```

### Configuration Elasticsearch
```yaml
# elasticsearch/config/elasticsearch.yml
cluster.name: veza-logs
node.name: veza-logs-node-1

network.host: 0.0.0.0
http.port: 9200

discovery.seed_hosts: ["veza-logs-node-1", "veza-logs-node-2"]
cluster.initial_master_nodes: ["veza-logs-node-1"]

xpack.security.enabled: true
xpack.monitoring.enabled: true
```

## 🔍 Recherche et Filtrage

### Requêtes Elasticsearch
```json
// Recherche d'erreurs récentes
{
  "query": {
    "bool": {
      "must": [
        { "match": { "level": "error" } },
        { "range": { "timestamp": { "gte": "now-1h" } } }
      ]
    }
  },
  "sort": [{ "timestamp": { "order": "desc" } }]
}

// Recherche par utilisateur
{
  "query": {
    "bool": {
      "must": [
        { "match": { "user_id": "user_123" } },
        { "range": { "timestamp": { "gte": "now-24h" } } }
      ]
    }
  }
}

// Recherche d'événements de sécurité
{
  "query": {
    "bool": {
      "must": [
        { "match": { "event": "security" } },
        { "range": { "timestamp": { "gte": "now-7d" } } }
      ]
    }
  }
}
```

## 📊 Visualisation avec Kibana

### Dashboards Kibana
```yaml
kibana_dashboards:
  - name: "Application Logs"
    panels:
      - title: "Logs par Niveau"
        type: "pie"
        field: "level"
        
      - title: "Logs par Service"
        type: "bar"
        field: "service"
        
      - title: "Erreurs dans le Temps"
        type: "line"
        field: "timestamp"
        filter: "level:error"
        
  - name: "Performance API"
    panels:
      - title: "Temps de Réponse"
        type: "line"
        field: "response_time_ms"
        
      - title: "Codes de Statut"
        type: "bar"
        field: "status_code"
```

## 📚 Ressources

- [Guide de Monitoring](../README.md)
- [Métriques](../metrics/README.md)
- [Alertes](../alerts/README.md)

---

**Dernière mise à jour** : $(date)
**Version** : 1.0.0 

# Logs - Veza Platform

Ce dossier regroupe la documentation sur la gestion et l'analyse des logs.

## Index
- À compléter : ajouter la documentation sur ELK, log format, etc.

## Navigation
- [Retour au schéma principal](../../diagrams/architecture-overview.md) 