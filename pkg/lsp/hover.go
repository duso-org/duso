package lsp

import (
	"fmt"
	"strings"

	"github.com/duso-org/duso/pkg/cli"
	"github.com/duso-org/duso/pkg/script"
)

// HoverInfo represents hover information
type HoverInfo struct {
	Contents string `json:"contents"` // Markdown formatted
	Range    *Range `json:"range,omitempty"`
}

// ProvideHover returns hover information for a position in the document
func ProvideHover(server *Server, doc *Document, pos Position, entry *ASTCacheEntry) *HoverInfo {
	if entry == nil || entry.AST == nil {
		return nil
	}

	// Convert LSP position to Duso position
	dusoPos := LSPPositionToDuso(pos)

	// Find the node at the position
	node := FindNodeAtPosition(entry.AST, dusoPos)
	if node == nil {
		return nil
	}

	// Generate hover information based on node type
	contents := generateHoverContents(server, node, doc.Text)
	if contents == "" {
		return nil
	}

	return &HoverInfo{
		Contents: contents,
	}
}

// generateHoverContents generates hover information for a node
func generateHoverContents(server *Server, node script.Node, text string) string {
	switch n := node.(type) {
	case *script.Identifier:
		// Look up identifier information
		return getIdentifierInfo(server, n.Name)

	case *script.FunctionDef:
		// Show function signature
		return fmt.Sprintf("```duso\nfunction %s(%s)\n```", n.Name, formatParameters(n.Parameters))

	case *script.CallExpr:
		// Show called function info if it's an identifier
		if ident, ok := n.Func.(*script.Identifier); ok {
			return getIdentifierInfo(server, ident.Name)
		}

	case *script.AssignStatement:
		// Show variable being assigned
		if ident, ok := n.Target.(*script.Identifier); ok {
			return fmt.Sprintf("Variable: `%s`", ident.Name)
		}
	}

	return ""
}

// getIdentifierInfo returns information about a built-in identifier from embedded docs
func getIdentifierInfo(server *Server, name string) string {
	if server == nil {
		return ""
	}

	// Try to read the documentation from embedded files
	docPath := fmt.Sprintf("docs/reference/%s.md", name)
	content, err := cli.EmbeddedFileRead(docPath)
	if err != nil {
		// No documentation found - return a minimal hover showing we know about it
		// This helps debug why some functions don't work
		return fmt.Sprintf("Built-in function: %s\n\n(Documentation not found)", name)
	}

	// Parse the markdown to extract summary and signature
	result := parseDocForHover(string(content))
	if result == "" {
		return fmt.Sprintf("Built-in function: %s", name)
	}
	return result
}

// parseDocForHover extracts summary and signature from markdown documentation
func parseDocForHover(content string) string {
	lines := strings.Split(content, "\n")
	var summary strings.Builder
	var signature strings.Builder
	var inSignatureSection bool

	for i, line := range lines {
		// Skip the title line
		if strings.HasPrefix(line, "# ") {
			continue
		}

		// Extract summary (first non-empty paragraph after title)
		if !inSignatureSection && strings.TrimSpace(line) != "" && !strings.HasPrefix(line, "#") {
			if summary.Len() == 0 && !strings.HasPrefix(line, "##") {
				summary.WriteString(line)
				summary.WriteString("\n")
			}
		}

		// Find and extract signature section
		if strings.HasPrefix(line, "## Signature") {
			inSignatureSection = true
			continue
		}

		// Extract signature code block
		if inSignatureSection && strings.HasPrefix(line, "```") {
			signature.WriteString(line)
			signature.WriteString("\n")
			// Collect next line(s) until closing ```
			for j := i + 1; j < len(lines); j++ {
				signature.WriteString(lines[j])
				signature.WriteString("\n")
				if strings.HasPrefix(lines[j], "```") {
					break
				}
			}
			break
		}
	}

	var result strings.Builder

	// Add summary
	summaryText := strings.TrimSpace(summary.String())
	if summaryText != "" {
		result.WriteString(summaryText)
		result.WriteString("\n\n")
	}

	// Add signature
	signatureText := strings.TrimSpace(signature.String())
	if signatureText != "" {
		result.WriteString(signatureText)
	}

	return result.String()
}

// formatParameters formats function parameters for display
func formatParameters(params []*script.Parameter) string {
	if len(params) == 0 {
		return ""
	}

	result := ""
	for i, param := range params {
		if i > 0 {
			result += ", "
		}
		result += param.Name
		if param.Default != nil {
			result += "?"
		}
	}
	return result
}
