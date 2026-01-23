// Package cli provides CLI-specific functions for Duso scripts.
// This file contains documentation and registration for Claude API functions.
package cli

import (
	"github.com/duso-org/duso/pkg/script"
)

// RegisterConversationAPI registers the conversation() and claude() functions
// in the given Duso environment.
//
// These functions provide integration with the Anthropic Claude API.
// They are NOT part of the core language and are only available when:
// 1. Using the duso CLI (automatically registered)
// 2. An embedded application explicitly enables them via cli.RegisterFunctions()
//
// The actual implementation is in pkg/script/conversation_api.go to keep Claude
// API logic in the core package and avoid circular imports.
// This function delegates to that implementation.
//
// # Available Functions
//
// ## conversation(system [, model] [, tokens] [, key]) -> object
//
// Creates a multi-turn conversation with Claude that maintains context.
//
// Example:
//
//	agent = conversation(
//	    system = "You are a helpful assistant",
//	    model = "claude-opus-4-5-20251101"
//	)
//	response1 = agent.prompt("Hello!")
//	response2 = agent.prompt("How are you?")  // Context is maintained
//
// Methods on conversation object:
//   - .prompt(message) - Send message and get response
//   - .system(prompt) - Update system prompt
//   - .model(modelID) - Change Claude model
//   - .key(apiKey) - Set API key
//   - .tokens(maxTokens) - Set max tokens per response
//   - .clear() - Clear conversation history
//   - .usage() - Get token usage statistics
//
// ## claude(prompt [, model] [, tokens] [, key]) -> string
//
// Single-shot query to Claude. Does not maintain context.
// Useful for one-off questions without conversation history.
//
// Example:
//
//	answer = claude("What is 2+2?")
//	code = claude("Write a hello world program", model = "claude-opus-4-5-20251101")
func RegisterConversationAPI(env *script.Environment) {
	// Delegate to the core implementation
	// This is defined in pkg/script/conversation_api.go
	script.RegisterConversationAPI(env)
}
