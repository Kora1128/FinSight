package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Kora1128/FinSight/internal/api/handlers"
	"github.com/Kora1128/FinSight/internal/api/routes"
	"github.com/Kora1128/FinSight/internal/broker"
	"github.com/Kora1128/FinSight/internal/cache"
	"github.com/Kora1128/FinSight/internal/config"
	"github.com/Kora1128/FinSight/internal/database"
	"github.com/Kora1128/FinSight/internal/news"
	"github.com/Kora1128/FinSight/internal/portfolio"
	"github.com/joho/godotenv" // Import the package
)

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Error loading .env file, using environment variables from OS")
	}

	// Load configuration
	cfg := config.New()

	// Initialize cache
	appCache := cache.New(cfg.CacheTTL, time.Hour)

	// Initialize database
	db, err := database.New(database.Config{
		ConnString: cfg.DBConnectionString,
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Log Supabase client status
	if cfg.SupabaseClient != nil {
		log.Println("Supabase client is available for use")
	} else {
		log.Println("Warning: Supabase client is not initialized")
	}

	// Initialize news engine components
	newsCache := news.NewRecommendationCache(news.CacheConfig{
		MaxItems:        1000,
		TTL:             24 * time.Hour,
		CleanupInterval: 1 * time.Hour,
	})
	processor := news.NewProcessor(newsCache, cfg.OpenAIAPIKey)
	fetcher := news.NewNewsFetcher()

	// Initialize repositories
	sessionRepo := database.NewSessionRepo(db)
	brokerCredentialsRepo := database.NewBrokerCredentialsRepo(db)
	portfolioRepo := database.NewPortfolioRepo(db)

	// Initialize broker manager
	brokerManager := broker.NewBrokerManager(brokerCredentialsRepo, appCache, 24*time.Hour, 1*time.Hour)

	// Initialize user portfolio service
	userPortfolioService := portfolio.NewUserService(portfolio.UserServiceConfig{
		BrokerManager:       brokerManager,
		PortfolioRepository: portfolioRepo,
	})

	// Set up background context for periodic news fetching
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start periodic news fetching
	go func() {
		ticker := time.NewTicker(3 * time.Hour)
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

	// Create handlers
	newsHandler := handlers.NewNewsHandler(processor, fetcher)
	userPortfolioHandler := handlers.NewUserPortfolioHandler(userPortfolioService)
	userRepo := database.NewUserRepo(db)
	sessionHandler := handlers.NewSessionHandler(
		appCache,
		sessionRepo,
		userRepo,
		brokerManager,
		24*time.Hour,
	)

	// Initialize router with routes
	router := routes.SetupRouter(
		newsHandler,
		userPortfolioHandler,
		sessionHandler,
		appCache, // Still keeping this for now in case other handlers need it
		sessionRepo,
		userRepo,
	)

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	// Start server in a goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Graceful shutdown
	log.Println("Shutting down server...")
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exiting")
}
