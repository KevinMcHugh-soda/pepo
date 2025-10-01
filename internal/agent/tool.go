package agent

import (
	"context"
	"encoding/json"
	"sort"
)

// Tool is implemented by capabilities that the agent can invoke during a conversation.
type Tool interface {
	Definition() ToolDefinition
	Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error)
}

// ToolRegistry manages lookup for registered tools by name.
type ToolRegistry struct {
	tools map[string]Tool
}

// NewToolRegistry creates a registry pre-populated with the provided tools.
func NewToolRegistry(toolList ...Tool) *ToolRegistry {
	reg := &ToolRegistry{tools: make(map[string]Tool, len(toolList))}
	for _, t := range toolList {
		def := t.Definition()
		reg.tools[def.Name] = t
	}
	return reg
}

// Add registers a tool and overwrites existing registrations with the same name.
func (r *ToolRegistry) Add(tool Tool) {
	if r == nil {
		return
	}
	if r.tools == nil {
		r.tools = make(map[string]Tool)
	}
	def := tool.Definition()
	r.tools[def.Name] = tool
}

// DefinitionList returns the tool definitions sorted by name for deterministic ordering.
func (r *ToolRegistry) DefinitionList() []ToolDefinition {
	if r == nil || len(r.tools) == 0 {
		return nil
	}
	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	sort.Strings(names)

	defs := make([]ToolDefinition, 0, len(r.tools))
	for _, name := range names {
		defs = append(defs, r.tools[name].Definition())
	}
	return defs
}

// Get returns a tool by name.
func (r *ToolRegistry) Get(name string) (Tool, bool) {
	if r == nil {
		return nil, false
	}
	tool, ok := r.tools[name]
	return tool, ok
}
