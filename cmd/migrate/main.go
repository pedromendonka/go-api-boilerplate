// Package main runs database migrations for the application.
package main

import (
	"context"
	"flag"
	"log"
	"time"

	"sanjow-main-api/config"
	"sanjow-main-api/internal/database"
)

func main() {
	migrationsPath := flag.String("path", "internal/database/migrations", "Path to migrations directory")
	flag.Parse()

	// Load configuration
	cfg := config.Load()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	// Connect to database
	dbPool, err := database.NewPool(ctx, cfg.Database.URL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer cancel()
	defer database.Close(dbPool)

	// Create migrator and run migrations
	migrator := database.NewMigrator(dbPool, *migrationsPath)
	if err := migrator.Migrate(ctx); err != nil {
		log.Printf("Migration failed: %v", err)
		return
	}

	log.Println("Migrations completed successfully")
}
