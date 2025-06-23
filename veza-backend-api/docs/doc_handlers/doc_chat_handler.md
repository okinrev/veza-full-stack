# ChatHandler Documentation

This file refer to : backend/internal/handlers/chat.go

## Package & Dependencies
```go
package handlers
imports: gin, database, middleware, models, strconv, strings, database/sql
```

## Struct
```go
type ChatHandler struct {
    db *database.DB
}
```

## Constructor
```go
func NewChatHandler(db *database.DB) *ChatHandler
```

## Types/Models
```go
type MessageResponse struct {
    ID, FromUser int
    ToUser *int
    Room *string
    Content, Timestamp, Username string
    IsRead bool
}

type MessageWithAvatar struct {
    MessageResponse + Avatar *string + EditedAt *string
}

type EnhancedConversationSummary struct {
    UserID, UnreadCount int
    Username string
    FirstName, LastName, Avatar *string
    LastMessage, LastActivity string
    IsOnline bool
    LastSeen *string
}

type SendMessageRequest struct {
    Content string `binding:"required,min=1,max=1000"`
}

type RoomResponse struct {
    ID int
    Name string
    Description *string
    IsPrivate bool
    CreatorID *int
    CreatorName, CreatedAt, UpdatedAt string
}

type CreateRoomRequest struct {
    Name string `binding:"required,min=1,max=50"`
    Description *string
    IsPrivate bool
}
```

## Methods

### Direct Messages
```go
func (h *ChatHandler) GetDirectMessages(c *gin.Context)
func (h *ChatHandler) SendDirectMessage(c *gin.Context)
```
- **Auth**: common.GetUserIDFromContext required
- **Params**: user_id (URL param), page/limit (query)
- **Process**: Message history with pagination, auto-mark as read
- **Database**: messages table with users JOIN
- **Validation**: User existence, no self-messaging

### Conversations
```go
func (h *ChatHandler) GetConversations(c *gin.Context)
```
- **Auth**: Required
- **Process**: Complex CTE query for conversation partners, last messages, unread counts
- **Response**: Enhanced conversation list with user details

### Message Management
```go
func (h *ChatHandler) MarkAsRead(c *gin.Context)
func (h *ChatHandler) EditMessage(c *gin.Context)
func (h *ChatHandler) DeleteMessage(c *gin.Context)
func (h *ChatHandler) GetUnreadCount(c *gin.Context)
```
- **Auth**: Required for all
- **Validation**: Message ownership for edit/delete
- **Process**: Soft delete (content replacement), edit tracking

### Room Management
```go
func (h *ChatHandler) GetPublicRooms(c *gin.Context)
func (h *ChatHandler) CreateRoom(c *gin.Context)
func (h *ChatHandler) GetRoomMessages(c *gin.Context)
func (h *ChatHandler) SendRoomMessage(c *gin.Context)
```
- **Auth**: Required for create/send, optional for public listing
- **Validation**: Room name uniqueness, membership for private rooms
- **Process**: Room creation, message posting, access control

## Route Mapping Expectations
- `GET /chat/messages/:user_id` → GetDirectMessages
- `POST /chat/messages/:user_id` → SendDirectMessage
- `GET /chat/conversations` → GetConversations
- `PUT /chat/messages/:user_id/read` → MarkAsRead
- `PUT /chat/messages/:message_id` → EditMessage
- `DELETE /chat/messages/:message_id` → DeleteMessage
- `GET /chat/unread` → GetUnreadCount
- `GET /chat/rooms` → GetPublicRooms
- `POST /chat/rooms` → CreateRoom
- `GET /chat/rooms/:room/messages` → GetRoomMessages
- `POST /chat/rooms/:room/messages` → SendRoomMessage

## Middleware Dependencies
- Authentication middleware for user context
- JSON binding for POST/PUT requests

## Database Tables
- messages (content, timestamps, read status)
- users (profile data, avatars)
- rooms (room metadata)
- room_members (membership tracking) - implied usage

## Features
- Avatar support in message responses
- Read/unread status tracking
- Message editing with timestamp tracking
- Soft delete for message removal
- Room-based messaging
- Conversation summary with last activity

## Error Patterns
- 401: User not authenticated
- 403: Not authorized (message ownership, room access)
- 404: User/message/room not found
- 409: Room name conflicts
- 500: Database operations