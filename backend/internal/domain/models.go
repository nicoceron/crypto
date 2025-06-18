package domain

import (
	"time"

	"github.com/google/uuid"
)

// StockRating represents a stock rating event from the API
type StockRating struct {
	RatingID    uuid.UUID  `json:"rating_id" db:"rating_id"`
	Ticker      string     `json:"ticker" db:"ticker"`
	Company     string     `json:"company" db:"company"`
	Brokerage   string     `json:"brokerage" db:"brokerage"`
	Action      string     `json:"action" db:"action"`
	RatingFrom  *string    `json:"rating_from" db:"rating_from"`
	RatingTo    string     `json:"rating_to" db:"rating_to"`
	TargetFrom  *float64   `json:"target_from" db:"target_from"`
	TargetTo    *float64   `json:"target_to" db:"target_to"`
	Time        time.Time  `json:"time" db:"time"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}

// EnrichedStockData represents additional data for recommendation analysis
type EnrichedStockData struct {
	Ticker           string                 `json:"ticker" db:"ticker"`
	HistoricalPrices map[string]interface{} `json:"historical_prices" db:"historical_prices"`
	NewsSentiment    map[string]interface{} `json:"news_sentiment" db:"news_sentiment"`
	UpdatedAt        time.Time              `json:"updated_at" db:"updated_at"`
}

// StockRecommendation represents a recommended stock
type StockRecommendation struct {
	Ticker           string    `json:"ticker"`
	Company          string    `json:"company"`
	Score            float64   `json:"score"`
	Rationale        string    `json:"rationale"`
	LatestRating     string    `json:"latest_rating"`
	TargetPrice      *float64  `json:"target_price"`
	TechnicalSignal  string    `json:"technical_signal"`
	SentimentScore   *float64  `json:"sentiment_score"`
	GeneratedAt      time.Time `json:"generated_at"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse[T any] struct {
	Data       []T        `json:"data"`
	Pagination Pagination `json:"pagination"`
}

// Pagination represents pagination metadata
type Pagination struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
}

// APIResponse represents the external API response format
type APIResponse struct {
	Items    []APIStockRating `json:"items"`
	NextPage *string          `json:"next_page"`
}

// APIStockRating represents the stock rating format from the external API
type APIStockRating struct {
	Ticker     string `json:"ticker"`
	Company    string `json:"company"`
	Brokerage  string `json:"brokerage"`
	Action     string `json:"action"`
	RatingFrom string `json:"rating_from"`
	RatingTo   string `json:"rating_to"`
	TargetFrom string `json:"target_from"`
	TargetTo   string `json:"target_to"`
	Time       string `json:"time"`
} 