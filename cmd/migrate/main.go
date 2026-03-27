// Package main runs database migrations for the application.
package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"time"

	"sanjow-nova-api/config"
	"sanjow-nova-api/internal/database"
	"sanjow-nova-api/internal/shared/logging"
)

func main() {
	migrationsPath := flag.String("path", "internal/database/migrations", "Path to migrations directory")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("configuration error", "error", err)
		os.Exit(1)
	}

	// Initialize structured logger
	logCfg := logging.DefaultConfig()
	logCfg.Format = cfg.Log.Format
	logger := logging.New(logCfg)
	slog.SetDefault(logger)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Connect to database
	dbPool, err := database.NewPool(ctx, cfg.Database.URL)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer database.Close(dbPool, logger)

	// Create migrator and run migrations
	migrator := database.NewMigrator(dbPool, *migrationsPath, logger)
	if err := migrator.Migrate(ctx); err != nil {
		logger.Error("migration failed", "error", err)
		os.Exit(1)
	}

	logger.Info("migrations completed successfully")
}
