package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	jwtpkg "github.com/nashirabbash/trackride/pkg/jwt"
)

// Auth returns a middleware that validates JWT access tokens
func Auth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHORIZED"})
			return
		}

		// Extract "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHORIZED"})
			return
		}

		token := parts[1]
		claims, err := jwtpkg.ParseToken(token, secret)
		if err != nil {
			errorCode := "TOKEN_INVALID"

			// Distinguish between expired and other invalid token errors
			if tokenErr, ok := err.(*jwtpkg.TokenError); ok && tokenErr.Code == "expired" {
				errorCode = "TOKEN_EXPIRED"
			}

			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": errorCode})
			return
		}

		c.Set("user_id", claims.UserID)
		c.Next()
	}
}

// GetUserID safely extracts the authenticated user ID from the request context
// Returns empty string if user_id is not found or has invalid type
func GetUserID(c *gin.Context) string {
	userID, exists := c.Get("user_id")
	if !exists {
		return ""
	}

	id, ok := userID.(string)
	if !ok {
		return ""
	}

	return id
}
