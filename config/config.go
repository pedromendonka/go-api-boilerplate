// Package config provides configuration loading and validation for the application.
package config

import (
	"fmt"
	"os"
)

// Config holds all application configuration.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
}

// ServerConfig holds server-related configuration.
type ServerConfig struct {
	Port string
}

// DatabaseConfig holds database connection configuration.
type DatabaseConfig struct {
	URL string
}

// JWTConfig holds JWT-related configuration.
type JWTConfig struct {
	Secret string
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
		},
		Database: DatabaseConfig{
			URL: getEnv("DATABASE_URL", ""),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", ""),
		},
	}
}

// Validate checks that all required environment variables are set
func (c *Config) Validate() error {
	var missing []string

	if c.Database.URL == "" {
		missing = append(missing, "DATABASE_URL")
	}
	if c.JWT.Secret == "" {
		missing = append(missing, "JWT_SECRET")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %v", missing)
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
