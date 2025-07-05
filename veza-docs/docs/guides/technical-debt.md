# Guide de la Dette Technique - Veza Platform

## Vue d'ensemble

Ce guide explique la notion de dette technique, ses causes, ses conséquences et les stratégies pour la gérer efficacement dans la plateforme Veza.

## 💡 Qu'est-ce que la Dette Technique ?
La dette technique désigne les compromis faits lors du développement qui facilitent une livraison rapide mais qui, à terme, complexifient la maintenance et l'évolution du code.

## 🏷️ Exemples de Dette Technique
- Code dupliqué ou non factorisé
- Absence de tests automatisés ([api-testing.md](./api-testing.md))
- Documentation incomplète ([documentation-standards.md](./documentation-standards.md))
- Utilisation de dépendances obsolètes
- Contournements temporaires non corrigés

## ✅ Bonnes Pratiques
- Documenter toute dette technique dans les PR et tickets
- Prioriser le remboursement de la dette critique
- Planifier des sprints de refactoring ([refactoring-guide.md](./refactoring-guide.md))
- Mettre en place des outils de suivi (SonarQube, backlog Jira)
- Sensibiliser l'équipe à l'impact de la dette

## ⚠️ Pièges à Éviter
- Ignorer la dette accumulée
- Reporter indéfiniment le remboursement
- Ne pas mesurer l'impact sur la vélocité
- Laisser la dette s'accumuler dans les modules critiques

## 🔗 Liens Utiles
- [refactoring-guide.md](./refactoring-guide.md)
- [code-quality.md](./code-quality.md)
- [api-testing.md](./api-testing.md)
- [debugging.md](./debugging.md)

---

**Dernière mise à jour** : $(date)
**Version** : 1.0.0 