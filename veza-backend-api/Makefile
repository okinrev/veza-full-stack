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