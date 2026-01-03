// Package service provides authentication business logic.
package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

// AuthUser contains only the fields needed for authentication
type AuthUser struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
}

// UserService defines what the auth domain needs from the user domain
type UserService interface {
	GetUserForAuth(ctx context.Context, email string) (*AuthUser, error)
	CheckPassword(passwordHash, password string) bool
}

// Service provides authentication operations.
type Service struct {
	userService UserService
	jwtSecret   []byte
}

// New creates a new Service for authentication.
func New(userService UserService, jwtSecret string) *Service {
	return &Service{
		userService: userService,
		jwtSecret:   []byte(jwtSecret),
	}
}

// Authenticate verifies credentials and returns a JWT token
func (s *Service) Authenticate(ctx context.Context, email, password string, expiry time.Duration) (string, error) {
	user, err := s.userService.GetUserForAuth(ctx, email)
	if err != nil {
		return "", errors.New("invalid credentials")
	}
	if !s.userService.CheckPassword(user.PasswordHash, password) {
		return "", errors.New("invalid credentials")
	}

	// Generate JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID.String(),
		"email":   user.Email,
		"exp":     time.Now().Add(expiry).Unix(),
		"iat":     time.Now().Unix(),
	})

	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", errors.New("failed to generate token")
	}

	return tokenString, nil
}
