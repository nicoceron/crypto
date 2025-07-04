package domain

import (
	"time"

	"github.com/google/uuid"
)

// StockRating represents a stock rating event from the API.
// This is the core domain entity that captures analyst recommendations
// and rating changes for publicly traded stocks.
type StockRating struct {
	RatingID   uuid.UUID `json:"rating_id" db:"rating_id"`     // Unique identifier for this rating event
	Ticker     string    `json:"ticker" db:"ticker"`           // Stock symbol
	Company    string    `json:"company" db:"company"`         // Full company name
	Brokerage  string    `json:"brokerage" db:"brokerage"`     // Analyst firm name
	Action     string    `json:"action" db:"action"`           // Rating action
	RatingFrom *string   `json:"rating_from" db:"rating_from"` // Previous rating (nullable)
	RatingTo   string    `json:"rating_to" db:"rating_to"`     // New/current rating
	TargetFrom *float64  `json:"target_from" db:"target_from"` // Previous price target (nullable)
	TargetTo   *float64  `json:"target_to" db:"target_to"`     // New price target (nullable)
	Time       time.Time `json:"time" db:"time"`               // When the rating was issued
	CreatedAt  time.Time `json:"created_at" db:"created_at"`   // When this record was created
}

// EnrichedStockData represents additional data for recommendation analysis.
// This entity stores supplementary information beyond basic ratings,
// including historical price data and sentiment analysis results.
//
// The data is stored as JSON for flexibility in handling various
// external API response formats.
type EnrichedStockData struct {
	Ticker           string                 `json:"ticker" db:"ticker"`                       // Stock symbol
	HistoricalPrices map[string]interface{} `json:"historical_prices" db:"historical_prices"` // Price history from external APIs
	NewsSentiment    map[string]interface{} `json:"news_sentiment" db:"news_sentiment"`       // Sentiment analysis data
	UpdatedAt        time.Time              `json:"updated_at" db:"updated_at"`               // Last refresh timestamp
}

// StockRecommendation represents a recommended stock generated by our AI analysis.
// This combines analyst ratings, technical analysis, and sentiment data
// to produce actionable investment recommendations.
//
// Score ranges from 0.0 to 1.0, where:
// - 0.0-0.3: Avoid/Sell
// - 0.3-0.7: Neutral/Hold
// - 0.7-1.0: Buy/Strong Buy
type StockRecommendation struct {
	Ticker          string    `json:"ticker"`           // Stock symbol
	Company         string    `json:"company"`          // Full company name
	Score           float64   `json:"score"`            // Recommendation score (0.0-1.0)
	Rationale       string    `json:"rationale"`        // Human-readable explanation
	LatestRating    string    `json:"latest_rating"`    // Most recent analyst rating
	TargetPrice     *float64  `json:"target_price"`     // Analyst price target (nullable)
	TechnicalSignal string    `json:"technical_signal"` // Technical analysis result
	SentimentScore  *float64  `json:"sentiment_score"`  // News sentiment score (nullable)
	GeneratedAt     time.Time `json:"generated_at"`     // When this recommendation was generated
}

// PaginatedResponse represents a paginated API response.
// This generic type provides consistent pagination across all endpoints
// that return lists of data.
type PaginatedResponse[T any] struct {
	Data       []T        `json:"data"`       // The actual data items for this page
	Pagination Pagination `json:"pagination"` // Pagination metadata
}

// Pagination represents pagination metadata.
// Used consistently across all paginated endpoints to provide
// navigation information to clients.
type Pagination struct {
	Page       int `json:"page"`        // Current page number (1-based)
	Limit      int `json:"limit"`       // Items per page
	TotalItems int `json:"total_items"` // Total number of items across all pages
	TotalPages int `json:"total_pages"` // Total number of pages
}

// APIResponse represents the external API response format.
// This matches the structure returned by our external stock ratings API
// and is used during the data ingestion process.
type APIResponse struct {
	Items    []APIStockRating `json:"items"`     // Array of rating items from the API
	NextPage *string          `json:"next_page"` // Pagination token for next page (nullable)
}

// APIStockRating represents the stock rating format from the external API.
// This is the raw format we receive from external data sources before
// transformation into our internal StockRating domain model.
//
// All fields are strings as received from the API and require
// parsing/validation before use.
type APIStockRating struct {
	Ticker     string `json:"ticker"`      // Stock symbol as string
	Company    string `json:"company"`     // Company name as string
	Brokerage  string `json:"brokerage"`   // Analyst firm name as string
	Action     string `json:"action"`      // Rating action as string
	RatingFrom string `json:"rating_from"` // Previous rating as string
	RatingTo   string `json:"rating_to"`   // New rating as string
	TargetFrom string `json:"target_from"` // Previous target as string
	TargetTo   string `json:"target_to"`   // New target as string
	Time       string `json:"time"`        // Rating time as ISO string
}
