package anthropic

import (
	"context"
	"os"
)

// Conversation manages a multi-turn conversation with Claude
type Conversation struct {
	client         *Client
	messages       []Message
	system         string
	model          string
	maxTokens      int
	inputTokens    int
	outputTokens   int
	cacheTokens    int
	apiKey         string // Can be overridden per conversation
}

// NewConversation creates a new conversation with Claude
func NewConversation() *Conversation {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	model := os.Getenv("ANTHROPIC_MODEL")
	if model == "" {
		model = "claude-opus-4-5-20251101"
	}

	return &Conversation{
		client:    NewClient(apiKey, false),
		messages:  []Message{},
		system:    "",
		model:     model,
		maxTokens: 4096,
		apiKey:    apiKey,
	}
}

// Prompt sends a message and returns Claude's response
func (c *Conversation) Prompt(msg string) (string, error) {
	// Append user message
	c.messages = append(c.messages, Message{
		Role:    "user",
		Content: msg,
	})

	// Build request
	req := CreateMessageRequest{
		Model:     c.model,
		Messages:  c.messages,
		MaxTokens: c.maxTokens,
	}

	if c.system != "" {
		req.System = c.system
	}

	// Make API call with override key if set
	client := c.client
	if c.apiKey != os.Getenv("ANTHROPIC_API_KEY") && c.apiKey != "" {
		// Use override key
		client = NewClient(c.apiKey, false)
	}

	ctx := context.Background()
	resp, err := client.CreateMessage(ctx, req)
	if err != nil {
		return "", err
	}

	// Extract text response
	var responseText string
	for _, block := range resp.Content {
		if block.Type == "text" {
			responseText += block.Text
		}
	}

	// Track tokens
	c.inputTokens += resp.Usage.InputTokens
	c.outputTokens += resp.Usage.OutputTokens
	if resp.Usage.CacheCreationInputTokens > 0 {
		c.cacheTokens += resp.Usage.CacheCreationInputTokens
	}
	if resp.Usage.CacheReadInputTokens > 0 {
		c.cacheTokens += resp.Usage.CacheReadInputTokens
	}

	// Append assistant response
	c.messages = append(c.messages, Message{
		Role:    "assistant",
		Content: responseText,
	})

	return responseText, nil
}

// System sets the system prompt
func (c *Conversation) System(text string) {
	c.system = text
}

// Model sets the model to use
func (c *Conversation) Model(name string) {
	c.model = name
}

// Key sets the API key (for override)
func (c *Conversation) Key(key string) {
	if key != "" {
		if err := ValidateAPIKey(key); err == nil {
			c.apiKey = key
		}
	}
}

// MaxTokens sets the maximum tokens for responses
func (c *Conversation) MaxTokens(tokens int) {
	if tokens > 0 {
		c.maxTokens = tokens
	}
}

// Clear clears the conversation history
func (c *Conversation) Clear() {
	c.messages = []Message{}
	c.inputTokens = 0
	c.outputTokens = 0
	c.cacheTokens = 0
}

// Usage returns usage statistics
func (c *Conversation) Usage() map[string]interface{} {
	return map[string]interface{}{
		"input_tokens":  c.inputTokens,
		"output_tokens": c.outputTokens,
		"cache_tokens":  c.cacheTokens,
		"total_tokens":  c.inputTokens + c.outputTokens + c.cacheTokens,
	}
}

// Claude is a convenience function for single-shot API calls
func Claude(prompt string, system string, model string, apiKey string, maxTokens int) (string, error) {
	if maxTokens <= 0 {
		maxTokens = 4096
	}

	if model == "" {
		model = os.Getenv("ANTHROPIC_MODEL")
		if model == "" {
			model = "claude-opus-4-5-20251101"
		}
	}

	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
	}

	if err := ValidateAPIKey(apiKey); err != nil {
		return "", err
	}

	client := NewClient(apiKey, false)

	req := CreateMessageRequest{
		Model:     model,
		MaxTokens: maxTokens,
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	if system != "" {
		req.System = system
	}

	ctx := context.Background()
	resp, err := client.CreateMessage(ctx, req)
	if err != nil {
		return "", err
	}

	var responseText string
	for _, block := range resp.Content {
		if block.Type == "text" {
			responseText += block.Text
		}
	}

	return responseText, nil
}
