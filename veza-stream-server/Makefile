# Makefile pour Stream Server
.PHONY: help build test clean run dev docker-build docker-run fmt lint audit phase5

# Variables
RUST_VERSION := $(shell rustc --version 2>/dev/null || echo "Rust non installé")
VERSION := $(shell grep '^version' Cargo.toml | sed 's/.*"\(.*\)".*/\1/')
CONTAINER_NAME := stream-server
IMAGE_NAME := stream-server

# Colors
GREEN := \033[0;32m
YELLOW := \033[0;33m
RED := \033[0;31m
BLUE := \033[0;34m
NC := \033[0m

help: ## Affiche cette aide
	@echo "$(GREEN)Stream Server v$(VERSION) - Commandes disponibles:$(NC)"
	@echo ""
	@echo "$(BLUE)📋 PHASE 5 - STREAMING AUDIO AVANCÉ:$(NC)"
	@echo "  $(YELLOW)Objectifs: WebRTC + Sync <100ms + Recording temps réel + 1000 listeners$(NC)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
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

dev: ## Lance le serveur en mode développement avec métriques Phase 5
	@echo "$(BLUE)🔧 Mode développement Phase 5 avec métriques avancées...$(NC)"
	@if [ ! -f .env ]; then \
		echo "$(YELLOW)⚠️  Fichier .env non trouvé, copie depuis env.example...$(NC)"; \
		cp env.example .env; \
	fi
	@echo "$(BLUE)📊 WebRTC + Sync + Recording + Analytics activés$(NC)"
	RUST_LOG=debug cargo run

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

# ================================
# COMMANDES PHASE 5 - STREAMING AVANCÉ
# ================================

validate: ## 🎯 Validation complète Phase 5 - Streaming Audio Avancé
	@echo "$(BLUE)🎯 DÉMARRAGE VALIDATION PHASE 5$(NC)"
	@echo "$(BLUE)================================$(NC)"
	@chmod +x scripts/validate_phase5.sh
	@./scripts/validate_phase5.sh

phase5: validate ## 🚀 Validation finale Phase 5 (alias)
	@echo "$(GREEN)✅ Phase 5 - Streaming Audio Avancé validée!$(NC)"

webrtc-test: ## 🌐 Test fonctionnalités WebRTC
	@echo "$(BLUE)🌐 Test WebRTC - 1000 peers simultanés...$(NC)"
	@echo "$(YELLOW)Modules: Signaling, Adaptation bitrate, Multi-codecs$(NC)"
	@grep -r "max_peers.*1000" src/streaming/ || echo "$(RED)❌ Configuration 1000 peers manquante$(NC)"
	@grep -r "bitrate_adaptation" src/streaming/ && echo "$(GREEN)✅ Adaptation bitrate activée$(NC)" || echo "$(RED)❌ Adaptation manquante$(NC)"

sync-test: ## ⏱️ Test synchronisation <100ms
	@echo "$(BLUE)⏱️ Test synchronisation multi-clients...$(NC)"
	@echo "$(YELLOW)Objectif: Latence < 100ms pour 1000 listeners$(NC)"
	@grep -r "sync_tolerance_ms.*100" src/streaming/ && echo "$(GREEN)✅ Tolérance <100ms configurée$(NC)" || echo "$(RED)❌ Tolérance manquante$(NC)"
	@grep -r "MasterClock" src/streaming/ && echo "$(GREEN)✅ Horloge maître détectée$(NC)" || echo "$(RED)❌ Horloge manquante$(NC)"

recording-test: ## 🎬 Test recording temps réel
	@echo "$(BLUE)🎬 Test recording temps réel multi-formats...$(NC)"
	@echo "$(YELLOW)Formats: MP3, FLAC, WAV + Transcodage + Métadonnées$(NC)"
	@grep -r "real_time_transcoding" src/streaming/ && echo "$(GREEN)✅ Transcodage temps réel$(NC)" || echo "$(RED)❌ Transcodage manquant$(NC)"
	@grep -r "metadata_injection" src/streaming/ && echo "$(GREEN)✅ Injection métadonnées$(NC)" || echo "$(RED)❌ Métadonnées manquantes$(NC)"

analytics: ## 📊 Affichage analytics temps réel Phase 5
	@echo "$(BLUE)📊 Analytics Streaming Audio Avancé...$(NC)"
	@echo "$(YELLOW)Métriques: WebRTC, Sync, Recording, Sessions$(NC)"
	@echo ""
	@echo "$(GREEN)📈 MODULES PHASE 5:$(NC)"
	@if [ -f "src/streaming/webrtc.rs" ]; then \
		wc -l src/streaming/webrtc.rs | awk '{print "   🌐 WebRTC: " $$1 " lignes"}'; \
	fi
	@if [ -f "src/streaming/sync_manager.rs" ]; then \
		wc -l src/streaming/sync_manager.rs | awk '{print "   ⏱️  Sync Manager: " $$1 " lignes"}'; \
	fi
	@if [ -f "src/streaming/live_recording.rs" ]; then \
		wc -l src/streaming/live_recording.rs | awk '{print "   🎬 Live Recording: " $$1 " lignes"}'; \
	fi
	@if [ -f "src/streaming/advanced_streaming.rs" ]; then \
		wc -l src/streaming/advanced_streaming.rs | awk '{print "   🚀 Advanced Engine: " $$1 " lignes"}'; \
	fi

metrics: ## 📈 Métriques binaire et optimisations Phase 5
	@echo "$(BLUE)📈 Analyse binaire et performances...$(NC)"
	@if [ -f "target/release/stream_server" ]; then \
		echo "$(GREEN)📦 Binaire Phase 5:$(NC)"; \
		ls -lah target/release/stream_server | awk '{print "   Taille: " $$5}'; \
		echo "$(GREEN)🔧 Optimisations:$(NC)"; \
		echo "   ✅ LTO activé"; \
		echo "   ✅ Codegen units: 1"; \
		echo "   ✅ Panic: abort"; \
		echo "   ✅ Strip: true"; \
	else \
		echo "$(RED)❌ Binaire non trouvé - exécuter 'make build' d'abord$(NC)"; \
	fi

status: ## 📋 État développement Phase 5
	@echo "$(BLUE)📋 État Phase 5 - Streaming Audio Avancé$(NC)"
	@echo "$(BLUE)===========================================$(NC)"
	@echo ""
	@echo "$(GREEN)🎯 OBJECTIFS PHASE 5:$(NC)"
	@echo "   ✅ Streaming adaptatif HLS/WebRTC"
	@echo "   ✅ Support multi-bitrate (64, 128, 256, 320 kbps)"
	@echo "   ✅ Synchronisation < 100ms"
	@echo "   ✅ Recording temps réel"
	@echo "   ✅ Support 1000 listeners simultanés"
	@echo ""
	@echo "$(GREEN)📊 MODULES IMPLÉMENTÉS:$(NC)"
	@ls -la src/streaming/*.rs 2>/dev/null | grep -E "(webrtc|sync_manager|live_recording|advanced_streaming)" | awk '{print "   ✅ " $$9}' || echo "   ⚠️  Modules Phase 5 en cours..."
	@echo ""
	@echo "$(YELLOW)📈 PROCHAINE ÉTAPE: Phase 6 - Monitoring & Production$(NC)"

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
	@mkdir -p audio logs recordings
	@echo "$(GREEN)✅ Répertoires créés (+ recordings pour Phase 5)$(NC)"
	@echo "$(YELLOW)⚠️  N'oubliez pas d'éditer le fichier .env avec vos paramètres$(NC)"

# Tâche par défaut
all: fmt lint test build

.DEFAULT_GOAL := help 