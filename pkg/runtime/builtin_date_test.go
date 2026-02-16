package runtime

import (
	"testing"
	"time"
)

func TestBuiltinNow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args map[string]any
	}{
		{"no args returns seconds", map[string]any{}},
		{"false returns seconds", map[string]any{"0": false}},
		{"true returns milliseconds", map[string]any{"0": true}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			evaluator := &Evaluator{}
			result, err := builtinNow(evaluator, tt.args)
			if err != nil {
				t.Errorf("error = %v", err)
			}

			timestamp := result.(float64)
			currentUnix := float64(time.Now().Unix())
			currentUnixMs := float64(time.Now().UnixMilli())

			if tt.args["0"] == true {
				// Milliseconds - should be ~1000x bigger
				if timestamp < currentUnixMs-1000 || timestamp > currentUnixMs+1000 {
					t.Errorf("timestamp %v not close to current millis %v", timestamp, currentUnixMs)
				}
			} else {
				// Seconds - should be close to current Unix
				if timestamp < currentUnix-1 || timestamp > currentUnix+1 {
					t.Errorf("timestamp %v not close to current Unix %v", timestamp, currentUnix)
				}
			}
		})
	}
}

func TestBuiltinFormatTime(t *testing.T) {
	t.Parallel()

	// Use a fixed timestamp for consistent testing
	timestamp := 1609459200.0 // 2021-01-01 00:00:00 UTC

	tests := []struct {
		name    string
		args    map[string]any
		want    string
		wantErr bool
	}{
		{"default format", map[string]any{"0": timestamp}, "2021-01-01 00:00:00", false},
		{"iso format", map[string]any{"0": timestamp, "1": "iso"}, "2021-01-01T00:00:00Z", false},
		{"date format", map[string]any{"0": timestamp, "1": "date"}, "2021-01-01", false},
		{"time format", map[string]any{"0": timestamp, "1": "time"}, "00:00:00", false},
		{"string timestamp", map[string]any{"0": "1609459200"}, "2021-01-01 00:00:00", false},
		{"non-number", map[string]any{"0": "not-a-number"}, "", true},
		{"missing argument", map[string]any{}, "", true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			evaluator := &Evaluator{}
			result, err := builtinFormatTime(evaluator, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && result != tt.want {
				t.Errorf("got %q, want %q", result, tt.want)
			}
		})
	}
}

func TestBuiltinParseTime(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		args      map[string]any
		wantOK    bool
		wantErr   bool
	}{
		{"iso format", map[string]any{"0": "2021-01-01T00:00:00Z"}, true, false},
		{"iso no z", map[string]any{"0": "2021-01-01T00:00:00"}, true, false},
		{"default format", map[string]any{"0": "2021-01-01 00:00:00"}, true, false},
		{"date only", map[string]any{"0": "2021-01-01"}, true, false},
		{"long date", map[string]any{"0": "January 1, 2021"}, true, false},
		{"short date", map[string]any{"0": "Jan 1, 2021"}, true, false},
		{"custom format", map[string]any{"0": "01/01/2021", "1": "MM/DD/YYYY"}, true, false},
		{"invalid format", map[string]any{"0": "not-a-date"}, false, true},
		{"non-string", map[string]any{"0": 42.0}, false, true},
		{"missing argument", map[string]any{}, false, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			evaluator := &Evaluator{}
			result, err := builtinParseTime(evaluator, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && tt.wantOK {
				if _, ok := result.(float64); !ok {
					t.Errorf("result type = %T, want float64", result)
				}
			}
		})
	}
}
