---
id: webhooks
title: Webhooks API
sidebar_label: Webhooks
---

# 🔗 Webhooks API - Veza

## 📋 Vue d'ensemble

Ce guide détaille l'utilisation des webhooks dans l'API Veza pour recevoir des notifications en temps réel.

## 🎯 Événements Disponibles

### Événements Utilisateur
```json
{
  "user.created": "Utilisateur créé",
  "user.updated": "Utilisateur modifié",
  "user.deleted": "Utilisateur supprimé",
  "user.login": "Connexion utilisateur",
  "user.logout": "Déconnexion utilisateur"
}
```

### Événements Chat
```json
{
  "message.sent": "Message envoyé",
  "message.received": "Message reçu",
  "room.created": "Salle créée",
  "room.joined": "Utilisateur rejoint une salle",
  "room.left": "Utilisateur quitte une salle"
}
```

### Événements Stream
```json
{
  "stream.started": "Stream démarré",
  "stream.stopped": "Stream arrêté",
  "stream.paused": "Stream en pause",
  "stream.resumed": "Stream repris"
}
```

## 📡 Configuration des Webhooks

### Créer un Webhook
```bash
curl -X POST /api/v1/webhooks \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://your-app.com/webhooks",
    "events": ["user.created", "message.sent"],
    "secret": "your_webhook_secret"
  }'
```

### Format de Payload
```json
{
  "event": "user.created",
  "timestamp": "2024-01-01T12:00:00Z",
  "data": {
    "user_id": "user_123",
    "email": "user@example.com",
    "created_at": "2024-01-01T12:00:00Z"
  },
  "webhook_id": "webhook_456"
}
```

## 🔐 Sécurité

### Signature HMAC
```javascript
// Vérification de la signature
const crypto = require('crypto');

function verifyWebhookSignature(payload, signature, secret) {
  const expectedSignature = crypto
    .createHmac('sha256', secret)
    .update(payload)
    .digest('hex');
    
  return crypto.timingSafeEqual(
    Buffer.from(signature),
    Buffer.from(expectedSignature)
  );
}
```

## 🛠️ Exemple d'Implémentation

```javascript
// Serveur webhook
app.post('/webhooks', (req, res) => {
  const signature = req.headers['x-veza-signature'];
  const payload = JSON.stringify(req.body);
  
  if (!verifyWebhookSignature(payload, signature, WEBHOOK_SECRET)) {
    return res.status(401).json({ error: 'Signature invalide' });
  }
  
  const { event, data } = req.body;
  
  switch (event) {
    case 'user.created':
      handleUserCreated(data);
      break;
    case 'message.sent':
      handleMessageSent(data);
      break;
    default:
      console.log('Événement non géré:', event);
  }
  
  res.status(200).json({ received: true });
});
```

## 📚 Ressources

- [Guide d'API](./README.md)
- [Authentification](./authentication.md)
- [Gestion d'Erreurs](./error-handling.md)

---

**Dernière mise à jour** : $(date)
**Version** : 1.0.0 