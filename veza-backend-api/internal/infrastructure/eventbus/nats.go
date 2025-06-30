package eventbus

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

// EventBus interface pour event bus
type EventBus interface {
	Connect(url string) error
	Close() error
	PublishChatEvent(event ChatEvent) error
	PublishStreamEvent(event StreamEvent) error
	PublishUserEvent(event UserEvent) error
	SubscribeChatEvents(handler ChatEventHandler) error
	SubscribeStreamEvents(handler StreamEventHandler) error
	SubscribeUserEvents(handler UserEventHandler) error
	GetStatus() EventBusStatus
}

// NATSEventBus implémentation NATS de l'event bus
type NATSEventBus struct {
	nc          *nats.Conn
	js          nats.JetStreamContext
	logger      *zap.Logger
	config      *EventBusConfig
	isConnected bool
}

// EventBusConfig configuration de l'event bus
type EventBusConfig struct {
	URL                string        `json:"url"`
	ClusterID          string        `json:"cluster_id"`
	ClientID           string        `json:"client_id"`
	ConnectTimeout     time.Duration `json:"connect_timeout"`
	MaxReconnect       int           `json:"max_reconnect"`
	ReconnectWait      time.Duration `json:"reconnect_wait"`
	MaxPubAcksInflight int           `json:"max_pub_acks_inflight"`
}

// EventType type d'événement
type EventType string

const (
	EventTypeChatMessage     EventType = "chat.message"
	EventTypeChatRoomCreated EventType = "chat.room.created"
	EventTypeChatRoomJoined  EventType = "chat.room.joined"
	EventTypeChatUserMuted   EventType = "chat.user.muted"
	EventTypeStreamStarted   EventType = "stream.started"
	EventTypeStreamStopped   EventType = "stream.stopped"
	EventTypeStreamJoined    EventType = "stream.joined"
	EventTypeStreamLeft      EventType = "stream.left"
	EventTypeUserRegistered  EventType = "user.registered"
	EventTypeUserLoggedIn    EventType = "user.logged_in"
	EventTypeUserLoggedOut   EventType = "user.logged_out"
	EventTypeUserUpdated     EventType = "user.updated"
)

// BaseEvent structure de base pour tous les événements
type BaseEvent struct {
	Type      EventType   `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	UserID    int64       `json:"user_id"`
	Data      interface{} `json:"data"`
	Source    string      `json:"source"`
	TraceID   string      `json:"trace_id"`
}

// ChatEvent événements liés au chat
type ChatEvent struct {
	BaseEvent
	RoomID    string `json:"room_id,omitempty"`
	MessageID string `json:"message_id,omitempty"`
}

// StreamEvent événements liés au streaming
type StreamEvent struct {
	BaseEvent
	StreamID   string `json:"stream_id,omitempty"`
	Quality    string `json:"quality,omitempty"`
	ListenerID int64  `json:"listener_id,omitempty"`
}

// UserEvent événements liés aux utilisateurs
type UserEvent struct {
	BaseEvent
	Email  string `json:"email,omitempty"`
	Role   string `json:"role,omitempty"`
	Status string `json:"status,omitempty"`
}

// EventHandlers
type ChatEventHandler func(event ChatEvent) error
type StreamEventHandler func(event StreamEvent) error
type UserEventHandler func(event UserEvent) error

// EventBusStatus status de l'event bus
type EventBusStatus struct {
	Connected         bool      `json:"connected"`
	LastConnected     time.Time `json:"last_connected"`
	MessagesPublished int64     `json:"messages_published"`
	MessagesReceived  int64     `json:"messages_received"`
	Errors            int64     `json:"errors"`
}

// NewNATSEventBus crée un nouveau NATS EventBus
func NewNATSEventBus(config *EventBusConfig, logger *zap.Logger) *NATSEventBus {
	if logger == nil {
		logger, _ = zap.NewProduction()
	}

	return &NATSEventBus{
		config: config,
		logger: logger,
	}
}

// Connect se connecte au serveur NATS
func (e *NATSEventBus) Connect(url string) error {
	if url == "" {
		url = e.config.URL
	}
	if url == "" {
		url = nats.DefaultURL
	}

	// Options de connexion
	opts := []nats.Option{
		nats.Name(e.config.ClientID),
		nats.Timeout(e.config.ConnectTimeout),
		nats.MaxReconnects(e.config.MaxReconnect),
		nats.ReconnectWait(e.config.ReconnectWait),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			e.logger.Warn("NATS disconnected", zap.Error(err))
			e.isConnected = false
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			e.logger.Info("NATS reconnected", zap.String("url", nc.ConnectedUrl()))
			e.isConnected = true
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			e.logger.Info("NATS connection closed")
			e.isConnected = false
		}),
	}

	// Connexion
	nc, err := nats.Connect(url, opts...)
	if err != nil {
		e.logger.Error("Failed to connect to NATS", zap.Error(err), zap.String("url", url))
		return fmt.Errorf("failed to connect to NATS: %w", err)
	}

	e.nc = nc
	e.isConnected = true

	// Créer JetStream context
	js, err := nc.JetStream()
	if err != nil {
		e.logger.Error("Failed to create JetStream context", zap.Error(err))
		return fmt.Errorf("failed to create JetStream context: %w", err)
	}
	e.js = js

	// Créer les streams nécessaires
	if err := e.createStreams(); err != nil {
		e.logger.Error("Failed to create streams", zap.Error(err))
		return fmt.Errorf("failed to create streams: %w", err)
	}

	e.logger.Info("Successfully connected to NATS", zap.String("url", url))
	return nil
}

// createStreams crée les streams JetStream nécessaires
func (e *NATSEventBus) createStreams() error {
	streams := []nats.StreamConfig{
		{
			Name:      "CHAT_EVENTS",
			Subjects:  []string{"chat.events.>"},
			Storage:   nats.FileStorage,
			Retention: nats.LimitsPolicy,
			MaxAge:    24 * time.Hour, // Retention 24h
		},
		{
			Name:      "STREAM_EVENTS",
			Subjects:  []string{"stream.events.>"},
			Storage:   nats.FileStorage,
			Retention: nats.LimitsPolicy,
			MaxAge:    24 * time.Hour,
		},
		{
			Name:      "USER_EVENTS",
			Subjects:  []string{"user.events.>"},
			Storage:   nats.FileStorage,
			Retention: nats.LimitsPolicy,
			MaxAge:    7 * 24 * time.Hour, // Retention 7j pour audit
		},
	}

	for _, streamConfig := range streams {
		// Vérifier si le stream existe déjà
		_, err := e.js.StreamInfo(streamConfig.Name)
		if err == nil {
			e.logger.Debug("Stream already exists", zap.String("stream", streamConfig.Name))
			continue
		}

		// Créer le stream
		_, err = e.js.AddStream(&streamConfig)
		if err != nil {
			return fmt.Errorf("failed to create stream %s: %w", streamConfig.Name, err)
		}

		e.logger.Info("Created stream", zap.String("stream", streamConfig.Name))
	}

	return nil
}

// Close ferme la connexion NATS
func (e *NATSEventBus) Close() error {
	if e.nc != nil && e.isConnected {
		e.nc.Close()
		e.isConnected = false
		e.logger.Info("NATS connection closed")
	}
	return nil
}

// PublishChatEvent publie un événement chat
func (e *NATSEventBus) PublishChatEvent(event ChatEvent) error {
	if !e.isConnected {
		return fmt.Errorf("NATS not connected")
	}

	event.Timestamp = time.Now()
	event.Source = "veza-backend"

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal chat event: %w", err)
	}

	subject := fmt.Sprintf("chat.events.%s", event.Type)
	_, err = e.js.Publish(subject, data)
	if err != nil {
		e.logger.Error("Failed to publish chat event", zap.Error(err), zap.String("subject", subject))
		return fmt.Errorf("failed to publish chat event: %w", err)
	}

	e.logger.Debug("Published chat event", zap.String("type", string(event.Type)), zap.String("subject", subject))
	return nil
}

// PublishStreamEvent publie un événement stream
func (e *NATSEventBus) PublishStreamEvent(event StreamEvent) error {
	if !e.isConnected {
		return fmt.Errorf("NATS not connected")
	}

	event.Timestamp = time.Now()
	event.Source = "veza-backend"

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal stream event: %w", err)
	}

	subject := fmt.Sprintf("stream.events.%s", event.Type)
	_, err = e.js.Publish(subject, data)
	if err != nil {
		e.logger.Error("Failed to publish stream event", zap.Error(err), zap.String("subject", subject))
		return fmt.Errorf("failed to publish stream event: %w", err)
	}

	e.logger.Debug("Published stream event", zap.String("type", string(event.Type)), zap.String("subject", subject))
	return nil
}

// PublishUserEvent publie un événement utilisateur
func (e *NATSEventBus) PublishUserEvent(event UserEvent) error {
	if !e.isConnected {
		return fmt.Errorf("NATS not connected")
	}

	event.Timestamp = time.Now()
	event.Source = "veza-backend"

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal user event: %w", err)
	}

	subject := fmt.Sprintf("user.events.%s", event.Type)
	_, err = e.js.Publish(subject, data)
	if err != nil {
		e.logger.Error("Failed to publish user event", zap.Error(err), zap.String("subject", subject))
		return fmt.Errorf("failed to publish user event: %w", err)
	}

	e.logger.Debug("Published user event", zap.String("type", string(event.Type)), zap.String("subject", subject))
	return nil
}

// SubscribeChatEvents s'abonne aux événements chat
func (e *NATSEventBus) SubscribeChatEvents(handler ChatEventHandler) error {
	_, err := e.js.Subscribe("chat.events.>", func(msg *nats.Msg) {
		var event ChatEvent
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			e.logger.Error("Failed to unmarshal chat event", zap.Error(err))
			return
		}

		if err := handler(event); err != nil {
			e.logger.Error("Chat event handler failed", zap.Error(err), zap.String("type", string(event.Type)))
		} else {
			msg.Ack()
		}
	}, nats.Durable("chat-events-consumer"))

	return err
}

// SubscribeStreamEvents s'abonne aux événements stream
func (e *NATSEventBus) SubscribeStreamEvents(handler StreamEventHandler) error {
	_, err := e.js.Subscribe("stream.events.>", func(msg *nats.Msg) {
		var event StreamEvent
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			e.logger.Error("Failed to unmarshal stream event", zap.Error(err))
			return
		}

		if err := handler(event); err != nil {
			e.logger.Error("Stream event handler failed", zap.Error(err), zap.String("type", string(event.Type)))
		} else {
			msg.Ack()
		}
	}, nats.Durable("stream-events-consumer"))

	return err
}

// SubscribeUserEvents s'abonne aux événements utilisateur
func (e *NATSEventBus) SubscribeUserEvents(handler UserEventHandler) error {
	_, err := e.js.Subscribe("user.events.>", func(msg *nats.Msg) {
		var event UserEvent
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			e.logger.Error("Failed to unmarshal user event", zap.Error(err))
			return
		}

		if err := handler(event); err != nil {
			e.logger.Error("User event handler failed", zap.Error(err), zap.String("type", string(event.Type)))
		} else {
			msg.Ack()
		}
	}, nats.Durable("user-events-consumer"))

	return err
}

// GetStatus retourne le status de l'event bus
func (e *NATSEventBus) GetStatus() EventBusStatus {
	status := EventBusStatus{
		Connected: e.isConnected,
	}

	if e.nc != nil {
		stats := e.nc.Stats()
		status.MessagesPublished = int64(stats.OutMsgs)
		status.MessagesReceived = int64(stats.InMsgs)
	}

	return status
}

// GetDefaultConfig retourne une configuration par défaut
func GetDefaultConfig() *EventBusConfig {
	return &EventBusConfig{
		URL:                nats.DefaultURL,
		ClusterID:          "veza-cluster",
		ClientID:           "veza-backend",
		ConnectTimeout:     10 * time.Second,
		MaxReconnect:       -1, // Reconnexion infinie
		ReconnectWait:      2 * time.Second,
		MaxPubAcksInflight: 1000,
	}
}
