package handlers

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/rs/xid"
	xidb "github.com/rs/xid/b"

	"pepo/internal/api"
	"pepo/internal/db"
	"pepo/templates"
)

type ActionHandler struct {
	queries *db.Queries
}

func NewActionHandler(queries *db.Queries) *ActionHandler {
	return &ActionHandler{
		queries: queries,
	}
}

// Helper function to convert database action row to API action
func convertToAPIAction(id, personID xidb.ID, occurredAt time.Time, description string, references sql.NullString, valence db.ValenceType, createdAt, updatedAt time.Time) api.Action {
	apiAction := api.Action{
		ID:          id.String(),
		PersonID:    personID.String(),
		OccurredAt:  occurredAt,
		Description: description,
		Valence:     api.ActionValence(valence),
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}

	if references.Valid {
		apiAction.References = api.OptNilString{Value: references.String, Set: true}
	}

	return apiAction
}

// Helper function to convert API action to template action
func convertToTemplateAction(apiAction api.Action) templates.Action {
	templateAction := templates.Action{
		ID:          apiAction.ID,
		PersonID:    apiAction.PersonID,
		OccurredAt:  apiAction.OccurredAt,
		Description: apiAction.Description,
		Valence:     string(apiAction.Valence),
		CreatedAt:   apiAction.CreatedAt,
		UpdatedAt:   apiAction.UpdatedAt,
	}

	if apiAction.References.IsSet() {
		templateAction.References = apiAction.References.Value
	}

	return templateAction
}

// API Handlers

func (h *ActionHandler) CreateAction(ctx context.Context, req *api.CreateActionRequest) (api.CreateActionRes, error) {
	// Validate request
	if req.Description == "" {
		return &api.CreateActionBadRequest{
			Message: "Description is required",
			Code:    "VALIDATION_ERROR",
		}, nil
	}

	// Generate new xid for the action
	actionID := xid.New().String()

	// Use current time if occurred_at is not provided
	occurredAt := time.Now()
	if req.OccurredAt.IsSet() {
		occurredAt = req.OccurredAt.Value
	}

	// Create action in database
	action, err := h.queries.CreateAction(ctx, db.CreateActionParams{
		XidStr:      actionID,
		XidStr_2:    req.PersonID,
		OccurredAt:  occurredAt,
		Description: req.Description,
		References:  sql.NullString{String: req.References.Or(""), Valid: req.References.IsSet()},
		Valence:     db.ValenceType(req.Valence),
	})
	if err != nil {
		log.Printf("Error creating action: %v", err)
		return &api.CreateActionInternalServerError{
			Message: "Failed to create action",
			Code:    "INTERNAL_ERROR",
		}, nil
	}

	// Convert to API response
	apiAction := &api.Action{
		ID:          action.ID.String(),
		PersonID:    action.PersonID.String(),
		OccurredAt:  action.OccurredAt,
		Description: action.Description,
		Valence:     api.ActionValence(action.Valence),
		CreatedAt:   action.CreatedAt,
		UpdatedAt:   action.UpdatedAt,
	}

	if action.References.Valid {
		apiAction.References = api.OptNilString{Value: action.References.String, Set: true}
	}

	return apiAction, nil
}

func (h *ActionHandler) GetAction(ctx context.Context, params api.GetActionParams) (api.GetActionRes, error) {
	action, err := h.queries.GetActionByID(ctx, params.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return &api.GetActionNotFound{
				Message: "Action not found",
				Code:    "NOT_FOUND",
			}, nil
		}
		log.Printf("Error getting action: %v", err)
		return &api.GetActionInternalServerError{
			Message: "Failed to get action",
			Code:    "INTERNAL_ERROR",
		}, nil
	}

	apiAction := &api.Action{
		ID:          action.ID.String(),
		PersonID:    action.PersonID.String(),
		OccurredAt:  action.OccurredAt,
		Description: action.Description,
		Valence:     api.ActionValence(action.Valence),
		CreatedAt:   action.CreatedAt,
		UpdatedAt:   action.UpdatedAt,
	}

	if action.References.Valid {
		apiAction.References = api.OptNilString{Value: action.References.String, Set: true}
	}

	return apiAction, nil
}

func (h *ActionHandler) ListActions(ctx context.Context, params api.ListActionsParams) (api.ListActionsRes, error) {
	limit := int32(20)
	if params.Limit.IsSet() {
		limit = int32(params.Limit.Value)
	}

	offset := int32(0)
	if params.Offset.IsSet() {
		offset = int32(params.Offset.Value)
	}

	var apiActions []api.Action
	var total int64
	var err error

	// Handle different filtering options
	if params.PersonID.IsSet() && params.Valence.IsSet() {
		// Filter by both person and valence
		actions, err := h.queries.ListActionsByPersonIDAndValence(ctx, db.ListActionsByPersonIDAndValenceParams{
			XidStr:  params.PersonID.Value,
			Valence: db.ValenceType(params.Valence.Value),
			Limit:   limit,
			Offset:  offset,
		})
		if err == nil {
			apiActions = make([]api.Action, len(actions))
			for i, a := range actions {
				apiActions[i] = convertToAPIAction(a.ID, a.PersonID, a.OccurredAt, a.Description, a.References, a.Valence, a.CreatedAt, a.UpdatedAt)
			}
			total, err = h.queries.CountActionsByPersonID(ctx, params.PersonID.Value)
		}
	} else if params.PersonID.IsSet() {
		// Filter by person only
		actions, err := h.queries.ListActionsByPersonID(ctx, db.ListActionsByPersonIDParams{
			XidStr: params.PersonID.Value,
			Limit:  limit,
			Offset: offset,
		})
		if err == nil {
			apiActions = make([]api.Action, len(actions))
			for i, a := range actions {
				apiActions[i] = convertToAPIAction(a.ID, a.PersonID, a.OccurredAt, a.Description, a.References, a.Valence, a.CreatedAt, a.UpdatedAt)
			}
			total, err = h.queries.CountActionsByPersonID(ctx, params.PersonID.Value)
		}
	} else if params.Valence.IsSet() {
		// Filter by valence only
		actions, err := h.queries.ListActionsByValence(ctx, db.ListActionsByValenceParams{
			Valence: db.ValenceType(params.Valence.Value),
			Limit:   limit,
			Offset:  offset,
		})
		if err == nil {
			apiActions = make([]api.Action, len(actions))
			for i, a := range actions {
				apiActions[i] = convertToAPIAction(a.ID, a.PersonID, a.OccurredAt, a.Description, a.References, a.Valence, a.CreatedAt, a.UpdatedAt)
			}
			total, err = h.queries.CountActions(ctx)
		}
	} else {
		// No filters
		actions, err := h.queries.ListActions(ctx, db.ListActionsParams{
			Limit:  limit,
			Offset: offset,
		})
		if err == nil {
			apiActions = make([]api.Action, len(actions))
			for i, a := range actions {
				apiActions[i] = convertToAPIAction(a.ID, a.PersonID, a.OccurredAt, a.Description, a.References, a.Valence, a.CreatedAt, a.UpdatedAt)
			}
			total, err = h.queries.CountActions(ctx)
		}
	}

	if err != nil {
		log.Printf("Error listing actions: %v", err)
		return &api.Error{
			Message: "Failed to list actions",
			Code:    "INTERNAL_ERROR",
		}, nil
	}

	return &api.ListActionsOK{
		Actions: apiActions,
		Total:   int(total),
	}, nil
}

func (h *ActionHandler) UpdateAction(ctx context.Context, req *api.UpdateActionRequest, params api.UpdateActionParams) (api.UpdateActionRes, error) {
	// Validate request
	if req.Description == "" {
		return &api.UpdateActionBadRequest{
			Message: "Description is required",
			Code:    "VALIDATION_ERROR",
		}, nil
	}

	action, err := h.queries.UpdateAction(ctx, db.UpdateActionParams{
		XidStr:      params.ID,
		XidStr_2:    req.PersonID,
		OccurredAt:  req.OccurredAt,
		Description: req.Description,
		References:  sql.NullString{String: req.References.Or(""), Valid: req.References.IsSet()},
		Valence:     db.ValenceType(req.Valence),
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return &api.UpdateActionNotFound{
				Message: "Action not found",
				Code:    "NOT_FOUND",
			}, nil
		}
		log.Printf("Error updating action: %v", err)
		return &api.UpdateActionInternalServerError{
			Message: "Failed to update action",
			Code:    "INTERNAL_ERROR",
		}, nil
	}

	apiAction := &api.Action{
		ID:          action.ID.String(),
		PersonID:    action.PersonID.String(),
		OccurredAt:  action.OccurredAt,
		Description: action.Description,
		Valence:     api.ActionValence(action.Valence),
		CreatedAt:   action.CreatedAt,
		UpdatedAt:   action.UpdatedAt,
	}

	if action.References.Valid {
		apiAction.References = api.OptNilString{Value: action.References.String, Set: true}
	}

	return apiAction, nil
}

func (h *ActionHandler) DeleteAction(ctx context.Context, params api.DeleteActionParams) (api.DeleteActionRes, error) {
	err := h.queries.DeleteAction(ctx, params.ID)
	if err != nil {
		log.Printf("Error deleting action: %v", err)
		return &api.DeleteActionInternalServerError{
			Message: "Failed to delete action",
			Code:    "INTERNAL_ERROR",
		}, nil
	}

	return &api.DeleteActionNoContent{}, nil
}

func (h *ActionHandler) GetPersonActions(ctx context.Context, params api.GetPersonActionsParams) (api.GetPersonActionsRes, error) {
	limit := int32(20)
	if params.Limit.IsSet() {
		limit = int32(params.Limit.Value)
	}

	offset := int32(0)
	if params.Offset.IsSet() {
		offset = int32(params.Offset.Value)
	}

	var apiActions []api.Action
	var err error

	if params.Valence.IsSet() {
		actions, err := h.queries.ListActionsByPersonIDAndValence(ctx, db.ListActionsByPersonIDAndValenceParams{
			XidStr:  params.ID,
			Valence: db.ValenceType(params.Valence.Value),
			Limit:   limit,
			Offset:  offset,
		})
		if err == nil {
			apiActions = make([]api.Action, len(actions))
			for i, a := range actions {
				apiActions[i] = convertToAPIAction(a.ID, a.PersonID, a.OccurredAt, a.Description, a.References, a.Valence, a.CreatedAt, a.UpdatedAt)
			}
		}
	} else {
		actions, err := h.queries.ListActionsByPersonID(ctx, db.ListActionsByPersonIDParams{
			XidStr: params.ID,
			Limit:  limit,
			Offset: offset,
		})
		if err == nil {
			apiActions = make([]api.Action, len(actions))
			for i, a := range actions {
				apiActions[i] = convertToAPIAction(a.ID, a.PersonID, a.OccurredAt, a.Description, a.References, a.Valence, a.CreatedAt, a.UpdatedAt)
			}
		}
	}

	if err != nil {
		log.Printf("Error getting person actions: %v", err)
		return &api.GetPersonActionsInternalServerError{
			Message: "Failed to get person actions",
			Code:    "INTERNAL_ERROR",
		}, nil
	}

	// Get total count for this person
	total, err := h.queries.CountActionsByPersonID(ctx, params.ID)
	if err != nil {
		log.Printf("Error counting person actions: %v", err)
		return &api.GetPersonActionsInternalServerError{
			Message: "Failed to count person actions",
			Code:    "INTERNAL_ERROR",
		}, nil
	}

	return &api.GetPersonActionsOK{
		Actions: apiActions,
		Total:   int(total),
	}, nil
}

// Form Handlers

func (h *ActionHandler) HandleCreateActionForm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	personID := r.FormValue("person_id")
	description := r.FormValue("description")
	references := r.FormValue("references")
	valence := r.FormValue("valence")
	occurredAtStr := r.FormValue("occurred_at")

	if personID == "" || description == "" || valence == "" {
		w.WriteHeader(http.StatusBadRequest)
		templates.ActionError("Person, description, and valence are required").Render(r.Context(), w)
		return
	}

	// Parse occurred_at if provided, otherwise use current time
	var occurredAt time.Time
	if occurredAtStr != "" {
		parsed, err := time.Parse("2006-01-02T15:04", occurredAtStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			templates.ActionError("Invalid date format").Render(r.Context(), w)
			return
		}
		occurredAt = parsed
	} else {
		occurredAt = time.Now()
	}

	// Create API request
	req := &api.CreateActionRequest{
		PersonID:    personID,
		Description: description,
		Valence:     api.CreateActionRequestValence(valence),
	}

	if occurredAtStr != "" {
		req.OccurredAt = api.OptDateTime{Value: occurredAt, Set: true}
	}

	if references != "" {
		req.References = api.OptNilString{Value: references, Set: true}
	}

	// Call the API handler internally
	result, err := h.CreateAction(r.Context(), req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		templates.ActionError("Failed to create action").Render(r.Context(), w)
		return
	}

	// Check if result is an action or error
	switch action := result.(type) {
	case *api.Action:
		templateAction := convertToTemplateAction(*action)
		w.Header().Set("Content-Type", "text/html")
		templates.ActionItem(templateAction).Render(r.Context(), w)
	default:
		w.WriteHeader(http.StatusBadRequest)
		templates.ActionError("Failed to create action").Render(r.Context(), w)
	}
}

func (h *ActionHandler) HandleListActionsHTML(w http.ResponseWriter, r *http.Request) {
	// Call the API handler internally
	params := api.ListActionsParams{
		Limit:  api.OptInt{Value: 50, Set: true},
		Offset: api.OptInt{Value: 0, Set: true},
	}

	result, err := h.ListActions(r.Context(), params)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		templates.ActionError("Failed to load actions").Render(r.Context(), w)
		return
	}

	switch listResult := result.(type) {
	case *api.ListActionsOK:
		// Convert to template actions
		templateActions := make([]templates.Action, len(listResult.Actions))
		for i, action := range listResult.Actions {
			templateActions[i] = convertToTemplateAction(action)
		}

		w.Header().Set("Content-Type", "text/html")
		templates.ActionList(templateActions).Render(r.Context(), w)
	default:
		w.WriteHeader(http.StatusInternalServerError)
		templates.ActionError("Failed to load actions").Render(r.Context(), w)
	}
}

func (h *ActionHandler) HandleDeleteActionForm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract action ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		w.WriteHeader(http.StatusBadRequest)
		templates.ActionError("Invalid action ID").Render(r.Context(), w)
		return
	}
	actionID := pathParts[4] // /forms/actions/delete/{id}

	// Call the API handler internally
	params := api.DeleteActionParams{ID: actionID}
	_, err := h.DeleteAction(r.Context(), params)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		templates.ActionError("Failed to delete action").Render(r.Context(), w)
		return
	}

	// Return empty content to remove the element
	w.WriteHeader(http.StatusOK)
}
