package models

import (
	"time"

	"github.com/google/uuid"
)

// UserSession represents a user session with broker credentials
type UserSession struct {
	UserID           string    `json:"userId"`
	SessionID        string    `json:"sessionId"`
	ZerodhaConnected bool      `json:"zerodhaConnected"`
	ICICIConnected   bool      `json:"iciciConnected"`
	CreatedAt        time.Time `json:"createdAt"`
	LastAccessedAt   time.Time `json:"lastAccessedAt"`
	ExpiresAt        time.Time `json:"expiresAt"`
}

// NewUserSession creates a new user session
func NewUserSession(sessionDuration time.Duration) *UserSession {
	userID := uuid.New().String()
	now := time.Now()
	return &UserSession{
		UserID:         userID,
		SessionID:      uuid.New().String(),
		CreatedAt:      now,
		LastAccessedAt: now,
		ExpiresAt:      now.Add(sessionDuration),
	}
}

// IsValid checks if the session is still valid
func (s *UserSession) IsValid() bool {
	return time.Now().Before(s.ExpiresAt)
}

// Touch updates the last accessed time
func (s *UserSession) Touch() {
	s.LastAccessedAt = time.Now()
}

// SessionInfo represents the public session information
type SessionInfo struct {
	UserID           string    `json:"userId"`
	ZerodhaConnected bool      `json:"zerodhaConnected"`
	ICICIConnected   bool      `json:"iciciConnected"`
	ExpiresAt        time.Time `json:"expiresAt"`
}

// GetInfo returns the public session information
func (s *UserSession) GetInfo() SessionInfo {
	return SessionInfo{
		UserID:           s.UserID,
		ZerodhaConnected: s.ZerodhaConnected,
		ICICIConnected:   s.ICICIConnected,
		ExpiresAt:        s.ExpiresAt,
	}
}

// UserCredentials represents user credentials for broker authentication
type UserCredentials struct {
	UserID       string `json:"userId"`
	BrokerType   string `json:"brokerType" binding:"required,oneof=zerodha icici_direct"`
	APIKey       string `json:"apiKey" binding:"required"`
	APISecret    string `json:"apiSecret" binding:"required"`
	RequestToken string `json:"requestToken"`
	Password     string `json:"password,omitempty"` // For ICICI Direct
}

// SessionResponse represents the response for session-related endpoints
type SessionResponse struct {
	Success bool        `json:"success"`
	Data    SessionInfo `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}
