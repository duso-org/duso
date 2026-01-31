# template()

Create a reusable template function from a string with `{{expression}}` syntax.

## Signature

```duso
template(string) → function
```

## Parameters

- `string` (string) - A string containing `{{expression}}` placeholders

## Returns

A function that evaluates the template with provided named arguments.

## Description

`template()` takes a string containing template expressions and returns a function. When you call this function with named arguments, the template is evaluated with those values. This allows:

- **Deferred evaluation** - Parse template once, use many times with different values
- **Template reuse** - Define a template pattern and apply it to multiple datasets
- **Cleaner code** - Avoid repetitive string building with concatenation
- **Template logic** - Use expressions and computations inside the template

## Examples

### Basic usage

```duso
// Create a template function
greeting = template("Hello {{name}}, you are {{age}} years old!")

// Evaluate with different values
msg1 = greeting(name = "Alice", age = 30)
msg2 = greeting(name = "Bob", age = 25)

print(msg1)  // Output: Hello Alice, you are 30 years old!
print(msg2)  // Output: Hello Bob, you are 25 years old!
```

### Complex expressions

```duso
// Templates can include any Duso expressions
calc = template("{{x}} + {{y}} = {{x + y}}")
result = calc(x = 5, y = 3)
print(result)  // Output: 5 + 3 = 8

// With object access
emailTemplate = template("Contact: {{user.email}} ({{user.name}})")
person = {name = "Alice", email = "alice@example.com"}
email = emailTemplate(user = person)
print(email)  // Output: Contact: alice@example.com (Alice)
```

### Email template example

```duso
emailTemplate = template("""
Dear {{user.name}},

Your order for {{product}} has been confirmed.
Order ID: {{order_id}}
Total: ${{amount}}

Thank you for your business!

Best regards,
The Team
""")

// Use with multiple customers
order1 = emailTemplate(
  user = {name = "Alice"},
  product = "Widget",
  order_id = "ORD-001",
  amount = 50.00
)

order2 = emailTemplate(
  user = {name = "Bob"},
  product = "Gadget",
  order_id = "ORD-002",
  amount = 75.50
)
```

### Dynamic SQL query builder

```duso
// Template for flexible SQL queries
queryBuilder = template("""
SELECT {{fields}}
FROM {{table}}
WHERE {{where_clause}}
LIMIT {{limit}}
""")

// Build different queries
users_query = queryBuilder(
  fields = "id, name, email",
  table = "users",
  where_clause = "age > 18",
  limit = 100
)

products_query = queryBuilder(
  fields = "id, name, price",
  table = "products",
  where_clause = "stock > 0",
  limit = 50
)
```

### Using with arrays and functions

```duso
// Template that formats a list
itemTemplate = template("  - {{item}}: {{value}}")

items = ["apple", "banana", "orange"]
prices = [1.50, 0.75, 2.00]

// Map over items
formatted = map(items, function(item, i)
  return itemTemplate(item = item, value = prices[i])
end)

for line in formatted do
  print(line)
end
// Output:
//   - apple: 1.5
//   - banana: 0.75
//   - orange: 2
```

### Stored templates with raw strings

```duso
// Use raw to store template without evaluating
template_text = raw """
Invoice for {{customer}}
Date: {{date}}
Items: {{item_count}}
Total: ${{total}}
"""

// Create template function from raw string
invoice = template(template_text)

// Use multiple times
inv1 = invoice(customer = "Alice", date = "2025-01-30", item_count = 3, total = 99.99)
inv2 = invoice(customer = "Bob", date = "2025-01-30", item_count = 1, total = 49.99)
```

## Undefined Variables

If the template references variables not provided to the function, those expressions remain literal:

```duso
t = template("User: {{user}}, Role: {{role}}")

// Missing 'role' argument
result = t(user = "Alice")
// Output: User: Alice, Role: {{role}}
```

## Performance Note

- Template compilation happens once when `template()` is called
- Subsequent evaluations are fast - the template structure is reused
- Good for scenarios where you use the same template many times

## Edge Cases and Common Patterns

### Conditional logic in templates

While templates don't support if/then syntax directly, you can pre-compute values:

```duso
t = template("Status: {{status_text}}")
status_text = is_active ? "Active" : "Inactive"
result = t(status_text = status_text)
```

### Nested templates

You can create template generators:

```duso
// Template that generates another template
templateFactory = template("Template pattern: {{pattern}}")
pattern = "Hello {{name}}, your ID is {{id}}"
description = templateFactory(pattern = pattern)
print(description)
```

## See Also

- [raw - Prevent template evaluation](/docs/reference/raw.md)
- [String templates - Template expression syntax](/docs/learning-duso.md#templates)
- [Strings reference - String operations](/docs/reference/string.md)
