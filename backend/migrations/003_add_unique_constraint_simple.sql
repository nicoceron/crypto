-- Simplified migration for CockroachDB to add unique constraint

-- Add updated_at column if it doesn't exist
ALTER TABLE stock_ratings ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW();

-- Clean existing duplicates first
-- Keep only the latest created_at for each duplicate group
WITH duplicates AS (
    SELECT 
        ticker, brokerage, rating_to, time,
        rating_id,
        ROW_NUMBER() OVER (
            PARTITION BY ticker, brokerage, rating_to, time 
            ORDER BY created_at DESC
        ) as rn
    FROM stock_ratings
)
DELETE FROM stock_ratings 
WHERE rating_id IN (
    SELECT rating_id FROM duplicates WHERE rn > 1
);

-- Create unique index to prevent future duplicates
CREATE UNIQUE INDEX IF NOT EXISTS idx_stock_ratings_unique 
ON stock_ratings (ticker, brokerage, rating_to, time);

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger for updated_at
DROP TRIGGER IF EXISTS update_stock_ratings_updated_at ON stock_ratings;
CREATE TRIGGER update_stock_ratings_updated_at 
    BEFORE UPDATE ON stock_ratings 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column(); 