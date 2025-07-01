#!/bin/bash

# 🚀 SCRIPT DE TEST - JOURS 15-16 : Communication gRPC & Event Bus
# Valide l'intégration avec le backend Go et NATS

set -e

echo "🚀 ===== JOURS 15-16 - COMMUNICATION gRPC & EVENT BUS ====="
echo "📅 $(date '+%Y-%m-%d %H:%M:%S')"
echo ""

echo "🔧 Compilation avec nouveaux modules..."
cargo build --release

echo ""
echo "🧪 Tests d'intégration gRPC..."

# Test des modules gRPC
echo "🔌 Test 1/4: Service gRPC Stream"
cargo test --release --lib -- grpc --nocapture

echo ""
echo "📡 Test 2/4: Event Bus NATS"
cargo test --release --lib -- eventbus --nocapture

echo ""
echo "🔐 Test 3/4: Service d'authentification"
cargo test --release --lib -- auth_service --nocapture

echo ""
echo "⚡ Test 4/4: Communication end-to-end"
cargo test --release test_grpc_eventbus_integration -- --nocapture

echo ""
echo "📊 Génération du rapport Jours 15-16..."

cat > /tmp/jours15-16_rapport.md << EOF
# 📊 RAPPORT JOURS 15-16 - Communication gRPC & Event Bus

**Date**: $(date '+%Y-%m-%d %H:%M:%S')

## ✅ Réalisations

### 🔌 15.1 Intégration Backend Go
- **Service gRPC Stream**: ✅ Implémenté
  - Création/gestion des streams
  - Communication avec backend Go
  - Métriques et monitoring intégrés
  
- **Service d'Authentification**: ✅ Implémenté
  - Validation JWT tokens
  - Permissions utilisateurs
  - Intégration avec RBAC backend
  
- **Client gRPC**: ✅ Implémenté
  - Connexion au backend Go
  - Retry automatique avec backoff exponentiel
  - Métriques de performance

### 📡 15.2 Event Bus Partagé NATS
- **Event Bus Core**: ✅ Implémenté
  - Configuration NATS flexible
  - Publication/souscription d'événements
  - Monitoring des métriques
  
- **Types d'événements**: ✅ Définis
  - StreamStarted/StreamEnded
  - ListenerJoined/ListenerLeft
  - UserAuthenticated
  - AnalyticsEvent
  
- **Communication asynchrone**: ✅ Opérationnelle
  - Découplage services
  - Scalabilité horizontale
  - Résilience réseau

## 🎯 Architecture

**Stream Server Rust** ←→ **gRPC** ←→ **Backend Go**
                ↓
          **NATS Event Bus**
                ↓
     **Services distribués**

## 📈 Métriques Intégrées

- **gRPC**: Requêtes/sec, latence, erreurs
- **Event Bus**: Événements publiés/reçus, throughput
- **Connexions**: État, retry, health checks

## 🚀 Prochaines Étapes

**Jours 17-18**: Tests Production
- Load Testing 100k+ connexions
- Chaos Testing pour résilience
- Performance benchmarking
- Stress testing infrastructure

**Status**: ✅ JOURS 15-16 TERMINÉS AVEC SUCCÈS

EOF

echo "📄 Rapport généré dans /tmp/jours15-16_rapport.md"
cat /tmp/jours15-16_rapport.md

echo ""
echo "🎉 ===== JOURS 15-16 TERMINÉS AVEC SUCCÈS ====="
echo ""
echo "🚀 Prêt pour JOURS 17-18 : TESTS PRODUCTION"
echo "   📅 Load Testing 100k+ connexions"
echo "   📅 Chaos Testing pour résilience"
echo "   📅 Performance benchmarking"
echo "   📅 Stress testing infrastructure" 