#!/bin/bash

echo "🛠️ Configuration de l'environnement de développement"
echo "=================================================="

# Vérifications des prérequis
echo "1. Vérification des prérequis..."

# Go
if ! command -v go &> /dev/null; then
    echo "❌ Go n'est pas installé"
    echo "Installez Go depuis https://golang.org/"
    exit 1
fi

GO_VERSION=$(go version | cut -d' ' -f3)
echo "✅ Go installé: $GO_VERSION"

# PostgreSQL
if ! command -v psql &> /dev/null; then
    echo "⚠️ PostgreSQL client non trouvé"
    echo "Installez postgresql-client pour les tests DB"
fi

# Git
if ! command -v git &> /dev/null; then
    echo "❌ Git n'est pas installé"
    exit 1
fi

echo "✅ Git installé: $(git --version)"

# Configuration de l'environnement
echo ""
echo "2. Configuration de l'environnement..."

# Créer .env si nécessaire
if [ ! -f .env ]; then
    echo "📝 Création du fichier .env..."
    cat > .env << 'EOF'
# Configuration de développement
DATABASE_URL=postgresql://postgres:password@localhost:5432/talas_dev
JWT_SECRET=dev-secret-key-change-in-production
PORT=8080
ENVIRONMENT=development

# Base de données
DB_HOST=localhost
DB_PORT=5432
DB_USERNAME=postgres
DB_PASSWORD=password
DB_NAME=talas_dev
DB_SSLMODE=disable

# Fichiers
UPLOAD_DIR=./static
MAX_FILE_SIZE=104857600

# Modules Rust
CHAT_SERVER_PORT=9001
STREAM_SERVER_PORT=8082
EOF
    
    echo "✅ Fichier .env créé"
    echo "⚠️ Modifiez les valeurs selon votre configuration"
else
    echo "✅ Fichier .env existe déjà"
fi

# Répertoires nécessaires
echo ""
echo "3. Création des répertoires..."

DIRS=(
    "static/audio"
    "static/uploads"
    "static/avatars"
    "static/shared_resources"
    "static/internal_docs"
    "logs"
)

for dir in "${DIRS[@]}"; do
    if [ ! -d "$dir" ]; then
        mkdir -p "$dir"
        echo "📁 Créé: $dir"
    else
        echo "✅ Existe: $dir"
    fi
done

# Permissions des scripts
echo ""
echo "4. Configuration des permissions..."
chmod +x scripts/*.sh 2>/dev/null || true
echo "✅ Permissions des scripts configurées"

# Installation des dépendances Go
echo ""
echo "5. Installation des dépendances Go..."
go mod tidy
if [ $? -eq 0 ]; then
    echo "✅ Dépendances installées"
else
    echo "❌ Erreur lors de l'installation des dépendances"
    exit 1
fi

# Test de compilation
echo ""
echo "6. Test de compilation..."
if go build -o tmp_server ./cmd/server; then
    rm -f tmp_server
    echo "✅ Compilation réussie"
else
    echo "❌ Erreur de compilation"
    echo "Vérifiez les imports et la structure du code"
    exit 1
fi

# Configuration Git hooks (optionnel)
echo ""
echo "7. Configuration des Git hooks..."
if [ -d .git ]; then
    cat > .git/hooks/pre-commit << 'EOF'
#!/bin/bash
# Pre-commit hook: vérification du code

echo "🔍 Vérification avant commit..."

# Format du code
go fmt ./...

# Tests rapides
go vet ./...

# Vérification de compilation
if ! go build ./cmd/server >/dev/null 2>&1; then
    echo "❌ Erreur de compilation"
    exit 1
fi

echo "✅ Vérifications passées"
EOF
    
    chmod +x .git/hooks/pre-commit
    echo "✅ Pre-commit hook configuré"
fi

# Résumé
echo ""
echo "🎉 Configuration terminée !"
echo "========================="
echo ""
echo "📋 Résumé:"
echo "• Environnement: $(go version)"
echo "• Configuration: .env créé"
echo "• Répertoires: créés"
echo "• Dépendances: installées"
echo "• Compilation: OK"
echo ""
echo "🚀 Commandes utiles:"
echo "• Lancer le serveur: go run cmd/server/main.go"
echo "• Tests: bash scripts/test_endpoints.sh"
echo "• Validation: bash scripts/validate_build.sh"
echo ""
echo "📚 Documentation:"
echo "• Architecture: docs/migration/"
echo "• API: docs/api/"
echo "• Handlers: docs/doc_*_handler.md"