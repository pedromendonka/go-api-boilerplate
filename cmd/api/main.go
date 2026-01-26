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

	// Initialize database connection
	logger.Info("connecting to database")
	dbPool, err := database.NewPool(ctx, cfg.Database.URL)
	if err != nil {
		logger.Error("failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer cancel()
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
	userDomain := user.New(dbPool, cfg.JWT.Secret)
	authDomain := auth.New(userDomain.Repository, cfg.JWT.Secret)

	// Initialize Gin router
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.RequestID())
	router.Use(middleware.Logger(logger))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
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

	// Register API routes
	api := router.Group("/api")
	{
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
