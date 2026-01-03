// Package handler provides HTTP handlers for user operations.
package handler

import (
	"errors"
	"net/http"
	"strconv"

	"sanjow-main-api/internal/domain/user/repository"
	"sanjow-main-api/internal/domain/user/service"
	"sanjow-main-api/internal/shared/ctx"
	"sanjow-main-api/internal/shared/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Handler handles HTTP requests for users
type Handler struct {
	service *service.Service
	repo    *repository.Repository
}

// New creates a new user handler
func New(svc *service.Service, repo *repository.Repository) *Handler {
	return &Handler{service: svc, repo: repo}
}

// RegisterRoutes registers user routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	users := router.Group("/users")
	{
		users.POST("", h.Create)
		users.GET("", h.List)
		users.GET(":id", h.GetByID)
		// Protect update and delete with Auth + current user injection
		users.PUT(":id", middleware.Auth(), h.injectCurrentUser(), h.Update)
		users.DELETE(":id", middleware.Auth(), h.injectCurrentUser(), h.Delete)
	}
}

// injectCurrentUser is a middleware that looks up the user by user_id from context and stores
// the domain response under the "current_user" key for handlers to use.
func (h *Handler) injectCurrentUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		v, ok := c.Get(string(ctx.UserID))
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing user id in context"})
			return
		}

		uid, ok := v.(uuid.UUID)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid user id in context"})
			return
		}

		userObj, err := h.repo.GetByID(c.Request.Context(), uid)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
				return
			}
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to load user"})
			return
		}

		// Convert to response and set in context
		c.Set(string(ctx.CurrentUser), service.ToResponse(userObj))
		c.Next()
	}
}

// CreateUserRequest represents the request body for creating a user
type CreateUserRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// UpdateUserRequest represents the request body for updating a user
type UpdateUserRequest struct {
	Email     *string `json:"email,omitempty"`
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
}

// Create handles POST /users
func (h *Handler) Create(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.service.Create(c.Request.Context(), service.CreateUserInput{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})
	if err != nil {
		switch err {
		case service.ErrUserAlreadyExists:
			c.JSON(http.StatusConflict, gin.H{"error": "user already exists"})
		case service.ErrInvalidInput:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusCreated, user)
}

// GetByID handles GET /users/:id
func (h *Handler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	user, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == service.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// List handles GET /users
func (h *Handler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	users, err := h.service.List(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      users,
		"page":      page,
		"page_size": pageSize,
	})
}

// Update handles PUT /users/:id
func (h *Handler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.service.Update(c.Request.Context(), id, service.UpdateUserInput{
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})
	if err != nil {
		if err == service.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// Delete handles DELETE /users/:id
func (h *Handler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
