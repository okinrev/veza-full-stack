---
id: authentication
title: Authentification
sidebar_label: Authentification
---

# Authentification - Veza

## Vue d'ensemble

Ce document décrit les mécanismes d'authentification de la plateforme Veza.

## Mécanismes d'Authentification

### JWT Tokens
- **Access Token** : 15 minutes
- **Refresh Token** : 7 jours
- **Signature** : RS256

### OAuth2
- **Google** : Connexion avec Google
- **GitHub** : Connexion avec GitHub
- **Magic Links** : Authentification sans mot de passe

## Configuration

```yaml
auth:
  jwt_secret: "your-secret-key"
  access_token_expiry: 15m
  refresh_token_expiry: 168h
```

---

**Dernière mise à jour** : $(date)
**Version** : 1.0.0 