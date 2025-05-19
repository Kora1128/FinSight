package middleware

import (
	"net/http"
	"strings"

	"github.com/Kora1128/FinSight/internal/database"
	"github.com/gin-gonic/gin"
)

// SessionAuthConfig holds configuration for session authentication middleware
type SessionAuthConfig struct {
	SessionRepo *database.SessionRepo
	UserRepo    *database.UserRepo
}

// SessionAuth returns middleware to check if a user is authenticated via user session
func SessionAuth(config SessionAuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("userId")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "User ID is required",
			})
			c.Abort()
			return
		}

		// First try to authenticate by session token
		// Get authorization header
		authHeader := c.GetHeader("Authorization")
		var sessionToken string

		// Extract token from Bearer format if needed
		if strings.HasPrefix(authHeader, "Bearer ") {
			sessionToken = strings.TrimPrefix(authHeader, "Bearer ")
		} else if authHeader != "" {
			sessionToken = authHeader
		} else {
			// Try from query string
			sessionToken = c.Query("sessionToken")
		}

		if sessionToken != "" {
			// Authenticate using the session token
			session, err := config.SessionRepo.GetSession(sessionToken)
			if err == nil && session != nil && session.IsValid() {
				// Check if session belongs to the requested user
				if session.UserID == userID {
					// Valid session, update last accessed time
					_ = config.SessionRepo.UpdateLastAccessed(sessionToken)
					// Continue with the request
					c.Next()
					return
				}
			}
		}
		
		// If token authentication failed, try to get a valid session for this user directly
		session, err := config.SessionRepo.GetUserSession(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to validate user",
			})
			c.Abort()
			return
		}

		// Check if a valid session exists for this user
		if session == nil || !session.IsValid() {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid or expired session",
			})
			c.Abort()
			return
		}

		// Update last accessed time
		_ = config.SessionRepo.UpdateLastAccessed(session.SessionID)

		// Session is valid, continue
		c.Next()
	}
}
