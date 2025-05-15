package zerodha

import (
	"context"
	"time"

	"github.com/Kora1128/FinSight/internal/models"
	kiteconnect "github.com/zerodha/gokiteconnect/v4"
)

// Client represents the Zerodha broker integration client
type Client struct {
	kc *kiteconnect.Client
}

// NewClient creates a new Zerodha client with the provided API key and secret
func NewClient(apiKey, apiSecret string) *Client {
	kc := kiteconnect.New(apiKey)
	return &Client{kc: kc}
}

// Login authenticates the user with Zerodha using the provided request token and apiSecret
func (c *Client) Login(requestToken, apiSecret string) error {
	user, err := c.kc.GenerateSession(requestToken, apiSecret)
	if err != nil {
		return err
	}
	c.kc.SetAccessToken(user.AccessToken)
	return nil
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
