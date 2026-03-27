package database

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Migrator handles database migrations
type Migrator struct {
	pool           *pgxpool.Pool
	migrationsPath string
	logger         *slog.Logger
}

// NewMigrator creates a new migrator instance
func NewMigrator(pool *pgxpool.Pool, migrationsPath string, logger *slog.Logger) *Migrator {
	if logger == nil {
		logger = slog.Default()
	}
	return &Migrator{
		pool:           pool,
		migrationsPath: migrationsPath,
		logger:         logger,
	}
}

// Migrate runs all pending migrations
func (m *Migrator) Migrate(ctx context.Context) error {
	// Create migrations table if it doesn't exist
	if err := m.createMigrationsTable(ctx); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get list of migration files
	files, err := m.getMigrationFiles()
	if err != nil {
		return fmt.Errorf("failed to get migration files: %w", err)
	}

	// Get applied migrations
	applied, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Run pending migrations
	for _, file := range files {
		if _, ok := applied[file]; !ok {
			if err := m.runMigration(ctx, file); err != nil {
				return fmt.Errorf("failed to run migration %s: %w", file, err)
			}
			m.logger.Info("applied migration", "file", file)
		}
	}

	return nil
}

// createMigrationsTable creates the migrations tracking table
func (m *Migrator) createMigrationsTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id SERIAL PRIMARY KEY,
			version VARCHAR(255) UNIQUE NOT NULL,
			applied_at TIMESTAMP NOT NULL DEFAULT NOW()
		);
	`
	_, err := m.pool.Exec(ctx, query)
	return err
}

// getMigrationFiles returns sorted list of migration files
func (m *Migrator) getMigrationFiles() ([]string, error) {
	entries, err := os.ReadDir(m.migrationsPath)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			files = append(files, entry.Name())
		}
	}

	sort.Strings(files)
	return files, nil
}

// getAppliedMigrations returns map of applied migration versions
func (m *Migrator) getAppliedMigrations(ctx context.Context) (map[string]bool, error) {
	query := "SELECT version FROM schema_migrations"
	rows, err := m.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}

	return applied, rows.Err()
}

// runMigration executes a single migration file
func (m *Migrator) runMigration(ctx context.Context, filename string) error {
	// Read migration file
	path := filepath.Join(m.migrationsPath, filename)
	// #nosec G304 -- The migration files are trusted and controlled, only .sql files from a known directory are loaded.
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Start transaction
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	// Execute migration
	if _, err := tx.Exec(ctx, string(content)); err != nil {
		return err
	}

	// Record migration
	query := "INSERT INTO schema_migrations (version) VALUES ($1)"
	if _, err := tx.Exec(ctx, query, filename); err != nil {
		return err
	}

	// Commit transaction
	return tx.Commit(ctx)
}
