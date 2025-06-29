# Guide d'Intégration Talas - Architecture Unifiée

## 🎯 Vue d'Ensemble

Talas est une plateforme unifiée composée de 4 modules principaux avec authentification JWT centralisée et communication fluide entre tous les services.

### Modules Principaux

1. **Backend Go** (Port 8080) - API REST principale, gestion utilisateurs, authentification
2. **Frontend React** (Port 5173) - Interface utilisateur moderne, TypeScript + Vite
3. **Chat Server Rust** (Port 3001) - WebSocket temps réel pour messagerie
4. **Stream Server Rust** (Port 3002) - WebSocket audio streaming

### Architecture Générale

```
┌─────────────────┐    ┌─────────────────┐
│  Frontend React │    │   Backend Go    │
│   Port 5173     │────│   Port 8080     │
└─────────────────┘    └─────────────────┘
         │                       │
         │              ┌─────────────────┐
         │              │   PostgreSQL    │
         │              │   Port 5432     │
         │              └─────────────────┘
         │                       │
┌─────────────────┐    ┌─────────────────┐
│  Chat Rust WS   │    │ Stream Rust WS  │
│   Port 3001     │    │   Port 3002     │
└─────────────────┘    └─────────────────┘
```

## 🔐 Authentification JWT Unifiée

### Configuration Partagée

Tous les services utilisent la configuration JWT centralisée :

**Fichier : `configs/jwt.config`**
```bash
JWT_SECRET=veza_unified_jwt_secret_key_2025_microservices_secure_32chars_minimum
JWT_ISSUER=veza-platform
JWT_AUDIENCE=veza-services
JWT_ALGORITHM=HS256
JWT_EXPIRATION=24h
JWT_REFRESH_EXPIRATION=168h
```

### Flux d'Authentification

1. **Login Frontend** → **Backend Go** : Génération token JWT
2. **Frontend** → **Chat/Stream WebSocket** : Authentification automatique avec token
3. **Toutes les requêtes** incluent le token JWT dans les headers
4. **Validation unifiée** dans tous les services

## 🚀 Guide de Déploiement

### 1. Configuration Initiale

```bash
# Configuration JWT unifiée
./scripts/talas-admin.sh setup
```

### 2. Compilation

```bash
# Compiler tous les services
./scripts/talas-admin.sh build
```

### 3. Démarrage

```bash
# Démarrer la plateforme complète
./scripts/talas-admin.sh start

# Vérifier l'état
./scripts/talas-admin.sh status
```

### 4. Tests

```bash
# Tests d'intégration
./scripts/talas-admin.sh test
```

## 🔧 Administration

### Commandes Principales

```bash
./scripts/talas-admin.sh setup    # Configuration initiale
./scripts/talas-admin.sh start    # Démarrer tous les services
./scripts/talas-admin.sh stop     # Arrêter tous les services
./scripts/talas-admin.sh status   # État des services
./scripts/talas-admin.sh logs     # Voir les logs
./scripts/talas-admin.sh restart  # Redémarrer
./scripts/talas-admin.sh clean    # Nettoyer
```

### Monitoring

- **Logs centralisés** : `logs/*.log`
- **Status en temps réel** : `talas-admin.sh status`
- **Health checks** : Endpoints `/health` sur chaque service

## 🌐 Communication Inter-Services

### Frontend → Backend
- **Protocole** : REST API via HTTP
- **Authentification** : JWT Bearer token
- **Format** : JSON

### Frontend → Chat/Stream
- **Protocole** : WebSocket
- **Authentification** : JWT dans message initial
- **Format** : JSON messages

### Exemple de Flow Complet

1. Utilisateur se connecte via Frontend
2. Backend génère JWT et retourne token
3. Frontend stocke token et se connecte aux WebSockets
4. Chat/Stream valident le JWT et acceptent la connexion
5. Communication fluide entre tous les services

## 🧪 Tests d'Intégration

### Test Manuel Complet

```bash
# 1. Démarrer tous les services
./scripts/talas-admin.sh start

# 2. Vérifier les endpoints
curl http://localhost:8080/health
curl http://localhost:3001/health
curl http://localhost:3002/health
curl http://localhost:5173/

# 3. Test d'authentification
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password"}'
```

### Test WebSocket

Le service WebSocket unifié (`websocketService.ts`) gère automatiquement :
- Connexion avec authentification JWT
- Reconnexion automatique
- Heartbeat
- Gestion des erreurs

## 📊 Métriques

### Performance Attendue

- **Démarrage** : < 10 secondes pour tous les services
- **Authentification** : < 100ms
- **WebSocket connexion** : < 1 seconde
- **API REST** : < 200ms

### Vérification de Santé

```bash
./scripts/talas-admin.sh status
```

Affiche :
- État de chaque service (Actif/Arrêté)
- Ports utilisés
- PIDs des processus

## 🚨 Dépannage

### Problèmes Courants

1. **Service ne démarre pas** : Vérifier les logs avec `logs <service>`
2. **Authentification échoue** : Vérifier la configuration JWT
3. **WebSocket ne connecte pas** : Vérifier les URLs dans `.env.local`

### Configuration de Debug

Le frontend inclut des logs détaillés pour le debug :
```typescript
// Logs automatiques dans la console du navigateur
🔵 [API] GET /api/v1/profile
🔌 [Talas WebSocket] Connexion au service chat
✅ [Talas WebSocket] Authentification chat réussie
```

---

## 🎉 Résultat Final

Une plateforme Talas complètement intégrée où :

✅ Un utilisateur peut se connecter une fois et accéder à tous les services  
✅ L'authentification JWT fonctionne de manière transparente  
✅ Les WebSocket se connectent automatiquement  
✅ La communication inter-services est fluide  
✅ L'administration est centralisée via `talas-admin.sh`  

**L'expérience utilisateur est unifiée et sans friction entre tous les modules.**
