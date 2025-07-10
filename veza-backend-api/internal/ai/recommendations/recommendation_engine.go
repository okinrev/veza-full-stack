package recommendations

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"go.uber.org/zap"
)

// RecommendationEngine moteur de recommandations AI
type RecommendationEngine struct {
	userProfileService    UserProfileService
	trackAnalyticsService TrackAnalyticsService
	mlModel               MLModel
	logger                *zap.Logger
	cache                 RecommendationCache
}

// UserProfileService interface pour le service de profil utilisateur
type UserProfileService interface {
	GetUserProfile(ctx context.Context, userID int64) (*UserProfile, error)
	UpdateUserProfile(ctx context.Context, userID int64, profile *UserProfile) error
}

// TrackAnalyticsService interface pour le service d'analytics des pistes
type TrackAnalyticsService interface {
	GetTrackAnalytics(ctx context.Context, trackID int64) (*TrackAnalytics, error)
	GetSimilarTracks(ctx context.Context, trackID int64, limit int) ([]int64, error)
	GetPopularTracks(ctx context.Context, genre string, limit int) ([]int64, error)
}

// MLModel interface pour le modèle ML
type MLModel interface {
	PredictUserPreferences(ctx context.Context, userProfile *UserProfile) (*UserPreferences, error)
	GetTrackEmbeddings(ctx context.Context, trackIDs []int64) (map[int64][]float64, error)
	CalculateSimilarity(embedding1, embedding2 []float64) float64
}

// RecommendationCache interface pour le cache des recommandations
type RecommendationCache interface {
	GetRecommendations(ctx context.Context, userID int64, context string) (*RecommendationResponse, error)
	SetRecommendations(ctx context.Context, userID int64, context string, recommendations *RecommendationResponse, ttl time.Duration) error
	InvalidateUserRecommendations(ctx context.Context, userID int64) error
}

// UserProfile profil utilisateur pour les recommandations
type UserProfile struct {
	UserID             int64              `json:"user_id"`
	Genres             map[string]float64 `json:"genres"`              // Genre -> Score (0-1)
	Artists            map[string]float64 `json:"artists"`             // Artist -> Score (0-1)
	Instruments        map[string]float64 `json:"instruments"`         // Instrument -> Score (0-1)
	MoodPreferences    map[string]float64 `json:"mood_preferences"`    // Mood -> Score (0-1)
	CollaborationStyle string             `json:"collaboration_style"` // "solo", "collaborative", "both"
	SkillLevel         string             `json:"skill_level"`         // "beginner", "intermediate", "advanced"
	ProductionStyle    string             `json:"production_style"`    // "electronic", "acoustic", "mixed"
	RecentActivity     []UserActivity     `json:"recent_activity"`
	LastUpdated        time.Time          `json:"last_updated"`
	Embedding          []float64          `json:"embedding"` // Vector embedding du profil
}

// UserActivity activité récente de l'utilisateur
type UserActivity struct {
	Type        string    `json:"type"` // "listen", "like", "share", "collaborate", "purchase"
	TrackID     int64     `json:"track_id"`
	ArtistID    int64     `json:"artist_id"`
	Genre       string    `json:"genre"`
	Timestamp   time.Time `json:"timestamp"`
	Duration    int       `json:"duration"`    // en secondes
	Interaction float64   `json:"interaction"` // score d'interaction (0-1)
}

// UserPreferences préférences prédites par le ML
type UserPreferences struct {
	GenreWeights      map[string]float64 `json:"genre_weights"`
	ArtistWeights     map[string]float64 `json:"artist_weights"`
	MoodWeights       map[string]float64 `json:"mood_weights"`
	CollaborationPref float64            `json:"collaboration_pref"`
	SkillLevelPref    string             `json:"skill_level_pref"`
	ProductionPref    string             `json:"production_pref"`
	Confidence        float64            `json:"confidence"`
}

// TrackAnalytics analytics d'une piste
type TrackAnalytics struct {
	TrackID            int64     `json:"track_id"`
	PlayCount          int64     `json:"play_count"`
	LikeCount          int64     `json:"like_count"`
	ShareCount         int64     `json:"share_count"`
	CollaborationCount int64     `json:"collaboration_count"`
	Genre              string    `json:"genre"`
	Mood               string    `json:"mood"`
	Instruments        []string  `json:"instruments"`
	BPM                int       `json:"bpm"`
	Key                string    `json:"key"`
	Duration           int       `json:"duration"`
	Embedding          []float64 `json:"embedding"`
	SimilarTracks      []int64   `json:"similar_tracks"`
	PopularityScore    float64   `json:"popularity_score"`
}

// RecommendationRequest requête de recommandation
type RecommendationRequest struct {
	UserID    int64   `json:"user_id"`
	Context   string  `json:"context"` // "discovery", "collaboration", "production", "learning"
	Limit     int     `json:"limit"`
	Filters   Filters `json:"filters"`
	Freshness float64 `json:"freshness"` // 0-1, 0 = très populaire, 1 = très récent
}

// Filters filtres pour les recommandations
type Filters struct {
	Genres        []string `json:"genres,omitempty"`
	Artists       []string `json:"artists,omitempty"`
	Instruments   []string `json:"instruments,omitempty"`
	Moods         []string `json:"moods,omitempty"`
	MinDuration   *int     `json:"min_duration,omitempty"`
	MaxDuration   *int     `json:"max_duration,omitempty"`
	MinBPM        *int     `json:"min_bpm,omitempty"`
	MaxBPM        *int     `json:"max_bpm,omitempty"`
	SkillLevel    string   `json:"skill_level,omitempty"`
	Collaboration bool     `json:"collaboration,omitempty"`
}

// RecommendationResponse réponse de recommandation
type RecommendationResponse struct {
	Tracks      []TrackRecommendation    `json:"tracks"`
	Artists     []ArtistRecommendation   `json:"artists"`
	Samples     []SampleRecommendation   `json:"samples"`
	Playlists   []PlaylistRecommendation `json:"playlists"`
	Confidence  float64                  `json:"confidence"`
	Context     string                   `json:"context"`
	GeneratedAt time.Time                `json:"generated_at"`
}

// TrackRecommendation recommandation de piste
type TrackRecommendation struct {
	TrackID       int64   `json:"track_id"`
	Title         string  `json:"title"`
	Artist        string  `json:"artist"`
	Genre         string  `json:"genre"`
	Mood          string  `json:"mood"`
	Similarity    float64 `json:"similarity"` // 0-1
	Popularity    float64 `json:"popularity"` // 0-1
	Freshness     float64 `json:"freshness"`  // 0-1
	Collaboration bool    `json:"collaboration"`
	Reason        string  `json:"reason"` // Explication de la recommandation
}

// ArtistRecommendation recommandation d'artiste
type ArtistRecommendation struct {
	ArtistID      int64   `json:"artist_id"`
	Name          string  `json:"name"`
	Genre         string  `json:"genre"`
	Similarity    float64 `json:"similarity"`
	Collaboration bool    `json:"collaboration"`
	Reason        string  `json:"reason"`
}

// SampleRecommendation recommandation de sample
type SampleRecommendation struct {
	SampleID   int64   `json:"sample_id"`
	Name       string  `json:"name"`
	Category   string  `json:"category"`
	Similarity float64 `json:"similarity"`
	Popularity float64 `json:"popularity"`
	Reason     string  `json:"reason"`
}

// PlaylistRecommendation recommandation de playlist
type PlaylistRecommendation struct {
	PlaylistID  int64   `json:"playlist_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Similarity  float64 `json:"similarity"`
	Reason      string  `json:"reason"`
}

// NewRecommendationEngine crée un nouveau moteur de recommandations
func NewRecommendationEngine(
	userProfileService UserProfileService,
	trackAnalyticsService TrackAnalyticsService,
	mlModel MLModel,
	cache RecommendationCache,
	logger *zap.Logger,
) *RecommendationEngine {
	return &RecommendationEngine{
		userProfileService:    userProfileService,
		trackAnalyticsService: trackAnalyticsService,
		mlModel:               mlModel,
		cache:                 cache,
		logger:                logger,
	}
}

// GetRecommendations génère des recommandations personnalisées
func (r *RecommendationEngine) GetRecommendations(ctx context.Context, req *RecommendationRequest) (*RecommendationResponse, error) {
	// Vérifier le cache d'abord
	if cached, err := r.cache.GetRecommendations(ctx, req.UserID, req.Context); err == nil && cached != nil {
		r.logger.Info("📦 Recommendations served from cache",
			zap.Int64("user_id", req.UserID),
			zap.String("context", req.Context),
		)
		return cached, nil
	}

	// Récupérer le profil utilisateur
	userProfile, err := r.userProfileService.GetUserProfile(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	// Prédire les préférences avec le ML
	preferences, err := r.mlModel.PredictUserPreferences(ctx, userProfile)
	if err != nil {
		return nil, fmt.Errorf("failed to predict user preferences: %w", err)
	}

	// Générer les recommandations
	recommendations, err := r.generateRecommendations(ctx, userProfile, preferences, req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate recommendations: %w", err)
	}

	// Mettre en cache
	if err := r.cache.SetRecommendations(ctx, req.UserID, req.Context, recommendations, 30*time.Minute); err != nil {
		r.logger.Warn("Failed to cache recommendations", zap.Error(err))
	}

	r.logger.Info("🎯 Generated personalized recommendations",
		zap.Int64("user_id", req.UserID),
		zap.String("context", req.Context),
		zap.Int("tracks_count", len(recommendations.Tracks)),
		zap.Float64("confidence", recommendations.Confidence),
	)

	return recommendations, nil
}

// generateRecommendations génère les recommandations basées sur le profil et les préférences
func (r *RecommendationEngine) generateRecommendations(
	ctx context.Context,
	userProfile *UserProfile,
	preferences *UserPreferences,
	req *RecommendationRequest,
) (*RecommendationResponse, error) {
	recommendations := &RecommendationResponse{
		Context:     req.Context,
		GeneratedAt: time.Now(),
	}

	// Générer recommandations de pistes
	tracks, err := r.generateTrackRecommendations(ctx, userProfile, preferences, req)
	if err != nil {
		return nil, err
	}
	recommendations.Tracks = tracks

	// Générer recommandations d'artistes
	artists, err := r.generateArtistRecommendations(ctx, userProfile, preferences, req)
	if err != nil {
		return nil, err
	}
	recommendations.Artists = artists

	// Générer recommandations de samples
	samples, err := r.generateSampleRecommendations(ctx, userProfile, preferences, req)
	if err != nil {
		return nil, err
	}
	recommendations.Samples = samples

	// Calculer la confiance globale
	recommendations.Confidence = r.calculateConfidence(userProfile, preferences, len(tracks))

	return recommendations, nil
}

// generateTrackRecommendations génère les recommandations de pistes
func (r *RecommendationEngine) generateTrackRecommendations(
	ctx context.Context,
	userProfile *UserProfile,
	preferences *UserPreferences,
	req *RecommendationRequest,
) ([]TrackRecommendation, error) {
	var recommendations []TrackRecommendation

	// Récupérer les pistes populaires par genre préféré
	for genre, weight := range preferences.GenreWeights {
		if weight < 0.3 { // Seuil minimum
			continue
		}

		trackIDs, err := r.trackAnalyticsService.GetPopularTracks(ctx, genre, req.Limit/2)
		if err != nil {
			r.logger.Warn("Failed to get popular tracks", zap.String("genre", genre), zap.Error(err))
			continue
		}

		for _, trackID := range trackIDs {
			analytics, err := r.trackAnalyticsService.GetTrackAnalytics(ctx, trackID)
			if err != nil {
				continue
			}

			// Calculer la similarité
			similarity := r.calculateTrackSimilarity(userProfile, analytics, preferences)

			// Appliquer les filtres
			if !r.applyTrackFilters(analytics, req.Filters) {
				continue
			}

			recommendation := TrackRecommendation{
				TrackID:       trackID,
				Title:         "Track Title", // À récupérer depuis la DB
				Artist:        "Artist Name", // À récupérer depuis la DB
				Genre:         analytics.Genre,
				Mood:          analytics.Mood,
				Similarity:    similarity,
				Popularity:    analytics.PopularityScore,
				Freshness:     req.Freshness,
				Collaboration: analytics.CollaborationCount > 0,
				Reason:        r.generateRecommendationReason(userProfile, analytics, preferences),
			}

			recommendations = append(recommendations, recommendation)
		}
	}

	// Trier par score combiné (similarité + popularité + fraîcheur)
	sort.Slice(recommendations, func(i, j int) bool {
		scoreI := recommendations[i].Similarity*0.5 + recommendations[i].Popularity*0.3 + recommendations[i].Freshness*0.2
		scoreJ := recommendations[j].Similarity*0.5 + recommendations[j].Popularity*0.3 + recommendations[j].Freshness*0.2
		return scoreI > scoreJ
	})

	// Limiter le nombre de recommandations
	if len(recommendations) > req.Limit {
		recommendations = recommendations[:req.Limit]
	}

	return recommendations, nil
}

// generateArtistRecommendations génère les recommandations d'artistes
func (r *RecommendationEngine) generateArtistRecommendations(
	ctx context.Context,
	userProfile *UserProfile,
	preferences *UserPreferences,
	req *RecommendationRequest,
) ([]ArtistRecommendation, error) {
	var recommendations []ArtistRecommendation

	// Logique de recommandation d'artistes basée sur les préférences
	// À implémenter selon les besoins

	return recommendations, nil
}

// generateSampleRecommendations génère les recommandations de samples
func (r *RecommendationEngine) generateSampleRecommendations(
	ctx context.Context,
	userProfile *UserProfile,
	preferences *UserPreferences,
	req *RecommendationRequest,
) ([]SampleRecommendation, error) {
	var recommendations []SampleRecommendation

	// Logique de recommandation de samples basée sur les préférences
	// À implémenter selon les besoins

	return recommendations, nil
}

// calculateTrackSimilarity calcule la similarité entre un utilisateur et une piste
func (r *RecommendationEngine) calculateTrackSimilarity(
	userProfile *UserProfile,
	trackAnalytics *TrackAnalytics,
	preferences *UserPreferences,
) float64 {
	similarity := 0.0

	// Similarité par genre
	if weight, exists := preferences.GenreWeights[trackAnalytics.Genre]; exists {
		similarity += weight * 0.4
	}

	// Similarité par mood
	if weight, exists := preferences.MoodWeights[trackAnalytics.Mood]; exists {
		similarity += weight * 0.3
	}

	// Similarité par instruments
	for _, instrument := range trackAnalytics.Instruments {
		if weight, exists := userProfile.Instruments[instrument]; exists {
			similarity += weight * 0.2
		}
	}

	// Similarité par embedding (si disponible)
	if len(userProfile.Embedding) > 0 && len(trackAnalytics.Embedding) > 0 {
		embeddingSimilarity := r.mlModel.CalculateSimilarity(userProfile.Embedding, trackAnalytics.Embedding)
		similarity += embeddingSimilarity * 0.1
	}

	return math.Min(similarity, 1.0)
}

// applyTrackFilters applique les filtres sur une piste
func (r *RecommendationEngine) applyTrackFilters(analytics *TrackAnalytics, filters Filters) bool {
	// Filtre par genre
	if len(filters.Genres) > 0 {
		found := false
		for _, genre := range filters.Genres {
			if genre == analytics.Genre {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Filtre par durée
	if filters.MinDuration != nil && analytics.Duration < *filters.MinDuration {
		return false
	}
	if filters.MaxDuration != nil && analytics.Duration > *filters.MaxDuration {
		return false
	}

	// Filtre par BPM
	if filters.MinBPM != nil && analytics.BPM < *filters.MinBPM {
		return false
	}
	if filters.MaxBPM != nil && analytics.BPM > *filters.MaxBPM {
		return false
	}

	// Filtre par collaboration
	if filters.Collaboration && analytics.CollaborationCount == 0 {
		return false
	}

	return true
}

// generateRecommendationReason génère une explication de la recommandation
func (r *RecommendationEngine) generateRecommendationReason(
	userProfile *UserProfile,
	analytics *TrackAnalytics,
	preferences *UserPreferences,
) string {
	reasons := []string{}

	// Raison basée sur le genre
	if weight, exists := preferences.GenreWeights[analytics.Genre]; exists && weight > 0.5 {
		reasons = append(reasons, fmt.Sprintf("Basé sur votre intérêt pour le %s", analytics.Genre))
	}

	// Raison basée sur le mood
	if weight, exists := preferences.MoodWeights[analytics.Mood]; exists && weight > 0.5 {
		reasons = append(reasons, fmt.Sprintf("Correspond à votre mood préféré: %s", analytics.Mood))
	}

	// Raison basée sur la popularité
	if analytics.PopularityScore > 0.8 {
		reasons = append(reasons, "Très populaire dans la communauté")
	}

	// Raison basée sur la collaboration
	if analytics.CollaborationCount > 0 {
		reasons = append(reasons, "Opportunité de collaboration")
	}

	if len(reasons) == 0 {
		return "Recommandation personnalisée basée sur vos préférences"
	}

	return reasons[0] // Retourner la première raison
}

// calculateConfidence calcule la confiance globale des recommandations
func (r *RecommendationEngine) calculateConfidence(
	userProfile *UserProfile,
	preferences *UserPreferences,
	tracksCount int,
) float64 {
	confidence := preferences.Confidence

	// Ajuster selon la quantité de données utilisateur
	if len(userProfile.RecentActivity) < 10 {
		confidence *= 0.7 // Moins de confiance si peu d'activité
	}

	// Ajuster selon le nombre de recommandations
	if tracksCount < 5 {
		confidence *= 0.8
	}

	return math.Min(confidence, 1.0)
}

// UpdateUserProfile met à jour le profil utilisateur après une interaction
func (r *RecommendationEngine) UpdateUserProfile(ctx context.Context, userID int64, activity UserActivity) error {
	// Récupérer le profil actuel
	profile, err := r.userProfileService.GetUserProfile(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user profile: %w", err)
	}

	// Mettre à jour les préférences basées sur l'activité
	r.updateProfileFromActivity(profile, activity)

	// Sauvegarder le profil mis à jour
	if err := r.userProfileService.UpdateUserProfile(ctx, userID, profile); err != nil {
		return fmt.Errorf("failed to update user profile: %w", err)
	}

	// Invalider le cache des recommandations
	if err := r.cache.InvalidateUserRecommendations(ctx, userID); err != nil {
		r.logger.Warn("Failed to invalidate recommendations cache", zap.Error(err))
	}

	r.logger.Info("🔄 Updated user profile from activity",
		zap.Int64("user_id", userID),
		zap.String("activity_type", activity.Type),
		zap.Int64("track_id", activity.TrackID),
	)

	return nil
}

// updateProfileFromActivity met à jour le profil basé sur une activité
func (r *RecommendationEngine) updateProfileFromActivity(profile *UserProfile, activity UserActivity) {
	// Ajouter l'activité à l'historique
	profile.RecentActivity = append(profile.RecentActivity, activity)

	// Limiter l'historique à 100 activités
	if len(profile.RecentActivity) > 100 {
		profile.RecentActivity = profile.RecentActivity[len(profile.RecentActivity)-100:]
	}

	// Mettre à jour les préférences de genre
	if activity.Genre != "" {
		currentWeight := profile.Genres[activity.Genre]
		newWeight := currentWeight + activity.Interaction*0.1
		profile.Genres[activity.Genre] = math.Min(newWeight, 1.0)
	}

	// Mettre à jour les préférences d'artiste
	if activity.ArtistID != 0 {
		// Logique de mise à jour des préférences d'artiste
		// À implémenter selon les besoins
	}

	// Mettre à jour le timestamp
	profile.LastUpdated = time.Now()
}
