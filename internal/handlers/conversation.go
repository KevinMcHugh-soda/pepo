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
