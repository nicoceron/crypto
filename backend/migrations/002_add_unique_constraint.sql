-- Migration to add unique constraint for preventing duplicate ratings
-- This migration adds the unique index that was missing from the initial schema

-- Add updated_at column if it doesn't exist
ALTER TABLE stock_ratings ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW();

-- Create unique index to prevent duplicate ratings
-- This will fail if there are existing duplicates, so we need to clean them first
DO $$
BEGIN
    -- Create the unique index
    CREATE UNIQUE INDEX IF NOT EXISTS idx_stock_ratings_unique 
    ON stock_ratings (ticker, brokerage, rating_to, time);
    
    RAISE NOTICE 'Unique index created successfully';
EXCEPTION
    WHEN others THEN
        RAISE NOTICE 'Could not create unique index - likely due to existing duplicates. Error: %', SQLERRM;
        
        -- If the index creation fails due to duplicates, we need to clean them first
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
        
        -- Now try to create the index again
        CREATE UNIQUE INDEX IF NOT EXISTS idx_stock_ratings_unique 
        ON stock_ratings (ticker, brokerage, rating_to, time);
        
        RAISE NOTICE 'Duplicates cleaned and unique index created';
END $$;

-- Create updated_at trigger function if it doesn't exist
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger for updated_at if it doesn't exist
DROP TRIGGER IF EXISTS update_stock_ratings_updated_at ON stock_ratings;
CREATE TRIGGER update_stock_ratings_updated_at 
    BEFORE UPDATE ON stock_ratings 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column(); 