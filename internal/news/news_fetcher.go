package news

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/mmcdole/gofeed"
)

// Common errors
var (
	ErrSourceExists   = errors.New("source already exists")
	ErrSourceNotFound = errors.New("source not found")
)

// NewsFetcher handles fetching news from RSS feeds
type NewsFetcher struct {
	parser    *gofeed.Parser
	sources   []Source
	sourcesMu sync.RWMutex
}

// NewNewsFetcher creates a new news fetcher
func NewNewsFetcher() *NewsFetcher {
	return &NewsFetcher{
		parser:  gofeed.NewParser(),
		sources: GetDefaultSources(),
	}
}

// GetSources returns all configured sources
func (f *NewsFetcher) GetSources() []Source {
	f.sourcesMu.RLock()
	defer f.sourcesMu.RUnlock()
	return f.sources
}

// AddSource adds a new source
func (f *NewsFetcher) AddSource(source Source) error {
	f.sourcesMu.Lock()
	defer f.sourcesMu.Unlock()

	// Check if source already exists
	for _, s := range f.sources {
		if s.Name == source.Name {
			return ErrSourceExists
		}
	}

	f.sources = append(f.sources, source)
	return nil
}

// RemoveSource removes a source by name
func (f *NewsFetcher) RemoveSource(name string) error {
	f.sourcesMu.Lock()
	defer f.sourcesMu.Unlock()

	for i, s := range f.sources {
		if s.Name == name {
			f.sources = append(f.sources[:i], f.sources[i+1:]...)
			return nil
		}
	}

	return ErrSourceNotFound
}

// FetchNews fetches news from all configured sources
func (f *NewsFetcher) FetchNews(ctx context.Context) ([]NewsItem, error) {
	f.sourcesMu.RLock()
	sources := f.sources
	f.sourcesMu.RUnlock()

	var allNews []NewsItem
	for _, source := range sources {
		feed, err := f.parser.ParseURLWithContext(source.URL, ctx)
		if err != nil {
			log.Printf("Error fetching from %s: %v", source.Name, err)
			continue
		}

		for _, item := range feed.Items {
			pubDate := time.Now()
			if item.PublishedParsed != nil {
				pubDate = *item.PublishedParsed
			}

			newsItem := NewsItem{
				Title:       item.Title,
				Description: item.Description,
				Link:        item.Link,
				Source:      source.Name,
				Category:    source.Category,
				PublishedAt: pubDate,
			}

			allNews = append(allNews, newsItem)
		}
	}

	return allNews, nil
}
