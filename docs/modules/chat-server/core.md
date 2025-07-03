# 🧩 Core Domain - Chat Server

## Rôle
- Logique métier centrale (user, room, message, permissions)
- Définition des entités et services principaux

## Principales responsabilités
- Gestion des entités (User, Room, Message)
- Services métier (création, update, suppression, validation)
- Gestion des permissions et rôles

## Interactions
- Appelé par le hub WebSocket
- Utilise PostgreSQL via SQLx
- Utilise Redis pour le cache

## Points clés
- Séparation claire entre domaine et infrastructure
- Testabilité (mocks, tests unitaires)
- Respect du DDD (Domain Driven Design)

---

*À compléter avec des exemples, schémas, et détails d’implémentation.* 