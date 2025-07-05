---
id: migrations-readme
title: Migrations Database
sidebar_label: Migrations
---

# Migrations de Base de Données - Veza

## Vue d'ensemble

Ce document décrit la gestion des migrations de base de données.

## Structure des Migrations

### Format des Fichiers
```sql
-- Migration: 001_create_users_table.up.sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);
```

### Rollback
```sql
-- Migration: 001_create_users_table.down.sql
DROP TABLE users;
```

## Commandes

```bash
# Créer une migration
make migration name=add_user_table

# Appliquer les migrations
make migrate

# Rollback
make migrate-rollback
```

---

**Dernière mise à jour** : $(date)
**Version** : 1.0.0 