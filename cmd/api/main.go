// Package main provides the entry point for the API server.
package main

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"sanjow-nova-api/config"
	"sanjow-nova-api/internal/database"
	"sanjow-nova-api/internal/database/db"
	"sanjow-nova-api/internal/domain/auth"
	"sanjow-nova-api/internal/domain/user"
	"sanjow-nova-api/internal/shared/logging"
	"sanjow-nova-api/internal/shared/middleware"
	"sanjow-nova-api/web"
)

const (
	appName        = "sanjow-nova-api"
	appDisplayName = "Sanjow Nova API"
	appVersion     = "1.0.0"
)

// @title Sanjow Nova API (SNAPI)
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

	// Landing page
	router.GET("/", serveLanding)

	// Static files (logo, etc.)
	staticFS, _ := fs.Sub(web.Assets, "static")
	router.StaticFS("/static", http.FS(staticFS))

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

	// API Documentation (optionally protected with Basic Auth)
	if cfg.DocsAuth.Enabled() {
		docsAuth := middleware.BasicAuth(cfg.DocsAuth.Username, cfg.DocsAuth.Password)
		router.GET("/docs", docsAuth, serveRedoc)
		router.GET("/swagger.json", docsAuth, serveSwaggerJSON)
		logger.Info("docs authentication enabled")
	} else {
		router.GET("/docs", serveRedoc)
		router.GET("/swagger.json", serveSwaggerJSON)
	}

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
			slog.String("url", fmt.Sprintf("http://localhost:%d/", cfg.Server.Port)),
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
	const colorOrange = "\033[38;5;208m"
	const colorReset = "\033[0m"

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
	fmt.Println(colorOrange + banner + colorReset)
	fmt.Printf("\n  %s v%s\n\n", appDisplayName, appVersion)
}

func serveLanding(c *gin.Context) {
	html, err := web.Assets.ReadFile("templates/landing.html")
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to load landing page")
		return
	}
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, string(html))
}

func serveRedoc(c *gin.Context) {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>SNAPI - Sanjow Nova API Documentation</title>
    <meta charset="utf-8"/>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;600;700&display=swap" rel="stylesheet">
    <style>
        body {
            margin: 0;
            padding: 0;
            background-color: #1a1a1a;
        }
        .custom-header {
            background-color: #1a1a1a;
            padding: 1rem 2rem;
            display: flex;
            align-items: center;
            gap: 1rem;
            border-bottom: 1px solid #333;
        }
        .custom-header img {
            height: 40px;
            width: auto;
        }
        .custom-header h1 {
            font-family: 'Inter', sans-serif;
            font-size: 1.25rem;
            font-weight: 600;
            color: #FF6B00;
            margin: 0;
        }
    </style>
</head>
<body>
    <div class="custom-header">
        <img src="/static/logo.png" alt="SNAPI" onerror="this.style.display='none'">
        <h1>SNAPI - Sanjow Nova API</h1>
    </div>
    <redoc
        spec-url='/swagger.json'
        theme='{
            "colors": {
                "primary": { "main": "#FF6B00" },
                "text": { "primary": "#ffffff", "secondary": "#b3b3b3" },
                "http": {
                    "get": "#61affe",
                    "post": "#49cc90",
                    "put": "#fca130",
                    "delete": "#f93e3e"
                }
            },
            "typography": {
                "fontFamily": "Inter, -apple-system, BlinkMacSystemFont, sans-serif",
                "headings": { "fontFamily": "Inter, sans-serif" },
                "code": { "fontFamily": "JetBrains Mono, Monaco, monospace" }
            },
            "sidebar": {
                "backgroundColor": "#1a1a1a",
                "textColor": "#ffffff"
            },
            "rightPanel": {
                "backgroundColor": "#0d0d0d"
            },
            "schema": {
                "nestedBackground": "#262626"
            }
        }'
    ></redoc>
    <script src="https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js"></script>
</body>
</html>`
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

func serveSwaggerJSON(c *gin.Context) {
	c.File("docs/swagger.json")
}
