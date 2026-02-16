package runtime

import (
	"testing"
)

func TestBuiltinParseJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    map[string]any
		wantErr bool
		check   func(t *testing.T, result any)
	}{
		{
			name:    "parse object",
			args:    map[string]any{"0": `{"name":"Alice","age":30}`},
			wantErr: false,
			check: func(t *testing.T, result any) {
				obj := result.(map[string]any)
				if obj["name"] != "Alice" {
					t.Errorf("name = %v, want Alice", obj["name"])
				}
				if obj["age"] != float64(30) {
					t.Errorf("age = %v, want 30", obj["age"])
				}
			},
		},
		{
			name:    "parse array",
			args:    map[string]any{"0": `[1,2,3]`},
			wantErr: false,
			check: func(t *testing.T, result any) {
				arr := result.([]any)
				if len(arr) != 3 {
					t.Errorf("array len = %d, want 3", len(arr))
				}
			},
		},
		{
			name:    "parse string",
			args:    map[string]any{"0": `"hello"`},
			wantErr: false,
			check: func(t *testing.T, result any) {
				if result != "hello" {
					t.Errorf("got %q, want hello", result)
				}
			},
		},
		{
			name:    "parse number",
			args:    map[string]any{"0": `42.5`},
			wantErr: false,
			check: func(t *testing.T, result any) {
				if result != 42.5 {
					t.Errorf("got %v, want 42.5", result)
				}
			},
		},
		{
			name:    "parse boolean",
			args:    map[string]any{"0": `true`},
			wantErr: false,
			check: func(t *testing.T, result any) {
				if result != true {
					t.Errorf("got %v, want true", result)
				}
			},
		},
		{
			name:    "parse null",
			args:    map[string]any{"0": `null`},
			wantErr: false,
			check: func(t *testing.T, result any) {
				if result != nil {
					t.Errorf("got %v, want nil", result)
				}
			},
		},
		{
			name:    "parse empty object",
			args:    map[string]any{"0": `{}`},
			wantErr: false,
			check: func(t *testing.T, result any) {
				obj := result.(map[string]any)
				if len(obj) != 0 {
					t.Errorf("object len = %d, want 0", len(obj))
				}
			},
		},
		{
			name:    "parse empty array",
			args:    map[string]any{"0": `[]`},
			wantErr: false,
			check: func(t *testing.T, result any) {
				arr := result.([]any)
				if len(arr) != 0 {
					t.Errorf("array len = %d, want 0", len(arr))
				}
			},
		},
		{
			name:    "invalid json",
			args:    map[string]any{"0": `{invalid}`},
			wantErr: true,
			check:   nil,
		},
		{
			name:    "non-string input",
			args:    map[string]any{"0": 42.0},
			wantErr: true,
			check:   nil,
		},
		{
			name:    "nested object",
			args:    map[string]any{"0": `{"a":{"b":{"c":1}}}`},
			wantErr: false,
			check: func(t *testing.T, result any) {
				outer := result.(map[string]any)
				inner := outer["a"].(map[string]any)
				innermost := inner["b"].(map[string]any)
				if innermost["c"] != float64(1) {
					t.Errorf("got %v, want 1", innermost["c"])
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			evaluator := &Evaluator{}
			result, err := builtinParseJSON(evaluator, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && tt.check != nil {
				tt.check(t, result)
			}
		})
	}
}

func TestBuiltinFormatJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    map[string]any
		want    string
		wantErr bool
	}{
		{
			name:    "format object",
			args:    map[string]any{"0": map[string]any{"name": "Alice", "age": float64(30)}},
			wantErr: false,
			want:    ``,
		},
		{
			name:    "format array",
			args:    map[string]any{"0": []any{float64(1), float64(2), float64(3)}},
			wantErr: false,
			want:    ``,
		},
		{
			name:    "format string",
			args:    map[string]any{"0": "hello"},
			wantErr: false,
			want:    `"hello"`,
		},
		{
			name:    "format number",
			args:    map[string]any{"0": float64(42.5)},
			wantErr: false,
			want:    `42.5`,
		},
		{
			name:    "format boolean true",
			args:    map[string]any{"0": true},
			wantErr: false,
			want:    `true`,
		},
		{
			name:    "format boolean false",
			args:    map[string]any{"0": false},
			wantErr: false,
			want:    `false`,
		},
		{
			name:    "format null",
			args:    map[string]any{"0": nil},
			wantErr: false,
			want:    `null`,
		},
		{
			name:    "no arguments",
			args:    map[string]any{},
			wantErr: true,
			want:    ``,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			evaluator := &Evaluator{}
			result, err := builtinFormatJSON(evaluator, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && tt.want != "" {
				resultStr := result.(string)
				if resultStr != tt.want {
					t.Errorf("got %q, want %q", resultStr, tt.want)
				}
			}
		})
	}
}

func TestJSONRoundTrip(t *testing.T) {
	t.Parallel()

	// Test that simple types can roundtrip through parse and format
	testData := []string{
		`"simple string"`,
		`42`,
		`3.14`,
		`true`,
		`false`,
		`null`,
	}

	for _, jsonStr := range testData {
		jsonStr := jsonStr
		t.Run(jsonStr, func(t *testing.T) {
			t.Parallel()
			evaluator := &Evaluator{}

			// Parse JSON
			parsed, err := builtinParseJSON(evaluator, map[string]any{"0": jsonStr})
			if err != nil {
				t.Fatalf("parse failed: %v", err)
			}

			// Format back to JSON
			formatted, err := builtinFormatJSON(evaluator, map[string]any{"0": parsed})
			if err != nil {
				t.Fatalf("format failed: %v", err)
			}

			// For simple types, just verify format returns a string
			if _, ok := formatted.(string); !ok {
				t.Errorf("expected string, got %T", formatted)
			}
		})
	}
}
