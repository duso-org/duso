package script

import (
	"fmt"

	"github.com/duso-org/duso/pkg/anthropic"
)

// ConversationManager manages active conversations for scripts.
// This is part of the core language but specifically for Claude API integration.
type ConversationManager struct {
	conversations map[float64]*anthropic.Conversation
	counter       float64
}

// NewConversationManager creates a new conversation manager
func NewConversationManager() *ConversationManager {
	return &ConversationManager{
		conversations: make(map[float64]*anthropic.Conversation),
		counter:       0,
	}
}

// RegisterConversationAPI registers the conversation() and claude() functions.
//
// CORE LANGUAGE NOTICE: These functions are Claude API specific and are not part of
// the minimal core language. They are typically only available when using the duso CLI.
// See pkg/cli/ for the CLI-specific registration.
//
// This function is called from cmd/duso/main.go indirectly via interp.RegisterConversationAPI().
func RegisterConversationAPI(env *Environment) {
	manager := NewConversationManager()

	// conversation() constructor
	env.Define("conversation", NewGoFunction(func(args map[string]any) (any, error) {
		conv := anthropic.NewConversation()

		// Set system prompt if provided
		if system, ok := args["system"]; ok {
			conv.System(fmt.Sprintf("%v", system))
		}

		// Set model if provided
		if model, ok := args["model"]; ok {
			conv.Model(fmt.Sprintf("%v", model))
		}

		// Set API key if provided
		if key, ok := args["key"]; ok {
			conv.Key(fmt.Sprintf("%v", key))
		}

		// Set tokens if provided
		if tokens, ok := args["tokens"]; ok {
			if t, ok := tokens.(float64); ok {
				conv.MaxTokens(int(t))
			}
		}

		// Assign unique ID
		manager.counter++
		convID := manager.counter
		manager.conversations[convID] = conv

		// Create object with methods
		objMap := make(map[string]Value)

		// prompt method
		objMap["prompt"] = NewGoFunction(func(methodArgs map[string]any) (any, error) {
			c := manager.conversations[convID]
			if c == nil {
				return nil, fmt.Errorf("conversation no longer exists")
			}

			msg := ""
			if prompt, ok := methodArgs["0"]; ok {
				msg = fmt.Sprintf("%v", prompt)
			} else if prompt, ok := methodArgs["prompt"]; ok {
				msg = fmt.Sprintf("%v", prompt)
			}

			if msg == "" {
				return nil, fmt.Errorf("prompt() requires a message argument")
			}

			response, err := c.Prompt(msg)
			return response, err
		})

		// system method
		objMap["system"] = NewGoFunction(func(methodArgs map[string]any) (any, error) {
			c := manager.conversations[convID]
			if c == nil {
				return nil, fmt.Errorf("conversation no longer exists")
			}

			if sys, ok := methodArgs["0"]; ok {
				c.System(fmt.Sprintf("%v", sys))
				return nil, nil
			}
			return nil, fmt.Errorf("system() requires a string argument")
		})

		// model method
		objMap["model"] = NewGoFunction(func(methodArgs map[string]any) (any, error) {
			c := manager.conversations[convID]
			if c == nil {
				return nil, fmt.Errorf("conversation no longer exists")
			}

			if model, ok := methodArgs["0"]; ok {
				c.Model(fmt.Sprintf("%v", model))
				return nil, nil
			}
			return nil, fmt.Errorf("model() requires a string argument")
		})

		// key method
		objMap["key"] = NewGoFunction(func(methodArgs map[string]any) (any, error) {
			c := manager.conversations[convID]
			if c == nil {
				return nil, fmt.Errorf("conversation no longer exists")
			}

			if key, ok := methodArgs["0"]; ok {
				c.Key(fmt.Sprintf("%v", key))
				return nil, nil
			}
			return nil, fmt.Errorf("key() requires a string argument")
		})

		// tokens method
		objMap["tokens"] = NewGoFunction(func(methodArgs map[string]any) (any, error) {
			c := manager.conversations[convID]
			if c == nil {
				return nil, fmt.Errorf("conversation no longer exists")
			}

			if tokens, ok := methodArgs["0"].(float64); ok {
				c.MaxTokens(int(tokens))
				return nil, nil
			}
			return nil, fmt.Errorf("tokens() requires a number argument")
		})

		// clear method
		objMap["clear"] = NewGoFunction(func(methodArgs map[string]any) (any, error) {
			c := manager.conversations[convID]
			if c == nil {
				return nil, fmt.Errorf("conversation no longer exists")
			}

			c.Clear()
			return nil, nil
		})

		// usage method
		objMap["usage"] = NewGoFunction(func(methodArgs map[string]any) (any, error) {
			c := manager.conversations[convID]
			if c == nil {
				return nil, fmt.Errorf("conversation no longer exists")
			}

			return c.Usage(), nil
		})

		// Return the Value directly - interfaceToValue now handles this
		return NewObject(objMap), nil
	}))

	// claude() single-shot helper
	env.Define("claude", NewGoFunction(func(args map[string]any) (any, error) {
		prompt := ""
		system := ""
		model := ""
		apiKey := ""
		maxTokens := 0

		// Support both positional and named arguments
		if p, ok := args["0"]; ok {
			prompt = fmt.Sprintf("%v", p)
		}
		if p, ok := args["prompt"]; ok {
			prompt = fmt.Sprintf("%v", p)
		}

		if s, ok := args["1"]; ok {
			system = fmt.Sprintf("%v", s)
		}
		if s, ok := args["system"]; ok {
			system = fmt.Sprintf("%v", s)
		}

		if m, ok := args["2"]; ok {
			model = fmt.Sprintf("%v", m)
		}
		if m, ok := args["model"]; ok {
			model = fmt.Sprintf("%v", m)
		}

		if k, ok := args["key"]; ok {
			apiKey = fmt.Sprintf("%v", k)
		}

		if t, ok := args["tokens"]; ok {
			if tNum, ok := t.(float64); ok {
				maxTokens = int(tNum)
			}
		}

		if prompt == "" {
			return nil, fmt.Errorf("claude() requires a prompt argument")
		}

		response, err := anthropic.Claude(prompt, system, model, apiKey, maxTokens)
		return response, err
	}))
}
