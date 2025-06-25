// Package main provides the AWS Lambda entry point for the Stock Analyzer backend.
//
// This Lambda function serves multiple purposes based on the FUNCTION_TYPE environment variable:
//   - "api" (default): Handles HTTP requests via API Gateway using Gin router
//   - "ingestion": Performs scheduled data ingestion from external APIs
//   - "scheduler": Executes maintenance and cleanup tasks
//
// The function is designed to be deployed as a single Lambda with different configurations
// for different use cases, allowing for cost optimization and simplified deployment.
//
// Environment Variables Required:
//   - DATABASE_URL: PostgreSQL/CockroachDB connection string
//   - ALPACA_API_KEY: Alpaca API key for market data
//   - ALPACA_API_SECRET: Alpaca API secret
//   - STOCK_API_TOKEN: External stock ratings API token
//   - FUNCTION_TYPE: Optional, defaults to "api"
//
// AWS Lambda Configuration:
//   - Runtime: Go 1.x
//   - Memory: 512MB (API), 1024MB (ingestion), 256MB (scheduler)
//   - Timeout: 30s (API), 15min (ingestion), 5min (scheduler)
//   - Environment: Set via Terraform or AWS Console
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
	// ginLambda is the Gin adapter for AWS Lambda, initialized once during cold start
	ginLambda *ginadapter.GinLambda
	
	// db is the database connection pool, shared across Lambda invocations
	db *sql.DB
)

// init performs one-time initialization during Lambda cold start.
// This includes database connection setup, service initialization,
// and router configuration. The initialization is expensive but only
// happens once per Lambda container lifecycle.
func init() {
	// Set Gin to release mode in Lambda to reduce log verbosity
	gin.SetMode(gin.ReleaseMode)

	// Load configuration from environment variables
	cfg := config.Load()

	// Initialize database connection pool
	// The connection will be reused across Lambda invocations
	var err error
	db, err = sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Test database connectivity during initialization
	// This ensures we fail fast if database is unreachable
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Initialize repositories with database connection
	stockRepo := storage.NewPostgresRepository(db)

	// Initialize business services with their dependencies
	ingestionSvc := ingestion.NewService(stockRepo, cfg.StockAPIURL, cfg.StockAPIToken)
	recommendationSvc := recommendation.NewService(stockRepo)
	alpacaSvc := alpaca.NewAdapter(cfg.AlpacaAPIKey, cfg.AlpacaAPISecret)

	// Setup HTTP router with all handlers and middleware
	router := api.SetupRouter(stockRepo, ingestionSvc, recommendationSvc, alpacaSvc)

	// Create Lambda adapter for Gin router
	// This allows the Gin application to handle Lambda events
	ginLambda = ginadapter.New(router)
}

// Handler is the main AWS Lambda function handler.
// It routes requests to different handlers based on the FUNCTION_TYPE environment variable.
// This allows a single Lambda deployment to serve multiple purposes with different configurations.
//
// Function Types:
//   - "api": Handles HTTP API requests via API Gateway (default)
//   - "ingestion": Performs data ingestion from external APIs
//   - "scheduler": Executes scheduled maintenance tasks
//
// The handler implements the standard AWS Lambda signature and returns
// API Gateway-compatible responses for HTTP functions.
func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Determine function type from environment variable
	functionType := os.Getenv("FUNCTION_TYPE")

	switch functionType {
	case "ingestion":
		return handleIngestion(ctx, req)
	case "scheduler":
		return handleScheduler(ctx, req)
	default:
		// Default to API handler for HTTP requests
		return ginLambda.ProxyWithContext(ctx, req)
	}
}

// handleIngestion processes background data ingestion tasks.
// This function is triggered by EventBridge on a schedule (typically every 4 hours)
// to fetch fresh stock ratings data from external APIs.
//
// The function performs the following operations:
//   1. Initializes ingestion service with current configuration
//   2. Executes complete data ingestion cycle with error handling
//   3. Returns success/failure status for monitoring
//
// Expected Trigger: EventBridge scheduled event
// Timeout: 15 minutes (configurable via Lambda settings)
// Memory: 1024MB (higher memory for batch processing)
func handleIngestion(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Println("Starting data ingestion...")

	// Load fresh configuration for this invocation
	cfg := config.Load()

	// Initialize repositories and services
	// We reinitialize here to ensure fresh configuration
	stockRepo := storage.NewPostgresRepository(db)
	ingestionSvc := ingestion.NewService(stockRepo, cfg.StockAPIURL, cfg.StockAPIToken)

	// Perform complete data ingestion cycle
	// This includes fetching, transforming, and storing data
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

// handleScheduler processes scheduled maintenance and cleanup tasks.
// This function runs daily to perform housekeeping operations that
// keep the system running efficiently.
//
// Potential tasks include:
//   - Cleaning up old data beyond retention period
//   - Generating daily reports and analytics
//   - Updating cached recommendation data
//   - Database maintenance and optimization
//
// Expected Trigger: EventBridge scheduled event (daily)
// Timeout: 5 minutes
// Memory: 256MB (lightweight maintenance tasks)
func handleScheduler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Println("Running scheduled tasks...")

	// TODO: Implement scheduled task logic
	// Examples:
	// - Clean up old enriched data (older than 30 days)
	// - Update recommendation cache
	// - Generate daily analytics reports
	// - Perform database maintenance queries

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       `{"message": "Scheduled tasks completed"}`,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}, nil
}

// main is the Lambda entry point that starts the AWS Lambda runtime.
// This function is called by the AWS Lambda service when the function is invoked.
// It registers our Handler function with the Lambda runtime and begins
// processing incoming events.
func main() {
	lambda.Start(Handler)
}
