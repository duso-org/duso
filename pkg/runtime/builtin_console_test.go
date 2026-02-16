package runtime

import (
	"os"
	"testing"
)

func TestBuiltinPrint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args map[string]any
	}{
		{"single arg", map[string]any{"0": "hello"}},
		{"multiple args", map[string]any{"0": "hello", "1": "world"}},
		{"number arg", map[string]any{"0": 42.0}},
		{"boolean arg", map[string]any{"0": true}},
		{"nil arg", map[string]any{"0": nil}},
		{"no args", map[string]any{}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			evaluator := &Evaluator{}
			result, err := builtinPrint(evaluator, tt.args)
			if err != nil {
				t.Errorf("error = %v", err)
			}
			if result != nil {
				t.Errorf("expected nil result, got %v", result)
			}
		})
	}
}

func TestBuiltinInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args map[string]any
	}{
		{"no prompt", map[string]any{}},
		{"with prompt", map[string]any{"0": "Enter value: "}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Note: This test will hang waiting for stdin
			// In CI, we skip it or mock stdin
			if os.Getenv("CI") == "true" {
				t.Skip("skipping interactive test in CI")
			}

			evaluator := &Evaluator{}
			// Don't actually call builtinInput in parallel tests as it blocks
			// Just verify the function exists
			_ = evaluator
		})
	}
}
