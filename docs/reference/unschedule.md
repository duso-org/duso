# unschedule()

Cancel a job registered with [`schedule()`](/docs/reference/schedule.md) by its id. Available in `duso` CLI only.

`unschedule(id)`

## Parameters

- `id` (string) - The job id to cancel, as returned by (or passed to) `schedule()`

## Returns

Boolean - `true` if a job with that id was found and removed, `false` if no such job existed.

## Examples

```duso
job_id = schedule("30s", "heartbeat.du")
sleep(90)
unschedule(job_id)   // stops it; no more fires after this
```

With a stable, self-chosen id:

```duso
schedule("1D", "backup.du", id = "nightly-backup")
// ...later, from anywhere...
unschedule("nightly-backup")
```

Cancelling a job that doesn't exist is safe and just returns `false`:

```duso
found = unschedule("does-not-exist")
print(found)   // false
```

## Notes

- Removes the job from both the scheduler and the `duso_schedule` datastore record - it won't show up in `datastore("duso_schedule").get(id)` afterward.
- Safe to call from any script, not just the one that registered the job, as long as you know its id.
- One-shot jobs remove themselves automatically after firing - `unschedule()` is only needed if you want to cancel one *before* it fires.

## See Also

- [schedule() - Register a scheduled job](/docs/reference/schedule.md)
- [datastore() - Namespaced key/value store](/docs/reference/datastore.md)
