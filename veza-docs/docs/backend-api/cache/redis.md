---
id: redis-cache
title: Cache Redis
sidebar_label: Redis Cache
---

# Cache Redis - Backend API

## Vue d'ensemble

Ce document décrit l'utilisation de Redis comme cache dans le backend API.

## Configuration

### Variables d'Environnement
```bash
REDIS_URL=redis://localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
```

### Connexion
```go
client := redis.NewClient(&redis.Options{
    Addr:     "localhost:6379",
    Password: "",
    DB:       0,
})
```

## Utilisations

### Cache de Sessions
- Stockage des tokens JWT
- Sessions utilisateur
- Rate limiting

### Cache de Données
- Résultats de requêtes fréquentes
- Métadonnées utilisateur
- Configuration

---

**Dernière mise à jour** : $(date)
**Version** : 1.0.0 