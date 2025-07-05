---
id: development-environment
title: Environnement de Développement
sidebar_label: Environnement
---

# Environnement de Développement - Veza

## Vue d'ensemble

Ce document décrit la configuration de l'environnement de développement.

## Prérequis

### Outils
- Go 1.21+
- Rust 1.70+
- Node.js 18+
- Docker et Docker Compose

### Services
- PostgreSQL 15+
- Redis 7+
- NATS 2.10+

## Configuration

### Variables d'Environnement
```bash
# .env
DATABASE_URL=postgresql://veza:password@localhost:5432/veza_dev
REDIS_URL=redis://localhost:6379
NATS_URL=nats://localhost:4222
```

### Docker Compose
```yaml
version: '3.8'
services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: veza_dev
      POSTGRES_USER: veza
      POSTGRES_PASSWORD: password
```

---

**Dernière mise à jour** : $(date)
**Version** : 1.0.0 