// Package handler provides HTTP handlers for user operations.
package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"sanjow-main-api/internal/domain/user/service"
	"sanjow-main-api/internal/shared/apperror"
	"sanjow-main-api/internal/shared/logging"
	"sanjow-main-api/internal/shared/middleware"
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

// handleError handles errors consistently across all handlers.
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
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// UpdateUserRequest represents the request body for updating a user.
type UpdateUserRequest struct {
	Email     *string `json:"email,omitempty"`
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
}

// Create handles POST /users.
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
		h.handleError(c, err)
		return
	}

	logger.Info("user created", slog.String("user_id", user.ID.String()))
	c.JSON(http.StatusCreated, user)
}

// GetByID handles GET /users/:id.
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
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

// List handles GET /users.
func (h *Handler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	users, err := h.service.List(c.Request.Context(), page, pageSize)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      users,
		"page":      page,
		"page_size": pageSize,
	})
}

// Update handles PUT /users/:id.
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
		h.handleError(c, err)
		return
	}

	logger.Info("user updated", slog.String("user_id", user.ID.String()))
	c.JSON(http.StatusOK, user)
}

// Delete handles DELETE /users/:id.
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
		h.handleError(c, err)
		return
	}

	logger.Info("user deleted", slog.String("user_id", id.String()))
	c.Status(http.StatusNoContent)
}
