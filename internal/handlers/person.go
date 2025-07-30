package handlers

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"strings"

	"github.com/rs/xid"

	"pepo/internal/api"
	"pepo/internal/db"
	"pepo/templates"
)

type PersonHandler struct {
	queries *db.Queries
}

func NewPersonHandler(queries *db.Queries) *PersonHandler {
	return &PersonHandler{
		queries: queries,
	}
}

// API Handlers

func (h *PersonHandler) CreatePerson(ctx context.Context, req *api.CreatePersonRequest) (api.CreatePersonRes, error) {
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

func (h *PersonHandler) GetPerson(ctx context.Context, params api.GetPersonParams) (api.GetPersonRes, error) {
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

func (h *PersonHandler) ListPersons(ctx context.Context, params api.ListPersonsParams) (api.ListPersonsRes, error) {
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

func (h *PersonHandler) UpdatePerson(ctx context.Context, req *api.UpdatePersonRequest, params api.UpdatePersonParams) (api.UpdatePersonRes, error) {
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

func (h *PersonHandler) DeletePerson(ctx context.Context, params api.DeletePersonParams) (api.DeletePersonRes, error) {
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

// Form Handlers

func (h *PersonHandler) HandleCreatePersonForm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := r.FormValue("name")
	if name == "" {
		w.WriteHeader(http.StatusBadRequest)
		templates.ActionError("Name is required").Render(r.Context(), w)
		return
	}

	// Call the API handler internally
	req := &api.CreatePersonRequest{Name: name}
	result, err := h.CreatePerson(r.Context(), req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		templates.ActionError("Failed to create person").Render(r.Context(), w)
		return
	}

	// Check if result is a person or error
	switch person := result.(type) {
	case *api.Person:
		// Convert to template person
		templatePerson := templates.Person{
			ID:        person.ID,
			Name:      person.Name,
			CreatedAt: person.CreatedAt,
			UpdatedAt: person.UpdatedAt,
		}
		w.Header().Set("Content-Type", "text/html")
		templates.PersonItem(templatePerson).Render(r.Context(), w)
	default:
		w.WriteHeader(http.StatusBadRequest)
		templates.ActionError("Failed to create person").Render(r.Context(), w)
	}
}

func (h *PersonHandler) HandleListPersonsHTML(w http.ResponseWriter, r *http.Request) {
	// Call the API handler internally
	params := api.ListPersonsParams{
		Limit:  api.OptInt{Value: 50, Set: true},
		Offset: api.OptInt{Value: 0, Set: true},
	}

	result, err := h.ListPersons(r.Context(), params)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		templates.ActionError("Failed to load persons").Render(r.Context(), w)
		return
	}

	switch listResult := result.(type) {
	case *api.ListPersonsOK:
		// Convert to template persons
		templatePersons := make([]templates.Person, len(listResult.Persons))
		for i, person := range listResult.Persons {
			templatePersons[i] = templates.Person{
				ID:        person.ID,
				Name:      person.Name,
				CreatedAt: person.CreatedAt,
				UpdatedAt: person.UpdatedAt,
			}
		}

		w.Header().Set("Content-Type", "text/html")
		templates.PersonList(templatePersons).Render(r.Context(), w)
	default:
		w.WriteHeader(http.StatusInternalServerError)
		templates.ActionError("Failed to load persons").Render(r.Context(), w)
	}
}

func (h *PersonHandler) HandleDeletePersonForm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract person ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		w.WriteHeader(http.StatusBadRequest)
		templates.ActionError("Invalid person ID").Render(r.Context(), w)
		return
	}
	personID := pathParts[4] // /forms/persons/delete/{id}

	// Call the API handler internally
	params := api.DeletePersonParams{ID: personID}
	_, err := h.DeletePerson(r.Context(), params)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		templates.ActionError("Failed to delete person").Render(r.Context(), w)
		return
	}

	// Return empty content to remove the element
	w.WriteHeader(http.StatusOK)
}

func (h *PersonHandler) HandleGetPersonsForSelect(w http.ResponseWriter, r *http.Request) {
	// Call the API handler internally to get persons for the select dropdown
	params := api.ListPersonsParams{
		Limit:  api.OptInt{Value: 100, Set: true},
		Offset: api.OptInt{Value: 0, Set: true},
	}

	result, err := h.ListPersons(r.Context(), params)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		templates.PersonSelectError().Render(r.Context(), w)
		return
	}

	switch listResult := result.(type) {
	case *api.ListPersonsOK:
		// Convert to template persons
		templatePersons := make([]templates.Person, len(listResult.Persons))
		for i, person := range listResult.Persons {
			templatePersons[i] = templates.Person{
				ID:        person.ID,
				Name:      person.Name,
				CreatedAt: person.CreatedAt,
				UpdatedAt: person.UpdatedAt,
			}
		}

		w.Header().Set("Content-Type", "text/html")
		templates.PersonSelectOptions(templatePersons).Render(r.Context(), w)
	default:
		w.WriteHeader(http.StatusInternalServerError)
		templates.PersonSelectError().Render(r.Context(), w)
	}
}
