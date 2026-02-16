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

// builtinNow returns current Unix timestamp (seconds by default, milliseconds if arg is true)
func builtinNow(evaluator *Evaluator, args map[string]any) (any, error) {
	// Optional undocumented boolean parameter: true = milliseconds, false/nil = seconds
	if ms, ok := args["0"].(bool); ok && ms {
		return float64(time.Now().UnixMilli()), nil
	}
	return float64(time.Now().Unix()), nil
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

