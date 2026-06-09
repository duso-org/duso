# Virtual Filesystems: /EMBED/ and /STORE/

Duso provides two virtual filesystem prefixes for working with files without direct filesystem access.

## Overview

| Prefix | Purpose | Access | Persistence |
|--------|---------|--------|-------------|
| `/EMBED/` | Read embedded resources (stdlib, docs, etc.) | Read-only | N/A (built into binary) |
| `/STORE/` | Create and manage scripts at runtime | Read/write | In-memory (optional persistence) |

## /EMBED/ - Embedded Files

The `/EMBED/` prefix provides read-only access to files embedded in the Duso binary during compilation:

- **stdlib/** - Standard library modules
- **docs/** - Documentation files
- **contrib/** - Contributed modules
- **examples/** - Example scripts
- **README.md** - Project README

Use `load()` to read embedded files: `readme = load("/EMBED/README.md")`

When using `require()` or `include()`, the system automatically searches embedded files under `/EMBED/stdlib/`.

## /STORE/ - Virtual Filesystem

The `/STORE/` prefix provides a virtual filesystem backed by the in-memory datastore. You can create, modify, and execute code at runtime without touching the real filesystem.

Standard file operations work with `/STORE/`:
- `save(path, content)` - Write a file
- `load(path)` - Read a file  
- `append_file(path, content)` - Append to a file
- `copy_file(src, dst)` - Copy a file
- `move_file(src, dst)` - Move/rename a file
- `remove_file(path)` - Delete a file
- `list_files(pattern)` - List files matching a pattern
- `file_exists(path)` - Check if a file exists

All paths support wildcard patterns: `*` matches any characters, `?` matches a single character.

### Modules from /STORE/

You can create reusable modules in `/STORE/` and load them with `require()`:

```duso
save("/STORE/helpers.du", """
  return {
    double = function(x) return x * 2 end,
    square = function(x) return x * x end
  }
""")

helpers = require("/STORE/helpers")
print(helpers.double(5))
```

### Datastore Backing

`/STORE/` is backed by the datastore "vfs" namespace. For persistence, configure the datastore:

```duso
store = datastore("vfs", {
  persist = "data.json",
  persist_interval = 60
})
```

## Module Search Order

When using `require()` or `include()`, Duso searches in this order:

1. Local filesystem (relative to script directory)
2. `/STORE/` virtual filesystem
3. `/EMBED/stdlib/` embedded stdlib
4. `/EMBED/contrib/` embedded contrib modules

## Security: The -no-files Flag

Restrict filesystem access to virtual filesystems only:

```bash
duso -no-files agent-script.du
```

With `-no-files`:
- ✅ Read from `/EMBED/`
- ✅ Read/write `/STORE/`
- ❌ Access to real filesystem
