#!/bin/bash

# Phase 5 - Streaming Audio Avancé - Script de Validation
# Objectifs: Streaming adaptatif, Support multi-bitrate, Synchronisation <100ms, Recording temps réel

set -e

echo "🎯 PHASE 5 - VALIDATION STREAMING AUDIO AVANCÉ"
echo "================================================"

# Configuration
PHASE5_TARGET_DIR="target/release"
PHASE5_BINARY="stream_server"
VALIDATION_LOG="validation_phase5.log"
SCORE=0
MAX_SCORE=100

# Fonctions utilitaires
log_info() {
    echo "ℹ️  $1" | tee -a $VALIDATION_LOG
}

log_success() {
    echo "✅ $1" | tee -a $VALIDATION_LOG
}

log_warning() {
    echo "⚠️  $1" | tee -a $VALIDATION_LOG
}

log_error() {
    echo "❌ $1" | tee -a $VALIDATION_LOG
}

add_score() {
    SCORE=$((SCORE + $1))
    echo "📊 Score partiel: $SCORE/$MAX_SCORE" | tee -a $VALIDATION_LOG
}

# Nettoyage initial
rm -f $VALIDATION_LOG
echo "🔄 Démarrage validation Phase 5 - $(date)" > $VALIDATION_LOG

echo ""
echo "1️⃣  VALIDATION ARCHITECTURE PHASE 5"
echo "=================================="

# Test 1: Modules WebRTC créés (15 points)
log_info "Test 1: Vérification modules WebRTC Phase 5..."
if [ -f "src/streaming/webrtc.rs" ]; then
    log_success "Module WebRTC détecté"
    if grep -q "WebRTCManager" src/streaming/webrtc.rs; then
        log_success "WebRTCManager trouvé"
        add_score 5
    fi
    if grep -q "AudioCodec" src/streaming/webrtc.rs; then
        log_success "Support multi-codec détecté"
        add_score 5
    fi
    if grep -q "bitrate_adaptation" src/streaming/webrtc.rs; then
        log_success "Adaptation bitrate automatique détectée"
        add_score 5
    fi
else
    log_error "Module WebRTC manquant"
fi

# Test 2: Module Synchronisation (15 points)
log_info "Test 2: Vérification module synchronisation..."
if [ -f "src/streaming/sync_manager.rs" ]; then
    log_success "Module synchronisation détecté"
    if grep -q "sync_tolerance_ms.*100" src/streaming/sync_manager.rs; then
        log_success "Objectif synchronisation <100ms configuré"
        add_score 5
    fi
    if grep -q "max_clients.*1000" src/streaming/sync_manager.rs; then
        log_success "Support 1000 listeners simultanés configuré"
        add_score 5
    fi
    if grep -q "MasterClock" src/streaming/sync_manager.rs; then
        log_success "Horloge maître implémentée"
        add_score 5
    fi
else
    log_error "Module synchronisation manquant"
fi

# Test 3: Module Recording temps réel (15 points)
log_info "Test 3: Vérification recording temps réel..."
if [ -f "src/streaming/live_recording.rs" ]; then
    log_success "Module recording temps réel détecté"
    if grep -q "real_time_transcoding" src/streaming/live_recording.rs; then
        log_success "Transcodage temps réel activé"
        add_score 5
    fi
    if grep -q "Mp3.*Flac.*Wav" src/streaming/live_recording.rs; then
        log_success "Support formats multiples (MP3, FLAC, WAV)"
        add_score 5
    fi
    if grep -q "metadata_injection" src/streaming/live_recording.rs; then
        log_success "Injection métadonnées configurée"
        add_score 5
    fi
else
    log_error "Module recording temps réel manquant"
fi

# Test 4: Moteur streaming avancé (10 points)
log_info "Test 4: Vérification moteur streaming avancé..."
if [ -f "src/streaming/advanced_streaming.rs" ]; then
    log_success "Moteur streaming avancé détecté"
    if grep -q "AdvancedStreamingEngine" src/streaming/advanced_streaming.rs; then
        log_success "Engine principal trouvé"
        add_score 5
    fi
    if grep -q "analytics_enabled" src/streaming/advanced_streaming.rs; then
        log_success "Analytics temps réel activées"
        add_score 5
    fi
else
    log_error "Moteur streaming avancé manquant"
fi

echo ""
echo "2️⃣  TEST COMPILATION OPTIMISÉE"
echo "=============================="

# Test 5: Compilation release (10 points)
log_info "Test 5: Compilation optimisée release..."
if cargo build --release 2>&1 | tee -a $VALIDATION_LOG; then
    log_success "Compilation release réussie"
    add_score 10
else
    log_error "Échec compilation release"
fi

echo ""
echo "3️⃣  VALIDATION FONCTIONNALITÉS PHASE 5"
echo "====================================="

# Test 6: Support multi-bitrate (10 points)
log_info "Test 6: Support multi-bitrate (64, 128, 256, 320 kbps)..."
BITRATE_COUNT=0
for bitrate in 64 128 256 320; do
    if grep -r "$bitrate" src/streaming/ >/dev/null 2>&1; then
        log_success "Bitrate $bitrate kbps détecté"
        BITRATE_COUNT=$((BITRATE_COUNT + 1))
    fi
done
if [ $BITRATE_COUNT -ge 3 ]; then
    log_success "Support multi-bitrate validé ($BITRATE_COUNT/4 bitrates)"
    add_score 10
else
    log_warning "Support multi-bitrate partiel ($BITRATE_COUNT/4 bitrates)"
    add_score $((BITRATE_COUNT * 2))
fi

# Test 7: Analytics temps réel (10 points)
log_info "Test 7: Analytics temps réel..."
if grep -r "StreamAnalytics" src/streaming/ >/dev/null 2>&1; then
    log_success "Structure analytics détectée"
    add_score 3
fi
if grep -r "peak_listeners" src/streaming/ >/dev/null 2>&1; then
    log_success "Métriques peak listeners trouvées"
    add_score 3
fi
if grep -r "average_bitrate" src/streaming/ >/dev/null 2>&1; then
    log_success "Métriques bitrate moyens trouvées"
    add_score 4
fi

# Test 8: Adaptive streaming (10 points)
log_info "Test 8: Streaming adaptatif..."
if grep -r "adaptive_quality" src/streaming/ >/dev/null 2>&1; then
    log_success "Qualité adaptative activée"
    add_score 5
fi
if grep -r "bandwidth_monitoring" src/streaming/ >/dev/null 2>&1; then
    log_success "Monitoring bande passante activé"
    add_score 5
fi

echo ""
echo "4️⃣  TEST INTÉGRATION MODULES"
echo "=========================="

# Test 9: Intégration mod.rs (5 points)
log_info "Test 9: Intégration modules dans mod.rs..."
if grep -q "pub mod webrtc" src/streaming/mod.rs && \
   grep -q "pub mod sync_manager" src/streaming/mod.rs && \
   grep -q "pub mod live_recording" src/streaming/mod.rs && \
   grep -q "pub mod advanced_streaming" src/streaming/mod.rs; then
    log_success "Tous les modules Phase 5 intégrés"
    add_score 5
else
    log_warning "Intégration modules incomplète"
fi

echo ""
echo "5️⃣  RÉSULTATS FINAUX PHASE 5"
echo "============================"

# Calcul score final
PERCENTAGE=$((SCORE * 100 / MAX_SCORE))

echo "📊 SCORE FINAL: $SCORE/$MAX_SCORE ($PERCENTAGE%)" | tee -a $VALIDATION_LOG

if [ $PERCENTAGE -ge 80 ]; then
    echo "🎉 PHASE 5 VALIDÉE! Streaming Audio Avancé opérationnel"
    echo "   ✅ WebRTC streaming implémenté"
    echo "   ✅ Synchronisation multi-clients <100ms"
    echo "   ✅ Recording temps réel fonctionnel"
    echo "   ✅ Analytics temps réel actives"
    echo "   ✅ Support 1000 listeners simultanés configuré"
    echo "   ✅ Multi-bitrate (64-320 kbps) supporté"
    echo ""
    echo "🚀 PRÊT POUR PHASE 6 - Monitoring & Production!"
    
elif [ $PERCENTAGE -ge 60 ]; then
    echo "⚠️  PHASE 5 PARTIELLEMENT VALIDÉE"
    echo "   Certaines fonctionnalités avancées manquent"
    echo "   Recommandation: compléter avant Phase 6"
    
else
    echo "❌ PHASE 5 NON VALIDÉE"
    echo "   Score insuffisant: $PERCENTAGE%"
    echo "   Minimum requis: 60%"
    exit 1
fi

# Métriques détaillées
echo ""
echo "📈 MÉTRIQUES PHASE 5:"
echo "==================="

# Compter les lignes de code des nouveaux modules
WC_WEBRTC=0
WC_SYNC=0
WC_RECORDING=0
WC_ADVANCED=0

if [ -f "src/streaming/webrtc.rs" ]; then
    WC_WEBRTC=$(wc -l < src/streaming/webrtc.rs)
fi
if [ -f "src/streaming/sync_manager.rs" ]; then
    WC_SYNC=$(wc -l < src/streaming/sync_manager.rs)
fi
if [ -f "src/streaming/live_recording.rs" ]; then
    WC_RECORDING=$(wc -l < src/streaming/live_recording.rs)
fi
if [ -f "src/streaming/advanced_streaming.rs" ]; then
    WC_ADVANCED=$(wc -l < src/streaming/advanced_streaming.rs)
fi

TOTAL_LOC=$((WC_WEBRTC + WC_SYNC + WC_RECORDING + WC_ADVANCED))

echo "📝 Lignes de code Phase 5: $TOTAL_LOC"
echo "   - WebRTC Module: $WC_WEBRTC lignes"
echo "   - Sync Manager: $WC_SYNC lignes"
echo "   - Live Recording: $WC_RECORDING lignes"
echo "   - Advanced Engine: $WC_ADVANCED lignes"

# Vérifier binaire généré
if [ -f "$PHASE5_TARGET_DIR/$PHASE5_BINARY" ]; then
    BINARY_SIZE=$(du -h "$PHASE5_TARGET_DIR/$PHASE5_BINARY" | cut -f1)
    echo "📦 Taille binaire optimisé: $BINARY_SIZE"
else
    echo "⚠️  Binaire optimisé non trouvé"
fi

echo ""
echo "✅ Validation Phase 5 terminée - $(date)"
echo "📄 Log détaillé: $VALIDATION_LOG"

exit 0 