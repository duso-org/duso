package runtime

import (
	"testing"
)

func TestBuiltinParallelBasic(t *testing.T) {
	t.Parallel()

	// Test error cases that don't require execution
	tests := []struct {
		name    string
		args    map[string]any
		wantErr bool
	}{
		{"no args", map[string]any{}, true},
		{"empty array", map[string]any{"0": &[]Value{}}, false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			evaluator := &Evaluator{}
			result, err := builtinParallel(evaluator, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr %v, got %v", tt.wantErr, err != nil)
			}
			if !tt.wantErr && result == nil {
				t.Errorf("expected result")
			}
		})
	}
}
