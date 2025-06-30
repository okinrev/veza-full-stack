package grpc

import (
	"context"
)

// ===================================
// INTERFACES TEMPORAIRES PROTOBUF
// ===================================

// AuthServiceClient interface pour le service d'authentification
type AuthServiceClient interface {
	ValidateToken(ctx context.Context, req *ValidateTokenRequest) (*ValidateTokenResponse, error)
}

// ChatServiceClient interface pour le service de chat
type ChatServiceClient interface {
	CreateRoom(ctx context.Context, req *CreateRoomRequest) (*CreateRoomResponse, error)
	SendMessage(ctx context.Context, req *SendMessageRequest) (*SendMessageResponse, error)
}

// StreamServiceClient interface pour le service de streaming
type StreamServiceClient interface {
	CreateStream(ctx context.Context, req *CreateStreamRequest) (*CreateStreamResponse, error)
	StartStream(ctx context.Context, req *StartStreamRequest) (*StartStreamResponse, error)
}

// Structures de base (temporaires)
type ValidateTokenRequest struct {
	Token   string `json:"token"`
	Service string `json:"service"`
}

type ValidateTokenResponse struct {
	Valid bool   `json:"valid"`
	Error string `json:"error,omitempty"`
}

type CreateRoomRequest struct {
	Name      string `json:"name"`
	CreatedBy int64  `json:"created_by"`
	AuthToken string `json:"auth_token"`
}

type CreateRoomResponse struct {
	RoomId string `json:"room_id,omitempty"`
	Error  string `json:"error,omitempty"`
}

type SendMessageRequest struct {
	RoomId    string `json:"room_id"`
	UserId    int64  `json:"user_id"`
	Content   string `json:"content"`
	AuthToken string `json:"auth_token"`
}

type SendMessageResponse struct {
	MessageId string `json:"message_id,omitempty"`
	Error     string `json:"error,omitempty"`
}

type CreateStreamRequest struct {
	Title     string `json:"title"`
	CreatedBy int64  `json:"created_by"`
	AuthToken string `json:"auth_token"`
}

type CreateStreamResponse struct {
	StreamId string `json:"stream_id,omitempty"`
	Error    string `json:"error,omitempty"`
}

type StartStreamRequest struct {
	StreamId  string `json:"stream_id"`
	AuthToken string `json:"auth_token"`
}

type StartStreamResponse struct {
	Success   bool   `json:"success"`
	StreamUrl string `json:"stream_url,omitempty"`
	Error     string `json:"error,omitempty"`
}
