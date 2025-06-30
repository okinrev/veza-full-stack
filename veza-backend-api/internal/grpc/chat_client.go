package grpc

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

// ChatClient client gRPC pour le service de chat
type ChatClient struct {
	conn        *grpc.ClientConn
	client      ChatServiceClient
	config      *ClientConfig
	logger      *zap.Logger
	isConnected bool
	mu          sync.RWMutex
}

// ClientConfig configuration pour les clients gRPC
type ClientConfig struct {
	Address        string        `json:"address"`
	Timeout        time.Duration `json:"timeout"`
	MaxRetries     int           `json:"max_retries"`
	RetryDelay     time.Duration `json:"retry_delay"`
	KeepAlive      time.Duration `json:"keep_alive"`
	ConnectTimeout time.Duration `json:"connect_timeout"`
	EnableTLS      bool          `json:"enable_tls"`
}

// NewChatClient crée un nouveau client gRPC pour le chat
func NewChatClient(config *ClientConfig, logger *zap.Logger) *ChatClient {
	if logger == nil {
		logger, _ = zap.NewProduction()
	}

	if config == nil {
		config = &ClientConfig{
			Address:        "localhost:50051", // Port par défaut du Chat Server Rust
			Timeout:        30 * time.Second,
			MaxRetries:     3,
			RetryDelay:     1 * time.Second,
			KeepAlive:      30 * time.Second,
			ConnectTimeout: 10 * time.Second,
			EnableTLS:      false, // Développement sans TLS
		}
	}

	return &ChatClient{
		config: config,
		logger: logger,
	}
}

// Connect établit la connexion gRPC vers le Chat Server
func (c *ChatClient) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isConnected && c.conn != nil {
		return nil
	}

	// Options de connexion gRPC
	opts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithTimeout(c.config.ConnectTimeout),
	}

	// TLS ou mode insecure
	if c.config.EnableTLS {
		// TODO: Ajouter les credentials TLS en production
		c.logger.Info("TLS enabled for gRPC connection")
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// Keep alive (supprimé temporairement - à implémenter avec protobuf complet)

	c.logger.Info("Connecting to Chat Server", zap.String("address", c.config.Address))

	// Connexion
	conn, err := grpc.Dial(c.config.Address, opts...)
	if err != nil {
		c.logger.Error("Failed to connect to Chat Server", zap.Error(err))
		return fmt.Errorf("failed to connect to chat server: %w", err)
	}

	c.conn = conn
	c.isConnected = true

	c.logger.Info("Successfully connected to Chat Server")
	return nil
}

// Close ferme la connexion gRPC
func (c *ChatClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		c.isConnected = false
		c.logger.Info("Chat client connection closed")
		return err
	}
	return nil
}

// IsConnected vérifie si la connexion est active
func (c *ChatClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.isConnected || c.conn == nil {
		return false
	}

	state := c.conn.GetState()
	return state == connectivity.Ready || state == connectivity.Idle
}

// Health vérifie la santé de la connexion
func (c *ChatClient) Health() error {
	if !c.IsConnected() {
		return fmt.Errorf("chat client not connected")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test simple avec une requête de validation
	_, err := c.ValidateConnection(ctx)
	return err
}

// ValidateConnection valide la connexion avec une requête de test
func (c *ChatClient) ValidateConnection(ctx context.Context) (bool, error) {
	if !c.IsConnected() {
		return false, fmt.Errorf("not connected")
	}

	// TODO: Implémenter un ping/health check vers le Chat Server
	// Pour l'instant, on retourne simplement l'état de la connexion
	return true, nil
}

// CreateRoom crée une nouvelle salle de chat
func (c *ChatClient) CreateRoom(ctx context.Context, req *CreateRoomRequest) (*CreateRoomResponse, error) {
	if !c.IsConnected() {
		if err := c.Connect(); err != nil {
			return nil, err
		}
	}

	c.logger.Debug("Creating room", zap.String("name", req.Name), zap.Int64("created_by", req.CreatedBy))

	// TODO: Remplacer par l'appel gRPC réel une fois les protobuf générés
	// Pour l'instant, simulation d'une réponse
	response := &CreateRoomResponse{
		RoomId: fmt.Sprintf("room_%d_%s", time.Now().Unix(), req.Name),
	}

	c.logger.Info("Room created successfully", zap.String("room_id", response.RoomId))
	return response, nil
}

// SendMessage envoie un message dans une salle
func (c *ChatClient) SendMessage(ctx context.Context, req *SendMessageRequest) (*SendMessageResponse, error) {
	if !c.IsConnected() {
		if err := c.Connect(); err != nil {
			return nil, err
		}
	}

	c.logger.Debug("Sending message",
		zap.String("room_id", req.RoomId),
		zap.Int64("user_id", req.UserId),
		zap.String("content", req.Content))

	// TODO: Remplacer par l'appel gRPC réel une fois les protobuf générés
	response := &SendMessageResponse{
		MessageId: fmt.Sprintf("msg_%d", time.Now().UnixNano()),
	}

	c.logger.Debug("Message sent successfully", zap.String("message_id", response.MessageId))
	return response, nil
}

// WithRetry exécute une fonction avec retry automatique
func (c *ChatClient) WithRetry(ctx context.Context, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			c.logger.Warn("Retrying gRPC call",
				zap.Int("attempt", attempt),
				zap.Error(lastErr))

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(c.config.RetryDelay * time.Duration(attempt)):
			}
		}

		if err := fn(); err != nil {
			lastErr = err
			continue
		}

		return nil
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

// GetStatus retourne le statut du client
func (c *ChatClient) GetStatus() ClientStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()

	status := ClientStatus{
		ServiceName: "chat",
		Address:     c.config.Address,
		Connected:   c.isConnected,
	}

	if c.conn != nil {
		status.ConnectionState = c.conn.GetState().String()
	}

	return status
}

// ClientStatus représente le statut d'un client gRPC
type ClientStatus struct {
	ServiceName     string `json:"service_name"`
	Address         string `json:"address"`
	Connected       bool   `json:"connected"`
	ConnectionState string `json:"connection_state"`
	LastError       string `json:"last_error,omitempty"`
}

// Reconnect tente de reconnecter le client
func (c *ChatClient) Reconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		c.conn.Close()
	}

	return c.Connect()
}
