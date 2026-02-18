# Virtual Filesystem Path Issue

## Problem
On Windows, `filepath.Join()` converts forward slashes to backslashes. This breaks `/EMBED/` and `/STORE/` paths:
- `filepath.Join("/EMBED/stdlib", "markdown")` → `\EMBED\stdlib\markdown` ❌

The Go `embed.FS` and virtual filesystem always expect forward slashes.

## Solution
Use `path.Join()` (not `filepath.Join()`) for virtual paths:
- `path.Join("/EMBED/stdlib", "markdown")` → `/EMBED/stdlib/markdown` ✅

`path.Join()` always uses forward slashes and normalizes paths (removes `//`, etc).

## Files to Fix
- `pkg/cli/module_resolver.go` - use `path.Join()` for `/EMBED/` paths
- `pkg/cli/builtin_load.go` - use `path.Join()` for `/EMBED/` and `/STORE/` paths
- `pkg/cli/functions.go` - use `path.Join()` for `/EMBED/` paths
- Remove the `JoinVirtualPath()` and `NormalizeEmbeddedPath()` functions added during failed attempts
