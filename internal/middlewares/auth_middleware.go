package middleware

import (
	"chatapp/internal/services"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func JWTAuth(jwtSvc *services.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := ""

		// 1. Try Header
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
				tokenString = parts[1]
			}
		}

		// 2. Try Query Param (for WebSockets)
		if tokenString == "" {
			tokenString = c.Query("token")
		}

		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header or token query param required"})
			return
		}

		claims, err := jwtSvc.ParseToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.Set("claims", claims)
		c.Next()
	}
}
