package portfolio

import (
	"time"

	"github.com/Kora1128/FinSight/internal/models"
)

// PortfolioRepository defines the interface for storing and retrieving portfolio data
type PortfolioRepository interface {
	// SaveHoldings saves portfolio holdings to the repository for a user
	SaveHoldings(userID string, holdings []models.Holding) error

	// GetHoldings retrieves portfolio holdings for a user
	GetHoldings(userID string) ([]models.Holding, error)

	// GetPlatformHoldings retrieves portfolio holdings for a user filtered by platform
	GetPlatformHoldings(userID string, platform string) ([]models.Holding, error)

	// GetHoldingsByType retrieves portfolio holdings for a user filtered by holding type
	GetHoldingsByType(userID string, holdingType models.HoldingType) ([]models.Holding, error)

	// GetPortfolioLastUpdated gets the timestamp when the portfolio was last updated
	GetPortfolioLastUpdated(userID string) (time.Time, bool, error)
}
