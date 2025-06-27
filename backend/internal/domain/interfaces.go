package domain

import (
	"context"
	"time"
)

// StockRepository defines the contract for stock data persistence.
// This interface abstracts the data layer, allowing for different storage
// implementations (PostgreSQL, MySQL, etc.) while maintaining consistent
// business logic interactions.
//
// All methods should handle context cancellation gracefully and return
// appropriate domain errors for business logic handling.
type StockRepository interface {
	// CreateStockRating stores a single stock rating in the database.
	// Returns an error if the rating already exists or if database constraints
	// are violated.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeouts
	//   - rating: The stock rating to store (must have valid RatingID)
	//
	// Returns:
	//   - error: nil on success, domain error on failure
	CreateStockRating(ctx context.Context, rating *StockRating) error

	// CreateStockRatingsBatch efficiently stores multiple stock ratings in a single transaction.
	// Implements duplicate handling by skipping ratings that already exist.
	// This method is optimized for bulk ingestion operations.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeouts  
	//   - ratings: Slice of ratings to store (empty slice is valid)
	//
	// Returns:
	//   - int: Number of ratings successfully inserted (excluding duplicates)
	//   - error: nil on success, domain error on failure
	CreateStockRatingsBatch(ctx context.Context, ratings []*StockRating) (int, error)

	// GetStockRatings retrieves paginated stock ratings with optional filtering and sorting.
	// Supports full-text search across ticker, company, and brokerage fields.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeouts
	//   - filters: Pagination, search, and sorting options
	//
	// Returns:
	//   - PaginatedResponse[StockRating]: Paginated results with metadata
	//   - error: nil on success, domain error on failure
	GetStockRatings(ctx context.Context, filters FilterOptions) (*PaginatedResponse[StockRating], error)

	// GetStockRatingsByTicker retrieves all ratings for a specific stock ticker.
	// Results are ordered by time descending (most recent first).
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeouts
	//   - ticker: Stock symbol (case-insensitive, e.g., "AAPL", "aapl")
	//
	// Returns:
	//   - []StockRating: All ratings for the ticker (empty slice if none found)
	//   - error: nil on success, domain error on failure
	GetStockRatingsByTicker(ctx context.Context, ticker string) ([]StockRating, error)

	// GetUniqueTickers retrieves all unique stock tickers that have ratings.
	// Useful for generating watchlists and ticker selection interfaces.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeouts
	//
	// Returns:
	//   - []string: Sorted list of unique ticker symbols
	//   - error: nil on success, domain error on failure
	GetUniqueTickers(ctx context.Context) ([]string, error)

	// CreateEnrichedStockData stores additional analysis data for a stock.
	// This supplements basic rating data with technical analysis, sentiment, etc.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeouts
	//   - data: Enriched data to store (JSON fields are flexible)
	//
	// Returns:
	//   - error: nil on success, domain error on failure
	CreateEnrichedStockData(ctx context.Context, data *EnrichedStockData) error

	// GetEnrichedStockData retrieves additional analysis data for a stock ticker.
	// Returns the most recent enriched data available.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeouts
	//   - ticker: Stock symbol to retrieve data for
	//
	// Returns:
	//   - *EnrichedStockData: The enriched data (nil if not found)
	//   - error: nil on success, domain error on failure
	GetEnrichedStockData(ctx context.Context, ticker string) (*EnrichedStockData, error)

	// GetLatestRatingsByTicker returns the most recent rating for each ticker.
	// This is optimized for recommendation generation where only the latest
	// analyst opinion matters.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeouts
	//
	// Returns:
	//   - map[string]*StockRating: Map of ticker -> latest rating
	//   - error: nil on success, domain error on failure
	GetLatestRatingsByTicker(ctx context.Context) (map[string]*StockRating, error)

	// DeleteOldEnrichedData removes enriched stock data records older than a given time.
	// Returns the number of records deleted.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeouts
	//   - olderThan: The timestamp threshold for deletion
	//
	// Returns:
	//   - int64: Number of records deleted
	//   - error: nil on success, domain error on failure
	DeleteOldEnrichedData(ctx context.Context, olderThan time.Time) (int64, error)
}

// IngestionService defines the contract for data ingestion from external APIs.
// This service handles fetching, transforming, and storing stock market data
// from various external sources with proper error handling and retry logic.
type IngestionService interface {
	// IngestAllData performs a complete data ingestion cycle.
	// Fetches all available data from external APIs, transforms it to our domain model,
	// and stores it in the database. Implements pagination, rate limiting, and
	// duplicate handling automatically.
	//
	// This method is typically called by scheduled jobs or manual triggers.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeouts (should have reasonable timeout)
	//
	// Returns:
	//   - error: nil on success, descriptive error on failure
	IngestAllData(ctx context.Context) error
}

// RecommendationService defines the contract for generating stock recommendations.
// This service combines analyst ratings, technical analysis, and sentiment data
// to produce actionable investment recommendations using AI/ML algorithms.
type RecommendationService interface {
	// GenerateRecommendations analyzes all available data and generates fresh stock recommendations.
	// This is a computationally expensive operation that should be called sparingly
	// (e.g., once per hour) and results should be cached.
	//
	// The algorithm considers:
	// - Recent analyst ratings and upgrades/downgrades
	// - Technical analysis signals
	// - News sentiment analysis
	// - Historical performance patterns
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeouts (should allow 30+ seconds)
	//
	// Returns:
	//   - []StockRecommendation: Ordered list of recommendations (best first)
	//   - error: nil on success, domain error on failure
	GenerateRecommendations(ctx context.Context) ([]StockRecommendation, error)

	// GetCachedRecommendations retrieves the latest generated recommendations from cache.
	// If cache is stale or empty, automatically triggers fresh generation.
	// This method is optimized for fast API responses.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeouts
	//
	// Returns:
	//   - []StockRecommendation: Cached or freshly generated recommendations
	//   - error: nil on success, domain error on failure
	GetCachedRecommendations(ctx context.Context) ([]StockRecommendation, error)
}

// PriceBar represents a single price bar/candle from market data.
// This structure is used for historical price data and technical analysis.
// All timestamps should be in UTC and follow ISO 8601 format.
type PriceBar struct {
	Timestamp string  `json:"timestamp"` // ISO 8601 timestamp in UTC (e.g., "2024-01-15T09:30:00Z")
	Open      float64 `json:"open"`      // Opening price for the period
	High      float64 `json:"high"`      // Highest price during the period
	Low       float64 `json:"low"`       // Lowest price during the period  
	Close     float64 `json:"close"`     // Closing price for the period
	Volume    int64   `json:"volume"`    // Number of shares traded during the period
}

// Snapshot represents current market snapshot data for real-time quotes.
// This provides the most recent market activity for a stock including
// trades, quotes, and intraday bars.
type Snapshot struct {
	Symbol       string    `json:"symbol"`                    // Stock symbol (e.g., "AAPL")
	LatestTrade  *Trade    `json:"latest_trade,omitempty"`    // Most recent trade (nullable)
	LatestQuote  *Quote    `json:"latest_quote,omitempty"`    // Most recent bid/ask quote (nullable)
	MinuteBar    *PriceBar `json:"minute_bar,omitempty"`      // Current minute bar (nullable)
	DailyBar     *PriceBar `json:"daily_bar,omitempty"`       // Current day's bar (nullable)
	PrevDailyBar *PriceBar `json:"prev_daily_bar,omitempty"`  // Previous day's bar for comparison (nullable)
}

// Trade represents a single trade execution.
// Contains the essential information about a completed stock transaction.
type Trade struct {
	Timestamp string  `json:"timestamp"` // ISO 8601 timestamp of the trade
	Price     float64 `json:"price"`     // Execution price per share
	Size      int64   `json:"size"`      // Number of shares traded
}

// Quote represents the current bid/ask spread for a stock.
// Shows the best available prices for buying and selling.
type Quote struct {
	Timestamp string  `json:"timestamp"` // ISO 8601 timestamp of the quote
	BidPrice  float64 `json:"bid_price"` // Highest price buyers are willing to pay
	AskPrice  float64 `json:"ask_price"` // Lowest price sellers are willing to accept
	BidSize   int64   `json:"bid_size"`  // Number of shares available at bid price
	AskSize   int64   `json:"ask_size"`  // Number of shares available at ask price
}

// AlpacaService defines the contract for Alpaca API interactions.
// Alpaca provides real-time and historical market data with rate limiting
// and authentication. All methods implement automatic retry logic and
// respect API rate limits.
type AlpacaService interface {
	// GetHistoricalBars fetches historical price data for technical analysis.
	// Supports various timeframes from minutes to months with automatic
	// pagination for large date ranges.
	GetHistoricalBars(ctx context.Context, symbol string, timeframe string, start, end time.Time) ([]PriceBar, error)

	// GetSnapshot fetches current market snapshot for real-time data.
	// Provides the most recent trade, quote, and bar data for a symbol.
	GetSnapshot(ctx context.Context, symbol string) (*Snapshot, error)

	// GetRecentBars fetches the most recent bars for a symbol.
	// Convenience method for getting latest price action without
	// specifying exact time ranges.
	GetRecentBars(ctx context.Context, symbol string) ([]PriceBar, error)

	// IsMarketHours checks if the US stock market is currently open.
	// Considers regular trading hours (9:30 AM - 4:00 PM ET) and
	// excludes weekends and market holidays.
	IsMarketHours() bool
}

// FilterOptions defines filtering and pagination options for data queries.
// Used consistently across repository methods to provide flexible data access.
type FilterOptions struct {
	Page     int    `json:"page"`      // Page number (1-based, default: 1)
	Limit    int    `json:"limit"`     // Items per page (default: 20, max: 100)
	Search   string `json:"search"`    // Search term for full-text search (optional)
	SortBy   string `json:"sort_by"`   // Field to sort by (default: "time")
	SortDesc bool   `json:"sort_desc"` // Sort direction (default: true for descending)
}
