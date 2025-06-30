package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// AuditEvent représente un événement d'audit
type AuditEvent struct {
	Timestamp    time.Time              `json:"timestamp"`
	EventType    string                 `json:"event_type"`
	UserID       string                 `json:"user_id,omitempty"`
	IP           string                 `json:"ip"`
	UserAgent    string                 `json:"user_agent"`
	Method       string                 `json:"method"`
	Path         string                 `json:"path"`
	StatusCode   int                    `json:"status_code"`
	Duration     time.Duration          `json:"duration_ms"`
	RequestSize  int64                  `json:"request_size"`
	ResponseSize int                    `json:"response_size"`
	Headers      map[string]string      `json:"headers,omitempty"`
	Body         string                 `json:"body,omitempty"`
	Error        string                 `json:"error,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// AuditConfig représente la configuration d'audit
type AuditConfig struct {
	LogLevel      zapcore.Level
	LogBody       bool
	LogHeaders    bool
	MaxBodySize   int64
	SensitiveData []string
	SkipPaths     []string
	LogSuccessful bool
	LogErrors     bool
}

// DefaultAuditConfig retourne la configuration d'audit par défaut
func DefaultAuditConfig() AuditConfig {
	return AuditConfig{
		LogLevel:      zapcore.InfoLevel,
		LogBody:       true,
		LogHeaders:    true,
		MaxBodySize:   1024, // 1KB max body log
		SensitiveData: []string{"password", "token", "secret", "key"},
		SkipPaths:     []string{"/health", "/metrics", "/favicon.ico"},
		LogSuccessful: true,
		LogErrors:     true,
	}
}

// responseBodyWriter capture la réponse pour l'audit
type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// AuditLogger retourne le middleware d'audit logging
func AuditLogger() gin.HandlerFunc {
	return AuditWithConfig(DefaultAuditConfig())
}

// AuditWithConfig retourne le middleware d'audit avec configuration personnalisée
func AuditWithConfig(config AuditConfig) gin.HandlerFunc {
	// Logger d'audit dédié
	auditLogger := createAuditLogger()

	return func(c *gin.Context) {
		// Skip si le path est dans la liste d'exclusion
		if shouldSkipPath(c.Request.URL.Path, config.SkipPaths) {
			c.Next()
			return
		}

		start := time.Now()

		// Capturer le body de la requête
		var requestBody string
		if config.LogBody && c.Request.Body != nil {
			requestBody = captureRequestBody(c, config.MaxBodySize)
		}

		// Capturer la réponse
		responseBuffer := &bytes.Buffer{}
		writer := &responseBodyWriter{
			ResponseWriter: c.Writer,
			body:           responseBuffer,
		}
		c.Writer = writer

		// Traiter la requête
		c.Next()

		// Calculer la durée
		duration := time.Since(start)

		// Créer l'événement d'audit
		event := createAuditEvent(c, requestBody, responseBuffer.String(), duration, config)

		// Logger selon la configuration
		logAuditEvent(auditLogger, event, config)
	}
}

// createAuditLogger crée un logger spécialisé pour l'audit
func createAuditLogger() *zap.Logger {
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"stdout", "logs/audit.log"}
	config.ErrorOutputPaths = []string{"stderr", "logs/audit_errors.log"}

	logger, err := config.Build()
	if err != nil {
		// Fallback sur logger simple
		logger = zap.NewNop()
	}

	return logger
}

// shouldSkipPath vérifie si le path doit être ignoré
func shouldSkipPath(path string, skipPaths []string) bool {
	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// captureRequestBody capture le body de la requête
func captureRequestBody(c *gin.Context, maxSize int64) string {
	if c.Request.Body == nil {
		return ""
	}

	// Lire le body
	body, err := io.ReadAll(io.LimitReader(c.Request.Body, maxSize))
	if err != nil {
		return ""
	}

	// Restaurer le body pour les handlers suivants
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	return sanitizeBody(string(body))
}

// sanitizeBody retire les données sensibles du body
func sanitizeBody(body string) string {
	sensitiveFields := []string{"password", "token", "secret", "key", "credential"}

	var data interface{}
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		// Si ce n'est pas du JSON, on retourne tel quel
		return body
	}

	// Sanitize récursivement
	sanitized := sanitizeObject(data, sensitiveFields)

	result, err := json.Marshal(sanitized)
	if err != nil {
		return "[sanitization_error]"
	}

	return string(result)
}

// sanitizeObject retire les champs sensibles d'un objet
func sanitizeObject(obj interface{}, sensitiveFields []string) interface{} {
	switch v := obj.(type) {
	case map[string]interface{}:
		sanitized := make(map[string]interface{})
		for key, value := range v {
			if isSensitiveField(key, sensitiveFields) {
				sanitized[key] = "[REDACTED]"
			} else {
				sanitized[key] = sanitizeObject(value, sensitiveFields)
			}
		}
		return sanitized
	case []interface{}:
		sanitized := make([]interface{}, len(v))
		for i, item := range v {
			sanitized[i] = sanitizeObject(item, sensitiveFields)
		}
		return sanitized
	default:
		return v
	}
}

// isSensitiveField vérifie si un champ est sensible
func isSensitiveField(field string, sensitiveFields []string) bool {
	fieldLower := strings.ToLower(field)
	for _, sensitive := range sensitiveFields {
		if strings.Contains(fieldLower, strings.ToLower(sensitive)) {
			return true
		}
	}
	return false
}

// createAuditEvent crée un événement d'audit
func createAuditEvent(c *gin.Context, requestBody, responseBody string, duration time.Duration, config AuditConfig) AuditEvent {
	event := AuditEvent{
		Timestamp:    time.Now(),
		EventType:    determineEventType(c),
		IP:           c.ClientIP(),
		UserAgent:    c.GetHeader("User-Agent"),
		Method:       c.Request.Method,
		Path:         c.Request.URL.Path,
		StatusCode:   c.Writer.Status(),
		Duration:     duration,
		RequestSize:  c.Request.ContentLength,
		ResponseSize: c.Writer.Size(),
	}

	// Ajouter l'ID utilisateur si disponible
	if userID, exists := c.Get("user_id"); exists {
		event.UserID = fmt.Sprintf("%v", userID)
	}

	// Ajouter les headers si configuré
	if config.LogHeaders {
		event.Headers = sanitizeHeaders(c.Request.Header)
	}

	// Ajouter le body si configuré
	if config.LogBody && len(requestBody) > 0 {
		event.Body = requestBody
	}

	// Ajouter les erreurs si présentes
	if len(c.Errors) > 0 {
		event.Error = c.Errors.String()
	}

	// Métadonnées additionnelles
	event.Metadata = map[string]interface{}{
		"referer":      c.GetHeader("Referer"),
		"content_type": c.GetHeader("Content-Type"),
		"query_params": c.Request.URL.RawQuery,
	}

	return event
}

// determineEventType détermine le type d'événement
func determineEventType(c *gin.Context) string {
	path := c.Request.URL.Path
	method := c.Request.Method

	// Événements d'authentification
	if strings.Contains(path, "/auth/") {
		switch {
		case strings.Contains(path, "/login"):
			return "AUTH_LOGIN"
		case strings.Contains(path, "/register"):
			return "AUTH_REGISTER"
		case strings.Contains(path, "/logout"):
			return "AUTH_LOGOUT"
		case strings.Contains(path, "/refresh"):
			return "AUTH_REFRESH"
		}
	}

	// Événements CRUD
	switch method {
	case "POST":
		return "CREATE"
	case "PUT", "PATCH":
		return "UPDATE"
	case "DELETE":
		return "DELETE"
	case "GET":
		return "READ"
	default:
		return "REQUEST"
	}
}

// sanitizeHeaders retire les headers sensibles
func sanitizeHeaders(headers http.Header) map[string]string {
	sanitized := make(map[string]string)
	sensitiveHeaders := []string{"authorization", "cookie", "x-api-key", "x-auth-token"}

	for key, values := range headers {
		keyLower := strings.ToLower(key)
		isSensitive := false

		for _, sensitive := range sensitiveHeaders {
			if strings.Contains(keyLower, sensitive) {
				isSensitive = true
				break
			}
		}

		if isSensitive {
			sanitized[key] = "[REDACTED]"
		} else if len(values) > 0 {
			sanitized[key] = values[0]
		}
	}

	return sanitized
}

// logAuditEvent enregistre l'événement d'audit
func logAuditEvent(logger *zap.Logger, event AuditEvent, config AuditConfig) {
	// Déterminer le niveau de log
	var level zapcore.Level
	if event.StatusCode >= 500 {
		level = zapcore.ErrorLevel
	} else if event.StatusCode >= 400 {
		level = zapcore.WarnLevel
	} else {
		level = zapcore.InfoLevel
	}

	// Skip si on ne veut pas logger les succès/erreurs
	if !config.LogSuccessful && event.StatusCode < 400 {
		return
	}
	if !config.LogErrors && event.StatusCode >= 400 {
		return
	}

	// Fields structurés pour Zap
	fields := []zap.Field{
		zap.String("event_type", event.EventType),
		zap.String("user_id", event.UserID),
		zap.String("ip", event.IP),
		zap.String("method", event.Method),
		zap.String("path", event.Path),
		zap.Int("status_code", event.StatusCode),
		zap.Duration("duration", event.Duration),
		zap.Int64("request_size", event.RequestSize),
		zap.Int("response_size", event.ResponseSize),
	}

	// Ajouter des champs optionnels
	if event.Error != "" {
		fields = append(fields, zap.String("error", event.Error))
	}
	if event.Body != "" {
		fields = append(fields, zap.String("request_body", event.Body))
	}

	// Logger avec le niveau approprié
	logger.Log(level, "API Audit Event", fields...)
}
