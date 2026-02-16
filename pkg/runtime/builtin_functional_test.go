package runtime

import (
	"testing"
)

func TestBuiltinMap(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]any
		wantErr bool
	}{
		{"no args", map[string]any{}, true},
		{"missing function", map[string]any{"0": &[]Value{}}, true},
		{"non-array", map[string]any{"0": "text", "1": nil}, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			evaluator := &Evaluator{}
			_, err := builtinMap(evaluator, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr %v, got %v", tt.wantErr, err != nil)
			}
		})
	}
}

func TestBuiltinFilter(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]any
		wantErr bool
	}{
		{"no args", map[string]any{}, true},
		{"missing function", map[string]any{"0": &[]Value{}}, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			evaluator := &Evaluator{}
			_, err := builtinFilter(evaluator, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr %v, got %v", tt.wantErr, err != nil)
			}
		})
	}
}

func TestBuiltinReduce(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]any
		wantErr bool
	}{
		{"no args", map[string]any{}, true},
		{"missing function", map[string]any{"0": &[]Value{}}, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			evaluator := &Evaluator{}
			_, err := builtinReduce(evaluator, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr %v, got %v", tt.wantErr, err != nil)
			}
		})
	}
}

func TestBuiltinSort(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]any
		wantErr bool
	}{
		{"empty array", map[string]any{"0": &[]Value{}}, false},
		{"non-array", map[string]any{"0": "text"}, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			evaluator := &Evaluator{}
			_, err := builtinSort(evaluator, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr %v, got %v", tt.wantErr, err != nil)
			}
		})
	}
}
