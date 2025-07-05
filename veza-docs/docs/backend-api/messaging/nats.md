---
id: nats-messaging
title: Messaging NATS
sidebar_label: NATS Messaging
---

# Messaging NATS - Backend API

## Vue d'ensemble

Ce document décrit l'utilisation de NATS pour la communication inter-services.

## Configuration

### Variables d'Environnement
```bash
NATS_URL=nats://localhost:4222
NATS_CLUSTER_ID=veza-cluster
```

### Connexion
```go
nc, err := nats.Connect("nats://localhost:4222")
if err != nil {
    log.Fatal(err)
}
```

## Patterns

### Pub/Sub
```go
// Publisher
nc.Publish("user.created", data)

// Subscriber
nc.Subscribe("user.created", func(msg *nats.Msg) {
    // Traitement du message
})
```

### Request/Reply
```go
// Request
response, err := nc.Request("user.get", data, time.Second)
```

---

**Dernière mise à jour** : $(date)
**Version** : 1.0.0 