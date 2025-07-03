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

echo "🎯 ÉTAT INITIAL"
echo "==============="
chat_initial=$(count_issues "$CHAT_SERVER")
stream_initial=$(count_issues "$STREAM_SERVER")
echo "Chat Server: $chat_initial warnings"
echo "Stream Server: $stream_initial warnings"
echo "TOTAL: $((chat_initial + stream_initial)) problèmes à corriger"

# CORRECTION CHAT SERVER (2 warnings)
echo ""
echo "🔧 CORRECTION CHAT SERVER"
echo "========================="
cd "$CHAT_SERVER"

# Fix unused variable limit_type
sed -i 's/limit_type/_limit_type/g' src/core/advanced_rate_limiter.rs

# Fix unused import SplitSink
sed -i 's/use futures_util::{stream::SplitSink, StreamExt, SinkExt};/use futures_util::{StreamExt, SinkExt};/' src/main.rs

chat_after=$(count_issues "$CHAT_SERVER")
echo "Chat Server après correction: $chat_after warnings"

# CORRECTION STREAM SERVER (140 warnings)
echo ""
echo "🔧 CORRECTION MASSIVE STREAM SERVER"
echo "==================================="
cd "$STREAM_SERVER"

# Corrections spécifiques détectées
echo "Correction imports MP3..."
sed -i 's/use crate::codecs::{AudioEncoder, AudioDecoder, EncoderConfig, DecodedAudio};/use crate::codecs::{AudioEncoder, AudioDecoder};/' src/codecs/mp3.rs

echo "Correction imports playback..."
sed -i 's/use crate::core::{StreamManager, StreamEvent};/use crate::core::StreamManager;/' src/soundcloud/playback.rs

echo "Correction imports social..."
sed -i 's/use std::time::{Duration, SystemTime, Instant};/use std::time::{Duration, SystemTime};/' src/soundcloud/social.rs
sed -i 's/use tokio::sync::{RwLock, broadcast, mpsc};/use tokio::sync::{RwLock, broadcast};/' src/soundcloud/social.rs
sed -i '/use parking_lot::Mutex;/d' src/soundcloud/social.rs
sed -i 's/use tracing::{info, debug, warn};/use tracing::{info, debug};/' src/soundcloud/social.rs

echo "Correction imports discovery..."
sed -i 's/use std::collections::{HashMap, BTreeMap, VecDeque};/use std::collections::{HashMap, VecDeque};/' src/soundcloud/discovery.rs
sed -i 's/use std::time::{Duration, SystemTime, Instant};/use std::time::{Duration, SystemTime};/' src/soundcloud/discovery.rs

stream_after=$(count_issues "$STREAM_SERVER")
echo "Stream Server après première correction: $stream_after warnings"

echo ""
echo "🎯 VÉRIFICATION FINALE"
echo "======================"
echo "Chat Server: $(count_issues "$CHAT_SERVER") warnings"
echo "Stream Server: $(count_issues "$STREAM_SERVER") warnings"
echo "TOTAL: $(($(count_issues "$CHAT_SERVER") + $(count_issues "$STREAM_SERVER"))) problèmes"

