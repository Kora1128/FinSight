package middleware

import (
	"net/http"

	"github.com/Kora1128/FinSight/internal/cache"
	"github.com/Kora1128/FinSight/internal/models"
	"github.com/gin-gonic/gin"
)

// AuthConfig defines configuration options for the auth middleware
type AuthConfig struct {
	Cache        *cache.Cache
	RequireAny   bool // If true, user must be logged in to at least one broker
	RequireAll   bool // If true, user must be logged in to all brokers
	ZerodhaOnly  bool // If true, user must be logged in to Zerodha
	ICICIOnly    bool // If true, user must be logged in to ICICI Direct
	RequiredRole string // Reserved for future use with role-based auth
}

// Auth returns a middleware for checking user authentication
func Auth(config AuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		zerodhaToken, zerodhaLoggedIn := config.Cache.GetZerodhaToken()
		iciciToken, iciciLoggedIn := config.Cache.GetICICIToken()

		zerodhaValid := zerodhaLoggedIn && zerodhaToken != ""
		iciciValid := iciciLoggedIn && iciciToken != ""

		// Decide authentication status based on config
		var authenticated bool
		var requiredBroker string

		switch {
		case config.RequireAll:
			authenticated = zerodhaValid && iciciValid
			if !authenticated {
				if !zerodhaValid && !iciciValid {
					requiredBroker = "Zerodha and ICICI Direct"
				} else if !zerodhaValid {
					requiredBroker = "Zerodha"
				} else {
					requiredBroker = "ICICI Direct"
				}
			}
		case config.ZerodhaOnly:
			authenticated = zerodhaValid
			if !authenticated {
				requiredBroker = "Zerodha"
			}
		case config.ICICIOnly:
			authenticated = iciciValid
			if !authenticated {
				requiredBroker = "ICICI Direct"
			}
		case config.RequireAny:
			authenticated = zerodhaValid || iciciValid
			if !authenticated {
				requiredBroker = "at least one broker (Zerodha or ICICI Direct)"
			}
		default:
			authenticated = zerodhaValid || iciciValid // Default to RequireAny
			if !authenticated {
				requiredBroker = "at least one broker (Zerodha or ICICI Direct)"
			}
		}

		if !authenticated {
			errorMsg := "Authentication required"
			if requiredBroker != "" {
				errorMsg = "Authentication required for " + requiredBroker
			}
			
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{
				Success: false,
				Error:   errorMsg,
			})
			return
		}

		// Store authentication state in the context for potential use in handlers
		c.Set("zerodha_authenticated", zerodhaValid)
		c.Set("icici_authenticated", iciciValid)

		c.Next()
	}
}
