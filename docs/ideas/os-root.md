# Idea: os.Root — kernel-enforced file sandboxing

Go 1.24 added `os.Root` (rounded out in 1.25 with ReadFile/WriteFile/MkdirAll/RemoveAll/
Rename/etc.): open a directory, and every operation through that handle is confined to it
*by the kernel* (openat + RESOLVE_BENEATH-style resolution). `..` traversal, absolute
paths, and — critically — **symlink escapes** are all blocked at the syscall level, not by
string-checking paths. TOCTOU-safe, no cleverness required on our side.

## Why this is big for duso

The pitch is "single binary running your web stack on a small VM." Scripts are often
AI-generated, sometimes user-uploaded, and run alongside an HTTP server. Today a script
(or a path built from request input) can `readfile("/etc/passwd")` or write anywhere the
process user can. With os.Root, duso can promise: **scripts touch nothing outside their
app directory** — enforced by the OS, zero runtime cost on the happy path. That's a
differentiator none of the usual script-stack competitors offer out of the box.

## Where it plugs in

All disk I/O already funnels through two seams:

- `ResolvePath` (pkg/cli/file_io_util.go:49) — maps /HERE/, /CWD/, bare paths, absolute
- The injected capabilities on the interpreter: `ScriptLoader`, `FileReader`, and the
  write-side equivalents (pkg/script/script.go:79)

So the change is contained: open an `os.Root` on the app dir at startup, and have the
capability functions call `root.ReadFile`/`root.WriteFile`/... instead of the `os.*`
package functions. Script-visible behavior of well-behaved scripts is unchanged.

## Sketch

- Root = the entry script's appDir (already frozen at startup for bare-path resolution).
- `/EMBED/` and `/STORE/` are virtual filesystems — unaffected, already sandboxed by
  construction.
- `/HERE/` stays inside the root by definition (script dirs live under appDir); if a
  required module outside appDir needs /HERE/, that's the interesting edge — possibly one
  root per allowed tree.
- `/CWD/` and absolute paths are the policy question: under sandbox they'd fail unless
  under the root. That's the point.
- CLI flags: `-root DIR` (override), `-no-sandbox` (opt out). Open question whether
  sandbox is opt-in for 1.x (compat) and default in 2.0, or default now for server mode
  (`listen()` present) and opt-in for shell-script mode where touching /etc and /var IS
  the job. The server/script split probably decides it.
- `exec()` of unix tools is a separate trust boundary — os.Root doesn't confine child
  processes. Document clearly; possibly pair with a future exec allowlist.

## Costs

- Stdlib only, zero new deps, negligible binary size (aligned with the lean ethos).
- Platform: works on macOS/Linux/Windows; on platforms without openat-style support Go
  falls back to careful userspace resolution — still safe, slightly slower.
- Needs a test matrix for the path prefixes × sandbox on/off.
