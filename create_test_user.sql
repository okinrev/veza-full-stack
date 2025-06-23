-- Script pour créer un utilisateur de test
-- Password: "secret" (hashé avec bcrypt)

INSERT INTO users (
    username, 
    email, 
    password_hash, 
    first_name, 
    last_name, 
    role, 
    is_active, 
    is_verified, 
    created_at, 
    updated_at
) VALUES (
    'testuser',
    'test@test.com',
    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
    'Test',
    'User',
    'user',
    true,
    true,
    NOW(),
    NOW()
) ON CONFLICT (email) DO NOTHING;

-- Vérifier que l'utilisateur a été créé
SELECT id, username, email, role, is_active, created_at FROM users WHERE email = 'test@test.com'; 