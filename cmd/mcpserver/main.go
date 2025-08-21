package main

import (
	"context"
	"fmt"
	"log"
	"time"

	mcp "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"

	"pepo/internal/config"
	"pepo/internal/database"
	"pepo/internal/db"
)

type ListActionsRequest struct {
	PersonID string `json:"person_id" jsonschema_description:"ID of the person" jsonschema:"required"`
	Limit    int    `json:"limit,omitempty" jsonschema_description:"Maximum number of actions to return" jsonschema:"minimum=1,maximum=100,default=10"`
	Offset   int    `json:"offset,omitempty" jsonschema_description:"Pagination offset" jsonschema:"minimum=0,default=0"`
}

type Action struct {
	ID          string    `json:"id"`
	PersonID    string    `json:"person_id"`
	OccurredAt  time.Time `json:"occurred_at"`
	Description string    `json:"description"`
	References  string    `json:"references,omitempty"`
	Valence     string    `json:"valence"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func main() {
	cfg := config.Load()

	dbConn, queries, err := database.Initialize(cfg.DatabaseURL, database.DefaultConnectionConfig())
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer func() {
		if err := database.Close(dbConn); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	s := mcpserver.NewMCPServer(
		"Pepo MCP Server",
		"1.0.0",
		mcpserver.WithToolCapabilities(false),
	)

	listTool := mcp.NewTool(
		"list_actions_by_person",
		mcp.WithDescription("List actions for a specific person"),
		mcp.WithInputSchema[ListActionsRequest](),
		mcp.WithOutputSchema[[]Action](),
	)

	s.AddTool(listTool, mcp.NewStructuredToolHandler(func(ctx context.Context, request mcp.CallToolRequest, args ListActionsRequest) ([]Action, error) {
		limit := args.Limit
		if limit <= 0 {
			limit = 10
		}
		offset := args.Offset
		if offset < 0 {
			offset = 0
		}

		rows, err := queries.ListActionsByPersonID(ctx, db.ListActionsByPersonIDParams{
			PersonID: args.PersonID,
			Offset:   int32(offset),
			Limit:    int32(limit),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list actions: %w", err)
		}

		actions := make([]Action, 0, len(rows))
		for _, row := range rows {
			a := row.Action
			action := Action{
				ID:          a.ID.String(),
				PersonID:    a.PersonID.String(),
				OccurredAt:  a.OccurredAt,
				Description: a.Description,
				Valence:     string(a.Valence),
				CreatedAt:   a.CreatedAt,
				UpdatedAt:   a.UpdatedAt,
			}
			if a.References.Valid {
				action.References = a.References.String
			}
			actions = append(actions, action)
		}
		return actions, nil
	}))

	if err := mcpserver.ServeStdio(s); err != nil {
		log.Printf("Server error: %v", err)
	}
}
