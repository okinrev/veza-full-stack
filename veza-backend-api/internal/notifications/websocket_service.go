package notifications

import (
	"context"
	"fmt"
	mrand "math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// WebSocketService gère les connexions WebSocket temps réel
type WebSocketService struct {
	connections map[string]*WebSocketConnection
	mutex       sync.RWMutex
	upgrader    websocket.Upgrader
	logger      *zap.Logger

	// Canaux pour diffusion
	broadcast  chan *Notification
	register   chan *WebSocketConnection
	unregister chan *WebSocketConnection

	// Statistiques
	stats *WebSocketStats
}

// WebSocketConnection représente une connexion client
type WebSocketConnection struct {
	ID           string
	UserID       string
	Username     string
	Conn         *websocket.Conn
	Send         chan *Notification
	Service      *WebSocketService
	ConnectedAt  time.Time
	LastActivity time.Time

	// Filtres et préférences
	SubscribedTypes []NotificationType
	Rooms           map[string]bool
}

// WebSocketStats statistiques des connexions WebSocket
type WebSocketStats struct {
	ActiveConnections int64            `json:"active_connections"`
	TotalConnections  int64            `json:"total_connections"`
	MessagesSent      int64            `json:"messages_sent"`
	ConnectionsByType map[string]int64 `json:"connections_by_type"`
	Uptime            time.Time        `json:"uptime"`
	LastActivity      time.Time        `json:"last_activity"`
}

// NotificationType types de notifications supportées
type NotificationType string

const (
	// Notifications système
	NotificationSystemMaintenance NotificationType = "system_maintenance"
	NotificationSystemDegraded    NotificationType = "system_degraded"
	NotificationSystemRestored    NotificationType = "system_restored"

	// Notifications sociales
	NotificationNewFollower NotificationType = "new_follower"
	NotificationNewLike     NotificationType = "new_like"
	NotificationNewComment  NotificationType = "new_comment"
	NotificationNewMessage  NotificationType = "new_message"

	// Notifications de contenu
	NotificationNewTrack     NotificationType = "new_track"
	NotificationTrackUpdated NotificationType = "track_updated"

	// Notifications de sécurité
	NotificationSecurityAlert   NotificationType = "security_alert"
	NotificationLoginFromNew    NotificationType = "login_from_new"
	NotificationPasswordChanged NotificationType = "password_changed"

	// Notifications business
	NotificationSubscriptionExpiring NotificationType = "subscription_expiring"
	NotificationPaymentFailed        NotificationType = "payment_failed"
	NotificationNewFeature           NotificationType = "new_feature"
)

// Notification structure unifée pour toutes les notifications
type Notification struct {
	ID          string                 `json:"id"`
	Type        NotificationType       `json:"type"`
	UserID      string                 `json:"user_id"`
	Title       string                 `json:"title"`
	Message     string                 `json:"message"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Priority    Priority               `json:"priority"`
	Channels    []Channel              `json:"channels"`
	CreatedAt   time.Time              `json:"created_at"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
	ReadAt      *time.Time             `json:"read_at,omitempty"`
	DeliveredAt *time.Time             `json:"delivered_at,omitempty"`

	// Métadonnées
	Source   string            `json:"source"`
	Tags     []string          `json:"tags"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Priority niveaux de priorité
type Priority string

const (
	PriorityLow       Priority = "low"
	PriorityNormal    Priority = "normal"
	PriorityHigh      Priority = "high"
	PriorityCritical  Priority = "critical"
	PriorityEmergency Priority = "emergency"
)

// Channel canaux de diffusion
type Channel string

const (
	ChannelWebSocket Channel = "websocket"
	ChannelEmail     Channel = "email"
	ChannelSMS       Channel = "sms"
	ChannelPush      Channel = "push"
	ChannelInApp     Channel = "inapp"
	ChannelWebhook   Channel = "webhook"
)

// NewWebSocketService crée une nouvelle instance du service WebSocket
func NewWebSocketService(logger *zap.Logger) *WebSocketService {
	return &WebSocketService{
		connections: make(map[string]*WebSocketConnection),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// TODO: Implémenter une vérification d'origine plus stricte en production
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		logger:     logger,
		broadcast:  make(chan *Notification, 1000),
		register:   make(chan *WebSocketConnection),
		unregister: make(chan *WebSocketConnection),
		stats: &WebSocketStats{
			ConnectionsByType: make(map[string]int64),
			Uptime:            time.Now(),
		},
	}
}

// Start démarre le service WebSocket
func (ws *WebSocketService) Start(ctx context.Context) {
	ws.logger.Info("🚀 Starting WebSocket notification service")

	go ws.run(ctx)

	// Worker de nettoyage des connexions inactives
	go ws.cleanupWorker(ctx)

	// Worker de statistiques
	go ws.statsWorker(ctx)
}

// HandleWebSocket gère une nouvelle connexion WebSocket
func (ws *WebSocketService) HandleWebSocket(c *gin.Context) {
	conn, err := ws.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		ws.logger.Error("Failed to upgrade WebSocket connection", zap.Error(err))
		return
	}

	// Récupérer les informations utilisateur depuis le contexte
	userID, exists := c.Get("user_id")
	if !exists {
		ws.logger.Warn("WebSocket connection without user authentication")
		conn.Close()
		return
	}

	username, _ := c.Get("username")

	// Créer la connexion WebSocket
	connection := &WebSocketConnection{
		ID:              generateConnectionID(),
		UserID:          fmt.Sprintf("%v", userID),
		Username:        fmt.Sprintf("%v", username),
		Conn:            conn,
		Send:            make(chan *Notification, 256),
		Service:         ws,
		ConnectedAt:     time.Now(),
		LastActivity:    time.Now(),
		SubscribedTypes: []NotificationType{}, // Par défaut, tous les types
		Rooms:           make(map[string]bool),
	}

	// Enregistrer la connexion
	ws.register <- connection

	// Démarrer les goroutines de lecture et d'écriture
	go connection.writePump()
	go connection.readPump()
}

// SendToUser envoie une notification à un utilisateur spécifique
func (ws *WebSocketService) SendToUser(userID string, notification *Notification) error {
	ws.mutex.RLock()
	defer ws.mutex.RUnlock()

	sent := false
	for _, conn := range ws.connections {
		if conn.UserID == userID && ws.shouldReceiveNotification(conn, notification) {
			select {
			case conn.Send <- notification:
				sent = true
			default:
				ws.logger.Warn("Failed to send notification to user",
					zap.String("user_id", userID),
					zap.String("notification_id", notification.ID))
			}
		}
	}

	if !sent {
		return fmt.Errorf("user %s not connected or filtered out", userID)
	}

	return nil
}

// Broadcast diffuse une notification à toutes les connexions
func (ws *WebSocketService) Broadcast(notification *Notification) {
	select {
	case ws.broadcast <- notification:
	default:
		ws.logger.Warn("Broadcast channel full, dropping notification",
			zap.String("notification_id", notification.ID))
	}
}

// SendToRoom envoie une notification à tous les utilisateurs d'une room
func (ws *WebSocketService) SendToRoom(roomID string, notification *Notification) {
	ws.mutex.RLock()
	defer ws.mutex.RUnlock()

	for _, conn := range ws.connections {
		if conn.Rooms[roomID] && ws.shouldReceiveNotification(conn, notification) {
			select {
			case conn.Send <- notification:
			default:
				ws.logger.Warn("Failed to send room notification",
					zap.String("room_id", roomID),
					zap.String("user_id", conn.UserID))
			}
		}
	}
}

// GetStats retourne les statistiques WebSocket
func (ws *WebSocketService) GetStats() *WebSocketStats {
	ws.mutex.RLock()
	defer ws.mutex.RUnlock()

	stats := *ws.stats
	stats.ActiveConnections = int64(len(ws.connections))
	stats.LastActivity = time.Now()

	return &stats
}

// ============================================================================
// MÉTHODES PRIVÉES
// ============================================================================

// run boucle principale du service
func (ws *WebSocketService) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			ws.logger.Info("Shutting down WebSocket service")
			return

		case connection := <-ws.register:
			ws.registerConnection(connection)

		case connection := <-ws.unregister:
			ws.unregisterConnection(connection)

		case notification := <-ws.broadcast:
			ws.broadcastNotification(notification)
		}
	}
}

// registerConnection enregistre une nouvelle connexion
func (ws *WebSocketService) registerConnection(conn *WebSocketConnection) {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	ws.connections[conn.ID] = conn
	ws.stats.TotalConnections++
	ws.stats.ConnectionsByType["websocket"]++

	ws.logger.Info("WebSocket connection registered",
		zap.String("connection_id", conn.ID),
		zap.String("user_id", conn.UserID),
		zap.Int("total_connections", len(ws.connections)))

	// Envoyer notification de bienvenue
	welcome := &Notification{
		ID:        generateNotificationID(),
		Type:      "system_connected",
		UserID:    conn.UserID,
		Title:     "Connecté",
		Message:   "Vous êtes maintenant connecté aux notifications temps réel",
		Priority:  PriorityLow,
		Channels:  []Channel{ChannelWebSocket},
		CreatedAt: time.Now(),
		Source:    "websocket_service",
	}

	select {
	case conn.Send <- welcome:
	default:
	}
}

// unregisterConnection désenregistre une connexion
func (ws *WebSocketService) unregisterConnection(conn *WebSocketConnection) {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	if _, exists := ws.connections[conn.ID]; exists {
		delete(ws.connections, conn.ID)
		close(conn.Send)

		ws.logger.Info("WebSocket connection unregistered",
			zap.String("connection_id", conn.ID),
			zap.String("user_id", conn.UserID),
			zap.Int("remaining_connections", len(ws.connections)))
	}
}

// broadcastNotification diffuse une notification à toutes les connexions
func (ws *WebSocketService) broadcastNotification(notification *Notification) {
	ws.mutex.RLock()
	defer ws.mutex.RUnlock()

	sent := 0
	for _, conn := range ws.connections {
		if ws.shouldReceiveNotification(conn, notification) {
			select {
			case conn.Send <- notification:
				sent++
			default:
				ws.logger.Warn("Failed to send broadcast notification",
					zap.String("user_id", conn.UserID))
			}
		}
	}

	ws.stats.MessagesSent += int64(sent)
	ws.logger.Debug("Notification broadcasted",
		zap.String("notification_id", notification.ID),
		zap.Int("recipients", sent))
}

// shouldReceiveNotification détermine si une connexion doit recevoir une notification
func (ws *WebSocketService) shouldReceiveNotification(conn *WebSocketConnection, notification *Notification) bool {
	// Vérifier si l'utilisateur est le destinataire
	if notification.UserID != "" && notification.UserID != conn.UserID {
		return false
	}

	// Vérifier les filtres par type
	if len(conn.SubscribedTypes) > 0 {
		found := false
		for _, subType := range conn.SubscribedTypes {
			if subType == notification.Type {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Vérifier l'expiration
	if notification.ExpiresAt != nil && time.Now().After(*notification.ExpiresAt) {
		return false
	}

	return true
}

// cleanupWorker nettoie les connexions inactives
func (ws *WebSocketService) cleanupWorker(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ws.cleanupInactiveConnections()
		}
	}
}

// cleanupInactiveConnections ferme les connexions inactives
func (ws *WebSocketService) cleanupInactiveConnections() {
	ws.mutex.RLock()
	inactiveConnections := make([]*WebSocketConnection, 0)

	for _, conn := range ws.connections {
		if time.Since(conn.LastActivity) > 5*time.Minute {
			inactiveConnections = append(inactiveConnections, conn)
		}
	}
	ws.mutex.RUnlock()

	for _, conn := range inactiveConnections {
		conn.Conn.Close()
		ws.unregister <- conn
	}

	if len(inactiveConnections) > 0 {
		ws.logger.Info("Cleaned up inactive connections",
			zap.Int("count", len(inactiveConnections)))
	}
}

// statsWorker met à jour les statistiques périodiquement
func (ws *WebSocketService) statsWorker(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ws.updateStats()
		}
	}
}

// updateStats met à jour les statistiques
func (ws *WebSocketService) updateStats() {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	ws.stats.LastActivity = time.Now()

	// Log des stats périodiques
	ws.logger.Info("WebSocket stats",
		zap.Int("active_connections", len(ws.connections)),
		zap.Int64("total_connections", ws.stats.TotalConnections),
		zap.Int64("messages_sent", ws.stats.MessagesSent))
}

// generateConnectionID génère un ID unique pour une connexion
func generateConnectionID() string {
	return fmt.Sprintf("ws_%d_%d", time.Now().UnixNano(), mrand.Intn(10000))
}

// generateNotificationID génère un ID unique pour une notification
func generateNotificationID() string {
	return fmt.Sprintf("notif_%d_%d", time.Now().UnixNano(), mrand.Intn(10000))
}

// ============================================================================
// MÉTHODES DE CONNEXION WEBSOCKET
// ============================================================================

// readPump lit les messages du client WebSocket
func (c *WebSocketConnection) readPump() {
	defer func() {
		c.Service.unregister <- c
		c.Conn.Close()
	}()

	// Configuration des timeouts
	c.Conn.SetReadLimit(512)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		c.LastActivity = time.Now()
		return nil
	})

	for {
		var message map[string]interface{}
		err := c.Conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.Service.logger.Error("WebSocket read error", zap.Error(err))
			}
			break
		}

		c.LastActivity = time.Now()
		c.handleClientMessage(message)
	}
}

// writePump écrit les messages vers le client WebSocket
func (c *WebSocketConnection) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case notification, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Envoyer la notification
			if err := c.Conn.WriteJSON(notification); err != nil {
				c.Service.logger.Error("Failed to write WebSocket message", zap.Error(err))
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleClientMessage traite un message reçu du client
func (c *WebSocketConnection) handleClientMessage(message map[string]interface{}) {
	messageType, ok := message["type"].(string)
	if !ok {
		c.Service.logger.Warn("Invalid message type from client", zap.String("user_id", c.UserID))
		return
	}

	data, _ := message["data"].(map[string]interface{})

	switch messageType {
	case "subscribe":
		c.handleSubscribe(data)
	case "unsubscribe":
		c.handleUnsubscribe(data)
	case "join_room":
		c.handleJoinRoom(data)
	case "leave_room":
		c.handleLeaveRoom(data)
	case "mark_read":
		c.handleMarkRead(data)
	case "ping":
		c.handlePing()
	default:
		c.Service.logger.Warn("Unknown message type", zap.String("type", messageType))
	}
}

// handleSubscribe gère l'abonnement à des types de notifications
func (c *WebSocketConnection) handleSubscribe(data map[string]interface{}) {
	types, ok := data["types"].([]interface{})
	if !ok {
		return
	}

	for _, t := range types {
		if typeStr, ok := t.(string); ok {
			notifType := NotificationType(typeStr)
			// Vérifier si pas déjà abonné
			found := false
			for _, existing := range c.SubscribedTypes {
				if existing == notifType {
					found = true
					break
				}
			}
			if !found {
				c.SubscribedTypes = append(c.SubscribedTypes, notifType)
			}
		}
	}

	c.Service.logger.Debug("User subscribed to notification types",
		zap.String("user_id", c.UserID),
		zap.Any("types", c.SubscribedTypes))

	// Confirmer l'abonnement
	response := &Notification{
		ID:        generateNotificationID(),
		Type:      "subscription_confirmed",
		UserID:    c.UserID,
		Title:     "Abonnement confirmé",
		Message:   fmt.Sprintf("Abonné à %d types de notifications", len(c.SubscribedTypes)),
		Priority:  PriorityLow,
		Channels:  []Channel{ChannelWebSocket},
		CreatedAt: time.Now(),
		Source:    "websocket_service",
	}

	select {
	case c.Send <- response:
	default:
	}
}

// handleUnsubscribe gère le désabonnement
func (c *WebSocketConnection) handleUnsubscribe(data map[string]interface{}) {
	types, ok := data["types"].([]interface{})
	if !ok {
		return
	}

	for _, t := range types {
		if typeStr, ok := t.(string); ok {
			notifType := NotificationType(typeStr)
			// Retirer de la liste
			newTypes := make([]NotificationType, 0)
			for _, existing := range c.SubscribedTypes {
				if existing != notifType {
					newTypes = append(newTypes, existing)
				}
			}
			c.SubscribedTypes = newTypes
		}
	}

	c.Service.logger.Debug("User unsubscribed from notification types",
		zap.String("user_id", c.UserID))
}

// handleJoinRoom gère l'adhésion à une room
func (c *WebSocketConnection) handleJoinRoom(data map[string]interface{}) {
	roomID, ok := data["room_id"].(string)
	if !ok {
		return
	}

	c.Rooms[roomID] = true
	c.Service.logger.Debug("User joined room",
		zap.String("user_id", c.UserID),
		zap.String("room_id", roomID))
}

// handleLeaveRoom gère la sortie d'une room
func (c *WebSocketConnection) handleLeaveRoom(data map[string]interface{}) {
	roomID, ok := data["room_id"].(string)
	if !ok {
		return
	}

	delete(c.Rooms, roomID)
	c.Service.logger.Debug("User left room",
		zap.String("user_id", c.UserID),
		zap.String("room_id", roomID))
}

// handleMarkRead marque une notification comme lue
func (c *WebSocketConnection) handleMarkRead(data map[string]interface{}) {
	notificationID, ok := data["notification_id"].(string)
	if !ok {
		return
	}

	// TODO: Intégrer avec le système de stockage des notifications
	c.Service.logger.Debug("Notification marked as read",
		zap.String("user_id", c.UserID),
		zap.String("notification_id", notificationID))

	// Répondre avec confirmation
	response := &Notification{
		ID:        generateNotificationID(),
		Type:      "read_confirmed",
		UserID:    c.UserID,
		Title:     "Notification lue",
		Message:   "Notification marquée comme lue",
		Priority:  PriorityLow,
		Channels:  []Channel{ChannelWebSocket},
		CreatedAt: time.Now(),
		Source:    "websocket_service",
		Data:      map[string]interface{}{"original_id": notificationID},
	}

	select {
	case c.Send <- response:
	default:
	}
}

// handlePing répond à un ping du client
func (c *WebSocketConnection) handlePing() {
	response := &Notification{
		ID:        generateNotificationID(),
		Type:      "pong",
		UserID:    c.UserID,
		Title:     "Pong",
		Message:   "Server is alive",
		Priority:  PriorityLow,
		Channels:  []Channel{ChannelWebSocket},
		CreatedAt: time.Now(),
		Source:    "websocket_service",
		Data:      map[string]interface{}{"timestamp": time.Now().Unix()},
	}

	select {
	case c.Send <- response:
	default:
	}
}
