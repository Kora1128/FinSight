package database

import (
	"database/sql"
	"errors"
	"time"

	"github.com/Kora1128/FinSight/internal/models"
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
func (r *UserRepo) CreateUser(userID string) error {
	_, err := r.db.Exec(
		"INSERT INTO users (user_id) VALUES ($1)",
		userID,
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
		"SELECT session_id, user_id, created_at, last_accessed_at, expires_at FROM sessions WHERE session_id = $1",
		sessionID,
	).Scan(&session.SessionID, &session.UserID, &session.CreatedAt, &session.LastAccessedAt, &session.ExpiresAt)

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
		"SELECT session_id, user_id, created_at, last_accessed_at, expires_at FROM sessions WHERE user_id = $1",
		userID,
	).Scan(&session.SessionID, &session.UserID, &session.CreatedAt, &session.LastAccessedAt, &session.ExpiresAt)

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
