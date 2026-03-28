# Sanjow Nova API (SNAPI)

A Go-based REST API using **Gin**, **pgx** (PostgreSQL driver), and **sqlc** (SQL code generator) for type-safe database operations. Environment management powered by **dotenvx** with encryption support.

> [Read the Complete Documentation](PGX_SQLC_DOCUMENTATION.md) — All documentation in one file, organized for easy reading.

## Why pgx + sqlc?

- **High Performance**: pgx is faster than database/sql
- **Type Safety**: Compile-time checked SQL queries
- **Full SQL Control**: Write raw SQL with complete control
- **No ORM Magic**: Explicit and predictable
- **Easy Testing**: Test with real SQL queries

## Quick Start

### Prerequisites

- Go 1.26+
- PostgreSQL 14+
- [dotenvx](https://dotenvx.com) — environment management with encryption

### 1. First-Time Setup

```bash
make setup    # Installs dotenvx, sqlc, downloads deps, generates code
```

### 2. Configure Environment

Environment variables are managed with **dotenvx** — the Go code reads only `os.Getenv()`, and dotenvx injects vars before the process starts.

Env files are split by environment and concern:

| File | Purpose |
|------|---------|
| `.env.dev.core` | Development core: DATABASE_URL, JWT_SECRET, SERVER_PORT, LOG_FORMAT, SKIP_DB |
| `.env.dev.svcs` | Development services: DOCS_USERNAME, DOCS_PASSWORD |
| `.env.prod.core` | Production core vars |
| `.env.prod.svcs` | Production services vars |

Set values using Makefile targets:

```bash
make env-dev-core-set KEY=DATABASE_URL VAL=postgres://user:password@localhost:5432/sanjow?sslmode=disable
make env-dev-core-set KEY=JWT_SECRET VAL=your-secret-key-minimum-32-chars
make env-dev-core-set KEY=LOG_FORMAT VAL=text
```

Verify a value:

```bash
make env-dev-core-get KEY=DATABASE_URL
```

### 3. Run Migrations

```bash
make db-migrate
```

### 4. Generate Database Code

```bash
make sqlc-generate
```

This generates type-safe Go code in `internal/database/db/` from your SQL queries.

### 5. Run the Application

```bash
make run    # Loads .env.dev.core + .env.dev.svcs via dotenvx
```

The API starts on `http://localhost:8080`. Visit the landing page for links to health check and API docs.

### 6. Test the API

```bash
# Health check
curl http://localhost:8080/health

# API Documentation (Redoc)
open http://localhost:8080/docs

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

### 7. Run with Hot Reload (Local Development)

```bash
make dev    # Installs Air if needed, loads dev env via dotenvx
```

Automatically restarts the API when Go source files change.

### 8. Lint the Code

```bash
make lint    # Installs golangci-lint if needed
```

## Available Commands

### Development

```bash
make run               # Run API server (loads dev env via dotenvx)
make dev               # Hot reload with Air (loads dev env via dotenvx)
make db-migrate        # Run database migrations (loads dev env via dotenvx)
make test              # Run all tests
make build             # Build binary to bin/api
make lint              # Run golangci-lint
make tidy              # Tidy go modules
make deps              # Download dependencies
make clean             # Clean build artifacts
```

### Code Generation

```bash
make sqlc-generate     # Generate Go code from SQL queries
make docs              # Generate Swagger/OpenAPI docs
```

### Environment Management (dotenvx)

```bash
# Dev core
make env-dev-core-set  KEY=x VAL=y    # Set variable
make env-dev-core-get  KEY=x          # Get variable

# Dev services
make env-dev-svcs-set  KEY=x VAL=y
make env-dev-svcs-get  KEY=x

# Prod core
make env-prod-core-set KEY=x VAL=y
make env-prod-core-get KEY=x

# Prod services
make env-prod-svcs-set KEY=x VAL=y
make env-prod-svcs-get KEY=x

# Encryption
make env-encrypt       # Encrypt all env files (keys saved to .env.keys)
make env-decrypt       # Decrypt all env files
```

### Tool Installation

```bash
make setup             # Full first-time setup (dotenvx, sqlc, deps, codegen)
make dotenvx-install   # Install dotenvx CLI
make sqlc-install      # Install sqlc
make air-install       # Install Air (hot reload)
make swag-install      # Install swag (OpenAPI generation)
```

### Docker

```bash
make docker-build      # Build Docker image
make docker-run        # Run container (pass DOTENV_PRIVATE_KEY)
make docker-push       # Push image to registry (set DOCKER_REGISTRY)
```

## Project Structure

```
.
├── cmd/
│   ├── api/                # Main API application
│   └── migrate/            # Migration runner
├── config/                 # Configuration (reads os.Getenv, no dotenv library)
├── internal/
│   ├── database/
│   │   ├── db/             # Generated sqlc code (auto-generated)
│   │   ├── migrations/     # SQL migration files
│   │   ├── queries/        # SQL query definitions
│   │   ├── database.go     # Connection pool setup
│   │   └── migrator.go     # Migration runner
│   ├── domain/
│   │   ├── auth/           # Authentication (login, JWT)
│   │   └── user/           # User CRUD operations
│   └── shared/
│       ├── apperror/       # Domain error types with HTTP mapping
│       ├── httputil/       # Shared HTTP utilities (HandleError)
│       ├── logging/        # Colored structured logging (slog)
│       └── middleware/     # HTTP middleware (auth, request logging)
├── web/                    # Embedded web assets
│   ├── static/             # Static files (logo, etc.)
│   ├── templates/          # HTML templates (landing page)
│   └── embed.go            # Go embed directives
├── docs/                   # Generated API documentation (swagger)
├── .env.dev.core           # Dev core env vars (gitignored)
├── .env.dev.svcs           # Dev services env vars (gitignored)
├── .env.prod.core          # Prod core env vars (gitignored / encrypted)
├── .env.prod.svcs          # Prod services env vars (gitignored / encrypted)
├── sqlc.yaml               # sqlc configuration
├── Dockerfile              # Container build (with dotenvx)
├── Makefile                # Development commands
└── README.md
```

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `DATABASE_URL` | Yes | — | PostgreSQL connection URL |
| `JWT_SECRET` | Yes | — | Secret key for JWT signing (min 32 chars) |
| `SERVER_PORT` | No | 8080 | HTTP server port |
| `LOG_FORMAT` | No | json | `"text"` (colored) or `"json"` (production) |
| `SKIP_DB` | No | false | Run without database connection |
| `DOCS_USERNAME` | No | — | Basic Auth username for /docs |
| `DOCS_PASSWORD` | No | — | Basic Auth password for /docs |

### Encryption Workflow

dotenvx encrypts env files so they can be safely committed to git:

```bash
# 1. Set your variables
make env-prod-core-set KEY=DATABASE_URL VAL=postgres://prod-host:5432/sanjow

# 2. Encrypt all files (private keys saved to .env.keys, gitignored)
make env-encrypt

# 3. Commit the encrypted env files
git add .env.prod.core .env.prod.svcs

# 4. At runtime, pass the private key to decrypt
DOTENV_PRIVATE_KEY=your-key make docker-run
```

## API Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/` | No | Landing page |
| GET | `/health` | No | Health check |
| GET | `/docs` | Optional | API documentation (Redoc) |
| GET | `/swagger.json` | Optional | OpenAPI spec |
| POST | `/api/login` | No | Authenticate user and get JWT |
| POST | `/api/users` | No | Create user |
| GET | `/api/users` | No | List users |
| GET | `/api/users/:id` | No | Get user by ID |
| PUT | `/api/users/:id` | Yes | Update user |
| DELETE | `/api/users/:id` | Yes | Delete user |

## Database Layer

### 1. Define Schema (Migrations)

Create SQL files in `internal/database/migrations/`:

```sql
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

```sql
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

### 3. Generate and Use

```bash
make sqlc-generate
```

```go
queries := db.New(dbPool)
post, err := queries.CreatePost(ctx, db.CreatePostParams{
    UserID:  userID,
    Title:   "My Post",
    Content: "Post content here",
})
```

## Architecture

Clean Architecture with domain-driven design:

```
Handler → Service → Repository → sqlc Queries → PostgreSQL
  ↓         ↓          ↓
HTTP    Business    Data Access
Layer    Logic       Layer
```

Each domain (`auth/`, `user/`) is self-contained with handler, service, and repository layers. Cross-domain dependencies use Go's consumer-defines-interface pattern.

## Development Workflow

1. **Add new feature**:
   - Create migration file in `internal/database/migrations/`
   - Add SQL queries in `internal/database/queries/`
   - Run `make sqlc-generate`
   - Create repository, service, and handler
   - Wire up in `cmd/api/main.go`

2. **Test locally**:
   - Configure env: `make env-dev-core-set KEY=... VAL=...`
   - `make db-migrate` — Run migrations
   - `make sqlc-generate` — Generate code
   - `make run` — Start API

3. **Before committing**:
   - `make lint` — Run linter
   - `make test` — Run tests
   - `make build` — Verify build

## Deployment

### Docker

The Dockerfile produces a minimal Alpine-based image with dotenvx installed:

```bash
# Build
make docker-build

# Run (pass private key to decrypt encrypted env files)
DOTENV_PRIVATE_KEY="your-key" make docker-run

# Push to registry
DOCKER_TAG=v1.0.0 DOCKER_REGISTRY=your-registry.com make docker-push
```

The image includes encrypted `.env.prod.*` files. At runtime, `DOTENV_PRIVATE_KEY` is used by dotenvx to decrypt them before starting the API.

## Resources

- [Complete Documentation](PGX_SQLC_DOCUMENTATION.md)
- [pgx Documentation](https://github.com/jackc/pgx)
- [sqlc Documentation](https://docs.sqlc.dev/)
- [dotenvx Documentation](https://dotenvx.com/docs)
- [PostgreSQL Docs](https://www.postgresql.org/docs/)

## License

MIT
