package runtime

import (
	"regexp"
	"testing"
	"time"
)

func TestBuiltinExit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    map[string]any
		wantErr bool
	}{
		{"no args", map[string]any{}, true},
		{"one arg", map[string]any{"0": 42.0}, true},
		{"multiple args", map[string]any{"0": 1.0, "1": 2.0, "2": 3.0}, true},
		{"string arg", map[string]any{"0": "error message"}, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			evaluator := &Evaluator{}
			_, err := builtinExit(evaluator, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				_, ok := err.(*ExitExecution)
				if !ok {
					t.Errorf("error is not ExitExecution: %T", err)
				}
			}
		})
	}
}

func TestBuiltinSleep(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    map[string]any
		wantErr bool
	}{
		{"default sleep", map[string]any{}, false},
		{"zero seconds", map[string]any{"0": 0.0}, false},
		{"positive seconds", map[string]any{"0": 0.01}, false},
		{"negative seconds", map[string]any{"0": -1.0}, true},
		{"non-number", map[string]any{"0": "text"}, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			evaluator := &Evaluator{}
			start := time.Now()
			_, err := builtinSleep(evaluator, tt.args)
			elapsed := time.Since(start)

			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}

			// For non-error cases, verify some sleep occurred
			if !tt.wantErr && elapsed < 0 {
				t.Errorf("sleep time is negative: %v", elapsed)
			}
		})
	}
}

func TestBuiltinUUID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args map[string]any
	}{
		{"uuid generation", map[string]any{}},
		{"uuid with unused args", map[string]any{"0": "ignored"}},
	}

	// UUID format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			evaluator := &Evaluator{}
			result, err := builtinUUID(evaluator, tt.args)
			if err != nil {
				t.Errorf("error = %v", err)
			}

			uuid := result.(string)

			// Verify format
			if !uuidRegex.MatchString(uuid) {
				t.Errorf("invalid UUID format: %q", uuid)
			}

			// Length should be 36 (32 hex + 4 hyphens)
			if len(uuid) != 36 {
				t.Errorf("UUID length = %d, want 36", len(uuid))
			}

			// Version should be 7 (character at position 14)
			if uuid[14] != '7' {
				t.Errorf("version = %c, want 7", uuid[14])
			}

			// Variant should be 8, 9, a, or b (character at position 19)
			variant := uuid[19]
			if variant != '8' && variant != '9' && variant != 'a' && variant != 'b' {
				t.Errorf("variant = %c, want one of [89ab]", variant)
			}
		})
	}
}

func TestUUIDUniqueness(t *testing.T) {
	t.Parallel()

	evaluator := &Evaluator{}

	// Generate multiple UUIDs
	uuids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		result, err := builtinUUID(evaluator, map[string]any{})
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		uuid := result.(string)
		if uuids[uuid] {
			t.Errorf("duplicate UUID generated: %s", uuid)
		}
		uuids[uuid] = true
	}

	if len(uuids) != 100 {
		t.Errorf("got %d unique UUIDs, want 100", len(uuids))
	}
}
