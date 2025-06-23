#!/bin/bash

# Veza Web App - Quick Run Script
# This script helps you quickly set up and run the Veza backend

set -e

# Colors for output
BLUE='\033[36m'
GREEN='\033[32m'
RED='\033[31m'
YELLOW='\033[33m'
RESET='\033[0m'

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${RESET} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${RESET} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${RESET} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${RESET} $1"
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to check prerequisites
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    # Check Go
    if ! command_exists go; then
        print_error "Go is not installed. Please install Go 1.21 or later."
        exit 1
    fi
    
    # Check Go version
    GO_VERSION=$(go version | grep -oE 'go[0-9]+\.[0-9]+' | sed 's/go//')
    REQUIRED_VERSION="1.21"
    if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then
        print_error "Go version $REQUIRED_VERSION or later is required. Found: $GO_VERSION"
        exit 1
    fi
    
    # Check PostgreSQL
    if ! command_exists psql; then
        print_warning "PostgreSQL client not found. Make sure PostgreSQL is installed and running."
    fi
    
    # Check Rust (optional)
    if ! command_exists rustc; then
        print_warning "Rust not found. Chat and stream servers won't be available."
    fi
    
    print_success "Prerequisites check completed"
}

# Function to setup environment
setup_environment() {
    print_status "Setting up environment..."
    
    # Create .env if it doesn't exist
    if [ ! -f .env ]; then
        if [ -f .env.example ]; then
            cp .env.example .env
            print_success "Created .env file from .env.example"
            print_warning "Please edit .env file with your actual configuration"
        else
            print_error ".env.example not found. Please create .env file manually."
            exit 1
        fi
    else
        print_status ".env file already exists"
    fi
    
    # Create necessary directories
    mkdir -p logs
    mkdir -p static/uploads
    mkdir -p static/shared
    mkdir -p static/shared_ressources
    mkdir -p bin
    
    print_success "Environment setup completed"
}

# Function to install dependencies
install_dependencies() {
    print_status "Installing Go dependencies..."
    
    if ! go mod download; then
        print_error "Failed to download Go dependencies"
        exit 1
    fi
    
    if ! go mod tidy; then
        print_error "Failed to tidy Go modules"
        exit 1
    fi
    
    print_success "Go dependencies installed"
}

# Function to run database migrations
run_migrations() {
    print_status "Running database migrations..."
    
    # Check if database is accessible
    if ! go run -tags migrate internal/database/migrate.go 2>/dev/null; then
        print_warning "Database migration failed. Please ensure:"
        print_warning "1. PostgreSQL is running"
        print_warning "2. Database exists and is accessible"
        print_warning "3. DATABASE_URL in .env is correct"
        print_warning "Continuing without migrations..."
    else
        print_success "Database migrations completed"
    fi
}

# Function to build the application
build_application() {
    print_status "Building application..."
    
    if ! go build -o bin/veza-api main.go; then
        print_error "Failed to build application"
        exit 1
    fi
    
    print_success "Application built successfully"
}

# Function to run the application
run_application() {
    print_status "Starting Veza Web App..."
    print_status "Server will be available at: http://localhost:$(grep PORT .env | cut -d'=' -f2 | tr -d ' ' || echo '8080')"
    print_status "Press Ctrl+C to stop the server"
    echo ""
    
    if [ "$1" = "dev" ]; then
        # Development mode with hot reload
        if command_exists air; then
            air
        else
            print_warning "Air not found. Running without hot reload."
            go run main.go
        fi
    else
        # Production mode
        ./bin/veza-api
    fi
}

# Function to build Rust modules
build_rust_modules() {
    if command_exists cargo; then
        print_status "Building Rust modules..."
        
        # Build chat server
        if [ -d "modules/chat_server" ]; then
            print_status "Building chat server..."
            (cd modules/chat_server && cargo build --release) || print_warning "Failed to build chat server"
        fi
        
        # Build stream server
        if [ -d "modules/stream_server" ]; then
            print_status "Building stream server..."
            (cd modules/stream_server && cargo build --release) || print_warning "Failed to build stream server"
        fi
        
        print_success "Rust modules built"
    else
        print_warning "Cargo not found. Skipping Rust modules."
    fi
}

# Function to show help
show_help() {
    echo "Veza Web App - Quick Run Script"
    echo ""
    echo "Usage: $0 [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  setup     - Complete project setup (run once)"
    echo "  run       - Build and run the application"
    echo "  dev       - Run in development mode with hot reload"
    echo "  build     - Build the application only"
    echo "  clean     - Clean build artifacts"
    echo "  help      - Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 setup     # First time setup"
    echo "  $0 dev       # Development with hot reload"
    echo "  $0 run       # Production build and run"
}

# Function to clean build artifacts
clean_build() {
    print_status "Cleaning build artifacts..."
    
    rm -rf bin/
    rm -f coverage.out coverage.html
    go clean
    
    if command_exists cargo; then
        [ -d "modules/chat_server" ] && (cd modules/chat_server && cargo clean)
        [ -d "modules/stream_server" ] && (cd modules/stream_server && cargo clean)
    fi
    
    print_success "Clean completed"
}

# Main script logic
main() {
    case "${1:-run}" in
        "setup")
            check_prerequisites
            setup_environment
            install_dependencies
            build_rust_modules
            run_migrations
            build_application
            print_success "Setup completed! Run '$0 dev' to start development server."
            ;;
        "run")
            check_prerequisites
            install_dependencies
            build_application
            run_application
            ;;
        "dev")
            check_prerequisites
            install_dependencies
            run_application dev
            ;;
        "build")
            check_prerequisites
            install_dependencies
            build_application
            build_rust_modules
            ;;
        "clean")
            clean_build
            ;;
        "help"|"-h"|"--help")
            show_help
            ;;
        *)
            print_error "Unknown command: $1"
            show_help
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"