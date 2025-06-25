# Data Ingestion Service Documentation

The Data Ingestion Service is responsible for fetching, transforming, and storing stock market data from external APIs. This document covers the service architecture, data flow, configuration, and operational procedures.

## Table of Contents

1. [Service Overview](#service-overview)
2. [Architecture](#architecture)
3. [Data Sources](#data-sources)
4. [Ingestion Pipeline](#ingestion-pipeline)
5. [Error Handling](#error-handling)
6. [Monitoring](#monitoring)
7. [Configuration](#configuration)
8. [Troubleshooting](#troubleshooting)

## Service Overview

### Purpose

The Ingestion Service continuously fetches stock market data from external APIs and stores it in our database for analysis and recommendations.

### Key Features

- **Scheduled Execution**: Runs every 4 hours via EventBridge
- **Data Validation**: Validates incoming data before storage
- **Duplicate Prevention**: Prevents duplicate records with unique constraints
- **Error Recovery**: Implements retry logic and error handling
- **Performance Optimization**: Batch processing and concurrent operations

### Service Implementation

```go
// Service interface
type IngestionService interface {
    IngestStockRatings(ctx context.Context) error
    GetIngestionStatus(ctx context.Context) (*IngestionStatus, error)
}

// Implementation
type Service struct {
    stockRepo   domain.StockRepository
    httpClient  *http.Client
    apiURL      string
    apiToken    string
    logger      *logrus.Logger
}
```

## Architecture

### Component Diagram

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   EventBridge   │───▶│ Lambda Function │───▶│ Ingestion Svc   │
│                 │    │                 │    │                 │
│ • Schedule      │    │ • Error Handling│    │ • Data Fetching │
│ • Trigger       │    │ • Logging       │    │ • Transformation│
│ • Rate: 4hrs    │    │ • Timeout: 15m  │    │ • Validation    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                                        │
┌─────────────────┐    ┌─────────────────┐             │
│   External API  │◀───│   HTTP Client   │◀────────────┘
│                 │    │                 │
│ • Stock Ratings │    │ • Retry Logic   │
│ • Pagination    │    │ • Rate Limiting │
│ • JSON Response │    │ • Timeout       │
└─────────────────┘    └─────────────────┘
                                │
┌─────────────────┐             │
│   CockroachDB   │◀────────────┘
│                 │
│ • Stock Ratings │
│ • Unique Index  │
│ • ACID Trans.   │
└─────────────────┘
```

### Data Flow

1. **Trigger**: EventBridge schedule triggers Lambda function
2. **Fetch**: Service makes paginated API calls to external source
3. **Transform**: Convert API response to domain models
4. **Validate**: Validate data integrity and business rules
5. **Store**: Batch insert into database with duplicate handling
6. **Log**: Record ingestion statistics and any errors

## Data Sources

### Primary Source: Stock Ratings API

**Endpoint**: `https://8j5baasof2.execute-api.us-west-2.amazonaws.com/production/swechallenge/list`

**Authentication**: Bearer token in Authorization header

**Response Format**:

```json
{
  "items": [
    {
      "ticker": "AAPL",
      "company": "Apple Inc.",
      "brokerage": "Goldman Sachs",
      "action": "upgrade",
      "rating_from": "Hold",
      "rating_to": "Buy",
      "target_from": "150.00",
      "target_to": "180.00",
      "time": "2024-12-24T08:30:00Z"
    }
  ],
  "next_page": "page_2_token"
}
```

**Pagination**: Uses `next_page` token for cursor-based pagination

### Data Transformation

```go
// Transform API response to domain model
func (s *Service) transformAPIRating(apiRating APIStockRating) (*domain.StockRating, error) {
    // Parse time
    parsedTime, err := time.Parse(time.RFC3339, apiRating.Time)
    if err != nil {
        return nil, fmt.Errorf("invalid time format: %w", err)
    }

    // Parse target prices
    var targetFrom, targetTo *float64
    if apiRating.TargetFrom != "" {
        if val, err := strconv.ParseFloat(apiRating.TargetFrom, 64); err == nil {
            targetFrom = &val
        }
    }
    if apiRating.TargetTo != "" {
        if val, err := strconv.ParseFloat(apiRating.TargetTo, 64); err == nil {
            targetTo = &val
        }
    }

    return &domain.StockRating{
        RatingID:   uuid.New(),
        Ticker:     strings.ToUpper(apiRating.Ticker),
        Company:    apiRating.Company,
        Brokerage:  apiRating.Brokerage,
        Action:     strings.ToLower(apiRating.Action),
        RatingFrom: nullableString(apiRating.RatingFrom),
        RatingTo:   apiRating.RatingTo,
        TargetFrom: targetFrom,
        TargetTo:   targetTo,
        Time:       parsedTime,
        CreatedAt:  time.Now(),
    }, nil
}
```

## Ingestion Pipeline

### 1. Data Fetching Process

```go
func (s *Service) IngestStockRatings(ctx context.Context) error {
    logger := s.logger.WithField("operation", "IngestStockRatings")

    startTime := time.Now()
    var totalFetched, totalStored int
    var errors []error

    logger.Info("Starting stock ratings ingestion")

    // Fetch all pages
    nextPage := ""
    for {
        ratings, nextPageToken, err := s.fetchStockRatingsPage(ctx, nextPage)
        if err != nil {
            errors = append(errors, err)
            break
        }

        if len(ratings) == 0 {
            break
        }

        totalFetched += len(ratings)

        // Transform and validate
        validRatings := s.validateAndTransformRatings(ratings)

        // Store in batches
        stored, err := s.storeRatingsBatch(ctx, validRatings)
        if err != nil {
            errors = append(errors, err)
            // Continue with next batch
        } else {
            totalStored += stored
        }

        // Check for next page
        if nextPageToken == "" {
            break
        }
        nextPage = nextPageToken

        // Rate limiting
        select {
        case <-time.After(100 * time.Millisecond):
        case <-ctx.Done():
            return ctx.Err()
        }
    }

    duration := time.Since(startTime)

    logger.WithFields(logrus.Fields{
        "total_fetched": totalFetched,
        "total_stored":  totalStored,
        "duration":      duration,
        "errors":        len(errors),
    }).Info("Stock ratings ingestion completed")

    if len(errors) > 0 {
        return fmt.Errorf("ingestion completed with %d errors: %v", len(errors), errors[0])
    }

    return nil
}
```

### 2. Batch Processing

```go
func (s *Service) storeRatingsBatch(ctx context.Context, ratings []domain.StockRating) (int, error) {
    if len(ratings) == 0 {
        return 0, nil
    }

    const batchSize = 100
    var totalStored int

    for i := 0; i < len(ratings); i += batchSize {
        end := i + batchSize
        if end > len(ratings) {
            end = len(ratings)
        }

        batch := ratings[i:end]
        stored, err := s.storeBatch(ctx, batch)
        if err != nil {
            s.logger.WithError(err).Errorf("Failed to store batch %d-%d", i, end)
            continue
        }

        totalStored += stored
    }

    return totalStored, nil
}

func (s *Service) storeBatch(ctx context.Context, ratings []domain.StockRating) (int, error) {
    tx, err := s.stockRepo.BeginTx(ctx)
    if err != nil {
        return 0, fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer tx.Rollback()

    var stored int
    for _, rating := range ratings {
        err := s.stockRepo.CreateStockRating(ctx, rating)
        if err != nil {
            if isDuplicateError(err) {
                // Skip duplicates
                continue
            }
            return stored, fmt.Errorf("failed to store rating: %w", err)
        }
        stored++
    }

    if err := tx.Commit(); err != nil {
        return 0, fmt.Errorf("failed to commit transaction: %w", err)
    }

    return stored, nil
}
```

### 3. Data Validation

```go
func (s *Service) validateRating(rating *domain.StockRating) error {
    var errors []string

    // Validate ticker
    if rating.Ticker == "" {
        errors = append(errors, "ticker is required")
    } else if !isValidTicker(rating.Ticker) {
        errors = append(errors, "invalid ticker format")
    }

    // Validate company
    if rating.Company == "" {
        errors = append(errors, "company is required")
    }

    // Validate brokerage
    if rating.Brokerage == "" {
        errors = append(errors, "brokerage is required")
    }

    // Validate action
    validActions := []string{"upgrade", "downgrade", "initiate", "maintain"}
    if !contains(validActions, rating.Action) {
        errors = append(errors, "invalid action")
    }

    // Validate rating_to
    if rating.RatingTo == "" {
        errors = append(errors, "rating_to is required")
    }

    // Validate time
    if rating.Time.IsZero() {
        errors = append(errors, "time is required")
    } else if rating.Time.After(time.Now()) {
        errors = append(errors, "time cannot be in the future")
    }

    // Validate target prices
    if rating.TargetFrom != nil && *rating.TargetFrom <= 0 {
        errors = append(errors, "target_from must be positive")
    }
    if rating.TargetTo != nil && *rating.TargetTo <= 0 {
        errors = append(errors, "target_to must be positive")
    }

    if len(errors) > 0 {
        return fmt.Errorf("validation failed: %s", strings.Join(errors, ", "))
    }

    return nil
}
```

## Error Handling

### Retry Strategy

```go
func (s *Service) fetchWithRetry(ctx context.Context, url string, maxRetries int) (*http.Response, error) {
    var lastErr error

    for attempt := 0; attempt <= maxRetries; attempt++ {
        if attempt > 0 {
            // Exponential backoff
            backoff := time.Duration(math.Pow(2, float64(attempt))) * time.Second
            s.logger.WithFields(logrus.Fields{
                "attempt": attempt,
                "backoff": backoff,
            }).Warn("Retrying request after error")

            select {
            case <-time.After(backoff):
            case <-ctx.Done():
                return nil, ctx.Err()
            }
        }

        resp, err := s.makeHTTPRequest(ctx, url)
        if err == nil {
            return resp, nil
        }

        lastErr = err

        // Don't retry client errors (4xx)
        if isClientError(err) {
            break
        }
    }

    return nil, fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}
```

### Error Classification

```go
type ErrorType int

const (
    ErrorTypeNetwork ErrorType = iota
    ErrorTypeTimeout
    ErrorTypeRateLimit
    ErrorTypeAuth
    ErrorTypeValidation
    ErrorTypeDatabase
)

func classifyError(err error) ErrorType {
    switch {
    case isNetworkError(err):
        return ErrorTypeNetwork
    case isTimeoutError(err):
        return ErrorTypeTimeout
    case isRateLimitError(err):
        return ErrorTypeRateLimit
    case isAuthError(err):
        return ErrorTypeAuth
    case isValidationError(err):
        return ErrorTypeValidation
    case isDatabaseError(err):
        return ErrorTypeDatabase
    default:
        return ErrorTypeNetwork
    }
}
```

### Circuit Breaker

```go
type CircuitBreaker struct {
    maxFailures int
    timeout     time.Duration
    state       State
    failures    int
    lastFailure time.Time
    mutex       sync.RWMutex
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
    cb.mutex.Lock()
    defer cb.mutex.Unlock()

    switch cb.state {
    case StateClosed:
        err := fn()
        if err != nil {
            cb.failures++
            cb.lastFailure = time.Now()
            if cb.failures >= cb.maxFailures {
                cb.state = StateOpen
            }
        } else {
            cb.failures = 0
        }
        return err

    case StateOpen:
        if time.Since(cb.lastFailure) > cb.timeout {
            cb.state = StateHalfOpen
            return cb.Execute(fn)
        }
        return errors.New("circuit breaker is open")

    case StateHalfOpen:
        err := fn()
        if err != nil {
            cb.state = StateOpen
            cb.failures++
            cb.lastFailure = time.Now()
        } else {
            cb.state = StateClosed
            cb.failures = 0
        }
        return err
    }

    return nil
}
```

## Monitoring

### Key Metrics

| Metric                   | Description                         | Alert Threshold              |
| ------------------------ | ----------------------------------- | ---------------------------- |
| `ingestion_duration`     | Time taken for full ingestion       | > 10 minutes                 |
| `ingestion_success_rate` | Percentage of successful ingestions | < 95%                        |
| `records_fetched`        | Number of records fetched per run   | < 100 (possible API issue)   |
| `records_stored`         | Number of records stored per run    | Deviation > 50% from average |
| `api_response_time`      | External API response time          | > 5 seconds                  |
| `duplicate_rate`         | Percentage of duplicate records     | > 20%                        |

### Logging Strategy

```go
func (s *Service) logIngestionMetrics(ctx context.Context, metrics IngestionMetrics) {
    s.logger.WithFields(logrus.Fields{
        "operation":       "ingestion_metrics",
        "duration":        metrics.Duration,
        "fetched":         metrics.RecordsFetched,
        "stored":          metrics.RecordsStored,
        "duplicates":      metrics.Duplicates,
        "errors":          metrics.Errors,
        "api_calls":       metrics.APICalls,
        "avg_response_time": metrics.AvgResponseTime,
    }).Info("Ingestion completed")

    // Send custom metrics to CloudWatch
    s.sendCloudWatchMetrics(ctx, metrics)
}
```

### CloudWatch Integration

```go
func (s *Service) sendCloudWatchMetrics(ctx context.Context, metrics IngestionMetrics) {
    // Custom metrics for monitoring
    metricData := []*cloudwatch.MetricDatum{
        {
            MetricName: aws.String("IngestionDuration"),
            Value:      aws.Float64(metrics.Duration.Seconds()),
            Unit:       aws.String("Seconds"),
            Timestamp:  aws.Time(time.Now()),
        },
        {
            MetricName: aws.String("RecordsFetched"),
            Value:      aws.Float64(float64(metrics.RecordsFetched)),
            Unit:       aws.String("Count"),
            Timestamp:  aws.Time(time.Now()),
        },
        {
            MetricName: aws.String("RecordsStored"),
            Value:      aws.Float64(float64(metrics.RecordsStored)),
            Unit:       aws.String("Count"),
            Timestamp:  aws.Time(time.Now()),
        },
    }

    // Send to CloudWatch (implementation depends on AWS SDK setup)
}
```

## Configuration

### Environment Variables

```env
# API Configuration
STOCK_API_URL=https://8j5baasof2.execute-api.us-west-2.amazonaws.com/production/swechallenge/list
STOCK_API_TOKEN=your_api_token

# Database
DATABASE_URL=postgresql://user:pass@host:26257/database

# Ingestion Settings
INGESTION_BATCH_SIZE=100
INGESTION_MAX_RETRIES=3
INGESTION_TIMEOUT=15m
INGESTION_RATE_LIMIT=100ms

# Monitoring
LOG_LEVEL=info
ENABLE_METRICS=true
```

### Service Configuration

```go
type Config struct {
    APIURL          string
    APIToken        string
    BatchSize       int
    MaxRetries      int
    Timeout         time.Duration
    RateLimit       time.Duration
    EnableMetrics   bool
}

func LoadConfig() *Config {
    return &Config{
        APIURL:        getEnv("STOCK_API_URL", ""),
        APIToken:      getEnv("STOCK_API_TOKEN", ""),
        BatchSize:     getEnvInt("INGESTION_BATCH_SIZE", 100),
        MaxRetries:    getEnvInt("INGESTION_MAX_RETRIES", 3),
        Timeout:       getEnvDuration("INGESTION_TIMEOUT", 15*time.Minute),
        RateLimit:     getEnvDuration("INGESTION_RATE_LIMIT", 100*time.Millisecond),
        EnableMetrics: getEnvBool("ENABLE_METRICS", true),
    }
}
```

## Troubleshooting

### Common Issues

#### 1. API Rate Limiting

**Symptoms**: 429 Too Many Requests errors

**Solutions**:

```go
// Implement adaptive rate limiting
func (s *Service) adaptiveRateLimit(responseHeaders http.Header) {
    if rateLimitRemaining := responseHeaders.Get("X-RateLimit-Remaining"); rateLimitRemaining != "" {
        if remaining, err := strconv.Atoi(rateLimitRemaining); err == nil && remaining < 10 {
            // Slow down requests
            s.rateLimit = s.rateLimit * 2
        }
    }
}
```

#### 2. Database Connection Issues

**Symptoms**: Connection timeouts, pool exhaustion

**Solutions**:

```go
// Implement connection pool monitoring
func (s *Service) monitorConnectionPool() {
    stats := s.db.Stats()
    if stats.OpenConnections >= stats.MaxOpenConnections-2 {
        s.logger.Warn("Database connection pool near capacity")
    }
}
```

#### 3. Memory Issues with Large Datasets

**Symptoms**: Out of memory errors, slow processing

**Solutions**:

```go
// Stream processing for large datasets
func (s *Service) streamProcess(ctx context.Context) error {
    const pageSize = 1000

    for page := 1; ; page++ {
        ratings, hasMore, err := s.fetchPage(ctx, page, pageSize)
        if err != nil {
            return err
        }

        if err := s.processBatch(ctx, ratings); err != nil {
            return err
        }

        if !hasMore {
            break
        }

        // Force garbage collection
        if page%10 == 0 {
            runtime.GC()
        }
    }

    return nil
}
```

### Debugging Tools

#### 1. Enable Debug Logging

```go
// Temporary debug mode
if os.Getenv("DEBUG_INGESTION") == "true" {
    s.logger.SetLevel(logrus.DebugLevel)
    s.logger.Debug("Debug mode enabled for ingestion")
}
```

#### 2. Manual Ingestion Trigger

```bash
# Trigger manual ingestion via API
curl -X POST "https://api.example.com/api/v1/ingest" \
  -H "Authorization: Bearer admin_token"

# Check ingestion status
curl -X GET "https://api.example.com/api/v1/ingest/status"
```

#### 3. Data Quality Checks

```sql
-- Check for recent ingestion
SELECT COUNT(*), MAX(created_at)
FROM stock_ratings
WHERE created_at > NOW() - INTERVAL '1 day';

-- Check for duplicates
SELECT ticker, brokerage, time, COUNT(*)
FROM stock_ratings
GROUP BY ticker, brokerage, time
HAVING COUNT(*) > 1;

-- Check data distribution
SELECT
    action,
    COUNT(*) as count,
    COUNT(*) * 100.0 / SUM(COUNT(*)) OVER() as percentage
FROM stock_ratings
WHERE created_at > NOW() - INTERVAL '1 day'
GROUP BY action;
```

---

_Ingestion Service Documentation v1.0 - Last updated: December 2024_
