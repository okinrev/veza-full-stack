# Makefile pour Stream Server
.PHONY: help build test clean run dev docker-build docker-run fmt lint audit

# Variables
RUST_VERSION := $(shell rustc --version 2>/dev/null || echo "Rust non installé")
VERSION := $(shell grep '^version' Cargo.toml | sed 's/.*"\(.*\)".*/\1/')
CONTAINER_NAME := stream-server
IMAGE_NAME := stream-server

# Colors
GREEN := \033[0;32m
YELLOW := \033[0;33m
RED := \033[0;31m
NC := \033[0m

help: ## Affiche cette aide
	@echo "$(GREEN)Stream Server v$(VERSION) - Commandes disponibles:$(NC)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "$(YELLOW)Rust version: $(RUST_VERSION)$(NC)"

build: ## Compile le projet en mode release
	@echo "$(GREEN)🔨 Compilation du projet...$(NC)"
	cargo build --release
	@echo "$(GREEN)✅ Compilation terminée$(NC)"

build-dev: ## Compile le projet en mode debug
	@echo "$(GREEN)🔨 Compilation en mode développement...$(NC)"
	cargo build
	@echo "$(GREEN)✅ Compilation terminée$(NC)"

build-tools: ## Compile les outils utilitaires
	@echo "$(GREEN)🔧 Compilation des outils...$(NC)"
	cd tools && cargo build --release
	@echo "$(GREEN)✅ Outils compilés dans tools/target/release/$(NC)"

test: ## Lance tous les tests
	@echo "$(GREEN)🧪 Exécution des tests...$(NC)"
	cargo test --all
	@echo "$(GREEN)✅ Tests terminés$(NC)"

clean: ## Nettoie les fichiers de build
	@echo "$(YELLOW)🧹 Nettoyage...$(NC)"
	cargo clean
	cd tools && cargo clean
	@echo "$(GREEN)✅ Nettoyage terminé$(NC)"

run: build ## Compile et lance le serveur
	@echo "$(GREEN)🚀 Démarrage du serveur...$(NC)"
	@if [ ! -f .env ]; then \
		echo "$(YELLOW)⚠️  Fichier .env non trouvé, copie depuis env.example...$(NC)"; \
		cp env.example .env; \
	fi
	./target/release/stream_server

dev: ## Lance le serveur en mode développement
	@echo "$(GREEN)🔧 Mode développement...$(NC)"
	@if [ ! -f .env ]; then \
		echo "$(YELLOW)⚠️  Fichier .env non trouvé, copie depuis env.example...$(NC)"; \
		cp env.example .env; \
	fi
	cargo run

fmt: ## Formate le code source
	@echo "$(GREEN)📝 Formatage du code...$(NC)"
	cargo fmt --all
	cd tools && cargo fmt --all
	@echo "$(GREEN)✅ Code formaté$(NC)"

lint: ## Vérifie la qualité du code avec clippy
	@echo "$(GREEN)🔍 Vérification avec clippy...$(NC)"
	cargo clippy --all-targets --all-features -- -D warnings
	cd tools && cargo clippy --all-targets --all-features -- -D warnings
	@echo "$(GREEN)✅ Vérification terminée$(NC)"

audit: ## Audit de sécurité des dépendances
	@echo "$(GREEN)🔒 Audit de sécurité...$(NC)"
	cargo audit
	@echo "$(GREEN)✅ Audit terminé$(NC)"

docker-build: ## Construit l'image Docker
	@echo "$(GREEN)🐳 Construction de l'image Docker...$(NC)"
	docker build -t $(IMAGE_NAME):$(VERSION) -t $(IMAGE_NAME):latest .
	@echo "$(GREEN)✅ Image Docker construite: $(IMAGE_NAME):$(VERSION)$(NC)"

docker-run: docker-build ## Lance le serveur dans Docker
	@echo "$(GREEN)🐳 Démarrage du conteneur Docker...$(NC)"
	docker run -d \
		--name $(CONTAINER_NAME) \
		-p 8082:8082 \
		-v "$(PWD)/audio:/app/audio:ro" \
		-v "$(PWD)/.env:/app/.env:ro" \
		$(IMAGE_NAME):latest
	@echo "$(GREEN)✅ Conteneur démarré: http://localhost:8082$(NC)"

docker-stop: ## Arrête et supprime le conteneur Docker
	@echo "$(YELLOW)⏹️  Arrêt du conteneur...$(NC)"
	-docker stop $(CONTAINER_NAME) 2>/dev/null || true
	-docker rm $(CONTAINER_NAME) 2>/dev/null || true
	@echo "$(GREEN)✅ Conteneur arrêté$(NC)"

docker-logs: ## Affiche les logs du conteneur
	@echo "$(GREEN)📋 Logs du conteneur:$(NC)"
	docker logs -f $(CONTAINER_NAME)

health-check: ## Vérifie la santé du serveur
	@echo "$(GREEN)🏥 Vérification de la santé du serveur...$(NC)"
	@curl -s http://localhost:8082/health | jq . || echo "$(RED)❌ Serveur non accessible$(NC)"

install-deps: ## Installe les dépendances système nécessaires
	@echo "$(GREEN)📦 Installation des dépendances...$(NC)"
	@if command -v apt-get >/dev/null 2>&1; then \
		sudo apt-get update && sudo apt-get install -y pkg-config libssl-dev; \
	elif command -v dnf >/dev/null 2>&1; then \
		sudo dnf install -y pkgconfig openssl-devel; \
	elif command -v pacman >/dev/null 2>&1; then \
		sudo pacman -S pkg-config openssl; \
	else \
		echo "$(YELLOW)⚠️  Gestionnaire de paquets non reconnu$(NC)"; \
	fi
	@echo "$(GREEN)✅ Dépendances installées$(NC)"

setup: install-deps ## Configuration initiale du projet
	@echo "$(GREEN)🛠️  Configuration du projet...$(NC)"
	@if [ ! -f .env ]; then \
		cp env.example .env; \
		echo "$(GREEN)✅ Fichier .env créé$(NC)"; \
	fi
	@mkdir -p audio logs
	@echo "$(GREEN)✅ Répertoires créés$(NC)"
	@echo "$(YELLOW)⚠️  N'oubliez pas d'éditer le fichier .env avec vos paramètres$(NC)"

# Tâche par défaut
all: fmt lint test build

.DEFAULT_GOAL := help 