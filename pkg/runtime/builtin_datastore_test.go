package runtime

import (
	"testing"
)

func TestBuiltinDatastore(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]any
		wantErr bool
	}{
		{"with namespace", map[string]any{"0": "test"}, false},
		{"no namespace", map[string]any{}, true},
		{"named namespace", map[string]any{"namespace": "test"}, false},
		{"sys namespace", map[string]any{"0": "sys"}, false},
		{"with config", map[string]any{"0": "test", "1": map[string]any{}}, false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			evaluator := &Evaluator{}
			result, err := builtinDatastore(evaluator, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr %v, got %v", tt.wantErr, err != nil)
			}
			if !tt.wantErr && result == nil {
				t.Errorf("expected result")
			}
		})
	}
}

func TestDatastoreCount(t *testing.T) {
	// GetDatastoreCount should return a number
	count := GetDatastoreCount()
	if count < 0 {
		t.Errorf("datastore count negative: %d", count)
	}
}
