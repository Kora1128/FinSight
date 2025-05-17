package handlers

import (
	"context"
	"net/http"

	"github.com/Kora1128/FinSight/internal/models"
	"github.com/Kora1128/FinSight/internal/portfolio"
	"github.com/gin-gonic/gin"
)

// PortfolioHandler handles portfolio-related HTTP requests
type PortfolioHandler struct {
	portfolioService portfolio.ServiceInterface
}

// NewPortfolioHandler creates a new portfolio handler
func NewPortfolioHandler(portfolioService portfolio.ServiceInterface) *PortfolioHandler {
	return &PortfolioHandler{
		portfolioService: portfolioService,
	}
}

// GetPortfolio retrieves the aggregated portfolio
func (h *PortfolioHandler) GetPortfolio(c *gin.Context) {
	// Parse query parameters
	var req models.PortfolioRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.PortfolioResponse{
			Success: false,
			Error:   "Invalid query parameters",
		})
		return
	}

	// Default to "all" if type is not provided
	holdingType := req.Type
	if holdingType == "" {
		holdingType = "all"
	}

	// Get portfolio from service
	ctx := context.Background()
	portfolio, err := h.portfolioService.GetPortfolio(ctx, false, holdingType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.PortfolioResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.PortfolioResponse{
		Success: true,
		Data:    *portfolio,
	})
}

// RefreshPortfolio forces a refresh of portfolio data from brokers
func (h *PortfolioHandler) RefreshPortfolio(c *gin.Context) {
	// Parse query parameters
	var req models.PortfolioRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.PortfolioResponse{
			Success: false,
			Error:   "Invalid query parameters",
		})
		return
	}

	// Default to "all" if type is not provided
	holdingType := req.Type
	if holdingType == "" {
		holdingType = "all"
	}

	// Force refresh
	ctx := context.Background()
	if err := h.portfolioService.RefreshPortfolio(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, models.PortfolioResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Get updated portfolio
	portfolio, err := h.portfolioService.GetPortfolio(ctx, false, holdingType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.PortfolioResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.PortfolioResponse{
		Success: true,
		Data:    *portfolio,
	})
}
