#!/bin/bash

set -e

echo "🚨 SCRIPT OBLIGATOIRE - ZÉRO ERREUR, ZÉRO WARNING 🚨"
echo "=================================================="

WORKSPACE_ROOT="/home/senke/Documents/veza-full-stack"
CHAT_SERVER="$WORKSPACE_ROOT/veza-chat-server"
STREAM_SERVER="$WORKSPACE_ROOT/veza-stream-server"

# Fonction pour compter les erreurs/warnings
count_issues() {
    local dir=$1
    cd "$dir"
    cargo check 2>&1 | grep -cE "error:|warning:" || echo "0"
}

# Fonction pour corriger les imports inutilisés
fix_unused_imports() {
    local dir=$1
    echo "🔧 Correction des imports inutilisés dans $dir"
    cd "$dir"
    
    # Obtenir la liste des imports inutilisés
    cargo check 2>&1 | grep -A 2 "unused import" | grep -E "^\s*[0-9]+\s*\|" | while read line; do
        echo "Import à corriger: $line"
    done
}

# Fonction pour corriger les variables inutilisées
fix_unused_variables() {
    local dir=$1
    echo "🔧 Correction des variables inutilisées dans $dir"
    cd "$dir"
    
    # Trouver tous les fichiers Rust
    find src -name "*.rs" | while read file; do
        # Préfixer les variables inutilisées avec _
        sed -i 's/\b\([a-zA-Z_][a-zA-Z0-9_]*\):/\_\1:/g' "$file" 2>/dev/null || true
    done
}

# ÉTAPE 1: Correction des 2 warnings restants du Chat Server
echo "🎯 ÉTAPE 1: Chat Server (2 warnings à corriger)"
echo "================================================"

cd "$CHAT_SERVER"
current_chat_warnings=$(count_issues "$CHAT_SERVER")
echo "Warnings actuels Chat Server: $current_chat_warnings"

# Correction spécifique du warning limit_type
echo "🔧 Correction warning limit_type..."
sed -i 's/limit_type/_limit_type/g' src/core/advanced_rate_limiter.rs

# Correction spécifique de l'import SplitSink
echo "🔧 Correction import SplitSink..."
sed -i 's/use futures_util::{stream::SplitSink, StreamExt, SinkExt};/use futures_util::{StreamExt, SinkExt};/' src/main.rs

# Vérification Chat Server
new_chat_warnings=$(count_issues "$CHAT_SERVER")
echo "Warnings Chat Server après correction: $new_chat_warnings"

# ÉTAPE 2: Correction massive du Stream Server (140 warnings)
echo ""
echo "🎯 ÉTAPE 2: Stream Server (140 warnings à corriger)"
echo "===================================================="

cd "$STREAM_SERVER"
current_stream_warnings=$(count_issues "$STREAM_SERVER")
echo "Warnings actuels Stream Server: $current_stream_warnings"

# Correction systématique des imports inutilisés
echo "🔧 Suppression imports inutilisés..."

# MP3 codec - suppression EncoderConfig et DecodedAudio
sed -i 's/use crate::codecs::{AudioEncoder, AudioDecoder, EncoderConfig, DecodedAudio};/use crate::codecs::{AudioEncoder, AudioDecoder};/' src/codecs/mp3.rs

# Playback - suppression StreamEvent
sed -i 's/use crate::core::{StreamManager, StreamEvent};/use crate::core::StreamManager;/' src/soundcloud/playback.rs

# Social - suppressions multiples
sed -i 's/use std::time::{Duration, SystemTime, Instant};/use std::time::{Duration, SystemTime};/' src/soundcloud/social.rs
sed -i 's/use tokio::sync::{RwLock, broadcast, mpsc};/use tokio::sync::{RwLock, broadcast};/' src/soundcloud/social.rs
sed -i '/use parking_lot::Mutex;/d' src/soundcloud/social.rs
sed -i 's/use tracing::{info, debug, warn};/use tracing::{info, debug};/' src/soundcloud/social.rs

# Discovery - suppressions
sed -i 's/use std::collections::{HashMap, BTreeMap, VecDeque};/use std::collections::{HashMap, VecDeque};/' src/soundcloud/discovery.rs
sed -i 's/use std::time::{Duration, SystemTime, Instant};/use std::time::{Duration, SystemTime};/' src/soundcloud/discovery.rs

echo "🔧 Correction systématique de tous les fichiers..."

# Correction automatique de tous les warnings d'imports
find src -name "*.rs" -exec grep -l "unused import" {} \; 2>/dev/null | while read file; do
    echo "Correction import dans: $file"
    # Suppression automatique des imports inutilisés courants
    sed -i '/use std::sync::mpsc;/d' "$file" 2>/dev/null || true
    sed -i '/use tokio::time::Instant;/d' "$file" 2>/dev/null || true
    sed -i '/use std::time::Instant;/d' "$file" 2>/dev/null || true
    sed -i '/use parking_lot::Mutex;/d' "$file" 2>/dev/null || true
    sed -i '/use tracing::warn;/d' "$file" 2>/dev/null || true
    sed -i 's/, warn//g' "$file" 2>/dev/null || true
    sed -i 's/warn, //g' "$file" 2>/dev/null || true
    sed -i 's/, Instant//g' "$file" 2>/dev/null || true
    sed -i 's/Instant, //g' "$file" 2>/dev/null || true
    sed -i 's/, BTreeMap//g' "$file" 2>/dev/null || true
    sed -i 's/BTreeMap, //g' "$file" 2>/dev/null || true
    sed -i 's/, mpsc//g' "$file" 2>/dev/null || true
    sed -i 's/mpsc, //g' "$file" 2>/dev/null || true
    sed -i 's/, StreamEvent//g' "$file" 2>/dev/null || true
    sed -i 's/StreamEvent, //g' "$file" 2>/dev/null || true
    sed -i 's/, EncoderConfig//g' "$file" 2>/dev/null || true
    sed -i 's/EncoderConfig, //g' "$file" 2>/dev/null || true
    sed -i 's/, DecodedAudio//g' "$file" 2>/dev/null || true
    sed -i 's/DecodedAudio, //g' "$file" 2>/dev/null || true
done

# BOUCLE DE CORRECTION CONTINUE
echo ""
echo "🔄 BOUCLE DE CORRECTION CONTINUE"
echo "================================="

iteration=1
while true; do
    echo "🔄 Itération $iteration"
    
    # Vérification Chat Server
    cd "$CHAT_SERVER"
    chat_warnings=$(count_issues "$CHAT_SERVER")
    
    # Vérification Stream Server  
    cd "$STREAM_SERVER"
    stream_warnings=$(count_issues "$STREAM_SERVER")
    
    echo "Chat Server: $chat_warnings warnings"
    echo "Stream Server: $stream_warnings warnings"
    
    if [ "$chat_warnings" -eq 0 ] && [ "$stream_warnings" -eq 0 ]; then
        echo ""
        echo "🎉 SUCCÈS ! ZÉRO ERREUR, ZÉRO WARNING !"
        echo "========================================"
        break
    fi
    
    if [ "$chat_warnings" -gt 0 ]; then
        echo "🔧 Correction Chat Server..."
        cd "$CHAT_SERVER"
        
        # Correction automatique avec cargo fix
        cargo fix --allow-dirty --allow-staged 2>/dev/null || true
        
        # Correction manuelle des warnings persistants
        cargo check 2>&1 | grep -A 1 "warning:" | grep "unused" | while read line; do
            if [[ "$line" =~ "unused variable:" ]]; then
                var_name=$(echo "$line" | grep -o '`[^`]*`' | sed 's/`//g')
                if [ ! -z "$var_name" ]; then
                    find src -name "*.rs" -exec sed -i "s/\\b$var_name\\b/_$var_name/g" {} \; 2>/dev/null || true
                fi
            fi
        done
    fi
    
    if [ "$stream_warnings" -gt 0 ]; then
        echo "🔧 Correction Stream Server..."
        cd "$STREAM_SERVER"
        
        # Correction automatique avec cargo fix
        cargo fix --allow-dirty --allow-staged 2>/dev/null || true
        
        # Analyse et correction des warnings spécifiques
        cargo check 2>&1 | grep -A 3 "warning:" | while read line; do
            if [[ "$line" =~ "unused import:" ]]; then
                # Extraction et suppression de l'import
                import_name=$(echo "$line" | grep -o '`[^`]*`' | sed 's/`//g')
                if [ ! -z "$import_name" ]; then
                    find src -name "*.rs" -exec sed -i "s/, $import_name//g; s/$import_name, //g; /use.*$import_name.*;/d" {} \; 2>/dev/null || true
                fi
            elif [[ "$line" =~ "unused variable:" ]]; then
                # Préfixage des variables inutilisées
                var_name=$(echo "$line" | grep -o '`[^`]*`' | sed 's/`//g')
                if [ ! -z "$var_name" ]; then
                    find src -name "*.rs" -exec sed -i "s/\\b$var_name\\b/_$var_name/g" {} \; 2>/dev/null || true
                fi
            fi
        done
    fi
    
    iteration=$((iteration + 1))
    
    # Sécurité : maximum 50 itérations
    if [ "$iteration" -gt 50 ]; then
        echo "⚠️  Limite d'itérations atteinte"
        break
    fi
done

# VALIDATION FINALE OBLIGATOIRE
echo ""
echo "🎯 VALIDATION FINALE OBLIGATOIRE"
echo "================================="

echo "=== CHAT SERVER ==="
cd "$CHAT_SERVER"
cargo clean
chat_errors=$(cargo check --all-features 2>&1 | grep -c "error:" || echo "0")
chat_warnings=$(cargo check --all-features 2>&1 | grep -c "warning:" || echo "0")
echo "Erreurs: $chat_errors (DOIT ÊTRE 0)"
echo "Warnings: $chat_warnings (DOIT ÊTRE 0)"

echo ""
echo "=== STREAM SERVER ==="
cd "$STREAM_SERVER"
cargo clean
stream_errors=$(cargo check --all-features 2>&1 | grep -c "error:" || echo "0")
stream_warnings=$(cargo check --all-features 2>&1 | grep -c "warning:" || echo "0")
echo "Erreurs: $stream_errors (DOIT ÊTRE 0)"
echo "Warnings: $stream_warnings (DOIT ÊTRE 0)"

echo ""
echo "🎯 RÉSULTAT FINAL"
echo "=================="
echo "Chat Server:"
echo "  Erreurs: $chat_errors"
echo "  Warnings: $chat_warnings"
echo ""
echo "Stream Server:"
echo "  Erreurs: $stream_errors"
echo "  Warnings: $stream_warnings"

total_issues=$((chat_errors + chat_warnings + stream_errors + stream_warnings))

if [ "$total_issues" -eq 0 ]; then
    echo ""
    echo "✅ MISSION ACCOMPLIE ! ZÉRO ERREUR, ZÉRO WARNING !"
    echo "=================================================="
    exit 0
else
    echo ""
    echo "❌ ÉCHEC ! IL RESTE $total_issues PROBLÈMES !"
    echo "=============================================="
    exit 1
fi 