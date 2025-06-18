package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"stock-analyzer/internal/alpaca"
	"stock-analyzer/internal/api"
	"stock-analyzer/internal/ingestion"
	"stock-analyzer/internal/recommendation"
	"stock-analyzer/internal/storage"
	"stock-analyzer/pkg/config"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load configuration
	cfg := config.Load()

	// Set up database connection
	db, err := setupDatabase(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize repositories and services using dependency injection
	stockRepo := storage.NewPostgresRepository(db)
	ingestionSvc := ingestion.NewService(stockRepo, cfg.StockAPIURL, cfg.StockAPIToken)
	recommendationSvc := recommendation.NewService(stockRepo)

	// Initialize Alpaca service
	alpacaSvc := alpaca.NewService(cfg.AlpacaAPIKey, cfg.AlpacaAPISecret)
	log.Printf("Initialized Alpaca service with API key: %s****", cfg.AlpacaAPIKey[:4])

	// Setup HTTP router with all services
	router := api.SetupRouter(stockRepo, ingestionSvc, recommendationSvc, alpacaSvc)

	// Configure HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Perform initial data ingestion if database is empty
	go func() {
		if shouldRunInitialIngestion(stockRepo) {
			log.Println("Starting initial data ingestion...")
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			defer cancel()

			if err := ingestionSvc.IngestAllData(ctx); err != nil {
				log.Printf("Initial ingestion failed: %v", err)
			} else {
				log.Println("Initial data ingestion completed successfully")
				
				// Enrich data for a few popular tickers
				tickers, _ := stockRepo.GetUniqueTickers(ctx)
				if len(tickers) > 0 {
					// Limit to first 10 tickers for enrichment
					if len(tickers) > 10 {
						tickers = tickers[:10]
					}
					log.Printf("Enriching data for %d tickers...", len(tickers))
					if err := ingestionSvc.EnrichStockData(ctx, tickers); err != nil {
						log.Printf("Data enrichment failed: %v", err)
					} else {
						log.Println("Data enrichment completed")
					}
				}
			}
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Give outstanding requests a deadline for completion
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

// setupDatabase initializes the database connection
func setupDatabase(databaseURL string) (*sql.DB, error) {
	if databaseURL == "" {
		// For development, use a local CockroachDB instance
		databaseURL = "postgresql://root@localhost:26257/stock_data?sslmode=disable"
		log.Println("Using default database URL for development")
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Database connection established successfully")
	return db, nil
}

// shouldRunInitialIngestion checks if we need to run initial data ingestion
func shouldRunInitialIngestion(stockRepo *storage.PostgresRepository) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if we have any data in the database
	ratings, _, err := stockRepo.GetStockRatings(ctx, 1, 1, "time", "desc", "")
	if err != nil {
		log.Printf("Error checking for existing data: %v", err)
		return false
	}

	// If no data exists, run initial ingestion
	return len(ratings) == 0
} 