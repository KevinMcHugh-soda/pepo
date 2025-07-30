package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/rs/xid"

	"pepo/internal/api"
	"pepo/internal/db"
)

func main() {
	// Load configuration from environment
	config := loadConfig()

	// Initialize database connection
	database, err := initDatabase(config.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Initialize database queries
	queries := db.New(database)

	// Create API handler
	apiHandler := &APIHandler{
		queries: queries,
	}

	// Create ogen server
	srv, err := api.NewServer(apiHandler)
	if err != nil {
		log.Fatalf("Failed to create API server: %v", err)
	}

	// Create HTTP router
	mux := setupRoutes(srv, apiHandler)

	// Initialize HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", config.Port),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %s", config.Port)
		log.Printf("Health check: http://localhost:%s/health", config.Port)
		log.Printf("API documentation: http://localhost:%s/api/v1", config.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
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

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

// Config holds application configuration
type Config struct {
	Port        string
	DatabaseURL string
	Environment string
}

// loadConfig loads configuration from environment variables
func loadConfig() *Config {
	config := &Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5433/pepo_dev?sslmode=disable"),
		Environment: getEnv("ENV", "development"),
	}

	return config
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// initDatabase initializes the database connection
func initDatabase(databaseURL string) (*sql.DB, error) {
	database, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := database.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Configure connection pool
	database.SetMaxOpenConns(25)
	database.SetMaxIdleConns(25)
	database.SetConnMaxLifetime(5 * time.Minute)

	log.Println("Database connection established")
	return database, nil
}

// APIHandler implements the generated API interface
type APIHandler struct {
	queries *db.Queries
}

// CreatePerson implements the CreatePerson operation
func (h *APIHandler) CreatePerson(ctx context.Context, req *api.CreatePersonRequest) (api.CreatePersonRes, error) {
	// Validate request
	if req.Name == "" {
		return &api.CreatePersonBadRequest{
			Message: "Name is required",
			Code:    "VALIDATION_ERROR",
		}, nil
	}

	// Generate new xid for the person
	personID := xid.New().String()

	// Create person in database
	person, err := h.queries.CreatePerson(ctx, db.CreatePersonParams{
		XidStr: personID,
		Name:   req.Name,
	})
	if err != nil {
		log.Printf("Error creating person: %v", err)
		return &api.CreatePersonInternalServerError{
			Message: "Failed to create person",
			Code:    "INTERNAL_ERROR",
		}, nil
	}

	// Convert to API response
	return &api.Person{
		ID:        person.ID,
		Name:      person.Name,
		CreatedAt: person.CreatedAt,
		UpdatedAt: person.UpdatedAt,
	}, nil
}

// GetPerson implements the GetPerson operation
func (h *APIHandler) GetPerson(ctx context.Context, params api.GetPersonParams) (api.GetPersonRes, error) {
	person, err := h.queries.GetPersonByID(ctx, params.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return &api.GetPersonNotFound{
				Message: "Person not found",
				Code:    "NOT_FOUND",
			}, nil
		}
		log.Printf("Error getting person: %v", err)
		return &api.GetPersonInternalServerError{
			Message: "Failed to get person",
			Code:    "INTERNAL_ERROR",
		}, nil
	}

	return &api.Person{
		ID:        person.ID,
		Name:      person.Name,
		CreatedAt: person.CreatedAt,
		UpdatedAt: person.UpdatedAt,
	}, nil
}

// ListPersons implements the ListPersons operation
func (h *APIHandler) ListPersons(ctx context.Context, params api.ListPersonsParams) (api.ListPersonsRes, error) {
	limit := int32(20)
	if params.Limit.IsSet() {
		limit = int32(params.Limit.Value)
	}

	offset := int32(0)
	if params.Offset.IsSet() {
		offset = int32(params.Offset.Value)
	}

	// Get total count
	total, err := h.queries.CountPersons(ctx)
	if err != nil {
		log.Printf("Error counting persons: %v", err)
		return &api.Error{
			Message: "Failed to count persons",
			Code:    "INTERNAL_ERROR",
		}, nil
	}

	// Get persons
	persons, err := h.queries.ListPersons(ctx, db.ListPersonsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		log.Printf("Error listing persons: %v", err)
		return &api.Error{
			Message: "Failed to list persons",
			Code:    "INTERNAL_ERROR",
		}, nil
	}

	// Convert to API response
	apiPersons := make([]api.Person, len(persons))
	for i, person := range persons {
		apiPersons[i] = api.Person{
			ID:        person.ID,
			Name:      person.Name,
			CreatedAt: person.CreatedAt,
			UpdatedAt: person.UpdatedAt,
		}
	}

	return &api.ListPersonsOK{
		Persons: apiPersons,
		Total:   int(total),
	}, nil
}

// UpdatePerson implements the UpdatePerson operation
func (h *APIHandler) UpdatePerson(ctx context.Context, req *api.UpdatePersonRequest, params api.UpdatePersonParams) (api.UpdatePersonRes, error) {
	// Validate request
	if req.Name == "" {
		return &api.UpdatePersonBadRequest{
			Message: "Name is required",
			Code:    "VALIDATION_ERROR",
		}, nil
	}

	person, err := h.queries.UpdatePerson(ctx, db.UpdatePersonParams{
		XidStr: params.ID,
		Name:   req.Name,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return &api.UpdatePersonNotFound{
				Message: "Person not found",
				Code:    "NOT_FOUND",
			}, nil
		}
		log.Printf("Error updating person: %v", err)
		return &api.UpdatePersonInternalServerError{
			Message: "Failed to update person",
			Code:    "INTERNAL_ERROR",
		}, nil
	}

	return &api.Person{
		ID:        person.ID,
		Name:      person.Name,
		CreatedAt: person.CreatedAt,
		UpdatedAt: person.UpdatedAt,
	}, nil
}

// DeletePerson implements the DeletePerson operation
func (h *APIHandler) DeletePerson(ctx context.Context, params api.DeletePersonParams) (api.DeletePersonRes, error) {
	err := h.queries.DeletePerson(ctx, params.ID)
	if err != nil {
		log.Printf("Error deleting person: %v", err)
		return &api.DeletePersonInternalServerError{
			Message: "Failed to delete person",
			Code:    "INTERNAL_ERROR",
		}, nil
	}

	return &api.DeletePersonNoContent{}, nil
}

// Form handlers for HTMX integration
func (h *APIHandler) handleCreatePersonForm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := r.FormValue("name")
	if name == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`<div class="text-red-500">Name is required</div>`))
		return
	}

	// Call the API handler internally
	req := &api.CreatePersonRequest{Name: name}
	result, err := h.CreatePerson(r.Context(), req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`<div class="text-red-500">Failed to create person</div>`))
		return
	}

	// Check if result is a person or error
	switch person := result.(type) {
	case *api.Person:
		// Return HTML fragment for the new person
		html := fmt.Sprintf(`
			<div class="border-b pb-2 mb-2" id="person-%s">
				<div class="flex justify-between items-center">
					<div>
						<span class="font-medium">%s</span>
						<span class="text-sm text-gray-500 ml-2">ID: %s</span>
					</div>
					<div class="space-x-2">
						<button hx-get="/forms/persons/%s/edit" hx-target="#person-%s" hx-swap="outerHTML"
								class="text-blue-500 hover:text-blue-700 text-sm">Edit</button>
						<button hx-delete="/forms/persons/delete/%s" hx-target="#person-%s" hx-swap="outerHTML"
								hx-confirm="Are you sure you want to delete this person?"
								class="text-red-500 hover:text-red-700 text-sm">Delete</button>
					</div>
				</div>
				<div class="text-xs text-gray-400 mt-1">
					Created: %s
				</div>
			</div>`,
			person.ID, person.Name, person.ID, person.ID, person.ID, person.ID, person.ID,
			person.CreatedAt.Format("2006-01-02 15:04:05"))
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	default:
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`<div class="text-red-500">Failed to create person</div>`))
	}
}

func (h *APIHandler) handleListPersonsHTML(w http.ResponseWriter, r *http.Request) {
	// Call the API handler internally
	params := api.ListPersonsParams{
		Limit:  api.OptInt{Value: 50, Set: true},
		Offset: api.OptInt{Value: 0, Set: true},
	}

	result, err := h.ListPersons(r.Context(), params)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`<div class="text-red-500">Failed to load persons</div>`))
		return
	}

	switch listResult := result.(type) {
	case *api.ListPersonsOK:
		if len(listResult.Persons) == 0 {
			w.Write([]byte(`<div class="text-gray-500 text-center py-4">No people found. Add someone above!</div>`))
			return
		}

		var html strings.Builder
		for _, person := range listResult.Persons {
			html.WriteString(fmt.Sprintf(`
				<div class="border-b pb-2 mb-2" id="person-%s">
					<div class="flex justify-between items-center">
						<div>
							<span class="font-medium">%s</span>
							<span class="text-sm text-gray-500 ml-2">ID: %s</span>
						</div>
						<div class="space-x-2">
							<button hx-get="/forms/persons/%s/edit" hx-target="#person-%s" hx-swap="outerHTML"
									class="text-blue-500 hover:text-blue-700 text-sm">Edit</button>
							<button hx-delete="/forms/persons/delete/%s" hx-target="#person-%s" hx-swap="outerHTML"
									hx-confirm="Are you sure you want to delete this person?"
									class="text-red-500 hover:text-red-700 text-sm">Delete</button>
						</div>
					</div>
					<div class="text-xs text-gray-400 mt-1">
						Created: %s | Updated: %s
					</div>
				</div>`,
				person.ID, person.Name, person.ID, person.ID, person.ID, person.ID, person.ID,
				person.CreatedAt.Format("2006-01-02 15:04:05"),
				person.UpdatedAt.Format("2006-01-02 15:04:05")))
		}

		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html.String()))
	default:
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`<div class="text-red-500">Failed to load persons</div>`))
	}
}

func (h *APIHandler) handleDeletePersonForm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract person ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`<div class="text-red-500">Invalid person ID</div>`))
		return
	}
	personID := pathParts[4] // /forms/persons/delete/{id}

	// Call the API handler internally
	params := api.DeletePersonParams{ID: personID}
	_, err := h.DeletePerson(r.Context(), params)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`<div class="text-red-500">Failed to delete person</div>`))
		return
	}

	// Return empty content to remove the element
	w.WriteHeader(http.StatusOK)
}

// setupRoutes sets up HTTP routes
func setupRoutes(apiServer *api.Server, formHandler *APIHandler) *http.ServeMux {
	mux := http.NewServeMux()

	// Serve static files
	fs := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`))
	})

	// Form handlers for HTMX
	mux.HandleFunc("/forms/persons/create", formHandler.handleCreatePersonForm)
	mux.HandleFunc("/forms/persons/list", formHandler.handleListPersonsHTML)
	mux.HandleFunc("/forms/persons/delete/", formHandler.handleDeletePersonForm)

	// API routes (mount the ogen server)
	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", apiServer))

	// Root endpoint - serve the main HTML page
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Performance Tracking</title>
    <script src="https://unpkg.com/htmx.org@1.9.8"></script>
    <script src="https://unpkg.com/htmx.org@1.9.8/dist/ext/json-enc.js"></script>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <h1 class="text-3xl font-bold text-gray-900 mb-8">Performance Tracking</h1>

        <!-- Add Person Form -->
        <div class="bg-white rounded-lg shadow p-6 mb-6">
            <h2 class="text-xl font-semibold mb-4">Add New Person</h2>
            <form hx-post="/forms/persons/create" hx-target="#persons-list" hx-swap="afterbegin">
                <div class="flex gap-4">
                    <input
                        type="text"
                        name="name"
                        placeholder="Person's name"
                        required
                        class="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    >
                    <button
                        type="submit"
                        class="px-4 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500"
                    >
                        Add Person
                    </button>
                </div>
            </form>
        </div>

        <!-- Persons List -->
        <div class="bg-white rounded-lg shadow p-6">
            <h2 class="text-xl font-semibold mb-4">People</h2>
            <div id="persons-list" hx-get="/forms/persons/list" hx-trigger="load">
                <p class="text-gray-500">Loading...</p>
            </div>
        </div>

        <!-- API Info -->
        <div class="mt-8 bg-blue-50 rounded-lg p-4">
            <h3 class="text-lg font-semibold text-blue-800 mb-2">API Endpoints</h3>
            <ul class="text-sm text-blue-600 space-y-1">
                <li><code>GET /api/v1/persons</code> - List all persons</li>
                <li><code>POST /api/v1/persons</code> - Create a new person</li>
                <li><code>GET /api/v1/persons/{id}</code> - Get person by ID</li>
                <li><code>PUT /api/v1/persons/{id}</code> - Update person</li>
                <li><code>DELETE /api/v1/persons/{id}</code> - Delete person</li>
            </ul>
        </div>
    </div>
</body>
</html>
		`))
	})

	// Demo endpoint to test xid generation
	mux.HandleFunc("/api/v1/demo/xid", func(w http.ResponseWriter, r *http.Request) {
		id := xid.New()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf(`{"id":"%s","timestamp":"%s"}`, id.String(), id.Time().Format(time.RFC3339))))
	})

	return mux
}
