// Package main provides the entry point for the API server.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"sanjow-main-api/config"
	"sanjow-main-api/internal/database"
	"sanjow-main-api/internal/database/db"
	"sanjow-main-api/internal/domain/auth"
	"sanjow-main-api/internal/domain/user"
	"sanjow-main-api/internal/shared/logging"
	"sanjow-main-api/internal/shared/middleware"
)

const (
	appName        = "sanjow-main-api"
	appDisplayName = "Sanjow Main API"
	appVersion     = "1.0.0"
)

// @title Sanjow Main API
// @version 1.0.0
// @description REST API for user authentication and management
// @host localhost:8080
// @BasePath /api
// @schemes http https
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description JWT token in format: Bearer {token}
func main() {
	printBanner()

	// Load and validate configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}

	// Initialize structured logger
	logCfg := logging.DefaultConfig()
	logCfg.Format = cfg.Log.Format
	logger := logging.New(logCfg)
	slog.SetDefault(logger)

	logger.Info("starting application",
		slog.String("app", appName),
		slog.String("version", appVersion),
	)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize database connection (unless SKIP_DB is set)
	var dbPool *pgxpool.Pool
	var userDomain *user.Domain
	var authDomain *auth.Domain

	if cfg.SkipDB {
		logger.Warn("SKIP_DB is set - running without database connection",
			slog.String("note", "API endpoints requiring DB will fail"),
		)
	} else {
		logger.Info("connecting to database")
		var err error
		dbPool, err = database.NewPool(ctx, cfg.Database.URL)
		if err != nil {
			logger.Error("failed to connect to database", slog.String("error", err.Error()))
			os.Exit(1)
		}
		defer database.Close(dbPool)
		logger.Info("database connected")

		// Run migrations
		logger.Info("running migrations")
		migrator := database.NewMigrator(dbPool, "internal/database/migrations")
		if err := migrator.Migrate(ctx); err != nil {
			logger.Warn("migration failed", slog.String("error", err.Error()))
		} else {
			logger.Info("migrations completed")
		}

		// Initialize sqlc queries
		queries := db.New(dbPool)
		_ = queries

		// Initialize domains
		userDomain = user.New(dbPool, cfg.JWT.Secret)
		authDomain = auth.New(userDomain.Repository, cfg.JWT.Secret)
	}

	// Initialize Gin router
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.RequestID())
	router.Use(middleware.Logger(logger))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		if cfg.SkipDB {
			c.JSON(http.StatusOK, gin.H{
				"status":   "healthy",
				"database": "skipped",
				"mode":     "no-db",
			})
			return
		}
		if err := dbPool.Ping(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unhealthy",
				"error":  "database connection failed",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":   "healthy",
			"database": "connected",
		})
	})

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	// API Documentation
	router.GET("/docs", serveRedoc)
	router.GET("/swagger.json", serveSwaggerJSON)

	// Register API routes (only when DB is available)
	api := router.Group("/api")
	if !cfg.SkipDB {
		userDomain.RegisterRoutes(api)
		authDomain.RegisterRoutes(api)
	}

	// Start server
	srv := &http.Server{
		Addr:         cfg.Server.Addr(),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("server started",
			slog.Int("port", cfg.Server.Port),
			slog.String("health", fmt.Sprintf("http://localhost:%d/health", cfg.Server.Port)),
			slog.String("api_base", fmt.Sprintf("http://localhost:%d/api", cfg.Server.Port)),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server forced to shutdown", slog.String("error", err.Error()))
		return
	}

	logger.Info("server exited gracefully")
}

func printBanner() {
	banner := `
╔═══════════════════════════════════════════════════════════╗
║                                                           ║
║   ███████╗ █████╗ ███╗   ██╗     ██╗ ██████╗ ██╗    ██╗   ║
║   ██╔════╝██╔══██╗████╗  ██║     ██║██╔═══██╗██║    ██║   ║
║   ███████╗███████║██╔██╗ ██║     ██║██║   ██║██║ █╗ ██║   ║
║   ╚════██║██╔══██║██║╚██╗██║██   ██║██║   ██║██║███╗██║   ║
║   ███████║██║  ██║██║ ╚████║╚█████╔╝╚██████╔╝╚███╔███╔╝   ║
║   ╚══════╝╚═╝  ╚═╝╚═╝  ╚═══╝ ╚════╝  ╚═════╝  ╚══╝╚══╝    ║
║                                                           ║
╚═══════════════════════════════════════════════════════════╝`
	fmt.Println(banner)
	fmt.Printf("\n  %s v%s\n\n", appDisplayName, appVersion)
}

func serveRedoc(c *gin.Context) {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Sanjow API Documentation</title>
    <meta charset="utf-8"/>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>body { margin: 0; padding: 0; }</style>
</head>
<body>
    <redoc spec-url='/swagger.json'></redoc>
    <script src="https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js"></script>
</body>
</html>`
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

func serveSwaggerJSON(c *gin.Context) {
	c.File("docs/swagger.json")
}
