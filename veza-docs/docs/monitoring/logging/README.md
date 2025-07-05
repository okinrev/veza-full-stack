---
id: logging-readme
title: Gestion des Logs
sidebar_label: Logs
---

# Gestion des Logs - Veza

## Vue d'ensemble

Ce document décrit la gestion des logs dans la plateforme Veza.

## Configuration des Logs

### Format Structuré
```json
{
  "timestamp": "2024-01-01T12:00:00Z",
  "level": "info",
  "service": "backend-api",
  "message": "Request processed",
  "request_id": "req_123"
}
```

### Niveaux de Log
- **DEBUG** : Informations de débogage
- **INFO** : Informations générales
- **WARN** : Avertissements
- **ERROR** : Erreurs
- **FATAL** : Erreurs critiques

## Centralisation

- **ELK Stack** : Elasticsearch, Logstash, Kibana
- **Fluentd** : Collecte des logs
- **Grafana Loki** : Stockage et requêtes

---

**Dernière mise à jour** : $(date)
**Version** : 1.0.0 