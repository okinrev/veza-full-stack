---
id: error-codes
title: Codes d'Erreur
sidebar_label: Codes d'Erreur
---

# ⚠️ Codes d'Erreur - Veza

## 📋 Vue d'ensemble

Ce guide répertorie tous les codes d'erreur utilisés dans la plateforme Veza.

## 🔢 Codes d'Erreur Système

### Erreurs HTTP Standards
```yaml
http_codes:
  400: "Bad Request - Requête malformée"
  401: "Unauthorized - Authentification requise"
  403: "Forbidden - Accès refusé"
  404: "Not Found - Ressource introuvable"
  409: "Conflict - Conflit de données"
  422: "Unprocessable Entity - Données invalides"
  429: "Too Many Requests - Rate limit dépassé"
  500: "Internal Server Error - Erreur serveur"
  502: "Bad Gateway - Erreur de gateway"
  503: "Service Unavailable - Service indisponible"
  504: "Gateway Timeout - Timeout de gateway"
```

### Codes d'Erreur Veza
```yaml
veza_codes:
  # Authentification
  AUTH_INVALID_CREDENTIALS: "Identifiants invalides"
  AUTH_TOKEN_EXPIRED: "Token expiré"
  AUTH_TOKEN_INVALID: "Token invalide"
  AUTH_INSUFFICIENT_PERMISSIONS: "Permissions insuffisantes"
  AUTH_ACCOUNT_LOCKED: "Compte verrouillé"
  AUTH_2FA_REQUIRED: "Authentification à deux facteurs requise"
  
  # Validation
  VALIDATION_REQUIRED_FIELD: "Champ requis manquant"
  VALIDATION_INVALID_FORMAT: "Format invalide"
  VALIDATION_INVALID_EMAIL: "Email invalide"
  VALIDATION_PASSWORD_TOO_WEAK: "Mot de passe trop faible"
  VALIDATION_INVALID_USERNAME: "Nom d'utilisateur invalide"
  
  # Ressources
  RESOURCE_NOT_FOUND: "Ressource introuvable"
  RESOURCE_ALREADY_EXISTS: "Ressource déjà existante"
  RESOURCE_DELETED: "Ressource supprimée"
  RESOURCE_UNAVAILABLE: "Ressource indisponible"
  
  # Chat
  CHAT_ROOM_FULL: "Salle de chat pleine"
  CHAT_MESSAGE_TOO_LONG: "Message trop long"
  CHAT_RATE_LIMIT_EXCEEDED: "Limite de messages dépassée"
  CHAT_USER_BLOCKED: "Utilisateur bloqué"
  
  # Streaming
  STREAM_NOT_FOUND: "Stream introuvable"
  STREAM_ALREADY_ACTIVE: "Stream déjà actif"
  STREAM_UNAVAILABLE: "Stream indisponible"
  STREAM_QUALITY_UNAVAILABLE: "Qualité de stream indisponible"
  
  # Fichiers
  FILE_TOO_LARGE: "Fichier trop volumineux"
  FILE_TYPE_NOT_ALLOWED: "Type de fichier non autorisé"
  FILE_UPLOAD_FAILED: "Échec de l'upload"
  FILE_NOT_FOUND: "Fichier introuvable"
  
  # Système
  SYSTEM_MAINTENANCE: "Système en maintenance"
  SYSTEM_OVERLOADED: "Système surchargé"
  SYSTEM_ERROR: "Erreur système"
  DATABASE_ERROR: "Erreur de base de données"
  CACHE_ERROR: "Erreur de cache"
  NETWORK_ERROR: "Erreur réseau"
```

## 📝 Format de Réponse d'Erreur

### Structure Standard
```json
{
  "error": {
    "code": "AUTH_INVALID_CREDENTIALS",
    "message": "Identifiants invalides",
    "details": {
      "field": "email",
      "reason": "Email non trouvé",
      "suggestion": "Vérifiez votre email"
    },
    "timestamp": "2024-01-01T12:00:00Z",
    "request_id": "req_123456789",
    "help_url": "https://docs.veza.com/errors/AUTH_INVALID_CREDENTIALS"
  }
}
```

### Exemples d'Erreurs

#### Erreur d'Authentification
```json
{
  "error": {
    "code": "AUTH_TOKEN_EXPIRED",
    "message": "Token expiré",
    "details": {
      "expired_at": "2024-01-01T11:30:00Z",
      "current_time": "2024-01-01T12:00:00Z"
    },
    "timestamp": "2024-01-01T12:00:00Z",
    "request_id": "req_123456789"
  }
}
```

#### Erreur de Validation
```json
{
  "error": {
    "code": "VALIDATION_REQUIRED_FIELD",
    "message": "Champ requis manquant",
    "details": {
      "field": "username",
      "value": null,
      "constraints": {
        "min_length": 3,
        "max_length": 50,
        "pattern": "^[a-zA-Z0-9_]+$"
      }
    },
    "timestamp": "2024-01-01T12:00:00Z",
    "request_id": "req_123456789"
  }
}
```

#### Erreur de Ressource
```json
{
  "error": {
    "code": "RESOURCE_NOT_FOUND",
    "message": "Ressource introuvable",
    "details": {
      "resource_type": "user",
      "resource_id": "user_123",
      "searched_fields": ["id", "email"]
    },
    "timestamp": "2024-01-01T12:00:00Z",
    "request_id": "req_123456789"
  }
}
```

## 🛠️ Gestion côté Client

### Gestion d'Erreurs JavaScript
```javascript
class VezaError extends Error {
  constructor(errorResponse) {
    super(errorResponse.error.message);
    this.code = errorResponse.error.code;
    this.details = errorResponse.error.details;
    this.timestamp = errorResponse.error.timestamp;
    this.requestId = errorResponse.error.request_id;
  }
}

async function handleApiError(response) {
  if (!response.ok) {
    const errorData = await response.json();
    throw new VezaError(errorData);
  }
  return response.json();
}

// Utilisation
try {
  const data = await fetch('/api/users')
    .then(handleApiError);
} catch (error) {
  if (error instanceof VezaError) {
    switch (error.code) {
      case 'AUTH_TOKEN_EXPIRED':
        // Renouveler le token
        await refreshToken();
        break;
      case 'VALIDATION_REQUIRED_FIELD':
        // Afficher erreur de validation
        showValidationError(error.details);
        break;
      default:
        // Erreur générique
        showGenericError(error.message);
    }
  }
}
```

## 📚 Ressources

- [Gestion d'Erreurs API](../../api/error-handling.md)
- [Guide d'API](../../api/README.md)
- [Authentification](../../api/authentication.md)

---

**Dernière mise à jour** : $(date)
**Version** : 1.0.0 