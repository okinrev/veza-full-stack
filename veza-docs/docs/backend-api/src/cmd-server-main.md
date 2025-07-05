---
id: cmd-server-main
title: Main Server
sidebar_label: Main Server
---

# Main Server - Backend API

## Vue d'ensemble

Ce document décrit le point d'entrée principal du serveur backend API de Veza.

## Structure

Le fichier `main.go` dans `cmd/server/` contient :

- Initialisation de la configuration
- Configuration de la base de données
- Démarrage du serveur HTTP
- Gestion des signaux d'arrêt

## Configuration

```go
// Exemple de configuration
type Config struct {
    Port     string
    Database DatabaseConfig
    Redis    RedisConfig
}
```

## Démarrage

```bash
go run cmd/server/main.go
```

---

**Dernière mise à jour** : $(date)
**Version** : 1.0.0
