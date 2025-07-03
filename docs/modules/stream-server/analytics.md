# 📊 Analytics - Stream Server

## Rôle
- Collecte et analyse des métriques d’utilisation du streaming
- Statistiques sur les streams, listeners, durée, géolocalisation
- Détection d’anomalies et reporting

## Principales responsabilités
- Agrégation des métriques (streams actifs, bitrate, etc.)
- Export vers Prometheus
- Génération de rapports/dashboards

## Interactions
- Utilise PostgreSQL pour l’historique
- Utilise Redis pour les compteurs temps réel
- Publie des events analytics (NATS)

## Points clés
- Faible impact sur la performance
- Données exploitables pour le business et l’optimisation
- Intégration Grafana

---

*À compléter avec des exemples, schémas, et détails d’implémentation.* 