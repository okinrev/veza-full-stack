---
id: grafana-templates
title: Templates Grafana
sidebar_label: Templates
---

# Templates Grafana - Veza

## Vue d'ensemble

Ce document contient les templates de dashboards Grafana.

## Templates Disponibles

### Dashboard Principal
- Métriques système
- Métriques application
- Métriques base de données

### Dashboard Sécurité
- Tentatives de connexion
- Erreurs d'authentification
- Violations de rate limit

## Utilisation

```bash
# Importer un template
curl -X POST http://grafana:3000/api/dashboards/db \
  -H "Content-Type: application/json" \
  -d @dashboard-template.json
```

---

**Dernière mise à jour** : $(date)
**Version** : 1.0.0 