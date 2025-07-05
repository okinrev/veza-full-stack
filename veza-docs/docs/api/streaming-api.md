---
id: streaming-api
title: API de Streaming - Veza Platform
sidebar_label: Streaming API
---

# API de Streaming - Veza Platform

> **Documentation complète de l'API de streaming audio de Veza**

## Vue d'ensemble

L'API de streaming permet de gérer les flux audio en temps réel, l'upload de fichiers, et la diffusion de contenu.

## Endpoints

### Gestion des Streams

#### Démarrer un Stream

```http
POST /api/v1/streams/start
```

**Headers :**
```
Authorization: Bearer <token>
Content-Type: application/json
```

**Body :**
```json
{
  "title": "Mon Stream",
  "description": "Description du stream",
  "is_public": true,
  "category": "music",
  "tags": ["rock", "live"]
}
```

**Réponse :**
```json
{
  "stream_id": "stream_123",
  "rtmp_url": "rtmp://stream.veza.com/live/stream_123",
  "hls_url": "https://stream.veza.com/hls/stream_123.m3u8",
  "status": "starting"
}
```

#### Arrêter un Stream

```http
POST /api/v1/streams/{stream_id}/stop
```

**Headers :**
```
Authorization: Bearer <token>
```

**Réponse :**
```json
{
  "stream_id": "stream_123",
  "status": "stopped",
  "duration": 3600,
  "viewers_peak": 150
}
```

#### Obtenir les Informations d'un Stream

```http
GET /api/v1/streams/{stream_id}
```

**Réponse :**
```json
{
  "stream_id": "stream_123",
  "user_id": "user_456",
  "title": "Mon Stream",
  "description": "Description du stream",
  "status": "live",
  "viewers_count": 120,
  "started_at": "2024-01-15T10:00:00Z",
  "category": "music",
  "tags": ["rock", "live"],
  "is_public": true
}
```

### Upload de Fichiers

#### Upload d'un Fichier Audio

```http
POST /api/v1/streams/upload
```

**Headers :**
```
Authorization: Bearer <token>
Content-Type: multipart/form-data
```

**Body :**
```
file: <fichier audio>
title: "Mon Titre"
artist: "Mon Artiste"
album: "Mon Album"
```

**Réponse :**
```json
{
  "file_id": "file_789",
  "filename": "song.mp3",
  "size": 5242880,
  "duration": 180,
  "url": "https://storage.veza.com/audio/file_789.mp3",
  "status": "uploaded"
}
```

#### Obtenir la Liste des Fichiers

```http
GET /api/v1/streams/files
```

**Query Parameters :**
- `page`: Numéro de page (défaut: 1)
- `limit`: Nombre d'éléments par page (défaut: 20)
- `category`: Filtrer par catégorie
- `search`: Recherche par titre/artiste

**Réponse :**
```json
{
  "files": [
    {
      "file_id": "file_789",
      "title": "Mon Titre",
      "artist": "Mon Artiste",
      "album": "Mon Album",
      "duration": 180,
      "size": 5242880,
      "uploaded_at": "2024-01-15T09:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 150,
    "pages": 8
  }
}
```

### Gestion des Playlists

#### Créer une Playlist

```http
POST /api/v1/playlists
```

**Headers :**
```
Authorization: Bearer <token>
Content-Type: application/json
```

**Body :**
```json
{
  "name": "Ma Playlist",
  "description": "Description de la playlist",
  "is_public": false,
  "files": ["file_789", "file_790"]
}
```

**Réponse :**
```json
{
  "playlist_id": "playlist_123",
  "name": "Ma Playlist",
  "description": "Description de la playlist",
  "files_count": 2,
  "duration": 360,
  "created_at": "2024-01-15T10:00:00Z"
}
```

#### Ajouter des Fichiers à une Playlist

```http
POST /api/v1/playlists/{playlist_id}/files
```

**Body :**
```json
{
  "files": ["file_791", "file_792"]
}
```

### Diffusion en Direct

#### Obtenir les Streams en Direct

```http
GET /api/v1/streams/live
```

**Query Parameters :**
- `category`: Filtrer par catégorie
- `limit`: Nombre de streams (défaut: 20)

**Réponse :**
```json
{
  "streams": [
    {
      "stream_id": "stream_123",
      "user_id": "user_456",
      "username": "dj_veza",
      "title": "Live Session",
      "viewers_count": 120,
      "category": "music",
      "started_at": "2024-01-15T10:00:00Z",
      "thumbnail_url": "https://cdn.veza.com/thumbnails/stream_123.jpg"
    }
  ]
}
```

#### Rejoindre un Stream

```http
POST /api/v1/streams/{stream_id}/join
```

**Headers :**
```
Authorization: Bearer <token>
```

**Réponse :**
```json
{
  "stream_id": "stream_123",
  "hls_url": "https://stream.veza.com/hls/stream_123.m3u8",
  "chat_room": "chat_room_123",
  "joined_at": "2024-01-15T10:30:00Z"
}
```

## WebSocket Events

### Événements de Streaming

#### Stream Started
```json
{
  "event": "stream_started",
  "data": {
    "stream_id": "stream_123",
    "user_id": "user_456",
    "title": "Live Session",
    "started_at": "2024-01-15T10:00:00Z"
  }
}
```

#### Stream Stopped
```json
{
  "event": "stream_stopped",
  "data": {
    "stream_id": "stream_123",
    "duration": 3600,
    "viewers_peak": 150,
    "stopped_at": "2024-01-15T11:00:00Z"
  }
}
```

#### Viewer Count Updated
```json
{
  "event": "viewer_count_updated",
  "data": {
    "stream_id": "stream_123",
    "viewers_count": 125,
    "updated_at": "2024-01-15T10:35:00Z"
  }
}
```

## Codes d'Erreur

| Code | Message | Description |
|------|---------|-------------|
| 400 | `INVALID_STREAM_DATA` | Données de stream invalides |
| 401 | `UNAUTHORIZED` | Token d'authentification manquant ou invalide |
| 403 | `STREAM_ACCESS_DENIED` | Accès refusé au stream |
| 404 | `STREAM_NOT_FOUND` | Stream introuvable |
| 409 | `STREAM_ALREADY_LIVE` | Stream déjà en cours |
| 413 | `FILE_TOO_LARGE` | Fichier trop volumineux |
| 415 | `UNSUPPORTED_FORMAT` | Format de fichier non supporté |
| 429 | `RATE_LIMIT_EXCEEDED` | Limite de taux dépassée |
| 500 | `STREAM_ERROR` | Erreur interne du serveur de streaming |

## Exemples d'Implémentation

### JavaScript/Node.js

```javascript
// Démarrer un stream
const startStream = async (streamData) => {
  const response = await fetch('/api/v1/streams/start', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(streamData)
  });
  
  return response.json();
};

// Upload d'un fichier
const uploadFile = async (file) => {
  const formData = new FormData();
  formData.append('file', file);
  formData.append('title', 'Mon Titre');
  formData.append('artist', 'Mon Artiste');
  
  const response = await fetch('/api/v1/streams/upload', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`
    },
    body: formData
  });
  
  return response.json();
};

// WebSocket pour les événements
const ws = new WebSocket('wss://api.veza.com/ws/streaming');

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  
  switch (data.event) {
    case 'stream_started':
      console.log('Stream démarré:', data.data);
      break;
    case 'viewer_count_updated':
      updateViewerCount(data.data.viewers_count);
      break;
  }
};
```

### Python

```python
import requests
import websocket
import json

class VezaStreamingAPI:
    def __init__(self, base_url, token):
        self.base_url = base_url
        self.headers = {'Authorization': f'Bearer {token}'}
    
    def start_stream(self, stream_data):
        response = requests.post(
            f'{self.base_url}/api/v1/streams/start',
            headers=self.headers,
            json=stream_data
        )
        return response.json()
    
    def upload_file(self, file_path, metadata):
        with open(file_path, 'rb') as f:
            files = {'file': f}
            data = metadata
            
            response = requests.post(
                f'{self.base_url}/api/v1/streams/upload',
                headers=self.headers,
                files=files,
                data=data
            )
            return response.json()
    
    def get_live_streams(self, category=None):
        params = {}
        if category:
            params['category'] = category
            
        response = requests.get(
            f'{self.base_url}/api/v1/streams/live',
            headers=self.headers,
            params=params
        )
        return response.json()

# Utilisation
api = VezaStreamingAPI('https://api.veza.com', 'your_token')

# Démarrer un stream
stream = api.start_stream({
    'title': 'Live Session',
    'description': 'Session en direct',
    'is_public': True,
    'category': 'music'
})

# Upload d'un fichier
file_info = api.upload_file('song.mp3', {
    'title': 'Ma Chanson',
    'artist': 'Mon Artiste'
})
```

## Performance et Optimisation

### Bonnes Pratiques

1. **Gestion des Connexions**
   - Utiliser des connexions persistantes
   - Implémenter un retry automatique
   - Gérer les timeouts appropriés

2. **Upload de Fichiers**
   - Utiliser le chunked upload pour les gros fichiers
   - Valider les formats avant upload
   - Compresser les métadonnées

3. **Streaming en Direct**
   - Utiliser HLS pour la compatibilité
   - Implémenter l'adaptive bitrate
   - Optimiser la latence

### Métriques de Performance

```bash
# Latence moyenne des requêtes
curl -w "@curl-format.txt" -o /dev/null -s "https://api.veza.com/api/v1/streams/live"

# Débit de streaming
ffprobe -v quiet -show_entries format=bit_rate -of csv=p=0 stream.m3u8

# Qualité de service
curl -s "https://api.veza.com/health" | jq '.streaming.quality'
```

---

## 🔗 Liens croisés

- [Architecture de Streaming](../architecture/stream-server-architecture.md)
- [WebSocket API](../api/websocket/README.md)
- [Base de Données](../database/schema.md)
- [Monitoring](../monitoring/metrics.md)

---

## Pour aller plus loin

- [Guide de Performance](../guides/performance.md)
- [Troubleshooting](../troubleshooting/README.md)
- [Sécurité](../security/README.md)
- [Déploiement](../deployment/README.md) 