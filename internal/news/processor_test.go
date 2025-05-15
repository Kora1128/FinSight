package news

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestProcessor_ProcessNews(t *testing.T) {
	cache := NewRecommendationCache(GetDefaultCacheConfig())
	processor := NewProcessor(cache)

	now := time.Now()
	newsItems := []NewsItem{
		{
			Title:       "RELIANCE announces strong Q3 results",
			Description: "Reliance Industries reported better-than-expected quarterly results with strong growth in all segments.",
			Link:        "http://example.com/news1",
			PublishedAt: now,
			Source: Source{
				Name: "MoneyControl",
			},
			Sentiment: 0.8,
		},
		{
			Title:       "TCS stock price target raised by analysts",
			Description: "Multiple analysts have raised their price targets for TCS following strong performance.",
			Link:        "http://example.com/news2",
			PublishedAt: now,
			Source: Source{
				Name: "Economic Times",
			},
			Sentiment: 0.6,
		},
		{
			Title:       "Market Update: Sensex gains 500 points",
			Description: "The benchmark Sensex gained 500 points in today's trading session.",
			Link:        "http://example.com/news3",
			PublishedAt: now,
			Source: Source{
				Name: "Business Standard",
			},
			Sentiment: 0.2,
		},
	}

	recommendations := processor.ProcessNews(context.Background(), newsItems)

	// Test that we got recommendations
	assert.NotEmpty(t, recommendations)

	// Test each recommendation
	for _, rec := range recommendations {
		// Verify basic fields are set
		assert.NotEmpty(t, rec.Title)
		assert.NotEmpty(t, rec.Description)
		assert.NotEmpty(t, rec.Link)
		assert.NotZero(t, rec.LastUpdated)

		// Verify relevance score is within bounds
		assert.GreaterOrEqual(t, rec.RelevanceScore, 0.5)
		assert.LessOrEqual(t, rec.RelevanceScore, 1.0)

		// Verify confidence is within bounds
		assert.GreaterOrEqual(t, rec.Confidence, 0.0)
		assert.LessOrEqual(t, rec.Confidence, 1.0)

		// Verify action is one of the expected values
		assert.Contains(t, []string{"BUY", "SELL", "HOLD", "WATCH"}, rec.Action)

		// For high sentiment news, verify it got BUY recommendation
		if rec.Sentiment > 0.3 && rec.RelevanceScore >= 0.5 {
			assert.Equal(t, "BUY", rec.Action)
		}
	}

	// Test that recommendations are cached
	for _, item := range newsItems {
		rec, found := cache.Get(item.Link)
		if rec.RelevanceScore >= 0.5 { // Only items with high relevance should be cached
			assert.True(t, found)
			assert.NotNil(t, rec)
		}
	}

	// Test that we got the expected number of recommendations
	// We expect 2 recommendations (RELIANCE and TCS) as they have high sentiment and relevance
	assert.Len(t, recommendations, 2)
}

func TestProcessor_GetRecommendationsByStock(t *testing.T) {
	cache := NewRecommendationCache(GetDefaultCacheConfig())
	processor := NewProcessor(cache)

	// Add some test recommendations
	recommendations := []Recommendation{
		{
			NewsItem: NewsItem{
				Title:       "RELIANCE announces strong Q3 results",
				Description: "Reliance Industries reported better-than-expected quarterly results.",
				Link:        "http://example.com/news1",
			},
			RelevanceScore: 0.8,
			Action:         "BUY",
			Confidence:     0.9,
		},
		{
			NewsItem: NewsItem{
				Title:       "TCS stock price target raised",
				Description: "Analysts raise TCS price target.",
				Link:        "http://example.com/news2",
			},
			RelevanceScore: 0.7,
			Action:         "BUY",
			Confidence:     0.8,
		},
	}

	for _, rec := range recommendations {
		cache.Set(rec.Link, rec)
	}

	// Test getting recommendations for RELIANCE
	relianceRecs := processor.GetRecommendationsByStock("RELIANCE")
	assert.Len(t, relianceRecs, 1)
	assert.Contains(t, relianceRecs[0].Title, "RELIANCE")

	// Test getting recommendations for TCS
	tcsRecs := processor.GetRecommendationsByStock("TCS")
	assert.Len(t, tcsRecs, 1)
	assert.Contains(t, tcsRecs[0].Title, "TCS")
}

func TestProcessor_GetLatestRecommendations(t *testing.T) {
	cache := NewRecommendationCache(GetDefaultCacheConfig())
	processor := NewProcessor(cache)

	// Add test recommendations with different timestamps
	now := time.Now()
	recommendations := []Recommendation{
		{
			NewsItem: NewsItem{
				Title: "Old news",
				Link:  "http://example.com/old",
			},
			LastUpdated: now.Add(-24 * time.Hour),
		},
		{
			NewsItem: NewsItem{
				Title: "Recent news",
				Link:  "http://example.com/recent",
			},
			LastUpdated: now,
		},
		{
			NewsItem: NewsItem{
				Title: "Middle news",
				Link:  "http://example.com/middle",
			},
			LastUpdated: now.Add(-12 * time.Hour),
		},
	}

	for _, rec := range recommendations {
		cache.Set(rec.Link, rec)
	}

	// Test getting latest recommendations
	latest := processor.GetLatestRecommendations(2)
	assert.Len(t, latest, 2)
	assert.Equal(t, "Recent news", latest[0].Title)
	assert.Equal(t, "Middle news", latest[1].Title)
}
