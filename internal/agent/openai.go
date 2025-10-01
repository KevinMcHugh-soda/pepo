package agent

import (
	"context"
	"encoding/json"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

// OpenAIClient wraps the go-openai client to implement the LLM interface.
type OpenAIClient struct {
	client *openai.Client
	model  string
}

// OpenAIOption configures an OpenAIClient instance.
type OpenAIOption func(*OpenAIClient)

// WithOpenAIModel overrides the default model used for chat completions.
func WithOpenAIModel(model string) OpenAIOption {
	return func(c *OpenAIClient) {
		c.model = model
	}
}

// NewOpenAIClient constructs an OpenAI-backed LLM implementation.
func NewOpenAIClient(client *openai.Client, opts ...OpenAIOption) *OpenAIClient {
	llm := &OpenAIClient{
		client: client,
		model:  "gpt-4o-mini",
	}
	for _, opt := range opts {
		opt(llm)
	}
	return llm
}

// Generate requests a chat completion from OpenAI using the provided conversation state and tool definitions.
func (c *OpenAIClient) Generate(ctx context.Context, params ChatParams) (ChatResponse, error) {
	if c.client == nil {
		return ChatResponse{}, fmt.Errorf("openai client is nil")
	}

	messages := make([]openai.ChatCompletionMessage, 0, len(params.Messages))
	for _, msg := range params.Messages {
		messages = append(messages, toOpenAIMessage(msg))
	}

	req := openai.ChatCompletionRequest{
		Model:       c.model,
		Messages:    messages,
		Temperature: 0,
	}

	if len(params.Tools) > 0 {
		req.Tools = make([]openai.Tool, 0, len(params.Tools))
		for _, tool := range params.Tools {
			req.Tools = append(req.Tools, openai.Tool{
				Type: openai.ToolTypeFunction,
				Function: &openai.FunctionDefinition{
					Name:        tool.Name,
					Description: tool.Description,
					Parameters:  tool.Parameters,
				},
			})
		}
	}

	resp, err := c.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("create chat completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return ChatResponse{}, fmt.Errorf("no choices returned from openai")
	}

	choice := resp.Choices[0]
	message := fromOpenAIMessage(choice.Message)
	toolCalls := make([]ToolCall, 0, len(choice.Message.ToolCalls))
	for _, call := range choice.Message.ToolCalls {
		toolCalls = append(toolCalls, ToolCall{
			ID:        call.ID,
			Name:      call.Function.Name,
			Arguments: json.RawMessage(call.Function.Arguments),
		})
	}

	return ChatResponse{
		Message:   message,
		ToolCalls: toolCalls,
	}, nil
}

func toOpenAIMessage(msg Message) openai.ChatCompletionMessage {
	oaMsg := openai.ChatCompletionMessage{
		Role:    string(msg.Role),
		Content: msg.Content,
		Name:    msg.Name,
	}

	if msg.ToolCallID != "" {
		oaMsg.ToolCallID = msg.ToolCallID
	}

	if msg.ToolCall != nil {
		call := openai.ToolCall{
			ID:   msg.ToolCall.ID,
			Type: openai.ToolTypeFunction,
			Function: openai.FunctionCall{
				Name:      msg.ToolCall.Name,
				Arguments: string(msg.ToolCall.Arguments),
			},
		}
		oaMsg.ToolCalls = []openai.ToolCall{call}
	}

	return oaMsg
}

func fromOpenAIMessage(msg openai.ChatCompletionMessage) Message {
	out := Message{
		Role:    Role(msg.Role),
		Content: msg.Content,
		Name:    msg.Name,
	}
	if len(msg.ToolCalls) > 0 {
		// Tool calls are stored separately in ChatResponse but we keep the first call on the message for completeness.
		call := msg.ToolCalls[0]
		out.ToolCallID = call.ID
		out.ToolCall = &ToolCall{
			ID:        call.ID,
			Name:      call.Function.Name,
			Arguments: json.RawMessage(call.Function.Arguments),
		}
	}
	return out
}
