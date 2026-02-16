package runtime

import (
	"testing"
)

func TestBuiltinContext(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args map[string]any
	}{
		{"no context", map[string]any{}},
		{"with unused args", map[string]any{"0": "ignored"}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			evaluator := &Evaluator{}
			// Context returns nil when no context getter is set
			result, err := builtinContext(evaluator, tt.args)
			if err != nil {
				t.Errorf("error = %v", err)
			}
			// Should return nil when no context
			if result != nil {
				t.Errorf("expected nil, got %v", result)
			}
		})
	}
}

func TestContextWithGetter(t *testing.T) {
	t.Parallel()

	evaluator := &Evaluator{}

	// Set a context getter
	gid := GetGoroutineID()
	SetContextGetter(gid, func() any {
		return "test context data"
	})
	defer ClearContextGetter(gid)

	result, err := builtinContext(evaluator, map[string]any{})
	if err != nil {
		t.Errorf("error = %v", err)
	}
	if result != "test context data" {
		t.Errorf("got %v, want test context data", result)
	}
}
