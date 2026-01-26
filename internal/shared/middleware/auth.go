// Package middleware provides HTTP middleware for the application.
package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"

	"sanjow-main-api/internal/shared/apperror"
	"sanjow-main-api/internal/shared/logging"
)

// Auth returns middleware that verifies a Bearer JWT and sets user context.
// The secret parameter is required and must not be empty.
func Auth(secret string) gin.HandlerFunc {
	if secret == "" {
		panic("middleware: Auth requires a non-empty JWT secret")
	}

	secretBytes := []byte(secret)

	return func(c *gin.Context) {
		logger := logging.FromContext(c.Request.Context())

		// Extract Authorization header
		auth := c.GetHeader("Authorization")
		if auth == "" {
			logger.Warn("auth failed: missing authorization header")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing authorization header",
				"code":  apperror.CodeUnauthorized,
			})
			return
		}

		// Parse Bearer token
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			logger.Warn("auth failed: invalid authorization header format")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid authorization header",
				"code":  apperror.CodeUnauthorized,
			})
			return
		}

		tokenStr := parts[1]

		// Parse and validate token
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrTokenUnverifiable
			}
			return secretBytes, nil
		})
		if err != nil || !token.Valid {
			logger.Warn("auth failed: invalid token", slog.String("error", err.Error()))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token",
				"code":  apperror.CodeUnauthorized,
			})
			return
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			logger.Warn("auth failed: invalid token claims")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token claims",
				"code":  apperror.CodeUnauthorized,
			})
			return
		}

		// Extract user ID from 'sub' claim (JWT standard)
		sub, ok := claims["sub"].(string)
		if !ok || sub == "" {
			logger.Warn("auth failed: missing sub claim in token")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token: missing subject",
				"code":  apperror.CodeUnauthorized,
			})
			return
		}

		// Parse UUID
		userID, err := uuid.Parse(sub)
		if err != nil {
			logger.Warn("auth failed: invalid user id format", slog.String("sub", sub))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token: invalid subject format",
				"code":  apperror.CodeUnauthorized,
			})
			return
		}

		// Set user ID in context for downstream handlers
		SetUserID(c, userID)

		logger.Debug("auth successful", slog.String("user_id", userID.String()))
		c.Next()
	}
}
