# Sanjow Main API

Go REST API backend service for user authentication and management.

## Tech Stack

- **Language:** Go 1.25
- **Framework:** Gin v1.11
- **Database:** PostgreSQL with pgx/v5
- **Code Generation:** sqlc (SQL-first, type-safe queries)
- **Auth:** JWT (HS256)
- **Password Hashing:** bcrypt

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
    ctx/                # Context keys
    middleware/         # HTTP middleware (JWT auth)
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

## Architecture

Clean Architecture with domain-driven design:

```
HTTP Handler → Service → Repository → Database
```

- **Handlers:** HTTP request/response handling
- **Services:** Business logic
- **Repositories:** Data access layer

Each domain (auth, user) is self-contained with handler/service/repository layers. Cross-domain dependencies use Go interfaces defined by the consumer (e.g., auth defines what it needs from user).

## Key Patterns

### Database Queries

SQL-first approach using sqlc. Edit queries in `internal/database/queries/`, then run `make sqlc-generate`. Never edit files in `internal/database/db/` directly.

### Error Handling

- Return explicit errors (Go idiom)
- Use `ErrUserNotFound` style sentinel errors
- Wrap errors with context: `fmt.Errorf("context: %w", err)`

### API Response Format

```json
{"data": {...}}     // Success
{"error": "msg"}    // Error
```

### Authentication

Protected routes use middleware chain:
1. `middleware.Auth()` - Verify JWT Bearer token
2. `injectCurrentUser()` - Load user from context

### Dependency Injection

Constructor-based DI via `New()` functions. Go idiom: consumer defines interfaces it needs (e.g., auth service defines `UserRepository` interface that user.Repository satisfies).

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
