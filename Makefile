.PHONY: help sqlc-generate sqlc-install setup test-build db-migrate run test build tidy deps clean dev lint \
	docker-build docker-run docker-push air-install lint-install swag-install swag-generate docs dotenvx-install \
	env-prod-core-set env-prod-core-get env-prod-svcs-set env-prod-svcs-get \
	env-dev-core-set env-dev-core-get env-dev-svcs-set env-dev-svcs-get \
	env-encrypt env-decrypt

# dotenvx env files loaded for local development
DOTENVX_DEV = dotenvx run -f .env.dev.core -f .env.dev.svcs --

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# =============================================================================
# Development
# =============================================================================

run: ## Run the application (loads dev env via dotenvx)
	$(DOTENVX_DEV) go run cmd/api/main.go

dev: ## Run the API with hot reload using Air
	@$(MAKE) air-install
	@echo "Starting API with hot reload (Air)..."
	$(DOTENVX_DEV) air

db-migrate: ## Run database migrations
	$(DOTENVX_DEV) go run cmd/migrate/main.go

test: ## Run tests
	go test -v ./...

build: ## Build the application
	go build -o bin/api cmd/api/main.go

tidy: ## Tidy go modules
	go mod tidy

deps: ## Download dependencies
	go mod download

clean: ## Clean build artifacts
	rm -rf bin/
	rm -rf internal/database/db/

lint: ## Run golangci-lint on the codebase
	@$(MAKE) lint-install
	@echo "Running golangci-lint..."
	PATH="$(shell go env GOPATH)/bin:$$PATH" golangci-lint run --timeout=5m

# =============================================================================
# Code generation
# =============================================================================

sqlc-generate: ## Generate Go code from SQL queries using sqlc
	@if command -v sqlc >/dev/null 2>&1; then \
		sqlc generate; \
	elif [ -f ~/go/bin/sqlc ]; then \
		~/go/bin/sqlc generate; \
	else \
		echo "sqlc not found. Run 'make sqlc-install' first"; \
		exit 1; \
	fi

docs: swag-generate ## Generate API documentation

swag-generate: ## Generate OpenAPI spec from annotations
	@if command -v swag >/dev/null 2>&1; then \
		swag init -g cmd/api/main.go -o docs; \
	elif [ -f ~/go/bin/swag ]; then \
		~/go/bin/swag init -g cmd/api/main.go -o docs; \
	else \
		echo "swag not found. Run 'make swag-install' first"; \
		exit 1; \
	fi

# =============================================================================
# Environment management (dotenvx)
# =============================================================================

env-dev-core-set: ## Set var in .env.dev.core (KEY=x VAL=y)
	@test -n "$(KEY)" || (echo "Usage: make env-dev-core-set KEY=x VAL=y"; exit 1)
	@dotenvx set $(KEY) "$(VAL)" -f .env.dev.core

env-dev-core-get: ## Get var from .env.dev.core (KEY=x)
	@test -n "$(KEY)" || (echo "Usage: make env-dev-core-get KEY=x"; exit 1)
	@dotenvx get $(KEY) -f .env.dev.core

env-dev-svcs-set: ## Set var in .env.dev.svcs (KEY=x VAL=y)
	@test -n "$(KEY)" || (echo "Usage: make env-dev-svcs-set KEY=x VAL=y"; exit 1)
	@dotenvx set $(KEY) "$(VAL)" -f .env.dev.svcs

env-dev-svcs-get: ## Get var from .env.dev.svcs (KEY=x)
	@test -n "$(KEY)" || (echo "Usage: make env-dev-svcs-get KEY=x"; exit 1)
	@dotenvx get $(KEY) -f .env.dev.svcs

env-prod-core-set: ## Set var in .env.prod.core (KEY=x VAL=y)
	@test -n "$(KEY)" || (echo "Usage: make env-prod-core-set KEY=x VAL=y"; exit 1)
	@dotenvx set $(KEY) "$(VAL)" -f .env.prod.core

env-prod-core-get: ## Get var from .env.prod.core (KEY=x)
	@test -n "$(KEY)" || (echo "Usage: make env-prod-core-get KEY=x"; exit 1)
	@dotenvx get $(KEY) -f .env.prod.core

env-prod-svcs-set: ## Set var in .env.prod.svcs (KEY=x VAL=y)
	@test -n "$(KEY)" || (echo "Usage: make env-prod-svcs-set KEY=x VAL=y"; exit 1)
	@dotenvx set $(KEY) "$(VAL)" -f .env.prod.svcs

env-prod-svcs-get: ## Get var from .env.prod.svcs (KEY=x)
	@test -n "$(KEY)" || (echo "Usage: make env-prod-svcs-get KEY=x"; exit 1)
	@dotenvx get $(KEY) -f .env.prod.svcs

env-encrypt: ## Encrypt all env files with dotenvx
	@[ -f .env.dev.core ] && dotenvx encrypt -f .env.dev.core || echo "Skipping .env.dev.core (not found)"
	@[ -f .env.dev.svcs ] && dotenvx encrypt -f .env.dev.svcs || echo "Skipping .env.dev.svcs (not found)"
	@[ -f .env.prod.core ] && dotenvx encrypt -f .env.prod.core || echo "Skipping .env.prod.core (not found)"
	@[ -f .env.prod.svcs ] && dotenvx encrypt -f .env.prod.svcs || echo "Skipping .env.prod.svcs (not found)"
	@echo "Done. Private keys saved to .env.keys (gitignored)."

env-decrypt: ## Decrypt all env files with dotenvx
	@[ -f .env.dev.core ] && dotenvx decrypt -f .env.dev.core || echo "Skipping .env.dev.core (not found)"
	@[ -f .env.dev.svcs ] && dotenvx decrypt -f .env.dev.svcs || echo "Skipping .env.dev.svcs (not found)"
	@[ -f .env.prod.core ] && dotenvx decrypt -f .env.prod.core || echo "Skipping .env.prod.core (not found)"
	@[ -f .env.prod.svcs ] && dotenvx decrypt -f .env.prod.svcs || echo "Skipping .env.prod.svcs (not found)"

# =============================================================================
# Install tools
# =============================================================================

dotenvx-install: ## Install dotenvx CLI
	@if ! command -v dotenvx >/dev/null 2>&1; then \
		curl -sfS https://dotenvx.sh | sh; \
	else \
		echo "dotenvx already installed"; \
	fi

sqlc-install: ## Install sqlc (if not already installed)
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

air-install: ## Install Air (hot reload) if not already installed
	@if ! command -v air >/dev/null 2>&1; then \
		go install github.com/air-verse/air@latest; \
	else \
		echo "Air already installed"; \
	fi

lint-install:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

swag-install: ## Install swag CLI for OpenAPI spec generation
	go install github.com/swaggo/swag/cmd/swag@latest

setup: ## First-time setup: install tools, download deps, generate code
	@echo "== Setup: install tools, download deps, generate code =="
	@$(MAKE) dotenvx-install
	@if ! command -v sqlc >/dev/null 2>&1; then \
		$(MAKE) sqlc-install; \
	else \
		echo "sqlc already installed"; \
	fi
	@echo "Downloading go dependencies..."
	go mod download
	@$(MAKE) sqlc-generate
	@echo "Setup complete."

test-build: ## Verify generated code exists and try a module-aware build of the API
	@echo "Testing build..."
	@if [ ! -d "internal/database/db" ]; then \
		echo "Generated code not found. Run: make sqlc-generate"; exit 1; \
	fi
	@echo "Building API..."
	@go build -o /tmp/sanjow-api-test ./cmd/api && echo "Build successful!" && rm -f /tmp/sanjow-api-test

# =============================================================================
# Docker
# =============================================================================

DOCKER_IMAGE ?= sanjow-nova-api
DOCKER_TAG ?= latest

docker-build: ## Build Docker image
	@echo "Building Docker image $(DOCKER_IMAGE):$(DOCKER_TAG)..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-run: ## Run Docker container locally (export DOTENV_PRIVATE_KEY first)
	@echo "Running Docker container..."
	@test -n "$$DOTENV_PRIVATE_KEY" || (echo "Error: export DOTENV_PRIVATE_KEY before running"; exit 1)
	docker run --rm -p 8080:8080 \
		-e DOTENV_PRIVATE_KEY \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

docker-push: ## Push Docker image to registry (set DOCKER_REGISTRY)
	@if [ -z "$(DOCKER_REGISTRY)" ]; then \
		echo "Error: DOCKER_REGISTRY is not set"; exit 1; \
	fi
	docker tag $(DOCKER_IMAGE):$(DOCKER_TAG) $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):$(DOCKER_TAG)
	docker push $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):$(DOCKER_TAG)
