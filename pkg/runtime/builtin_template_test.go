package runtime

import (
	"testing"
)

func TestBuiltinTemplate(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]any
		wantErr bool
	}{
		{"with template string", map[string]any{"0": "Hello {name}"}, false},
		{"empty string", map[string]any{"0": ""}, false},
		{"no args", map[string]any{}, true},
		{"non-string", map[string]any{"0": 42.0}, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			evaluator := &Evaluator{}
			result, err := builtinTemplate(evaluator, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr %v, got %v", tt.wantErr, err != nil)
			}
			if !tt.wantErr && result == nil {
				t.Errorf("expected result")
			}
		})
	}
}
