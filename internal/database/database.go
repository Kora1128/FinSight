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
	
	// Run migrations if needed
	if err = migrateTables(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return &DB{DB: db}, nil
}

// migrateTables performs any necessary migrations for existing tables
func migrateTables(db *sql.DB) error {
	// Check if email column exists in users table
	var emailExists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 
			FROM information_schema.columns 
			WHERE table_name='users' AND column_name='email'
		)
	`).Scan(&emailExists)

	if err != nil {
		return fmt.Errorf("failed to check for email column: %w", err)
	}

	// If email column doesn't exist in an existing table, add it
	if !emailExists {
		log.Println("Migrating users table to add email column")
		
		// First add the column allowing nulls temporarily
		_, err = db.Exec(`
			ALTER TABLE users 
			ADD COLUMN email TEXT
		`)
		
		if err != nil {
			return fmt.Errorf("failed to add email column: %w", err)
		}
		
		// Generate default emails for existing users based on their user_id
		_, err = db.Exec(`
			UPDATE users 
			SET email = CONCAT(user_id, '@temp_migration.com')
			WHERE email IS NULL
		`)
		
		if err != nil {
			return fmt.Errorf("failed to add default emails: %w", err)
		}
		
		// Now add the NOT NULL and UNIQUE constraints
		_, err = db.Exec(`
			ALTER TABLE users 
			ALTER COLUMN email SET NOT NULL;
			
			CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email
			ON users(email);
		`)
		
		if err != nil {
			return fmt.Errorf("failed to set constraints on email column: %w", err)
		}
		
		log.Println("Users table migration completed successfully")
	}
	
	return nil
}

// initDB creates the necessary tables if they don't exist
func initDB(db *sql.DB) error {
	// Create users table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			user_id TEXT PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
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

	// Perform any necessary migrations
	if err = migrateTables(db); err != nil {
		return fmt.Errorf("failed to migrate tables: %w", err)
	}

	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}
