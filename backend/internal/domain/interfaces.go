package domain

import "context"

// StockRepository defines the contract for stock data persistence
type StockRepository interface {
	// CreateStockRating stores a new stock rating
	CreateStockRating(ctx context.Context, rating *StockRating) error
	
	// CreateStockRatingsBatch stores multiple stock ratings in a single transaction
	CreateStockRatingsBatch(ctx context.Context, ratings []StockRating) error
	
	// GetStockRatings retrieves paginated stock ratings with optional filtering
	GetStockRatings(ctx context.Context, page, limit int, sortBy, order, search string) ([]StockRating, int, error)
	
	// GetStockRatingsByTicker retrieves all ratings for a specific ticker
	GetStockRatingsByTicker(ctx context.Context, ticker string) ([]StockRating, error)
	
	// GetUniqueTickers retrieves all unique ticker symbols
	GetUniqueTickers(ctx context.Context) ([]string, error)
	
	// CreateEnrichedStockData stores enriched stock data
	CreateEnrichedStockData(ctx context.Context, data *EnrichedStockData) error
	
	// GetEnrichedStockData retrieves enriched data for a ticker
	GetEnrichedStockData(ctx context.Context, ticker string) (*EnrichedStockData, error)
	
	// GetLatestRatingsByTicker gets the most recent rating for each ticker
	GetLatestRatingsByTicker(ctx context.Context) (map[string]*StockRating, error)
}

// IngestionService defines the contract for data ingestion
type IngestionService interface {
	// IngestAllData fetches and stores all data from the external API
	IngestAllData(ctx context.Context) error
	
	// EnrichStockData fetches additional data for stocks from external sources
	EnrichStockData(ctx context.Context, tickers []string) error
}

// RecommendationService defines the contract for generating stock recommendations
type RecommendationService interface {
	// GenerateRecommendations analyzes data and generates stock recommendations
	GenerateRecommendations(ctx context.Context) ([]StockRecommendation, error)
	
	// GetCachedRecommendations retrieves the latest generated recommendations
	GetCachedRecommendations(ctx context.Context) ([]StockRecommendation, error)
} 