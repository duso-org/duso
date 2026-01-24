// Package markdown provides markdown to ANSI formatting for terminal output.
package markdown

import (
	"regexp"
	"strings"
)

// ANSI color codes
const (
	// Standard colors
	Black   = "\033[30m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"

	// Bright colors
	BrightBlack   = "\033[90m"
	BrightRed     = "\033[91m"
	BrightGreen   = "\033[92m"
	BrightYellow  = "\033[93m"
	BrightBlue    = "\033[94m"
	BrightMagenta = "\033[95m"
	BrightCyan    = "\033[96m"
	BrightWhite   = "\033[97m"

	// Bold variants
	BoldBlack      = "\033[1;30m"
	BoldRed        = "\033[1;31m"
	BoldGreen      = "\033[1;32m"
	BoldYellow     = "\033[1;33m"
	BoldBlue       = "\033[1;34m"
	BoldMagenta    = "\033[1;35m"
	BoldCyan       = "\033[1;36m"
	BoldWhite      = "\033[1;37m"
	BoldBrightYellow = "\033[1;93m"

	// Text styles
	Underline = "\033[4m"

	// Reset
	Reset = "\033[0m"
)

// Theme defines the color scheme for markdown rendering
type Theme struct {
	Text   string // Regular text
	H1     string // # headers
	H2     string // ## headers
	H3     string // ### headers
	H4     string // #### headers
	Code   string // Code blocks and inline code
	Link   string // Links
	Bold   string // **bold**
	Italic string // *italic*
}

// DefaultTheme is the default color scheme
var DefaultTheme = Theme{
	Text:   White,
	H1:     BoldBrightYellow,
	H2:     Yellow,
	H3:     BoldWhite,
	H4:     BoldWhite,
	Code:   BrightCyan,
	Link:   BrightGreen,
	Bold:   BoldWhite,
	Italic: BoldWhite,
}

// ToANSI converts markdown to ANSI-formatted text using the default theme
func ToANSI(markdown string) string {
	return ToANSIWithTheme(markdown, DefaultTheme)
}

// ToANSIWithTheme converts markdown to ANSI-formatted text using a custom theme
func ToANSIWithTheme(markdown string, theme Theme) string {
	lines := strings.Split(markdown, "\n")
	var result []string
	inCodeBlock := false

	for _, line := range lines {
		// Skip blank lines in input - we'll control spacing through formatting
		if line == "" {
			continue
		}

		// Handle code block markers
		if strings.HasPrefix(line, "```") {
			if !inCodeBlock {
				// Entering code block - add blank line before if needed
				if len(result) > 0 && result[len(result)-1] != "" {
					result = append(result, "")
				}
			} else {
				// Exiting code block - add blank line after
				result = append(result, "")
			}
			inCodeBlock = !inCodeBlock
			// Skip the fence line itself
			continue
		}

		// Format code block content
		if inCodeBlock {
			result = append(result, theme.Code+line+Reset)
			continue
		}

		// Format headers (keep the # symbols)
		headerMatch := regexp.MustCompile(`^(#+)\s+(.*)$`).FindStringSubmatch(line)
		if len(headerMatch) == 3 {
			hashes := headerMatch[1]
			content := headerMatch[2]

			// Add blank line before header if needed
			if len(result) > 0 && result[len(result)-1] != "" {
				result = append(result, "")
			}

			// Choose color based on header level
			headerColor := theme.H4
			switch len(hashes) {
			case 1:
				headerColor = theme.H1
			case 2:
				headerColor = theme.H2
			case 3:
				headerColor = theme.H3
			}

			line = headerColor + Underline + content + Reset
			result = append(result, line)

			// Add blank line after header
			result = append(result, "")
			continue
		}

		// Format inline elements in the line
		line = formatInline(line, theme)
		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

// formatInline applies inline formatting to a line (bold, italic, code, links)
func formatInline(line string, theme Theme) string {
	// Format inline code backticks (must be before bold/italic to avoid conflicts)
	// Remove the backticks and apply code color
	line = regexp.MustCompile("`([^`]+)`").ReplaceAllString(line, theme.Code+"$1"+Reset)

	// Format links [text](url) -> text in link color with underline
	line = regexp.MustCompile(`\[([^\]]+)\]\(([^\)]+)\)`).ReplaceAllString(line, theme.Link+Underline+"$1"+Reset)

	// Format bold **text**
	line = regexp.MustCompile(`\*\*([^\*]+)\*\*`).ReplaceAllString(line, theme.Bold+"$1"+Reset)

	// Format italic *text* (but not already-formatted bold)
	line = regexp.MustCompile(`(?:\*([^\*\n]+)\*)`).ReplaceAllString(line, theme.Italic+"$1"+Reset)

	// Wrap entire line in text color if it doesn't start with a color code
	if !strings.HasPrefix(line, "\033[") && line != "" {
		line = theme.Text + line + Reset
	}

	return line
}
