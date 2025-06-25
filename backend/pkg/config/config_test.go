package config

import (
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad_WithDefaults(t *testing.T) {
	t.Log("Testing config Load: with default values")
	// Clear all environment variables for this test
	clearEnvVars()

	config := Load()

	assert.Equal(t, "8080", config.Port)
	assert.Equal(t, "", config.DatabaseURL)
	assert.Equal(t, "https://8j5baasof2.execute-api.us-west-2.amazonaws.com/production/swechallenge/list", config.StockAPIURL)
	assert.Equal(t, "", config.StockAPIToken)
	assert.Equal(t, "", config.AlphaVantageKey)
	assert.Equal(t, "", config.AlpacaAPIKey)
	assert.Equal(t, "", config.AlpacaAPISecret)
	assert.Equal(t, "development", config.Environment)
	assert.Equal(t, "info", config.LogLevel)
}

func TestLoad_WithEnvironmentVariables(t *testing.T) {
	t.Log("Testing config Load: with all environment variables set")
	// Set environment variables
	envVars := map[string]string{
		"PORT":              "3000",
		"DATABASE_URL":      "postgres://user:pass@localhost/db",
		"STOCK_API_URL":     "https://custom-api.com",
		"STOCK_API_TOKEN":   "custom-token",
		"ALPHA_VANTAGE_KEY": "av-key",
		"ALPACA_API_KEY":    "alpaca-key",
		"ALPACA_API_SECRET": "alpaca-secret",
		"ENVIRONMENT":       "production",
		"LOG_LEVEL":         "debug",
	}

	for key, value := range envVars {
		os.Setenv(key, value)
	}
	defer clearEnvVars()

	config := Load()

	assert.Equal(t, "3000", config.Port)
	assert.Equal(t, "postgres://user:pass@localhost/db", config.DatabaseURL)
	assert.Equal(t, "https://custom-api.com", config.StockAPIURL)
	assert.Equal(t, "custom-token", config.StockAPIToken)
	assert.Equal(t, "av-key", config.AlphaVantageKey)
	assert.Equal(t, "alpaca-key", config.AlpacaAPIKey)
	assert.Equal(t, "alpaca-secret", config.AlpacaAPISecret)
	assert.Equal(t, "production", config.Environment)
	assert.Equal(t, "debug", config.LogLevel)
}

func TestLoad_WithPartialEnvironmentVariables(t *testing.T) {
	t.Log("Testing config Load: with partial environment variables set")
	clearEnvVars()

	// Set only some environment variables
	os.Setenv("PORT", "9000")
	os.Setenv("ENVIRONMENT", "staging")
	defer clearEnvVars()

	config := Load()

	// Overridden values
	assert.Equal(t, "9000", config.Port)
	assert.Equal(t, "staging", config.Environment)

	// Default values for unset variables
	assert.Equal(t, "", config.DatabaseURL)
	assert.Equal(t, "https://8j5baasof2.execute-api.us-west-2.amazonaws.com/production/swechallenge/list", config.StockAPIURL)
	assert.Equal(t, "", config.StockAPIToken)
	assert.Equal(t, "info", config.LogLevel)
}

func TestLoad_WithEmptyEnvironmentVariables(t *testing.T) {
	t.Log("Testing config Load: with empty environment variables (should use defaults)")
	clearEnvVars()

	// Set environment variables to empty strings
	envVars := []string{
		"PORT", "DATABASE_URL", "STOCK_API_URL", "STOCK_API_TOKEN",
		"ALPHA_VANTAGE_KEY", "ALPACA_API_KEY", "ALPACA_API_SECRET",
		"ENVIRONMENT", "LOG_LEVEL",
	}

	for _, key := range envVars {
		os.Setenv(key, "")
	}
	defer clearEnvVars()

	config := Load()

	// Should use defaults when env vars are empty
	assert.Equal(t, "8080", config.Port)
	assert.Equal(t, "", config.DatabaseURL)
	assert.Equal(t, "https://8j5baasof2.execute-api.us-west-2.amazonaws.com/production/swechallenge/list", config.StockAPIURL)
	assert.Equal(t, "", config.StockAPIToken)
	assert.Equal(t, "", config.AlphaVantageKey)
	assert.Equal(t, "", config.AlpacaAPIKey)
	assert.Equal(t, "", config.AlpacaAPISecret)
	assert.Equal(t, "development", config.Environment)
	assert.Equal(t, "info", config.LogLevel)
}

func TestGetEnv_WithValue(t *testing.T) {
	t.Log("Testing utility: getEnv with a value set")
	key := "TEST_ENV_VAR"
	value := "test-value"
	fallback := "fallback-value"

	os.Setenv(key, value)
	defer os.Unsetenv(key)

	result := getEnv(key, fallback)
	assert.Equal(t, value, result)
}

func TestGetEnv_WithoutValue(t *testing.T) {
	t.Log("Testing utility: getEnv with no value set (should return default)")
	key := "NON_EXISTENT_ENV_VAR"
	fallback := "fallback-value"

	// Ensure the environment variable doesn't exist
	os.Unsetenv(key)

	result := getEnv(key, fallback)
	assert.Equal(t, fallback, result)
}

func TestGetEnv_WithEmptyValue(t *testing.T) {
	t.Log("Testing utility: getEnv with an empty value (should return default)")
	key := "EMPTY_ENV_VAR"
	fallback := "fallback-value"

	os.Setenv(key, "")
	defer os.Unsetenv(key)

	result := getEnv(key, fallback)
	assert.Equal(t, fallback, result)
}

func TestGetEnv_WithWhitespaceValue(t *testing.T) {
	t.Log("Testing utility: getEnv with a whitespace value (should return value)")
	key := "WHITESPACE_ENV_VAR"
	value := "  whitespace-value  "
	fallback := "fallback-value"

	os.Setenv(key, value)
	defer os.Unsetenv(key)

	result := getEnv(key, fallback)
	assert.Equal(t, value, result) // Should preserve whitespace
}

func TestGetEnvInt_WithValidInteger(t *testing.T) {
	t.Log("Testing utility: getEnvInt with a valid integer")
	key := "TEST_INT_VAR"
	value := "42"
	fallback := 10

	os.Setenv(key, value)
	defer os.Unsetenv(key)

	result := getEnvInt(key, fallback)
	assert.Equal(t, 42, result)
}

func TestGetEnvInt_WithInvalidInteger(t *testing.T) {
	t.Log("Testing utility: getEnvInt with an invalid integer (should return default)")
	key := "INVALID_INT_VAR"
	value := "not-a-number"
	fallback := 10

	os.Setenv(key, value)
	defer os.Unsetenv(key)

	result := getEnvInt(key, fallback)
	assert.Equal(t, fallback, result)
}

func TestGetEnvInt_WithoutValue(t *testing.T) {
	t.Log("Testing utility: getEnvInt with no value set (should return default)")
	key := "NON_EXISTENT_INT_VAR"
	fallback := 10

	os.Unsetenv(key)

	result := getEnvInt(key, fallback)
	assert.Equal(t, fallback, result)
}

func TestGetEnvInt_WithEmptyValue(t *testing.T) {
	t.Log("Testing utility: getEnvInt with an empty value (should return default)")
	key := "EMPTY_INT_VAR"
	fallback := 10

	os.Setenv(key, "")
	defer os.Unsetenv(key)

	result := getEnvInt(key, fallback)
	assert.Equal(t, fallback, result)
}

func TestGetEnvInt_WithZero(t *testing.T) {
	key := "ZERO_INT_VAR"
	value := "0"
	fallback := 10

	os.Setenv(key, value)
	defer os.Unsetenv(key)

	result := getEnvInt(key, fallback)
	assert.Equal(t, 0, result)
}

func TestGetEnvInt_WithNegativeInteger(t *testing.T) {
	key := "NEGATIVE_INT_VAR"
	value := "-42"
	fallback := 10

	os.Setenv(key, value)
	defer os.Unsetenv(key)

	result := getEnvInt(key, fallback)
	assert.Equal(t, -42, result)
}

func TestGetEnvInt_WithFloat(t *testing.T) {
	key := "FLOAT_VAR"
	value := "42.5"
	fallback := 10

	os.Setenv(key, value)
	defer os.Unsetenv(key)

	result := getEnvInt(key, fallback)
	assert.Equal(t, fallback, result) // Should use fallback for invalid int
}

func TestGetEnvInt_WithLargeInteger(t *testing.T) {
	key := "LARGE_INT_VAR"
	value := strconv.Itoa(int(^uint(0) >> 1)) // Max int
	fallback := 10

	os.Setenv(key, value)
	defer os.Unsetenv(key)

	result := getEnvInt(key, fallback)
	expected, _ := strconv.Atoi(value)
	assert.Equal(t, expected, result)
}

// Edge case tests
func TestLoad_MultipleCalls(t *testing.T) {
	clearEnvVars()

	// First call
	config1 := Load()

	// Set an environment variable
	os.Setenv("PORT", "9999")
	defer os.Unsetenv("PORT")

	// Second call should pick up the new environment variable
	config2 := Load()

	assert.Equal(t, "8080", config1.Port)
	assert.Equal(t, "9999", config2.Port)
}

func TestLoad_WithSpecialCharacters(t *testing.T) {
	clearEnvVars()

	specialValues := map[string]string{
		"DATABASE_URL":    "postgres://user:p@ssw0rd!@localhost:5432/db?sslmode=disable",
		"STOCK_API_TOKEN": "Bearer abc123!@#$%^&*()_+{}|:<>?[]\\;',./",
		"ALPACA_API_KEY":  "key-with-dashes-and_underscores.123",
		"ENVIRONMENT":     "test-environment_with.special-chars",
	}

	for key, value := range specialValues {
		os.Setenv(key, value)
	}
	defer clearEnvVars()

	config := Load()

	assert.Equal(t, specialValues["DATABASE_URL"], config.DatabaseURL)
	assert.Equal(t, specialValues["STOCK_API_TOKEN"], config.StockAPIToken)
	assert.Equal(t, specialValues["ALPACA_API_KEY"], config.AlpacaAPIKey)
	assert.Equal(t, specialValues["ENVIRONMENT"], config.Environment)
}

func TestLoad_WithUnicodeCharacters(t *testing.T) {
	clearEnvVars()

	unicodeValues := map[string]string{
		"ENVIRONMENT": "тест-среда",
		"LOG_LEVEL":   "デバッグ",
	}

	for key, value := range unicodeValues {
		os.Setenv(key, value)
	}
	defer clearEnvVars()

	config := Load()

	assert.Equal(t, unicodeValues["ENVIRONMENT"], config.Environment)
	assert.Equal(t, unicodeValues["LOG_LEVEL"], config.LogLevel)
}

// Performance tests
func BenchmarkLoad(b *testing.B) {
	clearEnvVars()

	// Set some environment variables
	os.Setenv("PORT", "8080")
	os.Setenv("DATABASE_URL", "postgres://user:pass@localhost/db")
	os.Setenv("ENVIRONMENT", "production")
	defer clearEnvVars()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Load()
	}
}

func BenchmarkGetEnv(b *testing.B) {
	key := "BENCHMARK_VAR"
	value := "benchmark-value"
	fallback := "fallback"

	os.Setenv(key, value)
	defer os.Unsetenv(key)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		getEnv(key, fallback)
	}
}

func BenchmarkGetEnvInt(b *testing.B) {
	key := "BENCHMARK_INT_VAR"
	value := "42"
	fallback := 10

	os.Setenv(key, value)
	defer os.Unsetenv(key)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		getEnvInt(key, fallback)
	}
}

// Stress tests
func TestLoad_HighFrequency(t *testing.T) {
	clearEnvVars()

	// Rapidly load configuration many times
	configs := make([]*Config, 1000)
	for i := 0; i < 1000; i++ {
		configs[i] = Load()
	}

	// All should be identical
	firstConfig := configs[0]
	for i, config := range configs {
		assert.Equal(t, firstConfig.Port, config.Port, "Config %d differs", i)
		assert.Equal(t, firstConfig.Environment, config.Environment, "Config %d differs", i)
		assert.Equal(t, firstConfig.LogLevel, config.LogLevel, "Config %d differs", i)
	}
}

func TestLoad_ConcurrentAccess(t *testing.T) {
	clearEnvVars()

	concurrency := 100
	done := make(chan *Config, concurrency)

	// Start many goroutines loading configuration
	for i := 0; i < concurrency; i++ {
		go func() {
			config := Load()
			done <- config
		}()
	}

	// Collect all results
	configs := make([]*Config, concurrency)
	for i := 0; i < concurrency; i++ {
		configs[i] = <-done
	}

	// All should be identical
	firstConfig := configs[0]
	for i, config := range configs {
		assert.Equal(t, firstConfig.Port, config.Port, "Config %d differs", i)
		assert.Equal(t, firstConfig.DatabaseURL, config.DatabaseURL, "Config %d differs", i)
		assert.Equal(t, firstConfig.Environment, config.Environment, "Config %d differs", i)
	}
}

// Test configuration validation scenarios
func TestConfig_DatabaseURL_Formats(t *testing.T) {
	clearEnvVars()

	testCases := []struct {
		name        string
		databaseURL string
	}{
		{"PostgreSQL URL", "postgres://user:pass@localhost:5432/dbname"},
		{"PostgreSQL URL with SSL", "postgres://user:pass@localhost:5432/dbname?sslmode=require"},
		{"MySQL URL", "mysql://user:pass@localhost:3306/dbname"},
		{"SQLite URL", "sqlite:///path/to/database.db"},
		{"Empty URL", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			os.Setenv("DATABASE_URL", tc.databaseURL)
			defer os.Unsetenv("DATABASE_URL")

			config := Load()
			assert.Equal(t, tc.databaseURL, config.DatabaseURL)
		})
	}
}

func TestConfig_Environment_Values(t *testing.T) {
	clearEnvVars()

	testCases := []string{
		"development",
		"staging",
		"production",
		"test",
		"local",
		"dev",
		"prod",
	}

	for _, env := range testCases {
		t.Run(env, func(t *testing.T) {
			os.Setenv("ENVIRONMENT", env)
			defer os.Unsetenv("ENVIRONMENT")

			config := Load()
			assert.Equal(t, env, config.Environment)
		})
	}
}

func TestConfig_LogLevel_Values(t *testing.T) {
	clearEnvVars()

	testCases := []string{
		"debug",
		"info",
		"warn",
		"error",
		"fatal",
		"trace",
	}

	for _, level := range testCases {
		t.Run(level, func(t *testing.T) {
			os.Setenv("LOG_LEVEL", level)
			defer os.Unsetenv("LOG_LEVEL")

			config := Load()
			assert.Equal(t, level, config.LogLevel)
		})
	}
}

// Helper function to clear all environment variables used by the config
func clearEnvVars() {
	envVars := []string{
		"PORT", "DATABASE_URL", "STOCK_API_URL", "STOCK_API_TOKEN",
		"ALPHA_VANTAGE_KEY", "ALPACA_API_KEY", "ALPACA_API_SECRET",
		"ENVIRONMENT", "LOG_LEVEL",
	}

	for _, key := range envVars {
		os.Unsetenv(key)
	}
}
