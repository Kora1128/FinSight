package broker

import (
	"context"

	"github.com/Kora1128/FinSight/internal/models"
)

// Client defines the interface that all broker clients must implement
type Client interface {
	// GetHoldings fetches the current portfolio holdings
	GetHoldings(ctx context.Context) ([]models.Holding, error)

	// GetPositions fetches the current positions
	GetPositions(ctx context.Context) ([]models.Holding, error)

	// Login authenticates the user with the broker
	Login(requestToken, apiSecret string) error
	
	// CanAutoRefresh checks if the client can refresh the token automatically
	CanAutoRefresh() bool
	
	// RefreshToken attempts to refresh the authentication token
	RefreshToken() error
	
	// GetAccessToken returns the current access token
	GetAccessToken() string
	
	// GetRefreshToken returns the current refresh token
	GetRefreshToken() string
	
	// SetRefreshToken sets the refresh token
	SetRefreshToken(token string)
	
	// GetLoginURL returns the login URL for OAuth flow
	GetLoginURL(redirectURI string) string
	
	// GetAPIKey returns the API key
	GetAPIKey() string
}
