package markdown

import (
	"strings"
	"testing"
)

// Mixed-content sample covering most of what production docs look like.
const benchSample = `# Heading One

This is a paragraph with **bold**, *italic*, ` + "`inline code`" + `, and a [link](https://example.com "title") in it. Soft line break here
and continuing.

## Heading Two

A second paragraph. ~~Strikethrough~~ and an autolink: <https://example.org/some/path>.

- bullet one
- bullet two with **emphasis**
  - nested bullet
  - another nested
- bullet three

1. first
2. second with ` + "`code`" + `
3. third

> A blockquote
> spanning multiple lines
> with **bold** inside.
>
> A second paragraph in the blockquote.

| Col A | Col B | Col C |
|:------|:-----:|------:|
| left  | center | right |
| more  | data   | here  |

` + "```go" + `
func hello() {
	fmt.Println("world")
	for i := 0; i < 10; i++ {
		_ = i * 2
	}
}
` + "```" + `

A final paragraph with [a reference link][ref] and an image: ![alt text](image.png "title").

[ref]: https://example.com/reference "Reference Title"

- [ ] task one
- [x] task two
- [ ] task three
`

// largeBench is benchSample repeated many times to model a long document.
var largeBench = strings.Repeat(benchSample, 50)

func BenchmarkToHTML_Small(b *testing.B) {
	b.SetBytes(int64(len(benchSample)))
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = ToHTML(benchSample, DefaultOptions())
	}
}

func BenchmarkToHTML_Large(b *testing.B) {
	b.SetBytes(int64(len(largeBench)))
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = ToHTML(largeBench, DefaultOptions())
	}
}

func BenchmarkToANSI(b *testing.B) {
	b.SetBytes(int64(len(benchSample)))
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = ToANSI(benchSample, DefaultOptions(), DefaultTheme())
	}
}

// Specialized benches that isolate hot paths.

func BenchmarkParagraphs(b *testing.B) {
	src := strings.Repeat("This is a paragraph with **bold** and *italic* and `code` inline.\n\n", 100)
	b.SetBytes(int64(len(src)))
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = ToHTML(src, Options{})
	}
}

func BenchmarkLists(b *testing.B) {
	var sb strings.Builder
	for i := 0; i < 100; i++ {
		sb.WriteString("- item ")
		sb.WriteString(strings.Repeat("a", 20))
		sb.WriteByte('\n')
	}
	src := sb.String()
	b.SetBytes(int64(len(src)))
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = ToHTML(src, Options{})
	}
}

func BenchmarkCodeBlocks(b *testing.B) {
	var sb strings.Builder
	for i := 0; i < 50; i++ {
		sb.WriteString("```\nsome code line one\nsome code line two\n```\n\n")
	}
	src := sb.String()
	b.SetBytes(int64(len(src)))
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = ToHTML(src, Options{})
	}
}
