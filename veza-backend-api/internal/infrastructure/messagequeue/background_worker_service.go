package messagequeue

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"sync"
	"time"

	"go.uber.org/zap"
)

// BackgroundWorkerService service pour les tâches en arrière-plan
type BackgroundWorkerService struct {
	natsService *NATSService
	logger      *zap.Logger
	config      *WorkerConfig

	// Pools de workers
	workerPools map[TaskType]*WorkerPool

	// Métriques
	metrics *WorkerMetrics

	// Contrôle de lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	mutex  sync.RWMutex
}

// WorkerConfig configuration des workers
type WorkerConfig struct {
	// Configuration générale
	DefaultWorkerCount int           `json:"default_worker_count"`
	TaskTimeout        time.Duration `json:"task_timeout"`
	MaxRetries         int           `json:"max_retries"`
	RetryDelay         time.Duration `json:"retry_delay"`

	// Configuration par type de tâche
	WorkerCounts map[string]int           `json:"worker_counts"`
	TaskTimeouts map[string]time.Duration `json:"task_timeouts"`

	// Configuration de monitoring
	HealthCheckInterval time.Duration `json:"health_check_interval"`
	MetricsInterval     time.Duration `json:"metrics_interval"`

	// Configuration de performance
	MaxConcurrentTasks int `json:"max_concurrent_tasks"`
	QueueCapacity      int `json:"queue_capacity"`
}

// Task représente une tâche à exécuter
type Task struct {
	ID          string                 `json:"id"`
	Type        TaskType               `json:"type"`
	Priority    TaskPriority           `json:"priority"`
	Data        map[string]interface{} `json:"data"`
	CreatedAt   time.Time              `json:"created_at"`
	ScheduledAt *time.Time             `json:"scheduled_at,omitempty"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	RetryCount  int                    `json:"retry_count"`
	MaxRetries  int                    `json:"max_retries"`
	UserID      *int64                 `json:"user_id,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Status      TaskStatus             `json:"status"`
}

// TaskType types de tâches
type TaskType string

const (
	// Tâches de fichiers
	TaskImageResize       TaskType = "image.resize"
	TaskImageOptimize     TaskType = "image.optimize"
	TaskVideoTranscode    TaskType = "video.transcode"
	TaskAudioProcess      TaskType = "audio.process"
	TaskFileAnalysis      TaskType = "file.analysis"
	TaskThumbnailGenerate TaskType = "thumbnail.generate"

	// Tâches de données
	TaskDataExport      TaskType = "data.export"
	TaskDataImport      TaskType = "data.import"
	TaskDataAnalysis    TaskType = "data.analysis"
	TaskReportGenerate  TaskType = "report.generate"
	TaskBackupCreate    TaskType = "backup.create"
	TaskDatabaseCleanup TaskType = "database.cleanup"

	// Tâches de communication
	TaskEmailBatch        TaskType = "email.batch"
	TaskNotificationBatch TaskType = "notification.batch"
	TaskSMSBatch          TaskType = "sms.batch"

	// Tâches de calcul
	TaskAnalyticsCompute  TaskType = "analytics.compute"
	TaskMLPrediction      TaskType = "ml.prediction"
	TaskSearchIndexUpdate TaskType = "search.index.update"
	TaskCacheWarmup       TaskType = "cache.warmup"

	// Tâches système
	TaskSystemMaintenance  TaskType = "system.maintenance"
	TaskLogRotation        TaskType = "log.rotation"
	TaskMetricsAggregation TaskType = "metrics.aggregation"
)

// TaskPriority priorité des tâches
type TaskPriority string

const (
	TaskPriorityLow      TaskPriority = "low"
	TaskPriorityNormal   TaskPriority = "normal"
	TaskPriorityHigh     TaskPriority = "high"
	TaskPriorityCritical TaskPriority = "critical"
)

// TaskStatus statut des tâches
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusQueued    TaskStatus = "queued"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusRetrying  TaskStatus = "retrying"
	TaskStatusCancelled TaskStatus = "cancelled"
)

// WorkerPool pool de workers pour un type de tâche
type WorkerPool struct {
	taskType    TaskType
	workerCount int
	workers     []*Worker
	taskQueue   chan *Task
	resultQueue chan *TaskResult
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	logger      *zap.Logger
}

// Worker worker individuel
type Worker struct {
	id          int
	taskType    TaskType
	processor   TaskProcessor
	taskQueue   chan *Task
	resultQueue chan *TaskResult
	ctx         context.Context
	logger      *zap.Logger
	metrics     *WorkerMetrics
}

// TaskResult résultat d'une tâche
type TaskResult struct {
	TaskID      string                 `json:"task_id"`
	Status      TaskStatus             `json:"status"`
	Result      map[string]interface{} `json:"result,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Duration    time.Duration          `json:"duration"`
	CompletedAt time.Time              `json:"completed_at"`
}

// TaskProcessor interface pour traiter les tâches
type TaskProcessor interface {
	ProcessTask(ctx context.Context, task *Task) (*TaskResult, error)
	GetTaskType() TaskType
}

// WorkerMetrics métriques des workers
type WorkerMetrics struct {
	TasksTotal        int64            `json:"tasks_total"`
	TasksCompleted    int64            `json:"tasks_completed"`
	TasksFailed       int64            `json:"tasks_failed"`
	TasksRetried      int64            `json:"tasks_retried"`
	TasksCancelled    int64            `json:"tasks_cancelled"`
	ActiveTasks       int64            `json:"active_tasks"`
	QueuedTasks       int64            `json:"queued_tasks"`
	AvgProcessingTime time.Duration    `json:"avg_processing_time"`
	TasksByType       map[string]int64 `json:"tasks_by_type"`
	TasksByPriority   map[string]int64 `json:"tasks_by_priority"`
	WorkersByType     map[string]int   `json:"workers_by_type"`

	mutex sync.RWMutex
}

// Sujets NATS pour les tâches
const (
	TaskSubjectScheduled = "tasks.scheduled"
	TaskSubjectImmediate = "tasks.immediate"
	TaskSubjectResults   = "tasks.results"
	TaskSubjectRetry     = "tasks.retry"
	TaskSubjectDLQ       = "tasks.dlq"
	TaskSubjectHeartbeat = "tasks.heartbeat"
)

// NewBackgroundWorkerService crée un nouveau service de workers
func NewBackgroundWorkerService(natsService *NATSService, config *WorkerConfig, logger *zap.Logger) (*BackgroundWorkerService, error) {
	if config == nil {
		config = &WorkerConfig{
			DefaultWorkerCount:  runtime.NumCPU(),
			TaskTimeout:         5 * time.Minute,
			MaxRetries:          3,
			RetryDelay:          1 * time.Minute,
			WorkerCounts:        make(map[string]int),
			TaskTimeouts:        make(map[string]time.Duration),
			HealthCheckInterval: 30 * time.Second,
			MetricsInterval:     1 * time.Minute,
			MaxConcurrentTasks:  100,
			QueueCapacity:       1000,
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	service := &BackgroundWorkerService{
		natsService: natsService,
		logger:      logger,
		config:      config,
		workerPools: make(map[TaskType]*WorkerPool),
		metrics: &WorkerMetrics{
			TasksByType:     make(map[string]int64),
			TasksByPriority: make(map[string]int64),
			WorkersByType:   make(map[string]int),
		},
		ctx:    ctx,
		cancel: cancel,
	}

	// Enregistrer les processors par défaut
	if err := service.registerDefaultProcessors(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to register default processors: %w", err)
	}

	// Démarrer les subscriptions NATS
	if err := service.startNATSSubscriptions(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to start NATS subscriptions: %w", err)
	}

	// Démarrer les services de monitoring
	go service.startHealthChecker()
	go service.startMetricsReporter()

	return service, nil
}

// ============================================================================
// SOUMISSION DE TÂCHES
// ============================================================================

// SubmitTask soumet une tâche pour exécution
func (b *BackgroundWorkerService) SubmitTask(ctx context.Context, task *Task) error {
	if task == nil {
		return fmt.Errorf("task cannot be nil")
	}

	// Enrichir la tâche
	if task.ID == "" {
		task.ID = b.generateTaskID()
	}
	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now()
	}
	if task.Status == "" {
		task.Status = TaskStatusQueued
	}
	if task.MaxRetries == 0 {
		task.MaxRetries = b.config.MaxRetries
	}

	// Valider la tâche
	if err := b.validateTask(task); err != nil {
		return fmt.Errorf("invalid task: %w", err)
	}

	// Déterminer le sujet selon la priorité et le scheduling
	subject := b.getSubjectForTask(task)

	// Créer l'événement
	event := &Event{
		ID:        task.ID,
		Type:      EventType(fmt.Sprintf("task.%s.submitted", task.Type)),
		Source:    "background_worker_service",
		Subject:   subject,
		Data:      task,
		Priority:  b.convertTaskPriority(task.Priority),
		Timestamp: time.Now(),
		UserID:    task.UserID,
	}

	// Publier l'événement
	if err := b.natsService.PublishEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to submit task: %w", err)
	}

	// Enregistrer les métriques
	b.recordTaskSubmitted(task.Type, task.Priority)

	b.logger.Debug("Task submitted",
		zap.String("id", task.ID),
		zap.String("type", string(task.Type)),
		zap.String("priority", string(task.Priority)))

	return nil
}

// SubmitTaskScheduled soumet une tâche programmée
func (b *BackgroundWorkerService) SubmitTaskScheduled(ctx context.Context, task *Task, scheduledAt time.Time) error {
	task.ScheduledAt = &scheduledAt
	return b.SubmitTask(ctx, task)
}

// SubmitBatchTasks soumet un lot de tâches
func (b *BackgroundWorkerService) SubmitBatchTasks(ctx context.Context, tasks []*Task) error {
	for _, task := range tasks {
		if err := b.SubmitTask(ctx, task); err != nil {
			return fmt.Errorf("failed to submit batch task %s: %w", task.ID, err)
		}
	}

	b.logger.Info("Batch tasks submitted", zap.Int("count", len(tasks)))
	return nil
}

// ============================================================================
// WORKER POOLS
// ============================================================================

// RegisterTaskProcessor enregistre un processor pour un type de tâche
func (b *BackgroundWorkerService) RegisterTaskProcessor(processor TaskProcessor) error {
	taskType := processor.GetTaskType()

	// Obtenir le nombre de workers pour ce type
	workerCount := b.config.DefaultWorkerCount
	if count, exists := b.config.WorkerCounts[string(taskType)]; exists {
		workerCount = count
	}

	// Créer le pool de workers
	pool, err := b.createWorkerPool(taskType, workerCount, processor)
	if err != nil {
		return fmt.Errorf("failed to create worker pool for %s: %w", taskType, err)
	}

	b.mutex.Lock()
	b.workerPools[taskType] = pool
	b.metrics.WorkersByType[string(taskType)] = workerCount
	b.mutex.Unlock()

	b.logger.Info("Task processor registered",
		zap.String("task_type", string(taskType)),
		zap.Int("worker_count", workerCount))

	return nil
}

// createWorkerPool crée un pool de workers
func (b *BackgroundWorkerService) createWorkerPool(taskType TaskType, workerCount int, processor TaskProcessor) (*WorkerPool, error) {
	ctx, cancel := context.WithCancel(b.ctx)

	pool := &WorkerPool{
		taskType:    taskType,
		workerCount: workerCount,
		workers:     make([]*Worker, workerCount),
		taskQueue:   make(chan *Task, b.config.QueueCapacity),
		resultQueue: make(chan *TaskResult, b.config.QueueCapacity),
		ctx:         ctx,
		cancel:      cancel,
		logger:      b.logger,
	}

	// Créer les workers
	for i := 0; i < workerCount; i++ {
		worker := &Worker{
			id:          i,
			taskType:    taskType,
			processor:   processor,
			taskQueue:   pool.taskQueue,
			resultQueue: pool.resultQueue,
			ctx:         ctx,
			logger:      b.logger,
			metrics:     b.metrics,
		}

		pool.workers[i] = worker

		// Démarrer le worker
		b.wg.Add(1)
		go worker.run(&b.wg)
	}

	// Démarrer le gestionnaire de résultats
	b.wg.Add(1)
	go pool.handleResults(&b.wg, b.natsService)

	return pool, nil
}

// ============================================================================
// WORKER EXECUTION
// ============================================================================

// run exécute le worker
func (w *Worker) run(wg *sync.WaitGroup) {
	defer wg.Done()

	w.logger.Debug("Worker started",
		zap.Int("worker_id", w.id),
		zap.String("task_type", string(w.taskType)))

	for {
		select {
		case task := <-w.taskQueue:
			w.processTask(task)

		case <-w.ctx.Done():
			w.logger.Debug("Worker stopping",
				zap.Int("worker_id", w.id),
				zap.String("task_type", string(w.taskType)))
			return
		}
	}
}

// processTask traite une tâche
func (w *Worker) processTask(task *Task) {
	start := time.Now()

	// Marquer la tâche comme en cours
	task.Status = TaskStatusRunning
	task.StartedAt = &start

	w.logger.Debug("Processing task",
		zap.String("task_id", task.ID),
		zap.String("task_type", string(task.Type)),
		zap.Int("worker_id", w.id))

	// Créer un contexte avec timeout
	timeout := 5 * time.Minute // Valeur par défaut
	ctx, cancel := context.WithTimeout(w.ctx, timeout)
	defer cancel()

	// Traiter la tâche
	result, err := w.processor.ProcessTask(ctx, task)
	duration := time.Since(start)

	// Créer le résultat
	if result == nil {
		result = &TaskResult{
			TaskID:      task.ID,
			Duration:    duration,
			CompletedAt: time.Now(),
		}
	}

	if err != nil {
		result.Status = TaskStatusFailed
		result.Error = err.Error()
		task.Error = err.Error()

		w.logger.Error("Task failed",
			zap.String("task_id", task.ID),
			zap.Error(err),
			zap.Duration("duration", duration))

		w.metrics.recordTaskFailed(task.Type, duration)
	} else {
		result.Status = TaskStatusCompleted
		task.Status = TaskStatusCompleted
		completedAt := time.Now()
		task.CompletedAt = &completedAt

		w.logger.Debug("Task completed",
			zap.String("task_id", task.ID),
			zap.Duration("duration", duration))

		w.metrics.recordTaskCompleted(task.Type, duration)
	}

	// Envoyer le résultat
	select {
	case w.resultQueue <- result:
	case <-w.ctx.Done():
		return
	}
}

// handleResults traite les résultats des tâches
func (p *WorkerPool) handleResults(wg *sync.WaitGroup, natsService *NATSService) {
	defer wg.Done()

	for {
		select {
		case result := <-p.resultQueue:
			// Publier le résultat via NATS
			event := &Event{
				ID:        result.TaskID + "_result",
				Type:      EventType("task.result"),
				Source:    "background_worker_service",
				Subject:   TaskSubjectResults,
				Data:      result,
				Priority:  PriorityNormal,
				Timestamp: time.Now(),
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			if err := natsService.PublishEvent(ctx, event); err != nil {
				p.logger.Error("Failed to publish task result",
					zap.String("task_id", result.TaskID),
					zap.Error(err))
			}
			cancel()

		case <-p.ctx.Done():
			return
		}
	}
}

// ============================================================================
// SUBSCRIPTION NATS
// ============================================================================

// startNATSSubscriptions démarre les subscriptions NATS
func (b *BackgroundWorkerService) startNATSSubscriptions() error {
	// Subscription pour tâches immédiates
	if err := b.natsService.SubscribeToSubject(TaskSubjectImmediate, b.handleImmediateTask); err != nil {
		return fmt.Errorf("failed to subscribe to immediate tasks: %w", err)
	}

	// Subscription pour tâches programmées
	if err := b.natsService.SubscribeToSubject(TaskSubjectScheduled, b.handleScheduledTask); err != nil {
		return fmt.Errorf("failed to subscribe to scheduled tasks: %w", err)
	}

	// Subscription pour retry
	if err := b.natsService.SubscribeToSubject(TaskSubjectRetry, b.handleRetryTask); err != nil {
		return fmt.Errorf("failed to subscribe to retry tasks: %w", err)
	}

	return nil
}

// handleImmediateTask traite une tâche immédiate
func (b *BackgroundWorkerService) handleImmediateTask(ctx context.Context, event *Event) error {
	var task Task
	if err := json.Unmarshal(event.Data.([]byte), &task); err != nil {
		return fmt.Errorf("failed to parse task: %w", err)
	}

	return b.dispatchTask(&task)
}

// handleScheduledTask traite une tâche programmée
func (b *BackgroundWorkerService) handleScheduledTask(ctx context.Context, event *Event) error {
	var task Task
	if err := json.Unmarshal(event.Data.([]byte), &task); err != nil {
		return fmt.Errorf("failed to parse task: %w", err)
	}

	// Vérifier si c'est le moment d'exécuter
	if task.ScheduledAt != nil && task.ScheduledAt.After(time.Now()) {
		// Reprogrammer la tâche
		return b.rescheduleTask(&task, *task.ScheduledAt)
	}

	return b.dispatchTask(&task)
}

// handleRetryTask traite une tâche en retry
func (b *BackgroundWorkerService) handleRetryTask(ctx context.Context, event *Event) error {
	var task Task
	if err := json.Unmarshal(event.Data.([]byte), &task); err != nil {
		return fmt.Errorf("failed to parse retry task: %w", err)
	}

	task.RetryCount++
	task.Status = TaskStatusRetrying

	return b.dispatchTask(&task)
}

// dispatchTask dispatch une tâche vers le bon pool
func (b *BackgroundWorkerService) dispatchTask(task *Task) error {
	b.mutex.RLock()
	pool, exists := b.workerPools[task.Type]
	b.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("no worker pool for task type: %s", task.Type)
	}

	// Envoyer la tâche au pool
	select {
	case pool.taskQueue <- task:
		b.recordTaskDispatched(task.Type)
		return nil
	default:
		return fmt.Errorf("worker pool queue full for task type: %s", task.Type)
	}
}

// ============================================================================
// PROCESSORS PAR DÉFAUT
// ============================================================================

// registerDefaultProcessors enregistre les processors par défaut
func (b *BackgroundWorkerService) registerDefaultProcessors() error {
	processors := []TaskProcessor{
		&ImageResizeProcessor{logger: b.logger},
		&DataExportProcessor{logger: b.logger},
		&EmailBatchProcessor{logger: b.logger},
		&AnalyticsComputeProcessor{logger: b.logger},
		&SystemMaintenanceProcessor{logger: b.logger},
	}

	for _, processor := range processors {
		if err := b.RegisterTaskProcessor(processor); err != nil {
			return fmt.Errorf("failed to register processor %T: %w", processor, err)
		}
	}

	return nil
}

// ImageResizeProcessor processor pour redimensionner les images
type ImageResizeProcessor struct {
	logger *zap.Logger
}

func (p *ImageResizeProcessor) GetTaskType() TaskType {
	return TaskImageResize
}

func (p *ImageResizeProcessor) ProcessTask(ctx context.Context, task *Task) (*TaskResult, error) {
	// Simuler le traitement d'image
	p.logger.Info("Processing image resize", zap.String("task_id", task.ID))

	// Simuler du travail
	time.Sleep(100 * time.Millisecond)

	result := &TaskResult{
		TaskID: task.ID,
		Result: map[string]interface{}{
			"resized_url": "https://example.com/resized/image.jpg",
			"width":       800,
			"height":      600,
		},
	}

	return result, nil
}

// DataExportProcessor processor pour exporter des données
type DataExportProcessor struct {
	logger *zap.Logger
}

func (p *DataExportProcessor) GetTaskType() TaskType {
	return TaskDataExport
}

func (p *DataExportProcessor) ProcessTask(ctx context.Context, task *Task) (*TaskResult, error) {
	p.logger.Info("Processing data export", zap.String("task_id", task.ID))

	// Simuler l'export de données
	time.Sleep(200 * time.Millisecond)

	result := &TaskResult{
		TaskID: task.ID,
		Result: map[string]interface{}{
			"export_url": "https://example.com/exports/data.csv",
			"row_count":  1000,
		},
	}

	return result, nil
}

// EmailBatchProcessor processor pour envoyer des emails en lot
type EmailBatchProcessor struct {
	logger *zap.Logger
}

func (p *EmailBatchProcessor) GetTaskType() TaskType {
	return TaskEmailBatch
}

func (p *EmailBatchProcessor) ProcessTask(ctx context.Context, task *Task) (*TaskResult, error) {
	p.logger.Info("Processing email batch", zap.String("task_id", task.ID))

	// Simuler l'envoi d'emails
	time.Sleep(500 * time.Millisecond)

	result := &TaskResult{
		TaskID: task.ID,
		Result: map[string]interface{}{
			"emails_sent":  100,
			"success_rate": 0.95,
		},
	}

	return result, nil
}

// AnalyticsComputeProcessor processor pour calculer les analytics
type AnalyticsComputeProcessor struct {
	logger *zap.Logger
}

func (p *AnalyticsComputeProcessor) GetTaskType() TaskType {
	return TaskAnalyticsCompute
}

func (p *AnalyticsComputeProcessor) ProcessTask(ctx context.Context, task *Task) (*TaskResult, error) {
	p.logger.Info("Processing analytics compute", zap.String("task_id", task.ID))

	// Simuler le calcul d'analytics
	time.Sleep(300 * time.Millisecond)

	result := &TaskResult{
		TaskID: task.ID,
		Result: map[string]interface{}{
			"dau":            1500,
			"mau":            45000,
			"retention_rate": 0.75,
		},
	}

	return result, nil
}

// SystemMaintenanceProcessor processor pour la maintenance système
type SystemMaintenanceProcessor struct {
	logger *zap.Logger
}

func (p *SystemMaintenanceProcessor) GetTaskType() TaskType {
	return TaskSystemMaintenance
}

func (p *SystemMaintenanceProcessor) ProcessTask(ctx context.Context, task *Task) (*TaskResult, error) {
	p.logger.Info("Processing system maintenance", zap.String("task_id", task.ID))

	// Simuler la maintenance
	time.Sleep(1 * time.Second)

	result := &TaskResult{
		TaskID: task.ID,
		Result: map[string]interface{}{
			"cleaned_files":  150,
			"freed_space_mb": 2048,
		},
	}

	return result, nil
}

// ============================================================================
// UTILITAIRES
// ============================================================================

// validateTask valide une tâche
func (b *BackgroundWorkerService) validateTask(task *Task) error {
	if task.ID == "" {
		return fmt.Errorf("task ID is required")
	}

	if task.Type == "" {
		return fmt.Errorf("task type is required")
	}

	// Vérifier que le processor existe
	b.mutex.RLock()
	_, exists := b.workerPools[task.Type]
	b.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("no processor registered for task type: %s", task.Type)
	}

	return nil
}

// getSubjectForTask retourne le sujet NATS pour une tâche
func (b *BackgroundWorkerService) getSubjectForTask(task *Task) string {
	if task.ScheduledAt != nil && task.ScheduledAt.After(time.Now()) {
		return TaskSubjectScheduled
	}
	return TaskSubjectImmediate
}

// convertTaskPriority convertit la priorité de tâche en priorité d'événement
func (b *BackgroundWorkerService) convertTaskPriority(priority TaskPriority) EventPriority {
	switch priority {
	case TaskPriorityLow:
		return PriorityLow
	case TaskPriorityNormal:
		return PriorityNormal
	case TaskPriorityHigh:
		return PriorityHigh
	case TaskPriorityCritical:
		return PriorityCritical
	default:
		return PriorityNormal
	}
}

// generateTaskID génère un ID unique pour la tâche
func (b *BackgroundWorkerService) generateTaskID() string {
	return fmt.Sprintf("task_%d_%d", time.Now().UnixNano(), b.metrics.TasksTotal)
}

// rescheduleTask reprogramme une tâche
func (b *BackgroundWorkerService) rescheduleTask(task *Task, scheduledAt time.Time) error {
	delay := time.Until(scheduledAt)

	go func() {
		time.Sleep(delay)

		// Remettre en queue
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := b.SubmitTask(ctx, task); err != nil {
			b.logger.Error("Failed to reschedule task",
				zap.String("task_id", task.ID),
				zap.Error(err))
		}
	}()

	return nil
}

// ============================================================================
// MÉTRIQUES
// ============================================================================

func (m *WorkerMetrics) recordTaskSubmitted(taskType TaskType, priority TaskPriority) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.TasksTotal++
	m.QueuedTasks++
	m.TasksByType[string(taskType)]++
	m.TasksByPriority[string(priority)]++
}

func (m *WorkerMetrics) recordTaskCompleted(taskType TaskType, duration time.Duration) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.TasksCompleted++
	m.ActiveTasks--
	if m.ActiveTasks < 0 {
		m.ActiveTasks = 0
	}
	m.AvgProcessingTime = (m.AvgProcessingTime + duration) / 2
}

func (m *WorkerMetrics) recordTaskFailed(taskType TaskType, duration time.Duration) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.TasksFailed++
	m.ActiveTasks--
	if m.ActiveTasks < 0 {
		m.ActiveTasks = 0
	}
}

func (b *BackgroundWorkerService) recordTaskSubmitted(taskType TaskType, priority TaskPriority) {
	b.metrics.recordTaskSubmitted(taskType, priority)
}

func (b *BackgroundWorkerService) recordTaskDispatched(taskType TaskType) {
	b.metrics.mutex.Lock()
	defer b.metrics.mutex.Unlock()

	b.metrics.QueuedTasks--
	b.metrics.ActiveTasks++
}

// GetMetrics retourne les métriques des workers
func (b *BackgroundWorkerService) GetMetrics() *WorkerMetrics {
	b.metrics.mutex.RLock()
	defer b.metrics.mutex.RUnlock()

	// Copier les métriques
	tasksByType := make(map[string]int64)
	for k, v := range b.metrics.TasksByType {
		tasksByType[k] = v
	}

	tasksByPriority := make(map[string]int64)
	for k, v := range b.metrics.TasksByPriority {
		tasksByPriority[k] = v
	}

	workersByType := make(map[string]int)
	for k, v := range b.metrics.WorkersByType {
		workersByType[k] = v
	}

	return &WorkerMetrics{
		TasksTotal:        b.metrics.TasksTotal,
		TasksCompleted:    b.metrics.TasksCompleted,
		TasksFailed:       b.metrics.TasksFailed,
		TasksRetried:      b.metrics.TasksRetried,
		TasksCancelled:    b.metrics.TasksCancelled,
		ActiveTasks:       b.metrics.ActiveTasks,
		QueuedTasks:       b.metrics.QueuedTasks,
		AvgProcessingTime: b.metrics.AvgProcessingTime,
		TasksByType:       tasksByType,
		TasksByPriority:   tasksByPriority,
		WorkersByType:     workersByType,
	}
}

// ============================================================================
// MONITORING
// ============================================================================

// startHealthChecker démarre le vérificateur de santé
func (b *BackgroundWorkerService) startHealthChecker() {
	ticker := time.NewTicker(b.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Publier un heartbeat
			event := &Event{
				ID:        fmt.Sprintf("heartbeat_%d", time.Now().Unix()),
				Type:      EventType("worker.heartbeat"),
				Source:    "background_worker_service",
				Subject:   TaskSubjectHeartbeat,
				Data:      b.GetMetrics(),
				Priority:  PriorityLow,
				Timestamp: time.Now(),
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			if err := b.natsService.PublishEvent(ctx, event); err != nil {
				b.logger.Warn("Failed to publish heartbeat", zap.Error(err))
			}
			cancel()

		case <-b.ctx.Done():
			return
		}
	}
}

// startMetricsReporter démarre le reporter de métriques
func (b *BackgroundWorkerService) startMetricsReporter() {
	ticker := time.NewTicker(b.config.MetricsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			metrics := b.GetMetrics()
			b.logger.Info("Worker metrics",
				zap.Int64("tasks_completed", metrics.TasksCompleted),
				zap.Int64("tasks_failed", metrics.TasksFailed),
				zap.Int64("active_tasks", metrics.ActiveTasks),
				zap.Int64("queued_tasks", metrics.QueuedTasks),
				zap.Duration("avg_processing_time", metrics.AvgProcessingTime))

		case <-b.ctx.Done():
			return
		}
	}
}

// HealthCheck vérifie la santé du service
func (b *BackgroundWorkerService) HealthCheck() error {
	b.mutex.RLock()
	poolCount := len(b.workerPools)
	b.mutex.RUnlock()

	if poolCount == 0 {
		return fmt.Errorf("no worker pools registered")
	}

	return nil
}

// Close ferme proprement le service
func (b *BackgroundWorkerService) Close() error {
	b.cancel()

	// Fermer tous les pools
	b.mutex.Lock()
	for _, pool := range b.workerPools {
		pool.cancel()
	}
	b.mutex.Unlock()

	// Attendre que tous les workers terminent
	done := make(chan struct{})
	go func() {
		b.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(30 * time.Second):
		b.logger.Warn("Timeout waiting for workers to finish")
	}

	b.logger.Info("Background worker service closed")
	return nil
}
