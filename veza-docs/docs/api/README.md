---
id: api-overview
title: API Overview
sidebar_label: Vue d'ensemble
---

# 🚀 API Veza - Vue d'ensemble

## 📋 Introduction

L'API Veza est une API RESTful moderne qui fournit tous les services nécessaires pour la plateforme Veza, incluant l'authentification, la gestion des utilisateurs, le chat en temps réel, et le streaming audio.

## 🔗 Endpoints Principaux

### Authentification
- `POST /auth/login` - Connexion utilisateur
- `POST /auth/register` - Inscription utilisateur
- `POST /auth/logout` - Déconnexion
- `POST /auth/refresh` - Renouvellement de token
- `GET /auth/oauth/{provider}` - OAuth2
- `GET /auth/oauth/{provider}/callback` - Callback OAuth2

### Utilisateurs
- `GET /api/v1/users` - Liste des utilisateurs
- `GET /api/v1/users/{id}` - Détails utilisateur
- `PUT /api/v1/users/{id}` - Mise à jour utilisateur
- `DELETE /api/v1/users/{id}` - Suppression utilisateur

### Chat
- `GET /api/v1/rooms` - Liste des salles
- `POST /api/v1/rooms` - Créer une salle
- `GET /api/v1/rooms/{id}/messages` - Messages d'une salle
- `POST /api/v1/rooms/{id}/messages` - Envoyer un message

### Streaming
- `GET /api/v1/streams` - Liste des streams
- `POST /api/v1/streams` - Créer un stream
- `GET /api/v1/streams/{id}` - Détails d'un stream
- `PUT /api/v1/streams/{id}` - Mettre à jour un stream

## 🔐 Authentification

L'API utilise JWT (JSON Web Tokens) pour l'authentification. Tous les endpoints protégés nécessitent un token valide dans l'en-tête `Authorization`.

```bash
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  https://api.veza.com/v1/users
```

## 📝 Format de Réponse

Toutes les réponses suivent un format standard :

```json
{
  "success": true,
  "data": {
    // Données de la réponse
  },
  "message": "Opération réussie",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

## ⚠️ Gestion d'Erreurs

Les erreurs suivent un format standard :

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Description de l'erreur",
    "details": {
      // Détails supplémentaires
    }
  },
  "timestamp": "2024-01-01T12:00:00Z"
}
```

## 📚 Documentation Détaillée

- [Authentification](./authentication.md)
- [Gestion d'Erreurs](./error-handling.md)
- [Webhooks](./webhooks.md)
- [Référence des Endpoints](./endpoints-reference.md)

## 🛠️ SDK et Outils

- **JavaScript SDK** : `npm install @veza/api-client`
- **Python SDK** : `pip install veza-api-client`
- **Collection Postman** : Disponible dans le repository GitHub

## 📞 Support

- **Documentation** : [docs.veza.com](https://docs.veza.com)
- **Support Email** : api-support@veza.com
- **Slack** : #api-support

---

**Dernière mise à jour** : $(date)
**Version API** : 1.0.0 