# Plan de Migration Backend Talas - Vue d'ensemble

## 🎯 Objectif
Reconstruire un backend Go totalement fonctionnel à partir de la structure actuelle fragmentée.

## 📁 Structure des fichiers de migration
```
backend/docs/migration/
├── 00_MASTER_PLAN.md           # Ce fichier (vue d'ensemble)
├── 01_consolidation_archi.md   # Étape 1 : Architecture
├── 02_correction_imports.md    # Étape 2 : Imports  
├── 03_implementation_handlers.md # Étape 3 : Handlers
├── 04_consolidation_services.md # Étape 4 : Services
├── 05_refactoring_routes.md    # Étape 5 : Routes
├── 06_tests_validation.md      # Étape 6 : Tests
└── scripts/                    # Scripts utilitaires
    ├── check_imports.sh
    ├── clean_duplicates.sh
    └── validate_build.sh
```

## 🚨 Problèmes identifiés

### Critiques (bloquent la compilation)
- ❌ Cycles d'import entre routes/handlers
- ❌ Imports incorrects (`veza-web-app/` vs `github.com/okinrev/veza-web-app/`)
- ❌ Double architecture (legacy + modulaire)
- ❌ Handlers API incomplets (TODO partout)

### Graves (dégradent la qualité)
- ⚠️ Doublons de services/handlers
- ⚠️ Routing complexe et fragile
- ⚠️ Configuration dispersée
- ⚠️ Middlewares incohérents

## 📊 Architecture actuelle vs cible

### 🔴 Actuelle (problématique)
```
cmd/server/main.go (hybride complexe)
├── internal/routes/ (legacy, non importé)
├── internal/handlers/ (legacy, documenté)
├── internal/api/ (nouveau, incomplet)
├── internal/services/ (doublons)
└── internal/models/ (ok)
```

### 🟢 Cible (recommandée)
```
cmd/server/main.go (simple)
├── internal/api/ (modules complets)
│   ├── auth/ (handler + service + routes)
│   ├── user/ (handler + service + routes)
│   └── [...]/
├── internal/common/ (partagé)
├── internal/config/ (centralisé)
└── internal/models/ (unifié)
```

## ⏱️ Planning estimé
- **Étape 1** : 30-45 min (architecture)
- **Étape 2** : 15-20 min (imports)
- **Étape 3** : 60-90 min (handlers)
- **Étape 4** : 30-45 min (services)
- **Étape 5** : 45-60 min (routes)
- **Étape 6** : 20-30 min (tests)

**Total estimé** : 3h30 - 4h30

## 🔄 Ordre d'exécution STRICT

1. **ÉTAPE 1** → Consolider l'architecture (base saine)
2. **ÉTAPE 2** → Corriger les imports (compilation possible)
3. **ÉTAPE 3** → Implémenter les handlers (fonctionnalités)
4. **ÉTAPE 4** → Consolider les services (éliminer doublons)
5. **ÉTAPE 5** → Refactorer les routes (système unifié)
6. **ÉTAPE 6** → Tester et valider (stabilité)

## ✅ Validation entre étapes

Après chaque étape, vérifier :
```bash
# Compilation
go mod tidy
go build ./cmd/server

# Structure
find . -name "*.go" | wc -l  # Comptage fichiers
go list -m all               # Modules

# Tests basiques
go test ./...
```

## 🆘 Points de rollback

Si une étape échoue :
1. **Git commit** avant chaque étape
2. **Backup** des fichiers modifiés
3. **Documentation** des changements dans `CHANGELOG.md`

## 📝 Notes importantes

- ⚠️ Ne pas mélanger les étapes
- ⚠️ Tester la compilation après chaque modification majeure
- ⚠️ Garder la documentation à jour
- ⚠️ Sauvegarder avant de commencer

---

**▶️ PROCHAINE ÉTAPE** : Lire `01_consolidation_archi.md`