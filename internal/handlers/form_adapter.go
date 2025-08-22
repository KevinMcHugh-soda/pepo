package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"pepo/internal/api"
)

// FormAdapter handles conversion between HTML form data and API request structures
type FormAdapter struct{}

// NewFormAdapter creates a new form adapter
func NewFormAdapter() *FormAdapter {
	return &FormAdapter{}
}

// ParseCreatePersonRequest converts form data to CreatePersonRequest
func (f *FormAdapter) ParseCreatePersonRequest(r *http.Request) (*api.CreatePersonRequest, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}

	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		return nil, &ValidationError{Field: "name", Message: "Name is required"}
	}

	return &api.CreatePersonRequest{
		Name: name,
	}, nil
}

// ParseCreateActionRequest converts form data to CreateActionRequest
func (f *FormAdapter) ParseCreateActionRequest(r *http.Request) (*api.CreateActionRequest, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}

	// Required fields
	personID := strings.TrimSpace(r.FormValue("person_id"))
	if personID == "" {
		return nil, &ValidationError{Field: "person_id", Message: "Person ID is required"}
	}

	description := strings.TrimSpace(r.FormValue("description"))
	if description == "" {
		return nil, &ValidationError{Field: "description", Message: "Description is required"}
	}

	valenceStr := strings.TrimSpace(r.FormValue("valence"))
	if valenceStr == "" {
		return nil, &ValidationError{Field: "valence", Message: "Valence is required"}
	}

	// Validate valence
	var valence api.CreateActionRequestValence
	switch valenceStr {
	case "positive":
		valence = api.CreateActionRequestValencePositive
	case "negative":
		valence = api.CreateActionRequestValenceNegative
	case "neutral":
		valence = api.CreateActionRequestValenceNeutral
	default:
		return nil, &ValidationError{Field: "valence", Message: "Invalid valence. Must be positive, negative, or neutral"}
	}

	// Parse occurred_at (optional, defaults to current time if not provided)
	var occurredAt time.Time
	occurredAtStr := strings.TrimSpace(r.FormValue("occurred_at"))
	if occurredAtStr != "" {
		// Try parsing different datetime formats
		var err error
		// HTML datetime-local format: 2006-01-02T15:04
		occurredAt, err = time.Parse("2006-01-02T15:04", occurredAtStr)
		if err != nil {
			// Try with seconds: 2006-01-02T15:04:05
			occurredAt, err = time.Parse("2006-01-02T15:04:05", occurredAtStr)
			if err != nil {
				// Try ISO format: 2006-01-02T15:04:05Z
				occurredAt, err = time.Parse(time.RFC3339, occurredAtStr)
				if err != nil {
					return nil, &ValidationError{Field: "occurred_at", Message: "Invalid date format. Use YYYY-MM-DDTHH:MM format"}
				}
			}
		}
	} else {
		occurredAt = time.Now()
	}

	// Optional references field
	references := strings.TrimSpace(r.FormValue("references"))

	req := &api.CreateActionRequest{
		PersonID:    personID,
		OccurredAt:  occurredAt,
		Description: description,
		Valence:     valence,
	}

	if references != "" {
		req.References = api.OptNilString{Value: references, Set: true}
	}

	// Parse existing theme IDs
	if themes := r.Form["themes"]; len(themes) > 0 {
		req.Themes = make([]string, 0, len(themes))
		for _, t := range themes {
			t = strings.TrimSpace(t)
			if t != "" {
				req.Themes = append(req.Themes, t)
			}
		}
	}

	return req, nil
}

// ParseUpdatePersonRequest converts form data to UpdatePersonRequest
func (f *FormAdapter) ParseUpdatePersonRequest(r *http.Request) (*api.UpdatePersonRequest, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}

	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		return nil, &ValidationError{Field: "name", Message: "Name is required"}
	}

	return &api.UpdatePersonRequest{
		Name: name,
	}, nil
}

// ParseUpdateActionRequest converts form data to UpdateActionRequest
func (f *FormAdapter) ParseUpdateActionRequest(r *http.Request) (*api.UpdateActionRequest, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}

	// Required fields
	personID := strings.TrimSpace(r.FormValue("person_id"))
	if personID == "" {
		return nil, &ValidationError{Field: "person_id", Message: "Person ID is required"}
	}

	description := strings.TrimSpace(r.FormValue("description"))
	if description == "" {
		return nil, &ValidationError{Field: "description", Message: "Description is required"}
	}

	valenceStr := strings.TrimSpace(r.FormValue("valence"))
	if valenceStr == "" {
		return nil, &ValidationError{Field: "valence", Message: "Valence is required"}
	}

	// Validate valence
	var valence api.UpdateActionRequestValence
	switch valenceStr {
	case "positive":
		valence = api.UpdateActionRequestValencePositive
	case "negative":
		valence = api.UpdateActionRequestValenceNegative
	default:
		return nil, &ValidationError{Field: "valence", Message: "Invalid valence. Must be positive, negative, or neutral"}
	}

	// Parse occurred_at (required for updates)
	occurredAtStr := strings.TrimSpace(r.FormValue("occurred_at"))
	if occurredAtStr == "" {
		return nil, &ValidationError{Field: "occurred_at", Message: "Occurred at time is required"}
	}

	var occurredAt time.Time
	var err error
	// Try parsing different datetime formats
	occurredAt, err = time.Parse("2006-01-02T15:04", occurredAtStr)
	if err != nil {
		occurredAt, err = time.Parse("2006-01-02T15:04:05", occurredAtStr)
		if err != nil {
			occurredAt, err = time.Parse(time.RFC3339, occurredAtStr)
			if err != nil {
				return nil, &ValidationError{Field: "occurred_at", Message: "Invalid date format. Use YYYY-MM-DDTHH:MM format"}
			}
		}
	}

	// Optional references field
	references := strings.TrimSpace(r.FormValue("references"))

	req := &api.UpdateActionRequest{
		PersonID:    personID,
		OccurredAt:  occurredAt,
		Description: description,
		Valence:     valence,
	}

	if references != "" {
		req.References = api.OptNilString{Value: references, Set: true}
	}

	return req, nil
}

// IsFormData checks if the request contains form data
func (f *FormAdapter) IsFormData(r *http.Request) bool {
	contentType := r.Header.Get("Content-Type")
	return strings.HasPrefix(contentType, "application/x-www-form-urlencoded") ||
		strings.HasPrefix(contentType, "multipart/form-data")
}

// IsHTMLRequest checks if the client prefers HTML response
func (f *FormAdapter) IsHTMLRequest(r *http.Request) bool {
	accept := r.Header.Get("Accept")
	return strings.Contains(accept, "text/html")
}

// ValidationError represents a form validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// ParseQueryParams extracts common query parameters
func (f *FormAdapter) ParseQueryParams(r *http.Request) QueryParams {
	query := r.URL.Query()

	var limit int = 10 // default
	if limitStr := query.Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	var offset int = 0 // default
	if offsetStr := query.Get("offset"); offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	return QueryParams{
		Limit:    limit,
		Offset:   offset,
		PersonID: query.Get("person_id"),
		Valence:  query.Get("valence"),
		Format:   query.Get("format"),
	}
}

// QueryParams represents common query parameters
type QueryParams struct {
	Limit    int
	Offset   int
	PersonID string
	Valence  string
	Format   string
}

// ToGetPersonsParams converts to API parameters
func (q QueryParams) ToGetPersonsParams() api.GetPersonsParams {
	params := api.GetPersonsParams{}

	if q.Limit > 0 {
		params.Limit = api.OptInt{Value: q.Limit, Set: true}
	}

	if q.Offset > 0 {
		params.Offset = api.OptInt{Value: q.Offset, Set: true}
	}

	return params
}

// ToGetActionsParams converts to API parameters
func (q QueryParams) ToGetActionsParams() api.GetActionsParams {
	params := api.GetActionsParams{}

	if q.Limit > 0 {
		params.Limit = api.OptInt{Value: q.Limit, Set: true}
	}

	if q.Offset > 0 {
		params.Offset = api.OptInt{Value: q.Offset, Set: true}
	}

	if q.PersonID != "" {
		params.PersonID = api.OptString{Value: q.PersonID, Set: true}
	}

	if q.Valence != "" {
		switch q.Valence {
		case "positive":
			params.Valence = api.OptGetActionsValence{Value: api.GetActionsValencePositive, Set: true}
		case "negative":
			params.Valence = api.OptGetActionsValence{Value: api.GetActionsValenceNegative, Set: true}
		}
	}

	return params
}

// ToGetPersonActionsParams converts to API parameters
func (q QueryParams) ToGetPersonActionsParams(personID string) api.GetPersonActionsParams {
	params := api.GetPersonActionsParams{
		ID: personID,
	}

	if q.Limit > 0 {
		params.Limit = api.OptInt{Value: q.Limit, Set: true}
	}

	if q.Offset > 0 {
		params.Offset = api.OptInt{Value: q.Offset, Set: true}
	}

	if q.Valence != "" {
		switch q.Valence {
		case "positive":
			params.Valence = api.OptGetPersonActionsValence{Value: api.GetPersonActionsValencePositive, Set: true}
		case "negative":
			params.Valence = api.OptGetPersonActionsValence{Value: api.GetPersonActionsValenceNegative, Set: true}
		}
	}

	return params
}
