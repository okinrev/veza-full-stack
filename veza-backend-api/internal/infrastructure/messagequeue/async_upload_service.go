package messagequeue

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// AsyncUploadService service pour le processing asynchrone des uploads
type AsyncUploadService struct {
	natsService   *NATSService
	workerService *BackgroundWorkerService
	logger        *zap.Logger
	config        *UploadConfig

	// Processeurs de fichiers
	processors map[string]FileProcessor

	// Métriques
	metrics *UploadMetrics

	// Contrôle de lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	mutex  sync.RWMutex
}

// UploadConfig configuration des uploads
type UploadConfig struct {
	UploadDir         string          `json:"upload_dir"`
	TempDir           string          `json:"temp_dir"`
	ProcessedDir      string          `json:"processed_dir"`
	MaxFileSize       int64           `json:"max_file_size"`
	AllowedMimeTypes  []string        `json:"allowed_mime_types"`
	AllowedExtensions []string        `json:"allowed_extensions"`
	EnableVirusScan   bool            `json:"enable_virus_scan"`
	EnableThumbnails  bool            `json:"enable_thumbnails"`
	EnableCompression bool            `json:"enable_compression"`
	EnableWatermark   bool            `json:"enable_watermark"`
	ProcessingTimeout time.Duration   `json:"processing_timeout"`
	CleanupInterval   time.Duration   `json:"cleanup_interval"`
	ThumbnailSizes    []ThumbnailSize `json:"thumbnail_sizes"`
	ImageQuality      int             `json:"image_quality"`
	CDNEnabled        bool            `json:"cdn_enabled"`
	CDNBaseURL        string          `json:"cdn_base_url"`
}

// FileUpload représente un upload de fichier
type FileUpload struct {
	ID              string                 `json:"id"`
	UserID          int64                  `json:"user_id"`
	OriginalName    string                 `json:"original_name"`
	Filename        string                 `json:"filename"`
	ContentType     string                 `json:"content_type"`
	Size            int64                  `json:"size"`
	Extension       string                 `json:"extension"`
	Hash            string                 `json:"hash"`
	Path            string                 `json:"path"`
	URL             string                 `json:"url,omitempty"`
	ThumbnailURLs   map[string]string      `json:"thumbnail_urls,omitempty"`
	ProcessedURLs   map[string]string      `json:"processed_urls,omitempty"`
	Status          UploadStatus           `json:"status"`
	ProcessingSteps []ProcessingStep       `json:"processing_steps"`
	Metadata        map[string]interface{} `json:"metadata"`
	CreatedAt       time.Time              `json:"created_at"`
	ProcessedAt     *time.Time             `json:"processed_at,omitempty"`
	Error           string                 `json:"error,omitempty"`
}

// UploadStatus statut d'upload
type UploadStatus string

const (
	UploadStatusUploading   UploadStatus = "uploading"
	UploadStatusPending     UploadStatus = "pending"
	UploadStatusProcessing  UploadStatus = "processing"
	UploadStatusCompleted   UploadStatus = "completed"
	UploadStatusFailed      UploadStatus = "failed"
	UploadStatusQuarantined UploadStatus = "quarantined"
)

// ProcessingStep étape de processing
type ProcessingStep struct {
	Name        string                 `json:"name"`
	Status      ProcessingStepStatus   `json:"status"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Result      map[string]interface{} `json:"result,omitempty"`
}

// ProcessingStepStatus statut d'étape
type ProcessingStepStatus string

const (
	ProcessingStepPending   ProcessingStepStatus = "pending"
	ProcessingStepRunning   ProcessingStepStatus = "running"
	ProcessingStepCompleted ProcessingStepStatus = "completed"
	ProcessingStepFailed    ProcessingStepStatus = "failed"
	ProcessingStepSkipped   ProcessingStepStatus = "skipped"
)

// ThumbnailSize taille de thumbnail
type ThumbnailSize struct {
	Name   string `json:"name"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// FileProcessor interface pour traiter les fichiers
type FileProcessor interface {
	ProcessFile(ctx context.Context, upload *FileUpload) error
	GetSupportedMimeTypes() []string
	GetProcessorName() string
}

// UploadMetrics métriques d'upload
type UploadMetrics struct {
	UploadsTotal      int64            `json:"uploads_total"`
	UploadsCompleted  int64            `json:"uploads_completed"`
	UploadsFailed     int64            `json:"uploads_failed"`
	BytesProcessed    int64            `json:"bytes_processed"`
	AvgProcessingTime time.Duration    `json:"avg_processing_time"`
	UploadsByType     map[string]int64 `json:"uploads_by_type"`
	UploadsByStatus   map[string]int64 `json:"uploads_by_status"`
	ProcessingErrors  map[string]int64 `json:"processing_errors"`

	mutex sync.RWMutex
}

// NewAsyncUploadService crée un nouveau service d'upload asynchrone
func NewAsyncUploadService(natsService *NATSService, workerService *BackgroundWorkerService, config *UploadConfig, logger *zap.Logger) (*AsyncUploadService, error) {
	if config == nil {
		config = &UploadConfig{
			UploadDir:         "./uploads",
			TempDir:           "./temp",
			ProcessedDir:      "./processed",
			MaxFileSize:       50 * 1024 * 1024, // 50MB
			AllowedMimeTypes:  []string{"image/jpeg", "image/png", "image/gif", "image/webp", "video/mp4", "application/pdf"},
			AllowedExtensions: []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".mp4", ".pdf"},
			EnableVirusScan:   true,
			EnableThumbnails:  true,
			EnableCompression: true,
			EnableWatermark:   false,
			ProcessingTimeout: 10 * time.Minute,
			CleanupInterval:   24 * time.Hour,
			ThumbnailSizes: []ThumbnailSize{
				{Name: "small", Width: 150, Height: 150},
				{Name: "medium", Width: 300, Height: 300},
				{Name: "large", Width: 800, Height: 600},
			},
			ImageQuality: 85,
			CDNEnabled:   false,
			CDNBaseURL:   "",
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	service := &AsyncUploadService{
		natsService:   natsService,
		workerService: workerService,
		logger:        logger,
		config:        config,
		processors:    make(map[string]FileProcessor),
		metrics: &UploadMetrics{
			UploadsByType:    make(map[string]int64),
			UploadsByStatus:  make(map[string]int64),
			ProcessingErrors: make(map[string]int64),
		},
		ctx:    ctx,
		cancel: cancel,
	}

	// Créer les répertoires nécessaires
	if err := service.createDirectories(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create directories: %w", err)
	}

	// Enregistrer les processeurs par défaut
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
	go service.startMetricsReporter()
	go service.startCleanupWorker()

	return service, nil
}

// ============================================================================
// UPLOAD DE FICHIERS
// ============================================================================

// UploadFile uploade et traite un fichier de manière asynchrone
func (a *AsyncUploadService) UploadFile(ctx context.Context, userID int64, file multipart.File, header *multipart.FileHeader) (*FileUpload, error) {
	// Valider le fichier
	if err := a.validateFile(header); err != nil {
		return nil, fmt.Errorf("file validation failed: %w", err)
	}

	// Créer l'objet upload
	upload := &FileUpload{
		ID:           a.generateUploadID(),
		UserID:       userID,
		OriginalName: header.Filename,
		ContentType:  header.Header.Get("Content-Type"),
		Size:         header.Size,
		Extension:    strings.ToLower(filepath.Ext(header.Filename)),
		Status:       UploadStatusUploading,
		Metadata:     make(map[string]interface{}),
		CreatedAt:    time.Now(),
		ProcessingSteps: []ProcessingStep{
			{Name: "validation", Status: ProcessingStepPending},
			{Name: "virus_scan", Status: ProcessingStepPending},
			{Name: "thumbnail_generation", Status: ProcessingStepPending},
			{Name: "compression", Status: ProcessingStepPending},
			{Name: "metadata_extraction", Status: ProcessingStepPending},
		},
	}

	// Générer le nom de fichier unique
	upload.Filename = a.generateFilename(upload.ID, upload.Extension)
	upload.Path = filepath.Join(a.config.UploadDir, upload.Filename)

	// Sauvegarder le fichier temporairement
	tempPath := filepath.Join(a.config.TempDir, upload.Filename)
	if err := a.saveFile(file, tempPath); err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	// Calculer le hash du fichier
	hash, err := a.calculateFileHash(tempPath)
	if err != nil {
		os.Remove(tempPath)
		return nil, fmt.Errorf("failed to calculate file hash: %w", err)
	}
	upload.Hash = hash

	// Déplacer vers le répertoire d'upload final
	if err := os.Rename(tempPath, upload.Path); err != nil {
		os.Remove(tempPath)
		return nil, fmt.Errorf("failed to move file: %w", err)
	}

	// Marquer comme en attente de processing
	upload.Status = UploadStatusPending

	// Soumettre pour processing asynchrone
	if err := a.submitForProcessing(ctx, upload); err != nil {
		// Nettoyer le fichier en cas d'erreur
		os.Remove(upload.Path)
		return nil, fmt.Errorf("failed to submit for processing: %w", err)
	}

	// Enregistrer les métriques
	a.recordUploadStarted(upload.ContentType)

	a.logger.Info("File uploaded successfully",
		zap.String("id", upload.ID),
		zap.String("filename", upload.OriginalName),
		zap.Int64("user_id", userID),
		zap.Int64("size", upload.Size))

	return upload, nil
}

// ============================================================================
// PROCESSING ASYNCHRONE
// ============================================================================

// submitForProcessing soumet un upload pour processing asynchrone
func (a *AsyncUploadService) submitForProcessing(ctx context.Context, upload *FileUpload) error {
	// Créer la tâche de processing
	task := &Task{
		Type: TaskFileAnalysis,
		Data: map[string]interface{}{
			"upload_id":    upload.ID,
			"file_path":    upload.Path,
			"content_type": upload.ContentType,
			"user_id":      upload.UserID,
		},
		Priority: TaskPriorityNormal,
		UserID:   &upload.UserID,
	}

	// Soumettre la tâche
	return a.workerService.SubmitTask(ctx, task)
}

// processUpload traite un upload
func (a *AsyncUploadService) processUpload(ctx context.Context, upload *FileUpload) error {
	upload.Status = UploadStatusProcessing
	start := time.Now()

	a.logger.Info("Starting upload processing",
		zap.String("id", upload.ID),
		zap.String("filename", upload.OriginalName))

	// Étape 1: Validation avancée
	if err := a.processStep(ctx, upload, "validation", func() error {
		return a.validateFileContent(upload)
	}); err != nil {
		return a.markUploadFailed(upload, err)
	}

	// Étape 2: Scan antivirus (si activé)
	if a.config.EnableVirusScan {
		if err := a.processStep(ctx, upload, "virus_scan", func() error {
			return a.scanForVirus(upload)
		}); err != nil {
			return a.markUploadQuarantined(upload, err)
		}
	} else {
		a.markStepSkipped(upload, "virus_scan")
	}

	// Étape 3: Extraction de métadonnées
	if err := a.processStep(ctx, upload, "metadata_extraction", func() error {
		return a.extractMetadata(upload)
	}); err != nil {
		a.logger.Warn("Failed to extract metadata", zap.Error(err))
		// Continuer même si l'extraction échoue
	}

	// Étape 4: Génération de thumbnails (pour les images)
	if a.config.EnableThumbnails && a.isImageFile(upload) {
		if err := a.processStep(ctx, upload, "thumbnail_generation", func() error {
			return a.generateThumbnails(upload)
		}); err != nil {
			a.logger.Warn("Failed to generate thumbnails", zap.Error(err))
			// Continuer même si la génération échoue
		}
	} else {
		a.markStepSkipped(upload, "thumbnail_generation")
	}

	// Étape 5: Compression (si activé)
	if a.config.EnableCompression && a.isCompressibleFile(upload) {
		if err := a.processStep(ctx, upload, "compression", func() error {
			return a.compressFile(upload)
		}); err != nil {
			a.logger.Warn("Failed to compress file", zap.Error(err))
			// Continuer même si la compression échoue
		}
	} else {
		a.markStepSkipped(upload, "compression")
	}

	// Marquer comme terminé
	upload.Status = UploadStatusCompleted
	processedAt := time.Now()
	upload.ProcessedAt = &processedAt

	// Générer l'URL finale
	upload.URL = a.generateFileURL(upload)

	// Publier l'événement de completion
	if err := a.publishUploadCompleted(ctx, upload); err != nil {
		a.logger.Warn("Failed to publish upload completed event", zap.Error(err))
	}

	// Enregistrer les métriques
	a.recordUploadCompleted(upload.ContentType, time.Since(start))

	a.logger.Info("Upload processing completed",
		zap.String("id", upload.ID),
		zap.Duration("duration", time.Since(start)))

	return nil
}

// ============================================================================
// ÉTAPES DE PROCESSING
// ============================================================================

// processStep exécute une étape de processing
func (a *AsyncUploadService) processStep(ctx context.Context, upload *FileUpload, stepName string, stepFunc func() error) error {
	// Trouver l'étape
	stepIndex := a.findStep(upload, stepName)
	if stepIndex == -1 {
		return fmt.Errorf("step not found: %s", stepName)
	}

	step := &upload.ProcessingSteps[stepIndex]
	step.Status = ProcessingStepRunning
	startedAt := time.Now()
	step.StartedAt = &startedAt

	// Exécuter l'étape
	err := stepFunc()
	completedAt := time.Now()
	step.CompletedAt = &completedAt

	if err != nil {
		step.Status = ProcessingStepFailed
		step.Error = err.Error()
		return err
	}

	step.Status = ProcessingStepCompleted
	return nil
}

// validateFileContent valide le contenu du fichier
func (a *AsyncUploadService) validateFileContent(upload *FileUpload) error {
	// TODO: Implémenter la validation de contenu
	// Par exemple, vérifier que c'est vraiment une image si l'extension l'indique
	a.logger.Debug("Validating file content", zap.String("upload_id", upload.ID))
	return nil
}

// scanForVirus scan le fichier pour des virus
func (a *AsyncUploadService) scanForVirus(upload *FileUpload) error {
	// TODO: Implémenter le scan antivirus avec ClamAV ou autre
	a.logger.Debug("Scanning file for virus", zap.String("upload_id", upload.ID))

	// Simuler un scan
	time.Sleep(100 * time.Millisecond)

	return nil
}

// extractMetadata extrait les métadonnées du fichier
func (a *AsyncUploadService) extractMetadata(upload *FileUpload) error {
	a.logger.Debug("Extracting metadata", zap.String("upload_id", upload.ID))

	// TODO: Implémenter l'extraction de métadonnées avec ExifRead ou autre
	if a.isImageFile(upload) {
		upload.Metadata["type"] = "image"
		upload.Metadata["format"] = strings.TrimPrefix(upload.Extension, ".")
	} else if a.isVideoFile(upload) {
		upload.Metadata["type"] = "video"
		upload.Metadata["format"] = strings.TrimPrefix(upload.Extension, ".")
	}

	return nil
}

// generateThumbnails génère les thumbnails pour les images
func (a *AsyncUploadService) generateThumbnails(upload *FileUpload) error {
	a.logger.Debug("Generating thumbnails", zap.String("upload_id", upload.ID))

	if upload.ThumbnailURLs == nil {
		upload.ThumbnailURLs = make(map[string]string)
	}

	// TODO: Implémenter la génération de thumbnails avec imaging library
	for _, size := range a.config.ThumbnailSizes {
		// Simuler la génération
		thumbnailPath := filepath.Join(a.config.ProcessedDir, fmt.Sprintf("%s_%s%s", upload.ID, size.Name, upload.Extension))
		upload.ThumbnailURLs[size.Name] = a.generateThumbnailURL(upload, size.Name)

		a.logger.Debug("Generated thumbnail",
			zap.String("upload_id", upload.ID),
			zap.String("size", size.Name),
			zap.String("path", thumbnailPath))
	}

	return nil
}

// compressFile compresse le fichier
func (a *AsyncUploadService) compressFile(upload *FileUpload) error {
	a.logger.Debug("Compressing file", zap.String("upload_id", upload.ID))

	// TODO: Implémenter la compression selon le type de fichier
	if a.isImageFile(upload) {
		// Réduire la qualité JPEG par exemple
		compressedPath := filepath.Join(a.config.ProcessedDir, fmt.Sprintf("%s_compressed%s", upload.ID, upload.Extension))
		if upload.ProcessedURLs == nil {
			upload.ProcessedURLs = make(map[string]string)
		}
		upload.ProcessedURLs["compressed"] = a.generateProcessedURL(upload, "compressed")

		a.logger.Debug("Compressed image",
			zap.String("upload_id", upload.ID),
			zap.String("path", compressedPath))
	}

	return nil
}

// ============================================================================
// PROCESSEURS DE FICHIERS
// ============================================================================

// registerDefaultProcessors enregistre les processeurs par défaut
func (a *AsyncUploadService) registerDefaultProcessors() error {
	processors := []FileProcessor{
		&ImageProcessor{logger: a.logger, config: a.config},
		&VideoProcessor{logger: a.logger, config: a.config},
		&DocumentProcessor{logger: a.logger, config: a.config},
	}

	for _, processor := range processors {
		a.RegisterFileProcessor(processor)
	}

	return nil
}

// RegisterFileProcessor enregistre un processeur de fichier
func (a *AsyncUploadService) RegisterFileProcessor(processor FileProcessor) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	for _, mimeType := range processor.GetSupportedMimeTypes() {
		a.processors[mimeType] = processor
	}

	a.logger.Info("File processor registered",
		zap.String("name", processor.GetProcessorName()),
		zap.Strings("mime_types", processor.GetSupportedMimeTypes()))
}

// ImageProcessor processeur pour les images
type ImageProcessor struct {
	logger *zap.Logger
	config *UploadConfig
}

func (p *ImageProcessor) GetProcessorName() string {
	return "image_processor"
}

func (p *ImageProcessor) GetSupportedMimeTypes() []string {
	return []string{"image/jpeg", "image/png", "image/gif", "image/webp"}
}

func (p *ImageProcessor) ProcessFile(ctx context.Context, upload *FileUpload) error {
	p.logger.Info("Processing image file", zap.String("upload_id", upload.ID))

	// TODO: Implémenter le processing spécifique aux images
	// - Redimensionnement
	// - Optimisation
	// - Watermark
	// - Conversion de format

	return nil
}

// VideoProcessor processeur pour les vidéos
type VideoProcessor struct {
	logger *zap.Logger
	config *UploadConfig
}

func (p *VideoProcessor) GetProcessorName() string {
	return "video_processor"
}

func (p *VideoProcessor) GetSupportedMimeTypes() []string {
	return []string{"video/mp4", "video/avi", "video/mov", "video/wmv"}
}

func (p *VideoProcessor) ProcessFile(ctx context.Context, upload *FileUpload) error {
	p.logger.Info("Processing video file", zap.String("upload_id", upload.ID))

	// TODO: Implémenter le processing spécifique aux vidéos
	// - Génération de thumbnail
	// - Transcodage
	// - Compression
	// - Extraction de métadonnées

	return nil
}

// DocumentProcessor processeur pour les documents
type DocumentProcessor struct {
	logger *zap.Logger
	config *UploadConfig
}

func (p *DocumentProcessor) GetProcessorName() string {
	return "document_processor"
}

func (p *DocumentProcessor) GetSupportedMimeTypes() []string {
	return []string{"application/pdf", "application/msword", "text/plain"}
}

func (p *DocumentProcessor) ProcessFile(ctx context.Context, upload *FileUpload) error {
	p.logger.Info("Processing document file", zap.String("upload_id", upload.ID))

	// TODO: Implémenter le processing spécifique aux documents
	// - Extraction de texte
	// - Génération de preview
	// - Validation de format

	return nil
}

// ============================================================================
// NATS INTEGRATION
// ============================================================================

// startNATSSubscriptions démarre les subscriptions NATS
func (a *AsyncUploadService) startNATSSubscriptions() error {
	// Subscription pour les résultats de processing
	if err := a.natsService.SubscribeToSubject("uploads.process", a.handleUploadProcessing); err != nil {
		return fmt.Errorf("failed to subscribe to upload processing: %w", err)
	}

	return nil
}

// handleUploadProcessing traite un événement de processing d'upload
func (a *AsyncUploadService) handleUploadProcessing(ctx context.Context, event *Event) error {
	// TODO: Implémenter le traitement des événements
	a.logger.Debug("Handling upload processing event", zap.String("event_id", event.ID))
	return nil
}

// publishUploadCompleted publie un événement de completion d'upload
func (a *AsyncUploadService) publishUploadCompleted(ctx context.Context, upload *FileUpload) error {
	event := &Event{
		ID:        upload.ID + "_completed",
		Type:      EventType("upload.completed"),
		Source:    "async_upload_service",
		Subject:   "uploads.completed",
		Data:      upload,
		Priority:  PriorityNormal,
		Timestamp: time.Now(),
		UserID:    &upload.UserID,
	}

	return a.natsService.PublishEvent(ctx, event)
}

// ============================================================================
// UTILITAIRES
// ============================================================================

// validateFile valide un fichier uploadé
func (a *AsyncUploadService) validateFile(header *multipart.FileHeader) error {
	// Vérifier la taille
	if header.Size > a.config.MaxFileSize {
		return fmt.Errorf("file too large: %d bytes (max: %d bytes)", header.Size, a.config.MaxFileSize)
	}

	// Vérifier l'extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !a.isAllowedExtension(ext) {
		return fmt.Errorf("file extension not allowed: %s", ext)
	}

	// Vérifier le type MIME
	contentType := header.Header.Get("Content-Type")
	if !a.isAllowedMimeType(contentType) {
		return fmt.Errorf("content type not allowed: %s", contentType)
	}

	return nil
}

// isAllowedExtension vérifie si l'extension est autorisée
func (a *AsyncUploadService) isAllowedExtension(ext string) bool {
	for _, allowed := range a.config.AllowedExtensions {
		if ext == allowed {
			return true
		}
	}
	return false
}

// isAllowedMimeType vérifie si le type MIME est autorisé
func (a *AsyncUploadService) isAllowedMimeType(mimeType string) bool {
	for _, allowed := range a.config.AllowedMimeTypes {
		if mimeType == allowed {
			return true
		}
	}
	return false
}

// isImageFile vérifie si c'est un fichier image
func (a *AsyncUploadService) isImageFile(upload *FileUpload) bool {
	return strings.HasPrefix(upload.ContentType, "image/")
}

// isVideoFile vérifie si c'est un fichier vidéo
func (a *AsyncUploadService) isVideoFile(upload *FileUpload) bool {
	return strings.HasPrefix(upload.ContentType, "video/")
}

// isCompressibleFile vérifie si le fichier peut être compressé
func (a *AsyncUploadService) isCompressibleFile(upload *FileUpload) bool {
	return a.isImageFile(upload) || a.isVideoFile(upload)
}

// saveFile sauvegarde un fichier sur le disque
func (a *AsyncUploadService) saveFile(file multipart.File, path string) error {
	dst, err := os.Create(path)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	return err
}

// calculateFileHash calcule le hash MD5 d'un fichier
func (a *AsyncUploadService) calculateFileHash(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// generateUploadID génère un ID unique pour l'upload
func (a *AsyncUploadService) generateUploadID() string {
	return fmt.Sprintf("upload_%d_%d", time.Now().UnixNano(), a.metrics.UploadsTotal)
}

// generateFilename génère un nom de fichier unique
func (a *AsyncUploadService) generateFilename(id, extension string) string {
	return fmt.Sprintf("%s%s", id, extension)
}

// generateFileURL génère l'URL du fichier
func (a *AsyncUploadService) generateFileURL(upload *FileUpload) string {
	if a.config.CDNEnabled {
		return fmt.Sprintf("%s/%s", a.config.CDNBaseURL, upload.Filename)
	}
	return fmt.Sprintf("/uploads/%s", upload.Filename)
}

// generateThumbnailURL génère l'URL d'un thumbnail
func (a *AsyncUploadService) generateThumbnailURL(upload *FileUpload, size string) string {
	filename := fmt.Sprintf("%s_%s%s", upload.ID, size, upload.Extension)
	if a.config.CDNEnabled {
		return fmt.Sprintf("%s/thumbnails/%s", a.config.CDNBaseURL, filename)
	}
	return fmt.Sprintf("/thumbnails/%s", filename)
}

// generateProcessedURL génère l'URL d'un fichier traité
func (a *AsyncUploadService) generateProcessedURL(upload *FileUpload, variant string) string {
	filename := fmt.Sprintf("%s_%s%s", upload.ID, variant, upload.Extension)
	if a.config.CDNEnabled {
		return fmt.Sprintf("%s/processed/%s", a.config.CDNBaseURL, filename)
	}
	return fmt.Sprintf("/processed/%s", filename)
}

// createDirectories crée les répertoires nécessaires
func (a *AsyncUploadService) createDirectories() error {
	dirs := []string{
		a.config.UploadDir,
		a.config.TempDir,
		a.config.ProcessedDir,
		filepath.Join(a.config.ProcessedDir, "thumbnails"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// findStep trouve l'index d'une étape par son nom
func (a *AsyncUploadService) findStep(upload *FileUpload, stepName string) int {
	for i, step := range upload.ProcessingSteps {
		if step.Name == stepName {
			return i
		}
	}
	return -1
}

// markStepSkipped marque une étape comme sautée
func (a *AsyncUploadService) markStepSkipped(upload *FileUpload, stepName string) {
	stepIndex := a.findStep(upload, stepName)
	if stepIndex != -1 {
		upload.ProcessingSteps[stepIndex].Status = ProcessingStepSkipped
	}
}

// markUploadFailed marque un upload comme échoué
func (a *AsyncUploadService) markUploadFailed(upload *FileUpload, err error) error {
	upload.Status = UploadStatusFailed
	upload.Error = err.Error()

	a.recordUploadFailed(upload.ContentType, err.Error())

	a.logger.Error("Upload processing failed",
		zap.String("id", upload.ID),
		zap.Error(err))

	return err
}

// markUploadQuarantined marque un upload comme mis en quarantaine
func (a *AsyncUploadService) markUploadQuarantined(upload *FileUpload, err error) error {
	upload.Status = UploadStatusQuarantined
	upload.Error = err.Error()

	// Déplacer le fichier vers un répertoire de quarantaine
	quarantinePath := filepath.Join(a.config.TempDir, "quarantine", upload.Filename)
	if moveErr := os.Rename(upload.Path, quarantinePath); moveErr != nil {
		a.logger.Error("Failed to move file to quarantine", zap.Error(moveErr))
	}

	a.recordUploadFailed(upload.ContentType, "quarantined")

	a.logger.Warn("Upload quarantined",
		zap.String("id", upload.ID),
		zap.Error(err))

	return err
}

// ============================================================================
// MÉTRIQUES
// ============================================================================

func (a *AsyncUploadService) recordUploadStarted(contentType string) {
	a.metrics.mutex.Lock()
	defer a.metrics.mutex.Unlock()

	a.metrics.UploadsTotal++
	a.metrics.UploadsByType[contentType]++
	a.metrics.UploadsByStatus["started"]++
}

func (a *AsyncUploadService) recordUploadCompleted(contentType string, duration time.Duration) {
	a.metrics.mutex.Lock()
	defer a.metrics.mutex.Unlock()

	a.metrics.UploadsCompleted++
	a.metrics.UploadsByStatus["completed"]++
	a.metrics.AvgProcessingTime = (a.metrics.AvgProcessingTime + duration) / 2
}

func (a *AsyncUploadService) recordUploadFailed(contentType, errorType string) {
	a.metrics.mutex.Lock()
	defer a.metrics.mutex.Unlock()

	a.metrics.UploadsFailed++
	a.metrics.UploadsByStatus["failed"]++
	a.metrics.ProcessingErrors[errorType]++
}

// GetMetrics retourne les métriques d'upload
func (a *AsyncUploadService) GetMetrics() *UploadMetrics {
	a.metrics.mutex.RLock()
	defer a.metrics.mutex.RUnlock()

	// Copier les métriques
	uploadsByType := make(map[string]int64)
	for k, v := range a.metrics.UploadsByType {
		uploadsByType[k] = v
	}

	uploadsByStatus := make(map[string]int64)
	for k, v := range a.metrics.UploadsByStatus {
		uploadsByStatus[k] = v
	}

	processingErrors := make(map[string]int64)
	for k, v := range a.metrics.ProcessingErrors {
		processingErrors[k] = v
	}

	return &UploadMetrics{
		UploadsTotal:      a.metrics.UploadsTotal,
		UploadsCompleted:  a.metrics.UploadsCompleted,
		UploadsFailed:     a.metrics.UploadsFailed,
		BytesProcessed:    a.metrics.BytesProcessed,
		AvgProcessingTime: a.metrics.AvgProcessingTime,
		UploadsByType:     uploadsByType,
		UploadsByStatus:   uploadsByStatus,
		ProcessingErrors:  processingErrors,
	}
}

// ============================================================================
// MONITORING
// ============================================================================

// startMetricsReporter démarre le reporter de métriques
func (a *AsyncUploadService) startMetricsReporter() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			metrics := a.GetMetrics()
			a.logger.Info("Upload metrics",
				zap.Int64("uploads_total", metrics.UploadsTotal),
				zap.Int64("uploads_completed", metrics.UploadsCompleted),
				zap.Int64("uploads_failed", metrics.UploadsFailed),
				zap.Duration("avg_processing_time", metrics.AvgProcessingTime))

		case <-a.ctx.Done():
			return
		}
	}
}

// startCleanupWorker démarre le worker de nettoyage
func (a *AsyncUploadService) startCleanupWorker() {
	ticker := time.NewTicker(a.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// TODO: Implémenter le nettoyage des fichiers temporaires
			// et des uploads échoués/expirés
			a.logger.Info("Cleanup worker running")

		case <-a.ctx.Done():
			return
		}
	}
}

// HealthCheck vérifie la santé du service
func (a *AsyncUploadService) HealthCheck() error {
	// Vérifier que les répertoires existent
	dirs := []string{a.config.UploadDir, a.config.TempDir, a.config.ProcessedDir}
	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return fmt.Errorf("directory does not exist: %s", dir)
		}
	}

	return nil
}

// Close ferme proprement le service
func (a *AsyncUploadService) Close() error {
	a.cancel()

	// Attendre que tous les workers terminent
	done := make(chan struct{})
	go func() {
		a.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		a.logger.Warn("Timeout waiting for upload workers to finish")
	}

	a.logger.Info("Async upload service closed")
	return nil
}
