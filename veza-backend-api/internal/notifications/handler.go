package notifications

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	
	"github.com/okinrev/veza-web-app/internal/response"
)

// Handler gestionnaire HTTP pour les notifications
type Handler struct {
	service          *NotificationService
	websocketService *WebSocketService
	logger           *zap.Logger
}

// NewHandler crée un nouveau handler de notifications
func NewHandler(service *NotificationService, websocketService *WebSocketService, logger *zap.Logger) *Handler {
	return &Handler{
		service:          service,
		websocketService: websocketService,
		logger:           logger,
	}
}

// ============================================================================
// ENDPOINTS REST API
// ============================================================================

// SendNotification envoie une nouvelle notification
// POST /api/v1/notifications/send
func (h *Handler) SendNotification(c *gin.Context) {
	var req NotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data: "+err.Error())
		return
	}

	// Vérifier l'autorisation (admin ou destinataire)
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	role, _ := c.Get("role")
	
	// Seuls les admins peuvent envoyer des notifications à d'autres utilisateurs
	if req.UserID != userID && role != "admin" {
		response.Forbidden(c, "Cannot send notifications to other users")
		return
	}

	notification, err := h.service.SendNotification(c.Request.Context(), &req)
	if err != nil {
		response.InternalServerError(c, "Failed to send notification: "+err.Error())
		return
	}

	response.Success(c, notification, "Notification sent successfully")
}

// SendBulkNotification envoie des notifications en masse
// POST /api/v1/notifications/bulk
func (h *Handler) SendBulkNotification(c *gin.Context) {
	// Vérifier les permissions admin
	role, exists := c.Get("role")
	if !exists || role != "admin" {
		response.Forbidden(c, "Admin privileges required")
		return
	}

	var req struct {
		UserIDs []string              `json:"user_ids" binding:"required"`
		NotificationRequest
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data: "+err.Error())
		return
	}

	notifications, err := h.service.SendBulkNotification(c.Request.Context(), req.UserIDs, &req.NotificationRequest)
	if err != nil {
		response.InternalServerError(c, "Failed to send bulk notifications: "+err.Error())
		return
	}

	response.Success(c, map[string]interface{}{
		"notifications": notifications,
		"sent_count":    len(notifications),
	}, "Bulk notifications sent successfully")
}

// GetUserNotifications récupère les notifications d'un utilisateur
// GET /api/v1/notifications
func (h *Handler) GetUserNotifications(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	// Paramètres de pagination
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")
	unreadOnlyStr := c.DefaultQuery("unread_only", "false")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		response.BadRequest(c, "Invalid limit parameter")
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		response.BadRequest(c, "Invalid offset parameter")
		return
	}

	unreadOnly, err := strconv.ParseBool(unreadOnlyStr)
	if err != nil {
		response.BadRequest(c, "Invalid unread_only parameter")
		return
	}

	notifications, err := h.service.GetUserNotifications(
		c.Request.Context(),
		userID.(string),
		limit,
		offset,
		unreadOnly,
	)
	if err != nil {
		response.InternalServerError(c, "Failed to get notifications: "+err.Error())
		return
	}

	// Obtenir le compte de non lues
	unreadCount, err := h.service.GetUnreadCount(c.Request.Context(), userID.(string))
	if err != nil {
		h.logger.Error("Failed to get unread count", zap.Error(err))
		unreadCount = 0
	}

	response.Success(c, map[string]interface{}{
		"notifications": notifications,
		"unread_count":  unreadCount,
		"total_count":   len(notifications),
		"limit":         limit,
		"offset":        offset,
	}, "Notifications retrieved successfully")
}

// MarkAsRead marque une notification comme lue
// PUT /api/v1/notifications/:id/read
func (h *Handler) MarkAsRead(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	notificationID := c.Param("id")
	if notificationID == "" {
		response.BadRequest(c, "Notification ID required")
		return
	}

	err := h.service.MarkAsRead(c.Request.Context(), notificationID, userID.(string))
	if err != nil {
		response.InternalServerError(c, "Failed to mark notification as read: "+err.Error())
		return
	}

	response.Success(c, map[string]interface{}{
		"notification_id": notificationID,
		"read_at":         time.Now(),
	}, "Notification marked as read")
}

// GetUnreadCount retourne le nombre de notifications non lues
// GET /api/v1/notifications/unread-count
func (h *Handler) GetUnreadCount(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	count, err := h.service.GetUnreadCount(c.Request.Context(), userID.(string))
	if err != nil {
		response.InternalServerError(c, "Failed to get unread count: "+err.Error())
		return
	}

	response.Success(c, map[string]interface{}{
		"unread_count": count,
	}, "Unread count retrieved successfully")
}

// ============================================================================
// WEBSOCKET ENDPOINT
// ============================================================================

// HandleWebSocket gère les connexions WebSocket
// GET /api/v1/notifications/ws
func (h *Handler) HandleWebSocket(c *gin.Context) {
	h.websocketService.HandleWebSocket(c)
}

// ============================================================================
// ENDPOINTS D'ADMINISTRATION
// ============================================================================

// GetStats retourne les statistiques des notifications
// GET /api/v1/notifications/stats
func (h *Handler) GetStats(c *gin.Context) {
	// Vérifier les permissions admin
	role, exists := c.Get("role")
	if !exists || role != "admin" {
		response.Forbidden(c, "Admin privileges required")
		return
	}

	wsStats := h.websocketService.GetStats()

	response.Success(c, map[string]interface{}{
		"websocket_stats": wsStats,
		"timestamp":       time.Now(),
	}, "Notification stats retrieved successfully")
}

// GetNotificationTypes retourne les types de notifications disponibles
// GET /api/v1/notifications/types
func (h *Handler) GetNotificationTypes(c *gin.Context) {
	types := []string{
		string(NotificationSystemMaintenance),
		string(NotificationSystemDegraded),
		string(NotificationSystemRestored),
		string(NotificationNewFollower),
		string(NotificationNewLike),
		string(NotificationNewComment),
		string(NotificationNewMessage),
		string(NotificationNewTrack),
		string(NotificationTrackUpdated),
		string(NotificationSecurityAlert),
		string(NotificationLoginFromNew),
		string(NotificationPasswordChanged),
		string(NotificationSubscriptionExpiring),
		string(NotificationPaymentFailed),
		string(NotificationNewFeature),
	}

	response.Success(c, map[string]interface{}{
		"types": types,
	}, "Notification types retrieved successfully")
}

// GetChannels retourne les canaux de notification disponibles
// GET /api/v1/notifications/channels
func (h *Handler) GetChannels(c *gin.Context) {
	channels := []string{
		string(ChannelWebSocket),
		string(ChannelEmail),
		string(ChannelSMS),
		string(ChannelPush),
		string(ChannelInApp),
		string(ChannelWebhook),
	}

	response.Success(c, map[string]interface{}{
		"channels": channels,
	}, "Notification channels retrieved successfully")
}

// ============================================================================
// PRÉFÉRENCES UTILISATEUR
// ============================================================================

// UserPreferences préférences de notification d'un utilisateur
type UserPreferences struct {
	UserID          string                              `json:"user_id"`
	EnabledChannels map[Channel]bool                    `json:"enabled_channels"`
	TypePreferences map[NotificationType][]Channel      `json:"type_preferences"`
	QuietHours      *QuietHours                         `json:"quiet_hours,omitempty"`
	EmailDigest     bool                                `json:"email_digest"`
	Language        string                              `json:"language"`
	Timezone        string                              `json:"timezone"`
}

// QuietHours heures de silence
type QuietHours struct {
	Enabled   bool   `json:"enabled"`
	StartHour int    `json:"start_hour"` // 0-23
	EndHour   int    `json:"end_hour"`   // 0-23
	Timezone  string `json:"timezone"`
}

// GetUserPreferences récupère les préférences de notification d'un utilisateur
// GET /api/v1/notifications/preferences
func (h *Handler) GetUserPreferences(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	// TODO: Récupérer les vraies préférences depuis la base de données
	// Pour l'instant, retourner des préférences par défaut
	preferences := &UserPreferences{
		UserID: userID.(string),
		EnabledChannels: map[Channel]bool{
			ChannelWebSocket: true,
			ChannelEmail:     true,
			ChannelPush:      true,
			ChannelSMS:       false,
			ChannelInApp:     true,
			ChannelWebhook:   false,
		},
		TypePreferences: map[NotificationType][]Channel{
			NotificationNewMessage:  {ChannelWebSocket, ChannelPush},
			NotificationNewFollower: {ChannelWebSocket, ChannelEmail},
			NotificationSecurityAlert: {ChannelWebSocket, ChannelEmail, ChannelSMS},
		},
		EmailDigest: true,
		Language:    "fr",
		Timezone:    "Europe/Paris",
	}

	response.Success(c, preferences, "User preferences retrieved successfully")
}

// UpdateUserPreferences met à jour les préférences de notification d'un utilisateur
// PUT /api/v1/notifications/preferences
func (h *Handler) UpdateUserPreferences(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	var preferences UserPreferences
	if err := c.ShouldBindJSON(&preferences); err != nil {
		response.BadRequest(c, "Invalid preferences data: "+err.Error())
		return
	}

	// S'assurer que l'utilisateur ne peut modifier que ses propres préférences
	preferences.UserID = userID.(string)

	// TODO: Sauvegarder les préférences en base de données

	h.logger.Info("User preferences updated",
		zap.String("user_id", userID.(string)))

	response.Success(c, preferences, "User preferences updated successfully")
}

// ============================================================================
// HELPERS
// ============================================================================

// TestNotification envoie une notification de test
// POST /api/v1/notifications/test
func (h *Handler) TestNotification(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	// Créer une notification de test
	req := &NotificationRequest{
		Type:     NotificationNewFeature,
		UserID:   userID.(string),
		Title:    "🧪 Notification de test",
		Message:  "Ceci est une notification de test pour vérifier que le système fonctionne correctement.",
		Priority: PriorityNormal,
		Channels: []Channel{ChannelWebSocket, ChannelInApp},
		Data: map[string]interface{}{
			"test":      true,
			"timestamp": time.Now().Unix(),
		},
		Tags: []string{"test", "system"},
	}

	notification, err := h.service.SendNotification(c.Request.Context(), req)
	if err != nil {
		response.InternalServerError(c, "Failed to send test notification: "+err.Error())
		return
	}

	response.Success(c, notification, "Test notification sent successfully")
}

// BroadcastNotification diffuse une notification à tous les utilisateurs connectés
// POST /api/v1/notifications/broadcast
func (h *Handler) BroadcastNotification(c *gin.Context) {
	// Vérifier les permissions admin
	role, exists := c.Get("role")
	if !exists || role != "admin" {
		response.Forbidden(c, "Admin privileges required")
		return
	}

	var req NotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data: "+err.Error())
		return
	}

	// Créer la notification de broadcast
	notification := &Notification{
		ID:        generateNotificationID(),
		Type:      req.Type,
		UserID:    "", // Broadcast = pas d'utilisateur spécifique
		Title:     req.Title,
		Message:   req.Message,
		Data:      req.Data,
		Priority:  req.Priority,
		Channels:  req.Channels,
		CreatedAt: time.Now(),
		Source:    "broadcast",
		Tags:      req.Tags,
		Metadata:  req.Metadata,
	}

	if req.ExpiresIn != nil {
		expiresAt := time.Now().Add(*req.ExpiresIn)
		notification.ExpiresAt = &expiresAt
	}

	// Diffuser via WebSocket
	h.websocketService.Broadcast(notification)

	response.Success(c, notification, "Broadcast notification sent successfully")
}
