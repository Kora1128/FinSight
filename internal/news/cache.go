package news

import (
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
)

// RecommendationCache handles caching of news recommendations
type RecommendationCache struct {
	cache    *cache.Cache
	mu       sync.RWMutex
	config   CacheConfig
	stopChan chan struct{}
}

// NewRecommendationCache creates a new recommendation cache
func NewRecommendationCache(config CacheConfig) *RecommendationCache {
	c := &RecommendationCache{
		cache:    cache.New(config.TTL, config.CleanupInterval),
		config:   config,
		stopChan: make(chan struct{}),
	}

	// Start cleanup goroutine
	go c.cleanup()

	return c
}

// Set adds a recommendation to the cache
func (c *RecommendationCache) Set(key string, recommendation Recommendation) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if we need to remove oldest items
	if c.cache.ItemCount() >= c.config.MaxItems {
		c.removeOldest()
	}

	c.cache.Set(key, recommendation, c.config.TTL)
}

// Get retrieves a recommendation from the cache
func (c *RecommendationCache) Get(key string) (Recommendation, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if item, found := c.cache.Get(key); found {
		return item.(Recommendation), true
	}
	return Recommendation{}, false
}

// GetAll returns all recommendations in the cache
func (c *RecommendationCache) GetAll() []Recommendation {
	c.mu.RLock()
	defer c.mu.RUnlock()

	items := c.cache.Items()
	recommendations := make([]Recommendation, 0, len(items))

	for _, item := range items {
		if rec, ok := item.Object.(Recommendation); ok {
			recommendations = append(recommendations, rec)
		}
	}

	return recommendations
}

// Remove removes a recommendation from the cache
func (c *RecommendationCache) Remove(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache.Delete(key)
}

// Clear removes all recommendations from the cache
func (c *RecommendationCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache.Flush()
}

// Close stops the cleanup goroutine
func (c *RecommendationCache) Close() {
	close(c.stopChan)
}

// removeOldest removes the oldest item from the cache
func (c *RecommendationCache) removeOldest() {
	items := c.cache.Items()
	var oldestKey string
	var oldestTime time.Time

	for key, item := range items {
		if oldestKey == "" || item.Expiration < oldestTime.UnixNano() {
			oldestKey = key
			oldestTime = time.Unix(0, item.Expiration)
		}
	}

	if oldestKey != "" {
		c.cache.Delete(oldestKey)
	}
}

// cleanup periodically removes expired items
func (c *RecommendationCache) cleanup() {
	ticker := time.NewTicker(c.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cache.DeleteExpired()
		case <-c.stopChan:
			return
		}
	}
}

// GetDefaultCacheConfig returns default cache configuration
func GetDefaultCacheConfig() CacheConfig {
	return CacheConfig{
		TTL:             24 * time.Hour,
		MaxItems:        1000,
		CleanupInterval: 1 * time.Hour,
	}
}
