package runtime

import (
	"reflect"
	"testing"

	"github.com/duso-org/duso/pkg/script"
)

// Duso arrays reach builtins as *[]Value, including arrays nested in a config
// object. These helpers read them as duso values rather than Go slice types.

func TestRouteMethodArg(t *testing.T) {
	arr := []Value{script.NewString("GET"), script.NewString("post")}
	got, err := routeMethodArg(&arr)
	if err != nil || !reflect.DeepEqual(got, []string{"GET", "post"}) {
		t.Errorf("array: got (%#v, %v), want []string{GET, post}", got, err)
	}
	if got, _ := routeMethodArg("GET"); got != "GET" {
		t.Errorf("string: got %#v", got)
	}
	if got, _ := routeMethodArg(nil); got != nil {
		t.Errorf("nil: got %#v", got)
	}

	empty := []Value{}
	if _, err := routeMethodArg(&empty); err == nil {
		t.Error("empty array: want an error")
	}
	mixed := []Value{script.NewString("GET"), script.NewNumber(1)}
	if _, err := routeMethodArg(&mixed); err == nil {
		t.Error("non-string element: want an error")
	}
}

func TestStringsFromValue(t *testing.T) {
	arr := []Value{script.NewString("a"), script.NewNumber(1), script.NewString("b")}
	if got := stringsFromValue(InterfaceToValue(&arr)); !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Errorf("array: got %#v, want [a b]", got)
	}
	if got := stringsFromValue(InterfaceToValue("x")); !reflect.DeepEqual(got, []string{"x"}) {
		t.Errorf("string: got %#v, want [x]", got)
	}
	if got := stringsFromValue(InterfaceToValue(3.0)); got != nil {
		t.Errorf("number: got %#v, want nil", got)
	}
}
