package routes

import (
	"github.com/Kora1128/FinSight/internal/api/handlers"
	"github.com/Kora1128/FinSight/internal/api/middleware"
	"github.com/Kora1128/FinSight/internal/cache"
	"github.com/Kora1128/FinSight/internal/database"
	"github.com/gin-gonic/gin"
)

// SetupRouter configures the API routes
func SetupRouter(
	newsHandler *handlers.NewsHandler,
	userPortfolioHandler *handlers.UserPortfolioHandler,
	sessionHandler *handlers.SessionHandler,
	cache *cache.Cache,
	sessionRepo *database.SessionRepo,
	userRepo *database.UserRepo,
) *gin.Engine {
	r := gin.New()

	// Middleware
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())

	// API Routes
	api := r.Group("/api/v1")
	{
		// User session routes
		sessions := api.Group("/sessions")
		{
			// Create a new session
			sessions.POST("", sessionHandler.CreateSession)
			
			// Get session info
			sessions.GET("/:userId", sessionHandler.GetSession)
			
			// Connect broker to session
			sessions.POST("/connect", sessionHandler.ConnectBroker)
			
			// Disconnect broker from session
			sessions.POST("/disconnect/:userId/:brokerType", sessionHandler.DisconnectBroker)
		}

		// User-specific portfolio routes - protected by session authentication
		userPortfolio := api.Group("/users/:userId/portfolio")
		userPortfolio.Use(middleware.SessionAuth(middleware.SessionAuthConfig{
			SessionRepo: sessionRepo,
			UserRepo:    userRepo,
		}))
		{
			userPortfolio.GET("", userPortfolioHandler.GetUserPortfolio)
			userPortfolio.POST("/refresh", userPortfolioHandler.RefreshUserPortfolio)
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
