-- Script d'initialisation de la base de données pour Veza
-- Compatible avec PostgreSQL 15+
-- Adaptation au schéma existant

-- Extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- Types ENUM (compatibles avec le schéma existant)
DO $$ BEGIN
    CREATE TYPE "public"."conversation_type" AS ENUM ('direct', 'group', 'room');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE "public"."message_status" AS ENUM ('sent', 'delivered', 'read', 'deleted');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE "public"."user_role" AS ENUM ('user', 'moderator', 'admin', 'super_admin');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Fonctions utilitaires (compatibles avec le schéma existant)
CREATE OR REPLACE FUNCTION "public"."cleanup_expired_sessions"() 
RETURNS integer
LANGUAGE plpgsql
AS $$
DECLARE
    deleted_count integer;
BEGIN
    DELETE FROM user_sessions WHERE expires_at < CURRENT_TIMESTAMP;
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$;

CREATE OR REPLACE FUNCTION "public"."current_user_id"() 
RETURNS bigint
LANGUAGE plpgsql
AS $$
BEGIN
    RETURN current_setting('app.current_user_id', true)::bigint;
END;
$$;

CREATE OR REPLACE FUNCTION "public"."update_updated_at_column"()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$;

-- Tables compatibles avec le schéma existant
-- Les tables seront créées seulement si elles n'existent pas déjà

-- Table users (compatible avec votre schéma)
CREATE TABLE IF NOT EXISTS "public"."users" (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID DEFAULT uuid_generate_v4() UNIQUE NOT NULL,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    avatar_url TEXT,
    role "public"."user_role" DEFAULT 'user',
    is_active BOOLEAN DEFAULT true,
    email_verified BOOLEAN DEFAULT false,
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Table conversations (compatible avec votre schéma)
CREATE TABLE IF NOT EXISTS "public"."conversations" (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID DEFAULT uuid_generate_v4() UNIQUE NOT NULL,
    name VARCHAR(255),
    description TEXT,
    type "public"."conversation_type" DEFAULT 'direct',
    is_private BOOLEAN DEFAULT false,
    created_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Table rooms (alias pour conversations de type 'room')
CREATE TABLE IF NOT EXISTS "public"."rooms" (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID DEFAULT uuid_generate_v4() UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_public BOOLEAN DEFAULT true,
    owner_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    max_members INTEGER DEFAULT 100,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Table des membres des conversations/salles (compatible avec votre schéma)
CREATE TABLE IF NOT EXISTS "public"."conversation_members" (
    id BIGSERIAL PRIMARY KEY,
    conversation_id BIGINT REFERENCES conversations(id) ON DELETE CASCADE,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    role "public"."user_role" DEFAULT 'user',
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    left_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(conversation_id, user_id)
);

-- Table des membres des salles (alias pour room_members)
CREATE TABLE IF NOT EXISTS "public"."room_members" (
    id BIGSERIAL PRIMARY KEY,
    room_id BIGINT REFERENCES rooms(id) ON DELETE CASCADE,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    role "public"."user_role" DEFAULT 'user',
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(room_id, user_id)
);

-- Table des messages (compatible avec votre schéma)
CREATE TABLE IF NOT EXISTS "public"."messages" (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID DEFAULT uuid_generate_v4() UNIQUE NOT NULL,
    conversation_id BIGINT REFERENCES conversations(id) ON DELETE CASCADE,
    room_id BIGINT REFERENCES rooms(id) ON DELETE CASCADE,
    author_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    message_type VARCHAR(50) DEFAULT 'text',
    parent_message_id BIGINT REFERENCES messages(id) ON DELETE SET NULL,
    thread_count INTEGER DEFAULT 0,
    is_pinned BOOLEAN DEFAULT false,
    is_edited BOOLEAN DEFAULT false,
    edit_count INTEGER DEFAULT 0,
    status "public"."message_status" DEFAULT 'sent',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    edited_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Table des conversations privées
CREATE TABLE IF NOT EXISTS conversations (
    id SERIAL PRIMARY KEY,
    uuid UUID DEFAULT uuid_generate_v4() UNIQUE NOT NULL,
    user1_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    user2_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    is_blocked BOOLEAN DEFAULT false,
    blocked_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user1_id, user2_id)
);

-- Table des messages privés
CREATE TABLE IF NOT EXISTS direct_messages (
    id SERIAL PRIMARY KEY,
    uuid UUID DEFAULT uuid_generate_v4() UNIQUE NOT NULL,
    conversation_id INTEGER REFERENCES conversations(id) ON DELETE CASCADE,
    sender_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    message_type VARCHAR(20) DEFAULT 'text' CHECK (message_type IN ('text', 'image', 'file', 'audio')),
    is_read BOOLEAN DEFAULT false,
    read_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Table des réactions aux messages
CREATE TABLE IF NOT EXISTS message_reactions (
    id SERIAL PRIMARY KEY,
    message_id INTEGER REFERENCES chat_messages(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    emoji VARCHAR(10) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(message_id, user_id, emoji)
);

-- Table des mentions
CREATE TABLE IF NOT EXISTS message_mentions (
    id SERIAL PRIMARY KEY,
    message_id INTEGER REFERENCES chat_messages(id) ON DELETE CASCADE,
    mentioned_user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(message_id, mentioned_user_id)
);

-- Table des pistes audio
CREATE TABLE IF NOT EXISTS audio_tracks (
    id SERIAL PRIMARY KEY,
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
    uploaded_by INTEGER REFERENCES users(id) ON DELETE CASCADE,
    is_public BOOLEAN DEFAULT true,
    play_count INTEGER DEFAULT 0,
    download_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Table des playlists
CREATE TABLE IF NOT EXISTS playlists (
    id SERIAL PRIMARY KEY,
    uuid UUID DEFAULT uuid_generate_v4() UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    owner_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    is_public BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Table des éléments de playlist
CREATE TABLE IF NOT EXISTS playlist_items (
    id SERIAL PRIMARY KEY,
    playlist_id INTEGER REFERENCES playlists(id) ON DELETE CASCADE,
    track_id INTEGER REFERENCES audio_tracks(id) ON DELETE CASCADE,
    position INTEGER NOT NULL,
    added_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(playlist_id, track_id)
);

-- Table des statistiques d'écoute
CREATE TABLE IF NOT EXISTS listening_stats (
    id SERIAL PRIMARY KEY,
    track_id INTEGER REFERENCES audio_tracks(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    session_id UUID DEFAULT uuid_generate_v4(),
    duration_listened INTEGER, -- en secondes
    completed BOOLEAN DEFAULT false,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Table des notifications
CREATE TABLE IF NOT EXISTS notifications (
    id SERIAL PRIMARY KEY,
    uuid UUID DEFAULT uuid_generate_v4() UNIQUE NOT NULL,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    content TEXT,
    metadata JSONB DEFAULT '{}',
    is_read BOOLEAN DEFAULT false,
    read_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes pour optimiser les performances
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_uuid ON users(uuid);

CREATE INDEX IF NOT EXISTS idx_chat_messages_room_id ON chat_messages(room_id);
CREATE INDEX IF NOT EXISTS idx_chat_messages_author_id ON chat_messages(author_id);
CREATE INDEX IF NOT EXISTS idx_chat_messages_created_at ON chat_messages(created_at);
CREATE INDEX IF NOT EXISTS idx_chat_messages_uuid ON chat_messages(uuid);

CREATE INDEX IF NOT EXISTS idx_direct_messages_conversation_id ON direct_messages(conversation_id);
CREATE INDEX IF NOT EXISTS idx_direct_messages_sender_id ON direct_messages(sender_id);
CREATE INDEX IF NOT EXISTS idx_direct_messages_created_at ON direct_messages(created_at);

CREATE INDEX IF NOT EXISTS idx_conversations_users ON conversations(user1_id, user2_id);

CREATE INDEX IF NOT EXISTS idx_audio_tracks_uploaded_by ON audio_tracks(uploaded_by);
CREATE INDEX IF NOT EXISTS idx_audio_tracks_created_at ON audio_tracks(created_at);
CREATE INDEX IF NOT EXISTS idx_audio_tracks_uuid ON audio_tracks(uuid);

CREATE INDEX IF NOT EXISTS idx_listening_stats_track_id ON listening_stats(track_id);
CREATE INDEX IF NOT EXISTS idx_listening_stats_user_id ON listening_stats(user_id);
CREATE INDEX IF NOT EXISTS idx_listening_stats_created_at ON listening_stats(created_at);

CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_is_read ON notifications(is_read);

-- Index pour la recherche full-text
CREATE INDEX IF NOT EXISTS idx_chat_messages_content_gin ON chat_messages USING gin(to_tsvector('french', content));
CREATE INDEX IF NOT EXISTS idx_audio_tracks_title_gin ON audio_tracks USING gin(to_tsvector('french', title));

-- Fonctions pour les triggers
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers pour mettre à jour updated_at automatiquement
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_chat_rooms_updated_at BEFORE UPDATE ON chat_rooms
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_chat_messages_updated_at BEFORE UPDATE ON chat_messages
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_conversations_updated_at BEFORE UPDATE ON conversations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_direct_messages_updated_at BEFORE UPDATE ON direct_messages
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_audio_tracks_updated_at BEFORE UPDATE ON audio_tracks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_playlists_updated_at BEFORE UPDATE ON playlists
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Données de test pour le développement
INSERT INTO users (username, email, password_hash, first_name, last_name, role) VALUES
('admin', 'admin@veza.com', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LeuyxpXL6Dqm8UHoi', 'Admin', 'User', 'admin'),
('alice', 'alice@example.com', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LeuyxpXL6Dqm8UHoi', 'Alice', 'Smith', 'user'),
('bob', 'bob@example.com', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LeuyxpXL6Dqm8UHoi', 'Bob', 'Johnson', 'user'),
('charlie', 'charlie@example.com', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LeuyxpXL6Dqm8UHoi', 'Charlie', 'Brown', 'moderator')
ON CONFLICT (username) DO NOTHING;

-- Salle de chat générale
INSERT INTO chat_rooms (name, description, is_public, owner_id) VALUES
('General', 'Salle de discussion générale', true, 1),
('Music', 'Discussion autour de la musique', true, 1),
('Tech', 'Discussions techniques', true, 1)
ON CONFLICT DO NOTHING;

-- Ajouter les utilisateurs aux salles
INSERT INTO room_members (room_id, user_id, role) VALUES
(1, 1, 'admin'),
(1, 2, 'member'),
(1, 3, 'member'),
(1, 4, 'moderator'),
(2, 1, 'admin'),
(2, 2, 'member'),
(2, 3, 'member'),
(3, 1, 'admin'),
(3, 4, 'moderator')
ON CONFLICT (room_id, user_id) DO NOTHING;

-- Messages de bienvenue
INSERT INTO chat_messages (room_id, author_id, content, message_type) VALUES
(1, 1, 'Bienvenue dans la salle générale !', 'system'),
(2, 1, 'Partagez votre musique préférée ici !', 'system'),
(3, 1, 'Discussions techniques et développement', 'system')
ON CONFLICT DO NOTHING;

COMMIT; 