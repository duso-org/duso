# File Functions

Duso provides built-in file and directory operations for reading, writing, and managing files on disk.

## Reading Files

- [load()](/docs/reference/load.md) - Read file as text string
- [load_binary()](/docs/reference/load_binary.md) - Read file as binary data

## Writing Files

- [save()](/docs/reference/save.md) - Write text string to file (create/overwrite)
- [save_binary()](/docs/reference/save_binary.md) - Write binary data to file
- [append_file()](/docs/reference/append_file.md) - Append text to file

## File Operations

- [copy_file()](/docs/reference/copy_file.md) - Copy file
- [move_file()](/docs/reference/move_file.md) - Move file to new location
- [rename_file()](/docs/reference/rename_file.md) - Rename file
- [remove_file()](/docs/reference/remove_file.md) - Delete file

## Directory Operations

- [list_dir()](/docs/reference/list_dir.md) - List directory contents
- [make_dir()](/docs/reference/make_dir.md) - Create directory (creates parents if needed)
- [remove_dir()](/docs/reference/remove_dir.md) - Remove empty directory
- [current_dir()](/docs/reference/current_dir.md) - Get current working directory

## Utilities

- [file_exists()](/docs/reference/file_exists.md) - Check if file or directory exists
- [file_type()](/docs/reference/file_type.md) - Get file type ("file" or "directory")
- [watch()](/docs/reference/watch.md) - Monitor file/directory for changes

## Quick Examples

### Reading and Writing Text

```duso
// Read file
content = load("README.md")
print(content)

// Write file
save("output.txt", "Hello World")

// Append to file
append_file("log.txt", "New log entry\n")
```

### Working with Binary Data

```duso
// Read image
image = load_binary("photo.jpg")

// Save modified image
save_binary(image, "output.jpg")
```

### Directory Operations

```duso
// List files
files = list_dir(".")
for f in files do
  print(f.name)
end

// Create directories
make_dir("output/images")

// Check if exists
if file_exists("data.txt") then
  print("File found")
end
```

### Monitoring Changes

```duso
// Watch for file changes (blocks until change or timeout)
changed = watch("config.json", 5)
if changed then
  print("Config updated")
end
```
