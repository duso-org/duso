package runtime

import (
	"fmt"
	"math"
	mathrand "math/rand"
	"time"
)

// Math functions

// builtinFloor rounds down to nearest integer
func builtinFloor(evaluator *Evaluator, args map[string]any) (any, error) {
	if arg, ok := args["0"].(float64); ok {
		return math.Floor(arg), nil
	}
	return nil, fmt.Errorf("floor() requires a number")
}

// builtinCeil rounds up to nearest integer
func builtinCeil(evaluator *Evaluator, args map[string]any) (any, error) {
	if arg, ok := args["0"].(float64); ok {
		return math.Ceil(arg), nil
	}
	return nil, fmt.Errorf("ceil() requires a number")
}

// builtinRound rounds to nearest integer
func builtinRound(evaluator *Evaluator, args map[string]any) (any, error) {
	if arg, ok := args["0"].(float64); ok {
		return math.Round(arg), nil
	}
	return nil, fmt.Errorf("round() requires a number")
}

// builtinAbs returns absolute value
func builtinAbs(evaluator *Evaluator, args map[string]any) (any, error) {
	if arg, ok := args["0"].(float64); ok {
		return math.Abs(arg), nil
	}
	return nil, fmt.Errorf("abs() requires a number")
}

// minMaxHelper computes min/max of numeric arguments
func minMaxHelper(args map[string]any, isMin bool) (any, error) {
	if len(args) == 0 {
		name := "min()"
		if !isMin {
			name = "max()"
		}
		return nil, fmt.Errorf("%s requires at least one argument", name)
	}

	var result float64
	var set bool

	for i := 0; ; i++ {
		key := fmt.Sprintf("%d", i)
		val, ok := args[key].(float64)
		if !ok {
			break
		}
		if !set {
			result = val
			set = true
		} else if isMin && val < result {
			result = val
		} else if !isMin && val > result {
			result = val
		}
	}

	if !set {
		name := "min()"
		if !isMin {
			name = "max()"
		}
		return nil, fmt.Errorf("%s requires numeric arguments", name)
	}
	return result, nil
}

// builtinMin returns minimum of arguments
func builtinMin(evaluator *Evaluator, args map[string]any) (any, error) {
	return minMaxHelper(args, true)
}

// builtinMax returns maximum of arguments
func builtinMax(evaluator *Evaluator, args map[string]any) (any, error) {
	return minMaxHelper(args, false)
}

// builtinSqrt returns square root
func builtinSqrt(evaluator *Evaluator, args map[string]any) (any, error) {
	if arg, ok := args["0"].(float64); ok {
		return math.Sqrt(arg), nil
	}
	return nil, fmt.Errorf("sqrt() requires a number")
}

// builtinPow returns x^y
func builtinPow(evaluator *Evaluator, args map[string]any) (any, error) {
	x, ok := args["0"].(float64)
	if !ok {
		return nil, fmt.Errorf("pow() requires a number as first argument")
	}

	y, ok := args["1"].(float64)
	if !ok {
		return nil, fmt.Errorf("pow() requires a number as second argument")
	}

	return math.Pow(x, y), nil
}

// builtinClamp clamps value between min and max
func builtinClamp(evaluator *Evaluator, args map[string]any) (any, error) {
	val, ok := args["0"].(float64)
	if !ok {
		return nil, fmt.Errorf("clamp() requires a number as first argument")
	}

	min, ok := args["1"].(float64)
	if !ok {
		return nil, fmt.Errorf("clamp() requires a number as second argument (min)")
	}

	max, ok := args["2"].(float64)
	if !ok {
		return nil, fmt.Errorf("clamp() requires a number as third argument (max)")
	}

	if val < min {
		return min, nil
	}
	if val > max {
		return max, nil
	}
	return val, nil
}

// Trigonometric functions

// builtinSin returns sine of angle in radians
func builtinSin(evaluator *Evaluator, args map[string]any) (any, error) {
	if arg, ok := args["0"].(float64); ok {
		return math.Sin(arg), nil
	}
	return nil, fmt.Errorf("sin() requires a number (angle in radians)")
}

// builtinCos returns cosine of angle in radians
func builtinCos(evaluator *Evaluator, args map[string]any) (any, error) {
	if arg, ok := args["0"].(float64); ok {
		return math.Cos(arg), nil
	}
	return nil, fmt.Errorf("cos() requires a number (angle in radians)")
}

// builtinTan returns tangent of angle in radians
func builtinTan(evaluator *Evaluator, args map[string]any) (any, error) {
	if arg, ok := args["0"].(float64); ok {
		return math.Tan(arg), nil
	}
	return nil, fmt.Errorf("tan() requires a number (angle in radians)")
}

// builtinAsin returns arcsine in radians (inverse of sine)
func builtinAsin(evaluator *Evaluator, args map[string]any) (any, error) {
	if arg, ok := args["0"].(float64); ok {
		return math.Asin(arg), nil
	}
	return nil, fmt.Errorf("asin() requires a number between -1 and 1")
}

// builtinAcos returns arccosine in radians (inverse of cosine)
func builtinAcos(evaluator *Evaluator, args map[string]any) (any, error) {
	if arg, ok := args["0"].(float64); ok {
		return math.Acos(arg), nil
	}
	return nil, fmt.Errorf("acos() requires a number between -1 and 1")
}

// builtinAtan returns arctangent in radians (inverse of tangent)
func builtinAtan(evaluator *Evaluator, args map[string]any) (any, error) {
	if arg, ok := args["0"].(float64); ok {
		return math.Atan(arg), nil
	}
	return nil, fmt.Errorf("atan() requires a number")
}

// builtinAtan2 returns arctangent of y/x in radians, handling quadrants correctly
func builtinAtan2(evaluator *Evaluator, args map[string]any) (any, error) {
	y, ok := args["0"].(float64)
	if !ok {
		return nil, fmt.Errorf("atan2() requires a number as first argument (y)")
	}

	x, ok := args["1"].(float64)
	if !ok {
		return nil, fmt.Errorf("atan2() requires a number as second argument (x)")
	}

	return math.Atan2(y, x), nil
}

// Exponential and logarithmic functions

// builtinExp returns e^x
func builtinExp(evaluator *Evaluator, args map[string]any) (any, error) {
	if arg, ok := args["0"].(float64); ok {
		return math.Exp(arg), nil
	}
	return nil, fmt.Errorf("exp() requires a number")
}

// builtinLog returns logarithm base 10
func builtinLog(evaluator *Evaluator, args map[string]any) (any, error) {
	if arg, ok := args["0"].(float64); ok {
		return math.Log10(arg), nil
	}
	return nil, fmt.Errorf("log() requires a number")
}

// builtinLn returns natural logarithm (base e)
func builtinLn(evaluator *Evaluator, args map[string]any) (any, error) {
	if arg, ok := args["0"].(float64); ok {
		return math.Log(arg), nil
	}
	return nil, fmt.Errorf("ln() requires a number")
}

// builtinPi returns the mathematical constant pi
func builtinPi(evaluator *Evaluator, args map[string]any) (any, error) {
	return math.Pi, nil
}

// builtinRandom returns a random float between 0 and 1
func builtinRandom(evaluator *Evaluator, args map[string]any) (any, error) {
	rng := mathrand.New(mathrand.NewSource(time.Now().UnixNano()))
	return rng.Float64(), nil
}

