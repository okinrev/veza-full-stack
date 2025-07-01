package messagequeue

import (
	"context"
	"encoding/json"
	"fmt"
	"net/smtp"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// NotificationQueueService service pour la gestion des queues de notifications
type NotificationQueueService struct {
	natsService    *NATSService
	emailService   *EmailService
	templateEngine *TemplateEngine
	logger         *zap.Logger
	config         *NotificationConfig

	// Métriques
	metrics *NotificationMetrics

	// Contrôle de lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NotificationConfig configuration des notifications
type NotificationConfig struct {
	// Configuration email
	SMTPHost     string `json:"smtp_host"`
	SMTPPort     int    `json:"smtp_port"`
	SMTPUsername string `json:"smtp_username"`
	SMTPPassword string `json:"smtp_password"`
	FromEmail    string `json:"from_email"`
	FromName     string `json:"from_name"`

	// Configuration retry
	MaxRetries   int           `json:"max_retries"`
	RetryDelay   time.Duration `json:"retry_delay"`
	RetryBackoff float64       `json:"retry_backoff"`

	// Configuration rate limiting
	EmailRateLimit  int           `json:"email_rate_limit"`
	EmailRateWindow time.Duration `json:"email_rate_window"`

	// Configuration templates
	TemplateDir     string `json:"template_dir"`
	DefaultLanguage string `json:"default_language"`

	// Configuration workers
	WorkerCount int `json:"worker_count"`
	BatchSize   int `json:"batch_size"`
}

// NotificationMessage message de notification
type NotificationMessage struct {
	ID          string                 `json:"id"`
	Type        NotificationType       `json:"type"`
	Recipients  []string               `json:"recipients"`
	Subject     string                 `json:"subject"`
	Template    string                 `json:"template"`
	Data        map[string]interface{} `json:"data"`
	Priority    NotificationPriority   `json:"priority"`
	ScheduledAt *time.Time             `json:"scheduled_at,omitempty"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
	Language    string                 `json:"language"`
	UserID      *int64                 `json:"user_id,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	RetryCount  int                    `json:"retry_count"`
	CreatedAt   time.Time              `json:"created_at"`
}

// NotificationType types de notifications
type NotificationType string

const (
	// Notifications par email
	NotificationEmail              NotificationType = "email"
	NotificationEmailWelcome       NotificationType = "email.welcome"
	NotificationEmailVerification  NotificationType = "email.verification"
	NotificationEmailPasswordReset NotificationType = "email.password_reset"
	NotificationEmailNewsletter    NotificationType = "email.newsletter"
	NotificationEmailAlert         NotificationType = "email.alert"

	// Notifications push
	NotificationPush        NotificationType = "push"
	NotificationPushMessage NotificationType = "push.message"
	NotificationPushAlert   NotificationType = "push.alert"

	// Notifications in-app
	NotificationInApp        NotificationType = "in_app"
	NotificationInAppMessage NotificationType = "in_app.message"
	NotificationInAppAlert   NotificationType = "in_app.alert"

	// Notifications SMS
	NotificationSMS             NotificationType = "sms"
	NotificationSMSVerification NotificationType = "sms.verification"
	NotificationSMSAlert        NotificationType = "sms.alert"
)

// NotificationPriority priorité des notifications
type NotificationPriority string

const (
	NotificationPriorityLow      NotificationPriority = "low"
	NotificationPriorityNormal   NotificationPriority = "normal"
	NotificationPriorityHigh     NotificationPriority = "high"
	NotificationPriorityCritical NotificationPriority = "critical"
)

// NotificationMetrics métriques des notifications
type NotificationMetrics struct {
	EmailsSent        int64            `json:"emails_sent"`
	EmailsFailed      int64            `json:"emails_failed"`
	EmailsRetried     int64            `json:"emails_retried"`
	EmailsQueued      int64            `json:"emails_queued"`
	PushSent          int64            `json:"push_sent"`
	PushFailed        int64            `json:"push_failed"`
	InAppSent         int64            `json:"in_app_sent"`
	SMSSent           int64            `json:"sms_sent"`
	SMSFailed         int64            `json:"sms_failed"`
	AvgProcessingTime time.Duration    `json:"avg_processing_time"`
	QueueSize         int64            `json:"queue_size"`
	ByType            map[string]int64 `json:"by_type"`
	ByPriority        map[string]int64 `json:"by_priority"`

	mutex sync.RWMutex
}

// EmailService service d'envoi d'emails
type EmailService struct {
	config *NotificationConfig
	auth   smtp.Auth
	logger *zap.Logger
}

// TemplateEngine moteur de templates
type TemplateEngine struct {
	templates map[string]*NotificationTemplate
	logger    *zap.Logger
	mutex     sync.RWMutex
}

// NotificationTemplate template de notification
type NotificationTemplate struct {
	Name      string            `json:"name"`
	Subject   string            `json:"subject"`
	Body      string            `json:"body"`
	HTMLBody  string            `json:"html_body"`
	Language  string            `json:"language"`
	Variables []string          `json:"variables"`
	Metadata  map[string]string `json:"metadata"`
}

// Sujets NATS pour les notifications
const (
	NotificationSubjectEmail = "notifications.email"
	NotificationSubjectPush  = "notifications.push"
	NotificationSubjectInApp = "notifications.in_app"
	NotificationSubjectSMS   = "notifications.sms"
	NotificationSubjectDLQ   = "notifications.dlq"
	NotificationSubjectRetry = "notifications.retry"
)

// NewNotificationQueueService crée un nouveau service de queue de notifications
func NewNotificationQueueService(natsService *NATSService, config *NotificationConfig, logger *zap.Logger) (*NotificationQueueService, error) {
	if config == nil {
		config = &NotificationConfig{
			SMTPHost:        "smtp.gmail.com",
			SMTPPort:        587,
			FromEmail:       "noreply@veza.com",
			FromName:        "Veza",
			MaxRetries:      3,
			RetryDelay:      1 * time.Minute,
			RetryBackoff:    2.0,
			EmailRateLimit:  100,
			EmailRateWindow: 1 * time.Minute,
			TemplateDir:     "./templates",
			DefaultLanguage: "fr",
			WorkerCount:     5,
			BatchSize:       10,
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Créer le service email
	emailService := &EmailService{
		config: config,
		logger: logger,
	}

	if config.SMTPUsername != "" && config.SMTPPassword != "" {
		emailService.auth = smtp.PlainAuth("", config.SMTPUsername, config.SMTPPassword, config.SMTPHost)
	}

	// Créer le moteur de templates
	templateEngine := &TemplateEngine{
		templates: make(map[string]*NotificationTemplate),
		logger:    logger,
	}

	service := &NotificationQueueService{
		natsService:    natsService,
		emailService:   emailService,
		templateEngine: templateEngine,
		logger:         logger,
		config:         config,
		metrics:        &NotificationMetrics{ByType: make(map[string]int64), ByPriority: make(map[string]int64)},
		ctx:            ctx,
		cancel:         cancel,
	}

	// Charger les templates par défaut
	if err := service.loadDefaultTemplates(); err != nil {
		logger.Warn("Failed to load default templates", zap.Error(err))
	}

	// Démarrer les workers
	if err := service.startWorkers(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to start workers: %w", err)
	}

	return service, nil
}

// ============================================================================
// GESTION DES QUEUES
// ============================================================================

// QueueEmail met un email en queue
func (n *NotificationQueueService) QueueEmail(ctx context.Context, recipients []string, subject, template string, data map[string]interface{}) error {
	return n.QueueNotification(ctx, &NotificationMessage{
		ID:         n.generateNotificationID(),
		Type:       NotificationEmail,
		Recipients: recipients,
		Subject:    subject,
		Template:   template,
		Data:       data,
		Priority:   NotificationPriorityNormal,
		Language:   n.config.DefaultLanguage,
		CreatedAt:  time.Now(),
	})
}

// QueueEmailWithPriority met un email en queue avec priorité
func (n *NotificationQueueService) QueueEmailWithPriority(ctx context.Context, recipients []string, subject, template string, data map[string]interface{}, priority NotificationPriority) error {
	return n.QueueNotification(ctx, &NotificationMessage{
		ID:         n.generateNotificationID(),
		Type:       NotificationEmail,
		Recipients: recipients,
		Subject:    subject,
		Template:   template,
		Data:       data,
		Priority:   priority,
		Language:   n.config.DefaultLanguage,
		CreatedAt:  time.Now(),
	})
}

// QueueScheduledEmail met un email programmé en queue
func (n *NotificationQueueService) QueueScheduledEmail(ctx context.Context, recipients []string, subject, template string, data map[string]interface{}, scheduledAt time.Time) error {
	return n.QueueNotification(ctx, &NotificationMessage{
		ID:          n.generateNotificationID(),
		Type:        NotificationEmail,
		Recipients:  recipients,
		Subject:     subject,
		Template:    template,
		Data:        data,
		Priority:    NotificationPriorityNormal,
		Language:    n.config.DefaultLanguage,
		ScheduledAt: &scheduledAt,
		CreatedAt:   time.Now(),
	})
}

// QueueNotification met une notification en queue
func (n *NotificationQueueService) QueueNotification(ctx context.Context, notification *NotificationMessage) error {
	if notification == nil {
		return fmt.Errorf("notification cannot be nil")
	}

	// Valider la notification
	if err := n.validateNotification(notification); err != nil {
		return fmt.Errorf("invalid notification: %w", err)
	}

	// Déterminer le sujet NATS
	subject := n.getSubjectForNotificationType(notification.Type)

	// Créer l'événement
	event := &Event{
		ID:        notification.ID,
		Type:      EventType(fmt.Sprintf("notification.%s.queued", notification.Type)),
		Source:    "notification_queue_service",
		Subject:   subject,
		Data:      notification,
		Priority:  n.convertPriority(notification.Priority),
		Timestamp: time.Now(),
		UserID:    notification.UserID,
	}

	// Publier l'événement
	if err := n.natsService.PublishEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to queue notification: %w", err)
	}

	// Enregistrer les métriques
	n.recordNotificationQueued(notification.Type, notification.Priority)

	n.logger.Debug("Notification queued",
		zap.String("id", notification.ID),
		zap.String("type", string(notification.Type)),
		zap.String("priority", string(notification.Priority)),
		zap.Int("recipients", len(notification.Recipients)))

	return nil
}

// ============================================================================
// WORKERS
// ============================================================================

// startWorkers démarre les workers de traitement
func (n *NotificationQueueService) startWorkers() error {
	// Worker pour emails
	if err := n.natsService.SubscribeToSubject(NotificationSubjectEmail, n.handleEmailNotification); err != nil {
		return fmt.Errorf("failed to subscribe to email notifications: %w", err)
	}

	// Worker pour push notifications
	if err := n.natsService.SubscribeToSubject(NotificationSubjectPush, n.handlePushNotification); err != nil {
		return fmt.Errorf("failed to subscribe to push notifications: %w", err)
	}

	// Worker pour in-app notifications
	if err := n.natsService.SubscribeToSubject(NotificationSubjectInApp, n.handleInAppNotification); err != nil {
		return fmt.Errorf("failed to subscribe to in-app notifications: %w", err)
	}

	// Worker pour SMS
	if err := n.natsService.SubscribeToSubject(NotificationSubjectSMS, n.handleSMSNotification); err != nil {
		return fmt.Errorf("failed to subscribe to SMS notifications: %w", err)
	}

	// Worker pour retry
	if err := n.natsService.SubscribeToSubject(NotificationSubjectRetry, n.handleRetryNotification); err != nil {
		return fmt.Errorf("failed to subscribe to retry notifications: %w", err)
	}

	// Démarrer le worker de nettoyage
	go n.startCleanupWorker()

	// Démarrer le worker de métriques
	go n.startMetricsWorker()

	n.logger.Info("Notification workers started",
		zap.Int("worker_count", n.config.WorkerCount))

	return nil
}

// handleEmailNotification traite une notification email
func (n *NotificationQueueService) handleEmailNotification(ctx context.Context, event *Event) error {
	start := time.Now()

	// Parser la notification
	var notification NotificationMessage
	if err := json.Unmarshal(event.Data.([]byte), &notification); err != nil {
		return fmt.Errorf("failed to parse notification: %w", err)
	}

	// Vérifier si c'est une notification programmée
	if notification.ScheduledAt != nil && notification.ScheduledAt.After(time.Now()) {
		// Reprogrammer la notification
		return n.rescheduleNotification(&notification, *notification.ScheduledAt)
	}

	// Vérifier si la notification a expiré
	if notification.ExpiresAt != nil && notification.ExpiresAt.Before(time.Now()) {
		n.logger.Warn("Notification expired", zap.String("id", notification.ID))
		return nil
	}

	// Envoyer l'email
	if err := n.sendEmail(ctx, &notification); err != nil {
		// Retry si nécessaire
		if notification.RetryCount < n.config.MaxRetries {
			return n.retryNotification(&notification, err)
		}

		// Envoyer vers DLQ
		return n.sendToDeadLetterQueue(&notification, err)
	}

	// Enregistrer les métriques
	n.recordEmailSent(time.Since(start))

	n.logger.Debug("Email notification processed",
		zap.String("id", notification.ID),
		zap.Duration("duration", time.Since(start)))

	return nil
}

// handlePushNotification traite une notification push
func (n *NotificationQueueService) handlePushNotification(ctx context.Context, event *Event) error {
	start := time.Now()

	// Parser la notification
	var notification NotificationMessage
	if err := json.Unmarshal(event.Data.([]byte), &notification); err != nil {
		return fmt.Errorf("failed to parse notification: %w", err)
	}

	// TODO: Implémenter l'envoi de push notifications
	// Pour l'instant, on simule un envoi réussi
	n.logger.Info("Push notification sent", zap.String("id", notification.ID))

	// Enregistrer les métriques
	n.recordPushSent(time.Since(start))

	return nil
}

// handleInAppNotification traite une notification in-app
func (n *NotificationQueueService) handleInAppNotification(ctx context.Context, event *Event) error {
	// TODO: Implémenter l'envoi de notifications in-app via WebSocket
	// Pour l'instant, on simule un envoi réussi
	n.logger.Info("In-app notification sent", zap.String("event_id", event.ID))

	n.recordInAppSent()

	return nil
}

// handleSMSNotification traite une notification SMS
func (n *NotificationQueueService) handleSMSNotification(ctx context.Context, event *Event) error {
	// TODO: Implémenter l'envoi de SMS
	// Pour l'instant, on simule un envoi réussi
	n.logger.Info("SMS notification sent", zap.String("event_id", event.ID))

	n.recordSMSSent()

	return nil
}

// handleRetryNotification traite une notification retry
func (n *NotificationQueueService) handleRetryNotification(ctx context.Context, event *Event) error {
	// Parser la notification
	var notification NotificationMessage
	if err := json.Unmarshal(event.Data.([]byte), &notification); err != nil {
		return fmt.Errorf("failed to parse retry notification: %w", err)
	}

	// Incrémenter le compteur de retry
	notification.RetryCount++

	// Remettre en queue
	return n.QueueNotification(ctx, &notification)
}

// ============================================================================
// ENVOI D'EMAILS
// ============================================================================

// sendEmail envoie un email
func (n *NotificationQueueService) sendEmail(ctx context.Context, notification *NotificationMessage) error {
	// Générer le contenu de l'email
	subject, body, htmlBody, err := n.renderEmailTemplate(notification)
	if err != nil {
		return fmt.Errorf("failed to render email template: %w", err)
	}

	// Envoyer l'email à chaque destinataire
	for _, recipient := range notification.Recipients {
		if err := n.emailService.SendEmail(recipient, subject, body, htmlBody); err != nil {
			return fmt.Errorf("failed to send email to %s: %w", recipient, err)
		}
	}

	return nil
}

// SendEmail envoie un email via SMTP
func (e *EmailService) SendEmail(to, subject, body, htmlBody string) error {
	// Construire le message
	msg := e.buildEmailMessage(to, subject, body, htmlBody)

	// Envoyer via SMTP
	addr := fmt.Sprintf("%s:%d", e.config.SMTPHost, e.config.SMTPPort)

	if e.auth != nil {
		return smtp.SendMail(addr, e.auth, e.config.FromEmail, []string{to}, []byte(msg))
	}

	// Fallback sans authentification pour les tests
	return smtp.SendMail(addr, nil, e.config.FromEmail, []string{to}, []byte(msg))
}

// buildEmailMessage construit le message email
func (e *EmailService) buildEmailMessage(to, subject, body, htmlBody string) string {
	var msg strings.Builder

	// Headers
	msg.WriteString(fmt.Sprintf("From: %s <%s>\r\n", e.config.FromName, e.config.FromEmail))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", to))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")

	if htmlBody != "" {
		// Email multipart
		boundary := "boundary-" + fmt.Sprintf("%d", time.Now().Unix())
		msg.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=%s\r\n", boundary))
		msg.WriteString("\r\n")

		// Partie texte
		msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
		msg.WriteString("\r\n")
		msg.WriteString(body)
		msg.WriteString("\r\n")

		// Partie HTML
		msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		msg.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
		msg.WriteString("\r\n")
		msg.WriteString(htmlBody)
		msg.WriteString("\r\n")

		msg.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	} else {
		// Email texte simple
		msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
		msg.WriteString("\r\n")
		msg.WriteString(body)
	}

	return msg.String()
}

// ============================================================================
// TEMPLATES
// ============================================================================

// renderEmailTemplate rend un template d'email
func (n *NotificationQueueService) renderEmailTemplate(notification *NotificationMessage) (subject, body, htmlBody string, err error) {
	template := n.templateEngine.GetTemplate(notification.Template, notification.Language)
	if template == nil {
		return "", "", "", fmt.Errorf("template not found: %s", notification.Template)
	}

	// Remplacer les variables dans le sujet
	subject = n.replaceVariables(template.Subject, notification.Data)
	if notification.Subject != "" {
		subject = notification.Subject // Override si spécifié
	}

	// Remplacer les variables dans le corps
	body = n.replaceVariables(template.Body, notification.Data)

	// Remplacer les variables dans le HTML
	if template.HTMLBody != "" {
		htmlBody = n.replaceVariables(template.HTMLBody, notification.Data)
	}

	return subject, body, htmlBody, nil
}

// replaceVariables remplace les variables dans un template
func (n *NotificationQueueService) replaceVariables(template string, data map[string]interface{}) string {
	result := template

	for key, value := range data {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}

	return result
}

// GetTemplate retourne un template
func (t *TemplateEngine) GetTemplate(name, language string) *NotificationTemplate {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	// Chercher avec la langue spécifiée
	key := fmt.Sprintf("%s-%s", name, language)
	if template, exists := t.templates[key]; exists {
		return template
	}

	// Fallback vers le template par défaut
	if template, exists := t.templates[name]; exists {
		return template
	}

	return nil
}

// AddTemplate ajoute un template
func (t *TemplateEngine) AddTemplate(template *NotificationTemplate) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	key := template.Name
	if template.Language != "" {
		key = fmt.Sprintf("%s-%s", template.Name, template.Language)
	}

	t.templates[key] = template
}

// ============================================================================
// UTILITAIRES
// ============================================================================

// loadDefaultTemplates charge les templates par défaut
func (n *NotificationQueueService) loadDefaultTemplates() error {
	templates := []*NotificationTemplate{
		{
			Name:     "welcome",
			Subject:  "Bienvenue sur Veza, {{username}} !",
			Body:     "Bonjour {{username}},\n\nBienvenue sur Veza ! Votre compte a été créé avec succès.\n\nCordialement,\nL'équipe Veza",
			HTMLBody: "<h1>Bienvenue sur Veza, {{username}} !</h1><p>Votre compte a été créé avec succès.</p>",
			Language: "fr",
		},
		{
			Name:     "verification",
			Subject:  "Vérification de votre compte Veza",
			Body:     "Bonjour {{username}},\n\nVeuillez cliquer sur ce lien pour vérifier votre compte : {{verification_link}}\n\nCordialement,\nL'équipe Veza",
			HTMLBody: "<h1>Vérification de compte</h1><p><a href=\"{{verification_link}}\">Cliquez ici pour vérifier</a></p>",
			Language: "fr",
		},
		{
			Name:     "password_reset",
			Subject:  "Réinitialisation de votre mot de passe",
			Body:     "Bonjour {{username}},\n\nVeuillez cliquer sur ce lien pour réinitialiser votre mot de passe : {{reset_link}}\n\nCordialement,\nL'équipe Veza",
			HTMLBody: "<h1>Réinitialisation de mot de passe</h1><p><a href=\"{{reset_link}}\">Cliquez ici pour réinitialiser</a></p>",
			Language: "fr",
		},
	}

	for _, template := range templates {
		n.templateEngine.AddTemplate(template)
	}

	return nil
}

// validateNotification valide une notification
func (n *NotificationQueueService) validateNotification(notification *NotificationMessage) error {
	if notification.ID == "" {
		return fmt.Errorf("notification ID is required")
	}

	if len(notification.Recipients) == 0 {
		return fmt.Errorf("at least one recipient is required")
	}

	if notification.Template == "" {
		return fmt.Errorf("template is required")
	}

	return nil
}

// getSubjectForNotificationType retourne le sujet NATS pour un type
func (n *NotificationQueueService) getSubjectForNotificationType(notifType NotificationType) string {
	switch {
	case strings.HasPrefix(string(notifType), "email"):
		return NotificationSubjectEmail
	case strings.HasPrefix(string(notifType), "push"):
		return NotificationSubjectPush
	case strings.HasPrefix(string(notifType), "in_app"):
		return NotificationSubjectInApp
	case strings.HasPrefix(string(notifType), "sms"):
		return NotificationSubjectSMS
	default:
		return NotificationSubjectEmail
	}
}

// convertPriority convertit la priorité de notification en priorité d'événement
func (n *NotificationQueueService) convertPriority(priority NotificationPriority) EventPriority {
	switch priority {
	case NotificationPriorityLow:
		return PriorityLow
	case NotificationPriorityNormal:
		return PriorityNormal
	case NotificationPriorityHigh:
		return PriorityHigh
	case NotificationPriorityCritical:
		return PriorityCritical
	default:
		return PriorityNormal
	}
}

// generateNotificationID génère un ID unique pour la notification
func (n *NotificationQueueService) generateNotificationID() string {
	return fmt.Sprintf("notif_%d_%d", time.Now().UnixNano(), n.metrics.EmailsQueued)
}

// rescheduleNotification reprogramme une notification
func (n *NotificationQueueService) rescheduleNotification(notification *NotificationMessage, scheduledAt time.Time) error {
	delay := time.Until(scheduledAt)

	go func() {
		time.Sleep(delay)

		// Remettre en queue
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := n.QueueNotification(ctx, notification); err != nil {
			n.logger.Error("Failed to reschedule notification",
				zap.String("id", notification.ID),
				zap.Error(err))
		}
	}()

	return nil
}

// retryNotification remet une notification en retry
func (n *NotificationQueueService) retryNotification(notification *NotificationMessage, err error) error {
	notification.RetryCount++

	// Calculer le délai de retry avec backoff exponentiel
	delay := time.Duration(float64(n.config.RetryDelay) * float64(notification.RetryCount) * n.config.RetryBackoff)

	go func() {
		time.Sleep(delay)

		// Créer l'événement de retry
		event := &Event{
			ID:        notification.ID + "_retry",
			Type:      EventType("notification.retry"),
			Source:    "notification_queue_service",
			Subject:   NotificationSubjectRetry,
			Data:      notification,
			Priority:  PriorityNormal,
			Timestamp: time.Now(),
		}

		// Publier l'événement de retry
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := n.natsService.PublishEvent(ctx, event); err != nil {
			n.logger.Error("Failed to retry notification",
				zap.String("id", notification.ID),
				zap.Error(err))
		}
	}()

	n.recordNotificationRetried()

	return nil
}

// sendToDeadLetterQueue envoie une notification vers la DLQ
func (n *NotificationQueueService) sendToDeadLetterQueue(notification *NotificationMessage, err error) error {
	// Créer l'événement DLQ
	event := &Event{
		ID:        notification.ID + "_dlq",
		Type:      EventType("notification.dlq"),
		Source:    "notification_queue_service",
		Subject:   NotificationSubjectDLQ,
		Data:      notification,
		Priority:  PriorityNormal,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"error":       err.Error(),
			"retry_count": notification.RetryCount,
			"failed_at":   time.Now(),
		},
	}

	// Publier l'événement DLQ
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := n.natsService.PublishEvent(ctx, event); err != nil {
		n.logger.Error("Failed to send notification to DLQ",
			zap.String("id", notification.ID),
			zap.Error(err))
		return err
	}

	n.logger.Warn("Notification sent to DLQ",
		zap.String("id", notification.ID),
		zap.String("error", err.Error()),
		zap.Int("retry_count", notification.RetryCount))

	return nil
}

// ============================================================================
// WORKERS DE MAINTENANCE
// ============================================================================

// startCleanupWorker démarre le worker de nettoyage
func (n *NotificationQueueService) startCleanupWorker() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// TODO: Nettoyer les notifications expirées
			n.logger.Debug("Cleanup worker running")

		case <-n.ctx.Done():
			return
		}
	}
}

// startMetricsWorker démarre le worker de métriques
func (n *NotificationQueueService) startMetricsWorker() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			metrics := n.GetMetrics()
			n.logger.Info("Notification metrics",
				zap.Int64("emails_sent", metrics.EmailsSent),
				zap.Int64("emails_failed", metrics.EmailsFailed),
				zap.Int64("emails_queued", metrics.EmailsQueued),
				zap.Int64("queue_size", metrics.QueueSize))

		case <-n.ctx.Done():
			return
		}
	}
}

// ============================================================================
// MÉTRIQUES
// ============================================================================

func (n *NotificationQueueService) recordNotificationQueued(notifType NotificationType, priority NotificationPriority) {
	n.metrics.mutex.Lock()
	defer n.metrics.mutex.Unlock()

	if strings.HasPrefix(string(notifType), "email") {
		n.metrics.EmailsQueued++
	}

	n.metrics.ByType[string(notifType)]++
	n.metrics.ByPriority[string(priority)]++
}

func (n *NotificationQueueService) recordEmailSent(duration time.Duration) {
	n.metrics.mutex.Lock()
	defer n.metrics.mutex.Unlock()

	n.metrics.EmailsSent++
	n.metrics.AvgProcessingTime = (n.metrics.AvgProcessingTime + duration) / 2
}

func (n *NotificationQueueService) recordEmailFailed() {
	n.metrics.mutex.Lock()
	defer n.metrics.mutex.Unlock()

	n.metrics.EmailsFailed++
}

func (n *NotificationQueueService) recordNotificationRetried() {
	n.metrics.mutex.Lock()
	defer n.metrics.mutex.Unlock()

	n.metrics.EmailsRetried++
}

func (n *NotificationQueueService) recordPushSent(duration time.Duration) {
	n.metrics.mutex.Lock()
	defer n.metrics.mutex.Unlock()

	n.metrics.PushSent++
	n.metrics.AvgProcessingTime = (n.metrics.AvgProcessingTime + duration) / 2
}

func (n *NotificationQueueService) recordPushFailed() {
	n.metrics.mutex.Lock()
	defer n.metrics.mutex.Unlock()

	n.metrics.PushFailed++
}

func (n *NotificationQueueService) recordInAppSent() {
	n.metrics.mutex.Lock()
	defer n.metrics.mutex.Unlock()

	n.metrics.InAppSent++
}

func (n *NotificationQueueService) recordSMSSent() {
	n.metrics.mutex.Lock()
	defer n.metrics.mutex.Unlock()

	n.metrics.SMSSent++
}

func (n *NotificationQueueService) recordSMSFailed() {
	n.metrics.mutex.Lock()
	defer n.metrics.mutex.Unlock()

	n.metrics.SMSFailed++
}

// GetMetrics retourne les métriques de notification
func (n *NotificationQueueService) GetMetrics() *NotificationMetrics {
	n.metrics.mutex.RLock()
	defer n.metrics.mutex.RUnlock()

	// Copier les métriques
	byType := make(map[string]int64)
	for k, v := range n.metrics.ByType {
		byType[k] = v
	}

	byPriority := make(map[string]int64)
	for k, v := range n.metrics.ByPriority {
		byPriority[k] = v
	}

	return &NotificationMetrics{
		EmailsSent:        n.metrics.EmailsSent,
		EmailsFailed:      n.metrics.EmailsFailed,
		EmailsRetried:     n.metrics.EmailsRetried,
		EmailsQueued:      n.metrics.EmailsQueued,
		PushSent:          n.metrics.PushSent,
		PushFailed:        n.metrics.PushFailed,
		InAppSent:         n.metrics.InAppSent,
		SMSSent:           n.metrics.SMSSent,
		SMSFailed:         n.metrics.SMSFailed,
		AvgProcessingTime: n.metrics.AvgProcessingTime,
		QueueSize:         n.metrics.QueueSize,
		ByType:            byType,
		ByPriority:        byPriority,
	}
}

// Close ferme proprement le service
func (n *NotificationQueueService) Close() error {
	n.cancel()

	// Attendre que tous les workers terminent
	done := make(chan struct{})
	go func() {
		n.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		n.logger.Warn("Timeout waiting for notification workers to finish")
	}

	n.logger.Info("Notification queue service closed")
	return nil
}
