package alpaca

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
)

// PriceBar represents a normalized price bar for our API
type PriceBar struct {
	Timestamp string  `json:"timestamp"`
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Volume    int64   `json:"volume"`
}

// Snapshot represents current market snapshot data
type Snapshot struct {
	Symbol       string    `json:"symbol"`
	LatestTrade  *Trade    `json:"latest_trade,omitempty"`
	LatestQuote  *Quote    `json:"latest_quote,omitempty"`
	MinuteBar    *PriceBar `json:"minute_bar,omitempty"`
	DailyBar     *PriceBar `json:"daily_bar,omitempty"`
	PrevDailyBar *PriceBar `json:"prev_daily_bar,omitempty"`
}

type Trade struct {
	Timestamp string  `json:"timestamp"`
	Price     float64 `json:"price"`
	Size      int64   `json:"size"`
}

type Quote struct {
	Timestamp string  `json:"timestamp"`
	BidPrice  float64 `json:"bid_price"`
	AskPrice  float64 `json:"ask_price"`
	BidSize   int64   `json:"bid_size"`
	AskSize   int64   `json:"ask_size"`
}

// RateLimiter implements a simple rate limiter for API calls
type RateLimiter struct {
	lastCall time.Time
	mutex    sync.Mutex
	delay    time.Duration
}

// NewRateLimiter creates a new rate limiter with the specified delay between calls
func NewRateLimiter(delay time.Duration) *RateLimiter {
	return &RateLimiter{
		delay: delay,
	}
}

// Wait blocks until it's safe to make the next API call
func (rl *RateLimiter) Wait() {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	
	elapsed := time.Since(rl.lastCall)
	if elapsed < rl.delay {
		waitTime := rl.delay - elapsed
		fmt.Printf("⏳ Rate limiting: waiting %v before next API call\n", waitTime)
		time.Sleep(waitTime)
	}
	rl.lastCall = time.Now()
}

// Service handles Alpaca API interactions using the official SDK
type Service struct {
	client      *marketdata.Client
	rateLimiter *RateLimiter
}

// NewService creates a new Alpaca service with rate limiting
func NewService(apiKey, apiSecret string) *Service {
	// Create Alpaca client using official SDK
	alpacaClient := marketdata.NewClient(marketdata.ClientOpts{
		APIKey:    apiKey,
		APISecret: apiSecret,
		BaseURL:   "https://data.alpaca.markets",
	})

	return &Service{
		client:      alpacaClient,
		rateLimiter: NewRateLimiter(250 * time.Millisecond), // 4 requests per second max
	}
}

// parseTimeFrame converts string timeframe to Alpaca TimeFrame
func (s *Service) parseTimeFrame(timeframe string) marketdata.TimeFrame {
	switch timeframe {
	case "1Min":
		return marketdata.OneMin
	case "5Min":
		return marketdata.NewTimeFrame(5, marketdata.Min)
	case "15Min":
		return marketdata.NewTimeFrame(15, marketdata.Min)
	case "30Min":
		return marketdata.NewTimeFrame(30, marketdata.Min)
	case "1Hour":
		return marketdata.OneHour
	case "1Day":
		return marketdata.OneDay
	case "1Week":
		return marketdata.OneWeek
	case "1Month":
		return marketdata.OneMonth
	default:
		return marketdata.OneDay // Default fallback
	}
}

// GetHistoricalBars fetches historical price data from Alpaca API with rate limiting
func (s *Service) GetHistoricalBars(ctx context.Context, symbol string, timeframe string, start, end time.Time) ([]PriceBar, error) {
	// Apply rate limiting
	s.rateLimiter.Wait()
	
	fmt.Printf("🔸 ALPACA SERVICE: GetHistoricalBars called for %s (%s) from %s to %s (%.1f hours) - WITH RATE LIMITING\n", 
		symbol, timeframe, start.Format("2006-01-02 15:04"), end.Format("2006-01-02 15:04"), end.Sub(start).Hours())
	return s.getAlpacaBars(ctx, symbol, timeframe, start, end)
}

// getAlpacaBars fetches from Alpaca API using official SDK
func (s *Service) getAlpacaBars(ctx context.Context, symbol string, timeframe string, start, end time.Time) ([]PriceBar, error) {
	// Parse the timeframe
	tf := s.parseTimeFrame(timeframe)
	
	// Create bars request using official SDK with dynamic timeframe
	req := marketdata.GetBarsRequest{
		TimeFrame: tf,
		Start:     start,
		End:       end,
		Feed:      marketdata.IEX, // Use IEX feed for better reliability
	}

	fmt.Printf("🔸 ALPACA API: Making %s request for %s from %s to %s (%.1f hours)\n", 
		timeframe, symbol, start.Format("2006-01-02 15:04"), end.Format("2006-01-02 15:04"), end.Sub(start).Hours())

	// Get bars using official SDK (single symbol)
	bars, err := s.client.GetBars(symbol, req)
	if err != nil {
		fmt.Printf("🔴 Alpaca API error for %s (%s): %v\n", symbol, timeframe, err)
		return nil, fmt.Errorf("failed to get bars from Alpaca: %w", err)
	}

	if len(bars) == 0 {
		fmt.Printf("No %s bars returned from Alpaca for %s between %s and %s\n", 
			timeframe, symbol, start.Format("2006-01-02 15:04"), end.Format("2006-01-02 15:04"))
		return []PriceBar{}, fmt.Errorf("no bars found for symbol %s in date range", symbol)
	}

	// Convert to our format
	priceBars := make([]PriceBar, len(bars))
	for i, bar := range bars {
		priceBars[i] = PriceBar{
			Timestamp: bar.Timestamp.Format(time.RFC3339),
			Open:      bar.Open,
			High:      bar.High,
			Low:       bar.Low,
			Close:     bar.Close,
			Volume:    int64(bar.Volume),
		}
	}

	fmt.Printf("✅ Alpaca SUCCESS: returned %d %s bars for %s (requested %s to %s, %.1f hours)\n", 
		len(priceBars), timeframe, symbol, start.Format("2006-01-02 15:04"), end.Format("2006-01-02 15:04"), end.Sub(start).Hours())
	return priceBars, nil
}

// GetSnapshot fetches current market snapshot for a symbol
func (s *Service) GetSnapshot(ctx context.Context, symbol string) (*Snapshot, error) {
	// Apply rate limiting
	s.rateLimiter.Wait()
	
	fmt.Printf("🔸 ALPACA SERVICE: GetSnapshot called for %s\n", symbol)

	req := marketdata.GetSnapshotRequest{
		Feed: marketdata.IEX,
	}

	snapshot, err := s.client.GetSnapshot(symbol, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshot from Alpaca: %w", err)
	}

	if snapshot == nil {
		return nil, fmt.Errorf("no snapshot data available for symbol %s", symbol)
	}

	// Convert to our format
	result := &Snapshot{
		Symbol: symbol,
	}

	// Convert latest trade if available
	if snapshot.LatestTrade != nil {
		result.LatestTrade = &Trade{
			Timestamp: snapshot.LatestTrade.Timestamp.Format(time.RFC3339),
			Price:     snapshot.LatestTrade.Price,
			Size:      int64(snapshot.LatestTrade.Size),
		}
	}

	// Convert latest quote if available
	if snapshot.LatestQuote != nil {
		result.LatestQuote = &Quote{
			Timestamp: snapshot.LatestQuote.Timestamp.Format(time.RFC3339),
			BidPrice:  snapshot.LatestQuote.BidPrice,
			AskPrice:  snapshot.LatestQuote.AskPrice,
			BidSize:   int64(snapshot.LatestQuote.BidSize),
			AskSize:   int64(snapshot.LatestQuote.AskSize),
		}
	}

	// Convert bars if available
	if snapshot.MinuteBar != nil {
		result.MinuteBar = &PriceBar{
			Timestamp: snapshot.MinuteBar.Timestamp.Format(time.RFC3339),
			Open:      snapshot.MinuteBar.Open,
			High:      snapshot.MinuteBar.High,
			Low:       snapshot.MinuteBar.Low,
			Close:     snapshot.MinuteBar.Close,
			Volume:    int64(snapshot.MinuteBar.Volume),
		}
	}

	if snapshot.DailyBar != nil {
		result.DailyBar = &PriceBar{
			Timestamp: snapshot.DailyBar.Timestamp.Format(time.RFC3339),
			Open:      snapshot.DailyBar.Open,
			High:      snapshot.DailyBar.High,
			Low:       snapshot.DailyBar.Low,
			Close:     snapshot.DailyBar.Close,
			Volume:    int64(snapshot.DailyBar.Volume),
		}
	}

	if snapshot.PrevDailyBar != nil {
		result.PrevDailyBar = &PriceBar{
			Timestamp: snapshot.PrevDailyBar.Timestamp.Format(time.RFC3339),
			Open:      snapshot.PrevDailyBar.Open,
			High:      snapshot.PrevDailyBar.High,
			Low:       snapshot.PrevDailyBar.Low,
			Close:     snapshot.PrevDailyBar.Close,
			Volume:    int64(snapshot.PrevDailyBar.Volume),
		}
	}

	return result, nil
}

// GetRecentBars fetches the most recent bars for a symbol (convenience method)
func (s *Service) GetRecentBars(ctx context.Context, symbol string) ([]PriceBar, error) {
	end := time.Now()
	start := end.Add(-24 * time.Hour)
	return s.GetHistoricalBars(ctx, symbol, "1Hour", start, end)
}

// IsMarketHours checks if the current time is during market hours
func (s *Service) IsMarketHours() bool {
	now := time.Now()
	// Simple US market hours check (9:30 AM - 4:00 PM ET, Monday-Friday)
	// This is a simplified version; production code might use more sophisticated timezone handling
	hour := now.Hour()
	weekday := now.Weekday()
	
	return weekday >= time.Monday && weekday <= time.Friday && hour >= 9 && hour < 16
} 