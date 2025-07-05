---
id: api-integration
sidebar_label: Intégration API Externe
---

# 🔌 Guide d'Intégration API Externe - Veza Platform

> **Guide complet pour intégrer la plateforme Veza dans vos applications**

## 📋 Table des Matières

- [Vue d'ensemble](#vue-densemble)
- [Authentification](#authentification)
- [API REST](#api-rest)
- [API gRPC](#api-grpc)
- [WebSocket](#websocket)
- [Gestion des Erreurs](#gestion-des-erreurs)
- [Exemples d'Intégration](#exemples-dintgration)
- [Bonnes Pratiques](#bonnes-pratiques)

## 🎯 Vue d'ensemble

Ce guide vous accompagne dans l'intégration de la plateforme Veza dans vos applications externes, que ce soit pour récupérer des données, envoyer des notifications ou automatiser des processus.

### 🌟 Types d'Intégration

- **📊 Analytics** : Récupération de métriques et statistiques
- **💬 Messagerie** : Envoi et réception de messages
- **🎵 Streaming** : Gestion des streams audio
- **👥 Utilisateurs** : Gestion des comptes et permissions
- **🔔 Notifications** : Système d'alertes et notifications
- **📁 Fichiers** : Upload et gestion de fichiers

### 🔧 Prérequis

- **Compte développeur** : Inscription sur [dev.veza.app](https://dev.veza.app)
- **Clé API** : Génération dans le dashboard développeur
- **Certificats SSL** : Pour les intégrations sécurisées
- **Rate Limiting** : Respect des limites d'API

## 🔐 Authentification

### 1. 🔑 API Keys

```bash
# Génération d'une clé API
curl -X POST https://api.veza.app/v1/auth/api-keys \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Integration",
    "permissions": ["read:users", "write:messages"],
    "expires_at": "2024-12-31T23:59:59Z"
  }'
```

**Réponse :**
```json
{
  "api_key": "veza_live_1234567890abcdef",
  "secret": "sk_live_1234567890abcdef",
  "permissions": ["read:users", "write:messages"],
  "created_at": "2024-01-01T00:00:00Z",
  "expires_at": "2024-12-31T23:59:59Z"
}
```

### 2. 🔐 OAuth 2.0

```javascript
// Configuration OAuth
const oauthConfig = {
  clientId: 'your_client_id',
  clientSecret: 'your_client_secret',
  redirectUri: 'https://your-app.com/callback',
  scope: 'read:users write:messages read:streams',
  authorizationUrl: 'https://api.veza.app/oauth/authorize',
  tokenUrl: 'https://api.veza.app/oauth/token'
};

// Flux d'authentification
async function authenticate() {
  // 1. Rediriger vers l'URL d'autorisation
  const authUrl = `${oauthConfig.authorizationUrl}?` +
    `client_id=${oauthConfig.clientId}&` +
    `redirect_uri=${encodeURIComponent(oauthConfig.redirectUri)}&` +
    `scope=${encodeURIComponent(oauthConfig.scope)}&` +
    `response_type=code`;
  
  window.location.href = authUrl;
}

// 2. Échanger le code contre un token
async function exchangeCode(code) {
  const response = await fetch(oauthConfig.tokenUrl, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded',
    },
    body: new URLSearchParams({
      grant_type: 'authorization_code',
      client_id: oauthConfig.clientId,
      client_secret: oauthConfig.clientSecret,
      code: code,
      redirect_uri: oauthConfig.redirectUri,
    }),
  });
  
  return response.json();
}
```

### 3. 🔑 JWT Tokens

```go
// internal/integration/jwt_client.go
package integration

import (
    "crypto/rsa"
    "time"
    
    "github.com/golang-jwt/jwt/v5"
)

type JWTClient struct {
    privateKey *rsa.PrivateKey
    issuer     string
    audience    string
}

func NewJWTClient(privateKeyPath string) (*JWTClient, error) {
    privateKey, err := loadPrivateKey(privateKeyPath)
    if err != nil {
        return nil, err
    }
    
    return &JWTClient{
        privateKey: privateKey,
        issuer:     "your-app",
        audience:   "veza-api",
    }, nil
}

func (jc *JWTClient) GenerateToken(userID int64, permissions []string) (string, error) {
    now := time.Now()
    claims := jwt.MapClaims{
        "user_id":     userID,
        "permissions": permissions,
        "iss":         jc.issuer,
        "aud":         jc.audience,
        "iat":         now.Unix(),
        "exp":         now.Add(1 * time.Hour).Unix(),
        "nbf":         now.Unix(),
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
    return token.SignedString(jc.privateKey)
}
```

## 📡 API REST

### 1. 🔍 Endpoints Principaux

```bash
# Base URL
https://api.veza.app/v1

# Endpoints disponibles
GET    /users                    # Liste des utilisateurs
GET    /users/{id}              # Détails d'un utilisateur
POST   /users                   # Créer un utilisateur
PUT    /users/{id}              # Modifier un utilisateur
DELETE /users/{id}              # Supprimer un utilisateur

GET    /chat/rooms              # Liste des salons
GET    /chat/rooms/{id}/messages # Messages d'un salon
POST   /chat/messages           # Envoyer un message
PUT    /chat/messages/{id}      # Modifier un message
DELETE /chat/messages/{id}      # Supprimer un message

GET    /streams                 # Liste des streams
POST   /streams                 # Créer un stream
PUT    /streams/{id}            # Modifier un stream
DELETE /streams/{id}            # Supprimer un stream

GET    /analytics/metrics       # Métriques analytics
GET    /analytics/reports       # Rapports détaillés
```

### 2. 📊 Exemples d'Appels

#### Récupération d'Utilisateurs

```python
# Python - Récupération d'utilisateurs
import requests

class VezaAPI:
    def __init__(self, api_key):
        self.api_key = api_key
        self.base_url = "https://api.veza.app/v1"
        self.headers = {
            "Authorization": f"Bearer {api_key}",
            "Content-Type": "application/json"
        }
    
    def get_users(self, page=1, limit=50, filters=None):
        params = {
            "page": page,
            "limit": limit
        }
        
        if filters:
            params.update(filters)
        
        response = requests.get(
            f"{self.base_url}/users",
            headers=self.headers,
            params=params
        )
        
        if response.status_code == 200:
            return response.json()
        else:
            raise Exception(f"Error: {response.status_code} - {response.text}")
    
    def create_user(self, user_data):
        response = requests.post(
            f"{self.base_url}/users",
            headers=self.headers,
            json=user_data
        )
        
        if response.status_code == 201:
            return response.json()
        else:
            raise Exception(f"Error: {response.status_code} - {response.text}")

# Utilisation
api = VezaAPI("your_api_key")
users = api.get_users(page=1, limit=10)
print(f"Found {len(users['data'])} users")
```

#### Envoi de Messages

```javascript
// JavaScript - Envoi de messages
class VezaChatAPI {
    constructor(apiKey) {
        this.apiKey = apiKey;
        this.baseUrl = 'https://api.veza.app/v1';
        this.headers = {
            'Authorization': `Bearer ${apiKey}`,
            'Content-Type': 'application/json'
        };
    }
    
    async sendMessage(roomId, content, options = {}) {
        const messageData = {
            room_id: roomId,
            content: content,
            message_type: options.type || 'text',
            metadata: options.metadata || {},
            reply_to_id: options.replyToId || null
        };
        
        const response = await fetch(`${this.baseUrl}/chat/messages`, {
            method: 'POST',
            headers: this.headers,
            body: JSON.stringify(messageData)
        });
        
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
        
        return response.json();
    }
    
    async getMessages(roomId, limit = 50, beforeId = null) {
        const params = new URLSearchParams({
            room_id: roomId,
            limit: limit
        });
        
        if (beforeId) {
            params.append('before_id', beforeId);
        }
        
        const response = await fetch(`${this.baseUrl}/chat/messages?${params}`, {
            headers: this.headers
        });
        
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
        
        return response.json();
    }
}

// Utilisation
const chatAPI = new VezaChatAPI('your_api_key');

// Envoyer un message
chatAPI.sendMessage(123, 'Hello from integration!', {
    type: 'text',
    metadata: { source: 'external_app' }
}).then(message => {
    console.log('Message sent:', message);
});

// Récupérer les messages
chatAPI.getMessages(123, 20).then(messages => {
    console.log('Messages:', messages);
});
```

#### Gestion des Streams

```go
// Go - Gestion des streams
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

type VezaStreamAPI struct {
    APIKey   string
    BaseURL  string
    Client   *http.Client
}

type Stream struct {
    ID          int64     `json:"id"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    UserID      int64     `json:"user_id"`
    Status      string    `json:"status"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type CreateStreamRequest struct {
    Title       string `json:"title"`
    Description string `json:"description"`
    Category    string `json:"category"`
    Visibility  string `json:"visibility"`
}

func NewVezaStreamAPI(apiKey string) *VezaStreamAPI {
    return &VezaStreamAPI{
        APIKey:  apiKey,
        BaseURL: "https://api.veza.app/v1",
        Client:  &http.Client{Timeout: 30 * time.Second},
    }
}

func (api *VezaStreamAPI) CreateStream(req CreateStreamRequest) (*Stream, error) {
    jsonData, err := json.Marshal(req)
    if err != nil {
        return nil, err
    }
    
    httpReq, err := http.NewRequest("POST", api.BaseURL+"/streams", bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, err
    }
    
    httpReq.Header.Set("Authorization", "Bearer "+api.APIKey)
    httpReq.Header.Set("Content-Type", "application/json")
    
    resp, err := api.Client.Do(httpReq)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusCreated {
        return nil, fmt.Errorf("HTTP %d: failed to create stream", resp.StatusCode)
    }
    
    var stream Stream
    if err := json.NewDecoder(resp.Body).Decode(&stream); err != nil {
        return nil, err
    }
    
    return &stream, nil
}

func (api *VezaStreamAPI) GetStream(streamID int64) (*Stream, error) {
    httpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/streams/%d", api.BaseURL, streamID), nil)
    if err != nil {
        return nil, err
    }
    
    httpReq.Header.Set("Authorization", "Bearer "+api.APIKey)
    
    resp, err := api.Client.Do(httpReq)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("HTTP %d: failed to get stream", resp.StatusCode)
    }
    
    var stream Stream
    if err := json.NewDecoder(resp.Body).Decode(&stream); err != nil {
        return nil, err
    }
    
    return &stream, nil
}

// Utilisation
func main() {
    api := NewVezaStreamAPI("your_api_key")
    
    // Créer un stream
    stream, err := api.CreateStream(CreateStreamRequest{
        Title:       "My Integration Stream",
        Description: "Stream created via API",
        Category:    "talk",
        Visibility:  "public",
    })
    
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Stream created: %s (ID: %d)\n", stream.Title, stream.ID)
}
```

## 🔌 API gRPC

### 1. 📋 Configuration Protobuf

```protobuf
// proto/veza_api.proto
syntax = "proto3";

package veza.api;

import "google/protobuf/timestamp.proto";

service VezaAPI {
  // Gestion des utilisateurs
  rpc GetUser(GetUserRequest) returns (User);
  rpc CreateUser(CreateUserRequest) returns (User);
  rpc UpdateUser(UpdateUserRequest) returns (User);
  rpc DeleteUser(DeleteUserRequest) returns (DeleteUserResponse);
  
  // Gestion des messages
  rpc SendMessage(SendMessageRequest) returns (Message);
  rpc GetMessages(GetMessagesRequest) returns (GetMessagesResponse);
  rpc StreamMessages(StreamMessagesRequest) returns (stream Message);
  
  // Gestion des streams
  rpc CreateStream(CreateStreamRequest) returns (Stream);
  rpc GetStream(GetStreamRequest) returns (Stream);
  rpc UpdateStream(UpdateStreamRequest) returns (Stream);
  rpc DeleteStream(DeleteStreamRequest) returns (DeleteStreamResponse);
  
  // Analytics
  rpc GetMetrics(GetMetricsRequest) returns (Metrics);
  rpc StreamMetrics(StreamMetricsRequest) returns (stream MetricData);
}

message User {
  int64 id = 1;
  string email = 2;
  string username = 3;
  string first_name = 4;
  string last_name = 5;
  repeated string roles = 6;
  google.protobuf.Timestamp created_at = 7;
  google.protobuf.Timestamp updated_at = 8;
}

message Message {
  int64 id = 1;
  int64 room_id = 2;
  int64 user_id = 3;
  string content = 4;
  string message_type = 5;
  bytes metadata = 6;
  google.protobuf.Timestamp created_at = 7;
}

message Stream {
  int64 id = 1;
  string title = 2;
  string description = 3;
  int64 user_id = 4;
  string status = 5;
  google.protobuf.Timestamp created_at = 6;
  google.protobuf.Timestamp updated_at = 7;
}
```

### 2. 🔧 Client gRPC

```go
// internal/integration/grpc_client.go
package integration

import (
    "context"
    "time"
    
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    
    pb "github.com/veza/proto"
)

type VezaGRPCClient struct {
    client pb.VezaAPIClient
    conn   *grpc.ClientConn
}

func NewVezaGRPCClient(serverAddr string) (*VezaGRPCClient, error) {
    conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        return nil, err
    }
    
    client := pb.NewVezaAPIClient(conn)
    
    return &VezaGRPCClient{
        client: client,
        conn:   conn,
    }, nil
}

func (vgc *VezaGRPCClient) Close() error {
    return vgc.conn.Close()
}

func (vgc *VezaGRPCClient) GetUser(ctx context.Context, userID int64) (*pb.User, error) {
    req := &pb.GetUserRequest{Id: userID}
    return vgc.client.GetUser(ctx, req)
}

func (vgc *VezaGRPCClient) SendMessage(ctx context.Context, roomID int64, content string) (*pb.Message, error) {
    req := &pb.SendMessageRequest{
        RoomId:  roomID,
        Content: content,
    }
    return vgc.client.SendMessage(ctx, req)
}

func (vgc *VezaGRPCClient) StreamMessages(ctx context.Context, roomID int64) (<-chan *pb.Message, error) {
    req := &pb.StreamMessagesRequest{RoomId: roomID}
    stream, err := vgc.client.StreamMessages(ctx, req)
    if err != nil {
        return nil, err
    }
    
    messages := make(chan *pb.Message)
    
    go func() {
        defer close(messages)
        for {
            msg, err := stream.Recv()
            if err != nil {
                return
            }
            messages <- msg
        }
    }()
    
    return messages, nil
}

// Utilisation
func main() {
    client, err := NewVezaGRPCClient("localhost:9090")
    if err != nil {
        panic(err)
    }
    defer client.Close()
    
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    // Récupérer un utilisateur
    user, err := client.GetUser(ctx, 123)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("User: %s (%s)\n", user.Username, user.Email)
    
    // Envoyer un message
    message, err := client.SendMessage(ctx, 456, "Hello from gRPC!")
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Message sent: %s\n", message.Content)
    
    // Écouter les messages en streaming
    messages, err := client.StreamMessages(ctx, 456)
    if err != nil {
        panic(err)
    }
    
    for msg := range messages {
        fmt.Printf("New message: %s\n", msg.Content)
    }
}
```

## 🔌 WebSocket

### 1. 🔗 Connexion WebSocket

```javascript
// JavaScript - Client WebSocket
class VezaWebSocketClient {
    constructor(apiKey, options = {}) {
        this.apiKey = apiKey;
        this.baseUrl = options.baseUrl || 'wss://api.veza.app/ws';
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = options.maxReconnectAttempts || 5;
        this.reconnectDelay = options.reconnectDelay || 1000;
        this.listeners = new Map();
        this.connected = false;
    }
    
    connect() {
        return new Promise((resolve, reject) => {
            this.ws = new WebSocket(`${this.baseUrl}?token=${this.apiKey}`);
            
            this.ws.onopen = () => {
                this.connected = true;
                this.reconnectAttempts = 0;
                console.log('WebSocket connected');
                resolve();
            };
            
            this.ws.onmessage = (event) => {
                try {
                    const data = JSON.parse(event.data);
                    this.handleMessage(data);
                } catch (error) {
                    console.error('Failed to parse message:', error);
                }
            };
            
            this.ws.onclose = () => {
                this.connected = false;
                console.log('WebSocket disconnected');
                this.attemptReconnect();
            };
            
            this.ws.onerror = (error) => {
                console.error('WebSocket error:', error);
                reject(error);
            };
        });
    }
    
    attemptReconnect() {
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
            this.reconnectAttempts++;
            console.log(`Attempting to reconnect (${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
            
            setTimeout(() => {
                this.connect().catch(console.error);
            }, this.reconnectDelay * this.reconnectAttempts);
        }
    }
    
    handleMessage(data) {
        const { type, payload } = data;
        
        if (this.listeners.has(type)) {
            this.listeners.get(type).forEach(callback => {
                try {
                    callback(payload);
                } catch (error) {
                    console.error(`Error in ${type} listener:`, error);
                }
            });
        }
    }
    
    on(event, callback) {
        if (!this.listeners.has(event)) {
            this.listeners.set(event, []);
        }
        this.listeners.get(event).push(callback);
    }
    
    off(event, callback) {
        if (this.listeners.has(event)) {
            const callbacks = this.listeners.get(event);
            const index = callbacks.indexOf(callback);
            if (index > -1) {
                callbacks.splice(index, 1);
            }
        }
    }
    
    send(type, payload) {
        if (this.connected) {
            this.ws.send(JSON.stringify({ type, payload }));
        } else {
            throw new Error('WebSocket not connected');
        }
    }
    
    subscribe(roomId) {
        this.send('subscribe', { room_id: roomId });
    }
    
    unsubscribe(roomId) {
        this.send('unsubscribe', { room_id: roomId });
    }
    
    sendMessage(roomId, content) {
        this.send('message', {
            room_id: roomId,
            content: content
        });
    }
}

// Utilisation
const wsClient = new VezaWebSocketClient('your_api_key');

// Écouter les nouveaux messages
wsClient.on('message', (message) => {
    console.log('New message:', message);
});

// Écouter les utilisateurs connectés
wsClient.on('user_joined', (user) => {
    console.log('User joined:', user.username);
});

wsClient.on('user_left', (user) => {
    console.log('User left:', user.username);
});

// Se connecter et s'abonner
wsClient.connect().then(() => {
    wsClient.subscribe(123); // Salon ID
    
    // Envoyer un message
    wsClient.sendMessage(123, 'Hello from WebSocket!');
}).catch(console.error);
```

### 2. 🔧 Serveur WebSocket

```go
// internal/websocket/server.go
package websocket

import (
    "context"
    "encoding/json"
    "log"
    "net/http"
    "sync"
    
    "github.com/gorilla/websocket"
)

type WebSocketServer struct {
    upgrader websocket.Upgrader
    clients  map[*Client]bool
    rooms    map[int64]*Room
    mutex    sync.RWMutex
}

type Client struct {
    conn     *websocket.Conn
    userID   int64
    username string
    server   *WebSocketServer
    send     chan []byte
}

type Room struct {
    ID      int64
    clients map[*Client]bool
    mutex   sync.RWMutex
}

type Message struct {
    Type    string      `json:"type"`
    Payload interface{} `json:"payload"`
}

func NewWebSocketServer() *WebSocketServer {
    return &WebSocketServer{
        upgrader: websocket.Upgrader{
            CheckOrigin: func(r *http.Request) bool {
                return true // À configurer selon vos besoins
            },
        },
        clients: make(map[*Client]bool),
        rooms:   make(map[int64]*Room),
    }
}

func (wss *WebSocketServer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
    conn, err := wss.upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Printf("WebSocket upgrade failed: %v", err)
        return
    }
    
    // Authentifier le client
    token := r.URL.Query().Get("token")
    userID, username, err := wss.authenticateToken(token)
    if err != nil {
        conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "Invalid token"))
        conn.Close()
        return
    }
    
    client := &Client{
        conn:     conn,
        userID:   userID,
        username: username,
        server:   wss,
        send:     make(chan []byte, 256),
    }
    
    wss.registerClient(client)
    
    go client.writePump()
    go client.readPump()
}

func (client *Client) readPump() {
    defer func() {
        client.server.unregisterClient(client)
        client.conn.Close()
    }()
    
    for {
        _, message, err := client.conn.ReadMessage()
        if err != nil {
            break
        }
        
        var msg Message
        if err := json.Unmarshal(message, &msg); err != nil {
            continue
        }
        
        client.handleMessage(msg)
    }
}

func (client *Client) writePump() {
    defer client.conn.Close()
    
    for {
        select {
        case message, ok := <-client.send:
            if !ok {
                client.conn.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }
            
            w, err := client.conn.NextWriter(websocket.TextMessage)
            if err != nil {
                return
            }
            
            w.Write(message)
            
            if err := w.Close(); err != nil {
                return
            }
        }
    }
}

func (client *Client) handleMessage(msg Message) {
    switch msg.Type {
    case "subscribe":
        if payload, ok := msg.Payload.(map[string]interface{}); ok {
            if roomID, ok := payload["room_id"].(float64); ok {
                client.server.subscribeToRoom(client, int64(roomID))
            }
        }
        
    case "unsubscribe":
        if payload, ok := msg.Payload.(map[string]interface{}); ok {
            if roomID, ok := payload["room_id"].(float64); ok {
                client.server.unsubscribeFromRoom(client, int64(roomID))
            }
        }
        
    case "message":
        if payload, ok := msg.Payload.(map[string]interface{}); ok {
            if roomID, ok := payload["room_id"].(float64); ok {
                if content, ok := payload["content"].(string); ok {
                    client.server.broadcastMessage(int64(roomID), Message{
                        Type: "message",
                        Payload: map[string]interface{}{
                            "room_id":   int64(roomID),
                            "user_id":   client.userID,
                            "username":  client.username,
                            "content":   content,
                            "timestamp": time.Now(),
                        },
                    })
                }
            }
        }
    }
}
```

## ⚠️ Gestion des Erreurs

### 1. 🔍 Codes d'Erreur

```go
// internal/integration/errors.go
package integration

import "fmt"

type VezaError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
}

func (e *VezaError) Error() string {
    return fmt.Sprintf("Veza API Error %d: %s", e.Code, e.Message)
}

var (
    ErrInvalidToken     = &VezaError{Code: 401, Message: "Invalid or expired token"}
    ErrInsufficientPerm = &VezaError{Code: 403, Message: "Insufficient permissions"}
    ErrResourceNotFound = &VezaError{Code: 404, Message: "Resource not found"}
    ErrRateLimitExceeded = &VezaError{Code: 429, Message: "Rate limit exceeded"}
    ErrInternalError    = &VezaError{Code: 500, Message: "Internal server error"}
)
```

### 2. 🔄 Retry Logic

```go
// internal/integration/retry.go
package integration

import (
    "context"
    "time"
)

type RetryConfig struct {
    MaxAttempts int
    BaseDelay    time.Duration
    MaxDelay     time.Duration
    BackoffFactor float64
}

func DefaultRetryConfig() *RetryConfig {
    return &RetryConfig{
        MaxAttempts:  3,
        BaseDelay:    time.Second,
        MaxDelay:     30 * time.Second,
        BackoffFactor: 2.0,
    }
}

func (rc *RetryConfig) Retry(ctx context.Context, operation func() error) error {
    var lastErr error
    
    for attempt := 0; attempt < rc.MaxAttempts; attempt++ {
        if err := operation(); err == nil {
            return nil
        } else {
            lastErr = err
        }
        
        if attempt < rc.MaxAttempts-1 {
            delay := rc.calculateDelay(attempt)
            select {
            case <-time.After(delay):
                continue
            case <-ctx.Done():
                return ctx.Err()
            }
        }
    }
    
    return lastErr
}

func (rc *RetryConfig) calculateDelay(attempt int) time.Duration {
    delay := rc.BaseDelay * time.Duration(rc.BackoffFactor*float64(attempt))
    if delay > rc.MaxDelay {
        delay = rc.MaxDelay
    }
    return delay
}
```

## 💻 Exemples d'Intégration

### 1. 🤖 Bot Discord

```python
# bot_discord.py
import discord
import asyncio
import aiohttp
from discord.ext import commands

class VezaBot(commands.Bot):
    def __init__(self):
        super().__init__(command_prefix='!')
        self.veza_api = VezaAPI('your_api_key')
        self.session = aiohttp.ClientSession()
    
    async def on_ready(self):
        print(f'{self.user} has connected to Discord!')
    
    @commands.command()
    async def stream(self, ctx, *, title):
        """Démarrer un stream Veza depuis Discord"""
        try:
            # Créer le stream
            stream = await self.veza_api.create_stream({
                'title': title,
                'description': f'Stream démarré depuis Discord par {ctx.author.name}',
                'category': 'talk',
                'visibility': 'public'
            })
            
            await ctx.send(f'🎵 Stream créé: {stream["title"]}\n'
                          f'🔗 Lien: https://veza.app/stream/{stream["id"]}')
        
        except Exception as e:
            await ctx.send(f'❌ Erreur: {str(e)}')
    
    @commands.command()
    async def chat(self, ctx, room_id: int, *, message):
        """Envoyer un message dans un salon Veza"""
        try:
            msg = await self.veza_api.send_message({
                'room_id': room_id,
                'content': f'[Discord] {ctx.author.name}: {message}',
                'metadata': {
                    'source': 'discord',
                    'discord_user': ctx.author.name,
                    'discord_channel': ctx.channel.name
                }
            })
            
            await ctx.send(f'✅ Message envoyé dans le salon {room_id}')
        
        except Exception as e:
            await ctx.send(f'❌ Erreur: {str(e)}')

# Utilisation
bot = VezaBot()
bot.run('your_discord_token')
```

### 2. 📊 Dashboard Analytics

```javascript
// dashboard.js
class VezaDashboard {
    constructor(apiKey) {
        this.apiKey = apiKey;
        this.charts = {};
        this.initCharts();
        this.startRealTimeUpdates();
    }
    
    initCharts() {
        // Graphique des utilisateurs actifs
        this.charts.activeUsers = new Chart(document.getElementById('activeUsers'), {
            type: 'line',
            data: {
                labels: [],
                datasets: [{
                    label: 'Utilisateurs Actifs',
                    data: [],
                    borderColor: 'rgb(75, 192, 192)',
                    tension: 0.1
                }]
            },
            options: {
                responsive: true,
                scales: {
                    y: {
                        beginAtZero: true
                    }
                }
            }
        });
        
        // Graphique des messages
        this.charts.messages = new Chart(document.getElementById('messages'), {
            type: 'bar',
            data: {
                labels: [],
                datasets: [{
                    label: 'Messages par Heure',
                    data: [],
                    backgroundColor: 'rgba(54, 162, 235, 0.2)',
                    borderColor: 'rgb(54, 162, 235)',
                    borderWidth: 1
                }]
            }
        });
    }
    
    async updateMetrics() {
        try {
            const response = await fetch('https://api.veza.app/v1/analytics/metrics', {
                headers: {
                    'Authorization': `Bearer ${this.apiKey}`
                }
            });
            
            const metrics = await response.json();
            
            // Mettre à jour les graphiques
            this.updateActiveUsersChart(metrics.active_users);
            this.updateMessagesChart(metrics.messages_per_hour);
            this.updateSystemStatus(metrics.system_status);
            
        } catch (error) {
            console.error('Failed to update metrics:', error);
        }
    }
    
    updateActiveUsersChart(data) {
        const chart = this.charts.activeUsers;
        const now = new Date().toLocaleTimeString();
        
        chart.data.labels.push(now);
        chart.data.datasets[0].data.push(data);
        
        // Garder seulement les 20 derniers points
        if (chart.data.labels.length > 20) {
            chart.data.labels.shift();
            chart.data.datasets[0].data.shift();
        }
        
        chart.update();
    }
    
    startRealTimeUpdates() {
        // Mettre à jour toutes les 30 secondes
        setInterval(() => {
            this.updateMetrics();
        }, 30000);
        
        // Mise à jour initiale
        this.updateMetrics();
    }
}

// Utilisation
const dashboard = new VezaDashboard('your_api_key');
```

## ✅ Bonnes Pratiques

### 1. 🔒 Sécurité

- **Toujours utiliser HTTPS/WSS** pour les connexions
- **Valider les tokens** à chaque requête
- **Implémenter le rate limiting** côté client
- **Logger les erreurs** pour le debugging
- **Ne jamais exposer les clés API** dans le code client

### 2. 🚀 Performance

- **Utiliser la pagination** pour les grandes listes
- **Implémenter le caching** pour les données statiques
- **Optimiser les requêtes** avec des filtres appropriés
- **Utiliser le streaming** pour les données en temps réel
- **Gérer les timeouts** pour éviter les blocages

### 3. 🔄 Robustesse

- **Implémenter le retry logic** avec backoff exponentiel
- **Gérer les erreurs réseau** gracieusement
- **Valider les réponses** avant traitement
- **Utiliser des timeouts** appropriés
- **Tester les intégrations** régulièrement

### 4. 📊 Monitoring

- **Logger les métriques** d'utilisation
- **Surveiller les erreurs** et timeouts
- **Tracer les performances** des appels API
- **Alertes** pour les problèmes critiques
- **Dashboard** pour visualiser l'état

---

## 🔗 Liens croisés

- [API Reference](../api/README.md)
- [Architecture Globale](../architecture/global-architecture.md)
- [Sécurité Avancée](../guides/advanced-security.md)
- [Performance](../guides/performance-optimization.md)

---

## Pour aller plus loin

- [Guide Utilisateur](../guides/user-guide.md)
- [Configuration Avancée](../guides/advanced-configuration.md)
- [Troubleshooting](../troubleshooting/README.md)
- [Tests](../testing/README.md) 