// Package user defines the user domain model and related logic.
package user

import (
	"sanjow-main-api/internal/domain/user/handler"
	"sanjow-main-api/internal/domain/user/repository"
	"sanjow-main-api/internal/domain/user/service"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Domain holds all user domain components
type Domain struct {
	Handler    *handler.Handler
	Service    *service.Service
	Repository *repository.Repository
}

// New creates a new user domain with all its components wired together
func New(pool *pgxpool.Pool) *Domain {
	repo := repository.New(pool)
	svc := service.New(repo)
	h := handler.New(svc, repo)

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
