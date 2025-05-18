package handlers

import (
	"net/http"
	"time"

	"github.com/Kora1128/FinSight/internal/models"
	"github.com/gin-gonic/gin"
)

// GetZerodhaAuthURL generates and returns the Zerodha authorization URL
func (h *AuthHandler) GetZerodhaAuthURL(c *gin.Context) {
	// Extract query parameters
	redirectURI := c.Query("redirectUri")
	apiKey := c.Query("apiKey")

	if apiKey == "" {
		apiKey = h.zerodhaClient.GetAPIKey() // Fallback to default API key
	}

	if apiKey == "" {
		c.JSON(http.StatusBadRequest, models.AuthURLResponse{
			Success: false,
			Error:   "API key is required",
		})
		return
	}

	// Generate the Zerodha authorization URL
	authURL := h.zerodhaClient.GetLoginURL(redirectURI)

	c.JSON(http.StatusOK, models.AuthURLResponse{
		Success: true,
		URL:     authURL,
	})
}

// GetICICIAuthURL generates and returns the ICICI Direct authorization URL
func (h *AuthHandler) GetICICIAuthURL(c *gin.Context) {
	// Extract query parameters
	redirectURI := c.Query("redirectUri")
	apiKey := c.Query("apiKey")

	if apiKey == "" {
		apiKey = h.iciciClient.GetAPIKey() // Fallback to default API key
	}

	if apiKey == "" {
		c.JSON(http.StatusBadRequest, models.AuthURLResponse{
			Success: false,
			Error:   "API key is required",
		})
		return
	}

	// Generate the ICICI Direct authorization URL
	authURL := h.iciciClient.GetLoginURL(redirectURI)

	c.JSON(http.StatusOK, models.AuthURLResponse{
		Success: true,
		URL:     authURL,
	})
}

// RefreshToken handles token refresh requests for both brokers
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req models.TokenRefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.TokenResponse{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	var token string
	var refreshToken string
	var expiresAt time.Time
	var err error

	switch req.BrokerType {
	case "zerodha":
		// Setting the refresh token for Zerodha client if not already set
		if h.zerodhaClient.GetRefreshToken() == "" {
			h.zerodhaClient.SetRefreshToken(req.RefreshToken)
		}

		// Refresh the token
		if err = h.zerodhaClient.RefreshToken(); err == nil {
			token = h.zerodhaClient.GetAccessToken()
			refreshToken = h.zerodhaClient.GetRefreshToken()
			expiresAt = time.Now().Add(24 * time.Hour)
			
			// Update the cache
			h.cache.SetZerodhaToken(token, h.sessionDuration)
		}

	case "icici":
		// Setting the refresh token for ICICI client if not already set
		if h.iciciClient.GetRefreshToken() == "" {
			h.iciciClient.SetRefreshToken(req.RefreshToken)
		}

		// Refresh the token
		if err = h.iciciClient.RefreshToken(); err == nil {
			token = h.iciciClient.GetAccessToken()
			refreshToken = h.iciciClient.GetRefreshToken()
			expiresAt = time.Now().Add(12 * time.Hour)
			
			// Update the cache
			h.cache.SetICICIToken(token, h.sessionDuration)
		}

	default:
		c.JSON(http.StatusBadRequest, models.TokenResponse{
			Success: false,
			Error:   "Invalid broker type",
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusUnauthorized, models.TokenResponse{
			Success: false,
			Error:   "Token refresh failed: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.TokenResponse{
		Success:      true,
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	})
}
