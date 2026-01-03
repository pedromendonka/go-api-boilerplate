# Build stage
FROM golang:1.25-alpine AS builder

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
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates for HTTPS calls
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user for security
RUN adduser -D -g '' appuser

# Copy binary from builder
COPY --from=builder /app/api .

# Copy migrations (needed at runtime)
COPY --from=builder /app/internal/database/migrations ./internal/database/migrations

# Use non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./api"]

