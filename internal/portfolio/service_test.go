package portfolio

import (
	"context"
	"testing"
	"time"

	"github.com/Kora1128/FinSight/internal/broker"
	"github.com/Kora1128/FinSight/internal/cache"
	"github.com/Kora1128/FinSight/internal/models"
	"github.com/stretchr/testify/assert"
)

// MockBrokerClient implements the broker.Client interface for testing
type MockBrokerClient struct {
	holdings  []models.Holding
	positions []models.Holding
	err       error
}

// Ensure MockBrokerClient implements broker.Client interface
var _ broker.Client = (*MockBrokerClient)(nil)

// GetHoldings returns mock holdings
func (m *MockBrokerClient) GetHoldings(ctx context.Context) ([]models.Holding, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.holdings, nil
}

// GetPositions returns mock positions
func (m *MockBrokerClient) GetPositions(ctx context.Context) ([]models.Holding, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.positions, nil
}

// Login implements the login method
func (m *MockBrokerClient) Login(requestToken, apiSecret string) error {
	return nil
}

// CanAutoRefresh implements the CanAutoRefresh method
func (m *MockBrokerClient) CanAutoRefresh() bool {
	return true
}

// RefreshToken implements the RefreshToken method
func (m *MockBrokerClient) RefreshToken() error {
	return nil
}

// GetAccessToken implements the GetAccessToken method
func (m *MockBrokerClient) GetAccessToken() string {
	return "mock-access-token"
}

// GetRefreshToken implements the GetRefreshToken method
func (m *MockBrokerClient) GetRefreshToken() string {
	return "mock-refresh-token"
}

// SetRefreshToken implements the SetRefreshToken method
func (m *MockBrokerClient) SetRefreshToken(token string) {
	// No-op in mock
}

// GetLoginURL implements the GetLoginURL method
func (m *MockBrokerClient) GetLoginURL(redirectURI string) string {
	return "https://mock-login-url.com"
}

// GetAPIKey implements the GetAPIKey method
func (m *MockBrokerClient) GetAPIKey() string {
	return "mock-api-key"
}

// TestGetPortfolio tests retrieving the portfolio
func TestGetPortfolio(t *testing.T) {
	// Create a cache
	c := cache.New(time.Hour, time.Hour)

	// Create mock holdings
	zerodhaHoldings := []models.Holding{
		{
			ItemName:         "INFY",
			ISIN:             "INE009A01021",
			Quantity:         10,
			AveragePrice:     1000,
			LastTradedPrice:  1100,
			CurrentValue:     11000,
			DayChange:        100,
			DayChangePercent: 1,
			TotalPnL:         1000,
			Platform:         models.PlatformZerodha,
			Type:             models.HoldingTypeStock,
			LastUpdated:      time.Now(),
		},
	}

	iciciHoldings := []models.Holding{
		{
			ItemName:         "TCS",
			ISIN:             "INE467B01029",
			Quantity:         5,
			AveragePrice:     2000,
			LastTradedPrice:  2200,
			CurrentValue:     11000,
			DayChange:        200,
			DayChangePercent: 2,
			TotalPnL:         1000,
			Platform:         models.PlatformICICIDirect,
			Type:             models.HoldingTypeStock,
			LastUpdated:      time.Now(),
		},
	}

	// Create mock clients
	zerodhaClient := &MockBrokerClient{
		holdings:  zerodhaHoldings,
		positions: []models.Holding{},
	}

	iciciClient := &MockBrokerClient{
		holdings:  iciciHoldings,
		positions: []models.Holding{},
	}

	// Create the service
	service := NewService(ServiceConfig{
		ZerodhaClient: zerodhaClient,
		ICICIClient:   iciciClient,
		Cache:         c,
		CacheTTL:      time.Hour,
	})

	// First, manually refresh the portfolio to populate the cache
	ctx := context.Background()
	err := service.RefreshPortfolio(ctx)
	assert.NoError(t, err)
	
	// Now get the portfolio from cache
	portfolio, err := service.GetPortfolio(ctx, false, "")
	assert.NoError(t, err)
	assert.NotNil(t, portfolio)
	assert.Equal(t, 2, len(portfolio.Holdings))
	assert.Equal(t, 22000.0, portfolio.TotalValue)
	assert.Equal(t, 300.0, portfolio.TotalDayChange)
	assert.Equal(t, 2000.0, portfolio.TotalPnL)

	// Test filtering by stock type
	stockPortfolio, err := service.GetPortfolio(ctx, false, models.HoldingTypeStock)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(stockPortfolio.Holdings))

	// Test forcing refresh
	forcedPortfolio, err := service.GetPortfolio(ctx, true, "")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(forcedPortfolio.Holdings))
}

// TestMergeHoldings tests the merging of holdings from different sources
func TestMergeHoldings(t *testing.T) {
	now := time.Now()
	earlier := now.Add(-time.Hour)

	// Create test holdings
	holdingsA := []models.Holding{
		{
			ItemName:         "INFY",
			ISIN:             "INE009A01021",
			Quantity:         10,
			AveragePrice:     1000,
			LastTradedPrice:  1100,
			CurrentValue:     11000,
			Platform:         models.PlatformZerodha,
			Type:             models.HoldingTypeStock,
			LastUpdated:      earlier,
		},
	}

	holdingsB := []models.Holding{
		{
			ItemName:         "TCS",
			ISIN:             "INE467B01029",
			Quantity:         5,
			AveragePrice:     2000,
			LastTradedPrice:  2200,
			CurrentValue:     11000,
			Platform:         models.PlatformICICIDirect,
			Type:             models.HoldingTypeStock,
			LastUpdated:      now,
		},
	}

	holdingsC := []models.Holding{
		{
			ItemName:         "Infosys", // Different name but same ISIN
			ISIN:             "INE009A01021",
			Quantity:         15,
			AveragePrice:     1050,
			LastTradedPrice:  1150,
			CurrentValue:     17250,
			Platform:         models.PlatformICICIDirect,
			Type:             models.HoldingTypeStock,
			LastUpdated:      now, // More recent
		},
	}

	merged := mergeHoldings(holdingsA, holdingsB, holdingsC)
	
	// Should have 2 unique holdings (merged A+C by ISIN, and B)
	assert.Equal(t, 2, len(merged))

	// Find INFY/Infosys holding
	var infyHolding models.Holding
	var tcsHolding models.Holding
	
	for _, h := range merged {
		if h.ISIN == "INE009A01021" {
			infyHolding = h
		} else if h.ISIN == "INE467B01029" {
			tcsHolding = h
		}
	}

	// Verify INFY merged properly (should use more recent data from holdingsC)
	assert.Equal(t, "Infosys", infyHolding.ItemName)
	assert.Equal(t, 15.0, infyHolding.Quantity)
	assert.Equal(t, models.PlatformICICIDirect, infyHolding.Platform)
	
	// Verify TCS data
	assert.Equal(t, "TCS", tcsHolding.ItemName)
	assert.Equal(t, 5.0, tcsHolding.Quantity)
	assert.Equal(t, models.PlatformICICIDirect, tcsHolding.Platform)
}

// TestRefreshPortfolio tests the portfolio refresh functionality
func TestRefreshPortfolio(t *testing.T) {
	// Create a cache
	c := cache.New(time.Hour, time.Hour)

	// Create mock holdings
	zerodhaHoldings := []models.Holding{
		{
			ItemName:         "INFY",
			ISIN:             "INE009A01021",
			Quantity:         10,
			AveragePrice:     1000,
			LastTradedPrice:  1100,
			CurrentValue:     11000,
			Platform:         models.PlatformZerodha,
			Type:             models.HoldingTypeStock,
			LastUpdated:      time.Now(),
		},
	}

	// Create mock clients
	zerodhaClient := &MockBrokerClient{
		holdings:  zerodhaHoldings,
		positions: []models.Holding{},
	}

	iciciClient := &MockBrokerClient{
		holdings:  []models.Holding{},
		positions: []models.Holding{},
	}

	// Create the service
	service := NewService(ServiceConfig{
		ZerodhaClient: zerodhaClient,
		ICICIClient:   iciciClient,
		Cache:         c,
		CacheTTL:      time.Hour,
	})

	// Test refresh
	ctx := context.Background()
	err := service.RefreshPortfolio(ctx)
	assert.NoError(t, err)

	// Check that portfolio was cached
	portfolio, found := c.GetPortfolio()
	assert.True(t, found)
	assert.Equal(t, 1, len(portfolio.Holdings))
	assert.Equal(t, "INFY", portfolio.Holdings[0].ItemName)
}
