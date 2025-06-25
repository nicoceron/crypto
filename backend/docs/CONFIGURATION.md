# Configuration Documentation

This document provides comprehensive information about configuring the Stock Analyzer application across different environments and deployment scenarios.

## Table of Contents

1. [Environment Variables](#environment-variables)
2. [Configuration Structure](#configuration-structure)
3. [Environment-Specific Configs](#environment-specific-configs)
4. [Terraform Configuration](#terraform-configuration)
5. [Runtime Configuration](#runtime-configuration)
6. [Security Configuration](#security-configuration)
7. [Performance Tuning](#performance-tuning)
8. [Troubleshooting](#troubleshooting)

## Environment Variables

### Core Application Variables

| Variable       | Description                   | Required | Default       | Example                                                      |
| -------------- | ----------------------------- | -------- | ------------- | ------------------------------------------------------------ |
| `DATABASE_URL` | CockroachDB connection string | ✅       | -             | `postgresql://user:pass@host:26257/database?sslmode=require` |
| `PORT`         | Server port                   | ❌       | `8080`        | `8080`                                                       |
| `ENVIRONMENT`  | Deployment environment        | ❌       | `development` | `production`, `staging`, `development`                       |
| `LOG_LEVEL`    | Logging level                 | ❌       | `info`        | `debug`, `info`, `warn`, `error`                             |

### External API Configuration

| Variable            | Description                        | Required | Default       | Example                        |
| ------------------- | ---------------------------------- | -------- | ------------- | ------------------------------ |
| `ALPACA_API_KEY`    | Alpaca API key for market data     | ✅       | -             | `PKTEST...`                    |
| `ALPACA_API_SECRET` | Alpaca API secret                  | ✅       | -             | `abc123...`                    |
| `STOCK_API_URL`     | Stock ratings API endpoint         | ❌       | `https://...` | `https://api.example.com/data` |
| `STOCK_API_TOKEN`   | Stock ratings API token            | ✅       | -             | `token123...`                  |
| `ALPHA_VANTAGE_KEY` | Alpha Vantage API key (future use) | ❌       | -             | `ABCD1234`                     |

### AWS Lambda Configuration

| Variable        | Description          | Required | Default     | Example                         |
| --------------- | -------------------- | -------- | ----------- | ------------------------------- |
| `FUNCTION_TYPE` | Lambda function type | ❌       | `api`       | `api`, `ingestion`, `scheduler` |
| `AWS_REGION`    | AWS region           | ❌       | `us-west-2` | `us-west-2`, `us-east-1`        |

## Configuration Structure

### Go Configuration (`pkg/config/config.go`)

```go
// Config holds all configuration for our application
type Config struct {
    // Server configuration
    Port string

    // Database configuration
    DatabaseURL string

    // External API configuration
    StockAPIURL     string
    StockAPIToken   string
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
        StockAPIURL:     getEnv("STOCK_API_URL", "https://..."),
        StockAPIToken:   getEnv("STOCK_API_TOKEN", ""),
        AlphaVantageKey: getEnv("ALPHA_VANTAGE_KEY", ""),
        AlpacaAPIKey:    getEnv("ALPACA_API_KEY", ""),
        AlpacaAPISecret: getEnv("ALPACA_API_SECRET", ""),
        Environment:     getEnv("ENVIRONMENT", "development"),
        LogLevel:        getEnv("LOG_LEVEL", "info"),
    }
}
```

### Configuration Validation

```go
// ValidateConfig checks if all required configuration is present
func (c *Config) Validate() error {
    var errors []string

    if c.DatabaseURL == "" {
        errors = append(errors, "DATABASE_URL is required")
    }

    if c.AlpacaAPIKey == "" {
        errors = append(errors, "ALPACA_API_KEY is required")
    }

    if c.AlpacaAPISecret == "" {
        errors = append(errors, "ALPACA_API_SECRET is required")
    }

    if c.StockAPIToken == "" {
        errors = append(errors, "STOCK_API_TOKEN is required")
    }

    if len(errors) > 0 {
        return fmt.Errorf("configuration validation failed: %s", strings.Join(errors, ", "))
    }

    return nil
}
```

## Environment-Specific Configs

### Development Environment

**File**: `.env.development`

```env
# Development Configuration
ENVIRONMENT=development
LOG_LEVEL=debug
PORT=8080

# Database
DATABASE_URL=postgresql://user:password@localhost:26257/stock_data_dev?sslmode=require

# External APIs (use test/sandbox keys)
ALPACA_API_KEY=PKTEST_sandbox_key
ALPACA_API_SECRET=sandbox_secret
STOCK_API_TOKEN=test_token

# Optional services
ALPHA_VANTAGE_KEY=demo_key
```

### Staging Environment

**File**: `.env.staging`

```env
# Staging Configuration
ENVIRONMENT=staging
LOG_LEVEL=info
PORT=8080

# Database (staging cluster)
DATABASE_URL=postgresql://user:password@staging-cluster:26257/stock_data_staging?sslmode=require

# External APIs (staging/test keys)
ALPACA_API_KEY=PKTEST_staging_key
ALPACA_API_SECRET=staging_secret
STOCK_API_TOKEN=staging_token

# Performance settings
DB_MAX_OPEN_CONNS=15
DB_MAX_IDLE_CONNS=5
```

### Production Environment

**File**: `.env.production` (managed via AWS Secrets Manager)

```env
# Production Configuration
ENVIRONMENT=production
LOG_LEVEL=warn
PORT=8080

# Database (production cluster)
DATABASE_URL=postgresql://user:password@prod-cluster:26257/stock_data?sslmode=require

# External APIs (production keys)
ALPACA_API_KEY=PKPROD_live_key
ALPACA_API_SECRET=live_secret
STOCK_API_TOKEN=production_token

# Performance settings
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=10
```

## Terraform Configuration

### Variable Definitions (`terraform/variables.tf`)

```hcl
variable "aws_region" {
  description = "AWS region where resources will be created"
  type        = string
  default     = "us-west-2"
}

variable "environment" {
  description = "Environment name (e.g., dev, staging, prod)"
  type        = string
  default     = "dev"
}

variable "cockroachdb_connection_string" {
  description = "CockroachDB connection string"
  type        = string
  sensitive   = true
}

variable "alpaca_api_key" {
  description = "Alpaca API key"
  type        = string
  sensitive   = true
}

variable "alpaca_api_secret" {
  description = "Alpaca API secret"
  type        = string
  sensitive   = true
}
```

### Environment-Specific Terraform Files

**Development**: `terraform/dev.tfvars`

```hcl
# Development Infrastructure
aws_region = "us-west-2"
environment = "dev"
project_name = "stock-analyzer"

# Network Configuration
vpc_cidr = "10.0.0.0/16"
availability_zones = ["us-west-2a", "us-west-2b"]

# Lambda Configuration
lambda_memory_size = 512
lambda_timeout = 30

# Database Configuration
cockroachdb_connection_string = "postgresql://dev_user:dev_pass@dev-cluster:26257/stock_data_dev"

# API Keys (development/sandbox)
alpaca_api_key = "PKTEST_dev_key"
alpaca_api_secret = "dev_secret"
stock_api_token = "dev_token"

# Monitoring
enable_detailed_monitoring = false
log_retention_days = 7

# Tags
common_tags = {
  Project     = "stock-analyzer"
  Environment = "dev"
  Owner       = "development-team"
  CostCenter  = "engineering"
}
```

**Production**: `terraform/prod.tfvars`

```hcl
# Production Infrastructure
aws_region = "us-west-2"
environment = "prod"
project_name = "stock-analyzer"

# Network Configuration
vpc_cidr = "10.1.0.0/16"
availability_zones = ["us-west-2a", "us-west-2b", "us-west-2c"]

# Lambda Configuration
lambda_memory_size = 1024
lambda_timeout = 30
enable_provisioned_concurrency = true

# Database Configuration
cockroachdb_connection_string = "postgresql://prod_user:prod_pass@prod-cluster:26257/stock_data"

# API Keys (production)
alpaca_api_key = "PKPROD_live_key"
alpaca_api_secret = "live_secret"
stock_api_token = "production_token"

# Monitoring
enable_detailed_monitoring = true
log_retention_days = 30
enable_x_ray_tracing = true

# Performance
api_gateway_caching_enabled = true
cloudfront_cache_ttl = 3600

# Security
enable_waf = true
enable_shield = true

# Tags
common_tags = {
  Project     = "stock-analyzer"
  Environment = "prod"
  Owner       = "platform-team"
  CostCenter  = "trading"
  Compliance  = "required"
}
```

## Runtime Configuration

### Database Connection Pool Settings

```go
// Production optimized settings
func ConfigureDatabase(db *sql.DB, env string) {
    switch env {
    case "production":
        db.SetMaxOpenConns(25)
        db.SetMaxIdleConns(10)
        db.SetConnMaxLifetime(5 * time.Minute)
        db.SetConnMaxIdleTime(1 * time.Minute)
    case "staging":
        db.SetMaxOpenConns(15)
        db.SetMaxIdleConns(5)
        db.SetConnMaxLifetime(3 * time.Minute)
        db.SetConnMaxIdleTime(30 * time.Second)
    default: // development
        db.SetMaxOpenConns(5)
        db.SetMaxIdleConns(2)
        db.SetConnMaxLifetime(1 * time.Minute)
        db.SetConnMaxIdleTime(30 * time.Second)
    }
}
```

### Logging Configuration

```go
// ConfigureLogging sets up structured logging based on environment
func ConfigureLogging(env, level string) *logrus.Logger {
    logger := logrus.New()

    // Set log level
    switch level {
    case "debug":
        logger.SetLevel(logrus.DebugLevel)
    case "info":
        logger.SetLevel(logrus.InfoLevel)
    case "warn":
        logger.SetLevel(logrus.WarnLevel)
    case "error":
        logger.SetLevel(logrus.ErrorLevel)
    default:
        logger.SetLevel(logrus.InfoLevel)
    }

    // Set formatter based on environment
    if env == "production" {
        logger.SetFormatter(&logrus.JSONFormatter{
            TimestampFormat: time.RFC3339,
        })
    } else {
        logger.SetFormatter(&logrus.TextFormatter{
            FullTimestamp: true,
            ForceColors:   true,
        })
    }

    return logger
}
```

### HTTP Client Configuration

```go
// ConfigureHTTPClient creates optimized HTTP clients for external APIs
func ConfigureHTTPClient(env string) *http.Client {
    transport := &http.Transport{
        MaxIdleConns:       100,
        IdleConnTimeout:    90 * time.Second,
        DisableCompression: false,
    }

    var timeout time.Duration
    switch env {
    case "production":
        timeout = 30 * time.Second
        transport.MaxIdleConnsPerHost = 20
    case "staging":
        timeout = 45 * time.Second
        transport.MaxIdleConnsPerHost = 10
    default:
        timeout = 60 * time.Second
        transport.MaxIdleConnsPerHost = 5
    }

    return &http.Client{
        Timeout:   timeout,
        Transport: transport,
    }
}
```

## Security Configuration

### API Key Management

**Development**:

```bash
# Store in .env file (not committed)
export ALPACA_API_KEY="PKTEST_sandbox_key"
export ALPACA_API_SECRET="sandbox_secret"
```

**Production**:

```bash
# Use AWS Secrets Manager
aws secretsmanager create-secret \
  --name "stock-analyzer/prod/alpaca" \
  --description "Alpaca API credentials for production" \
  --secret-string '{"api_key":"PKPROD_key","api_secret":"secret"}'
```

### Database Security

```yaml
# CockroachDB Connection Security
Connection String Components:
  - SSL Mode: require (always use SSL)
  - Certificate Verification: full
  - Connection Encryption: TLS 1.2+
  - Authentication: Username/Password or Certificate

Security Best Practices:
  - Rotate credentials quarterly
  - Use read-only users for reporting
  - Monitor connection patterns
  - Enable audit logging
```

### AWS IAM Configuration

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "arn:aws:logs:*:*:*"
    },
    {
      "Effect": "Allow",
      "Action": ["secretsmanager:GetSecretValue"],
      "Resource": "arn:aws:secretsmanager:*:*:secret:stock-analyzer/*"
    }
  ]
}
```

## Performance Tuning

### Lambda Function Configuration

```yaml
API Function:
  Memory: 512MB (development) / 1024MB (production)
  Timeout: 30 seconds
  Reserved Concurrency: null (auto-scaling)
  Provisioned Concurrency: 2 (production only)

Ingestion Function:
  Memory: 1024MB
  Timeout: 15 minutes
  Reserved Concurrency: 1 (prevent duplicate runs)

Scheduler Function:
  Memory: 256MB
  Timeout: 5 minutes
  Reserved Concurrency: 1
```

### API Gateway Configuration

```yaml
Production Settings:
  Throttle Burst Limit: 5000
  Throttle Rate Limit: 2000
  Cache TTL: 300 seconds (5 minutes)
  Cache Key Parameters: symbol, period

Development Settings:
  Throttle Burst Limit: 200
  Throttle Rate Limit: 100
  Cache TTL: 60 seconds
  Cache Enabled: false
```

### CloudFront Configuration

```yaml
Production CDN:
  Default Cache Behavior:
    TTL: 3600 seconds (1 hour)
    Compress: true
    Viewer Protocol Policy: redirect-to-https

  Static Assets Cache:
    TTL: 86400 seconds (24 hours)
    Compress: true

  API Responses Cache:
    TTL: 300 seconds (5 minutes)
    Compress: true
    Cache Based on Headers: Authorization
```

## Troubleshooting

### Common Configuration Issues

#### 1. Database Connection Failures

**Error**: `connection refused` or `SSL required`

**Solutions**:

```bash
# Check connection string format
DATABASE_URL="postgresql://user:password@host:26257/database?sslmode=require"

# Verify SSL configuration
DATABASE_URL="postgresql://user:password@host:26257/database?sslmode=require&sslcert=client.crt&sslkey=client.key&sslrootcert=ca.crt"

# Test connection
psql "$DATABASE_URL" -c "SELECT 1;"
```

#### 2. API Key Authentication Failures

**Error**: `401 Unauthorized` or `403 Forbidden`

**Solutions**:

```bash
# Verify API key format
echo $ALPACA_API_KEY | grep -E "^PK(TEST|PROD)_"

# Check environment-specific keys
if [[ "$ENVIRONMENT" == "production" ]]; then
    # Should start with PKPROD_
    [[ "$ALPACA_API_KEY" =~ ^PKPROD_ ]]
else
    # Should start with PKTEST_
    [[ "$ALPACA_API_KEY" =~ ^PKTEST_ ]]
fi
```

#### 3. Lambda Environment Variable Issues

**Error**: Missing or incorrect environment variables

**Debugging**:

```go
// Add debugging to Lambda function
func logEnvironmentVariables() {
    envVars := []string{
        "DATABASE_URL", "ALPACA_API_KEY", "ALPACA_API_SECRET",
        "STOCK_API_TOKEN", "FUNCTION_TYPE", "ENVIRONMENT",
    }

    for _, envVar := range envVars {
        value := os.Getenv(envVar)
        if value == "" {
            log.Printf("WARNING: %s is not set", envVar)
        } else {
            // Log first/last 4 characters for sensitive values
            if strings.Contains(envVar, "KEY") || strings.Contains(envVar, "SECRET") || strings.Contains(envVar, "TOKEN") {
                masked := maskSensitiveValue(value)
                log.Printf("%s: %s", envVar, masked)
            } else {
                log.Printf("%s: %s", envVar, value)
            }
        }
    }
}

func maskSensitiveValue(value string) string {
    if len(value) <= 8 {
        return "****"
    }
    return value[:4] + "****" + value[len(value)-4:]
}
```

### Configuration Validation Script

```bash
#!/bin/bash
# validate-config.sh - Validate configuration before deployment

set -e

echo "🔍 Validating configuration..."

# Check required environment variables
required_vars=(
    "DATABASE_URL"
    "ALPACA_API_KEY"
    "ALPACA_API_SECRET"
    "STOCK_API_TOKEN"
)

for var in "${required_vars[@]}"; do
    if [[ -z "${!var}" ]]; then
        echo "❌ ERROR: $var is not set"
        exit 1
    else
        echo "✅ $var is set"
    fi
done

# Validate database connection
echo "🔗 Testing database connection..."
if psql "$DATABASE_URL" -c "SELECT 1;" &>/dev/null; then
    echo "✅ Database connection successful"
else
    echo "❌ Database connection failed"
    exit 1
fi

# Validate API keys format
echo "🔑 Validating API key formats..."
if [[ "$ALPACA_API_KEY" =~ ^PK(TEST|PROD)_ ]]; then
    echo "✅ Alpaca API key format is valid"
else
    echo "❌ Alpaca API key format is invalid"
    exit 1
fi

echo "🎉 Configuration validation completed successfully!"
```

### Environment-Specific Validation

```bash
# Run validation for specific environment
./scripts/validate-config.sh development
./scripts/validate-config.sh staging
./scripts/validate-config.sh production
```

---

_Configuration Documentation v1.0 - Last updated: December 2024_
