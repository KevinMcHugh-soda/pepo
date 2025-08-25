package handlers

import (
	"context"
	"database/sql"

	"github.com/rs/xid"

	"pepo/internal/api"
	"pepo/internal/db"

	"go.uber.org/zap"
)

type ConversationHandler struct {
	queries *db.Queries
}

func NewConversationHandler(queries *db.Queries) *ConversationHandler {
	return &ConversationHandler{queries: queries}
}

func convertToAPIConversation(conv db.Conversation) api.Conversation {
	id, _ := xid.FromBytes(conv.ID)
	return api.Conversation{
		ID:          id.String(),
		OccurredAt:  conv.OccurredAt,
		Description: conv.Description,
		CreatedAt:   conv.CreatedAt,
		UpdatedAt:   conv.UpdatedAt,
	}
}

func (h *ConversationHandler) CreateConversation(ctx context.Context, req *api.CreateConversationRequest) (api.CreateConversationRes, error) {
	convID := xid.New().String()

	conv, err := h.queries.CreateConversation(ctx, db.CreateConversationParams{
		ID:          convID,
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

	// Associate actions
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
			ConversationID: convID,
		}); err != nil {
			zap.L().Error("error adding action to conversation", zap.Error(err))
			return &api.CreateConversationInternalServerError{
				Message: "Failed to associate action",
				Code:    "INTERNAL_ERROR",
			}, nil
		}
	}

	// Associate themes
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
			ConversationID: convID,
			ThemeID:        tID,
		}); err != nil {
			zap.L().Error("error adding theme to conversation", zap.Error(err))
			return &api.CreateConversationInternalServerError{
				Message: "Failed to associate theme",
				Code:    "INTERNAL_ERROR",
			}, nil
		}
	}

	apiConv := convertToAPIConversation(conv)
	return &apiConv, nil
}
