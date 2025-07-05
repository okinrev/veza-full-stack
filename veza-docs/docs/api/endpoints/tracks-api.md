# API Tracks - Documentation Complète

## Vue d'ensemble

L'API Tracks de Veza gère l'écosystème complet de streaming audio, incluant l'upload, la gestion, la distribution et la consommation de contenu musical. Cette documentation détaille tous les endpoints avec exemples complets, validation, cas d'usage et optimisations performance.

### Architecture Audio

```mermaid
graph TB
    subgraph "Écosystème Audio Veza"
        A[Upload Client] --> B[API Backend]
        B --> C[Validation Audio]
        C --> D[Stockage S3/MinIO]
        D --> E[Processing Pipeline]
        E --> F[CDN Distribution]
        
        B --> G[Base de Données]
        G --> H[Métadonnées]
        G --> I[Tags & Recherche]
        
        F --> J[Stream Server Rust]
        J --> K[Adaptive Streaming]
        K --> L[Client Players]
        
        B --> M[Analytics Engine]
        M --> N[Statistiques Écoute]
        M --> O[Recommandations]
    end
    
    subgraph "Formats Supportés"
        P[MP3] --> C
        Q[WAV] --> C
        R[FLAC] --> C
        S[OGG] --> C
        T[M4A] --> C
        U[AAC] --> C
    end
    
    subgraph "Qualités Streaming"
        K --> V[128 kbps]
        K --> W[320 kbps]
        K --> X[Lossless]
        K --> Y[Adaptive]
    end
```

## Endpoints

### 1. Liste des tracks

**GET** `/api/v1/tracks`

Récupère la liste paginée des tracks publiques avec filtrage avancé.

#### Authentification
- **Requis** : Non pour tracks publiques
- **Optionnel** : JWT pour accéder aux tracks privées de l'utilisateur

#### Paramètres de requête

| Paramètre | Type | Requis | Défaut | Description |
|-----------|------|--------|--------|-------------|
| `page` | integer | Non | 1 | Numéro de page (min 1) |
| `limit` | integer | Non | 20 | Éléments par page (1-100) |
| `sort` | string | Non | created_at | Critère de tri |
| `order` | string | Non | desc | Ordre (asc/desc) |
| `artist` | string | Non | - | Filtrer par artiste |
| `genre` | string | Non | - | Filtrer par genre |
| `tags` | array | Non | - | Filtrer par tags |
| `min_duration` | integer | Non | - | Durée minimale (secondes) |
| `max_duration` | integer | Non | - | Durée maximale (secondes) |
| `search` | string | Non | - | Recherche textuelle |
| `uploader_id` | integer | Non | - | Tracks d'un utilisateur |
| `show_private` | boolean | Non | false | Inclure tracks privées (auth requis) |
| `format` | string | Non | - | Filtrer par format audio |
| `quality` | string | Non | - | Filtrer par qualité |

#### Options de tri

- `created_at` : Date de création (défaut)
- `updated_at` : Dernière modification
- `title` : Titre alphabétique
- `artist` : Artiste alphabétique
- `duration` : Durée
- `plays` : Nombre d'écoutes
- `likes` : Nombre de likes
- `popularity` : Score de popularité calculé

#### Exemples de requête

```bash
# Liste basique des tracks publiques
curl -X GET "http://localhost:8080/api/v1/tracks"

# Recherche avec filtres avancés
curl -X GET "http://localhost:8080/api/v1/tracks?search=electronic&genre=techno&min_duration=180&sort=popularity&order=desc&limit=10"

# Tracks d'un artiste spécifique
curl -X GET "http://localhost:8080/api/v1/tracks?artist=Alice%20Producer&sort=created_at&order=desc"

# Avec authentification pour tracks privées
curl -X GET "http://localhost:8080/api/v1/tracks?show_private=true&uploader_id=123" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# Filtrage par tags multiples
curl -X GET "http://localhost:8080/api/v1/tracks?tags=electronic,ambient,chill&sort=likes&order=desc"
```

#### Réponse succès (200)

```json
{
  "success": true,
  "message": "Tracks retrieved successfully",
  "data": [
    {
      "id": 123,
      "title": "Midnight Dreams",
      "artist": "Alice Producer",
      "filename": "midnight_dreams_v2.mp3",
      "duration_seconds": 245,
      "tags": ["electronic", "ambient", "chill"],
      "is_public": true,
      "uploader_id": 456,
      "uploader_name": "alice_producer",
      "created_at": "2024-01-20T14:30:00Z",
      "updated_at": "2024-01-20T14:30:00Z",
      "stream_url": "https://cdn.veza.app/audio/123/stream?token=abc123",
      "waveform_url": "https://cdn.veza.app/waveforms/123.json",
      "thumbnail_url": "https://cdn.veza.app/thumbnails/123.jpg",
      "stats": {
        "plays": 1234,
        "likes": 89,
        "downloads": 12,
        "shares": 5
      },
      "audio_info": {
        "format": "mp3",
        "bitrate": 320,
        "sample_rate": 44100,
        "channels": 2,
        "file_size": 9876543
      }
    }
  ],
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 1450,
    "total_pages": 73,
    "has_next": true,
    "has_prev": false,
    "filters_applied": {
      "search": "electronic",
      "genre": "techno",
      "min_duration": 180
    }
  }
}
```

#### Logique métier

1. **Parsing des paramètres** : Validation et normalisation
2. **Construction requête SQL** : Jointures avec tables users et stats
3. **Application des filtres** : WHERE clauses dynamiques
4. **Recherche textuelle** : Full-text search sur title/artist
5. **Pagination** : LIMIT/OFFSET avec comptage total
6. **Enrichissement** : URLs streaming, statistiques, métadonnées
7. **Cache** : Redis 3 minutes pour requêtes populaires

#### Performance

- **Index optimisés** : Composite sur (is_public, created_at), (artist), (tags)
- **Cache L1** : Application (1 min)
- **Cache L2** : Redis (3 min)
- **Recherche** : Index GIN PostgreSQL pour full-text
- **CDN** : Métadonnées en edge cache

---

### 2. Détails d'un track

**GET** `/api/v1/tracks/{id}`

Récupère les détails complets d'un track spécifique.

#### Authentification
- **Requis** : Non pour tracks publiques
- **Requis** : Oui pour tracks privées (propriétaire uniquement)

#### Paramètres

| Paramètre | Type | Description |
|-----------|------|-------------|
| `id` | integer | ID du track (path parameter) |
| `include_stats` | boolean | Inclure statistiques détaillées |
| `include_comments` | boolean | Inclure derniers commentaires |
| `include_related` | boolean | Inclure tracks similaires |

#### Exemple de requête

```bash
# Track public
curl -X GET "http://localhost:8080/api/v1/tracks/123?include_stats=true&include_related=true"

# Track privé
curl -X GET "http://localhost:8080/api/v1/tracks/456" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

#### Réponse succès (200)

```json
{
  "success": true,
  "message": "Track retrieved successfully",
  "data": {
    "id": 123,
    "title": "Midnight Dreams",
    "artist": "Alice Producer",
    "description": "A journey through electronic soundscapes...",
    "filename": "midnight_dreams_v2.mp3",
    "duration_seconds": 245,
    "tags": ["electronic", "ambient", "chill", "experimental"],
    "genre": "Electronic",
    "bpm": 128,
    "key": "Am",
    "is_public": true,
    "is_downloadable": true,
    "license": "Creative Commons BY-SA",
    "uploader_id": 456,
    "uploader": {
      "id": 456,
      "username": "alice_producer",
      "display_name": "Alice Producer",
      "avatar_url": "https://cdn.veza.app/avatars/456.jpg",
      "is_verified": true,
      "follower_count": 1234
    },
    "created_at": "2024-01-20T14:30:00Z",
    "updated_at": "2024-01-20T14:30:00Z",
    "urls": {
      "stream": "https://cdn.veza.app/audio/123/stream?token=abc123&quality=high",
      "download": "https://cdn.veza.app/audio/123/download?token=def456",
      "waveform": "https://cdn.veza.app/waveforms/123.json",
      "thumbnail": "https://cdn.veza.app/thumbnails/123.jpg",
      "artwork": "https://cdn.veza.app/artwork/123_1024.jpg"
    },
    "audio_info": {
      "format": "mp3",
      "bitrate": 320,
      "sample_rate": 44100,
      "channels": 2,
      "file_size": 9876543,
      "codec": "LAME 3.100",
      "quality": "high"
    },
    "stats": {
      "plays": 1234,
      "unique_listeners": 890,
      "likes": 89,
      "dislikes": 3,
      "downloads": 12,
      "shares": 5,
      "comments": 23,
      "reposts": 7,
      "plays_today": 45,
      "plays_this_week": 178,
      "plays_this_month": 456
    },
    "related_tracks": [
      {
        "id": 789,
        "title": "Similar Track",
        "artist": "Other Artist",
        "thumbnail_url": "https://cdn.veza.app/thumbnails/789.jpg",
        "similarity_score": 0.85
      }
    ],
    "recent_comments": [
      {
        "id": 1001,
        "user": {
          "id": 101,
          "username": "music_fan",
          "avatar_url": "https://cdn.veza.app/avatars/101.jpg"
        },
        "content": "Amazing track! Love the ambient textures.",
        "timestamp": 156,
        "created_at": "2024-01-20T16:45:00Z",
        "likes": 5
      }
    ]
  }
}
```

#### Erreurs possibles

| Code | Message | Description |
|------|---------|-------------|
| 404 | Track not found | Track inexistant |
| 403 | Access denied | Track privé sans autorisation |
| 410 | Track deleted | Track supprimé |

---

### 3. Upload de track

**POST** `/api/v1/tracks`

Upload un nouveau track audio avec métadonnées.

#### Authentification
- **Requis** : Oui (JWT Bearer Token)
- **Rôles** : Utilisateurs authentifiés

#### Content Type
- `multipart/form-data` : Fichier audio + métadonnées JSON

#### Schéma de validation

```json
{
  "type": "object",
  "required": ["title", "artist", "audio"],
  "properties": {
    "title": {
      "type": "string",
      "minLength": 1,
      "maxLength": 200
    },
    "artist": {
      "type": "string",
      "minLength": 1,
      "maxLength": 100
    },
    "description": {
      "type": "string",
      "maxLength": 2000
    },
    "tags": {
      "type": "array",
      "maxItems": 10,
      "items": {
        "type": "string",
        "maxLength": 30
      }
    },
    "genre": {
      "type": "string",
      "maxLength": 50
    },
    "bpm": {
      "type": "integer",
      "minimum": 40,
      "maximum": 300
    },
    "key": {
      "type": "string",
      "pattern": "^[A-G][#b]?[m]?$"
    },
    "is_public": {
      "type": "boolean",
      "default": true
    },
    "is_downloadable": {
      "type": "boolean",
      "default": false
    },
    "license": {
      "type": "string",
      "enum": ["All Rights Reserved", "Creative Commons BY", "Creative Commons BY-SA", "Creative Commons BY-NC", "Public Domain"]
    }
  }
}
```

#### Contraintes fichier audio

- **Formats** : MP3, WAV, FLAC, OGG, M4A, AAC
- **Taille max** : 100 MB
- **Durée max** : 10 minutes
- **Qualité min** : 128 kbps pour MP3
- **Sample rate** : 22050 Hz minimum

#### Exemple de requête

```bash
curl -X POST "http://localhost:8080/api/v1/tracks" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -F 'metadata={
    "title": "Midnight Dreams",
    "artist": "Alice Producer",
    "description": "A journey through electronic soundscapes",
    "tags": ["electronic", "ambient", "chill"],
    "genre": "Electronic",
    "bpm": 128,
    "key": "Am",
    "is_public": true,
    "is_downloadable": false,
    "license": "Creative Commons BY-SA"
  }' \
  -F 'audio=@/path/to/track.mp3' \
  -F 'artwork=@/path/to/cover.jpg'
```

#### Réponse succès (201)

```json
{
  "success": true,
  "message": "Track uploaded successfully",
  "data": {
    "id": 124,
    "title": "Midnight Dreams",
    "artist": "Alice Producer",
    "filename": "midnight_dreams_20240120.mp3",
    "duration_seconds": 245,
    "tags": ["electronic", "ambient", "chill"],
    "genre": "Electronic",
    "bpm": 128,
    "key": "Am",
    "is_public": true,
    "is_downloadable": false,
    "license": "Creative Commons BY-SA",
    "uploader_id": 456,
    "created_at": "2024-01-20T16:45:00Z",
    "updated_at": "2024-01-20T16:45:00Z",
    "urls": {
      "stream": "https://cdn.veza.app/audio/124/stream?token=xyz789",
      "waveform": "https://cdn.veza.app/waveforms/124.json",
      "thumbnail": "https://cdn.veza.app/thumbnails/124.jpg"
    },
    "processing_status": {
      "status": "processing",
      "progress": 0,
      "estimated_completion": "2024-01-20T16:50:00Z",
      "tasks": ["audio_analysis", "waveform_generation", "thumbnail_creation", "quality_variants"]
    }
  }
}
```

#### Pipeline de traitement

```mermaid
sequenceDiagram
    participant C as Client
    participant A as API
    participant S as Storage
    participant P as Processor
    participant D as Database
    participant CDN as CDN

    C->>A: Upload audio + metadata
    A->>A: Validate file & metadata
    A->>S: Store original file
    A->>D: Save track record
    A->>C: Return track info (processing)
    A->>P: Queue processing job
    
    P->>S: Fetch original file
    P->>P: Audio analysis
    P->>P: Generate waveform
    P->>P: Create thumbnail
    P->>P: Encode quality variants
    P->>CDN: Upload processed files
    P->>D: Update processing status
    P->>A: Notify completion
    A->>C: Webhook notification (optionnel)
```

#### Logique métier

1. **Validation upload** : Type MIME, taille, format
2. **Extraction métadonnées** : FFmpeg pour infos audio
3. **Stockage sécurisé** : S3/MinIO avec path structuré
4. **Enregistrement BDD** : Transaction avec retry
5. **Queue processing** : Jobs asynchrones pour traitement
6. **Génération assets** : Waveform, thumbnail, variants qualité
7. **Indexation recherche** : Mise à jour index Elasticsearch
8. **Notifications** : Events pour analytics et recommandations

---

### 4. Mise à jour track

**PUT** `/api/v1/tracks/{id}`

Met à jour les métadonnées d'un track existant.

#### Authentification
- **Requis** : Oui (JWT Bearer Token)
- **Permissions** : Propriétaire du track uniquement

#### Paramètres

| Paramètre | Type | Description |
|-----------|------|-------------|
| `id` | integer | ID du track (path) |

#### Schéma de validation

```json
{
  "type": "object",
  "properties": {
    "title": {
      "type": "string",
      "minLength": 1,
      "maxLength": 200
    },
    "artist": {
      "type": "string",
      "minLength": 1,
      "maxLength": 100
    },
    "description": {
      "type": "string",
      "maxLength": 2000
    },
    "tags": {
      "type": "array",
      "maxItems": 10,
      "items": {
        "type": "string",
        "maxLength": 30
      }
    },
    "genre": {
      "type": "string",
      "maxLength": 50
    },
    "bpm": {
      "type": "integer",
      "minimum": 40,
      "maximum": 300
    },
    "key": {
      "type": "string",
      "pattern": "^[A-G][#b]?[m]?$"
    },
    "is_public": {
      "type": "boolean"
    },
    "is_downloadable": {
      "type": "boolean"
    },
    "license": {
      "type": "string",
      "enum": ["All Rights Reserved", "Creative Commons BY", "Creative Commons BY-SA", "Creative Commons BY-NC", "Public Domain"]
    }
  }
}
```

#### Exemple de requête

```bash
curl -X PUT "http://localhost:8080/api/v1/tracks/124" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Midnight Dreams (Remix)",
    "description": "Updated description with remix info",
    "tags": ["electronic", "ambient", "chill", "remix"],
    "genre": "Electronic",
    "bpm": 130,
    "is_public": true,
    "is_downloadable": true
  }'
```

#### Réponse succès (200)

```json
{
  "success": true,
  "message": "Track updated successfully",
  "data": {
    "id": 124,
    "title": "Midnight Dreams (Remix)",
    "artist": "Alice Producer",
    "description": "Updated description with remix info",
    "tags": ["electronic", "ambient", "chill", "remix"],
    "genre": "Electronic",
    "bpm": 130,
    "is_public": true,
    "is_downloadable": true,
    "uploader_id": 456,
    "updated_at": "2024-01-20T17:30:00Z",
    "change_log": [
      {
        "field": "title",
        "old_value": "Midnight Dreams",
        "new_value": "Midnight Dreams (Remix)",
        "changed_at": "2024-01-20T17:30:00Z"
      },
      {
        "field": "bpm",
        "old_value": 128,
        "new_value": 130,
        "changed_at": "2024-01-20T17:30:00Z"
      }
    ]
  }
}
```

#### Erreurs possibles

| Code | Message | Description |
|------|---------|-------------|
| 403 | Unauthorized to update | Pas propriétaire |
| 404 | Track not found | Track inexistant |
| 422 | Track is being processed | Processing en cours |

---

### 5. Suppression track

**DELETE** `/api/v1/tracks/{id}`

Supprime définitivement un track et tous ses assets associés.

#### Authentification
- **Requis** : Oui (JWT Bearer Token)
- **Permissions** : Propriétaire du track ou admin

#### Paramètres

| Paramètre | Type | Description |
|-----------|------|-------------|
| `id` | integer | ID du track (path) |
| `force` | boolean | Supprimer même si utilisé dans playlists |

#### Exemple de requête

```bash
curl -X DELETE "http://localhost:8080/api/v1/tracks/124?force=true" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

#### Réponse succès (200)

```json
{
  "success": true,
  "message": "Track deleted successfully",
  "data": {
    "deleted_track_id": 124,
    "title": "Midnight Dreams (Remix)",
    "deleted_at": "2024-01-20T18:00:00Z",
    "cleanup_status": {
      "database_record": "deleted",
      "audio_file": "deleted",
      "waveform": "deleted",
      "thumbnail": "deleted",
      "cdn_cache": "purged",
      "search_index": "removed"
    },
    "impact": {
      "playlists_affected": 3,
      "user_libraries_affected": 12,
      "stats_preserved": true
    }
  }
}
```

#### Processus de suppression

```mermaid
graph LR
    A[DELETE Request] --> B[Validate Ownership]
    B --> C[Check Dependencies]
    C --> D[Soft Delete DB]
    D --> E[Queue Cleanup Jobs]
    E --> F[Remove Files]
    F --> G[Purge CDN]
    G --> H[Update Search Index]
    H --> I[Notify Services]
```

#### Logique métier

1. **Vérification propriété** : User ID ou rôle admin
2. **Analyse impact** : Playlists, favoris, partages
3. **Soft delete** : Marquage deleted_at
4. **Queue cleanup** : Jobs asynchrones de nettoyage
5. **Suppression fichiers** : Storage et CDN
6. **Mise à jour index** : Search et recommandations
7. **Notifications** : Services dépendants
8. **Conservation stats** : Historique anonymisé

---

### 6. Recherche de tracks

**GET** `/api/v1/tracks/search`

Moteur de recherche avancé pour découvrir des tracks.

#### Authentification
- **Requis** : Non (recherche publique)
- **Optionnel** : JWT pour personnalisation

#### Paramètres de requête

| Paramètre | Type | Description |
|-----------|------|-------------|
| `q` | string | Requête de recherche principale |
| `type` | string | Type de recherche (simple, advanced, semantic) |
| `artist` | string | Nom d'artiste |
| `genre` | string | Genre musical |
| `tags` | array | Tags spécifiques |
| `bpm_min` | integer | BPM minimum |
| `bpm_max` | integer | BPM maximum |
| `duration_min` | integer | Durée min (secondes) |
| `duration_max` | integer | Durée max (secondes) |
| `key` | string | Tonalité musicale |
| `license` | string | Type de licence |
| `sort` | string | Critère de tri |
| `limit` | integer | Nombre de résultats (max 50) |
| `offset` | integer | Décalage pour pagination |

#### Algorithme de recherche

```mermaid
graph TB
    A[Requête utilisateur] --> B[Tokenisation]
    B --> C[Recherche exacte]
    B --> D[Recherche floue]
    B --> E[Recherche sémantique]
    
    C --> F[Score 1.0]
    D --> G[Score 0.6-0.9]
    E --> H[Score 0.3-0.8]
    
    F --> I[Fusion scores]
    G --> I
    H --> I
    
    I --> J[Application filtres]
    J --> K[Tri final]
    K --> L[Résultats]
```

#### Score de pertinence

- **Titre exact** : 1.0
- **Artiste exact** : 0.9
- **Titre partiel** : 0.8
- **Tags match** : 0.7
- **Description** : 0.6
- **Similarité phonétique** : 0.5
- **Métadonnées** : 0.4

#### Exemple de requête

```bash
# Recherche simple
curl -X GET "http://localhost:8080/api/v1/tracks/search?q=electronic ambient&limit=10"

# Recherche avancée
curl -X GET "http://localhost:8080/api/v1/tracks/search?q=deep house&genre=Electronic&bpm_min=120&bpm_max=130&duration_min=300&sort=popularity"

# Recherche par tonalité
curl -X GET "http://localhost:8080/api/v1/tracks/search?key=Am&genre=Ambient&sort=created_at"
```

#### Réponse succès (200)

```json
{
  "success": true,
  "message": "Search completed successfully",
  "data": {
    "query": "electronic ambient",
    "total_results": 156,
    "search_time_ms": 45,
    "results": [
      {
        "id": 123,
        "title": "Midnight Dreams",
        "artist": "Alice Producer",
        "score": 0.95,
        "match_reasons": ["title_partial", "tags_match"],
        "highlighted": {
          "title": "Midnight <em>Dreams</em>",
          "tags": ["<em>electronic</em>", "<em>ambient</em>", "chill"]
        },
        "duration_seconds": 245,
        "genre": "Electronic",
        "bpm": 128,
        "key": "Am",
        "thumbnail_url": "https://cdn.veza.app/thumbnails/123.jpg",
        "stream_url": "https://cdn.veza.app/audio/123/stream?token=abc123",
        "uploader": {
          "id": 456,
          "username": "alice_producer",
          "display_name": "Alice Producer"
        },
        "stats": {
          "plays": 1234,
          "likes": 89
        }
      }
    ],
    "facets": {
      "genres": [
        {"name": "Electronic", "count": 89},
        {"name": "Ambient", "count": 67}
      ],
      "artists": [
        {"name": "Alice Producer", "count": 12},
        {"name": "Bob Ambient", "count": 8}
      ],
      "bpm_ranges": [
        {"range": "100-120", "count": 34},
        {"range": "120-140", "count": 78}
      ]
    },
    "suggestions": ["electronic music", "ambient techno", "alice producer"],
    "related_searches": ["deep house", "experimental electronic"]
  }
}
```

---

### 7. Streaming audio

**GET** `/api/v1/tracks/{id}/stream`

Génère une URL de streaming sécurisée pour un track.

#### Authentification
- **Requis** : Non pour tracks publiques
- **Requis** : Oui pour tracks privées

#### Paramètres

| Paramètre | Type | Description |
|-----------|------|-------------|
| `id` | integer | ID du track |
| `quality` | string | Qualité audio (low, medium, high, lossless) |
| `format` | string | Format de sortie (mp3, ogg, m4a) |
| `range` | string | Range bytes (pour streaming adaptatif) |

#### Exemple de requête

```bash
# Stream public
curl -X GET "http://localhost:8080/api/v1/tracks/123/stream?quality=high&format=mp3"

# Stream privé
curl -X GET "http://localhost:8080/api/v1/tracks/456/stream" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

#### Réponse succès (200)

```json
{
  "success": true,
  "message": "Stream URL generated successfully",
  "data": {
    "track_id": 123,
    "stream_url": "https://stream.veza.app/audio/123/high.mp3?token=signed_jwt_token&expires=1705756800",
    "expires_at": "2024-01-20T19:00:00Z",
    "quality": "high",
    "format": "mp3",
    "bitrate": 320,
    "estimated_size": 9876543,
    "duration": 245,
    "streaming_info": {
      "supports_range": true,
      "chunk_size": 32768,
      "adaptive_streaming": true,
      "cdn_endpoints": [
        "https://cdn-eu.veza.app",
        "https://cdn-us.veza.app"
      ]
    }
  }
}
```

---

## Modèles de données

### Track

```typescript
interface Track {
  id: number;
  title: string;
  artist: string;
  description?: string;
  filename: string;
  duration_seconds?: number;
  tags: string[];
  genre?: string;
  bpm?: number;
  key?: string;
  is_public: boolean;
  is_downloadable: boolean;
  license: string;
  uploader_id: number;
  created_at: string;
  updated_at: string;
}
```

### TrackWithMetadata

```typescript
interface TrackWithMetadata extends Track {
  uploader: UserProfile;
  urls: {
    stream: string;
    download?: string;
    waveform: string;
    thumbnail: string;
    artwork?: string;
  };
  audio_info: {
    format: string;
    bitrate: number;
    sample_rate: number;
    channels: number;
    file_size: number;
    codec?: string;
    quality: string;
  };
  stats: {
    plays: number;
    unique_listeners: number;
    likes: number;
    dislikes: number;
    downloads: number;
    shares: number;
    comments: number;
    reposts: number;
    plays_today: number;
    plays_this_week: number;
    plays_this_month: number;
  };
}
```

### UploadRequest

```typescript
interface UploadRequest {
  title: string;
  artist: string;
  description?: string;
  tags?: string[];
  genre?: string;
  bpm?: number;
  key?: string;
  is_public?: boolean;
  is_downloadable?: boolean;
  license?: string;
  audio: File;
  artwork?: File;
}
```

## Validation et contraintes

### Formats audio supportés

| Format | Extension | Max Bitrate | Qualité |
|--------|-----------|-------------|---------|
| MP3 | .mp3 | 320 kbps | Standard |
| WAV | .wav | Lossless | Haute |
| FLAC | .flac | Lossless | Très haute |
| OGG | .ogg | 500 kbps | Standard |
| M4A | .m4a | 256 kbps | Standard |
| AAC | .aac | 320 kbps | Standard |

### Limites par plan

| Plan | Taille Max | Durée Max | Uploads/Mois | Stockage |
|------|------------|-----------|--------------|----------|
| Free | 50 MB | 6 min | 10 | 1 GB |
| Basic | 100 MB | 10 min | 50 | 10 GB |
| Pro | 500 MB | 30 min | 200 | 100 GB |
| Artist | 1 GB | Illimité | 1000 | 1 TB |

## Rate Limiting

| Endpoint | Limite | Fenêtre | Notes |
|----------|--------|---------|-------|
| `GET /tracks` | 100 | 1 minute | Cache agressif |
| `GET /tracks/{id}` | 200 | 1 minute | CDN + cache |
| `POST /tracks` | 5 | 1 heure | Upload processing |
| `PUT /tracks/{id}` | 20 | 1 minute | Metadata only |
| `DELETE /tracks/{id}` | 10 | 1 heure | Safety limit |
| `GET /tracks/search` | 50 | 1 minute | Search optimization |

## Codes d'erreur

| Code | Erreur | Description |
|------|--------|-------------|
| `TRACK_NOT_FOUND` | 404 | Track introuvable |
| `INVALID_AUDIO_FORMAT` | 400 | Format audio non supporté |
| `FILE_TOO_LARGE` | 413 | Fichier trop volumineux |
| `DURATION_EXCEEDED` | 400 | Durée maximale dépassée |
| `UPLOAD_QUOTA_EXCEEDED` | 429 | Quota d'upload dépassé |
| `PROCESSING_ERROR` | 500 | Erreur de traitement audio |
| `UNAUTHORIZED_ACCESS` | 403 | Accès non autorisé |
| `TRACK_PROCESSING` | 422 | Track en cours de traitement |
| `INVALID_METADATA` | 400 | Métadonnées invalides |
| `STORAGE_ERROR` | 500 | Erreur de stockage |

## Performance et optimisation

### Infrastructure

```mermaid
graph TB
    subgraph "Upload Pipeline"
        A[Client Upload] --> B[Load Balancer]
        B --> C[API Servers]
        C --> D[Storage Gateway]
        D --> E[Object Storage]
        
        C --> F[Message Queue]
        F --> G[Processing Workers]
        G --> H[Transcoding]
        H --> I[CDN Distribution]
    end
    
    subgraph "Streaming Pipeline"
        J[Stream Request] --> K[CDN Edge]
        K --> L[Origin Servers]
        L --> M[Adaptive Bitrate]
        M --> N[Client Player]
    end
    
    subgraph "Caching Layers"
        O[Browser Cache]
        P[CDN Cache]
        Q[Redis Cache]
        R[Database]
    end
```

### Optimisations

#### Base de données
```sql
-- Index composite pour recherche
CREATE INDEX CONCURRENTLY idx_tracks_search 
ON tracks USING gin(to_tsvector('english', title || ' ' || artist));

-- Index pour filtrage par genre et BPM
CREATE INDEX CONCURRENTLY idx_tracks_genre_bpm 
ON tracks (genre, bpm) WHERE is_public = true;

-- Index pour tri par popularité
CREATE INDEX CONCURRENTLY idx_tracks_popularity 
ON tracks (play_count DESC, created_at DESC) WHERE is_public = true;

-- Partitioning par date pour les gros volumes
CREATE TABLE tracks_2024 PARTITION OF tracks 
FOR VALUES FROM ('2024-01-01') TO ('2025-01-01');
```

#### Cache strategy
- **L1 (Application)** : 30 secondes pour métadonnées
- **L2 (Redis)** : 5 minutes pour tracks populaires
- **L3 (CDN)** : 1 heure pour assets statiques
- **L4 (Browser)** : 24h pour images/waveforms

#### Streaming optimizations
- **Adaptive bitrate** : 128/256/320 kbps variants
- **Chunked transfer** : 32KB chunks pour stream fluide
- **Edge caching** : Distribution géographique
- **Preloading** : Buffer intelligent côté client

### Métriques de performance

| Métrique | Target | Actuel | Actions |
|----------|--------|--------|---------|
| Upload time (50MB) | &lt;30s | 25s | OK |
| Stream latency | &lt;200ms | 150ms | OK |
| Search response | &lt;100ms | 80ms | OK |
| CDN hit ratio | &gt;90% | 94% | OK |
| Transcode time | &lt;2x duration | 1.5x | OK |
| Error rate | &lt;0.1% | 0.05% | OK |

## Monitoring et observabilité

### Métriques business
- **Uploads par jour** : Tendance croissance
- **Temps d'écoute total** : Engagement utilisateur
- **Ratio découverte** : Efficacité recommandations
- **Conversions premium** : Monétisation

### Métriques techniques
- **Latence P95** : Performance perçue
- **Taux d'erreur** : Fiabilité service
- **Throughput** : Capacité système
- **Utilisation ressources** : Optimisation coûts

### Alertes critiques
- Upload failure rate > 1%
- Stream start latency > 500ms
- Storage usage > 80%
- CDN errors > 0.5%
- Processing queue depth > 100

## Sécurité

### Protection uploads
- **Validation MIME** : Vérification réelle du contenu
- **Scan antivirus** : ClamAV intégré
- **Content analysis** : Détection contenu inapproprié
- **Rate limiting** : Protection contre abus

### Streaming sécurisé
- **URLs signées** : JWT avec expiration
- **Referrer checking** : Anti-hotlinking
- **Geo-blocking** : Restrictions géographiques
- **Watermarking** : Protection copyright

### Audit et conformité
- **DMCA compliance** : Takedown automatique
- **Rights management** : Tracking ownership
- **Usage analytics** : Reporting détaillé
- **Data protection** : RGPD/CCPA compliant 