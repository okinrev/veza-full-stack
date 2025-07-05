---
id: main
title: Main Stream Server
sidebar_label: Main Stream Server
---

# Main Stream Server

## Vue d'ensemble

Ce document décrit le point d'entrée principal du serveur de streaming Veza.

## Structure

Le fichier `main.rs` contient :

- Initialisation du serveur de streaming
- Configuration des codecs audio
- Gestion des flux de données
- Démarrage du serveur

## Configuration

```rust
// Exemple de configuration
struct StreamServerConfig {
    port: u16,
    buffer_size: usize,
    codec: AudioCodec,
}
```

## Démarrage

```bash
cargo run --bin stream-server
```

---

**Dernière mise à jour** : $(date)
**Version** : 1.0.0 