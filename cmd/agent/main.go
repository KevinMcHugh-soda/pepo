package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"

	"pepo/internal/agent"
	"pepo/internal/agent/tools"
)

func main() {
	var (
		apiBaseURL   = flag.String("api-base-url", getEnv("PEPO_API_BASE_URL", "http://localhost:8080"), "Base URL for the Pepo API")
		model        = flag.String("model", getEnv("OPENAI_MODEL", ""), "OpenAI model to use for chat completions")
		systemPrompt = flag.String("system", getEnv("PEPO_AGENT_SYSTEM", "You are Pepo, an assistant that helps managers understand their team's performance."), "System prompt for the agent")
		maxRounds    = flag.Int("max-rounds", getEnvInt("PEPO_AGENT_MAX_ROUNDS", 6), "Maximum number of tool/response rounds per user message (0 = unlimited)")
		timeout      = flag.Duration("timeout", getEnvDuration("PEPO_AGENT_TIMEOUT", 2*time.Minute), "Timeout for each LLM request")
	)
	flag.Parse()

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "OPENAI_API_KEY environment variable is required")
		os.Exit(1)
	}

	ctx := context.Background()

	openaiClient := openai.NewClient(apiKey)
	llmOpts := make([]agent.OpenAIOption, 0, 1)
	if *model != "" {
		llmOpts = append(llmOpts, agent.WithOpenAIModel(*model))
	}
	llm := agent.NewOpenAIClient(openaiClient, llmOpts...)

	apiTool, err := tools.NewAPITool(*apiBaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create API tool: %v\n", err)
		os.Exit(1)
	}

	toolRegistry := agent.NewToolRegistry(apiTool)
	agentOpts := []agent.AgentOption{
		agent.WithSystemMessage(*systemPrompt),
	}
	if *maxRounds >= 0 {
		agentOpts = append(agentOpts, agent.WithMaxRounds(*maxRounds))
	}

	pepoAgent := agent.New(llm, toolRegistry, agentOpts...)

	reader := bufio.NewScanner(os.Stdin)
	reader.Buffer(make([]byte, 0, 1024), 1024*1024)

	fmt.Println("Starting Pepo agent conversation. Type 'exit' or 'quit' to end.")

	for {
		fmt.Print("You: ")
		if !reader.Scan() {
			fmt.Println()
			break
		}
		input := strings.TrimSpace(reader.Text())
		if input == "" {
			continue
		}
		lower := strings.ToLower(input)
		if lower == "exit" || lower == "quit" {
			fmt.Println("Goodbye!")
			break
		}

		reqCtx, cancel := context.WithTimeout(ctx, *timeout)
		response, err := pepoAgent.HandleUserMessage(reqCtx, input)
		cancel()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			continue
		}

		fmt.Printf("Agent: %s\n", strings.TrimSpace(response.Content))
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		var parsed int
		if _, err := fmt.Sscanf(value, "%d", &parsed); err == nil {
			return parsed
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return fallback
}
