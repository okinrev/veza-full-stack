# 🔗 gRPC - Chat Server

## Rôle
- Communication inter-services (Go <-> Rust)
- Exposition de services (auth, user, room, message)

## Principales responsabilités
- Définition des services protobuf
- Implémentation des clients/serveurs gRPC
- Gestion des erreurs et retries

## Interactions
- Appelé par le backend Go pour certaines opérations
- Peut appeler d'autres services (stream, analytics)

## Points clés
- Haute performance (protobuf, streaming)
- Sécurité (auth, TLS)
- Observabilité (tracing, logs)

---

*À compléter avec des exemples, schémas, et détails d'implémentation.* 