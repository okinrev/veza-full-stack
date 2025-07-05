---
id: grafana-plugins
title: Plugins Grafana
sidebar_label: Plugins
---

# Plugins Grafana - Veza

## Vue d'ensemble

Ce document liste les plugins Grafana utilisés.

## Plugins Installés

### Data Sources
- **Prometheus** : Métriques temps réel
- **PostgreSQL** : Requêtes directes
- **Redis** : Métriques cache

### Panels
- **Graph** : Graphiques temporels
- **Stat** : Métriques simples
- **Table** : Données tabulaires

## Configuration

```ini
[plugins]
allow_loading_unsigned_plugins = prometheus-datasource,postgres-datasource
```

---

**Dernière mise à jour** : $(date)
**Version** : 1.0.0 