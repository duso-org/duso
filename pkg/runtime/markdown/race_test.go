package markdown

import (
	"sync"
	"testing"
)

// TestConcurrentToHTML hammers ToHTML from many goroutines simultaneously,
// with different inputs and shared option/theme values, to exercise any
// potential data race under `go test -race`.
func TestConcurrentToHTML(t *testing.T) {
	inputs := []string{
		"# Heading\n\nA paragraph with **bold** and *italic*.\n",
		"- item one\n- item two\n  - nested\n- item three\n",
		"> A blockquote\n> with **strong** content.\n",
		"| a | b |\n|---|---|\n| 1 | 2 |\n",
		"```go\nfunc main() {}\n```\n",
		"Reference [link][ref].\n\n[ref]: /url \"title\"\n",
	}
	opts := DefaultOptions()
	theme := DefaultTheme()

	var wg sync.WaitGroup
	for g := 0; g < 32; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for n := 0; n < 200; n++ {
				src := inputs[(g+n)%len(inputs)]
				_ = ToHTML(src, opts)
				_ = ToANSI(src, theme)
				_ = ToText(src)
			}
		}(g)
	}
	wg.Wait()
}
