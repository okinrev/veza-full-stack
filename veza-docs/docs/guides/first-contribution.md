---
id: first-contribution
title: Première contribution
sidebar_label: Première contribution
description: Guide complet pour faire votre première contribution au projet Veza
---

# Première contribution

Ce guide vous accompagne étape par étape pour faire votre première contribution au projet Veza, que vous soyez développeur expérimenté ou débutant.

## Prérequis

### Compétences requises
- **Bases de programmation** : Connaissance d'au moins un langage (Go, Rust, JavaScript)
- **Git** : Maîtrise des commandes de base (clone, commit, push, pull)
- **Anglais** : Compréhension basique pour lire les issues et la documentation
- **Communication** : Capacité à expliquer vos changements

### Outils nécessaires
- **Git** : Version 2.30+
- **IDE** : VS Code, IntelliJ, ou éditeur de votre choix
- **Environnement de développement** : Voir [Guide d'environnement](./development-environment.md)

## Premiers pas

### 1. Fork du projet
```bash
# Aller sur GitHub et cliquer sur "Fork"
# https://github.com/veza/veza-full-stack

# Cloner votre fork
git clone https://github.com/VOTRE_USERNAME/veza-full-stack.git
cd veza-full-stack

# Ajouter le repository original comme upstream
git remote add upstream https://github.com/veza/veza-full-stack.git
```

### 2. Configuration de l'environnement
```bash
# Suivre le guide d'environnement de développement
# Voir: ./development-environment.md

# Vérifier que tout fonctionne
./scripts/dev-start.sh
```

### 3. Exploration du projet
```bash
# Structure du projet
tree -L 3 -I 'node_modules|target|.git'

# Lire les fichiers importants
cat README.md
cat CONTRIBUTING.md
cat CODE_OF_CONDUCT.md
```

## Choisir une première contribution

### Types de contributions recommandées pour débuter

#### 1. Documentation
- **Niveau** : Facile
- **Impact** : Élevé
- **Exemples** :
  - Corriger des fautes de frappe
  - Améliorer la clarté des explications
  - Ajouter des exemples de code
  - Traduire en français

#### 2. Tests
- **Niveau** : Facile à Moyen
- **Impact** : Moyen
- **Exemples** :
  - Ajouter des tests unitaires
  - Améliorer la couverture de tests
  - Ajouter des tests d'intégration

#### 3. Bugs simples
- **Niveau** : Moyen
- **Impact** : Élevé
- **Exemples** :
  - Corriger des erreurs de validation
  - Améliorer la gestion d'erreurs
  - Optimiser des requêtes

#### 4. Features mineures
- **Niveau** : Moyen à Difficile
- **Impact** : Moyen
- **Exemples** :
  - Ajouter des endpoints simples
  - Améliorer l'UX
  - Optimiser les performances

### Trouver des issues appropriées

#### Labels utiles
- `good first issue` : Parfait pour débuter
- `help wanted` : Besoin d'aide
- `documentation` : Amélioration de la documentation
- `bug` : Correction de bugs
- `enhancement` : Amélioration de fonctionnalités

#### Recherche d'issues
```bash
# Issues ouvertes avec le label "good first issue"
# https://github.com/veza/veza-full-stack/issues?q=is%3Aopen+is%3Aissue+label%3A%22good+first+issue%22

# Issues récentes
# https://github.com/veza/veza-full-stack/issues?q=is%3Aopen+is%3Aissue+sort%3Aupdated-desc
```

## Workflow de contribution

### 1. Créer une branche
```bash
# Synchroniser avec upstream
git fetch upstream
git checkout main
git merge upstream/main

# Créer une nouvelle branche
git checkout -b feature/ma-premiere-contribution

# Nommage des branches
# feature/nom-de-la-feature
# bugfix/nom-du-bug
# docs/amélioration-documentation
# test/ajout-tests
```

### 2. Développer votre contribution

#### Exemple : Ajouter un test
```go
// veza-backend-api/internal/services/user_service_test.go
package services

import (
    "testing"
    "context"
    "github.com/stretchr/testify/assert"
)

func TestUserService_GetUserByEmail(t *testing.T) {
    // Arrange
    mockRepo := &MockUserRepository{}
    service := NewUserService(mockRepo)
    ctx := context.Background()
    
    expectedUser := &User{
        ID:       1,
        Email:    "test@example.com",
        Username: "testuser",
    }
    
    mockRepo.On("GetByEmail", ctx, "test@example.com").Return(expectedUser, nil)
    
    // Act
    user, err := service.GetUserByEmail(ctx, "test@example.com")
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, expectedUser, user)
    mockRepo.AssertExpectations(t)
}
```

#### Exemple : Améliorer la documentation
```markdown
# veza-docs/docs/api/endpoints/users-api.md

## Authentification

### Login utilisateur

Authentifie un utilisateur avec email et mot de passe.

**Endpoint :** `POST /api/v1/auth/login`

**Corps de la requête :**
```json
{
  "email": "user@example.com",
  "password": "motdepasse123"
}
```

**Réponse de succès (200) :**
```json
{
  "user": {
    "id": 123,
    "email": "user@example.com",
    "username": "user123"
  },
  "tokens": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
    "expires_in": 900
  }
}
```

**Codes d'erreur :**
- `401 Unauthorized` : Identifiants invalides
- `422 Unprocessable Entity` : Données de validation invalides
```

### 3. Tester votre contribution
```bash
# Tests unitaires
go test ./... -v
cargo test
npm test

# Tests d'intégration
go test ./... -tags=integration
cargo test --features integration

# Vérification du code
golangci-lint run
cargo clippy
npm run lint

# Build complet
./scripts/dev-build.sh
```

### 4. Commiter vos changements
```bash
# Voir les changements
git status
git diff

# Ajouter les fichiers
git add .

# Créer un commit descriptif
git commit -m "feat: ajouter test pour UserService.GetUserByEmail

- Ajoute un test unitaire pour la méthode GetUserByEmail
- Couvre le cas de succès avec un utilisateur valide
- Utilise le pattern Arrange-Act-Assert pour la clarté

Closes #123"
```

#### Convention de commit
```bash
# Format: type(scope): description

# Types
feat:      nouvelle fonctionnalité
fix:       correction de bug
docs:      documentation
style:     formatage, espaces, etc.
refactor:  refactoring
test:      ajout/modification de tests
chore:     tâches de maintenance

# Exemples
feat(auth): ajouter support OAuth2 Google
fix(api): corriger validation email
docs(api): améliorer documentation endpoints
test(user): ajouter tests unitaires UserService
```

### 5. Pousser et créer une Pull Request
```bash
# Pousser votre branche
git push origin feature/ma-premiere-contribution

# Aller sur GitHub et créer une Pull Request
# https://github.com/veza/veza-full-stack/compare
```

#### Template de Pull Request
```markdown
## Description

Brève description de votre contribution.

## Type de changement

- [ ] Bug fix (correction qui résout un problème)
- [ ] New feature (fonctionnalité qui ajoute quelque chose)
- [ ] Breaking change (correction ou fonctionnalité qui casse quelque chose)
- [ ] Documentation update

## Tests

- [ ] J'ai ajouté des tests qui prouvent que ma correction fonctionne
- [ ] J'ai ajouté des tests qui prouvent que ma nouvelle fonctionnalité fonctionne
- [ ] Tous les tests existants passent

## Checklist

- [ ] Mon code suit les conventions de style du projet
- [ ] J'ai effectué un auto-review de mon propre code
- [ ] J'ai commenté mon code, particulièrement dans les zones difficiles à comprendre
- [ ] J'ai fait les changements correspondants dans la documentation
- [ ] Mes changements ne génèrent pas de nouveaux warnings
- [ ] J'ai ajouté des tests qui prouvent que ma correction fonctionne ou que ma fonctionnalité fonctionne
- [ ] Les tests unitaires et d'intégration passent localement avec mes changements
- [ ] Tout nouveau code ou changement de logique a une couverture de tests appropriée

## Screenshots (si applicable)

Ajoutez des captures d'écran pour aider à expliquer votre changement.

## Informations supplémentaires

Toute information supplémentaire ou contexte que vous souhaitez ajouter.
```

## Code Review

### 1. Attendre la review
- Les maintainers examineront votre PR
- Ils peuvent demander des modifications
- Soyez patient et réactif

### 2. Répondre aux commentaires
```bash
# Si des modifications sont demandées
git add .
git commit -m "fix: adresser les commentaires de review

- Corriger la validation d'email
- Ajouter des tests supplémentaires
- Améliorer la documentation"
git push origin feature/ma-premiere-contribution
```

### 3. Merge de votre PR
- Une fois approuvée, votre PR sera mergée
- Votre contribution sera intégrée au projet
- Félicitations ! 🎉

## Bonnes pratiques

### 1. Communication
- **Soyez clair** : Expliquez ce que vous faites et pourquoi
- **Posez des questions** : N'hésitez pas à demander de l'aide
- **Soyez respectueux** : Respectez les autres contributeurs

### 2. Code
- **Suivez les conventions** : Respectez le style de code du projet
- **Écrivez des tests** : Testez votre code
- **Documentez** : Commentez votre code si nécessaire

### 3. Git
- **Commits atomiques** : Un commit = une fonctionnalité
- **Messages descriptifs** : Expliquez ce que fait le commit
- **Branches à jour** : Synchronisez régulièrement avec main

### 4. Tests
```bash
# Avant de soumettre votre PR
./scripts/dev-test.sh
./scripts/dev-build.sh

# Vérification de la qualité
golangci-lint run
cargo clippy
npm run lint
```

## Ressources d'aide

### Documentation
- [Guide de développement](./development-environment.md)
- [Standards de code](./code-review.md)
- [Architecture du projet](../architecture/backend-architecture.md)

### Communauté
- **Discussions GitHub** : [GitHub Discussions](https://github.com/veza/veza-full-stack/discussions)
- **Issues** : [GitHub Issues](https://github.com/veza/veza-full-stack/issues)
- **Documentation** : [Veza Docs](https://docs.veza.app)

### Outils utiles
- **GitHub Desktop** : Interface graphique pour Git
- **GitKraken** : Client Git avancé
- **Postman** : Test d'APIs
- **DBeaver** : Client de base de données

## Exemples de contributions

### 1. Correction de documentation
```markdown
# Issue : #456 - Documentation API incomplète

## Changements
- Ajouter des exemples de requêtes pour tous les endpoints
- Corriger les descriptions des codes d'erreur
- Ajouter des exemples de réponses

## Fichiers modifiés
- docs/api/endpoints/users-api.md
- docs/api/endpoints/tracks-api.md
- docs/api/endpoints/chat-api.md
```

### 2. Ajout de tests
```go
// Issue : #789 - Manque de tests pour UserService

func TestUserService_ValidateEmail(t *testing.T) {
    tests := []struct {
        name    string
        email   string
        wantErr bool
    }{
        {"valid email", "test@example.com", false},
        {"invalid email", "invalid-email", true},
        {"empty email", "", true},
        {"missing @", "testexample.com", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validateEmail(tt.email)
            if (err != nil) != tt.wantErr {
                t.Errorf("validateEmail() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### 3. Correction de bug
```go
// Issue : #123 - Validation email trop permissive

// Avant
func validateEmail(email string) error {
    if !strings.Contains(email, "@") {
        return errors.New("invalid email")
    }
    return nil
}

// Après
func validateEmail(email string) error {
    if email == "" {
        return errors.New("email is required")
    }
    
    if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
        return errors.New("invalid email format")
    }
    
    // Validation plus stricte avec regex
    emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
    if !emailRegex.MatchString(email) {
        return errors.New("invalid email format")
    }
    
    return nil
}
```

## Prochaines étapes

### Après votre première contribution
1. **Continuez à contribuer** : Trouvez d'autres issues
2. **Améliorez vos compétences** : Apprenez de nouvelles technologies
3. **Aidez les autres** : Répondez aux questions dans les discussions
4. **Proposez des améliorations** : Créez des issues pour des idées

### Évolution dans le projet
- **Contributeur régulier** : Plusieurs contributions
- **Maintainer** : Responsabilité sur certaines parties
- **Core maintainer** : Accès complet au projet

## Conclusion

Félicitations pour votre première contribution ! Chaque contribution, même petite, aide à améliorer le projet Veza. N'hésitez pas à continuer à contribuer et à faire partie de notre communauté.

### Ressources supplémentaires
- [Guide de développement](./development-environment.md)
- [Standards de code](./code-review.md)
- [Architecture du projet](../architecture/backend-architecture.md) 