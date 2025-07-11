package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/hashicorp/vault/api"
	"go.uber.org/zap"
)

// AudioAnalysis représente l'analyse audio complète
type AudioAnalysis struct {
	TrackID          int64     `json:"track_id"`
	BPM              float64   `json:"bpm"`
	Key              string    `json:"key"`
	Scale            string    `json:"scale"`
	Energy           float64   `json:"energy"`
	SpectralCentroid float64   `json:"spectral_centroid"`
	RMS              float64   `json:"rms"`
	Duration         float64   `json:"duration"`
	AnalyzedAt       time.Time `json:"analyzed_at"`
}

// TrackTags représente les tags d'un track
type TrackTags struct {
	TrackID     int64     `json:"track_id"`
	Genre       string    `json:"genre"`
	Mood        string    `json:"mood"`
	Instruments []string  `json:"instruments"`
	BPM         float64   `json:"bpm"`
	Key         string    `json:"key"`
	Energy      float64   `json:"energy"`
	Tags        []string  `json:"tags"`
	CreatedAt   time.Time `json:"created_at"`
}

// TaggingService fournit l'auto-tagging algorithmique
type TaggingService struct {
	db          *sql.DB
	cache       *redis.Client
	vaultClient *api.Client
	logger      *zap.Logger
	analysisDir string
}

// Règles de classification
type ClassificationRules struct {
	Genres map[string]GenreRule `json:"genres"`
	Moods  map[string]MoodRule  `json:"moods"`
}

type GenreRule struct {
	BPMRange     [2]float64 `json:"bpm_range"`
	EnergyHigh   bool       `json:"energy_high"`
	EnergyMedium bool       `json:"energy_medium"`
	EnergyLow    bool       `json:"energy_low"`
	Instruments  []string   `json:"instruments"`
}

type MoodRule struct {
	KeyMajor  bool    `json:"key_major"`
	KeyMinor  bool    `json:"key_minor"`
	EnergyMin float64 `json:"energy_min"`
	EnergyMax float64 `json:"energy_max"`
}

// NewTaggingService crée une nouvelle instance du service
func NewTaggingService(db *sql.DB, cache *redis.Client, vaultClient *api.Client, logger *zap.Logger, analysisDir string) *TaggingService {
	return &TaggingService{
		db:          db,
		cache:       cache,
		vaultClient: vaultClient,
		logger:      logger,
		analysisDir: analysisDir,
	}
}

// AutoTagTrack analyse et tag automatiquement un track
func (s *TaggingService) AutoTagTrack(trackID int64, audioPath string) (*TrackTags, error) {
	// Vérifier si l'analyse existe déjà
	analysisPath := filepath.Join(s.analysisDir, fmt.Sprintf("%d.json", trackID))
	if _, err := os.Stat(analysisPath); err == nil {
		return s.loadExistingTags(trackID)
	}

	// Analyser l'audio
	analysis, err := s.analyzeAudio(audioPath)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze audio: %w", err)
	}

	analysis.TrackID = trackID
	analysis.AnalyzedAt = time.Now()

	// Sauvegarder l'analyse
	err = s.saveAnalysis(analysis)
	if err != nil {
		return nil, fmt.Errorf("failed to save analysis: %w", err)
	}

	// Classifier et tagger
	tags, err := s.classifyAndTag(analysis)
	if err != nil {
		return nil, fmt.Errorf("failed to classify and tag: %w", err)
	}

	// Sauvegarder les tags
	err = s.saveTags(tags)
	if err != nil {
		return nil, fmt.Errorf("failed to save tags: %w", err)
	}

	// Mettre à jour la base de données
	err = s.updateTrackTags(trackID, tags)
	if err != nil {
		return nil, fmt.Errorf("failed to update track tags: %w", err)
	}

	s.logger.Info("Track auto-tagged", zap.Int64("track_id", trackID), zap.String("genre", tags.Genre))
	return tags, nil
}

// analyzeAudio analyse l'audio (simulation)
func (s *TaggingService) analyzeAudio(audioPath string) (*AudioAnalysis, error) {
	// En production, utiliser aubio et essentia
	// Ici, on simule l'analyse

	analysis := &AudioAnalysis{
		BPM:              120.0 + float64(len(audioPath)%40), // 120-160 BPM
		Key:              "C",
		Scale:            "major",
		Energy:           0.5 + float64(len(audioPath)%50)*0.01, // 0.5-1.0
		SpectralCentroid: 2000.0 + float64(len(audioPath)%1000),
		RMS:              0.3 + float64(len(audioPath)%70)*0.01,
		Duration:         180.0, // 3 minutes
	}

	// Déterminer la tonalité basée sur le chemin
	if len(audioPath)%2 == 0 {
		analysis.Scale = "minor"
	}

	// Déterminer la clé basée sur le chemin
	keys := []string{"C", "D", "E", "F", "G", "A", "B"}
	analysis.Key = keys[len(audioPath)%len(keys)]

	return analysis, nil
}

// saveAnalysis sauvegarde l'analyse dans un fichier JSON
func (s *TaggingService) saveAnalysis(analysis *AudioAnalysis) error {
	err := os.MkdirAll(s.analysisDir, 0755)
	if err != nil {
		return err
	}

	analysisPath := filepath.Join(s.analysisDir, fmt.Sprintf("%d.json", analysis.TrackID))
	data, err := json.MarshalIndent(analysis, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(analysisPath, data, 0644)
}

// classifyAndTag classe et tag le track
func (s *TaggingService) classifyAndTag(analysis *AudioAnalysis) (*TrackTags, error) {
	tags := &TrackTags{
		TrackID:   analysis.TrackID,
		BPM:       analysis.BPM,
		Key:       fmt.Sprintf("%s %s", analysis.Key, analysis.Scale),
		Energy:    analysis.Energy,
		CreatedAt: time.Now(),
	}

	// Classifier le genre
	tags.Genre = s.classifyGenre(analysis)

	// Classifier le mood
	tags.Mood = s.classifyMood(analysis)

	// Détecter les instruments
	tags.Instruments = s.detectInstruments(analysis)

	// Générer les tags combinés
	tags.Tags = s.generateTags(tags)

	return tags, nil
}

// classifyGenre classe le genre selon les règles heuristiques
func (s *TaggingService) classifyGenre(analysis *AudioAnalysis) string {
	rules := s.getGenreRules()

	for genre, rule := range rules {
		// Vérifier le BPM
		if analysis.BPM >= rule.BPMRange[0] && analysis.BPM <= rule.BPMRange[1] {
			// Vérifier l'énergie
			if rule.EnergyHigh && analysis.Energy > 0.7 {
				return genre
			} else if rule.EnergyMedium && analysis.Energy >= 0.3 && analysis.Energy <= 0.7 {
				return genre
			} else if rule.EnergyLow && analysis.Energy < 0.3 {
				return genre
			}
		}
	}

	return "unknown"
}

// classifyMood classe le mood selon les règles heuristiques
func (s *TaggingService) classifyMood(analysis *AudioAnalysis) string {
	if analysis.Scale == "major" {
		if analysis.Energy > 0.7 {
			return "energetic"
		} else if analysis.Energy > 0.4 {
			return "happy"
		} else {
			return "calm"
		}
	} else { // minor
		if analysis.Energy > 0.7 {
			return "dark"
		} else if analysis.Energy > 0.4 {
			return "melancholic"
		} else {
			return "sad"
		}
	}
}

// detectInstruments détecte les instruments (simulation)
func (s *TaggingService) detectInstruments(analysis *AudioAnalysis) []string {
	var instruments []string

	// Règles heuristiques basées sur le spectre
	if analysis.SpectralCentroid > 2500 {
		instruments = append(instruments, "synth")
	}

	if analysis.Energy > 0.6 {
		instruments = append(instruments, "drums")
	}

	if analysis.SpectralCentroid < 1500 && analysis.Energy > 0.4 {
		instruments = append(instruments, "bass")
	}

	if analysis.SpectralCentroid > 3000 && analysis.Energy < 0.5 {
		instruments = append(instruments, "piano")
	}

	// Ajouter des instruments basés sur la tonalité
	if analysis.Scale == "major" {
		instruments = append(instruments, "strings")
	} else {
		instruments = append(instruments, "pad")
	}

	return instruments
}

// generateTags génère les tags combinés
func (s *TaggingService) generateTags(tags *TrackTags) []string {
	var allTags []string

	// Ajouter le genre
	if tags.Genre != "unknown" {
		allTags = append(allTags, tags.Genre)
	}

	// Ajouter le mood
	allTags = append(allTags, tags.Mood)

	// Ajouter les instruments
	allTags = append(allTags, tags.Instruments...)

	// Ajouter des tags basés sur le BPM
	if tags.BPM > 140 {
		allTags = append(allTags, "fast")
	} else if tags.BPM < 80 {
		allTags = append(allTags, "slow")
	} else {
		allTags = append(allTags, "medium")
	}

	// Ajouter des tags basés sur l'énergie
	if tags.Energy > 0.7 {
		allTags = append(allTags, "high-energy")
	} else if tags.Energy < 0.3 {
		allTags = append(allTags, "low-energy")
	} else {
		allTags = append(allTags, "medium-energy")
	}

	// Ajouter la tonalité
	allTags = append(allTags, strings.ToLower(tags.Key))

	return allTags
}

// getGenreRules retourne les règles de classification des genres
func (s *TaggingService) getGenreRules() map[string]GenreRule {
	return map[string]GenreRule{
		"electronic": {
			BPMRange:    [2]float64{120, 140},
			EnergyHigh:  true,
			Instruments: []string{"synth", "drums"},
		},
		"rock": {
			BPMRange:    [2]float64{80, 120},
			EnergyHigh:  true,
			Instruments: []string{"drums", "guitar"},
		},
		"jazz": {
			BPMRange:     [2]float64{60, 120},
			EnergyMedium: true,
			Instruments:  []string{"piano", "bass", "drums"},
		},
		"classical": {
			BPMRange:    [2]float64{40, 180},
			EnergyLow:   true,
			Instruments: []string{"strings", "piano"},
		},
		"hip-hop": {
			BPMRange:     [2]float64{80, 100},
			EnergyMedium: true,
			Instruments:  []string{"drums", "bass"},
		},
		"ambient": {
			BPMRange:    [2]float64{60, 90},
			EnergyLow:   true,
			Instruments: []string{"pad", "synth"},
		},
	}
}

// saveTags sauvegarde les tags dans un fichier JSON
func (s *TaggingService) saveTags(tags *TrackTags) error {
	err := os.MkdirAll(s.analysisDir, 0755)
	if err != nil {
		return err
	}

	tagsPath := filepath.Join(s.analysisDir, fmt.Sprintf("%d_tags.json", tags.TrackID))
	data, err := json.MarshalIndent(tags, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(tagsPath, data, 0644)
}

// loadExistingTags charge les tags existants
func (s *TaggingService) loadExistingTags(trackID int64) (*TrackTags, error) {
	tagsPath := filepath.Join(s.analysisDir, fmt.Sprintf("%d_tags.json", trackID))

	data, err := os.ReadFile(tagsPath)
	if err != nil {
		return nil, err
	}

	var tags TrackTags
	err = json.Unmarshal(data, &tags)
	if err != nil {
		return nil, err
	}

	return &tags, nil
}

// updateTrackTags met à jour les tags dans la base de données
func (s *TaggingService) updateTrackTags(trackID int64, tags *TrackTags) error {
	// Mettre à jour la table tracks
	query := `
		UPDATE tracks 
		SET 
			tags_extracted = true,
			tags_updated_at = NOW(),
			genre = $1,
			bpm = $2,
			key_signature = $3,
			energy_level = $4
		WHERE id = $5
	`

	_, err := s.db.Exec(query, tags.Genre, tags.BPM, tags.Key, tags.Energy, trackID)
	if err != nil {
		return err
	}

	// Insérer les tags dans la table track_tags
	for _, tagName := range tags.Tags {
		err = s.insertTrackTag(trackID, tagName)
		if err != nil {
			s.logger.Warn("Failed to insert tag", zap.String("tag", tagName), zap.Error(err))
		}
	}

	return nil
}

// insertTrackTag insère un tag pour un track
func (s *TaggingService) insertTrackTag(trackID int64, tagName string) error {
	// Vérifier si le tag existe, sinon le créer
	var tagID int64
	query := `SELECT id FROM tags WHERE name = $1`
	err := s.db.QueryRow(query, tagName).Scan(&tagID)
	if err == sql.ErrNoRows {
		// Créer le tag
		query = `INSERT INTO tags (name, created_at) VALUES ($1, NOW()) RETURNING id`
		err = s.db.QueryRow(query, tagName).Scan(&tagID)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	// Insérer l'association track-tag
	query = `
		INSERT INTO track_tags (track_id, tag_id, created_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (track_id, tag_id) DO NOTHING
	`

	_, err = s.db.Exec(query, trackID, tagID)
	return err
}

// GetTrackTags récupère les tags d'un track
func (s *TaggingService) GetTrackTags(trackID int64) (*TrackTags, error) {
	// Essayer de charger depuis le cache
	ctx := context.Background()
	cacheKey := fmt.Sprintf("tags:%d", trackID)
	cached, err := s.cache.Get(ctx, cacheKey).Result()
	if err == nil {
		var tags TrackTags
		err = json.Unmarshal([]byte(cached), &tags)
		if err == nil {
			return &tags, nil
		}
	}

	// Charger depuis le fichier
	tags, err := s.loadExistingTags(trackID)
	if err != nil {
		return nil, err
	}

	// Mettre en cache
	if data, err := json.Marshal(tags); err == nil {
		s.cache.Set(ctx, cacheKey, string(data), time.Hour)
	}

	return tags, nil
}

// UpdateUserTag permet à l'utilisateur de modifier un tag
func (s *TaggingService) UpdateUserTag(trackID int64, tagName string, action string) error {
	// action: "add" ou "remove"
	if action == "add" {
		return s.insertTrackTag(trackID, tagName)
	} else if action == "remove" {
		query := `
			DELETE FROM track_tags 
			WHERE track_id = $1 AND tag_id = (SELECT id FROM tags WHERE name = $2)
		`
		_, err := s.db.Exec(query, trackID, tagName)
		return err
	}

	return fmt.Errorf("invalid action: %s", action)
}

// GetPopularTags récupère les tags les plus populaires
func (s *TaggingService) GetPopularTags(limit int) ([]string, error) {
	query := `
		SELECT t.name, COUNT(tt.track_id) as usage_count
		FROM tags t
		JOIN track_tags tt ON t.id = tt.tag_id
		GROUP BY t.id, t.name
		ORDER BY usage_count DESC
		LIMIT $1
	`

	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tagName string
		var usageCount int
		err := rows.Scan(&tagName, &usageCount)
		if err != nil {
			continue
		}
		tags = append(tags, tagName)
	}

	return tags, nil
}
