package runtime

import (
	"testing"
)

func TestBuiltinThrow(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]any
		wantErr bool
	}{
		{"with string", map[string]any{"0": "error message"}, true},
		{"with number", map[string]any{"0": 42.0}, true},
		{"no args", map[string]any{}, true},
		{"named arg", map[string]any{"message": "error"}, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			evaluator := &Evaluator{}
			_, err := builtinThrow(evaluator, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr %v, got %v", tt.wantErr, err != nil)
			}
			if tt.wantErr && err != nil {
				if _, ok := err.(*DusoError); !ok {
					t.Errorf("expected DusoError, got %T", err)
				}
			}
		})
	}
}

func TestBuiltinBreakpoint(t *testing.T) {
	t.Parallel()

	// Without debug mode, should return nil
	evaluator := &Evaluator{}
	result, err := builtinBreakpoint(evaluator, map[string]any{})
	if result != nil {
		t.Errorf("expected nil result without debug mode")
	}
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestBuiltinWatchSkipped(t *testing.T) {
	t.Skip("Watch requires execution context")
}
