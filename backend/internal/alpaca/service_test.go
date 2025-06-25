package alpaca

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestServer creates a new test service and a mock HTTP server.
// The service is configured to send requests to the mock server.
func setupTestServer(t *testing.T, handler http.HandlerFunc) (*Service, *httptest.Server) {
	server := httptest.NewServer(handler)
	// Use the test constructor to point the service to the mock server
	service := newTestService(server.URL)
	return service, server
}

func TestNewService(t *testing.T) {
	t.Log("Testing NewService: initialization")
	service := NewService("test-key", "test-secret")
	assert.NotNil(t, service)
	assert.NotNil(t, service.client)
	assert.NotNil(t, service.rateLimiter)
	assert.Equal(t, 250*time.Millisecond, service.rateLimiter.delay)
}

func TestRateLimiter_Wait(t *testing.T) {
	t.Log("Testing RateLimiter: ensures delay between calls")
	delay := 50 * time.Millisecond
	rateLimiter := NewRateLimiter(delay)

	// First call should have no delay
	start1 := time.Now()
	rateLimiter.Wait()
	assert.Less(t, time.Since(start1), 10*time.Millisecond)

	// Second call should be delayed
	start2 := time.Now()
	rateLimiter.Wait()
	assert.GreaterOrEqual(t, time.Since(start2), delay-10*time.Millisecond)
}

func TestParseTimeFrame(t *testing.T) {
	t.Log("Testing utility: parseTimeFrame")
	service := NewService("test-key", "test-secret")

	assert.Equal(t, "1Min", service.parseTimeFrame("1Min").String())
	assert.Equal(t, "1Day", service.parseTimeFrame("1Day").String())
	assert.Equal(t, "1Day", service.parseTimeFrame("invalid").String()) // Default
}

func TestGetHistoricalBars_Success(t *testing.T) {
	t.Log("Testing GetHistoricalBars: successful data retrieval")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/stocks/bars", r.URL.Path)
		assert.Equal(t, "AAPL", r.URL.Query().Get("symbols"))
		w.Header().Set("Content-Type", "application/json")
		// This is the raw JSON response the Alpaca v2 API sends for a single symbol
		rawJSON := `{
			"bars": {
				"AAPL": [
					{
						"t": "2023-01-01T10:00:00Z",
						"o": 150.0,
						"h": 151.0,
						"l": 149.0,
						"c": 150.5,
						"v": 100000,
						"n": 100,
						"vw": 150.2
					}
				]
			},
			"next_page_token": null
		}`
		fmt.Fprint(w, rawJSON)
	})

	service, server := setupTestServer(t, handler)
	defer server.Close()

	start := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)

	bars, err := service.GetHistoricalBars(context.Background(), "AAPL", "1Day", start, end)

	require.NoError(t, err)
	require.Len(t, bars, 1)
	assert.Equal(t, 150.5, bars[0].Close)
}

func TestGetHistoricalBars_APIError(t *testing.T) {
	t.Log("Testing GetHistoricalBars: handles API error")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, `{"message": "internal server error"}`)
	})

	service, server := setupTestServer(t, handler)
	defer server.Close()

	start := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)

	_, err := service.GetHistoricalBars(context.Background(), "AAPL", "1Day", start, end)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get bars from Alpaca")
}

func TestGetHistoricalBars_NoData(t *testing.T) {
	t.Log("Testing GetHistoricalBars: handles no data response")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// The v2 API returns an empty map in the "bars" field for no data
		rawJSON := `{"bars": {}, "next_page_token": null}`
		fmt.Fprint(w, rawJSON)
	})

	service, server := setupTestServer(t, handler)
	defer server.Close()

	start := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)

	bars, err := service.GetHistoricalBars(context.Background(), "AAPL", "1Day", start, end)

	require.Error(t, err)
	assert.Len(t, bars, 0)
	assert.Contains(t, err.Error(), "no bars found for symbol")
}

func TestGetSnapshot_Success(t *testing.T) {
	t.Log("Testing GetSnapshot: successful retrieval")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/stocks/snapshots", r.URL.Path)
		assert.Equal(t, "AAPL", r.URL.Query().Get("symbols"))
		w.Header().Set("Content-Type", "application/json")
		// The v2 API returns a map of symbol to snapshot object.
		rawJSON := `{
			"AAPL": {
				"latestTrade": { "t": "2023-01-01T10:00:00Z", "p": 150.0, "s": 100 },
				"latestQuote": { "t": "2023-01-01T10:00:00Z", "bp": 149.99, "bs": 10, "ap": 150.01, "as": 10 },
				"minuteBar": { "t": "2023-01-01T10:00:00Z", "o": 150.0, "h": 151.0, "l": 149.0, "c": 150.5, "v": 100000 },
				"dailyBar": { "t": "2023-01-01T05:00:00Z", "o": 148.0, "h": 152.0, "l": 147.0, "c": 150.5, "v": 2000000 },
				"prevDailyBar": { "t": "2022-12-30T05:00:00Z", "o": 147.0, "h": 148.0, "l": 146.0, "c": 147.5, "v": 1500000 }
			}
		}`
		fmt.Fprint(w, rawJSON)
	})

	service, server := setupTestServer(t, handler)
	defer server.Close()

	snapshot, err := service.GetSnapshot(context.Background(), "AAPL")

	require.NoError(t, err)
	require.NotNil(t, snapshot)
	assert.NotNil(t, snapshot.LatestTrade)
	assert.Equal(t, 150.0, snapshot.LatestTrade.Price)
}

func TestIsMarketHours(t *testing.T) {
	t.Log("Testing IsMarketHours: confirms it runs without panic")
	// This test is basic, just ensuring the method doesn't panic, as it has no side effects.
	service := NewService("any-key", "any-secret")
	result := service.IsMarketHours()
	assert.IsType(t, false, result)
}
