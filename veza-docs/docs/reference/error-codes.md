---
id: error-codes
title: Codes d'erreur
sidebar_label: Codes d'erreur
description: Référence complète des codes d'erreur de l'API Veza
---

# Codes d'erreur

Ce document référence tous les codes d'erreur utilisés dans l'API Veza, avec leurs descriptions, causes possibles et solutions.

## Structure des erreurs

### Format de réponse d'erreur
```json
{
  "error": {
    "code": "USER_NOT_FOUND",
    "message": "Utilisateur introuvable",
    "details": {
      "user_id": 123,
      "suggestion": "Vérifiez l'ID utilisateur"
    },
    "timestamp": "2024-01-15T10:30:00Z",
    "request_id": "req_abc123_def456"
  }
}
```

### Codes HTTP associés
| Code HTTP | Description | Usage |
|-----------|-------------|-------|
| 400 | Bad Request | Données invalides |
| 401 | Unauthorized | Authentification requise |
| 403 | Forbidden | Permissions insuffisantes |
| 404 | Not Found | Ressource introuvable |
| 409 | Conflict | Conflit de données |
| 422 | Unprocessable Entity | Validation échouée |
| 429 | Too Many Requests | Rate limit dépassé |
| 500 | Internal Server Error | Erreur serveur |
| 503 | Service Unavailable | Service temporairement indisponible |

## Codes d'erreur par domaine

### Authentification (AUTH_*)

| Code | HTTP | Description | Cause | Solution |
|------|------|-------------|-------|----------|
| `AUTH_INVALID_CREDENTIALS` | 401 | Identifiants invalides | Login/mot de passe incorrect | Vérifier les identifiants |
| `AUTH_TOKEN_EXPIRED` | 401 | Token expiré | JWT expiré | Renouveler le token |
| `AUTH_TOKEN_INVALID` | 401 | Token invalide | Format ou signature incorrect | Reconnecter l'utilisateur |
| `AUTH_INSUFFICIENT_PERMISSIONS` | 403 | Permissions insuffisantes | Rôle insuffisant | Vérifier les permissions |
| `AUTH_ACCOUNT_LOCKED` | 403 | Compte verrouillé | Trop de tentatives | Contacter le support |
| `AUTH_2FA_REQUIRED` | 401 | 2FA requis | Authentification à deux facteurs nécessaire | Compléter la 2FA |
| `AUTH_MAGIC_LINK_EXPIRED` | 401 | Lien magique expiré | Lien de connexion expiré | Demander un nouveau lien |

### Utilisateurs (USER_*)

| Code | HTTP | Description | Cause | Solution |
|------|------|-------------|-------|----------|
| `USER_NOT_FOUND` | 404 | Utilisateur introuvable | ID utilisateur invalide | Vérifier l'ID utilisateur |
| `USER_ALREADY_EXISTS` | 409 | Utilisateur déjà existant | Email déjà utilisé | Utiliser un autre email |
| `USER_PROFILE_INCOMPLETE` | 422 | Profil incomplet | Données manquantes | Compléter le profil |
| `USER_ACCOUNT_DISABLED` | 403 | Compte désactivé | Compte suspendu | Contacter le support |
| `USER_EMAIL_NOT_VERIFIED` | 422 | Email non vérifié | Email non confirmé | Vérifier l'email |

### Chat (CHAT_*)

| Code | HTTP | Description | Cause | Solution |
|------|------|-------------|-------|----------|
| `CHAT_ROOM_NOT_FOUND` | 404 | Salon introuvable | ID salon invalide | Vérifier l'ID salon |
| `CHAT_ACCESS_DENIED` | 403 | Accès refusé | Permissions insuffisantes | Demander l'accès |
| `CHAT_MESSAGE_TOO_LONG` | 422 | Message trop long | Limite de caractères dépassée | Raccourcir le message |
| `CHAT_RATE_LIMIT_EXCEEDED` | 429 | Trop de messages | Rate limit dépassé | Attendre avant de renvoyer |
| `CHAT_USER_BLOCKED` | 403 | Utilisateur bloqué | Blocage mutuel | Débloquer l'utilisateur |

### Streaming (STREAM_*)

| Code | HTTP | Description | Cause | Solution |
|------|------|-------------|-------|----------|
| `STREAM_SESSION_NOT_FOUND` | 404 | Session introuvable | ID session invalide | Vérifier l'ID session |
| `STREAM_ALREADY_ACTIVE` | 409 | Stream déjà actif | Session en cours | Arrêter le stream actuel |
| `STREAM_QUALITY_UNAVAILABLE` | 422 | Qualité non disponible | Bitrate non supporté | Choisir une autre qualité |
| `STREAM_BANDWIDTH_EXCEEDED` | 429 | Bande passante dépassée | Limite de bande passante | Réduire la qualité |
| `STREAM_GEO_BLOCKED` | 403 | Géoblocage | Région non autorisée | Utiliser un VPN autorisé |

### Tracks (TRACK_*)

| Code | HTTP | Description | Cause | Solution |
|------|------|-------------|-------|----------|
| `TRACK_NOT_FOUND` | 404 | Track introuvable | ID track invalide | Vérifier l'ID track |
| `TRACK_UPLOAD_FAILED` | 500 | Échec upload | Erreur serveur | Réessayer l'upload |
| `TRACK_FILE_TOO_LARGE` | 413 | Fichier trop volumineux | Taille dépassée | Réduire la taille |
| `TRACK_INVALID_FORMAT` | 422 | Format invalide | Format non supporté | Convertir le format |
| `TRACK_PROCESSING_ERROR` | 500 | Erreur traitement | Échec transcodage | Contacter le support |
| `TRACK_QUOTA_EXCEEDED` | 429 | Quota dépassé | Limite d'upload | Attendre ou upgrade |

### Fichiers (FILE_*)

| Code | HTTP | Description | Cause | Solution |
|------|------|-------------|-------|----------|
| `FILE_NOT_FOUND` | 404 | Fichier introuvable | ID fichier invalide | Vérifier l'ID fichier |
| `FILE_UPLOAD_FAILED` | 500 | Échec upload | Erreur serveur | Réessayer l'upload |
| `FILE_INVALID_TYPE` | 422 | Type invalide | Type non autorisé | Changer le type |
| `FILE_SIZE_EXCEEDED` | 413 | Taille dépassée | Limite de taille | Réduire la taille |
| `FILE_CORRUPTED` | 422 | Fichier corrompu | Données invalides | Re-uploader le fichier |

### Validation (VALIDATION_*)

| Code | HTTP | Description | Cause | Solution |
|------|------|-------------|-------|----------|
| `VALIDATION_REQUIRED_FIELD` | 422 | Champ requis | Champ manquant | Remplir le champ |
| `VALIDATION_INVALID_FORMAT` | 422 | Format invalide | Format incorrect | Corriger le format |
| `VALIDATION_STRING_TOO_LONG` | 422 | Chaîne trop longue | Limite dépassée | Raccourcir la chaîne |
| `VALIDATION_STRING_TOO_SHORT` | 422 | Chaîne trop courte | Minimum non atteint | Allonger la chaîne |
| `VALIDATION_INVALID_EMAIL` | 422 | Email invalide | Format email incorrect | Corriger l'email |
| `VALIDATION_INVALID_URL` | 422 | URL invalide | Format URL incorrect | Corriger l'URL |

### Rate Limiting (RATE_LIMIT_*)

| Code | HTTP | Description | Cause | Solution |
|------|------|-------------|-------|----------|
| `RATE_LIMIT_EXCEEDED` | 429 | Limite dépassée | Trop de requêtes | Attendre avant de renvoyer |
| `RATE_LIMIT_IP_BLOCKED` | 429 | IP bloquée | Comportement suspect | Contacter le support |
| `RATE_LIMIT_USER_BLOCKED` | 429 | Utilisateur bloqué | Abus détecté | Contacter le support |

### Base de données (DB_*)

| Code | HTTP | Description | Cause | Solution |
|------|------|-------------|-------|----------|
| `DB_CONNECTION_ERROR` | 503 | Erreur connexion | Base de données indisponible | Réessayer plus tard |
| `DB_QUERY_TIMEOUT` | 500 | Timeout requête | Requête trop longue | Optimiser la requête |
| `DB_CONSTRAINT_VIOLATION` | 409 | Violation contrainte | Données en conflit | Corriger les données |
| `DB_DEADLOCK_DETECTED` | 500 | Deadlock détecté | Conflit de verrous | Réessayer la transaction |

### Cache (CACHE_*)

| Code | HTTP | Description | Cause | Solution |
|------|------|-------------|-------|----------|
| `CACHE_CONNECTION_ERROR` | 503 | Erreur connexion cache | Redis indisponible | Réessayer plus tard |
| `CACHE_KEY_NOT_FOUND` | 404 | Clé cache introuvable | Données expirées | Recharger les données |
| `CACHE_WRITE_ERROR` | 500 | Erreur écriture cache | Cache plein | Nettoyer le cache |

### Services externes (EXTERNAL_*)

| Code | HTTP | Description | Cause | Solution |
|------|------|-------------|-------|----------|
| `EXTERNAL_SERVICE_UNAVAILABLE` | 503 | Service externe indisponible | API tierce down | Réessayer plus tard |
| `EXTERNAL_SERVICE_TIMEOUT` | 500 | Timeout service externe | Service lent | Réessayer plus tard |
| `EXTERNAL_SERVICE_ERROR` | 500 | Erreur service externe | Erreur API tierce | Contacter le support |

## Gestion des erreurs côté client

### JavaScript/TypeScript
```typescript
interface ApiError {
  code: string;
  message: string;
  details?: Record<string, any>;
  timestamp: string;
  request_id: string;
}

class ApiClient {
  async handleError(response: Response): Promise<never> {
    const error: ApiError = await response.json();
    
    switch (error.code) {
      case 'AUTH_TOKEN_EXPIRED':
        await this.refreshToken();
        break;
      case 'RATE_LIMIT_EXCEEDED':
        await this.waitAndRetry();
        break;
      case 'VALIDATION_REQUIRED_FIELD':
        this.showValidationError(error.details);
        break;
      default:
        this.showGenericError(error);
    }
    
    throw new Error(error.message);
  }
}
```

### Go
```go
type ApiError struct {
    Code      string                 `json:"code"`
    Message   string                 `json:"message"`
    Details   map[string]interface{} `json:"details,omitempty"`
    Timestamp time.Time              `json:"timestamp"`
    RequestID string                 `json:"request_id"`
}

func handleApiError(err *ApiError) error {
    switch err.Code {
    case "AUTH_TOKEN_EXPIRED":
        return refreshToken()
    case "RATE_LIMIT_EXCEEDED":
        return waitAndRetry()
    case "VALIDATION_REQUIRED_FIELD":
        return showValidationError(err.Details)
    default:
        return showGenericError(err)
    }
}
```

## Monitoring des erreurs

### Métriques à surveiller
- **Taux d'erreur par code** : Identifier les erreurs fréquentes
- **Temps de résolution** : Mesurer l'impact sur l'expérience utilisateur
- **Distribution géographique** : Détecter les problèmes régionaux
- **Corrélation avec les releases** : Identifier les régressions

### Alertes critiques
```yaml
# Alertes pour erreurs critiques
- alert: HighErrorRate
  expr: rate(api_errors_total[5m]) > 0.05
  for: 2m
  labels:
    severity: critical
  annotations:
    summary: "High API error rate detected"
    description: "{{ $value }} errors per second"

- alert: AuthErrors
  expr: rate(api_errors_total{code=~"AUTH_.*"}[5m]) > 0.01
  for: 1m
  labels:
    severity: warning
  annotations:
    summary: "Authentication errors detected"
    description: "{{ $value }} auth errors per second"
```

## Dépannage

### Erreurs communes et solutions

#### 401 Unauthorized
**Problème** : Token d'authentification invalide ou expiré
**Solution** :
1. Vérifier la validité du token
2. Renouveler le token si nécessaire
3. Reconnecter l'utilisateur si le refresh échoue

#### 422 Unprocessable Entity
**Problème** : Données de validation invalides
**Solution** :
1. Vérifier le format des données envoyées
2. Consulter la documentation des endpoints
3. Corriger les champs manquants ou invalides

#### 429 Too Many Requests
**Problème** : Rate limit dépassé
**Solution** :
1. Implémenter un backoff exponentiel
2. Réduire la fréquence des requêtes
3. Optimiser les requêtes pour réduire le nombre d'appels

#### 500 Internal Server Error
**Problème** : Erreur serveur interne
**Solution** :
1. Vérifier les logs serveur
2. Contacter le support avec le request_id
3. Réessayer après un délai

## Conclusion

Cette documentation des codes d'erreur permet de comprendre rapidement les problèmes rencontrés et d'implémenter une gestion d'erreur robuste côté client. Pour toute question ou suggestion d'amélioration, n'hésitez pas à créer une issue. 