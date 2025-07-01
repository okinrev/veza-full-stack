package notifications

import (
	"context"
	"strconv"
	"time"
	"github.com/gin-gonic/gin"
	"database/sql"
	"os"

	"go.uber.org/zap"
)

// NotificationSystem système complet de notifications
type NotificationSystem struct {
	Service          *NotificationService
	WebSocketService *WebSocketService
	Handler          *Handler
	Storage          NotificationStorage
}

// Config configuration du système de notifications
type Config struct {
	// WebSocket
	EnableWebSocket bool

	// Email
	EnableEmail bool
	SMTPHost    string
	SMTPPort    string
	SMTPUser    string
	SMTPPass    string
	FromEmail   string
	FromName    string

	// SMS
	EnableSMS  bool
	SMSAPIKey  string
	SMSAPIURL  string
	SMSSender  string

	// Push Notifications
	EnablePush bool
	FCMAPIKey  string
	APNSKey    string

	// Webhook
	EnableWebhook bool

	// Général
	MaxRetries int
	Workers    int
}

// LoadConfigFromEnv charge la configuration depuis les variables d'environnement
func LoadConfigFromEnv() *Config {
	return &Config{
		// WebSocket
		EnableWebSocket: getEnvBool("NOTIFICATION_ENABLE_WEBSOCKET", true),

		// Email
		EnableEmail: getEnvBool("NOTIFICATION_ENABLE_EMAIL", true),
		SMTPHost:    getEnv("SMTP_HOST", "localhost"),
		SMTPPort:    getEnv("SMTP_PORT", "587"),
		SMTPUser:    getEnv("SMTP_USER", ""),
		SMTPPass:    getEnv("SMTP_PASS", ""),
		FromEmail:   getEnv("FROM_EMAIL", "noreply@veza.com"),
		FromName:    getEnv("FROM_NAME", "Veza"),

		// SMS
		EnableSMS: getEnvBool("NOTIFICATION_ENABLE_SMS", false),
		SMSAPIKey: getEnv("SMS_API_KEY", ""),
		SMSAPIURL: getEnv("SMS_API_URL", ""),
		SMSSender: getEnv("SMS_SENDER", "Veza"),

		// Push
		EnablePush: getEnvBool("NOTIFICATION_ENABLE_PUSH", true),
		FCMAPIKey:  getEnv("FCM_API_KEY", ""),
		APNSKey:    getEnv("APNS_KEY", ""),

		// Webhook
		EnableWebhook: getEnvBool("NOTIFICATION_ENABLE_WEBHOOK", false),

		// Général
		MaxRetries: getEnvInt("NOTIFICATION_MAX_RETRIES", 3),
		Workers:    getEnvInt("NOTIFICATION_WORKERS", 5),
	}
}

// InitializeNotificationSystem initialise le système complet de notifications
func InitializeNotificationSystem(db *sql.DB, logger *zap.Logger, config *Config) (*NotificationSystem, error) {
	if config == nil {
		config = LoadConfigFromEnv()
	}

	// Initialiser le storage
	storage := NewPostgreSQLStorage(db, logger)

	// Initialiser le service WebSocket
	websocketService := NewWebSocketService(logger)

	// Initialiser les services de canal
	var emailService EmailSender
	var smsService SMSSender
	var pushService PushSender
	var webhookService WebhookSender

	if config.EnableEmail {
		emailService = NewEmailService(
			logger,
			config.SMTPHost,
			config.SMTPPort,
			config.SMTPUser,
			config.SMTPPass,
			config.FromEmail,
			config.FromName,
		)
		logger.Info("📧 Email notification service initialized")
	}

	if config.EnableSMS {
		smsService = NewSMSService(
			logger,
			config.SMSAPIKey,
			config.SMSAPIURL,
			config.SMSSender,
		)
		logger.Info("📱 SMS notification service initialized")
	}

	if config.EnablePush {
		pushService = NewPushService(
			logger,
			config.FCMAPIKey,
			config.APNSKey,
		)
		logger.Info("🔔 Push notification service initialized")
	}

	if config.EnableWebhook {
		webhookService = NewWebhookService(logger)
		logger.Info("🔗 Webhook notification service initialized")
	}

	// Configuration du service principal
	serviceConfig := &NotificationConfig{
		EnableWebSocket: config.EnableWebSocket,
		EnableEmail:     config.EnableEmail,
		EnableSMS:       config.EnableSMS,
		EnablePush:      config.EnablePush,
		EnableWebhook:   config.EnableWebhook,
		MaxRetries:      config.MaxRetries,
	}

	// Initialiser le service principal
	notificationService := NewNotificationService(
		logger,
		websocketService,
		emailService,
		smsService,
		pushService,
		webhookService,
		storage,
		serviceConfig,
	)

	// Initialiser le handler HTTP
	handler := NewHandler(notificationService, websocketService, logger)

	logger.Info("🚀 Notification system initialized successfully",
		zap.Bool("websocket", config.EnableWebSocket),
		zap.Bool("email", config.EnableEmail),
		zap.Bool("sms", config.EnableSMS),
		zap.Bool("push", config.EnablePush),
		zap.Bool("webhook", config.EnableWebhook))

	return &NotificationSystem{
		Service:          notificationService,
		WebSocketService: websocketService,
		Handler:          handler,
		Storage:          storage,
	}, nil
}

// Start démarre tous les services du système de notifications
func (ns *NotificationSystem) Start(ctx context.Context) error {
	ns.WebSocketService.Start(ctx)
	ns.Service.Start(ctx)
	return nil
}

// Shutdown arrête proprement le système de notifications
func (ns *NotificationSystem) Shutdown(ctx context.Context) error {
	// TODO: Implémenter l'arrêt propre des services
	return nil
}

// ============================================================================
// FONCTIONS UTILITAIRES
// ============================================================================

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1"
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// ============================================================================
// INTÉGRATION AVEC L'APPLICATION PRINCIPALE
// ============================================================================

// AddToApplication ajoute le système de notifications à l'application
func (ns *NotificationSystem) AddToApplication(router *gin.RouterGroup) {
	RegisterRoutes(router, ns.Handler, zap.L())
}

// SendQuickNotification envoie rapidement une notification simple
func (ns *NotificationSystem) SendQuickNotification(ctx context.Context, userID, title, message string, priority Priority) error {
	req := &NotificationRequest{
		Type:     "general",
		UserID:   userID,
		Title:    title,
		Message:  message,
		Priority: priority,
		Channels: []Channel{ChannelWebSocket, ChannelInApp},
	}

	_, err := ns.Service.SendNotification(ctx, req)
	return err
}

// SendSystemAlert envoie une alerte système
func (ns *NotificationSystem) SendSystemAlert(ctx context.Context, userID, title, message string) error {
	req := &NotificationRequest{
		Type:     NotificationSystemDegraded,
		UserID:   userID,
		Title:    title,
		Message:  message,
		Priority: PriorityCritical,
		Channels: []Channel{ChannelWebSocket, ChannelEmail, ChannelPush},
		Tags:     []string{"system", "alert"},
	}

	_, err := ns.Service.SendNotification(ctx, req)
	return err
}

// SendSecurityAlert envoie une alerte de sécurité
func (ns *NotificationSystem) SendSecurityAlert(ctx context.Context, userID, title, message string) error {
	req := &NotificationRequest{
		Type:     NotificationSecurityAlert,
		UserID:   userID,
		Title:    title,
		Message:  message,
		Priority: PriorityHigh,
		Channels: []Channel{ChannelWebSocket, ChannelEmail, ChannelSMS},
		Tags:     []string{"security", "alert"},
	}

	_, err := ns.Service.SendNotification(ctx, req)
	return err
}

// BroadcastSystemMessage diffuse un message système à tous les utilisateurs connectés
func (ns *NotificationSystem) BroadcastSystemMessage(ctx context.Context, title, message string, priority Priority) error {
	notification := &Notification{
		ID:        generateNotificationID(),
		Type:      NotificationSystemMaintenance,
		UserID:    "", // Broadcast
		Title:     title,
		Message:   message,
		Priority:  priority,
		Channels:  []Channel{ChannelWebSocket},
		CreatedAt: time.Now(),
		Source:    "system_broadcast",
		Tags:      []string{"system", "broadcast"},
	}

	ns.WebSocketService.Broadcast(notification)
	return nil
}
