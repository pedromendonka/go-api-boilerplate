# Sanjow Main API

A Go-based REST API using **pgx** (PostgreSQL driver) and **sqlc** (SQL code generator) for type-safe database operations.

> 📖 **[Read the Complete Documentation](PGX_SQLC_DOCUMENTATION.md)** - All documentation in one file, organized for easy reading!

## Why pgx + sqlc?

- ✅ **High Performance**: pgx is faster than database/sql
- ✅ **Type Safety**: Compile-time checked SQL queries
- ✅ **Full SQL Control**: Write raw SQL with complete control
- ✅ **No ORM Magic**: Explicit and predictable
- ✅ **Easy Testing**: Test with real SQL queries

## Quick Start

### Prerequisites

- Go 1.25+
- PostgreSQL 14+ (external database)
- sqlc

### 1. Clone and Install Dependencies

```bash
cd /path/to/sanjow-main-api
go mod download
```

### 2. Install sqlc

```bash
make sqlc-install
```

### 3. Configure Database Connection

Copy `.env.example` to `.env` and set your PostgreSQL connection URL:

```bash
cp .env.example .env
```

Edit `.env` with your database credentials:
```env
DATABASE_URL=postgres://user:password@host:5432/dbname?sslmode=disable
```

### 4. Run Migrations

```bash
make db-migrate
```

### 5. Generate Database Code

```bash
make sqlc-generate
```

This generates type-safe Go code in `internal/database/db/` from your SQL queries.

### 6. Run the Application

```bash
make run
```

The API will start on `http://localhost:8080`.

### 7. Test the API

```bash
# Health check
curl http://localhost:8080/health

# Create a user
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "first_name": "Test",
    "last_name": "User"
  }'
```

See [Complete Documentation](PGX_SQLC_DOCUMENTATION.md) for more examples.

### 8. Run with Hot Reload (Local Development)

Install [Air](https://github.com/cosmtrek/air) if you haven't already:

```bash
go install github.com/cosmtrek/air@latest
```

Then start the API with hot reload:

```bash
make dev
```

This will automatically restart the API when you change Go source files (except generated code in `internal/database/db/`).

### 9. Lint the Code (Required Before Commit/Push)

Install [golangci-lint](https://github.com/golangci/golangci-lint) if you haven't already:

```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

Run linting before every commit or push:

```bash
make lint
```

This will check your code for style, bugs, and best practices using a standard set of linters. Linting is also recommended in CI for team-wide code quality.

## Project Structure

```
.
├── cmd/
│   ├── api/              # Main API application
│   └── migrate/          # Migration runner
├── config/               # Configuration management
├── internal/
│   ├── database/
│   │   ├── db/          # Generated sqlc code (auto-generated)
│   │   ├── migrations/  # SQL migration files
│   │   ├── queries/     # SQL query definitions
│   │   ├── database.go  # Connection pool setup
│   │   └── migrator.go  # Migration runner
│   ├── domain/
│   │   └── user/        # User domain
│   │       ├── user.go       # Domain wiring
│   │       ├── handler/      # HTTP handlers
│   │       │   └── http.go
│   │       ├── service/      # Business logic
│   │       │   └── service.go
│   │       └── repository/   # Data access
│   │           └── postgres.go
│   └── shared/           # Internal shared packages
│       ├── ctx/          # Context keys
│       ├── errors/       # Error utilities
│       ├── logger/       # Logging
│       ├── middleware/   # HTTP middleware (auth)
│       ├── response/     # API response helpers
│       └── utils/        # Utility functions
├── sqlc.yaml            # sqlc configuration
├── Dockerfile           # Container build configuration
├── Makefile             # Development commands
└── README.md
```

## Available Commands

```bash
make help              # Show all available commands
make setup             # Install sqlc, download deps, generate sqlc code
make sqlc-generate     # Generate Go code from SQL
make db-migrate        # Run database migrations
make run               # Run the application
make dev               # Run with hot reload (Air)
make build             # Build the application
make test              # Run tests
make lint              # Run golangci-lint
make clean             # Clean generated files
make docker-build      # Build Docker image
make docker-run        # Run Docker container locally
make docker-push       # Push image to registry
```

## Database Layer

### 1. Define Schema (Migrations)

Create SQL files in `internal/database/migrations/`:

```text
-- internal/database/migrations/002_add_posts.sql
CREATE TABLE posts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    title VARCHAR(255) NOT NULL,
    content TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

### 2. Write Queries

Create query files in `internal/database/queries/`:

```text
-- internal/database/queries/posts.sql

-- name: CreatePost :one
INSERT INTO posts (user_id, title, content)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetPostByID :one
SELECT * FROM posts WHERE id = $1;

-- name: ListPostsByUser :many
SELECT * FROM posts WHERE user_id = $1 ORDER BY created_at DESC;
```

### 3. Generate Code

```bash
make sqlc-generate
```

### 4. Use in Your Application

```text
// In your repository
queries := db.New(dbPool)

post, err := queries.CreatePost(ctx, db.CreatePostParams{
    UserID:  userID,
    Title:   "My Post",
    Content: "Post content here",
})
```

## Architecture Pattern

This project follows Clean Architecture / Domain-Driven Design:

```
Handler → Service → Repository → Database
  ↓         ↓          ↓
HTTP    Business    Data Access
Layer    Logic       Layer
```

**Example Flow**:
1. **Handler** receives HTTP request, validates input
2. **Service** contains business logic, orchestrates operations
3. **Repository** wraps sqlc queries, handles data access
4. **Generated Code** (sqlc) provides type-safe database operations

## Environment Variables

Copy `.env.example` to `.env` and configure:

```env
SERVER_PORT=8080

# PostgreSQL connection URL
DATABASE_URL=postgres://user:password@host:5432/dbname?sslmode=disable

JWT_SECRET=your-super-secret-key-change-in-production
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Health check |
| GET | `/ping` | Ping endpoint |
| POST | `/api/login` | Authenticate user and get JWT |
| POST | `/api/users` | Create user |
| GET | `/api/users` | List users |
| GET | `/api/users/:id` | Get user by ID |
| PUT | `/api/users/:id` | Update user |
| DELETE | `/api/users/:id` | Delete user |

See [Complete Documentation](PGX_SQLC_DOCUMENTATION.md) for detailed examples.

## Development Workflow

1. **Add new feature**:
   - Create migration file in `internal/database/migrations/`
   - Add SQL queries in `internal/database/queries/`
   - Run `make sqlc-generate`
   - Create repository, service, and handler
   - Wire up in `cmd/api/main.go`

2. **Test locally**:
   - Configure `DATABASE_URL` in `.env`
   - `make db-migrate` - Run migrations
   - `make sqlc-generate` - Generate code
   - `make run` - Start API
   - Test with curl or Postman

3. **Before committing**:
   - `make lint` - Run linter
   - `make test` - Run tests
   - `make build` - Verify build

## Deployment

### Build Docker Image

```bash
make docker-build
```

This creates a minimal Alpine-based image (~20MB) with:
- Multi-stage build for small image size
- Non-root user for security
- Health check endpoint configured
- Migrations included

### Run with Docker

```bash
# Set environment variables and run
DATABASE_URL="postgres://user:pass@host:5432/db" \
JWT_SECRET="your-secret" \
make docker-run
```

### Push to Registry

```bash
# Push to a container registry
DOCKER_REGISTRY=your-registry.com make docker-push

# Or with custom tag
DOCKER_TAG=v1.0.0 DOCKER_REGISTRY=your-registry.com make docker-push
```

### Environment Variables for Production

| Variable | Required | Description |
|----------|----------|-------------|
| `DATABASE_URL` | Yes | PostgreSQL connection URL |
| `JWT_SECRET` | Yes | Secret key for JWT signing |
| `SERVER_PORT` | No | Server port (default: 8080) |

## Resources

- 📖 **[Complete Documentation](PGX_SQLC_DOCUMENTATION.md)**
- 📚 [pgx Documentation](https://github.com/jackc/pgx)
- 📚 [sqlc Documentation](https://docs.sqlc.dev/)
- 📚 [PostgreSQL Docs](https://www.postgresql.org/docs/)

## License

MIT

