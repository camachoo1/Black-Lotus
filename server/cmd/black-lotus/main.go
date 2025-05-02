package main

import (
	"log"
	"os"
	"time"

	"black-lotus/api"
	"black-lotus/db"
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

	// Create and configure the server
	server := api.NewServer()

	// Setup routes
	api.SetupRouter(server.Echo())

	// Get port from environment or use default
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	log.Printf("Server starting on port %s", port)
	log.Fatal(server.Start(port))
}
