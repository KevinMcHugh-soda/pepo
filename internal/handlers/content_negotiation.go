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
				templatePersons, err := h.combinedHandler.GetPersonsWithLastActivity(ctx, params)
				if err != nil {
					return &api.Error{
						Message: "Failed to get people with last actions",
						Code:    "INTERNAL_ERROR",
					}, nil
				}

				// Render person list with last action template
				return &api.GetPersonsOKTextHTML{
					Data: renderTemplate(templates.PersonWithLastActivityTable(templatePersons)),
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

				// Fetch the person's timeline for the detail view
				timelineParams := api.GetPersonTimelineParams{
					ID:    params.ID,
					Limit: api.OptInt{Value: 100, Set: true},
				}

				timelineResult, err := h.combinedHandler.GetPersonTimeline(ctx, timelineParams)
				if err != nil {
					return &api.GetPersonByIdOKTextHTML{
						Data: renderTemplate(templates.PersonDetail(templatePerson, []templates.TimelineItem{})),
					}, nil
				}

				var templateItems []templates.TimelineItem
				if timelineJSON, ok := timelineResult.(*api.GetPersonTimelineOKApplicationJSON); ok {
					templateItems = make([]templates.TimelineItem, len(timelineJSON.Items))
					for i, item := range timelineJSON.Items {
						switch item.Type {
						case api.TimelineItemTypeAction:
							tmplAction := &templates.Action{
								ID:          item.ID,
								PersonID:    item.PersonID,
								OccurredAt:  item.OccurredAt,
								Description: item.Description,
								References:  item.References.Or(""),
								Valence:     string(item.Valence.Or("")),
								CreatedAt:   item.CreatedAt,
								UpdatedAt:   item.UpdatedAt,
							}
							if len(item.Themes) > 0 {
								tmplThemes := make([]templates.Theme, len(item.Themes))
								for j, th := range item.Themes {
									tmplThemes[j] = templates.Theme{ID: th.ID, Text: th.Text}
								}
								tmplAction.Themes = tmplThemes
							}
							templateItems[i] = templates.TimelineItem{Type: "action", Action: tmplAction}
						case api.TimelineItemTypeConversation:
							tmplConv := &templates.Conversation{
								ID:          item.ID,
								PersonID:    item.PersonID,
								OccurredAt:  item.OccurredAt,
								Description: item.Description,
								CreatedAt:   item.CreatedAt,
								UpdatedAt:   item.UpdatedAt,
							}
							if len(item.Themes) > 0 {
								tmplThemes := make([]templates.Theme, len(item.Themes))
								for j, th := range item.Themes {
									tmplThemes[j] = templates.Theme{ID: th.ID, Text: th.Text}
								}
								tmplConv.Themes = tmplThemes
							}
							templateItems[i] = templates.TimelineItem{Type: "conversation", Conversation: tmplConv}
						}
					}
				}

				return &api.GetPersonByIdOKTextHTML{
					Data: renderTemplate(templates.PersonDetail(templatePerson, templateItems)),
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
						PersonName:  action.PersonName.Value,
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
			format := req.URL.Query().Get("format")

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
					PersonName:  jsonResult.PersonName.Value,
				}

				if format == "edit" {
					return &api.GetActionByIdOKTextHTML{
						Data: renderTemplate(templates.EditActionForm(templateAction)),
					}, nil
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

// CreateConversation handles both JSON and HTML requests for creating a conversation
func (h *ContentNegotiatingHandler) CreateConversation(ctx context.Context, req *api.CreateConversationRequest) (api.CreateConversationRes, error) {
	result, err := h.combinedHandler.CreateConversation(ctx, req)
	if err != nil {
		return result, err
	}

	if httpReq := h.getRequestFromContext(ctx); httpReq != nil {
		if h.determineResponseType(httpReq) == "text/html" {
			switch conv := result.(type) {
			case *api.Conversation:
				tmplConv := templates.Conversation{
					ID:          conv.ID,
					PersonID:    conv.PersonID,
					OccurredAt:  conv.OccurredAt,
					Description: conv.Description,
					CreatedAt:   conv.CreatedAt,
					UpdatedAt:   conv.UpdatedAt,
				}
				return &api.CreateConversationCreatedTextHTML{
					Data: renderTemplate(templates.ConversationItem(tmplConv)),
				}, nil
			}
		}
	}

	return result, nil
}

// GetConversationById handles both JSON and HTML requests for getting a conversation by ID
func (h *ContentNegotiatingHandler) GetConversationById(ctx context.Context, params api.GetConversationByIdParams) (api.GetConversationByIdRes, error) {
	result, err := h.combinedHandler.GetConversationById(ctx, params)
	if err != nil {
		return result, err
	}

	if req := h.getRequestFromContext(ctx); req != nil {
		if h.determineResponseType(req) == "text/html" {
			format := req.URL.Query().Get("format")
			switch conv := result.(type) {
			case *api.Conversation:
				tmplConv := templates.Conversation{
					ID:          conv.ID,
					PersonID:    conv.PersonID,
					OccurredAt:  conv.OccurredAt,
					Description: conv.Description,
					CreatedAt:   conv.CreatedAt,
					UpdatedAt:   conv.UpdatedAt,
				}
				if format == "edit" {
					return &api.GetConversationByIdOKTextHTML{Data: renderTemplate(templates.EditConversationForm(tmplConv))}, nil
				}
				return &api.GetConversationByIdOKTextHTML{Data: renderTemplate(templates.ConversationItem(tmplConv))}, nil
			}
		}
	}

	return result, nil
}

// UpdateConversation handles both JSON and HTML requests for updating a conversation
func (h *ContentNegotiatingHandler) UpdateConversation(ctx context.Context, req *api.UpdateConversationRequest, params api.UpdateConversationParams) (api.UpdateConversationRes, error) {
	result, err := h.combinedHandler.UpdateConversation(ctx, req, params)
	if err != nil {
		return result, err
	}

	if httpReq := h.getRequestFromContext(ctx); httpReq != nil {
		if h.determineResponseType(httpReq) == "text/html" {
			switch conv := result.(type) {
			case *api.Conversation:
				tmplConv := templates.Conversation{
					ID:          conv.ID,
					PersonID:    conv.PersonID,
					OccurredAt:  conv.OccurredAt,
					Description: conv.Description,
					CreatedAt:   conv.CreatedAt,
					UpdatedAt:   conv.UpdatedAt,
				}
				return &api.UpdateConversationOKTextHTML{Data: renderTemplate(templates.ConversationItem(tmplConv))}, nil
			}
		}
	}

	return result, nil
}

// DeleteConversation handles deleting a conversation
func (h *ContentNegotiatingHandler) DeleteConversation(ctx context.Context, params api.DeleteConversationParams) (api.DeleteConversationRes, error) {
	return h.combinedHandler.DeleteConversation(ctx, params)
}

// GetPersonTimeline handles both JSON and HTML requests for a person's timeline
func (h *ContentNegotiatingHandler) GetPersonTimeline(ctx context.Context, params api.GetPersonTimelineParams) (api.GetPersonTimelineRes, error) {
	result, err := h.combinedHandler.GetPersonTimeline(ctx, params)
	if err != nil {
		return result, err
	}

	if req := h.getRequestFromContext(ctx); req != nil {
		if h.determineResponseType(req) == "text/html" {
			if jsonResult, ok := result.(*api.GetPersonTimelineOKApplicationJSON); ok {
				templateItems := make([]templates.TimelineItem, len(jsonResult.Items))
				for i, item := range jsonResult.Items {
					switch item.Type {
					case api.TimelineItemTypeAction:
						tmplAction := &templates.Action{
							ID:          item.ID,
							PersonID:    item.PersonID,
							OccurredAt:  item.OccurredAt,
							Description: item.Description,
							References:  item.References.Or(""),
							Valence:     string(item.Valence.Or("")),
							CreatedAt:   item.CreatedAt,
							UpdatedAt:   item.UpdatedAt,
						}
						if len(item.Themes) > 0 {
							tmplThemes := make([]templates.Theme, len(item.Themes))
							for j, th := range item.Themes {
								tmplThemes[j] = templates.Theme{ID: th.ID, Text: th.Text}
							}
							tmplAction.Themes = tmplThemes
						}
						templateItems[i] = templates.TimelineItem{Type: "action", Action: tmplAction}
					case api.TimelineItemTypeConversation:
						tmplConv := &templates.Conversation{
							ID:          item.ID,
							PersonID:    item.PersonID,
							OccurredAt:  item.OccurredAt,
							Description: item.Description,
							CreatedAt:   item.CreatedAt,
							UpdatedAt:   item.UpdatedAt,
						}
						if len(item.Themes) > 0 {
							tmplThemes := make([]templates.Theme, len(item.Themes))
							for j, th := range item.Themes {
								tmplThemes[j] = templates.Theme{ID: th.ID, Text: th.Text}
							}
							tmplConv.Themes = tmplThemes
						}
						templateItems[i] = templates.TimelineItem{Type: "conversation", Conversation: tmplConv}
					}
				}
				return &api.GetPersonTimelineOKTextHTML{
					Data: renderTemplate(templates.TimelineList(templateItems)),
				}, nil
			}
		}
	}

	return result, nil
}
