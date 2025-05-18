package handlers

import (
	"net/http"
	"time"

	"github.com/Kora1128/FinSight/internal/broker"
	"github.com/Kora1128/FinSight/internal/cache"
	"github.com/Kora1128/FinSight/internal/database"
	"github.com/Kora1128/FinSight/internal/models"
	"github.com/gin-gonic/gin"
)

// SessionHandler handles user session-related HTTP requests
type SessionHandler struct {
	cache         *cache.Cache   // Only used for temporary storage
	sessionRepo   *database.SessionRepo
	userRepo      *database.UserRepo
	brokerManager *broker.BrokerManager
	sessionTTL    time.Duration
}

// NewSessionHandler creates a new session handler
func NewSessionHandler(
	cache *cache.Cache,
	sessionRepo *database.SessionRepo,
	userRepo *database.UserRepo,
	brokerManager *broker.BrokerManager,
	sessionTTL time.Duration,
) *SessionHandler {
	return &SessionHandler{
		cache:         cache,
		sessionRepo:   sessionRepo,
		userRepo:      userRepo,
		brokerManager: brokerManager,
		sessionTTL:    sessionTTL,
	}
}

// CreateSessionRequest represents the request payload for session creation
type CreateSessionRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// CreateSession creates a new user session
func (h *SessionHandler) CreateSession(c *gin.Context) {
	var req CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.SessionResponse{
			Success: false,
			Error:   "Invalid request format: " + err.Error(),
		})
		return
	}
	
	// Find existing session for this email if any
	userID, exists, err := h.userRepo.GetUserByEmail(req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.SessionResponse{
			Success: false,
			Error:   "Failed to check for existing user: " + err.Error(),
		})
		return
	}
	
	// If user exists, check for an active session
	if exists {
		existingSession, err := h.sessionRepo.GetUserSession(userID)
		if err == nil && existingSession != nil && existingSession.IsValid() {
			// Existing valid session found, update last accessed time
			_ = h.sessionRepo.UpdateLastAccessed(existingSession.SessionID)
			_ = h.userRepo.UpdateLastAccessed(userID)
			c.JSON(http.StatusOK, models.SessionResponse{
				Success: true,
				Data:    existingSession.GetInfo(),
			})
			return
		}
	}
	
	// Get or create user ID
	userID, err = h.userRepo.FindOrCreateUserByEmail(req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.SessionResponse{
			Success: false,
			Error:   "Failed to process user account: " + err.Error(),
		})
		return
	}
	
	// Create a new session
	session := models.NewUserSession(req.Email, h.sessionTTL)
	session.UserID = userID  // Use the found or created user ID
	
	// Create session in the database
	if err := h.sessionRepo.CreateSession(session); err != nil {
		c.JSON(http.StatusInternalServerError, models.SessionResponse{
			Success: false,
			Error:   "Failed to create session: " + err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, models.SessionResponse{
		Success: true,
		Data:    session.GetInfo(),
	})
}

// GetSession retrieves the current session
func (h *SessionHandler) GetSession(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, models.SessionResponse{
			Success: false,
			Error:   "User ID is required",
		})
		return
	}
	
	// Get session from database
	session, err := h.sessionRepo.GetUserSession(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.SessionResponse{
			Success: false,
			Error:   "Failed to retrieve session: " + err.Error(),
		})
		return
	}
	
	if session == nil {
		c.JSON(http.StatusNotFound, models.SessionResponse{
			Success: false,
			Error:   "Session not found",
		})
		return
	}
	
	if !session.IsValid() {
		// Delete the expired session
		_ = h.sessionRepo.DeleteSession(session.SessionID)
		
		c.JSON(http.StatusUnauthorized, models.SessionResponse{
			Success: false,
			Error:   "Session expired",
		})
		return
	}
	
	// Update last accessed time
	_ = h.sessionRepo.UpdateLastAccessed(session.SessionID)
	
	c.JSON(http.StatusOK, models.SessionResponse{
		Success: true,
		Data:    session.GetInfo(),
	})
}

// ConnectBroker connects a broker to the user's session
func (h *SessionHandler) ConnectBroker(c *gin.Context) {
	var req models.UserCredentials
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.SessionResponse{
			Success: false,
			Error:   "Invalid request body: " + err.Error(),
		})
		return
	}
	
	// Check if session exists
	session, err := h.sessionRepo.GetUserSession(req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.SessionResponse{
			Success: false,
			Error:   "Failed to retrieve session: " + err.Error(),
		})
		return
	}
	
	if session == nil {
		c.JSON(http.StatusNotFound, models.SessionResponse{
			Success: false,
			Error:   "Session not found",
		})
		return
	}
	
	if !session.IsValid() {
		// Delete the expired session
		_ = h.sessionRepo.DeleteSession(session.SessionID)
		
		c.JSON(http.StatusUnauthorized, models.SessionResponse{
			Success: false,
			Error:   "Session expired",
		})
		return
	}
	
	// Connect to broker and store credentials in database
	clientType := broker.ClientType(req.BrokerType)
	creds := broker.ClientCredentials{
		UserID:       req.UserID,
		APIKey:       req.APIKey,
		APISecret:    req.APISecret,
		RequestToken: req.RequestToken,
		Password:     req.Password,
	}
	
	_, err = h.brokerManager.GetOrCreateClient(clientType, creds)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.SessionResponse{
			Success: false,
			Error:   "Failed to connect to broker: " + err.Error(),
		})
		return
	}
	
	// Update last accessed time
	_ = h.sessionRepo.UpdateLastAccessed(session.SessionID)
	
	// Get the updated session information (with correct connection status)
	updatedSession, _ := h.sessionRepo.GetUserSession(req.UserID)
	if updatedSession == nil {
		updatedSession = session // Fallback to previous session if something went wrong
	}
	
	c.JSON(http.StatusOK, models.SessionResponse{
		Success: true,
		Data:    updatedSession.GetInfo(),
	})
}

// DisconnectBroker disconnects a broker from the user's session
func (h *SessionHandler) DisconnectBroker(c *gin.Context) {
	userID := c.Param("userId")
	brokerType := c.Param("brokerType")
	
	if userID == "" || brokerType == "" {
		c.JSON(http.StatusBadRequest, models.SessionResponse{
			Success: false,
			Error:   "User ID and broker type are required",
		})
		return
	}
	
	// Check if session exists
	session, err := h.sessionRepo.GetUserSession(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.SessionResponse{
			Success: false,
			Error:   "Failed to retrieve session: " + err.Error(),
		})
		return
	}
	
	if session == nil {
		c.JSON(http.StatusNotFound, models.SessionResponse{
			Success: false,
			Error:   "Session not found",
		})
		return
	}
	
	if !session.IsValid() {
		// Delete the expired session
		_ = h.sessionRepo.DeleteSession(session.SessionID)
		
		c.JSON(http.StatusUnauthorized, models.SessionResponse{
			Success: false,
			Error:   "Session expired",
		})
		return
	}
	
	// Disconnect from broker (this will remove credentials from database via repository)
	clientType := broker.ClientType(brokerType)
	h.brokerManager.RemoveClient(userID, clientType)
	
	// Update last accessed time
	_ = h.sessionRepo.UpdateLastAccessed(session.SessionID)
	
	// Get the updated session information (with correct connection status)
	updatedSession, _ := h.sessionRepo.GetUserSession(userID)
	if updatedSession == nil {
		updatedSession = session // Fallback to previous session if something went wrong
	}
	
	c.JSON(http.StatusOK, models.SessionResponse{
		Success: true,
		Data:    updatedSession.GetInfo(),
	})
}
