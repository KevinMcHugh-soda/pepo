package handlers

import (
	"context"

	"pepo/internal/api"
	"pepo/templates"
)

// CombinedAPIHandler implements all ogen interfaces by delegating to specific handlers
type CombinedAPIHandler struct {
	personHandler       *PersonHandler
	actionHandler       *ActionHandler
	conversationHandler *ConversationHandler
}

// NewCombinedAPIHandler creates a new combined API handler
func NewCombinedAPIHandler(personHandler *PersonHandler, actionHandler *ActionHandler, conversationHandler *ConversationHandler) *CombinedAPIHandler {
	return &CombinedAPIHandler{
		personHandler:       personHandler,
		actionHandler:       actionHandler,
		conversationHandler: conversationHandler,
	}
}

// Person API methods
func (h *CombinedAPIHandler) CreatePerson(ctx context.Context, req *api.CreatePersonRequest) (api.CreatePersonRes, error) {
	return h.personHandler.CreatePerson(ctx, req)
}

func (h *CombinedAPIHandler) GetPersonById(ctx context.Context, params api.GetPersonByIdParams) (api.GetPersonByIdRes, error) {
	return h.personHandler.GetPersonById(ctx, params)
}

func (h *CombinedAPIHandler) GetPersons(ctx context.Context, params api.GetPersonsParams) (api.GetPersonsRes, error) {
	return h.personHandler.GetPersons(ctx, params)
}

func (h *CombinedAPIHandler) UpdatePerson(ctx context.Context, req *api.UpdatePersonRequest, params api.UpdatePersonParams) (api.UpdatePersonRes, error) {
	return h.personHandler.UpdatePerson(ctx, req, params)
}

func (h *CombinedAPIHandler) DeletePerson(ctx context.Context, params api.DeletePersonParams) (api.DeletePersonRes, error) {
	return h.personHandler.DeletePerson(ctx, params)
}

// Action API methods
func (h *CombinedAPIHandler) CreateAction(ctx context.Context, req *api.CreateActionRequest) (api.CreateActionRes, error) {
	return h.actionHandler.CreateAction(ctx, req)
}

func (h *CombinedAPIHandler) GetActionById(ctx context.Context, params api.GetActionByIdParams) (api.GetActionByIdRes, error) {
	return h.actionHandler.GetActionById(ctx, params)
}

func (h *CombinedAPIHandler) GetActions(ctx context.Context, params api.GetActionsParams) (api.GetActionsRes, error) {
	return h.actionHandler.GetActions(ctx, params)
}

func (h *CombinedAPIHandler) UpdateAction(ctx context.Context, req *api.UpdateActionRequest, params api.UpdateActionParams) (api.UpdateActionRes, error) {
	return h.actionHandler.UpdateAction(ctx, req, params)
}

func (h *CombinedAPIHandler) DeleteAction(ctx context.Context, params api.DeleteActionParams) (api.DeleteActionRes, error) {
	return h.actionHandler.DeleteAction(ctx, params)
}

func (h *CombinedAPIHandler) GetPersonActions(ctx context.Context, params api.GetPersonActionsParams) (api.GetPersonActionsRes, error) {
	return h.actionHandler.GetPersonActions(ctx, params)
}

func (h *CombinedAPIHandler) GetPersonTimeline(ctx context.Context, params api.GetPersonTimelineParams) (api.GetPersonTimelineRes, error) {
	return h.personHandler.GetPersonTimeline(ctx, params)
}

func (h *CombinedAPIHandler) GetPersonsWithLastActivity(ctx context.Context, params api.GetPersonsParams) ([]templates.PersonWithLastActivity, error) {
	return h.personHandler.GetPersonsWithLastActivity(ctx, params)
}

// Conversation API methods
func (h *CombinedAPIHandler) CreateConversation(ctx context.Context, req *api.CreateConversationRequest) (api.CreateConversationRes, error) {
	return h.conversationHandler.CreateConversation(ctx, req)
}

func (h *CombinedAPIHandler) GetConversationById(ctx context.Context, params api.GetConversationByIdParams) (api.GetConversationByIdRes, error) {
	return h.conversationHandler.GetConversationById(ctx, params)
}

func (h *CombinedAPIHandler) UpdateConversation(ctx context.Context, req *api.UpdateConversationRequest, params api.UpdateConversationParams) (api.UpdateConversationRes, error) {
	return h.conversationHandler.UpdateConversation(ctx, req, params)
}

func (h *CombinedAPIHandler) DeleteConversation(ctx context.Context, params api.DeleteConversationParams) (api.DeleteConversationRes, error) {
	return h.conversationHandler.DeleteConversation(ctx, params)
}
