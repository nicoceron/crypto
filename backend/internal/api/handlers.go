package api

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"stock-analyzer/internal/alpaca"
	"stock-analyzer/internal/domain"
	apperrors "stock-analyzer/pkg/errors"

	"github.com/gin-gonic/gin"
)

// PriceBar represents a single price bar/candle
type PriceBar struct {
	Timestamp string  `json:"timestamp"`
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Volume    int64   `json:"volume"`
}

// StockPriceResponse represents the price data response
type StockPriceResponse struct {
	Symbol string     `json:"symbol"`
	Bars   []PriceBar `json:"bars"`
}

// StockLogoResponse represents the logo response
type StockLogoResponse struct {
	Symbol  string `json:"symbol"`
	LogoURL string `json:"logo_url"`
}

// Handlers contains all the HTTP handlers
type Handlers struct {
	stockRepo       domain.StockRepository
	ingestionSvc    domain.IngestionService
	recommendationSvc domain.RecommendationService
	alpacaSvc       *alpaca.Service
}

// NewHandlers creates a new handlers instance
func NewHandlers(stockRepo domain.StockRepository, ingestionSvc domain.IngestionService, recommendationSvc domain.RecommendationService, alpacaSvc *alpaca.Service) *Handlers {
	return &Handlers{
		stockRepo:       stockRepo,
		ingestionSvc:    ingestionSvc,
		recommendationSvc: recommendationSvc,
		alpacaSvc:       alpacaSvc,
	}
}

// GetStockPrice retrieves historical price data for a stock using Alpaca API
func (h *Handlers) GetStockPrice(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		HandleError(c, apperrors.ErrValidationFailure.WithDetails("symbol parameter is required"))
		return
	}

	// Convert symbol to uppercase for consistency
	symbol = strings.ToUpper(symbol)

	// Parse period parameter with enhanced logic for mini charts
	period := c.DefaultQuery("period", "1M")
	
	// For mini charts, we want more granular data
	var timeframe string
	var start time.Time
	end := time.Now()
	
	switch period {
	case "1W":
		start = end.AddDate(0, 0, -7)
		timeframe = "1Hour" // Hourly data for 1 week = ~168 data points (7 days * 24 hours)
		fmt.Printf("🔹 HANDLER: 1W period - using HOURLY data from %s\n", start.Format("2006-01-02"))
	case "1M":
		start = end.AddDate(0, -1, 0)
		timeframe = "1Hour" // Hourly data for 1 month = ~720 data points
		fmt.Printf("🔹 HANDLER: 1M period - using HOURLY data from %s\n", start.Format("2006-01-02"))
	case "3M":
		start = end.AddDate(0, -3, 0)
		timeframe = "1Day" // Daily data for 3 months = ~90 data points
		fmt.Printf("🔹 HANDLER: 3M period - using DAILY data from %s\n", start.Format("2006-01-02"))
	case "6M":
		start = end.AddDate(0, -6, 0)
		timeframe = "1Day" // Daily data for 6 months = ~180 data points
		fmt.Printf("🔹 HANDLER: 6M period - using DAILY data from %s\n", start.Format("2006-01-02"))
	case "1Y":
		start = end.AddDate(-1, 0, 0)
		timeframe = "1Day" // Daily data for 1 year = ~365 data points
		fmt.Printf("🔹 HANDLER: 1Y period - using DAILY data from %s\n", start.Format("2006-01-02"))
	case "2Y":
		start = end.AddDate(-2, 0, 0)
		timeframe = "1Day" // Daily data for 2 years = ~730 data points
		fmt.Printf("🔹 HANDLER: 2Y period - using DAILY data from %s\n", start.Format("2006-01-02"))
	default:
		start = end.AddDate(0, -1, 0)
		timeframe = "1Hour"
		fmt.Printf("🔹 HANDLER: Default period - using HOURLY data from %s\n", start.Format("2006-01-02"))
	}

	// Get price data with specified time range
	fmt.Printf("🔹 HANDLER: Final date range for %s period %s: %s to %s (%.0f days)\n", symbol, period, start.Format("2006-01-02"), end.Format("2006-01-02"), end.Sub(start).Hours()/24)
	alpacaBars, err := h.alpacaSvc.GetHistoricalBars(c.Request.Context(), symbol, timeframe, start, end)
	if err != nil {
		fmt.Printf("🔴 ERROR fetching data from Alpaca for %s: %v\n", symbol, err)
		// Return the raw error for debugging instead of wrapping it
		HandleError(c, err)
		return
	}

	// Convert Alpaca bars to our format
	bars := make([]PriceBar, len(alpacaBars))
	for i, alpacaBar := range alpacaBars {
		bars[i] = PriceBar{
			Timestamp: alpacaBar.Timestamp,
			Open:      alpacaBar.Open,
			High:      alpacaBar.High,
			Low:       alpacaBar.Low,
			Close:     alpacaBar.Close,
			Volume:    alpacaBar.Volume,
		}
	}

	// If no data from Alpaca, return error
	if len(bars) == 0 {
		fmt.Printf("No data returned from Alpaca for %s\n", symbol)
		HandleError(c, apperrors.ErrNotFound.WithDetails(fmt.Sprintf("No price data available for %s", symbol)))
		return
	}

	fmt.Printf("🔹 HANDLER: Returning %d bars for %s period %s (first: %s, last: %s)\n", len(bars), symbol, period, 
		func() string { if len(bars) > 0 { return bars[0].Timestamp } else { return "N/A" } }(),
		func() string { if len(bars) > 0 { return bars[len(bars)-1].Timestamp } else { return "N/A" } }())

	response := StockPriceResponse{
		Symbol: symbol,
		Bars:   bars,
	}

	c.JSON(http.StatusOK, response)
}

// GetStockLogo retrieves the logo URL for a stock
func (h *Handlers) GetStockLogo(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		fmt.Printf("🔴 ERROR in GetStockLogo: symbol parameter is required\n")
		HandleError(c, apperrors.ErrValidationFailure.WithDetails("symbol parameter is required"))
		return
	}

	// Convert to uppercase for consistency
	symbol = strings.ToUpper(symbol)
	fmt.Printf("🔍 HANDLER: Getting logo for symbol: %s\n", symbol)

	// Use Clearbit Logo API as primary source with fallbacks
	logoURL := fmt.Sprintf("https://logo.clearbit.com/%s.com", strings.ToLower(symbol))

	response := StockLogoResponse{
		Symbol:  symbol,
		LogoURL: logoURL,
	}

	// Add caching headers for logo responses (cache for 1 hour)
	c.Header("Cache-Control", "public, max-age=3600")
	c.Header("ETag", fmt.Sprintf(`"%s"`, symbol))
	
	fmt.Printf("✅ HANDLER: Successfully returning logo URL for %s: %s\n", symbol, logoURL)
	c.JSON(http.StatusOK, response)
}

// GetStockRatings retrieves paginated stock ratings with optional filtering
func (h *Handlers) GetStockRatings(c *gin.Context) {
	// Parse query parameters with defaults
	page, err := parseIntQuery(c, "page", 1)
	if err != nil {
		HandleError(c, apperrors.ErrValidationFailure.WithDetails("invalid page parameter"))
		return
	}

	limit, err := parseIntQuery(c, "limit", 20)
	if err != nil {
		HandleError(c, apperrors.ErrValidationFailure.WithDetails("invalid limit parameter"))
		return
	}

	// Validate limits
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	sortBy := c.DefaultQuery("sort_by", "time")
	order := c.DefaultQuery("order", "desc")
	search := c.Query("search")

	// Get ratings from repository
	fmt.Printf("🔍 HANDLER: Getting ratings - page:%d, limit:%d, sortBy:%s, order:%s, search:%s\n", page, limit, sortBy, order, search)
	ratings, totalCount, err := h.stockRepo.GetStockRatings(c.Request.Context(), page, limit, sortBy, order, search)
	if err != nil {
		fmt.Printf("🔴 ERROR in GetStockRatings: %v\n", err)
		HandleError(c, err)
		return
	}
	fmt.Printf("✅ HANDLER: Successfully retrieved %d ratings (total: %d)\n", len(ratings), totalCount)

	// Calculate pagination metadata
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

	response := domain.PaginatedResponse[domain.StockRating]{
		Data: ratings,
		Pagination: domain.Pagination{
			Page:       page,
			Limit:      limit,
			TotalItems: totalCount,
			TotalPages: totalPages,
		},
	}

	c.JSON(http.StatusOK, response)
}

// GetStockRatingsByTicker retrieves all ratings for a specific ticker
func (h *Handlers) GetStockRatingsByTicker(c *gin.Context) {
	ticker := c.Param("ticker")
	if ticker == "" {
		HandleError(c, apperrors.ErrValidationFailure.WithDetails("ticker parameter is required"))
		return
	}

	ratings, err := h.stockRepo.GetStockRatingsByTicker(c.Request.Context(), ticker)
	if err != nil {
		HandleError(c, err)
		return
	}

	if len(ratings) == 0 {
		HandleError(c, apperrors.ErrNotFound.WithDetails("no ratings found for ticker "+ticker))
		return
	}

	c.JSON(http.StatusOK, ratings)
}

// GetRecommendations retrieves stock recommendations
func (h *Handlers) GetRecommendations(c *gin.Context) {
	fmt.Printf("🔍 HANDLER: Starting GetRecommendations\n")
	recommendations, err := h.recommendationSvc.GetCachedRecommendations(c.Request.Context())
	if err != nil {
		fmt.Printf("🔴 ERROR in GetRecommendations: %v\n", err)
		HandleError(c, err)
		return
	}

	fmt.Printf("✅ HANDLER: Successfully retrieved %d recommendations\n", len(recommendations))
	c.JSON(http.StatusOK, recommendations)
}

// TriggerIngestion manually triggers a full data ingestion process
func (h *Handlers) TriggerIngestion(c *gin.Context) {
	// This should be an async operation in production
	go func() {
		if err := h.ingestionSvc.IngestAllData(c.Request.Context()); err != nil {
			// Log error in production
			// For now, just print to console
			println("Ingestion error:", err.Error())
		}
	}()

	c.JSON(http.StatusAccepted, gin.H{
		"message": "Data ingestion started",
		"status":  "accepted",
	})
}

// HealthCheck returns the health status of the service
func (h *Handlers) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "stock-analyzer",
		"timestamp": gin.H{"time": "now"},
	})
}

// parseIntQuery parses an integer query parameter with a default value
func parseIntQuery(c *gin.Context, key string, defaultValue int) (int, error) {
	str := c.Query(key)
	if str == "" {
		return defaultValue, nil
	}

	value, err := strconv.Atoi(str)
	if err != nil {
		return 0, err
	}

	return value, nil
} 