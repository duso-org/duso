package runtime

import (
	"math"
	"testing"
)

func TestBuiltinFloor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    map[string]any
		want    float64
		wantErr bool
	}{
		{"3.7", map[string]any{"0": 3.7}, 3.0, false},
		{"3.2", map[string]any{"0": 3.2}, 3.0, false},
		{"negative", map[string]any{"0": -3.7}, -4.0, false},
		{"integer", map[string]any{"0": 5.0}, 5.0, false},
		{"zero", map[string]any{"0": 0.0}, 0.0, false},
		{"non-number", map[string]any{"0": "text"}, 0, true},
		{"no argument", map[string]any{}, 0, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			evaluator := &Evaluator{}
			result, err := builtinFloor(evaluator, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && result != tt.want {
				t.Errorf("got %v, want %v", result, tt.want)
			}
		})
	}
}

func TestBuiltinCeil(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    map[string]any
		want    float64
		wantErr bool
	}{
		{"3.2", map[string]any{"0": 3.2}, 4.0, false},
		{"3.7", map[string]any{"0": 3.7}, 4.0, false},
		{"negative", map[string]any{"0": -3.7}, -3.0, false},
		{"integer", map[string]any{"0": 5.0}, 5.0, false},
		{"zero", map[string]any{"0": 0.0}, 0.0, false},
		{"non-number", map[string]any{"0": "text"}, 0, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			evaluator := &Evaluator{}
			result, err := builtinCeil(evaluator, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && result != tt.want {
				t.Errorf("got %v, want %v", result, tt.want)
			}
		})
	}
}

func TestBuiltinRound(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    map[string]any
		want    float64
		wantErr bool
	}{
		{"3.2", map[string]any{"0": 3.2}, 3.0, false},
		{"3.5", map[string]any{"0": 3.5}, 4.0, false},
		{"3.7", map[string]any{"0": 3.7}, 4.0, false},
		{"negative", map[string]any{"0": -3.7}, -4.0, false},
		{"integer", map[string]any{"0": 5.0}, 5.0, false},
		{"non-number", map[string]any{"0": "text"}, 0, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			evaluator := &Evaluator{}
			result, err := builtinRound(evaluator, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && result != tt.want {
				t.Errorf("got %v, want %v", result, tt.want)
			}
		})
	}
}

func TestBuiltinAbs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    map[string]any
		want    float64
		wantErr bool
	}{
		{"positive", map[string]any{"0": 5.0}, 5.0, false},
		{"negative", map[string]any{"0": -5.0}, 5.0, false},
		{"zero", map[string]any{"0": 0.0}, 0.0, false},
		{"float", map[string]any{"0": -3.14}, 3.14, false},
		{"non-number", map[string]any{"0": "text"}, 0, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			evaluator := &Evaluator{}
			result, err := builtinAbs(evaluator, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && result != tt.want {
				t.Errorf("got %v, want %v", result, tt.want)
			}
		})
	}
}

func TestBuiltinMin(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    map[string]any
		want    float64
		wantErr bool
	}{
		{"two args", map[string]any{"0": 5.0, "1": 3.0}, 3.0, false},
		{"three args", map[string]any{"0": 5.0, "1": 3.0, "2": 8.0}, 3.0, false},
		{"one arg", map[string]any{"0": 5.0}, 5.0, false},
		{"negative", map[string]any{"0": -5.0, "1": 3.0}, -5.0, false},
		{"no args", map[string]any{}, 0, true},
		{"non-number", map[string]any{"0": "text"}, 0, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			evaluator := &Evaluator{}
			result, err := builtinMin(evaluator, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && result != tt.want {
				t.Errorf("got %v, want %v", result, tt.want)
			}
		})
	}
}

func TestBuiltinMax(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    map[string]any
		want    float64
		wantErr bool
	}{
		{"two args", map[string]any{"0": 5.0, "1": 3.0}, 5.0, false},
		{"three args", map[string]any{"0": 5.0, "1": 3.0, "2": 8.0}, 8.0, false},
		{"one arg", map[string]any{"0": 5.0}, 5.0, false},
		{"negative", map[string]any{"0": -5.0, "1": -3.0}, -3.0, false},
		{"no args", map[string]any{}, 0, true},
		{"non-number", map[string]any{"0": "text"}, 0, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			evaluator := &Evaluator{}
			result, err := builtinMax(evaluator, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && result != tt.want {
				t.Errorf("got %v, want %v", result, tt.want)
			}
		})
	}
}

func TestBuiltinSqrt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    map[string]any
		want    float64
		wantErr bool
	}{
		{"4", map[string]any{"0": 4.0}, 2.0, false},
		{"9", map[string]any{"0": 9.0}, 3.0, false},
		{"0", map[string]any{"0": 0.0}, 0.0, false},
		{"2", map[string]any{"0": 2.0}, math.Sqrt(2.0), false},
		{"non-number", map[string]any{"0": "text"}, 0, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			evaluator := &Evaluator{}
			result, err := builtinSqrt(evaluator, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && result != tt.want {
				t.Errorf("got %v, want %v", result, tt.want)
			}
		})
	}
}

func TestBuiltinPow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    map[string]any
		want    float64
		wantErr bool
	}{
		{"2^3", map[string]any{"0": 2.0, "1": 3.0}, 8.0, false},
		{"3^2", map[string]any{"0": 3.0, "1": 2.0}, 9.0, false},
		{"5^0", map[string]any{"0": 5.0, "1": 0.0}, 1.0, false},
		{"2^-1", map[string]any{"0": 2.0, "1": -1.0}, 0.5, false},
		{"missing y", map[string]any{"0": 2.0}, 0, true},
		{"non-number x", map[string]any{"0": "text", "1": 3.0}, 0, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			evaluator := &Evaluator{}
			result, err := builtinPow(evaluator, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && result != tt.want {
				t.Errorf("got %v, want %v", result, tt.want)
			}
		})
	}
}

func TestBuiltinClamp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    map[string]any
		want    float64
		wantErr bool
	}{
		{"within range", map[string]any{"0": 5.0, "1": 0.0, "2": 10.0}, 5.0, false},
		{"below min", map[string]any{"0": -5.0, "1": 0.0, "2": 10.0}, 0.0, false},
		{"above max", map[string]any{"0": 15.0, "1": 0.0, "2": 10.0}, 10.0, false},
		{"at min", map[string]any{"0": 0.0, "1": 0.0, "2": 10.0}, 0.0, false},
		{"at max", map[string]any{"0": 10.0, "1": 0.0, "2": 10.0}, 10.0, false},
		{"missing max", map[string]any{"0": 5.0, "1": 0.0}, 0, true},
		{"non-number", map[string]any{"0": "text", "1": 0.0, "2": 10.0}, 0, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			evaluator := &Evaluator{}
			result, err := builtinClamp(evaluator, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && result != tt.want {
				t.Errorf("got %v, want %v", result, tt.want)
			}
		})
	}
}

func TestBuiltinSin(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    map[string]any
		want    float64
		wantErr bool
	}{
		{"0", map[string]any{"0": 0.0}, 0.0, false},
		{"pi/2", map[string]any{"0": math.Pi / 2}, 1.0, false},
		{"pi", map[string]any{"0": math.Pi}, 0.0, false},
		{"non-number", map[string]any{"0": "text"}, 0, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			evaluator := &Evaluator{}
			result, err := builtinSin(evaluator, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil {
				res := result.(float64)
				if math.IsNaN(res) && math.IsNaN(tt.want) {
					return
				}
				if math.Abs(res-tt.want) > 1e-9 {
					t.Errorf("got %v, want %v", res, tt.want)
				}
			}
		})
	}
}

func TestBuiltinCos(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    map[string]any
		want    float64
		wantErr bool
	}{
		{"0", map[string]any{"0": 0.0}, 1.0, false},
		{"pi/2", map[string]any{"0": math.Pi / 2}, 0.0, false},
		{"pi", map[string]any{"0": math.Pi}, -1.0, false},
		{"non-number", map[string]any{"0": "text"}, 0, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			evaluator := &Evaluator{}
			result, err := builtinCos(evaluator, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil {
				res := result.(float64)
				if math.Abs(res-tt.want) > 1e-9 {
					t.Errorf("got %v, want %v", res, tt.want)
				}
			}
		})
	}
}

func TestBuiltinTan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    map[string]any
		want    float64
		wantErr bool
	}{
		{"0", map[string]any{"0": 0.0}, 0.0, false},
		{"pi/4", map[string]any{"0": math.Pi / 4}, 1.0, false},
		{"non-number", map[string]any{"0": "text"}, 0, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			evaluator := &Evaluator{}
			result, err := builtinTan(evaluator, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil {
				res := result.(float64)
				if math.Abs(res-tt.want) > 1e-9 {
					t.Errorf("got %v, want %v", res, tt.want)
				}
			}
		})
	}
}

func TestBuiltinPi(t *testing.T) {
	t.Parallel()

	evaluator := &Evaluator{}
	result, err := builtinPi(evaluator, map[string]any{})
	if err != nil {
		t.Errorf("error = %v", err)
	}
	res := result.(float64)
	if res != math.Pi {
		t.Errorf("got %v, want %v", res, math.Pi)
	}
}
