# Sanjow Main API 
# Complete PGX + SQLC Documentation
## PostgreSQL with pgx + sqlc

> **Last Updated:** December 29, 2025  
> **Version:** 1.0  
> **Status:** Production Ready ✅

---

**Table of Contents**

1. [Overview](#overview)
2. [Getting Started](#getting-started)
3. [Quick Start Guide](#quick-start-guide)
4. [Project Architecture](#project-architecture)
5. [API Testing](#api-testing)
6. [Development Workflow](#development-workflow)
7. [Quick Reference](#quick-reference)
8. [Troubleshooting](#troubleshooting)

---

# Overview

Sanjow API has been configured with **pgx** (high-performance PostgreSQL driver) and **sqlc** (SQL code generator) for type-safe database operations.

## Why pgx + sqlc?

### Key Benefits
- ✅ **2-10x faster** than GORM (no reflection overhead)
- ✅ **Compile-time type safety** - SQL errors caught before runtime
- ✅ **Full SQL control** - write any query you need
- ✅ **Explicit code** - no magic, easy to debug
- ✅ **Production-ready** - used by major companies

### Technology Stack
- **pgx/v5** - High-performance PostgreSQL driver
- **sqlc** - Type-safe SQL code generation
- **Gin** - Fast HTTP framework
- **Clean Architecture** - Handler → Service → Repository → Database

### Comparison with GORM

| Feature | GORM | pgx + sqlc |
|---------|------|------------|
| Performance | Slow (reflection) | **2-10x Faster** ⚡ |
| Type Safety | Runtime errors | **Compile-time** 🛡️ |
| SQL Visibility | Hidden queries | **Explicit SQL** 👀 |
| Query Control | Limited | **Full Control** 🎯 |
| Complex Queries | Awkward | **Natural SQL** ✨ |
| Debugging | Difficult | **Easy** 🐛 |
| Learning Curve | ORM concepts | Standard SQL |
| Memory Usage | Higher | **40-60% Lower** 💾 |

---

# Getting Started

## Build Status: SUCCESS ✓

The API compiles successfully with:
- ✅ pgx/v5 - High-performance PostgreSQL driver
- ✅ sqlc - Type-safe SQL code generation
- ✅ Clean architecture with handlers, services, and repositories
- ✅ User CRUD operations fully implemented

## Prerequisites

Before you start, ensure you have:

- **Go 1.25+** installed
- **Docker & Docker Compose** (or PostgreSQL 14+)
- **sqlc** code generator (installed via `make sqlc-install`)

## Quick Start (3 Steps)

### Step 1: Start PostgreSQL

```bash
make docker-up
```

This starts:
- PostgreSQL on `localhost:5432`
- pgAdmin on `http://localhost:5050` (login: admin@sanjow.com / admin)

**Alternative (if you have PostgreSQL already):**
```bash
createdb -h localhost -U postgres sanjow
```

### Step 2: Run Migrations

```bash
make db-migrate
```

This creates the users table and schema_migrations table.

### Step 3: Start the API

```bash
make run
```

The API will start on `http://localhost:8080`.

## Verify Your Setup

In another terminal:

```bash
# Health check
curl http://localhost:8080/health

# Expected response:
# {"status":"healthy","database":"connected"}

# Create a user
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "password123",
    "first_name": "John",
    "last_name": "Doe"
  }'

# List users
curl http://localhost:8080/api/users
```

---

# Quick Start Guide

## How It Works

### 1. Define Your Schema

Create SQL migration files in `internal/database/migrations/`:

```text
-- internal/database/migrations/001_init.sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;
```

### 2. Write SQL Queries

Create query files in `internal/database/queries/`:

```text
-- internal/database/queries/users.sql

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1 AND deleted_at IS NULL;

-- name: CreateUser :one
INSERT INTO users (email, password_hash, first_name, last_name)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListUsers :many
SELECT * FROM users 
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateUser :one
UPDATE users
SET 
    email = COALESCE($2, email),
    first_name = COALESCE($3, first_name),
    last_name = COALESCE($4, last_name),
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteUser :exec
UPDATE users SET deleted_at = NOW() WHERE id = $1;
```

### 3. Generate Go Code

Run `make sqlc-generate` and sqlc creates:

```text
// internal/database/db/querier.go
type Querier interface {
    GetUserByID(ctx context.Context, id uuid.UUID) (User, error)
    CreateUser(ctx context.Context, arg CreateUserParams) (User, error)
    ListUsers(ctx context.Context, arg ListUsersParams) ([]User, error)
    UpdateUser(ctx context.Context, arg UpdateUserParams) (User, error)
    SoftDeleteUser(ctx context.Context, id uuid.UUID) error
}

// internal/database/db/models.go
type User struct {
    ID           uuid.UUID          `db:"id" json:"id"`
    Email        string             `db:"email" json:"email"`
    PasswordHash string             `db:"password_hash" json:"password_hash"`
    FirstName    *string            `db:"first_name" json:"first_name"`
    LastName     *string            `db:"last_name" json:"last_name"`
    CreatedAt    time.Time          `db:"created_at" json:"created_at"`
    UpdatedAt    time.Time          `db:"updated_at" json:"updated_at"`
    DeletedAt    pgtype.Timestamptz `db:"deleted_at" json:"deleted_at"`
}
```

### 4. Use in Your Code

```text
// Create queries instance
queries := db.New(dbPool)

// Use type-safe methods
user, err := queries.GetUserByID(ctx, userID)
if err != nil {
    return err
}

// All parameters are type-checked at compile time!

newUser, err := queries.CreateUser(ctx, db.CreateUserParams{
    Email:        "user@example.com",
    PasswordHash: hashedPassword,
    FirstName:    &firstName,
    LastName:     &lastName,
})
```

## sqlc Query Annotations

- **`:one`** - Returns single row (or error if not found)
- **`:many`** - Returns slice of rows (empty slice if none)
- **`:exec`** - Executes without returning data
- **`:execrows`** - Returns number of affected rows

## Example: Adding a New Query

Let's say you want to search users by email:

**1. Add SQL query:**
```text
-- internal/database/queries/users.sql
-- name: SearchUsersByEmail :many
SELECT * FROM users 
WHERE email LIKE $1 AND deleted_at IS NULL
ORDER BY created_at DESC;
```

**2. Generate code:**
```bash
make sqlc-generate
```

**3. Use in repository:**
```text
// internal/domain/user/repository/postgres.go
func (r *Repository) SearchByEmail(ctx context.Context, pattern string) ([]db.User, error) {
    return r.queries.SearchUsersByEmail(ctx, pattern+"%")
}
```

**4. That's it!** Type-safe, compiled, ready to use.

---

# Project Architecture

## File Structure

```
sanjow-main-api/
│
├── 📄 Configuration Files
│   ├── sqlc.yaml              ← sqlc configuration
│   ├── docker-compose.yml     ← PostgreSQL + pgAdmin
│   ├── .env.example           ← Environment template
│   ├── Makefile               ← Development commands
│   └── go.mod                 ← Go dependencies
│
├── 🗄️ Database Layer
│   └── internal/database/
│       ├── database.go        ← Connection pool setup
│       ├── migrator.go        ← Migration runner
│       ├── migrations/        ← SQL schema files (YOU WRITE)
│       │   └── 001_init.sql
│       ├── queries/           ← SQL query definitions (YOU WRITE)
│       │   └── users.sql
│       └── db/                ← Generated by sqlc (DON'T EDIT)
│           ├── db.go
│           ├── models.go
│           ├── querier.go
│           └── users.sql.go
│
├── 🎯 Domain Layer
│   └── internal/domain/user/
│       ├── user.go            ← Domain wiring
│       ├── handler/           ← HTTP handlers
│       │   └── http.go
│       ├── service/           ← Business logic
│       │   └── service.go
│       └── repository/        ← Data access (uses sqlc)
│           └── postgres.go
│
├── 🔧 Shared Internal Packages
│   └── internal/shared/
│       ├── ctx/               ← Context keys
│       ├── errors/            ← Error utilities
│       ├── logger/            ← Logging
│       ├── middleware/        ← HTTP middleware (auth)
│       ├── response/          ← API response helpers
│       └── utils/             ← Utility functions
│
├── ⚙️ Config
│   └── config/
│       └── config.go          ← Configuration management
│
└── 🚀 Applications
    └── cmd/
        ├── api/main.go        ← Main API application
        └── migrate/main.go    ← Migration runner
```

## Clean Architecture Flow

```
┌─────────────────────────────────────────────┐
│           HTTP Request                       │
└─────────────────┬───────────────────────────┘
                  ↓
┌─────────────────────────────────────────────┐
│  Handler Layer (internal/domain/*/handler)   │
│  • Validates input                           │
│  • Parses requests                           │
│  • Returns responses                         │
└─────────────────┬───────────────────────────┘
                  ↓
┌─────────────────────────────────────────────┐
│  Service Layer (internal/domain/*/service)   │
│  • Business logic                            │
│  • Orchestrates operations                   │
│  • Error handling                            │
└─────────────────┬───────────────────────────┘
                  ↓
┌─────────────────────────────────────────────┐
│ Repository Layer (internal/domain/*/repo)    │
│  • Wraps sqlc queries                        │
│  • Data access abstraction                   │
└─────────────────┬───────────────────────────┘
                  ↓
┌─────────────────────────────────────────────┐
│  Generated Code (internal/database/db/)      │
│  • Type-safe queries                         │
│  • Auto-generated by sqlc                    │
└─────────────────┬───────────────────────────┘
                  ↓
┌─────────────────────────────────────────────┐
│  pgx Driver                                  │
│  • High-performance PostgreSQL driver        │
└─────────────────┬───────────────────────────┘
                  ↓
┌─────────────────────────────────────────────┐
│  PostgreSQL Database                         │
└─────────────────────────────────────────────┘
```

## Key Files to Know

### You Write SQL Here:
- `internal/database/migrations/*.sql` - Database schema
- `internal/database/queries/*.sql` - SQL queries

### sqlc Generates This (Don't Edit):
- `internal/database/db/*.go` - Type-safe Go code from your SQL

### Your Business Logic:
- `internal/domain/user/user.go` - Domain wiring
- `internal/domain/user/repository/postgres.go` - Data access
- `internal/domain/user/service/service.go` - Business logic
- `internal/domain/user/handler/http.go` - HTTP handlers

### Shared Internal Packages:
- `internal/shared/ctx/` - Context keys for typed context values
- `internal/shared/middleware/` - HTTP middleware (auth)
- `internal/shared/errors/` - Error utilities
- `internal/shared/response/` - API response helpers
- `internal/shared/logger/` - Logging
- `internal/shared/utils/` - Utility functions

### Configuration:
- `sqlc.yaml` - sqlc configuration
- `config/config.go` - App configuration
- `.env` - Environment variables (create from `.env.example`)

---

# API Testing

## Available Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Health check with DB status |
| GET | `/ping` | Simple ping endpoint |
| POST | `/api/users` | Create new user |
| GET | `/api/users` | List users (paginated) |
| GET | `/api/users/:id` | Get user by ID |
| PUT | `/api/users/:id` | Update user |
| DELETE | `/api/users/:id` | Delete user (soft delete) |

## Health Check

```bash
# Check API health
curl http://localhost:8080/health

# Expected response:
{
  "status": "healthy",
  "database": "connected"
}

# Ping endpoint
curl http://localhost:8080/ping

# Expected response:
{
  "message": "pong"
}
```

## User Endpoints

### Create a User

```bash
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "password123",
    "first_name": "John",
    "last_name": "Doe"
  }'
```

**Expected Response (201 Created):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "john@example.com",
  "first_name": "John",
  "last_name": "Doe",
  "created_at": "2025-12-28T10:30:00Z",
  "updated_at": "2025-12-28T10:30:00Z"
}
```

### Get User by ID

```bash
# Replace {user_id} with actual UUID
curl http://localhost:8080/api/users/{user_id}
```

**Example:**
```bash
curl http://localhost:8080/api/users/550e8400-e29b-41d4-a716-446655440000
```

**Expected Response (200 OK):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "john@example.com",
  "first_name": "John",
  "last_name": "Doe",
  "created_at": "2025-12-28T10:30:00Z",
  "updated_at": "2025-12-28T10:30:00Z"
}
```

### List Users

```bash
# Default pagination (page 1, 20 items)
curl http://localhost:8080/api/users

# With pagination parameters
curl "http://localhost:8080/api/users?page=1&page_size=10"
```

**Expected Response (200 OK):**
```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "email": "john@example.com",
      "first_name": "John",
      "last_name": "Doe",
      "created_at": "2025-12-28T10:30:00Z",
      "updated_at": "2025-12-28T10:30:00Z"
    }
  ],
  "page": 1,
  "page_size": 20
}
```

### Update User

```bash
# Replace {user_id} with actual UUID
curl -X PUT http://localhost:8080/api/users/{user_id} \
  -H "Content-Type: application/json" \
  -d '{
    "email": "newemail@example.com",
    "first_name": "Jane",
    "last_name": "Smith"
  }'
```

**Expected Response (200 OK):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "newemail@example.com",
  "first_name": "Jane",
  "last_name": "Smith",
  "created_at": "2025-12-28T10:30:00Z",
  "updated_at": "2025-12-28T10:35:00Z"
}
```

### Delete User (Soft Delete)

```bash
# Replace {user_id} with actual UUID
curl -X DELETE http://localhost:8080/api/users/{user_id}
```

**Expected Response (204 No Content):**
```
(Empty response body)
```

## Error Responses

### 404 Not Found
```json
{
  "error": "user not found"
}
```

### 400 Bad Request
```json
{
  "error": "invalid user id"
}
```

### 409 Conflict
```json
{
  "error": "user already exists"
}
```

## Using HTTPie (Alternative to curl)

If you have HTTPie installed:

```bash
# Create user
http POST localhost:8080/api/users \
  email=john@example.com \
  password=password123 \
  first_name=John \
  last_name=Doe

# Get user
http localhost:8080/api/users/{user_id}

# List users
http localhost:8080/api/users page==1 page_size==10

# Update user
http PUT localhost:8080/api/users/{user_id} \
  email=newemail@example.com

# Delete user
http DELETE localhost:8080/api/users/{user_id}
```

## Complete Testing Workflow

1. **Start the database:**
```bash
make docker-up
```

2. **Run migrations:**
```bash
make db-migrate
```

3. **Generate sqlc code:**
```bash
make sqlc-generate
```

4. **Start the API:**
```bash
make run
```

5. **Test the endpoints** using the curl commands above

---

# Development Workflow

## Common Tasks

### Adding a New Table

**1. Create migration:**
```text
-- internal/database/migrations/002_posts.sql
CREATE TABLE posts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    title VARCHAR(255) NOT NULL,
    content TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_posts_user_id ON posts(user_id);
```

**2. Run migration:**
```bash
make db-migrate
```

### Adding New Queries

**1. Create/edit query file:**
```text
-- internal/database/queries/posts.sql
-- name: CreatePost :one
INSERT INTO posts (user_id, title, content)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetPostsByUser :many
SELECT * FROM posts WHERE user_id = $1 ORDER BY created_at DESC;
```

**2. Generate code:**
```bash
make sqlc-generate
```

**3. Use in repository:**
```text
// internal/domain/post/repository/postgres.go
package repository

import (
    "context"
    "sanjow-main-api/internal/database/db"
    "github.com/google/uuid"
    "github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
    queries *db.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
    return &Repository{
        queries: db.New(pool),
    }
}

func (r *Repository) Create(ctx context.Context, userID uuid.UUID, title, content string) (*db.Post, error) {
    post, err := r.queries.CreatePost(ctx, db.CreatePostParams{
        UserID:  userID,
        Title:   title,
        Content: &content,
    })
    return &post, err
}

func (r *Repository) GetByUser(ctx context.Context, userID uuid.UUID) ([]db.Post, error) {
    return r.queries.GetPostsByUser(ctx, userID)
}
```

### Updating Existing Queries

**1. Edit the SQL file:**
```text
-- internal/database/queries/users.sql
-- name: GetActiveUsers :many
SELECT * FROM users WHERE deleted_at IS NULL ORDER BY created_at DESC;
```

**2. Regenerate:**
```bash
make sqlc-generate
```

**3. Use the new method:**
```text
activeUsers, err := queries.GetActiveUsers(ctx)
```

### Database Migrations Best Practices

- ✅ **Never edit old migrations** - always create new ones
- ✅ **Use sequential numbering** - 001, 002, 003, etc.
- ✅ **One change per migration** - easier to track and rollback
- ✅ **Test migrations** - run them on a test database first
- ✅ **Use TIMESTAMPTZ** - not TIMESTAMP (for timezone support)
- ✅ **Add indexes** - for foreign keys and frequently queried columns

### Development Loop

```
1. Write/update SQL  → internal/database/queries/ or migrations/
2. Generate code     → make sqlc-generate (or make db-migrate)
3. Implement logic   → repository/service/handler
4. Build & test      → make run
5. Verify            → curl or API client
```

## Environment Configuration

### Environment Variables

Copy `.env.example` to `.env` and configure:

```env
# Server Configuration
SERVER_PORT=8080

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=sanjow
DB_SSLMODE=disable
```

### Docker Services

The `docker-compose.yml` provides:

- **PostgreSQL 16** - Database server
- **pgAdmin 4** - Web-based database management

Access pgAdmin at `http://localhost:5050`:
- Email: `admin@sanjow.com`
- Password: `admin`

### sqlc Configuration

The `sqlc.yaml` file configures code generation:

```yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "internal/database/queries"
    schema: "internal/database/migrations"
    gen:
      go:
        package: "db"
        out: "internal/database/db"
        sql_package: "pgx/v5"
        emit_json_tags: true
        emit_interface: true
        emit_db_tags: true
        overrides:
          - db_type: "uuid"
            go_type: "github.com/google/uuid.UUID"
          - db_type: "timestamptz"
            go_type:
              import: "time"
              type: "Time"
```

---

# Quick Reference

## Essential Commands

```bash
# Setup & Installation
make sqlc-install      # Install sqlc code generator
make deps              # Download Go dependencies
make tidy              # Tidy go.mod

# Database
make docker-up         # Start PostgreSQL & pgAdmin
make docker-down       # Stop Docker containers
make docker-logs       # View PostgreSQL logs
make db-migrate        # Run database migrations
make db-create         # Create database (without Docker)
make db-drop           # Drop database (without Docker)

# Code Generation
make sqlc-generate     # Generate Go code from SQL

# Development
make run               # Run the API
make build             # Build binary to bin/api
make test              # Run tests
make clean             # Clean generated files

# Utilities
make help              # Show all available commands
```

## Quick Start (Copy & Paste)

```bash
# Start everything
make docker-up && sleep 5 && make db-migrate && make run

# In another terminal, test it
curl http://localhost:8080/health
```

## Database Connection Info

- **Host:** localhost
- **Port:** 5432
- **Database:** sanjow
- **User:** postgres
- **Password:** postgres
- **pgAdmin:** http://localhost:5050 (admin@sanjow.com / admin)

## File Locations Quick Reference

| What | Where |
|------|-------|
| **SQL Schema** | `internal/database/migrations/*.sql` |
| **SQL Queries** | `internal/database/queries/*.sql` |
| **Generated Code** | `internal/database/db/*.go` (auto-generated) |
| **Domain Wiring** | `internal/domain/*/user.go` |
| **Repositories** | `internal/domain/*/repository/postgres.go` |
| **Services** | `internal/domain/*/service/service.go` |
| **Handlers** | `internal/domain/*/handler/http.go` |
| **Shared Packages** | `internal/shared/*` |
| **Main App** | `cmd/api/main.go` |
| **Config** | `sqlc.yaml`, `config/config.go` |

## Common Workflow

```
1. Write SQL      → internal/database/queries/
2. Generate       → make sqlc-generate
3. Use in code    → internal/domain/*/repository/postgres.go
4. Build & test   → make run
```

## Tips & Best Practices

- ✅ Always run `make sqlc-generate` after changing SQL files
- ✅ Use `TIMESTAMPTZ` instead of `TIMESTAMP` for timezone support
- ✅ Non-nullable timestamps → `time.Time`
- ✅ Nullable timestamps → `pgtype.Timestamptz`
- ✅ Write SQL in query files, not in Go code
- ✅ Let sqlc handle type conversions
- ✅ Create new migrations, don't edit old ones
- ✅ One migration per database change

---

# Troubleshooting

## Common Issues & Solutions

### "cannot find package db"

**Problem:** The generated database code doesn't exist.

**Solution:**
```bash
make sqlc-generate
```

### "connection refused"

**Problem:** PostgreSQL is not running.

**Solution:**
```bash
# Check if Docker is running
docker ps

# Start PostgreSQL
make docker-up

# Wait a few seconds for PostgreSQL to start
sleep 5

# Try again
make run
```

### "relation does not exist" or "table does not exist"

**Problem:** Database migrations haven't been run.

**Solution:**
```bash
# Run migrations
make db-migrate

# Verify tables exist
psql -h localhost -U postgres -d sanjow -c "\dt"
```

### "sqlc: command not found"

**Problem:** sqlc is not installed or not in PATH.

**Solution:**
```bash
# Install sqlc
make sqlc-install

# Or manually
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Verify installation
~/go/bin/sqlc version
```

### Build errors after changing SQL

**Problem:** Generated code is out of sync with SQL.

**Solution:**
```bash
# Clean and regenerate
rm -rf internal/database/db
make sqlc-generate

# Rebuild
go build cmd/api/main.go
```

### Type mismatch errors (UUID, Timestamp)

**Problem:** sqlc configuration might be incorrect.

**Solution:**
Check `sqlc.yaml` has proper overrides:
```yaml
overrides:
  - db_type: "uuid"
    go_type: "github.com/google/uuid.UUID"
  - db_type: "timestamptz"
    go_type:
      import: "time"
      type: "Time"
```

Then regenerate:
```bash
rm -rf internal/database/db
make sqlc-generate
```

### Docker port already in use

**Problem:** Port 5432 or 8080 is already in use.

**Solution:**
```bash
# Check what's using the port
lsof -i :5432
lsof -i :8080

# Stop conflicting services or change ports in .env
```

### Migration already applied

**Problem:** Trying to run the same migration twice.

**Solution:**
Migrations track what's been applied. Just create a new migration file with a higher number (002_, 003_, etc.).

### Can't connect to pgAdmin

**Problem:** pgAdmin container not running or wrong credentials.

**Solution:**
```bash
# Check if containers are running
docker-compose ps

# Restart if needed
make docker-down
make docker-up

# Access at http://localhost:5050
# Email: admin@sanjow.com
# Password: admin
```

## Debugging Tips

### Check Database Connection

```bash
# Via psql
psql -h localhost -U postgres -d sanjow

# List tables
\dt

# Query users
SELECT * FROM users;

# Check migrations
SELECT * FROM schema_migrations;
```

### Verify Generated Code

```bash
# List generated files
ls -la internal/database/db/

# Should see:
# - db.go
# - models.go
# - querier.go
# - users.sql.go
```

### Check Application Logs

```bash
# Run with verbose logging
make run

# Or build and run manually
go build -o /tmp/api cmd/api/main.go
/tmp/api
```

### Verify Makefile Commands

```bash
# Show all available commands
make help

# Test build without running
make build

# Check if binary was created
ls -la bin/
```

## Getting Help

If you encounter issues not covered here:

1. **Check the documentation files** in the repository
2. **Verify your setup** against the checklist
3. **Check error logs** when running the application
4. **Verify PostgreSQL is running**: `docker ps`
5. **Verify generated code exists**: `ls internal/database/db/`
6. **Check Go version**: `go version` (should be 1.25+)

---

# Appendix

## Performance Benchmarks

Compared to GORM:

| Operation | GORM | pgx+sqlc | Improvement |
|-----------|------|----------|-------------|
| Simple SELECT | 1.0x | **2-3x** | 🚀 2-3x faster |
| INSERT | 1.0x | **2-3x** | 🚀 2-3x faster |
| Complex JOIN | 1.0x | **3-5x** | 🚀 3-5x faster |
| Bulk INSERT | 1.0x | **5-10x** | 🚀 5-10x faster |
| Memory Usage | High | **Low** | 💾 40-60% less |

## Resources

### Official Documentation
- [pgx GitHub](https://github.com/jackc/pgx) - PostgreSQL driver documentation
- [sqlc Documentation](https://docs.sqlc.dev/) - sqlc guide and reference
- [PostgreSQL Docs](https://www.postgresql.org/docs/) - PostgreSQL documentation
- [Gin Framework](https://gin-gonic.com/docs/) - Gin web framework

### Learning Resources
- [Go by Example](https://gobyexample.com/) - Go programming examples
- [PostgreSQL Tutorial](https://www.postgresqltutorial.com/) - Learn PostgreSQL
- [SQL Style Guide](https://www.sqlstyle.guide/) - SQL best practices

## Project Information

### Dependencies

```text
require (
    github.com/gin-gonic/gin v1.11.0      // Web framework
    github.com/google/uuid v1.6.0          // UUID support
    github.com/jackc/pgx/v5 v5.8.0         // PostgreSQL driver
)
```

### License

MIT License - See LICENSE file for details

### Contributors

Built with ❤️ for high-performance Go APIs

---

**End of Documentation**

For the most up-to-date information, check the individual MD files in the repository or visit the project repository.

**Last Updated:** December 29, 2025
