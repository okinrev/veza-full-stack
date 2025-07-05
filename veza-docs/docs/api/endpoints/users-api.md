# API Utilisateurs - Documentation Complète

## Vue d'ensemble

L'API Utilisateurs de Veza gère tous les aspects liés aux profils utilisateurs, authentification, préférences et interactions sociales. Cette documentation couvre l'ensemble des endpoints disponibles avec des exemples complets, schémas de validation et cas d'usage.

### Architecture

```mermaid
graph TB
    subgraph "API Utilisateurs"
        A[Handler] --> B[Service]
        B --> C[Repository]
        C --> D[PostgreSQL]
        
        B --> E[Redis Cache]
        B --> F[File Storage]
        B --> G[Event Bus]
    end
    
    subgraph "Fonctionnalités"
        H[Profils] --> A
        I[Avatars] --> A
        J[Préférences] --> A
        K[Recherche] --> A
        L[Statistiques] --> A
    end
    
    subgraph "Intégrations"
        A --> M[Notification Service]
        A --> N[Analytics Service]
        A --> O[CDN Service]
    end
```

## Endpoints

### 1. Liste des utilisateurs

**GET** `/api/v1/users`

Récupère la liste paginée des utilisateurs publics de la plateforme.

#### Authentification
- **Requis** : Oui (JWT Bearer Token)
- **Rôles** : Tous les utilisateurs authentifiés

#### Paramètres de requête

| Paramètre | Type | Requis | Défaut | Description |
|-----------|------|--------|--------|-------------|
| `page` | integer | Non | 1 | Numéro de page (commence à 1) |
| `limit` | integer | Non | 20 | Nombre d'éléments par page (1-100) |
| `sort` | string | Non | username | Critère de tri (username, created_at, popularity, last_active) |
| `order` | string | Non | asc | Ordre de tri (asc, desc) |
| `role` | string | Non | - | Filtrer par rôle (user, admin, super_admin) |
| `status` | string | Non | active | Filtrer par statut (active, inactive, banned) |
| `search` | string | Non | - | Recherche par username ou nom complet |
| `created_after` | date | Non | - | Utilisateurs créés après cette date |
| `created_before` | date | Non | - | Utilisateurs créés avant cette date |

#### Exemples de requête

```bash
# Liste basique
curl -X GET "http://localhost:8080/api/v1/users" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# Recherche avec filtres
curl -X GET "http://localhost:8080/api/v1/users?search=alice&role=user&sort=popularity&order=desc&limit=10" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# Utilisateurs récents
curl -X GET "http://localhost:8080/api/v1/users?created_after=2024-01-01&sort=created_at&order=desc" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

#### Réponse succès (200)

```json
{
  "success": true,
  "message": "Users retrieved successfully",
  "data": [
    {
      "id": 123,
      "username": "alice_doe",
      "display_name": "Alice Doe",
      "bio": "Passionate music lover and audio engineer",
      "avatar_url": "https://cdn.veza.app/avatars/123.jpg",
      "banner_url": "https://cdn.veza.app/banners/123.jpg",
      "role": "user",
      "is_verified": true,
      "is_online": false,
      "created_at": "2024-01-15T10:30:00Z",
      "last_active": "2024-01-20T15:45:00Z",
      "location": "Paris, France",
      "website": "https://alice-music.com",
      "social_links": {
        "twitter": "alice_doe",
        "instagram": "alice.music",
        "soundcloud": "alice-producer"
      },
      "stats": {
        "tracks": 15,
        "followers": 89,
        "following": 156,
        "total_plays": 1234
      },
      "skills": ["producer", "mixing", "electronic"]
    }
  ],
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 1250,
    "total_pages": 63,
    "has_next": true,
    "has_prev": false
  }
}
```

#### Erreurs possibles

| Code | Message | Description |
|------|---------|-------------|
| 400 | Invalid query parameters | Paramètres de requête invalides |
| 401 | Unauthorized | Token manquant ou invalide |
| 500 | Internal server error | Erreur serveur |

#### Logique métier

1. **Validation des paramètres** : Vérification des limites et formats
2. **Construction de la requête** : Assemblage des filtres SQL
3. **Exécution optimisée** : Utilisation des index sur username, created_at
4. **Cache** : Mise en cache Redis 5 minutes pour les requêtes populaires
5. **Filtrage des données** : Exclusion des informations sensibles

#### Performance

- **Index** : `idx_users_username`, `idx_users_created_at`, `idx_users_role`
- **Cache** : Redis 5 minutes pour les 100 requêtes les plus populaires
- **Pagination** : Curseur pour les grandes collections (>10k)
- **Lazy loading** : Avatars chargés à la demande

---

### 2. Profil utilisateur connecté

**GET** `/api/v1/users/me`

Récupère le profil complet de l'utilisateur actuellement authentifié.

#### Authentification
- **Requis** : Oui (JWT Bearer Token)
- **Rôles** : Utilisateur connecté uniquement

#### Exemple de requête

```bash
curl -X GET "http://localhost:8080/api/v1/users/me" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

#### Réponse succès (200)

```json
{
  "success": true,
  "message": "User profile retrieved successfully",
  "data": {
    "id": 123,
    "username": "alice_doe",
    "email": "alice@example.com",
    "display_name": "Alice Doe",
    "bio": "Passionate music lover and audio engineer",
    "avatar_url": "https://cdn.veza.app/avatars/123.jpg",
    "banner_url": "https://cdn.veza.app/banners/123.jpg",
    "role": "user",
    "is_verified": true,
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-20T10:15:00Z",
    "last_login": "2024-01-20T15:45:00Z",
    "timezone": "Europe/Paris",
    "language": "fr",
    "phone": "+33123456789",
    "location": "Paris, France",
    "website": "https://alice-music.com",
    "social_links": {
      "twitter": "alice_doe",
      "instagram": "alice.music",
      "soundcloud": "alice-producer"
    },
    "preferences": {
      "theme": "dark",
      "notifications": {
        "email": true,
        "push": true,
        "desktop": false,
        "new_followers": true,
        "track_comments": true,
        "direct_messages": true,
        "mentions": true,
        "likes": true
      },
      "privacy": {
        "show_email": false,
        "show_activity": true,
        "allow_dm": true,
        "track_visibility": "public",
        "profile_visibility": "public"
      },
      "audio": {
        "auto_play": false,
        "quality": "high",
        "volume": 0.8,
        "crossfade": 3
      }
    },
    "stats": {
      "tracks_uploaded": 15,
      "total_plays": 1234,
      "total_likes": 89,
      "followers": 89,
      "following": 156,
      "playlists": 8,
      "total_listening_time": 45680
    },
    "subscription": {
      "plan": "free",
      "expires_at": null,
      "features": ["basic_upload", "public_tracks"],
      "limits": {
        "upload_size_mb": 100,
        "tracks_per_month": 10,
        "storage_gb": 1,
        "bandwidth_gb": 10
      },
      "usage": {
        "tracks_this_month": 3,
        "storage_used_gb": 0.45,
        "bandwidth_used_gb": 2.1
      }
    },
    "security": {
      "two_factor_enabled": false,
      "last_password_change": "2024-01-15T10:30:00Z",
      "active_sessions": 2,
      "login_attempts": {
        "failed_today": 0,
        "last_failed": null
      },
      "trusted_devices": [
        {
          "id": "device_abc123",
          "name": "MacBook Pro - Chrome",
          "last_used": "2024-01-20T15:45:00Z",
          "ip_address": "192.168.1.100",
          "location": "Paris, France"
        }
      ]
    },
    "recent_activity": [
      {
        "type": "track_upload",
        "timestamp": "2024-01-20T14:30:00Z",
        "details": {
          "track_id": 456,
          "track_title": "Summer Vibes"
        }
      },
      {
        "type": "new_follower",
        "timestamp": "2024-01-20T12:15:00Z",
        "details": {
          "follower_id": 789,
          "follower_username": "music_fan"
        }
      }
    ]
  }
}
```

---

### 3. Mise à jour du profil

**PUT** `/api/v1/users/me`

Met à jour les informations du profil de l'utilisateur connecté.

#### Authentification
- **Requis** : Oui (JWT Bearer Token)
- **Rôles** : Utilisateur connecté uniquement

#### Content Types supportés
- `application/json` : Données JSON uniquement
- `multipart/form-data` : Données + fichiers images

#### Schéma de validation

```json
{
  "type": "object",
  "properties": {
    "display_name": {
      "type": "string",
      "minLength": 1,
      "maxLength": 100
    },
    "bio": {
      "type": "string",
      "maxLength": 500
    },
    "location": {
      "type": "string",
      "maxLength": 100
    },
    "website": {
      "type": "string",
      "format": "uri",
      "maxLength": 255
    },
    "timezone": {
      "type": "string",
      "pattern": "^[A-Za-z_/]+$"
    },
    "language": {
      "type": "string",
      "pattern": "^[a-z]{2}$"
    },
    "social_links": {
      "type": "object",
      "properties": {
        "twitter": {"type": "string", "maxLength": 50},
        "instagram": {"type": "string", "maxLength": 50},
        "soundcloud": {"type": "string", "maxLength": 50}
      }
    },
    "skills": {
      "type": "array",
      "maxItems": 10,
      "items": {
        "type": "string",
        "maxLength": 30
      }
    },
    "preferences": {
      "type": "object"
    }
  }
}
```

#### Exemples de requête

```bash
# Mise à jour JSON
curl -X PUT "http://localhost:8080/api/v1/users/me" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "display_name": "Alice Producer",
    "bio": "Electronic music producer from Paris",
    "location": "Paris, France",
    "website": "https://alice-music.com",
    "skills": ["producer", "mixing", "mastering"],
    "preferences": {
      "theme": "dark",
      "notifications": {
        "email": true,
        "push": false
      }
    }
  }'

# Mise à jour avec avatar
curl -X PUT "http://localhost:8080/api/v1/users/me" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -F 'profile_data={"display_name":"Alice Producer"}' \
  -F 'avatar=@/path/to/avatar.jpg' \
  -F 'banner=@/path/to/banner.jpg'
```

#### Réponse succès (200)

```json
{
  "success": true,
  "message": "Profile updated successfully",
  "data": {
    "id": 123,
    "username": "alice_doe",
    "display_name": "Alice Producer",
    "bio": "Electronic music producer from Paris",
    "avatar_url": "https://cdn.veza.app/avatars/123_v2.jpg",
    "updated_at": "2024-01-20T16:45:00Z"
  }
}
```

#### Erreurs possibles

| Code | Message | Description |
|------|---------|-------------|
| 400 | Validation failed | Données invalides |
| 413 | File too large | Fichier image trop volumineux |
| 415 | Unsupported media type | Type de fichier non supporté |
| 422 | Username already taken | Nom d'utilisateur déjà utilisé |

#### Logique métier

1. **Validation** : Schéma JSON + validation des fichiers
2. **Upload d'images** : Redimensionnement automatique (256x256, 512x512)
3. **Mise à jour atomique** : Transaction SQL pour cohérence
4. **Invalidation cache** : Suppression du cache utilisateur
5. **Événements** : Publication profile_updated pour synchronisation
6. **Audit** : Enregistrement des modifications

#### Contraintes d'upload

- **Avatar** : JPEG/PNG, max 5MB, min 64x64px, max 2048x2048px
- **Bannière** : JPEG/PNG, max 10MB, ratio 3:1 recommandé
- **Formats** : Conversion automatique en WebP pour optimisation
- **CDN** : Upload direct vers S3/MinIO avec URL signée

---

### 4. Changement de mot de passe

**PUT** `/api/v1/users/me/password`

Change le mot de passe de l'utilisateur connecté avec validation sécurisée.

#### Authentification
- **Requis** : Oui (JWT Bearer Token)
- **Rôles** : Utilisateur connecté uniquement

#### Schéma de validation

```json
{
  "type": "object",
  "required": ["current_password", "new_password", "confirm_password"],
  "properties": {
    "current_password": {
      "type": "string"
    },
    "new_password": {
      "type": "string",
      "minLength": 8,
      "maxLength": 128,
      "pattern": "^(?=.*[a-z])(?=.*[A-Z])(?=.*\\d).+$"
    },
    "confirm_password": {
      "type": "string"
    },
    "logout_all_devices": {
      "type": "boolean",
      "default": true
    }
  }
}
```

#### Exemple de requête

```bash
curl -X PUT "http://localhost:8080/api/v1/users/me/password" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "current_password": "OldPassword123!",
    "new_password": "NewSecurePass456!",
    "confirm_password": "NewSecurePass456!",
    "logout_all_devices": true
  }'
```

#### Réponse succès (200)

```json
{
  "success": true,
  "message": "Password changed successfully",
  "data": {
    "changed_at": "2024-01-20T16:45:00Z",
    "devices_logged_out": 3,
    "new_tokens_required": true
  }
}
```

#### Processus de sécurité

1. **Vérification** : Hash bcrypt du mot de passe actuel
2. **Validation** : Force du nouveau mot de passe
3. **Hash sécurisé** : bcrypt avec coût 12
4. **Révocation** : Tous les tokens existants invalidés
5. **Notification** : Email de confirmation automatique
6. **Audit** : Log sécurisé avec IP et user-agent
7. **Rate limiting** : 3 tentatives par heure maximum

#### Politique de mot de passe

- **Longueur** : 8-128 caractères
- **Complexité** : Majuscule + minuscule + chiffre
- **Historique** : Différent des 5 derniers
- **Dictionnaire** : Protection contre mots courants
- **Entropie** : Score minimal requis

---

### 5. Autres utilisateurs (sauf moi)

**GET** `/api/v1/users/except-me`

Récupère la liste des utilisateurs en excluant l'utilisateur connecté.

#### Authentification
- **Requis** : Oui (JWT Bearer Token)

#### Paramètres de requête

| Paramètre | Type | Description |
|-----------|------|-------------|
| `page` | integer | Numéro de page |
| `limit` | integer | Éléments par page |
| `search` | string | Recherche par nom |
| `active_only` | boolean | Utilisateurs actifs récemment |
| `exclude_blocked` | boolean | Exclure les utilisateurs bloqués |
| `sort_by_interaction` | boolean | Trier par fréquence d'interaction |

#### Exemple de requête

```bash
curl -X GET "http://localhost:8080/api/v1/users/except-me?active_only=true&sort_by_interaction=true&limit=10" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

#### Cas d'usage

- **Interface de chat** : Sélection de destinataire pour messages directs
- **Système de suivi** : Suggestions d'utilisateurs à suivre
- **Mentions** : Autocomplétion dans commentaires et posts
- **Collaborations** : Recherche de partenaires musicaux

---

### 6. Recherche d'utilisateurs

**GET** `/api/v1/users/search`

Moteur de recherche avancé pour trouver des utilisateurs.

#### Authentification
- **Requis** : Oui (JWT Bearer Token)

#### Paramètres de requête

| Paramètre | Type | Requis | Description |
|-----------|------|--------|-------------|
| `q` | string | Oui | Terme de recherche (min 2 chars) |
| `limit` | integer | Non | Nombre max de résultats (1-50) |
| `role` | string | Non | Filtrer par rôle |
| `verified_only` | boolean | Non | Utilisateurs vérifiés uniquement |
| `has_tracks` | boolean | Non | Utilisateurs avec tracks |
| `location` | string | Non | Filtrer par localisation |
| `skills` | array | Non | Recherche par compétences |
| `cursor` | string | Non | Pagination curseur |

#### Algorithme de recherche

```mermaid
graph TB
    A[Requête utilisateur] --> B[Normalisation]
    B --> C[Recherche exacte]
    B --> D[Recherche fuzzy]
    C --> E[Score 1.0]
    D --> F[Score 0.1-0.9]
    E --> G[Fusion résultats]
    F --> G
    G --> H[Tri par pertinence]
    H --> I[Application filtres]
    I --> J[Résultats finaux]
```

#### Score de pertinence

- **Correspondance exacte** : 1.0
- **Début du username** : 0.9
- **Début du display_name** : 0.8
- **Contient dans username** : 0.7
- **Contient dans display_name** : 0.6
- **Correspondance phonétique** : 0.5
- **Tags/compétences** : 0.4
- **Bio/description** : 0.3

#### Exemple de requête

```bash
curl -X GET "http://localhost:8080/api/v1/users/search?q=alice&verified_only=true&has_tracks=true&skills=producer,mixing&limit=10" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

#### Réponse succès (200)

```json
{
  "success": true,
  "message": "Search completed successfully",
  "data": {
    "results": [
      {
        "id": 123,
        "username": "alice_doe",
        "display_name": "Alice Doe",
        "avatar_url": "https://cdn.veza.app/avatars/123.jpg",
        "role": "user",
        "is_verified": true,
        "is_online": false,
        "search_score": 0.95,
        "match_reasons": ["username_exact"],
        "stats": {
          "tracks": 15,
          "followers": 89
        }
      },
      {
        "id": 124,
        "username": "alice_producer",
        "display_name": "Alice Producer",
        "avatar_url": "https://cdn.veza.app/avatars/124.jpg",
        "role": "user",
        "is_verified": true,
        "is_online": true,
        "search_score": 0.87,
        "match_reasons": ["username_partial", "skill_match"],
        "stats": {
          "tracks": 42,
          "followers": 234
        }
      }
    ],
    "suggestions": ["alice_doe", "alice_producer", "alice_music"],
    "total_count": 156,
    "next_cursor": "eyJsYXN0X2lkIjoxNDU="
  }
}
```

#### Performance

- **Index** : ElasticSearch pour recherche textuelle
- **Cache** : Redis pour requêtes populaires (10 minutes)
- **Autocomplétion** : Index Trie pour suggestions rapides
- **Pagination** : Curseur pour navigation fluide

---

### 7. Avatar utilisateur

**GET** `/api/v1/users/{id}/avatar`

Récupère l'avatar d'un utilisateur avec optimisations multiples.

#### Authentification
- **Requis** : Non (endpoint public)

#### Paramètres

| Paramètre | Type | Description |
|-----------|------|-------------|
| `id` | integer | ID de l'utilisateur (path) |
| `size` | integer | Taille en pixels (32,64,128,256,512) |
| `format` | string | Format image (jpeg,png,webp,avif) |
| `quality` | integer | Qualité compression (1-100) |

#### Exemple de requête

```bash
# Avatar standard
curl -X GET "http://localhost:8080/api/v1/users/123/avatar"

# Avatar optimisé
curl -X GET "http://localhost:8080/api/v1/users/123/avatar?size=256&format=webp&quality=85"
```

#### Headers de réponse

```http
Cache-Control: public, max-age=2592000
ETag: "abc123def456"
Content-Type: image/webp
Content-Length: 8192
Last-Modified: Wed, 20 Jan 2024 16:45:00 GMT
```

#### Optimisations

- **CDN** : Cache global avec invalidation
- **Formats modernes** : WebP/AVIF avec fallback
- **Redimensionnement** : À la volée avec cache
- **Compression** : Optimale selon format et taille
- **ETag** : Validation cache client

#### Avatar par défaut

Si l'utilisateur n'a pas d'avatar personnalisé :
- **Initiales** : Génération automatique avec couleur
- **Identicon** : Pattern unique basé sur l'ID
- **Gravatar** : Fallback via hash email (optionnel)

---

## Modèles de données

### PublicUserProfile

```typescript
interface PublicUserProfile {
  id: number;
  username: string;
  display_name: string;
  bio?: string;
  avatar_url?: string;
  banner_url?: string;
  role: 'user' | 'admin' | 'super_admin';
  is_verified: boolean;
  is_online: boolean;
  created_at: string;
  last_active: string;
  location?: string;
  website?: string;
  social_links?: {
    twitter?: string;
    instagram?: string;
    soundcloud?: string;
  };
  stats: {
    tracks: number;
    followers: number;
    following: number;
    total_plays: number;
  };
  skills: string[];
}
```

### CompleteUserProfile

Étend `PublicUserProfile` avec les champs privés :

```typescript
interface CompleteUserProfile extends PublicUserProfile {
  email: string;
  phone?: string;
  timezone: string;
  language: string;
  updated_at: string;
  last_login: string;
  preferences: UserPreferences;
  subscription: UserSubscription;
  security: UserSecurity;
  recent_activity: UserActivity[];
}
```

### UserPreferences

```typescript
interface UserPreferences {
  theme: 'light' | 'dark' | 'auto';
  notifications: {
    email: boolean;
    push: boolean;
    desktop: boolean;
    new_followers: boolean;
    track_comments: boolean;
    direct_messages: boolean;
    mentions: boolean;
    likes: boolean;
  };
  privacy: {
    show_email: boolean;
    show_activity: boolean;
    allow_dm: boolean;
    track_visibility: 'public' | 'followers' | 'private';
    profile_visibility: 'public' | 'registered' | 'private';
  };
  audio: {
    auto_play: boolean;
    quality: 'low' | 'medium' | 'high' | 'lossless';
    volume: number; // 0-1
    crossfade: number; // secondes
  };
}
```

## Codes d'erreur

| Code | Erreur | Description |
|------|--------|-------------|
| `USER_NOT_FOUND` | 404 | Utilisateur introuvable |
| `VALIDATION_ERROR` | 400 | Données invalides |
| `USERNAME_TAKEN` | 409 | Nom d'utilisateur déjà utilisé |
| `EMAIL_TAKEN` | 409 | Email déjà utilisé |
| `WEAK_PASSWORD` | 400 | Mot de passe trop faible |
| `INVALID_CURRENT_PASSWORD` | 401 | Mot de passe actuel incorrect |
| `FILE_TOO_LARGE` | 413 | Fichier trop volumineux |
| `UNSUPPORTED_FORMAT` | 415 | Format de fichier non supporté |
| `RATE_LIMITED` | 429 | Trop de requêtes |

## Rate Limiting

| Endpoint | Limite | Fenêtre |
|----------|--------|---------|
| `GET /users` | 100 | 1 minute |
| `GET /users/me` | 60 | 1 minute |
| `PUT /users/me` | 10 | 1 minute |
| `PUT /users/me/password` | 3 | 1 heure |
| `GET /users/search` | 30 | 1 minute |
| `GET /users/{id}/avatar` | 200 | 1 minute |

## Sécurité

### Protection des données
- **Chiffrement** : TLS 1.3 en transit
- **Hash** : bcrypt coût 12 pour mots de passe
- **Sanitization** : HTML/SQL injection prevention
- **Validation** : Schémas stricts côté serveur

### Audit Trail
- **Connexions** : IP, user-agent, géolocalisation
- **Modifications** : Avant/après avec timestamp
- **Tentatives** : Échecs de connexion et patterns suspects
- **Accès** : Logs d'accès aux données sensibles

### Conformité
- **RGPD** : Droit à l'oubli, portabilité des données
- **CCPA** : Protection des résidents californiens
- **SOC2** : Contrôles de sécurité et audit

## Performance

### Optimisations base de données
```sql
-- Index pour recherche utilisateurs
CREATE INDEX CONCURRENTLY idx_users_search_gin 
ON users USING gin(to_tsvector('french', username || ' ' || display_name));

-- Index pour tri par popularité
CREATE INDEX CONCURRENTLY idx_users_popularity 
ON users (follower_count DESC, created_at DESC);

-- Index pour filtrage par activité
CREATE INDEX CONCURRENTLY idx_users_active 
ON users (last_active DESC) WHERE last_active > NOW() - INTERVAL '30 days';
```

### Cache Strategy
- **L1** : Application cache (Go maps) - 1 minute
- **L2** : Redis cache - 5 minutes
- **L3** : CDN cache - 30 minutes pour avatars
- **Invalidation** : Événementielle via message queue

### Métriques clés
- **Latence P95** : < 200ms pour GET, < 500ms pour PUT
- **Throughput** : 1000 req/s par instance
- **Cache hit ratio** : > 80% pour les lectures
- **Erreur rate** : < 0.1% en conditions normales

## Tests

### Tests unitaires
```go
func TestGetUsers(t *testing.T) {
    // Arrange
    mockDB := &MockUserRepository{}
    service := NewUserService(mockDB)
    
    // Act
    users, meta, err := service.GetUsers(1, 20, UserFilters{})
    
    // Assert
    assert.NoError(t, err)
    assert.Len(t, users, 20)
    assert.Equal(t, 1, meta.Page)
}

func TestUpdateProfile_ValidationError(t *testing.T) {
    // Test des erreurs de validation
}

func TestChangePassword_Security(t *testing.T) {
    // Test des règles de sécurité
}
```

### Tests d'intégration
```bash
# Test avec base de données réelle
go test -tags=integration ./internal/api/user/...

# Test de performance
go test -bench=. -benchmem ./internal/api/user/...
```

### Tests end-to-end
```javascript
describe('User API', () => {
  it('should create and retrieve user profile', async () => {
    // Test complet du cycle de vie utilisateur
  });
  
  it('should handle avatar upload correctly', async () => {
    // Test upload avec validation
  });
});
``` 