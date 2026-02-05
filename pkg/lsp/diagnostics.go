package lsp

import (
	"github.com/duso-org/duso/pkg/script"
)

// DiagnosticSeverity represents the severity of a diagnostic
type DiagnosticSeverity int

const (
	DiagnosticSeverityError       DiagnosticSeverity = 1
	DiagnosticSeverityWarning     DiagnosticSeverity = 2
	DiagnosticSeverityInformation DiagnosticSeverity = 3
	DiagnosticSeverityHint        DiagnosticSeverity = 4
)

// Diagnostic represents a diagnostic message (error, warning, etc.)
type Diagnostic struct {
	Range              Range                  `json:"range"`
	Severity           *DiagnosticSeverity    `json:"severity,omitempty"`
	Code               *string                `json:"code,omitempty"`
	Source             *string                `json:"source,omitempty"`
	Message            string                 `json:"message"`
	Tags               []int                  `json:"tags,omitempty"`
	RelatedInformation []*DiagnosticRelated   `json:"relatedInformation,omitempty"`
	CodeDescription    *CodeDescription       `json:"codeDescription,omitempty"`
}

// DiagnosticRelated represents related information in a diagnostic
type DiagnosticRelated struct {
	Location Location `json:"location"`
	Message  string   `json:"message"`
}

// CodeDescription represents code description information
type CodeDescription struct {
	Href string `json:"href"`
}

// ConvertDusoError converts a Duso error to an LSP Diagnostic
func ConvertDusoError(err error, uri string) *Diagnostic {
	if err == nil {
		return nil
	}

	dusoErr, ok := err.(*script.DusoError)
	if !ok {
		// Generic error with unknown position
		severity := DiagnosticSeverityError
		return &Diagnostic{
			Range: Range{
				Start: Position{Line: 0, Character: 0},
				End:   Position{Line: 0, Character: 0},
			},
			Severity: &severity,
			Source:   strPtr("duso"),
			Message:  err.Error(),
		}
	}

	severity := DiagnosticSeverityError
	pos := DusoPositionToLSP(dusoErr.Position)

	return &Diagnostic{
		Range: Range{
			Start: pos,
			End:   pos,
		},
		Severity: &severity,
		Source:   strPtr("duso"),
		Message:  dusoErr.Message,
	}
}

// ConvertDusoErrors converts multiple Duso errors to LSP Diagnostics
func ConvertDusoErrors(err error, uri string) []*Diagnostic {
	diagnostics := []*Diagnostic{}

	if err == nil {
		return diagnostics
	}

	// Try to convert as a Duso error
	if diag := ConvertDusoError(err, uri); diag != nil {
		diagnostics = append(diagnostics, diag)
	}

	return diagnostics
}

// strPtr returns a pointer to a string
func strPtr(s string) *string {
	return &s
}

// intPtr returns a pointer to an int
func intPtr(i int) *int {
	return &i
}

// DiagnosticsFromParseError creates diagnostics from a parse error
// The parser may return a simple error or DusoError
func DiagnosticsFromParseError(err error, uri string) []*Diagnostic {
	return ConvertDusoErrors(err, uri)
}

// DiagnosticsFromAST would validate the AST and return diagnostics
// For Phase 1, we only report parse errors
func DiagnosticsFromAST(ast *script.Program, uri string) []*Diagnostic {
	// Phase 1: No semantic analysis, return empty array
	// Phase 2: Add type checking, undefined variables, etc.
	return []*Diagnostic{}
}
