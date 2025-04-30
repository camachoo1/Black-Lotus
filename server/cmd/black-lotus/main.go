package main

import (
	"log"
	"net/http"
	"os"
	"time"

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
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
        AllowOrigins:     []string{"http://localhost:3000"},
        AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
        AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, "X-CSRF-TOKEN"},
				ExposeHeaders: []string{"Set-Cookie"},
        AllowCredentials: true, // This is crucial for sending cookies
    }))
	e.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
				TokenLookup:    "header:X-CSRF-Token",
				CookieName:     "csrf_token",
				CookiePath:     "/",
				CookieHTTPOnly: false,
				CookieMaxAge:   3600,  // 1 hour
		}))
	
	// Rate limiting to prevent abuse
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(20))) // 20 requests per second

	// Setup routes
	api.AuthRoutes(e)

	// Get port from environment or use default
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	log.Printf("Server starting on port %s", port)
	log.Fatal(e.Start(":" + port))
}