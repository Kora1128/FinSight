package models

import "time"

// AuthToken represents an authentication token with metadata
type AuthToken struct {
	Token        string    `json:"token"`
	RefreshToken string    `json:"refreshToken,omitempty"`
	ExpiresAt    time.Time `json:"expiresAt"`
	TokenType    string    `json:"tokenType,omitempty"` // e.g., "Bearer"
}

// TokenRefreshRequest represents a request to refresh an authentication token
type TokenRefreshRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
	BrokerType   string `json:"brokerType" binding:"required"` // "zerodha" or "icici"
}

// AuthURLResponse represents a response containing an authentication URL
type AuthURLResponse struct {
	Success bool   `json:"success"`
	URL     string `json:"url,omitempty"`
	Error   string `json:"error,omitempty"`
}

// TokenResponse represents a response containing authentication tokens
type TokenResponse struct {
	Success      bool      `json:"success"`
	Token        string    `json:"token,omitempty"`
	RefreshToken string    `json:"refreshToken,omitempty"`
	ExpiresAt    time.Time `json:"expiresAt,omitempty"`
	Error        string    `json:"error,omitempty"`
}
