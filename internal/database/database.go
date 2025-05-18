package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// DB represents the database connection
type DB struct {
	*sql.DB
}

// Config holds database configuration
type Config struct {
	ConnString string // PostgreSQL connection string
}

// New creates a new database connection
func New(config Config) (*DB, error) {
	db, err := sql.Open("postgres", config.ConnString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Initialize database
	if err = initDB(db); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return &DB{DB: db}, nil
}

// initDB creates the necessary tables if they don't exist
func initDB(db *sql.DB) error {
	// Create users table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			user_id TEXT PRIMARY KEY,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			last_accessed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	// Create sessions table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS sessions (
			session_id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			last_accessed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			expires_at TIMESTAMP NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users (user_id)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create sessions table: %w", err)
	}

	// Create broker_credentials table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS broker_credentials (
			id SERIAL PRIMARY KEY,
			user_id TEXT NOT NULL,
			broker_type TEXT NOT NULL,
			api_key TEXT NOT NULL,
			api_secret TEXT NOT NULL,
			access_token TEXT,
			token_expiry TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users (user_id),
			UNIQUE (user_id, broker_type)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create broker_credentials table: %w", err)
	}

	// Create portfolio_holdings table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS portfolio_holdings (
			id SERIAL PRIMARY KEY,
			user_id TEXT NOT NULL,
			item_name TEXT NOT NULL,
			isin TEXT,
			quantity REAL NOT NULL,
			average_price REAL NOT NULL,
			last_traded_price REAL NOT NULL,
			current_value REAL NOT NULL,
			day_change REAL NOT NULL,
			day_change_percent REAL NOT NULL,
			total_pnl REAL NOT NULL,
			platform TEXT NOT NULL,
			holding_type TEXT NOT NULL,
			last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users (user_id)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create portfolio_holdings table: %w", err)
	}

	log.Println("Database initialized successfully")
	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}
