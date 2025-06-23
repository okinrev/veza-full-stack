# 🎵 Veza - Application Web Unifiée

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org/)
[![React Version](https://img.shields.io/badge/React-18.2+-blue.svg)](https://reactjs.org/)
[![Rust Version](https://img.shields.io/badge/Rust-1.70+-orange.svg)](https://www.rust-lang.org/)

Une application web moderne et performante combinant chat en temps réel, streaming audio, et une interface utilisateur réactive. Construite avec une architecture microservices utilisant Go, React, et Rust.

## 🚀 Démarrage Rapide

### Prérequis

- **Docker** et **Docker Compose** (recommandé)
- Ou :
  - **Go** 1.21+
  - **Node.js** 18+ et **npm**
  - **Rust** 1.70+
  - **PostgreSQL** 15+
  - **Redis** 7+

### Installation Express avec Docker

```bash
# Cloner le projet
git clone <repository-url>
cd veza-full-stack

# Configuration initiale
make setup

# Démarrer tous les services
make docker-up

# Accéder à l'application
open http://localhost
```

### Installation Développement Local

```bash
# Configuration initiale
make setup

# Démarrer en mode développement (tous les services)
make dev
```

L'application sera disponible sur :
- 🎨 **Frontend** : http://localhost:5173 (dev) / http://localhost (prod)
- ⚙️ **Backend API** : http://localhost:8080
- 💬 **Chat WebSocket** : ws://localhost:8081/ws
- 🎵 **Stream WebSocket** : ws://localhost:8082/ws
- ⚖️ **HAProxy Stats** : http://localhost:8404/stats

## 🏗️ Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Frontend      │    │   Backend API    │    │   Modules Rust  │
│   React + TS    │◄──►│   Go + Gin       │◄──►│   Chat + Stream │
│   Port 5173     │    │   Port 8080      │    │   Ports 8081/82 │
└─────────────────┘    └──────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
         └───────────────────────▼───────────────────────┘
                    ┌─────────────────┐    ┌─────────────────┐
                    │   PostgreSQL    │    │     Redis       │
                    │   Port 5432     │    │   Port 6379     │
                    └─────────────────┘    └─────────────────┘
```

### 🎯 Services

#### 🎨 Frontend React
- **Framework** : React 18 + TypeScript + Vite
- **UI** : Tailwind CSS + shadcn/ui
- **État** : Zustand + React Query
- **Fonctionnalités** : Chat temps réel, lecteur audio, interface responsive

#### ⚙️ Backend API Go
- **Framework** : Gin + GORM
- **Base de données** : PostgreSQL
- **Authentification** : JWT avec refresh tokens
- **APIs** : REST endpoints + WebSocket proxy

#### 💬 Chat Server Rust
- **Framework** : Tokio + Axum
- **WebSocket** : Messages temps réel, salles, DM
- **Fonctionnalités** : Modération, réactions, mentions, threads

#### 🎵 Stream Server Rust
- **Framework** : Tokio + Axum
- **Streaming** : Audio adaptatif, range requests
- **Fonctionnalités** : Upload, transcoding, analytics

## 📚 Documentation

### 📖 Documentation Complète
- [**Backend API Go**](./veza-backend-api/docs/README.md) - Configuration, endpoints, authentification
- [**Frontend React**](./veza-frontend/docs/README.md) - Architecture, composants, intégration
- [**Chat Server Rust**](./veza-chat-server/docs/README.md) - WebSocket API, protocoles, sécurité
- [**Stream Server Rust**](./veza-stream-server/docs/README.md) - Streaming, compression, analytics

### 🔧 Guides d'Intégration
- [Intégration Go ↔ Rust](./veza-chat-server/docs/integration_go.md)
- [Intégration React ↔ WebSocket](./veza-chat-server/docs/integration_react.md)
- [Protocoles WebSocket](./veza-backend-api/docs/websocket-protocol.md)
- [Architecture Frontend](./veza-frontend/docs/architecture/overview.md)

## 🛠️ Commandes Make

```bash
# Développement
make dev              # Démarrer tous les services en dev
make dev-frontend     # Frontend uniquement
make dev-backend      # Backend uniquement

# Production
make build            # Construire tous les services
make docker-up        # Démarrer avec Docker
make docker-down      # Arrêter Docker

# Tests
make test             # Tous les tests
make test-frontend    # Tests React
make test-backend     # Tests Go
make test-e2e         # Tests end-to-end

# Utilitaires
make clean            # Nettoyer les builds
make logs             # Voir les logs
make health           # Vérifier la santé des services
make format           # Formater le code

make help             # Aide complète
```

## 🔧 Configuration

### Variables d'Environnement

Copiez `env.example` vers `.env` et configurez :

```bash
# Base de données
DATABASE_URL=postgres://veza_user:veza_password@localhost:5432/veza_db
REDIS_URL=redis://localhost:6379

# JWT
JWT_SECRET=your-super-secret-key-change-in-production

# Services
API_PORT=8080
CHAT_PORT=8081
STREAM_PORT=8082

# Frontend
VITE_API_URL=http://localhost:8080/api/v1
VITE_WS_CHAT_URL=ws://localhost:8081/ws
VITE_WS_STREAM_URL=ws://localhost:8082/ws
```

### 🐳 Docker Compose

Services inclus :
- **PostgreSQL** avec migration automatique
- **Redis** pour le cache et sessions
- **HAProxy** comme load balancer avancé
- Tous les services applicatifs avec haute disponibilité

## 🎯 Fonctionnalités

### 💬 Chat Temps Réel
- Salles publiques et privées
- Messages directs
- Réactions et mentions
- Modération avancée
- Threads de discussion

### 🎵 Streaming Audio
- Upload de fichiers audio
- Streaming adaptatif
- Playlists collaboratives
- Analytics d'écoute
- Génération de waveforms

### 👤 Gestion Utilisateur
- Authentification JWT sécurisée
- Profils utilisateur
- Système de rôles
- Sessions multiples

### 🛡️ Sécurité
- Authentification multi-niveaux
- Rate limiting intelligent
- Validation stricte des entrées
- Headers de sécurité
- Audit complet

## 🚀 Déploiement

### Développement Local

```bash
# Avec Docker
make docker-up

# Manuel
make dev
```

### Production

```bash
# Build pour production
make build

# Déploiement Docker
docker-compose -f docker-compose.prod.yml up -d

# Avec reverse proxy
make docker-up  # Nginx inclus sur port 80
```

### Variables de Production

```bash
ENVIRONMENT=production
DEBUG=false
DATABASE_URL=postgres://prod_user:strong_password@prod-db:5432/veza_prod
JWT_SECRET=super-secure-production-secret-minimum-32-characters
ALLOWED_ORIGINS=https://yourdomain.com
```

### Migration Base de Données Existante

Si vous avez déjà une base de données Veza existante :

```bash
# 1. Importer votre dump existant
docker-compose exec postgres psql -U veza_user -d veza_db < veza_db_dump_21_06_2025.sql

# 2. Exécuter la migration pour ajouter les nouvelles fonctionnalités
make db-migrate-existing

# 3. Vérifier la migration
docker-compose exec postgres psql -U veza_user -d veza_db -c "\dt"
```

### HAProxy Load Balancing

Configuration avancée avec :
- **Round Robin** pour l'API backend
- **Source IP** pour les WebSockets (sticky sessions)
- **Least Connections** pour le streaming
- **Health Checks** automatiques
- **Failover** avec serveurs de backup
- **Rate Limiting** et protection DDoS
- **Interface d'administration** : http://localhost:8404/stats

## 🧪 Tests

### Tests Unitaires
```bash
make test-frontend    # React + Vitest
make test-backend     # Go tests
make test-chat        # Rust tests
make test-stream      # Rust tests
```

### Tests d'Intégration
```bash
make test-e2e         # Playwright E2E
```

### Tests de Performance
```bash
# Load testing des WebSockets
cd scripts && ./load-test-websocket.sh

# Load testing de l'API
cd scripts && ./load-test-api.sh
```

## 📊 Monitoring

### Health Checks
```bash
make health           # Vérifier tous les services

# Endpoints individuels
curl http://localhost:8080/health    # Backend
curl http://localhost:8081/health    # Chat
curl http://localhost:8082/health    # Stream
```

### Métriques
```bash
# Métriques Prometheus
curl http://localhost:8080/metrics

# Logs structurés
make logs
```

## 🤝 Contribution

### Structure du Code

```
veza-full-stack/
├── veza-backend-api/     # 🟢 Backend Go
├── veza-frontend/        # 🔵 Frontend React
├── veza-chat-server/     # 🟠 Chat Rust
├── veza-stream-server/   # 🟠 Stream Rust
├── docker-compose.yml    # 🐳 Orchestration
├── nginx.conf           # 🌐 Reverse proxy
├── init-db.sql          # 🗄️ Schema BDD
└── Makefile            # 🔧 Commandes
```

### Workflow de Développement

1. **Fork** le projet
2. **Créer** une branche feature
3. **Développer** avec les commandes make
4. **Tester** avec `make test`
5. **Formater** avec `make format`
6. **Commit** et push
7. **Pull Request** avec description

### Standards de Code

- **Go** : `go fmt`, `golint`, tests obligatoires
- **TypeScript** : ESLint, Prettier, tests Jest/Vitest
- **Rust** : `cargo fmt`, `cargo clippy`, tests intégrés

## 📝 Changelog

### v1.0.0 (En développement)
- ✅ Architecture microservices complète
- ✅ Chat temps réel WebSocket
- ✅ Streaming audio adaptatif
- ✅ Interface React moderne
- ✅ Authentification JWT sécurisée
- ✅ Docker Compose complet
- 🔄 Tests E2E
- 🔄 Documentation API complète

## 📜 Licence

MIT License - voir [LICENSE](LICENSE) pour les détails.

## 🙏 Remerciements

- [Gin](https://gin-gonic.com/) - Framework web Go
- [React](https://reactjs.org/) - Bibliothèque UI
- [Tokio](https://tokio.rs/) - Runtime async Rust
- [Tailwind CSS](https://tailwindcss.com/) - Framework CSS
- [shadcn/ui](https://ui.shadcn.com/) - Composants UI

---

**⭐ N'hésitez pas à laisser une étoile si ce projet vous aide !**

# 🚀 Veza Full-Stack - Infrastructure Déployée

## 📋 État Actuel

✅ **Infrastructure complètement déployée et opérationnelle !**

### 🏗️ Architecture Microservices Incus

L'infrastructure Veza est maintenant déployée avec 8 containers Incus interconnectés :

| Service | Container | IP | Port | Runtime | Status |
|---------|-----------|----|----- |---------|--------|
| 🖥️ **Frontend** | `veza-frontend` | `10.5.191.41` | `5173` | Node.js 20.19.2 | ✅ Ready |
| 🔧 **Backend API** | `veza-backend` | `10.5.191.241` | `8080` | Go 1.21.5 | ✅ Ready |
| 💬 **Chat Server** | `veza-chat` | `10.5.191.49` | `8081` | Rust 1.87.0 | ✅ Ready |
| 🎵 **Stream Server** | `veza-stream` | `10.5.191.196` | `8082` | Rust 1.87.0 | ✅ Ready |
| 🗄️ **PostgreSQL** | `veza-postgres` | `10.5.191.134` | `5432` | PostgreSQL 15 | ✅ Active |
| 🔴 **Redis** | `veza-redis` | `10.5.191.186` | `6379` | Redis 7.0.15 | ✅ Active |
| ⚖️ **HAProxy** | `veza-haproxy` | `10.5.191.133` | `80/8404` | HAProxy 2.6.12 | ✅ Active |
| 📁 **Storage** | `veza-storage` | `10.5.191.206` | `2049` | NFS | ✅ Ready |

## 🌐 Points d'Accès

- **🌍 Application** : http://10.5.191.133
- **📊 HAProxy Stats** : http://10.5.191.133:8404/stats
- **⚛️ Frontend Dev** : http://10.5.191.41:5173
- **🔧 Backend API** : http://10.5.191.241:8080

## 🛠️ Gestion de l'Infrastructure

### Script de Gestion Intégré

```bash
# Voir l'état complet
./scripts/veza-manage.sh status

# Tester la connectivité
./scripts/veza-manage.sh test

# Voir les logs
./scripts/veza-manage.sh logs haproxy
./scripts/veza-manage.sh logs postgres

# Entrer dans un container
./scripts/veza-manage.sh shell veza-backend

# Redémarrer les services
./scripts/veza-manage.sh restart

# Sauvegarder la base de données
./scripts/veza-manage.sh backup

# Ouvrir l'interface de monitoring
./scripts/veza-manage.sh monitor
```

## 📁 Structure du Projet

```
veza-full-stack/
├── 📁 configs/                    # Configuration centralisée
│   ├── haproxy.cfg                # Configuration HAProxy
│   ├── infrastructure.yaml        # Config infrastructure complète
│   ├── env.example               # Variables d'environnement
│   ├── init-db.sql              # Script d'initialisation DB
│   ├── migrate-existing-db.sql   # Migration DB existante
│   └── incus-config.yaml        # Configuration Incus
├── 📁 scripts/                   # Scripts d'automatisation
│   ├── veza-manage.sh           # 🎛️ Gestionnaire principal
│   ├── incus-deploy-final.sh    # Déploiement infrastructure
│   └── incus-clean.sh           # Nettoyage complet
├── 📁 veza-frontend/            # Code source React
├── 📁 veza-backend-api/         # Code source Go API
├── 📁 veza-chat-server/         # Code source Chat Rust
├── 📁 veza-stream-server/       # Code source Stream Rust
└── 📁 backups/                  # Sauvegardes automatiques
```

## 🔄 Prochaines Étapes

### 1. **Déploiement du Code Source** 📦
- [ ] Copier le code dans les containers
- [ ] Configurer les variables d'environnement
- [ ] Construire les applications

### 2. **Configuration de la Base de Données** 🗄️
- [ ] Initialiser le schéma avec le dump existant
- [ ] Configurer les connexions
- [ ] Tester les migrations

### 3. **Démarrage des Services** 🚀
- [ ] Démarrer le backend Go (port 8080)
- [ ] Démarrer le frontend React (port 5173)
- [ ] Démarrer le chat server Rust (port 8081)
- [ ] Démarrer le stream server Rust (port 8082)

### 4. **Tests d'Intégration** 🧪
- [ ] Tester l'API backend
- [ ] Tester les WebSockets
- [ ] Tester le routage HAProxy
- [ ] Tester l'application complète

## 🔧 Commandes Utiles

### Gestion des Containers

```bash
# Lister tous les containers
incus list

# Entrer dans un container
incus exec veza-frontend -- bash

# Voir les logs d'un service
incus exec veza-haproxy -- journalctl -u haproxy -f

# Redémarrer un container
incus restart veza-backend
```

### Monitoring

```bash
# Statut des services
systemctl status postgresql  # Dans veza-postgres
systemctl status redis-server  # Dans veza-redis
systemctl status haproxy     # Dans veza-haproxy

# Test de connectivité
ping 10.5.191.133  # HAProxy
curl http://10.5.191.133:8404/stats  # Stats HAProxy
```

## 🏗️ Architecture Technique

### Réseau
- **Type** : Réseau par défaut Incus (incusbr0)
- **Plage IP** : 10.5.191.0/24
- **DHCP** : Attribution automatique des IPs

### Load Balancing (HAProxy)
- **Algorithme API** : Round Robin
- **Algorithme Chat** : Source IP (sticky sessions)
- **Algorithme Stream** : Least Connections
- **Health Checks** : Automatiques sur tous les backends

### Sécurité
- **Isolation** : Containers Incus séparés
- **Réseau** : Réseau privé isolé
- **Rate Limiting** : Protection DDoS via HAProxy
- **Headers** : Sécurité HTTP configurée

## 🚨 Résolution de Problèmes

### Container ne démarre pas
```bash
incus info veza-[nom]
incus exec veza-[nom] -- systemctl status
```

### Service inaccessible
```bash
./scripts/veza-manage.sh test
incus exec veza-haproxy -- systemctl status haproxy
```

### Problème de connectivité
```bash
ping 10.5.191.133
curl -I http://10.5.191.133:8404/stats
```

## 📊 Monitoring et Logs

- **HAProxy Stats** : Interface web complète sur le port 8404
- **Logs centralisés** : Via `journalctl` dans chaque container
- **Health checks** : Automatiques avec alertes
- **Métriques** : Accessibles via l'interface HAProxy

---

## 🎯 Objectif Atteint

✅ **Infrastructure microservices complètement opérationnelle !**

L'architecture Veza est maintenant prête pour recevoir le code source et devenir une application web complète. Tous les services de base (base de données, cache, load balancer) sont configurés et fonctionnels.

**Prêt pour la phase de déploiement applicatif ! 🚀** 