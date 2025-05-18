package database

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Kora1128/FinSight/internal/models"
	"github.com/google/uuid"
)

// UserRepo handles user operations in the database
type UserRepo struct {
	db *DB
}

// NewUserRepo creates a new user repository
func NewUserRepo(db *DB) *UserRepo {
	return &UserRepo{db: db}
}

// CreateUser creates a new user in the database
func (r *UserRepo) CreateUser(userID string, email string) error {
	_, err := r.db.Exec(
		"INSERT INTO users (user_id, email) VALUES ($1, $2)",
		userID, email,
	)
	return err
}

// GetUser retrieves a user from the database
func (r *UserRepo) GetUser(userID string) (bool, error) {
	var id string
	err := r.db.QueryRow(
		"SELECT user_id FROM users WHERE user_id = $1",
		userID,
	).Scan(&id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// GetUserByEmail checks if a user exists with the given email
func (r *UserRepo) GetUserByEmail(email string) (string, bool, error) {
	var userID string
	err := r.db.QueryRow(
		"SELECT user_id FROM users WHERE email = $1",
		email,
	).Scan(&userID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, nil
		}
		return "", false, err
	}

	return userID, true, nil
}

// FindOrCreateUserByEmail finds a user by email or creates one if not found
func (r *UserRepo) FindOrCreateUserByEmail(email string) (string, error) {
	// First try to find the user
	userID, exists, err := r.GetUserByEmail(email)
	if err != nil {
		return "", fmt.Errorf("error checking for existing user: %w", err)
	}
	
	if exists {
		return userID, nil
	}
	
	// User doesn't exist, create a new one
	newUserID := uuid.New().String()
	if err := r.CreateUser(newUserID, email); err != nil {
		return "", fmt.Errorf("error creating new user: %w", err)
	}
	
	return newUserID, nil
}

// UpdateLastAccessed updates the last_accessed_at field for a user
func (r *UserRepo) UpdateLastAccessed(userID string) error {
	_, err := r.db.Exec(
		"UPDATE users SET last_accessed_at = $1 WHERE user_id = $2",
		time.Now(), userID,
	)
	return err
}

// SessionRepo handles session operations in the database
type SessionRepo struct {
	db *DB
}

// NewSessionRepo creates a new session repository
func NewSessionRepo(db *DB) *SessionRepo {
	return &SessionRepo{db: db}
}

// CreateSession creates a new session in the database
func (r *SessionRepo) CreateSession(session *models.UserSession) error {
	_, err := r.db.Exec(
		"INSERT INTO sessions (session_id, user_id, expires_at) VALUES ($1, $2, $3)",
		session.SessionID, session.UserID, session.ExpiresAt,
	)
	return err
}

// GetSession retrieves a session from the database
func (r *SessionRepo) GetSession(sessionID string) (*models.UserSession, error) {
	session := &models.UserSession{}
	err := r.db.QueryRow(
		`SELECT s.session_id, s.user_id, s.created_at, s.last_accessed_at, s.expires_at, u.email 
		FROM sessions s
		JOIN users u ON s.user_id = u.user_id
		WHERE s.session_id = $1`,
		sessionID,
	).Scan(&session.SessionID, &session.UserID, &session.CreatedAt, &session.LastAccessedAt, &session.ExpiresAt, &session.Email)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	// Get broker connections
	brokerRepo := NewBrokerCredentialsRepo(r.db)
	zerodhaConnected, err := brokerRepo.HasCredentials(session.UserID, models.PlatformZerodha)
	if err != nil {
		return nil, err
	}
	
	iciciConnected, err := brokerRepo.HasCredentials(session.UserID, models.PlatformICICIDirect)
	if err != nil {
		return nil, err
	}
	
	session.ZerodhaConnected = zerodhaConnected
	session.ICICIConnected = iciciConnected

	return session, nil
}

// GetUserSession retrieves a session by user ID from the database
func (r *SessionRepo) GetUserSession(userID string) (*models.UserSession, error) {
	session := &models.UserSession{}
	err := r.db.QueryRow(
		`SELECT s.session_id, s.user_id, s.created_at, s.last_accessed_at, s.expires_at, u.email 
		FROM sessions s
		JOIN users u ON s.user_id = u.user_id
		WHERE s.user_id = $1`,
		userID,
	).Scan(&session.SessionID, &session.UserID, &session.CreatedAt, &session.LastAccessedAt, &session.ExpiresAt, &session.Email)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	// Get broker connections
	brokerRepo := NewBrokerCredentialsRepo(r.db)
	zerodhaConnected, err := brokerRepo.HasCredentials(session.UserID, models.PlatformZerodha)
	if err != nil {
		return nil, err
	}
	
	iciciConnected, err := brokerRepo.HasCredentials(session.UserID, models.PlatformICICIDirect)
	if err != nil {
		return nil, err
	}
	
	session.ZerodhaConnected = zerodhaConnected
	session.ICICIConnected = iciciConnected

	return session, nil
}

// UpdateLastAccessed updates the last_accessed_at field for a session
func (r *SessionRepo) UpdateLastAccessed(sessionID string) error {
	_, err := r.db.Exec(
		"UPDATE sessions SET last_accessed_at = $1 WHERE session_id = $2",
		time.Now(), sessionID,
	)
	return err
}

// DeleteSession deletes a session from the database
func (r *SessionRepo) DeleteSession(sessionID string) error {
	_, err := r.db.Exec(
		"DELETE FROM sessions WHERE session_id = $1",
		sessionID,
	)
	return err
}
