package runtime

import (
	"testing"
)

func TestIsInteger(t *testing.T) {
	tests := []struct {
		val  float64
		want bool
	}{
		{0, true},
		{1, true},
		{-5, true},
		{1e10, true},
		{3.14, false},
		{0.5, false},
		{1.0, true},
	}

	for _, tt := range tests {
		if IsInteger(tt.val) != tt.want {
			t.Errorf("IsInteger(%v) = %v, want %v", tt.val, IsInteger(tt.val), tt.want)
		}
	}
}

func TestInterfaceToValue(t *testing.T) {
	tests := []struct {
		input interface{}
		check func(Value) bool
	}{
		{42.0, func(v Value) bool { return v.IsNumber() }},
		{"hello", func(v Value) bool { return v.IsString() }},
		{true, func(v Value) bool { return v.IsBool() }},
		{false, func(v Value) bool { return v.IsBool() }},
		{nil, func(v Value) bool { return v.IsNil() }},
		{[]interface{}{}, func(v Value) bool { return v.IsArray() }},
		{map[string]interface{}{}, func(v Value) bool { return v.IsObject() }},
	}

	for _, tt := range tests {
		v := InterfaceToValue(tt.input)
		if !tt.check(v) {
			t.Errorf("InterfaceToValue(%T) check failed", tt.input)
		}
	}
}

func TestValueToInterface(t *testing.T) {
	tests := []struct {
		v     Value
		check func(interface{}) bool
	}{
		{NewNumber(42), func(i interface{}) bool { n, ok := i.(float64); return ok && n == 42 }},
		{NewString("hi"), func(i interface{}) bool { s, ok := i.(string); return ok && s == "hi" }},
		{NewBool(true), func(i interface{}) bool { b, ok := i.(bool); return ok && b }},
		{NewNil(), func(i interface{}) bool { return i == nil }},
	}

	for _, tt := range tests {
		v := ValueToInterface(tt.v)
		if !tt.check(v) {
			t.Errorf("ValueToInterface check failed")
		}
	}
}

func TestNewNumber(t *testing.T) {
	tests := []float64{0, 1, -5, 3.14, 1e10, -1e-5}
	for _, num := range tests {
		v := NewNumber(num)
		if v.AsNumber() != num {
			t.Errorf("NewNumber(%v) != %v", num, v.AsNumber())
		}
	}
}

func TestNewString(t *testing.T) {
	tests := []string{"", "hello", "with spaces", "你好", "\n\t"}
	for _, s := range tests {
		v := NewString(s)
		if v.AsString() != s {
			t.Errorf("NewString(%q) != %q", s, v.AsString())
		}
	}
}

func TestNewBool(t *testing.T) {
	v := NewBool(true)
	if !v.AsBool() {
		t.Errorf("NewBool(true).AsBool() = false")
	}

	v = NewBool(false)
	if v.AsBool() {
		t.Errorf("NewBool(false).AsBool() = true")
	}
}

func TestNewNil(t *testing.T) {
	v := NewNil()
	if !v.IsNil() {
		t.Errorf("NewNil().IsNil() = false")
	}
	if v.AsNumber() != 0 {
		t.Errorf("NewNil().AsNumber() != 0")
	}
}

func TestNewArray(t *testing.T) {
	arr := NewArray([]Value{NewNumber(1), NewString("two")})
	if !arr.IsArray() {
		t.Errorf("NewArray().IsArray() = false")
	}
	vals := arr.AsArray()
	if len(vals) != 2 {
		t.Errorf("array len = %d, want 2", len(vals))
	}
}

func TestNewObject(t *testing.T) {
	obj := NewObject(map[string]Value{"a": NewNumber(1)})
	if !obj.IsObject() {
		t.Errorf("NewObject().IsObject() = false")
	}
	vals := obj.AsObject()
	if len(vals) != 1 {
		t.Errorf("object len = %d, want 1", len(vals))
	}
}
