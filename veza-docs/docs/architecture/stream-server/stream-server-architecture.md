---
id: stream-server-architecture
title: Architecture Stream Server
sidebar_label: Architecture Stream Server
---

# Architecture Stream Server - Veza

## Vue d'ensemble

Ce document décrit l'architecture du serveur de streaming Veza.

## Architecture de Streaming

```mermaid
graph TB
    subgraph "Clients"
        WebClient[Web Client]
        MobileClient[Mobile Client]
    end
    
    subgraph "Stream Server"
        HTTP[HTTP Handler]
        StreamManager[Stream Manager]
        CodecManager[Codec Manager]
        FileManager[File Manager]
    end
    
    subgraph "Infrastructure"
        DB[(PostgreSQL)]
        Storage[Object Storage]
    end
    
    WebClient --> HTTP
    MobileClient --> HTTP
    HTTP --> StreamManager
    StreamManager --> CodecManager
    CodecManager --> FileManager
    FileManager --> DB
    FileManager --> Storage
```

## Technologies

- **Langage** : Rust
- **Framework** : Axum
- **Codecs** : FFmpeg
- **Base de données** : PostgreSQL
- **Storage** : S3/Cloud Storage

---

**Dernière mise à jour** : $(date)
**Version** : 1.0.0 