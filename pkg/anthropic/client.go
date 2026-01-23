package anthropic

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// truncateJSONForLog truncates long text fields in JSON for logging
// Specifically truncates system prompts, message text, and response text to maxLen characters
func truncateJSONForLog(jsonData []byte, maxLen int) string {
	var data map[string]any
	if err := json.Unmarshal(jsonData, &data); err != nil {
		// If we can't parse, just return as-is
		return string(jsonData)
	}

	// Truncate system prompt (can be string or array of blocks)
	if system, ok := data["system"]; ok {
		if systemStr, ok := system.(string); ok && len(systemStr) > maxLen {
			data["system"] = systemStr[:maxLen] + "..."
		} else if systemBlocks, ok := system.([]any); ok {
			// Handle system as array of blocks
			for i, block := range systemBlocks {
				if blockMap, ok := block.(map[string]any); ok {
					if text, ok := blockMap["text"].(string); ok && len(text) > maxLen {
						blockMap["text"] = text[:maxLen] + "..."
						systemBlocks[i] = blockMap
					}
				}
			}
			data["system"] = systemBlocks
		}
	}

	// Truncate messages array (request messages)
	if messages, ok := data["messages"].([]any); ok {
		for i, msg := range messages {
			if msgMap, ok := msg.(map[string]any); ok {
				// Handle content as string
				if content, ok := msgMap["content"].(string); ok && len(content) > maxLen {
					msgMap["content"] = content[:maxLen] + "..."
					messages[i] = msgMap
				}
				// Handle content as array of blocks
				if contentBlocks, ok := msgMap["content"].([]any); ok {
					for j, block := range contentBlocks {
						if blockMap, ok := block.(map[string]any); ok {
							if text, ok := blockMap["text"].(string); ok && len(text) > maxLen {
								blockMap["text"] = text[:maxLen] + "..."
								contentBlocks[j] = blockMap
							}
							// Also truncate tool_result content
							if content, ok := blockMap["content"].(string); ok && len(content) > maxLen {
								blockMap["content"] = content[:maxLen] + "..."
								contentBlocks[j] = blockMap
							}
						}
					}
					msgMap["content"] = contentBlocks
					messages[i] = msgMap
				}
			}
		}
		data["messages"] = messages
	}

	// Truncate content blocks (response text)
	if content, ok := data["content"].([]any); ok {
		for i, block := range content {
			if blockMap, ok := block.(map[string]any); ok {
				if text, ok := blockMap["text"].(string); ok && len(text) > maxLen {
					blockMap["text"] = text[:maxLen] + "..."
					content[i] = blockMap
				}
			}
		}
		data["content"] = content
	}

	// Re-encode to JSON
	truncated, err := json.Marshal(data)
	if err != nil {
		return string(jsonData)
	}
	return string(truncated)
}

const (
	apiURL     = "https://api.anthropic.com/v1/messages"
	modelsURL  = "https://api.anthropic.com/v1/models"
	apiVersion = "2023-06-01"
)

// Client handles communication with the Anthropic API
type Client struct {
	apiKey     string
	httpClient *http.Client
	verbose    bool
}

// NewClient creates a new Anthropic API client
func NewClient(apiKey string, verbose bool) *Client {
	return &Client{
		apiKey:  apiKey,
		verbose: verbose,
		httpClient: &http.Client{
			Timeout: 15 * time.Minute, // Allow up to 15 minutes for large responses with continuation
		},
	}
}

// ValidateAPIKey checks if the API key has the correct format
func ValidateAPIKey(apiKey string) error {
	if apiKey == "" {
		return fmt.Errorf("API key is empty")
	}
	if len(apiKey) < 10 {
		return fmt.Errorf("API key is too short")
	}
	if apiKey[:7] != "sk-ant-" {
		return fmt.Errorf("API key should start with 'sk-ant-'")
	}
	return nil
}

// Message represents a message in the conversation
type Message struct {
	Role    string      `json:"role"`    // "user" or "assistant"
	Content any `json:"content"` // String for regular messages, []ContentBlock for tool results
}

// Tool represents a tool that Claude can use
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]any `json:"input_schema"`
}

// CacheControl represents cache control settings for prompt caching
type CacheControl struct {
	Type string `json:"type"` // "ephemeral"
}

// SystemBlock represents a block in the system prompt with optional caching
type SystemBlock struct {
	Type         string        `json:"type"` // "text"
	Text         string        `json:"text"`
	CacheControl *CacheControl `json:"cache_control,omitempty"`
}

// CreateMessageRequest represents the request to create a message
type CreateMessageRequest struct {
	Model       string      `json:"model"`
	MaxTokens   int         `json:"max_tokens,omitempty"` // omitempty: if 0, field is omitted from JSON
	Messages    []Message   `json:"messages"`
	System      any `json:"system,omitempty"` // string or []SystemBlock
	Temperature float64     `json:"temperature,omitempty"`
	Stream      bool        `json:"stream,omitempty"`
	Tools       []Tool      `json:"tools,omitempty"`
}

// ContentBlock represents a content block in the response
type ContentBlock struct {
	Type         string                 `json:"type"` // "text", "tool_use", or "tool_result"
	Text         string                 `json:"text,omitempty"`
	ID           string                 `json:"id,omitempty"`            // For tool_use blocks
	Name         string                 `json:"name,omitempty"`          // Tool name
	Input        map[string]any `json:"input,omitempty"`         // Tool input
	ToolUseID    string                 `json:"tool_use_id,omitempty"`   // For tool_result blocks
	Content      string                 `json:"content,omitempty"`       // For tool_result blocks
	IsError      bool                   `json:"is_error,omitempty"`      // For tool_result blocks
	CacheControl *CacheControl          `json:"cache_control,omitempty"` // For prompt caching
}

// MarshalJSON implements custom JSON marshaling for ContentBlock
// Ensures tool_use blocks always include "input" field (even if empty) to satisfy API requirements
func (c ContentBlock) MarshalJSON() ([]byte, error) {
	// Use type alias to avoid recursion
	type Alias ContentBlock

	// For tool_use blocks, ensure Input is always present
	if c.Type == "tool_use" {
		if c.Input == nil {
			c.Input = make(map[string]any)
		}
		// Create anonymous struct with Input field that doesn't have omitempty
		return json.Marshal(&struct {
			Type         string                 `json:"type"`
			Text         string                 `json:"text,omitempty"`
			ID           string                 `json:"id,omitempty"`
			Name         string                 `json:"name,omitempty"`
			Input        map[string]any `json:"input"` // No omitempty!
			ToolUseID    string                 `json:"tool_use_id,omitempty"`
			Content      string                 `json:"content,omitempty"`
			IsError      bool                   `json:"is_error,omitempty"`
			CacheControl *CacheControl          `json:"cache_control,omitempty"`
		}{
			Type:         c.Type,
			Text:         c.Text,
			ID:           c.ID,
			Name:         c.Name,
			Input:        c.Input,
			ToolUseID:    c.ToolUseID,
			Content:      c.Content,
			IsError:      c.IsError,
			CacheControl: c.CacheControl,
		})
	}

	// For other types, use default marshaling with omitempty
	return json.Marshal((*Alias)(&c))
}

// CreateMessageResponse represents the response from the API
type CreateMessageResponse struct {
	ID           string         `json:"id"`
	Type         string         `json:"type"`
	Role         string         `json:"role"`
	Content      []ContentBlock `json:"content"`
	Model        string         `json:"model"`
	StopReason   string         `json:"stop_reason"`
	StopSequence string         `json:"stop_sequence"`
	Usage        struct {
		InputTokens              int `json:"input_tokens"`
		OutputTokens             int `json:"output_tokens"`
		CacheCreationInputTokens int `json:"cache_creation_input_tokens,omitempty"`
		CacheReadInputTokens     int `json:"cache_read_input_tokens,omitempty"`
	} `json:"usage"`
}

// ErrorResponse represents an error response from the API
type ErrorResponse struct {
	Type  string `json:"type"`
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// RateLimitError represents a rate limit error with retry information
type RateLimitError struct {
	Message    string
	RetryAfter int // seconds to wait before retrying
	StatusCode int
}

func (e *RateLimitError) Error() string {
	if e.RetryAfter > 0 {
		return fmt.Sprintf("Rate limit exceeded. Please retry after %d seconds", e.RetryAfter)
	}
	return "Rate limit exceeded. Please wait before retrying"
}

// ContextLengthError represents a context length exceeded error
type ContextLengthError struct {
	Message       string
	TokensUsed    int
	TokensMaximum int
	StatusCode    int
}

func (e *ContextLengthError) Error() string {
	if e.TokensUsed > 0 && e.TokensMaximum > 0 {
		return fmt.Sprintf("Context length exceeded: %d tokens used, %d maximum", e.TokensUsed, e.TokensMaximum)
	}
	return "Context length exceeded. The conversation is too long."
}

// parseRetryAfter extracts the retry delay from the Retry-After header
// Returns 0 if header is not present or invalid
func parseRetryAfter(header string) int {
	if header == "" {
		return 0
	}

	// Try parsing as integer (seconds)
	if seconds, err := strconv.Atoi(header); err == nil {
		return seconds
	}

	// Could also parse HTTP-date format, but for simplicity we'll just use default
	return 0
}

// parseContextLengthError extracts token counts from context length error message
// Expected format: "prompt is too long: 202838 tokens > 200000 maximum"
func parseContextLengthError(message string) (tokensUsed int, tokensMax int) {
	re := regexp.MustCompile(`(\d+)\s+tokens?\s*>\s*(\d+)\s+maximum`)
	matches := re.FindStringSubmatch(message)
	if len(matches) >= 3 {
		tokensUsed, _ = strconv.Atoi(matches[1])
		tokensMax, _ = strconv.Atoi(matches[2])
	}
	return
}

// CreateMessage sends a message to Claude and returns the response
// The context can be used to cancel the request
func (c *Client) CreateMessage(ctx context.Context, req CreateMessageRequest) (*CreateMessageResponse, error) {
	// If MaxTokens is 0, it will be omitted from the JSON request (omitempty)
	// allowing the API to use its own defaults/limits

	// Validate model
	if req.Model == "" {
		return nil, fmt.Errorf("model is required")
	}

	// Marshal request to JSON
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	if c.verbose {
		log.Printf("[DEBUG] API Request: %s", truncateJSONForLog(jsonData, 80))
	}

	// Create HTTP request with context for cancellation support
	httpReq, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", apiVersion)

	if c.verbose {
		log.Printf("[DEBUG] Request headers: Content-Type=%s, anthropic-version=%s",
			httpReq.Header.Get("Content-Type"),
			httpReq.Header.Get("anthropic-version"))
	}

	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if c.verbose {
		log.Printf("[DEBUG] API Response (HTTP %d): %s", resp.StatusCode, truncateJSONForLog(body, 80))
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		// Check for rate limit error (HTTP 429)
		if resp.StatusCode == http.StatusTooManyRequests {
			retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"))
			if retryAfter == 0 {
				// Default to 60 seconds if no Retry-After header
				retryAfter = 60
			}

			var errResp ErrorResponse
			message := "Rate limit exceeded"
			if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error.Message != "" {
				message = errResp.Error.Message
			}

			return nil, &RateLimitError{
				Message:    message,
				RetryAfter: retryAfter,
				StatusCode: resp.StatusCode,
			}
		}

		// Handle other errors
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			// If we can't parse the error response, return the raw body
			return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
		}

		// Check for context length error (HTTP 400 with "prompt is too long")
		if resp.StatusCode == http.StatusBadRequest &&
			errResp.Error.Type == "invalid_request_error" &&
			regexp.MustCompile(`prompt is too long`).MatchString(errResp.Error.Message) {

			tokensUsed, tokensMax := parseContextLengthError(errResp.Error.Message)
			return nil, &ContextLengthError{
				Message:       errResp.Error.Message,
				TokensUsed:    tokensUsed,
				TokensMaximum: tokensMax,
				StatusCode:    resp.StatusCode,
			}
		}

		// Return detailed error information
		if errResp.Error.Type != "" {
			return nil, fmt.Errorf("%s: %s (HTTP %d)", errResp.Error.Type, errResp.Error.Message, resp.StatusCode)
		}
		return nil, fmt.Errorf("API error: %s (HTTP %d)", errResp.Error.Message, resp.StatusCode)
	}

	// Parse response
	var msgResp CreateMessageResponse
	if err := json.Unmarshal(body, &msgResp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	// Normalize tool_use blocks to ensure Input is always initialized
	msgResp.normalizeToolUses()

	return &msgResp, nil
}

// GetResponseText extracts the text content from a response
func (r *CreateMessageResponse) GetResponseText() string {
	for _, block := range r.Content {
		if block.Type == "text" && block.Text != "" {
			return block.Text
		}
	}
	return ""
}

// HasToolUse checks if the response contains tool use requests
func (r *CreateMessageResponse) HasToolUse() bool {
	for _, block := range r.Content {
		if block.Type == "tool_use" {
			return true
		}
	}
	return false
}

// normalizeToolUses ensures all tool_use blocks have Input initialized
// This fixes the issue where tools with no parameters come back with Input: nil
// which causes the API to reject the message when sent back
func (r *CreateMessageResponse) normalizeToolUses() {
	for i := range r.Content {
		if r.Content[i].Type == "tool_use" && r.Content[i].Input == nil {
			r.Content[i].Input = make(map[string]any)
		}
	}
}

// GetToolUses extracts all tool use blocks from the response
func (r *CreateMessageResponse) GetToolUses() []ContentBlock {
	var tools []ContentBlock
	for _, block := range r.Content {
		if block.Type == "tool_use" {
			tools = append(tools, block)
		}
	}
	return tools
}

// MakeCacheableSystem creates a system prompt with caching enabled
// This is useful for large system prompts that are reused across requests
func MakeCacheableSystem(systemPrompt string) []SystemBlock {
	return []SystemBlock{
		{
			Type: "text",
			Text: systemPrompt,
			CacheControl: &CacheControl{
				Type: "ephemeral",
			},
		},
	}
}

// Model represents a Claude model
type Model struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Type        string `json:"type"`
	CreatedAt   string `json:"created_at"`
}

// ModelsResponse represents the response from the models endpoint
type ModelsResponse struct {
	Data []Model `json:"data"`
}

// GetModels fetches the list of available models
func (c *Client) GetModels() ([]Model, error) {
	// Create HTTP request
	httpReq, err := http.NewRequest("GET", modelsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", apiVersion)

	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	// Models API response logging removed - too verbose

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
		}
		return nil, fmt.Errorf("%s: %s (HTTP %d)", errResp.Error.Type, errResp.Error.Message, resp.StatusCode)
	}

	// Parse response
	var modelsResp ModelsResponse
	if err := json.Unmarshal(body, &modelsResp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return modelsResp.Data, nil
}

// GetDefaultModels returns a hardcoded list of models as fallback
func GetDefaultModels() []Model {
	return []Model{
		{ID: "claude-sonnet-4-5-20250929", DisplayName: "Claude Sonnet 4.5", Type: "chat"},
		{ID: "claude-haiku-4-5-20250929", DisplayName: "Claude Haiku 4.5", Type: "chat"},
		{ID: "claude-3-5-sonnet-20241022", DisplayName: "Claude 3.5 Sonnet", Type: "chat"},
		{ID: "claude-3-5-haiku-20241022", DisplayName: "Claude 3.5 Haiku", Type: "chat"},
		{ID: "claude-3-opus-20240229", DisplayName: "Claude 3 Opus", Type: "chat"},
		{ID: "claude-3-sonnet-20240229", DisplayName: "Claude 3 Sonnet", Type: "chat"},
		{ID: "claude-3-haiku-20240307", DisplayName: "Claude 3 Haiku", Type: "chat"},
	}
}

// ResolveModel finds the best matching model from available models
// Handles version date changes by matching major version prefix
func ResolveModel(configModel string, availableModels []Model) (string, bool) {
	if configModel == "" {
		return "", false
	}

	// Try exact match first
	for _, model := range availableModels {
		if model.ID == configModel {
			return configModel, false // Exact match, no resolution needed
		}
	}

	// Extract major version prefix (e.g., "claude-sonnet-4-5" from "claude-sonnet-4-5-20250929")
	// Model format: claude-{name}-{major}-{minor}-{date}
	parts := strings.Split(configModel, "-")
	if len(parts) >= 4 {
		// Build prefix: claude-{name}-{major}-{minor}
		prefix := strings.Join(parts[:4], "-") // e.g., "claude-sonnet-4-5"

		// Find first model matching this prefix
		for _, model := range availableModels {
			if strings.HasPrefix(model.ID, prefix) {
				return model.ID, true // Resolved to different version
			}
		}
	}

	// No match found - return original (will likely fail, but let API handle it)
	return configModel, false
}

// StreamEvent represents a streaming event from the API
type StreamEvent struct {
	Type         string                 `json:"type"`
	Index        int                    `json:"index,omitempty"`
	Delta        *StreamDelta           `json:"delta,omitempty"`
	Message      *CreateMessageResponse `json:"message,omitempty"`
	ContentBlock *ContentBlock          `json:"content_block,omitempty"`
	Usage        *struct {
		InputTokens  int `json:"input_tokens,omitempty"`
		OutputTokens int `json:"output_tokens,omitempty"`
	} `json:"usage,omitempty"`
}

// StreamDelta represents incremental content in a stream
type StreamDelta struct {
	Type        string `json:"type"`
	Text        string `json:"text,omitempty"`
	PartialJSON string `json:"partial_json,omitempty"` // For input_json_delta
	StopReason  string `json:"stop_reason,omitempty"`
}

// CreateMessageStreaming sends a streaming message to Claude and accumulates the complete response
// This uses streaming internally to help with rate limits, but returns a complete response like CreateMessage
// The context can be used to cancel the request mid-stream
func (c *Client) CreateMessageStreaming(ctx context.Context, req CreateMessageRequest) (*CreateMessageResponse, error) {
	// If MaxTokens is 0, it will be omitted from the JSON request (omitempty)
	// allowing the API to use its own defaults/limits

	// Validate model
	if req.Model == "" {
		return nil, fmt.Errorf("model is required")
	}

	// Enable streaming
	req.Stream = true

	// Marshal request to JSON
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	if c.verbose {
		log.Printf("[DEBUG] Streaming API Request: %s", truncateJSONForLog(jsonData, 80))
	}

	// Create HTTP request with context for cancellation support
	httpReq, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", apiVersion)

	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		// Read error body
		body, _ := io.ReadAll(resp.Body)

		// Check for rate limit error (HTTP 429)
		if resp.StatusCode == http.StatusTooManyRequests {
			retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"))
			if retryAfter == 0 {
				retryAfter = 60
			}

			var errResp ErrorResponse
			message := "Rate limit exceeded"
			if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error.Message != "" {
				message = errResp.Error.Message
			}

			return nil, &RateLimitError{
				Message:    message,
				RetryAfter: retryAfter,
				StatusCode: resp.StatusCode,
			}
		}

		// Handle other errors
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
		}

		// Check for context length error
		if resp.StatusCode == http.StatusBadRequest &&
			errResp.Error.Type == "invalid_request_error" &&
			regexp.MustCompile(`prompt is too long`).MatchString(errResp.Error.Message) {

			tokensUsed, tokensMax := parseContextLengthError(errResp.Error.Message)
			return nil, &ContextLengthError{
				Message:       errResp.Error.Message,
				TokensUsed:    tokensUsed,
				TokensMaximum: tokensMax,
				StatusCode:    resp.StatusCode,
			}
		}

		if errResp.Error.Type != "" {
			return nil, fmt.Errorf("%s: %s (HTTP %d)", errResp.Error.Type, errResp.Error.Message, resp.StatusCode)
		}
		return nil, fmt.Errorf("API error: %s (HTTP %d)", errResp.Error.Message, resp.StatusCode)
	}

	// Parse streaming response
	return c.parseStreamingResponse(resp.Body)
}

// parseStreamingResponse reads SSE events and accumulates a complete response
func (c *Client) parseStreamingResponse(body io.Reader) (*CreateMessageResponse, error) {
	scanner := bufio.NewScanner(body)

	// Initialize response
	response := &CreateMessageResponse{
		Content: make([]ContentBlock, 0),
	}

	var currentBlock *ContentBlock
	var currentInputJSON string // Accumulate tool input JSON
	var eventType string
	var eventData []byte

	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines (event boundaries)
		if line == "" {
			if eventType != "" && len(eventData) > 0 {
				// Process accumulated event
				if err := c.processStreamEvent(eventType, eventData, response, &currentBlock, &currentInputJSON); err != nil {
					return nil, err
				}
				eventType = ""
				eventData = nil
			}
			continue
		}

		// Parse SSE format: "event: type" or "data: json"
		if strings.HasPrefix(line, "event: ") {
			eventType = strings.TrimPrefix(line, "event: ")
		} else if strings.HasPrefix(line, "data: ") {
			eventData = []byte(strings.TrimPrefix(line, "data: "))
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading stream: %w", err)
	}

	if c.verbose {
		log.Printf("[DEBUG] Streaming complete. Response ID: %s, Blocks: %d", response.ID, len(response.Content))
	}

	return response, nil
}

// processStreamEvent handles a single SSE event
func (c *Client) processStreamEvent(eventType string, data []byte, response *CreateMessageResponse, currentBlock **ContentBlock, currentInputJSON *string) error {
	var event StreamEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return fmt.Errorf("parsing event %s: %w", eventType, err)
	}

	switch eventType {
	case "message_start":
		// Initialize message metadata
		if event.Message != nil {
			response.ID = event.Message.ID
			response.Type = event.Message.Type
			response.Role = event.Message.Role
			response.Model = event.Message.Model
			response.Usage = event.Message.Usage
		}

	case "content_block_start":
		// Start a new content block
		if event.ContentBlock != nil {
			response.Content = append(response.Content, *event.ContentBlock)
			*currentBlock = &response.Content[len(response.Content)-1]
			// Reset input JSON accumulator for tool_use blocks
			if (*currentBlock).Type == "tool_use" {
				*currentInputJSON = ""
			}
		}

	case "content_block_delta":
		if event.Delta != nil && *currentBlock != nil {
			// Handle text delta
			if event.Delta.Type == "text_delta" {
				(*currentBlock).Text += event.Delta.Text
			}
			// Handle tool input JSON delta
			if event.Delta.Type == "input_json_delta" {
				*currentInputJSON += event.Delta.PartialJSON
			}
		}

	case "content_block_stop":
		// Parse accumulated tool input JSON if this was a tool_use block
		if *currentBlock != nil && (*currentBlock).Type == "tool_use" && *currentInputJSON != "" {
			var input map[string]any
			if err := json.Unmarshal([]byte(*currentInputJSON), &input); err != nil {
				if c.verbose {
					log.Printf("[DEBUG] Failed to parse tool input JSON: %v", err)
				}
			} else {
				(*currentBlock).Input = input
			}
			*currentInputJSON = ""
		}
		// Finish current block
		*currentBlock = nil

	case "message_delta":
		// Update stop reason and usage
		if event.Delta != nil && event.Delta.StopReason != "" {
			response.StopReason = event.Delta.StopReason
		}
		if event.Usage != nil {
			if event.Usage.OutputTokens > 0 {
				response.Usage.OutputTokens = event.Usage.OutputTokens
			}
		}

	case "message_stop":
		// Stream complete
		if c.verbose {
			log.Printf("[DEBUG] Stream stopped. Total output tokens: %d", response.Usage.OutputTokens)
		}

	case "ping":
		// Keep-alive, ignore

	default:
		if c.verbose {
			log.Printf("[DEBUG] Unknown event type: %s", eventType)
		}
	}

	return nil
}
