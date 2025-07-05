---
id: main
title: Main Chat Server
sidebar_label: Main Chat Server
---

# Main Chat Server

## Vue d'ensemble

Ce document décrit le point d'entrée principal du serveur de chat Veza.

## Structure

Le fichier `main.rs` contient :

- Initialisation du serveur WebSocket
- Configuration des canaux de chat
- Gestion des connexions clients
- Démarrage du serveur

## Configuration

```rust
// Exemple de configuration
struct ChatServerConfig {
    port: u16,
    max_connections: usize,
    heartbeat_interval: Duration,
}
```

## Démarrage

```bash
cargo run --bin chat-server
```

---

**Dernière mise à jour** : $(date)
**Version** : 1.0.0 