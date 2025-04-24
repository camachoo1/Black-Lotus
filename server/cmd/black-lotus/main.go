package main

import (
	"log"
	"time"
	"os"

	"black-lotus/api"
	"black-lotus/db"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
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
	// Get port from environment or use default
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	log.Printf("Server starting on port %s", port)
	log.Fatal(e.Start(":" + port))
}