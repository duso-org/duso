package script

import (
	"testing"
)

// Integration tests combine multiple language features in realistic scenarios

// TestIntegration_DataProcessingPipeline tests complete data transformation
func TestIntegration_DataProcessingPipeline(t *testing.T) {
	code := `
// Parse data
json_data = """[
  {"name": "Alice", "score": 85},
  {"name": "Bob", "score": 92},
  {"name": "Charlie", "score": 78}
]"""

students = parse_json(json_data)

// Filter high scorers
high_scorers = filter(students, function(s)
  return s.score >= 85
end)

// Extract names
names = map(high_scorers, function(s)
  return s.name
end)

// Sort and output
result = join(sort(names), ", ")
print(result)
`
	test(t, code, "Alice, Bob\n")
}

// TestIntegration_StatisticsCalculation tests building statistics
func TestIntegration_StatisticsCalculation(t *testing.T) {
	code := `
scores = [85, 92, 78, 95, 88]

sum = reduce(scores, function(acc, x) return acc + x end, 0)
count = len(scores)
average = sum / count

min_score = reduce(scores, function(acc, x)
  return x < acc ? x : acc
end, scores[0])

max_score = reduce(scores, function(acc, x)
  return x > acc ? x : acc
end, scores[0])

print("Sum: " + sum)
print("Average: " + average)
print("Min: " + min_score)
print("Max: " + max_score)
`
	expected := "Sum: 438\nAverage: 87.6\nMin: 78\nMax: 95\n"
	test(t, code, expected)
}

// TestIntegration_RecursiveDataStructure tests recursion with complex data
func TestIntegration_RecursiveDataStructure(t *testing.T) {
	code := `
function count_nodes(obj)
  if type(obj) == "array" then
    sum = 0
    for item in obj do
      sum = sum + count_nodes(item)
    end
    return sum
  elseif type(obj) == "object" then
    sum = 0
    for key in obj do
      sum = sum + count_nodes(obj[key])
    end
    return sum
  else
    return 1
  end
end

tree = {
  value = 1,
  children = [
    {value = 2, children = []},
    {value = 3, children = [
      {value = 4, children = []}
    ]}
  ]
}

print(count_nodes(tree))
`
	test(t, code, "4\n")
}

// TestIntegration_TextProcessing tests string manipulation pipeline
func TestIntegration_TextProcessing(t *testing.T) {
	code := `
text = "Hello World Duso"

// Convert to lowercase and split
words = split(lower(text), " ")

// Process each word
processed = map(words, function(word)
  return upper(substr(word, 0, 1)) + substr(word, 1)
end)

// Join back
result = join(processed, " ")
print(result)
`
	test(t, code, "Hello World Duso\n")
}

// TestIntegration_NestedFunctionCalls tests deep function composition
func TestIntegration_NestedFunctionCalls(t *testing.T) {
	code := `
function is_even(n)
  return n % 2 == 0
end

function double(n)
  return n * 2
end

function triple(n)
  return n * 3
end

numbers = [1, 2, 3, 4, 5]

// Chain: filter evens, double them, triple each, sum
result = reduce(
  map(
    filter(numbers, is_even),
    function(n) return double(n) end
  ),
  function(acc, n) return acc + triple(n) end,
  0
)

print(result)
`
	test(t, code, "36\n")
}

// TestIntegration_ObjectOrientedPattern tests object methods and state
func TestIntegration_ObjectOrientedPattern(t *testing.T) {
	code := `
Account = {
  balance = 0,
  deposit = function(amount)
    balance = balance + amount
    return "Deposited " + amount
  end,
  withdraw = function(amount)
    if amount > balance then
      return "Insufficient funds"
    end
    balance = balance - amount
    return "Withdrew " + amount
  end,
  get_balance = function()
    return balance
  end
}

// Create account instance
account = Account()
print(account.deposit(100))
print(account.deposit(50))
print(account.withdraw(30))
print("Balance: " + account.get_balance())
`
	expected := "Deposited 100\nDeposited 50\nWithdrew 30\nBalance: 120\n"
	test(t, code, expected)
}

// TestIntegration_DynamicDataStructure tests building structures dynamically
func TestIntegration_DynamicDataStructure(t *testing.T) {
	code := `
function build_map(keys_arr, values_arr)
  result = {}
  for i = 0, len(keys_arr) - 1 do
    key = keys_arr[i]
    value = values_arr[i]
    result[key] = value
  end
  return result
end

keys = ["a", "b", "c"]
values = [1, 2, 3]
map_obj = build_map(keys, values)

print(map_obj.a)
print(map_obj.b)
print(map_obj.c)
`
	test(t, code, "1\n2\n3\n")
}

// TestIntegration_ErrorHandlingInPipeline tests error recovery
func TestIntegration_ErrorHandlingInPipeline(t *testing.T) {
	code := `
try
  data = parse_json("""{"name":"Alice"}""")
  print(data.name)
catch (e)
  print("Error: " + e)
end

print("Done")
`
	test(t, code, "Alice\nDone\n")
}

// TestIntegration_TemplateStringWithLogic tests template rendering
func TestIntegration_TemplateStringWithLogic(t *testing.T) {
	code := `
function render_person(person)
  status = person.age >= 18 ? "adult" : "minor"
  return """
Name: {{person.name}}
Age: {{person.age}}
Status: {{status}}
"""
end

person = {name = "Alice", age = 25}
print(render_person(person))
`
	test(t, code, "Name: Alice\nAge: 25\nStatus: adult\n")
}

// TestIntegration_SortingWithComparator tests custom sort comparator
func TestIntegration_SortingWithComparator(t *testing.T) {
	code := `
people = [
  {name = "Alice", age = 30},
  {name = "Bob", age = 25},
  {name = "Charlie", age = 35}
]

// Sort by age
sorted = sort(people, function(a, b)
  return a.age < b.age
end)

for person in sorted do
  print(person.name + " (" + person.age + ")")
end
`
	test(t, code, "Bob (25)\nAlice (30)\nCharlie (35)\n")
}

// TestIntegration_ComplexConditionalLogic tests nested conditions
func TestIntegration_ComplexConditionalLogic(t *testing.T) {
	code := `
function classify(score)
  if score >= 90 then
    return "A"
  elseif score >= 80 then
    return "B"
  elseif score >= 70 then
    return "C"
  elseif score >= 60 then
    return "D"
  else
    return "F"
  end
end

scores = [95, 85, 75, 65, 55]
for score in scores do
  print(score + " -> " + classify(score))
end
`
	test(t, code, "95 -> A\n85 -> B\n75 -> C\n65 -> D\n55 -> F\n")
}

// TestIntegration_MultiLevelNesting tests deeply nested structures
func TestIntegration_MultiLevelNesting(t *testing.T) {
	code := `
config = {
  database = {
    host = "localhost",
    port = 5432,
    credentials = {
      user = "admin",
      password = "secret"
    }
  },
  cache = {
    enabled = true,
    ttl = 3600
  }
}

print(config.database.credentials.user)
print(config.cache.ttl)
`
	test(t, code, "admin\n3600\n")
}

// TestIntegration_ArrayTransformations tests multiple array operations
func TestIntegration_ArrayTransformations(t *testing.T) {
	code := `
// Start with range, filter, transform, reduce
result = reduce(
  map(
    filter(
      range(1, 10),
      function(x) return x % 2 == 0 end
    ),
    function(x) return x * x end
  ),
  function(acc, x) return acc + x end,
  0
)

print("Sum of squares of even numbers 1-10: " + result)
`
	test(t, code, "Sum of squares of even numbers 1-10: 220\n")
}

// TestIntegration_FunctionFactories tests creating parameterized functions
func TestIntegration_FunctionFactories(t *testing.T) {
	code := `
function make_multiplier(factor)
  return function(x) return x * factor end
end

function make_adder(amount)
  return function(x) return x + amount end
end

multiply_by_3 = make_multiplier(3)
add_10 = make_adder(10)

x = 5
result = multiply_by_3(add_10(x))
print(result)
`
	test(t, code, "45\n")
}

// TestIntegration_StateManagement tests maintaining mutable state
func TestIntegration_StateManagement(t *testing.T) {
	code := `
var state = {
  users = [],
  add_user = function(name)
    users = append(users, {name = name, id = len(users) + 1})
  end,
  get_user = function(id)
    for user in users do
      if user.id == id then
        return user
      end
    end
    return nil
  end
}

state.add_user("Alice")
state.add_user("Bob")

user1 = state.get_user(1)
user2 = state.get_user(2)

print(user1.name)
print(user2.name)
`
	test(t, code, "Alice\nBob\n")
}

// TestIntegration_MixedTypeHandling tests working with multiple types
func TestIntegration_MixedTypeHandling(t *testing.T) {
	code := `
data = [
  {type = "string", value = "hello"},
  {type = "number", value = 42},
  {type = "boolean", value = true},
  {type = "array", value = [1, 2, 3]}
]

for item in data do
  print(item.type + ": " + tostring(item.value))
end
`
	test(t, code, "string: hello\nnumber: 42\nboolean: true\narray: [1 2 3]\n")
}

// TestIntegration_ConditionalFlowWithFunctions tests conditions affecting execution
func TestIntegration_ConditionalFlowWithFunctions(t *testing.T) {
	code := `
function process(items, predicate, transform)
  result = []
  for item in items do
    if predicate(item) then
      result = append(result, transform(item))
    end
  end
  return result
end

numbers = [1, 2, 3, 4, 5]
evens = process(
  numbers,
  function(x) return x % 2 == 0 end,
  function(x) return x * x end
)

print(format_json(evens))
`
	test(t, code, "[4,16]\n")
}
