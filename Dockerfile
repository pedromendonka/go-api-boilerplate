# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Install dependencies for building
RUN apk add --no-cache git ca-certificates

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/api ./cmd/api

# Final stage - minimal image
FROM alpine:3.21

WORKDIR /app

# Install dotenvx from pinned release (no curl-pipe-sh)
ARG DOTENVX_VERSION=1.30.1
ARG TARGETARCH=amd64
RUN apk add --no-cache ca-certificates tzdata curl && \
    curl -fsSL "https://github.com/dotenvx/dotenvx/releases/download/v${DOTENVX_VERSION}/dotenvx-linux-${TARGETARCH}.tar.gz" \
      -o /tmp/dotenvx.tar.gz && \
    tar -xzf /tmp/dotenvx.tar.gz -C /usr/local/bin dotenvx && \
    chmod +x /usr/local/bin/dotenvx && \
    rm /tmp/dotenvx.tar.gz && \
    apk del curl

# Create non-root user for security
RUN adduser -D -g '' appuser

# Copy binary from builder
COPY --from=builder /app/api .

# Copy migrations (needed at runtime)
COPY --from=builder /app/internal/database/migrations ./internal/database/migrations

# Copy encrypted env files (decrypted at runtime via DOTENV_PRIVATE_KEY)
COPY .env.prod.core .env.prod.svcs ./

# Use non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application (dotenvx decrypts env files using DOTENV_PRIVATE_KEY)
CMD ["dotenvx", "run", "-f", ".env.prod.core", "-f", ".env.prod.svcs", "--", "./api"]
