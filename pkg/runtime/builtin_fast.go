package runtime

import (
	"fmt"
	"math"
	"strings"
	"unicode/utf8"

	"github.com/duso-org/duso/pkg/script"
)

// Fast-path ([]Value) variants of hot builtins. Each mirrors the semantics of
// its map-based counterpart exactly; the evaluator uses these for direct
// positional calls and falls back to the map form everywhere else.

func registerFastBuiltins() {
	script.RegisterBuiltinFast("push", fastPush)
	script.RegisterBuiltinFast("pop", fastPop)
	script.RegisterBuiltinFast("shift", fastShift)
	script.RegisterBuiltinFast("len", fastLen)
	script.RegisterBuiltinFast("keys", fastKeys)
	script.RegisterBuiltinFast("values", fastValues)
	script.RegisterBuiltinFast("join", fastJoin)
	script.RegisterBuiltinFast("floor", fastFloor)
	script.RegisterBuiltinFast("ceil", fastCeil)
	script.RegisterBuiltinFast("round", fastRound)
	script.RegisterBuiltinFast("abs", fastAbs)
	script.RegisterBuiltinFast("sqrt", fastSqrt)
	script.RegisterBuiltinFast("min", fastMin)
	script.RegisterBuiltinFast("max", fastMax)
	script.RegisterBuiltinFast("fibonacci", fastFibonacci)
}

func fastFibonacci(evaluator *Evaluator, args []Value) (Value, error) {
	if len(args) < 1 || !args[0].IsNumber() {
		return script.NewNil(), fmt.Errorf("fibonacci() requires a number")
	}
	n := int64(args[0].AsNumber())
	if n < 0 {
		return script.NewNil(), fmt.Errorf("fibonacci() requires a non-negative integer")
	}
	return script.NewNumber(float64(fibonacci(n))), nil
}

func fastPush(evaluator *Evaluator, args []Value) (Value, error) {
	if len(args) < 1 || !args[0].IsArray() {
		return script.NewNil(), fmt.Errorf("push() requires an array as first argument")
	}
	arrPtr := args[0].AsArrayPtr()
	*arrPtr = append(*arrPtr, args[1:]...)
	return script.NewNumber(float64(len(*arrPtr))), nil
}

func fastPop(evaluator *Evaluator, args []Value) (Value, error) {
	if len(args) < 1 || !args[0].IsArray() {
		return script.NewNil(), fmt.Errorf("pop() requires an array as first argument")
	}
	arrPtr := args[0].AsArrayPtr()
	arr := *arrPtr
	if len(arr) == 0 {
		return script.NewNil(), nil
	}
	last := arr[len(arr)-1]
	*arrPtr = arr[:len(arr)-1]
	return last, nil
}

func fastShift(evaluator *Evaluator, args []Value) (Value, error) {
	if len(args) < 1 || !args[0].IsArray() {
		return script.NewNil(), fmt.Errorf("shift() requires an array as first argument")
	}
	arrPtr := args[0].AsArrayPtr()
	arr := *arrPtr
	if len(arr) == 0 {
		return script.NewNil(), nil
	}
	first := arr[0]
	*arrPtr = arr[1:]
	return first, nil
}

func fastLen(evaluator *Evaluator, args []Value) (Value, error) {
	if len(args) < 1 {
		return script.NewNil(), fmt.Errorf("len() requires an argument")
	}
	v := args[0]
	switch {
	case v.IsNil():
		return script.NewNumber(0), nil
	case v.IsArray():
		return script.NewNumber(float64(len(v.AsArray()))), nil
	case v.IsObject():
		return script.NewNumber(float64(len(v.AsObject()))), nil
	case v.IsString():
		return script.NewNumber(float64(utf8.RuneCountInString(v.AsString()))), nil
	case v.IsBinary():
		if bin := v.AsBinary(); bin != nil && bin.Data != nil {
			return script.NewNumber(float64(len(*bin.Data))), nil
		}
	}
	return script.NewNil(), fmt.Errorf("len() requires array, object, string, or binary")
}

func fastKeys(evaluator *Evaluator, args []Value) (Value, error) {
	if len(args) < 1 || !args[0].IsObject() {
		return script.NewNil(), fmt.Errorf("keys() requires an object")
	}
	obj := args[0].AsObject()
	keys := make([]Value, 0, len(obj))
	for k := range obj {
		keys = append(keys, script.NewString(k))
	}
	return script.NewArray(keys), nil
}

func fastValues(evaluator *Evaluator, args []Value) (Value, error) {
	if len(args) < 1 || !args[0].IsObject() {
		return script.NewNil(), fmt.Errorf("values() requires an object")
	}
	obj := args[0].AsObject()
	values := make([]Value, 0, len(obj))
	for _, v := range obj {
		values = append(values, v)
	}
	return script.NewArray(values), nil
}

func fastJoin(evaluator *Evaluator, args []Value) (Value, error) {
	if len(args) < 1 || !args[0].IsArray() {
		return script.NewNil(), fmt.Errorf("join() requires an array as first argument")
	}
	if len(args) < 2 || !args[1].IsString() {
		return script.NewNil(), fmt.Errorf("join() requires a string separator as second argument")
	}
	arr := args[0].AsArray()
	parts := make([]string, len(arr))
	for i, item := range arr {
		parts[i] = item.String()
	}
	return script.NewString(strings.Join(parts, args[1].AsString())), nil
}

func fastNumArg(args []Value, name string) (float64, error) {
	if len(args) < 1 || !args[0].IsNumber() {
		return 0, fmt.Errorf("%s() requires a number", name)
	}
	return args[0].AsNumber(), nil
}

func fastFloor(evaluator *Evaluator, args []Value) (Value, error) {
	n, err := fastNumArg(args, "floor")
	if err != nil {
		return script.NewNil(), err
	}
	return script.NewNumber(math.Floor(n)), nil
}

func fastCeil(evaluator *Evaluator, args []Value) (Value, error) {
	n, err := fastNumArg(args, "ceil")
	if err != nil {
		return script.NewNil(), err
	}
	return script.NewNumber(math.Ceil(n)), nil
}

func fastRound(evaluator *Evaluator, args []Value) (Value, error) {
	n, err := fastNumArg(args, "round")
	if err != nil {
		return script.NewNil(), err
	}
	return script.NewNumber(math.Round(n)), nil
}

func fastAbs(evaluator *Evaluator, args []Value) (Value, error) {
	n, err := fastNumArg(args, "abs")
	if err != nil {
		return script.NewNil(), err
	}
	return script.NewNumber(math.Abs(n)), nil
}

func fastSqrt(evaluator *Evaluator, args []Value) (Value, error) {
	n, err := fastNumArg(args, "sqrt")
	if err != nil {
		return script.NewNil(), err
	}
	return script.NewNumber(math.Sqrt(n)), nil
}

func fastMinMax(args []Value, name string, isMin bool) (Value, error) {
	if len(args) == 0 {
		return script.NewNil(), fmt.Errorf("%s() requires at least one argument", name)
	}
	// Mirrors minMaxHelper: stop at the first non-number rather than erroring
	var best float64
	var set bool
	for _, v := range args {
		if !v.IsNumber() {
			break
		}
		n := v.AsNumber()
		if !set || (isMin && n < best) || (!isMin && n > best) {
			best = n
			set = true
		}
	}
	if !set {
		return script.NewNil(), fmt.Errorf("%s() requires numeric arguments", name)
	}
	return script.NewNumber(best), nil
}

func fastMin(evaluator *Evaluator, args []Value) (Value, error) {
	return fastMinMax(args, "min", true)
}

func fastMax(evaluator *Evaluator, args []Value) (Value, error) {
	return fastMinMax(args, "max", false)
}
