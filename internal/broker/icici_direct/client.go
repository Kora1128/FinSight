package icici_direct

import (
	"context"
	"errors"
	"time"

	"github.com/Kora1128/FinSight/internal/broker/types"
	"github.com/Kora1128/FinSight/internal/models"
	"github.com/Kora1128/icici-breezeconnect-go/breezeconnect"
	"github.com/Kora1128/icici-breezeconnect-go/breezeconnect/services"
)

// Ensure Client implements types.Client interface
var _ types.Client = (*Client)(nil)

// Client represents the ICICI Direct broker integration client
type Client struct {
	apiKey       string
	apiSecret    string
	requestToken string
	client       *breezeconnect.Client
	accessToken  string
	refreshToken string
	expiresAt    time.Time
}

// NewClient creates a new ICICI Direct client with the provided API key and secret
func NewClient(apiKey, apiSecret, requestToken string) *Client {
	client := breezeconnect.NewClient(apiKey, apiSecret)
	return &Client{
		apiKey:       apiKey,
		apiSecret:    apiSecret,
		requestToken: requestToken,
		client:       client,
	}
}

// Login authenticates the user with ICICI Direct using the provided request token and apiSecret
func (c *Client) Login() error {
	if c.requestToken == "" || c.apiSecret == "" {
		return errors.New("invalid request token or api secret")
	}
	customerService := services.NewCustomerService(c.client)
	resp, err := customerService.GetCustomerDetails(c.requestToken)
	if err == nil && resp != nil {
		c.accessToken = resp.Success.SessionToken    // Store the token
		c.expiresAt = time.Now().Add(12 * time.Hour) // ICICI tokens typically expire in 12 hours
	}
	return err
}

// CanAutoRefresh checks if the client can refresh the token automatically
func (c *Client) CanAutoRefresh() bool {
	return c.accessToken != "" && c.apiSecret != "" && time.Until(c.expiresAt) < 1*time.Hour
}

// RefreshToken attempts to refresh the authentication token
func (c *Client) RefreshToken() error {
	// Check if we need to refresh yet
	if time.Until(c.expiresAt) > 1*time.Hour {
		return nil // No need to refresh yet
	}

	if c.accessToken == "" || c.apiSecret == "" {
		return errors.New("missing credentials for token refresh")
	}

	// ICICI Direct may have a specific refresh token API
	// This is a simplified implementation
	customerService := services.NewCustomerService(c.client)
	_, err := customerService.GetCustomerDetails(c.accessToken)
	if err == nil {
		// Successfully refreshed, extend expiry
		c.expiresAt = time.Now().Add(12 * time.Hour)
	}
	return err
}

// SetAccessToken sets the access token for the client
func (c *Client) SetAccessToken(token string) {
	c.accessToken = token
	c.client.SetSessionKey(token)
}

// GetAccessToken returns the current access token
func (c *Client) GetAccessToken() string {
	return c.accessToken
}

// GetAPIKey returns the API key
func (c *Client) GetAPIKey() string {
	return c.apiKey
}

// GetHoldings fetches the current portfolio holdings from ICICI Direct and normalizes them into the common Holding struct
func (c *Client) GetHoldings(ctx context.Context) ([]models.Holding, error) {
	portfolioService := services.NewPortfolioService(c.client)
	holdings, err := portfolioService.GetPortfolioHoldings()
	if err != nil {
		return nil, err
	}

	var normalizedHoldings []models.Holding
	for _, h := range holdings {
		holdingType := models.HoldingTypeStock
		// TODO: Add logic to differentiate mutual funds if needed
		normalizedHolding := models.Holding{
			ItemName:         h.Symbol,
			ISIN:             h.ISIN,
			Quantity:         float64(h.Quantity),
			AveragePrice:     h.AveragePrice,
			LastTradedPrice:  h.LastTradedPrice,
			CurrentValue:     h.TotalValue,
			DayChange:        0, // Not available
			DayChangePercent: 0, // Not available
			TotalPnL:         h.PnL,
			Platform:         models.PlatformICICIDirect,
			Type:             holdingType,
			LastUpdated:      time.Now(),
		}
		normalizedHoldings = append(normalizedHoldings, normalizedHolding)
	}
	return normalizedHoldings, nil
}

// GetPositions fetches the current positions from ICICI Direct and normalizes them into the common Holding struct
func (c *Client) GetPositions(ctx context.Context) ([]models.Holding, error) {
	portfolioService := services.NewPortfolioService(c.client)
	positions, err := portfolioService.GetPositions()
	if err != nil {
		return nil, err
	}

	var normalizedHoldings []models.Holding
	for _, p := range positions {
		holdingType := models.HoldingTypeStock
		normalizedHolding := models.Holding{
			ItemName:         p.Symbol,
			ISIN:             "", // ISIN not available in Position struct
			Quantity:         float64(p.Quantity),
			AveragePrice:     p.AveragePrice,
			LastTradedPrice:  p.LastTradedPrice,
			CurrentValue:     p.TotalValue,
			DayChange:        0, // Not available
			DayChangePercent: 0, // Not available
			TotalPnL:         p.PnL,
			Platform:         models.PlatformICICIDirect,
			Type:             holdingType,
			LastUpdated:      time.Now(),
		}
		normalizedHoldings = append(normalizedHoldings, normalizedHolding)
	}
	return normalizedHoldings, nil
}
