package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
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
	// Create content negotiating handler
	contentHandler := handlers.NewContentNegotiatingHandler(apiHandler)

	// Create ogen server with content negotiating handler
	apiServer, err := api.NewServer(contentHandler)
	if err != nil {
		return nil, fmt.Errorf("failed to create API server: %w", err)
	}

	// Setup routes
	mux := setupRoutes(apiServer, personHandler, actionHandler)

	// Wrap with middleware
	handler := middleware.Chain(mux,
		middleware.AddRequestToContext,
		middleware.FormToJSONMiddleware,
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

	// Legacy form handlers for HTMX compatibility (keeping for now)
	mux.HandleFunc("/forms/people/select", personHandler.HandleGetPersonsForSelect)

	// Consolidated API routes with content negotiation (supports both JSON and HTML)
	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", apiServer))

	// Convenience routes that serve the same endpoints without /api/v1 prefix
	mux.Handle("/people/", createConvenienceHandler(apiServer, "/people"))
	mux.Handle("/people", createConvenienceHandler(apiServer, "/people"))
	mux.HandleFunc("/actions/", createActionHandler(apiServer, actionHandler))
	mux.Handle("/actions", createConvenienceHandler(apiServer, "/actions"))

	// Static file serving for development
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	return mux
}

// createConvenienceHandler creates a handler that forwards requests to the API server
// with proper path mapping for convenience routes
func createConvenienceHandler(apiServer *api.Server, prefix string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Forward the request to the API server
		apiServer.ServeHTTP(w, r)
	}
}

// createActionHandler handles action routes and serves a dedicated edit page
// while forwarding other requests to the API server.
func createActionHandler(apiServer *api.Server, actionHandler *handlers.ActionHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/edit") {
			id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/actions/"), "/edit")
			params := api.GetActionByIdParams{ID: id}
			res, err := actionHandler.GetActionById(r.Context(), params)
			if err != nil {
				log.Printf("Error getting action: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			switch action := res.(type) {
			case *api.Action:
				tmplAction := templates.Action{
					ID:          action.ID,
					PersonID:    action.PersonID,
					OccurredAt:  action.OccurredAt,
					Description: action.Description,
					References:  action.References.Or(""),
					Valence:     string(action.Valence),
					CreatedAt:   action.CreatedAt,
					UpdatedAt:   action.UpdatedAt,
				}
				w.Header().Set("Content-Type", "text/html")
				w.WriteHeader(http.StatusOK)
				if err := templates.EditActionPage(tmplAction).Render(r.Context(), w); err != nil {
					log.Printf("Error rendering template: %v", err)
				}
				return
			case *api.GetActionByIdNotFound:
				http.NotFound(w, r)
				return
			default:
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		}

		// Forward other requests to the API server
		apiServer.ServeHTTP(w, r)
	}
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
