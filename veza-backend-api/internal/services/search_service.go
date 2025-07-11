package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/hashicorp/vault/api"
	"go.uber.org/zap"
)

// AudioFeatures représente les features extraites d'un audio
type AudioFeatures struct {
	TrackID          int64     `json:"track_id"`
	MFCC             []float64 `json:"mfcc"`   // 20 coefficients
	Chroma           []float64 `json:"chroma"` // 12 dimensions
	SpectralCentroid float64   `json:"spectral_centroid"`
	ZeroCrossingRate float64   `json:"zero_crossing_rate"`
	Duration         float64   `json:"duration"`
	ExtractedAt      time.Time `json:"extracted_at"`
}

// SearchResult représente un résultat de recherche
type SearchResult struct {
	TrackID    int64   `json:"track_id"`
	Title      string  `json:"title"`
	Artist     string  `json:"artist"`
	Similarity float64 `json:"similarity"`
	Distance   float64 `json:"distance"`
}

// SearchService fournit la recherche audio algorithmique
type SearchService struct {
	db          *sql.DB
	cache       *redis.Client
	vaultClient *api.Client
	logger      *zap.Logger
	featuresDir string
}

// NewSearchService crée une nouvelle instance du service
func NewSearchService(db *sql.DB, cache *redis.Client, vaultClient *api.Client, logger *zap.Logger, featuresDir string) *SearchService {
	return &SearchService{
		db:          db,
		cache:       cache,
		vaultClient: vaultClient,
		logger:      logger,
		featuresDir: featuresDir,
	}
}

// ExtractAndStoreFeatures extrait et stocke les features d'un audio
func (s *SearchService) ExtractAndStoreFeatures(trackID int64, audioPath string) error {
	// Vérifier si les features existent déjà
	featuresPath := filepath.Join(s.featuresDir, fmt.Sprintf("%d.json", trackID))
	if _, err := os.Stat(featuresPath); err == nil {
		s.logger.Info("Features already exist", zap.Int64("track_id", trackID))
		return nil
	}

	// Extraire les features (simulation - en production, utiliser librosa)
	features, err := s.extractFeatures(audioPath)
	if err != nil {
		return fmt.Errorf("failed to extract features: %w", err)
	}

	features.TrackID = trackID
	features.ExtractedAt = time.Now()

	// Sauvegarder les features
	err = s.saveFeatures(features)
	if err != nil {
		return fmt.Errorf("failed to save features: %w", err)
	}

	// Mettre à jour la base de données
	err = s.updateTrackFeatures(trackID, features)
	if err != nil {
		return fmt.Errorf("failed to update track features: %w", err)
	}

	s.logger.Info("Features extracted and stored", zap.Int64("track_id", trackID))
	return nil
}

// extractFeatures extrait les features audio (simulation)
func (s *SearchService) extractFeatures(audioPath string) (*AudioFeatures, error) {
	// En production, utiliser librosa pour extraire les vraies features
	// Ici, on simule l'extraction

	features := &AudioFeatures{
		MFCC:             make([]float64, 20),
		Chroma:           make([]float64, 12),
		SpectralCentroid: 2000.0 + float64(len(audioPath)%1000), // Simulation
		ZeroCrossingRate: 0.1 + float64(len(audioPath)%10)*0.01, // Simulation
		Duration:         180.0,                                 // 3 minutes par défaut
	}

	// Simuler MFCC (20 coefficients)
	for i := 0; i < 20; i++ {
		features.MFCC[i] = float64(i)*0.1 + float64(len(audioPath)%100)*0.01
	}

	// Simuler Chroma (12 dimensions)
	for i := 0; i < 12; i++ {
		features.Chroma[i] = float64(i)*0.05 + float64(len(audioPath)%50)*0.01
	}

	return features, nil
}

// saveFeatures sauvegarde les features dans un fichier JSON
func (s *SearchService) saveFeatures(features *AudioFeatures) error {
	// Créer le répertoire si nécessaire
	err := os.MkdirAll(s.featuresDir, 0755)
	if err != nil {
		return err
	}

	// Sauvegarder en JSON
	featuresPath := filepath.Join(s.featuresDir, fmt.Sprintf("%d.json", features.TrackID))
	data, err := json.MarshalIndent(features, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(featuresPath, data, 0644)
}

// updateTrackFeatures met à jour les features dans la base de données
func (s *SearchService) updateTrackFeatures(trackID int64, features *AudioFeatures) error {
	query := `
		UPDATE tracks 
		SET 
			features_extracted = true,
			features_updated_at = NOW(),
			audio_duration = $1,
			spectral_centroid = $2,
			zero_crossing_rate = $3
		WHERE id = $4
	`

	_, err := s.db.Exec(query, features.Duration, features.SpectralCentroid, features.ZeroCrossingRate, trackID)
	return err
}

// SearchSimilar recherche les audios similaires
func (s *SearchService) SearchSimilar(queryTrackID int64, limit int) ([]SearchResult, error) {
	ctx := context.Background()

	// Vérifier le cache
	cacheKey := fmt.Sprintf("search:similar:%d:%d", queryTrackID, limit)
	cached, err := s.cache.Get(ctx, cacheKey).Result()
	if err == nil {
		var results []SearchResult
		err = json.Unmarshal([]byte(cached), &results)
		if err == nil {
			s.logger.Info("Search results served from cache", zap.Int64("query_track_id", queryTrackID))
			return results, nil
		}
	}

	// Charger les features de la requête
	queryFeatures, err := s.loadFeatures(queryTrackID)
	if err != nil {
		return nil, fmt.Errorf("failed to load query features: %w", err)
	}

	// Rechercher les tracks similaires
	results, err := s.findSimilarTracks(queryFeatures, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to find similar tracks: %w", err)
	}

	// Mettre en cache pour 30 minutes
	if data, err := json.Marshal(results); err == nil {
		s.cache.Set(ctx, cacheKey, string(data), 30*time.Minute)
	}

	return results, nil
}

// loadFeatures charge les features d'un track
func (s *SearchService) loadFeatures(trackID int64) (*AudioFeatures, error) {
	featuresPath := filepath.Join(s.featuresDir, fmt.Sprintf("%d.json", trackID))

	data, err := os.ReadFile(featuresPath)
	if err != nil {
		return nil, err
	}

	var features AudioFeatures
	err = json.Unmarshal(data, &features)
	if err != nil {
		return nil, err
	}

	return &features, nil
}

// findSimilarTracks trouve les tracks similaires
func (s *SearchService) findSimilarTracks(queryFeatures *AudioFeatures, limit int) ([]SearchResult, error) {
	// Récupérer tous les tracks avec features
	query := `
		SELECT id, title, artist, spectral_centroid, zero_crossing_rate
		FROM tracks
		WHERE features_extracted = true AND id != $1
	`

	rows, err := s.db.Query(query, queryFeatures.TrackID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var trackID int64
		var title, artist string
		var spectralCentroid, zeroCrossingRate float64

		err := rows.Scan(&trackID, &title, &artist, &spectralCentroid, &zeroCrossingRate)
		if err != nil {
			continue
		}

		// Charger les features complètes pour ce track
		trackFeatures, err := s.loadFeatures(trackID)
		if err != nil {
			continue
		}

		// Calculer la similarité
		similarity := s.calculateSimilarity(queryFeatures, trackFeatures)

		results = append(results, SearchResult{
			TrackID:    trackID,
			Title:      title,
			Artist:     artist,
			Similarity: similarity,
			Distance:   1.0 - similarity, // Distance euclidienne normalisée
		})
	}

	// Trier par similarité décroissante
	sort.Slice(results, func(i, j int) bool {
		return results[i].Similarity > results[j].Similarity
	})

	// Limiter les résultats
	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// calculateSimilarity calcule la similarité entre deux sets de features
func (s *SearchService) calculateSimilarity(features1, features2 *AudioFeatures) float64 {
	// Similarité cosinus sur les features MFCC
	mfccSimilarity := s.cosineSimilarity(features1.MFCC, features2.MFCC)

	// Similarité cosinus sur les features Chroma
	chromaSimilarity := s.cosineSimilarity(features1.Chroma, features2.Chroma)

	// Similarité sur les features spectrales
	spectralSimilarity := 1.0 - math.Abs(features1.SpectralCentroid-features2.SpectralCentroid)/10000.0
	zcrSimilarity := 1.0 - math.Abs(features1.ZeroCrossingRate-features2.ZeroCrossingRate)

	// Pondération des similarités
	totalSimilarity := mfccSimilarity*0.4 + chromaSimilarity*0.3 + spectralSimilarity*0.2 + zcrSimilarity*0.1

	return math.Max(0, totalSimilarity)
}

// cosineSimilarity calcule la similarité cosinus entre deux vecteurs
func (s *SearchService) cosineSimilarity(vec1, vec2 []float64) float64 {
	if len(vec1) != len(vec2) || len(vec1) == 0 {
		return 0
	}

	dotProduct := 0.0
	magnitude1 := 0.0
	magnitude2 := 0.0

	for i := 0; i < len(vec1); i++ {
		dotProduct += vec1[i] * vec2[i]
		magnitude1 += vec1[i] * vec1[i]
		magnitude2 += vec2[i] * vec2[i]
	}

	if magnitude1 == 0 || magnitude2 == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(magnitude1) * math.Sqrt(magnitude2))
}

// SearchByTags recherche par tags
func (s *SearchService) SearchByTags(tags []string, limit int) ([]SearchResult, error) {
	query := `
		SELECT DISTINCT t.id, t.title, t.artist
		FROM tracks t
		JOIN track_tags tt ON t.id = tt.track_id
		JOIN tags tag ON tt.tag_id = tag.id
		WHERE tag.name = ANY($1)
		ORDER BY t.created_at DESC
		LIMIT $2
	`

	rows, err := s.db.Query(query, tags, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var trackID int64
		var title, artist string

		err := rows.Scan(&trackID, &title, &artist)
		if err != nil {
			continue
		}

		results = append(results, SearchResult{
			TrackID:    trackID,
			Title:      title,
			Artist:     artist,
			Similarity: 1.0, // Pas de similarité pour la recherche par tags
		})
	}

	return results, nil
}

// BuildSearchIndex construit l'index de recherche
func (s *SearchService) BuildSearchIndex() error {
	s.logger.Info("Building search index...")

	// Récupérer tous les tracks sans features
	query := `
		SELECT id, audio_path
		FROM tracks
		WHERE features_extracted = false AND audio_path IS NOT NULL
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var trackID int64
		var audioPath string

		err := rows.Scan(&trackID, &audioPath)
		if err != nil {
			continue
		}

		// Extraire les features
		err = s.ExtractAndStoreFeatures(trackID, audioPath)
		if err != nil {
			s.logger.Warn("Failed to extract features", zap.Int64("track_id", trackID), zap.Error(err))
			continue
		}

		count++
		if count%100 == 0 {
			s.logger.Info("Indexed tracks", zap.Int("count", count))
		}
	}

	s.logger.Info("Search index built", zap.Int("total_indexed", count))
	return nil
}
