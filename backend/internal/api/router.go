package api

import (
	"stock-analyzer/internal/domain"

	"github.com/gin-gonic/gin"
)

// SetupRouter creates and configures the HTTP router
func SetupRouter(stockRepo domain.StockRepository, ingestionSvc domain.IngestionService, recommendationSvc domain.RecommendationService, alpacaSvc domain.AlpacaService) *gin.Engine {
	// Create Gin router
	router := gin.New()

	// Add middleware
	router.Use(gin.Logger())
	router.Use(ErrorHandler())
	router.Use(CORS())

	// Create handlers
	handlers := NewHandlers(stockRepo, ingestionSvc, recommendationSvc, alpacaSvc)

	// Health check endpoint
	router.GET("/health", handlers.HealthCheck)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Stock ratings endpoints
		v1.GET("/ratings", handlers.GetStockRatings)
		v1.GET("/ratings/:ticker", handlers.GetStockRatingsByTicker)

		// Recommendations endpoint
		v1.GET("/recommendations", handlers.GetRecommendations)

		// Stock price data endpoints
		v1.GET("/stocks/:symbol/price", handlers.GetStockPrice)
		v1.GET("/stocks/:symbol/logo", handlers.GetStockLogo)

		// Admin/utility endpoints
		v1.POST("/ingest", handlers.TriggerIngestion)
	}

	return router
}
