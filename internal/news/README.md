# News Module

The news module provides functionality for fetching, processing, and analyzing financial news to generate investment recommendations. It includes sentiment analysis, relevance scoring, and caching mechanisms.

## Features

- RSS feed fetching from multiple financial news sources
- Sentiment analysis of news articles
- Relevance scoring for investment decisions
- Recommendation caching with TTL
- Stock-specific recommendation filtering
- Latest recommendations retrieval

## Usage

### 1. Initialize the News Engine

```go
import "your-project/internal/news"

// Create a cache configuration
cacheConfig := news.CacheConfig{
    TTL:             24 * time.Hour,    // Cache items expire after 24 hours
    MaxItems:        1000,              // Maximum number of items in cache
    CleanupInterval: 1 * time.Hour,     // Cleanup expired items every hour
}

// Create the recommendation cache
cache := news.NewRecommendationCache(cacheConfig)

// Create the news processor
processor := news.NewProcessor(cache)

// Create the RSS fetcher with default sources
fetcher := news.NewRSSFetcher(news.GetDefaultSources())
```

### 2. Fetch and Process News

```go
// Fetch news from all configured sources
newsItems, err := fetcher.FetchAll(context.Background())
if err != nil {
    // Handle error
}

// Process news items to generate recommendations
recommendations := processor.ProcessNews(context.Background(), newsItems)
```

### 3. Access Recommendations

```go
// Get recommendations for a specific stock
stockRecs := processor.GetRecommendationsByStock("RELIANCE")

// Get latest recommendations (e.g., last 5)
latestRecs := processor.GetLatestRecommendations(5)
```

### 4. Cache Management

```go
// Get a specific recommendation from cache
rec, found := cache.Get("news-item-url")
if found {
    // Use the recommendation
}

// Remove a specific recommendation
cache.Remove("news-item-url")

// Clear all recommendations
cache.Clear()

// Close the cache when done
cache.Close()
```

## Recommendation Structure

Each recommendation contains:

```go
type Recommendation struct {
    NewsItem        // Original news item
    RelevanceScore float64   // 0 to 1, indicating relevance for investment
    Action         string    // "BUY", "SELL", "HOLD", or "WATCH"
    Confidence     float64   // 0 to 1, indicating confidence in recommendation
    LastUpdated    time.Time // When the recommendation was last updated
}
```

## Configuration

### Cache Configuration

```go
type CacheConfig struct {
    TTL             time.Duration // Time-to-live for cached items
    MaxItems        int          // Maximum number of items in cache
    CleanupInterval time.Duration // How often to clean up expired items
}
```

### Default Values

- TTL: 24 hours
- MaxItems: 1000
- CleanupInterval: 1 hour

## News Sources

The module comes with pre-configured sources:

- MoneyControl
- Economic Times
- Business Standard

To add custom sources:

```go
customSources := []news.Source{
    {
        Name:        "Your Source",
        URL:         "https://your-source.com/rss",
        Description: "Description of your source",
        Category:    "Financial",
    },
}
fetcher := news.NewRSSFetcher(customSources)
```

## Best Practices

1. **Cache Management**
   - Set appropriate TTL based on your needs
   - Monitor cache size and adjust MaxItems accordingly
   - Call `cache.Close()` when shutting down

2. **Error Handling**
   - Always check for errors when fetching news
   - Handle cases where recommendations might be empty

3. **Performance**
   - Use appropriate context timeouts
   - Consider implementing rate limiting for RSS feeds
   - Monitor memory usage with large cache sizes

4. **Recommendation Usage**
   - Consider both relevance score and confidence
   - Use stock-specific recommendations for targeted analysis
   - Regularly update recommendations by fetching new news

## Example Flow

```go
func main() {
    // Initialize components
    cache := news.NewRecommendationCache(news.GetDefaultCacheConfig())
    processor := news.NewProcessor(cache)
    fetcher := news.NewRSSFetcher(news.GetDefaultSources())

    // Create a ticker to fetch news periodically
    ticker := time.NewTicker(1 * time.Hour)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            // Fetch and process news
            newsItems, err := fetcher.FetchAll(context.Background())
            if err != nil {
                log.Printf("Error fetching news: %v", err)
                continue
            }

            // Process news and get recommendations
            recommendations := processor.ProcessNews(context.Background(), newsItems)

            // Use recommendations
            for _, rec := range recommendations {
                if rec.RelevanceScore > 0.7 && rec.Confidence > 0.8 {
                    // Handle high-confidence recommendations
                    log.Printf("High confidence recommendation: %s - %s", rec.Title, rec.Action)
                }
            }
        }
    }
}
```

## Error Handling

The module provides several error cases to handle:

1. RSS Feed Errors
   - Network issues
   - Invalid feed format
   - Source unavailable

2. Processing Errors
   - Invalid news item format
   - Missing required fields

3. Cache Errors
   - Cache full
   - Item expiration

Always implement proper error handling in your application.

## Contributing

When adding new features or modifying existing ones:

1. Update the constants in `constants.go`
2. Add new tests in `processor_test.go`
3. Update this documentation
4. Follow the existing code style and patterns 