// Package handler | HTTP handler for authentication-related endpoints.
package handler

import (
	"net/http"
	"time"

	"sanjow-main-api/internal/domain/auth/service"

	"github.com/gin-gonic/gin"
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
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents a login response payload.
type LoginResponse struct {
	Token string `json:"token"`
}

// RegisterRoutes registers all auth routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/login", h.Login)
}

// Login handles user authentication
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Default expiry of 24 hours, can be parameterized
	token, err := h.service.Authenticate(c.Request.Context(), req.Email, req.Password, 24*time.Hour)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{Token: token})
}
