// Package service provides authentication business logic.
package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"

	"sanjow-nova-api/internal/database/db"
	"sanjow-nova-api/internal/shared/apperror"
	"sanjow-nova-api/internal/shared/logging"
)

// UserRepository defines what auth needs from user domain (Go idiom: consumer defines interface).
type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (*db.User, error)
}

// Service provides authentication operations.
type Service struct {
	userRepo  UserRepository
	jwtSecret []byte
}

// New creates a new Service for authentication.
func New(userRepo UserRepository, jwtSecret string) *Service {
	return &Service{
		userRepo:  userRepo,
		jwtSecret: []byte(jwtSecret),
	}
}

// Authenticate verifies credentials and returns a JWT token.
func (s *Service) Authenticate(ctx context.Context, email, password string, expiry time.Duration) (string, error) {
	logger := logging.FromContext(ctx)

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// Log the actual error for debugging, return generic message to client
		logger.Warn("authentication failed: user lookup",
			slog.String("email", email),
			slog.String("error", err.Error()),
		)
		return "", apperror.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		logger.Warn("authentication failed: password mismatch",
			slog.String("email", email),
		)
		return "", apperror.ErrInvalidCredentials
	}

	// Generate JWT with standard claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   user.ID.String(), // subject (user ID) per JWT spec
		"email": user.Email,
		"exp":   time.Now().Add(expiry).Unix(),
		"iat":   time.Now().Unix(),
	})

	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		logger.Error("failed to sign JWT token",
			slog.String("error", err.Error()),
		)
		return "", apperror.Wrap(apperror.CodeTokenGenFailed, "failed to generate token", err)
	}

	logger.Info("user authenticated successfully",
		slog.String("user_id", user.ID.String()),
	)

	return tokenString, nil
}
