-- Migration pour rendre la base de données compatible avec tous les composants
-- Garde la compatibilité avec le backend Go et le frontend React
-- Ajoute les fonctionnalités nécessaires pour le serveur de chat Rust

BEGIN;

-- ==========================================
-- 1. MISE À JOUR DE LA TABLE MESSAGES
-- ==========================================

-- Ajouter les colonnes nécessaires pour le serveur de chat Rust
-- Tout en gardant les colonnes existantes pour la compatibilité Go/React

-- Ajouter author_username pour éviter les JOIN constants (dénormalisation contrôlée)
ALTER TABLE messages 
ADD COLUMN IF NOT EXISTS author_username VARCHAR(255);

-- Ajouter recipient_username pour les DM (dénormalisation contrôlée)  
ALTER TABLE messages 
ADD COLUMN IF NOT EXISTS recipient_username VARCHAR(255);

-- Ajouter room_id numérique tout en gardant room texte
ALTER TABLE messages 
ADD COLUMN IF NOT EXISTS room_id INTEGER REFERENCES rooms(id);

-- Ajouter original_content pour l'historique des éditions
ALTER TABLE messages 
ADD COLUMN IF NOT EXISTS original_content TEXT;

-- Corriger le mapping de author_id (dans le dump c'est author_id, dans Go c'est from_user)
-- Ajouter un alias pour la compatibilité
ALTER TABLE messages 
ADD COLUMN IF NOT EXISTS from_user INTEGER;

-- Ajouter recipient_id comme alias de to_user pour la compatibilité Rust
ALTER TABLE messages 
ADD COLUMN IF NOT EXISTS recipient_id INTEGER;

-- ==========================================
-- 2. CRÉATION DE LA TABLE MESSAGE_READ_STATUS
-- ==========================================

CREATE TABLE IF NOT EXISTS message_read_status (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    message_id INTEGER NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    read_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, message_id)
);

CREATE INDEX IF NOT EXISTS idx_message_read_status_user ON message_read_status(user_id);
CREATE INDEX IF NOT EXISTS idx_message_read_status_message ON message_read_status(message_id);

-- ==========================================
-- 3. MISE À JOUR DE LA TABLE MESSAGE_MENTIONS
-- ==========================================

-- Ajouter user_id comme alias de mentioned_user_id pour la compatibilité Rust
ALTER TABLE message_mentions 
ADD COLUMN IF NOT EXISTS user_id INTEGER;

-- ==========================================
-- 4. TRIGGERS POUR MAINTENIR LA COHÉRENCE
-- ==========================================

-- Fonction pour synchroniser les colonnes de compatibilité
CREATE OR REPLACE FUNCTION sync_message_compatibility() RETURNS TRIGGER AS $$
BEGIN
    -- Synchroniser author_id/from_user
    IF NEW.author_id IS NOT NULL AND NEW.from_user IS NULL THEN
        NEW.from_user := NEW.author_id;
    END IF;
    IF NEW.from_user IS NOT NULL AND NEW.author_id IS NULL THEN
        NEW.author_id := NEW.from_user;
    END IF;
    
    -- Synchroniser to_user/recipient_id
    IF NEW.to_user IS NOT NULL AND NEW.recipient_id IS NULL THEN
        NEW.recipient_id := NEW.to_user;
    END IF;
    IF NEW.recipient_id IS NOT NULL AND NEW.to_user IS NULL THEN
        NEW.to_user := NEW.recipient_id;
    END IF;
    
    -- Remplir author_username automatiquement
    IF NEW.author_username IS NULL AND NEW.author_id IS NOT NULL THEN
        SELECT username INTO NEW.author_username 
        FROM users 
        WHERE id = NEW.author_id;
    END IF;
    
    -- Remplir recipient_username automatiquement
    IF NEW.recipient_username IS NULL AND NEW.recipient_id IS NOT NULL THEN
        SELECT username INTO NEW.recipient_username 
        FROM users 
        WHERE id = NEW.recipient_id;
    END IF;
    
    -- Associer room et room_id
    IF NEW.room IS NOT NULL AND NEW.room_id IS NULL THEN
        SELECT id INTO NEW.room_id 
        FROM rooms 
        WHERE name = NEW.room;
    END IF;
    IF NEW.room_id IS NOT NULL AND NEW.room IS NULL THEN
        SELECT name INTO NEW.room 
        FROM rooms 
        WHERE id = NEW.room_id;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Appliquer le trigger
DROP TRIGGER IF EXISTS sync_message_compatibility_trigger ON messages;
CREATE TRIGGER sync_message_compatibility_trigger
    BEFORE INSERT OR UPDATE ON messages
    FOR EACH ROW EXECUTE FUNCTION sync_message_compatibility();

-- Fonction pour synchroniser message_mentions
CREATE OR REPLACE FUNCTION sync_mention_compatibility() RETURNS TRIGGER AS $$
BEGIN
    -- Synchroniser mentioned_user_id/user_id
    IF NEW.mentioned_user_id IS NOT NULL AND NEW.user_id IS NULL THEN
        NEW.user_id := NEW.mentioned_user_id;
    END IF;
    IF NEW.user_id IS NOT NULL AND NEW.mentioned_user_id IS NULL THEN
        NEW.mentioned_user_id := NEW.user_id;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Appliquer le trigger
DROP TRIGGER IF EXISTS sync_mention_compatibility_trigger ON message_mentions;
CREATE TRIGGER sync_mention_compatibility_trigger
    BEFORE INSERT OR UPDATE ON message_mentions
    FOR EACH ROW EXECUTE FUNCTION sync_mention_compatibility();

-- ==========================================
-- 5. SYNCHRONISATION DES DONNÉES EXISTANTES
-- ==========================================

-- Remplir les nouvelles colonnes avec les données existantes

-- Synchroniser author_id et from_user
UPDATE messages SET from_user = author_id WHERE from_user IS NULL AND author_id IS NOT NULL;
UPDATE messages SET author_id = from_user WHERE author_id IS NULL AND from_user IS NOT NULL;

-- Synchroniser to_user et recipient_id  
UPDATE messages SET recipient_id = to_user WHERE recipient_id IS NULL AND to_user IS NOT NULL;
UPDATE messages SET to_user = recipient_id WHERE to_user IS NULL AND recipient_id IS NOT NULL;

-- Remplir author_username
UPDATE messages 
SET author_username = u.username 
FROM users u 
WHERE messages.author_id = u.id AND messages.author_username IS NULL;

-- Remplir recipient_username
UPDATE messages 
SET recipient_username = u.username 
FROM users u 
WHERE messages.recipient_id = u.id AND messages.recipient_username IS NULL;

-- Associer room_id pour les rooms existantes
UPDATE messages 
SET room_id = r.id 
FROM rooms r 
WHERE messages.room = r.name AND messages.room_id IS NULL;

-- Synchroniser message_mentions
UPDATE message_mentions 
SET user_id = mentioned_user_id 
WHERE user_id IS NULL AND mentioned_user_id IS NOT NULL;

-- ==========================================
-- 6. VUES DE COMPATIBILITÉ (OPTIONNEL)
-- ==========================================

-- Vue pour le backend Go (format attendu)
CREATE OR REPLACE VIEW messages_go_view AS
SELECT 
    id,
    from_user,
    to_user,
    room,
    content,
    created_at,
    message_type,
    reply_to_id,
    is_edited,
    edited_at,
    metadata,
    is_pinned,
    thread_count,
    status,
    uuid,
    conversation_id,
    parent_message_id,
    updated_at
FROM messages;

-- Vue pour le serveur de chat Rust (format attendu)
CREATE OR REPLACE VIEW messages_rust_view AS
SELECT 
    id,
    author_id,
    author_username,
    recipient_id,
    recipient_username,
    room_id,
    room,
    content,
    created_at,
    message_type,
    parent_message_id,
    thread_count,
    status,
    is_pinned,
    is_edited,
    original_content,
    updated_at
FROM messages;

-- ==========================================
-- 7. NOUVEAUX INDEX POUR LES PERFORMANCES
-- ==========================================

-- Index pour les nouvelles colonnes
CREATE INDEX IF NOT EXISTS idx_messages_author_username ON messages(author_username);
CREATE INDEX IF NOT EXISTS idx_messages_recipient_username ON messages(recipient_username);
CREATE INDEX IF NOT EXISTS idx_messages_room_id ON messages(room_id);
CREATE INDEX IF NOT EXISTS idx_messages_original_content ON messages(original_content) WHERE original_content IS NOT NULL;

COMMIT;

-- ==========================================
-- NOTES D'UTILISATION
-- ==========================================

/*
Ce script de migration permet :

1. **Compatibilité totale** avec le backend Go existant
   - Conserve toutes les colonnes utilisées par Go
   - Les triggers maintiennent la cohérence automatiquement

2. **Support complet** pour le serveur de chat Rust  
   - Ajoute toutes les colonnes attendues par Rust
   - Dénormalisation contrôlée pour les performances

3. **Flexibilité** pour le frontend React
   - Peut utiliser soit les colonnes Go soit les colonnes Rust
   - Les vues facilitent l'adoption progressive

4. **Migration transparente**
   - Les applications existantes continuent de fonctionner
   - Pas de changement de code requis côté Go/React
   - Le serveur Rust peut utiliser le nouveau schéma

UTILISATION :
- Le backend Go continue d'utiliser from_user, to_user, room
- Le serveur Rust utilise author_id, recipient_id, room_id  
- Les triggers synchronisent automatiquement les deux formats
- Les usernames sont dénormalisés pour éviter les JOIN constants
*/ 