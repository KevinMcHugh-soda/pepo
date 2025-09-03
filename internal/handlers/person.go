package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"sort"

	"github.com/rs/xid"

	"pepo/internal/api"
	"pepo/internal/db"
	"pepo/templates"

	"go.uber.org/zap"
)

type PersonHandler struct {
	queries *db.Queries
}

func NewPersonHandler(queries *db.Queries) *PersonHandler {
	return &PersonHandler{
		queries: queries,
	}
}

// API Handlers

func (h *PersonHandler) CreatePerson(ctx context.Context, req *api.CreatePersonRequest) (api.CreatePersonRes, error) {
	// Validate request
	if req.Name == "" {
		return &api.CreatePersonBadRequest{
			Message: "Name is required",
			Code:    "VALIDATION_ERROR",
		}, nil
	}

	// Generate new xid for the person
	personID := xid.New().String()

	// Create person in database
	person, err := h.queries.CreatePerson(ctx, db.CreatePersonParams{
		ID:   personID,
		Name: req.Name,
	})
	if err != nil {
		zap.L().Error("error creating person", zap.Error(err))
		return &api.CreatePersonInternalServerError{
			Message: "Failed to create person",
			Code:    "INTERNAL_ERROR",
		}, nil
	}

	// Convert to API response
	return &api.Person{
		ID:        person.ID,
		Name:      person.Name,
		CreatedAt: person.CreatedAt,
		UpdatedAt: person.UpdatedAt,
	}, nil
}

func (h *PersonHandler) GetPersonById(ctx context.Context, params api.GetPersonByIdParams) (api.GetPersonByIdRes, error) {
	person, err := h.queries.GetPersonByID(ctx, params.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return &api.GetPersonByIdNotFound{
				Message: "Person not found",
				Code:    "NOT_FOUND",
			}, nil
		}
		zap.L().Error("error getting person", zap.Error(err))
		return &api.GetPersonByIdInternalServerError{
			Message: "Failed to get person",
			Code:    "INTERNAL_ERROR",
		}, nil
	}

	return &api.Person{
		ID:        person.ID,
		Name:      person.Name,
		CreatedAt: person.CreatedAt,
		UpdatedAt: person.UpdatedAt,
	}, nil
}

func (h *PersonHandler) GetPersons(ctx context.Context, params api.GetPersonsParams) (api.GetPersonsRes, error) {
	limit := int32(10)
	if params.Limit.IsSet() {
		limit = int32(params.Limit.Value)
	}

	offset := int32(0)
	if params.Offset.IsSet() {
		offset = int32(params.Offset.Value)
	}

	// Get total count
	total, err := h.queries.CountPersons(ctx)
	if err != nil {
		zap.L().Error("error counting persons", zap.Error(err))
		return &api.Error{
			Message: "Failed to count people",
			Code:    "INTERNAL_ERROR",
		}, nil
	}

	// Get persons
	persons, err := h.queries.ListPersons(ctx, db.ListPersonsParams{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		zap.L().Error("error listing persons", zap.Error(err))
		return &api.Error{
			Message: "Failed to list people",
			Code:    "INTERNAL_ERROR",
		}, nil
	}

	// Convert to API response
	apiPersons := make([]api.Person, len(persons))
	for i, person := range persons {
		apiPersons[i] = api.Person{
			ID:        person.ID,
			Name:      person.Name,
			CreatedAt: person.CreatedAt,
			UpdatedAt: person.UpdatedAt,
		}
	}

	return &api.GetPersonsOKApplicationJSON{
		Persons: apiPersons,
		Total:   int(total),
	}, nil
}

func (h *PersonHandler) GetPersonsWithLastActivity(ctx context.Context, params api.GetPersonsParams) ([]templates.PersonWithLastActivity, error) {
	limit := int32(10)
	if params.Limit.IsSet() {
		limit = int32(params.Limit.Value)
	}

	offset := int32(0)
	if params.Offset.IsSet() {
		offset = int32(params.Offset.Value)
	}

	persons, err := h.queries.ListPersonsWithLastActivity(ctx, db.ListPersonsWithLastActivityParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		zap.L().Error("error listing persons with last activity", zap.Error(err))
		return nil, err
	}

	templatePersons := make([]templates.PersonWithLastActivity, len(persons))
	for i, person := range persons {
		tmpl := templates.PersonWithLastActivity{
			ID:        person.ID,
			Name:      person.Name,
			CreatedAt: person.CreatedAt,
			UpdatedAt: person.UpdatedAt,
		}
		if person.LastActionDesc != "" {
			tmpl.LastActionDesc = person.LastActionDesc
		}
		if !person.LastActionAt.IsZero() {
			t := person.LastActionAt
			tmpl.LastActionAt = &t
		}
		if person.LastConversationDesc != "" {
			tmpl.LastConversationDesc = person.LastConversationDesc
		}
		if !person.LastConversationAt.IsZero() {
			t := person.LastConversationAt
			tmpl.LastConversationAt = &t
		}
		templatePersons[i] = tmpl
	}

	return templatePersons, nil
}

func (h *PersonHandler) GetPersonTimeline(ctx context.Context, params api.GetPersonTimelineParams) (api.GetPersonTimelineRes, error) {
	// Verify person exists
	if _, err := h.queries.GetPersonByID(ctx, params.ID); err != nil {
		if err == sql.ErrNoRows {
			return &api.GetPersonTimelineNotFound{
				Message: "Person not found",
				Code:    "NOT_FOUND",
			}, nil
		}
		zap.L().Error("error getting person", zap.Error(err))
		return &api.GetPersonTimelineInternalServerError{
			Message: "Failed to get timeline",
			Code:    "INTERNAL_ERROR",
		}, nil
	}

	limit := int32(10)
	if params.Limit.IsSet() {
		limit = int32(params.Limit.Value)
	}

	offset := int32(0)
	if params.Offset.IsSet() {
		offset = int32(params.Offset.Value)
	}

	fetchLimit := limit + offset

	actions, err := h.queries.ListActionsByPersonID(ctx, db.ListActionsByPersonIDParams{
		PersonID: params.ID,
		Offset:   0,
		Limit:    fetchLimit,
	})
	if err != nil {
		zap.L().Error("error listing actions for timeline", zap.Error(err))
		return &api.GetPersonTimelineInternalServerError{
			Message: "Failed to get actions",
			Code:    "INTERNAL_ERROR",
		}, nil
	}

	conversations, err := h.queries.ListConversationsByPersonID(ctx, db.ListConversationsByPersonIDParams{
		PersonID: params.ID,
		Offset:   0,
		Limit:    fetchLimit,
	})
	if err != nil {
		zap.L().Error("error listing conversations for timeline", zap.Error(err))
		return &api.GetPersonTimelineInternalServerError{
			Message: "Failed to get conversations",
			Code:    "INTERNAL_ERROR",
		}, nil
	}

	var items []api.TimelineItem
	for _, a := range actions {
		act := a.Action
		item := api.TimelineItem{
			Type:        api.TimelineItemTypeAction,
			ID:          act.ID.String(),
			PersonID:    act.PersonID.String(),
			OccurredAt:  act.OccurredAt,
			Description: act.Description,
			CreatedAt:   act.CreatedAt,
			UpdatedAt:   act.UpdatedAt,
		}
		if act.References.Valid {
			item.References = api.OptNilString{Value: act.References.String, Set: true}
		}
		item.Valence = api.OptNilTimelineItemValence{Value: api.TimelineItemValence(act.Valence), Set: true}
		items = append(items, item)
	}

	for _, c := range conversations {
		convID, _ := xid.FromBytes(c.Conversation.ID)
		item := api.TimelineItem{
			Type:        api.TimelineItemTypeConversation,
			ID:          convID.String(),
			PersonID:    params.ID,
			OccurredAt:  c.Conversation.OccurredAt,
			Description: c.Conversation.Description,
			CreatedAt:   c.Conversation.CreatedAt,
			UpdatedAt:   c.Conversation.UpdatedAt,
		}
		items = append(items, item)
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].OccurredAt.After(items[j].OccurredAt)
	})

	totalActions, _ := h.queries.CountActionsByPersonID(ctx, params.ID)
	totalConversations, _ := h.queries.CountConversationsByPersonID(ctx, params.ID)
	total := int(totalActions + totalConversations)

	start := int(offset)
	if start > len(items) {
		start = len(items)
	}
	end := start + int(limit)
	if end > len(items) {
		end = len(items)
	}
	page := items[start:end]

	return &api.GetPersonTimelineOKApplicationJSON{
		Items: page,
		Total: total,
	}, nil
}

func (h *PersonHandler) UpdatePerson(ctx context.Context, req *api.UpdatePersonRequest, params api.UpdatePersonParams) (api.UpdatePersonRes, error) {
	// Validate request
	if req.Name == "" {
		return &api.UpdatePersonBadRequest{
			Message: "Name is required",
			Code:    "VALIDATION_ERROR",
		}, nil
	}

	person, err := h.queries.UpdatePerson(ctx, db.UpdatePersonParams{
		ID:   params.ID,
		Name: req.Name,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return &api.UpdatePersonNotFound{
				Message: "Person not found",
				Code:    "NOT_FOUND",
			}, nil
		}
		zap.L().Error("error updating person", zap.Error(err))
		return &api.UpdatePersonInternalServerError{
			Message: "Failed to update person",
			Code:    "INTERNAL_ERROR",
		}, nil
	}

	return &api.Person{
		ID:        person.ID,
		Name:      person.Name,
		CreatedAt: person.CreatedAt,
		UpdatedAt: person.UpdatedAt,
	}, nil
}

func (h *PersonHandler) DeletePerson(ctx context.Context, params api.DeletePersonParams) (api.DeletePersonRes, error) {
	err := h.queries.DeletePerson(ctx, params.ID)
	if err != nil {
		zap.L().Error("error deleting person", zap.Error(err))
		return &api.DeletePersonInternalServerError{
			Message: "Failed to delete person",
			Code:    "INTERNAL_ERROR",
		}, nil
	}

	return &api.DeletePersonNoContent{}, nil
}

func (h *PersonHandler) HandleGetPersonsForSelect(w http.ResponseWriter, r *http.Request) {
	// Call the API handler internally to get persons for the select dropdown
	params := api.GetPersonsParams{
		Limit:  api.OptInt{Value: 100, Set: true},
		Offset: api.OptInt{Value: 0, Set: true},
	}

	result, err := h.GetPersons(r.Context(), params)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		templates.PersonSelectError().Render(r.Context(), w)
		return
	}

	switch listResult := result.(type) {
	case *api.GetPersonsOKApplicationJSON:
		// Convert to template persons
		templatePersons := make([]templates.Person, len(listResult.Persons))
		for i, person := range listResult.Persons {
			templatePersons[i] = templates.Person{
				ID:        person.ID,
				Name:      person.Name,
				CreatedAt: person.CreatedAt,
				UpdatedAt: person.UpdatedAt,
			}
		}

		w.Header().Set("Content-Type", "text/html")
		templates.PersonSelectOptions(templatePersons).Render(r.Context(), w)
	default:
		w.WriteHeader(http.StatusInternalServerError)
		templates.PersonSelectError().Render(r.Context(), w)
	}
}
