package news

import "time"

// Source represents a news source
type Source struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

// NewsItem represents a news article
type NewsItem struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Link        string    `json:"link"`
	Source      string    `json:"source"`
	Category    string    `json:"category"`
	PublishedAt time.Time `json:"published_at"`
	Sentiment   float64   `json:"sentiment"`
}

// Recommendation represents an investment recommendation based on news
type Recommendation struct {
	StockSymbol string    `json:"stock_symbol"`
	Action      string    `json:"action"` // BUY, SELL, HOLD, WATCH
	Confidence  float64   `json:"confidence"`
	Reason      string    `json:"reason"`
	NewsItem    NewsItem  `json:"news_item"`
	CreatedAt   time.Time `json:"created_at"`
}

// GetDefaultSources returns the default list of news sources
func GetDefaultSources() []Source {
	return []Source{
		{
			Name:        "MoneyControl",
			URL:         "https://www.moneycontrol.com/rss/business.xml",
			Description: "Business news from MoneyControl",
			Category:    "Business",
		},
		// {
		// 	Name:        "Economic Times",
		// 	URL:         "https://economictimes.indiatimes.com/rssfeedstopstories.cms",
		// 	Description: "Top stories from Economic Times",
		// 	Category:    "Business",
		// },
		{
			Name:        "Business Standard Markets",
			URL:         "https://www.business-standard.com/rss/markets-106.rss",
			Description: "Current topics from Business Standard",
			Category:    "Business",
		},
		{
			Name:        "Business Standard Stock Market",
			URL:         "https://www.business-standard.com/rss/markets/stock-market-news-10618.rss",
			Description: "Current topics from Business Standard",
			Category:    "Business",
		},
	}
}

// CacheConfig represents the configuration for caching recommendations
type CacheConfig struct {
	TTL             time.Duration
	MaxItems        int
	CleanupInterval time.Duration
}
