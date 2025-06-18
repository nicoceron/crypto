-- CockroachDB Cloud Migration - Initial Schema
-- Note: Database 'stock_data' should already exist in the cloud cluster

-- Create stock_ratings table with UUID primary key to prevent hotspots
CREATE TABLE IF NOT EXISTS stock_ratings (
    rating_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticker VARCHAR(10) NOT NULL,
    company VARCHAR(255) NOT NULL,
    brokerage VARCHAR(255) NOT NULL,
    action VARCHAR(50) NOT NULL,
    rating_from VARCHAR(50),
    rating_to VARCHAR(50) NOT NULL,
    target_from DECIMAL(10, 2),
    target_to DECIMAL(10, 2),
    time TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create enriched_stock_data table for additional data
CREATE TABLE IF NOT EXISTS enriched_stock_data (
    ticker VARCHAR(10) PRIMARY KEY,
    historical_prices JSONB,
    news_sentiment JSONB,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes for performance optimization
CREATE INDEX IF NOT EXISTS idx_stock_ratings_ticker ON stock_ratings(ticker);
CREATE INDEX IF NOT EXISTS idx_stock_ratings_time ON stock_ratings(time DESC);
CREATE INDEX IF NOT EXISTS idx_stock_ratings_ticker_time ON stock_ratings(ticker, time DESC);
CREATE INDEX IF NOT EXISTS idx_stock_ratings_company ON stock_ratings USING GIN(company gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_stock_ratings_brokerage ON stock_ratings USING GIN(brokerage gin_trgm_ops);

-- Enable trigram extension for full-text search (if not already enabled)
-- Note: This might already be available in CockroachDB cloud
CREATE EXTENSION IF NOT EXISTS pg_trgm; 