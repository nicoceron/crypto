package recommendation

import (
	"context"
	"testing"
	"time"

	"stock-analyzer/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockStockRepository is a mock implementation of domain.StockRepository
type MockStockRepository struct {
	mock.Mock
}

func (m *MockStockRepository) CreateStockRating(ctx context.Context, rating *domain.StockRating) error {
	args := m.Called(ctx, rating)
	return args.Error(0)
}

func (m *MockStockRepository) CreateStockRatingsBatch(ctx context.Context, ratings []domain.StockRating) error {
	args := m.Called(ctx, ratings)
	return args.Error(0)
}

func (m *MockStockRepository) GetStockRatings(ctx context.Context, page, limit int, sortBy, order, search string) ([]domain.StockRating, int, error) {
	args := m.Called(ctx, page, limit, sortBy, order, search)
	return args.Get(0).([]domain.StockRating), args.Int(1), args.Error(2)
}

func (m *MockStockRepository) GetStockRatingsByTicker(ctx context.Context, ticker string) ([]domain.StockRating, error) {
	args := m.Called(ctx, ticker)
	return args.Get(0).([]domain.StockRating), args.Error(1)
}

func (m *MockStockRepository) GetUniqueTickers(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockStockRepository) CreateEnrichedStockData(ctx context.Context, data *domain.EnrichedStockData) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

func (m *MockStockRepository) GetEnrichedStockData(ctx context.Context, ticker string) (*domain.EnrichedStockData, error) {
	args := m.Called(ctx, ticker)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.EnrichedStockData), args.Error(1)
}

func (m *MockStockRepository) GetLatestRatingsByTicker(ctx context.Context) (map[string]*domain.StockRating, error) {
	args := m.Called(ctx)
	return args.Get(0).(map[string]*domain.StockRating), args.Error(1)
}

func TestService_filterPositiveRatings(t *testing.T) {
	service := NewService(nil)

	tests := []struct {
		name     string
		ratings  map[string]*domain.StockRating
		expected int
	}{
		{
			name: "filters buy ratings",
			ratings: map[string]*domain.StockRating{
				"AAPL": {
					Ticker:   "AAPL",
					RatingTo: "Buy",
					Action:   "upgraded by",
				},
				"GOOGL": {
					Ticker:   "GOOGL",
					RatingTo: "Hold",
					Action:   "maintained by",
				},
			},
			expected: 1,
		},
		{
			name: "filters strong buy ratings",
			ratings: map[string]*domain.StockRating{
				"TSLA": {
					Ticker:   "TSLA",
					RatingTo: "Strong Buy",
					Action:   "initiated by",
				},
			},
			expected: 1,
		},
		{
			name: "filters positive actions",
			ratings: map[string]*domain.StockRating{
				"MSFT": {
					Ticker:   "MSFT",
					RatingTo: "Hold",
					Action:   "upgraded by",
				},
			},
			expected: 1,
		},
		{
			name:     "empty ratings",
			ratings:  map[string]*domain.StockRating{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.filterPositiveRatings(tt.ratings)
			assert.Len(t, result, tt.expected)
		})
	}
}

func TestService_isUpgrade(t *testing.T) {
	service := NewService(nil)

	tests := []struct {
		name     string
		from     *string
		to       *string
		expected bool
	}{
		{
			name:     "upgrade from hold to buy",
			from:     stringPtr("Hold"),
			to:       stringPtr("Buy"),
			expected: true,
		},
		{
			name:     "upgrade from sell to buy",
			from:     stringPtr("Sell"),
			to:       stringPtr("Buy"),
			expected: true,
		},
		{
			name:     "downgrade from buy to hold",
			from:     stringPtr("Buy"),
			to:       stringPtr("Hold"),
			expected: false,
		},
		{
			name:     "same rating",
			from:     stringPtr("Buy"),
			to:       stringPtr("Buy"),
			expected: false,
		},
		{
			name:     "nil from rating",
			from:     nil,
			to:       stringPtr("Buy"),
			expected: false,
		},
		{
			name:     "nil to rating",
			from:     stringPtr("Hold"),
			to:       nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.isUpgrade(tt.from, tt.to)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestService_analyzeTechnical(t *testing.T) {
	service := NewService(nil)

	tests := []struct {
		name           string
		historicalData map[string]interface{}
		expectedSignal string
		expectedScore  float64
	}{
		{
			name: "positive trend",
			historicalData: map[string]interface{}{
				"data": []map[string]interface{}{
					{"close": 100.0},
					{"close": 105.0},
				},
			},
			expectedSignal: "Golden Cross",
			expectedScore:  0.8,
		},
		{
			name: "negative trend",
			historicalData: map[string]interface{}{
				"data": []map[string]interface{}{
					{"close": 100.0},
					{"close": 90.0},
				},
			},
			expectedSignal: "Death Cross",
			expectedScore:  0.2,
		},
		{
			name: "sideways trend",
			historicalData: map[string]interface{}{
				"data": []map[string]interface{}{
					{"close": 100.0},
					{"close": 99.0},
				},
			},
			expectedSignal: "Sideways",
			expectedScore:  0.5,
		},
		{
			name:           "insufficient data",
			historicalData: map[string]interface{}{},
			expectedSignal: "Insufficient Data",
			expectedScore:  0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signal, score := service.analyzeTechnical(tt.historicalData)
			assert.Equal(t, tt.expectedSignal, signal)
			assert.Equal(t, tt.expectedScore, score)
		})
	}
}

func TestService_analyzeSentiment(t *testing.T) {
	service := NewService(nil)

	tests := []struct {
		name          string
		sentimentData map[string]interface{}
		expected      *float64
	}{
		{
			name: "positive sentiment",
			sentimentData: map[string]interface{}{
				"sentiment_score": 0.8,
			},
			expected: float64Ptr(0.9), // (0.8 + 1) / 2
		},
		{
			name: "negative sentiment",
			sentimentData: map[string]interface{}{
				"sentiment_score": -0.6,
			},
			expected: float64Ptr(0.2), // (-0.6 + 1) / 2
		},
		{
			name: "neutral sentiment",
			sentimentData: map[string]interface{}{
				"sentiment_score": 0.0,
			},
			expected: float64Ptr(0.5), // (0.0 + 1) / 2
		},
		{
			name:          "no sentiment data",
			sentimentData: map[string]interface{}{},
			expected:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.analyzeSentiment(tt.sentimentData)
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.InDelta(t, *tt.expected, *result, 0.01)
			}
		})
	}
}

func TestService_GenerateRecommendations(t *testing.T) {
	mockRepo := new(MockStockRepository)
	service := NewService(mockRepo)

	// Setup test data
	latestRatings := map[string]*domain.StockRating{
		"AAPL": {
			RatingID: uuid.New(),
			Ticker:   "AAPL",
			Company:  "Apple Inc.",
			RatingTo: "Buy",
			Action:   "upgraded by",
			Time:     time.Now(),
		},
		"GOOGL": {
			RatingID: uuid.New(),
			Ticker:   "GOOGL",
			Company:  "Alphabet Inc.",
			RatingTo: "Hold",
			Action:   "maintained by",
			Time:     time.Now(),
		},
	}

	// Mock repository calls
	mockRepo.On("GetLatestRatingsByTicker", mock.Anything).Return(latestRatings, nil)

	// Execute test
	ctx := context.Background()
	recommendations, err := service.GenerateRecommendations(ctx)

	// Assertions
	assert.NoError(t, err)
	assert.Len(t, recommendations, 1) // Only AAPL should be recommended
	assert.Equal(t, "AAPL", recommendations[0].Ticker)
	assert.Equal(t, "Apple Inc.", recommendations[0].Company)
	assert.Greater(t, recommendations[0].Score, 0.6) // Should meet threshold

	mockRepo.AssertExpectations(t)
}

func TestService_createBasicRecommendation(t *testing.T) {
	service := NewService(nil)

	tests := []struct {
		name          string
		rating        *domain.StockRating
		expectedScore float64
	}{
		{
			name: "strong buy rating",
			rating: &domain.StockRating{
				Ticker:   "AAPL",
				Company:  "Apple Inc.",
				RatingTo: "Strong Buy",
			},
			expectedScore: 0.9, // 0.7 base + 0.2 bonus
		},
		{
			name: "buy rating",
			rating: &domain.StockRating{
				Ticker:   "MSFT",
				Company:  "Microsoft Corp.",
				RatingTo: "Buy",
			},
			expectedScore: 0.85, // 0.7 base + 0.15 bonus
		},
		{
			name: "outperform rating",
			rating: &domain.StockRating{
				Ticker:   "GOOGL",
				Company:  "Alphabet Inc.",
				RatingTo: "Outperform",
			},
			expectedScore: 0.8, // 0.7 base + 0.1 bonus
		},
		{
			name: "unknown rating",
			rating: &domain.StockRating{
				Ticker:   "TSLA",
				Company:  "Tesla Inc.",
				RatingTo: "Unknown",
			},
			expectedScore: 0.7, // 0.7 base only
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.createBasicRecommendation(tt.rating)
			assert.Equal(t, tt.rating.Ticker, result.Ticker)
			assert.Equal(t, tt.rating.Company, result.Company)
			assert.InDelta(t, tt.expectedScore, result.Score, 0.01)
			assert.Equal(t, "Pending Analysis", result.TechnicalSignal)
			assert.Nil(t, result.SentimentScore)
		})
	}
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func float64Ptr(f float64) *float64 {
	return &f
} 