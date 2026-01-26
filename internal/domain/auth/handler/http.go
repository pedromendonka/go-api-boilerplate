// Package handler provides HTTP handlers for authentication endpoints.
package handler

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"sanjow-nova-api/internal/domain/auth/service"
	"sanjow-nova-api/internal/shared/apperror"
	"sanjow-nova-api/internal/shared/logging"
)

// Handler handles authentication HTTP requests.
type Handler struct {
	service *service.Service
}

// New creates a new Handler for authentication.
func New(svc *service.Service) *Handler {
	return &Handler{service: svc}
}

// LoginRequest represents a login request payload.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"password123"`
}

// LoginResponse represents a login response payload.
type LoginResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// RegisterRoutes registers all auth routes.
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/login", h.Login)
}

// handleError handles errors consistently.
func (h *Handler) handleError(c *gin.Context, err error) {
	logger := logging.FromContext(c.Request.Context())

	if appErr, ok := apperror.AsAppError(err); ok {
		if appErr.Err != nil {
			logger.Error("request failed",
				slog.String("code", string(appErr.Code)),
				slog.String("error", appErr.Err.Error()),
			)
		} else {
			logger.Warn("request failed",
				slog.String("code", string(appErr.Code)),
			)
		}
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

// Login godoc
// @Summary User login
// @Description Authenticate with email and password to receive a JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} LoginResponse "JWT token"
// @Failure 400 {object} apperror.ErrorResponse "Invalid request"
// @Failure 401 {object} apperror.ErrorResponse "Invalid credentials"
// @Router /login [post]
func (h *Handler) Login(c *gin.Context) {
	logger := logging.FromContext(c.Request.Context())

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("invalid login request", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request",
			"code":  apperror.CodeInvalidInput,
		})
		return
	}

	token, err := h.service.Authenticate(c.Request.Context(), req.Email, req.Password, 24*time.Hour)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, LoginResponse{Token: token})
}
