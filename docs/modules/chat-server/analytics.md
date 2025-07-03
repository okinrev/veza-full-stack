# 📊 Analytics - Chat Server

## Rôle
- Collecte et analyse des métriques d’utilisation du chat
- Statistiques sur les messages, utilisateurs, rooms
- Détection d’anomalies et reporting

## Principales responsabilités
- Agrégation des métriques (messages/min, users actifs, etc.)
- Export vers Prometheus
- Génération de rapports/dashboards

## Interactions
- Utilise PostgreSQL pour l’historique
- Utilise Redis pour les compteurs temps réel
- Publie des events analytics (NATS)

## Points clés
- Faible impact sur la performance
- Données exploitables pour le business et la modération
- Intégration Grafana

---

*À compléter avec des exemples, schémas, et détails d’implémentation.* 