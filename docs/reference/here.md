# here

Returns the directory of the file the call is written in.

## Syntax

```duso
here()
```

## Description

`here()` returns the absolute directory of the source file containing the call — the same directory the [`/HERE/`](/docs/files-and-modules.md#path-roots) prefix resolves to.

It is **lexical**: the answer depends on where the code was written, not on who called it. A function a module exports still reports the module's own directory when the caller invokes it from somewhere else. The parser folds `here()` to a string constant as the file is parsed, so there is no runtime cost and no call stack to get wrong.

Use `/HERE/foo.txt` for a literal path. Reach for `here()` when the path is assembled at runtime:

```duso
load(here() + "/locales/" + lang + ".json")
```

## Returns

`string` — absolute directory path, with no trailing slash.

For a module parsed out of the embedded filesystem, that path is its `/EMBED/` directory, so bundled builds keep working unchanged.

In code with no source file to fold against — the REPL, `duso eval`, `parse()` on a string — `here()` falls back to the working directory.

## Examples

Loading a resource that ships with a module:

```duso
// lib/mailer/mailer.du
function render(name)
  return load(here() + "/templates/" + name + ".html")
end

{ render = render }
```

```duso
// app.du, in a different directory
m = require("lib/mailer/mailer.du")
html = m.render("welcome")   // reads lib/mailer/templates/welcome.html
```

Launching a worker that sits beside its module:

```duso
spawn(here() + "/worker.du", {job = j})
```

## Notes

- `here()` is folded at parse time, so defining your own `here` function will not override it inside a script file.
- `here()` answers "what file am I in?". [`current_dir()`](/docs/reference/current_dir.md) answers "where was the process started?" — these differ whenever a script is run from another directory.

## See Also

- [Files, Modules, and Paths](/docs/files-and-modules.md) - the full path contract
- [current_dir()](/docs/reference/current_dir.md) - process working directory
- [load()](/docs/reference/load.md) - read a file
- [require()](/docs/reference/require.md) - load a module
