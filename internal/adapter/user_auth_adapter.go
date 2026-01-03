// Package adapter provides adapters for various services
package adapter

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	authservice "sanjow-main-api/internal/domain/auth/service"
	"sanjow-main-api/internal/domain/user/repository"
)

// ErrUserNotFound is returned when a user is not found
var ErrUserNotFound = errors.New("user not found")

// UserAuthAdapter adapts user repository to auth's UserService interface
type UserAuthAdapter struct {
	repo *repository.Repository
}

// NewUserAuthAdapter creates a new UserAuthAdapter
func NewUserAuthAdapter(repo *repository.Repository) *UserAuthAdapter {
	return &UserAuthAdapter{repo: repo}
}

// GetUserForAuth retrieves a user by email for authentication purposes
func (a *UserAuthAdapter) GetUserForAuth(ctx context.Context, email string) (*authservice.AuthUser, error) {
	user, err := a.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &authservice.AuthUser{
		ID:           user.ID,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
	}, nil
}

// CheckPassword verifies a password against a hash
func (a *UserAuthAdapter) CheckPassword(passwordHash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)) == nil
}
