---
id: health
title: Health Checks
sidebar_label: Health Checks
---

# Health Checks - Monitoring Veza

## Vue d'ensemble

Ce document décrit les health checks de la plateforme Veza pour assurer la surveillance de l'état des services.

## Endpoints de Health Check

### Backend API
```bash
GET /health
GET /ready
GET /live
```

### Chat Server
```bash
GET /health
GET /ready
```

### Stream Server
```bash
GET /health
GET /ready
```

## Format de Réponse

```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T12:00:00Z",
  "services": {
    "database": "healthy",
    "redis": "healthy",
    "external_apis": "healthy"
  }
}
```

## Configuration

```yaml
health_checks:
  interval: 30s
  timeout: 10s
  retries: 3
```

---

**Dernière mise à jour** : $(date)
**Version** : 1.0.0 