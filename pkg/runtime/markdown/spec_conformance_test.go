package markdown

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"
)

// specCase mirrors entries in CommonMark's spec.json.
type specCase struct {
	Markdown string `json:"markdown"`
	HTML     string `json:"html"`
	Example  int    `json:"example"`
	Section  string `json:"section"`
}

var (
	specPath = flag.String("spec", "/tmp/cm-spec.json", "path to CommonMark spec.json")
	verbose  = flag.Bool("v_fails", false, "print each failing case")
	section  = flag.String("section", "", "limit to a single section")
)

// TestCommonMarkSpec runs the CommonMark spec.json cases against ToHTML and
// reports pass/fail counts overall and per section. It does NOT fail the
// test — its purpose is to surface conformance as a metric.
func TestCommonMarkSpec(t *testing.T) {
	data, err := os.ReadFile(*specPath)
	if err != nil {
		t.Skipf("spec.json not available at %s: %v (download from https://spec.commonmark.org/0.31.2/spec.json)", *specPath, err)
	}
	var cases []specCase
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatalf("parse spec.json: %v", err)
	}

	type sect struct {
		pass, total int
		failed      []int
	}
	sects := map[string]*sect{}
	totalPass, totalFail := 0, 0

	// We use a CommonMark-flavored options set: no GFM extensions, no
	// smartquotes, no heading IDs (matches reference output).
	opts := Options{}

	for _, c := range cases {
		if *section != "" && c.Section != *section {
			continue
		}
		got := ToHTML(c.Markdown, opts)
		want := c.HTML
		if _, ok := sects[c.Section]; !ok {
			sects[c.Section] = &sect{}
		}
		s := sects[c.Section]
		s.total++
		if got == want {
			s.pass++
			totalPass++
			continue
		}
		totalFail++
		s.failed = append(s.failed, c.Example)
		if *verbose {
			t.Logf("FAIL ex %d (%s)\nin:    %q\nwant:  %q\ngot:   %q\n", c.Example, c.Section, c.Markdown, want, got)
		}
	}

	var names []string
	for k := range sects {
		names = append(names, k)
	}
	sort.Strings(names)
	t.Logf("\n=== CommonMark conformance: %d / %d (%.1f%%) ===",
		totalPass, totalPass+totalFail, 100*float64(totalPass)/float64(totalPass+totalFail))
	for _, n := range names {
		s := sects[n]
		t.Logf("  %-40s  %3d / %3d", n, s.pass, s.total)
	}

	if testing.Verbose() && !*verbose {
		// Print one example failure per section so we have actionable signal.
		t.Logf("\nFirst failing example per section:")
		for _, n := range names {
			s := sects[n]
			if len(s.failed) == 0 {
				continue
			}
			t.Logf("  %s: ex %v", n, s.failed)
		}
	}
	_ = strings.TrimSpace
	_ = fmt.Sprintf
}

func firstN(xs []int, n int) []int {
	if len(xs) <= n {
		return xs
	}
	return xs[:n]
}
