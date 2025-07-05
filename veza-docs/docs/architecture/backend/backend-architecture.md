---
id: backend-architecture
title: Architecture Backend
sidebar_label: Architecture Backend
---

# Architecture Backend - Veza

## Vue d'ensemble

Ce document décrit l'architecture du backend API de Veza.

## Architecture Hexagonale

```mermaid
graph TB
    subgraph "Adapters"
        HTTP[HTTP Adapter]
        GRPC[gRPC Adapter]
    end
    
    subgraph "Application"
        Services[Services]
        UseCases[Use Cases]
    end
    
    subgraph "Domain"
        Entities[Entities]
        Repositories[Repositories]
    end
    
    subgraph "Infrastructure"
        DB[(PostgreSQL)]
        Redis[(Redis)]
        NATS[NATS]
    end
    
    HTTP --> Services
    GRPC --> Services
    Services --> UseCases
    UseCases --> Repositories
    Repositories --> DB
    Repositories --> Redis
    Services --> NATS
```

## Technologies

- **Langage** : Go 1.21+
- **Framework** : Gin
- **ORM** : GORM
- **Base de données** : PostgreSQL
- **Cache** : Redis
- **Message Queue** : NATS

---

**Dernière mise à jour** : $(date)
**Version** : 1.0.0 