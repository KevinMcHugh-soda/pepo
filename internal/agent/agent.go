package agent

import (
	"context"
	"encoding/json"
	"fmt"
)

// Conversation holds the accumulated history between the user and the agent.
type Conversation struct {
	messages []Message
}

// Messages returns a copy of the conversation history.
func (c *Conversation) Messages() []Message {
	out := make([]Message, len(c.messages))
	copy(out, c.messages)
	return out
}

// Append adds a message to the conversation history.
func (c *Conversation) Append(msg Message) {
	c.messages = append(c.messages, msg)
}

// Agent coordinates the interaction loop between an LLM and the registered tools.
type Agent struct {
	llm             LLM
	tools           *ToolRegistry
	convo           Conversation
	maxRounds       int
	initialMessages []Message
}

// AgentOption mutates configuration when constructing a new Agent.
type AgentOption func(*Agent)

// WithMaxRounds sets an upper bound on the number of LLM response iterations per input.
func WithMaxRounds(rounds int) AgentOption {
	return func(a *Agent) {
		a.maxRounds = rounds
	}
}

// WithSystemMessage adds a system-level instruction that is prepended to the conversation history.
func WithSystemMessage(content string) AgentOption {
	return func(a *Agent) {
		a.initialMessages = append(a.initialMessages, Message{Role: RoleSystem, Content: content})
	}
}

// WithInitialMessages seeds the conversation with a predefined set of messages.
func WithInitialMessages(messages ...Message) AgentOption {
	return func(a *Agent) {
		a.initialMessages = append(a.initialMessages, messages...)
	}
}

// New creates a new Agent instance.
func New(llm LLM, tools *ToolRegistry, opts ...AgentOption) *Agent {
	agent := &Agent{
		llm:       llm,
		tools:     tools,
		maxRounds: 6,
	}
	for _, opt := range opts {
		opt(agent)
	}
	agent.applyInitialMessages()
	return agent
}

func (a *Agent) applyInitialMessages() {
	if len(a.initialMessages) == 0 || len(a.convo.messages) > 0 {
		return
	}
	for _, msg := range a.initialMessages {
		a.convo.Append(msg)
	}
}

// Conversation returns a copy of the message history.
func (a *Agent) Conversation() []Message {
	return a.convo.Messages()
}

// Reset clears the conversation history while preserving initial messages.
func (a *Agent) Reset() {
	a.convo = Conversation{}
	a.applyInitialMessages()
}

// HandleUserMessage processes a new user message and returns the assistant's final response content.
func (a *Agent) HandleUserMessage(ctx context.Context, content string) (Message, error) {
	a.applyInitialMessages()
	a.convo.Append(Message{Role: RoleUser, Content: content})

	for round := 0; a.maxRounds == 0 || round < a.maxRounds; round++ {
		var toolDefs []ToolDefinition
		if a.tools != nil {
			toolDefs = a.tools.DefinitionList()
		}
		response, err := a.llm.Generate(ctx, ChatParams{
			Messages: a.convo.Messages(),
			Tools:    toolDefs,
		})
		if err != nil {
			return Message{}, fmt.Errorf("llm generate: %w", err)
		}

		// Append the assistant message regardless of tool usage for auditability.
		assistantMsg := response.Message
		if assistantMsg.Role == "" {
			assistantMsg.Role = RoleAssistant
		}
		a.convo.Append(assistantMsg)

		if len(response.ToolCalls) == 0 {
			return assistantMsg, nil
		}

		for idx, call := range response.ToolCalls {
			toolCallID := call.ID
			if toolCallID == "" {
				toolCallID = fmt.Sprintf("%d", idx)
			}

			tool, ok := a.tools.Get(call.Name)
			if !ok {
				// Surface an error back to the LLM so it can recover.
				errMsg := fmt.Sprintf("unknown tool %q", call.Name)
				a.convo.Append(Message{
					Role:       RoleTool,
					Name:       call.Name,
					Content:    errMsg,
					ToolCallID: toolCallID,
				})
				continue
			}

			toolResp, err := tool.Call(ctx, call.Arguments)
			content := string(toolResp)
			if err != nil {
				content = fmt.Sprintf("tool error: %v", err)
			}

			a.convo.Append(Message{
				Role:       RoleTool,
				Name:       call.Name,
				Content:    content,
				ToolCallID: toolCallID,
			})
		}
	}

	return Message{}, fmt.Errorf("max rounds exceeded")
}

// MarshalConversation returns the conversation history as JSON for persistence or debugging purposes.
func (a *Agent) MarshalConversation() ([]byte, error) {
	messages := a.convo.Messages()
	return json.Marshal(messages)
}
