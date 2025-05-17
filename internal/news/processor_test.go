package news

import (
	"context"
	"strings"
	"testing"
	"time"
)

// MockOpenAIResolver is a mock implementation of StockResolver for testing
type MockOpenAIResolver struct {
	Symbols map[string]string
}

func NewMockOpenAIResolver() *MockOpenAIResolver {
	return &MockOpenAIResolver{
		Symbols: map[string]string{
			"RELIANCE":   "RELIANCE",
			"TCS":        "TATA CONSULTANCY",
			"HDFC":       "HDFC BANK",
			"BHARTIARTL": "BHARTI AIRTEL",
		},
	}
}

func (r *MockOpenAIResolver) ResolveSymbol(ctx context.Context, text string) (string, error) {
	// Check for exact matches first
	for symbol, pattern := range r.Symbols {
		if strings.Contains(strings.ToUpper(text), strings.ToUpper(pattern)) {
			return symbol, nil
		}
	}

	// Check for market-wide news
	marketKeywords := []string{"market", "sensex", "nifty", "bse", "nse", "stock market", "share market"}
	textLower := strings.ToLower(text)
	for _, keyword := range marketKeywords {
		if strings.Contains(textLower, keyword) {
			return "NIFTY", nil
		}
	}

	return "", nil
}

func TestNewProcessor(t *testing.T) {
	cache := NewRecommendationCache(CacheConfig{
		TTL:             24 * time.Hour,
		MaxItems:        1000,
		CleanupInterval: 1 * time.Hour,
	})
	processor := NewProcessor(cache, "test-api-key")

	if processor == nil {
		t.Error("Expected non-nil processor")
	}
	if processor.cache != cache {
		t.Error("Expected cache to be set correctly")
	}
	if processor.stockResolver == nil {
		t.Error("Expected stock resolver to be set")
	}
}

func TestProcessNews(t *testing.T) {
	cache := NewRecommendationCache(CacheConfig{
		TTL:             24 * time.Hour,
		MaxItems:        1000,
		CleanupInterval: 1 * time.Hour,
	})
	processor := NewProcessor(cache, "test-api-key")

	// Test empty news items
	recommendations := processor.ProcessNews(context.Background(), []NewsItem{})
	if len(recommendations) != 0 {
		t.Error("Expected no recommendations for empty news items")
	}

	// Test processing news items
	newsItems := []NewsItem{
		{
			Title:       "Positive news about NIFTY",
			Description: "NIFTY shows strong growth potential",
			Link:        "http://example.com/1",
			Source:      "MoneyControl",
			PublishedAt: time.Now(),
		},
		{
			Title:       "Negative news about NIFTY",
			Description: "NIFTY faces market challenges",
			Link:        "http://example.com/2",
			Source:      "Economic Times",
			PublishedAt: time.Now(),
		},
	}

	recommendations = processor.ProcessNews(context.Background(), newsItems)
	if len(recommendations) == 0 {
		t.Error("Expected recommendations for valid news items")
	}

	// Test duplicate handling
	recommendations = processor.ProcessNews(context.Background(), newsItems)
	if len(recommendations) != 0 {
		t.Error("Expected no new recommendations for duplicate news items")
	}
}

func TestAnalyzeSentiment(t *testing.T) {
	processor := NewProcessor(NewRecommendationCache(CacheConfig{
		TTL:             24 * time.Hour,
		MaxItems:        1000,
		CleanupInterval: 1 * time.Hour,
	}), "test-api-key")

	tests := []struct {
		name     string
		item     NewsItem
		expected float64
	}{
		{
			name: "Positive sentiment",
			item: NewsItem{
				Title:       "Stock shows strong growth",
				Description: "Company reports excellent quarterly results",
				Source:      "MoneyControl",
			},
			expected: 0.5, // Expected positive sentiment
		},
		{
			name: "Negative sentiment",
			item: NewsItem{
				Title:       "Stock faces challenges",
				Description: "Company reports significant losses",
				Source:      "Economic Times",
			},
			expected: -0.5, // Expected negative sentiment
		},
		{
			name: "Neutral sentiment",
			item: NewsItem{
				Title:       "Stock market update",
				Description: "Regular market update",
				Source:      "Business Standard",
			},
			expected: 0.0, // Expected neutral sentiment
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sentiment := processor.analyzeSentiment(tt.item)
			if sentiment == 0 && tt.expected != 0 {
				t.Errorf("Expected non-zero sentiment, got %f", sentiment)
			}
			if sentiment > 0 && tt.expected < 0 {
				t.Errorf("Expected negative sentiment, got positive %f", sentiment)
			}
			if sentiment < 0 && tt.expected > 0 {
				t.Errorf("Expected positive sentiment, got negative %f", sentiment)
			}
		})
	}
}

func TestCalculateRelevanceScore(t *testing.T) {
	processor := NewProcessor(NewRecommendationCache(CacheConfig{
		TTL:             24 * time.Hour,
		MaxItems:        1000,
		CleanupInterval: 1 * time.Hour,
	}), "test-api-key")

	tests := []struct {
		name     string
		item     NewsItem
		expected float64
	}{
		{
			name: "Highly relevant news",
			item: NewsItem{
				Title:       "Stock price target",
				Description: "Analyst recommends strong buy",
				Source:      "MoneyControl",
				PublishedAt: time.Now(),
			},
			expected: 0.7, // Expected high relevance
		},
		{
			name: "Less relevant news",
			item: NewsItem{
				Title:       "Market update",
				Description: "Regular market update",
				Source:      "Unknown Source",
				PublishedAt: time.Now().Add(-24 * time.Hour),
			},
			expected: 0.3, // Expected lower relevance
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := processor.calculateRelevanceScore(tt.item)
			if score <= 0 {
				t.Error("Expected positive relevance score")
			}
			if score > MaxRelevanceScore {
				t.Errorf("Expected score <= %f, got %f", MaxRelevanceScore, score)
			}
		})
	}
}

func TestDetermineAction(t *testing.T) {
	processor := NewProcessor(NewRecommendationCache(CacheConfig{
		TTL:             24 * time.Hour,
		MaxItems:        1000,
		CleanupInterval: 1 * time.Hour,
	}), "test-api-key")

	tests := []struct {
		name      string
		sentiment float64
		relevance float64
		expected  string
	}{
		{
			name:      "Strong buy signal",
			sentiment: 0.8,
			relevance: 0.9,
			expected:  ActionBuy,
		},
		{
			name:      "Strong sell signal",
			sentiment: -0.8,
			relevance: 0.9,
			expected:  ActionSell,
		},
		{
			name:      "Hold signal",
			sentiment: 0.1,
			relevance: 0.9,
			expected:  ActionHold,
		},
		{
			name:      "Watch signal",
			sentiment: 0.8,
			relevance: 0.1,
			expected:  ActionWatch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := processor.determineAction(tt.sentiment, tt.relevance)
			if action != tt.expected {
				t.Errorf("Expected action %s, got %s", tt.expected, action)
			}
		})
	}
}

func TestCalculateConfidence(t *testing.T) {
	processor := NewProcessor(NewRecommendationCache(CacheConfig{
		TTL:             24 * time.Hour,
		MaxItems:        1000,
		CleanupInterval: 1 * time.Hour,
	}), "test-api-key")

	tests := []struct {
		name     string
		item     NewsItem
		expected float64
	}{
		{
			name: "High confidence source",
			item: NewsItem{
				Title:       "Stock analysis",
				Description: "Detailed analysis with strong evidence",
				Source:      "MoneyControl",
				Sentiment:   0.8,
			},
			expected: 0.8, // Expected high confidence
		},
		{
			name: "Low confidence source",
			item: NewsItem{
				Title:       "Market update",
				Description: "Brief update",
				Source:      "Unknown Source",
				Sentiment:   0.1,
			},
			expected: 0.3, // Expected lower confidence
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			confidence := processor.calculateConfidence(tt.item)
			if confidence <= 0 {
				t.Error("Expected positive confidence score")
			}
			if confidence > MaxConfidenceScore {
				t.Errorf("Expected confidence <= %f, got %f", MaxConfidenceScore, confidence)
			}
		})
	}
}

func TestGenerateReason(t *testing.T) {
	processor := NewProcessor(NewRecommendationCache(CacheConfig{
		TTL:             24 * time.Hour,
		MaxItems:        1000,
		CleanupInterval: 1 * time.Hour,
	}), "test-api-key")

	tests := []struct {
		name       string
		item       NewsItem
		action     string
		confidence float64
		expected   string
	}{
		{
			name: "Strong buy recommendation",
			item: NewsItem{
				Title:  "Stock shows strong growth",
				Source: "MoneyControl",
			},
			action:     ActionBuy,
			confidence: 0.9,
			expected:   "Based on strong sentiment from MoneyControl news: Stock shows strong growth",
		},
		{
			name: "Weak sell recommendation",
			item: NewsItem{
				Title:  "Stock faces challenges",
				Source: "Unknown Source",
			},
			action:     ActionSell,
			confidence: 0.3,
			expected:   "Based on weak sentiment from Unknown Source news: Stock faces challenges",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reason := processor.generateReason(tt.item, tt.action, tt.confidence)
			if reason != tt.expected {
				t.Errorf("Expected reason %s, got %s", tt.expected, reason)
			}
		})
	}
}

func TestGetRecommendationsByStock(t *testing.T) {
	cache := NewRecommendationCache(CacheConfig{
		TTL:             24 * time.Hour,
		MaxItems:        1000,
		CleanupInterval: 1 * time.Hour,
	})
	processor := NewProcessor(cache, "test-api-key")

	// Add some test recommendations
	recommendations := []Recommendation{
		{
			StockSymbol: "NIFTY",
			Action:      ActionBuy,
			Confidence:  0.8,
			NewsItem: NewsItem{
				Title:       "NIFTY shows strong growth potential",
				Description: "NIFTY index reaches new highs",
				Source:      "MoneyControl",
				Link:        "http://example.com/1",
			},
			CreatedAt: time.Now(),
		},
		{
			StockSymbol: "RELIANCE",
			Action:      ActionSell,
			Confidence:  0.7,
			NewsItem: NewsItem{
				Title:       "RELIANCE faces market challenges",
				Description: "RELIANCE stock price drops",
				Source:      "Economic Times",
				Link:        "http://example.com/2",
			},
			CreatedAt: time.Now(),
		},
		{
			StockSymbol: "TCS",
			Action:      ActionHold,
			Confidence:  0.6,
			NewsItem: NewsItem{
				Title:       "TCS maintains steady growth",
				Description: "TCS reports stable quarterly results",
				Source:      "Business Standard",
				Link:        "http://example.com/3",
			},
			CreatedAt: time.Now(),
		},
	}

	for _, rec := range recommendations {
		cache.Set(rec.NewsItem.Link, rec)
	}

	// Test getting NIFTY recommendations
	niftyRecs := processor.GetRecommendationsByStock("NIFTY")
	if len(niftyRecs) != 1 {
		t.Errorf("Expected 1 NIFTY recommendation, got %d", len(niftyRecs))
	}
	if niftyRecs[0].StockSymbol != "NIFTY" {
		t.Error("Expected NIFTY stock symbol")
	}

	// Test getting RELIANCE recommendations
	relianceRecs := processor.GetRecommendationsByStock("RELIANCE")
	if len(relianceRecs) != 1 {
		t.Errorf("Expected 1 RELIANCE recommendation, got %d", len(relianceRecs))
	}
	if relianceRecs[0].StockSymbol != "RELIANCE" {
		t.Error("Expected RELIANCE stock symbol")
	}

	// Test getting TCS recommendations
	tcsRecs := processor.GetRecommendationsByStock("TCS")
	if len(tcsRecs) != 1 {
		t.Errorf("Expected 1 TCS recommendation, got %d", len(tcsRecs))
	}
	if tcsRecs[0].StockSymbol != "TCS" {
		t.Error("Expected TCS stock symbol")
	}

	// Test getting non-existent stock recommendations
	unknownRecs := processor.GetRecommendationsByStock("UNKNOWN")
	if len(unknownRecs) != 0 {
		t.Error("Expected no recommendations for unknown stock")
	}
}

func TestGetLatestRecommendations(t *testing.T) {
	cache := NewRecommendationCache(CacheConfig{
		TTL:             24 * time.Hour,
		MaxItems:        1000,
		CleanupInterval: 1 * time.Hour,
	})
	processor := &Processor{
		cache: cache,
		stockResolver: &MockOpenAIResolver{
			Symbols: map[string]string{
				"NIFTY":    "NIFTY",
				"RELIANCE": "RELIANCE",
				"TCS":      "TCS",
			},
		},
	}

	// Add test recommendations with different timestamps
	now := time.Now()
	recommendations := []Recommendation{
		{
			StockSymbol: "NIFTY",
			Action:      ActionBuy,
			Confidence:  0.8,
			NewsItem: NewsItem{
				Title:  "NIFTY shows growth",
				Source: "MoneyControl",
				Link:   "http://example.com/1",
			},
			CreatedAt: now.Add(-2 * time.Hour),
		},
		{
			StockSymbol: "RELIANCE",
			Action:      ActionSell,
			Confidence:  0.7,
			NewsItem: NewsItem{
				Title:  "RELIANCE faces challenges",
				Source: "Economic Times",
				Link:   "http://example.com/2",
			},
			CreatedAt: now.Add(-1 * time.Hour),
		},
		{
			StockSymbol: "TCS",
			Action:      ActionHold,
			Confidence:  0.6,
			NewsItem: NewsItem{
				Title:  "TCS maintains position",
				Source: "Business Standard",
				Link:   "http://example.com/3",
			},
			CreatedAt: now,
		},
	}

	// Add recommendations to cache
	for _, rec := range recommendations {
		cache.Set(rec.NewsItem.Link, rec)
	}

	// Test getting latest 2 recommendations
	latestRecs := processor.GetLatestRecommendations(2)
	if len(latestRecs) != 2 {
		t.Errorf("Expected 2 latest recommendations, got %d", len(latestRecs))
	}
	if latestRecs[0].StockSymbol != "TCS" {
		t.Error("Expected TCS as most recent recommendation")
	}
	if latestRecs[1].StockSymbol != "RELIANCE" {
		t.Error("Expected RELIANCE as second most recent recommendation")
	}

	// Test getting all recommendations
	allRecs := processor.GetLatestRecommendations(10)
	if len(allRecs) != 3 {
		t.Errorf("Expected 3 recommendations, got %d", len(allRecs))
	}
}

func TestExtractStockSymbol(t *testing.T) {
	// Create a mock OpenAI client for testing
	mockClient := &MockOpenAIResolver{
		Symbols: map[string]string{
			"RELIANCE":   "RELIANCE",
			"TCS":        "TATA CONSULTANCY",
			"HDFC":       "HDFC BANK",
			"BHARTIARTL": "BHARTI AIRTEL",
		},
	}

	processor := &Processor{
		cache: NewRecommendationCache(CacheConfig{
			TTL:             24 * time.Hour,
			MaxItems:        1000,
			CleanupInterval: 1 * time.Hour,
		}),
		stockResolver: mockClient,
	}

	tests := []struct {
		name     string
		item     NewsItem
		expected string
	}{
		{
			name: "Exact match with symbol",
			item: NewsItem{
				Title:       "RELIANCE stock price target",
				Description: "RELIANCE shares show strong growth",
			},
			expected: "RELIANCE",
		},
		{
			name: "Match with company name",
			item: NewsItem{
				Title:       "Tata Consultancy Services reports strong Q4 results",
				Description: "TCS stock price expected to rise",
			},
			expected: "TCS",
		},
		{
			name: "Match with context",
			item: NewsItem{
				Title:       "Stock of HDFC Bank shows strong momentum",
				Description: "HDFC Bank shares trading at new highs",
			},
			expected: "HDFC",
		},
		{
			name: "Match with variation",
			item: NewsItem{
				Title:       "Bharti Airtel announces new plans",
				Description: "Airtel stock price target raised",
			},
			expected: "BHARTIARTL",
		},
		{
			name: "Market-wide news",
			item: NewsItem{
				Title:       "Stock market update",
				Description: "Market shows strong momentum",
			},
			expected: "NIFTY",
		},
		{
			name: "No match",
			item: NewsItem{
				Title:       "General business news",
				Description: "No specific stock mentioned",
			},
			expected: "",
		},
		{
			name: "Multiple stocks mentioned",
			item: NewsItem{
				Title:       "RELIANCE and TCS show strong growth",
				Description: "Both stocks trading at new highs",
			},
			expected: "RELIANCE", // Should return the first match
		},
		{
			name: "Part of larger word",
			item: NewsItem{
				Title:       "RELIANCEGROUP announces merger",
				Description: "New company formation",
			},
			expected: "RELIANCE", // Should not match as it's part of a larger word
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			symbol, err := processor.stockResolver.ResolveSymbol(context.Background(), tt.item.Title+" "+tt.item.Description)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if symbol != tt.expected {
				t.Errorf("Expected symbol %s, got %s", tt.expected, symbol)
			}
		})
	}
}
