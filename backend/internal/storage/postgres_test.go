package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"stock-analyzer/internal/domain"
	apperrors "stock-analyzer/pkg/errors"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *PostgresRepository) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)

	repo := NewPostgresRepository(db)
	return db, mock, repo
}

func setupMockDBForBenchmark(b *testing.B) (*sql.DB, sqlmock.Sqlmock, *PostgresRepository) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(b, err)

	repo := NewPostgresRepository(db)
	return db, mock, repo
}

func TestCreateStockRating_Success(t *testing.T) {
	t.Log("Testing CreateStockRating: successful creation")
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	rating := &domain.StockRating{
		RatingID:   uuid.New(),
		Ticker:     "AAPL",
		Company:    "Apple Inc.",
		Brokerage:  "Goldman Sachs",
		Action:     "upgraded by",
		RatingFrom: stringPtr("Hold"),
		RatingTo:   "Buy",
		TargetFrom: float64Ptr(150.0),
		TargetTo:   float64Ptr(180.0),
		Time:       time.Now(),
	}

	mock.ExpectExec(`
		INSERT INTO stock_ratings (
			rating_id, ticker, company, brokerage, action, 
			rating_from, rating_to, target_from, target_to, time
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`).
		WithArgs(rating.RatingID, rating.Ticker, rating.Company, rating.Brokerage,
			rating.Action, rating.RatingFrom, rating.RatingTo, rating.TargetFrom,
			rating.TargetTo, rating.Time).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.CreateStockRating(context.Background(), rating)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateStockRating_DatabaseError(t *testing.T) {
	t.Log("Testing CreateStockRating: handles database error")
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	rating := &domain.StockRating{
		RatingID:  uuid.New(),
		Ticker:    "AAPL",
		Company:   "Apple Inc.",
		Brokerage: "Goldman Sachs",
		Action:    "upgraded by",
		RatingTo:  "Buy",
		Time:      time.Now(),
	}

	mock.ExpectExec(`
		INSERT INTO stock_ratings (
			rating_id, ticker, company, brokerage, action, 
			rating_from, rating_to, target_from, target_to, time
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`).
		WithArgs(rating.RatingID, rating.Ticker, rating.Company, rating.Brokerage,
			rating.Action, rating.RatingFrom, rating.RatingTo, rating.TargetFrom,
			rating.TargetTo, rating.Time).
		WillReturnError(fmt.Errorf("database connection error"))

	err := repo.CreateStockRating(context.Background(), rating)

	assert.Error(t, err)
	var appErr *apperrors.AppError
	assert.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperrors.ErrCodeDatabase, appErr.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateStockRatingsBatch_Success(t *testing.T) {
	t.Log("Testing CreateStockRatingsBatch: successful batch insert")
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	ratings := []*domain.StockRating{
		{
			RatingID:  uuid.New(),
			Ticker:    "AAPL",
			Company:   "Apple Inc.",
			Brokerage: "Goldman Sachs",
			Action:    "upgraded by",
			RatingTo:  "Buy",
			Time:      time.Now(),
		},
		{
			RatingID:  uuid.New(),
			Ticker:    "GOOGL",
			Company:   "Alphabet Inc.",
			Brokerage: "Morgan Stanley",
			Action:    "initiated by",
			RatingTo:  "Strong Buy",
			Time:      time.Now().Add(-time.Hour),
		},
	}

	mock.ExpectBegin()
	mock.ExpectPrepare(`
		INSERT INTO stock_ratings (
			rating_id, ticker, company, brokerage, action, 
			rating_from, rating_to, target_from, target_to, time
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (ticker, brokerage, rating_to, time) DO NOTHING`)

	for _, rating := range ratings {
		mock.ExpectExec(`
		INSERT INTO stock_ratings (
			rating_id, ticker, company, brokerage, action, 
			rating_from, rating_to, target_from, target_to, time
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (ticker, brokerage, rating_to, time) DO NOTHING`).
			WithArgs(rating.RatingID, rating.Ticker, rating.Company, rating.Brokerage,
				rating.Action, rating.RatingFrom, rating.RatingTo, rating.TargetFrom,
				rating.TargetTo, rating.Time).
			WillReturnResult(sqlmock.NewResult(1, 1))
	}

	mock.ExpectCommit()

	insertedCount, err := repo.CreateStockRatingsBatch(context.Background(), ratings)
	assert.NoError(t, err)
	assert.Equal(t, 2, insertedCount)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateStockRatingsBatch_EmptySlice(t *testing.T) {
	t.Log("Testing CreateStockRatingsBatch: handles empty input slice")
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	insertedCount, err := repo.CreateStockRatingsBatch(context.Background(), []*domain.StockRating{})
	assert.NoError(t, err)
	assert.Equal(t, 0, insertedCount)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateStockRatingsBatch_TransactionError(t *testing.T) {
	t.Log("Testing CreateStockRatingsBatch: handles transaction begin error")
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	ratings := []*domain.StockRating{
		{
			RatingID:  uuid.New(),
			Ticker:    "AAPL",
			Company:   "Apple Inc.",
			Brokerage: "Goldman Sachs",
			Action:    "upgraded by",
			RatingTo:  "Buy",
			Time:      time.Now(),
		},
	}

	mock.ExpectBegin().WillReturnError(fmt.Errorf("transaction begin error"))

	insertedCount, err := repo.CreateStockRatingsBatch(context.Background(), ratings)

	assert.Error(t, err)
	assert.Equal(t, 0, insertedCount)
	var appErr *apperrors.AppError
	assert.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperrors.ErrCodeDatabase, appErr.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateStockRatingsBatch_DuplicateHandling(t *testing.T) {
	t.Log("Testing CreateStockRatingsBatch: handles duplicate primary keys")
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	ratings := []*domain.StockRating{
		{
			RatingID:  uuid.New(),
			Ticker:    "AAPL",
			Company:   "Apple Inc.",
			Brokerage: "Goldman Sachs",
			Action:    "upgraded by",
			RatingTo:  "Buy",
			Time:      time.Now(),
		},
		{
			RatingID:  uuid.New(),
			Ticker:    "AAPL", // Duplicate
			Company:   "Apple Inc.",
			Brokerage: "Goldman Sachs",
			Action:    "upgraded by",
			RatingTo:  "Buy",
			Time:      time.Now(),
		},
	}

	mock.ExpectBegin()
	mock.ExpectPrepare(`
		INSERT INTO stock_ratings (
			rating_id, ticker, company, brokerage, action, 
			rating_from, rating_to, target_from, target_to, time
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (ticker, brokerage, rating_to, time) DO NOTHING`)

	// First insert succeeds
	mock.ExpectExec(`
		INSERT INTO stock_ratings (
			rating_id, ticker, company, brokerage, action, 
			rating_from, rating_to, target_from, target_to, time
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (ticker, brokerage, rating_to, time) DO NOTHING`).
		WithArgs(ratings[0].RatingID, ratings[0].Ticker, ratings[0].Company,
			ratings[0].Brokerage, ratings[0].Action, ratings[0].RatingFrom,
			ratings[0].RatingTo, ratings[0].TargetFrom, ratings[0].TargetTo, ratings[0].Time).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Second insert is ignored due to conflict
	mock.ExpectExec(`
		INSERT INTO stock_ratings (
			rating_id, ticker, company, brokerage, action, 
			rating_from, rating_to, target_from, target_to, time
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (ticker, brokerage, rating_to, time) DO NOTHING`).
		WithArgs(ratings[1].RatingID, ratings[1].Ticker, ratings[1].Company,
			ratings[1].Brokerage, ratings[1].Action, ratings[1].RatingFrom,
								ratings[1].RatingTo, ratings[1].TargetFrom, ratings[1].TargetTo, ratings[1].Time).
		WillReturnResult(sqlmock.NewResult(0, 0)) // No rows affected due to conflict

	mock.ExpectCommit()

	insertedCount, err := repo.CreateStockRatingsBatch(context.Background(), ratings)
	assert.NoError(t, err)
	assert.Equal(t, 1, insertedCount)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetStockRatings_Success(t *testing.T) {
	t.Log("Testing GetStockRatings: successful retrieval of ratings")
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	expectedRatings := []domain.StockRating{
		{
			RatingID:  uuid.New(),
			Ticker:    "AAPL",
			Company:   "Apple Inc.",
			Brokerage: "Goldman Sachs",
			Action:    "upgraded by",
			RatingTo:  "Buy",
			Time:      time.Now(),
			CreatedAt: time.Now(),
		},
	}

	// Mock count query
	mock.ExpectQuery("SELECT COUNT(*) FROM stock_ratings ").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// Mock main query
	rows := sqlmock.NewRows([]string{
		"rating_id", "ticker", "company", "brokerage", "action",
		"rating_from", "rating_to", "target_from", "target_to", "time", "created_at",
	})

	for _, rating := range expectedRatings {
		rows.AddRow(rating.RatingID, rating.Ticker, rating.Company, rating.Brokerage,
			rating.Action, rating.RatingFrom, rating.RatingTo, rating.TargetFrom,
			rating.TargetTo, rating.Time, rating.CreatedAt)
	}

	mock.ExpectQuery(`
		SELECT rating_id, ticker, company, brokerage, action, rating_from, 
			   rating_to, target_from, target_to, time, created_at
		FROM stock_ratings  ORDER BY time DESC LIMIT $1 OFFSET $2`).
		WithArgs(20, 0).
		WillReturnRows(rows)

	filters := domain.FilterOptions{Page: 1, Limit: 20, SortBy: "time", SortDesc: true}
	response, err := repo.GetStockRatings(context.Background(), filters)

	assert.NoError(t, err)
	assert.Equal(t, 1, response.Pagination.TotalItems)
	assert.Len(t, response.Data, 1)
	assert.Equal(t, expectedRatings[0].Ticker, response.Data[0].Ticker)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetStockRatings_WithSearch(t *testing.T) {
	t.Log("Testing GetStockRatings: with search query")
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	searchTerm := "Apple"

	// Mock count query with search
	mock.ExpectQuery("SELECT COUNT(*) FROM stock_ratings WHERE (company ILIKE $1 OR ticker ILIKE $1 OR brokerage ILIKE $1)").
		WithArgs("%Apple%").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// Mock main query with search
	rows := sqlmock.NewRows([]string{
		"rating_id", "ticker", "company", "brokerage", "action",
		"rating_from", "rating_to", "target_from", "target_to", "time", "created_at",
	}).AddRow(uuid.New(), "AAPL", "Apple Inc.", "Goldman Sachs", "upgraded by",
		nil, "Buy", nil, nil, time.Now(), time.Now())

	mock.ExpectQuery(`
		SELECT rating_id, ticker, company, brokerage, action, rating_from, 
			   rating_to, target_from, target_to, time, created_at
		FROM stock_ratings WHERE (company ILIKE $1 OR ticker ILIKE $1 OR brokerage ILIKE $1) ORDER BY time DESC LIMIT $2 OFFSET $3`).
		WithArgs("%Apple%", 20, 0).
		WillReturnRows(rows)

	filters := domain.FilterOptions{Page: 1, Limit: 20, SortBy: "time", SortDesc: true, Search: searchTerm}
	response, err := repo.GetStockRatings(context.Background(), filters)

	assert.NoError(t, err)
	assert.Equal(t, 1, response.Pagination.TotalItems)
	assert.Len(t, response.Data, 1)
	assert.Contains(t, response.Data[0].Company, "Apple")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetStockRatings_InvalidSortField(t *testing.T) {
	t.Log("Testing GetStockRatings: handles invalid sort field")
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	// Mock count query
	mock.ExpectQuery("SELECT COUNT(*) FROM stock_ratings ").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// Mock main query with default sort field (time)
	mock.ExpectQuery(`
		SELECT rating_id, ticker, company, brokerage, action, rating_from, 
			   rating_to, target_from, target_to, time, created_at
		FROM stock_ratings  ORDER BY time DESC LIMIT $1 OFFSET $2`).
		WithArgs(20, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"rating_id", "ticker", "company", "brokerage", "action",
			"rating_from", "rating_to", "target_from", "target_to", "time", "created_at",
		}))

	// Try to sort by invalid field
	filters := domain.FilterOptions{Page: 1, Limit: 20, SortBy: "invalid_field", SortDesc: true}
	response, err := repo.GetStockRatings(context.Background(), filters)

	assert.NoError(t, err)
	assert.Equal(t, 0, response.Pagination.TotalItems)
	assert.Len(t, response.Data, 0)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetStockRatings_CountError(t *testing.T) {
	t.Log("Testing GetStockRatings: handles error during count query")
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	mock.ExpectQuery("SELECT COUNT(*) FROM stock_ratings ").
		WillReturnError(fmt.Errorf("count query error"))

	filters := domain.FilterOptions{Page: 1, Limit: 20, SortBy: "time", SortDesc: true}
	response, err := repo.GetStockRatings(context.Background(), filters)

	assert.Error(t, err)
	assert.Nil(t, response)

	var appErr *apperrors.AppError
	assert.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperrors.ErrCodeDatabase, appErr.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetStockRatingsByTicker_Success(t *testing.T) {
	t.Log("Testing GetStockRatingsByTicker: successful retrieval for a ticker")
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	ticker := "AAPL"

	rows := sqlmock.NewRows([]string{
		"rating_id", "ticker", "company", "brokerage", "action",
		"rating_from", "rating_to", "target_from", "target_to", "time", "created_at",
	}).AddRow(uuid.New(), "AAPL", "Apple Inc.", "Goldman Sachs", "upgraded by",
		nil, "Buy", nil, nil, time.Now(), time.Now())

	mock.ExpectQuery(`
		SELECT rating_id, ticker, company, brokerage, action, rating_from, 
			   rating_to, target_from, target_to, time, created_at
		FROM stock_ratings 
		WHERE ticker = $1 
		ORDER BY time DESC`).
		WithArgs(ticker).
		WillReturnRows(rows)

	ratings, err := repo.GetStockRatingsByTicker(context.Background(), ticker)

	assert.NoError(t, err)
	assert.Len(t, ratings, 1)
	assert.Equal(t, ticker, ratings[0].Ticker)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetStockRatingsByTicker_NotFound(t *testing.T) {
	t.Log("Testing GetStockRatingsByTicker: handles ticker not found")
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	ticker := "NONEXISTENT"

	mock.ExpectQuery(`
		SELECT rating_id, ticker, company, brokerage, action, rating_from, 
			   rating_to, target_from, target_to, time, created_at
		FROM stock_ratings 
		WHERE ticker = $1 
		ORDER BY time DESC`).
		WithArgs(ticker).
		WillReturnRows(sqlmock.NewRows([]string{
			"rating_id", "ticker", "company", "brokerage", "action",
			"rating_from", "rating_to", "target_from", "target_to", "time", "created_at",
		}))

	ratings, err := repo.GetStockRatingsByTicker(context.Background(), ticker)

	assert.NoError(t, err)
	assert.Len(t, ratings, 0)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetStockRatingsByTicker_DatabaseError(t *testing.T) {
	t.Log("Testing GetStockRatingsByTicker: handles database error")
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	ticker := "AAPL"

	mock.ExpectQuery(`
		SELECT rating_id, ticker, company, brokerage, action, rating_from, 
			   rating_to, target_from, target_to, time, created_at
		FROM stock_ratings 
		WHERE ticker = $1 
		ORDER BY time DESC`).
		WithArgs(ticker).
		WillReturnError(fmt.Errorf("database error"))

	ratings, err := repo.GetStockRatingsByTicker(context.Background(), ticker)

	assert.Error(t, err)
	assert.Nil(t, ratings)

	var appErr *apperrors.AppError
	assert.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperrors.ErrCodeDatabase, appErr.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUniqueTickers_Success(t *testing.T) {
	t.Log("Testing GetUniqueTickers: successful retrieval of unique tickers")
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	expectedTickers := []string{"AAPL", "GOOGL", "MSFT"}

	rows := sqlmock.NewRows([]string{"ticker"})
	for _, ticker := range expectedTickers {
		rows.AddRow(ticker)
	}

	mock.ExpectQuery("SELECT DISTINCT ticker FROM stock_ratings ORDER BY ticker").
		WillReturnRows(rows)

	tickers, err := repo.GetUniqueTickers(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedTickers, tickers)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateEnrichedStockData_Success(t *testing.T) {
	t.Log("Testing CreateEnrichedStockData: successful creation")
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	enrichedData := &domain.EnrichedStockData{
		Ticker: "AAPL",
		HistoricalPrices: map[string]interface{}{
			"data": []map[string]interface{}{
				{"close": 150.0, "volume": 1000000},
			},
		},
		NewsSentiment: map[string]interface{}{
			"sentiment_score": 0.7,
			"articles_count":  10,
		},
		UpdatedAt: time.Now(),
	}

	historicalJSON, _ := json.Marshal(enrichedData.HistoricalPrices)
	sentimentJSON, _ := json.Marshal(enrichedData.NewsSentiment)

	mock.ExpectExec(`
		INSERT INTO enriched_stock_data (ticker, historical_prices, news_sentiment, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (ticker) DO UPDATE SET
			historical_prices = EXCLUDED.historical_prices,
			news_sentiment = EXCLUDED.news_sentiment,
			updated_at = NOW()`).
		WithArgs(enrichedData.Ticker, historicalJSON, sentimentJSON).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.CreateEnrichedStockData(context.Background(), enrichedData)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetEnrichedStockData_Success(t *testing.T) {
	t.Log("Testing GetEnrichedStockData: successful retrieval")
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	ticker := "AAPL"
	historicalJSON := `{"data":[{"close":150,"volume":1000000}]}`
	sentimentJSON := `{"articles_count":10,"sentiment_score":0.7}`
	updatedAt := time.Now()

	rows := sqlmock.NewRows([]string{"ticker", "historical_prices", "news_sentiment", "updated_at"}).
		AddRow(ticker, historicalJSON, sentimentJSON, updatedAt)

	mock.ExpectQuery(`
		SELECT ticker, historical_prices, news_sentiment, updated_at
		FROM enriched_stock_data 
		WHERE ticker = $1`).
		WithArgs(ticker).
		WillReturnRows(rows)

	data, err := repo.GetEnrichedStockData(context.Background(), ticker)

	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, ticker, data.Ticker)
	assert.Contains(t, data.HistoricalPrices, "data")
	assert.Contains(t, data.NewsSentiment, "sentiment_score")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetEnrichedStockData_NotFound(t *testing.T) {
	t.Log("Testing GetEnrichedStockData: handles not found error")
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	ticker := "NONEXISTENT"

	mock.ExpectQuery(`
		SELECT ticker, historical_prices, news_sentiment, updated_at
		FROM enriched_stock_data 
		WHERE ticker = $1`).
		WithArgs(ticker).
		WillReturnError(sql.ErrNoRows)

	data, err := repo.GetEnrichedStockData(context.Background(), ticker)

	assert.Error(t, err)
	assert.Nil(t, data)

	var appErr *apperrors.AppError
	assert.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperrors.ErrCodeNotFound, appErr.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetLatestRatingsByTicker_Success(t *testing.T) {
	t.Log("Testing GetLatestRatingsByTicker: successful retrieval of latest ratings")
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	rows := sqlmock.NewRows([]string{
		"ticker", "rating_id", "company", "brokerage", "action",
		"rating_from", "rating_to", "target_from", "target_to", "time", "created_at",
	}).
		AddRow("AAPL", uuid.New(), "Apple Inc.", "Goldman Sachs", "upgraded by",
			nil, "Buy", nil, nil, time.Now(), time.Now()).
		AddRow("GOOGL", uuid.New(), "Alphabet Inc.", "Morgan Stanley", "initiated by",
			nil, "Strong Buy", nil, nil, time.Now().Add(-time.Hour), time.Now().Add(-time.Hour))

	mock.ExpectQuery(`
		SELECT DISTINCT ON (ticker) ticker, rating_id, company, brokerage, action, 
			   rating_from, rating_to, target_from, target_to, time, created_at
		FROM stock_ratings 
		ORDER BY ticker, time DESC`).
		WillReturnRows(rows)

	ratingsMap, err := repo.GetLatestRatingsByTicker(context.Background())

	assert.NoError(t, err)
	assert.Len(t, ratingsMap, 2)
	assert.Contains(t, ratingsMap, "AAPL")
	assert.Contains(t, ratingsMap, "GOOGL")
	assert.Equal(t, "Buy", ratingsMap["AAPL"].RatingTo)
	assert.Equal(t, "Strong Buy", ratingsMap["GOOGL"].RatingTo)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Benchmark tests for expensive operations
func BenchmarkCreateStockRatingsBatch(b *testing.B) {
	b.Log("Benchmarking CreateStockRatingsBatch")
	db, mock, repo := setupMockDBForBenchmark(b)
	defer db.Close()

	// Create a large batch of ratings
	ratings := make([]*domain.StockRating, 1000)
	for i := 0; i < 1000; i++ {
		ratings[i] = &domain.StockRating{
			RatingID:  uuid.New(),
			Ticker:    fmt.Sprintf("TICK%d", i),
			Company:   fmt.Sprintf("Company %d", i),
			Brokerage: "Test Brokerage",
			Action:    "upgraded by",
			RatingTo:  "Buy",
			Time:      time.Now().Add(-time.Duration(i) * time.Minute),
		}
	}

	// Mock expectations for each benchmark iteration
	for i := 0; i < b.N; i++ {
		mock.ExpectBegin()
		mock.ExpectPrepare(`
		INSERT INTO stock_ratings (
			rating_id, ticker, company, brokerage, action, 
			rating_from, rating_to, target_from, target_to, time
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (ticker, brokerage, rating_to, time) DO NOTHING`)

		for _, rating := range ratings {
			mock.ExpectExec(`
		INSERT INTO stock_ratings (
			rating_id, ticker, company, brokerage, action, 
			rating_from, rating_to, target_from, target_to, time
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (ticker, brokerage, rating_to, time) DO NOTHING`).
				WithArgs(rating.RatingID, rating.Ticker, rating.Company, rating.Brokerage,
					rating.Action, rating.RatingFrom, rating.RatingTo, rating.TargetFrom,
					rating.TargetTo, rating.Time).
				WillReturnResult(sqlmock.NewResult(1, 1))
		}

		mock.ExpectCommit()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.CreateStockRatingsBatch(context.Background(), ratings)
		require.NoError(b, err)
	}
}

func BenchmarkGetStockRatings(b *testing.B) {
	b.Log("Benchmarking GetStockRatings")
	db, mock, repo := setupMockDBForBenchmark(b)
	defer db.Close()

	// Mock large result set
	rows := sqlmock.NewRows([]string{
		"rating_id", "ticker", "company", "brokerage", "action",
		"rating_from", "rating_to", "target_from", "target_to", "time", "created_at",
	})

	for i := 0; i < 100; i++ {
		rows.AddRow(uuid.New(), fmt.Sprintf("TICK%d", i), fmt.Sprintf("Company %d", i),
			"Test Brokerage", "upgraded by", nil, "Buy", nil, nil,
			time.Now().Add(-time.Duration(i)*time.Hour), time.Now().Add(-time.Duration(i)*time.Hour))
	}

	for i := 0; i < b.N; i++ {
		mock.ExpectQuery("SELECT COUNT(*) FROM stock_ratings ").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10000))

		mock.ExpectQuery(`
		SELECT rating_id, ticker, company, brokerage, action, rating_from, 
			   rating_to, target_from, target_to, time, created_at
		FROM stock_ratings  ORDER BY time DESC LIMIT $1 OFFSET $2`).
			WithArgs(20, 0).
			WillReturnRows(rows)
	}

	filters := domain.FilterOptions{Page: 1, Limit: 20, SortBy: "time", SortDesc: true}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.GetStockRatings(context.Background(), filters)
		require.NoError(b, err)
	}
}

// Stress test for concurrent operations
func TestConcurrentCreateStockRating(t *testing.T) {
	t.Log("Testing CreateStockRating: with high concurrency")
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	concurrency := 10
	done := make(chan error, concurrency)

	// Mock expectations for concurrent inserts
	for i := 0; i < concurrency; i++ {
		mock.ExpectExec(`
		INSERT INTO stock_ratings (
			rating_id, ticker, company, brokerage, action, 
			rating_from, rating_to, target_from, target_to, time
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`).
			WillReturnResult(sqlmock.NewResult(1, 1))
	}

	for i := 0; i < concurrency; i++ {
		go func(id int) {
			rating := &domain.StockRating{
				RatingID:  uuid.New(),
				Ticker:    fmt.Sprintf("TICK%d", id),
				Company:   fmt.Sprintf("Company %d", id),
				Brokerage: "Test Brokerage",
				Action:    "upgraded by",
				RatingTo:  "Buy",
				Time:      time.Now(),
			}

			err := repo.CreateStockRating(context.Background(), rating)
			done <- err
		}(i)
	}

	// Wait for all operations to complete
	for i := 0; i < concurrency; i++ {
		err := <-done
		assert.NoError(t, err)
	}

	assert.NoError(t, mock.ExpectationsWereMet())
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func float64Ptr(f float64) *float64 {
	return &f
}
