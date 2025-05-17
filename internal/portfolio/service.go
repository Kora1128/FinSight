package portfolio

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Kora1128/FinSight/internal/broker"
	"github.com/Kora1128/FinSight/internal/cache"
	"github.com/Kora1128/FinSight/internal/models"
)

// Ensure Service implements ServiceInterface
var _ ServiceInterface = (*Service)(nil)

// Service handles portfolio operations
type Service struct {
	zerodhaClient   broker.Client
	iciciClient     broker.Client
	cache           *cache.Cache
	cacheTTL        time.Duration
	lastRefreshTime time.Time
	mu              sync.RWMutex
}

// ServiceConfig holds configuration for the portfolio service
type ServiceConfig struct {
	ZerodhaClient broker.Client
	ICICIClient   broker.Client
	Cache         *cache.Cache
	CacheTTL      time.Duration
}

// NewService creates a new portfolio service
func NewService(config ServiceConfig) *Service {
	return &Service{
		zerodhaClient:   config.ZerodhaClient,
		iciciClient:     config.ICICIClient,
		cache:           config.Cache,
		cacheTTL:        config.CacheTTL,
		lastRefreshTime: time.Now(),
	}
}

// GetPortfolio retrieves the aggregated portfolio from cache or refreshes if needed
func (s *Service) GetPortfolio(ctx context.Context, forceRefresh bool, holdingType models.HoldingType) (*models.Portfolio, error) {
	s.mu.RLock()
	timeSinceLastRefresh := time.Since(s.lastRefreshTime)
	needsRefresh := forceRefresh || timeSinceLastRefresh > s.cacheTTL
	s.mu.RUnlock()

	if needsRefresh {
		if err := s.RefreshPortfolio(ctx); err != nil {
			return nil, err
		}
	}

	portfolio, found := s.cache.GetPortfolio()
	if !found {
		return nil, fmt.Errorf("portfolio not found in cache")
	}

	// Filter by holding type if requested
	if holdingType != "" && holdingType != "all" {
		filteredPortfolio := &models.Portfolio{
			TotalValue:        0,
			TotalDayChange:    0,
			TotalDayChangePct: 0,
			TotalPnL:          0,
			LastUpdated:       portfolio.LastUpdated,
		}

		for _, holding := range portfolio.Holdings {
			if holding.Type == holdingType {
				filteredPortfolio.Holdings = append(filteredPortfolio.Holdings, holding)
				filteredPortfolio.TotalValue += holding.CurrentValue
				filteredPortfolio.TotalDayChange += holding.DayChange
				filteredPortfolio.TotalPnL += holding.TotalPnL
			}
		}

		// Calculate day change percentage for the filtered portfolio
		if filteredPortfolio.TotalValue > 0 {
			filteredPortfolio.TotalDayChangePct = (filteredPortfolio.TotalDayChange / filteredPortfolio.TotalValue) * 100
		}

		return filteredPortfolio, nil
	}

	return portfolio, nil
}

// RefreshPortfolio fetches fresh data from all brokers and updates the cache
func (s *Service) RefreshPortfolio(ctx context.Context) error {
	s.mu.Lock()
	defer func() {
		s.lastRefreshTime = time.Now()
		s.mu.Unlock()
	}()

	// Fetch all holdings in parallel
	var wg sync.WaitGroup
	var zerodhaHoldings, zerodhaPositions, iciciHoldings, iciciPositions []models.Holding
	var zerodhaHoldingsErr, zerodhaPositionsErr, iciciHoldingsErr, iciciPositionsErr error
	
	// Zerodha holdings
	wg.Add(1)
	go func() {
		defer wg.Done()
		zerodhaHoldings, zerodhaHoldingsErr = s.fetchZerodhaHoldings(ctx)
	}()

	// Zerodha positions
	wg.Add(1)
	go func() {
		defer wg.Done()
		zerodhaPositions, zerodhaPositionsErr = s.fetchZerodhaPositions(ctx)
	}()

	// ICICI holdings
	wg.Add(1)
	go func() {
		defer wg.Done()
		iciciHoldings, iciciHoldingsErr = s.fetchICICIHoldings(ctx)
	}()

	// ICICI positions
	wg.Add(1)
	go func() {
		defer wg.Done()
		iciciPositions, iciciPositionsErr = s.fetchICICIPositions(ctx)
	}()

	wg.Wait()

	// Check for errors
	if zerodhaHoldingsErr != nil {
		fmt.Printf("Error fetching Zerodha holdings: %v\n", zerodhaHoldingsErr)
	}
	if zerodhaPositionsErr != nil {
		fmt.Printf("Error fetching Zerodha positions: %v\n", zerodhaPositionsErr)
	}
	if iciciHoldingsErr != nil {
		fmt.Printf("Error fetching ICICI holdings: %v\n", iciciHoldingsErr)
	}
	if iciciPositionsErr != nil {
		fmt.Printf("Error fetching ICICI positions: %v\n", iciciPositionsErr)
	}

	// Combine all holdings
	allHoldings := mergeHoldings(
		zerodhaHoldings, 
		zerodhaPositions, 
		iciciHoldings, 
		iciciPositions,
	)

	// Calculate totals
	portfolio := &models.Portfolio{
		Holdings:    allHoldings,
		LastUpdated: time.Now(),
	}

	for _, holding := range allHoldings {
		portfolio.TotalValue += holding.CurrentValue
		portfolio.TotalDayChange += holding.DayChange
		portfolio.TotalPnL += holding.TotalPnL
	}

	if portfolio.TotalValue > 0 {
		portfolio.TotalDayChangePct = (portfolio.TotalDayChange / portfolio.TotalValue) * 100
	}

	// Update cache
	s.cache.SetPortfolio(portfolio)

	return nil
}

// fetchZerodhaHoldings fetches holdings from Zerodha with error handling
func (s *Service) fetchZerodhaHoldings(ctx context.Context) ([]models.Holding, error) {
	if s.zerodhaClient == nil {
		return []models.Holding{}, nil
	}

	holdings, err := s.zerodhaClient.GetHoldings(ctx)
	if err != nil {
		return []models.Holding{}, fmt.Errorf("failed to fetch Zerodha holdings: %w", err)
	}
	return holdings, nil
}

// fetchZerodhaPositions fetches positions from Zerodha with error handling
func (s *Service) fetchZerodhaPositions(ctx context.Context) ([]models.Holding, error) {
	if s.zerodhaClient == nil {
		return []models.Holding{}, nil
	}

	positions, err := s.zerodhaClient.GetPositions(ctx)
	if err != nil {
		return []models.Holding{}, fmt.Errorf("failed to fetch Zerodha positions: %w", err)
	}
	return positions, nil
}

// fetchICICIHoldings fetches holdings from ICICI Direct with error handling
func (s *Service) fetchICICIHoldings(ctx context.Context) ([]models.Holding, error) {
	if s.iciciClient == nil {
		return []models.Holding{}, nil
	}

	holdings, err := s.iciciClient.GetHoldings(ctx)
	if err != nil {
		return []models.Holding{}, fmt.Errorf("failed to fetch ICICI holdings: %w", err)
	}
	return holdings, nil
}

// fetchICICIPositions fetches positions from ICICI Direct with error handling
func (s *Service) fetchICICIPositions(ctx context.Context) ([]models.Holding, error) {
	if s.iciciClient == nil {
		return []models.Holding{}, nil
	}

	positions, err := s.iciciClient.GetPositions(ctx)
	if err != nil {
		return []models.Holding{}, fmt.Errorf("failed to fetch ICICI positions: %w", err)
	}
	return positions, nil
}

// mergeHoldings combines and deduplicates holdings from different sources
func mergeHoldings(holdingsList ...[]models.Holding) []models.Holding {
	holdingsMap := make(map[string]models.Holding)

	for _, holdings := range holdingsList {
		for _, holding := range holdings {
			key := holding.ISIN
			if key == "" {
				// If ISIN is not available, use platform + name as key
				key = holding.Platform + ":" + holding.ItemName
			}

			if existingHolding, exists := holdingsMap[key]; exists {
				// Merge if same security from different platform or update with latest data
				holdingsMap[key] = mergeTwoHoldings(existingHolding, holding)
			} else {
				holdingsMap[key] = holding
			}
		}
	}

	// Convert map back to slice
	result := make([]models.Holding, 0, len(holdingsMap))
	for _, holding := range holdingsMap {
		result = append(result, holding)
	}

	return result
}

// mergeTwoHoldings combines data from two holdings of the same security
func mergeTwoHoldings(a, b models.Holding) models.Holding {
	// Use the more recent data if same security
	if a.LastUpdated.After(b.LastUpdated) {
		return a
	}
	return b
}
