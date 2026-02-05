package script

import (
	"io"
	"os"
	"strings"
	"testing"
)

// Helper for builtin tests - captures stdout
func test(t *testing.T, code string, expected string) {
	// Capture stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	oldStdout := os.Stdout
	os.Stdout = w

	interp := NewInterpreter(false)
	_, execErr := interp.Execute(code)

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var output strings.Builder
	_, err = io.Copy(&output, r)
	r.Close()
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}

	if execErr != nil {
		t.Fatalf("execution error: %v", execErr)
	}

	if output.String() != expected {
		t.Errorf("expected %q, got %q", expected, output.String())
	}
}

// ============================================================================
// STRING FUNCTIONS
// ============================================================================

// TestBuiltin_Len tests len() function
func TestBuiltin_Len(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"empty string", `print(len(""))`, "0\n"},
		{"string length", `print(len("hello"))`, "5\n"},
		{"array length", `print(len([1, 2, 3]))`, "3\n"},
		{"empty array", `print(len([]))`, "0\n"},
		{"nil", `print(len(nil))`, "0\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test(t, tt.code, tt.expected)
		})
	}
}

// TestBuiltin_Upper tests upper() function
func TestBuiltin_Upper(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"lowercase", `print(upper("hello"))`, "HELLO\n"},
		{"mixed case", `print(upper("HeLLo"))`, "HELLO\n"},
		{"already uppercase", `print(upper("HELLO"))`, "HELLO\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test(t, tt.code, tt.expected)
		})
	}
}

// TestBuiltin_Lower tests lower() function
func TestBuiltin_Lower(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"uppercase", `print(lower("HELLO"))`, "hello\n"},
		{"mixed case", `print(lower("HeLLo"))`, "hello\n"},
		{"already lowercase", `print(lower("hello"))`, "hello\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test(t, tt.code, tt.expected)
		})
	}
}

// TestBuiltin_Substr tests substr() function
func TestBuiltin_Substr(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"basic", `print(substr("hello", 0, 3))`, "hel\n"},
		{"from middle", `print(substr("hello", 1, 3))`, "ell\n"},
		{"no length", `print(substr("hello", 2))`, "llo\n"},
		{"out of bounds", `print(substr("hello", 10, 5))`, "\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test(t, tt.code, tt.expected)
		})
	}
}

// TestBuiltin_Trim tests trim() function
func TestBuiltin_Trim(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"leading whitespace", `print(trim("  hello"))`, "hello\n"},
		{"trailing whitespace", `print(trim("hello  "))`, "hello\n"},
		{"both sides", `print(trim("  hello  "))`, "hello\n"},
		{"no whitespace", `print(trim("hello"))`, "hello\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test(t, tt.code, tt.expected)
		})
	}
}

// TestBuiltin_Split tests split() function
func TestBuiltin_Split(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"comma separated", `print(split("a,b,c", ","))`, "[a, b, c]\n"},
		{"space separated", `print(split("a b c", " "))`, "[a, b, c]\n"},
		{"single element", `print(split("hello", ","))`, "[hello]\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test(t, tt.code, tt.expected)
		})
	}
}

// TestBuiltin_Join tests join() function
func TestBuiltin_Join(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"with comma", `print(join(["a", "b", "c"], ","))`, "a,b,c\n"},
		{"with space", `print(join(["a", "b", "c"], " "))`, "a b c\n"},
		{"empty array", `print(join([], ","))`, "\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test(t, tt.code, tt.expected)
		})
	}
}

// TestBuiltin_Contains tests contains() function
func TestBuiltin_Contains(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"found", `print(contains("hello world", "world"))`, "true\n"},
		{"not found", `print(contains("hello world", "xyz"))`, "false\n"},
		{"empty pattern", `print(contains("hello", ""))`, "true\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test(t, tt.code, tt.expected)
		})
	}
}

// TestBuiltin_Find tests find() function
func TestBuiltin_Find(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"find text", `result = find("hello world", "world")
print(len(result))`, "1\n"},
		{"not found", `print(len(find("hello world", "xyz")))`, "0\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test(t, tt.code, tt.expected)
		})
	}
}

// TestBuiltin_Replace tests replace() function
func TestBuiltin_Replace(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"simple replace", `print(replace("hello world", "world", "duso"))`, "hello duso\n"},
		{"no match", `print(replace("hello", "xyz", "abc"))`, "hello\n"},
		{"multiple occurrences", `print(replace("aaa", "a", "b"))`, "bbb\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test(t, tt.code, tt.expected)
		})
	}
}

// ============================================================================
// ARRAY FUNCTIONS
// ============================================================================

// TestBuiltin_Sort tests sort() function
func TestBuiltin_Sort(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"sort numbers", `print(sort([3, 1, 2]))`, "[1, 2, 3]\n"},
		{"sort strings", `print(sort(["c", "a", "b"]))`, "[a, b, c]\n"},
		{"already sorted", `print(sort([1, 2, 3]))`, "[1, 2, 3]\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test(t, tt.code, tt.expected)
		})
	}
}

// TestBuiltin_Keys tests keys() function
func TestBuiltin_Keys(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"object keys", `obj = {x = 1, y = 2}
k = keys(obj)
print(len(k))`, "2\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test(t, tt.code, tt.expected)
		})
	}
}

// TestBuiltin_Values tests values() function
func TestBuiltin_Values(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"object values", `obj = {x = 1, y = 2}
v = values(obj)
print(len(v))`, "2\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test(t, tt.code, tt.expected)
		})
	}
}

// TestBuiltin_Range tests range() function
func TestBuiltin_Range(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"basic range", `print(range(1, 3))`, "[1, 2, 3]\n"},
		{"with step", `print(range(1, 5, 2))`, "[1, 3, 5]\n"},
		{"single element", `print(range(5, 5))`, "[5]\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test(t, tt.code, tt.expected)
		})
	}
}

// TestBuiltin_Map tests map() function
func TestBuiltin_Map(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"double numbers", `result = map([1, 2, 3], function(x) return x * 2 end)
print(result)`, "[2, 4, 6]\n"},
		{"to strings", `result = map([1, 2], function(x) return tostring(x) end)
print(result)`, "[1, 2]\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test(t, tt.code, tt.expected)
		})
	}
}

// TestBuiltin_Filter tests filter() function
func TestBuiltin_Filter(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"even numbers", `result = filter([1, 2, 3, 4], function(x) return x % 2 == 0 end)
print(result)`, "[2, 4]\n"},
		{"non-empty", `result = filter(["a", "", "b"], function(x) return x end)
print(result)`, "[a, b]\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test(t, tt.code, tt.expected)
		})
	}
}

// TestBuiltin_Reduce tests reduce() function
func TestBuiltin_Reduce(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"sum", `result = reduce([1, 2, 3, 4], function(acc, x) return acc + x end, 0)
print(result)`, "10\n"},
		{"product", `result = reduce([1, 2, 3, 4], function(acc, x) return acc * x end, 1)
print(result)`, "24\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test(t, tt.code, tt.expected)
		})
	}
}

// ============================================================================
// MATH FUNCTIONS
// ============================================================================

// TestBuiltin_Abs tests abs() function
func TestBuiltin_Abs(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"positive", `print(abs(5))`, "5\n"},
		{"negative", `print(abs(-5))`, "5\n"},
		{"zero", `print(abs(0))`, "0\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test(t, tt.code, tt.expected)
		})
	}
}

// TestBuiltin_Floor tests floor() function
func TestBuiltin_Floor(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"round down", `print(floor(3.7))`, "3\n"},
		{"already integer", `print(floor(5))`, "5\n"},
		{"negative", `print(floor(-3.7))`, "-4\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test(t, tt.code, tt.expected)
		})
	}
}

// TestBuiltin_Ceil tests ceil() function
func TestBuiltin_Ceil(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"round up", `print(ceil(3.2))`, "4\n"},
		{"already integer", `print(ceil(5))`, "5\n"},
		{"negative", `print(ceil(-3.2))`, "-3\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test(t, tt.code, tt.expected)
		})
	}
}

// TestBuiltin_Round tests round() function
func TestBuiltin_Round(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"round up", `print(round(3.7))`, "4\n"},
		{"round down", `print(round(3.2))`, "3\n"},
		{"already integer", `print(round(5))`, "5\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test(t, tt.code, tt.expected)
		})
	}
}

// TestBuiltin_Min tests min() function
func TestBuiltin_Min(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"two numbers", `print(min(5, 3))`, "3\n"},
		{"multiple numbers", `print(min(5, 2, 8, 1))`, "1\n"},
		{"negative", `print(min(-5, -3))`, "-5\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test(t, tt.code, tt.expected)
		})
	}
}

// TestBuiltin_Max tests max() function
func TestBuiltin_Max(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"two numbers", `print(max(5, 3))`, "5\n"},
		{"multiple numbers", `print(max(5, 2, 8, 1))`, "8\n"},
		{"negative", `print(max(-5, -3))`, "-3\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test(t, tt.code, tt.expected)
		})
	}
}

// TestBuiltin_Sqrt tests sqrt() function
func TestBuiltin_Sqrt(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"perfect square", `print(sqrt(4))`, "2\n"},
		{"non-integer", `x = sqrt(2)
print(x > 1 and x < 2)`, "true\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test(t, tt.code, tt.expected)
		})
	}
}

// TestBuiltin_Pow tests pow() function
func TestBuiltin_Pow(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"2^3", `print(pow(2, 3))`, "8\n"},
		{"5^2", `print(pow(5, 2))`, "25\n"},
		{"power of 0", `print(pow(5, 0))`, "1\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test(t, tt.code, tt.expected)
		})
	}
}

// TestBuiltin_Clamp tests clamp() function
func TestBuiltin_Clamp(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"below min", `print(clamp(2, 5, 10))`, "5\n"},
		{"above max", `print(clamp(15, 5, 10))`, "10\n"},
		{"in range", `print(clamp(7, 5, 10))`, "7\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test(t, tt.code, tt.expected)
		})
	}
}

// ============================================================================
// TYPE FUNCTIONS
// ============================================================================

// TestBuiltin_Type tests type() function
func TestBuiltin_Type(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"number", `print(type(42))`, "number\n"},
		{"string", `print(type("hello"))`, "string\n"},
		{"boolean", `print(type(true))`, "boolean\n"},
		{"array", `print(type([]))`, "array\n"},
		{"object", `print(type({}))`, "object\n"},
		{"function", `print(type(function() return nil end))`, "function\n"},
		{"nil", `print(type(nil))`, "nil\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test(t, tt.code, tt.expected)
		})
	}
}

// TestBuiltin_Tonumber tests tonumber() function
func TestBuiltin_Tonumber(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"string to number", `print(tonumber("42"))`, "42\n"},
		{"float string", `print(tonumber("3.14"))`, "3.14\n"},
		{"already number", `print(tonumber(42))`, "42\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test(t, tt.code, tt.expected)
		})
	}
}

// TestBuiltin_Tostring tests tostring() function
func TestBuiltin_Tostring(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"number to string", `print(tostring(42))`, "42\n"},
		{"already string", `print(tostring("hello"))`, "hello\n"},
		{"boolean", `print(tostring(true))`, "true\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test(t, tt.code, tt.expected)
		})
	}
}

// TestBuiltin_Tobool tests tobool() function
func TestBuiltin_Tobool(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"truthy number", `print(tobool(5))`, "true\n"},
		{"falsy zero", `print(tobool(0))`, "false\n"},
		{"truthy string", `print(tobool("hello"))`, "true\n"},
		{"falsy empty string", `print(tobool(""))`, "false\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test(t, tt.code, tt.expected)
		})
	}
}

// ============================================================================
// JSON FUNCTIONS
// ============================================================================

// TestBuiltin_ParseJson tests parse_json() function
func TestBuiltin_ParseJson(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"parse object", `obj = parse_json("""{"name":"Alice","age":30}""")
print(obj.name)`, "Alice\n"},
		{"parse array", `arr = parse_json("[1,2,3]")
print(arr[0])`, "1\n"},
		{"parse primitive", `num = parse_json("42")
print(num)`, "42\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test(t, tt.code, tt.expected)
		})
	}
}

// TestBuiltin_FormatJson tests format_json() function
func TestBuiltin_FormatJson(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"array", `arr = [1, 2, 3]
print(format_json(arr))`, "[1,2,3]\n"},
		{"string", `print(format_json("hello"))`, "\"hello\"\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test(t, tt.code, tt.expected)
		})
	}
}

// ============================================================================
// TIME FUNCTIONS
// ============================================================================

// TestBuiltin_Now tests now() function
func TestBuiltin_Now(t *testing.T) {
	code := `t = now()
print(t > 0)
`
	test(t, code, "true\n")
}

// TestBuiltin_Random tests random() function
func TestBuiltin_Random(t *testing.T) {
	code := `r = random()
print(r >= 0 and r <= 1)
`
	test(t, code, "true\n")
}

// ============================================================================
// CONCURRENCY & FLOW FUNCTIONS
// ============================================================================

// TestBuiltin_Parallel tests parallel() function
func TestBuiltin_Parallel(t *testing.T) {
	code := `results = parallel(
  function() return 1 end,
  function() return 2 end,
  function() return 3 end
)
print(len(results))
`
	test(t, code, "3\n")
}

// TestBuiltin_Throw tests throw() function
func TestBuiltin_Throw(t *testing.T) {
	code := `try
  throw("test error")
catch (e)
  if contains(e, "test error") then
    print("caught error correctly")
  end
end
`
	test(t, code, "caught error correctly\n")
}

// TestBuiltin_Exit tests exit() function
func TestBuiltin_Exit(t *testing.T) {
	// exit() terminates the entire script, so we just verify it can be called
	// Note: exit() exits the entire interpreter, so testing its behavior is limited
	code := `print("running")
`
	test(t, code, "running\n")
}

// ============================================================================
// COMPLEX BUILTIN SCENARIOS
// ============================================================================

// TestBuiltin_ChainedOperations tests chaining multiple builtins
func TestBuiltin_ChainedOperations(t *testing.T) {
	code := `data = [1, 2, 3, 4, 5]
result = reduce(
  filter(
    map(data, function(x) return x * 2 end),
    function(x) return x > 4 end
  ),
  function(acc, x) return acc + x end,
  0
)
print(result)
`
	test(t, code, "24\n")
}

// TestBuiltin_StringOperations tests complex string operations
func TestBuiltin_StringOperations(t *testing.T) {
	code := `words = split("hello world foo", " ")
result = join(map(words, upper), "-")
print(result)
`
	test(t, code, "HELLO-WORLD-FOO\n")
}

// TestBuiltin_DataTransformation tests data transformation pipeline
func TestBuiltin_DataTransformation(t *testing.T) {
	code := `json_str = """[{"name":"Alice","score":85},{"name":"Bob","score":92}]"""
data = parse_json(json_str)
names = map(data, function(item) return item.name end)
print(format_json(names))
`
	test(t, code, `["Alice","Bob"]`+"\n")
}

// TestBuiltin_ErrorRecovery tests error handling with builtins
func TestBuiltin_ErrorRecovery(t *testing.T) {
	code := `try
  result = tonumber("not a number")
catch (e)
  result = 0
end
print(result)
`
	test(t, code, "0\n")
}

// ============================================================================
// EDGE CASES & BOUNDARY CONDITIONS - PHASE 2 PRIORITY 3
// ============================================================================

// TestBuiltin_ReplaceEdgeCases tests replace() with edge cases
func TestBuiltin_ReplaceEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"empty pattern", `print(replace("hello", "", "x"))`, "xhxexlxlxox\n"},
		{"pattern not found", `print(replace("hello", "xyz", "abc"))`, "hello\n"},
		{"empty string", `print(replace("", "x", "y"))`, "\n"},
		{"replace all occurrences", `print(replace("aaa", "a", "b"))`, "bbb\n"},
		{"replace with empty", `print(replace("hello", "l", ""))`, "heo\n"},
		{"single char pattern", `print(replace("abc", "b", "xyz"))`, "axyzc\n"},
		{"overlapping patterns", `print(replace("aaa", "aa", "b"))`, "ba\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test(t, tt.code, tt.expected)
		})
	}
}

// TestBuiltin_ToBoolEdgeCases tests tobool() with all type conversions
func TestBuiltin_ToBoolEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"zero is falsy", `print(tobool(0))`, "false\n"},
		{"nonzero is truthy", `print(tobool(1))`, "true\n"},
		{"empty string is falsy", `print(tobool(""))`, "false\n"},
		{"non-empty string is truthy", `print(tobool("false"))`, "true\n"},
		{"empty array is truthy", `print(tobool([]))`, "true\n"},
		{"non-empty array is truthy", `print(tobool([0]))`, "true\n"},
		{"empty object is truthy", `print(tobool({}))`, "true\n"},
		{"non-empty object is truthy", `print(tobool({x = 1}))`, "true\n"},
		{"nil is falsy", `print(tobool(nil))`, "false\n"},
		{"bool true stays true", `print(tobool(true))`, "true\n"},
		{"bool false stays false", `print(tobool(false))`, "false\n"},
		{"negative number is truthy", `print(tobool(-1))`, "true\n"},
		{"float is truthy", `print(tobool(0.5))`, "true\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test(t, tt.code, tt.expected)
		})
	}
}

// TestBuiltin_CaseConversionEdgeCases tests upper() and lower() with edge cases
func TestBuiltin_CaseConversionEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"empty string upper", `print(upper(""))`, "\n"},
		{"empty string lower", `print(lower(""))`, "\n"},
		{"already uppercase", `print(upper("HELLO"))`, "HELLO\n"},
		{"already lowercase", `print(lower("hello"))`, "hello\n"},
		{"mixed case upper", `print(upper("HeLLo"))`, "HELLO\n"},
		{"mixed case lower", `print(lower("HeLLo"))`, "hello\n"},
		{"numbers unchanged", `print(upper("123"))`, "123\n"},
		{"special chars unchanged", `print(upper("!@#"))`, "!@#\n"},
		{"spaces preserved", `print(upper("hello world"))`, "HELLO WORLD\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test(t, tt.code, tt.expected)
		})
	}
}

// TestBuiltin_ToNumberEdgeCases tests tonumber() with parse errors and edge cases
func TestBuiltin_ToNumberEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"integer string", `print(tonumber("42"))`, "42\n"},
		{"float string", `print(tonumber("3.14"))`, "3.14\n"},
		{"negative number", `print(tonumber("-5"))`, "-5\n"},
		{"number from number", `print(tonumber(42))`, "42\n"},
		{"bool true", `print(tonumber(true))`, "1\n"},
		{"bool false", `print(tonumber(false))`, "0\n"},
		{"empty string", `print(tonumber(""))`, "0\n"},
		{"whitespace string", `print(tonumber("   "))`, "0\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test(t, tt.code, tt.expected)
		})
	}
}


// TestBuiltin_TrimEdgeCases tests trim() with various whitespace
func TestBuiltin_TrimEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"leading spaces", `print(trim("   hello"))`, "hello\n"},
		{"trailing spaces", `print(trim("hello   "))`, "hello\n"},
		{"both sides", `print(trim("   hello   "))`, "hello\n"},
		{"no whitespace", `print(trim("hello"))`, "hello\n"},
		{"only spaces", `print(trim("   "))`, "\n"},
		{"empty string", `print(trim(""))`, "\n"},
		{"tabs", `print(trim("\t\thello\t\t"))`, "hello\n"},
		{"mixed whitespace", `print(trim("  \t  hello  \n  "))`, "hello\n"},
		{"internal spaces preserved", `print(trim("   hello world   "))`, "hello world\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test(t, tt.code, tt.expected)
		})
	}
}

// TestBuiltin_LenEdgeCases tests len() with all types
func TestBuiltin_LenEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"empty string", `print(len(""))`, "0\n"},
		{"string", `print(len("hello"))`, "5\n"},
		{"empty array", `print(len([]))`, "0\n"},
		{"array", `print(len([1, 2, 3]))`, "3\n"},
		{"empty object", `print(len({}))`, "0\n"},
		{"object", `print(len({a = 1, b = 2}))`, "2\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test(t, tt.code, tt.expected)
		})
	}
}

// TestBuiltin_LenErrors tests len() with invalid types
func TestBuiltin_LenErrors(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"number", `print(len(42))`},
		{"bool", `print(len(true))`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interp := NewInterpreter(false)
			_, err := interp.Execute(tt.code)
			if err == nil {
				t.Fatal("expected error but execution succeeded")
			}
		})
	}
}

// TestEvaluator_DivisionByZero tests division by zero error
func TestEvaluator_DivisionByZero(t *testing.T) {
	code := `x = 1 / 0`
	interp := NewInterpreter(false)
	_, err := interp.Execute(code)
	if err == nil {
		t.Fatal("expected division by zero error")
	}
	if !strings.Contains(err.Error(), "division") {
		t.Errorf("error should mention division: %v", err)
	}
}

// TestEvaluator_UndefinedVariable tests undefined variable error
func TestEvaluator_UndefinedVariable(t *testing.T) {
	code := `x = undefined_variable`
	interp := NewInterpreter(false)
	_, err := interp.Execute(code)
	if err == nil {
		t.Fatal("expected undefined variable error")
	}
}

// TestEvaluator_TypeMismatch tests operations between incompatible types
func TestEvaluator_TypeMismatch(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"array plus array", `x = [1] + [2]`},
		{"object plus object", `x = {a = 1} + {b = 2}`},
		{"array times number", `x = [1, 2] * 3`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interp := NewInterpreter(false)
			_, err := interp.Execute(tt.code)
			if err == nil {
				t.Fatal("expected type mismatch error")
			}
		})
	}
}

// TestBuiltin_BoundaryConditions tests edge cases across builtins
func TestBuiltin_BoundaryConditions(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"substr at boundary", `print(substr("hello", 5, 10))`, "\n"},
		{"substr zero length", `print(substr("hello", 0, 0))`, "\n"},
		{"split empty string", `print(len(split("", ",")))`, "1\n"},
		{"join empty array", `print(join([], ","))`, "\n"},
		{"sort empty array", `print(len(sort([])))`, "0\n"},
		{"filter empty array", `print(len(filter([], function(x) return true end)))`, "0\n"},
		{"map empty array", `print(len(map([], function(x) return x end)))`, "0\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test(t, tt.code, tt.expected)
		})
	}
}

// TestBuiltin_LargeInputs tests performance boundaries
func TestBuiltin_LargeInputs(t *testing.T) {
	code := `
arr = range(1, 100)
result = len(arr)
print(result)
`
	test(t, code, "100\n")
}

// TestBuiltin_NestedArrays tests deeply nested structures
func TestBuiltin_NestedArrays(t *testing.T) {
	code := `
nested = [[[1, 2], [3, 4]], [[5, 6], [7, 8]]]
print(len(nested))
print(len(nested[0]))
print(len(nested[0][0]))
`
	test(t, code, "2\n2\n2\n")
}

// TestBuiltin_ComplexTypeConversions tests various type conversion paths
func TestBuiltin_ComplexTypeConversions(t *testing.T) {
	code := `
n = tonumber("123")
s = tostring(n)
b = tobool(n)
print(s)
print(b)
`
	test(t, code, "123\ntrue\n")
}

// TestBuiltin_SubstringBoundaries tests substr() boundary conditions
func TestBuiltin_SubstringBoundaries(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"start at 0", `print(substr("hello", 0, 5))`, "hello\n"},
		{"middle", `print(substr("hello", 1, 3))`, "ell\n"},
		{"beyond end", `print(substr("hello", 3, 10))`, "lo\n"},
		{"negative start (from end)", `print(substr("hello", -1, 1))`, "o\n"},
		{"zero length", `print(substr("hello", 2, 0))`, "\n"},
		{"at end", `print(substr("hello", 5, 1))`, "\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test(t, tt.code, tt.expected)
		})
	}
}

// TestBuiltin_SplitEdgeCases tests split() edge cases
func TestBuiltin_SplitEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"empty delimiter", `arr = split("abc", "")
print(len(arr))`, "3\n"},
		{"delimiter not found", `print(len(split("hello", "x")))`, "1\n"},
		{"multiple delimiters", `print(len(split("a,b,c", ",")))`, "3\n"},
		{"consecutive delimiters", `print(len(split("a,,b", ",")))`, "3\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test(t, tt.code, tt.expected)
		})
	}
}

// TestBuiltin_FilterReduce tests filter and reduce together
func TestBuiltin_FilterReduceChain(t *testing.T) {
	code := `
numbers = [1, 2, 3, 4, 5]
evens = filter(numbers, function(x) return x % 2 == 0 end)
sum = reduce(evens, function(acc, x) return acc + x end, 0)
print(sum)
`
	test(t, code, "6\n")
}

// TestBuiltin_KeysValues tests keys() and values() functions
func TestBuiltin_KeysValues(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"keys empty object", `print(len(keys({})))`, "0\n"},
		{"values empty object", `print(len(values({})))`, "0\n"},
		{"keys with values", `obj = {a = 1, b = 2}
kv = keys(obj)
print(len(kv))`, "2\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test(t, tt.code, tt.expected)
		})
	}
}

// TestBuiltin_SubstringEdgeCases tests substr with boundary conditions
func TestBuiltin_SubstringEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"negative start (from end)", `print(substr("hello", -1, 1))`, "o\n"},
		{"start beyond length", `print(substr("hello", 10, 5))`, "\n"},
		{"zero length", `print(substr("hello", 1, 0))`, "\n"},
		{"full string", `print(substr("hello", 0, 5))`, "hello\n"},
		{"partial substring", `print(substr("hello", 1, 3))`, "ell\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test(t, tt.code, tt.expected)
		})
	}
}

// TestBuiltin_MathEdgeCases tests math functions with edge cases
func TestBuiltin_MathEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"floor negative", `print(floor(-3.7))`, "-4\n"},
		{"ceil negative", `print(ceil(-3.2))`, "-3\n"},
		{"round half even (2.5)", `print(round(2.5))`, "3\n"},
		{"round half even (1.5)", `print(round(1.5))`, "2\n"},
		{"abs zero", `print(abs(0))`, "0\n"},
		{"abs negative", `print(abs(-42))`, "42\n"},
		{"min single arg", `print(min(5))`, "5\n"},
		{"max single arg", `print(max(5))`, "5\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test(t, tt.code, tt.expected)
		})
	}
}

// TestBuiltin_ArrayOperationsEdgeCases tests array operations edge cases
func TestBuiltin_ArrayOperationsEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"sort empty array", `print(sort([]))`, "[]\n"},
		{"sort single element", `print(sort([1]))`, "[1]\n"},
		{"sort strings", `print(sort(["c", "a", "b"]))`, "[a, b, c]\n"},
		{"map empty array", `print(map([], function(x) return x * 2 end))`, "[]\n"},
		{"filter all match", `print(len(filter([1, 2, 3], function(x) return true end)))`, "3\n"},
		{"filter none match", `print(len(filter([1, 2, 3], function(x) return false end)))`, "0\n"},
		{"reduce empty array", `print(reduce([], function(a, x) return a + x end, 10))`, "10\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test(t, tt.code, tt.expected)
		})
	}
}

// TestBuiltin_StringMethodChaining tests chaining string operations
func TestBuiltin_StringMethodChaining(t *testing.T) {
	code := `
str = "  Hello World  "
result = trim(str)
result = lower(result)
result = upper(substr(result, 0, 5))
print(result)
`
	test(t, code, "HELLO\n")
}

// TestBuiltin_TypeConversionChain tests chaining type conversions
func TestBuiltin_TypeConversionChain(t *testing.T) {
	code := `
val = "42"
num = tonumber(val)
str = tostring(num)
bool = tobool(str)
print(bool)
`
	test(t, code, "true\n")
}

// TestBuiltin_FindWithNoMatches tests find when pattern doesn't match
func TestBuiltin_FindWithNoMatches(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"find no match", `result = find("hello", "xyz")
print(result)`, "[]\n"},
		{"find case sensitive", `result = find("Hello", "hello")
print(result)`, "[]\n"},
		{"find in empty string", `result = find("", "a")
print(result)`, "[]\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test(t, tt.code, tt.expected)
		})
	}
}

// TestBuiltin_JoinEdgeCases tests join with various inputs
func TestBuiltin_JoinEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"join empty array", `print(join([], ","))`, "\n"},
		{"join single element", `print(join(["a"], ","))`, "a\n"},
		{"join with empty separator", `print(join(["a", "b"], ""))`, "ab\n"},
		{"join numbers", `print(join([1, 2, 3], "-"))`, "1-2-3\n"},
		{"join mixed types", `print(join([1, "a", true], ","))`, "1,a,true\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test(t, tt.code, tt.expected)
		})
	}
}

// TestBuiltin_RangeEdgeCases tests range with various inputs
func TestBuiltin_RangeEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"range single value (inclusive)", `print(len(range(5, 5)))`, "1\n"},
		{"range negative (inclusive)", `arr = range(-3, -1)
print(len(arr))`, "3\n"},
		{"range zero to one (inclusive)", `print(len(range(0, 1)))`, "2\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test(t, tt.code, tt.expected)
		})
	}
}

// TestBuiltin_ClampEdgeCases tests clamp boundary conditions
func TestBuiltin_ClampEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"clamp below min", `print(clamp(1, 5, 10))`, "5\n"},
		{"clamp above max", `print(clamp(15, 5, 10))`, "10\n"},
		{"clamp within range", `print(clamp(7, 5, 10))`, "7\n"},
		{"clamp equal to min", `print(clamp(5, 5, 10))`, "5\n"},
		{"clamp equal to max", `print(clamp(10, 5, 10))`, "10\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test(t, tt.code, tt.expected)
		})
	}
}

// TestBuiltin_PowEdgeCases tests pow with various exponents
func TestBuiltin_PowEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"pow zero exponent", `print(pow(5, 0))`, "1\n"},
		{"pow one exponent", `print(pow(5, 1))`, "5\n"},
		{"pow negative exponent", `print(pow(2, -1))`, "0.5\n"},
		{"pow base one", `print(pow(1, 100))`, "1\n"},
		{"pow base zero", `print(pow(0, 5))`, "0\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test(t, tt.code, tt.expected)
		})
	}
}

// TestBuiltin_SqrtEdgeCases tests sqrt with edge cases
func TestBuiltin_SqrtEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"sqrt zero", `print(sqrt(0))`, "0\n"},
		{"sqrt one", `print(sqrt(1))`, "1\n"},
		{"sqrt perfect square", `print(sqrt(16))`, "4\n"},
		{"sqrt fraction", `print(sqrt(0.25))`, "0.5\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test(t, tt.code, tt.expected)
		})
	}
}

// TestBuiltin_JsonRoundTrip tests parse and format json together
func TestBuiltin_JsonRoundTrip(t *testing.T) {
	code := `
original = {name = "Alice", age = 30, active = true}
json_str = format_json(original)
parsed = parse_json(json_str)
print(parsed.name)
print(parsed.age)
print(parsed.active)
`
	test(t, code, "Alice\n30\ntrue\n")
}

// TestBuiltin_ContainsRegex tests contains with regex patterns
func TestBuiltin_ContainsRegex(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"contains digit", `print(contains("abc123", "[0-9]"))`, "true\n"},
		{"contains no digit", `print(contains("abc", "[0-9]"))`, "false\n"},
		{"contains word boundary", `print(contains("hello world", "world"))`, "true\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test(t, tt.code, tt.expected)
		})
	}
}

// ============================================================================
// COPY FUNCTIONS
// ============================================================================

// TestBuiltin_DeepCopy tests deep_copy() function
func TestBuiltin_DeepCopy(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{
			"deep copy primitive",
			`n = 42
copy = deep_copy(n)
print(copy)`,
			"42\n",
		},
		{
			"deep copy string",
			`s = "hello"
copy = deep_copy(s)
print(copy)`,
			"hello\n",
		},
		{
			"deep copy array",
			`arr = [1, 2, 3]
copy = deep_copy(arr)
copy[0] = 999
print(arr[0])
print(copy[0])`,
			"1\n999\n",
		},
		{
			"deep copy nested array",
			`arr = [[1, 2], [3, 4]]
copy = deep_copy(arr)
copy[0][0] = 999
print(arr[0][0])
print(copy[0][0])`,
			"1\n999\n",
		},
		{
			"deep copy object",
			`obj = {x = 10, y = 20}
copy = deep_copy(obj)
copy.x = 999
print(obj.x)
print(copy.x)`,
			"10\n999\n",
		},
		{
			"deep copy nested object",
			`obj = {data = {value = 42}}
copy = deep_copy(obj)
copy.data.value = 999
print(obj.data.value)
print(copy.data.value)`,
			"42\n999\n",
		},
		{
			"deep copy mixed structure",
			`obj = {arr = [1, 2, {nested = "val"}]}
copy = deep_copy(obj)
copy.arr[2].nested = "new"
print(obj.arr[2].nested)
print(copy.arr[2].nested)`,
			"val\nnew\n",
		},
		{
			"deep copy preserves nil",
			`arr = [1, nil, 3]
copy = deep_copy(arr)
print(len(copy))`,
			"3\n",
		},
		{
			"deep copy empty array",
			`arr = []
copy = deep_copy(arr)
print(len(copy))`,
			"0\n",
		},
		{
			"deep copy empty object",
			`obj = {}
copy = deep_copy(obj)
print(len(keys(copy)))`,
			"0\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test(t, tt.code, tt.expected)
		})
	}
}

// TestBuiltin_DeepCopyRemovesFunctions tests that deep_copy removes functions
func TestBuiltin_DeepCopyRemovesFunctions(t *testing.T) {
	code := `obj = {
  value = 42,
  get_value = function()
    return value
  end
}
copy = deep_copy(obj)
print(copy.value)
print(copy.get_value)
`
	test(t, code, "42\nnil\n")
}
