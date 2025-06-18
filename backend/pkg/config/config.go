package config

import (
	"os"
	"strconv"
)

// Config holds all configuration for our application
type Config struct {
	// Server configuration
	Port string
	
	// Database configuration
	DatabaseURL string
	
	// External API configuration
	StockAPIURL    string
	StockAPIToken  string
	AlphaVantageKey string
	
	// Alpaca API configuration
	AlpacaAPIKey    string
	AlpacaAPISecret string
	
	// Application configuration
	Environment string
	LogLevel    string
}

// Load reads configuration from environment variables
func Load() *Config {
	return &Config{
		Port:            getEnv("PORT", "8080"),
		DatabaseURL:     getEnv("DATABASE_URL", ""),
		StockAPIURL:     getEnv("STOCK_API_URL", "https://8j5baasof2.execute-api.us-west-2.amazonaws.com/production/swechallenge/list"),
		StockAPIToken:   getEnv("STOCK_API_TOKEN", ""),
		AlphaVantageKey: getEnv("ALPHA_VANTAGE_KEY", ""),
		AlpacaAPIKey:    getEnv("ALPACA_API_KEY", ""),
		AlpacaAPISecret: getEnv("ALPACA_API_SECRET", ""),
		Environment:     getEnv("ENVIRONMENT", "development"),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
	}
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// getEnvInt gets an environment variable as an integer with a fallback value
func getEnvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return fallback
} 