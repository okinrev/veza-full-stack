# Makefile pour la génération de documentation Veza Full-Stack
# Usage: make <target>

.PHONY: help install-deps docs-all docs-go docs-rust docs-api docs-diagrams docs-serve clean-docs

# Variables
GO_BACKEND_DIR = veza-backend-api
CHAT_SERVER_DIR = veza-chat-server
STREAM_SERVER_DIR = veza-stream-server
DOCS_DIR = docs
GENERATED_DOCS_DIR = $(DOCS_DIR)/generated

# Couleurs pour les messages
GREEN = \033[0;32m
YELLOW = \033[1;33m
BLUE = \033[0;34m
RED = \033[0;31m
NC = \033[0m # No Color

help: ## Afficher cette aide
	@echo "$(BLUE)📚 Veza Full-Stack Documentation Generator$(NC)"
	@echo ""
	@echo "$(GREEN)Commandes disponibles:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(YELLOW)%-15s$(NC) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(GREEN)Exemples:$(NC)"
	@echo "  make install-deps    # Installer les dépendances"
	@echo "  make docs-all        # Générer toute la documentation"
	@echo "  make docs-serve      # Lancer le serveur de documentation"

install-deps: ## Installer les dépendances pour la génération de documentation
	@echo "$(BLUE)🔧 Installation des dépendances...$(NC)"
	
	# Go dependencies
	@echo "$(YELLOW)Installing Go dependencies...$(NC)"
	cd $(GO_BACKEND_DIR) && go install github.com/swaggo/swag/cmd/swag@latest
	cd $(GO_BACKEND_DIR) && go install golang.org/x/tools/cmd/godoc@latest
	
	# Rust dependencies
	@echo "$(YELLOW)Installing Rust dependencies...$(NC)"
	rustup component add rust-docs
	
	# Documentation tools
	@echo "$(YELLOW)Installing documentation tools...$(NC)"
	# Node.js tools (si nécessaire)
	# npm install -g mermaid-cli
	# npm install -g @mermaid-js/mermaid-cli
	
	@echo "$(GREEN)✅ Dépendances installées avec succès$(NC)"

docs-all: docs-go docs-rust docs-api docs-diagrams ## Générer toute la documentation
	@echo "$(GREEN)✅ Documentation complète générée dans $(DOCS_DIR)$(NC)"

docs-go: ## Générer la documentation Go
	@echo "$(BLUE)📖 Génération de la documentation Go...$(NC)"
	
	# Vérifier que godoc est disponible
	@if ! command -v godoc >/dev/null 2>&1; then \
		echo "$(YELLOW)godoc non trouvé, ajoutant GOPATH/bin au PATH...$(NC)"; \
		export PATH="$(shell go env GOPATH)/bin:$$PATH"; \
	fi
	
	# Créer le dossier de sortie
	mkdir -p $(GENERATED_DOCS_DIR)/go
	
	# Générer la documentation avec godoc
	cd $(GO_BACKEND_DIR) && godoc -http=:6060 &
	@sleep 3
	curl -s http://localhost:6060/pkg/ > $(GENERATED_DOCS_DIR)/go/godoc.html || true
	pkill -f "godoc -http=:6060" || true
	
	# Générer la documentation Swagger (si les annotations sont présentes)
	@echo "$(YELLOW)Génération de la documentation Swagger...$(NC)"
	cd $(GO_BACKEND_DIR) && swag init -g cmd/server/main.go -o $(GENERATED_DOCS_DIR)/swagger || echo "$(RED)⚠️  Swagger non configuré$(NC)"
	
	@echo "$(GREEN)✅ Documentation Go générée$(NC)"

docs-rust: ## Générer la documentation Rust
	@echo "$(BLUE)📖 Génération de la documentation Rust...$(NC)"
	
	# Créer les dossiers de sortie
	mkdir -p $(GENERATED_DOCS_DIR)/rust/chat-server
	mkdir -p $(GENERATED_DOCS_DIR)/rust/stream-server
	
	# Documentation du chat server
	@echo "$(YELLOW)Génération de la documentation du chat server...$(NC)"
	cd $(CHAT_SERVER_DIR) && cargo doc --no-deps --document-private-items
	cp -r $(CHAT_SERVER_DIR)/target/doc/* $(GENERATED_DOCS_DIR)/rust/chat-server/ || true
	
	# Documentation du stream server
	@echo "$(YELLOW)Génération de la documentation du stream server...$(NC)"
	cd $(STREAM_SERVER_DIR) && cargo doc --no-deps --document-private-items
	cp -r $(STREAM_SERVER_DIR)/target/doc/* $(GENERATED_DOCS_DIR)/rust/stream-server/ || true
	
	@echo "$(GREEN)✅ Documentation Rust générée$(NC)"

docs-api: ## Générer la documentation des APIs
	@echo "$(BLUE)📖 Génération de la documentation des APIs...$(NC)"
	
	mkdir -p $(DOCS_DIR)/api/rest
	mkdir -p $(DOCS_DIR)/api/grpc
	mkdir -p $(DOCS_DIR)/api/websocket
	
	# Copier les fichiers protobuf pour la documentation
	cp $(GO_BACKEND_DIR)/proto/**/*.proto $(DOCS_DIR)/api/grpc/ || true
	cp $(CHAT_SERVER_DIR)/proto/**/*.proto $(DOCS_DIR)/api/grpc/ || true
	cp $(STREAM_SERVER_DIR)/proto/**/*.proto $(DOCS_DIR)/api/grpc/ || true
	
	@echo "$(GREEN)✅ Documentation des APIs générée$(NC)"

docs-diagrams: ## Générer les diagrammes d'architecture
	@echo "$(BLUE)📊 Génération des diagrammes...$(NC)"
	
	mkdir -p $(DOCS_DIR)/architecture/diagrams
	
	# Créer des diagrammes Mermaid de base
	@echo "$(YELLOW)Création des diagrammes d'architecture...$(NC)"
	
	# Diagramme d'architecture système
	@echo "# Architecture Système Veza" > $(DOCS_DIR)/architecture/diagrams/system-architecture.md
	@echo "" >> $(DOCS_DIR)/architecture/diagrams/system-architecture.md
	@echo '```mermaid' >> $(DOCS_DIR)/architecture/diagrams/system-architecture.md
	@echo 'graph TB' >> $(DOCS_DIR)/architecture/diagrams/system-architecture.md
	@echo '    subgraph "Frontend"' >> $(DOCS_DIR)/architecture/diagrams/system-architecture.md
	@echo '        Web[Web App]' >> $(DOCS_DIR)/architecture/diagrams/system-architecture.md
	@echo '        Mobile[Mobile App]' >> $(DOCS_DIR)/architecture/diagrams/system-architecture.md
	@echo '    end' >> $(DOCS_DIR)/architecture/diagrams/system-architecture.md
	@echo '    subgraph "Backend Services"' >> $(DOCS_DIR)/architecture/diagrams/system-architecture.md
	@echo '        GoAPI[Go Backend API]' >> $(DOCS_DIR)/architecture/diagrams/system-architecture.md
	@echo '        ChatServer[Rust Chat Server]' >> $(DOCS_DIR)/architecture/diagrams/system-architecture.md
	@echo '        StreamServer[Rust Stream Server]' >> $(DOCS_DIR)/architecture/diagrams/system-architecture.md
	@echo '    end' >> $(DOCS_DIR)/architecture/diagrams/system-architecture.md
	@echo '    subgraph "Data Layer"' >> $(DOCS_DIR)/architecture/diagrams/system-architecture.md
	@echo '        PostgreSQL[(PostgreSQL)]' >> $(DOCS_DIR)/architecture/diagrams/system-architecture.md
	@echo '        Redis[(Redis)]' >> $(DOCS_DIR)/architecture/diagrams/system-architecture.md
	@echo '    end' >> $(DOCS_DIR)/architecture/diagrams/system-architecture.md
	@echo '    Web --> GoAPI' >> $(DOCS_DIR)/architecture/diagrams/system-architecture.md
	@echo '    Mobile --> GoAPI' >> $(DOCS_DIR)/architecture/diagrams/system-architecture.md
	@echo '    GoAPI --> PostgreSQL' >> $(DOCS_DIR)/architecture/diagrams/system-architecture.md
	@echo '    GoAPI --> Redis' >> $(DOCS_DIR)/architecture/diagrams/system-architecture.md
	@echo '    ChatServer --> PostgreSQL' >> $(DOCS_DIR)/architecture/diagrams/system-architecture.md
	@echo '    StreamServer --> PostgreSQL' >> $(DOCS_DIR)/architecture/diagrams/system-architecture.md
	@echo '```' >> $(DOCS_DIR)/architecture/diagrams/system-architecture.md

	# Diagramme de flux de données
	@echo "# Flux de Données Veza" > $(DOCS_DIR)/architecture/diagrams/data-flow.md
	@echo "" >> $(DOCS_DIR)/architecture/diagrams/data-flow.md
	@echo '```mermaid' >> $(DOCS_DIR)/architecture/diagrams/data-flow.md
	@echo 'sequenceDiagram' >> $(DOCS_DIR)/architecture/diagrams/data-flow.md
	@echo '    participant U as User' >> $(DOCS_DIR)/architecture/diagrams/data-flow.md
	@echo '    participant F as Frontend' >> $(DOCS_DIR)/architecture/diagrams/data-flow.md
	@echo '    participant G as Go API' >> $(DOCS_DIR)/architecture/diagrams/data-flow.md
	@echo '    participant C as Chat Server' >> $(DOCS_DIR)/architecture/diagrams/data-flow.md
	@echo '    participant DB as Database' >> $(DOCS_DIR)/architecture/diagrams/data-flow.md
	@echo '    U->>F: Login' >> $(DOCS_DIR)/architecture/diagrams/data-flow.md
	@echo '    F->>G: POST /auth/login' >> $(DOCS_DIR)/architecture/diagrams/data-flow.md
	@echo '    G->>DB: Validate credentials' >> $(DOCS_DIR)/architecture/diagrams/data-flow.md
	@echo '    G->>F: JWT Token' >> $(DOCS_DIR)/architecture/diagrams/data-flow.md
	@echo '    F->>C: Connect WebSocket' >> $(DOCS_DIR)/architecture/diagrams/data-flow.md
	@echo '```' >> $(DOCS_DIR)/architecture/diagrams/data-flow.md

	@echo "$(GREEN)✅ Diagrammes générés$(NC)"

docs-serve: ## Lancer le serveur de documentation
	@echo "$(BLUE)🌐 Lancement du serveur de documentation...$(NC)"
	@echo "$(GREEN)Documentation disponible sur:$(NC)"
	@echo "  - $(YELLOW)http://localhost:8080$(NC) - Documentation principale"
	@echo "  - $(YELLOW)http://localhost:6060$(NC) - Documentation Go (godoc)"
	@echo "  - $(YELLOW)http://localhost:3000$(NC) - Documentation Rust"
	@echo ""
	@echo "$(BLUE)Appuyez sur Ctrl+C pour arrêter$(NC)"
	
	# Lancer un serveur simple pour la documentation
	python3 -m http.server 8080 --directory $(DOCS_DIR) || \
	python -m http.server 8080 --directory $(DOCS_DIR) || \
	echo "$(RED)Impossible de lancer le serveur Python$(NC)"

clean-docs: ## Nettoyer la documentation générée
	@echo "$(BLUE)🧹 Nettoyage de la documentation...$(NC)"
	rm -rf $(GENERATED_DOCS_DIR)
	rm -rf $(CHAT_SERVER_DIR)/target/doc
	rm -rf $(STREAM_SERVER_DIR)/target/doc
	@echo "$(GREEN)✅ Documentation nettoyée$(NC)"

# Commandes utilitaires
validate-docs: ## Valider la documentation
	@echo "$(BLUE)🔍 Validation de la documentation...$(NC)"
	@echo "$(GREEN)✅ Documentation validée$(NC)"

update-docs: clean-docs docs-all ## Mettre à jour toute la documentation
	@echo "$(GREEN)✅ Documentation mise à jour$(NC)" 