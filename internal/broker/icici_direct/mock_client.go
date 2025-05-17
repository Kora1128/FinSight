package icici_direct

import (
	"context"
	"time"

	"github.com/Kora1128/FinSight/internal/models"
)

// MockClient represents a mock ICICI Direct client for testing
type MockClient struct {
	// Configurable responses
	LoginError     error
	Holdings       []models.Holding
	HoldingsError  error
	Positions      []models.Holding
	PositionsError error
}

// Ensure MockClient implements ICICIClient interface
var _ ICICIClient = (*MockClient)(nil)

// NewMockClient creates a new mock ICICI Direct client
func NewMockClient() *MockClient {
	return &MockClient{}
}

// Login simulates the login process
func (m *MockClient) Login(requestToken, apiSecret string) error {
	return m.LoginError
}

// GetHoldings returns mock holdings data
func (m *MockClient) GetHoldings(ctx context.Context) ([]models.Holding, error) {
	if m.HoldingsError != nil {
		return nil, m.HoldingsError
	}
	return m.Holdings, nil
}

// GetPositions returns mock positions data
func (m *MockClient) GetPositions(ctx context.Context) ([]models.Holding, error) {
	if m.PositionsError != nil {
		return nil, m.PositionsError
	}
	return m.Positions, nil
}

// WithMockHoldings sets mock holdings data
func (m *MockClient) WithMockHoldings(holdings []models.Holding) *MockClient {
	m.Holdings = holdings
	return m
}

// WithMockPositions sets mock positions data
func (m *MockClient) WithMockPositions(positions []models.Holding) *MockClient {
	m.Positions = positions
	return m
}

// WithLoginError sets a mock login error
func (m *MockClient) WithLoginError(err error) *MockClient {
	m.LoginError = err
	return m
}

// WithHoldingsError sets a mock holdings error
func (m *MockClient) WithHoldingsError(err error) *MockClient {
	m.HoldingsError = err
	return m
}

// WithPositionsError sets a mock positions error
func (m *MockClient) WithPositionsError(err error) *MockClient {
	m.PositionsError = err
	return m
}

// GetDefaultMockHoldings returns a set of default mock holdings for testing
func GetDefaultMockHoldings() []models.Holding {
	return []models.Holding{
		{
			ItemName:         "RELIANCE",
			ISIN:             "INE002A01018",
			Quantity:         10,
			AveragePrice:     2500.0,
			LastTradedPrice:  2600.0,
			CurrentValue:     26000.0,
			DayChange:        100.0,
			DayChangePercent: 4.0,
			TotalPnL:         1000.0,
			Platform:         models.PlatformICICIDirect,
			Type:             models.HoldingTypeStock,
			LastUpdated:      time.Now(),
		},
		{
			ItemName:         "TCS",
			ISIN:             "INE467B01029",
			Quantity:         5,
			AveragePrice:     3500.0,
			LastTradedPrice:  3600.0,
			CurrentValue:     18000.0,
			DayChange:        500.0,
			DayChangePercent: 2.86,
			TotalPnL:         500.0,
			Platform:         models.PlatformICICIDirect,
			Type:             models.HoldingTypeStock,
			LastUpdated:      time.Now(),
		},
	}
}

// GetDefaultMockPositions returns a set of default mock positions for testing
func GetDefaultMockPositions() []models.Holding {
	return []models.Holding{
		{
			ItemName:         "INFY",
			ISIN:             "INE009A01021",
			Quantity:         20,
			AveragePrice:     1500.0,
			LastTradedPrice:  1550.0,
			CurrentValue:     31000.0,
			DayChange:        1000.0,
			DayChangePercent: 3.33,
			TotalPnL:         1000.0,
			Platform:         models.PlatformICICIDirect,
			Type:             models.HoldingTypeStock,
			LastUpdated:      time.Now(),
		},
	}
}
