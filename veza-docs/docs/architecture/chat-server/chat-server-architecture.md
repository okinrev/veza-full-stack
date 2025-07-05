---
id: chat-server-architecture
title: Architecture Chat Server
sidebar_label: Architecture Chat Server
---

# Architecture Chat Server - Veza

## Vue d'ensemble

Ce document décrit l'architecture du serveur de chat Veza.

## Architecture WebSocket

```mermaid
graph TB
    subgraph "Clients"
        WebClient[Web Client]
        MobileClient[Mobile Client]
    end
    
    subgraph "Chat Server"
        WebSocket[WebSocket Handler]
        Hub[Message Hub]
        Channels[Channel Manager]
    end
    
    subgraph "Infrastructure"
        DB[(PostgreSQL)]
        Redis[(Redis)]
    end
    
    WebClient --> WebSocket
    MobileClient --> WebSocket
    WebSocket --> Hub
    Hub --> Channels
    Channels --> DB
    Hub --> Redis
```

## Technologies

- **Langage** : Rust
- **Framework** : Axum
- **WebSocket** : tokio-tungstenite
- **Base de données** : PostgreSQL
- **Cache** : Redis

## Composants principaux
- Serveur WebSocket (Rust)
- PostgreSQL
- Redis

## Diagramme d'architecture

```mermaid
graph TD
    ChatServer[Chat Server] --> DB[PostgreSQL]
    ChatServer --> Redis[Redis Cache]
```

## Ressources
- [README général](../README.md)

---

**Dernière mise à jour** : $(date)
**Version** : 1.0.0 