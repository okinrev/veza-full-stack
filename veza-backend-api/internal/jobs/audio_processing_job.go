package jobs

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Shopify/sarama"
	"github.com/go-redis/redis/v8"
	"github.com/hashicorp/vault/api"
	"go.uber.org/zap"
)

// AudioProcessingJob gère le traitement audio asynchrone
type AudioProcessingJob struct {
	db          *sql.DB
	cache       *redis.Client
	vaultClient *api.Client
	logger      *zap.Logger
	producer    sarama.SyncProducer
	consumer    sarama.Consumer
}

// AudioProcessingMessage représente un message de traitement audio
type AudioProcessingMessage struct {
	TrackID   int64     `json:"track_id"`
	AudioPath string    `json:"audio_path"`
	Tasks     []string  `json:"tasks"` // "tagging", "search", "stems", "mastering"
	Priority  int       `json:"priority"`
	CreatedAt time.Time `json:"created_at"`
}

// AudioProcessingResult représente le résultat du traitement
type AudioProcessingResult struct {
	TrackID        int64                  `json:"track_id"`
	Tasks          []string               `json:"tasks"`
	Status         string                 `json:"status"` // "completed", "failed", "partial"
	Results        map[string]interface{} `json:"results"`
	ProcessingTime float64                `json:"processing_time"`
	CreatedAt      time.Time              `json:"created_at"`
}

// NewAudioProcessingJob crée une nouvelle instance du job
func NewAudioProcessingJob(
	db *sql.DB,
	cache *redis.Client,
	vaultClient *api.Client,
	logger *zap.Logger,
	producer sarama.SyncProducer,
	consumer sarama.Consumer,
) *AudioProcessingJob {
	return &AudioProcessingJob{
		db:          db,
		cache:       cache,
		vaultClient: vaultClient,
		logger:      logger,
		producer:    producer,
		consumer:    consumer,
	}
}

// EnqueueAudioProcessing ajoute un track au traitement audio
func (j *AudioProcessingJob) EnqueueAudioProcessing(trackID int64, audioPath string, tasks []string, priority int) error {
	message := AudioProcessingMessage{
		TrackID:   trackID,
		AudioPath: audioPath,
		Tasks:     tasks,
		Priority:  priority,
		CreatedAt: time.Now(),
	}

	// Sérialiser le message
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Publier sur Kafka
	msg := &sarama.ProducerMessage{
		Topic: "audio-processing",
		Key:   sarama.StringEncoder(fmt.Sprintf("%d", trackID)),
		Value: sarama.ByteEncoder(data),
	}

	partition, offset, err := j.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	j.logger.Info("Audio processing enqueued",
		zap.Int64("track_id", trackID),
		zap.Int32("partition", partition),
		zap.Int64("offset", offset))

	return nil
}

// StartConsumer démarre le consommateur Kafka
func (j *AudioProcessingJob) StartConsumer(ctx context.Context) error {
	partitionConsumer, err := j.consumer.ConsumePartition("audio-processing", 0, sarama.OffsetNewest)
	if err != nil {
		return fmt.Errorf("failed to create partition consumer: %w", err)
	}
	defer partitionConsumer.Close()

	j.logger.Info("Audio processing consumer started")

	for {
		select {
		case <-ctx.Done():
			j.logger.Info("Audio processing consumer stopped")
			return nil

		case msg := <-partitionConsumer.Messages():
			err := j.processMessage(msg)
			if err != nil {
				j.logger.Error("Failed to process message", zap.Error(err))
			}
		}
	}
}

// processMessage traite un message de traitement audio
func (j *AudioProcessingJob) processMessage(msg *sarama.ConsumerMessage) error {
	var message AudioProcessingMessage
	err := json.Unmarshal(msg.Value, &message)
	if err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	j.logger.Info("Processing audio message",
		zap.Int64("track_id", message.TrackID),
		zap.Strings("tasks", message.Tasks))

	startTime := time.Now()
	result := &AudioProcessingResult{
		TrackID:   message.TrackID,
		Tasks:     message.Tasks,
		Status:    "completed",
		Results:   make(map[string]interface{}),
		CreatedAt: time.Now(),
	}

	// Traiter chaque tâche
	for _, task := range message.Tasks {
		taskResult, err := j.processTask(message.TrackID, message.AudioPath, task)
		if err != nil {
			j.logger.Error("Task failed",
				zap.String("task", task),
				zap.Int64("track_id", message.TrackID),
				zap.Error(err))
			result.Status = "partial"
		} else {
			result.Results[task] = taskResult
		}
	}

	result.ProcessingTime = time.Since(startTime).Seconds()

	// Sauvegarder le résultat
	err = j.saveProcessingResult(result)
	if err != nil {
		return fmt.Errorf("failed to save processing result: %w", err)
	}

	// Publier le résultat
	err = j.publishResult(result)
	if err != nil {
		return fmt.Errorf("failed to publish result: %w", err)
	}

	j.logger.Info("Audio processing completed",
		zap.Int64("track_id", message.TrackID),
		zap.String("status", result.Status),
		zap.Float64("processing_time", result.ProcessingTime))

	return nil
}

// processTask traite une tâche spécifique
func (j *AudioProcessingJob) processTask(trackID int64, audioPath, task string) (interface{}, error) {
	switch task {
	case "tagging":
		return j.processTagging(trackID, audioPath)
	case "search":
		return j.processSearch(trackID, audioPath)
	case "stems":
		return j.processStems(trackID, audioPath)
	case "mastering":
		return j.processMastering(trackID, audioPath)
	default:
		return nil, fmt.Errorf("unknown task: %s", task)
	}
}

// processTagging traite l'auto-tagging
func (j *AudioProcessingJob) processTagging(trackID int64, audioPath string) (interface{}, error) {
	// Créer le service de tagging
	taggingService := NewTaggingService(j.db, j.cache, j.vaultClient, j.logger, "/tmp/analysis")

	// Appliquer l'auto-tagging
	tags, err := taggingService.AutoTagTrack(trackID, audioPath)
	if err != nil {
		return nil, err
	}

	return tags, nil
}

// processSearch traite l'indexation pour la recherche
func (j *AudioProcessingJob) processSearch(trackID int64, audioPath string) (interface{}, error) {
	// Créer le service de recherche
	searchService := NewSearchService(j.db, j.cache, j.vaultClient, j.logger, "/tmp/features")

	// Extraire et stocker les features
	err := searchService.ExtractAndStoreFeatures(trackID, audioPath)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"features_extracted": true,
		"track_id":           trackID,
	}, nil
}

// processStems traite la séparation de stems
func (j *AudioProcessingJob) processStems(trackID int64, audioPath string) (interface{}, error) {
	// Créer le service de séparation de stems
	stemService := NewStemSeparationService(j.db, j.cache, j.vaultClient, j.logger, "/tmp/stems")

	// Séparer les stems
	result, err := stemService.SeparateStems(trackID, audioPath)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// processMastering traite le mastering
func (j *AudioProcessingJob) processMastering(trackID int64, audioPath string) (interface{}, error) {
	// Créer le service de mastering
	masteringService := NewMasteringService(j.db, j.cache, j.vaultClient, j.logger, "/tmp/mastered")

	// Appliquer le mastering (profil streaming par défaut)
	result, err := masteringService.MasterTrack(trackID, audioPath, "streaming")
	if err != nil {
		return nil, err
	}

	return result, nil
}

// saveProcessingResult sauvegarde le résultat du traitement
func (j *AudioProcessingJob) saveProcessingResult(result *AudioProcessingResult) error {
	// Sauvegarder dans Redis
	ctx := context.Background()
	cacheKey := fmt.Sprintf("audio_processing:result:%d", result.TrackID)

	data, err := json.Marshal(result)
	if err != nil {
		return err
	}

	err = j.cache.Set(ctx, cacheKey, string(data), 24*time.Hour).Err()
	if err != nil {
		return err
	}

	// Mettre à jour la base de données
	query := `
		UPDATE tracks 
		SET 
			audio_processed = true,
			audio_processed_at = NOW(),
			audio_processing_status = $1
		WHERE id = $2
	`

	_, err = j.db.Exec(query, result.Status, result.TrackID)
	return err
}

// publishResult publie le résultat sur Kafka
func (j *AudioProcessingJob) publishResult(result *AudioProcessingResult) error {
	data, err := json.Marshal(result)
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: "audio-processing-results",
		Key:   sarama.StringEncoder(fmt.Sprintf("%d", result.TrackID)),
		Value: sarama.ByteEncoder(data),
	}

	_, _, err = j.producer.SendMessage(msg)
	return err
}

// GetProcessingStatus récupère le statut du traitement
func (j *AudioProcessingJob) GetProcessingStatus(trackID int64) (*AudioProcessingResult, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("audio_processing:result:%d", trackID)

	// Essayer de récupérer depuis le cache
	cached, err := j.cache.Get(ctx, cacheKey).Result()
	if err == nil {
		var result AudioProcessingResult
		err = json.Unmarshal([]byte(cached), &result)
		if err == nil {
			return &result, nil
		}
	}

	// Vérifier dans la base de données
	query := `
		SELECT audio_processed, audio_processing_status, audio_processed_at
		FROM tracks
		WHERE id = $1
	`

	var processed bool
	var status string
	var processedAt *time.Time

	err = j.db.QueryRow(query, trackID).Scan(&processed, &status, &processedAt)
	if err != nil {
		return nil, err
	}

	if !processed {
		return &AudioProcessingResult{
			TrackID: trackID,
			Status:  "pending",
		}, nil
	}

	return &AudioProcessingResult{
		TrackID:   trackID,
		Status:    status,
		CreatedAt: *processedAt,
	}, nil
}

// BatchProcessTracks traite plusieurs tracks en lot
func (j *AudioProcessingJob) BatchProcessTracks(trackIDs []int64, tasks []string) error {
	for _, trackID := range trackIDs {
		// Récupérer le chemin audio
		var audioPath string
		query := `SELECT audio_path FROM tracks WHERE id = $1`
		err := j.db.QueryRow(query, trackID).Scan(&audioPath)
		if err != nil {
			j.logger.Warn("Failed to get track path", zap.Int64("track_id", trackID), zap.Error(err))
			continue
		}

		// Ajouter au traitement
		err = j.EnqueueAudioProcessing(trackID, audioPath, tasks, 1)
		if err != nil {
			j.logger.Warn("Failed to enqueue processing", zap.Int64("track_id", trackID), zap.Error(err))
		}
	}

	return nil
}
