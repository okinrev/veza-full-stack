# Documentation - cmd/server/main.go

## Vue d'ensemble

Le fichier `cmd/server/main.go` est le point d'entrée principal de l'application backend. Il configure et démarre le serveur HTTP avec tous les middlewares, routes, et services nécessaires.

## Fonctions Principales

### `main()`

**Description** : Fonction principale qui initialise et démarre l'application.

**Étapes d'exécution** :
1. Chargement des variables d'environnement depuis `.env`
2. Création de la configuration système
3. Configuration du mode Gin (développement/production)
4. Connexion à la base de données PostgreSQL
5. Exécution des migrations automatiques
6. Configuration du serveur de fichiers statiques React
7. Initialisation du gestionnaire WebSocket pour le chat
8. Configuration des routes API
9. Démarrage du serveur HTTP

**Code exemple** :
```go
func main() {
    // Load .env
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found")
    }

    // Configuration
    cfg := config.New()

    // Mode Gin
    if cfg.Server.Environment != "development" {
        gin.SetMode(gin.ReleaseMode)
    }

    // Database connection et migrations...
    // WebSocket setup...
    // Router configuration...
}
```

### `getProjectRoot()`

**Description** : Retourne le chemin absolu vers la racine du projet.

**Signature** :
```go
func getProjectRoot() string
```

**Utilisation** :
- Déterminer le chemin vers les fichiers statiques du frontend React
- Configuration des chemins relatifs dans l'application

**Retour** : `string` - Chemin absolu vers la racine du projet

### `setupRouter()` (fonction utilitaire)

**Description** : Configure le routeur Gin avec tous les middlewares et routes nécessaires.

**Paramètres** :
- `cfg *config.Config` - Configuration de l'application

**Retour** : `*gin.Engine` - Instance du routeur Gin configuré

**Middlewares configurés** :
- `gin.Logger()` - Logging des requêtes HTTP
- `gin.Recovery()` - Récupération des paniques
- `cors.New()` - Configuration CORS pour le frontend
- Health check endpoint sur `/health`

## Configuration du Serveur

### Serveur de Fichiers Statiques

Le serveur configure automatiquement la diffusion des fichiers du build React :

```go
// Middleware pour servir les fichiers statiques du build React
router.Static("/assets", filepath.Join(frontendPath, "assets"))
router.StaticFile("/favicon.ico", filepath.Join(frontendPath, "favicon.ico"))
```

### Configuration SPA (Single Page Application)

Pour supporter le routing côté client de React :

```go
router.NoRoute(func(c *gin.Context) {
    // Si c'est une requête API, renvoyer 404
    if strings.HasPrefix(c.Request.URL.Path, "/api/") {
        c.JSON(http.StatusNotFound, gin.H{"error": "API endpoint not found"})
        return
    }
    
    // Pour toutes les autres routes, servir index.html (SPA routing)
    indexPath := filepath.Join(frontendPath, "index.html")
    if _, err := os.Stat(indexPath); err == nil {
        c.File(indexPath)
    } else {
        // Fallback en cas d'absence du build
        c.JSON(http.StatusNotFound, gin.H{
            "error": "Frontend not built", 
            "message": "Run 'npm run build' in talas-frontend/ directory"
        })
    }
})
```

## WebSocket Configuration

### Gestionnaire de Chat

```go
// Initialiser le gestionnaire WebSocket
chatManager := websocket.NewChatManager(cfg.JWT.Secret)
go chatManager.Run()

// Route WebSocket pour le chat
router.GET("/ws/chat", func(c *gin.Context) {
    chatManager.HandleWebSocket(c)
})
```

**Fonctionnalités** :
- Authentification JWT via token
- Gestion des connexions multiples
- Diffusion de messages en temps réel
- Nettoyage automatique des connexions fermées

## Configuration CORS

```go
corsConfig := cors.Config{
    AllowOrigins:     []string{"http://localhost:3000", "http://localhost:8080", "http://localhost:5173"},
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
    ExposeHeaders:    []string{"Content-Length", "Content-Disposition"},
    AllowCredentials: true,
    MaxAge:           12 * time.Hour,
}
```

**Origines autorisées** :
- `localhost:3000` - Serveur de développement React (Create React App)
- `localhost:8080` - Serveur backend
- `localhost:5173` - Serveur de développement Vite

## Health Check

Endpoint de vérification de santé du service :

```go
router.GET("/health", func(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "status":      "ok",
        "service":     "talas-backend",
        "version":     "1.0.0",
        "timestamp":   time.Now().Unix(),
        "environment": cfg.Server.Environment,
    })
})
```

**Usage** : `GET /health`

**Réponse** :
```json
{
    "status": "ok",
    "service": "talas-backend",
    "version": "1.0.0",
    "timestamp": 1699123456,
    "environment": "development"
}
```

## Gestion d'Erreurs

### Connexion Base de Données

```go
db, err := database.NewConnection(cfg.Database.URL)
if err != nil {
    log.Fatal("Database connection failed:", err)
}
defer db.Close()
```

### Migrations

```go
if err := database.RunMigrations(db); err != nil {
    log.Printf("Migration warning: %v", err)
}
```

**Note** : Les erreurs de migration sont loggées mais n'arrêtent pas le serveur.

## Démarrage du Serveur

```go
port := cfg.Server.Port
if port == "" {
    port = "8080"
}

log.Printf("Serveur démarré sur le port %s", port)
if err := router.Run(":" + port); err != nil {
    log.Fatalf("Erreur lors du démarrage du serveur: %v", err)
}
```

## Variables d'Environnement Utilisées

- `PORT` - Port d'écoute du serveur (défaut: 8080)
- `ENVIRONMENT` - Environnement d'exécution (development/production)
- `DATABASE_URL` - URL de connexion PostgreSQL
- `JWT_SECRET` - Clé secrète pour la signature JWT

## Intégration Frontend

### Développement

En mode développement, le frontend React peut tourner sur son propre serveur (port 3000 ou 5173). Le backend gère les requêtes API et sert de fallback.

### Production

En production, le backend sert directement les fichiers statiques du build React depuis le dossier `dist/`.

## Intégration Modules Rust

Le backend Go expose les endpoints nécessaires pour que les modules Rust puissent :
- Authentifier les utilisateurs via JWT
- Accéder aux données de la base PostgreSQL
- Communiquer via WebSocket

## Points d'Attention

1. **Chemin Frontend** : Le chemin vers le frontend est calculé dynamiquement mais peut nécessiter un ajustement selon l'environnement de déploiement.

2. **CORS Production** : En production, restreindre les origines CORS aux domaines autorisés uniquement.

3. **Sécurité JWT** : Le secret JWT doit être robuste et unique en production.

4. **Ressources** : Assurer la fermeture propre des connexions base de données et WebSocket.

## Améliorations Possibles

1. **Configuration avancée** : Timeout configurable pour les connexions
2. **Monitoring** : Ajout de métriques et monitoring
3. **Graceful shutdown** : Gestion propre de l'arrêt du serveur
4. **Load balancing** : Support pour déploiement multi-instances 