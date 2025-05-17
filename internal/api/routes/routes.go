package routes

import (
	"github.com/Kora1128/FinSight/internal/api/handlers"
	"github.com/Kora1128/FinSight/internal/api/middleware"
	"github.com/Kora1128/FinSight/internal/cache"
	"github.com/gin-gonic/gin"
)

// SetupRouter configures the API routes
func SetupRouter(
	newsHandler *handlers.NewsHandler,
	portfolioHandler *handlers.PortfolioHandler,
	authHandler *handlers.AuthHandler,
	cache *cache.Cache,
) *gin.Engine {
	r := gin.New()

	// Middleware
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())

	// API Routes
	api := r.Group("/api/v1")
	{
		// Authentication routes
		auth := api.Group("/auth")
		{
			// Auth URL endpoints (for OAuth flow)
			auth.GET("/zerodha/url", authHandler.GetZerodhaAuthURL)
			auth.GET("/icici/url", authHandler.GetICICIAuthURL)
			
			// Login endpoints
			login := auth.Group("/login")
			{
				login.POST("/zerodha", authHandler.ZerodhaLogin)
				login.POST("/icici", authHandler.ICICILogin)
			}
			
			// Logout endpoints
			logout := auth.Group("/logout")
			{
				logout.POST("/zerodha", authHandler.LogoutZerodha)
				logout.POST("/icici", authHandler.LogoutICICI)
			}
			
			// Token refresh endpoint
			auth.POST("/refresh", authHandler.RefreshToken)
			
			// User status endpoint
			auth.GET("/status", authHandler.GetUserStatus)
		}

		// Portfolio routes - protected by authentication
		portfolio := api.Group("/portfolio")
		// Apply authentication middleware to portfolio routes
		portfolio.Use(middleware.Auth(middleware.AuthConfig{
			Cache:      cache,
			RequireAny: true, // Require login to at least one broker
		}))
		{
			portfolio.GET("", portfolioHandler.GetPortfolio)
			portfolio.POST("/refresh", portfolioHandler.RefreshPortfolio)
		}

		// News/Recommendation routes
		news := api.Group("/recommendations")
		{
			news.GET("", newsHandler.GetRecommendations)
			news.GET("/latest", newsHandler.GetLatestRecommendations)
			news.GET("/stock/:symbol", newsHandler.GetRecommendationsByStock)
		}

		// News sources routes
		sources := api.Group("/news/sources")
		{
			sources.GET("", newsHandler.GetSources)
			sources.POST("", newsHandler.AddSource)
			sources.DELETE("/:name", newsHandler.RemoveSource)
		}
	}

	return r
}
