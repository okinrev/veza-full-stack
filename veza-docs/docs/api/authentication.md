---
id: authentication
title: Authentification API
sidebar_label: Authentification
---

# 🔐 Authentification API - Veza

## 📋 Vue d'ensemble

Ce guide détaille les méthodes d'authentification disponibles dans l'API Veza, incluant JWT, OAuth2, et les API Keys.

## 🔑 Méthodes d'Authentification

### JWT Tokens
```bash
# Login
curl -X POST /auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "password"}'

# Response
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "refresh_token_here",
  "expires_in": 3600
}
```

### OAuth2
```bash
# Initier OAuth
GET /auth/oauth/google

# Callback
GET /auth/oauth/google/callback?code=authorization_code
```

### API Keys
```bash
# Utilisation
curl -H "X-API-Key: your_api_key" \
  https://api.veza.com/v1/users
```

## 📚 Ressources

- [Guide d'Authentification](../api/authentication/README.md)
- [Sécurité](../security/README.md)
- [Configuration](../deployment/environment-variables.md)

---

**Dernière mise à jour** : $(date)
**Version** : 1.0.0 