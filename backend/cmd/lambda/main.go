package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"stock-analyzer/internal/alpaca"
	"stock-analyzer/internal/api"
	"stock-analyzer/internal/ingestion"
	"stock-analyzer/internal/recommendation"
	"stock-analyzer/internal/storage"
	"stock-analyzer/pkg/config"
)

var (
	ginLambda *ginadapter.GinLambda
	db        *sql.DB
)

func init() {
	// Set Gin to release mode in Lambda
	gin.SetMode(gin.ReleaseMode)

	// Load configuration
	cfg := config.Load()

	// Initialize database connection
	var err error
	db, err = sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Initialize repositories
	stockRepo := storage.NewPostgresRepository(db)

	// Initialize services
	ingestionSvc := ingestion.NewService(stockRepo, cfg.StockAPIURL, cfg.StockAPIToken)
	recommendationSvc := recommendation.NewService(stockRepo)
	alpacaSvc := alpaca.NewService(cfg.AlpacaAPIKey, cfg.AlpacaAPISecret)

	// Setup router
	router := api.SetupRouter(stockRepo, ingestionSvc, recommendationSvc, alpacaSvc)

	// Create Lambda adapter
	ginLambda = ginadapter.New(router)
}

// Handler is the Lambda function handler
func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Handle different function types based on environment variable
	functionType := os.Getenv("FUNCTION_TYPE")
	
	switch functionType {
	case "ingestion":
		return handleIngestion(ctx, req)
	case "scheduler":
		return handleScheduler(ctx, req)
	default:
		// Default to API handler
		return ginLambda.ProxyWithContext(ctx, req)
	}
}

// handleIngestion handles background ingestion tasks
func handleIngestion(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Println("Starting data ingestion...")
	
	// Load configuration
	cfg := config.Load()
	
	// Initialize repositories
	stockRepo := storage.NewPostgresRepository(db)
	
	// Initialize ingestion service
	ingestionSvc := ingestion.NewService(stockRepo, cfg.StockAPIURL, cfg.StockAPIToken)
	
	// Perform ingestion
	err := ingestionSvc.IngestAllData(ctx)
	if err != nil {
		log.Printf("Ingestion failed: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       `{"error": "Ingestion failed"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}, nil
	}
	
	log.Println("Data ingestion completed successfully")
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       `{"message": "Ingestion completed successfully"}`,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}, nil
}

// handleScheduler handles scheduled tasks
func handleScheduler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Println("Running scheduled tasks...")
	
	// Add your scheduled task logic here
	// For example: cleanup old data, generate reports, etc.
	
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       `{"message": "Scheduled tasks completed"}`,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}, nil
}

func main() {
	lambda.Start(Handler)
} 