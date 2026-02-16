package runtime

import (
	"os"
	"testing"
)

func TestBuiltinEnv(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func()
		cleanup func()
		args    map[string]any
		want    string
		wantErr bool
	}{
		{
			name: "existing env var PATH",
			args: map[string]any{"0": "PATH"},
			want: os.Getenv("PATH"),
		},
		{
			name: "existing env var HOME",
			args: map[string]any{"0": "HOME"},
			want: os.Getenv("HOME"),
		},
		{
			name: "nonexistent var",
			args: map[string]any{"0": "DUSO_TEST_NONEXISTENT_VAR_XYZ"},
			want: "",
		},
		{
			name: "empty string var",
			args: map[string]any{"0": ""},
			want: "",
		},
		{
			name: "named argument",
			args: map[string]any{"varname": "PATH"},
			want: os.Getenv("PATH"),
		},
		{
			name:    "non-string input",
			args:    map[string]any{"0": 42.0},
			wantErr: true, // Requires string
		},
		{
			name:    "no argument",
			args:    map[string]any{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.setup != nil {
				tt.setup()
			}
			defer func() {
				if tt.cleanup != nil {
					tt.cleanup()
				}
			}()

			evaluator := &Evaluator{}
			result, err := builtinEnv(evaluator, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && result != tt.want {
				t.Errorf("got %q, want %q", result, tt.want)
			}
		})
	}
}

func TestEnvCustomVariable(t *testing.T) {
	t.Parallel()

	// Set a custom environment variable for testing
	testKey := "DUSO_TEST_CUSTOM_ENV_VAR"
	testValue := "test_value_12345"
	os.Setenv(testKey, testValue)
	defer os.Unsetenv(testKey)

	evaluator := &Evaluator{}
	result, err := builtinEnv(evaluator, map[string]any{"0": testKey})
	if err != nil {
		t.Errorf("error = %v", err)
	}
	if result != testValue {
		t.Errorf("got %q, want %q", result, testValue)
	}
}
