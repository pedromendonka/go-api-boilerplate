// Package middleware | Gin middleware for JWT authentication, setting user context.
package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"

	"sanjow-main-api/internal/shared/ctx"
)

// Auth middleware verifies a Bearer JWT and sets the ctx.User and typed ctx.UserID values in context
func Auth() gin.HandlerFunc {
	secret := os.Getenv("JWT_SECRET")
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
			return
		}

		tokenStr := parts[1]

		if secret == "" {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "jwt secret not configured"})
			return
		}

		// Parse and validate token
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			// Verify signing method
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrTokenUnverifiable
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		// Extract claims as map
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			return
		}

		// Try to extract user id from 'sub' or 'user_id' claim
		var idStr string
		if sub, ok := claims["sub"].(string); ok && sub != "" {
			idStr = sub
		} else if uid, ok := claims["user_id"].(string); ok && uid != "" {
			idStr = uid
		}

		if idStr == "" {
			// no typed user id available; proceed but without typed user_id
			c.Set(string(ctx.User), claims)
			c.Next()
			return
		}

		// Parse UUID
		uid, err := uuid.Parse(idStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid user id in token"})
			return
		}

		// Set claims and typed user id in context for handlers
		c.Set(string(ctx.User), claims)
		c.Set(string(ctx.UserID), uid)
		c.Next()
	}
}
