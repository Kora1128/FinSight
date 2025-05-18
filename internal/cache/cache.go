package cache

import (
	"sync"
	"time"

	"github.com/Kora1128/FinSight/internal/models"
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
	KeyICICIToken      = "icici_token"
)

// SetPortfolio stores the portfolio data in cache
func (c *Cache) SetPortfolio(portfolio *models.Portfolio) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Set(KeyPortfolio, portfolio, cache.DefaultExpiration)
}

// GetPortfolio retrieves the portfolio data from cache
func (c *Cache) GetPortfolio() (*models.Portfolio, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if x, found := c.Get(KeyPortfolio); found {
		return x.(*models.Portfolio), true
	}
	return nil, false
}

// SetRecommendations stores the recommendations data in cache
func (c *Cache) SetRecommendations(recommendations []models.Recommendation) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Set(KeyRecommendations, recommendations, cache.DefaultExpiration)
}

// GetRecommendations retrieves the recommendations data from cache
func (c *Cache) GetRecommendations() ([]models.Recommendation, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if x, found := c.Get(KeyRecommendations); found {
		return x.([]models.Recommendation), true
	}
	return nil, false
}

// SetZerodhaToken stores the Zerodha access token in cache
func (c *Cache) SetZerodhaToken(token string, expiration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Set(KeyZerodhaToken, token, expiration)
}

// GetZerodhaToken retrieves the Zerodha access token from cache
func (c *Cache) GetZerodhaToken() (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if x, found := c.Get(KeyZerodhaToken); found {
		return x.(string), true
	}
	return "", false
}

// SetICICIToken stores the ICICI Direct session token in cache
func (c *Cache) SetICICIToken(token string, expiration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Set(KeyICICIToken, token, expiration)
}

// GetICICIToken retrieves the ICICI Direct session token from cache
func (c *Cache) GetICICIToken() (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if x, found := c.Get(KeyICICIToken); found {
		return x.(string), true
	}
	return "", false
}

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
