# binary

Immutable binary data type for handling files, images, and other binary content.

## Overview

A `binary` value represents immutable binary data (raw bytes) loaded from files or created within scripts. Binary values are memory-efficient - they use pointer sharing when passed between script scopes, so multiple workers can hold references to the same binary data without copying.

## Creating Binary Values

### load_binary(filename)

Load a file as binary data:

```duso
image = load_binary("avatar.png")
data = load_binary("/STORE/backup.bin")
```

Path resolution follows the same rules as `load()`:
- Absolute paths and `/STORE/`, `/EMBED/` → used as-is
- Relative paths → tries script dir, `/STORE/`, `/EMBED/` in order

## Operations

### len(binary)

Get the size in bytes:

```duso
image = load_binary("photo.png")
print(len(image))  // prints byte count
```

### Metadata Access

Access metadata fields using bracket notation:

```duso
image = load_binary("photo.png")
filename = image["filename"]  // original filename
```

**Available metadata:**
- `filename` - original filename from `load_binary()`
- Other metadata may be populated by HTTP file uploads (content_type, etc.)

## Saving Binary Values

### save_binary(binary, filename)

Write binary data to a file:

```duso
image = load_binary("original.png")
save_binary(image, "copy.png")
save_binary(image, "/STORE/backups/image.png")
```

Path resolution:
- Absolute paths and `/STORE/` → used as-is
- Relative paths → written to script directory

## Type Checking

Use `type()` to check if a value is binary:

```duso
data = load_binary("file.bin")
if type(data) == "binary" then
  print("It's binary!")
end
```

## Truthiness

Binary values are truthy (non-empty binaries evaluate to true):

```duso
image = load_binary("photo.png")
if image then
  print("Has data")
end
```

## Properties

- **Immutable** - Binary data cannot be modified after creation
- **No indexing** - Cannot access individual bytes like arrays
- **Pointer-based** - When passed to other scripts, only the pointer is copied, not the data
- **Memory efficient** - Multiple workers/threads can reference the same binary without overhead
- **GC managed** - Automatically cleaned up when last reference is dropped

## HTTP File Uploads

In HTTP handlers, uploaded files are automatically converted to binary values with metadata:

```duso
ctx = context()
req = ctx.request()

// Access uploaded file
file = req.files["avatar"]  // binary value with metadata
print(file["filename"])      // original filename
print(len(file))             // file size
save_binary(file, "./uploads/" + file["filename"])
```

## Examples

### Copy a file

```duso
data = load_binary("source.dat")
save_binary(data, "backup.dat")
```

### Process binary in workers

```duso
// Main script
binary_data = load_binary("large-file.bin")

// Spawn workers - each gets efficient pointer to same data
for i = 1, 100 do
  spawn("process_worker.du", {data = binary_data})
end
```

```duso
// process_worker.du
ctx = context()
data = ctx.data
print("Processing", len(data), "bytes")
// ... process without copying
```

### Validate file before saving

```duso
uploaded = load_binary("temp_upload.bin")

if len(uploaded) > 10000000 then
  print("File too large")
else
  save_binary(uploaded, "uploads/file.bin")
end
```

## Limitations

- **No partial reads** - Must load entire file into memory
- **No modification** - Binary values are immutable (cannot edit bytes)
- **No direct inspection** - Cannot index individual bytes; use metadata or file operations instead
- **Text encoding** - For text files, use `load()` instead for UTF-8 string handling

## See Also

- [load() / save() - Text file operations](/docs/reference/load.md)
- [type() - Type checking](/docs/reference/type.md)
- [len() - Length operations](/docs/reference/len.md)
- [http_server() - HTTP file uploads](/docs/reference/http_server.md)
