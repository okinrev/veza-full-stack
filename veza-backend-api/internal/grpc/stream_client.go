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

// StreamClient client gRPC pour le service de streaming
type StreamClient struct {
	conn        *grpc.ClientConn
	client      StreamServiceClient
	config      *ClientConfig
	logger      *zap.Logger
	isConnected bool
	mu          sync.RWMutex
}

// NewStreamClient crée un nouveau client gRPC pour le streaming
func NewStreamClient(config *ClientConfig, logger *zap.Logger) *StreamClient {
	if logger == nil {
		logger, _ = zap.NewProduction()
	}

	if config == nil {
		config = &ClientConfig{
			Address:        "localhost:50052", // Port par défaut du Stream Server Rust
			Timeout:        30 * time.Second,
			MaxRetries:     3,
			RetryDelay:     1 * time.Second,
			KeepAlive:      30 * time.Second,
			ConnectTimeout: 10 * time.Second,
			EnableTLS:      false, // Développement sans TLS
		}
	}

	return &StreamClient{
		config: config,
		logger: logger,
	}
}

// Connect établit la connexion gRPC vers le Stream Server
func (s *StreamClient) Connect() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isConnected && s.conn != nil {
		return nil
	}

	// Options de connexion gRPC
	opts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithTimeout(s.config.ConnectTimeout),
	}

	// TLS ou mode insecure
	if s.config.EnableTLS {
		s.logger.Info("TLS enabled for gRPC connection")
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	s.logger.Info("Connecting to Stream Server", zap.String("address", s.config.Address))

	// Connexion
	conn, err := grpc.Dial(s.config.Address, opts...)
	if err != nil {
		s.logger.Error("Failed to connect to Stream Server", zap.Error(err))
		return fmt.Errorf("failed to connect to stream server: %w", err)
	}

	s.conn = conn
	s.isConnected = true

	s.logger.Info("Successfully connected to Stream Server")
	return nil
}

// Close ferme la connexion gRPC
func (s *StreamClient) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.conn != nil {
		err := s.conn.Close()
		s.conn = nil
		s.isConnected = false
		s.logger.Info("Stream client connection closed")
		return err
	}
	return nil
}

// IsConnected vérifie si la connexion est active
func (s *StreamClient) IsConnected() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.isConnected || s.conn == nil {
		return false
	}

	state := s.conn.GetState()
	return state == connectivity.Ready || state == connectivity.Idle
}

// Health vérifie la santé de la connexion
func (s *StreamClient) Health() error {
	if !s.IsConnected() {
		return fmt.Errorf("stream client not connected")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test simple avec une requête de validation
	_, err := s.ValidateConnection(ctx)
	return err
}

// ValidateConnection valide la connexion avec une requête de test
func (s *StreamClient) ValidateConnection(ctx context.Context) (bool, error) {
	if !s.IsConnected() {
		return false, fmt.Errorf("not connected")
	}

	// TODO: Implémenter un ping/health check vers le Stream Server
	return true, nil
}

// CreateStream crée un nouveau stream audio
func (s *StreamClient) CreateStream(ctx context.Context, req *CreateStreamRequest) (*CreateStreamResponse, error) {
	if !s.IsConnected() {
		if err := s.Connect(); err != nil {
			return nil, err
		}
	}

	s.logger.Debug("Creating stream", zap.String("title", req.Title), zap.Int64("created_by", req.CreatedBy))

	// TODO: Remplacer par l'appel gRPC réel une fois les protobuf générés
	response := &CreateStreamResponse{
		StreamId: fmt.Sprintf("stream_%d_%s", time.Now().Unix(), req.Title),
	}

	s.logger.Info("Stream created successfully", zap.String("stream_id", response.StreamId))
	return response, nil
}

// StartStream démarre un stream audio
func (s *StreamClient) StartStream(ctx context.Context, req *StartStreamRequest) (*StartStreamResponse, error) {
	if !s.IsConnected() {
		if err := s.Connect(); err != nil {
			return nil, err
		}
	}

	s.logger.Debug("Starting stream", zap.String("stream_id", req.StreamId))

	// TODO: Remplacer par l'appel gRPC réel une fois les protobuf générés
	response := &StartStreamResponse{
		Success:   true,
		StreamUrl: fmt.Sprintf("ws://localhost:8081/stream/%s", req.StreamId),
	}

	s.logger.Info("Stream started successfully",
		zap.String("stream_id", req.StreamId),
		zap.String("stream_url", response.StreamUrl))
	return response, nil
}

// WithRetry exécute une fonction avec retry automatique
func (s *StreamClient) WithRetry(ctx context.Context, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt <= s.config.MaxRetries; attempt++ {
		if attempt > 0 {
			s.logger.Warn("Retrying gRPC call",
				zap.Int("attempt", attempt),
				zap.Error(lastErr))

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(s.config.RetryDelay * time.Duration(attempt)):
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
func (s *StreamClient) GetStatus() ClientStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := ClientStatus{
		ServiceName: "stream",
		Address:     s.config.Address,
		Connected:   s.isConnected,
	}

	if s.conn != nil {
		status.ConnectionState = s.conn.GetState().String()
	}

	return status
}

// Reconnect tente de reconnecter le client
func (s *StreamClient) Reconnect() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.conn != nil {
		s.conn.Close()
	}

	return s.Connect()
}
