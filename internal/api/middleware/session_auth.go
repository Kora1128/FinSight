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

		if sessionToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Session token is required",
			})
			c.Abort()
			return
		}

		// Get session from database
		session, err := config.SessionRepo.GetSession(sessionToken)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to validate session",
			})
			c.Abort()
			return
		}

		// Check if session exists and belongs to the requested user
		if session == nil || session.UserID != userID {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid or expired session token",
			})
			c.Abort()
			return
		}

		// Update last accessed time
		_ = config.SessionRepo.UpdateLastAccessed(sessionToken)

		// Session is valid, continue
		c.Next()
	}
}
