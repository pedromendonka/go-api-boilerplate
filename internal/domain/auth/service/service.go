// Package service provides authentication business logic.
package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"

	"sanjow-main-api/internal/database/db"
)

// UserRepository defines what auth needs from user domain (Go idiom: consumer defines interface)
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

// Authenticate verifies credentials and returns a JWT token
func (s *Service) Authenticate(ctx context.Context, email, password string, expiry time.Duration) (string, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
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
