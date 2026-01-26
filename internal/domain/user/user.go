// Package user defines the user domain model and related logic.
package user

import (
	"sanjow-nova-api/internal/domain/user/handler"
	"sanjow-nova-api/internal/domain/user/repository"
	"sanjow-nova-api/internal/domain/user/service"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Domain holds all user domain components
type Domain struct {
	Handler    *handler.Handler
	Service    *service.Service
	Repository *repository.Repository
}

// New creates a new user domain with all its components wired together.
// jwtSecret is required for protected route authentication.
func New(pool *pgxpool.Pool, jwtSecret string) *Domain {
	repo := repository.New(pool)
	svc := service.New(repo)
	h := handler.New(svc, jwtSecret)

	return &Domain{
		Handler:    h,
		Service:    svc,
		Repository: repo,
	}
}

// RegisterRoutes registers all user routes
func (d *Domain) RegisterRoutes(router *gin.RouterGroup) {
	d.Handler.RegisterRoutes(router)
}
