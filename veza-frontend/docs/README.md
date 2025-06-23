# Documentation Complète - Frontend React Talas

## Vue d'ensemble

Cette documentation fournit une description complète de l'architecture, des composants, des services et de l'intégration du frontend React Talas. Elle est conçue pour faciliter l'intégration avec :

- **Backend API Go** : Serveur principal avec endpoints REST et WebSocket
- **Module Chat Rust** : Service de chat en temps réel avec WebSocket
- **Module Stream Rust** : Service de streaming audio/vidéo

## Architecture Globale

```
Frontend React (TypeScript)
├── App Layer          → Interface utilisateur et routing
├── Feature Layer      → Modules métier (auth, chat, products, tracks)
├── Shared Layer       → Services, composants et utilitaires partagés
└── Integration Layer  → API clients et WebSocket managers
```

## Intégration Backend

### 1. Backend Go (API REST)
- **Base URL** : `http://localhost:8080`
- **Authentification** : JWT avec access/refresh tokens
- **Endpoints** : Voir `/docs/api/endpoints.md`
- **Types** : Voir `/docs/api/types.md`

### 2. Module Chat Rust (WebSocket)
- **URL WebSocket** : `ws://localhost:8081/ws`
- **Protocole** : Messages JSON avec types définis
- **Events** : Voir `/docs/chat/websocket-protocol.md`

### 3. Module Stream Rust (WebSocket/WebRTC)
- **URL WebSocket** : `ws://localhost:8082/ws`
- **Protocole** : WebRTC signaling + streaming
- **Types** : Voir `/docs/streaming/protocol.md`

## Structure de la Documentation

### 📁 `/docs/architecture/`
- `overview.md` - Vue d'ensemble de l'architecture
- `layers.md` - Description des couches applicatives
- `data-flow.md` - Flux de données dans l'application

### 📁 `/docs/api/`
- `client.md` - Configuration du client API
- `endpoints.md` - Tous les endpoints disponibles
- `types.md` - Types TypeScript pour l'API
- `authentication.md` - Gestion de l'authentification

### 📁 `/docs/features/`
- `auth/` - Module d'authentification
- `chat/` - Module de chat
- `products/` - Module de gestion des produits
- `tracks/` - Module de gestion des pistes audio
- `profile/` - Module de profil utilisateur

### 📁 `/docs/shared/`
- `components/` - Composants UI partagés
- `hooks/` - Hooks React personnalisés
- `services/` - Services et utilitaires
- `stores/` - Gestion d'état avec Zustand

### 📁 `/docs/chat/`
- `websocket-protocol.md` - Protocole WebSocket pour le chat
- `message-types.md` - Types de messages supportés
- `integration.md` - Intégration avec le module Rust

### 📁 `/docs/streaming/`
- `protocol.md` - Protocole de streaming
- `webrtc-integration.md` - Intégration WebRTC
- `audio-handling.md` - Gestion des pistes audio

### 📁 `/docs/configuration/`
- `environment.md` - Variables d'environnement
- `build.md` - Configuration de build
- `deployment.md` - Guide de déploiement

### 📁 `/docs/integration/`
- `backend-go.md` - Intégration avec le backend Go
- `chat-rust.md` - Intégration avec le module chat Rust
- `stream-rust.md` - Intégration avec le module stream Rust

## Quick Start pour Intégration

### 1. Configuration Backend Go

```go
// Exemple de configuration CORS pour le backend Go
func setupCORS() {
    c.Header("Access-Control-Allow-Origin", "http://localhost:5173")
    c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
    c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
    c.Header("Access-Control-Allow-Credentials", "true")
}
```

### 2. Configuration Module Chat Rust

```rust
// Configuration WebSocket pour le chat
let addr = "127.0.0.1:8081";
let listener = TcpListener::bind(&addr).await?;
println!("Chat WebSocket server running on: {}", addr);
```

### 3. Variables d'Environnement Frontend

```env
VITE_API_URL=http://localhost:8080
VITE_WS_CHAT_URL=ws://localhost:8081/ws
VITE_WS_STREAM_URL=ws://localhost:8082/ws
```

## Conventions de Code

- **TypeScript** : Types stricts, interfaces bien définies
- **React** : Composants fonctionnels avec hooks
- **Styling** : Tailwind CSS + shadcn/ui
- **State Management** : Zustand pour l'état global
- **API** : Axios avec intercepteurs pour l'authentification

## Flux de Données Typiques

### Authentification
```
LoginForm → authStore.login() → API Backend → JWT Storage → Navigation
```

### Chat en Temps Réel
```
ChatInput → ChatService → WebSocket Rust → Backend Sync → UI Update
```

### Upload de Fichiers
```
FileUpload → API Client → Backend Go → Storage → Metadata Update
```

---

**Note** : Cette documentation est mise à jour en continu. Chaque fichier de code dispose de sa propre documentation détaillée dans les dossiers correspondants. 