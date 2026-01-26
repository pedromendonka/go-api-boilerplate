// Package config provides configuration loading and validation for the application.
package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

// Default configuration values.
const (
	defaultPort      = 8080
	defaultLogFormat = "json"
	minJWTSecretLen  = 32
)

// Config holds all application configuration.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Log      LogConfig
}

// ServerConfig holds server-related configuration.
type ServerConfig struct {
	Port int
}

// DatabaseConfig holds database connection configuration.
type DatabaseConfig struct {
	URL string
}

// JWTConfig holds JWT-related configuration.
type JWTConfig struct {
	Secret string
}

// LogConfig holds logging configuration.
type LogConfig struct {
	Format string // "json" or "text"
}

// ValidationError represents a configuration validation error.
type ValidationError struct {
	Field   string
	Message string
	Hint    string
}

func (e *ValidationError) Error() string {
	if e.Hint != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Field, e.Message, e.Hint)
	}
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors is a collection of configuration validation errors.
type ValidationErrors []*ValidationError

func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("configuration validation failed:\n")
	for _, e := range ve {
		b.WriteString("  - ")
		b.WriteString(e.Error())
		b.WriteString("\n")
	}
	return b.String()
}

// Load loads and validates configuration from environment variables.
// Returns an error if required variables are missing or invalid.
func Load() (*Config, error) {
	var errs ValidationErrors

	// Required environment variables
	dbURL := required("DATABASE_URL", "format: postgres://user:pass@host:port/dbname", &errs)
	jwtSecret := required("JWT_SECRET", fmt.Sprintf("minimum %d characters", minJWTSecretLen), &errs)

	// Optional environment variables with defaults
	port := optionalInt("SERVER_PORT", defaultPort, &errs)
	logFormat := optional("LOG_FORMAT", defaultLogFormat)

	// Fail fast if required vars are missing
	if len(errs) > 0 {
		return nil, errs
	}

	cfg := &Config{
		Server: ServerConfig{
			Port: port,
		},
		Database: DatabaseConfig{
			URL: dbURL,
		},
		JWT: JWTConfig{
			Secret: jwtSecret,
		},
		Log: LogConfig{
			Format: logFormat,
		},
	}

	// Validate configuration values
	if err := cfg.validate(&errs); len(errs) > 0 {
		return nil, err
	}

	return cfg, nil
}

// validate checks that all configuration values are valid.
func (c *Config) validate(errs *ValidationErrors) ValidationErrors {
	// Validate server port range
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		*errs = append(*errs, &ValidationError{
			Field:   "SERVER_PORT",
			Message: fmt.Sprintf("invalid port %d", c.Server.Port),
			Hint:    "must be between 1 and 65535",
		})
	}

	// Validate database URL format
	if err := validateDatabaseURL(c.Database.URL); err != nil {
		*errs = append(*errs, &ValidationError{
			Field:   "DATABASE_URL",
			Message: err.Error(),
			Hint:    "format: postgres://user:pass@host:port/dbname",
		})
	}

	// Validate JWT secret length
	if len(c.JWT.Secret) < minJWTSecretLen {
		*errs = append(*errs, &ValidationError{
			Field:   "JWT_SECRET",
			Message: fmt.Sprintf("too short (%d characters)", len(c.JWT.Secret)),
			Hint:    fmt.Sprintf("minimum %d characters for security", minJWTSecretLen),
		})
	}

	// Validate log format
	if c.Log.Format != "json" && c.Log.Format != "text" {
		*errs = append(*errs, &ValidationError{
			Field:   "LOG_FORMAT",
			Message: fmt.Sprintf("invalid format %q", c.Log.Format),
			Hint:    "must be \"json\" or \"text\"",
		})
	}

	return *errs
}

// validateDatabaseURL checks if the database URL is valid.
func validateDatabaseURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	if u.Scheme != "postgres" && u.Scheme != "postgresql" {
		return fmt.Errorf("invalid scheme %q, expected postgres or postgresql", u.Scheme)
	}

	if u.Host == "" {
		return fmt.Errorf("missing host")
	}

	return nil
}

// Addr returns the server address in host:port format.
func (c *ServerConfig) Addr() string {
	return fmt.Sprintf(":%d", c.Port)
}

// Environment variable helpers

// required loads a required environment variable.
// Adds an error if the variable is not set.
func required(key, hint string, errs *ValidationErrors) string {
	value := os.Getenv(key)
	if value == "" {
		*errs = append(*errs, &ValidationError{
			Field:   key,
			Message: "required but not set",
			Hint:    hint,
		})
	}
	return value
}

// optional loads an optional environment variable with a default value.
func optional(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// optionalInt loads an optional integer environment variable with a default value.
// Adds an error if the value is set but not a valid integer.
func optionalInt(key string, defaultValue int, errs *ValidationErrors) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		*errs = append(*errs, &ValidationError{
			Field:   key,
			Message: fmt.Sprintf("invalid integer %q", value),
			Hint:    "must be a valid number",
		})
		return defaultValue
	}
	return parsed
}
