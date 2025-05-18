package handlers

import (
	"context"
	"net/http"

	"github.com/Kora1128/FinSight/internal/models"
	"github.com/Kora1128/FinSight/internal/portfolio"
	"github.com/gin-gonic/gin"
)

// UserPortfolioHandler handles user-specific portfolio-related HTTP requests
type UserPortfolioHandler struct {
	userPortfolioService *portfolio.UserService
}

// NewUserPortfolioHandler creates a new user portfolio handler
func NewUserPortfolioHandler(userPortfolioService *portfolio.UserService) *UserPortfolioHandler {
	return &UserPortfolioHandler{
		userPortfolioService: userPortfolioService,
	}
}

// GetUserPortfolio retrieves the portfolio for a specific user
func (h *UserPortfolioHandler) GetUserPortfolio(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, models.PortfolioResponse{
			Success: false,
			Error:   "User ID is required",
		})
		return
	}

	var req models.PortfolioRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.PortfolioResponse{
			Success: false,
			Error:   "Invalid request parameters",
		})
		return
	}

	// Check if we should force refresh
	forceRefresh := c.Query("refresh") == "true"

	// Get the portfolio
	portfolio, err := h.userPortfolioService.GetPortfolio(context.Background(), userID, forceRefresh, req.Type)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.PortfolioResponse{
			Success: false,
			Error:   "Failed to retrieve portfolio: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.PortfolioResponse{
		Success: true,
		Data:    *portfolio,
	})
}

// RefreshUserPortfolio forces a refresh of portfolio data for a specific user
func (h *UserPortfolioHandler) RefreshUserPortfolio(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, models.PortfolioResponse{
			Success: false,
			Error:   "User ID is required",
		})
		return
	}

	// Refresh the portfolio
	err := h.userPortfolioService.RefreshPortfolio(context.Background(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.PortfolioResponse{
			Success: false,
			Error:   "Failed to refresh portfolio: " + err.Error(),
		})
		return
	}

	// Get the updated portfolio
	portfolio, err := h.userPortfolioService.GetPortfolio(context.Background(), userID, false, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.PortfolioResponse{
			Success: false,
			Error:   "Failed to retrieve refreshed portfolio: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.PortfolioResponse{
		Success: true,
		Data:    *portfolio,
	})
}
