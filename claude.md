# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Sanjow Nova API (SNAPI) — Go REST API for user authentication and management. Uses Gin, PostgreSQL (pgx/v5), sqlc for type-safe queries, JWT auth (HS256), bcrypt password hashing, and slog structured logging.

## Commands

```bash
make run              # Run API server
make dev              # Hot reload with Air
make test             # Run all tests
make lint             # Run golangci-lint (installs if needed)
make build            # Build binary to bin/api
make sqlc-generate    # Regenerate code after SQL query changes
make db-migrate       # Run database migrations
make docs             # Generate Swagger/OpenAPI docs
make setup            # Install deps & generate code (first-time setup)
```

Single test: `go test -v -run TestFunctionName ./internal/domain/auth/handler/`

## Architecture

Clean Architecture with domain-driven design:

```
cmd/api/main.go → domain.New(pool, secret) → domain.RegisterRoutes(router)
                        ↓
              Handler → Service → Repository → sqlc Queries → PostgreSQL
```

### Domain Package Structure

Each domain (`auth/`, `user/`) is self-contained with the same layout:

```
internal/domain/<name>/
  <name>.go              # Domain struct + New() wiring + RegisterRoutes()
  handler/http.go        # Gin handlers, request/response types, Swagger annotations
  service/service.go     # Business logic, input/response DTOs, consumer-defined interfaces
  repository/postgres.go # Data access (sqlc queries wrapper) — user only; auth reuses user's repo
```

- **Domain root file** (`user.go`, `auth.go`) is the wiring factory — `New()` constructs all layers internally and exposes only `Domain` struct with `RegisterRoutes()`.
- **Input/response types** live in the **service layer**, not the handler.
- **`ToResponse()`** in service converts `db.User` → `UserResponse` (strips password hash).

### Cross-Domain Dependencies

Go consumer-defines-interface pattern: `auth/service` defines its own `UserRepository` interface (only needs `GetByEmail`). In `main.go`, `userDomain.Repository` (concrete) is passed to `auth.New()` and satisfies the interface implicitly.

### Dependency Injection in main.go

Manual, linear wiring: config → logger → db pool → migrate → `user.New(pool, secret)` → `auth.New(userDomain.Repository, secret)` → Gin setup → register routes. `SKIP_DB=true` skips all domain initialization (server runs with landing page, health, docs only).

## Key Patterns

### Error Handling

All domain errors use `apperror` package. Never return raw errors from service/handler layers.

```go
return nil, apperror.ErrUserNotFound          // Sentinel errors for common cases
return nil, apperror.Wrap(apperror.CodeInternal, "failed to create user", err)  // Wrap with context
```

Handler error pattern (same in all handlers):
```go
if appErr, ok := apperror.AsAppError(err); ok {
    c.JSON(appErr.HTTPStatus(), gin.H{"error": appErr.Message, "code": appErr.Code})
    return
}
c.JSON(500, gin.H{"error": "internal error", "code": apperror.CodeInternal})
```

The `Err` field (wrapped error) is for logging only — never exposed to API clients. Only `Message` and `Code` reach the response.

### Logging

Request-scoped logger propagation: `middleware.Logger()` creates a child logger with `request_id`, `method`, `path` and stores it in context. Retrieve with `logging.FromContext(ctx)`.

### Database / sqlc Workflow

1. Add/edit SQL in `internal/database/queries/*.sql`
2. Run `make sqlc-generate`
3. **Never edit** `internal/database/db/` — it's fully generated

Query annotations: `:one` (single row), `:many` (slice), `:exec` (no return). Partial updates use `COALESCE(sqlc.narg('field'), field)` pattern with `*string` nil pointers.

Soft delete pattern: all queries filter `WHERE deleted_at IS NULL`. `SoftDeleteUser` sets `deleted_at = NOW()`.

UUID type override: PostgreSQL `uuid` maps to `github.com/google/uuid.UUID` (configured in `sqlc.yaml`).

### Middleware Chain

Protected routes: `middleware.RequestID()` → `middleware.Logger()` → `middleware.Auth(secret)` → `handler.requireUser()` (verifies user still exists in DB).

Context helpers: `middleware.SetUserID(c, id)` / `middleware.GetUserID(c)` with typed `uuid.UUID`.

Docs protection: `middleware.BasicAuth(username, password)` — only applied when `DOCS_USERNAME` and `DOCS_PASSWORD` are set.

### Testing

Tests use testify (`assert`/`mock`). Hand-rolled mocks implement consumer-defined interfaces. Test pattern: create mock → wire real service with mock repo → create handler → `httptest.NewRequest` → `router.ServeHTTP` → assert status/body. Setup helper `setupTestRouter(h)` creates a Gin engine in test mode.

Currently only handler-level tests exist (`auth/handler/http_test.go`).

## Environment

Loads `.env` and `.env.local` via godotenv. Required: `DATABASE_URL`, `JWT_SECRET` (min 32 chars). Optional: `SERVER_PORT` (default 8080), `LOG_FORMAT` ("text"/"json"), `SKIP_DB`, `DOCS_USERNAME`, `DOCS_PASSWORD`.

## Deployment

Multi-stage Dockerfile: builds with `golang:1.26-alpine`, runs on `alpine:3.19` as non-root user. Migrations SQL files are copied to the image (read from disk at runtime, not embedded). Web assets (`web/`) are Go-embedded in the binary.