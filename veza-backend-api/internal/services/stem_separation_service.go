package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/hashicorp/vault/api"
	"go.uber.org/zap"
)

// StemSeparationResult représente le résultat de la séparation
type StemSeparationResult struct {
	TrackID        int64     `json:"track_id"`
	InputPath      string    `json:"input_path"`
	OutputDir      string    `json:"output_dir"`
	VocalsPath     string    `json:"vocals_path"`
	DrumsPath      string    `json:"drums_path"`
	BassPath       string    `json:"bass_path"`
	OtherPath      string    `json:"other_path"`
	ProcessingTime float64   `json:"processing_time"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

// StemSeparationService fournit la séparation de stems algorithmique
type StemSeparationService struct {
	db          *sql.DB
	cache       *redis.Client
	vaultClient *api.Client
	logger      *zap.Logger
	outputDir   string
}

// NewStemSeparationService crée une nouvelle instance du service
func NewStemSeparationService(db *sql.DB, cache *redis.Client, vaultClient *api.Client, logger *zap.Logger, outputDir string) *StemSeparationService {
	return &StemSeparationService{
		db:          db,
		cache:       cache,
		vaultClient: vaultClient,
		logger:      logger,
		outputDir:   outputDir,
	}
}

// SeparateStems sépare les stems d'un track
func (s *StemSeparationService) SeparateStems(trackID int64, inputPath string) (*StemSeparationResult, error) {
	startTime := time.Now()

	// Créer le répertoire de sortie
	outputDir := filepath.Join(s.outputDir, fmt.Sprintf("%d", trackID))
	err := os.MkdirAll(outputDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Générer les chemins de sortie
	vocalsPath := filepath.Join(outputDir, "vocals.wav")
	drumsPath := filepath.Join(outputDir, "drums.wav")
	bassPath := filepath.Join(outputDir, "bass.wav")
	otherPath := filepath.Join(outputDir, "other.wav")

	// Appliquer la séparation HPSS (Harmonic-Percussive)
	err = s.applyHPSS(inputPath, vocalsPath, drumsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to apply HPSS: %w", err)
	}

	// Appliquer la séparation NMF (Non-negative Matrix Factorization)
	err = s.applyNMF(inputPath, bassPath, otherPath)
	if err != nil {
		return nil, fmt.Errorf("failed to apply NMF: %w", err)
	}

	// Créer le résultat
	result := &StemSeparationResult{
		TrackID:        trackID,
		InputPath:      inputPath,
		OutputDir:      outputDir,
		VocalsPath:     vocalsPath,
		DrumsPath:      drumsPath,
		BassPath:       bassPath,
		OtherPath:      otherPath,
		ProcessingTime: time.Since(startTime).Seconds(),
		Status:         "completed",
		CreatedAt:      time.Now(),
	}

	// Sauvegarder le résultat
	err = s.saveSeparationResult(result)
	if err != nil {
		return nil, fmt.Errorf("failed to save separation result: %w", err)
	}

	// Mettre à jour la base de données
	err = s.updateTrackStems(trackID, result)
	if err != nil {
		return nil, fmt.Errorf("failed to update track stems: %w", err)
	}

	s.logger.Info("Stems separated successfully",
		zap.Int64("track_id", trackID),
		zap.Float64("processing_time", result.ProcessingTime))

	return result, nil
}

// applyHPSS applique la séparation Harmonic-Percussive
func (s *StemSeparationService) applyHPSS(inputPath, vocalsPath, drumsPath string) error {
	// Utiliser sox pour la séparation HPSS (simulation)
	// En production, utiliser librosa ou un outil spécialisé

	// Séparation des fréquences hautes (vocals) et basses (drums)
	// Vocals: filtrage passe-haut
	vocalsCmd := exec.Command("sox", inputPath, vocalsPath,
		"highpass", "200", "lowpass", "8000")

	err := vocalsCmd.Run()
	if err != nil {
		return fmt.Errorf("vocals separation failed: %w", err)
	}

	// Drums: filtrage passe-bas
	drumsCmd := exec.Command("sox", inputPath, drumsPath,
		"lowpass", "200")

	err = drumsCmd.Run()
	if err != nil {
		return fmt.Errorf("drums separation failed: %w", err)
	}

	return nil
}

// applyNMF applique la séparation NMF
func (s *StemSeparationService) applyNMF(inputPath, bassPath, otherPath string) error {
	// Utiliser un script Python pour NMF (simulation)
	// En production, utiliser librosa.decompose.NMF

	// Script Python pour NMF
	pythonScript := `
import librosa
import numpy as np
from sklearn.decomposition import NMF
import soundfile as sf

# Charger l'audio
y, sr = librosa.load('` + inputPath + `', sr=44100)

# Extraire les features
stft = librosa.stft(y)
magnitude = np.abs(stft)

# NMF decomposition (2 composantes: bass et other)
nmf = NMF(n_components=2, random_state=42)
W = nmf.fit_transform(magnitude.T)
H = nmf.components_

# Reconstruire les stems
bass_magnitude = np.dot(W[:, 0:1], H[0:1, :]).T
other_magnitude = np.dot(W[:, 1:2], H[1:2, :]).T

# Reconstruire l'audio
bass_stft = bass_magnitude * np.exp(1j * np.angle(stft))
other_stft = other_magnitude * np.exp(1j * np.angle(stft))

bass_audio = librosa.istft(bass_stft)
other_audio = librosa.istft(other_stft)

# Sauvegarder
sf.write('` + bassPath + `', bass_audio, sr)
sf.write('` + otherPath + `', other_audio, sr)
`

	// Écrire le script temporaire
	scriptPath := filepath.Join(os.TempDir(), "nmf_separation.py")
	err := os.WriteFile(scriptPath, []byte(pythonScript), 0644)
	if err != nil {
		return fmt.Errorf("failed to write Python script: %w", err)
	}
	defer os.Remove(scriptPath)

	// Exécuter le script Python
	cmd := exec.Command("python3", scriptPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Python NMF script failed: %s, %w", string(output), err)
	}

	return nil
}

// saveSeparationResult sauvegarde le résultat de la séparation
func (s *StemSeparationService) saveSeparationResult(result *StemSeparationResult) error {
	resultPath := filepath.Join(result.OutputDir, "separation_result.json")

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(resultPath, data, 0644)
}

// updateTrackStems met à jour les informations de stems dans la base
func (s *StemSeparationService) updateTrackStems(trackID int64, result *StemSeparationResult) error {
	query := `
		UPDATE tracks 
		SET 
			stems_separated = true,
			stems_separated_at = NOW(),
			stems_output_dir = $1,
			stems_vocals_path = $2,
			stems_drums_path = $3,
			stems_bass_path = $4,
			stems_other_path = $5
		WHERE id = $6
	`

	_, err := s.db.Exec(query,
		result.OutputDir,
		result.VocalsPath,
		result.DrumsPath,
		result.BassPath,
		result.OtherPath,
		trackID)

	return err
}

// GetStemSeparationResult récupère le résultat de la séparation
func (s *StemSeparationService) GetStemSeparationResult(trackID int64) (*StemSeparationResult, error) {
	// Essayer le cache d'abord
	ctx := context.Background()
	cacheKey := fmt.Sprintf("stems:result:%d", trackID)
	cached, err := s.cache.Get(ctx, cacheKey).Result()
	if err == nil {
		var result StemSeparationResult
		err = json.Unmarshal([]byte(cached), &result)
		if err == nil {
			return &result, nil
		}
	}

	// Charger depuis le fichier
	resultPath := filepath.Join(s.outputDir, fmt.Sprintf("%d/separation_result.json", trackID))

	data, err := os.ReadFile(resultPath)
	if err != nil {
		return nil, err
	}

	var result StemSeparationResult
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}

	// Mettre en cache
	if data, err := json.Marshal(result); err == nil {
		s.cache.Set(ctx, cacheKey, string(data), time.Hour)
	}

	return &result, nil
}

// ExtractCenterChannel extrait le canal central (karaoké)
func (s *StemSeparationService) ExtractCenterChannel(trackID int64, inputPath string) (string, error) {
	outputPath := filepath.Join(s.outputDir, fmt.Sprintf("%d_center.wav", trackID))

	// Créer le répertoire si nécessaire
	err := os.MkdirAll(filepath.Dir(outputPath), 0755)
	if err != nil {
		return "", err
	}

	// Extraire le canal central avec sox
	// Center = L - R (pour les fichiers stéréo)
	cmd := exec.Command("sox", inputPath, outputPath, "channels", "1")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("center channel extraction failed: %s, %w", string(output), err)
	}

	// Mettre à jour la base de données
	query := `UPDATE tracks SET center_channel_path = $1 WHERE id = $2`
	_, err = s.db.Exec(query, outputPath, trackID)
	if err != nil {
		s.logger.Warn("Failed to update center channel path", zap.Error(err))
	}

	return outputPath, nil
}

// BatchSeparateStems applique la séparation en lot
func (s *StemSeparationService) BatchSeparateStems(trackIDs []int64) ([]*StemSeparationResult, error) {
	var results []*StemSeparationResult

	for _, trackID := range trackIDs {
		// Récupérer le chemin du track
		var audioPath string
		query := `SELECT audio_path FROM tracks WHERE id = $1`
		err := s.db.QueryRow(query, trackID).Scan(&audioPath)
		if err != nil {
			s.logger.Warn("Failed to get track path", zap.Int64("track_id", trackID), zap.Error(err))
			continue
		}

		// Appliquer la séparation
		result, err := s.SeparateStems(trackID, audioPath)
		if err != nil {
			s.logger.Warn("Failed to separate stems", zap.Int64("track_id", trackID), zap.Error(err))
			continue
		}

		results = append(results, result)
	}

	return results, nil
}

// GetStemStats récupère les statistiques des stems
func (s *StemSeparationService) GetStemStats(trackID int64) (map[string]interface{}, error) {
	result, err := s.GetStemSeparationResult(trackID)
	if err != nil {
		return nil, err
	}

	stats := make(map[string]interface{})

	// Vérifier l'existence des fichiers
	stats["vocals_exists"] = fileExists(result.VocalsPath)
	stats["drums_exists"] = fileExists(result.DrumsPath)
	stats["bass_exists"] = fileExists(result.BassPath)
	stats["other_exists"] = fileExists(result.OtherPath)

	// Obtenir les tailles des fichiers
	if stats["vocals_exists"].(bool) {
		stats["vocals_size"] = getFileSize(result.VocalsPath)
	}
	if stats["drums_exists"].(bool) {
		stats["drums_size"] = getFileSize(result.DrumsPath)
	}
	if stats["bass_exists"].(bool) {
		stats["bass_size"] = getFileSize(result.BassPath)
	}
	if stats["other_exists"].(bool) {
		stats["other_size"] = getFileSize(result.OtherPath)
	}

	stats["processing_time"] = result.ProcessingTime
	stats["status"] = result.Status

	return stats, nil
}

// fileExists vérifie si un fichier existe
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// getFileSize retourne la taille d'un fichier
func getFileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}
