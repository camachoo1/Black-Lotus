package main

import (
	"log"
	"time"

	"black-lotus/internal/api"
	"black-lotus/internal/db"
	"black-lotus/internal/utils"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Load environment variables
	utils.LoadEnv()

	// Initialize database connection
	if err := db.Initialize(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()
	log.Println("Successfully connected to PostgreSQL")

	// Start the cleanup job for expired records
	db.StartCleanupJob(1 * time.Hour) // Run cleanup every hour
	log.Println("Started database cleanup job")

	// Initialize Echo
	e := echo.New()
	
	// Add middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	
	// Rate limiting to prevent abuse
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(20))) // 20 requests per second

	// Setup routes
	api.RegisterRoutes(e)

	// Get port from environment
	port := utils.GetEnvWithDefault("SERVER_PORT", "8080")

	// Start server
	log.Printf("Server starting on port %s", port)
	log.Fatal(e.Start(":" + port))
}