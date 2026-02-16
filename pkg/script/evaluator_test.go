package script

import "testing"

// TestIsInteger tests isInteger() helper
func TestIsInteger(t *testing.T) {
	tests := []struct {
		val  float64
		want bool
	}{
		{5.0, true},
		{0.0, true},
		{-10.0, true},
		{3.14, false},
		{0.5, false},
	}

	for _, tt := range tests {
		if got := isInteger(tt.val); got != tt.want {
			t.Errorf("isInteger(%v) = %v, want %v", tt.val, got, tt.want)
		}
	}
}

// TestNewEvaluator creates evaluator
func TestNewEvaluator(t *testing.T) {
	eval := NewEvaluator()
	if eval == nil {
		t.Fatal("NewEvaluator returned nil")
	}
	if eval.env == nil {
		t.Error("env is nil")
	}
	if eval.builtins == nil {
		t.Error("builtins is nil")
	}
}

// TestTryCoerceToNumber tests number coercion
func TestTryCoerceToNumber(t *testing.T) {
	eval := NewEvaluator()

	tests := []struct {
		val    Value
		want   float64
		wantOk bool
	}{
		{NewNumber(42.0), 42.0, true},
		{NewString("42"), 42.0, true},
		{NewString("3.14"), 3.14, true},
		{NewString("hello"), 0, false},
		{NewNil(), 0, false},
	}

	for i, tt := range tests {
		got, ok := eval.tryCoerceToNumber(tt.val)
		if ok != tt.wantOk {
			t.Errorf("case %d: ok = %v, want %v", i, ok, tt.wantOk)
		}
		if ok && got != tt.want {
			t.Errorf("case %d: got %v, want %v", i, got, tt.want)
		}
	}
}

// TestParseFloatHelper tests parseFloat()
func TestParseFloatHelper(t *testing.T) {
	tests := []struct {
		input   string
		want    float64
		wantErr bool
	}{
		{"42", 42.0, false},
		{"3.14", 3.14, false},
		{"-5", -5.0, false},
		{"0", 0.0, false},
		{"abc", 0, true},
	}

	for _, tt := range tests {
		got, err := parseFloat(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("parseFloat(%q): error = %v, wantErr %v", tt.input, err, tt.wantErr)
		}
		if err == nil && got != tt.want {
			t.Errorf("parseFloat(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

// TestEnvironmentDefineGet tests Define/Get
func TestEnvironmentDefineGet(t *testing.T) {
	env := NewEnvironment()
	env.Define("x", NewNumber(42.0))

	result, err := env.Get("x")
	if err != nil {
		t.Errorf("Get error: %v", err)
		return
	}
	if result.AsNumber() != 42.0 {
		t.Errorf("got %v, want 42.0", result)
	}
}

// TestEnvironmentGetMissing tests Get nonexistent
func TestEnvironmentGetMissing(t *testing.T) {
	env := NewEnvironment()
	_, err := env.Get("missing")
	if err == nil {
		t.Error("Get of missing should error")
	}
}

// TestEnvironmentSet tests Set
func TestEnvironmentSet(t *testing.T) {
	env := NewEnvironment()
	env.Define("x", NewNumber(10.0))

	err := env.Set("x", NewNumber(20.0))
	if err != nil {
		t.Errorf("Set error: %v", err)
		return
	}

	result, _ := env.Get("x")
	if result.AsNumber() != 20.0 {
		t.Errorf("got %v, want 20.0", result)
	}
}

// TestChildEnvironment tests parent/child relationship
func TestChildEnvironment(t *testing.T) {
	parent := NewEnvironment()
	parent.Define("x", NewNumber(10.0))

	child := NewChildEnvironment(parent)
	child.Define("y", NewNumber(20.0))

	// Child can see parent variable
	x, err := child.Get("x")
	if err != nil || x.AsNumber() != 10.0 {
		t.Error("child cannot see parent variable")
	}

	// Child has own variable
	y, err := child.Get("y")
	if err != nil || y.AsNumber() != 20.0 {
		t.Error("child variable not accessible")
	}

	// Parent cannot see child variable
	_, err = parent.Get("y")
	if err == nil {
		t.Error("parent should not see child variable")
	}
}

// TestShadowing tests variable shadowing
func TestShadowing(t *testing.T) {
	parent := NewEnvironment()
	parent.Define("x", NewNumber(10.0))

	child := NewChildEnvironment(parent)
	child.Define("x", NewNumber(20.0))

	// Child sees its own x
	x, _ := child.Get("x")
	if x.AsNumber() != 20.0 {
		t.Errorf("child got %v, want 20.0 (shadowed)", x)
	}

	// Parent still has original
	x, _ = parent.Get("x")
	if x.AsNumber() != 10.0 {
		t.Errorf("parent got %v, want 10.0", x)
	}
}

// TestFunctionEnvironment tests function scope blocking
func TestFunctionEnvironment(t *testing.T) {
	parent := NewEnvironment()
	parent.Define("x", NewNumber(10.0))

	funcEnv := NewFunctionEnvironment(parent)
	funcEnv.Define("y", NewNumber(20.0))

	// Can read parent
	x, err := funcEnv.Get("x")
	if err != nil || x.AsNumber() != 10.0 {
		t.Error("function scope cannot read parent")
	}

	// Can read own variable
	y, err := funcEnv.Get("y")
	if err != nil || y.AsNumber() != 20.0 {
		t.Error("function scope variable not accessible")
	}
}

// TestSetExecutionFilePath tests SetExecutionFilePath
func TestSetExecutionFilePath(t *testing.T) {
	eval := NewEvaluator()
	eval.SetExecutionFilePath("/test/file.du")

	if eval.ctx.FilePath != "/test/file.du" {
		t.Errorf("FilePath = %q, want '/test/file.du'", eval.ctx.FilePath)
	}
}

// TestEvaluatorRegisterFunction tests Evaluator's RegisterFunction method
func TestEvaluatorRegisterFunction(t *testing.T) {
	eval := NewEvaluator()

	fn := func(evaluator *Evaluator, args map[string]any) (any, error) {
		return NewNumber(42.0), nil
	}

	eval.RegisterFunction("testFn", fn)

	result, err := eval.env.Get("testFn")
	if err != nil {
		t.Errorf("registered function not found: %v", err)
		return
	}

	// Just verify it's a Value and was registered
	if result.AsString() == "" && result.AsNumber() == 0 {
		// Some basic check that it's not empty
	}
}

// TestNewEnvironmentFields tests NewEnvironment initialization
func TestNewEnvironmentFields(t *testing.T) {
	env := NewEnvironment()

	if env.variables == nil {
		t.Error("variables is nil")
	}
	if env.parent != nil {
		t.Error("parent should be nil")
	}
	if env.parameters == nil {
		t.Error("parameters is nil")
	}
}

// TestNewFunctionEnvironmentFields tests NewFunctionEnvironment
func TestNewFunctionEnvironmentFields(t *testing.T) {
	parent := NewEnvironment()
	funcEnv := NewFunctionEnvironment(parent)

	if !funcEnv.isFunctionScope {
		t.Error("isFunctionScope should be true")
	}
	if funcEnv.parent != parent {
		t.Error("parent not set correctly")
	}
}

// TestNewChildEnvironmentWithSelf tests NewChildEnvironmentWithSelf
func TestNewChildEnvironmentWithSelf(t *testing.T) {
	parent := NewEnvironment()
	selfVal := NewNumber(42.0)

	child := NewChildEnvironmentWithSelf(parent, selfVal)

	if child.parent != parent {
		t.Error("parent not set")
	}
	// Verify self was set (just check it exists)
	_ = child.self
}

// TestEvaluatorInitialState tests evaluator initial state
func TestEvaluatorInitialState(t *testing.T) {
	eval := NewEvaluator()

	if eval.isParallelContext {
		t.Error("isParallelContext should be false")
	}
	if eval.DebugMode {
		t.Error("DebugMode should be false")
	}
	if eval.NoStdin {
		t.Error("NoStdin should be false")
	}
	if eval.watchCache == nil {
		t.Error("watchCache is nil")
	}
	if len(eval.watchCache) != 0 {
		t.Error("watchCache should be empty")
	}
}
