package zerodha

import (
	"context"
	"errors"
	"time"

	"github.com/Kora1128/FinSight/internal/broker/types"
	"github.com/Kora1128/FinSight/internal/models"
	kiteconnect "github.com/zerodha/gokiteconnect/v4"
)

// Ensure Client implements types.Client interface
var _ types.Client = (*Client)(nil)

// Client represents the Zerodha broker integration client
type Client struct {
	kc           *kiteconnect.Client
	apiKey       string
	apiSecret    string
	requestToken string
	accessToken  string
	refreshToken string
	expiresAt    time.Time
}

// NewClient creates a new Zerodha client with the provided API key and secret
func NewClient(apiKey, apiSecret, requestToken string) *Client {
	kc := kiteconnect.New(apiKey)
	return &Client{
		kc:           kc,
		apiKey:       apiKey,
		requestToken: requestToken,
		apiSecret:    apiSecret,
	}
}

// Login authenticates the user with Zerodha using the provided request token and apiSecret
func (c *Client) Login() error {
	user, err := c.kc.GenerateSession(c.requestToken, c.apiSecret)
	if err != nil {
		return err
	}
	c.kc.SetAccessToken(user.AccessToken)
	c.accessToken = user.AccessToken
	c.refreshToken = user.RefreshToken
	c.expiresAt = time.Now().Add(24 * time.Hour) // Zerodha tokens typically expire after 24 hours
	return nil
}

// CanAutoRefresh checks if the client can refresh the token automatically
func (c *Client) CanAutoRefresh() bool {
	return c.refreshToken != "" && c.apiSecret != ""
}

// RefreshToken attempts to refresh the authentication token
func (c *Client) RefreshToken() error {
	// Check if token is about to expire (within 1 hour)
	if time.Until(c.expiresAt) > 1*time.Hour {
		return nil // No need to refresh yet
	}

	// Zerodha doesn't have a direct refresh token mechanism in their API
	// We would need to re-authenticate or use the refresh token to get a new access token
	// This is a simplified implementation
	if c.refreshToken == "" {
		return errors.New("no refresh token available")
	}

	// In a real implementation, we would call Zerodha's API to refresh the token
	// For now, we'll simulate a successful refresh by extending the expiry
	c.expiresAt = time.Now().Add(24 * time.Hour)
	return nil
}

// SetAccessToken sets the access token for the client
func (c *Client) SetAccessToken(token string) {
	c.accessToken = token
	c.kc.SetAccessToken(token)
}

// GetAccessToken returns the current access token
func (c *Client) GetAccessToken() string {
	return c.accessToken
}

// GetAPIKey returns the API key
func (c *Client) GetAPIKey() string {
	return c.apiKey
}

// GetHoldings fetches the current portfolio holdings from Zerodha and normalizes them into the common Holding struct
func (c *Client) GetHoldings(ctx context.Context) ([]models.Holding, error) {
	holdings, err := c.kc.GetHoldings()
	if err != nil {
		return nil, err
	}

	var normalizedHoldings []models.Holding
	for _, h := range holdings {
		holdingType := models.HoldingTypeStock
		// TODO: Add logic to differentiate mutual funds if needed
		normalizedHolding := models.Holding{
			ItemName:         h.Tradingsymbol,
			ISIN:             h.ISIN,
			Quantity:         float64(h.Quantity),
			AveragePrice:     h.AveragePrice,
			LastTradedPrice:  h.LastPrice,
			CurrentValue:     float64(h.Quantity) * h.LastPrice,
			DayChange:        h.DayChange,
			DayChangePercent: h.DayChangePercentage,
			TotalPnL:         h.PnL,
			Platform:         models.PlatformZerodha,
			Type:             holdingType,
			LastUpdated:      time.Now(),
		}
		normalizedHoldings = append(normalizedHoldings, normalizedHolding)
	}
	return normalizedHoldings, nil
}

// GetPositions fetches the current positions from Zerodha and normalizes them into the common Holding struct
func (c *Client) GetPositions(ctx context.Context) ([]models.Holding, error) {
	positions, err := c.kc.GetPositions()
	if err != nil {
		return nil, err
	}

	var normalizedHoldings []models.Holding
	for _, p := range positions.Net {
		holdingType := models.HoldingTypeStock
		normalizedHolding := models.Holding{
			ItemName:         p.Tradingsymbol,
			ISIN:             "", // ISIN not available in Position struct
			Quantity:         float64(p.Quantity),
			AveragePrice:     p.AveragePrice,
			LastTradedPrice:  p.LastPrice,
			CurrentValue:     float64(p.Quantity) * p.LastPrice,
			DayChange:        p.M2M,
			DayChangePercent: 0, // Not available
			TotalPnL:         p.PnL,
			Platform:         models.PlatformZerodha,
			Type:             holdingType,
			LastUpdated:      time.Now(),
		}
		normalizedHoldings = append(normalizedHoldings, normalizedHolding)
	}
	return normalizedHoldings, nil
}
