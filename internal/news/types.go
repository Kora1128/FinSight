package news

import "time"

// Source represents a news source
type Source struct {
	Name        string
	URL         string
	Description string
	Category    string
}

// NewsItem represents a single news article
type NewsItem struct {
	Title       string
	Description string
	Link        string
	PublishedAt time.Time
	Source      Source
	Categories  []string
	Keywords    []string
	Sentiment   float64 // -1 to 1, where -1 is negative, 0 is neutral, 1 is positive
}

// Recommendation represents a filtered and processed news item
type Recommendation struct {
	NewsItem
	RelevanceScore float64 // 0 to 1, indicating how relevant this news is for investment decisions
	Action         string  // e.g., "BUY", "SELL", "HOLD", "WATCH"
	Confidence     float64 // 0 to 1, indicating confidence in the recommendation
	LastUpdated    time.Time
}

// CacheConfig represents the configuration for caching recommendations
type CacheConfig struct {
	TTL             time.Duration
	MaxItems        int
	CleanupInterval time.Duration
}
