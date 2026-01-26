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

func (ce ValidationErrors) Error() string {
	if len(ce) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("configuration validation failed:\n")
	for _, e := range ce {
		b.WriteString("  - ")
		b.WriteString(e.Error())
		b.WriteString("\n")
	}
	return b.String()
}

// Load loads configuration from environment variables.
func Load() *Config {
	port := defaultPort
	if portStr := os.Getenv("SERVER_PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	logFormat := defaultLogFormat
	if format := os.Getenv("LOG_FORMAT"); format != "" {
		logFormat = format
	}

	return &Config{
		Server: ServerConfig{
			Port: port,
		},
		Database: DatabaseConfig{
			URL: os.Getenv("DATABASE_URL"),
		},
		JWT: JWTConfig{
			Secret: os.Getenv("JWT_SECRET"),
		},
		Log: LogConfig{
			Format: logFormat,
		},
	}
}

// Validate checks that all configuration values are valid.
func (c *Config) Validate() error {
	var errs ValidationErrors

	// Validate server port
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		errs = append(errs, &ValidationError{
			Field:   "SERVER_PORT",
			Message: fmt.Sprintf("invalid port %d", c.Server.Port),
			Hint:    "must be between 1 and 65535",
		})
	}

	// Validate database URL
	if c.Database.URL == "" {
		errs = append(errs, &ValidationError{
			Field:   "DATABASE_URL",
			Message: "required but not set",
			Hint:    "format: postgres://user:pass@host:port/dbname",
		})
	} else if err := validateDatabaseURL(c.Database.URL); err != nil {
		errs = append(errs, &ValidationError{
			Field:   "DATABASE_URL",
			Message: err.Error(),
			Hint:    "format: postgres://user:pass@host:port/dbname",
		})
	}

	// Validate JWT secret
	if c.JWT.Secret == "" {
		errs = append(errs, &ValidationError{
			Field:   "JWT_SECRET",
			Message: "required but not set",
			Hint:    fmt.Sprintf("minimum %d characters recommended", minJWTSecretLen),
		})
	} else if len(c.JWT.Secret) < minJWTSecretLen {
		errs = append(errs, &ValidationError{
			Field:   "JWT_SECRET",
			Message: fmt.Sprintf("too short (%d characters)", len(c.JWT.Secret)),
			Hint:    fmt.Sprintf("minimum %d characters recommended for security", minJWTSecretLen),
		})
	}

	// Validate log format
	if c.Log.Format != "json" && c.Log.Format != "text" {
		errs = append(errs, &ValidationError{
			Field:   "LOG_FORMAT",
			Message: fmt.Sprintf("invalid format %q", c.Log.Format),
			Hint:    "must be \"json\" or \"text\"",
		})
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
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
