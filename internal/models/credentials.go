package models

import "time"

// Credentials represents the stored broker credentials
type Credentials struct {
	ID          int64
	UserID      string
	BrokerType  string
	APIKey      string
	APISecret   string
	AccessToken string
	TokenExpiry time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
