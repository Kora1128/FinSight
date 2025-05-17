package news

import (
	"fmt"
	"testing"
	"time"
)

func TestNewRecommendationCache(t *testing.T) {
	config := CacheConfig{
		TTL:             24 * time.Hour,
		MaxItems:        1000,
		CleanupInterval: 1 * time.Hour,
	}

	cache := NewRecommendationCache(config)
	if cache == nil {
		t.Error("Expected non-nil cache")
	}
	if cache.config != config {
		t.Error("Expected cache config to be set correctly")
	}
	if cache.cache == nil {
		t.Error("Expected cache to be initialized")
	}
}

func TestCache_SetAndGet(t *testing.T) {
	cache := NewRecommendationCache(CacheConfig{
		TTL:             24 * time.Hour,
		MaxItems:        1000,
		CleanupInterval: 1 * time.Hour,
	})

	// Test setting and getting a recommendation
	rec := Recommendation{
		StockSymbol: "NIFTY",
		Action:      ActionBuy,
		Confidence:  0.8,
		NewsItem: NewsItem{
			Title:  "Test News",
			Source: "Test Source",
		},
		CreatedAt: time.Now(),
	}

	cache.Set("test-key", rec)

	// Test getting existing recommendation
	got, found := cache.Get("test-key")
	if !found {
		t.Error("Expected to find recommendation")
	}
	if got.StockSymbol != rec.StockSymbol {
		t.Errorf("Expected stock symbol %s, got %s", rec.StockSymbol, got.StockSymbol)
	}
	if got.Action != rec.Action {
		t.Errorf("Expected action %s, got %s", rec.Action, got.Action)
	}
	if got.Confidence != rec.Confidence {
		t.Errorf("Expected confidence %f, got %f", rec.Confidence, got.Confidence)
	}

	// Test getting non-existent recommendation
	_, found = cache.Get("non-existent-key")
	if found {
		t.Error("Expected not to find non-existent recommendation")
	}
}

func TestCache_GetAll(t *testing.T) {
	cache := NewRecommendationCache(CacheConfig{
		TTL:             24 * time.Hour,
		MaxItems:        1000,
		CleanupInterval: 1 * time.Hour,
	})

	// Add test recommendations
	recommendations := []Recommendation{
		{
			StockSymbol: "NIFTY",
			Action:      ActionBuy,
			Confidence:  0.8,
			NewsItem: NewsItem{
				Title:  "Test News 1",
				Source: "Test Source 1",
			},
			CreatedAt: time.Now(),
		},
		{
			StockSymbol: "RELIANCE",
			Action:      ActionSell,
			Confidence:  0.7,
			NewsItem: NewsItem{
				Title:  "Test News 2",
				Source: "Test Source 2",
			},
			CreatedAt: time.Now(),
		},
	}

	for i, rec := range recommendations {
		cache.Set(fmt.Sprintf("key-%d", i), rec)
	}

	// Test getting all recommendations
	allRecs := cache.GetAll()
	if len(allRecs) != len(recommendations) {
		t.Errorf("Expected %d recommendations, got %d", len(recommendations), len(allRecs))
	}

	// Verify all recommendations are present
	for _, rec := range recommendations {
		found := false
		for _, got := range allRecs {
			if got.StockSymbol == rec.StockSymbol && got.Action == rec.Action {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected to find recommendation for %s", rec.StockSymbol)
		}
	}
}

func TestCache_Expiration(t *testing.T) {
	cache := NewRecommendationCache(CacheConfig{
		TTL:             100 * time.Millisecond,
		MaxItems:        1000,
		CleanupInterval: 50 * time.Millisecond,
	})

	// Add a recommendation
	rec := Recommendation{
		StockSymbol: "NIFTY",
		Action:      ActionBuy,
		Confidence:  0.8,
		NewsItem: NewsItem{
			Title:  "Test News",
			Source: "Test Source",
		},
		CreatedAt: time.Now(),
	}

	cache.Set("test-key", rec)

	// Verify recommendation is present
	_, found := cache.Get("test-key")
	if !found {
		t.Error("Expected to find recommendation before expiration")
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Verify recommendation is expired
	_, found = cache.Get("test-key")
	if found {
		t.Error("Expected recommendation to be expired")
	}
}

func TestCache_MaxItems(t *testing.T) {
	cache := NewRecommendationCache(CacheConfig{
		TTL:             24 * time.Hour,
		MaxItems:        2,
		CleanupInterval: 1 * time.Hour,
	})

	// Add more recommendations than MaxItems
	recommendations := []Recommendation{
		{
			StockSymbol: "NIFTY",
			Action:      ActionBuy,
			Confidence:  0.8,
			NewsItem: NewsItem{
				Title:  "Test News 1",
				Source: "Test Source 1",
			},
			CreatedAt: time.Now(),
		},
		{
			StockSymbol: "RELIANCE",
			Action:      ActionSell,
			Confidence:  0.7,
			NewsItem: NewsItem{
				Title:  "Test News 2",
				Source: "Test Source 2",
			},
			CreatedAt: time.Now(),
		},
		{
			StockSymbol: "TCS",
			Action:      ActionHold,
			Confidence:  0.6,
			NewsItem: NewsItem{
				Title:  "Test News 3",
				Source: "Test Source 3",
			},
			CreatedAt: time.Now(),
		},
	}

	for i, rec := range recommendations {
		cache.Set(fmt.Sprintf("key-%d", i), rec)
	}

	// Verify only MaxItems recommendations are kept
	allRecs := cache.GetAll()
	if len(allRecs) > cache.config.MaxItems {
		t.Errorf("Expected at most %d recommendations, got %d", cache.config.MaxItems, len(allRecs))
	}
}
