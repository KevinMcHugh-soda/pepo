package agent

import (
	"context"
	"encoding/json"
)

// ToolDefinition describes a callable tool that can be invoked by the LLM.
type ToolDefinition struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

// ChatParams wraps the inputs required to generate a chat completion from an LLM.
type ChatParams struct {
	Messages []Message
	Tools    []ToolDefinition
}

// ChatResponse encapsulates the assistant output from the LLM.
type ChatResponse struct {
	Message   Message
	ToolCalls []ToolCall
}

// LLM defines the interface that must be implemented to plug a model provider into the agent.
type LLM interface {
	Generate(ctx context.Context, params ChatParams) (ChatResponse, error)
}
