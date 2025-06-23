#!/bin/bash

# Script de test complet pour l'API Talas/Veza
# Teste toutes les fonctionnalités en profondeur
# Auteur: Assistant Claude pour Senke
# Date: 2025-06-23

set -e  # Arrêter en cas d'erreur

# Configuration
API_BASE="http://localhost:8080/api/v1"
TEMP_DIR="/tmp/veza_tests"
LOG_FILE="$TEMP_DIR/test_results.log"

# Couleurs pour l'affichage
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Variables globales pour les tests
ACCESS_TOKEN=""
REFRESH_TOKEN=""
USER_ID=""
USERNAME=""
EMAIL=""
TEST_USER2_TOKEN=""
TEST_USER2_ID=""
ROOM_ID=""
LISTING_ID=""
TRACK_ID=""

# Compteurs de tests
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Création du dossier temporaire
mkdir -p "$TEMP_DIR"
echo "📋 Démarrage des tests API Talas - $(date)" > "$LOG_FILE"

# Fonctions utilitaires
print_header() {
    echo -e "\n${BLUE}======================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}======================================${NC}\n"
}

print_test() {
    echo -e "${CYAN}🧪 Test: $1${NC}"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
    PASSED_TESTS=$((PASSED_TESTS + 1))
    echo "[✅] $1" >> "$LOG_FILE"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
    FAILED_TESTS=$((FAILED_TESTS + 1))
    echo "[❌] $1" >> "$LOG_FILE"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
    echo "[⚠️ ] $1" >> "$LOG_FILE"
}

print_info() {
    echo -e "${PURPLE}ℹ️  $1${NC}"
}

# Fonction pour faire une requête avec gestion d'erreur
make_request() {
    local method="$1"
    local url="$2"
    local data="$3"
    local headers="$4"
    local expected_status="$5"
    
    local curl_cmd="curl -s"
    
    if [ -n "$headers" ]; then
        curl_cmd="$curl_cmd $headers"
    fi
    
    curl_cmd="$curl_cmd -w '%{http_code}' -o $TEMP_DIR/response.json"
    
    if [ "$method" != "GET" ]; then
        curl_cmd="$curl_cmd -X $method"
    fi
    
    if [ -n "$data" ]; then
        curl_cmd="$curl_cmd -d '$data'"
    fi
    
    curl_cmd="$curl_cmd '$url'"
    
    local status_code=$(eval $curl_cmd)
    local response=$(cat "$TEMP_DIR/response.json" 2>/dev/null || echo "{}")
    
    echo "$response" > "$TEMP_DIR/last_response.json"
    echo "$status_code" > "$TEMP_DIR/last_status.txt"
    
    if [ -n "$expected_status" ] && [ "$status_code" != "$expected_status" ]; then
        return 1
    fi
    
    return 0
}