package portfolio

import (
	"context"
	"time"

	"github.com/Kora1128/FinSight/internal/broker"
	"github.com/Kora1128/FinSight/internal/models"
)

// UserServiceConfig holds configuration for user-specific portfolio service
type UserServiceConfig struct {
	BrokerManager       *broker.BrokerManager
	PortfolioRepository PortfolioRepository
}

// UserService manages portfolios for specific users
type UserService struct {
	brokerManager       *broker.BrokerManager
	portfolioRepository PortfolioRepository
}

// NewUserService creates a new user-specific portfolio service
func NewUserService(config UserServiceConfig) *UserService {
	return &UserService{
		brokerManager:       config.BrokerManager,
		portfolioRepository: config.PortfolioRepository,
	}
}

// GetPortfolio retrieves the portfolio for a specific user
func (s *UserService) GetPortfolio(ctx context.Context, userID string, forceRefresh bool, holdingType models.HoldingType) (*models.Portfolio, error) {
	// Check if we need to refresh
	if forceRefresh {
		if err := s.RefreshPortfolio(ctx, userID); err != nil {
			return nil, err
		}
	}

	// Get portfolio from database
	var holdings []models.Holding
	var err error

	if holdingType != "" && holdingType != "all" {
		holdings, err = s.portfolioRepository.GetHoldingsByType(userID, holdingType)
	} else {
		holdings, err = s.portfolioRepository.GetHoldings(userID)
	}

	if err != nil {
		return nil, err
	}

	// If there are no holdings, return empty portfolio
	if len(holdings) == 0 {
		return &models.Portfolio{
			Holdings:          []models.Holding{},
			TotalValue:        0,
			TotalDayChange:    0,
			TotalDayChangePct: 0,
			TotalPnL:          0,
			LastUpdated:       time.Now(),
		}, nil
	}

	// Calculate portfolio totals
	totalValue := 0.0
	totalDayChange := 0.0
	totalPnL := 0.0

	for _, holding := range holdings {
		totalValue += holding.CurrentValue
		totalDayChange += holding.DayChange
		totalPnL += holding.TotalPnL
	}

	// Create portfolio object
	portfolio := &models.Portfolio{
		Holdings:       holdings,
		TotalValue:     totalValue,
		TotalDayChange: totalDayChange,
		TotalPnL:       totalPnL,
		LastUpdated:    time.Now(),
	}

	// Calculate percent change safely
	if totalValue > 0 {
		portfolio.TotalDayChangePct = (totalDayChange / totalValue) * 100
	}

	return portfolio, nil
}

// RefreshPortfolio updates the portfolio for a specific user
func (s *UserService) RefreshPortfolio(ctx context.Context, userID string) error {
	allHoldings := []models.Holding{}

	// Get zerodha client if available
	if zerodhaClient, exists := s.brokerManager.GetClient(userID, broker.ClientTypeZerodha); exists {
		zerodhaHoldings, err := zerodhaClient.GetHoldings(ctx)
		if err == nil {
			// Update platform info
			for i := range zerodhaHoldings {
				zerodhaHoldings[i].Platform = models.PlatformZerodha
				zerodhaHoldings[i].LastUpdated = time.Now()
			}
			allHoldings = append(allHoldings, zerodhaHoldings...)
		}

		zerodhaPositions, err := zerodhaClient.GetPositions(ctx)
		if err == nil {
			// Update platform info
			for i := range zerodhaPositions {
				zerodhaPositions[i].Platform = models.PlatformZerodha
				zerodhaPositions[i].LastUpdated = time.Now()
			}
			allHoldings = append(allHoldings, zerodhaPositions...)
		}
	}

	// Get ICICI Direct client if available
	if iciciClient, exists := s.brokerManager.GetClient(userID, broker.ClientTypeICICIDirect); exists {
		iciciHoldings, err := iciciClient.GetHoldings(ctx)
		if err == nil {
			// Update platform info
			for i := range iciciHoldings {
				iciciHoldings[i].Platform = models.PlatformICICIDirect
				iciciHoldings[i].LastUpdated = time.Now()
			}
			allHoldings = append(allHoldings, iciciHoldings...)
		}

		iciciPositions, err := iciciClient.GetPositions(ctx)
		if err == nil {
			// Update platform info
			for i := range iciciPositions {
				iciciPositions[i].Platform = models.PlatformICICIDirect
				iciciPositions[i].LastUpdated = time.Now()
			}
			allHoldings = append(allHoldings, iciciPositions...)
		}
	}

	// Merge holdings with the same ISIN
	mergedHoldings := mergeHoldings(allHoldings)

	// Save to database
	if err := s.portfolioRepository.SaveHoldings(userID, mergedHoldings); err != nil {
		return err
	}

	return nil
}

// Helper function to merge holdings with the same ISIN
func mergeHoldings(holdings []models.Holding) []models.Holding {
	// Group by ISIN
	holdingsByISIN := make(map[string][]models.Holding)
	for _, holding := range holdings {
		if holding.ISIN != "" {
			holdingsByISIN[holding.ISIN] = append(holdingsByISIN[holding.ISIN], holding)
		} else {
			// If no ISIN, we will use a combination of name and platform as key
			key := holding.ItemName + "_" + holding.Platform
			holdingsByISIN[key] = append(holdingsByISIN[key], holding)
		}
	}

	// Merge holdings with the same ISIN
	var mergedHoldings []models.Holding
	for _, holdingsGroup := range holdingsByISIN {
		if len(holdingsGroup) == 1 {
			mergedHoldings = append(mergedHoldings, holdingsGroup[0])
			continue
		}

		// Merge multiple holdings with the same ISIN
		merged := models.Holding{
			ItemName:    holdingsGroup[0].ItemName,
			ISIN:        holdingsGroup[0].ISIN,
			Platform:    holdingsGroup[0].Platform,
			Type:        holdingsGroup[0].Type,
			LastUpdated: time.Now(),
		}

		for _, h := range holdingsGroup {
			merged.Quantity += h.Quantity
			merged.CurrentValue += h.CurrentValue
			merged.DayChange += h.DayChange
			merged.TotalPnL += h.TotalPnL
		}

		// Recalculate average price and day change percent
		if merged.Quantity > 0 {
			merged.AveragePrice = merged.CurrentValue / merged.Quantity
		}
		if merged.CurrentValue > 0 {
			merged.DayChangePercent = (merged.DayChange / (merged.CurrentValue - merged.DayChange)) * 100
		}

		mergedHoldings = append(mergedHoldings, merged)
	}

	return mergedHoldings
}
