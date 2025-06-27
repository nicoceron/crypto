package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"stock-analyzer/internal/domain"
	apperrors "stock-analyzer/pkg/errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockStockRepository is a mock implementation of domain.StockRepository
type MockStockRepository struct {
	mock.Mock
}

func (m *MockStockRepository) CreateStockRating(ctx context.Context, rating *domain.StockRating) error {
	args := m.Called(ctx, rating)
	return args.Error(0)
}

func (m *MockStockRepository) CreateStockRatingsBatch(ctx context.Context, ratings []*domain.StockRating) (int, error) {
	args := m.Called(ctx, ratings)
	return args.Int(0), args.Error(1)
}

func (m *MockStockRepository) GetStockRatings(ctx context.Context, filters domain.FilterOptions) (*domain.PaginatedResponse[domain.StockRating], error) {
	args := m.Called(ctx, filters)
	return args.Get(0).(*domain.PaginatedResponse[domain.StockRating]), args.Error(1)
}

func (m *MockStockRepository) GetStockRatingsByTicker(ctx context.Context, ticker string) ([]domain.StockRating, error) {
	args := m.Called(ctx, ticker)
	return args.Get(0).([]domain.StockRating), args.Error(1)
}

func (m *MockStockRepository) GetUniqueTickers(ctx context.Context) ([]string, error) {
	args := m.Called()
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

func (m *MockStockRepository) DeleteOldEnrichedData(ctx context.Context, olderThan time.Time) (int64, error) {
	args := m.Called(ctx, olderThan)
	return args.Get(0).(int64), args.Error(1)
}

// MockIngestionService is a mock implementation of domain.IngestionService
type MockIngestionService struct {
	mock.Mock
}

func (m *MockIngestionService) IngestAllData(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockIngestionService) EnrichStockData(ctx context.Context, tickers []string) error {
	args := m.Called(ctx, tickers)
	return args.Error(0)
}

// MockRecommendationService is a mock implementation of domain.RecommendationService
type MockRecommendationService struct {
	mock.Mock
}

func (m *MockRecommendationService) GenerateRecommendations(ctx context.Context) ([]domain.StockRecommendation, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.StockRecommendation), args.Error(1)
}

func (m *MockRecommendationService) GetCachedRecommendations(ctx context.Context) ([]domain.StockRecommendation, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.StockRecommendation), args.Error(1)
}

// MockAlpacaService is a mock implementation of alpaca.Service
type MockAlpacaService struct {
	mock.Mock
}

func (m *MockAlpacaService) GetHistoricalBars(ctx context.Context, symbol string, timeframe string, start, end time.Time) ([]domain.PriceBar, error) {
	args := m.Called(ctx, symbol, timeframe, start, end)
	return args.Get(0).([]domain.PriceBar), args.Error(1)
}

func (m *MockAlpacaService) GetSnapshot(ctx context.Context, symbol string) (*domain.Snapshot, error) {
	args := m.Called(ctx, symbol)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Snapshot), args.Error(1)
}

func (m *MockAlpacaService) GetRecentBars(ctx context.Context, symbol string) ([]domain.PriceBar, error) {
	args := m.Called(ctx, symbol)
	return args.Get(0).([]domain.PriceBar), args.Error(1)
}

func (m *MockAlpacaService) IsMarketHours() bool {
	args := m.Called()
	return args.Bool(0)
}

func setupTestHandlers() (*Handlers, *MockStockRepository, *MockIngestionService, *MockRecommendationService, *MockAlpacaService) {
	stockRepo := &MockStockRepository{}
	ingestionSvc := &MockIngestionService{}
	recommendationSvc := &MockRecommendationService{}
	alpacaSvc := &MockAlpacaService{}

	handlers := NewHandlers(stockRepo, ingestionSvc, recommendationSvc, alpacaSvc)

	return handlers, stockRepo, ingestionSvc, recommendationSvc, alpacaSvc
}

func setupGinRouter(handlers *Handlers) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add middleware
	router.Use(ErrorHandler())

	// Setup routes
	v1 := router.Group("/api/v1")
	{
		v1.GET("/ratings", handlers.GetStockRatings)
		v1.GET("/ratings/:ticker", handlers.GetStockRatingsByTicker)
		v1.GET("/recommendations", handlers.GetRecommendations)
		v1.GET("/stocks/:symbol/price", handlers.GetStockPrice)
		v1.GET("/stocks/:symbol/logo", handlers.GetStockLogo)
		v1.POST("/ingest", handlers.TriggerIngestion)
	}

	return router
}

func TestGetStockRatings_Success(t *testing.T) {
	t.Log("Testing GetStockRatings: successful retrieval")
	handlers, stockRepo, _, _, _ := setupTestHandlers()
	router := setupGinRouter(handlers)

	// Test data
	ratings := []domain.StockRating{
		{
			RatingID:  uuid.New(),
			Ticker:    "AAPL",
			Company:   "Apple Inc.",
			Brokerage: "Goldman Sachs",
			Action:    "upgraded by",
			RatingTo:  "Buy",
			Time:      time.Now(),
			CreatedAt: time.Now(),
		},
		{
			RatingID:  uuid.New(),
			Ticker:    "GOOGL",
			Company:   "Alphabet Inc.",
			Brokerage: "Morgan Stanley",
			Action:    "initiated by",
			RatingTo:  "Strong Buy",
			Time:      time.Now().Add(-time.Hour),
			CreatedAt: time.Now().Add(-time.Hour),
		},
	}

	expectedResponse := &domain.PaginatedResponse[domain.StockRating]{
		Data: ratings,
		Pagination: domain.Pagination{
			Page:       1,
			Limit:      20,
			TotalItems: 2,
			TotalPages: 1,
		},
	}
	stockRepo.On("GetStockRatings", mock.Anything, mock.MatchedBy(func(filters domain.FilterOptions) bool {
		return filters.Page == 1 && filters.Limit == 20 && filters.SortBy == "time" && filters.SortDesc && filters.Search == ""
	})).Return(expectedResponse, nil)

	// Test request
	req, _ := http.NewRequest("GET", "/api/v1/ratings", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response domain.PaginatedResponse[domain.StockRating]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Len(t, response.Data, 2)
	assert.Equal(t, 1, response.Pagination.Page)
	assert.Equal(t, 20, response.Pagination.Limit)
	assert.Equal(t, 2, response.Pagination.TotalItems)
	assert.Equal(t, 1, response.Pagination.TotalPages)

	stockRepo.AssertExpectations(t)
}

func TestGetStockRatings_WithPagination(t *testing.T) {
	t.Log("Testing GetStockRatings: with pagination")
	handlers, stockRepo, _, _, _ := setupTestHandlers()
	router := setupGinRouter(handlers)

	// Mock large dataset
	ratings := make([]domain.StockRating, 5)
	for i := 0; i < 5; i++ {
		ratings[i] = domain.StockRating{
			RatingID:  uuid.New(),
			Ticker:    fmt.Sprintf("TICK%d", i),
			Company:   fmt.Sprintf("Company %d", i),
			Brokerage: "Test Brokerage",
			Action:    "upgraded by",
			RatingTo:  "Buy",
			Time:      time.Now().Add(-time.Duration(i) * time.Hour),
			CreatedAt: time.Now().Add(-time.Duration(i) * time.Hour),
		}
	}

	expectedResponse := &domain.PaginatedResponse[domain.StockRating]{
		Data: ratings,
		Pagination: domain.Pagination{
			Page:       2,
			Limit:      5,
			TotalItems: 100,
			TotalPages: 20,
		},
	}
	stockRepo.On("GetStockRatings", mock.Anything, mock.MatchedBy(func(filters domain.FilterOptions) bool {
		return filters.Page == 2 && filters.Limit == 5 && filters.SortBy == "time" && filters.SortDesc && filters.Search == ""
	})).Return(expectedResponse, nil)

	// Test request with pagination
	req, _ := http.NewRequest("GET", "/api/v1/ratings?page=2&limit=5", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response domain.PaginatedResponse[domain.StockRating]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Len(t, response.Data, 5)
	assert.Equal(t, 2, response.Pagination.Page)
	assert.Equal(t, 5, response.Pagination.Limit)
	assert.Equal(t, 100, response.Pagination.TotalItems)
	assert.Equal(t, 20, response.Pagination.TotalPages)

	stockRepo.AssertExpectations(t)
}

func TestGetStockRatings_WithSearch(t *testing.T) {
	t.Log("Testing GetStockRatings: with search parameter")
	handlers, stockRepo, _, _, _ := setupTestHandlers()
	router := setupGinRouter(handlers)

	ratings := []domain.StockRating{
		{
			RatingID:  uuid.New(),
			Ticker:    "AAPL",
			Company:   "Apple Inc.",
			Brokerage: "Goldman Sachs",
			Action:    "upgraded by",
			RatingTo:  "Buy",
			Time:      time.Now(),
			CreatedAt: time.Now(),
		},
	}

	expectedResponse := &domain.PaginatedResponse[domain.StockRating]{
		Data: ratings,
		Pagination: domain.Pagination{
			Page:       1,
			Limit:      20,
			TotalItems: 1,
			TotalPages: 1,
		},
	}
	stockRepo.On("GetStockRatings", mock.Anything, mock.MatchedBy(func(filters domain.FilterOptions) bool {
		return filters.Page == 1 && filters.Limit == 20 && filters.SortBy == "time" && filters.SortDesc && filters.Search == "Apple"
	})).Return(expectedResponse, nil)

	// Test request with search
	req, _ := http.NewRequest("GET", "/api/v1/ratings?search=Apple", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response domain.PaginatedResponse[domain.StockRating]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Len(t, response.Data, 1)
	assert.Contains(t, response.Data[0].Company, "Apple")

	stockRepo.AssertExpectations(t)
}

func TestGetStockRatings_WithSorting(t *testing.T) {
	t.Log("Testing GetStockRatings: with sorting parameters")
	handlers, stockRepo, _, _, _ := setupTestHandlers()
	router := setupGinRouter(handlers)

	ratings := []domain.StockRating{
		{
			RatingID:  uuid.New(),
			Ticker:    "AAPL",
			Company:   "Apple Inc.",
			Brokerage: "Goldman Sachs",
			Action:    "upgraded by",
			RatingTo:  "Buy",
			Time:      time.Now(),
			CreatedAt: time.Now(),
		},
	}

	expectedResponse := &domain.PaginatedResponse[domain.StockRating]{
		Data: ratings,
		Pagination: domain.Pagination{
			Page:       1,
			Limit:      20,
			TotalItems: 1,
			TotalPages: 1,
		},
	}
	stockRepo.On("GetStockRatings", mock.Anything, mock.MatchedBy(func(filters domain.FilterOptions) bool {
		return filters.Page == 1 && filters.Limit == 20 && filters.SortBy == "ticker" && !filters.SortDesc && filters.Search == ""
	})).Return(expectedResponse, nil)

	// Test request with sorting
	req, _ := http.NewRequest("GET", "/api/v1/ratings?sort_by=ticker&order=asc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response domain.PaginatedResponse[domain.StockRating]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Len(t, response.Data, 1)

	stockRepo.AssertExpectations(t)
}

func TestGetStockRatings_InvalidParameters(t *testing.T) {
	t.Log("Testing GetStockRatings: with invalid pagination parameters")
	handlers, stockRepo, _, _, _ := setupTestHandlers()
	router := setupGinRouter(handlers)

	// Test invalid page parameter
	req, _ := http.NewRequest("GET", "/api/v1/ratings?page=invalid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errorResp ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.Contains(t, errorResp.Details, "invalid page parameter")

	// Test invalid limit parameter
	req, _ = http.NewRequest("GET", "/api/v1/ratings?limit=invalid", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.Contains(t, errorResp.Details, "invalid limit parameter")

	stockRepo.AssertExpectations(t)
}

func TestGetStockRatings_DatabaseError(t *testing.T) {
	t.Log("Testing GetStockRatings: repository returns an error")
	handlers, stockRepo, _, _, _ := setupTestHandlers()
	router := setupGinRouter(handlers)

	stockRepo.On("GetStockRatings", mock.Anything, mock.Anything).Return((*domain.PaginatedResponse[domain.StockRating])(nil), apperrors.ErrDatabaseFailure)

	req, _ := http.NewRequest("GET", "/api/v1/ratings", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var errorResp ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.Equal(t, apperrors.ErrCodeDatabase, errorResp.Code)

	stockRepo.AssertExpectations(t)
}

func TestGetStockRatingsByTicker_Success(t *testing.T) {
	t.Log("Testing GetStockRatingsByTicker: successful retrieval for a single ticker")
	handlers, stockRepo, _, _, _ := setupTestHandlers()
	router := setupGinRouter(handlers)

	ratings := []domain.StockRating{
		{
			RatingID:  uuid.New(),
			Ticker:    "AAPL",
			Company:   "Apple Inc.",
			Brokerage: "Goldman Sachs",
			Action:    "upgraded by",
			RatingTo:  "Buy",
			Time:      time.Now(),
			CreatedAt: time.Now(),
		},
	}

	stockRepo.On("GetStockRatingsByTicker", mock.Anything, "AAPL").Return(ratings, nil)

	req, _ := http.NewRequest("GET", "/api/v1/ratings/AAPL", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []domain.StockRating
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Len(t, response, 1)
	assert.Equal(t, "AAPL", response[0].Ticker)

	stockRepo.AssertExpectations(t)
}

func TestGetStockRatingsByTicker_NotFound(t *testing.T) {
	t.Log("Testing GetStockRatingsByTicker: ticker not found")
	handlers, stockRepo, _, _, _ := setupTestHandlers()
	router := setupGinRouter(handlers)

	stockRepo.On("GetStockRatingsByTicker", mock.Anything, "NONEXISTENT").Return([]domain.StockRating{}, nil)

	req, _ := http.NewRequest("GET", "/api/v1/ratings/NONEXISTENT", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var errorResp ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.Contains(t, errorResp.Details, "no ratings found for ticker")

	stockRepo.AssertExpectations(t)
}

func TestGetStockPrice_Success(t *testing.T) {
	t.Log("Testing GetStockPrice: successful retrieval of price data")
	handlers, _, _, _, alpacaSvc := setupTestHandlers()
	router := setupGinRouter(handlers)

	priceBars := []domain.PriceBar{
		{
			Timestamp: "2023-12-01T09:30:00Z",
			Open:      100.0,
			High:      105.0,
			Low:       99.0,
			Close:     104.0,
			Volume:    1000000,
		},
		{
			Timestamp: "2023-12-01T10:30:00Z",
			Open:      104.0,
			High:      106.0,
			Low:       103.0,
			Close:     105.5,
			Volume:    800000,
		},
	}

	alpacaSvc.On("GetHistoricalBars", mock.Anything, "AAPL", "1Hour", mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).Return(priceBars, nil)

	req, _ := http.NewRequest("GET", "/api/v1/stocks/AAPL/price?period=1M", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response StockPriceResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "AAPL", response.Symbol)
	assert.Len(t, response.Bars, 2)
	assert.Equal(t, 100.0, response.Bars[0].Open)
	assert.Equal(t, 105.5, response.Bars[1].Close)

	alpacaSvc.AssertExpectations(t)
}

func TestGetStockPrice_DifferentPeriods(t *testing.T) {
	t.Log("Testing GetStockPrice: handling different time periods")
	handlers, _, _, _, alpacaSvc := setupTestHandlers()
	router := setupGinRouter(handlers)

	testCases := []struct {
		period     string
		expectedTF string
	}{
		{"1W", "1Hour"},
		{"1M", "1Hour"},
		{"3M", "1Day"},
		{"6M", "1Day"},
		{"1Y", "1Day"},
		{"2Y", "1Day"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("period_%s", tc.period), func(t *testing.T) {
			priceBars := []domain.PriceBar{
				{
					Timestamp: "2023-12-01T09:30:00Z",
					Open:      100.0,
					High:      105.0,
					Low:       99.0,
					Close:     104.0,
					Volume:    1000000,
				},
			}

			alpacaSvc.On("GetHistoricalBars", mock.Anything, "AAPL", tc.expectedTF, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).Return(priceBars, nil).Once()

			req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/stocks/AAPL/price?period=%s", tc.period), nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response StockPriceResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, "AAPL", response.Symbol)
			assert.Len(t, response.Bars, 1)
		})
	}

	alpacaSvc.AssertExpectations(t)
}

func TestGetStockPrice_NoData(t *testing.T) {
	t.Log("Testing GetStockPrice: when Alpaca service returns no data")
	handlers, _, _, _, alpacaSvc := setupTestHandlers()
	router := setupGinRouter(handlers)

	alpacaSvc.On("GetHistoricalBars", mock.Anything, "INVALID", "1Hour", mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).Return([]domain.PriceBar{}, nil)

	req, _ := http.NewRequest("GET", "/api/v1/stocks/INVALID/price", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var errorResp ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.Contains(t, errorResp.Details, "No price data available")

	alpacaSvc.AssertExpectations(t)
}

func TestGetStockPrice_AlpacaError(t *testing.T) {
	t.Log("Testing GetStockPrice: when Alpaca service returns an error")
	handlers, _, _, _, alpacaSvc := setupTestHandlers()
	router := setupGinRouter(handlers)

	alpacaSvc.On("GetHistoricalBars", mock.Anything, "AAPL", "1Hour", mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).Return([]domain.PriceBar{}, fmt.Errorf("API error"))

	req, _ := http.NewRequest("GET", "/api/v1/stocks/AAPL/price", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	alpacaSvc.AssertExpectations(t)
}

func TestGetStockLogo_Success(t *testing.T) {
	t.Log("Testing GetStockLogo: successful retrieval of logo URL")
	handlers, _, _, _, _ := setupTestHandlers()
	router := setupGinRouter(handlers)

	req, _ := http.NewRequest("GET", "/api/v1/stocks/AAPL/logo", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response StockLogoResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "AAPL", response.Symbol)
	assert.Contains(t, response.LogoURL, "clearbit.com")
	assert.Contains(t, response.LogoURL, "aapl.com")

	// Check cache headers
	assert.Equal(t, "public, max-age=3600", w.Header().Get("Cache-Control"))
	assert.Equal(t, `"AAPL"`, w.Header().Get("ETag"))
}

func TestGetStockLogo_MissingSymbol(t *testing.T) {
	t.Log("Testing GetStockLogo: when symbol parameter is missing")
	handlers, _, _, _, _ := setupTestHandlers()
	router := setupGinRouter(handlers)

	req, _ := http.NewRequest("GET", "/api/v1/stocks//logo", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errorResp ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.Contains(t, errorResp.Details, "symbol parameter is required")
}

func TestGetRecommendations_Success(t *testing.T) {
	t.Log("Testing GetRecommendations: successful retrieval")
	handlers, _, _, recommendationSvc, _ := setupTestHandlers()
	router := setupGinRouter(handlers)

	recommendations := []domain.StockRecommendation{
		{
			Ticker:          "AAPL",
			Company:         "Apple Inc.",
			Score:           0.85,
			Rationale:       "Strong fundamentals and positive sentiment",
			LatestRating:    "Buy",
			TargetPrice:     &[]float64{180.0}[0],
			TechnicalSignal: "Golden Cross",
			SentimentScore:  &[]float64{0.7}[0],
			GeneratedAt:     time.Now(),
		},
	}

	recommendationSvc.On("GetCachedRecommendations", mock.Anything).Return(recommendations, nil)

	req, _ := http.NewRequest("GET", "/api/v1/recommendations", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []domain.StockRecommendation
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Len(t, response, 1)
	assert.Equal(t, "AAPL", response[0].Ticker)
	assert.Equal(t, 0.85, response[0].Score)

	recommendationSvc.AssertExpectations(t)
}

func TestGetRecommendations_ServiceError(t *testing.T) {
	t.Log("Testing GetRecommendations: when recommendation service returns an error")
	handlers, _, _, recommendationSvc, _ := setupTestHandlers()
	router := setupGinRouter(handlers)

	recommendationSvc.On("GetCachedRecommendations", mock.Anything).Return([]domain.StockRecommendation{}, fmt.Errorf("service error"))

	req, _ := http.NewRequest("GET", "/api/v1/recommendations", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	recommendationSvc.AssertExpectations(t)
}

func TestTriggerIngestion_Success(t *testing.T) {
	t.Log("Testing TriggerIngestion: successfully triggers ingestion service")
	handlers, _, ingestionSvc, _, _ := setupTestHandlers()
	router := setupGinRouter(handlers)

	// The handler starts ingestion asynchronously, so we need to handle this carefully
	ingestionSvc.On("IngestAllData", mock.Anything).Return(nil).Maybe()

	req, _ := http.NewRequest("POST", "/api/v1/ingest", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Data ingestion started", response["message"])
	assert.Equal(t, "accepted", response["status"])

	// Wait a bit for the goroutine to potentially start
	time.Sleep(10 * time.Millisecond)
}

func TestHealthCheck(t *testing.T) {
	t.Log("Testing HealthCheck: endpoint returns OK")
	handlers, _, _, _, _ := setupTestHandlers()
	router := setupGinRouter(handlers)

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// Add health check route
	router.GET("/health", handlers.HealthCheck)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, "stock-analyzer", response["service"])
	assert.Contains(t, response, "timestamp")
}

func TestParseIntQuery(t *testing.T) {
	t.Log("Testing utility: ParseIntQuery")
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		queryValue   string
		defaultValue int
		expected     int
		expectError  bool
	}{
		{
			name:         "valid integer",
			queryValue:   "42",
			defaultValue: 10,
			expected:     42,
			expectError:  false,
		},
		{
			name:         "empty query uses default",
			queryValue:   "",
			defaultValue: 10,
			expected:     10,
			expectError:  false,
		},
		{
			name:         "invalid integer",
			queryValue:   "invalid",
			defaultValue: 10,
			expected:     0,
			expectError:  true,
		},
		{
			name:         "negative integer",
			queryValue:   "-5",
			defaultValue: 10,
			expected:     -5,
			expectError:  false,
		},
		{
			name:         "zero",
			queryValue:   "0",
			defaultValue: 10,
			expected:     0,
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("  - Sub-test: %s", tt.name)
			// Create a test context with query parameter
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			if tt.queryValue != "" {
				req, _ := http.NewRequest("GET", fmt.Sprintf("/?test=%s", tt.queryValue), nil)
				c.Request = req
			} else {
				req, _ := http.NewRequest("GET", "/", nil)
				c.Request = req
			}

			result, err := parseIntQuery(c, "test", tt.defaultValue)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.expected, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// Benchmarks for expensive operations
func BenchmarkGetStockRatings(b *testing.B) {
	b.Log("Benchmarking GetStockRatings endpoint")
	handlers, stockRepo, _, _, _ := setupTestHandlers()
	router := setupGinRouter(handlers)

	// Setup large dataset
	ratings := make([]domain.StockRating, 1000)
	for i := 0; i < 1000; i++ {
		ratings[i] = domain.StockRating{
			RatingID:  uuid.New(),
			Ticker:    fmt.Sprintf("TICK%d", i),
			Company:   fmt.Sprintf("Company %d", i),
			Brokerage: "Test Brokerage",
			Action:    "upgraded by",
			RatingTo:  "Buy",
			Time:      time.Now().Add(-time.Duration(i) * time.Hour),
			CreatedAt: time.Now().Add(-time.Duration(i) * time.Hour),
		}
	}

	expectedResponse := &domain.PaginatedResponse[domain.StockRating]{
		Data: ratings,
		Pagination: domain.Pagination{
			Page:       1,
			Limit:      20,
			TotalItems: 10000,
			TotalPages: 500,
		},
	}
	stockRepo.On("GetStockRatings", mock.Anything, mock.Anything).Return(expectedResponse, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/api/v1/ratings", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkGetStockPrice(b *testing.B) {
	b.Log("Benchmarking GetStockPrice endpoint")
	handlers, _, _, _, alpacaSvc := setupTestHandlers()
	router := setupGinRouter(handlers)

	// Setup large price data
	priceBars := make([]domain.PriceBar, 1000)
	for i := 0; i < 1000; i++ {
		priceBars[i] = domain.PriceBar{
			Timestamp: time.Now().Add(-time.Duration(i) * time.Hour).Format(time.RFC3339),
			Open:      100.0 + float64(i)*0.1,
			High:      105.0 + float64(i)*0.1,
			Low:       99.0 + float64(i)*0.1,
			Close:     104.0 + float64(i)*0.1,
			Volume:    int64(1000000 + i*1000),
		}
	}

	alpacaSvc.On("GetHistoricalBars", mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).Return(priceBars, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/api/v1/stocks/AAPL/price", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// Stress tests for concurrent operations
func TestConcurrentGetStockRatings(t *testing.T) {
	t.Log("Testing GetStockRatings: with high concurrency")
	handlers, stockRepo, _, _, _ := setupTestHandlers()
	router := setupGinRouter(handlers)

	ratings := []domain.StockRating{
		{
			RatingID:  uuid.New(),
			Ticker:    "AAPL",
			Company:   "Apple Inc.",
			Brokerage: "Goldman Sachs",
			Action:    "upgraded by",
			RatingTo:  "Buy",
			Time:      time.Now(),
			CreatedAt: time.Now(),
		},
	}

	expectedResponse := &domain.PaginatedResponse[domain.StockRating]{
		Data: ratings,
		Pagination: domain.Pagination{
			Page:       1,
			Limit:      20,
			TotalItems: 1,
			TotalPages: 1,
		},
	}
	stockRepo.On("GetStockRatings", mock.Anything, mock.Anything).Return(expectedResponse, nil)

	// Run 100 concurrent requests
	concurrency := 100
	done := make(chan bool, concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			req, _ := http.NewRequest("GET", "/api/v1/ratings", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			done <- true
		}()
	}

	// Wait for all requests to complete
	for i := 0; i < concurrency; i++ {
		<-done
	}

	stockRepo.AssertExpectations(t)
}

func TestConcurrentGetStockPrice(t *testing.T) {
	t.Log("Testing GetStockPrice: with high concurrency")
	handlers, _, _, _, alpacaSvc := setupTestHandlers()
	router := setupGinRouter(handlers)

	priceBars := []domain.PriceBar{
		{
			Timestamp: "2023-12-01T09:30:00Z",
			Open:      100.0,
			High:      105.0,
			Low:       99.0,
			Close:     104.0,
			Volume:    1000000,
		},
	}

	alpacaSvc.On("GetHistoricalBars", mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).Return(priceBars, nil)

	// Run 50 concurrent requests
	concurrency := 50
	done := make(chan bool, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(id int) {
			symbol := fmt.Sprintf("STOCK%d", id%10) // Use 10 different symbols
			req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/stocks/%s/price", symbol), nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			done <- true
		}(i)
	}

	// Wait for all requests to complete
	for i := 0; i < concurrency; i++ {
		<-done
	}

	alpacaSvc.AssertExpectations(t)
}
