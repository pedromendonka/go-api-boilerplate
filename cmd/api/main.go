// Package main provides the entry point for the API server.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sanjow-main-api/config"
	"sanjow-main-api/internal/database"
	"sanjow-main-api/internal/database/db"
	"sanjow-main-api/internal/domain/auth"
	"sanjow-main-api/internal/domain/user"

	"github.com/gin-gonic/gin"
)

const (
	appName    = "Sanjow Main API"
	appVersion = "1.0.0"
)

func main() {
	// Print startup banner
	printBanner()

	// Load configuration
	cfg := config.Load()

	// Validate required environment variables
	if err := cfg.Validate(); err != nil {
		log.Fatalf("❌ Configuration error: %v", err)
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())

	// Initialize database connection
	log.Println("📦 Connecting to database...")
	dbPool, err := database.NewPool(ctx, cfg.Database.URL)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer cancel()
	defer database.Close(dbPool)
	log.Println("✅ Database connected")

	// Run migrations
	log.Println("🔄 Running migrations...")
	migrator := database.NewMigrator(dbPool, "internal/database/migrations")
	if err := migrator.Migrate(ctx); err != nil {
		log.Printf("⚠️  Warning: Migration failed: %v", err)
	} else {
		log.Println("✅ Migrations completed")
	}

	// Initialize sqlc queries
	queries := db.New(dbPool)
	_ = queries // Use this in your repositories

	// Initialize user domain
	userDomain := user.New(dbPool)

	// Initialize auth domain (uses user repository directly - no circular dependency)
	authDomain := auth.New(userDomain.Repository, cfg.JWT.Secret)

	// Initialize Gin router
	router := gin.Default()

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
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	api := router.Group("/api")
	{
		userDomain.RegisterRoutes(api)
		authDomain.RegisterRoutes(api)
	}

	// Start server
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Run server in a goroutine
	go func() {
		printStartupInfo(cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("\n🛑 Shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("❌ Server forced to shutdown: %v", err)
		return
	}

	log.Println("👋 Server exited gracefully")
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
	fmt.Printf("\n  %s v%s\n\n", appName, appVersion)
}

func printStartupInfo(port string) {
	log.Println("═══════════════════════════════════════════════════════════")
	log.Printf("🚀 %s is running!", appName)
	log.Println("═══════════════════════════════════════════════════════════")
	log.Printf("   Version:     %s", appVersion)
	log.Printf("   Port:        %s", port)
	log.Printf("   Health:      http://localhost:%s/health", port)
	log.Printf("   API Base:    http://localhost:%s/api", port)
	log.Println("═══════════════════════════════════════════════════════════")
	log.Println("   Press Ctrl+C to stop")
	log.Println("═══════════════════════════════════════════════════════════")
}
