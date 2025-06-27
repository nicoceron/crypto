package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"stock-analyzer/internal/domain"
	apperrors "stock-analyzer/pkg/errors"
	"strings"
	"time"
)

// PostgresRepository implements the StockRepository interface for PostgreSQL/CockroachDB
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository creates a new PostgresRepository instance
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// CreateStockRating stores a new stock rating
func (r *PostgresRepository) CreateStockRating(ctx context.Context, rating *domain.StockRating) error {
	query := `
		INSERT INTO stock_ratings (
			rating_id, ticker, company, brokerage, action, 
			rating_from, rating_to, target_from, target_to, time
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.db.ExecContext(ctx, query,
		rating.RatingID, rating.Ticker, rating.Company, rating.Brokerage,
		rating.Action, rating.RatingFrom, rating.RatingTo, rating.TargetFrom,
		rating.TargetTo, rating.Time)

	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeDatabase, "failed to create stock rating")
	}

	return nil
}

// CreateStockRatingsBatch stores multiple stock ratings in a single transaction
func (r *PostgresRepository) CreateStockRatingsBatch(ctx context.Context, ratings []*domain.StockRating) (int, error) {
	if len(ratings) == 0 {
		return 0, nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, apperrors.Wrap(err, apperrors.ErrCodeDatabase, "failed to begin transaction")
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO stock_ratings (
			rating_id, ticker, company, brokerage, action, 
			rating_from, rating_to, target_from, target_to, time
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (ticker, brokerage, rating_to, time) DO NOTHING`)
	if err != nil {
		return 0, apperrors.Wrap(err, apperrors.ErrCodeDatabase, "failed to prepare statement")
	}
	defer stmt.Close()

	insertedCount := 0
	for _, rating := range ratings {
		result, err := stmt.ExecContext(ctx,
			rating.RatingID, rating.Ticker, rating.Company, rating.Brokerage,
			rating.Action, rating.RatingFrom, rating.RatingTo, rating.TargetFrom,
			rating.TargetTo, rating.Time)
		if err != nil {
			// With "ON CONFLICT DO NOTHING", an error here is unexpected.
			// We'll rollback and return the error.
			return 0, apperrors.Wrap(err, apperrors.ErrCodeDatabase, "failed to insert rating")
		}

		// Check if a row was actually inserted
		if rowsAffected, err := result.RowsAffected(); err == nil && rowsAffected > 0 {
			insertedCount++
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, apperrors.Wrap(err, apperrors.ErrCodeDatabase, "failed to commit transaction")
	}

	fmt.Printf("📊 Database batch insert: %d attempted → %d inserted (skipped %d duplicates)\n",
		len(ratings), insertedCount, len(ratings)-insertedCount)

	return insertedCount, nil
}

// GetStockRatings retrieves paginated stock ratings with optional filtering
func (r *PostgresRepository) GetStockRatings(ctx context.Context, filters domain.FilterOptions) (*domain.PaginatedResponse[domain.StockRating], error) {
	page := filters.Page
	if page < 1 {
		page = 1
	}
	limit := filters.Limit
	if limit < 1 || limit > 100 {
		limit = 20
	}
	sortBy := filters.SortBy
	if sortBy == "" {
		sortBy = "time"
	}
	search := filters.Search
	offset := (page - 1) * limit

	// Build WHERE clause for search
	whereClause := ""
	args := []interface{}{}
	argCount := 0

	if search != "" {
		whereClause = "WHERE (company ILIKE $1 OR ticker ILIKE $1 OR brokerage ILIKE $1)"
		args = append(args, "%"+search+"%")
		argCount = 1
	}

	// Validate and build ORDER BY clause
	validSortFields := map[string]bool{
		"time":      true,
		"ticker":    true,
		"company":   true,
		"brokerage": true,
	}

	if !validSortFields[sortBy] {
		sortBy = "time"
	}

	order := "desc"
	if !filters.SortDesc {
		order = "asc"
	}

	orderClause := fmt.Sprintf("ORDER BY %s %s", sortBy, strings.ToUpper(order))

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM stock_ratings %s", whereClause)
	var totalCount int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeDatabase, "failed to get total count")
	}

	// Get paginated results
	query := fmt.Sprintf(`
		SELECT rating_id, ticker, company, brokerage, action, rating_from, 
			   rating_to, target_from, target_to, time, created_at
		FROM stock_ratings %s %s LIMIT $%d OFFSET $%d`,
		whereClause, orderClause, argCount+1, argCount+2)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeDatabase, "failed to query stock ratings")
	}
	defer rows.Close()

	var ratings []domain.StockRating
	for rows.Next() {
		var rating domain.StockRating
		err := rows.Scan(
			&rating.RatingID, &rating.Ticker, &rating.Company, &rating.Brokerage,
			&rating.Action, &rating.RatingFrom, &rating.RatingTo, &rating.TargetFrom,
			&rating.TargetTo, &rating.Time, &rating.CreatedAt)
		if err != nil {
			return nil, apperrors.Wrap(err, apperrors.ErrCodeDatabase, "failed to scan rating")
		}
		ratings = append(ratings, rating)
	}

	if err := rows.Err(); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeDatabase, "error iterating over ratings")
	}

	// Calculate pagination metadata
	totalPages := (totalCount + limit - 1) / limit

	response := &domain.PaginatedResponse[domain.StockRating]{
		Data: ratings,
		Pagination: domain.Pagination{
			Page:       page,
			Limit:      limit,
			TotalItems: totalCount,
			TotalPages: totalPages,
		},
	}

	return response, nil
}

// GetStockRatingsByTicker retrieves all ratings for a specific ticker
func (r *PostgresRepository) GetStockRatingsByTicker(ctx context.Context, ticker string) ([]domain.StockRating, error) {
	query := `
		SELECT rating_id, ticker, company, brokerage, action, rating_from, 
			   rating_to, target_from, target_to, time, created_at
		FROM stock_ratings 
		WHERE ticker = $1 
		ORDER BY time DESC`

	rows, err := r.db.QueryContext(ctx, query, ticker)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeDatabase, "failed to query ratings by ticker")
	}
	defer rows.Close()

	var ratings []domain.StockRating
	for rows.Next() {
		var rating domain.StockRating
		err := rows.Scan(
			&rating.RatingID, &rating.Ticker, &rating.Company, &rating.Brokerage,
			&rating.Action, &rating.RatingFrom, &rating.RatingTo, &rating.TargetFrom,
			&rating.TargetTo, &rating.Time, &rating.CreatedAt)
		if err != nil {
			return nil, apperrors.Wrap(err, apperrors.ErrCodeDatabase, "failed to scan rating")
		}
		ratings = append(ratings, rating)
	}

	if err := rows.Err(); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeDatabase, "error iterating over ratings")
	}

	return ratings, nil
}

// GetUniqueTickers retrieves all unique ticker symbols
func (r *PostgresRepository) GetUniqueTickers(ctx context.Context) ([]string, error) {
	query := "SELECT DISTINCT ticker FROM stock_ratings ORDER BY ticker"

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeDatabase, "failed to query unique tickers")
	}
	defer rows.Close()

	var tickers []string
	for rows.Next() {
		var ticker string
		if err := rows.Scan(&ticker); err != nil {
			return nil, apperrors.Wrap(err, apperrors.ErrCodeDatabase, "failed to scan ticker")
		}
		tickers = append(tickers, ticker)
	}

	if err := rows.Err(); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeDatabase, "error iterating over unique tickers")
	}

	return tickers, nil
}

// CreateEnrichedStockData stores enriched stock data
func (r *PostgresRepository) CreateEnrichedStockData(ctx context.Context, data *domain.EnrichedStockData) error {
	histPricesJSON, err := json.Marshal(data.HistoricalPrices)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeValidation, "failed to marshal historical prices")
	}

	sentimentJSON, err := json.Marshal(data.NewsSentiment)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeValidation, "failed to marshal news sentiment")
	}

	query := `
		INSERT INTO enriched_stock_data (ticker, historical_prices, news_sentiment, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (ticker) DO UPDATE SET
			historical_prices = EXCLUDED.historical_prices,
			news_sentiment = EXCLUDED.news_sentiment,
			updated_at = NOW()`

	_, err = r.db.ExecContext(ctx, query, data.Ticker, histPricesJSON, sentimentJSON)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeDatabase, "failed to create enriched stock data")
	}

	return nil
}

// GetEnrichedStockData retrieves enriched data for a ticker
func (r *PostgresRepository) GetEnrichedStockData(ctx context.Context, ticker string) (*domain.EnrichedStockData, error) {
	query := `
		SELECT ticker, historical_prices, news_sentiment, updated_at
		FROM enriched_stock_data 
		WHERE ticker = $1`

	var data domain.EnrichedStockData
	var histPricesJSON, sentimentJSON []byte

	err := r.db.QueryRowContext(ctx, query, ticker).Scan(
		&data.Ticker, &histPricesJSON, &sentimentJSON, &data.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, apperrors.ErrNotFound.WithDetails(fmt.Sprintf("enriched data for ticker %s not found", ticker))
	}
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeDatabase, "failed to get enriched stock data")
	}

	if err := json.Unmarshal(histPricesJSON, &data.HistoricalPrices); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeDatabase, "failed to unmarshal historical prices")
	}

	if err := json.Unmarshal(sentimentJSON, &data.NewsSentiment); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeDatabase, "failed to unmarshal news sentiment")
	}

	return &data, nil
}

// GetLatestRatingsByTicker gets the most recent rating for each ticker
func (r *PostgresRepository) GetLatestRatingsByTicker(ctx context.Context) (map[string]*domain.StockRating, error) {
	query := `
		SELECT DISTINCT ON (ticker) ticker, rating_id, company, brokerage, action, 
			   rating_from, rating_to, target_from, target_to, time, created_at
		FROM stock_ratings 
		ORDER BY ticker, time DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeDatabase, "failed to query latest ratings")
	}
	defer rows.Close()

	result := make(map[string]*domain.StockRating)
	for rows.Next() {
		var rating domain.StockRating
		err := rows.Scan(
			&rating.Ticker, &rating.RatingID, &rating.Company, &rating.Brokerage,
			&rating.Action, &rating.RatingFrom, &rating.RatingTo, &rating.TargetFrom,
			&rating.TargetTo, &rating.Time, &rating.CreatedAt)
		if err != nil {
			return nil, apperrors.Wrap(err, apperrors.ErrCodeDatabase, "failed to scan latest rating")
		}
		result[rating.Ticker] = &rating
	}

	if err := rows.Err(); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeDatabase, "error iterating over latest ratings")
	}

	return result, nil
}

// DeleteOldEnrichedData removes enriched stock data records older than a given time
func (r *PostgresRepository) DeleteOldEnrichedData(ctx context.Context, olderThan time.Time) (int64, error) {
	query := `DELETE FROM enriched_stock_data WHERE updated_at < $1`

	result, err := r.db.ExecContext(ctx, query, olderThan)
	if err != nil {
		return 0, apperrors.Wrap(err, apperrors.ErrCodeDatabase, "failed to delete old enriched data")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, apperrors.Wrap(err, apperrors.ErrCodeDatabase, "failed to get affected rows after deletion")
	}

	return rowsAffected, nil
}
