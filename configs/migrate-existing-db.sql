-- Script de migration pour adapter la base de données existante à l'architecture unifiée Veza
-- Ce script ajoute uniquement les éléments manquants sans toucher aux données existantes

-- =============================================================================
-- Extensions et Types (si pas encore présents)
-- =============================================================================

-- Extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- Ajout des colonnes manquantes dans les tables existantes (si nécessaire)
-- =============================================================================

-- Ajouter UUID aux tables existantes si manquant
DO $$ 
BEGIN
    -- Ajouter UUID à la table users si elle n'en a pas
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'uuid') THEN
        ALTER TABLE users ADD COLUMN uuid UUID DEFAULT uuid_generate_v4() UNIQUE;
        UPDATE users SET uuid = uuid_generate_v4() WHERE uuid IS NULL;
    END IF;
    
    -- Ajouter UUID à la table conversations si elle n'en a pas
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'conversations' AND column_name = 'uuid') THEN
        ALTER TABLE conversations ADD COLUMN uuid UUID DEFAULT uuid_generate_v4() UNIQUE;
        UPDATE conversations SET uuid = uuid_generate_v4() WHERE uuid IS NULL;
    END IF;
    
    -- Ajouter UUID à la table messages si elle n'en a pas
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'messages' AND column_name = 'uuid') THEN
        ALTER TABLE messages ADD COLUMN uuid UUID DEFAULT uuid_generate_v4() UNIQUE;
        UPDATE messages SET uuid = uuid_generate_v4() WHERE uuid IS NULL;
    END IF;
    
    -- Ajouter UUID à la table rooms si elle n'en a pas
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'rooms' AND column_name = 'uuid') THEN
        ALTER TABLE rooms ADD COLUMN uuid UUID DEFAULT uuid_generate_v4() UNIQUE;
        UPDATE rooms SET uuid = uuid_generate_v4() WHERE uuid IS NULL;
    END IF;
END $$;

-- =============================================================================
-- Tables pour le streaming audio (nouvelles fonctionnalités)
-- =============================================================================

-- Table des pistes audio (nouvelle fonctionnalité)
CREATE TABLE IF NOT EXISTS "public"."audio_tracks" (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID DEFAULT uuid_generate_v4() UNIQUE NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    artist VARCHAR(255),
    album VARCHAR(255),
    duration_seconds INTEGER,
    file_size BIGINT,
    file_path TEXT NOT NULL,
    original_filename VARCHAR(255),
    mime_type VARCHAR(100),
    sample_rate INTEGER,
    bit_rate INTEGER,
    channels INTEGER,
    waveform_data JSONB,
    thumbnail_url TEXT,
    uploaded_by BIGINT REFERENCES users(id) ON DELETE CASCADE,
    is_public BOOLEAN DEFAULT true,
    play_count INTEGER DEFAULT 0,
    download_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Table des playlists
CREATE TABLE IF NOT EXISTS "public"."playlists" (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID DEFAULT uuid_generate_v4() UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    owner_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    is_public BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Table des éléments de playlist
CREATE TABLE IF NOT EXISTS "public"."playlist_items" (
    id BIGSERIAL PRIMARY KEY,
    playlist_id BIGINT REFERENCES playlists(id) ON DELETE CASCADE,
    track_id BIGINT REFERENCES audio_tracks(id) ON DELETE CASCADE,
    position INTEGER NOT NULL,
    added_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(playlist_id, track_id)
);

-- Table des statistiques d'écoute
CREATE TABLE IF NOT EXISTS "public"."listening_stats" (
    id BIGSERIAL PRIMARY KEY,
    track_id BIGINT REFERENCES audio_tracks(id) ON DELETE CASCADE,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    session_id UUID DEFAULT uuid_generate_v4(),
    duration_listened INTEGER, -- en secondes
    completed BOOLEAN DEFAULT false,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- =============================================================================
-- Tables de notifications étendues (si pas déjà complètes)
-- =============================================================================

-- Améliorer la table notifications si nécessaire
DO $$
BEGIN
    -- Ajouter UUID si manquant
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'notifications' AND column_name = 'uuid') THEN
        ALTER TABLE notifications ADD COLUMN uuid UUID DEFAULT uuid_generate_v4() UNIQUE;
        UPDATE notifications SET uuid = uuid_generate_v4() WHERE uuid IS NULL;
    END IF;
    
    -- Ajouter metadata si manquant
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'notifications' AND column_name = 'metadata') THEN
        ALTER TABLE notifications ADD COLUMN metadata JSONB DEFAULT '{}';
    END IF;
END $$;

-- =============================================================================
-- Tables de session étendues pour l'architecture unifiée
-- =============================================================================

-- Sessions utilisateur étendues (si la table user_sessions est différente)
CREATE TABLE IF NOT EXISTS "public"."user_sessions_extended" (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    session_token VARCHAR(255) UNIQUE NOT NULL,
    refresh_token VARCHAR(255) UNIQUE NOT NULL,
    device_info JSONB DEFAULT '{}',
    ip_address INET,
    user_agent TEXT,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_activity_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- =============================================================================
-- Indexes pour optimiser les performances (si pas déjà présents)
-- =============================================================================

-- Indexes sur les UUIDs
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_uuid ON users(uuid);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_conversations_uuid ON conversations(uuid);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_messages_uuid ON messages(uuid);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_rooms_uuid ON rooms(uuid);

-- Indexes sur les messages pour les nouvelles fonctionnalités
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_messages_conversation_id ON messages(conversation_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_messages_room_id ON messages(room_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_messages_author_id ON messages(author_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_messages_created_at ON messages(created_at);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_messages_status ON messages(status);

-- Indexes pour les audio tracks
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audio_tracks_uploaded_by ON audio_tracks(uploaded_by);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audio_tracks_created_at ON audio_tracks(created_at);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audio_tracks_is_public ON audio_tracks(is_public);

-- Indexes pour les statistiques d'écoute
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_listening_stats_track_id ON listening_stats(track_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_listening_stats_user_id ON listening_stats(user_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_listening_stats_created_at ON listening_stats(created_at);

-- Index pour la recherche full-text
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_messages_content_gin ON messages USING gin(to_tsvector('french', content));
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audio_tracks_title_gin ON audio_tracks USING gin(to_tsvector('french', title));

-- =============================================================================
-- Triggers pour les nouvelles tables
-- =============================================================================

-- Triggers pour mettre à jour updated_at automatiquement sur les nouvelles tables
DO $$
BEGIN
    -- Trigger pour audio_tracks
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_audio_tracks_updated_at') THEN
        CREATE TRIGGER update_audio_tracks_updated_at 
        BEFORE UPDATE ON audio_tracks
        FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    END IF;
    
    -- Trigger pour playlists
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_playlists_updated_at') THEN
        CREATE TRIGGER update_playlists_updated_at 
        BEFORE UPDATE ON playlists
        FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    END IF;
END $$;

-- =============================================================================
-- Vues pour faciliter l'intégration entre les services
-- =============================================================================

-- Vue pour les messages avec informations utilisateur
CREATE OR REPLACE VIEW messages_with_user AS
SELECT 
    m.*,
    u.username,
    u.avatar_url,
    CASE 
        WHEN m.conversation_id IS NOT NULL THEN 'conversation'
        WHEN m.room_id IS NOT NULL THEN 'room'
        ELSE 'unknown'
    END as message_context
FROM messages m
JOIN users u ON m.author_id = u.id
WHERE m.deleted_at IS NULL;

-- Vue pour les tracks avec informations utilisateur
CREATE OR REPLACE VIEW tracks_with_user AS
SELECT 
    t.*,
    u.username as uploaded_by_username,
    u.avatar_url as uploaded_by_avatar
FROM audio_tracks t
JOIN users u ON t.uploaded_by = u.id;

-- =============================================================================
-- Fonctions utilitaires pour l'API
-- =============================================================================

-- Fonction pour obtenir les messages récents d'une conversation
CREATE OR REPLACE FUNCTION get_recent_messages(
    p_conversation_id BIGINT DEFAULT NULL,
    p_room_id BIGINT DEFAULT NULL,
    p_limit INTEGER DEFAULT 50,
    p_offset INTEGER DEFAULT 0
)
RETURNS TABLE (
    id BIGINT,
    uuid UUID,
    content TEXT,
    author_id BIGINT,
    username VARCHAR,
    avatar_url TEXT,
    message_type VARCHAR,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE
)
LANGUAGE plpgsql
AS $$
BEGIN
    RETURN QUERY
    SELECT 
        m.id,
        m.uuid,
        m.content,
        m.author_id,
        u.username,
        u.avatar_url,
        m.message_type,
        m.created_at,
        m.updated_at
    FROM messages m
    JOIN users u ON m.author_id = u.id
    WHERE 
        (p_conversation_id IS NULL OR m.conversation_id = p_conversation_id)
        AND (p_room_id IS NULL OR m.room_id = p_room_id)
        AND m.deleted_at IS NULL
    ORDER BY m.created_at DESC
    LIMIT p_limit
    OFFSET p_offset;
END;
$$;

-- =============================================================================
-- Données de test pour les nouvelles fonctionnalités (optionnel)
-- =============================================================================

-- Insérer des données de test uniquement si elles n'existent pas
INSERT INTO audio_tracks (title, description, artist, file_path, uploaded_by, is_public)
SELECT 
    'Sample Track 1',
    'Un morceau de démonstration',
    'Artiste Demo',
    '/audio/sample1.mp3',
    1,
    true
WHERE NOT EXISTS (SELECT 1 FROM audio_tracks WHERE title = 'Sample Track 1')
AND EXISTS (SELECT 1 FROM users WHERE id = 1);

-- Commit final
COMMIT;

-- Afficher un message de confirmation
DO $$
BEGIN
    RAISE NOTICE 'Migration de la base de données existante terminée avec succès !';
    RAISE NOTICE 'Nouvelles fonctionnalités ajoutées : audio streaming, playlists, notifications étendues';
    RAISE NOTICE 'Les données existantes ont été préservées';
END $$; 