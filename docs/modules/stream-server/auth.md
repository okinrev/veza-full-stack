# 🔐 Authentification - Stream Server

## Rôle
- Authentification des connexions WebSocket/stream
- Validation des tokens JWT
- Gestion des permissions utilisateur

## Principales responsabilités
- Vérification des tokens à la connexion
- Extraction des claims (user_id, roles, etc.)
- Gestion des refresh tokens
- Intégration 2FA/TOTP

## Interactions
- Appelé par le module streaming à chaque connexion
- Utilise Redis pour les sessions
- Peut interroger le backend Go via gRPC

## Points clés
- Sécurité maximale (expiration, rotation, blacklist)
- Support OAuth2
- Audit log des connexions

---

*À compléter avec des exemples, schémas, et détails d’implémentation.* 