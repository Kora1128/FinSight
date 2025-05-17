package routes

import (
	"github.com/Kora1128/FinSight/internal/api/handlers"
	"github.com/Kora1128/FinSight/internal/api/middleware"
	"github.com/gin-gonic/gin"
)

// SetupRouter configures the API routes
func SetupRouter(
	newsHandler *handlers.NewsHandler,
	portfolioHandler *handlers.PortfolioHandler,
	authHandler *handlers.AuthHandler,
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
		auth := api.Group("/login")
		{
			auth.POST("/zerodha", authHandler.ZerodhaLogin)
			auth.POST("/icici", authHandler.ICICILogin)
		}

		// User status
		api.GET("/user/status", authHandler.GetUserStatus)

		// Portfolio routes
		portfolio := api.Group("/portfolio")
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
