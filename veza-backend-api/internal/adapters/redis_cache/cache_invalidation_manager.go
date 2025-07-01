package redis_cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// CacheInvalidationManager gestionnaire d'invalidation intelligent
type CacheInvalidationManager struct {
	redis           *redis.Client
	multiLevelCache *MultiLevelCacheService
	rbacCache       *RBACCacheService
	queryCache      *QueryCacheService
	logger          *zap.Logger

	// Channels pour coordination
	invalidationChannel chan InvalidationEvent

	// Patterns d'invalidation
	invalidationPatterns map[string][]InvalidationRule

	// Métriques
	metrics *InvalidationMetrics

	// Configuration
	batchSize  int
	maxRetries int
	retryDelay time.Duration
}

// InvalidationEvent événement d'invalidation
type InvalidationEvent struct {
	Type       InvalidationType       `json:"type"`
	Resource   string                 `json:"resource"`
	ResourceID interface{}            `json:"resource_id,omitempty"`
	UserID     int64                  `json:"user_id,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
	Source     string                 `json:"source"`
	Priority   Priority               `json:"priority"`
}

// InvalidationType types d'invalidation
type InvalidationType string

const (
	InvalidationTypeUser       InvalidationType = "user"
	InvalidationTypeRole       InvalidationType = "role"
	InvalidationTypePermission InvalidationType = "permission"
	InvalidationTypeSession    InvalidationType = "session"
	InvalidationTypeQuery      InvalidationType = "query"
	InvalidationTypeTable      InvalidationType = "table"
	InvalidationTypePattern    InvalidationType = "pattern"
	InvalidationTypeGlobal     InvalidationType = "global"
)

// Priority priorité d'invalidation
type Priority int

const (
	PriorityLow Priority = iota
	PriorityNormal
	PriorityHigh
	PriorityCritical
)

// InvalidationRule règle d'invalidation
type InvalidationRule struct {
	CacheType    CacheType      `json:"cache_type"`
	Pattern      string         `json:"pattern"`
	Action       Action         `json:"action"`
	Dependencies []string       `json:"dependencies,omitempty"`
	TTLOverride  *time.Duration `json:"ttl_override,omitempty"`
}

// CacheType types de cache
type CacheType string

const (
	CacheTypeSession CacheType = "session"
	CacheTypeRBAC    CacheType = "rbac"
	CacheTypeQuery   CacheType = "query"
	CacheTypeAll     CacheType = "all"
)

// Action actions d'invalidation
type Action string

const (
	ActionDelete  Action = "delete"
	ActionExpire  Action = "expire"
	ActionRefresh Action = "refresh"
	ActionTag     Action = "tag"
)

// InvalidationMetrics métriques d'invalidation
type InvalidationMetrics struct {
	TotalInvalidations      int64            `json:"total_invalidations"`
	InvalidationsByType     map[string]int64 `json:"invalidations_by_type"`
	InvalidationsByPriority map[string]int64 `json:"invalidations_by_priority"`
	SuccessfulInvalidations int64            `json:"successful_invalidations"`
	FailedInvalidations     int64            `json:"failed_invalidations"`
	AvgProcessingTimeMs     int64            `json:"avg_processing_time_ms"`
	BatchedInvalidations    int64            `json:"batched_invalidations"`

	mutex sync.RWMutex
}

// NewCacheInvalidationManager crée un nouveau gestionnaire d'invalidation
func NewCacheInvalidationManager(
	redisClient *redis.Client,
	multiLevelCache *MultiLevelCacheService,
	rbacCache *RBACCacheService,
	queryCache *QueryCacheService,
	logger *zap.Logger,
) *CacheInvalidationManager {
	manager := &CacheInvalidationManager{
		redis:                redisClient,
		multiLevelCache:      multiLevelCache,
		rbacCache:            rbacCache,
		queryCache:           queryCache,
		logger:               logger,
		invalidationChannel:  make(chan InvalidationEvent, 1000),
		invalidationPatterns: make(map[string][]InvalidationRule),
		metrics: &InvalidationMetrics{
			InvalidationsByType:     make(map[string]int64),
			InvalidationsByPriority: make(map[string]int64),
		},
		batchSize:  100,
		maxRetries: 3,
		retryDelay: 100 * time.Millisecond,
	}

	// Initialiser les patterns d'invalidation
	manager.initializeInvalidationPatterns()

	// Démarrer le processeur d'événements
	go manager.startEventProcessor()

	// Démarrer le processeur de batches
	go manager.startBatchProcessor()

	// Démarrer le monitoring
	go manager.startMetricsReporter()

	return manager
}

// ============================================================================
// INVALIDATION INTELLIGENTE
// ============================================================================

// InvalidateUser invalide tous les caches liés à un utilisateur
func (c *CacheInvalidationManager) InvalidateUser(ctx context.Context, userID int64) error {
	event := InvalidationEvent{
		Type:       InvalidationTypeUser,
		Resource:   "user",
		ResourceID: userID,
		UserID:     userID,
		Timestamp:  time.Now(),
		Source:     "api",
		Priority:   PriorityHigh,
	}

	return c.processInvalidationEvent(ctx, event)
}

// InvalidateUserSession invalide la session d'un utilisateur
func (c *CacheInvalidationManager) InvalidateUserSession(ctx context.Context, userID int64) error {
	event := InvalidationEvent{
		Type:       InvalidationTypeSession,
		Resource:   "session",
		ResourceID: userID,
		UserID:     userID,
		Timestamp:  time.Now(),
		Source:     "auth",
		Priority:   PriorityCritical,
	}

	return c.processInvalidationEvent(ctx, event)
}

// InvalidatePermissions invalide les permissions d'un rôle
func (c *CacheInvalidationManager) InvalidatePermissions(ctx context.Context, role string) error {
	event := InvalidationEvent{
		Type:       InvalidationTypePermission,
		Resource:   "permission",
		ResourceID: role,
		Timestamp:  time.Now(),
		Source:     "rbac",
		Priority:   PriorityHigh,
		Metadata:   map[string]interface{}{"role": role},
	}

	return c.processInvalidationEvent(ctx, event)
}

// InvalidateTable invalide tous les caches liés à une table
func (c *CacheInvalidationManager) InvalidateTable(ctx context.Context, tableName string) error {
	event := InvalidationEvent{
		Type:       InvalidationTypeTable,
		Resource:   "table",
		ResourceID: tableName,
		Timestamp:  time.Now(),
		Source:     "database",
		Priority:   PriorityNormal,
		Metadata:   map[string]interface{}{"table": tableName},
	}

	return c.processInvalidationEvent(ctx, event)
}

// InvalidateByPattern invalide par pattern personnalisé
func (c *CacheInvalidationManager) InvalidateByPattern(ctx context.Context, pattern string, priority Priority) error {
	event := InvalidationEvent{
		Type:       InvalidationTypePattern,
		Resource:   "pattern",
		ResourceID: pattern,
		Timestamp:  time.Now(),
		Source:     "custom",
		Priority:   priority,
		Metadata:   map[string]interface{}{"pattern": pattern},
	}

	return c.processInvalidationEvent(ctx, event)
}

// processInvalidationEvent traite un événement d'invalidation
func (c *CacheInvalidationManager) processInvalidationEvent(ctx context.Context, event InvalidationEvent) error {
	start := time.Now()

	c.recordInvalidationAttempt(event)

	// Récupérer les règles d'invalidation
	rules := c.getInvalidationRules(event)

	var errors []error
	var wg sync.WaitGroup

	// Exécuter les invalidations en parallèle
	for _, rule := range rules {
		wg.Add(1)
		go func(rule InvalidationRule) {
			defer wg.Done()
			if err := c.executeInvalidationRule(ctx, event, rule); err != nil {
				errors = append(errors, err)
			}
		}(rule)
	}

	wg.Wait()

	// Enregistrer les métriques
	processingTime := time.Since(start)
	if len(errors) == 0 {
		c.recordSuccessfulInvalidation(processingTime)
		c.logger.Debug("Invalidation réussie",
			zap.String("type", string(event.Type)),
			zap.String("resource", event.Resource),
			zap.Any("resource_id", event.ResourceID),
			zap.Duration("processing_time", processingTime),
			zap.Int("rules_executed", len(rules)))
	} else {
		c.recordFailedInvalidation()
		c.logger.Warn("Invalidation partiellement échouée",
			zap.String("type", string(event.Type)),
			zap.String("resource", event.Resource),
			zap.Int("errors", len(errors)),
			zap.Errors("invalidation_errors", errors))
	}

	// Publier l'événement pour les autres services
	go c.publishInvalidationEvent(event)

	if len(errors) > 0 {
		return fmt.Errorf("invalidation partiellement échouée: %d erreurs", len(errors))
	}

	return nil
}

// executeInvalidationRule exécute une règle d'invalidation spécifique
func (c *CacheInvalidationManager) executeInvalidationRule(ctx context.Context, event InvalidationEvent, rule InvalidationRule) error {
	switch rule.CacheType {
	case CacheTypeSession:
		return c.invalidateSessionCache(ctx, event, rule)
	case CacheTypeRBAC:
		return c.invalidateRBACCache(ctx, event, rule)
	case CacheTypeQuery:
		return c.invalidateQueryCache(ctx, event, rule)
	case CacheTypeAll:
		// Invalider tous les types de cache
		errors := []error{}
		if err := c.invalidateSessionCache(ctx, event, rule); err != nil {
			errors = append(errors, err)
		}
		if err := c.invalidateRBACCache(ctx, event, rule); err != nil {
			errors = append(errors, err)
		}
		if err := c.invalidateQueryCache(ctx, event, rule); err != nil {
			errors = append(errors, err)
		}
		if len(errors) > 0 {
			return fmt.Errorf("erreurs invalidation multiple: %v", errors)
		}
		return nil
	default:
		return fmt.Errorf("type de cache non supporté: %s", rule.CacheType)
	}
}

// invalidateSessionCache invalide le cache de session
func (c *CacheInvalidationManager) invalidateSessionCache(ctx context.Context, event InvalidationEvent, rule InvalidationRule) error {
	if c.multiLevelCache == nil {
		return nil
	}

	switch event.Type {
	case InvalidationTypeUser, InvalidationTypeSession:
		if userID, ok := event.ResourceID.(int64); ok {
			return c.multiLevelCache.InvalidateUserSession(ctx, userID)
		}
		if userIDFloat, ok := event.ResourceID.(float64); ok {
			return c.multiLevelCache.InvalidateUserSession(ctx, int64(userIDFloat))
		}
	case InvalidationTypePattern:
		// Pour les patterns, nous devons implémenter une invalidation par pattern
		// Pour l'instant, on peut laisser vide ou implémenter selon les besoins
		return nil
	}

	return nil
}

// invalidateRBACCache invalide le cache RBAC
func (c *CacheInvalidationManager) invalidateRBACCache(ctx context.Context, event InvalidationEvent, rule InvalidationRule) error {
	if c.rbacCache == nil {
		return nil
	}

	switch event.Type {
	case InvalidationTypeUser:
		if userID, ok := event.ResourceID.(int64); ok {
			return c.rbacCache.InvalidateUserPermissions(ctx, userID)
		}
		if userIDFloat, ok := event.ResourceID.(float64); ok {
			return c.rbacCache.InvalidateUserPermissions(ctx, int64(userIDFloat))
		}
	case InvalidationTypeRole, InvalidationTypePermission:
		if role, ok := event.ResourceID.(string); ok {
			return c.rbacCache.InvalidateRolePermissions(ctx, role)
		}
	}

	return nil
}

// invalidateQueryCache invalide le cache de requêtes
func (c *CacheInvalidationManager) invalidateQueryCache(ctx context.Context, event InvalidationEvent, rule InvalidationRule) error {
	if c.queryCache == nil {
		return nil
	}

	switch event.Type {
	case InvalidationTypeTable:
		if tableName, ok := event.ResourceID.(string); ok {
			return c.queryCache.InvalidateByTable(ctx, tableName)
		}
	case InvalidationTypePattern:
		if pattern, ok := event.ResourceID.(string); ok {
			return c.queryCache.InvalidateByPattern(ctx, pattern)
		}
	}

	return nil
}

// ============================================================================
// RÈGLES D'INVALIDATION
// ============================================================================

// initializeInvalidationPatterns initialise les patterns d'invalidation
func (c *CacheInvalidationManager) initializeInvalidationPatterns() {
	// Patterns pour l'invalidation utilisateur
	c.invalidationPatterns["user"] = []InvalidationRule{
		{CacheType: CacheTypeSession, Pattern: "user_session:*", Action: ActionDelete},
		{CacheType: CacheTypeRBAC, Pattern: "perm:*", Action: ActionDelete},
		{CacheType: CacheTypeQuery, Pattern: "query:*user*", Action: ActionDelete},
	}

	// Patterns pour l'invalidation de session
	c.invalidationPatterns["session"] = []InvalidationRule{
		{CacheType: CacheTypeSession, Pattern: "user_session:*", Action: ActionDelete},
	}

	// Patterns pour l'invalidation de permissions
	c.invalidationPatterns["permission"] = []InvalidationRule{
		{CacheType: CacheTypeRBAC, Pattern: "role_perm:*", Action: ActionDelete},
		{CacheType: CacheTypeRBAC, Pattern: "perm:*", Action: ActionDelete},
	}

	// Patterns pour l'invalidation de table
	c.invalidationPatterns["table"] = []InvalidationRule{
		{CacheType: CacheTypeQuery, Pattern: "query:*", Action: ActionDelete},
		{CacheType: CacheTypeSession, Pattern: "user_session:*", Action: ActionExpire, TTLOverride: &[]time.Duration{5 * time.Minute}[0]},
	}

	// Patterns pour l'invalidation globale
	c.invalidationPatterns["global"] = []InvalidationRule{
		{CacheType: CacheTypeAll, Pattern: "*", Action: ActionDelete},
	}
}

// getInvalidationRules récupère les règles d'invalidation pour un événement
func (c *CacheInvalidationManager) getInvalidationRules(event InvalidationEvent) []InvalidationRule {
	rules := []InvalidationRule{}

	// Récupérer les règles basées sur le type d'événement
	if eventRules, exists := c.invalidationPatterns[string(event.Type)]; exists {
		rules = append(rules, eventRules...)
	}

	// Ajouter des règles spécifiques selon la priorité
	if event.Priority == PriorityCritical {
		// Pour les événements critiques, ajouter des règles supplémentaires
		rules = append(rules, InvalidationRule{
			CacheType:   CacheTypeAll,
			Pattern:     "*",
			Action:      ActionExpire,
			TTLOverride: &[]time.Duration{1 * time.Minute}[0],
		})
	}

	return rules
}

// ============================================================================
// TRAITEMENT EN BATCH
// ============================================================================

// startBatchProcessor démarre le processeur de batches
func (c *CacheInvalidationManager) startBatchProcessor() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	batch := []InvalidationEvent{}

	for {
		select {
		case event := <-c.invalidationChannel:
			batch = append(batch, event)

			// Traiter le batch si plein ou événement critique
			if len(batch) >= c.batchSize || event.Priority == PriorityCritical {
				c.processBatch(batch)
				batch = []InvalidationEvent{}
			}

		case <-ticker.C:
			// Traiter le batch périodiquement
			if len(batch) > 0 {
				c.processBatch(batch)
				batch = []InvalidationEvent{}
			}
		}
	}
}

// processBatch traite un batch d'événements d'invalidation
func (c *CacheInvalidationManager) processBatch(events []InvalidationEvent) {
	if len(events) == 0 {
		return
	}

	start := time.Now()
	ctx := context.Background()

	// Grouper les événements par type pour optimisation
	eventGroups := make(map[InvalidationType][]InvalidationEvent)
	for _, event := range events {
		eventGroups[event.Type] = append(eventGroups[event.Type], event)
	}

	var wg sync.WaitGroup

	// Traiter chaque groupe en parallèle
	for eventType, groupEvents := range eventGroups {
		wg.Add(1)
		go func(eventType InvalidationType, events []InvalidationEvent) {
			defer wg.Done()
			for _, event := range events {
				if err := c.processInvalidationEvent(ctx, event); err != nil {
					c.logger.Warn("Erreur traitement événement batch",
						zap.String("type", string(eventType)),
						zap.Error(err))
				}
			}
		}(eventType, groupEvents)
	}

	wg.Wait()

	c.recordBatchProcessing(len(events), time.Since(start))

	c.logger.Debug("Batch d'invalidation traité",
		zap.Int("events_count", len(events)),
		zap.Int("groups_count", len(eventGroups)),
		zap.Duration("processing_time", time.Since(start)))
}

// startEventProcessor démarre le processeur d'événements
func (c *CacheInvalidationManager) startEventProcessor() {
	// Le processeur principal est le batch processor
	// Ici on peut ajouter de la logique supplémentaire si nécessaire
}

// ============================================================================
// PUBLICATION D'ÉVÉNEMENTS
// ============================================================================

// publishInvalidationEvent publie un événement d'invalidation
func (c *CacheInvalidationManager) publishInvalidationEvent(event InvalidationEvent) {
	eventData, err := json.Marshal(event)
	if err != nil {
		c.logger.Warn("Erreur sérialisation événement invalidation", zap.Error(err))
		return
	}

	// Publier sur Redis pub/sub pour les autres instances
	channel := fmt.Sprintf("cache_invalidation:%s", event.Type)
	if err := c.redis.Publish(context.Background(), channel, eventData).Err(); err != nil {
		c.logger.Warn("Erreur publication événement invalidation", zap.Error(err))
	}
}

// ============================================================================
// MÉTRIQUES ET MONITORING
// ============================================================================

func (c *CacheInvalidationManager) recordInvalidationAttempt(event InvalidationEvent) {
	c.metrics.mutex.Lock()
	defer c.metrics.mutex.Unlock()

	c.metrics.TotalInvalidations++
	c.metrics.InvalidationsByType[string(event.Type)]++
	c.metrics.InvalidationsByPriority[c.priorityToString(event.Priority)]++
}

func (c *CacheInvalidationManager) recordSuccessfulInvalidation(processingTime time.Duration) {
	c.metrics.mutex.Lock()
	defer c.metrics.mutex.Unlock()

	c.metrics.SuccessfulInvalidations++
	c.metrics.AvgProcessingTimeMs = (c.metrics.AvgProcessingTimeMs + processingTime.Milliseconds()) / 2
}

func (c *CacheInvalidationManager) recordFailedInvalidation() {
	c.metrics.mutex.Lock()
	defer c.metrics.mutex.Unlock()

	c.metrics.FailedInvalidations++
}

func (c *CacheInvalidationManager) recordBatchProcessing(eventCount int, processingTime time.Duration) {
	c.metrics.mutex.Lock()
	defer c.metrics.mutex.Unlock()

	c.metrics.BatchedInvalidations += int64(eventCount)
}

func (c *CacheInvalidationManager) priorityToString(priority Priority) string {
	switch priority {
	case PriorityLow:
		return "low"
	case PriorityNormal:
		return "normal"
	case PriorityHigh:
		return "high"
	case PriorityCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// GetInvalidationMetrics retourne les métriques d'invalidation
func (c *CacheInvalidationManager) GetInvalidationMetrics() *InvalidationMetrics {
	c.metrics.mutex.RLock()
	defer c.metrics.mutex.RUnlock()

	// Copier les maps
	typesCopy := make(map[string]int64)
	for k, v := range c.metrics.InvalidationsByType {
		typesCopy[k] = v
	}

	priorityCopy := make(map[string]int64)
	for k, v := range c.metrics.InvalidationsByPriority {
		priorityCopy[k] = v
	}

	return &InvalidationMetrics{
		TotalInvalidations:      c.metrics.TotalInvalidations,
		InvalidationsByType:     typesCopy,
		InvalidationsByPriority: priorityCopy,
		SuccessfulInvalidations: c.metrics.SuccessfulInvalidations,
		FailedInvalidations:     c.metrics.FailedInvalidations,
		AvgProcessingTimeMs:     c.metrics.AvgProcessingTimeMs,
		BatchedInvalidations:    c.metrics.BatchedInvalidations,
	}
}

func (c *CacheInvalidationManager) startMetricsReporter() {
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		metrics := c.GetInvalidationMetrics()

		successRate := float64(0)
		if metrics.TotalInvalidations > 0 {
			successRate = float64(metrics.SuccessfulInvalidations) / float64(metrics.TotalInvalidations)
		}

		c.logger.Info("Métriques invalidation de cache",
			zap.Int64("total_invalidations", metrics.TotalInvalidations),
			zap.Float64("success_rate", successRate),
			zap.Int64("avg_processing_time_ms", metrics.AvgProcessingTimeMs),
			zap.Int64("batched_invalidations", metrics.BatchedInvalidations),
			zap.Any("by_type", metrics.InvalidationsByType))
	}
}

// HealthCheck vérifie la santé du gestionnaire d'invalidation
func (c *CacheInvalidationManager) HealthCheck(ctx context.Context) error {
	// Test d'invalidation simple
	testEvent := InvalidationEvent{
		Type:       InvalidationTypePattern,
		Resource:   "test",
		ResourceID: "health_check",
		Timestamp:  time.Now(),
		Source:     "health_check",
		Priority:   PriorityLow,
	}

	if err := c.processInvalidationEvent(ctx, testEvent); err != nil {
		return fmt.Errorf("échec test invalidation: %w", err)
	}

	return nil
}
