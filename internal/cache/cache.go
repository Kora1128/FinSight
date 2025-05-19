package cache

import (
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
)

// Cache represents the application's cache
type Cache struct {
	*cache.Cache
	mu sync.RWMutex
}

// New creates a new Cache instance
func New(defaultExpiration, cleanupInterval time.Duration) *Cache {
	return &Cache{
		Cache: cache.New(defaultExpiration, cleanupInterval),
	}
}

// Cache keys
const (
	KeyPortfolio       = "portfolio"
	KeyRecommendations = "recommendations"
	KeyZerodhaToken    = "zerodha_token"
	KeyICICIToken      = "icici_direct_token"
)

// ClearAll clears all cached data
func (c *Cache) ClearAll() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Flush()
}

// Delete removes an item from the cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Cache.Delete(key)
}
