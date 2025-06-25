# Database Documentation

The Stock Analyzer uses **CockroachDB Serverless** (PostgreSQL-compatible) as its primary database. This document covers the database schema, migration strategy, indexing, performance optimization, and operational procedures.

## Table of Contents

1. [Database Overview](#database-overview)
2. [Schema Design](#schema-design)
3. [Database Tables](#database-tables)
4. [Indexes and Performance](#indexes-and-performance)
5. [Migration Strategy](#migration-strategy)
6. [Query Patterns](#query-patterns)
7. [Performance Optimization](#performance-optimization)
8. [Backup and Recovery](#backup-and-recovery)
9. [Monitoring](#monitoring)
10. [Troubleshooting](#troubleshooting)

## Database Overview

### Technology Stack

- **Database**: CockroachDB Serverless
- **Protocol**: PostgreSQL wire protocol
- **Driver**: `lib/pq` (PostgreSQL driver for Go)
- **Connection**: SSL/TLS encrypted connections
- **Scaling**: Auto-scaling based on workload

### Key Features

- **Distributed SQL**: Horizontal scaling across multiple nodes
- **ACID Transactions**: Full ACID compliance
- **PostgreSQL Compatibility**: Standard PostgreSQL SQL syntax
- **Serverless**: Automatic scaling and management
- **Global Distribution**: Multi-region deployment capability

## Schema Design

The database follows a **normalized relational design** with proper foreign key relationships and constraints:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   stock_ratings │    │ enriched_data   │    │recommendations  │
│                 │    │                 │    │                 │
│ • rating_id (PK)│    │ • ticker (PK)   │    │ • ticker        │
│ • ticker        │◄───┤ • hist_prices   │    │ • score         │
│ • company       │    │ • news_sentiment│    │ • rationale     │
│ • brokerage     │    │ • updated_at    │    │ • generated_at  │
│ • action        │    │                 │    │                 │
│ • rating_to     │    └─────────────────┘    └─────────────────┘
│ • target_to     │
│ • time          │
│ • created_at    │
└─────────────────┘
```

### Design Principles

1. **Normalization**: Tables are normalized to 3NF to eliminate data redundancy
2. **Referential Integrity**: Foreign key constraints ensure data consistency
3. **Temporal Data**: Proper handling of time-based data with appropriate indexes
4. **Scalability**: Schema designed for horizontal scaling
5. **Performance**: Optimized for read-heavy workloads with proper indexing

## Database Tables

### 1. `stock_ratings` Table

**Purpose**: Stores analyst stock ratings and recommendations

```sql
CREATE TABLE stock_ratings (
    rating_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticker VARCHAR(10) NOT NULL,
    company TEXT NOT NULL,
    brokerage VARCHAR(100) NOT NULL,
    action VARCHAR(20) NOT NULL,
    rating_from VARCHAR(20),
    rating_to VARCHAR(20) NOT NULL,
    target_from DECIMAL(10, 2),
    target_to DECIMAL(10, 2),
    time TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
```

**Fields Description**:

- `rating_id`: Unique identifier (UUID v4)
- `ticker`: Stock symbol (e.g., 'AAPL', 'MSFT')
- `company`: Company name (e.g., 'Apple Inc.')
- `brokerage`: Rating firm (e.g., 'Goldman Sachs')
- `action`: Rating action ('upgrade', 'downgrade', 'initiate', 'maintain')
- `rating_from`: Previous rating (nullable for new coverage)
- `rating_to`: New rating ('Buy', 'Hold', 'Sell', etc.)
- `target_from`: Previous price target (nullable)
- `target_to`: New price target (nullable)
- `time`: When the rating was issued by the analyst
- `created_at`: When the record was created in our system

**Constraints**:

```sql
-- Ensure valid ticker symbols
CONSTRAINT valid_ticker CHECK (ticker ~ '^[A-Z]{1,10}$'),

-- Ensure valid actions
CONSTRAINT valid_action CHECK (action IN ('upgrade', 'downgrade', 'initiate', 'maintain')),

-- Ensure target prices are positive
CONSTRAINT positive_target_from CHECK (target_from IS NULL OR target_from > 0),
CONSTRAINT positive_target_to CHECK (target_to IS NULL OR target_to > 0),

-- Ensure time is not in the future
CONSTRAINT rating_time_valid CHECK (time <= NOW())
```

### 2. `enriched_stock_data` Table

**Purpose**: Stores additional data for recommendation analysis

```sql
CREATE TABLE enriched_stock_data (
    ticker VARCHAR(10) PRIMARY KEY,
    historical_prices JSONB,
    news_sentiment JSONB,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
```

**Fields Description**:

- `ticker`: Stock symbol (primary key)
- `historical_prices`: JSON containing price history and technical indicators
- `news_sentiment`: JSON containing sentiment analysis data
- `updated_at`: Last update timestamp

**JSON Structure Examples**:

```json
// historical_prices format
{
  "prices": [
    {
      "date": "2024-12-20",
      "open": 150.25,
      "high": 152.80,
      "low": 149.90,
      "close": 151.45,
      "volume": 2500000
    }
  ],
  "indicators": {
    "sma_20": 148.50,
    "sma_50": 145.30,
    "rsi": 65.4,
    "macd": 1.2
  }
}

// news_sentiment format
{
  "overall_score": 0.75,
  "articles": [
    {
      "headline": "Company reports strong Q4 earnings",
      "sentiment": 0.85,
      "source": "Reuters",
      "date": "2024-12-20"
    }
  ],
  "summary": {
    "positive": 12,
    "neutral": 5,
    "negative": 3
  }
}
```

### 3. `stock_recommendations` Table (Materialized View)

**Purpose**: Cached AI-generated stock recommendations

```sql
CREATE TABLE stock_recommendations (
    ticker VARCHAR(10) PRIMARY KEY,
    company TEXT NOT NULL,
    score DECIMAL(3, 1) NOT NULL,
    rationale TEXT NOT NULL,
    latest_rating VARCHAR(20),
    target_price DECIMAL(10, 2),
    technical_signal VARCHAR(20) NOT NULL,
    sentiment_score DECIMAL(3, 2),
    generated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
```

**Fields Description**:

- `ticker`: Stock symbol (primary key)
- `company`: Company name
- `score`: Recommendation score (0.0 to 10.0)
- `rationale`: Human-readable explanation
- `latest_rating`: Most recent analyst rating
- `target_price`: Consensus price target
- `technical_signal`: 'bullish', 'bearish', or 'neutral'
- `sentiment_score`: News sentiment score (-1.0 to 1.0)
- `generated_at`: When recommendation was generated

## Indexes and Performance

### Primary Indexes

```sql
-- Primary key indexes (automatically created)
CREATE UNIQUE INDEX idx_stock_ratings_pkey ON stock_ratings (rating_id);
CREATE UNIQUE INDEX idx_enriched_data_pkey ON enriched_stock_data (ticker);
CREATE UNIQUE INDEX idx_recommendations_pkey ON stock_recommendations (ticker);
```

### Secondary Indexes

```sql
-- Optimize ticker-based queries
CREATE INDEX idx_stock_ratings_ticker ON stock_ratings (ticker);

-- Optimize time-based queries (for pagination and sorting)
CREATE INDEX idx_stock_ratings_time ON stock_ratings (time DESC);

-- Optimize combined ticker and time queries
CREATE INDEX idx_stock_ratings_ticker_time ON stock_ratings (ticker, time DESC);

-- Optimize action-based filtering
CREATE INDEX idx_stock_ratings_action ON stock_ratings (action);

-- Optimize brokerage-based filtering
CREATE INDEX idx_stock_ratings_brokerage ON stock_ratings (brokerage);

-- Composite index for common query patterns
CREATE INDEX idx_stock_ratings_composite ON stock_ratings (ticker, action, time DESC);

-- Optimize updated_at queries for enriched data
CREATE INDEX idx_enriched_data_updated ON enriched_stock_data (updated_at DESC);

-- Optimize score-based queries for recommendations
CREATE INDEX idx_recommendations_score ON stock_recommendations (score DESC);
```

### Index Usage Patterns

| Query Pattern                                  | Optimal Index                   |
| ---------------------------------------------- | ------------------------------- |
| `WHERE ticker = 'AAPL'`                        | `idx_stock_ratings_ticker`      |
| `ORDER BY time DESC`                           | `idx_stock_ratings_time`        |
| `WHERE ticker = 'AAPL' ORDER BY time DESC`     | `idx_stock_ratings_ticker_time` |
| `WHERE action = 'upgrade'`                     | `idx_stock_ratings_action`      |
| `WHERE ticker = 'AAPL' AND action = 'upgrade'` | `idx_stock_ratings_composite`   |

## Migration Strategy

### Migration Files Structure

```
migrations/
├── 001_initial_schema_cloud.sql     # Initial table creation
├── 002_add_unique_constraint.sql    # Add unique constraints
└── 003_add_unique_constraint_simple.sql # Simplified constraints
```

### Migration 001: Initial Schema

```sql
-- Create initial tables with basic structure
CREATE TABLE stock_ratings (
    rating_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticker VARCHAR(10) NOT NULL,
    company TEXT NOT NULL,
    -- ... other fields
);

-- Create initial indexes
CREATE INDEX idx_stock_ratings_ticker ON stock_ratings (ticker);
CREATE INDEX idx_stock_ratings_time ON stock_ratings (time DESC);
```

### Migration 002: Add Constraints

```sql
-- Add unique constraint to prevent duplicate ratings
ALTER TABLE stock_ratings
ADD CONSTRAINT unique_rating
UNIQUE (ticker, brokerage, time, action);

-- Add check constraints for data validation
ALTER TABLE stock_ratings
ADD CONSTRAINT valid_ticker
CHECK (ticker ~ '^[A-Z]{1,10}$');
```

### Migration 003: Optimization

```sql
-- Add composite indexes for better query performance
CREATE INDEX idx_stock_ratings_composite
ON stock_ratings (ticker, action, time DESC);

-- Add partial indexes for specific use cases
CREATE INDEX idx_recent_ratings
ON stock_ratings (ticker, time DESC)
WHERE time > NOW() - INTERVAL '30 days';
```

### Running Migrations

```bash
# Run all pending migrations
go run cmd/migrate/main.go

# Run migrations with specific database URL
export DATABASE_URL="postgresql://user:pass@host:26257/database"
go run cmd/migrate/main.go
```

## Query Patterns

### Common Query Examples

#### 1. Get Recent Ratings for a Ticker

```sql
SELECT rating_id, company, brokerage, action, rating_to, target_to, time
FROM stock_ratings
WHERE ticker = $1
ORDER BY time DESC
LIMIT $2 OFFSET $3;
```

**Index Used**: `idx_stock_ratings_ticker_time`

#### 2. Get Paginated Ratings with Filtering

```sql
SELECT rating_id, ticker, company, brokerage, action, rating_to, time
FROM stock_ratings
WHERE ($1::text IS NULL OR ticker = $1)
  AND ($2::text IS NULL OR action = $2)
ORDER BY time DESC
LIMIT $3 OFFSET $4;
```

**Index Used**: `idx_stock_ratings_composite` or `idx_stock_ratings_time`

#### 3. Get Recommendation Data

```sql
SELECT ticker, company, score, rationale, latest_rating, target_price
FROM stock_recommendations
ORDER BY score DESC
LIMIT $1;
```

**Index Used**: `idx_recommendations_score`

#### 4. Get Enriched Data for Analysis

```sql
SELECT ticker, historical_prices, news_sentiment
FROM enriched_stock_data
WHERE updated_at > NOW() - INTERVAL '1 hour'
ORDER BY updated_at DESC;
```

**Index Used**: `idx_enriched_data_updated`

### Query Performance Guidelines

1. **Always use parameterized queries** to prevent SQL injection
2. **Include LIMIT clauses** for pagination to prevent large result sets
3. **Use appropriate indexes** by checking query execution plans
4. **Avoid SELECT \*** and specify only needed columns
5. **Use composite indexes** for multi-column WHERE clauses

## Performance Optimization

### Connection Management

```go
// Optimal connection pool configuration
db.SetMaxOpenConns(25)        // Maximum concurrent connections
db.SetMaxIdleConns(5)         // Idle connections to maintain
db.SetConnMaxLifetime(5 * time.Minute)  // Connection lifetime
```

### Query Optimization Techniques

#### 1. Use EXPLAIN for Query Analysis

```sql
EXPLAIN (ANALYZE, BUFFERS)
SELECT ticker, company, action, time
FROM stock_ratings
WHERE ticker = 'AAPL'
ORDER BY time DESC
LIMIT 20;
```

#### 2. Batch Inserts for Better Performance

```go
// Instead of individual inserts
stmt, err := tx.Prepare(`
    INSERT INTO stock_ratings
    (ticker, company, brokerage, action, rating_to, time)
    VALUES ($1, $2, $3, $4, $5, $6)
`)

// Use batch inserts with transactions
for _, rating := range ratings {
    _, err = stmt.Exec(rating.Ticker, rating.Company, ...)
}
```

#### 3. Use Connection Pooling

```go
// Reuse database connections
type Repository struct {
    db *sql.DB  // Connection pool, not individual connection
}
```

### CockroachDB-Specific Optimizations

#### 1. Use UPSERT for Idempotent Operations

```sql
-- Instead of INSERT ... ON CONFLICT
UPSERT INTO enriched_stock_data (ticker, historical_prices, updated_at)
VALUES ($1, $2, NOW());
```

#### 2. Optimize for Distributed Architecture

```sql
-- Use UUID primary keys for better distribution
CREATE TABLE stock_ratings (
    rating_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    -- ...
);
```

#### 3. Leverage JSONB for Flexible Data

```sql
-- Use JSONB indexes for JSON queries
CREATE INDEX idx_prices_jsonb ON enriched_stock_data
USING GIN (historical_prices);

-- Query JSON data efficiently
SELECT ticker FROM enriched_stock_data
WHERE historical_prices @> '{"indicators": {"rsi": 65.4}}';
```

## Backup and Recovery

### CockroachDB Serverless Backup

- **Automatic Backups**: CockroachDB Serverless provides automatic daily backups
- **Point-in-Time Recovery**: Recovery to any point within the backup retention period
- **Cross-Region Replication**: Data is automatically replicated across regions

### Manual Backup Commands

```bash
# Export data for migration or analysis
cockroach sql --url="$DATABASE_URL" \
  --execute="SELECT * FROM stock_ratings" \
  --format=csv > stock_ratings_backup.csv

# Dump schema for documentation
cockroach sql --url="$DATABASE_URL" \
  --execute="SHOW CREATE TABLE stock_ratings"
```

### Recovery Procedures

1. **Point-in-Time Recovery**: Contact CockroachDB support for PITR
2. **Data Restoration**: Re-run data ingestion if recent data is lost
3. **Schema Recovery**: Apply migrations from version control

## Monitoring

### Key Metrics to Monitor

#### Database Performance

- **Query Latency**: P95 and P99 response times
- **Connection Count**: Active and idle connections
- **Query Rate**: Queries per second
- **Error Rate**: Failed query percentage

#### CockroachDB Specific

- **Node Health**: Cluster node status
- **Replication Lag**: Data replication delays
- **Storage Usage**: Disk space utilization
- **Memory Usage**: Buffer and cache utilization

### Monitoring Queries

```sql
-- Check table sizes
SELECT
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
FROM pg_tables
WHERE schemaname = 'public';

-- Check index usage
SELECT
    indexrelname as index_name,
    idx_scan as times_used,
    pg_size_pretty(pg_relation_size(indexrelname::regclass)) as index_size
FROM pg_stat_user_indexes
ORDER BY idx_scan DESC;

-- Check slow queries
SELECT
    query,
    calls,
    total_time,
    mean_time,
    rows
FROM pg_stat_statements
ORDER BY mean_time DESC
LIMIT 10;
```

### Alerting Thresholds

| Metric            | Warning | Critical |
| ----------------- | ------- | -------- |
| Query Latency P95 | > 100ms | > 500ms  |
| Connection Usage  | > 80%   | > 95%    |
| Error Rate        | > 1%    | > 5%     |
| Storage Usage     | > 80%   | > 90%    |

## Troubleshooting

### Common Issues

#### 1. Connection Pool Exhaustion

**Symptoms**: "too many connections" errors

**Solutions**:

```go
// Reduce max connections
db.SetMaxOpenConns(10)

// Reduce connection lifetime
db.SetConnMaxLifetime(1 * time.Minute)

// Ensure connections are closed
defer rows.Close()
```

#### 2. Slow Query Performance

**Symptoms**: High query latency

**Diagnosis**:

```sql
-- Check query execution plan
EXPLAIN (ANALYZE, BUFFERS) SELECT ...;

-- Check missing indexes
SELECT * FROM pg_stat_user_tables WHERE idx_scan = 0;
```

**Solutions**:

- Add appropriate indexes
- Optimize WHERE clauses
- Use LIMIT for pagination

#### 3. Data Inconsistency

**Symptoms**: Duplicate or missing data

**Diagnosis**:

```sql
-- Check for duplicates
SELECT ticker, brokerage, time, COUNT(*)
FROM stock_ratings
GROUP BY ticker, brokerage, time
HAVING COUNT(*) > 1;
```

**Solutions**:

- Add unique constraints
- Implement proper error handling
- Use transactions for data consistency

### Debugging Tools

#### 1. Query Performance

```sql
-- Enable statement statistics
SET cluster setting sql.metrics.statement_details.enabled = true;

-- View slow queries
SELECT * FROM crdb_internal.node_statement_statistics
WHERE service_lat > 1000000;  -- > 1 second
```

#### 2. Connection Monitoring

```go
// Log database stats
stats := db.Stats()
log.Printf("Open connections: %d", stats.OpenConnections)
log.Printf("In use: %d", stats.InUse)
log.Printf("Idle: %d", stats.Idle)
```

#### 3. Transaction Analysis

```sql
-- Check long-running transactions
SELECT * FROM crdb_internal.cluster_transactions
WHERE age(now(), start) > INTERVAL '1 minute';
```

---

_Database Documentation v1.0 - Last updated: December 2024_
