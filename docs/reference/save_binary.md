# save_binary

Write binary data to a file.

## Syntax

```duso
save_binary(binary, filename)
```

## Parameters

- `binary` - A binary value to write
- `filename` - Path where the file should be saved

## Returns

Nothing (void). Throws an error if the file cannot be written.

## Description

Writes binary data to a file. This is useful for saving images, archives, and other binary content processed or downloaded by your script.

## Path Resolution

- **Absolute paths** - Used as-is
- **`/STORE/` paths** - Used as-is (persistent storage)
- **Relative paths** - Written to the script directory

## Examples

### Copy a binary file

```duso
image = load_binary("original.png")
save_binary(image, "copy.png")
```

### Save to persistent storage

```duso
data = load_binary("source.bin")
save_binary(data, "/STORE/backups/backup.bin")
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

### Save file from HTTP upload

```duso
ctx = context()
req = ctx.request()

// Access uploaded file
file = req.files["avatar"]  // binary value with metadata
if file then
  save_binary(file, "./uploads/" + file["filename"])
end
```

## Processing in Workers

Binary values are pointer-based and memory-efficient when passed to workers:

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
// ... process without copying the data
```

## See Also

- [binary - Binary data type overview](/docs/reference/binary.md)
- [load_binary() - Load binary files](/docs/reference/load_binary.md)
- [encode_base64() - Encode binary to base64 text](/docs/reference/encode_base64.md)
- [decode_base64() - Decode base64 text to binary](/docs/reference/decode_base64.md)
- [len() - Get size in bytes](/docs/reference/len.md)
- [http_server() - HTTP file uploads](/docs/reference/http_server.md)
