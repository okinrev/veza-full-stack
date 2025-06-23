# Chat WebSocket Protocol - Documentation

## Vue d'ensemble

Le module chat utilise un protocole WebSocket pour la communication temps réel entre le frontend React et le module chat Rust.

## Configuration

```env
VITE_WS_CHAT_URL=ws://localhost:8081/ws
```

## Types de Messages

### Messages Sortants (Frontend → Rust)

```typescript
interface WebSocketOutboundMessage {
  type: 'join' | 'message' | 'dm' | 'room_history' | 'dm_history';
  room?: string;      // Pour join, message, room_history
  content?: string;   // Pour message, dm
  to?: number;        // Pour dm (user_id destinataire)
  with?: number;      // Pour dm_history
  limit?: number;     // Pour historique
}
```

### Messages Entrants (Rust → Frontend)

```typescript
interface WebSocketInboundMessage {
  type: 'message' | 'dm' | 'room_history' | 'dm_history' | 'error';
  data?: any;
  message?: string;
}
```

## Protocole

### 1. Rejoindre un salon
```json
{
  "type": "join",
  "room": "general"
}
```

### 2. Envoyer message salon
```json
{
  "type": "message",
  "room": "general",
  "content": "Hello everyone!"
}
```

### 3. Message direct
```json
{
  "type": "dm",
  "to": 123,
  "content": "Hello!"
}
```

### 4. Demander historique
```json
{
  "type": "room_history",
  "room": "general",
  "limit": 50
}
```

## Implémentation Rust Requise

```rust
// Structure suggérée pour le serveur chat Rust
pub struct ChatServer {
    connections: HashMap<u32, WebSocketSender>,
    rooms: HashMap<String, HashSet<u32>>,
    http_client: reqwest::Client,
}

impl ChatServer {
    async fn handle_message(&mut self, user_id: u32, message: Value) -> Result<(), Error> {
        let msg_type = message.get("type").and_then(|v| v.as_str())?;
        
        match msg_type {
            "join" => self.handle_join(user_id, message).await,
            "message" => self.handle_room_message(user_id, message).await,
            "dm" => self.handle_direct_message(user_id, message).await,
            _ => Err("Unknown message type".into())
        }
    }
    
    async fn broadcast_to_room(&self, room: &str, message: Value) -> Result<(), Error> {
        // Diffuser le message à tous les membres du salon
    }
    
    async fn persist_message(&self, message: &ChatMessage) -> Result<(), Error> {
        // Synchroniser avec le backend Go pour la persistance
        self.http_client
            .post("http://localhost:8080/api/v1/chat/messages")
            .json(message)
            .send()
            .await?;
        Ok(())
    }
}
```

## Authentification WebSocket

```rust
// Validation du token dans la connexion WebSocket
async fn validate_token_with_backend(token: &str) -> Result<User, Error> {
    let client = reqwest::Client::new();
    let response = client
        .get("http://localhost:8080/api/v1/auth/me")
        .bearer_auth(token)
        .send()
        .await?;
        
    if response.status().is_success() {
        Ok(response.json().await?)
    } else {
        Err("Invalid token".into())
    }
}
```

Cette documentation permet l'intégration complète du chat en temps réel. 