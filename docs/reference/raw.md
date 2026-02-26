# raw

Prefix a string literal to prevent template expression evaluation.

## Syntax

```duso
raw "string literal"
raw 'string literal'
raw """multiline string"""
raw '''multiline string'''
```

## Description

By default, string literals with `{{expression}}` syntax are evaluated as templates. The `raw` keyword prevents this evaluation, keeping template expressions as literal text. The string is still unescaped (backslash sequences are processed normally).

Use `raw` when you need to:
- Pass a template string to a function without evaluating it
- Store template patterns for later evaluation
- Embed Duso code or other languages that use `{{}}` syntax

## Examples

### Default behavior - templates are evaluated

```duso
name = "Alice"
msg = "Hello {{name}}"
print(msg)  // Output: Hello Alice
```

### Using raw - templates are NOT evaluated

```duso
name = "Alice"
msg = raw "Hello {{name}}"
print(msg)  // Output: Hello {{name}}
```

### Storing template patterns

```duso
// Store a template pattern without evaluating it
template_pattern = raw "Dear {{customer}}, your balance is {{balance}}"
print(template_pattern)  // Output: Dear {{customer}}, your balance is {{balance}}

// Later, evaluate it with template()
greeting = template(template_pattern)
result = greeting(customer = "Bob", balance = "100")
print(result)  // Output: Dear Bob, your balance is 100
```

### Embedding other languages

```duso
// Pass a template string to an external system without Duso evaluating it
jinja_template = raw "{% for user in users %}\n  - {{user.name}}\n{% endfor %}"
print(jinja_template)  // Sends Jinja syntax to external template engine

// JSON with template markers
json_template = raw """
{
  "user": "{{username}}",
  "role": "{{role}}"
}
"""
```

### Multiline raw strings

```duso
sql_template = raw """
  SELECT * FROM users
  WHERE age > {{min_age}}
    AND status = '{{status}}'
"""
print(sql_template)
```

## Raw vs Regular Strings

| Feature | Regular String | Raw String |
|---------|---|---|
| Template evaluation | ✓ Evaluated | ✗ Not evaluated |
| Escape sequences | ✓ Processed | ✓ Processed |
| `{{}}` syntax | Becomes values | Stays literal |
| Use case | Dynamic content | Patterns, templates to store |

## Common Pattern: Deferred Evaluation

```duso
// Store a template without evaluating it
emailTemplate = raw """
Dear {{name}},

Thank you for your purchase of {{product}}.
Your order total is {{amount}}.

Best regards,
The Team
"""

// Later, evaluate with specific values
mailer = template(emailTemplate)
email1 = mailer(name = "Alice", product = "Widget", amount = "$50")
email2 = mailer(name = "Bob", product = "Gadget", amount = "$75")
```

## See Also

- [template() - Create reusable template functions](/docs/reference/template.md)
- [String templates - Template expression syntax](/docs/learning-duso.md#templates)
