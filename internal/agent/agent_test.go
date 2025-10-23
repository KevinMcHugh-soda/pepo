package agent

import (
	"context"
	"encoding/json"
	"testing"
)

type mockLLM struct {
	responses []ChatResponse
	calls     int
}

func (m *mockLLM) Generate(ctx context.Context, params ChatParams) (ChatResponse, error) {
	if m.calls >= len(m.responses) {
		return ChatResponse{}, nil
	}
	resp := m.responses[m.calls]
	m.calls++
	return resp, nil
}

type mockTool struct {
	definition ToolDefinition
	output     json.RawMessage
}

func (m *mockTool) Definition() ToolDefinition {
	return m.definition
}

func (m *mockTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	if len(m.output) > 0 {
		return m.output, nil
	}
	return input, nil
}

func TestHandleUserMessage_NoTool(t *testing.T) {
	llm := &mockLLM{responses: []ChatResponse{
		{
			Message: Message{Role: RoleAssistant, Content: "Hello!"},
		},
	}}
	agent := New(llm, NewToolRegistry())

	msg, err := agent.HandleUserMessage(context.Background(), "Hi")
	if err != nil {
		t.Fatalf("HandleUserMessage returned error: %v", err)
	}
	if msg.Content != "Hello!" {
		t.Fatalf("expected response content to be %q, got %q", "Hello!", msg.Content)
	}
	if llm.calls != 1 {
		t.Fatalf("expected llm to be called once, got %d", llm.calls)
	}
}

func TestHandleUserMessage_ToolInvocation(t *testing.T) {
	toolResponse := json.RawMessage(`{"result":"ok"}`)
	llm := &mockLLM{responses: []ChatResponse{
		{
			Message: Message{
				Role:       RoleAssistant,
				Content:    "",
				ToolCallID: "call-1",
				ToolCall: &ToolCall{
					ID:        "call-1",
					Name:      "call_api",
					Arguments: json.RawMessage(`{"method":"GET","path":"/people"}`),
				},
			},
			ToolCalls: []ToolCall{
				{
					ID:        "call-1",
					Name:      "call_api",
					Arguments: json.RawMessage(`{"method":"GET","path":"/people"}`),
				},
			},
		},
		{
			Message: Message{Role: RoleAssistant, Content: "Done"},
		},
	}}

	tool := &mockTool{
		definition: ToolDefinition{
			Name:        "call_api",
			Description: "Call the API",
			Parameters:  json.RawMessage(`{}`),
		},
		output: toolResponse,
	}

	agent := New(llm, NewToolRegistry(tool))

	msg, err := agent.HandleUserMessage(context.Background(), "Do something")
	if err != nil {
		t.Fatalf("HandleUserMessage returned error: %v", err)
	}
	if msg.Content != "Done" {
		t.Fatalf("expected final response to be %q, got %q", "Done", msg.Content)
	}
	if llm.calls != 2 {
		t.Fatalf("expected llm to be called twice, got %d", llm.calls)
	}

	convo := agent.Conversation()
	if len(convo) != 4 {
		t.Fatalf("expected 4 conversation entries, got %d", len(convo))
	}
	toolMsg := convo[2]
	if toolMsg.Role != RoleTool {
		t.Fatalf("expected tool message role to be %q, got %q", RoleTool, toolMsg.Role)
	}
	if toolMsg.ToolCallID != "call-1" {
		t.Fatalf("expected tool call id to be %q, got %q", "call-1", toolMsg.ToolCallID)
	}
	if toolMsg.Content != string(toolResponse) {
		t.Fatalf("expected tool response content to match, got %q", toolMsg.Content)
	}
}
