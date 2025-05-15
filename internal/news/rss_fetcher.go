package news

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/mmcdole/gofeed"
)

// RSSFetcher handles fetching and parsing RSS feeds
type RSSFetcher struct {
	client  *http.Client
	parser  *gofeed.Parser
	sources []Source
}

// NewRSSFetcher creates a new RSS fetcher
func NewRSSFetcher(sources []Source) *RSSFetcher {
	return &RSSFetcher{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		parser:  gofeed.NewParser(),
		sources: sources,
	}
}

// FetchAll fetches news from all configured sources
func (f *RSSFetcher) FetchAll(ctx context.Context) ([]NewsItem, error) {
	var allNews []NewsItem

	for _, source := range f.sources {
		news, err := f.FetchSource(ctx, source)
		if err != nil {
			// Log error but continue with other sources
			fmt.Printf("Error fetching from %s: %v\n", source.Name, err)
			continue
		}
		allNews = append(allNews, news...)
	}

	return allNews, nil
}

// FetchSource fetches news from a specific source
func (f *RSSFetcher) FetchSource(ctx context.Context, source Source) ([]NewsItem, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", source.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching feed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	feed, err := f.parser.ParseString(string(body))
	if err != nil {
		return nil, fmt.Errorf("error parsing feed: %w", err)
	}

	var newsItems []NewsItem
	for _, item := range feed.Items {
		publishedAt := time.Now()
		if item.PublishedParsed != nil {
			publishedAt = *item.PublishedParsed
		}

		newsItem := NewsItem{
			Title:       item.Title,
			Description: item.Description,
			Link:        item.Link,
			PublishedAt: publishedAt,
			Source:      source,
			Categories:  item.Categories,
		}
		newsItems = append(newsItems, newsItem)
	}

	return newsItems, nil
}

// GetDefaultSources returns a list of default financial news sources
func GetDefaultSources() []Source {
	return []Source{
		{
			Name:        "MoneyControl",
			URL:         "https://www.moneycontrol.com/rss/latestnews.xml",
			Description: "Latest news from MoneyControl",
			Category:    "Financial",
		},
		{
			Name:        "Economic Times",
			URL:         "https://economictimes.indiatimes.com/markets/rssfeeds/1977021501.cms",
			Description: "Market news from Economic Times",
			Category:    "Financial",
		},
		{
			Name:        "Business Standard",
			URL:         "https://www.business-standard.com/rss/markets-1010101.xml",
			Description: "Market news from Business Standard",
			Category:    "Financial",
		},
	}
}
