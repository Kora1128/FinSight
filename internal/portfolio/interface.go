package portfolio

import (
	"context"

	"github.com/Kora1128/FinSight/internal/models"
)

// ServiceInterface defines the interface that the portfolio service must implement
type ServiceInterface interface {
	// GetPortfolio retrieves the aggregated portfolio
	GetPortfolio(ctx context.Context, forceRefresh bool, holdingType models.HoldingType) (*models.Portfolio, error)

	// RefreshPortfolio forces a refresh of portfolio data from brokers
	RefreshPortfolio(ctx context.Context) error
}
