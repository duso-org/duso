package script

import (
	"testing"
)

// TestValueConstructors tests Value constructor functions directly
func TestValueConstructors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		fn       func() Value
		validate func(t *testing.T, v Value)
	}{
		{
			name: "NewNil",
			fn:   NewNil,
			validate: func(t *testing.T, v Value) {
				if !v.IsNil() {
					t.Fatal("expected nil value")
				}
			},
		},
		{
			name: "NewNumber",
			fn: func() Value {
				return NewNumber(42.5)
			},
			validate: func(t *testing.T, v Value) {
				if !v.IsNumber() {
					t.Fatal("expected number")
				}
				if v.AsNumber() != 42.5 {
					t.Fatalf("expected 42.5, got %v", v.AsNumber())
				}
			},
		},
		{
			name: "NewString",
			fn: func() Value {
				return NewString("hello")
			},
			validate: func(t *testing.T, v Value) {
				if !v.IsString() {
					t.Fatal("expected string")
				}
				if v.AsString() != "hello" {
					t.Fatalf("expected 'hello', got %v", v.AsString())
				}
			},
		},
		{
			name: "NewBool true",
			fn: func() Value {
				return NewBool(true)
			},
			validate: func(t *testing.T, v Value) {
				if !v.IsBool() {
					t.Fatal("expected bool")
				}
				if !v.AsBool() {
					t.Fatal("expected true")
				}
			},
		},
		{
			name: "NewBool false",
			fn: func() Value {
				return NewBool(false)
			},
			validate: func(t *testing.T, v Value) {
				if !v.IsBool() {
					t.Fatal("expected bool")
				}
				if v.AsBool() {
					t.Fatal("expected false")
				}
			},
		},
		{
			name: "NewArray",
			fn: func() Value {
				return NewArray([]Value{NewNumber(1), NewNumber(2), NewNumber(3)})
			},
			validate: func(t *testing.T, v Value) {
				if !v.IsArray() {
					t.Fatal("expected array")
				}
				arr := v.AsArray()
				if len(arr) != 3 {
					t.Fatalf("expected length 3, got %d", len(arr))
				}
			},
		},
		{
			name: "NewArray empty",
			fn: func() Value {
				return NewArray([]Value{})
			},
			validate: func(t *testing.T, v Value) {
				if !v.IsArray() {
					t.Fatal("expected array")
				}
				arr := v.AsArray()
				if len(arr) != 0 {
					t.Fatalf("expected length 0, got %d", len(arr))
				}
			},
		},
		{
			name: "NewObject",
			fn: func() Value {
				return NewObject(map[string]Value{
					"a": NewNumber(1),
					"b": NewString("two"),
				})
			},
			validate: func(t *testing.T, v Value) {
				if !v.IsObject() {
					t.Fatal("expected object")
				}
				obj := v.AsObject()
				if len(obj) != 2 {
					t.Fatalf("expected 2 keys, got %d", len(obj))
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := tt.fn()
			tt.validate(t, v)
		})
	}
}

// TestValueTypeChecks tests Value type checking methods
func TestValueTypeChecks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		value     Value
		expectNil bool
		expectNum bool
		expectStr bool
		expectBool bool
		expectArr bool
		expectObj bool
		expectFn  bool
	}{
		{
			name:      "nil value",
			value:     NewNil(),
			expectNil: true,
		},
		{
			name:      "number value",
			value:     NewNumber(42),
			expectNum: true,
		},
		{
			name:      "string value",
			value:     NewString("test"),
			expectStr: true,
		},
		{
			name:      "bool value",
			value:     NewBool(true),
			expectBool: true,
		},
		{
			name:      "array value",
			value:     NewArray([]Value{}),
			expectArr: true,
		},
		{
			name:      "object value",
			value:     NewObject(map[string]Value{}),
			expectObj: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.expectNil != tt.value.IsNil() {
				t.Fatalf("IsNil mismatch: expected %v", tt.expectNil)
			}
			if tt.expectNum != tt.value.IsNumber() {
				t.Fatalf("IsNumber mismatch: expected %v", tt.expectNum)
			}
			if tt.expectStr != tt.value.IsString() {
				t.Fatalf("IsString mismatch: expected %v", tt.expectStr)
			}
			if tt.expectBool != tt.value.IsBool() {
				t.Fatalf("IsBool mismatch: expected %v", tt.expectBool)
			}
			if tt.expectArr != tt.value.IsArray() {
				t.Fatalf("IsArray mismatch: expected %v", tt.expectArr)
			}
			if tt.expectObj != tt.value.IsObject() {
				t.Fatalf("IsObject mismatch: expected %v", tt.expectObj)
			}
			if tt.expectFn != tt.value.IsFunction() {
				t.Fatalf("IsFunction mismatch: expected %v", tt.expectFn)
			}
		})
	}
}

// TestValueTruthiness tests IsTruthy for all value types
func TestValueTruthiness(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		value     Value
		expectTruthy bool
	}{
		{
			name:      "nil is falsy",
			value:     NewNil(),
			expectTruthy: false,
		},
		{
			name:      "zero is falsy",
			value:     NewNumber(0),
			expectTruthy: false,
		},
		{
			name:      "positive number is truthy",
			value:     NewNumber(42),
			expectTruthy: true,
		},
		{
			name:      "negative number is truthy",
			value:     NewNumber(-1),
			expectTruthy: true,
		},
		{
			name:      "empty string is falsy",
			value:     NewString(""),
			expectTruthy: false,
		},
		{
			name:      "non-empty string is truthy",
			value:     NewString("hello"),
			expectTruthy: true,
		},
		{
			name:      "false is falsy",
			value:     NewBool(false),
			expectTruthy: false,
		},
		{
			name:      "true is truthy",
			value:     NewBool(true),
			expectTruthy: true,
		},
		{
			name:      "empty array is falsy",
			value:     NewArray([]Value{}),
			expectTruthy: false,
		},
		{
			name:      "non-empty array is truthy",
			value:     NewArray([]Value{NewNumber(1)}),
			expectTruthy: true,
		},
		{
			name:      "empty object is falsy",
			value:     NewObject(map[string]Value{}),
			expectTruthy: false,
		},
		{
			name:      "non-empty object is truthy",
			value:     NewObject(map[string]Value{"a": NewNumber(1)}),
			expectTruthy: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			actual := tt.value.IsTruthy()
			if actual != tt.expectTruthy {
				t.Fatalf("expected %v, got %v", tt.expectTruthy, actual)
			}
		})
	}
}

// TestValueConversions tests AsX() methods with edge cases
func TestValueConversions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		value     Value
		checkFn   func(t *testing.T, v Value)
	}{
		{
			name:  "AsNumber on number",
			value: NewNumber(42.5),
			checkFn: func(t *testing.T, v Value) {
				if v.AsNumber() != 42.5 {
					t.Fatalf("expected 42.5, got %v", v.AsNumber())
				}
			},
		},
		{
			name:  "AsNumber on nil",
			value: NewNil(),
			checkFn: func(t *testing.T, v Value) {
				// Converting nil to number should give 0
				if v.AsNumber() != 0 {
					t.Fatalf("expected 0 for nil, got %v", v.AsNumber())
				}
			},
		},
		{
			name:  "AsString on string",
			value: NewString("hello"),
			checkFn: func(t *testing.T, v Value) {
				if v.AsString() != "hello" {
					t.Fatalf("expected 'hello', got %v", v.AsString())
				}
			},
		},
		{
			name:  "AsString on number returns empty",
			value: NewNumber(42),
			checkFn: func(t *testing.T, v Value) {
				// AsString() only returns value if type is actually string, else ""
				str := v.AsString()
				if str != "" {
					t.Fatalf("expected empty string, got '%s'", str)
				}
			},
		},
		{
			name:  "AsString on nil",
			value: NewNil(),
			checkFn: func(t *testing.T, v Value) {
				str := v.AsString()
				if str != "" && str != "<nil>" {
					t.Fatalf("expected empty or nil string, got %v", str)
				}
			},
		},
		{
			name:  "AsBool on bool",
			value: NewBool(true),
			checkFn: func(t *testing.T, v Value) {
				if !v.AsBool() {
					t.Fatal("expected true")
				}
			},
		},
		{
			name:  "AsArray on array",
			value: NewArray([]Value{NewNumber(1), NewNumber(2)}),
			checkFn: func(t *testing.T, v Value) {
				arr := v.AsArray()
				if len(arr) != 2 {
					t.Fatalf("expected length 2, got %d", len(arr))
				}
			},
		},
		{
			name:  "AsArrayPtr returns pointer",
			value: NewArray([]Value{NewNumber(1)}),
			checkFn: func(t *testing.T, v Value) {
				ptr := v.AsArrayPtr()
				if ptr == nil {
					t.Fatal("expected non-nil pointer")
				}
				if len(*ptr) != 1 {
					t.Fatalf("expected length 1, got %d", len(*ptr))
				}
			},
		},
		{
			name:  "AsObject on object",
			value: NewObject(map[string]Value{"a": NewNumber(1)}),
			checkFn: func(t *testing.T, v Value) {
				obj := v.AsObject()
				if len(obj) != 1 {
					t.Fatalf("expected 1 key, got %d", len(obj))
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.checkFn(t, tt.value)
		})
	}
}

// TestInterfaceConversions tests converting between Value and interface{}
func TestInterfaceConversions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		value      Value
		validateFn func(t *testing.T, iface any)
	}{
		{
			name:  "nil to interface",
			value: NewNil(),
			validateFn: func(t *testing.T, iface any) {
				if iface != nil {
					t.Fatalf("expected nil, got %v", iface)
				}
			},
		},
		{
			name:  "number to interface",
			value: NewNumber(42.5),
			validateFn: func(t *testing.T, iface any) {
				f, ok := iface.(float64)
				if !ok || f != 42.5 {
					t.Fatalf("expected 42.5, got %v", iface)
				}
			},
		},
		{
			name:  "string to interface",
			value: NewString("hello"),
			validateFn: func(t *testing.T, iface any) {
				s, ok := iface.(string)
				if !ok || s != "hello" {
					t.Fatalf("expected 'hello', got %v", iface)
				}
			},
		},
		{
			name:  "bool to interface",
			value: NewBool(true),
			validateFn: func(t *testing.T, iface any) {
				b, ok := iface.(bool)
				if !ok || !b {
					t.Fatalf("expected true, got %v", iface)
				}
			},
		},
		{
			name:  "array to interface returns pointer",
			value: NewArray([]Value{NewNumber(1), NewNumber(2)}),
			validateFn: func(t *testing.T, iface any) {
				// ValueToInterface returns pointer to array for in-place mutations
				arrPtr, ok := iface.(*[]Value)
				if !ok || arrPtr == nil || len(*arrPtr) != 2 {
					t.Fatalf("expected *[]Value with 2 elements, got %v", iface)
				}
			},
		},
		{
			name:  "object to interface",
			value: NewObject(map[string]Value{"a": NewNumber(1)}),
			validateFn: func(t *testing.T, iface any) {
				obj, ok := iface.(map[string]any)
				if !ok || len(obj) != 1 {
					t.Fatalf("expected object with 1 key, got %v", iface)
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			iface := ValueToInterface(tt.value)
			tt.validateFn(t, iface)
		})
	}
}

// TestInterfaceToValueConversion tests converting interface{} back to Value
func TestInterfaceToValueConversion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		iface      interface{}
		validateFn func(t *testing.T, v Value)
	}{
		{
			name:  "nil interface",
			iface: nil,
			validateFn: func(t *testing.T, v Value) {
				if !v.IsNil() {
					t.Fatal("expected nil value")
				}
			},
		},
		{
			name:  "float64 interface",
			iface: 42.5,
			validateFn: func(t *testing.T, v Value) {
				if !v.IsNumber() {
					t.Fatal("expected number")
				}
				if v.AsNumber() != 42.5 {
					t.Fatalf("expected 42.5, got %v", v.AsNumber())
				}
			},
		},
		{
			name:  "string interface",
			iface: "hello",
			validateFn: func(t *testing.T, v Value) {
				if !v.IsString() {
					t.Fatal("expected string")
				}
				if v.AsString() != "hello" {
					t.Fatalf("expected 'hello', got %v", v.AsString())
				}
			},
		},
		{
			name:  "bool interface",
			iface: true,
			validateFn: func(t *testing.T, v Value) {
				if !v.IsBool() {
					t.Fatal("expected bool")
				}
				if !v.AsBool() {
					t.Fatal("expected true")
				}
			},
		},
		{
			name:  "slice interface",
			iface: []interface{}{float64(1), float64(2)},
			validateFn: func(t *testing.T, v Value) {
				if !v.IsArray() {
					t.Fatal("expected array")
				}
				arr := v.AsArray()
				if len(arr) != 2 {
					t.Fatalf("expected length 2, got %d", len(arr))
				}
			},
		},
		{
			name:  "map interface",
			iface: map[string]interface{}{"a": float64(1)},
			validateFn: func(t *testing.T, v Value) {
				if !v.IsObject() {
					t.Fatal("expected object")
				}
				obj := v.AsObject()
				if len(obj) != 1 {
					t.Fatalf("expected 1 key, got %d", len(obj))
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := InterfaceToValue(tt.iface)
			tt.validateFn(t, v)
		})
	}
}

// TestDeepCopy tests DeepCopy functionality with various value types
func TestDeepCopy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		value      Value
		validateFn func(t *testing.T, original Value, copy Value)
	}{
		{
			name:  "copy nil",
			value: NewNil(),
			validateFn: func(t *testing.T, original Value, copy Value) {
				if !copy.IsNil() {
					t.Fatal("expected nil copy")
				}
			},
		},
		{
			name:  "copy number",
			value: NewNumber(42.5),
			validateFn: func(t *testing.T, original Value, copy Value) {
				if !copy.IsNumber() || copy.AsNumber() != 42.5 {
					t.Fatal("number copy failed")
				}
			},
		},
		{
			name:  "copy string",
			value: NewString("hello"),
			validateFn: func(t *testing.T, original Value, copy Value) {
				if !copy.IsString() || copy.AsString() != "hello" {
					t.Fatal("string copy failed")
				}
			},
		},
		{
			name:  "copy array",
			value: NewArray([]Value{NewNumber(1), NewNumber(2), NewNumber(3)}),
			validateFn: func(t *testing.T, original Value, copy Value) {
				if !copy.IsArray() {
					t.Fatal("expected array copy")
				}
				origArr := original.AsArray()
				copyArr := copy.AsArray()
				if len(copyArr) != len(origArr) {
					t.Fatal("array length mismatch")
				}
				// Modifying copy should not affect original
				copyPtr := copy.AsArrayPtr()
				(*copyPtr)[0] = NewNumber(999)
				if origArr[0].AsNumber() == 999 {
					t.Fatal("deep copy failed - original was modified")
				}
			},
		},
		{
			name: "copy object",
			value: NewObject(map[string]Value{
				"a": NewNumber(1),
				"b": NewString("hello"),
			}),
			validateFn: func(t *testing.T, original Value, copy Value) {
				if !copy.IsObject() {
					t.Fatal("expected object copy")
				}
				origObj := original.AsObject()
				copyObj := copy.AsObject()
				if len(copyObj) != len(origObj) {
					t.Fatal("object size mismatch")
				}
				// Modify copy
				copyObj["a"] = NewNumber(999)
				if origObj["a"].AsNumber() == 999 {
					t.Fatal("deep copy failed - original was modified")
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			copy := DeepCopy(tt.value)
			tt.validateFn(t, tt.value, copy)
		})
	}
}

// TestValueStringRepresentation tests Value.String() method
func TestValueStringRepresentation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		value     Value
		checkFn   func(t *testing.T, str string)
	}{
		{
			name:  "nil string",
			value: NewNil(),
			checkFn: func(t *testing.T, str string) {
				if str != "nil" {
					t.Fatalf("expected 'nil', got '%s'", str)
				}
			},
		},
		{
			name:  "number string with decimal",
			value: NewNumber(42.5),
			checkFn: func(t *testing.T, str string) {
				if str != "42.5" {
					t.Fatalf("expected '42.5', got '%s'", str)
				}
			},
		},
		{
			name:  "number string integer",
			value: NewNumber(42),
			checkFn: func(t *testing.T, str string) {
				if str != "42" && str != "42.0" {
					t.Fatalf("expected '42' or '42.0', got '%s'", str)
				}
			},
		},
		{
			name:  "string string",
			value: NewString("hello"),
			checkFn: func(t *testing.T, str string) {
				if str != "hello" {
					t.Fatalf("expected 'hello', got '%s'", str)
				}
			},
		},
		{
			name:  "bool true string",
			value: NewBool(true),
			checkFn: func(t *testing.T, str string) {
				if str != "true" {
					t.Fatalf("expected 'true', got '%s'", str)
				}
			},
		},
		{
			name:  "bool false string",
			value: NewBool(false),
			checkFn: func(t *testing.T, str string) {
				if str != "false" {
					t.Fatalf("expected 'false', got '%s'", str)
				}
			},
		},
		{
			name:  "array string",
			value: NewArray([]Value{NewNumber(1), NewNumber(2)}),
			checkFn: func(t *testing.T, str string) {
				if str == "" {
					t.Fatal("expected non-empty array string")
				}
			},
		},
		{
			name:  "object string",
			value: NewObject(map[string]Value{"a": NewNumber(1)}),
			checkFn: func(t *testing.T, str string) {
				if str == "" {
					t.Fatal("expected non-empty object string")
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			str := tt.value.String()
			tt.checkFn(t, str)
		})
	}
}
