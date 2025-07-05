---
id: authentication-readme
title: Authentification API
sidebar_label: Authentification
---

# Authentification API - Veza

## Vue d'ensemble

Ce document décrit les mécanismes d'authentification de l'API Veza.

## Endpoints d'Authentification

### POST /api/v1/auth/login
Connexion utilisateur avec email et mot de passe.

### POST /api/v1/auth/register
Inscription d'un nouvel utilisateur.

### POST /api/v1/auth/logout
Déconnexion et invalidation du token.

### POST /api/v1/auth/refresh
Renouvellement du token d'accès.

## Configuration

```yaml
auth:
  jwt_secret: "your-secret-key"
  access_token_expiry: 15m
  refresh_token_expiry: 168h
```

## Liens Utiles

- [Guide de Sécurité](../../security/README.md)
- [Variables d'Environnement](../../deployment/environment-variables.md)
- [Monitoring](../../monitoring/README.md)

---

**Dernière mise à jour** : $(date)
**Version** : 1.0.0 