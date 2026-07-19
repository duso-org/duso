package runtime

import (
	"testing"

	"github.com/duso-org/duso/pkg/script"
)

// TestCompoundAssignSingleEval verifies that compound assignment and
// increment/decrement evaluate side-effecting object/index expressions exactly
// once (a[f(i)] += 1 must call f once, not twice).
func TestCompoundAssignSingleEval(t *testing.T) {
	registerOnce.Do(RegisterBuiltins)

	src := `
calls = 0
function idx()
  calls += 1
  return 0
end

a = [10]
a[idx()] += 5
if calls != 1 then
  throw("compound assign: index evaluated " + calls + " times, want 1")
end
if a[0] != 15 then
  throw("compound assign: a[0] = " + a[0] + ", want 15")
end

calls = 0
a[idx()]++
if calls != 1 then
  throw("post-increment: index evaluated " + calls + " times, want 1")
end
if a[0] != 16 then
  throw("post-increment: a[0] = " + a[0] + ", want 16")
end

calls = 0
function obj()
  calls += 1
  return o
end
o = {n = 1}
obj().n += 2
if calls != 1 then
  throw("property compound assign: object evaluated " + calls + " times, want 1")
end
if o.n != 3 then
  throw("property compound assign: o.n = " + o.n + ", want 3")
end
`
	interp := script.NewInterpreter()
	if _, err := interp.Execute(src); err != nil {
		t.Fatalf("single-eval script failed: %v", err)
	}
}
