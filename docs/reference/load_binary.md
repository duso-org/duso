# load_binary

Load a file as immutable binary data.

## Syntax

```duso
load_binary(filename)
```

## Parameters

- `filename` - Path to the file to load

## Returns

A `binary` value containing the file's raw bytes, or throws an error if the file cannot be read.

## Description

Loads an entire file into memory as immutable binary data. This is useful for handling images, archives, and other binary files.

## Path Resolution

Path resolution follows the same rules as `load()`:
- **Absolute paths** - Used as-is
- **`/STORE/` paths** - Used as-is (persistent storage)
- **`/EMBED/` paths** - Used as-is (embedded files)
- **Relative paths** - Tried in this order:
  1. Script directory
  2. `/STORE/`
  3. `/EMBED/`

## Examples

### Load an image file

```duso
image = load_binary("avatar.png")
print(len(image))  // prints byte count
```

### Load from absolute path

```duso
data = load_binary("/STORE/backup.bin")
```

### Access file metadata

```duso
file = load_binary("document.pdf")
filename = file["filename"]  // original filename
print(filename)
```

**Available metadata:**
- `filename` - original filename from `load_binary()`
- `content_type` - may be populated by HTTP file uploads

## Properties

- **Immutable** - Binary data cannot be modified after creation
- **No indexing** - Cannot access individual bytes like arrays
- **Pointer-based** - When passed to other scripts, only the pointer is copied, not the data
- **Memory efficient** - Multiple workers/threads can reference the same binary without overhead
- **GC managed** - Automatically cleaned up when last reference is dropped

## Type Checking

```duso
data = load_binary("file.bin")
if type(data) == "binary" then
  print("It's binary!")
end
```

## See Also

- [binary - Binary data type overview](/docs/reference/binary.md)
- [save_binary() - Write binary data to files](/docs/reference/save_binary.md)
- [encode_base64() - Encode binary to base64 text](/docs/reference/encode_base64.md)
- [decode_base64() - Decode base64 text to binary](/docs/reference/decode_base64.md)
- [load() - Load text files](/docs/reference/load.md)
- [len() - Get size in bytes](/docs/reference/len.md)
