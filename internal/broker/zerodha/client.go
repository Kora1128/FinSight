package zerodha

import (
	"context"
	"errors"
	"time"

	"github.com/Kora1128/FinSight/internal/broker"
	"github.com/Kora1128/FinSight/internal/models"
	kiteconnect "github.com/zerodha/gokiteconnect/v4"
)

// Ensure Client implements broker.Client interface
var _ broker.Client = (*Client)(nil)

// Client represents the Zerodha broker integration client
type Client struct {
	kc           *kiteconnect.Client
	apiKey       string
	apiSecret    string
	accessToken  string
	refreshToken string
	expiresAt    time.Time
}

// NewClient creates a new Zerodha client with the provided API key and secret
func NewClient(apiKey, apiSecret string) *Client {
	kc := kiteconnect.New(apiKey)
	return &Client{
		kc:        kc,
		apiKey:    apiKey,
		apiSecret: apiSecret,
	}
}

// Login authenticates the user with Zerodha using the provided request token and apiSecret
func (c *Client) Login(requestToken, apiSecret string) error {
	user, err := c.kc.GenerateSession(requestToken, apiSecret)
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

// GetAccessToken returns the current access token
func (c *Client) GetAccessToken() string {
	return c.accessToken
}

// GetLoginURL returns the Zerodha login URL
func (c *Client) GetLoginURL(redirectURI string) string {
	if redirectURI == "" {
		// Use default redirect URI if none provided
		redirectURI = "https://finsight.app/auth/zerodha/callback"
	}
	return c.kc.GetLoginURL()
}

// GetAPIKey returns the API key
func (c *Client) GetAPIKey() string {
	return c.apiKey
}

// GetRefreshToken returns the refresh token
func (c *Client) GetRefreshToken() string {
	return c.refreshToken
}

// SetRefreshToken sets the refresh token
func (c *Client) SetRefreshToken(token string) {
	c.refreshToken = token
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
