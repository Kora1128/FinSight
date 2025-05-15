package models

import "time"

// RecommendationSource represents the source of a stock recommendation
type RecommendationSource struct {
	Name string `json:"name"`
	Type string `json:"type"` // "news" or "broker"
}

// Recommendation represents a stock recommendation
type Recommendation struct {
	Symbol         string               `json:"symbol"`
	Source         RecommendationSource `json:"source"`
	Date           time.Time            `json:"date"`
	Title          string               `json:"title"`
	Description    string               `json:"description"`
	Link           string               `json:"link"`
	Recommendation string               `json:"recommendation"` // "buy", "sell", "hold", etc.
	TargetPrice    float64              `json:"targetPrice,omitempty"`
	CurrentPrice   float64              `json:"currentPrice,omitempty"`
}

// RecommendationsResponse represents the response for recommendations endpoint
type RecommendationsResponse struct {
	Success bool             `json:"success"`
	Data    []Recommendation `json:"data"`
	Error   string           `json:"error,omitempty"`
}

// TrustedSource represents a trusted news or broker source
type TrustedSource struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // "news" or "broker"
	URL         string `json:"url,omitempty"`
	Description string `json:"description,omitempty"`
}
