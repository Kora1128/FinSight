package icici_direct

import (
	"context"
	"errors"
	"time"

	"github.com/Kora1128/FinSight/internal/models"
	"github.com/Kora1128/icici-breezeconnect-go/breezeconnect"
	"github.com/Kora1128/icici-breezeconnect-go/breezeconnect/services"
)

// Client represents the ICICI Direct broker integration client
type Client struct {
	apiKey    string
	apiSecret string
	client    *breezeconnect.Client
}

// NewClient creates a new ICICI Direct client with the provided API key and secret
func NewClient(apiKey, apiSecret string) *Client {
	client := breezeconnect.NewClient(apiKey, apiSecret)
	return &Client{
		apiKey:    apiKey,
		apiSecret: apiSecret,
		client:    client,
	}
}

// Login authenticates the user with ICICI Direct using the provided request token and apiSecret
func (c *Client) Login(requestToken, apiSecret string) error {
	if requestToken == "" || apiSecret == "" {
		return errors.New("invalid request token or api secret")
	}
	customerService := services.NewCustomerService(c.client)
	_, err := customerService.GetCustomerDetails(requestToken)
	return err
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
