package ingestion

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"stock-analyzer/internal/domain"
	apperrors "stock-analyzer/pkg/errors"

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

func (m *MockStockRepository) DeleteOldEnrichedData(ctx context.Context, olderThan time.Time) (int64, error) {
	args := m.Called(ctx, olderThan)
	return args.Get(0).(int64), args.Error(1)
}

func createMockAPIResponse(items []domain.APIStockRating, nextPage *string) *domain.APIResponse {
	return &domain.APIResponse{
		Items:    items,
		NextPage: nextPage,
	}
}

func createMockAPIItems(count int) []domain.APIStockRating {
	items := make([]domain.APIStockRating, count)
	for i := 0; i < count; i++ {
		items[i] = domain.APIStockRating{
			Ticker:     fmt.Sprintf("TICK%d", i),
			Company:    fmt.Sprintf("Company %d", i),
			Brokerage:  "Test Brokerage",
			Action:     "upgraded by",
			RatingFrom: "Hold",
			RatingTo:   "Buy",
			TargetFrom: "150.00",
			TargetTo:   "180.00",
			Time:       time.Now().Add(-time.Duration(i) * time.Hour).Format(time.RFC3339),
		}
	}
	return items
}

func TestNewService(t *testing.T) {
	t.Log("Testing NewService: initialization")
	stockRepo := &MockStockRepository{}
	apiURL := "https://api.example.com"
	apiToken := "test-token"

	service := NewService(stockRepo, apiURL, apiToken)

	assert.NotNil(t, service)
	assert.Equal(t, stockRepo, service.stockRepo)
	assert.Equal(t, apiURL, service.apiURL)
	assert.Equal(t, apiToken, service.apiToken)
	assert.NotNil(t, service.client)
	assert.Equal(t, 30*time.Second, service.client.Timeout)
}

func TestIngestAllData_Success_SinglePage(t *testing.T) {
	t.Log("Testing IngestAllData: success with a single page of data")
	stockRepo := &MockStockRepository{}

	// Mock API server
	items := createMockAPIItems(5)
	response := createMockAPIResponse(items, nil)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	service := NewService(stockRepo, server.URL, "test-token")

	// Mock repository expectation
	stockRepo.On("CreateStockRatingsBatch", mock.Anything, mock.MatchedBy(func(ratings []*domain.StockRating) bool {
		return len(ratings) == 5
	})).Return(5, nil)

	err := service.IngestAllData(context.Background())

	assert.NoError(t, err)
	stockRepo.AssertExpectations(t)
}

func TestIngestAllData_Success_MultiplePage(t *testing.T) {
	t.Log("Testing IngestAllData: success with multiple pages of data")
	stockRepo := &MockStockRepository{}

	// Mock API server with pagination
	page1Items := createMockAPIItems(3)
	page1Response := createMockAPIResponse(page1Items, stringPtr("page2"))

	page2Items := createMockAPIItems(2)
	page2Response := createMockAPIResponse(page2Items, nil)

	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++

		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("next_page") == "page2" {
			json.NewEncoder(w).Encode(page2Response)
		} else {
			json.NewEncoder(w).Encode(page1Response)
		}
	}))
	defer server.Close()

	service := NewService(stockRepo, server.URL, "test-token")

	// Mock repository expectations
	stockRepo.On("CreateStockRatingsBatch", mock.Anything, mock.MatchedBy(func(ratings []*domain.StockRating) bool {
		return len(ratings) == 3
	})).Return(3, nil).Once()

	stockRepo.On("CreateStockRatingsBatch", mock.Anything, mock.MatchedBy(func(ratings []*domain.StockRating) bool {
		return len(ratings) == 2
	})).Return(2, nil).Once()

	err := service.IngestAllData(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, 2, requestCount)
	stockRepo.AssertExpectations(t)
}

func TestIngestAllData_EmptyResponse(t *testing.T) {
	t.Log("Testing IngestAllData: handles empty API response")
	stockRepo := &MockStockRepository{}

	response := createMockAPIResponse([]domain.APIStockRating{}, nil)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	service := NewService(stockRepo, server.URL, "test-token")

	err := service.IngestAllData(context.Background())

	assert.NoError(t, err)
	stockRepo.AssertNotCalled(t, "CreateStockRatingsBatch")
}

func TestIngestAllData_APIError(t *testing.T) {
	t.Log("Testing IngestAllData: handles API error (e.g., 500 status)")
	stockRepo := &MockStockRepository{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	service := NewService(stockRepo, server.URL, "test-token")

	err := service.IngestAllData(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch data from API")

	var appErr *apperrors.AppError
	assert.ErrorAs(t, err, &appErr)
	stockRepo.AssertNotCalled(t, "CreateStockRatingsBatch")
}

func TestIngestAllData_TransformationError(t *testing.T) {
	t.Log("Testing IngestAllData: handles data transformation error")
	stockRepo := &MockStockRepository{}

	// Invalid data that will cause transformation to fail
	invalidItems := []domain.APIStockRating{
		{
			Ticker:     "AAPL",
			Company:    "Apple Inc.",
			Brokerage:  "Goldman Sachs",
			Action:     "upgraded by",
			RatingTo:   "Buy",
			TargetFrom: "invalid-number", // This will cause parsing error
			TargetTo:   "180.00",
			Time:       "invalid-time", // This will cause parsing error
		},
	}
	response := createMockAPIResponse(invalidItems, nil)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	service := NewService(stockRepo, server.URL, "test-token")

	err := service.IngestAllData(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to transform API ratings")
	stockRepo.AssertNotCalled(t, "CreateStockRatingsBatch")
}

func TestIngestAllData_RepositoryError(t *testing.T) {
	t.Log("Testing IngestAllData: handles repository error on batch create")
	stockRepo := &MockStockRepository{}

	items := createMockAPIItems(3)
	response := createMockAPIResponse(items, nil)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	service := NewService(stockRepo, server.URL, "test-token")

	// Mock repository error
	stockRepo.On("CreateStockRatingsBatch", mock.Anything, mock.Anything).Return(0, apperrors.ErrDatabaseFailure)

	err := service.IngestAllData(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to store ratings batch")
	stockRepo.AssertExpectations(t)
}

func TestIngestAllData_ContextCancellation(t *testing.T) {
	t.Log("Testing IngestAllData: handles context cancellation")
	stockRepo := &MockStockRepository{}

	// Server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)

		items := createMockAPIItems(2)
		response := createMockAPIResponse(items, nil)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	service := NewService(stockRepo, server.URL, "test-token")

	// Context that cancels quickly
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := service.IngestAllData(ctx)

	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "context canceled"))
	stockRepo.AssertNotCalled(t, "CreateStockRatingsBatch")
}

func TestFetchDataFromAPI_Success(t *testing.T) {
	t.Log("Testing fetchDataFromAPI: successful fetch")
	stockRepo := &MockStockRepository{}

	items := createMockAPIItems(3)
	expectedResponse := createMockAPIResponse(items, stringPtr("next_page_token"))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedResponse)
	}))
	defer server.Close()

	service := NewService(stockRepo, server.URL, "test-token")

	response, err := service.fetchDataFromAPI(context.Background(), nil)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Len(t, response.Items, 3)
	assert.Equal(t, "next_page_token", *response.NextPage)
}

func TestFetchDataFromAPI_WithNextPage(t *testing.T) {
	t.Log("Testing fetchDataFromAPI: includes next_page parameter")
	stockRepo := &MockStockRepository{}

	items := createMockAPIItems(2)
	expectedResponse := createMockAPIResponse(items, nil)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify next_page parameter is passed
		assert.Equal(t, "page_token_123", r.URL.Query().Get("next_page"))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedResponse)
	}))
	defer server.Close()

	service := NewService(stockRepo, server.URL, "test-token")

	nextPage := "page_token_123"
	response, err := service.fetchDataFromAPI(context.Background(), &nextPage)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Len(t, response.Items, 2)
}

func TestFetchDataFromAPI_HTTPError(t *testing.T) {
	t.Log("Testing fetchDataFromAPI: handles non-200 status code")
	stockRepo := &MockStockRepository{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad Request"))
	}))
	defer server.Close()

	service := NewService(stockRepo, server.URL, "test-token")

	response, err := service.fetchDataFromAPI(context.Background(), nil)

	assert.Error(t, err)
	assert.Nil(t, response)

	var appErr *apperrors.AppError
	assert.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperrors.ErrCodeUpstreamAPI, appErr.Code)
}

func TestFetchDataFromAPI_InvalidJSON(t *testing.T) {
	t.Log("Testing fetchDataFromAPI: handles invalid JSON response")
	stockRepo := &MockStockRepository{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	service := NewService(stockRepo, server.URL, "test-token")

	response, err := service.fetchDataFromAPI(context.Background(), nil)

	assert.Error(t, err)
	assert.Nil(t, response)

	var appErr *apperrors.AppError
	assert.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperrors.ErrCodeUpstreamAPI, appErr.Code)
}

func TestMakeRequestWithRetry_Success(t *testing.T) {
	t.Log("Testing makeRequestWithRetry: success on first attempt")
	stockRepo := &MockStockRepository{}

	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer server.Close()

	service := NewService(stockRepo, server.URL, "test-token")

	req, err := http.NewRequest("GET", server.URL, nil)
	require.NoError(t, err)

	resp, err := service.makeRequestWithRetry(context.Background(), req, 3)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 1, requestCount) // Should succeed on first try

	resp.Body.Close()
}

func TestMakeRequestWithRetry_SuccessAfterRetries(t *testing.T) {
	t.Log("Testing makeRequestWithRetry: success after a few retries")
	stockRepo := &MockStockRepository{}

	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if requestCount <= 2 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("server error"))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		}
	}))
	defer server.Close()

	service := NewService(stockRepo, server.URL, "test-token")

	req, err := http.NewRequest("GET", server.URL, nil)
	require.NoError(t, err)

	resp, err := service.makeRequestWithRetry(context.Background(), req, 3)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 3, requestCount) // Should succeed on third try

	resp.Body.Close()
}

func TestMakeRequestWithRetry_MaxRetriesExceeded(t *testing.T) {
	t.Log("Testing makeRequestWithRetry: fails after max retries")
	stockRepo := &MockStockRepository{}

	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("server error"))
	}))
	defer server.Close()

	service := NewService(stockRepo, server.URL, "test-token")

	req, err := http.NewRequest("GET", server.URL, nil)
	require.NoError(t, err)

	resp, err := service.makeRequestWithRetry(context.Background(), req, 2)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 3, requestCount) // Initial attempt + 2 retries

	var appErr *apperrors.AppError
	assert.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperrors.ErrCodeUpstreamAPI, appErr.Code)
}

func TestMakeRequestWithRetry_NonRetryableError(t *testing.T) {
	t.Log("Testing makeRequestWithRetry: stops for non-retryable errors (e.g., 404)")
	stockRepo := &MockStockRepository{}

	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusBadRequest) // 4xx errors should not be retried
		w.Write([]byte("bad request"))
	}))
	defer server.Close()

	service := NewService(stockRepo, server.URL, "test-token")

	req, err := http.NewRequest("GET", server.URL, nil)
	require.NoError(t, err)

	resp, err := service.makeRequestWithRetry(context.Background(), req, 3)

	assert.NoError(t, err) // No error, but status code indicates failure
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Equal(t, 1, requestCount) // Should not retry 4xx errors

	resp.Body.Close()
}

func TestTransformAPIRatings_Success(t *testing.T) {
	t.Log("Testing transformAPIRatings: successful transformation")
	stockRepo := &MockStockRepository{}
	service := NewService(stockRepo, "test-url", "test-token")

	apiRatings := []domain.APIStockRating{
		{
			Ticker:     "AAPL",
			Company:    "Apple Inc.",
			Brokerage:  "Goldman Sachs",
			Action:     "upgraded by",
			RatingFrom: "Hold",
			RatingTo:   "Buy",
			TargetFrom: "150.00",
			TargetTo:   "180.00",
			Time:       "2023-12-01T10:30:00Z",
		},
		{
			Ticker:     "GOOGL",
			Company:    "Alphabet Inc.",
			Brokerage:  "Morgan Stanley",
			Action:     "initiated by",
			RatingFrom: "",
			RatingTo:   "Strong Buy",
			TargetFrom: "",
			TargetTo:   "200.00",
			Time:       "2023-12-01T09:00:00Z",
		},
	}

	ratings, err := service.transformAPIRatings(apiRatings)

	assert.NoError(t, err)
	assert.Len(t, ratings, 2)

	// Create a map to find ratings by ticker since order is not guaranteed
	ratingMap := make(map[string]domain.StockRating)
	for _, rating := range ratings {
		ratingMap[rating.Ticker] = rating
	}

	// Check AAPL rating
	aaplRating, exists := ratingMap["AAPL"]
	assert.True(t, exists)
	assert.Equal(t, "Apple Inc.", aaplRating.Company)
	assert.Equal(t, "Goldman Sachs", aaplRating.Brokerage)
	assert.Equal(t, "upgraded by", aaplRating.Action)
	assert.NotNil(t, aaplRating.RatingFrom)
	assert.Equal(t, "Hold", *aaplRating.RatingFrom)
	assert.Equal(t, "Buy", aaplRating.RatingTo)
	assert.NotNil(t, aaplRating.TargetFrom)
	assert.Equal(t, 150.0, *aaplRating.TargetFrom)
	assert.NotNil(t, aaplRating.TargetTo)
	assert.Equal(t, 180.0, *aaplRating.TargetTo)

	// Check GOOGL rating (with missing fields)
	googlRating, exists := ratingMap["GOOGL"]
	assert.True(t, exists)
	assert.Equal(t, "Alphabet Inc.", googlRating.Company)
	assert.Equal(t, "Morgan Stanley", googlRating.Brokerage)
	assert.Equal(t, "initiated by", googlRating.Action)
	assert.Equal(t, "Strong Buy", googlRating.RatingTo)
	assert.Nil(t, googlRating.RatingFrom)
	assert.Nil(t, googlRating.TargetFrom)
	assert.NotNil(t, googlRating.TargetTo)
	assert.Equal(t, 200.0, *googlRating.TargetTo)
}

func TestTransformAPIRatings_InvalidTime(t *testing.T) {
	t.Log("Testing transformAPIRatings: handles invalid time format")
	stockRepo := &MockStockRepository{}
	service := NewService(stockRepo, "test-url", "test-token")

	apiRatings := []domain.APIStockRating{
		{
			Ticker:    "AAPL",
			Company:   "Apple Inc.",
			Brokerage: "Goldman Sachs",
			Action:    "upgraded by",
			RatingTo:  "Buy",
			Time:      "invalid-time-format",
		},
	}

	ratings, err := service.transformAPIRatings(apiRatings)

	assert.Error(t, err)
	assert.Nil(t, ratings)
	assert.Contains(t, err.Error(), "failed to parse time")
}

func TestTransformAPIRatings_InvalidTargetPrice(t *testing.T) {
	t.Log("Testing transformAPIRatings: handles invalid target price")
	stockRepo := &MockStockRepository{}
	service := NewService(stockRepo, "test-url", "test-token")

	apiRatings := []domain.APIStockRating{
		{
			Ticker:     "AAPL",
			Company:    "Apple Inc.",
			Brokerage:  "Goldman Sachs",
			Action:     "upgraded by",
			RatingTo:   "Buy",
			TargetFrom: "invalid-price",
			Time:       "2023-12-01T10:30:00Z",
		},
	}

	ratings, err := service.transformAPIRatings(apiRatings)

	// The function should succeed but skip the invalid target price
	assert.NoError(t, err)
	assert.Len(t, ratings, 1)
	assert.Nil(t, ratings[0].TargetFrom) // Invalid price should be skipped
	assert.Equal(t, "AAPL", ratings[0].Ticker)
}

func TestEnrichStockData_NotImplemented(t *testing.T) {
	t.Log("Testing EnrichStockData: confirms it runs without error (as it's not implemented)")
	stockRepo := &MockStockRepository{}
	service := NewService(stockRepo, "test-url", "test-token")

	err := service.EnrichStockData(context.Background(), []string{"AAPL", "GOOGL"})

	// The method should return nil as it's not implemented yet
	assert.NoError(t, err)
}

// Benchmark tests for expensive operations
func BenchmarkIngestAllData(b *testing.B) {
	b.Log("Benchmarking IngestAllData with 1000 items")
	stockRepo := &MockStockRepository{}

	// Create large dataset
	items := createMockAPIItems(1000)
	response := createMockAPIResponse(items, nil)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	service := NewService(stockRepo, server.URL, "test-token")

	// Mock repository to accept any batch
	stockRepo.On("CreateStockRatingsBatch", mock.Anything, mock.Anything).Return(1000, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := service.IngestAllData(context.Background())
		require.NoError(b, err)
	}
}

func BenchmarkTransformAPIRatings(b *testing.B) {
	b.Log("Benchmarking transformAPIRatings with 1000 items")
	stockRepo := &MockStockRepository{}
	service := NewService(stockRepo, "test-url", "test-token")

	// Create large dataset for transformation
	apiRatings := make([]domain.APIStockRating, 1000)
	for i := 0; i < 1000; i++ {
		apiRatings[i] = domain.APIStockRating{
			Ticker:     fmt.Sprintf("TICK%d", i),
			Company:    fmt.Sprintf("Company %d", i),
			Brokerage:  "Test Brokerage",
			Action:     "upgraded by",
			RatingFrom: "Hold",
			RatingTo:   "Buy",
			TargetFrom: "150.00",
			TargetTo:   "180.00",
			Time:       "2023-12-01T10:30:00Z",
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.transformAPIRatings(apiRatings)
		require.NoError(b, err)
	}
}

// Stress test for concurrent operations
func TestConcurrentIngestAllData(t *testing.T) {
	t.Log("Testing IngestAllData: with high concurrency")
	stockRepo := &MockStockRepository{}

	items := createMockAPIItems(10)
	response := createMockAPIResponse(items, nil)

	var requestCount int32 // Use atomic for race-free counter
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		time.Sleep(10 * time.Millisecond) // Simulate some processing time

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	service := NewService(stockRepo, server.URL, "test-token")

	// Mock repository to accept batches concurrently
	stockRepo.On("CreateStockRatingsBatch", mock.Anything, mock.Anything).Return(10, nil)

	concurrency := 5
	done := make(chan error, concurrency)

	// Run multiple ingestion operations concurrently
	for i := 0; i < concurrency; i++ {
		go func() {
			err := service.IngestAllData(context.Background())
			done <- err
		}()
	}

	// Wait for all operations to complete
	for i := 0; i < concurrency; i++ {
		err := <-done
		assert.NoError(t, err)
	}

	assert.Equal(t, int32(concurrency), atomic.LoadInt32(&requestCount))
	stockRepo.AssertExpectations(t)
}

// Edge case tests
func TestIngestAllData_VeryLargeResponse(t *testing.T) {
	t.Log("Testing IngestAllData: handles a very large API response (10,000 items)")
	stockRepo := &MockStockRepository{}

	// Create very large response
	items := createMockAPIItems(10000)
	response := createMockAPIResponse(items, nil)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	service := NewService(stockRepo, server.URL, "test-token")

	// Mock repository expectation
	stockRepo.On("CreateStockRatingsBatch", mock.Anything, mock.MatchedBy(func(ratings []*domain.StockRating) bool {
		return len(ratings) == 10000
	})).Return(10000, nil)

	err := service.IngestAllData(context.Background())

	assert.NoError(t, err)
	stockRepo.AssertExpectations(t)
}

func TestIngestAllData_SlowAPI(t *testing.T) {
	t.Log("Testing IngestAllData: handles a slow API response")
	stockRepo := &MockStockRepository{}

	items := createMockAPIItems(5)
	response := createMockAPIResponse(items, nil)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond) // Simulate slow API
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	service := NewService(stockRepo, server.URL, "test-token")

	// Mock repository expectation
	stockRepo.On("CreateStockRatingsBatch", mock.Anything, mock.Anything).Return(5, nil)

	start := time.Now()
	err := service.IngestAllData(context.Background())
	duration := time.Since(start)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, duration, 200*time.Millisecond)
	stockRepo.AssertExpectations(t)
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}
