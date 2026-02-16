package script

import (
	"testing"
)

// TestEnvironmentDefineAndGet tests basic Define/Get operations
func TestEnvironmentDefineAndGet(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		varName  string
		varVal   Value
		wantErr  bool
	}{
		{"define number", "x", NewNumber(42), false},
		{"define string", "msg", NewString("hello"), false},
		{"define bool", "flag", NewBool(true), false},
		{"define nil", "empty", NewNil(), false},
		{"define array", "arr", NewArray([]Value{}), false},
		{"define object", "obj", NewObject(map[string]Value{}), false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			env := NewEnvironment()
			env.Define(tt.varName, tt.varVal)

			v, err := env.Get(tt.varName)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get error = %v, wantErr %v", err != nil, tt.wantErr)
			}
			if err == nil && v.Type != tt.varVal.Type {
				t.Errorf("value type mismatch: got %v, want %v", v.Type, tt.varVal.Type)
			}
		})
	}
}

// TestEnvironmentSetUpdates tests Set operation on existing variables
func TestEnvironmentSetUpdates(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		initial   Value
		updated   Value
		wantErr   bool
		isNew     bool
	}{
		{"update number", NewNumber(10), NewNumber(20), false, false},
		{"update string", NewString("old"), NewString("new"), false, false},
		{"update to different type", NewNumber(5), NewString("five"), false, false},
		{"set nonexistent", NewNumber(0), NewNumber(42), false, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			env := NewEnvironment()

			if !tt.isNew {
				env.Define("var", tt.initial)
			}

			err := env.Set("var", tt.updated)
			if (err != nil) != tt.wantErr {
				t.Errorf("Set error = %v, wantErr %v", err != nil, tt.wantErr)
			}
		})
	}
}

// TestChildEnvironmentScoping tests parent/child scoping
func TestChildEnvironmentScoping(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		parentDef     string
		childDef      string
		childCanSee   bool
		parentCanSee  bool
	}{
		{
			"child sees parent",
			"x",
			"",
			true,  // child can see parent's x
			false, // parent can't see child's undefined var
		},
		{
			"child defines different var",
			"x",
			"y",
			true,  // child can see parent's x
			false, // parent doesn't see child-only var y
		},
		{
			"child-only var",
			"",
			"z",
			true,  // we know z exists in child
			false, // parent doesn't see child vars
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			parent := NewEnvironment()
			if tt.parentDef != "" {
				parent.Define(tt.parentDef, NewNumber(1))
			}

			child := NewChildEnvironment(parent)
			if tt.childDef != "" {
				child.Define(tt.childDef, NewNumber(2))
			}

			// Check child can access parent vars
			if tt.childCanSee && tt.parentDef != "" {
				_, err := child.Get(tt.parentDef)
				if err != nil {
					t.Errorf("child should see parent var")
				}
			}

			// Check parent access to child vars (errors if parentCanSee=false)
			if tt.childDef != "" {
				_, err := parent.Get(tt.childDef)
				// If parentCanSee is false, we expect error; if true, we expect no error
				hasError := err != nil
				shouldHaveError := !tt.parentCanSee
				if hasError != shouldHaveError {
					t.Errorf("parent access child var: hasError=%v, shouldHaveError=%v", hasError, shouldHaveError)
				}
			}
		})
	}
}

// TestNestedEnvironmentChain tests multiple levels of nesting
func TestNestedEnvironmentChain(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		levels    int
		varName   string
		canFind   bool
	}{
		{"2 levels", 2, "x", true},
		{"3 levels", 3, "x", true},
		{"4 levels", 4, "y", true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Create root with variable
			root := NewEnvironment()
			root.Define(tt.varName, NewNumber(1))

			// Create chain
			current := root
			for i := 1; i < tt.levels; i++ {
				current = NewChildEnvironment(current)
				current.Define("level"+string(rune(48+i)), NewNumber(float64(i)))
			}

			// Verify we can find root variable from deepest level
			_, err := current.Get(tt.varName)
			if (err == nil) != tt.canFind {
				t.Errorf("find root var: err=%v, canFind=%v", err != nil, tt.canFind)
			}
		})
	}
}

// TestEnvironmentVariableShadowing tests shadowing behavior
func TestEnvironmentVariableShadowing(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		shadow    bool
		parentVal float64
		childVal  float64
	}{
		{"no shadow", false, 10, 20},
		{"shadow same name", true, 100, 200},
		{"modify parent", false, 50, 75},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			parent := NewEnvironment()
			parent.Define("x", NewNumber(tt.parentVal))

			child := NewChildEnvironment(parent)
			if tt.shadow {
				child.Define("x", NewNumber(tt.childVal))
			} else {
				// Don't shadow, but can modify parent's x through Set
				child.Set("x", NewNumber(tt.childVal))
			}

			// Check parent value
			parentX, _ := parent.Get("x")
			if !tt.shadow && parentX.AsNumber() != tt.childVal {
				t.Errorf("parent.x = %v, want %v", parentX.AsNumber(), tt.childVal)
			}

			// Check child value
			childX, _ := child.Get("x")
			if tt.shadow && childX.AsNumber() != tt.childVal {
				t.Errorf("child.x = %v, want %v (shadowed)", childX.AsNumber(), tt.childVal)
			}
		})
	}
}

// TestEnvironmentTypes tests different environment types
func TestEnvironmentTypes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		create    func(*Environment) *Environment
		isFunc    bool
	}{
		{
			"child environment",
			func(p *Environment) *Environment { return NewChildEnvironment(p) },
			false,
		},
		{
			"function environment",
			func(p *Environment) *Environment { return NewFunctionEnvironment(p) },
			true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			parent := NewEnvironment()
			parent.Define("x", NewNumber(10))

			env := tt.create(parent)

			// Both should be able to read parent
			_, err := env.Get("x")
			if err != nil {
				t.Errorf("can't read parent variable")
			}
		})
	}
}

// TestEnvironmentSetLocal tests SetLocal isolation
func TestEnvironmentSetLocal(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		setup    func(*Environment, *Environment)
		wantErr  bool
	}{
		{
			"setlocal on own var",
			func(p, c *Environment) {
				c.Define("y", NewNumber(20))
			},
			false,
		},
		{
			"setlocal updates current scope",
			func(p, c *Environment) {
				c.Define("y", NewNumber(30))
			},
			false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			parent := NewEnvironment()
			parent.Define("x", NewNumber(10))

			child := NewChildEnvironment(parent)
			tt.setup(parent, child)

			err := child.SetLocal("y", NewNumber(25))
			if (err != nil) != tt.wantErr {
				t.Errorf("SetLocal error = %v, wantErr %v", err != nil, tt.wantErr)
			}
		})
	}
}

// TestEnvironmentInitialization tests environment initialization state
func TestEnvironmentInitialization(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		create func() *Environment
		check  func(*Environment) bool
	}{
		{
			"root environment",
			func() *Environment { return NewEnvironment() },
			func(e *Environment) bool { return e.parent == nil },
		},
		{
			"child has parent",
			func() *Environment {
				p := NewEnvironment()
				return NewChildEnvironment(p)
			},
			func(e *Environment) bool { return e.parent != nil },
		},
		{
			"function has parent",
			func() *Environment {
				p := NewEnvironment()
				return NewFunctionEnvironment(p)
			},
			func(e *Environment) bool { return e.parent != nil },
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			env := tt.create()
			if !tt.check(env) {
				t.Errorf("initialization check failed")
			}
		})
	}
}

// TestEnvironmentComplexValues tests storing complex values
func TestEnvironmentComplexValues(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		value Value
	}{
		{"array", NewArray([]Value{NewNumber(1), NewNumber(2)})},
		{"object", NewObject(map[string]Value{"a": NewNumber(1)})},
		{"nested array", NewArray([]Value{NewArray([]Value{NewNumber(1)})})},
		{"nested object", NewObject(map[string]Value{
			"inner": NewObject(map[string]Value{"x": NewNumber(1)}),
		})},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			env := NewEnvironment()
			env.Define("val", tt.value)

			retrieved, err := env.Get("val")
			if err != nil {
				t.Errorf("failed to retrieve: %v", err)
			}
			if retrieved.Type != tt.value.Type {
				t.Errorf("value type mismatch: got %v, want %v", retrieved.Type, tt.value.Type)
			}
		})
	}
}

// TestEnvironmentMultipleVariables tests storing many variables
func TestEnvironmentMultipleVariables(t *testing.T) {
	t.Parallel()
	env := NewEnvironment()

	// Define many variables
	varCount := 100
	for i := 0; i < varCount; i++ {
		env.Define("var"+string(rune(48+i%10)), NewNumber(float64(i)))
	}

	// Verify we can retrieve them
	for i := 0; i < varCount; i++ {
		_, err := env.Get("var" + string(rune(48+i%10)))
		if err != nil {
			t.Errorf("failed to get var at index %d", i)
		}
	}
}
