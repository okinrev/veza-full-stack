---
id: queries-readme
title: Requêtes Database
sidebar_label: Requêtes
---

# Requêtes de Base de Données - Veza

## Vue d'ensemble

Ce document contient les requêtes SQL importantes pour Veza.

## Requêtes Principales

### Utilisateurs
```sql
-- Récupérer un utilisateur par email
SELECT * FROM users WHERE email = $1;

-- Créer un utilisateur
INSERT INTO users (email, username) VALUES ($1, $2);
```

### Messages
```sql
-- Messages d'une salle
SELECT * FROM messages WHERE room_id = $1 ORDER BY created_at DESC;
```

## Optimisations

- Index sur les colonnes fréquemment utilisées
- Requêtes préparées
- Pagination pour les grandes listes

---

**Dernière mise à jour** : $(date)
**Version** : 1.0.0 