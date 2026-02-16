package script

import (
	"testing"
)

// TestValueConstructors tests all Value constructors
func TestValueConstructors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		create    func() Value
		checkType func(Value) bool
		checkVal  any
	}{
		{
			name:      "NewNumber",
			create:    func() Value { return NewNumber(42.5) },
			checkType: func(v Value) bool { return v.IsNumber() },
			checkVal:  42.5,
		},
		{
			name:      "NewString",
			create:    func() Value { return NewString("hello") },
			checkType: func(v Value) bool { return v.IsString() },
			checkVal:  "hello",
		},
		{
			name:      "NewBool true",
			create:    func() Value { return NewBool(true) },
			checkType: func(v Value) bool { return v.IsBool() },
			checkVal:  true,
		},
		{
			name:      "NewBool false",
			create:    func() Value { return NewBool(false) },
			checkType: func(v Value) bool { return v.IsBool() },
			checkVal:  false,
		},
		{
			name:      "NewNil",
			create:    func() Value { return NewNil() },
			checkType: func(v Value) bool { return v.IsNil() },
			checkVal:  nil,
		},
		{
			name: "NewArray",
			create: func() Value {
				return NewArray([]Value{NewNumber(1), NewNumber(2)})
			},
			checkType: func(v Value) bool { return v.IsArray() },
			checkVal:  2, // length
		},
		{
			name: "NewObject",
			create: func() Value {
				return NewObject(map[string]Value{"a": NewNumber(1)})
			},
			checkType: func(v Value) bool { return v.IsObject() },
			checkVal:  1, // length
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := tt.create()
			if !tt.checkType(v) {
				t.Errorf("type check failed for %s", tt.name)
			}
		})
	}
}

// TestValueTypeChecks tests all type checking methods
func TestValueTypeChecks(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		value    Value
		isNum    bool
		isStr    bool
		isBool   bool
		isNil    bool
		isArr    bool
		isObj    bool
	}{
		{"number", NewNumber(42), true, false, false, false, false, false},
		{"string", NewString("hi"), false, true, false, false, false, false},
		{"bool", NewBool(true), false, false, true, false, false, false},
		{"nil", NewNil(), false, false, false, true, false, false},
		{"array", NewArray([]Value{}), false, false, false, false, true, false},
		{"object", NewObject(map[string]Value{}), false, false, false, false, false, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.value.IsNumber() != tt.isNum {
				t.Errorf("IsNumber() = %v, want %v", tt.value.IsNumber(), tt.isNum)
			}
			if tt.value.IsString() != tt.isStr {
				t.Errorf("IsString() = %v, want %v", tt.value.IsString(), tt.isStr)
			}
			if tt.value.IsBool() != tt.isBool {
				t.Errorf("IsBool() = %v, want %v", tt.value.IsBool(), tt.isBool)
			}
			if tt.value.IsNil() != tt.isNil {
				t.Errorf("IsNil() = %v, want %v", tt.value.IsNil(), tt.isNil)
			}
			if tt.value.IsArray() != tt.isArr {
				t.Errorf("IsArray() = %v, want %v", tt.value.IsArray(), tt.isArr)
			}
			if tt.value.IsObject() != tt.isObj {
				t.Errorf("IsObject() = %v, want %v", tt.value.IsObject(), tt.isObj)
			}
		})
	}
}

// TestValueAsConversions tests As* accessor methods (type-specific getters)
func TestValueAsConversions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		value  Value
		asNum  float64
		asStr  string
		asBool bool
	}{
		{"number 42", NewNumber(42), 42.0, "", false},
		{"number 0", NewNumber(0), 0, "", false},
		{"string hello", NewString("hello"), 0, "hello", false},
		{"empty string", NewString(""), 0, "", false},
		{"true", NewBool(true), 0, "", true},
		{"false", NewBool(false), 0, "", false},
		{"nil", NewNil(), 0, "", false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.value.AsNumber() != tt.asNum {
				t.Errorf("AsNumber() = %v, want %v", tt.value.AsNumber(), tt.asNum)
			}
			if tt.value.AsString() != tt.asStr {
				t.Errorf("AsString() = %q, want %q", tt.value.AsString(), tt.asStr)
			}
			if tt.value.AsBool() != tt.asBool {
				t.Errorf("AsBool() = %v, want %v", tt.value.AsBool(), tt.asBool)
			}
		})
	}
}

// TestValueCollections tests array and object operations
func TestValueCollections(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		create    func() Value
		checkLen  int
		checkElem func(Value) bool
	}{
		{
			name: "empty array",
			create: func() Value {
				return NewArray([]Value{})
			},
			checkLen: 0,
		},
		{
			name: "array with numbers",
			create: func() Value {
				return NewArray([]Value{NewNumber(1), NewNumber(2), NewNumber(3)})
			},
			checkLen:  3,
			checkElem: func(v Value) bool { return v.IsNumber() },
		},
		{
			name: "empty object",
			create: func() Value {
				return NewObject(map[string]Value{})
			},
			checkLen: 0,
		},
		{
			name: "object with mixed values",
			create: func() Value {
				return NewObject(map[string]Value{
					"name": NewString("Alice"),
					"age":  NewNumber(30),
					"admin": NewBool(true),
				})
			},
			checkLen: 3,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := tt.create()

			if v.IsArray() {
				if len(v.AsArray()) != tt.checkLen {
					t.Errorf("array len = %d, want %d", len(v.AsArray()), tt.checkLen)
				}
			} else if v.IsObject() {
				if len(v.AsObject()) != tt.checkLen {
					t.Errorf("object len = %d, want %d", len(v.AsObject()), tt.checkLen)
				}
			}
		})
	}
}

// TestNestedStructures tests nested arrays and objects
func TestNestedStructures(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		create func() Value
		check  func(Value) bool
	}{
		{
			name: "nested array",
			create: func() Value {
				return NewArray([]Value{
					NewArray([]Value{NewNumber(1), NewNumber(2)}),
					NewArray([]Value{NewNumber(3), NewNumber(4)}),
				})
			},
			check: func(v Value) bool {
				arr := v.AsArray()
				return len(arr) == 2 && arr[0].IsArray()
			},
		},
		{
			name: "nested object",
			create: func() Value {
				return NewObject(map[string]Value{
					"person": NewObject(map[string]Value{
						"name": NewString("Alice"),
						"age":  NewNumber(30),
					}),
				})
			},
			check: func(v Value) bool {
				obj := v.AsObject()
				return len(obj) == 1 && obj["person"].IsObject()
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := tt.create()
			if !tt.check(v) {
				t.Errorf("structure check failed for %s", tt.name)
			}
		})
	}
}

// TestValueNumbers tests number values across different ranges
func TestValueNumbers(t *testing.T) {
	t.Parallel()
	tests := []float64{0, 1, -5, 3.14, 1e10, -1e-5, 9999999999}

	for _, num := range tests {
		num := num
		t.Run("number", func(t *testing.T) {
			t.Parallel()
			v := NewNumber(num)
			if !v.IsNumber() {
				t.Errorf("IsNumber() = false for %v", num)
			}
			if v.AsNumber() != num {
				t.Errorf("AsNumber() = %v, want %v", v.AsNumber(), num)
			}
		})
	}
}

// TestValueStrings tests string values with various content
func TestValueStrings(t *testing.T) {
	t.Parallel()
	tests := []string{"", "hello", "with spaces", "你好", "\n\t", "special!@#$%"}

	for _, s := range tests {
		s := s
		t.Run("string", func(t *testing.T) {
			t.Parallel()
			v := NewString(s)
			if !v.IsString() {
				t.Errorf("IsString() = false")
			}
			if v.AsString() != s {
				t.Errorf("AsString() = %q, want %q", v.AsString(), s)
			}
		})
	}
}

// TestValueStringRepresentation tests String() method
func TestValueStringRepresentation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		value    Value
		contains string
	}{
		{"number", NewNumber(42), "42"},
		{"string", NewString("hello"), "hello"},
		{"bool true", NewBool(true), "true"},
		{"bool false", NewBool(false), "false"},
		{"nil", NewNil(), "nil"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			str := tt.value.String()
			if str == "" {
				t.Errorf("String() returned empty")
			}
		})
	}
}
