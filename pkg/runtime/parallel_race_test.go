package runtime

import (
	"sync"
	"testing"

	"github.com/duso-org/duso/pkg/script"
)

var registerOnce sync.Once

// TestParallelParentScopeReads exercises the concurrency invariant that lets
// Environment run unsynchronized: parallel() branches read parent scope vars
// concurrently while writing to their own scopes, and the parent goroutine is
// blocked for the duration. Run with -race to verify.
func TestParallelParentScopeReads(t *testing.T) {
	registerOnce.Do(RegisterBuiltins)
	interp := script.NewInterpreter()

	src := `
base = 1000
other = 7
results = parallel([
  function()
    s = 0
    for i = 1, 5000 do
      s += base
    end
    return s
  end,
  function()
    s = 0
    for i = 1, 5000 do
      s += base + other
    end
    return s
  end,
  function()
    s = 0
    for i = 1, 5000 do
      s += other
    end
    return s
  end
])

if results[0] != 5000000 then
  throw("branch 0 expected 5000000, got " + results[0])
end
if results[1] != 5035000 then
  throw("branch 1 expected 5035000, got " + results[1])
end
if results[2] != 35000 then
  throw("branch 2 expected 35000, got " + results[2])
end
`
	if _, err := interp.Execute(src); err != nil {
		t.Fatalf("parallel script failed: %v", err)
	}
}
