package main

import (
	"fmt"
	"log"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
)

func main() {
	// Initialize Alpaca client with your credentials
	client := marketdata.NewClient(marketdata.ClientOpts{
		APIKey:    "PKP5DG1IOUNBL8LSA8VT",
		APISecret: "QiMcNZ8C1ftFQhWAqhDrXddiA7QSWa1bh7rE2R1z",
		BaseURL:   "https://data.alpaca.markets",
	})

	symbol := "DLR"
	
	// Test different date ranges
	testCases := []struct {
		name  string
		start time.Time
		end   time.Time
	}{
		{
			name:  "1 Week",
			start: time.Now().AddDate(0, 0, -7),
			end:   time.Now(),
		},
		{
			name:  "1 Month", 
			start: time.Now().AddDate(0, -1, 0),
			end:   time.Now(),
		},
		{
			name:  "3 Months",
			start: time.Now().AddDate(0, -3, 0),
			end:   time.Now(),
		},
		{
			name:  "6 Months",
			start: time.Now().AddDate(0, -6, 0),
			end:   time.Now(),
		},
		{
			name:  "1 Year",
			start: time.Now().AddDate(-1, 0, 0),
			end:   time.Now(),
		},
		{
			name:  "2 Years",
			start: time.Now().AddDate(-2, 0, 0),
			end:   time.Now(),
		},
	}

	for _, tc := range testCases {
		fmt.Printf("\n=== Testing %s ===\n", tc.name)
		fmt.Printf("Requesting %s from %s to %s\n", symbol, tc.start.Format("2006-01-02"), tc.end.Format("2006-01-02"))
		
		req := marketdata.GetBarsRequest{
			TimeFrame: marketdata.OneDay,
			Start:     tc.start,
			End:       tc.end,
			Feed:      marketdata.IEX,
		}

		bars, err := client.GetBars(symbol, req)
		if err != nil {
			log.Printf("Error for %s: %v", tc.name, err)
			continue
		}

		fmt.Printf("Received %d bars\n", len(bars))
		if len(bars) > 0 {
			fmt.Printf("First bar: %s\n", bars[0].Timestamp.Format("2006-01-02"))
			fmt.Printf("Last bar:  %s\n", bars[len(bars)-1].Timestamp.Format("2006-01-02"))
		}
	}
} 