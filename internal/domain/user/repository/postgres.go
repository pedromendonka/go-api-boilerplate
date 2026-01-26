// Package repository provides the PostgreSQL implementation for user repository.
package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"sanjow-nova-api/internal/database/db"
)

// Repository handles user data operations
type Repository struct {
	queries *db.Queries
	pool    *pgxpool.Pool
}

// New creates a new user repository
func New(pool *pgxpool.Pool) *Repository {
	return &Repository{
		queries: db.New(pool),
		pool:    pool,
	}
}

// Create creates a new user
func (r *Repository) Create(ctx context.Context, email, passwordHash, firstName, lastName string) (*db.User, error) {
	user, err := r.queries.CreateUser(ctx, db.CreateUserParams{
		Email:        email,
		PasswordHash: passwordHash,
		FirstName:    &firstName,
		LastName:     &lastName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	return &user, nil
}

// GetByID retrieves a user by ID
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*db.User, error) {
	user, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *Repository) GetByEmail(ctx context.Context, email string) (*db.User, error) {
	user, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return &user, nil
}

// List retrieves a paginated list of users
func (r *Repository) List(ctx context.Context, limit, offset int32) ([]db.User, error) {
	users, err := r.queries.ListUsers(ctx, db.ListUsersParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return users, nil
}

// Update updates a user's information
func (r *Repository) Update(ctx context.Context, id uuid.UUID, email, firstName, lastName *string) (*db.User, error) {
	user, err := r.queries.UpdateUser(ctx, db.UpdateUserParams{
		ID:        id,
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}
	return &user, nil
}

// UpdatePassword updates a user's password hash
func (r *Repository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	err := r.queries.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		ID:           id,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return fmt.Errorf("failed to update user password: %w", err)
	}
	return nil
}

// SoftDelete soft deletes a user
func (r *Repository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	err := r.queries.SoftDeleteUser(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to soft delete user: %w", err)
	}
	return nil
}

// HardDelete permanently deletes a user
func (r *Repository) HardDelete(ctx context.Context, id uuid.UUID) error {
	err := r.queries.HardDeleteUser(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to hard delete user: %w", err)
	}
	return nil
}

// Count returns the total number of active users
func (r *Repository) Count(ctx context.Context) (int64, error) {
	count, err := r.queries.CountUsers(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}
	return count, nil
}
