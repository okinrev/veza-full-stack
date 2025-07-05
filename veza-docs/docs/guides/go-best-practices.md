# Bonnes Pratiques Go - Veza Platform

## Vue d'ensemble

Ce guide présente les conventions, astuces et pièges à éviter pour écrire du code Go robuste, lisible et performant dans le contexte Veza.

## 🏆 Principes de Base
- Utiliser gofmt/golint pour le formatage
- Privilégier la simplicité et la clarté
- Favoriser la composition sur l'héritage
- Gérer explicitement les erreurs
- Utiliser les contextes pour la gestion du cycle de vie

## ✍️ Exemples

### Gestion d'erreur idiomatique
```go
result, err := DoSomething()
if err != nil {
    return nil, fmt.Errorf("échec DoSomething: %w", err)
}
```

### Utilisation des contextes
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
```

## ✅ Bonnes Pratiques
- Toujours vérifier les erreurs
- Utiliser des types explicites
- Préférer les slices aux arrays
- Documenter chaque fonction/export ([documentation-standards.md](./documentation-standards.md))
- Structurer le projet par feature/module
- Utiliser les channels pour la concurrence
- Protéger les accès concurrents (mutex, atomic)
- Écrire des tests avec testing et testify

## ⚠️ Pièges à Éviter
- Panic non contrôlé
- Utilisation abusive des variables globales
- Oublier de fermer les ressources (fichiers, connexions)
- Boucles infinies sans timeout
- Utilisation de l'interface{} sans contrôle de type

## 🔗 Liens Utiles
- [backend-development.md](./backend-development.md)
- [debugging.md](./debugging.md)
- [performance-profiling.md](./performance-profiling.md)

---

**Dernière mise à jour** : $(date)
**Version** : 1.0.0 