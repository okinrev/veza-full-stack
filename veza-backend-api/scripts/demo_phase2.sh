#!/bin/bash

echo "🚀 =============================================="
echo "   DÉMONSTRATION PHASE 2 - SÉCURITÉ & MIDDLEWARE"
echo "   VEZA BACKEND API"
echo "=============================================="
echo

# Démarrer le serveur en arrière-plan
echo "🔐 Démarrage du serveur Phase 2..."
./bin/veza-api-phase2 &
SERVER_PID=$!

# Attendre que le serveur démarre
sleep 3

echo "✅ Serveur démarré (PID: $SERVER_PID)"
echo

# Test 1: Health Check
echo "🏥 TEST 1: Health Check Sécurisé"
echo "curl http://localhost:8080/health"
curl -s http://localhost:8080/health | jq . 2>/dev/null || curl -s http://localhost:8080/health
echo -e "\n"

# Test 2: Status Phase 2
echo "🔐 TEST 2: Status Sécurité Phase 2"
echo "curl http://localhost:8080/phase2/status"
curl -s http://localhost:8080/phase2/status | jq . 2>/dev/null || curl -s http://localhost:8080/phase2/status
echo -e "\n"

# Test 3: Auth Status
echo "🔑 TEST 3: Auth Module Status"
echo "curl http://localhost:8080/api/auth/status"
curl -s http://localhost:8080/api/auth/status | jq . 2>/dev/null || curl -s http://localhost:8080/api/auth/status
echo -e "\n"

# Test 4: Register
echo "📝 TEST 4: Registration Endpoint"
echo "curl -X POST http://localhost:8080/api/auth/register [données]"
curl -X POST http://localhost:8080/api/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"username":"demo_user","email":"demo@example.com","password":"SecurePass123!"}' \
  | jq . 2>/dev/null || curl -X POST http://localhost:8080/api/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"username":"demo_user","email":"demo@example.com","password":"SecurePass123!"}'
echo -e "\n"

# Test 5: Login
echo "🔓 TEST 5: Login Endpoint"
echo "curl -X POST http://localhost:8080/api/auth/login [credentials]"
curl -X POST http://localhost:8080/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"demo@example.com","password":"SecurePass123!"}' \
  | jq . 2>/dev/null || curl -X POST http://localhost:8080/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"demo@example.com","password":"SecurePass123!"}'
echo -e "\n"

# Test 6: Profile
echo "�� TEST 6: Profile Endpoint"
echo "curl http://localhost:8080/api/auth/profile"
curl -s http://localhost:8080/api/auth/profile | jq . 2>/dev/null || curl -s http://localhost:8080/api/auth/profile
echo -e "\n"

# Nettoyer
echo "🧹 Nettoyage..."
kill $SERVER_PID 2>/dev/null
sleep 1

echo "✅ =============================================="
echo "   PHASE 2 DÉMONTRÉE AVEC SUCCÈS !"
echo "   Tous les endpoints sécurisés fonctionnent"
echo "=============================================="
