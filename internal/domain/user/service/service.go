// Package service handles user business logic.
package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"sanjow-nova-api/internal/database/db"
	"sanjow-nova-api/internal/domain/user/repository"
	"sanjow-nova-api/internal/shared/apperror"
)

// Service handles user business logic.
type Service struct {
	repo *repository.Repository
}

// New creates a new user service.
func New(repo *repository.Repository) *Service {
	return &Service{repo: repo}
}

// CreateUserInput represents input for creating a user.
type CreateUserInput struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
}

// UpdateUserInput represents input for updating a user.
type UpdateUserInput struct {
	Email     *string
	FirstName *string
	LastName  *string
}

// UserResponse represents a user response (without sensitive data).
type UserResponse struct {
	ID        uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email     string    `json:"email" example:"user@example.com"`
	FirstName *string   `json:"first_name,omitempty" example:"John"`
	LastName  *string   `json:"last_name,omitempty" example:"Doe"`
	CreatedAt time.Time `json:"created_at" example:"2024-01-15T10:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2024-01-15T10:00:00Z"`
}

// ToResponse converts a db.User to UserResponse.
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

// Create creates a new user.
func (s *Service) Create(ctx context.Context, input CreateUserInput) (*UserResponse, error) {
	if input.Email == "" || input.Password == "" {
		return nil, apperror.ErrInvalidInput
	}

	// Check if user already exists
	_, err := s.repo.GetByEmail(ctx, input.Email)
	if err == nil {
		return nil, apperror.ErrUserAlreadyExists
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to check existing user", err)
	}

	// Hash password
	passwordHash, err := hashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	// Create user
	user, err := s.repo.Create(ctx, input.Email, passwordHash, input.FirstName, input.LastName)
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to create user", err)
	}

	return ToResponse(user), nil
}

// GetByID retrieves a user by ID.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*UserResponse, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrUserNotFound
		}
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to get user", err)
	}

	return ToResponse(user), nil
}

// GetByEmail retrieves a user by email.
func (s *Service) GetByEmail(ctx context.Context, email string) (*UserResponse, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrUserNotFound
		}
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to get user", err)
	}

	return ToResponse(user), nil
}

// List retrieves a paginated list of users.
func (s *Service) List(ctx context.Context, page, pageSize int) ([]*UserResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	// #nosec G115 -- pageSize <= 100 and page is bounded, offset fits in int32
	users, err := s.repo.List(ctx, int32(pageSize), int32(offset))
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to list users", err)
	}

	responses := make([]*UserResponse, len(users))
	for i, user := range users {
		responses[i] = ToResponse(&user)
	}

	return responses, nil
}

// Update updates a user's information.
func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateUserInput) (*UserResponse, error) {
	user, err := s.repo.Update(ctx, id, input.Email, input.FirstName, input.LastName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrUserNotFound
		}
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to update user", err)
	}

	return ToResponse(user), nil
}

// Delete soft deletes a user.
func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.SoftDelete(ctx, id); err != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to delete user", err)
	}
	return nil
}

// hashPassword hashes a password using bcrypt.
func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", apperror.Wrap(apperror.CodeHashingFailed, "failed to hash password", err)
	}
	return string(hash), nil
}
