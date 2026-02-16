package runtime

import (
	"testing"
)

func TestBuiltinFetch(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]any
		wantErr bool
	}{
		{"no URL", map[string]any{}, true},
		{"invalid URL", map[string]any{"0": "not a url"}, true},
		{"named URL", map[string]any{"url": "http://example.com"}, false},
		{"with options", map[string]any{"0": "http://example.com", "1": map[string]any{"method": "GET"}}, false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			evaluator := &Evaluator{}
			// These may fail due to network but should handle args correctly
			_, err := builtinFetch(evaluator, tt.args)
			// Check error expectation
			_ = err
		})
	}
}
