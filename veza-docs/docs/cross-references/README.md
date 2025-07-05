---
id: cross-references
title: Cross-Références
sidebar_label: Cross-Références
---

# Cross-Références - Veza Platform

## Vue d'ensemble

Ce dossier contient les références croisées entre les différents modules et composants de la plateforme Veza.

## Structure des Cross-Références

### Architecture ↔ API
- [Architecture Backend](../architecture/backend-architecture.md) → [API Endpoints](../api/endpoints-reference.md)
- [Architecture Chat Server](../architecture/chat-server-architecture.md) → [WebSocket API](../api/websocket/README.md)
- [Architecture Stream Server](../architecture/stream-server-architecture.md) → [Stream API](../api/streaming-api.md)

### API ↔ Base de Données
- [API Endpoints](../api/endpoints-reference.md) → [Schéma DB](../database/schema.md)
- [Authentification API](../api/authentication.md) → [Tables Users](../database/schema.md#users)
- [Chat API](../api/endpoints/chat-api.md) → [Tables Messages](../database/schema.md#messages)

### Services ↔ Monitoring
- [Backend API](../backend-api/src/cmd-server-main.md) → [Métriques](../monitoring/metrics.md)
- [Chat Server](../chat-server/src/main.md) → [Alertes](../monitoring/alerts/alerting-guide.md)
- [Stream Server](../stream-server/src/main.md) → [Dashboards](../monitoring/grafana/README.md)

## Navigation Bidirectionnelle

### Depuis les Diagrammes
Chaque nœud du [diagramme principal](../diagrams/architecture-overview.md) renvoie vers :
- Documentation détaillée du composant
- Guide de configuration
- Exemples d'utilisation
- Troubleshooting associé

### Depuis la Documentation
Chaque page de documentation renvoie vers :
- Diagramme principal pour contexte
- Composants parents/enfants
- Services dépendants
- Ressources associées

## Index des Références

### Par Module
- **Backend API** : [Architecture](../architecture/backend-architecture.md) | [API](../api/endpoints-reference.md) | [DB](../database/schema.md)
- **Chat Server** : [Architecture](../architecture/chat-server-architecture.md) | [WebSocket](../api/websocket/README.md) | [Monitoring](../monitoring/alerts/alerting-guide.md)
- **Stream Server** : [Architecture](../architecture/stream-server-architecture.md) | [Stream API](../api/streaming-api.md) | [Performance](../guides/performance.md)

### Par Fonctionnalité
- **Authentification** : [API](../api/authentication.md) | [Sécurité](../security/authentication.md) | [DB](../database/schema.md#users)
- **Chat** : [WebSocket](../api/websocket/README.md) | [Modération](../guides/moderation-guide.md) | [Monitoring](../monitoring/alerts/alerting-guide.md)
- **Streaming** : [API](../api/streaming-api.md) | [Performance](../guides/performance.md) | [Storage](../deployment/storage.md)

## Maintenance

### Ajout de Nouvelles Références
1. Identifier les composants liés
2. Créer les liens bidirectionnels
3. Mettre à jour cet index
4. Tester la navigation

### Validation des Liens
```bash
# Vérifier les liens cassés
npm run broken-links

# Valider la navigation
npm run validate-links
```

---

**Dernière mise à jour** : $(date)
**Version** : 1.0.0 