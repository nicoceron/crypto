package alpaca

import (
	"context"
	"fmt"
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

// Service handles Alpaca API interactions using the official SDK
type Service struct {
	client *marketdata.Client
}

// NewService creates a new Alpaca service using the official SDK
func NewService(apiKey, apiSecret string) *Service {
	// Create Alpaca client using official SDK
	alpacaClient := marketdata.NewClient(marketdata.ClientOpts{
		APIKey:    apiKey,
		APISecret: apiSecret,
		BaseURL:   "https://data.alpaca.markets",
	})

	return &Service{
		client: alpacaClient,
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

// GetHistoricalBars fetches historical price data from Alpaca API
func (s *Service) GetHistoricalBars(ctx context.Context, symbol string, timeframe string, start, end time.Time) ([]PriceBar, error) {
	fmt.Printf("🔸 ALPACA SERVICE: GetHistoricalBars called for %s (%s) from %s to %s (%.1f hours)\n", 
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
		fmt.Printf("Alpaca API error for %s (%s): %v\n", symbol, timeframe, err)
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
	fmt.Printf("🔸 ALPACA SERVICE: GetSnapshot called for %s\n", symbol)
	
	// Get snapshot using official SDK (v3 method)
	req := marketdata.GetSnapshotRequest{
		Feed: marketdata.IEX,
	}
	snapshot, err := s.client.GetSnapshot(symbol, req)
	if err != nil {
		fmt.Printf("Alpaca snapshot API error for %s: %v\n", symbol, err)
		return nil, fmt.Errorf("failed to get snapshot from Alpaca: %w", err)
	}

	if snapshot == nil {
		return nil, fmt.Errorf("no snapshot data found for symbol %s", symbol)
	}

	// Convert to our format
	result := &Snapshot{
		Symbol: symbol,
	}

	// Latest trade
	if snapshot.LatestTrade != nil {
		result.LatestTrade = &Trade{
			Timestamp: snapshot.LatestTrade.Timestamp.Format(time.RFC3339),
			Price:     snapshot.LatestTrade.Price,
			Size:      int64(snapshot.LatestTrade.Size),
		}
	}

	// Latest quote
	if snapshot.LatestQuote != nil {
		result.LatestQuote = &Quote{
			Timestamp: snapshot.LatestQuote.Timestamp.Format(time.RFC3339),
			BidPrice:  snapshot.LatestQuote.BidPrice,
			AskPrice:  snapshot.LatestQuote.AskPrice,
			BidSize:   int64(snapshot.LatestQuote.BidSize),
			AskSize:   int64(snapshot.LatestQuote.AskSize),
		}
	}

	// Minute bar
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

	// Daily bar
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

	// Previous daily bar
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

	fmt.Printf("✅ Alpaca snapshot SUCCESS for %s\n", symbol)
	return result, nil
}

// GetRecentBars gets recent historical data (last 30 days with 1D timeframe)
func (s *Service) GetRecentBars(ctx context.Context, symbol string) ([]PriceBar, error) {
	end := time.Now()
	start := end.AddDate(0, 0, -30) // 30 days ago

	return s.GetHistoricalBars(ctx, symbol, "1Day", start, end)
}

// IsMarketHours checks if current time is during market hours (9:30 AM - 4:00 PM ET)
func (s *Service) IsMarketHours() bool {
	now := time.Now()
	
	// Convert to ET timezone
	et, err := time.LoadLocation("America/New_York")
	if err != nil {
		return false
	}
	
	nowET := now.In(et)
	
	// Check if it's a weekday
	if nowET.Weekday() == time.Saturday || nowET.Weekday() == time.Sunday {
		return false
	}
	
	// Check if it's during market hours (9:30 AM - 4:00 PM ET)
	hour := nowET.Hour()
	minute := nowET.Minute()
	
	// Market opens at 9:30 AM
	if hour < 9 || (hour == 9 && minute < 30) {
		return false
	}
	
	// Market closes at 4:00 PM
	if hour >= 16 {
		return false
	}
	
	return true
} 