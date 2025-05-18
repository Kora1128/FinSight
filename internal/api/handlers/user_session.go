package handlers

import (
	"net/http"
	"time"

	"github.com/Kora1128/FinSight/internal/broker"
	"github.com/Kora1128/FinSight/internal/cache"
	"github.com/Kora1128/FinSight/internal/models"
	"github.com/gin-gonic/gin"
)

// SessionHandler handles user session-related HTTP requests
type SessionHandler struct {
	cache         *cache.Cache
	brokerManager *broker.BrokerManager
	sessionTTL    time.Duration
}

// NewSessionHandler creates a new session handler
func NewSessionHandler(cache *cache.Cache, brokerManager *broker.BrokerManager, sessionTTL time.Duration) *SessionHandler {
	return &SessionHandler{
		cache:         cache,
		brokerManager: brokerManager,
		sessionTTL:    sessionTTL,
	}
}

// CreateSession creates a new user session
func (h *SessionHandler) CreateSession(c *gin.Context) {
	session := models.NewUserSession(h.sessionTTL)
	h.cache.Set("session:"+session.UserID, session, h.sessionTTL)
	
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
	
	sessionKey := "session:" + userID
	sessionObj, found := h.cache.Get(sessionKey)
	if !found {
		c.JSON(http.StatusNotFound, models.SessionResponse{
			Success: false,
			Error:   "Session not found",
		})
		return
	}
	
	session := sessionObj.(*models.UserSession)
	if !session.IsValid() {
		h.cache.Delete(sessionKey)
		c.JSON(http.StatusUnauthorized, models.SessionResponse{
			Success: false,
			Error:   "Session expired",
		})
		return
	}
	
	session.Touch()
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
	sessionKey := "session:" + req.UserID
	sessionObj, found := h.cache.Get(sessionKey)
	if !found {
		c.JSON(http.StatusNotFound, models.SessionResponse{
			Success: false,
			Error:   "Session not found",
		})
		return
	}
	
	session := sessionObj.(*models.UserSession)
	if !session.IsValid() {
		h.cache.Delete(sessionKey)
		c.JSON(http.StatusUnauthorized, models.SessionResponse{
			Success: false,
			Error:   "Session expired",
		})
		return
	}
	
	// Update session
	session.Touch()
	
	// Connect to broker
	clientType := broker.ClientType(req.BrokerType)
	creds := broker.ClientCredentials{
		UserID:       req.UserID,
		APIKey:       req.APIKey,
		APISecret:    req.APISecret,
		RequestToken: req.RequestToken,
		Password:     req.Password,
	}
	
	_, err := h.brokerManager.GetOrCreateClient(clientType, creds)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.SessionResponse{
			Success: false,
			Error:   "Failed to connect to broker: " + err.Error(),
		})
		return
	}
	
	// Update session with connection status
	switch clientType {
	case broker.ClientTypeZerodha:
		session.ZerodhaConnected = true
	case broker.ClientTypeICICIDirect:
		session.ICICIConnected = true
	}
	
	// Update session in cache
	h.cache.Set(sessionKey, session, h.sessionTTL)
	
	c.JSON(http.StatusOK, models.SessionResponse{
		Success: true,
		Data:    session.GetInfo(),
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
	sessionKey := "session:" + userID
	sessionObj, found := h.cache.Get(sessionKey)
	if !found {
		c.JSON(http.StatusNotFound, models.SessionResponse{
			Success: false,
			Error:   "Session not found",
		})
		return
	}
	
	session := sessionObj.(*models.UserSession)
	if !session.IsValid() {
		h.cache.Delete(sessionKey)
		c.JSON(http.StatusUnauthorized, models.SessionResponse{
			Success: false,
			Error:   "Session expired",
		})
		return
	}
	
	// Update session
	session.Touch()
	
	// Disconnect from broker
	clientType := broker.ClientType(brokerType)
	h.brokerManager.RemoveClient(userID, clientType)
	
	// Update session with connection status
	switch clientType {
	case broker.ClientTypeZerodha:
		session.ZerodhaConnected = false
	case broker.ClientTypeICICIDirect:
		session.ICICIConnected = false
	}
	
	// Update session in cache
	h.cache.Set(sessionKey, session, h.sessionTTL)
	
	c.JSON(http.StatusOK, models.SessionResponse{
		Success: true,
		Data:    session.GetInfo(),
	})
}
