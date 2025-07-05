---
title: Dépannage
sidebar_label: Dépannage
---

# 🛠️ Dépannage

Ce guide présente les procédures de dépannage pour Veza.

# Guide de Dépannage - Veza Platform

## Vue d'ensemble

Ce guide couvre les problèmes courants rencontrés lors du développement, du déploiement et de la maintenance de la plateforme Veza.

## 🔍 **Diagnostic Initial**

### Vérification de l'État du Système

```bash
# Vérifier l'état des services
docker ps -a
systemctl status veza-*

# Vérifier les logs
docker logs veza-backend-api
docker logs veza-chat-server
docker logs veza-stream-server

# Vérifier les métriques
curl http://localhost:9090/metrics
```

### Outils de Diagnostic

- **Prometheus** : Métriques système
- **Grafana** : Visualisation des données
- **Jaeger** : Traçage distribué
- **ELK Stack** : Centralisation des logs

## 🚨 **Problèmes Courants**

### 1. Problèmes d'Authentification

#### Symptômes
- Erreurs 401/403
- Tokens JWT expirés
- Sessions perdues

#### Solutions
```bash
# Vérifier la configuration JWT
curl -X POST http://localhost:8080/auth/refresh \
  -H "Authorization: Bearer <token>"

# Réinitialiser les sessions Redis
redis-cli FLUSHDB
```

#### Prévention
- Configurer l'expiration des tokens
- Implémenter le refresh automatique
- Monitorer les tentatives d'authentification

### 2. Problèmes de Base de Données

#### Symptômes
- Timeouts de connexion
- Erreurs de contrainte
- Performances dégradées

#### Solutions
```sql
-- Vérifier les connexions actives
SELECT * FROM pg_stat_activity;

-- Analyser les requêtes lentes
SELECT query, mean_time, calls 
FROM pg_stat_statements 
ORDER BY mean_time DESC 
LIMIT 10;

-- Vérifier l'espace disque
SELECT schemaname, tablename, pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename))
FROM pg_tables 
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```

#### Prévention
- Configurer le connection pooling
- Optimiser les requêtes
- Mettre en place des index appropriés

### 3. Problèmes de WebSocket

#### Symptômes
- Connexions déconnectées
- Messages perdus
- Latence élevée

#### Solutions
```bash
# Vérifier les connexions WebSocket
netstat -an | grep :8081

# Tester la connectivité
websocat ws://localhost:8081/chat

# Vérifier les logs du chat server
docker logs veza-chat-server --tail 100
```

#### Prévention
- Implémenter la reconnexion automatique
- Monitorer la latence réseau
- Configurer les timeouts appropriés

### 4. Problèmes de Streaming

#### Symptômes
- Audio coupé
- Qualité dégradée
- Buffering excessif

#### Solutions
```bash
# Vérifier les ressources système
htop
iotop

# Analyser le trafic réseau
iftop -i eth0

# Vérifier les logs du stream server
docker logs veza-stream-server --tail 100
```

#### Prévention
- Optimiser les codecs audio
- Configurer la bande passante
- Monitorer les performances

### 5. Problèmes de Performance

#### Symptômes
- Temps de réponse élevés
- Utilisation CPU/mémoire excessive
- Timeouts fréquents

#### Solutions
```bash
# Analyser les performances
docker stats

# Vérifier les métriques Prometheus
curl http://localhost:9090/api/v1/query?query=rate(http_requests_total[5m])

# Analyser les logs d'accès
tail -f /var/log/nginx/access.log | grep -E "(4[0-9]{2}|5[0-9]{2})"
```

#### Prévention
- Configurer le cache Redis
- Optimiser les requêtes
- Mettre en place le load balancing

## 🔧 **Procédures de Récupération**

### Redémarrage des Services

```bash
# Redémarrer tous les services
docker-compose down
docker-compose up -d

# Redémarrer un service spécifique
docker-compose restart veza-backend-api
```

### Restauration de Base de Données

```bash
# Sauvegarde
pg_dump -h localhost -U veza_user veza_db > backup.sql

# Restauration
psql -h localhost -U veza_user veza_db < backup.sql
```

### Nettoyage du Cache

```bash
# Vider le cache Redis
redis-cli FLUSHALL

# Redémarrer Redis
docker-compose restart redis
```

## 📊 **Monitoring et Alerting**

### Métriques Clés à Surveiller

- **Temps de réponse API** : < 200ms
- **Taux d'erreur** : < 1%
- **Utilisation CPU** : < 80%
- **Utilisation mémoire** : < 85%
- **Espace disque** : < 90%

### Alertes Configurées

```yaml
# Exemple d'alerte Prometheus
groups:
- name: veza_alerts
  rules:
  - alert: HighErrorRate
    expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.1
    for: 2m
    labels:
      severity: critical
    annotations:
      summary: "Taux d'erreur élevé"
      description: "Le taux d'erreur HTTP 5xx est supérieur à 10%"
```

## 🛠️ **Outils de Diagnostic**

### Scripts Utilitaires

```bash
#!/bin/bash
# health_check.sh

echo "=== Vérification de l'état du système ==="

# Services
echo "Services Docker:"
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

# Base de données
echo "Connexions DB:"
psql -h localhost -U veza_user -d veza_db -c "SELECT count(*) FROM pg_stat_activity;"

# Redis
echo "Redis:"
redis-cli ping

# Métriques
echo "Métriques API:"
curl -s http://localhost:8080/health | jq .
```

### Logs à Surveiller

```bash
# Logs d'application
tail -f /var/log/veza/application.log

# Logs d'erreur
tail -f /var/log/veza/error.log

# Logs d'accès
tail -f /var/log/nginx/access.log
```

## 📞 **Escalade et Support**

### Niveaux d'Escalade

1. **Niveau 1** : Équipe de développement
2. **Niveau 2** : DevOps/SRE
3. **Niveau 3** : Architecte système

### Contacts d'Urgence

- **DevOps** : devops@veza.com
- **SRE** : sre@veza.com
- **Architecte** : architect@veza.com

## 📚 **Ressources Supplémentaires**

- [Guide de Monitoring](../monitoring/README.md)
- [Guide de Sécurité](../security/security-guide.md)
- [Architecture de la Plateforme](../architecture/backend-architecture.md)
- [Guide de Déploiement](../deployment/guide.md)

---

**Dernière mise à jour** : $(date)
**Version** : 1.0.0 