// Package auth provides authentication endpoints and logic.
package auth

import (
	"sanjow-main-api/internal/domain/auth/handler"
	"sanjow-main-api/internal/domain/auth/service"

	"github.com/gin-gonic/gin"
)

// Domain holds all auth domain components
type Domain struct {
	Handler *handler.Handler
	Service *service.Service
}

// New creates a new auth domain with all its components wired together
func New(userService service.UserService, jwtSecret string) *Domain {
	svc := service.New(userService, jwtSecret)
	h := handler.New(svc)
	return &Domain{
		Handler: h,
		Service: svc,
	}
}

// RegisterRoutes registers all auth routes
func (d *Domain) RegisterRoutes(router *gin.RouterGroup) {
	d.Handler.RegisterRoutes(router)
}
