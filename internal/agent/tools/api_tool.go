package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"pepo/internal/agent"
)

const (
	apiToolLogBodyLimit   = 2048
	apiAllowedPathPattern = `^/(people(?:/[A-Za-z0-9_-]+(?:/(actions|timeline))?)?|actions(?:/[A-Za-z0-9_-]+)?|conversations(?:/[A-Za-z0-9_-]+)?)$`
)

type endpointRule struct {
	pattern *regexp.Regexp
	methods map[string]struct{}
}

var apiToolParameters = json.RawMessage([]byte(fmt.Sprintf(`{
  "type": "object",
  "properties": {
    "method": {
      "type": "string",
      "description": "HTTP method to use when calling the API",
      "enum": ["GET", "POST", "PUT", "DELETE"]
    },
    "path": {
      "type": "string",
      "description": "Path of the API endpoint, relative to the API base URL. Allowed endpoints include /people, /people/{id}, /people/{id}/actions, /people/{id}/timeline, /actions, /actions/{id}, /conversations, /conversations/{id}. Substitute actual resource IDs for {id}.",
      "pattern": %q,
      "examples": ["/people", "/people/ab12cd34ef56gh78ij90", "/people/ab12cd34ef56gh78ij90/actions"]
    },
    "query": {
      "type": "object",
      "description": "Optional query string parameters",
      "additionalProperties": {"type": "string"}
    },
    "headers": {
      "type": "object",
      "description": "Additional HTTP headers to include in the request",
      "additionalProperties": {"type": "string"}
    },
    "body": {
      "description": "Optional JSON body to send with the request",
      "anyOf": [
        {"type": "object"},
        {"type": "array", "items": {}},
        {"type": "string"},
        {"type": "number"},
        {"type": "boolean"},
        {"type": "null"}
      ]
    }
  },
  "required": ["method", "path"],
  "additionalProperties": false
}`, apiAllowedPathPattern)))

var allowedEndpointRules = []endpointRule{
	{
		pattern: regexp.MustCompile(`^/people$`),
		methods: methodSet("GET", "POST"),
	},
	{
		pattern: regexp.MustCompile(`^/people/[A-Za-z0-9_-]+$`),
		methods: methodSet("GET", "PUT", "DELETE"),
	},
	{
		pattern: regexp.MustCompile(`^/people/[A-Za-z0-9_-]+/actions$`),
		methods: methodSet("GET"),
	},
	{
		pattern: regexp.MustCompile(`^/people/[A-Za-z0-9_-]+/timeline$`),
		methods: methodSet("GET"),
	},
	{
		pattern: regexp.MustCompile(`^/actions$`),
		methods: methodSet("GET", "POST"),
	},
	{
		pattern: regexp.MustCompile(`^/actions/[A-Za-z0-9_-]+$`),
		methods: methodSet("GET", "PUT", "DELETE"),
	},
	{
		pattern: regexp.MustCompile(`^/conversations$`),
		methods: methodSet("POST"),
	},
	{
		pattern: regexp.MustCompile(`^/conversations/[A-Za-z0-9_-]+$`),
		methods: methodSet("GET", "PUT", "DELETE"),
	},
}

func methodSet(methods ...string) map[string]struct{} {
	set := make(map[string]struct{}, len(methods))
	for _, method := range methods {
		set[strings.ToUpper(method)] = struct{}{}
	}
	return set
}

func validateAllowedEndpoint(method, path string) error {
	method = strings.ToUpper(method)
	for _, rule := range allowedEndpointRules {
		if rule.pattern.MatchString(path) {
			if _, ok := rule.methods[method]; ok {
				return nil
			}
			return fmt.Errorf("method %q is not allowed for endpoint %q", method, path)
		}
	}
	return fmt.Errorf("path %q is not an allowed API endpoint", path)
}

// APITool allows the agent to communicate with the Pepo HTTP API using structured tool calls.
type APITool struct {
	name           string
	description    string
	client         *http.Client
	baseURL        *url.URL
	defaultHeaders http.Header
}

// APIToolOption configures the API tool.
type APIToolOption func(*APITool)

// WithName changes the function name exposed to the LLM.
func WithName(name string) APIToolOption {
	return func(t *APITool) {
		t.name = name
	}
}

// WithDescription overrides the tool description presented to the LLM.
func WithDescription(desc string) APIToolOption {
	return func(t *APITool) {
		t.description = desc
	}
}

// WithHTTPClient provides a custom http.Client for outbound requests.
func WithHTTPClient(client *http.Client) APIToolOption {
	return func(t *APITool) {
		t.client = client
	}
}

// WithDefaultHeader sets a header that will be included on every request made by the tool.
func WithDefaultHeader(key, value string) APIToolOption {
	return func(t *APITool) {
		if t.defaultHeaders == nil {
			t.defaultHeaders = make(http.Header)
		}
		t.defaultHeaders.Set(key, value)
	}
}

// NewAPITool constructs a new tool instance configured to call the application's API.
func NewAPITool(baseURL string, opts ...APIToolOption) (*APITool, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse base url: %w", err)
	}

	tool := &APITool{
		name:        "call_api",
		description: "Call the Pepo HTTP API using REST semantics.",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: parsed,
	}

	for _, opt := range opts {
		opt(tool)
	}

	if tool.client == nil {
		tool.client = &http.Client{Timeout: 30 * time.Second}
	}

	return tool, nil
}

func (t *APITool) Definition() agent.ToolDefinition {
	return agent.ToolDefinition{
		Name:        t.name,
		Description: t.description,
		Parameters:  apiToolParameters,
	}
}

func (t *APITool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	if t.baseURL == nil {
		return nil, fmt.Errorf("api tool missing base url")
	}

	var payload struct {
		Method  string            `json:"method"`
		Path    string            `json:"path"`
		Query   map[string]string `json:"query"`
		Headers map[string]string `json:"headers"`
		Body    json.RawMessage   `json:"body"`
	}

	if err := json.Unmarshal(input, &payload); err != nil {
		return nil, fmt.Errorf("unmarshal tool input: %w", err)
	}

	if payload.Method == "" || payload.Path == "" {
		return nil, fmt.Errorf("method and path are required")
	}

	payload.Method = strings.ToUpper(payload.Method)

	relPath := payload.Path
	if !strings.HasPrefix(relPath, "/") {
		relPath = "/" + relPath
	}

	if err := validateAllowedEndpoint(payload.Method, relPath); err != nil {
		return nil, err
	}

	endpoint, err := t.baseURL.Parse(relPath)
	if err != nil {
		return nil, fmt.Errorf("build url: %w", err)
	}

	if len(payload.Query) > 0 {
		q := endpoint.Query()
		for key, value := range payload.Query {
			q.Set(key, value)
		}
		endpoint.RawQuery = q.Encode()
	}

	var body io.Reader
	hasBody := len(payload.Body) > 0 && string(bytes.TrimSpace(payload.Body)) != "null"
	if hasBody {
		body = bytes.NewReader(payload.Body)
	}

	logAPIToolRequest(payload.Method, endpoint.String(), hasBody, payload.Body)

	req, err := http.NewRequestWithContext(ctx, payload.Method, endpoint.String(), body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	for key, values := range t.defaultHeaders {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	for key, value := range payload.Headers {
		req.Header.Set(key, value)
	}

	if hasBody && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	start := time.Now()
	resp, err := t.client.Do(req)
	if err != nil {
		logAPIToolError(err)
		return nil, fmt.Errorf("perform request: %w", err)
	}
	duration := time.Since(start)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	logAPIToolResponse(resp.StatusCode, duration, respBody)

	result := map[string]any{
		"status":  resp.StatusCode,
		"headers": resp.Header,
	}

	trimmed := bytes.TrimSpace(respBody)
	if len(trimmed) == 0 {
		result["body"] = ""
	} else if json.Valid(trimmed) {
		var decoded any
		if err := json.Unmarshal(trimmed, &decoded); err == nil {
			result["body"] = decoded
		} else {
			result["body"] = string(respBody)
		}
	} else {
		result["body"] = string(respBody)
	}

	return json.Marshal(result)
}

func logAPIToolRequest(method, endpoint string, hasBody bool, body []byte) {
	if method == "" {
		method = "GET"
	}
	fmt.Fprintf(os.Stderr, "[pepo-agent] call_api -> %s %s\n", strings.ToUpper(method), endpoint)
	if hasBody {
		logAPIToolBody("call_api request body", body)
	}
}

func logAPIToolResponse(status int, duration time.Duration, body []byte) {
	statusText := http.StatusText(status)
	if statusText == "" {
		statusText = "Unknown"
	}
	fmt.Fprintf(os.Stderr, "[pepo-agent] call_api <- %d %s (%s)\n", status, statusText, duration.Truncate(time.Millisecond))
	logAPIToolBody("call_api response body", body)
}

func logAPIToolError(err error) {
	fmt.Fprintf(os.Stderr, "[pepo-agent] call_api error: %v\n", err)
}

func logAPIToolBody(label string, raw []byte) {
	formatted := formatAPIToolBody(raw)
	if formatted == "" {
		return
	}
	if strings.Contains(formatted, "\n") {
		fmt.Fprintf(os.Stderr, "[pepo-agent] %s:\n%s\n", label, formatted)
		return
	}
	fmt.Fprintf(os.Stderr, "[pepo-agent] %s: %s\n", label, formatted)
}

func formatAPIToolBody(raw []byte) string {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return ""
	}
	if first := trimmed[0]; first == '{' || first == '[' {
		var buf bytes.Buffer
		if err := json.Indent(&buf, trimmed, "", "  "); err == nil {
			return truncateForLog(buf.String(), apiToolLogBodyLimit)
		}
	}
	return truncateForLog(string(trimmed), apiToolLogBodyLimit)
}

func truncateForLog(text string, limit int) string {
	if limit <= 0 {
		return ""
	}
	runeCount := 0
	for idx := range text {
		if runeCount == limit {
			return text[:idx] + "...(truncated)"
		}
		runeCount++
	}
	return text
}
