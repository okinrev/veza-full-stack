package messagequeue

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// EventSourcingService service pour l'event sourcing et audit logs
type EventSourcingService struct {
	natsService *NATSService
	logger      *zap.Logger
	config      *EventSourcingConfig

	// Stockage des événements
	eventStore EventStore

	// Projections
	projections map[string]Projection

	// Métriques
	metrics *EventSourcingMetrics

	// Contrôle de lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	mutex  sync.RWMutex
}

// EventSourcingConfig configuration de l'event sourcing
type EventSourcingConfig struct {
	BatchSize          int           `json:"batch_size"`
	FlushInterval      time.Duration `json:"flush_interval"`
	RetentionPeriod    time.Duration `json:"retention_period"`
	CompressionEnabled bool          `json:"compression_enabled"`
	EncryptionEnabled  bool          `json:"encryption_enabled"`
	EnableProjections  bool          `json:"enable_projections"`
	EnableSnapshots    bool          `json:"enable_snapshots"`
	SnapshotInterval   int           `json:"snapshot_interval"`
}

// AuditEvent événement d'audit
type AuditEvent struct {
	ID            string                 `json:"id"`
	AggregateID   string                 `json:"aggregate_id"`
	AggregateType string                 `json:"aggregate_type"`
	EventType     AuditEventType         `json:"event_type"`
	EventData     map[string]interface{} `json:"event_data"`
	Metadata      AuditMetadata          `json:"metadata"`
	Timestamp     time.Time              `json:"timestamp"`
	Version       int64                  `json:"version"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
	CausationID   string                 `json:"causation_id,omitempty"`
	UserID        *int64                 `json:"user_id,omitempty"`
	SessionID     *string                `json:"session_id,omitempty"`
	IPAddress     string                 `json:"ip_address,omitempty"`
	UserAgent     string                 `json:"user_agent,omitempty"`
	Severity      AuditSeverity          `json:"severity"`
	Category      AuditCategory          `json:"category"`
}

// AuditEventType types d'événements d'audit
type AuditEventType string

const (
	// Événements d'authentification
	AuditUserLogin         AuditEventType = "user.login"
	AuditUserLogout        AuditEventType = "user.logout"
	AuditUserLoginFailed   AuditEventType = "user.login.failed"
	AuditPasswordChanged   AuditEventType = "user.password.changed"
	AuditTwoFactorEnabled  AuditEventType = "user.2fa.enabled"
	AuditTwoFactorDisabled AuditEventType = "user.2fa.disabled"

	// Événements de données
	AuditDataCreated  AuditEventType = "data.created"
	AuditDataUpdated  AuditEventType = "data.updated"
	AuditDataDeleted  AuditEventType = "data.deleted"
	AuditDataExported AuditEventType = "data.exported"
	AuditDataImported AuditEventType = "data.imported"
	AuditDataAccessed AuditEventType = "data.accessed"

	// Événements administratifs
	AuditAdminAction       AuditEventType = "admin.action"
	AuditPermissionChanged AuditEventType = "admin.permission.changed"
	AuditRoleAssigned      AuditEventType = "admin.role.assigned"
	AuditRoleRevoked       AuditEventType = "admin.role.revoked"
	AuditConfigChanged     AuditEventType = "admin.config.changed"

	// Événements de sécurité
	AuditSecurityBreach     AuditEventType = "security.breach"
	AuditSuspiciousActivity AuditEventType = "security.suspicious"
	AuditAccessDenied       AuditEventType = "security.access.denied"
	AuditRateLimitExceeded  AuditEventType = "security.rate_limit.exceeded"

	// Événements système
	AuditSystemStartup     AuditEventType = "system.startup"
	AuditSystemShutdown    AuditEventType = "system.shutdown"
	AuditSystemError       AuditEventType = "system.error"
	AuditSystemMaintenance AuditEventType = "system.maintenance"
)

// AuditSeverity niveau de sévérité
type AuditSeverity string

const (
	AuditSeverityInfo     AuditSeverity = "info"
	AuditSeverityWarning  AuditSeverity = "warning"
	AuditSeverityError    AuditSeverity = "error"
	AuditSeverityCritical AuditSeverity = "critical"
)

// AuditCategory catégorie d'audit
type AuditCategory string

const (
	AuditCategoryAuth     AuditCategory = "authentication"
	AuditCategoryData     AuditCategory = "data"
	AuditCategoryAdmin    AuditCategory = "administration"
	AuditCategorySecurity AuditCategory = "security"
	AuditCategorySystem   AuditCategory = "system"
)

// AuditMetadata métadonnées d'audit
type AuditMetadata struct {
	RequestID   string            `json:"request_id,omitempty"`
	TraceID     string            `json:"trace_id,omitempty"`
	Source      string            `json:"source"`
	Environment string            `json:"environment"`
	Application string            `json:"application"`
	Component   string            `json:"component"`
	Action      string            `json:"action"`
	Resource    string            `json:"resource,omitempty"`
	ResourceID  string            `json:"resource_id,omitempty"`
	Changes     []FieldChange     `json:"changes,omitempty"`
	Context     map[string]string `json:"context,omitempty"`
}

// FieldChange changement de champ
type FieldChange struct {
	Field    string      `json:"field"`
	OldValue interface{} `json:"old_value"`
	NewValue interface{} `json:"new_value"`
}

// EventStore interface pour le stockage des événements
type EventStore interface {
	SaveEvent(ctx context.Context, event *AuditEvent) error
	GetEvents(ctx context.Context, aggregateID string, fromVersion int64) ([]*AuditEvent, error)
	GetEventsByType(ctx context.Context, eventType AuditEventType, limit int) ([]*AuditEvent, error)
	GetEventsByTimeRange(ctx context.Context, start, end time.Time) ([]*AuditEvent, error)
	GetEventsByUser(ctx context.Context, userID int64, limit int) ([]*AuditEvent, error)
}

// Projection interface pour les projections
type Projection interface {
	ProcessEvent(ctx context.Context, event *AuditEvent) error
	GetProjectionName() string
	Reset() error
}

// EventSourcingMetrics métriques d'event sourcing
type EventSourcingMetrics struct {
	EventsProcessed    int64            `json:"events_processed"`
	EventsSaved        int64            `json:"events_saved"`
	EventsFailed       int64            `json:"events_failed"`
	ProjectionsUpdated int64            `json:"projections_updated"`
	AvgProcessingTime  time.Duration    `json:"avg_processing_time"`
	EventsByType       map[string]int64 `json:"events_by_type"`
	EventsBySeverity   map[string]int64 `json:"events_by_severity"`
	EventsByCategory   map[string]int64 `json:"events_by_category"`

	mutex sync.RWMutex
}

// NewEventSourcingService crée un nouveau service d'event sourcing
func NewEventSourcingService(natsService *NATSService, eventStore EventStore, config *EventSourcingConfig, logger *zap.Logger) (*EventSourcingService, error) {
	if config == nil {
		config = &EventSourcingConfig{
			BatchSize:          100,
			FlushInterval:      5 * time.Second,
			RetentionPeriod:    365 * 24 * time.Hour, // 1 an
			CompressionEnabled: true,
			EncryptionEnabled:  true,
			EnableProjections:  true,
			EnableSnapshots:    true,
			SnapshotInterval:   1000,
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	service := &EventSourcingService{
		natsService: natsService,
		logger:      logger,
		config:      config,
		eventStore:  eventStore,
		projections: make(map[string]Projection),
		metrics: &EventSourcingMetrics{
			EventsByType:     make(map[string]int64),
			EventsBySeverity: make(map[string]int64),
			EventsByCategory: make(map[string]int64),
		},
		ctx:    ctx,
		cancel: cancel,
	}

	// Démarrer les subscriptions NATS
	if err := service.startNATSSubscriptions(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to start NATS subscriptions: %w", err)
	}

	// Enregistrer les projections par défaut
	if config.EnableProjections {
		service.registerDefaultProjections()
	}

	// Démarrer les services de monitoring
	go service.startMetricsReporter()
	go service.startCleanupWorker()

	return service, nil
}

// ============================================================================
// CRÉATION D'ÉVÉNEMENTS D'AUDIT
// ============================================================================

// RecordAuditEvent enregistre un événement d'audit
func (e *EventSourcingService) RecordAuditEvent(ctx context.Context, event *AuditEvent) error {
	if event == nil {
		return fmt.Errorf("audit event cannot be nil")
	}

	// Enrichir l'événement
	if event.ID == "" {
		event.ID = e.generateEventID()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	if event.Metadata.Environment == "" {
		event.Metadata.Environment = "production"
	}
	if event.Metadata.Application == "" {
		event.Metadata.Application = "veza-backend"
	}

	// Valider l'événement
	if err := e.validateAuditEvent(event); err != nil {
		return fmt.Errorf("invalid audit event: %w", err)
	}

	// Sauvegarder dans l'event store
	if err := e.eventStore.SaveEvent(ctx, event); err != nil {
		e.recordEventFailed()
		return fmt.Errorf("failed to save audit event: %w", err)
	}

	// Publier l'événement via NATS pour processing en temps réel
	if err := e.publishAuditEvent(ctx, event); err != nil {
		e.logger.Warn("Failed to publish audit event", zap.Error(err))
		// Ne pas retourner d'erreur car l'événement est déjà sauvegardé
	}

	// Mettre à jour les métriques
	e.recordEventProcessed(event.EventType, event.Severity, event.Category)

	e.logger.Debug("Audit event recorded",
		zap.String("id", event.ID),
		zap.String("type", string(event.EventType)),
		zap.String("aggregate_id", event.AggregateID))

	return nil
}

// RecordUserLoginEvent enregistre un événement de connexion utilisateur
func (e *EventSourcingService) RecordUserLoginEvent(ctx context.Context, userID int64, sessionID string, ipAddress, userAgent string, success bool) error {
	eventType := AuditUserLogin
	severity := AuditSeverityInfo

	if !success {
		eventType = AuditUserLoginFailed
		severity = AuditSeverityWarning
	}

	event := &AuditEvent{
		AggregateID:   fmt.Sprintf("user-%d", userID),
		AggregateType: "user",
		EventType:     eventType,
		EventData: map[string]interface{}{
			"success":    success,
			"session_id": sessionID,
		},
		Metadata: AuditMetadata{
			Source:    "authentication_service",
			Component: "auth_handler",
			Action:    "login",
			Resource:  "user",
		},
		UserID:    &userID,
		SessionID: &sessionID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Severity:  severity,
		Category:  AuditCategoryAuth,
	}

	return e.RecordAuditEvent(ctx, event)
}

// RecordDataChangeEvent enregistre un événement de changement de données
func (e *EventSourcingService) RecordDataChangeEvent(ctx context.Context, userID int64, resource, resourceID, action string, changes []FieldChange) error {
	var eventType AuditEventType
	switch action {
	case "create":
		eventType = AuditDataCreated
	case "update":
		eventType = AuditDataUpdated
	case "delete":
		eventType = AuditDataDeleted
	default:
		eventType = AuditDataAccessed
	}

	event := &AuditEvent{
		AggregateID:   fmt.Sprintf("%s-%s", resource, resourceID),
		AggregateType: resource,
		EventType:     eventType,
		EventData: map[string]interface{}{
			"action":      action,
			"resource":    resource,
			"resource_id": resourceID,
			"changes":     changes,
		},
		Metadata: AuditMetadata{
			Source:     "data_service",
			Component:  "data_handler",
			Action:     action,
			Resource:   resource,
			ResourceID: resourceID,
			Changes:    changes,
		},
		UserID:   &userID,
		Severity: AuditSeverityInfo,
		Category: AuditCategoryData,
	}

	return e.RecordAuditEvent(ctx, event)
}

// RecordSecurityEvent enregistre un événement de sécurité
func (e *EventSourcingService) RecordSecurityEvent(ctx context.Context, eventType AuditEventType, userID *int64, ipAddress string, details map[string]interface{}) error {
	event := &AuditEvent{
		AggregateID:   fmt.Sprintf("security-%d", time.Now().Unix()),
		AggregateType: "security",
		EventType:     eventType,
		EventData:     details,
		Metadata: AuditMetadata{
			Source:    "security_service",
			Component: "security_handler",
			Action:    string(eventType),
			Resource:  "security",
		},
		UserID:    userID,
		IPAddress: ipAddress,
		Severity:  AuditSeverityCritical,
		Category:  AuditCategorySecurity,
	}

	return e.RecordAuditEvent(ctx, event)
}

// ============================================================================
// PROJECTIONS
// ============================================================================

// RegisterProjection enregistre une projection
func (e *EventSourcingService) RegisterProjection(projection Projection) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.projections[projection.GetProjectionName()] = projection

	e.logger.Info("Projection registered",
		zap.String("name", projection.GetProjectionName()))
}

// processProjections traite les projections pour un événement
func (e *EventSourcingService) processProjections(ctx context.Context, event *AuditEvent) {
	e.mutex.RLock()
	projections := make([]Projection, 0, len(e.projections))
	for _, projection := range e.projections {
		projections = append(projections, projection)
	}
	e.mutex.RUnlock()

	for _, projection := range projections {
		if err := projection.ProcessEvent(ctx, event); err != nil {
			e.logger.Error("Failed to process projection",
				zap.String("projection", projection.GetProjectionName()),
				zap.String("event_id", event.ID),
				zap.Error(err))
		} else {
			e.recordProjectionUpdated()
		}
	}
}

// ============================================================================
// PROJECTIONS PAR DÉFAUT
// ============================================================================

// UserActivityProjection projection d'activité utilisateur
type UserActivityProjection struct {
	logger *zap.Logger
}

func (p *UserActivityProjection) GetProjectionName() string {
	return "user_activity"
}

func (p *UserActivityProjection) ProcessEvent(ctx context.Context, event *AuditEvent) error {
	if event.UserID == nil {
		return nil // Pas d'activité utilisateur
	}

	// TODO: Implémenter la logique de projection
	// Par exemple, mettre à jour une table d'activité utilisateur
	p.logger.Debug("Processing user activity projection",
		zap.String("event_id", event.ID),
		zap.Int64("user_id", *event.UserID))

	return nil
}

func (p *UserActivityProjection) Reset() error {
	// TODO: Implémenter la réinitialisation
	return nil
}

// SecurityEventsProjection projection d'événements de sécurité
type SecurityEventsProjection struct {
	logger *zap.Logger
}

func (p *SecurityEventsProjection) GetProjectionName() string {
	return "security_events"
}

func (p *SecurityEventsProjection) ProcessEvent(ctx context.Context, event *AuditEvent) error {
	if event.Category != AuditCategorySecurity {
		return nil // Pas un événement de sécurité
	}

	// TODO: Implémenter la logique de projection sécurité
	// Par exemple, alerter sur des événements critiques
	if event.Severity == AuditSeverityCritical {
		p.logger.Warn("Critical security event detected",
			zap.String("event_id", event.ID),
			zap.String("event_type", string(event.EventType)))
	}

	return nil
}

func (p *SecurityEventsProjection) Reset() error {
	// TODO: Implémenter la réinitialisation
	return nil
}

// DataChangesProjection projection de changements de données
type DataChangesProjection struct {
	logger *zap.Logger
}

func (p *DataChangesProjection) GetProjectionName() string {
	return "data_changes"
}

func (p *DataChangesProjection) ProcessEvent(ctx context.Context, event *AuditEvent) error {
	if event.Category != AuditCategoryData {
		return nil // Pas un changement de données
	}

	// TODO: Implémenter la logique de projection données
	// Par exemple, maintenir un historique des changements
	p.logger.Debug("Processing data changes projection",
		zap.String("event_id", event.ID),
		zap.String("aggregate_id", event.AggregateID))

	return nil
}

func (p *DataChangesProjection) Reset() error {
	// TODO: Implémenter la réinitialisation
	return nil
}

// registerDefaultProjections enregistre les projections par défaut
func (e *EventSourcingService) registerDefaultProjections() {
	projections := []Projection{
		&UserActivityProjection{logger: e.logger},
		&SecurityEventsProjection{logger: e.logger},
		&DataChangesProjection{logger: e.logger},
	}

	for _, projection := range projections {
		e.RegisterProjection(projection)
	}
}

// ============================================================================
// NATS INTEGRATION
// ============================================================================

// startNATSSubscriptions démarre les subscriptions NATS
func (e *EventSourcingService) startNATSSubscriptions() error {
	// Subscription pour les événements d'audit
	if err := e.natsService.SubscribeToSubject("audit.events", e.handleAuditEvent); err != nil {
		return fmt.Errorf("failed to subscribe to audit events: %w", err)
	}

	return nil
}

// publishAuditEvent publie un événement d'audit via NATS
func (e *EventSourcingService) publishAuditEvent(ctx context.Context, event *AuditEvent) error {
	natsEvent := &Event{
		ID:        event.ID + "_audit",
		Type:      EventType("audit." + string(event.EventType)),
		Source:    "event_sourcing_service",
		Subject:   "audit.events",
		Data:      event,
		Priority:  e.convertSeverityToPriority(event.Severity),
		Timestamp: event.Timestamp,
		UserID:    event.UserID,
	}

	return e.natsService.PublishEvent(ctx, natsEvent)
}

// handleAuditEvent traite un événement d'audit reçu via NATS
func (e *EventSourcingService) handleAuditEvent(ctx context.Context, event *Event) error {
	data, err := json.Marshal(event.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal audit event data: %w", err)
	}

	var auditEvent AuditEvent
	if err := json.Unmarshal(data, &auditEvent); err != nil {
		return fmt.Errorf("failed to parse audit event: %w", err)
	}

	// Traiter les projections
	if e.config.EnableProjections {
		e.processProjections(ctx, &auditEvent)
	}

	return nil
}

// ============================================================================
// REQUÊTES
// ============================================================================

// GetEventsByAggregate récupère les événements pour un agrégat
func (e *EventSourcingService) GetEventsByAggregate(ctx context.Context, aggregateID string, fromVersion int64) ([]*AuditEvent, error) {
	return e.eventStore.GetEvents(ctx, aggregateID, fromVersion)
}

// GetEventsByType récupère les événements par type
func (e *EventSourcingService) GetEventsByType(ctx context.Context, eventType AuditEventType, limit int) ([]*AuditEvent, error) {
	return e.eventStore.GetEventsByType(ctx, eventType, limit)
}

// GetEventsByTimeRange récupère les événements dans une plage de temps
func (e *EventSourcingService) GetEventsByTimeRange(ctx context.Context, start, end time.Time) ([]*AuditEvent, error) {
	return e.eventStore.GetEventsByTimeRange(ctx, start, end)
}

// GetUserActivity récupère l'activité d'un utilisateur
func (e *EventSourcingService) GetUserActivity(ctx context.Context, userID int64, limit int) ([]*AuditEvent, error) {
	return e.eventStore.GetEventsByUser(ctx, userID, limit)
}

// ============================================================================
// UTILITAIRES
// ============================================================================

// validateAuditEvent valide un événement d'audit
func (e *EventSourcingService) validateAuditEvent(event *AuditEvent) error {
	if event.ID == "" {
		return fmt.Errorf("event ID is required")
	}

	if event.AggregateID == "" {
		return fmt.Errorf("aggregate ID is required")
	}

	if event.AggregateType == "" {
		return fmt.Errorf("aggregate type is required")
	}

	if event.EventType == "" {
		return fmt.Errorf("event type is required")
	}

	return nil
}

// generateEventID génère un ID unique pour l'événement
func (e *EventSourcingService) generateEventID() string {
	return fmt.Sprintf("audit_%d_%d", time.Now().UnixNano(), e.metrics.EventsProcessed)
}

// convertSeverityToPriority convertit la sévérité en priorité d'événement
func (e *EventSourcingService) convertSeverityToPriority(severity AuditSeverity) EventPriority {
	switch severity {
	case AuditSeverityInfo:
		return PriorityNormal
	case AuditSeverityWarning:
		return PriorityHigh
	case AuditSeverityError:
		return PriorityHigh
	case AuditSeverityCritical:
		return PriorityCritical
	default:
		return PriorityNormal
	}
}

// ============================================================================
// MÉTRIQUES
// ============================================================================

func (e *EventSourcingService) recordEventProcessed(eventType AuditEventType, severity AuditSeverity, category AuditCategory) {
	e.metrics.mutex.Lock()
	defer e.metrics.mutex.Unlock()

	e.metrics.EventsProcessed++
	e.metrics.EventsSaved++
	e.metrics.EventsByType[string(eventType)]++
	e.metrics.EventsBySeverity[string(severity)]++
	e.metrics.EventsByCategory[string(category)]++
}

func (e *EventSourcingService) recordEventFailed() {
	e.metrics.mutex.Lock()
	defer e.metrics.mutex.Unlock()

	e.metrics.EventsFailed++
}

func (e *EventSourcingService) recordProjectionUpdated() {
	e.metrics.mutex.Lock()
	defer e.metrics.mutex.Unlock()

	e.metrics.ProjectionsUpdated++
}

// GetMetrics retourne les métriques d'event sourcing
func (e *EventSourcingService) GetMetrics() *EventSourcingMetrics {
	e.metrics.mutex.RLock()
	defer e.metrics.mutex.RUnlock()

	// Copier les métriques
	eventsByType := make(map[string]int64)
	for k, v := range e.metrics.EventsByType {
		eventsByType[k] = v
	}

	eventsBySeverity := make(map[string]int64)
	for k, v := range e.metrics.EventsBySeverity {
		eventsBySeverity[k] = v
	}

	eventsByCategory := make(map[string]int64)
	for k, v := range e.metrics.EventsByCategory {
		eventsByCategory[k] = v
	}

	return &EventSourcingMetrics{
		EventsProcessed:    e.metrics.EventsProcessed,
		EventsSaved:        e.metrics.EventsSaved,
		EventsFailed:       e.metrics.EventsFailed,
		ProjectionsUpdated: e.metrics.ProjectionsUpdated,
		AvgProcessingTime:  e.metrics.AvgProcessingTime,
		EventsByType:       eventsByType,
		EventsBySeverity:   eventsBySeverity,
		EventsByCategory:   eventsByCategory,
	}
}

// ============================================================================
// MONITORING
// ============================================================================

// startMetricsReporter démarre le reporter de métriques
func (e *EventSourcingService) startMetricsReporter() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			metrics := e.GetMetrics()
			e.logger.Info("Event sourcing metrics",
				zap.Int64("events_processed", metrics.EventsProcessed),
				zap.Int64("events_saved", metrics.EventsSaved),
				zap.Int64("events_failed", metrics.EventsFailed),
				zap.Int64("projections_updated", metrics.ProjectionsUpdated))

		case <-e.ctx.Done():
			return
		}
	}
}

// startCleanupWorker démarre le worker de nettoyage
func (e *EventSourcingService) startCleanupWorker() {
	ticker := time.NewTicker(24 * time.Hour) // Nettoyage quotidien
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// TODO: Implémenter le nettoyage des anciens événements
			// selon la période de rétention configurée
			e.logger.Info("Cleanup worker running")

		case <-e.ctx.Done():
			return
		}
	}
}

// HealthCheck vérifie la santé du service
func (e *EventSourcingService) HealthCheck() error {
	// TODO: Vérifier la connectivité à l'event store
	return nil
}

// Close ferme proprement le service
func (e *EventSourcingService) Close() error {
	e.cancel()

	// Attendre que tous les workers terminent
	done := make(chan struct{})
	go func() {
		e.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		e.logger.Warn("Timeout waiting for event sourcing workers to finish")
	}

	e.logger.Info("Event sourcing service closed")
	return nil
}
