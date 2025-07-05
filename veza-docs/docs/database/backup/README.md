---
id: backup-readme
title: Sauvegarde Database
sidebar_label: Sauvegarde
---

# Sauvegarde de Base de Données - Veza

## Vue d'ensemble

Ce document décrit les procédures de sauvegarde de la base de données.

## Stratégies de Sauvegarde

### Sauvegarde Complète
```bash
pg_dump -h localhost -U veza -d veza > backup_$(date +%Y%m%d).sql
```

### Sauvegarde Incrémentale
```bash
pg_dump -h localhost -U veza -d veza --data-only > incremental_backup.sql
```

## Récupération

```bash
psql -h localhost -U veza -d veza < backup_file.sql
```

## Planification

- Sauvegarde quotidienne à 2h du matin
- Rétention de 30 jours
- Test de restauration mensuel

---

**Dernière mise à jour** : $(date)
**Version** : 1.0.0 