// Package httputil provides shared HTTP utilities for Gin handlers.
package httputil

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"sanjow-nova-api/internal/shared/apperror"
	"sanjow-nova-api/internal/shared/logging"
)

// HandleError handles errors consistently across all handlers.
func HandleError(c *gin.Context, err error) {
	logger := logging.FromContext(c.Request.Context())

	if appErr, ok := apperror.AsAppError(err); ok {
		level := slog.LevelWarn
		if appErr.HTTPStatus() >= 500 {
			level = slog.LevelError
		}
		attrs := []slog.Attr{slog.String("code", string(appErr.Code))}
		if appErr.Err != nil {
			attrs = append(attrs, slog.String("error", appErr.Err.Error()))
		}
		logger.LogAttrs(c.Request.Context(), level, "request failed", attrs...)
		c.JSON(appErr.HTTPStatus(), gin.H{
			"error": appErr.Message,
			"code":  appErr.Code,
		})
		return
	}

	logger.Error("unexpected error", slog.String("error", err.Error()))
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": "internal server error",
		"code":  apperror.CodeInternal,
	})
}
