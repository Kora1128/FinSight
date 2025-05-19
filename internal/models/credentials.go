package models

import "time"

// Credentials represents the stored broker credentials
type Credentials struct {
	ID          int64     `json:"id"`
	UserID      string    `json:"user_id"`
	BrokerType  string    `json:"broker_type"`
	APIKey      string    `json:"api_key"`
	APISecret   string    `json:"api_secret"`
	AccessToken string    `json:"access_token"`
	TokenExpiry time.Time `json:"token_expiry"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
