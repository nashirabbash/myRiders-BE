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
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "TOKEN_EXPIRED"})
			return
		}

		c.Set("user_id", claims.UserID)
		c.Next()
	}
}

// GetUserID extracts the authenticated user ID from the request context
func GetUserID(c *gin.Context) string {
	userID, _ := c.Get("user_id")
	return userID.(string)
}
