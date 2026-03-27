// Package handler provides HTTP handlers for user operations.
package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"sanjow-nova-api/internal/domain/user/service"
	"sanjow-nova-api/internal/shared/apperror"
	"sanjow-nova-api/internal/shared/httputil"
	"sanjow-nova-api/internal/shared/logging"
	"sanjow-nova-api/internal/shared/middleware"
)

// Handler handles HTTP requests for users.
type Handler struct {
	service   *service.Service
	jwtSecret string
}

// New creates a new user handler.
func New(svc *service.Service, jwtSecret string) *Handler {
	return &Handler{
		service:   svc,
		jwtSecret: jwtSecret,
	}
}

// RegisterRoutes registers user routes.
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	authMiddleware := middleware.Auth(h.jwtSecret)

	users := router.Group("/users")
	{
		users.POST("", h.Create)
		users.GET("", h.List)
		users.GET(":id", h.GetByID)
		users.PUT(":id", authMiddleware, h.requireUser(), h.Update)
		users.DELETE(":id", authMiddleware, h.requireUser(), h.Delete)
	}
}

// requireUser is a middleware that validates the authenticated user exists.
func (h *Handler) requireUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := logging.FromContext(c.Request.Context())

		userID, ok := middleware.GetUserID(c)
		if !ok {
			logger.Warn("missing user id in context")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized",
				"code":  apperror.CodeUnauthorized,
			})
			return
		}

		// Verify user exists in database
		if _, err := h.service.GetByID(c.Request.Context(), userID); err != nil {
			if apperror.Is(err, apperror.CodeNotFound) {
				logger.Warn("authenticated user not found", slog.String("user_id", userID.String()))
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": "user not found",
					"code":  apperror.CodeUnauthorized,
				})
				return
			}
			logger.Error("failed to verify user", slog.String("error", err.Error()))
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "internal error",
				"code":  apperror.CodeInternal,
			})
			return
		}

		c.Next()
	}
}

// CreateUserRequest represents the request body for creating a user.
type CreateUserRequest struct {
	Email     string `json:"email" binding:"required,email" example:"user@example.com"`
	Password  string `json:"password" binding:"required,min=8" example:"password123"`
	FirstName string `json:"first_name" example:"John"`
	LastName  string `json:"last_name" example:"Doe"`
}

// UpdateUserRequest represents the request body for updating a user.
type UpdateUserRequest struct {
	Email     *string `json:"email,omitempty" example:"newemail@example.com"`
	FirstName *string `json:"first_name,omitempty" example:"Jane"`
	LastName  *string `json:"last_name,omitempty" example:"Smith"`
}

// Create godoc
// @Summary Create a new user
// @Description Create a new user account with email, password, and optional name
// @Tags users
// @Accept json
// @Produce json
// @Param request body CreateUserRequest true "User registration data"
// @Success 201 {object} service.UserResponse "Created user"
// @Failure 400 {object} apperror.ErrorResponse "Invalid input"
// @Failure 409 {object} apperror.ErrorResponse "Email already exists"
// @Router /users [post]
func (h *Handler) Create(c *gin.Context) {
	logger := logging.FromContext(c.Request.Context())

	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("invalid request body", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"code":  apperror.CodeInvalidInput,
		})
		return
	}

	user, err := h.service.Create(c.Request.Context(), service.CreateUserInput{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})
	if err != nil {
		httputil.HandleError(c, err)
		return
	}

	logger.Info("user created", slog.String("user_id", user.ID.String()))
	c.JSON(http.StatusCreated, user)
}

// GetByID godoc
// @Summary Get user by ID
// @Description Retrieve a user by their UUID
// @Tags users
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Success 200 {object} service.UserResponse "User data"
// @Failure 400 {object} apperror.ErrorResponse "Invalid user ID"
// @Failure 404 {object} apperror.ErrorResponse "User not found"
// @Router /users/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	logger := logging.FromContext(c.Request.Context())

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		logger.Warn("invalid user id", slog.String("id", idStr))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user id",
			"code":  apperror.CodeInvalidInput,
		})
		return
	}

	user, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		httputil.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

// List godoc
// @Summary List all users
// @Description Retrieve a paginated list of all users
// @Tags users
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Items per page" default(20)
// @Success 200 {object} map[string]any "Paginated user list"
// @Router /users [get]
func (h *Handler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	users, err := h.service.List(c.Request.Context(), page, pageSize)
	if err != nil {
		httputil.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      users,
		"page":      page,
		"page_size": pageSize,
	})
}

// Update godoc
// @Summary Update user
// @Description Update user profile information (requires authentication)
// @Tags users
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "User ID (UUID)"
// @Param request body UpdateUserRequest true "Updated user data"
// @Success 200 {object} service.UserResponse "Updated user"
// @Failure 400 {object} apperror.ErrorResponse "Invalid input"
// @Failure 401 {object} apperror.ErrorResponse "Unauthorized"
// @Failure 404 {object} apperror.ErrorResponse "User not found"
// @Router /users/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	logger := logging.FromContext(c.Request.Context())

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		logger.Warn("invalid user id", slog.String("id", idStr))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user id",
			"code":  apperror.CodeInvalidInput,
		})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("invalid request body", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"code":  apperror.CodeInvalidInput,
		})
		return
	}

	user, err := h.service.Update(c.Request.Context(), id, service.UpdateUserInput{
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})
	if err != nil {
		httputil.HandleError(c, err)
		return
	}

	logger.Info("user updated", slog.String("user_id", user.ID.String()))
	c.JSON(http.StatusOK, user)
}

// Delete godoc
// @Summary Delete user
// @Description Delete a user account (requires authentication)
// @Tags users
// @Security Bearer
// @Param id path string true "User ID (UUID)"
// @Success 204 "User deleted"
// @Failure 401 {object} apperror.ErrorResponse "Unauthorized"
// @Failure 404 {object} apperror.ErrorResponse "User not found"
// @Router /users/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	logger := logging.FromContext(c.Request.Context())

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		logger.Warn("invalid user id", slog.String("id", idStr))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user id",
			"code":  apperror.CodeInvalidInput,
		})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		httputil.HandleError(c, err)
		return
	}

	logger.Info("user deleted", slog.String("user_id", id.String()))
	c.Status(http.StatusNoContent)
}
