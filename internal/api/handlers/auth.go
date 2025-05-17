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
	// Channel to handle token refresh operations
	refreshChan chan struct{}
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(cache *cache.Cache, zerodhaClient *zerodha.Client, iciciClient *icici_direct.Client, sessionDuration time.Duration) *AuthHandler {
	handler := &AuthHandler{
		cache:           cache,
		zerodhaClient:   zerodhaClient,
		iciciClient:     iciciClient,
		sessionDuration: sessionDuration,
		refreshChan:     make(chan struct{}, 1),
	}
	
	// Start the token refresh background process
	go handler.startTokenRefreshWorker()
	
	return handler
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

	// Store access token in cache
	accessToken := h.zerodhaClient.GetAccessToken()
	h.cache.SetZerodhaToken(accessToken, h.sessionDuration)

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

	// Store access token in cache
	accessToken := h.iciciClient.GetAccessToken()
	h.cache.SetICICIToken(accessToken, h.sessionDuration)

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

// startTokenRefreshWorker starts a background goroutine that periodically
// checks token expiration and refreshes tokens that are about to expire
func (h *AuthHandler) startTokenRefreshWorker() {
	ticker := time.NewTicker(15 * time.Minute) // Check every 15 minutes
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			h.refreshTokensIfNeeded()
		case <-h.refreshChan:
			h.refreshTokensIfNeeded()
		}
	}
}

// refreshTokensIfNeeded checks and refreshes tokens if they're close to expiration
func (h *AuthHandler) refreshTokensIfNeeded() {
	// Refresh Zerodha token if needed
	if token, found := h.cache.GetZerodhaToken(); found && token != "" {
		// Zerodha tokens typically expire after 24 hours
		// Try to refresh if we have valid credentials
		if h.zerodhaClient.CanAutoRefresh() {
			if err := h.zerodhaClient.RefreshToken(); err == nil {
				// Get the new access token and store it
				newToken := h.zerodhaClient.GetAccessToken()
				h.cache.SetZerodhaToken(newToken, h.sessionDuration)
			}
		}
	}

	// Refresh ICICI token if needed
	if token, found := h.cache.GetICICIToken(); found && token != "" {
		// ICICI tokens may have different expiry mechanisms
		if h.iciciClient.CanAutoRefresh() {
			if err := h.iciciClient.RefreshToken(); err == nil {
				newToken := h.iciciClient.GetAccessToken()
				h.cache.SetICICIToken(newToken, h.sessionDuration)
			}
		}
	}
}

// LogoutZerodha handles logout requests for Zerodha
func (h *AuthHandler) LogoutZerodha(c *gin.Context) {
	// Remove token from cache
	h.cache.Delete(cache.KeyZerodhaToken)
	
	c.JSON(http.StatusOK, models.LoginResponse{
		Success: true,
		Message: "Successfully logged out from Zerodha",
	})
}

// LogoutICICI handles logout requests for ICICI Direct
func (h *AuthHandler) LogoutICICI(c *gin.Context) {
	// Remove token from cache
	h.cache.Delete(cache.KeyICICIToken)
	
	c.JSON(http.StatusOK, models.LoginResponse{
		Success: true,
		Message: "Successfully logged out from ICICI Direct",
	})
}
