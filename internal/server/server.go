package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pepo/internal/api"
	"pepo/internal/config"
	"pepo/internal/handlers"
	"pepo/internal/middleware"
	"pepo/internal/version"
	"pepo/templates"
)

// Server wraps the HTTP server and provides setup/shutdown methods
type Server struct {
	httpServer *http.Server
	config     *config.Config
}

// New creates a new server instance
func New(cfg *config.Config, apiHandler *handlers.CombinedAPIHandler, personHandler *handlers.PersonHandler, actionHandler *handlers.ActionHandler) (*Server, error) {
	// Create ogen server
	apiServer, err := api.NewServer(apiHandler)
	if err != nil {
		return nil, fmt.Errorf("failed to create API server: %w", err)
	}

	// Setup routes
	mux := setupRoutes(apiServer, personHandler, actionHandler)

	// Wrap with middleware
	handler := middleware.Chain(mux,
		middleware.RecoveryMiddleware,
		middleware.LoggingMiddleware,
		middleware.SecurityHeadersMiddleware,
	)

	// Add CORS middleware in development
	if cfg.IsDevelopment() {
		handler = middleware.CORSMiddleware(handler)
	}

	// Create HTTP server with timeouts
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		httpServer: httpServer,
		config:     cfg,
	}, nil
}

// Start starts the HTTP server
func (s *Server) Start() error {
	log.Printf("Server starting on port %s", s.config.Port)
	log.Printf("Health check: http://localhost:%s/health", s.config.Port)
	log.Printf("API documentation: http://localhost:%s/api/v1", s.config.Port)
	log.Printf("Web interface: http://localhost:%s/", s.config.Port)

	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server failed to start: %w", err)
	}
	return nil
}

// StartWithGracefulShutdown starts the server and handles graceful shutdown
func (s *Server) StartWithGracefulShutdown() error {
	// Start server in a goroutine
	go func() {
		if err := s.Start(); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	log.Println("Server exited")
	return nil
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// setupRoutes configures all HTTP routes
func setupRoutes(apiServer *api.Server, personHandler *handlers.PersonHandler, actionHandler *handlers.ActionHandler) *http.ServeMux {
	mux := http.NewServeMux()

	// Health check endpoint (both at root and API level)
	healthHandler := createHealthHandler()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/api/v1/health", healthHandler)

	// Root endpoint - serve the main HTML page using templ
	mux.HandleFunc("/", handleRootPage)

	// Form handlers for HTMX - Person routes
	mux.HandleFunc("/forms/persons/create", personHandler.HandleCreatePersonForm)
	mux.HandleFunc("/forms/persons/list", personHandler.HandleListPersonsHTML)
	mux.HandleFunc("/forms/persons/delete/", personHandler.HandleDeletePersonForm)
	mux.HandleFunc("/forms/persons/select", personHandler.HandleGetPersonsForSelect)

	// Form handlers for HTMX - Action routes
	mux.HandleFunc("/forms/actions/create", actionHandler.HandleCreateActionForm)
	mux.HandleFunc("/forms/actions/list", actionHandler.HandleListActionsHTML)
	mux.HandleFunc("/forms/actions/delete/", actionHandler.HandleDeleteActionForm)

	// API routes (mount the ogen server)
	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", apiServer))

	// Static file serving for development
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	return mux
}

// createHealthHandler creates a health check handler
func createHealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		timestamp := time.Now().Format(time.RFC3339)
		versionInfo := version.Get()
		response := fmt.Sprintf(`{"status":"ok","timestamp":"%s","version":"%s","commit":"%s","go_version":"%s"}`,
			timestamp, versionInfo.Version, versionInfo.Commit, versionInfo.GoVersion)
		w.Write([]byte(response))
	}
}

// handleRootPage handles the root page request
func handleRootPage(w http.ResponseWriter, r *http.Request) {
	// Only serve the root path, return 404 for any other path
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)

	// Render the main page template
	if err := templates.Index().Render(r.Context(), w); err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
