package services

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/hashicorp/vault/api"
	"go.uber.org/zap"
)

// Interaction représente une interaction utilisateur-track
type Interaction struct {
	UserID    int64     `json:"user_id"`
	TrackID   int64     `json:"track_id"`
	Type      string    `json:"type"` // "listen", "like", "purchase", "share"
	Weight    float64   `json:"weight"`
	Timestamp time.Time `json:"timestamp"`
}

// TrackSimilarity représente la similarité entre deux tracks
type TrackSimilarity struct {
	TrackID     int64   `json:"track_id"`
	Similarity  float64 `json:"similarity"`
	CommonUsers int     `json:"common_users"`
}

// RecommendationService fournit des recommandations algorithmiques
type RecommendationService struct {
	db          *sql.DB
	cache       *redis.Client
	vaultClient *api.Client
	logger      *zap.Logger
}

// NewRecommendationService crée une nouvelle instance du service
func NewRecommendationService(db *sql.DB, cache *redis.Client, vaultClient *api.Client, logger *zap.Logger) *RecommendationService {
	return &RecommendationService{
		db:          db,
		cache:       cache,
		vaultClient: vaultClient,
		logger:      logger,
	}
}

// GetRecommendations retourne les recommandations pour un utilisateur
func (s *RecommendationService) GetRecommendations(userID int64, limit int) ([]Track, error) {
	ctx := context.Background()

	// Vérifier le cache d'abord
	cacheKey := fmt.Sprintf("recommendations:%d:%d", userID, limit)
	cached, err := s.cache.Get(ctx, cacheKey).Result()
	if err == nil {
		// TODO: Désérialiser depuis le cache
		s.logger.Info("Recommendations served from cache", zap.Int64("user_id", userID))
		return []Track{}, nil // Placeholder
	}

	// 1. Récupérer les interactions utilisateur
	interactions, err := s.getUserInteractions(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user interactions: %w", err)
	}

	// Si pas d'interactions, retourner les plus populaires
	if len(interactions) == 0 {
		return s.getPopularTracks(limit)
	}

	// 2. Calculer la similarité cosinus
	similarities, err := s.calculateCosineSimilarity(interactions)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate similarities: %w", err)
	}

	// 3. Obtenir les recommandations top-N
	recommendations, err := s.getTopRecommendations(similarities, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get top recommendations: %w", err)
	}

	// Mettre en cache pour 1 heure
	s.cache.Set(ctx, cacheKey, "placeholder", time.Hour)

	return recommendations, nil
}

// getUserInteractions récupère les interactions d'un utilisateur
func (s *RecommendationService) getUserInteractions(userID int64) ([]Interaction, error) {
	query := `
		SELECT user_id, track_id, interaction_type, 
		       CASE 
		           WHEN interaction_type = 'listen' THEN 1.0
		           WHEN interaction_type = 'like' THEN 2.0
		           WHEN interaction_type = 'purchase' THEN 5.0
		           WHEN interaction_type = 'share' THEN 3.0
		           ELSE 1.0
		       END as weight,
		       created_at
		FROM user_track_interactions
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var interactions []Interaction
	for rows.Next() {
		var interaction Interaction
		err := rows.Scan(&interaction.UserID, &interaction.TrackID, &interaction.Type, &interaction.Weight, &interaction.Timestamp)
		if err != nil {
			return nil, err
		}
		interactions = append(interactions, interaction)
	}

	return interactions, nil
}

// calculateCosineSimilarity calcule la similarité cosinus entre tracks
func (s *RecommendationService) calculateCosineSimilarity(interactions []Interaction) ([]TrackSimilarity, error) {
	if len(interactions) == 0 {
		return []TrackSimilarity{}, nil
	}

	// Construire la matrice d'interactions
	userTracks := make(map[int64]map[int64]float64)
	for _, interaction := range interactions {
		if userTracks[interaction.UserID] == nil {
			userTracks[interaction.UserID] = make(map[int64]float64)
		}
		userTracks[interaction.UserID][interaction.TrackID] += interaction.Weight
	}

	// Calculer les similarités cosinus
	similarities := make(map[int64]float64)
	trackCounts := make(map[int64]int)

	// Pour chaque track de l'utilisateur
	for userID, tracks := range userTracks {
		for trackID, weight := range tracks {
			// Trouver les autres utilisateurs qui ont écouté ce track
			otherUsers, err := s.getUsersWhoListenedToTrack(trackID)
			if err != nil {
				s.logger.Warn("Failed to get users for track", zap.Int64("track_id", trackID), zap.Error(err))
				continue
			}

			// Calculer la similarité avec chaque autre track
			for _, otherUser := range otherUsers {
				if otherUser == userID {
					continue
				}

				otherTracks, err := s.getUserTrackWeights(otherUser)
				if err != nil {
					continue
				}

				// Calculer la similarité cosinus
				similarity := s.cosineSimilarity(tracks, otherTracks)
				if similarity > 0.1 { // Seuil minimal
					for otherTrackID := range otherTracks {
						if otherTrackID != trackID {
							similarities[otherTrackID] += similarity
							trackCounts[otherTrackID]++
						}
					}
				}
			}
		}
	}

	// Normaliser et trier les similarités
	var result []TrackSimilarity
	for trackID, totalSimilarity := range similarities {
		if count := trackCounts[trackID]; count > 0 {
			avgSimilarity := totalSimilarity / float64(count)
			result = append(result, TrackSimilarity{
				TrackID:     trackID,
				Similarity:  avgSimilarity,
				CommonUsers: count,
			})
		}
	}

	// Trier par similarité décroissante
	sort.Slice(result, func(i, j int) bool {
		return result[i].Similarity > result[j].Similarity
	})

	return result, nil
}

// cosineSimilarity calcule la similarité cosinus entre deux vecteurs
func (s *RecommendationService) cosineSimilarity(vec1, vec2 map[int64]float64) float64 {
	dotProduct := 0.0
	magnitude1 := 0.0
	magnitude2 := 0.0

	// Calculer le produit scalaire et les magnitudes
	for trackID, weight1 := range vec1 {
		weight2 := vec2[trackID]
		dotProduct += weight1 * weight2
		magnitude1 += weight1 * weight1
	}

	for _, weight2 := range vec2 {
		magnitude2 += weight2 * weight2
	}

	if magnitude1 == 0 || magnitude2 == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(magnitude1) * math.Sqrt(magnitude2))
}

// getUsersWhoListenedToTrack récupère les utilisateurs qui ont écouté un track
func (s *RecommendationService) getUsersWhoListenedToTrack(trackID int64) ([]int64, error) {
	query := `
		SELECT DISTINCT user_id
		FROM user_track_interactions
		WHERE track_id = $1
	`

	rows, err := s.db.Query(query, trackID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userIDs []int64
	for rows.Next() {
		var userID int64
		err := rows.Scan(&userID)
		if err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}

	return userIDs, nil
}

// getUserTrackWeights récupère les poids des tracks d'un utilisateur
func (s *RecommendationService) getUserTrackWeights(userID int64) (map[int64]float64, error) {
	query := `
		SELECT track_id, 
		       SUM(CASE 
		           WHEN interaction_type = 'listen' THEN 1.0
		           WHEN interaction_type = 'like' THEN 2.0
		           WHEN interaction_type = 'purchase' THEN 5.0
		           WHEN interaction_type = 'share' THEN 3.0
		           ELSE 1.0
		       END) as weight
		FROM user_track_interactions
		WHERE user_id = $1
		GROUP BY track_id
	`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	weights := make(map[int64]float64)
	for rows.Next() {
		var trackID int64
		var weight float64
		err := rows.Scan(&trackID, &weight)
		if err != nil {
			return nil, err
		}
		weights[trackID] = weight
	}

	return weights, nil
}

// getTopRecommendations récupère les tracks recommandés
func (s *RecommendationService) getTopRecommendations(similarities []TrackSimilarity, userID int64, limit int) ([]Track, error) {
	if len(similarities) == 0 {
		return s.getPopularTracks(limit)
	}

	// Extraire les track IDs
	var trackIDs []int64
	for _, similarity := range similarities {
		trackIDs = append(trackIDs, similarity.TrackID)
		if len(trackIDs) >= limit {
			break
		}
	}

	// Récupérer les détails des tracks
	query := `
		SELECT id, title, artist, genre, duration, created_at
		FROM tracks
		WHERE id = ANY($1)
		ORDER BY array_position($1, id)
	`

	rows, err := s.db.Query(query, trackIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tracks []Track
	for rows.Next() {
		var track Track
		err := rows.Scan(&track.ID, &track.Title, &track.Artist, &track.Genre, &track.Duration, &track.CreatedAt)
		if err != nil {
			return nil, err
		}
		tracks = append(tracks, track)
	}

	return tracks, nil
}

// getPopularTracks retourne les tracks les plus populaires
func (s *RecommendationService) getPopularTracks(limit int) ([]Track, error) {
	query := `
		SELECT t.id, t.title, t.artist, t.genre, t.duration, t.created_at
		FROM tracks t
		LEFT JOIN user_track_interactions uti ON t.id = uti.track_id
		GROUP BY t.id
		ORDER BY COUNT(uti.id) DESC, t.created_at DESC
		LIMIT $1
	`

	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tracks []Track
	for rows.Next() {
		var track Track
		err := rows.Scan(&track.ID, &track.Title, &track.Artist, &track.Genre, &track.Duration, &track.CreatedAt)
		if err != nil {
			return nil, err
		}
		tracks = append(tracks, track)
	}

	return tracks, nil
}

// UpdateUserInteraction met à jour les interactions utilisateur
func (s *RecommendationService) UpdateUserInteraction(userID, trackID int64, interactionType string) error {
	query := `
		INSERT INTO user_track_interactions (user_id, track_id, interaction_type, created_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (user_id, track_id, interaction_type) 
		DO UPDATE SET created_at = NOW()
	`

	_, err := s.db.Exec(query, userID, trackID, interactionType)
	if err != nil {
		return fmt.Errorf("failed to update user interaction: %w", err)
	}

	// Invalider le cache des recommandations
	cacheKey := fmt.Sprintf("recommendations:%d:*", userID)
	s.cache.Del(context.Background(), cacheKey)

	return nil
}
