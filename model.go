package main

// TickerQuotes .
type TickerQuotes struct {
	Status string            `json:"status"`
	Data   map[string]Ticker `json:"data"`
}

// Ticker .
type Ticker struct {
	Price float64 `json:"last_price"`
}
