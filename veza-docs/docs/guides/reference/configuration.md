---
id: configuration-reference
title: Référence Configuration
sidebar_label: Configuration
---

# Référence Configuration - Veza

## Vue d'ensemble

Ce document liste toutes les variables de configuration de Veza.

## Variables d'Environnement

### Base de Données
```bash
DATABASE_URL=postgresql://user:pass@localhost:5432/veza
DB_MAX_CONNECTIONS=10
DB_TIMEOUT=30s
```

### Redis
```bash
REDIS_URL=redis://localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
```

### API
```bash
API_PORT=8080
API_HOST=0.0.0.0
JWT_SECRET=your-secret-key
```

### Monitoring
```bash
PROMETHEUS_PORT=9090
GRAFANA_PORT=3000
LOG_LEVEL=info
```

---

**Dernière mise à jour** : $(date)
**Version** : 1.0.0 