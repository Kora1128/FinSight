package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"
	
	"github.com/supabase-community/supabase-go"
)

// Config holds all configuration for the application
type Config struct {
	// Server configuration
	Port         string
	Environment  string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration

	// OpenAI configuration
	OpenAIAPIKey string

	// Cache configuration
	CacheTTL time.Duration

	// Database configuration (Supabase)
	SupabaseURL        string
	SupabaseAPIKey     string // Public API key for Supabase client
	SupabasePassword   string // Database password for direct PostgreSQL connections
	DBConnectionString string // For direct PostgreSQL connection
	SupabaseClient     *supabase.Client // Supabase client for easy API access

	// News configuration
	TrustedSources []string
}

// New creates a new Config instance with values from environment variables
func New() *Config {
	cfg := &Config{
		// Server configuration
		Port:         getEnv("APP_PORT", "8080"),
		Environment:  getEnv("APP_ENV", "development"),
		ReadTimeout:  getDurationEnv("APP_READ_TIMEOUT", 10*time.Second),
		WriteTimeout: getDurationEnv("APP_WRITE_TIMEOUT", 10*time.Second),

		// OpenAI configuration
		OpenAIAPIKey: getEnv("OPENAI_API_KEY", ""),

		// Cache configuration
		CacheTTL: getDurationEnv("CACHE_TTL", 15*time.Minute),

		// Database configuration (Supabase)
		SupabaseURL:      getEnv("SUPABASE_URL", ""),
		SupabaseAPIKey:   getEnv("SUPABASE_API_KEY", ""),
		SupabasePassword: getEnv("SUPABASE_PASSWORD", ""),

		// News configuration
		TrustedSources: getTrustedSources(),
	}

	// Initialize Supabase client if URL and API key are provided
	if cfg.SupabaseURL != "" && cfg.SupabaseAPIKey != "" {
		supaClient, err := supabase.NewClient(cfg.SupabaseURL, cfg.SupabaseAPIKey, nil)
		if err != nil {
			fmt.Printf("Warning: Failed to initialize Supabase client: %v\n", err)
		} else {
			cfg.SupabaseClient = supaClient
			fmt.Println("Supabase client initialized successfully")
		}
	} else {
		fmt.Println("Warning: SUPABASE_URL or SUPABASE_API_KEY is not set. Supabase client will not be available.")
	}
	
	// Construct PostgreSQL DSN from Supabase credentials for direct DB access
	if cfg.SupabaseURL != "" && cfg.SupabasePassword != "" {
		parsedURL, err := url.Parse(cfg.SupabaseURL)
		if err == nil && parsedURL.Host != "" {
			// Extract project reference from URL
			projectRef := strings.Split(parsedURL.Host, ".")[0]
			dbHost := fmt.Sprintf("db.%s.supabase.co", projectRef)
			
			// Standard Supabase PostgreSQL connection details
			// User: postgres
			// DB Name: postgres
			// Port: 5432
			cfg.DBConnectionString = fmt.Sprintf("postgresql://postgres:%s@%s:5432/postgres", url.QueryEscape(cfg.SupabasePassword), dbHost)
		} else {
			// Handle error or set a default/empty DSN if URL parsing fails
			fmt.Printf("Warning: Could not parse SUPABASE_URL ('%s') to construct DBConnectionString\n", cfg.SupabaseURL)
			cfg.DBConnectionString = "" // Or handle as a fatal error
		}
	} else {
		fmt.Println("Warning: SUPABASE_URL or SUPABASE_PASSWORD is not set. DBConnectionString will be empty.")
		cfg.DBConnectionString = "" // Or handle as a fatal error
	}

	return cfg
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getDurationEnv gets a duration from an environment variable or returns a default value
func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// getTrustedSources returns the list of trusted sources from environment variables
func getTrustedSources() []string {
	// Default trusted sources
	defaultSources := []string{
		"Economic Times",
		"Business Standard",
		"Moneycontrol",
		"Livemint",
		"Reuters India",
		"BloombergQuint",
		"HDFC Securities",
		"ICICI Direct Research",
		"SBI Securities",
		"Motilal Oswal",
	}

	// Get custom sources from environment variable
	if sources := getEnv("TRUSTED_SOURCES", ""); sources != "" {
		// Split by comma and trim spaces
		return strings.Split(sources, ",")
	}

	return defaultSources
}
