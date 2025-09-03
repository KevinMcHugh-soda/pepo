package handlers

import (
	"context"

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

	return &api.Conversation{
		ID:          id,
		PersonID:    req.PersonID,
		OccurredAt:  conv.OccurredAt,
		Description: conv.Description,
		CreatedAt:   conv.CreatedAt,
		UpdatedAt:   conv.UpdatedAt,
	}, nil
}
