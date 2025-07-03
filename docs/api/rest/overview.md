# 🌐 APIs REST Veza

## Vue d'ensemble

L'API REST Veza fournit une interface HTTP complète pour gérer tous les aspects de la plateforme : authentification, utilisateurs, chat, streaming, et administration.

## 🔗 Base URL

```
Development:  http://localhost:8080/api/v1
Staging:      https://api-staging.veza.com/api/v1
Production:   https://api.veza.com/api/v1
```

## 🔐 Authentification

### JWT Bearer Token
```http
Authorization: Bearer <jwt_token>
```

### API Key (pour les services)
```http
X-API-Key: <api_key>
```

## 📊 Codes de Réponse

| Code | Description |
|------|-------------|
| 200 | Succès |
| 201 | Créé avec succès |
| 400 | Requête invalide |
| 401 | Non authentifié |
| 403 | Non autorisé |
| 404 | Ressource non trouvée |
| 429 | Rate limit dépassé |
| 500 | Erreur serveur |

## 📝 Format des Réponses

### Succès
```json
{
  "success": true,
  "data": {
    // Données de la réponse
  },
  "message": "Opération réussie",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### Erreur
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Données invalides",
    "details": {
      "field": "email",
      "reason": "Format invalide"
    }
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

## 🔄 Pagination

Toutes les listes supportent la pagination :

```http
GET /api/v1/users?page=1&limit=20&sort=created_at&order=desc
```

**Paramètres :**
- `page` : Numéro de page (défaut: 1)
- `limit` : Nombre d'éléments par page (défaut: 20, max: 100)
- `sort` : Champ de tri
- `order` : Ordre (asc/desc)

**Réponse :**
```json
{
  "success": true,
  "data": {
    "items": [...],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 150,
      "pages": 8,
      "has_next": true,
      "has_prev": false
    }
  }
}
```

## 🔍 Filtrage et Recherche

### Filtres simples
```http
GET /api/v1/users?status=active&role=user
```

### Recherche textuelle
```http
GET /api/v1/users?search=john&search_fields=name,email
```

### Filtres avancés
```http
GET /api/v1/users?filters[created_at][gte]=2024-01-01&filters[age][between]=18,30
```

## 📁 Upload de Fichiers

### Single file
```http
POST /api/v1/files/upload
Content-Type: multipart/form-data

file: <binary_data>
```

### Multiple files
```http
POST /api/v1/files/upload/batch
Content-Type: multipart/form-data

files[]: <binary_data_1>
files[]: <binary_data_2>
```

## 🔒 Rate Limiting

- **Authentifié** : 1000 req/min par utilisateur
- **Non authentifié** : 100 req/min par IP
- **Upload** : 10 fichiers/min par utilisateur

**Headers de réponse :**
```http
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1640995200
```

## 📊 Métriques

### Endpoints de monitoring
- `GET /health` - Santé de l'API
- `GET /metrics` - Métriques Prometheus
- `GET /ready` - Readiness check

### Métriques disponibles
- **Requests/sec** par endpoint
- **Response time** (p50, p95, p99)
- **Error rate** par endpoint
- **Active users** par minute
- **Database connections** actives

## 🔗 Endpoints Principaux

### Authentication
- `POST /auth/login` - Connexion
- `POST /auth/register` - Inscription
- `POST /auth/refresh` - Rafraîchir token
- `POST /auth/logout` - Déconnexion
- `POST /auth/forgot-password` - Mot de passe oublié
- `POST /auth/reset-password` - Réinitialiser mot de passe

### Users
- `GET /users` - Liste des utilisateurs
- `GET /users/{id}` - Détails utilisateur
- `PUT /users/{id}` - Modifier utilisateur
- `DELETE /users/{id}` - Supprimer utilisateur
- `GET /users/{id}/profile` - Profil utilisateur
- `PUT /users/{id}/profile` - Modifier profil

### Chat
- `GET /rooms` - Liste des salles
- `POST /rooms` - Créer une salle
- `GET /rooms/{id}` - Détails salle
- `PUT /rooms/{id}` - Modifier salle
- `DELETE /rooms/{id}` - Supprimer salle
- `GET /rooms/{id}/messages` - Messages d'une salle
- `POST /rooms/{id}/messages` - Envoyer message

### Stream
- `GET /streams` - Liste des streams
- `POST /streams` - Créer un stream
- `GET /streams/{id}` - Détails stream
- `PUT /streams/{id}` - Modifier stream
- `DELETE /streams/{id}` - Supprimer stream
- `POST /streams/{id}/start` - Démarrer stream
- `POST /streams/{id}/stop` - Arrêter stream

### Admin
- `GET /admin/analytics` - Analytics globales
- `GET /admin/users` - Gestion utilisateurs
- `GET /admin/rooms` - Gestion salles
- `GET /admin/streams` - Gestion streams
- `GET /admin/logs` - Logs système

## 🔧 SDKs et Clients

### JavaScript/TypeScript
```bash
npm install @veza/api-client
```

```javascript
import { VezaClient } from '@veza/api-client';

const client = new VezaClient({
  baseURL: 'https://api.veza.com/api/v1',
  token: 'your-jwt-token'
});

// Utilisation
const users = await client.users.list();
const user = await client.users.get(123);
```

### Python
```bash
pip install veza-api-client
```

```python
from veza_api import VezaClient

client = VezaClient(
    base_url='https://api.veza.com/api/v1',
    token='your-jwt-token'
)

# Utilisation
users = client.users.list()
user = client.users.get(123)
```

### Go
```bash
go get github.com/veza/api-client-go
```

```go
import "github.com/veza/api-client-go"

client := veza.NewClient("https://api.veza.com/api/v1", "your-jwt-token")

// Utilisation
users, err := client.Users.List()
user, err := client.Users.Get(123)
```

## 🧪 Tests

### Collection Postman
```bash
# Importer la collection
curl -X GET https://api.veza.com/postman/collection.json
```

### Tests automatisés
```bash
# Lancer les tests d'intégration
make test-api

# Tests avec données de test
make test-api-with-fixtures
```

## 📚 Documentation Interactive

### Swagger UI
```
https://api.veza.com/docs
```

### ReDoc
```
https://api.veza.com/docs/redoc
```

## 🔄 Webhooks

### Configuration
```http
POST /api/v1/webhooks
{
  "url": "https://your-app.com/webhook",
  "events": ["user.created", "message.sent"],
  "secret": "your-webhook-secret"
}
```

### Événements disponibles
- `user.created` - Nouvel utilisateur
- `user.updated` - Utilisateur modifié
- `user.deleted` - Utilisateur supprimé
- `room.created` - Nouvelle salle
- `room.updated` - Salle modifiée
- `message.sent` - Nouveau message
- `stream.started` - Stream démarré
- `stream.stopped` - Stream arrêté

## 🚀 Migration et Versioning

### Versioning
- **Version actuelle** : v1
- **Support** : 2 versions majeures
- **Dépréciation** : 6 mois d'avis

### Migration guide
```http
GET /api/v1/migration-guide
```

---

*Dernière mise à jour : 2024-01-01*
*Version API : v1* 