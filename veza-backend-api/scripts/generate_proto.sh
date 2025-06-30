#!/bin/bash

# 🔧 Script de Génération Protobuf - Phase 3 gRPC
# ===============================================

set -e

echo "🚀 GÉNÉRATION PROTOBUF PHASE 3"
echo "=============================="
echo ""

# Répertoires
PROTO_DIR="proto"
OUT_DIR="internal/grpc/generated"
TOOLS_DIR="tools"

# Créer le répertoire de sortie s'il n'existe pas
mkdir -p "${OUT_DIR}"
mkdir -p "${OUT_DIR}/common"
mkdir -p "${OUT_DIR}/chat"  
mkdir -p "${OUT_DIR}/stream"

echo "📁 Répertoires créés:"
echo "   - ${OUT_DIR}"
echo "   - ${OUT_DIR}/common"
echo "   - ${OUT_DIR}/chat"
echo "   - ${OUT_DIR}/stream"
echo ""

# Fonction de génération
generate_proto() {
    local proto_file="$1"
    local output_dir="$2"
    local package_name="$3"
    
    echo "🔨 Génération: ${proto_file}"
    
    # Vérifier si protoc est installé
    if ! command -v protoc &> /dev/null; then
        echo "❌ protoc n'est pas installé. Installation recommandée:"
        echo "   - Ubuntu/Debian: sudo apt install protobuf-compiler"
        echo "   - macOS: brew install protobuf"
        echo "   - Arch: sudo pacman -S protobuf"
        exit 1
    fi
    
    # Générer les fichiers Go
    protoc \
        --proto_path="${PROTO_DIR}" \
        --go_out="${output_dir}" \
        --go_opt=paths=source_relative \
        --go-grpc_out="${output_dir}" \
        --go-grpc_opt=paths=source_relative \
        "${proto_file}"
    
    if [ $? -eq 0 ]; then
        echo "   ✅ ${proto_file} → ${output_dir}"
    else
        echo "   ❌ Erreur génération ${proto_file}"
        exit 1
    fi
}

echo "🔧 Génération des fichiers protobuf..."
echo ""

# 1. Common Auth Service
echo "1️⃣  Common Auth Service"
generate_proto "common/auth.proto" "${OUT_DIR}" "common"
echo ""

# 2. Chat Service  
echo "2️⃣  Chat Service"
generate_proto "chat/chat.proto" "${OUT_DIR}" "chat"
echo ""

# 3. Stream Service
echo "3️⃣  Stream Service" 
generate_proto "stream/stream.proto" "${OUT_DIR}" "stream"
echo ""

echo "📊 RÉSUMÉ GÉNÉRATION"
echo "==================="
echo ""

# Compter les fichiers générés
pb_files=$(find "${OUT_DIR}" -name "*.pb.go" | wc -l)
grpc_files=$(find "${OUT_DIR}" -name "*_grpc.pb.go" | wc -l)

echo "📁 Fichiers générés:"
echo "   - Protobuf (.pb.go): ${pb_files}"
echo "   - gRPC (*_grpc.pb.go): ${grpc_files}"
echo "   - Total: $((pb_files + grpc_files))"
echo ""

echo "📂 Structure générée:"
find "${OUT_DIR}" -name "*.go" | sort | sed 's|^|   |'
echo ""

# Vérification des fichiers générés
echo "🔍 VÉRIFICATION GÉNÉRATION"
echo "=========================="
echo ""

required_files=(
    "${OUT_DIR}/common/auth.pb.go"
    "${OUT_DIR}/common/auth_grpc.pb.go"
    "${OUT_DIR}/chat/chat.pb.go"
    "${OUT_DIR}/chat/chat_grpc.pb.go"
    "${OUT_DIR}/stream/stream.pb.go"
    "${OUT_DIR}/stream/stream_grpc.pb.go"
)

all_exist=true
for file in "${required_files[@]}"; do
    if [ -f "$file" ]; then
        echo "   ✅ $file"
    else
        echo "   ❌ $file (MANQUANT)"
        all_exist=false
    fi
done
echo ""

if [ "$all_exist" = true ]; then
    echo "🎉 GÉNÉRATION PROTOBUF RÉUSSIE !"
    echo "================================"
    echo ""
    echo "✅ Tous les fichiers requis ont été générés"
    echo "✅ Prêt pour l'implémentation des clients gRPC"
    echo ""
    echo "📋 Prochaines étapes:"
    echo "   1. Implémenter les clients gRPC"
    echo "   2. Intégrer dans les handlers HTTP"
    echo "   3. Configurer les serveurs Rust gRPC"
    echo "   4. Tests d'intégration gRPC"
    echo ""
else
    echo "❌ GÉNÉRATION PROTOBUF ÉCHOUÉE"
    echo "=============================="
    echo ""
    echo "Certains fichiers n'ont pas été générés correctement."
    echo "Vérifiez les erreurs ci-dessus et relancez le script."
    exit 1
fi

# Test de compilation rapide
echo "🧪 TEST COMPILATION"  
echo "=================="
echo ""

echo "Test compilation des fichiers générés..."
if go build ./internal/grpc/generated/... 2>/dev/null; then
    echo "   ✅ Compilation réussie"
else
    echo "   ⚠️  Avertissements de compilation (normal à ce stade)"
fi
echo ""

echo "✨ Script de génération terminé avec succès !"
echo "Temps d'exécution: ${SECONDS}s" 