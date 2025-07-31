package handlers

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"

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

func (h *PersonHandler) GetPersonById(ctx context.Context, params api.GetPersonByIdParams) (api.GetPersonByIdRes, error) {
	person, err := h.queries.GetPersonByID(ctx, params.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return &api.GetPersonByIdNotFound{
				Message: "Person not found",
				Code:    "NOT_FOUND",
			}, nil
		}
		log.Printf("Error getting person: %v", err)
		return &api.GetPersonByIdInternalServerError{
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

func (h *PersonHandler) GetPersons(ctx context.Context, params api.GetPersonsParams) (api.GetPersonsRes, error) {
	limit := int32(10)
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
			Message: "Failed to count people",
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
			Message: "Failed to list people",
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

	return &api.GetPersonsOKApplicationJSON{
		Persons: apiPersons,
		Total:   int(total),
	}, nil
}

func (h *PersonHandler) GetPersonsWithLastAction(ctx context.Context, params api.GetPersonsParams) ([]templates.PersonWithLastAction, error) {
	limit := int32(10)
	if params.Limit.IsSet() {
		limit = int32(params.Limit.Value)
	}

	offset := int32(0)
	if params.Offset.IsSet() {
		offset = int32(params.Offset.Value)
	}

	// Get persons with last action data
	persons, err := h.queries.ListPersonsWithLastAction(ctx, db.ListPersonsWithLastActionParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		log.Printf("Error listing persons with last action: %v", err)
		return nil, err
	}

	// Convert to template PersonWithLastAction
	templatePersons := make([]templates.PersonWithLastAction, len(persons))
	for i, person := range persons {
		templatePerson := templates.PersonWithLastAction{
			ID:        person.ID,
			Name:      person.Name,
			CreatedAt: person.CreatedAt,
			UpdatedAt: person.UpdatedAt,
		}

		// Handle the last_action_at which might be nil or different types
		if person.LastActionAt != nil {
			switch v := person.LastActionAt.(type) {
			case time.Time:
				if !v.IsZero() {
					templatePerson.LastActionAt = &v
				}
			case *time.Time:
				if v != nil && !v.IsZero() {
					templatePerson.LastActionAt = v
				}
			}
		}

		templatePersons[i] = templatePerson
	}

	return templatePersons, nil
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

func (h *PersonHandler) HandleGetPersonsForSelect(w http.ResponseWriter, r *http.Request) {
	// Call the API handler internally to get persons for the select dropdown
	params := api.GetPersonsParams{
		Limit:  api.OptInt{Value: 100, Set: true},
		Offset: api.OptInt{Value: 0, Set: true},
	}

	result, err := h.GetPersons(r.Context(), params)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		templates.PersonSelectError().Render(r.Context(), w)
		return
	}

	switch listResult := result.(type) {
	case *api.GetPersonsOKApplicationJSON:
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
