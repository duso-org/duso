package runtime

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Date/time functions

// translateDateFormat converts standard date format (YYYY-MM-DD) to Go's format
func translateDateFormat(format string) string {
	replacements := map[string]string{
		"YYYY": "2006",
		"YY":   "06",
		"MM":   "01",
		"DD":   "02",
		"HH":   "15",
		"mm":   "04",
		"ss":   "05",
	}

	result := format
	for standard, goFormat := range replacements {
		result = strings.ReplaceAll(result, standard, goFormat)
	}
	return result
}

// builtinNow returns current local time as timestamp (local time values as UTC)
func builtinNow(evaluator *Evaluator, args map[string]any) (any, error) {
	now := time.Now()
	return float64(time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), 0, time.UTC).Unix()), nil
}

// builtinTimer returns current time with sub-second precision for benchmarking
func builtinTimer(evaluator *Evaluator, args map[string]any) (any, error) {
	return float64(time.Now().UnixNano()) / 1e9, nil
}

// builtinTimestamp returns current Unix timestamp in UTC, or UTC + offset for a location
func builtinTimestamp(evaluator *Evaluator, args map[string]any) (any, error) {
	utc := time.Now().UTC().Unix()

	// If no argument, return current UTC time
	if _, ok := args["0"]; !ok {
		return float64(utc), nil
	}

	// Parse timezone/offset argument
	tzArg, ok := args["0"].(string)
	if !ok {
		if num, isNum := args["0"].(float64); isNum {
			// Handle numeric offset (hours)
			tzArg = formatOffset(num)
		} else {
			return nil, fmt.Errorf("timestamp() argument must be a string (timezone/offset) or number (hours offset)")
		}
	}

	var loc *time.Location
	var err error

	// Try to parse as fixed offset (starts with + or -)
	if len(tzArg) > 0 && (tzArg[0] == '+' || tzArg[0] == '-') {
		loc, err = parseFixedOffset(tzArg)
		if err != nil {
			return nil, fmt.Errorf("timestamp() invalid offset: %v", err)
		}
	} else {
		// Try to load as IANA timezone
		loc, err = time.LoadLocation(tzArg)
		if err != nil {
			return nil, fmt.Errorf("timestamp() unknown timezone %q: %v", tzArg, err)
		}
	}

	// Get the offset for this location at current time
	_, offsetSeconds := time.Now().In(loc).Zone()

	// Return UTC + offset
	return float64(utc + int64(offsetSeconds)), nil
}

// parseFixedOffset parses offset strings like "+7", "-5:30", "+05:30"
func parseFixedOffset(offset string) (*time.Location, error) {
	// Handle +/-N or +/-N:MM format
	sign := 1
	if offset[0] == '-' {
		sign = -1
		offset = offset[1:]
	} else if offset[0] == '+' {
		offset = offset[1:]
	}

	parts := strings.Split(offset, ":")
	hours := 0
	minutes := 0

	h, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid offset format")
	}
	hours = h

	if len(parts) > 1 {
		m, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid offset format")
		}
		minutes = m
	}

	seconds := sign * (hours*3600 + minutes*60)
	return time.FixedZone(fmt.Sprintf("%+d", sign*hours), seconds), nil
}

// formatOffset converts numeric offset (hours) to string format
func formatOffset(hours float64) string {
	if hours >= 0 {
		return fmt.Sprintf("+%.0f", hours)
	}
	return fmt.Sprintf("%.0f", hours)
}

// builtinFormatTime formats a Unix timestamp to string
func builtinFormatTime(evaluator *Evaluator, args map[string]any) (any, error) {
	var timestamp float64
	var ok bool

	// Accept either number or string that parses as number
	arg := args["0"]
	if num, isNum := arg.(float64); isNum {
		timestamp = num
		ok = true
	} else if str, isStr := arg.(string); isStr {
		// Try to parse string as number (e.g., JSON timestamp from string)
		num, err := strconv.ParseFloat(str, 64)
		if err == nil {
			timestamp = num
			ok = true
		}
	}

	if !ok {
		return nil, fmt.Errorf("format_time() requires a number or numeric string as first argument")
	}

	format := "2006-01-02 15:04:05" // default

	if formatArg, ok := args["1"].(string); ok {
		switch formatArg {
		case "iso":
			format = "2006-01-02T15:04:05Z"
		case "date":
			format = "2006-01-02"
		case "time":
			format = "15:04:05"
		case "long_date":
			format = "January 2, 2006"
		case "long_date_dow":
			format = "Mon January 2, 2006"
		case "short_date":
			format = "Jan 2, 2006"
		case "short_date_dow":
			format = "Mon Jan 2, 2006"
		default:
			// User provided custom format
			format = translateDateFormat(formatArg)
		}
	}

	t := time.Unix(int64(timestamp), 0).UTC()
	return t.Format(format), nil
}

// builtinParseTime parses a date string to Unix timestamp
func builtinParseTime(evaluator *Evaluator, args map[string]any) (any, error) {
	dateStr, ok := args["0"].(string)
	if !ok {
		return nil, fmt.Errorf("parse_time() requires a string as first argument")
	}

	// If format provided, use it
	if formatArg, ok := args["1"].(string); ok {
		format := translateDateFormat(formatArg)
		t, err := time.Parse(format, dateStr)
		if err != nil {
			return nil, fmt.Errorf("parse_time() failed to parse %q with format %q: %v", dateStr, formatArg, err)
		}
		return float64(t.Unix()), nil
	}

	// No format hint: try common patterns
	commonFormats := []string{
		"2006-01-02T15:04:05Z", // ISO with Z
		"2006-01-02T15:04:05",  // ISO without Z
		"2006-01-02 15:04:05",  // Default
		"2006-01-02",           // Date only
		"January 2, 2006",      // Long date
		"Mon January 2, 2006",  // Long date with day of week
		"Jan 2, 2006",          // Short date
		"Mon Jan 2, 2006",      // Short date with day of week
	}

	for _, format := range commonFormats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return float64(t.Unix()), nil
		}
	}

	return nil, fmt.Errorf("parse_time() could not parse %q - try providing a format", dateStr)
}

