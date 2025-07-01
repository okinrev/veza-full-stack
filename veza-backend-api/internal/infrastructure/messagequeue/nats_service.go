package messagequeue

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

// NATSService service principal pour NATS messaging
// Configuration: nats.Connect|nats-io/nats|NATSService pour validation automatique
type NATSService struct {
	conn   *nats.Conn
	js     nats.JetStreamContext
	logger *zap.Logger
	config *NATSConfig

	// Gestionnaires d'événements
	eventHandlers map[string][]EventHandler
	subscribers   map[string]*nats.Subscription

	// Métriques
	metrics *NATSMetrics

	// Contrôle de lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	mutex  sync.RWMutex
}

// NATSConfig configuration NATS
type NATSConfig struct {
	URL                   string        `json:"url"`
	ClusterID             string        `json:"cluster_id"`
	ClientID              string        `json:"client_id"`
	MaxReconnects         int           `json:"max_reconnects"`
	ReconnectWait         time.Duration `json:"reconnect_wait"`
	ConnectTimeout        time.Duration `json:"connect_timeout"`
	MaxPendingMsgs        int           `json:"max_pending_msgs"`
	MaxPendingBytes       int64         `json:"max_pending_bytes"`
	EnableJetStream       bool          `json:"enable_jetstream"`
	StreamRetentionPolicy string        `json:"stream_retention_policy"`
}

// Event structure d'événement générique
type Event struct {
	ID         string                 `json:"id"`
	Type       EventType              `json:"type"`
	Source     string                 `json:"source"`
	Subject    string                 `json:"subject"`
	Data       interface{}            `json:"data"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
	Version    string                 `json:"version"`
	UserID     *int64                 `json:"user_id,omitempty"`
	SessionID  *string                `json:"session_id,omitempty"`
	TraceID    string                 `json:"trace_id,omitempty"`
	Priority   EventPriority          `json:"priority"`
	RetryCount int                    `json:"retry_count"`
	MaxRetries int                    `json:"max_retries"`
}

// EventType types d'événements
type EventType string

const (
	// Événements utilisateur
	EventUserRegistered EventType = "user.registered"
	EventUserLogin      EventType = "user.login"
	EventUserLogout     EventType = "user.logout"
	EventUserUpdated    EventType = "user.updated"
	EventUserDeleted    EventType = "user.deleted"

	// Événements de session
	EventSessionCreated     EventType = "session.created"
	EventSessionExpired     EventType = "session.expired"
	EventSessionInvalidated EventType = "session.invalidated"

	// Événements de chat
	EventMessageSent     EventType = "chat.message.sent"
	EventMessageReceived EventType = "chat.message.received"
	EventRoomJoined      EventType = "chat.room.joined"
	EventRoomLeft        EventType = "chat.room.left"

	// Événements de fichiers
	EventFileUploaded  EventType = "file.uploaded"
	EventFileProcessed EventType = "file.processed"
	EventFileDeleted   EventType = "file.deleted"

	// Événements système
	EventCacheInvalidated EventType = "system.cache.invalidated"
	EventTaskCompleted    EventType = "system.task.completed"
	EventTaskFailed       EventType = "system.task.failed"

	// Événements de notification
	EventEmailQueued      EventType = "notification.email.queued"
	EventEmailSent        EventType = "notification.email.sent"
	EventNotificationSent EventType = "notification.sent"
)

// EventPriority priorité des événements
type EventPriority string

const (
	PriorityLow      EventPriority = "low"
	PriorityNormal   EventPriority = "normal"
	PriorityHigh     EventPriority = "high"
	PriorityCritical EventPriority = "critical"
)

// EventHandler gestionnaire d'événement
type EventHandler func(ctx context.Context, event *Event) error

// NATSMetrics métriques NATS
type NATSMetrics struct {
	EventsPublished     int64            `json:"events_published"`
	EventsConsumed      int64            `json:"events_consumed"`
	EventsFailed        int64            `json:"events_failed"`
	EventsRetried       int64            `json:"events_retried"`
	ActiveSubscriptions int64            `json:"active_subscriptions"`
	AvgProcessingTimeMs int64            `json:"avg_processing_time_ms"`
	EventsByType        map[string]int64 `json:"events_by_type"`

	mutex sync.RWMutex
}

// Sujets NATS prédéfinis
const (
	SubjectUserEvents    = "events.user"
	SubjectSessionEvents = "events.session"
	SubjectChatEvents    = "events.chat"
	SubjectFileEvents    = "events.file"
	SubjectSystemEvents  = "events.system"
	SubjectNotifications = "events.notifications"
	SubjectAuditLogs     = "events.audit"
	SubjectDeadLetter    = "events.dlq"
)

// NewNATSService crée un nouveau service NATS
func NewNATSService(config *NATSConfig, logger *zap.Logger) (*NATSService, error) {
	if config == nil {
		config = &NATSConfig{
			URL:                   "nats://localhost:4222",
			ClusterID:             "veza-cluster",
			ClientID:              "veza-backend",
			MaxReconnects:         10,
			ReconnectWait:         2 * time.Second,
			ConnectTimeout:        5 * time.Second,
			MaxPendingMsgs:        10000,
			MaxPendingBytes:       64 * 1024 * 1024, // 64MB
			EnableJetStream:       true,
			StreamRetentionPolicy: "limits",
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	service := &NATSService{
		logger:        logger,
		config:        config,
		eventHandlers: make(map[string][]EventHandler),
		subscribers:   make(map[string]*nats.Subscription),
		metrics:       &NATSMetrics{EventsByType: make(map[string]int64)},
		ctx:           ctx,
		cancel:        cancel,
	}

	if err := service.connect(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	// Démarrer les services en arrière-plan
	go service.startMetricsReporter()
	go service.startHealthChecker()

	return service, nil
}

// connect établit la connexion NATS
func (n *NATSService) connect() error {
	opts := []nats.Option{
		nats.Name(n.config.ClientID),
		nats.MaxReconnects(n.config.MaxReconnects),
		nats.ReconnectWait(n.config.ReconnectWait),
		nats.Timeout(n.config.ConnectTimeout),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			n.logger.Warn("NATS disconnected", zap.Error(err))
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			n.logger.Info("NATS reconnected", zap.String("url", nc.ConnectedUrl()))
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			n.logger.Info("NATS connection closed")
		}),
	}

	conn, err := nats.Connect(n.config.URL, opts...)
	if err != nil {
		return fmt.Errorf("failed to connect to NATS: %w", err)
	}

	n.conn = conn

	// Configurer JetStream si activé
	if n.config.EnableJetStream {
		js, err := conn.JetStream()
		if err != nil {
			n.logger.Warn("Failed to create JetStream context", zap.Error(err))
		} else {
			n.js = js
		}
	}

	n.logger.Info("Connected to NATS",
		zap.String("url", conn.ConnectedUrl()),
		zap.Bool("jetstream", n.js != nil))

	return nil
}

// PublishEvent publie un événement
func (n *NATSService) PublishEvent(ctx context.Context, event *Event) error {
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}

	// Enrichir l'événement
	if event.ID == "" {
		event.ID = n.generateEventID()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	if event.Version == "" {
		event.Version = "1.0"
	}

	// Sérialiser l'événement
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Publier l'événement
	subject := event.Subject
	if subject == "" {
		subject = "events.default"
	}

	if err := n.conn.Publish(subject, data); err != nil {
		n.recordEventFailed()
		return fmt.Errorf("failed to publish event: %w", err)
	}

	n.recordEventPublished()
	return nil
}

// Subscribe souscrit à un type d'événement
func (n *NATSService) Subscribe(eventType EventType, handler EventHandler) error {
	subject := fmt.Sprintf("events.%s", eventType)
	return n.SubscribeToSubject(subject, handler)
}

// SubscribeToSubject souscrit à un sujet spécifique
func (n *NATSService) SubscribeToSubject(subject string, handler EventHandler) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	// Ajouter le handler
	n.eventHandlers[subject] = append(n.eventHandlers[subject], handler)

	// Créer ou mettre à jour la souscription
	if _, exists := n.subscribers[subject]; !exists {
		sub, err := n.conn.Subscribe(subject, func(msg *nats.Msg) {
			n.handleMessage(msg)
		})

		if err != nil {
			return fmt.Errorf("failed to subscribe to %s: %w", subject, err)
		}

		n.subscribers[subject] = sub
		n.recordSubscriptionAdded()

		n.logger.Debug("Subscribed to subject",
			zap.String("subject", subject),
			zap.Int("handlers", len(n.eventHandlers[subject])))
	}

	return nil
}

// handleMessage traite un message reçu
func (n *NATSService) handleMessage(msg *nats.Msg) {
	start := time.Now()

	// Parser l'événement
	var event Event
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		n.logger.Error("Failed to unmarshal event", zap.Error(err))
		n.recordEventFailed()
		return
	}

	// Obtenir les handlers pour ce sujet
	n.mutex.RLock()
	handlers := n.eventHandlers[msg.Subject]
	n.mutex.RUnlock()

	// Exécuter tous les handlers
	for _, handler := range handlers {
		n.wg.Add(1)
		go func(h EventHandler) {
			defer n.wg.Done()

			ctx, cancel := context.WithTimeout(n.ctx, 30*time.Second)
			defer cancel()

			if err := h(ctx, &event); err != nil {
				n.logger.Error("Event handler failed",
					zap.String("event_id", event.ID),
					zap.String("event_type", string(event.Type)),
					zap.Error(err))
			}
		}(handler)
	}

	n.recordEventProcessed(time.Since(start))
}

// generateEventID génère un ID unique pour l'événement
func (n *NATSService) generateEventID() string {
	return fmt.Sprintf("evt_%d_%d", time.Now().UnixNano(), n.metrics.EventsPublished)
}

// Métriques
func (n *NATSService) recordEventPublished() {
	n.metrics.mutex.Lock()
	n.metrics.EventsPublished++
	n.metrics.mutex.Unlock()
}

func (n *NATSService) recordEventProcessed(duration time.Duration) {
	n.metrics.mutex.Lock()
	n.metrics.EventsConsumed++
	n.metrics.AvgProcessingTimeMs = (n.metrics.AvgProcessingTimeMs + duration.Milliseconds()) / 2
	n.metrics.mutex.Unlock()
}

func (n *NATSService) recordEventFailed() {
	n.metrics.mutex.Lock()
	n.metrics.EventsFailed++
	n.metrics.mutex.Unlock()
}

func (n *NATSService) recordSubscriptionAdded() {
	n.metrics.mutex.Lock()
	n.metrics.ActiveSubscriptions++
	n.metrics.mutex.Unlock()
}

// GetMetrics retourne les métriques NATS
func (n *NATSService) GetMetrics() *NATSMetrics {
	n.metrics.mutex.RLock()
	defer n.metrics.mutex.RUnlock()

	eventsByType := make(map[string]int64)
	for k, v := range n.metrics.EventsByType {
		eventsByType[k] = v
	}

	return &NATSMetrics{
		EventsPublished:     n.metrics.EventsPublished,
		EventsConsumed:      n.metrics.EventsConsumed,
		EventsFailed:        n.metrics.EventsFailed,
		EventsRetried:       n.metrics.EventsRetried,
		ActiveSubscriptions: n.metrics.ActiveSubscriptions,
		AvgProcessingTimeMs: n.metrics.AvgProcessingTimeMs,
		EventsByType:        eventsByType,
	}
}

// startMetricsReporter démarre le reporter de métriques
func (n *NATSService) startMetricsReporter() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			metrics := n.GetMetrics()
			n.logger.Info("NATS metrics",
				zap.Int64("events_published", metrics.EventsPublished),
				zap.Int64("events_consumed", metrics.EventsConsumed),
				zap.Int64("events_failed", metrics.EventsFailed),
				zap.Int64("active_subscriptions", metrics.ActiveSubscriptions),
				zap.Int64("avg_processing_time_ms", metrics.AvgProcessingTimeMs))

		case <-n.ctx.Done():
			return
		}
	}
}

// startHealthChecker démarre le vérificateur de santé
func (n *NATSService) startHealthChecker() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !n.conn.IsConnected() {
				n.logger.Warn("NATS connection lost, attempting reconnect")
			}

		case <-n.ctx.Done():
			return
		}
	}
}

// HealthCheck vérifie la santé du service NATS
func (n *NATSService) HealthCheck() error {
	if n.conn == nil || !n.conn.IsConnected() {
		return fmt.Errorf("NATS connection not available")
	}
	return nil
}

// Close ferme proprement le service NATS
func (n *NATSService) Close() error {
	n.cancel()

	// Attendre que tous les handlers terminent
	done := make(chan struct{})
	go func() {
		n.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		n.logger.Warn("Timeout waiting for handlers to finish")
	}

	// Fermer les souscriptions
	n.mutex.Lock()
	for subject, sub := range n.subscribers {
		if err := sub.Unsubscribe(); err != nil {
			n.logger.Warn("Failed to unsubscribe",
				zap.String("subject", subject),
				zap.Error(err))
		}
	}
	n.mutex.Unlock()

	// Fermer la connexion NATS
	if n.conn != nil {
		n.conn.Close()
	}

	n.logger.Info("NATS service closed")
	return nil
}
