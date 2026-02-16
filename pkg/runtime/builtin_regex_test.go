package runtime

import (
	"testing"
)

func TestBuiltinContains(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    map[string]any
		want    bool
		wantErr bool
	}{
		{"simple match", map[string]any{"0": "hello world", "1": "world"}, true, false},
		{"no match", map[string]any{"0": "hello world", "1": "xyz"}, false, false},
		{"regex pattern", map[string]any{"0": "test123", "1": "\\d+"}, true, false},
		{"case insensitive", map[string]any{"0": "Hello", "1": "hello", "ignore_case": true}, true, false},
		{"case sensitive no match", map[string]any{"0": "Hello", "1": "hello"}, false, false},
		{"empty string", map[string]any{"0": "hello", "1": ""}, true, false},
		{"non-string input", map[string]any{"0": 42.0, "1": "test"}, false, true},
		{"non-string pattern", map[string]any{"0": "hello", "1": 42.0}, false, true},
		{"invalid regex", map[string]any{"0": "hello", "1": "["}, false, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			evaluator := &Evaluator{}
			result, err := builtinContains(evaluator, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && result != tt.want {
				t.Errorf("got %v, want %v", result, tt.want)
			}
		})
	}
}

func TestBuiltinFind(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    map[string]any
		wantLen int
		wantErr bool
	}{
		{"find single match", map[string]any{"0": "hello world", "1": "world"}, 1, false},
		{"find multiple matches", map[string]any{"0": "abc123def456", "1": "\\d+"}, 2, false},
		{"no matches", map[string]any{"0": "hello", "1": "\\d+"}, 0, false},
		{"case insensitive", map[string]any{"0": "Hello World", "1": "hello", "ignore_case": true}, 1, false},
		{"non-string input", map[string]any{"0": 42.0, "1": "test"}, 0, true},
		{"non-string pattern", map[string]any{"0": "hello", "1": 42.0}, 0, true},
		{"invalid regex", map[string]any{"0": "hello", "1": "["}, 0, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			evaluator := &Evaluator{}
			result, err := builtinFind(evaluator, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				arr := result.(Value)
				matches := arr.AsArray()
				if len(matches) != tt.wantLen {
					t.Errorf("got %d matches, want %d", len(matches), tt.wantLen)
				}
			}
		})
	}
}

func TestFindMatchObject(t *testing.T) {
	t.Parallel()

	evaluator := &Evaluator{}
	result, err := builtinFind(evaluator, map[string]any{"0": "hello123world", "1": "\\d+"})
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	arr := result.(Value)
	matches := arr.AsArray()
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}

	matchObj := matches[0].AsObject()
	text := matchObj["text"]
	pos := matchObj["pos"]
	len := matchObj["len"]

	if text.AsString() != "123" {
		t.Errorf("text = %q, want 123", text.AsString())
	}
	if pos.AsNumber() != 5.0 {
		t.Errorf("pos = %v, want 5", pos.AsNumber())
	}
	if len.AsNumber() != 3.0 {
		t.Errorf("len = %v, want 3", len.AsNumber())
	}
}

func TestBuiltinReplace(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    map[string]any
		want    string
		wantErr bool
	}{
		{"simple replace", map[string]any{"0": "hello world", "1": "world", "2": "Duso"}, "hello Duso", false},
		{"regex replace", map[string]any{"0": "abc123def456", "1": "\\d+", "2": "X"}, "abcXdefX", false},
		{"replace all", map[string]any{"0": "aaa", "1": "a", "2": "b"}, "bbb", false},
		{"case insensitive replace", map[string]any{"0": "Hello", "1": "hello", "2": "hi", "ignore_case": true}, "hi", false},
		{"non-string input", map[string]any{"0": 42.0, "1": "test", "2": "replace"}, "", true},
		{"non-string pattern", map[string]any{"0": "hello", "1": 42.0, "2": "replace"}, "", true},
		{"missing replacement", map[string]any{"0": "hello", "1": "world"}, "", true},
		{"invalid regex", map[string]any{"0": "hello", "1": "[", "2": "replace"}, "", true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			evaluator := &Evaluator{}
			result, err := builtinReplace(evaluator, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && result != tt.want {
				t.Errorf("got %q, want %q", result, tt.want)
			}
		})
	}
}
