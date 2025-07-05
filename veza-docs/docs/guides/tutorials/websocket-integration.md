---
id: websocket-integration-tutorial
title: Tutoriel WebSocket Integration
sidebar_label: WebSocket Integration
---

# Tutoriel WebSocket Integration - Veza

## Vue d'ensemble

Ce tutoriel explique l'intégration WebSocket pour le chat en temps réel.

## Configuration

### Côté Serveur
```rust
async fn handle_websocket(ws: WebSocket) {
    // Gestion des connexions WebSocket
}
```

### Côté Client
```javascript
const ws = new WebSocket('ws://localhost:3001');
ws.onmessage = (event) => {
    console.log('Message reçu:', event.data);
};
```

## Messages

### Format JSON
```json
{
  "type": "message",
  "room_id": "room_123",
  "content": "Hello world!"
}
```

---

**Dernière mise à jour** : $(date)
**Version** : 1.0.0 