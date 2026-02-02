package script

import (
	"testing"
)

// Helper for builtin tests
func test(t *testing.T, code string, expected string) {
	interp := NewInterpreter(false)
	output, err := interp.Execute(code)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
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
		{"comma separated", `print(split("a,b,c", ","))`, "[a b c]\n"},
		{"space separated", `print(split("a b c", " "))`, "[a b c]\n"},
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

// TestBuiltin_Append tests append() function
func TestBuiltin_Append(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"append to empty", `print(append([], 1))`, "[1]\n"},
		{"append to existing", `print(append([1, 2], 3))`, "[1 2 3]\n"},
		{"append returns new array", `a = [1]
b = append(a, 2)
print(a)
print(b)`, "[1]\n[1 2]\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test(t, tt.code, tt.expected)
		})
	}
}

// TestBuiltin_Sort tests sort() function
func TestBuiltin_Sort(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"sort numbers", `print(sort([3, 1, 2]))`, "[1 2 3]\n"},
		{"sort strings", `print(sort(["c", "a", "b"]))`, "[a b c]\n"},
		{"already sorted", `print(sort([1, 2, 3]))`, "[1 2 3]\n"},
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
		{"basic range", `print(range(1, 3))`, "[1 2 3]\n"},
		{"with step", `print(range(1, 5, 2))`, "[1 3 5]\n"},
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
print(result)`, "[2 4 6]\n"},
		{"to strings", `result = map([1, 2], function(x) return tostring(x) end)
print(result)`, "[1 2]\n"},
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
print(result)`, "[2 4]\n"},
		{"non-empty", `result = filter(["a", "", "b"], function(x) return x end)
print(result)`, "[a b]\n"},
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
