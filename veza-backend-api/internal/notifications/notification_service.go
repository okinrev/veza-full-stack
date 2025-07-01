package notifications

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// NotificationService service principal de notifications multi-canal
type NotificationService struct {
	logger           *zap.Logger
	websocketService *WebSocketService
	emailService     EmailSender
	smsService       SMSSender
	pushService      PushSender
	webhookService   WebhookSender
	storage          NotificationStorage
	
	// Configuration
	config *NotificationConfig
	
	// Canaux de traitement
	notificationQueue chan *Notification
	retryQueue       chan *Notification
}

// NotificationConfig configuration du service de notifications
type NotificationConfig struct {
	EnableWebSocket bool                   `json:"enable_websocket"`
	EnableEmail     bool                   `json:"enable_email"`
	EnableSMS       bool                   `json:"enable_sms"`
	EnablePush      bool                   `json:"enable_push"`
	EnableWebhook   bool                   `json:"enable_webhook"`
	
	// Configuration des tentatives
	MaxRetries      int           `json:"max_retries"`
	RetryDelay      time.Duration `json:"retry_delay"`
	
	// Templates par défaut
	DefaultTemplates map[NotificationType]*NotificationTemplate `json:"default_templates"`
}

// NotificationTemplate template de notification
type NotificationTemplate struct {
	Subject      string                 `json:"subject"`
	HTMLBody     string                 `json:"html_body"`
	TextBody     string                 `json:"text_body"`
	Variables    map[string]interface{} `json:"variables"`
	Localization map[string]*LocalizedTemplate `json:"localization"`
}

// LocalizedTemplate template localisé
type LocalizedTemplate struct {
	Subject  string `json:"subject"`
	HTMLBody string `json:"html_body"`
	TextBody string `json:"text_body"`
}

// Interfaces des services de canal
type EmailSender interface {
	SendEmail(ctx context.Context, notification *Notification) error
}

type SMSSender interface {
	SendSMS(ctx context.Context, notification *Notification) error
}

type PushSender interface {
	SendPushNotification(ctx context.Context, notification *Notification) error
}

type WebhookSender interface {
	SendWebhook(ctx context.Context, notification *Notification) error
}

// NotificationStorage interface de stockage des notifications
type NotificationStorage interface {
	Store(ctx context.Context, notification *Notification) error
	GetByUser(ctx context.Context, userID string, limit, offset int) ([]*Notification, error)
	MarkAsRead(ctx context.Context, notificationID, userID string) error
	GetUnreadCount(ctx context.Context, userID string) (int, error)
	DeleteExpired(ctx context.Context) error
}

// NotificationRequest requête de création de notification
type NotificationRequest struct {
	Type        NotificationType       `json:"type" binding:"required"`
	UserID      string                 `json:"user_id" binding:"required"`
	Title       string                 `json:"title" binding:"required"`
	Message     string                 `json:"message" binding:"required"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Priority    Priority               `json:"priority"`
	Channels    []Channel              `json:"channels"`
	ExpiresIn   *time.Duration         `json:"expires_in,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Metadata    map[string]string      `json:"metadata,omitempty"`
}

// NotificationStats statistiques des notifications
type NotificationStats struct {
	TotalSent      int64                     `json:"total_sent"`
	TotalDelivered int64                     `json:"total_delivered"`
	TotalRead      int64                     `json:"total_read"`
	DeliveryRate   float64                   `json:"delivery_rate"`
	ReadRate       float64                   `json:"read_rate"`
	ChannelStats   map[Channel]*ChannelStats `json:"channel_stats"`
	TypeStats      map[NotificationType]*TypeStats `json:"type_stats"`
}

// ChannelStats statistiques par canal
type ChannelStats struct {
	Sent               int64         `json:"sent"`
	Delivered          int64         `json:"delivered"`
	Failed             int64         `json:"failed"`
	AverageDeliveryTime time.Duration `json:"average_delivery_time"`
}

// TypeStats statistiques par type
type TypeStats struct {
	Sent             int64         `json:"sent"`
	Read             int64         `json:"read"`
	AverageReadTime  time.Duration `json:"average_read_time"`
}

// NewNotificationService crée une nouvelle instance du service
func NewNotificationService(
	logger *zap.Logger,
	websocketService *WebSocketService,
	emailService EmailSender,
	smsService SMSSender,
	pushService PushSender,
	webhookService WebhookSender,
	storage NotificationStorage,
	config *NotificationConfig,
) *NotificationService {
	
	if config == nil {
		config = &NotificationConfig{
			EnableWebSocket: true,
			EnableEmail:     true,
			EnableSMS:       false,
			EnablePush:      true,
			EnableWebhook:   false,
			MaxRetries:      3,
			RetryDelay:      5 * time.Second,
			DefaultTemplates: make(map[NotificationType]*NotificationTemplate),
		}
	}

	return &NotificationService{
		logger:            logger,
		websocketService:  websocketService,
		emailService:      emailService,
		smsService:        smsService,
		pushService:       pushService,
		webhookService:    webhookService,
		storage:          storage,
		config:           config,
		notificationQueue: make(chan *Notification, 10000),
		retryQueue:       make(chan *Notification, 1000),
	}
}

// Start démarre le service de notifications
func (ns *NotificationService) Start(ctx context.Context) {
	ns.logger.Info("🚀 Starting notification service")
	
	// Démarrer les workers de traitement
	for i := 0; i < 5; i++ {
		go ns.notificationWorker(ctx, i)
	}
	
	// Worker de retry
	go ns.retryWorker(ctx)
	
	// Worker de nettoyage
	go ns.cleanupWorker(ctx)
}

// SendNotification envoie une notification
func (ns *NotificationService) SendNotification(ctx context.Context, req *NotificationRequest) (*Notification, error) {
	// Créer la notification
	notification := &Notification{
		ID:        generateNotificationID(),
		Type:      req.Type,
		UserID:    req.UserID,
		Title:     req.Title,
		Message:   req.Message,
		Data:      req.Data,
		Priority:  req.Priority,
		Channels:  req.Channels,
		CreatedAt: time.Now(),
		Source:    "notification_service",
		Tags:      req.Tags,
		Metadata:  req.Metadata,
	}

	// Définir l'expiration
	if req.ExpiresIn != nil {
		expiresAt := time.Now().Add(*req.ExpiresIn)
		notification.ExpiresAt = &expiresAt
	}

	// Définir priorité par défaut
	if notification.Priority == "" {
		notification.Priority = PriorityNormal
	}

	// Définir canaux par défaut si non spécifiés
	if len(notification.Channels) == 0 {
		notification.Channels = ns.getDefaultChannels(notification.Priority)
	}

	// Stocker la notification
	if ns.storage != nil {
		if err := ns.storage.Store(ctx, notification); err != nil {
			ns.logger.Error("Failed to store notification", zap.Error(err))
		}
	}

	// Ajouter à la queue de traitement
	select {
	case ns.notificationQueue <- notification:
		ns.logger.Debug("Notification queued",
			zap.String("id", notification.ID),
			zap.String("type", string(notification.Type)))
	default:
		ns.logger.Error("Notification queue full",
			zap.String("id", notification.ID))
		return nil, fmt.Errorf("notification queue full")
	}

	return notification, nil
}

// SendBulkNotification envoie des notifications en masse
func (ns *NotificationService) SendBulkNotification(ctx context.Context, userIDs []string, req *NotificationRequest) ([]*Notification, error) {
	notifications := make([]*Notification, 0, len(userIDs))
	
	for _, userID := range userIDs {
		userReq := *req
		userReq.UserID = userID
		
		notification, err := ns.SendNotification(ctx, &userReq)
		if err != nil {
			ns.logger.Error("Failed to send bulk notification",
				zap.String("user_id", userID),
				zap.Error(err))
			continue
		}
		
		notifications = append(notifications, notification)
	}
	
	return notifications, nil
}

// GetUserNotifications récupère les notifications d'un utilisateur
func (ns *NotificationService) GetUserNotifications(ctx context.Context, userID string, limit, offset int, unreadOnly bool) ([]*Notification, error) {
	if ns.storage == nil {
		return []*Notification{}, nil
	}
	
	notifications, err := ns.storage.GetByUser(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	
	// Filtrer les non lues si demandé
	if unreadOnly {
		filtered := make([]*Notification, 0)
		for _, notif := range notifications {
			if notif.ReadAt == nil {
				filtered = append(filtered, notif)
			}
		}
		return filtered, nil
	}
	
	return notifications, nil
}

// MarkAsRead marque une notification comme lue
func (ns *NotificationService) MarkAsRead(ctx context.Context, notificationID, userID string) error {
	if ns.storage == nil {
		return fmt.Errorf("storage not configured")
	}
	
	return ns.storage.MarkAsRead(ctx, notificationID, userID)
}

// GetUnreadCount retourne le nombre de notifications non lues
func (ns *NotificationService) GetUnreadCount(ctx context.Context, userID string) (int, error) {
	if ns.storage == nil {
		return 0, nil
	}
	
	return ns.storage.GetUnreadCount(ctx, userID)
}

// ============================================================================
// WORKERS DE TRAITEMENT
// ============================================================================

// notificationWorker traite les notifications de la queue
func (ns *NotificationService) notificationWorker(ctx context.Context, workerID int) {
	ns.logger.Info("Starting notification worker", zap.Int("worker_id", workerID))
	
	for {
		select {
		case <-ctx.Done():
			return
		case notification := <-ns.notificationQueue:
			ns.processNotification(ctx, notification)
		}
	}
}

// processNotification traite une notification
func (ns *NotificationService) processNotification(ctx context.Context, notification *Notification) {
	startTime := time.Now()
	
	// Vérifier l'expiration
	if notification.ExpiresAt != nil && time.Now().After(*notification.ExpiresAt) {
		ns.logger.Debug("Notification expired, skipping",
			zap.String("id", notification.ID))
		return
	}
	
	// Traiter chaque canal
	for _, channel := range notification.Channels {
		if err := ns.deliverToChannel(ctx, notification, channel); err != nil {
			ns.logger.Error("Failed to deliver notification",
				zap.String("id", notification.ID),
				zap.String("channel", string(channel)),
				zap.Error(err))
			
			// Programmer un retry pour les échecs
			ns.scheduleRetry(notification, channel, err)
		} else {
			ns.logger.Debug("Notification delivered successfully",
				zap.String("id", notification.ID),
				zap.String("channel", string(channel)),
				zap.Duration("duration", time.Since(startTime)))
		}
	}
	
	// Marquer comme délivrée
	now := time.Now()
	notification.DeliveredAt = &now
}

// deliverToChannel livre une notification via un canal spécifique
func (ns *NotificationService) deliverToChannel(ctx context.Context, notification *Notification, channel Channel) error {
	switch channel {
	case ChannelWebSocket:
		if ns.config.EnableWebSocket && ns.websocketService != nil {
			return ns.websocketService.SendToUser(notification.UserID, notification)
		}
	case ChannelEmail:
		if ns.config.EnableEmail && ns.emailService != nil {
			return ns.emailService.SendEmail(ctx, notification)
		}
	case ChannelSMS:
		if ns.config.EnableSMS && ns.smsService != nil {
			return ns.smsService.SendSMS(ctx, notification)
		}
	case ChannelPush:
		if ns.config.EnablePush && ns.pushService != nil {
			return ns.pushService.SendPushNotification(ctx, notification)
		}
	case ChannelWebhook:
		if ns.config.EnableWebhook && ns.webhookService != nil {
			return ns.webhookService.SendWebhook(ctx, notification)
		}
	default:
		return fmt.Errorf("unknown channel: %s", channel)
	}
	
	return fmt.Errorf("channel %s not enabled or not configured", channel)
}

// scheduleRetry programme un retry
func (ns *NotificationService) scheduleRetry(notification *Notification, channel Channel, err error) {
	// TODO: Implémenter la logique de retry avec backoff exponentiel
	ns.logger.Debug("Scheduling retry",
		zap.String("notification_id", notification.ID),
		zap.String("channel", string(channel)))
}

// retryWorker traite les retries
func (ns *NotificationService) retryWorker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case notification := <-ns.retryQueue:
			// Attendre avant de réessayer
			time.Sleep(ns.config.RetryDelay)
			ns.processNotification(ctx, notification)
		}
	}
}

// cleanupWorker nettoie les notifications expirées
func (ns *NotificationService) cleanupWorker(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if ns.storage != nil {
				if err := ns.storage.DeleteExpired(ctx); err != nil {
					ns.logger.Error("Failed to cleanup expired notifications", zap.Error(err))
				}
			}
		}
	}
}

// getDefaultChannels retourne les canaux par défaut selon la priorité
func (ns *NotificationService) getDefaultChannels(priority Priority) []Channel {
	switch priority {
	case PriorityEmergency, PriorityCritical:
		return []Channel{ChannelWebSocket, ChannelEmail, ChannelPush, ChannelSMS}
	case PriorityHigh:
		return []Channel{ChannelWebSocket, ChannelEmail, ChannelPush}
	case PriorityNormal:
		return []Channel{ChannelWebSocket, ChannelInApp}
	case PriorityLow:
		return []Channel{ChannelInApp}
	default:
		return []Channel{ChannelWebSocket, ChannelInApp}
	}
}
