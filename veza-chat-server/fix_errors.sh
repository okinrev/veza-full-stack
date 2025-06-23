#!/bin/bash

echo "🔧 Correction automatique des erreurs de compilation..."

# 1. Corriger les erreurs de types dans message_store.rs
echo "📝 Correction des erreurs de types dans message_store.rs..."

# Corriger les conversions de types
sed -i 's/message\.id,/message.id as i32,/g' src/message_store.rs
sed -i 's/message\.room_id,/message.room_id.map(|r| r.parse::<i32>().unwrap_or(0)),/g' src/message_store.rs
sed -i 's/message\.created_at,/message.created_at.naive_utc(),/g' src/message_store.rs

# 2. Corriger les erreurs dans simple_message_store.rs
echo "📝 Correction des erreurs dans simple_message_store.rs..."
sed -i 's/ChatError::database_error("operation", &format!/ChatError::configuration_error("Database error")/g' src/simple_message_store.rs
sed -i 's/ChatError::validation_error/ChatError::configuration_error/g' src/simple_message_store.rs

# 3. Corriger les erreurs dans test_simple_store.rs
echo "📝 Correction des erreurs dans test_simple_store.rs..."
sed -i 's/ChatError::database_error("operation", &format!/ChatError::configuration_error("Database error")/g' src/test_simple_store.rs

# 4. Corriger les erreurs de signatures de fonctions
echo "📝 Correction des signatures de fonctions..."

# Corriger room_enhanced.rs
sed -i 's/limit,/limit.into(),/g' src/hub/room_enhanced.rs
sed -i 's/channels::send_room_message(hub, room_id, user_id, username, content, parent_id, None)/channels::send_room_message(hub, room_id, user_id, username, content, parent_id, None)/g' src/hub/room_enhanced.rs

# 5. Corriger message_handler.rs
echo "📝 Correction des erreurs dans message_handler.rs..."
sed -i 's/limit,/limit.into(),/g' src/message_handler.rs

# 6. Corriger les accès aux champs de row
echo "📝 Correction des accès aux champs de row..."
sed -i 's/row\.try_get(/row.get(/g' src/message_store.rs

echo "✅ Corrections appliquées. Teste maintenant la compilation..." 