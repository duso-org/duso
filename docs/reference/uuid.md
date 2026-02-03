# uuid()

Generate a UUID v7 (RFC 9562) universally unique identifier.

## Signature

```duso
uuid()
```

## Parameters

None

## Returns

A UUID v7 string in the format `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`

## Details

UUID v7 is a time-ordered, sortable UUID that combines:

- **48-bit Unix timestamp (milliseconds)**: Ensures UUIDs generated at the same time are monotonically increasing, making them ideal for database primary keys and indexes
- **74 bits of random data**: Provides uniqueness within the same millisecond
- **Version/Variant bits**: Encoded to identify it as a v7 UUID

**Key advantages over UUID v4:**
- **Time-sortable**: Monotonically increasing, enabling efficient database B-tree indexing
- **Privacy**: Doesn't embed MAC addresses like UUID v1
- **Database performance**: Up to 35% better insertion performance in relational databases (PostgreSQL, MySQL, etc.) compared to random UUIDs
- **Distributed-friendly**: Safe to generate independently in distributed systems

## Examples

Generate a unique ID:

```duso
id = uuid()
print(id)  // "018f5c7a-d7f8-7e2c-81a3-f9c4d6e1b5a0"
```

Use as a database primary key:

```duso
record = {
  id = uuid(),
  name = "Alice",
  created = now()
}
```

Generate multiple UUIDs (each unique and time-ordered):

```duso
ids = []
for i = 1, 5 do
  ids = append(ids, uuid())
end
print(ids)
```

## See Also

- [now() - Get current timestamp](/docs/reference/now.md)
- [random() - Get random float](/docs/reference/random.md)
