# Makefile pour Veza - Application Web Unifiée
# Commandes pour développement, build et déploiement

.PHONY: help dev build start stop clean logs test docker-build docker-up docker-down setup

# Configuration par défaut
ENV ?= development
DOCKER_COMPOSE_FILE ?= docker-compose.yml

# Couleurs pour les messages
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
NC := \033[0m # No Color

# Aide par défaut
help: ## Afficher l'aide
	@echo "$(BLUE)Veza - Application Web Unifiée$(NC)"
	@echo "=================================="
	@echo ""
	@echo "$(GREEN)Commandes disponibles:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(YELLOW)%-20s$(NC) %s\n", $$1, $$2}'
	@echo ""

# =============================================================================
# Setup et Installation
# =============================================================================

setup: ## Initialiser le projet (première fois)
	@echo "$(BLUE)🔧 Configuration initiale du projet Veza$(NC)"
	@echo "$(YELLOW)Création des répertoires nécessaires...$(NC)"
	@mkdir -p logs uploads audio ssl
	@echo "$(YELLOW)Copie du fichier de configuration...$(NC)"
	@cp env.example .env
	@echo "$(YELLOW)Installation des dépendances Frontend...$(NC)"
	@cd veza-frontend && npm install
	@echo "$(GREEN)✅ Projet configuré avec succès!$(NC)"
	@echo "$(BLUE)📝 N'oubliez pas d'éditer le fichier .env avec vos paramètres$(NC)"

install-deps: ## Installer toutes les dépendances
	@echo "$(BLUE)📦 Installation des dépendances$(NC)"
	@echo "$(YELLOW)Frontend (React)...$(NC)"
	@cd veza-frontend && npm install
	@echo "$(YELLOW)Backend (Go)...$(NC)"
	@cd veza-backend-api && go mod tidy
	@echo "$(YELLOW)Chat Server (Rust)...$(NC)"
	@cd veza-chat-server && cargo build --release
	@echo "$(YELLOW)Stream Server (Rust)...$(NC)"
	@cd veza-stream-server && cargo build --release
	@echo "$(GREEN)✅ Toutes les dépendances installées$(NC)"

# =============================================================================
# Développement Local
# =============================================================================

dev: incus-dev ## Démarrer l'environnement de développement (Incus par défaut)

dev-local: ## Démarrer tous les services en mode développement local (sans containers)
	@echo "$(BLUE)🚀 Démarrage en mode développement local$(NC)"
	@echo "$(YELLOW)Services qui vont démarrer:$(NC)"
	@echo "  - Frontend React: http://localhost:5173"
	@echo "  - Backend API Go: http://localhost:8080"
	@echo "  - Chat Server Rust: http://localhost:8081"
	@echo "  - Stream Server Rust: http://localhost:8082"
	@echo ""
	@echo "$(BLUE)💡 Utilisez Ctrl+C pour arrêter tous les services$(NC)"
	@echo ""
	@make -j4 dev-frontend dev-backend dev-chat dev-stream

dev-frontend: ## Démarrer le frontend React uniquement
	@echo "$(BLUE)🎨 Démarrage du Frontend React$(NC)"
	@cd veza-frontend && npm run dev

dev-backend: ## Démarrer le backend Go uniquement
	@echo "$(BLUE)⚙️ Démarrage du Backend Go$(NC)"
	@cd veza-backend-api && go run cmd/server/main.go

dev-chat: ## Démarrer le chat server Rust uniquement
	@echo "$(BLUE)💬 Démarrage du Chat Server$(NC)"
	@cd veza-chat-server && cargo run

dev-stream: ## Démarrer le stream server Rust uniquement
	@echo "$(BLUE)🎵 Démarrage du Stream Server$(NC)"
	@cd veza-stream-server && cargo run

# =============================================================================
# Build et Production
# =============================================================================

build: ## Construire tous les services pour la production
	@echo "$(BLUE)🔨 Construction pour la production$(NC)"
	@make build-frontend
	@make build-backend
	@make build-chat
	@make build-stream
	@echo "$(GREEN)✅ Tous les services construits$(NC)"

build-frontend: ## Construire le frontend React
	@echo "$(YELLOW)🎨 Construction du Frontend...$(NC)"
	@cd veza-frontend && npm run build

build-backend: ## Construire le backend Go
	@echo "$(YELLOW)⚙️ Construction du Backend...$(NC)"
	@cd veza-backend-api && go build -o bin/server cmd/server/main.go

build-chat: ## Construire le chat server Rust
	@echo "$(YELLOW)💬 Construction du Chat Server...$(NC)"
	@cd veza-chat-server && cargo build --release

build-stream: ## Construire le stream server Rust
	@echo "$(YELLOW)🎵 Construction du Stream Server...$(NC)"
	@cd veza-stream-server && cargo build --release

# =============================================================================
# Déploiement Unifié avec Incus
# =============================================================================

deploy: ## Déploiement complet avec le système unifié
	@echo "$(BLUE)🚀 Déploiement complet Veza$(NC)"
	@chmod +x scripts/deploy.sh
	@./scripts/deploy.sh deploy

deploy-dev: ## Déploiement en mode développement
	@echo "$(BLUE)🚀 Déploiement développement Veza$(NC)"
	@chmod +x scripts/deploy.sh
	@./scripts/deploy.sh deploy --dev

deploy-apps: ## Déployer seulement les applications
	@echo "$(BLUE)🚀 Déploiement applications uniquement$(NC)"
	@chmod +x scripts/deploy.sh
	@./scripts/deploy.sh apps

deploy-infrastructure: ## Déployer seulement l'infrastructure
	@echo "$(BLUE)🚀 Déploiement infrastructure uniquement$(NC)"
	@chmod +x scripts/deploy.sh
	@./scripts/deploy.sh infrastructure

deploy-rebuild: ## Reconstruction complète avec nettoyage
	@echo "$(BLUE)🚀 Reconstruction complète$(NC)"
	@chmod +x scripts/deploy.sh
	@./scripts/deploy.sh rebuild

deploy-test: ## Tester le déploiement
	@echo "$(BLUE)🧪 Tests du déploiement$(NC)"
	@chmod +x scripts/test.sh
	@./scripts/test.sh

deploy-test-quick: ## Tests rapides du déploiement
	@echo "$(BLUE)🧪 Tests rapides$(NC)"
	@chmod +x scripts/test.sh
	@./scripts/test.sh --quick

# Commandes legacy (redirigées vers le nouveau système)
incus-setup: ## Configuration initiale d'Incus (réseau et profils)
	@echo "$(YELLOW)⚠️ Commande legacy - utilisez 'make deploy' à la place$(NC)"
	@chmod +x scripts/deploy.sh  
	@./scripts/deploy.sh setup

incus-deploy: ## Déployer tous les containers Incus
	@echo "$(YELLOW)⚠️ Commande legacy - utilisez 'make deploy' à la place$(NC)"
	@chmod +x scripts/deploy.sh
	@./scripts/deploy.sh deploy

incus-status: ## Vérifier le statut des containers Incus
	@chmod +x scripts/incus-status.sh
	@./scripts/incus-status.sh

incus-logs: ## Afficher les logs des containers (usage: make incus-logs CONTAINER=veza-backend)
	@chmod +x scripts/incus-logs.sh
	@./scripts/incus-logs.sh $(CONTAINER)

incus-restart: ## Redémarrer un container ou tous (usage: make incus-restart CONTAINER=veza-backend)
	@chmod +x scripts/incus-restart.sh
	@./scripts/incus-restart.sh $(CONTAINER) restart

incus-stop: ## Arrêter tous les containers Incus
	@chmod +x scripts/incus-restart.sh
	@./scripts/incus-restart.sh all stop

incus-start: ## Démarrer tous les containers Incus
	@chmod +x scripts/incus-restart.sh
	@./scripts/incus-restart.sh all start

incus-clean: ## Nettoyer complètement l'environnement Incus
	@echo "$(YELLOW)🧹 Nettoyage de l'environnement Incus$(NC)"
	@echo "$(RED)⚠️ Ceci va supprimer tous les containers et profils Veza$(NC)"
	@read -p "Continuer ? (y/N) " confirm && [ "$$confirm" = "y" ] || exit 1
	@chmod +x scripts/deploy.sh
	@./scripts/deploy.sh clean

incus-dev: ## Démarrer l'environnement de développement Incus complet
	@echo "$(YELLOW)⚠️ Commande legacy - utilisez 'make deploy-dev' à la place$(NC)"
	@make deploy-dev

# =============================================================================
# Docker (ARCHIVÉ - Utilisez Incus à la place)
# =============================================================================

docker-build: ## [ARCHIVE] Construire les images Docker - UTILISEZ INCUS
	@echo "$(RED)⚠️ Docker est archivé, utilisez 'make incus-dev' à la place$(NC)"
	@echo "$(BLUE)📁 Fichiers Docker disponibles dans : archive/docker/$(NC)"
	@exit 1

docker-up: ## [ARCHIVE] Démarrer avec Docker Compose - UTILISEZ INCUS
	@echo "$(RED)⚠️ Docker est archivé, utilisez 'make incus-dev' à la place$(NC)"
	@echo "$(BLUE)📁 Fichiers Docker disponibles dans : archive/docker/$(NC)"
	@exit 1

docker-down: ## [ARCHIVE] Arrêter Docker Compose - UTILISEZ INCUS
	@echo "$(RED)⚠️ Docker est archivé, utilisez 'make incus-stop' à la place$(NC)"
	@exit 1

docker-restart: ## [ARCHIVE] Redémarrer Docker Compose - UTILISEZ INCUS
	@echo "$(RED)⚠️ Docker est archivé, utilisez 'make incus-restart CONTAINER=all' à la place$(NC)"
	@exit 1

docker-logs: ## [ARCHIVE] Voir les logs Docker - UTILISEZ INCUS
	@echo "$(RED)⚠️ Docker est archivé, utilisez 'make incus-logs CONTAINER=all' à la place$(NC)"
	@exit 1

docker-clean: ## [ARCHIVE] Nettoyer Docker - UTILISEZ INCUS
	@echo "$(RED)⚠️ Docker est archivé, utilisez 'make incus-clean' à la place$(NC)"
	@exit 1

# =============================================================================
# Base de Données
# =============================================================================

db-setup: ## Initialiser la base de données (Incus)
	@echo "$(BLUE)🗄️ Initialisation de la base de données$(NC)"
	@incus exec veza-postgres -- psql -U veza_user -d veza_db -f /tmp/init-db.sql
	@echo "$(GREEN)✅ Base de données initialisée$(NC)"

db-migrate: ## Exécuter les migrations
	@echo "$(BLUE)📊 Exécution des migrations$(NC)"
	@cd veza-backend-api && go run cmd/migrate/main.go
	@echo "$(GREEN)✅ Migrations exécutées$(NC)"

db-migrate-existing: ## Migrer une base de données existante (Incus)
	@echo "$(BLUE)🔄 Migration de la base de données existante$(NC)"
	@incus exec veza-postgres -- psql -U veza_user -d veza_db -f /tmp/migrate-existing-db.sql
	@echo "$(GREEN)✅ Migration existante terminée$(NC)"

db-reset: ## Réinitialiser la base de données (Incus)
	@echo "$(YELLOW)⚠️ Réinitialisation de la base de données$(NC)"
	@echo "$(RED)⚠️ Ceci va supprimer toutes les données !$(NC)"
	@read -p "Continuer ? (y/N) " confirm && [ "$$confirm" = "y" ] || exit 1
	@incus stop veza-postgres 2>/dev/null || true
	@incus delete veza-postgres 2>/dev/null || true
	@rm -rf data/postgres/*
	@make incus-deploy
	@sleep 10
	@make db-setup

# =============================================================================
# Tests
# =============================================================================

test: ## Exécuter tous les tests
	@echo "$(BLUE)🧪 Exécution de tous les tests$(NC)"
	@make test-frontend
	@make test-backend
	@make test-chat
	@make test-stream
	@echo "$(GREEN)✅ Tous les tests terminés$(NC)"

test-frontend: ## Tests du frontend React
	@echo "$(YELLOW)🎨 Tests Frontend...$(NC)"
	@cd veza-frontend && npm test

test-backend: ## Tests du backend Go
	@echo "$(YELLOW)⚙️ Tests Backend...$(NC)"
	@cd veza-backend-api && go test ./...

test-chat: ## Tests du chat server Rust
	@echo "$(YELLOW)💬 Tests Chat Server...$(NC)"
	@cd veza-chat-server && cargo test

test-stream: ## Tests du stream server Rust
	@echo "$(YELLOW)🎵 Tests Stream Server...$(NC)"
	@cd veza-stream-server && cargo test

test-e2e: ## Tests end-to-end
	@echo "$(YELLOW)🔄 Tests End-to-End...$(NC)"
	@cd veza-frontend && npm run test:e2e

# =============================================================================
# Monitoring et Logs
# =============================================================================

logs: ## Voir les logs de tous les services (Incus)
	@echo "$(BLUE)📝 Logs des services$(NC)"
	@make incus-logs CONTAINER=all

logs-backend: ## Logs du backend uniquement (Incus)
	@make incus-logs CONTAINER=veza-backend

logs-chat: ## Logs du chat server uniquement (Incus)
	@make incus-logs CONTAINER=veza-chat

logs-stream: ## Logs du stream server uniquement (Incus)
	@make incus-logs CONTAINER=veza-stream

logs-frontend: ## Logs du frontend uniquement (Incus)
	@make incus-logs CONTAINER=veza-frontend

logs-postgres: ## Logs de PostgreSQL (Incus)
	@make incus-logs CONTAINER=veza-postgres

logs-redis: ## Logs de Redis (Incus)
	@make incus-logs CONTAINER=veza-redis

logs-haproxy: ## Logs d'HAProxy (Incus)
	@make incus-logs CONTAINER=veza-haproxy

health: ## Vérifier la santé des services (Incus)
	@echo "$(BLUE)🏥 Vérification de la santé des services$(NC)"
	@make incus-status

# =============================================================================
# Gestion ZFS Storage
# =============================================================================

zfs-status: ## Afficher le statut du pool ZFS
	@chmod +x scripts/incus-zfs-manage.sh
	@./scripts/incus-zfs-manage.sh status

zfs-snapshot: ## Créer des snapshots ZFS des volumes de stockage
	@chmod +x scripts/incus-zfs-manage.sh
	@./scripts/incus-zfs-manage.sh snapshot

zfs-list: ## Lister tous les snapshots ZFS
	@chmod +x scripts/incus-zfs-manage.sh
	@./scripts/incus-zfs-manage.sh list-snapshots

zfs-compress: ## Afficher les statistiques de compression ZFS
	@chmod +x scripts/incus-zfs-manage.sh
	@./scripts/incus-zfs-manage.sh compress-stats

zfs-cleanup: ## Nettoyer les anciens snapshots ZFS (>30 jours)
	@chmod +x scripts/incus-zfs-manage.sh
	@./scripts/incus-zfs-manage.sh cleanup

zfs-expand: ## Étendre la taille des volumes ZFS
	@chmod +x scripts/incus-zfs-manage.sh
	@./scripts/incus-zfs-manage.sh expand

zfs-monitor: ## Monitoring en temps réel du stockage ZFS
	@chmod +x scripts/incus-zfs-manage.sh
	@./scripts/incus-zfs-manage.sh monitor

zfs-restore: ## Instructions pour restaurer depuis un snapshot ZFS
	@chmod +x scripts/incus-zfs-manage.sh
	@./scripts/incus-zfs-manage.sh restore

# =============================================================================
# Utilitaires
# =============================================================================

clean: ## Nettoyer les fichiers de build
	@echo "$(BLUE)🧹 Nettoyage des fichiers de build$(NC)"
	@rm -rf veza-frontend/dist
	@rm -rf veza-frontend/node_modules/.cache
	@rm -rf veza-backend-api/bin
	@rm -rf veza-chat-server/target
	@rm -rf veza-stream-server/target
	@rm -rf logs/*
	@echo "$(GREEN)✅ Nettoyage terminé$(NC)"

backup: ## Sauvegarder la base de données (Incus)
	@echo "$(BLUE)💾 Sauvegarde de la base de données$(NC)"
	@mkdir -p backups
	@incus exec veza-postgres -- pg_dump -U veza_user veza_db > backups/backup_$(shell date +%Y%m%d_%H%M%S).sql
	@echo "$(GREEN)✅ Sauvegarde créée dans backups/$(NC)"

format: ## Formater le code (tous les langages)
	@echo "$(BLUE)✨ Formatage du code$(NC)"
	@cd veza-frontend && npm run lint --fix
	@cd veza-backend-api && go fmt ./...
	@cd veza-chat-server && cargo fmt
	@cd veza-stream-server && cargo fmt
	@echo "$(GREEN)✅ Code formaté$(NC)"

# =============================================================================
# Environnements
# =============================================================================

env-dev: ## Configurer pour le développement
	@echo "$(BLUE)🔧 Configuration développement$(NC)"
	@cp env.example .env
	@echo "$(GREEN)✅ Environnement de développement configuré$(NC)"

env-prod: ## Configurer pour la production
	@echo "$(BLUE)🔧 Configuration production$(NC)"
	@echo "$(YELLOW)⚠️ Vérifiez et modifiez les variables dans .env$(NC)"
	@echo "$(YELLOW)⚠️ N'oubliez pas de changer les mots de passe et secrets!$(NC)"

# =============================================================================
# Informations
# =============================================================================

status: ## Afficher le statut des services (Incus)
	@echo "$(BLUE)📊 Statut des services$(NC)"
	@make incus-status

info: ## Afficher les informations du projet Veza (Incus)
	@echo "$(BLUE)📋 Informations du projet Veza - Architecture Incus$(NC)"
	@echo "======================================================"
	@echo ""
	@echo "$(GREEN)🌐 Services Web:$(NC)"
	@echo "$(YELLOW)  • Application HAProxy:$(NC) http://10.100.0.16"
	@echo "$(YELLOW)  • HAProxy Stats:$(NC) http://10.100.0.16:8404/stats"
	@echo "$(YELLOW)  • Frontend React:$(NC) http://10.100.0.11:5173 (dev)"
	@echo "$(YELLOW)  • Backend API Go:$(NC) http://10.100.0.12:8080"
	@echo ""
	@echo "$(GREEN)🔌 Services WebSocket:$(NC)"
	@echo "$(YELLOW)  • Chat Server Rust:$(NC) ws://10.100.0.13:8081/ws"
	@echo "$(YELLOW)  • Stream Server Rust:$(NC) ws://10.100.0.14:8082/ws"
	@echo ""
	@echo "$(GREEN)🗄️ Services Infrastructure:$(NC)"
	@echo "$(YELLOW)  • PostgreSQL:$(NC) 10.100.0.15:5432"
	@echo "$(YELLOW)  • Redis:$(NC) 10.100.0.17:6379"
	@echo "$(YELLOW)  • ZFS Storage (NFS):$(NC) 10.100.0.18:2049"
	@echo ""
	@echo "$(GREEN)🐧 Containers Incus:$(NC)"
	@echo "$(YELLOW)  • veza-postgres$(NC)  - Base de données PostgreSQL"
	@echo "$(YELLOW)  • veza-redis$(NC)     - Cache Redis"
	@echo "$(YELLOW)  • veza-storage$(NC)   - Stockage ZFS + NFS"
	@echo "$(YELLOW)  • veza-backend$(NC)   - API Backend Go"
	@echo "$(YELLOW)  • veza-chat$(NC)      - Serveur Chat Rust"
	@echo "$(YELLOW)  • veza-stream$(NC)    - Serveur Stream Rust"
	@echo "$(YELLOW)  • veza-frontend$(NC)  - Interface React"
	@echo "$(YELLOW)  • veza-haproxy$(NC)   - Load Balancer HAProxy"
	@echo ""
	@echo "$(GREEN)📖 Documentation:$(NC) Voir les dossiers docs/ dans chaque projet"
	@echo "$(GREEN)🔧 Configuration:$(NC) Voir le fichier .env"
	@echo "$(GREEN)🔧 Migration BDD existante:$(NC) make db-migrate-existing"
	@echo "$(GREEN)🚀 Démarrage rapide:$(NC) make dev (ou make incus-dev)" 