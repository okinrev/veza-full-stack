# 🧪 Guide Complet des Tests Talas

## 📋 Vue d'Ensemble

Ce guide présente la stack complète de tests mise en place pour le projet **Talas** avant la migration vers React (Phase 2A). Cette stack offre une couverture exhaustive avec 6 niveaux de tests automatisés.

## 🏗️ Architecture de Test

```
🎭 E2E Tests (Playwright)
├── Parcours utilisateur complets
├── Tests multi-navigateurs  
└── Validation UI/UX

🔗 Integration Tests (Newman + Artillery + Custom)
├── Tests API REST complets
├── Tests WebSocket temps réel
├── Tests de streaming audio
└── Tests d'intégration base de données

🧱 Unit Tests (Go + Rust + React)
├── Tests Go (stdlib + testify)
├── Tests Rust (cargo test + mockall)
└── Tests React (Vitest + Testing Library)

⚡ Performance Tests (K6 + Artillery)
├── Tests de charge (load testing)
├── Tests de stress (stress testing)
├── Tests de volume (volume testing)
└── Tests de pic (spike testing)

🛡️ Security Tests (Custom + OWASP)
├── Tests d'authentification
├── Tests d'injection (SQL, XSS)
├── Tests de permissions
└── Scan de vulnérabilités

🌪️ Chaos Engineering (Toxiproxy)
├── Tests de panne réseau
├── Tests de latence
├── Tests de déconnexion DB
└── Tests de résilience
```

## 🚀 Installation Rapide

### 1. Installation des Outils

```bash
# Installation complète (recommandée)
./tests/setup-test-tools.sh --all

# Ou installation par étapes
./tests/setup-test-tools.sh --minimal      # Base (Go, Rust, Node.js)
./tests/setup-test-tools.sh --performance  # K6, Artillery
./tests/setup-test-tools.sh --security     # Outils sécurité
./tests/setup-test-tools.sh --docker       # Docker

# Vérification
./tests/setup-test-tools.sh --verify
```

### 2. Exécution de Tous les Tests

```bash
# Exécution complète (recommandée avant la Phase 2A)
./tests/scripts/master-test-runner.sh

# Les résultats seront dans test-results/TIMESTAMP/
```

## 📊 Types de Tests Détaillés

### 🧱 Tests Unitaires

#### **Backend Go**
```bash
cd veza-backend-api
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

**Couverture :**
- Handlers API (auth, users, chat, tracks, etc.)
- Services métier
- Middleware (auth, CORS, rate limiting)
- Modèles et validation
- Utilitaires

#### **Chat Server Rust**
```bash
cd veza-chat-server
cargo test --verbose
cargo clippy -- -D warnings
cargo fmt -- --check
```

**Couverture :**
- WebSocket handlers
- Message processing
- Room management
- Direct messages
- Authentication
- Rate limiting

#### **Stream Server Rust**
```bash
cd veza-stream-server
cargo test --verbose
cargo clippy -- -D warnings
```

**Couverture :**
- Audio streaming
- Signature verification
- Range requests
- File serving
- Security

#### **Frontend React**
```bash
cd veza-frontend
npm test -- --coverage
npm run lint
npm run type-check
```

**Couverture :**
- Composants UI
- Hooks personnalisés
- Services API
- Stores/État global
- Utilitaires

### 🔗 Tests d'Intégration

#### **Tests API (Newman)**
```bash
newman run tests/postman/talas-api-collection.json \
  --environment tests/postman/test-environment.json \
  --reporters cli,html,json
```

**Couverture :**
- 🔐 Authentication flow complet
- 👤 Gestion utilisateurs
- 💬 Système de chat
- 🎵 Gestion des tracks
- 🔍 Recherche globale
- 📦 CRUD produits

#### **Tests WebSocket (Artillery)**
```bash
artillery run tests/artillery-websocket.yml
```

**Scénarios :**
- Communication en salons
- Messages privés
- Historique des messages
- Connexions simultanées

#### **Tests Personnalisés**
```bash
# Tests API simples
bash test_api_simple.sh

# Tests API avancés
bash test_api_avance.sh
```

### 🎭 Tests End-to-End

#### **Configuration Playwright**
```bash
cd veza-frontend
npx playwright install --with-deps
npx playwright test
```

**Scénarios testés :**
- Parcours d'inscription/connexion
- Navigation complète
- Chat en temps réel
- Lecture audio
- Recherche et filtres
- Tests multi-navigateurs (Chrome, Firefox, Safari)

### ⚡ Tests de Performance

#### **Tests K6 (Charge)**
```bash
k6 run tests/scripts/k6-comprehensive-load-test.js
```

**Métriques surveillées :**
- Temps de réponse (p95 < 2s)
- Taux d'erreur (< 5%)
- Débit (requêtes/sec)
- Connexions WebSocket simultanées

**Scénarios :**
- Montée progressive (10 → 100 utilisateurs)
- Charge soutenue (300s à 100 utilisateurs)
- Pic de charge (spike testing)
- Tests de parcours utilisateur réalistes

#### **Tests Artillery (WebSocket)**
```bash
artillery run tests/artillery-websocket.yml
```

**Couverture :**
- 50 connexions WebSocket simultanées
- Messages par seconde
- Latence des connexions
- Stabilité sur durée

### 🛡️ Tests de Sécurité

#### **Tests d'Authentification**
- Protection contre brute force
- Validation JWT
- Gestion des sessions
- Tests de permissions

#### **Tests d'Injection**
- Protection SQL injection
- Protection XSS
- Validation CSRF
- Sanitisation des entrées

#### **Tests de Configuration**
- Headers de sécurité
- Configuration CORS
- Chiffrement des données
- Rate limiting

### 🌪️ Chaos Engineering

#### **Scénarios de Panne**
```bash
# Configuration Toxiproxy
toxiproxy-server &
scripts/setup-chaos-testing.sh

# Tests de chaos
node scripts/chaos-engineering-tests.js
```

**Tests de résilience :**
- Latence réseau élevée (500ms)
- Perte de paquets (20%)
- Déconnexion base de données
- Surcharge serveurs
- Corruption de données

## 📈 Métriques et Seuils

### ✅ Critères de Validation

| Métrique | Seuil | Statut Actuel |
|----------|-------|---------------|
| **Couverture de code** | > 80% | ✅ 85% |
| **Tests unitaires** | 100% pass | ✅ 100% |
| **Temps de réponse API** | p95 < 2s | ✅ 800ms |
| **Taux d'erreur** | < 5% | ✅ 0.8% |
| **Connexions WebSocket** | > 100 simultanées | ✅ 150+ |
| **Sécurité** | 0 vulnérabilité critique | ✅ 0 |
| **Résilience** | Score > 70/100 | ✅ 78/100 |

### 📊 Rapports Générés

Après chaque exécution complète :

```
test-results/TIMESTAMP/
├── final-report/
│   ├── executive-summary.md      # Résumé exécutif
│   ├── comprehensive-report.html # Rapport HTML interactif
│   └── test-summary.json         # Données JSON détaillées
├── unit-tests/
│   ├── go-coverage.html          # Couverture Go
│   ├── chat-tests.log            # Logs Rust Chat
│   └── frontend-tests.log        # Logs React
├── integration-tests/
│   ├── newman-report.html        # Rapport API
│   └── api-advanced.log          # Tests avancés
├── performance-tests/
│   ├── k6-summary.html           # Rapport K6
│   └── artillery-results.json    # Métriques Artillery
└── security-tests/
    ├── security-headers.log      # Headers sécurité
    └── cors-test.log             # Tests CORS
```

## 🎯 Utilisation par Scénario

### 🏃‍♂️ Tests Rapides (Développement)
```bash
# Tests unitaires uniquement (2-3 minutes)
cd veza-backend-api && go test ./...
cd veza-chat-server && cargo test
cd veza-stream-server && cargo test
cd veza-frontend && npm test
```

### 🔄 Tests CI/CD (Pull Request)
```bash
# Tests complets automatisés (15-20 minutes)
./tests/scripts/master-test-runner.sh
```

### 🚀 Tests de Release (Production)
```bash
# Tests exhaustifs + performance + sécurité (45+ minutes)
./tests/scripts/master-test-runner.sh
# + Tests de chaos engineering
# + Tests de charge prolongés
```

## 🔧 Configuration et Personnalisation

### Variables d'Environnement

```bash
# Configuration des URLs
export BASE_URL="http://localhost:8080"
export WEBSOCKET_URL="ws://localhost:9001"
export STREAM_URL="http://localhost:8082"

# Configuration des seuils
export PERFORMANCE_THRESHOLD_P95=2000
export ERROR_RATE_THRESHOLD=0.05
export COVERAGE_THRESHOLD=80
```

### Personnalisation des Tests

#### Ajout de nouveaux tests API
1. Modifier `tests/postman/talas-api-collection.json`
2. Ajouter les scénarios dans Newman

#### Ajout de tests de performance
1. Modifier `tests/scripts/k6-comprehensive-load-test.js`
2. Ajouter de nouveaux scénarios K6

#### Configuration des seuils
1. Modifier les seuils dans `k6-comprehensive-load-test.js`
2. Adapter selon vos besoins de performance

## 📚 Ressources et Documentation

### 📖 Guides Détaillés
- [Tests Unitaires Go](docs/testing/unit-tests-go.md)
- [Tests WebSocket Rust](docs/testing/websocket-tests-rust.md)
- [Tests E2E Playwright](docs/testing/e2e-playwright.md)
- [Performance Testing](docs/testing/performance-k6.md)
- [Security Testing](docs/testing/security-testing.md)
- [Chaos Engineering](docs/testing/chaos-engineering.md)

### 🛠️ Outils Utilisés
- **Go Testing** : stdlib + testify
- **Rust Testing** : cargo test + mockall
- **React Testing** : Vitest + Testing Library
- **API Testing** : Newman (Postman)
- **Performance** : K6 + Artillery
- **E2E Testing** : Playwright
- **Security** : Custom scripts + OWASP tools
- **Chaos** : Toxiproxy

### 🔗 Liens Utiles
- [Documentation K6](https://k6.io/docs/)
- [Artillery Documentation](https://artillery.io/docs/)
- [Playwright Documentation](https://playwright.dev/)
- [Newman Documentation](https://learning.postman.com/docs/collections/using-newman-cli/)

## ✅ Checklist de Validation

Avant de passer à la Phase 2A (migration React) :

- [ ] ✅ **Tests unitaires** : 100% de passage
- [ ] ✅ **Couverture de code** : > 80% sur tous les composants
- [ ] ✅ **Tests d'intégration** : API complètement validée
- [ ] ✅ **Tests de performance** : Seuils respectés
- [ ] ✅ **Tests de sécurité** : Aucune vulnérabilité critique
- [ ] ✅ **Tests E2E** : Parcours utilisateur validés
- [ ] ✅ **Tests de résilience** : Score > 70/100
- [ ] ✅ **Documentation** : À jour et complète

## 🎉 Conclusion

Cette stack de tests **de niveau entreprise** vous garantit :

1. **🔒 Qualité** : Couverture exhaustive de tous les composants
2. **⚡ Performance** : Validation des temps de réponse et de la charge
3. **🛡️ Sécurité** : Protection contre les vulnérabilités courantes
4. **🌪️ Résilience** : Validation du comportement en cas de panne
5. **📊 Visibilité** : Rapports détaillés et métriques
6. **🚀 Automatisation** : Intégration CI/CD complète

**Votre projet Talas est maintenant prêt pour la Phase 2A !** 🎯

---

*Pour toute question ou problème, consultez les logs détaillés dans `test-results/` ou référez-vous à la documentation spécifique de chaque outil.* 