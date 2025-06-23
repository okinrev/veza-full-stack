# RoomHandler Documentation

This file refer to : backend/internal/handlers/room.go

## Package & Dependencies
```go
package handlers
imports: gin, database, middleware, strconv, strings, fmt, net/http
```

## Struct
```go
type RoomHandler struct {
    db *database.DB
}
```

## Constructor
```go
func NewRoomHandler(db *database.DB) *RoomHandler
```

## Types/Models
```go
type RoomResponse struct {
    ID, MemberCount, OnlineCount, UnreadCount int
    Name, CreatorName, UserRole string
    Description *string
    IsPrivate, IsMember bool
    CreatorID *int
    LastActivity, LastMessage *string
    CreatedAt, UpdatedAt string
}

type CreateRoomRequest struct {
    Name string `binding:"required,min=1,max=50"`
    Description *string
    IsPrivate bool
    Password *string
}

type UpdateRoomRequest struct {
    Name, Description *string
    IsPrivate *bool
    Password *string
}

type RoomMemberResponse struct {
    UserID, MessageCount int
    Username, Role, JoinedAt string
    Avatar, LastSeen *string
    IsOnline bool
}

type RoomMessageResponse struct {
    ID, RoomID, UserID int
    Username, Content, MessageType, CreatedAt string
    Avatar *string
    EditedAt *string
}

type JoinRoomRequest struct {
    Password *string
}
```

## Methods

### Room Management
```go
func (h *RoomHandler) GetPublicRooms(c *gin.Context)
func (h *RoomHandler) CreateRoom(c *gin.Context)
func (h *RoomHandler) GetRoom(c *gin.Context)
func (h *RoomHandler) JoinRoom(c *gin.Context)
func (h *RoomHandler) LeaveRoom(c *gin.Context)
```

### Room Content
```go
func (h *RoomHandler) GetRoomMembers(c *gin.Context)
func (h *RoomHandler) GetRoomMessages(c *gin.Context)
func (h *RoomHandler) SendRoomMessage(c *gin.Context)
```

### GetPublicRooms
- **Auth**: None
- **Params**: page, limit, search
- **Process**: Public rooms with member counts, last activity
- **Response**: Paginated rooms ordered by activity

### CreateRoom
- **Auth**: common.GetUserIDFromContext required
- **Validation**: Name uniqueness
- **Process**: Create room + add creator as owner + system message
- **Features**: Password protection (plain text - NOT SECURE)

### JoinRoom/LeaveRoom
- **Auth**: Required
- **Validation**: Room existence, password check, membership status
- **Process**: Add/remove from room_members, system messages
- **Restriction**: Owners cannot leave (must transfer/delete)

### GetRoomMessages/SendRoomMessage
- **Auth**: Required, membership validation
- **Process**: Paginated message history, message posting
- **Features**: Message types (message, join, leave, system)

## Helper Methods
```go
func (h *RoomHandler) getRoomByID(roomID, userID int) (*RoomResponse, error)
func (h *RoomHandler) hasRoomAccess(roomID, userID int) bool
func (h *RoomHandler) addSystemMessage(roomID int, content string)
func getUsernameByID(db *database.DB, userID int) string
```

## Route Mapping Expectations
- `GET /rooms` → GetPublicRooms
- `POST /rooms` → CreateRoom
- `GET /rooms/:id` → GetRoom
- `POST /rooms/:id/join` → JoinRoom
- `POST /rooms/:id/leave` → LeaveRoom
- `GET /rooms/:id/members` → GetRoomMembers
- `GET /rooms/:id/messages` → GetRoomMessages
- `POST /rooms/:id/messages` → SendRoomMessage

## Middleware Dependencies
- Authentication middleware for most endpoints
- JSON binding for POST requests

## Database Tables
- rooms (room metadata, passwords)
- room_members (membership with roles)
- room_messages (messages with types)
- users (member details)

## Member Roles
- `owner`: Room creator, cannot leave
- `admin`: Elevated permissions (not fully implemented)
- `member`: Regular member

## Message Types
- `message`: Regular user message
- `join`: User joined room
- `leave`: User left room
- `system`: System announcements

## Access Control
- Public rooms: Anyone can view, members can message
- Private rooms: Members only for all operations
- Password protection: Optional additional security

## Error Patterns
- 401: User not authenticated
- 403: Access denied (room access, ownership)
- 404: Room not found
- 409: Room name conflicts, already member
- 500: Database errors