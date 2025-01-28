package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"crossfire-grafana/internal/routes" // Import the routes package
)

func main() {
	// Load environment variables from .env
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Get environment variables
	projectID := os.Getenv("PROJECT_ID")
	databaseID := os.Getenv("DATABASE_ID")

	if projectID == "" || databaseID == "" {
		log.Fatalf("Environment variables PROJECT_ID and DATABASE_ID must be set.")
	}

	// Set up the HTTP server
	router := routes.SetupRouter(projectID, databaseID)

	// Start the server
	log.Println("Server is running on port 4000")
	if err := router.Run(":4000"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
