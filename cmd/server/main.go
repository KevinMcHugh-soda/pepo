package main

import (
	"log"

	"pepo/internal/config"
	"pepo/internal/database"
	"pepo/internal/handlers"
	"pepo/internal/server"
	"pepo/internal/version"
)

func main() {
	// Print version information
	versionInfo := version.Get()
	log.Printf("=== %s ===", versionInfo.String())

	// Load configuration from environment
	cfg := config.Load()
	log.Printf("Starting server in %s mode on port %s", cfg.Environment, cfg.Port)

	// Initialize database connection
	log.Printf("Connecting to database...")
	db, queries, err := database.Initialize(cfg.DatabaseURL, database.DefaultConnectionConfig())
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer func() {
		if err := database.Close(db); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	// Create handlers
	log.Printf("Initializing application handlers...")
	personHandler := handlers.NewPersonHandler(queries)
	actionHandler := handlers.NewActionHandler(queries)
	combinedAPIHandler := handlers.NewCombinedAPIHandler(personHandler, actionHandler)

	// Create and configure server
	log.Printf("Setting up HTTP server...")
	srv, err := server.New(cfg, combinedAPIHandler, personHandler, actionHandler)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Start server with graceful shutdown
	log.Printf("Server initialization complete, starting...")
	if err := srv.StartWithGracefulShutdown(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
