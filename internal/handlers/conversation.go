package handlers

import (
	"context"
	"database/sql"

	"github.com/rs/xid"
	"go.uber.org/zap"

	"pepo/internal/api"
	"pepo/internal/db"
)

type ConversationHandler struct {
	queries *db.Queries
}

func NewConversationHandler(queries *db.Queries) *ConversationHandler {
	return &ConversationHandler{queries: queries}
}

// CreateConversation handles creating a conversation
func (h *ConversationHandler) CreateConversation(ctx context.Context, req *api.CreateConversationRequest) (api.CreateConversationRes, error) {
	if req.PersonID == "" {
		return &api.CreateConversationBadRequest{
			Message: "Person ID is required",
			Code:    "VALIDATION_ERROR",
		}, nil
	}
	if req.Description == "" {
		return &api.CreateConversationBadRequest{
			Message: "Description is required",
			Code:    "VALIDATION_ERROR",
		}, nil
	}

	id := xid.New().String()

	row, err := h.queries.CreateConversation(ctx, db.CreateConversationParams{
		ID:          id,
		PersonID:    req.PersonID,
		Description: req.Description,
		OccurredAt:  req.OccurredAt,
	})
	if err != nil {
		zap.L().Error("error creating conversation", zap.Error(err))
		return &api.CreateConversationInternalServerError{
			Message: "Failed to create conversation",
			Code:    "INTERNAL_ERROR",
		}, nil
	}

	conv := row.Conversation

	// Associate provided actions with the new conversation
	for _, aID := range req.Actions {
		if _, err := h.queries.GetActionByID(ctx, aID); err != nil {
			if err == sql.ErrNoRows {
				return &api.CreateConversationBadRequest{
					Message: "Action not found",
					Code:    "INVALID_ACTION",
				}, nil
			}
			zap.L().Error("error getting action", zap.Error(err))
			return &api.CreateConversationInternalServerError{
				Message: "Failed to associate action",
				Code:    "INTERNAL_ERROR",
			}, nil
		}
		if err := h.queries.AddActionToConversation(ctx, db.AddActionToConversationParams{
			ActionID:       aID,
			ConversationID: id,
		}); err != nil {
			zap.L().Error("error adding action to conversation", zap.Error(err))
			return &api.CreateConversationInternalServerError{
				Message: "Failed to associate action",
				Code:    "INTERNAL_ERROR",
			}, nil
		}
	}

	// Associate provided themes with the new conversation
	for _, tID := range req.Themes {
		if _, err := h.queries.GetThemeByID(ctx, tID); err != nil {
			if err == sql.ErrNoRows {
				return &api.CreateConversationBadRequest{
					Message: "Theme not found",
					Code:    "INVALID_THEME",
				}, nil
			}
			zap.L().Error("error getting theme", zap.Error(err))
			return &api.CreateConversationInternalServerError{
				Message: "Failed to associate theme",
				Code:    "INTERNAL_ERROR",
			}, nil
		}
		if err := h.queries.AddThemeToConversation(ctx, db.AddThemeToConversationParams{
			ConversationID: id,
			ThemeID:        tID,
		}); err != nil {
			zap.L().Error("error adding theme to conversation", zap.Error(err))
			return &api.CreateConversationInternalServerError{
				Message: "Failed to associate theme",
				Code:    "INTERNAL_ERROR",
			}, nil
		}
	}

	return &api.Conversation{
		ID:          id,
		PersonID:    req.PersonID,
		OccurredAt:  conv.OccurredAt,
		Description: conv.Description,
		CreatedAt:   conv.CreatedAt,
		UpdatedAt:   conv.UpdatedAt,
	}, nil
}

// GetConversationById retrieves a conversation by ID
func (h *ConversationHandler) GetConversationById(ctx context.Context, params api.GetConversationByIdParams) (api.GetConversationByIdRes, error) {
	row, err := h.queries.GetConversationByID(ctx, params.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return &api.GetConversationByIdNotFound{
				Message: "Conversation not found",
				Code:    "NOT_FOUND",
			}, nil
		}
		zap.L().Error("error getting conversation", zap.Error(err))
		return &api.GetConversationByIdInternalServerError{
			Message: "Failed to get conversation",
			Code:    "INTERNAL_ERROR",
		}, nil
	}

	c := row.Conversation
	convID, _ := xid.FromBytes(c.ID)
	personID, _ := xid.FromBytes(c.PersonID)
	return &api.Conversation{
		ID:          convID.String(),
		PersonID:    personID.String(),
		OccurredAt:  c.OccurredAt,
		Description: c.Description,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}, nil
}

// UpdateConversation handles updating a conversation
func (h *ConversationHandler) UpdateConversation(ctx context.Context, req *api.UpdateConversationRequest, params api.UpdateConversationParams) (api.UpdateConversationRes, error) {
	if req.PersonID == "" {
		return &api.UpdateConversationBadRequest{
			Message: "Person ID is required",
			Code:    "VALIDATION_ERROR",
		}, nil
	}
	if req.Description == "" {
		return &api.UpdateConversationBadRequest{
			Message: "Description is required",
			Code:    "VALIDATION_ERROR",
		}, nil
	}

	row, err := h.queries.UpdateConversation(ctx, db.UpdateConversationParams{
		ID:          params.ID,
		PersonID:    req.PersonID,
		Description: req.Description,
		OccurredAt:  req.OccurredAt,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return &api.UpdateConversationNotFound{
				Message: "Conversation not found",
				Code:    "NOT_FOUND",
			}, nil
		}
		zap.L().Error("error updating conversation", zap.Error(err))
		return &api.UpdateConversationInternalServerError{
			Message: "Failed to update conversation",
			Code:    "INTERNAL_ERROR",
		}, nil
	}

	// Update associated actions
	if err := h.queries.DeleteActionsByConversationID(ctx, params.ID); err != nil {
		zap.L().Error("error deleting conversation actions", zap.Error(err))
		return &api.UpdateConversationInternalServerError{
			Message: "Failed to update actions",
			Code:    "INTERNAL_ERROR",
		}, nil
	}
	for _, aID := range req.Actions {
		if err := h.queries.AddActionToConversation(ctx, db.AddActionToConversationParams{
			ActionID:       aID,
			ConversationID: params.ID,
		}); err != nil {
			zap.L().Error("error adding action to conversation", zap.Error(err))
			return &api.UpdateConversationInternalServerError{
				Message: "Failed to update actions",
				Code:    "INTERNAL_ERROR",
			}, nil
		}
	}

	// Update associated themes
	if err := h.queries.DeleteThemesByConversationID(ctx, params.ID); err != nil {
		zap.L().Error("error deleting conversation themes", zap.Error(err))
		return &api.UpdateConversationInternalServerError{
			Message: "Failed to update themes",
			Code:    "INTERNAL_ERROR",
		}, nil
	}
	for _, tID := range req.Themes {
		if err := h.queries.AddThemeToConversation(ctx, db.AddThemeToConversationParams{
			ConversationID: params.ID,
			ThemeID:        tID,
		}); err != nil {
			zap.L().Error("error adding theme to conversation", zap.Error(err))
			return &api.UpdateConversationInternalServerError{
				Message: "Failed to update themes",
				Code:    "INTERNAL_ERROR",
			}, nil
		}
	}

	conv := row.Conversation
	convID, _ := xid.FromBytes(conv.ID)
	personID, _ := xid.FromBytes(conv.PersonID)
	return &api.Conversation{
		ID:          convID.String(),
		PersonID:    personID.String(),
		OccurredAt:  conv.OccurredAt,
		Description: conv.Description,
		CreatedAt:   conv.CreatedAt,
		UpdatedAt:   conv.UpdatedAt,
	}, nil
}

// DeleteConversation deletes a conversation
func (h *ConversationHandler) DeleteConversation(ctx context.Context, params api.DeleteConversationParams) (api.DeleteConversationRes, error) {
	if err := h.queries.DeleteActionsByConversationID(ctx, params.ID); err != nil {
		zap.L().Error("error deleting conversation actions", zap.Error(err))
	}
	if err := h.queries.DeleteThemesByConversationID(ctx, params.ID); err != nil {
		zap.L().Error("error deleting conversation themes", zap.Error(err))
	}
	if err := h.queries.DeleteConversation(ctx, params.ID); err != nil {
		zap.L().Error("error deleting conversation", zap.Error(err))
		return &api.DeleteConversationInternalServerError{
			Message: "Failed to delete conversation",
			Code:    "INTERNAL_ERROR",
		}, nil
	}
	return &api.DeleteConversationNoContent{}, nil
}
