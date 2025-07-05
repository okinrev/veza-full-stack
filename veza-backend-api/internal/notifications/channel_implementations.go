package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"text/template"
	"time"

	"go.uber.org/zap"
)

// ============================================================================
// EMAIL SERVICE IMPLEMENTATION
// ============================================================================

// EmailServiceImpl implémentation du service email
type EmailServiceImpl struct {
	logger    *zap.Logger
	smtpHost  string
	smtpPort  string
	smtpUser  string
	smtpPass  string
	fromEmail string
	fromName  string
	templates map[NotificationType]*EmailTemplate
}

// EmailTemplate template d'email
type EmailTemplate struct {
	Subject  string
	HTMLBody string
	TextBody string
}

// NewEmailService crée un nouveau service email
func NewEmailService(logger *zap.Logger, smtpHost, smtpPort, smtpUser, smtpPass, fromEmail, fromName string) *EmailServiceImpl {
	return &EmailServiceImpl{
		logger:    logger,
		smtpHost:  smtpHost,
		smtpPort:  smtpPort,
		smtpUser:  smtpUser,
		smtpPass:  smtpPass,
		fromEmail: fromEmail,
		fromName:  fromName,
		templates: make(map[NotificationType]*EmailTemplate),
	}
}

// SendEmail envoie un email
func (e *EmailServiceImpl) SendEmail(ctx context.Context, notification *Notification) error {
	// Obtenir le template
	emailTemplate := e.getTemplate(notification.Type)

	// Rendre le template avec les données
	subject, htmlBody, textBody, err := e.renderTemplate(emailTemplate, notification)
	if err != nil {
		return fmt.Errorf("failed to render email template: %w", err)
	}

	// Composer l'email
	message := e.composeEmail(notification.UserID, subject, htmlBody, textBody)

	// Envoyer via SMTP
	auth := smtp.PlainAuth("", e.smtpUser, e.smtpPass, e.smtpHost)
	addr := fmt.Sprintf("%s:%s", e.smtpHost, e.smtpPort)

	// TODO: Obtenir l'email de l'utilisateur depuis la base de données
	userEmail := fmt.Sprintf("user_%s@example.com", notification.UserID) // Temporaire

	err = smtp.SendMail(addr, auth, e.fromEmail, []string{userEmail}, []byte(message))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	e.logger.Info("📧 Email sent successfully",
		zap.String("user_id", notification.UserID),
		zap.String("notification_id", notification.ID),
		zap.String("type", string(notification.Type)))

	return nil
}

// getTemplate obtient le template pour un type de notification
func (e *EmailServiceImpl) getTemplate(notificationType NotificationType) *EmailTemplate {
	if template, exists := e.templates[notificationType]; exists {
		return template
	}

	// Template par défaut
	return &EmailTemplate{
		Subject:  "{{.Title}}",
		HTMLBody: "<h2>{{.Title}}</h2><p>{{.Message}}</p>",
		TextBody: "{{.Title}}\n\n{{.Message}}",
	}
}

// renderTemplate rend un template avec les données de notification
func (e *EmailServiceImpl) renderTemplate(emailTemplate *EmailTemplate, notification *Notification) (string, string, string, error) {
	data := map[string]interface{}{
		"Title":     notification.Title,
		"Message":   notification.Message,
		"UserID":    notification.UserID,
		"Type":      notification.Type,
		"Priority":  notification.Priority,
		"CreatedAt": notification.CreatedAt.Format("2006-01-02 15:04:05"),
		"Data":      notification.Data,
	}

	// Rendre le sujet
	subjectTmpl, err := template.New("subject").Parse(emailTemplate.Subject)
	if err != nil {
		return "", "", "", err
	}
	var subjectBuf bytes.Buffer
	if err := subjectTmpl.Execute(&subjectBuf, data); err != nil {
		return "", "", "", err
	}

	// Rendre le corps HTML
	htmlTmpl, err := template.New("html").Parse(emailTemplate.HTMLBody)
	if err != nil {
		return "", "", "", err
	}
	var htmlBuf bytes.Buffer
	if err := htmlTmpl.Execute(&htmlBuf, data); err != nil {
		return "", "", "", err
	}

	// Rendre le corps texte
	textTmpl, err := template.New("text").Parse(emailTemplate.TextBody)
	if err != nil {
		return "", "", "", err
	}
	var textBuf bytes.Buffer
	if err := textTmpl.Execute(&textBuf, data); err != nil {
		return "", "", "", err
	}

	return subjectBuf.String(), htmlBuf.String(), textBuf.String(), nil
}

// composeEmail compose le message email complet
func (e *EmailServiceImpl) composeEmail(userID, subject, htmlBody, textBody string) string {
	boundary := "veza-notification-boundary"

	message := fmt.Sprintf(`From: %s <%s>
To: %s
Subject: %s
MIME-Version: 1.0
Content-Type: multipart/alternative; boundary="%s"

--%s
Content-Type: text/plain; charset=UTF-8

%s

--%s
Content-Type: text/html; charset=UTF-8

%s

--%s--
`, e.fromName, e.fromEmail, userID, subject, boundary, boundary, textBody, boundary, htmlBody, boundary)

	return message
}

// ============================================================================
// SMS SERVICE IMPLEMENTATION
// ============================================================================

// SMSServiceImpl implémentation du service SMS
type SMSServiceImpl struct {
	logger *zap.Logger
	apiKey string
	apiURL string
	sender string
}

// NewSMSService crée un nouveau service SMS
func NewSMSService(logger *zap.Logger, apiKey, apiURL, sender string) *SMSServiceImpl {
	return &SMSServiceImpl{
		logger: logger,
		apiKey: apiKey,
		apiURL: apiURL,
		sender: sender,
	}
}

// SendSMS envoie un SMS
func (s *SMSServiceImpl) SendSMS(ctx context.Context, notification *Notification) error {
	// Composer le message SMS (limité à 160 caractères)
	message := s.formatSMSMessage(notification)

	// TODO: Implémenter la logique SMS réelle
	phoneNumber := "+33600000000" // Temporaire

	// Préparer la requête API
	payload := map[string]interface{}{
		"to":      phoneNumber,
		"from":    s.sender,
		"message": message,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal SMS payload: %w", err)
	}

	// Créer la requête HTTP
	req, err := http.NewRequestWithContext(ctx, "POST", s.apiURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create SMS request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	// Envoyer la requête
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send SMS request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("SMS API returned error: %d", resp.StatusCode)
	}

	s.logger.Info("📱 SMS sent successfully",
		zap.String("user_id", notification.UserID),
		zap.String("notification_id", notification.ID),
		zap.String("phone", phoneNumber))

	return nil
}

// formatSMSMessage formate le message pour SMS
func (s *SMSServiceImpl) formatSMSMessage(notification *Notification) string {
	message := fmt.Sprintf("%s: %s", notification.Title, notification.Message)

	// Limiter à 160 caractères
	if len(message) > 160 {
		message = message[:157] + "..."
	}

	return message
}

// ============================================================================
// PUSH NOTIFICATION SERVICE IMPLEMENTATION
// ============================================================================

// PushServiceImpl implémentation du service push notifications
type PushServiceImpl struct {
	logger    *zap.Logger
	fcmAPIKey string
	apnsKey   string
}

// NewPushService crée un nouveau service push
func NewPushService(logger *zap.Logger, fcmAPIKey, apnsKey string) *PushServiceImpl {
	return &PushServiceImpl{
		logger:    logger,
		fcmAPIKey: fcmAPIKey,
		apnsKey:   apnsKey,
	}
}

// SendPushNotification envoie une push notification
func (p *PushServiceImpl) SendPushNotification(ctx context.Context, notification *Notification) error {
	// TODO: Obtenir les tokens de device de l'utilisateur depuis la DB
	deviceTokens := []string{"example_device_token"} // Temporaire

	for _, token := range deviceTokens {
		if err := p.sendToDevice(ctx, notification, token); err != nil {
			p.logger.Error("Failed to send push to device",
				zap.String("token", token),
				zap.Error(err))
		}
	}

	p.logger.Info("🔔 Push notifications sent",
		zap.String("user_id", notification.UserID),
		zap.String("notification_id", notification.ID),
		zap.Int("devices", len(deviceTokens)))

	return nil
}

// sendToDevice envoie une push notification à un device spécifique
func (p *PushServiceImpl) sendToDevice(ctx context.Context, notification *Notification, deviceToken string) error {
	// Préparer la payload FCM
	payload := map[string]interface{}{
		"to": deviceToken,
		"notification": map[string]interface{}{
			"title": notification.Title,
			"body":  notification.Message,
		},
		"data": map[string]interface{}{
			"notification_id":   notification.ID,
			"notification_type": string(notification.Type),
			"priority":          string(notification.Priority),
		},
	}

	if notification.Data != nil {
		if data, ok := payload["data"].(map[string]interface{}); ok {
			for k, v := range notification.Data {
				data[k] = v
			}
		}
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal push payload: %w", err)
	}

	// Envoyer via FCM
	req, err := http.NewRequestWithContext(ctx, "POST", "https://fcm.googleapis.com/fcm/send", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create push request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "key="+p.fcmAPIKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send push request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("FCM API returned error: %d", resp.StatusCode)
	}

	return nil
}

// ============================================================================
// WEBHOOK SERVICE IMPLEMENTATION
// ============================================================================

// WebhookServiceImpl implémentation du service webhook
type WebhookServiceImpl struct {
	logger *zap.Logger
}

// NewWebhookService crée un nouveau service webhook
func NewWebhookService(logger *zap.Logger) *WebhookServiceImpl {
	return &WebhookServiceImpl{
		logger: logger,
	}
}

// SendWebhook envoie une notification via webhook
func (w *WebhookServiceImpl) SendWebhook(ctx context.Context, notification *Notification) error {
	// TODO: Obtenir les URLs de webhook de l'utilisateur depuis la DB
	webhookURLs := []string{"https://example.com/webhook"} // Temporaire

	for _, url := range webhookURLs {
		if err := w.sendToWebhook(ctx, notification, url); err != nil {
			w.logger.Error("Failed to send webhook",
				zap.String("url", url),
				zap.Error(err))
		}
	}

	w.logger.Info("🔗 Webhooks sent",
		zap.String("user_id", notification.UserID),
		zap.String("notification_id", notification.ID),
		zap.Int("webhooks", len(webhookURLs)))

	return nil
}

// sendToWebhook envoie vers une URL de webhook spécifique
func (w *WebhookServiceImpl) sendToWebhook(ctx context.Context, notification *Notification, webhookURL string) error {
	payloadBytes, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Veza-Notifications/1.0")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned error: %d", resp.StatusCode)
	}

	return nil
}
