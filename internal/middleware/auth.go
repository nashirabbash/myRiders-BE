package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	jwtpkg "github.com/nashirabbash/trackride/pkg/jwt"
)

// Auth creates a Gin middleware that validates JWT access tokens
func Auth(accessSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHORIZED"})
			return
		}

		// Expect "Bearer <token>" format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHORIZED"})
			return
		}

		tokenString := parts[1]

		// Parse and validate token
		claims, err := jwtpkg.ParseToken(tokenString, accessSecret)
		if err != nil {
			if tokenErr, ok := err.(jwtpkg.TokenError); ok {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": tokenErr.Code})
				return
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHORIZED"})
			return
		}

		// Verify token type
		if claims.Type != "access" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHORIZED"})
			return
		}

		// Store user_id in context for downstream handlers
		c.Set("user_id", claims.UserID())
		c.Next()
	}
}

// GetUserID safely retrieves the user_id from the Gin context
// Returns the user_id and a boolean indicating success
func GetUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", false
	}

	id, ok := userID.(string)
	if !ok {
		return "", false
	}

	// Empty user_id indicates malformed token or missing subject claim
	if id == "" || strings.TrimSpace(id) == "" {
		return "", false
	}

	return id, true
}
