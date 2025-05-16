package config

import (
	"os"
	"strings"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	// Server configuration
	Port         string
	Environment  string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration

	// Zerodha API configuration
	ZerodhaAPIKey    string
	ZerodhaAPISecret string

	// ICICI Direct API configuration
	ICICIAPIKey    string
	ICICIAPISecret string
	ICICIPassword  string

	// OpenAI configuration
	OpenAIAPIKey string

	// Cache configuration
	CacheTTL time.Duration

	// News configuration
	NewsRefreshInterval time.Duration
	TrustedSources      []string
}

// New creates a new Config instance with values from environment variables
func New() *Config {
	config := &Config{
		// Server configuration
		Port:         getEnv("APP_PORT", "8080"),
		Environment:  getEnv("APP_ENV", "development"),
		ReadTimeout:  getDurationEnv("APP_READ_TIMEOUT", 10*time.Second),
		WriteTimeout: getDurationEnv("APP_WRITE_TIMEOUT", 10*time.Second),

		// Zerodha API configuration
		ZerodhaAPIKey:    getEnv("ZERODHA_API_KEY", ""),
		ZerodhaAPISecret: getEnv("ZERODHA_API_SECRET", ""),

		// ICICI Direct API configuration
		ICICIAPIKey:    getEnv("ICICI_API_KEY", ""),
		ICICIAPISecret: getEnv("ICICI_API_SECRET", ""),
		ICICIPassword:  getEnv("ICICI_PASSWORD", ""),

		// OpenAI configuration
		OpenAIAPIKey: getEnv("OPENAI_API_KEY", ""),

		// Cache configuration
		CacheTTL: getDurationEnv("CACHE_TTL", 15*time.Minute),

		// News configuration
		NewsRefreshInterval: getDurationEnv("NEWS_REFRESH_INTERVAL", 24*time.Hour),
		TrustedSources:      getTrustedSources(),
	}

	return config
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
