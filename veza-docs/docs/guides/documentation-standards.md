# Standards de Documentation - Veza Platform

## Vue d'ensemble

Ce guide définit les standards de documentation pour tous les projets Veza, incluant le format, la structure, la profondeur et les outils recommandés.

## 📚 Principes Généraux
- Chaque fichier doit avoir un en-tête descriptif
- Chaque fonction/export doit avoir un JSDoc/TSDoc
- Les composants doivent documenter props et state
- Les endpoints API doivent avoir des exemples et schémas OpenAPI
- Les modèles de données doivent être documentés
- Les variables doivent être typées et commentées

## 📝 Exemples

### En-tête de fichier
```go
// Fichier : user_service.go
// Service de gestion des utilisateurs pour Veza.
// Fournit les opérations CRUD et la gestion des rôles.
```

### JSDoc pour fonction
```ts
/**
 * Calcule la somme de deux nombres.
 * @param a Premier nombre
 * @param b Deuxième nombre
 * @returns Somme
 * @example
 *   add(2, 3) // 5
 */
function add(a: number, b: number): number { return a + b; }
```

### OpenAPI pour endpoint
```yaml
paths:
  /users:
    get:
      summary: Liste les utilisateurs
      responses:
        '200':
          description: Succès
```

## ✅ Bonnes Pratiques
- Utiliser un langage clair et précis
- Illustrer avec des exemples concrets
- Croiser les références internes ([code-quality.md](./code-quality.md))
- Mettre à jour la documentation à chaque modification
- Générer la documentation automatiquement (Docusaurus, Swagger)

## ⚠️ Pièges à Éviter
- Documentation obsolète ou incomplète
- Absence d'exemples
- Manque de liens croisés
- Oublier de documenter les erreurs et cas limites

## 🔗 Liens Utiles
- [code-quality.md](./code-quality.md)
- [refactoring-guide.md](./refactoring-guide.md)
- [api-testing.md](./api-testing.md)
- [debugging.md](./debugging.md)

---

**Dernière mise à jour** : $(date)
**Version** : 1.0.0 