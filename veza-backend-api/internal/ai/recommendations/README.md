# 🤖 AI Recommendations Engine

## Vue d'ensemble

Le moteur de recommandations AI de Veza génère des recommandations personnalisées de pistes, artistes et samples basées sur le profil utilisateur et l'historique d'activité.

## Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   API Handler   │───▶│ Recommendation   │───▶│   ML Model      │
│                 │    │    Engine        │    │                 │
└─────────────────┘    └──────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Cache Redis   │    │ User Profile     │    │ Track Analytics │
│                 │    │   Service        │    │   Service       │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

## Composants

### 1. RecommendationEngine
- **Rôle** : Orchestrateur principal du système de recommandations
- **Fonctionnalités** :
  - Génération de recommandations personnalisées
  - Mise à jour des profils utilisateur
  - Gestion du cache
  - Calcul de similarité

### 2. SimpleMLModel
- **Rôle** : Modèle ML basique pour la prédiction des préférences
- **Fonctionnalités** :
  - Prédiction des préférences utilisateur
  - Calcul d'embeddings de pistes
  - Calcul de similarité cosinus
  - Analyse des activités récentes

### 3. RedisRecommendationCache
- **Rôle** : Cache Redis pour les recommandations
- **Fonctionnalités** :
  - Mise en cache des recommandations (TTL: 30min)
  - Invalidation par utilisateur
  - Statistiques du cache

### 4. Handler HTTP
- **Rôle** : API REST pour les recommandations
- **Endpoints** :
  - `GET /api/v1/recommendations` - Obtenir des recommandations
  - `POST /api/v1/recommendations/users/{user_id}/activity` - Mettre à jour l'activité
  - `GET /api/v1/recommendations/users/{user_id}/profile` - Obtenir le profil
  - `GET /api/v1/recommendations/stats` - Statistiques

## Algorithmes

### 1. Prédiction des Préférences
```go
// Analyse des activités récentes (30 derniers jours)
// Calcul des poids par genre basé sur la fréquence
// Analyse du style de collaboration
// Détermination du niveau de compétence
```

### 2. Calcul de Similarité
```go
// Similarité cosinus entre embeddings
// Poids par genre (40%)
// Poids par mood (30%)
// Poids par instruments (20%)
// Similarité d'embedding (10%)
```

### 3. Filtrage et Tri
```go
// Application des filtres (genre, durée, BPM, etc.)
// Tri par score combiné :
//   - Similarité (50%)
//   - Popularité (30%)
//   - Fraîcheur (20%)
```

## API Usage

### Obtenir des recommandations
```bash
GET /api/v1/recommendations?user_id=123&context=discovery&limit=20&freshness=0.5
```

**Paramètres :**
- `user_id` (requis) : ID de l'utilisateur
- `context` (optionnel) : "discovery", "collaboration", "production", "learning"
- `limit` (optionnel) : Nombre de recommandations (1-100)
- `freshness` (optionnel) : Fraîcheur (0-1, 0=populaire, 1=récent)
- `filters` (optionnel) : Filtres JSON

**Réponse :**
```json
{
  "tracks": [
    {
      "track_id": 123,
      "title": "Amazing Track",
      "artist": "Great Artist",
      "genre": "electronic",
      "mood": "energetic",
      "similarity": 0.85,
      "popularity": 0.7,
      "freshness": 0.5,
      "collaboration": true,
      "reason": "Basé sur votre intérêt pour l'electronic"
    }
  ],
  "artists": [...],
  "samples": [...],
  "playlists": [...],
  "confidence": 0.75,
  "context": "discovery",
  "generated_at": "2024-01-15T10:30:00Z"
}
```

### Mettre à jour l'activité
```bash
POST /api/v1/recommendations/users/123/activity
Content-Type: application/json

{
  "type": "listen",
  "track_id": 456,
  "artist_id": 789,
  "genre": "electronic",
  "timestamp": "2024-01-15T10:30:00Z",
  "duration": 180,
  "interaction": 0.8
}
```

## Métriques et Monitoring

### Métriques Clés
- **Confiance moyenne** : Qualité des recommandations
- **Hit rate du cache** : Performance du cache
- **Temps de réponse** : Performance API
- **Taux de conversion** : Utilisateurs qui cliquent sur les recommandations

### Logs Structurés
```json
{
  "level": "info",
  "message": "🎯 Generated recommendations",
  "user_id": 123,
  "context": "discovery",
  "tracks_count": 15,
  "confidence": 0.75,
  "response_time_ms": 150
}
```

## Configuration

### Variables d'environnement
```env
# Cache Redis
REDIS_URL=redis://localhost:6379
REDIS_DB=0

# ML Model
ML_MODEL_TYPE=simple
ML_EMBEDDING_DIM=128

# Cache TTL
RECOMMENDATIONS_CACHE_TTL=30m
```

### Configuration par défaut
```go
type RecommendationConfig struct {
    CacheTTL           time.Duration `json:"cache_ttl"`
    MaxRecommendations int           `json:"max_recommendations"`
    MinConfidence      float64       `json:"min_confidence"`
    EnableCache        bool          `json:"enable_cache"`
}
```

## Tests

### Tests Unitaires
```bash
cd veza-backend-api/internal/ai/recommendations
go test -v
```

### Tests d'Intégration
```bash
# Tester avec Redis réel
go test -v -tags=integration
```

## Roadmap

### Phase 1 (Actuel)
- ✅ Moteur de recommandations basique
- ✅ Cache Redis
- ✅ API REST
- ✅ Tests unitaires

### Phase 2 (Prochain)
- 🔄 Modèle ML avancé (TensorFlow/PyTorch)
- 🔄 Embeddings réels des pistes
- 🔄 A/B testing
- 🔄 Dashboard analytics

### Phase 3 (Futur)
- 📋 Recommandations temps-réel
- 📋 Collaborative filtering
- 📋 Deep learning embeddings
- 📋 Multi-modal (audio + metadata)

## Performance

### Objectifs
- **Temps de réponse** : < 200ms P95
- **Hit rate cache** : > 80%
- **Précision** : > 85%
- **Disponibilité** : > 99.9%

### Optimisations
- Cache Redis avec TTL adaptatif
- Calculs asynchrones pour les embeddings
- Indexation des métadonnées
- Compression des réponses

## Sécurité

### Bonnes Pratiques
- Validation des paramètres d'entrée
- Rate limiting sur les API
- Sanitisation des données utilisateur
- Logs d'audit pour les recommandations

### GDPR Compliance
- Anonymisation des données sensibles
- Droit à l'oubli pour les profils utilisateur
- Consentement pour l'utilisation des données
- Chiffrement des données en transit

## Support

### Debugging
```bash
# Activer les logs détaillés
export LOG_LEVEL=debug

# Vérifier le cache Redis
redis-cli keys "recommendations:*"

# Tester l'API
curl -X GET "http://localhost:8080/api/v1/recommendations?user_id=123"
```

### Monitoring
- Métriques Prometheus
- Logs structurés avec Zap
- Alertes sur les erreurs
- Dashboard Grafana

---

*Documentation créée par le Lead Innovation Engineer*  
*Dernière mise à jour : 2024-01-15* 