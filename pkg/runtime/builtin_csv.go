package runtime

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"github.com/duso-org/duso/pkg/script"
)

// csvComma resolves the optional delimiter argument, defaulting to a comma.
// The first *rune* is taken, not the first byte, so multi-byte delimiters like
// "→" work instead of decoding to a garbage rune that never matches.
func csvComma(args map[string]any) rune {
	delimiterArg := GetArg(args, 1, "delimiter")
	if delimiterArg == nil {
		return ','
	}
	delimiterStr, ok := delimiterArg.(string)
	if !ok {
		return ','
	}
	if runes := []rune(delimiterStr); len(runes) > 0 {
		return runes[0]
	}
	return ','
}

// csvField renders one cell. nil becomes an empty field rather than the text
// "nil" — empty is how CSV spells "no value", and "nil" would read back from
// parse_csv() as a four-character string.
func csvField(v Value) string {
	if v.IsNil() {
		return ""
	}
	return v.String()
}

// csvCommaFast is csvComma for the []Value calling convention.
func csvCommaFast(args []Value) rune {
	if len(args) < 2 || !args[1].IsString() {
		return ','
	}
	if runes := []rune(args[1].AsString()); len(runes) > 0 {
		return runes[0]
	}
	return ','
}

// parseCSVRecords parses CSV text into an array of arrays of strings, built as
// Values directly. Going through []any would mean allocating the whole result
// twice — once here and once more when the evaluator converts it back — which
// for a large file is most of the cost of parsing it.
func parseCSVRecords(text string, comma rune) (Value, error) {
	reader := csv.NewReader(strings.NewReader(text))
	reader.Comma = comma
	// Safe because each field is copied into a Value before the next Read.
	reader.ReuseRecord = true

	records := []Value{}
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return script.NewNil(), fmt.Errorf("parse_csv() error: %v", err)
		}

		fields := make([]Value, len(record))
		for i, field := range record {
			fields[i] = script.NewString(field)
		}
		records = append(records, script.NewArray(fields))
	}

	return script.NewArray(records), nil
}

// builtinParseCSV parses a CSV string to array of arrays
func builtinParseCSV(evaluator *Evaluator, args map[string]any) (any, error) {
	csvString := GetArg(args, 0, "str")
	if csvString == nil {
		return nil, fmt.Errorf("parse_csv() requires a string as first argument")
	}

	stringValue, ok := csvString.(string)
	if !ok {
		return nil, fmt.Errorf("parse_csv() requires a string as first argument")
	}

	return parseCSVRecords(stringValue, csvComma(args))
}

// fastParseCSV is the []Value form of parse_csv (see builtin_fast.go)
func fastParseCSV(evaluator *Evaluator, args []Value) (Value, error) {
	if len(args) < 1 || !args[0].IsString() {
		return script.NewNil(), fmt.Errorf("parse_csv() requires a string as first argument")
	}
	return parseCSVRecords(args[0].AsString(), csvCommaFast(args))
}

// builtinFormatCSV formats array of arrays to CSV string
func builtinFormatCSV(evaluator *Evaluator, args map[string]any) (any, error) {
	arrayArg := GetArg(args, 0, "array")
	if arrayArg == nil {
		return nil, fmt.Errorf("format_csv() requires an array as first argument")
	}

	arrayPtr, ok := arrayArg.(*[]Value)
	if !ok {
		return nil, fmt.Errorf("format_csv() requires an array as first argument")
	}

	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)
	writer.Comma = csvComma(args)

	for _, row := range *arrayPtr {
		var record []string

		// Handle both array of values and array of arrays. Test the type rather
		// than the slice: an empty array yields a nil slice, and treating that
		// as "not a row" would write out its Duso rendering, "[]".
		if row.IsArray() {
			rowArray := row.AsArray()
			record = make([]string, 0, len(rowArray))
			for _, field := range rowArray {
				record = append(record, csvField(field))
			}
		} else {
			record = append(record, csvField(row))
		}

		if err := writer.Write(record); err != nil {
			return nil, fmt.Errorf("format_csv() error: %v", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("format_csv() error: %v", err)
	}

	return buffer.String(), nil
}
