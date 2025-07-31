package handlers

import (
	"context"
	"database/sql"
	"log"

	"github.com/rs/xid"

	"pepo/internal/api"
	"pepo/internal/db"
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
func convertToAPIAction(action db.Action) api.Action {
	apiAction := api.Action{
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

	return apiAction
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

	// Use the provided occurred_at time
	occurredAt := req.OccurredAt

	// Create action in database
	row, err := h.queries.CreateAction(ctx, db.CreateActionParams{
		XidStr:      actionID,
		XidStr_2:    req.PersonID,
		OccurredAt:  occurredAt,
		Description: req.Description,
		References:  sql.NullString{String: req.References.Or(""), Valid: req.References.IsSet()},
		Valence:     db.ValenceType(req.Valence),
	})
	action := row.Action
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

func (h *ActionHandler) GetActionById(ctx context.Context, params api.GetActionByIdParams) (api.GetActionByIdRes, error) {
	row, err := h.queries.GetActionByID(ctx, params.ID)
	action := row.Action
	if err != nil {
		if err == sql.ErrNoRows {
			return &api.GetActionByIdNotFound{
				Message: "Action not found",
				Code:    "NOT_FOUND",
			}, nil
		}
		log.Printf("Error getting action: %v", err)
		return &api.GetActionByIdInternalServerError{
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

func (h *ActionHandler) GetActions(ctx context.Context, params api.GetActionsParams) (api.GetActionsRes, error) {
	limit := int32(10)
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
		rows, err := h.queries.ListActionsByPersonIDAndValence(ctx, db.ListActionsByPersonIDAndValenceParams{
			XidStr:  params.PersonID.Value,
			Valence: db.ValenceType(params.Valence.Value),
			Limit:   limit,
			Offset:  offset,
		})
		if err == nil {
			apiActions = make([]api.Action, len(rows))
			for i, row := range rows {
				a := row.Action
				apiActions[i] = convertToAPIAction(a)
			}
			total, err = h.queries.CountActionsByPersonID(ctx, params.PersonID.Value)
		}
	} else if params.PersonID.IsSet() {
		// Filter by person only
		rows, err := h.queries.ListActionsByPersonID(ctx, db.ListActionsByPersonIDParams{
			XidStr: params.PersonID.Value,
			Limit:  limit,
			Offset: offset,
		})
		if err == nil {
			apiActions = make([]api.Action, len(rows))
			for i, row := range rows {
				a := row.Action
				apiActions[i] = convertToAPIAction(a)
			}
			total, err = h.queries.CountActionsByPersonID(ctx, params.PersonID.Value)
		}
	} else if params.Valence.IsSet() {
		// Filter by valence only
		rows, err := h.queries.ListActionsByValence(ctx, db.ListActionsByValenceParams{
			Valence: db.ValenceType(params.Valence.Value),
			Limit:   limit,
			Offset:  offset,
		})
		if err == nil {
			apiActions = make([]api.Action, len(rows))
			for i, row := range rows {
				a := row.Action
				apiActions[i] = convertToAPIAction(a)
			}
			total, err = h.queries.CountActions(ctx)
		}
	} else {
		// No filters
		rows, err := h.queries.ListActions(ctx, db.ListActionsParams{
			Limit:  limit,
			Offset: offset,
		})
		if err == nil {
			apiActions = make([]api.Action, len(rows))
			for i, row := range rows {
				a := row.Action
				action := convertToAPIAction(a)
				action.PersonName = api.NewOptString(row.PersonName)
				apiActions[i] = action
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

	return &api.GetActionsOKApplicationJSON{
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

	row, err := h.queries.UpdateAction(ctx, db.UpdateActionParams{
		XidStr:      params.ID,
		XidStr_2:    req.PersonID,
		OccurredAt:  req.OccurredAt,
		Description: req.Description,
		References:  sql.NullString{String: req.References.Or(""), Valid: req.References.IsSet()},
		Valence:     db.ValenceType(req.Valence),
	})
	action := row.Action
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
		rows, err := h.queries.ListActionsByPersonIDAndValence(ctx, db.ListActionsByPersonIDAndValenceParams{
			XidStr:  params.ID,
			Valence: db.ValenceType(params.Valence.Value),
			Limit:   limit,
			Offset:  offset,
		})
		if err == nil {
			apiActions = make([]api.Action, len(rows))
			for i, r := range rows {
				a := r.Action
				apiActions[i] = convertToAPIAction(a)
			}
		}
	} else {
		rows, err := h.queries.ListActionsByPersonID(ctx, db.ListActionsByPersonIDParams{
			XidStr: params.ID,
			Limit:  limit,
			Offset: offset,
		})
		if err == nil {
			apiActions = make([]api.Action, len(rows))
			for i, r := range rows {
				a := r.Action

				apiActions[i] = convertToAPIAction(a)
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

	return &api.GetPersonActionsOKApplicationJSON{
		Actions: apiActions,
		Total:   int(total),
	}, nil
}
