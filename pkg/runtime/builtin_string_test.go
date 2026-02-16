package runtime

import (
	"testing"
)

func TestBuiltinUpper(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    map[string]any
		want    string
		wantErr bool
	}{
		{"hello", map[string]any{"0": "hello"}, "HELLO", false},
		{"HELLO", map[string]any{"0": "HELLO"}, "HELLO", false},
		{"Hello World", map[string]any{"0": "Hello World"}, "HELLO WORLD", false},
		{"empty string", map[string]any{"0": ""}, "", false},
		{"number coerced", map[string]any{"0": 42.0}, "42", false},
		{"no argument", map[string]any{}, "", true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			evaluator := &Evaluator{}
			result, err := builtinUpper(evaluator, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && result != tt.want {
				t.Errorf("got %q, want %q", result, tt.want)
			}
		})
	}
}

func TestBuiltinLower(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    map[string]any
		want    string
		wantErr bool
	}{
		{"HELLO", map[string]any{"0": "HELLO"}, "hello", false},
		{"hello", map[string]any{"0": "hello"}, "hello", false},
		{"Hello World", map[string]any{"0": "Hello World"}, "hello world", false},
		{"empty string", map[string]any{"0": ""}, "", false},
		{"number coerced", map[string]any{"0": 42.0}, "42", false},
		{"no argument", map[string]any{}, "", true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			evaluator := &Evaluator{}
			result, err := builtinLower(evaluator, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && result != tt.want {
				t.Errorf("got %q, want %q", result, tt.want)
			}
		})
	}
}

func TestBuiltinSubstr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    map[string]any
		want    string
		wantErr bool
	}{
		{"start 0", map[string]any{"0": "hello", "1": 0.0}, "hello", false},
		{"start 1", map[string]any{"0": "hello", "1": 1.0}, "ello", false},
		{"start 2 length 2", map[string]any{"0": "hello", "1": 2.0, "2": 2.0}, "ll", false},
		{"start 0 length 5", map[string]any{"0": "hello", "1": 0.0, "2": 5.0}, "hello", false},
		{"start beyond string", map[string]any{"0": "hello", "1": 10.0}, "", false},
		{"negative start", map[string]any{"0": "hello", "1": -1.0}, "o", false},
		{"negative start with length", map[string]any{"0": "hello", "1": -2.0, "2": 2.0}, "lo", false},
		{"non-string", map[string]any{"0": 42.0, "1": 0.0}, "", true},
		{"non-number index", map[string]any{"0": "hello", "1": "bad"}, "", true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			evaluator := &Evaluator{}
			result, err := builtinSubstr(evaluator, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && result != tt.want {
				t.Errorf("got %q, want %q", result, tt.want)
			}
		})
	}
}

func TestBuiltinTrim(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    map[string]any
		want    string
		wantErr bool
	}{
		{"no whitespace", map[string]any{"0": "hello"}, "hello", false},
		{"leading space", map[string]any{"0": "  hello"}, "hello", false},
		{"trailing space", map[string]any{"0": "hello  "}, "hello", false},
		{"both sides", map[string]any{"0": "  hello  "}, "hello", false},
		{"empty string", map[string]any{"0": ""}, "", false},
		{"tabs and newlines", map[string]any{"0": "\t\nhello\n\t"}, "hello", false},
		{"non-string", map[string]any{"0": 42.0}, "", true},
		{"no argument", map[string]any{}, "", true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			evaluator := &Evaluator{}
			result, err := builtinTrim(evaluator, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && result != tt.want {
				t.Errorf("got %q, want %q", result, tt.want)
			}
		})
	}
}
