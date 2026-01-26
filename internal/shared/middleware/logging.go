package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"sanjow-nova-api/internal/shared/logging"
)

// RequestID adds a unique request ID to each request.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// Logger creates a request logging middleware.
func Logger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Get or generate request ID
		requestID, _ := c.Get("request_id")
		reqIDStr, _ := requestID.(string)

		// Create request-scoped logger with request attributes
		reqLogger := logging.WithAttrs(logger,
			slog.String("request_id", reqIDStr),
			slog.String("method", c.Request.Method),
			slog.String("path", path),
		)

		// Attach logger to context for handlers to use
		ctx := logging.WithContext(c.Request.Context(), reqLogger)
		c.Request = c.Request.WithContext(ctx)

		// Process request
		c.Next()

		// Log request completion
		latency := time.Since(start)
		status := c.Writer.Status()

		// Build log attributes
		attrs := []any{
			slog.Int("status", status),
			slog.Duration("latency", latency),
			slog.String("client_ip", c.ClientIP()),
		}

		if query != "" {
			attrs = append(attrs, slog.String("query", query))
		}

		if len(c.Errors) > 0 {
			attrs = append(attrs, slog.String("errors", c.Errors.String()))
		}

		// Log at appropriate level based on status
		msg := "request completed"
		switch {
		case status >= 500:
			reqLogger.Error(msg, attrs...)
		case status >= 400:
			reqLogger.Warn(msg, attrs...)
		default:
			reqLogger.Info(msg, attrs...)
		}
	}
}
