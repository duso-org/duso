package runtime

import (
	"container/heap"
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/duso-org/duso/pkg/core"
	"github.com/duso-org/duso/pkg/script"
)

// intervalTokenRe matches a count followed by a unit word, e.g. "30s", "1day",
// "2mo", "1yr". The word itself is deliberately lenient (see classifyUnitWord) -
// full words and common abbreviations all work, not just a single fixed letter.
var intervalTokenRe = regexp.MustCompile(`^(\d+(?:\.\d+)?)([A-Za-z]+)$`)

// dateTokenRe matches a date-shaped token like "2026-01-15" or "2026/01/15".
// "/" and "-" are treated as interchangeable separators.
var dateTokenRe = regexp.MustCompile(`^(\d{4})[/-](\d{1,2})[/-](\d{1,2})$`)

// timeTokenRe matches a time-of-day token like "06:00" or "15:00:00".
var timeTokenRe = regexp.MustCompile(`^(\d{1,2}):(\d{2})(?::(\d{2}))?$`)

// classifyUnitWord maps a unit word (any case) to one of the six canonical
// units addInterval understands (Y/M/D/h/m/s) plus a multiplier to fold in
// units addInterval doesn't have natively (week -> 7 days).
//
// Case-insensitive throughout - month vs minute (the one pair that used to
// need case-sensitivity to disambiguate) is instead resolved by length:
// month always requires at least "mo" (mo/mon/month/months); a bare "m", or
// anything starting "mi" (min/minute/minutes), means minute. That's a more
// forgiving rule than relying on case, which is easy to get wrong when
// composing these strings by hand or generating them programmatically.
func classifyUnitWord(word string) (unit string, multiplier float64, ok bool) {
	w := strings.ToLower(word)
	switch {
	case strings.HasPrefix(w, "mo"): // mo, mon, month, months
		return "M", 1, true
	case w == "m" || strings.HasPrefix(w, "mi"): // m, min, minute, minutes
		return "m", 1, true
	case strings.HasPrefix(w, "y"): // y, yr, year, years
		return "Y", 1, true
	case strings.HasPrefix(w, "w"): // w, wk, week, weeks -> 7 days
		return "D", 7, true
	case strings.HasPrefix(w, "d"): // d, day, days
		return "D", 1, true
	case strings.HasPrefix(w, "h"): // h, hr, hour, hours
		return "h", 1, true
	case strings.HasPrefix(w, "s"): // s, sec, second, seconds
		return "s", 1, true
	default:
		return "", 0, false
	}
}

// parseIntervalToken parses a single interval token like "30s" or "2mo" into a count and unit.
func parseIntervalToken(s string) (float64, string, error) {
	match := intervalTokenRe.FindStringSubmatch(s)
	if match == nil {
		return 0, "", fmt.Errorf("invalid interval %q - expected a count followed by a unit (e.g. \"30s\", \"1day\", \"2mo\", \"1yr\")", s)
	}
	var count float64
	if _, err := fmt.Sscanf(match[1], "%g", &count); err != nil {
		return 0, "", fmt.Errorf("invalid interval count in %q: %v", s, err)
	}
	if count <= 0 {
		return 0, "", fmt.Errorf("invalid interval %q - count must be positive", s)
	}
	unit, multiplier, ok := classifyUnitWord(match[2])
	if !ok {
		return 0, "", fmt.Errorf("invalid interval %q - unrecognized unit %q", s, match[2])
	}
	return count * multiplier, unit, nil
}

// parseDateToken parses a date-shaped token into year/month/day.
func parseDateToken(s string) (year, month, day int, err error) {
	match := dateTokenRe.FindStringSubmatch(s)
	if match == nil {
		return 0, 0, 0, fmt.Errorf("invalid date %q - expected YYYY-MM-DD or YYYY/MM/DD", s)
	}
	fmt.Sscanf(match[1], "%d", &year)
	fmt.Sscanf(match[2], "%d", &month)
	fmt.Sscanf(match[3], "%d", &day)
	if month < 1 || month > 12 {
		return 0, 0, 0, fmt.Errorf("invalid date %q - month out of range", s)
	}
	if day < 1 || day > 31 {
		return 0, 0, 0, fmt.Errorf("invalid date %q - day out of range", s)
	}
	return year, month, day, nil
}

// parseTimeToken parses a time-of-day-shaped token into hour/min/sec.
func parseTimeToken(s string) (hour, min, sec int, err error) {
	match := timeTokenRe.FindStringSubmatch(s)
	if match == nil {
		return 0, 0, 0, fmt.Errorf("invalid time %q - expected HH:MM or HH:MM:SS", s)
	}
	fmt.Sscanf(match[1], "%d", &hour)
	fmt.Sscanf(match[2], "%d", &min)
	if match[3] != "" {
		fmt.Sscanf(match[3], "%d", &sec)
	}
	if hour < 0 || hour > 23 || min < 0 || min > 59 || sec < 0 || sec > 59 {
		return 0, 0, 0, fmt.Errorf("invalid time %q - out of range", s)
	}
	return hour, min, sec, nil
}

// parseScheduleSpec tokenizes a schedule spec string on whitespace and classifies
// each token independently by shape - an interval token (Y/M/D/h/m/s suffix), a
// date token ("/" or "-"), or a time-of-day token (":"). Order doesn't matter,
// since the three shapes never overlap: "06:00 1d" and "1d 06:00" parse the same.
//
// unit == "" means no interval was given (a one-shot). anchor == nil means no
// date/time was given (a pure interval, anchored at schedule()-call time).
// A date or time token alone (no interval) composes a one-shot anchor, filling
// in whichever of date/time is missing from the server's local "now" - the
// wall-clock portion is always interpreted in the server's local timezone
// (time.Local), so DST is handled correctly without per-job timezone config;
// the resulting time.Time still carries the correct absolute UTC instant
// regardless of Location, so nothing downstream needs to convert it.
func parseScheduleSpec(spec string) (count float64, unit string, anchor *time.Time, err error) {
	var haveDate, haveTime, haveInterval, haveRelative bool
	var year, month, day, hour, min, sec int
	var relCount float64
	var relUnit string

	for _, tok := range strings.Fields(spec) {
		switch {
		case strings.HasPrefix(tok, "+") && intervalTokenRe.MatchString(strings.TrimPrefix(tok, "+")):
			if haveRelative {
				return 0, "", nil, fmt.Errorf("schedule spec %q has more than one relative (+...) token", spec)
			}
			relCount, relUnit, err = parseIntervalToken(strings.TrimPrefix(tok, "+"))
			if err != nil {
				return 0, "", nil, err
			}
			haveRelative = true
		case intervalTokenRe.MatchString(tok):
			if haveInterval {
				return 0, "", nil, fmt.Errorf("schedule spec %q has more than one interval token", spec)
			}
			count, unit, err = parseIntervalToken(tok)
			if err != nil {
				return 0, "", nil, err
			}
			haveInterval = true
		case strings.Contains(tok, ":"):
			if haveTime {
				return 0, "", nil, fmt.Errorf("schedule spec %q has more than one time token", spec)
			}
			hour, min, sec, err = parseTimeToken(tok)
			if err != nil {
				return 0, "", nil, err
			}
			haveTime = true
		case strings.Contains(tok, "/") || strings.Contains(tok, "-"):
			if haveDate {
				return 0, "", nil, fmt.Errorf("schedule spec %q has more than one date token", spec)
			}
			year, month, day, err = parseDateToken(tok)
			if err != nil {
				return 0, "", nil, err
			}
			haveDate = true
		default:
			return 0, "", nil, fmt.Errorf("schedule spec %q has an unrecognized token %q", spec, tok)
		}
	}

	if haveRelative && (haveDate || haveTime) {
		return 0, "", nil, fmt.Errorf("schedule spec %q combines a relative offset (+...) with an absolute date/time - pick one", spec)
	}

	if haveRelative {
		t := addInterval(time.Now(), relCount, relUnit)
		return count, unit, &t, nil
	}

	if !haveDate && !haveTime {
		return count, unit, nil, nil
	}

	now := time.Now().In(time.Local)
	if !haveDate {
		year, month, day = now.Year(), int(now.Month()), now.Day()
	}
	if !haveTime {
		hour, min, sec = 0, 0, 0
	}
	t := time.Date(year, time.Month(month), day, hour, min, sec, 0, time.Local)
	return count, unit, &t, nil
}

// addInterval advances t by count units of unit, clamping calendar arithmetic
// (Y/M) to the last valid day of the target month instead of overflowing into
// the next month (Go's time.AddDate overflows: Jan 31 + 1 month -> Mar 3).
func addInterval(t time.Time, count float64, unit string) time.Time {
	switch unit {
	case "s":
		return t.Add(time.Duration(count) * time.Second)
	case "m":
		return t.Add(time.Duration(count) * time.Minute)
	case "h":
		return t.Add(time.Duration(count) * time.Hour)
	case "D":
		return t.AddDate(0, 0, int(count))
	case "M":
		return addDateClamped(t, 0, int(count))
	case "Y":
		return addDateClamped(t, int(count), 0)
	default:
		// Unreachable given intervalTokenRe's character class, but keep a safe fallback.
		return t.AddDate(0, 0, int(count))
	}
}

// addDateClamped adds years/months to t, clamping the day-of-month to the last
// valid day of the resulting month rather than letting it overflow (Go's
// AddDate normalizes Jan 31 + 1 month into Mar 3; this clamps to Feb 28/29).
func addDateClamped(t time.Time, years, months int) time.Time {
	year, month, day := t.Date()
	targetMonthIndex := int(month) - 1 + months
	targetYear := year + years + targetMonthIndex/12
	targetMonth := time.Month(targetMonthIndex%12 + 1)
	if targetMonth <= 0 {
		targetMonth += 12
		targetYear--
	}

	lastDay := lastDayOfMonth(targetYear, targetMonth)
	if day > lastDay {
		day = lastDay
	}

	return time.Date(targetYear, targetMonth, day, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
}

func lastDayOfMonth(year int, month time.Month) int {
	// Day 0 of the following month is the last day of this month.
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// scheduleJob is one entry in the scheduler's min-heap, and mirrors what's
// stored (minus the live heap index) in the duso_schedule datastore.
type scheduleJob struct {
	id          string
	count       float64
	unit        string
	nextFire    time.Time
	scriptPath  string
	contextData any
	index       int // heap index, maintained by container/heap
}

// scheduleHeapT implements container/heap.Interface: a min-heap sorted by nextFire.
type scheduleHeapT []*scheduleJob

func (h scheduleHeapT) Len() int            { return len(h) }
func (h scheduleHeapT) Less(i, j int) bool  { return h[i].nextFire.Before(h[j].nextFire) }
func (h scheduleHeapT) Swap(i, j int)       { h[i], h[j] = h[j], h[i]; h[i].index = i; h[j].index = j }
func (h *scheduleHeapT) Push(x any) {
	job := x.(*scheduleJob)
	job.index = len(*h)
	*h = append(*h, job)
}
func (h *scheduleHeapT) Pop() any {
	old := *h
	n := len(old)
	job := old[n-1]
	old[n-1] = nil
	job.index = -1
	*h = old[0 : n-1]
	return job
}

var (
	schedMu      sync.Mutex
	schedHeap    = scheduleHeapT{}
	schedIndex   = make(map[string]*scheduleJob) // id -> live heap entry, for unschedule()
	schedWake    = make(chan struct{}, 1)
	schedStarted sync.Once
)

// wakeScheduler nudges the scheduler loop to recompute its wait, e.g. after a
// new job is pushed that fires sooner than whatever the loop was already
// waiting on, or after a job is removed.
func wakeScheduler() {
	select {
	case schedWake <- struct{}{}:
	default:
	}
}

// ensureSchedulerRunning starts the background scheduler loop exactly once,
// lazily, the first time schedule() is called in this process.
func ensureSchedulerRunning() {
	schedStarted.Do(func() {
		go schedulerLoop()
	})
}

// schedulerLoop is the single background goroutine driving all scheduled jobs.
// It always sleeps via a one-shot timer reset to the earliest due job (never a
// periodic poll), so idle cost is ~zero regardless of how many jobs exist or
// how far out the next one is - modeled on the existing ExpiryHeap, but with
// a one-shot time.Timer instead of a fixed-interval ticker since job intervals
// are typically much coarser than TTL sweep granularity.
func schedulerLoop() {
	defer core.RecoverPanic("schedule scheduler loop")

	timer := time.NewTimer(time.Hour)
	if !timer.Stop() {
		<-timer.C
	}

	for {
		schedMu.Lock()
		var wait time.Duration
		hasJob := schedHeap.Len() > 0
		if hasJob {
			wait = time.Until(schedHeap[0].nextFire)
			if wait < 0 {
				wait = 0
			}
		}
		schedMu.Unlock()

		if hasJob {
			timer.Reset(wait)
		}

		if hasJob {
			select {
			case <-timer.C:
				fireDueJobs()
			case <-schedWake:
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
			}
		} else {
			<-schedWake
		}
	}
}

// fireDueJobs pops every job due at or before now, fires each one, and
// reschedules recurring jobs by pushing their next occurrence back onto the heap.
func fireDueJobs() {
	now := time.Now()
	var due []*scheduleJob

	schedMu.Lock()
	for schedHeap.Len() > 0 && !schedHeap[0].nextFire.After(now) {
		job := heap.Pop(&schedHeap).(*scheduleJob)
		delete(schedIndex, job.id)
		due = append(due, job)
	}
	schedMu.Unlock()

	ds := GetDatastore("duso_schedule", nil)

	for _, job := range due {
		fireJob(job)

		if job.unit == "" {
			// One-shot: fired once, done. No reschedule, no lingering record.
			ds.Delete(job.id)
			continue
		}

		job.nextFire = addInterval(job.nextFire, job.count, job.unit)

		schedMu.Lock()
		heap.Push(&schedHeap, job)
		schedIndex[job.id] = job
		schedMu.Unlock()

		ds.Set(job.id, scheduleRecord(job))
	}
}

// fireJob runs the job's target script in a background goroutine, fire-and-forget,
// the same execution shape as spawn() (fresh evaluator, no parent frame since
// the scheduler fires independently of any calling script).
func fireJob(job *scheduleJob) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Fprintf(os.Stderr, "schedule: panic in %s (job %q): %v\n", job.scriptPath, job.id, r)
			}
		}()

		program, err := globalInterpreter.ParseScript(job.scriptPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "schedule: failed to parse %s (job %q): %v\n", job.scriptPath, job.id, err)
			return
		}

		frame := &script.InvocationFrame{
			Filename: job.scriptPath,
			Line:     1,
			Col:      1,
			Reason:   "schedule",
			Details:  map[string]any{},
		}

		eval := script.NewEvaluator()
		procCtx, cancel := context.WithCancel(context.Background())
		defer cancel()

		runCtx := &script.RequestContext{
			Frame:        frame,
			ProcessCtx:   procCtx,
			Interpreter:  globalInterpreter,
			Evaluator:    eval,
			OutputWriter: globalInterpreter.OutputWriter,
		}

		gid := script.GetGoroutineID()
		contextCopy := script.DeepCopyAny(job.contextData)
		script.SetRequestContextWithData(gid, runCtx, contextCopy)
		defer script.ClearRequestContext(gid)

		SetContextGetter(gid, func() any {
			ctx, ok := script.GetRequestContext(gid)
			if !ok {
				return nil
			}
			return ctx.Data
		})
		defer ClearContextGetter(gid)

		result := script.ExecuteScript(program, globalInterpreter, frame, runCtx, procCtx)
		if result != nil && result.Error != nil {
			var msg string
			if dusoErr, ok := result.Error.(*script.DusoError); ok {
				msg = script.FormatErrorWithStack(dusoErr)
			} else {
				msg = result.Error.Error()
			}
			fmt.Fprintf(os.Stderr, "schedule: error in %s (job %q): %s\n", job.scriptPath, job.id, msg)
		}
	}()
}

func scheduleRecord(job *scheduleJob) map[string]any {
	var every any
	if job.unit != "" {
		every = fmt.Sprintf("%g%s", job.count, job.unit)
	}
	return map[string]any{
		"every":     every,
		"script":    job.scriptPath,
		"next_fire": float64(job.nextFire.Unix()),
		"context":   job.contextData,
	}
}

// builtinSchedule registers a recurring job: schedule([id,] every, script[, context]).
// Positional: "0"=every (or id, if "1" is also present and "2" is the script -
// kept simple for now: id is always named "id", every/script are positional 0/1).
func builtinSchedule(evaluator *Evaluator, args map[string]any) (any, error) {
	var everyStr string
	if e, ok := args["every"]; ok {
		if s, ok := e.(string); ok {
			everyStr = s
		}
	} else if e, ok := args["0"]; ok {
		if s, ok := e.(string); ok {
			everyStr = s
		}
	}
	if everyStr == "" {
		return nil, fmt.Errorf("schedule() requires a spec string (e.g. \"30s\", \"06:00 1D\", \"2026-01-15 15:00\")")
	}

	count, unit, anchor, err := parseScheduleSpec(everyStr)
	if err != nil {
		return nil, err
	}
	if unit == "" && anchor == nil {
		return nil, fmt.Errorf("schedule() spec %q didn't resolve to an interval or a date/time", everyStr)
	}

	var scriptPath string
	if s, ok := args["script"]; ok {
		if str, ok := s.(string); ok {
			scriptPath = str
		}
	} else if s, ok := args["1"]; ok {
		if str, ok := s.(string); ok {
			scriptPath = str
		}
	}
	if scriptPath == "" {
		return nil, fmt.Errorf("schedule() requires a script path argument")
	}

	var contextData any
	if c, ok := args["context"]; ok {
		contextData = c
	} else if c, ok := args["2"]; ok {
		contextData = c
	}

	var id string
	if i, ok := args["id"]; ok {
		if str, ok := i.(string); ok {
			id = str
		}
	} else if i, ok := args["3"]; ok {
		if str, ok := i.(string); ok {
			id = str
		}
	}
	if id == "" {
		generated, err := builtinUUID(evaluator, map[string]any{})
		if err != nil {
			return nil, fmt.Errorf("schedule() failed to generate id: %v", err)
		}
		id = generated.(string)
	}

	// Resolve the script path relative to the calling script's directory, same as spawn()/run().
	if ctx, ok := script.CurrentRequestContext(evaluator); ok && ctx.Frame != nil && ctx.Frame.Filename != "" {
		scriptPath = script.ResolveScriptPath(scriptPath, ctx.Frame.Filename)
	}

	ensureSchedulerRunning()

	var nextFire time.Time
	switch {
	case anchor != nil && unit == "":
		// One-shot at a specific instant. If it's already past, it fires
		// almost immediately - the heap/timer handle a near-zero wait for free.
		nextFire = *anchor
	case anchor != nil && unit != "":
		// Anchored recurring: start at the anchor, but if it's already in the
		// past, silently fast-forward to the next future occurrence instead of
		// firing once per missed tick.
		nextFire = *anchor
		now := time.Now()
		for !nextFire.After(now) {
			nextFire = addInterval(nextFire, count, unit)
		}
	default:
		// Plain interval, no anchor: first fire is one interval from now.
		nextFire = addInterval(time.Now(), count, unit)
	}

	job := &scheduleJob{
		id:          id,
		count:       count,
		unit:        unit,
		nextFire:    nextFire,
		scriptPath:  scriptPath,
		contextData: contextData,
	}

	schedMu.Lock()
	if existing, ok := schedIndex[id]; ok {
		heap.Remove(&schedHeap, existing.index)
	}
	heap.Push(&schedHeap, job)
	schedIndex[id] = job
	schedMu.Unlock()
	wakeScheduler()

	ds := GetDatastore("duso_schedule", nil)
	ds.Set(id, scheduleRecord(job))

	return id, nil
}

// builtinUnschedule cancels a job by id: unschedule(id).
func builtinUnschedule(evaluator *Evaluator, args map[string]any) (any, error) {
	var id string
	if i, ok := args["id"]; ok {
		if str, ok := i.(string); ok {
			id = str
		}
	} else if i, ok := args["0"]; ok {
		if str, ok := i.(string); ok {
			id = str
		}
	}
	if id == "" {
		return nil, fmt.Errorf("unschedule() requires an id argument")
	}

	schedMu.Lock()
	found := false
	if existing, ok := schedIndex[id]; ok {
		heap.Remove(&schedHeap, existing.index)
		delete(schedIndex, id)
		found = true
	}
	schedMu.Unlock()

	ds := GetDatastore("duso_schedule", nil)
	ds.Delete(id)

	return found, nil
}
