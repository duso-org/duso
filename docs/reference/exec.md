# exec()

Run an external command and capture its output. Blocks until the command finishes. Available in `duso` CLI only.

`exec(command [, options])`

**exec() is disabled unless the operator starts duso with `-allow-exec`.** See **Permission** below — the flag is the whole of what exec() can reach, and a script cannot widen it.

## Parameters

- `command` (string) - Command to run. Must be permitted by `-allow-exec`.
- `options` (optional, object):
  - `args` (array) - Arguments, passed to the process exactly as given. Strings, numbers, and booleans only.
  - `env` (object) - Environment variables for the child
  - `inherit_env` (boolean, default: `false`) - Pass duso's entire environment to the child
  - `dir` (string) - Working directory
  - `timeout` (number) - Seconds to wait before giving up. No timeout by default, same as `fetch()`.
  - `max_output` (number, default: 1048576) - Bytes captured per stream
  - `stdin` (string) - Written to the child's standard input

## Returns

An object:

- `ok` (boolean) - The command exited 0
- `code` (number) - Exit status
- `stdout` (string) - Captured standard output
- `stderr` (string) - Captured standard error
- `truncated` (boolean) - Either stream hit `max_output`

## Examples

```duso
result = exec("systemctl", {args = ["restart", "arland"]})
if not result.ok then
  throw("restart failed: " + result.stderr)
end
```

Reading output:

```duso
result = exec("certbot", {args = ["certificates"]})
for line in split(result.stdout, "\n") do
  if contains(line, "Expiry Date") then print(trim(line)) end
end
```

Writing to the child's stdin:

```duso
result = exec("gzip", {args = ["-c"], stdin = document})
```

Giving up on a command that hangs:

```duso
try
  result = exec("rsync", {args = ["-a", src, dst], timeout = 300})
catch (e)
  print("rsync gave up: " + e)
end
```

## No shell is involved

`exec()` starts the program directly. Arguments reach the process as written — nothing expands variables, splits on spaces, or treats `;`, `|`, `&&`, or backticks as syntax:

```duso
exec("echo", {args = ["hi; whoami"]})    // prints: hi; whoami
exec("echo", {args = ["$HOME"]})          // prints: $HOME
```

This is what makes it safe to build an argument out of request data. The value becomes one argument, never a second command.

A shell is reachable only by asking for one by name, and only if the operator permitted it:

```duso
exec("sh", {args = ["-c", "ps aux | grep arland"]})
```

Everything inside that string is shell syntax again, with all that implies. Interpolating untrusted data into it is a command injection, exactly as it would be anywhere else.

## Errors vs. results

`exec()` splits failures the same way `fetch()` does:

- **The command ran and exited non-zero** — a result, not an error. Check `result.ok`. A grep that matched nothing or a `git diff --quiet` that found changes are answers, not failures.
- **The command could not be started** — throws. Not permitted, not found, not executable, bad working directory.
- **The timeout fired, or `kill()` cancelled the script** — throws. Any output captured up to that point is lost with the result object.

```duso
result = exec("grep", {args = ["needle", "haystack.txt"]})
if result.code == 1 then
  print("no match")          // grep's normal "found nothing"
elseif not result.ok then
  throw(result.stderr)        // grep actually failed
end
```

## Environment

`inherit_env` is `false` by default because duso's environment is usually where the database URL and the API keys live, and a command that does not need them should not see them.

An empty environment breaks most real tools, so a small set passes through regardless: `PATH`, `HOME`, `TMPDIR`, `LANG`, `TZ`. These say where the machine keeps things and nothing about who it talks to. Anything else has to be named:

```duso
exec("mysqldump", {
  args = ["mydb"],
  env = {MYSQL_PWD = load("/run/secrets/db")}
})
```

`env` entries are added on top of whatever the child would otherwise get, so they also work alongside `inherit_env = true`.

## Timeouts

There is no default timeout — a command runs as long as it runs. Pass `timeout` for anything that could hang.

When the timeout fires, `exec()` throws immediately and control returns to your script at the deadline. Cleaning up the child happens behind you: it gets `SIGTERM`, then `SIGKILL` two seconds later if it is still alive. A process that ignores signals cannot hold up the call the timeout exists to unblock.

On unix the whole process group is signalled, so a command that backgrounds work goes down with it. On Windows only the command itself is killed and anything it started may survive.

`kill(pid)` on a spawned script does the same to whatever it is currently running, so killing a worker does not leave a process behind.

## Output limits

Each stream is captured up to `max_output` bytes (1 MB by default) and the rest is discarded, with `truncated` set. The command keeps running either way — it never blocks on a full pipe.

```duso
result = exec("journalctl", {args = ["-u", "arland"], max_output = 65536})
if result.truncated then print("(output clipped)") end
```

For output larger than you want in memory, have the command write a file and read that instead.

## Permission

exec() is off unless duso is started with `-allow-exec`, whose value is a comma-separated list:

```bash
duso app.du -allow-exec "certbot,systemctl start arland,systemctl stop arland,systemctl restart arland"
```

Each entry is a command, optionally followed by arguments that a call must supply verbatim as its leading arguments:

- `certbot` — certbot, with any arguments
- `systemctl restart arland` — only that exact invocation, plus anything after it

A call outside the list throws, and the error names the flag that would permit it:

```
exec: ls not permitted; start duso with -allow-exec "ls"
```

Every entry is resolved to an absolute path once, at startup, and that path is what runs. There is no `PATH` lookup at call time, so nothing a script or a request does later can change which binary a permitted name refers to. An entry that cannot be resolved refuses to boot:

```
allow-exec: certbot not found
```

Better to hear that at deploy time than the first time the handler fires.

Every call is logged to stderr with its arguments. The environment is never logged.

`-allow-exec` with no list permits anything and warns loudly at startup. It exists for local experimentation; do not ship it.

## The allowlist is not a sandbox

It controls **which program** runs, not what that program can be talked into doing. Most real commands have an escape hatch — `certbot --pre-hook`, `git -c core.pager=`, `tar --to-command`, `find -exec`. Permitting one of those means trusting everyone who can influence its arguments.

What the allowlist reliably buys you is that argument data can never choose the program. A compromised handler cannot reach for a shell, or for `curl`, if you did not allow them.

So:

- Prefer entries pinned to full argument prefixes (`systemctl restart arland`) over bare command names.
- Treat a bare command name as granting everything that command can do.
- Allowing a shell (`sh`, `bash`, `zsh`, ...) is allowing arbitrary execution. duso warns at startup when you do.

A child process is also outside anything duso enforces about its own script — file sandboxing applies to the script, not to a program it starts.

## Notes

- Blocking. Long-running commands belong in [`spawn()`](/docs/reference/spawn.md).
- Not available in embedded Go applications unless the host wires it up.
- The child's stdin is closed when `stdin` is not supplied, so a command that reads input sees EOF instead of hanging.
- No streaming, no binary stdout, no async handles yet.

## See Also

- [spawn() - Run a script in the background](/docs/reference/spawn.md)
- [kill() - Terminate a spawned process](/docs/reference/kill.md)
- [fetch() - Make HTTP requests](/docs/reference/fetch.md)
- [sys() - Read CLI flags](/docs/reference/sys.md)
