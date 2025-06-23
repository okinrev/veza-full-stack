# Stream Server - Documentation Complète

## Vue d'ensemble

Stream Server est un serveur de streaming audio haute performance écrit en Rust, conçu pour être intégré dans des architectures modernes avec des backends API (Go, Node.js, Python, etc.) et des frontends web (React, Vue, Angular, etc.).

## Architecture Globale

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │   Backend API   │    │   Stream        │
│   (React)       │◄──►│   (Go/Node/...)  │◄──►│   Server        │
│                 │    │                 │    │   (Rust)        │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                                        │
        ┌───────────────────────────────────────────────┼───────────────────────────────────────────────┐
        │                                               │                                               │
        ▼                                               ▼                                               ▼
┌─────────────────┐                           ┌─────────────────┐                           ┌─────────────────┐
│   SQLite        │                           │   File System   │                           │   Redis         │
│   (Analytics)   │                           │   (Audio Files) │                           │   (Optional)    │
└─────────────────┘                           └─────────────────┘                           └─────────────────┘
```

## Composants Principaux

### 1. **Configuration et Environment** (`src/config/`)
- Configuration centralisée via variables d'environnement
- Validation automatique des paramètres
- Support multi-environnement (dev, staging, production)

### 2. **Authentification et Autorisation** (`src/auth/`)
- JWT tokens avec refresh automatique
- Système de rôles et permissions granulaires
- Protection contre les attaques courantes

### 3. **Streaming Audio** (`src/streaming/`)
- Streaming adaptatif basé sur la bande passante
- Support Range Requests pour le seek
- WebSocket pour les événements temps réel
- Génération de playlists HLS

### 4. **Traitement Audio** (`src/audio/`)
- Extraction de métadonnées automatique
- Génération de waveforms
- Compression multi-format
- Analyse spectrale

### 5. **Analytics et Monitoring** (`src/analytics/`, `src/health/`)
- Tracking des sessions d'écoute
- Métriques de performance en temps réel
- Health checks détaillés
- Alertes automatiques

### 6. **Notifications** (`src/notifications/`)
- Multi-canal (WebSocket, Email, SMS, Push)
- Templates personnalisables
- Gestion de la fréquence et des heures de silence

### 7. **Cache et Performance** (`src/cache/`, `src/utils/`)
- Cache LRU pour les métadonnées
- Optimisations de performance
- Métriques de monitoring

## Flux de Données Typique

### 1. **Authentification**
```
Frontend → Backend API → Stream Server (validation JWT)
```

### 2. **Génération d'URL Signée**
```
Frontend → Backend API (génère signature) → Frontend (URL signée)
```

### 3. **Streaming Audio**
```
Frontend → Stream Server (avec URL signée) → Audio Stream
```

### 4. **Events Temps Réel**
```
Stream Server → WebSocket → Frontend (événements de lecture)
```

## Points d'Intégration

### Backend API (Go/Node.js/Python)
- **Authentification** : Validation des utilisateurs et génération de JWT
- **Autorisation** : Vérification des permissions d'accès aux fichiers
- **Signature d'URLs** : Génération des signatures temporelles sécurisées
- **Analytics** : Récupération des statistiques d'écoute
- **Administration** : Gestion des utilisateurs et du contenu

### Frontend (React/Vue/Angular)
- **Player Audio** : Intégration du streaming avec gestion des erreurs
- **Interface Utilisateur** : Affichage des métadonnées et waveforms
- **WebSocket** : Réception des événements temps réel
- **Analytics** : Tracking des interactions utilisateur

## Sécurité

### Authentification Multi-Niveaux
1. **JWT Tokens** pour l'authentification utilisateur
2. **URLs Signées** avec expiration temporelle
3. **Rate Limiting** par IP et par utilisateur
4. **CORS** configuré selon l'environnement

### Protection des Données
- Validation stricte de tous les inputs
- Protection contre les attaques par traversée de répertoire
- Headers de sécurité automatiques
- Chiffrement des communications sensibles

## Performance

### Optimisations Côté Serveur
- **Streaming par chunks** pour les gros fichiers
- **Cache LRU** pour les métadonnées fréquemment accédées
- **Workers parallèles** pour le traitement audio
- **Compression HTTP** automatique

### Optimisations Côté Client
- **Range Requests** pour le seek instantané
- **Streaming adaptatif** selon la bande passante
- **Mise en cache** des métadonnées côté client
- **Preloading** intelligent

## Monitoring et Observabilité

### Métriques Disponibles
- Nombre de streams actifs
- Bande passante utilisée
- Temps de réponse
- Taux d'erreur
- Usage des ressources système

### Health Checks
- `/health` - Status basique
- `/health/detailed` - Diagnostics complets
- `/metrics` - Métriques Prometheus

### Logging
- Logs structurés avec niveaux configurables
- Tracking des événements de sécurité
- Audit des accès aux fichiers

## Déploiement

### Environnements Supportés
- **Développement** : Configuration simplifiée, logs verbeux
- **Staging** : Configuration proche de la production
- **Production** : Sécurité maximale, performance optimisée

### Options de Déploiement
- **Docker** : Containerisation complète
- **Docker Compose** : Stack complète avec dépendances
- **Kubernetes** : Déploiement scalable
- **Binaire Standalone** : Déploiement direct sur serveur

## Structure de la Documentation

### Modules Détaillés
- [`docs/modules/config.md`](./modules/config.md) - Configuration système
- [`docs/modules/auth.md`](./modules/auth.md) - Authentification et autorisation
- [`docs/modules/streaming.md`](./modules/streaming.md) - Streaming audio
- [`docs/modules/audio.md`](./modules/audio.md) - Traitement audio
- [`docs/modules/analytics.md`](./modules/analytics.md) - Analytics et métriques
- [`docs/modules/notifications.md`](./modules/notifications.md) - Système de notifications
- [`docs/modules/health.md`](./modules/health.md) - Monitoring et santé
- [`docs/modules/cache.md`](./modules/cache.md) - Système de cache
- [`docs/modules/middleware.md`](./modules/middleware.md) - Middlewares de sécurité
- [`docs/modules/utils.md`](./modules/utils.md) - Utilitaires et helpers

### Guides d'Intégration
- [`docs/integration/go-backend.md`](./integration/go-backend.md) - Intégration backend Go
- [`docs/integration/react-frontend.md`](./integration/react-frontend.md) - Intégration frontend React
- [`docs/integration/architecture.md`](./integration/architecture.md) - Architecture complète

### Référence API
- [`docs/api/endpoints.md`](./api/endpoints.md) - Documentation complète des endpoints
- [`docs/api/websockets.md`](./api/websockets.md) - API WebSocket
- [`docs/api/authentication.md`](./api/authentication.md) - Authentification
- [`docs/api/errors.md`](./api/errors.md) - Codes d'erreur et gestion

### Exemples Pratiques
- [`docs/examples/curl-examples.md`](./examples/curl-examples.md) - Tests avec cURL
- [`docs/examples/go-examples.md`](./examples/go-examples.md) - Code Go d'exemple
- [`docs/examples/react-examples.md`](./examples/react-examples.md) - Code React d'exemple
- [`docs/examples/deployment.md`](./examples/deployment.md) - Exemples de déploiement

## Support et Maintenance

### Versioning
Le projet suit le versioning sémantique (SemVer) :
- **Major** : Changements incompatibles de l'API
- **Minor** : Nouvelles fonctionnalités compatibles
- **Patch** : Corrections de bugs

### Migration
- Documentation des changements entre versions
- Scripts de migration pour les bases de données
- Guide de mise à jour des intégrations

### Troubleshooting
- Guide de diagnostic des problèmes courants
- Logs détaillés pour le debugging
- Métriques pour identifier les goulots d'étranglement

---

**Version** : 0.2.0  
**Dernière mise à jour** : $(date '+%Y-%m-%d')  
**Licence** : MIT 