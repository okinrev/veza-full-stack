# Guide Développement Backend - Veza Platform

## Vue d'ensemble

Ce guide présente les standards, outils et bonnes pratiques pour le développement backend sur la plateforme Veza (Go, PostgreSQL, Redis, NATS, etc.).

## 🏗️ Architecture Backend
- Architecture hexagonale (voir [architecture-decisions.md](./architecture-decisions.md))
- Séparation claire entre domaine, application, infrastructure
- API RESTful et WebSocket
- Microservices et communication par événements

## ✍️ Exemples

### Structure d'un handler Go
```go
// Handler pour créer un utilisateur
func (h *UserHandler) CreateUser(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    user, err := h.service.CreateUser(req)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    c.JSON(201, user)
}
```

### Exemple de migration SQL
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT now()
);
```

## ✅ Bonnes Pratiques
- Respecter la séparation des couches
- Utiliser des interfaces pour l'injection de dépendances
- Gérer explicitement les erreurs ([debugging.md](./debugging.md))
- Écrire des tests unitaires et d'intégration ([api-testing.md](./api-testing.md))
- Documenter chaque endpoint ([documentation-standards.md](./documentation-standards.md))
- Utiliser les transactions pour la cohérence des données
- Monitorer les performances ([performance-profiling.md](./performance-profiling.md))

## ⚠️ Pièges à Éviter
- Logique métier dans les handlers
- Accès direct à la base sans repository
- Oublier la gestion des transactions
- Absence de gestion des erreurs
- Ne pas valider les entrées utilisateur

## 🔗 Liens Utiles
- [database-design.md](./database-design.md)
- [api-testing.md](./api-testing.md)
- [debugging.md](./debugging.md)
- [performance-profiling.md](./performance-profiling.md)

---

**Dernière mise à jour** : $(date)
**Version** : 1.0.0 