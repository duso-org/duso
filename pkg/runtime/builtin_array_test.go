package runtime

import (
	"testing"
)

func TestBuiltinKeys(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    map[string]any
		wantLen int
		wantErr bool
	}{
		{
			name:    "object with keys",
			args:    map[string]any{"0": map[string]any{"a": 1.0, "b": 2.0, "c": 3.0}},
			wantLen: 3,
			wantErr: false,
		},
		{
			name:    "empty object",
			args:    map[string]any{"0": map[string]any{}},
			wantLen: 0,
			wantErr: false,
		},
		{
			name:    "non-object",
			args:    map[string]any{"0": "text"},
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			evaluator := &Evaluator{}
			result, err := builtinKeys(evaluator, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil {
				keys := result.([]any)
				if len(keys) != tt.wantLen {
					t.Errorf("got len %d, want %d", len(keys), tt.wantLen)
				}
			}
		})
	}
}

func TestBuiltinValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    map[string]any
		wantLen int
		wantErr bool
	}{
		{
			name:    "object with values",
			args:    map[string]any{"0": map[string]any{"a": 1.0, "b": 2.0, "c": 3.0}},
			wantLen: 3,
			wantErr: false,
		},
		{
			name:    "empty object",
			args:    map[string]any{"0": map[string]any{}},
			wantLen: 0,
			wantErr: false,
		},
		{
			name:    "non-object",
			args:    map[string]any{"0": "text"},
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			evaluator := &Evaluator{}
			result, err := builtinValues(evaluator, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil {
				values := result.([]any)
				if len(values) != tt.wantLen {
					t.Errorf("got len %d, want %d", len(values), tt.wantLen)
				}
			}
		})
	}
}

func TestBuiltinSplit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    map[string]any
		wantLen int
		wantErr bool
	}{
		{
			name:    "simple split",
			args:    map[string]any{"0": "a,b,c", "1": ","},
			wantLen: 3,
			wantErr: false,
		},
		{
			name:    "split with spaces",
			args:    map[string]any{"0": "hello world foo", "1": " "},
			wantLen: 3,
			wantErr: false,
		},
		{
			name:    "no separator found",
			args:    map[string]any{"0": "hello", "1": ","},
			wantLen: 1,
			wantErr: false,
		},
		{
			name:    "empty string",
			args:    map[string]any{"0": "", "1": ","},
			wantLen: 1,
			wantErr: false,
		},
		{
			name:    "non-string input",
			args:    map[string]any{"0": 42.0, "1": ","},
			wantLen: 0,
			wantErr: true,
		},
		{
			name:    "non-string separator",
			args:    map[string]any{"0": "a,b,c", "1": 42.0},
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			evaluator := &Evaluator{}
			result, err := builtinSplit(evaluator, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil {
				parts := result.([]any)
				if len(parts) != tt.wantLen {
					t.Errorf("got len %d, want %d", len(parts), tt.wantLen)
				}
			}
		})
	}
}

func TestBuiltinRange(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    map[string]any
		wantLen int
		wantErr bool
	}{
		{
			name:    "range 0 to 5",
			args:    map[string]any{"0": 0.0, "1": 5.0},
			wantLen: 6,
			wantErr: false,
		},
		{
			name:    "range 1 to 3",
			args:    map[string]any{"0": 1.0, "1": 3.0},
			wantLen: 3,
			wantErr: false,
		},
		{
			name:    "range with step",
			args:    map[string]any{"0": 0.0, "1": 10.0, "2": 2.0},
			wantLen: 6,
			wantErr: false,
		},
		{
			name:    "negative range",
			args:    map[string]any{"0": 5.0, "1": 0.0, "2": -1.0},
			wantLen: 6,
			wantErr: false,
		},
		{
			name:    "zero step",
			args:    map[string]any{"0": 0.0, "1": 5.0, "2": 0.0},
			wantLen: 0,
			wantErr: true,
		},
		{
			name:    "non-number start",
			args:    map[string]any{"0": "text", "1": 5.0},
			wantLen: 0,
			wantErr: true,
		},
		{
			name:    "non-number end",
			args:    map[string]any{"0": 0.0, "1": "text"},
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			evaluator := &Evaluator{}
			result, err := builtinRange(evaluator, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil {
				nums := result.([]any)
				if len(nums) != tt.wantLen {
					t.Errorf("got len %d, want %d", len(nums), tt.wantLen)
				}
			}
		})
	}
}

func TestBuiltinPush(t *testing.T) {
	t.Parallel()

	t.Run("push single item", func(t *testing.T) {
		t.Parallel()
		arr := []Value{}
		arrPtr := &arr
		args := map[string]any{"0": arrPtr, "1": 42.0}
		evaluator := &Evaluator{}
		result, err := builtinPush(evaluator, args)
		if err != nil {
			t.Errorf("error = %v", err)
		}
		if result != 1.0 {
			t.Errorf("got %v, want 1.0", result)
		}
		if len(*arrPtr) != 1 {
			t.Errorf("array len = %d, want 1", len(*arrPtr))
		}
	})

	t.Run("push multiple items", func(t *testing.T) {
		t.Parallel()
		arr := []Value{}
		arrPtr := &arr
		args := map[string]any{"0": arrPtr, "1": 1.0, "2": 2.0, "3": 3.0}
		evaluator := &Evaluator{}
		result, err := builtinPush(evaluator, args)
		if err != nil {
			t.Errorf("error = %v", err)
		}
		if result != 3.0 {
			t.Errorf("got %v, want 3.0", result)
		}
		if len(*arrPtr) != 3 {
			t.Errorf("array len = %d, want 3", len(*arrPtr))
		}
	})

	t.Run("push to non-array", func(t *testing.T) {
		t.Parallel()
		args := map[string]any{"0": "text", "1": 42.0}
		evaluator := &Evaluator{}
		_, err := builtinPush(evaluator, args)
		if err == nil {
			t.Error("expected error")
		}
	})
}

func TestBuiltinPop(t *testing.T) {
	t.Parallel()

	t.Run("pop from array", func(t *testing.T) {
		t.Parallel()
		arr := []Value{NewNumber(1.0), NewNumber(2.0), NewNumber(3.0)}
		arrPtr := &arr
		args := map[string]any{"0": arrPtr}
		evaluator := &Evaluator{}
		result, err := builtinPop(evaluator, args)
		if err != nil {
			t.Errorf("error = %v", err)
		}
		if len(*arrPtr) != 2 {
			t.Errorf("array len = %d, want 2", len(*arrPtr))
		}
		if result != NewNumber(3.0) {
			t.Errorf("popped value incorrect")
		}
	})

	t.Run("pop from empty array", func(t *testing.T) {
		t.Parallel()
		arr := []Value{}
		arrPtr := &arr
		args := map[string]any{"0": arrPtr}
		evaluator := &Evaluator{}
		result, err := builtinPop(evaluator, args)
		if err != nil {
			t.Errorf("error = %v", err)
		}
		if result != nil {
			t.Errorf("expected nil")
		}
	})
}

func TestBuiltinShift(t *testing.T) {
	t.Parallel()

	t.Run("shift from array", func(t *testing.T) {
		t.Parallel()
		arr := []Value{NewNumber(1.0), NewNumber(2.0), NewNumber(3.0)}
		arrPtr := &arr
		args := map[string]any{"0": arrPtr}
		evaluator := &Evaluator{}
		result, err := builtinShift(evaluator, args)
		if err != nil {
			t.Errorf("error = %v", err)
		}
		if len(*arrPtr) != 2 {
			t.Errorf("array len = %d, want 2", len(*arrPtr))
		}
		if result != NewNumber(1.0) {
			t.Errorf("shifted value incorrect")
		}
	})

	t.Run("shift from empty array", func(t *testing.T) {
		t.Parallel()
		arr := []Value{}
		arrPtr := &arr
		args := map[string]any{"0": arrPtr}
		evaluator := &Evaluator{}
		result, err := builtinShift(evaluator, args)
		if err != nil {
			t.Errorf("error = %v", err)
		}
		if result != nil {
			t.Errorf("expected nil")
		}
	})
}

func TestBuiltinUnshift(t *testing.T) {
	t.Parallel()

	t.Run("unshift single item", func(t *testing.T) {
		t.Parallel()
		arr := []Value{NewNumber(2.0), NewNumber(3.0)}
		arrPtr := &arr
		args := map[string]any{"0": arrPtr, "1": 1.0}
		evaluator := &Evaluator{}
		result, err := builtinUnshift(evaluator, args)
		if err != nil {
			t.Errorf("error = %v", err)
		}
		if result != 3.0 {
			t.Errorf("got %v, want 3.0", result)
		}
		if len(*arrPtr) != 3 {
			t.Errorf("array len = %d, want 3", len(*arrPtr))
		}
	})

	t.Run("unshift multiple items", func(t *testing.T) {
		t.Parallel()
		arr := []Value{NewNumber(3.0)}
		arrPtr := &arr
		args := map[string]any{"0": arrPtr, "1": 1.0, "2": 2.0}
		evaluator := &Evaluator{}
		result, err := builtinUnshift(evaluator, args)
		if err != nil {
			t.Errorf("error = %v", err)
		}
		if result != 3.0 {
			t.Errorf("got %v, want 3.0", result)
		}
		if len(*arrPtr) != 3 {
			t.Errorf("array len = %d, want 3", len(*arrPtr))
		}
	})
}

func TestBuiltinJoin(t *testing.T) {
	t.Parallel()

	t.Run("join with comma", func(t *testing.T) {
		t.Parallel()
		arr := []Value{NewString("a"), NewString("b"), NewString("c")}
		arrPtr := &arr
		args := map[string]any{"0": arrPtr, "1": ","}
		evaluator := &Evaluator{}
		result, err := builtinJoin(evaluator, args)
		if err != nil {
			t.Errorf("error = %v", err)
		}
		if result != "a,b,c" {
			t.Errorf("got %q, want %q", result, "a,b,c")
		}
	})

	t.Run("join with space", func(t *testing.T) {
		t.Parallel()
		arr := []Value{NewString("hello"), NewString("world")}
		arrPtr := &arr
		args := map[string]any{"0": arrPtr, "1": " "}
		evaluator := &Evaluator{}
		result, err := builtinJoin(evaluator, args)
		if err != nil {
			t.Errorf("error = %v", err)
		}
		if result != "hello world" {
			t.Errorf("got %q, want %q", result, "hello world")
		}
	})

	t.Run("join empty array", func(t *testing.T) {
		t.Parallel()
		arr := []Value{}
		arrPtr := &arr
		args := map[string]any{"0": arrPtr, "1": ","}
		evaluator := &Evaluator{}
		result, err := builtinJoin(evaluator, args)
		if err != nil {
			t.Errorf("error = %v", err)
		}
		if result != "" {
			t.Errorf("got %q, want %q", result, "")
		}
	})

	t.Run("non-array", func(t *testing.T) {
		t.Parallel()
		args := map[string]any{"0": "text", "1": ","}
		evaluator := &Evaluator{}
		_, err := builtinJoin(evaluator, args)
		if err == nil {
			t.Error("expected error")
		}
	})
}
