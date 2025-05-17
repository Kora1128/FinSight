package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Kora1128/FinSight/internal/models"
	"github.com/Kora1128/FinSight/internal/portfolio"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// MockPortfolioService implements the portfolio.ServiceInterface for testing
type MockPortfolioService struct {
	portfolio    *models.Portfolio
	refreshError error
	getError     error
}

// Ensure MockPortfolioService implements portfolio.ServiceInterface
var _ portfolio.ServiceInterface = (*MockPortfolioService)(nil)

// GetPortfolio mocks the portfolio service's GetPortfolio method
func (m *MockPortfolioService) GetPortfolio(ctx context.Context, forceRefresh bool, holdingType models.HoldingType) (*models.Portfolio, error) {
	if m.getError != nil {
		return nil, m.getError
	}

	if m.portfolio == nil {
		return &models.Portfolio{
			Holdings:    []models.Holding{},
			TotalValue:  0,
			LastUpdated: time.Now(),
		}, nil
	}

	// Filter by type if requested
	if holdingType != "" && holdingType != "all" {
		filteredPortfolio := &models.Portfolio{
			Holdings:    []models.Holding{},
			TotalValue:  0,
			LastUpdated: m.portfolio.LastUpdated,
		}

		for _, holding := range m.portfolio.Holdings {
			if holding.Type == holdingType {
				filteredPortfolio.Holdings = append(filteredPortfolio.Holdings, holding)
				filteredPortfolio.TotalValue += holding.CurrentValue
			}
		}

		return filteredPortfolio, nil
	}

	return m.portfolio, nil
}

// RefreshPortfolio mocks the portfolio service's RefreshPortfolio method
func (m *MockPortfolioService) RefreshPortfolio(ctx context.Context) error {
	return m.refreshError
}

// setupTest sets up the test environment
func setupTest() (*gin.Engine, *MockPortfolioService) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	mockService := &MockPortfolioService{}
	handler := NewPortfolioHandler(mockService)

	r.GET("/api/v1/portfolio", handler.GetPortfolio)
	r.POST("/api/v1/portfolio/refresh", handler.RefreshPortfolio)

	return r, mockService
}

// TestGetPortfolio tests the GetPortfolio handler
func TestGetPortfolio(t *testing.T) {
	r, mockService := setupTest()

	// Set up mock data
	mockService.portfolio = &models.Portfolio{
		Holdings: []models.Holding{
			{
				ItemName:        "INFY",
				ISIN:            "INE009A01021",
				Quantity:        10,
				AveragePrice:    1000,
				LastTradedPrice: 1100,
				CurrentValue:    11000,
				Platform:        models.PlatformZerodha,
				Type:            models.HoldingTypeStock,
			},
			{
				ItemName:        "HDFCBANK",
				ISIN:            "INE040A01034",
				Quantity:        5,
				AveragePrice:    1500,
				LastTradedPrice: 1600,
				CurrentValue:    8000,
				Platform:        models.PlatformICICIDirect,
				Type:            models.HoldingTypeMutualFund,
			},
		},
		TotalValue:  19000,
		LastUpdated: time.Now(),
	}

	// Test successful response
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/portfolio", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.PortfolioResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, 2, len(response.Data.Holdings))
	assert.Equal(t, 19000.0, response.Data.TotalValue)

	// Test with type filter (stocks only)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/portfolio?type=stock", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, 1, len(response.Data.Holdings))
	assert.Equal(t, "INFY", response.Data.Holdings[0].ItemName)

	// Test error case
	mockService.getError = errors.New("failed to get portfolio")

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/portfolio", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "failed to get portfolio", response.Error)
}

// TestRefreshPortfolio tests the RefreshPortfolio handler
func TestRefreshPortfolio(t *testing.T) {
	r, mockService := setupTest()

	// Set up mock data
	mockService.portfolio = &models.Portfolio{
		Holdings: []models.Holding{
			{
				ItemName:        "INFY",
				ISIN:            "INE009A01021",
				Quantity:        10,
				AveragePrice:    1000,
				LastTradedPrice: 1100,
				CurrentValue:    11000,
				Platform:        models.PlatformZerodha,
				Type:            models.HoldingTypeStock,
			},
		},
		TotalValue:  11000,
		LastUpdated: time.Now(),
	}

	// Test successful refresh
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/portfolio/refresh", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.PortfolioResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, 1, len(response.Data.Holdings))

	// Test refresh error
	mockService.refreshError = errors.New("refresh failed")

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/portfolio/refresh", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "refresh failed", response.Error)
}
