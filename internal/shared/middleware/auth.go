// Package middleware provides HTTP middleware for the application.
package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"sanjow-nova-api/internal/shared/apperror"
	"sanjow-nova-api/internal/shared/logging"
)

// BasicAuth returns middleware that requires HTTP Basic Authentication.
// Used to protect endpoints like /docs with a simple username/password.
func BasicAuth(username, password string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, pass, hasAuth := c.Request.BasicAuth()
		if !hasAuth || user != username || pass != password {
			c.Header("WWW-Authenticate", `Basic realm="API Documentation"`)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Next()
	}
}

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

		// Parse and validate token with typed claims
		var claims jwt.RegisteredClaims
		token, err := jwt.ParseWithClaims(tokenStr, &claims, func(t *jwt.Token) (any, error) {
			return secretBytes, nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
		if err != nil {
			logger.Warn("auth failed: invalid token", slog.String("error", err.Error()))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token",
				"code":  apperror.CodeUnauthorized,
			})
			return
		}
		_ = token // claims are already populated

		// Extract user ID from 'sub' claim (JWT standard)
		if claims.Subject == "" {
			logger.Warn("auth failed: missing sub claim in token")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token: missing subject",
				"code":  apperror.CodeUnauthorized,
			})
			return
		}

		// Parse UUID
		userID, err := uuid.Parse(claims.Subject)
		if err != nil {
			logger.Warn("auth failed: invalid user id format", slog.String("sub", claims.Subject))
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
