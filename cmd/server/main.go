package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Kora1128/FinSight/internal/api/handlers"
	"github.com/Kora1128/FinSight/internal/api/middleware"
	"github.com/Kora1128/FinSight/internal/news"
)

func main() {
	// Initialize news engine components
	cache := news.NewRecommendationCache(news.CacheConfig{
		MaxItems:        1000,
		TTL:             24 * time.Hour,
		CleanupInterval: 1 * time.Hour,
	})
	processor := news.NewProcessor(cache)
	fetcher := news.NewNewsFetcher()

	// Set up background context for periodic news fetching
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start periodic news fetching
	go func() {
		ticker := time.NewTicker(15 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				newsItems, err := fetcher.FetchNews(ctx)
				if err != nil {
					log.Printf("Error fetching news: %v", err)
					continue
				}

				recommendations := processor.ProcessNews(ctx, newsItems)
				log.Printf("Processed %d news items, generated %d recommendations", len(newsItems), len(recommendations))
			}
		}
	}()

	// Initialize Gin router
	router := gin.Default()

	// Add middleware
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())
	router.Use(middleware.CORS())

	// Create news handler
	newsHandler := handlers.NewNewsHandler(processor, fetcher)

	// Set up API routes
	api := router.Group("/api/v1")
	{
		// News recommendations
		api.GET("/recommendations", newsHandler.GetRecommendations)
		api.GET("/recommendations/latest", newsHandler.GetLatestRecommendations)
		api.GET("/recommendations/stock/:symbol", newsHandler.GetRecommendationsByStock)

		// News sources
		api.GET("/sources", newsHandler.GetSources)
		api.POST("/sources", newsHandler.AddSource)
		api.DELETE("/sources/:name", newsHandler.RemoveSource)
	}

	// Create HTTP server
	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Graceful shutdown
	log.Println("Shutting down server...")
	cancel() // Stop news fetching

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
