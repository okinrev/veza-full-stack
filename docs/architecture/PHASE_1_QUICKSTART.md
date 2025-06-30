# 🚀 Phase 1 - Démarrage Rapide Architecture Hexagonale

## 📋 Objectif Phase 1

Transformer le backend Go en architecture hexagonale avec :
- ✅ Structure des dossiers hexagonale 
- ✅ Injection de dépendances
- ✅ Repository pattern
- ✅ Services métier isolés
- ✅ Cache Redis intégré
- ✅ Configuration avancée

## 🏗️ Structure Créée

```
veza-backend-api/
├── internal/
│   ├── domain/                  # 🆕 Couche métier
│   │   ├── entities/           # Entités métier
│   │   │   └── user.go        # ✅ Entité User complète
│   │   ├── repositories/      # Interfaces persistence  
│   │   │   └── user_repository.go # ✅ Interface UserRepository
│   │   └── services/          # Services métier
│   │       ├── auth_service.go    # ✅ Service authentification
│   │       └── cache_service.go   # ✅ Interface cache
│   ├── ports/                  # 🆕 Interfaces externes
│   │   └── http/              # Handlers HTTP
│   │       └── auth_handler.go    # ✅ Handler auth hexagonal
│   ├── adapters/              # 🆕 Implémentations infrastructure
│   │   ├── postgres/          # Repository PostgreSQL
│   │   ├── redis_cache/       # Implémentation cache Redis
│   │   └── grpc/             # Clients gRPC
│   └── infrastructure/        # 🆕 Configuration & DI
│       ├── config/           
│       │   └── app_config.go     # ✅ Configuration complète
│       └── container/
│           └── container.go       # ✅ Injection dépendances
```

## 🔧 Prochaines Étapes Immédiates

### 1. Installation des Dépendances

```bash
cd veza-backend-api
go mod tidy
go mod download
```

### 2. Configuration Environment

```bash
# Copier le template de configuration
cp .env.example .env

# Éditer avec vos paramètres
nano .env
```

**Contenu .env minimum :**
```env
# Base de données
DATABASE_URL=postgres://postgres:password@localhost:5432/veza_db?sslmode=disable

# Redis
REDIS_ENABLED=true
REDIS_HOST=localhost
REDIS_PORT=6379

# JWT
JWT_ACCESS_SECRET=your-super-secret-access-key-change-in-production
JWT_REFRESH_SECRET=your-super-secret-refresh-key-change-in-production

# Serveur
PORT=8080
ENVIRONMENT=development

# Monitoring
PROMETHEUS_ENABLED=true
LOG_LEVEL=info
```

### 3. Créer les Adapters Manquants

**a) Adapter PostgreSQL :**
```bash
mkdir -p internal/adapters/postgres
```

**b) Adapter Redis :**
```bash
mkdir -p internal/adapters/redis_cache
```

**c) Clients gRPC :**
```bash 
mkdir -p internal/adapters/grpc
```

### 4. Services à Implémenter

- [ ] `internal/adapters/postgres/user_repository.go`
- [ ] `internal/adapters/redis_cache/cache_service.go`
- [ ] `internal/adapters/grpc/chat_client.go`
- [ ] `internal/adapters/grpc/stream_client.go`

### 5. Tests à Créer

```bash
mkdir -p internal/domain/entities_test
mkdir -p internal/domain/services_test
mkdir -p internal/ports/http_test
```

## 🧪 Plan de Tests Phase 1

### Tests Unitaires
- [x] Entité User (validation, règles métier)
- [ ] AuthService (logique authentification)
- [ ] AuthHandler (endpoints HTTP)
- [ ] UserRepository (interface)

### Tests d'Intégration
- [ ] Base de données PostgreSQL
- [ ] Cache Redis  
- [ ] Endpoints API complets

### Tests de Performance
- [ ] Connexions base de données
- [ ] Performance cache Redis
- [ ] Latence API < 50ms

## 📊 Métriques de Succès Phase 1

| Critère | Objectif | Validation |
|---------|----------|------------|
| Architecture | Hexagonale complète | ✅ Structure créée |
| Tests Coverage | > 80% | 🔄 En cours |
| Performance API | < 50ms latence | 🔄 À mesurer |
| Cache Redis | Opérationnel | 🔄 À tester |
| Documentation | Complète | ✅ En cours |

## 🚦 État Actuel

### ✅ Terminé
- Structure dossiers hexagonale
- Configuration avancée
- Entité User complète avec validation
- Interface UserRepository
- Service AuthService (structure)
- Handler AuthHandler 
- Container injection dépendances

### 🔄 En Cours
- Implémentation adapters
- Tests unitaires
- Validation configuration

### ⏳ À Faire
- Adapter PostgreSQL
- Adapter Redis
- Clients gRPC
- Tests d'intégration
- Métriques Prometheus

## 🔥 Actions Prioritaires

### 1. Compléter les Adapters (2h)
```bash
# Créer l'adapter PostgreSQL
touch internal/adapters/postgres/user_repository.go
touch internal/adapters/postgres/connection.go

# Créer l'adapter Redis
touch internal/adapters/redis_cache/cache_service.go
touch internal/adapters/redis_cache/client.go
```

### 2. Tests de Base (1h)
```bash
# Tests entité User
touch internal/domain/entities/user_test.go

# Tests service Auth  
touch internal/domain/services/auth_service_test.go
```

### 3. Main Temporaire (30min)
```bash
# Point d'entrée simple pour tester
touch cmd/server/main_hexagonal.go
```

## 🎯 Validation Phase 1

**Critères de validation :**
- [ ] Architecture hexagonale fonctionnelle
- [ ] Injection de dépendances opérationnelle  
- [ ] Tests unitaires > 80% coverage
- [ ] API auth fonctionnelle
- [ ] Cache Redis connecté
- [ ] Performance < 50ms par requête

**Tests de validation :**
```bash
# Lancer les tests
go test ./internal/domain/... -v -cover

# Vérifier la compilation
go build ./cmd/server/

# Tester les endpoints
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","email":"test@example.com","password":"Test123!"}'
```

---

**🎉 Prêt pour Phase 2 ?**
Une fois Phase 1 validée, nous passerons à la sécurité avancée et aux middlewares !

**📞 Support :** 
Pour toute question sur la Phase 1, demandez assistance avec le détail spécifique. 