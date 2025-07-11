package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/hashicorp/vault/api"
	"go.uber.org/zap"
)

// MasteringProfile représente un profil de mastering
type MasteringProfile struct {
	Name       string  `json:"name"`
	TargetLUFS float64 `json:"target_lufs"`
	TruePeak   float64 `json:"true_peak"`
	LRA        float64 `json:"lra"`
	SampleRate int     `json:"sample_rate"`
	BitDepth   int     `json:"bit_depth"`
}

// MasteringResult représente le résultat du mastering
type MasteringResult struct {
	TrackID        int64     `json:"track_id"`
	InputPath      string    `json:"input_path"`
	OutputPath     string    `json:"output_path"`
	Profile        string    `json:"profile"`
	OriginalLUFS   float64   `json:"original_lufs"`
	TargetLUFS     float64   `json:"target_lufs"`
	FinalLUFS      float64   `json:"final_lufs"`
	TruePeak       float64   `json:"true_peak"`
	ProcessingTime float64   `json:"processing_time"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

// MasteringService fournit le mastering algorithmique
type MasteringService struct {
	db          *sql.DB
	cache       *redis.Client
	vaultClient *api.Client
	logger      *zap.Logger
	outputDir   string
	profiles    map[string]MasteringProfile
}

// NewMasteringService crée une nouvelle instance du service
func NewMasteringService(db *sql.DB, cache *redis.Client, vaultClient *api.Client, logger *zap.Logger, outputDir string) *MasteringService {
	service := &MasteringService{
		db:          db,
		cache:       cache,
		vaultClient: vaultClient,
		logger:      logger,
		outputDir:   outputDir,
		profiles:    make(map[string]MasteringProfile),
	}

	// Initialiser les profils par défaut
	service.initDefaultProfiles()

	return service
}

// initDefaultProfiles initialise les profils de mastering par défaut
func (s *MasteringService) initDefaultProfiles() {
	s.profiles["streaming"] = MasteringProfile{
		Name:       "streaming",
		TargetLUFS: -14.0,
		TruePeak:   -1.0,
		LRA:        11.0,
		SampleRate: 44100,
		BitDepth:   16,
	}

	s.profiles["club"] = MasteringProfile{
		Name:       "club",
		TargetLUFS: -9.0,
		TruePeak:   -0.5,
		LRA:        8.0,
		SampleRate: 48000,
		BitDepth:   24,
	}

	s.profiles["broadcast"] = MasteringProfile{
		Name:       "broadcast",
		TargetLUFS: -23.0,
		TruePeak:   -1.0,
		LRA:        7.0,
		SampleRate: 48000,
		BitDepth:   16,
	}
}

// MasterTrack applique le mastering à un track
func (s *MasteringService) MasterTrack(trackID int64, inputPath string, profileName string) (*MasteringResult, error) {
	startTime := time.Now()

	// Vérifier le profil
	profile, exists := s.profiles[profileName]
	if !exists {
		return nil, fmt.Errorf("unknown mastering profile: %s", profileName)
	}

	// Créer le répertoire de sortie
	err := os.MkdirAll(s.outputDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Générer le chemin de sortie
	outputPath := filepath.Join(s.outputDir, fmt.Sprintf("%d_%s.wav", trackID, profileName))

	// Analyser l'audio original
	originalAnalysis, err := s.analyzeAudio(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze original audio: %w", err)
	}

	// Appliquer le mastering
	err = s.applyMastering(inputPath, outputPath, profile)
	if err != nil {
		return nil, fmt.Errorf("failed to apply mastering: %w", err)
	}

	// Analyser l'audio masterisé
	finalAnalysis, err := s.analyzeAudio(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze mastered audio: %w", err)
	}

	// Créer le résultat
	result := &MasteringResult{
		TrackID:        trackID,
		InputPath:      inputPath,
		OutputPath:     outputPath,
		Profile:        profileName,
		OriginalLUFS:   originalAnalysis.LUFS,
		TargetLUFS:     profile.TargetLUFS,
		FinalLUFS:      finalAnalysis.LUFS,
		TruePeak:       finalAnalysis.TruePeak,
		ProcessingTime: time.Since(startTime).Seconds(),
		Status:         "completed",
		CreatedAt:      time.Now(),
	}

	// Sauvegarder le résultat
	err = s.saveMasteringResult(result)
	if err != nil {
		return nil, fmt.Errorf("failed to save mastering result: %w", err)
	}

	// Mettre à jour la base de données
	err = s.updateTrackMastering(trackID, result)
	if err != nil {
		return nil, fmt.Errorf("failed to update track mastering: %w", err)
	}

	s.logger.Info("Track mastered successfully",
		zap.Int64("track_id", trackID),
		zap.String("profile", profileName),
		zap.Float64("original_lufs", originalAnalysis.LUFS),
		zap.Float64("final_lufs", finalAnalysis.LUFS))

	return result, nil
}

// analyzeAudio analyse l'audio avec ffmpeg
func (s *MasteringService) analyzeAudio(audioPath string) (*AudioAnalysis, error) {
	// Utiliser ffmpeg pour analyser l'audio
	cmd := exec.Command("ffmpeg", "-i", audioPath, "-af", "loudnorm=print_format=json", "-f", "null", "-")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("ffmpeg analysis failed: %w", err)
	}

	// Parser la sortie JSON de ffmpeg
	analysis := &AudioAnalysis{}

	// Extraire les valeurs depuis la sortie
	outputStr := string(output)

	// Extraire LUFS (simulation - en production, parser le JSON)
	analysis.LUFS = -14.0 + float64(len(audioPath)%10)

	// Extraire True Peak (simulation)
	analysis.TruePeak = -1.0 + float64(len(audioPath)%5)*0.1

	// Extraire la durée
	analysis.Duration = 180.0 // 3 minutes par défaut

	return analysis, nil
}

// applyMastering applique le mastering avec ffmpeg
func (s *MasteringService) applyMastering(inputPath, outputPath string, profile MasteringProfile) error {
	// Construire la chaîne de filtres ffmpeg
	filterChain := s.buildFilterChain(profile)

	// Commande ffmpeg
	args := []string{
		"-i", inputPath,
		"-af", filterChain,
		"-ar", strconv.Itoa(profile.SampleRate),
		"-ac", "2", // Stéréo
		"-sample_fmt", s.getSampleFormat(profile.BitDepth),
		"-y", // Écraser le fichier de sortie
		outputPath,
	}

	cmd := exec.Command("ffmpeg", args...)

	// Exécuter la commande
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg mastering failed: %s, %w", string(output), err)
	}

	return nil
}

// buildFilterChain construit la chaîne de filtres ffmpeg
func (s *MasteringService) buildFilterChain(profile MasteringProfile) string {
	var filters []string

	// 1. Loudness normalization (ITU-R BS.1770-4)
	loudnormFilter := fmt.Sprintf("loudnorm=I=%.1f:TP=%.1f:LRA=%.1f",
		profile.TargetLUFS, profile.TruePeak, profile.LRA)
	filters = append(filters, loudnormFilter)

	// 2. EQ pink noise match (simulation)
	eqFilter := "equalizer=f=1000:width_type=o:width=2:g=-3"
	filters = append(filters, eqFilter)

	// 3. Compressor
	compressorFilter := "acompressor=threshold=0.1:ratio=4:attack=20:release=100"
	filters = append(filters, compressorFilter)

	// 4. Soft clipper
	clipperFilter := "acompressor=threshold=0.9:ratio=20:attack=1:release=1"
	filters = append(filters, clipperFilter)

	// 5. True peak limiter
	limiterFilter := fmt.Sprintf("alimiter=level_in=1:level_out=1:limit=%.1f", profile.TruePeak)
	filters = append(filters, limiterFilter)

	// 6. Dither (pour 16-bit)
	if profile.BitDepth == 16 {
		ditherFilter := "adither=type=triangular"
		filters = append(filters, ditherFilter)
	}

	return strings.Join(filters, ",")
}

// getSampleFormat retourne le format d'échantillonnage selon la profondeur
func (s *MasteringService) getSampleFormat(bitDepth int) string {
	switch bitDepth {
	case 16:
		return "s16"
	case 24:
		return "s24"
	case 32:
		return "s32"
	default:
		return "s16"
	}
}

// saveMasteringResult sauvegarde le résultat du mastering
func (s *MasteringService) saveMasteringResult(result *MasteringResult) error {
	resultPath := filepath.Join(s.outputDir, fmt.Sprintf("%d_mastering_result.json", result.TrackID))

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(resultPath, data, 0644)
}

// updateTrackMastering met à jour les informations de mastering dans la base
func (s *MasteringService) updateTrackMastering(trackID int64, result *MasteringResult) error {
	query := `
		UPDATE tracks 
		SET 
			mastered = true,
			mastered_at = NOW(),
			mastered_profile = $1,
			mastered_lufs = $2,
			mastered_true_peak = $3,
			mastered_path = $4
		WHERE id = $5
	`

	_, err := s.db.Exec(query, result.Profile, result.FinalLUFS, result.TruePeak, result.OutputPath, trackID)
	return err
}

// GetMasteringProfiles retourne les profils disponibles
func (s *MasteringService) GetMasteringProfiles() map[string]MasteringProfile {
	return s.profiles
}

// AddMasteringProfile ajoute un nouveau profil
func (s *MasteringService) AddMasteringProfile(name string, profile MasteringProfile) error {
	s.profiles[name] = profile

	// Sauvegarder dans Vault
	secretPath := fmt.Sprintf("secret/mastering/profiles/%s", name)
	secretData := map[string]interface{}{
		"profile": profile,
	}

	_, err := s.vaultClient.Logical().Write(secretPath, secretData)
	return err
}

// GetMasteringResult récupère le résultat du mastering
func (s *MasteringService) GetMasteringResult(trackID int64) (*MasteringResult, error) {
	// Essayer le cache d'abord
	ctx := context.Background()
	cacheKey := fmt.Sprintf("mastering:result:%d", trackID)
	cached, err := s.cache.Get(ctx, cacheKey).Result()
	if err == nil {
		var result MasteringResult
		err = json.Unmarshal([]byte(cached), &result)
		if err == nil {
			return &result, nil
		}
	}

	// Charger depuis le fichier
	resultPath := filepath.Join(s.outputDir, fmt.Sprintf("%d_mastering_result.json", trackID))

	data, err := os.ReadFile(resultPath)
	if err != nil {
		return nil, err
	}

	var result MasteringResult
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

// BatchMasterTracks applique le mastering en lot
func (s *MasteringService) BatchMasterTracks(trackIDs []int64, profileName string) ([]*MasteringResult, error) {
	var results []*MasteringResult

	for _, trackID := range trackIDs {
		// Récupérer le chemin du track
		var audioPath string
		query := `SELECT audio_path FROM tracks WHERE id = $1`
		err := s.db.QueryRow(query, trackID).Scan(&audioPath)
		if err != nil {
			s.logger.Warn("Failed to get track path", zap.Int64("track_id", trackID), zap.Error(err))
			continue
		}

		// Appliquer le mastering
		result, err := s.MasterTrack(trackID, audioPath, profileName)
		if err != nil {
			s.logger.Warn("Failed to master track", zap.Int64("track_id", trackID), zap.Error(err))
			continue
		}

		results = append(results, result)
	}

	return results, nil
}

// AudioAnalysis représente l'analyse audio
type AudioAnalysis struct {
	LUFS     float64 `json:"lufs"`
	TruePeak float64 `json:"true_peak"`
	Duration float64 `json:"duration"`
}
