#!/bin/bash
set -e

echo "🧪 Testing backend migration..."

echo "1. Testing Go module..."
go mod tidy
echo "✅ Go module OK"

echo "2. Testing compilation with old main.go..."
if go build -o /tmp/test-old main.go; then
    echo "✅ Old structure compiles"
    rm -f /tmp/test-old
else
    echo "❌ Old structure compilation failed"
    exit 1
fi

echo "3. Testing new structure imports..."
go list ./internal/... > /dev/null
echo "✅ New structure imports OK"

echo "4. Testing handlers import..."
go build -c internal/admin/handlers/products.go > /dev/null
echo "✅ New handlers compile"

echo ""
echo "🎉 Migration test successful!"
echo "You can now:"
echo "  - make dev     # Run with old main.go"
echo "  - make test    # Run tests"
echo "  - git status   # See what's changed"
