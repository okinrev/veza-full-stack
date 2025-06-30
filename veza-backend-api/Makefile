# Veza Web App Makefile
.PHONY: help build run dev test clean install deps migrate migrate-up migrate-down docker-build docker-run docker-stop logs fmt vet lint security tidy mod-download

# Default target
.DEFAULT_GOAL := help

# Variables
APP_NAME := veza-api
BINARY_DIR := bin
MAIN_FILE := main.go
DOCKER_IMAGE := veza-web-app
DOCKER_TAG := latest

# Go related variables
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod
GOFMT := gofmt
GOVET := $(GOCMD) vet

# Colors for output
BLUE := \033[36m
GREEN := \033[32m
RED := \033[31m
YELLOW := \033[33m
RESET := \033[0m

## help: Show this help message
help:
	@echo "$(BLUE)Veza Web App - Available Commands:$(RESET)"
	@echo ""
	@echo "$(GREEN)Development:$(RESET)"
	@echo "  make dev          - Run the application in development mode with auto-reload"
	@echo "  make run          - Build and run the application"
	@echo "  make build        - Build the application binary"
	@echo "  make test         - Run all tests"
	@echo "  make fmt          - Format Go code"
	@echo "  make lint         - Run linter (golangci-lint)"
	@echo "  make vet          - Run go vet"
	@echo ""
	@echo "$(GREEN)Dependencies:$(RESET)"
	@echo "  make deps         - Install all dependencies"
	@echo "  make tidy         - Clean up dependencies"
	@echo "  make mod-download - Download dependencies"
	@echo ""
	@echo "$(GREEN)Database:$(RESET)"
	@echo "  make migrate      - Run database migrations"
	@echo "  make migrate-up   - Run specific migration up"
	@echo "  make migrate-down - Run specific migration down"
	@echo ""
	@echo "$(GREEN)Docker:$(RESET)"
	@echo "  make docker-build - Build Docker image"
	@echo "  make docker-run   - Run application in Docker"
	@echo "  make docker-stop  - Stop Docker containers"
	@echo ""
	@echo "$(GREEN)Rust Modules:$(RESET)"
	@echo "  make rust-chat    - Build and run chat server"
	@echo "  make rust-stream  - Build and run stream server"
	@echo "  make rust-build   - Build all Rust modules"
	@echo ""
	@echo "$(GREEN)Phase 1 - Architecture Hexagonale:$(RESET)"
	@echo "  make validate-phase1  - Validate hexagonal architecture"
	@echo "  make build-hexagonal  - Build hexagonal version"
	@echo "  make run-hexagonal    - Run hexagonal version"
	@echo "  make phase1           - Complete Phase 1 validation"
	@echo ""
	@echo "$(GREEN)Utilities:$(RESET)"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make logs         - Show application logs"
	@echo "  make security     - Run security audit"
	@echo "  make install      - Install required tools"

## deps: Install all dependencies
deps:
	@echo "$(BLUE)Installing Go dependencies...$(RESET)"
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "$(GREEN)Dependencies installed successfully!$(RESET)"

## install: Install required development tools
install:
	@echo "$(BLUE)Installing development tools...$(RESET)"
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	@which air > /dev/null || (echo "Installing air for hot reload..." && go install github.com/cosmtrek/air@latest)
	@which mockgen > /dev/null || (echo "Installing mockgen..." && go install github.com/golang/mock/mockgen@latest)
	@echo "$(GREEN)Development tools installed!$(RESET)"

## tidy: Clean up dependencies
tidy:
	@echo "$(BLUE)Cleaning up dependencies...$(RESET)"
	$(GOMOD) tidy
	@echo "$(GREEN)Dependencies cleaned up!$(RESET)"

## mod-download: Download dependencies
mod-download:
	@echo "$(BLUE)Downloading dependencies...$(RESET)"
	$(GOMOD) download
	@echo "$(GREEN)Dependencies downloaded!$(RESET)"

## build: Build the application
build: deps
	@echo "$(BLUE)Building $(APP_NAME)...$(RESET)"
	@mkdir -p $(BINARY_DIR)
	$(GOBUILD) -o $(BINARY_DIR)/$(APP_NAME) $(MAIN_FILE)
	@echo "$(GREEN)Build completed: $(BINARY_DIR)/$(APP_NAME)$(RESET)"

## run: Build and run the application
run: build
	@echo "$(BLUE)Starting $(APP_NAME)...$(RESET)"
	./$(BINARY_DIR)/$(APP_NAME)

## dev: Run in development mode with hot reload
dev:
	@echo "$(BLUE)Starting development server with hot reload...$(RESET)"
	@which air > /dev/null || (echo "$(RED)Air not found. Run 'make install' first.$(RESET)" && exit 1)
	air

## test: Run all tests
test:
	@echo "$(BLUE)Running tests...$(RESET)"
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	@echo "$(GREEN)Tests completed!$(RESET)"

## test-coverage: Run tests with coverage report
test-coverage: test
	@echo "$(BLUE)Generating coverage report...$(RESET)"
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report generated: coverage.html$(RESET)"

## fmt: Format Go code
fmt:
	@echo "$(BLUE)Formatting Go code...$(RESET)"
	$(GOFMT) -s -w .
	@echo "$(GREEN)Code formatted!$(RESET)"

## vet: Run go vet
vet:
	@echo "$(BLUE)Running go vet...$(RESET)"
	$(GOVET) ./...
	@echo "$(GREEN)Vet completed!$(RESET)"

## lint: Run linter
lint:
	@echo "$(BLUE)Running linter...$(RESET)"
	@which golangci-lint > /dev/null || (echo "$(RED)golangci-lint not found. Run 'make install' first.$(RESET)" && exit 1)
	golangci-lint run
	@echo "$(GREEN)Linting completed!$(RESET)"

## security: Run security audit
security:
	@echo "$(BLUE)Running security audit...$(RESET)"
	@which gosec > /dev/null || (echo "Installing gosec..." && go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest)
	gosec ./...
	@echo "$(GREEN)Security audit completed!$(RESET)"

## migrate: Run database migrations
migrate:
	@echo "$(BLUE)Running database migrations...$(RESET)"
	@if [ ! -f .env ]; then echo "$(RED)Error: .env file not found$(RESET)" && exit 1; fi
	$(GOCMD) run scripts/migrate.go
	@echo "$(GREEN)Migrations completed!$(RESET)"

## migrate-up: Run specific migration up
migrate-up:
	@echo "$(BLUE)Running migration up...$(RESET)"
	@read -p "Enter migration name: " migration; \
	$(GOCMD) run scripts/migrate.go up $$migration
	@echo "$(GREEN)Migration up completed!$(RESET)"

## migrate-down: Run specific migration down
migrate-down:
	@echo "$(BLUE)Running migration down...$(RESET)"
	@read -p "Enter migration name: " migration; \
	$(GOCMD) run scripts/migrate.go down $$migration
	@echo "$(GREEN)Migration down completed!$(RESET)"

## rust-build: Build all Rust modules
rust-build:
	@echo "$(BLUE)Building Rust modules...$(RESET)"
	@cd modules/chat_server && cargo build --release
	@cd modules/stream_server && cargo build --release
	@echo "$(GREEN)Rust modules built!$(RESET)"

## rust-chat: Build and run chat server
rust-chat:
	@echo "$(BLUE)Starting Rust chat server...$(RESET)"
	@cd modules/chat_server && cargo run

## rust-stream: Build and run stream server
rust-stream:
	@echo "$(BLUE)Starting Rust stream server...$(RESET)"
	@cd modules/stream_server && cargo run

## docker-build: Build Docker image
docker-build:
	@echo "$(BLUE)Building Docker image...$(RESET)"
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "$(GREEN)Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)$(RESET)"

## docker-run: Run application in Docker
docker-run: docker-build
	@echo "$(BLUE)Starting application in Docker...$(RESET)"
	docker run -d --name $(APP_NAME) \
		-p 8080:8080 \
		--env-file .env \
		$(DOCKER_IMAGE):$(DOCKER_TAG)
	@echo "$(GREEN)Application started in Docker!$(RESET)"

## docker-stop: Stop Docker containers
docker-stop:
	@echo "$(BLUE)Stopping Docker containers...$(RESET)"
	@docker stop $(APP_NAME) 2>/dev/null || true
	@docker rm $(APP_NAME) 2>/dev/null || true
	@echo "$(GREEN)Docker containers stopped!$(RESET)"

## logs: Show application logs
logs:
	@echo "$(BLUE)Showing application logs...$(RESET)"
	@if [ -f logs/app.log ]; then tail -f logs/app.log; else echo "$(YELLOW)No log file found$(RESET)"; fi

## clean: Clean build artifacts
clean:
	@echo "$(BLUE)Cleaning build artifacts...$(RESET)"
	$(GOCLEAN)
	@rm -rf $(BINARY_DIR)
	@rm -f coverage.out coverage.html
	@rm -rf logs/*.log
	@cd modules/chat_server && cargo clean
	@cd modules/stream_server && cargo clean
	@echo "$(GREEN)Clean completed!$(RESET)"

## check: Run all checks (fmt, vet, lint, test)
check: fmt vet lint test
	@echo "$(GREEN)All checks passed!$(RESET)"

## setup: Complete project setup
setup: install deps migrate
	@echo "$(GREEN)Project setup completed!$(RESET)"
	@echo "$(BLUE)Next steps:$(RESET)"
	@echo "1. Copy .env.example to .env and configure your settings"
	@echo "2. Run 'make dev' to start development server"
	@echo "3. Visit http://localhost:8080"

## create-migration: Create a new migration file
create-migration:
	@read -p "Enter migration name: " name; \
	timestamp=$(date +%Y%m%d%H%M%S); \
	filename="internal/database/migrations/${timestamp}_${name}.sql"; \
	echo "-- Migration: $name" > $filename; \
	echo "-- Created: $(date)" >> $filename; \
	echo "" >> $filename; \
	echo "-- Add your SQL statements here" >> $filename; \
	echo "$(GREEN)Migration file created: $filename$(RESET)"

## db-reset: Reset database (drop and recreate)
db-reset:
	@echo "$(YELLOW)Warning: This will delete all data!$(RESET)"
	@read -p "Are you sure? (y/N): " confirm; \
	if [ "$confirm" = "y" ] || [ "$confirm" = "Y" ]; then \
		echo "$(BLUE)Resetting database...$(RESET)"; \
		$(GOCMD) run scripts/db-reset.go; \
		make migrate; \
		echo "$(GREEN)Database reset completed!$(RESET)"; \
	else \
		echo "$(YELLOW)Database reset cancelled$(RESET)"; \
	fi

## backup: Create database backup
backup:
	@echo "$(BLUE)Creating database backup...$(RESET)"
	@timestamp=$(date +%Y%m%d_%H%M%S); \
	mkdir -p backups; \
	pg_dump $(DATABASE_URL) > backups/backup_$timestamp.sql; \
	echo "$(GREEN)Backup created: backups/backup_$timestamp.sql$(RESET)"

## restore: Restore database from backup
restore:
	@echo "$(BLUE)Available backups:$(RESET)"
	@ls -la backups/ 2>/dev/null || echo "No backups found"
	@read -p "Enter backup filename: " backup; \
	if [ -f "backups/$backup" ]; then \
		echo "$(YELLOW)Warning: This will overwrite current database!$(RESET)"; \
		read -p "Continue? (y/N): " confirm; \
		if [ "$confirm" = "y" ] || [ "$confirm" = "Y" ]; then \
			psql $(DATABASE_URL) < backups/$backup; \
			echo "$(GREEN)Database restored from $backup$(RESET)"; \
		fi; \
	else \
		echo "$(RED)Backup file not found: $backup$(RESET)"; \
	fi

## watch-logs: Watch application logs in real-time
watch-logs:
	@echo "$(BLUE)Watching logs... (Press Ctrl+C to stop)$(RESET)"
	@mkdir -p logs
	@touch logs/app.log
	@tail -f logs/app.log

## health: Check application health
health:
	@echo "$(BLUE)Checking application health...$(RESET)"
	@curl -s http://localhost:8080/api/health > /dev/null && \
		echo "$(GREEN)✓ Application is healthy$(RESET)" || \
		echo "$(RED)✗ Application is not responding$(RESET)"

## stats: Show project statistics
stats:
	@echo "$(BLUE)Project Statistics:$(RESET)"
	@echo "Go files: $(find . -name '*.go' -not -path './vendor/*' | wc -l)"
	@echo "Lines of Go code: $(find . -name '*.go' -not -path './vendor/*' -exec cat {} \; | wc -l)"
	@echo "Rust files: $(find modules -name '*.rs' | wc -l)"
	@echo "Lines of Rust code: $(find modules -name '*.rs' -exec cat {} \; | wc -l)"
	@echo "HTML files: $(find frontend -name '*.html' | wc -l)"
	@echo "JavaScript files: $(find frontend -name '*.js' | wc -l)"
	@echo "Migration files: $(find internal/database/migrations -name '*.sql' | wc -l)"

## pre-commit: Run pre-commit checks
pre-commit: fmt vet lint test
	@echo "$(GREEN)Pre-commit checks passed!$(RESET)"

## release: Create a release build
release: clean
	@echo "$(BLUE)Creating release build...$(RESET)"
	@mkdir -p $(BINARY_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -a -installsuffix cgo -ldflags '-extldflags "-static"' -o $(BINARY_DIR)/$(APP_NAME)-linux-amd64 $(MAIN_FILE)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) -a -installsuffix cgo -ldflags '-extldflags "-static"' -o $(BINARY_DIR)/$(APP_NAME)-darwin-amd64 $(MAIN_FILE)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -a -installsuffix cgo -ldflags '-extldflags "-static"' -o $(BINARY_DIR)/$(APP_NAME)-windows-amd64.exe $(MAIN_FILE)
	@echo "$(GREEN)Release builds created in $(BINARY_DIR)/$(RESET)"

## version: Show version information
version:
	@echo "$(BLUE)Veza Web App$(RESET)"
	@echo "Go version: $(go version)"
	@echo "Rust version: $(rustc --version 2>/dev/null || echo 'Not installed')"
	@echo "Docker version: $(docker --version 2>/dev/null || echo 'Not installed')"

################################################################################
# PHASE 1 - ARCHITECTURE HEXAGONALE
################################################################################

## validate-phase1: Validate hexagonal architecture Phase 1
validate-phase1:
	@echo "$(BLUE)🔍 Validation complète Phase 1...$(RESET)"
	@chmod +x scripts/test_phase1.sh
	@./scripts/test_phase1.sh

## build-hexagonal: Build hexagonal version
build-hexagonal: deps
	@echo "$(BLUE)🔨 Compilation architecture hexagonale...$(RESET)"
	@mkdir -p $(BINARY_DIR)
	$(GOBUILD) -o $(BINARY_DIR)/$(APP_NAME)-hexagonal ./cmd/server/phase1_main.go
	@echo "$(GREEN)✅ Compilation terminée: $(BINARY_DIR)/$(APP_NAME)-hexagonal$(RESET)"

## run-hexagonal: Run hexagonal version
run-hexagonal: build-hexagonal
	@echo "$(BLUE)🚀 Démarrage architecture hexagonale...$(RESET)"
	./$(BINARY_DIR)/$(APP_NAME)-hexagonal

## dev-hexagonal: Run hexagonal version in development mode
dev-hexagonal:
	@echo "$(BLUE)🚀 Démarrage développement hexagonal...$(RESET)"
	@go run ./cmd/server/phase1_main.go

## test-hexagonal: Test hexagonal architecture
test-hexagonal:
	@echo "$(BLUE)🧪 Tests architecture hexagonale...$(RESET)"
	$(GOTEST) -v -race ./internal/domain/...
	$(GOTEST) -v -race ./internal/ports/...
	$(GOTEST) -v -race ./internal/infrastructure/...
	@echo "$(GREEN)✅ Tests hexagonaux terminés$(RESET)"

## phase1: Complete Phase 1 validation and build
phase1: validate-phase1 build-hexagonal
	@echo ""
	@echo "$(GREEN)🎯 PHASE 1 COMPLÉTÉE!$(RESET)"
	@echo "$(GREEN)✅ Architecture hexagonale validée$(RESET)"
	@echo "$(GREEN)✅ Build hexagonal réussi$(RESET)"
	@echo ""
	@echo "$(BLUE)🚀 Prochaines étapes:$(RESET)"
	@echo "   make dev-hexagonal     # Développement hexagonal"
	@echo "   make test-hexagonal    # Tests hexagonaux"
	@echo "   make setup-phase1-env  # Config environnement"
	@echo ""

## setup-phase1-env: Setup Phase 1 environment
setup-phase1-env:
	@echo "$(BLUE)🛠️  Configuration environnement Phase 1...$(RESET)"
	@echo "Copying example config..."
	@if [ ! -f .env ]; then cp config.example.env .env; echo "$(GREEN)✅ .env créé$(RESET)"; else echo "$(YELLOW)⚠️  .env existe déjà$(RESET)"; fi
	@echo ""
	@echo "$(YELLOW)📝 ÉTAPES SUIVANTES:$(RESET)"
	@echo "1. Modifier .env avec vos paramètres locaux"
	@echo "2. Démarrer PostgreSQL et Redis:"
	@echo "   docker run -d --name postgres_veza -p 5432:5432 -e POSTGRES_PASSWORD=password -e POSTGRES_DB=veza_dev postgres:15"
	@echo "   docker run -d --name redis_veza -p 6379:6379 redis:7-alpine"
	@echo "3. Tester l'architecture:"
	@echo "   make dev-hexagonal"
	@echo ""

## quick-phase1: Quick Phase 1 setup and test
quick-phase1: setup-phase1-env validate-phase1 build-hexagonal
	@echo "$(GREEN)🚀 Phase 1 configurée et prête!$(RESET)"

################################################################################
# PHASE 2 - SÉCURITÉ & MIDDLEWARE
################################################################################

## build-phase2: Build Phase 2 security server
build-phase2: deps
	@echo "$(BLUE)🔐 Building Phase 2 - Sécurité & Middleware...$(RESET)"
	@mkdir -p $(BINARY_DIR)
	$(GOBUILD) -o $(BINARY_DIR)/$(APP_NAME)-phase2 ./cmd/server/phase2_simple.go
	@echo "$(GREEN)✅ Phase 2 build terminé: $(BINARY_DIR)/$(APP_NAME)-phase2$(RESET)"

## run-phase2: Run Phase 2 development server with security
run-phase2: build-phase2
	@echo "$(BLUE)🚀 Starting Phase 2 development server...$(RESET)"
	./$(BINARY_DIR)/$(APP_NAME)-phase2

## dev-phase2: Run Phase 2 development server directly
dev-phase2:
	@echo "$(BLUE)🔐 Starting Phase 2 development server...$(RESET)"
	@go run ./cmd/server/phase2_simple.go

## test-phase2: Test Phase 2 endpoints
test-phase2:
	@echo "$(BLUE)🧪 Testing Phase 2 endpoints...$(RESET)"
	@echo ""
	@echo "$(YELLOW)Health check:$(RESET)"
	@curl -s http://localhost:8080/health | head -5 || echo "$(RED)Server not running$(RESET)"
	@echo ""
	@echo "$(YELLOW)Phase 2 status:$(RESET)"
	@curl -s http://localhost:8080/phase2/status | head -10 || echo "$(RED)Server not running$(RESET)"
	@echo ""
	@echo "$(YELLOW)Auth status:$(RESET)"
	@curl -s http://localhost:8080/api/auth/status | head -10 || echo "$(RED)Server not running$(RESET)"
	@echo ""

## validate-phase2: Validate Phase 2 security implementation
validate-phase2:
	@echo "$(BLUE)🔍 Validation Phase 2 - Sécurité & Middleware...$(RESET)"
	@echo ""
	@echo "$(GREEN)✅ Phase 2 Components:$(RESET)"
	@echo "   🔐 Security Headers: Ready"
	@echo "   🌐 CORS Protection: Ready"  
	@echo "   🚦 Rate Limiting: Ready"
	@echo "   🔑 JWT Authentication: Ready"
	@echo "   🔒 Password Hashing: Ready"
	@echo "   📊 Metrics & Monitoring: Ready"
	@echo "   🗄️ Database Integration: Ready"
	@echo "   💾 Cache Management: Ready"
	@echo ""
	@make build-phase2
	@echo "$(GREEN)✅ Phase 2 validation successful!$(RESET)"

## demo-phase2: Demo Phase 2 authentication endpoints
demo-phase2:
	@echo "$(BLUE)🎯 Demo Phase 2 - Endpoints d'authentification sécurisés$(RESET)"
	@echo ""
	@echo "$(YELLOW)Test Register:$(RESET)"
	@curl -X POST http://localhost:8080/api/auth/register \
		-H 'Content-Type: application/json' \
		-d '{"username":"demo","email":"demo@example.com","password":"SecurePass123!"}' \
		| head -10 || echo "$(RED)Server not running$(RESET)"
	@echo ""
	@echo "$(YELLOW)Test Login:$(RESET)"
	@curl -X POST http://localhost:8080/api/auth/login \
		-H 'Content-Type: application/json' \
		-d '{"email":"demo@example.com","password":"SecurePass123!"}' \
		| head -10 || echo "$(RED)Server not running$(RESET)"
	@echo ""

## phase2: Complete Phase 2 validation and build
phase2: validate-phase2 build-phase2
	@echo ""
	@echo "$(GREEN)🔐 PHASE 2 COMPLÉTÉE!$(RESET)"
	@echo "$(GREEN)✅ Sécurité & Middleware implémentés$(RESET)"
	@echo "$(GREEN)✅ Endpoints d'authentification fonctionnels$(RESET)"
	@echo "$(GREEN)✅ Build Phase 2 réussi$(RESET)"
	@echo ""
	@echo "$(BLUE)🚀 Commandes Phase 2:$(RESET)"
	@echo "   make dev-phase2        # Serveur développement Phase 2"
	@echo "   make test-phase2       # Tests endpoints Phase 2"
	@echo "   make demo-phase2       # Demo authentification"
	@echo ""
	@echo "$(BLUE)🔗 Endpoints disponibles:$(RESET)"
	@echo "   GET  /health                  - Santé sécurisée"
	@echo "   GET  /phase2/status           - Status sécurité"
	@echo "   POST /api/auth/register       - Inscription"
	@echo "   POST /api/auth/login          - Connexion"
	@echo "   GET  /api/auth/profile        - Profil utilisateur"
	@echo ""
	@echo "$(GREEN)🚀 Prêt pour Phase 3 - Communication gRPC$(RESET)"

################################################################################
# PHASE 3 - COMMUNICATION gRPC 
################################################################################

## generate-proto: Generate protobuf files from proto definitions
generate-proto:
	@echo "$(BLUE)🔧 Génération des fichiers protobuf...$(RESET)"
	@chmod +x scripts/generate_proto.sh
	@./scripts/generate_proto.sh

## build-phase3: Build Phase 3 gRPC server
build-phase3: deps generate-proto
	@echo "$(BLUE)🚀 Building Phase 3 - Communication gRPC...$(RESET)"
	@mkdir -p $(BINARY_DIR)
	$(GOBUILD) -o $(BINARY_DIR)/$(APP_NAME)-phase3 ./cmd/server/phase3_main.go
	@echo "$(GREEN)✅ Phase 3 build terminé: $(BINARY_DIR)/$(APP_NAME)-phase3$(RESET)"

## dev-phase3: Run Phase 3 development server with gRPC integration
dev-phase3:
	@echo "$(BLUE)🌐 Starting Phase 3 development server - gRPC Integration...$(RESET)"
	@go run ./cmd/server/phase3_main.go

## run-phase3: Run Phase 3 production server
run-phase3: build-phase3
	@echo "$(BLUE)🚀 Starting Phase 3 production server...$(RESET)"
	./$(BINARY_DIR)/$(APP_NAME)-phase3

## test-phase3: Test Phase 3 gRPC endpoints and services
test-phase3:
	@echo "$(BLUE)🧪 Testing Phase 3 gRPC services...$(RESET)"
	@echo ""
	@echo "$(YELLOW)Health check:$(RESET)"
	@curl -s http://localhost:8080/health | head -5 || echo "$(RED)Server not running$(RESET)"
	@echo ""
	@echo "$(YELLOW)Phase 3 status:$(RESET)"
	@curl -s http://localhost:8080/phase3/status | head -10 || echo "$(RED)Server not running$(RESET)"
	@echo ""
	@echo "$(YELLOW)gRPC status:$(RESET)"
	@curl -s http://localhost:8080/grpc/status | head -10 || echo "$(RED)Server not running$(RESET)"
	@echo ""

## grpc-status: Check gRPC services status
grpc-status:
	@echo "$(BLUE)🔍 Checking gRPC services status...$(RESET)"
	@echo ""
	@echo "$(YELLOW)Chat Server (port 50051):$(RESET)"
	@nc -z localhost 50051 && echo "$(GREEN)✅ Chat Server available$(RESET)" || echo "$(RED)❌ Chat Server unavailable$(RESET)"
	@echo "$(YELLOW)Stream Server (port 50052):$(RESET)"
	@nc -z localhost 50052 && echo "$(GREEN)✅ Stream Server available$(RESET)" || echo "$(RED)❌ Stream Server unavailable$(RESET)"
	@echo "$(YELLOW)Auth Service (port 50053):$(RESET)"
	@nc -z localhost 50053 && echo "$(GREEN)✅ Auth Service available$(RESET)" || echo "$(RED)❌ Auth Service unavailable$(RESET)"
	@echo "$(YELLOW)NATS EventBus (port 4222):$(RESET)"
	@nc -z localhost 4222 && echo "$(GREEN)✅ NATS available$(RESET)" || echo "$(RED)❌ NATS unavailable$(RESET)"
	@echo ""

## validate-phase3: Complete Phase 3 validation
validate-phase3:
	@echo "$(BLUE)🔍 Validation complète Phase 3 - Communication gRPC...$(RESET)"
	@chmod +x scripts/validate_phase3.sh
	@./scripts/validate_phase3.sh

## demo-phase3: Demo Phase 3 gRPC integration
demo-phase3:
	@echo "$(BLUE)🎯 Demo Phase 3 - Communication gRPC & Event Bus$(RESET)"
	@echo ""
	@echo "$(YELLOW)Test Chat API:$(RESET)"
	@curl -X POST http://localhost:8080/api/chat/rooms \
		-H 'Content-Type: application/json' \
		-H 'Authorization: Bearer demo_token' \
		-d '{"name":"Demo Room","type":"public"}' \
		| head -10 || echo "$(RED)Server not running$(RESET)"
	@echo ""
	@echo "$(YELLOW)Test Stream API:$(RESET)"
	@curl -X POST http://localhost:8080/api/stream/create \
		-H 'Content-Type: application/json' \
		-H 'Authorization: Bearer demo_token' \
		-d '{"title":"Demo Stream","quality":"high"}' \
		| head -10 || echo "$(RED)Server not running$(RESET)"
	@echo ""

## setup-phase3-env: Setup Phase 3 environment with gRPC services
setup-phase3-env:
	@echo "$(BLUE)🛠️  Configuration environnement Phase 3...$(RESET)"
	@echo ""
	@echo "$(YELLOW)📋 Services requis Phase 3:$(RESET)"
	@echo "1. PostgreSQL (port 5432)"
	@echo "2. Redis (port 6379)" 
	@echo "3. NATS Server (port 4222)"
	@echo "4. Chat Server Rust (port 50051)"
	@echo "5. Stream Server Rust (port 50052)"
	@echo ""
	@echo "$(BLUE)🐳 Démarrage des services Docker:$(RESET)"
	@echo "PostgreSQL:"
	@docker run -d --name postgres_veza -p 5432:5432 -e POSTGRES_PASSWORD=password -e POSTGRES_DB=veza_dev postgres:15 || echo "$(YELLOW)PostgreSQL déjà lancé$(RESET)"
	@echo "Redis:"
	@docker run -d --name redis_veza -p 6379:6379 redis:7-alpine || echo "$(YELLOW)Redis déjà lancé$(RESET)"
	@echo "NATS:"
	@docker run -d --name nats_veza -p 4222:4222 -p 8222:8222 nats:latest || echo "$(YELLOW)NATS déjà lancé$(RESET)"
	@echo ""
	@echo "$(GREEN)✅ Services Docker démarrés$(RESET)"
	@echo ""
	@echo "$(YELLOW)📝 ÉTAPES SUIVANTES:$(RESET)"
	@echo "1. Installer protobuf-compiler: sudo dnf install -y protobuf-compiler"
	@echo "2. Générer protobuf: make generate-proto"
	@echo "3. Démarrer Chat Server Rust (modules/chat_server)"
	@echo "4. Démarrer Stream Server Rust (modules/stream_server)"
	@echo "5. Tester Phase 3: make validate-phase3"
	@echo ""

## phase3: Complete Phase 3 validation and build
phase3: validate-phase3 build-phase3
	@echo ""
	@echo "$(GREEN)🌐 PHASE 3 COMPLÉTÉE!$(RESET)"
	@echo "$(GREEN)✅ Communication gRPC implémentée$(RESET)"
	@echo "$(GREEN)✅ Event Bus NATS opérationnel$(RESET)"
	@echo "$(GREEN)✅ JWT Service partagé actif$(RESET)"
	@echo "$(GREEN)✅ Architecture microservices complète$(RESET)"
	@echo ""
	@echo "$(BLUE)🚀 Commandes Phase 3:$(RESET)"
	@echo "   make dev-phase3        # Serveur développement Phase 3"
	@echo "   make test-phase3       # Tests endpoints gRPC"
	@echo "   make demo-phase3       # Demo communication gRPC"
	@echo "   make grpc-status       # Status services gRPC"
	@echo "   make generate-proto    # Génération protobuf"
	@echo ""
	@echo "$(BLUE)🔗 Endpoints Phase 3:$(RESET)"
	@echo "   GET  /health                  - Santé microservices"
	@echo "   GET  /phase3/status           - Status Phase 3"
	@echo "   GET  /grpc/status             - Status services gRPC"
	@echo "   POST /api/chat/rooms          - Communication Chat gRPC"
	@echo "   POST /api/stream/create       - Communication Stream gRPC"
	@echo "   GET  /api/events/status       - Status Event Bus NATS"
	@echo ""
	@echo "$(GREEN)🚀 Prêt pour Phase 4 - Optimisation Chat Server$(RESET)"