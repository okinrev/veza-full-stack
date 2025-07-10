package recommendations

import (
	"context"
	"math"
	"math/rand"
	"time"

	"go.uber.org/zap"
)

// SimpleMLModel implémentation basique du modèle ML
type SimpleMLModel struct {
	logger *zap.Logger
}

// NewSimpleMLModel crée un nouveau modèle ML simple
func NewSimpleMLModel(logger *zap.Logger) *SimpleMLModel {
	return &SimpleMLModel{
		logger: logger,
	}
}

// PredictUserPreferences prédit les préférences utilisateur
func (m *SimpleMLModel) PredictUserPreferences(ctx context.Context, userProfile *UserProfile) (*UserPreferences, error) {
	preferences := &UserPreferences{
		GenreWeights:      make(map[string]float64),
		ArtistWeights:     make(map[string]float64),
		MoodWeights:       make(map[string]float64),
		CollaborationPref: 0.5,
		SkillLevelPref:    "intermediate",
		ProductionPref:    "mixed",
		Confidence:        0.7,
	}

	// Analyser les activités récentes pour prédire les préférences
	m.analyzeRecentActivity(userProfile, preferences)

	// Analyser les genres préférés
	m.analyzeGenrePreferences(userProfile, preferences)

	// Analyser le style de collaboration
	m.analyzeCollaborationStyle(userProfile, preferences)

	// Analyser le niveau de compétence
	m.analyzeSkillLevel(userProfile, preferences)

	// Calculer la confiance basée sur la quantité de données
	preferences.Confidence = m.calculateConfidence(userProfile)

	m.logger.Info("🤖 Predicted user preferences",
		zap.Int64("user_id", userProfile.UserID),
		zap.Float64("confidence", preferences.Confidence),
		zap.Int("genres_count", len(preferences.GenreWeights)),
	)

	return preferences, nil
}

// GetTrackEmbeddings récupère les embeddings des pistes
func (m *SimpleMLModel) GetTrackEmbeddings(ctx context.Context, trackIDs []int64) (map[int64][]float64, error) {
	embeddings := make(map[int64][]float64)

	// Pour l'instant, générer des embeddings aléatoires
	// En production, cela viendrait d'un vrai modèle ML
	for _, trackID := range trackIDs {
		embedding := make([]float64, 128) // Embedding de 128 dimensions
		for i := range embedding {
			embedding[i] = rand.Float64()*2 - 1 // Valeurs entre -1 et 1
		}
		embeddings[trackID] = embedding
	}

	return embeddings, nil
}

// CalculateSimilarity calcule la similarité entre deux embeddings
func (m *SimpleMLModel) CalculateSimilarity(embedding1, embedding2 []float64) float64 {
	if len(embedding1) != len(embedding2) {
		return 0.0
	}

	// Calcul de la similarité cosinus
	var dotProduct, norm1, norm2 float64

	for i := range embedding1 {
		dotProduct += embedding1[i] * embedding2[i]
		norm1 += embedding1[i] * embedding1[i]
		norm2 += embedding2[i] * embedding2[i]
	}

	if norm1 == 0 || norm2 == 0 {
		return 0.0
	}

	similarity := dotProduct / (math.Sqrt(norm1) * math.Sqrt(norm2))
	return math.Max(0, similarity) // Retourner une valeur positive
}

// analyzeRecentActivity analyse les activités récentes pour prédire les préférences
func (m *SimpleMLModel) analyzeRecentActivity(userProfile *UserProfile, preferences *UserPreferences) {
	if len(userProfile.RecentActivity) == 0 {
		return
	}

	// Analyser les 30 derniers jours
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	recentActivity := make([]UserActivity, 0)

	for _, activity := range userProfile.RecentActivity {
		if activity.Timestamp.After(thirtyDaysAgo) {
			recentActivity = append(recentActivity, activity)
		}
	}

	// Calculer les poids basés sur la fréquence et l'intensité
	genreCounts := make(map[string]int)
	totalInteractions := 0.0

	for _, activity := range recentActivity {
		if activity.Genre != "" {
			genreCounts[activity.Genre]++
		}
		if activity.Type == "like" || activity.Type == "share" {
			totalInteractions += activity.Interaction
		}
	}

	// Normaliser les poids des genres
	maxCount := 1
	for _, count := range genreCounts {
		if count > maxCount {
			maxCount = count
		}
	}

	for genre, count := range genreCounts {
		preferences.GenreWeights[genre] = float64(count) / float64(maxCount)
	}

	// Ajuster la confiance basée sur l'activité récente
	if len(recentActivity) > 10 {
		preferences.Confidence = math.Min(preferences.Confidence+0.1, 0.9)
	}
}

// analyzeGenrePreferences analyse les préférences de genre
func (m *SimpleMLModel) analyzeGenrePreferences(userProfile *UserProfile, preferences *UserPreferences) {
	// Utiliser les préférences existantes du profil
	for genre, weight := range userProfile.Genres {
		preferences.GenreWeights[genre] = weight
	}

	// Ajouter des genres populaires si peu de données
	if len(preferences.GenreWeights) < 3 {
		popularGenres := []string{"electronic", "hip-hop", "rock", "pop", "jazz"}
		for _, genre := range popularGenres {
			if _, exists := preferences.GenreWeights[genre]; !exists {
				preferences.GenreWeights[genre] = 0.3 // Poids par défaut
			}
		}
	}
}

// analyzeCollaborationStyle analyse le style de collaboration
func (m *SimpleMLModel) analyzeCollaborationStyle(userProfile *UserProfile, preferences *UserPreferences) {
	collaborationCount := 0
	totalActivities := len(userProfile.RecentActivity)

	for _, activity := range userProfile.RecentActivity {
		if activity.Type == "collaborate" {
			collaborationCount++
		}
	}

	if totalActivities > 0 {
		collaborationRatio := float64(collaborationCount) / float64(totalActivities)
		preferences.CollaborationPref = collaborationRatio

		if collaborationRatio > 0.5 {
			preferences.CollaborationPref = 0.8
		} else if collaborationRatio > 0.2 {
			preferences.CollaborationPref = 0.5
		} else {
			preferences.CollaborationPref = 0.2
		}
	}
}

// analyzeSkillLevel analyse le niveau de compétence
func (m *SimpleMLModel) analyzeSkillLevel(userProfile *UserProfile, preferences *UserPreferences) {
	// Basé sur la fréquence d'activité et les types d'interactions
	advancedActivities := 0
	totalActivities := len(userProfile.RecentActivity)

	for _, activity := range userProfile.RecentActivity {
		if activity.Type == "collaborate" || activity.Type == "purchase" {
			advancedActivities++
		}
	}

	if totalActivities > 0 {
		advancedRatio := float64(advancedActivities) / float64(totalActivities)

		if advancedRatio > 0.6 {
			preferences.SkillLevelPref = "advanced"
		} else if advancedRatio > 0.3 {
			preferences.SkillLevelPref = "intermediate"
		} else {
			preferences.SkillLevelPref = "beginner"
		}
	}
}

// calculateConfidence calcule la confiance du modèle
func (m *SimpleMLModel) calculateConfidence(userProfile *UserProfile) float64 {
	confidence := 0.5 // Confiance de base

	// Augmenter la confiance avec plus d'activité
	if len(userProfile.RecentActivity) > 20 {
		confidence += 0.2
	} else if len(userProfile.RecentActivity) > 10 {
		confidence += 0.1
	}

	// Augmenter la confiance avec plus de genres
	if len(userProfile.Genres) > 5 {
		confidence += 0.1
	} else if len(userProfile.Genres) > 2 {
		confidence += 0.05
	}

	// Augmenter la confiance avec des activités récentes
	recentActivity := 0
	for _, activity := range userProfile.RecentActivity {
		if activity.Timestamp.After(time.Now().AddDate(0, 0, -7)) {
			recentActivity++
		}
	}

	if recentActivity > 5 {
		confidence += 0.1
	}

	return math.Min(confidence, 0.9) // Max 90% de confiance
}
