package types

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
	
	// RefreshToken refreshes the access token
	RefreshToken() error
	
	// GetAccessToken returns the current access token
	GetAccessToken() string
	
	// GetRefreshToken returns the refresh token
	GetRefreshToken() string
}
