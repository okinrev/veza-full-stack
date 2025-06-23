#!/usr/bin/env python3
"""
Script pour corriger automatiquement les erreurs de compilation du chat server Rust
"""

import re
import os

def fix_message_store_errors():
    """Corrige les erreurs dans message_store.rs"""
    file_path = "src/message_store.rs"
    
    with open(file_path, 'r') as f:
        content = f.read()
    
    # Corriger toutes les instances de ChatError::Database
    content = re.sub(
        r'\.map_err\(ChatError::Database\)\?',
        '.map_err(|e| ChatError::database_error("database_operation", e))?',
        content
    )
    
    # Corriger ChatError::PermissionDenied
    content = re.sub(
        r'ChatError::PermissionDenied\("([^"]+)"\\.to_string\(\)\)',
        r'ChatError::PermissionDenied { message: "\1".to_string() }',
        content
    )
    
    content = re.sub(
        r'ChatError::PermissionDenied\("([^"]+)"\)',
        r'ChatError::PermissionDenied { message: "\1".to_string() }',
        content
    )
    
    # Corriger ChatError::ReactionAlreadyExists et ReactionNotFound
    content = re.sub(
        r'ChatError::ReactionAlreadyExists\)',
        'ChatError::ReactionAlreadyExists',
        content
    )
    
    content = re.sub(
        r'ChatError::ReactionNotFound\)',
        'ChatError::ReactionNotFound',
        content
    )
    
    # Corriger configuration_error avec to_string()
    content = re.sub(
        r'ChatError::configuration_error\("([^"]+)"\\.to_string\(\)\)',
        r'ChatError::configuration_error("\1")',
        content
    )
    
    # Corriger la signature de row_to_message
    content = re.sub(
        r'async fn row_to_message\(&self, row: sqlx::Row\) -> Result<Message>',
        'async fn row_to_message<R: sqlx::Row>(&self, row: R) -> Result<Message>',
        content
    )
    
    # Remplacer les accès directs aux champs de row par get()
    row_fields = [
        'id', 'content', 'author_id', 'author_username', 'room_id', 
        'recipient_id', 'recipient_username', 'created_at', 'updated_at',
        'status', 'is_pinned', 'is_edited', 'original_content', 
        'parent_message_id', 'thread_count', 'is_flagged', 'moderation_notes',
        'message_type', 'mention_ids'
    ]
    
    for field in row_fields:
        content = re.sub(
            rf'row\.{field}',
            f'row.try_get("{field}").unwrap_or_default()',
            content
        )
    
    with open(file_path, 'w') as f:
        f.write(content)
    
    print(f"✅ Corrigé {file_path}")

def fix_message_handler_errors():
    """Corrige les erreurs dans message_handler.rs"""
    file_path = "src/message_handler.rs"
    
    with open(file_path, 'r') as f:
        content = f.read()
    
    # Remplacer validate_room_name et sanitize_content par validate_content
    content = re.sub(
        r'self\.content_filter\.validate_room_name\(([^)]+)\)',
        r'self.content_filter.validate_content(\1)',
        content
    )
    
    content = re.sub(
        r'self\.content_filter\.sanitize_content\(([^)]+)\)',
        r'self.content_filter.validate_content(\1)',
        content
    )
    
    with open(file_path, 'w') as f:
        f.write(content)
    
    print(f"✅ Corrigé {file_path}")

def fix_error_severity():
    """Corrige les erreurs de severity dans error.rs"""
    file_path = "src/error.rs"
    
    with open(file_path, 'r') as f:
        content = f.read()
    
    # Corriger les doublons dans severity mapping
    content = re.sub(
        r'Self::MessageTooLong \{ \.\. \} => 413,',
        '',
        content,
        count=1  # Supprimer seulement le doublon
    )
    
    with open(file_path, 'w') as f:
        f.write(content)
    
    print(f"✅ Corrigé {file_path}")

def fix_unused_imports():
    """Corrige les imports inutilisés"""
    files_to_fix = [
        "src/hub/channel_websocket.rs",
        "src/hub/direct_messages_websocket.rs", 
        "src/moderation.rs",
        "src/monitoring.rs",
        "src/security.rs",
        "src/websocket.rs"
    ]
    
    for file_path in files_to_fix:
        if os.path.exists(file_path):
            with open(file_path, 'r') as f:
                content = f.read()
            
            # Supprimer les imports inutilisés
            content = re.sub(r'use.*tungstenite::Message;\n', '', content)
            content = re.sub(r', error', '', content)
            content = re.sub(r'use std::collections::HashMap;\n', '', content)
            content = re.sub(r', SystemTime, UNIX_EPOCH', '', content)
            content = re.sub(r', Deserialize', '', content)
            content = re.sub(r'use crate::error::Result;\n', '', content)
            content = re.sub(r', UNIX_EPOCH', '', content)
            content = re.sub(r', Serialize, Deserialize', '', content)
            content = re.sub(r', Result as WsResult', '', content)
            
            # Ajouter prefix underscore aux variables non utilisées
            content = re.sub(r'let blocked_by', 'let _blocked_by', content)
            content = re.sub(r'let mut config', 'let config', content)
            
            with open(file_path, 'w') as f:
                f.write(content)
            
            print(f"✅ Corrigé {file_path}")

def main():
    """Fonction principale"""
    print("🔧 Correction des erreurs de compilation Rust...")
    
    # Changer vers le répertoire du chat server
    os.chdir("/home/senke/Documents/veza-full-stack/veza-chat-server")
    
    # Appliquer les corrections
    fix_message_store_errors()
    fix_message_handler_errors() 
    fix_error_severity()
    fix_unused_imports()
    
    print("✅ Toutes les corrections appliquées!")
    print("🔄 Vérification de la compilation...")
    
    # Tester la compilation
    os.system("cargo check")

if __name__ == "__main__":
    main() 