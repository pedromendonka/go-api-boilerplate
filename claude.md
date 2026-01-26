# Sanjow Main API

Go REST API backend service for user authentication and management.

## Tech Stack

- **Language:** Go 1.25
- **Framework:** Gin v1.11
- **Database:** PostgreSQL with pgx/v5
- **Code Generation:** sqlc (SQL-first, type-safe queries)
- **Auth:** JWT (HS256)
- **Password Hashing:** bcrypt
- **Logging:** slog (structured JSON logging)

## Project Structure

```
cmd/
  api/main.go           # API server entry point
  migrate/main.go       # Database migration CLI
config/
  config.go             # Environment configuration
internal/
  database/
    db/                 # Generated sqlc code (DO NOT EDIT)
    migrations/         # SQL migration files
    queries/            # SQL query definitions for sqlc
    database.go         # Connection pool setup
    migrator.go         # Migration runner
  domain/
    auth/               # Authentication (login, JWT)
    user/               # User CRUD operations
  shared/
    apperror/           # Domain error types with HTTP mapping
    ctx/                # Context keys
    logging/            # Structured logging with slog
    middleware/         # HTTP middleware (auth, request logging)
```

## Commands

```bash
make setup          # Install deps & generate code
make run            # Run API server
make dev            # Hot reload with Air
make test           # Run tests
make lint           # Run golangci-lint
make sqlc-generate  # Regenerate sqlc code after SQL changes
make db-migrate     # Run database migrations
make build          # Build binary
```

## Environment Variables

Required in `.env`:
- `DATABASE_URL` - PostgreSQL connection string
- `JWT_SECRET` - Secret for signing JWTs
- `SERVER_PORT` - (optional, defaults to 8080)
- `LOG_FORMAT` - (optional, "json" or "text", defaults to "json")

## Architecture

Clean Architecture with domain-driven design:

```
HTTP Handler → Service → Repository → Database
```

- **Handlers:** HTTP request/response handling, error mapping
- **Services:** Business logic, logging
- **Repositories:** Data access layer

Each domain (auth, user) is self-contained. Cross-domain dependencies use Go interfaces defined by the consumer.

## Key Patterns

### Error Handling

Use `apperror` package for all domain errors:

```go
// Return predefined errors
return nil, apperror.ErrUserNotFound

// Wrap errors with context
return nil, apperror.Wrap(apperror.CodeInternal, "failed to create user", err)

// Check error types
if apperror.Is(err, apperror.CodeNotFound) { ... }
```

Error codes map to HTTP status:
- `NOT_FOUND` → 404
- `ALREADY_EXISTS` → 409
- `INVALID_INPUT` → 400
- `UNAUTHORIZED` → 401
- `INTERNAL_ERROR` → 500

### Logging

Use `logging` package with context propagation:

```go
logger := logging.FromContext(ctx)
logger.Info("user created", slog.String("user_id", id))
logger.Error("failed", slog.String("error", err.Error()))
```

Logs are JSON-formatted for production. Request logging middleware adds request_id, method, path, latency automatically.

### API Response Format

```json
{"id": "...", "email": "..."}           // Success
{"error": "message", "code": "CODE"}    // Error
```

### Database Queries

SQL-first approach using sqlc. Edit queries in `internal/database/queries/`, then run `make sqlc-generate`. Never edit files in `internal/database/db/` directly.

### Authentication

Protected routes use middleware chain:
1. `middleware.RequestID()` - Add request ID
2. `middleware.Logger()` - Log requests
3. `middleware.Auth(secret)` - Verify JWT Bearer token (uses `sub` claim per JWT spec)
4. `injectCurrentUser()` - Load user from context

### Dependency Injection

Constructor-based DI via `New()` functions. Go idiom: consumer defines interfaces it needs.

## API Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/health` | No | Health check |
| GET | `/ping` | No | Ping |
| POST | `/api/login` | No | Login → JWT |
| POST | `/api/users` | No | Create user |
| GET | `/api/users` | No | List users |
| GET | `/api/users/:id` | No | Get user |
| PUT | `/api/users/:id` | Yes | Update user |
| DELETE | `/api/users/:id` | Yes | Delete user |

## Testing

- Use testify for assertions and mocks
- HTTP tests with `httptest`
- Run: `make test`

## Code Style

- Run `make lint` before committing
- Follow Go naming conventions (PascalCase types, camelCase funcs)
- Use snake_case for JSON and database fields
- Keep packages focused on single responsibility
- Use structured logging, not fmt.Printf
- Always handle errors explicitly
