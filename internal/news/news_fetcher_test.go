package news

import (
	"context"
	"testing"
	"time"
)

func TestNewNewsFetcher(t *testing.T) {
	fetcher := NewNewsFetcher()
	if fetcher == nil {
		t.Error("Expected non-nil fetcher")
	}
	if len(fetcher.sources) == 0 {
		t.Error("Expected default sources to be initialized")
	}
}

func TestAddSource(t *testing.T) {
	fetcher := NewNewsFetcher()

	// Test adding a valid source
	err := fetcher.AddSource(Source{
		Name: "Test Source",
		URL:  "http://example.com/feed",
	})
	if err != nil {
		t.Errorf("Expected no error when adding valid source, got %v", err)
	}

	// Test adding duplicate source
	err = fetcher.AddSource(Source{
		Name: "Test Source",
		URL:  "http://example.com/feed",
	})
	if err != ErrSourceExists {
		t.Errorf("Expected ErrSourceExists when adding duplicate source, got %v", err)
	}

	// Test adding source with empty name
	err = fetcher.AddSource(Source{
		Name: "",
		URL:  "http://example.com/feed",
	})
	if err == nil {
		t.Error("Expected error when adding source with empty name")
	}

	// Test adding source with empty URL
	err = fetcher.AddSource(Source{
		Name: "Test Source",
		URL:  "",
	})
	if err == nil {
		t.Error("Expected error when adding source with empty URL")
	}
}

func TestRemoveSource(t *testing.T) {
	fetcher := NewNewsFetcher()

	// Add a test source
	err := fetcher.AddSource(Source{
		Name: "Test Source",
		URL:  "http://example.com/feed",
	})
	if err != nil {
		t.Fatalf("Failed to add test source: %v", err)
	}

	// Test removing existing source
	err = fetcher.RemoveSource("Test Source")
	if err != nil {
		t.Errorf("Expected no error when removing existing source, got %v", err)
	}

	// Test removing non-existent source
	err = fetcher.RemoveSource("Non-existent Source")
	if err != ErrSourceNotFound {
		t.Errorf("Expected ErrSourceNotFound when removing non-existent source, got %v", err)
	}
}

func TestGetSources(t *testing.T) {
	fetcher := NewNewsFetcher()

	// Add test sources
	testSources := []Source{
		{
			Name: "Source 1",
			URL:  "http://example.com/feed1",
		},
		{
			Name: "Source 2",
			URL:  "http://example.com/feed2",
		},
	}

	for _, source := range testSources {
		err := fetcher.AddSource(source)
		if err != nil {
			t.Fatalf("Failed to add test source %s: %v", source.Name, err)
		}
	}

	// Get sources
	sources := fetcher.GetSources()

	// Verify all test sources are present
	for _, testSource := range testSources {
		found := false
		for _, source := range sources {
			if source.Name == testSource.Name && source.URL == testSource.URL {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected to find source %s with URL %s", testSource.Name, testSource.URL)
		}
	}
}

func TestFetchNews(t *testing.T) {
	fetcher := NewNewsFetcher()

	// Test fetching with no sources
	newsItems, err := fetcher.FetchNews(context.Background())
	if err != nil {
		t.Errorf("Expected no error when fetching with no sources, got %v", err)
	}
	if len(newsItems) != 0 {
		t.Errorf("Expected no news items when no sources, got %d", len(newsItems))
	}

	// Test fetching with invalid source
	err = fetcher.AddSource(Source{
		Name: "Invalid Source",
		URL:  "http://invalid-url",
	})
	if err != nil {
		t.Fatalf("Failed to add invalid source: %v", err)
	}

	newsItems, err = fetcher.FetchNews(context.Background())
	if err == nil {
		t.Error("Expected error when fetching from invalid source")
	}
	if len(newsItems) != 0 {
		t.Errorf("Expected no news items from invalid source, got %d", len(newsItems))
	}

	// Test fetching with valid source (mock test)
	// Note: In a real test, you would use a mock HTTP server to test actual fetching
	// This is just a basic structure test
	err = fetcher.AddSource(Source{
		Name: "Valid Source",
		URL:  "http://example.com/valid-feed",
	})
	if err != nil {
		t.Fatalf("Failed to add valid source: %v", err)
	}

	newsItems, err = fetcher.FetchNews(context.Background())
	if err != nil {
		t.Errorf("Expected no error when fetching from valid source, got %v", err)
	}
	// Note: In a real test, you would verify the content of newsItems
}

func TestFetchNewsWithTimeout(t *testing.T) {
	fetcher := NewNewsFetcher()

	// Add a source that will timeout
	err := fetcher.AddSource(Source{
		Name: "Timeout Source",
		URL:  "http://example.com/timeout",
	})
	if err != nil {
		t.Fatalf("Failed to add timeout source: %v", err)
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Test fetching with timeout
	newsItems, err := fetcher.FetchNews(ctx)
	if err == nil {
		t.Error("Expected error when fetching with timeout")
	}
	if len(newsItems) != 0 {
		t.Errorf("Expected no news items when timeout, got %d", len(newsItems))
	}
}

func TestFetchNewsWithCancellation(t *testing.T) {
	fetcher := NewNewsFetcher()

	// Add a source
	err := fetcher.AddSource(Source{
		Name: "Test Source",
		URL:  "http://example.com/feed",
	})
	if err != nil {
		t.Fatalf("Failed to add test source: %v", err)
	}

	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Test fetching with cancelled context
	newsItems, err := fetcher.FetchNews(ctx)
	if err == nil {
		t.Error("Expected error when fetching with cancelled context")
	}
	if len(newsItems) != 0 {
		t.Errorf("Expected no news items when cancelled, got %d", len(newsItems))
	}
}
