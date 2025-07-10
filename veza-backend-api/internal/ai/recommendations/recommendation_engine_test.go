package recommendations

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockUserProfileService mock pour le service de profil utilisateur
type MockUserProfileService struct {
	mock.Mock
}

func (m *MockUserProfileService) GetUserProfile(ctx context.Context, userID int64) (*UserProfile, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserProfile), args.Error(1)
}

func (m *MockUserProfileService) UpdateUserProfile(ctx context.Context, userID int64, profile *UserProfile) error {
	args := m.Called(ctx, userID, profile)
	return args.Error(0)
}

// MockTrackAnalyticsService mock pour le service d'analytics des pistes
type MockTrackAnalyticsService struct {
	mock.Mock
}

func (m *MockTrackAnalyticsService) GetTrackAnalytics(ctx context.Context, trackID int64) (*TrackAnalytics, error) {
	args := m.Called(ctx, trackID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TrackAnalytics), args.Error(1)
}

func (m *MockTrackAnalyticsService) GetSimilarTracks(ctx context.Context, trackID int64, limit int) ([]int64, error) {
	args := m.Called(ctx, trackID, limit)
	return args.Get(0).([]int64), args.Error(1)
}

func (m *MockTrackAnalyticsService) GetPopularTracks(ctx context.Context, genre string, limit int) ([]int64, error) {
	args := m.Called(ctx, genre, limit)
	return args.Get(0).([]int64), args.Error(1)
}

// MockMLModel mock pour le modèle ML
type MockMLModel struct {
	mock.Mock
}

func (m *MockMLModel) PredictUserPreferences(ctx context.Context, userProfile *UserProfile) (*UserPreferences, error) {
	args := m.Called(ctx, userProfile)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserPreferences), args.Error(1)
}

func (m *MockMLModel) GetTrackEmbeddings(ctx context.Context, trackIDs []int64) (map[int64][]float64, error) {
	args := m.Called(ctx, trackIDs)
	return args.Get(0).(map[int64][]float64), args.Error(1)
}

func (m *MockMLModel) CalculateSimilarity(embedding1, embedding2 []float64) float64 {
	args := m.Called(embedding1, embedding2)
	return args.Get(0).(float64)
}

// MockRecommendationCache mock pour le cache
type MockRecommendationCache struct {
	mock.Mock
}

func (m *MockRecommendationCache) GetRecommendations(ctx context.Context, userID int64, context string) (*RecommendationResponse, error) {
	args := m.Called(ctx, userID, context)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*RecommendationResponse), args.Error(1)
}

func (m *MockRecommendationCache) SetRecommendations(ctx context.Context, userID int64, context string, recommendations *RecommendationResponse, ttl time.Duration) error {
	args := m.Called(ctx, userID, context, recommendations, ttl)
	return args.Error(0)
}

func (m *MockRecommendationCache) InvalidateUserRecommendations(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func TestRecommendationEngine_GetRecommendations(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	mockUserProfileService := &MockUserProfileService{}
	mockTrackAnalyticsService := &MockTrackAnalyticsService{}
	mockMLModel := &MockMLModel{}
	mockCache := &MockRecommendationCache{}

	engine := NewRecommendationEngine(
		mockUserProfileService,
		mockTrackAnalyticsService,
		mockMLModel,
		mockCache,
		logger,
	)

	ctx := context.Background()

	// Test data
	userProfile := &UserProfile{
		UserID: 1,
		Genres: map[string]float64{
			"electronic": 0.8,
			"hip-hop":    0.6,
		},
		RecentActivity: []UserActivity{
			{
				Type:        "listen",
				TrackID:     123,
				Genre:       "electronic",
				Timestamp:   time.Now(),
				Duration:    180,
				Interaction: 0.8,
			},
		},
		LastUpdated: time.Now(),
	}

	preferences := &UserPreferences{
		GenreWeights: map[string]float64{
			"electronic": 0.8,
			"hip-hop":    0.6,
		},
		Confidence: 0.7,
	}

	// Mock expectations
	mockCache.On("GetRecommendations", ctx, int64(1), "discovery").Return(nil, nil)
	mockUserProfileService.On("GetUserProfile", ctx, int64(1)).Return(userProfile, nil)
	mockMLModel.On("PredictUserPreferences", ctx, userProfile).Return(preferences, nil)
	mockTrackAnalyticsService.On("GetPopularTracks", ctx, "electronic", 10).Return([]int64{123, 456}, nil)
	mockTrackAnalyticsService.On("GetPopularTracks", ctx, "hip-hop", 10).Return([]int64{789}, nil)
	mockTrackAnalyticsService.On("GetTrackAnalytics", ctx, int64(123)).Return(&TrackAnalytics{
		TrackID:         123,
		Genre:           "electronic",
		PopularityScore: 0.8,
	}, nil)
	mockTrackAnalyticsService.On("GetTrackAnalytics", ctx, int64(456)).Return(&TrackAnalytics{
		TrackID:         456,
		Genre:           "electronic",
		PopularityScore: 0.7,
	}, nil)
	mockTrackAnalyticsService.On("GetTrackAnalytics", ctx, int64(789)).Return(&TrackAnalytics{
		TrackID:         789,
		Genre:           "hip-hop",
		PopularityScore: 0.6,
	}, nil)
	mockCache.On("SetRecommendations", mock.Anything, int64(1), "discovery", mock.Anything, mock.Anything).Return(nil)

	// Execute
	req := &RecommendationRequest{
		UserID:  1,
		Context: "discovery",
		Limit:   20,
		Filters: Filters{},
	}

	result, err := engine.GetRecommendations(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "discovery", result.Context)
	assert.True(t, len(result.Tracks) > 0)
	assert.True(t, result.Confidence > 0)

	// Verify mocks
	mockCache.AssertExpectations(t)
	mockUserProfileService.AssertExpectations(t)
	mockTrackAnalyticsService.AssertExpectations(t)
	mockMLModel.AssertExpectations(t)
}

func TestRecommendationEngine_UpdateUserProfile(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	mockUserProfileService := &MockUserProfileService{}
	mockTrackAnalyticsService := &MockTrackAnalyticsService{}
	mockMLModel := &MockMLModel{}
	mockCache := &MockRecommendationCache{}

	engine := NewRecommendationEngine(
		mockUserProfileService,
		mockTrackAnalyticsService,
		mockMLModel,
		mockCache,
		logger,
	)

	ctx := context.Background()

	// Test data
	userProfile := &UserProfile{
		UserID: 1,
		Genres: map[string]float64{
			"electronic": 0.5,
		},
		RecentActivity: []UserActivity{},
		LastUpdated:    time.Now(),
	}

	activity := UserActivity{
		Type:        "listen",
		TrackID:     123,
		Genre:       "electronic",
		Timestamp:   time.Now(),
		Duration:    180,
		Interaction: 0.8,
	}

	// Mock expectations
	mockUserProfileService.On("GetUserProfile", ctx, int64(1)).Return(userProfile, nil)
	mockUserProfileService.On("UpdateUserProfile", ctx, int64(1), mock.Anything).Return(nil)
	mockCache.On("InvalidateUserRecommendations", ctx, int64(1)).Return(nil)

	// Execute
	err := engine.UpdateUserProfile(ctx, 1, activity)

	// Assert
	assert.NoError(t, err)

	// Verify mocks
	mockUserProfileService.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestRecommendationEngine_CalculateTrackSimilarity(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	mockUserProfileService := &MockUserProfileService{}
	mockTrackAnalyticsService := &MockTrackAnalyticsService{}
	mockMLModel := &MockMLModel{}
	mockCache := &MockRecommendationCache{}

	engine := NewRecommendationEngine(
		mockUserProfileService,
		mockTrackAnalyticsService,
		mockMLModel,
		mockCache,
		logger,
	)

	// Test data
	userProfile := &UserProfile{
		UserID: 1,
		Genres: map[string]float64{
			"electronic": 0.8,
		},
		Instruments: map[string]float64{
			"synth": 0.7,
		},
	}

	preferences := &UserPreferences{
		GenreWeights: map[string]float64{
			"electronic": 0.8,
		},
		MoodWeights: map[string]float64{
			"energetic": 0.6,
		},
	}

	trackAnalytics := &TrackAnalytics{
		TrackID:     123,
		Genre:       "electronic",
		Mood:        "energetic",
		Instruments: []string{"synth"},
	}

	// Execute
	similarity := engine.calculateTrackSimilarity(userProfile, trackAnalytics, preferences)

	// Assert
	assert.True(t, similarity > 0)
	assert.True(t, similarity <= 1.0)
}

func TestRecommendationEngine_ApplyTrackFilters(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	mockUserProfileService := &MockUserProfileService{}
	mockTrackAnalyticsService := &MockTrackAnalyticsService{}
	mockMLModel := &MockMLModel{}
	mockCache := &MockRecommendationCache{}

	engine := NewRecommendationEngine(
		mockUserProfileService,
		mockTrackAnalyticsService,
		mockMLModel,
		mockCache,
		logger,
	)

	// Test cases
	testCases := []struct {
		name           string
		analytics      *TrackAnalytics
		filters        Filters
		expectedResult bool
	}{
		{
			name: "passes all filters",
			analytics: &TrackAnalytics{
				Genre:    "electronic",
				Duration: 180,
				BPM:      120,
			},
			filters: Filters{
				Genres:      []string{"electronic"},
				MinDuration: intPtr(60),
				MaxDuration: intPtr(300),
				MinBPM:      intPtr(100),
				MaxBPM:      intPtr(140),
			},
			expectedResult: true,
		},
		{
			name: "fails genre filter",
			analytics: &TrackAnalytics{
				Genre: "rock",
			},
			filters: Filters{
				Genres: []string{"electronic"},
			},
			expectedResult: false,
		},
		{
			name: "fails duration filter",
			analytics: &TrackAnalytics{
				Duration: 30,
			},
			filters: Filters{
				MinDuration: intPtr(60),
			},
			expectedResult: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := engine.applyTrackFilters(tc.analytics, tc.filters)
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}

// Helper function
func intPtr(i int) *int {
	return &i
}
