#!/bin/bash

# Script pour forcer la migration de la table refresh_tokens

echo "🔧 Force migration: refresh_tokens expires_at fix"

# Configuration par défaut
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_NAME=${DB_NAME:-veza}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-postgres}

export PGPASSWORD=$DB_PASSWORD

echo "📦 Base de données: $DB_NAME sur $DB_HOST:$DB_PORT"

# Nettoyer la table et changer le type
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME << EOF
-- Afficher la structure actuelle
\d refresh_tokens;

-- Nettoyer les données
TRUNCATE TABLE refresh_tokens CASCADE;

-- Changer le type de colonne
ALTER TABLE refresh_tokens ALTER COLUMN expires_at TYPE BIGINT USING 0;

-- Réinitialiser la séquence
ALTER SEQUENCE refresh_tokens_id_seq RESTART WITH 1;

-- Afficher la nouvelle structure
\d refresh_tokens;

-- Vérifier que tout est OK
SELECT 'Migration refresh_tokens COMPLETED' as status;
EOF

echo "✅ Migration terminée. Redémarrez le serveur Go." 