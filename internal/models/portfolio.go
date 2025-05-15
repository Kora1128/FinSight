package models

import "time"

// HoldingType represents the type of holding (stock or mutual fund)
type HoldingType string

const (
	HoldingTypeStock      HoldingType = "stock"
	HoldingTypeMutualFund HoldingType = "mutualfund"
)

// Platform represents the broker platform
type Platform string

const (
	PlatformZerodha Platform = "zerodha"
	PlatformICICI   Platform = "icici"
)

// Holding represents a normalized holding item from any broker
type Holding struct {
	ItemName         string      `json:"itemName"`
	ISIN             string      `json:"isin"`
	Quantity         float64     `json:"quantity"`
	AveragePrice     float64     `json:"averagePrice"`
	LastTradedPrice  float64     `json:"lastTradedPrice"`
	CurrentValue     float64     `json:"currentValue"`
	DayChange        float64     `json:"dayChange"`
	DayChangePercent float64     `json:"dayChangePercent"`
	TotalPnL         float64     `json:"totalPnL"`
	Platform         Platform    `json:"platform"`
	Type             HoldingType `json:"type"`
	LastUpdated      time.Time   `json:"lastUpdated"`
}

// Portfolio represents the aggregated portfolio
type Portfolio struct {
	Holdings          []Holding `json:"holdings"`
	TotalValue        float64   `json:"totalValue"`
	TotalDayChange    float64   `json:"totalDayChange"`
	TotalDayChangePct float64   `json:"totalDayChangePct"`
	TotalPnL          float64   `json:"totalPnL"`
	LastUpdated       time.Time `json:"lastUpdated"`
}

// PortfolioRequest represents the request parameters for portfolio endpoints
type PortfolioRequest struct {
	Type HoldingType `form:"type" binding:"omitempty,oneof=stock mutualfund all"`
}

// PortfolioResponse represents the response for portfolio endpoints
type PortfolioResponse struct {
	Success bool      `json:"success"`
	Data    Portfolio `json:"data"`
	Error   string    `json:"error,omitempty"`
}
