package handlers

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"

	"github.com/a-h/templ"

	"pepo/internal/api"
	"pepo/internal/middleware"
	"pepo/templates"
)

// ContentNegotiatingHandler wraps the existing handlers with content negotiation
type ContentNegotiatingHandler struct {
	combinedHandler *CombinedAPIHandler
}

// NewContentNegotiatingHandler creates a new content negotiating handler
func NewContentNegotiatingHandler(combinedHandler *CombinedAPIHandler) *ContentNegotiatingHandler {
	return &ContentNegotiatingHandler{
		combinedHandler: combinedHandler,
	}
}

// determineResponseType checks Accept header and returns preferred content type
func (h *ContentNegotiatingHandler) determineResponseType(r *http.Request) string {
	accept := r.Header.Get("Accept")
	// Check for explicit HTML preference
	if strings.Contains(accept, "text/html") {
		return "text/html"
	}
	// Check for HTMX requests (they typically want HTML)
	if r.Header.Get("HX-Request") == "true" {
		return "text/html"
	}
	return "application/json"
}

// getRequestFromContext safely extracts HTTP request from context
func (h *ContentNegotiatingHandler) getRequestFromContext(ctx context.Context) *http.Request {
	if req, ok := ctx.Value(middleware.HTTPRequestKey).(*http.Request); ok {
		return req
	}
	// Fallback: try other context keys
	if req, ok := ctx.Value("http_request").(*http.Request); ok {
		return req
	}
	return nil
}

// renderTemplate renders a template to an io.Reader
func renderTemplate(component templ.Component) io.Reader {
	var buf bytes.Buffer
	component.Render(context.Background(), &buf)
	return &buf
}

// GetPersons handles both JSON and HTML requests for listing people
func (h *ContentNegotiatingHandler) GetPersons(ctx context.Context, params api.GetPersonsParams) (api.GetPersonsRes, error) {
	// Check if we have request context to determine response type
	if req := h.getRequestFromContext(ctx); req != nil {
		responseType := h.determineResponseType(req)

		if responseType == "text/html" {
			// Check if this is a request for select options format
			if req.URL.Query().Get("format") == "select" {
				// For select options, we still need the basic person data
				result, err := h.combinedHandler.GetPersons(ctx, params)
				if err != nil {
					return result, err
				}

				switch jsonResult := result.(type) {
				case *api.GetPersonsOKApplicationJSON:
					// Convert API persons to template people
					templatePersons := make([]templates.Person, len(jsonResult.Persons))
					for i, person := range jsonResult.Persons {
						templatePersons[i] = templates.Person{
							ID:        person.ID,
							Name:      person.Name,
							CreatedAt: person.CreatedAt,
							UpdatedAt: person.UpdatedAt,
						}
					}
					// Render select options template
					return &api.GetPersonsOKTextHTML{
						Data: renderTemplate(templates.PersonSelectOptions(templatePersons)),
					}, nil
				}
			} else {
				// For regular list, get persons with last action data
				templatePersons, err := h.combinedHandler.GetPersonsWithLastAction(ctx, params)
				if err != nil {
					return &api.Error{
						Message: "Failed to get people with last actions",
						Code:    "INTERNAL_ERROR",
					}, nil
				}

				// Render person list with last action template
				return &api.GetPersonsOKTextHTML{
					Data: renderTemplate(templates.PersonWithLastActionList(templatePersons)),
				}, nil
			}
		}
	}

	// Call the business logic for JSON response
	result, err := h.combinedHandler.GetPersons(ctx, params)
	if err != nil {
		return result, err
	}

	// Return JSON response (default)
	return result, nil
}

// GetPersonById handles both JSON and HTML requests for getting a person by ID
func (h *ContentNegotiatingHandler) GetPersonById(ctx context.Context, params api.GetPersonByIdParams) (api.GetPersonByIdRes, error) {
	// Call the business logic
	result, err := h.combinedHandler.GetPersonById(ctx, params)
	if err != nil {
		return result, err
	}

	// Check if we have request context to determine response type
	if req := h.getRequestFromContext(ctx); req != nil {
		responseType := h.determineResponseType(req)

		if responseType == "text/html" {
			// Handle HTML response
			switch jsonResult := result.(type) {
			case *api.Person:
				// Convert API person to template person
				templatePerson := templates.Person{
					ID:        jsonResult.ID,
					Name:      jsonResult.Name,
					CreatedAt: jsonResult.CreatedAt,
					UpdatedAt: jsonResult.UpdatedAt,
				}

				// Fetch the person's actions for the detail view
				actionsParams := api.GetPersonActionsParams{
					ID:    params.ID,
					Limit: api.OptInt{Value: 100, Set: true}, // Get more actions for detail view
				}

				actionsResult, err := h.combinedHandler.GetPersonActions(ctx, actionsParams)
				if err != nil {
					// If we can't get actions, still show the person with empty actions
					return &api.GetPersonByIdOKTextHTML{
						Data: renderTemplate(templates.PersonDetail(templatePerson, []templates.Action{})),
					}, nil
				}

				var templateActions []templates.Action
				if actionsJSON, ok := actionsResult.(*api.GetPersonActionsOKApplicationJSON); ok {
					templateActions = make([]templates.Action, len(actionsJSON.Actions))
					for i, action := range actionsJSON.Actions {
						templateActions[i] = templates.Action{
							ID:          action.ID,
							PersonID:    action.PersonID,
							OccurredAt:  action.OccurredAt,
							Description: action.Description,
							References:  action.References.Or(""),
							Valence:     string(action.Valence),
							CreatedAt:   action.CreatedAt,
							UpdatedAt:   action.UpdatedAt,
						}
					}
				}

				// Render PersonDetail template with person and actions
				return &api.GetPersonByIdOKTextHTML{
					Data: renderTemplate(templates.PersonDetail(templatePerson, templateActions)),
				}, nil
			}
		}
	}

	// Return JSON response (default)
	return result, nil
}

// CreatePerson handles both JSON and HTML requests for creating a person
func (h *ContentNegotiatingHandler) CreatePerson(ctx context.Context, req *api.CreatePersonRequest) (api.CreatePersonRes, error) {
	// Call the business logic
	result, err := h.combinedHandler.CreatePerson(ctx, req)
	if err != nil {
		return result, err
	}

	// Check if we have request context to determine response type
	if httpReq := h.getRequestFromContext(ctx); httpReq != nil {
		responseType := h.determineResponseType(httpReq)

		if responseType == "text/html" {
			// Handle HTML response
			switch jsonResult := result.(type) {
			case *api.Person:
				// Convert API person to template person
				templatePerson := templates.Person{
					ID:        jsonResult.ID,
					Name:      jsonResult.Name,
					CreatedAt: jsonResult.CreatedAt,
					UpdatedAt: jsonResult.UpdatedAt,
				}

				// Render template and return HTML response
				return &api.CreatePersonCreatedTextHTML{
					Data: renderTemplate(templates.PersonItem(templatePerson)),
				}, nil
			}
		}
	}

	// Return JSON response (default)
	return result, nil
}

// UpdatePerson handles both JSON and HTML requests for updating a person
func (h *ContentNegotiatingHandler) UpdatePerson(ctx context.Context, req *api.UpdatePersonRequest, params api.UpdatePersonParams) (api.UpdatePersonRes, error) {
	// Call the business logic
	result, err := h.combinedHandler.UpdatePerson(ctx, req, params)
	if err != nil {
		return result, err
	}

	// Check if we have request context to determine response type
	if httpReq := h.getRequestFromContext(ctx); httpReq != nil {
		responseType := h.determineResponseType(httpReq)

		if responseType == "text/html" {
			// Handle HTML response
			switch jsonResult := result.(type) {
			case *api.Person:
				// Convert API person to template person
				templatePerson := templates.Person{
					ID:        jsonResult.ID,
					Name:      jsonResult.Name,
					CreatedAt: jsonResult.CreatedAt,
					UpdatedAt: jsonResult.UpdatedAt,
				}

				// Render template and return HTML response
				return &api.UpdatePersonOKTextHTML{
					Data: renderTemplate(templates.PersonItem(templatePerson)),
				}, nil
			}
		}
	}

	// Return JSON response (default)
	return result, nil
}

// DeletePerson handles delete requests (no content negotiation needed for 204 responses)
func (h *ContentNegotiatingHandler) DeletePerson(ctx context.Context, params api.DeletePersonParams) (api.DeletePersonRes, error) {
	return h.combinedHandler.DeletePerson(ctx, params)
}

// GetActions handles both JSON and HTML requests for listing actions
func (h *ContentNegotiatingHandler) GetActions(ctx context.Context, params api.GetActionsParams) (api.GetActionsRes, error) {
	// Call the business logic
	result, err := h.combinedHandler.GetActions(ctx, params)
	if err != nil {
		return result, err
	}

	// Check if we have request context to determine response type
	if req := h.getRequestFromContext(ctx); req != nil {
		responseType := h.determineResponseType(req)

		if responseType == "text/html" {
			// Handle HTML response
			switch jsonResult := result.(type) {
			case *api.GetActionsOKApplicationJSON:
				// Convert API actions to template actions
				templateActions := make([]templates.Action, len(jsonResult.Actions))
				for i, action := range jsonResult.Actions {
					templateActions[i] = templates.Action{
						ID:          action.ID,
						PersonID:    action.PersonID,
						OccurredAt:  action.OccurredAt,
						Description: action.Description,
						References:  action.References.Or(""),
						Valence:     string(action.Valence),
						CreatedAt:   action.CreatedAt,
						UpdatedAt:   action.UpdatedAt,
					}
				}

				// Render template and return HTML response
				return &api.GetActionsOKTextHTML{
					Data: renderTemplate(templates.ActionList(templateActions)),
				}, nil
			}
		}
	}

	// Return JSON response (default)
	return result, nil
}

// GetActionById handles both JSON and HTML requests for getting an action by ID
func (h *ContentNegotiatingHandler) GetActionById(ctx context.Context, params api.GetActionByIdParams) (api.GetActionByIdRes, error) {
	// Call the business logic
	result, err := h.combinedHandler.GetActionById(ctx, params)
	if err != nil {
		return result, err
	}

	// Check if we have request context to determine response type
	if req := h.getRequestFromContext(ctx); req != nil {
		responseType := h.determineResponseType(req)

		if responseType == "text/html" {
			// Handle HTML response
			switch jsonResult := result.(type) {
			case *api.Action:
				// Convert API action to template action
				templateAction := templates.Action{
					ID:          jsonResult.ID,
					PersonID:    jsonResult.PersonID,
					OccurredAt:  jsonResult.OccurredAt,
					Description: jsonResult.Description,
					References:  jsonResult.References.Or(""),
					Valence:     string(jsonResult.Valence),
					CreatedAt:   jsonResult.CreatedAt,
					UpdatedAt:   jsonResult.UpdatedAt,
				}

				// Render template and return HTML response
				return &api.GetActionByIdOKTextHTML{
					Data: renderTemplate(templates.ActionItem(templateAction)),
				}, nil
			}
		}
	}

	// Return JSON response (default)
	return result, nil
}

// CreateAction handles both JSON and HTML requests for creating an action
func (h *ContentNegotiatingHandler) CreateAction(ctx context.Context, req *api.CreateActionRequest) (api.CreateActionRes, error) {
	// Call the business logic
	result, err := h.combinedHandler.CreateAction(ctx, req)
	if err != nil {
		return result, err
	}

	// Check if we have request context to determine response type
	if httpReq := h.getRequestFromContext(ctx); httpReq != nil {
		responseType := h.determineResponseType(httpReq)

		if responseType == "text/html" {
			// Handle HTML response
			switch jsonResult := result.(type) {
			case *api.Action:
				// Convert API action to template action
				templateAction := templates.Action{
					ID:          jsonResult.ID,
					PersonID:    jsonResult.PersonID,
					OccurredAt:  jsonResult.OccurredAt,
					Description: jsonResult.Description,
					References:  jsonResult.References.Or(""),
					Valence:     string(jsonResult.Valence),
					CreatedAt:   jsonResult.CreatedAt,
					UpdatedAt:   jsonResult.UpdatedAt,
				}

				// Render template and return HTML response
				return &api.CreateActionCreatedTextHTML{
					Data: renderTemplate(templates.ActionItem(templateAction)),
				}, nil
			}
		}
	}

	// Return JSON response (default)
	return result, nil
}

// UpdateAction handles both JSON and HTML requests for updating an action
func (h *ContentNegotiatingHandler) UpdateAction(ctx context.Context, req *api.UpdateActionRequest, params api.UpdateActionParams) (api.UpdateActionRes, error) {
	// Call the business logic
	result, err := h.combinedHandler.UpdateAction(ctx, req, params)
	if err != nil {
		return result, err
	}

	// Check if we have request context to determine response type
	if httpReq := h.getRequestFromContext(ctx); httpReq != nil {
		responseType := h.determineResponseType(httpReq)

		if responseType == "text/html" {
			// Handle HTML response
			switch jsonResult := result.(type) {
			case *api.Action:
				// Convert API action to template action
				templateAction := templates.Action{
					ID:          jsonResult.ID,
					PersonID:    jsonResult.PersonID,
					OccurredAt:  jsonResult.OccurredAt,
					Description: jsonResult.Description,
					References:  jsonResult.References.Or(""),
					Valence:     string(jsonResult.Valence),
					CreatedAt:   jsonResult.CreatedAt,
					UpdatedAt:   jsonResult.UpdatedAt,
				}

				// Render template and return HTML response
				return &api.UpdateActionOKTextHTML{
					Data: renderTemplate(templates.ActionItem(templateAction)),
				}, nil
			}
		}
	}

	// Return JSON response (default)
	return result, nil
}

// DeleteAction handles delete requests (no content negotiation needed for 204 responses)
func (h *ContentNegotiatingHandler) DeleteAction(ctx context.Context, params api.DeleteActionParams) (api.DeleteActionRes, error) {
	return h.combinedHandler.DeleteAction(ctx, params)
}

// GetPersonActions handles both JSON and HTML requests for getting a person's actions
func (h *ContentNegotiatingHandler) GetPersonActions(ctx context.Context, params api.GetPersonActionsParams) (api.GetPersonActionsRes, error) {
	// Call the business logic
	result, err := h.combinedHandler.GetPersonActions(ctx, params)
	if err != nil {
		return result, err
	}

	// Check if we have request context to determine response type
	if req := h.getRequestFromContext(ctx); req != nil {
		responseType := h.determineResponseType(req)

		if responseType == "text/html" {
			// Handle HTML response
			switch jsonResult := result.(type) {
			case *api.GetPersonActionsOKApplicationJSON:
				// Convert API actions to template actions
				templateActions := make([]templates.Action, len(jsonResult.Actions))
				for i, action := range jsonResult.Actions {
					templateActions[i] = templates.Action{
						ID:          action.ID,
						PersonID:    action.PersonID,
						OccurredAt:  action.OccurredAt,
						Description: action.Description,
						References:  action.References.Or(""),
						Valence:     string(action.Valence),
						CreatedAt:   action.CreatedAt,
						UpdatedAt:   action.UpdatedAt,
					}
				}

				// Render template and return HTML response
				return &api.GetPersonActionsOKTextHTML{
					Data: renderTemplate(templates.ActionList(templateActions)),
				}, nil
			}
		}
	}

	// Return JSON response (default)
	return result, nil
}
