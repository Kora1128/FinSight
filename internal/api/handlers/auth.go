package handlers

import (
	"net/http"
	"time"

	"github.com/Kora1128/FinSight/internal/broker/icici_direct"
	"github.com/Kora1128/FinSight/internal/broker/zerodha"
	"github.com/Kora1128/FinSight/internal/cache"
	"github.com/Kora1128/FinSight/internal/models"
	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	cache           *cache.Cache
	zerodhaClient   *zerodha.Client
	iciciClient     *icici_direct.Client
	sessionDuration time.Duration
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(cache *cache.Cache, zerodhaClient *zerodha.Client, iciciClient *icici_direct.Client, sessionDuration time.Duration) *AuthHandler {
	return &AuthHandler{
		cache:           cache,
		zerodhaClient:   zerodhaClient,
		iciciClient:     iciciClient,
		sessionDuration: sessionDuration,
	}
}

// ZerodhaLogin handles login requests for Zerodha
func (h *AuthHandler) ZerodhaLogin(c *gin.Context) {
	var req models.ZerodhaLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.LoginResponse{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	// Login to Zerodha
	if err := h.zerodhaClient.Login(req.RequestToken, req.APISecret); err != nil {
		c.JSON(http.StatusUnauthorized, models.LoginResponse{
			Success: false,
			Error:   "Failed to login to Zerodha: " + err.Error(),
		})
		return
	}

	// Store token in cache (in a real app, you'd store the access token)
	h.cache.SetZerodhaToken("logged_in", h.sessionDuration)

	c.JSON(http.StatusOK, models.LoginResponse{
		Success: true,
		Message: "Successfully logged in to Zerodha",
	})
}

// ICICILogin handles login requests for ICICI Direct
func (h *AuthHandler) ICICILogin(c *gin.Context) {
	var req models.ICICILoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.LoginResponse{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	// Login to ICICI Direct
	if err := h.iciciClient.Login(req.APIKey, req.APISecret); err != nil {
		c.JSON(http.StatusUnauthorized, models.LoginResponse{
			Success: false,
			Error:   "Failed to login to ICICI Direct: " + err.Error(),
		})
		return
	}

	// Store token in cache (in a real app, you'd store the actual token)
	h.cache.SetICICIToken("logged_in", h.sessionDuration)

	c.JSON(http.StatusOK, models.LoginResponse{
		Success: true,
		Message: "Successfully logged in to ICICI Direct",
	})
}

// GetUserStatus returns the login status for both brokers
func (h *AuthHandler) GetUserStatus(c *gin.Context) {
	// Check Zerodha login status
	zerodhaToken, zerodhaLoggedIn := h.cache.GetZerodhaToken()
	
	// Check ICICI Direct login status
	iciciToken, iciciLoggedIn := h.cache.GetICICIToken()

	status := models.UserStatus{
		ZerodhaLoggedIn: zerodhaLoggedIn && zerodhaToken != "",
		ICICILoggedIn:   iciciLoggedIn && iciciToken != "",
	}

	c.JSON(http.StatusOK, models.UserStatusResponse{
		Success: true,
		Data:    status,
	})
}
