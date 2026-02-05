package lsp

import (
	"embed"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/duso-org/duso/pkg/script"
)

// Server represents an LSP server
type Server struct {
	documents   *DocumentManager
	cache       *ASTCache
	interp      *script.Interpreter
	embeddedFS  embed.FS
	mu          sync.RWMutex
	initialized bool
}

// NewServer creates a new LSP server with embedded filesystem for docs
func NewServer(interp *script.Interpreter, fs embed.FS) *Server {
	return &Server{
		documents:   NewDocumentManager(),
		cache:       NewASTCache(100),
		interp:      interp,
		embeddedFS:  fs,
		initialized: false,
	}
}

// InitializeRequest represents the initialize request
type InitializeRequest struct {
	ProcessID             *int              `json:"processId"`
	RootPath              *string           `json:"rootPath"`
	RootURI               *string           `json:"rootUri"`
	InitializationOptions json.RawMessage   `json:"initializationOptions"`
	Capabilities          ClientCapabilities `json:"capabilities"`
}

// ClientCapabilities represents client capabilities
type ClientCapabilities struct {
	Workspace      map[string]interface{} `json:"workspace,omitempty"`
	TextDocument   map[string]interface{} `json:"textDocument,omitempty"`
	Experimental   map[string]interface{} `json:"experimental,omitempty"`
}

// InitializeResponse represents the initialize response
type InitializeResponse struct {
	Capabilities ServerCapabilities `json:"capabilities"`
	ServerInfo   *ServerInfo        `json:"serverInfo,omitempty"`
}

// ServerCapabilities describes what the server can do
type ServerCapabilities struct {
	TextDocumentSync   int                     `json:"textDocumentSync"`
	HoverProvider      bool                    `json:"hoverProvider"`
	DefinitionProvider bool                    `json:"definitionProvider"`
	ReferencesProvider bool                    `json:"referencesProvider,omitempty"`
	DocumentSymbolProvider bool                `json:"documentSymbolProvider,omitempty"`
	RenameProvider     bool                    `json:"renameProvider,omitempty"`
	CompletionProvider *CompletionOptions      `json:"completionProvider,omitempty"`
}

// CompletionOptions describes completion provider options
type CompletionOptions struct {
	TriggerCharacters []string `json:"triggerCharacters,omitempty"`
}

// ServerInfo contains server information
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Initialize handles the initialize request
func (s *Server) Initialize() *InitializeResponse {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.initialized = true

	textDocSyncFull := 1 // Full sync

	return &InitializeResponse{
		Capabilities: ServerCapabilities{
			TextDocumentSync:    textDocSyncFull,
			HoverProvider:       true,
			DefinitionProvider:  true,
			ReferencesProvider:  true,
			CompletionProvider: &CompletionOptions{
				TriggerCharacters: []string{},
			},
		},
		ServerInfo: &ServerInfo{
			Name:    "Duso",
			Version: "1.0.0",
		},
	}
}

// Shutdown handles the shutdown request
func (s *Server) Shutdown() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.initialized = false
	return nil
}

// DidOpenParams represents parameters for the didOpen notification
type DidOpenParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

// TextDocumentItem represents a document
type TextDocumentItem struct {
	URI        string `json:"uri"`
	LanguageID string `json:"languageId"`
	Version    int    `json:"version"`
	Text       string `json:"text"`
}

// DidOpen handles the didOpen notification
func (s *Server) DidOpen(params DidOpenParams) []*Diagnostic {
	uri := params.TextDocument.URI
	doc := s.documents.Open(uri, params.TextDocument.Version, params.TextDocument.Text)

	// Parse and cache AST
	return s.parseAndCache(doc)
}

// DidChangeParams represents parameters for the didChange notification
type DidChangeParams struct {
	TextDocument   VersionedTextDocumentIdentifier `json:"textDocument"`
	ContentChanges []TextDocumentContentChangeEvent `json:"contentChanges"`
}

// VersionedTextDocumentIdentifier identifies a versioned document
type VersionedTextDocumentIdentifier struct {
	URI     string `json:"uri"`
	Version int    `json:"version"`
}

// TextDocumentContentChangeEvent represents a content change
type TextDocumentContentChangeEvent struct {
	Range       *Range `json:"range,omitempty"`
	RangeLength *int   `json:"rangeLength,omitempty"`
	Text        string `json:"text"`
}

// DidChange handles the didChange notification
func (s *Server) DidChange(params DidChangeParams) []*Diagnostic {
	uri := params.TextDocument.URI
	version := params.TextDocument.Version

	// Get existing document
	doc := s.documents.Get(uri)
	if doc == nil {
		doc = &Document{URI: uri, Version: version, Text: ""}
	}

	// Apply changes (Phase 1: Full sync)
	for _, change := range params.ContentChanges {
		if change.Range == nil {
			// Full document sync
			doc.Text = change.Text
		} else {
			// Incremental sync (Phase 2)
			// For now, treat as full sync
			doc.Text = change.Text
		}
	}

	// Update document
	doc.Version = version
	s.documents.Update(uri, version, doc.Text)

	// Parse and cache AST
	return s.parseAndCache(doc)
}

// DidCloseParams represents parameters for the didClose notification
type DidCloseParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// DidClose handles the didClose notification
func (s *Server) DidClose(params DidCloseParams) {
	uri := params.TextDocument.URI
	s.documents.Close(uri)
	s.cache.Delete(uri)
}

// HoverParams represents parameters for the hover request
type HoverParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

// Hover handles the hover request
func (s *Server) Hover(params HoverParams) *HoverInfo {
	doc := s.documents.Get(params.TextDocument.URI)
	if doc == nil {
		return nil
	}

	entry, _ := s.cache.Get(params.TextDocument.URI, doc.Version)
	if entry == nil {
		return nil
	}

	return ProvideHover(s, doc, params.Position, entry)
}

// Definition handles the definition request
func (s *Server) Definition(params DefinitionParams) *Location {
	doc := s.documents.Get(params.TextDocument.URI)
	if doc == nil {
		return nil
	}

	entry, _ := s.cache.Get(params.TextDocument.URI, doc.Version)
	if entry == nil {
		return nil
	}

	return ProvideDefinition(doc, params.Position, entry)
}

// ReferenceParams represents parameters for the references request
type ReferenceParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
	Context      ReferenceContext       `json:"context"`
}

// ReferenceContext represents reference context
type ReferenceContext struct {
	IncludeDeclaration bool `json:"includeDeclaration"`
}

// References handles the references request
func (s *Server) References(params ReferenceParams) []*Location {
	doc := s.documents.Get(params.TextDocument.URI)
	if doc == nil {
		return nil
	}

	entry, _ := s.cache.Get(params.TextDocument.URI, doc.Version)
	if entry == nil {
		return nil
	}

	// Convert LSP position to Duso position
	dusoPos := LSPPositionToDuso(params.Position)

	// Find the node at the position
	node := FindNodeAtPosition(entry.AST, dusoPos)
	if node == nil {
		return nil
	}

	// Get the identifier name
	var identName string
	switch n := node.(type) {
	case *script.Identifier:
		identName = n.Name
	default:
		return nil
	}

	if identName == "" {
		return nil
	}

	// Find references
	return FindReferences(entry.AST, identName, doc.URI)
}

// CompletionParams represents parameters for the completion request
type CompletionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
	Context      CompletionContext      `json:"context,omitempty"`
}

// CompletionContext represents completion context
type CompletionContext struct {
	TriggerKind      int    `json:"triggerKind"`
	TriggerCharacter string `json:"triggerCharacter,omitempty"`
}

// Completion handles the completion request
func (s *Server) Completion(params CompletionParams) []*CompletionItem {
	doc := s.documents.Get(params.TextDocument.URI)
	if doc == nil {
		return nil
	}

	entry, _ := s.cache.Get(params.TextDocument.URI, doc.Version)
	if entry == nil {
		return nil
	}

	return ProvideCompletion(s, doc, params.Position, entry)
}

// parseAndCache parses a document and caches the result
func (s *Server) parseAndCache(doc *Document) []*Diagnostic {
	if doc == nil {
		return []*Diagnostic{}
	}

	// Parse the document
	lexer := script.NewLexer(doc.Text)
	tokens := lexer.Tokenize()
	parser := script.NewParserWithFile(tokens, URIToString(doc.URI))

	ast, err := parser.Parse()

	diagnostics := []*Diagnostic{}
	if err != nil {
		diagnostics = DiagnosticsFromParseError(err, doc.URI)
	}

	// Also check AST for issues
	if ast != nil {
		astDiags := DiagnosticsFromAST(ast, doc.URI)
		diagnostics = append(diagnostics, astDiags...)
	}

	// Cache the result
	s.cache.Set(doc.URI, doc.Version, ast, diagnostics)

	return diagnostics
}

// PublishDiagnosticsParams represents parameters for the publishDiagnostics notification
type PublishDiagnosticsParams struct {
	URI         string         `json:"uri"`
	Diagnostics []*Diagnostic  `json:"diagnostics"`
}

// String returns a string representation for debugging
func (s *Server) String() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return fmt.Sprintf("LSP Server (initialized=%v, documents=%d)", s.initialized, s.documents.Count())
}
