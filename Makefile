.PHONY: help sqlc-generate sqlc-install setup test-build db-migrate run test build tidy deps clean dev lint docker-build docker-run docker-push air-install lint-install

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

sqlc-generate: ## Generate Go code from SQL queries using sqlc
	@if command -v sqlc >/dev/null 2>&1; then \
		sqlc generate; \
	elif [ -f ~/go/bin/sqlc ]; then \
		~/go/bin/sqlc generate; \
	else \
		echo "sqlc not found. Run 'make sqlc-install' first"; \
		exit 1; \
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

setup: ## Install sqlc, download deps, and generate sqlc code
	@echo "== Setup: install sqlc, download deps, sqlc generate =="
	@if ! command -v sqlc >/dev/null 2>&1; then \
		$(MAKE) sqlc-install; \
	else \
		echo "sqlc already installed"; \
	fi
	@echo "Downloading go dependencies..."
	go mod download
	@$(MAKE) sqlc-generate
	@echo "Setup complete. See README.md for next steps."

test-build: ## Verify generated code exists and try a module-aware build of the API
	@echo "🧪 Testing build..."
	@if [ ! -d "internal/database/db" ]; then \
		echo "❌ Generated code not found. Run: make sqlc-generate"; exit 1; \
	fi
	@echo "Building API..."
	@go build -o /tmp/sanjow-api-test ./cmd/api && echo "✅ Build successful!" && rm -f /tmp/sanjow-api-test


db-migrate: ## Run database migrations
	go run cmd/migrate/main.go

run: ## Run the application
	go run cmd/api/main.go

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

dev: ## Run the API with hot reload using Air
	@$(MAKE) air-install
	@echo "Starting API with hot reload (Air)..."
	@air

lint: ## Run golangci-lint on the codebase
	@$(MAKE) lint-install
	@echo "Running golangci-lint..."
	PATH="$(shell go env GOPATH)/bin:$$PATH" golangci-lint run --timeout=5m

# Docker targets
DOCKER_IMAGE ?= sanjow-main-api
DOCKER_TAG ?= latest

docker-build: ## Build Docker image
	@echo "Building Docker image $(DOCKER_IMAGE):$(DOCKER_TAG)..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-run: ## Run Docker container locally
	@echo "Running Docker container..."
	docker run --rm -p 8080:8080 \
		-e DATABASE_URL="$(DATABASE_URL)" \
		-e JWT_SECRET="$(JWT_SECRET)" \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

docker-push: ## Push Docker image to registry (set DOCKER_REGISTRY)
	@if [ -z "$(DOCKER_REGISTRY)" ]; then \
		echo "Error: DOCKER_REGISTRY is not set"; exit 1; \
	fi
	docker tag $(DOCKER_IMAGE):$(DOCKER_TAG) $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):$(DOCKER_TAG)
	docker push $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):$(DOCKER_TAG)
