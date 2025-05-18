package broker

import (
	"time"
)

// CredentialsRepository defines the interface for storing and retrieving broker credentials
type CredentialsRepository interface {
	// SaveCredentials saves broker credentials to the repository
	SaveCredentials(userID string, brokerType string, apiKey string, apiSecret string) error

	// GetCredentials retrieves broker credentials from the repository
	GetCredentials(userID string, brokerType string) (*Credentials, error)

	// UpdateAccessToken updates the access token and expiry time for broker credentials
	UpdateAccessToken(userID string, brokerType string, accessToken string, expiryTime time.Time) error

	// HasCredentials checks if the user has credentials for a specific broker
	HasCredentials(userID string, brokerType string) (bool, error)

	// DeleteCredentials deletes broker credentials from the repository
	DeleteCredentials(userID string, brokerType string) error

	// GetCredentialsForAllUsers retrieves all broker credentials from the repository
	GetCredentialsForAllUsers() ([]*Credentials, error)

	// GetExpiredTokens retrieves credentials with expired tokens
	GetExpiredTokens() ([]*Credentials, error)
}
