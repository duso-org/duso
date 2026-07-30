# schedule()

Register a job that runs a script on a recurring interval, at a specific date/time, or once relative to now. Available in `duso` CLI only.

`schedule(spec, script [, context] [, id])`

## Parameters

- `spec` (string) - What to run and when. See **Spec Format** below.
- `script` (string) - Path to the script to run, resolved relative to the calling script's directory
- `context` (optional, object) - Context object passed to the fired script, retrievable there with [`context()`](/docs/reference/context.md)
- `id` (optional, string) - A stable identifier for this job - also the key it's stored under in the `duso_schedule` datastore (see below). Calling `schedule()` again with the same `id` replaces the existing job rather than creating a duplicate. If omitted, a `uuid()` is generated and returned. Accepted as the 4th positional argument or as `id = "..."` - named is recommended.

## Returns

The job's id (string) - either the one you passed, or the generated one if you didn't pass one. Keep it if you might want to `unschedule()` this job later.

## Spec Format

A `spec` is one or more whitespace-separated tokens. Each token is classified independently by its own shape, so order never matters - `"06:00 1D"` and `"1D 06:00"` parse identically.

| Token shape | Example | Means |
|---|---|---|
| count + unit | `30s`, `1D`, `2mo` | Recurring interval |
| `+` + count + unit | `+30s`, `+2h` | One-shot (or anchor), offset from *now* |
| date | `2026-01-15`, `2026/01/15` | Anchor date (`/` and `-` are interchangeable) |
| time | `06:00`, `15:00:00` | Anchor time of day |

**Interval units** are lenient about spelling - full words, plurals, and common abbreviations all work: `y`/`yr`/`year`/`years`, `mo`/`month`/`months`, `w`/`wk`/`week`/`weeks` (folds into 7 days), `d`/`day`/`days`, `h`/`hr`/`hour`/`hours`, `s`/`sec`/`second`/`seconds`. Month always requires at least `mo` (`mo`, `mon`, `month`...); a bare `m`, or anything starting `mi` (`min`, `minute`...), always means minute.

Year/month intervals clamp to the last valid day of the target month instead of overflowing (`2026-01-31` + 1 month lands on `2026-02-28`, not `2026-03-03`).

Date and time tokens are parsed in the **server's local timezone** (whatever the OS/container's `TZ` is set to), so daylight saving is handled correctly without any per-job timezone configuration. Internally everything is tracked as an absolute UTC instant regardless.

### Combining tokens

- **Interval alone** (`"30s"`) - recurring, first fire one interval from now.
- **Date and/or time alone** (`"2026-01-15 09:00"`, or just `"09:00"`, or just `"2026-01-15"`) - one-shot at that instant. Missing date fills in today (server-local); missing time defaults to midnight. If the instant is already in the past, it fires almost immediately rather than erroring.
- **Date/time + interval** (`"2026-01-01 03:00 1D"`) - anchored recurring: first fire is the anchor, then repeats on the interval. If the anchor is already in the past, missed occurrences are silently skipped (fast-forwarded) rather than firing once per miss.
- **`+offset` alone** (`"+30s"`) - one-shot, fires exactly that far from now.
- **`+offset` + interval** (`"+2h 1D"`) - anchored recurring, starting that far from now (e.g. "start in 2 hours, then run daily").
- **`+offset` combined with a date/time token** - rejected as an error.
- At most one of each token shape is allowed; more than one date, time, interval, or `+offset` token in the same spec is an error, as is any unrecognized token.

## Examples

Simple recurring interval:

```duso
schedule("30s", "heartbeat.du")
```

Friendly unit spelling:

```duso
schedule("1yr", "renew_cert.du")
schedule("1mo", "invoice.du")
schedule("1min", "poll.du")   // minute, not month
```

Daily at a specific time, regardless of which token comes first:

```duso
schedule("06:00 1D", "morning_report.du")
schedule("1D 06:00", "morning_report.du")   // identical
```

Anchored monthly billing run:

```duso
schedule("2026-01-01 03:00 1mo", "billing.du")
```

One-shot at a specific future date/time:

```duso
schedule("2026-08-02 09:00:00", "send_reminder.du")
```

One-shot relative to now:

```duso
schedule("+30s", "send_confirmation.du")
```

With a stable id and context:

```duso
schedule("1D", "backup.du", {retention_days = 30}, id = "nightly-backup")
```

## Inspecting scheduled jobs

Job records live in a reserved `duso_schedule` datastore namespace - writable only through `schedule()`/`unschedule()`, but freely readable like any other datastore:

```duso
jobs = datastore("duso_schedule")
jobs.get("nightly-backup")            // this job's record
jobs.select(function(id, v) return true end) // list every scheduled job
```

A record looks like (the id is the datastore key, not a field in the value):

```duso
{every="1D", script="backup.du", next_fire=1785345648, context={retention_days=30}, last_pid=42}
```

`every` is `nil` for one-shot jobs, which are also removed from the datastore automatically right after they fire. `last_pid` is the PID of the most recent (or currently running) firing, and is killable like any [`spawn()`](/docs/reference/spawn.md)ed process:

```duso
kill(rec.last_pid)   // stop a run that's hung, without affecting the schedule itself
```

## Overlap handling

`schedule()` reschedules a job's next occurrence purely from interval math, the moment it fires - it doesn't track whether the previous run has finished. If a handler can occasionally take longer than its own interval, that's deliberately left to the handler itself to guard against, using the datastore's existing `set_once()` as a lock:

```duso
if not datastore("locks").set_once("nightly-backup-running", true) then
  exit()   // previous run still in flight - skip this tick
end

// ... do the work ...

datastore("locks").delete("nightly-backup-running")
```

## Notes

- Jobs run in a fresh, isolated script instance, the same execution model as [`spawn()`](/docs/reference/spawn.md) - no shared globals with the script that called `schedule()`. Use the datastore to communicate results back.
- `context` passed to `schedule()` is available inside the fired script via [`context()`](/docs/reference/context.md).
- Multi-node deployments should keep server local time consistent across nodes, since date/time tokens are resolved against each node's own local timezone.

## See Also

- [unschedule() - Cancel a scheduled job](/docs/reference/unschedule.md)
- [spawn() - Run a script in the background](/docs/reference/spawn.md)
- [kill() - Terminate a running process by PID](/docs/reference/kill.md)
- [context() - Access runtime context](/docs/reference/context.md)
- [datastore() - Namespaced key/value store](/docs/reference/datastore.md)
- [parse_time() - Parse a time string to a timestamp](/docs/reference/parse_time.md)
