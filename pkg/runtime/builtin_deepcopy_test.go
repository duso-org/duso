package runtime

import (
	"testing"
)

func TestBuiltinDeepCopy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    map[string]any
		wantErr bool
	}{
		{"copy number", map[string]any{"0": 42.0}, false},
		{"copy string", map[string]any{"0": "hello"}, false},
		{"copy boolean", map[string]any{"0": true}, false},
		{"copy nil", map[string]any{"0": nil}, false},
		{"no argument", map[string]any{}, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			evaluator := &Evaluator{}
			result, err := builtinDeepCopy(evaluator, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && result == nil {
				t.Errorf("expected result, got nil")
			}
		})
	}
}

func TestDeepCopyPrimitive(t *testing.T) {
	t.Parallel()

	evaluator := &Evaluator{}

	// Test number
	result, _ := builtinDeepCopy(evaluator, map[string]any{"0": 42.0})
	val := result.(Value)
	if val.AsNumber() != 42.0 {
		t.Errorf("copy number failed")
	}

	// Test string
	result, _ = builtinDeepCopy(evaluator, map[string]any{"0": "test"})
	val = result.(Value)
	if val.AsString() != "test" {
		t.Errorf("copy string failed")
	}

	// Test bool
	result, _ = builtinDeepCopy(evaluator, map[string]any{"0": true})
	val = result.(Value)
	if !val.AsBool() {
		t.Errorf("copy bool failed")
	}
}
