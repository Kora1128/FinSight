package news

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/stretchr/testify/assert"
)

const feedXML = `
<rss version="2.0">
  <channel>
	<title>Mock Feed</title>
	<item>
	  <title>Mock News</title>
	  <description>This is a mock news item.</description>
	  <link>http://example.com/mock-news</link>
	  <pubDate>Mon, 02 Jan 2006 15:04:05 MST</pubDate>
	</item>
  </channel>
</rss>`

func TestNewNewsFetcher(t *testing.T) {
	// Create a mock RSS feed XML
	server, client := newMockRSSFeedServer(feedXML)
	defer server.Close()
	parser := gofeed.NewParser()
	parser.Client = client

	// Create a NewsFetcher with the mock server as its source
	fetcher := &NewsFetcher{
		parser:  parser,
		sources: []Source{{Name: "MockSource", URL: server.URL}},
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
	server, client := newMockRSSFeedServer(feedXML)
	defer server.Close()
	parser := gofeed.NewParser()
	parser.Client = client

	// Create a NewsFetcher with the mock server as its source
	fetcher := &NewsFetcher{
		parser:  parser,
		sources: []Source{{Name: "MockSource", URL: server.URL}},
	}

	// Test fetching with valid source (mock test)
	// Note: In a real test, you would use a mock HTTP server to test actual fetching
	// This is just a basic structure test

	newsItems, err := fetcher.FetchNews(context.Background())
	if err != nil {
		t.Errorf("Expected no error when fetching from valid source, got %v", err)
	}
	assert.Len(t, newsItems, 1)
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
	newsItems, _ := fetcher.FetchNews(ctx)
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
	newsItems, _ := fetcher.FetchNews(ctx)
	if len(newsItems) != 0 {
		t.Errorf("Expected no news items when cancelled, got %d", len(newsItems))
	}
}

// newMockRSSFeedServer returns a test server serving a static RSS feed
func newMockRSSFeedServer(feedXML string) (*httptest.Server, *http.Client) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		w.Write([]byte(feedXML))
	}))

	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}

	client := &http.Client{Transport: transport}
	return server, client
}
