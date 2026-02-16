package runtime

import (
	"testing"
)

func TestTypeFunction(t *testing.T) {
	tests := []struct {
		name  string
		val   any
		want  string
	}{
		{"number", 42.0, "number"},
		{"string", "hello", "string"},
		{"boolean true", true, "boolean"},
		{"boolean false", false, "boolean"},
		{"nil", nil, "nil"},
	}

	for _, tt := range tests {
		evaluator := &Evaluator{}
		result, err := builtinType(evaluator, map[string]any{"0": tt.val})
		if err != nil {
			t.Errorf("%s: error: %v", tt.name, err)
			continue
		}
		if result != tt.want {
			t.Errorf("%s: got %q, want %q", tt.name, result, tt.want)
		}
	}
}

func TestLenFunction(t *testing.T) {
	tests := []struct {
		name  string
		val   any
		want  float64
		err   bool
	}{
		{"string hello", "hello", 5.0, false},
		{"empty string", "", 0.0, false},
		{"string with spaces", "a b c", 5.0, false},
		{"unicode", "你好", 6.0, false}, // utf-8 encoded, each character is multiple bytes
		{"number error", 42.0, 0, true},
		{"nil returns 0", nil, 0.0, false}, // nil returns 0, not error
	}

	for _, tt := range tests {
		evaluator := &Evaluator{}
		result, err := builtinLen(evaluator, map[string]any{"0": tt.val})
		if (err != nil) != tt.err {
			t.Errorf("%s: error = %v, want err = %v", tt.name, err, tt.err)
		}
		if err == nil {
			if result != tt.want {
				t.Errorf("%s: got %v, want %v", tt.name, result, tt.want)
			}
		}
	}
}

func TestToNumberFunction(t *testing.T) {
	tests := []struct {
		name  string
		val   any
		want  float64
		err   bool
	}{
		{"number", 42.0, 42.0, false},
		{"string number", "123", 123.0, false},
		{"string float", "3.14", 3.14, false},
		{"zero", 0.0, 0.0, false},
		{"negative", -5.0, -5.0, false},
		{"bool true", true, 1.0, false},
		{"bool false", false, 0.0, false},
		{"string non-number", "abc", 0.0, false}, // Returns 0.0 on parse error, not error
	}

	for _, tt := range tests {
		evaluator := &Evaluator{}
		result, err := builtinToNumber(evaluator, map[string]any{"0": tt.val})
		if (err != nil) != tt.err {
			t.Errorf("%s: error = %v, want err = %v", tt.name, err, tt.err)
		}
		if err == nil {
			if result != tt.want {
				t.Errorf("%s: got %v, want %v", tt.name, result, tt.want)
			}
		}
	}
}

func TestToStringFunction(t *testing.T) {
	tests := []struct {
		name string
		val  any
		want string
	}{
		{"number 42", 42.0, "42"},
		{"number 0", 0.0, "0"},
		{"string", "hello", "hello"},
		{"empty string", "", ""},
		{"true", true, "true"},
		{"false", false, "false"},
		{"nil", nil, "<nil>"}, // fmt.Sprintf("%v", nil) returns "<nil>"
	}

	for _, tt := range tests {
		evaluator := &Evaluator{}
		result, err := builtinToString(evaluator, map[string]any{"0": tt.val})
		if err != nil {
			t.Errorf("%s: error: %v", tt.name, err)
			continue
		}
		if result != tt.want {
			t.Errorf("%s: got %q, want %q", tt.name, result, tt.want)
		}
	}
}

func TestToBoolFunction(t *testing.T) {
	tests := []struct {
		name string
		val  any
		want bool
	}{
		{"number 0", 0.0, false},
		{"number 1", 1.0, true},
		{"number -1", -1.0, true},
		{"string empty", "", false},
		{"string non-empty", "hello", true},
		{"bool true", true, true},
		{"bool false", false, false},
		{"nil", nil, false},
	}

	for _, tt := range tests {
		evaluator := &Evaluator{}
		result, err := builtinToBool(evaluator, map[string]any{"0": tt.val})
		if err != nil {
			t.Errorf("%s: error: %v", tt.name, err)
			continue
		}
		if result != tt.want {
			t.Errorf("%s: got %v, want %v", tt.name, result, tt.want)
		}
	}
}
