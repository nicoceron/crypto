-- CockroachDB Cloud Migration - Initial Schema
-- Note: Database 'stock_data' should already exist in the cloud cluster

-- Enable necessary extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

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
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create unique index to prevent duplicate ratings
CREATE UNIQUE INDEX idx_stock_ratings_unique 
ON stock_ratings (ticker, brokerage, rating_to, time);

-- Create indexes for performance optimization
CREATE INDEX IF NOT EXISTS idx_stock_ratings_ticker ON stock_ratings(ticker);
CREATE INDEX IF NOT EXISTS idx_stock_ratings_time ON stock_ratings(time DESC);
CREATE INDEX IF NOT EXISTS idx_stock_ratings_ticker_time ON stock_ratings(ticker, time DESC);
CREATE INDEX IF NOT EXISTS idx_stock_ratings_company ON stock_ratings USING GIN(company gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_stock_ratings_brokerage ON stock_ratings USING GIN(brokerage gin_trgm_ops);

-- Enable trigram extension for full-text search (if not already enabled)
-- Note: This might already be available in CockroachDB cloud
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Create enriched_stock_data table for additional data
CREATE TABLE IF NOT EXISTS enriched_stock_data (
    ticker VARCHAR(10) PRIMARY KEY,
    company_name VARCHAR(255),
    sector VARCHAR(100),
    industry VARCHAR(100),
    market_cap BIGINT,
    pe_ratio DECIMAL(10,2),
    dividend_yield DECIMAL(5,4),
    beta DECIMAL(5,2),
    week_52_high DECIMAL(10,2),
    week_52_low DECIMAL(10,2),
    avg_volume BIGINT,
    additional_data JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes for enriched data
CREATE INDEX IF NOT EXISTS idx_enriched_stock_data_sector ON enriched_stock_data (sector);
CREATE INDEX IF NOT EXISTS idx_enriched_stock_data_industry ON enriched_stock_data (industry);
CREATE INDEX IF NOT EXISTS idx_enriched_stock_data_market_cap ON enriched_stock_data (market_cap DESC);

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at
CREATE TRIGGER update_stock_ratings_updated_at BEFORE UPDATE ON stock_ratings FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_enriched_stock_data_updated_at BEFORE UPDATE ON enriched_stock_data FOR EACH ROW EXECUTE FUNCTION update_updated_at_column(); 