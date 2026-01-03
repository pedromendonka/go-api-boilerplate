// Package service handles user business logic for users.
package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"sanjow-main-api/internal/database/db"
	"sanjow-main-api/internal/domain/user/repository"
)

var (
	// ErrUserNotFound is returned when a user is not found in the repository.
	ErrUserNotFound = errors.New("user not found")
	// ErrUserAlreadyExists is returned when a user with the same email already exists.
	ErrUserAlreadyExists = errors.New("user already exists")
	// ErrInvalidInput is returned when the input provided to a function is invalid.
	ErrInvalidInput = errors.New("invalid input")
)

// Service handles user business logic
type Service struct {
	repo *repository.Repository
}

// New creates a new user service
func New(repo *repository.Repository) *Service {
	return &Service{repo: repo}
}

// CreateUserInput represents input for creating a user
type CreateUserInput struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
}

// UpdateUserInput represents input for updating a user
type UpdateUserInput struct {
	Email     *string
	FirstName *string
	LastName  *string
}

// UserResponse represents a user response (without sensitive data)
type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	FirstName *string   `json:"first_name,omitempty"`
	LastName  *string   `json:"last_name,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToResponse converts a db.User to UserResponse
func ToResponse(user *db.User) *UserResponse {
	return &UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

// Create creates a new user
func (s *Service) Create(ctx context.Context, input CreateUserInput) (*UserResponse, error) {
	// Validate input
	if input.Email == "" || input.Password == "" {
		return nil, ErrInvalidInput
	}

	// Check if user already exists
	_, err := s.repo.GetByEmail(ctx, input.Email)
	if err == nil {
		return nil, ErrUserAlreadyExists
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	// Hash password using bcrypt
	passwordHash := hashPassword(input.Password)

	// Create user
	user, err := s.repo.Create(ctx, input.Email, passwordHash, input.FirstName, input.LastName)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return ToResponse(user), nil
}

// GetByID retrieves a user by ID
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*UserResponse, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return ToResponse(user), nil
}

// GetByEmail retrieves a user by email
func (s *Service) GetByEmail(ctx context.Context, email string) (*UserResponse, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return ToResponse(user), nil
}

// List retrieves a paginated list of users
func (s *Service) List(ctx context.Context, page, pageSize int) ([]*UserResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	const maxInt32 = int(^uint32(0) >> 1)
	if offset < 0 || offset > maxInt32 {
		return nil, fmt.Errorf("offset exceeds int32 bounds")
	}
	if pageSize > maxInt32 {
		return nil, fmt.Errorf("pageSize exceeds int32 bounds")
	}
	// #nosec G115 -- Safe conversion after bounds check above
	pageSize32 := int32(pageSize)
	// #nosec G115 -- Safe conversion after bounds check above
	offset32 := int32(offset)
	users, err := s.repo.List(ctx, pageSize32, offset32)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	responses := make([]*UserResponse, len(users))
	for i, user := range users {
		responses[i] = ToResponse(&user)
	}

	return responses, nil
}

// Update updates a user's information
func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateUserInput) (*UserResponse, error) {
	user, err := s.repo.Update(ctx, id, input.Email, input.FirstName, input.LastName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return ToResponse(user), nil
}

// Delete soft deletes a user
func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// hashPassword hashes a password using bcrypt
func hashPassword(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		// In production, handle this error properly
		return ""
	}
	return string(hash)
}
