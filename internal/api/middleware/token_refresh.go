package middleware

import (
	"time"

	"github.com/Kora1128/FinSight/internal/cache"
	"github.com/gin-gonic/gin"
)

// TokenRefreshConfig defines configuration options for the token refresh middleware
type TokenRefreshConfig struct {
	Cache              *cache.Cache
	RefreshThreshold   time.Duration
	ZerodhaRefreshFunc func() error
	ICICIRefreshFunc   func() error
}

// TokenRefresh returns a middleware for automatic token refresh
func TokenRefresh(config TokenRefreshConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Default refresh threshold to 1 hour if not specified
		threshold := config.RefreshThreshold
		if threshold == 0 {
			threshold = 1 * time.Hour
		}

		// Check if Zerodha token needs refresh
		zerodhaToken, zerodhaFound := config.Cache.GetZerodhaToken()
		if zerodhaFound && zerodhaToken != "" {
			// In a real implementation, we would check token expiration time
			// For now, we'll assume refresh is needed based on configured functions
			if config.ZerodhaRefreshFunc != nil {
				err := config.ZerodhaRefreshFunc()
				if err == nil {
					// Token was refreshed, nothing to do
				}
			}
		}

		// Check if ICICI token needs refresh
		iciciToken, iciciFound := config.Cache.GetICICIToken()
		if iciciFound && iciciToken != "" {
			// Similar to above
			if config.ICICIRefreshFunc != nil {
				err := config.ICICIRefreshFunc()
				if err == nil {
					// Token was refreshed, nothing to do
				}
			}
		}

		c.Next()
	}
}
